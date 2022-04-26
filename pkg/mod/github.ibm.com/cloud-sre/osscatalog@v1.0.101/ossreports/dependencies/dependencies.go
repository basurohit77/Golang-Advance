package dependencies

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/debug"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

const emptycell = " "

var countPattern = regexp.MustCompile(`count=(\d+)`)

type ossEntryRecordType struct {
	OSSName           string `column:"OSS (CRN) service-name,15"`
	OSSDisplayName    string `column:"OSS Display Name,20"`
	EntryType         string `column:"OSS Type"`
	OperationalStatus string `column:"OSS Operational Status,6"`
	NumCHEntries      int    `column:"Number of CH Entries,7"`
	SourceCH          int    `column:"Dependencies from ClearingHouse,7"`
	SourceOther       int    `column:"Dependencies from Other Sources,7"`
	NotOSS            int    `column:"Dependencies Not OSS records,7"`
	NotCloud          int    `column:"Dependencies Not in Cloud?,7"`
	WithIssues        int    `column:"Dependencies with Issues,7"`
	StatusComplete    int    `column:"Status: Complete,7"`
	StatusOther       int    `column:"Status: Other,7"`
	TypeUsage         int    `column:"Type: Usage,7"`
	TypeOther         int    `column:"Type: Other,7"`
	Dependencies      string `column:"Dependencies"`
}

var allInboundOSSEntry = make([]interface{}, 0, 500)
var allOutboundOSSEntry = make([]interface{}, 0, 500)

type dependencyPairRecordType struct {
	key         string `column:"-"`
	srcOSSEntry string `column:"-"`
	Originator  string `column:"Originator,15"`
	Provider    string `column:"Provider,15"`
	Issues      int    `column:"Issues,7"`
	Source      string `column:"Source,15"`
	Status      string `column:"Status,10"`
	Type        string `column:"Type,10"`
	NotOSS      string `column:"NotOSS,10"`
	AllTags     string `column:"All Tags"`
	Notes       string `column:"Notes"`
}

var allDependencyPairs = make(map[string]*dependencyPairRecordType)

func recordOSSEntry(oss *ossrecordextended.OSSServiceExtended, inbound bool) {
	r := &ossEntryRecordType{}
	//TODO: Need better logic to pick the best OSS name, not just the first one
	r.OSSName = string(oss.OSSService.ReferenceResourceName)
	r.OSSDisplayName = oss.OSSService.ReferenceDisplayName
	r.OperationalStatus = string(oss.OSSService.GeneralInfo.OperationalStatus)
	r.EntryType = string(oss.OSSService.GeneralInfo.EntryType)
	if len(oss.OSSService.ProductInfo.ClearingHouseReferences) == 1 && oss.OSSService.ProductInfo.ClearingHouseReferences[0].ID == ossrecord.ProductInfoNone {
		r.NumCHEntries = 0
	} else {
		r.NumCHEntries = len(oss.OSSService.ProductInfo.ClearingHouseReferences)
	}
	var target ossrecord.Dependencies
	if inbound {
		target = oss.OSSService.DependencyInfo.InboundDependencies
	} else {
		target = oss.OSSService.DependencyInfo.OutboundDependencies
	}
	var numNotCloudEntries int
	buf := strings.Builder{}
	for _, d := range target {
		if strings.HasPrefix(d.Service, "~~~") {
			numNotCloudEntries++
			continue
		}
		if buf.Len() > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(d.Service)
		recordDependencyPair(string(oss.OSSService.ReferenceResourceName), d, inbound)
	}
	r.Dependencies = buf.String()
	var notCloudNamedIssue *ossvalidation.NamedValidationIssue
	if inbound {
		notCloudNamedIssue = ossvalidation.CHDependencyInboundNotCloud
	} else {
		notCloudNamedIssue = ossvalidation.CHDependencyOutboundNotCloud
	}
	if is, found := oss.OSSValidation.FindNamedIssue(notCloudNamedIssue); found {
		m := countPattern.FindStringSubmatch(is.GetText())
		if m != nil {
			count, err := strconv.Atoi(m[1])
			if err != nil {
				debug.PrintError("Cannot parse count of NotCloud items in entry for %s -- issue=%s", r.OSSName, is.GetText())
			} else {
				r.NotCloud = count
			}
		} else {
			debug.PrintError("Cannot parse count of NotCloud items in entry for %s (not a number?) -- issue=%s", r.OSSName, is.GetText())
		}
	} else {
		r.NotCloud = 0
	}
	r.SourceCH = target.CountTag("Src:ClearingHouse") - numNotCloudEntries + r.NotCloud
	r.SourceOther = len(target) - target.CountTag("Src:ClearingHouse")
	r.NotOSS = target.CountTag("NotOSS")
	r.WithIssues = target.CountTag("Issues:")
	r.StatusComplete = target.CountTag("Status:Complete")
	r.StatusOther = len(target) - r.StatusComplete - numNotCloudEntries
	r.TypeUsage = target.CountTag("Type:Usage")
	r.TypeOther = len(target) - r.TypeUsage - numNotCloudEntries

	if inbound {
		allInboundOSSEntry = append(allInboundOSSEntry, r)
	} else {
		allOutboundOSSEntry = append(allOutboundOSSEntry, r)
	}
}

func recordDependencyPair(srcOSSEntry string, dependency *ossrecord.Dependency, inbound bool) {
	r := &dependencyPairRecordType{}
	r.srcOSSEntry = srcOSSEntry
	if inbound {
		r.Originator = dependency.Service
		r.Provider = srcOSSEntry
	} else {
		r.Originator = srcOSSEntry
		r.Provider = dependency.Service
	}
	r.key = fmt.Sprintf("%s::%s", r.Originator, r.Provider)
	if issuesStr := dependency.FindTag("Issues:"); issuesStr != "" {
		countIssues, err := strconv.Atoi(issuesStr)
		if err != nil {
			debug.PrintError("Cannot parse number of issues for dependency %s in OSS entry %s: %v", r.key, r.srcOSSEntry, err)
		} else {
			r.Issues = countIssues
		}
	}
	r.Source = dependency.FindTag("Src:")
	r.Status = dependency.FindTag("Status:")
	r.Type = dependency.FindTag("Type:")
	r.NotOSS = dependency.FindTag("NotOSS")
	r.AllTags = fmt.Sprintf("%q", dependency.Tags)
	r.Notes = ""

	if prior, found := allDependencyPairs[r.key]; found {
		if r.AllTags != prior.AllTags {
			r.key = fmt.Sprintf("%s(from %s)", r.key, r.srcOSSEntry)
			r.Notes = fmt.Sprintf("dependency found in OSS entry %s", r.srcOSSEntry)
			prior.Notes = fmt.Sprintf("dependency found in OSS entry %s", prior.srcOSSEntry)
		}
		allDependencyPairs[r.key] = r
	} else {
		allDependencyPairs[r.key] = r
	}
}

// Dependencies generates a summary report of all Dependency information in all OSS entries
func Dependencies(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {

	err := ossmerge.ListAllServices(pattern, func(si *ossmerge.ServiceInfo) {
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		recordOSSEntry(&si.OSSServiceExtended, true)
		recordOSSEntry(&si.OSSServiceExtended, false)
	})
	if err != nil {
		return debug.WrapError(err, "Error listing OSS records for report sheet")
	}

	xl := ossreports.CreateExcel(w, "Dependencies")

	// Outbound sheet
	sort.Slice(allOutboundOSSEntry, func(i, j int) bool {
		return allOutboundOSSEntry[i].(*ossEntryRecordType).OSSName < allOutboundOSSEntry[j].(*ossEntryRecordType).OSSName
	})
	err = xl.AddSheet(allOutboundOSSEntry, 100.0, "Outbound")
	if err != nil {
		return err
	}

	// Inbound sheet
	sort.Slice(allInboundOSSEntry, func(i, j int) bool {
		return allInboundOSSEntry[i].(*ossEntryRecordType).OSSName < allInboundOSSEntry[j].(*ossEntryRecordType).OSSName
	})
	err = xl.AddSheet(allInboundOSSEntry, 100.0, "Inbound")
	if err != nil {
		return err
	}

	// All Dependency Pairs sheet
	pairs := make([]interface{}, 0, len(allDependencyPairs))
	for _, p := range allDependencyPairs {
		pairs = append(pairs, p)
	}
	sort.Slice(pairs, func(i, j int) bool {
		key1 := pairs[i].(*dependencyPairRecordType).key
		key2 := pairs[j].(*dependencyPairRecordType).key
		if key1 == key2 {
			return pairs[i].(*dependencyPairRecordType).Notes < pairs[j].(*dependencyPairRecordType).Notes
		}
		return key1 < key2
	})
	err = xl.AddSheet(pairs, 100.0, "All Dependency Pairs")
	if err != nil {
		return err
	}

	err = xl.Finalize()
	if err != nil {
		return err
	}

	return nil
}

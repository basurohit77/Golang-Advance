package clearinghousenames

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

const emptycell = " "

func generateOSSEntriesSheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	type ossEntryRecordType struct {
		OSSName                string `column:"OSS (CRN) service-name,15"`
		CHName                 string `column:"ClearingHouse Name,15"`
		CHID                   string `column:"ClearingHouse ID,35"`
		OSSDisplayName         string `column:"OSS Display Name,20"`
		EntryType              string `column:"OSS Type"`
		OperationalStatus      string `column:"OSS Operational Status,6"`
		AdditionalOSSNames     string `column:"Additional OSS Names,15"`
		AdditionalCHReferences string `column:"Additional ClearingHouse References,15"`
		CHMissingCRN           string `column:"CH missing CRN service-name attribute,10"`
		CHBadCRN               string `column:"CH has bad CRN service-name attribute,10"`
		CHFoundByCRN           int    `column:"CH found by CRN service-name,10"`
		CHSomeFoundByCRNOnly   int    `column:"CH found by CRN service-name only,10"`
		CHNotFoundByCRN        int    `column:"CH not found by CRN service-name,10"`
		CHSomeFoundByPIDOnly   int    `column:"CH found by PID only,10"`
		CHNotFoundByPID        int    `column:"CH not found by PID,10"`
		CHSomeFoundByNameOnly  int    `column:"CH found by Name only,10"`
		CHNotFoundByName       int    `column:"CH not found by Name,10"`
		Notes                  string `column:"Merge Control Notes,20"`
		OSSTags                string `column:"OSS Tags"`
	}
	var allData = make([]interface{}, 0, 500)

	err := ossmerge.ListAllServices(pattern, func(si *ossmerge.ServiceInfo) {
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}

		r := &ossEntryRecordType{}
		allData = append(allData, r)
		//TODO: Need better logic to pick the best OSS name, not just the first one
		r.OSSName = string(si.OSSService.ReferenceResourceName)
		r.OSSDisplayName = si.OSSService.ReferenceDisplayName
		r.OperationalStatus = string(si.OSSService.GeneralInfo.OperationalStatus)
		r.EntryType = string(si.OSSService.GeneralInfo.EntryType)
		if si.OSSMergeControl != nil {
			r.Notes = si.OSSMergeControl.Notes
		} else {
			r.Notes = emptycell
		}
		r.OSSTags = si.OSSService.GeneralInfo.OSSTags.String()
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHMultipleOSS); found {
			r.AdditionalOSSNames = is.Details
		} else {
			r.AdditionalOSSNames = emptycell
		}
		if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHMissingCRN); found {
			r.CHMissingCRN = `"1"`
		} else {
			r.CHMissingCRN = emptycell
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHBadCRN); found {
			r.CHBadCRN = is.Details
		} else {
			r.CHBadCRN = emptycell
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHFoundByCRN); found {
			r.CHFoundByCRN = len(parseNameList(is.Details))
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHSomeFoundByCRNOnly); found {
			r.CHSomeFoundByCRNOnly = len(parseNameList(is.Details))
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHNotFoundByCRN); found {
			r.CHNotFoundByCRN = len(parseNameList(is.Details))
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHSomeFoundByPIDOnly); found {
			r.CHSomeFoundByPIDOnly = len(parseNameList(is.Details))
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHNotFoundByPID); found {
			r.CHNotFoundByPID = len(parseNameList(is.Details))
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHSomeFoundByNameOnly); found {
			r.CHSomeFoundByNameOnly = len(parseNameList(is.Details))
		}
		if is, found := si.OSSValidation.FindNamedIssue(ossvalidation.CHNotFoundByName); found {
			r.CHNotFoundByName = len(parseNameList(is.Details))
		}
		if len(si.OSSService.ProductInfo.ClearingHouseReferences) > 0 {
			chrefs := si.OSSService.ProductInfo.ClearingHouseReferences
			//TODO: Need better logic to pick the best ClearingHouse entry, not just the first one
			r.CHID = chrefs[0].ID
			r.CHName = chrefs[0].Name
			buf := strings.Builder{}
			for i := range chrefs {
				if i == 0 {
					continue
				}
				buf.WriteString(clearinghouse.MakeCHLabel(chrefs[i].Name, clearinghouse.DeliverableID(chrefs[i].ID)))
			}
			if buf.Len() > 0 {
				r.AdditionalCHReferences = buf.String()
			} else {
				r.AdditionalCHReferences = emptycell
			}
		} else {
			r.CHID = emptycell
			r.CHName = emptycell
			r.AdditionalCHReferences = emptycell
		}
	})
	if err != nil {
		return debug.WrapError(err, "Error listing ServiceInfo records for report sheet")
	}

	sort.Slice(allData, func(i, j int) bool {
		if allData[i].(*ossEntryRecordType).OSSName == allData[j].(*ossEntryRecordType).OSSName {
			return allData[i].(*ossEntryRecordType).CHName < allData[j].(*ossEntryRecordType).CHName
		}
		return allData[i].(*ossEntryRecordType).OSSName < allData[j].(*ossEntryRecordType).OSSName
	})

	err = xl.AddSheet(allData, 100.0, "OSS Entries")
	if err != nil {
		return err
	}

	return nil
}

func generateUnmatchedCHSheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	type unmatchedCHRecordType struct {
		CHNames           string `column:"ClearingHouse Names,35"`
		CHIDs             string `column:"ClearingHouse IDs,35"`
		CHCRNServiceNames string `column:"CRN service names in ClearingHouse entries,35"`
	}
	var allData = make([]interface{}, 0, 500)

	err := ossmerge.ListNameGroups(func(ng *ossmerge.NameGroup) {
		if len(ng.CHIDs) == 0 || len(ng.OSSNames) > 0 {
			return
		}
		r := &unmatchedCHRecordType{}
		allCHNames := make([]string, 0, len(ng.CHIDs))
		allCRNs := make([]string, 0, len(ng.CHIDs))
		for _, chid := range ng.CHIDs {
			if _, found := ossmerge.LookupServicesByCHID(chid); found {
				return
			}
			if ch, found := clearinghouse.LookupSummaryEntryByID(chid, ""); found {
				allCHNames = append(allCHNames, ch.Name)
				if ch.CRNServiceName != "" {
					allCRNs = append(allCRNs, ch.CRNServiceName)
				}
			} else {
				panic(fmt.Sprintf(`ClearingHouseNames() CHID="%s" not found for record=%+v/%p`, chid, ng, ng))
			}
		}
		// If we got this far, it means we really did not find any ServiceInfo record, i.e. no matching OSS entry
		allData = append(allData, r)
		r.CHNames = fmt.Sprintf("%q", allCHNames)
		r.CHIDs = fmt.Sprintf("%q", ng.CHIDs)
		r.CHCRNServiceNames = fmt.Sprintf("%q", allCRNs)
	})
	if err != nil {
		return debug.WrapError(err, "Error listing Name Groups for report sheet")
	}

	sort.Slice(allData, func(i, j int) bool {
		return allData[i].(*unmatchedCHRecordType).CHNames < allData[j].(*unmatchedCHRecordType).CHNames
	})

	err = xl.AddSheet(allData, 30.0, "Unmatched CH Entries")
	if err != nil {
		return err
	}

	return nil
}

func generateNameGroupsSheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	type nameGroupRecordType struct {
		OSSName            string `column:"First OSS (CRN) service-name,15"`
		CHName             string `column:"ClearingHouse Name,15"`
		CHID               string `column:"ClearingHouse ID,35"`
		OSSDisplayName     string `column:"OSS Display Name,20"`
		EntryType          string `column:"OSS Type"`
		OperationalStatus  string `column:"OSS Operational Status,6"`
		AdditionalOSSNames string `column:"Additional OSS Names,15"`
		AdditionalCHNames  string `column:"Additional ClearingHouse Names,15"`
		AdditionalCHIDs    string `column:"Additional ClearingHouse IDs,15"`
		Notes              string `column:"Merge Control Notes,20"`
		OSSTags            string `column:"OSS Tags"`
	}
	var allData = make([]interface{}, 0, 500)

	err := ossmerge.ListNameGroups(func(ng *ossmerge.NameGroup) {
		r := &nameGroupRecordType{}
		allData = append(allData, r)
		if len(ng.OSSNames) > 0 {
			//TODO: Need better logic to pick the best OSS name, not just the first one
			r.OSSName = string(ng.OSSNames[0])
			if len(ng.OSSNames) > 1 {
				r.AdditionalOSSNames = fmt.Sprintf("%q", ng.OSSNames[1:])
			} else {
				r.AdditionalOSSNames = emptycell
			}
			cname := ossmerge.MakeComparableName(string(r.OSSName))
			if si, found := ossmerge.LookupService(cname, false); found {
				r.OSSDisplayName = si.OSSService.ReferenceDisplayName
				r.OperationalStatus = string(si.OSSService.GeneralInfo.OperationalStatus)
				r.EntryType = string(si.OSSService.GeneralInfo.EntryType)
				if si.OSSMergeControl != nil {
					r.Notes = si.OSSMergeControl.Notes
				} else {
					r.Notes = emptycell
				}
				r.OSSTags = si.OSSService.GeneralInfo.OSSTags.String()
			} else {
				panic(fmt.Sprintf(`ClearingHouseNames() OSSName="%s" not found for record=%+v/%p`, r.OSSName, ng, ng))
			}
		}
		if len(ng.CHIDs) > 0 {
			var allCHNames []string
			for _, chid := range ng.CHIDs {
				if ch, found := clearinghouse.LookupSummaryEntryByID(chid, ""); found {
					allCHNames = append(allCHNames, ch.Name)
				} else {
					panic(fmt.Sprintf(`ClearingHouseNames() CHID="%s" not found for record=%+v/%p`, chid, ng, ng))
				}
			}
			//TODO: Need better logic to pick the best ClearingHouse entry, not just the first one
			r.CHID = string(ng.CHIDs[0])
			if len(ng.CHIDs) > 1 {
				r.AdditionalCHIDs = fmt.Sprintf("%q", ng.CHIDs[1:])
			} else {
				r.AdditionalCHIDs = emptycell
			}
			r.CHName = allCHNames[0]
			if len(allCHNames) > 1 {
				r.AdditionalCHNames = fmt.Sprintf("%q", allCHNames[1:])
			} else {
				r.AdditionalCHNames = emptycell
			}
		} else {
			r.CHID = emptycell
			r.AdditionalCHIDs = emptycell
			r.CHName = emptycell
			r.AdditionalCHNames = emptycell
		}
	})
	if err != nil {
		return debug.WrapError(err, "Error listing Name Groups for report sheet")
	}

	sort.Slice(allData, func(i, j int) bool {
		if allData[i].(*nameGroupRecordType).OSSName == allData[j].(*nameGroupRecordType).OSSName {
			return allData[i].(*nameGroupRecordType).CHName < allData[j].(*nameGroupRecordType).CHName
		}
		return allData[i].(*nameGroupRecordType).OSSName < allData[j].(*nameGroupRecordType).OSSName
	})

	err = xl.AddSheet(allData, 100.0, "All Name Groups")
	if err != nil {
		return err
	}

	return nil
}

// ClearingHouseNames generates a summary report of OSS and ClearingHouse entries that share a common name
func ClearingHouseNames(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	if !clearinghouse.HasCHInfo() {
		return fmt.Errorf("Cannot generate ClearingHouseNames names report because ClearingHouse input file has not been loaded")
	}

	xl := ossreports.CreateExcel(w, "ClearingHouseNames")

	err := generateOSSEntriesSheet(xl, pattern)
	if err != nil {
		debug.WrapError(err, "Error generating the OSS Entries sheet of the ClearingHouseNames report")
		return err
	}

	err = generateUnmatchedCHSheet(xl, pattern)
	if err != nil {
		debug.WrapError(err, "Error generating the Unmatched ClearingHouse Entries sheet of the ClearingHouseNames report")
		return err
	}

	err = generateNameGroupsSheet(xl, pattern)
	if err != nil {
		debug.WrapError(err, "Error generating the Name Groups sheet of the ClearingHouseNames report")
		return err
	}

	err = xl.Finalize()
	if err != nil {
		return err
	}

	return nil
}

//var parseNameListPattern = regexp.MustCompile(`\["([^"]*)"(?:,"([^"]*)")*\]`)
//var parseNameListPattern = regexp.MustCompile(`"[^"]+"`)
var parseNameListPattern = regexp.MustCompile(`(\s*\[\s*"?)|("\s*,\s*")|("?\s*\]\s*)`)

func parseNameList(input string) []string {
	m := parseNameListPattern.Split(input, -1)
	result := make([]string, 0, len(m))
	for _, s := range m {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

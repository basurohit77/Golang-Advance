package services4pnp

import (
	"io"
	"regexp"
	"sort"
	"strconv"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

// mainRecordType represents one row in the main sheet of the report
type mainRecordType struct {
	CRNServiceName      string `column:"CRN service-name,15"`
	DisplayName         string `column:"Display Name,20"`
	EntryType           string `column:"Type"`
	OperationalStatus   string `column:"Operational Status,6"`
	PnPEnabled          string `column:"PnP Enabled,6"`
	PnPCandidate        string `column:"PnP Candidate?,6"`
	CRNStatus           string `column:"CRN Validation Status,10"`
	OneCloud            string `column:"OneCloud,6"`
	ClientFacingFlag    string `column:"Client-facing Flag,6"`
	CatalogClientFacing string `column:"Client Facing (from Catalog only),7"`
	//	ClientFacingComputed  string `column:"Client-facing (Computed),10"`
	CategoryID            string `column:"CategoryID,10"`
	CategoryParent        string `column:"CategoryParent,10"`
	NumSameCategory       string `column:"Number of entries with same CategoryID, 6"`
	ServiceNowCategoryID  string `column:"ServiceNow CategoryID,10"`
	ServiceNowName        string `column:"ServiceNow Name\n(if not CRN),15"`
	ServiceNowOnboarded   string `column:"ServiceNow Onboarded,6"`
	SupportEnabled        string `column:"Support Enabled,7"`
	OperationsEnabled     string `column:"Operations Enabled,7"`
	CatalogName           string `column:"Catalog Name(if not CRN),15"`
	CatalogHasDeployments string `column:"Catalog Entry uses Deployments,10"`
	NumStatusIssues       string `column:"Number of potential issues for Status/Notifications,10"`
	Segment               string `column:"Segment"`
	Tribe                 string `column:"Tribe"`
	Notes                 string `column:"Merge Control Notes,20"`
	OSSOnboardingPhase    string `column:"OSS Onboarding Phase,12"`
	OSSTags               string `column:"OSS Tags"`
}

type issuesRecordType struct {
	CRNServiceName string `column:"CRN service-name,15"`
	PnPEnabled     string `column:"PnP Enabled,6"`
	PnPCandidate   string `column:"PnP Candidate?,6"`
	Severity       string `column:"Severity, 10"`
	Issue          string `column:"Issue"`
}

// Labels for columns in the spreadsheet
const (
	OK                        = "ok"
	NO                        = "*NO*"
	OKIAAS                    = "ok-IaaS"
	NOTENABLEDIAAS            = "*IaaS BUT NOT ENABLED*"
	NOTENABLED                = "*NOT-Enabled*"
	MULTIPLE                  = "*MULTIPLE*"
	MISSING                   = "*MISSING*"
	MISMATCH                  = "*MISMATCH*: "
	EMPTY                     = "*EMPTY*"
	NOTFOUND                  = "*NOTFOUND*: "
	PNPNOTENABLEDBUTCANDIDATE = "*pnp_candidate but cannot be enabled*"
	PNPNOTENABLEDBUTINCLUDE   = "*pnp_include but cannot be enabled*"
	OKCANDIDATE               = "ok-candidate"
	OKCANDIDATEUNNEEDED       = "ok-candidate(unneeded)"
	OKINCLUDE                 = "ok-include"
	OKINCLUDEUNNEEDED         = "ok-include(unneeded)"
	EXCLUDE                   = "exclude"
	EXCLUDEUNNEEDED           = "exclude(unneeded)"
	ERROR                     = "*???*"
)

// Services4PnP generates a summary report of status/notification metdadata for all services/components, for use by the PnP project (as an Excel spreadsheet)
func Services4PnP(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	var err error
	mainData := make([]interface{}, 0, 500)
	issuesData := make([]interface{}, 0, 500)

	err = ossmerge.ListAllServices(pattern, func(si *ossmerge.ServiceInfo) {
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		//var issues int
		oss := &si.OSSService
		cat := si.GetSourceMainCatalog()
		r := mainRecordType{}
		r.CRNServiceName = string(oss.ReferenceResourceName)
		if cat != nil {
			if len(si.AdditionalMainCatalog) > 0 {
				r.CatalogName = MULTIPLE
				//issues++
			} else if string(oss.ReferenceResourceName) != cat.Name {
				r.CatalogName = MISMATCH + cat.Name
				//issues++
			} else {
				r.CatalogName = OK
			}
		} else {
			r.CatalogName = MISSING
			//issues++
		}
		if oss.ReferenceDisplayName != "" {
			r.DisplayName = oss.ReferenceDisplayName
		} else {
			r.DisplayName = EMPTY
			//issues++
		}
		r.EntryType = string(oss.GeneralInfo.EntryType)
		r.OperationalStatus = string(oss.GeneralInfo.OperationalStatus)
		if oss.GeneralInfo.OSSTags.Contains(osstags.PnPEnabled) {
			if oss.GeneralInfo.OSSTags.Contains(osstags.PnPEnabledIaaS) {
				r.PnPEnabled = OKIAAS
			} else {
				r.PnPEnabled = OK
			}
		} else {
			if oss.GeneralInfo.OSSTags.Contains(osstags.PnPEnabledIaaS) {
				r.PnPEnabled = NOTENABLEDIAAS
			} else {
				r.PnPEnabled = NOTENABLED
			}
		}
		if si.OSSValidation != nil {
			if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPNotEnabledButCandidate); found {
				r.PnPCandidate = PNPNOTENABLEDBUTCANDIDATE
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPNotEnabledWithoutCandidate); found {
				r.PnPCandidate = ""
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPNotEnabledMissBasicCriteria); found {
				r.PnPCandidate = ""
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPNotEnabledButInclude); found {
				r.PnPCandidate = PNPNOTENABLEDBUTINCLUDE
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPNotEnabledWithExclude); found {
				r.PnPCandidate = EXCLUDE
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPNotEnabledWithUnnecessaryExclude); found {
				r.PnPCandidate = EXCLUDEUNNEEDED
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPEnabledWithoutCandidate); found {
				r.PnPCandidate = ""
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPEnabledWithUnnecessaryCandidate); found {
				r.PnPCandidate = OKCANDIDATEUNNEEDED
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPEnabledWithInclude); found {
				r.PnPCandidate = OKINCLUDE
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.PnPEnabledWithUnnecessaryInclude); found {
				r.PnPCandidate = OKINCLUDEUNNEEDED
			} else {
				r.PnPCandidate = ERROR
			}
		} else {
			r.PnPCandidate = ERROR
		}
		r.CRNStatus = oss.GeneralInfo.OSSTags.GetCRNStatus().StringStatus()
		r.OneCloud = strconv.FormatBool(oss.GeneralInfo.OSSTags.Contains(osstags.OneCloud) || oss.GeneralInfo.OSSTags.Contains(osstags.OneCloudComponent))
		r.ClientFacingFlag = strconv.FormatBool(oss.GeneralInfo.ClientFacing)
		r.CatalogClientFacing = strconv.FormatBool(oss.CatalogInfo.CatalogClientFacing)
		//	r.ClientFacingComputed = "???"
		r.CategoryID = oss.StatusPage.CategoryID
		if oss.StatusPage.CategoryID != "" {
			r.CategoryID = oss.StatusPage.CategoryID
		} else {
			r.CategoryID = MISSING
			//issues++
		}
		if oss.StatusPage.CategoryParent != "" {
			if _, ok := ossmerge.LookupService(ossmerge.MakeComparableName(string(oss.StatusPage.CategoryParent)), false); ok {
				r.CategoryParent = string(oss.StatusPage.CategoryParent)
			} else {
				r.CategoryParent = NOTFOUND + string(oss.StatusPage.CategoryParent)
			}
		} else if si.OSSValidation.StatusCategoryCount > 1 {
			r.CategoryParent = MISSING
			//issues++
		} else {
			r.CategoryParent = ""
		}
		r.NumSameCategory = strconv.Itoa(si.OSSValidation.StatusCategoryCount)
		r.ServiceNowCategoryID = si.SourceServiceNow.StatusPage.CategoryID
		if si.HasSourceServiceNow() {
			if len(si.AdditionalServiceNow) > 0 {
				r.ServiceNowName = MULTIPLE
				//issues++
			} else if string(oss.ReferenceResourceName) != si.SourceServiceNow.CRNServiceName {
				r.ServiceNowName = MISMATCH + si.SourceServiceNow.CRNServiceName
				//issues++
			} else {
				r.ServiceNowName = OK
			}
			if oss.Compliance.ServiceNowOnboarded {
				r.ServiceNowOnboarded = OK
			} else {
				r.ServiceNowOnboarded = NO
				//issues++
			}
			if !oss.ServiceNowInfo.SupportNotApplicable {
				r.SupportEnabled = OK
			} else {
				r.SupportEnabled = NO
				//issues++ // not part of default rules
			}
			if !oss.ServiceNowInfo.OperationsNotApplicable {
				r.OperationsEnabled = OK
			} else {
				r.OperationsEnabled = NO
				//issues++ // not part of default rules
			}
		} else {
			r.ServiceNowName = MISSING
			r.ServiceNowOnboarded = ""
			r.SupportEnabled = ""
			r.OperationsEnabled = ""
			//issues++
		}
		if cat != nil && cat.Kind == "service" {
			r.CatalogHasDeployments = OK
		} else {
			r.CatalogHasDeployments = NO
			//issues++
		}
		r.OSSOnboardingPhase = string(oss.GeneralInfo.OSSOnboardingPhase)
		r.OSSTags = oss.GeneralInfo.OSSTags.String()
		//		r.NumStatusIssues = strconv.Itoa(issues)
		if si.OSSValidation != nil {
			allIssues := si.OSSValidation.GetIssues([]ossvalidation.Tag{ossvalidation.TagStatusPage})
			count := 0
			for _, v := range allIssues {
				sev := v.GetSeverity()
				if sev == ossvalidation.INFO {
					continue
				}
				r1 := issuesRecordType{}
				r1.CRNServiceName = r.CRNServiceName
				r1.PnPEnabled = r.PnPEnabled
				r1.PnPCandidate = r.PnPCandidate
				r1.Severity = string(sev)
				r1.Issue = v.GetText()
				issuesData = append(issuesData, r1)
				count++
			}
			r.NumStatusIssues = strconv.Itoa(count)
		} else {
			r.NumStatusIssues = "???"
		}
		if si.OSSMergeControl != nil {
			r.Notes = si.OSSMergeControl.Notes
		}
		r.Segment = si.Ownership.SegmentName
		r.Tribe = si.Ownership.TribeName
		mainData = append(mainData, r)
	})
	if err != nil {
		return err
	}

	sort.Slice(mainData, func(i, j int) bool {
		return mainData[i].(mainRecordType).CRNServiceName < mainData[j].(mainRecordType).CRNServiceName
	})

	sort.Slice(issuesData, func(i, j int) bool {
		ri := issuesData[i].(issuesRecordType)
		rj := issuesData[j].(issuesRecordType)
		if ri.CRNServiceName == rj.CRNServiceName {
			return ri.Issue < rj.Issue
		}
		return ri.CRNServiceName < rj.CRNServiceName
	})

	xl := ossreports.CreateExcel(w, "Services4PnP")
	err = xl.AddSheet(mainData, 90.0, "Services + Components")
	if err != nil {
		return err
	}
	err = xl.AddSheet(issuesData, 40.0, "Issues")
	if err != nil {
		return err
	}
	err = xl.Finalize()
	if err != nil {
		return err
	}

	return nil
}

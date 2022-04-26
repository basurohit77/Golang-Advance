package supportreport

// TODO: Should reuse code from AllOSSEntries report instead of copying it

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/supportcenter"

	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

const emptycell = " "
const yescell = "Yes"
const truecell = "True"
const mismatchcell = "*MISMATCH*"
const notfoundcell = "*NOTFOUND*"

const osscatviewer = "https://osscatviewer.us-south.cf.test.appdomain.cloud"
const scorecard = "https://cloud.ibm.com/scorecard"

type serviceRecord struct {
	CRNServiceName           ossreports.ExcelLink `column:"CRN service-name,20"`
	IsSupportCenterCandidate string               `column:"Support Center Candidates List,7"`
	SupportCenterHandling    string               `column:"Support Center Handling,10"`
	SupportCenterNotes       string               `column:"Support Center Notes,20"`
	DisplayName              string               `column:"Display Name,20"`
	EntryType                string               `column:"Type"`
	OperationalStatus        string               `column:"Status,10"`
	ClientFacing             string               `column:"Client Facing,7"`
	CatalogClientFacing      string               `column:"Client Facing (from Catalog only),7"`
	InServiceNow             string               `column:"In ServiceNow,7"`
	SupportNotApplicable     string               `column:"Support Not Applicable,7"`
	CRNStatus                string               `column:"CRN Status,8,center"`
	OverallStatus            string               `column:"Overall Status,8,center"`
	NumMajorIssues           int                  `column:"Number of Major Issues,8,center"`
	NumWarningIssues         int                  `column:"Number of Warnings,8,center"`
	NumMinorIssues           int                  `column:"Number of Minor Issues,8,center"`
	OneCloudStatus           string               `column:"One Cloud,15"`
	SegmentName              string               `column:"Segment Name,15"`
	TribeName                string               `column:"Tribe Name,15"`
	OfferingManager          string               `column:"Offering Manager"`
	OnboardingContact        string               `column:"Onboarding Contact"`
	InCatalog                string               `column:"In Main Catalog,7"`
	InScorecardV1            string               `column:"In ScorecardV1,7"`
	InIAM                    string               `column:"In IAM,7"`
	InEDB                    ossreports.ExcelLink `column:"Has data in EDB,7"`
	OSSUID                   string               `column:"OSS UID"`
	OSSOnboardingPhase       string               `column:"OSS Onboarding Phase,12"`
	OSSTags                  string               `column:"OSS Tags"`
}

// RunReport generates a summary report for all OSS records in the Global Catalog (as an Excel spreadsheet)
func RunReport(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	// Create the report
	xl := ossreports.CreateExcel(w, "SupportCenter")
	var err error

	if ossrunactions.Services.IsEnabled() {
		err = buildServicesSheet(xl, pattern)
		if err != nil {
			return err
		}
	}

	// Finalize/close the report
	err = xl.Finalize()
	if err != nil {
		return err
	}

	return nil
}

func buildServicesSheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	// Prepare Services sheet
	allServices := make([]interface{}, 0, 400)
	err := ossmerge.ListAllServices(pattern, func(si *ossmerge.ServiceInfo) {
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		si1 := supportcenter.ServiceInfo{Model: si}
		sci := si1.GetSupportCenterInfo(nil)
		var candidate *supportcenter.Candidate
		if sci != nil {
			candidate = sci.Candidate
		}
		if (!si.IsValid() || si.IsDeletable()) && candidate == nil {
			return
		}
		r := serviceRecord{}
		allServices = append(allServices, &r)
		if si.IsValid() {
			oss := &si.OSSService
			if candidate != nil {
				r.IsSupportCenterCandidate = yescell
				if candidate.CRNServiceName != string(oss.ReferenceResourceName) {
					debug.PrintError(`SupportCenter report: found candidate with non-canonical CRNServiceName - candidate="%s"   canonical="%s"`, candidate.CRNServiceName, oss.ReferenceResourceName)
					r.SupportCenterNotes = fmt.Sprintf(`*** non canonical name: %s`, candidate.CRNServiceName)
				}
			}
			issues := si.OSSValidation.CountIssues(nil)
			r.CRNServiceName = ossreports.ExcelLink{
				URL:  fmt.Sprintf("%s/view/%s", osscatviewer, oss.ReferenceResourceName),
				Text: string(oss.ReferenceResourceName),
			}
			r.DisplayName = oss.ReferenceDisplayName
			r.EntryType = string(oss.GeneralInfo.EntryType)
			r.OperationalStatus = string(oss.GeneralInfo.OperationalStatus)
			r.OneCloudStatus = oss.GeneralInfo.OSSTags.GetTagByGroup(osstags.GroupOneCloud).StringStatusShort()
			r.CRNStatus = oss.GeneralInfo.OSSTags.GetCRNStatus().StringStatusShort()
			r.OverallStatus = oss.GeneralInfo.OSSTags.GetOverallStatus().StringStatusShort()
			r.NumMajorIssues = issues[ossvalidation.CRITICAL] + issues[ossvalidation.SEVERE]
			r.NumWarningIssues = issues[ossvalidation.WARNING]
			r.NumMinorIssues = issues[ossvalidation.MINOR]
			r.SegmentName = oss.Ownership.SegmentName
			r.TribeName = oss.Ownership.TribeName
			r.OfferingManager = oss.Ownership.OfferingManager.String()
			r.OnboardingContact = oss.Compliance.OnboardingContact.String()
			if oss.GeneralInfo.ClientFacing {
				r.ClientFacing = yescell
			} else {
				r.ClientFacing = emptycell
			}
			if oss.CatalogInfo.CatalogClientFacing {
				r.CatalogClientFacing = yescell
			} else {
				r.CatalogClientFacing = emptycell
			}
			if si.HasSourceMainCatalog() {
				if ossmerge.CompareCompositeAndCanonicalName(si.GetSourceMainCatalog().Name, oss.ReferenceResourceName) {
					r.InCatalog = yescell
				} else {
					r.InCatalog = mismatchcell
				}
			}
			if si.HasSourceServiceNow() {
				if si.GetSourceServiceNow().CRNServiceName == string(oss.ReferenceResourceName) {
					r.InServiceNow = yescell
				} else {
					r.InServiceNow = mismatchcell
				}
			}
			if oss.ServiceNowInfo.SupportNotApplicable {
				r.SupportNotApplicable = truecell
			}
			if si.HasSourceScorecardV1Detail() {
				if si.GetSourceScorecardV1Detail().Name == string(oss.ReferenceResourceName) {
					r.InScorecardV1 = yescell
				} else {
					r.InScorecardV1 = mismatchcell
				}
			}
			if si.HasSourceIAM() {
				if ossmerge.CompareCompositeAndCanonicalName(si.GetSourceIAM().Name, oss.ReferenceResourceName) {
					r.InIAM = yescell
				} else {
					r.InIAM = mismatchcell
				}
			}
			if len(oss.MonitoringInfo.Metrics) > 0 {
				if len(oss.Ownership.SegmentName) > 100 || len(oss.Ownership.TribeName) > 100 || strings.Contains(oss.Ownership.SegmentName, "/") || strings.Contains(oss.Ownership.TribeName, "/") {
					r.InEDB = ossreports.ExcelLink{
						Text: yescell,
					}
				} else {
					r.InEDB = ossreports.ExcelLink{
						URL:  fmt.Sprintf("%s/resources/%s/%s/%s", scorecard, url.PathEscape(oss.Ownership.SegmentName), url.PathEscape(oss.Ownership.TribeName), oss.ReferenceResourceName),
						Text: yescell,
					}
				}
			} else {
				r.InEDB = ossreports.ExcelLink{
					Text: emptycell,
				}
			}
			r.OSSUID = oss.ProductInfo.OSSUID
			r.OSSOnboardingPhase = string(oss.GeneralInfo.OSSOnboardingPhase)
			r.OSSTags = oss.GeneralInfo.OSSTags.String()
		} else if candidate != nil {
			r.IsSupportCenterCandidate = yescell
			r.CRNServiceName = ossreports.ExcelLink{
				URL:  "",
				Text: candidate.CRNServiceName,
			}
			r.DisplayName = candidate.DisplayName
		}
	})
	if err != nil {
		return err
	}
	sort.Slice(allServices, func(i, j int) bool {
		return allServices[i].(*serviceRecord).CRNServiceName.Text < allServices[j].(*serviceRecord).CRNServiceName.Text
	})
	err = xl.AddSheet(allServices, 75.0, "Services+Components")
	if err != nil {
		return err
	}
	return nil
}

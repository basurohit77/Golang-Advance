package ownership

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

const emptycell = " "
const yescell = "Yes"
const mismatchcell = "*MISMATCH*"
const notfoundcell = "*NOTFOUND*"

const osscatviewer = "https://osscatviewer.us-south.cf.test.appdomain.cloud"
const scorecard = "https://cloud.ibm.com/scorecard"

type recordType struct {
	CRNServiceName     ossreports.ExcelLink `column:"CRN service-name,15"`
	CatalogName        string               `column:"Catalog Name\n(if not CRN),15"`
	DisplayName        string               `column:"Display Name,20"`
	EntryType          string               `column:"Type"`
	OperationalStatus  string               `column:"Status,6"`
	Sources            string               `column:"Found In,15"`
	SegmentName        string               `column:"Segment"`
	SegmentOwner       string               `column:"Segment Owner"`
	TribeName          string               `column:"Tribe"`
	TribeOwner         string               `column:"Tribe Owner"`
	OfferingManager    string               `column:"Offering Manager"`
	OnboardingContact  string               `column:"Onboarding Contact"`
	DevelopmentManager string               `column:"Development Manager"`
	//	TechnicalContact       string               `column:"Technical Contact"`
	ArchitectureFocal      string               `column:"Architecture Focal"`
	SupportManager         string               `column:"Support Manager"`
	OperationsManager      string               `column:"Operations Manager"`
	CatalogProviderName    string               `column:"Provider,8"`
	CatalogProviderContact string               `column:"Provider Contact"`
	Division               string               `column:"Division,8"`
	MajorUnitUTL10         string               `column:"Major Unit UTL10,15"`
	MinorUnitUTL15         string               `column:"Minor Unit UTL15,15"`
	MarketUTL17            string               `column:"Market UTL17,15"`
	PortfolioUTL20         string               `column:"Portfolio UTL20,15"`
	OfferingUTL30          string               `column:"Offering UTL30,15"`
	OwnershipMissing       string               `column:"Ownership Missing"`
	ClientFacing           string               `column:"Client Facing,7"`
	InEDB                  ossreports.ExcelLink `column:"Has data in EDB,7"`
	OSSOnboardingPhase     string               `column:"OSS Onboarding Phase,12"`
	OSSTags                string               `column:"OSS Tags"`
}

type recordType2 struct {
	CRNServiceName    ossreports.ExcelLink `column:"CRN service-name,15"`
	CatalogName       string               `column:"Catalog Name\n(if not CRN),15"`
	DisplayName       string               `column:"Display Name,20"`
	EntryType         string               `column:"Type"`
	OperationalStatus string               `column:"Status,6"`
	Sources           string               `column:"Found In,15"`
	OnboardingContact string               `column:"Onboarding Contact"`
	SegmentName       string               `column:"Segment"`
	SegmentOwner      string               `column:"Segment Owner"`
	TribeName         string               `column:"Tribe"`
	TribeOwner        string               `column:"Tribe Owner"`
	Division          string               `column:"Division,8"`
	OSSTags           string               `column:"OSS Tags"`
}

// RunReportXL generates a summary report of ownership information for all services/components in the Global Catalog (as an Excel spreadsheet)
func RunReportXL(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	var err error
	data := make([]interface{}, 0, 500)
	dataNoOwnerIBM := make([]interface{}, 0, 500)
	dataNoOwner3P := make([]interface{}, 0, 500)

	var handler = func(si *ossmerge.ServiceInfo) {
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		cat := si.GetSourceMainCatalog()
		if cat != nil && cat.Group {
			return
		}
		oss := &si.OSSService
		r := recordType{}
		r.CRNServiceName = ossreports.ExcelLink{
			URL:  fmt.Sprintf("%s/view/%s", osscatviewer, oss.ReferenceResourceName),
			Text: string(oss.ReferenceResourceName),
		}
		if len(si.AdditionalMainCatalog) > 0 {
			r.CatalogName = "***MULTIPLE***"
		} else if cat == nil {
			r.CatalogName = "<none>"
		} else if string(oss.ReferenceResourceName) != cat.Name {
			r.CatalogName = cat.Name
		}
		r.DisplayName = oss.ReferenceDisplayName
		r.EntryType = string(oss.GeneralInfo.EntryType)
		r.OperationalStatus = string(oss.GeneralInfo.OperationalStatus)
		if si.OSSValidation == nil {
			debug.PrintError("OwnershipXL report found ServiceInfo entry with nil OSSValidation: %s", si.String())
		} else {
			r.Sources = fmt.Sprint(si.OSSValidation.AllSources())
			if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.NoValidOwnership); found {
				r.OwnershipMissing = "*No owner*"
			}
		}
		r.SegmentName = oss.Ownership.SegmentName
		r.SegmentOwner = oss.Ownership.SegmentOwner.String()
		r.TribeName = oss.Ownership.TribeName
		r.TribeOwner = oss.Ownership.TribeOwner.String()
		r.OfferingManager = oss.Ownership.OfferingManager.String()
		r.OnboardingContact = oss.Compliance.OnboardingContact.String()
		r.DevelopmentManager = oss.Ownership.DevelopmentManager.String()
		//		r.TechnicalContact = oss.Ownership.ArchitectureFocal.String()
		r.ArchitectureFocal = oss.Compliance.ArchitectureFocal.String()
		r.SupportManager = oss.Support.Manager.String()
		r.OperationsManager = oss.Operations.Manager.String()
		r.CatalogProviderName = oss.CatalogInfo.Provider.String()
		r.CatalogProviderContact = fmt.Sprintf("%s %s %s", oss.CatalogInfo.ProviderContact, oss.CatalogInfo.ProviderSupportEmail, oss.CatalogInfo.ProviderPhone)
		r.Division = oss.ProductInfo.Division
		r.MajorUnitUTL10 = oss.ProductInfo.Taxonomy.MajorUnitUTL10
		r.MinorUnitUTL15 = oss.ProductInfo.Taxonomy.MinorUnitUTL15
		r.MarketUTL17 = oss.ProductInfo.Taxonomy.MarketUTL17
		r.PortfolioUTL20 = oss.ProductInfo.Taxonomy.PortfolioUTL20
		r.OfferingUTL30 = oss.ProductInfo.Taxonomy.OfferingUTL30
		if oss.GeneralInfo.ClientFacing {
			r.ClientFacing = yescell
		} else {
			r.ClientFacing = emptycell
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
		r.OSSOnboardingPhase = string(oss.GeneralInfo.OSSOnboardingPhase)
		r.OSSTags = oss.GeneralInfo.OSSTags.String()
		data = append(data, r)
	}

	err = ossmerge.ListAllServices(pattern, handler)
	if err != nil {
		return err
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].(recordType).CRNServiceName.Text < data[j].(recordType).CRNServiceName.Text
	})

	// Go through all rows and pick the ones with missing owners, to put on the second and third tabs.
	// We do this after the sort, so that we do not have to sort a second time
	for _, r0 := range data {
		r1 := r0.(recordType)
		if r1.OwnershipMissing != "" {
			r2 := recordType2{}
			r2.CRNServiceName = r1.CRNServiceName
			r2.CatalogName = r1.CatalogName
			r2.DisplayName = r1.DisplayName
			r2.EntryType = r1.EntryType
			r2.OperationalStatus = r1.OperationalStatus
			r2.Sources = r1.Sources
			r2.OnboardingContact = r1.OnboardingContact
			r2.SegmentName = r1.SegmentName
			r2.SegmentOwner = r1.SegmentOwner
			r2.TribeName = r1.TribeName
			r2.TribeOwner = r1.TribeOwner
			r2.Division = r1.Division
			r2.OSSTags = r1.OSSTags
			if r2.OperationalStatus == string(ossrecord.THIRDPARTY) || r2.OperationalStatus == string(ossrecord.COMMUNITY) {
				dataNoOwner3P = append(dataNoOwner3P, r2)
			} else {
				dataNoOwnerIBM = append(dataNoOwnerIBM, r2)
			}
		}
	}

	xl := ossreports.CreateExcel(w, "Ownership")

	err = xl.AddSheet(data, 50.0, "Ownership")
	if err != nil {
		return err
	}

	err = xl.AddSheet(dataNoOwnerIBM, 50.0, "No Owner - IBM")
	if err != nil {
		return err
	}

	err = xl.AddSheet(dataNoOwner3P, 50.0, "No Owner - Third-Party")
	if err != nil {
		return err
	}

	err = xl.Finalize()
	if err != nil {
		return err
	}

	return nil
}

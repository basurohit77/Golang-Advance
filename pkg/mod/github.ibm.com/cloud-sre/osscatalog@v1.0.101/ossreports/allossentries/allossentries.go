package allossentries

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/stats"

	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

const emptycell = " "
const yescell = "Yes"
const disabledcell = "Yes/disabled"
const mismatchcell = "*MISMATCH*"
const notfoundcell = "*NOTFOUND*"

const osscatviewer = "https://osscatviewer.us-south.cf.test.appdomain.cloud"
const scorecard = "https://cloud.ibm.com/scorecard"

type serviceRecord struct {
	CRNServiceName      ossreports.ExcelLink `column:"CRN service-name,15"`
	DisplayName         string               `column:"Display Name,20"`
	EntryType           string               `column:"Type"`
	OperationalStatus   string               `column:"Status,10"`
	OSSOnboardingPhase  string               `column:"OSS Onboarding Phase,12"`
	OneCloudStatus      string               `column:"One Cloud,15"`
	CRNStatus           string               `column:"CRN Status,8,center"`
	OverallStatus       string               `column:"Overall Status,8,center"`
	NumMajorIssues      int                  `column:"Number of Major Issues,8,center"`
	NumWarningIssues    int                  `column:"Number of Warnings,8,center"`
	NumMinorIssues      int                  `column:"Number of Minor Issues,8,center"`
	SegmentName         string               `column:"Segment Name,15"`
	TribeName           string               `column:"Tribe Name,15"`
	OfferingManager     string               `column:"Offering Manager"`
	OnboardingContact   string               `column:"Onboarding Contact"`
	ClientFacing        string               `column:"Client Facing,7"`
	CatalogClientFacing string               `column:"Client Facing (from Catalog only),7"`
	InCatalog           string               `column:"In Main Catalog,7"`
	InRMC               string               `column:"In RMC,7"`
	InServiceNow        string               `column:"In ServiceNow,7"`
	InScorecardV1       string               `column:"In ScorecardV1,7"`
	InIAM               string               `column:"In IAM,7"`
	InEDB               ossreports.ExcelLink `column:"Has data in EDB,7"`
	SupportEnabled      string               `column:"Support Enabled,7"`
	OperationsEnabled   string               `column:"Operations Enabled,7"`
	CatalogCreated      string               `column:"Catalog Entry Created"`
	OSSCreated          string               `column:"OSS Entry Created"`
	OSSUpdated          string               `column:"OSS Entry Updated"`
	RMCType             string               `column:"RMC Type"`
	RMCMaturity         string               `column:"RMC Maturity"`
	RMCOneCloud         string               `column:"RMC One Cloud"`
	RMCManagedBy        string               `column:"RMC Managed By"`
	OSSUID              string               `column:"OSS UID"`
	OSSTags             string               `column:"OSS Tags"`
}

type tribeRecord struct {
	SegmentName               ossreports.ExcelLink `column:"Segment Name,25"`
	TribeName                 ossreports.ExcelLink `column:"Tribe Name,25"`
	SegmentType               string               `column:"Segment Type,20"`
	SegmentOwner              string               `column:"Segment Owner,20"`
	SegmentTechnicalContact   string               `column:"Segment Technical Contact,20"`
	TribeOwner                string               `column:"Tribe Owner,20"`
	CRBApprovers              string               `column:"Number of CRB Approvers"`
	TribeID                   string               `column:"Tribe ID"`
	InScorecardV1             string               `column:"In ScorecardV1,7"`
	SegmentOSSOnboardingPhase string               `column:"Segment OSS Onboarding Phase,12"`
	SegmentOSSTags            string               `column:"Segment OSS Tags"`
	SegmentOSSCreated         string               `column:"Segment OSS Entry Created"`
	SegmentOSSUpdated         string               `column:"Segment OSS Entry Updated"`
	TribeOSSOnboardingPhase   string               `column:"Tribe OSS Onboarding Phase,12"`
	TribeOSSTags              string               `column:"Tribe OSS Tags"`
	TribeOSSCreated           string               `column:"Tribe OSS Entry Created"`
	TribeOSSUpdated           string               `column:"Tribe OSS Entry Updated"`
}

type environmentRecord struct {
	EnvironmentID       ossreports.ExcelLink `column:"Environment ID,35"`
	DisplayName         string               `column:"Display Name,20"`
	EnvironmentType     string               `column:"Type,20"`
	EnvironmentStatus   string               `column:"Status,10"`
	OwningSegmentName   string               `column:"Owning Segment,20"`
	OwningClient        string               `column:"Owning Client,20"`
	InCatalog           string               `column:"In Main Catalog,7"`
	InDoctorEnvironment string               `column:"In Doctor - Environment Records,7"`
	InDoctorRegionID    string               `column:"In Doctor - RegionID Records,7"`
	OSSOnboardingPhase  string               `column:"OSS Onboarding Phase,12"`
	OSSTags             string               `column:"OSS Tags"`
	OSSCreated          string               `column:"OSS Entry Created"`
	OSSUpdated          string               `column:"OSS Entry Updated"`
}

type validationIssueRecord struct {
	Count    int    `column:"Count,center"`
	Severity string `column:"Severity"`
	Tags     string `column:"Tags,30"`
	Title    string `column:"Issue Title,100"`
	OSSType  string `column:"OSS Record Type,20"`
}

type validationSummaryRecord struct {
	Label          string                 `column:"Category,35"`
	Total          int                    `column:"Total,10,center"`
	CRNGreen       ossreports.NonNegative `column:"CRN Status: Green,8,center"`
	CRNYellow      ossreports.NonNegative `column:"CRN Status: Yellow,8,center"`
	CRNRed         ossreports.NonNegative `column:"CRN Status: Red,8,center"`
	CRNUnknown     ossreports.NonNegative `column:"CRN Status: Unknown,8,center"`
	OverallGreen   ossreports.NonNegative `column:"Overall Status: Green,8,center"`
	OverallYellow  ossreports.NonNegative `column:"Overall Status: Yellow,8,center"`
	OverallRed     ossreports.NonNegative `column:"Overall Status: Red,8,center"`
	OverallUnknown ossreports.NonNegative `column:"Overall Status: Unknown,8,center"`
}

// RunReport generates a summary report for all OSS records in the Global Catalog (as an Excel spreadsheet)
func RunReport(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	// Create the report
	xl := ossreports.CreateExcel(w, "AllOSSEntries")
	var err error

	if ossrunactions.Services.IsEnabled() {
		err = buildServicesSheet(xl, pattern)
		if err != nil {
			return err
		}
	}

	if ossrunactions.Tribes.IsEnabled() {
		err = buildTribesSheet(xl, pattern)
		if err != nil {
			return err
		}
	}

	if ossrunactions.Environments.IsEnabled() {
		err = buildEnvironmentsSheet(xl, pattern)
		if err != nil {
			return err
		}
	}

	err = buildValidationSummarySheet(xl, pattern)
	if err != nil {
		return err
	}

	err = buildValidationIssuesSheet(xl, pattern)
	if err != nil {
		return err
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
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		if si.OSSValidation != nil {
			recordValidationIssues(si.OSSValidation.GetIssues(nil), "OSSService")
		}
		oss := &si.OSSService
		issues := si.OSSValidation.CountIssues(nil)
		r := serviceRecord{}
		allServices = append(allServices, &r)
		r.CRNServiceName = ossreports.ExcelLink{
			URL:  fmt.Sprintf("%s/view/%s", osscatviewer, oss.ReferenceResourceName),
			Text: string(oss.ReferenceResourceName),
		}
		r.DisplayName = oss.ReferenceDisplayName
		r.EntryType = string(oss.GeneralInfo.EntryType)
		r.OperationalStatus = string(oss.GeneralInfo.OperationalStatus)
		r.OSSOnboardingPhase = string(oss.GeneralInfo.OSSOnboardingPhase)
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
			r.CatalogCreated = si.GetSourceMainCatalog().Created
		}
		if si.HasSourceRMC() {
			if si.GetSourceRMC().CRNServiceName == oss.ReferenceResourceName {
				r.InRMC = yescell
			} else {
				r.InRMC = mismatchcell
			}
		}
		if si.HasSourceServiceNow() {
			if si.GetSourceServiceNow().CRNServiceName == string(oss.ReferenceResourceName) {
				r.InServiceNow = yescell
			} else {
				r.InServiceNow = mismatchcell
			}
			if !oss.ServiceNowInfo.SupportNotApplicable {
				r.SupportEnabled = yescell
			}
			if !oss.ServiceNowInfo.OperationsNotApplicable {
				r.OperationsEnabled = yescell
			}
		}
		if si.HasSourceScorecardV1Detail() {
			if si.GetSourceScorecardV1Detail().Name == string(oss.ReferenceResourceName) {
				r.InScorecardV1 = yescell
			} else {
				r.InScorecardV1 = mismatchcell
			}
		} else if len(si.OSSValidation.SourceNames(ossvalidation.SCORECARDV1DISABLED)) > 0 {
			r.InScorecardV1 = disabledcell
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
		r.OSSCreated = si.Created
		r.OSSUpdated = si.Updated
		if si.HasSourceRMC() {
			r.RMCType = si.GetSourceRMC().Type
			r.RMCMaturity = si.GetSourceRMC().Maturity
			r.RMCOneCloud = strconv.FormatBool(si.GetSourceRMC().OneCloudService)
			r.RMCManagedBy = si.GetSourceRMC().ManagedBy
		}
		r.OSSUID = oss.ProductInfo.OSSUID
		r.OSSTags = oss.GeneralInfo.OSSTags.String()
	})
	if err != nil {
		return err
	}
	sort.Slice(allServices, func(i, j int) bool {
		return allServices[i].(*serviceRecord).CRNServiceName.Text < allServices[j].(*serviceRecord).CRNServiceName.Text
	})
	err = xl.AddSheet(allServices, 65.0, "Services+Components")
	if err != nil {
		return err
	}
	return nil
}

func buildTribesSheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	// Prepare Segments/Tribes sheet
	allTribes := make([]interface{}, 0, 400)
	err := ossmerge.ListAllSegments(pattern, func(seg *ossmerge.SegmentInfo) {
		if !seg.IsValid() || seg.IsDeletable() {
			return
		}
		if seg.OSSSegment.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		if seg.OSSValidation != nil {
			recordValidationIssues(seg.OSSValidation.GetIssues(nil), "OSSSegment")
		}
		var numTribes int
		err2 := seg.ListAllTribes(pattern, func(tr *ossmerge.TribeInfo) {
			if !tr.IsValid() || tr.IsDeletable() {
				return
			}
			if tr.OSSTribe.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
				return
			}
			numTribes++
			if tr.OSSValidation != nil {
				recordValidationIssues(tr.OSSValidation.GetIssues(nil), "OSSTribe")
			}
			r := tribeRecord{}
			allTribes = append(allTribes, &r)
			r.SegmentName = ossreports.ExcelLink{
				URL:  fmt.Sprintf("%s/segment/%s", osscatviewer, seg.OSSSegment.SegmentID),
				Text: seg.OSSSegment.DisplayName,
			}
			r.TribeName = ossreports.ExcelLink{
				URL:  fmt.Sprintf("%s/tribe/%s", osscatviewer, tr.OSSTribe.TribeID),
				Text: tr.OSSTribe.DisplayName,
			}
			r.SegmentType = string(seg.SegmentType)
			if seg.OSSSegment.Owner.IsValid() {
				r.SegmentOwner = seg.OSSSegment.Owner.String()
			} else {
				r.SegmentOwner = emptycell
			}
			if seg.OSSSegment.TechnicalContact.IsValid() {
				r.SegmentTechnicalContact = seg.OSSSegment.TechnicalContact.String()
			} else {
				r.SegmentTechnicalContact = emptycell
			}
			if tr.OSSTribe.Owner.IsValid() {
				r.TribeOwner = tr.OSSTribe.Owner.String()
			} else {
				r.TribeOwner = emptycell
			}
			if len(tr.OSSTribe.ChangeApprovers) > 0 {
				r.CRBApprovers = fmt.Sprintf("   %d", len(tr.OSSTribe.ChangeApprovers))
			} else {
				r.CRBApprovers = emptycell
			}
			r.TribeID = string(tr.OSSTribe.TribeID)
			if tr.HasSourceScorecardV1() {
				if tr.GetSourceScorecardV1().GetTribeID() == tr.OSSTribe.TribeID {
					r.InScorecardV1 = yescell
				} else {
					r.InScorecardV1 = mismatchcell
				}
			} else if tr.OSSValidation != nil && len(tr.OSSValidation.SourceNames(ossvalidation.SCORECARDV1DISABLED)) > 0 {
				r.InScorecardV1 = disabledcell
			}
			r.SegmentOSSOnboardingPhase = string(seg.OSSOnboardingPhase)
			r.SegmentOSSTags = seg.OSSTags.String()
			r.SegmentOSSCreated = seg.Created
			r.SegmentOSSUpdated = seg.Updated
			r.TribeOSSOnboardingPhase = string(tr.OSSOnboardingPhase)
			r.TribeOSSTags = tr.OSSTags.String()
			r.TribeOSSCreated = tr.Created
			r.TribeOSSUpdated = tr.Updated
		})
		if err2 != nil {
			debug.PrintError("Error in ListAllTribes(): %v", err2)
			panic("Unrecoverable error")
		}
		if numTribes == 0 {
			r := tribeRecord{}
			allTribes = append(allTribes, &r)
			r.SegmentName = ossreports.ExcelLink{
				URL:  fmt.Sprintf("%s/segment/%s", osscatviewer, seg.OSSSegment.SegmentID),
				Text: seg.OSSSegment.DisplayName,
			}
			r.TribeName = ossreports.ExcelLink{Text: "*NO VALID TRIBES*"}
			r.CRBApprovers = emptycell
			r.SegmentOwner = seg.OSSSegment.Owner.String()
		}
	})
	if err != nil {
		return err
	}
	sort.Slice(allTribes, func(i, j int) bool {
		left := allTribes[i].(*tribeRecord)
		right := allTribes[j].(*tribeRecord)
		if left.SegmentName == right.SegmentName {
			return left.TribeName.Text < right.TribeName.Text
		}
		return left.SegmentName.Text < right.SegmentName.Text
	})
	err = xl.AddSheet(allTribes, 65.0, "Segments+Tribes")
	if err != nil {
		return err
	}
	return nil
}

func buildEnvironmentsSheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	// Prepare Environments sheet
	allEnvironments := make([]interface{}, 0, 400)
	err := ossmerge.ListAllEnvironments(pattern, func(env *ossmerge.EnvironmentInfo) {
		if !env.IsValid() || env.IsDeletable() {
			return
		}
		if env.OSSEnvironment.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		if env.OSSValidation != nil {
			recordValidationIssues(env.OSSValidation.GetIssues(nil), "OSSEnvironment")
		}
		oss := &env.OSSEnvironment
		r := environmentRecord{}
		allEnvironments = append(allEnvironments, &r)
		r.EnvironmentID = ossreports.ExcelLink{
			URL:  fmt.Sprintf("%s/environment/%s", osscatviewer, oss.EnvironmentID),
			Text: string(oss.EnvironmentID),
		}
		r.DisplayName = oss.DisplayName
		r.EnvironmentType = string(oss.Type)
		r.EnvironmentStatus = string(oss.Status)
		if oss.OwningSegment != "" {
			if seg, found := ossmerge.LookupSegment(oss.OwningSegment, false); found {
				r.OwningSegmentName = seg.OSSSegment.DisplayName
			} else {
				r.OwningSegmentName = notfoundcell
			}
		}
		if oss.OwningClient != "" {
			r.OwningClient = oss.OwningClient
			// } else {
			//			r.OwningClient = emptycell
		}
		if env.HasSourceMainCatalog() {
			r.InCatalog = yescell
		}
		if env.HasSourceDoctorEnvironment() {
			r.InDoctorEnvironment = yescell
		} else if len(env.OSSValidation.SourceNames(ossvalidation.DOCTORENVDISABLED)) > 0 {
			r.InDoctorEnvironment = disabledcell
		}
		if env.HasSourceDoctorRegionID() {
			r.InDoctorRegionID = yescell
		} else if len(env.OSSValidation.SourceNames(ossvalidation.DOCTORREGIONIDDISABLED)) > 0 {
			r.InDoctorRegionID = disabledcell
		}
		r.OSSOnboardingPhase = string(oss.OSSOnboardingPhase)
		r.OSSTags = oss.OSSTags.String()
		r.OSSCreated = env.Created
		r.OSSUpdated = env.Updated
	})
	if err != nil {
		return err
	}
	sort.Slice(allEnvironments, func(i, j int) bool {
		return allEnvironments[i].(*environmentRecord).EnvironmentID.Text < allEnvironments[j].(*environmentRecord).EnvironmentID.Text
	})
	err = xl.AddSheet(allEnvironments, 65.0, "Environments")
	if err != nil {
		return err
	}
	return nil
}

func buildValidationSummarySheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	// Prepare Validation Summary sheet
	data := stats.GetGlobalActualStats()
	data.Finalize()
	allRows := make([]interface{}, 0, 50)

	if ossrunactions.Services.IsEnabled() {
		keys := make([]osstags.Tag, 0, len(data.ServicesOneCloudValidationStatus))
		for k := range data.ServicesOneCloudValidationStatus {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})
		for i, k := range keys {
			sc := data.ServicesOneCloudValidationStatus[k]
			r := validationSummaryRecord{}
			allRows = append(allRows, &r)
			r.Label = fmt.Sprintf("%s Services/Components", k)
			r.Total = sc.CRN.Green + sc.CRN.Yellow + sc.CRN.Red + sc.CRN.Unknown
			total2 := sc.Overall.Green + sc.Overall.Yellow + sc.Overall.Red + sc.Overall.Unknown
			if r.Total != total2 {
				panic(fmt.Sprintf("Total number of %s services by CRN status(=%d) not equal to total by Overall status(=%d)", k, r.Total, total2))
			}
			if i == 0 && r.Total != data.NumServicesActual {
				panic(fmt.Sprintf("Total number of %s services by CRN status(=%d) not equal to count by recorded actions(=%d)", k, r.Total, data.NumServicesActual))
			}
			r.CRNGreen = ossreports.NonNegative(sc.CRN.Green)
			r.CRNYellow = ossreports.NonNegative(sc.CRN.Yellow)
			r.CRNRed = ossreports.NonNegative(sc.CRN.Red)
			r.CRNUnknown = ossreports.NonNegative(sc.CRN.Unknown)
			r.OverallGreen = ossreports.NonNegative(sc.Overall.Green)
			r.OverallYellow = ossreports.NonNegative(sc.Overall.Yellow)
			r.OverallRed = ossreports.NonNegative(sc.Overall.Red)
			r.OverallUnknown = ossreports.NonNegative(sc.Overall.Unknown)
		}
	}

	if ossrunactions.Tribes.IsEnabled() {
		r1 := validationSummaryRecord{}
		allRows = append(allRows, &r1)
		r1.Label = "ALL Segments"
		r1.Total = data.NumSegmentsActual
		r1.CRNGreen = ossreports.NonNegative(-1)
		r1.CRNYellow = ossreports.NonNegative(-1)
		r1.CRNRed = ossreports.NonNegative(-1)
		r1.CRNUnknown = ossreports.NonNegative(-1)
		r1.OverallGreen = ossreports.NonNegative(-1)
		r1.OverallYellow = ossreports.NonNegative(-1)
		r1.OverallRed = ossreports.NonNegative(-1)
		r1.OverallUnknown = ossreports.NonNegative(-1)

		r2 := validationSummaryRecord{}
		allRows = append(allRows, &r2)
		r2.Label = "ALL Tribes"
		r2.Total = data.NumTribesActual
		r2.CRNGreen = ossreports.NonNegative(-1)
		r2.CRNYellow = ossreports.NonNegative(-1)
		r2.CRNRed = ossreports.NonNegative(-1)
		r2.CRNUnknown = ossreports.NonNegative(-1)
		r2.OverallGreen = ossreports.NonNegative(-1)
		r2.OverallYellow = ossreports.NonNegative(-1)
		r2.OverallRed = ossreports.NonNegative(-1)
		r2.OverallUnknown = ossreports.NonNegative(-1)
	}

	if ossrunactions.Environments.IsEnabled() {
		r := validationSummaryRecord{}
		allRows = append(allRows, &r)
		r.Label = "ALL Environments"
		r.Total = data.NumEnvironmentsActual
		r.CRNGreen = ossreports.NonNegative(-1)
		r.CRNYellow = ossreports.NonNegative(-1)
		r.CRNRed = ossreports.NonNegative(-1)
		r.CRNUnknown = ossreports.NonNegative(-1)
		r.OverallGreen = ossreports.NonNegative(-1)
		r.OverallYellow = ossreports.NonNegative(-1)
		r.OverallRed = ossreports.NonNegative(-1)
		r.OverallUnknown = ossreports.NonNegative(-1)
	}

	err := xl.AddSheet(allRows, 65.0, "Validation Summary")
	if err != nil {
		return err
	}
	return nil
}

func buildValidationIssuesSheet(xl *ossreports.ExcelReport, pattern *regexp.Regexp) error {
	// Prepare Validation Issues list sheet
	allRows := make([]interface{}, 0, len(allValidationIssues))
	buffers := make(map[string][]*validationIssueRecord)
	for _, r := range allValidationIssues {
		buffers[r.Severity] = append(buffers[r.Severity], r)
	}
	for _, s := range ossvalidation.AllSeverityList() {
		cur := buffers[string(s)]
		sort.SliceStable(cur, func(i, j int) bool {
			return cur[i].Title < cur[j].Title
		})
		for _, r := range cur {
			allRows = append(allRows, r)
		}
	}

	err := xl.AddSheet(allRows, 65.0, "All Validation Issues")
	if err != nil {
		return err
	}
	return nil
}

var allValidationIssues = make(map[string]*validationIssueRecord)

func recordValidationIssues(issues []*ossvalidation.ValidationIssue, ossType string) {
	for _, v := range issues {
		switch v.Severity {
		case ossvalidation.INFO, ossvalidation.IGNORE /*, ossvalidation.DEFERRED*/ :
			// skip
		default:
			if r, ok := allValidationIssues[v.Title]; ok {
				tags := fmt.Sprintf("%v", v.Tags)
				/* We allow the issue for ResetForRMC() to apply to any entry type
				if ossType != r.OSSType {
					debug.PrintError(`AllOSSEntries report found two validation issues with same title in different OSS record types %s/%s   title="%s"`, r.OSSType, ossType, v.Title)
				}
				*/
				if string(v.Severity) != r.Severity || tags != r.Tags {
					debug.PrintError(`AllOSSEntries report found two validation issues with same title but different severity or tags: severity=%s/%s   tags=%s/%s   title="%s"`, r.Severity, string(v.Severity), r.Tags, tags, v.Title)
				}
				r.Count++
			} else {
				r := &validationIssueRecord{
					Count:    1,
					Severity: string(v.Severity),
					Tags:     fmt.Sprintf("%v", v.Tags),
					Title:    v.Title,
					OSSType:  ossType,
				}
				allValidationIssues[v.Title] = r
			}
		}
	}
}

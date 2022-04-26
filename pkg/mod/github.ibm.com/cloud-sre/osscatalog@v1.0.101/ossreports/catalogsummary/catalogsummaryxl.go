package catalogsummary

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

const emptycell = " "
const yescell = "Yes"
const mismatchcell = "*MISMATCH*"
const hiddencell = "(***HIDDEN/INACTIVE***)"
const missingcell = "***MISSING***"
const dataunavailcell = "***DATA UNAVAILABLE***"
const multiplecell = "***MULTIPLE***"

type recordType struct {
	CRNServiceName          string                `column:"CRN service-name,15"`
	CatalogName             string                `column:"Catalog Name\n(if not CRN),15"`
	DisplayName             string                `column:"Display Name,20"`
	TileNeedsCorrection     string                `column:"Tile label requires correction"`
	EntryType               string                `column:"Type"`
	Category                string                `column:"Category,15"`
	OfferingManager         string                `column:"Offering Manager"`
	Description             string                `column:"Description,30"`
	Provider                string                `column:"Provider,8"`
	OperationalStatus       string                `column:"Status,7"`
	ClientFacing            string                `column:"Client Facing,7"`
	CatalogClientFacing     string                `column:"Client Facing (from Catalog only),7"`
	FreePlan                string                `column:"Free plan(s)"`
	Plans                   string                `column:"Plans,15"`
	Locations               string                `column:"Locations"`
	PartNumbers             string                `column:"Part Numbers,15"`
	ProductIDs              string                `column:"Product IDs,15"`
	ProductIDSource         string                `column:"Product ID Source,15"`
	Division                string                `column:"Division,8"`
	ClearingHouseReferences string                `column:"ClearingHouse References,15"`
	ClearingHouseLink       string                `column:"ClearingHouse Link,15"`
	MajorUnitUTL10          string                `column:"Major Unit UTL10,15"`
	MinorUnitUTL15          string                `column:"Minor Unit UTL15,15"`
	MarketUTL17             string                `column:"Market UTL17,15"`
	PortfolioUTL20          string                `column:"Portfolio UTL20,15"`
	OfferingUTL30           string                `column:"Offering UTL30,15"`
	CreationDate            string                `column:"Creation Date"`
	LastModifiedDate        string                `column:"Last Modification"`
	ServiceKeySupport       string                `column:"Service Key Support"`
	LongDescription         string                `column:"Long Description,30"`
	URL                     string                `column:"URL"`
	CatalogID               string                `column:"Catalog ID"`
	CHOfficialName          string                `column:"ClearingHouse Official Name,20"`
	InCatalog               string                `column:"In Main Catalog,7"`
	InServiceNow            string                `column:"In ServiceNow,7"`
	InScorecardV1           string                `column:"In ScorecardV1,7"`
	InIAM                   string                `column:"In IAM,7"`
	SupportEnabled          string                `column:"Support Enabled,8"`
	OperationsEnabled       string                `column:"Operations Enabled,8"`
	OSSUID                  string                `column:"OSS UID"`
	OSSOnboardingPhase      string                `column:"OSS Onboarding Phase,12"`
	OSSTags                 string                `column:"OSS Tags"`
	locationsList           collections.StringSet `column:"-"`
}
type planRecordType struct {
	CRNServiceName string `column:"CRN service-name,15"`
	DisplayName    string `column:"Display Name,20"`
	EntryType      string `column:"Type"`
	Status         string `column:"Status"`
	PlanName       string `column:"Plan Name"`
	PlanCatalogID  string `column:"Plan Catalog ID"`
}

var allLocations = make(map[string]struct{})

// RunReport generates a summary report for all services/components in the Global Catalog (as an Excel spreadsheet)
func RunReport(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	var err error
	data := make([]interface{}, 0, 500)
	plans := make([]interface{}, 0, 500)

	var handler = func(si *ossmerge.ServiceInfo) {
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		oss := &si.OSSService
		r := recordType{}
		r.CRNServiceName = string(oss.ReferenceResourceName)
		r.DisplayName = oss.ReferenceDisplayName
		r.EntryType = string(oss.GeneralInfo.EntryType)
		r.OfferingManager = oss.Ownership.OfferingManager.String()
		r.OperationalStatus = string(oss.GeneralInfo.OperationalStatus)
		if oss.GeneralInfo.ClientFacing {
			r.ClientFacing = "Yes"
		} else {
			r.ClientFacing = emptycell
		}
		if oss.CatalogInfo.CatalogClientFacing {
			r.CatalogClientFacing = yescell
		} else {
			r.CatalogClientFacing = emptycell
		}
		r.OSSUID = oss.ProductInfo.OSSUID
		r.OSSTags = oss.GeneralInfo.OSSTags.String()
		r.OSSOnboardingPhase = string(oss.GeneralInfo.OSSOnboardingPhase)
		if cat := si.GetSourceMainCatalog(); cat != nil {
			if len(si.AdditionalMainCatalog) > 0 {
				r.CatalogName = "***MULTIPLE***"
			} else if string(oss.ReferenceResourceName) != cat.Name {
				r.CatalogName = cat.Name
			}
			r.Category = oss.CatalogInfo.CategoryTags
			r.Description = cat.OverviewUI.En.Description
			r.Provider = cat.Provider.Name
			r.FreePlan = catalog.ScanFreePricing(cat)
			if cex := si.GetCatalogExtra(false); cex != nil {
				if len(cex.Plans) > 0 {
					r.Plans = si.ListPlans()
					for pi := range cex.Plans {
						p := planRecordType{}
						p.CRNServiceName = r.CRNServiceName
						p.DisplayName = r.DisplayName
						p.EntryType = r.EntryType
						p.Status = string(si.GeneralInfo.OperationalStatus)
						p.PlanName = pi.Name
						p.PlanCatalogID = pi.CatalogID
						plans = append(plans, p)
					}
				} else {
					r.Plans = "***NOTFOUND***"
				}
				if ossrunactions.Deployments.IsEnabled() {
					r.locationsList = cex.Locations
					if cex.Locations.Len() > 0 {
						r.Locations = cex.Locations.String()
						for _, loc := range cex.Locations.Slice() {
							allLocations[loc] = struct{}{}
						}
					} else {
						r.Locations = "***NOTFOUND***"
					}
				} else {
					r.Locations = "***NOTCOMPUTED***"
				}
			} else {
				if cat.Kind == "service" || cat.Kind == "iaas" {
					r.Plans = "***NOTFOUND***"
					r.Locations = "***NOTFOUND***"
				}
				// } else {
				// ignore
				//}
			}
			r.CreationDate = cat.Created
			r.LastModifiedDate = cat.Updated
			r.LongDescription = cat.OverviewUI.En.LongDescription
			r.CatalogID = cat.ID
		} else if si.IgnoredMainCatalog != nil {
			r.CatalogName = si.IgnoredMainCatalog.Name + hiddencell
			r.InCatalog = hiddencell
		} else {
			r.CatalogName = missingcell
			// r.InCatalog = missingcell
		}
		r.TileNeedsCorrection = "???"
		r.URL = "???"
		r.ServiceKeySupport = "???"
		r.PartNumbers = fmt.Sprintf("%v", oss.ProductInfo.PartNumbers)
		r.ProductIDs = fmt.Sprintf("%v", oss.ProductInfo.ProductIDs)
		r.ProductIDSource = string(oss.ProductInfo.ProductIDSource)
		r.Division = oss.ProductInfo.Division
		if len(oss.ProductInfo.ClearingHouseReferences) > 0 {
			chrefs := oss.ProductInfo.ClearingHouseReferences
			buf := strings.Builder{}
			for i := range chrefs {
				buf.WriteString(clearinghouse.MakeCHLabel(chrefs[i].Name, clearinghouse.DeliverableID(chrefs[i].ID)))
			}
			r.ClearingHouseReferences = buf.String()
			r.ClearingHouseLink = clearinghouse.GetCHEntryUI(clearinghouse.DeliverableID(chrefs[0].ID))
		}
		r.MajorUnitUTL10 = oss.ProductInfo.Taxonomy.MajorUnitUTL10
		r.MinorUnitUTL15 = oss.ProductInfo.Taxonomy.MinorUnitUTL15
		r.MarketUTL17 = oss.ProductInfo.Taxonomy.MarketUTL17
		r.PortfolioUTL20 = oss.ProductInfo.Taxonomy.PortfolioUTL20
		r.OfferingUTL30 = oss.ProductInfo.Taxonomy.OfferingUTL30
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
		}
		if si.HasSourceIAM() {
			if ossmerge.CompareCompositeAndCanonicalName(si.GetSourceIAM().Name, oss.ReferenceResourceName) {
				r.InIAM = yescell
			} else {
				r.InIAM = mismatchcell
			}
		}
		if clearinghouse.HasCHInfo() {
			switch len(oss.ProductInfo.ClearingHouseReferences) {
			case 0:
				r.CHOfficialName = emptycell
			case 1:
				if entry, err := clearinghouse.GetFullRecordByID(clearinghouse.DeliverableID(oss.ProductInfo.ClearingHouseReferences[0].ID)); err == nil {
					r.CHOfficialName = entry.OfficialName
				} else {
					r.CHOfficialName = dataunavailcell
				}
			default:
				r.CHOfficialName = multiplecell
			}
		} else {
			r.CHOfficialName = dataunavailcell
		}
		data = append(data, r)
	}

	// Generate all the data
	err = ossmerge.ListAllServices(pattern, handler)
	if err != nil {
		return err
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].(recordType).CRNServiceName < data[j].(recordType).CRNServiceName
	})

	// Create the report
	xl := ossreports.CreateExcel(w, "CatalogSummaryXL")

	// Create the first sheet, listing all services
	err = xl.AddSheet(data, 65.0, "Services")
	if err != nil {
		return err
	}

	// Create the second sheet, listing all locations
	if ossrunactions.Deployments.IsEnabled() {
		allLocationsList := make([]string, 0, len(allLocations))
		for loc := range allLocations {
			allLocationsList = append(allLocationsList, loc)
		}
		allLocationsList = ossmerge.SortLocationsList(allLocationsList)
		locHeaders := append([]string{"CRN service-name,15", "Display Name,20", "Type", "Status,15"}, allLocationsList...)
		locSheet, err := xl.AddEmptySheet(locHeaders, 30.0, "Locations")
		if err != nil {
			return debug.WrapError(err, "Error while creating a the Locations sheet")
		}
		for _, r := range data {
			r0 := r.(recordType)
			row := make([]interface{}, 0, len(locHeaders))
			row = append(row, r0.CRNServiceName)
			row = append(row, r0.DisplayName)
			row = append(row, r0.EntryType)
			row = append(row, r0.OperationalStatus)
			for _, loc := range allLocationsList {
				if r0.locationsList != nil && r0.locationsList.Contains(loc) {
					row = append(row, "X")
				} else {
					row = append(row, "")
				}
			}
			err := xl.AppendRowSlice(locSheet, row)
			if err != nil {
				return debug.WrapError(err, "Error while appending a row to the Locations sheet")
			}
		}
	} else {
		locHeaders := []string{`Locations information not available: -run Deployments not enabled in this run`}
		_, err = xl.AddEmptySheet(locHeaders, 30.0, "Locations UNAVAILABLE")
		if err != nil {
			return debug.WrapError(err, "Error while creating a the Locations sheet")
		}
	}

	// Create the third sheet listing all the Plans
	sort.Slice(plans, func(i, j int) bool {
		if plans[i].(planRecordType).CRNServiceName == plans[j].(planRecordType).CRNServiceName {
			if plans[i].(planRecordType).PlanName == plans[j].(planRecordType).PlanName {
				return plans[i].(planRecordType).PlanCatalogID < plans[j].(planRecordType).PlanCatalogID
			}
			return plans[i].(planRecordType).PlanName < plans[j].(planRecordType).PlanName
		}
		return plans[i].(planRecordType).CRNServiceName < plans[j].(planRecordType).CRNServiceName
	})
	err = xl.AddSheet(plans, 65.0, "Plans")
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

/*
func getPlans(r *catalogapi.Resource) string {
	// TODO: Combine Pricing tags from all Individual Plans with the Pricing Tags from the parent entry
	var result strings.Builder
	var count = 0
	err := catalog.ListPlans(r, func(rp *catalogapi.Resource) {
		if count > 0 {
			result.WriteString(`,`)
		}
		result.WriteString(fmt.Sprintf(`"%s"`, rp.OverviewUI.En.DisplayName))
		count++
	})
	if err != nil {
		debug.PrintError("CatalogSummaryXL: error getting Plans for resource %s: %v", r.Name, err)
		return "***error***"
	}

	return result.String()
}
*/

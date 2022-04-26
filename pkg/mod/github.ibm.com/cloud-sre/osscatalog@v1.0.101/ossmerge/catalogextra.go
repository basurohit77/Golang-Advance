package ossmerge

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// PlanInfo captures the information associated with one Plan object found in the Catalog
type PlanInfo struct {
	Name      string
	CatalogID string
}

// CatalogExtra captures any extra informatiom loaded from the Global Catalog, that is not contained in the catalog.Resource (ServiceInfo.SourceMainCatalog) itself
// e.g. pricing, plans, deployments, etc.
type CatalogExtra struct {
	PartNumbers   collections.StringSet // Part numbers obtained from the Main Catalog entries during loading
	Plans         map[PlanInfo]struct{} // All the Plans associated with this Main Catalog entry
	Locations     collections.StringSet // All locations from Deployments associated with this Main Catalog entry
	ChildrenKinds collections.StringSet // All Catalog "kind" attributes for direct children
}

// AddPlan adds a new Plan entry to the CatalogExtra record in this ServiceInfo, unless it is a duplicate.
func (si *ServiceInfo) AddPlan(name string, id string) {
	pi := PlanInfo{Name: name, CatalogID: id}
	if _, found := si.CatalogExtra.Plans[pi]; found {
		debug.Warning(`Ignoring duplicate plan in %s: %+v`, si.String(), pi)
	} else {
		si.CatalogExtra.Plans[pi] = struct{}{}
	}
}

// ListPlans returns the list of plans from the CatalogExtra record in this ServiceInfo, as a comma separated string
func (si *ServiceInfo) ListPlans() string {
	cx := si.CatalogExtra
	if cx == nil {
		return ""
	}
	result := collections.NewStringSet()
	for pi := range cx.Plans {
		result.Add(pi.Name)
	}
	return result.String()
}

// GetCatalogExtra returns the CatalogExtra structure from a ServiceInfo record, allocating one if necessary (and if the "allocate" flag is true)
func (si *ServiceInfo) GetCatalogExtra(allocate bool) *CatalogExtra {
	if !si.HasSourceMainCatalog() && si.IgnoredMainCatalog == nil {
		if si.CatalogExtra != nil {
			panic(fmt.Sprintf("ServiceInfo(%s) has CatalogExtra!=nil but HasSourceMainCatalog()==false and IgnoredMainCatalog==nil --  %+v", si.String(), si.CatalogExtra))
		}
		return nil
	}
	if si.CatalogExtra == nil && allocate {
		si.CatalogExtra = new(CatalogExtra)
		si.CatalogExtra.PartNumbers = collections.NewStringSet()
		si.CatalogExtra.Plans = make(map[PlanInfo]struct{})
		si.CatalogExtra.Locations = collections.NewStringSet()
		si.CatalogExtra.ChildrenKinds = collections.NewStringSet()
	}
	return si.CatalogExtra
}

// MergeCatalogExtra merges the content of a second CatalogExtra record into the one for this ServiceInfo
func (si *ServiceInfo) MergeCatalogExtra(cx2 *CatalogExtra) {
	if cx2 == nil {
		debug.Debug(debug.Pricing, "MergeCatalogExtra(%s) <- null input CatalogExtra", si.String())
		return
	}
	cx1 := si.GetCatalogExtra(true)

	if cx2.PartNumbers.Len() > 0 {
		if cx2.PartNumbers.Contains(ossrecord.ProductInfoNone) {
			if cx1.PartNumbers.Len() == 0 {
				cx1.PartNumbers.Add(ossrecord.ProductInfoNone)
			}
		} else {
			if cx1.PartNumbers.Contains(ossrecord.ProductInfoNone) {
				cx1.PartNumbers = collections.NewStringSet()
			}
			cx1.PartNumbers.Add(cx2.PartNumbers.Slice()...)
		}
	}

	for pi := range cx2.Plans {
		si.AddPlan(pi.Name, pi.CatalogID)
	}
	cx1.Locations.Add(cx2.Locations.Slice()...)
	cx1.ChildrenKinds.Add(cx2.ChildrenKinds.Slice()...)

	debug.Debug(debug.Pricing, "MergeCatalogExtra(%s) -> PartNumbers=%v", si.String(), cx1.PartNumbers)
}

// RecordPricingInfo records in the OSS entry any Catalog pricing info (esp. part numbers) found in the specified service and its plans
func RecordPricingInfo(si *ServiceInfo, r *catalogapi.Resource) {
	cx := si.GetCatalogExtra(true)
	if debug.IsDebugEnabled(debug.Pricing) {
		defer func() {
			debug.Debug(debug.Pricing, "RecordPricingInfo(%s) -> PartNumbers=%v", si.String(), cx.PartNumbers)
		}()
	}
	allParts := make(map[string]bool)
	/*
		err := catalog.ListPricingInfoFromCatalog(r, func(p *catalogapi.Pricing) {
			for _, m := range p.Metrics {
				part := strings.TrimSpace(m.MetricID)
				if part != "" {
					allParts[part] = true
				}
			}
		})
	*/
	if hasCachedPricingInfo() {
		if pi := getCachedPricingInfo(r.ID); pi != nil {
			for _, part := range pi.PartNumbers {
				allParts[part] = true
			}
			for _, v := range pi.Issues {
				si.OSSValidation.AddIssuePreallocated(v)
			}
		}
	} else {
		err := catalog.ListPricingInfo(r, func(p *catalog.BSSPricing) {
			for _, m := range p.Metrics {
				part := strings.TrimSpace(m.PartNumber)
				if part != "" {
					allParts[part] = true
				}
			}
		})
		if err != nil {
			cx.PartNumbers = nil
			debug.PrintError("Error recording pricing info for %s: %v", si.String(), err)
			si.AddValidationIssue(ossvalidation.WARNING, "Error recording pricing info", "%s", err.Error()).TagProductInfo()
			return
		}
	}

	if len(allParts) > 0 {
		if cx.PartNumbers.Len() == 1 && ossrecord.IsProductInfoNone(cx.PartNumbers.Slice()) {
			cx.PartNumbers = collections.NewStringSet()
		}
		for part := range allParts {
			cx.PartNumbers.Add(part)
		}
	} else {
		if cx.PartNumbers.Len() == 0 {
			cx.PartNumbers.Add(ossrecord.ProductInfoNone)
		}
	}
}

var azPattern = regexp.MustCompile("-[0-9]$")

// SortLocationsList sorts a list of Locations with "global" and multi-zone regions first, then datacenters and pops, then satellite containers,  then availability zones
func SortLocationsList(list []string) []string {
	maxSize := len(list)
	globalBucket := make([]string, 0, 1)
	regionsBucket := make([]string, 0, maxSize)
	azBucket := make([]string, 0, maxSize)
	satconBucket := make([]string, 0, maxSize)
	restBucket := make([]string, 0, maxSize)
	for _, s := range list {
		switch {
		case s == "global":
			globalBucket = append(globalBucket, s) // I know this bucket can only have one item, but I like symmetry..
		case azPattern.MatchString(s):
			azBucket = append(azBucket, s)
		case strings.HasPrefix(s, "satcon"):
			satconBucket = append(satconBucket, s)
		case strings.Contains(s, "-"):
			regionsBucket = append(regionsBucket, s)
		default:
			restBucket = append(restBucket, s)
		}
	}
	result := make([]string, 0, maxSize)
	sort.Strings(globalBucket)
	result = append(result, globalBucket...)
	sort.Strings(regionsBucket)
	result = append(result, regionsBucket...)
	sort.Strings(restBucket)
	result = append(result, restBucket...)
	sort.Strings(satconBucket)
	result = append(result, satconBucket...)
	sort.Strings(azBucket)
	result = append(result, azBucket...)
	return result
}

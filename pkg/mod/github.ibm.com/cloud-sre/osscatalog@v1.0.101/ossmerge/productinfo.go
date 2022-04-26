package ossmerge

import (
	"fmt"
	"sort"

	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/ossuid"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/partsinput"
)

var servicesByPartNumber = make(map[string][]*ServiceInfo)
var servicesByProductID = make(map[string][]*ServiceInfo)

// highestOSSUID is the highest OSS UID encountered across all OSS records seen in this run
var highestOSSUID = ossuid.BaseValue

// cloudPlatformProductIDs is the ProductID for the `ibm-cloud-platform` entry (inherited by all PLATFORMCOMPONENTS)
var cloudPlatformProductIDs []string

const cloudPlatformEntryName = `ibm-cloud-platform`

// mergeProductInfoPhaseOne updates the ProductIDs based on the available Part Numbers, and other product information
// First pass, one entry at a time (no relationships)
func (si *ServiceInfo) mergeProductInfoPhaseOne() {
	if debug.IsDebugEnabled(debug.Pricing) {
		debug.Debug(debug.Pricing, "mergeProductInfo(%s) <- HasSourceMainCatalog=%v CatalogExtra=%+v   HasPriorOSS=%v  PriorOSS.ProductInfo.PartNumbers=%v",
			si.String(), si.HasSourceMainCatalog(), si.CatalogExtra, si.HasPriorOSS(), si.PriorOSS.ProductInfo.PartNumbers)
		defer func() {
			debug.Debug(debug.Pricing, "mergeProductInfo(%s) -> ProductInfo.PartNumbers=%v",
				si.String(), si.OSSService.ProductInfo.PartNumbers)
		}()
	}
	si.OSSService.ProductInfo.PartNumbers = nil
	si.OSSService.ProductInfo.PartNumbersRefreshed = ""
	si.OSSService.ProductInfo.ProductIDs = nil
	si.OSSService.ProductInfo.ProductIDSource = ""
	var newPartNumbers []string
	if options.GlobalOptions().RefreshPricing && si.HasSourceMainCatalog() {
		cx := si.GetCatalogExtra(true)
		if cx.PartNumbers.Len() > 0 {
			newPartNumbers = cx.PartNumbers.Slice()
		} else {
			panic(fmt.Sprintf(`Unexpected empty PartNumbers=%v in refresh-pricing mode for entry "%s"`, cx.PartNumbers, si.String()))
		}
	} else if options.GlobalOptions().IncludePricing && si.HasSourceMainCatalog() {
		cx := si.GetCatalogExtra(true)
		if cx.PartNumbers.Len() > 0 {
			newPartNumbers = cx.PartNumbers.Slice()
		} else if si.HasPriorOSS() && si.GetPriorOSS().ProductInfo.PartNumbersRefreshed == "" {
			// Special case when the PriorOSS has been reset due to duplicate merging
			si.AddValidationIssue(ossvalidation.MINOR, `CatalogInfo.PartNumbers cannot be set because of recent merge with another duplicate entry`, "").TagProductInfo()
		} else if !si.HasPriorOSS() || len(si.GetPriorOSS().ProductInfo.PartNumbers) == 0 {
			panic(fmt.Sprintf(`Unexpected empty PartNumbers=%v in include-pricing mode for entry "%s" -- PriorOSS=%v   PriorOSS.PartNumbers=%v`, cx.PartNumbers, si.String(), si.HasPriorOSS(), si.PriorOSS.ProductInfo.PartNumbers))
		} else {
			// Not unexpected to have empty PartNumbers -- we will copy from PriorOSS
		}
	}
	if len(newPartNumbers) > 0 {
		if si.HasPriorOSS() && len(si.GetPriorOSS().ProductInfo.PartNumbers) > 0 && !ossrecord.IsProductInfoNone(si.GetPriorOSS().ProductInfo.PartNumbers) {
			if ossrecord.IsProductInfoNone(newPartNumbers) {
				si.AddValidationIssue(ossvalidation.MINOR, fmt.Sprintf(`CatalogInfo.PartNumbers reset to "%s" list after refresh from Catalog`, ossrecord.ProductInfoNone), "old value=%v", si.GetPriorOSS().ProductInfo.PartNumbers).TagProductInfo()
			}
		}
		si.OSSService.ProductInfo.PartNumbers = newPartNumbers
		si.OSSService.ProductInfo.PartNumbersRefreshed = options.GlobalOptions().LogTimeStamp
	} else {
		if si.HasPriorOSS() {
			if len(si.GetPriorOSS().ProductInfo.PartNumbers) > 0 {
				si.AddValidationIssue(ossvalidation.INFO, "CatalogInfo.PartNumbers copied over from old value -- no refresh from Catalog in this run", "").TagProductInfo()
				si.OSSService.ProductInfo.PartNumbers = si.GetPriorOSS().ProductInfo.PartNumbers
				si.OSSService.ProductInfo.PartNumbersRefreshed = si.GetPriorOSS().ProductInfo.PartNumbersRefreshed
			} else {
				si.AddValidationIssue(ossvalidation.INFO, "No CatalogInfo.PartNumbers - no previous value and no refresh from Catalog in this run", "").TagProductInfo()
				si.OSSService.ProductInfo.PartNumbers = nil
				si.OSSService.ProductInfo.PartNumbersRefreshed = si.GetPriorOSS().ProductInfo.PartNumbersRefreshed
			}
		}
	}
	for _, pn := range si.OSSService.ProductInfo.PartNumbers {
		if pn == ossrecord.ProductInfoNone {
			continue
		}
		servicesByPartNumber[pn] = append(servicesByPartNumber[pn], si)
	}

	if si.HasPriorOSS() && si.GetPriorOSS().ProductInfo.ProductIDSource == ossrecord.ProductIDSourceManual {
		si.OSSService.ProductInfo.ProductIDs = si.GetPriorOSS().ProductInfo.ProductIDs
		si.OSSService.ProductInfo.ProductIDSource = si.GetPriorOSS().ProductInfo.ProductIDSource
	}

	if len(si.OSSService.ProductInfo.PartNumbers) > 0 && !ossrecord.IsProductInfoNone(si.OSSService.ProductInfo.PartNumbers) {
		if partsinput.HasPartNumbers() {
			var newPids = collections.NewStringSet()
			for _, part := range si.OSSService.ProductInfo.PartNumbers {
				if part == ossrecord.ProductInfoNone {
					continue
				}
				if entry, found := partsinput.LookupPartNumber(part); found {
					if entry.ProductID != "" {
						newPids.Add(entry.ProductID)
					} else {
						si.AddValidationIssue(ossvalidation.WARNING, "Found product info for part number but pid is blank", "part number=%s   entry=%+v", part, entry).TagProductInfo()
					}
				} else {
					si.AddValidationIssue(ossvalidation.WARNING, "Cannot find product info for part number", "part number=%s", part).TagProductInfo()
				}
			}
			if si.OSSService.ProductInfo.ProductIDs == nil {
				if newPids.Len() == 0 {
					if si.HasPriorOSS() && len(si.GetPriorOSS().ProductInfo.ProductIDs) > 0 && !ossrecord.IsProductInfoNone(si.GetPriorOSS().ProductInfo.ProductIDs) {
						si.AddValidationIssue(ossvalidation.MINOR, fmt.Sprintf(`ProductInfo.ProductIDs reset to "%s" list after refresh from Catalog and Parts Info file`, ossrecord.ProductInfoNone), "old value=%v/%s", si.GetPriorOSS().ProductInfo.ProductIDs, si.GetPriorOSS().ProductInfo.ProductIDSource).TagProductInfo()
					}
					si.OSSService.ProductInfo.ProductIDs = []string{ossrecord.ProductInfoNone}
				} else {
					si.OSSService.ProductInfo.ProductIDs = newPids.Slice()
					si.OSSService.ProductInfo.ProductIDSource = ossrecord.ProductIDSourcePartNumbers
				}
			} else {
				si.AddValidationIssue(ossvalidation.INFO, `Ignoring ProductID derived from Part Numbers because there is an overriding value`, "from part numbers=%v,  overriding value=%v/%s", newPids, si.OSSService.ProductInfo.ProductIDs, si.OSSService.ProductInfo.ProductIDSource).TagProductInfo()
			}
		} else if si.OSSService.ProductInfo.ProductIDs == nil && si.HasPriorOSS() {
			prior := &si.GetPriorOSS().ProductInfo
			if prior.ProductIDSource == ossrecord.ProductIDSourcePartNumbers {
				si.AddValidationIssue(ossvalidation.INFO, "ProductInfo.ProductIDs copied over from old value -- no refresh from Catalog and Parts Info file in this run", "source=%s", prior.ProductIDSource).TagProductInfo()
				si.OSSService.ProductInfo.ProductIDs = prior.ProductIDs
				si.OSSService.ProductInfo.ProductIDSource = prior.ProductIDSource
			}
		}
		for _, pid := range si.OSSService.ProductInfo.ProductIDs {
			if pid == ossrecord.ProductInfoNone {
				continue
			}
			servicesByProductID[pid] = append(servicesByProductID[pid], si)
		}
	}

	// Take note of the ProductID for IBM Cloud Platform (for use by default with all PLATFORMCOMPONENT entries)
	if si.OSSService.ReferenceResourceName == cloudPlatformEntryName {
		if cloudPlatformProductIDs != nil {
			debug.Critical(`Found more than one entry for the IBM Cloud Platform (%s)`, cloudPlatformEntryName)
		} else if len(si.OSSService.ProductInfo.ProductIDs) == 1 && !ossrecord.IsProductInfoNone(si.OSSService.ProductInfo.ProductIDs) {
			cloudPlatformProductIDs = make([]string, len(si.OSSService.ProductInfo.ProductIDs))
			copy(cloudPlatformProductIDs, si.OSSService.ProductInfo.ProductIDs)
			debug.Info(`Found the ProductID for the IBM Cloud Platform (%s): %v`, cloudPlatformEntryName, cloudPlatformProductIDs)
		} else {
			debug.PrintError(`Did not find a valid (and single) ProductID for the IBM Cloud Platform (%s): %v`, cloudPlatformEntryName, si.OSSService.ProductInfo.ProductIDs)
		}
	}

	// Figure out the highest OSS UID of all the records examined
	if si.HasPriorOSS() && si.GetPriorOSS().ProductInfo.OSSUID != "" {
		uid := si.GetPriorOSS().ProductInfo.OSSUID
		si.ProductInfo.OSSUID = uid
		if si.GetPriorOSS().ProductInfo.OSSUID > highestOSSUID { // string compare is as good as numerical compare, given the structure of OSS UIDs
			highestOSSUID = si.GetPriorOSS().ProductInfo.OSSUID
		}
	}
}

// mergeProductInfoPhaseTwo updates additional ProductInfo attributes at the end of a run
func (si *ServiceInfo) mergeProductInfoPhaseTwo() {
	si.checkMergePhase(mergePhaseServicesTwo)
	// Nothing to do
}

// mergeProductInfoPhaseThree updates additional ProductInfo attributes at the end of a run,
// after all records have been scanned once AND all relationships between records have been set-up (i.e. in Phase 3)
// and after we have loaded all ClearingHouse links, if any.
// - the Division attribute: based on PIDs obtained both through part numbers and through ClearingHouse entries
func (si *ServiceInfo) mergeProductInfoPhaseThree() {
	// Guard against reentrant calls while recursing through the ParentResourceName name links
	if si.mergeWorkArea.mergeProductInfoPhaseThreeDone {
		return
	}
	si.mergeWorkArea.mergeProductInfoPhaseThreeDone = true

	si.checkGlobalMergePhaseMultiple(mergePhaseServicesThree)
	si.checkEntryMergePhaseMultiple(mergePhaseServicesTwo, mergePhaseServicesThree) // Note: some parent entries may be checked from a call from mergeProductInfoPhaseThree in a child entry and may not themselves be in phase THREE yet

	pinfo := &si.OSSService.ProductInfo

	// Check for ProductIDS inherited from parents, platform
	if si.OSSService.GeneralInfo.ParentResourceName != "" {
		if parent, found := LookupService(string(si.OSSService.GeneralInfo.ParentResourceName), false); found {
			parent.mergeProductInfoPhaseThree() // Recurse to ensure we find the ProductInfo for parents before the children
			if len(parent.OSSService.ProductInfo.ProductIDs) > 0 && !ossrecord.IsProductInfoNone(parent.OSSService.ProductInfo.ProductIDs) {
				if len(pinfo.ProductIDs) == 0 || ossrecord.IsProductInfoNone(pinfo.ProductIDs) {
					pinfo.ProductIDs = parent.OSSService.ProductInfo.ProductIDs
					pinfo.ProductIDSource = ossrecord.ProductIDSourceParent
				} else {
					si.AddValidationIssue(ossvalidation.INFO, `Ignoring ProductIDs from parent as this entry has its own ProductIDs`, "parent=%s %v", si.OSSService.GeneralInfo.ParentResourceName, parent.OSSService.ProductInfo.ProductIDs).TagProductInfo()
				}
			}
		} else {
			debug.Critical(`Cannot find parent record "%s" referenced in OSSService "%s"`, si.OSSService.GeneralInfo.ParentResourceName, si.String())
		}
	} else if si.OSSService.GeneralInfo.EntryType == ossrecord.PLATFORMCOMPONENT && cloudPlatformProductIDs != nil {
		if len(pinfo.ProductIDs) == 0 || ossrecord.IsProductInfoNone(pinfo.ProductIDs) {
			pinfo.ProductIDs = cloudPlatformProductIDs
			pinfo.ProductIDSource = ossrecord.ProductIDSourceCloudPlatform
		} else {
			si.AddValidationIssue(ossvalidation.INFO, `Ignoring ProductIDs from IBM Cloud Platform as this PLATFORMCOMPONENT entry has its own ProductIDs`, "platform=%s %v", cloudPlatformEntryName, cloudPlatformProductIDs).TagProductInfo()
		}
	}

	// Sanity check for ProductIDS
	if len(pinfo.ProductIDs) == 0 || ossrecord.IsProductInfoNone(pinfo.ProductIDs) {
		if pinfo.ProductIDSource != "" {
			panic(fmt.Sprintf(`mergeProductInfoPhaseThree(%s): ProductIDSource is not empty (%s) but ProductIDs is empty (%v)`, si.String(), pinfo.ProductIDSource, pinfo.ProductIDs))
		}
	} else {
		if pinfo.ProductIDSource == "" {
			panic(fmt.Sprintf(`mergeProductInfoPhaseThree(%s): ProductIDSource is empty but ProductIDs not empty (%v)`, si.String(), pinfo.ProductIDs))
		}
	}

	// Last ditch: preserve historical ProductIDs
	if pinfo.ProductIDSource == "" && si.HasPriorOSS() {
		prior := si.GetPriorOSS().ProductInfo
		if len(prior.ProductIDs) > 0 && !ossrecord.IsProductInfoNone(prior.ProductIDs) {
			pinfo.ProductIDs = prior.ProductIDs
			pinfo.ProductIDSource = ossrecord.ProductIDSourceHistorical
			si.AddValidationIssue(ossvalidation.WARNING, `Copying historical ProductIDs from prior OSS record`, "pid=%v  historical_source=%s", prior.ProductIDs, prior.ProductIDSource).TagProductInfo()
		}
	}

	// Figure out the division code
	if partsinput.HasPartNumbers() {
		if len(pinfo.ProductIDs) != 0 && !ossrecord.IsProductInfoNone(pinfo.ProductIDs) {
			divs := collections.NewStringSet()
			for _, pid := range pinfo.ProductIDs {
				if entries, ok := partsinput.LookupProductID(pid); ok {
					for _, e := range entries {
						divs.Add(e.Division)
					}
				} else {
					si.AddValidationIssue(ossvalidation.WARNING, "Cannot find Division code for Product ID (not in Parts Info file)", "pid=%s", pid).TagProductInfo()
				}
			}
			switch divs.Len() {
			case 0:
				pinfo.Division = ""
			case 1:
				pinfo.Division = divs.Slice()[0]
			default:
				pinfo.Division = "(multiple)"
				si.AddValidationIssue(ossvalidation.SEVERE, "Found multiple Division codes based on Product IDs ", "pids=%v  divs=%v", pinfo.ProductIDs, divs.Slice()).TagProductInfo()
			}
		} else {
			si.AddValidationIssue(ossvalidation.MINOR, "Cannot find Division code - no Product IDs", "").TagProductInfo()
		}
	} else if si.HasPriorOSS() {
		if len(si.GetPriorOSS().ProductInfo.Division) > 0 {
			si.AddValidationIssue(ossvalidation.INFO, "ProductInfo.Division copied over from old value -- no refresh from Catalog and Parts Info file in this run", "").TagProductInfo()
			pinfo.Division = si.GetPriorOSS().ProductInfo.Division
		} else {
			si.AddValidationIssue(ossvalidation.INFO, "No ProductInfo.Division code -- no previous value and no refresh from Catalog and Parts Info file in this run", "").TagProductInfo()
		}
	}

	// Allocate a new OSS UID if we do not have one yet
	if si.ProductInfo.OSSUID == "" {
		highestOSSUID = ossuid.Parse(highestOSSUID).Increment().String()
		si.ProductInfo.OSSUID = highestOSSUID
	}

}

// checkProductInfo checks for issues related to multiple entries sharing the same ProductInfo
// Note that this function must be run only ONCE - it automatically covers all known entries
func checkProductInfo() {
	for pn, silist := range servicesByPartNumber {
		if len(silist) > 1 {
			sinames := make([]string, 0, len(silist))
			for _, si := range silist {
				sinames = append(sinames, si.String())
			}
			sort.Strings(sinames)
			for _, si := range silist {
				si.AddValidationIssue(ossvalidation.WARNING, "Found multiple entries referencing the same Part Number", "PartNumber=%s   entries=%v", pn, sinames).TagProductInfo()
			}
		}
	}

	for pid, silist := range servicesByProductID {
		if len(silist) > 1 {
			sinames := make([]string, 0, len(silist))
			for _, si := range silist {
				sinames = append(sinames, si.String())
			}
			sort.Strings(sinames)
			for _, si := range silist {
				si.AddValidationIssue(ossvalidation.WARNING, "Found multiple entries referencing the same Product ID", "ProductID=%s   entries=%v", pid, sinames).TagProductInfo()
			}
		}
	}

	for cid, silist := range servicesByClearingHouseID {
		if len(silist) > 1 {
			sinames := make([]string, 0, len(silist))
			for _, si := range silist {
				sinames = append(sinames, si.String())
			}
			sort.Strings(sinames)
			for _, si := range silist {
				si.AddValidationIssue(ossvalidation.WARNING, "Found multiple entries referencing the same ClearingHouse ID", "ClearingHouseID=%s   entries=%v", cid, sinames).TagProductInfo()
			}
		}
	}
}

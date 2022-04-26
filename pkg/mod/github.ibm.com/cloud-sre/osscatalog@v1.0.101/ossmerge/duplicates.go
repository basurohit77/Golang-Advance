package ossmerge

import (
	"fmt"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// processDuplicateNames goes through all entries in the Model and combines the ones that
// have been marked as potential duplicates, according to the RawDuplicateNames and DoNotMergeNames
// attributes in the OSSControl records
func processDuplicateNames() error {
	allPattern := regexp.MustCompile(".*")
	errorCount := 0

	// Build a list of all duplicate names
	allDupNames := make(map[string]*ServiceInfo)
	err := ListAllServices(allPattern, func(si *ServiceInfo) {
		if si.OSSMergeControl != nil {
			if dups := si.OSSMergeControl.RawDuplicateNames; len(dups) > 0 {
				for _, n := range dups {
					if _, found := allDoNotMergeNames[n]; found {
						// TODO: the error should report which two OSSMergeControl entries contain the conflict
						// TODO: what if we do not want to merge a raw name, but still want to merge/dup the folded/comparable name flavor of the same name
						// TODO: should we copy each do-not-merge name into all concerned OSSMergeControl entries
						// (i.e the one that will represent the do-not-merge name itself and the one in which it would otherwise have been folded
						debug.PrintError("Name \"%s\" is claimed as both a duplicate and a do-not-merge name in two separate OSSMergeControl entries", n)
						errorCount++
					} else {
						cn := MakeComparableName(n)
						prior, found := allDupNames[cn]
						if found {
							debug.PrintError("Name \"%s\" is claimed as a duplicate by two separate OSSMergeControl entries: \"%s\" and \"%s\"", n, si.OSSMergeControl.CanonicalName, prior.OSSMergeControl.CanonicalName)
							errorCount++
						} else {
							allDupNames[cn] = si
						}
					}
				}
			}
		}
	})
	if err != nil {
		return debug.WrapError(err, "processDuplicateNames(): error while traversing all entries - aborting")
	}
	if errorCount > 0 {
		return fmt.Errorf("processDuplicateNames(): found %d conflicts in list of all duplicate names - aborting", errorCount)
	}

	// Go through all duplicate entries and merge them
	for cn, si := range allDupNames {
		if dup, found := LookupService(cn, false); found {
			debug.Info(`Merging into entry "%s" <- %s`, si.GetServiceName(), dup.String())
			if dup.DuplicateOf != "" {
				panic(fmt.Sprintf(`Marking entry "%s" as duplicate of "%s" but it is already marked as a duplicate of entry "%s"`, dup.GetServiceName(), si.GetServiceName(), dup.DuplicateOf))
			}
			if dup.OSSMergeControl != nil && !dup.OSSMergeControl.IsEmptyExceptNotes() {
				debug.PrintError(`Attempting to merge a duplicate entry for "%s" but OSSMergeControl is not empty: %s`, dup.ComparableName, dup.OSSMergeControl.OneLineString())
				errorCount++
				continue
			}

			if dup.HasPriorOSS() && si.HasPriorOSS() {
				// Special handling for information that would normally be copied over from a prior OSS merge, that is not recomputed in this run:
				// invalidate if we are merging a duplicate record for the first time (i.e, it itself has a PriorOSS record)
				//
				// XXX TODO: revise special handling of optional merge actions when merging duplicates
				si.AddValidationIssue(ossvalidation.INFO, "Resetting cached PartNumbers because of merging a new duplicate entry for the first time: %s", dup.String())
				si.GetPriorOSS().ProductInfo.PartNumbersRefreshed = ""
				si.GetPriorOSS().ProductInfo.PartNumbers = nil
				si.GetPriorOSS().ProductInfo.ProductIDs = nil
				si.GetPriorOSS().ProductInfo.ClearingHouseReferences = nil
			}

			if dup.HasSourceMainCatalog() {
				if si.HasSourceMainCatalog() {
					si.AdditionalMainCatalog = append(si.AdditionalMainCatalog, &dup.SourceMainCatalog)
				} else {
					si.SourceMainCatalog = dup.SourceMainCatalog
				}
				dupcx := dup.GetCatalogExtra(false)
				if dupcx != nil {
					si.MergeCatalogExtra(dupcx)
				}
			}
			if len(dup.AdditionalMainCatalog) > 0 {
				si.AdditionalMainCatalog = append(si.AdditionalMainCatalog, dup.AdditionalMainCatalog...)
			}

			if dup.HasSourceScorecardV1Detail() {
				if si.HasSourceScorecardV1Detail() {
					si.AdditionalScorecardV1Detail = append(si.AdditionalScorecardV1Detail, &dup.SourceScorecardV1Detail)
				} else {
					si.SourceScorecardV1Detail = dup.SourceScorecardV1Detail
				}
			}
			if len(dup.AdditionalScorecardV1Detail) > 0 {
				si.AdditionalScorecardV1Detail = append(si.AdditionalScorecardV1Detail, dup.AdditionalScorecardV1Detail...)
			}

			if dup.HasSourceServiceNow() {
				if si.HasSourceServiceNow() {
					si.AdditionalServiceNow = append(si.AdditionalServiceNow, &dup.SourceServiceNow)
				} else {
					si.SourceServiceNow = dup.SourceServiceNow
				}
			}
			if len(dup.AdditionalServiceNow) > 0 {
				si.AdditionalServiceNow = append(si.AdditionalServiceNow, dup.AdditionalServiceNow...)
			}

			if dup.IgnoredMainCatalog != nil {
				if si.IgnoredMainCatalog == nil {
					si.IgnoredMainCatalog = dup.IgnoredMainCatalog
				} else {
					// TODO: Handle merging of multiple duplicate entries with IgnoredMainCatalog info
					debug.Debug(debug.Merge, `** Cannot merge more than one entry with IgnoredMainCatalog info: dest="%s" <- %s`, cn, dup.String())
				}
			}

			if dup.MonitoringInfo != nil {
				// TODO: merge duplicate MonitoringInfo
			}

			dup.DuplicateOf = si.ComparableName
			dup.OSSService.ReferenceResourceName = ossrecord.CRNServiceName(dup.GetServiceName())
		}

	}
	if errorCount > 0 {
		return fmt.Errorf("processDuplicateNames(): skipped merging %d duplicate entries with errors", errorCount)
	}

	return nil
}

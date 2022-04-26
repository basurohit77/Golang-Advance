package ossmerge

import (
	"fmt"
	"sort"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"gopkg.in/yaml.v2"
)

type cachedCHDependenciesEntry struct {
	CHLabel                  string
	OutboundDependencies     ossrecord.Dependencies
	InboundDependencies      ossrecord.Dependencies
	Issues                   []*ossvalidation.ValidationIssue
	ListInboundNotOSS        []string
	ListInboundNotCloud      []string
	ListOutboundNotOSS       []string
	ListOutboundNotCloud     []string
	CountIgnoredDependencies int
}

var cachedCHDependencies = make(map[clearinghouse.DeliverableID]*cachedCHDependenciesEntry)

// mergeCHDependencies merges dependency information from ClearingHouse
func (si *ServiceInfo) mergeCHDependencies() {
	si.checkMergePhase(mergePhaseServicesThree)

	if ossrunactions.DependenciesClearingHouse.IsEnabled() {
		//		si.AddValidationIssue(ossvalidation.INFO, "Re-computing Dependency-ClearingHouse information in this run", "").TagDependencies()
		si.OSSValidation.RecordRunAction(ossrunactions.DependenciesClearingHouse)
		var listInboundNotOSS []string
		var listInboundNotCloud []string
		var listOutboundNotOSS []string
		var listOutboundNotCloud []string
		for _, chref := range si.OSSService.ProductInfo.ClearingHouseReferences {
			cid := chref.ID
			if cid == ossrecord.ProductInfoNone {
				continue
			}
			cached, found := cachedCHDependencies[clearinghouse.DeliverableID(cid)]
			if !found {
				newCached := &cachedCHDependenciesEntry{}
				rec, err := clearinghouse.GetFullRecordByID(clearinghouse.DeliverableID(cid))
				if err != nil {
					issue := ossvalidation.NewIssue(ossvalidation.CRITICAL, "Invalid ClearingHouse DeliverableID - cannot fetch dependency information", "CHID=%q  err=%v", cid, err).TagRunAction(ossrunactions.DependenciesClearingHouse)
					debug.Warning("mergeCHDependencies(service=%s): %s", si.String(), issue.String())
					newCached.CHLabel = clearinghouse.MakeCHLabel("", clearinghouse.DeliverableID(cid))
					newCached.Issues = append(newCached.Issues, issue)
				} else {
					newCached.CHLabel = clearinghouse.MakeCHLabel(rec.Name, clearinghouse.DeliverableID(rec.ID))
					_parseCHDependencies(true, rec, newCached)
					_parseCHDependencies(false, rec, newCached)
					sort.Strings(newCached.ListInboundNotCloud)
					sort.Strings(newCached.ListOutboundNotCloud)
					sort.Strings(newCached.ListInboundNotOSS)
					sort.Strings(newCached.ListOutboundNotOSS)
					// TODO: validate that the inbound and outbound dependencices from ClearingHouse match across records
				}
				cachedCHDependencies[clearinghouse.DeliverableID(cid)] = newCached
				cached = newCached
				if debug.IsDebugEnabled(debug.ClearingHouse) {
					data, err := yaml.Marshal(newCached)
					if err != nil {
						data = ([]byte)(fmt.Sprintf("*** %v", err))
					}
					debug.Debug(debug.ClearingHouse, "ossmerge.mergeCHDependencies(%s): Recording cached dependencies for CHID=%s:\n%s", si.String(), cid, string(data))
				}
			}
			for _, d := range cached.OutboundDependencies {
				si.OSSService.DependencyInfo.OutboundDependencies.AddDependency(d.Service, d.Tags...)
			}
			for _, d := range cached.InboundDependencies {
				si.OSSService.DependencyInfo.InboundDependencies.AddDependency(d.Service, d.Tags...)
			}
			for _, issue := range cached.Issues {
				si.OSSValidation.AddIssuePreallocated(issue)
			}
			listInboundNotOSS, _ = collections.AppendSliceStringNoDups(listInboundNotOSS, cached.ListInboundNotOSS...)
			listOutboundNotOSS, _ = collections.AppendSliceStringNoDups(listOutboundNotOSS, cached.ListOutboundNotOSS...)
			listInboundNotCloud, _ = collections.AppendSliceStringNoDups(listInboundNotCloud, cached.ListInboundNotCloud...)
			listOutboundNotCloud, _ = collections.AppendSliceStringNoDups(listOutboundNotCloud, cached.ListOutboundNotCloud...)
			if len(cached.ListInboundNotCloud) > 0 {
				issue := si.OSSValidation.AddNamedIssue(ossvalidation.CHDependencyInboundNotCloud, "count=%d   CHentry=%s", len(cached.ListInboundNotCloud), cached.CHLabel)
				data, err := yaml.Marshal(cached.ListInboundNotCloud)
				if err != nil {
					data = ([]byte)(fmt.Sprintf("*** %v", err))
				}
				debug.PlainLogEntry(debug.LevelINFO, si.String(), "Merging %s for Service %s: %s\n%s", ossrunactions.DependenciesClearingHouse.Name(), si.String(), issue.GetText(), string(data))
			}
			if len(cached.ListOutboundNotCloud) > 0 {
				issue := si.OSSValidation.AddNamedIssue(ossvalidation.CHDependencyOutboundNotCloud, "count=%d   CHentry=%s", len(cached.ListOutboundNotCloud), cached.CHLabel)
				data, err := yaml.Marshal(cached.ListOutboundNotCloud)
				if err != nil {
					data = ([]byte)(fmt.Sprintf("*** %v", err))
				}
				debug.PlainLogEntry(debug.LevelINFO, si.String(), "Merging %s for Service %s: %s\n%s", ossrunactions.DependenciesClearingHouse.Name(), si.String(), issue.GetText(), string(data))
			}
			if len(cached.ListInboundNotOSS) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHDependencyInboundNotOSS, "count=%d   CHentry=%s", len(cached.ListInboundNotOSS), cached.CHLabel)
			}
			if len(cached.ListOutboundNotOSS) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHDependencyOutboundNotOSS, "count=%d   CHentry=%s", len(cached.ListOutboundNotOSS), cached.CHLabel)
			}
			if cached.CountIgnoredDependencies > 0 {
				si.AddValidationIssue(ossvalidation.IGNORE, "Ignoring one or more ClearingHouse Dependency records in Cancel or Reject status", "count=%d   CHentry=%s", cached.CountIgnoredDependencies, cached.CHLabel).TagRunAction(ossrunactions.DependenciesClearingHouse)
			}
		}
		// Add a dummy dependency entry to reflect the count of non-Cloud items
		// Note that we prefix with "~~~" to ensure that it gets sorted to the end
		if len(listInboundNotCloud) > 0 {
			_addDependency(&si.OSSService.DependencyInfo.InboundDependencies, fmt.Sprintf("~~~ + %d potentially non-Cloud items", len(listInboundNotCloud)), "", "", nil, "NotCloud")
		}
		if len(listOutboundNotCloud) > 0 {
			_addDependency(&si.OSSService.DependencyInfo.OutboundDependencies, fmt.Sprintf("~~~ + %d potentially non-Cloud items", len(listOutboundNotCloud)), "", "", nil, "NotCloud")
		}
	} else {
		si.OSSValidation.CopyRunAction(si.GetPriorOSSValidation(), ossrunactions.DependenciesClearingHouse)
		if si.HasPriorOSS() && si.HasPriorOSSValidation() {
			//			si.AddValidationIssue(ossvalidation.INFO, "Not computing Dependency-ClearingHouse information in this run -- copying data from an earlier run", "").TagDependencies()
			// Need to copy dependencies explicitly one by one rather than replace the whole structure,
			// because we might also have some non-ClearingHouse dependencies in the mix, added in this merge
			for _, d := range si.GetPriorOSS().DependencyInfo.OutboundDependencies {
				si.OSSService.DependencyInfo.OutboundDependencies.AddDependency(d.Service, d.Tags...)
			}
			for _, d := range si.GetPriorOSS().DependencyInfo.InboundDependencies {
				si.OSSService.DependencyInfo.InboundDependencies.AddDependency(d.Service, d.Tags...)
			}
		} else {
			//			si.AddValidationIssue(ossvalidation.INFO, "Not computing Dependency-ClearingHouse information in this run -- no data from a previous run", "").TagDependencies()
		}
	}
}

func _parseCHDependencies(inbound bool, chEntry *clearinghouse.CHDeliverableWithDependencies, chDep *cachedCHDependenciesEntry) {
	var label string
	var src *struct {
		Dependencies []clearinghouse.CHDependency `json:"dependencies"`
	}
	var dest *ossrecord.Dependencies
	if inbound {
		label = "DependencyOriginators (inbound)"
		src = chEntry.DependencyOriginators
		dest = &chDep.InboundDependencies
	} else {
		label = "DependencyProviders (outbound)"
		src = chEntry.DependencyProviders
		dest = &chDep.OutboundDependencies
	}
	if src == nil {
		issue := ossvalidation.NewIssue(ossvalidation.INFO, fmt.Sprintf("%s for ClearingHouse entry is nil", label), "this CH entry=%s", chEntry.String()).TagRunAction(ossrunactions.DependenciesClearingHouse)
		chDep.Issues = append(chDep.Issues, issue)
		return
	} else if len(src.Dependencies) == 0 {
		issue := ossvalidation.NewIssue(ossvalidation.INFO, fmt.Sprintf("%s for ClearingHouse entry is empty", label), "this CH entry=%s", chEntry.String()).TagRunAction(ossrunactions.DependenciesClearingHouse)
		chDep.Issues = append(chDep.Issues, issue)
		return
	}
	for _, chd := range src.Dependencies {
		if chd.CommitStatus == "Cancel" || chd.CommitStatus == "Reject" {
			chDep.CountIgnoredDependencies++
			continue
		}
		var targetCHID, myCHID clearinghouse.DeliverableID
		var targetCHName, myCHName string
		if inbound {
			targetCHID = clearinghouse.DeliverableID(chd.OriginatorID)
			targetCHName = chd.OriginatorName
			myCHID = clearinghouse.DeliverableID(chd.ProviderID)
			myCHName = chd.ProviderName
		} else {
			targetCHID = clearinghouse.DeliverableID(chd.ProviderID)
			targetCHName = chd.ProviderName
			myCHID = clearinghouse.DeliverableID(chd.OriginatorID)
			myCHName = chd.OriginatorName
		}
		var localIssues []*ossvalidation.ValidationIssue
		if myCHID != clearinghouse.DeliverableID(chEntry.ID) {
			issue := ossvalidation.NewIssue(ossvalidation.CRITICAL, fmt.Sprintf("%s for ClearingHouse entry has unexpected local ID", label), `this CH entry=%s   local=%s   DependencyID=%s`, chEntry.String(), clearinghouse.MakeCHLabel(myCHName, myCHID), chd.DependencyID).TagRunAction(ossrunactions.DependenciesClearingHouse)
			debug.PrintError("mergeCHDependencies(): %s", issue.String())
			localIssues = append(localIssues, issue)
			// XXX Should we continue and record this dependency anyway, or skip it?
		}
		if targetServices, found := LookupServicesByCHID(clearinghouse.DeliverableID(targetCHID)); found {
			if len(targetServices) == 0 {
				panic(fmt.Sprintf("ossmerge.parseCHDependencies(): LookupServicesByCHID(%s) returned empty list of ServiceInfos", targetCHID))
			} else if len(targetServices) > 1 {
				targetNames := make([]string, 0, len(targetServices))
				for _, si := range targetServices {
					targetNames = append(targetNames, string(si.OSSService.ReferenceResourceName))
				}
				sort.Strings(targetNames)
				// Note: we cannot include the DependencyID in this ValidationIssue, because we frequently have multiple CH dependency records referencing the same OSS entry/entries
				issue := ossvalidation.NewIssue(ossvalidation.MINOR, fmt.Sprintf("%s for a ClearingHouse entry references multiple OSS records from a single Dependency record", label), `this CH entry=%s   targets=%v`, chEntry.String(), targetNames).TagRunAction(ossrunactions.DependenciesClearingHouse)
				localIssues = append(localIssues, issue)
			}
			// XXX Should also check for multiple CH Dependency records referencing the same OSS entry
			for _, si := range targetServices {
				_addDependency(dest, string(si.OSSService.ReferenceResourceName), chd.DependencyType, chd.CommitStatus, localIssues)
			}
		} else {
			targetOSSName := clearinghouse.MakeCHLabel(targetCHName, targetCHID)
			if _, found := clearinghouse.LookupSummaryEntryByID(targetCHID, targetCHName); found {
				// We did not find a OSS entry, but at least it is in ClearingHouse as a Cloud entry
				if inbound {
					chDep.ListInboundNotOSS, _ = collections.AppendSliceStringNoDups(chDep.ListInboundNotOSS, targetOSSName)
				} else {
					chDep.ListOutboundNotOSS, _ = collections.AppendSliceStringNoDups(chDep.ListOutboundNotOSS, targetOSSName)
				}
				_addDependency(dest, targetOSSName, chd.DependencyType, chd.CommitStatus, localIssues, ossrecord.DependencyTagNotOSS)
			} else {
				// Probably not even a Cloud entry (not in ClearingHouse as deployment_target=Cloud)
				if inbound {
					chDep.ListInboundNotCloud, _ = collections.AppendSliceStringNoDups(chDep.ListInboundNotCloud, targetOSSName)
				} else {
					chDep.ListOutboundNotCloud, _ = collections.AppendSliceStringNoDups(chDep.ListOutboundNotCloud, targetOSSName)
				}
				/*
					_addDependency(dest, ossrecord.CRNServiceName(targetOSSName), chd.DependencyType, localIssues, false)
				*/
			}
			/*
				issue := ossvalidation.NewIssue(ossvalidation.WARNING, fmt.Sprintf("%s for ClearingHouse entry points to another ClearingHouse entry with no known OSS record", label), `this CH entry=%s   target=%s   DependencyID=%s`, chEntry.String(), targetOSSName, chd.DependencyID).TagRunAction(ossrunactions.DependenciesClearingHouse)
				localIssues = append(localIssues, issue)
			*/
		}
		// Append to the global list of issues for this CH entry -- but avoid duplicate issues from multiple Dependency records
		for _, issue := range localIssues {
			var isDup bool
			for _, issue0 := range chDep.Issues {
				if issue0.IsEqual(issue) {
					isDup = true
					break
				}
			}
			if !isDup {
				chDep.Issues = append(chDep.Issues, issue)
			}
		}
	}
	return
}

func _addDependency(dest *ossrecord.Dependencies, targetOSSName string, depType string, depStatus string, issues []*ossvalidation.ValidationIssue, moreTags ...string) {
	tags := make([]string, 0, 3)
	tags = append(tags, ossrecord.DependencyTagSourceClearingHouse)
	if depType != "" {
		tags = append(tags, fmt.Sprintf("%s%s", ossrecord.DependencyTagType, depType))
	}
	if depStatus != "" {
		tags = append(tags, fmt.Sprintf("%s%s", ossrecord.DependencyTagStatus, depStatus))
	}
	if len(issues) != 0 {
		tags = append(tags, fmt.Sprintf("%s%d", ossrecord.DependencyTagIssues, len(issues)))
	}
	if len(moreTags) > 0 {
		tags = append(tags, moreTags...)
	}
	dest.AddDependency(targetOSSName, tags...)
}

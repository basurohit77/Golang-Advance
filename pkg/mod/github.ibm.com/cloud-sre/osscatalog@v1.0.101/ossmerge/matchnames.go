package ossmerge

import (
	"fmt"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// sharedNamesEntry is one entry that brings together all ServiceInfo and
// all ClearingHouse records that share a common (comparable) name
type sharedNamesEntry struct {
	ComparableName string
	ServiceInfos   []*struct {
		ServiceInfo *ServiceInfo
		SourceNames []string
	}
	CHEntries []*struct {
		CHEntry     *clearinghouse.CHSummaryEntry
		SourceNames []string
	}
}

// clearingHouseExcludedNames is a list of regex patterns for ClearingHouse entry names
// that should be EXCLUDED from the loading, because we know they are not actually IBM Cloud offerings
var clearingHouseExcludedNames = []*regexp.Regexp{
	regexp.MustCompile(`for Marketplace\s*$`),
	regexp.MustCompile(`on AWS\s*$`),
}

var allSharedNames = make(map[string]*sharedNamesEntry)

var nameIssuesByCHID = make(map[clearinghouse.DeliverableID][]*ossvalidation.ValidationIssue)

//var allNameGroups = make(map[])

func addOneServiceInfoName(name string, source string, si *ServiceInfo) {
	if name == "" {
		return
	}
	trimmed := strings.TrimSpace(name)
	sourceName := fmt.Sprintf("%s[%s]", trimmed, strings.TrimSpace(source))
	if isNameNotMergeable(trimmed) {
		nameIssue := ossvalidation.NewIssue(ossvalidation.MINOR, "Ignoring not mergeable name in OSS record", "%q", sourceName).TagProductInfo()
		si.OSSValidation.AddIssuePreallocated(nameIssue)
		debug.Debug(debug.Fine, "addOneServiceInfoName(%s) recording issue: %q", name, nameIssue.GetText())
		return
	}
	comparableName := MakeComparableName(name)
	if sne, found := allSharedNames[comparableName]; found {
		for _, si0 := range sne.ServiceInfos {
			if si == si0.ServiceInfo {
				for _, n0 := range si0.SourceNames {
					if sourceName == n0 {
						return
					}
				}
				si0.SourceNames = append(si0.SourceNames, sourceName)
				return
			}
		}
		si1 := &struct {
			ServiceInfo *ServiceInfo
			SourceNames []string
		}{
			ServiceInfo: si,
			SourceNames: []string{sourceName},
		}
		sne.ServiceInfos = append(sne.ServiceInfos, si1)
	} else {
		sne = &sharedNamesEntry{ComparableName: comparableName}
		si1 := &struct {
			ServiceInfo *ServiceInfo
			SourceNames []string
		}{
			ServiceInfo: si,
			SourceNames: []string{sourceName},
		}
		sne.ServiceInfos = append(sne.ServiceInfos, si1)
		allSharedNames[comparableName] = sne
	}
}

func addClearingHouseName(name string, source string, ch *clearinghouse.CHSummaryEntry) {
	if name == "" {
		return
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return
	}
	sourceName := fmt.Sprintf("%s[%s]", trimmed, strings.TrimSpace(source))
	if isNameNotMergeable(trimmed) {
		nameIssue := ossvalidation.NewIssue(ossvalidation.MINOR, "Ignoring not mergeable name in ClearingHouse record", "%s: %q", ch.String(), sourceName).TagProductInfo()
		nameIssuesByCHID[ch.DeliverableID] = append(nameIssuesByCHID[ch.DeliverableID], nameIssue)
		debug.Debug(debug.Fine, "addClearingHouseName(%s) recording issue: %q", name, nameIssue.GetText())
		return
	}
	comparableName := MakeComparableName(name)
	if sne, found := allSharedNames[comparableName]; found {
		for _, ch0 := range sne.CHEntries {
			if ch == ch0.CHEntry {
				for _, n0 := range ch0.SourceNames {
					if sourceName == n0 {
						return
					}
				}
				ch0.SourceNames = append(ch0.SourceNames, sourceName)
				return
			}
		}
		ch1 := &struct {
			CHEntry     *clearinghouse.CHSummaryEntry
			SourceNames []string
		}{
			CHEntry:     ch,
			SourceNames: []string{sourceName},
		}
		sne.CHEntries = append(sne.CHEntries, ch1)
	} else {
		sne = &sharedNamesEntry{ComparableName: comparableName}
		ch1 := &struct {
			CHEntry     *clearinghouse.CHSummaryEntry
			SourceNames []string
		}{
			CHEntry:     ch,
			SourceNames: []string{sourceName},
		}
		sne.CHEntries = append(sne.CHEntries, ch1)
		allSharedNames[comparableName] = sne
	}
}

// loadAllNames loads all known names from ServiceInfo and ClearingHouse records into one common table
func loadAllNames() error {

	// List all known names from ServiceInfo entries (already loaded from various sources)
	err := ListAllServices(nil, func(si *ServiceInfo) {
		if si.IsDeletable() {
			return
		}
		addOneServiceInfoName(si.ComparableName, "ossmerge.ComparableName", si)
		addOneServiceInfoName(string(si.OSSService.ReferenceResourceName), "ossmerge.ReferenceResourceName", si)
		addOneServiceInfoName(si.OSSService.ReferenceDisplayName, "ossmerge.ReferenceDisplayName", si)
		if si.HasSourceMainCatalog() {
			addOneServiceInfoName(si.GetSourceMainCatalog().Name, string(ossvalidation.CATALOG)+".Name", si)
			addOneServiceInfoName(si.GetSourceMainCatalog().OverviewUI.En.DisplayName, string(ossvalidation.CATALOG)+".DisplayName", si)
		}
		for _, e := range si.AdditionalMainCatalog {
			addOneServiceInfoName(e.Name, string(ossvalidation.CATALOG)+".Name", si)
			addOneServiceInfoName(e.OverviewUI.En.DisplayName, string(ossvalidation.CATALOG)+".DisplayName", si)
		}
		if si.HasSourceServiceNow() {
			addOneServiceInfoName(si.GetSourceServiceNow().CRNServiceName, string(ossvalidation.SERVICENOW)+".CRNServiceName", si)
			addOneServiceInfoName(si.GetSourceServiceNow().DisplayName, string(ossvalidation.SERVICENOW)+".DisplayName", si)
		}
		for _, e := range si.AdditionalServiceNow {
			addOneServiceInfoName(e.CRNServiceName, string(ossvalidation.SERVICENOW)+".CRNServiceName", si)
			addOneServiceInfoName(e.DisplayName, string(ossvalidation.SERVICENOW)+".DisplayName", si)
		}
		if si.HasSourceScorecardV1Detail() {
			addOneServiceInfoName(si.GetSourceScorecardV1Detail().Name, string(ossvalidation.SCORECARDV1)+".Name", si)
			addOneServiceInfoName(si.GetSourceScorecardV1Detail().DisplayName, string(ossvalidation.SCORECARDV1)+".DisplayName", si)
		}
		for _, e := range si.AdditionalScorecardV1Detail {
			addOneServiceInfoName(e.Name, string(ossvalidation.SCORECARDV1)+".Name", si)
			addOneServiceInfoName(e.DisplayName, string(ossvalidation.SCORECARDV1)+".DisplayName", si)
		}
		names := si.OSSValidation.SourceNames(ossvalidation.SCORECARDV1DISABLED)
		for _, n := range names {
			addOneServiceInfoName(n, string(ossvalidation.SCORECARDV1DISABLED)+".Name", si)
		}
	})
	if err != nil {
		return err
	}

	// List all known names from ClearingHouse entries
	err = clearinghouse.ListSummaryEntries(nil, func(ch *clearinghouse.CHSummaryEntry) {
		for _, pat := range clearingHouseExcludedNames {
			if pat.FindString(ch.Name) != "" {
				debug.Debug(debug.Fine, `loadAllNames() ignoring CH entry "%s"/%s that matches the clearingHouseExcludedNames patterns`, ch.Name, ch.DeliverableID)
				return
			}
		}
		addClearingHouseName(ch.Name, "ClearingHouse.Name", ch)
		addClearingHouseName(ch.CodeName, "ClearingHouse.CodeName", ch)
		addClearingHouseName(ch.OfficialName, "ClearingHouse.OfficialName", ch)
		addClearingHouseName(ch.ShortName, "ClearingHouse.ShortName", ch)
		addClearingHouseName(ch.CRNServiceName, "ClearingHouse.CRNServiceName", ch)
	})
	if err != nil {
		return err
	}

	return nil
}

// dump produces a multi-line string with detailed contents of one sharedNamesEntry
func (sne *sharedNamesEntry) dump(firstPrefix, mainPrefix string) string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("%s%q:\n", firstPrefix, sne.ComparableName))
	for _, si := range sne.ServiceInfos {
		buf.WriteString(fmt.Sprintf("%s    -OSSName: %q    sources: %q\n", mainPrefix, si.ServiceInfo.String(), si.SourceNames))
	}
	for _, ch := range sne.CHEntries {
		buf.WriteString(fmt.Sprintf("%s    -CHEntry: %s    sources: %q\n", mainPrefix, ch.CHEntry.String(), ch.SourceNames))
	}
	return buf.String()
}

func mergeNameGroups() error {
	type Label int
	const MaxLabel Label = 999999
	var labelsEquivalency = make(map[Label]Label)
	var lastLabel Label
	var labelsByOSSName = make(map[ossrecord.CRNServiceName]Label)
	var labelsByCHID = make(map[clearinghouse.DeliverableID]Label)
	var nameGroupsByLabel = make(map[Label]*NameGroup)
	var traceSharedNamesByLabel = make(map[Label][]*sharedNamesEntry) // for producing diagnostics information about the source of each NameGroup

	//	previousDebug := debug.SetDebugFlags(debug.Fine)
	//	defer debug.SetDebugFlags(previousDebug)

	// Utility function: find the root of the equivalency class for a group of labels
	var findRootLabel = func(label Label) Label {
		currentLabel := label
		firstRootLabel, found := labelsEquivalency[currentLabel]
		rootLabel := firstRootLabel
		for found && currentLabel != rootLabel {
			if !found {
				panic(fmt.Sprintf("ossmerge.mergeNameGroups() found unknown label %d while processing equivalency classess", currentLabel))
			}
			currentLabel = rootLabel
			rootLabel, found = labelsEquivalency[currentLabel]
		}
		// for performance improvement for next time: record the direct path to root
		if rootLabel != firstRootLabel {
			labelsEquivalency[label] = rootLabel
		}
		return rootLabel
	}

	// Utility function: add a new label to an equivalency class
	var addEquivalency = func(label1, label2 Label) {
		if label1 == label2 {
			return
		}
		label1Root := findRootLabel(label1)
		label2Root := findRootLabel(label2)
		if label1Root == label2Root {
			// already same root -- nothing to do
		} else if label1Root < label2Root {
			labelsEquivalency[label2Root] = label1Root
			labelsEquivalency[label2] = label1Root
		} else if label1Root > label2Root {
			labelsEquivalency[label1Root] = label2Root
			labelsEquivalency[label1] = label2Root
		}
	}

	// First pass: go through each raw sharedNamesEntry record and assign the lowest possible label to each ServiceInfo or CH entry.
	// Make note of equivalency classes between labels.
	debug.Debug(debug.Fine, `mergeNameGroups(): starting first pass`)
	for _, sne := range allSharedNames {
		debug.Debug(debug.Fine, `mergeNameGroups() processing sharedNamesEntry %s`, sne.dump("", ""))
		var mainLabel = MaxLabel
		for _, si := range sne.ServiceInfos {
			label := labelsByOSSName[si.ServiceInfo.OSSService.ReferenceResourceName]
			if label != 0 && label < mainLabel {
				mainLabel = label
			}
		}
		for _, ch := range sne.CHEntries {
			label := labelsByCHID[ch.CHEntry.DeliverableID]
			if label != 0 && label < mainLabel {
				mainLabel = label
			}
		}
		if mainLabel == MaxLabel {
			// Assign a new label
			lastLabel++
			mainLabel = lastLabel
			labelsEquivalency[mainLabel] = mainLabel // Start as the root of its class
		} else {
			mainLabel = findRootLabel(mainLabel)
		}
		traceSharedNamesByLabel[mainLabel] = append(traceSharedNamesByLabel[mainLabel], sne)
		for _, si := range sne.ServiceInfos {
			label := labelsByOSSName[si.ServiceInfo.OSSService.ReferenceResourceName]
			switch label {
			case 0:
				// no label yet -- give it the main label
				labelsByOSSName[si.ServiceInfo.OSSService.ReferenceResourceName] = mainLabel
			case mainLabel:
				// already has the right label -- do nothing
			default:
				// has a different label -- replace with the main label (which is the lowest number) and record equivalency between these labels
				labelsByOSSName[si.ServiceInfo.OSSService.ReferenceResourceName] = mainLabel
				addEquivalency(mainLabel, label)
			}
		}
		for _, ch := range sne.CHEntries {
			label := labelsByCHID[ch.CHEntry.DeliverableID]
			switch label {
			case 0:
				// no label yet -- give it the main label
				labelsByCHID[ch.CHEntry.DeliverableID] = mainLabel
			case mainLabel:
				// already has the right label -- do nothing
			default:
				// has a different label -- replace with the main label (which is the lowest number) and record equivalency between these labels
				labelsByCHID[ch.CHEntry.DeliverableID] = mainLabel
				addEquivalency(mainLabel, label)
			}
		}
	}
	debug.Debug(debug.Fine, `mergeNameGroups(): completed first pass; allocated %d labels across %d sharedNamesEntry records`, lastLabel, len(allSharedNames))

	// Second pass: process the equivalency classes between labels, which might span multiple sharedNamesEntry records.
	// Create one NameGroup entry for each equivalency class
	debug.Debug(debug.Fine, `mergeNameGroups(): starting second pass`)
	for name, label := range labelsByOSSName {
		rootLabel := findRootLabel(label)
		if label != rootLabel {
			labelsByOSSName[name] = rootLabel
		}
		if ng, found := nameGroupsByLabel[rootLabel]; found {
			ng.addOSSName(name, traceSharedNamesByLabel[label])
		} else {
			ng = newNameGroup()
			ng.addOSSName(name, traceSharedNamesByLabel[label])
			nameGroupsByLabel[rootLabel] = ng
		}
	}
	for chid, label := range labelsByCHID {
		rootLabel := findRootLabel(label)
		if label != rootLabel {
			labelsByCHID[chid] = rootLabel
		}
		if ng, found := nameGroupsByLabel[rootLabel]; found {
			ng.addCHID(chid, traceSharedNamesByLabel[label])
		} else {
			ng = newNameGroup()
			ng.addCHID(chid, traceSharedNamesByLabel[label])
			nameGroupsByLabel[rootLabel] = ng
		}
	}
	/*
		for _, sne := range allSharedNames {
			for name, la := range sne.ServiceInfos {
				label := labelsByOSSName[si.ServiceInfo.OSSService.ReferenceResourceName]
				rootLabel := findRootLabel(label)
				if label != rootLabel {
					labelsByOSSName[si.ServiceInfo.OSSService.ReferenceResourceName] = rootLabel
				}
				if ng, found := nameGroupsByLabel[rootLabel]; found {
					ng.addOSSName(si.ServiceInfo)
				} else {
					ng = newNameGroup()
					ng.addOSSName(si.ServiceInfo)
					nameGroupsByLabel[rootLabel] = ng
				}
			}
			for _, ch := range sne.CHEntries {
				label := labelsByCHID[ch.CHEntry.DeliverableID]
				rootLabel := findRootLabel(label)
				if label != rootLabel {
					labelsByCHID[ch.CHEntry.DeliverableID] = rootLabel
				}
				if ng, found := nameGroupsByLabel[rootLabel]; found {
					ng.addCHID(ch.CHEntry)
				} else {
					ng = newNameGroup()
					ng.addCHID(ch.CHEntry)
					nameGroupsByLabel[rootLabel] = ng
				}
			}
		}
	*/
	debug.Debug(debug.Fine, `mergeNameGroups(): completed second pass; found %d label equivalency classes / NameGroups`, len(nameGroupsByLabel))

	return nil
}

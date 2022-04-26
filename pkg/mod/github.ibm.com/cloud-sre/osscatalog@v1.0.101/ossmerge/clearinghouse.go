package ossmerge

import (
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

var servicesByClearingHouseID = make(map[string][]*ServiceInfo)

// addCHEntry adds one ClearingHouse entry to the list associated with this ServiceInfo record
func (si *ServiceInfo) addCHEntry(e *clearinghouse.CHSummaryEntry, tags ...string) {
	pinfo := &si.OSSService.ProductInfo
	isNew := pinfo.ClearingHouseReferences.AddClearingHouseReference(e.Name, string(e.DeliverableID), tags...)
	if isNew {
		if nameIssues, found := nameIssuesByCHID[e.DeliverableID]; found {
			for _, nameIssue := range nameIssues {
				si.OSSValidation.AddIssuePreallocated(nameIssue)
			}
		}
		if e.CRNServiceName != "" {
			if e.CRNServiceName == string(si.OSSService.ReferenceResourceName) {
				// Good case: the CRN service-name in the CH entry corresponds to the one for the OSS record to which we are associating the CH entry
			} else {
				si1, found := LookupService(MakeComparableName(e.CRNServiceName), false)
				if !found || e.CRNServiceName != string(si1.OSSService.ReferenceResourceName) {
					si.OSSValidation.AddNamedIssue(ossvalidation.CHBadCRN, e.String())
					// We do not differentiate between complete miss or CH entry that does not use the canonical CRN service-name
				}
			}
		} else {
			si.OSSValidation.AddNamedIssue(ossvalidation.CHMissingCRN, e.String())
		}
		if nameGroup, ok := LookupNameGroupByCHID(e.DeliverableID); ok {
			var found bool
			for _, n := range nameGroup.OSSNames {
				if n == si.ReferenceResourceName {
					found = true
					break
				}
			}
			if !found {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHMismatchNameGroup, e.String())
			}
		} else {
			si.OSSValidation.AddNamedIssue(ossvalidation.CHNoNameGroup, e.String())
		}
	}
}

// mergeClearingHouse links and merges the content of any ClearingHouse entries found to be associated with this ServiceInfo
// This function considers both the CH entries found trough matching PIDs and through matching of similar service names
func (si *ServiceInfo) mergeClearingHouse() {
	si.checkMergePhase(mergePhaseServicesTwo)

	pinfo := &si.OSSService.ProductInfo
	if clearinghouse.HasCHInfo() {
		// Match ClearingHouse entries by their CRN service-name attribute
		if entries, found := clearinghouse.LookupSummaryEntryByCRNServiceName(string(si.OSSService.ReferenceResourceName), si.String()); found {
			if len(entries) == 0 {
				panic(fmt.Sprintf("MergeClearingHouse(%s): CRNServiceName(%s) found but ClearingHouse entry list is empty", si.String(), si.OSSService.ReferenceResourceName))
			}
			for _, e := range entries {
				si.addCHEntry(e, ossrecord.ClearingHouseReferenceTagSourceCRNAttribute)
			}
		}

		// Match ClearingHouse entries by PID
		if pinfo.ProductIDSource != ossrecord.ProductIDSourceCloudPlatform && pinfo.ProductIDSource != ossrecord.ProductIDSourceParent {
			for _, pid := range pinfo.ProductIDs {
				if pid == ossrecord.ProductInfoNone {
					continue
				}
				if entries, found := clearinghouse.LookupSummaryEntryByPID(pid, si.String()); found {
					if len(entries) == 0 {
						panic(fmt.Sprintf("MergeClearingHouse(%s): PID(%s) found but ClearingHouse entry list is empty", si.String(), pid))
					}
					for _, e := range entries {
						si.addCHEntry(e, ossrecord.ClearingHouseReferenceTagSourcePID)
					}
				}
			}
		} else {
			si.AddValidationIssue(ossvalidation.INFO, `Not looking for ClearingHouse links based on ProductIDs that are only inherited from another entry`, "ProductIDSource=%s", pinfo.ProductIDSource).TagProductInfo()
		}

		// Match ClearingHouse entries by similar names -- only if we found no other ClearingHouse linkages
		// (because this method is not very reliable)
		if len(pinfo.ClearingHouseReferences) == 0 {
			if ng, found := LookupNameGroupByOSSName(si.OSSService.ReferenceResourceName); found {
				if len(ng.OSSNames) > 1 {
					issue := si.OSSValidation.AddNamedIssue(ossvalidation.CHMultipleOSS, fmt.Sprintf("%q", ng.OSSNames))
					debug.PlainLogEntry(debug.LevelINFO, si.String(), "Merging ProductInfo information for Service %s: %s\n  Sources:\n%s", si.String(), issue.GetText(), ng.dumpTraceSharedNames("  ", "  "))
				}
				for _, chid := range ng.CHIDs {
					if e, found := clearinghouse.LookupSummaryEntryByID(clearinghouse.DeliverableID(chid), si.String()); found {
						si.addCHEntry(e, ossrecord.ClearingHouseReferenceTagSourceNames)
					} else {
						panic(fmt.Sprintf("MergeClearingHouse(%s): ClearingHouse ID %s not found", si.String(), chid))
					}
				}
			} else {
				//			issue := si.AddValidationIssue(ossvalidation.CRITICAL, "Entry not found in any NameGroup", "").TagProductInfo()
				//			debug.PrintError(strings.TrimSpace(issue.String()))
				panic(fmt.Sprintf(`Entry not found in any NameGroup: %s`, si.OSSService.ReferenceResourceName))
			}
		}

		// Generate validation issues related to ClearingHouse matching
		if len(pinfo.ClearingHouseReferences) == 0 {
			if len(pinfo.ProductIDs) > 0 && !ossrecord.IsProductInfoNone(pinfo.ProductIDs) {
				si.AddValidationIssue(ossvalidation.WARNING, "Entry not found in ClearingHouse, despite having some known productIDs", "productIDs=%v", pinfo.ProductIDs).TagProductInfo()
			}
		} else {
			if len(pinfo.ClearingHouseReferences) > 1 {
				buf := strings.Builder{}
				for i := range pinfo.ClearingHouseReferences {
					buf.WriteString(clearinghouse.MakeCHLabel(pinfo.ClearingHouseReferences[i].Name, clearinghouse.DeliverableID(pinfo.ClearingHouseReferences[i].ID)))
				}
				si.AddValidationIssue(ossvalidation.WARNING, "Found multiple ClearingHouse records for this entry", buf.String()).TagProductInfo()
			}
			// Check for differences between the list of ClearingHouseIDs found through  various sources
			var foundCRN, crnOnly, crnNot, pidOnly, pidNot, nameOnly, nameNot []string
			for i := range pinfo.ClearingHouseReferences {
				if pinfo.ClearingHouseReferences[i].FindTag(ossrecord.ClearingHouseReferenceTagSourceCRNAttribute) != "" {
					foundCRN = append(foundCRN, pinfo.ClearingHouseReferences[i].Name)
					if pinfo.ClearingHouseReferences[i].FindTag(ossrecord.ClearingHouseReferenceTagSourcePID) != "" {
						if pinfo.ClearingHouseReferences[i].FindTag(ossrecord.ClearingHouseReferenceTagSourceNames) != "" {
						} else {
							nameNot = append(nameNot, pinfo.ClearingHouseReferences[i].Name)
						}
					} else {
						pidNot = append(pidNot, pinfo.ClearingHouseReferences[i].Name)
						if pinfo.ClearingHouseReferences[i].FindTag(ossrecord.ClearingHouseReferenceTagSourceNames) != "" {
						} else {
							crnOnly = append(crnOnly, pinfo.ClearingHouseReferences[i].Name)
							nameNot = append(nameNot, pinfo.ClearingHouseReferences[i].Name)
						}
					}
				} else {
					crnNot = append(crnNot, pinfo.ClearingHouseReferences[i].Name)
					if pinfo.ClearingHouseReferences[i].FindTag(ossrecord.ClearingHouseReferenceTagSourcePID) != "" {
						if pinfo.ClearingHouseReferences[i].FindTag(ossrecord.ClearingHouseReferenceTagSourceNames) != "" {
						} else {
							pidOnly = append(pidOnly, pinfo.ClearingHouseReferences[i].Name)
							nameNot = append(nameNot, pinfo.ClearingHouseReferences[i].Name)
						}
					} else {
						pidNot = append(pidNot, pinfo.ClearingHouseReferences[i].Name)
						if pinfo.ClearingHouseReferences[i].FindTag(ossrecord.ClearingHouseReferenceTagSourceNames) != "" {
							nameOnly = append(nameOnly, pinfo.ClearingHouseReferences[i].Name)
						} else {
							nameNot = append(nameNot, pinfo.ClearingHouseReferences[i].Name)
						}
					}
				}
			}
			if len(foundCRN) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHFoundByCRN, "%q", foundCRN)
			}
			if len(pidOnly) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHSomeFoundByPIDOnly, "%q", pidOnly)
			}
			if len(pidNot) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHNotFoundByPID, "%q", pidNot)
			}
			if len(nameOnly) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHSomeFoundByNameOnly, "%q", nameOnly)
			}
			if len(nameNot) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHNotFoundByName, "%q", nameNot)
			}
			if len(crnOnly) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHSomeFoundByCRNOnly, "%q", crnOnly)
			}
			if len(crnNot) > 0 {
				si.OSSValidation.AddNamedIssue(ossvalidation.CHNotFoundByCRN, "%q", crnNot)
			}
		}

		// Fill-in missing ProductIDs obtained from ClearingHouse references
		// (ignore values for references that were found *through* a ProductID)
		var pidsFromCHNames = collections.NewStringSet()
		var pidsFromCHCRNs = collections.NewStringSet()
		for _, chref := range pinfo.ClearingHouseReferences {
			if e, err := clearinghouse.GetFullRecordByID(clearinghouse.DeliverableID(chref.ID)); err == nil {
				for _, pid := range e.PidNumber {
					pid = ossrecord.NormalizeProductID(pid)
					var found bool
					for _, pid0 := range pinfo.ProductIDs {
						if pid0 == pid {
							found = true
							break
						}
					}
					if !found {
						if chref.FindTag(ossrecord.ClearingHouseReferenceTagSourceCRNAttribute) != "" {
							pidsFromCHCRNs.Add(pid)
						} else if chref.FindTag(ossrecord.ClearingHouseReferenceTagSourceNames) != "" {
							pidsFromCHNames.Add(pid)
						}
					}
				}
			} else {
				issue := ossvalidation.NewIssue(ossvalidation.CRITICAL, "ProductInfo contains an invalid ClearingHouse reference", "%s: %v", clearinghouse.MakeCHLabel(chref.Name, clearinghouse.DeliverableID(chref.ID)), err).TagRunAction(ossrunactions.ProductInfoClearingHouse)
				debug.Warning("mergeClearingHouse(service=%s): %s", si.String(), issue.String())
			}
		}
		addCHProductIDs(si, pidsFromCHCRNs, pidsFromCHNames)

		// Find the best Taxonomy info across the available ClearingHouse entries
		if len(pinfo.ClearingHouseReferences) == 1 {
			pinfo.Taxonomy = *getTaxonomy(&pinfo.ClearingHouseReferences[0], si)
		} else if len(pinfo.ClearingHouseReferences) > 1 {
			var found bool
			// First look for a CH reference based on CRN (in alphabetical order)
			for _, chref := range pinfo.ClearingHouseReferences {
				chref := chref
				if chref.FindTag(ossrecord.ClearingHouseReferenceTagSourceCRNAttribute) == "" {
					continue
				}
				t := getTaxonomy(&chref, si)
				if t.IsValid() {
					pinfo.Taxonomy = *t
					found = true
					break
				}
			}
			if !found {
				// Second look for CH reference based on PID (in alphabetical order)
				for _, chref := range pinfo.ClearingHouseReferences {
					if chref.FindTag(ossrecord.ClearingHouseReferenceTagSourcePID) == "" {
						continue
					}
					chref := chref
					t := getTaxonomy(&chref, si)
					if t.IsValid() {
						pinfo.Taxonomy = *t
						found = true
						break
					}
				}
			}
			if !found {
				// Last ... any CH reference will do (in alphabetical order)
				for _, chref := range pinfo.ClearingHouseReferences {
					chref := chref
					t := getTaxonomy(&chref, si)
					if t.IsValid() {
						pinfo.Taxonomy = *t
						found = true
					}
				}
			}
			if found {
				// Look for conflicts in multiple CH references
				conflicts := collections.NewStringSet()
				for _, chref := range pinfo.ClearingHouseReferences {
					chref := chref
					t := getTaxonomy(&chref, si)
					if pinfo.Taxonomy != *t && t.IsValid() {
						conflicts.Add(fmt.Sprintf("%+v", t))
					}
				}
				for _, c := range conflicts.Slice() {
					si.AddValidationIssue(ossvalidation.WARNING, "Found multiple values for Taxonomy in ClearingHouse (first found prevails)", "val1=%+v   val2=%s", &pinfo.Taxonomy, c).TagProductInfo()
				}
			}
		} else {
			// No ClearingHouse entries -> leave Taxonomy empty
		}
	} else {
		// Copy info from Prior OSS record, if any
		if si.HasPriorOSS() && len(si.GetPriorOSS().ProductInfo.ClearingHouseReferences) > 0 {
			si.AddValidationIssue(ossvalidation.INFO, "ProductInfo ClearingHouse-related information copied over from old values -- no refresh in this run", "").TagProductInfo()
			prior := si.GetPriorOSS().ProductInfo
			pinfo.ClearingHouseReferences = prior.ClearingHouseReferences
			pinfo.Taxonomy = prior.Taxonomy
			switch prior.ProductIDSource {
			case ossrecord.ProductIDSourceCHCRN:
				addCHProductIDs(si, collections.NewStringSet(prior.ProductIDs...), collections.NewStringSet())
			case ossrecord.ProductIDSourceCHName:
				addCHProductIDs(si, collections.NewStringSet(), collections.NewStringSet(prior.ProductIDs...))
			default:
				// Just for sanity check
				addCHProductIDs(si, collections.NewStringSet(), collections.NewStringSet())
			}
		}
	}
	for _, chref := range pinfo.ClearingHouseReferences {
		if chref.ID == ossrecord.ProductInfoNone {
			continue
		}
		servicesByClearingHouseID[chref.ID] = append(servicesByClearingHouseID[chref.ID], si)
		debug.Debug(debug.ClearingHouse, "ossmerge.mergeClearingHouse(%s): Recording ClearingHouseID to ServiceInfo mapping: %s -> %s", si.String(), chref.ID, si.String())
	}
}

// LookupServicesByCHID returns the ServiceInfo record that contains a particular ClearingHouse DeliverableID
func LookupServicesByCHID(chid clearinghouse.DeliverableID) (sis []*ServiceInfo, food bool) {
	sis, found := servicesByClearingHouseID[string(chid)]
	return sis, found
}

func getTaxonomy(chref *ossrecord.ClearingHouseReference, si *ServiceInfo) *ossrecord.Taxonomy {
	var e *clearinghouse.CHDeliverableWithDependencies
	var err error
	if e, err = clearinghouse.GetFullRecordByID(clearinghouse.DeliverableID(chref.ID)); err == nil {
		return &e.CopiedTaxonomy
	}
	issue := ossvalidation.NewIssue(ossvalidation.CRITICAL, "Invalid ClearingHouse reference while fetching Taxonomy info", "%s: %v", clearinghouse.MakeCHLabel(chref.Name, clearinghouse.DeliverableID(chref.ID)), err).TagRunAction(ossrunactions.ProductInfoClearingHouse)
	debug.Warning("getTaxonomy(service=%s): %s", si.String(), issue.String())
	return &ossrecord.Taxonomy{}
}

func addCHProductIDs(si *ServiceInfo, pidsFromCHCRNs, pidsFromCHNames collections.StringSet) {
	pinfo := &si.OSSService.ProductInfo
	if len(pinfo.ProductIDs) == 0 || ossrecord.IsProductInfoNone(pinfo.ProductIDs) {
		if pinfo.ProductIDSource != "" {
			panic(fmt.Sprintf(`mergeClearingHouse(%s): ProductIDSource is not empty (%s) but ProductIDs is empty (%v)`, si.String(), pinfo.ProductIDSource, pinfo.ProductIDs))
		}
	} else {
		if pinfo.ProductIDSource == "" {
			panic(fmt.Sprintf(`mergeClearingHouse(%s): ProductIDSource is empty but ProductIDs not empty (%v)`, si.String(), pinfo.ProductIDs))
		}
	}
	if pidsFromCHCRNs.Len() > 0 {
		if pinfo.ProductIDSource == "" {
			pinfo.ProductIDs = pidsFromCHCRNs.Slice()
			pinfo.ProductIDSource = ossrecord.ProductIDSourceCHCRN
		} else {
			for _, pid := range pidsFromCHCRNs.Slice() {
				found := false
				for _, pid0 := range pinfo.ProductIDs {
					if pid == pid0 {
						found = true
						break
					}
				}
				if !found {
					si.AddValidationIssue(ossvalidation.SEVERE, "ProductID found through ClearingHouse CRN Attribute missing from other source", "pid=%s  other_source=%s", pid, pinfo.ProductIDSource).TagProductInfo()
				}
			}
		}
	}
	if pidsFromCHNames.Len() > 0 {
		if pinfo.ProductIDSource == "" {
			pinfo.ProductIDs = pidsFromCHNames.Slice()
			pinfo.ProductIDSource = ossrecord.ProductIDSourceCHName
			si.AddValidationIssue(ossvalidation.MINOR, "ProductIDs only found indirectly through matching of ClearingHouse names -- may not be correct", "%v", pinfo.ProductIDs).TagProductInfo()
		} else {
			for _, pid := range pidsFromCHCRNs.Slice() {
				found := false
				for _, pid0 := range pinfo.ProductIDs {
					if pid == pid0 {
						found = true
						break
					}
				}
				if !found {
					si.AddValidationIssue(ossvalidation.SEVERE, "Ignoring additional ProductID found through matching of ClearingHouse names", "pid=%s  other_source=%s", pid, pinfo.ProductIDSource).TagProductInfo()
				}
			}
		}
	}
}

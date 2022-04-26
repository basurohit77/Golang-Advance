package ossmerge

import (
	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

const kindComposite = `composite`

// checkComposite examines Composite objects in Catalog and their children
func (si *ServiceInfo) checkComposite() {
	si.checkMergePhase(mergePhaseServicesTwo)

	if !si.HasSourceMainCatalog() {
		return
	}

	if si.GetSourceMainCatalog().Kind == kindComposite {
		// Parent Composite
		if base, _ := ParseCompositeName(si.GetSourceMainCatalog().Name); base != "" {
			si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite has a name that itself a composite child name`, "name=%s", si.GetSourceMainCatalog().Name).TagCRN().TagCatalogComposite()
		}
		comp := si.GetSourceMainCatalog().ObjectMetaData.Other.Composite
		if comp == nil {
			si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry is of kind=composite but does not contain a ObjectMetaData.Other.Composite section`, "").TagCRN().TagCatalogComposite()
			return
		}
		if comp.CompositeKind == kindComposite {
			si.AddValidationIssue(ossvalidation.CRITICAL, `Main Catalog entry is of kind=composite) but its composite_kind (for child entries) is itself "composite" (not supported by this tool)`, "").TagCRN().TagCatalogComposite()
		}
		for _, childEntry := range comp.Children {
			childInfo, found := LookupService(MakeComparableName(childEntry.Name), false)
			base, _ := ParseCompositeName(childEntry.Name)
			if base != si.GetSourceMainCatalog().Name {
				si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite has a child entry name that is not formed using the Composite parent name as a base`, "child=%s   parent=%s", childEntry.Name, si.GetSourceMainCatalog().Name).TagCRN().TagCatalogComposite()
			}
			if childEntry.Kind != comp.CompositeKind {
				si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite has a child entry Kind that is not the same as the main composite_kind`, "child=%s   composite_kind=%s   child.kind=%s", childEntry.Name, comp.CompositeKind, childEntry.Kind).TagCRN().TagCatalogComposite()
			}
			if !found || childInfo.OSSValidation.NumTrueSources() == 0 {
				si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite references a child object that is not found in any source`, "child=%s", childEntry.Name).TagCRN().TagCatalogComposite()
				continue
			}
			issues := 0
			if childInfo.mergeWorkArea.compositeParent != "" {
				if childInfo.mergeWorkArea.compositeParent == si.OSSService.ReferenceResourceName {
					childInfo.AddValidationIssue(ossvalidation.CRITICAL, "Main Catalog entry is referenced as a child more than once from the same Composite parent", `parent="%s"`, childInfo.mergeWorkArea.compositeParent).TagCRN().TagCatalogComposite()
				} else {
					childInfo.AddValidationIssue(ossvalidation.CRITICAL, "Main Catalog entry is referenced as a child of more than one Composite parent", `parent1="%s"   parent2="%s"`, childInfo.mergeWorkArea.compositeParent, si.OSSService.ReferenceResourceName).TagCRN().TagCatalogComposite()
				}
				issues++
			} else {
				childInfo.mergeWorkArea.compositeParent = si.OSSService.ReferenceResourceName
				debug.Debug(debug.Composite, `Found composite child in parent %s -> %s`, si.OSSService.ReferenceResourceName, childInfo.OSSService.ReferenceResourceName)
			}
			if !childInfo.HasSourceMainCatalog() {
				si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite references a child object found in some sources but not in the Catalog iself`, "child=%s   sources=%v%v", childEntry.Name, childInfo.OSSValidation.CanonicalNameSources, childInfo.OSSValidation.OtherNamesSources).TagCRN().TagCatalogComposite()
				childInfo.AddValidationIssue(ossvalidation.SEVERE, `Entry is a child of a Composite entry in the Catalog but is not itself in the Catalog`, `parent_name="%s"`, si.OSSService.ReferenceResourceName).TagCRN().TagCatalogComposite()
				/*
					if debug.IsDebugEnabled(debug.Composite) {
						buf, _ := json.MarshalIndent(childInfo, "    ", "    ")
						debug.Debug(debug.Composite, "checkComposite(%s,%s): child found but not in Catalog: childEntry=%+v   childServiceInfo=%s", si.String(), childEntry, &childEntry, buf)
					}
				*/
				continue
			}
			if childInfo.GetSourceMainCatalog().Name != childEntry.Name {
				si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite has a child object that is found but whose name is not an exact match`, "expected=%s   got=%s", childEntry.Name, childInfo.GetSourceMainCatalog().Name).TagCRN().TagCatalogComposite()
				childInfo.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry is a child of a Composite entry but its name is not an exact match for the reference in the parent", `parent_name="%s"  actual_child_name="%s"   expected_child_name="%s"`, si.OSSService.ReferenceResourceName, childInfo.GetSourceMainCatalog().Name, childEntry.Name).TagCRN().TagCatalogComposite()
				issues++
			}
			if base2, _ := ParseCompositeName(childInfo.GetSourceMainCatalog().Name); base2 != si.GetSourceMainCatalog().Name {
				childInfo.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry is a child of a Composite entry but its name is not formed using the Composite parent name as a base", `parent_name="%s"  child_name="%s"  child_base_name="%s"`, si.OSSService.ReferenceResourceName, childInfo.GetSourceMainCatalog().Name, base2).TagCRN().TagCatalogComposite()
				issues++
			}
			if childInfo.GetSourceMainCatalog().Kind != childEntry.Kind {
				si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite has a child object that is not of the expected Kind`, "child=%s   expected_kind=%s   got=%s", childEntry.Name, childEntry.Kind, childInfo.GetSourceMainCatalog().Kind).TagCRN().TagCatalogComposite()
				childInfo.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry is a child of a Composite entry but does not have the expected Kind as specified in the parent", `parent_name="%s"  child_kind="%s"   expected_kind="%s"`, si.OSSService.ReferenceResourceName, childInfo.GetSourceMainCatalog().Kind, childEntry.Kind).TagCRN().TagCatalogComposite()
				issues++
			}
			if !catalog.SearchTags(childInfo.GetSourceMainCatalog(), comp.CompositeTag) {
				si.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry of kind=composite has a child object that does not contain the expected composite_tag`, "child=%s   expected_tag=%s", childEntry.Name, comp.CompositeTag).TagCRN().TagCatalogComposite()
				childInfo.AddValidationIssue(ossvalidation.SEVERE, `Main Catalog entry is a child of a Composite entry but does not contain the expected composite_tag as specified in the parent`, `parent_name="%s"  expected_tag="%s"`, si.OSSService.ReferenceResourceName, comp.CompositeTag).TagCRN().TagCatalogComposite()
				issues++
			}
			if issues == 0 {
				childInfo.AddValidationIssue(ossvalidation.INFO, `Main Catalog entry is a child of a Composite entry -- no issues`, "").TagCRN().TagCatalogComposite()
			}
		}
	} else if base, _ := ParseCompositeName(si.GetSourceMainCatalog().Name); base != "" {
		// Child of a Composite parent
		parent, found := LookupService(MakeComparableName(base), false)
		if !found {
			si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry looks like a child of a Composite entry but the parent Composite is not found from any sources", `parent="%s"`, base).TagCRN().TagCatalogComposite()
			return
		}
		if !parent.HasSourceMainCatalog() {
			si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry looks like a child of a Composite entry but the parent Composite is found in some sources but not in the Catalog iself", `parent="%s"     sources=%v%v"`, base, parent.OSSValidation.CanonicalNameSources, parent.OSSValidation.OtherNamesSources).TagCRN().TagCatalogComposite()
			return
		}
		if parent.GetSourceMainCatalog().Kind != kindComposite {
			si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry looks like a child of a Composite entry but the parent is not actually of kind=composite", `parent="%s"  parent.kind="%s"`, parent.GetSourceMainCatalog().Name, parent.GetSourceMainCatalog().Kind).TagCRN().TagCatalogComposite()
			return
		}
		comp := parent.GetSourceMainCatalog().ObjectMetaData.Other.Composite
		if comp == nil {
			si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry looks like a child of a Composite entry but the parent Composite does not contain a ObjectMetaData.Other.Composite section", `parent="%s"  parent.kind="%s"`, parent.GetSourceMainCatalog().Name, parent.GetSourceMainCatalog().Kind).TagCRN().TagCatalogComposite()
			return
		}
		var childFound int
		for _, childEntry := range comp.Children {
			// TODO: should we check for fuzzy matches that would be found through the "comparable name" logic even though not an exact match (we do in the other direction, from parent to child...)
			if si.GetSourceMainCatalog().Name == childEntry.Name {
				childFound++
			}
		}
		if childFound == 0 {
			si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry looks like a child of a Composite entry but the parent Composite does not have a reference to this child", `parent_name="%s"  child_name="%s"`, parent.GetSourceMainCatalog().Name, si.GetSourceMainCatalog().Name).TagCRN().TagCatalogComposite()
			if parent.GetSourceMainCatalog().Name != base {
				// If the child was found in the parent, this warning would have been generated from the parent
				si.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry looks like a child of a Composite entry but the parent Composite name is not an exact match (and does not contain a reference to this child)", `expected_parent="%s"  actual_parent="%s"`, base, parent.GetSourceMainCatalog().Name).TagCRN().TagCatalogComposite()
			}
		} else if childFound > 1 {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Main Catalog entry is a child of a Composite entry but the parent Composite contains more than one child entry with the same name", `parent_name="%s"  matches=%d`, parent.GetSourceMainCatalog().Name, childFound).TagCRN().TagCatalogComposite()
		}
		if si.HasSourceRMC() && si.GetSourceRMC().ParentCompositeService.Name != string(parent.ReferenceResourceName) {
			si.AddValidationIssue(ossvalidation.CRITICAL, "Main Catalog entry looks like a child of a Composite entry but the parent does not match the RMC entry", `Catalog_parent="%s"  RMC_parent="%s"`, parent.ReferenceResourceName, si.GetSourceRMC().ParentCompositeService.Name).TagCRN().TagCatalogComposite().TagRunAction(ossrunactions.RMC)
		}
	} else if si.HasSourceRMC() && si.GetSourceRMC().ParentCompositeService.Name != "" {
		si.AddValidationIssue(ossvalidation.CRITICAL, "RMC indicates this entry is a child of a Composite entry but entry in the Catalog does not look like a child of Composite", `RMC_parent="%s"`, si.GetSourceRMC().ParentCompositeService.Name).TagCRN().TagCatalogComposite().TagRunAction(ossrunactions.RMC)
		// Force the compositeParent anyway, to set the ParentResourceName field later.
	}

}

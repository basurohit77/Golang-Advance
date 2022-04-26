package ossmerge

import (
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// checkIAM examines the IAM related flags and registrations
func (si *ServiceInfo) checkIAM() {
	si.checkMergePhase(mergePhaseServicesOne)

	if si.HasSourceIAM() {
		if si.GetSourceIAM().Enabled {
			if si.HasSourceMainCatalog() {
				cat := si.GetSourceMainCatalog()
				if cat.ObjectMetaData.Service != nil && cat.ObjectMetaData.Service.IAMCompatible {
					// OK
				} else {
					si.AddValidationIssue(ossvalidation.WARNING, "Service is registered in IAM but the corresponding entry in Global Catalog has IAM_Compatible=false", "%s", si.GetSourceIAM().String()).TagIAM()
				}
			} else {
				si.AddValidationIssue(ossvalidation.WARNING, "Service is registered in IAM but has no corresponding entry in Global Catalog", "%s", si.GetSourceIAM().String()).TagIAM()
			}
		} else {
			if si.HasSourceMainCatalog() {
				cat := si.GetSourceMainCatalog()
				if cat.ObjectMetaData.Service != nil && cat.ObjectMetaData.Service.IAMCompatible {
					si.AddValidationIssue(ossvalidation.WARNING, "Service is registered but not enabled in IAM but the corresponding entry in Global Catalog has IAM_Compatible=false", "%s", si.GetSourceIAM().String()).TagIAM()
				} else {
					si.AddValidationIssue(ossvalidation.MINOR, "Service is registered but not enabled in IAM and the corresponding entry in Global Catalog has IAM_Compatible=false", "%s", si.GetSourceIAM().String()).TagIAM()
				}
			} else {
				si.AddValidationIssue(ossvalidation.INFO, "Service is registered but not enabled in IAM and has no corresponding entry in Global Catalog", "%s", si.GetSourceIAM().String()).TagIAM()
			}
		}
	} else {
		if si.HasSourceMainCatalog() {
			cat := si.GetSourceMainCatalog()
			if cat.ObjectMetaData.Service != nil && cat.ObjectMetaData.Service.IAMCompatible {
				si.AddValidationIssue(ossvalidation.SEVERE, "Service is NOT registered in IAM but the corresponding entry in Global Catalog has IAM_Compatible=true", "").TagIAM()
			}
		}
	}
}

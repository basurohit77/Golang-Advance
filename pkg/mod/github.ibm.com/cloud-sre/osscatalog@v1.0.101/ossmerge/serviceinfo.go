package ossmerge

// Functions and declarations for the data model in the oss-catalog tool

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"

	"github.ibm.com/cloud-sre/osscatalog/rmc"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/legacy"
	"github.ibm.com/cloud-sre/osscatalog/monitoringinfo"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/supportcenter"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/iam"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/scorecardv1"
	"github.ibm.com/cloud-sre/osscatalog/servicenow"
)

// ServiceInfo is the main data structure for the data model in the oss-catalog tool.
// There is one record for each service or component for which we maintain OSS information.
// The main element is a OSSService structure, which represents the record that represents each service or component in the Global Catalog
//
// NOTE that we use "inline" versions of the member structs rather than pointers, to simplify conditional access to
// individual attributes in situations where one or more of these sources is not available.
// We provide special "HasX()" methods to check for availability, rather than checking for a nil pointer
type ServiceInfo struct {
	ComparableName                       string                          // The "comparable name" for this entry, which folds multiple possible variations of the names from various sources
	ossrecordextended.OSSServiceExtended                                 // The full OSS record under construction
	PriorOSS                             ossrecord.OSSService            // Prior OSSService entry for this service/component, found in Catalog
	PriorOSSValidation                   *ossvalidation.OSSValidation    // Prios OSSValidation entry for this service/component, found in Catalog
	PriorOSSValidationChecksum           string                          // Checksum for the prior OSSValidation entry for this service/component, if any, found in Catalog
	SourceMainCatalog                    catalogapi.Resource             // Main entry for this service/component in Global Catalog (not necessarily the entry containing the OSSService metadata)
	SourceServiceNow                     servicenow.ConfigurationItem    // Original entry for this service/component in ServiceNow
	SourceScorecardV1Detail              scorecardv1.DetailEntry         // Original entry for this service/component in ScorecardV1 (internal API)
	AdditionalMainCatalog                []*catalogapi.Resource          // Additional entries from Global Catalog, that map to the same comparableName
	AdditionalServiceNow                 []*servicenow.ConfigurationItem // Additional entries from ServiceNow, that map to the same comparableName
	AdditionalScorecardV1Detail          []*scorecardv1.DetailEntry      // Additional entries from ScorecardV1 (internal API), that map to the same comparableName
	IgnoredMainCatalog                   *catalogapi.Resource            // Saved copy of the original SourceMainCatalog, in case we decide to ignore it because it is not visible (used for some reports)
	SourceIAM                            iam.Service                     // Main entry for this service in IAM (if any)
	AdditionalIAM                        []*iam.Service                  // Additional entries from IAM, that map to the same comparableName
	SourceRMC                            rmc.SummaryEntry                // Main entry for this service in RMC (if any)
	AdditionalRMC                        []*rmc.SummaryEntry             // Additional entries from RMC, that map to the same comparableName
	DuplicateOf                          string                          // Name of the ServiceInfo record that this one has been merged into
	Legacy                               *legacy.Entry                   // Info from the legacy, python-based CRN Validation report
	CatalogExtra                         *CatalogExtra                   // Extra informatiom loaded from the Global Catalog, that is not contained in the catalog.Resource (ServiceInfo.SourceMainCatalog) itself
	MonitoringInfo                       *monitoringinfo.MonitoringInfo
	SupportCenterInfo                    *supportcenter.Info
	mergeWorkArea                        struct { // work area for intermediate results from the merge
		skipMerge                      bool                     // cache the result of SkipMerge()
		skipMergeReason                string                   // cache the result of SkipMerge()
		mergePhase                     mergePhaseType           // sanity check to verify if the merging for this record has properly gone through the corect phases in the right order, and been finalized
		pnpStatus                      pnpStatusType            // *private*: to cache the result of checkPnPEnablement()
		pnpCategoryParentIssues        int                      // to record issues with the setting of the StatusPage.CategoryParent
		compositeParent                ossrecord.CRNServiceName // the parent of the child entry of a Catalog entry with kind=composite -- if applicable
		mergeProductInfoPhaseThreeDone bool                     // to manage recursion as we traverse the ParentResourceName links in mergeProductInfoPhaseThree()
	}
	//	SourceScorecardV1Entry	ScorecardV1Entry		// Original entry for this service/component in ScorecardV1 (official API)
}

// HasPriorOSS returns true if the PriorOSS sub-record in this ServiceInfo is present, false otherwise
func (si *ServiceInfo) HasPriorOSS() bool {
	return si.PriorOSS.ReferenceResourceName != ""
}

// GetPriorOSS returns a pointer to the PriorOSS record in this ServiceInfo if it is valid, nil otherwise
func (si *ServiceInfo) GetPriorOSS() *ossrecord.OSSService {
	if si.HasPriorOSS() {
		return &si.PriorOSS
	}
	return nil
}

// HasPriorOSSValidation returns true if the PriorOSSValidation sub-record in this ServiceInfo is present, false otherwise
func (si *ServiceInfo) HasPriorOSSValidation() bool {
	return si.PriorOSSValidation != nil
}

// GetPriorOSSValidation returns a pointer to the PriorOSSValidation record in this ServiceInfo if it is valid, nil otherwise
func (si *ServiceInfo) GetPriorOSSValidation() *ossvalidation.OSSValidation {
	if si.HasPriorOSSValidation() {
		return si.PriorOSSValidation
	}
	return nil
}

// HasSourceMainCatalog returns true if the SourceMainCatalog sub-record in this ServiceInfo is present, false otherwise
func (si *ServiceInfo) HasSourceMainCatalog() bool {
	return si.SourceMainCatalog.Name != ""
}

// GetSourceMainCatalog returns a pointer to the SourceMainCatalog record in this ServiceInfo if it is valid, nil otherwise
func (si *ServiceInfo) GetSourceMainCatalog() *catalogapi.Resource {
	if si.HasSourceMainCatalog() {
		return &si.SourceMainCatalog
	}
	return nil
}

// HasSourceServiceNow returns true if the SourceServiceNow sub-record in this ServiceInfo is present, false otherwise
func (si *ServiceInfo) HasSourceServiceNow() bool {
	return si.SourceServiceNow.CRNServiceName != ""
}

// GetSourceServiceNow returns a pointer to the SourceServiceNow record in this ServiceInfo if it is valid, nil otherwise
func (si *ServiceInfo) GetSourceServiceNow() *servicenow.ConfigurationItem {
	if si.HasSourceServiceNow() {
		return &si.SourceServiceNow
	}
	return nil
}

// HasSourceScorecardV1Detail returns true if the SourceScorecardV1Detail sub-record in this ServiceInfo is present, false otherwise
func (si *ServiceInfo) HasSourceScorecardV1Detail() bool {
	return si.SourceScorecardV1Detail.Name != ""
}

// GetSourceScorecardV1Detail returns a pointer to the SourceScorecardV1Detail sub-record in this ServiceInfo if it is valid, nil otherwise
func (si *ServiceInfo) GetSourceScorecardV1Detail() *scorecardv1.DetailEntry {
	if si.HasSourceScorecardV1Detail() {
		return &si.SourceScorecardV1Detail
	}
	return nil
}

// HasSourceIAM returns true if the SourceIAM sub-record in this ServiceInfo is present, false otherwise
func (si *ServiceInfo) HasSourceIAM() bool {
	return si.SourceIAM.Name != ""
}

// GetSourceIAM returns a pointer to the SourceIAM record in this ServiceInfo if it is valid, nil otherwise
func (si *ServiceInfo) GetSourceIAM() *iam.Service {
	if si.HasSourceIAM() {
		return &si.SourceIAM
	}
	return nil
}

// HasSourceRMC returns true if the SourceRMC sub-record in this ServiceInfo is present, false otherwise
func (si *ServiceInfo) HasSourceRMC() bool {
	return si.SourceRMC.CRNServiceName != ""
}

// GetSourceRMC returns a pointer to the SourceRMC record in this ServiceInfo if it is valid, nil otherwise
func (si *ServiceInfo) GetSourceRMC() *rmc.SummaryEntry {
	if si.HasSourceRMC() {
		return &si.SourceRMC
	}
	return nil
}

// GetServiceName returns the best available user-facing name for this ServiceInfo record
func (si *ServiceInfo) GetServiceName() string {
	if si.OSSService.ReferenceResourceName != "" {
		return string(si.OSSService.ReferenceResourceName)
	}
	if si.PriorOSS.ReferenceResourceName != "" {
		return string(si.PriorOSS.ReferenceResourceName)
	}
	return si.ComparableName
}

// String returns a short string identifier for this ServiceInfo
func (si *ServiceInfo) String() string {
	if si.OSSService.ReferenceResourceName != "" {
		return string(si.OSSService.ReferenceResourceName)
	}
	if si.PriorOSS.ReferenceResourceName != "" {
		return string(si.PriorOSS.ReferenceResourceName)
	}
	return fmt.Sprintf("%s/comparableName", si.ComparableName)
}

var serviceByComparableName = make(map[string]*ServiceInfo)

// LookupService returns the ServiceInfo record associated with a given "comparable name" or creates a new record if appropriate.
// If no record exists for a comprableName and the parameter 'createIfNeeded' is false, 'nil' is returned.
func LookupService(name string, createIfNeeded bool) (si *ServiceInfo, found bool) {
	name1 := MakeComparableName(name)
	si, found = serviceByComparableName[name1]
	if !found && createIfNeeded {
		si := new(ServiceInfo)
		si.OSSValidation = ossvalidation.New("", options.GlobalOptions().LogTimeStamp)
		si.ComparableName = name1
		serviceByComparableName[name1] = si
		return si, true
	}
	return si, found
}

// ListAllServices invokes the handler function on all ServiceInfo records whose name matches the specified pattern
// (or all records, if the pattern is empty)
func ListAllServices(pattern *regexp.Regexp, handler func(si *ServiceInfo)) error {
	for name, si := range serviceByComparableName {
		if pattern == nil || pattern.FindString(name) != "" {
			handler(si)
		}
	}
	return nil
}

type serviceInfoRegistryType string

// serviceInfoRegistrySingleton is the singleton object for access to the registry of ServiceInfo objects managed in this package.
const serviceInfoRegistrySingleton serviceInfoRegistryType = "ServiceInfoRegistrySingleton"

func (registry serviceInfoRegistryType) LookupModel(id string, createIfNeeded bool) (m ossmergemodel.Model, ok bool) {
	si, found := LookupService(id, createIfNeeded)
	return si, found
}

func (registry serviceInfoRegistryType) ListAllModels(pattern *regexp.Regexp, handler func(m ossmergemodel.Model)) error {
	err := ListAllServices(pattern, func(si *ServiceInfo) {
		handler(si)
	})
	return err
}

// ResetForTesting resets any persistent data structures -- used to setup tests
func ResetForTesting() ossmergemodel.ModelRegistry {
	serviceByComparableName = make(map[string]*ServiceInfo)
	return serviceInfoRegistrySingleton
}

// GetDupOfServiceName returns the canonical name of another entry that this entry is a duplicate of
// or the empty string if this entry is not known to be a duplicate of any other entry
func (si *ServiceInfo) GetDupOfServiceName() string {
	if si.DuplicateOf != "" {
		target, found := LookupService(si.DuplicateOf, false)
		if found {
			return target.GetServiceName()
		}
		return si.DuplicateOf
	}
	return ""
}

/*

var serviceBySysid = make(map[ossrecord.ServiceNowSysid]*ServiceInfo)

// RecordServiceNowSysid records the ServiceNow sysid associated with the service represented by the target ServiceInfo record.
// This function checks for duplicates / mismatched sysids
func (si *ServiceInfo) RecordServiceNowSysid(sysid ossrecord.ServiceNowSysid) error {
	if (si.OSSService.ServiceNowSysid != "") && (si.OSSService.ServiceNowSysid != sysid) {
		return fmt.Errorf("found duplicate ServiceNow sysid for entry \"%s\":  existing=%v    new=%v", si.OSSService.ReferenceResourceName, si.OSSService.ServiceNowSysid, sysid)
	}
	existing := serviceBySysid[sysid]
	if (existing != nil) && (existing.OSSService.ServiceNowSysid != sysid) {
		return fmt.Errorf("found two services with the same sysid - name1=\"%s\"   name2=\"%s\"   sysid=%v", existing.OSSService.ReferenceResourceName, si.OSSService.ReferenceResourceName, sysid)
	}
	si.OSSService.ServiceNowSysid = sysid
	serviceBySysid[sysid] = si
	return nil
}
*/

// AddValidationIssue creates a new ValidationIssue record and adds it to the target ServiceInfo record
func (si *ServiceInfo) AddValidationIssue(severity ossvalidation.Severity, title string, detailsFmt string, a ...interface{}) *ossvalidation.ValidationIssue {
	return si.OSSValidation.AddIssue(severity, title, detailsFmt, a...)
}

// AddNamedValidationIssue creates a new ValidationIssue record based on a named pattern and adds it to the target ServiceInfo record
func (si *ServiceInfo) AddNamedValidationIssue(pattern *ossvalidation.NamedValidationIssue, detailsFmt string, a ...interface{}) {
	si.OSSValidation.AddNamedIssue(pattern, detailsFmt, a...)
}

// GetAllServices returns a slice that contains pointers to every ServiceInfo record that matches the given name pattern
func GetAllServices(pattern *regexp.Regexp) ([]*ServiceInfo, error) {
	var result = make([]*ServiceInfo, 0, len(serviceByComparableName))
	for name, si := range serviceByComparableName {
		if pattern.FindString(name) != "" {
			result = append(result, si)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OSSService.ReferenceResourceName < result[j].OSSService.ReferenceResourceName
	})
	return result, nil
}

// IsValid returns true if this ServiceInfo record represents a valid entry that
// can be updated or merged into the Catalog
func (si *ServiceInfo) IsValid() bool {
	si.checkMergePhase(mergePhaseFinalized)
	return si.DuplicateOf == "" && si.OSSService.ReferenceResourceName != ""
}

// IsDeletable returns true if there is no reason to keep this OSS record in the Catalog (i.e it has no source, and no non-zero OSSControl data)
func (si *ServiceInfo) IsDeletable() bool {
	// XXX We must allow this call to be made during the merge itself; especially at the beginning of phase TWO when we merge all the name groups
	// to attempt to match ClearingHouse entries
	si.checkGlobalMergePhaseMultiple(mergePhaseServicesTwo, mergePhaseServicesThree, mergePhaseFinalized)
	if si.DuplicateOf != "" || si.OSSService.ReferenceResourceName == "" || si.OSSServiceExtended.IsDeletable() {
		if !ossrunactions.ScorecardV1.IsEnabled() && si.HasPriorOSS() && si.GetPriorOSS().GeneralInfo.OSSOnboardingPhase == "" {
			if si.OSSValidation != nil {
				si.OSSValidation.AddIssueIgnoreDup(ossvalidation.INFO, `Cannot confirm if entry is deletable when ScorecardV1 is disabled`, ``).TagPriorOSS()
			} else {
				debug.Info(`Cannot confirm if entry is deletable when ScorecardV1 is disabled: %s`, si.String())
			}
			return false
		}
		return true
	}
	return false
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (si *ServiceInfo) IsUpdatable() bool {
	return si.OSSService.IsUpdatable()
}

// Diffs returns a list of differences between the new and priorOSS versions of the OSS record in this ServiceInfo
func (si *ServiceInfo) Diffs() *compare.Output {
	if !si.Finalized() {
		panic(fmt.Sprintf("ServiceInfo(%s) merge has not been finalized", si.String()))
	}
	out := compare.Output{}
	compare.DeepCompare("before", &si.PriorOSS, "after", &si.OSSService, &out)
	if si.OSSValidation.Checksum() != si.PriorOSSValidationChecksum {
		compare.DeepCompare("before.OSSValidation", si.PriorOSSValidation, "after.OSSValidation", si.OSSValidation, &out)
		//out.AddOtherDiff("OSSValidation updated")
	}
	return &out
}

// GlobalSortKey returns a global key for sorting across all ossmerge.Model entries of all types
func (si *ServiceInfo) GlobalSortKey() string {
	return "4." + si.String()
}

// OSSEntry returns the contents of this ServiceInfo as a OSSEntry type
func (si *ServiceInfo) OSSEntry() ossrecord.OSSEntry {
	return &si.OSSServiceExtended
}

// PriorOSSEntry returns pointer to the PriorOSS record in this ServiceInfo as a OSSEntry type, if it is valid, nil otherwise
func (si *ServiceInfo) PriorOSSEntry() ossrecord.OSSEntry {
	return si.GetPriorOSS()
}

// GetAccessor returns the Accessor function for a given sub-record within the ServiceInfo
func (si *ServiceInfo) GetAccessor(t reflect.Type) ossmergemodel.AccessorFunc {
	switch t {
	case monitoringinfo.SubRecordType:
		f := func(m ossmergemodel.Model, createFunc ossmergemodel.CreateFunc) interface{} {
			m1 := m.(*ServiceInfo)
			if m1.MonitoringInfo == nil && createFunc != nil {
				m1.MonitoringInfo = createFunc().(*monitoringinfo.MonitoringInfo)
			}
			return m1.MonitoringInfo
		}
		return f
	case supportcenter.SubRecordType:
		f := func(m ossmergemodel.Model, createFunc ossmergemodel.CreateFunc) interface{} {
			m1 := m.(*ServiceInfo)
			if m1.SupportCenterInfo == nil && createFunc != nil {
				m1.SupportCenterInfo = createFunc().(*supportcenter.Info)
			}
			return m1.SupportCenterInfo
		}
		return f
	default:
		panic(fmt.Sprintf("ossmerge.ServiceInfo.GetAccessor() called with unsupported type %v", t))
	}
}

// GetMonitoringInfo returns the MonitoringInfo sub-record contained in this ServiceInfo
func (si *ServiceInfo) GetMonitoringInfo(createFunc func() *monitoringinfo.MonitoringInfo) *monitoringinfo.MonitoringInfo {
	return (&monitoringinfo.ServiceInfo{Model: si}).GetMonitoringInfo(createFunc)
}

// GetOSSValidation returns the OSSValidation record in this ServiceInfo, or nil if there is none
func (si *ServiceInfo) GetOSSValidation() *ossvalidation.OSSValidation {
	return si.OSSValidation
}

// SkipMerge returns true if we should skip the OSS merge for this record and simply copy it from
// the prior OSS record.
// If not empty, the "reason" string provides details on why this record should or should not be merged
func (si *ServiceInfo) SkipMerge() (skip bool, reason string) {
	if si.mergeWorkArea.skipMergeReason == "" {
		si.mergeWorkArea.skipMerge, si.mergeWorkArea.skipMergeReason = ossmergemodel.SkipMerge(si)
	}
	return si.mergeWorkArea.skipMerge, si.mergeWorkArea.skipMergeReason
}

var _ ossmergemodel.Model = &ServiceInfo{} // verify

var _ ossmergemodel.ModelRegistry = serviceInfoRegistrySingleton // verify

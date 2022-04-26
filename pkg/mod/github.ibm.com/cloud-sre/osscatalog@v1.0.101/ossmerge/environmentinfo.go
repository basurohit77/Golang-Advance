package ossmerge

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/doctor"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// EnvironmentInfo represents the data model for OSS environments information for the oss-catalog tool
type EnvironmentInfo struct {
	ComparableCRNMask                        crn.Mask
	ossrecordextended.OSSEnvironmentExtended                              // The full OSS record under construction
	PriorOSS                                 ossrecord.OSSEnvironment     // Prior OSSEnvironment entry for this environment, found in Catalog
	PriorOSSValidation                       *ossvalidation.OSSValidation // Prios OSSValidation entry for this environment, found in Catalog
	PriorOSSValidationChecksum               string                       // Checksum for the prior OSSValidation entry for this environment, if any, found in Catalog
	SourceMainCatalog                        catalogapi.Resource
	AdditionalPriorOSS                       []*ossrecord.OSSEnvironment // Additional OSS entries that map to the same EnvironmentID
	AdditionalMainCatalog                    []*catalogapi.Resource      // Additional entries from Global Catalog that map to the same EnvironmentID
	SourceDoctorEnvironments                 []*doctor.EnvironmentEntry
	SourceDoctorRegionIDs                    []*doctor.RegionID
	//	DuplicateIDs []string
	mergeWorkArea struct { // work area for intermediate results from the merge
		finalCRNMask             crn.Mask // parsed version of OSSEnvironment.EnvironmentID
		catalogCRNMask           crn.Mask
		doctorEnvironmentCRNMask crn.Mask
	}
}

// HasPriorOSS returns true if the PriorOSS sub-record in this EnvironmentInfo is present, false otherwise
func (env *EnvironmentInfo) HasPriorOSS() bool {
	return env.PriorOSS.EnvironmentID != ""
}

// GetPriorOSS returns a pointer to the PriorOSS record in this EnvironmentInfo if it is valid, nil otherwise
func (env *EnvironmentInfo) GetPriorOSS() *ossrecord.OSSEnvironment {
	if env.HasPriorOSS() {
		return &env.PriorOSS
	}
	return nil
}

// HasPriorOSSValidation returns true if the PriorOSSValidation sub-record in this EnvironmentInfo is present, false otherwise
func (env *EnvironmentInfo) HasPriorOSSValidation() bool {
	return env.PriorOSSValidation != nil
}

// GetPriorOSSValidation returns a pointer to the PriorOSSValidation record in this EnvironmentInfo if it is valid, nil otherwise
func (env *EnvironmentInfo) GetPriorOSSValidation() *ossvalidation.OSSValidation {
	if env.HasPriorOSSValidation() {
		return env.PriorOSSValidation
	}
	return nil
}

// HasSourceMainCatalog returns true if the SourceMainCatalog sub-record in this EnvironmentInfo is present, false otherwise
func (env *EnvironmentInfo) HasSourceMainCatalog() bool {
	return env.SourceMainCatalog.Name != ""
}

// GetSourceMainCatalog returns a pointer to the SourceMainCatalog sub-record in this EnvironmentInfo if it is valid, nil otherwise
func (env *EnvironmentInfo) GetSourceMainCatalog() *catalogapi.Resource {
	if env.HasSourceMainCatalog() {
		return &env.SourceMainCatalog
	}
	return nil
}

// HasSourceDoctorEnvironment returns true if the (first) SourceDoctorEnvironment sub-record in this EnvironmentInfo if present, false otherwise
func (env *EnvironmentInfo) HasSourceDoctorEnvironment() bool {
	if len(env.SourceDoctorEnvironments) > 0 && env.SourceDoctorEnvironments[0].Doctor != "" {
		return true
	}
	return false
}

// GetSourceDoctorEnvironment returns a pointer to the (first) SourceDoctorEnvironment sub-record in this EnvironmentInfo if it is valid, nil otherwise
func (env *EnvironmentInfo) GetSourceDoctorEnvironment() *doctor.EnvironmentEntry {
	if env.HasSourceDoctorEnvironment() {
		return env.SourceDoctorEnvironments[0]
	}
	return nil
}

// HasSourceDoctorRegionID returns true if the (first) SourceDoctorRegionID sub-record in this EnvironmentInfo if present, false otherwise
func (env *EnvironmentInfo) HasSourceDoctorRegionID() bool {
	if len(env.SourceDoctorRegionIDs) > 0 && env.SourceDoctorRegionIDs[0].Name != "" {
		return true
	}
	return false
}

// GetSourceDoctorRegionID returns a pointer to the (first) SourceDoctorRegionID sub-record in this EnvironmentInfo if it is valid, nil otherwise
func (env *EnvironmentInfo) GetSourceDoctorRegionID() *doctor.RegionID {
	if env.HasSourceDoctorRegionID() {
		return env.SourceDoctorRegionIDs[0]
	}
	return nil
}

// AddValidationIssue creates a new ValidationIssue record for this this EnvironmentInfo record
func (env *EnvironmentInfo) AddValidationIssue(severity ossvalidation.Severity, title string, detailsFmt string, a ...interface{}) *ossvalidation.ValidationIssue {
	//	issue := ossvalidation.NewIssue(severity, title, detailsFmt, a...)
	issue := env.OSSValidation.AddIssue(severity, title, detailsFmt, a...)
	/*
		switch issue.GetSeverity() {
		case ossvalidation.CRITICAL, ossvalidation.SEVERE:
			debug.PrintError("%-66s  %s", env.String(), strings.TrimSpace(issue.String()))
		case ossvalidation.WARNING:
			debug.Warning("%-66s  %s", env.String(), strings.TrimSpace(issue.String()))
		default:
			debug.Info("%-66s  %s", env.String(), strings.TrimSpace(issue.String()))
		}
	*/
	return issue
}

// String returns a short string identifier for this EnvironmentInfo
func (env *EnvironmentInfo) String() string {
	var displayName string
	id := crn.String(env.OSSEnvironment.EnvironmentID)
	if id == "" {
		id = env.ComparableCRNMask.ToCRNString()
		displayName = "<<<comparable-crn-mask>>>"
	} else {
		if !env.mergeWorkArea.finalCRNMask.IsZero() {
			actual := env.mergeWorkArea.finalCRNMask.ToCRNString()
			if actual != id {
				panic(fmt.Sprintf("EnvironmentInfo has actualCRNMask=%s but OSSEnvironment.EnvironmentID=%s", actual, id))
			}
		}
		displayName = env.OSSEnvironment.DisplayName
		if displayName == "" {
			displayName = env.PriorOSS.DisplayName
		}
	}
	return fmt.Sprintf(`Environment(%s[%s])`, displayName, id)
}

// allEnvironmentsByEnvironmentID records all the environments, indexed by their EnvironmentID
var allEnvironmentsByEnvironmentID = make(map[crn.Mask]*EnvironmentInfo)

// makeComparableCRNMask transforms a raw CRN Mask string into a simplified version that can be used
// to compare between different sources
// Note that if the CRN Mask is to be normalized, this needs to happen before it passed as input to this function
func makeComparableCRNMask(input crn.Mask) crn.Mask {
	switch input.CType {
	case "local":
		input.CName = strings.TrimPrefix(input.CName, "l-")
		input.Location = "<trimmed>" // TODO: This hack is valid only if we never have two truly valid Dedicated/Local environments with the same cname but different locations
	case "dedicated":
		input.CName = strings.TrimPrefix(input.CName, "d-")
		input.Location = "<trimmed>" // TODO: This hack is valid only if we never have two truly valid Dedicated/Local environments with the same cname but different locations
	}
	return input // Note that we actually return a copy of the input here, since it is passed by value
}

// LookupEnvironment returns the EnvironmentInfo record associated with a given CRN mask or creates a new record if appropriate.
// If no record exists and the parameter 'createIfNeeded' is false, 'nil' is returned.
func LookupEnvironment(crnMask crn.Mask, createIfNeeded bool) (env *EnvironmentInfo, found bool) {
	comparable := makeComparableCRNMask(crnMask)
	if env, found = allEnvironmentsByEnvironmentID[comparable]; found {
		return env, true
	}
	if createIfNeeded {
		env := new(EnvironmentInfo)
		env.ComparableCRNMask = comparable
		env.OSSValidation = ossvalidation.New("", options.GlobalOptions().LogTimeStamp)
		allEnvironmentsByEnvironmentID[comparable] = env
		if env.LegacyMCCPID != "" {
			list := append(allEnvironmentsByMCCPID[env.LegacyMCCPID], env)
			allEnvironmentsByMCCPID[env.LegacyMCCPID] = list
		}
		return env, true
	}
	return nil, false
}

// ListAllEnvironments invokes the handler function on all EnvironmentInfo records whose display_name matches the specified pattern
// (or all records, if the pattern is empty)
func ListAllEnvironments(pattern *regexp.Regexp, handler func(env *EnvironmentInfo)) error {
	for _, env := range allEnvironmentsByEnvironmentID {
		name := env.OSSEnvironment.DisplayName
		if pattern == nil || pattern.FindString(name) != "" {
			handler(env)
		}
	}
	return nil
}

// allEnvironmentsByEnvironmentID records all the environments, indexed by their MCCPID from Doctor (for those environments that do have a MCCPID)
var allEnvironmentsByMCCPID = make(map[string][]*EnvironmentInfo)

// LookupEnvironmentsByMCCPID returns a list of EnvironmentInfo records that are associated with the given legacy MCCPID.
// Note that there should normally be 0 or 1 environments for each MCCPID, but we return a list just in case,
// to facilitate error reporting.
func LookupEnvironmentsByMCCPID(mccpid string) (envs []*EnvironmentInfo, found bool) {
	envs, found = allEnvironmentsByMCCPID[mccpid]
	return envs, found
}

// environmentsByReferenceCatalogID records all the environments by the CatalogID (if any) of the Main record in Global Catalog
var environmentsByReferenceCatalogID = make(map[ossrecord.CatalogID]*EnvironmentInfo)

// RecordEnvironmentByReferenceCatalogID records a mapping between the ID of a Main Catalog resource representing an environment
// and the EnvironmentInfo record used to construct the OSSEnvironment record
func RecordEnvironmentByReferenceCatalogID(id ossrecord.CatalogID, env *EnvironmentInfo) {
	if env0, found := environmentsByReferenceCatalogID[id]; found {
		debug.PrintError(`Found duplicate Catalog ID for environment-related resources: id="%s"   %s   %s`, id, env0.String(), env.String())
	} else {
		environmentsByReferenceCatalogID[id] = env
	}
}

// LookupEnvironmentByReferenceCatalogID returns the EnvironmentInfo record associated with
// the ID of a Main Catalog resource representing an environment, if there is one
func LookupEnvironmentByReferenceCatalogID(id ossrecord.CatalogID) (env *EnvironmentInfo, found bool) {
	env, found = environmentsByReferenceCatalogID[id]
	return env, found
}

// OSSEntry returns the contents of this EnvironmentInfo as a OSSEntry type
func (env *EnvironmentInfo) OSSEntry() ossrecord.OSSEntry {
	return &env.OSSEnvironmentExtended
}

// PriorOSSEntry returns pointer to the PriorOSS record in this EnvironmentInfo as a OSSEntry type, if it is valid, nil otherwise
func (env *EnvironmentInfo) PriorOSSEntry() ossrecord.OSSEntry {
	return env.GetPriorOSS()
}

// IsDeletable returns true if there is no reason to keep this OSS record in the Catalog (i.e it has no source, and no non-zero OSSControl or RMC data)
func (env *EnvironmentInfo) IsDeletable() bool {
	if env.OSSEnvironment.OSSTags.Contains(osstags.OSSOnly) {
		return false
	}
	if env.OSSEnvironment.OSSTags.Contains(osstags.OSSTest) {
		return false
	}
	if env.OSSEnvironment.OSSOnboardingPhase != "" {
		return false
	}
	if env.HasSourceMainCatalog() {
		return false
	}
	if env.HasSourceDoctorEnvironment() {
		return false
	}
	if env.HasSourceDoctorRegionID() {
		return false
	}
	if !ossrunactions.Doctor.IsEnabled() && env.HasPriorOSS() && env.GetPriorOSS().OSSOnboardingPhase == "" {
		if env.OSSValidation != nil {
			env.OSSValidation.AddIssueIgnoreDup(ossvalidation.INFO, `Cannot confirm if entry is deletable when Doctor is disabled`, ``).TagPriorOSS()
		} else {
			debug.Info(`Cannot confirm if entry is deletable when Doctor is disabled: %s`, env.String())
		}
		return false
	}
	return true
}

// IsUpdatable returns true if this OSS record can be updated in the Catalog (i.e. it is a real standalone OSS record, not a placeholder for a native Catalog entry)
func (env *EnvironmentInfo) IsUpdatable() bool {
	return env.OSSEnvironment.IsUpdatable()
}

// IsValid returns true if this EnvironmentInfo record represents a valid entry that
// can be updated or merged into the Catalog
func (env *EnvironmentInfo) IsValid() bool {
	return env.OSSEnvironment.EnvironmentID != ""
}

// Header returns a short header representing this EnvironmentInfo record, suitable for printing in logs and reports
func (env *EnvironmentInfo) Header() string {
	result := strings.Builder{}
	result.WriteString(env.OSSEnvironment.Header())
	if env.OSSValidation != nil {
		result.WriteString(env.OSSValidation.Header())
	}
	return result.String()
}

// Details returns a long text representing this EnvironmentInfo record (including a Header), suitable for printing in logs and reports
func (env *EnvironmentInfo) Details() string {
	result := strings.Builder{}
	result.WriteString(env.OSSEnvironment.Header())
	if env.OSSValidation != nil {
		result.WriteString(env.OSSValidation.Details())
	}
	return result.String()
}

// Diffs returns a list of differences between the new and priorOSS versions of the OSSEnvironment record in this EnvironmentInfo
func (env *EnvironmentInfo) Diffs() *compare.Output {
	out := compare.Output{}
	compare.DeepCompare("before", &env.PriorOSS, "after", &env.OSSEnvironment, &out)
	if env.OSSValidation.Checksum() != env.PriorOSSValidationChecksum {
		compare.DeepCompare("before.OSSValidation", env.PriorOSSValidation, "after.OSSValidation", env.OSSValidation, &out)
		//out.AddOtherDiff("OSSValidation updated")
	}
	return &out
}

// GlobalSortKey returns a global key for sorting across all ossmerge.Model entries of all types
func (env *EnvironmentInfo) GlobalSortKey() string {
	return "3." + env.String()
}

// GetAccessor returns the Accessor function for a given sub-record within the EnvironmentInfo
func (env *EnvironmentInfo) GetAccessor(t reflect.Type) ossmergemodel.AccessorFunc {
	panic("GetAccessor() not defined for type ossmerge.EnvironmentInfo")
}

// GetOSSValidation returns the OSSValidation record in this EnvironmentInfo, or nil if there is none
func (env *EnvironmentInfo) GetOSSValidation() *ossvalidation.OSSValidation {
	return env.OSSValidation
}

// SkipMerge returns true if we should skip the OSS merge for this record and simply copy it from
// the prior OSS record.
// If not empty, the "reason" string provides details on why this record should or should not be merged
func (env *EnvironmentInfo) SkipMerge() (skip bool, reason string) {
	return ossmergemodel.SkipMerge(env)
}

var _ ossmergemodel.Model = &EnvironmentInfo{} // verify

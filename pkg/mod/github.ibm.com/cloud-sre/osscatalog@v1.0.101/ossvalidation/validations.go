package ossvalidation

// Functions and declarations tracking validation issues while analyzing or constructing OSS records from multiple sources

import (
	"crypto/sha1" // #nosec G505
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// Source represents the source of some information used during merging and validation (e.g. Catalog, ServiceNow, ScorecardV1, ...)
type Source string

// Possible values for the Source attributes in ValidationInfo records
const (
	PRIOROSS               Source = "Catalog-OSS"
	CATALOG                Source = "Catalog-Main"
	CATALOGIGNORED         Source = "Catalog-Main-IGNORED"
	SERVICENOW             Source = "ServiceNow"
	SERVICENOWRETIRED      Source = "ServiceNow-RETIRED"
	SCORECARDV1            Source = "ScorecardV1"
	SCORECARDV1DISABLED    Source = "ScorecardV1-DISABLED" // Special value to track old entries in ScorecardV1 when the ScorecardV1 RunAction is disabled
	DOCTORENV              Source = "Doctor-Env"
	DOCTORENVIGNORED       Source = "Doctor-Env-IGNORED"
	DOCTORENVDISABLED      Source = "Doctor-Env-DISABLED" // Special value to track old entries in Doctor when the Doctor RunAction is disabled
	DOCTORREGIONID         Source = "Doctor-RegionID"
	DOCTORREGIONIDIGNORED  Source = "Doctor-RegionID-IGNORED"
	DOCTORREGIONIDDISABLED Source = "Doctor-RegionID-DISABLED" // Special value to track old entries in Doctor when the Doctor RunAction is disabled
	IAM                    Source = "IAM"
	IAMDISABLED            Source = "IAM-DISABLED"
	RMC                    Source = "RMC"
	COMPUTED               Source = "(computed)"
)

// ComparableString forces comparisons of Sources using the special string slice comparator
func (s Source) ComparableString() string {
	return string(s)
}

// OSSValidation encapsulates all the ValidationIssues and other validation information gathered
// while merging and analyzing a ServiceInfo record
type OSSValidation struct {
	CanonicalName        string              `json:"canonical_name"`
	CanonicalNameSources []Source            `json:"canonical_name_sources"`
	OtherNamesSources    map[string][]Source `json:"other_names_sources"`
	CatalogVisibility    struct {
		EffectiveRestrictions string `json:"effective_restrictions"`
		LocalRestrictions     string `json:"local_restrictions"`
		Active                bool   `json:"active"`
		Hidden                bool   `json:"hidden"`
		Disabled              bool   `json:"disabled"`
	} `json:"catalog_visibility,omitempty"`
	StatusCategoryCount int                           `json:"status_category_count,omitempty"`
	LastRunTimestamp    string                        `json:"last_update_run"` // timestamp of the last osscatimporter run that actually updated the OSSEntry or any of the OSSValidation info
	LastRunActions      map[string]string             `json:"last_run_actions"`
	Issues              []*ValidationIssue            `json:"issues"`
	issuesMap           map[string][]*ValidationIssue // Not written with the JSON
}

// String returns a short string representation of this OSSValidation
func (ossv *OSSValidation) String() string {
	return fmt.Sprintf(`OSSValidation("%s")`, ossv.CanonicalName)
}

// AllSources returns a list of all the Sources for this record, including both the CanonicalNameSources and the OtherNamesSources
func (ossv *OSSValidation) AllSources() []Source {
	allSources := make(map[Source]struct{})
	result := make([]Source, 0, len(ossv.CanonicalNameSources)+len(ossv.OtherNamesSources)) // We expect one Source per other name, but the slice will adjust if needed
	for _, s := range ossv.CanonicalNameSources {
		if _, found := allSources[s]; !found {
			result = append(result, s)
			allSources[s] = struct{}{}
		}
	}
	for _, srcs := range ossv.OtherNamesSources {
		for _, s := range srcs {
			if _, found := allSources[s]; !found {
				result = append(result, s)
				allSources[s] = struct{}{}
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

// AllNames returns a list of all the names for this record, starting with CanonicalNameS followed by the OtherNames
func (ossv *OSSValidation) AllNames() []string {
	if len(ossv.OtherNamesSources) > 0 {
		otherNames := make([]string, 0, len(ossv.OtherNamesSources))
		for n := range ossv.OtherNamesSources {
			otherNames = append(otherNames, n)
		}
		sort.Strings(otherNames)
		if ossv.CanonicalName != "" {
			return append([]string{ossv.CanonicalName}, otherNames...)
		}
		return otherNames
	} else if ossv.CanonicalName != "" {
		return []string{ossv.CanonicalName}
	}
	return nil
}

// NameSources returns all the sources associated with a given name
func (ossv *OSSValidation) NameSources(name string) []Source {
	var result []Source
	if name == ossv.CanonicalName {
		result = ossv.CanonicalNameSources
	} else {
		result = ossv.OtherNamesSources[name]
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

// TrueSources returns a list of all the "true" Sources for this record, excluding ignored entries from Catalog, ScorecardV1, ServiceNow,
// or other types of entries that do not really confirm the existence of a real service/component in actual use {
func (ossv *OSSValidation) TrueSources() []Source {
	trueSources := make(map[Source]struct{})
	result := make([]Source, 0, len(ossv.CanonicalNameSources)+len(ossv.OtherNamesSources)) // We expect one Source per other name, but the slice will adjust if needed
	for _, s := range ossv.CanonicalNameSources {
		switch s {
		case PRIOROSS, COMPUTED, CATALOGIGNORED, IAM:
			// do nothing
		default:
			if _, found := trueSources[s]; !found {
				result = append(result, s)
				trueSources[s] = struct{}{}
			}
		}
	}
	for _, srcs := range ossv.OtherNamesSources {
		for _, s := range srcs {
			switch s {
			case PRIOROSS, COMPUTED, CATALOGIGNORED, IAM:
				// do nothing
			default:
				if _, found := trueSources[s]; !found {
					result = append(result, s)
					trueSources[s] = struct{}{}
				}
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

// NumTrueSources returns the number of "true" Sources for this record, excluding ignored entries from Catalog, ScorecardV1, ServiceNow,
// or other types of entries that do not really confirm the existence of a real service/component in actual use
func (ossv *OSSValidation) NumTrueSources() int {
	return len(ossv.TrueSources())
}

// New creates a new, empty OSSValidation record
func New(name string, lastRunTimestamp string) *OSSValidation {
	result := new(OSSValidation)
	if name != "" {
		result.SetSourceNameCanonical(name)
	}
	result.LastRunTimestamp = lastRunTimestamp
	return result
}

// AddIssuePreallocated adds a pre-allocation ValidationIssue (created with NewIssue()) to the list in this OSS ValidationInfo object
func (ossv *OSSValidation) AddIssuePreallocated(v *ValidationIssue) {
	if ossv.issuesMap == nil {
		ossv.issuesMap = make(map[string][]*ValidationIssue)
	}
	if vlist, found := ossv.issuesMap[v.Title]; !found {
		ossv.issuesMap[v.Title] = []*ValidationIssue{v}
	} else {
		for _, v0 := range vlist {
			if v0.Details == v.Details {
				debug.PrintError("Duplicate ValidationIssue (same Details) for %s: %s  /  %s", ossv.CanonicalName, v0.String(), v.String())
				//debug.Debug(debug.Merge, "Duplicate ValidationIssue (same Details) for %s: %s  /  %s", ossv.CanonicalName, v0.String(), v.String())
				//panic(fmt.Sprintf("Duplicate ValidationIssue (same Details) for %s: %s  /  %s", ossv.CanonicalName, v0.String(), v.String()))
			}
			if v0.Severity != v.Severity {
				debug.PrintError("Duplicate ValidationIssue (different Severity) for %s: %s  /  %s", ossv.CanonicalName, v0.String(), v.String())
				//debug.Debug(debug.Merge, "Duplicate ValidationIssue (different Severity) for %s: %s  /  %s", ossv.CanonicalName, v0.String(), v.String())
				//panic(fmt.Sprintf("Duplicate ValidationIssue (different Severity) for %s: %s  /  %s", ossv.CanonicalName, v0.String(), v.String()))
			}
		}
		ossv.issuesMap[v.Title] = append(vlist, v)
	}
	ossv.Issues = append(ossv.Issues, v)
}

// AddIssue creates a new ValidationIssue record and adds it to the list in this OSS ValidationInfo object
func (ossv *OSSValidation) AddIssue(severity Severity, title string, detailsFmt string, a ...interface{}) *ValidationIssue {
	details := fmt.Sprintf(detailsFmt, a...)
	v := ValidationIssue{Severity: severity, Title: title, Details: details}
	ossv.AddIssuePreallocated(&v)
	return &v
}

// AddNamedIssue creates a new instance of a ValidationIssue record based on a named pattern and adds it to the list in this OSS ValidationInfo object
func (ossv *OSSValidation) AddNamedIssue(pattern *NamedValidationIssue, detailsFmt string, a ...interface{}) *ValidationIssue {
	details := fmt.Sprintf(detailsFmt, a...)
	v := ValidationIssue{Severity: pattern.Severity, Title: pattern.Title, Details: details}
	v.AddTag(pattern.Tags...)
	ossv.AddIssuePreallocated(&v)
	return &v
}

// AddIssueIgnoreDup creates a new ValidationIssue record and adds it to the list in this OSS ValidationInfo object, unless there is already a ValidationIssue with the same title
func (ossv *OSSValidation) AddIssueIgnoreDup(severity Severity, title string, detailsFmt string, a ...interface{}) *ValidationIssue {
	if vlist, found := ossv.issuesMap[title]; found {
		return vlist[0]
	}
	details := fmt.Sprintf(detailsFmt, a...)
	v := ValidationIssue{Severity: severity, Title: title, Details: details}
	ossv.AddIssuePreallocated(&v)
	return &v
}

// AddSource registers one Source for a name variant for this OSS ValidationInfo object
func (ossv *OSSValidation) AddSource(name string, source Source) {
	// TODO: Detect duplicate entries from the same source
	if ossv.OtherNamesSources == nil {
		ossv.OtherNamesSources = make(map[string][]Source)
	}
	if name == string(ossv.CanonicalName) {
		ossv.CanonicalNameSources = append(ossv.CanonicalNameSources, source)
	} else {
		list := ossv.OtherNamesSources[name]
		ossv.OtherNamesSources[name] = append(list, source)
	}
}

// SetSourceNameCanonical designates one particular Source name as being the canonical name for this OSS entry
func (ossv *OSSValidation) SetSourceNameCanonical(name string) {
	if ossv.CanonicalName != "" {
		panic(fmt.Sprintf("ossvalidation.SetSourceNameCanonical() called more than once   - name1=\"%s\"    name2=\"%s\"", ossv.CanonicalName, name))
	}
	ossv.CanonicalName = name
	list, found := ossv.OtherNamesSources[string(name)]
	if found {
		delete(ossv.OtherNamesSources, string(name))
		ossv.CanonicalNameSources = append(ossv.CanonicalNameSources, list...)
	} else {
		ossv.CanonicalNameSources = append(ossv.CanonicalNameSources, COMPUTED)
	}
}

// SourceNames returns a list of names associated with the given Source, or nil if this Source is not part of this OSSValidation
// Note that this method returns a list (slice) because there may be multiple Source records of the same type, with different names
// (after name merging)
func (ossv *OSSValidation) SourceNames(source Source) []string {
	var result = collections.NewStringSet()
	for _, cns0 := range ossv.CanonicalNameSources {
		if cns0 == source {
			result.Add(ossv.CanonicalName)
		}
	}
	for name, ons := range ossv.OtherNamesSources {
		for _, ons0 := range ons {
			if ons0 == source {
				result.Add(name)
			}
		}
	}
	return result.Slice()
}

// FindNamedIssue checks if this OSSValidation object contains a particular NamedValidationIssue, and return the actual ValidationIssue instance if found
// XXX If the object contains multiple issues with the same type, only the first one is returned
func (ossv *OSSValidation) FindNamedIssue(pattern *NamedValidationIssue) (issue *ValidationIssue, found bool) {
	// TODO: Use faster lookup that linear search
	if vlist, found := ossv.issuesMap[pattern.Title]; found {
		return vlist[0], true
	}
	return nil, false
}

// CountIssues returns the number of ValidationIssues of various severities contained in this OSSValidation object,
// filtering by a given set of validation Tags (or all if filter = nil),
// as a map keyed by Severity, plus one special "TOTAL" key
func (ossv *OSSValidation) CountIssues(filter []Tag) map[Severity]int {
	// TODO: Should cache severity counts
	var counts = make(map[Severity]int)
	var totalIssues = 0
	if filter != nil {
		var filterMap = make(map[Tag]bool)
		for _, f := range filter {
			filterMap[f] = true
		}
		for i := 0; i < len(ossv.Issues); i++ {
			v := ossv.Issues[i]
			for _, vt := range v.Tags {
				if filterMap[vt] {
					counts[v.Severity]++
					switch v.Severity {
					case IGNORE, INFO:
					default:
						totalIssues++
					}
					break
				}
			}
		}
	} else {
		for i := 0; i < len(ossv.Issues); i++ {
			v := ossv.Issues[i]
			counts[v.Severity]++
			switch v.Severity {
			case IGNORE, INFO:
			default:
				totalIssues++
			}
		}
	}
	counts["TOTAL"] = totalIssues
	return counts
}

// GetIssues returns all the ValidationIssues contained in this OSSValidation object,
// filtering by a given set of validation Tags (or all if filter = nil),
func (ossv *OSSValidation) GetIssues(filter []Tag) []*ValidationIssue {
	var result = make([]*ValidationIssue, 0, len(ossv.Issues))
	if filter != nil {
		var filterMap = make(map[Tag]bool)
		for _, f := range filter {
			filterMap[f] = true
		}
		for i := 0; i < len(ossv.Issues); i++ {
			v := ossv.Issues[i]
			for _, vt := range v.Tags {
				if filterMap[vt] {
					result = append(result, v)
				}
			}
		}
	} else {
		for i := 0; i < len(ossv.Issues); i++ {
			v := ossv.Issues[i]
			result = append(result, v)
		}
	}
	return result
}

// SummaryOverallStatus computes the summary OSS overall validation status (Red, Yellow, Green) from a set of validation issues
func (ossv *OSSValidation) SummaryOverallStatus(tags osstags.TagSet) osstags.Tag {
	counts := ossv.CountIssues(nil)
	switch {
	case (counts[CRITICAL] + counts[SEVERE]) > 0:
		if tags.Contains(osstags.NotReady) {
			return osstags.StatusYellow
		}
		return osstags.StatusRed
	case counts[WARNING] > 0:
		return osstags.StatusYellow
	default:
		// Note that we do not consider MINOR issues as bad enough to make use Yellow
		return osstags.StatusGreen
	}
}

// SummaryCRNStatus computes the summary OSS CRN validation status (Red, Yellow, Green) from a set of validation issues
func (ossv *OSSValidation) SummaryCRNStatus(tags osstags.TagSet) osstags.Tag {
	counts := ossv.CountIssues([]Tag{TagCRN})
	switch {
	case (counts[CRITICAL] + counts[SEVERE]) > 0:
		if tags.Contains(osstags.NotReady) {
			return osstags.StatusCRNYellow
		}
		return osstags.StatusCRNRed
	case counts[WARNING] > 0:
		return osstags.StatusCRNYellow
	default:
		// Note that we do not consider MINOR issues as bad enough to make use Yellow
		return osstags.StatusCRNGreen
	}
}

// Sort sorts all the ValidationIssues in-place in a ValidationInfo record
func (ossv *OSSValidation) Sort() {
	var buffers = make(map[Severity][]*ValidationIssue)
	for i := 0; i < len(ossv.Issues); i++ {
		v := ossv.Issues[i]
		buffers[v.Severity] = append(buffers[v.Severity], v)
	}
	newList := make([]*ValidationIssue, 0, len(ossv.Issues))
	for _, s := range AllSeverityList() {
		cur := buffers[s]
		sort.SliceStable(cur, func(i, j int) bool {
			if cur[i].Title == cur[j].Title {
				return cur[i].Details < cur[j].Details
			}
			return cur[i].Title < cur[j].Title
		})
		newList = append(newList, cur...)
	}
	ossv.Issues = newList

	// Sort the sources
	sort.Slice(ossv.CanonicalNameSources, func(i, j int) bool {
		return ossv.CanonicalNameSources[i] < ossv.CanonicalNameSources[j]
	})
	// TODO: Sort the other source names themselves, not just the list of sources under each name
	for _, src := range ossv.OtherNamesSources {
		sort.Slice(src, func(i, j int) bool {
			return src[i] < src[j]
		})
	}

}

// Checksum computes a checksum for this OSSValidation record and returns it as a string
// Used for quick comparison between OSSValidation records
func (ossv *OSSValidation) Checksum() string {
	// Exclude the LastRunTimestamp from the checksum - we do not want to update records if that is the only difference
	savedLastRunTimestamp := ossv.LastRunTimestamp
	ossv.LastRunTimestamp = ""
	data, _ := json.Marshal(ossv)
	sum := sha1.Sum(data) // #nosec G401
	newCksum := hex.EncodeToString(sum[:len(sum)])
	ossv.LastRunTimestamp = savedLastRunTimestamp
	return newCksum
}

// RecordCatalogVisibility Records the Catalog entry visibility information in this OSSValidation record.
// Note that we cannot simply pass a catalogapi.Resource as parameter, because this would cause a circular package dependency
// TODO: use actual Catalog API definitions for the CatalogVisibility info in OSSValidation (fix circular package dependency)
func (ossv *OSSValidation) RecordCatalogVisibility(effectiveRestrictions, localRestrictions string, active, hidden, disabled bool) {
	ossv.CatalogVisibility.EffectiveRestrictions = effectiveRestrictions
	ossv.CatalogVisibility.LocalRestrictions = localRestrictions
	ossv.CatalogVisibility.Active = active
	ossv.CatalogVisibility.Hidden = hidden
	ossv.CatalogVisibility.Disabled = disabled
}

// ResetForRMC resets the validation info (status and issues) that might become obsolete when an entry is being edited in RMC
func (ossv *OSSValidation) ResetForRMC() {
	ossv.AddIssueIgnoreDup(InvalidatedByRMC.Severity, InvalidatedByRMC.Title, "").AddTag(InvalidatedByRMC.Tags...)
	/*
		for n := range ossv.LastRunActions {
			ossv.LastRunActions[n] = "<unknown>"
		}
	*/
	ossv.Sort()
}

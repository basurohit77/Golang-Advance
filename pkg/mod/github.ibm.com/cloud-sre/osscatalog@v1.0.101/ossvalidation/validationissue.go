package ossvalidation

import "fmt"

// Severity of a ValidationIssue
type Severity string

// Possible values for Severity of a ValidationIssue
const (
	CRITICAL Severity = "(critical)"
	SEVERE   Severity = "(severe)"
	WARNING  Severity = "(warning)"
	MINOR    Severity = "(minor)"
	DEFERRED Severity = "(deferred)"
	INFO     Severity = "(info)"
	IGNORE   Severity = "(ignore)"
)

// AllSeverityList returns a list of all Severity values, in the order in which we shoud report ValidationIssues of various severities
func AllSeverityList() []Severity {
	return []Severity{CRITICAL, SEVERE, WARNING, MINOR, DEFERRED, INFO, IGNORE}
}

var severityIndex = map[Severity]int{
	CRITICAL: 1,
	SEVERE:   2,
	WARNING:  3,
	MINOR:    4,
	DEFERRED: 5,
	INFO:     6,
	IGNORE:   7,
}

// GetSeverityIndex returns an integer index associated with a Severity
func GetSeverityIndex(s Severity) int {
	if ix, ok := severityIndex[s]; ok {
		return ix
	}
	panic(fmt.Sprintf("GetSeverityIndex(%s): unknown Severity", s))
}

// ActionableSeverityList returns a list of "actionable" Severity values, i.e excluding INFO and IGNORE
func ActionableSeverityList() []Severity {
	return []Severity{CRITICAL, SEVERE, WARNING, MINOR}
}

// ValidationIssue represents one issue/conflict encountered when reconciling data from multiple sources for one OSS record
type ValidationIssue struct {
	Title    string
	Details  string
	Severity Severity
	Tags     []Tag
}

// NamedValidationIssue represents a pattern for one particular type of ValidationIssue, of which instances may be found
// in many OSSValidation objects
type NamedValidationIssue ValidationIssue

// NewIssue allocates a new ValidationIssue without attaching it to an OSS ValidationInfo object
func NewIssue(severity Severity, title string, detailsFmt string, a ...interface{}) *ValidationIssue {
	details := fmt.Sprintf(detailsFmt, a...)
	v := ValidationIssue{Severity: severity, Title: title, Details: details}
	return &v
}

// newNamedIssue create a new ValidationIssue (without adding it to any OSSValidation record)
// Used to created named issue patterns
func newNamedIssue(severity Severity, title string, tags ...Tag) *NamedValidationIssue {
	v := NamedValidationIssue{Severity: severity, Title: title, Tags: tags}
	return &v
}

// IsInstance returns true if this specified Validation issue is an instance of the specified NamedValidationIssue
func (vi *ValidationIssue) IsInstance(ni *NamedValidationIssue) bool {
	return ni.Title == vi.Title
}

// String produces a string representation for one ValidationIssue entry
func (vi *ValidationIssue) String() string {
	var result string
	if vi.Details != "" {
		result = fmt.Sprintf("%-10s %-15s %s: %s\n", vi.Severity, fmt.Sprintf("%v", vi.Tags), vi.Title, vi.Details)
	} else {
		result = fmt.Sprintf("%-10s %-15s %s\n", vi.Severity, fmt.Sprintf("%v", vi.Tags), vi.Title)
	}
	return result
}

// GetSeverity returns the severity of this ValidationIssue
func (vi *ValidationIssue) GetSeverity() Severity {
	return vi.Severity
}

// GetText returns the full text of this ValidationIssue (but without severity or tags)
func (vi *ValidationIssue) GetText() string {
	var result string
	if vi.Details != "" {
		result = fmt.Sprintf("%s: %s", vi.Title, vi.Details)
	} else {
		result = fmt.Sprintf("%s", vi.Title)
	}
	return result
}

// IsEqual returns true if two ValidationIssues are identical
func (vi *ValidationIssue) IsEqual(vi2 *ValidationIssue) bool {
	// XXX We do not compare the Tags, for efficiency
	if vi.Title == vi2.Title && vi.Details == vi2.Details && vi.Severity == vi2.Severity {
		return true
	}
	return false
}

// ComparableString forces comparisons of ValidationIssue records to be on a single line with all attributes together
func (vi *ValidationIssue) ComparableString() string {
	var result string
	if vi.Details != "" {
		result = fmt.Sprintf(`%d%-10s %s: %s %-15s`, GetSeverityIndex(vi.Severity), vi.Severity, vi.Title, vi.Details, fmt.Sprintf("%v", vi.Tags))
	} else {
		result = fmt.Sprintf(`%d%-10s %s %-15s`, GetSeverityIndex(vi.Severity), vi.Severity, vi.Title, fmt.Sprintf("%v", vi.Tags))
	}
	return result
}

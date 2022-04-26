package compare

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// DiffKind represents the type of difference for one item in the output (one field)
type DiffKind int

// Possible values for DiffType
const (
	DiffLOnly DiffKind = iota
	DiffROnly
	DiffValue
	DiffType
	DiffEqual
	DiffOther
)

// OneDiff represents one difference item in the output (one field)
type OneDiff struct {
	Typ   DiffKind
	LName string
	LVal  string
	RName string
	RVal  string
}

// Output contains all the output (differences) from a DeepCompare operation
type Output struct {
	allDiffs []*OneDiff
	lOnly    []string
	rOnly    []string
	diff     []string
	equal    []string
	other    []string

	// IncludeEqual is a global flag that controls whether this Output object will emit output related to
	// entries that are _equal_, or only entries that have differences.
	// Default is false.
	IncludeEqual bool
}

func (out *Output) addValueDiff(lName string, lVal reflect.Value, rName string, rVal reflect.Value) {
	d := OneDiff{
		Typ:   DiffValue,
		LName: lName,
		LVal:  fmt.Sprintf("%#v", lVal),
		RName: rName,
		RVal:  fmt.Sprintf("%#v", rVal),
	}
	out.allDiffs = append(out.allDiffs, &d)
	out.diff = append(out.diff, fmt.Sprintf(`DIFF VALUE:      %s=%s    %s=%s`, d.LName, d.LVal, d.RName, d.RVal))
	//	checkBadString(out.diff[len(out.diff)-1], "in Output.addValueDiff")
}

func (out *Output) addKindDiff(lName string, lVal reflect.Value, rName string, rVal reflect.Value) {
	d := OneDiff{
		Typ:   DiffType,
		LName: lName,
		LVal:  fmt.Sprintf("Type(%v)=%T", lVal.Type().Kind().String(), lVal.Interface()),
		RName: rName,
		RVal:  fmt.Sprintf("Type(%v)=%T", rVal.Type().Kind().String(), rVal.Interface()),
	}
	out.allDiffs = append(out.allDiffs, &d)
	out.diff = append(out.diff, fmt.Sprintf(`DIFF TYPE:       %s.%s    %s.%s`, d.LName, d.LVal, d.RName, d.RVal))
}

// AddOtherDiff appends a generic "diff" message in this Output record
func (out *Output) AddOtherDiff(msg string) {
	d := OneDiff{
		Typ:   DiffOther,
		LName: msg,
	}
	out.allDiffs = append(out.allDiffs, &d)
	out.other = append(out.other, fmt.Sprintf(`OTHER:           %s`, d.LName))
}

func (out *Output) addLOnly(lName string, lVal reflect.Value) {
	d := OneDiff{
		Typ:   DiffLOnly,
		LName: lName,
		LVal:  fmt.Sprintf("%#v", lVal),
		RName: lName,
	}
	out.allDiffs = append(out.allDiffs, &d)
	out.lOnly = append(out.lOnly, fmt.Sprintf(`DIFF LEFT ONLY:  %s=%s`, d.LName, d.LVal))
}
func (out *Output) addROnly(rName string, rVal reflect.Value) {
	d := OneDiff{
		Typ:   DiffROnly,
		LName: rName,
		RName: rName,
		RVal:  fmt.Sprintf("%#v", rVal),
	}
	out.allDiffs = append(out.allDiffs, &d)
	out.rOnly = append(out.rOnly, fmt.Sprintf(`DIFF RIGHT ONLY:         %s=%s`, d.RName, d.RVal))
}

func (out *Output) addValueEqual(lName string, lVal reflect.Value, rName string, rVal reflect.Value) {
	d := OneDiff{
		Typ:   DiffEqual,
		LName: lName,
		LVal:  fmt.Sprintf("%#v", lVal),
		RName: rName,
		RVal:  fmt.Sprintf("%#v", rVal),
	}
	out.allDiffs = append(out.allDiffs, &d)
	out.equal = append(out.equal, fmt.Sprintf(`EQUAL:           %s=%s`, d.LName, d.LVal))
}

// ToStrings produces a slice of strings containing all the entries in the given Output object, in a predictable order
func (out *Output) ToStrings() []string {
	var result []string
	// Note that we do not sort the entries here. We expect that they were appended in the order that they should be displayed
	if out.IncludeEqual {
		result = append(result, out.equal...)
	}
	result = append(result, out.diff...)
	result = append(result, out.lOnly...)
	result = append(result, out.rOnly...)
	result = append(result, out.other...)
	return result
}

// NumDiffs returns the total number of differences recorded in this Output record
// (including left-only items, right-only items and items that are actually different on each side)
func (out *Output) NumDiffs() int {
	return len(out.lOnly) + len(out.rOnly) + len(out.diff) + len(out.other)
}

var sectionRegex = regexp.MustCompile(`^[^.]+\.([^.=]+)`)

// nonCoreSections lists the keys for diffs that are to be counted outside the "(core)" category in Summary()
var nonCoreSections = map[string]struct{}{
	`SchemaVersion`:  struct{}{},
	`ProductInfo`:    struct{}{},
	`MonitoringInfo`: struct{}{},
	`DependencyInfo`: struct{}{},
	`CatalogInfo`:    struct{}{},
	`OSSValidation`:  struct{}{},
}

// Summary return a short string summary of the counts of differences recorded in this Output record
func (out *Output) Summary() string {
	var buffer strings.Builder
	if len(out.lOnly) > 0 {
		if buffer.Len() > 0 {
			buffer.WriteString("  ")
		}
		buffer.WriteString(fmt.Sprintf("LOnly:%d", len(out.lOnly)))
	}
	if len(out.rOnly) > 0 {
		if buffer.Len() > 0 {
			buffer.WriteString("  ")
		}
		buffer.WriteString(fmt.Sprintf("ROnly:%d", len(out.rOnly)))
	}
	if len(out.diff) > 0 {
		if buffer.Len() > 0 {
			buffer.WriteString("  ")
		}
		buffer.WriteString(fmt.Sprintf("Diff:%d", len(out.diff)))
	}
	if len(out.other) > 0 {
		if buffer.Len() > 0 {
			buffer.WriteString("  ")
		}
		buffer.WriteString(fmt.Sprintf("Other:%d", len(out.other)))
	}
	coreCount := 0
	sectionsMap := make(map[string]int)
	for _, d := range out.allDiffs {
		switch d.Typ {
		case DiffLOnly, DiffROnly, DiffValue, DiffType /*, DiffOther */ :
			m := sectionRegex.FindStringSubmatch(d.RName)
			if m == nil {
				m = sectionRegex.FindStringSubmatch(d.LName)
			}
			if m != nil && m[1] != "" {
				section := strings.TrimSpace(m[1])
				if _, found := nonCoreSections[section]; found {
					sectionsMap[section]++
				} else {
					coreCount++
				}
			} else {
				sectionsMap[`???`]++
			}
		default:
			// nothing to do
		}
	}
	if coreCount > 0 {
		if buffer.Len() > 0 {
			buffer.WriteString("  ")
		}
		buffer.WriteString(fmt.Sprintf("(core):%d", coreCount))
	}
	sectionsNames := make([]string, 0, len(sectionsMap))
	for s := range sectionsMap {
		sectionsNames = append(sectionsNames, s)
	}
	sort.Strings(sectionsNames)
	for _, s := range sectionsNames {
		count := sectionsMap[s]
		if buffer.Len() > 0 {
			buffer.WriteString("  ")
		}
		buffer.WriteString(fmt.Sprintf("%s:%d", s, count))
	}
	if buffer.Len() == 0 {
		buffer.WriteString("(none)")
	}
	return buffer.String()
}

// StringWithPrefix converts the given Output record to a single multi-line printable string, including specified prefix
func (out *Output) StringWithPrefix(prefix string) string {
	var buffer strings.Builder
	strings := out.ToStrings()
	for _, s := range strings {
		buffer.WriteString(prefix + s + "\n")
	}
	return buffer.String()
}

// String converts the given Output record to a single multi-line printable string
func (out *Output) String() string {
	return out.StringWithPrefix("")
}

// GetDiffs returns a slice containing all the difference items in the output, in the order that they were found
func (out *Output) GetDiffs() []*OneDiff {
	result := make([]*OneDiff, 0, len(out.allDiffs))
	for _, d := range out.allDiffs {
		if !out.IncludeEqual && d.Typ == DiffEqual {
			continue
		}
		result = append(result, d)
	}
	return result
}

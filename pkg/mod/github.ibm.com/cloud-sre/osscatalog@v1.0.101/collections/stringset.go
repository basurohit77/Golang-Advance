package collections

import (
	"fmt"
	"sort"
	"strings"
)

var _stringSetMapThreshold = 10 // Use a map for fast lookup, if the size reaches this threshold

// StringSet represents a set of string values, with no duplicates, sorted alphabetically
type StringSet interface {
	Add(vals ...string) (numDups int)
	Contains(val string) bool
	Compare(right StringSet) (leftOnly, rightOnly StringSet)
	Len() int
	Slice() []string
	String() string
}

type stringSetData struct {
	// TODO: use a map for better performance when the size of the data becomes larger
	sliceData []string
	mapData   map[string]struct{}
}

// NewStringSet allocates a new StringSet initialied with the supplied values
func NewStringSet(vals ...string) StringSet {
	result := new(stringSetData)
	for _, val := range vals {
		result.Add(val)
	}
	return result
}

func (ssd *stringSetData) Add(vals ...string) (numDups int) {
	for _, val := range vals {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		if ssd.mapData != nil {
			if _, found := ssd.mapData[val]; found {
				numDups++
			} else {
				ssd.sliceData = append(ssd.sliceData, val)
				ssd.mapData[val] = struct{}{}
			}
		} else {
			var isDup bool
			for _, elem := range ssd.sliceData {
				if elem == val {
					numDups++
					isDup = true
					break
				}
			}
			if !isDup {
				ssd.sliceData = append(ssd.sliceData, val)
				if len(ssd.sliceData) >= _stringSetMapThreshold {
					ssd.mapData = make(map[string]struct{})
					for _, e := range ssd.sliceData {
						ssd.mapData[e] = struct{}{}
					}
				}
			}
		}
	}
	return numDups
}

func (ssd *stringSetData) Contains(val string) bool {
	val = strings.TrimSpace(val)
	if ssd.mapData != nil {
		if _, found := ssd.mapData[val]; found {
			return true
		}
	} else {
		for _, elem := range ssd.sliceData {
			if elem == val {
				return true
			}
		}
	}
	return false
}

func (ssd *stringSetData) Len() int {
	return len(ssd.sliceData)
}

func (ssd *stringSetData) Slice() []string {
	if len(ssd.sliceData) == 0 {
		return nil
	}
	sort.Strings(ssd.sliceData)
	return ssd.sliceData
}

func (ssd *stringSetData) String() string {
	var result strings.Builder
	for i, v := range ssd.Slice() {
		if i > 0 {
			result.WriteString(`,`)
		}
		result.WriteString(fmt.Sprintf(`"%s"`, v))
	}
	return result.String()
}

// Compare compares two StringSets and reports on all differences
// Note: this has the side effect of sorting both StringSets
func (ssd *stringSetData) Compare(right StringSet) (leftOnly, rightOnly StringSet) {
	// Re-sort just to be safe
	leftSlice := ssd.Slice()
	rigthSlice := right.Slice()

	leftOnlySlice, rightOnlySlice := CompareSliceString(leftSlice, rigthSlice)

	if len(leftOnlySlice) > 0 {
		leftOnly = &stringSetData{sliceData: leftOnlySlice}
	} else {
		leftOnly = nil
	}
	if len(rightOnlySlice) > 0 {
		rightOnly = &stringSetData{sliceData: rightOnlySlice}
	} else {
		rightOnly = nil
	}

	return leftOnly, rightOnly
}

package collections

import (
	"fmt"
	"sort"
	"strings"
)

var _genericSetMapThreshold = 4 // Use a map for fast lookup, if the size reaches this threshold

// GenericSet represents a set of arbitrary values, with no duplicates, sorted according to a specified comparison function
type GenericSet interface {
	Add(vals ...interface{}) (numDups int)
	Contains(val interface{}) bool
	Find(val interface{}) interface{}
	Len() int
	Slice() []interface{}
	String() string
}

type genericSetData struct {
	sliceData []interface{}
	mapData   map[string]interface{}
	keyFunc   func(interface{}) string
}

// NewGenericSet allocates a new GenericSet initialied with the supplied values
func NewGenericSet(keyFunc func(interface{}) string, vals ...interface{}) GenericSet {
	result := &genericSetData{keyFunc: keyFunc}
	for _, val := range vals {
		result.Add(val)
	}
	return result
}

func (gsd *genericSetData) Add(vals ...interface{}) (numDups int) {
	for _, val := range vals {
		if gsd.mapData != nil {
			key := gsd.keyFunc(val)
			if _, found := gsd.mapData[key]; found {
				numDups++
			} else {
				gsd.sliceData = append(gsd.sliceData, val)
				gsd.mapData[key] = val
			}
		} else {
			var isDup bool
			for _, elem := range gsd.sliceData {
				if gsd.keyFunc(elem) == gsd.keyFunc(val) {
					numDups++
					isDup = true
					break
				}
			}
			if !isDup {
				gsd.sliceData = append(gsd.sliceData, val)
				if len(gsd.sliceData) >= _genericSetMapThreshold {
					gsd.mapData = make(map[string]interface{})
					for _, e := range gsd.sliceData {
						gsd.mapData[gsd.keyFunc(e)] = e
					}
				}
			}
		}
	}
	return numDups
}

func (gsd *genericSetData) Contains(val interface{}) bool {
	return gsd.Find(val) != nil
}

func (gsd *genericSetData) Find(val interface{}) interface{} {
	if gsd.mapData != nil {
		if item, found := gsd.mapData[gsd.keyFunc(val)]; found {
			return item
		}
	} else {
		for _, elem := range gsd.sliceData {
			if gsd.keyFunc(elem) == gsd.keyFunc(val) {
				return elem
			}
		}
	}
	return nil
}

func (gsd *genericSetData) Len() int {
	return len(gsd.sliceData)
}

func (gsd *genericSetData) Slice() []interface{} {
	sort.Slice(gsd.sliceData, func(i, j int) bool {
		return gsd.keyFunc(gsd.sliceData[i]) < gsd.keyFunc(gsd.sliceData[j])
	})
	return gsd.sliceData
}

func (gsd *genericSetData) String() string {
	var result strings.Builder
	for i, v := range gsd.Slice() {
		if i > 0 {
			result.WriteString(`,`)
		}
		result.WriteString(fmt.Sprintf(`{%v}`, v))
	}
	return result.String()
}

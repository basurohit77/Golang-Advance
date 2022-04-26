// Package compare (part of the osscatalog project) contains utility functions to compare two data structures,
// esp. to compare a raw JSON returned from a REST API with the data structure in which we unmarshal it
package compare

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// Comparable is an interface to be implemented by objects that do not want to be simply compared by comparing each of their individual elements,
// but that define a special "ComparableString" method instead.
// Two objects are deemed identical if the "ComparableString" method returns identical strings for both objects
type Comparable interface {
	ComparableString() string
}

const comparableStringMethodName = "ComparableString"

var comparableType = reflect.TypeOf((*Comparable)(nil)).Elem()

// DeepCompare compares two complex objects (structs, arrays, maps, etc.) and produces a list of all differences
func DeepCompare(lName string, lVal interface{}, rName string, rVal interface{}, out *Output) {
	deepCompareValues(lName, reflect.ValueOf(lVal), rName, reflect.ValueOf(rVal), out)
}

// deepCompareValues compares two complex objects (structs, arrays, maps, etc.) represented as reflect.Values and produces a list of all differences
func deepCompareValues(lName string, lv reflect.Value, rName string, rv reflect.Value, out *Output) {
	if !lv.IsValid() && !rv.IsValid() {
		// Both zero/nil
		zv := reflect.ValueOf("<nil>")
		out.addValueEqual(lName, zv, rName, zv)
		return
	}
	if lv.IsValid() && !rv.IsValid() {
		out.addLOnly(lName, lv)
		return
	} else if !lv.IsValid() && rv.IsValid() {
		out.addROnly(rName, rv)
		return
	}
	lType := lv.Type()
	lKind := lType.Kind()
	rType := rv.Type()
	rKind := rType.Kind()
	if lKind != rKind {
		switch {
		case lKind == reflect.Struct && isMapOfInterfaces(rType):
			// special mapping to allow comparison with JSON unmarshall structs
			lv = reflect.ValueOf(convertStructValueToMap(lv))
			deepCompareMapValues(lName, lv, rName, rv, out)
			return
		case isMapOfInterfaces(lType) && rKind == reflect.Struct:
			// special mapping to allow comparison with JSON unmarshall structs
			rv = reflect.ValueOf(convertStructValueToMap(rv))
			deepCompareMapValues(lName, lv, rName, rv, out)
			return
		case lKind == reflect.Ptr && isMapOfInterfaces(rType):
			// special mapping to allow comparison with JSON unmarshall structs containing pointers for optional items
			lv = lv.Elem()
			deepCompareValues(lName, lv, rName, rv, out)
			return
		case isMapOfInterfaces(lType) && rKind == reflect.Ptr:
			// special mapping to allow comparison with JSON unmarshall structs containing pointers for optional items
			rv = rv.Elem()
			deepCompareValues(lName, lv, rName, rv, out)
			return
		case lKind == reflect.Array && rKind == reflect.Slice:
			lv = reflect.ValueOf(convertArrayValueToSlice(lv))
			deepCompareSliceValues(lName, lv, rName, rv, out)
			return
		case lKind == reflect.Slice && rKind == reflect.Array:
			rv = reflect.ValueOf(convertArrayValueToSlice(rv))
			deepCompareSliceValues(lName, lv, rName, rv, out)
			return
		case lKind == reflect.Interface && rKind != reflect.Interface:
			// XXX do we really want to allow comparing interface to non-interface?
			// -- needed for Array/Slice comparisons
			lv = lv.Elem()
			deepCompareValues(lName, lv, rName, rv, out)
			return
		case lKind != reflect.Interface && rKind == reflect.Interface:
			// XXX do we really want to allow comparing interface to non-interface?
			// -- needed for Array/Slice comparisons
			rv = rv.Elem()
			deepCompareValues(lName, lv, rName, rv, out)
			return
		case lKind == reflect.Float64 && rKind == reflect.Int64:
			// special mapping for numbers unmarshaled from JSON
			rv = reflect.ValueOf(float64(rv.Int()))
			deepCompareValues(lName, lv, rName, rv, out)
			return
		case lKind == reflect.Int64 && rKind == reflect.Float64:
			// special mapping for numbers unmarshaled from JSON
			lv = reflect.ValueOf(float64(lv.Int()))
			deepCompareValues(lName, lv, rName, rv, out)
			return
		default:
			out.addKindDiff(lName, lv, rName, rv)
			return
		}
	}
	var lBase, rBase interface{}
	if lType.Implements(comparableType) && rType.Implements(comparableType) {
		lcmp := lv.MethodByName(comparableStringMethodName)
		rcmp := rv.MethodByName(comparableStringMethodName)
		//		if !lcmp.IsZero() && !rcmp.IsZero() {
		lret := lcmp.Call(nil)
		rret := rcmp.Call(nil)
		if len(lret) == 1 && len(rret) == 1 {
			deepCompareValues(lName, lret[0], rName, rret[0], out)
			return
		}
		//		}
	}
	switch lKind {
	case reflect.Struct:
		lv = reflect.ValueOf(convertStructValueToMap(lv))
		rv = reflect.ValueOf(convertStructValueToMap(rv))
		deepCompareMapValues(lName, lv, rName, rv, out)
		return
	case reflect.Slice:
		deepCompareSliceValues(lName, lv, rName, rv, out)
		return
	case reflect.Array:
		lv = reflect.ValueOf(convertArrayValueToSlice(lv))
		rv = reflect.ValueOf(convertArrayValueToSlice(rv))
		deepCompareSliceValues(lName, lv, rName, rv, out)
		return
	case reflect.Map:
		deepCompareMapValues(lName, lv, rName, rv, out)
		return
	case reflect.Ptr, reflect.Interface:
		lv = lv.Elem()
		rv = rv.Elem()
		deepCompareValues(lName, lv, rName, rv, out)
		return
	case reflect.String:
		lBase = lv.String()
		rBase = rv.String()
	case reflect.Bool:
		lBase = lv.Bool()
		rBase = rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		lBase = lv.Int()
		rBase = rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		lBase = lv.Uint()
		rBase = rv.Uint()
	case reflect.Float32, reflect.Float64:
		lBase = lv.Float()
		rBase = rv.Float()
	default:
		// TODO: Handle comparison of additional user-defined base types
		// TODO: Handle comparison of int vs. uint, int vs. float
		if lType != rType {
			out.addKindDiff(lName, lv, rName, rv)
		} else if lv.Interface() != rv.Interface() {
			out.addValueDiff(lName, lv, rName, rv)
		} else {
			out.addValueEqual(lName, lv, rName, rv)
		}
		return
	}
	if lBase != rBase {
		out.addValueDiff(lName, lv, rName, rv)
	} else {
		out.addValueEqual(lName, lv, rName, rv)
	}
}

func deepCompareMapValues(lName string, lvin reflect.Value, rName string, rvin reflect.Value, out *Output) {
	// Special handling if we are called with a convertedStruct object
	lv, lFieldNames, lFieldOrder := checkForConvertedStruct(lvin)
	rv, rFieldNames, rFieldOrder := checkForConvertedStruct(rvin)

	lType := lv.Type()
	if lType.Kind() != reflect.Map {
		panic(fmt.Sprintf("deepCompareMapValues() called with left value not a map: %#v", lv))
	}
	rType := rv.Type()
	if rType.Kind() != reflect.Map {
		panic(fmt.Sprintf("deepCompareMapValues() called with right value not a map: %#v", rv))
	}

	rKey := rType.Key().Kind()
	lKey := lType.Key().Kind()
	rElem := rType.Elem().Kind()
	lElem := lType.Elem().Kind()
	//	rTypeStr := rType.String()
	//	lTypeStr := lType.String()
	//	fmt.Println("DEBUG Map types:", rTypeStr, lTypeStr)
	if !((rKey == lKey) && ((rElem == lElem) || (lElem == reflect.Interface) || (rElem == reflect.Interface))) {
		out.addKindDiff(lName, lv, rName, rv)
		return
	}

	// FIXME: Does not deal with keys that do not map uniquely to a string
	allKeys := make(map[string]bool)
	if lFieldOrder == nil {
		lFieldOrder = lv.MapKeys()
		sort.SliceStable(lFieldOrder, func(i, j int) bool {
			return lFieldOrder[i].String() < lFieldOrder[j].String()
		})
	}
	for _, k := range lFieldOrder {
		name := k.String()
		allKeys[name] = true
		leftEntryValue := lv.MapIndex(k)
		rightEntryValue := rv.MapIndex(k)
		if rightEntryValue.IsValid() {
			deepCompareValues(buildNamePath(lName, name, lFieldNames), leftEntryValue, buildNamePath(rName, name, rFieldNames), rightEntryValue, out)
		} else {
			out.addLOnly(buildNamePath(lName, name, lFieldNames), leftEntryValue)
		}
	}
	if rFieldOrder == nil {
		rFieldOrder = rv.MapKeys()
		sort.SliceStable(rFieldOrder, func(i, j int) bool {
			return rFieldOrder[i].String() < rFieldOrder[j].String()
		})
	}
	for _, k := range rFieldOrder {
		name := k.String()
		if allKeys[name] == false {
			rightEntryValue := rv.MapIndex(k)
			out.addROnly(buildNamePath(rName, name, rFieldNames), rightEntryValue)
		}
	}
}

func buildNamePath(basename string, name string, fieldNames map[string]string) string {
	/*
		var typeName string
		typ1 := val.Type()
		if typ1.Kind() == reflect.Interface {
			val2 := val.Elem()
			typeName = val2.Type().Name()
		} else {
			typeName = val.Type().Name()
		}
		if typeName == "convertedStruct" {
			return basename + "." + name
		}
	*/
	if fieldNames != nil {
		fn := fieldNames[name]
		return basename + "." + fn
	}
	return basename + "[" + name + "]"
}

func deepCompareSliceValues(lName string, lv reflect.Value, rName string, rv reflect.Value, out *Output) {
	debug.Debug(debug.Compare, `deepCompareSliceValues(): comparing %q vs. %q`, lv, rv)
	if lv.Type().Kind() != reflect.Slice {
		panic(fmt.Sprintf("deepCompareMap() called with left value not a slice: %#v", lv))
	}
	if rv.Type().Kind() != reflect.Slice {
		panic(fmt.Sprintf("deepCompareMap() called with right value not a slice: %#v", rv))
	}

	llen := lv.Len()
	rlen := rv.Len()

	// First check for special case where both slices are Comparable
	if llen > 0 {
		debug.Debug(debug.Compare, `deepCompareSliceValues(): left first element val=%v  type=%v  comparableType=%v`, lv.Index(0), lv.Index(0).Type(), lv.Index(0).Type().Implements(comparableType))
	}
	if rlen > 0 {
		debug.Debug(debug.Compare, `deepCompareSliceValues(): right first element val=%v  type=%v  comparableType=%v`, rv.Index(0), rv.Index(0).Type(), rv.Index(0).Type().Implements(comparableType))
	}
	if llen > 0 && rlen > 0 && lv.Index(0).Type().Implements(comparableType) && lv.Index(0).Type().Implements(comparableType) {
		lStrings := make([]string, llen)
		for i := 0; i < llen; i++ {
			elem := lv.Index(i)
			if elem.Type().Implements(comparableType) {
				lcmp := elem.MethodByName(comparableStringMethodName)
				lret := lcmp.Call(nil)
				if len(lret) == 1 {
					lStrings[i] = lret[0].String()
				} else {
					debug.Debug(debug.Compare, `deepCompareSliceValues(): left element %d is Comparable type but converts to unexpected value %v`, i, lret)
					goto fallback
				}
			} else {
				debug.Debug(debug.Compare, `deepCompareSliceValues(): left element %d is not Comparable type`, i)
				goto fallback
			}
		}
		if !sort.StringsAreSorted(lStrings) {
			debug.Debug(debug.Compare, `deepCompareSliceValues(): left slice is Comparable type but not sorted`)
			goto fallback
		}
		rStrings := make([]string, rlen)
		for i := 0; i < rlen; i++ {
			elem := rv.Index(i)
			if elem.Type().Implements(comparableType) {
				rcmp := elem.MethodByName(comparableStringMethodName)
				rret := rcmp.Call(nil)
				if len(rret) == 1 {
					rStrings[i] = rret[0].String()
				} else {
					debug.Debug(debug.Compare, `deepCompareSliceValues(): right element %d is Comparable type but converts to unexpected value %v`, i, rret)
					goto fallback
				}
			} else {
				debug.Debug(debug.Compare, `deepCompareSliceValues(): right element %d is not Comparable type`, i)
				goto fallback
			}
		}
		if !sort.StringsAreSorted(rStrings) {
			debug.Debug(debug.Compare, `deepCompareSliceValues(): right slice is Comparable type but not sorted`)
			goto fallback
		}

		var lix, rix int
		for lix < llen || rix < rlen {
			lElemName := lName + "[" + strconv.Itoa(lix) + "]"
			rElemName := rName + "[" + strconv.Itoa(rix) + "]"
			if lix < llen && rix < rlen && lStrings[lix] == rStrings[rix] {
				out.addValueEqual(lElemName, reflect.ValueOf(lStrings[lix]), rElemName, reflect.ValueOf(rStrings[rix]))
				lix++
				rix++
			} else if lix < llen && (rix >= rlen || lStrings[lix] < rStrings[rix]) {
				out.addLOnly(lElemName, reflect.ValueOf(lStrings[lix]))
				lix++
			} else if rix < rlen && (lix >= llen || lStrings[lix] > rStrings[rix]) {
				out.addROnly(rElemName, reflect.ValueOf(rStrings[rix]))
				rix++
			} else {
				panic(fmt.Sprintf(`deepCompareSliceValues() - unexpected state lix=%d/%d  rix=%d/%d\nlStrings=%q\nrStrings=%q`, lix, llen, rix, rlen, lStrings, rStrings))
			}
		}
		return
	}

	// Default handling: iterate through both slices and compare values at the same index
fallback:
	debug.Debug(debug.Compare, `deepCompareSliceValues(): fallback mode - elements not Comparable as strings`)
	switch {
	case llen == rlen:
		for i := 0; i < llen; i++ {
			deepCompareValues(lName+"["+strconv.Itoa(i)+"]", lv.Index(i), rName+"["+strconv.Itoa(i)+"]", rv.Index(i), out)
		}
	case llen < rlen:
		for i := 0; i < llen; i++ {
			deepCompareValues(lName+"["+strconv.Itoa(i)+"]", lv.Index(i), rName+"["+strconv.Itoa(i)+"]", rv.Index(i), out)
		}
		for i := llen; i < rlen; i++ {
			out.addROnly(rName+"["+strconv.Itoa(i)+"]", rv.Index(i))
		}
	case llen > rlen:
		for i := 0; i < rlen; i++ {
			deepCompareValues(lName+"["+strconv.Itoa(i)+"]", lv.Index(i), rName+"["+strconv.Itoa(i)+"]", rv.Index(i), out)
		}
		for i := rlen; i < llen; i++ {
			out.addLOnly(lName+"["+strconv.Itoa(i)+"]", lv.Index(i))
		}
	}
}

// convertedStruct is a special type that represents a struct that has been converted into a Map to facilitate comparison functions
type convertedStruct struct {
	TheMap     map[string]interface{}
	FieldNames map[string]string
	FieldOrder []reflect.Value
}

// checkForConvertedStruct checks if its argument is a Value associated with a convertedStruct object
// and if so, it returns the underlying Map containing the values and the Map of field names
func checkForConvertedStruct(v reflect.Value) (theMapValue reflect.Value, fieldNames map[string]string, fieldOrder []reflect.Value) {
	if v.Type().Name() == "convertedStruct" {
		theMapValue = v.FieldByName("TheMap")
		fieldNames = v.FieldByName("FieldNames").Interface().(map[string]string)
		fieldOrder = v.FieldByName("FieldOrder").Interface().([]reflect.Value)
		if theMapValue.Len() != len(fieldOrder) || len(fieldNames) != len(fieldOrder) {
			panic(fmt.Sprintf("ASSERTION FAILED: checkForConvertedStruct() mismatched lengths: TheMap.len=%d   FieldNames.len=%d   FieldOrder.len=%d", theMapValue.Len(), len(fieldNames), len(fieldOrder)))
		}
	} else {
		theMapValue = v
		fieldNames = nil
		fieldOrder = nil
	}
	return
}

// MapOfInterfaces is a map[string]interface{} such as those created by json.Unmarshall, that can be compared to a struct
type MapOfInterfaces map[string]interface{}

// isMapOfInterfaces returns true if the given Type represents a map[string]interface{} such as those created by json.Unmarshall, that can be compared to a struct
func isMapOfInterfaces(t reflect.Type) bool {
	if t.Kind() == reflect.Map && t.Key().Kind() == reflect.String && t.Elem().Kind() == reflect.Interface {
		return true
	}
	return false
}

// convertStructValueToMap() converts the reflect.Value of a Struct into a Map to facilitate comparison functions
// TODO: should populate the result with reflect.Value objects instead of interface{}
func convertStructValueToMap(s reflect.Value) convertedStruct {
	typ := s.Type()
	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("convertStructValueToMap() called with object not a structure: %#v", s))
	}
	numFields := typ.NumField()
	var result = convertedStruct{
		TheMap:     make(map[string]interface{}),
		FieldNames: make(map[string]string),
		FieldOrder: make([]reflect.Value, 0, typ.NumField()),
	}
	for i := 0; i < numFields; i++ {
		f := typ.Field(i)
		name := f.Name
		if f.PkgPath == "" {
			value := s.Field(i)
			jsonName := ""
			jsonModifier := ""
			var ok bool
			if jsonName, ok = f.Tag.Lookup("json"); ok {
				jsonName = strings.TrimSpace(jsonName)
				if ix := strings.Index(jsonName, ","); ix >= 0 {
					jsonModifier = strings.TrimSpace(jsonName[ix+1:])
					jsonName = strings.TrimSpace(jsonName[:ix])
				}
			}
			if f.Anonymous && (jsonName == "" || jsonModifier == "squash") { // Merge embedded struct inside the parent
				embedded := convertStructValueToMap(value)
				for k, v := range embedded.FieldNames {
					if _, found := result.FieldNames[name]; found {
						panic(fmt.Sprintf("convertStructValueToMap()(FieldNames) found duplicate field name \"%s\" in embedded field \"%s\" in struct %#v", name, k, s.Interface()))
					}
					result.FieldNames[k] = v
				}
				for k, v := range embedded.TheMap {
					if _, found := result.TheMap[name]; found {
						panic(fmt.Sprintf("convertStructValueToMap()(TheMap) found duplicate field name \"%s\" in embedded field \"%s\" in struct %#v", name, k, s.Interface()))
					}
					result.TheMap[k] = v
				}
				result.FieldOrder = append(result.FieldOrder, embedded.FieldOrder...)
			} else if jsonName != "" {
				if jsonName == "-" {
					continue
				}
				// The IsZero function only exists in Golang 1.13+, but not all services using the osscatalog library
				// are at this Golang level yet. Since this check is a nice to have, commenting out for now:
				/*if jsonModifier == "omitempty" && value.IsZero() {
					continue
				}*/
				if _, found := result.FieldNames[jsonName]; found {
					panic(fmt.Sprintf("convertStructValueToMap() found duplicate json tag \"%s\" in struct %#v", jsonName, s.Interface()))
				}
				result.FieldNames[jsonName] = name
				result.TheMap[jsonName] = value.Interface()
				result.FieldOrder = append(result.FieldOrder, reflect.ValueOf(jsonName))
			} else {
				result.FieldNames[name] = name
				result.TheMap[name] = value.Interface()
				result.FieldOrder = append(result.FieldOrder, reflect.ValueOf(name))
			}
		}
	}
	return result
}

// convertedArray is a slice containing all the elements extracted from an array
type convertedArray []interface{}

// convertArrayValueToSlice converts a reflect.Value representing an array of arbitrary types
// into a slice of interface{} objects for each element of the original array
// TODO: should populate the result with reflect.Value objects instead of interface{}
func convertArrayValueToSlice(a reflect.Value) convertedArray {
	len := a.Len()
	s := make(convertedArray, len)
	for i := 0; i < len; i++ {
		s[i] = a.Index(i).Interface()
	}
	return s
}

/*
func checkBadString(s string, msg string) {
	bytes := []byte(s)
	for _, b := range bytes {
		if b == byte(1) {
			fmt.Println("*** Got bad string", msg, s, bytes)
			return
		}
	}
}
*/

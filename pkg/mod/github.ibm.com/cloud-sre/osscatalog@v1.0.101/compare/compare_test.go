package compare

import (
	"reflect"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
)

func TestConvertStuctValueToMap(t *testing.T) {
	type innerStruct struct {
		FieldStringInner string
	}
	var in = struct {
		FieldString          string
		FieldZeroString      string
		FieldInt             int `json:"field_int"`
		fieldNotExported     string
		FieldStruct          innerStruct `json:"field_struct,omitempty"`
		FieldNil             *innerStruct
		FieldPointerToStruct *innerStruct
	}{
		FieldString:          "StringA",
		FieldZeroString:      "",
		FieldInt:             42,
		fieldNotExported:     "not exported",
		FieldStruct:          innerStruct{"StringB"},
		FieldPointerToStruct: &innerStruct{"StringC"},
	}

	result := convertStructValueToMap(reflect.ValueOf(in))

	//	fmt.Println("Result: ", result)

	if len(result.TheMap) != 6 {
		t.Errorf("Generated values map length=%d, expected 6", len(result.TheMap))
	}
	if len(result.FieldNames) != 6 {
		t.Errorf("Generated field names map length=%d, expected 6", len(result.FieldNames))
	}
	if len(result.FieldOrder) != 6 {
		t.Errorf("Generated field order list length=%d, expected 6", len(result.FieldOrder))
	}

	if result.TheMap["FieldString"].(string) != "StringA" {
		t.Errorf("Did not find expected value for FieldString in generated map: %#v  (expected \"StringA\")", result.TheMap["FieldString"])
	}
	if result.FieldNames["FieldString"] != "FieldString" {
		t.Errorf("Did not find expected field name for FieldString in generated map: %#v  (expected \"FieldString\")", result.FieldNames["FieldString"])
	}
	if result.FieldOrder[0].String() != "FieldString" {
		t.Errorf("Did not find expected field order for slot 0 - got %#v  (expected \"FieldString\")", result.FieldOrder[0].String())
	}
	if result.TheMap["FieldZeroString"].(string) != "" {
		t.Errorf("Did not find expected value for FieldZeroString in generated map: %#v  (expected \"\")", result.TheMap["FieldZeroString"])
	}
	if result.FieldOrder[1].String() != "FieldZeroString" {
		t.Errorf("Did not find expected field order for slot 1 - got %#v  (expected \"FieldZeroString\")", result.FieldOrder[1].String())
	}
	if result.TheMap["field_int"].(int) != 42 {
		t.Errorf("Did not find expected value for field_int in generated map: %#v  (expected 42)", result.TheMap["field_int"])
	}
	if result.FieldNames["field_int"] != "FieldInt" {
		t.Errorf("Did not find expected field name for field_int in generated map: %#v  (expected \"FieldInt\")", result.FieldNames["field_int"])
	}
	if result.FieldOrder[2].String() != "field_int" {
		t.Errorf("Did not find expected field order for slot 2 - got %#v  (expected \"field_int\")", result.FieldOrder[2].String())
	}
	if val, ok := result.TheMap["fieldNotExported"]; ok {
		t.Errorf("Found unexpected fieldNotExported in generated map: %#v", val)
	}
	if result.TheMap["field_struct"].(innerStruct) != (innerStruct{"StringB"}) {
		t.Errorf("Did not find expected value for field_struct in generated map: %#v  (expected {\"StringB\"})", result.TheMap["field_struct"])
	}
	if result.FieldNames["field_struct"] != "FieldStruct" {
		t.Errorf("Did not find expected field name for field_struct in generated map: %#v  (expected \"FieldStruct\")", result.FieldNames["field_struct"])
	}
	if result.FieldOrder[3].String() != "field_struct" {
		t.Errorf("Did not find expected field order for slot 3 - got %#v  (expected \"field_struct\")", result.FieldOrder[3].String())
	}
	if val, ok := result.TheMap["FieldNil"].(*innerStruct); !ok || val != nil {
		t.Errorf("Did not find expected value for FieldNil in generated map: %#v  (expected nil)", result.TheMap["FieldNil"])
	}
	if result.FieldOrder[4].String() != "FieldNil" {
		t.Errorf("Did not find expected field order for slot 4 - got %#v  (expected \"FieldNil\")", result.FieldOrder[4].String())
	}
	if *(result.TheMap["FieldPointerToStruct"].(*innerStruct)) != (innerStruct{"StringC"}) {
		t.Errorf("Did not find expected value for FieldFieldPointerToStruct in generated map: %#v  (expected {\"StringC\"})", result.TheMap["FieldPointerToStruct"])
	}
	if result.FieldOrder[5].String() != "FieldPointerToStruct" {
		t.Errorf("Did not find expected field order for slot 5 - got %#v  (expected \"FieldPointerToStruct\")", result.FieldOrder[5].String())
	}
}

func TestConvertArrayValueToSlice(t *testing.T) {
	in := [2]string{"elem1", "elem2"}

	result := convertArrayValueToSlice(reflect.ValueOf(in))
	if len(result) != 2 {
		t.Errorf("Expected len=2 got len=%d", len(result))
	}
	if result[0].(string) != "elem1" {
		t.Errorf("Expected result[0]=\"elem1\" got %#v", result[0])
	}
	if result[1].(string) != "elem2" {
		t.Errorf("Expected result[1]=\"elem2\" got %#v", result[1])
	}
}

func oneCompareTest(t *testing.T, lVal interface{}, rVal interface{}, includeEqual bool, expected []string) {
	out := Output{IncludeEqual: includeEqual}
	DeepCompare("left", lVal, "right", rVal, &out)
	diffs := out.ToStrings()
	if !reflect.DeepEqual(diffs, expected) {
		//		checkBadString(expected[0], "oneCompareTest.expected[0]")
		//		checkBadString(diffs[0], "oneCompareTest.diffs[0]")
		t.Errorf("%s: got %v  expected %v", t.Name(), diffs, expected)
	}
}

func TestDeepCompare(t *testing.T) {

	// testing for basic types
	t.Run("int=", func(t *testing.T) { oneCompareTest(t, 42, 42, false, nil) })
	t.Run("int!=", func(t *testing.T) { oneCompareTest(t, 42, 43, false, []string{`DIFF VALUE:      left=42    right=43`}) })
	t.Run("string=", func(t *testing.T) { oneCompareTest(t, "foo", "foo", false, nil) })
	t.Run("string!=", func(t *testing.T) {
		oneCompareTest(t, "foo", "bar", false, []string{`DIFF VALUE:      left="foo"    right="bar"`})
	})
	t.Run("string-int", func(t *testing.T) {
		oneCompareTest(t, "foo", 42, false, []string{`DIFF TYPE:       left.Type(string)=string    right.Type(int)=int`})
	})
	t.Run("string-nil", func(t *testing.T) { oneCompareTest(t, "foo", nil, false, []string{`DIFF LEFT ONLY:  left="foo"`}) })
	t.Run("nil-string", func(t *testing.T) {
		oneCompareTest(t, nil, "foo", false, []string{`DIFF RIGHT ONLY:         right="foo"`})
	})
	t.Run("string-zero", func(t *testing.T) {
		oneCompareTest(t, "foo", "", false, []string{`DIFF VALUE:      left="foo"    right=""`})
	})
	t.Run("zero-string", func(t *testing.T) {
		oneCompareTest(t, "", "foo", false, []string{`DIFF VALUE:      left=""    right="foo"`})
	})

	// testing for simple maps
	map0 := map[string]string{}
	map1 := map[string]string{"Key2": "Value2", "Key1": "Value1"} // Note non-alpha order of keys (should not matter for a map)
	map2 := map[string]string{"Key1": "Value1", "Key2": "Value2"}
	map3 := map[string]string{"Key1": "Value1", "Key2": "Value2.X"}
	map4 := map[string]string{"Key1": "Value1", "Key3": "Value3"}
	t.Run("map=", func(t *testing.T) { oneCompareTest(t, map1, map2, false, nil) })
	t.Run("map=(includeEqual", func(t *testing.T) { // Note expect the order Key1, Key2 (alphabetical for a map)
		oneCompareTest(t, map1, map2, true, []string{`EQUAL:           left[Key1]="Value1"`, `EQUAL:           left[Key2]="Value2"`})
	})
	t.Run("map!=values", func(t *testing.T) {
		oneCompareTest(t, map1, map3, false, []string{`DIFF VALUE:      left[Key2]="Value2"    right[Key2]="Value2.X"`})
	})
	t.Run("map!=values(includeEqual)", func(t *testing.T) {
		oneCompareTest(t, map1, map3, true, []string{`EQUAL:           left[Key1]="Value1"`, `DIFF VALUE:      left[Key2]="Value2"    right[Key2]="Value2.X"`})
	})
	t.Run("map!=keys", func(t *testing.T) {
		oneCompareTest(t, map1, map4, false, []string{`DIFF LEFT ONLY:  left[Key2]="Value2"`, `DIFF RIGHT ONLY:         right[Key3]="Value3"`})
	})
	t.Run("map-empty", func(t *testing.T) {
		oneCompareTest(t, map1, map0, false, []string{`DIFF LEFT ONLY:  left[Key1]="Value1"`, `DIFF LEFT ONLY:  left[Key2]="Value2"`})
	})
	t.Run("empty-map", func(t *testing.T) {
		oneCompareTest(t, map0, map1, false, []string{`DIFF RIGHT ONLY:         right[Key1]="Value1"`, `DIFF RIGHT ONLY:         right[Key2]="Value2"`})
	})

	// testing for simple slices
	slice1 := []string{"elem1", "elem2"}
	slice2 := []string{"elem1", "elem2"}
	slice3 := []string{"elem1", "elem2.X"}
	slice4 := []string{"elem1", "elem2", "elem3"}
	t.Run("slice=", func(t *testing.T) { oneCompareTest(t, slice1, slice2, false, nil) })
	t.Run("slice!=values", func(t *testing.T) {
		oneCompareTest(t, slice1, slice3, false, []string{`DIFF VALUE:      left[1]="elem2"    right[1]="elem2.X"`})
	})
	t.Run("slice-extra-right", func(t *testing.T) {
		oneCompareTest(t, slice1, slice4, false, []string{`DIFF RIGHT ONLY:         right[2]="elem3"`})
	})
	t.Run("slice-extra-left", func(t *testing.T) {
		oneCompareTest(t, slice4, slice1, false, []string{`DIFF LEFT ONLY:  left[2]="elem3"`})
	})

	// testing for simple structs
	struct1 := struct { // Note non-alpha order of fields (should be respected)
		Key2 string `json:"field_key2,omitempty"`
		Key1 string
	}{
		Key2: "Value2",
		Key1: "Value1",
	}
	struct2 := struct {
		Key1 string
		Key2 string `json:"field_key2"`
	}{
		Key1: "Value1",
		Key2: "Value2",
	}
	struct3 := struct {
		Key1 string
		Key2 string `json:"field_key2"`
	}{
		Key1: "Value1",
		Key2: "Value2.X",
	}
	struct4 := struct {
		Key1 string
		Key3 string `json:"field_key3"`
	}{
		Key1: "Value1",
		Key3: "Value3",
	}
	t.Run("struct=", func(t *testing.T) {
		oneCompareTest(t, struct1, struct2, false, nil)
	})
	t.Run("struct=(includeEqual)", func(t *testing.T) { // Note we expect the order Key2, Key1 (matches the order of fields in the struct declaration)
		oneCompareTest(t, struct1, struct2, true, []string{`EQUAL:           left.Key2="Value2"`, `EQUAL:           left.Key1="Value1"`})
	})
	t.Run("struct!=values", func(t *testing.T) {
		oneCompareTest(t, struct1, struct3, false, []string{`DIFF VALUE:      left.Key2="Value2"    right.Key2="Value2.X"`})
	})
	t.Run("struct!=values(includeEqual)", func(t *testing.T) {
		oneCompareTest(t, struct1, struct3, true, []string{`EQUAL:           left.Key1="Value1"`, `DIFF VALUE:      left.Key2="Value2"    right.Key2="Value2.X"`})
	})
	t.Run("struct!=fields", func(t *testing.T) {
		oneCompareTest(t, struct1, struct4, false, []string{`DIFF LEFT ONLY:  left.Key2="Value2"`, `DIFF RIGHT ONLY:         right.Key3="Value3"`})
	})

	// testing for struct pointers
	t.Run("*struct=", func(t *testing.T) { oneCompareTest(t, &struct1, &struct2, false, nil) })
	t.Run("*struct!=values", func(t *testing.T) {
		oneCompareTest(t, &struct1, &struct3, false, []string{`DIFF VALUE:      left.Key2="Value2"    right.Key2="Value2.X"`})
	})

	// testing for maps of interfaces
	mapIfc1 := MapOfInterfaces{"Key1": "Value1", "field_key2": "Value2"}
	mapIfc2 := MapOfInterfaces{"Key1": "Value1", "field_key2": "Value2"}
	mapIfc3 := MapOfInterfaces{"Key1": "Value1", "field_key2": "Value2.X"}
	mapIfc4 := MapOfInterfaces{"Key1": "Value1", "Key3": "Value3"}
	t.Run("map!=mapIfc", func(t *testing.T) {
		oneCompareTest(t, map1, mapIfc1, false, []string{`DIFF LEFT ONLY:  left[Key2]="Value2"`, `DIFF RIGHT ONLY:         right[field_key2]="Value2"`})
	})
	t.Run("mapIfc=", func(t *testing.T) {
		oneCompareTest(t, mapIfc1, mapIfc2, false, nil)
	})
	t.Run("mapIfc!=fields", func(t *testing.T) {
		oneCompareTest(t, mapIfc1, mapIfc4, false, []string{`DIFF LEFT ONLY:  left[field_key2]="Value2"`, `DIFF RIGHT ONLY:         right[Key3]="Value3"`})
	})
	t.Run("&mapIfc=mapIfc", func(t *testing.T) {
		oneCompareTest(t, &mapIfc1, mapIfc2, false, nil)
	})
	t.Run("mapIfc=&mapIfc", func(t *testing.T) {
		oneCompareTest(t, mapIfc1, &mapIfc2, false, nil)
	})

	// testing for structs against maps (of interfaces)
	t.Run("struct!=map", func(t *testing.T) {
		oneCompareTest(t, struct1, map1, false, []string{`DIFF TYPE:       left.Type(struct)=struct { Key2 string "json:\"field_key2,omitempty\""; Key1 string }    right.Type(map)=map[string]string`})
	})
	t.Run("struct=mapIfc", func(t *testing.T) {
		oneCompareTest(t, struct1, mapIfc1, false, nil)
	})
	t.Run("mapIfc=struct", func(t *testing.T) {
		oneCompareTest(t, mapIfc1, struct1, false, nil)
	})
	t.Run("struct!=mapIfc.values", func(t *testing.T) {
		oneCompareTest(t, struct1, mapIfc3, false, []string{`DIFF VALUE:      left.Key2="Value2"    right[field_key2]="Value2.X"`})
	})
	t.Run("struct!=mapIfc.fields", func(t *testing.T) {
		oneCompareTest(t, struct1, mapIfc4, false, []string{`DIFF LEFT ONLY:  left.Key2="Value2"`, `DIFF RIGHT ONLY:         right[Key3]="Value3"`})
	})
	t.Run("&struct=mapIfc", func(t *testing.T) {
		oneCompareTest(t, &struct1, mapIfc1, false, nil)
	})
	t.Run("mapIfc=&struct", func(t *testing.T) {
		oneCompareTest(t, mapIfc1, &struct1, false, nil)
	})

	// testing for nested structs and maps of interfaces
	structNested1 := struct {
		Outer1 string
		Inner1 struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}
	}{
		Outer1: "outer1",
	}
	structNested1.Inner1.Key1 = "Value1"
	structNested1.Inner1.Key2 = "Value2"
	structNested3 := struct {
		Outer1 string
		Inner1 struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}
	}{
		Outer1: "outer1",
	}
	structNested3.Inner1.Key1 = "Value1"
	structNested3.Inner1.Key2 = "Value2.X"
	nestedMapIfc3 := MapOfInterfaces{
		"Outer1": "outer1",
		"Inner1": MapOfInterfaces{"Key1": "Value1", "field_key2": "Value2.X"},
	}
	t.Run("nested-struct!=values", func(t *testing.T) {
		oneCompareTest(t, structNested1, structNested3, false, []string{`DIFF VALUE:      left.Inner1.Key2="Value2"    right.Inner1.Key2="Value2.X"`})
	})
	t.Run("nested-struct!=nested-mapIfc.values", func(t *testing.T) {
		oneCompareTest(t, structNested1, nestedMapIfc3, false, []string{`DIFF VALUE:      left.Inner1.Key2="Value2"    right[Inner1][field_key2]="Value2.X"`})
	})

	// testing for arrays
	array1 := [2]string{"elem1", "elem2"}
	array3 := [2]string{"elem1", "elem2.X"}
	t.Run("array!=values", func(t *testing.T) {
		oneCompareTest(t, array1, array3, false, []string{`DIFF VALUE:      left[1]="elem2"    right[1]="elem2.X"`})
	})
	t.Run("array!=slice", func(t *testing.T) {
		oneCompareTest(t, array1, slice3, false, []string{`DIFF VALUE:      left[1]="elem2"    right[1]="elem2.X"`})
	})
	t.Run("slice!=array", func(t *testing.T) {
		oneCompareTest(t, slice1, array3, false, []string{`DIFF VALUE:      left[1]="elem2"    right[1]="elem2.X"`})
	})

	// testing for user-defined types
	type MyString string
	var myString MyString = "theString"
	t.Run("user-defined-string=", func(t *testing.T) {
		oneCompareTest(t, myString, "theString", false, nil)
	})
	t.Run("user-defined-string!=", func(t *testing.T) {
		oneCompareTest(t, myString, "theStringX", false, []string{`DIFF VALUE:      left="theString"    right="theStringX"`})
	})
	t.Run("=user-defined-string", func(t *testing.T) {
		oneCompareTest(t, "theString", myString, false, nil)
	})
	type MyBool bool
	var myBool MyBool = true
	t.Run("user-defined-bool=", func(t *testing.T) {
		oneCompareTest(t, myBool, true, false, nil)
	})
	type MyUInt uint
	var myUInt MyUInt = 43
	t.Run("user-defined-uint=", func(t *testing.T) {
		oneCompareTest(t, myUInt, uint(43), false, nil)
	})
	type MyFloat32 float32
	var myFloat32 MyFloat32 = 42.1
	t.Run("user-defined-float32=", func(t *testing.T) {
		oneCompareTest(t, myFloat32, float32(42.1), false, nil)
	})

	// testing for pointers in struct members and nil pointers
	structWithPointer1 := struct {
		Outer1 string
		Inner1 *struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}
	}{
		Outer1: "outer1.1",
		Inner1: &struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}{
			Key1: "inner1.1",
		},
	}
	structWithPointer2 := struct {
		Outer1 string
		Inner1 *struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}
	}{
		Outer1: "outer1.2",
		Inner1: &struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}{
			Key1: "inner1.2",
		},
	}
	structWithPointer3 := struct {
		Outer1 string
		Inner1 *struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}
	}{
		Outer1: "outer1.3",
		Inner1: nil,
	}
	structWithPointer4 := struct {
		Outer1 string
		Inner1 *struct {
			Key1 string
			Key2 string `json:"field_key2"`
		}
	}{
		Outer1: "outer1.4",
		Inner1: nil,
	}
	t.Run("structWithPointer-nonnil-nonnil", func(t *testing.T) {
		oneCompareTest(t, structWithPointer1, structWithPointer2, true, []string{
			`EQUAL:           left.Inner1.Key2=""`,
			`DIFF VALUE:      left.Outer1="outer1.1"    right.Outer1="outer1.2"`,
			`DIFF VALUE:      left.Inner1.Key1="inner1.1"    right.Inner1.Key1="inner1.2"`})
	})
	t.Run("structWithPointer-nil-nil", func(t *testing.T) {
		oneCompareTest(t, structWithPointer3, structWithPointer4, true, []string{
			`EQUAL:           left.Inner1="<nil>"`,
			`DIFF VALUE:      left.Outer1="outer1.3"    right.Outer1="outer1.4"`})
	})
	t.Run("structWithPointer-nonnil-nil", func(t *testing.T) {
		oneCompareTest(t, structWithPointer1, structWithPointer4, false, []string{
			`DIFF VALUE:      left.Outer1="outer1.1"    right.Outer1="outer1.4"`,
			`DIFF LEFT ONLY:  left.Inner1=struct { Key1 string; Key2 string "json:\"field_key2\"" }{Key1:"inner1.1", Key2:""}`})
	})
	t.Run("structWithPointer-nil-nonnil", func(t *testing.T) {
		oneCompareTest(t, structWithPointer3, structWithPointer2, false, []string{
			`DIFF VALUE:      left.Outer1="outer1.3"    right.Outer1="outer1.2"`,
			`DIFF RIGHT ONLY:         right.Inner1=struct { Key1 string; Key2 string "json:\"field_key2\"" }{Key1:"inner1.2", Key2:""}`})
	})

}

type testComparable string

func (tc testComparable) ComparableString() string {
	return string(tc)
}

func oneSliceTest(t *testing.T, label string, left []testComparable, right []testComparable, includeEqual bool, expected []string) {
	t.Run(label, func(t *testing.T) {
		out := Output{IncludeEqual: includeEqual}
		deepCompareSliceValues("left", reflect.ValueOf(left), "right", reflect.ValueOf(right), &out)
		diffs := out.ToStrings()
		if !reflect.DeepEqual(diffs, expected) {
			t.Errorf("%s: got %v  expected %v", t.Name(), diffs, expected)
		}
	})
}

func TestDeepCompareSliceValuesAsString(t *testing.T) {
	debug.SetDebugFlags(0 /* | debug.Compare /* XXX */)

	oneSliceTest(t, "identical",
		[]testComparable{"a", "b"},
		[]testComparable{"a", "b"},
		true,
		[]string{
			`EQUAL:           left[0]="a"`,
			`EQUAL:           left[1]="b"`,
		})

	oneSliceTest(t, "left empty",
		[]testComparable{},
		[]testComparable{"a", "b"},
		true,
		[]string{
			`DIFF RIGHT ONLY:         right[0]="a"`,
			`DIFF RIGHT ONLY:         right[1]="b"`,
		})

	oneSliceTest(t, "right empty",
		[]testComparable{"a", "b"},
		[]testComparable{},
		true,
		[]string{
			`DIFF LEFT ONLY:  left[0]="a"`,
			`DIFF LEFT ONLY:  left[1]="b"`,
		})

	oneSliceTest(t, "left short",
		[]testComparable{"a", "b"},
		[]testComparable{"a", "b", "c"},
		true,
		[]string{
			`EQUAL:           left[0]="a"`,
			`EQUAL:           left[1]="b"`,
			`DIFF RIGHT ONLY:         right[2]="c"`,
		})

	oneSliceTest(t, "right short",
		[]testComparable{"a", "b", "c"},
		[]testComparable{"a", "b"},
		true,
		[]string{
			`EQUAL:           left[0]="a"`,
			`EQUAL:           left[1]="b"`,
			`DIFF LEFT ONLY:  left[2]="c"`,
		})

	oneSliceTest(t, "insert left middle",
		[]testComparable{"a", "b", "b2", "b3", "c"},
		[]testComparable{"a", "b", "c"},
		true,
		[]string{
			`EQUAL:           left[0]="a"`,
			`EQUAL:           left[1]="b"`,
			`EQUAL:           left[4]="c"`,
			`DIFF LEFT ONLY:  left[2]="b2"`,
			`DIFF LEFT ONLY:  left[3]="b3"`,
		})

	oneSliceTest(t, "insert right middle",
		[]testComparable{"a", "b", "c"},
		[]testComparable{"a", "b", "b1", "b2", "c"},
		true,
		[]string{
			`EQUAL:           left[0]="a"`,
			`EQUAL:           left[1]="b"`,
			`EQUAL:           left[2]="c"`,
			`DIFF RIGHT ONLY:         right[2]="b1"`,
			`DIFF RIGHT ONLY:         right[3]="b2"`,
		},
	)

	oneSliceTest(t, "insert left and right middle",
		[]testComparable{"a", "b", "b2", "c"},
		[]testComparable{"a", "b", "b1", "c"},
		true,
		[]string{
			`EQUAL:           left[0]="a"`,
			`EQUAL:           left[1]="b"`,
			`EQUAL:           left[3]="c"`,
			`DIFF LEFT ONLY:  left[2]="b2"`,
			`DIFF RIGHT ONLY:         right[2]="b1"`,
		})

	oneSliceTest(t, "insert left start",
		[]testComparable{"0a", "a", "b"},
		[]testComparable{"a", "b"},
		true,
		[]string{
			`EQUAL:           left[1]="a"`,
			`EQUAL:           left[2]="b"`,
			`DIFF LEFT ONLY:  left[0]="0a"`,
		})

	oneSliceTest(t, "insert right start",
		[]testComparable{"a", "b"},
		[]testComparable{"0a", "a", "b"},
		true,
		[]string{
			`EQUAL:           left[0]="a"`,
			`EQUAL:           left[1]="b"`,
			`DIFF RIGHT ONLY:         right[0]="0a"`,
		})

	oneSliceTest(t, "insert left and right start",
		[]testComparable{"1a", "a", "b"},
		[]testComparable{"0a", "a", "b"},
		true,
		[]string{
			`EQUAL:           left[1]="a"`,
			`EQUAL:           left[2]="b"`,
			`DIFF LEFT ONLY:  left[0]="1a"`,
			`DIFF RIGHT ONLY:         right[0]="0a"`,
		})
}

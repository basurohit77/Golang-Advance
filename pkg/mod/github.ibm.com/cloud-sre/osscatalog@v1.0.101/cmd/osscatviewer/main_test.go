package main

import (
	"net/url"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadOneFormField(t *testing.T) {
	var values = url.Values{}
	values.Add("zero string", "")
	values.Add("empty string", `""`)
	values.Add("string1", "string1-value")
	values.Add("string2", `"string2-value"`)
	values.Add("empty slice", "[]")
	values.Add("slice1", ` [ "item1",   "item2" ]`)
	values.Add("empty map", "{}")
	values.Add("map1", `{"key1": "value1", "key2": "value2"}`)

	var errorBuffer = strings.Builder{}
	var testString string
	var testSlice []string
	var testTagSet osstags.TagSet
	var testMap map[string]string

	readOneFormField(values, "zero string", &testString, &errorBuffer)
	testhelper.AssertEqual(t, "zero string", "", testString)

	readOneFormField(values, "empty string", &testString, &errorBuffer)
	testhelper.AssertEqual(t, "empty string", "", testString)

	readOneFormField(values, "string1", &testString, &errorBuffer)
	testhelper.AssertEqual(t, "string1", "string1-value", testString)

	readOneFormField(values, "string2", &testString, &errorBuffer)
	testhelper.AssertEqual(t, "string2", "string2-value", testString)

	readOneFormField(values, "empty slice", &testSlice, &errorBuffer)
	testhelper.AssertEqual(t, "empty slice", []string{}, testSlice)

	readOneFormField(values, "slice1", &testSlice, &errorBuffer)
	testhelper.AssertEqual(t, "slice1", []string{"item1", "item2"}, testSlice)

	readOneFormField(values, "slice1", &testTagSet, &errorBuffer)
	testhelper.AssertEqual(t, "tagset1", osstags.TagSet{"item1", "item2"}, testTagSet)

	readOneFormField(values, "empty map", &testMap, &errorBuffer)
	testhelper.AssertEqual(t, "empty map", map[string]string{}, testMap)

	readOneFormField(values, "map1", &testMap, &errorBuffer)
	testhelper.AssertEqual(t, "map1", map[string]string{"key1": "value1", "key2": "value2"}, testMap)

	if errorBuffer.Len() > 0 {
		t.Errorf(`Expected empty errorBuffer but got \n %s`, errorBuffer.String())
	}
}

func TestReadOneFormFieldMismatch(t *testing.T) {
	var values = url.Values{}
	values.Add("string1", "string1-value")
	values.Add("slice1", ` [ "item1",   "item2" ]`)

	var errorBuffer = strings.Builder{}
	var testString string
	var testSlice []string

	readOneFormField(values, "string1", &testSlice, &errorBuffer)
	testhelper.AssertEqual(t, "string1", ([]string)(nil), testSlice)

	readOneFormField(values, "slice1", &testString, &errorBuffer)
	testhelper.AssertEqual(t, "tagset1", "", testString)

	testhelper.AssertEqual(t, "errors",
		"Error parsing value for form field \"string1\" (string1-value): invalid character 's' looking for beginning of value\nError parsing value for form field \"slice1\" ([ \"item1\",   \"item2\" ]): json: cannot unmarshal array into Go value of type string\n",
		errorBuffer.String())
}

func TestGetUniversalLink(t *testing.T) {
	var link string

	link = getUniversalLink("my-service-name")
	testhelper.AssertEqual(t, "OSS entry", "/view/my-service-name", link)

	link = getUniversalLink(`{Some Entry Name [chid:12345GH56]}`)
	testhelper.AssertEqual(t, "CH entry", "https://clearinghousev2.raleigh.ibm.com/CHNewCHRDM/CCHMServlet#&nature=wlhNDE&deliverableId=12345GH56", link)

	link = getUniversalLink("~~~ something")
	testhelper.AssertEqual(t, "something else", "", link)

}

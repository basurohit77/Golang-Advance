package ossmerge

import (
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"

	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

type FromStruct struct {
	RMCNumber         string                      `json:"rmc_number"`
	OperationalStatus ossrecord.OperationalStatus `json:"operational_status"`
	ClientFacing      bool                        `json:"client_facing"`
	EntryType         ossrecord.EntryType         `json:"entry_type"`
	FullCRN           string                      `json:"full_crn"`
	OSSDescription    string                      `json:"oss_description"`
	OnlyInFrom        string                      `json:"only_in_from"`
}

type ToStruct struct {
	RMCNumber         string                      `json:"rmc_number"`
	OperationalStatus ossrecord.OperationalStatus `json:"operational_status"`
	ClientFacing      bool                        `json:"client_facing"`
	EntryType         ossrecord.EntryType         `json:"entry_type"`
	FullCRN           string                      `json:"full_crn"`
	OSSDescription    string                      `json:"oss_description"`
	OnlyInTo          string                      `json:"only_in_to"`
}

func TestMergeStruct(t *testing.T) {
	from := FromStruct{
		OperationalStatus: ossrecord.GA,
		ClientFacing:      true,
		OSSDescription:    "From description",
		OnlyInFrom:        "only in From",
	}

	to := ToStruct{
		OperationalStatus: ossrecord.BETA,
		ClientFacing:      false,
		OSSDescription:    "To description",
		OnlyInTo:          "only in To",
	}

	MergeStruct(&to, from)

	if to.OnlyInTo != "only in To" {
		t.Error("Lost content of OnlyInTo field")
	}
	if to.OperationalStatus != ossrecord.GA {
		t.Error("Did not override OperationalStatus field")
	}
	if to.ClientFacing != true {
		t.Error("Did not override ClientFacing field")
	}
	if to.OSSDescription != "From description" {
		t.Error("Did not override OSSDescription field")
	}

	//	fmt.Println("From: ", from)
	//	fmt.Println("To:   ", to)

}

func TestMergeValues(t *testing.T) {

	type testDataStruct struct {
		title          string
		hasServiceNow  bool
		servicenowVal  interface{}
		hasScorecardV1 bool
		scorecardv1Val interface{}
		hasPriorOSS    bool
		priorOSSVal    interface{}
		expected       interface{}
		validations    []string
	}
	const (
		Diff         = "has different values from different sources (first source prevails)"
		MissingSN    = "is missing from ServiceNow"
		MissingSC    = "is missing from ScorecardV1"
		NotSet       = "cannot be set in OSS record from any available source"
		NoSources    = "cannot be set in OSS record because there are no source records containing this attribute"
		FromPriorOSS = "value taken from PriorOSS record"
	)
	testFunc := func(t *testing.T, td testDataStruct) {
		si := &ServiceInfo{}
		si.OSSMergeControl = ossmergecontrol.New("the-service-name")
		si.OSSValidation = ossvalidation.New("the-service-name", "test-timestamp")
		si.OSSService.GeneralInfo.RMCNumber = "str-oss"
		if td.hasServiceNow {
			si.SourceServiceNow.CRNServiceName = "the-service-name"
		} else {
			si.SourceServiceNow.CRNServiceName = ""
		}
		if td.hasScorecardV1 {
			si.SourceScorecardV1Detail.Name = "the-service-name"
		} else {
			si.SourceScorecardV1Detail.Name = ""
		}
		if td.hasPriorOSS {
			si.PriorOSS.ReferenceResourceName = "the-resource-name"
		} else {
			si.PriorOSS.ReferenceResourceName = ""
		}
		var testParams []parameter
		testParams = append(testParams, SeverityIfMissing{V: ossvalidation.WARNING})
		testParams = append(testParams, TagIfMissing{V: ossvalidation.TagStatusPage})
		if td.servicenowVal != nil {
			testParams = append(testParams, ServiceNow{V: td.servicenowVal})
		}
		if td.scorecardv1Val != nil {
			testParams = append(testParams, ScorecardV1{V: td.scorecardv1Val})
		}
		if td.priorOSSVal != nil {
			testParams = append(testParams, PriorOSS{V: td.priorOSSVal})
		}
		result := si.MergeValues(td.title, testParams...)

		if result != td.expected {
			t.Errorf("%s mismatched result- expected %#v got %#v", t.Name(), td.expected, result)
		}
		var numValidations int
		ossValidations := len(si.OSSValidation.Issues)
		if len(td.validations) > ossValidations {
			numValidations = len(td.validations)
		} else {
			numValidations = ossValidations
		}
		for i := 0; i < numValidations; i++ {
			switch {
			case i < len(td.validations) && i < ossValidations:
				if !strings.HasSuffix(si.OSSValidation.Issues[i].Title, td.validations[i]) {
					t.Errorf("%s mismatched ValidationIssue - expected %#v got %#v", t.Name(), td.validations[i], si.OSSValidation.Issues[i].Title)
				}
			case i >= len(td.validations):
				t.Errorf("%s mismatched ValidationIssue - expected %#v got %#v", t.Name(), "", si.OSSValidation.Issues[i].Title)
			case i >= ossValidations:
				t.Errorf("%s mismatched ValidationIssue - expected %#v got %#v", t.Name(), td.validations[i], "")
			}
		}
	}
	testData := []testDataStruct{
		{"string A and A", true, "str-A", true, "str-A", true, "str-P", "str-A", []string{}},
		{"string A and B", true, "str-A", true, "str-B", true, "str-P", "str-A", []string{Diff}},
		{"string ZERO and B", true, "", true, "str-B", true, "str-P", "str-B", []string{MissingSN}},
		{"string A and ZERO", true, "str-A", true, "", true, "str-P", "str-A", []string{MissingSC}},
		{"string ZERO and ZERO", true, "", true, "", true, "str-P", "", []string{MissingSN, MissingSC, NotSet}},
		{"int 1 and 2", true, 1, true, 2, true, 99, 1, []string{Diff}},
		{"bool true and false", true, true, true, false, true, false, true, []string{Diff}},
		{"bool false and true", true, false, true, true, true, true, false, []string{Diff}}, // XXX cannot tell if boolean is present but false, or not present
		{"string NONE and B", false, "str-A", true, "str-B", true, "str-P", "str-B", []string{}},
		{"string A and NONE", true, "str-A", false, "str-B", true, "str-P", "str-A", []string{}},
		{"string NONE and NONE", false, "str-A", false, "str-B", true, "str-P", "", []string{NoSources}},
		{"Person A and B", true, ossrecord.Person{Name: "nameA", W3ID: "emailA"}, true, ossrecord.Person{Name: "nameB", W3ID: "emailB"}, true, ossrecord.Person{Name: "nameP", W3ID: "emailP"}, ossrecord.Person{Name: "nameA", W3ID: "emailA"}, []string{Diff}},
		{"Person ZERO and B", true, ossrecord.Person{Name: "", W3ID: ""}, true, ossrecord.Person{Name: "nameB", W3ID: "emailB"}, true, ossrecord.Person{Name: "nameP", W3ID: "emailP"}, ossrecord.Person{Name: "nameB", W3ID: "emailB"}, []string{MissingSN}},
		{"Person partial-ZERO and B", true, ossrecord.Person{Name: "", W3ID: "emailA"}, true, ossrecord.Person{Name: "nameB", W3ID: "emailB"}, true, ossrecord.Person{Name: "nameP", W3ID: "emailP"}, ossrecord.Person{Name: "", W3ID: "emailA"}, []string{Diff}},
		{"Person NONE and B", false, ossrecord.Person{Name: "nameA", W3ID: "emailA"}, true, ossrecord.Person{Name: "nameB", W3ID: "emailB"}, true, ossrecord.Person{Name: "nameP", W3ID: "emailP"}, ossrecord.Person{Name: "nameB", W3ID: "emailB"}, []string{}},
		{"Person NONE and NONE", false, ossrecord.Person{Name: "nameA", W3ID: "emailA"}, false, ossrecord.Person{Name: "nameB", W3ID: "emailB"}, true, ossrecord.Person{Name: "nameP", W3ID: "emailP"}, ossrecord.Person{Name: "", W3ID: ""}, []string{NoSources}},
		{"string A and nil", true, "str-A", true, nil, true, "str-P", "str-A", []string{}},
		{"string ZERO and nil", true, "", true, nil, true, "str-P", "", []string{MissingSN, NotSet}},
		{"string nil and B", true, nil, true, "str-B", true, "str-P", "str-B", []string{}},
		{"string nil and ZERO", true, nil, true, "", true, "str-P", "", []string{MissingSC, NotSet}},
		{"bool true and nil", true, true, true, nil, true, false, true, []string{}},
		{"bool false and nil", true, false, true, nil, true, true, false, []string{}},
		{"bool nil and true", true, nil, true, true, true, false, true, []string{}},
		{"bool nil and false", true, nil, true, false, true, true, false, []string{}},
		{"string nil and nil + prior", true, nil, true, nil, true, "str-P", "str-P", []string{ /*FromPriorOSS*/ }},
		//{"string nil and nil", true, nil, true, nil, ""},
	}

	savedScorecardV1RunAction := ossrunactions.ScorecardV1.IsEnabled()
	ossrunactions.Enable([]string{"ScorecardV1"})

	for _, td := range testData {
		t.Run(td.title, func(t *testing.T) { testFunc(t, td) })
	}

	if !savedScorecardV1RunAction {
		ossrunactions.Disable([]string{"ScorecardV1"})
	}
}

func TestDeepCopy(t *testing.T) {
	type MyType struct {
		Attr1 string   `json:"attr1"`
		Attr2 []string `json:"attr2"`
	}

	src := MyType{
		Attr1: "attribute1.A",
		Attr2: []string{"attribute2.A.1", "attribute2.A.2"},
	}
	dest := MyType{}

	err := DeepCopy(&dest, src)
	testhelper.AssertError(t, err)

	testhelper.AssertEqual(t, "Attr1", src.Attr1, dest.Attr1)
	testhelper.AssertEqual(t, "Attr2", src.Attr2, dest.Attr2)

	dest.Attr1 = "attribute1.B"
	dest.Attr2 = append(dest.Attr2, "attribute2.B.3")
	testhelper.AssertEqual(t, "Attr1 modified", "attribute1.B", dest.Attr1)
	testhelper.AssertEqual(t, "Attr2 modified", []string{"attribute2.A.1", "attribute2.A.2", "attribute2.B.3"}, dest.Attr2)
	testhelper.AssertEqual(t, "Attr1 src", "attribute1.A", src.Attr1)
	testhelper.AssertEqual(t, "Attr2 src", []string{"attribute2.A.1", "attribute2.A.2"}, src.Attr2)
}

func TestCompareValues(t *testing.T) {
	a := []string{"x", "y", "z"}
	b := []string{"x", "y", "z"}
	comparsionResult := compareValues(a, b)
	testhelper.AssertEqual(t, "Same array content and same order", true, comparsionResult)

	a = []string{"x", "y", "z"}
	b = []string{"z", "y", "x"}
	comparsionResult = compareValues(a, b)
	testhelper.AssertEqual(t, "Same array content but different order", true, comparsionResult)

	a = []string{"x", "y", "z"}
	b = []string{"x1", "y", "z"}
	comparsionResult = compareValues(a, b)
	testhelper.AssertEqual(t, "Different array value at beginning", false, comparsionResult)

	a = []string{"x", "y", "z"}
	b = []string{"x", "y1", "z"}
	comparsionResult = compareValues(a, b)
	testhelper.AssertEqual(t, "Different array value at middle", false, comparsionResult)

	a = []string{"x", "y", "z"}
	b = []string{"x", "y", "z1"}
	comparsionResult = compareValues(a, b)
	testhelper.AssertEqual(t, "Different array value at end", false, comparsionResult)

	a = []string{"x", "y", "z"}
	b = []string{"x", "y"}
	comparsionResult = compareValues(a, b)
	testhelper.AssertEqual(t, "Different array sizes", false, comparsionResult)

	a = []string{}
	b = []string{}
	comparsionResult = compareValues(a, b)
	testhelper.AssertEqual(t, "Empty arrays", true, comparsionResult)
}

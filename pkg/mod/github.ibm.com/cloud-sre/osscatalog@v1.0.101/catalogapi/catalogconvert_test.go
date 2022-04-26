package catalogapi

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestGetOperationalStatus(t *testing.T) {
	type testDataStruct struct {
		title          string
		inputTags      []string
		expectedResult ossrecord.OperationalStatus
		expectedError  string
	}
	testData := []testDataStruct{
		{"simple-beta", []string{"foo", "ibm_beta", "bar"}, ossrecord.BETA, ``},
		{"simple-ga", []string{"foo", "bar"}, ossrecord.GA, ``},
		{"conflicting-tags", []string{"foo", "ibm_beta", "bar", "ibm_experimental"}, ossrecord.OperationalStatusUnknown, `Conflicting Catalog tags for support level: ["foo" "ibm_beta" "bar" "ibm_experimental"] -- resetting`},
		{"duplicate-tags", []string{"foo", "ibm_beta", "bar", "ibm_beta"}, ossrecord.OperationalStatusUnknown, `Conflicting Catalog tags for support level: ["foo" "ibm_beta" "bar" "ibm_beta"] -- resetting`},
		{"empty", []string{}, ossrecord.GA, ``},
		{"ibm_created", []string{"ibm_created"}, ossrecord.GA, ``},
		{"ibm_created and beta", []string{"ibm_created", "ibm_beta"}, ossrecord.BETA, ``},
		{"ibm_third_party", []string{"ibm_third_party"}, ossrecord.THIRDPARTY, ``},
		{"ibm_third_party and beta", []string{"ibm_third_party", "ibm_beta"}, ossrecord.THIRDPARTY, `Catalog tags specify both "ibm_third_party" and other sources/levels: ["ibm_third_party" "ibm_beta"] -- "ibm_third_party" takes precedence`},
		{"ibm_created and ibm_third_party", []string{"ibm_created", "ibm_third_party"}, ossrecord.GA, `Catalog tags specify both "ibm_created" and other sources: ["ibm_created" "ibm_third_party"] -- "ibm_created" takes precedence`},
		{"ibm_created and ibm_third_party and beta", []string{"ibm_created", "ibm_third_party", "ibm_beta"}, ossrecord.BETA, `Catalog tags specify both "ibm_created" and other sources: ["ibm_created" "ibm_third_party" "ibm_beta"] -- "ibm_created" takes precedence`},
		{"ibm_created and ibm_community", []string{"ibm_created", "ibm_community"}, ossrecord.GA, `Catalog tags specify both "ibm_created" and other sources: ["ibm_created" "ibm_community"] -- "ibm_created" takes precedence`},
		{"ibm_third_party and ibm_community", []string{"ibm_third_party", "ibm_community"}, ossrecord.THIRDPARTY, `Catalog tags specify both "ibm_third_party" and other sources/levels: ["ibm_third_party" "ibm_community"] -- "ibm_third_party" takes precedence`},
		{"ibm_community", []string{"ibm_community"}, ossrecord.COMMUNITY, ``},
		{"ibm_third_party and community", []string{"ibm_third_party", "ibm_community"}, ossrecord.THIRDPARTY, `Catalog tags specify both "ibm_third_party" and other sources/levels: ["ibm_third_party" "ibm_community"] -- "ibm_third_party" takes precedence`},
	}

	for _, td := range testData {
		t.Run(td.title, func(t *testing.T) {
			r := Resource{Tags: td.inputTags}
			result, err := r.GetOperationalStatus()
			testhelper.AssertEqual(t, "result", td.expectedResult, result)
			if err != nil {
				testhelper.AssertEqual(t, "result error", td.expectedError, err.Error())
			} else {
				testhelper.AssertEqual(t, "result error", td.expectedError, "")
			}
		})
	}
}

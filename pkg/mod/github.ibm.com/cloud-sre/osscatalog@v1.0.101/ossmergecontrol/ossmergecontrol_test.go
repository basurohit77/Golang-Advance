package ossmergecontrol

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"

	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

func createTestRecord() *OSSMergeControl {
	result := &OSSMergeControl{
		CanonicalName:     "oss-catalog-testing",
		OSSTags:           osstags.TagSet{"tag1", "tag2"},
		RawDuplicateNames: []string{"name1", "name2"},
		DoNotMergeNames:   []string{"name3"},
		Notes:             "some notes",
		Overrides: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		LastUpdate: "2018-07-27 16:00 UTC",
		UpdatedBy:  "none@nowhere.com",
	}
	result.RefreshChecksum(true)
	return result
}

func TestRefreshChecksum(t *testing.T) {
	f := func(title string, ossc *OSSMergeControl, expected bool) string {
		valid := ossc.RefreshChecksum(true)
		if *testhelper.VeryVerbose {
			fmt.Printf("Checksum for %s: %v\n", title, ossc.Checksum)
		}
		testhelper.AssertEqual(t, title+" - is valid", expected, valid)
		l := len(ossc.Checksum)
		if l == 0 {
			t.Errorf("Unexpected length for Checksum for %s", title)
		}
		return ossc.Checksum
	}
	empty := &OSSMergeControl{}
	f("empty record", empty, true)
	ossc := createTestRecord()
	tags1 := ossc.OSSTags
	cksum1 := f("original full record", ossc, true)
	cksum2 := f("original full record (repeat)", ossc, true)
	testhelper.AssertEqual(t, "repeated checksum", cksum1, cksum2)
	ossc.OSSTags.AddTag("tag3")
	cksum3 := f("after modification", ossc, false)
	if cksum3 == cksum2 {
		t.Errorf("Expected different checksum after modification  - got %v", cksum3)
	}
	ossc.OSSTags = tags1
	cksum4 := f("after restore to original value", ossc, false)
	testhelper.AssertEqual(t, "checksum after restore original value", cksum1, cksum4)
}

func TestIsEmpty(t *testing.T) {
	empty := &OSSMergeControl{}
	testhelper.AssertEqual(t, "empty record", true, empty.IsEmpty())
	empty.CanonicalName = "oss-catalog-testing"
	empty.LastUpdate = "2018-07-27 16:00 UTC"
	empty.UpdatedBy = "none@nowhere.com"
	empty.RefreshChecksum(true)
	testhelper.AssertEqual(t, "empty record with metadata fields", true, empty.IsEmpty())
	ossc := createTestRecord()
	testhelper.AssertEqual(t, "full record", false, ossc.IsEmpty())
}

func TestOneLineString(t *testing.T) {
	rec1 := createTestRecord()
	out1 := rec1.OneLineString()
	testhelper.AssertEqual(t, "full record", `{"canonical_name":"oss-catalog-testing","oss_tags":["tag1","tag2"],"raw_duplicate_names":["name1","name2"],"do_not_merge_names":["name3"],"notes":"some notes","overrides":{"key1":"value1","key2":"value2"},"last_update":"2018-07-27 16:00 UTC","updated_by":"none@nowhere.com"}`, out1)
	rec2 := New("test-record")
	rec2.OSSTags = osstags.TagSet{"tag1"}
	buf := strings.Builder{}
	fmt.Fprintln(&buf, "line1")
	fmt.Fprintln(&buf, "line2")
	rec2.Notes = buf.String()
	out2 := rec2.OneLineString()
	testhelper.AssertEqual(t, "sparse record", `{"canonical_name":"test-record","oss_tags":["tag1"],"notes":"line1\nline2\n"}`, out2)
	if *testhelper.VeryVerbose {
		fmt.Printf("out2=%sEND\n", out2)
	}
}

func TestAddOverride(t *testing.T) {
	rec1 := createTestRecord()
	validOverrides["Override_String"] = reflect.TypeOf(string("bar"))
	validOverrides["Override_CRNServiceName"] = reflect.TypeOf(ossrecord.CRNServiceName("bar"))

	test := func(t *testing.T, name string, value string, expected interface{}) {
		err := rec1.AddOverride(name, value)
		switch expected1 := expected.(type) {
		case error:
			if err == nil {
				t.Errorf("%s: expected an error but got success (result=%v)", name, rec1.Overrides[name])
			} else if expected1.Error() != err.Error() {
				t.Errorf(`%s: got unexpected error "%s"  (expected "%s")`, name, err.Error(), expected1.Error())
			}
		default:
			testhelper.AssertError(t, err)
			testhelper.AssertEqual(t, name, expected1, rec1.Overrides[name])
		}
	}
	test(t, "Override_String", "test-string", string("test-string"))
	test(t, "Override_CRNServiceName", "test-crn", ossrecord.CRNServiceName("test-crn"))
	test(t, "Override_BAD", "test-bad", fmt.Errorf(`Invalid Override "Override_BAD=test-bad" for entry oss-catalog-testing`))
}

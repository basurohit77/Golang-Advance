package osstags

import (
	"fmt"
	"reflect"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestAddTag(t *testing.T) {
	var tags = TagSet{}

	tags.AddTag(NotReady)
	tags.AddTag(NotReady)
	tags.AddTag(TypeOtherOSS)

	testhelper.AssertEqual(t, "", 2, len(tags))
	testhelper.AssertEqual(t, "", NotReady, tags[0])
	testhelper.AssertEqual(t, "", TypeOtherOSS, tags[1])
}

func TestRemoveTag(t *testing.T) {
	var tags = TagSet{Deprecated, PnPEnabled, NotReady + ">490102", OneCloud}

	testhelper.AssertEqual(t, "before remove", TagSet{Deprecated, PnPEnabled, NotReady + ">490102", OneCloud}, tags)

	tags.RemoveTag(PnPEnabled)
	tags.RemoveTag(NotReady)
	testhelper.AssertEqual(t, "after 1st remove", TagSet{Deprecated, OneCloud}, tags)

	tags.RemoveTag(PnPEnabled)
	testhelper.AssertEqual(t, "after 2nd remove", TagSet{Deprecated, OneCloud}, tags)
}

func TestRemoveTagWithDuplicate(t *testing.T) {
	var tags = TagSet{Deprecated, PnPEnabled, Deprecated, OneCloud}

	testhelper.AssertEqual(t, "before remove", TagSet{Deprecated, PnPEnabled, Deprecated, OneCloud}, tags)

	tags.RemoveTag(Deprecated)
	testhelper.AssertEqual(t, "after 1st remove", TagSet{PnPEnabled, OneCloud}, tags)

	tags.RemoveTag(Deprecated)
	testhelper.AssertEqual(t, "after 2nd remove", TagSet{PnPEnabled, OneCloud}, tags)
}

func TestContains(t *testing.T) {
	var tags = TagSet{NotReady, TypeOtherOSS, OneCloud + ">490102", OneCloudWave1 + ">000103", PnPEnabledIaaS}

	testhelper.AssertEqual(t, string(TypeOtherOSS), true, tags.Contains(TypeOtherOSS))
	testhelper.AssertEqual(t, string(NotReady), true, tags.Contains(NotReady))
	testhelper.AssertEqual(t, string(IaaSGen2), false, tags.Contains(IaaSGen2))
	testhelper.AssertEqual(t, string(OneCloud+">490102"), true, tags.Contains(OneCloud))
	testhelper.AssertEqual(t, string(OneCloudWave1+">000103"), false, tags.Contains(OneCloudWave1))     // expired
	testhelper.AssertEqual(t, string(PnPEnabled)+"(check substring)", false, tags.Contains(PnPEnabled)) // testing with tags that have a common prefix unrelated to expiration date
}

func TestSetOverallStatus(t *testing.T) {
	var tags = TagSet{NotReady, OneCloud + ">490102", TypeOtherOSS}

	tags.SetOverallStatus(StatusYellow)
	testhelper.AssertEqual(t, "first set to yellow", TagSet{StatusYellow, NotReady, OneCloud + ">490102", TypeOtherOSS}, tags)

	tags.SetOverallStatus(StatusRed)
	testhelper.AssertEqual(t, "second set to red", TagSet{StatusRed, NotReady, OneCloud + ">490102", TypeOtherOSS}, tags)
}

func TestGetOverallStatus(t *testing.T) {
	var tags1 = TagSet{StatusGreen, StatusCRNYellow, NotReady, OneCloud + ">490102", TypeOtherOSS}
	testhelper.AssertEqual(t, "green", StatusGreen, tags1.GetOverallStatus())

	var tags2 = TagSet{NotReady, StatusCRNYellow, OneCloud + ">490102", TypeOtherOSS}
	testhelper.AssertEqual(t, "no status", Tag(""), tags2.GetOverallStatus())
}

func TestSetCRNStatus(t *testing.T) {
	var tags = TagSet{NotReady, OneCloud + ">490102", TypeOtherOSS}

	tags.SetCRNStatus(StatusCRNYellow)
	testhelper.AssertEqual(t, "first set to yellow", TagSet{StatusCRNYellow, NotReady, OneCloud + ">490102", TypeOtherOSS}, tags)

	tags.SetCRNStatus(StatusCRNRed)
	testhelper.AssertEqual(t, "second set to red", TagSet{StatusCRNRed, NotReady, OneCloud + ">490102", TypeOtherOSS}, tags)
}

func TestGetCRNStatus(t *testing.T) {
	var tags1 = TagSet{StatusCRNGreen, StatusYellow, NotReady, OneCloud + ">490102", TypeOtherOSS}
	testhelper.AssertEqual(t, "green", StatusCRNGreen, tags1.GetCRNStatus())

	var tags2 = TagSet{NotReady, StatusYellow, OneCloud + ">490102", TypeOtherOSS}
	testhelper.AssertEqual(t, "no status", Tag(""), tags2.GetCRNStatus())
}
func TestString(t *testing.T) {
	var tags1 = TagSet{}
	var tags2 = TagSet{NotReady, OneCloud + ">490102", TypeOtherOSS}

	testhelper.AssertEqual(t, "empty", "[]", tags1.String())
	testhelper.AssertEqual(t, "not empty", "[not_ready onecloud>490102 type_otheross]", tags2.String())
}

func TestWithoutStatus(t *testing.T) {
	ts := TagSet{OneCloudWave1, StatusRed, PnPEnabled, StatusCRNGreen, NotReady, OneCloud + ">490102"}
	tsw := ts.WithoutStatus()
	testhelper.AssertEqual(t, "", "[not_ready onecloud>490102 onecloud_wave1]", tsw.String())
}

func TestParseTagName(t *testing.T) {
	f := func(name string, expected Tag, allowStatusTags bool) {
		var testName string
		if allowStatusTags {
			testName = name
		} else {
			testName = testName + " (allowStatusTags=false)"
		}
		_, parsed, err := parseTagName(name, allowStatusTags)
		if expected != "" {
			if err != nil {
				t.Errorf(`"%s" -> expected no error but got err="%v" and tag="%s"`, testName, err, parsed.StringStatus())
			}
			testhelper.AssertEqual(t, testName+" -> tag", expected.StringStatus(), parsed.StringStatus())
		} else {
			if err == nil {
				t.Errorf(`invalid "%s" -> expected an error but got err==nil and tag="%s"`, testName, parsed.StringStatus())
			}
		}
	}

	if *testhelper.VeryVerbose {
		for n, t := range tagNameMap {
			fmt.Printf("Valid tag entry: \"%s\" -> %v\n", n, t)
		}
	}

	f("OneCloudWave1", OneCloudWave1, true)
	f("one-cloud-wave-1", OneCloudWave1, true)
	f("OneCloudGroup 1", "", true)

	f("StatusRed", StatusRed, true)
	f("OSSStatusRed", StatusRed, true)
	f("Overall:Red", StatusRed, true)
	f("oss-status-red", StatusRed, true)
	f("status-red", StatusRed, true)
	f("Red", StatusRed, true)
	f("CRNGreen", StatusCRNGreen, true)
	f("CRN:Green", StatusCRNGreen, true)

	f("StatusRed", "", false)
	f("OSSStatusRed", "", false)
	f("oss-status-red", "", false)
	f("status-red", "", false)
	f("Red", "", false)
	f("CRNGreen (allowStatusTags=false)", "", false)

	f("NotReady>490102", NotReady+">490102", true)
	f("NotReady>000102", NotReady+">000102", true)
	f("NotReady>49100", "", true)
	f("NotReady>", "", true)
}

func TestValidate(t *testing.T) {
	f := func(name string, expectedResult string, expectedError string, tags ...Tag) {
		ts := TagSet(tags)
		err := ts.Validate(true)
		if expectedError != "" {
			testhelper.AssertEqual(t, name+".error", expectedError, err.Error())
		} else {
			if err != nil {
				t.Errorf("%s/%s test expected no error but got: %v", t.Name(), name, err)
			}
			testhelper.AssertEqual(t, name+".result", expectedResult, fmt.Sprintf("%v", ts))
		}
	}

	f("valid combo", "[retired type_component]", "", "typecomponent", Retired)
	f("invalid tag", "", "Invalid TagSet [retired XYZ type_component]: \n  Invalid tag \"XYZ\": Invalid tag name: \"XYZ\"\\n", Retired, "XYZ", TypeComponent)
	f("invalid combo", "", "Invalid TagSet [retired type_component deprecated]: \n  Tag \"deprecated\" cannot be used in conjunction with Tag \"retired\"\\n", Retired, TypeComponent, Deprecated)
	f("valid expiration date", "[not_ready>490102 onecloud type_component]", "", "oNeCloud", "not-ready>490102", TypeComponent)
	f("valid expiration date (expired)", "[not_ready>000102 onecloud type_component]", "", OneCloud, NotReady+">000102", TypeComponent)
	f("invalid expiration date", "", "Invalid TagSet [onecloud not_ready>49010 type_component]: \n  Invalid tag \"not_ready>49010\": Invalid expiration date format in tag: \"not_ready>49010\": parsing time \"49010\" as \"060102\": cannot parse \"0\" as \"02\"\\n", OneCloud, NotReady+">49010", TypeComponent)
	f("invalid expiration date format", "", "Invalid TagSet [onecloud not_ready<490102 type_component]: \n  Invalid tag \"not_ready<490102\": Invalid tag name: \"not_ready<490102\"\\n", OneCloud, NotReady+"<490102", TypeComponent) // using < instead of >
}

func TestGetExpiredTags(t *testing.T) {
	var tags = TagSet{OneCloud + ">490103", NotReady + ">000102", TypeOtherOSS}

	expired := tags.GetExpiredTags()
	testhelper.AssertEqual(t, "", "[not_ready>000102]", fmt.Sprintf("%v", expired))
}

func TestGetTagByGroup(t *testing.T) {
	var tags = TagSet{OneCloud + ">490103", NotReady + ">000102", TypeOtherOSS}

	testhelper.AssertEqual(t, "valid tag", TypeOtherOSS, tags.GetTagByGroup(GroupEntryType))
	testhelper.AssertEqual(t, "valid tag with expiration", Tag(OneCloud+">490103"), tags.GetTagByGroup(GroupOneCloud))
	testhelper.AssertEqual(t, "tag not found", Tag(""), tags.GetTagByGroup(GroupClientFacing))
}

func TestCheckOSSTestTag(t *testing.T) {
	originalDisplayName := "My Display Name"
	var displayName string
	var prefix string
	var tags TagSet
	var result bool

	displayName = originalDisplayName + ""
	tags = TagSet{}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, "not a test record -> result", false, result)
	testhelper.AssertEqual(t, "not a test record -> name", originalDisplayName, displayName)
	testhelper.AssertEqual(t, "not a test record -> tags", TagSet{}, tags)

	displayName = originalDisplayName + ""
	tags = TagSet{OSSTest}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, "tags only -> result", true, result)
	testhelper.AssertEqual(t, "tags only -> name", ossTestDisplayName+" "+originalDisplayName, displayName)
	testhelper.AssertEqual(t, "tags only -> tags", TagSet{OSSTest}, tags)

	prefix = ossTestDisplayName
	displayName = prefix + " " + originalDisplayName
	tags = TagSet{}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> result`, true, result)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> name`, ossTestDisplayName+" "+originalDisplayName, displayName)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> tags`, TagSet{OSSTest}, tags)

	prefix = ossTestDisplayName
	displayName = prefix + " " + originalDisplayName
	tags = TagSet{OSSTest}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, `both tags and prefix "`+prefix+`" -> result`, true, result)
	testhelper.AssertEqual(t, `both tags and prefix "`+prefix+`" -> name`, ossTestDisplayName+" "+originalDisplayName, displayName)
	testhelper.AssertEqual(t, `both tags and prefix "`+prefix+`" -> tags`, TagSet{OSSTest}, tags)

	testhelper.AssertEqual(t, "verify originalDisplayName", "My Display Name", originalDisplayName)
}

func TestCheckOSSTestTagAltName(t *testing.T) {
	originalDisplayName := "My Display Name"
	var displayName string
	var prefix string
	var tags TagSet
	var result bool

	prefix = "*TEST RECORD*"
	displayName = prefix + " " + originalDisplayName
	tags = TagSet{}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> result`, true, result)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> name`, ossTestDisplayName+" "+originalDisplayName, displayName)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> tags`, TagSet{OSSTest}, tags)

	prefix = "*Test Record*"
	displayName = prefix + " " + originalDisplayName
	tags = TagSet{OSSTest}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, `both tags and prefix "`+prefix+`" -> result`, true, result)
	testhelper.AssertEqual(t, `both tags and prefix "`+prefix+`" -> name`, ossTestDisplayName+" "+originalDisplayName, displayName)
	testhelper.AssertEqual(t, `both tags and prefix "`+prefix+`" -> tags`, TagSet{OSSTest}, tags)

	prefix = "test-record:"
	displayName = prefix + " " + originalDisplayName
	tags = TagSet{}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> result`, true, result)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> name`, ossTestDisplayName+" "+originalDisplayName, displayName)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> tags`, TagSet{OSSTest}, tags)

	prefix = "TEST___record"
	displayName = prefix + " " + originalDisplayName
	tags = TagSet{}
	result = CheckOSSTestTag(&displayName, &tags)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> result`, true, result)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> name`, ossTestDisplayName+" "+originalDisplayName, displayName)
	testhelper.AssertEqual(t, `prefix "`+prefix+`" only -> tags`, TagSet{OSSTest}, tags)

	testhelper.AssertEqual(t, "verify originalDisplayName", "My Display Name", originalDisplayName)
}

func TestRMCManaged(t *testing.T) {
	testhelper.AssertEqual(t, "Deprecated(RMC)", true, Deprecated.IsRMCManaged())
	testhelper.AssertEqual(t, "LenientCRNName(not-RMC)", false, LenientCRNName.IsRMCManaged())
}

func TestIStatusTag(t *testing.T) {
	testhelper.AssertEqual(t, "Deprecated", false, Deprecated.IsStatusTag())
	testhelper.AssertEqual(t, "StatusRed", true, StatusRed.IsStatusTag())
	testhelper.AssertEqual(t, "StatusCRNGreen", true, StatusCRNGreen.IsStatusTag())
	testhelper.AssertEqual(t, "PnPEnabled", true, PnPEnabled.IsStatusTag())
	testhelper.AssertEqual(t, "PnPEnabledIaaS", false, PnPEnabledIaaS.IsStatusTag())
}

func TestCompareTagSet(t *testing.T) {
	debug.SetDebugFlags(0 /* | debug.Compare /* XXX */)

	var tags1 = &TagSet{StatusGreen, StatusCRNYellow, NotReady, OneCloud + ">490102", TypeOtherOSS}
	tags1.Validate(true)
	var tags2 = &TagSet{StatusCRNYellow, OneCloud + ">490102", TypeOtherOSS}
	tags2.Validate(true)

	/* XXX Disabled: We want to use the string slice comparator to see exact differences
	var expected = []string{`DIFF VALUE:   left="[not_ready onecloud>490102 oss_status_crn_yellow oss_status_green type_otheross]"    right="[onecloud>490102 oss_status_crn_yellow type_otheross]"`}
	*/
	var expected = []string{
		`EQUAL:           left[1]="onecloud>490102"`,
		`EQUAL:           left[2]="oss_status_crn_yellow"`,
		`EQUAL:           left[4]="type_otheross"`,
		`DIFF LEFT ONLY:  left[0]="not_ready"`,
		`DIFF LEFT ONLY:  left[3]="oss_status_green"`,
	}

	out := compare.Output{IncludeEqual: true}
	compare.DeepCompare("left", *tags1, "right", *tags2, &out)
	diffs := out.ToStrings()
	if !reflect.DeepEqual(diffs, expected) {
		t.Errorf("%s: got %v  expected %v", t.Name(), diffs, expected)
	}
}

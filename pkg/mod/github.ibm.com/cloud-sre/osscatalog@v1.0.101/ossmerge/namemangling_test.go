package ossmerge

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestMakeCanonicalName(t *testing.T) {
	t.Run("already canonical", func(t *testing.T) {
		result := MakeCanonicalName("some-service-name")
		testhelper.AssertEqual(t, "", result, ossrecord.CRNServiceName("some-service-name"))
	})

	t.Run("not canonical", func(t *testing.T) {
		result := MakeCanonicalName("Some Service Name")
		testhelper.AssertEqual(t, "", result, ossrecord.CRNServiceName("some-service-name"))
	})

	t.Run("prefix+canonical", func(t *testing.T) {
		result := MakeCanonicalName("is.some-service-name")
		testhelper.AssertEqual(t, "", result, ossrecord.CRNServiceName("is-some-service-name"))
	})

	t.Run("prefix+not canonical", func(t *testing.T) {
		result := MakeCanonicalName("is.Some Service Name")
		testhelper.AssertEqual(t, "", result, ossrecord.CRNServiceName("is-some-service-name"))
	})

	t.Run("bad prefix", func(t *testing.T) {
		result := MakeCanonicalName("iss.some-service-name")
		testhelper.AssertEqual(t, "", result, ossrecord.CRNServiceName("iss-some-service-name"))
	})

	t.Run("empty", func(t *testing.T) {
		result := MakeCanonicalName("")
		testhelper.AssertEqual(t, "", result, ossrecord.CRNServiceName(""))
	})
}

func TestMakeComparableName(t *testing.T) {
	t.Run("already flat", func(t *testing.T) {
		result := MakeComparableName("someservicename")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("canonical", func(t *testing.T) {
		result := MakeComparableName("some-service-name")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("spaces and upper", func(t *testing.T) {
		result := MakeComparableName("Some Service Name")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("IBM", func(t *testing.T) {
		result := MakeComparableName("ibm-some-service-name")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("ibmcloud", func(t *testing.T) {
		result := MakeComparableName("ibmcloud-some-service-name")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("ibm-cloud", func(t *testing.T) {
		result := MakeComparableName("ibm-cloud-some-service-name")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("bluemix", func(t *testing.T) {
		result := MakeComparableName("bluemix-some-service-name")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("ibm-bluemix", func(t *testing.T) {
		result := MakeComparableName("ibm-bluemix-some-service-name")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("is prefix", func(t *testing.T) {
		result := MakeComparableName("is.some-service-name")
		testhelper.AssertEqual(t, "", result, "issomeservicename")
	})

	t.Run("empty", func(t *testing.T) {
		result := MakeComparableName("")
		testhelper.AssertEqual(t, "", result, "")
	})

	t.Run("suffix for ibm cloud", func(t *testing.T) {
		result := MakeComparableName("some-service-name-for-ibm-cloud")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("prefix+suffix", func(t *testing.T) {
		result := MakeComparableName("ibm-some-service-name-for-ibm-cloud")
		testhelper.AssertEqual(t, "", result, "someservicename")
	})

	t.Run("prefix+suffix+replacements", func(t *testing.T) {
		result := MakeComparableName("testxyz some-service-name testzyx")
		testhelper.AssertEqual(t, "", result, "xxyyzzsomeservicenamezzyyxx")
	})
}

func TestParseCompositeName(t *testing.T) {
	f := func(name string, expectedBase string, expectedSuffix string) {
		base, suffix := ParseCompositeName(name)
		testhelper.AssertEqual(t, name+"(base)", expectedBase, base)
		testhelper.AssertEqual(t, name+"(suffix)", expectedSuffix, suffix)
	}

	f("is.volume", "is", "volume")
	f("is", "", "")
	f("is.floating-ip", "is", "floating-ip")
	f("is.foo.bar", "", "")
	f("IS.bar", "", "")
	f("is.BAR", "", "")
	f("is-bar", "", "")
	f("is-BAR", "", "")
	f("aa-bb.cc", "", "")
}

func TestConvertCompositeToCanonicalName(t *testing.T) {
	f := func(name string, expectedCanonical string, expectedComposite bool) {
		canonical, isComposite := ConvertCompositeToCanonicalName(name)
		testhelper.AssertEqual(t, name+"(canonical)", ossrecord.CRNServiceName(expectedCanonical), canonical)
		testhelper.AssertEqual(t, name+"(isComposite)", expectedComposite, isComposite)
	}

	f("is.volume", "is-volume", true)
	f("is", "is", false)
	f("is.floating-ip", "is-floating-ip", true)
	f("is.foo.bar", "is-foo-bar", false)
	f("IS.bar", "is-bar", false)
	f("is.BAR", "is-bar", false)
	f("is-bar", "is-bar", false)
	f("is-BAR", "is-bar", false)
	f("aa-bb.cc", "aa-bb-cc", false)
}

func TestCompareCompositeAndCanonicalName(t *testing.T) {
	f := func(name string, canonical string, expected bool) {
		result := CompareCompositeAndCanonicalName(name, ossrecord.CRNServiceName(canonical))
		testhelper.AssertEqual(t, name+"/"+canonical, expected, result)
	}

	f("is.volume", "is-volume", true)
	f("is", "is", true)
	f("is.floating-ip", "is-floating-ip", true)
	f("is.foo.bar", "is-foo-bar", false)
	f("IS.bar", "is-bar", false)
	f("is.BAR", "is-bar", false)
	f("is-bar", "is-bar", true)
	f("is-BAR", "is-bar", false)
	f("aa-bb.cc", "aa-bb-cc", false)
}

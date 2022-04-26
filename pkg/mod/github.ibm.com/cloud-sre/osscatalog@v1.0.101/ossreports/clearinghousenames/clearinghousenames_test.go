package clearinghousenames

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestParseNameList(t *testing.T) {

	var f = func(input string, expected ...string) {
		if expected == nil {
			expected = []string{}
		}
		m := parseNameList(input)
		testhelper.AssertEqual(t, input, expected, m)
	}

	f(``)
	f(`[]`)
	f(`["foo"]`, "foo")
	f(`["foo","bar"]`, "foo", "bar")
	f(`["foo","bar","baz"]`, "foo", "bar", "baz")

	f(`  [  "  foo  "  ,  "  bar  "  ]  `, "  foo  ", "  bar  ")
}

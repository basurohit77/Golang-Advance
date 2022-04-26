package ossuid

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestParseAndString(t *testing.T) {
	var p UID
	var s string

	p = Parse("OSS0002")
	testhelper.AssertEqual(t, "Parse(OSS0002)", 2, p.numeric)
	s = p.String()
	testhelper.AssertEqual(t, "String(OSS0002)", "OSS0002", s)

	p = Parse("OSS00A2")
	testhelper.AssertEqual(t, "Parse(OSS00A2)", 362, p.numeric)
	s = p.String()
	testhelper.AssertEqual(t, "String(OSS00A2)", "OSS00A2", s)

	p = Parse("OSS0A00")
	testhelper.AssertEqual(t, "Parse(OSS0A00)", 12960, p.numeric)
	s = p.String()
	testhelper.AssertEqual(t, "String(OSS0A00)", "OSS0A00", s)
}

func TestIncrement(t *testing.T) {
	uid1 := Parse("OSS0A00")
	uid2 := uid1.Increment()
	testhelper.AssertEqual(t, "Increment(OSS0A00)", "OSS0A01", uid2.String())
}

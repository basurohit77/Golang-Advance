package ossrecord

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestAddDependency(t *testing.T) {
	deps := new(Dependencies)

	deps.AddDependency("serviceB", "tag2")
	deps.AddDependency("serviceA", "tag1")
	deps.AddDependency("serviceB", "tag2", "tag3", "tag1")
	deps.AddDependency("serviceA")
	deps.AddDependency("serviceA", "tag1")

	testhelper.AssertEqual(t, "dependency0.Service", "serviceA", (*deps)[0].Service)
	testhelper.AssertEqual(t, "dependency0.Tags", []string{"tag1"}, (*deps)[0].Tags)
	testhelper.AssertEqual(t, "dependency1.Service", "serviceB", (*deps)[1].Service)
	testhelper.AssertEqual(t, "dependency1.Tags", []string{"tag1", "tag2", "tag3"}, (*deps)[1].Tags)
}

func TestCountTag(t *testing.T) {

	deps := new(Dependencies)

	deps.AddDependency("serviceB", "tag1:val1.1", "tag2")
	deps.AddDependency("serviceA", "tag1:val1.2", "tag2:val2.1", "tag3")
	deps.AddDependency("serviceC", "tag3")

	testhelper.AssertEqual(t, "all with value", 2, deps.CountTag("tag1:"))
	testhelper.AssertEqual(t, "all with value no colon", 0, deps.CountTag("tag1"))
	testhelper.AssertEqual(t, "specific value", 1, deps.CountTag("tag1:val1.2"))
	testhelper.AssertEqual(t, "mixed value", 1, deps.CountTag("tag2:"))
	testhelper.AssertEqual(t, "mixed value no colon", 1, deps.CountTag("tag2"))
	testhelper.AssertEqual(t, "none with value", 2, deps.CountTag("tag3"))
	testhelper.AssertEqual(t, "nonexistent", 0, deps.CountTag("tag4"))
}

func TestFindTag(t *testing.T) {
	deps := new(Dependencies)

	deps.AddDependency("serviceB", "tag1:val1.1", "tag2")

	testhelper.AssertEqual(t, "with value", "val1.1", (*deps)[0].FindTag("tag1:"))
	testhelper.AssertEqual(t, "specific value", "tag1:val1.1", (*deps)[0].FindTag("tag1:val1.1"))
	testhelper.AssertEqual(t, "with value no colon", "", (*deps)[0].FindTag("tag1"))
	testhelper.AssertEqual(t, "no value", "tag2", (*deps)[0].FindTag("tag2"))
	testhelper.AssertEqual(t, "no value with colon", "", (*deps)[0].FindTag("tag2:"))
	testhelper.AssertEqual(t, "nonexistent", "", (*deps)[0].FindTag("tag3"))

}

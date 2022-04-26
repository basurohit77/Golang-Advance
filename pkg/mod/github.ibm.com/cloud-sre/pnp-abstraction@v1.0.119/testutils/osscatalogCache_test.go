package testutils

import (
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

func TestOSS(t *testing.T) {
	err := MyTestListFunction(regexp.MustCompile(".*"), catalog.IncludeServices, testCatcher)
	if err != nil {
		t.Fatal(err)
	}
}

func testCatcher(r ossrecord.OSSEntry) {
}

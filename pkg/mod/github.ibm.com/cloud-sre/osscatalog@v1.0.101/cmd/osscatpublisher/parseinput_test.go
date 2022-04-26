package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestParseJSONInput(t *testing.T) {
	/* *testhelper.VeryVerbose = true /* XXX */

	options.LoadGlobalOptions("-keyfile <none>", true)

	filename := "testdata/input.json"
	inputFile, err := os.Open(filename)
	testhelper.AssertError(t, err)
	pattern := regexp.MustCompile(".*")

	numEntries, err := parseJSONInput(inputFile, pattern, func(r ossrecord.OSSEntry) {
		if *testhelper.VeryVerbose {
			fmt.Printf("--> got entry: %s\n", r.String())
		}
	})
	testhelper.AssertError(t, err)

	testhelper.AssertEqual(t, "Number of records found", 6, numEntries)
}

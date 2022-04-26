package testutils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

const (
	catalogName = "catalogOss_test.json"
)

// MyTestListFunction simulates the OSSListEntiries function in the OSS catalog library, but uses UT data.
func MyTestListFunction(pattern *regexp.Regexp, incl catalog.IncludeOptions, pFunc func(r ossrecord.OSSEntry)) error {
	_, filename, _, _ := runtime.Caller(0) // Find the path of this source file so we can locate the data file
	path := filename[:strings.LastIndex(filename, "/")]
	filename = path + "/testdata/" + catalogName
	return baseListFunction(filename, pattern, incl, pFunc)
}

// [2019-06-20] The structure of the osscatalog changed. Here we compensate by reintroducing the 'OSS' field, then
// we map it to the new OSSService type.
type oldOssRecExt struct {
	OSS ossrecord.OSSService
}

func baseListFunction(fn string, pattern *regexp.Regexp, incl catalog.IncludeOptions, pFunc func(r ossrecord.OSSEntry)) error {
	fnCleaned := filepath.Clean(fn)
	data, err := ioutil.ReadFile(fnCleaned)

	ossRecords := make([]*oldOssRecExt, 0)
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&ossRecords); err != nil {
		log.Println("baseListFunction: Error parsing OSS records from UT file ", fnCleaned, " error=", err)
		return err
	}

	for _, v := range ossRecords {
		// OSS Catalog had a bug where the json name of a field was misspelled. That later got fixed, but introduced a new
		// field in the record.  The below compensates for that.
		if v.OSS.StatusPage.CategoryID == "" {
			v.OSS.StatusPage.CategoryID = v.OSS.StatusPage.CategoryIDMisspelled
		}
		v.OSS.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled) // add pnpenabled flag so category cache will use this entry
		pFunc(&v.OSS)
	}

	log.Println("Completed listing function")
	return err
}

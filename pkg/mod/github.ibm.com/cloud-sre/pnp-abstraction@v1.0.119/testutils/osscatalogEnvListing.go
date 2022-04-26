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
)

const (
	envDataFile = "test_environments.json"
)

// MyTestEnvironmentsListFunction simulates the OSSListEnvironments function in the OSS catalog library, but uses UT data.
func MyTestEnvironmentsListFunction(pattern *regexp.Regexp, incl catalog.IncludeOptions, pFunc func(r ossrecord.OSSEntry)) error {
	_, filename, _, _ := runtime.Caller(0) // Find the path of this source file so we can locate the data file
	path := filename[:strings.LastIndex(filename, "/")]
	filename = path + "/testdata/" + envDataFile

	filenameCleaned := filepath.Clean(filename)
	data, err := ioutil.ReadFile(filenameCleaned)

	ossRecords := make([]ossrecord.OSSEnvironment, 0)
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&ossRecords); err != nil {
		log.Println("MyTestEnvironmentListFunction: Error parsing OSS records from UT file ", filenameCleaned, " error=", err)
		return err
	}

	for i := range ossRecords {
		//To correct G601 (CWE-118): Implicit memory aliasing in for loop scan error
		pFunc(&ossRecords[i])
	}

	log.Println("Completed listing function")
	return err
}

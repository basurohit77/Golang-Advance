package options

import (
	"flag"
	"sort"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadConfigFileOK(t *testing.T) {

	if *testhelper.VeryVerbose {
		debug.SetDebugFlags(debug.Options)
	}

	saved := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("TestReadConfigFileOK", flag.ExitOnError)
	defer func() { flag.CommandLine = saved }()

	keyfileFlag := flag.String("keyfile", "", "")
	lenientFlag := flag.Bool("lenient", false, "")
	option1Flag := flag.String("option1", "", "")
	option2Flag := flag.Bool("option2", false, "")

	err := ReadConfigFile("testdata/test-config.yaml")

	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "keyfile", "default", *keyfileFlag)
	testhelper.AssertEqual(t, "lenient", true, *lenientFlag)
	testhelper.AssertEqual(t, "option1", "option1Value", *option1Flag)
	testhelper.AssertEqual(t, "option2", true, *option2Flag)
}

func TestReadConfigFileWithErrors(t *testing.T) {

	if *testhelper.VeryVerbose {
		debug.SetDebugFlags(debug.Options)
	}

	saved := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("TestReadConfigFileWithErrors", flag.ExitOnError)
	defer func() { flag.CommandLine = saved }()

	keyfileFlag := flag.String("keyfile", "", "")
	f := flag.Lookup("keyfile")
	f.Value.Set("not-default")
	lenientFlag := flag.Bool("lenient", false, "")
	//	option1Flag := flag.String("option1", "", "")
	option2Flag := flag.String("option2", "", "")

	err := ReadConfigFile("testdata/test-config.yaml")

	errors := strings.Split(err.Error(), `\n`)
	sort.Strings(errors)
	//fmt.Printf("ERRORS: %#v\n", errors)
	expectedErrors := []string{
		"",
		"Empty parameter value in config file: \"option2\"=\"\" does not appear to be for a boolean flag (actual type: string)",
		"Errors encountered while processing parameters config file testdata/test-config.yaml: ",
		"Parameter in config file \"keyfile\"=\"default\" is already set to a non-default value \"not-default\" (from command line?)",
		"Unknown parameter in config file: \"option1\"=\"option1Value\""}

	testhelper.AssertEqual(t, "errors", expectedErrors, errors)
	testhelper.AssertEqual(t, "keyfile", "not-default", *keyfileFlag)
	testhelper.AssertEqual(t, "lenient", true, *lenientFlag)
	//	testhelper.AssertEqual(t, "option1", "option1Value", *option1Flag)
	testhelper.AssertEqual(t, "option2", "", *option2Flag)
}

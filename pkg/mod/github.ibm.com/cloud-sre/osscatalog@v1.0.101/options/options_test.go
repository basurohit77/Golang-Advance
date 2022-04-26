package options

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestLoadGlobalOptions(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestLoadGlobalOptions in short mode")
	}

	//	fmt.Println("DEBUG first call - VCAP=", os.Getenv("VCAP_SERVICES"))
	LoadGlobalOptions("-verbose -keyfile default", true)
	testhelper.AssertEqual(t, "-verbose - first call with parameter", true, GlobalOptions().Verbose)
	testhelper.AssertEqual(t, "-lenient - first call with parameter", false, GlobalOptions().Lenient)

	//	fmt.Println("DEBUG second call - VCAP=", os.Getenv("VCAP_SERVICES"))
	LoadGlobalOptions("-lenient -keyfile default", true)
	testhelper.AssertEqual(t, "-verbose - second call with parameter", false, GlobalOptions().Verbose)
	testhelper.AssertEqual(t, "-lenient - second call with parameter", true, GlobalOptions().Lenient)

	// TODO: Need testing in CF environment
	/*
		saved, exists := os.LookupEnv("VCAP_SERVICES")
		os.Setenv("VCAP_SERVICES", "dummy")
		if exists {
			defer os.Setenv("VCAP_SERVICES", saved)
		} else {
			defer os.Unsetenv("VCAP_SERVICES")
		}

		//	fmt.Println("DEBUG third call - VCAP=", os.Getenv("VCAP_SERVICES"))
		LoadGlobalOptions("", true)
		testhelper.AssertEqual(t, "-verbose - third call with command line", false, GlobalOptions().Verbose)
		testhelper.AssertEqual(t, "-lenient - third call with command line", false, GlobalOptions().Lenient)

		//	fmt.Println("DEBUG fourth call - VCAP=", os.Getenv("VCAP_SERVICES"))
		LoadGlobalOptions("", true)
		testhelper.AssertEqual(t, "-verbose - fourth call with command line", false, GlobalOptions().Verbose)
		testhelper.AssertEqual(t, "-lenient - fourth call with command line", false, GlobalOptions().Lenient)
	*/
}

func TestDefaultGlobalOptions(t *testing.T) {
	// Ensure there are default options:
	if GlobalOptions() == nil {
		t.Logf("Default global options are nil. Expected to be non-nil.")
		t.Fail()
	}

	// Ensure that if LoadGlobalOptions is called the default options are not returned:
	LoadGlobalOptions("-visibility ibm_only -keyfile <none>", true)
	testhelper.AssertEqual(t, "GlobalOptions().rawVisibilityRestrictions", "ibm_only", GlobalOptions().rawVisibilityRestrictions)
}

func TestVcapFile(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestVcapFile in short mode")
	}
	LoadGlobalOptions("-vcap testdata/test-vcap.json", true)

	key, err := rest.GetKey("service2")
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "key(service2)", "AAA", key)

	val := options.CFServices["ServiceType2"][0].Credentials["token"]
	testhelper.AssertEqual(t, "token(service3)", "DDD", val)

	name2 := options.CFServices["user-provided"][0].Name
	testhelper.AssertEqual(t, "service1.Name", "service1", name2)
	val2 := options.CFServices["user-provided"][0].Credentials["clientId"]
	testhelper.AssertEqual(t, "clientId(service1)", "YYY", val2)

	if *testhelper.VeryVerbose {
		var out strings.Builder
		out.WriteString("Parsed VCAP data:\n")
		json, err := json.MarshalIndent(options.CFServices, "    ", "    ")
		testhelper.AssertError(t, err)
		out.Write(json)
		fmt.Println(out.String())
	}
}

func TestIsCheckOwnerSpecified(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test TestIsCheckOwnerSpecified() in short mode")
	}

	var ret bool

	// XXX Note this only works if we check without the flag first, because we cannot reset the flag.Visit method
	LoadGlobalOptions("-keyfile default -lenient", true)
	ret = IsCheckOwnerSpecified()
	testhelper.AssertEqual(t, "without flag", false, ret)

	LoadGlobalOptions("-keyfile default -check-owner", true)
	ret = IsCheckOwnerSpecified()
	testhelper.AssertEqual(t, "with flag", true, ret)
}

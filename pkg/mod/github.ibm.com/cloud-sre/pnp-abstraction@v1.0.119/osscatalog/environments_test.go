package osscatalog

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/pnp-abstraction/testutils"
)

func TestEnvCheck(t *testing.T) {

	globalEnvironmentListingFunc = testutils.MyTestEnvironmentsListFunction

	crn := "crn:v1:bluemix:public::eu-de::::"
	ok := IsEnvPnPEnabled(crn)
	if !ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:bluemix:public::dal02-b::::" // BAD location
	ok = IsEnvPnPEnabled(crn)
	if !ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:bluemix:public::ch-ctu::::" //ZONE
	ok = IsEnvPnPEnabled(crn)
	if !ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:d-aa1:dedicated::us-east::::" //DEDICATED
	ok = IsEnvPnPEnabled(crn)
	if ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:bluemix:dedicated::eu-de:::" // Public cname (bluemix) but private ctype (dedicated)
	ok = IsEnvPnPEnabled(crn)
	if ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:d-aa1:public::eu-de:::" // Public ctype (public) but private cname (d-aa1)
	ok = IsEnvPnPEnabled(crn)
	if ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:bluemix:public:myservice:eu-de::::" // we deal with service names
	ok = IsEnvPnPEnabled(crn)
	if !ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:bluemix:public::eu-de:::" // BAD CRN Format
	ok = IsEnvPnPEnabled(crn)
	if ok {
		t.Error("CRN failed ", crn)
	}

	crn = "crn:v1:bluemix:public:myservice:eu-de:a/foo:::" // has account
	ok = IsEnvPnPEnabled(crn)
	if !ok {
		t.Error("CRN failed ", crn)
	}

}

func TestEnvList(t *testing.T) {

	list, err := GetEnvironments()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Number of environments:", len(list))
	if len(list) < 20 {
		t.Error("bad returned list")
	}

	for _, e := range list {
		catID := e.ReferenceCatalogID
		if catID == "" {
			catID = "     "
		}
		fmt.Printf("%s -> %s -> %s\n", catID, e.EnvironmentID, e.Type)

	}

	list, err = GetCloudServiceEnvironments()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Cloud Environments")

	for _, e := range list {
		catID := e.ReferenceCatalogID
		if catID == "" {
			catID = "     "
		}
		fmt.Printf("%s -> %s -> %s\n", catID, e.EnvironmentID, e.Type)

	}

}

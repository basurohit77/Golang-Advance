package ossmerge

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestIMSIDRegex(t *testing.T) {
	m := imsIDRegex.FindStringSubmatch("SoftLayer AMS03 (814994)")
	testhelper.AssertEqual(t, "len(m)", 2, len(m))
	testhelper.AssertEqual(t, "IMS ID", "814994", m[1])
}

func TestDoctorEnvRegex(t *testing.T) {
	m := doctorEnvRegex.FindString("ABN AMRO - decommissioned\n\n    DoctorEnvironment(ABN AMRO - decommissioned/D_ABNAMRO[crn:v1:d-abnamro:dedicated::ams03::::]\n    DoctorRegionID(ABN AMRO - decommissioned[crn:v1:d-abnamro:dedicated::ams03::::]/mccpid=abnamro:prod:eu-nl)\n")
	testhelper.AssertEqual(t, "1", "DoctorEnvironment(ABN AMRO - decommissioned/D_ABNAMRO[crn:v1:d-abnamro:dedicated::ams03::::]", m)

	m = doctorEnvRegex.FindString("AT&T 1 (deployed as WBMDN-28801)\n\n    DoctorEnvironment(AT&T 1 (deployed as WBMDN-28801)/D_WBMDN-28801[crn:v1:d-wbmdn-28801:dedicated::us-south::::]\n    DoctorRegionID(AT&T 1 (deployed as WBMDN-28801)[crn:v1:d-wbmdn-28801:dedicated::us-south::::]/mccpid=wbmdn-28801:prod:us-south)\n")
	testhelper.AssertEqual(t, "2", "DoctorEnvironment(AT&T 1 (deployed as WBMDN-28801)/D_WBMDN-28801[crn:v1:d-wbmdn-28801:dedicated::us-south::::]", m)

	m = doctorEnvRegex.FindString("Accenture - decommissioned\n\n    DoctorRegionID(Accenture - decommissioned[]/mccpid=acn:prod:us-ne)\n")
	testhelper.AssertEqual(t, "3", "", m)

	m = doctorEnvRegex.FindString("Amsterdam 01\n\n    Catalog{Path:\"/eu/nl/ams/ams01\", Kind:\"dc\", ID:\"ams01\"}\n    DoctorRegionID(SoftLayer AMS01 (265592)[crn:v1:softlayer:public::ams01::::]/mccpid=SoftLayer AMS01 (265592))\n")
	testhelper.AssertEqual(t, "4", "", m)
}

func TestDoctorRegionIDRegex(t *testing.T) {
	m := doctorRegionIDRegex.FindString("ABN AMRO - decommissioned\n\n    DoctorEnvironment(ABN AMRO - decommissioned/D_ABNAMRO[crn:v1:d-abnamro:dedicated::ams03::::]\n    DoctorRegionID(ABN AMRO - decommissioned[crn:v1:d-abnamro:dedicated::ams03::::]/mccpid=abnamro:prod:eu-nl)\n")
	testhelper.AssertEqual(t, "1", "DoctorRegionID(ABN AMRO - decommissioned[crn:v1:d-abnamro:dedicated::ams03::::]/mccpid=abnamro:prod:eu-nl)", m)

	m = doctorRegionIDRegex.FindString("AT&T 1 (deployed as WBMDN-28801)\n\n    DoctorEnvironment(AT&T 1 (deployed as WBMDN-28801)/D_WBMDN-28801[crn:v1:d-wbmdn-28801:dedicated::us-south::::]\n    DoctorRegionID(AT&T 1 (deployed as WBMDN-28801)[crn:v1:d-wbmdn-28801:dedicated::us-south::::]/mccpid=wbmdn-28801:prod:us-south)\n")
	testhelper.AssertEqual(t, "2", "DoctorRegionID(AT&T 1 (deployed as WBMDN-28801)[crn:v1:d-wbmdn-28801:dedicated::us-south::::]/mccpid=wbmdn-28801:prod:us-south)", m)

	m = doctorRegionIDRegex.FindString("Accenture - decommissioned\n\n    DoctorRegionID(Accenture - decommissioned[]/mccpid=acn:prod:us-ne)\n")
	testhelper.AssertEqual(t, "3", "DoctorRegionID(Accenture - decommissioned[]/mccpid=acn:prod:us-ne)", m)

	m = doctorRegionIDRegex.FindString("Amsterdam 01\n\n    Catalog{Path:\"/eu/nl/ams/ams01\", Kind:\"dc\", ID:\"ams01\"}\n    DoctorRegionID(SoftLayer AMS01 (265592)[crn:v1:softlayer:public::ams01::::]/mccpid=SoftLayer AMS01 (265592))\n")
	testhelper.AssertEqual(t, "4", "DoctorRegionID(SoftLayer AMS01 (265592)[crn:v1:softlayer:public::ams01::::]/mccpid=SoftLayer AMS01 (265592))", m)
}

package ossrecord

import "github.ibm.com/cloud-sre/osscatalog/osstags"

// Note this is not a test file per se, hence it is not
// named *_test.go. Rather, this is a common utility file
// used to write tests in *other* packages.

// CreateTestRecord creates a dummy OSSService record that can be used in various testing functions
func CreateTestRecord() *OSSService {
	oss := &OSSService{}

	oss.SchemaVersion = OSSCurrentSchema
	oss.ReferenceResourceName = "osscatalog-testing"
	oss.ReferenceDisplayName = "OSS Catalog Testing"

	gi := &oss.GeneralInfo
	gi.RMCNumber = "12345"
	gi.OperationalStatus = BETA
	gi.OSSTags = osstags.TagSet{}
	gi.OSSTags.AddTag(osstags.StatusGreen)
	gi.OSSTags.AddTag(osstags.NotReady)
	gi.ClientFacing = true
	gi.EntryType = SERVICE
	gi.ServiceNowSysid = "ea56778eebed67"
	gi.OSSDescription = "This is a test record for oss-catalog functions"
	gi.ParentResourceName = ""
	gi.Domain = COMMERCIAL

	ow := &oss.Ownership
	ow.OfferingManager = Person{Name: "John Doe", W3ID: "johndoe@us.ibm.com"}
	ow.DevelopmentManager = Person{Name: "Jane Somebody"}
	ow.TechnicalContactDEPRECATED = Person{W3ID: "howard@us.ibm.com"}
	ow.SegmentName = "Watson and Cloud CTO"
	ow.SegmentOwner = Person{Name: "Bryson Koehler"}
	ow.TribeName = "CTO Global Technology Operations"
	ow.TribeOwner = Person{Name: "Shaun Smith"}

	su := &oss.Support
	su.Manager = Person{Name: "Shawn Bramblett"}
	su.ClientExperience = ACSSUPPORTED
	su.SpecialInstructions = "Make sure to collect MustGather"
	su.Tier2EscalationType = GITHUB
	su.Slack = "#osscatalog-support"
	su.Tier2Repository = "https://github.ibm.com/cloud-sre/osscatalog"

	op := &oss.Operations
	op.Manager = Person{Name: "Shawn Bramblett"}
	op.SpecialInstructions = "This service never fails!"
	op.TIPOnboarded = true

	st := &oss.StatusPage
	st.Group = "StatusGroup1"
	st.CategoryID = "blah.blah.literal.l133"

	co := &oss.Compliance
	co.ServiceNowOnboarded = true
	co.BCDRFocal = Person{Name: "Batman"}
	co.ArchitectureFocal = Person{Name: "Robin"}

	oss.AdditionalContacts = "See Bluepages"

	sn := &oss.ServiceNowInfo
	sn.SupportTier1AG = "My Assignment Group"
	sn.SupportTier2AG = "My Other Assignment Group"

	cat := &oss.CatalogInfo
	cat.Provider = Person{Name: "IBM", W3ID: "bshawn@us.ibm.com"}
	cat.ProviderContact = "Shawn Bramblett"
	cat.ProviderSupportEmail = "bshawn@us.ibm.com"
	cat.ProviderPhone = "+1-720-349-602"

	return oss
}

package crn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestParse(t *testing.T) {
	input := "crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"

	crn, err := Parse(input)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "Parse", Mask{"cname", "ctype", "service-name", "location", "scope", "service-instance", "resource-type", "resource"}, crn)

	output := crn.ToCRNString()
	testhelper.AssertEqual(t, "CRNMask.String()", String(input), output)
}

func TestNormalizeCRNMask(t *testing.T) {
	crn0, _ := Parse("crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource")
	normalized0, err0 := crn0.Normalize(ExpectEnvironment, ExpectServiceName, AllowOtherSegments)
	testhelper.AssertEqual(t, "Normalize() OK normal case", nil, err0)
	testhelper.AssertEqual(t, "Normalize() OK normal case", crn0, normalized0)

	crn1, _ := Parse("crn:v1:::service-name:location:scope:service-instance:resource-type:resource")
	normalized1, err1 := crn1.Normalize(ExpectEnvironment, ExpectServiceName, AllowBlankCName, AllowOtherSegments)
	expected1, _ := Parse("crn:v1:bluemix:public:service-name:location:scope:service-instance:resource-type:resource")
	testhelper.AssertEqual(t, "Normalize() OK blank CName", nil, err1)
	testhelper.AssertEqual(t, "Normalize() OK blank CName", expected1, normalized1)

	crn2, _ := Parse("crn:v1:softlayer:public:service-name:location:scope:service-instance:resource-type:resource")
	normalized2, err2 := crn2.Normalize(ExpectEnvironment, ExpectServiceName, AllowSoftLayerCName, AllowOtherSegments)
	expected2, _ := Parse("crn:v1:bluemix:public:service-name:location:scope:service-instance:resource-type:resource")
	testhelper.AssertEqual(t, "Normalize() OK SoftLayer CName", nil, err2)
	testhelper.AssertEqual(t, "Normalize() OK SofLayer CName", expected2, normalized2)

	crn3, _ := Parse("crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource")
	normalized3, err3 := crn3.Normalize(ExpectEnvironmentPublicCloud, ExpectServiceName, AllowOtherSegments)
	testhelper.AssertEqual(t, "Normalize() expected Public Cloud", "Invalid CRN Mask \"crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource\": [\"Invalid Environment information (cname/ctype/location): expected IBM Public Cloud\"]", err3.Error())
	testhelper.AssertEqual(t, "Normalize() expected Public Cloud", crn3, normalized3)

	crn4, _ := Parse("crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource")
	normalized4, err4 := crn4.Normalize(AllowOtherSegments)
	expected4, _ := Parse("crn:v1:::::scope:service-instance:resource-type:resource")
	testhelper.AssertEqual(t, "Validate() various errors",
		"Invalid CRN Mask \"crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource\": ["+
			"\"Unexpected Environment (cname/ctype/location) segments\" "+
			"\"Unexpected \\\"service-name\\\" segment\"]",
		err4.Error())
	testhelper.AssertEqual(t, "Normalize() expected Public Cloud", expected4, normalized4)

	crn5, _ := Parse("crn:v1:bluemix:public:service-name:location:scope:service-instance:resource-type:resource")
	normalized5, err5 := crn5.Normalize(ExpectEnvironment, ExpectServiceName, AllowSoftLayerCName, AllowOtherSegments)
	expected5, _ := Parse("crn:v1:bluemix:public:service-name:location:scope:service-instance:resource-type:resource")
	testhelper.AssertEqual(t, "Normalize() OK cname=bluemix", nil, err5)
	testhelper.AssertEqual(t, "Normalize() OK Socname=bluemix", expected5, normalized5)

	crn6, _ := Parse("crn:v1:ibmcloud:public:service-name:location:scope:service-instance:resource-type:resource")
	normalized6, err6 := crn6.Normalize(ExpectEnvironment, ExpectServiceName, AllowSoftLayerCName, AllowOtherSegments)
	expected6, _ := Parse("crn:v1:bluemix:public:service-name:location:scope:service-instance:resource-type:resource")
	testhelper.AssertEqual(t, "Normalize() OK cname=ibmcloud", nil, err6)
	testhelper.AssertEqual(t, "Normalize() OK cname=ibmcloud", expected6, normalized6)

}

func TestParseAll(t *testing.T) {
	input := "crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"

	crn, err := ParseAll(input)
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "Parse", MaskAll{"crn", "v1", "cname", "ctype", "service-name", "location", "scope", "service-instance", "resource-type", "resource"}, crn)

	output := crn.ToCRNString()
	testhelper.AssertEqual(t, "CRNMask.String()", String(input), output)
}

func TestIsAnyIBMCloud(t *testing.T) {
	input := "crn:v1:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"

	crn, _ := Parse(input)

	output := crn.IsAnyIBMCloud()
	assert.Equal(t, false, output)
}

func TestIsAnyIBMPublic(t *testing.T) {
	input := "crn:v1:bluemix:public:service-name:location:scope:service-instance:resource-type:resource"

	crn, _ := Parse(input)

	output := crn.IsIBMPublicCloud()
	assert.Equal(t, true, output)
}

func TestSetPublicCloud(t *testing.T) {
	input := "crn:v1:::service-name::scope:service-instance:resource-type:resource"
	crn, _ := Parse(input)
	expected := Mask{"bluemix", "public", "service-name", "us-east", "scope", "service-instance", "resource-type", "resource"}
	actual := crn.SetPublicCloud("us-east")
	assert.Equal(t, expected, actual)
}

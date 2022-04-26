package osscatalog

import "testing"

func TestValidate(t *testing.T) {

	crn := "crn:v1:bluemix:public:cloudantnosqldb:us-east::::"
	ok := IsValidCRNFormat(crn)
	if !ok {
		t.Error("CRN failed: " + crn)
	}

	crn = "crn:v1:bluemix:public:cloudant nosqldb:us-east::::"
	ok = IsValidCRNFormat(crn)
	if ok {
		t.Error("CRN should have failed: " + crn)
	}

	crn = "crn:v1:bluemix:public:cloudant&nosqldb:us-east::::"
	ok = IsValidCRNFormat(crn)
	if ok {
		t.Error("CRN should have failed: " + crn)
	}

	crn = "crn:v1:bluemix:public:cloudantNosqldb:us-east::::"
	ok = IsValidCRNFormat(crn)
	if ok {
		t.Error("CRN should have failed: " + crn)
	}

	crn = "crn:v2:bluemix:public:cloudantnosqldb:us-east::::"
	ok = IsValidCRNFormat(crn)
	if ok {
		t.Error("CRN should have failed: " + crn)
	}

	crn = "crr:v1:bluemix:public:cloudantnosqldb:us-east::::"
	ok = IsValidCRNFormat(crn)
	if ok {
		t.Error("CRN should have failed: " + crn)
	}
}

func TestParse(t *testing.T) {

	crn := "crn:v1:bluemix:public:cloudantnosqldb:us-east::::"
	parsed := ParseCRN(crn)
	if parsed == nil {
		t.Error("CRN cannot parse: " + crn)
	}

	if parsed.Version != "v1" {
		t.Error("Invalid version" + parsed.Version)
	}
	if parsed.Cname != "bluemix" {
		t.Error("Invalid Cname" + parsed.Cname)
	}
	if parsed.Ctype != "public" {
		t.Error("Invalid Ctype" + parsed.Ctype)
	}
	if parsed.ServiceName != "cloudantnosqldb" {
		t.Error("Invalid ServiceName" + parsed.ServiceName)
	}
	if parsed.Location != "us-east" {
		t.Error("Invalid Location " + parsed.Location)
	}
	if parsed.Scope != "" {
		t.Error("Invalid Scope")
	}
	if parsed.ServiceInstance != "" {
		t.Error("Invalid ServiceInstance")
	}
	if parsed.ResourceType != "" {
		t.Error("Invalid ResourceType")
	}
	if parsed.Resource != "" {
		t.Error("Invalid Resource")
	}

}

func TestParseCRN(t *testing.T) {
	crn := "crn:v1:bluemix:public:file-storage:dal03::::"
	pc := ParseCRN(crn)

	if pc.Location != "dal03" {
		t.Error("Wrong location " + pc.Location)
	}

	if crn != pc.String() {
		t.Error("Could not create crn")
	}
}

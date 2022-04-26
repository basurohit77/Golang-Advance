package main

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestStructIterator(t *testing.T) {
	oss := ossrecord.CreateTestRecord()

	it := NewStructIterator(&oss.GeneralInfo)

	fieldIndex := 0
	test := func(expectedName string, expectedValue string) {
		n, v := it.Next()
		if n != expectedName {
			t.Errorf(`Unexpected name at field index %d: expected "%v" got "%v"`, fieldIndex, expectedName, n)
		}
		actualValue := fmt.Sprintf("%v", v)
		if actualValue != expectedValue {
			t.Errorf(`Unexpected value at field index %d: expected "%v" got "%v"`, fieldIndex, expectedValue, v)
		}
		fieldIndex++
	}

	test("RMCNumber", "12345")
	test("OperationalStatus", "BETA")
	test("FutureOperationalStatus", "")
	test("OSSTags", "[oss_status_green not_ready]")
	test("OSSOnboardingPhase", "")
	test("OSSOnboardingApprover", "")
	test("OSSOnboardingApprovalDate", "")
	test("ClientFacing", "true")
	test("EntryType", "SERVICE")
	test("OSSDescription", "This is a test record for oss-catalog functions")
	test("ServiceNowSysid", "ea56778eebed67")
	test("ServiceNowCIURL", "")
	test("ParentResourceName", "")
	test("Domain", "COMMERCIAL")
	test("", "<nil>")
	test("", "<nil>")
}

func TestStructIteratorSlice(t *testing.T) {
	oss := ossrecord.CreateTestRecord()

	it := NewStructIterator(&oss.GeneralInfo)

	slice := it.Slice()

	testhelper.AssertEqual(t, "slice[0].Name", "RMCNumber", slice[0].Name)
}

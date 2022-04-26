package ossmerge

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"

	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

func TestProcessDuplicateNames(t *testing.T) {
	options.LoadGlobalOptions("-keyfile <none>", true)

	oss1 := ossrecord.CreateTestRecord()
	oss1.ReferenceResourceName = ossrecord.CRNServiceName("service-1")
	m1, _ := LookupService(MakeComparableName(string(oss1.ReferenceResourceName)), true)
	m1.OSSService = *oss1
	m1.OSSService.ReferenceDisplayName = "Service 1"
	m1.SourceMainCatalog.Name = string(oss1.ReferenceResourceName)
	m1.OSSMergeControl = ossmergecontrol.New("service-1")
	m1.OSSValidation = ossvalidation.New("service-1", "test-timestamp")

	oss2 := ossrecord.CreateTestRecord()
	oss2.ReferenceResourceName = ossrecord.CRNServiceName("service-2")
	m2, _ := LookupService(MakeComparableName(string(oss2.ReferenceResourceName)), true)
	m2.OSSService = *oss2
	m2.OSSService.ReferenceDisplayName = "Service 2"
	m2.SourceMainCatalog.Name = string(oss2.ReferenceResourceName)
	m2.OSSMergeControl = ossmergecontrol.New("service-2")
	m2.OSSMergeControl.RawDuplicateNames = []string{"service-1", "service-4"}
	m2.OSSValidation = ossvalidation.New("service-2", "test-timestamp")

	oss3 := ossrecord.CreateTestRecord()
	oss3.ReferenceResourceName = ossrecord.CRNServiceName("service-3")
	m3, _ := LookupService(MakeComparableName(string(oss3.ReferenceResourceName)), true)
	m3.OSSService = *oss3
	m3.OSSService.ReferenceDisplayName = "Service 3"
	m3.SourceMainCatalog.Name = string(oss3.ReferenceResourceName)
	m3.OSSMergeControl = ossmergecontrol.New("service-3")
	m3.OSSMergeControl.RawDuplicateNames = []string{"service-5"}
	m3.OSSValidation = ossvalidation.New("service-3", "test-timestamp")

	if *testhelper.VeryVerbose {
		debug.SetDebugFlags(debug.Merge)
	}

	err := processDuplicateNames()
	testhelper.AssertError(t, err)
	var found bool
	m1After, _ := LookupService(MakeComparableName(string(oss1.ReferenceResourceName)), false)
	testhelper.AssertEqual(t, "duplicate record", "ServiceInfo(service1)  DuplicateOf=\"service2\"  CatalogMain=\"service-1\"  ScorecardV1=\"\"  ServiceNow=\"\"", m1After.Dump())

	if found {
		t.Errorf("Did not expect to find %s", m1After.String())
	}
	m2After, _ := LookupService(MakeComparableName(string(oss2.ReferenceResourceName)), false)
	testhelper.AssertEqual(t, "merged record", "ServiceInfo(service2)  CatalogMain=\"service-2\"+[\"service-1\",]  ScorecardV1=\"\"  ServiceNow=\"\"   OSSMergeControl=<non-empty>", m2After.Dump())
	m3After, _ := LookupService(MakeComparableName(string(oss3.ReferenceResourceName)), false)
	testhelper.AssertEqual(t, "intact record", m3.String(), m3After.String())

}

package ownership

import (
	"bytes"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestOwnershipXL(t *testing.T) {
	ossmerge.ResetForTesting()

	options.LoadGlobalOptions("-keyfile <none>", true)
	pattern := regexp.MustCompile(".*")
	//	pattern := regexp.MustCompile(regexp.QuoteMeta("service-1"))

	oss1 := ossrecord.CreateTestRecord()
	oss1.ReferenceResourceName = ossrecord.CRNServiceName("service-1")
	m1, _ := ossmerge.LookupService(string(oss1.ReferenceResourceName), true)
	m1.OSSService = *oss1
	m1.OSSService.ReferenceDisplayName = "Service 1"
	m1.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
	m1.OSSService.GeneralInfo.EntryType = ossrecord.SERVICE
	m1.SourceMainCatalog.Name = string(oss1.ReferenceResourceName)
	m1.SourceMainCatalog.Active = true
	m1.SourceMainCatalog.Kind = "service"
	m1.OSSMergeControl = ossmergecontrol.New("service-1")
	m1.OSSValidation = ossvalidation.New("service-1", "test-timestamp")
	m1.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m1.OSSService.Ownership.SegmentName = "<none>"
	m1.OSSService.Ownership.TribeName = "<none>"
	m1.OSSService.MonitoringInfo.Metrics = append(m1.OSSService.MonitoringInfo.Metrics, &ossrecord.Metric{})

	m1.SetFinalized()

	oss2 := ossrecord.CreateTestRecord()
	oss2.ReferenceResourceName = ossrecord.CRNServiceName("service-2")
	m2, _ := ossmerge.LookupService(string(oss2.ReferenceResourceName), true)
	m2.OSSService = *oss2
	m2.OSSService.ReferenceDisplayName = "Service 2"
	m2.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
	m2.OSSService.GeneralInfo.EntryType = ossrecord.RUNTIME
	m2.SourceMainCatalog.Name = string(oss2.ReferenceResourceName)
	m2.SourceMainCatalog.Active = true
	m2.SourceMainCatalog.Kind = "runtime"
	m2.OSSMergeControl = ossmergecontrol.New("service-2")
	m2.OSSValidation = ossvalidation.New("service-2", "test-timestamp")
	m2.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m2.SetFinalized()

	oss3 := ossrecord.CreateTestRecord()
	oss3.ReferenceResourceName = ossrecord.CRNServiceName("service-3")
	m3, _ := ossmerge.LookupService(string(oss3.ReferenceResourceName), true)
	m3.OSSService = *oss3
	m3.OSSService.ReferenceDisplayName = "Service 3"
	m3.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
	m3.OSSService.GeneralInfo.EntryType = ossrecord.SERVICE
	m3.SourceMainCatalog.Name = string(oss3.ReferenceResourceName)
	m3.SourceMainCatalog.Active = true
	m3.SourceMainCatalog.Kind = "service"
	m3.OSSMergeControl = ossmergecontrol.New("service-3")
	m3.OSSValidation = ossvalidation.New("service-3", "test-timestamp")
	m3.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m3.SetFinalized()
	m3.AddNamedValidationIssue(ossvalidation.NoValidOwnership, "")

	if testing.Verbose() {
		previous := debug.SetDebugFlags(debug.Reports)
		defer debug.SetDebugFlags(previous)
	}

	buffer := new(bytes.Buffer)

	err := RunReportXL(buffer, timeStamp, pattern)
	testhelper.AssertError(t, err)

	if buffer.Len() < 9000 {
		t.Errorf("Output appears to be too short (%d characters)", buffer.Len())
	}

	/*
		fname := "TestOwnershipXL.xlsx"
		file, err := os.Create(fname)
		testhelper.AssertError(t, err)
		defer file.Close()
		_, err = buffer.WriteTo(file)
		testhelper.AssertError(t, err)
	*/
}

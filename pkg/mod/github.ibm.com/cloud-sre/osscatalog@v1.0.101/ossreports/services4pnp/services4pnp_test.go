package services4pnp

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

var timeLocation = func() *time.Location {
	loc, _ := time.LoadLocation("UTC")
	return loc
}()
var timeStamp = time.Now().In(timeLocation).Format("2006-01-02T1504Z")

func TestServices4PnP(t *testing.T) {
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
	m1.OSSService.GeneralInfo.OSSTags = osstags.TagSet{osstags.StatusCRNGreen, osstags.OneCloud, osstags.PnPEnabled}
	m1.OSSService.StatusPage.CategoryID = "category-1"
	m1.OSSService.StatusPage.CategoryParent = oss1.ReferenceResourceName
	m1.OSSService.Compliance.ServiceNowOnboarded = true
	m1.SourceServiceNow.CRNServiceName = string(oss1.ReferenceResourceName)
	m1.SourceServiceNow.StatusPage.CategoryID = "category-1"
	m1.OSSMergeControl = ossmergecontrol.New("service-1")
	m1.OSSValidation = ossvalidation.New("service-1", "test-timestamp")
	m1.OSSValidation.AddSource("service-1", ossvalidation.CATALOG)
	m1.OSSValidation.AddSource("service-1", ossvalidation.SERVICENOW)
	m1.OSSValidation.StatusCategoryCount = 2
	m1.OSSValidation.AddNamedIssue(ossvalidation.PnPEnabledWithoutCandidate, "")
	m1.OSSValidation.AddIssue(ossvalidation.SEVERE, "Some test issue (1)", "").TagStatusPage()
	m1.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m1.SetFinalized()

	oss2 := ossrecord.CreateTestRecord()
	oss2.ReferenceResourceName = ossrecord.CRNServiceName("service-2")
	m2, _ := ossmerge.LookupService(string(oss2.ReferenceResourceName), true)
	m2.OSSService = *oss2
	m2.OSSService.ReferenceDisplayName = "Service 2"
	m2.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
	m2.OSSService.GeneralInfo.EntryType = ossrecord.RUNTIME
	m2.SourceMainCatalog.Name = ""
	m2.SourceMainCatalog.Active = true
	m2.SourceMainCatalog.Kind = "runtime"
	m2.OSSService.GeneralInfo.OSSTags = osstags.TagSet{osstags.StatusCRNGreen, osstags.PnPCandidate}
	m2.OSSService.StatusPage.CategoryID = "category-1"
	m2.OSSService.StatusPage.CategoryParent = "service-4"
	m2.OSSService.Compliance.ServiceNowOnboarded = true
	m2.SourceServiceNow.CRNServiceName = string(oss2.ReferenceResourceName) + "_X"
	m2.SourceServiceNow.StatusPage.CategoryID = "category-1"
	m2.OSSMergeControl = ossmergecontrol.New("service-2")
	m2.OSSMergeControl.Notes = "Some notes"
	m2.OSSValidation = ossvalidation.New("service-2", "test-timestamp")
	m2.OSSValidation.AddSource("service-2", ossvalidation.CATALOG)
	m2.OSSValidation.AddSource("service-2", ossvalidation.SERVICENOW)
	m2.OSSValidation.StatusCategoryCount = 2
	m2.OSSValidation.AddNamedIssue(ossvalidation.PnPNotEnabledButCandidate, "")
	m2.OSSValidation.AddIssue(ossvalidation.SEVERE, "Some test issue (2)", "").TagStatusPage()
	m2.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m2.SetFinalized()

	oss3 := ossrecord.CreateTestRecord()
	oss3.ReferenceResourceName = ossrecord.CRNServiceName("service-3")
	m3, _ := ossmerge.LookupService(string(oss3.ReferenceResourceName), true)
	m3.OSSService = *oss3
	m3.OSSService.ReferenceDisplayName = "Service 3"
	m3.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
	m3.OSSService.GeneralInfo.EntryType = ossrecord.SERVICE
	m3.SourceMainCatalog.Name = string(oss3.ReferenceResourceName) + "_Y"
	m3.SourceMainCatalog.Active = true
	m3.SourceMainCatalog.Kind = "service"
	m3.OSSService.GeneralInfo.OSSTags = osstags.TagSet{osstags.StatusCRNGreen, osstags.PnPInclude}
	m3.OSSService.StatusPage.CategoryID = "category-3"
	m3.OSSService.StatusPage.CategoryParent = ""
	m3.OSSService.Compliance.ServiceNowOnboarded = true
	m3.SourceServiceNow.CRNServiceName = ""
	m3.SourceServiceNow.StatusPage.CategoryID = "category-3"
	m3.OSSMergeControl = ossmergecontrol.New("service-3")
	m3.OSSValidation = ossvalidation.New("service-3", "test-timestamp")
	m3.OSSValidation.AddSource("service-3", ossvalidation.CATALOG)
	m3.OSSValidation.AddSource("service-3", ossvalidation.SERVICENOW)
	m3.OSSValidation.StatusCategoryCount = 0
	m3.OSSValidation.AddNamedIssue(ossvalidation.PnPNotEnabledButInclude, "")
	m3.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m3.SetFinalized()

	if testing.Verbose() {
		previous := debug.SetDebugFlags(debug.Reports)
		defer debug.SetDebugFlags(previous)
	}

	buffer := new(bytes.Buffer)

	err := Services4PnP(buffer, timeStamp, pattern)
	testhelper.AssertError(t, err)

	if buffer.Len() < 8000 {
		t.Errorf("Output appears to be too short (%d characters)", buffer.Len())
	}

	/*
		fname := "TestServices4PnP.xlsx"
		file, err := os.Create(fname)
		testhelper.AssertError(t, err)
		defer file.Close()
		_, err = buffer.WriteTo(file)
		testhelper.AssertError(t, err)
	*/
}

package catalogvisibilitygroups

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
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

func TestCatalogVisibilityGroups(t *testing.T) {
	options.LoadGlobalOptions("-keyfile <none>", true)
	pattern := regexp.MustCompile(".*")
	//	pattern := regexp.MustCompile(regexp.QuoteMeta("service-1"))

	oss1 := ossrecord.CreateTestRecord()
	oss1.ReferenceResourceName = ossrecord.CRNServiceName("service-1")
	m1, _ := ossmerge.LookupService(string(oss1.ReferenceResourceName), true)
	m1.OSSService = *oss1
	m1.OSSService.ReferenceDisplayName = "Service 1"
	m1.SourceMainCatalog.Name = string(oss1.ReferenceResourceName)
	m1.SourceMainCatalog.Kind = "service"
	m1.SourceMainCatalog.Tags = []string{"deprecated"}
	m1.SourceMainCatalog.EffectiveVisibility.Restrictions = string(catalogapi.VisibilityIBMOnly)
	m1.OSSMergeControl = ossmergecontrol.New("service-1")
	m1.OSSValidation = ossvalidation.New("service-1", "test-timestamp")
	m1.OSSValidation.CatalogVisibility.EffectiveRestrictions = m1.SourceMainCatalog.EffectiveVisibility.Restrictions
	m1.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m1.SetFinalized()

	oss2 := ossrecord.CreateTestRecord()
	oss2.ReferenceResourceName = ossrecord.CRNServiceName("service-2")
	m2, _ := ossmerge.LookupService(string(oss2.ReferenceResourceName), true)
	m2.OSSService = *oss2
	m2.OSSService.ReferenceDisplayName = "Service 2"
	m2.SourceMainCatalog.Name = string(oss2.ReferenceResourceName)
	m2.SourceMainCatalog.Kind = "service"
	m2.SourceMainCatalog.EffectiveVisibility.Restrictions = string(catalogapi.VisibilityIBMOnly)
	m2.OSSMergeControl = ossmergecontrol.New("service-2")
	m2.OSSValidation = ossvalidation.New("service-2", "test-timestamp")
	m2.OSSValidation.CatalogVisibility.EffectiveRestrictions = m2.SourceMainCatalog.EffectiveVisibility.Restrictions
	m2.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m2.SetFinalized()

	oss3 := ossrecord.CreateTestRecord()
	oss3.ReferenceResourceName = ossrecord.CRNServiceName("service-3")
	m3, _ := ossmerge.LookupService(string(oss3.ReferenceResourceName), true)
	m3.OSSService = *oss3
	m3.OSSService.ReferenceDisplayName = "Service 3"
	m3.SourceMainCatalog.Name = string(oss3.ReferenceResourceName)
	m3.SourceMainCatalog.Kind = "runtime"
	m3.SourceMainCatalog.EffectiveVisibility.Restrictions = string(catalogapi.VisibilityPrivate)
	m3.OSSMergeControl = ossmergecontrol.New("service-3")
	m3.OSSValidation = ossvalidation.New("service-3", "test-timestamp")
	m3.OSSValidation.CatalogVisibility.EffectiveRestrictions = m3.SourceMainCatalog.EffectiveVisibility.Restrictions
	m3.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m3.SetFinalized()

	var buffer strings.Builder
	//	w := os.Stdout
	err := CatalogVisibilityGroups(&buffer, timeStamp, pattern)

	if *testhelper.VeryVerbose {
		fmt.Print(buffer.String())
	}

	testhelper.AssertError(t, err)
	if buffer.Len() < 800 {
		t.Errorf("Output appears to be too short (%d characters)", buffer.Len())
	}
}

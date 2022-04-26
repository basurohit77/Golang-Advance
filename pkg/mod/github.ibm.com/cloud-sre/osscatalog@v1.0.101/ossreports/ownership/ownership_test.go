package ownership

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

var timeLocation = func() *time.Location {
	loc, _ := time.LoadLocation("UTC")
	return loc
}()
var timeStamp = time.Now().In(timeLocation).Format("2006-01-02T1504Z")

func TestOwnership(t *testing.T) {
	options.LoadGlobalOptions("-keyfile <none>", true)
	pattern := regexp.MustCompile(".*")

	m1, _ := ossmerge.LookupService("service-1", true)
	m1.OSSService.ReferenceResourceName = "service-1"
	m1.OSSService.ReferenceDisplayName = "Service 1"
	m1.OSSService.GeneralInfo.EntryType = ossrecord.SERVICE
	m1.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
	m1.OSSService.GeneralInfo.OSSTags = osstags.TagSet{osstags.StatusGreen, osstags.NotReady}
	m1.OSSService.Ownership.OfferingManager = ossrecord.Person{Name: "John Doe", W3ID: "johndoe@us.ibm.com"}
	m1.OSSService.Ownership.DevelopmentManager = ossrecord.Person{Name: "Jane Somebody"}
	//	m1.OSSService.Ownership.TechnicalContact = ossrecord.Person{W3ID: "howard@us.ibm.com"}
	m1.OSSService.Compliance.ArchitectureFocal = ossrecord.Person{W3ID: "howard@us.ibm.com"}
	m1.OSSService.Ownership.SegmentName = "Watson and Cloud CTO"
	m1.OSSService.Ownership.SegmentOwner = ossrecord.Person{Name: "Bryson Koehler"}
	m1.OSSService.Ownership.TribeName = "CTO Global Technology Operations"
	m1.OSSService.Ownership.TribeOwner = ossrecord.Person{Name: "Shaun Smith"}
	m1.OSSService.Support.Manager = ossrecord.Person{Name: "Shawn Bramblett"}
	m1.OSSService.Operations.Manager = ossrecord.Person{Name: "Shawn Bramblett"}
	m1.OSSService.CatalogInfo.Provider = ossrecord.Person{Name: "IBM", W3ID: "bshawn@us.ibm.com"}
	m1.OSSService.CatalogInfo.ProviderContact = "Shawn Bramblett"
	m1.OSSService.CatalogInfo.ProviderSupportEmail = "bshawn@us.ibm.com"
	m1.OSSService.CatalogInfo.ProviderPhone = "+1-720-349-602"
	m1.OSSMergeControl = ossmergecontrol.New("service-1")
	m1.OSSValidation = ossvalidation.New("service-1", "test-timestamp")
	m1.OSSValidation.AddSource("service-1", ossvalidation.CATALOG)
	m1.OSSValidation.AddSource("service-1", ossvalidation.PRIOROSS)
	m1.OSSValidation.AddSource("name-variant-1a", ossvalidation.SCORECARDV1)
	m1.OSSValidation.AddSource("name-variant-1b", ossvalidation.SERVICENOW)
	m1.OSSValidation.AddSource(string(m1.OSSValidation.CanonicalName), ossvalidation.CATALOG)
	m1.SetFinalized()

	m2, _ := ossmerge.LookupService("service-2", true)
	*m2 = *m1
	m2.OSSService.ReferenceResourceName = "service-2"
	m2.OSSService.ReferenceDisplayName = "Service 2"
	m2.OSSMergeControl.CanonicalName = "service-2"
	m2.OSSValidation.CanonicalName = "service-2"

	var buffer strings.Builder
	//	w := os.Stdout
	err := RunReport(&buffer, timeStamp, pattern)

	if *testhelper.VeryVerbose {
		fmt.Print(buffer.String())
	}

	testhelper.AssertError(t, err)
	if buffer.Len() < 1900 {
		t.Errorf("Output appears to be too short (%d characters)", buffer.Len())
	}
}

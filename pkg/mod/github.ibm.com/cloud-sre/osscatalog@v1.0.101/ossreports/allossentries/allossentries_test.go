package allossentries

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"

	"github.ibm.com/cloud-sre/osscatalog/crn"

	"github.ibm.com/cloud-sre/osscatalog/stats"

	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
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

func TestAllOSSEntries(t *testing.T) {
	options.LoadGlobalOptions("-keyfile <none>", true)
	pattern := regexp.MustCompile(".*")
	//	pattern := regexp.MustCompile(regexp.QuoteMeta("service-1"))

	ossrunactions.Enable([]string{"services", "tribes", "environments"})

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
	m1.OSSValidation.AddIssue(ossvalidation.MINOR, "Sample minor issue", "#1")
	m1.GetCatalogExtra(true).Locations = collections.NewStringSet("us-south", "us-east", "sjc01", "par02")
	m1.OSSService.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNGreen)
	m1.OSSService.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusYellow)
	m1.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OneCloud)
	m1.OSSService.MonitoringInfo.Metrics = append(m1.OSSService.MonitoringInfo.Metrics, &ossrecord.Metric{})
	m1.SetFinalized()
	stats.GetGlobalActualStats().RecordAction(stats.ActionCreate, options.RunModeRW, false, m1.OSSEntry(), nil)

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
	m2.GetCatalogExtra(true).Locations = collections.NewStringSet("us-south", "eu-gb", "sjc01", "tor03")
	m2.OSSService.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNYellow)
	m2.OSSService.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusRed)
	m2.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OneCloudWave1)
	m2.OSSValidation.AddIssue(ossvalidation.WARNING, "Sample warning issue", "")
	m2.OSSValidation.AddIssue(ossvalidation.MINOR, "Sample minor issue", "#2")
	m2.OSSService.MonitoringInfo.Metrics = append(m2.OSSService.MonitoringInfo.Metrics, &ossrecord.Metric{})
	m2.SetFinalized()
	stats.GetGlobalActualStats().RecordAction(stats.ActionCreate, options.RunModeRW, false, m2.OSSEntry(), nil)

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
	m3.GetCatalogExtra(true).Locations = collections.NewStringSet("us-south", "sao04")
	m3.OSSService.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNRed)
	m3.OSSService.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusGreen)
	m3.OSSValidation.AddIssue(ossvalidation.SEVERE, "Sample severe issue", "")
	m3.OSSValidation.AddIssue(ossvalidation.CRITICAL, "Sample critical issue", "")
	m3.SetFinalized()
	stats.GetGlobalActualStats().RecordAction(stats.ActionCreate, options.RunModeRW, false, m3.OSSEntry(), nil)

	seg1, _ := ossmerge.LookupSegment("segment1", true)
	seg1.OSSSegment.DisplayName = "Segment 1"
	seg1.OSSSegment.Owner = ossrecord.Person{Name: "John Doe"}
	seg1.SourceScorecardV1.Name = seg1.OSSSegment.DisplayName
	stats.GetGlobalActualStats().RecordAction(stats.ActionCreate, options.RunModeRW, false, seg1.OSSEntry(), nil)
	tr1, _ := seg1.LookupTribe("tribe1", true)
	tr1.OSSTribe.DisplayName = "Tribe 1"
	tr1.OSSTribe.Owner = ossrecord.Person{Name: "Fred Flintstone"}
	tr1.SourceScorecardV1.Name = tr1.OSSTribe.DisplayName
	stats.GetGlobalActualStats().RecordAction(stats.ActionCreate, options.RunModeRW, false, tr1.OSSEntry(), nil)
	tr2, _ := seg1.LookupTribe("tribe2", true)
	tr2.OSSTribe.DisplayName = "Tribe 2"
	tr2.OSSTribe.Owner = ossrecord.Person{Name: "Barney Rubble"}
	tr2.OSSTribe.ChangeApprovers = append(tr2.OSSTribe.ChangeApprovers, &ossrecord.PersonListEntry{Member: ossrecord.Person{Name: "Larry"}})
	tr2.OSSTribe.ChangeApprovers = append(tr2.OSSTribe.ChangeApprovers, &ossrecord.PersonListEntry{Member: ossrecord.Person{Name: "Curly"}})
	tr2.OSSTribe.ChangeApprovers = append(tr2.OSSTribe.ChangeApprovers, &ossrecord.PersonListEntry{Member: ossrecord.Person{Name: "Moe"}})
	tr2.SourceScorecardV1.Name = tr2.OSSTribe.DisplayName
	stats.GetGlobalActualStats().RecordAction(stats.ActionCreate, options.RunModeRW, false, tr2.OSSEntry(), nil)

	env1ID, _ := crn.Parse("crn:v1:the-north:gaas::eu-gb::::")
	env1, _ := ossmerge.LookupEnvironment(env1ID, true)
	env1.OSSEnvironment.EnvironmentID = ossrecord.EnvironmentID(env1ID.ToCRNString())
	env1.OSSEnvironment.DisplayName = "*TEST RECORD* The North"
	env1.OSSEnvironment.OSSTags.AddTag(osstags.OSSTest)
	osstags.CheckOSSTestTag(&env1.OSSEnvironment.DisplayName, &env1.OSSEnvironment.OSSTags)
	env1.OSSEnvironment.Type = ossrecord.EnvironmentGAAS
	env1.OSSEnvironment.Status = ossrecord.EnvironmentActive
	env1.OSSValidation = ossvalidation.New(string(env1.OSSEnvironment.EnvironmentID), "test-timestamp")
	env1.OSSValidation.AddIssue(ossvalidation.SEVERE, "Sample severe issue for environment", "")
	stats.GetGlobalActualStats().RecordAction(stats.ActionCreate, options.RunModeRW, false, env1.OSSEntry(), nil)

	if *testhelper.VeryVerbose {
		previous := debug.SetDebugFlags(debug.Reports)
		defer debug.SetDebugFlags(previous)
	}

	buffer := new(bytes.Buffer)

	if testing.Verbose() {
		previous := debug.SetDebugFlags(debug.Reports)
		defer debug.SetDebugFlags(previous)
	}

	err := RunReport(buffer, timeStamp, pattern)
	testhelper.AssertError(t, err)

	if buffer.Len() < 7900 {
		t.Errorf("Output appears to be too short (%d characters)", buffer.Len())
	}

	/*
		fname := "AllOSSEntries.xlsx"
		file, err := os.Create(fname)
		testhelper.AssertError(t, err)
		defer file.Close()
		_, err = buffer.WriteTo(file)
		testhelper.AssertError(t, err)
	*/
}

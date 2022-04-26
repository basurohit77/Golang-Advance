package main

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestEmitJSONShort(t *testing.T) {
	var v = &ViewData{}
	var out string

	out = v.EmitJSONShort("")
	testhelper.AssertEqual(t, "empty string", ``, out)

	out = v.EmitJSONShort("string1-value")
	testhelper.AssertEqual(t, "string1", `string1-value`, out)

	out = v.EmitJSONShort([]string{})
	testhelper.AssertEqual(t, "empty slice", `[]`, out)

	out = v.EmitJSONShort(([]string)(nil))
	testhelper.AssertEqual(t, "nil slice", `[]`, out)

	out = v.EmitJSONShort([]string{"name1", "name2"})
	testhelper.AssertEqual(t, "slice1", `["name1","name2"]`, out)

	out = v.EmitJSONShort((map[string]string)(nil))
	testhelper.AssertEqual(t, "nil map", `{}`, out)
}

func TestEmitJSONLong(t *testing.T) {
	var v = &ViewData{}
	var out string

	out = v.EmitJSONLong("")
	testhelper.AssertEqual(t, "empty string", ``, out)

	out = v.EmitJSONLong("string1-value")
	testhelper.AssertEqual(t, "string1", `string1-value`, out)

	out = v.EmitJSONLong([]string{})
	testhelper.AssertEqual(t, "empty slice", `[]`, out)

	out = v.EmitJSONLong(([]string)(nil))
	testhelper.AssertEqual(t, "nil slice", `[]`, out)

	out = v.EmitJSONLong([]string{"name1", "name2"})
	testhelper.AssertEqual(t, "slice1", "[\n\"name1\",\n\"name2\"\n]", out)

	out = v.EmitJSONLong((map[string]string)(nil))
	testhelper.AssertEqual(t, "nil map", `{}`, out)
}

func TestRenderView(t *testing.T) {
	var err error

	//	debug.SetDebugFlags(debug.MergeControl)

	options.LoadGlobalOptions("-keyfile <none>", true)

	viewFuncs := make(map[string]interface{})
	viewFuncs["getUniversalLink"] = getUniversalLink
	viewTemplate, err := template.New("view.html").Funcs(viewFuncs).ParseFiles("views/view.html")
	if err != nil {
		t.Fatalf("Could not create the view.html template: %v", err)
	}

	userInfo := "someone@somewhere.com"

	ar := createTestCache()

	//	ossrec := catalog.NewOSSServiceExtended("test-oss-record")
	ossrec := ossrecordextended.NewOSSServiceExtended("osscatalog-testing.1") // Reference a name from the cache, in order to pick-up children information
	ossrec.OSSService.ReferenceDisplayName = "Display name for a test OSS record. This could be a long name"
	ossrec.OSSService.AdditionalContacts = "Name1 Email1 Phone1;  Name2 Email2 Phone2;  Name3 Email3 Phone3\n(second line) Name4 Email4 Phone4;  Name5 Email6 Phone7;"
	ossrec.OSSService.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNGreen)
	ossrec.OSSService.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusRed)
	ossrec.OSSService.GeneralInfo.ServiceNowCIURL = "https://watson.service-now.com"
	ossrec.OSSService.GeneralInfo.ParentResourceName = ossrecord.CRNServiceName("osscatalog-testing.2")
	ossrec.OSSService.Ownership.SegmentName = "A Segment Name"
	ossrec.OSSService.Ownership.SegmentID = "oss_segment.a-segment-id"
	ossrec.OSSService.Ownership.TribeName = "A Tribe Name"
	ossrec.OSSService.Ownership.TribeID = "oss_tribe.a-tribe-id"
	ossrec.OSSService.ServiceNowInfo.SupportNotApplicable = true
	ossrec.OSSService.Support.SpecialInstructions = "Should not display, because SupportNotApplicable=true"
	ossrec.OSSService.ServiceNowInfo.OperationsNotApplicable = false
	ossrec.OSSService.Operations.SpecialInstructions = "Should display, because OperationsNotApplicable=false"
	ossrec.OSSService.ProductInfo.ClearingHouseReferences = []ossrecord.ClearingHouseReference{
		{Name: "CHName1", ID: "CHID1", Tags: []string{"tag1", "tag2"}},
		{Name: "CHName2", ID: "CHID2", Tags: []string{"tag3"}},
	}
	ossrec.OSSService.ProductInfo.Taxonomy.MajorUnitUTL10 = "myUTL10"
	ossrec.OSSService.ProductInfo.Taxonomy.MinorUnitUTL15 = "myUTL15"
	ossrec.OSSService.DependencyInfo.OutboundDependencies = []*ossrecord.Dependency{
		{Service: "service1", Tags: []string{"tag1", "tag2"}},
		{Service: "service2", Tags: []string{"tag3"}},
		{Service: `{service3[chid:12345]}`, Tags: []string{"tag1"}},
	}
	ossrec.OSSService.DependencyInfo.InboundDependencies = []*ossrecord.Dependency{
		{Service: "service1", Tags: []string{"tag1", "tag2"}},
		{Service: "service2", Tags: []string{"tag3"}},
		{Service: `{service3[chid:12345]}`, Tags: []string{"tag1"}},
		{Service: `~~~ + 1 potentially non-Cloud items`, Tags: []string{"tag1"}},
	}
	ossrec.OSSService.MonitoringInfo.Metrics = []*ossrecord.Metric{
		{Type: ossrecord.MetricProvisioning, PlanOrLabel: "Plan1", Environment: "crn:v1:bluemix:public::au-syd::::", Tags: []string{"Src:EDB"}},
		{Type: ossrecord.MetricConsumption, PlanOrLabel: "Plan2", Environment: "crn:v1:bluemix:public::us-south::::", Tags: []string{"Src:EDB"}},
	}
	ossrec.OSSMergeControl.RawDuplicateNames = []string{"name1", "name2"}
	ossrec.OSSMergeControl.DoNotMergeNames = []string{"name3", "name4"}
	ossrec.OSSMergeControl.Notes = "Some notes about the merge for this entry. Could possibly expand to more than one line.\nLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."
	ossrec.OSSMergeControl.Overrides = map[string]interface{}{"Attribute1": "Value1", "Attribute2": "Value2", "Attribute3": "Value3", "Attribute4": "Value4"}
	ossrec.OSSMergeControl.UpdatedBy = userInfo
	ossrec.OSSMergeControl.LastUpdate = "2018-07-27 16:00 UTC"
	ossrec.OSSMergeControl.RefreshChecksum(true)
	ossrec.OSSValidation.CanonicalNameSources = []ossvalidation.Source{"source1", "source2"}
	ossrec.OSSValidation.OtherNamesSources = make(map[string][]ossvalidation.Source)
	ossrec.OSSValidation.OtherNamesSources["name5"] = []ossvalidation.Source{"source5.1", "source5.2"}
	ossrec.OSSValidation.OtherNamesSources["name6"] = []ossvalidation.Source{"source6.1", "source6.2"}
	ossrec.OSSValidation.CatalogVisibility.EffectiveRestrictions = string(catalogapi.VisibilityIBMOnly)
	ossrec.OSSValidation.StatusCategoryCount = 2
	ossrec.OSSValidation.AddIssue(ossvalidation.CRITICAL, "issue5", "").TagDataMissing()
	ossrec.OSSValidation.AddIssue(ossvalidation.WARNING, "issue1", "details1").TagDataMismatch()
	ossrec.OSSValidation.AddIssue(ossvalidation.SEVERE, "issue2", "details2").TagCRN()
	ossrec.OSSValidation.AddIssue(ossvalidation.CRITICAL, "issue3", "details3").TagRunAction(ossrunactions.DependenciesClearingHouse)
	ossrec.OSSValidation.AddIssue(ossvalidation.INFO, "issue4", "details4")
	ossrec.OSSValidation.RecordRunAction(ossrunactions.ProductInfoParts)
	ossrec.OSSValidation.Sort()

	viewData1 := &ViewData{}
	viewData1.populate(ossrec, ar, userInfo, false, false)
	debug.Debug(debug.MergeControl, `SignatureWarning="%s"    DeleteWarning="%s"`, viewData1.SignatureWarning, viewData1.DeleteWarning)
	var buffer1 strings.Builder
	err = viewTemplate.Execute(&buffer1, viewData1)
	if err != nil {
		t.Fatalf("Error executing template (update disabled): %v", err)
	}

	viewData2 := &ViewData{}
	viewData2.populate(ossrec, ar, userInfo, true, true)
	viewData2.EditMode = true
	debug.Debug(debug.MergeControl, `SignatureWarning="%s"    DeleteWarning="%s"`, viewData2.SignatureWarning, viewData2.DeleteWarning)
	var buffer2 strings.Builder
	err = viewTemplate.Execute(&buffer2, viewData2)
	if err != nil {
		t.Fatalf("Error executing template (update enabled): %v", err)
	}

	viewData3 := &ViewData{}
	ossrec.OSSMergeControl = ossmergecontrol.New(string(ossrec.OSSService.ReferenceResourceName))
	ossrec.OSSValidation = ossvalidation.New("test-entry", "test-timestamp")
	ossrec.OSSValidation.AddSource(string(ossrec.OSSService.ReferenceResourceName), ossvalidation.PRIOROSS)
	viewData3.populate(ossrec, ar, userInfo, true, false)
	debug.Debug(debug.MergeControl, `SignatureWarning="%s"    DeleteWarning="%s"`, viewData3.SignatureWarning, viewData3.DeleteWarning)
	var buffer3 strings.Builder
	err = viewTemplate.Execute(&buffer3, viewData3)
	if err != nil {
		t.Fatalf("Error executing template (with warnings): %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer1.String())
		fmt.Println(buffer2.String())
		fmt.Println(buffer3.String())
	}

	out1, _ := os.Create("/Users/dpj/00tmp/TestRender.view.RO.html")
	defer out1.Close()
	io.WriteString(out1, fmt.Sprint(buffer1.String()))
	out2, _ := os.Create("/Users/dpj/00tmp/TestRender.view.EDIT.html")
	defer out2.Close()
	io.WriteString(out2, fmt.Sprint(buffer2.String()))
	out3, _ := os.Create("/Users/dpj/00tmp/TestRender.view.WARN.html")
	defer out3.Close()
	io.WriteString(out3, fmt.Sprint(buffer3.String()))
	/*
	 */
}

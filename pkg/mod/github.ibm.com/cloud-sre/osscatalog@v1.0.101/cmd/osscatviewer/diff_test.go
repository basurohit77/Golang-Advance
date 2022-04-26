package main

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRenderDiff(t *testing.T) {

	//	debug.SetDebugFlags(debug.MergeControl)

	diffTemplate, err := template.ParseFiles("views/diff.html")
	if err != nil {
		t.Fatalf("Could not create the diff.html template: %v", err)
	}

	userInfo := "someone@somewhere.com"

	ossrec := ossrecord.OSSService{ReferenceResourceName: "test-oss-record"}
	ossrec.ReferenceDisplayName = "Display name for a test OSS record. This could be a long name"
	ossrec.AdditionalContacts = "Name1 Email1 Phone1;  Name2 Email2 Phone2;  Name3 Email3 Phone3\n(second line) Name4 Email4 Phone4;  Name5 Email6 Phone7;"
	ossrec.GeneralInfo.OperationalStatus = ossrecord.GA
	ossrec.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNGreen)
	ossrec.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusRed)
	ossrec.GeneralInfo.ServiceNowCIURL = "https://watson.service-now.com"
	ossrec.Ownership.SegmentName = "A Segment Name"
	ossrec.Ownership.SegmentID = "oss_segment.a-segment-id"
	ossrec.Ownership.TribeName = "A Tribe Name"
	ossrec.Ownership.TribeID = "oss_tribe.a-tribe-id"
	ossrec.Support.Manager = ossrecord.Person{Name: "SupportManager", W3ID: "support@ibm.com"}

	ossrecProduction := ossrec
	ossrecProduction.ReferenceDisplayName = ossrec.ReferenceDisplayName + ".PRODUCTION"
	ossrecProduction.GeneralInfo.OperationalStatus = ""
	ossrecProduction.GeneralInfo.EntryType = ossrecord.SERVICE
	ossrecProduction.Support.Manager = ossrecord.Person{Name: "SupportManager", W3ID: "supportProduction@ibm.com"}

	diffData1 := &DiffData{}
	diffData1.populate(&ossrec, &ossrecProduction, userInfo)
	var buffer1 strings.Builder
	err = diffTemplate.Execute(&buffer1, diffData1)
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}
	diffData2 := &DiffData{}
	diffData2.populate(&ossrec, nil, userInfo)
	var buffer2 strings.Builder
	err = diffTemplate.Execute(&buffer2, diffData2)
	if err != nil {
		t.Fatalf("Error executing template (staging only): %v", err)
	}
	diffData3 := &DiffData{}
	diffData3.populate(nil, &ossrecProduction, userInfo)
	var buffer3 strings.Builder
	err = diffTemplate.Execute(&buffer3, diffData3)
	if err != nil {
		t.Fatalf("Error executing template (production only): %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer1.String())
		fmt.Println(buffer2.String())
		fmt.Println(buffer3.String())
	}

	/*
		out1, _ := os.Create("/Users/dpj/00tmp/TestRender.diff.html")
		defer out1.Close()
		io.WriteString(out1, fmt.Sprint(buffer1.String()))
		out2, _ := os.Create("/Users/dpj/00tmp/TestRender.diff.staging-only.html")
		defer out2.Close()
		io.WriteString(out2, fmt.Sprint(buffer2.String()))
		out3, _ := os.Create("/Users/dpj/00tmp/TestRender.diff.production-only.html")
		defer out3.Close()
		io.WriteString(out3, fmt.Sprint(buffer3.String()))
	*/
}

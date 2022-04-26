package main

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRenderSegment(t *testing.T) {

	segmentTemplate, err := template.ParseFiles("views/segment.html")
	if err != nil {
		t.Fatalf("Could not create the segment.html template: %v", err)
	}

	userInfo := "someone@somewhere.com"

	ar := createTestCache()

	seg := ossrecordextended.NewOSSSegmentExtended(ar.segments[0].SegmentID)
	seg.DisplayName = ar.getSegmentName(seg.SegmentID)
	seg.Owner = ossrecord.Person{Name: "John Doe Owner", W3ID: "johndoe@us.ibm.com"}
	seg.TechnicalContact = ossrecord.Person{Name: "John Doe Tech Contact", W3ID: "johndoe@us.ibm.com"}
	seg.ChangeCommApprovers = make([]*ossrecord.PersonListEntry, 2, 10)
	seg.ChangeCommApprovers[0] = &ossrecord.PersonListEntry{
		Member: ossrecord.Person{Name: "Pinky", W3ID: "pinky@us.ibm.com"},
		Tags:   []string{},
	}
	seg.ChangeCommApprovers[1] = &ossrecord.PersonListEntry{
		Member: ossrecord.Person{Name: "Brain", W3ID: "brain@us.ibm.com"},
		Tags:   []string{"the_boss"},
	}
	seg.ERCAApprovers = make([]*ossrecord.PersonListEntry, 2, 10)
	seg.ERCAApprovers[0] = &ossrecord.PersonListEntry{
		Member: ossrecord.Person{Name: "Ren", W3ID: "ren@us.ibm.com"},
		Tags:   []string{"dog"},
	}
	seg.ERCAApprovers[1] = &ossrecord.PersonListEntry{
		Member: ossrecord.Person{Name: "Stimpy", W3ID: "stimpy@us.ibm.com"},
		Tags:   []string{"cat"},
	}
	seg.SchemaVersion = ossrecord.OSSCurrentSchema

	segmentData1 := &SegmentData{}
	segmentData1.populate(seg, ar, userInfo, false, false)
	var buffer1 strings.Builder
	err = segmentTemplate.Execute(&buffer1, segmentData1)
	if err != nil {
		t.Fatalf("Error executing template (update disabled): %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer1.String())
	}

	/*
		out1, _ := os.Create("/Users/dpj/00tmp/TestRender.segment.html")
		defer out1.Close()
		io.WriteString(out1, fmt.Sprint(buffer1.String()))
	*/
}

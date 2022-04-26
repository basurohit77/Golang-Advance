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

func TestRenderTribe(t *testing.T) {

	tribeTemplate, err := template.ParseFiles("views/tribe.html")
	if err != nil {
		t.Fatalf("Could not create the tribe.html template: %v", err)
	}

	userInfo := "someone@somewhere.com"

	ar := createTestCache()

	tr := ossrecordextended.NewOSSTribeExtended(ar.segments[0].Tribes[0].TribeID)
	tr.SegmentID = ar.segments[0].SegmentID
	tr.DisplayName = ar.segments[0].Tribes[0].DisplayName
	tr.Owner = ossrecord.Person{Name: "John Doe Owner", W3ID: "johndoe@us.ibm.com"}
	tr.ChangeApprovers = make([]*ossrecord.PersonListEntry, 2, 10)
	tr.ChangeApprovers[0] = &ossrecord.PersonListEntry{
		Member: ossrecord.Person{Name: "John Doe Approver2", W3ID: "johndoe@us.ibm.com"},
		Tags:   []string{"tag1", "tag2"},
	}
	tr.ChangeApprovers[1] = &ossrecord.PersonListEntry{
		Member: ossrecord.Person{Name: "John Doe Approver1", W3ID: "johndoe@us.ibm.com"},
		Tags:   []string{"tag1"},
	}
	tr.SchemaVersion = ossrecord.OSSCurrentSchema

	tribeData1 := &TribeData{}
	tribeData1.populate(tr, ar, userInfo, false, false)
	var buffer1 strings.Builder
	err = tribeTemplate.Execute(&buffer1, tribeData1)
	if err != nil {
		t.Fatalf("Error executing template (update disabled): %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer1.String())
	}

	/*
		out1, _ := os.Create("/Users/dpj/00tmp/TestRender.tribe.html")
		defer out1.Close()
		io.WriteString(out1, fmt.Sprint(buffer1.String()))
	*/
}

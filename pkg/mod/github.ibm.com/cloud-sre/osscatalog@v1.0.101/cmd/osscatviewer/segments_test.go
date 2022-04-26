package main

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRenderSegments(t *testing.T) {

	segmentsTemplate, err := template.ParseFiles("views/segments.html")
	if err != nil {
		t.Fatalf("Could not create the segments.html template: %v", err)
	}

	userInfo := "someone@somewhere.com"

	ar := createTestCache()

	segmentsData1 := &SegmentsData{}
	segmentsData1.LastRefresh = ar.lastRefresh.Format(timeFormat)
	segmentsData1.LoggedInUser = userInfo
	segmentsData1.Segments = ar.segments

	var buffer1 strings.Builder
	err = segmentsTemplate.Execute(&buffer1, segmentsData1)
	if err != nil {
		t.Fatalf("Error executing template (update disabled): %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer1.String())
	}

	/*
		out1, _ := os.Create("/Users/dpj/00tmp/TestRender.segments.html")
		defer out1.Close()
		io.WriteString(out1, fmt.Sprint(buffer1.String()))
	*/
}

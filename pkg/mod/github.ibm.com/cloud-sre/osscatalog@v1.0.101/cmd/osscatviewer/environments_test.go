package main

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRenderEnvironments(t *testing.T) {

	environmentsTemplate, err := template.ParseFiles("views/environments.html")
	if err != nil {
		t.Fatalf("Could not create the environments.html template: %v", err)
	}

	userInfo := "someone@somewhere.com"

	ar := createTestCache()

	environmentsData1 := &EnvironmentsData{}
	environmentsData1.LastRefresh = ar.lastRefresh.Format(timeFormat)
	environmentsData1.LoggedInUser = userInfo
	environmentsData1.Environments = ar.environments

	var buffer1 strings.Builder
	err = environmentsTemplate.Execute(&buffer1, environmentsData1)
	if err != nil {
		t.Fatalf("Error executing template (update disabled): %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer1.String())
	}

	/*
		out1, _ := os.Create("/Users/dpj/00tmp/TestRender.environments.html")
		defer out1.Close()
		io.WriteString(out1, fmt.Sprint(buffer1.String()))
	*/
}

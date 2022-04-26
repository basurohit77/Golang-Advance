package main

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRenderHome(t *testing.T) {

	userInfo := "someone@somewhere.com"

	ar := createTestCache()

	homeData := HomeData{}
	homeData.Pattern = ".*"
	homeData.LastRefresh = ar.lastRefresh.Format(timeFormat)
	homeData.LoggedInUser = userInfo
	homeData.Services = ar.services

	homeTemplate, err := template.ParseFiles("views/home.html")
	if err != nil {
		t.Fatalf("Could not create the home.html template: %v", err)
	}

	var buffer strings.Builder

	err = homeTemplate.Execute(&buffer, &homeData)
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer.String())
	}
	/*
		out, _ := os.Create("/Users/dpj/00tmp/TestRender.home.html")
		defer out.Close()
		io.WriteString(out, fmt.Sprint(buffer.String()))
	*/
}

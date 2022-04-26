package main

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestRenderEnvironment(t *testing.T) {
	options.LoadGlobalOptions("-keyfile <none>", true)

	environmentTemplate, err := template.ParseFiles("views/environment.html")
	if err != nil {
		t.Fatalf("Could not create the environment.html template: %v", err)
	}

	userInfo := "someone@somewhere.com"

	ar := createTestCache()

	env := &ossrecordextended.OSSEnvironmentExtended{}
	env.OSSValidation = ossvalidation.New("test-id", "test-timestamp")
	env.EnvironmentID = ar.environments[0].EnvironmentID
	env.DisplayName = ar.environments[0].DisplayName
	env.Type = ar.environments[0].Type
	env.Status = ar.environments[0].Status
	env.OSSTags = osstags.TagSet{osstags.IBMCloudDefaultSegment}
	env.SchemaVersion = ossrecord.OSSCurrentSchema
	env.OSSValidation.AddIssue(ossvalidation.INFO, "Info #1", "").TagEnvironments()
	env.OSSValidation.AddIssue(ossvalidation.WARNING, "Warning #1", "").TagEnvironments()
	env.OSSValidation.AddSource("name1", ossvalidation.CATALOG)
	env.OSSValidation.AddSource("name2", ossvalidation.DOCTORENV)

	environmentData1 := &EnvironmentData{}
	environmentData1.populate(env, ar, userInfo, false, false)
	var buffer1 strings.Builder
	err = environmentTemplate.Execute(&buffer1, environmentData1)
	if err != nil {
		t.Fatalf("Error executing template (update disabled): %v", err)
	}

	if *testhelper.VeryVerbose {
		fmt.Println(buffer1.String())
	}

	/*
		out1, _ := os.Create("/Users/dpj/00tmp/TestRender.environment.html")
		defer out1.Close()
		io.WriteString(out1, fmt.Sprint(buffer1.String()))
	*/
}

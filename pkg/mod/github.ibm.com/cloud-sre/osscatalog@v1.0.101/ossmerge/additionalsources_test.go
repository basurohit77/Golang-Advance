package ossmerge

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/scorecardv1"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestDebugPrintAdditionalNames(t *testing.T) {
	input := make([]*catalogapi.Resource, 3)
	input[0] = &catalogapi.Resource{Name: "name0", Kind: "kind0", ID: "id0"}
	input[1] = &catalogapi.Resource{Name: "name1", Kind: "kind1", ID: "id1"}
	input[2] = &catalogapi.Resource{Name: "name2", Kind: "kind2", ID: "id2"}

	result := debugPrintAdditionalNames(input)

	if *testhelper.VeryVerbose /* || true /* XXX */ {
		fmt.Println("Result -> ", result)
	}

	testhelper.AssertEqual(t, "", `[Catalog{Name:"name0", Kind:"kind0", ID:"id0"}  Catalog{Name:"name1", Kind:"kind1", ID:"id1"}  Catalog{Name:"name2", Kind:"kind2", ID:"id2"}]`, result)
}

func TestCheckAdditionalScorecardV1(t *testing.T) {
	si := &ServiceInfo{}
	si.PriorOSS.ReferenceResourceName = "cli-repo"
	si.SourceScorecardV1Detail = scorecardv1.DetailEntry{Name: "bluemix-cli-repo"}
	si.AdditionalScorecardV1Detail = append(si.AdditionalScorecardV1Detail, &scorecardv1.DetailEntry{Name: "cli-repo"})

	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Fine)
	}

	si.checkAdditionalScorecardV1()

	testhelper.AssertEqual(t, "SourceScorecardV1Detail", "cli-repo", si.SourceScorecardV1Detail.Name)
	testhelper.AssertEqual(t, "len(AdditionalScorecardV1Detail)", 1, len(si.AdditionalScorecardV1Detail))
	testhelper.AssertEqual(t, "AdditionalScorecardV1Detail[0]", "bluemix-cli-repo", si.AdditionalScorecardV1Detail[0].Name)
}

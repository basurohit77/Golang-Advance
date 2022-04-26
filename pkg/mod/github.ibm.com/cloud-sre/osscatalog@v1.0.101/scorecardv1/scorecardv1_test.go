package scorecardv1

import (
	"fmt"
	"regexp"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestReadScorecardV1Record(t *testing.T) {
	t.Skip("Skipping test TestReadScorecardV1Record() - Doctor disabled")

	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestReadScorecardV1Record() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.ScorecardV1)
	}

	name := "appid"
	productionReadiness := false // Note: not working with productionReadiness=true for appid

	rest.LoadDefaultKeyFile()

	result, err := ReadScorecardV1Record(ossrecord.CRNServiceName(name), productionReadiness)

	if err != nil {
		t.Errorf("ReadScorecardV1Record failed: %v", err)
	}
	if result.Name != name {
		t.Errorf("ReadScorecardV1Record did not return a record with the expected name \"%s\"", name)
	}

	if productionReadiness && result.ProductionReadinessData == nil {
		t.Errorf("ReadScorecardV1Record did not find ProductionReadiness info for \"%s\"", name)
	}
}

// FIXME: not working with productionReadiness=true for appid
func NOTestReadScorecardV1ProductionReadiness(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestReadScorecardV1Record() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.ScorecardV1)
	}

	name := "appid"
	productionReadiness := true

	rest.LoadDefaultKeyFile()

	result, err := ReadScorecardV1Record(ossrecord.CRNServiceName(name), productionReadiness)

	if err != nil {
		t.Errorf("ReadScorecardV1Record failed: %v", err)
	}
	if result.Name != name {
		t.Errorf("ReadScorecardV1Record did not return a record with the expected name \"%s\"", name)
	}

	if productionReadiness && result.ProductionReadinessData == nil {
		t.Errorf("ReadScorecardV1Record did not find ProductionReadiness info for \"%s\"", name)
	}
}
func TestReadScorecardV1Detail(t *testing.T) {
	t.Skip("Skipping test TestReadScorecardV1Detail() - Doctor disabled")

	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestReadScorecardV1Detail() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ScorecardV1 /* | debug.Fine /* XXX */)
	}

	name := "appid"

	rest.LoadDefaultKeyFile()

	result, err := ReadScorecardV1Detail(ossrecord.CRNServiceName(name))

	if err != nil {
		t.Errorf("ReadScorecardV1Detail failed: %v", err)
	} else if result.Name != name {
		t.Errorf("ReadScorecardV1Detail did not return a record with the expected name \"%s\"", name)
	}
}

func TestListSegments(t *testing.T) {
	t.Skip("Skipping test TestListSegments() - Doctor disabled")

	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListSegments() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ScorecardV1 | debug.IAM /* | debug.Fine /* XXX */)
		*testhelper.VeryVerbose = true // XXX
	}

	rest.LoadDefaultKeyFile()

	countResults := 0

	err := ListSegments(func(e *SegmentResource) error {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Printf(" -> found entry %q   %+v\n", e.Name, e)
		}
		return nil
	})
	testhelper.AssertError(t, err)
	if countResults < 5 {
		t.Errorf("ListSegments() returned only %d entries -- too few", countResults)
	}
}

func TestListTribes(t *testing.T) {
	t.Skip("Skipping test TestListTribes() - Doctor disabled")

	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListTribes() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.ScorecardV1 | debug.IAM)
	}

	segment := SegmentResource{
		Name: "Core Platform",
		//Href: "https://api-oss.bluemix.net/scorecard/api/segmenttribe/v1/segments/58eda55b9babda00075a50da",
		Href: "https://pnp-api-oss.cloud.ibm.com/scorecard/api/segmenttribe/v1/segments/58eda55b9babda00075a50da",
	}
	/*
		segment := SegmentResource{
			Name: "Cloud Platform Client Success",
			Href: "https://api-oss.bluemix.net/scorecard/api/segmenttribe/v1/segments/58eda55b9babda00075a50dd",
		}
	*/
	segment.Tribes.Href = segment.Href + "/tribes"

	rest.LoadDefaultKeyFile()

	countResults := 0

	err := ListTribes(&segment, func(e *TribeResource) error {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Println(" -> found entry", e.Name, e.Href)
		}
		return nil
	})
	testhelper.AssertError(t, err)
	if countResults < 5 {
		t.Errorf("TestListTribes(%s) returned only %d entries -- too few", segment.Name, countResults)
	}
}

func TestListScorecardV1Details(t *testing.T) {
	t.Skip("Skipping test TestListScorecardV1Details() - Doctor disabled")

	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListScorecardV1Details() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.ScorecardV1 /* | debug.Fine /* XXX */)
		/* *testhelper.VeryVerbose = true /* XXX */
	}

	rest.LoadDefaultKeyFile()

	//	pattern := regexp.MustCompile(".*node.*")
	pattern := regexp.MustCompile(".*")

	countResults := 0

	err := ListScorecardV1Details(pattern, func(e *DetailEntry) {
		countResults++
		if *testhelper.VeryVerbose {
			fmt.Println(" -> found entry", e.Name)
		}
	})

	if err != nil {
		t.Errorf("TestListScorecardV1Details failed: %v", err)
	}
	if countResults < 100 {
		t.Errorf("TestListScorecardV1Details returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d entries from ScorecardV1\n", countResults)
	}

}

func TestGetSegmentID(t *testing.T) {
	segment := SegmentResource{
		Name: "Core Platform",
		//Href: "https://api-oss.bluemix.net/scorecard/api/segmenttribe/v1/segments/58eda55b9babda00075a50da",
		Href: "https://pnp-api-oss.cloud.ibm.com/scorecard/api/segmenttribe/v1/segments/58eda55b9babda00075a50da",
	}

	id := segment.GetSegmentID()
	testhelper.AssertEqual(t, "", "58eda55b9babda00075a50da", string(id))
}

func TestGetTribeID(t *testing.T) {
	tribe := TribeResource{
		Name: "Core Platform",
		//Href: "https://api-oss.bluemix.net/scorecard/api/segmenttribe/v1/segments/58eda55b9babda00075a50da/tribes/58eda5669babda00075a50e8",
		Href: "https://pnp-api-oss.cloud.ibm.com/scorecard/api/segmenttribe/v1/segments/58eda55b9babda00075a50da/tribes/58eda5669babda00075a50e8",
	}

	id := tribe.GetTribeID()
	testhelper.AssertEqual(t, "", "58eda55b9babda00075a50da-58eda5669babda00075a50e8", string(id))
}

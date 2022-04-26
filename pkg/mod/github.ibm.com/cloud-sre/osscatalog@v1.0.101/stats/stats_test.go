package stats

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"

	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
)

func TestReport(t *testing.T) {
	//*testhelper.VeryVerbose = true // XXX

	options.LoadGlobalOptions("-keyfile <none>", true)

	data := newData()
	var svc *ossrecordextended.OSSServiceExtended
	var prior ossrecord.OSSService

	svc = ossrecordextended.NewOSSServiceExtended("service-1")
	svc.GeneralInfo.OSSTags.AddTag(osstags.StatusCRNYellow)
	svc.GeneralInfo.OSSTags.AddTag(osstags.StatusRed)
	data.RecordAction(ActionCreate, options.RunModeRW, false, &svc.OSSService, nil)

	svc = ossrecordextended.NewOSSServiceExtended("service-2")
	svc.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	data.RecordAction(ActionCreate, options.RunModeRW, false, &svc.OSSService, nil)

	svc = ossrecordextended.NewOSSServiceExtended("service-3")
	prior = svc.OSSService
	svc.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	svc.GeneralInfo.OSSTags.AddTag(osstags.OneCloud)
	svc.GeneralInfo.OSSTags.AddTag(osstags.StatusCRNGreen)
	svc.GeneralInfo.OSSTags.AddTag(osstags.StatusYellow)
	data.RecordAction(ActionUpdate, options.RunModeRW, false, &svc.OSSService, &prior)

	svc = ossrecordextended.NewOSSServiceExtended("service-4")
	prior = svc.OSSService
	prior.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	data.RecordAction(ActionUpdate, options.RunModeRW, false, &svc.OSSService, &prior)

	svc = ossrecordextended.NewOSSServiceExtended("service-5")
	prior = svc.OSSService
	data.RecordAction(ActionDelete, options.RunModeRW, false, &svc.OSSService, &prior)

	svc = ossrecordextended.NewOSSServiceExtended("service-6")
	prior = svc.OSSService
	prior.GeneralInfo.OSSTags.AddTag(osstags.PnPEnabled)
	data.RecordAction(ActionDelete, options.RunModeRW, false, &svc.OSSService, &prior)

	numEntries := data.NumEntries()
	testhelper.AssertEqual(t, "NumEntries()", 6, numEntries)

	result := data.Report("TEST: ")
	if *testhelper.VeryVerbose {
		fmt.Println(result)
	}
	testhelper.AssertEqual(t, "len(Report())", 2438, len(result))
}

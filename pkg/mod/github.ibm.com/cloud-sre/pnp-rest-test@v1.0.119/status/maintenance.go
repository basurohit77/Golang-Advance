package status

import (
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/testutils"
)

// MaintenanceTest runs the maintenance test
func (api *API) MaintenanceTest() {

	lg.Info("MaintenanceTest", "Executing maintenance tests")

	api.AllMaintenanceTest()

}

// AllMaintenanceTest gets all maintenances
func (api *API) AllMaintenanceTest() {

	METHOD := "AllMaintenanceTest"
	lg.Info("AllMaintenanceTest", "Querying all maintenances test")

	list := new(MaintenanceList)
	err := api.cat.Server.GetAndDecode(METHOD, "maintenance.MaintenanceList", api.apiInfo.Maintenances.Href, list)
	if err != nil {
		return
	}

	api.checkMaintenanceList(METHOD, list)

}

func (api *API) checkMaintenanceList(fct string, list *MaintenanceList) {

	METHOD := fct + "->checkMaintenanceList"

	if len(list.Resources) == 0 {
		lg.Err(METHOD, nil, "No resources returned in query.")
		return
	}

	expected := 200
	if list.Limit > expected {
		lg.Err(METHOD, nil, "Limit for maintenances is greater than expected maximum. Is %d, should be %d.", list.Limit, expected)
	}

	checkList := new(MaintenanceList)
	err := api.cat.Server.GetAndDecode(METHOD, "maintenance.MaintenanceList", list.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on maintenance list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on maintenance list length using href")
	}

	checkList = new(MaintenanceList)
	err = api.cat.Server.GetAndDecode(METHOD, "maintenance.MaintenanceList", list.First.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get First on maintenance list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on maintenance list length using first")
	}

	checkList = new(MaintenanceList)
	err = api.cat.Server.GetAndDecode(METHOD, "maintenance.MaintenanceList", list.Last.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Last on maintenance list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, nil, "Did not get a match on maintenance list length using last")
	}

	for _, item := range list.Resources {
		api.checkMaintenance(METHOD, item)
	}
}

func (api *API) checkMaintenance(fct string, inc MaintenanceGet) {
	METHOD := fct + "->checkMaintenance"

	if inc.RecordID == "" {
		lg.Err(METHOD, nil, "No record ID in the maintenance")
	}

	testutils.CheckTime(METHOD, "Maintenance.CreationTime("+inc.RecordID+")", inc.CreationTime)
	testutils.CheckTime(METHOD, "Maintenance.UpdateTime("+inc.RecordID+")", inc.UpdateTime)
	testutils.CheckTime(METHOD, "Maintenance.PlannedStart("+inc.RecordID+")", inc.PlannedStart)

	if inc.PlannedEnd != "" {
		testutils.CheckTime(METHOD, "Maintenance.PlannedEnd("+inc.RecordID+")", inc.PlannedEnd)
	}

	testutils.CheckEnum(METHOD, "Maintenance.Kind", inc.Kind, "maintenance")

	testutils.CheckNoBlankValue(METHOD, "ShortDescription", inc.ShortDescription)
	testutils.CheckNoBlankValue(METHOD, "LongDescription", inc.LongDescription)
	testutils.CheckNoBlankValue(METHOD, "Source", inc.Source)
	testutils.CheckNoBlankValue(METHOD, "SourceID", inc.SourceID)

	if len(inc.CRNMasks) == 0 {
		lg.Err(METHOD, nil, "No CRN Masks in maintenance")
	}

	testutils.CheckEnum(METHOD, "Maintenance.State", inc.State, "new", "scheduled", "in-progress", "complete")
	testutils.CheckEnum(METHOD, "Maintenance.Disruptive", inc.Disruptive, "true", "false")

	checkMaintenance := new(MaintenanceGet)
	err := api.cat.Server.GetAndDecode(METHOD, "maintenance.Maintenance", inc.Href, checkMaintenance)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on maintenance.")
	}
	if checkMaintenance.RecordID != inc.RecordID {
		lg.Err(METHOD, err, "Href maintenance does not match original")
	}
}

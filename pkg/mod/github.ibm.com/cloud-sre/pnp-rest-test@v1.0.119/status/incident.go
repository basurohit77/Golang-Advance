package status

import (
	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/testutils"
)

const (
	// KnownIncidentID is the ID of a SN incident that should exist that is a sev 1 confirmed cie
	KnownIncidentID = "INC0277848"
	// KnownIncidentSD is the ShortDescription of the known incident
	KnownIncidentSD = "authoring requests failing"
)

// IncidentTest runs the incident test
func (api *API) IncidentTest() {
	METHOD := "IncidentTest"

	lg.Info("IncidentTest", "Executing incident tests")

	api.AllIncidentTest(METHOD)
	api.GetByMask(METHOD)

}

// AllIncidentTest gets all incidents
func (api *API) AllIncidentTest(fct string) {

	METHOD := fct + "->AllIncidentTest"
	lg.Info("AllIncidentTest", "Querying all incidents test")

	list := new(IncidentList)
	err := api.cat.Server.GetAndDecode(METHOD, "incident.IncidentList", api.apiInfo.Incidents.Href, list)
	if err != nil {
		return
	}

	api.checkIncidentList(METHOD, list)
}

func (api *API) checkIncidentList(fct string, list *IncidentList) {

	METHOD := fct + "->checkIncidentList"

	if len(list.Resources) == 0 {
		lg.Err(METHOD, nil, "No resources returned in query.")
		return
	}

	expected := 200
	if list.Limit > expected {
		lg.Err(METHOD, nil, "Limit for incidents is greater than expected maximum. Is %d, should be %d.", list.Limit, expected)
	}

	checkList := new(IncidentList)
	err := api.cat.Server.GetAndDecode(METHOD, "incident.IncidentList", list.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on incident list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on incident list length using href %d != %d", list.Count, checkList.Count)
	}

	checkList = new(IncidentList)
	err = api.cat.Server.GetAndDecode(METHOD, "incident.IncidentList", list.First.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get First on incident list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on incident list length using first")
	}

	checkList = new(IncidentList)
	err = api.cat.Server.GetAndDecode(METHOD, "incident.IncidentList", list.Last.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Last on incident list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, nil, "Did not get a match on incident list length using last")
	}

	for _, item := range list.Resources {
		api.checkIncident(METHOD, item)
	}
}

func (api *API) checkIncident(fct string, inc IncidentGet) {
	METHOD := fct + "->checkIncident"

	if inc.RecordID == "" {
		lg.Err(METHOD, nil, "No record ID in the incident")
	}

	testutils.CheckTime(METHOD, "Incident.CreationTime", inc.CreationTime)
	testutils.CheckTime(METHOD, "Incident.UpdateTime", inc.UpdateTime)
	testutils.CheckTime(METHOD, "Incident.OutageStart", inc.OutageStart)
	testutils.CheckTime(METHOD, "Incident.OutageEnd", inc.OutageEnd)

	testutils.CheckEnum(METHOD, "Incident.Kind", inc.Kind, "incident")

	testutils.CheckNoBlankValue(METHOD, "ShortDescription", inc.ShortDescription)
	testutils.CheckNoBlankValue(METHOD, "LongDescription", inc.LongDescription)
	testutils.CheckNoBlankValue(METHOD, "Source", inc.Source)
	testutils.CheckNoBlankValue(METHOD, "SourceID", inc.SourceID)

	if len(inc.CRNMasks) == 0 {
		lg.Err(METHOD, nil, "No CRN Masks in incident")
	}

	testutils.CheckEnum(METHOD, "Incident.State", inc.State, "new", "in-progress", "resolved")
	testutils.CheckEnum(METHOD, "Incident.Classification", inc.Classification, "confirmed-cie", "potential-cie", "normal")

	if inc.Severity < 1 || inc.Severity > 4 {
		lg.Err(METHOD, nil, "Incident severity is incorrect %d", inc.Severity)
	}

	checkIncident := new(IncidentGet)
	err := api.cat.Server.GetAndDecode(METHOD, "incident.Incident", inc.Href, checkIncident)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on incident.")
	}
	if checkIncident.RecordID != inc.RecordID {
		lg.Err(METHOD, err, "Href incident does not match original")
	}
}

// GetByMask will try to retrieve a known incident by ID.
func (api *API) GetByMask(fct string) {
	// Need to get a valid service to continue with this test
	//METHOD := fct + "->GetByMask"
	//crnMask := "crn:v1:bluemix:public:compose-for-postgresql:us-south::::"

}

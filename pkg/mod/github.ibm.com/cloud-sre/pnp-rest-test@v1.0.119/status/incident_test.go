package status

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/pnp-rest-test/catalog"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
	"github.ibm.com/cloud-sre/pnp-rest-test/testutils"
)

var incident = IncidentGet{
	//Href:             tsSingle,
	//RecordID:         KnownIncident,
	CreationTime:     testutils.GetDate(time.Hour),
	UpdateTime:       testutils.GetDate(time.Hour),
	Kind:             "incident",
	ShortDescription: "Short description 123 ",
	LongDescription:  "Long description 123",
	CRNMasks:         []string{"crn:v1:bluemix:public:cloudant:::::"},
	State:            "resolved",
	Classification:   "confirmed-cie",
	Severity:         1,
	OutageStart:      testutils.GetDate(time.Hour),
	OutageEnd:        testutils.GetDate(time.Hour),
	Source:           "SOURCE",
	SourceID:         "SOURCEID",
}

var incidentList *IncidentList

func TestIncidents(t *testing.T) {

	tsIncident := testutils.NewDataServer(serveIncident)
	incident.Href = tsIncident.URL
	defer tsIncident.Close()

	tsList := testutils.NewDataServer(serveIncidentList)
	defer tsList.Close()

	list := new(IncidentList)
	list.Offset = 0
	list.Limit = 1
	list.Count = 1
	list.Href = tsList.URL
	list.First.Href = tsList.URL
	list.Last.Href = tsList.URL

	list.Resources = append(list.Resources, incident)
	incidentList = list

	srvr := testutils.NewJSONServer(list)
	defer srvr.Close()

	apiInfo := new(APIInfo)
	apiInfo.Incidents.Href = srvr.URL

	api := &API{apiInfo: apiInfo, cat: &catalog.Catalog{Server: &rest.Server{}}}

	api.AllIncidentTest("TestIncidents")

}

func serveIncidentList(url *url.URL) ([]byte, int) {
	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(incidentList); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}

func serveIncident(url *url.URL) ([]byte, int) {

	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(incident); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}

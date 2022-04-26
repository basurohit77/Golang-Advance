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

var maintenance = MaintenanceGet{
	//Href:             tsSingle,
	RecordID:         "RECORDIDxyz",
	CreationTime:     testutils.GetDate(time.Hour),
	UpdateTime:       testutils.GetDate(time.Hour),
	Kind:             "maintenance",
	ShortDescription: "Short description 123 ",
	LongDescription:  "Long description 123",
	CRNMasks:         []string{"crn:v1:bluemix:public:cloudant:::::"},
	State:            "in-progress",
	Disruptive:       "true",
	PlannedStart:     testutils.GetDate(time.Hour),
	PlannedEnd:       testutils.GetDate(time.Hour),
	Source:           "SOURCE",
	SourceID:         "SOURCEID",
}

var maintenanceList *MaintenanceList

func TestMaintenances(t *testing.T) {

	tsMaintenance := testutils.NewDataServer(serveMaintenance)
	maintenance.Href = tsMaintenance.URL
	defer tsMaintenance.Close()

	tsList := testutils.NewDataServer(serveMaintList)
	defer tsList.Close()

	list := new(MaintenanceList)
	list.Offset = 0
	list.Limit = 1
	list.Count = 1
	list.Href = tsList.URL
	list.First.Href = tsList.URL
	list.Last.Href = tsList.URL

	list.Resources = append(list.Resources, maintenance)
	maintenanceList = list

	srvr := testutils.NewJSONServer(list)
	defer srvr.Close()

	apiInfo := new(APIInfo)
	apiInfo.Maintenances.Href = srvr.URL

	api := &API{apiInfo: apiInfo, cat: &catalog.Catalog{Server: &rest.Server{}}}

	api.AllMaintenanceTest()

}

func serveMaintList(url *url.URL) ([]byte, int) {
	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(maintenanceList); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}

func serveMaintenance(url *url.URL) ([]byte, int) {

	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(maintenance); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}

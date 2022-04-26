package status

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/pnp-rest-test/catalog"
	"github.ibm.com/cloud-sre/pnp-rest-test/common"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
	"github.ibm.com/cloud-sre/pnp-rest-test/testutils"
)

var resource = ResourceGet{
	//Href:             tsSingle,
	RecordID:         "RECORDID123",
	CreationTime:     testutils.GetDate(time.Hour),
	UpdateTime:       testutils.GetDate(time.Hour),
	Kind:             "resource",
	CRNMask:          "crn:v1:bluemix:public:cloudant:::::",
	State:            "ok",
	Status:           "ok",
	Visibility:       []string{"clientFacing", "hasStatus"},
	StatusUpdateTime: testutils.GetDate(time.Hour),
	Tags:             []common.SimpleTag{common.SimpleTag{ID: "tag1"}},
	DisplayName:      []common.TranslatedString{common.TranslatedString{Language: "en", Text: "cloudant"}},
}

var resourceList *ResourceList

func TestResources(t *testing.T) {

	tsResource := testutils.NewDataServer(serveResource)
	resource.Href = tsResource.URL
	defer tsResource.Close()

	tsList := testutils.NewDataServer(serveResList)
	defer tsList.Close()

	list := new(ResourceList)
	list.Offset = 0
	list.Limit = 1
	list.Count = 1
	list.Href = tsList.URL
	list.First.Href = tsList.URL
	list.Last.Href = tsList.URL

	list.Resources = append(list.Resources, resource)
	resourceList = list

	srvr := testutils.NewJSONServer(list)
	defer srvr.Close()

	apiInfo := new(APIInfo)
	apiInfo.Resources.Href = srvr.URL

	api := &API{apiInfo: apiInfo, cat: &catalog.Catalog{Server: &rest.Server{}}}

	api.AllResourceTest()

}

func serveResList(url *url.URL) ([]byte, int) {
	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(resourceList); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}

func serveResource(url *url.URL) ([]byte, int) {

	var err error
	buffer := new(bytes.Buffer)
	if err = json.NewEncoder(buffer).Encode(resource); err != nil {
		return nil, http.StatusInternalServerError
	}
	return buffer.Bytes(), http.StatusOK
}

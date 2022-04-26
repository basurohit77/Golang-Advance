package status

import (
	"strings"

	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/testutils"
)

// ResourceTest runs the resource test
func (api *API) ResourceTest() {

	lg.Info("ResourceTest", "Executing resource tests")

	api.AllResourceTest()

}

// AllResourceTest gets all resources
func (api *API) AllResourceTest() {

	METHOD := "AllResourceTest"
	lg.Info("AllResourceTest", "Querying all resources test")

	list := new(ResourceList)
	err := api.cat.Server.GetAndDecode(METHOD, "resource.ResourceList", api.apiInfo.Resources.Href, list)
	if err != nil {
		return
	}

	api.checkResourceList(METHOD, list)
}

func (api *API) checkResourceList(fct string, list *ResourceList) {

	METHOD := fct + "->checkResourceList"

	if len(list.Resources) == 0 {
		lg.Err(METHOD, nil, "No resources returned in query.")
		return
	}

	expected := 200
	if list.Limit > expected {
		lg.Err(METHOD, nil, "Limit for resources is greater than expected maximum. Is %d, should be %d.", list.Limit, expected)
	}

	checkList := new(ResourceList)
	err := api.cat.Server.GetAndDecode(METHOD, "resource.ResourceList", list.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on resource list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on resource list length using href")
	}

	checkList = new(ResourceList)
	err = api.cat.Server.GetAndDecode(METHOD, "resource.ResourceList", list.First.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get First on resource list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, err, "Did not get a match on resource list length using first")
	}

	checkList = new(ResourceList)
	err = api.cat.Server.GetAndDecode(METHOD, "resource.ResourceList", list.Last.Href, checkList)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Last on resource list.")
	}
	if list.Count != checkList.Count {
		lg.Err(METHOD, nil, "Did not get a match on resource list length using last original(%d) != last(%d)", list.Count, checkList.Count)
	}

	for _, item := range list.Resources {
		api.checkResource(METHOD, item)
	}

	api.tagCheckResourceList(METHOD, list)
}

func (api *API) checkResource(fct string, inc ResourceGet) {
	METHOD := fct + "->checkResource"

	if inc.RecordID == "" {
		lg.Err(METHOD, nil, "No record ID in the resource")
	}

	testutils.CheckTime(METHOD, "Resource.CreationTime", inc.CreationTime)
	testutils.CheckTime(METHOD, "Resource.UpdateTime", inc.UpdateTime)

	testutils.CheckEnum(METHOD, "Resource.Status", inc.Status, "", "ok", "degraded", "failed")
	if inc.Status != "" {
		testutils.CheckTime(METHOD, "Resource.StatusUpdateTime", inc.StatusUpdateTime)
	}

	testutils.CheckEnum(METHOD, "Resource.Kind", inc.Kind, "resource")

	testutils.CheckNoBlankValue(METHOD, "Resource.CRNMask", inc.CRNMask)

	testutils.CheckEnum(METHOD, "Resource.State", inc.State, "ok", "archived")

	if len(inc.DisplayName) == 0 {
		lg.Err(METHOD, nil, "Display name in resource is empty")
	}

	for _, v := range inc.Visibility {
		testutils.CheckEnum(METHOD, "Resource.Visibility", v, "hasStatus", "clientFacing")
	}

	checkResource := new(ResourceGet)
	err := api.cat.Server.GetAndDecode(METHOD, "resource.Resource", inc.Href, checkResource)
	if err != nil {
		lg.Err(METHOD, err, "Failed to get Href on resource.")
	}
	if checkResource.RecordID != inc.RecordID {
		lg.Err(METHOD, err, "Href resource does not match original")
	}
}

func (api *API) tagCheckResourceList(fct string, list *ResourceList) {
	METHOD := fct + "->tagCheckResourceList"

	hasStatus := 0
	clientFacings := 0

	for len(list.Resources) > 0 {

		for _, r := range list.Resources {
			for _, t := range r.Visibility {
				switch t {
				case "hasStatus":
					hasStatus++
				case "clientFacing":
					clientFacings++
				default:
				}
			}
		}

		if strings.TrimSpace(list.Next.Href) == "" {
			break
		} else {

			next := list.Next.Href
			list = new(ResourceList)
			err := api.cat.Server.GetAndDecode(METHOD, "resource.ResourceList", next, list)
			if err != nil {
				lg.Err(METHOD, err, "Failed to decode resource list from Next link.")
				return
			}
		}
	}

	if hasStatus == 0 {
		lg.Err(METHOD, nil, "No resources contain the 'hasStatus' tag")
	}
	if clientFacings == 0 {
		lg.Err(METHOD, nil, "No resources contain the 'clientFacing' tag")
	}
	lg.Info(METHOD, "tags: hasStatus=%d, clientFacing=%d\n", hasStatus, clientFacings)

}

package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

var (
	kongURL                    = "http://kongurl"
	catalogURL                 = "http://catalogurl"
	catalogTimeout             = 5
	catalogCategoryID          = "catalogCategoryID"
	catalogCategoryName        = "catalogCategoryName"
	catalogCategoryDescription = "Test catalogCategory description"
	catalogClientID            = "testCategoryApi"

	categoryResp = `{
	      "href": "%s",
	      "id": "%s",
	      "name": "%s",
	      "description": "%s"
	    }`

	roleImpl = `{
      "href": "%s",
      "id": "%s",
      "clientId": "%s",
      "categories": %s,
      "sourceInfo": {
        "href": "%s"
      }
    }`
	roleImplList = `{
	  "href": "%s",
	  "impls": %s
	}`

	categoryHref         = catalogURL + catalogCategoryPath + catalogCategoryID
	apiResponse_category = fmt.Sprintf(categoryResp, categoryHref, catalogCategoryID, catalogCategoryName, catalogCategoryDescription)

	role_href                  = kongURL + "/api/catalog/impl/" + catalogClientID
	role_source_info           = kongURL + "/" + catalogClientID + "/api/info"
	apiResponse_role           = fmt.Sprintf(roleImpl, role_href, catalogClientID, catalogClientID, "[\""+catalogCategoryID+"\"]", role_source_info)
	apiResponse_roleList_empty = fmt.Sprintf(roleImplList, kongURL+catalogAPIPathImpls, "[]")
	apiResponse_roleList_exist = fmt.Sprintf(roleImplList, kongURL+catalogAPIPathImpls, "["+apiResponse_role+"]")
)

func Test_NewCategory(t *testing.T) {
	//Mock category response
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	//query category - not found
	httpmock.RegisterResponder("GET", catalogURL+catalogCategoryPath+"/"+catalogCategoryID, httpmock.NewStringResponder(404, "Not Found- The ID of the category was not found."))
	var isPosted = false
	//register category
	httpmock.RegisterResponder("POST", catalogURL+catalogCategoriesPath, func(req *http.Request) (*http.Response, error) {
		isPosted = true
		return httpmock.NewStringResponse(http.StatusOK, apiResponse_category), nil
	})
	httpmock.RegisterResponder("GET", catalogURL+catalogAPIPathImpls, httpmock.NewStringResponder(200, apiResponse_roleList_empty))
	t.Log(apiResponse_role)
	httpmock.RegisterResponder("POST", catalogURL+catalogAPIPathImpls, func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, apiResponse_role), nil
	})

	result := RegisterAPIWithCatalog(kongURL, catalogURL, catalogTimeout, catalogCategoryID, catalogCategoryName, catalogCategoryDescription, catalogClientID)

	if !isPosted || !result {
		t.Fail()
	}
}
func Test_ExistingCategory(t *testing.T) {
	//Mock category response
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	//query category - not found
	httpmock.RegisterResponder("GET", catalogURL+catalogCategoryPath+"/"+catalogCategoryID, httpmock.NewStringResponder(200, apiResponse_category))
	var isPosted = false
	//register category
	httpmock.RegisterResponder("POST", catalogURL+catalogCategoriesPath, func(req *http.Request) (*http.Response, error) {
		t.Errorf("Should NOT register category")
		isPosted = true
		return httpmock.NewStringResponse(http.StatusOK, apiResponse_category), nil
	})
	httpmock.RegisterResponder("GET", catalogURL+catalogAPIPathImpls, httpmock.NewStringResponder(200, apiResponse_roleList_empty))
	httpmock.RegisterResponder("POST", catalogURL+catalogAPIPathImpls, func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, apiResponse_role), nil
	})

	result := RegisterAPIWithCatalog(kongURL, catalogURL, catalogTimeout, catalogCategoryID, catalogCategoryName, catalogCategoryDescription, catalogClientID)

	if isPosted || !result {
		t.Fail()
	}
}

func Test_RegisterAPINew(t *testing.T) {
	//Mock service registry response
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	//query category
	httpmock.RegisterResponder("GET", catalogURL+catalogCategoryPath+"/"+catalogCategoryID, httpmock.NewStringResponder(200, apiResponse_category))
	//query catalog impls - not exist
	t.Log(apiResponse_roleList_empty)
	httpmock.RegisterResponder("GET", catalogURL+catalogAPIPathImpls, httpmock.NewStringResponder(200, apiResponse_roleList_empty))
	//register catalog
	isPosted := false
	httpmock.RegisterResponder("POST", catalogURL+catalogAPIPathImpls, func(req *http.Request) (*http.Response, error) {
		isPosted = true
		t.Logf("%s", apiResponse_role)
		return httpmock.NewStringResponse(http.StatusOK, apiResponse_role), nil
	})
	httpmock.RegisterResponder("PUT", catalogURL+catalogAPIPathImpl+"/"+catalogClientID, func(req *http.Request) (*http.Response, error) {
		t.Errorf("Should NOT be PUT")
		return httpmock.NewStringResponse(http.StatusNotFound, ""), nil
	})

	result := RegisterAPIWithCatalog(kongURL, catalogURL, catalogTimeout, catalogCategoryID, catalogCategoryName, catalogCategoryDescription, catalogClientID)

	if !isPosted || !result {
		t.Fail()
	}
	//verify result
}

func Test_registerAPIUpdate_nochange(t *testing.T) {
	//Mock service registry response
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	//query category
	httpmock.RegisterResponder("GET", catalogURL+catalogCategoryPath+"/"+catalogCategoryID, httpmock.NewStringResponder(200, apiResponse_category))
	//query catalog impls - not exist
	httpmock.RegisterResponder("GET", catalogURL+catalogAPIPathImpls, httpmock.NewStringResponder(200, apiResponse_roleList_exist))
	//register catalog
	httpmock.RegisterResponder("POST", catalogURL+catalogAPIPathImpls, func(req *http.Request) (*http.Response, error) {
		t.Errorf("Should NOT be POST")
		return httpmock.NewStringResponse(http.StatusConflict, apiResponse_role), nil
	})
	isPut := false
	httpmock.RegisterResponder("PUT", catalogURL+catalogAPIPathImpl+"/"+catalogClientID, func(req *http.Request) (*http.Response, error) {
		isPut = true
		t.Errorf("Data no change, should NOT be PUT")
		return httpmock.NewStringResponse(http.StatusOK, apiResponse_role), nil
	})

	result := RegisterAPIWithCatalog(kongURL, catalogURL, catalogTimeout, catalogCategoryID, catalogCategoryName, catalogCategoryDescription, catalogClientID)

	if isPut || !result {
		t.Fail()
	}
}
func Test_registerAPIUpdate(t *testing.T) {
	//Mock service registry response
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	//query category
	httpmock.RegisterResponder("GET", catalogURL+catalogCategoryPath+"/"+catalogCategoryID, httpmock.NewStringResponder(200, apiResponse_category))
	//query catalog impls - not exist
	role_source_info = kongURL + "/" + catalogClientID + "/api/info_old"
	apiResponse_role = fmt.Sprintf(roleImpl, role_href, catalogClientID, catalogClientID, "[\""+catalogCategoryID+"\"]", role_source_info)
	apiResponse_roleList_exist = fmt.Sprintf(roleImplList, kongURL+catalogAPIPathImpls, "["+apiResponse_role+"]")
	httpmock.RegisterResponder("GET", catalogURL+catalogAPIPathImpls, httpmock.NewStringResponder(200, apiResponse_roleList_exist))
	//register catalog
	httpmock.RegisterResponder("POST", catalogURL+catalogAPIPathImpls, func(req *http.Request) (*http.Response, error) {
		t.Errorf("Should NOT be POST")
		return httpmock.NewStringResponse(http.StatusConflict, apiResponse_role), nil
	})
	isPut := false
	httpmock.RegisterResponder("PUT", catalogURL+catalogAPIPathImpl+"/"+catalogClientID, func(req *http.Request) (*http.Response, error) {
		isPut = true
		role_source_info = kongURL + "/" + catalogClientID + "/api/info"
		apiResponse_role = fmt.Sprintf(roleImpl, role_href, catalogClientID, catalogClientID, "[\""+catalogCategoryID+"\"]", role_source_info)
		return httpmock.NewStringResponse(http.StatusOK, apiResponse_role), nil
	})

	result := RegisterAPIWithCatalog(kongURL, catalogURL, catalogTimeout, catalogCategoryID, catalogCategoryName, catalogCategoryDescription, catalogClientID)

	if !isPut || !result {
		t.Fail()
	}
}

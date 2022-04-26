package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.ibm.com/cloud-sre/pnp-abstraction/testutils"
)

var (
	//	kongURL                    = "http://kongurl"
	//	catalogURL                 = "http://catalogurl"
	//	catalogTimeout             = 5
	//	catalogCategoryID          = "catalogCategoryID"
	//	catalogCategoryName        = "catalogCategoryName"
	//	catalogCategoryDescription = "Test catalogCategory description"
	//	catalogClientID            = "testCategoryApi"
	infoHandlerId = "test"
	infoResponse  = `{"categories":%s,"clientId":"%s","description":"%s","%s":{"href":"%s"}}`
)

const (
	apiInfoPath = "/api/info"
)

func createInfoHandler() *InfoHandler {
	infoHandler := new(InfoHandler)
	infoHandler.ClientID = catalogClientID
	infoHandler.CategoryID = catalogCategoryID
	infoHandler.Description = catalogCategoryDescription
	infoHandler.KongURL = kongURL
	infoHandler.APIRequestPathPrefix = catalogClientID

	var infoHandlerHrefs []InfoHandlerHref
	infoHandlerHrefs = append(infoHandlerHrefs, *createInfoHandlerHref(infoHandlerId, apiInfoPath))
	infoHandler.Hrefs = infoHandlerHrefs

	return infoHandler
}
func createInfoHandlerHref(id string, subpath string) *InfoHandlerHref {
	infoHandlerHref := new(InfoHandlerHref)
	infoHandlerHref.ID = id
	infoHandlerHref.SubPath = subpath
	return infoHandlerHref
}

func Test_info(t *testing.T) {
	validMethods := make(map[string]string)
	validMethods[http.MethodGet] = http.MethodGet
	infoHandler := createInfoHandler()
	infoEndpointHandler := CreateEndpointHandler(apiInfoPath, infoHandler.ServeInfo, validMethods)

	infoHref := kongURL + "/" + catalogClientID + apiInfoPath
	expectedResponse := fmt.Sprintf(infoResponse, "[\""+catalogCategoryID+"\"]", catalogClientID, catalogCategoryDescription, infoHandlerId, infoHref)

	testutils.HttpGet(t, apiInfoPath, expectedResponse, http.StatusOK, infoEndpointHandler.HandlerFunc)
}

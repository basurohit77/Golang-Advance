package pep

import (
	"net/http"
	"os"
	"testing"
)

func TestMocking_gopepExpectPermit(t *testing.T) {
	permitTest := true
	runGopepTest(t, permitTest)
}

func TestMocking_gopepExpectDeny(t *testing.T) {
	permitTest := false
	runGopepTest(t, permitTest)
}

func runGopepTest(t *testing.T, permitTest bool) {

	mockedPermitJSON := ` 
	{
		"cacheKeyPattern": {
		 "order": [
		  "subject",
		  "resource",
		  "action"
		 ],
		 "subject": [
		  [
		   "id"
		  ],
		  [
		   "id",
		   "scope"
		  ]
		 ],
		 "resource": [
		  [],
		  [
		   "serviceName"
		  ],
		  [
		   "serviceName",
		   "accountId"
		  ],
		  [
		   "serviceName",
		   "accountId",
		   "serviceInstance"
		  ],
		  [
		   "serviceName",
		   "accountId",
		   "serviceInstance",
		   "resourceType"
		  ],
		  [
		   "serviceName",
		   "accountId",
		   "serviceInstance",
		   "resourceType",
		   "resource"
		  ]
		 ]
		},
		"decisions": [
		 {
		  "decision": "Permit",
		  "obligation": {
		   "actions": [
			"gopep.books.write",
			"gopep.books.burn",
			"gopep.books.read",
			"gopep.books.eat"
		   ],
		   "environment": {
			"attributes": null
		   },
		   "maxCacheAgeSeconds": 600,
		   "subject": {
			"attributes": {
			 "id": "IBMid-3100015XDS"
			}
		   },
		   "resource": {
			"attributes": {
			 "accountId": "2c17c4e5587783961ce4a0aa415054e7",
			 "serviceName": "gopep"
			}
		   }
		  }
		 }
		]
	   }
	   `
	mockedDenyJSON := `{
		"decisions": [
			{
		  		"decision": "Deny"
			}
		]
	}`

	mockedContextDenyJSON := `{
		"decisions": [
			{
				"decision": "Deny",
				"reason":   "Context"
			}
		]
	}`

	resource := Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	cachedActions := []string{
		"gopep.books.write",
		"gopep.books.burn",
		"gopep.books.eat",
	}

	nonCachedAction := "gopep.books.give"

	responders := []JSONMockerFunc{
		func(req *http.Request) mockedResponse {
			return mockedResponse{
				json:       mockedPermitJSON,
				statusCode: 200,
			}
		},
		func(req *http.Request) mockedResponse {
			return mockedResponse{
				json:       mockedDenyJSON,
				statusCode: 200,
			}
		},
	}

	if !permitTest {
		responders = []JSONMockerFunc{
			func(req *http.Request) mockedResponse {
				return mockedResponse{
					json:       mockedDenyJSON,
					statusCode: 200,
				}
			},
			func(req *http.Request) mockedResponse {
				return mockedResponse{
					json:       mockedContextDenyJSON,
					statusCode: 200,
				}
			},
		}
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
	}

	// Running a series of tests
	if permitTest {
		runPermitTest(t, config, request, cachedActions, nonCachedAction)
	} else {
		runDenyTest(t, config, request, cachedActions, nonCachedAction)
	}
}

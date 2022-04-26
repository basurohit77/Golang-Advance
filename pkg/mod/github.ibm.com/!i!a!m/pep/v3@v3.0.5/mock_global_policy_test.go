package pep

import (
	"net/http"
	"os"
	"testing"
)

func TestMocking_globalPolicyExpectPermit(t *testing.T) {
	permitTest := true
	globalGodPolicy(t, permitTest)
}

func TestMocking_globalPolicyExpectDeny(t *testing.T) {
	permitTest := false
	globalGodPolicy(t, permitTest)
}

func globalGodPolicy(t *testing.T, permitTest bool) {
	mockedPermitJSON := `{
		"responses": [
			{
				"status": "200",
				"authorizationDecision": {
					"permitted": true,
					"obligation": {
						"actions": [
							"iam.policy.update",
							"iam.role.read",
							"iam.policy.read",
							"iam.policy.create",
							"iam.service.create",
							"iam.operator.update",
							"iam.delegationPolicy.create",
							"iam.policy.delete",
							"iam.delegationPolicy.update",
							"iam.operator.read",
							"iam.service.delete",
							"iam.service.read",
							"iam.role.assign"
						],
						"environment": {},
						"resources": [
							"crn:v1:staging:public:toolchain:us-south::::"
						],
						"maxCacheAgeSeconds": 600,
						"subject": {
							"attributes": {
								"scope": "ibm openid otc",
								"id": "IBMid-550005146S"
							}
						},
						"resource": {
							"attributes": {}
						}
					}
				}
			}
		],
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
		}
	}`

	mockedDenyJSON := `{
		"responses": [
			{
				"status": "200",
				"authorizationDecision": {
					"permitted": false
				}
			}
		]
	}`

	mockedContextDenyJSON := `{
		"responses": [
			{
				"status": "200",
				"authorizationDecision": {
					"permitted": false,
					"reason"   : "Context"
				}
			}
		]
	}`

	subject := Attributes{
		"scope": "ibm openid otc",
		"id":    "IBMid-550005146S",
	}
	resource := Attributes{
		"crn": "crn:v1:staging:public:toolchain:us-south::::",
	}

	action := "iam.policy.read"

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	cachedActions := []string{
		"iam.policy.update",
		"iam.role.read",
		"iam.policy.create",
		"iam.service.create",
		"iam.operator.update",
		"iam.delegationPolicy.create",
		"iam.policy.delete",
		"iam.delegationPolicy.update",
		"iam.operator.read",
		"iam.service.delete",
		"iam.service.read",
		"iam.role.assign",
	}

	nonCachedAction := "iam.policy.write"

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

	mocker := NewAuthzMocker(t, responders)

	config := Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
		HTTPClient:  mocker.httpClient,
	}

	if permitTest {
		runPermitTest(t, config, request, cachedActions, nonCachedAction)
	} else {
		runDenyTest(t, config, request, cachedActions, nonCachedAction)
	}
}

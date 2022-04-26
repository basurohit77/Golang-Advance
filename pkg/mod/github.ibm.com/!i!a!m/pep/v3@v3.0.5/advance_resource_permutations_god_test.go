package pep

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvanceResourcePermutations_godPolicyRequirement1(t *testing.T) {

	// Adjust this accordingly if you need to test the cache expiration
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getGodPolicyResponder(),
	}

	mocker := NewAuthzMocker(t, responders)

	config := Config{
		Environment:     Staging,
		APIKey:          os.Getenv("API_KEY"),
		LogLevel:        LevelError,
		CacheDefaultTTL: cacheTTL,
		HTTPClient:      mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}

		pdpCallCount := 0
		t.Run("initial call", func(t *testing.T) {
			trace := "txid-god-policy-req1-initial-call"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

		t.Run("same call", func(t *testing.T) {
			trace := "txid-god-policy-req1-same-call"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.True(t, response.Decisions[0].Cached, "should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")
			assert.Equal(t, pdpCallCount, mocker.counter(), "should not make a call to pdp")
		})

		t.Run("same call with different resource", func(t *testing.T) {
			trace := "txid-god-policy-req1-same-call-with-different-resource"

			resource := Attributes{
				"crn": "crn:v1:staging:public:toolchain:us-east::::",
			}
			request["resource"] = resource
			requests = &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.True(t, response.Decisions[0].Cached, "should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")
			assert.Equal(t, pdpCallCount, mocker.counter(), "should not make a call to pdp")
		})

		t.Run("same call with an applicable action", func(t *testing.T) {
			trace := "txid-god-policy-req1-same-call-with-applicable-action"

			request["action"] = "iam.policy.update"
			requests = &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.True(t, response.Decisions[0].Cached, "should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")
			assert.Equal(t, pdpCallCount, mocker.counter(), "should not make a call to pdp")
		})

		t.Run("same call with a non applicable action", func(t *testing.T) {
			trace := "txid-god-policy-req1-same-call-with-non-applicable-action"

			request["action"] = "non-application-action"
			requests = &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "should get a deny")
			assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

		t.Run("repeat the non applicable action call", func(t *testing.T) {
			trace := "txid-god-policy-req1-repeat-same-call-with-non-applicable-action"

			request["action"] = "non-application-action"
			requests = &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "should get a deny")
			assert.True(t, response.Decisions[0].Cached, "should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")
			assert.Equal(t, pdpCallCount, mocker.counter(), "should not make a call to pdp")
		})

	}

	runTest(t, config, getRequestForGodPolicy())
}

func TestAdvanceResourcePermutations_godPolicyRequirement2(t *testing.T) {

	// Adjust this accordingly if you need to test the cache expiration
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getGodPolicyResponder(),
	}

	mocker := NewAuthzMocker(t, responders)

	config := Config{
		Environment:     Staging,
		APIKey:          os.Getenv("API_KEY"),
		LogLevel:        LevelError,
		CacheDefaultTTL: cacheTTL,
		HTTPClient:      mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}

		pdpCallCount := 0

		t.Run("initial call", func(t *testing.T) {
			trace := "txid-god-policy-req2-initial-call"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

		t.Run("same call with non-applicable action", func(t *testing.T) {
			trace := "txid-god-policy-req2-same-call-non-applicable-action"

			request["action"] = "non-applicable-action"
			requests := &Requests{request}
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "should get a deny")
			assert.False(t, response.Decisions[0].Cached, "should not be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

		t.Run("same call with different subject id", func(t *testing.T) {
			trace := "txid-god-policy-req2-same-call-with-different-identity"

			request := getRequestForGodPolicy()
			request["subject"] = Attributes{
				"scope": "ibm openid otc",
				"id":    "IBMid-111111",
			}
			requests := &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "should get a deny")
			assert.False(t, response.Decisions[0].Cached, "should not be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

		t.Run("same call with different subject scope", func(t *testing.T) {
			trace := "txid-god-policy-req2-same-call-with-different-scope"

			request := getRequestForGodPolicy()
			request["subject"] = Attributes{
				"scope": "external openid otc",
				"id":    "IBMid-550005146S",
			}
			requests := &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "should get a deny")
			assert.False(t, response.Decisions[0].Cached, "should not be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

	}

	runTest(t, config, getRequestForGodPolicy())
}

func TestAdvanceResourcePermutations_ServiceGlobalPolicyRequirement3(t *testing.T) {

	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getGlobalServicePolicyResponder(),
	}

	mocker := NewBulkMocker(t, responders)

	config := Config{
		Environment:     Staging,
		APIKey:          os.Getenv("API_KEY"),
		LogLevel:        LevelError,
		CacheDefaultTTL: cacheTTL,
		HTTPClient:      mocker.httpClient,
	}

	runTest := func(t *testing.T, config Config, request Request) {

		err := Configure(&config)
		require.NoError(t, err)

		requests := &Requests{request}

		pdpCallCount := 0

		t.Run("initial call", func(t *testing.T) {
			trace := "txid-service-global-policy-req3-initial-call"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.False(t, response.Decisions[0].Cached, "should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

		t.Run("same request", func(t *testing.T) {
			trace := "txid-service-global-policy-req3-same-request"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.True(t, response.Decisions[0].Cached, "should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "should not expire since expired cache not enabled")

			assert.Equal(t, pdpCallCount, mocker.counter(), "should NOT make a call to pdp")
		})

		t.Run("same call with different applicable action", func(t *testing.T) {
			trace := "txid-service-global-policy-req3-different-applicable-action"

			// supported action (in cached list)
			request["action"] = "containers-kubernetes.group.create"
			requests := &Requests{request}
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.True(t, response.Decisions[0].Cached, "should be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			assert.Equal(t, pdpCallCount, mocker.counter(), "should NOT make a call to pdp")
		})

		t.Run("same call with different accountId,", func(t *testing.T) {
			trace := "txid-service-global-policy-req3-different-accountId"

			// same serviceName, different account
			request["resource"] = Attributes{
				"serviceName": "containers-kubernetes",
				"accountId":   "44444444444444444444444444444444",
			}

			requests := &Requests{request}
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.True(t, response.Decisions[0].Cached, "should be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			assert.Equal(t, pdpCallCount, mocker.counter(), "should NOT make a call to pdp")
		})

		t.Run("same call with different serviceInstance,", func(t *testing.T) {
			trace := "txid-service-global-policy-req3-different-serviceInstance"

			// same account and serviceName, new serviceInstance value
			request["resource"] = Attributes{
				"serviceName":     "containers-kubernetes",
				"accountId":       "4e507a1cd8624b028da5bf7380271c3c",
				"serviceInstance": "456789",
			}

			requests := &Requests{request}
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "should get a permit")
			assert.True(t, response.Decisions[0].Cached, "should be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			assert.Equal(t, pdpCallCount, mocker.counter(), "should NOT make a call to pdp")
		})

		t.Run("same call with different serviceName", func(t *testing.T) {
			trace := "txid-service-global-policy-req3-different-serviceName"

			// different serviceName
			request["resource"] = Attributes{
				"serviceName": "cloud-object-storage",
				"accountId":   "4e507a1cd8624b028da5bf7380271c3c",
			}

			requests := &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Cached, "should not be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

		t.Run("same call with unsupported action", func(t *testing.T) {
			trace := "txid-service-global-policy-req3-unsupported-action"

			// unsupported action (not in cache action list)
			request["action"] = "containers-kubernetes.something.unsupported"
			requests := &Requests{request}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Cached, "should not be cached")
			assert.False(t, response.Decisions[0].Expired, "should not expired")

			pdpCallCount = pdpCallCount + 1
			assert.Equal(t, pdpCallCount, mocker.counter(), "should make a call to pdp")
		})

	}

	runTest(t, config, getRequestForGlobalServicePolicy())
}

type pdpReq struct {
	httpReq *http.Request
}

func (pr pdpReq) isValid() bool {
	if pr.httpReq == nil {
		return false
	}

	if pr.httpReq.ContentLength == 0 {
		return false
	}
	return true
}

func (pr pdpReq) hasAnyAction(actions []string) bool {

	if !pr.isValid() {
		return false
	}

	if len(actions) < 1 {
		return false
	}

	var body interface{}

	buf := pr.getBody()
	err := json.Unmarshal(buf, &body)

	if err != nil {
		return false
	}

	// Is it authz?
	if authz, ok := body.([]interface{}); ok && len(authz) > 0 {
		if firstRequest, ok := authz[0].(map[string]interface{}); ok {
			for _, a := range actions {
				if a == firstRequest["action"] {
					return true
				}
			}
			return false

		}
	}

	// Is it bulk?
	if bulk, ok := body.(map[string]interface{}); ok {
		if action, ok := bulk["action"]; ok {
			for _, a := range actions {
				if a == action {
					return true
				}
			}
		}
	}
	return false
}

func (pr pdpReq) hasSubject(scope string, id string) bool {

	if !pr.isValid() {
		return false
	}

	var body interface{}

	buf := pr.getBody()
	err := json.Unmarshal(buf, &body)

	if err != nil {
		return false
	}

	// is it authz?
	if authz, ok := body.([]interface{}); ok && len(authz) > 0 {
		if firstRequest, ok := authz[0].(map[string]interface{}); ok {
			if subject, hasSubject := firstRequest["subject"].(map[string]interface{}); hasSubject {
				if attr, hasAttr := subject["attributes"].(map[string]interface{}); hasAttr {
					return attr["id"] == id && attr["scope"] == scope
				}
			}

		}

	}

	// is it bulk?
	if bulk, ok := body.(map[string]interface{}); ok {
		if subject, ok := bulk["subject"].(map[string]interface{}); ok {
			if attr, ok := subject["attributes"].(map[string]interface{}); ok {
				return attr["id"] == id && attr["scope"] == scope
			}
		}
	}

	return false
}

func (pr pdpReq) getBody() []byte {
	if pr.httpReq == nil {
		return []byte{}
	}
	buf, _ := ioutil.ReadAll(pr.httpReq.Body)
	pr.httpReq.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	return bytes.NewBuffer(buf).Bytes()
}

func getGodPolicyResponder() JSONMockerFunc {

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

	cacheableActions := []string{
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
		"iam.role.assign",
	}

	return func(req *http.Request) mockedResponse {

		mockedJSON := mockedDenyJSON

		r := pdpReq{req}

		if r.hasAnyAction(cacheableActions) && r.hasSubject("ibm openid otc", "IBMid-550005146S") {
			mockedJSON = mockedPermitJSON
		}
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}

	}

}

func getRequestForGodPolicy() Request {

	subject := Attributes{
		"scope": "ibm openid otc",
		"id":    "IBMid-550005146S",
	}
	resource := Attributes{
		"crn": "crn:v1:staging:public:gopep:us-south::::",
	}

	action := "iam.policy.read"

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	return request
}

func getGlobalServicePolicyResponder() JSONMockerFunc {

	mockedPermitJSON := `{
		"decisions": [
			{
				"decision": "Permit",
				"obligation": {
					"actions": [
						"containers-kubernetes.key.retrieve_all",
						"containers-kubernetes.group.update",
						"containers-kubernetes.instance.replay",
						"containers-kubernetes.quota.update",
						"containers-kubernetes.instance.update_state",
						"containers-kubernetes.cluster.read",
						"containers-kubernetes.instance.create_id",
						"containers-kubernetes.alias.retrieve_all",
						"containers-kubernetes.instance.retrieve_all",
						"containers-kubernetes.instance.update_sync",
						"containers-kubernetes.instance.retrieve_state",
						"containers-kubernetes.group.retrieve",
						"containers-kubernetes.group.update_quota",
						"containers-kubernetes.quota.retrieve",
						"containers-kubernetes.group.create",
						"containers-kubernetes.broker.retrieve_all",
						"containers-kubernetes.instance.retrieve_by_crn_segments",
						"containers-kubernetes.quota.create",
						"containers-kubernetes.cluster.update",
						"containers-kubernetes.instance.delete_force",
						"containers-kubernetes.instance.retrieve_history",
						"containers-kubernetes.binding.retrieve_all"
					],
					"maxCacheAgeSeconds": 600,
					"subject": {
						"attributes": {
							"scope": "ibm openid otc",
							"id": "iam-ServiceId-023ef47c-1d54-4088-9a46-f5bc736ef8d2"
						}
					},
					"resource": {
						"attributes": {
							"serviceName": "containers-kubernetes"
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
			"decisions": [
				{
					"decision": "Deny"
				}
			]
			}`

	cacheableActions := []string{
		"containers-kubernetes.key.retrieve_all",
		"containers-kubernetes.group.update",
		"containers-kubernetes.instance.replay",
		"containers-kubernetes.quota.update",
		"containers-kubernetes.instance.update_state",
		"containers-kubernetes.cluster.read",
		"containers-kubernetes.instance.create_id",
		"containers-kubernetes.alias.retrieve_all",
		"containers-kubernetes.instance.retrieve_all",
		"containers-kubernetes.instance.update_sync",
		"containers-kubernetes.instance.retrieve_state",
		"containers-kubernetes.group.retrieve",
		"containers-kubernetes.group.update_quota",
		"containers-kubernetes.quota.retrieve",
		"containers-kubernetes.group.create",
		"containers-kubernetes.broker.retrieve_all",
		"containers-kubernetes.instance.retrieve_by_crn_segments",
		"containers-kubernetes.quota.create",
		"containers-kubernetes.cluster.update",
		"containers-kubernetes.instance.delete_force",
		"containers-kubernetes.instance.retrieve_history",
		"containers-kubernetes.binding.retrieve_all",
	}

	return func(req *http.Request) mockedResponse {

		mockedJSON := mockedDenyJSON

		r := pdpReq{req}

		if r.hasAnyAction(cacheableActions) && r.hasSubject("ibm openid otc", "iam-ServiceId-023ef47c-1d54-4088-9a46-f5bc736ef8d2") {
			mockedJSON = mockedPermitJSON
		}

		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}

	}
}

func getRequestForGlobalServicePolicy() Request {

	subject := Attributes{
		"scope": "ibm openid otc",
		"id":    "iam-ServiceId-023ef47c-1d54-4088-9a46-f5bc736ef8d2",
	}
	resource := Attributes{
		"serviceName": "containers-kubernetes",
		"accountId":   "4e507a1cd8624b028da5bf7380271c3c",
	}

	action := "containers-kubernetes.cluster.read"

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  subject,
	}

	return request
}

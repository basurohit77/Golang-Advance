package pep

import (
	io "io/ioutil"
	"os"
	"testing"
)

func TestRequests_statsFuctions(t *testing.T) {

	tooLargeRequest, _ := generateBulkRequests(1001, false, false)
	largeRequest, largeResponse := generateBulkRequests(1000, false, false)
	largeResponse.Trace = "txid-large-request"

	authzTestLarge := authzTestRequest{
		useDifferentAccount: true,
		requestCount:        98,
		trace:               "txid-large-authz",
	}
	largeAuthzRequest, largeAuthzResp, _ := authzTestLarge.generateAuthzRequest()

	authzAccount := authzTestRequest{
		useDifferentAccount: true,
		requestCount:        98,
		trace:               "txid-authz-different-accounts",
	}
	reqAccount, respAccount, _ := authzAccount.generateAuthzRequest()
	cachedRespAccount := getExpectedCachedResponse(respAccount)

	authzServiceName := authzTestRequest{
		useDifferentServiceName: true,
		requestCount:            98,
		trace:                   "txid-authz-different-servicenames",
	}
	reqServiceName, respServiceName, errServiceName := authzServiceName.generateAuthzRequest()

	authzSubject := authzTestRequest{
		useDifferentSubject: true,
		requestCount:        98,
		trace:               "txid-authz-different-subjects",
	}
	// Error is ignored since there is no error
	reqSubject, respSubject, _ := authzSubject.generateAuthzRequest()
	cachedRespSubject := getExpectedCachedResponse(respSubject)

	authzAction := authzTestRequest{
		useDifferentAction: true,
		requestCount:       98,
		trace:              "txid-authz-different-actions",
	}
	// Error is ignored since there is no error
	reqAction, respAction, _ := authzAction.generateAuthzRequest()
	cachedRespAction := getExpectedCachedResponse(respAction)

	tests := []isAuthorizedTestTemplate{
		{
			name: "permitted",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "cached permitted",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-cache"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-cache",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "denied",
			reqs: &deniedRequest,
			args: isAuthorizedTraceArgs{"txid-denied-1"},
			want: AuthzResponse{
				Trace: "txid-denied-1",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
						Reason:     DenyReasonIAM,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "cached denied",
			reqs: &deniedRequest,
			args: isAuthorizedTraceArgs{"txid-denied-from-cached"},
			want: AuthzResponse{
				Trace: "txid-denied-from-cached",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
						Reason:     DenyReasonIAM,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "cached no transaction id",
			reqs: &deniedRequest,
			args: isAuthorizedTraceArgs{""},
			want: AuthzResponse{
				Trace: "",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
						Reason:     DenyReasonIAM,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "error due to wrong servicename",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep1",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			args: isAuthorizedTraceArgs{"txid-denied-due-to-wrong-servicename"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  403,
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz/bulk",
				Trace:       "txid-denied-due-to-wrong-servicename",
			},
		},
		{
			name: "denied due to wrong accountId",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "5c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			args: isAuthorizedTraceArgs{"txid-denied-due-to-wrong-account-id"},
			want: AuthzResponse{
				Trace: "txid-denied-due-to-wrong-account-id",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
						Reason:     DenyReasonIAM,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "denied due to wrong subject",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "3c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS1",
					},
				},
			},
			args: isAuthorizedTraceArgs{"txid-denied-due-to-wrong-subject"},
			want: AuthzResponse{
				Trace: "txid-denied-due-to-wrong-subject",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
						Reason:     DenyReasonIAM,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "denied due to wrong action",
			reqs: &Requests{
				{
					"action": "gopep.books.read1",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "3c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			args: isAuthorizedTraceArgs{"txid-denied-due-to-wrong-action"},
			want: AuthzResponse{
				Trace: "txid-denied-due-to-wrong-action",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
						Reason:     DenyReasonIAM,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "error due to invalid request",
			reqs: &Requests{},
			args: isAuthorizedTraceArgs{"txid-denied-due-to-invalid-request"},
			want: AuthzResponse{},
			wantErr: &InternalError{
				Message: "Invalid request: The minimum number of requests is 1",
				Trace:   "txid-denied-due-to-invalid-request",
			},
		},
		{
			name: "multiple requests",
			reqs: &multipleRequests,
			args: isAuthorizedTraceArgs{"txid-multiple-requests"},
			want: AuthzResponse{
				Trace: "txid-multiple-requests",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
					},
					{
						Permitted:  true,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
					},
					{
						Permitted:  true,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "multiple v2 authz requests with error in request #0",
			reqs: &multipleRequestsWithErrorInRequest0,
			args: isAuthorizedTraceArgs{"txid-multiple-requests-with-error-0"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  403,
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz",
				Trace:       "txid-multiple-requests-with-error-0",
				Message:     "Request #0 contains error: unauthorized to make this request",
			},
		},
		{
			name: "multiple v2 authz requests with error in request #1",
			reqs: &multipleRequestsWithErrorInRequest1,
			args: isAuthorizedTraceArgs{"txid-multiple-requests-with-error-1"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  403,
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz",
				Trace:       "txid-multiple-requests-with-error-1",
				Message:     "Request #1 contains error: unauthorized to make this request",
			},
		},
		{
			name: "Request with more than 1k sub requests should return an error",
			reqs: &tooLargeRequest,
			args: isAuthorizedTraceArgs{"txid-too-large-request"},
			want: AuthzResponse{},
			wantErr: &InternalError{
				Trace:   "txid-too-large-request",
				Message: "Invalid request: The maximum number of requests allowed is 1k",
			},
		},
		{
			name:    "Request with 1k sub requests should return the response",
			reqs:    &largeRequest,
			args:    isAuthorizedTraceArgs{"txid-large-request"},
			want:    largeResponse,
			wantErr: nil,
		},
		{
			name:    "large authz request",
			reqs:    &largeAuthzRequest,
			args:    isAuthorizedTraceArgs{largeAuthzResp.Trace},
			want:    largeAuthzResp,
			wantErr: nil,
		},
		{
			name:    "authz with different accounts",
			reqs:    &reqAccount,
			args:    isAuthorizedTraceArgs{respAccount.Trace},
			want:    respAccount,
			wantErr: nil,
		},
		{
			// Note this test depends on a cached response of "authz with different accounts"
			name:    "cached authz with different accounts",
			reqs:    &reqAccount,
			args:    isAuthorizedTraceArgs{respAccount.Trace},
			want:    cachedRespAccount,
			wantErr: nil,
		},
		{
			// Note this test depends on a cached response of "authz with different accounts"
			name:    "cached authz with different accounts without trace",
			reqs:    &reqAccount,
			args:    isAuthorizedTraceArgs{""},
			want:    cachedRespAccount,
			wantErr: nil,
		},
		{
			name:    "authz with different service names",
			reqs:    &reqServiceName,
			args:    isAuthorizedTraceArgs{respServiceName.Trace},
			want:    AuthzResponse{},
			wantErr: &errServiceName,
		},
		{
			name:    "cached authz with different service names",
			reqs:    &reqServiceName,
			args:    isAuthorizedTraceArgs{respServiceName.Trace},
			want:    AuthzResponse{},
			wantErr: &errServiceName,
		},
		{
			name:    "authz with different subjects",
			reqs:    &reqSubject,
			args:    isAuthorizedTraceArgs{respSubject.Trace},
			want:    respSubject,
			wantErr: nil,
		},
		{
			name:    "cached authz with different subjects",
			reqs:    &reqSubject,
			args:    isAuthorizedTraceArgs{respSubject.Trace},
			want:    cachedRespSubject,
			wantErr: nil,
		},
		{
			name:    "authz with different actions",
			reqs:    &reqAction,
			args:    isAuthorizedTraceArgs{respAction.Trace},
			want:    respAction,
			wantErr: nil,
		},
		{
			name:    "cached authz with different actions",
			reqs:    &reqAction,
			args:    isAuthorizedTraceArgs{respAction.Trace},
			want:    cachedRespAction,
			wantErr: nil,
		},
		{
			name:    "cached 1authz with different actions",
			reqs:    &reqAction,
			args:    isAuthorizedTraceArgs{respAction.Trace},
			want:    cachedRespAction,
			wantErr: nil,
		},
		{
			name: "single invalid request",
			reqs: &Requests{
				{},
			},
			args: isAuthorizedTraceArgs{"txid-single-invalid-request"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  400,
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz",
				Trace:       "txid-single-invalid-request",
				Message:     "Request #0 contains error: An action needs to be provided for a valid request.",
			},
		},
		{
			name: "multiple requests no action",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{},
			},
			args: isAuthorizedTraceArgs{"txid-multiple-requests-no-action"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  400,
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz",
				Trace:       "txid-multiple-requests-no-action",
				Message:     "Request #1 contains error: An action needs to be provided for a valid request.",
			},
		},
		{
			name: "multiple requests no subject",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
				},
			},
			args: isAuthorizedTraceArgs{"txid-multiple-requests-no-subject"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  400,
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz",
				Trace:       "txid-multiple-requests-no-subject",
				Message:     "Request #1 contains error: A subject needs to be provided for a valid request.",
			},
		},
		{
			name: "multiple requests no resource",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			args: isAuthorizedTraceArgs{"txid-multiple-requests-no-resource"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  400,
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz",
				Trace:       "txid-multiple-requests-no-resource",
				Message:     "Request #1 contains error: A resource needs to be provided for a valid request.",
			},
		},
	}

	pepConfig := &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 48,
		LogOutput:         io.Discard,
	}

	err := Configure(pepConfig)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	for _, tt := range tests {
		authzRunner(tt, t)
	}

	conf := GetConfig().(*Config)
	badRequest, err := conf.Statistics.GetStat("status-code-400")

	if err != nil {
		t.Errorf("There should be no error getting statistics")
	}
	if badRequest != 1 {
		t.Errorf("Status code status-code-400 returned %d, want %d", badRequest, 1)
	}

	forbidden, err := conf.Statistics.GetStat("status-code-403")

	if err != nil {
		t.Errorf("There should be no error getting statistics")
	}
	if forbidden != 1 {
		t.Errorf("Status code status-code-403 returned %d, want %d", forbidden, 1)
	}

	partialContentCount, err := conf.Statistics.GetStat("status-code-206")

	if err != nil {
		t.Errorf("There should be no error getting statistics")
	}
	if partialContentCount != 7 {
		t.Errorf("Status code status-code-206 returned %d, want %d", partialContentCount, 7)
	}
}

func Test_statsStartup(t *testing.T) {
	pepConfig := &Config{
		Environment: Staging,
		//APIKey: os.Getenv("API_KEY"), this will not cause a misconfigured PEP that returns an error
	}

	err := Configure(pepConfig)

	// some configurations might do this, so we want to protect ourselves against possible nil pointer exceptions
	go GetStatistics()

	if err != nil {
		t.Errorf("pep.Configure should not fail for missing API Key")
	}
}

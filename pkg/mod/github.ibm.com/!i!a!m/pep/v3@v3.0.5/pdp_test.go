package pep

import (
	"bytes"
	"encoding/json"
	"fmt"
	io "io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/IAM/pep/v3/cache"
)

type isAuthorizedTraceArgs struct {
	trace string
}
type isAuthorizedTestTemplate struct {
	name    string
	reqs    *Requests
	args    isAuthorizedTraceArgs
	want    AuthzResponse
	wantErr error
}

var permittedRequest = Requests{
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
}

// this request should get a deny due to the wrong account ID
var deniedRequest = Requests{
	{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book1",
			"accountId":       "3c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
}

var multipleRequests = Requests{
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
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book2",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
	{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book3",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
}

var multipleRequestsWithErrorInRequest0 = Requests{
	{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "cloudant",
			"serviceInstance": "chatko-book1",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
	{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book2",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
	{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book3",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
}

var multipleRequestsWithErrorInRequest1 = Requests{
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
		"resource": Attributes{
			"serviceName":     "cloudant",
			"serviceInstance": "chatko-book2",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
	{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book3",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	},
}

//test CRN token
/* #nosec G101 */
var crnToken = "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJjcm4tY3JuOnYxOmJsdWVtaXg6cHVibGljOmdvcGVwOjphLzJjMTdjNGU1NTg3NzgzOTYxY2U0YTBhYTQxNTA1NGU3OmdvcGVwMTIzOjoiLCJpZCI6ImNybi1jcm46djE6Ymx1ZW1peDpwdWJsaWM6Z29wZXA6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6Z29wZXAxMjM6OiIsInJlYWxtaWQiOiJjcm4iLCJqdGkiOiIwYzFkNTZjZS1mOGZkLTQ3NjUtODBjYy03N2E3ZTk2YmY4MDEiLCJpZGVudGlmaWVyIjoiY3JuOnYxOmJsdWVtaXg6cHVibGljOmdvcGVwOjphLzJjMTdjNGU1NTg3NzgzOTYxY2U0YTBhYTQxNTA1NGU3OmdvcGVwMTIzOjoiLCJzdWIiOiJjcm46djE6Ymx1ZW1peDpwdWJsaWM6Z29wZXA6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6Z29wZXAxMjM6OiIsInN1Yl90eXBlIjoiQ1JOIiwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTciLCJmcm96ZW4iOnRydWV9LCJpYXQiOjE1OTg5MDQ1ODIsImV4cCI6MTU5ODkwODEwNSwiaXNzIjoiaHR0cHM6Ly9pYW0uc3RhZ2UxLmJsdWVtaXgubmV0L2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6aWFtLWF1dGh6Iiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFjciI6MCwiYW1yIjpbXX0.bDndRAzcNaVdG1_g-UH03_kUAty4qOmlE3HkIswvrKEpWo7po59fn8zf89oIv3B-pXpQ9kwO3LxLJ-o3CF45D8_xnyUHFl1JWec7NCUJeTtcwbGshjT25x62fKuHIknGfijNqsRQ807hdQmv8RLWLzo62nIee3TKN0YHvR7ju3ctTa_C5Xv3O72SWRXQ-MqoPrfD9C2TJq9R-UH-r7FQ0URDnyIHX_2q0joUNAB35ujle0uBuoOJSMfizFGdKigIbA3R5qDqC2qHE2NhPbJtGTZjpj3jPPiCTepziS_LhgAoGFmSjcuXVmEzSCkJyCfqm9zXudPD08_UJM61dcfptQ"

/* #nosec G101 */
var crnTokenNoRealm = "eyJraWQiOiIyMDIwMDgyMjE2NTYiLCJhbGciOiJIUzI1NiJ9.eyJpYW1faWQiOiJjcm46djE6Ymx1ZW1peDpwdWJsaWM6Y29tbWl0bWVudC1kZXZpY2U6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6aW5zdGFuY2UxMjM0NTo6IiwiaWQiOiJjcm46djE6Ymx1ZW1peDpwdWJsaWM6Y29tbWl0bWVudC1kZXZpY2U6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6aW5zdGFuY2UxMjM0NTo6IiwicmVhbG1pZCI6ImNybiIsImlkZW50aWZpZXIiOiJjcm46djE6Ymx1ZW1peDpwdWJsaWM6Y29tbWl0bWVudC1kZXZpY2U6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6aW5zdGFuY2UxMjM0NTo6Iiwic3ViIjoiY3JuOnYxOmJsdWVtaXg6cHVibGljOmNvbW1pdG1lbnQtZGV2aWNlOjphLzJjMTdjNGU1NTg3NzgzOTYxY2U0YTBhYTQxNTA1NGU3Omluc3RhbmNlMTIzNDU6OiIsInN1Yl90eXBlIjoiQ1JOIiwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTciLCJmcm96ZW4iOnRydWV9LCJpYXQiOjE1OTgzMDU1NDUsImV4cCI6MTU5ODMwOTEwNSwiaXNzIjoiaHR0cHM6Ly9pYW0uc3RhZ2UxLmJsdWVtaXgubmV0L2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6aWFtLWF1dGh6Iiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFjciI6MCwiYW1yIjpbXX0.gk3V8goRCkvselH88rmljsGp3OBaAPgU46dQoNPNmlU"

//test user token
/* #nosec G101 */
var userToken = "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwiaWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiNzkyMjg4NTAtZWM1Yy00MWJjLWI4ZjItN2RkMzYzNWFhNjI4IiwiaWRlbnRpZmllciI6IjI3MDAwM0dVU1giLCJnaXZlbl9uYW1lIjoiQWxleCIsImZhbWlseV9uYW1lIjoiSHVkaWNpIiwibmFtZSI6IkFsZXggSHVkaWNpIiwiZW1haWwiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJzdWIiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiIyYzE3YzRlNTU4Nzc4Mzk2MWNlNGEwYWE0MTUwNTRlNyJ9LCJpYXQiOjE1OTg5MDE2MTAsImV4cCI6MTU5ODkwNTIxMCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6cGFzc2NvZGUiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.fMCKyCDbmff8DtlFhbpw84wvZyBSO3k-zAKN494AkN7hvQtHROWPX-TmLpgA170xhYHTNNariOE4c7JitcPdjHinbF6Rro8yUL2HPUZ6To_T8u0xw3yXhR8UhLgHiomUTy2qLY2TK_rUz_R7_yQRUD-91CJwR8iUH8NP4lX0bCQC9o7pxFv8XMNkNSyoyrZYXw2Wfpu0JUK3iZp0Iq9RnAS_7VYJ-Na11tWOugvNvk-PdwWE3OxbyIGL2FDwWbL9kVIv3_oA5mvTI7Zw9vpODaoObstCWFALmGUt7g0t-nE1BbUWOgB5ajkCh1judvKv5tId7OHicIbBVY0bIK3yUw"

//test service token
/* #nosec G101 */
var serviceToken = "eyJraWQiOiIyMDE3MDkxOS0xOTowMDowMCIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLWMzM2FkNzJmLTE1NDYtNGZkNC04ZTk0LTM0MThlZDBmYjZlNCIsImlkIjoiaWFtLVNlcnZpY2VJZC1jMzNhZDcyZi0xNTQ2LTRmZDQtOGU5NC0zNDE4ZWQwZmI2ZTQiLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC1jMzNhZDcyZi0xNTQ2LTRmZDQtOGU5NC0zNDE4ZWQwZmI2ZTQiLCJzdWIiOiJTZXJ2aWNlSWQtYzMzYWQ3MmYtMTU0Ni00ZmQ0LThlOTQtMzQxOGVkMGZiNmU0Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7ImJzcyI6IjU4Y2Y5M2JmYWIzMzJjODA1ZjU4NzgxMzNhYmI0YTFmIn0sImlhdCI6MTUwOTExNDM0MCwiZXhwIjoxNTA5MTE3OTQwLCJpc3MiOiJodHRwczovL2lhbS5zdGFnZTEubmcuYmx1ZW1peC5uZXQvb2lkYy90b2tlbiIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFsZyI6IkhTMjU2In0.BWSrvt4fsHaWMIy5csdVeax4X1IvTchHzo-2ORbV8bSKKXT0cdgcLIHBYPEi0fLAdQEYJ8cM7ZkJMetoxIpVnIRh2Iiaim2ypDuKTFIjsm7sW3WBN6sMzIhIYuII68IHtJVofQ09HUNwTed61BDryOchvzJ6sZnbo3NAW0atH8r2udHz1uLtpg-ITdg_zIRvp5PZxJKmPHkKxEUvWPCeGJldkZPgahtYXhsPq_HA9NEgZCJANOdAQCm1qoCyZ-HDngysbu9SYopDKzTUf0by6CkkLtIjzg2LabtxTB_1n72CWO5GRA1q5xA70RIorvYap9MvsY7obWF310LYEXUj2A"

/* #nosec G101 */
var userToken2 = "eyJraWQiOiIyMDIwMTIxMDE0NDkiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0zMTAwMDE1WERTIiwiaWQiOiJJQk1pZC0zMTAwMDE1WERTIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiOGQzYjY1YzctMjk3MC00M2QxLTgwNDQtM2U1NmE4NzhhNTk4IiwiaWRlbnRpZmllciI6IjMxMDAwMTVYRFMiLCJnaXZlbl9uYW1lIjoiQ2hyaXMiLCJmYW1pbHlfbmFtZSI6IkhhdGtvIiwibmFtZSI6IkNocmlzIEhhdGtvIiwiZW1haWwiOiJjaGF0a29AY2EuaWJtLmNvbSIsInN1YiI6ImNoYXRrb0BjYS5pYm0uY29tIiwiaWF0IjoxNjA3NzQ0MDAyLCJleHAiOjE2MDc3NDc2MDIsImlzcyI6Imh0dHBzOi8vaWFtLnRlc3QuY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOnBhc3Njb2RlIiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiYngiLCJhY3IiOjEsImFtciI6WyJwd2QiXX0.JWZ8kKyOvv_JxExX1kLA8G2VOWfpJB4dd3BBvWvn2BFkFcQRF2SElhWvfsCY9TSr7dx-7M4_-4a10R7sMtInNF3d6Kq5ieF5ltcsOtjYmE4q7IuVclw4tOrvmInmj7o1HiM3Fk1VidBtfh2SxDs-JKbO0rATtMVmRFPHUPu6lDq_HzSu4pF5olKWUQQ1-DhbtuOEYcicvqrQ4qgealK2bytWOHHTbCNvvFzMsYx0OZO5nFW0rzYtkncaqfacAhe8PwkFQoUGn5Te67-9qO1WqA7q920YlZWSFjgmaxqvWQ7Qfx0akkI0z9G7L5p34WxDOcY35Cd91ywSYyDL1rgKvw"

func generateBulkRequests(requestCount int, cached bool, expired bool) (requests Requests, response AuthzResponse) {

	decision := Decision{
		Permitted:  true,
		Cached:     cached,
		Expired:    expired,
		RetryCount: 0,
	}

	for i := 0; i < requestCount; i++ {

		request := Attributes{
			"action": "gopep.books.read",
			"resource": Attributes{
				"serviceName":     "gopep",
				"serviceInstance": fmt.Sprintf("chatko-book%d", i),
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
			},
			"subject": Attributes{
				"id": "IBMid-3100015XDS",
			},
		}

		requests = append(requests, request)
		response.Decisions = append(response.Decisions, decision)
	}
	return
}

type authzTestRequest struct {
	requestCount            int
	trace                   string
	useDifferentSubject     bool
	useDifferentAction      bool
	useDifferentServiceName bool
	useDifferentAccount     bool
}

// Genereate requests for authz i.e. requests that do not pass any of the bulk request conditions below
// - No unique subject
// - No unique action
// - No unique service name
// - No unique accountId
// - No common attributes
func (r authzTestRequest) generateAuthzRequest() (requests Requests, response AuthzResponse, apiError APIError) {

	response.Trace = r.trace

	decisionTemplate := Decision{
		Permitted:  true,
		Cached:     false,
		Expired:    false,
		RetryCount: 0,
	}

	var subjectIndex = -1
	var actionIndex = -1
	var serviceNameIndex = -1
	var accountIndex = -1

	if r.requestCount == 0 {
		r.requestCount = 2
	}

	if r.useDifferentSubject {
		/* #nosec G404 */
		subjectIndex = rand.Intn(r.requestCount)
	}

	if r.useDifferentAction {
		/* #nosec G404 */
		actionIndex = rand.Intn(r.requestCount)
	}

	if r.useDifferentServiceName {
		/* #nosec G404 */
		serviceNameIndex = rand.Intn(r.requestCount)
		apiError.StatusCode = 403
		apiError.Trace = r.trace
		apiError.Message = fmt.Sprintf("Request #%d contains error: unauthorized to make this request", serviceNameIndex)
		apiError.EndpointURI = "https://iam.test.cloud.ibm.com/v2/authz"
	}

	if r.useDifferentAccount {
		/* #nosec G404 */
		accountIndex = rand.Intn(r.requestCount)
	}

	for i := 0; i < r.requestCount; i++ {

		decision := decisionTemplate

		action := "gopep.books.read"
		if r.useDifferentAction && i == actionIndex {
			// action that has no mapping
			action = "gopep.books.share"
			decision.Permitted = false
			decision.Reason = DenyReasonIAM
		}

		subject := "IBMid-3100015XDS"
		if r.useDifferentSubject && i == subjectIndex {
			// creates 1 entry in entire list that is not permitted
			subject = "IBMid-3100015XDT"
			decision.Permitted = false
			decision.Reason = DenyReasonIAM
		}

		serviceName := "gopep"
		if r.useDifferentServiceName && i == serviceNameIndex {
			//This should return a 403 which has precedence over other responses.
			serviceName = "cloudant"
			decision.Permitted = false
			decision.Reason = DenyReasonIAM
		}

		accountID := "2c17c4e5587783961ce4a0aa415054e7"
		if r.useDifferentAccount && i == accountIndex {
			accountID = "3c17c4e5587783961ce4a0aa415054e7"
			decision.Permitted = false
			decision.Reason = DenyReasonIAM
		}

		request := Attributes{
			"action": action,
			"resource": Attributes{
				"serviceName":     serviceName,
				"serviceInstance": fmt.Sprintf("chatko-book%d", i),
				"accountId":       accountID,
			},
			"subject": Attributes{
				"id": subject,
			},
		}

		requests = append(requests, request)
		response.Decisions = append(response.Decisions, decision)
	}
	return
}

func getExpectedCachedResponse(resp AuthzResponse) (cachedResp AuthzResponse) {
	cachedResp.Trace = resp.Trace
	for _, decision := range resp.Decisions {
		decision.Cached = true
		cachedResp.Decisions = append(cachedResp.Decisions, decision)
	}
	return
}

func TestRequests_isAuthorized(t *testing.T) {

	tooLargeRequest, _ := generateBulkRequests(1001, false, false)
	largeRequest, largeResponse := generateBulkRequests(1000, false, false)
	largeResponse.Trace = "txid-large-request"

	authzTestLarge := authzTestRequest{
		useDifferentAccount: true,
		requestCount:        98,
		trace:               "txid-large-authz",
	}
	largeAuthzRequest, largeAuthzResp, _ := authzTestLarge.generateAuthzRequest()

	authz101TestLarge := authzTestRequest{
		useDifferentAccount: true,
		requestCount:        101,
		trace:               "txid-large-authz-morethan-100",
	}
	large101AuthzRequest, large101AuthzResp, _ := authz101TestLarge.generateAuthzRequest()

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

	permittedInsertRequest := Attributes{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book999",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	}

	deniedInsertRequest := Attributes{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book999",
			"accountId":       "3c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDT",
		},
	}

	partialCachedReqSubject := make(Requests, len(reqSubject))
	copy(partialCachedReqSubject, reqSubject)
	partialCachedRespSubject := AuthzResponse{
		Trace:     "txid-authz-different-subjects-new-items",
		Decisions: make([]Decision, len(respSubject.Decisions)),
	}

	copy(partialCachedRespSubject.Decisions, respSubject.Decisions)

	// Inserting a new permitted request into the request slice
	partialCachedReqSubject[authzSubject.requestCount-1] = permittedInsertRequest
	// Inserting a new denied request into the request slice
	partialCachedReqSubject[authzSubject.requestCount-2] = deniedInsertRequest

	partialCachedRespSubject.Decisions[authzSubject.requestCount-1].Permitted = true
	partialCachedRespSubject.Decisions[authzSubject.requestCount-2].Permitted = false
	partialCachedRespSubject.Decisions[authzSubject.requestCount-2].Reason = DenyReasonIAM

	expectedPartialCachedRespSubject := getExpectedCachedResponse(partialCachedRespSubject)
	expectedPartialCachedRespSubject.Decisions[authzSubject.requestCount-1].Cached = true // true now since policy gives permit and this request falls into the policy
	expectedPartialCachedRespSubject.Decisions[authzSubject.requestCount-2].Cached = false

	authzAction := authzTestRequest{
		useDifferentAction: true,
		requestCount:       98,
		trace:              "txid-authz-different-actions",
	}
	// Error is ignored since there is no error
	reqAction, respAction, _ := authzAction.generateAuthzRequest()
	cachedRespAction := getExpectedCachedResponse(respAction)

	//permittedTokenSubjectRequest, err := permittedRequest[0].duplicate()
	//assert.Nil(t, err)
	permittedCRNTokenSubjectRequest := Attributes{
		"action": "iam.policy.read",
		"resource": Attributes{
			"serviceName":     "kitchen-tracker",
			"serviceInstance": "mykitchen",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
	}
	crnTokenBodySubject, err := GetSubjectFromToken(crnToken, true)
	assert.Nil(t, err)
	permittedCRNTokenSubjectRequest["subject"] = crnTokenBodySubject
	permittedCRNTokenSubjectRequests := Requests{permittedCRNTokenSubjectRequest}

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
				Trace:       "txid-denied-due-to-wrong-servicename",
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz/bulk",
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
			name: "More than 100 authz request",
			reqs: &large101AuthzRequest,
			args: isAuthorizedTraceArgs{authz101TestLarge.trace},
			want: large101AuthzResp,
			wantErr: &InternalError{
				Trace:   authz101TestLarge.trace,
				Message: "Invalid request: The maximum number of requests allowed is 100",
			},
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
			name:    "cached authz with different subjects and new items",
			reqs:    &partialCachedReqSubject,
			args:    isAuthorizedTraceArgs{partialCachedRespSubject.Trace},
			want:    expectedPartialCachedRespSubject,
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
		{
			name: "permitted crn token subject",
			reqs: &permittedCRNTokenSubjectRequests,
			args: isAuthorizedTraceArgs{"txid-permitted-crn-token-from-pdp"},
			want: AuthzResponse{
				Trace: "txid-permitted-crn-token-from-pdp",
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
			name: "cached permitted token subject",
			reqs: &permittedCRNTokenSubjectRequests,
			args: isAuthorizedTraceArgs{"txid-permitted-crn-token-from-cache"},
			want: AuthzResponse{
				Trace: "txid-permitted-crn-token-from-cache",
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
	}

	pepConfig := &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 32, //48,
		LogOutput:         io.Discard,
	}

	err = Configure(pepConfig)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	for _, tt := range tests {
		authzRunner(tt, t)
	}
}

func TestRequests_isAuthorizedWithSubjectToken(t *testing.T) {

	pepConfig := &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 32,
		LogOutput:         io.Discard}

	err := Configure(pepConfig)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	serviceSubject, err := GetSubjectFromToken(serviceToken, true)
	assert.Nil(t, err)
	userSubject, err := GetSubjectFromToken(userToken, true)
	assert.Nil(t, err)
	crnSubject, err := GetSubjectFromToken(crnToken, true)
	assert.Nil(t, err)
	crnSubjectNoRealm, err := GetSubjectFromToken(crnTokenNoRealm, true)
	assert.Nil(t, err)
	mixedPermitAttributesSubject, err := GetSubjectFromToken(crnToken, true)
	assert.Nil(t, err)
	mixedErrorAttributesSubject, err := GetSubjectFromToken(crnToken, true)
	assert.Nil(t, err)
	mixedPermitAttributesSubject["test-attribute"] = "chatko-book1"
	mixedErrorAttributesSubject["resource"] = "chatko-book1"

	tests := []isAuthorizedTestTemplate{
		{
			name: "token subject permitted single request service token",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": serviceSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-service-token"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-service-token",
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
			name: "token subject permitted single request user token",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": userSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-user-token"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-user-token",
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
			name: "token subject permitted single request crn token",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": crnSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-crn-token"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-crn-token",
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
			name: "token subject permit single request mixedAttributesSubject token",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": mixedPermitAttributesSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permit-from-pdp-crn-token-mixedAttributesSubject"},
			want: AuthzResponse{
				Trace: "txid-permit-from-pdp-crn-token-mixedAttributesSubject",
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
			name: "token subject error single request mixedAttributesSubject token",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": mixedErrorAttributesSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-error-from-pdp-crn-token-mixedAttributesSubject"},
			want: AuthzResponse{},
			wantErr: &APIError{
				EndpointURI: "https://iam.test.cloud.ibm.com/v2/authz/bulk",
				StatusCode:  400,
				Trace:       "txid-error-from-pdp-crn-token-mixedAttributesSubject",
				Message:     "",
			},
		},
		{
			name: "token subject permitted single request crn token without realm encoding",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": crnSubjectNoRealm,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-crn-token-norealm"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-crn-token-norealm",
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
			name: "token subject permitted single request crn token subject",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": crnSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-crn-token-subject"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-crn-token-subject",
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
			name: "token subject permitted single request crn token subject without realm encoding",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": crnSubjectNoRealm,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-crn-token-subject-norealm"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-crn-token-subject-norealm",
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
			name: "token subject permitted multi request subject attributes",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": crnSubject,
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": userSubject,
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": serviceSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-multiple-requests-subject-attributes"},
			want: AuthzResponse{
				Trace: "txid-permitted-multiple-requests-subject-attributes",
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
			name: "token subject permitted multiple request crn subject claims without realm encoding",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": crnSubjectNoRealm,
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": userSubject,
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": serviceSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-multi-crn-subject-claims-norealm"},
			want: AuthzResponse{
				Trace: "txid-permitted-multi-crn-subject-claims-norealm",
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
			name: "token subject permitted multiple request crn token without realm encoding",
			reqs: &Requests{
				{
					"action": "iam.policy.read",
					"resource": Attributes{
						"serviceName":     "kitchen-tracker",
						"serviceInstance": "mykitchen",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": crnSubjectNoRealm,
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": userSubject,
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": serviceSubject,
				},
			},
			args: isAuthorizedTraceArgs{"txid-permitted-multi-crn-token-norealm"},
			want: AuthzResponse{
				Trace: "txid-permitted-multi-crn-token-norealm",
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
	}

	for _, tt := range tests {
		authzRunner(tt, t)
	}

}

func authzRunner(tt isAuthorizedTestTemplate, t *testing.T) {
	t.Run(tt.name, func(t *testing.T) {
		// tests for the cache must be prefixed with "cached"
		if !strings.HasPrefix(tt.name, "cached") {
			cache := getDecisionCache()
			cache.Reset()
		}

		// if !strings.Contains(tt.name, "mixedAttributesSubject") {
		// 	t.Skip("skip")
		// }

		token, err := GetToken()

		if err != nil {
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Requests.isAuthorized() test %+v error = %+v, wantErr %+v", tt.name, err, tt.wantErr)
			}
			return
		}

		got, err := tt.reqs.isAuthorized(tt.args.trace, token)
		if err != nil {
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Requests.isAuthorized() test %+v error = %+v, wantErr %+v", tt.name, err, tt.wantErr)
			}
			return
		}
		if tt.want.Trace == "" || tt.args.trace == "" {
			tt.want.Trace = got.Trace
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Requests.isAuthorized() = %+v, want %+v", got, tt.want)
		}
	})
}

func TestRequests_expiredIsAuthorized(t *testing.T) {
	pepConfig := &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 48,
		CacheDefaultTTL:   time.Duration(1 * time.Nanosecond),
		LogLevel:          LevelError,
	}

	err := Configure(pepConfig)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	// create and store cache key pattern
	c := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{"serviceName"},
			{"serviceName", "accountId"},
			{"serviceName", "accountId", "serviceInstance"},
			{"serviceName", "accountId", "serviceInstance", "resourceType"},
			{"serviceName", "accountId", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	currentCacheKeyInfo.storeCacheKeyPattern(c)

	req := &Requests{
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
	}

	token, err := GetToken()

	if err != nil {
		t.Errorf("Requests.isAuthorized() error = %+v, wantErr %+v", err, "nil")
	}

	got, err := req.isAuthorized("txid-check-cache-expiry", token)

	if err != nil {
		t.Errorf("Requests.isAuthorized() error = %+v, wantErr %+v", err, "nil")
	}

	if !got.Decisions[0].Permitted == true || !got.Decisions[0].Cached == false || !got.Decisions[0].Expired == false {
		t.Errorf("Wrong state of authz response: %+v", got)
	}

	time.Sleep(1 * time.Millisecond)

	token, err = GetToken()

	if err != nil {
		t.Errorf("Requests.isAuthorized() error = %+v, wantErr %+v", err, "nil")
	}

	got, err = req.isAuthorized("txid-check-cache-expiry2", token)

	if err != nil {
		t.Errorf("Requests.isAuthorized() error = %+v, wantErr %+v", err, "nil")
	}

	if !got.Decisions[0].Permitted == true || !got.Decisions[0].Cached == false || !got.Decisions[0].Expired == false || got.Trace != "txid-check-cache-expiry2" {
		t.Errorf("Wrong state of authz response, wanted permitted: True, Cached:false, Expired: false. Instead got: %+v", got)
	}
}

func Test_toAuthzResponse(t *testing.T) {
	type args struct {
		cachedDecisions []cache.CachedDecision
		trace           string
	}
	tests := []struct {
		name string
		args args
		want AuthzResponse
	}{
		{
			name: "empty cached decisions",
			args: args{
				trace:           "txid-empty-cached-decisions",
				cachedDecisions: []cache.CachedDecision{},
			},
			want: AuthzResponse{
				Trace:     "txid-empty-cached-decisions",
				Decisions: []Decision{},
			},
		},
		{
			name: "1 permitted cached decision",
			args: args{
				trace: "txid-1-permitted-cached-decision",
				cachedDecisions: []cache.CachedDecision{
					{
						Permitted: true,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
				},
			},
			want: AuthzResponse{
				Trace: "txid-1-permitted-cached-decision",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
			},
		},
		{
			name: "1 expired permitted cached decision",
			args: args{
				trace: "txid-1-expired-permitted-cached-decision", cachedDecisions: []cache.CachedDecision{
					{
						Permitted: true,
						ExpiresAt: time.Now(),
					},
				},
			},
			want: AuthzResponse{
				Trace: "txid-1-expired-permitted-cached-decision",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    true,
						RetryCount: 0,
					},
				},
			},
		},
		{
			name: "1 denied cached decision",
			args: args{
				trace: "txid-1-denied-cached-decision",
				cachedDecisions: []cache.CachedDecision{
					{
						Permitted: false,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
				},
			},
			want: AuthzResponse{
				Trace: "txid-1-denied-cached-decision",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
			},
		},
		{
			name: "1 expired denied cached decision",
			args: args{
				trace: "txid-1-expired-denied-cached-decision",
				cachedDecisions: []cache.CachedDecision{
					{
						Permitted: false,
						ExpiresAt: time.Now(),
					},
				},
			},
			want: AuthzResponse{
				Trace: "txid-1-expired-denied-cached-decision",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     true,
						Expired:    true,
						RetryCount: 0,
					},
				},
			},
		},
		{
			name: "multi permitted/denied cached decision",
			args: args{
				trace: "txid-multiple-cached-decision",
				cachedDecisions: []cache.CachedDecision{
					{
						Permitted: false,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
					{
						Permitted: true,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
					{
						Permitted: true,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
					{
						Permitted: false,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
					{
						Permitted: true,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
					{
						Permitted: false,
						ExpiresAt: time.Now().Add(1 * time.Minute),
					},
				},
			},
			want: AuthzResponse{
				Trace: "txid-multiple-cached-decision",
				Decisions: []Decision{
					{
						Permitted:  false,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
					{
						Permitted:  false,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
					{
						Permitted:  false,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toAuthzResponse(tt.args.cachedDecisions, tt.args.trace)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toAuthzResponse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRequests_buildAuthzRequests(t *testing.T) {
	tests := []struct {
		name          string
		reqs          *Requests
		wantAuthzReqs AuthzRequests
	}{
		{
			name:          "empty input request",
			reqs:          &Requests{},
			wantAuthzReqs: AuthzRequests{},
		},
		{
			name: "single request",
			reqs: &permittedRequest,
			wantAuthzReqs: AuthzRequests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"attributes": Attributes{
							"serviceName":     "gopep",
							"serviceInstance": "chatko-book1",
							"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						},
					},
					"subject": Attributes{

						"attributes": Attributes{
							"id": "IBMid-3100015XDS",
						},
					},
				},
			},
		},
		{
			name: "multiple requests",
			reqs: &multipleRequests,
			wantAuthzReqs: AuthzRequests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"attributes": Attributes{

							"serviceName":     "gopep",
							"serviceInstance": "chatko-book1",
							"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						},
					},
					"subject": Attributes{
						"attributes": Attributes{

							"id": "IBMid-3100015XDS",
						},
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"attributes": Attributes{

							"serviceName":     "gopep",
							"serviceInstance": "chatko-book2",
							"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						},
					},
					"subject": Attributes{
						"attributes": Attributes{

							"id": "IBMid-3100015XDS",
						},
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"attributes": Attributes{

							"serviceName":     "gopep",
							"serviceInstance": "chatko-book3",
							"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						},
					},
					"subject": Attributes{
						"attributes": Attributes{

							"id": "IBMid-3100015XDS",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuthzReqs := tt.reqs.buildAuthzRequests()
			if !reflect.DeepEqual(gotAuthzReqs, tt.wantAuthzReqs) {
				t.Errorf("Requests.buildAuthzRequests() = %v, want %v", gotAuthzReqs, tt.wantAuthzReqs)
			}
		})
	}
}

func TestRequests_smallest(t *testing.T) {
	tests := []struct {
		name  string
		reqs  *Requests
		want  int
		want1 Attributes
	}{
		{
			name:  "empty request",
			reqs:  &Requests{},
			want:  -1,
			want1: Attributes{},
		},
		{
			name: "a single request",
			reqs: &permittedRequest,
			want: 0,
			want1: Attributes{
				"serviceName":     "gopep",
				"serviceInstance": "chatko-book1",
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
			},
		},
		{
			name: "multiple requests",
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: 1,
			want1: Attributes{
				"serviceName":     "gopep",
				"serviceInstance": "chatko-book2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.reqs.smallest()
			if got != tt.want {
				t.Errorf("Requests.smallest() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Requests.smallest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestAttributes_hasKeyAndValue(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		a    Attributes
		args args
		want bool
	}{
		{
			name: "empty resource",
			a:    Attributes{},
			args: args{
				key:   "accountId",
				value: "account id value",
			},
			want: false,
		},
		{
			name: "non empty resource but with no target key",
			a: Attributes{
				"serviceName":     "gopep",
				"serviceInstance": "chatko-book1",
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
			},
			args: args{
				key:   "region",
				value: "2c17c4e5587783961ce4a0aa415054e7",
			},
			want: false,
		},
		{
			name: "non empty resource but with no target value",
			a: Attributes{
				"serviceName":     "gopep",
				"serviceInstance": "chatko-book1",
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
			},
			args: args{
				key:   "accountId",
				value: "account id value",
			},
			want: false,
		},
		{
			name: "non empty resource",
			a: Attributes{

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
			args: args{
				key:   "accountId",
				value: "2c17c4e5587783961ce4a0aa415054e7",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.hasKeyAndValue(tt.args.key, tt.args.value); got != tt.want {
				t.Errorf("Attributes.hasKeyAndValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttributes_duplicate(t *testing.T) {
	tests := []struct {
		name    string
		a       *Attributes
		want    Attributes
		wantErr bool
	}{
		{
			name:    "empty attributes",
			a:       &Attributes{},
			want:    Attributes{},
			wantErr: false,
		},
		{
			name:    "Nil attributes",
			a:       nil,
			want:    Attributes{},
			wantErr: false,
		},
		{
			name: "Non empty attributes",
			a: &Attributes{
				"resource": Attributes{
					"serviceName":     "gopep",
					"serviceInstance": "chatko-book1",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				},
			},
			want: Attributes{
				"resource": Attributes{
					"serviceName":     "gopep",
					"serviceInstance": "chatko-book1",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.duplicate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Attributes.duplicate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes.duplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequests_extractAttributes(t *testing.T) {
	tests := []struct {
		name    string
		reqs    *Requests
		want    []Attributes
		wantErr bool
	}{
		{
			name:    "empty request",
			reqs:    &Requests{},
			want:    []Attributes{},
			wantErr: false,
		},
		{
			name: "request with no resource",
			reqs: &Requests{
				{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "request with resource",
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
			},
			want: []Attributes{
				{
					"serviceName":     "gopep",
					"serviceInstance": "chatko-book1",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				},
			},
			wantErr: false,
		},
		{
			name: "multi requests with resource",
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book3",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: []Attributes{
				{
					"serviceName":     "gopep",
					"serviceInstance": "chatko-book1",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				},
				{

					"serviceName":     "gopep",
					"serviceInstance": "chatko-book2",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				},
				{

					"serviceName":     "gopep",
					"serviceInstance": "chatko-book3",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				},
			},
			wantErr: false,
		},
		{
			name: "multi requests with missing resource",
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
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book3",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.reqs.extractAttributes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Requests.extractAttributes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Requests.extractAttributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequests_hasUniqueSubject(t *testing.T) {
	tests := []struct {
		name string
		reqs *Requests
		want bool
	}{
		{
			name: "empty request",
			reqs: &Requests{},
			want: false,
		},
		{
			name: "a permitted request",
			reqs: &permittedRequest,
			want: true,
		},
		{
			name: "multiple requests",
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: true,
		},
		{
			name: "multiple requests without first subject",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests with different subjects",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDR",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reqs.hasUniqueSubject(); got != tt.want {
				t.Errorf("Requests.hasUniqueSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequests_hasUniqueAction(t *testing.T) {
	tests := []struct {
		name string
		reqs *Requests
		want bool
	}{
		{
			name: "empty request",
			reqs: &Requests{},
			want: false,
		},
		{
			name: "a permitted request",
			reqs: &permittedRequest,
			want: true,
		},
		{
			name: "multiple requests",
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: true,
		},
		{
			name: "multiple requests without first action",
			reqs: &Requests{
				{
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests with different actions",
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
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without action in the 2nd request",
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reqs.hasUniqueAction(); got != tt.want {
				t.Errorf("Requests.hasUniqueAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequests_hasUniqueAccountID(t *testing.T) {
	tests := []struct {
		name string
		reqs *Requests
		want bool
	}{
		{
			name: "empty request",
			reqs: &Requests{},
			want: false,
		},
		{
			name: "a permitted request",
			reqs: &permittedRequest,
			want: true,
		},
		{
			name: "multiple requests",
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: true,
		},
		{
			name: "multiple requests with different account id",
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
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without first account id",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without first resource",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without second resource",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
				},
				{
					"action": "gopep.books.write",
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without second account id",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
				},
				{
					"action": "gopep.books.write",
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reqs.hasUniqueAccountID(); got != tt.want {
				t.Errorf("Requests.hasUniqueAccountID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequests_hasUniqueServiceName(t *testing.T) {
	tests := []struct {
		name string
		reqs *Requests
		want bool
	}{
		{
			name: "empty request",
			reqs: &Requests{},
			want: false,
		},
		{
			name: "a permitted request",
			reqs: &permittedRequest,
			want: true,
		},
		{
			name: "multiple requests",
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
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: true,
		},
		{
			name: "multiple requests with different service name",
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
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep1",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without first resource",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep1",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without first service name",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep1",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without 2nd service name",
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
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reqs.hasUniqueServiceName(); got != tt.want {
				t.Errorf("Requests.hasUniqueServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequests_getCacheKey(t *testing.T) {
	// create and store cache key pattern
	cSRA := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	currentCacheKeyInfo.storeCacheKeyPattern(cSRA)

	cRAS := CacheKeyPattern{
		Order: []string{"resource", "action", "subject"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	cRSA := CacheKeyPattern{
		Order: []string{"resource", "subject", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	cARS := CacheKeyPattern{
		Order: []string{"action", "resource", "subject"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	// variables for tests
	validSmallRequestcSRA := CacheKey("id:Ron;serviceName:Hogwarts;action:wingardium.leviosa")
	validSmallRequestcRAS := CacheKey("serviceName:Hogwarts;action:wingardium.leviosa;id:Ron")
	validSmallRequestcRSA := CacheKey("serviceName:Hogwarts;id:Ron;action:wingardium.leviosa")
	validSmallRequestcARS := CacheKey("action:wingardium.leviosa;serviceName:Hogwarts;id:Ron")

	validSmallRequestStruct := []struct {
		name    string
		reqs    Request
		want    CacheKey
		pattern CacheKeyPattern
	}{
		{
			name: "valid small request action resource subject",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want:    validSmallRequestcARS,
			pattern: cARS,
		},
		{
			name: "valid small request resource action subject",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want:    validSmallRequestcRAS,
			pattern: cRAS,
		},
		{
			name: "valid small request resource subject action",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want:    validSmallRequestcRSA,
			pattern: cRSA,
		},
	}

	validLargeRequest := CacheKey("id:Ron,scope:ibm magic;serviceName:Hogwarts,accountID:934,serviceInstance:class,resourceType:1st year,resource:spells;action:wingardium.leviosa")

	validIDNoResourceRequest := CacheKey("id:Ron;;action:wingardium.leviosa")

	validIDScopeNoResourceRequest := CacheKey("id:Ron,scope:ibm magic;;action:wingardium.leviosa")

	validOutOfSubjectPatternRequest := Request{
		"subject": Attributes{
			"id":        "Ron",
			"accoundID": "3",
		},
		"resource": Attributes{
			"serviceName": "Hogwarts",
		},
		"action": "wingardium.leviosa",
	}

	expectedValidOutOfSubjectPatternRequest, _ := json.Marshal(validOutOfSubjectPatternRequest)

	noSubjectIDRequest := Request{
		"subject": Attributes{
			"accoundID": "3",
		},
		"resource": Attributes{
			"serviceName": "Hogwarts",
		},
		"action": "wingardium.leviosa",
	}

	expectedNoSubjectIDRequest, _ := json.Marshal(noSubjectIDRequest)

	validCustomResourceRequest := Request{
		"subject": Attributes{
			"id": "Ron",
		},
		"resource": Attributes{
			"serviceName":     "Hogwarts",
			"customAttribute": "1234",
		},
		"action": "wingardium.leviosa",
	}

	expectedValidCustomResourceRequest, _ := json.Marshal(validCustomResourceRequest)

	validOutOfResourcePatternRequest := Request{
		"subject": Attributes{
			"id": "Ron",
		},
		"resource": Attributes{
			"serviceName":     "Hogwarts",
			"serviceInstance": "class",
		},
		"action": "wingardium.leviosa",
	}

	expectedValidOutOfResourcePatternRequest, _ := json.Marshal(validOutOfResourcePatternRequest)

	noServiceNamePatternRequest := Request{
		"subject": Attributes{
			"id": "Ron",
		},
		"resource": Attributes{
			"accountID":       "934",
			"serviceInstance": "class",
			"resourceType":    "1st year",
			"resource":        "spells",
		},
		"action": "wingardium.leviosa",
	}

	expectedNoServiceNamePatternRequest, _ := json.Marshal(noServiceNamePatternRequest)

	missingSeveralAttributesRequest := Request{
		"subject": Attributes{
			"id":    "Ron",
			"scope": "ibm magic",
		},
		"resource": Attributes{
			"accountID":       "934",
			"serviceInstance": "class",
			"resource":        "spells",
		},
		"action": "wingardium.leviosa",
	}

	expectedMissingSeveralAttributesRequest, _ := json.Marshal(missingSeveralAttributesRequest)

	unsupportedResourceAttributeRequest := Request{
		"subject": Attributes{
			"id": "Ron",
		},
		"resource": Attributes{
			"unsupportedName": "unsupportedValue",
		},
		"action": "wingardium.leviosa",
	}

	expectedUnsupportedResourceAttributesRequest, _ := json.Marshal(unsupportedResourceAttributeRequest)

	allKeysTests := []struct {
		name string
		reqs Request
		want CacheKey
	}{
		{
			name: "empty request",
			reqs: Request{},
			want: nil,
		},
		{
			name: "nil request",
			reqs: nil,
			want: nil,
		},
		{
			name: "resource missing",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"action": "wingardium.leviosa",
			},
			want: nil,
		},
		{
			name: "empty resource attributes",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{},
				"action":   "wingardium.leviosa",
			},
			want: validIDNoResourceRequest,
		},
		{
			name: "empty object resource attributes",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": "",
				"action":   "wingardium.leviosa",
			},
			want: nil,
		},
		{
			name: "unsupported object resource attributes",
			reqs: unsupportedResourceAttributeRequest,
			want: expectedUnsupportedResourceAttributesRequest,
		},
		{
			name: "subject missing",
			reqs: Request{
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want: nil,
		},
		{
			name: "empty subject attributes",
			reqs: Request{
				"subject": Attributes{},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want: nil,
		},
		{
			name: "action missing attributes",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
			},
			want: nil,
		},
		{
			name: "action missing attributes",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "",
			},
			want: nil,
		},
		{
			name: "valid out of subject pattern request",
			reqs: validOutOfSubjectPatternRequest,
			want: expectedValidOutOfSubjectPatternRequest,
		},
		{
			name: "no subject ID request",
			reqs: noSubjectIDRequest,
			want: expectedNoSubjectIDRequest,
		},
		{
			name: "valid custom resource request",
			reqs: validCustomResourceRequest,
			want: expectedValidCustomResourceRequest,
		},
		{
			name: "valid out of resource pattern request",
			reqs: validOutOfResourcePatternRequest,
			want: expectedValidOutOfResourcePatternRequest,
		},
		{
			name: "no serviceName pattern request",
			reqs: noServiceNamePatternRequest,
			want: expectedNoServiceNamePatternRequest,
		},
		{
			name: "missing several attributes request",
			reqs: missingSeveralAttributesRequest,
			want: expectedMissingSeveralAttributesRequest,
		},
		{
			name: "valid small request",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want: validSmallRequestcSRA,
		},
		{
			name: "valid large request",
			reqs: Request{
				"subject": Attributes{
					"id":    "Ron",
					"scope": "ibm magic",
				},
				"resource": Attributes{
					"serviceName":     "Hogwarts",
					"accountID":       "934",
					"serviceInstance": "class",
					"resourceType":    "1st year",
					"resource":        "spells",
				},
				"action": "wingardium.leviosa",
			},
			want: validLargeRequest,
		},
	}

	advancedKeysTests := []struct {
		name string
		reqs Request
		want CacheKey
	}{
		{
			name: "advanced only empty request",
			reqs: Request{},
			want: nil,
		},
		{
			name: "advanced only nil request",
			reqs: nil,
			want: nil,
		},
		{
			name: "advanced only resource missing",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"action": "wingardium.leviosa",
			},
			want: nil,
		},
		{
			name: "advanced only empty resource attributes",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{},
				"action":   "wingardium.leviosa",
			},
			want: validIDNoResourceRequest,
		},
		{
			name: "advanced only subject missing",
			reqs: Request{
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want: nil,
		},
		{
			name: "advanced only empty subject attributes",
			reqs: Request{
				"subject": Attributes{},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want: nil,
		},
		{
			name: "advanced only action missing attributes",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
			},
			want: nil,
		},
		{
			name: "advanced only action missing attributes",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "",
			},
			want: nil,
		},
		{
			name: "advanced only valid out of subject pattern request",
			reqs: validOutOfSubjectPatternRequest,
			want: nil,
		},
		{
			name: "advanced only no subject ID request",
			reqs: noSubjectIDRequest,
			want: nil,
		},
		{
			name: "advanced only invalid advanced custom resource request",
			reqs: validCustomResourceRequest,
			want: nil,
		},
		{
			name: "advanced only valid out of resource pattern request",
			reqs: validOutOfResourcePatternRequest,
			want: nil,
		},
		{
			name: "advanced only invalid no serviceName pattern request",
			reqs: noServiceNamePatternRequest,
			want: nil,
		},
		{
			name: "advanced only invalid missing several attributes request",
			reqs: missingSeveralAttributesRequest,
			want: nil,
		},
		{
			name: "advanced only valid small request",
			reqs: Request{
				"subject": Attributes{
					"id": "Ron",
				},
				"resource": Attributes{
					"serviceName": "Hogwarts",
				},
				"action": "wingardium.leviosa",
			},
			want: validSmallRequestcSRA,
		},
		{
			name: "advanced only valid large request",
			reqs: Request{
				"subject": Attributes{
					"id":    "Ron",
					"scope": "ibm magic",
				},
				"resource": Attributes{
					"serviceName":     "Hogwarts",
					"accountID":       "934",
					"serviceInstance": "class",
					"resourceType":    "1st year",
					"resource":        "spells",
				},
				"action": "wingardium.leviosa",
			},
			want: validLargeRequest,
		},
		{
			name: "advanced ID no resource valid request",
			reqs: Request{
				"subject": Attributes{
					"id":    "Ron",
					"scope": "ibm magic",
				},
				"resource": Attributes{},
				"action":   "wingardium.leviosa",
			},
			want: validIDScopeNoResourceRequest,
		},
		{
			name: "advanced ID and scope invalid resource object request",
			reqs: Request{
				"subject": Attributes{
					"id":    "Ron",
					"scope": "ibm magic",
				},
				"resource": "",
				"action":   "wingardium.leviosa",
			},
			want: nil,
		},
	}

	for _, tt := range allKeysTests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.reqs.getCacheKey(false)

			if !bytes.Equal(got, tt.want) {
				t.Errorf("Requests.getCacheKey() = %v, want %v", string(got), string(tt.want))
			}
		})
	}

	for _, tt := range advancedKeysTests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.reqs.getCacheKey(true)

			if !bytes.Equal(got, tt.want) {
				t.Errorf("Requests.getCacheKey() = %v, want %v", string(got), string(tt.want))
			}
		})
	}

	for _, tt := range validSmallRequestStruct {
		t.Run(tt.name, func(t *testing.T) {
			currentCacheKeyInfo.storeCacheKeyPattern(tt.pattern)
			got := tt.reqs.getCacheKey(false)

			if !bytes.Equal(got, tt.want) {
				t.Errorf("Requests.getCacheKey() = %v, want %v", string(got), string(tt.want))
			}
		})
	}

}

func Test_getCacheFutureProofing(t *testing.T) {
	// create and store cache key pattern
	cEqualLength := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "resourceType", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}, {"id", "expiry"}},
	}

	cEqualLengthDifferentOrder := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "resourceType", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}, {"id", "expiry"}},
	}

	cDuplicateResourcePattern := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	cDuplicateSubjectPattern := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}, {"id", "scope"}},
	}

	cDuplicatePatternMix := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}, {"id", "scope"}},
	}

	cIDandScopeOnly := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id", "scope"}},
	}

	cNewAttributes := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"global"},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource", "unit"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}, {"id", "scope", "expiry"}},
	}

	cMixOriginalOrder := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource", "unit"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}, {"id", "scope", "expiry"}},
	}

	cMixShuffledOrder := CacheKeyPattern{
		Order: []string{"action", "subject", "resource"},
		Resource: [][]string{
			{},
			{"serviceName"},
			{"serviceName", "accountID"},
			{"serviceName", "accountID", "serviceInstance"},
			{"serviceName", "accountID", "serviceInstance", "resourceType"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
			{"serviceName", "accountID", "serviceInstance", "resourceType", "resource", "unit"},
		},
		Subject: [][]string{{"id", "scope"}, {"id", "scope", "expiry"}, {"id"}},
	}

	requests := []struct {
		name    string
		reqs    Request
		pattern CacheKeyPattern
		want    CacheKey
	}{
		{
			name: "Equal length pattern resourceType",
			reqs: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			pattern: cEqualLength,
			want:    CacheKey("id:IBMid-3100015XDS;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "Equal length pattern resourceType",
			reqs: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			pattern: cEqualLengthDifferentOrder,
			want:    CacheKey("id:IBMid-3100015XDS;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "Equal length pattern resource",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resource":        "spells",
				},
				"action": "gopep.books.write",
			},
			pattern: cEqualLength,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resource:spells;action:gopep.books.write"),
		},
		{
			name: "Equal length pattern subject expiry",
			reqs: Request{
				"subject": Attributes{
					"id":     "IBMid-3100015XDS",
					"expiry": "20770101",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cEqualLength,
			want:    CacheKey("id:IBMid-3100015XDS,expiry:20770101;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "Equal length pattern scope",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
				},
				"action": "gopep.books.write",
			},
			pattern: cEqualLength,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2;action:gopep.books.write"),
		},
		{
			name: "Equal length pattern scope",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cDuplicateResourcePattern,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "Equal length pattern scope",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cDuplicateSubjectPattern,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "duplicate pattern mix no pattern found for expiry",
			reqs: Request{
				"subject": Attributes{
					"id":     "IBMid-3100015XDS",
					"expiry": "20770101",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cDuplicatePatternMix,
			want:    nil,
		},
		{
			name: "duplicate pattern mix",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cDuplicatePatternMix,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "ID and Scope only no match",
			reqs: Request{
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cIDandScopeOnly,
			want:    nil,
		},
		{
			name: "ID and Scope only pattern",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cIDandScopeOnly,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "New attributes in pattern full resource",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
					"resource":        "spells",
					"unit":            "1",
				},
				"action": "gopep.books.write",
			},
			pattern: cNewAttributes,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket,resource:spells,unit:1;action:gopep.books.write"),
		},
		{
			name: "New attributes in pattern single",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"global": "globoservice",
				},
				"action": "gopep.books.write",
			},
			pattern: cNewAttributes,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;global:globoservice;action:gopep.books.write"),
		},
		{
			name: "Mix original order 3 attribute resource",
			reqs: Request{
				"subject": Attributes{
					"id":     "IBMid-3100015XDS",
					"scope":  "ibm openid",
					"expiry": "20770101",
				},
				"resource": Attributes{
					"serviceName":  "gopep1",
					"accountID":    "2c17c4e5587783961ce4a0aa415054e8",
					"resourceType": "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cMixOriginalOrder,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid,expiry:20770101;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,resourceType:bucket;action:gopep.books.write"),
		},
		{
			name: "Mix original order 6 attribute resource",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
					"resource":        "spells",
					"unit":            "1",
				},
				"action": "gopep.books.write",
			},
			pattern: cMixOriginalOrder,
			want:    CacheKey("id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket,resource:spells,unit:1;action:gopep.books.write"),
		},
		{
			name: "Mix shuffled order 3 attribute resource no key",
			reqs: Request{
				"subject": Attributes{
					"id":     "IBMid-3100015XDS",
					"scope":  "ibm openid",
					"expiry": "20770101",
				},
				"resource": Attributes{
					"serviceName":  "gopep1",
					"accountID":    "2c17c4e5587783961ce4a0aa415054e8",
					"resourceType": "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cMixShuffledOrder,
			want:    nil,
		},
		{
			name: "Mix shuffled order 3 attribute resource",
			reqs: Request{
				"subject": Attributes{
					"id":     "IBMid-3100015XDS",
					"scope":  "ibm openid",
					"expiry": "20770101",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
				},
				"action": "gopep.books.write",
			},
			pattern: cMixShuffledOrder,
			want:    CacheKey("action:gopep.books.write;id:IBMid-3100015XDS,scope:ibm openid,expiry:20770101;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket"),
		},
		{
			name: "Mix original order 6 attribute resource",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"resourceType":    "bucket",
					"resource":        "spells",
					"unit":            "1",
				},
				"action": "gopep.books.write",
			},
			pattern: cMixShuffledOrder,
			want:    CacheKey("action:gopep.books.write;id:IBMid-3100015XDS,scope:ibm openid;serviceName:gopep1,accountID:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2,resourceType:bucket,resource:spells,unit:1"),
		},
		{
			name: "Mix original order 6 attribute resource",
			reqs: Request{
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm openid",
				},
				"resource": Attributes{
					"serviceName":            "gopep1",
					"accountID":              "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance":        "chatko-book2",
					"resourceType":           "bucket",
					"resource":               "spells",
					"unit":                   "1",
					"anotherAttNotInPattern": "2",
				},
				"action": "gopep.books.write",
			},
			pattern: cMixShuffledOrder,
			want:    nil,
		},
		{
			name: "No key match for equal length pattern",
			reqs: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"imadethisup":     "bucket",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			pattern: cEqualLength,
			want:    nil,
		},
		{
			name: "No key match for no action request",
			reqs: Request{
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"imadethisup":     "bucket",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			pattern: cEqualLength,
			want:    nil,
		},
		{
			name: "No key match for no action pattern",
			reqs: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"accountID":       "2c17c4e5587783961ce4a0aa415054e8",
					"serviceInstance": "chatko-book2",
					"imadethisup":     "bucket",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			pattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
				Resource: [][]string{
					{},
					{"serviceName"},
					{"serviceName", "accountID"},
					{"serviceName", "accountID", "serviceInstance"},
					{"serviceName", "accountID", "serviceInstance", "resourceType"},
					{"serviceName", "accountID", "serviceInstance", "resourceType", "resource"},
				},
				Subject: [][]string{{"id"}, {"id", "scope"}},
			},
			want: nil,
		},
	}

	for i, tt := range requests {
		if i == len(requests)-1 {
			t.Run(tt.name, func(t *testing.T) {
				currentCacheKeyInfo.storeCacheKeyPattern(tt.pattern)
				got := tt.reqs.getCacheKey(true)

				if !bytes.Equal(got, tt.want) {
					t.Errorf("Requests.getCacheKey() = %v, want %v", string(got), string(tt.want))
				}
			})
		}
	}

}

func TestRequests_retrieveCachedDecision(t *testing.T) {

	pepConfig := &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 32, //48,
		LogLevel:          LevelError,
	}

	err := Configure(pepConfig)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	// create and store cache key pattern
	c := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{"serviceName"},
			{"serviceName", "accountId"},
			{"serviceName", "accountId", "serviceInstance"},
			{"serviceName", "accountId", "serviceInstance", "resourceType"},
			{"serviceName", "accountId", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	currentCacheKeyInfo.storeCacheKeyPattern(c)

	toCache := map[string]bool{
		"id:IBMid-3100015XDS;serviceName:gopep1,accountId:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book2;action:gopep.books.write":                                           true,
		"id:IBMid-3100015XDS,scope:cos;serviceName:gopep2;action:gopep.books.write":                                                                                                         true,
		"id:IBMid-3100015XDS;serviceName:gopep1,accountId:2c17c4e5587783961ce4a0aa415054e8,serviceInstance:chatko-book3;action:gopep.books.write":                                           true,
		"id:IBMid-3100015XDS;serviceName:gopep3;action:gopep.books.write":                                                                                                                   true,
		"id:iam-ServiceId-c33ad72f-1546-4fd4-8e94-3418ed0fb6e4,scope:ibm;serviceName:gopep,accountId:2c17c4e5587783961ce4a0aa415054e7,serviceInstance:chatko-book1;action:gopep.books.read": true, //service token
		"id:IBMid-270003GUSX,scope:ibm openid;serviceName:gopep;action:gopep.books.read":                                                                                                    true, //usertoken
		//"id:IBMid-270003GUSX,scope:ibm openid;serviceName:gopep2;action:gopep.books.write":                                                                                                    true, //crn token
	}

	for keyString, decision := range toCache {
		key := CacheKey(keyString)
		key.cacheDecision(decision, 0, DenyReasonNone)
	}

	npk := Request{
		"action": "gopep.books.write",
		"resource": Attributes{
			"serviceName":  "gopep1",
			"accountId":    "1234",
			"resourceType": "bucket",
			"resource":     "test-bucket-public-access",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	}

	nonPatternKeyBytes, _ := json.Marshal(npk) // serialize into bytes a request that won't fit a key pattern

	nonPatternKey := CacheKey(nonPatternKeyBytes) // generate a CacheKey object

	nonPatternKey.cacheDecision(true, 0, DenyReasonNone) // cache the generated key

	serviceSubject, err := GetSubjectFromToken(serviceToken, true)
	assert.Nil(t, err)
	userSubject, err := GetSubjectFromToken(userToken, true)
	assert.Nil(t, err)
	// crnSubject, err := GetSubjectFromToken(crnToken, true)
	// assert.Nil(t, err)

	serviceTokenSubjectRequest := Request{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book1",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": serviceSubject,
	}
	userTokenSubjectRequest := Request{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book1",
			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": userSubject,
	}
	// crnSubjectRequest := Request{
	// 	{
	// 		"action": "iam.policy.read",
	// 		"resource": Attributes{
	// 			"serviceName":     "kitchen-tracker",
	// 			"serviceInstance": "mykitchen",
	// 			"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	// 		},
	// 		"subject": crnSubject,
	// 	},
	// }

	tests := []struct {
		name       string
		attributes Request
		want       *cache.CachedDecision
	}{
		{
			name:       "empty attributes",
			attributes: Request{},
			want:       nil,
		},
		{
			name: "missing subject attributes",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book2",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
				},
			},
			want: nil,
		},
		{
			name: "missing resource attributes",
			attributes: Request{
				"action": "gopep.books.write",
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			want: nil,
		},
		{
			name: "simple request with subject id only attributes",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book2",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name: "simple request with different serviceInstance and only subject id",
			attributes: Request{ // small req 2 only subject id
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book3",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name: "simple request with subject id only and resource type",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book2",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					"resourceType":    "bucket",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name: "simple request with subject id and scope attributes",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book2",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
				},
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "cos",
				},
			},
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name: "high granularity request with subject id only attributes",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book2",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					"resourceType":    "bucket",
					"resource":        "test-bucket-public-access",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name: "high granularity request with subject id and scope attributes",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book2",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					"resourceType":    "bucket",
					"resource":        "test-bucket-public-access",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				}},
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name: "high granularity request with chatko-book3",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "chatko-book3",
					"accountId":       "2c17c4e5587783961ce4a0aa415054e8",
					"resourceType":    "bucket",
					"resource":        "test-bucket-public-access",
				},
				"subject": Attributes{
					"id":    "IBMid-3100015XDS",
					"scope": "ibm",
				},
			},
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name: "non-pattern values",
			attributes: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":  "gopep1",
					"accountId":    "1234",
					"resourceType": "bucket",
					"resource":     "test-bucket-public-access",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			want: &cache.CachedDecision{
				Permitted: true,
			},
			// {
			// 	name: "separator in resource name",
			// },
			// {
			// 	name: "separator in subject name",
			// },
		},
		{
			name:       "cached permitted serviceTokenSubjectRequest",
			attributes: serviceTokenSubjectRequest,
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
		{
			name:       "cached permitted userTokenSubjectRequest",
			attributes: userTokenSubjectRequest,
			want: &cache.CachedDecision{
				Permitted: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.attributes.retrieveCachedDecision(); !compareCachedDecision(got, tt.want) {
				t.Errorf("Requests.retrieveCachedDecision() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isAuthorizedNoPDP(t *testing.T) {
	called := 0
	ts := httptest.NewServer(http.TimeoutHandler(http.HandlerFunc(serverErrorMaker(&called)), 20*time.Millisecond, "server timeout"))
	defer ts.Close()

	largeRequest, largeResponse := generateBulkRequests(10, true, true)
	largeResponse.Trace = "txid-permitted-from-pdp-non-functioning-endpoint-cached"
	largeResponse.ErrorForExpiredResults = "API error code: 0 calling http://127.0.0.1:1337/list for txid: txid-permitted-from-pdp-non-functioning-endpoint-cached Details: dial tcp 127.0.0.1:1337: connect: connection refused"

	largeResponse500 := largeResponse
	largeResponse500.Trace = "txid-permitted-from-pdp-500-error-cached"
	largeResponse500.ErrorForExpiredResults = "API error code: 500 calling " + ts.URL + "/v2/authz/bulk for txid: txid-permitted-from-pdp-500-error-cached Details: 500 - Something bad happened!"

	partialCachedReq := make(Requests, len(largeRequest))
	copy(partialCachedReq, largeRequest)

	insertedRequest := Attributes{
		"action": "gopep.books.read",
		"resource": Attributes{
			"serviceName":     "gopep",
			"serviceInstance": "chatko-book999",
			"accountId":       "3c17c4e5587783961ce4a0aa415054e7",
		},
		"subject": Attributes{
			"id": "IBMid-3100015XDS",
		},
	}

	// Inserting a new permitted request into the request slice
	partialCachedReq[len(largeResponse500.Decisions)-1] = insertedRequest

	action := "gopep.books.read"

	subject, err := GetSubjectFromToken(userToken2, true)
	assert.Nil(t, err)

	chrisTokenRequest := &Requests{
		{
			"action": action,
			"resource": Attributes{
				"serviceName":     "gopep",
				"serviceInstance": "chatko-book5",
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
			},
			"subject": subject,
		},
	}

	subject, err = GetSubjectFromToken(userToken2, true)
	assert.Nil(t, err)

	chrisTokenRequestResource := &Requests{
		{
			"action": action,
			"resource": Attributes{
				"serviceName":     "gopep",
				"serviceInstance": "chatko-book5",
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				"resource":        "blah",
			},
			"subject": subject,
		},
	}

	chrisTokenRequestFullPattern := &Requests{
		{
			"action": action,
			"resource": Attributes{
				"serviceName":     "gopep",
				"serviceInstance": "chatko-book5",
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				"resourceType":    "book",
				"resource":        "blah",
			},
			"subject": subject,
		},
	}

	tests := []isAuthorizedTestTemplate{
		{
			name: "permitted non-functioning endpoint",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-non-functioning-endpoint"},
			want: AuthzResponse{
				Trace:     "txid-permitted-from-pdp-non-functioning-endpoint",
				Decisions: []Decision{},
			},
			wantErr: &APIError{
				EndpointURI: "http://127.0.0.1:1337/list",
				StatusCode:  0,
				Trace:       "txid-permitted-from-pdp-non-functioning-endpoint",
				Message:     "dial tcp 127.0.0.1:1337: connect: connection refused",
			},
		},
		{
			name: "permitted 500 no cache",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-500-no-cache"},
			want: AuthzResponse{
				Trace:     "txid-permitted-from-pdp-500-no-cache",
				Decisions: []Decision{},
			},
			wantErr: &APIError{
				StatusCode:  500,
				EndpointURI: ts.URL + "/v2/authz/bulk",
				Trace:       "txid-permitted-from-pdp-500-no-cache",
				Message:     "500 - Something bad happened!",
			},
		},
		{
			name: "permitted 501 no cache",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-501-no-cache"},
			want: AuthzResponse{
				Trace:     "txid-permitted-from-pdp-501-no-cache",
				Decisions: []Decision{},
			},
			wantErr: &APIError{
				EndpointURI: ts.URL + "/v2/authz/bulk",
				StatusCode:  501,
				Trace:       "txid-permitted-from-pdp-501-no-cache",
				Message:     "501 - Something bad happened!",
			},
		},
		{
			name: "permitted timeout",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-timeout"},
			want: AuthzResponse{
				Trace:     "txid-permitted-from-pdp-timeout",
				Decisions: []Decision{},
			},
			wantErr: &APIError{
				EndpointURI: ts.URL + "/v2/authz/bulk",
				StatusCode:  0,
				Trace:       "txid-permitted-from-pdp-timeout",
				Message:     "context deadline exceeded (Client.Timeout exceeded while awaiting headers)",
			},
		},
		{
			name: "call real pdp wrong endpoint v3/authz without cache",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-v3-no-cache"},
			want: AuthzResponse{
				Trace:                  "txid-permitted-from-pdp-v3-no-cache",
				Decisions:              []Decision{},
				ErrorForExpiredResults: "",
			},
			wantErr: &APIError{
				EndpointURI: "https://iam.test.cloud.ibm.com/v3/authz/bulk",
				StatusCode:  404,
				Trace:       "txid-permitted-from-pdp-v3-no-cache",
				Message:     "EOF",
			},
		},
		{
			name: "call real pdp",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-real-pdp"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-real-pdp",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "",
			},
			wantErr: nil,
		},
		{
			name: "cached call real pdp wrong endpoint v3/authz",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-v3-cache"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-v3-cache",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "",
			},
			wantErr: nil,
		},
		{
			name: "cached permitted 500",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-500-cache"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-500-cache",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "",
			},
			wantErr: nil,
		},
		{
			name: "cached permitted 500 expired",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-500-cache-expired"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-500-cache-expired",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    true,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "API error code: 500 calling " + ts.URL + "/v2/authz/bulk for txid: txid-permitted-from-pdp-500-cache-expired Details: 500 - Something bad happened!",
			},
			wantErr: nil,
		},
		{
			name: "cached permitted non-functioning endpoint",
			reqs: &permittedRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-non-functioning-endpoint-cached"},
			want: AuthzResponse{
				Trace: "txid-permitted-from-pdp-non-functioning-endpoint-cached",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    true,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "API error code: 0 calling http://127.0.0.1:1337/list for txid: txid-permitted-from-pdp-non-functioning-endpoint-cached Details: dial tcp 127.0.0.1:1337: connect: connection refused",
			},
			wantErr: nil,
		},
		{
			name:    "cached multiple requests non-functional endpoint",
			reqs:    &largeRequest,
			args:    isAuthorizedTraceArgs{"txid-permitted-from-pdp-non-functioning-endpoint-cached"},
			want:    largeResponse,
			wantErr: nil,
		},
		{
			name:    "cached multiple requests 500 error",
			reqs:    &largeRequest,
			args:    isAuthorizedTraceArgs{"txid-permitted-from-pdp-500-error-cached"},
			want:    largeResponse500,
			wantErr: nil,
		},
		{
			name: "cached multiple partially cached",
			reqs: &partialCachedReq,
			args: isAuthorizedTraceArgs{"txid-permitted-from-pdp-500-error-partial-cached"},
			want: AuthzResponse{},
			wantErr: &APIError{
				StatusCode:  500,
				EndpointURI: ts.URL + "/v2/authz/bulk",
				Trace:       "txid-permitted-from-pdp-500-error-partial-cached",
				Message:     "500 - Something bad happened!",
			},
		},
		{
			name: "permitted userTokenSubjectRequest",
			reqs: chrisTokenRequest,
			args: isAuthorizedTraceArgs{"txid-permitted-chris-token"},
			want: AuthzResponse{
				Trace: "txid-permitted-chris-token",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     false,
						Expired:    false,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "",
			},
			wantErr: nil,
		},
		{
			name: "cached permitted userTokenSubjectRequest with resource no restype",
			reqs: chrisTokenRequestResource,
			args: isAuthorizedTraceArgs{"txid-permitted-chris-token-resource"},
			want: AuthzResponse{
				Trace: "txid-permitted-chris-token-resource",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "",
			},
			wantErr: nil,
		},
		{
			name: "cached permitted userTokenSubjectRequest",
			reqs: chrisTokenRequestFullPattern,
			args: isAuthorizedTraceArgs{"txid-permitted-chris-token-full-pattern"},
			want: AuthzResponse{
				Trace: "txid-permitted-chris-token-full-pattern",
				Decisions: []Decision{
					{
						Permitted:  true,
						Cached:     true,
						Expired:    false,
						RetryCount: 0,
					},
				},
				ErrorForExpiredResults: "",
			},
			wantErr: nil,
		},
	}

	// non-functioning endpoint
	pc := &Config{
		Environment:       Custom,
		APIKey:            os.Getenv("API_KEY"),
		AuthzEndpoint:     "http://127.0.0.1:1337/authz",
		ListEndpoint:      "http://127.0.0.1:1337/list",
		TokenEndpoint:     stagingTokenEndpoint,
		KeyEndpoint:       stagingKeyEndpoint,
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
	}

	err = Configure(pc)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	authzRunner(tests[0], t)

	// 500 range errors no cache, expired entries not enabled
	pc = &Config{
		Environment:       Custom,
		APIKey:            os.Getenv("API_KEY"),
		AuthzEndpoint:     ts.URL + "/v2/authz",
		ListEndpoint:      ts.URL + "/v2/authz/bulk",
		TokenEndpoint:     stagingTokenEndpoint,
		KeyEndpoint:       stagingKeyEndpoint,
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
	}

	err = Configure(pc)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	called = 1
	authzRunner(tests[1], t)
	called = 2
	authzRunner(tests[2], t)

	pc = &Config{
		Environment:       Custom,
		APIKey:            os.Getenv("API_KEY"),
		AuthzEndpoint:     ts.URL + "/v2/authz",
		ListEndpoint:      ts.URL + "/v2/authz/bulk",
		TokenEndpoint:     stagingTokenEndpoint,
		KeyEndpoint:       stagingKeyEndpoint,
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
		PDPTimeout:        20 * time.Millisecond,
	}

	err = Configure(pc)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}
	called = 3
	authzRunner(tests[3], t)

	// valid endpoint with bad path no cache, expired entries not enabled
	pc = &Config{
		Environment:       Custom,
		APIKey:            os.Getenv("API_KEY"),
		AuthzEndpoint:     "https://iam.test.cloud.ibm.com/v3/authz",
		ListEndpoint:      "https://iam.test.cloud.ibm.com/v3/authz/bulk",
		TokenEndpoint:     stagingTokenEndpoint,
		KeyEndpoint:       stagingKeyEndpoint,
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
	}

	err = Configure(pc)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	authzRunner(tests[4], t)

	// valid pdp endpoint, valid request, no cache. fetches advanced obligation and pattern
	pc = &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
	}

	err = Configure(pc)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	// tests permitted userTokenSubjectRequest and
	authzRunner(tests[13], t)

	// tests cached permitted userTokenSubjectRequest with resource no restype
	authzRunner(tests[14], t)

	// tests cached permitted userTokenSubjectRequest
	authzRunner(tests[15], t)

	// tests call real pdp
	authzRunner(tests[5], t)

	// valid endpoint with bad path with cache
	pepConfig.ListEndpoint = "https://iam.test.cloud.ibm.com/v3/authz/bulk"
	authzRunner(tests[6], t)

	// cached permitted 500
	called = 1
	pepConfig.ListEndpoint = ts.URL + "/v2/authz/bulk"
	authzRunner(tests[7], t)

	// reset cache with short expiry
	pc = &Config{
		Environment:        Staging,
		APIKey:             os.Getenv("API_KEY"),
		DecisionCacheSize:  32,
		LogLevel:           LevelError,
		CacheDefaultTTL:    time.Duration(1 * time.Second),
		EnableExpiredCache: true,
	}

	err = Configure(pc)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	authzRunner(tests[5], t)
	authzRunner(tests[6], t)

	// cached permitted 500 expired
	time.Sleep(1 * time.Second)
	pepConfig.ListEndpoint = ts.URL + "/v2/authz/bulk"

	authzRunner(tests[8], t)

	// cached unreachable
	pepConfig.ListEndpoint = "http://127.0.0.1:1337/list"
	authzRunner(tests[9], t)

	// cached multiple unreachable
	authzRunner(tests[10], t)

	// cached multiple requests 500
	pepConfig.ListEndpoint = ts.URL + "/v2/authz/bulk"
	authzRunner(tests[11], t)

	// cached partial multiple requests 500
	pepConfig.AuthzEndpoint = ts.URL + "/v2/authz/bulk"
	authzRunner(tests[12], t)

}

func TestPerformAuthorizationSingleSubject(t *testing.T) {
	pepConfig := &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
	}

	err := Configure(pepConfig)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}

	//"crn:v1:staging:public:secrets-manager:us-south:a/791f5fb10986423e97aa8512f18b7e65:64be543a-3901-4f54-9d60-854382b21f29::"
	crnToParse := "crn:v1:staging:public:gopep:us-south:a/2c17c4e5587783961ce4a0aa415054e7:chatko-book3::"

	parsedCRN := ParseCRNToAttributes(crnToParse)

	res := Attributes{
		"serviceName":     parsedCRN["serviceName"],     // "gopep",
		"serviceInstance": parsedCRN["serviceInstance"], // "chatko-book3",
		"accountId":       parsedCRN["accountId"],       // "2c17c4e5587783961ce4a0aa415054e7",
		"region":          parsedCRN["region"],
		//"resourceType":    parsedCRN["resourceType"],
		//"resource":        parsedCRN["resource"],
	} //it's the different resource type thats messing it up, /v2/bulk don't like that

	assert.NotNil(t, res, "Failed parsing CRN: %s\n", crnToParse)

	// Create the target map
	groupRes := Attributes{}

	// Copy from the original map to the target map
	for key, value := range res {
		groupRes[key] = value
	}

	groupRes["resourceType"] = "bucket" // "bucket",

	groupRes["resource"] = "bucket-of-books" // "bucket-of-books",

	req := Requests{
		Attributes{
			"subject":  Attributes{"id": "IBMid-3100015XDS"}, // "iam-" + authContext.Manager.ServiceId,
			"action":   "gopep.books.read",                   // managerAction,
			"resource": res,
		},
		Attributes{
			"subject":  Attributes{"id": "IBMid-3100015XDS"}, // "iam-" + authContext.Manager.ServiceId,
			"action":   "gopep.books.read",                   //managerAction,
			"resource": groupRes,
		},
	}
	response, err := PerformAuthorization(&req, "TestPerformAuthorizationSingleSubject")
	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *InternalError:
			// Handling the internal error
			t.Errorf("PerformAuthorization should succeed but failed with internal error: %v\n", err.Error())
		case *APIError:
			// Handling the API error
			t.Errorf("PerformAuthorization should succeed but failed with api error: %v\n", err.Error())
		default:
			t.Errorf("PerformAuthorization should succeed but failed with error: %v\n", err.Error())
		}
	} else if response.Decisions[0].Permitted == true && response.Decisions[1].Permitted == true {
		t.Logf("Success! got the expected decisions %#+v\n", response.Decisions)
	} else {
		t.Errorf("Failue! got unexpected expected decisions %#+v\n", response.Decisions)
	}
}

func TestRequests_hasUniqueResourceType(t *testing.T) {
	tests := []struct {
		name string
		reqs *Requests
		want bool
	}{
		{
			name: "empty request",
			reqs: &Requests{},
			want: false,
		},
		{
			name: "a permitted request",
			reqs: &permittedRequest,
			want: true,
		},
		{
			name: "multiple requests same resource type",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: true,
		},
		{
			name: "multiple requests missing one resource type",
			reqs: &Requests{
				{
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests with different resource types",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.write",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "pail",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "container",
						"resource":        "container-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests without two resource types",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
		{
			name: "multiple requests with mixed resource types",
			reqs: &Requests{
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book1",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "bucket",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
				{
					"action": "gopep.books.read",
					"resource": Attributes{
						"serviceName":     "gopep",
						"serviceInstance": "chatko-book2",
						"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
						"resourceType":    "pail",
						"resource":        "bucket-of-books",
					},
					"subject": Attributes{
						"id": "IBMid-3100015XDS",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reqs.hasUniqueResourceType(); got != tt.want {
				t.Errorf("Requests.hasUniqueAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isAuthorizedRetry(t *testing.T) {
	attempt := 0
	var attemptMtx sync.Mutex
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptMtx.Lock()
			if (r.URL.String() == "/v2/authz" || r.URL.String() == "/v2/authz/bulk") && attempt == 5 {
				// no error here
				attempt++
				attemptMtx.Unlock()
				decision := Decision{
					Permitted:  true,
					Cached:     false,
					Expired:    false,
					RetryCount: 0,
				}

				var response AuthzResponse

				response.Decisions = append(response.Decisions, decision)

				w.Header().Set("Vary", "Accept-Encoding")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				jsonResponse := []byte(`{
					"decisions": [
						{
							"decision": "Permit",
							"obligation": {
								"actions": [
									"create",
									"unwrap",
									"rewrap",
									"write"
								],
								"maxCacheAgeSeconds": 600,
								"subject": {
									"attributes": {
										"serviceName": "cloud-object-storage",
										"serviceInstance": "1w1",
										"accountId": "abc"
									}
								}
							}
						}
					]
				}`)

				if _, err := w.Write(jsonResponse); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_, err := w.Write([]byte("400 - Error in test server generating response to authz req."))
					if err != nil {
						fmt.Println(err)
					}
				}

			} else if (r.URL.String() == "/v2/authz" || r.URL.String() == "/v2/authz/bulk") && attempt >= 0 && attempt < 2 {
				// 500 error
				attempt++
				attemptMtx.Unlock()
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("500 - Error in test server generating response to getting token."))
				if err != nil {
					fmt.Println(err)
				}
			} else if (r.URL.String() == "/v2/authz" || r.URL.String() == "/v2/authz/bulk") && attempt == 2 {
				// 502 temporary error
				attempt++
				attemptMtx.Unlock()
				w.WriteHeader(http.StatusBadGateway)
				_, err := w.Write([]byte("502 - Temporary error."))
				if err != nil {
					fmt.Println(err)
				}
			} else if (r.URL.String() == "/v2/authz" || r.URL.String() == "/v2/authz/bulk") && attempt == 3 {
				// timeout error
				attempt++
				attemptMtx.Unlock()
				time.Sleep(15 * time.Second)
				return
			} else if (r.URL.String() == "/v2/authz" || r.URL.String() == "/v2/authz/bulk") && attempt >= 3 && attempt < 5 {
				// 429 error
				attempt++
				attemptMtx.Unlock()
				w.WriteHeader(http.StatusTooManyRequests)
				_, err := w.Write([]byte("429 - Too many requests."))
				if err != nil {
					fmt.Println(err)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte("404 - Invalid path in test."))
				if err != nil {
					fmt.Println(err)
				}
			}
		}))
	defer ts.Close()

	pc := &Config{
		Environment:       Custom,
		APIKey:            os.Getenv("API_KEY"),
		AuthzEndpoint:     ts.URL + "/v2/authz",
		ListEndpoint:      ts.URL + "/v2/authz/bulk",
		TokenEndpoint:     stagingTokenEndpoint,
		KeyEndpoint:       stagingKeyEndpoint,
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
		AuthzRetry:        true,
	}

	err := Configure(pc)
	if err != nil {
		t.Errorf("pep.Configure failed: %v", err)
	}
	token, err := GetToken()
	assert.Nil(t, err)
	authresp, err := permittedRequest.isAuthorized("retry_test", token)
	assert.NotNil(t, err)
	assert.Empty(t, authresp.Decisions)
	assert.Equal(t, "retry_test", authresp.Trace)

	token, err = GetToken()
	assert.Nil(t, err)
	authresp, err = permittedRequest.isAuthorized("retry_test", token)
	assert.Nil(t, err)
	assert.Equal(t, "retry_test", authresp.Trace)
	assert.True(t, authresp.Decisions[0].Permitted)

}

func TestBuildCacheKeysWithUserRequest(t *testing.T) {
	// create and store cache key pattern
	c := CacheKeyPattern{
		Order: []string{"subject", "resource", "action"},
		Resource: [][]string{
			{"serviceName"},
			{"serviceName", "accountId"},
			{"serviceName", "accountId", "serviceInstance"},
			{"serviceName", "accountId", "serviceInstance", "resourceType"},
			{"serviceName", "accountId", "serviceInstance", "resourceType", "resource"},
		},
		Subject: [][]string{{"id"}, {"id", "scope"}},
	}

	currentCacheKeyInfo.storeCacheKeyPattern(c)

	tests := []struct {
		name        string
		userRequest Request
		subject     Attributes
		resource    Attributes
		actions     []string
		want        *[]CacheKey
	}{
		{
			name: "resource attributes are available in response",
			userRequest: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "gopep-book1",
					"accountId":       "2c17c4e5587783",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			subject: Attributes{
				"id": "IBMid-3100015XDS",
			},
			resource: Attributes{
				"serviceName": "gopep1",
			},
			actions: []string{"gopep.books.write", "gopep.books.read", "gopep.books.burn"},
			want: &[]CacheKey{
				CacheKey("id:IBMid-3100015XDS;serviceName:gopep1;action:gopep.books.write"),
				CacheKey("id:IBMid-3100015XDS;serviceName:gopep1;action:gopep.books.read"),
				CacheKey("id:IBMid-3100015XDS;serviceName:gopep1;action:gopep.books.burn"),
			},
		},
		{
			name: "resource attributes are absent in response",
			userRequest: Request{
				"action": "gopep.books.write",
				"resource": Attributes{
					"serviceName":     "gopep1",
					"serviceInstance": "gopep-book1",
					"accountId":       "2c17c4e5587783",
				},
				"subject": Attributes{
					"id": "IBMid-3100015XDS",
				},
			},
			subject: Attributes{
				"id": "IBMid-3100015XDS",
			},
			actions: []string{"gopep.books.write", "gopep.books.read", "gopep.books.burn"},
			want: &[]CacheKey{
				CacheKey("id:IBMid-3100015XDS;serviceName:gopep1,accountId:2c17c4e5587783,serviceInstance:gopep-book1;action:gopep.books.write"),
				CacheKey("id:IBMid-3100015XDS;serviceName:gopep1,accountId:2c17c4e5587783,serviceInstance:gopep-book1;action:gopep.books.read"),
				CacheKey("id:IBMid-3100015XDS;serviceName:gopep1,accountId:2c17c4e5587783,serviceInstance:gopep-book1;action:gopep.books.burn"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCacheKeysWithUserRequest(tt.userRequest, tt.subject, tt.resource, tt.actions)
			assert.Equal(t, len(*tt.want), len(got), "should contain expected number of cache elements")
			for i, key := range *tt.want {
				if !bytes.Equal(got[i], key) {
					t.Errorf("buildCacheKeysWithUserRequest = %v, want %v", string(got[i]), string(key))
				}
			}
		})
	}
}

func TestBuildSubjectTokenBodyClaims(t *testing.T) {
	pepConfig := &Config{
		Environment:       Staging,
		APIKey:            os.Getenv("API_KEY"),
		DecisionCacheSize: 32,
		LogLevel:          LevelError,
	}

	err := Configure(pepConfig)
	assert.Nil(t, err)

	tests := []struct {
		name  string
		token string
		want  Attributes
	}{
		{
			name:  "User Token",
			token: userToken,
			want: Attributes{
				"id":    "IBMid-270003GUSX",
				"scope": "ibm openid",
			},
		},
		{
			name:  "Service Token",
			token: serviceToken,
			want: Attributes{
				"id":    "iam-ServiceId-c33ad72f-1546-4fd4-8e94-3418ed0fb6e4",
				"scope": "ibm",
			},
		},
		{
			name:  "CRN Token",
			token: crnToken,
			want: Attributes{
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				"cname":           "bluemix",
				"crn":             "crn",
				"ctype":           "public",
				"organizationId":  "",
				"projectId":       "",
				"realmid":         "crn",
				"region":          "",
				"resource":        "",
				"resourceType":    "",
				"scope":           "ibm openid",
				"serviceInstance": "gopep123",
				"serviceName":     "gopep",
				"spaceId":         "",
				"version":         "v1",
			},
		},
		{
			name:  "CRN Token No Realm",
			token: crnTokenNoRealm,
			want: Attributes{
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				"cname":           "bluemix",
				"crn":             "crn",
				"ctype":           "public",
				"organizationId":  "",
				"projectId":       "",
				"realmid":         "",
				"region":          "",
				"resource":        "",
				"resourceType":    "",
				"scope":           "ibm openid",
				"serviceInstance": "instance12345",
				"serviceName":     "commitment-device",
				"spaceId":         "",
				"version":         "v1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenBody, err := getTokenBody(tt.token)
			assert.Nil(t, err)
			got, err := buildSubjectTokenBodyClaims(tokenBody)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got, "should contain expected attributes")
		})
	}
}

func serverErrorMaker(called *int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if *called == 0 {
			return
		} else if *called == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("500 - Something bad happened!"))
			return
		} else if *called == 2 {
			w.WriteHeader(http.StatusNotImplemented)
			_, _ = w.Write([]byte("501 - Something bad happened!"))
			return
		} else if *called == 3 {
			w.WriteHeader(http.StatusOK)
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func compareCachedDecision(a *cache.CachedDecision, b *cache.CachedDecision) bool {
	if a == nil {
		return b == nil
	}

	if b == nil {
		return a == nil
	}

	return a.Permitted == b.Permitted
}

// toAuthzResponse converts the specified cached decisions into an AuthzResponse
func toAuthzResponse(cachedDecisions []cache.CachedDecision, trace string) AuthzResponse {

	authzResponse := AuthzResponse{
		Trace:     trace,
		Decisions: []Decision{},
	}

	for i := range cachedDecisions {
		authzResponse.Decisions = append(authzResponse.Decisions, toDecision(&cachedDecisions[i]))
	}
	return authzResponse
}

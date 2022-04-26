package pep

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessTokenInputNoAdvanceResourceObligations_onlyL1AndL2Caching(t *testing.T) {
	// TTL of 40 millisecond should be enough to run 10 mocked authz requests
	cacheTTL := time.Duration(40) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithoutResourceObligations(),
		getBulkDenyResponder(),
		getBulkDenyResponder(),
		getBulkPermitResponderWithoutResourceObligations(),
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

		t.Run("call #0, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#0-non-cached-permit"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #1, same request", func(t *testing.T) {
			trace := "txid-call#1-same-request"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #2, same request with different action", func(t *testing.T) {
			trace := "txid-call#2-same-request-with-different-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.object.copy",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with different action should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with different action should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #3, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#3-same-request-with-unsupported-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.unsupported.action",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with unsupported action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with unsupported action should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with a set of matching subject attribute claims", func(t *testing.T) {
			trace := "txid-call#4-same-request-with-matching-subject-claims"
			subject := Attributes{
				"id":    "IBMid-270003GUSX",
				"scope": "ibm openid",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with a set of matching subject attribute claims should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a set of matching subject attribute claims should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with modified token body (date changed)", func(t *testing.T) {
			trace := "txid-call#5-same-request-with-matching-subject-claims"
			/* #nosec G101 */
			updatedUserToken := "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwiaWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiNzkyMjg4NTAtZWM1Yy00MWJjLWI4ZjItN2RkMzYzNWFhNjI4IiwiaWRlbnRpZmllciI6IjI3MDAwM0dVU1giLCJnaXZlbl9uYW1lIjoiQWxleCIsImZhbWlseV9uYW1lIjoiSHVkaWNpIiwibmFtZSI6IkFsZXggSHVkaWNpIiwiZW1haWwiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJzdWIiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiIyYzE3YzRlNTU4Nzc4Mzk2MWNlNGEwYWE0MTUwNTRlNyJ9LCJpYXQiOjE1OTk5MDE2MTAsImV4cCI6MTU5OTkwNTIxMCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6cGFzc2NvZGUiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.fMCKyCDbmff8DtlFhbpw84wvZyBSO3k-zAKN494AkN7hvQtHROWPX-TmLpgA170xhYHTNNariOE4c7JitcPdjHinbF6Rro8yUL2HPUZ6To_T8u0xw3yXhR8UhLgHiomUTy2qLY2TK_rUz_R7_yQRUD-91CJwR8iUH8NP4lX0bCQC9o7pxFv8XMNkNSyoyrZYXw2Wfpu0JUK3iZp0Iq9RnAS_7VYJ-Na11tWOugvNvk-PdwWE3OxbyIGL2FDwWbL9kVIv3_oA5mvTI7Zw9vpODaoObstCWFALmGUt7g0t-nE1BbUWOgB5ajkCh1judvKv5tId7OHicIbBVY0bIK3yUw"
			tokenSubject, err := GetSubjectFromToken(updatedUserToken, true)
			require.NoError(t, err)
			assert.NotEqual(t, request["subject"], tokenSubject, "The new token should be updated (and different) from the original")

			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  tokenSubject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with a modified token body should get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a modified token body should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

	}

	runTest(t, config, getAuthzRequestWithTokenBody(t))
}

func TestAccessTokenInputWithAdvanceResourceObligations_allLevelsOfCaching(t *testing.T) {
	// TTL of 100 millisecond should be enough to run 12 mocked authz requests
	cacheTTL := time.Duration(100) * time.Millisecond

	responders := []JSONMockerFunc{
		getBulkPermitResponderWithResourceObligations(),
		getBulkDenyResponder(),
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

		t.Run("call #0, expect a non-cached permit", func(t *testing.T) {
			trace := "txid-call#0-non-cached-permit"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "1st call should get a permit")
			assert.False(t, response.Decisions[0].Cached, "1st call should not hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #1, same request", func(t *testing.T) {
			trace := "txid-call#1-same-request"
			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #2, same request with different action", func(t *testing.T) {
			trace := "txid-call#2-same-request-with-different-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.object.copy",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with different action should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with different action should hit the cache ")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #3, same request with a set of matching subject attribute claims", func(t *testing.T) {
			trace := "txid-call#3-same-request-with-matching-subject-claims"
			subject := Attributes{
				"id":    "IBMid-270003GUSX",
				"scope": "ibm openid",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with a set of matching subject attribute claims should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with a set of matching subject attribute claims should be a cache hit")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #4, same request with modified token body (date changed)", func(t *testing.T) {
			trace := "txid-call#4-same-request-with-matching-subject-claims"
			/* #nosec G101 */
			updatedUserToken := "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwiaWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiNzkyMjg4NTAtZWM1Yy00MWJjLWI4ZjItN2RkMzYzNWFhNjI4IiwiaWRlbnRpZmllciI6IjI3MDAwM0dVU1giLCJnaXZlbl9uYW1lIjoiQWxleCIsImZhbWlseV9uYW1lIjoiSHVkaWNpIiwibmFtZSI6IkFsZXggSHVkaWNpIiwiZW1haWwiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJzdWIiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiIyYzE3YzRlNTU4Nzc4Mzk2MWNlNGEwYWE0MTUwNTRlNyJ9LCJpYXQiOjE1OTk5MDE2MTAsImV4cCI6MTU5OTkwNTIxMCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6cGFzc2NvZGUiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.fMCKyCDbmff8DtlFhbpw84wvZyBSO3k-zAKN494AkN7hvQtHROWPX-TmLpgA170xhYHTNNariOE4c7JitcPdjHinbF6Rro8yUL2HPUZ6To_T8u0xw3yXhR8UhLgHiomUTy2qLY2TK_rUz_R7_yQRUD-91CJwR8iUH8NP4lX0bCQC9o7pxFv8XMNkNSyoyrZYXw2Wfpu0JUK3iZp0Iq9RnAS_7VYJ-Na11tWOugvNvk-PdwWE3OxbyIGL2FDwWbL9kVIv3_oA5mvTI7Zw9vpODaoObstCWFALmGUt7g0t-nE1BbUWOgB5ajkCh1judvKv5tId7OHicIbBVY0bIK3yUw"
			tokenSubject, err := GetSubjectFromToken(updatedUserToken, true)
			require.NoError(t, err)
			assert.NotEqual(t, request["subject"], tokenSubject, "The new token should be updated (and different) from the original")

			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  tokenSubject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "same call with a modified token body should get a permit")
			assert.True(t, response.Decisions[0].Cached, "same call with a modified token body should be a cache hit")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #5, same request with unsupported action", func(t *testing.T) {
			trace := "txid-call#5-same-request-with-unsupported-action"
			requests := &Requests{
				{
					"action":   "cloud-object-storage.unsupported.action",
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with unsupported action should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with unsupported action should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #6, same request with a set of missing scope attribute", func(t *testing.T) {
			trace := "txid-call#6-same-request-with-missing-scope-attribute"
			subject := Attributes{
				"id": "IBMid-270003GUSX",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with a set of missing scope attribute should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a set of missing scope attribute should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #7, same request with a set of different (mismatch) scope attribute value", func(t *testing.T) {
			trace := "txid-call#7-same-request-with-different-scope-attribute"
			subject := Attributes{
				"id":    "IBMid-270003GUSX",
				"scope": "ibm sso",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with a set of different (mismatch) scope attribute value should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a set of different (mismatch) scope attribute value should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #8, same request with a set of missing id attribute", func(t *testing.T) {
			trace := "txid-call#8-same-request-with-missing-id-attribute"
			subject := Attributes{
				"scope": "ibm openid",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with a set of missing id attribute should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a set of missing id attribute should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #9, same request with a set of different (mismatch) id attribute value", func(t *testing.T) {
			trace := "txid-call#9-same-request-with-different-id-attribute"
			subject := Attributes{
				"id":    "IBMid-550003J8QY",
				"scope": "ibm openid",
			}
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  subject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with a set of different (mismatch) id attribute value should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a set of different (mismatch) id attribute value should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #10, same request with a modified token body (different id)", func(t *testing.T) {
			trace := "txid-call#10-same-request-with-modified-token-id"
			/* #nosec G101 */
			updatedUserToken := "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC01NTAwMDNKOFFZIiwiaWQiOiJJQk1pZC01NTAwMDNKOFFZIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiNzkyMjg4NTAtZWM1Yy00MWJjLWI4ZjItN2RkMzYzNWFhNjI4IiwiaWRlbnRpZmllciI6IjI3MDAwM0dVU1giLCJnaXZlbl9uYW1lIjoiQWxleCIsImZhbWlseV9uYW1lIjoiSHVkaWNpIiwibmFtZSI6IkFsZXggSHVkaWNpIiwiZW1haWwiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJzdWIiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiIyYzE3YzRlNTU4Nzc4Mzk2MWNlNGEwYWE0MTUwNTRlNyJ9LCJpYXQiOjE1OTg5MDE2MTAsImV4cCI6MTU5ODkwNTIxMCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6cGFzc2NvZGUiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.fMCKyCDbmff8DtlFhbpw84wvZyBSO3k-zAKN494AkN7hvQtHROWPX-TmLpgA170xhYHTNNariOE4c7JitcPdjHinbF6Rro8yUL2HPUZ6To_T8u0xw3yXhR8UhLgHiomUTy2qLY2TK_rUz_R7_yQRUD-91CJwR8iUH8NP4lX0bCQC9o7pxFv8XMNkNSyoyrZYXw2Wfpu0JUK3iZp0Iq9RnAS_7VYJ-Na11tWOugvNvk-PdwWE3OxbyIGL2FDwWbL9kVIv3_oA5mvTI7Zw9vpODaoObstCWFALmGUt7g0t-nE1BbUWOgB5ajkCh1judvKv5tId7OHicIbBVY0bIK3yUw"
			tokenSubject, err := GetSubjectFromToken(updatedUserToken, true)
			require.NoError(t, err)
			assert.NotEqual(t, request["subject"], tokenSubject, "The new token should be updated (and different) from the original")
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  tokenSubject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with a modified token body (different id) should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a modified token body (different id) should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

		t.Run("call #11, same request with a modified token body (different scope)", func(t *testing.T) {
			trace := "txid-call#11-same-request-with-modified-token-scope"
			/* #nosec G101 */
			updatedUserToken := "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwiaWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiNzkyMjg4NTAtZWM1Yy00MWJjLWI4ZjItN2RkMzYzNWFhNjI4IiwiaWRlbnRpZmllciI6IjI3MDAwM0dVU1giLCJnaXZlbl9uYW1lIjoiQWxleCIsImZhbWlseV9uYW1lIjoiSHVkaWNpIiwibmFtZSI6IkFsZXggSHVkaWNpIiwiZW1haWwiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJzdWIiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiIyYzE3YzRlNTU4Nzc4Mzk2MWNlNGEwYWE0MTUwNTRlNyJ9LCJpYXQiOjE1OTg5MDE2MTAsImV4cCI6MTU5ODkwNTIxMCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6cGFzc2NvZGUiLCJzY29wZSI6ImlibSBzc28iLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.fMCKyCDbmff8DtlFhbpw84wvZyBSO3k-zAKN494AkN7hvQtHROWPX-TmLpgA170xhYHTNNariOE4c7JitcPdjHinbF6Rro8yUL2HPUZ6To_T8u0xw3yXhR8UhLgHiomUTy2qLY2TK_rUz_R7_yQRUD-91CJwR8iUH8NP4lX0bCQC9o7pxFv8XMNkNSyoyrZYXw2Wfpu0JUK3iZp0Iq9RnAS_7VYJ-Na11tWOugvNvk-PdwWE3OxbyIGL2FDwWbL9kVIv3_oA5mvTI7Zw9vpODaoObstCWFALmGUt7g0t-nE1BbUWOgB5ajkCh1judvKv5tId7OHicIbBVY0bIK3yUw"
			tokenSubject, err := GetSubjectFromToken(updatedUserToken, true)
			require.NoError(t, err)
			assert.NotEqual(t, request["subject"], tokenSubject, "The new token should be updated (and different) from the original")
			requests := &Requests{
				{
					"action":   request["action"],
					"resource": request["resource"],
					"subject":  tokenSubject,
				},
			}

			response, err := PerformAuthorization(requests, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.False(t, response.Decisions[0].Permitted, "same call with a modified token body (different scope) should not get a permit")
			assert.False(t, response.Decisions[0].Cached, "same call with a modified token body (different scope) should be a cache miss")
			assert.False(t, response.Decisions[0].Expired, "entry should not expire since expired cache not enabled")
		})

	}

	runTest(t, config, getAuthzRequestWithTokenBody(t))
}

func getAuthzRequestWithTokenBody(t *testing.T) Request {
	/* #nosec G101 */
	userToken := "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwiaWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiNzkyMjg4NTAtZWM1Yy00MWJjLWI4ZjItN2RkMzYzNWFhNjI4IiwiaWRlbnRpZmllciI6IjI3MDAwM0dVU1giLCJnaXZlbl9uYW1lIjoiQWxleCIsImZhbWlseV9uYW1lIjoiSHVkaWNpIiwibmFtZSI6IkFsZXggSHVkaWNpIiwiZW1haWwiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJzdWIiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiIyYzE3YzRlNTU4Nzc4Mzk2MWNlNGEwYWE0MTUwNTRlNyJ9LCJpYXQiOjE1OTg5MDE2MTAsImV4cCI6MTU5ODkwNTIxMCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6cGFzc2NvZGUiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.fMCKyCDbmff8DtlFhbpw84wvZyBSO3k-zAKN494AkN7hvQtHROWPX-TmLpgA170xhYHTNNariOE4c7JitcPdjHinbF6Rro8yUL2HPUZ6To_T8u0xw3yXhR8UhLgHiomUTy2qLY2TK_rUz_R7_yQRUD-91CJwR8iUH8NP4lX0bCQC9o7pxFv8XMNkNSyoyrZYXw2Wfpu0JUK3iZp0Iq9RnAS_7VYJ-Na11tWOugvNvk-PdwWE3OxbyIGL2FDwWbL9kVIv3_oA5mvTI7Zw9vpODaoObstCWFALmGUt7g0t-nE1BbUWOgB5ajkCh1judvKv5tId7OHicIbBVY0bIK3yUw"
	tokenSubject, err := GetSubjectFromToken(userToken, true)
	require.NoError(t, err)
	assert.Equal(t, strings.Split(userToken, ".")[1], tokenSubject["accessTokenBody"])

	action := "cloud-object-storage.object.get"
	resource := Attributes{
		"serviceName": "cloud-object-storage",
		"accountId":   "12345",
	}

	request := Request{
		"action":   action,
		"resource": resource,
		"subject":  tokenSubject,
	}

	return request
}

func getBulkPermitResponderWithoutResourceObligations() JSONMockerFunc {
	mockedJSON := `
	{
		"decisions": [
			{
				"decision": "Permit",
				"obligation": {
					"actions": [
						"cloud-object-storage.object.get",
						"cloud-object-storage.object.put_acl",
						"cloud-object-storage.bucket.get_metrics_monitoring",
						"cloud-object-storage.object.copy_get_version",
						"cloud-object-storage.object.post_md",
						"cloud-object-storage.object.get_tagging",
						"cloud-object-storage.bucket.put_lifecycle",
						"cloud-object-storage.object.copy",
						"cloud-object-storage.object.copy_part",
						"cloud-object-storage.bucket.get_basic",
						"cloud-object-storage.bucket.list_crk_id",
						"cloud-object-storage.object.head"
					],
					"maxCacheAgeSeconds": 600,
					"subject": {
						"attributes": {
							"scope": "ibm openid",
							"id": "IBMid-270003GUSX"
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
	}
	`
	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}
	}
	return responder
}

func getBulkPermitResponderWithResourceObligations() JSONMockerFunc {
	mockedJSON := `
	{
		"decisions": [
			{
				"decision": "Permit",
				"obligation": {
					"actions": [
						"cloud-object-storage.object.get",
						"cloud-object-storage.object.put_acl",
						"cloud-object-storage.bucket.get_metrics_monitoring",
						"cloud-object-storage.object.copy_get_version",
						"cloud-object-storage.object.post_md",
						"cloud-object-storage.object.get_tagging",
						"cloud-object-storage.bucket.put_lifecycle",
						"cloud-object-storage.object.copy",
						"cloud-object-storage.object.copy_part",
						"cloud-object-storage.bucket.get_basic",
						"cloud-object-storage.bucket.list_crk_id",
						"cloud-object-storage.object.head"
					],
					"maxCacheAgeSeconds": 600,
					"subject": {
						"attributes": {
							"scope": "ibm openid",
							"id": "IBMid-270003GUSX"
						}
					},
					"resource": {
						"attributes": {
							"serviceName": "cloud-object-storage",
							"accountId": "12345"
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
	}
	`
	responder := func(req *http.Request) mockedResponse {
		return mockedResponse{
			json:       mockedJSON,
			statusCode: 200,
		}
	}
	return responder
}

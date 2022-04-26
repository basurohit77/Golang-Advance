package pep

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//NewAuthzMocker returns *PDPMocker that mocks the v2/authz using the specified mockedResponses
func NewAuthzMocker(t *testing.T, responders []JSONMockerFunc) *PDPMocker {
	option := mockedData{
		responders: responders,
		bulk:       false,
	}
	return newMocker(t, option)
}

//NewBulkMocker returns *PDPMocker that mocks the v2/authz/bulk using the specified mockedResponses
func NewBulkMocker(t *testing.T, responders []JSONMockerFunc) *PDPMocker {
	option := mockedData{
		responders: responders,
		bulk:       true,
	}
	return newMocker(t, option)
}

//1st call, we expect a non-cached permit
//With the same call, we expect a cached permit
//With cached actions, we expecte a cached permit
//With non-cached action, we expect non-cached deny
//
//Note: Keep this function simple, i.e.
// 1 - DO NOT make this function difficult to read/debug.
// 2 - DO NOT attempt to combine wth runDenyTest
// 3 - copy and modify if needed
func runPermitTest(t *testing.T, config Config, request Request, actions []string, nonCachedAction string) {

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{request}
	t.Run("1st call, we expect a non-cached permit", func(t *testing.T) {
		trace := "txid-permit-1st-call"
		response, err := PerformAuthorization(requests, trace)
		require.NoError(t, err)

		require.NotEmpty(t, response.Decisions)
		assert.True(t, response.Decisions[0].Permitted, "1st call decision")
		assert.False(t, response.Decisions[0].Cached, "1st call cache hit")
	})

	t.Run("2nd call, we expect a cached permit", func(t *testing.T) {
		trace := "txid-cached-permit-2nd-call"
		response, err := PerformAuthorization(requests, trace)
		require.NoError(t, err)

		require.NotEmpty(t, response.Decisions)
		assert.True(t, response.Decisions[0].Permitted, "2nd call decision")
		assert.True(t, response.Decisions[0].Cached, "2nd call cache hit")
	})

	t.Run("with cached actions, we expect a cached permit", func(t *testing.T) {
		for _, action := range actions {

			reqs := &Requests{
				{
					"action":   action,
					"resource": request["resource"],
					"subject":  request["subject"],
				},
			}

			trace := fmt.Sprintf("txid-permit-cached-action-%s", action)
			response, err := PerformAuthorization(reqs, trace)
			require.NoError(t, err)

			require.NotEmpty(t, response.Decisions)
			assert.True(t, response.Decisions[0].Permitted, "decision from cached actions")

			assert.True(t, response.Decisions[0].Cached, "cache hit for cached actions")
		}
	})

	t.Run("with non-cached action, we expect non-cached deny", func(t *testing.T) {
		rNoCachedAction := Requests{
			{
				"action":   nonCachedAction,
				"resource": request["resource"],
				"subject":  request["subject"],
			},
		}
		trace := fmt.Sprintf("txid-deny-non-cached-action-%s", nonCachedAction)
		response, err := PerformAuthorization(&rNoCachedAction, trace)
		require.NoError(t, err)

		require.NotEmpty(t, response.Decisions)
		assert.False(t, response.Decisions[0].Permitted, "decision for unknown non-cached action")
		assert.False(t, response.Decisions[0].Cached, "no cache hit for non-cache action")
		assert.Equal(t, response.Decisions[0].Reason, DenyReasonIAM, "expect DenyReasonIAM")
	})
}

// Run tests similar to runPermitTest but for deny cases
func runDenyTest(t *testing.T, config Config, request Request, actions []string, nonCachedAction string) {

	err := Configure(&config)
	require.NoError(t, err)

	requests := &Requests{request}
	t.Run("1st call, we expect a non-cached deny", func(t *testing.T) {
		trace := "txid-deny-1st-call"
		response, err := PerformAuthorization(requests, trace)
		require.NoError(t, err)

		require.NotEmpty(t, response.Decisions)
		assert.False(t, response.Decisions[0].Permitted, "1st call decision, expect a deny")
		assert.False(t, response.Decisions[0].Cached, "1st call, expect no cache hit")
		assert.Equal(t, response.Decisions[0].Reason, DenyReasonIAM, "1st call, expect DenyReasonIAM")
	})

	t.Run("2nd call, we expect a cached deny", func(t *testing.T) {
		trace := "txid-cached-deny-2nd-call"
		response, err := PerformAuthorization(requests, trace)
		require.NoError(t, err)

		require.NotEmpty(t, response.Decisions)
		assert.False(t, response.Decisions[0].Permitted, "2nd call decision, expect a deny")
		assert.True(t, response.Decisions[0].Cached, "2nd call, expect a cache hit")
		assert.Equal(t, response.Decisions[0].Reason, DenyReasonIAM, "2nd call, expect DenyReasonIAM")
	})

	t.Run("3rd call, we expect a non-cached deny with DenyReasonContext", func(t *testing.T) {
		rNoCachedAction := Requests{
			{
				"action":   nonCachedAction,
				"resource": request["resource"],
				"subject":  request["subject"],
			},
		}

		trace := fmt.Sprintf("txid-deny-non-cached-action-%s", nonCachedAction)
		response, err := PerformAuthorization(&rNoCachedAction, trace)
		require.NoError(t, err)

		require.NotEmpty(t, response.Decisions)
		assert.False(t, response.Decisions[0].Permitted, "3rd call decision, expect a deny")
		assert.False(t, response.Decisions[0].Cached, "3rd call, expect no cache hit")
		assert.Equal(t, response.Decisions[0].Reason, DenyReasonContext, "3rd call, expect DenyReasonContext")
	})
}

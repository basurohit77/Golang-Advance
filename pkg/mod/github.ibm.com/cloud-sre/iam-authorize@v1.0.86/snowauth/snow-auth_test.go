package snowauth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
)

type mockNRTransaction struct {
	newrelicPreModules.Transaction
}

// only using AddAttribute in this lib at this point
func (nr mockNRTransaction) AddAttribute(key string, value interface{}) error {
	return nil
}

const svc1, svc2, svc3 = "svc1", "svc2", "svc3"

var (
	snAuth          *SnowAuth
	mockSNowAuthAPI *httptest.Server

	email        = "email@ibm.com"
	nonCloudUser = "noncloud@ibm.com"
	invalidUser  = "invalid@ibm.com"
	crns         = []string{"crn:v1:bluemix:public:" + svc1 + ":::::", "crn:v1:internal:dedicated:" + svc2 + ":::::", "crn:v1:::" + svc3 + ":::::"}
)

func TestMain(m *testing.M) {

	mockSNowAuthAPI = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data = snowAuthResp{snowAuthRespResult: snowAuthRespResult{Result: make([]result, 3)}}
		res := result{
			UserName:    email,
			CRN:         crns[1],
			UserType:    "cloud",
			ServiceType: "noncloud",
			Authorized: authorized{
				Valid:   false,
				Message: "some msg",
			},
		}

		data.snowAuthRespResult.Result[0] = res

		res = result{
			UserName:    nonCloudUser,
			CRN:         crns[0],
			UserType:    "noncloud",
			ServiceType: "cloud",
			Authorized: authorized{
				Valid:   true,
				Message: "some msg",
			},
		}

		data.snowAuthRespResult.Result[1] = res

		res = result{
			UserName:    invalidUser,
			CRN:         crns[2],
			UserType:    "NO_TYPE",
			ServiceType: "false",
			Authorized: authorized{
				Valid:   false,
				Message: "some msg",
			},
		}

		data.snowAuthRespResult.Result[2] = res

		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Fatal(err)
		}

	}))

	defer mockSNowAuthAPI.Close()

	snAuth = NewSNowAuth(mockSNowAuthAPI.URL, "snowtoken")
	snAuth.Cache.CleanupInterval = 1 * time.Second
	snAuth.Cache.ExpirationTime = 1 * time.Second

	os.Exit(m.Run())
}

func TestSNowAuthorizationBypassFlagTrue(t *testing.T) {

	ByPassFlag = true

	ctx := context.Background()
	ctx = newrelicPreModules.NewContext(ctx, mockNRTransaction{})
	ctx = newrelic.NewContext(ctx, &newrelic.Transaction{})

	snAuthSvcs, err := snAuth.SNowAuthorization(ctx, email, crns)
	if err != nil {
		t.Error(err)
	}

	// svc1 should return true because its CRN is public bluemix
	assert.Equal(t, true, snAuthSvcs[crns[0]])

	// svc2 should return true now because the bypass flag is on
	assert.Equal(t, true, snAuthSvcs[crns[1]])
}

func TestSNowAuthorizationBypassFlagFalse(t *testing.T) {

	ByPassFlag = false

	// just enough time for the cache to cleanup
	time.Sleep(2 * time.Second)

	ctx := context.Background()
	ctx = newrelicPreModules.NewContext(ctx, mockNRTransaction{})
	ctx = newrelic.NewContext(ctx, &newrelic.Transaction{})

	snAuthSvcs, err := snAuth.SNowAuthorization(ctx, email, crns)
	if err != nil {
		t.Error(err)
	}

	// svc1 should return true because its CRN is public bluemix
	assert.Equal(t, true, snAuthSvcs[crns[0]])

	// svc2 should return false because its CRN is not public bluemix
	// this will call SNow Auth API, which returns false
	assert.Equal(t, false, snAuthSvcs[crns[1]])

}

func TestNewSNowAuthEmptyURLToken(t *testing.T) {

	snAuth = NewSNowAuth("", "")

	assert.Equal(t, &SnowAuth{Cache: snAuth.Cache}, snAuth)

}

func TestSNowAuthorizationEmptyEmailCRNs(t *testing.T) {

	snAuth = NewSNowAuth("invalid.snow.url", "snowtoken")

	ctx := context.Background()
	ctx = newrelicPreModules.NewContext(ctx, mockNRTransaction{})
	ctx = newrelic.NewContext(ctx, &newrelic.Transaction{})

	snAuthSvcs, err := snAuth.SNowAuthorization(ctx, "", []string{})
	if err == nil {
		t.Error("should have returned no email error")
	}

	assert.Equal(t, errNoEmail, err.Error())
	assert.Equal(t, snAuthSvcs, make(SvcsAuth))

	snAuthSvcs, err = snAuth.SNowAuthorization(ctx, email, []string{})
	if err == nil {
		t.Error("should have returned no crns error")
	}

	assert.Equal(t, errNoCRN, err.Error())
	assert.Equal(t, snAuthSvcs, make(SvcsAuth))

}

func TestGetUserType(t *testing.T) {

	snAuth = NewSNowAuth(mockSNowAuthAPI.URL, "snowtoken")

	assert.Equal(t, "cloud", snAuth.GetUserType(email, crns))
	assert.Equal(t, "noncloud", snAuth.GetUserType(nonCloudUser, crns))
	assert.Equal(t, "NO_TYPE", snAuth.GetUserType(invalidUser, crns))

}

func TestGetServiceType(t *testing.T) {

	snAuth = NewSNowAuth(mockSNowAuthAPI.URL, "snowtoken")

	assert.Equal(t, "cloud", snAuth.GetServiceType(svc1, email))
	assert.Equal(t, "noncloud", snAuth.GetServiceType(svc2, email))
	assert.Equal(t, "false", snAuth.GetServiceType(svc3, email))

}

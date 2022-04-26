package pep

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PerformAuthorizationForAuthzShouldSucceedIntegration(t *testing.T) {
	t.Skip("Not implemented yet")
	// Create an AuthzRequest manually
	resource := map[string]interface{}{
		"ServiceName": "myService",
		"AccountID":   "myAccountID",
	}

	subject := map[string]interface{}{
		"iamID": "myID",
	}

	authzRequest := Requests{
		{
			"Scope":    "a/acountid",
			"Resource": resource,
			"Subject":  subject,
		},
	}

	trace := "12345"
	// Ask the pep to perform authz check
	authzResponse, err := PerformAuthorization(&authzRequest, trace)

	if err != nil {
		t.Errorf("%+v\n", err)
	}

	if !authzResponse.Decisions[0].Permitted {
		t.Errorf("PerformAuthorization returns %v; want true", authzResponse.Decisions[0].Permitted)
	}
}

func Test_PerformAuthorizationForAuthzShouldFailIntegration(t *testing.T) {

	t.Skip("Not implemented yet")
	// Create an AuthzRequest manually
	subject := map[string]interface{}{
		"iamID": "myID",
	}

	authzRequest := Requests{
		{
			"Scope":   "a/acountid",
			"Subject": subject,
		},
	}

	trace := "12345"

	// Ask the pep to perform authz check
	authzResponse, err := PerformAuthorization(&authzRequest, trace)

	if err != nil {
		t.Errorf("%+v\n", err)
	}

	if authzResponse.Decisions[0].Permitted {
		t.Errorf("PerformAuthorization returns %v; want false", authzResponse.Decisions[0].Permitted)
	}

}

func Test_PerformAuthorizationForListAPIShouldSucceedIntegration(t *testing.T) {
	t.Skip("Not implemented yet")
}

func Test_PerformAuthorizationForListAPIShouldFailIntegration(t *testing.T) {
	t.Skip("Not implemented yet")
}

func Test_GetTokenAndClaims(t *testing.T) {
	pc := &Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
	}

	err := Configure(pc)

	assert.Nil(t, err)

	tk, err := GetToken()
	assert.Nil(t, err)

	claims, err := GetClaims(tk, false)

	assert.Nil(t, err)
	assert.NotNil(t, claims)

	IAMIDsubject, err := GetSubjectAsIAMIDClaim(tk, false)

	assert.Nil(t, err)
	assert.NotEmpty(t, IAMIDsubject)

	assert.Equal(t, IAMIDsubject, claims.IAMID)
}

func Test_InvalidTokenFetch(t *testing.T) {
	invalidToken := []byte(`eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6MTU1NzI2OTQ5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.WP6VoxIdOe-LLsvFcYoQ4vtc8Q3pDZ7irqF2wJ2rLRmf6_XJ2ID6YObr_fURBWghQK1APwJNY97xUz2pWBaicqOHBPz66mi_AqC6NMWStkyhYNAevVar1AvwpfBUzK_P9MrgW5XV3UjchnIiibjkFx1Dpz0x98x7zt3g1t0-5_OynvaOZ2beZMKOaiVy93jKSPylcjS2aNTjAvu0axXlw3Aib5vwpwnVOCtwFaJo0H-Gq4GwC4YsRmoyWLXwSpyBqdY-Wdvh6fJFpBW_fiSbhrdSnvzsYCl0xeC3-wwqmA3y4nOShlferKHiH4HIn47EqOMRpuswOlcJyqdmppgztQ`)

	invalidClaims, err := GetClaims(string(invalidToken), false)

	assert.NotNil(t, err)
	assert.Nil(t, invalidClaims)

	invalidIAMIDsubject, err := GetSubjectAsIAMIDClaim(string(invalidToken), false)

	assert.NotNil(t, err)
	assert.Empty(t, invalidIAMIDsubject)
}

func Test_ParseCRNToAttribute(t *testing.T) {
	crn := "test-crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:foo:barResType:barRes"

	attributes := ParseCRNToAttributes(crn)

	assert.Equal(t, "test", attributes["realmid"])
	assert.Equal(t, "crn", attributes["crn"])
	assert.Equal(t, "v1", attributes["version"])
	assert.Equal(t, "staging", attributes["cname"])
	assert.Equal(t, "public", attributes["ctype"])
	assert.Equal(t, "gopep", attributes["serviceName"])
	assert.Equal(t, "foo", attributes["serviceInstance"])
	assert.Equal(t, "", attributes["region"])
	assert.Equal(t, "barResType", attributes["resourceType"])
	assert.Equal(t, "barRes", attributes["resource"])
	assert.Equal(t, "", attributes["spaceId"])
	assert.Equal(t, "2c17c4e5587783961ce4a0aa415054e7", attributes["accountId"])
	assert.Equal(t, "", attributes["organizationId"])
	assert.Equal(t, "", attributes["projectId"])
}

func Test_GetSubjectFromToken(t *testing.T) {
	subject, err := GetSubjectFromToken(crnToken, true)

	assert.Nil(t, err)
	assert.Equal(t, strings.Split(crnToken, ".")[1], subject["accessTokenBody"])

	subject, err = GetSubjectFromToken(userToken, true)
	assert.Nil(t, err)
	assert.Equal(t, strings.Split(userToken, ".")[1], subject["accessTokenBody"])

	subject, err = GetSubjectFromToken(serviceToken, true)
	assert.Nil(t, err)
	assert.Equal(t, strings.Split(serviceToken, ".")[1], subject["accessTokenBody"])
}

func Test_reportStatisticsStats(t *testing.T) {
	conf := &Config{}

	stats := conf.reportStatisticsStats()

	assert.Equal(t, Stats{}, stats)
}

func Test_PerformAuthorizationNoAPIKey(t *testing.T) {
	pc := &Config{
		Environment: Staging,
		LogLevel:    LevelError,
	}

	err := Configure(pc)

	assert.Nil(t, err)

	subject := map[string]interface{}{
		"iamID": "myID",
	}
	authzRequest := Requests{
		{
			"Scope":   "a/acountid",
			"Subject": subject,
		},
	}
	trace := "12345"
	// Ask the pep to perform authz check
	authzResponse, err := PerformAuthorization(&authzRequest, trace)

	assert.Equal(t, AuthzResponse{}, authzResponse)
	assert.NotNil(t, err)
	assert.Equal(t, "PEP is not initialized with an API Key", err.Error())
}

func Test_PerformAuthorizationWithToken_EmptyToken(t *testing.T) {
	pc := &Config{
		Environment: Staging,
		LogLevel:    LevelError,
	}

	err := Configure(pc)

	assert.Nil(t, err)

	subject := map[string]interface{}{
		"iamID": "myID",
	}
	authzRequest := Requests{
		{
			"Scope":   "a/acountid",
			"Subject": subject,
		},
	}
	trace := "12345"
	// Ask the pep to perform authz check
	authzResponse, err := PerformAuthorizationWithToken(&authzRequest, trace, "")

	assert.Equal(t, AuthzResponse{}, authzResponse)
	assert.NotNil(t, err)
	assert.Equal(t, "authorization token is required", err.Error())
}

func Test_GetAuthorizedRolesNoAPIKey(t *testing.T) {
	pc := &Config{
		Environment: Staging,
		LogLevel:    LevelError,
	}

	err := Configure(pc)

	assert.Nil(t, err)

	subject := map[string]interface{}{
		"iamID": "myID",
	}
	request := Requests{
		{
			"Scope":   "a/acountid",
			"Subject": subject,
		},
	}
	trace := "12345"

	response, err := GetAuthorizedRoles(&request, trace)

	assert.Equal(t, []RolesResponse{}, response)
	assert.NotNil(t, err)
	assert.Equal(t, "PEP is not initialized with an API Key", err.Error())
}

func Test_GetAuthorizedRolesWithToken_EmptyToken(t *testing.T) {
	pc := &Config{
		Environment: Staging,
		LogLevel:    LevelError,
	}

	err := Configure(pc)

	assert.Nil(t, err)

	subject := map[string]interface{}{
		"iamID": "myID",
	}
	request := Requests{
		{
			"Scope":   "a/acountid",
			"Subject": subject,
		},
	}
	trace := "12345"

	response, err := GetAuthorizedRolesWithToken(&request, trace, "")

	assert.Equal(t, []RolesResponse{}, response)
	assert.NotNil(t, err)
	assert.Equal(t, "authorization token is required", err.Error())
}

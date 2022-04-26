package common

import (
	"testing"
)

var (
	//fake apikey
	apikey              = "nhgchkde-bgt67cQPmxLNzBLamzF0Ox5ZWNUzcbLL34e"
	accountID           = "accountID1"
	allowedAccessGroups = "testGroup1"
)

func TestApiKeyToToken(t *testing.T) {
	accessToken, err := apiKeyToToken(apikey)
	if err != nil {
		t.Error(err)
	}
	t.Log(accessToken)
}

func TestGetUserAccessGroups(t *testing.T) {
	accessToken, _ := apiKeyToToken(apikey)
	groups, err := getUserAccessGroups(accessToken, accountID)
	if err != nil {
		t.Error(err)
	}
	t.Log(groups)
}

func TestVerifyDeveloperGroup(t *testing.T) {
	valid, err := VerifyUser(apikey, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}

func TestVerifyDeveloperGroupBearerAccessToken(t *testing.T) {
	accessToken, _ := apiKeyToToken(apikey)
	valid, err := VerifyUser("Bearer "+accessToken, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}

func TestVerifyAdminGroup(t *testing.T) {
	valid, err := VerifyUser(apikey, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}

func TestVerifyAdminGroupBearerToken(t *testing.T) {
	accessToken, _ := apiKeyToToken(apikey)
	valid, err := VerifyUser("Bearer "+accessToken, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}
func TestVerifySchematicsAdminSSHAdminGroups(t *testing.T) {
	valid, err := VerifyUser(apikey, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}

func TestVerifySchematicsAdminSSHAdminGroupsBearerToken(t *testing.T) {
	accessToken, _ := apiKeyToToken(apikey)
	valid, err := VerifyUser("Bearer "+accessToken, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}
func TestVerifyEmptyGroup(t *testing.T) {
	valid, err := VerifyUser(apikey, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}

func TestVerifyEmptyGroupAccessTokenBearer(t *testing.T) {
	apikey = ""
	accessToken, _ := apiKeyToToken(apikey)
	valid, err := VerifyUser("Bearer "+accessToken, accountID, allowedAccessGroups)
	if err != nil {
		t.Error(err)
	}
	t.Log(valid)
}

func TestVerifyEmptyApikey(t *testing.T) {
	apikey = ""
	valid, _ := VerifyUser("", accountID, allowedAccessGroups)
	if !valid {
		t.Log("should be authorized")
	}
}

func TestVerifyInvalidApikey(t *testing.T) {
	valid, _ := VerifyUser("key", accountID, allowedAccessGroups)
	if !valid {
		t.Log("should be authorized")
	}
}

func TestVerifyInvalidkey(t *testing.T) {
	valid, err := VerifyUser("key", accountID, allowedAccessGroups)
	if !valid || err != nil {
		t.Log("should be authorized")
	}
}

func TestCheckGroups(t *testing.T) {
	gd := groupData{
		Limit:      10,
		Offset:     2,
		TotalCount: 20,
		Groups: []Group{
			{
				ID: "testGroup",
			},
		},
	}
	valid := checkGroups("testGroup", &gd)
	if !valid {
		t.Log("allowed access groups are not in user access groups")
	}
}

func TestCheckGroupsInvalid(t *testing.T) {
	gd := groupData{
		Limit:      10,
		Offset:     2,
		TotalCount: 20,
		Groups: []Group{
			{
				ID: "group",
			},
		},
	}
	valid := checkGroups("testGroup", &gd)
	if !valid {
		t.Log("allowed access groups are not in user access groups")
	}
}

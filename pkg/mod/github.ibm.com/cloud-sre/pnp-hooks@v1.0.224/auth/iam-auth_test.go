package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	iammock "github.ibm.com/cloud-sre/common-mocks/iam"
)

// TestNewIAMAuth tests the creation of an IAMAuth instance and a call to IsIAMAuthorized
func TestNewIAMAuth(t *testing.T) {
	iammock.SetupIAMMockCustomHandler(iammock.GenericHandler)
	os.Setenv(`IAM_MODE`, `test`)
	os.Setenv(`IAM_URL`, iammock.MockIAM.URL)

	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)

	iamAuth := NewIAMAuth(nil)
	auth, err := iamAuth.IsIAMAuthorized(request, nil)
	if auth != nil {
		t.Error("Expected auth to not be nil, but it was nil")
	}
	if err == nil {
		t.Error("Expected error but did not receive one")
	}
}

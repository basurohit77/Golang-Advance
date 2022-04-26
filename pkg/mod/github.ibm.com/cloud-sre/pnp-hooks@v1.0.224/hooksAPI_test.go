package main

import (
	"os"
	"testing"

	"github.ibm.com/cloud-sre/pnp-hooks/auth"
	iammock "github.ibm.com/cloud-sre/common-mocks/iam"
)

func TestCreateHealthzHandler(t *testing.T) {
	s := createServerObject()
	h := createHealthzHandler(s)

	if h.APIRequestPathPrefix != "pnphooks" {
		t.Fatal("APIRequestPathPrefix is not hooks")
	}
}

func TestConnectRabbitMQ(t *testing.T) {
	connectRabbitmq()
}

func TestSetupHandlers(t *testing.T) {
	iammock.SetupIAMMockCustomHandler(iammock.GenericHandler)
	os.Setenv(`IAM_MODE`, `test`)
	os.Setenv(`IAM_URL`, iammock.MockIAM.URL)

	s := createServerObject()

	iamAuth := auth.NewIAMAuth(nil)
	authInstance := auth.NewAuth(iamAuth)

	setupHandlers(s, authInstance)
}

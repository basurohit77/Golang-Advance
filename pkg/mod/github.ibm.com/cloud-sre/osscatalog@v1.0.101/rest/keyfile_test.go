package rest

import (
	"os"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

//var debugFlag = flag.Bool("debug", false, "-debug: enable debugging information")

func TestGetKey(t *testing.T) {
	err := LoadKeyFile("testdata/dummy-keyfile.key")
	if err != nil {
		t.Fatalf("Error in LoadKeyFile(): %v", err)
	}
	key, err := GetKey("catalog-yp")
	if err != nil {
		t.Fatalf("Error getting key: %v", err)
	}
	testhelper.AssertEqual(t, "Key value", "key-value-for-catalog-yp", key)
}

func TestGetTokenFromFile(t *testing.T) {
	err := LoadKeyFile("testdata/dummy-keyfile.key")
	if err != nil {
		t.Fatalf("Error in LoadKeyFile(): %v", err)
	}
	token, err := GetToken("servicenow-watsondev")
	if err != nil {
		t.Fatalf("Error getting token from file: %v", err)
	}
	testhelper.AssertEqual(t, "Token value", "token-value-for-servicenow-watsondev", token)
}

func TestGetTokenFromIAM(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestGetTokenFromIAM() in short mode")
	}
	if *testhelper.TestDebugFlag {
		debug.SetDebugFlags(debug.IAM)
	}

	err := LoadDefaultKeyFile()
	if err != nil {
		t.Fatalf("Error in LoadKeyFile(): %v", err)
	}
	token, err := GetToken("catalog-ys1")
	if err != nil {
		t.Fatalf("Error getting token from IAM: %v", err)
	}
	if token == "" {
		t.Fatalf("Token is empty")
	}
}

func TestLoadEnvironmentSecrets(t *testing.T) {
	os.Setenv("SECRETKEY_key1", "key1_key")
	os.Setenv("SECRETID_key1", "key1_id")
	os.Setenv("SECRETTOKEN_key1", "key1_token")
	os.Setenv("SECRETCREDENTIALS_key2", `{"name": "key2", "key": "key2_key"}`)

	err := LoadEnvironmentSecrets()
	if err != nil {
		t.Fatalf("Error in LoadEnvironmentSecrets(): %v", err)
	}

	val, err := GetKey("key1")
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "Key", "key1_key", val)

	val, err = GetID("key1")
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "Key", "key1_id", val)

	val, err = GetToken("key1")
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "Key", "key1_token", val)

	val, err = GetCredentialsString("key2", "key")
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "Creds.key", "key2_key", val)

	val, err = GetKey("key2")
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "Key", "key2_key", val)
}

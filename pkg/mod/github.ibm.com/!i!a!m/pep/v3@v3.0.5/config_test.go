package pep_test

import (
	crand "crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	io "io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	guuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gojwks "github.ibm.com/IAM/go-jwks"
	. "github.ibm.com/IAM/pep/v3"
	token "github.ibm.com/IAM/token/v5"
)

func TestConfigureClientID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(mockIdentityService(3600)))
	defer ts.Close()

	config := &Config{
		Environment:   Custom,
		APIKey:        "fake",
		ClientID:      "testval1",
		ClientSecret:  "testval2",
		LogOutput:     io.Discard,
		TokenEndpoint: ts.URL + token.TokenPath,
		KeyEndpoint:   ts.URL + token.KeyPath,
		AuthzEndpoint: ts.URL,
		ListEndpoint:  ts.URL,
		RolesEndpoint: ts.URL,
	}

	err := Configure(config)

	if err != nil {
		t.Errorf("Configure() error = %v", err)
		t.FailNow()
	}

	pepConf := GetConfig().(*Config)

	if pepConf.ClientID != "testval1" {
		t.Errorf("PEP config does not contain expected ClientId")
	}

	if pepConf.ClientSecret != "testval2" {
		t.Errorf("PEP config does not contain expected ClientSecret")
	}
}

func TestDeploymentEnvironment_String(t *testing.T) {
	tests := []struct {
		name string
		d    DeploymentEnvironment
		want string
	}{
		{name: "Staging", d: Staging, want: "Staging"},
		{name: "Production", d: Production, want: "Production"},
		{name: "Custom", d: Custom, want: "Custom"},
		{name: "PrivateProduction", d: PrivateProduction, want: "PrivateProduction"},
		{name: "PrivateStaging", d: PrivateStaging, want: "PrivateStaging"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("DeploymentEnvironment.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Can be re-enabled once custom http client feature is finished
/*func TestDefaultConfig(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters

		if req.URL.String() == (token.ProdHostURL + token.TokenPath) {
			token, err := tokenGen(3600)
			assert.Nil(t, err)

			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBuffer(token)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		} else if req.URL.String() == (token.ProdHostURL + token.TokenPath) {
			verifyKeys, err := createKey()
			assert.Nil(t, err)

			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBuffer(verifyKeys)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		} else {
			return &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`Big Error`)),
				Header:     make(http.Header),
			}
		}
	})

	conf := &Config{APIKey: "dummy key", LogOutput: io.Discard, HTTPClient: client}

	err := Configure(conf)
	assert.Nil(t, err)

	pepConf := GetConfig().(*Config)

	assert.Equal(t, pepConf.Environment, Production)
}*/

func TestConfigure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(mockIdentityService(3600)))
	defer ts.Close()

	type args struct {
		c *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		/*{// Can be re-enabled once custom http client feature is finished
			name: "Production (default)",
			args: args{&Config{APIKey: "dummy key", LogOutput: io.Discard, TokenEndpoint: ts.URL + token.TokenPath,
				KeyEndpoint:   ts.URL + token.KeyPath,
				AuthzEndpoint: ts.URL,
				ListEndpoint:  ts.URL,
				RolesEndpoint: ts.URL}},
			wantErr: false,
		},*/
		{
			name:    "Staging",
			args:    args{&Config{Environment: Staging, APIKey: os.Getenv("API_KEY"), LogOutput: io.Discard}},
			wantErr: false,
		},
		/*{
			name:    "Production",
			args:    args{&Config{Environment: Production, APIKey: "dummy key", LogOutput: io.Discard}},
			wantErr: false,
		},*/
		{
			name: "Custom missing list and token endpoints", args: args{
				&Config{
					Environment: Custom, APIKey: "dummy key",
					AuthzEndpoint: "http://localhost/authz",
					LogOutput:     io.Discard,
				},
			},
			wantErr: true,
		},
		{
			name: "Custom missing token endpoint", args: args{
				&Config{
					Environment: Custom, APIKey: "dummy key",
					AuthzEndpoint: "http://localhost/authz",
					ListEndpoint:  "http://localhost/list",
					LogOutput:     io.Discard,
				},
			},
			wantErr: true,
		},
		{
			name: "Custom", args: args{
				&Config{
					Environment: Custom, APIKey: "dummy key",
					TokenEndpoint: ts.URL + token.TokenPath,
					KeyEndpoint:   ts.URL + token.KeyPath,
					AuthzEndpoint: ts.URL,
					ListEndpoint:  ts.URL,
					RolesEndpoint: ts.URL,
					LogOutput:     io.Discard,
				},
			},
			wantErr: false,
		},
		{
			name:    "Custom missing all endpoints",
			args:    args{&Config{Environment: Custom, APIKey: "dummy key", LogOutput: io.Discard}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Configure(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigureHttpClient(t *testing.T) {
	transport := &http.Transport{}

	timeout := time.Duration(15) * time.Second
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	config := &Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		HTTPClient:  client,
		LogOutput:   io.Discard,
	}

	err := Configure(config)

	if err != nil {
		t.Errorf("Configure() error = %v", err)
	}

	pepConf := GetConfig().(*Config)

	if pepConf.HTTPClient.Timeout != timeout {
		t.Errorf("PEP http client does not contain expected timeout")
	}

	if pepConf.HTTPClient.Transport != transport {
		t.Errorf("PEP http client does not contain expected transport")
	}

	if pepConf.HTTPClient != client {
		t.Errorf("PEP http client is not the supplied one")
	}
}

func TestCustomConfigInvalidEndpoints(t *testing.T) {
	config := &Config{
		Environment:   Custom,
		APIKey:        os.Getenv("API_KEY"),
		LogLevel:      LevelError,
		AuthzEndpoint: "http://mocked",
		ListEndpoint:  "http://mocked",
		TokenEndpoint: "http://mocked",
		KeyEndpoint:   "http://mocked",
	}

	err := Configure(config)

	assert.Contains(t, err.Error(), "no such host")
}

func TestConfigureDisableCache(t *testing.T) {
	config := &Config{
		Environment:  Staging,
		APIKey:       os.Getenv("API_KEY"),
		DisableCache: true,
	}

	err := Configure(config)
	assert.Nil(t, err)

	stats := GetStatistics()

	assert.NotNil(t, stats)
	assert.NotNil(t, stats.Usage)
	assert.NotNil(t, stats.Cache)
	assert.Equal(t, int(0), stats.Usage.OriginalRequestsToPDP, "Expect zero request stored in the usage stats")
	assert.Equal(t, uint64(0), stats.Cache.EntriesCount, "Expect zero entries stored in the cache")
}

func mockIdentityService(expiryTime int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == token.TokenPath {
			// path should be '/identity/keys'
			w.Header().Add("Content-Type", "application/json")

			tokenString, err := tokenGen(expiryTime)
			if err != nil {
				fmt.Println(err)
			}
			_, _ = w.Write([]byte(tokenString))
		} else if r.URL.String() == token.KeyPath {
			verifyKeys, err := createKey()

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Error while parsing key Token")
				log.Printf("Token Signing error: %v\n", err)
				return
			}

			verifyKeysObj := &gojwks.Keys{}

			err = json.Unmarshal(verifyKeys, verifyKeysObj)

			if err != nil {
				fmt.Println(err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Connection", "keep-alive")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(verifyKeysObj)

			if err != nil {
				fmt.Println(err)
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("400 - Invalid path in test."))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

type iKeys struct {
	Keys []iKey `json:"Keys"`
}
type iKey struct {
	Kty string `json:"Kty"`
	N   string `json:"N"`
	E   string `json:"E"`
	Alg string `json:"Alg"`
	Kid string `json:"Kid"`
}

var reader = crand.Reader
var bitSize = 2048

var key, _ = rsa.GenerateKey(reader, bitSize)

var publicKey *rsa.PublicKey = &key.PublicKey

func tokenGen(expirySeconds int) ([]byte, error) {

	t := jwt.New(jwt.SigningMethodRS256) //(jwt.GetSigningMethod("RS256"))
	t.Header["kid"] = "20190122"
	now := jwt.TimeFunc().Unix()
	expiresAt := now + int64(expirySeconds)

	// set our claims
	standardClaims := &jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Id:        guuid.New().String(),
		IssuedAt:  now,
		Issuer:    "https://iam.test.cloud.ibm.com/identity",

		Subject: "ServiceID-1234",
	}

	account := token.Account{
		Valid: true,
		Bss:   "00000000000000000000000000000000",
	}

	t.Claims = &token.IAMAccessTokenClaims{
		// see http://tools.ietf.org/html/draft-ietf-oauth-json-web-token-20#section-4.1.4
		StandardClaims: standardClaims,
		IAMID:          "iam-ServiceID-1234",                     // IAMID
		ID:             "iam-ServiceID-1234",                     // ID
		RealmID:        "iam",                                    // RealmID
		Identifier:     "ServiceID-1234",                         // Identifier
		GivenName:      "",                                       // GivenName
		FamilyName:     "",                                       // FamilyName
		Name:           "gopep",                                  // Name
		Email:          "",                                       // Email
		Account:        account,                                  // Account
		GrantType:      "urn:ibm:params:oauth:grant-type:apikey", // GrantType
		Scope:          "ibm openid",                             // Scope
		ClientID:       "default",                                // ClientID
		ACR:            1,                                        // ACR
		AMR: []string{
			"pwd", // AMR
		},
		Sub:                "ServiceID-1234", // Sub
		SubType:            "ServiceId",      // SubType
		UniqueInstanceCrns: []string{},       // UniqueInstanceCrns
	}

	// Create token string
	tokenString, err := t.SignedString(key)

	if err != nil {
		return nil, err
	}

	token := token.IAMToken{
		AccessToken:  tokenString,
		RefreshToken: "string",
		TokenType:    "Bearer",
		ExpiresIn:    expirySeconds,
		Expiration:   int(expiresAt),
		Scope:        "ibm openid",
	}

	return json.Marshal(token)
}

func createKey() ([]byte, error) {

	verifyKey := []iKey{
		{
			Kty: "RSA",
			N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
			E:   "AQAB", // No encoding for default
			Alg: "RS256",
			Kid: "20190122",
		},
	}

	verifyKeys := iKeys{Keys: verifyKey}

	keyBytes, err := json.Marshal(verifyKeys)

	if err != nil {
		return nil, err
	}

	return keyBytes, nil

}

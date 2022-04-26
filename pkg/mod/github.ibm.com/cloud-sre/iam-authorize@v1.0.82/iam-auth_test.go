package iamauth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	instana "github.com/instana/go-sensor"
	"github.ibm.com/cloud-sre/iam-authorize/iam"
)

var (
	iamauth              *IAMAuth
	mockPublicIAMServer  *httptest.Server
	req                  *http.Request
	publicKey            string
	rsaCert              *rsa.PrivateKey
	expiredAccessToken   = "eyJraWQiOiIyMDUwMDkyMjE4MzMiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0xMjM0NURPQSIsImlkIjoiSUJNaWQtMTIzNDVET0EiLCJyZWFsbWlkIjoiSUJNaWQiLCJpZGVudGlmaWVyIjoiMTIzNDVET0EiLCJnaXZlbl9uYW1lIjoiRmFrZSIsImZhbWlseV9uYW1lIjoiTmFtZSIsIm5hbWUiOiJGYWtlIE5hbWUiLCJlbWFpbCI6ImZha2VtYWlsQGlibS5jb20iLCJzdWIiOiJmYWtlbWFpbEBpYm0uY29tIiwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiMTIzNDUifSwiaWF0IjoxNTUyNjY3MzY4LCJleHAiOjE1NDcwNzg0MDAsImlzcyI6Imh0dHBzOi8vaWFtLnN0YWdlMS5ibHVlbWl4Lm5ldC9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImJ4IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.c"
	validAccessToken     = createJWTAndSign(2556057600)
	cacheTestAccessToken string
	fakeEmail            = "fakemail@ibm.com"
)

func TestMain(m *testing.M) {

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	var err error

	req, err = http.NewRequest("POST", "https://localhost/api/catalog/impls", nil)
	if err != nil {
		log.Fatal(err)
	}

	mockPublicIAMServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data string

		if r.URL.Path == "/identity/token" {
			if req.Header.Get("TestHeader") == "true" {
				data = `{"access_token":"` + expiredAccessToken + `","refresh_token":"invalid.refresh.token","token_type":"Bearer","expires_in":3600,"expiration":1553784508,"scope":"ibm openid"}`
			} else if req.Header.Get("TestInvalidAccessToken") == "true" {
				data = `invalid.access.code`
			} else {
				data = `{"access_token":"` + validAccessToken + `","refresh_token":"invalid.refresh.token","token_type":"Bearer","expires_in":3600,"expiration":1553784508,"scope":"ibm openid"}`
			}

			if req.Header.Get("TestHeader") == "400" {
				w.WriteHeader(400)
			}

			// add all the private key creation and sign to this file, or toa pkg
			if req.Header.Get("TestHeader") == "token" {
				if cacheTestAccessToken == "" {
					cacheTestAccessToken = createJWTAndSign(time.Now().Unix() + 5)
				}

				data = `{"access_token":"` + cacheTestAccessToken + `","refresh_token":"invalid.refresh.token","token_type":"Bearer","expires_in":3600,"expiration":1553784508,"scope":"ibm openid"}`
			}
		}

		if r.URL.Path == "/v2/authz" || r.URL.Path == "/v2/authz/bulk" {
			if req.Header.Get("TestHeader") == "unauthorized" {
				data = `{"Trace":"btqb7ojpc98ll6esnsr0", "Decisions":[{"Permitted":false, "Cached":false, "Expired":false, "RetryCount":0}], "ErrorForExpiredResults":""}`
			} else {
				data = `{"decisions":[{"decision":"Permit","obligation":{"actions":["pnp-api-oss.rest.patch","pnp-api-oss.rest.put","pnp-api-oss.rest.post","pnp-api-oss.rest.delete"],"maxCacheAgeSeconds":600,"subject":{"attributes":{"id":"IBMid-12345DOA"}}}}],"cacheKeyPattern":{"order":["subject","resource","action"],"subject":[["id"],["id","scope"]],"resource":[[],["serviceName"],["serviceName","accountId"],["serviceName","accountId","serviceInstance"],["serviceName","accountId","serviceInstance","resourceType"],["serviceName","accountId","serviceInstance","resourceType","resource"]]}}`
			}
		}

		if r.URL.Path == "/identity/keys" {
			data = `{"keys":[{"kty":"RSA","n":"` + publicKey + `","e":"AQAB","alg":"RS256","kid":"205009221833"}]}`
		}

		_, err := w.Write([]byte(data))
		if err != nil {
			log.Fatal(err)
		}
	}))

	os.Setenv("IAM_URL", mockPublicIAMServer.URL)

	m.Run()
}

func TestIAMAuthorizationNoAPIKeyToken(t *testing.T) {

	os.Setenv("CLOUD_SERVICE_API_KEY", "key")

	iamauth = NewIAMAuth("")

	iamauth.SetIAMURL(mockPublicIAMServer.URL)
	iamauth.SetCRNServiceName("my-svc-123")
	iamauth.SetResourceTypeName("some-resource-type")

	monConfig := &iam.NRWrapperConfig{
		NRApp:             nil,
		InstanaSensor:     instana.NewSensor("my-sensor"),
		Environment:       "staging",
		Region:            "us-east",
		URL:               "https://iam.test.cloud.ibm.com",
		SourceServiceName: "my-svc",
	}
	iamauth.SetNRWrapperConfig(monConfig)

	// no api key/token provided
	_, err := iamauth.IsIAMAuthorized(req, nil)
	if err == nil {
		t.Fatal(err)
	}

	expected := "no access token or api key provided"
	assess(err.Error(), expected, t)
}

func TestIAMAuthorizationNoServiceAPIKey(t *testing.T) {

	req.Header.Set("Authorization", "my-api-key")
	iamauth.serviceAPIKey = ""
	os.Setenv("CLOUD_SERVICE_API_KEY", "")

	// no service api key provided
	_, err := iamauth.IsIAMAuthorized(req, nil)
	if err == nil {
		t.Fatal(err)
	}
	expected := "No Cloud Service API Key has been set. Please set it with the SetServiceAPIKey() function or by setting env variable CLOUD_SERVICE_API_KEY"
	assess(err.Error(), expected, t)

}

func TestIAMAuthorizationSuccess(t *testing.T) {

	iamauth.SetServiceAPIKey("service-api-key")

	// successful call
	email, err := iamauth.IsIAMAuthorized(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	assess(email.Email, fakeEmail, t)
}

func TestIAMAuthorizationSuccessBasicAuth(t *testing.T) {

	iamauth.SetServiceAPIKey("service-api-key")
	req.SetBasicAuth("apikey", "my.api.key")

	// successful call
	email, err := iamauth.IsIAMAuthorized(req, nil)
	if err != nil {
		t.Fatal(err)
	}
	assess(email.Email, fakeEmail, t)
}

func TestIAMAuthorizationUnauthorized(t *testing.T) {
	// unauthorized
	req.Header.Set("TestHeader", "unauthorized")

	cp := req.URL.Path
	req.URL.Path = "/subscription/api/concern/subscriptions"

	email, err := iamauth.IsIAMAuthorized(req, nil)
	if err == nil || email != nil {
		t.Fatal(err, email)
	}
	assess(err.Error(), "unauthorized", t)
	req.URL.Path = cp

	req.Header.Set("TestHeader", "")
}

func TestIAMAuthorizationInvalidToken(t *testing.T) {
	req.Header.Set("Authorization", "my.access.token")
	// this returns an invalid token from mock iam, which will make the code fail
	req.Header.Set("TestInvalidAccessToken", "true")

	_, err := iamauth.IsIAMAuthorized(req, nil)
	if err == nil {
		t.Fatal(err)
	}
	expected := `invalid character 'i' looking for beginning of value`
	assess(err.Error(), expected, t)
	req.Header.Set("TestInvalidAccessToken", "")
}

func TestIAMAuthorizationInvalidTokenAPIKey(t *testing.T) {

	req.Header.Set("Authorization", "Bearer my-api-key")
	_, err := iamauth.IsIAMAuthorized(req, nil)
	if err == nil {
		t.Fatal(err)
	}
	expected := `unauthorized`
	assess(err.Error(), expected, t)

}

func TestGetAccessTokenExpiredToken(t *testing.T) {
	req.Header.Set("TestHeader", "true")

	// return expired access token
	token, err := iamauth.GetAccessToken("my-api-key200")
	if err != nil {
		t.Fatal(err)
	}
	assess(token, expiredAccessToken, t)
	req.Header.Set("TestHeader", "")
}

func TestGetAccessTokenEmptyAccessToken(t *testing.T) {

	// empty access token
	_, err := iamauth.GetAccessToken("")
	if err == nil {
		t.Fatal(err)
	}
	expected := "empty api key provided in function call"
	assess(err.Error(), expected, t)
}

func TestGetAccessTokenError400(t *testing.T) {
	req.Header.Set("TestHeader", "400")
	// server response code is 400
	_, err := iamauth.GetAccessToken("apik")
	if err == nil {
		t.Fatal(err)
	}
	expected := "Request to URL " + mockPublicIAMServer.URL + "/identity/token failed with status code 400"
	if !strings.HasPrefix(err.Error(), expected) {
		t.Fatalf("returned result does not start with expected.\ngot=\n\t%s\nexpected=\n\t%s\n", err.Error(), expected)
	}
	req.Header.Set("TestHeader", "")
}

func TestGetAccountIDInvalidBase64Token(t *testing.T) {
	// no base64 token
	_, err := iamauth.GetAccountID("invalid.token.")
	if err == nil {
		t.Fatal(err)
	}
	expected := "illegal base64 data at input byte 4"
	assess(err.Error(), expected, t)
}

func TestGetAccountIDInvalidToken(t *testing.T) {
	_, err := iamauth.GetAccountID("invalid.token")
	if err == nil {
		t.Fatal(err)
	}
	expected := "invalid token"
	assess(err.Error(), expected, t)
}

func TestGetAccountIDTokenExpired(t *testing.T) {

	// token is expired
	_, err := iamauth.GetAccountID(expiredAccessToken)
	if err == nil {
		t.Fatal(err)
	}
	expected := "token is expired"
	assess(err.Error(), expected, t)

}

func TestGetEmail(t *testing.T) {

	req.Header.Set("Authorization", expiredAccessToken)

	// token is expired
	email, err := iamauth.GetEmail(req)
	if err != nil {
		t.Fatal(err)
	}

	assess(email, "fakemail@ibm.com", t)

}

func TestCache(t *testing.T) {

	defer mockPublicIAMServer.Close()

	iamauth.SetCacheCleanupInterval(10 * time.Second)
	iamauth.SetCacheMaxSize(100)
	iamauth.SetCacheSizeThreshold(10)

	iamauth.cleanUpAPIKeyCache()

	iamauth.cache.Lock()
	previousInCache := len(iamauth.cache.apikCache)
	iamauth.cache.Unlock()

	// add 101 items in a cache of size 100. it should only add 100.
	// should have in the log "unable to add api key to cache, cache is at the maximum size..."
	for i := 0; i < 101; i++ {

		req.Header.Set("Authorization", "api-key-"+strconv.Itoa(i))
		if i >= 10 {
			req.Header.Set("TestHeader", "token")
		}

		// email, err := iamauth.IsIAMAuthorized(req,nil, "resource-123", "permission-123")
		email, err := iamauth.IsIAMAuthorized(req, nil)
		if err != nil {
			t.Fatal(err)
		}
		assess(email.Email, fakeEmail, t)

	}

	iamauth.cache.Lock()
	assess(fmt.Sprint(len(iamauth.cache.apikCache)), fmt.Sprint(iamauth.cache.maxSize), t)
	iamauth.cache.Unlock()

	// sleep so that the cache cleanup takes place(happens every 10s)
	time.Sleep(22 * time.Second)

	iamauth.cache.Lock()
	defer iamauth.cache.Unlock()
	count := len(iamauth.cache.apikCache)
	assess(fmt.Sprint(count), fmt.Sprint(10+previousInCache), t) // add 10 api keys(from the for loop above) plus what already was in cache before the for loop
}

func TestGetToken(t *testing.T) {
	assess(getToken("Bearer tokenValue1"), "tokenValue1", t)
	assess(getToken("bearer tokenValue1"), "tokenValue1", t)
	assess(getToken("  bearer tokenValue1  "), "tokenValue1", t)
	assess(getToken("tokenValue1"), "", t)
}

func assess(got, expected string, t *testing.T) {
	if got != expected {
		t.Fatalf("returned result does not match the expected.\ngot=\n\t%s\nexpected=\n\t%s\n", got, expected)
	}
}

/*

	Code below is for creating tokens and singing each of them with the generated certificate's private key

*/

type Claims struct {
	ID                 string      `json:"id,omitempty"`
	IAMid              string      `json:"iam_id,omitempty"`
	RealmID            string      `json:"realmid,omitempty"`
	Identifier         string      `json:"identifier,omitempty"`
	Sub                string      `json:"sub,omitempty"`
	SubType            string      `json:"sub_type,omitempty"`
	Account            account     `json:"account,omitempty"`
	Iat                json.Number `json:"iat,Number,omitempty"`
	Exp                json.Number `json:"exp,Number,omitempty"`
	Email              string      `json:"email,omitempty"`
	Iss                string      `json:"iss,omitempty"`
	GrantType          string      `json:"grant_type,omitempty"`
	Scope              string      `json:"scope,omitempty"`
	ClientID           string      `json:"client_id,omitempty"`
	UniqueInstanceCrns []string    `json:"unique_instance_crns,omitempty"`
	jwt.StandardClaims
}

type account struct {
	Ims string `json:"ims,omitempty"`
	V   bool   `json:"valid,omitempty"`
	Bss string `json:"bss,omitempty"`
}

// generateKey generates a digital ceritifcate with private and public keys
func generateKey() {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println(err)
	}

	rsaCert = privKey

	publicKey = base64.RawURLEncoding.EncodeToString(privKey.PublicKey.N.Bytes())
}

// createJWTAndSign creates a token(1st adn 2nd parts) with the signed(3rd part) of the token
func createJWTAndSign(expAt int64) string {

	if rsaCert == nil {
		generateKey()
	}

	accs := account{V: true}

	// Create the JWT claims
	claims := &Claims{
		ID:         "IBMid-12345DOA",
		IAMid:      "IBMid-12345DOA",
		RealmID:    "IBMid",
		Identifier: "12345DOA",
		Email:      "fakemail@ibm.com",
		Sub:        "fakemail@ibm.com",
		Account:    accs,
		Exp:        json.Number(strconv.FormatInt(expAt, 10)),
		Iat:        "1552667368",
		GrantType:  "urn:ibm:params:oauth:grant-type:apikey",
		Scope:      "ibm openid",
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expAt,
			Issuer:    "test",
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	delete(token.Header, "typ")
	token.Header["kid"] = "205009221833"

	// Create the JWT string
	tokenString, err := token.SignedString(rsaCert)
	if err != nil {
		log.Println(err)
	}

	return tokenString
}

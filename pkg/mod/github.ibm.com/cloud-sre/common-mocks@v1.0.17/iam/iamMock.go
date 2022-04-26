package iammock

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

const MockAccessToken = `eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJJQk1pZC0xMjM0NURPQSIsImlkIjoiSUJNaWQtMTIzNDVET0EiLCJyZWFsbWlkIjoiSUJNaWQiLCJpZGVudGlmaWVyIjoiMTIzNDVET0EiLCJnaXZlbl9uYW1lIjoiRmFrZSIsImZhbWlseV9uYW1lIjoiTmFtZSIsIm5hbWUiOiJGYWtlIE5hbWUiLCJlbWFpbCI6ImZha2VtYWlsQGlibS5jb20iLCJzdWIiOiJmYWtlbWFpbEBpYm0uY29tIiwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiMTIzNDUifSwiaWF0IjoxNTUyNjY3MzY4LCJleHAiOjI1NTYwNTc2MDAsImlzcyI6Imh0dHBzOi8vaWFtLnN0YWdlMS5ibHVlbWl4Lm5ldC9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIG9wZW5pZCBhbGwiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.c` // #nosec G101

var (
	MockIAM   *httptest.Server
	PublicKey string
)

// generatePublicKey generates a digital ceritifcate with private and public keys
func generatePublicKey() {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println(err)
	}

	PublicKey = base64.RawURLEncoding.EncodeToString(privKey.PublicKey.N.Bytes())
}

var GenericHandler = func(w http.ResponseWriter, r *http.Request) {
	var data string

	if r.URL.Path == `/identity/token` {
		data = `{"access_token":"` + MockAccessToken + `","refresh_token":"invalid.refresh.token","token_type":"Bearer","expires_in":3600,"expiration":1553784508,"scope":"ibm openid all"}`
	}

	if strings.HasPrefix(r.URL.Path, `/v2/authz`) {
		data = `{"decisions":[{"decision":"Permit","obligation":{"actions":["pnp-api-oss.rest.patch","pnp-api-oss.rest.put","pnp-api-oss.rest.post","pnp-api-oss.rest.delete"],"maxCacheAgeSeconds":600,"subject":{"attributes":{"id":"IBMid-12345DOA"}}}}],"cacheKeyPattern":{"order":["subject","resource","action"],"subject":[["id"],["id","scope"]],"resource":[[],["serviceName"],["serviceName","accountId"],["serviceName","accountId","serviceInstance"],["serviceName","accountId","serviceInstance","resourceType"],["serviceName","accountId","serviceInstance","resourceType","resource"]]}}`
	}

	if r.URL.Path == `/identity/keys` {
		data = `{"keys":[{"kty":"RSA","n":"` + PublicKey + `","e":"AQAB","alg":"RS256","kid":"205009221833"}]}`
	}

	w.Write([]byte(data))
}

// SetupIAMMockGeneric will setup a mock IAM server to be used for testing
func SetupIAMMockGeneric() {
	os.Setenv(`IAM_MODE`, `test`)
	os.Setenv(`IAM_BYPASS_FLAG`, `true`)

	SetupIAMMockCustomHandler(GenericHandler)

	os.Setenv(`IAM_URL`, MockIAM.URL)
}

// SetupIAMMockCustomHandler will setup a mock IAM server
// with a customed handler function to be used for testing
func SetupIAMMockCustomHandler(hf http.HandlerFunc) {
	generatePublicKey()
	MockIAM = httptest.NewServer(hf)
}

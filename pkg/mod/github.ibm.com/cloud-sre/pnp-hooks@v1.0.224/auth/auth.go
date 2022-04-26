package auth

import (
	"github.ibm.com/cloud-sre/pnp-abstraction/monitoring"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
	"log"
	"net/http"
)

// Auth defines the functions supported
type Auth interface {
	Middleware(handler http.HandlerFunc) http.HandlerFunc
	IsRequestAuthorized(resp http.ResponseWriter, req *http.Request) (isAuth bool)
}

// NewAuth creates a new authorization instance that can be used to authorize a request
func NewAuth(iamAuth IAMAuth) Auth {
	return &defaultAuth{IAMAuth: iamAuth}
}

// defaultAuth is the default implementation of the Auth interface
type defaultAuth struct {
	IAMAuth IAMAuth
}

// Middleware - endpoint handler that authorizes requests
func (auth *defaultAuth) Middleware(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.IsRequestAuthorized(w, r) {
			handler(w, r)
		} else {
			return
		}
	})
}

// IsRequestAuthorized takes the request and will return true if the request is authorized, false otherwise.
// If false, the response will be filled in with an appropriate error
func (auth *defaultAuth) IsRequestAuthorized(resp http.ResponseWriter, req *http.Request) (isAuth bool) {
	isAuth = false

	// Look for the hardcoded accepted token first:
	token := req.Header.Get("Authorization")
	isAuth = shared.HasValidToken(token) && token != ""

	// If the hardcoded accepted token was not found, ask IAM to authorize the request:
	if !isAuth {
		authresp, err := auth.IAMAuth.IsIAMAuthorized(req, resp)
		if err != nil {
			log.Print("IsRequestAuthorized: IAM Authorization error: ", err, "\n\t", authresp)
		} else if authresp != nil {
			monitoring.AddCustomAttribute(resp, "iamAuthorization", authresp.Email)
			isAuth = true
		}
	}

	if !isAuth {
		resp.WriteHeader(http.StatusUnauthorized)
	}

	return isAuth
}


package auth

import (
	"errors"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	iam "github.ibm.com/cloud-sre/iam-authorize"
)

// Mock IAM auth used for testing:
type mockIAMAuth struct {
	IsIAMAuthorizedImpl func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error)
}

func (iamAuth *mockIAMAuth) IsIAMAuthorized(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
	return iamAuth.IsIAMAuthorizedImpl(req, resp)
}

// TestMiddlewareAuthorized tests a call to the Middleware function where the request is authorized
func TestMiddlewareAuthorized(t *testing.T) {
	// Create mock IAM auth instance:
	iamAuth := &mockIAMAuth{
		IsIAMAuthorizedImpl: func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
			// Successful authorization result:
			return &iam.Authorization{Email: "fred@ibm.com", Source: "test"}, nil
		},
	}

	// Create auth instance using mock IAM auth instance:
	auth := NewAuth(iamAuth)

	// Create mock request and get the associated response writer:
	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)
	responseWriter := httptest.NewRecorder()

	// Call the Middleware function to get the handler, and then call the handler and verify the response:
	handler := auth.Middleware(func(resp http.ResponseWriter, req *http.Request) {})
	handler(responseWriter, request)
	response := responseWriter.Result()
	AssertEqual(t, "Status code", 200, response.StatusCode)
}

// TestMiddlewareUnauthorized tests a call to the Middleware function where the request is unauthorized
func TestMiddlewareUnauthorized(t *testing.T) {
	// Create mock IAM auth instance:
	iamAuth := &mockIAMAuth{
		IsIAMAuthorizedImpl: func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
			// Unsuccessful authorization result:
			return nil, errors.New("You are not allowed!")
		},
	}

	// Create auth instance using mock IAM auth instance:
	auth := NewAuth(iamAuth)

	// Create mock request and get the associated response writer:
	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)
	responseWriter := httptest.NewRecorder()

	// Call the Middleware function to get the handler, and then call the handler and verify the response:
	handler := auth.Middleware(func(resp http.ResponseWriter, req *http.Request) {})
	handler(responseWriter, request)
	response := responseWriter.Result()
	AssertEqual(t, "Status code", 401, response.StatusCode)
}

// TestIsRequestAuthorizedHardcodedToken tests the IsRequestAuthorized function when the request includes the known hardcoded token
func TestIsRequestAuthorizedHardcodedToken(t *testing.T) {
	// Create mock IAM auth instance:
	iamAuth := &mockIAMAuth{
		IsIAMAuthorizedImpl: func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
			// Unsuccessful authorization result - never called in this unit test:
			return nil, errors.New("You are not allowed!")
		},
	}

	// Create auth instance using mock IAM auth instance:
	auth := NewAuth(iamAuth)

	// Create mock request and get the associated response writer:
	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)
	request.Header.Add("Authorization", "0123456789")
	shared.SnowToken = "0123456789"
	responseWriter := httptest.NewRecorder()

	// Check for authorization:
	isAuthorized := auth.IsRequestAuthorized(responseWriter, request)
	AssertEqual(t, "isAuthorized", true, isAuthorized)
	response := responseWriter.Result()
	AssertEqual(t, "Status code", 200, response.StatusCode)
}

// TestIsRequestAuthorizedAuthorized tests the IsRequestAuthorized function when the request includes a valid authorization
func TestIsRequestAuthorizedAuthorized(t *testing.T) {
	// Create mock IAM auth instance:
	iamAuth := &mockIAMAuth{
		IsIAMAuthorizedImpl: func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
			// Successful authorization result:
			return &iam.Authorization{Email: "fred@ibm.com", Source: "test"}, nil
		},
	}

	// Create auth instance using mock IAM auth instance:
	auth := NewAuth(iamAuth)

	// Create mock request and get the associated response writer:
	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)
	request.Header.Add("Authorization", "def456")
	responseWriter := httptest.NewRecorder()

	// Check for authorization:
	isAuthorized := auth.IsRequestAuthorized(responseWriter, request)
	AssertEqual(t, "isAuthorized", true, isAuthorized)
	response := responseWriter.Result()
	AssertEqual(t, "Status code", 200, response.StatusCode)
}

// TestIsRequestAuthorizedUnauthorized tests the IsRequestAuthorized function when the request includes an invalid authorization
func TestIsRequestAuthorizedUnauthorized(t *testing.T) {
	// Create mock IAM auth instance:
	iamAuth := &mockIAMAuth{
		IsIAMAuthorizedImpl: func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
			// Unsuccessful authorization result:
			return nil, errors.New("You are not allowed!")
		},
	}

	// Create auth instance using mock IAM auth instance:
	auth := NewAuth(iamAuth)

	// Create mock request and get the associated response writer:
	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)
	request.Header.Add("Authorization", "def456")
	responseWriter := httptest.NewRecorder()

	// Check for authorization:
	isAuthorized := auth.IsRequestAuthorized(responseWriter, request)
	AssertEqual(t, "isAuthorized", false, isAuthorized)
	response := responseWriter.Result()
	AssertEqual(t, "Status code", 401, response.StatusCode)
}

// TestIsRequestAuthorizedEmptyToken tests the IsRequestAuthorized function when the request includes an empty token
func TestIsRequestAuthorizedEmptyToken(t *testing.T) {
	// Create mock IAM auth instance:
	iamAuth := &mockIAMAuth{
		IsIAMAuthorizedImpl: func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
			// Unsuccessful authorization result:
			return nil, errors.New("You are not allowed!")
		},
	}

	// Create auth instance using mock IAM auth instance:
	auth := NewAuth(iamAuth)

	// Create mock request and get the associated response writer:
	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)
	request.Header.Add("Authorization", "Bearer ")
	responseWriter := httptest.NewRecorder()

	// Check for authorization:
	isAuthorized := auth.IsRequestAuthorized(responseWriter, request)
	AssertEqual(t, "isAuthorized", false, isAuthorized)
	response := responseWriter.Result()
	AssertEqual(t, "Status code", 401, response.StatusCode)
}

// TestIsRequestAuthorizedEmptyAPIKey tests the IsRequestAuthorized function when the request includes an empty API key
func TestIsRequestAuthorizedEmptyAPIKey(t *testing.T) {
	// Create mock IAM auth instance:
	iamAuth := &mockIAMAuth{
		IsIAMAuthorizedImpl: func(req *http.Request, resp http.ResponseWriter) (*iam.Authorization, error) {
			// Unsuccessful authorization result:
			return nil, errors.New("You are not allowed!")
		},
	}

	// Create auth instance using mock IAM auth instance:
	auth := NewAuth(iamAuth)

	// Create mock request and get the associated response writer:
	request := httptest.NewRequest(http.MethodPost, "https://localhost:8080", nil)
	request.Header.Add("Authorization", "")
	responseWriter := httptest.NewRecorder()

	// Check for authorization:
	isAuthorized := auth.IsRequestAuthorized(responseWriter, request)
	AssertEqual(t, "isAuthorized", false, isAuthorized)
	response := responseWriter.Result()
	AssertEqual(t, "Status code", 401, response.StatusCode)
}

// AssertEqual reports a testing failure when a given actual item (interface{}) is not equal to the expected value
func AssertEqual(t *testing.T, item string, expected, actual interface{}) {
	if actual == nil {
		if expected != nil {
			t.Logf("%s %s mismatch - expected %T %v got %T %v", t.Name(), item, expected, expected, actual, actual)
			t.Fail()
		}
	} else if reflect.TypeOf(actual).Comparable() && reflect.TypeOf(actual).Kind() != reflect.Ptr {
		// Note that pointers are comparable but this doesn't apply to comparing de-referenced values
		if expected != actual {
			t.Logf("%s %s mismatch - expected %T %v got %T %v", t.Name(), item, expected, expected, actual, actual)
			t.Fail()
		}
	} else {
		if !reflect.DeepEqual(expected, actual) {
			t.Logf("%s %s mismatch - expected %T %v got %T %v", t.Name(), item, expected, expected, actual, actual)
			t.Fail()
		}
	}
}

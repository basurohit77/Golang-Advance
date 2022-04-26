// Package pep is a Policy Enforcement Point that performs authorization check with PDP.
package pep

import (
	"errors"
	"strings"

	"github.ibm.com/IAM/token/v5"
)

// Deny Reason definitions
type DenyReason int

const (
	DenyReasonNone DenyReason = iota
	DenyReasonIAM
	DenyReasonContext
)

// Decision contains information relevant to the decision obtained from the pdp.
type Decision struct {
	Permitted  bool       // the decision
	Cached     bool       // is the decision from the cache
	Expired    bool       // is the decision already expired
	RetryCount int        // How many times the retry to pdp
	Reason     DenyReason // Reason for non-permit decision
}

// AuthzResponse is an authorization response of an authorization request
// to the PDP.
type AuthzResponse struct {
	Trace                  string
	Decisions              []Decision
	ErrorForExpiredResults string
}

// RolesResponse - The response to a PDP /v2/authz/roles request
type RolesResponse struct {
	trace         string
	Attributes    RoleAttributes
	ServiceRoles  []RoleActions
	PlatformRoles []RoleActions
}

// AuthzRequest is the interface that wraps a the authorization request to the PDP.
type AuthzRequest interface {
	isAuthorized(trace string, serviceOperatorToken string) (AuthzResponse, error)
}

// RolesRequest - The PDP roles request
type RolesRequest interface {
	getRoles(trace string, serviceOperatorToken string) ([]RolesResponse, error)
}

// PerformAuthorization makes an authorization request to PDP.
// Any returned error is of type *pep.APIError or *pep.InternalError
func PerformAuthorization(r AuthzRequest, trace string) (AuthzResponse, error) {
	conf := GetConfig().(*Config)

	if !conf.IsInitialized {
		return AuthzResponse{}, errors.New("PEP is not initialized with an API Key")
	}

	conf.Statistics.Inc(requestsServicedByPEP)

	token, err := GetToken()

	if err != nil {
		conf.Statistics.Inc(failedUserRequests)
		return AuthzResponse{}, err
	}

	resp, err := r.isAuthorized(trace, token)

	if err != nil {
		conf.Statistics.Inc(failedUserRequests)
	}

	return resp, err
}

// PerformAuthorizationWithToken makes an authorization request to PDP with given serviceOperatorToken.
// The "serviceOperatorToken" parameter in this call is the JWT used to authenticate and authorize with the IAM PDP (generally representing the service operator).
// It is NOT the "subject" of the actual authorization request.
// Any returned error is of type *pep.APIError or *pep.InternalError
func PerformAuthorizationWithToken(r AuthzRequest, trace string, serviceOperatorToken string) (AuthzResponse, error) {
	if serviceOperatorToken == "" {
		return AuthzResponse{}, errors.New("authorization token is required")
	}

	conf := GetConfig().(*Config)

	conf.Statistics.Inc(requestsServicedByPEP)

	resp, err := r.isAuthorized(trace, serviceOperatorToken)

	if err != nil {
		conf.Statistics.Inc(failedUserRequests)
	}

	return resp, err
}

// GetAuthorizedRoles performs the role check against the PDP using the specified request.
// Any returned error is of type *pep.APIError or *pep.InternalError
func GetAuthorizedRoles(r RolesRequest, trace string) ([]RolesResponse, error) {
	conf := GetConfig().(*Config)

	if !conf.IsInitialized {
		return []RolesResponse{}, errors.New("PEP is not initialized with an API Key")
	}

	conf.Statistics.Inc(requestsServicedByPEP)

	token, err := GetToken()

	if err != nil {
		conf.Statistics.Inc(failedUserRequests)
		return []RolesResponse{}, err
	}

	resp, err := r.getRoles(trace, token)

	if err != nil {
		conf.Statistics.Inc(failedUserRequests)
	}

	return resp, err
}

// GetAuthorizedRolesWithToken performs the role check against the PDP using the specified request and given serviceOperatorToken.
// The "serviceOperatorToken" parameter in this call is the JWT used to authenticate and authorize with the IAM PDP (generally representing the service operator).
// It is NOT the "subject" of the actual authorization request.
// Any returned error is of type *pep.APIError or *pep.InternalError
func GetAuthorizedRolesWithToken(r RolesRequest, trace string, serviceOperatorToken string) ([]RolesResponse, error) {
	if serviceOperatorToken == "" {
		return []RolesResponse{}, errors.New("authorization token is required")
	}

	conf := GetConfig().(*Config)

	conf.Statistics.Inc(requestsServicedByPEP)

	resp, err := r.getRoles(trace, serviceOperatorToken)

	if err != nil {
		conf.Statistics.Inc(failedUserRequests)
	}

	return resp, err
}

// GetStatistics returns the available runtime statistics of the PEP
func GetStatistics() Stats {
	conf := GetConfig().(*Config)

	return conf.reportStatisticsStats()
}

// GetToken returns the cached access token
func GetToken() (string, error) {
	conf := GetConfig().(*Config)

	if !conf.IsInitialized {
		return "", errors.New("PEP is not initialized with an API Key")
	}

	return conf.tokenManager.GetToken()
}

// GetClaims consumes a JWT and returns a list of token claims. Can skip JWT validation.
func GetClaims(token string, skipValidation bool) (*token.IAMAccessTokenClaims, error) {
	conf := GetConfig().(*Config)

	if !conf.IsInitialized {
		return nil, errors.New("PEP is not initialized with an API Key")
	}

	return conf.tokenManager.GetClaims(token, skipValidation)
}

// GetSubjectAsIAMIDClaim Consumes a JWT and returns the iam_id claim. Can skip JTW validation.
func GetSubjectAsIAMIDClaim(token string, skipValidation bool) (string, error) {
	conf := GetConfig().(*Config)

	if !conf.IsInitialized {
		return "", errors.New("PEP is not initialized with an API Key")
	}

	return conf.tokenManager.GetSubjectAsIAMIDClaim(token, skipValidation)
}

// GetSubjectFromToken returns the an `Attributes` object from a user, service, or crn token that can be
// used in authorization requests
func GetSubjectFromToken(token string, skipValidation bool) (subject Attributes, err error) {

	if !skipValidation {
		conf := GetConfig().(*Config)

		if !conf.IsInitialized {
			return nil, errors.New("PEP is not initialized with an API Key")
		}

		_, err := conf.tokenManager.GetClaims(token, skipValidation)

		if err != nil {
			return nil, err
		}
	}

	// parse the body from the full token was provided
	tokenBody, err := getTokenBody(token)

	if err != nil {
		return nil, err
	}

	subject = Attributes{
		"accessTokenBody": tokenBody,
	}

	return subject, nil
}

// ParseCRNToAttributes returns an attributes object containing the fields extracted from a CRN string.
func ParseCRNToAttributes(crnString string) (crnAttributes Attributes) {
	var accountID, organizationID, spaceID, projectID string

	crnFields := strings.Split(crnString, ":")

	if len(crnFields) != 10 {
		return nil
	}

	scopeFields := strings.Split(crnFields[6], "/")

	if len(scopeFields) == 2 {
		if scopeFields[0] == "a" {
			accountID = scopeFields[1]
		} else if scopeFields[0] == "o" {
			organizationID = scopeFields[1]
		} else if scopeFields[0] == "s" {
			spaceID = scopeFields[1]
		} else if scopeFields[0] == "p" {
			projectID = scopeFields[1]
		}
	}

	realm := strings.Split(crnFields[0], "-")

	if len(realm) == 2 {
		crnAttributes = Attributes{
			"realmid":         realm[0],
			"crn":             realm[1],
			"version":         crnFields[1],
			"cname":           crnFields[2],
			"ctype":           crnFields[3],
			"serviceName":     crnFields[4],
			"serviceInstance": crnFields[7],
			"region":          crnFields[5],
			"resourceType":    crnFields[8],
			"resource":        crnFields[9],
			"spaceId":         spaceID,
			"accountId":       accountID,
			"organizationId":  organizationID,
			"projectId":       projectID,
		}
	} else {
		crnAttributes = Attributes{
			"realmid":         "",
			"crn":             crnFields[0],
			"version":         crnFields[1],
			"cname":           crnFields[2],
			"ctype":           crnFields[3],
			"serviceName":     crnFields[4],
			"serviceInstance": crnFields[7],
			"region":          crnFields[5],
			"resourceType":    crnFields[8],
			"resource":        crnFields[9],
			"spaceId":         spaceID,
			"accountId":       accountID,
			"organizationId":  organizationID,
			"projectId":       projectID,
		}
	}

	return crnAttributes
}

// GetDelegationToken returns the delegation token for the provided desiredIAMID. The token the PEP manages (via the configured APIKEY) is used as the credential in this exchange. This means that only delegations related to the permissions of the configured API key will be valid.
func GetDelegationToken(desiredIAMID string) (string, error) {
	conf := GetConfig().(*Config)

	if !conf.IsInitialized {
		return "", errors.New("PEP is not initialized with an API Key")
	}

	return conf.tokenManager.GetDelegationToken(desiredIAMID)

}

// getTokenBody returns the body of a JWT token and an error if it does not match JWT format or if the body is empty
func getTokenBody(token string) (string, error) {
	tokenParts := strings.Split(token, ".")

	if len(tokenParts) != 3 {
		return "", errors.New("failed to parse, provided string does not match 3 section . separated token format")
	}

	if tokenParts[1] == "" {
		return "", errors.New("token body is empty")
	}

	return tokenParts[1], nil
}

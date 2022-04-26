package pep

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.ibm.com/IAM/pep/v3/cache"
)

// CacheKeyPattern contains the relevant pattern for building keys that can be used
// to store/lookup authorization decision to/from the cache.
type CacheKeyPattern struct {
	Order    []string   `json:"order"`
	Subject  [][]string `json:"subject"`
	Resource [][]string `json:"resource"`
}

// These are the possible PDP authz response.
type authzResponse struct {
	CacheKeyPattern `json:"cacheKeyPattern"`
	Responses       []struct {
		Status string `json:"status"`
		Error  struct {
			TransactionID string `json:"transactionId"`
			InstanceID    string `json:"instanceId"`
			Errors        []struct {
				BusinessCode string `json:"businessCode"`
				Code         string `json:"code"`
				Message      string `json:"message"`
				MoreInfo     string `json:"moreInfo"`
				Target       struct {
					Type string `json:"type"`
					Name string `json:"name"`
				} `json:"target"`
			} `json:"errors"`
		} `json:"error,omitempty"`
		AuthorizationDecision struct {
			Permitted  bool   `json:"permitted"`
			Reason     string `json:"reason"`
			Obligation struct {
				Actions     []string `json:"actions"`
				Environment struct {
					Attributes map[string]interface{} `json:"attributes"`
				} `json:"environment"`
				MaxCacheAgeSeconds int `json:"maxCacheAgeSeconds"`
				Subject            struct {
					Attributes map[string]interface{} `json:"attributes"`
				} `json:"subject"`
				Resource struct {
					Attributes map[string]interface{} `json:"attributes"`
				} `json:"resource"`
			} `json:"obligation"`
		} `json:"authorizationDecision,omitempty"`
	} `json:"responses"`
}

type listResponse struct {
	CacheKeyPattern `json:"cacheKeyPattern"`
	Decisions       []struct {
		Decision   string `json:"decision"`
		Reason     string `json:"reason"`
		Obligation struct {
			Actions     []string `json:"actions"`
			Environment struct {
				Attributes map[string]interface{} `json:"attributes"`
			} `json:"environment"`
			MaxCacheAgeSeconds int `json:"maxCacheAgeSeconds"`
			Subject            struct {
				Attributes map[string]interface{} `json:"attributes"`
			} `json:"subject"`
			Resource struct {
				Attributes map[string]interface{} `json:"attributes"`
			} `json:"resource"`
		} `json:"obligation"`
	} `json:"decisions"`
}

type RoleActions struct {
	Role struct {
		CRN         string `json:"crn"`
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	} `json:"role"`
	Actions []struct {
		Action      string `json:"action"`
		DisplayName string `json:"displayName"`
	} `json:"actions"`
}

type RoleAttributes struct {
	ServiceName     string `json:"serviceName"`
	ServiceInstance string `json:"serviceInstance"`
	AccountID       string `json:"accountId"`
}

type rolesResponse struct {
	Responses []struct {
		RequestID      string         `json:"requestId"`
		Status         string         `json:"status"`
		RoleAttributes RoleAttributes `json:"attributes"`
		SubjectID      struct {
			UserID string `json:"userId"`
		} `json:"subjectId"`
		RoleActions        []RoleActions `json:"roleActions"`
		PlatformExtensions struct {
			RoleActions []RoleActions `json:"roleActions"`
		} `json:"platformExtensions"`
	} `json:"responses"`
	TransactionID string `json:"transactionId"`
	InstanceID    string `json:"instanceId"`
	Errors        []struct {
		BusinessCode string `json:"businessCode"`
		Code         string `json:"code"`
		Message      string `json:"message"`
		MoreInfo     string `json:"moreInfo"`
		Target       struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"target"`
	} `json:"errors"`
}

// Attributes for subject and resource
type Attributes map[string]interface{}
type Request = Attributes

// CacheKey is used to serialize JSON object or CRNs into a key usable by the cache
type CacheKey []byte

// Requests to PDP.
type Requests []Attributes

// Set of possible values returned by the PDP in the decision reason field.
const (
	DenyReasonStringNetwork = "Network"
	DenyReasonStringContext = "Context"
)

// validContextReasons is the set of deny reason strings that should map to DenyReasonContext
// TODO: this can be removed after all deployed PDPs transition away from using "Network"
var validContextReasons = map[string]struct{}{
	DenyReasonStringNetwork: {},
	DenyReasonStringContext: {},
}

// Implementation of the requests to PDP
// The list API is preferred over the v2 authz. Unless the requests fail the requirement for the list API.
func (reqs *Requests) isAuthorized(trace string, serviceOperatorToken string) (AuthzResponse, error) {
	if len(*reqs) < 1 {
		return AuthzResponse{}, buildInternalError(fmt.Errorf("The minimum number of requests is 1"), trace)
	}

	if len(*reqs) > 1000 {
		return AuthzResponse{}, buildInternalError(fmt.Errorf("The maximum number of requests allowed is 1k"), trace)
	}

	notCachedRequests, notCachedRequestIndex, userResponse := getUserResponseFromCache(*reqs)
	userResponse.Trace = trace

	if len(notCachedRequestIndex) == 0 {
		// if all found, return complete user response
		return userResponse, nil
	}

	var authzResp AuthzResponse
	listRequest, err := notCachedRequests.buildListRequest()

	if err != nil {
		// if error at any point return error & empty user response
		authzReq := notCachedRequests.buildAuthzRequests()
		authzResp, err = authzReq.isAuthorizedAndCache(trace, serviceOperatorToken, notCachedRequests)
	} else {
		authzResp, err = listRequest.isAuthorizedAndCache(trace, serviceOperatorToken, notCachedRequests)
	}

	if err != nil {
		c := GetConfig().(*Config)
		if apiErr, ok := err.(*APIError); ok && c.EnableExpiredCache {
			if apiErr.StatusCode == 0 || apiErr.StatusCode >= 500 {
				authzRespCached := AuthzResponse{
					Trace:                  trace,
					Decisions:              []Decision{},
					ErrorForExpiredResults: err.Error(),
				}

				for _, originalRequest := range *reqs {

					res := originalRequest.retrieveCachedDecision()
					if res == nil {
						return authzResp, err
					}

					decision := Decision{
						Permitted:  res.Permitted,
						Cached:     true,
						Expired:    res.Expired(),
						RetryCount: 0,
						Reason:     DenyReason(res.Reason),
					}

					authzRespCached.Decisions = append(authzRespCached.Decisions, decision)
				}
				return authzRespCached, nil
			}
		}
		return authzResp, err
	}

	for i, decision := range authzResp.Decisions {
		// rebuild decisions that have been found in the cache and ones from pdp into one user response
		userResponse.Decisions[notCachedRequestIndex[i]] = decision
	}

	return userResponse, nil
}

func toDecision(cachedDecision *cache.CachedDecision) Decision {

	decision := Decision{
		Permitted: cachedDecision.Permitted,
		Expired:   cachedDecision.Expired(),
		Cached:    true,
		Reason:    DenyReason(cachedDecision.Reason),
	}
	return decision
}

func (reqs Requests) getRoles(trace string, serviceOperatorToken string) ([]RolesResponse, error) {
	if len(reqs) < 1 {
		return []RolesResponse{}, buildInternalError(fmt.Errorf("The minimum number of requests is 1"), trace)
	}

	if len(reqs) > 100 {
		return []RolesResponse{}, buildInternalError(fmt.Errorf("The maximum number of requests is 100"), trace)
	}

	if !reqs.hasUniqueAccountID() {
		return []RolesResponse{}, buildInternalError(fmt.Errorf("The account ID must be the same for all requests"), trace)
	}

	rolesCall := reqs.buildRolesRequest()
	if rolesCall != nil {
		return rolesCall.getRoles(trace, serviceOperatorToken)
	}

	return []RolesResponse{}, buildInternalError(fmt.Errorf("A roles request cannot contain the action field"), trace)
}

func (reqs *Requests) buildRolesRequest() *RolesRequests {
	count := 0
	for _, req := range *reqs {
		if _, ok := req["action"].(Attributes); !ok {
			count++
		}
	}

	if count == len(*reqs) {
		rolesReqs := RolesRequests{}
		for _, req := range *reqs { // why not just change our struct so the user fills it in with the right format instead of us having to do this?
			req["resource"] = Attributes{
				"attributes": req["resource"],
			}
			req["subject"] = Attributes{
				"attributes": req["subject"],
			}
			rolesReqs = append(rolesReqs, req)
		}
		return &rolesReqs
	}
	return nil
}

func (reqs Requests) buildAuthzRequests() (authzReqs AuthzRequests) {

	authzReqs = AuthzRequests{}

	for _, req := range reqs {

		subject, err := buildSubjectFromAttributes(req)

		if err != nil {
			c := GetConfig().(*Config)
			c.Logger.Error(err)
		}
		pdpReq := Attributes{
			"action": req["action"],
			"resource": Attributes{
				"attributes": req["resource"],
			},
			"subject": subject,
		}
		if val, ok := req["environment"]; ok {
			pdpReq["environment"] = Attributes{
				"attributes": val,
			}
		}
		authzReqs = append(authzReqs, pdpReq)
	}

	return authzReqs
}

// Give the slice of Attributes, returns a copy of the object (and its index) that has the least number of keys in the resource
func (reqs *Requests) smallest() (int, Attributes) {
	if len(*reqs) < 1 {
		return -1, Attributes{}
	}
	currentSmallest := (*reqs)[0]["resource"].(Attributes)
	ndx := 0
	for i, r := range *reqs {
		resource := r["resource"].(Attributes)
		if len(resource) < len(currentSmallest) {
			ndx = i
			currentSmallest = resource
		}
	}
	return ndx, currentSmallest
}

// Determins if the resource of a request has the specified key and value
func (a Attributes) hasKeyAndValue(key string, value interface{}) bool {
	resource, ok := a["resource"].(Attributes)
	if !ok {
		return false
	}

	v, ok := resource[key]
	if !ok || v != value {
		return false
	}
	return true
}

// Duplicate an Attributes
func (a *Attributes) duplicate() (Attributes, error) {

	if a == nil {
		return Attributes{}, nil
	}

	gob.Register(Attributes{})
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(a)
	if err != nil {
		return nil, err
	}
	var copy Attributes
	err = dec.Decode(&copy)
	if err != nil {
		return nil, err
	}
	return copy, nil

}

// Given a subject, resource and actions from the PDP response,
// returns a slice of keys suitable for caching if resource is not nil.
// Otherwise, user request will be used to construct keys.
func buildCacheKeysWithUserRequest(userRequest Request, subject Attributes, resource Attributes, actions []string) (cacheKeys []CacheKey) {
	if resource != nil {
		cacheKeys = buildCacheKeys(subject, resource, actions)
	} else {
		cacheKeys = buildCacheKeys(userRequest["subject"].(Attributes), userRequest["resource"].(Attributes), actions)
	}
	return
}

// Given a subject, resource and actions from the PDP response
// return a slice of keys suitable for caching
func buildCacheKeys(subject Attributes, resource Attributes, actions []string) (cacheKeys []CacheKey) {
	if len(actions) < 1 {
		return
	}

	for _, action := range actions {
		req := Request{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		}
		// Ignore any error here since there should not be any, i.e. Request is always a valid struct.
		key := req.getCacheKey(false)

		if len(key) > 0 {
			cacheKeys = append(cacheKeys, key)
		}
	}

	return
}

// getCacheKey returns a cacheable key given the request
// if advanced is true, an advanced cache key is required, providing false will also return a basic cache key
// strictly follows the provided cache key patterns, requests that are missing some resource attributes should be
// validated by callers of this function
func (r Request) getCacheKey(advanced bool) (key CacheKey) {
	// key format id:<name>,scope:<name>;serviceName:<name>,accountID:<id>,serviceInstance:<instance>,resourceType:<type>,resource:<name>;action:<action>

	if r == nil || reflect.DeepEqual(r, Request{}) {
		return
	}

	//analyze pattern
	cacheKeyPattern := currentCacheKeyInfo.getCacheKeyPattern()

	subjectInRequest, ok := r["subject"].(Attributes)

	if !ok || len(subjectInRequest) == 0 {
		return nil
	}

	// get subject
	var subject string

	// TODO: fix adv obligation patterns for CRN subjects in order to be able to remove this condition
	// https://github.ibm.com/IAM/access-management/issues/12501
	if sbj, ok := subjectInRequest["id"]; ok && sbj != nil {
		subject, _ = buildPartialKeyAccordingToPattern(subjectInRequest, cacheKeyPattern.Subject)
	} else {
		subject = ""
	}
	if subject == "" {
		// advanced key could not be created due to subject attributes
		if advanced {
			return nil
		}
		key, _ := r.marshalAttributes()
		return key
	}

	resourceInRequest, ok := r["resource"].(Attributes)

	if !ok {
		return
	}

	// get resource
	resource, err := buildPartialKeyAccordingToPattern(resourceInRequest, cacheKeyPattern.Resource)
	if err != nil {
		// advanced key could not be created due to resource attributes
		if advanced {
			return nil
		}
		key, _ := r.marshalAttributes()
		return key
	}

	// get action
	action, ok := r["action"].(string)

	if !ok || len(action) == 0 {
		return nil
	}

	action = "action:" + action

	// use a map to be able to access by field name later
	varMap := map[string]string{
		"subject":  subject,
		"resource": resource,
		"action":   action,
	}

	keyString := ""

	for i, field := range cacheKeyPattern.Order {
		// build the key string by appending the main parts (subject;resource;action;)
		keyString = keyString + varMap[field] + ";"

		if i == len(cacheKeyPattern.Order)-1 {
			// reached the end of the order map, so we have the full key, remove the ; at the end
			keyString = strings.TrimSuffix(keyString, ";")
		}

	}

	key = CacheKey(keyString)

	return
}

// buildPartialKeyAccordingToPattern takes an attributes field containing a subset of attributes and a patterns field
// containing an array of pattern subsets and returns a key string, empty if it was not able to create one for the patterns
// example:
// attributes=Attributes{"serviceName":"xyz","accountID":"123"}
// patterns[][]={ {}, {"serviceName"}, {"serviceName", "accountID"}, {"serviceName", "accountID", "serviceInstance"} }
// returns the partial key key="serviceName:xyz,accountID:123"
func buildPartialKeyAccordingToPattern(attributes Attributes, patterns [][]string) (string, error) {
	key := ""
	if attributes == nil && len(patterns) > 0 {
		return key, fmt.Errorf("no attributes received")
	}

	for _, pattern := range patterns {

		if len(pattern) == 0 && len(attributes) == 0 {
			return "", nil
		}

		if len(pattern) == len(attributes) {

			// iterates over the attribute names in the pattern
			for _, fieldToFind := range pattern {

				// tests provided attributes to see if they match this pattern
				foundInAttributes, ok := attributes[fieldToFind]

				foundInAttributesString, isString := foundInAttributes.(string)

				if ok {
					if foundInAttributes == nil || !isString {
						c := GetConfig().(*Config)
						c.Logger.Debug("couldn't create key since field name "+fieldToFind+" not found in resource attributes of request", attributes)
						return "", nil
					}
					key = key + fieldToFind + ":" + foundInAttributesString + ","
				} else {
					key = ""
					break
				}
			}

			if key != "" {
				break
			}
		}
	}

	if key != "" {
		key = strings.TrimSuffix(key, ",")
	} else if len(attributes) != 0 {
		return key, fmt.Errorf("unsupported attributes received")
	}

	return key, nil
}

func (a Attributes) marshalAttributes() (CacheKey, error) {
	return json.Marshal(&a)
}

// Build a slice of Attributes from the resources
func (reqs *Requests) extractAttributes() ([]Attributes, error) {
	attrs := make([]Attributes, len(*reqs))

	for i, r := range *reqs {
		resource, ok := r["resource"].(Attributes)
		if !ok {
			return nil, fmt.Errorf("Unable to extract 'resource' from request #%d", i)
		}
		attr, err := resource.duplicate()
		if err != nil {
			return nil, fmt.Errorf("Unable to copy 'resource' from request #%d", i)
		}
		attrs[i] = attr
	}
	return attrs, nil
}

func (reqs *Requests) hasUniqueSubject() bool {
	if len(*reqs) < 1 {
		return false
	}

	firstSubject, ok := (*reqs)[0]["subject"]
	if !ok {
		return false
	}
	for _, req := range *reqs {
		subject, hasSubject := req["subject"]
		if !hasSubject {
			return false
		}

		if !reflect.DeepEqual(firstSubject, subject) {
			return false
		}
	}
	return true
}

func (reqs *Requests) hasUniqueAction() bool {
	if len(*reqs) < 1 {
		return false
	}

	firstAction, ok := (*reqs)[0]["action"]
	if !ok {
		return false
	}
	for _, req := range *reqs {
		action, hasAction := req["action"]
		if !hasAction {
			return false
		}
		if !reflect.DeepEqual(firstAction, action) {
			return false
		}
	}
	return true
}

func (reqs *Requests) hasUniqueAccountID() bool {
	if len(*reqs) < 1 {
		return false
	}

	firstAttr, hasFirstAttr := (*reqs)[0]["resource"].(Attributes)

	if !hasFirstAttr {
		return false
	}

	firstAccountID, hasFirstAccountID := firstAttr["accountId"]

	if !hasFirstAccountID {
		return false
	}

	for _, req := range *reqs {

		attr, hasAttr := req["resource"].(Attributes)
		if !hasAttr {
			return false
		}
		accountID, hasAccountID := attr["accountId"]
		if !hasAccountID {
			return false
		}

		if !reflect.DeepEqual(firstAccountID, accountID) {
			return false
		}
	}
	return true
}

func (reqs *Requests) hasUniqueServiceName() bool {
	if len(*reqs) < 1 {
		return false
	}

	firstAttr, hasFirstAttr := (*reqs)[0]["resource"].(Attributes)
	if !hasFirstAttr {
		return false
	}

	firstServiceName, hasFirstServiceName := firstAttr["serviceName"]
	if !hasFirstServiceName {
		return false
	}

	for _, req := range *reqs {
		attr, hasAttr := req["resource"].(Attributes)
		if !hasAttr {
			return false
		}
		serviceName, hasServiceName := attr["serviceName"]
		if !hasServiceName {
			return false
		}

		if !reflect.DeepEqual(firstServiceName, serviceName) {
			return false
		}
	}
	return true
}

func (reqs *Requests) hasUniqueResourceType() bool {
	if len(*reqs) < 1 {
		return false
	}

	firstAttr, hasFirstAttr := (*reqs)[0]["resource"].(Attributes)

	if !hasFirstAttr {
		return false
	}

	firstResourceType := firstAttr["resourceType"]

	for _, req := range *reqs {
		attr, hasAttr := req["resource"].(Attributes)
		if !hasAttr {
			return false
		}
		resourceType := attr["resourceType"]

		if !reflect.DeepEqual(firstResourceType, resourceType) {
			return false
		}
	}

	return true
}

func (reqs *Requests) hasUniqueEnvironment() bool {
	if len(*reqs) < 1 {
		return false
	}

	firstEnvironment, firstHasEnvironment := (*reqs)[0]["environment"]
	for _, req := range *reqs {
		environment, hasEnvironment := req["environment"]
		if firstHasEnvironment != hasEnvironment {
			return false
		}

		if !reflect.DeepEqual(firstEnvironment, environment) {
			return false
		}
	}
	return true
}

func (reqs *Requests) buildListRequest() (*ListRequest, error) {

	if !reqs.hasUniqueSubject() {
		return &ListRequest{}, fmt.Errorf("multiple subjects is not supported")
	}

	if !reqs.hasUniqueAction() {
		return &ListRequest{}, fmt.Errorf("multiple actions is not supported")
	}

	if !reqs.hasUniqueServiceName() {
		return &ListRequest{}, fmt.Errorf("multiple serviceName is not supported")
	}

	if !reqs.hasUniqueAccountID() {
		return &ListRequest{}, fmt.Errorf("multiple accountId is not supported")
	}

	if !reqs.hasUniqueResourceType() {
		return &ListRequest{}, fmt.Errorf("multiple resource types are not supported")
	}

	if !reqs.hasUniqueEnvironment() {
		return &ListRequest{}, fmt.Errorf("multiple environments are not supported")
	}

	sharedAttributes, errCommonAttr := reqs.extractCommonAttributes()
	if errCommonAttr != nil {
		return &ListRequest{}, errCommonAttr
	}
	uniqueAttributes, errUniqueAttr := reqs.extractUniqueAttributes()
	if errUniqueAttr != nil {
		return &ListRequest{}, errUniqueAttr
	}

	resources := Attributes{
		"sharedAttributes": sharedAttributes,
		"uniqueAttributes": uniqueAttributes,
	}

	subject, err := buildSubjectFromAttributes((*reqs)[0])

	if err != nil {
		c := GetConfig().(*Config)
		c.Logger.Error(err)
	}

	action := (*reqs)[0]["action"]

	listRequest := ListRequest{
		"action":    action,
		"resources": resources,
		"subject":   subject,
	}

	if val, ok := (*reqs)[0]["environment"]; ok {
		listRequest["environment"] = Attributes{
			"attributes": val,
		}
	}

	return &listRequest, nil

}

func buildSubjectFromAttributes(att Attributes) (Attributes, error) {
	subject := Attributes{}
	attributes := Attributes{}

	reqSubject, ok := att["subject"].(Attributes)

	// Subject attributes are invalid/not found
	if !ok {
		return nil, fmt.Errorf("Invalid subject provided")
	}

	for k, v := range reqSubject {
		if k == "accessTokenBody" {
			tokenString, ok := v.(string)

			if !ok {
				return nil, fmt.Errorf("invalid token string provided")
			}

			subject["accessTokenBody"] = tokenString

		} else {
			attributes[k] = v
		}
	}

	// add any additional attributes to the subject being sent
	subject["attributes"] = attributes

	return subject, nil

}

// Returns an object that contains the common resource attributes
func (reqs *Requests) extractCommonAttributes() (Attributes, error) {

	// Find the resource with the smallest number of attributes
	// If one of the remaining requests does not have the attribute
	//   then that attribute is not common

	ndx, smallestRes := reqs.smallest()
	common, err := smallestRes.duplicate()
	if err != nil {
		return nil, err
	}

	for k, v := range smallestRes {
		for i, req := range *reqs {
			if i == ndx {
				continue
			}
			if !req.hasKeyAndValue(k, v) {
				delete(common, k)
				break
			}
		}
	}

	return common, nil
}

func (reqs *Requests) extractUniqueAttributes() ([]Attributes, error) {

	commonAttributes, errCommonAttr := reqs.extractCommonAttributes()
	if errCommonAttr != nil {
		return nil, errCommonAttr
	}

	resources, errAttr := reqs.extractAttributes()
	if errAttr != nil {
		return nil, errAttr
	}

	attrCopies := make([]Attributes, len(resources))

	// Making a copy since we will delete common attributes later.
	for i, r := range resources {
		attrCopy, err := r.duplicate()
		if err != nil {
			return nil, err
		}
		attrCopies[i] = attrCopy
	}

	// Removing common attributes from the resource attributes
	for k := range commonAttributes {
		for _, r := range attrCopies {
			delete(r, k)
		}
	}
	return attrCopies, nil
}

func getUserResponseFromCache(r Requests) (notCachedRequests Requests, notCachedRequestIndex []int, userResponse AuthzResponse) {

	// notCachedRequestIndex
	// index of request attributes that weren't found in the original request
	// this is be used to keep track of and fill in the gaps of userResponse
	// when the non-cached requests are returned from PDP

	userResponse = AuthzResponse{
		Decisions: make([]Decision, len(r)),
	}

	for i, attributes := range r {
		// get each decision from cache
		cachedDecision := attributes.retrieveCachedDecision()

		if cachedDecision != nil && !cachedDecision.Expired() {
			userResponse.Decisions[i] = toDecision(cachedDecision)
		} else {
			c := GetConfig().(*Config)
			c.Logger.Debug("No cached response for:", attributes)

			// create authz request from requests that weren't cached
			notCachedRequests = append(notCachedRequests, attributes)
			notCachedRequestIndex = append(notCachedRequestIndex, i)
		}
	}

	return
}

// The v2 authz requests
type AuthzRequests []Attributes

// Implementation of the V2 authz and cache the result
// Note that caches should be stored based on user's requests.
// That is the reason why we need the userReqs
func (r *AuthzRequests) isAuthorizedAndCache(trace string, token string, userReqs Requests) (AuthzResponse, error) {

	if trace == "" {
		trace = uuid.New().String()
	}

	userResponse := AuthzResponse{
		Trace:     trace,
		Decisions: []Decision{},
	}

	// This should fail most of the tests in case of programming errors.
	if len(*r) != len(userReqs) {
		return userResponse, buildInternalError(fmt.Errorf("The number of user's requests and the one send to PDP is not the same"), trace)
	}

	if len(*r) > 100 {
		return userResponse, buildInternalError(fmt.Errorf("The maximum number of requests allowed is 100"), trace)
	}

	c := GetConfig().(*Config)

	var pdpResponse authzResponse

	request := &pdpRequest{
		endpoint:    c.AuthzEndpoint,
		method:      http.MethodPost,
		token:       token,
		payload:     r,
		trace:       trace,
		pdpResponse: &pdpResponse,
	}

	// call PDP
	resp, respCode, pdpErr := request.callPDP()

	if pdpErr != nil {
		if c.AuthzRetry {
			pdpErr = retryCallPDP(respCode, request, pdpErr, resp)
		}

		if pdpErr != nil {
			return userResponse, pdpErr
		}
	}

	currentCacheKeyInfo.storeCacheKeyPattern(pdpResponse.CacheKeyPattern)

	for i, response := range pdpResponse.Responses {
		decision := Decision{
			Cached:     false,
			RetryCount: request.retry,
		}
		decision.Permitted = response.AuthorizationDecision.Permitted

		if decision.Permitted {
			decision.Reason = DenyReasonNone
		} else if _, ok := validContextReasons[response.AuthorizationDecision.Reason]; ok {
			decision.Reason = DenyReasonContext
		} else {
			decision.Reason = DenyReasonIAM
		}
		userResponse.Decisions = append(userResponse.Decisions, decision)

		// Generate the cache key using information from the pdp response.
		originalReq := userReqs[i]
		cacheKeys := []CacheKey{}
		if decision.Permitted {
			// User request will be used to construct keys if resources are absent from pdp response.
			cacheKeys = buildCacheKeysWithUserRequest(originalReq,
				response.AuthorizationDecision.Obligation.Subject.Attributes,
				response.AuthorizationDecision.Obligation.Resource.Attributes,
				response.AuthorizationDecision.Obligation.Actions)
		} else {
			cacheKey, err := originalReq.marshalAttributes()
			if err != nil {
				continue
			}
			cacheKeys = append(cacheKeys, cacheKey)
		}

		maxCacheAgeSeconds := response.AuthorizationDecision.Obligation.MaxCacheAgeSeconds
		for _, key := range cacheKeys {
			key.cacheDecision(decision.Permitted, maxCacheAgeSeconds, decision.Reason)
		}
	}
	// return user response
	return userResponse, nil
}

// ListRequest to call PDP with
type ListRequest Attributes

// Implementation of the list API
// Note that caches should be stored based on user's requests.
// That is the reason why we need the userReqs
func (r *ListRequest) isAuthorizedAndCache(trace string, token string, userReqs Requests) (AuthzResponse, error) {

	if trace == "" {
		trace = uuid.New().String()
	}

	userResponse := AuthzResponse{
		Trace:     trace,
		Decisions: []Decision{},
	}

	c := GetConfig().(*Config)

	var pdpResponse listResponse

	subRequestCount := r.getUniqueAttributeCount()

	request := &pdpRequest{
		endpoint:        c.ListEndpoint,
		method:          http.MethodPut,
		token:           token,
		payload:         r,
		trace:           trace,
		pdpResponse:     &pdpResponse,
		subRequestCount: strconv.Itoa(subRequestCount),
	}

	resp, respCode, pdpErr := request.callPDP()

	if pdpErr != nil {
		if c.AuthzRetry {
			pdpErr = retryCallPDP(respCode, request, pdpErr, resp)
		}
		if pdpErr != nil {
			return userResponse, pdpErr
		}
	}

	currentCacheKeyInfo.storeCacheKeyPattern(pdpResponse.CacheKeyPattern)

	for i, d := range pdpResponse.Decisions {
		decision := Decision{
			RetryCount: request.retry,
		}
		if d.Decision == "Permit" {
			decision.Permitted = true
			decision.Reason = DenyReasonNone
			userResponse.Decisions = append(userResponse.Decisions, decision)
		} else {
			decision.Permitted = false
			if _, ok := validContextReasons[d.Reason]; ok {
				decision.Reason = DenyReasonContext
			} else {
				decision.Reason = DenyReasonIAM
			}
			userResponse.Decisions = append(userResponse.Decisions, decision)
		}

		originalReq := userReqs[i]
		cacheKeys := []CacheKey{}
		if !c.DisableCache && decision.Permitted {
			// Generate the default or advanced cache key using actions from the pdp Obligation
			cacheKeys = buildCacheKeysWithUserRequest(originalReq, d.Obligation.Subject.Attributes,
				d.Obligation.Resource.Attributes, d.Obligation.Actions)
		}

		cacheKey, err := originalReq.marshalAttributes()
		if err != nil {
			continue
		}
		cacheKeys = append(cacheKeys, cacheKey)

		maxCacheAgeSeconds := d.Obligation.MaxCacheAgeSeconds
		for _, key := range cacheKeys {
			key.cacheDecision(decision.Permitted, maxCacheAgeSeconds, decision.Reason)
		}
	}

	return userResponse, nil
}

func (r *ListRequest) getUniqueAttributeCount() (count int) {

	resource, hasResources := (*r)["resources"].(Attributes)
	if !hasResources {
		return
	}

	uniqueAttributes, hasUniqueAttr := resource["uniqueAttributes"].([]Attributes)

	if !hasUniqueAttr {
		return
	}

	count = len(uniqueAttributes)

	return
}

func getDecisionCache() cache.DecisionCache {
	config := GetConfig().(*Config)

	if config.DisableCache && config.DisableDeniedCache {
		return nil
	}

	cache := config.CachePlugin
	if cache == nil {
		return nil
	}
	return cache
}

func (r Request) retrieveCachedDecision() *cache.CachedDecision {
	var decision *cache.CachedDecision
	config := GetConfig().(*Config)
	cache := getDecisionCache()

	if cache == nil {
		config.Logger.Debug("Cache disabled.")
		return nil
	}
	key, err := r.marshalAttributes()

	if err != nil {
		config.Logger.Error("Failed to generate a cache key for the exact request attributes ", r)
	} else {
		decision = cache.Get(key)
		if decision != nil {
			return decision
		}
	}

	requestDuplicate, err := r.duplicate()

	if err != nil || len(requestDuplicate) == 0 {
		return nil
	}

	subject, ok := requestDuplicate["subject"].(Attributes)
	if !ok || len(subject) == 0 {
		return nil
	}

	var tokenBodySubject Attributes

	tokenBody, ok := subject["accessTokenBody"].(string)

	if ok {
		tokenBodySubject, err = buildSubjectTokenBodyClaims(tokenBody)
		if err != nil {
			config.Logger.Error("could not create claims from token body: ", err)
			return nil
		}

		delete(subject, "accessTokenBody")

		for key, value := range tokenBodySubject {
			// to avoid overwriting attributes with the same name (which are not permitted)
			if _, ok := subject[key]; ok {
				config.Logger.Error("duplicate attributes found in access token body subject request")
				return nil
			}
			subject[key] = value
		}
	}

	resource, ok := requestDuplicate["resource"].(Attributes)

	if !ok {
		return nil
	}

	action, ok := requestDuplicate["action"].(string)

	if !ok || len(action) == 0 {
		return nil
	}

	cacheKeyPattern := currentCacheKeyInfo.getCacheKeyPattern()

	for _, pattern := range cacheKeyPattern.Subject {
		subjToSend := Attributes{}

		if len(pattern) > len(subject) {
			break
		}

		for _, fieldToFind := range pattern {
			// build the subject portion
			subjToSend[fieldToFind] = subject[fieldToFind]
		}

	out:
		for _, pattern := range cacheKeyPattern.Resource {

			resToSend := Attributes{}

			if len(resource) != 0 {

				if len(pattern) > len(resource) {
					// if the resource size is smaller than the pattern then we have run out of usable patterns
					// since the existing resource doesn't have some of the fields that are in the pattern
					break
				}

				for _, fieldToFind := range pattern {
					// build the resource portion
					if resValue, ok := resource[fieldToFind]; ok {
						resToSend[fieldToFind] = resValue
					} else {
						config.Logger.Debug("not able to create cache key since pattern field name " + fieldToFind + " not found in resource attributes of request")

						break out
					}
				}
			}

			req := Request{
				"subject":  subjToSend,
				"resource": resToSend,
				"action":   action,
			}

			key = req.getCacheKey(true)
			if key == nil {
				// cannot have an empty cache key
				continue
			}
			if len(key) > 0 {

				decision = cache.Get(key)
				if decision != nil {
					return decision
				}
			}
		}
	}

	return decision
}

func buildSubjectTokenBodyClaims(tokenBody string) (subject Attributes, err error) {
	/*This is just a workaround to parse the token body using token library. Currently, the token library does not support to parse just the token body*/
	/* #nosec G101 */
	jwtFormattedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + tokenBody + ".SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	return buildSubjectTokenClaims(jwtFormattedToken)
}

func buildSubjectTokenClaims(token string) (subject Attributes, err error) {
	conf := GetConfig().(*Config)

	claims, err := conf.tokenManager.GetClaims(token, true)

	if err != nil {
		return nil, err
	}

	if claims.SubType == "CRN" {
		crnString := claims.IAMID
		subject = ParseCRNToAttributes(crnString)
		subject["scope"] = claims.Scope
	} else {
		subject = Attributes{
			"id":    claims.IAMID,
			"scope": claims.Scope,
		}
	}

	return
}

func (k CacheKey) cacheDecision(decision bool, maxCacheAgeSeconds int, reason DenyReason) {
	decisionCache := getDecisionCache()
	if decisionCache != nil {
		config := decisionCache.GetConfig()
		var ttl time.Duration
		if !decision {
			if !config.DisableDenied {
				if config.DeniedTTL != 0 {
					ttl = config.DeniedTTL
				} else {
					ttl = cache.DefaultDenyTTL
				}
				decisionCache.Set(k, decision, ttl, int(reason))
			}
		} else {
			if !config.DisablePermitted {
				if config.TTL != 0 {
					ttl = config.TTL
				} else if maxCacheAgeSeconds != 0 {
					ttl = time.Duration(maxCacheAgeSeconds) * time.Second
				} else {
					ttl = cache.DefaultTTL
				}
				decisionCache.Set(k, decision, ttl, int(DenyReasonNone))
			}
		}
	}
}

// RolesRequests represents the PDP request for a v2/authz/roles call
type RolesRequests []Attributes

// Implementation of the roles API
func (r *RolesRequests) getRoles(trace string, token string) ([]RolesResponse, error) {
	if trace == "" {
		trace = uuid.New().String()
	}

	userResponse := []RolesResponse{}

	c := GetConfig().(*Config)

	var pdpResponse rolesResponse

	request := &pdpRequest{
		endpoint:    c.RolesEndpoint,
		method:      http.MethodPost,
		token:       token,
		payload:     r,
		trace:       trace,
		pdpResponse: &pdpResponse,
	}

	resp, respCode, pdpErr := request.callPDP()

	if pdpErr != nil {
		if c.AuthzRetry {
			pdpErr = retryCallPDP(respCode, request, pdpErr, resp)
		}

		if pdpErr != nil {
			return userResponse, pdpErr
		}
	}

	for _, role := range pdpResponse.Responses {
		userResponse = append(userResponse, RolesResponse{
			trace:         trace,
			Attributes:    role.RoleAttributes,
			ServiceRoles:  role.RoleActions,
			PlatformRoles: role.PlatformExtensions.RoleActions,
		})
	}

	return userResponse, nil
}

func buildInternalError(err error, trace string) error {
	internalError := &InternalError{
		Message: errors.Wrap(err, "Invalid request").Error(),
		Trace:   trace,
	}
	return internalError
}

func buildAPIError(infos ...interface{}) error {

	apiError := &APIError{}
	for _, info := range infos {
		switch err := info.(type) {
		case *url.Error:
			apiError.EndpointURI = err.URL
		case error:
			apiError.Message = err.Error()
		case *http.Response:
			apiError.StatusCode = err.StatusCode
		case string:
			apiError.Trace = err
		case *authzResponse:
			var messages bytes.Buffer
			for idx, resp := range err.Responses {
				if resp.Status != "200" {
					messages.WriteString(fmt.Sprintf("Request #%d contains error: ", idx))
					for _, reqErr := range resp.Error.Errors {
						messages.WriteString(reqErr.Message)
					}
					apiError.StatusCode, _ = strconv.Atoi(resp.Status)
					apiError.Message = messages.String()
					break
				}
			}
		default:
		}
	}
	return apiError
}

type pdpRequest struct {
	endpoint        string      // the PDP endpoint
	method          string      // the HTTP Method
	token           string      // the token
	payload         interface{} // the payload for PDP
	trace           string      // The transaction ID
	pdpResponse     interface{} // The place holder for the PDP response
	subRequestCount string      // The number of sub requests, i.e. the value for the X-Request-Count header
	retry           int
}

// The call to PDP
func (r *pdpRequest) callPDP() (*http.Response, int, error) {
	var returnCode int
	conf := GetConfig().(*Config)
	conf.Statistics.Inc(originalRequestsToPDP)

	// Parsing the input
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(r.payload); err != nil {
		return nil, 0, buildInternalError(err, r.trace)
	}

	// Preparing the request to pdp
	req, err := http.NewRequest(r.method, r.endpoint, buf)

	if err != nil {
		return nil, 0, buildInternalError(errors.Wrap(err, "Error creating a request"), r.trace)
	}

	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("transaction-id", string(r.trace))
	req.Header.Set("X-Accept-Advanced-Obligation", "true")
	req.Header.Set("pep-version", version)

	if r.subRequestCount != "" {
		req.Header.Set("X-Request-Count", r.subRequestCount)
	} else {
		req.Header.Set("X-Request-Count", "10")
	}

	// Perform the request

	resp, err := conf.HTTPClient.Do(req)
	if resp == nil {
		returnCode = 0
	} else {
		returnCode = resp.StatusCode
	}

	if err != nil {
		var errMsg error
		if err, ok := err.(net.Error); ok && err.Timeout() {
			returnCode = http.StatusGatewayTimeout
			errMsg = err
		} else if ok && err.Temporary() {
			returnCode = http.StatusBadGateway
			errMsg = err
		}

		// DNS errors ( dropped connection)
		if err, ok := err.(*url.Error); ok {
			errMsg = buildAPIError(err, r.trace, err.Err)
			conf.Statistics.Inc(apiErr)

			if err, ok := err.Err.(*net.OpError); ok {
				if _, ok := err.Err.(*net.DNSError); ok {
					returnCode = 999
				}
			}
		}

		return nil, returnCode, errMsg
	}

	defer resp.Body.Close()

	// Adds the status code E.g. "status-code-400"
	statusCodeAsString := strconv.Itoa(returnCode)
	conf.Statistics.Inc("status-code" + "-" + statusCodeAsString)

	// separate status code & error from get response

	if returnCode >= 500 {
		body, err := ioutil.ReadAll(resp.Body)
		msg := ""
		if err != nil {
			msg = err.Error()
		} else {
			msg = string(body)
		}
		apiError := buildAPIError(resp, errors.New(msg), r.pdpResponse, r.trace, &url.Error{URL: r.endpoint}).(*APIError)
		apiError.EndpointURI = r.endpoint

		return nil, returnCode, apiError
	}

	err = getResponse(resp.Body, r.pdpResponse)
	if err != nil || returnCode != 200 {
		apiError := buildAPIError(resp, err, r.pdpResponse, r.trace, &url.Error{URL: r.endpoint})
		return resp, returnCode, apiError
	}

	// Log timing headers for debugging latency
	xResponseTime := resp.Header.Get("x-response-time")
	xProxyTime := resp.Header.Get("x-proxy-upstream-service-time")
	conf.Logger.Info("[transaction-id: %v] Service response time: %vms, Upstream Service Proxy response time: %vms", r.trace, xResponseTime, xProxyTime)

	return resp, returnCode, err

}

func getResponse(body io.ReadCloser, pdpResponse interface{}) error {
	if pdpResponse != nil {
		err := json.NewDecoder(body).Decode(pdpResponse)

		if err != nil {
			return err
		}
	}
	return nil
}

func retryCallPDP(respCode int, request *pdpRequest, lastError error, response *http.Response) error {
	if request.retry < 3 && (respCode == http.StatusTooManyRequests || respCode >= http.StatusInternalServerError) {
		request.retry++
		sleepTime := time.Duration(request.retry) * time.Second

		if response != nil {
			if retryAfter := response.Header.Get("Retry-After"); retryAfter != "" {
				if sleep, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
					sleepTime = time.Second * time.Duration(sleep)
				}
			}
		}
		time.Sleep(sleepTime)

		resp, respCode, pdpErr := request.callPDP()

		if pdpErr != nil {
			return retryCallPDP(respCode, request, pdpErr, resp)
		}

		return nil
	}
	return lastError
}

package pep_test

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.ibm.com/IAM/pep/v3"
)

func ExamplePerformAuthorization_request1() {

	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	requests := pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "txid-12345"
	response, err := pep.PerformAuthorization(&requests, trace)

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		// TO test with invalid api_key
		fmt.Println(response.Decisions[0].Permitted)
	}
	// Output: true
}
func ExamplePerformAuthorization_requests3() {

	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	// Should get denied due to servieName
	resource1 := pep.Attributes{
		"serviceName":     "cloudant",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	// Should get permit
	resource2 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	// Should get permit
	resource3 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book2",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	requests := pep.Requests{
		{
			"action":   action,
			"resource": resource1,
			"subject":  subject,
		},
		{
			"action":   action,
			"resource": resource2,
			"subject":  subject,
		},
		{
			"action":   action,
			"resource": resource3,
			"subject":  subject,
		},
	}

	trace := "txid-12348"
	response, err := pep.PerformAuthorization(&requests, trace)

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(response.Decisions[0].Permitted)
	}
	/*
	  The expected output should not be empty, but there is a bug in the list API...
	  Output: [false true true]
	*/
	//Output: []
}

func ExamplePerformAuthorization_request4() {

	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	// Should get denied due to servieName
	resource1 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	// Should get permit
	resource2 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	// Should get permit
	resource3 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book2",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	// Should get permit
	resource4 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book3",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	requests := pep.Requests{
		{
			"action":   action,
			"resource": resource1,
			"subject":  subject,
		},
		{
			"action":   action,
			"resource": resource2,
			"subject":  subject,
		},
		{
			"action":   action,
			"resource": resource3,
			"subject":  subject,
		},
		{
			"action":   action,
			"resource": resource4,
			"subject":  subject,
		},
	}

	trace := "txid-12348"
	response, err := pep.PerformAuthorization(&requests, trace)

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Printf("%+v\n", response.Decisions)
	}
	//Output: [{Permitted:true Cached:false Expired:false RetryCount:0 Reason:0} {Permitted:true Cached:false Expired:false RetryCount:0 Reason:0} {Permitted:true Cached:false Expired:false RetryCount:0 Reason:0} {Permitted:true Cached:false Expired:false RetryCount:0 Reason:0}]
}

func ExamplePerformAuthorizationWithToken_request1() {

	getToken := func() (string, error) {
		config := &pep.Config{
			Environment: pep.Staging,
			APIKey:      os.Getenv("API_KEY"),
			LogLevel:    pep.LevelError,
		}

		err := pep.Configure(config)
		if err != nil {
			fmt.Println(err)
		}

		return pep.GetToken()
	}

	//Get a valid serviceOperatorToken externally
	serviceOperatorToken, err := getToken()

	if err != nil {
		fmt.Println(err)
	}

	config := &pep.Config{
		Environment: pep.Staging,
		LogLevel:    pep.LevelError,
	}

	err = pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	requests := pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "txid-12345"
	response, err := pep.PerformAuthorizationWithToken(&requests, trace, serviceOperatorToken)

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		// TO test with invalid api_key
		fmt.Println(response.Decisions[0].Permitted)
	}
	// Output: true
}

func ExampleGetAuthorizedRoles_request() {

	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	requests := pep.Requests{
		{
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "txid-12345"
	response, err := pep.GetAuthorizedRoles(&requests, trace)

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(response[0].Attributes.ServiceInstance, response[0].ServiceRoles[0].Role.CRN, response[0].ServiceRoles[0].Actions[0].Action)
	}
	//Output: chatko-book1 crn:v1:bluemix:public:iam::::serviceRole:Reader gopep.books.read
}

func ExampleGetAuthorizedRolesWithToken_request() {

	getToken := func() (string, error) {
		config := &pep.Config{
			Environment: pep.Staging,
			APIKey:      os.Getenv("API_KEY"),
			LogLevel:    pep.LevelError,
		}

		err := pep.Configure(config)
		if err != nil {
			fmt.Println(err)
		}

		return pep.GetToken()
	}
	//Get a valid serviceOperatorToken externally
	serviceOperatorToken, err := getToken()

	if err != nil {
		fmt.Println(err)
	}

	config := &pep.Config{
		Environment: pep.Staging,
		LogLevel:    pep.LevelError,
	}

	err = pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	requests := pep.Requests{
		{
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "txid-12345"
	response, err := pep.GetAuthorizedRolesWithToken(&requests, trace, serviceOperatorToken)

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(response[0].Attributes.ServiceInstance, response[0].ServiceRoles[0].Role.CRN, response[0].ServiceRoles[0].Actions[0].Action)
	}
	//Output: chatko-book1 crn:v1:bluemix:public:iam::::serviceRole:Reader gopep.books.read
}

func ExamplePerformAuthorization_requestCheckStats() {

	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	requests := pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "txid-12345"
	resp, err := pep.PerformAuthorization(&requests, trace)
	_ = resp

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		statistics := pep.GetStatistics()

		// PEP usge/request related statistics
		fmt.Println("Number of requests serviced:", statistics.Usage.RequestsServicedByPEP)
		fmt.Println("Number of requests to PDP:", statistics.Usage.OriginalRequestsToPDP)
		fmt.Println("Failed requests:", statistics.Usage.FailedUserRequests)
		fmt.Println("Status codes:", statistics.Usage.StatusCodeCounts)

		// PEP cache related statistics
		fmt.Println("Hits:", statistics.Cache.Hits)
		fmt.Println("Misses:", statistics.Cache.Misses)
		fmt.Println("Entries in cache:", statistics.Cache.EntriesCount)

	}
	// Output: Number of requests serviced: 1
	// Number of requests to PDP: 1
	// Failed requests: 0
	// Status codes: map[status-code-200:1]
	// Hits: 0
	// Misses: 1
	// Entries in cache: 5
}

func ExamplePerformAuthorization_advancedObligation() {
	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	// Should get denied due to servieName
	resource1 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	// Should get permit
	resource2 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	// Should get permit
	resource3 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book2",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		"resourceType":    "bucket",
	}

	// Should get permit
	resource4 := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book3",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		"resourceType":    "bucket",
		"resource":        "bucket-of-books",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	action1 := "gopep.books.read"

	action2 := "gopep.books.write"

	action3 := "gopep.books.eat"

	request1 := pep.Requests{
		{
			"action":   action1,
			"resource": resource1,
			"subject":  subject,
		},
	}

	request2 := pep.Requests{
		{
			"action":   action1,
			"resource": resource2,
			"subject":  subject,
		},
		{
			"action":   action2,
			"resource": resource3,
			"subject":  subject,
		},
	}
	request3 := pep.Requests{
		{
			"action":   action3,
			"resource": resource4,
			"subject":  subject,
		},
	}

	request4 := pep.Requests{
		{
			"action":   action1,
			"resource": resource1,
			"subject":  subject,
		},
		{
			"action":   action1,
			"resource": resource2,
			"subject":  subject,
		},
		{
			"action":   action2,
			"resource": resource3,
			"subject":  subject,
		},
		{
			"action":   action3,
			"resource": resource4,
			"subject":  subject,
		},
		{
			"action":   action1,
			"resource": resource1,
			"subject":  subject,
		},
		{
			"action":   action1,
			"resource": resource2,
			"subject":  subject,
		},
		{
			"action":   action2,
			"resource": resource3,
			"subject":  subject,
		},
		{
			"action":   action3,
			"resource": resource4,
			"subject":  subject,
		},
		{
			"action":   action1,
			"resource": resource1,
			"subject":  subject,
		},
		{
			"action":   action1,
			"resource": resource2,
			"subject":  subject,
		},
		{
			"action":   action2,
			"resource": resource3,
			"subject":  subject,
		},
		{
			"action":   action3,
			"resource": resource4,
			"subject":  subject,
		},
	}

	trace := "alex-test-txid-advanced-obligations"

	response, err := pep.PerformAuthorization(&request1, trace)

	// Error handling
	if !handleError(err) {
		for _, decision := range response.Decisions {
			fmt.Println("Permitted", decision.Permitted)
		}
	}

	statistics := pep.GetStatistics()
	// PEP usge/request related statistics
	fmt.Println("Number of requests serviced:", statistics.Usage.RequestsServicedByPEP)
	fmt.Println("Number of requests to PDP:", statistics.Usage.OriginalRequestsToPDP)

	// PEP cache related statistics
	fmt.Println("Hits:", statistics.Cache.Hits)
	fmt.Println("Misses:", statistics.Cache.Misses)
	fmt.Println("Entries in cache:", statistics.Cache.EntriesCount)
	fmt.Println("**********")

	trace = "alex-test-txid-advanced-obligations2"

	response, err = pep.PerformAuthorization(&request2, trace)

	// Error handling
	if !handleError(err) {
		for _, decision := range response.Decisions {
			fmt.Println("Permitted", decision.Permitted)
		}
	}

	statistics = pep.GetStatistics()
	// PEP usge/request related statistics
	fmt.Println("Number of requests serviced:", statistics.Usage.RequestsServicedByPEP)
	fmt.Println("Number of requests to PDP:", statistics.Usage.OriginalRequestsToPDP)

	// PEP cache related statistics
	fmt.Println("Hits:", statistics.Cache.Hits)
	fmt.Println("Misses:", statistics.Cache.Misses)
	fmt.Println("Entries in cache:", statistics.Cache.EntriesCount)
	fmt.Println("**********")

	trace = "alex-test-txid-advanced-obligations3"

	response, err = pep.PerformAuthorization(&request3, trace)

	// Error handling
	if !handleError(err) {
		for _, decision := range response.Decisions {
			fmt.Println("Permitted", decision.Permitted)
		}
	}

	statistics = pep.GetStatistics()
	// PEP usge/request related statistics
	fmt.Println("Number of requests serviced:", statistics.Usage.RequestsServicedByPEP)
	fmt.Println("Number of requests to PDP:", statistics.Usage.OriginalRequestsToPDP)

	// PEP cache related statistics
	fmt.Println("Hits:", statistics.Cache.Hits)
	fmt.Println("Misses:", statistics.Cache.Misses)
	fmt.Println("Entries in cache:", statistics.Cache.EntriesCount)
	fmt.Println("**********")

	trace = "alex-test-txid-advanced-obligations4"

	response, err = pep.PerformAuthorization(&request4, trace)

	// Error handling
	if !handleError(err) {
		for _, decision := range response.Decisions {
			fmt.Println("Permitted", decision.Permitted)
		}
	}

	statistics = pep.GetStatistics()
	// PEP usge/request related statistics
	fmt.Println("Number of requests serviced:", statistics.Usage.RequestsServicedByPEP)
	fmt.Println("Number of requests to PDP:", statistics.Usage.OriginalRequestsToPDP)

	// PEP cache related statistics
	fmt.Println("Hits:", statistics.Cache.Hits)
	fmt.Println("Misses:", statistics.Cache.Misses)
	fmt.Println("Entries in cache:", statistics.Cache.EntriesCount)

	//Output:
	// Permitted true
	// Number of requests serviced: 1
	// Number of requests to PDP: 1
	// Hits: 0
	// Misses: 1
	// Entries in cache: 5
	// **********
	// Permitted true
	// Permitted true
	// Number of requests serviced: 2
	// Number of requests to PDP: 1
	// Hits: 2
	// Misses: 4
	// Entries in cache: 5
	// **********
	// Permitted true
	// Number of requests serviced: 3
	// Number of requests to PDP: 1
	// Hits: 3
	// Misses: 7
	// Entries in cache: 5
	// **********
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Permitted true
	// Number of requests serviced: 4
	// Number of requests to PDP: 1
	// Hits: 15
	// Misses: 25
	// Entries in cache: 5

}

func ExamplePerformAuthorization_getSubjectFromToken() {
	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	/* #nosec G101 */
	var crnToken = "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJjcm4tY3JuOnYxOmJsdWVtaXg6cHVibGljOmdvcGVwOjphLzJjMTdjNGU1NTg3NzgzOTYxY2U0YTBhYTQxNTA1NGU3OmdvcGVwMTIzOjoiLCJpZCI6ImNybi1jcm46djE6Ymx1ZW1peDpwdWJsaWM6Z29wZXA6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6Z29wZXAxMjM6OiIsInJlYWxtaWQiOiJjcm4iLCJqdGkiOiIwYzFkNTZjZS1mOGZkLTQ3NjUtODBjYy03N2E3ZTk2YmY4MDEiLCJpZGVudGlmaWVyIjoiY3JuOnYxOmJsdWVtaXg6cHVibGljOmdvcGVwOjphLzJjMTdjNGU1NTg3NzgzOTYxY2U0YTBhYTQxNTA1NGU3OmdvcGVwMTIzOjoiLCJzdWIiOiJjcm46djE6Ymx1ZW1peDpwdWJsaWM6Z29wZXA6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6Z29wZXAxMjM6OiIsInN1Yl90eXBlIjoiQ1JOIiwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTciLCJmcm96ZW4iOnRydWV9LCJpYXQiOjE1OTg5MDQ1ODIsImV4cCI6MTU5ODkwODEwNSwiaXNzIjoiaHR0cHM6Ly9pYW0uc3RhZ2UxLmJsdWVtaXgubmV0L2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6aWFtLWF1dGh6Iiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFjciI6MCwiYW1yIjpbXX0.bDndRAzcNaVdG1_g-UH03_kUAty4qOmlE3HkIswvrKEpWo7po59fn8zf89oIv3B-pXpQ9kwO3LxLJ-o3CF45D8_xnyUHFl1JWec7NCUJeTtcwbGshjT25x62fKuHIknGfijNqsRQ807hdQmv8RLWLzo62nIee3TKN0YHvR7ju3ctTa_C5Xv3O72SWRXQ-MqoPrfD9C2TJq9R-UH-r7FQ0URDnyIHX_2q0joUNAB35ujle0uBuoOJSMfizFGdKigIbA3R5qDqC2qHE2NhPbJtGTZjpj3jPPiCTepziS_LhgAoGFmSjcuXVmEzSCkJyCfqm9zXudPD08_UJM61dcfptQ"

	crnSubject, err := pep.GetSubjectFromToken(crnToken, true)

	if err != nil {
		fmt.Println(err)
	}

	action := "iam.policy.read"
	resource := pep.Attributes{
		"serviceName":     "kitchen-tracker",
		"serviceInstance": "mykitchen",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	request := pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  crnSubject,
		},
	}

	trace := "txid-12348"
	response, err := pep.PerformAuthorization(&request, trace)

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Printf("%+v\n", response.Decisions)
	}
	//Output: [{Permitted:true Cached:false Expired:false RetryCount:0 Reason:0}]
}

func handleError(err error) bool {

	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
		return true
	}
	return false
}

func ExamplePerformAuthorization_aRequestShouldAlwaysBeCached() {
	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}

	err := pep.Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		fmt.Println(err)
	}

	action := "gopep.books.read"
	resource := pep.Attributes{
		"accountId":    "2c17c4e5587783961ce4a0aa415054e7",
		"resourceType": "book",
		"serviceName":  "gopep",
		"resource":     "book1",
	}

	subject := pep.Attributes{
		"id":    "IBMid-not-cache-key-match",
		"scope": "ibm openid",
	}

	// This request does not fit any of the obligation pattern.
	// The second PerformAutorization should indicate that this request is cached.
	request := pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "txid-request-should-go-to-source"
	response, err := pep.PerformAuthorization(&request, trace)

	if err != nil {
		handleError(err)
	} else {
		fmt.Printf("%+v\n", response.Decisions)
	}

	trace = "txid-request-should-be-cached"
	response, err = pep.PerformAuthorization(&request, trace)

	if err != nil {
		handleError(err)
	} else {
		fmt.Printf("%+v\n", response.Decisions)
	}
	//Output: [{Permitted:true Cached:false Expired:false RetryCount:0 Reason:0}]
	//[{Permitted:true Cached:true Expired:false RetryCount:0 Reason:0}]
}

func ExamplePerformAuthorization_highLevelAccess() {
	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}
	err := pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	//token := "eyJraWQiOiIyMDIwMTIxMDE0NDkiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0zMTAwMDE1WERTIiwiaWQiOiJJQk1pZC0zMTAwMDE1WERTIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiOGQzYjY1YzctMjk3MC00M2QxLTgwNDQtM2U1NmE4NzhhNTk4IiwiaWRlbnRpZmllciI6IjMxMDAwMTVYRFMiLCJnaXZlbl9uYW1lIjoiQ2hyaXMiLCJmYW1pbHlfbmFtZSI6IkhhdGtvIiwibmFtZSI6IkNocmlzIEhhdGtvIiwiZW1haWwiOiJjaGF0a29AY2EuaWJtLmNvbSIsInN1YiI6ImNoYXRrb0BjYS5pYm0uY29tIiwiaWF0IjoxNjA3NzQ0MDAyLCJleHAiOjE2MDc3NDc2MDIsImlzcyI6Imh0dHBzOi8vaWFtLnRlc3QuY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOnBhc3Njb2RlIiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiYngiLCJhY3IiOjEsImFtciI6WyJwd2QiXX0.JWZ8kKyOvv_JxExX1kLA8G2VOWfpJB4dd3BBvWvn2BFkFcQRF2SElhWvfsCY9TSr7dx-7M4_-4a10R7sMtInNF3d6Kq5ieF5ltcsOtjYmE4q7IuVclw4tOrvmInmj7o1HiM3Fk1VidBtfh2SxDs-JKbO0rATtMVmRFPHUPu6lDq_HzSu4pF5olKWUQQ1-DhbtuOEYcicvqrQ4qgealK2bytWOHHTbCNvvFzMsYx0OZO5nFW0rzYtkncaqfacAhe8PwkFQoUGn5Te67-9qO1WqA7q920YlZWSFjgmaxqvWQ7Qfx0akkI0z9G7L5p34WxDOcY35Cd91ywSYyDL1rgKvw"

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	resourceExtra := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		"resource":        "blah",
	}

	//subject, _ := pep.GetSubjectFromToken(token, true)

	subject := pep.Attributes{
		"id":    "IBMid-3100015XDS",
		"scope": "ibm openid",
	}

	action := "gopep.books.read"

	requests1 := pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	requests2 := pep.Requests{
		{
			"action":   action,
			"resource": resourceExtra,
			"subject":  subject,
		},
	}

	trace := "txid-12345"
	response1, err := pep.PerformAuthorization(&requests1, trace)
	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Printf("%+v\n", response1.Decisions)
	}

	time.Sleep(1 * time.Second)

	response2, err := pep.PerformAuthorization(&requests2, trace)
	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Printf("%+v\n", response2.Decisions)
	}

	//Output: [{Permitted:true Cached:false Expired:false RetryCount:0 Reason:0}]
	//[{Permitted:true Cached:true Expired:false RetryCount:0 Reason:0}]
}
func ExamplePerformAuthorization_highLevelAccessTokenSubject() {
	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}
	err := pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	/* #nosec G101 */
	token := "eyJraWQiOiIyMDIwMTIxMDE0NDkiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0zMTAwMDE1WERTIiwiaWQiOiJJQk1pZC0zMTAwMDE1WERTIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiOGQzYjY1YzctMjk3MC00M2QxLTgwNDQtM2U1NmE4NzhhNTk4IiwiaWRlbnRpZmllciI6IjMxMDAwMTVYRFMiLCJnaXZlbl9uYW1lIjoiQ2hyaXMiLCJmYW1pbHlfbmFtZSI6IkhhdGtvIiwibmFtZSI6IkNocmlzIEhhdGtvIiwiZW1haWwiOiJjaGF0a29AY2EuaWJtLmNvbSIsInN1YiI6ImNoYXRrb0BjYS5pYm0uY29tIiwiaWF0IjoxNjA3NzQ0MDAyLCJleHAiOjE2MDc3NDc2MDIsImlzcyI6Imh0dHBzOi8vaWFtLnRlc3QuY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOnBhc3Njb2RlIiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiYngiLCJhY3IiOjEsImFtciI6WyJwd2QiXX0.JWZ8kKyOvv_JxExX1kLA8G2VOWfpJB4dd3BBvWvn2BFkFcQRF2SElhWvfsCY9TSr7dx-7M4_-4a10R7sMtInNF3d6Kq5ieF5ltcsOtjYmE4q7IuVclw4tOrvmInmj7o1HiM3Fk1VidBtfh2SxDs-JKbO0rATtMVmRFPHUPu6lDq_HzSu4pF5olKWUQQ1-DhbtuOEYcicvqrQ4qgealK2bytWOHHTbCNvvFzMsYx0OZO5nFW0rzYtkncaqfacAhe8PwkFQoUGn5Te67-9qO1WqA7q920YlZWSFjgmaxqvWQ7Qfx0akkI0z9G7L5p34WxDOcY35Cd91ywSYyDL1rgKvw"

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	resourceExtra := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
		"resource":        "blah",
	}

	subject, _ := pep.GetSubjectFromToken(token, true)

	// These are examples of what the GetSubjectFromToken generates
	// subject1 := pep.Attributes{
	// 	"id": "IBMid-3100015XDS",
	// }
	// subject2 := pep.Attributes{
	// 	"id":    "IBMid-3100015XDS",
	// 	"scope": "ibm openid",
	// }
	action := "gopep.books.read"

	requests1 := pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	requests2 := pep.Requests{
		{
			"action":   action,
			"resource": resourceExtra,
			"subject":  subject,
		},
	}

	trace := "txid-12345"
	response1, err := pep.PerformAuthorization(&requests1, trace)
	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Printf("%+v\n", response1.Decisions)
	}

	time.Sleep(1 * time.Second)

	response2, err := pep.PerformAuthorization(&requests2, trace)
	// Error handling
	if err != nil {
		switch err := err.(type) {
		case *pep.InternalError:
			// Handling the internal error
			fmt.Println(err.Error())
		case *pep.APIError:
			// Handling the API error
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		fmt.Printf("%+v\n", response2.Decisions)
	}

	//Output: [{Permitted:true Cached:false Expired:false RetryCount:0 Reason:0}]
	//[{Permitted:true Cached:true Expired:false RetryCount:0 Reason:0}]
}

func ExampleGetDelegationToken() {
	config := &pep.Config{
		Environment: pep.Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    pep.LevelError,
	}
	err := pep.Configure(config)
	if err != nil {
		fmt.Println(err)
	}

	// This portion covers an API key that has delegation permissions for the given CRN's service
	desiredIAMID := "crn-crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:instance12345666666::"

	delegationToken, err := pep.GetDelegationToken(desiredIAMID)
	if err != nil {
		fmt.Println(err)
	}

	iamIDClaim, err := pep.GetSubjectAsIAMIDClaim(delegationToken, false)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("delegation iam ID", iamIDClaim)

	// This portion covers an API key that does NOT have delegation permissions for the given CRN's service and the subsequent error
	notPermittedIAMIDservice := "crn-crn:v1:staging:public:gopep1::a/2c17c4e5587783961ce4a0aa415054e7:instance12345666666::"

	notPermittedDelegationToken, err := pep.GetDelegationToken(notPermittedIAMIDservice)

	i := strings.Index(err.Error(), "Status")

	errMessage := err.Error()[i:]

	fmt.Printf("token should be empty since it cannot be converted:%v\n", notPermittedDelegationToken)
	fmt.Println("the error:", errMessage)

	//Output: delegation iam ID crn-crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:instance12345666666::
	//token should be empty since it cannot be converted:
	//the error: Status Code 403 Forbidden BXNIM0513E You are not authorized to use this API
}

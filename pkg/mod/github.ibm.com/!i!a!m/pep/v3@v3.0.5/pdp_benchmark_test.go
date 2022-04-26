package pep

import (
	"fmt"
	"os"
	"testing"
)

func BenchmarkCachedPermittedAuthorization(b *testing.B) {

	config := &Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
	}

	err := Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	resource := Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	request := Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "benchmark0001"

	token, err := GetToken()
	if err != nil {
		b.Errorf("error: %+v", err)
	}

	_, err = request.isAuthorized(trace, token)
	if err != nil {
		b.Errorf("error: %+v", err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		token, err := GetToken()
		if err != nil {
			b.Errorf("error: %+v", err)
		}
		_, err = request.isAuthorized(trace, token)
		if err != nil {
			b.Errorf("error: %+v", err)
		}
	}

}

func BenchmarkCachedDeniedAuthorization(b *testing.B) {

	config := &Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
	}

	err := Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	resource := Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "3c17c4e5587783961ce4a0aa415054e7",
	}

	subject := Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	request := Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	trace := "benchmark0001"
	token, err := GetToken()
	if err != nil {
		b.Errorf("error: %+v", err)
	}
	_, err = request.isAuthorized(trace, token)
	if err != nil {
		b.Errorf("error: %+v", err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		token, err := GetToken()
		if err != nil {
			b.Errorf("error: %+v", err)
		}
		_, err = request.isAuthorized(trace, token)
		if err != nil {
			b.Errorf("error: %+v", err)
		}
	}

}

func BenchmarkCachedLargeRequest(b *testing.B) {

	config := &Config{
		Environment: Staging,
		APIKey:      os.Getenv("API_KEY"),
		LogLevel:    LevelError,
	}

	err := Configure(config)

	if err != nil {
		fmt.Println(err)
	}

	largeRequest, _ := generateBulkRequests(1000, false, false)

	trace := "benchmark-large-request-0001"
	token, err := GetToken()
	if err != nil {
		b.Errorf("error: %+v", err)
	}
	_, err = largeRequest.isAuthorized(trace, token)
	if err != nil {
		b.Errorf("error: %+v", err)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		token, err := GetToken()
		if err != nil {
			b.Errorf("error: %+v", err)
		}
		_, err = largeRequest.isAuthorized(trace, token)
		if err != nil {
			b.Errorf("error: %+v", err)
		}
	}

}

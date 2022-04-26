package elasticsearch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type indexInput struct {
	Test string
}

type searchInput struct {
	Test string
}

// mockTransport is a test HTTP transport used for unit testing
type mockTransport struct {
	RoundTripImpl func(req *http.Request) (*http.Response, error)
}

// RoundTrip is used for mocking HTTP calls
func (transport *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport.RoundTripImpl(req)
}

// TestElasticsearchNew tests that a instance of the Elasticsearch interface can be created
func TestElasticsearchNew(t *testing.T) {
	createElasticSearchInstance(t)
}

// TestElasticsearchNewWithConfig tests that an instance of the Elasticsearch interface can be created with a config
func TestElasticsearchNewWithConfig(t *testing.T) {
	loadEnvironment()
	config := Config{
		BulkIndexerDefaultIndex:      "fake",
		BulkIndexerFlushIntervalSecs: 30,
		BulkIndexerNumWorkers:        2,
	}
	elasticSearch, err := NewWithConfig(config)
	AssertError(t, err)
	if elasticSearch == nil {
		t.Error("Elasticsearch is nil")
	}
}

// TestElasticsearchNewWithTransport tests that a instance of the Elasticsearch interface with a mock HTTP transport
func TestElasticsearchNewWithTransport(t *testing.T) {
	createElasticSearchInstanceTransport(t, http.DefaultTransport)
}

// TestElasticsearchNewWithoutEnvLoaded tests whether an Elasticsearch instance can be created when the
// environment has not been loaded
func TestElasticsearchNewWithoutEnvLoaded(t *testing.T) {
	os.Setenv("SECRETCREDENTIALS_es_key", ``)
	_, err := New()
	AssertEqual(t, "error", "unexpected end of JSON input", err.Error())
}

// TestElasticsearchNewWithEnvLoadedFailure tests when the environment has been loaded but Elasticsearch is disabled
func TestElasticsearchNewWithEnvLoadedDisabled(t *testing.T) {
	os.Setenv("SECRETCREDENTIALS_es_key", `{"enabled":"false","url":"http://localhost:8080","user1":"fred1@ibm.com","password1":"abc123","user2":"fred2@ibm.com","password2":"def456"}`)

	_, err := New()
	AssertEqual(t, "error", "elasticsearch is disabled", err.Error())
}

// TestElasticsearchNewWithEnvLoadedFailure tests when the environment has been loaded but credentials are missing
func TestElasticsearchNewWithEnvLoadedInvalid(t *testing.T) {
	os.Setenv("SECRETCREDENTIALS_es_key", `{"enabled":"true","url":"http://localhost:8080","user1":"fred1@ibm.com","password1":"","user2":"fred2@ibm.com","password2":""}`)

	_, err := New()
	AssertEqual(t, "error", "missing elasticsearch credentials", err.Error())
}

// TestElasticsearchValidateCredentials tests that credentials are not loaded successfully if environment is not
// initialized
func TestElasticsearchValidateCredentials(t *testing.T) {
	var sc SecretCredentials

	ok := validateCredentials(sc)
	AssertEqual(t, "validateCredentials", false, ok)
}

// TestElasticsearchLoadCredentialsSuccess tests that credentials are loaded successfully
func TestElasticsearchLoadCredentialsSuccess(t *testing.T) {
	loadEnvironment()
	var sc SecretCredentials
	json.Unmarshal([]byte(os.Getenv("SECRETCREDENTIALS_es_key")), &sc)

	credentials := loadCredentials(sc)
	AssertEqual(t, "user1", "fred1@ibm.com", credentials[1].user)
	AssertEqual(t, "password1", "abc123", credentials[1].password)
	AssertEqual(t, "user2", "fred2@ibm.com", credentials[2].user)
	AssertEqual(t, "password2", "def456", credentials[2].password)
}

// TestElasticsearchCreateElasticSearchClients tests the creation of the Elasticsearch clients
func TestElasticsearchCreateElasticSearchClients(t *testing.T) {
	loadEnvironment()
	var sc SecretCredentials
	json.Unmarshal([]byte(os.Getenv("SECRETCREDENTIALS_es_key")), &sc)
	credentials := loadCredentials(sc)
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			// Return response and nil error:
			return &http.Response{
				Status:     "201 Created",
				StatusCode: http.StatusCreated,
				Body:       ioutil.NopCloser(strings.NewReader("test")),
			}, nil
		},
	}
	clients, err := createElasticSearchClients(transport, "http://localhost:8080", credentials)
	AssertError(t, err)
	AssertEqual(t, "len of clients", 2, len(clients))
}

// TestElasticsearchCreateElasticSearchBulkIndexers tests that the createElasticSearchInstanceTransport function
// returns 2 bulk indexers
func TestElasticsearchCreateElasticSearchBulkIndexers(t *testing.T) {
	es := createElasticSearchInstanceTransport(t, http.DefaultTransport)
	esStruct := es.(*elasticsearch)
	bulkIndexers, err := esStruct.createElasticSearchBulkIndexers()
	AssertError(t, err)
	AssertEqual(t, "len of bulk indexers", 2, len(bulkIndexers))
}

// isUnauthorizedError tests the isUnauthorizedError function
func TestIsUnauthorizedError(t *testing.T) {
	err := errors.New("foo")
	isAuth := isUnauthorizedError(context.Background(), err)
	AssertEqual(t, "first check", false, isAuth)

	err = errors.New("foo 401 Unauthorized bar")
	isAuth = isUnauthorizedError(context.Background(), err)
	AssertEqual(t, "second check", true, isAuth)

	isAuth = isUnauthorizedError(context.Background(), nil)
	AssertEqual(t, "third check", false, isAuth)
}

// TestElasticsearchSearchSuccess tests a successful search from Elasticsearch
func TestElasticsearchSearchSuccess(t *testing.T) {
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			// Header check: https://github.com/elastic/go-elasticsearch/pull/324/files
			if req.Method == http.MethodGet {
				// Return response with required header in 7.15.x and nil error:
				return &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("{}")),
					Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				}, nil
			}

			// Verify operation after header check:
			AssertEqual(t, "operation", http.MethodPost, req.Method)
			// Verify fullURL:
			AssertEqual(t, "url", "http://localhost:8080/fake/_search?track_total_hits=true", req.URL.String())
			// Verify authorization:
			authorization := req.Header.Get("Authorization")
			authPrefix := "Basic "
			if !strings.HasPrefix(authorization, authPrefix) {
				t.Errorf("Authorization does not start with '%s'. Auth=%s", authPrefix, authorization)
			}
			decodedAuth, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authorization, authPrefix))
			AssertError(t, err)
			AssertEqual(t, "authorization", "fred1@ibm.com:abc123", string(decodedAuth))
			// Create return object:
			response := SearchResponse{}
			hit := SearchResponseHit{}
			hit.ID = "id1"
			var hits [1]SearchResponseHit
			hits[0] = hit
			response.Hits.Hits = hits[0:]
			responseBody, err := toJSON(response)
			// Return response and nil error:
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
			}, nil
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := searchInput{Test: "TestElasticsearchSearchSuccess"}
	response, err := elasticSearch.Search(input, "fake")
	AssertError(t, err)
	AssertEqual(t, "hits", 1, len(response.Hits.Hits))
}

// TestElasticsearchSearchSuccessInputString tests a successful search from Elasticsearch
func TestElasticsearchSearchSuccessInputString(t *testing.T) {
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			// Header check: https://github.com/elastic/go-elasticsearch/pull/324/files
			if req.Method == http.MethodGet {
				// Return response with required header in 7.15.x and nil error:
				return &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("{}")),
					Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				}, nil
			}

			// Verify operation after header check:
			AssertEqual(t, "operation", http.MethodPost, req.Method)
			// Verify fullURL:
			AssertEqual(t, "url", "http://localhost:8080/fake/_search?track_total_hits=true", req.URL.String())
			// Verify authorization:
			authorization := req.Header.Get("Authorization")
			authPrefix := "Basic "
			if !strings.HasPrefix(authorization, authPrefix) {
				t.Errorf("Authorization does not start with '%s'. Auth=%s", authPrefix, authorization)
			}
			decodedAuth, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authorization, authPrefix))
			AssertError(t, err)
			AssertEqual(t, "authorization", "fred1@ibm.com:abc123", string(decodedAuth))
			// Create return object:
			response := SearchResponse{}
			hit := SearchResponseHit{}
			hit.ID = "id1"
			var hits [1]SearchResponseHit
			hits[0] = hit
			response.Hits.Hits = hits[0:]
			responseBody, err := toJSON(response)
			// Return response and nil error:
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
			}, nil
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := "{\"test\":\"TestElasticsearchSearchSuccess\"}"
	response, err := elasticSearch.Search(input, "fake")
	AssertError(t, err)
	AssertEqual(t, "hits", 1, len(response.Hits.Hits))
}

// TestElasticsearchSearchError tests the case when the search request to Elasticsearch always returns a non-401 error
func TestElasticsearchSearchError(t *testing.T) {
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			// Return error:
			return nil, errors.New("Something bad happened")
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := searchInput{Test: "TestElasticsearchSearchError"}
	_, err := elasticSearch.Search(input, "fake")
	if err == nil {
		t.Errorf("Expected that Elasticsearch index call would return error but did not receive one")
	}
}

// TestElasticsearchSearch401Success tests the case that the first search returns a 401 error, and the second call
// returns success
func TestElasticsearchSearch401Success(t *testing.T) {
	callCounter := 0
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			callCounter++
			if callCounter == 1 {
				return &http.Response{
					Status:     "401 Unauthorized",
					StatusCode: http.StatusUnauthorized,
					Body:       ioutil.NopCloser(strings.NewReader("Failure")),
					Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				}, nil
			}
			// Return response and nil error:
			response := SearchResponse{}
			responseBody, err := toJSON(response)
			AssertError(t, err)
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
			}, nil
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := searchInput{Test: "TestElasticsearchSearch401Success"}
	_, err := elasticSearch.Search(input, "fake")
	AssertError(t, err)
	// Elasticsearch Search API should have been called twice (first failed, second succeeded):
	AssertEqual(t, "counter", "2", strconv.Itoa(callCounter))
}

// TestElasticsearchSearch401Error tests the case that the first and second search calls return 401 errors
func TestElasticsearchSearch401Error(t *testing.T) {
	callCounter := 0
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			callCounter++
			return &http.Response{
				Status:     "401 Unauthorized",
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(strings.NewReader("Failure")),
				Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
			}, nil
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := searchInput{Test: "TestElasticsearchSearch401Error"}
	_, err := elasticSearch.Search(input, "fake")
	if err == nil {
		t.Errorf("Expected that Elasticsearch search would return error but did not receive one")
	}
	// Elasticsearch Search API should have been called four times (both failed, a header check call and an actual call each):
	AssertEqual(t, "counter", "4", strconv.Itoa(callCounter))
}

// TestElasticsearchIndexSuccess tests a successful index to Elasticsearch
func TestElasticsearchIndexSuccess(t *testing.T) {
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			// Header check: https://github.com/elastic/go-elasticsearch/pull/324/files
			if req.Method == http.MethodGet {
				// Return response with required header in 7.15.x and nil error:
				return &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("{}")),
					Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				}, nil
			}

			// Verify operation after header check:
			AssertEqual(t, "operation", http.MethodPost, req.Method)
			// Verify fullURL:
			AssertEqual(t, "url", "http://localhost:8080/fake/_doc", req.URL.String())
			// Verify authorization:
			authorization := req.Header.Get("Authorization")
			authPrefix := "Basic "
			if !strings.HasPrefix(authorization, authPrefix) {
				t.Errorf("Authorization does not start with '%s'. Auth=%s", authPrefix, authorization)
			}
			decodedAuth, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authorization, authPrefix))
			AssertError(t, err)
			AssertEqual(t, "authorization", "fred1@ibm.com:abc123", string(decodedAuth))
			// Create return object:
			indexResponse := IndexResponse{}
			indexResponse.ID = "id1"
			indexResponse.Index = "index1"
			indexResponse.PrimaryTerm = 2
			indexResponse.SeqNo = 3
			indexResponse.Shard = make(map[string]interface{})
			indexResponse.Shard["failed"] = 4
			indexResponse.Shard["successful"] = 5
			indexResponse.Shard["total"] = 6
			indexResponse.Type = "type1"
			indexResponse.Version = 7
			indexResponse.Result = "result1"
			responseBody, err := toJSON(indexResponse)
			AssertError(t, err)
			// Return response and nil error:
			return &http.Response{
				Status:     "201 Created",
				StatusCode: http.StatusCreated,
				Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
			}, nil
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := indexInput{Test: "TestElasticsearchIndexSuccess"}
	response, err := elasticSearch.Index(input, "fake")
	AssertError(t, err)
	AssertEqual(t, "Id", "id1", response.ID)
	AssertEqual(t, "Index", "index1", response.Index)
	AssertEqual(t, "PrimaryTerm", 2, response.PrimaryTerm)
	AssertEqual(t, "SeqNo", 3, response.SeqNo)
	AssertEqual(t, "Shard.failed", float64(4), response.Shard["failed"])
	AssertEqual(t, "Shard.successful", float64(5), response.Shard["successful"])
	AssertEqual(t, "Shard.total", float64(6), response.Shard["total"])
	AssertEqual(t, "Type", "type1", response.Type)
	AssertEqual(t, "Version", 7, response.Version)
	AssertEqual(t, "Result", "result1", response.Result)
}

// TestElasticsearchIndexError tests the case when the index to Elasticsearch always returns a non-401 error
func TestElasticsearchIndexError(t *testing.T) {
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			// Return error:
			return nil, errors.New("Something bad happened")
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := indexInput{Test: "TestElasticsearchIndexError"}
	_, err := elasticSearch.Index(input, "fake")
	if err == nil {
		t.Errorf("Expected that Elasticsearch index call would return error but did not receive one")
	}
}

// TestElasticsearchIndex401Success tests the case that the first index returns a 401 error, and the second call
// returns success
func TestElasticsearchIndex401Success(t *testing.T) {
	callCounter := 0
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			callCounter++
			if callCounter == 1 {
				return &http.Response{
					Status:     "401 Unauthorized",
					StatusCode: http.StatusUnauthorized,
					Body:       ioutil.NopCloser(strings.NewReader("Failure")),
				}, nil
			}
			// Return response and nil error:
			indexResponse := IndexResponse{}
			responseBody, err := toJSON(indexResponse)
			AssertError(t, err)
			return &http.Response{
				Status:     "201 Created",
				StatusCode: http.StatusCreated,
				Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
			}, nil
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := indexInput{Test: "TestElasticsearchIndex401Success"}
	_, err := elasticSearch.Index(input, "fake")
	AssertError(t, err)
	// Elasticsearch Index API should have been called twice (first failed, second succeeded):
	AssertEqual(t, "counter", "2", strconv.Itoa(callCounter))
}

// TestElasticsearchIndex401Error tests the case that the first and second index calls return 401 errors
func TestElasticsearchIndex401Error(t *testing.T) {
	callCounter := 0
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			callCounter++
			return &http.Response{
				Status:     "401 Unauthorized",
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(strings.NewReader("Failure")),
			}, nil
		},
	}
	elasticSearch := createElasticSearchInstanceTransport(t, transport)
	input := indexInput{Test: "TestElasticsearchIndex401Error"}
	_, err := elasticSearch.Index(input, "fake")
	if err == nil {
		t.Errorf("Expected that Elasticsearch index would return error but did not receive one")
	}
	// Elasticsearch Index API should have been called four times (both failed, a header check call and an actual call each):
	AssertEqual(t, "counter", "4", strconv.Itoa(callCounter))
}

// TestElasticsearchBulkIndexSuccess tests the BulkIndex function. In particular, it tests that multiple
// calls to the BulkIndex function result in a single call to Elasticsearch after the flush
// interval expires, and that the bulk indexer can be used again.
func TestElasticsearchBulkIndexSuccess(t *testing.T) {
	roundTripImplCounter := 0
	transport := &mockTransport{
		RoundTripImpl: func(req *http.Request) (*http.Response, error) {
			roundTripImplCounter++
			// Header check: https://github.com/elastic/go-elasticsearch/pull/324/files
			if req.Method == http.MethodGet {
				// Return response with required header in 7.15.x and nil error:
				return &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("{}")),
					Header:     http.Header{"X-Elastic-Product": []string{"Elasticsearch"}},
				}, nil
			}

			// Verify operation after header check:
			AssertEqual(t, "operation", http.MethodPost, req.Method)
			// Verify fullURL:
			AssertEqual(t, "url", "http://localhost:8080/fake/_bulk", req.URL.String())
			// Verify authorization:
			authorization := req.Header.Get("Authorization")
			authPrefix := "Basic "
			if !strings.HasPrefix(authorization, authPrefix) {
				t.Errorf("Authorization does not start with '%s'. Auth=%s", authPrefix, authorization)
			}
			decodedAuth, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authorization, authPrefix))
			AssertError(t, err)
			AssertEqual(t, "authorization", "fred1@ibm.com:abc123", string(decodedAuth))
			// Create return object:
			indexResponse := IndexResponse{}
			responseBody, err := toJSON(indexResponse)
			AssertError(t, err)
			// Return response and nil error:
			return &http.Response{
				Status:     "201 Created",
				StatusCode: http.StatusCreated,
				Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
			}, nil
		},
	}

	// Create Elasticsearch instance using mock transport from above and the following config:
	config := Config{
		BulkIndexerDefaultIndex:      "fake",
		BulkIndexerFlushIntervalSecs: 1,
		BulkIndexerNumWorkers:        1,
	}
	elasticSearch := createElasticSearchInstanceTransportAndConfig(t, transport, config)

	// Call BulkIndex:
	input := indexInput{Test: "TestElasticsearchBulkIndexSuccess"}
	err := elasticSearch.BulkIndex(input, "fake", "TestElasticsearchBulkIndexSuccess")
	AssertError(t, err)
	AssertEqual(t, "counter after initial BulkIndex call", 0, roundTripImplCounter)

	// Call BulkIndex again:
	err = elasticSearch.BulkIndex(input, "fake", "TestElasticsearchBulkIndexSuccess")
	AssertError(t, err)
	AssertEqual(t, "counter after second BulkIndex call", 0, roundTripImplCounter)

	// Sleep to give bulk indexer time to flush:
	time.Sleep(2 * time.Second)

	AssertEqual(t, "counter after first flush", 2, roundTripImplCounter) // one header check and one actual call

	// Call BulkIndex again:
	err = elasticSearch.BulkIndex(input, "fake", "TestElasticsearchBulkIndexSuccess")
	AssertError(t, err)
	AssertEqual(t, "counter after third BulkIndex call", 2, roundTripImplCounter)

	// Sleep to give bulk indexer time to flush again:
	time.Sleep(2 * time.Second)

	AssertEqual(t, "counter after second flush", 3, roundTripImplCounter)
}

// TestBulkIndexOnErrorHandler tests the bulkIndexOnErrorHandler function to ensure that the current client switching works correctly
func TestBulkIndexOnErrorHandler(t *testing.T) {
	// Create an instance of the Elasticsearch interface:
	elasticSearch := createElasticSearchInstance(t)

	// Cast to the elasticsearch struct to get access to information and functions not exposed by the Elasticsearch interface:
	es, ok := elasticSearch.(*elasticsearch)
	if !ok {
		t.Error("Cast from Elasticsearch interface to struct failed")
	}

	// Initially the current client should be 1:
	AssertEqual(t, "current client", 1, es.currentClient) // first client

	// Call the error handler with the first client being the client that encountering an error:
	err := errors.New("401 Unauthorized")
	es.bulkIndexOnErrorHandler(context.Background(), err, 1) // first client
	AssertEqual(t, "current client", 2, es.currentClient)    // expect a switch to the second client

	// Call the error handler again with the first client being the client that encountering an error:
	err = errors.New("401 Unauthorized")
	es.bulkIndexOnErrorHandler(context.Background(), err, 1) // first client
	AssertEqual(t, "current client", 2, es.currentClient)    // no switch expected

	// Call the error handler with the second client being the client that encountering an error:
	err = errors.New("401 Unauthorized")
	es.bulkIndexOnErrorHandler(context.Background(), err, 2) // second client
	AssertEqual(t, "current client", 1, es.currentClient)    // expect a switch to the first client

	// Call the error handler again with the second client being the client that encountering an error:
	err = errors.New("401 Unauthorized")
	es.bulkIndexOnErrorHandler(context.Background(), err, 2) // second client
	AssertEqual(t, "current client", 1, es.currentClient)    // no switch expected

	// Call the error handler with the first client being the client that encountering an error:
	err = errors.New("401 Unauthorized")
	es.bulkIndexOnErrorHandler(context.Background(), err, 1) // first client
	AssertEqual(t, "current client", 2, es.currentClient)    // expect a switch to the second client

	// Call the error handler with the first client being the client that encountering an error, but with a non-401 error:
	err = errors.New("Something bad happened")
	es.bulkIndexOnErrorHandler(context.Background(), err, 1) // first client
	AssertEqual(t, "current client", 2, es.currentClient)    // no switch expected because non-401 error encountered

	// Call the error handler with the second client being the client that encountering an error, but with a non-401 error:
	err = errors.New("Something bad happened")
	es.bulkIndexOnErrorHandler(context.Background(), err, 2) // second client
	AssertEqual(t, "current client", 2, es.currentClient)    // no switch expected because non-401 error encountered
}

// loadEnvironment is a helper function that loads the environment with needed variables
func loadEnvironment() {
	// Set and load environment variables (all values below are fake):
	os.Setenv("SECRETCREDENTIALS_es_key", `{"enabled":"true","url":"http://localhost:8080","user1":"fred1@ibm.com","password1":"abc123","user2":"fred2@ibm.com","password2":"def456"}`)
}

// createElasticSearchInstance is a helper function that creates an instance of Elasticsearch
func createElasticSearchInstance(t *testing.T) Elasticsearch {
	loadEnvironment()
	elasticSearch, err := New()
	AssertError(t, err)
	if elasticSearch == nil {
		t.Error("Elasticsearch is nil")
	}
	return elasticSearch
}

// createElasticSearchInstanceTransport is a helper function that creates an instance of Elasticsearch with an
// http client
func createElasticSearchInstanceTransport(t *testing.T, transport http.RoundTripper) Elasticsearch {
	loadEnvironment()
	config := Config{
		BulkIndexerDefaultIndex:      "fake",
		BulkIndexerFlushIntervalSecs: 10,
		BulkIndexerNumWorkers:        1,
	}
	elasticSearch, err := newWithTransport(config, transport)
	AssertError(t, err)
	if elasticSearch == nil {
		t.Error("Elasticsearch is nil")
	}
	return elasticSearch
}

// createElasticSearchInstanceTransportAndConfig is a helper function that creates an instance of Elasticsearch
// with an http client and config
func createElasticSearchInstanceTransportAndConfig(t *testing.T, transport http.RoundTripper, config Config) Elasticsearch {
	loadEnvironment()
	elasticSearch, err := newWithTransport(config, transport)
	AssertError(t, err)
	if elasticSearch == nil {
		t.Error("Elasticsearch is nil")
	}
	return elasticSearch
}

// toJSON returns JSON representation of the provided interface
func toJSON(j interface{}) (string, error) {
	result := ""
	var out []byte
	var err error
	out, err = json.MarshalIndent(j, "", "    ")
	if err == nil {
		result = string(out)
	}
	return result, err
}

// AssertError reports a testing failure when a given error is non nil
func AssertError(t *testing.T, err error) {
	if err != nil {
		t.Logf("%s operation failed: %v", t.Name(), err)
		t.Fail()
	}
}

// AssertEqual reports a testing failure when a given actual item (interface{}) is not equal to the expected value
func AssertEqual(t *testing.T, item string, expected, actual interface{}) {
	if actual == nil {
		if expected != nil {
			t.Logf("%s %s mismatch - expected %T %v got %T %v", t.Name(), item, expected, expected, actual, actual)
			t.Fail()
		}
	} else if reflect.TypeOf(actual).Comparable() {
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

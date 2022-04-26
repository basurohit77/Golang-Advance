package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/oss-secrets/secret"

	esc "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

// Provides the ability to call an Elasticsearch instance.
//
// Call New() to create an instance.
//
// The following environment variable, or similar configuration in the environment, is needed to successfully create
// an instance:
//
// SECRETCREDENTIALS_es_key=
// '{"enabled":"true","url":"http://localhost:8080","user1":"fred@ibm.com","password1":"abc123)","user2":"fred@ibm.com","password2":"def456"}'
//
// Note that:
// - The enabled property determines whether Elasticsearch should be enabled in this environment
// - The url property is the Elasticsearch URL (for example,
//   tip-elasticsearch-elasticsearch-client.tip.svc.cluster.local:9200)
// - Two users need to be provided that can be used to authenticate with Elasticsearch. The Elasticsearch code will
//   try both. This allows easier rotation of passwords.

// Elasticsearch defines the functions supported
type Elasticsearch interface {
	BulkIndex(input interface{}, index string, id string) error
	Index(input interface{}, index string) (IndexResponse, error)
	Search(input interface{}, index string) (SearchResponse, error)
}

// IndexResponse is the response from calling the Elasticsearch Index API
// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html
type IndexResponse struct {
	ID          string `json:"_id"`
	Index       string `json:"_index"`
	PrimaryTerm int    `json:"_primary_term"`
	SeqNo       int    `json:"_seq_no"`
	// The '_shards.failed' field is dynamic, so using an interface map:
	Shard   map[string]interface{} `json:"_shards"`
	Type    string                 `json:"_type"`
	Version int                    `json:"_version"`
	Result  string                 `json:"result"`
}

// SearchResponse is the response from calling the Elasticsearch Search API
type SearchResponse struct {
	Hits struct {
		Hits []SearchResponseHit `json:"hits"`
	} `json:"hits"`
	Aggregations interface{} `json:"aggregations"`
}

// SearchResponseHit represents a single hit in the response from calling the Elasticsearch Search API
type SearchResponseHit struct {
	ID     string                 `json:"_id"`
	Index  string                 `json:"_index"`
	Source map[string]interface{} `json:"_source"`
}

// elasticsearch is the internal implementation of the Elasticsearch interface
type elasticsearch struct {
	url                string
	clients            map[int]*esc.Client
	bulkIndexers       map[int]*esutil.BulkIndexer
	currentClient      int
	currentClientMutex sync.RWMutex
	config             Config
}

// esCredentials represents an individual Elasticsearch credential
type esCredentials struct {
	user     string
	password string
}

// SecretCredentials is the environment variable stored in vault
type SecretCredentials struct {
	Enabled   string
	Url       string
	User1     string
	Password1 string
	User2     string
	Password2 string
}

// Config is the configuration used to create an instance of Elasticsearch
type Config struct {
	BulkIndexerDefaultIndex      string // Default index name used by the bulk indexer
	BulkIndexerFlushIntervalSecs int    // Time in seconds until the bulk indexer will flush documents to Elasticsearch
	BulkIndexerNumWorkers        int    // Number of workers that will perform the flush to Elasticsearch
}

// New creates a new instance with the default configuration
func New() (Elasticsearch, error) {
	config := Config{BulkIndexerDefaultIndex: "breakglass", BulkIndexerFlushIntervalSecs: 300, BulkIndexerNumWorkers: 1}
	return NewWithConfig(config)
}

// NewWithConfig creates a new instance with the provided configuration
func NewWithConfig(config Config) (Elasticsearch, error) {
	return newWithTransport(config, http.DefaultTransport)
}

// newWithTransport is an internal function that creates a new instance and uses the provided HTTP transport. This injection of the
// transport allows for easier unit testing.
func newWithTransport(config Config, transport http.RoundTripper) (Elasticsearch, error) {
	// SECRETCREDENTIALS_es_key = `{"Enabled":"true","Url":"http://tip-elasticsearch-elasticsearch-client.tip.svc.cluster.local:9200",
	// "User1":"xxxxxxxx@ibm.com","Password1":"xxxxxx","User2":"xxxxxxxx@ibm.com","Password2":"xxxxxx"}`

	// First get the SecretCredentials object (Elasticsearch enable boolean, url and credentials) from the environment:
	var sc SecretCredentials
	err := json.Unmarshal([]byte(secret.Get("SECRETCREDENTIALS_es_key")), &sc)
	if err != nil {
		return nil, err
	}

	// Check if all the credentials are initialized:
	if !validateCredentials(sc) {
		return nil, errors.New("missing elasticsearch credentials")
	}

	// Second check if Elasticsearch is enabled:
	if sc.Enabled != "true" {
		return nil, errors.New("elasticsearch is disabled")
	}

	// Third create the Elasticsearch clients with url and credentials:
	clients, err := createElasticSearchClients(transport, sc.Url, loadCredentials(sc))
	if err != nil {
		return nil, err
	}

	// Forth create an instance of Elasticsearch:
	es := &elasticsearch{
		url:           sc.Url,
		clients:       clients,
		currentClient: 1, // try the first user name and password first
		config:        config,
	}

	// Fifth create the Elasticsearch bulk indexers and sets in the instance of Elasticsearch:
	bulkIndexers, err := es.createElasticSearchBulkIndexers()
	if err != nil {
		return nil, err
	}
	es.bulkIndexers = bulkIndexers

	return es, nil
}

// validateCredentials checks for uninitialized credentials
func validateCredentials(sc SecretCredentials) bool {
	return sc.Enabled != "" && sc.Url != "" && sc.User1 != "" && sc.User2 != "" && sc.Password1 != "" && sc.Password2 != ""
}

// loadCredentials loads the user and password information from the environment
func loadCredentials(sc SecretCredentials) map[int]esCredentials {
	credentials := make(map[int]esCredentials)
	credentials[1] = esCredentials{user: sc.User1, password: sc.Password1}
	credentials[2] = esCredentials{user: sc.User2, password: sc.Password2}
	return credentials
}

// createElasticSearchClients creates Elasticsearch clients given the provided HTTP transport, url, and credentials
func createElasticSearchClients(transport http.RoundTripper, url string, credentials map[int]esCredentials) (clients map[int]*esc.Client, err error) {
	clients = make(map[int]*esc.Client)
	config1 := esc.Config{
		Transport: transport,
		Addresses: []string{
			url,
		},
		Username: credentials[1].user,
		Password: credentials[1].password,
	}
	var client1 *esc.Client
	client1, err = esc.NewClient(config1)
	if err != nil {
		return
	}
	clients[1] = client1

	config2 := esc.Config{
		Transport: transport,
		Addresses: []string{
			url,
		},
		Username: credentials[2].user,
		Password: credentials[2].password,
	}
	var client2 *esc.Client
	client2, err = esc.NewClient(config2)
	if err != nil {
		return
	}
	clients[2] = client2

	return
}

// createElasticSearchBulkIndexers creates the Elasticsearch bulk indexers
func (es *elasticsearch) createElasticSearchBulkIndexers() (bulkIndexers map[int]*esutil.BulkIndexer, err error) {
	bulkIndexers = make(map[int]*esutil.BulkIndexer)

	var bulkIndexer1 esutil.BulkIndexer
	bulkIndexer1, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         es.config.BulkIndexerDefaultIndex,
		Client:        es.clients[1],
		FlushInterval: time.Duration(es.config.BulkIndexerFlushIntervalSecs) * time.Second,
		NumWorkers:    es.config.BulkIndexerNumWorkers,
		OnError:       func(ctx context.Context, err error) { es.bulkIndexOnErrorHandler(ctx, err, 1) },
	})
	bulkIndexers[1] = &bulkIndexer1

	var bulkIndexer2 esutil.BulkIndexer
	bulkIndexer2, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         es.config.BulkIndexerDefaultIndex,
		Client:        es.clients[2],
		FlushInterval: time.Duration(es.config.BulkIndexerFlushIntervalSecs) * time.Second,
		NumWorkers:    es.config.BulkIndexerNumWorkers,
		OnError:       func(ctx context.Context, err error) { es.bulkIndexOnErrorHandler(ctx, err, 2) },
	})
	bulkIndexers[2] = &bulkIndexer2

	return bulkIndexers, err
}

// bulkIndexOnErrorHandler handles an onError failure for the bulk indexer clients. The current client may be
// switched in the case of a 401 error.
func (es *elasticsearch) bulkIndexOnErrorHandler(ctx context.Context, err error, clientWithError int) {
	const fct = "bulkIndexOnErrorHandler: "
	if isUnauthorizedError(ctx, err) {
		// Switch the current client if necessary:
		es.currentClientMutex.Lock()
		defer es.currentClientMutex.Unlock()
		clientToSwitchTo := es.currentClient
		// Only switch clients if the current client is the same as the client with the error. This avoid
		// unnecessary flipping back and forth in the case that both clients become active at the same time:
		if es.currentClient == 1 && clientWithError == 1 {
			clientToSwitchTo = 2
		} else if es.currentClient == 2 && clientWithError == 2 {
			clientToSwitchTo = 1
		}
		if es.currentClient != clientToSwitchTo {
			// Need to switch clients
			log.Println(fmt.Sprintf("%sBulk indexer %d received a 401 Unauthorized error writing data to Elasticsearch, switching from Elasticsearch client %d to %d", fct, clientWithError, es.currentClient, clientToSwitchTo))
			es.currentClient = clientToSwitchTo
		} else {
			// Do not need to switch clients
			log.Println(fmt.Sprintf("%sBulk indexer %d received a 401 Unauthorized error writing data to Elasticsearch, however switch of Elasticsearch client %d to %d is not needed.", fct, clientWithError, es.currentClient, clientToSwitchTo))
		}
	} else {
		log.Println(fmt.Sprintf("Bulk indexer received an error writing data to Elasticsearch: %s", err))
	}
}

// isUnauthorizedError checks whether the provided error is a 401 unauthorized error
func isUnauthorizedError(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	// Don't like this check, but the err provided is a plain error instance so there is no way
	// programmatically to get the return code:
	return strings.Contains(err.Error(), "401 Unauthorized")
}

// Search attempts to find the provided documents from the provided index
func (es *elasticsearch) Search(input interface{}, index string) (SearchResponse, error) {
	response, err := es.callElasticSearch(input, index, esSearchHandler)
	if err != nil {
		return SearchResponse{}, err
	}
	searchResponse, ok := response.(SearchResponse)
	if !ok {
		err = errors.New("could not cast response of call to Elasticsearch to SearchResponse")
	}

	return searchResponse, err
}

// Index attempts to create or update the provided input in the provided index. Data returned from the
// post (if any) is returned along with an error (if any).
func (es *elasticsearch) Index(input interface{}, index string) (IndexResponse, error) {
	response, err := es.callElasticSearch(input, index, esIndexHandler)
	if err != nil {
		return IndexResponse{}, err
	}
	indexResponse, ok := response.(IndexResponse)
	if !ok {
		err = errors.New("could not cast response of call to Elasticsearch to IndexResponse")
	}

	return indexResponse, err
}

// BulkIndex attempts to create or update the provided input in the provided index. The create or update
// is done as part of a bulk index so will not necessarily be written to Elasticsearch right away.
func (es *elasticsearch) BulkIndex(input interface{}, index string, id string) error {
	const fct = "BulkIndex: "

	itemData, err := json.Marshal(input)
	if err != nil {
		log.Println(fct + "Error occurred marshalling input for bulk indexing")
		return err
	}

	// Add bulk index item to the current bulk indexer:
	es.currentClientMutex.RLock()
	defer es.currentClientMutex.RUnlock()
	err = (*es.bulkIndexers[es.currentClient]).Add(
		context.Background(),
		esutil.BulkIndexerItem{
			Index:      index,
			Action:     "index",
			DocumentID: id,
			Body:       bytes.NewReader(itemData),
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				// nothing to do
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				log.Println(fct + "Bulk indexer encountered an error writing data to Elasticsearch")
			},
		},
	)
	if err != nil {
		log.Println(fct + "Error occurred adding item to bulk indexer")
		return err
	}

	return nil
}

// callElasticSearch wraps calls to Elasticsearch and tries both Elasticsearch clients (user name and passwords) if needed
func (es *elasticsearch) callElasticSearch(input interface{}, index string, handler esHandlerFunc) (interface{}, error) {
	const fct = "callElasticSearch: "

	var err error
	var status int
	var output interface{}
	es.currentClientMutex.RLock()
	startingClient := es.currentClient // remember client that we are going to try first
	es.currentClientMutex.RUnlock()
	currentClient := startingClient
	for {
		output, status, err = handler(input, index, es.clients[currentClient]) // make call through handler
		if err == nil {
			// Success, set current client to the one we just successfully used and break:
			es.currentClientMutex.Lock()
			es.currentClient = currentClient
			es.currentClientMutex.Unlock()
			break
		} else if err != nil {
			log.Println(fmt.Sprintf("%sError occurred calling Elasticsearch using the '%s' index", fct, index))
			is401Error := status == http.StatusUnauthorized
			if !is401Error {
				// No 401 error calling Elasticsearch, set current client to the one we just
				// successfully used and break:
				es.currentClientMutex.Lock()
				es.currentClient = currentClient
				es.currentClientMutex.Unlock()
				break
			} else if is401Error && currentClient == 1 && startingClient != 2 {
				// Client 1 received a 401 error so try client 2 if we have not already tried
				// 2:
				currentClient = 2
			} else if is401Error && currentClient == 2 && startingClient != 1 {
				// Client 2 received a 401 error so try client 1 if we have not already tried
				// 1:
				currentClient = 1
			} else {
				// 401 error, but we have already exhausted all of the clients, so break:
				break
			}
		}
	}
	return output, err
}

// esHandlerFunc is the definition of a function supported by the callElasticSearch function
type esHandlerFunc func(input interface{}, index string, client *esc.Client) (interface{}, int, error)

// esSearchHandler calls the Elasticsearch Search API with the provided input and returns the response
func esSearchHandler(input interface{}, index string, client *esc.Client) (interface{}, int, error) {
	result := SearchResponse{}

	// If input is a string, convert it to JSON first to remove whitespace etc:
	switch inputValue := input.(type) {
	case string:
		var inputAsJSON map[string]interface{}
		err := json.Unmarshal([]byte(inputValue), &inputAsJSON)
		if err != nil {
			return result, -1, err
		}
		input = inputAsJSON
	}

	// Marshal the provided input (query document):
	buf, err := json.Marshal(input)
	if err != nil {
		return result, -1, err
	}
	content := bytes.NewBuffer(buf)

	// Make call to search Elasticsearch:
	resp, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(content),
		client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return result, -1, err
	}
	defer resp.Body.Close()

	// Check for expected status code:
	if resp.StatusCode != http.StatusOK {
		return result, resp.StatusCode, fmt.Errorf("unexpected status code:%d", resp.StatusCode)
	}

	// Read response:
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, -1, err
	}

	// Unmarshal the response:
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, resp.StatusCode, err
	}

	return result, resp.StatusCode, nil
}

// esIndexHandler calls the Elasticsearch Index API with the provided input and returns the response
func esIndexHandler(input interface{}, index string, client *esc.Client) (interface{}, int, error) {
	result := IndexResponse{}

	// Marshal the provided input (query document):
	buf, err := json.Marshal(input)
	if err != nil {
		return result, -1, err
	}
	content := bytes.NewBuffer(buf)

	// Make call to index document in Elasticsearch:
	resp, err := client.Index(index, content)
	if err != nil {
		return result, -1, err
	}
	defer resp.Body.Close()

	// Check for expected status code:
	if resp.StatusCode != http.StatusCreated {
		return result, resp.StatusCode, fmt.Errorf("unexpected status code:%d", resp.StatusCode)
	}

	// Read response:
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, -1, err
	}

	// Unmarshal the response:
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, resp.StatusCode, err
	}

	return result, resp.StatusCode, nil
}

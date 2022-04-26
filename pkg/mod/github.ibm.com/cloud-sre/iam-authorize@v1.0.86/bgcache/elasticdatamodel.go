package bgcache

import (
	"github.ibm.com/cloud-sre/iam-authorize/elasticsearch"
)

const (
	breakglassIndex = "breakglass"

	queryAllCacheEntries = `{
		"query": {
			"match_all": {}
		},
		"size": 10000,
		"sort": [
			{
				"api_keys.last_updated": {
					"nested_path": "api_keys",
					"order": "desc"
				}
			}
		]
	}`
)

var elasticSearch elasticsearch.Elasticsearch

type Permissions struct {
	Permission string `json:"permission"`
	Time int64 `json:"time"`
}

type Resources struct {
	Resource string `json:"resource"`
	ResourcePermissions []Permissions `json:"permissions"`
}

type Key struct {
	APIKey string `json:"api_key"`
	Source string `json:"source"`
	Token string `json:"token"`
	KeyID int64 `json:"key_id"`
	ESResources []Resources `json:"resources"`
	LastUpdated int64 `json:"last_updated"`
}

type ESCacheRecord struct {
	User string `json:"user"`
	IsAPI string `json:"isapi"`
	APIKeys []Key `json:"api_keys"`
}

func getElasticsearch() (elasticsearch.Elasticsearch, error){
	if elasticSearch == nil {
		es, err := elasticsearch.New()
		if err != nil {
			return nil, err
		}

		elasticSearch = es
	}

	return elasticSearch, nil
}

func setElasticsearch(es elasticsearch.Elasticsearch) {
	elasticSearch = es
}

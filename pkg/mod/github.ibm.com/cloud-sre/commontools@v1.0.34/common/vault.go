package common

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	h "github.ibm.com/cloud-sre/commontools/http"
)

const (
	VaultEndpoint   = "https://vserv-us.sos.ibm.com:8200"
	VaultPathPrefix = "/v1/generic/project/oss-kubectl-plugin/"
)

var (
	VaultCache *cache.Cache = cache.New(6*time.Hour, 6*time.Hour)
)

type VaultData struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Data          struct {
		Sha   string `json:"sha"`
		Value string `json:"value"`
	} `json:"data"`
	WrapInfo interface{} `json:"wrap_info"`
	Warnings interface{} `json:"warnings"`
	Auth     interface{} `json:"auth"`
}

type VaultRole struct {
	RequestID     string      `json:"request_id"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int         `json:"lease_duration"`
	Data          interface{} `json:"data"`
	WrapInfo      interface{} `json:"wrap_info"`
	Warnings      interface{} `json:"warnings"`
	Auth          struct {
		ClientToken   string   `json:"client_token"`
		Accessor      string   `json:"accessor"`
		Policies      []string `json:"policies"`
		TokenPolicies []string `json:"token_policies"`
		Metadata      struct {
			Data     string `json:"data"`
			Owner    string `json:"owner"`
			RoleName string `json:"role_name"`
		} `json:"metadata"`
		LeaseDuration int    `json:"lease_duration"`
		Renewable     bool   `json:"renewable"`
		EntityID      string `json:"entity_id"`
		TokenType     string `json:"token_type"`
		Orphan        bool   `json:"orphan"`
	} `json:"auth"`
}

func vaultToken(cacheKey string, vaultRoleID string, vaultSecretID string) (string, error) {
	//curl -X POST -d '{"role_id":"yyy","secret_id":"xxx"}' https://vserv-us.sos.ibm.com:8200/v1/auth/approle/login
	log.Debug("Get client token from vault.")
	body := make(map[string]string)
	body["role_id"] = vaultRoleID
	body["secret_id"] = vaultSecretID

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	url := VaultEndpoint + "/v1/auth/approle/login"

	log.WithFields(log.Fields{"vaultEndpoint": VaultEndpoint}).Debug("Request to get client token.")
	md := &h.Metadata{
		URL:     url,
		Headers: headers,
		Method:  http.MethodPost,
		Body:    body,
		Cookies: nil,
		Timeout: 0,
	}
	result := md.FireHTTPRequest()
	data := result.Data
	err := result.Err

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("HTTP request failed.")
		return "", err
	}

	var vaultRole *VaultRole
	log.Debug("Unmarshal http response data to vaultRole.")
	err = json.Unmarshal(data, &vaultRole)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("Failed to unmarshal http response data to vaultRole.")
		return "", err
	}

	token := vaultRole.Auth.ClientToken
	lease := vaultRole.Auth.LeaseDuration
	log.Debug("Get client token from vault successfully, set it to token.")
	VaultCache.Set(cacheKey, token, time.Duration(lease)*time.Second)
	return token, nil
}

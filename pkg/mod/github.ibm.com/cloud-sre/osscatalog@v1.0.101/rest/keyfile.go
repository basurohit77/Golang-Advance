package rest

// Functions for managing API keys and tokens

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// Credentials represents one entry in a keyfile
type Credentials map[string]string

var keyData map[string]Credentials

//cachedTokens caches IAM tokens obtained from the IAM service
var cachedTokens = make(map[string]cachedTokenType)
var tokensMutex = &sync.Mutex{}

// ResetCachedTokens resets the cache of authentication tokens
// XXX To be used for testing only
func ResetCachedTokens() {
	cachedTokens = make(map[string]cachedTokenType)
}

type cachedTokenType struct {
	token       string
	lastRefresh time.Time
}

// LoadDefaultKeyFile loads the default keyfile from the user's home directory into memory
func LoadDefaultKeyFile() error {
	//	return LoadKeyFile("~/.keys/default.key")
	return LoadKeyFile("~/.keys/osscat.key")
}

// LoadKeyFile loads a given keyfile into memory
func LoadKeyFile(keyFileName string) error {
	ResetCachedTokens()
	if strings.HasPrefix(keyFileName, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set - but required to expand keyFileName=%q", keyFileName)
		}
		keyFileName = home + keyFileName[1:]
	}
	file, err := os.Open(keyFileName) // #nosec G304
	if err != nil {
		return debug.WrapError(err, "Cannot open keyfile %s", keyFileName)
	}
	defer file.Close() // #nosec G307
	rawData, err := ioutil.ReadAll(file)
	if err != nil {
		return debug.WrapError(err, "Error reading keyfile %s", keyFileName)
	}
	err = json.Unmarshal(rawData, &keyData)
	if err != nil {
		return debug.WrapError(err, "Error parsing keyfile %s", keyFileName)
	}
	//	debug.Debug(debug.IAM|debug.Fine, "Keyfile Data: %+v", keyData)
	return nil
}

// GetCredentials returns the named entry from the keyfile
func GetCredentials(name string) (Credentials, error) {
	entry, found := keyData[name]
	if !found {
		return nil, fmt.Errorf("cannot find entry \"%s\" in keyfile", name)
	}
	return entry, nil
}

// GetCredentialsString returns one particular item from the named entry from the keyfile
func GetCredentialsString(name, item string) (string, error) {
	entry, found := keyData[name]
	if !found {
		return "", fmt.Errorf("cannot find entry \"%s\" in keyfile", name)
	}
	if value, found := entry[item]; found {
		return value, nil
	}
	return "", fmt.Errorf("entry %q in keyfile does not contain a %q", name, item)
}

// GetKey returns an API Key with a given name from the keyfile
func GetKey(name string) (string, error) {
	key, err := GetCredentialsString(name, "key")
	if err != nil {
		return "", err
	}
	return key, nil
}

// GetID returns an ID from the entry with a given name from the keyfile
func GetID(name string) (string, error) {
	id, err := GetCredentialsString(name, "id")
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetToken return IAM Token with a given name, either by reading it directly from the keyfile or by generating it from a API Key contained in the keyfile
func GetToken(name string) (string, error) {
	tokensMutex.Lock()
	// TODO: Should avoid keeping the mutex on the entire map through long REST calls
	defer tokensMutex.Unlock()

	if cached, ok := cachedTokens[name]; ok {
		now := time.Now()
		if now.Sub(cached.lastRefresh) < (45 * time.Minute) {
			debug.Debug(debug.IAM, "Returning cached IAM token for key \"%s\"", name)
			return cached.token, nil
		}
		debug.Debug(debug.IAM, "Cached IAM token for key \"%s\" is too old - refreshing", name)
	}
	entry, found := keyData[name]
	if !found {
		return "", fmt.Errorf("cannot find entry \"%s\" in keyfile", name)
	}
	if token, found := entry["token"]; found {
		return token, nil
	}
	var key string
	if key, found = entry["key"]; !found {
		return "", fmt.Errorf("entry `\"%s\" in keyfile does not contain a token or a key", name)
	}

	debug.Debug(debug.IAM, "Obtaining a IAM bearer token from API key entry \"%s\"", name)

	request := make(url.Values)
	request.Set("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	request.Set("response_type", "cloud_iam")
	request.Set("apikey", key)
	response := struct {
		AccessToken string `json:"access_token"`
	}{}

	var actualURL string
	if strings.Contains(strings.ToLower(name), "ys1") || strings.Contains(strings.ToLower(name), "staging") {
		actualURL = IAMTokenURLStaging
	} else {
		if os.Getenv("IAM_URL") != "" {
			actualURL = os.Getenv("IAM_URL") + "/identity/token"
		} else {
			actualURL = IAMTokenURL
		}
	}
	err := DoHTTPPostOrPut("POST", actualURL, "", nil, request, &response, "IAM Token", debug.IAM)
	if err != nil {
		return "", debug.WrapError(err, "could not get IAM access token from API key for entry \"%s\" (HTTP GET error from IAM)", name)
	}
	if response.AccessToken == "" {
		return "", fmt.Errorf("could not get IAM access token from API key for entry \"%s\" (empty response from IAM)", name)
	}
	debug.Debug(debug.IAM, "Obtained access_token for entry \"%s\"", name)
	token := "Bearer " + response.AccessToken
	cached := cachedTokenType{token: token, lastRefresh: time.Now()}
	cachedTokens[name] = cached
	return token, nil
}

// LoadVCAPServices loads key data from a VCAP_SERVICES environment from Cloud Foundry
func LoadVCAPServices(services cfenv.Services) error {
	keyData = make(map[string]Credentials)
	for _, s1 := range services {
		for _, s := range s1 {
			keyName := s.Name
			if keyName == "" {
				continue
			}
			entry := make(Credentials)
			for k, v := range s.Credentials {
				if vs, ok := v.(string); ok {
					entry[k] = vs
					debug.Debug(debug.Main, `LoadVCAPServices() obtained key %q.%q`, keyName, k)
				} else {
					return fmt.Errorf("LoadVCAPServices(): cred %q.%q is not a string: %T %v", keyName, k, v, v)
				}
			}
			keyData[keyName] = entry
		}
	}
	if len(keyData) == 0 {
		return fmt.Errorf("LoadVCAPServices() did not find any keys")
	}
	return nil
}

var secretPattern = regexp.MustCompile(`(SECRETKEY|SECRETID|SECRETTOKEN|SECRETCREDENTIALS)_([^=]+)=(.*)`)

// LoadEnvironmentSecrets loads all secrets (API keys, etc.) from the process environment
func LoadEnvironmentSecrets() error {
	debug.Info("Loading secrets from environment variables")
	keyData = make(map[string]Credentials)
	env := os.Environ()
	for _, e := range env {
		if strings.HasPrefix(e, "SECRET") {
			if m := secretPattern.FindStringSubmatch(e); m != nil {
				keyType := m[1]
				keyName := m[2]
				keyValue := m[3]
				keyName = strings.ReplaceAll(keyName, "_", "-")
				var entry Credentials
				var found bool
				if entry, found = keyData[keyName]; !found {
					entry = make(Credentials)
					entry["name"] = keyName
					keyData[keyName] = entry
				}
				switch keyType {
				case "SECRETKEY":
					entry["key"] = keyValue
				case "SECRETID":
					entry["id"] = keyValue
				case "SECRETTOKEN":
					entry["token"] = keyValue
				case "SECRETCREDENTIALS":
					err := json.Unmarshal(([]byte)(keyValue), &entry)
					if err != nil {
						return fmt.Errorf("error parsing SECRETCREDENTIALS environment variable: %w", err)
					}
				default:
					return fmt.Errorf("unknown secret key type in process environment: %s", e)
				}
				debug.Debug(debug.Main, "Loaded secret %s=%s", keyName, "***" /*keyValue*/)
			} else {
				return fmt.Errorf("unknown secret key variable in process environment: %s", e)
			}
		}
	}
	return nil
}

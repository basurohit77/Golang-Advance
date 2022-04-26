package token

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetKeysIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	//assert := require.New(t)
	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "9H9cXOTxNW-WIiYAgJNNnJainWa91X6Dqrsp95Sh8Py5aOr9XhZWiQ5T8tXNev4GLzRevsgvWUn3zRQpTDZk3aDURj-936Hlfx-AlbyGAC2cbUrYMRSZ3obdQV8k1LKPuf7FEdXyLxz18_h-XYMcnuwWVmXcw8wSELaJgHMn93aaoM7L8J5SdXZEkO5oEscarp4dnutO2ktf26QnCHBqkpzHPNjpV3dgwYnETQ3ryKDyazuZ2MUjSHAIPXBlLGhPUtz-uX8zML-thiD4Svun_swon1ZGcTiDpIzHMSOuU8bk9Y3xrSNYexXqJuA7f5gy8W01Ph1iz72gzhNunkftpQ",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20190122"
						},
						{
							"kty": "RSA",
							"n": "1wGR_xspvgMBZ8YdIhXlAv9UNdsZczjgUTmb_QwLdngxfnfUgm4dtL2tnPb1tZE8Ji2cV68bO8oKHyG93SyzChSZyB5SislSCxBZDQfvP4dt5mfU4RiQDH2UQ_4YBt43OVfWT5GBzOALIzlR-oRuARglUNO0ZGWThcHkWlbTzLTOwKMxi3c1XP30uQCOydm0yY1meRIa-HwUWu5_hms234nV4-stmSEZHyWbPSgSZGXE3lsD7YeM5o6a_d7zQu_wZ2u-UHDdX94ePkqmlDMhhslbyoI9W8BQcrgABAGqVDeP2jE3dm88mSLas0ekpnyN0PeRnt1nBONc9eJYIYjr3Q",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20180701"
						},
						{
							"kty": "RSA",
							"n": "1YuoV3LKKU6qTcBrWzuMPXmBTlbusvCxrtxY1nD7rIpyTHCVK59vovFeJlPcZD7aCmAyQWmfy5sp8kc8bvH_2uFhMyQIaTr3AK7ORS8kuiWh-0P_8d31ID8zpt9KG-rdfOnzCh3L0CZxCAHWUk-U29JDstQ6FACNDFZq1bdo3oXXNGNNG4dSxpUK8nBkZGWv42AE_0FH-1fvBbdOL78VNNPcjySvjuyW82cTO6O1ju22vRYbp946wr-WIp9O7qzZA72B9bd19LeCRcRLgfgILr2XLOTQ-BFAQT6ueivTqKsq3658TsiL3TwH-Y7RlbWNJwechsZLPXS7qGTJud3mLw",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20171025-16:27:10"
						}
					]
				}`))
		}),
	)

	defer ts.Close()

	// test the refresh channel
	endpoints := &Endpoints{
		TokenEndpoint: ts.URL,
		KeyEndpoint:   ts.URL,
	}

	initializeEnvKeyCacheIfNeeded(endpoints.KeyEndpoint, defaultKeyCacheExpiry)
	//keyCache[endpoints.Host]. .initCache()

	if keyCache[endpoints.KeyEndpoint].endpoint != ts.URL {
		t.Errorf("Configuration was not updated correctly, expected %v, got %v", ts.URL, keyCache[endpoints.KeyEndpoint].endpoint)
	}

	time.Sleep(2 * time.Second)
	keys := GetKeys(ts.URL)

	contains2019 := strings.Contains(keys.Keys[0].Kid, "2019")
	contains2017 := strings.Contains(keys.Keys[2].N, "1YuoV3LKK")

	if !contains2019 || !contains2017 {
		t.Errorf("GetKeys() returns %v %v; want 2019, 1YuoV3LKK", keys.Keys[0].Kid, keys.Keys[2].N)
	}

	keyCache[ts.URL].quitCacheLoop <- true
	delete(keyCache, ts.URL)
}

func TestCacheRefreshIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	timesCalled := 0
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if timesCalled == 0 {
				timesCalled++
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "9H9cXOTxNW-WIiYAgJNNnJainWa91X6Dqrsp95Sh8Py5aOr9XhZWiQ5T8tXNev4GLzRevsgvWUn3zRQpTDZk3aDURj-936Hlfx-AlbyGAC2cbUrYMRSZ3obdQV8k1LKPuf7FEdXyLxz18_h-XYMcnuwWVmXcw8wSELaJgHMn93aaoM7L8J5SdXZEkO5oEscarp4dnutO2ktf26QnCHBqkpzHPNjpV3dgwYnETQ3ryKDyazuZ2MUjSHAIPXBlLGhPUtz-uX8zML-thiD4Svun_swon1ZGcTiDpIzHMSOuU8bk9Y3xrSNYexXqJuA7f5gy8W01Ph1iz72gzhNunkftpQ",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20190122"
						}
					]
				}`))

			} else if timesCalled == 1 {
				timesCalled++
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "1wGR_xspvgMBZ8YdIhXlAv9UNdsZczjgUTmb_QwLdngxfnfUgm4dtL2tnPb1tZE8Ji2cV68bO8oKHyG93SyzChSZyB5SislSCxBZDQfvP4dt5mfU4RiQDH2UQ_4YBt43OVfWT5GBzOALIzlR-oRuARglUNO0ZGWThcHkWlbTzLTOwKMxi3c1XP30uQCOydm0yY1meRIa-HwUWu5_hms234nV4-stmSEZHyWbPSgSZGXE3lsD7YeM5o6a_d7zQu_wZ2u-UHDdX94ePkqmlDMhhslbyoI9W8BQcrgABAGqVDeP2jE3dm88mSLas0ekpnyN0PeRnt1nBONc9eJYIYjr3Q",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20180701"
						}
					]
				}`))
			} else {
				if timesCalled != 2 {
					t.Errorf("Keys endpoint should have been called twice, but was called %v", timesCalled)
				}
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "1YuoV3LKKU6qTcBrWzuMPXmBTlbusvCxrtxY1nD7rIpyTHCVK59vovFeJlPcZD7aCmAyQWmfy5sp8kc8bvH_2uFhMyQIaTr3AK7ORS8kuiWh-0P_8d31ID8zpt9KG-rdfOnzCh3L0CZxCAHWUk-U29JDstQ6FACNDFZq1bdo3oXXNGNNG4dSxpUK8nBkZGWv42AE_0FH-1fvBbdOL78VNNPcjySvjuyW82cTO6O1ju22vRYbp946wr-WIp9O7qzZA72B9bd19LeCRcRLgfgILr2XLOTQ-BFAQT6ueivTqKsq3658TsiL3TwH-Y7RlbWNJwechsZLPXS7qGTJud3mLw",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20171025-16:27:10"
						}
					]
				}`))

			}
		}),
	)

	defer ts.Close()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Errorf("API_KEY must be set for testing.\n")
		t.Fail()
	}

	// test the refresh channel

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL,
		KeyEndpoint:   ts.URL,
	}

	initializeEnvKeyCacheIfNeeded(endpoints.KeyEndpoint, 3)

	keyCacheMutex.Lock()
	key, ok := keyCache[ts.URL]
	keyCacheMutex.Unlock()

	assert.True(t, ok)

	if key.endpoint != ts.URL || key.keyCacheExpiry != 3 {
		t.Errorf("Configuration was not updated correctly, expected %v %v seconds, got %v %v seconds", ts.URL, 3, key.endpoint, key.keyCacheExpiry)
	}
	time.Sleep(1 * time.Second)
	keys := GetKeys(ts.URL)

	// validate first get
	keyValueModulus := strings.Contains(keys.Keys[0].N, "9H9cXOTxNW-")
	kidValue := strings.Contains(keys.Keys[0].Kid, "2019")

	if !keyValueModulus || !kidValue {
		t.Errorf("GetKeys() returns %v, %v; want 9H9cXOTxNW-, 2019", keys.Keys[0].N, keys.Keys[0].Kid)
	}

	// eventually consistent
	// do refresh of the cache

	for i := 0; i < 5; i++ {
		time.Sleep(1000 * time.Millisecond) // give some time for the gojwks call to happen
		//validate second get
		keys = GetKeys(ts.URL)
		key := strings.Contains(keys.Keys[0].N, "1wGR_xspvgM")
		if key {
			break
		}
	}
	keyValueModulus = strings.Contains(keys.Keys[0].N, "1wGR_xspvgM")
	kidValue = strings.Contains(keys.Keys[0].Kid, "2018")

	if !keyValueModulus || !kidValue {
		t.Errorf("GetKeys() returns %v, %v; want 1wGR_xspvgM, 2018", keys.Keys[0].N, keys.Keys[0].Kid)
	}

	// test quit channel
	keyCache[endpoints.KeyEndpoint].quitCacheLoop <- true

	// should be equal to 2018 KID value
	time.Sleep(4 * time.Second)
	keys = GetKeys(ts.URL)
	keyValueModulus = strings.Contains(keys.Keys[0].N, "1wGR_xspvgM")
	kidValue = strings.Contains(keys.Keys[0].Kid, "2018")

	if !keyValueModulus || !kidValue {
		t.Errorf("GetKeys() returns %v, %v; want 1wGR_xspvgM, 2018", keys.Keys[0].N, keys.Keys[0].Kid)
	}

	delete(keyCache, endpoints.KeyEndpoint)
}

func TestCacheRefreshRecoveryFromFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cache refresh recovery unit test")
	}

	timesCalled := 0
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if timesCalled == 0 {
				timesCalled++
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "9H9cXOTxNW-WIiYAgJNNnJainWa91X6Dqrsp95Sh8Py5aOr9XhZWiQ5T8tXNev4GLzRevsgvWUn3zRQpTDZk3aDURj-936Hlfx-AlbyGAC2cbUrYMRSZ3obdQV8k1LKPuf7FEdXyLxz18_h-XYMcnuwWVmXcw8wSELaJgHMn93aaoM7L8J5SdXZEkO5oEscarp4dnutO2ktf26QnCHBqkpzHPNjpV3dgwYnETQ3ryKDyazuZ2MUjSHAIPXBlLGhPUtz-uX8zML-thiD4Svun_swon1ZGcTiDpIzHMSOuU8bk9Y3xrSNYexXqJuA7f5gy8W01Ph1iz72gzhNunkftpQ",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20190122"
						}
					]
				}`))
			} else if timesCalled == 1 {
				timesCalled++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("500 - Error in test server generating response to getting token."))
				if err != nil {
					fmt.Println(err)
				}
			} else if timesCalled == 2 {
				timesCalled++
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "1wGR_xspvgMBZ8YdIhXlAv9UNdsZczjgUTmb_QwLdngxfnfUgm4dtL2tnPb1tZE8Ji2cV68bO8oKHyG93SyzChSZyB5SislSCxBZDQfvP4dt5mfU4RiQDH2UQ_4YBt43OVfWT5GBzOALIzlR-oRuARglUNO0ZGWThcHkWlbTzLTOwKMxi3c1XP30uQCOydm0yY1meRIa-HwUWu5_hms234nV4-stmSEZHyWbPSgSZGXE3lsD7YeM5o6a_d7zQu_wZ2u-UHDdX94ePkqmlDMhhslbyoI9W8BQcrgABAGqVDeP2jE3dm88mSLas0ekpnyN0PeRnt1nBONc9eJYIYjr3Q",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20180701"
						}
					]
				}`))
			} else {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{}`))
				if err != nil {
					fmt.Println(err)
				}
			}
		}),
	)

	defer ts.Close()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Errorf("API_KEY must be set for testing.\n")
	}

	// test the refresh channel
	os.Setenv("KEY_CACHE_EXPIRY", "3")

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL,
		KeyEndpoint:   ts.URL,
	}

	initializeEnvKeyCacheIfNeeded(endpoints.KeyEndpoint, 3)

	keyCacheMutex.Lock()
	key, ok := keyCache[ts.URL]
	keyCacheMutex.Unlock()

	assert.True(t, ok)

	if key.endpoint != ts.URL || key.keyCacheExpiry != 3 {
		t.Errorf("Configuration was not updated correctly, expected %v %v seconds, got %v %v seconds", ts.URL, 3, key.endpoint, key.keyCacheExpiry)
	}

	time.Sleep(1 * time.Second)
	// first get should succeed
	keys := GetKeys(ts.URL)

	// validate second get
	keyValueModulus := strings.Contains(keys.Keys[0].N, "9H9cXOTxNW-")
	kidValue := strings.Contains(keys.Keys[0].Kid, "2019")

	if !keyValueModulus || !kidValue {
		t.Errorf("GetKeys() returns %v, %v; want 9H9cXOTxNW-, 2019", keys.Keys[0].N, keys.Keys[0].Kid)
	}

	// do refresh of the cache
	time.Sleep(4500 * time.Millisecond) // give some time for the gojwks call to happen

	//validate second get, should fail to retreive key from identity endpoint but should still contain old keys
	keys = GetKeys(ts.URL)

	keyValueModulus = strings.Contains(keys.Keys[0].N, "9H9cXOTxNW-")
	kidValue = strings.Contains(keys.Keys[0].Kid, "2019")

	if !keyValueModulus || !kidValue {
		t.Errorf("GetKeys() returns %v, %v; want 1wGR_xsp, 20180701", keys.Keys[0].N, keys.Keys[0].Kid)
	}

	// 3rd fetch should update keys to new values
	time.Sleep(4 * time.Second) // give some time for the gojwks call to happen

	keys = GetKeys(ts.URL)

	keyValueModulus = strings.Contains(keys.Keys[0].N, "1wGR_x")
	kidValue = strings.Contains(keys.Keys[0].Kid, "20180701")

	if !keyValueModulus || !kidValue {
		t.Errorf("GetKeys() returns %v, %v; want 1wGR_xs, 20180701", keys.Keys[0].N, keys.Keys[0].Kid)
	}

	// 4th fetch should have 200 with invalid object, but function should return old keys from last fetch
	time.Sleep(3 * time.Second) // give some time for the gojwks call to happen

	keys = GetKeys(ts.URL)

	keyValueModulus = strings.Contains(keys.Keys[0].N, "1wGR_x")
	kidValue = strings.Contains(keys.Keys[0].Kid, "20180701")

	if !keyValueModulus || !kidValue {
		t.Errorf("GetKeys() returns %v, %v; want 9H9cXOTxNW-, 2019", keys.Keys[0].N, keys.Keys[0].Kid)
	}

	// test quit channel
	keyCache[ts.URL].quitCacheLoop <- true
	time.Sleep(1 * time.Second)
	delete(keyCache, ts.URL)
}

func Test_MultipleCacheKeys(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() == "/key1"+KeyPath {
				// path should be '/identity/keys'
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "1wGR_xspvgMBZ8YdIhXlAv9UNdsZczjgUTmb_QwLdngxfnfUgm4dtL2tnPb1tZE8Ji2cV68bO8oKHyG93SyzChSZyB5SislSCxBZDQfvP4dt5mfU4RiQDH2UQ_4YBt43OVfWT5GBzOALIzlR-oRuARglUNO0ZGWThcHkWlbTzLTOwKMxi3c1XP30uQCOydm0yY1meRIa-HwUWu5_hms234nV4-stmSEZHyWbPSgSZGXE3lsD7YeM5o6a_d7zQu_wZ2u-UHDdX94ePkqmlDMhhslbyoI9W8BQcrgABAGqVDeP2jE3dm88mSLas0ekpnyN0PeRnt1nBONc9eJYIYjr3Q",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20180701"
						}
					]
				}`))
			} else if r.URL.String() == "/key2"+KeyPath {
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "9H9cXOTxNW-WIiYAgJNNnJainWa91X6Dqrsp95Sh8Py5aOr9XhZWiQ5T8tXNev4GLzRevsgvWUn3zRQpTDZk3aDURj-936Hlfx-AlbyGAC2cbUrYMRSZ3obdQV8k1LKPuf7FEdXyLxz18_h-XYMcnuwWVmXcw8wSELaJgHMn93aaoM7L8J5SdXZEkO5oEscarp4dnutO2ktf26QnCHBqkpzHPNjpV3dgwYnETQ3ryKDyazuZ2MUjSHAIPXBlLGhPUtz-uX8zML-thiD4Svun_swon1ZGcTiDpIzHMSOuU8bk9Y3xrSNYexXqJuA7f5gy8W01Ph1iz72gzhNunkftpQ",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20190122"
						}
					]
				}`))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("400 - Invalid path in test."))
				if err != nil {
					fmt.Println(err)
				}
			}
		}))
	defer ts.Close()

	initializeEnvKeyCacheIfNeeded(ts.URL+"/key1"+KeyPath, defaultKeyCacheExpiry)
	keys1, err := GetKeyByKID("20180701", ts.URL+"/key1"+KeyPath)
	assert.Nil(t, err)
	assert.NotEmpty(t, keys1)

	initializeEnvKeyCacheIfNeeded(ts.URL+"/key2"+KeyPath, defaultKeyCacheExpiry)
	keys2, err := GetKeyByKID("20190122", ts.URL+"/key2"+KeyPath)
	assert.Nil(t, err)
	assert.NotEmpty(t, keys2)

	assert.NotEqual(t, keys1, keys2)

	// test invalid KID

	keys3, err := GetKeyByKID("123456", ts.URL+"/key1"+KeyPath)
	assert.Empty(t, keys3)
	errString := "KID 123456 not found for environment " + ts.URL + "/key1" + KeyPath + ". Available keys for this environment are: {Keys:[{Kid:20180701 Kty:RSA Alg:RS256 Use: X5c:[] X5t: N:1wGR_xspvgMBZ8YdIhXlAv9UNdsZczjgUTmb_QwLdngxfnfUgm4dtL2tnPb1tZE8Ji2cV68bO8oKHyG93SyzChSZyB5SislSCxBZDQfvP4dt5mfU4RiQDH2UQ_4YBt43OVfWT5GBzOALIzlR-oRuARglUNO0ZGWThcHkWlbTzLTOwKMxi3c1XP30uQCOydm0yY1meRIa-HwUWu5_hms234nV4-stmSEZHyWbPSgSZGXE3lsD7YeM5o6a_d7zQu_wZ2u-UHDdX94ePkqmlDMhhslbyoI9W8BQcrgABAGqVDeP2jE3dm88mSLas0ekpnyN0PeRnt1nBONc9eJYIYjr3Q E:AQAB}]}"
	assert.Contains(t, err.Error(), errString)
}

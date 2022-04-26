package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

var authConfig map[string]interface{} // to hold the cfenv AppID Credentials

type authCacheEntry struct {
	AccessToken string
	User        string
	WriteAccess bool
	Expiration  time.Time
}

// authCache is the global cache of authenticated users and their tokens, in the running server
var authCache = struct {
	mutex sync.Mutex
	data  map[string]*authCacheEntry
}{
	mutex: sync.Mutex{},
	data:  make(map[string]*authCacheEntry),
}

type userEntry struct {
	Email       string `json:"email"`
	WriteAccess bool   `json:"write_access"`
}

// authorizedUsers is the global list of known users and their authorization, used to validate users when they authenticate to the server
var authorizedUsers = make(map[string]*userEntry)

func setupAuth(authFileName string) error {
	var services cfenv.Services

	// Read the VCAP_SERVICES info to obtain the AppID credentials
	if services = options.GlobalOptions().CFServices; services == nil {
		return fmt.Errorf("Could not find any VCAP information")
	}
	matches, err := services.WithNameUsingPattern("appid.*")
	if err != nil {
		return fmt.Errorf("Found no appid service in the VCAP information: %v", err)
	}
	if len(matches) != 1 {
		results := make([]string, 0, len(matches))
		for _, r := range matches {
			results = append(results, r.Name)
		}
		return fmt.Errorf("Expected exactly one appid service in the VCAP information but got: %v", results)
	}
	serviceName := matches[0].Name
	creds := matches[0].Credentials
	if creds["clientId"] == "" {
		return fmt.Errorf("Missing clientId for appid service %s in the VCAP information", serviceName)
	}
	if creds["oauthServerUrl"] == "" {
		return fmt.Errorf("Missing oauthServerUrl for appid service %s in the VCAP information", serviceName)
	}
	if creds["secret"] == "" {
		return fmt.Errorf("Missing secret for appid service %s in the VCAP information", serviceName)
	}

	// Compute the appropriate AppID redirect URL, based on this server's URL
	var appURL string
	if opt := options.GlobalOptions().CFEnv; opt != nil {
		if len(opt.ApplicationURIs) > 0 {
			appURL = opt.ApplicationURIs[0]
		}
	}
	if appURL != "" {
		creds["afterAuthUrl"] = fmt.Sprintf("https://%s/afterauth", appURL)
	} else {
		creds["afterAuthUrl"] = fmt.Sprintf("http://localhost%s/afterauth", getPort())
	}
	debug.Info("Successfully configured to use appid service %s - afterAuth=%s", serviceName, creds["afterAuthUrl"])

	authConfig = creds

	// Read the list of authorized users
	authFile, err := os.Open(authFileName) // #nosec G304
	if err != nil {
		return debug.WrapError(err, "Cannot open authorized users file %s", authFileName)
	}
	defer authFile.Close() // #nosec G307
	users := make([]userEntry, 0)
	rawData, err := ioutil.ReadAll(authFile)
	if err != nil {
		return debug.WrapError(err, "Error reading authorized users file %s", authFileName)
	}
	err = json.Unmarshal(rawData, &users)
	if err != nil {
		return debug.WrapError(err, "Error parsing authorized users file %s", authFileName)
	}
	for _, u := range users {
		debug.Debug(debug.AppID, "Registering authorized user %v", u)
		ue := u
		authorizedUsers[u.Email] = &ue
	}
	debug.Info("Registered %d authorized users from file %s", len(authorizedUsers), authFileName)

	return nil
}

func checkToken(w http.ResponseWriter, r *http.Request) (loggedInUser string, updateEnabled bool, loginMessage string) {
	var token = ""
	cookie, err := r.Cookie(CookieName)
	if err == nil {
		// TODO: check for errors other than ErrNoCookie
		debug.Debug(debug.AppID, "checkToken() got cookie: %#v", cookie)
		if cookie != nil {
			token = cookie.Value
		}
	} else {
		debug.Debug(debug.AppID, "checkToken() no cookie received (err=%v)", err)
	}
	if token != "" {
		if *dummyAuth && token == dummyAuthToken {
			return "Dummy Auth", true, ""
		}
		authCache.mutex.Lock()
		defer authCache.mutex.Unlock()

		if entry := authCache.data[token]; entry != nil {
			now := time.Now()
			if now.Sub(entry.Expiration) < 0 {
				// This is the normal exit from this function, for logged-in users
				return entry.User, entry.WriteAccess, ""
			}
			return fmt.Sprintf("%s(token expired)", entry.User), false, fmt.Sprintf("Access token for user %s has expired - forcing a new login", entry.User)
		}
		clearCookie(w, r)
		return "(unknown)", false, "Access token not found in cache - forcing a new login"
	}
	return "", false, "No Access token - initiating login"
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := checkRequest(w, r, false); !ok {
		return
	}

	authURL := fmt.Sprintf("%s/authorization?response_type=code&client_id=%s&redirect_uri=%s&scope=openid", authConfig["oauthServerUrl"], authConfig["clientId"], authConfig["afterAuthUrl"])
	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

func afterAuthHandler(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := checkRequest(w, r, false); !ok {
		return
	}

	var now = time.Now()

	var tokenRequest struct {
		GrantType   string `json:"grant_type"`
		RedirectURI string `json:"redirect_uri"`
		Code        string `json:"code"`
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	var userinfoResponse struct {
		Name       string                   `json:"name"`
		Email      string                   `json:"email"`
		Picture    string                   `json:"picture"`
		Locale     string                   `json:"locale"`
		Gender     string                   `json:"gender"`
		Identities []map[string]interface{} `json:"identities"`
		Sub        string                   `json:"sub"`
	}

	tokenURL := fmt.Sprintf("%s/token", authConfig["oauthServerUrl"])
	userinfoURL := fmt.Sprintf("%s/userinfo", authConfig["oauthServerUrl"])

	codes := r.URL.Query()["code"]
	if len(codes) != 1 {
		errorPage(w, http.StatusInternalServerError, "afterAuthHandler expected a single authorization code but got %d codes", len(codes))
		return
	}
	code := codes[0]
	if len(code) < 20 {
		errorPage(w, http.StatusInternalServerError, "afterAuthHandler got empty/short authorization code (len=%d)", len(code))
		return
	}
	debug.Debug(debug.AppID, "afterAuthHandler() got authorization code: %s", code)

	/*
		// "Cheat" to create a Basic Auth header
		dummyReq, _ := http.NewRequest(http.MethodPost, tokenURL, nil)
		dummyReq.SetBasicAuth(authConfig["clientId"].(string), authConfig["secret"].(string))
		dummyReq.BasicAuth() // Force generation of the header?
		tokenAuthorization := r.Header.Get("Authorization")
		debug.Debug(debug.AppID, "afterAuthHandler() using authorization header: %s", tokenAuthorization)
	*/

	// Get the tokens
	tokenAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(authConfig["clientId"].(string)+":"+authConfig["secret"].(string)))
	tokenRequest.GrantType = "authorization_code"
	tokenRequest.RedirectURI = authConfig["afterAuthUrl"].(string)
	tokenRequest.Code = code
	err := rest.DoHTTPPostOrPut(http.MethodPost, tokenURL, tokenAuthorization, nil, &tokenRequest, &tokenResponse, "AppId/tokens", debug.AppID)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "afterAuthHandler error getting tokens from AppID: %v", err)
		return
	}
	debug.Debug(debug.AppID, "afterAuthHandler() got access token: %s", tokenResponse.AccessToken)

	// Get the userinfo
	userinfoAuthorization := "Bearer " + tokenResponse.AccessToken
	err = rest.DoHTTPGet(userinfoURL, userinfoAuthorization, nil, "AppId/userinfo", debug.AppID, &userinfoResponse)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "afterAuthHandler error getting userinfo from AppID: %v", err)
		return
	}
	debug.Debug(debug.AppID, "afterAuthHandler() got userinfo: %s", userinfoResponse.Email)

	// Setup an entry in the authenticated users cache
	authCache.mutex.Lock()
	defer authCache.mutex.Unlock()
	entry := authCache.data[tokenResponse.AccessToken]
	if entry == nil {
		entry = &authCacheEntry{}
	}
	entry.AccessToken = tokenResponse.AccessToken
	entry.Expiration = now.Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	entry.User = userinfoResponse.Email
	user := authorizedUsers[userinfoResponse.Email]
	if user != nil && user.WriteAccess {
		entry.WriteAccess = true
	} else {
		entry.WriteAccess = false
	}
	authCache.data[tokenResponse.AccessToken] = entry
	debug.Debug(debug.AppID, "afterAuthHandler() writing auth cache entry user=%s writeAcces=%v expiration=%v", entry.User, entry.WriteAccess, entry.Expiration)
	debug.Audit("Successful login for user=%s writeAcces=%v expiration=%v", entry.User, entry.WriteAccess, entry.Expiration.UTC().Format(timeFormat))

	cookie := http.Cookie{
		Name:  CookieName,
		Value: entry.AccessToken,
		Path:  "/",
		//		Domain:  r.Host,
		Expires: entry.Expiration,
	}
	ix := strings.Index(r.Host, ":")
	if ix >= 0 {
		cookie.Domain = r.Host[:ix]
	} else {
		cookie.Domain = r.Host
	}
	http.SetCookie(w, &cookie)

	debug.Debug(debug.AppID, "afterAuthHandler() complete - setting %s cookie and redirecting to home page", CookieName)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearCookie(w http.ResponseWriter, r *http.Request) {
	debug.Debug(debug.AppID, "Clearing the %s cookie", CookieName)
	cookie := http.Cookie{
		Name:    CookieName,
		Value:   "",
		Path:    "/",
		Domain:  r.Host,
		Expires: time.Now(),
	}
	http.SetCookie(w, &cookie)
}

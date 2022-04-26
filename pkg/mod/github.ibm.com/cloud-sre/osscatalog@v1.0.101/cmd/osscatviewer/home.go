package main

// Functions for the home page of osscatviewer

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

var homeTemplate *template.Template

// HomeData is used to pass all the data for the home template
type HomeData struct {
	Services     []*cachedService
	Pattern      string
	LastRefresh  string
	LoggedInUser string
}

var lastLogDNAFLush time.Time

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, false); !ok {
		return
	}

	if r.URL.Path != "/" {
		errorPage(w, http.StatusNotFound, `Page not found: %s`, r.URL.Path)
		return
	}

	defer func() {
		now := time.Now()
		if now.Sub(lastLogDNAFLush) > 5*time.Minute {
			debug.FlushLogDNA()
			lastLogDNAFLush = now
		}
	}()

	services0, lastRefresh := allRecords.getAllServices()
	if services0 == nil {
		errorPage(w, http.StatusNoContent, `Initialization in progress -- please try again in a few minutes`)
		err := allRecords.refresh()
		if err != nil {
			errorPage(w, http.StatusInternalServerError, `Error refreshing OSS data from Global Catalog: %v`, err)
			return
		}
		services0, lastRefresh = allRecords.getAllServices()
		return
	}

	patternString := r.URL.Query().Get("pattern")

	var homeData = HomeData{}
	homeData.LastRefresh = lastRefresh.UTC().Format(timeFormat)
	homeData.Pattern = patternString

	if patternString != "" {
		// TODO: Cache the filtered records
		pattern, err := regexp.Compile(patternString)
		if err != nil {
			errorPage(w, http.StatusInternalServerError, `Invalid pattern="%s": %v`, patternString, err)
			return
		}
		// No need to re-sort the filtered services, as we started with a sorted list already
		homeData.Services = make([]*cachedService, 0, 250)
		for _, e := range services0 {
			if pattern.FindString(e.FilterString) != "" {
				homeData.Services = append(homeData.Services, e)
			}
		}
	} else {
		homeData.Services = services0
	}
	homeData.LoggedInUser = loggedInUser

	if *dummyAuth {
		cookie := http.Cookie{
			Name:  CookieName,
			Value: dummyAuthToken,
			Path:  "/",
		}
		debug.PrintError("*** For debugging only: setting dummy authentication cookie %s", CookieName)
		http.SetCookie(w, &cookie)
	}

	err := homeTemplate.Execute(w, &homeData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := checkRequest(w, r, false); !ok {
		return
	}

	err := allRecords.refresh()
	if err != nil {
		errorPage(w, http.StatusInternalServerError, `Error refreshing OSS data from Global Catalog: %v`, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// makeFilterString creates a filter string used for filtering across service entries
// The filter string contains a combination of name, type, status, tags, etc.
func makeFilterString(r *ossrecord.OSSService) string {
	return fmt.Sprintf("%s/%s/type:%s/status:%s->%s/tag:%s/phase:%s", r.ReferenceResourceName, r.ReferenceDisplayName, r.GeneralInfo.EntryType, r.GeneralInfo.OperationalStatus, r.GeneralInfo.FutureOperationalStatus, r.GeneralInfo.OSSTags, r.GeneralInfo.OSSOnboardingPhase)

}

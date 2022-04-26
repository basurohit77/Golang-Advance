package main

import (
	"html/template"
	"net/http"
)

var environmentsTemplate *template.Template

// EnvironmentsData is used to pass all the data for the environments template
type EnvironmentsData struct {
	Environments []*cachedEnvironment
	LastRefresh  string
	LoggedInUser string
}

func environmentsHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, false); !ok {
		return
	}

	environments, lastRefresh := allRecords.getAllEnvironments()
	if environments == nil {
		err := allRecords.refresh()
		if err != nil {
			errorPage(w, http.StatusInternalServerError, `Error refreshing OSS data from Global Catalog: %v`, err)
			return
		}
		environments, lastRefresh = allRecords.getAllEnvironments()
	}

	var environmentsData = EnvironmentsData{}
	environmentsData.LastRefresh = lastRefresh.UTC().Format(timeFormat)
	environmentsData.Environments = environments
	environmentsData.LoggedInUser = loggedInUser

	err := environmentsTemplate.Execute(w, &environmentsData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

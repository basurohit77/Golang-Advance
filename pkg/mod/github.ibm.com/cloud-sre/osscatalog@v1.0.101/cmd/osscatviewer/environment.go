package main

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
)

var environmentTemplate *template.Template

// EnvironmentData is used to pass all the data for the environment template
type EnvironmentData struct {
	*ossrecordextended.OSSEnvironmentExtended        // embedded type
	Created                                   string // Creation time of the OSS entry in Catalog
	Updated                                   string // Last update time of the OSS entry in Catalog
	LastRefresh                               string
	LinkToOSSRecord                           string
	ProductionDiffID                          ossrecord.CatalogID
	LoggedInUser                              string // ID of the user, if logged-in
}

func (v *EnvironmentData) populate(env *ossrecordextended.OSSEnvironmentExtended, cache *AllRecordsType, loggedInUser string, updateEnabled, editMode bool) {
	v.OSSEnvironmentExtended = env
	v.Created = env.Created
	v.Updated = env.Updated
	var lastRefresh time.Time
	v.LastRefresh = lastRefresh.UTC().Format(timeFormat)
	v.LinkToOSSRecord, _ = catalog.GetOSSEntryUI(env)
	v.ProductionDiffID = ossrecord.CatalogID(env.GetOSSEntryID())
	v.LoggedInUser = loggedInUser
	return
}

// environmentHandler is the HTTP handler for the /environment page
func environmentHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, false); !ok {
		return
	}

	environmentID := ossrecord.EnvironmentID((strings.TrimPrefix(r.URL.Path, "/environment/")))
	if environmentID == "" {
		errorPage(w, http.StatusNotFound, `Empty environmentID. Use the URI "/environment/<environment-id>" to view information about <environment-id>`)
		return
	}

	env0, err := catalog.ReadOSSEntryByID(ossrecord.MakeOSSEnvironmentID(environmentID), catalog.IncludeEnvironments|catalog.IncludeOSSValidation|catalog.IncludeOSSTimestamps) /* We include the OSSValidation info for viewing */
	if err != nil {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": %v", environmentID, err)
		return
	}
	env, ok := env0.(*ossrecordextended.OSSEnvironmentExtended)
	if !ok {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": invalid type %T", environmentID, env0)
		return
	}

	environmentData := &EnvironmentData{}
	environmentData.populate(env, &allRecords, loggedInUser, false, false)

	err = environmentTemplate.Execute(w, environmentData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

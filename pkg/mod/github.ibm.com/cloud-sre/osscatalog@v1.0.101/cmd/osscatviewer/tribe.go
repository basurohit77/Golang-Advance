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

var tribeTemplate *template.Template

// TribeData is used to pass all the data for the tribe template
type TribeData struct {
	*ossrecordextended.OSSTribeExtended        // embedded type
	Created                             string // Creation time of the OSS entry in Catalog
	Updated                             string // Last update time of the OSS entry in Catalog
	SegmentName                         string
	Services                            []*cachedService
	LastRefresh                         string
	LinkToOSSRecord                     string
	ProductionDiffID                    ossrecord.CatalogID
	LoggedInUser                        string // ID of the user, if logged-in
}

func (v *TribeData) populate(tr *ossrecordextended.OSSTribeExtended, cache *AllRecordsType, loggedInUser string, updateEnabled, editMode bool) {
	v.OSSTribeExtended = tr
	v.Created = tr.Created
	v.Updated = tr.Updated
	v.SegmentName = cache.getSegmentName(tr.SegmentID)
	var lastRefresh time.Time
	v.Services, lastRefresh = cache.getServicesForTribeID(tr.TribeID)
	v.LastRefresh = lastRefresh.UTC().Format(timeFormat)
	v.LinkToOSSRecord, _ = catalog.GetOSSEntryUI(tr)
	v.ProductionDiffID = ossrecord.CatalogID(tr.GetOSSEntryID())
	v.LoggedInUser = loggedInUser
	return
}

// tribeHandler is the HTTP handler for the /tribe page
func tribeHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, false); !ok {
		return
	}

	tribeID := strings.TrimPrefix(r.URL.Path, "/tribe/")
	if tribeID == "" {
		errorPage(w, http.StatusNotFound, `Empty tribeID. Use the URI "/tribe/<tribe-id>" to view information about <tribe-id>`)
		return
	}

	tr := allRecords.getPlaceHolderTribe(ossrecord.TribeID(tribeID))
	if tr == nil {
		var err error
		tr0, err := catalog.ReadOSSEntryByID(ossrecord.MakeOSSTribeID(ossrecord.TribeID(tribeID)), catalog.IncludeTribes|catalog.IncludeAll)
		if err != nil {
			errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": %v", tribeID, err)
			return
		}
		tr, ok = tr0.(*ossrecordextended.OSSTribeExtended)
		if !ok {
			errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": invalid type %T", tribeID, tr0)
			return
		}
	}

	tribeData := &TribeData{}
	tribeData.populate(tr, &allRecords, loggedInUser, false, false)

	err := tribeTemplate.Execute(w, tribeData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

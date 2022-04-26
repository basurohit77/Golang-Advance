package main

import (
	"html/template"
	"net/http"
)

var segmentsTemplate *template.Template

// SegmentsData is used to pass all the data for the segments template
type SegmentsData struct {
	Segments     []*cachedSegment
	LastRefresh  string
	LoggedInUser string
}

func segmentsHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, false); !ok {
		return
	}

	segments, lastRefresh := allRecords.getAllSegments()
	if segments == nil {
		err := allRecords.refresh()
		if err != nil {
			errorPage(w, http.StatusInternalServerError, `Error refreshing OSS data from Global Catalog: %v`, err)
			return
		}
		segments, lastRefresh = allRecords.getAllSegments()
	}

	var segmentsData = SegmentsData{}
	segmentsData.LastRefresh = lastRefresh.UTC().Format(timeFormat)
	segmentsData.Segments = segments
	segmentsData.LoggedInUser = loggedInUser

	err := segmentsTemplate.Execute(w, &segmentsData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

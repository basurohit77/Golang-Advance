package main

// Functions for viewing differences between records in Staging and Production Catalog

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

var diffTemplate *template.Template

// DiffData is used to pass all the data for the diff template
type DiffData struct {
	EntryName      string
	EntryID        ossrecord.OSSEntryID
	Diffs          []*compare.OneDiff
	StagingOnly    bool
	ProductionOnly bool
	LoggedInUser   string // ID of the user, if logged-in
}

func (v *DiffData) populate(staging, production ossrecord.OSSEntry, loggedInUser string) {
	if staging != nil {
		v.EntryName = staging.String()
		v.EntryID = staging.GetOSSEntryID()
		if production != nil {
			out := &compare.Output{}
			compare.DeepCompare("", staging, "", production, out)
			v.Diffs = out.GetDiffs()
		} else {
			v.StagingOnly = true
			var dummy ossrecord.OSSEntry
			switch staging.(type) {
			case *ossrecord.OSSService:
				dummy = &ossrecord.OSSService{}
			case *ossrecord.OSSSegment:
				dummy = &ossrecord.OSSSegment{}
			case *ossrecord.OSSTribe:
				dummy = &ossrecord.OSSTribe{}
			case *ossrecord.OSSEnvironment:
				dummy = &ossrecord.OSSEnvironment{}
			default:
				panic(fmt.Sprintf("Unexpected entry type %T", staging))
			}
			out := &compare.Output{}
			compare.DeepCompare("", staging, "", dummy, out)
			v.Diffs = out.GetDiffs()
		}
	} else {
		v.ProductionOnly = true
		v.EntryName = production.String()
		v.EntryID = production.GetOSSEntryID()
		var dummy ossrecord.OSSEntry
		switch production.(type) {
		case *ossrecord.OSSService:
			dummy = &ossrecord.OSSService{}
		case *ossrecord.OSSSegment:
			dummy = &ossrecord.OSSSegment{}
		case *ossrecord.OSSTribe:
			dummy = &ossrecord.OSSTribe{}
		case *ossrecord.OSSEnvironment:
			dummy = &ossrecord.OSSEnvironment{}
		default:
			panic(fmt.Sprintf("Unexpected entry type %T", production))
		}
		out := &compare.Output{}
		compare.DeepCompare("", dummy, "", production, out)
		v.Diffs = out.GetDiffs()
	}
	v.LoggedInUser = loggedInUser
}

// diffHandler is the HTTP handler for the /diff page
func diffHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, false); !ok {
		return
	}

	var prefix = "/diff/"
	entryID := ossrecord.OSSEntryID(strings.TrimPrefix(r.URL.Path, prefix))
	if entryID == "" {
		errorPage(w, http.StatusNotFound, `Empty entry-id. Use the URI "/diff/<entry-id>" to view all diffs for one entry`)
		return
	}

	staging, err := catalog.ReadOSSEntryByID(entryID, catalog.IncludeNone)
	if rest.IsEntryNotFound(err) {
		staging = nil
	} else if err != nil {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog Staging entry for \"%s\": %v", entryID, err)
		return
	}

	production, err := catalog.ReadOSSEntryByIDProduction(entryID, catalog.IncludeNone)
	if rest.IsEntryNotFound(err) {
		production = nil
	} else if err != nil {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog Production entry for \"%s\": %v", entryID, err)
		return
	}

	if staging == nil && production == nil {
		errorPage(w, http.StatusNotFound, "Entry \"%s\" not found in Staging or Production Catalog", entryID)
		return
	}

	diffData := &DiffData{}
	diffData.populate(staging, production, loggedInUser)

	err = diffTemplate.Execute(w, diffData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

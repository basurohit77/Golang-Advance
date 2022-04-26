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

var segmentTemplate *template.Template

// SegmentData is used to pass all the data for the segment template
type SegmentData struct {
	*ossrecordextended.OSSSegmentExtended        // embedded type
	Created                               string // Creation time of the OSS entry in Catalog
	Updated                               string // Last update time of the OSS entry in Catalog
	Tribes                                []*cachedTribe
	LastRefresh                           string
	LinkToOSSRecord                       string
	ProductionDiffID                      ossrecord.CatalogID
	LoggedInUser                          string // ID of the user, if logged-in
}

func (v *SegmentData) populate(seg *ossrecordextended.OSSSegmentExtended, cache *AllRecordsType, loggedInUser string, updateEnabled, editMode bool) {
	v.OSSSegmentExtended = seg
	v.Created = seg.Created
	v.Updated = seg.Updated
	var lastRefresh time.Time
	v.Tribes, lastRefresh = cache.getTribesForSegmentID(seg.SegmentID)
	v.LastRefresh = lastRefresh.UTC().Format(timeFormat)
	v.LinkToOSSRecord, _ = catalog.GetOSSEntryUI(seg)
	v.ProductionDiffID = ossrecord.CatalogID(seg.GetOSSEntryID())
	v.LoggedInUser = loggedInUser
	return
}

// segmentHandler is the HTTP handler for the /segment page
func segmentHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, false); !ok {
		return
	}

	segmentID := strings.TrimPrefix(r.URL.Path, "/segment/")
	if segmentID == "" {
		errorPage(w, http.StatusNotFound, `Empty segmentID. Use the URI "/segment/<segment-id>" to view information about <segment-id>`)
		return
	}

	seg := allRecords.getPlaceHolderSegment(ossrecord.SegmentID(segmentID))
	if seg == nil {
		var err error
		seg0, err := catalog.ReadOSSEntryByID(ossrecord.MakeOSSSegmentID(ossrecord.SegmentID(segmentID)), catalog.IncludeTribes|catalog.IncludeAll)
		if err != nil {
			errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": %v", segmentID, err)
			return
		}
		seg, ok = seg0.(*ossrecordextended.OSSSegmentExtended)
		if !ok {
			errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": invalid type %T", segmentID, seg0)
			return
		}
	}

	segmentData := &SegmentData{}
	segmentData.populate(seg, &allRecords, loggedInUser, false, false)

	err := segmentTemplate.Execute(w, segmentData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

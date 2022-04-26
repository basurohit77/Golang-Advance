package main

// Functions for viewing and editing OSS service/component records

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/clearinghouse"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

var viewTemplate *template.Template

// ViewData is used to pass all the data for the view template
type ViewData struct {
	Main                    *ossrecord.OSSService
	GeneralInfo             []Element
	Ownership               []Element
	Support                 []Element
	Operations              []Element
	StatusPage              []Element
	Compliance              []Element
	ServiceNowInfo          []Element
	CatalogInfo             []Element
	ProductInfo             []Element
	Taxonomy                []Element
	OSSMergeControl         *ossmergecontrol.OSSMergeControl
	MergeControlTags        string // Use a separate string to avoid automatic quoting of expiration dates in template library
	IgnoredMergeControlTags string // Use a separate string to avoid automatic quoting of expiration dates in template library
	OSSValidation           *ossvalidation.OSSValidation
	LinkToOSSRecord         string
	LinkToMainRecord        string
	LinkToCHRecords         []struct {
		Name string
		Link string
		Tags []string
	}
	ProductionDiffID  ossrecord.CatalogID
	TotalIssues       int
	IssuesCounts      map[ossvalidation.Severity]int
	IssuesLabels      []ossvalidation.Severity
	CatalogVisibility string
	Children          []ossrecord.CRNServiceName
	LastRefresh       string
	Created           string // Creation time of the OSS entry in Catalog
	Updated           string // Last update time of the OSS entry in Catalog
	UpdateEnabled     bool   // true if this record can be updated by the client (authorized)
	EditMode          bool   // true if we are currently in edit mode (show form inputs instead of plain fields for editable items)
	SignatureWarning  string // if not empty, this is a message indicating that the OSSMergeControl.LastUpdate and OSSMergeControl.UpdatedBy might not be valid
	DeleteWarning     string // if not empty this is a message indicating that this record might be deleted on the next merge
	LoggedInUser      string // ID of the user, if logged-in
}

func (v *ViewData) populate(ossrec *ossrecordextended.OSSServiceExtended, cache *AllRecordsType, loggedInUser string, updateEnabled, editMode bool) {
	v.Main = &ossrec.OSSService
	v.GeneralInfo = NewStructIterator(&ossrec.OSSService.GeneralInfo).Slice()
	v.Ownership = NewStructIterator(&ossrec.OSSService.Ownership).Slice()
	v.Support = NewStructIterator(&ossrec.OSSService.Support).Slice()
	v.Operations = NewStructIterator(&ossrec.OSSService.Operations).Slice()
	v.StatusPage = NewStructIterator(&ossrec.OSSService.StatusPage).Slice()
	v.Compliance = NewStructIterator(&ossrec.OSSService.Compliance).Slice()
	v.ServiceNowInfo = NewStructIterator(&ossrec.OSSService.ServiceNowInfo).Slice()
	v.CatalogInfo = NewStructIterator(&ossrec.OSSService.CatalogInfo).Slice()
	v.ProductInfo = NewStructIterator(&ossrec.OSSService.ProductInfo).Slice()
	v.Taxonomy = NewStructIterator(&ossrec.OSSService.ProductInfo.Taxonomy).Slice()
	v.OSSMergeControl = ossrec.OSSMergeControl
	if tags := ossrec.OSSMergeControl.OSSTags.WithoutStatus(); len(*tags) > 0 {
		buf := strings.Builder{}
		buf.WriteString("[")
		for i, t := range *tags {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(`"`)
			buf.WriteString(string(t))
			buf.WriteString(`"`)
		}
		buf.WriteString("]")
		v.MergeControlTags = buf.String()
	} else {
		v.MergeControlTags = "[]"
	}
	if ignoredTags := ossrec.OSSMergeControl.IgnoredOSSTags.WithoutStatus(); len(*ignoredTags) > 0 {
		buf := strings.Builder{}
		buf.WriteString("[")
		for i, t := range *ignoredTags {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(`"`)
			buf.WriteString(string(t))
			buf.WriteString(`"`)
		}
		buf.WriteString("]")
		v.IgnoredMergeControlTags = buf.String()
	} else {
		v.IgnoredMergeControlTags = ""
	}
	v.OSSValidation = ossrec.OSSValidation
	v.LinkToOSSRecord, _ = catalog.GetOSSEntryUI(ossrec)
	v.LinkToMainRecord, _ = catalog.GetMainCatalogEntryUI(ossrec.OSSService.ReferenceCatalogID)
	chrefs := ossrec.OSSService.ProductInfo.ClearingHouseReferences
	for i := range chrefs {
		link := clearinghouse.GetCHEntryUI(clearinghouse.DeliverableID(chrefs[i].ID))
		v.LinkToCHRecords = append(v.LinkToCHRecords, struct {
			Name string
			Link string
			Tags []string
		}{chrefs[i].Name, link, chrefs[i].Tags})
	}
	v.ProductionDiffID = ossrecord.CatalogID(ossrec.GetOSSEntryID())
	v.IssuesCounts = ossrec.OSSValidation.CountIssues(nil)
	v.IssuesLabels = ossvalidation.ActionableSeverityList()
	v.TotalIssues = v.IssuesCounts["TOTAL"]
	if ossrec.OSSValidation.CatalogVisibility.EffectiveRestrictions != "" {
		v.CatalogVisibility = fmt.Sprintf("%+v", ossrec.OSSValidation.CatalogVisibility)
	} else {
		v.CatalogVisibility = ""
	}
	var lastRefresh time.Time
	v.Children, lastRefresh = cache.getServiceChildren(ossrec.OSSService.ReferenceResourceName)
	v.LastRefresh = lastRefresh.UTC().Format(timeFormat)
	v.Created = ossrec.Created
	v.Updated = ossrec.Updated
	v.UpdateEnabled = updateEnabled
	v.LoggedInUser = loggedInUser
	v.EditMode = editMode
	if !ossrec.OSSMergeControl.RefreshChecksum(false) {
		v.SignatureWarning = "WARNING: this record may have been updated outside the osscatviewer tool"
	} else {
		v.SignatureWarning = ""
	}
	if ossrec.IsDeletable() {
		v.DeleteWarning = "This record is empty and may be deleted after the next merge"
	} else {
		v.DeleteWarning = ""
	}

	// "Doctor" the ServiceNow CI link to always point to Production
	ciLink := ossrec.OSSService.GeneralInfo.ServiceNowCIURL
	if ciLink != "" {
		ciLink = strings.Replace(ciLink, "watsondev.service-now.com", "watson.service-now.com", 1)
		ciLink = strings.Replace(ciLink, "watsontest.service-now.com", "watson.service-now.com", 1)
		ossrec.OSSService.GeneralInfo.ServiceNowCIURL = ciLink
	}
	return
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	viewOrEditHandler(w, r, false)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	viewOrEditHandler(w, r, true)
}

// viewOrEditHandler is a common body for the /view and /edit HTTP handlers
func viewOrEditHandler(w http.ResponseWriter, r *http.Request, editMode bool) {
	var loggedInUser string
	var updateEnabled bool
	var ok bool
	if loggedInUser, updateEnabled, ok = checkRequest(w, r, editMode); !ok {
		return
	}

	var prefix string
	if editMode {
		prefix = "/edit/"
	} else {
		prefix = "/view/"
	}
	serviceName := strings.TrimPrefix(r.URL.Path, prefix)
	if serviceName == "" {
		errorPage(w, http.StatusNotFound, `Empty serviceName. Use the URI "/view/<service-name>" to view information about <service-name>`)
		return
	}

	ossrec0, err := catalog.ReadOSSEntryByID(ossrecord.MakeOSSServiceID(ossrecord.CRNServiceName(serviceName)), catalog.IncludeServices|catalog.IncludeAll)
	if err != nil {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": %v", serviceName, err)
		return
	}
	ossrec, ok := ossrec0.(*ossrecordextended.OSSServiceExtended)
	if !ok {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": invalid type %T", serviceName, ossrec0)
		return
	}

	viewData := &ViewData{}
	viewData.populate(ossrec, &allRecords, loggedInUser, updateEnabled, editMode)

	err = viewTemplate.Execute(w, viewData)
	if err != nil {
		errorPage(w, http.StatusInternalServerError, "Error executing template: %v", err)
		return
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser string
	var ok bool
	if loggedInUser, _, ok = checkRequest(w, r, true); !ok {
		return
	}

	if r.Method != "POST" {
		errorPage(w, http.StatusBadRequest, `/update/ only supports POST`)
		return
	}

	serviceName := strings.TrimPrefix(r.URL.Path, "/update/")
	if serviceName == "" {
		errorPage(w, http.StatusNotFound, `Empty serviceName. Use the URI "/update/<service-name>" to update information about <service-name>`)
		return
	}

	ossrec0, err := catalog.ReadOSSEntryByID(ossrecord.MakeOSSServiceID(ossrecord.CRNServiceName(serviceName)), catalog.IncludeServices|catalog.IncludeAll)
	if err != nil {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": %v", serviceName, err)
		return
	}
	ossrec, ok := ossrec0.(*ossrecordextended.OSSServiceExtended)
	if !ok {
		errorPage(w, http.StatusNotFound, "Error reading OSS Catalog entry for \"%s\": invalid type %T", serviceName, ossrec0)
		return
	}

	err = r.ParseForm()
	if err != nil {
		errorPage(w, http.StatusNotFound, `Error parsing /update/ form for "%s": %v`, serviceName, err)
		return
	}

	prior := ossrec.OSSMergeControl.OneLineString()
	var inputErrors strings.Builder
	var inputOSSTags = osstags.TagSet{}
	var inputOverrides = make(map[string]string)
	readOneFormField(r.PostForm, "osstags", &inputOSSTags, &inputErrors)
	readOneFormField(r.PostForm, "rawduplicatenames", &ossrec.OSSMergeControl.RawDuplicateNames, &inputErrors)
	readOneFormField(r.PostForm, "donotmergenames", &ossrec.OSSMergeControl.DoNotMergeNames, &inputErrors)
	readOneFormField(r.PostForm, "overrides", &inputOverrides, &inputErrors)
	readOneFormField(r.PostForm, "notes", &ossrec.OSSMergeControl.Notes, &inputErrors)
	// TODO: Should combine the tag parsing logic with the validation/sorting that occurs when saving in catalogoss.UpdateOSSEntry
	err = inputOSSTags.Validate(false)
	if err != nil {
		inputErrors.WriteString(fmt.Sprintf("%v\n", err))
	}
	var oldOverrides = ossrec.OSSMergeControl.Overrides
	ossrec.OSSMergeControl.Overrides = make(map[string]interface{})
	for on, ov := range inputOverrides {
		err := ossrec.OSSMergeControl.AddOverride(on, ov)
		if err != nil {
			inputErrors.WriteString(fmt.Sprintf("Invalid Override: \"%s\"\n", err))
			ossrec.OSSMergeControl.Overrides = nil
		}
	}
	if inputErrors.Len() > 0 {
		ossrec.OSSMergeControl.Overrides = oldOverrides
		errorPage(w, http.StatusBadRequest, "Error parsing inputs for /update/ form for \"%s\"\n\n%v", serviceName, inputErrors.String())
		return
	}
	ossrec.OSSMergeControl.OSSTags = inputOSSTags
	ossrec.OSSMergeControl.UpdatedBy = loggedInUser
	ossrec.OSSMergeControl.LastUpdate = time.Now().UTC().Format("2006-01-02T15:04.000Z")
	ossrec.OSSMergeControl.RefreshChecksum(true)

	err = debug.AuditWithOptions(serviceName, "UPDATING: User %s updating OSS Catalog entry for \"%s\": \n  old=%s\n  new=%s.", loggedInUser, serviceName, prior, ossrec.OSSMergeControl.OneLineString())
	if err != nil {
		errorPage(w, http.StatusInternalServerError, `Cannot update OSS Catalog entry for "%s" because the update cannot be logged in LogDNA: %v`, serviceName, err)
		return
	}

	err = catalog.UpdateOSSEntry(ossrec, catalog.IncludeAll)
	if err != nil {
		errorPage(w, http.StatusNotFound, `Error updating OSS Catalog entry for "%s": %v`, serviceName, err)
		return
	}
	debug.Info("UPDATED: Successfully updated OSS Catalog entry for \"%s\": \n  old=%s\n  new=%s.", serviceName, prior, ossrec.OSSMergeControl.OneLineString())
	http.Redirect(w, r, fmt.Sprintf("/view/%s", serviceName), http.StatusSeeOther)
}

// EmitJSONShort generates a short (one line) string containing the JSON representation of its argument
func (v *ViewData) EmitJSONShort(val interface{}) string {
	var result string
	switch val0 := val.(type) {
	case string:
		if len(val0) == 0 {
			return ""
		}
		switch val0[0] {
		case '{', '[', '*':
			return fmt.Sprintf(`"%s"`, val0)
		default:
			return val0
		}
	case []string:
		if len(val0) == 0 {
			return "[]"
		}
	case osstags.TagSet:
		if len(val0) == 0 {
			return "[]"
		}
	case map[string]string:
		if len(val0) == 0 {
			return "{}"
		}
	}
	buf, err := json.Marshal(val)
	if err == nil {
		result = string(buf)
	} else {
		result = err.Error()
	}
	return result
}

// EmitJSONLong generates a long (multi line) string containing the JSON representation of its argument
func (v *ViewData) EmitJSONLong(val interface{}) string {
	var result string
	switch val0 := val.(type) {
	case string:
		if len(val0) == 0 {
			return ""
		}
		switch val0[0] {
		case '{', '[', '*':
			return fmt.Sprintf(`"%s"`, val0)
		default:
			return val0
		}
	case []string:
		if len(val0) == 0 {
			return "[]"
		}
	case osstags.TagSet:
		if len(val0) == 0 {
			return "[]"
		}
	case map[string]string:
		if len(val0) == 0 {
			return "{}"
		}
	}
	buf, err := json.MarshalIndent(val, "", "")
	if err == nil {
		result = string(buf)
	} else {
		result = err.Error()
	}
	return result
}

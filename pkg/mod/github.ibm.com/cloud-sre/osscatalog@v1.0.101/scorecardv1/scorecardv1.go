package scorecardv1

//TODO: Decide if using offical ScorecardV1 API or internal API

import (
	"fmt"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// Functions for accessing entries in Doctor ScorecardV1

// segmentsGet represents the response from GET calls in ScorecardV1 to obtain Segment information (in JSON)
type segmentsGet struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
	Count  int64 `json:"count"`
	First  struct {
		Href string `json:"href"`
	} `json:"first"`
	Last struct {
		Href string `json:"href"`
	} `json:"last"`
	Prev struct {
		Href string `json:"href"`
	} `json:"prev"`
	Next struct {
		Href string `json:"href"`
	} `json:"next"`
	Resources []SegmentResource `json:"resources"`
}

// SegmentResource represents the info about one Segment in ScorecardV1 (in JSON)
type SegmentResource struct {
	Href      string `json:"href"`
	Name      string `json:"name"`
	MgmtName  string `json:"mgmtName"`
	MgmtEmail string `json:"mgmtEmail"`
	TechName  string `json:"techName"`
	TechEmail string `json:"techEmail"`
	Tribes    struct {
		Href string `json:"href"`
	} `json:"tribes"`
}

// GetSegmentID returns the SegmentID associated with this SegmentResource object from ScorecardV1
func (seg *SegmentResource) GetSegmentID() ossrecord.SegmentID {
	ix := strings.LastIndex(seg.Href, "/segments/")
	var id ossrecord.SegmentID
	if ix <= 0 {
		id = ""
	} else {
		id = ossrecord.SegmentID(seg.Href[ix+10:]) // skip past "/segments/"
	}
	return id
}

// tribeGet represents the response from a GET call in ScorecardV1 to obtain all the Tribes for one Segment (in JSON)
type tribeGet struct {
	Href      string          `json:"href"`
	Resources []TribeResource `json:"resources"`
}

// TribeResource represents the info about one Tribe in ScorecardV1 (in JSON)
type TribeResource struct {
	Href         string `json:"href"`
	Name         string `json:"name"`
	OwnerContact string `json:"ownerContact"`
	OwnerEmail   string `json:"ownerEmail"`
	Segment      struct {
		Name string `json:"name"`
		Href string `json:"href"`
	} `json:"segment"`
	Services struct {
		Href string `json:"href"`
	} `json:"services"`
	ChangeApprovers []struct {
		Member ossrecord.Person `json:"member"`
		Tags   []string         `json:"tags"`
	} `json:"change_approvers"`
}

var tribeIDPattern = regexp.MustCompile(`/segments/([^/]+)/tribes/([^/]+)$`)

// GetTribeID returns the TribeID associated with this SegmentResource object from ScorecardV1
func (tr *TribeResource) GetTribeID() ossrecord.TribeID {
	var id ossrecord.TribeID
	matches := tribeIDPattern.FindStringSubmatch(tr.Href)
	if matches == nil {
		id = "" // Error
	} else {
		id = ossrecord.TribeID(fmt.Sprintf("%s-%s", matches[1], matches[2]))
	}
	return id
}

// scorecardv1EntryGet represents the response from GET calls in ScorecardV1 to obtain service entries (in JSON)
type scorecardv1EntryGet struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
	Count  int64 `json:"count"`
	Next   struct {
		Href string `json:"href"`
	} `json:"next"`
	Resources []ExternalEntry `json:"resources"`
}

// ExternalEntry represents the record for each service/component in Doctor ScorecardV1
type ExternalEntry struct {
	Href     string
	CRN      string
	Manifest struct {
		Href string
	}
	Name                    string
	TOC                     string
	MgmtEmail               string
	ExcludeReports          bool
	SecurityFocal           string
	TechEmail               string
	Status                  string
	CMMonitor               string
	CMMonitorByEnvironments []struct {
		Env     string
		Monitor string
	}
	CMMonitorIsSeparated       bool
	CMMonitorUsesProvisionData bool
	PMMonitor                  string
	PMMonitorByEnvironments    []struct {
		Env     string
		Monitor string
	}
	PMMonitorIsSeparated bool
	TOCBypass            bool
	EscalationPolicies   []struct {
		Href       string
		Name       string
		SystemType string
	}
	//	CentralRepositoryLocation string
	Segment struct {
		Name string
		Href string
	}
	Tribe struct {
		Name string
		Href string
	}
	ProductionReadiness struct {
		Href string
	}
	ProductionReadinessData *ProductionReadiness
}

/* Sample ExternalEntry as of 2018-11-27
       {
DEBUG:          "cmMonitor": null,
DEBUG:          "cmMonitorByEnvironment": [],
DEBUG:          "cmMonitorIsSeparated": null,
DEBUG:          "cmMonitorUsesProvisionData": null,
DEBUG:          "crn": "crn:v1:bluemix:public:databases-for-etcd:::::",
DEBUG:          "escalationPolicies": [],
DEBUG:          "excludeReports": null,
DEBUG:          "href": "https://api-oss.bluemix.net/scorecardv1/api/segmenttribe/v1/segments/58eda55b9babda00075a50d5/tribes/58eda5669babda00075a510c/services/5bef337c616ee3002b41db65",
DEBUG:          "manifest": {
DEBUG:             "href": null
DEBUG:          },
DEBUG:          "mgmtEmail": "Christopher Quinones",
DEBUG:          "name": "databases-for-etcd",
DEBUG:          "pmMonitor": null,
DEBUG:          "pmMonitorByEnvironment": [],
DEBUG:          "pmMonitorIsSeparated": null,
DEBUG:          "productionReadiness": {
DEBUG:             "href": "https://api-oss.bluemix.net/scorecardv1/api/segmenttribe/v1/segments/58eda55b9babda00075a50d5/tribes/58eda5669babda00075a510c/services/5bef337c616ee3002b41db65/production_readiness"
DEBUG:          },
DEBUG:          "securityFocal": "Khoi Dang/Costa Mesa/IBM",
DEBUG:          "segment": {
DEBUG:             "href": "https://api-oss.bluemix.net/scorecardv1/api/segmenttribe/v1/segments/58eda55b9babda00075a50d5",
DEBUG:             "name": "Watson Data Platform"
DEBUG:          },
DEBUG:          "status": "GA",
DEBUG:          "techEmail": "Ben Anderson",
DEBUG:          "toc": "",
DEBUG:          "tocBypass": null,
DEBUG:          "tribe": {
DEBUG:             "href": "https://api-oss.bluemix.net/scorecardv1/api/segmenttribe/v1/segments/58eda55b9babda00075a50d5/tribes/58eda5669babda00075a510c",
DEBUG:             "name": "Persistence Services"
DEBUG:          }
DEBUG:       },
*/

// scorecardv1ProductionReadinessGet represents the response from GET calls in ScorecardV1 to obtain Production Readiness info
type scorecardv1ProductionReadinessGet struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
	Count  int64 `json:"count"`
	Next   struct {
		Href string `json:"href"`
	} `json:"next"`
	Resources []ProductionReadiness `json:"resources"`
}

// ProductionReadiness represents the Production Readiness record for each service/component in Doctor ScorecardV1
// ("child" of the main service/component record)
// TODO: some of these sections are deprecated
type ProductionReadiness struct {
	ServiceInfo         string
	CRN                 string
	ProductionReadiness struct {
		Href                                string
		State                               string
		CentralizedVersionControlCompliance struct {
			State  string
			Issues []string
		}
		OnCallRotationCompliance struct {
			State  string
			Issues []string
		}
		PagerCompliance struct {
			State  string
			Issues []string
		}
		EscalationPolicyCompliance struct {
			State  string
			Issues []string
		}
		ReliabilityDesignCompliance struct {
			State  string
			Issues []string
		}
		BypassCompliance struct {
			State  string
			Issues []string
		}
		AVMEnabledCompliance struct {
			State  string
			Issues []string
		}
		RunbookEnabledCompliance struct {
			State  string
			Issues []string
		}
	}
}

// DetailEntry represents the record for one service/component in Doctor ScorecardV1, using the internal Doctor API
type DetailEntry struct {
	Name           string   `json:"name"`
	DisplayName    string   `json:"displayname"`
	CRN            string   `json:"crn"`
	ComponentNames []string `json:"componentNames"`
	// TODO: sopID is not a consistent JSON type in JSON output
	//	SOPID                            string   `json:"sopID"`
	Status                           string   `json:"status"`
	BusinessUnit                     string   `json:"businessUnit"`
	Tribe                            string   `json:"tribe"`
	MgmtContact                      string   `json:"mgmtContact"`
	MgmtContactEmail                 string   `json:"mgmtContactEmail"`
	TechContact                      string   `json:"techContact"`
	TechContactEmail                 string   `json:"techContactEmail"`
	TOC                              string   `json:"toc"`
	SupportPublicSlackChannel        string   `json:"supportPublicSlackChannel"`
	CentralRepositoryLocation        string   `json:"centralRepositoryLocation"`
	CentralizedVersionControl        bool     `json:"centralizedVersionControl"`
	CMMonitor                        string   `json:"cmMonitor"`
	PMMonitor                        string   `json:"pmMonitor"`
	BCDRFocal                        string   `json:"bcdrfocal"`
	SOPSecurityFocal                 string   `json:"sopSecurityFocal"`
	BypassPRC                        string   `json:"bypassPRC"`
	EUAccessEmergencyUSAMServiceName string   `json:"euaccess_emerg_usam_service_name"`
	AVMEnabled                       bool     `json:"avmEnabled"`
	IsServiceNowOnboarded            bool     `json:"isServicenowOnboarded"`
	TIPOnboarded                     bool     `json:"tipOnboarded"`
	RunbookEnabled                   bool     `json:"runbookEnabled"`
	ManuallyTOC                      bool     `json:"manuallyToc"`
	ReliabilityDesignReview          bool     `json:"reliabilityDesignReview"`
	OnCallRotation                   bool     `json:"oncallRotation"`
	PagerDuty                        []string `json:"pager_duty"`
	PagerDutyDetails                 []struct {
		PDType string `json:"type"`
		URL    string `json:"url"`
		ID     string `json:"id"`
		Name   string `json:"name"`
	} `json:"pager_duty_details"`
	BaileyID                 string   `json:"baileyID"`
	BaileyProject            string   `json:"bailey_project"`
	BaileyURL                string   `json:"bailey_url"`
	BypassSupportCompliances string   `json:"bypassSupportCompliances"`
	CertificateManagerCRNs   []string `json:"cm_crns"`
	SupportCompliancesOSS007 bool     `json:"supportCompliances_OSS007"`
}

// scorecardv1DetailGet represents the response from GET calls in ScorecardV1 to obtain Production Readiness info
type scorecardv1DetailGet struct {
	Result string        `json:"result"`
	Detail []DetailEntry `json:"detail"`
}

// ReadScorecardV1Record reads one service entry through the ScorecardV1 API, given its name
func ReadScorecardV1Record(name ossrecord.CRNServiceName, productionReadiness bool) (*ExternalEntry, error) {
	actualURL := fmt.Sprintf(scorecardv1ServiceEntryURL, string(name))
	key, err := rest.GetKey(scorecardv1KeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for ScorecardV1")
		return nil, err
	}
	var result = new(scorecardv1EntryGet)
	err = rest.DoHTTPGet(actualURL, key, nil, "ScorecardV1", debug.ScorecardV1, result)
	if err != nil {
		return nil, err
	}

	switch len(result.Resources) {
	case 1:
		// good
	case 0:
		err = rest.MakeHTTPError(nil, nil, true, `ScorecardV1 entry "%s" not found`, name)
		return nil, err
	default:
		err = fmt.Errorf("ScorecardV1 GET: expected 1 entry got %d  (URL=%s)", len(result.Resources), actualURL)
		return nil, err
	}

	record := &result.Resources[0]
	if productionReadiness {
		prURL := record.ProductionReadiness.Href
		if prURL == "" {
			err = fmt.Errorf("ScorecardV1 GET: ProductionReadiness.Href is empty  (URL=%s)", actualURL)
			return nil, err
		}
		prGet := new(scorecardv1ProductionReadinessGet)
		err = rest.DoHTTPGet(prURL, key, nil, "ScorecardV1ProductionReadiness", debug.ScorecardV1, prGet)
		if err != nil {
			return nil, err
		}
		if prGet.Resources != nil {
			if len(prGet.Resources) != 1 {
				err = fmt.Errorf("ScorecardV1 GET: ProductionReadiness expected 1 entry got %d  (URL=%s)", len(prGet.Resources), prURL)
				return nil, err
			}
			debug.Debug(debug.ScorecardV1, "Got Production Readiness for \"%s\"", string(name))
			record.ProductionReadinessData = &prGet.Resources[0]
		} else {
			debug.Debug(debug.ScorecardV1, "Got no Production Readiness for \"%s\"", string(name))
		}
	}
	return record, nil
}

// ReadScorecardV1Detail reads one service entry through the internal Doctor API, given its name
func ReadScorecardV1Detail(name ossrecord.CRNServiceName) (*DetailEntry, error) {
	// FIXME: Need a ScorecardV1 API to fetch a single entry
	// The current authenticated "external" APIU returns all entries, and the "internal" API is disabled
	// See issue https://github.ibm.com/cloud-sre/osscatalog/issues/128
	if true {
		pattern, err := regexp.Compile(regexp.QuoteMeta(string(name)))
		if err != nil {
			return nil, debug.WrapError(err, `ReadScorecardV1Detail() Cannot compile a regex pattern for "%s"`, name)
		}
		var results []*DetailEntry
		err = ListScorecardV1Details(pattern, func(e *DetailEntry) {
			results = append(results, e)
		})
		if err != nil {
			return nil, debug.WrapError(err, `ReadScorecardV1Detail(): ListScorecardV1Details returned error for "%s"`, name)
		}
		switch len(results) {
		case 1:
			return results[0], nil
		case 0:
			err = rest.MakeHTTPError(nil, nil, true, `ScorecardV1 entry "%s" not found (using ListScorecardV1Details)`, name)
			return nil, err
		default:
			err = fmt.Errorf(`ReadScorecardV1Detail: expected 1 entry for "%s" got %d (using ListScorecardV1Details)`, name, len(results))
			return nil, err
		}
	} else {
		actualURL := fmt.Sprintf(scorecardv1ServiceDetailURL, string(name))

		key, err := rest.GetKey(scorecardv1KeyName)
		if err != nil {
			err = debug.WrapError(err, "Cannot get key for ScorecardV1")
			return nil, err
		}

		if strings.HasPrefix(actualURL, "https://doctor") {
			// FIXME: remove the SetDisableTLSVerify once ScorecardV1 URL certif is fixed.
			// XXX Note that this is not thread-safe
			oldTLS := rest.SetDisableTLSVerify(true)
			defer rest.SetDisableTLSVerify(oldTLS)
		}

		var result = new(scorecardv1DetailGet)
		err = rest.DoHTTPGet(actualURL, key, nil, "ScorecardV1", debug.ScorecardV1, result)
		if err != nil {
			return nil, err
		}
		if result.Result != "success" {
			err = fmt.Errorf("ScorecardV1 GET: result=%s  (URL=%s)", result.Result, actualURL)
			return nil, err
		}

		switch len(result.Detail) {
		case 1:
			// good
		case 0:
			err = rest.MakeHTTPError(nil, nil, true, `ScorecardV1 entry "%s" not found`, name)
			return nil, err
		default:
			err = fmt.Errorf("ScorecardV1 GET: expected 1 entry got %d  (URL=%s)", len(result.Detail), actualURL)
			return nil, err
		}

		record := &result.Detail[0]

		return record, nil
	}
}

// ListSegments lists all the Segments from ScorecardV1 and calls the special handler function for each entry
func ListSegments(handler func(e *SegmentResource) error) error {
	var numEntries = 0

	actualURL := scorecardv1SegmentsURL

	if strings.HasPrefix(actualURL, "https://doctor") {
		// FIXME: remove the SetDisableTLSVerify once ScorecardV1 URL certif is fixed.
		// XXX Note that this is not thread-safe
		oldTLS := rest.SetDisableTLSVerify(true)
		defer rest.SetDisableTLSVerify(oldTLS)
	}

	key, err := rest.GetKey(scorecardv1KeyName)
	//key, err := rest.GetToken(scorecardv1KeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for ScorecardV1")
		return err
	}

	for {
		var result = new(segmentsGet)
		err = rest.DoHTTPGet(actualURL, key, nil, "ScorecardV1.Segments", debug.ScorecardV1, result)
		if err != nil {
			return err
		}
		if len(result.Resources) == 0 {
			break
		}
		for i := 0; i < len(result.Resources); i++ {
			segment := &result.Resources[i]
			if segment.Name == "" {
				debug.PrintError("scorecardv1.ListSegments(): ignoring segment with empty name: %v#", segment)
				continue
			}
			err2 := handler(segment)
			if err2 != nil {
				return debug.WrapError(err2, "Aborting scorecardv1.ListSegments()")
			}
			numEntries++
		}
		// TODO: check if we should use the provided OSS Platform API URL, or substitute a direct Doctor API URL
		actualURL = result.Next.Href
		if actualURL == "" {
			break
		}
	}

	debug.Debug(debug.ScorecardV1, "ListSegments() loaded %d entries", numEntries)
	return nil
}

// ListTribes lists all the Tribes for one Segment from ScorecardV1 and calls the special handler function for each entry
func ListTribes(segment *SegmentResource, handler func(e *TribeResource) error) error {
	var numEntries = 0

	actualURL := segment.Tribes.Href
	// Special substitution to use the direct Doctor API instead of the OSS Platform API
	if scorecardv1TribesURL != "" {
		tempURL := scorecardv1TribesURL // to avoid a compiler warning about missing format specs in the Sprintf if scorecardv1TribesURL is empty
		actualURL = fmt.Sprintf(tempURL, segment.GetSegmentID())
	}

	if strings.HasPrefix(actualURL, "https://doctor") {
		// FIXME: remove the SetDisableTLSVerify once ScorecardV1 URL certif is fixed.
		// XXX Note that this is not thread-safe
		oldTLS := rest.SetDisableTLSVerify(true)
		defer rest.SetDisableTLSVerify(oldTLS)
	}

	key, err := rest.GetKey(scorecardv1KeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for ScorecardV1")
		return err
	}

	var result = new(tribeGet)
	err = rest.DoHTTPGet(actualURL, key, nil, "ScorecardV1.Tribes", debug.ScorecardV1, result)
	if err != nil {
		return err
	}
	if len(result.Resources) == 0 {
		debug.PrintError("ListTribes(%s) found 0 Tribes in ScorecardV1", segment.Name) // XXX
		return nil
		//		return fmt.Errorf("ListTribes(%s) found 0 Tribes in ScorecardV1", segment.Name)
	}
	for i := 0; i < len(result.Resources); i++ {
		tribe := &result.Resources[i]
		if tribe.Name == "" {
			debug.PrintError("scorecardv1.ListSegments(): ignoring tribe with empty name for Segment(%s): %v#", segment.Name, tribe)
			continue
		}
		err2 := handler(tribe)
		if err2 != nil {
			return debug.WrapError(err2, "Aborting scorecardv1.ListTribes()")
		}
		numEntries++
	}

	debug.Debug(debug.ScorecardV1, "ListTribes(%s) loaded %d entries", segment.Name, numEntries)
	return nil
}

// ListScorecardV1Details lists all ScorecardV1 entries (using the internal API) and calls the special handler function for each entry
func ListScorecardV1Details(pattern *regexp.Regexp, handler func(e *DetailEntry)) error {
	rawEntries := 0
	totalEntries := 0

	actualURL := scorecardv1DetailListURL

	key, err := rest.GetKey(scorecardv1KeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for ScorecardV1")
		return err
	}

	if strings.HasPrefix(actualURL, "https://doctor") {
		// FIXME: remove the SetDisableTLSVerify once ScorecardV1 URL certif is fixed.
		// XXX Note that this is not thread-safe
		oldTLS := rest.SetDisableTLSVerify(true)
		defer rest.SetDisableTLSVerify(oldTLS)
	}

	var result = new(scorecardv1DetailGet)
	debug.Info("Loading one batch of entries from ScorecardV1 (%d/%d entries so far)", totalEntries, rawEntries)
	err = rest.DoHTTPGet(actualURL, key, nil, "ScorecardV1", debug.ScorecardV1, result)
	if err != nil {
		return err
	}

	if result.Result != "success" {
		err = fmt.Errorf("ScorecardV1 GET: result=%s  (URL=%s)", result.Result, actualURL)
		return err
	}

	for i := 0; i < len(result.Detail); i++ {
		rawEntries++
		if (rawEntries % 30) == 0 {
			debug.Info("Loading one batch of entries from ScorecardV1 (%d/%d entries so far)", totalEntries, rawEntries)
		}
		e := &result.Detail[i]
		if pattern.FindString(e.Name) == "" {
			continue
		}
		handler(e)
		totalEntries++
	}

	debug.Info("Read %d entries from ScorecardV1", totalEntries)
	return nil
}

// String returns a short string representation of this ScorecardV1 DetailEntry record
func (r *DetailEntry) String() string {
	return fmt.Sprintf(`ScorecardV1Detail{Name:"%s", Status:"%s"}`, r.Name, r.Status)
}

var _ fmt.Stringer = &DetailEntry{}

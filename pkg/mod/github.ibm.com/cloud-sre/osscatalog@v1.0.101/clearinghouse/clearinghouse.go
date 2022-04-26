package clearinghouse

import (
	"fmt"
	"net/http"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

var countClearingHouseAPICalls int

// CHDeliverable represents one Deliverable record in ClearingHouse (without dependency information)
type CHDeliverable struct {
	CloudServiceType     string        `json:"cloud_service_type"`
	Created              string        `json:"created"`
	CreatedBy            string        `json:"created_by"`
	DeliveryType         string        `json:"delivery_type"`
	DeploymentTarget     string        `json:"deployment_target"`
	FixPacks             []interface{} `json:"fix_packs"`
	FreeOrTrial          bool          `json:"free_or_trial"`
	GbmsID               string        `json:"gbms_id"`
	ID                   string        `json:"id"`
	ModLevels            []interface{} `json:"mod_levels"`
	Name                 string        `json:"name"`
	Next                 []interface{} `json:"next"`
	OfficialName         string        `json:"official_name"`
	Ownership            string        `json:"ownership"`
	Phase                string        `json:"phase"`
	PidNumber            []string      `json:"pid_number"`
	Previous             []interface{} `json:"previous"`
	ProcessType          string        `json:"process_type"`
	Published            bool          `json:"published"`
	ReleaseType          string        `json:"release_type"`
	Segment              string        `json:"segment"`
	ShortName            string        `json:"short_name"`
	SmartProtectLink     string        `json:"smart_protect_link"`
	Team                 string        `json:"team"`
	Updated              string        `json:"updated"`
	UpdatedBy            string        `json:"updated_by"`
	CRNServiceName       string        `json:"cloud_resource_name"`
	Utlevel10Code        string        `json:"utlevel10_code"`
	Utlevel10Description string        `json:"utlevel10_description"`
	Utlevel15Code        string        `json:"utlevel15_code"`
	Utlevel15Description string        `json:"utlevel15_description"`
	Utlevel17Code        string        `json:"utlevel17_code"`
	Utlevel17Description string        `json:"utlevel17_description"`
	Utlevel20Code        string        `json:"utlevel20_code"`
	Utlevel20Description string        `json:"utlevel20_description"`
	Utlevel30Code        string        `json:"utlevel30_code"`
	Utlevel30Description string        `json:"utlevel30_description"`
	UtlevelSetByPid      bool          `json:"utlevel_set_by_pid"`
	Version              string        `json:"version"`

	/*
				{
			        "id": "2EBB5860B34311E7A9EB066095601ABB",
		            "name": "IBM Cloudant Dedicated Cluster",
		            "version": "",
		            "code_name": "",
		            "short_name": "Cloudant Enterprise",
		            "official_name": "IBM Cloudant Dedicated Cluster",
		            "pid_number": [
		                "5725-S99"
		            ],
		            "ownership": "ibm",
		            "free_or_trial": false,
		            "phase": "Deliver",
		            "deployment_target": "Cloud",
		            "delivery_type": "Continuous",
		            "process_type": "OMOM",
		            "utlevel10_description": "Hybrid Cloud Dual",
		            "utlevel15_description": "Watson Data Platform",
		            "utlevel17_description": "Data and Streams",
		            "utlevel20_description": "Cloud Databases",
		            "utlevel10_code": "10A00",
		            "utlevel15_code": "15WDF",
		            "utlevel17_code": "17SAM",
		            "utlevel20_code": "20D03",
		            "utlevel30_code": "30DE7",
		            "utlevel30_description": "Cloudant NoSQLDB",
		            "utlevel_set_by_pid": true,
		            "segment": "Persistence Services",
		            "team": "CDS",
		            "published": false,
		            "smart_protect_link": "https://smartprotect.raleigh.ibm.com/Web/review.jsp?id=708",
		            "created": "2017-10-17 09:57:58.629",
		            "created_by": "Joshua.Mintz@ibm.com",
		            "updated": "2018-10-20 23:06:17.233",
		            "updated_by": "chaccess@us.ibm.com",
		            "release_type": "Product"
		        },
	*/

}

// CHDependency represents one dependency record from ClearingHouse
type CHDependency struct {
	CommitStatus      string `json:"commit_status"`
	DependencyID      string `json:"dependency_id"`
	DependencyType    string `json:"dependency_type"`
	OriginatorID      string `json:"originator_id"`
	OriginatorName    string `json:"originator_name"`
	OriginatorVersion string `json:"originator_version"`
	ProviderID        string `json:"provider_id"`
	ProviderName      string `json:"provider_name"`
	ProviderVersion   string `json:"provider_version"`
}

// CHDeliverableWithDependencies represents one Deliverable record in ClearingHouse (including dependency information)
type CHDeliverableWithDependencies struct {
	CloudServiceType      string        `json:"cloud_service_type"`
	Created               string        `json:"created"`
	CreatedBy             string        `json:"created_by"`
	DeliveryType          string        `json:"delivery_type"`
	DeploymentTarget      string        `json:"deployment_target"`
	FixPacks              []interface{} `json:"fix_packs"`
	FreeOrTrial           bool          `json:"free_or_trial"`
	GbmsID                string        `json:"gbms_id"`
	ID                    string        `json:"id"`
	ModLevels             []interface{} `json:"mod_levels"`
	Name                  string        `json:"name"`
	Next                  []interface{} `json:"next"`
	OfficialName          string        `json:"official_name"`
	Ownership             string        `json:"ownership"`
	Phase                 string        `json:"phase"`
	PidNumber             []string      `json:"pid_number"`
	Previous              []interface{} `json:"previous"`
	ProcessType           string        `json:"process_type"`
	Published             bool          `json:"published"`
	ReleaseType           string        `json:"release_type"`
	Segment               string        `json:"segment"`
	ShortName             string        `json:"short_name"`
	SmartProtectLink      string        `json:"smart_protect_link"`
	Team                  string        `json:"team"`
	Updated               string        `json:"updated"`
	UpdatedBy             string        `json:"updated_by"`
	CRNServiceName        string        `json:"cloud_resource_name"`
	Utlevel10Code         string        `json:"utlevel10_code"`
	Utlevel10Description  string        `json:"utlevel10_description"`
	Utlevel15Code         string        `json:"utlevel15_code"`
	Utlevel15Description  string        `json:"utlevel15_description"`
	Utlevel17Code         string        `json:"utlevel17_code"`
	Utlevel17Description  string        `json:"utlevel17_description"`
	Utlevel20Code         string        `json:"utlevel20_code"`
	Utlevel20Description  string        `json:"utlevel20_description"`
	Utlevel30Code         string        `json:"utlevel30_code"`
	Utlevel30Description  string        `json:"utlevel30_description"`
	UtlevelSetByPid       bool          `json:"utlevel_set_by_pid"`
	Version               string        `json:"version"`
	DependencyOriginators *struct {
		Dependencies []CHDependency `json:"dependencies"`
	} `json:"dependency_originators"`
	DependencyProviders *struct {
		Dependencies []CHDependency `json:"dependencies"`
	} `json:"dependency_providers"`
	// We copy the Taxonomy info into a struct, to simplify assignments and comparisons
	CopiedTaxonomy ossrecord.Taxonomy `json:"-"`

	/*
				{
		            "id": "2052E430379B11E58B2CB2A838CE4F20",
		            "name": "IBM Cloudant for IBM Cloud",
		            "version": "2018.04",
		            "short_name": "",
		            "official_name": "IBM Cloudant for IBM Cloud",
		            "pid_number": [
		                "5725-R48"
		            ],
		            "gbms_id": "159540",
		            "ownership": "ibm",
		            "free_or_trial": false,
		            "phase": "Deliver",
		            "deployment_target": "Cloud",
		            "delivery_type": "Continuous",
		            "cloud_service_type": "SaaS",
		            "previous": [],
		            "next": [],
		            "mod_levels": [],
		            "fix_packs": [],
		            "process_type": "OMOM",
		            "utlevel10_description": "Hybrid Cloud Dual",
		            "utlevel15_description": "Watson Data Platform",
		            "utlevel17_description": "Data and Streams",
		            "utlevel20_description": "Cloud Databases",
		            "utlevel10_code": "10A00",
		            "utlevel15_code": "15WDF",
		            "utlevel17_code": "17SAM",
		            "utlevel20_code": "20D03",
		            "utlevel30_code": "30DE7",
		            "utlevel30_description": "Cloudant NoSQLDB",
		            "utlevel_set_by_pid": true,
		            "segment": "Persistence Services",
		            "team": "CDS",
		            "published": false,
		            "smart_protect_link": "https://smartprotect.raleigh.ibm.com/Web/review.jsp?id=708",
		            "created": "2015-07-31 11:45:10",
		            "created_by": "stacybrn@us.ibm.com",
		            "updated": "2019-01-30 11:05:21",
		            "updated_by": "dbotros@us.ibm.com",
		            "release_type": "Product",
		            "dependency_originators": {
		                "dependencies": [
		                    {
		                        "dependency_id": "1456259951560",
		                        "dependency_type": "Usage",
		                        "commit_status": "Cancel",
		                        "provider_name": "IBM Cloudant for IBM Cloud",
		                        "provider_version": "2018.04",
		                        "provider_id": "2052E430379B11E58B2CB2A838CE4F20",
		                        "originator_name": "IBM Cloud eDiscovery",
		                        "originator_version": "1.0",
		                        "originator_id": "28F85A207F3F11E5B676827145285BB5"
							}
						],
					},
		            "dependency_providers": {
		                "dependencies": [
		                    {
		                        "dependency_id": "1438357728660",
		                        "dependency_type": "Usage",
		                        "commit_status": "Complete",
		                        "provider_name": "IBM Cloud Platform - Public",
		                        "provider_version": "- Continuous Delivery",
		                        "provider_id": "1380559625333",
		                        "originator_name": "IBM Cloudant for IBM Cloud",
		                        "originator_version": "2018.04",
		                        "originator_id": "2052E430379B11E58B2CB2A838CE4F20"
							},
						]
					}
				}

	*/
}

// MakeCHLabel creates a string label to represent a ClearingHouse entry in reports, logs and UI.
// Note that we intentionally picked "{" as the first character to ensure that this would sort
// alphabetically after all CRN service-names
func MakeCHLabel(chname string, chid DeliverableID) string {
	return (fmt.Sprintf(`{%s [chid:%s]}`, chname, chid))
}

var chLabelRegex = regexp.MustCompile(`^{([^\[]+?)\s*\[chid:(\w+)\]}$`)

// ParseCHLabel parses a ClearinHouse entry label generated by MakeCHLabel() and returns
// the entry name and ID
func ParseCHLabel(label string) (chname string, chid DeliverableID) {
	m := chLabelRegex.FindStringSubmatch(string(label))
	if m == nil {
		return "", ""
	}
	return m[1], DeliverableID(m[2])
}

// String returns a short string identifier for the CHDeliverable entry
func (ch *CHDeliverable) String() string {
	return MakeCHLabel(ch.Name, DeliverableID(ch.ID))
}

// String returns a short string identifier for the CHDeliverableWithDependencies entry
func (ch *CHDeliverableWithDependencies) String() string {
	return MakeCHLabel(ch.Name, DeliverableID(ch.ID))
}

// GetCHEntryUI returns a link (URL) for accessing the UI showing
// one particular ClearingHouse Deliverable entry, designated by its ID
func GetCHEntryUI(chid DeliverableID) string {
	return fmt.Sprintf(chDeliverablePageURL, string(chid))
}

// GetCHDependencyUI returns a link (URL) for accessing the UI showing
// one particular ClearingHouse Dependency entry, designated by its dependency ID
func GetCHDependencyUI(depid string) string {
	return fmt.Sprintf(chDependencyPageURL, depid)
}

// GetCHEntryUIFromLabel returns a link (URL) for accessing the UI showing one particular ClearingHouse entry,
// designated by its label
func GetCHEntryUIFromLabel(label string) string {
	_, chid := ParseCHLabel(label)
	if chid != "" {
		link := GetCHEntryUI(chid)
		return link
	}
	return ""
}

// SearchRecordsByName searches ClearingHouse records that match a given name
// TODO: consider caching results of SearchRecordsByName()
func SearchRecordsByName(name string) ([]*CHDeliverable, error) {
	countClearingHouseAPICalls++
	actualURL := fmt.Sprintf(chSearchURL, name)
	token, err := rest.GetToken(chTokenName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get token for ClearingHouse")
		return nil, err
	}
	clientID, err := rest.GetID(chTokenName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get client-id for ClearingHouse")
		return nil, err
	}
	var resultContainer struct {
		ProductReleases []*CHDeliverable `json:"product_releases"`
		Offset          float64          `json:"offset"`
		Limit           float64          `json:"limit"`
		TotalCount      float64          `json:"total_count"`
	}

	headers := make(http.Header)
	headers.Set("X-IBM-Client-Id", clientID)
	err = rest.DoHTTPGet(actualURL, "Bearer "+token, headers, "ClearingHouse", debug.ClearingHouse, &resultContainer)
	if err != nil {
		return nil, err
	}

	entries := resultContainer.ProductReleases

	return entries, nil
}

// GetFullRecordByID returns the full ClearingHouse record with a given DeliverableID, including dependency information
func GetFullRecordByID(id DeliverableID) (*CHDeliverableWithDependencies, error) {
	if cachedFullRecords == nil {
		resetFullRecordsCache()
	}
	if c, ok := cachedFullRecords[id]; ok {
		return c.entry, c.err
	}
	e, err := getFullRecordByIDNotCached(id)
	cachedFullRecords[id] = cacheEntry{e, err}
	return e, err
}

type cacheEntry struct {
	entry *CHDeliverableWithDependencies
	err   error
}

var cachedFullRecords map[DeliverableID]cacheEntry

// for testing
func resetFullRecordsCache() {
	cachedFullRecords = make(map[DeliverableID]cacheEntry)
}

// getFullRecordByIDNotCached - internal implementation of GetFullRecordByID
func getFullRecordByIDNotCached(id DeliverableID) (*CHDeliverableWithDependencies, error) {
	countClearingHouseAPICalls++
	actualURL := fmt.Sprintf(chByIDURL, id)
	token, err := rest.GetToken(chTokenName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get token for ClearingHouse")
		return nil, err
	}
	clientID, err := rest.GetID(chTokenName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get client-id for ClearingHouse")
		return nil, err
	}

	var resultContainer struct {
		ProductReleases []*CHDeliverableWithDependencies `json:"product_releases"`
		// ProductReleases []map[string]interface{} `json:"product_releases"`
	}

	headers := make(http.Header)
	headers.Set("X-IBM-Client-Id", clientID)
	err = rest.DoHTTPGet(actualURL, "Bearer "+token, headers, "ClearingHouse", debug.ClearingHouse, &resultContainer)
	if err != nil {
		return nil, err
	}

	if len(resultContainer.ProductReleases) != 1 {
		return nil, fmt.Errorf("GetRecordByDeliverableID(%s) expected 1 record got %d", id, len(resultContainer.ProductReleases))
	}
	if resultContainer.ProductReleases[0] == nil {
		return nil, fmt.Errorf("GetRecordByDeliverableID(%s) got a nil record: %v", id, resultContainer)
	}
	result := resultContainer.ProductReleases[0]
	result.CopiedTaxonomy.MajorUnitUTL10 = makeUTLString(result.Utlevel10Description, result.Utlevel10Code)
	result.CopiedTaxonomy.MinorUnitUTL15 = makeUTLString(result.Utlevel15Description, result.Utlevel15Code)
	result.CopiedTaxonomy.MarketUTL17 = makeUTLString(result.Utlevel17Description, result.Utlevel17Code)
	result.CopiedTaxonomy.PortfolioUTL20 = makeUTLString(result.Utlevel20Description, result.Utlevel20Code)
	result.CopiedTaxonomy.OfferingUTL30 = makeUTLString(result.Utlevel30Description, result.Utlevel30Code)
	return result, nil
}

// GetCountClearingHouseAPICalls returns the total number of calls made to the ClearingHouse API, since this process started
func GetCountClearingHouseAPICalls() int {
	return countClearingHouseAPICalls
}

func makeUTLString(desc, code string) string {
	if desc != "" || code != "" {
		return fmt.Sprintf("%s (%s)", desc, code)
	}
	return ""
}

package testDefs

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	hlp "github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

var incidentMgmtAPIPath = "/incidentmgmtapi/api/info"
var incidentMgmtAPIConcernPath = "/incidentmgmtapi/api/v1/incidentmgmt/concerns"

var catalogURL = os.Getenv("CATALOG_URL")
var snHost = os.Getenv("snHost")
var SnURL = "https://" + snHost + ".service-now.com"
var SnAPIGetURL = SnURL + "/api/ibmwc/v2/incident/getIncidents"
var SnAPIURL = SnURL + "/api/ibmwc/v2/incident"

type CatalogResponse struct {
	ClientID string `json:"clientId"`
	Concerns struct {
		Href string `json:"href"`
	} `json:"concerns"`
}

type SNRecord struct {
	Number                    string   `json:"number"`
	ShortDescription          string   `json:"short_description,omitempty"`
	CurrentStatus             string   `json:"u_current_status,omitempty"`
	DescriptionCustomerImpact string   `json:"u_description_customer_impact,omitempty"`
	SysID                     string   `json:"sys_id"`
	SysCreatedOn              string   `json:"sys_created_on"`
	AffectedActivity          string   `json:"u_affected_activity"`
	CmdbCi                    string   `json:"cmdb_ci"`
	UEnvironment              string   `json:"u_environment,omitempty"`
	Description               string   `json:"description,omitempty"`
	UStatus                   string   `json:"u_status,omitempty"`
	IncidentState             string   `json:"incident_state,omitempty"`
	SysUpdatedOn              string   `json:"sys_updated_on"`
	DisruptionBegan           string   `json:"u_disruption_began,omitempty"`
	DisruptionEnded           string   `json:"u_disruption_ended,omitempty"`
	UDetectionSource          string   `json:"u_detection_source,omitempty"`
	SysCreatedBy              string   `json:"sys_created_by"`
	SysUpdatedBy              string   `json:"sys_updated_by"`
	Severity                  string   `json:"priority"`
	UCRNs                     []string `json:"u_crn"`
	CRNs                      []string `json:"crn"`
	Process                   string
}

type SNRecordResult struct {
	Result SNRecord `json:"result"`
}
type SNRecords struct {
	Results []SNRecord `json:"results"`
}

type Crn struct {
	Version         string `json:"version"`
	Cname           string `json:"cname"`
	Ctype           string `json:"ctype"`
	Location        string `json:"location"`
	ServiceName     string `json:"service_name"`
	Scope           string `json:"scope"`
	ServiceInstance string `json:"service_instance"`
	Resource        string `json:"resource"`
}
type Extras struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type ConcernFormat struct {
	Crn               Crn      `json:"crn"`
	RunbookURL        string   `json:"runbook_url"`
	CustomerImpacting string   `json:"customer_impacting"`
	DisablePager      string   `json:"disable_pager"`
	Extras            []Extras `json:"extras"`
	Severity          int      `json:"severity"`
	ShortDescription  string   `json:"short_description"`
	LongDescription   string   `json:"long_description"`
	Source            string   `json:"source"`
	Journal           []string `json:"journal"`
	Timestamp         string   `json:"timestamp"`
	TipMsgType        string   `json:"tip_msg_type"`
	TribeName         string   `json:"tribe_name"`
	Version           string   `json:"version"`
}
type ConcernResponse struct {
	IncidentID string `json:"incidentId"`
	UIURL      string `json:"uiUrl"`
}

type IncidentMgmtApiIncident struct {
	AffectedActivity string `json:"affected_activity"`
	AlertID          string `json:"alert_id"`
	Situation        string `json:"situation"`
	AssignedTo       string `json:"assigned_to"`
	ClosedBy         string `json:"closed_by"`
	CloseCode        string `json:"close_code"`
	CloseNotes       string `json:"close_notes"`
	CorrelationID    string `json:"correlation_id"`
	Crn              struct {
		Version     string `json:"version"`
		ServiceName string `json:"service_name"`
	} `json:"crn"`
	CustomerImpactDescription string `json:"customer_impact_description"`
	CustomersImpacted         string `json:"customers_impacted"`
	DisablePager              string `json:"disable_pager"`
	IncidentID                string `json:"incident_id"`
	IncidentState             string `json:"incident_state"`
	IncidentUIURL             string `json:"incident_ui_url"`
	LongDescription           string `json:"long_description"`
	OutageEnd                 string `json:"outage_end"`
	OutageStart               string `json:"outage_start"`
	Severity                  int    `json:"severity"`
	ShortDescription          string `json:"short_description"`
	Source                    string `json:"source"`
	Timestamp                 string `json:"timestamp"`
	Version                   string `json:"version"`
}
type IncidentsFromIncidentMgmtApi struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Count  int `json:"count"`
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
	Resources []IncidentMgmtApiIncident `json:"resources"`
}

func GetConcernURL(server *rest.Server) string {
	const fct = "[GetConcernURL]"

	if server == nil {
		server = &rest.Server{}
		server.SNToken = os.Getenv("SERVER_KEY")
	}

	url := catalogURL + incidentMgmtAPIPath
	catalogResponse := new(CatalogResponse)
	if err := server.GetAndDecode(fct, "GetConcernURL", url, catalogResponse); err != nil {
		log.Println(fct, err.Error())
		return ""
	}

	log.Println(fct, hlp.GetPrettyJson(catalogResponse))
	return catalogResponse.Concerns.Href
}

func GetIncidentFromSN(server *rest.Server, incidentId string) (sysid string, err error) {
	const fct = "[GetIncidentFromSN]"
	log.Println(fct, "starting and sleeping")
	time.Sleep(10 * time.Second)
	log.Println(fct, "done sleeping")

	if server == nil {
		server = &rest.Server{}
		server.SNToken = os.Getenv("SN_KEY")
	}

	if incidentId == "" {
		return "", errors.New("no incident ID provided")
	}

	incidents := new(SNRecords)

	req, err := http.NewRequest(http.MethodGet, SnAPIGetURL, nil)
	if err != nil {
		return "", err
	}

	req.Header["Authorization"] = []string{"Bearer " + server.SNToken}
	req.Header["query"] = []string{"number=" + incidentId}
	req.Header["end"] = []string{"1"}
	req.Header["start"] = []string{"1"}
	req.Header["updated_by"] = []string{"dchang8@us.ibm.com"}

	if err := server.GetRequestAndDecode(fct, "GetIncidentFromSN", req, incidents); err != nil {
		log.Println(fct, err.Error())
		return "", err
	}

	log.Println(fct, hlp.GetPrettyJson(incidents))

	if len(incidents.Results) == 0 {
		errMsg := "Error: no incidents returned"
		log.Println(fct, errMsg)
		return "", errors.New(errMsg)
	}

	sysid = incidents.Results[0].SysID
	log.Println(fct, "*** Incident returned, SysID:", sysid)
	return sysid, err
}

func UpdateIncidentToCIE(incidentNumber string) (snrecord *SNRecord, err error) {
	snPayload := `{"u_environment":"Washington DC REGION","u_disruption_time":"1970-01-01 00:00:45","state":"2","u_status":"21","u_affected_activity":"Application Availability"}`
	return UpdateIncident(incidentNumber, snPayload)
}

func UpdateIncidentToClosed(incidentNumber string) (snrecord *SNRecord, err error) {
	const fct = "UpdateIncidentToClosed"
	snPayload := `{"state":"7", }`

	if incidentNumber == "" {
		log.Println(fct, errors.New("No incidentNumber provided"))
		return nil, err
	}

	snrecord, err = UpdateIncident(incidentNumber, snPayload)

	return snrecord, err
}

func UpdateIncidentToResolved(incidentNumber string) (snrecord *SNRecord, err error) {
	const fct = "UpdateIncidentToResolved"
	snPayload := `{"u_status":"22"}`

	if incidentNumber == "" {
		log.Println(fct, errors.New("No incidentNumber provided"))
		return nil, err
	}

	snrecord, err = UpdateIncident(incidentNumber, snPayload)
	if err != nil {
		return snrecord, err
	}
	snPayload = `{"state":"6", "close_code": "Closed/Resolved by Caller","close_notes":"Closing out test" }`
	snrecord, err = UpdateIncident(incidentNumber, snPayload)
	return snrecord, err
}

func UpdateIncident(incidentNumber string, snPayload string) (snRecord *SNRecord, err error) {
	const fct = "[UpdateIncident]"

	if incidentNumber == "" {
		return nil, errors.New("no incidentNumber provided")
	}

	log.Println(fct, "Request body:", snPayload)
	SnAPIUpdateURL := SnAPIURL + "/" + incidentNumber
	req, err := createSNRequest(http.MethodPatch, SnAPIUpdateURL, bytes.NewBuffer([]byte(snPayload)))
	if err != nil {
		log.Println(fct, err.Error())
		return nil, err
	}

	server := &rest.Server{}
	server.SNToken = os.Getenv("SN_KEY")

	result := new(SNRecordResult)
	err = server.PostRequestAndDecode(fct, "SNRecordResult", req, result)
	if err != nil {
		log.Println(fct, err.Error())
		return nil, err
	}

	snRecord = &result.Result
	log.Println(fct, "*** Incident updated, SysID:", snRecord.SysID)
	return snRecord, err
}

// Creates an incident via the Incident Management API
// Due to limitations, you need to use this to create the incident and
// update the incident (via the table api) with the UpdateIncidentToCIE() function.
// Params can be nil and "" as they will go to defaults.
// URL is based off the environment variable CATALOG_URL
func CreateIncident() (incidentID string, sysID string, err error) {
	const fct = "[CreateIncident]"

	incidentToCreate := SNRecord{
		ShortDescription: "Test incident created by test program",
		CmdbCi:           "cloud-object-storage",
		Description:      "This is a sample alert created from the API explorer application. It is not a real incident. This should only appear in the test environment.",
		IncidentState:    "2",
		UDetectionSource: "Monitoring Tool",
		Severity:         "1",
	}

	bodyBytes, err := json.Marshal(incidentToCreate)
	if err != nil {
		log.Println(fct, err.Error())
		return incidentID, sysID, err
	}
	log.Println(fct, "Request body:", string(bodyBytes))

	req, err := createSNRequest(http.MethodPost, SnAPIURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Println(fct, err.Error())
		return incidentID, sysID, err
	}

	server := &rest.Server{}
	server.SNToken = os.Getenv("SN_KEY")

	result := new(SNRecordResult)
	err = server.PostRequestAndDecode(fct, "SNRecordResult", req, result)
	if err != nil {
		log.Println(fct, err.Error())
		return incidentID, sysID, err
	}

	incident := result.Result
	log.Println(fct, "*** Incident created, SysID:", incident.SysID)
	return incident.Number, incident.SysID, err
}

// GetFromIncidentsMgmtApi We get way too many incidents this way
/*
func GetFromIncidentsMgmtApi(serv *rest.Server, url string, query string) (incidents []IncidentMgmtApiIncident) {

	const fct = "GetFromIncidentMgmtApi"
	if serv == nil {
		serv = &rest.Server{}
		serv.SNToken = os.Getenv("SERVER_KEY")
	}

	if query == "" {
		query = "incident_state=confirmed-cie&severity=1"
	}
	if url == "" {
		url = catalogURL + incidentMgmtAPIConcernPath + "?" + query
	}

	list := new(IncidentsFromIncidentMgmtApi)
	serv1 := rest.Server{}
	serv1.Token = os.Getenv("SERVER_KEY")

	err := serv1.GetAndDecode("GET", "resource.ResourceList", url, list)
	if err != nil {
		log.Println("GET", nil, err.Error())
		return nil
	}

	if len(list.Resources) == 0 {
		log.Println("GET", nil, "No resources returned in query.")
		return nil
	}
	continueLoop := true
	offset := 0
	limit := 25
	for continueLoop {
		checkList := new(IncidentsFromIncidentMgmtApi)
		offset += limit
		fullUrl := url + "&offset=" + strconv.Itoa(offset) + "&limit=25"
		err = serv1.GetAndDecode("GET", "resource.ResourceList", fullUrl, checkList)
		if err != nil {
			log.Println("GET", err, "Failed to get Last on resource list.")
		}
		if len(checkList.Resources) > 0 {
			list.Resources = append(list.Resources, checkList.Resources...)
		} else {
			continueLoop = false
		}

	}
	return list.Resources
}
*/
func createSNRequest(method, url string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header["Authorization"] = []string{"Bearer " + os.Getenv("SN_KEY")}
	req.Header["updated_by"] = []string{"michael_lee@us.ibm.com"}
	req.Header["Content-Type"] = []string{"application/json"}

	return
}

/*
func CreateIncident(){

	const fct = "CreateIncident"
	if serv == nil {
		serv = &rest.Server{}
		serv.Token = os.Getenv("SN_KEY")
	}

	if subscription != nil {
		url = subscription.Watches.URL
	} else if url == "" {
		log.Println(fct, "No url or subscription provided")
		return nil, errors.New("No url or subscription provided")
	}
	if err := serv.GetAndDecode(fct, "WatchReturnArray", subscription.Watches.URL, &watches); err != nil {
		log.Println(fct, err.Error())
		return nil, err

	}

	log.Println(fct, "watches: \n"+hlp.GetPrettyJson(watches))
	return watches, err


}
*/

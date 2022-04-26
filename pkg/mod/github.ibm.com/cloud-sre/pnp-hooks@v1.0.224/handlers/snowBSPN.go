package handlers

// This handler will handle BSPNs when they are pushed from ServerNow.  The general format of the BSPN is represented by this sample JSON:
// {"type":"notification","regions":["ibm:yp:eu-gb"],"start_date":1536748020000,"end_date":1536764700000,"title":"RESOLVED: Issues with App Connect service","description":"<p>SERVICES/COMPONENTS AFFECTED:<br />&nbsp; - App Connect</p>\n<p><br />IMPACT:<br />&nbsp; - Users may experience problems running event-driven flows. Users will notice that their flows are not running. The flows will show an error state in the UI.<br />&nbsp; - Pure API-driven flows are not impacted.</p>\n<p><br />STATUS:<br />&nbsp; - 2018-09-12 11:00 UTC - INVESTIGATING - The operations team is aware of the issues and is currently investigating.<br />&nbsp; - 2018-09-12 15:25 UTC - RESOLVED - The issues have been resolved as of 15:05 UTC.</p>\n<p>&nbsp;</p>","severity":"Sev - 1","components":["cloudoe.sop.enum.paratureCategory.literal.l377"],"id":"BSP0002135","parent_id":"INC0279855","status":"CIE In Progress","affected_activities":"Application Availability","modified":1536765943588,"crn":["crn:v1:d-aa2:dedicated:velocity_test:us-south::::","crn:v1:d-aa1:dedicated:velocity_test:us-east::::","crn:v1:d-alfaevolution:dedicated:velocity_test:mil01::::"]}

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	ossadapter "github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
	adapter "github.ibm.com/cloud-sre/pnp-notifications-adapter/api"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/initadapter"
)

// BSPN is the struct used to parse the incoming request from ServiceNow for BSPN information
type BSPN struct {
	Type               string   `json:"type"`
	Regions            []string `json:"regions"`
	CreateDate         int64    `json:"created_on"` // In milliseconds
	StartDate          int64    `json:"start_date"` // In milliseconds
	EndDate            int64    `json:"end_date"`   // In milliseconds
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Severity           string   `json:"severity"`
	Components         []string `json:"components"` // Notification category IDs
	ID                 string   `json:"id"`
	ParentID           string   `json:"parent_id"` // Incident ID
	Status             string   `json:"status"`    //e.g. "CIE In Progress"
	AffectedActivities string   `json:"affected_activities"`
	Modified           int64    `json:"modified"` // In milliseconds
	CRN                []string `json:"crn"`
}

var (
	creds *adapter.SourceConfig
)

// Update is a constant for the message type representing an update
const Update = datastore.NotificationMsgType("update")

// ServeSnowBSPN - services /bspn requests
func ServeSnowBSPN(res http.ResponseWriter, req *http.Request) {

	// intercept panics: print error and stacktrace
	defer api.HandlePanics(res)
	defer req.Body.Close()

	log.Print("Invoking SNow BSPN")

	ctx, _ := buildContext()
	msgs, err := transformMessage(ctx, req)
	if err != nil {
		log.Println(err)
		return
	}

	for _, msg := range msgs {

		// post message to rabbitMQ
		httpStatus, err := shared.HandleMessage(bytes.NewReader(msg), "notification")
		if err != nil && httpStatus != http.StatusOK {
			log.Println(err)
			res.WriteHeader(httpStatus)
			return
		}
	}

	//log.Println("successfully produced BSPN to RabbitMQ")
}

var ctxCount = 0
var nrmon *exmon.Monitor

func buildContext() (ctx ctxt.Context, err error) {

	FCT := "buildContext"

	logID := fmt.Sprintf("hooks%d", ctxCount)
	ctxCount++

	if nrmon == nil {
		initadapter.Initialize() // Note for UT you can replace this function via function pointer. See initialize.go
		nrmon, err = exmon.CreateMonitor()
		if err != nil {
			log.Printf("ERROR (%s, %s): Cannot create NR monitor. New Relic monitoring disabled!! %s", FCT, logID, err.Error())
		}
	}

	return ctxt.Context{LogID: logID, NRMon: nrmon}, nil
}

// transformMessage will change the incoming BSPN message to the format that the nq2ds consumer expects
// it is possible, and likely, that this method will return multiple messages to be sent
func transformMessage(ctx ctxt.Context, req *http.Request) (result [][]byte, err error) {

	METHOD := "transformMessage"

	if creds == nil {
		creds, err = adapter.GetCredentials()
		if err != nil {
			log.Printf("ERROR (%s): Failed to read credentials: %s", METHOD, err.Error())
			return nil, err
		}
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("ERROR (%s): Failed to read data from Body: %s", METHOD, err.Error())
		return nil, err
	}

	rawMessage := string(data)
	log.Print("Raw BSPN message before transform: ", rawMessage)

	b, err := parseBSPN(data)
	if err != nil {
		return nil, err
	}

	notifications, err := BSPNToNoteInsert(ctx, b)
	if err != nil {
		return nil, err
	}

	result = make([][]byte, 0, 1)
	var foundPrimary = false // Marks whether the notification used for subscription has been identified

	for _, n := range notifications {

		// Needed since this goes to notification queue -> subscription consumer/nq2ds.
		// It could get kicked out by nq2ds and then we have a notification without a db record
		// This in cases where a BSPN has at least one CRN that would not go to the public status page
		if !api.IsCrnPnpValid(n.CRNFull) {
			log.Printf("INFO (%s): CRN is not a public CRN, stopping processing of notification. CRN=%s", METHOD, n.CRNFull)
			continue
		}

		//wrap in mq.NotificationMsg struct for https://github.ibm.com/cloud-sre/pnp-hooks/issues/40
		noteMsg := datastore.NotificationMsg{MsgType: Update, NotificationInsert: *n}
		if foundPrimary == false &&
			noteMsg.PnPRemoved == false {
			noteMsg.IsPrimary = true
			foundPrimary = true
		}

		buffer := new(bytes.Buffer)
		if err = json.NewEncoder(buffer).Encode(noteMsg); err != nil {
			return nil, fmt.Errorf("ERROR (%s): cannot encode notification [%s]", METHOD, err.Error())
		}

		result = append(result, buffer.Bytes())

	}
	reverseResult := make([][]byte, 0, 1)
	for i := len(result) - 1; i >= 0; i-- {
		log.Print("BSPN messages after transform: ", string(result[i]))
		reverseResult = append(reverseResult, result[i])
	}

	return reverseResult, nil
}

// BSPNToNoteInsert is used to transform a bspn to a NotificationInsert
func BSPNToNoteInsert(ctx ctxt.Context, b *BSPN) (list []*datastore.NotificationInsert, err error) {

	METHOD := "BSPNToNoteInsert"

	if len(b.Components) > 1 { // Note there should only be one of these
		log.Printf("INFO (%s): More than one component from ServiceNow. Expected only 1, got %d.", METHOD, len(b.Components))
	}

	n := new(datastore.NotificationInsert)

	n.SourceUpdateTime = msToNoteTime(b.Modified)
	n.SourceCreationTime = msToNoteTime(b.CreateDate)
	n.EventTimeStart = msToNoteTime(b.StartDate)
	n.EventTimeEnd = msToNoteTime(b.EndDate)
	n.Source = "servicenow"
	n.SourceID = b.ID
	n.Type = "incident"
	n.IncidentID = b.ParentID
	n.ShortDescription = []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: b.Title}}
	n.LongDescription = []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: b.Description}}

	// The next items need some lookups and we need the service name.  Use OSS catalog to find the service name.
	serviceName, err := ossadapter.CategoryIDToServiceName(ctx, b.Components[0])
	if err != nil {
		return nil, err
	}

	if serviceName == "" {
		log.Printf("WARNING (%s): Service name not found in OSS catalog.", METHOD)
		return nil, errors.New("Servicename not found")
	}

	n.ResourceDisplayNames, err = getDisplayNames(ctx, b.Components[0], serviceName)
	if err != nil {
		return nil, err
	}

	n.Category, err = getEntryType(ctx, b.Components[0])
	if err != nil {
		return nil, err
	}

	// Now need to create a separate notification record for each CRN
	list = make([]*datastore.NotificationInsert, 0, len(b.CRN))

	for _, crn := range b.CRN {

		o := new(datastore.NotificationInsert)

		o.SourceUpdateTime = n.SourceUpdateTime
		o.EventTimeStart = n.EventTimeStart
		o.EventTimeEnd = n.EventTimeEnd
		o.Source = n.Source
		o.SourceID = n.SourceID
		o.Type = n.Type
		o.IncidentID = n.IncidentID
		o.ShortDescription = n.ShortDescription
		o.LongDescription = n.LongDescription

		o.SourceCreationTime = n.SourceCreationTime
		if len(n.SourceCreationTime) == 0 {
			o.SourceCreationTime = n.SourceUpdateTime // Use source update time as an emergency backup
		}

		o.Category = n.Category
		o.CRNFull = buildCRN(crn, serviceName)
		o.ResourceDisplayNames = n.ResourceDisplayNames

		list = append(list, o)
	}

	return list, nil
}

func getEntryType(ctx ctxt.Context, categoryID string) (string, error) {
	et, err := ossadapter.CategoryIDToEntryType(ctx, categoryID)
	if err != nil {
		return "", err
	}

	switch et {
	case string(ossrecord.PLATFORMCOMPONENT):
		et = "platform"
	case string(ossrecord.SERVICE):
		et = "services"
	case string(ossrecord.RUNTIME):
		et = "runtimes"
	default:
		et = strings.ToLower(et)
	}

	return et, nil
}

// GetEntryType returns the entry type based on the categoryID
// exported function to be used by other components such as nq2ds
func GetEntryType(ctx ctxt.Context, categoryID string) (string, error) {
	return getEntryType(ctx, categoryID)
}

var crnRegex *regexp.Regexp

// buildCRN will recreate the crn using the provided service name.  This is needed
// in case the notification category ID pointed to a parent resource different from the crn
func buildCRN(crn, serviceName string) string {

	// example: "crn:v1:bluemix:public:cloudantnosqldb:us-south::::"
	if crnRegex == nil {
		crnRegex = regexp.MustCompile("(crn:v1:[^:]*:[^:]*:)[^:]*(:[^:]*:[^:]*:[^:]*:[^:]*:.*)")
	}

	matches := crnRegex.FindStringSubmatch(crn)

	if len(matches) == 3 {
		crn = matches[1] + serviceName + matches[2]
	}

	return crn
}

// BuildCRN recreates the crn using the provides service name
// exported function to be used by other components such as nq2ds
func BuildCRN(crn, serviceName string) string {
	return buildCRN(crn, serviceName)
}

func getDisplayNames(ctx ctxt.Context, categoryID, serviceName string) (result []datastore.DisplayName, err error) {

	METHOD := "getDisplayNames"

	if creds == nil {
		creds, err = adapter.GetCredentials()
		if err != nil {
			log.Printf("ERROR (%s): Failed to read credentials: %s", METHOD, err.Error())
			return nil, err
		}
	}

	cd := &adapter.CRNData{}
	cd.AddDisplayNames(ctx, serviceName, categoryID, creds, nil)

	list := make([]datastore.DisplayName, 0, len(cd.DisplayNames))

	for _, dn := range cd.DisplayNames {
		list = append(list, datastore.DisplayName{Language: dn.Language, Name: dn.Text})
	}

	return list, nil
}

// GetDisplayNames returns the list of display names for a specific categoryID and service name
// exported function to be used by other components such as nq2ds
func GetDisplayNames(ctx ctxt.Context, categoryID, serviceName string) (result []datastore.DisplayName, err error) {
	return getDisplayNames(ctx, categoryID, serviceName)
}

// msToNoteTime will create a formatted timestamp using the input milliseconds
func msToNoteTime(ms int64) string {

	// ServiceNow sends 0 for not set values, in this case we don't want to default to epoch start:
	if ms == 0 {
		return ""
	}

	ns := ms * 1000000

	t := time.Unix(0, ns).UTC()

	return t.Format("2006-01-02T15:04:05.000Z")
}

// parseBSPN parses byte data into bspn struct
func parseBSPN(data []byte) (*BSPN, error) {

	METHOD := "parseBSPN"

	// Decode the json result into a struct
	b := new(BSPN)
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(b); err != nil {
		msg := fmt.Sprintf("ERROR (%s): Failed to decode the BSPN input from ServiceNow: %s", METHOD, err.Error())
		log.Println(msg, string(data))
		return nil, errors.New(msg)

	}

	return b, nil
}

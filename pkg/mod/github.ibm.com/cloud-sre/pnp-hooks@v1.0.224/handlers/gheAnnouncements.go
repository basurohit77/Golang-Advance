package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/go-github/github"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	gitnotice "github.ibm.com/cloud-sre/pnp-gitnotice"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
	"github.ibm.com/sosat/githubzen/v2"
)

const (
	// SecurityLabel is the label on an issue indicating it is a security notice. This is added by humans.
	SecurityLabel = "security"
	// AnnouncementLabel is the label on an issue indicating it is an announcement. This is added by humans.
	AnnouncementLabel = "announcement"
	// ErrorLabel is the label on an issue indicating that an error occured during processing. This is added by this code.
	ErrorLabel = "ERROR"
	// ExpiredLabel is the label on an issue indicating that the notification is expored. This is added by this code.
	ExpiredLabel = "EXPIRED"
	// PublishLabel is the label on an issue indicating it should be available from PnP. This is added by humans.
	PublishLabel = "publish"
	// PrivateLabel is a label that indicates if an announcement is supposed to have a private audience
	PrivateLabel = "private"
	// RemovedLabel is the label on an issue indicating that a previously published issue is to be removed. The publish
	// label should also still be on the issue. This essentially adds the pnpRemoved tag to a notification.
	RemovedLabel = "removed"
	// OpenedAction is a github event action occuring when someone creates a new issue
	OpenedAction = "opened"
	// LabeledAction is a github event action occuring when someone adds a label to an issue
	LabeledAction = "labeled"
	// UnlabeledAction is a github event action occuring when someone removes a label from an issue
	UnlabeledAction = "unlabeled"
	// EditedAction is a github event action occuring when someone edits the issue description
	EditedAction = "edited"
	// ClosedState is a state of the git issue when it is closed
	ClosedState = "closed"
)

var (
	// EnableNotificationAudience indicates if the 'audience' attribute should be enabled for notifications
	EnableNotificationAudience = os.Getenv("NOTIFAUDIENCE") == "true"

	shaKey         = []byte(os.Getenv("PNP_GITHUB_SHAKEY"))       // SHA key to validate incoming messages
	shaKeyOlder    = []byte(os.Getenv("PNP_GITHUB_SHAKEY_OLDER")) // Older SHA key to validate incoming messages
	gitAccessToken = os.Getenv("PNP_GITHUB_TOKEN")                // Token to access github
	pnpServiceID   = os.Getenv("PNP_GITHUB_ID")                   // Login ID that the github token represents

	errCIMissing = errors.New("no CI present in notice")
)

// Simple counter to correlate transactions in the log
var logCounter int32

// ServeGheAnnouncements will accept push notices from GitHub enterprise to handle security notices and announcements.
func ServeGheAnnouncements(res http.ResponseWriter, req *http.Request) {

	FCT := "ServeGheAnnouncements"

	// intercept panics: print error and stacktrace
	defer api.HandlePanics(res)
	defer req.Body.Close()

	if len(shaKey) == 0 || len(gitAccessToken) == 0 {
		log.Println("ERROR: You must configure PNP_GITHUB_SHAKEY and PNP_GITHUB_TOKEN")
		return
	}

	// Create log transaction ID to make it easier to track this transaction in logs
	logTxn := fmt.Sprintf("[lg%d]", atomic.AddInt32(&logCounter, 1))
	FCT += " " + logTxn // Attach transaction ID to the Method name so it's included in logs

	payload, err := validatePayload(logTxn, res, req)
	if err != nil {
		return
	}

	rawMessage := string(payload)
	log.Println(FCT, ": Raw GHE message in hooks: ", rawMessage)

	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		log.Println("ERROR", FCT, ": Cannot parse GHE payload. error=", err)
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch event := event.(type) {
	case *github.IssuesEvent:
		enqueueIssueEvent(logTxn, event) // Place item on internal work queue (not rabbit)
	default:
		log.Printf("ERROR %s: Ignoring unrecognized event type. %T", FCT, event)
		return
	}

}

// validatePayload is used to validate the payload that is received in the request from Github.
// GHE works with different tokens than other APIs.  GHE does not allow customization of
// the Authorization token.  They have their own custom token, so that's what we use here.  The signature
// needs the body text to be calculated, so initially we just check if there is a signature, but we validate
// the signature after reading the body.
// [2022-03-04] To enable shakey rotation, we need to enable 2 keys to be valid at the same time. So attempt to
// validate one key, and if no good, try to validate the other. Rotation of this key should be done in this order:
// 1. Decide what the new shakey will be. Remember you can choose what it will be.  However, do not set it in GHE at this time.
// 2. Copy the existing key from ghe/shakey to ghe/shakey_older in vault
// 3. Set the vault value for ghe/shakey to the value decided upon in step #1
// 4. Redeploy all pods.
// 5. In GHE, set the shakey for the pnphooks to the value decided on in step #1
// 6. The vault value in ghe/shakey_older can be set to an empty string
// 7. Redeploy all pods.
func validatePayload(logTxn string, res http.ResponseWriter, req *http.Request) (payload []byte, err error) {

	FCT := "validatePayload " + logTxn // Attach transaction ID to the Method name so it's included in logs

	// Because we may have to validate the payload twice, we need to save the request body for possible reuse.
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			log.Println("ERROR", FCT, logTxn, ": Cannot read GHE payload prior to validation. error=", err)
			return nil, err
		}
	}

	if req.Body != nil { // keep nil request body as is, otherwise reset request body
		req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}
	payload, err = github.ValidatePayload(req, shaKey)
	if err != nil {
		if len(shaKeyOlder) > 0 {
			log.Println("INFO", FCT, logTxn, ": Cannot validate GHE payload with first shakey. We could be in rotation, so check backup. error=", err, string(shaKey))
			if req.Body != nil { // keep nil request body as is, otherwise reset request body
				req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			}
			payload, err = github.ValidatePayload(req, shaKeyOlder)
		}
		if err != nil {
			log.Println("ERROR", FCT, logTxn, ": Cannot validate GHE payload. Likely invalid or missing GHE shakey. If rotating, be sure to fill in vault values for github new and old shakeys error=", err, string(shaKeyOlder))
			res.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		// Swap the keys to give preference to the other key since it worked
		badKey := shaKey
		shaKey = shaKeyOlder
		shaKeyOlder = badKey
	}

	return payload, err
}

type gheWorkItem struct {
	Event  *github.IssuesEvent
	LogTxn string
}

var gheQueue chan *gheWorkItem

var gheQueueMux = new(sync.Mutex)

// enqueueIssueEvent will post an issue event to a channel for processing. We do it
// this way so that we can return the request thread of execution quicker so that
// github doesn't have to wait for things like the stocking of our global catalog cache
// which could take up to 45 seconds.  This could cause a timeout error on the github
// side which looks like an error, but it isn't
func enqueueIssueEvent(logTxn string, event *github.IssuesEvent) {

	gheQueueMux.Lock()
	defer gheQueueMux.Unlock()

	if gheQueue == nil {
		gheQueue = make(chan *gheWorkItem)
		go pullIssueEvent()
	}

	item := &gheWorkItem{Event: event, LogTxn: logTxn}

	gheQueue <- item
}

func pullIssueEvent() {

	for {
		wi := <-gheQueue
		if wi != nil {
			processGheIssueEvent(wi.LogTxn, wi.Event)
		}
	}
}

func processGheIssueEvent(logTxn string, event *github.IssuesEvent) {
	FCT := "processGheIssueEvent" + logTxn

	defer handlePanics(logTxn)

	if event == nil {
		log.Printf("%s: ERROR nil event passed", FCT)
		return
	}
	if event.Issue == nil {
		log.Printf("%s: ERROR nil issue in event", FCT)
		return
	}
	if event.Issue.GetNumber() <= 2 { // Issues 1 and 2 are reserved issue numbers used for testing and perhaps other special purposes in any repo
		log.Printf("%s: Taking no action. Issue is special issue number %d", FCT, event.Issue.GetNumber())
		return
	}
	if event.GetSender() != nil && event.GetSender().GetLogin() == pnpServiceID { // don't process events caused by this code
		log.Printf("%s: Taking no action. Issue event was caused by the PNP service ID %s", FCT, event.GetSender().GetLogin())
		return
	}
	if event.Issue.GetBody() == "" {
		log.Printf("%s: ERROR no body in issue", FCT)
		return
	}

	// The issue can have any number of actions, but we pay attention to as few as possible to reduce complexity. Types of actions we are interested in:
	// - opened : Indicates an issue was created. It may or may not be published at this time. Ignore if the publish label is missing.
	// - edited : Indicates a change was made to the description.  Ignore this if the publish label is missing.
	// - labeled : Indicates a label was added. Very important for things like publish or removed.
	// - unlabeled : Indicates a label was removed.
	//
	// Actions we ignore
	// - deleted : This should not occur. Indicates the issue was deleted. I don't know if this is even possible, so we ignore this.
	// - transferred
	// - pinned
	// - unpinned
	// - closed : Issue closure really has no meaning for us.  We work from labels. Humans can use this for their own needs.
	// - reopened
	// - assigned
	// - unassigned
	// - locked
	// - unlocked
	// - milestoned
	// - demilestoned

	var notes []*datastore.NotificationInsert

	var nError *gitnotice.Error

	// Currently we handle the different cases the same. Case statement left here incase we determine
	// some variation needed in the future.
	switch event.GetAction() {
	case OpenedAction:
		notes, nError = processEdited(logTxn, event)
	case EditedAction:
		notes, nError = processEdited(logTxn, event)
	case LabeledAction:
		notes, nError = processLabeled(logTxn, event)
	case UnlabeledAction:
		notes, nError = processEdited(logTxn, event)
	default:
		log.Printf("%s: Ignoring action %s", FCT, event.GetAction())
		return
	}

	if nError != nil {
		log.Printf("%s: ERROR could not process the event: %s", FCT, nError.String())
		sendProblemToGit(logTxn, nError, event)
		return
	}

	// No notifications to process, this is a normal condition for example if
	// a new notification comes in that is not published.
	if len(notes) == 0 {
		log.Printf("%s: No notices to be processed.", FCT)
		return
	}

	// Build and send MQ messages based on the constructed notifications
	msgs, err := createMqMsgs(logTxn, notes)
	if err != nil {
		// This error should not occur. It would mean there is a Decoding problem which should be caught in development.
		log.Printf("%s: ERROR Failed to generate the MQ messages: %s", FCT, err.Error())
		sendProblemToGit(logTxn, gitnotice.NewError(err, "Failed to generate MQ Messages"), event)
		return
	}

	for _, msg := range msgs {

		// post message to rabbitMQ
		httpStatus, err := shared.HandleMessage(bytes.NewReader(msg), "notification")
		if err != nil && httpStatus != http.StatusOK {
			log.Printf("%s: ERROR when sending message to mq:%s", FCT, err.Error())
			log.Println(FCT + "msg=" + string(msg))
			sendProblemToGit(logTxn, gitnotice.NewError(err, "Error when sending to message queue: %d", httpStatus), event)
			return
		}
	}

	// When we are here, it appears everything went well, so post back to github to let user know
	// we successfully got the update
	sendSuccessToGit(logTxn, event)
}

// processEdited is used to process an issue that has been updated.
func processEdited(logTxn string, event *github.IssuesEvent) (notifications []*datastore.NotificationInsert, nError *gitnotice.Error) {

	FCT := "processEdited " + logTxn

	if event.Issue == nil {
		log.Printf("%s: issue is empty", FCT)
		return nil, nil
	}

	var number int
	if event.Issue.Number != nil {
		number = *event.Issue.Number
	}

	if !hasLabel(PublishLabel, event.Issue) {
		// Any updated issues need to have the publish flag enabled for us to process it.
		log.Printf("%s: issue %d does not have publish label; ignoring the issue", FCT, number)
		return nil, nil
	}

	if isState(ClosedState, event.Issue) {
		// Ignore updates from closed issues
		log.Printf("%s: issue %d is closed; ignoring the issue", FCT, number)
		return nil, nil
	}

	return notificationInsertFromGhe(logTxn, event)
}

// processLabeled is used to process an issue that has been updated.
func processLabeled(logTxn string, event *github.IssuesEvent) (notifications []*datastore.NotificationInsert, nError *gitnotice.Error) {

	notifications, nError = processEdited(logTxn, event)

	if nError != nil && event.GetLabel().GetName() == PublishLabel {
		// The author just added the publish label, and an error occured, so lets remove the publish label
		removeLabel(logTxn, event, PublishLabel)
	}

	return notifications, nError
}

func isState(state string, issue *github.Issue) (ok bool) {
	if issue == nil {
		return false
	}

	if issue.State != nil && strings.ToLower(*issue.State) == state {
		return true
	}

	return false
}

func hasLabel(label string, issue *github.Issue) (ok bool) {
	if issue == nil {
		return false
	}

	for _, l := range issue.Labels {
		if strings.ToLower(l.GetName()) == strings.ToLower(label) {
			return true
		}
	}

	return false
}

func createMqMsgs(logTxn string, notifications []*datastore.NotificationInsert) (result [][]byte, err error) {
	FCT := "createMqMsgs " + logTxn

	result = make([][]byte, 0, 1)

	var foundPrimary = false // Marks whether the notification used for subscription has been identified
	for _, n := range notifications {

		// Needed since this goes to notification queue -> subscription consumer/nq2ds.
		// It could get kicked out by nq2ds and then we have a notification without a db record
		// This in cases where a BSPN has at least one CRN that would not go to the public status page
		if !api.IsCrnPnpValid(n.CRNFull) {
			log.Printf("INFO (%s): CRN is not a public CRN, stopping processing of notification. CRN=%s", FCT, n.CRNFull)
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
			return nil, fmt.Errorf("ERROR (%s %s): cannot encode notification [%s]", FCT, logTxn, err.Error())
		}

		result = append(result, buffer.Bytes())

	}

	reverseResult := make([][]byte, 0, 1)
	for i := len(result) - 1; i >= 0; i-- {
		log.Print("GHE messages after transform: ", string(result[i]))
		reverseResult = append(reverseResult, result[i])
	}

	return reverseResult, nil
}

// notificationInsertFromGhe generate a NotificationInsert based on GHE event
func notificationInsertFromGhe(logTxn string, event *github.IssuesEvent) (list []*datastore.NotificationInsert, nError *gitnotice.Error) {

	FCT := "notificationInsertFromGhe " + logTxn

	ctx, err := getContextForGhe(logTxn)
	if err != nil {
		log.Printf("%s: ERROR could not get context. Will continue without NewRelic support", FCT)
		ctx = ctxt.Context{LogID: logTxn}
	}

	// Parse the git issue Description field to pull out the individual notice pieces
	// Note that the error returned from BuildNotice is an enhanced type of error that contains text we can
	// post back to GHE to inform the creator of their error.
	notice, nError := gitnotice.BuildNotice(event.Issue.GetBody())
	if nError != nil {
		return nil, nError
	}

	// Now check that the CI and environments are OK.  We need to do this here instead of nq2ds so that when
	// we put the notification on the message queue, a subscriber doesn't get a notification with a bad CI or environment.
	nError = checkCI(logTxn, notice)
	if nError != nil {
		return nil, nError
	}

	n := new(datastore.NotificationInsert)

	// Slug out some of the easy fields
	tFormat := "2006-01-02T15:04:05.000Z"
	n.SourceUpdateTime = event.Issue.GetUpdatedAt().Format(tFormat)
	n.SourceCreationTime = event.Issue.GetCreatedAt().Format(tFormat)
	n.Source = "ghe"
	n.SourceID = gheSourceID(logTxn, event)
	n.Type = getNotificationType(event)
	//n.IncidentID = No incident associated with Announcement or security notice
	n.ShortDescription = []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: notice.Title}}
	n.LongDescription = []datastore.DisplayName{datastore.DisplayName{Language: "en", Name: notice.Description}}
	n.PnPRemoved = hasLabel(RemovedLabel, event.Issue) || (event.GetAction() == LabeledAction && event.Label.GetName() == RemovedLabel)
	if EnableNotificationAudience {
		n.Tags = getTags(notice)
	} else {
		log.Println("notification audience feature not enabled")
	}

	if notice.Start.IsZero() {
		// [2019-08-05]: Bill does not want to specify a separate start time.
		// So we fall in here when no start time is specified in the git issue which will be the usual case
		n.EventTimeStart = n.SourceCreationTime
	} else { // Just in case a start is ever present, we will use it
		n.EventTimeStart = notice.Start.Format(tFormat)
	}
	//n.EventTimeEnd = No end time to specify

	// For now we only support one service being passed from the GHE record.  The structure has an array just in
	// case the future holds more services.  Here we just check that we have one, then we just use one.
	if len(notice.Services) == 0 {
		log.Printf("%s: Did not get any service names", FCT)
		return nil, gitnotice.NewError(nil, "A service name is required.")
	}
	serviceName := notice.Services[0]

	// Here we need to do a quick check if this is a "roll-up" scenario.  In a roll-up, the user specified a service name that
	// we actually do not show to customers. Instead we use the service's parent to show to customers.
	serviceName, err = normalizeServiceName(serviceName, ctx)
	if err != nil {
		log.Printf("%s: Could not lookup parent for servicename %s. error=%s", FCT, serviceName, err.Error())
		return nil, gitnotice.NewError(err, "Could not lookup parent for servicename %s.", serviceName)
	}

	categoryID, err := osscatalog.ServiceNameToCategoryID(ctx, serviceName)
	if err != nil {
		log.Printf("%s: Could not convert servicename %s to a category ID. error=%s", FCT, serviceName, err.Error())
		return nil, gitnotice.NewError(err, "Could not find a category ID for service name %s.", serviceName)
	}

	n.ResourceDisplayNames, err = GetDisplayNames(ctx, categoryID, serviceName)
	if err != nil {
		log.Printf("%s: Database error retrieving display names for servicename %s. error=%s", FCT, serviceName, err.Error())
		return nil, gitnotice.NewError(err, "Could not find display names for service name %s.", serviceName)
	}

	n.Category, err = GetEntryType(ctx, categoryID)
	if err != nil {
		log.Printf("%s: Could not determine category for servicename %s, categoryID %s. error=%s", FCT, serviceName, categoryID, err.Error())
		n.Category = "services" // in case of error, set this to services as we don't want to fail just because of this
	}

	// Now need to create a separate notification record for each CRN
	list = make([]*datastore.NotificationInsert, 0, len(notice.Locations))

	for _, crn := range notice.Locations {

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
		o.PnPRemoved = n.PnPRemoved
		o.Tags = n.Tags

		o.SourceCreationTime = n.SourceCreationTime
		if len(n.SourceCreationTime) == 0 {
			o.SourceCreationTime = n.SourceUpdateTime // Use source update time as an emergency backup
		}

		o.Category = n.Category
		o.CRNFull = BuildCRN(crn.String(), serviceName)
		o.ResourceDisplayNames = n.ResourceDisplayNames

		list = append(list, o)
	}

	return list, nil
}

// getTags will return tags for the announcement. At this time the only tag is
// one that represents a private audience value. This will be interpreted by the
// pnp-status component.  Added as a tag for now to avoid a database update.
func getTags(notice *gitnotice.Notice) string {

	prefix := "audienceis"

	if len(notice.Audience) > 0 {

		a := strings.ToLower(strings.TrimSpace(notice.Audience))

		switch a {
		case "private":
			log.Println("setting audience tag", prefix+"private")
			return prefix + "private"
		case "public":
			log.Println("setting audience tag", prefix+"public")
			return prefix + "public"
		default:
			log.Println("unrecognized tag", a)
		}
	}

	return ""
}

// gheSourceID returns the sourceID for a git issue. We include the repo information
// in the source ID since just the number would not be unique enough.
func gheSourceID(logTxn string, event *github.IssuesEvent) string {
	FCT := "gheSourceID " + logTxn

	// Repo URL looks like: "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test"

	pathElts := strings.Split(event.Issue.GetRepositoryURL(), "/")
	lp := len(pathElts)

	id := ""
	if lp > 3 {
		id = fmt.Sprintf("%s/%s/%d", pathElts[lp-2], pathElts[lp-1], event.Issue.GetNumber())
		log.Printf("%s: INFO: Assigning source id: %s", FCT, id)
	} else { // Should not occur case
		log.Printf("%s: WARNING: Could not parse repo URL (%s), using basic number as id %d", FCT, event.Issue.GetRepositoryURL(), event.Issue.GetNumber())
		id += strconv.Itoa(event.Issue.GetNumber())
	}

	return id
}

// checkCI will check the Notice to ensure that the CI and the environments
// are ok to include in a notification.  It returns an enhanced error type
// that will include text we can include back to the GHE originator
func checkCI(logTxn string, notice *gitnotice.Notice) *gitnotice.Error {
	FCT := "checkCI " + logTxn

	// make sure the notice.Services slice is not empty - issue https://github.ibm.com/cloud-sre/ToolsPlatform/issues/12105
	if len(notice.Services) == 0 {
		return gitnotice.NewError(errCIMissing, "missing CI in notice")
	}

	ok, err := osscatalog.IsServicePnPEnabled(notice.Services[0])
	if !ok || err != nil {
		log.Printf("%s: Service %s is not PnPEnabled!", FCT, notice.Services[0])
		if err != nil {
			log.Println(FCT, err)
		}
		return gitnotice.NewError(err, "Service %s is not PnP enabled.", notice.Services[0])
	}

	badEnvs := ""
	for _, crn := range notice.Locations {
		if !osscatalog.IsEnvPnPEnabled(crn.String()) {
			badEnvs += crn.String() + ","
		}
	}

	if len(badEnvs) > 0 {
		log.Printf("%s: Bad environments provided for service %s. Environments: %s", FCT, notice.Services[0], badEnvs)
		return gitnotice.NewError(nil, "Incorrect locations provided. Check environment type and status: %s.", badEnvs)
	}

	return nil
}

// sendProblemToGit actually posts a comment back to github that describes
// the issue encountered.  It would be better to do this in nq2ds, but the
// problem is we need to figure out the error here before putting the
// notification on the queue to prevent subscribers from seeing bad notifications
func sendProblemToGit(logTxn string, nError *gitnotice.Error, event *github.IssuesEvent) {
	FCT := "sendProblemToGit " + logTxn

	issueConn, err := getIssueConnForIssue(logTxn, event)
	if err != nil {
		log.Println(FCT+"Failed to get Git issue connection", err)
		return
	}

	issueConn.AddLabels([]string{ErrorLabel})
	var labelList string
	for _, l := range event.GetIssue().Labels {
		labelList += l.GetName() + "\n"
	}
	issueConn.CreateComment("FAILED to process the last update:\nUser performing action: " + event.GetSender().GetLogin() + "\nAction performed: " + event.GetAction() + "\n\n---\nError:\n" + nError.String() + "\n\n---\nContent:\n" + event.GetIssue().GetBody() + "\n\n---\nLabels:\n" + labelList)
	return
}

// sendSuccessToGit sends a message back to the git issue indicating that the
// message was successfully processed
func sendSuccessToGit(logTxn string, event *github.IssuesEvent) {

	FCT := "sendSuccessToGit " + logTxn

	issueConn, _ := removeLabel(logTxn, event, ErrorLabel)
	if issueConn == nil {
		// If label removal failed, that's ok, but if we didn't get the issue connection, that's a problem
		log.Printf("%s: Failed to get issue connection", FCT)
		return
	}

	var labelList string
	for _, l := range event.GetIssue().Labels {
		labelList += l.GetName() + "\n"
	}
	issueConn.CreateComment("Successfully received update in PnP.\n" + "User performing action: " + event.GetSender().GetLogin() + "\nAction performed: " + event.GetAction() + "\n\n---\nContent:\n" + event.GetIssue().GetBody() + "\n\n---\nLabels:\n" + labelList)
}

func removeLabel(logTxn string, event *github.IssuesEvent, label string) (*githubzen.Issue, error) {
	FCT := "removeLabel " + logTxn

	issueConn, err := getIssueConnForIssue(logTxn, event)
	if err != nil {
		log.Println(FCT+"Failed to get Git issue connection", err)
		return nil, err
	}

	return issueConn, issueConn.RemoveLabel(label)
}

func getIssueConnForIssue(logTxn string, event *github.IssuesEvent) (issue *githubzen.Issue, err error) {

	// Extract the URL for the git API only, for example `https://github.ibm.com/api/v3` from a URL like this:
	//"url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues/5"

	gitURL := event.Issue.GetURL()[0:strings.Index(event.Issue.GetURL(), "/repos/")]

	return getGitIssueConnection(logTxn,
		gitAccessToken,
		gitURL,
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		event.Issue.GetNumber())
}

var gitServer *githubzen.Server
var gitRepo *githubzen.Repo
var gitMutex = new(sync.Mutex)

// getGitIssueConnection is used to get a connection to an issue object that can be used to send changes
// back to git.  The issue that we get on the hooks is just JSON, but if we want to set a label or
// comment, we need to authenticate back to the Git server, so this method will return an issue
// upon which we can make calls to make changes in Git on that issue.
// Note this function does assume that this insteance of the GHE hook will only ever deal with
// one git server and one git repo; therefore it caches those values once acquired.
func getGitIssueConnection(logTxn, accessToken, gitURL, owner, name string, number int) (issue *githubzen.Issue, err error) {

	FCT := "getGitIssueConnection " + logTxn

	gitMutex.Lock()
	defer gitMutex.Unlock()

	if gitServer == nil {
		gitServer, err = githubzen.SetupGit(gitURL, accessToken)
		if err != nil {
			log.Printf("%s: Failed to get GIT server: %s  error:%s", FCT, gitURL, err.Error())
			return nil, err
		}
	}

	if gitRepo == nil {
		gitRepo, err = gitServer.SetupRepo(owner, name)
		if err != nil {
			log.Printf("%s: Failed to get GIT repo. owner: %s, name: %s: error: %s", FCT, owner, name, err.Error())
			return nil, err
		}
	}

	issue, err = gitServer.GetIssue(gitRepo, number)
	if err != nil {
		log.Printf("%s: Failed to get GIT issue connection. owner: %s, name: %s: number: %d  error: %s", FCT, owner, name, number, err.Error())
		return nil, err
	}

	return issue, nil

}

// getContext gets the context to use with calls back to the pnp_abstraction. The
// context is pretty simple as it really just carries the NR monitor. We call
// the buildContext method from snowBSPN.go
func getContextForGhe(logTxn string) (ctx ctxt.Context, err error) {

	FCT := "getContextForGhe " + logTxn

	ctx, err = buildContext()
	if err != nil {
		log.Println(FCT, ": ERROR getting context", err)
		return ctxt.Context{}, err
	}
	ctx.LogID = logTxn

	return ctx, nil
}

// getNotificationType determines if the git issue represents an announcement or a security notice
// The label indicates if it is an 'announcement' or 'security' notice. However, there could be some
// different conditions that we meet:
// - Both labels could be present. This is not necessarily a mistake. The issue can be in transition.
// - Neither is provided. Again this could be a transition state.
// We will try to check if the action is a label or unlabel to determine what the type will be. However,
// if we are in a state where we cannot concretely determine the type, we will default to 'announcement'
func getNotificationType(event *github.IssuesEvent) string {
	result := AnnouncementLabel

	if hasLabel(SecurityLabel, event.Issue) && !hasLabel(AnnouncementLabel, event.Issue) {
		result = SecurityLabel
	}

	return result
}

// normalizeServiceName will ensure that a service name reflects its categoryID parent if it exists
func normalizeServiceName(servicename string, context ctxt.Context) (normalizedServiceName string, err error) {
	FCT := "normalizeServiceName"
	normalizedServiceName = servicename

	record, err := osscatalog.ServiceNameToOSSRecord(context, servicename)
	if err != nil {
		log.Println(FCT, "Error getting category ID", err)
		return "", err
	}

	parent := string(record.StatusPage.CategoryParent)

	if parent != "" {
		normalizedServiceName = parent
	}

	return normalizedServiceName, nil
}

// handlePanics will handle panics within the processing thread
func handlePanics(logTxn string) {
	const FCT = "HandlePanics: "
	if err := recover(); err != nil {
		stacktraceBuffer := make([]byte, 4096)
		count := runtime.Stack(stacktraceBuffer, true)

		log.Printf(FCT+"PANIC: "+logTxn, err)
		log.Printf(FCT+"STACKTRACE %s (%d bytes): %s\n", logTxn, count, stacktraceBuffer[:count])
		return
	}
}

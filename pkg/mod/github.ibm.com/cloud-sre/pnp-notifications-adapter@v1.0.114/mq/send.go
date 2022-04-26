package mq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
	"github.ibm.com/cloud-sre/pnp-abstraction/utils"
)

const (
	// EnhancedLogging enables very explicit logging
	EnhancedLogging = true
)

// EnhancedLog is used to create a more descriptive log for debugging
type EnhancedLog struct {
	EventStart       string `json:"eventStartTime"`
	UpdateTime       string `json:"updateTime"`
	ShortDescription string `json:"shortDescription"`
	Type             string `json:"type"`
	DiffField        string `json:"diff,omitempty"`
	DiffNewVal       string `json:"nval,omitempty"`
	DiffOldVal       string `json:"oval,omitempty"`
}

// UTInternalSend Pointer to internal method used for unit testing only
var UTInternalSend = internalSendNotifications

// UTInternalCompareAndSend Pointer to internal method used for unit testing only
var UTInternalCompareAndSend = internalSendCompareAndUpdateNotifications

// SendNotifications will take the input list and send all of the
// notifications on it to the message queue
func (conn *Connection) SendNotifications(ctx ctxt.Context, nList []datastore.NotificationInsert, msgType datastore.NotificationMsgType) (err error) {
	return internalSendNotifications(ctx, conn, nList, msgType)
}

func internalSendNotifications(ctx ctxt.Context, conn IConnection, nList []datastore.NotificationInsert, msgType datastore.NotificationMsgType) (err error) {
	METHOD := "mq.SendNotifications"

	log.Printf("INFO (%s): Sending %d notifications", METHOD, len(nList))

	for i, n := range nList {

		txn := ctx.NRMon.StartTransaction(exmon.MQProduceNotification)
		if msgType == datastore.Update {
			txn.AddCustomAttribute(exmon.Operation, exmon.OperationUpdate)
		} else {
			txn.AddCustomAttribute(exmon.Operation, exmon.OperationInsert)
		}
		txn.AddCustomAttribute(exmon.Kind, exmon.KindNotification)
		txn.AddCustomAttribute(exmon.Source, n.Source)
		txn.AddCustomAttribute(exmon.SourceID, n.SourceID)
		txn.AddIntCustomAttribute(exmon.RecordCount, 1)

		wrap, err := WrapNotification(&nList[i], msgType)
		if err != nil {
			txn.End()
			return fmt.Errorf("ERROR (%s): cannot wrap notification [%s]", METHOD, err.Error())
		}

		buffer := new(bytes.Buffer)
		if err = json.NewEncoder(buffer).Encode(wrap); err != nil {
			txn.End()
			return fmt.Errorf("ERROR (%s): cannot encode notification [%s]", METHOD, err.Error())
		}

		err2 := conn.Produce(buffer.String())

		if err2 != nil {
			txn.AddBoolCustomAttribute(exmon.MQPostFailed, true)
			// TODO: Need to do some kind of retry
			err = err2
		} else {
			txn.AddBoolCustomAttribute(exmon.MQPostFailed, false)
		}

		txn.End()
	}

	return err
}

// SendCompareAndUpdateNotifications will compare an existing list against a new list and only
// send the new list of notifications
func (conn *Connection) SendCompareAndUpdateNotifications(ctx ctxt.Context, existingList, newList []datastore.NotificationInsert) error {
	return internalSendCompareAndUpdateNotifications(ctx, conn, existingList, newList)
}

// InternalSendCompareAndUpdateNotifications is an internally used function to execute send and update.  Exposed externally for
// unit testing purposes.
func internalSendCompareAndUpdateNotifications(ctx ctxt.Context, conn IConnection, existingList, newList []datastore.NotificationInsert) error {

	METHOD := "internalSendCompareAndUpdateNotifications"

	log.Printf("INFO (%s): len(existing)=%d, len(new)=%d\n", METHOD, len(existingList), len(newList))

	output := make([]datastore.NotificationInsert, 0)

	myHashOfExisting := make(map[string]datastore.NotificationInsert)

	// Create hashed table for easy lookup
	for _, e := range existingList {

		myHashOfExisting[fmt.Sprintf("%s+%s+%s", e.Source, e.SourceID, e.CRNFull)] = e
	}

	var notUpdatedLog []EnhancedLog
	var updatedLog []EnhancedLog

	for _, n := range newList {

		updateRequired := true
		diffLabel := "DOES_NOT_EXIST"
		diffNewVal := ""
		diffOldVal := ""

		e, exists := myHashOfExisting[fmt.Sprintf("%s+%s+%s", n.Source, n.SourceID, n.CRNFull)]

		if exists {
			updateRequired = false
			diffLabel = ""
			diffNewVal = ""
			diffOldVal = ""

			updateRequired, diffLabel, diffOldVal, diffNewVal = timeCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "SourceUpdateTime", e.SourceUpdateTime, n.SourceUpdateTime)

			/* [2019-06-28] Previously we made all the below checks, but as it turns out this introduced
			                a problem if there was a change in resource display names in the global catalog.
							The problem resulted from that change happening is this adapter would detect the change
							and send an update to nq2ds. However, nq2ds would look at the record source update time
							which would not have changed, so it would not make the change. Next time this adapter ran
							it would detec the same difference and that cycle continually repeated.  So at this point
							we will only update the record if the source update time changes.  The update time should be
							changed whenever the majority of these changes occur.  The resource display name is the only
							"out-of-record" change that could happen, and if that is the case, then we won't automatically
							pick it up.  But for the most part that only would apply to old notifications, and it's probably
							ok to ignore those.
			// If the existing item has been pnpRemoved from the PNP database, then we only consider the above update time.
			if !e.PnPRemoved {
				updateRequired, diffLabel, diffOldVal, diffNewVal = timeCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "SourceCreationTime", e.SourceCreationTime, n.SourceCreationTime)
				updateRequired, diffLabel, diffOldVal, diffNewVal = timeCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "EventTimeStart", e.EventTimeStart, n.EventTimeStart)
				updateRequired, diffLabel, diffOldVal, diffNewVal = timeCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "EventTimeEnd", e.EventTimeEnd, n.EventTimeEnd)
				updateRequired, diffLabel, diffOldVal, diffNewVal = strCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "Type", e.Type, n.Type)
				updateRequired, diffLabel, diffOldVal, diffNewVal = strCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "Category", e.Category, n.Category)
				updateRequired, diffLabel, diffOldVal, diffNewVal = strCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "IncidentID", e.IncidentID, n.IncidentID)
				updateRequired, diffLabel, diffOldVal, diffNewVal = strCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "CRNFull", e.CRNFull, n.CRNFull)

				updateRequired, diffLabel, diffOldVal, diffNewVal = xsCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "LongDescription", e.LongDescription, n.LongDescription)
				updateRequired, diffLabel, diffOldVal, diffNewVal = xsCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "ShortDescription", e.ShortDescription, n.ShortDescription)
				updateRequired, diffLabel, diffOldVal, diffNewVal = xsCheck(updateRequired, diffLabel, diffOldVal, diffNewVal, "ResourceDisplayNames", e.ResourceDisplayNames, n.ResourceDisplayNames)
			}
			*/

		} else {
			log.Println("DEBUG These are not found!: send.go", n.Source, n.SourceID, n.CRNFull)
		}

		logName := ""
		if len(n.ShortDescription) > 0 {
			logName = n.ShortDescription[0].Name
		}

		if updateRequired {
			output = append(output, n)

			if EnhancedLogging {
				updatedLog = append(updatedLog, EnhancedLog{EventStart: n.EventTimeStart, ShortDescription: logName, Type: n.Type, UpdateTime: n.SourceUpdateTime, DiffField: diffLabel, DiffOldVal: diffOldVal, DiffNewVal: diffNewVal})
			}
		} else {
			if EnhancedLogging {
				notUpdatedLog = append(notUpdatedLog, EnhancedLog{EventStart: n.EventTimeStart, ShortDescription: logName, Type: n.Type, UpdateTime: n.SourceUpdateTime})
			}
		}
	}

	if len(output) > 0 {

		if EnhancedLogging {

			buffer := new(bytes.Buffer)
			if err := json.NewEncoder(buffer).Encode(updatedLog); err == nil {
				log.Println(METHOD + ": UPDATING: " + buffer.String())
			}
			//buffer = new(bytes.Buffer)
			//if err := json.NewEncoder(buffer).Encode(notUpdatedLog); err == nil {
			//	log.Println(METHOD + "NOT UPDATING: " + buffer.String())
			//}

		}

		return conn.SendNotifications(ctx, output, datastore.Update)
	}

	return nil
}

func strCheck(isUpdating bool, defaultLabel, defaultOldVal, defaultNewVal, label, value1, value2 string) (resultUpdating bool, resultLabel, resultOldVal, resultNewVal string) {

	resultUpdating = isUpdating
	resultLabel = defaultLabel
	resultOldVal = defaultOldVal
	resultNewVal = defaultNewVal

	if !isUpdating { // if already updating, then no further check needed
		resultUpdating = value1 != value2
		if resultUpdating {
			resultLabel = label
			resultOldVal = value1
			resultNewVal = value2
		}
	}

	return resultUpdating, resultLabel, resultOldVal, resultNewVal
}

func timeCheck(isUpdating bool, defaultLabel, defaultOldVal, defaultNewVal, label, value1, value2 string) (resultUpdating bool, resultLabel, resultOldVal, resultNewVal string) {

	resultUpdating = isUpdating
	resultLabel = defaultLabel
	resultOldVal = defaultOldVal
	resultNewVal = defaultNewVal

	if !resultUpdating { // if already updating, then no further check needed

		if len(value1) == 0 && len(value2) == 0 {
			return resultUpdating, resultLabel, resultOldVal, resultNewVal
		}

		resultUpdating = !utils.CompareTimeStr(value1, value2)
		if resultUpdating {
			resultLabel = label
			resultOldVal = value1
			resultNewVal = value2
		}
	}

	return resultUpdating, resultLabel, resultOldVal, resultNewVal
}

func xsCheck(isUpdating bool, defaultLabel, defaultOldVal, defaultNewVal, label string, value1, value2 []datastore.DisplayName) (resultUpdating bool, resultLabel, resultOldVal, resultNewVal string) {

	resultUpdating = isUpdating
	resultLabel = defaultLabel
	resultOldVal = defaultOldVal
	resultNewVal = defaultNewVal

	if !isUpdating { // if already updating, then no further check needed

		if len(value1) != len(value2) {
			resultOldVal = fmt.Sprintf("LEN=%d", len(value1))
			resultNewVal = fmt.Sprintf("LEN=%d", len(value2))
			return true, label, resultOldVal, resultNewVal
		}

		lm := make(map[string]string)

		for _, xs := range value1 {
			lm[xs.Language] = xs.Name
		}

		for _, xs := range value2 {
			if lm[xs.Language] != xs.Name {
				resultOldVal = lm[xs.Language]
				resultNewVal = xs.Name
				return true, label, resultOldVal, resultNewVal
			}
		}

	}

	return resultUpdating, resultLabel, resultOldVal, resultNewVal
}

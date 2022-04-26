package testDefs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	encryption "github.ibm.com/cloud-sre/pnp-data-encryption"
	hlp "github.ibm.com/cloud-sre/pnp-nq2ds/helper"
	rabbitmq "github.ibm.com/cloud-sre/pnp-rabbitmq-connector"
	"github.ibm.com/cloud-sre/pnp-rest-test/common"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
	"github.ibm.com/cloud-sre/pnp-rest-test/status"
)

func watchInArray(watches []datastore.WatchReturn, matchString string) bool {
	for _, watch := range watches {
		if watch.RecordID == matchString {
			return true
		}
	}
	return false
}

func cleanup(server rest.Server, sub *datastore.SubscriptionSingleResponse) error {
	const fct = "[cleanup]"
	_, err := server.Delete(fct, sub.SubscriptionURL)
	if err != nil {
		log.Println(fct, err.Error())
		return err
	}
	return nil
}

/*
func dumpWatchMap(w http.ResponseWriter) error {
	const fct = "PostdumpWatchMap2RMQ"
	time.Sleep(1 * time.Second)
	serv := rest.Server{}
	url := dumpWatchMapUrl

	resp, err := serv.Get(fct, url)

	if err != nil {
		Lg(w, fct, "Error = "+err.Error())
		return err
	}
	htdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Lg(w, fct, "Error = "+err.Error())
		return err
	}
	Lg(w, fct, string(htdata))
	return nil

}
*/

// Consumes through the messages channel
// validation updates come in the form of "Validation update|-1|validationString|Description"
// where "Validation update" tells it what kind of message it is
// "-1" indicates whether to append or update a validation string(-1 is append.  An integer is the index to update)
func listenForWebhook(validationStrings []validation, numValidations int) string {
	const fct = "[listenForWebhook]"
	const listenInterval = 5 // listen for messages for 5 minutes

	countValidations := 0

	// We only get the record id created by the system from the status consumer.
	log.Println(fct, "Starting --- numValidations:", numValidations, "validationStrings:\n", printValidationStrings(validationStrings))

	now := time.Now()
	doneString := "done " + now.Format("2006-01-02 15:04:05")

	go func(doneString string) {
		log.Println(fct, "Start sleep for", listenInterval, "minutes")
		time.Sleep(listenInterval * time.Minute)
		log.Println(fct, "Sleep completed :", doneString)
		Messages <- doneString
	}(doneString)

	for msg := range Messages {
		log.Println(fct, "*** Message received :", msg)
		if bodyIndex := strings.Index(msg, "Body"); bodyIndex > 0 {
			log.Println(fct, "Message body:", msg[bodyIndex:])
		}

		if msg == doneString {
			log.Println(fct, "****************************************************")
			log.Println(fct, "FAIL: TIMED OUT")
			log.Println(fct, "--- Found:", countValidations)
			log.Println(fct, "--- Need", len(validationStrings), "more validations")
			log.Println(fct, "****************************************************")
			log.Println(fct, "Validations left:\n", printValidationStrings(validationStrings))

			return FAIL_MSG
		}

		// Validation message is in the form: Validation update|<index>|<validation_value>|<validation_description>
		if strings.Index(msg, "Validation update|") == 0 {
			validationUpdate := strings.Split(msg, "|")
			newValidation := validation{
				Value: validationUpdate[2],
				Desc:  validationUpdate[3],
			}

			updateIndex, err := strconv.Atoi(validationUpdate[1])
			if err != nil {
				log.Println(fct, "Failed to get updateIndex:", err.Error())
			}

			if updateIndex == -1 {
				validationStrings = append(validationStrings, newValidation)
				log.Println(fct, "--- Appended new validation string:\n\t", validationStrings[len(validationStrings)-1])
			} else {
				mutexValidation.Lock()
				validationStrings[updateIndex] = newValidation
				mutexValidation.Unlock()
				log.Println(fct, "--- Updated validation string at index:", updateIndex, "\n\t", validationStrings[updateIndex])
			}

			log.Println(fct, "*********** Validation Strings *******************\n", printValidationStrings(validationStrings))
			continue
		}

		mutexValidation.Lock()
		log.Println(fct, "Checking message against validation strings")
		for i := len(validationStrings) - 1; i >= 0; i-- { // loop backwards through array to avoid indexing a removed value
			v := validationStrings[i]
			if strings.Contains(msg, v.Value) {
				countValidations++
				log.Println(fct, "--- Found validation string:", v.Value, "-", v.Desc)
				log.Println(fct, "countValidations:", countValidations, "numValidations:", numValidations)

				if !isValidBody(msg, v.Value) {
					log.Println(fct, "FAIL: Valid string match but invalid response from status API for", v.Desc)
					mutexValidation.Unlock()
					return FAIL_MSG
				}

				if len(validationStrings) > 1 || countValidations < numValidations {
					validationStrings = append(validationStrings[:i], validationStrings[i+1:]...)
					log.Println(fct, "--- Removed validation string at index:", i)
				}

				log.Println(fct, "---", len(validationStrings), "validation strings left:\n", printValidationStrings(validationStrings))

				if countValidations == numValidations {
					log.Println(fct, "--- All validations have been found. Returning success ---")
					log.Println(fct, "validationStrings:\n", printValidationStrings(validationStrings))
					mutexValidation.Unlock()
					return SUCCESS_MSG
				}
			}
		}
		mutexValidation.Unlock()
	}

	log.Println(fct, "FAIL: NOT ALL VALIDATIONS FOUND:", validationStrings)
	return FAIL_MSG
}

func printValidationStrings(validations []validation) string {
	var printStr string
	for i, v := range validations {
		printStr += fmt.Sprintf("\t[%d] %s - %s\n", i, v.Value, v.Desc)
	}
	return printStr
}

func PostMaintenance2RMQ() error {
	const fct = "[PostMaintenance2RMQ]"
	time.Sleep(1 * time.Second)

	var RMQUrls []string
	var p1 *rabbitmq.AMQPProducer
	if isTargetMessagesForRabbitMQ() {
		RMQUrls = append(RMQUrls, RMQAMQPSEndpoint)
		p1 = rabbitmq.NewSSLProducer(RMQUrls, RMQTLSCert, RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	} else {
		RMQUrls = append(RMQUrls, RMQUrl)
		p1 = rabbitmq.NewProducer(RMQUrls, RMQRoutingKeyMaintenance, exchangeName, exchangeType)
	}
	log.Println(fct, RMQUrls, RMQRoutingKeyMaintenance, exchangeName, exchangeType)

	encMaintMsg, err := encryption.Encrypt(CreateMaintenanceMessageinRMQ)
	if err != nil {
		return err
	}

	err = p1.ProduceOnce(string(encMaintMsg))
	if err != nil {
		return err
	}

	log.Println(fct, "Posted maintenance message to Rabbit MQ")
	return nil
}

// Finds href in the JSON body of the msg parameter
// Calls server using the href to get a response
// Checks that the response href contains the validationString parameter
// Checks that the response contains a recordID
func isValidBody(msg, validationString string) bool {
	const fct = "[isValidBody]"
	log.Println(fct, "--- Validating msg:", msg, "\n\tvalidation string:", validationString)
	time.Sleep(5 * time.Second)

	hrefString := strings.Split(msg, "Body :")[1]
	href := common.Href{}
	err := json.Unmarshal([]byte(hrefString), &href)
	if err != nil {
		log.Println(fct, "FAILED to unmarshal:", err.Error())
		return false
	}

	server := &rest.Server{}
	server.Token = os.Getenv("SERVER_KEY")

	incident := new(status.IncidentGet)
	loopCount := 0
	maxRetry := 10
	for loopCount < maxRetry {
		err := server.GetAndDecode(fct, "IncidentGet", href.Href, incident)
		if err != nil {
			log.Println(fct, err.Error())
			time.Sleep(3 * time.Second)
			loopCount++
			log.Println(fct, fmt.Sprintf("Retrying...(%d/%d)", loopCount, maxRetry))
		} else {
			loopCount = maxRetry
		}
	}

	log.Println(fct, "Response body:", hlp.GetPrettyJson(incident))
	if incident.Href == "" {
		log.Println(fct, "Returned empty href")
		return false
	}
	if !strings.Contains(incident.Href, validationString) {
		log.Println(fct, "Returned invalid href:", incident.Href)
		return false
	}

	if incident.RecordID == "" {
		log.Println(fct, "Returned empty recordID")
		return false
	}

	log.Println(fct, "Successfully validated:", validationString)
	return true
}

func isTargetMessagesForRabbitMQ() bool {
	return RMQEnableMessages == "true"
}

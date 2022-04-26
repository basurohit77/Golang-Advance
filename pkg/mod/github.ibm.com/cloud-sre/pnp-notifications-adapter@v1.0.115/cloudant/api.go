package cloudant

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/exmon"
)

// GetNotifications will reach out to cloudant to pull the notifications
func GetNotifications(ctx ctxt.Context, url, user, password string) (result *NotificationsResult, err error) {

	METHOD := "cloudant.api.GetNotifications"

	result = getNotificationsCache(url)
	if result != nil {
		return result, nil
	}

	txn := ctx.NRMon.StartTransaction(exmon.CloudantGetNotifications)
	defer txn.End()
	txn.AddCustomAttribute(exmon.Operation, exmon.OperationQuery)

	resp, err := GetFromCloudant(METHOD, http.MethodGet, url, user, password)
	if err != nil {
		txn.AddBoolCustomAttribute(exmon.CloudantFailed, true)
		return nil, err
	}
	txn.AddBoolCustomAttribute(exmon.CloudantFailed, false)

	// Make sure to close the body when we are all done
	defer resp.Body.Close()

	// Decode the json result into a struct
	cloudantRecords := new(NotificationsResult)
	if err := json.NewDecoder(resp.Body).Decode(cloudantRecords); err != nil {

		msg := fmt.Sprintf("ERROR (%s): Failed to decode the result from Cloudant: %s", METHOD, err.Error())
		log.Println(msg)
		txn.AddBoolCustomAttribute(exmon.ParseFailed, true)
		return nil, errors.New(msg)

	}

	txn.AddBoolCustomAttribute(exmon.ParseFailed, false)
	txn.AddIntCustomAttribute(exmon.RecordCount, len(cloudantRecords.Rows))

	setNotificationsCache(cloudantRecords, url)
	return cloudantRecords, nil
}

// GetNameMapping retrieves a name map from Cloudant
func GetNameMapping(url, user, password string) (result *NameMapping, err error) {

	METHOD := "cloudant.api.GetNameMapping"

	result = getNameMappingCache(url)
	if result != nil {
		return result, nil
	}

	resp, err := GetFromCloudant(METHOD, http.MethodGet, url, user, password)
	if err != nil {
		return nil, err
	}

	// Make sure to close the body when we are all done
	defer resp.Body.Close()

	// Decode the json result into a struct
	cloudantRecords := new(NameMapping)
	if err := json.NewDecoder(resp.Body).Decode(cloudantRecords); err != nil {

		msg := fmt.Sprintf("ERROR (%s): Failed to decode the result from Cloudant: %s", METHOD, err.Error())
		log.Println(msg)
		return nil, errors.New(msg)

	}

	setNameMappingCache(cloudantRecords, url)
	return cloudantRecords, nil
}

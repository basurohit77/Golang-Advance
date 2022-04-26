package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	pq "github.com/lib/pq"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

// UpdateSubscriptionStatement updates subscription_table by record id value
func UpdateSubscriptionStatement(dbConnection *sql.DB, itemToInsert *datastore.SubscriptionInsert, recordID string) (*datastore.SubscriptionInsert, error, int) {

	query := "UPDATE " + SUBSCRIPTION_TABLE_NAME + " SET "

	var vals []interface{}
	if itemToInsert.Expiration != "" {
		vals = append(vals, itemToInsert.Expiration)
		query += SUBSCRIPTION_COLUMN_EXPIRATION + "=$" + strconv.Itoa(len(vals))
	}
	if itemToInsert.TargetToken != "" {
		vals = append(vals, itemToInsert.TargetToken)
		if len(vals) > 1 {
			query += ", "
		}
		query += SUBSCRIPTION_COLUMN_TARGET_TOKEN + "=$" + strconv.Itoa(len(vals))
	}

	if itemToInsert.TargetAddress != "" {
		vals = append(vals, itemToInsert.TargetAddress)
		if len(vals) > 1 {
			query += ", "
		}
		query += SUBSCRIPTION_COLUMN_TARGET_ADDRESS + "=$" + strconv.Itoa(len(vals))
	}
	if itemToInsert.Name != "" {
		vals = append(vals, itemToInsert.Name)
		if len(vals) > 1 {
			query += ", "
		}
		query += SUBSCRIPTION_COLUMN_NAME + "=$" + strconv.Itoa(len(vals))
	}
	vals = append(vals, recordID)
	query += " WHERE " + SUBSCRIPTION_COLUMN_RECORD_ID + "=$" + strconv.Itoa(len(vals))

	if len(vals) == 1 {
		log.Println(tlog.Log() + "Nothing to update")
		return &datastore.SubscriptionInsert{}, errors.New("Nothing to update " + fmt.Sprintf("%+v", itemToInsert)), http.StatusOK
	}
	log.Println(tlog.Log()+"DEBUG: Query: ", query)
	log.Println(tlog.Log()+"Values:", vals)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		_, err := dbConnection.Exec(query, vals...)

		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+"Error: ", err.Error())
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return itemToInsert, err, http.StatusInternalServerError
			}

		} else {
			return itemToInsert, err, http.StatusOK
		}
	}
	return itemToInsert, nil, http.StatusOK
}

// UpdateIncident get the metadata of the incident and added to the database
func UpdateIncident(database *sql.DB, itemToInsert *datastore.IncidentInsert) (error, int) {

	if strings.TrimSpace(itemToInsert.Source) == "" {
		return errors.New(ERR_NO_SOURCE), http.StatusBadRequest
	}

	if strings.TrimSpace(itemToInsert.SourceID) == "" {
		return errors.New(ERR_NO_SOURCEID), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", itemToInsert.Source, "SourceID: ", itemToInsert.SourceID)

	if CheckEnumFields(itemToInsert.State, "new", "in-progress", "resolved") == false {
		return errors.New(ERR_BAD_STATE), http.StatusBadRequest
	}

	if CheckEnumFields(itemToInsert.Classification, "confirmed-cie", "potential-cie", "normal") == false {
		return errors.New(ERR_BAD_CLASSIFICATION), http.StatusBadRequest
	}

	if !itemToInsert.PnPRemoved { // enforce CRN only when PnPRemoved flag is false
		if itemToInsert.CRNFull == nil || len(itemToInsert.CRNFull) == 0 {
			return errors.New(ERR_NO_CRN), http.StatusBadRequest
		}
	}

	err := VerifyCRNArray(itemToInsert.CRNFull)
	if err != nil {
		return err, http.StatusBadRequest
	}

	recordID := CreateRecordIDFromSourceSourceID(itemToInsert.Source, itemToInsert.SourceID)
	log.Println(tlog.Log()+"recordID:", recordID, "source:", itemToInsert.Source, "sourceID:", itemToInsert.SourceID)

	incidentGet, err, _ := GetIncidentByRecordIDStatement(database, recordID)
	if err == sql.ErrNoRows {
		msg := "No incident with source '" + itemToInsert.Source + "', and source_id '" + itemToInsert.SourceID + "' found. No record is updated."
		log.Println(tlog.Log() + msg)
		return errors.New(msg), http.StatusBadRequest
	}

	incidentGetRecordID := incidentGet.RecordID

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting update transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}

		err = updateIncidentStatement(tx, itemToInsert, recordID)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				log.Println(tlog.Log(), err)
				return err, http.StatusInternalServerError
			}
		}

		incidentJuncGet, errA, _ := GetIncidentJunctionByIncidentID(database, incidentGetRecordID)
		if errA != nil {
			return errA, http.StatusInternalServerError
		}
		incidentJuncMap := map[string]string{}
		for _, ij := range *incidentJuncGet {
			incidentJuncMap[ij.ResourceID] = ij.IncidentID
		}

		// Check that IncidentJunction table is in sync with update
		for i := 0; i < len(itemToInsert.CRNFull); i++ {

			// change crn filter if it is a Gaas Resource
			queryStr, err, statusCode := CreateCrnFilter(itemToInsert.CRNFull[i], "")
			if err != nil {
				log.Println(tlog.Log(), err)
				return err, statusCode
			}

			resources, err1, rc := GetResourceByQuerySimple(database, queryStr)
			if err1 != nil {
				if rc == http.StatusBadRequest {
					//programming error in caller
					return err1, http.StatusNotImplemented
				}
				log.Println(tlog.Log()+"Error getting resource_id: ", err1)
				//Just log the error for now
			}
			for i := 0; i < len(*resources); i++ {
				//Check if already exists in table
				_, ok := incidentJuncMap[(*resources)[i].RecordID]
				if !ok {
					_, err1 = insertIncidentJunctionStatement(tx, (*resources)[i].RecordID, incidentGetRecordID)
					if err1 != nil {
						log.Println(tlog.Log()+"Error inserting junction: ", err1)
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							return err1, http.StatusInternalServerError
						}
					}
				} else {
					delete(incidentJuncMap, (*resources)[i].RecordID)
				}
			}
			if restartTransaction {
				break
			}
		}
		if restartTransaction {
			// tx has already rollback
			Delay()
			continue
		}

		//delete any remaining incidentjunctionget
		if len(incidentJuncMap) > 0 {
			for r, i := range incidentJuncMap {
				err2 := deleteIncidentJunctionStatement(tx, r, i)
				if err2 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err2) {
						restartTransaction = true
						break
					} else {
						return err2, http.StatusInternalServerError
					}
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}
		}

		// commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Println(tlog.Log()+"Error in Commit: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	} //for nretry
	return err, http.StatusOK

}

func updateIncidentStatement(tx *sql.Tx, itemToInsert *datastore.IncidentInsert, recordID string) error {
	log.Println(tlog.Log()+"recordID:", recordID)

	// Check if there is a reference to `[targeted notification](URL)` in the incoming
	// itemToInsert.
	// If there is a URL defined, parse it and store it into TargetedURL.
	// If there isn't a URL defined, the function returns ErrLongDescNoMatch.
	// This error is logged as a warning.

	// Targeted URL is not longer part of the long description it is not set in a field called u_targeted_notification_url passed from SN
	// if itemToInsert.TargetedURL == "" {
	// 	tURL, err := targurl.URLFromLongDescription(itemToInsert.LongDescription)
	// 	if err != nil {
	// 		log.Println(tlog.Log()+" Warn:", err)
	// 	}

	// 	itemToInsert.TargetedURL = tURL
	// }

	if itemToInsert.Audience == "" || len(itemToInsert.Audience) == 0 {
		itemToInsert.Audience = SNnill2PnP
	}

	// Alex Torres R 2020-12-18 based in Gabriel's code
	// Deal with inserts with a customer description field longer than the DB field (4k)
	if len(itemToInsert.CustomerImpactDescription) >= 4000 {
		itemToInsert.CustomerImpactDescription = itemToInsert.CustomerImpactDescription[:3999]
	}

	// Source and SourceID cannot be changed
	result, err := tx.Exec("UPDATE "+INCIDENT_TABLE_NAME+" SET "+
		INCIDENT_COLUMN_PNP_UPDATE_TIME+"=$1,"+
		INCIDENT_COLUMN_SOURCE_CREATION_TIME+"=$2,"+
		INCIDENT_COLUMN_SOURCE_UPDATE_TIME+"=$3,"+
		INCIDENT_COLUMN_START_TIME+"=$4,"+
		INCIDENT_COLUMN_END_TIME+"=$5,"+
		INCIDENT_COLUMN_SHORT_DESCRIPTION+"=$6,"+
		INCIDENT_COLUMN_LONG_DESCRIPTION+"=$7,"+
		INCIDENT_COLUMN_STATE+"=$8,"+
		INCIDENT_COLUMN_CLASSIFICATION+"=$9,"+
		INCIDENT_COLUMN_SEVERITY+"=$10,"+
		INCIDENT_COLUMN_CRN_FULL+"=$11,"+
		INCIDENT_COLUMN_REGULATORY_DOMAIN+"=$12,"+
		INCIDENT_COLUMN_AFFECTED_ACTIVITY+"=$13,"+
		INCIDENT_COLUMN_CUSTOMER_IMPACT_DESCRIPTION+"=$14,"+
		INCIDENT_COLUMN_PNP_REMOVED+"=$15,"+
		INCIDENT_COLUMN_TARGETED_URL+"=$16,"+
		INCIDENT_COLUMN_AUDIENCE+"=$17"+
		" WHERE "+INCIDENT_COLUMN_RECORD_ID+"=$18",
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		itemToInsert.SourceCreationTime,
		NewNullString(itemToInsert.SourceUpdateTime),
		NewNullString(itemToInsert.OutageStartTime),
		NewNullString(itemToInsert.OutageEndTime),
		NewNullString(itemToInsert.ShortDescription),
		NewNullString(itemToInsert.LongDescription),
		itemToInsert.State,
		itemToInsert.Classification,
		itemToInsert.Severity,
		pq.Array(itemToInsert.CRNFull),
		NewNullString(itemToInsert.RegulatoryDomain),
		NewNullString(itemToInsert.AffectedActivity),
		NewNullString(itemToInsert.CustomerImpactDescription),
		strconv.FormatBool(itemToInsert.PnPRemoved),
		NewNullString(itemToInsert.TargetedURL),
		NewNullString(itemToInsert.Audience),
		recordID)

	if err != nil {
		// failed to execute SQL statements. Rollback
		if result != nil {
			rowsAffected, err2 := result.RowsAffected()
			if err2 != nil {
				log.Println(tlog.Log()+"Error getting rows affected: ", err2)
			}
			log.Println(tlog.Log()+"Error: ", err, rowsAffected)
		}
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in updateIncidentStatement Rollback: ", err1)
		}
	}
	return err

}

// UpdateMaintenance - update the maintenance record with new data
func UpdateMaintenance(database *sql.DB, itemToInsert *datastore.MaintenanceInsert) (error, int) {
	if strings.TrimSpace(itemToInsert.Source) == "" {
		return errors.New("Source is not valid"), http.StatusBadRequest
	}

	if strings.TrimSpace(itemToInsert.SourceID) == "" {
		return errors.New("SourceId is not valid"), http.StatusBadRequest
	}

	if CheckEnumFields(itemToInsert.State, "new", "scheduled", "in-progress", "complete") == false {
		return errors.New("State is not valid"), http.StatusBadRequest
	}

	if itemToInsert.CRNFull == nil || len(itemToInsert.CRNFull) == 0 {
		return errors.New("MaintenanceInsert.CRNFull cannot be nil or empty"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", itemToInsert.Source, "SourceID: ", itemToInsert.SourceID)

	err := VerifyCRNArray(itemToInsert.CRNFull)
	if err != nil {
		log.Println(tlog.Log(), err)
		return err, http.StatusBadRequest
	}

	recordID := CreateRecordIDFromSourceSourceID(itemToInsert.Source, itemToInsert.SourceID)

	maintGet, err, _ := GetMaintenanceByRecordIDStatement(database, recordID)
	if err == sql.ErrNoRows {
		msg := "No maintenance with source '" + itemToInsert.Source + "', and source_id '" + itemToInsert.SourceID + "' found. No record is updated."
		log.Println(tlog.Log(), msg)
		return errors.New(msg), http.StatusBadRequest
	}

	maintGetRecordID := maintGet.RecordID

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting update transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}

		err = updateMaintenanceStatement(tx, itemToInsert, recordID)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				log.Println(tlog.Log(), err)
				return err, http.StatusInternalServerError
			}

		}

		maintJuncGet, errA, _ := GetMaintenanceJunctionByMaintenanceID(database, maintGetRecordID)
		if errA != nil {
			return errA, http.StatusInternalServerError
		}
		maintJuncMap := map[string]string{}
		for _, mj := range *maintJuncGet {
			maintJuncMap[mj.ResourceID] = mj.MaintenanceID
		}

		//Insert into MaintenanceJunction table
		//First get the resourceID from crnFull
		for i := 0; i < len(itemToInsert.CRNFull); i++ {

			// change crn filter if it is a Gaas Resource
			queryStr, err, statusCode := CreateCrnFilter(itemToInsert.CRNFull[i], "")
			if err != nil {
				return err, statusCode
			}

			resources, err1, rc := GetResourceByQuerySimple(database, queryStr)
			if err1 != nil {
				if rc == http.StatusBadRequest {
					//programming error in caller
					return err1, http.StatusNotImplemented
				}
				log.Println(tlog.Log()+"Error getting resource_id: ", err1)
				//Log error for now
			}
			for i := 0; i < len(*resources); i++ {
				//Check if already exists in table
				_, ok := maintJuncMap[(*resources)[i].RecordID]
				if !ok {
					_, err1 = insertMaintenanceJunctionStatement(tx, (*resources)[i].RecordID, maintGetRecordID)
					if err1 != nil {
						log.Println(tlog.Log()+"Error inserting junction: ", err1)
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							return err1, http.StatusInternalServerError
						}
					}
				} else {
					delete(maintJuncMap, (*resources)[i].RecordID)
				}
			}
			if restartTransaction {
				break
			}
		}
		if restartTransaction {
			// tx has already rollback
			Delay()
			continue
		}
		//delete any remaining maintenacejunctionget
		if len(maintJuncMap) > 0 {
			for r, m := range maintJuncMap {
				err2 := deleteMaintenanceJunctionStatement(tx, r, m)
				if err2 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err2) {
						// tx has already rollback
						Delay()
						continue
					} else {
						return err2, http.StatusInternalServerError
					}
				}
			}
		}
		// commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Println(tlog.Log()+"Error in Commit: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}
	return err, http.StatusOK
}

func updateMaintenanceStatement(tx *sql.Tx, itemToInsert *datastore.MaintenanceInsert, recordID string) error {
	log.Println(tlog.Log()+"recordID:", recordID)

	itemToInsert.RecordHash = ComputeMaintenanceRecordHash(itemToInsert)

	// Check if there is a reference to `[targeted notification](URL)` in the incoming
	// itemToInsert.
	// If there is a URL defined, parse it and store it into TargetedURL.
	// If there isn't a URL defined, the function returns ErrLongDescNoMatch.
	// This error is logged as a warning.

	// if itemToInsert.TargetedURL == "" {
	// 	tURL, err := targurl.URLFromLongDescription(itemToInsert.LongDescription)
	// 	if err != nil {
	// 		log.Println(tlog.Log()+" Warn:", err)
	// 	}

	// 	itemToInsert.TargetedURL = tURL
	// }

	if itemToInsert.Audience == "" || len(itemToInsert.Audience) == 0 {
		itemToInsert.Audience = SNnill2PnP
	}

	// Source and SourceID cannot be changed
	result, err := tx.Exec("UPDATE "+MAINTENANCE_TABLE_NAME+" SET "+
		MAINTENANCE_COLUMN_PNP_UPDATE_TIME+"=$1,"+
		MAINTENANCE_COLUMN_SOURCE_CREATION_TIME+"=$2,"+
		MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME+"=$3,"+
		MAINTENANCE_COLUMN_START_TIME+"=$4,"+
		MAINTENANCE_COLUMN_END_TIME+"=$5,"+
		MAINTENANCE_COLUMN_SHORT_DESCRIPTION+"=$6,"+
		MAINTENANCE_COLUMN_LONG_DESCRIPTION+"=$7,"+
		MAINTENANCE_COLUMN_CRN_FULL+"=$8,"+
		MAINTENANCE_COLUMN_STATE+"=$9,"+
		MAINTENANCE_COLUMN_DISRUPTIVE+"=$10,"+
		MAINTENANCE_COLUMN_RECORD_HASH+"=$11,"+
		MAINTENANCE_COLUMN_MAINTENANCE_DURATION+"=$12,"+
		MAINTENANCE_COLUMN_DISRUPTION_TYPE+"=$13,"+
		MAINTENANCE_COLUMN_DISRUPTION_DESCRIPTION+"=$14,"+
		MAINTENANCE_COLUMN_DISRUPTION_DURATION+"=$15,"+
		MAINTENANCE_COLUMN_REGULATORY_DOMAIN+"=$16,"+
		MAINTENANCE_COLUMN_PNP_REMOVED+"=$17,"+
		MAINTENANCE_COLUMN_COMPLETION_CODE+"=$18,"+
		MAINTENANCE_COLUMN_TARGETED_URL+"=$19,"+
		MAINTENANCE_COLUMN_AUDIENCE+"=$20"+
		" WHERE "+MAINTENANCE_COLUMN_RECORD_ID+" =$21",
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		itemToInsert.SourceCreationTime,
		NewNullString(itemToInsert.SourceUpdateTime),
		NewNullString(itemToInsert.PlannedStartTime),
		NewNullString(itemToInsert.PlannedEndTime),
		NewNullString(itemToInsert.ShortDescription),
		NewNullString(itemToInsert.LongDescription),
		pq.Array(itemToInsert.CRNFull),
		itemToInsert.State,
		strconv.FormatBool(itemToInsert.Disruptive),
		NewNullString(itemToInsert.RecordHash),
		itemToInsert.MaintenanceDuration,
		NewNullString(itemToInsert.DisruptionType),
		NewNullString(itemToInsert.DisruptionDescription),
		itemToInsert.DisruptionDuration,
		NewNullString(itemToInsert.RegulatoryDomain),
		strconv.FormatBool(itemToInsert.PnPRemoved),
		NewNullString(itemToInsert.CompletionCode),
		NewNullString(itemToInsert.TargetedURL),
		NewNullString(itemToInsert.Audience),
		recordID)

	if err != nil {
		// failed to execute SQL statements. Rollback
		if result != nil {
			rowsAffected, err2 := result.RowsAffected()
			if err2 != nil {
				log.Println(tlog.Log()+"Error getting rows affected: ", err2)
			}
			log.Println(tlog.Log()+"Error: ", err, rowsAffected)
		}
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in updateMaintenanceStatement Rollback: ", err1)
		}
	}
	return err
}

// UpdateResource Similar to InsertResource except with steps to ensure DisplayName, VisibilityJunction, and Member tables are in sync with changes
// After updating Resource table, it will get all existing records from DisplayName table with Resource recordId and store it in a map.
// Then as it loops through the resourceToInsert.DisplayName array, it will either insert a new DisplayName record if it did not exist before
// or it will remove the particular DisplayName record from the map if DisplayName exists already. After loop exits, it will delete any DisplayName
// records that is left in the map.
// Similar functionality is performed for the VisibilityJunction tables and Member tables
func UpdateResource(database *sql.DB, resourceToInsert *datastore.ResourceInsert) (error, int) {

	if strings.TrimSpace(resourceToInsert.Source) == "" {
		return errors.New("Source is not valid"), http.StatusBadRequest
	}

	if strings.TrimSpace(resourceToInsert.SourceID) == "" {
		return errors.New("SourceId is not valid"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", resourceToInsert.Source, "SourceID: ", resourceToInsert.SourceID, "Crn: ", resourceToInsert.CRNFull)

	//if CheckEnumFields(resourceToInsert.State, "ok", "archived") == false {
	//	return errors.New("State is not valid"), http.StatusBadRequest
	//}

	//if CheckEnumFields(resourceToInsert.OperationalStatus, "none", "ga", "experiment", "deprecated") == false {
	//	return errors.New("OperationalStatus is not valid"), http.StatusBadRequest
	//}

	//if CheckEnumFields(resourceToInsert.Status, "ok", "degraded", "failed", "maintenance") == false {
	//	return errors.New("Status is not valid"), http.StatusBadRequest
	//}

	// check DisplayName
	if resourceToInsert.DisplayNames != nil && checkDisplayNames(resourceToInsert.DisplayNames) == false {
		return errors.New("One or more DisplayName.Name or DisplayName.Language have an empty string"), http.StatusBadRequest
	}

	// check Tags
	if resourceToInsert.Tags != nil && checkTags(resourceToInsert.Tags) == false {
		return errors.New("One or more Tag.ID have an empty string"), http.StatusBadRequest
	}

	// store crn in lowercase
	resourceToInsert.CRNFull = strings.ToLower(resourceToInsert.CRNFull)

	// if resourceToInsert.CRNFull is not IBM Public Cloud, then only save the resource as "crn:v1::service-name:::::" for Gaas
	isGaasResource, crnFull, err0 := ConvertToResourceLookupCRN(resourceToInsert.CRNFull)
	log.Println(tlog.Log()+"crnFull: ", resourceToInsert.CRNFull, ", isGaasResource: ", isGaasResource, ", crnFull is converted to "+crnFull)
	if err0 != nil {
		log.Println(tlog.Log()+"ERROR: ConvertToResourceLookupCRN returns error: ", err0)
		return err0, http.StatusBadRequest
	} else if isGaasResource {
		resourceToInsert.CRNFull = crnFull
	}

	//crnStruct, ok := ParseCRNString(resourceToInsert.CRNFull)
	crnStruct, err := crn.ParseAll(resourceToInsert.CRNFull)

	if err != nil {
		msg := "resourceToInsert.CRNFull has incorrect format."
		log.Println(tlog.Log(), msg)
		return errors.New(msg), http.StatusBadRequest
	}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting update transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return err, http.StatusInternalServerError
			}
		}

		resourceRecordID := CreateRecordIDFromSourceSourceID(resourceToInsert.Source, resourceToInsert.SourceID)

		err = updateResourceStatement(tx, resourceToInsert, crnStruct, resourceRecordID)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				log.Println(tlog.Log(), err)
				return err, http.StatusInternalServerError
			}
		}

		// insert to Display Name table
		dispNameMap := map[string]string{}
		// get DisplayNames currently in DB for Resource
		displayNames, err1 := getDisplayNamesByResourceIDStatement(database, resourceRecordID)
		// ignore error and continue
		if err1 == nil && displayNames != nil {
			for _, dn := range *displayNames {
				dispNameMap[dn.Name] = dn.Language
			}
		}
		if resourceToInsert.DisplayNames != nil && len(resourceToInsert.DisplayNames) > 0 {
			for i := range resourceToInsert.DisplayNames {
				// to fix Go scan error G601 (CWE-118): Implicit memory aliasing in for loop
				lang, ok := dispNameMap[resourceToInsert.DisplayNames[i].Name]
				if !ok || lang != resourceToInsert.DisplayNames[i].Language {
					_, err1 := insertDisplayNameStatement(tx, &resourceToInsert.DisplayNames[i], resourceRecordID)
					if err1 != nil {
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							return err1, http.StatusInternalServerError
						}
					}
				} else {
					delete(dispNameMap, resourceToInsert.DisplayNames[i].Name)
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}

		}
		//Delete any remaining DisplayNames
		if len(dispNameMap) > 0 {
			for n, l := range dispNameMap {
				//deletevisibilibity
				err2 := deleteDisplayNameStatement(tx, resourceRecordID, n, l)
				if err2 != nil {
					return err2, http.StatusInternalServerError
				}
			}
		}

		// insert to Visibility and Visibility Junction table
		visibilityJuncMap := map[string]string{}
		visibilityJuncGet, err1 := getVisibilityJunctionByResourceIDStatement(database, resourceRecordID)
		// ignore error and continue
		if err1 == nil && visibilityJuncGet != nil {
			for _, vj := range visibilityJuncGet {
				visibilityJuncMap[vj.VisibilityID] = vj.ResourceID
			}
		}

		var visibilityRecordID string
		if resourceToInsert.Visibility != nil {

			for _, v := range resourceToInsert.Visibility {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				visibilityGet, err1, rc := getVisibilityByNameStatement(database, v)
				if err1 != nil {
					if rc == http.StatusOK {
						// cannot find visibility, insert it
						visibility := datastore.VisibilityInsert{
							Name:        v,
							Description: v,
						}
						visibilityRecordID, err1 = insertVisibilityStatement(tx, &visibility)
						if err1 != nil {
							if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
								restartTransaction = true
								break
							} else {
								return err1, http.StatusInternalServerError
							}
						}

					} else {
						log.Println(tlog.Log(), err1)
						return err1, http.StatusInternalServerError
					}
				} else {
					visibilityRecordID = visibilityGet.RecordID
				}

				//check if already inserted in visibilityjunction table
				_, ok := visibilityJuncMap[visibilityRecordID]
				if !ok {
					// insert record to VisibilityJunction table
					_, err1 = insertVisibilityJunctionStatement(tx, resourceRecordID, visibilityRecordID)
					if err1 != nil {
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							log.Println(tlog.Log(), err1)
							return err1, http.StatusInternalServerError
						}
					}
				} else {
					delete(visibilityJuncMap, visibilityRecordID)
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}
		}

		//Delete any remaining VisibilityJunction
		if len(visibilityJuncMap) > 0 {
			for v, r := range visibilityJuncMap {
				//deletevisibilibity
				err2 := deleteVisibilityJunctionStatement(tx, r, v)
				if err2 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err2) {
						restartTransaction = true
						break
					} else {
						log.Println(tlog.Log(), err2)
						return err2, http.StatusInternalServerError
					}
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}

		}

		// insert to Tag and Tag Junction table
		tagJuncMap := map[string]string{}
		tagJuncGet, err1 := getTagJunctionByResourceIDStatement(database, resourceRecordID)
		// ignore error and continue
		if err1 == nil && tagJuncGet != nil {
			for _, tj := range tagJuncGet {
				tagJuncMap[tj.TagID] = tj.ResourceID
			}
		}

		var tagRecordID string
		if resourceToInsert.Tags != nil {

			for _, t := range resourceToInsert.Tags {
				t.ID = strings.TrimSpace(t.ID)
				if t.ID == "" {
					continue
				}
				tagGet, err1, rc := getTagByRecordIDStatement(database, CreateRecordIDFromString(t.ID))
				if err1 != nil {
					if rc == http.StatusOK {
						// cannot find visibility, insert it
						tag := datastore.TagInsert{
							ID: t.ID,
						}
						tagRecordID, err1 = insertTagStatement(tx, &tag)
						if err1 != nil {
							if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
								restartTransaction = true
								break
							} else {
								return err1, http.StatusInternalServerError
							}
						}

					} else {
						log.Println(tlog.Log(), err1)
						return err1, http.StatusInternalServerError
					}
				} else {
					tagRecordID = tagGet.RecordID
				}

				//check if already inserted in tagjunction table
				_, ok := tagJuncMap[tagRecordID]
				if !ok {
					// insert record to TagJunction table
					_, err1 = insertTagJunctionStatement(tx, resourceRecordID, tagRecordID)
					if err1 != nil {
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							log.Println(tlog.Log(), err1)
							return err1, http.StatusInternalServerError
						}
					}
				} else {
					delete(tagJuncMap, tagRecordID)
				}

			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}
		}

		//Delete any remaining TagJunction
		if len(tagJuncMap) > 0 {
			for t, r := range tagJuncMap {
				//delete tag
				err2 := deleteTagJunctionStatement(tx, r, t)
				if err2 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err2) {
						restartTransaction = true
						break
					} else {
						log.Println(tlog.Log(), err2)
						return err2, http.StatusInternalServerError
					}
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}
		}

		// commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Println(tlog.Log()+"Error in Commit: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				log.Println(tlog.Log(), err)
				return err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}
	return nil, http.StatusOK
}

func updateResourceStatement(tx *sql.Tx, resourceToInsert *datastore.ResourceInsert, crnStruct crn.MaskAll, recordID string) error {
	log.Println(tlog.Log()+"recordID:", recordID)
	log.Println(tlog.Log()+"rcrnStruct:", crnStruct)

	resourceToInsert.RecordHash = ComputeResourceRecordHash(resourceToInsert)

	// Source and SourceID cannot be changed
	result, err := tx.Exec("UPDATE "+RESOURCE_TABLE_NAME+" SET "+
		RESOURCE_COLUMN_PNP_UPDATE_TIME+"=$1,"+
		RESOURCE_COLUMN_SOURCE_CREATION_TIME+"=$2,"+
		RESOURCE_COLUMN_SOURCE_UPDATE_TIME+"=$3,"+
		RESOURCE_COLUMN_CRN_FULL+"=$4,"+
		RESOURCE_COLUMN_STATE+"=$5,"+
		RESOURCE_COLUMN_OPERATIONAL_STATUS+"=$6,"+
		RESOURCE_COLUMN_STATUS+"=$7,"+
		RESOURCE_COLUMN_STATUS_UPDATE_TIME+"=$8,"+
		RESOURCE_COLUMN_REGULATORY_DOMAIN+"=$9,"+
		RESOURCE_COLUMN_CATEGORY_ID+"=$10,"+
		RESOURCE_COLUMN_CATEGORY_PARENT+"=$11,"+
		RESOURCE_COLUMN_VERSION+"=$12,"+
		RESOURCE_COLUMN_CNAME+"=$13,"+
		RESOURCE_COLUMN_CTYPE+"=$14,"+
		RESOURCE_COLUMN_SERVICE_NAME+"=$15,"+
		RESOURCE_COLUMN_LOCATION+"=$16,"+
		RESOURCE_COLUMN_SCOPE+"=$17,"+
		RESOURCE_COLUMN_SERVICE_INSTANCE+"=$18,"+
		RESOURCE_COLUMN_RESOURCE_TYPE+"=$19,"+
		RESOURCE_COLUMN_RESOURCE+"=$20,"+
		RESOURCE_COLUMN_IS_CATALOG_PARENT+"=$21,"+
		RESOURCE_COLUMN_CATALOG_PARENT_ID+"=$22,"+
		RESOURCE_COLUMN_RECORD_HASH+"=$23 WHERE "+RESOURCE_COLUMN_RECORD_ID+"=$24",
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		NewNullString(resourceToInsert.SourceCreationTime),
		NewNullString(resourceToInsert.SourceUpdateTime),
		resourceToInsert.CRNFull,
		NewNullString(resourceToInsert.State),
		NewNullString(resourceToInsert.OperationalStatus),
		NewNullString(resourceToInsert.Status),
		NewNullString(resourceToInsert.StatusUpdateTime),
		NewNullString(resourceToInsert.RegulatoryDomain),
		NewNullString(resourceToInsert.CategoryID),
		strconv.FormatBool(resourceToInsert.CategoryParent),
		crnStruct.Version,
		crnStruct.CName,
		crnStruct.CType,
		crnStruct.ServiceName,
		crnStruct.Location,
		crnStruct.Scope,
		crnStruct.ServiceInstance,
		crnStruct.ResourceType,
		crnStruct.Resource,
		strconv.FormatBool(resourceToInsert.IsCatalogParent),
		NewNullString(resourceToInsert.CatalogParentID),
		NewNullString(resourceToInsert.RecordHash),
		recordID)

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return err
}

// UpdateNotification Similar to InsertNotification except with steps to ensure NotificationDescription table is in sync with changes
// After updating Notification table, it will get all existing records from NotificationDescription table with Notification recordId and store it in a map.
// Then as it loops through the itemToInsert.DisplayName array, it will either insert a new NotificationDescription record if it did not exist before
// or it will remove the particular NotificationDescription record from the map if NotificationDescription exists already. After loop exits, it will delete any
// NotificationDescription records that is left in the map.
func UpdateNotification(database *sql.DB, itemToInsert *datastore.NotificationInsert) (error, int) {
	if strings.TrimSpace(itemToInsert.Source) == "" {
		return errors.New("Source is not valid"), http.StatusBadRequest
	}

	if strings.TrimSpace(itemToInsert.SourceID) == "" {
		return errors.New("SourceId is not valid"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", itemToInsert.Source, "SourceID: ", itemToInsert.SourceID, "Crn: ", itemToInsert.CRNFull)

	// check ShortDescription
	if itemToInsert.ShortDescription != nil && checkDisplayNames(itemToInsert.ShortDescription) == false {
		return errors.New("One or more ShortDescription.Name or ShortDescription.Language have an empty string."), http.StatusBadRequest
	}
	// check LongDescription
	if itemToInsert.LongDescription != nil && checkDisplayNames(itemToInsert.LongDescription) == false {
		return errors.New("One or more LongDescription.Name or LongDescription.Language have an empty string."), http.StatusBadRequest
	}

	// store crn in lowercase
	itemToInsert.CRNFull = strings.ToLower(itemToInsert.CRNFull)

	crnStruct, err := crn.ParseAll(itemToInsert.CRNFull)
	if err != nil {
		msg := "itemToInsert.CRNFull has incorrect format."
		log.Println(tlog.Log(), msg)
		return errors.New(msg), http.StatusBadRequest
	}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting update transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				log.Println(tlog.Log(), err)
				return err, http.StatusInternalServerError
			}
		}

		recordID := CreateNotificationRecordID(itemToInsert.Source, itemToInsert.SourceID, itemToInsert.CRNFull, itemToInsert.IncidentID, itemToInsert.Type)

		err = updateNotificationStatement(tx, itemToInsert, crnStruct, recordID)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				log.Println(tlog.Log(), err)
				return err, http.StatusInternalServerError
			}
		}
		// update NotificationDescription table
		ndMap := map[string]datastore.NotificationDescriptionGet{}

		// get NotificationDescription currently in DB for this notification recordID
		ndGetArr, err1 := getNotificationDescriptionByNotificationIDStatement(database, recordID)
		// ignore error and continue
		if err1 == nil && ndGetArr != nil {
			for _, ndGet := range *ndGetArr {
				ndMap[ndGet.RecordID] = ndGet
			}
		}

		if itemToInsert.LongDescription != nil && len(itemToInsert.LongDescription) > 0 {
			for _, longDesc := range itemToInsert.LongDescription {
				ndRecordID := CreateRecordIDFromString(recordID + longDesc.Language)
				ndGet, ok := ndMap[ndRecordID]
				if !ok {
					displayName := datastore.DisplayName{Name: longDesc.Name, Language: longDesc.Language}
					_, err1 := insertNotificationDescriptionStatement(tx, &displayName, recordID)
					if err1 != nil {
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							log.Println(tlog.Log(), err1)
							return err1, http.StatusInternalServerError
						}
					}
				} else if ndGet.LongDescription == longDesc.Name {
					delete(ndMap, ndRecordID) // nothing change, delete from map
				} else {
					err1 := updateNotificationDescriptionStatement(tx, longDesc.Name, ndRecordID)
					if err1 != nil {
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							log.Println(tlog.Log(), err1)
							return err1, http.StatusInternalServerError
						}
					}
					delete(ndMap, ndRecordID) // updated, delete from map
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}
		}

		//Delete any remaining notification description records
		if len(ndMap) > 0 {
			for recID := range ndMap {
				//delete notification description
				err2 := deleteNotificationDescriptionStatement(tx, recID)
				if err2 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err2) {
						restartTransaction = true
						break
					} else {
						log.Println(tlog.Log(), err2)
						return err2, http.StatusInternalServerError
					}
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}
		}

		// commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Println(tlog.Log()+"Error in Commit: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				log.Println(tlog.Log(), err)
				return err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}
	return nil, http.StatusOK
}

func updateNotificationStatement(tx *sql.Tx, itemToInsert *datastore.NotificationInsert, crnStruct crn.MaskAll, recordID string) error {
	log.Println(tlog.Log()+"recordID:", recordID)
	log.Println(tlog.Log()+"crnStruct:", crnStruct)

	// Convert ShortDescription to string array
	shortDescription := []string{}
	for _, desc := range itemToInsert.ShortDescription {
		shortDescription = append(shortDescription, PadRight(desc.Language, " ", NOTIFICATION_LANGUAGE_LENGTH)+desc.Name)
	}

	// Convert ResourceDisplayNames to string array
	resourceDisplayNames := []string{}
	for _, name := range itemToInsert.ResourceDisplayNames {
		resourceDisplayNames = append(resourceDisplayNames, PadRight(name.Language, " ", NOTIFICATION_LANGUAGE_LENGTH)+name.Name)
	}

	// Source, SourceID and crn cannot be changed for they form the primary key
	result, err := tx.Exec("UPDATE "+NOTIFICATION_TABLE_NAME+" SET "+
		NOTIFICATION_COLUMN_PNP_UPDATE_TIME+"=$1,"+
		NOTIFICATION_COLUMN_SOURCE_CREATION_TIME+"=$2,"+
		NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME+"=$3,"+
		NOTIFICATION_COLUMN_TYPE+"=$4,"+
		NOTIFICATION_COLUMN_CATEGORY+"=$5,"+
		NOTIFICATION_COLUMN_EVENT_TIME_START+"=$6,"+
		NOTIFICATION_COLUMN_EVENT_TIME_END+"=$7,"+
		NOTIFICATION_COLUMN_INCIDENT_ID+"=$8,"+
		NOTIFICATION_COLUMN_SHORT_DESCRIPTION+"=$9,"+
		NOTIFICATION_COLUMN_RESOURCE_DISPLAY_NAMES+"=$10,"+
		NOTIFICATION_COLUMN_TAGS+"=$11,"+
		NOTIFICATION_COLUMN_PNP_REMOVED+"=$12,"+
		NOTIFICATION_COLUMN_RELEASE_NOTE_URL+"=$13,"+
		NOTIFICATION_COLUMN_RECORD_RETRACTION_TIME+"=$14 WHERE "+NOTIFICATION_COLUMN_RECORD_ID+"=$15",
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		NewNullString(itemToInsert.SourceCreationTime),
		NewNullString(itemToInsert.SourceUpdateTime),
		NewNullString(itemToInsert.Type),
		NewNullString(itemToInsert.Category),
		NewNullString(itemToInsert.EventTimeStart),
		NewNullString(itemToInsert.EventTimeEnd),
		NewNullString(itemToInsert.IncidentID),
		pq.Array(shortDescription),
		pq.Array(resourceDisplayNames),
		NewNullString(itemToInsert.Tags),
		strconv.FormatBool(itemToInsert.PnPRemoved),
		NewNullString(itemToInsert.ReleaseNoteUrl),
		NewNullString(itemToInsert.RecordRetractionTime),
		recordID)

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return err
}

// updateNotificationDescriptionStatement - can only update longDescription, for recordID is hash of notification_id + language
func updateNotificationDescriptionStatement(tx *sql.Tx, longDescription string, recordID string) error {
	log.Println(tlog.Log()+"recordID:", recordID)

	result, err := tx.Exec("UPDATE "+NOTIFICATION_DESCRIPTION_TABLE_NAME+" SET "+
		NOTIFICATIONDESCRIPTION_COLUMN_LONG_DESCRIPTION+"=$1 WHERE "+NOTIFICATIONDESCRIPTION_COLUMN_RECORD_ID+"=$2",
		longDescription,
		recordID)

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return err
}

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

// GenerateUUID Generated a new Universal Unique Identifier for a record id primary key
func GenerateUUID() string {
	newUUID, err := uuid.NewV4()
	if err != nil {
		log.Printf(tlog.Log()+"ERROR Generating UUID: %s", err)
	}

	return newUUID.String()
}

// InsertResource - return resource_record_id, error, http status code
func InsertResource(database *sql.DB, resourceToInsert *datastore.ResourceInsert) (string, error, int) {
	resourceRecordID := ""

	if strings.TrimSpace(resourceToInsert.Source) == "" {
		return "", errors.New("Source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(resourceToInsert.SourceID) == "" {
		return "", errors.New("SourceID cannot be empty"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", resourceToInsert.Source, "SourceID: ", resourceToInsert.SourceID)

	//if CheckEnumFields(resourceToInsert.State, "ok", "archived") == false {
	//	return "", errors.New("State is not valid"), http.StatusBadRequest
	//}

	//if CheckEnumFields(resourceToInsert.OperationalStatus, "none", "ga", "experiment", "deprecated") == false {
	//	return "", errors.New("OperationalStatus is not valid"), http.StatusBadRequest
	//}

	//if CheckEnumFields(resourceToInsert.Status, "ok", "degraded", "failed", "maintenance") == false {
	//	return "", errors.New("Status is not valid"), http.StatusBadRequest
	//}

	// check DisplayName
	if resourceToInsert.DisplayNames != nil && checkDisplayNames(resourceToInsert.DisplayNames) == false {
		return "", errors.New("One or more DisplayNameInsert.Name or DisplayNameInsert.Language have an empty string"), http.StatusBadRequest
	}

	// check Tag
	if resourceToInsert.Tags != nil && checkTags(resourceToInsert.Tags) == false {
		return "", errors.New("One or more Tag.ID have an empty string"), http.StatusBadRequest
	}
	// store crn in lowercase
	resourceToInsert.CRNFull = strings.ToLower(resourceToInsert.CRNFull)

	// if resourceToInsert.CRNFull is not IBM Public Cloud, then only save the resource as "crn:v1::service-name:::::" for Gaas
	isGaasResource, crnFull, err0 := ConvertToResourceLookupCRN(resourceToInsert.CRNFull)
	if err0 != nil {
		log.Println(tlog.Log()+"ERROR: ConvertToResourceLookupCRN returns error: ", err0)
		return "", err0, http.StatusBadRequest
	} else if isGaasResource {
		resourceToInsert.CRNFull = crnFull
		log.Println(tlog.Log()+"isGaasResource, crnFull: ", resourceToInsert.CRNFull+" is converted to "+crnFull)
	}
	crnStruct, err := crn.ParseAll(resourceToInsert.CRNFull)
	if err != nil {
		return "", errors.New("resourceToInsert.CRNFull has incorrect format"), http.StatusBadRequest
	}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting insert transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return "", err, http.StatusInternalServerError
			}
		}

		resourceRecordID, err = insertResourceStatement(tx, resourceToInsert, crnStruct)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				return resourceRecordID, err, http.StatusInternalServerError
			}
		}

		// insert to Display Name table
		if resourceToInsert.DisplayNames != nil && len(resourceToInsert.DisplayNames) > 0 {
			for i := range resourceToInsert.DisplayNames {
				// To fix scan error G601 (CWE-118): Implicit memory aliasing in for loop.
				_, err1 := insertDisplayNameStatement(tx, &resourceToInsert.DisplayNames[i], resourceRecordID)
				if err1 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
						restartTransaction = true
						break
					} else {
						return resourceRecordID, err1, http.StatusInternalServerError
					}
				}
			}
			if restartTransaction {
				// tx has already rollback
				Delay()
				continue
			}
		}

		// insert to Visibility and Visibility Junction table
		var visibilityRecordID string
		if resourceToInsert.Visibility != nil {
			for _, v := range resourceToInsert.Visibility {
				if strings.TrimSpace(v) == "" {
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
								return resourceRecordID, err1, http.StatusInternalServerError
							}
						}
					} else if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
						restartTransaction = true
						break
					} else {
						return resourceRecordID, err1, http.StatusInternalServerError
					}
				} else {
					visibilityRecordID = visibilityGet.RecordID
				}

				// insert record to VisibilityJunction table
				_, err1 = insertVisibilityJunctionStatement(tx, resourceRecordID, visibilityRecordID)
				if err1 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
						restartTransaction = true
						break
					} else {
						return resourceRecordID, err1, http.StatusInternalServerError
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
		var tagRecordID string
		if resourceToInsert.Tags != nil {
			for _, t := range resourceToInsert.Tags {
				if strings.TrimSpace(t.ID) == "" {
					continue
				}
				tagGet, err1, rc := getTagByRecordIDStatement(database, CreateRecordIDFromString(t.ID))
				if err1 != nil {
					if rc == http.StatusOK {
						// cannot find tag, insert it
						tag := datastore.TagInsert{
							ID: t.ID,
						}
						tagRecordID, err1 = insertTagStatement(tx, &tag)
						if err1 != nil {
							if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
								restartTransaction = true
								break
							} else {
								return resourceRecordID, err1, http.StatusInternalServerError
							}
						}
					} else {
						if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
							restartTransaction = true
							break
						} else {
							return resourceRecordID, err1, http.StatusInternalServerError
						}
					}
				} else {
					tagRecordID = tagGet.RecordID
				}

				// insert record to TagJunction table
				_, err1 = insertTagJunctionStatement(tx, resourceRecordID, tagRecordID)
				if err1 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
						restartTransaction = true
						break
					} else {
						return resourceRecordID, err1, http.StatusInternalServerError
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
				return resourceRecordID, err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}

	}
	return resourceRecordID, nil, http.StatusOK
}

// insertResourceStatement - return recordID, error
func insertResourceStatement(tx *sql.Tx, resourceToInsert *datastore.ResourceInsert, crnStruct crn.MaskAll) (string, error) {
	recordID := CreateRecordIDFromSourceSourceID(resourceToInsert.Source, resourceToInsert.SourceID)

	resourceToInsert.RecordHash = ComputeResourceRecordHash(resourceToInsert)

	result, err := tx.Exec("INSERT INTO "+RESOURCE_TABLE_NAME+"("+
		RESOURCE_COLUMN_RECORD_ID+","+
		RESOURCE_COLUMN_PNP_CREATION_TIME+","+
		RESOURCE_COLUMN_PNP_UPDATE_TIME+","+
		RESOURCE_COLUMN_SOURCE_CREATION_TIME+","+
		RESOURCE_COLUMN_SOURCE_UPDATE_TIME+","+
		RESOURCE_COLUMN_CRN_FULL+","+
		RESOURCE_COLUMN_STATE+","+
		RESOURCE_COLUMN_OPERATIONAL_STATUS+","+
		RESOURCE_COLUMN_SOURCE+","+
		RESOURCE_COLUMN_SOURCE_ID+","+
		RESOURCE_COLUMN_STATUS+","+
		RESOURCE_COLUMN_STATUS_UPDATE_TIME+","+
		RESOURCE_COLUMN_REGULATORY_DOMAIN+","+
		RESOURCE_COLUMN_CATEGORY_ID+","+
		RESOURCE_COLUMN_CATEGORY_PARENT+","+
		RESOURCE_COLUMN_VERSION+","+
		RESOURCE_COLUMN_CNAME+","+
		RESOURCE_COLUMN_CTYPE+","+
		RESOURCE_COLUMN_SERVICE_NAME+","+
		RESOURCE_COLUMN_LOCATION+","+
		RESOURCE_COLUMN_SCOPE+","+
		RESOURCE_COLUMN_SERVICE_INSTANCE+","+
		RESOURCE_COLUMN_RESOURCE_TYPE+","+
		RESOURCE_COLUMN_RESOURCE+","+
		RESOURCE_COLUMN_IS_CATALOG_PARENT+","+
		RESOURCE_COLUMN_CATALOG_PARENT_ID+","+
		RESOURCE_COLUMN_RECORD_HASH+") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)",
		recordID,
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		NewNullString(resourceToInsert.SourceCreationTime),
		NewNullString(resourceToInsert.SourceUpdateTime),
		resourceToInsert.CRNFull,
		NewNullString(resourceToInsert.State),
		NewNullString(resourceToInsert.OperationalStatus),
		resourceToInsert.Source,
		resourceToInsert.SourceID,
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
		NewNullString(resourceToInsert.RecordHash))

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return recordID, err
}

// InsertIncident - return incident_record_id, error, http status code
func InsertIncident(database *sql.DB, itemToInsert *datastore.IncidentInsert) (string, error, int) {
	var incidentRecordID string

	if strings.TrimSpace(itemToInsert.Source) == "" {
		return "", errors.New("Source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(itemToInsert.SourceID) == "" {
		return "", errors.New("SourceID cannot be empty"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", itemToInsert.Source, "SourceID: ", itemToInsert.SourceID)

	if CheckEnumFields(itemToInsert.State, "new", "in-progress", "resolved") == false {
		return "", errors.New("State is not valid"), http.StatusBadRequest
	}

	if CheckEnumFields(itemToInsert.Classification, "confirmed-cie", "potential-cie", "normal") == false {
		return "", errors.New("Classification is not valid"), http.StatusBadRequest
	}

	if itemToInsert.CRNFull == nil || len(itemToInsert.CRNFull) == 0 {
		return "", errors.New("IncidentInsert.CRNFull cannot be nil or empty"), http.StatusBadRequest
	}

	err := VerifyCRNArray(itemToInsert.CRNFull)
	if err != nil {
		return "", err, http.StatusBadRequest
	}

	if itemToInsert.Audience == "" || len(itemToInsert.Audience) == 0 {
		itemToInsert.Audience = SNnill2PnP
	}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println("Error in starting insert transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return "", err, http.StatusInternalServerError
			}
		}

		//Insert into incidents table
		incidentRecordID, err = insertIncidentStatement(tx, itemToInsert)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				return incidentRecordID, err, http.StatusInternalServerError
			}
		}

		//Insert into IncidentJunction table
		//First get the resourceID from crnFull
		for i := 0; i < len(itemToInsert.CRNFull); i++ {

			// change crn filter if it is a Gaas Resource
			queryStr, err, statusCode := CreateCrnFilter(itemToInsert.CRNFull[i], "")
			if err != nil {
				return incidentRecordID, err, statusCode
			}
			resources, err1, rc := GetResourceByQuerySimple(database, queryStr)
			if err1 != nil {
				if rc == http.StatusBadRequest {
					//programming error in caller
					return incidentRecordID, err1, http.StatusNotImplemented
				}
				log.Println(tlog.Log()+"Error getting resource_id: ", err1)
				// return incident_record_id, err, http.StatusInternalServerError
			}
			for i := 0; i < len(*resources); i++ {
				_, err1 = insertIncidentJunctionStatement(tx, (*resources)[i].RecordID, incidentRecordID)
				if err1 != nil {
					log.Println(tlog.Log()+"Error inserting junction: ", err1)
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
						restartTransaction = true
						break
					} else {
						return incidentRecordID, err1, http.StatusNotImplemented
					}

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

		// commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Println(tlog.Log()+"Error in Commit: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return incidentRecordID, err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}

	return incidentRecordID, nil, http.StatusOK

}

// insertIncidentStatement - return recordID, error
func insertIncidentStatement(tx *sql.Tx, itemToInsert *datastore.IncidentInsert) (string, error) {
	recordID := CreateRecordIDFromSourceSourceID(itemToInsert.Source, itemToInsert.SourceID)

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

	// Gabriel Avila 2020-12-08
	// Deal with inserts with a customer description field longer than the DB field (4k)
	if len(itemToInsert.CustomerImpactDescription) >= 4000 {
		itemToInsert.CustomerImpactDescription = itemToInsert.CustomerImpactDescription[:3999]
	}

	// Code reapply from https://github.ibm.com/cloud-sre/pnp-abstraction/pull/694/files
	var re = regexp.MustCompile(`\${SN_RECORD_ID}`)
	itemToInsert.TargetedURL = re.ReplaceAllString(itemToInsert.TargetedURL, itemToInsert.SourceID)

	result, err := tx.Exec("INSERT INTO "+INCIDENT_TABLE_NAME+"("+
		INCIDENT_COLUMN_RECORD_ID+","+
		INCIDENT_COLUMN_PNP_CREATION_TIME+","+
		INCIDENT_COLUMN_PNP_UPDATE_TIME+","+
		INCIDENT_COLUMN_SOURCE_CREATION_TIME+","+
		INCIDENT_COLUMN_SOURCE_UPDATE_TIME+","+
		INCIDENT_COLUMN_START_TIME+","+
		INCIDENT_COLUMN_END_TIME+","+
		INCIDENT_COLUMN_SHORT_DESCRIPTION+","+
		INCIDENT_COLUMN_LONG_DESCRIPTION+","+
		INCIDENT_COLUMN_STATE+","+
		INCIDENT_COLUMN_CLASSIFICATION+","+
		INCIDENT_COLUMN_SEVERITY+","+
		INCIDENT_COLUMN_CRN_FULL+","+
		INCIDENT_COLUMN_SOURCE_ID+","+
		INCIDENT_COLUMN_SOURCE+","+
		INCIDENT_COLUMN_REGULATORY_DOMAIN+","+
		INCIDENT_COLUMN_AFFECTED_ACTIVITY+","+
		INCIDENT_COLUMN_CUSTOMER_IMPACT_DESCRIPTION+","+
		INCIDENT_COLUMN_PNP_REMOVED+","+
		INCIDENT_COLUMN_TARGETED_URL+","+
		INCIDENT_COLUMN_AUDIENCE+
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)",
		recordID,
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
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
		itemToInsert.SourceID,
		itemToInsert.Source,
		NewNullString(itemToInsert.RegulatoryDomain),
		NewNullString(itemToInsert.AffectedActivity),
		NewNullString(itemToInsert.CustomerImpactDescription),
		strconv.FormatBool(itemToInsert.PnPRemoved),
		NewNullString(itemToInsert.TargetedURL),
		NewNullString(itemToInsert.Audience))

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in insertIncidentStatement Rollback: ", err1)
		}
	}
	return recordID, err

}

// InsertMaintenance - return maintenance_record_id, error, http status code
func InsertMaintenance(database *sql.DB, itemToInsert *datastore.MaintenanceInsert) (string, error, int) {
	var maintenanceRecordID string

	if strings.TrimSpace(itemToInsert.Source) == "" {
		return "", errors.New("Source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(itemToInsert.SourceID) == "" {
		return "", errors.New("SourceID cannot be empty"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", itemToInsert.Source, "SourceID: ", itemToInsert.SourceID)

	if CheckEnumFields(strings.ToLower(itemToInsert.State), "new", "scheduled", "in-progress", "complete") == false {
		return "", errors.New("State is not valid"), http.StatusBadRequest
	}

	if itemToInsert.CRNFull == nil || len(itemToInsert.CRNFull) == 0 {
		return "", errors.New("MaintenanceInsert.CRNFull cannot be nil or empty"), http.StatusBadRequest
	}

	err := VerifyCRNArray(itemToInsert.CRNFull)
	if err != nil {
		return "", err, http.StatusBadRequest
	}

	if itemToInsert.Audience == "" || len(itemToInsert.Audience) == 0 {
		itemToInsert.Audience = SNnill2PnP
	}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting insert transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return "", err, http.StatusInternalServerError
			}
		}

		//Insert into incidents table
		maintenanceRecordID, err = insertMaintenanceStatement(tx, itemToInsert)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				return maintenanceRecordID, err, http.StatusInternalServerError
			}
		}

		//Insert into MaintenanceJunction table
		//First get the resourceID from crnFull
		for i := 0; i < len(itemToInsert.CRNFull); i++ {

			// change crn filter if it is a Gaas Resource
			queryStr, err, statusCode := CreateCrnFilter(itemToInsert.CRNFull[i], "")
			if err != nil {
				return maintenanceRecordID, err, statusCode
			}
			resources, err1, rc := GetResourceByQuerySimple(database, queryStr)
			if err1 != nil {
				if rc == http.StatusBadRequest {
					//programming error in caller
					return maintenanceRecordID, err1, http.StatusNotImplemented
				}
				log.Println(tlog.Log()+"Error getting resource_id: ", err1)
				// return maintenanceRecordID, err, http.StatusInternalServerError
			}
			for i := 0; i < len(*resources); i++ {
				_, err1 = insertMaintenanceJunctionStatement(tx, (*resources)[i].RecordID, maintenanceRecordID)
				if err1 != nil {
					log.Println(tlog.Log()+"Error inserting junction: ", err1)
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
						restartTransaction = true
						break
					} else {
						return maintenanceRecordID, err, http.StatusInternalServerError
					}
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

		// commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Println(tlog.Log()+"Error in Commit: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return maintenanceRecordID, err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}

	return maintenanceRecordID, nil, http.StatusOK
}

// insertMaintenanceStatement - return recordID, error
func insertMaintenanceStatement(tx *sql.Tx, itemToInsert *datastore.MaintenanceInsert) (string, error) {
	recordID := CreateRecordIDFromSourceSourceID(itemToInsert.Source, itemToInsert.SourceID)

	itemToInsert.RecordHash = ComputeMaintenanceRecordHash(itemToInsert)

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

	// Code reapply from https://github.ibm.com/cloud-sre/pnp-abstraction/pull/694/files
	var re = regexp.MustCompile(`\${SN_RECORD_ID}`)
	itemToInsert.TargetedURL = re.ReplaceAllString(itemToInsert.TargetedURL, itemToInsert.SourceID)

	result, err := tx.Exec("INSERT INTO "+MAINTENANCE_TABLE_NAME+"("+
		MAINTENANCE_COLUMN_RECORD_ID+","+
		MAINTENANCE_COLUMN_PNP_CREATION_TIME+","+
		MAINTENANCE_COLUMN_PNP_UPDATE_TIME+","+
		MAINTENANCE_COLUMN_SOURCE_CREATION_TIME+","+
		MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME+","+
		MAINTENANCE_COLUMN_START_TIME+","+
		MAINTENANCE_COLUMN_END_TIME+","+
		MAINTENANCE_COLUMN_SHORT_DESCRIPTION+","+
		MAINTENANCE_COLUMN_LONG_DESCRIPTION+","+
		MAINTENANCE_COLUMN_CRN_FULL+","+
		MAINTENANCE_COLUMN_STATE+","+
		MAINTENANCE_COLUMN_DISRUPTIVE+","+
		MAINTENANCE_COLUMN_SOURCE_ID+","+
		MAINTENANCE_COLUMN_SOURCE+","+
		MAINTENANCE_COLUMN_RECORD_HASH+","+
		MAINTENANCE_COLUMN_MAINTENANCE_DURATION+","+
		MAINTENANCE_COLUMN_DISRUPTION_TYPE+","+
		MAINTENANCE_COLUMN_DISRUPTION_DESCRIPTION+","+
		MAINTENANCE_COLUMN_DISRUPTION_DURATION+","+
		MAINTENANCE_COLUMN_REGULATORY_DOMAIN+","+
		MAINTENANCE_COLUMN_PNP_REMOVED+","+
		MAINTENANCE_COLUMN_COMPLETION_CODE+","+
		MAINTENANCE_COLUMN_TARGETED_URL+","+
		MAINTENANCE_COLUMN_AUDIENCE+
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)",
		recordID,
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
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
		itemToInsert.SourceID,
		itemToInsert.Source,
		NewNullString(itemToInsert.RecordHash),
		itemToInsert.MaintenanceDuration,
		NewNullString(itemToInsert.DisruptionType),
		NewNullString(itemToInsert.DisruptionDescription),
		itemToInsert.DisruptionDuration,
		NewNullString(itemToInsert.RegulatoryDomain),
		strconv.FormatBool(itemToInsert.PnPRemoved),
		NewNullString(itemToInsert.CompletionCode),
		NewNullString(itemToInsert.TargetedURL),
		NewNullString(itemToInsert.Audience))

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in insertMaintenanceStatement Rollback: ", err1)
		}
	}
	return recordID, err
}

// InsertSubscriptionStatement creates a new subscription record and relationships - returns recordID, err, statusCode
func InsertSubscriptionStatement(dbConnection *sql.DB, itemToInsert *datastore.SubscriptionInsert) (string, error, int) {
	newUUID := GenerateUUID()
	log.Println(tlog.Log()+"newUUID:", newUUID)

	counter := 0
	var dataToAdd []interface{}
	query := "INSERT INTO " + SUBSCRIPTION_TABLE_NAME + "("
	reflectVal := reflect.Indirect(reflect.ValueOf(itemToInsert))
	numFields := reflectVal.NumField()
	skipped := 0
	for i := 0; i < numFields; i++ {
		if reflectVal.Field(i).Interface() != nil {

			fieldValue := reflectVal.Field(i).Interface().(string)
			if fieldValue == "" {
				skipped++
				if i >= numFields-skipped-1 {
					query += SUBSCRIPTION_COLUMN_RECORD_ID
					break
				}
			} else {
				dataToAdd = append(dataToAdd, fieldValue)
				counter++
				if i >= numFields-skipped-1 {
					query += reflectVal.Type().Field(i).Tag.Get("col") + "," + SUBSCRIPTION_COLUMN_RECORD_ID

				} else {
					query += reflectVal.Type().Field(i).Tag.Get("col") + ","

				}
			}

		}

	}

	dataToAdd = append(dataToAdd, newUUID)
	query += ") VALUES ("
	for i := 1; i <= counter; i++ {
		if i == counter {
			query += "$" + strconv.Itoa(i) + "," + "$" + strconv.Itoa(i+1) + ")"
		} else {
			query += "$" + strconv.Itoa(i) + ","
		}
	}
	log.Println(tlog.Log()+"DEBUG Query : ", query, dataToAdd)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		result, err := dbConnection.Exec(query, dataToAdd...)

		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+"Error: ", err, result)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return newUUID, err, http.StatusInternalServerError
			}

		} else {
			retry = false
			if result != nil {
				lastInsertID, err1 := result.LastInsertId()
				if err1 != nil {
					log.Println(tlog.Log()+"Error getting last insert id: ", err1)
				}
				rowsAffected, err2 := result.RowsAffected()
				if err2 != nil {
					log.Println(tlog.Log()+"Error getting rows affected: ", err2)
				}
				log.Println(tlog.Log()+"Result: ", lastInsertID, rowsAffected)
			}
		}
	}

	return newUUID, nil, http.StatusOK
}

/*
func InsertWatchStatement(database *sql.DB, itemToInsert *datastore.WatchInsert) (string, error) {

	FCT := "InsertWatchStatement: "

	newUUID := GenerateUUID()

	result, err := database.Exec("INSERT INTO "+WATCH_TABLE_NAME+"("+
		WATCH_COLUMN_RECORD_ID+","+
		WATCH_COLUMN_SUBSCRIPTION_ID+","+
		WATCH_COLUMN_KIND+","+
		WATCH_COLUMN_PATH+","+
		WATCH_COLUMN_CRN_FULL+","+
		WATCH_COLUMN_WILDCARDS+","+
		WATCH_COLUMN_RECORD_ID_TO_WATCH+") VALUES ($1, $2, $3, $4, $5, $6, $7)",
		newUUID,
		itemToInsert.SubscriptionRecordID,
		itemToInsert.Kind,
		itemToInsert.Path,
		itemToInsert.CRNFull,
		strconv.FormatBool(itemToInsert.Wildcards),
		itemToInsert.RecordIDToWatch)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(FCT+"Error: ", err, result)
	}
	return newUUID, err
}
*/

// InsertWatchByTypeForSubscriptionStatement  Insert a Watch record by type for Subscription
func InsertWatchByTypeForSubscriptionStatement(database *sql.DB, itemToInsert *datastore.WatchInsert, apiURL string) (*datastore.WatchReturn, error, int) {
	start := time.Now()
	newUUID := GenerateUUID()
	log.Println(tlog.Log()+"newUUID:", newUUID)

	counter := 0
	var dataToAdd []interface{}
	query := "INSERT INTO " + WATCH_TABLE_NAME + "("
	retQuery := " RETURNING record_id, subscription_id, kind, path, wildcards, record_id_to_watch, crn_full, subscription_email "

	// store crn's in database in lowercase
	for i := range itemToInsert.CRNFull {
		itemToInsert.CRNFull[i] = strings.ToLower(itemToInsert.CRNFull[i])
	}

	reflectVal := reflect.Indirect(reflect.ValueOf(itemToInsert))
	numFields := reflectVal.NumField()
	for i := 0; i < numFields; i++ {
		if reflectVal.Field(i).Interface() != nil {

			switch v := reflectVal.Field(i).Interface().(type) {
			case int:
				dataToAdd = append(dataToAdd, strconv.Itoa(v))
			case bool:
				dataToAdd = append(dataToAdd, strconv.FormatBool(v))
			case string:
				if v == "" {
					continue
				}
				dataToAdd = append(dataToAdd, v)

			default:
				if reflectVal.Type().Field(i).Name == "RecordIDToWatch" {
					records := "{"
					for i, v := range itemToInsert.RecordIDToWatch {
						records += v
						if i != len(itemToInsert.RecordIDToWatch)-1 {
							records += ","
						}
					}
					records += "}"
					dataToAdd = append(dataToAdd, records)
				}
				if reflectVal.Type().Field(i).Name == "CRNFull" {
					crns := "{"
					for i, v := range itemToInsert.CRNFull {
						crns += v
						if i != len(itemToInsert.CRNFull)-1 {
							crns += ","
						}
					}
					crns += "}"
					dataToAdd = append(dataToAdd, crns)
				}
			}

			counter++
			query += reflectVal.Type().Field(i).Tag.Get("col") + ","
		}
	}

	query += WATCH_COLUMN_RECORD_ID

	dataToAdd = append(dataToAdd, newUUID)
	query += ") VALUES ("
	for i := 1; i <= counter; i++ {
		if i == counter {
			query += "$" + strconv.Itoa(i) + "," + "$" + strconv.Itoa(i+1) + ")"
		} else {
			query += "$" + strconv.Itoa(i) + ","
		}
	}
	query += retQuery
	log.Println(tlog.Log()+"DEBUG: Query:", query)
	var watch datastore.WatchGet
	var watchReturn datastore.WatchReturn

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		// restartTransaction := false   https://github.ibm.com/cloud-sre/pnp-abstraction/issues/646

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting insert transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}

		err = tx.QueryRow(query, dataToAdd...).Scan(&watch.RecordID, &watch.SubscriptionRecordID, &watch.Kind, &watch.Path, &watch.Wildcards, &watch.RecordIDToWatch, &watch.CRNFull, &watch.SubscriptionEmail)
		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+"Error:  :", err.Error())
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}

		watchReturn = datastore.ConvertWatchGetToWatchReturn(watch, apiURL)

		// Insert into WatchJunction table
		// First get the resourceID from crnFull
		for i := 0; i < len(itemToInsert.CRNFull); i++ {

			// change crn filter if it is a Gaas Resource
			queryStr, err, statusCode := CreateCrnFilter(itemToInsert.CRNFull[i], itemToInsert.Wildcards)
			if err != nil {
				return &watchReturn, err, statusCode
			}
			resources, err1, rc := GetResourceByQuerySimple(database, queryStr)

			if err1 != nil || resources == nil {
				if rc == http.StatusBadRequest {
					//programming error in caller
					return &watchReturn, err1, http.StatusNotImplemented
				}
				log.Println(tlog.Log()+"Error getting resource for that id: ", err1)
				//return &watchReturn, err1, http.StatusInternalServerError

			}

			// Commented out https://github.ibm.com/cloud-sre/pnp-abstraction/issues/646
			// for i := 0; i < len(*resources); i++ {
			// 	_, err1 = insertWatchJunctionStatement(tx, (*resources)[i].RecordID, watch.RecordID)
			// 	if err1 != nil {
			// 		log.Println(FCT+"Error inserting junction: ", err1)
			// 		if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
			// 			restartTransaction = true
			// 			break
			// 		} else {
			// 			return &watchReturn, err1, http.StatusInternalServerError
			// 		}

			// 	}
			// }
			// if restartTransaction {
			// 	break
			// }
		}
		// https://github.ibm.com/cloud-sre/pnp-abstraction/issues/646
		// if restartTransaction {
		// 	// tx has already rollback
		// 	Delay()
		// 	continue
		// }

		// commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Println(tlog.Log()+"Error in Commit: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return &watchReturn, err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}

	fmt.Println(tlog.Log(), time.Since(start))
	return &watchReturn, nil, http.StatusOK
}

// InsertCase create a new record in case_table
func InsertCase(database *sql.DB, itemToInsert *datastore.CaseInsert) (string, error, int) {
	if strings.TrimSpace(itemToInsert.Source) == "" {
		return "", errors.New("Source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(itemToInsert.SourceID) == "" {
		return "", errors.New("SourceID cannot be empty"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", itemToInsert.Source, "SourceID: ", itemToInsert.SourceID)

	recordID := CreateRecordIDFromSourceSourceID(itemToInsert.Source, itemToInsert.SourceID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		result, err := database.Exec("INSERT INTO "+CASE_TABLE_NAME+"("+
			CASE_COLUMN_RECORD_ID+","+
			CASE_COLUMN_SOURCE+","+
			CASE_COLUMN_SOURCE_ID+","+
			CASE_COLUMN_SOURCE_SYS_ID+") VALUES ($1, $2, $3, $4)",
			recordID,
			itemToInsert.Source,
			itemToInsert.SourceID,
			itemToInsert.SourceSysID)

		if err != nil {
			// failed to execute SQL statements. Exit
			log.Println(tlog.Log()+"Error: ", err, result)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return recordID, err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}
	return recordID, nil, http.StatusOK
}

// insertVisibilityStatement - return recordID, err
func insertVisibilityStatement(tx *sql.Tx, itemToInsert *datastore.VisibilityInsert) (string, error) {
	recordID := CreateRecordIDFromString(itemToInsert.Name)
	log.Println(tlog.Log()+"recordID:", recordID)

	result, err := tx.Exec("INSERT INTO "+VISIBILITY_TABLE_NAME+"("+
		VISIBILITY_COLUMN_RECORD_ID+","+
		VISIBILITY_COLUMN_NAME+","+
		VISIBILITY_COLUMN_DESCRIPTION+") VALUES ($1, $2, $3)",
		recordID,
		itemToInsert.Name,
		itemToInsert.Description)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println("Error: ", err1)
		}
	}
	return recordID, err
}

// insertTagStatement - return recordID, err
func insertTagStatement(tx *sql.Tx, itemToInsert *datastore.TagInsert) (string, error) {
	recordID := CreateRecordIDFromString(itemToInsert.ID)
	log.Println(tlog.Log()+"recordID:", recordID)

	result, err := tx.Exec("INSERT INTO "+TAG_TABLE_NAME+"("+
		TAG_COLUMN_RECORD_ID+","+
		TAG_COLUMN_ID+") VALUES ($1, $2)",
		recordID,
		itemToInsert.ID)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error: ", err1)
		}
	}
	return recordID, err
}

func insertDisplayNameStatement(tx *sql.Tx, itemToInsert *datastore.DisplayName, resourceRecordID string) (string, error) {
	newUUID := GenerateUUID()
	log.Println(tlog.Log()+"newUUID:", newUUID)

	result, err := tx.Exec("INSERT INTO "+DISPLAY_NAMES_TABLE_NAME+"("+
		DISPLAYNAMES_COLUMN_RECORD_ID+","+
		DISPLAYNAMES_COLUMN_NAME+","+
		DISPLAYNAMES_COLUMN_LANGUAGE+","+
		DISPLAYNAMES_COLUMN_RESOURCE_ID+") VALUES ($1, $2, $3, $4)",
		newUUID,
		itemToInsert.Name,
		itemToInsert.Language,
		resourceRecordID)

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return newUUID, err
}

func insertVisibilityJunctionStatement(tx *sql.Tx, resourceID string, visibilityID string) (string, error) {
	newUUID := GenerateUUID()
	log.Println(tlog.Log()+"newUUID:", newUUID)

	result, err := tx.Exec("INSERT INTO "+VISIBILITY_JUNCTION_TABLE_NAME+"("+
		VISIBILITYJUNCTION_COLUMN_RECORD_ID+","+
		VISIBILITYJUNCTION_COLUMN_RESOURCE_ID+","+
		VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID+") VALUES ($1, $2, $3)",
		newUUID,
		resourceID,
		visibilityID)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return newUUID, err
}

func insertTagJunctionStatement(tx *sql.Tx, resourceID string, tagID string) (string, error) {
	newUUID := GenerateUUID()
	log.Println(tlog.Log()+"newUUID:", newUUID)

	result, err := tx.Exec("INSERT INTO "+TAG_JUNCTION_TABLE_NAME+"("+
		TAGJUNCTION_COLUMN_RECORD_ID+","+
		TAGJUNCTION_COLUMN_RESOURCE_ID+","+
		TAGJUNCTION_COLUMN_TAG_ID+") VALUES ($1, $2, $3)",
		newUUID,
		resourceID,
		tagID)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return newUUID, err
}

func insertIncidentJunctionStatement(tx *sql.Tx, resourceID string, incidentID string) (string, error) {
	newUUID := GenerateUUID()
	log.Println(tlog.Log()+"newUUID:", newUUID)

	result, err := tx.Exec("INSERT INTO "+INCIDENT_JUNCTION_TABLE_NAME+"("+
		INCIDENTJUNCTION_COLUMN_RECORD_ID+","+
		INCIDENTJUNCTION_COLUMN_RESOURCE_ID+","+
		INCIDENTJUNCTION_COLUMN_INCIDENT_ID+") VALUES ($1, $2, $3)",
		newUUID,
		resourceID,
		incidentID)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return newUUID, err
}

func insertMaintenanceJunctionStatement(tx *sql.Tx, resourceID string, maintenanceID string) (string, error) {
	newUUID := GenerateUUID()
	log.Println(tlog.Log()+"newUUID:", newUUID)

	result, err := tx.Exec("INSERT INTO "+MAINTENANCE_JUNCTION_TABLE_NAME+"("+
		MAINTENANCEJUNCTION_COLUMN_RECORD_ID+","+
		MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID+","+
		MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID+") VALUES ($1, $2, $3)",
		newUUID,
		resourceID,
		maintenanceID)

	if err != nil {
		// failed to execute SQL statements. Exit
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return newUUID, err
}

//Commented out https://github.ibm.com/cloud-sre/pnp-abstraction/issues/646
// func insertWatchJunctionStatement(tx *sql.Tx, resourceID string, watchID string) (string, error) {

// 	FCT := "insertWatchJunctionStatement: "

// 	newUUID := GenerateUUID()
// 	log.Println(FCT+"newUUID:", newUUID)

// 	result, err := tx.Exec("INSERT INTO "+WATCH_JUNCTION_TABLE_NAME+"("+
// 		WATCHJUNCTION_COLUMN_RECORD_ID+","+
// 		WATCHJUNCTION_COLUMN_RESOURCE_ID+","+
// 		WATCHJUNCTION_COLUMN_WATCH_ID+") VALUES ($1, $2, $3)",
// 		newUUID,
// 		resourceID,
// 		watchID)

// 	if err != nil {
// 		// failed to execute SQL statements. Exit
// 		log.Println(FCT+"Error: ", err, result)
// 		err1 := tx.Rollback()
// 		if err1 != nil {
// 			log.Println(FCT+"Error in Rollback: ", err1)
// 		}
// 	}
// 	return newUUID, err
// }

func insertNotificationDescriptionStatement(tx *sql.Tx, itemToInsert *datastore.DisplayName, notificationRecordID string) (string, error) {
	recordID := CreateRecordIDFromString(notificationRecordID + itemToInsert.Language)
	log.Println(tlog.Log()+"recordID:", recordID)

	result, err := tx.Exec("INSERT INTO "+NOTIFICATION_DESCRIPTION_TABLE_NAME+"("+
		NOTIFICATIONDESCRIPTION_COLUMN_RECORD_ID+","+
		NOTIFICATIONDESCRIPTION_COLUMN_LONG_DESCRIPTION+","+
		NOTIFICATIONDESCRIPTION_COLUMN_LANGUAGE+","+
		NOTIFICATIONDESCRIPTION_COLUMN_NOTIFICATION_ID+") VALUES ($1, $2, $3, $4)",
		recordID,
		itemToInsert.Name,
		itemToInsert.Language,
		notificationRecordID)

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return recordID, err
}

// insertNotificationStatement - return recordID, error
func insertNotificationStatement(tx *sql.Tx, itemToInsert *datastore.NotificationInsert, crnStruct crn.MaskAll) (string, error) {
	recordID := CreateNotificationRecordID(itemToInsert.Source, itemToInsert.SourceID, itemToInsert.CRNFull, itemToInsert.IncidentID, itemToInsert.Type)
	log.Println(tlog.Log()+"recordID:", recordID)

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

	result, err := tx.Exec("INSERT INTO "+NOTIFICATION_TABLE_NAME+"("+
		NOTIFICATION_COLUMN_RECORD_ID+","+
		NOTIFICATION_COLUMN_PNP_CREATION_TIME+","+
		NOTIFICATION_COLUMN_PNP_UPDATE_TIME+","+
		NOTIFICATION_COLUMN_SOURCE_CREATION_TIME+","+
		NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME+","+
		NOTIFICATION_COLUMN_EVENT_TIME_START+","+
		NOTIFICATION_COLUMN_EVENT_TIME_END+","+
		NOTIFICATION_COLUMN_SOURCE+","+
		NOTIFICATION_COLUMN_SOURCE_ID+","+
		NOTIFICATION_COLUMN_TYPE+","+
		NOTIFICATION_COLUMN_CATEGORY+","+
		NOTIFICATION_COLUMN_INCIDENT_ID+","+
		NOTIFICATION_COLUMN_SHORT_DESCRIPTION+","+
		NOTIFICATION_COLUMN_RESOURCE_DISPLAY_NAMES+","+
		NOTIFICATION_COLUMN_CRN_FULL+","+
		NOTIFICATION_COLUMN_VERSION+","+
		NOTIFICATION_COLUMN_CNAME+","+
		NOTIFICATION_COLUMN_CTYPE+","+
		NOTIFICATION_COLUMN_SERVICE_NAME+","+
		NOTIFICATION_COLUMN_LOCATION+","+
		NOTIFICATION_COLUMN_SCOPE+","+
		NOTIFICATION_COLUMN_SERVICE_INSTANCE+","+
		NOTIFICATION_COLUMN_RESOURCE_TYPE+","+
		NOTIFICATION_COLUMN_RESOURCE+","+
		NOTIFICATION_COLUMN_TAGS+","+
		NOTIFICATION_COLUMN_PNP_REMOVED+","+
		NOTIFICATION_COLUMN_RELEASE_NOTE_URL+","+
		NOTIFICATION_COLUMN_RECORD_RETRACTION_TIME+") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27,$28)",
		recordID,
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		NewNullString(itemToInsert.SourceCreationTime),
		NewNullString(itemToInsert.SourceUpdateTime),
		NewNullString(itemToInsert.EventTimeStart),
		NewNullString(itemToInsert.EventTimeEnd),
		itemToInsert.Source,
		itemToInsert.SourceID,
		itemToInsert.Type,
		NewNullString(itemToInsert.Category),
		NewNullString(itemToInsert.IncidentID),
		pq.Array(shortDescription),
		pq.Array(resourceDisplayNames),
		itemToInsert.CRNFull,
		crnStruct.Version,
		crnStruct.CName,
		crnStruct.CType,
		crnStruct.ServiceName,
		crnStruct.Location,
		crnStruct.Scope,
		crnStruct.ServiceInstance,
		crnStruct.ResourceType,
		crnStruct.Resource,
		NewNullString(itemToInsert.Tags),
		strconv.FormatBool(itemToInsert.PnPRemoved),
		NewNullString(itemToInsert.ReleaseNoteUrl),
		NewNullString(itemToInsert.RecordRetractionTime))

	if err != nil {
		// failed to execute SQL statements. Rollback
		log.Println(tlog.Log()+"Error: ", err, result)
		err1 := tx.Rollback()
		if err1 != nil {
			log.Println(tlog.Log()+"Error in Rollback: ", err1)
		}
	}
	return recordID, err
}

// InsertNotification - return record_id, error, http status code
func InsertNotification(database *sql.DB, itemToInsert *datastore.NotificationInsert) (string, error, int) {
	recordID := ""

	if strings.TrimSpace(itemToInsert.Source) == "" {
		return "", errors.New("Source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(itemToInsert.SourceID) == "" {
		return "", errors.New("SourceID cannot be empty"), http.StatusBadRequest
	}

	log.Println(tlog.Log()+"Source: ", itemToInsert.Source, "SourceID: ", itemToInsert.SourceID, "Crn: ", itemToInsert.CRNFull)

	if strings.TrimSpace(itemToInsert.Type) == "" {
		return "", errors.New("Type cannot be empty"), http.StatusBadRequest
	}

	// check ShortDescription
	if itemToInsert.ShortDescription != nil && checkDisplayNames(itemToInsert.ShortDescription) == false {
		return "", errors.New("One or more ShortDescription.Name or ShortDescription.Language have an empty string"), http.StatusBadRequest
	}

	// check LongDescription
	if itemToInsert.LongDescription != nil && checkDisplayNames(itemToInsert.LongDescription) == false {
		return "", errors.New("One or more LongDescription.Name or LongDescription.Language have an empty string"), http.StatusBadRequest
	}

	if len(itemToInsert.CRNFull) == 0 {
		return "", errors.New("CRNFull cannot be nil or empty"), http.StatusBadRequest
	}

	// store crn in lowercase
	itemToInsert.CRNFull = strings.ToLower(itemToInsert.CRNFull)

	crnStruct, err := crn.ParseAll(itemToInsert.CRNFull)
	if err != nil {
		return "", errors.New("itemToInsert.CRNFull has incorrect format"), http.StatusBadRequest
	}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		restartTransaction := false

		tx, err := database.Begin()
		if err != nil {
			log.Println(tlog.Log()+"Error in starting insert transaction: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return "", err, http.StatusInternalServerError
			}
		}

		recordID, err = insertNotificationStatement(tx, itemToInsert, crnStruct)
		if err != nil {
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// tx has already rollback
				Delay()
				continue
			} else {
				return recordID, err, http.StatusInternalServerError
			}
		}

		// insert to Notification Description table
		if itemToInsert.LongDescription != nil && len(itemToInsert.LongDescription) > 0 && recordID != "" {
			for i := range itemToInsert.LongDescription {
				// to fix go scan error G601 (CWE-118): Implicit memory aliasing in for loop
				_, err1 := insertNotificationDescriptionStatement(tx, &itemToInsert.LongDescription[i], recordID)
				if err1 != nil {
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err1) {
						restartTransaction = true
						break
					} else {
						return recordID, err1, http.StatusInternalServerError
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
				return recordID, err, http.StatusInternalServerError
			}
		} else {
			retry = false
		}
	}

	return recordID, nil, http.StatusOK
}

// CheckEnumFields Returns true if possibleValues contains value
func CheckEnumFields(value string, possibleValues ...string) bool {
	for _, val := range possibleValues {
		if value == val {
			return true
		}
	}
	return false
}

func checkDisplayNames(displayNames []datastore.DisplayName) bool {
	for _, displayName := range displayNames {
		if displayName.Name == "" || strings.TrimSpace(displayName.Language) == "" {
			return false
		}
	}
	return true
}

func checkTags(tags []datastore.Tag) bool {
	for _, tag := range tags {
		if strings.TrimSpace(tag.ID) == "" {
			return false
		}
	}
	return true
}

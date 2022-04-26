// New updates base on requirement https://github.ibm.com/cloud-sre/pnp-status/issues/340
// A new filed notification_table.pnp_removed was added to flag record in the process to be removed
// Implying to modify these test cases due to pnp_removed will be passsed as defaul search value with
// a value of false
// Alejandro Torres Rojas Jun-2019

package db

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/oss-globals/tlog"
	datastore "github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

const (
	// SELECT_CASE select fields statement for case table
	SELECT_CASE = "SELECT " +
		CASE_COLUMN_RECORD_ID + "," +
		CASE_COLUMN_SOURCE + "," +
		CASE_COLUMN_SOURCE_ID + "," +
		CASE_COLUMN_SOURCE_SYS_ID

	// SELECT_CASE_FROM from and where part of the select case statement
	SELECT_CASE_FROM = " FROM " + CASE_TABLE_NAME + " WHERE "

	SELECT_CASE_WITHOUT_COUNT = SELECT_CASE + SELECT_CASE_FROM

	// Select Resource
	RESOURCE_TABLE_ALIAS  = "r"
	RESOURCE_TABLE_PREFIX = "r."

	SELECT_RESOURCE_COLUMNS = RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_PNP_CREATION_TIME + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_PNP_UPDATE_TIME + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_SOURCE_CREATION_TIME + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_SOURCE_UPDATE_TIME + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_STATE + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_OPERATIONAL_STATUS + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_SOURCE + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_SOURCE_ID + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_STATUS + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_STATUS_UPDATE_TIME + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_REGULATORY_DOMAIN + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_CATEGORY_ID + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_CATEGORY_PARENT + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_CRN_FULL + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_IS_CATALOG_PARENT + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_CATALOG_PARENT_ID + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_HASH

	SELECT_RESOURCE_HASHES = RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_SOURCE + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_SOURCE_ID + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_STATUS + "," +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_HASH

	SELECT_RESOURCE_FROM = " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS + " WHERE "

	SELECT_RESOURCE_BY_QUERY = "SELECT " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + SELECT_RESOURCE_FROM

	SELECT_DISTINCT_RESOURCE_BY_QUERY = "SELECT DISTINCT " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID

	SELECT_RESOURCE_BY_RECORDID_STATEMENT = "SELECT DISTINCT " + SELECT_RESOURCE_COLUMNS + "," + SELECT_VISIBILITY_COLUMNS + "," + SELECT_TAG_COLUMNS + "," + SELECT_DISPLAYNAMES_COLUMNS +
		" FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS +
		" full join " + VISIBILITY_JUNCTION_TABLE_NAME + " " + JUNCTION_TABLE_ALIAS + " on " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "=" + JUNCTION_TABLE_PREFIX + VISIBILITYJUNCTION_COLUMN_RESOURCE_ID +
		" full join " + VISIBILITY_TABLE_NAME + " " + VISIBILITY_TABLE_ALIAS + " on " + JUNCTION_TABLE_PREFIX + VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID + "=" + VISIBILITY_TABLE_PREFIX + VISIBILITY_COLUMN_RECORD_ID +
		" full join " + TAG_JUNCTION_TABLE_NAME + " " + TAGJUNCTION_TABLE_ALIAS + " on " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "=" + TAGJUNCTION_TABLE_PREFIX + TAGJUNCTION_COLUMN_RESOURCE_ID +
		" full join " + TAG_TABLE_NAME + " " + TAG_TABLE_ALIAS + " on " + TAGJUNCTION_TABLE_PREFIX + TAGJUNCTION_COLUMN_TAG_ID + "=" + TAG_TABLE_PREFIX + TAG_COLUMN_RECORD_ID +
		" full join " + DISPLAY_NAMES_TABLE_NAME + " " + DISPLAYNAMES_TABLE_ALIAS + " on " + DISPLAYNAMES_TABLE_PREFIX + DISPLAYNAMES_COLUMN_RESOURCE_ID + "=" + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID +
		" WHERE " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "="

	// Select Incident
	INCIDENT_TABLE_ALIAS  = "i"
	INCIDENT_TABLE_PREFIX = "i."

	SELECT_INCIDENT_COLUMNS = INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_RECORD_ID + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_PNP_CREATION_TIME + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_PNP_UPDATE_TIME + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_SOURCE_CREATION_TIME + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_SOURCE_UPDATE_TIME + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_START_TIME + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_END_TIME + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_SHORT_DESCRIPTION + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_LONG_DESCRIPTION + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_STATE + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_CLASSIFICATION + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_SEVERITY + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_SOURCE_ID + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_SOURCE + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_REGULATORY_DOMAIN + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_CRN_FULL + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_AFFECTED_ACTIVITY + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_CUSTOMER_IMPACT_DESCRIPTION + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_PNP_REMOVED + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_TARGETED_URL + "," +
		INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_AUDIENCE

	JUNCTION_TABLE_ALIAS  = "j"
	JUNCTION_TABLE_PREFIX = "j."

	// Select Maintenance
	MAINTENANCE_TABLE_ALIAS  = "m"
	MAINTENANCE_TABLE_PREFIX = "m."

	SELECT_MAINTENANCE_COLUMNS = MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_RECORD_ID + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_PNP_CREATION_TIME + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_PNP_UPDATE_TIME + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_SOURCE_CREATION_TIME + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_START_TIME + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_END_TIME + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_SHORT_DESCRIPTION + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_LONG_DESCRIPTION + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_STATE + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_DISRUPTIVE + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_SOURCE_ID + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_SOURCE + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_REGULATORY_DOMAIN + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_RECORD_HASH + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_MAINTENANCE_DURATION + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_DISRUPTION_TYPE + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_DISRUPTION_DESCRIPTION + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_DISRUPTION_DURATION + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_COMPLETION_CODE + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_CRN_FULL + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_PNP_REMOVED + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_TARGETED_URL + "," +
		MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_AUDIENCE

	// Select Watch
	WATCH_TABLE_ALIAS  = "w"
	WATCH_TABLE_PREFIX = "w."

	SELECT_WATCH_COLUMNS = WATCH_TABLE_PREFIX + WATCH_COLUMN_RECORD_ID + "," +
		WATCH_TABLE_PREFIX + WATCH_COLUMN_SUBSCRIPTION_ID + "," +
		WATCH_TABLE_PREFIX + WATCH_COLUMN_KIND + "," +
		WATCH_TABLE_PREFIX + WATCH_COLUMN_PATH + "," +
		WATCH_TABLE_PREFIX + WATCH_COLUMN_CRN_FULL + "," +
		WATCH_TABLE_PREFIX + WATCH_COLUMN_WILDCARDS + "," +
		WATCH_TABLE_PREFIX + WATCH_COLUMN_RECORD_ID_TO_WATCH + "," +
		WATCH_TABLE_PREFIX + WATCH_COLUMN_SUBSCRIPTION_EMAIL

	// Visibility table
	VISIBILITY_TABLE_ALIAS  = "v"
	VISIBILITY_TABLE_PREFIX = "v."

	SELECT_VISIBILITY_COLUMNS = VISIBILITY_TABLE_PREFIX + VISIBILITY_COLUMN_NAME

	// Tag table
	TAG_TABLE_ALIAS  = "t"
	TAG_TABLE_PREFIX = "t."

	SELECT_TAG_COLUMNS = TAG_TABLE_PREFIX + TAG_COLUMN_ID

	// Tag Junction table
	TAGJUNCTION_TABLE_ALIAS  = "tj"
	TAGJUNCTION_TABLE_PREFIX = "tj."

	// DisplayNames table
	DISPLAYNAMES_TABLE_ALIAS  = "d"
	DISPLAYNAMES_TABLE_PREFIX = "d."

	SELECT_DISPLAYNAMES_COLUMNS = DISPLAYNAMES_TABLE_PREFIX + DISPLAYNAMES_COLUMN_NAME + "," + DISPLAYNAMES_TABLE_PREFIX + DISPLAYNAMES_COLUMN_LANGUAGE

	// Select Notification
	NOTIFICATION_TABLE_ALIAS  = "n"
	NOTIFICATION_TABLE_PREFIX = "n."

	SELECT_NOTIFICATION_COLUMNS = NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_RECORD_ID + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_PNP_CREATION_TIME + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_PNP_UPDATE_TIME + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_SOURCE_CREATION_TIME + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_EVENT_TIME_START + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_EVENT_TIME_END + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_SOURCE + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_SOURCE_ID + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_TYPE + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_CATEGORY + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_INCIDENT_ID + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_CRN_FULL + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_RESOURCE_DISPLAY_NAMES + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_SHORT_DESCRIPTION + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_TAGS + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_RECORD_RETRACTION_TIME + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_PNP_REMOVED + "," +
		NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_RELEASE_NOTE_URL

	// Select NotificationDescription
	NOTIFICATIONDESCRIPTION_TABLE_ALIAS  = "nd"
	NOTIFICATIONDESCRIPTION_TABLE_PREFIX = "nd."

	SELECT_NOTIFICATIONDESCRIPTION_COLUMNS = NOTIFICATIONDESCRIPTION_TABLE_PREFIX + NOTIFICATIONDESCRIPTION_COLUMN_LONG_DESCRIPTION + "," + NOTIFICATIONDESCRIPTION_TABLE_PREFIX + NOTIFICATIONDESCRIPTION_COLUMN_LANGUAGE

	SELECT_NOTIFICATION_BY_RECORDID_STATEMENT = "SELECT DISTINCT " + SELECT_NOTIFICATION_COLUMNS + "," + SELECT_NOTIFICATIONDESCRIPTION_COLUMNS +
		" FROM " + NOTIFICATION_TABLE_NAME + " " + NOTIFICATION_TABLE_ALIAS +
		" full join " + NOTIFICATION_DESCRIPTION_TABLE_NAME + " " + NOTIFICATIONDESCRIPTION_TABLE_ALIAS + " on " + NOTIFICATIONDESCRIPTION_TABLE_PREFIX + NOTIFICATIONDESCRIPTION_COLUMN_NOTIFICATION_ID + "=" + NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_RECORD_ID +
		" WHERE " + NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_RECORD_ID + "="
)

// GetCaseBySourceID gets case_table recors by passed isource id
func GetCaseBySourceID(database *sql.DB, source string, sourceID string) (*datastore.CaseReturn, error, int) {

	if strings.TrimSpace(source) == "" {
		return nil, errors.New("source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(sourceID) == "" {
		return nil, errors.New("sourceID cannot be empty"), http.StatusBadRequest
	}

	recordID := CreateRecordIDFromSourceSourceID(source, sourceID)

	return GetCaseByRecordID(database, recordID)

}

// GetCaseByRecordID gets case_table recors by passed record id
func GetCaseByRecordID(database *sql.DB, recordID string) (*datastore.CaseReturn, error, int) {
	var caseToReturn datastore.CaseReturn

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow(SELECT_CASE_WITHOUT_COUNT+
			CASE_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&caseToReturn.RecordID,
			&caseToReturn.Source,
			&caseToReturn.SourceID,
			&caseToReturn.SourceSysID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Case with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		default:
			//log.Printf("Case is %s\n", caseToReturn.RecordID)
			retry = false
		}
	}
	return &caseToReturn, nil, http.StatusOK
}

// GetResourceBySourceID - Gets a resource_table record by resource_table.source_id and resource_table.source
func GetResourceBySourceID(database *sql.DB, source string, sourceID string) (*datastore.ResourceReturn, error, int) {

	if strings.TrimSpace(source) == "" {
		return nil, errors.New("source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(sourceID) == "" {
		return nil, errors.New("sourceID cannot be empty"), http.StatusBadRequest
	}

	resourceRecordID := CreateRecordIDFromSourceSourceID(source, sourceID)

	return GetResourceByRecordID(database, resourceRecordID)
}

// GetResourceByRecordID gets a resource record by quering the passed resource_table.record_id
func GetResourceByRecordID(database *sql.DB, resourceRecordID string) (*datastore.ResourceReturn, error, int) {
	log.Println(tlog.Log()+"DEBUG: resourceRecordID:", resourceRecordID)

	resourceToReturn := datastore.ResourceReturn{}

	selectStmt := SELECT_RESOURCE_BY_RECORDID_STATEMENT + "'" + resourceRecordID + "'"

	//log.Println(tlog.Log()+"selectStmt: ", selectStmt)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(selectStmt)
		rowsReturned := 0

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Resource found.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		default:
			defer rows.Close()

			resourceToReturn = datastore.ResourceReturn{}
			for rows.Next() {
				rowsReturned++
				r := datastore.ResourceGet{}
				var visibilityName, tagID, displayNameName, displayNameLanguage sql.NullString

				err = rows.Scan(&r.RecordID, &r.PnpCreationTime, &r.PnpUpdateTime, &r.SourceCreationTime, &r.SourceUpdateTime,
					&r.State, &r.OperationalStatus, &r.Source, &r.SourceID, &r.Status, &r.StatusUpdateTime, &r.RegulatoryDomain,
					&r.CategoryID, &r.CategoryParent, &r.CRNFull, &r.IsCatalogParent, &r.CatalogParentID, &r.RecordHash,
					&visibilityName, &tagID, &displayNameName, &displayNameLanguage)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
					return nil, err, http.StatusInternalServerError
				}

				if resourceToReturn.RecordID == "" { // only do it once
					resourceToReturn = ConvertResourceGetToResourceReturn(r)
				}

				if visibilityName.Valid && !Contains(resourceToReturn.Visibility, visibilityName.String) {
					resourceToReturn.Visibility = append(resourceToReturn.Visibility, visibilityName.String)
				}

				if tagID.Valid && !ContainTag(resourceToReturn.Tags, tagID.String) {
					tag := datastore.Tag{ID: tagID.String}
					resourceToReturn.Tags = append(resourceToReturn.Tags, tag)
				}

				if displayNameName.Valid && displayNameLanguage.Valid && !ContainDisplayName(resourceToReturn.DisplayNames, displayNameName.String, displayNameLanguage.String) {
					dn := datastore.DisplayName{Name: displayNameName.String, Language: displayNameLanguage.String}
					resourceToReturn.DisplayNames = append(resourceToReturn.DisplayNames, dn)
				} else if displayNameName.Valid && !displayNameLanguage.Valid && !ContainDisplayName(resourceToReturn.DisplayNames, displayNameName.String, "") {
					dn := datastore.DisplayName{Name: displayNameName.String, Language: ""}
					resourceToReturn.DisplayNames = append(resourceToReturn.DisplayNames, dn)
				} else if !displayNameName.Valid && displayNameLanguage.Valid && !ContainDisplayName(resourceToReturn.DisplayNames, "", displayNameLanguage.String) {
					dn := datastore.DisplayName{Name: "", Language: displayNameLanguage.String}
					resourceToReturn.DisplayNames = append(resourceToReturn.DisplayNames, dn)
				}
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err, http.StatusInternalServerError
				}
			}
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}
			if rowsReturned == 0 {
				return nil, sql.ErrNoRows, http.StatusOK
			}
			retry = false
		}
	}

	//log.Println(tlog.Log() + "resourceToReturn: ", resourceToReturn)
	return &resourceToReturn, nil, http.StatusOK
}

// GetResourceByQuery query is a string that has a format like "x=1&y=2&y=3", where the key like x,y,z, etc. must be one of the following values:
// crn, version, cname, ctype, service_name, location, scope, service_instance, resource_type, resource, visibility.
// When key is crn, it must a crn string format, and it can only support single value, cannot be comma separated values, for example,
// "crn=crn:version:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"
// If a crn attribute is missing, wildcard is applied in matching.
//
// For other keys, their values can be comma separated. For example, "cname=abc,def&visibility=tag1,tag2"
//
// The rows returned are sorted by record_id.
// If the query string cannot be parsed by net/url/ParseQuery function, error will be returned.
// limit must be > 0, offset must be >= 0. If limit is <=0, will be no limit
//
// See some examples in GetWatchesByQuery function
func GetResourceByQuery(database *sql.DB, query string, limit int, offset int) (*[]datastore.ResourceReturn, int, error, int) {
	log.Println(tlog.Log()+"query:", query, "limit:", limit, "offset:", offset)

	queryParms, err := url.ParseQuery(query)
	if err != nil {
		//log.Println(tlog.Log()+"Error: ", err)
		return nil, 0, err, http.StatusBadRequest
	}

	// checking query
	if len(queryParms[RESOURCE_QUERY_VERSION]) == 0 && len(queryParms[RESOURCE_QUERY_CNAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CTYPE]) == 0 && len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_LOCATION]) == 0 && len(queryParms[RESOURCE_QUERY_SCOPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) == 0 && len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_RESOURCE]) == 0 && len(queryParms[RESOURCE_QUERY_VISIBILITY]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CRN]) == 0 && len(queryParms[RESOURCE_QUERY_CATALOG_PARENT_ID]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CREATION_TIME_START]) == 0 && len(queryParms[RESOURCE_QUERY_CREATION_TIME_END]) == 0 &&
		len(queryParms[RESOURCE_QUERY_UPDATE_TIME_START]) == 0 && len(queryParms[RESOURCE_QUERY_UPDATE_TIME_END]) == 0 &&
		len(queryParms[RESOURCE_QUERY_PNP_CREATION_TIME_START]) == 0 && len(queryParms[RESOURCE_QUERY_PNP_CREATION_TIME_END]) == 0 &&
		len(queryParms[RESOURCE_QUERY_PNP_UPDATE_TIME_START]) == 0 && len(queryParms[RESOURCE_QUERY_PNP_UPDATE_TIME_END]) == 0 {

		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters. : query:[" + query + "]"
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	hasQuery := false
	hasCRNQuery := false
	hasVisibilityQuery := false
	hasOtherQuery := false

	if len(queryParms[RESOURCE_QUERY_VERSION]) > 0 || len(queryParms[RESOURCE_QUERY_CNAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_CTYPE]) > 0 || len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_LOCATION]) > 0 || len(queryParms[RESOURCE_QUERY_SCOPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) > 0 || len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_RESOURCE]) > 0 {

		if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
			msg := tlog.Log() + "ERROR: You cannot have crn string and also crn attribute in the query. : query:[" + query + "]"
			//log.Println(msg)
			return nil, 0, errors.New(msg), http.StatusBadRequest
		}
		hasCRNQuery = true
	} else if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
		hasCRNQuery = true
	}

	if len(queryParms[RESOURCE_QUERY_CREATION_TIME_START]) > 0 || len(queryParms[RESOURCE_QUERY_CREATION_TIME_END]) > 0 ||
		len(queryParms[RESOURCE_QUERY_PNP_UPDATE_TIME_START]) > 0 || len(queryParms[RESOURCE_QUERY_PNP_UPDATE_TIME_END]) > 0 ||
		len(queryParms[RESOURCE_QUERY_PNP_CREATION_TIME_START]) > 0 || len(queryParms[RESOURCE_QUERY_CREATION_TIME_END]) > 0 ||
		len(queryParms[RESOURCE_QUERY_UPDATE_TIME_END]) > 0 || len(queryParms[RESOURCE_QUERY_UPDATE_TIME_START]) > 0 {
		hasOtherQuery = true
	}

	// ======== Form where clause ========
	whereClause, orderBy, err := CreateCRNSearchWhereClauseFromQuery(queryParms, RESOURCE_TABLE_PREFIX)

	// if there is query on visibility, we have to do further searching
	if strings.TrimSpace(whereClause) == "" {
		hasCRNQuery = false
		hasQuery = false
	} else {
		hasCRNQuery = true
		hasQuery = true
	}

	// visibility
	visibilityQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_VISIBILITY, false)

	if len(visibilityQueryList) == 0 {
		hasVisibilityQuery = false
	} else {
		hasVisibilityQuery = true

		// =========== other query where clause ===========
		// where source_creation_time start
		whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, visibilityQueryList, VISIBILITY_COLUMN_NAME, VISIBILITY_TABLE_PREFIX)
	}

	if hasOtherQuery {
		// creation_time_start
		createStartQueryList, createStartErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_CREATION_TIME_START)
		if createStartErr != nil {
			//log.Println(tlog.Log() + createStartErr.Error())
			return nil, 0, createStartErr, http.StatusBadRequest
		}

		// creation_time_end
		createEndQueryList, createEndErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_CREATION_TIME_END)
		if createEndErr != nil {
			//log.Println(tlog.Log() + createEndErr.Error())
			return nil, 0, createEndErr, http.StatusBadRequest
		}

		// pnp_creation_time_start
		pnpCreateStartQueryList, pnpCreateStartErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_PNP_CREATION_TIME_START)
		if pnpCreateStartErr != nil {
			//log.Println(tlog.Log() + pnpCreateStartErr.Error())
			return nil, 0, pnpCreateStartErr, http.StatusBadRequest
		}

		// pnp_creation_time_end
		pnpCreateEndQueryList, pnpCreateEndErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_PNP_CREATION_TIME_END)
		if pnpCreateEndErr != nil {
			//log.Println(tlog.Log() + pnpCreateEndErr.Error())
			return nil, 0, pnpCreateEndErr, http.StatusBadRequest
		}

		// pnp_update_time_start
		pnpUpdateStartQueryList, pnpUpdateStartErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_PNP_UPDATE_TIME_START)
		if pnpUpdateStartErr != nil {
			//log.Println(tlog.Log() + pnpUpdateStartErr.Error())
			return nil, 0, pnpUpdateStartErr, http.StatusBadRequest
		}

		// pnp_update_time_end
		pnpUpdateEndQueryList, pnpUpdateEndErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_PNP_UPDATE_TIME_END)
		if pnpUpdateEndErr != nil {
			//log.Println(tlog.Log() + pnpUpdateEndErr.Error())
			return nil, 0, pnpUpdateEndErr, http.StatusBadRequest
		}

		// update_time_start
		updateStartQueryList, updateStartErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_UPDATE_TIME_START)
		if updateStartErr != nil {
			//log.Println(tlog.Log() + updateStartErr.Error())
			return nil, 0, updateStartErr, http.StatusBadRequest
		}

		// update_time_end
		updateEndQueryList, updateEndErr := ParseQueryTimeCondition(queryParms, RESOURCE_QUERY_UPDATE_TIME_END)
		if updateEndErr != nil {
			//log.Println(tlog.Log() + updateEndErr.Error())
			return nil, 0, updateEndErr, http.StatusBadRequest
		}

		// =========== other query where clause ===========
		// where source_creation_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, createStartQueryList, RESOURCE_COLUMN_SOURCE_CREATION_TIME, RESOURCE_TABLE_PREFIX, ">=")

		// where source_creation_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, createEndQueryList, RESOURCE_COLUMN_SOURCE_CREATION_TIME, RESOURCE_TABLE_PREFIX, "<=")

		// where pnp_create_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpCreateStartQueryList, RESOURCE_COLUMN_PNP_CREATION_TIME, RESOURCE_TABLE_PREFIX, ">=")

		// where pnp_create_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpCreateEndQueryList, RESOURCE_COLUMN_PNP_CREATION_TIME, RESOURCE_TABLE_PREFIX, "<=")

		// where pnp_update_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpUpdateStartQueryList, RESOURCE_COLUMN_PNP_UPDATE_TIME, RESOURCE_TABLE_PREFIX, ">=")

		// where pnp_update_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpUpdateEndQueryList, RESOURCE_COLUMN_PNP_UPDATE_TIME, RESOURCE_TABLE_PREFIX, "<=")

		// where source_update_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, updateStartQueryList, RESOURCE_COLUMN_SOURCE_UPDATE_TIME, RESOURCE_TABLE_PREFIX, ">=")

		// where source_update_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, updateEndQueryList, RESOURCE_COLUMN_SOURCE_UPDATE_TIME, RESOURCE_TABLE_PREFIX, "<=")

	}

	selectStmt := ""

	// ======== ORDER BY =========
	orderBy = RESOURCE_COLUMN_RECORD_ID // order by RecordID
	orderByClause := ""
	if orderBy != "" {
		orderByClause = " ORDER BY " + orderBy
	}

	// ======== LIMIT =========
	limitClause := ""
	if limit > 0 {
		limitClause = " LIMIT " + strconv.Itoa(limit)
	}

	// ======== OFFSET =========
	offsetClause := ""
	if offset > 0 {
		offsetClause = " OFFSET " + strconv.Itoa(offset)
	}

	// ========= database Query =========
	//log.Println("whereClause: ", whereClause)
	//log.Println("orderByClause: ", orderByClause)
	//log.Println("limitClause: ", limitClause)
	//log.Println("offsetClause: ", offsetClause)

	if hasQuery {
		if hasCRNQuery && !hasVisibilityQuery {

			selectStmt = SELECT_RESOURCE_BY_QUERY + whereClause

		} else if hasVisibilityQuery || hasOtherQuery {

			selectStmt = SELECT_DISTINCT_RESOURCE_BY_QUERY + " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS +
				" inner join " + VISIBILITY_JUNCTION_TABLE_NAME + " " + JUNCTION_TABLE_ALIAS + " on " +
				RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "=" + JUNCTION_TABLE_PREFIX + VISIBILITYJUNCTION_COLUMN_RESOURCE_ID +
				" inner join " + VISIBILITY_TABLE_NAME + " " + VISIBILITY_TABLE_ALIAS + " on " +
				JUNCTION_TABLE_PREFIX + VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID + "=" + VISIBILITY_TABLE_PREFIX + VISIBILITY_COLUMN_RECORD_ID +
				" WHERE " + whereClause
		}
	} else {
		// select all
		//selectStmt = "SELECT " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " FROM " + RESOURCE_TABLE_NAME + " " +
		//	RESOURCE_TABLE_ALIAS
		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters."
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	//log.Println(FCT+"selectStmt: ", selectStmt)

	// The execute statment will be something like this:
	//
	// WITH cte AS (SELECT DISTINCT r.record_id FROM resource_table1 r
	//              inner join visibility_junction_table1 j on r.record_id=j.resource_id
	//              inner join visibility_table1 v on j.visibility_id=v.record_id
	//              WHERE  (r.cname='cname2') AND (r.ctype='ctype2') AND (r.service_name='service-name2')
	//                     AND (r.version='1') AND (v.name='vis2' OR v.name='vis3')
	//             )
	// SELECT DISTINCT * FROM
	//                        (SELECT DISTINCT r.record_id,r.pnp_creation_time,r.pnp_update_time,
	//                         r.source_creation_time,r.source_update_time,r.state,r.operational_status,
	//                         r.source,r.source_id,r.status,r.status_update_time,r.regulatory_domain,
	//                         r.category_id,r.category_parent,r.crn_full,v2.name,t2.id,d2.name,d2.language
	//                         FROM resource_table1 r2
	//                              full join visibility_junction_table1 j2 on r.record_id=j2.resource_id
	//                              full join visibility_table1 v2 on j2.visibility_id=v2.record_id
	//                              full join tag_junction_table1 tj on r.record_id=tj.resource_id
	//                              full join tag_table1 t2 on tj.tag_id=t2.record_id
	//                              full join display_names_table1 d2 on d2.resource_id=r.record_id
	//                         WHERE r.record_id = ANY
	//                                                  (SELECT * FROM cte ORDER BY record_id OFFSET 1 LIMIT 2)
	//                        ) AS r2
	//                        RIGHT JOIN (SELECT count(*) FROM cte) c(total_count) ON true
	//                        ORDER BY record_id;

	executeStmt := "WITH cte AS (" + selectStmt + ") " +
		"SELECT DISTINCT * FROM ( " + SELECT_RESOURCE_BY_RECORDID_STATEMENT + " ANY " +
		"(SELECT * FROM cte " + orderByClause + limitClause + offsetClause +
		") ) AS r2 RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true ORDER BY record_id;"

	log.Println(tlog.Log()+"DEBUG: executeStmt: ", executeStmt)
	resourcesReturn := []datastore.ResourceReturn{}
	totalCount := 0

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(executeStmt)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Resource found.")
			return nil, 0, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, 0, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			//			totalCount, resourcesToReturn, err, status := getMultipleResourcesToReturnFromRows(rows)
			resourcesReturn = []datastore.ResourceReturn{}
			resourceToReturn := datastore.ResourceReturn{}
			rg := datastore.ResourceGet{}
			totalCount = 0
			rowsReturned := 0
			currentRecordID := ""
			var visibilityName, tagID, displayNameName, displayNameLanguage sql.NullString

			for rows.Next() {
				rowsReturned++
				r := datastore.ResourceGetNull{}

				err := rows.Scan(&r.RecordID, &r.PnpCreationTime, &r.PnpUpdateTime, &r.SourceCreationTime, &r.SourceUpdateTime,
					&r.State, &r.OperationalStatus, &r.Source, &r.SourceID, &r.Status, &r.StatusUpdateTime, &r.RegulatoryDomain,
					&r.CategoryID, &r.CategoryParent, &r.CRNFull, &r.IsCatalogParent, &r.CatalogParentID, &r.RecordHash, &visibilityName, &tagID, &displayNameName, &displayNameLanguage, &totalCount)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to ResourceGetNull Error: %v", err)
					return nil, totalCount, err, http.StatusInternalServerError
				}

				if r.RecordID.Valid && currentRecordID != r.RecordID.String {
					if currentRecordID != "" {
						resourcesReturn = append(resourcesReturn, resourceToReturn)
						resourceToReturn = datastore.ResourceReturn{}
					}
					rg = datastore.ResourceGet{}
					err = rows.Scan(&rg.RecordID, &rg.PnpCreationTime, &rg.PnpUpdateTime, &rg.SourceCreationTime, &rg.SourceUpdateTime,
						&rg.State, &rg.OperationalStatus, &rg.Source, &rg.SourceID, &rg.Status, &rg.StatusUpdateTime, &rg.RegulatoryDomain,
						&rg.CategoryID, &rg.CategoryParent, &rg.CRNFull, &rg.IsCatalogParent, &rg.CatalogParentID, &r.RecordHash, &visibilityName, &tagID, &displayNameName, &displayNameLanguage, &totalCount)

					if err != nil {
						log.Printf(tlog.Log()+"Row Scan to ResourceGet Error: %v", err)
						return nil, totalCount, err, http.StatusInternalServerError
					}

					resourceToReturn = ConvertResourceGetToResourceReturn(rg)
					currentRecordID = resourceToReturn.RecordID

				} else if !r.RecordID.Valid && totalCount > 0 && currentRecordID != "" {
					resourcesReturn = append(resourcesReturn, resourceToReturn)
					return &resourcesReturn, totalCount, nil, http.StatusOK

				} else if !r.RecordID.Valid {
					return &resourcesReturn, totalCount, nil, http.StatusOK
				}

				if visibilityName.Valid && !Contains(resourceToReturn.Visibility, visibilityName.String) {
					resourceToReturn.Visibility = append(resourceToReturn.Visibility, visibilityName.String)
				}

				if tagID.Valid && !ContainTag(resourceToReturn.Tags, tagID.String) {
					tag := datastore.Tag{ID: tagID.String}
					resourceToReturn.Tags = append(resourceToReturn.Tags, tag)
				}

				if displayNameName.Valid && displayNameLanguage.Valid && !ContainDisplayName(resourceToReturn.DisplayNames, displayNameName.String, displayNameLanguage.String) {
					dn := datastore.DisplayName{Name: displayNameName.String, Language: displayNameLanguage.String}
					resourceToReturn.DisplayNames = append(resourceToReturn.DisplayNames, dn)
				} else if displayNameName.Valid && !displayNameLanguage.Valid && !ContainDisplayName(resourceToReturn.DisplayNames, displayNameName.String, "") {
					dn := datastore.DisplayName{Name: displayNameName.String, Language: ""}
					resourceToReturn.DisplayNames = append(resourceToReturn.DisplayNames, dn)
				} else if !displayNameName.Valid && displayNameLanguage.Valid && !ContainDisplayName(resourceToReturn.DisplayNames, "", displayNameLanguage.String) {
					dn := datastore.DisplayName{Name: "", Language: displayNameLanguage.String}
					resourceToReturn.DisplayNames = append(resourceToReturn.DisplayNames, dn)
				}
			}
			var status int
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return &resourcesReturn, 0, err, http.StatusInternalServerError
				}
			}
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}
			if rowsReturned == 0 {
				//				return nil, totalCount, sql.ErrNoRows, http.StatusOK
				err = sql.ErrNoRows
				status = http.StatusOK
			} else if currentRecordID == resourceToReturn.RecordID {
				resourcesReturn = append(resourcesReturn, resourceToReturn)
				status = http.StatusOK
			}

			if err != nil && status != http.StatusOK {
				log.Printf(tlog.Log()+"Row Scan to recordId, total_count Error: %v", err)
				return nil, 0, err, status
			}
			retry = false
		}
	}
	return &resourcesReturn, totalCount, nil, http.StatusOK
}

// GetResourceRecordIDsByVisibility - From a list of resource RecordIDs, find all the resources that has the specified visibility, and return those recordIDs
func GetResourceRecordIDsByVisibility(database *sql.DB, resourceRecordIds []string, visibility string) ([]string, error, int) {
	log.Println(tlog.Log()+"resourceRecordIds:", resourceRecordIds, "visibility:", visibility)

	resourcesToReturn := []string{}

	if len(resourceRecordIds) == 0 {
		msg := tlog.Log() + "resourceRecordIds is empty"
		//log.Println(msg)
		return resourcesToReturn, errors.New(msg), http.StatusBadRequest
	}
	if strings.TrimSpace(visibility) == "" {
		msg := tlog.Log() + "visibility is empty"
		//log.Println(msg)
		return resourcesToReturn, errors.New(msg), http.StatusBadRequest
	}

	// create where clause
	whereClause := "(" + VISIBILITY_TABLE_PREFIX + VISIBILITY_COLUMN_NAME + "='" + visibility + "') AND (" + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + " IN ("

	for i, sourceID := range resourceRecordIds {
		if i > 0 {
			whereClause += ","
		}
		whereClause += "'" + strings.TrimSpace(sourceID) + "'"
	}
	whereClause += "))"

	// Select statement
	selectStmt := "SELECT DISTINCT " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID +
		" FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS +
		" inner join " + VISIBILITY_JUNCTION_TABLE_NAME + " " + JUNCTION_TABLE_ALIAS + " on " +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "=" + JUNCTION_TABLE_PREFIX + VISIBILITYJUNCTION_COLUMN_RESOURCE_ID +
		" inner join " + VISIBILITY_TABLE_NAME + " " + VISIBILITY_TABLE_ALIAS + " on " +
		JUNCTION_TABLE_PREFIX + VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID + "=" + VISIBILITY_TABLE_PREFIX + VISIBILITY_COLUMN_RECORD_ID +
		" WHERE " + whereClause + " ORDER BY " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID

	//log.Println("selectStmt: ", selectStmt)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(selectStmt)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Resources found.")
			return resourcesToReturn, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return resourcesToReturn, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			resourcesToReturn = []string{}
			for rows.Next() {
				var recordID string
				err = rows.Scan(&recordID)

				if err != nil {
					log.Println(tlog.Log()+"Row Scan Error: ", err)
					return resourcesToReturn, err, http.StatusInternalServerError
				}

				resourcesToReturn = append(resourcesToReturn, recordID)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err, http.StatusInternalServerError
				}
			}
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}
			retry = false
		}
	}
	return resourcesToReturn, nil, http.StatusOK
}

// GetAllResources get all resources by SELECT r.record_id,r.source,r.source_id,r.status,r.record_hash FROM resource_table r
// return ResourceReturn only record_id,source,source_id,status and record_hash fields
func GetAllResources(database *sql.DB) (*[]datastore.ResourceReturn, error) {
	log.Println(tlog.Log())

	resourcesToReturn := []datastore.ResourceReturn{}
	strQry := "SELECT " + SELECT_RESOURCE_HASHES + " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS
	// rowCnt := GetRecordsCount(database, RESOURCE_TABLE_NAME, "")
	// log.Println(tlog.Log()+"select string: ", strQry, "rows: ", rowCnt)
	//The above lines were created for debugging use, I forget to remove them

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(strQry)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Resource found.")
			return nil, err
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err
			}

		default:
			defer rows.Close()

			resourcesToReturn = []datastore.ResourceReturn{}
			for rows.Next() {
				var r datastore.ResourceGet
				err = rows.Scan(&r.RecordID, &r.Source, &r.SourceID, &r.Status, &r.RecordHash)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
				}
				resReturn := ConvertResourceGetToResourceReturn(r)
				resourcesToReturn = append(resourcesToReturn, resReturn)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err
				}
			}
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}
			retry = false
		}
	}
	log.Println(tlog.Log(), resourcesToReturn)
	return &resourcesToReturn, nil
}

// GetWatchByRecordIDStatement get a watch record from the database using the record_id passed
func GetWatchByRecordIDStatement(database *sql.DB, recordID string, url string) (*datastore.WatchReturn, error, int) {
	log.Println(tlog.Log()+"recordID:", recordID, "url:", url)

	var watch datastore.WatchGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow("SELECT "+
			WATCH_COLUMN_RECORD_ID+","+
			WATCH_COLUMN_SUBSCRIPTION_ID+","+
			WATCH_COLUMN_KIND+","+
			WATCH_COLUMN_PATH+","+
			WATCH_COLUMN_CRN_FULL+","+
			WATCH_COLUMN_WILDCARDS+","+
			WATCH_COLUMN_RECORD_ID_TO_WATCH+","+
			WATCH_COLUMN_SUBSCRIPTION_EMAIL+" FROM "+
			WATCH_TABLE_NAME+" WHERE "+
			WATCH_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&watch.RecordID,
			&watch.SubscriptionRecordID,
			&watch.Kind,
			&watch.Path,
			&watch.CRNFull,
			&watch.Wildcards,
			&watch.RecordIDToWatch,
			&watch.SubscriptionEmail)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Watch with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			log.Printf("Watch is %s\n", watch.RecordID)
			watchToReturn := datastore.ConvertWatchGetToWatchReturn(watch, url)
			return &watchToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusOK

}

// GetWatchByRecordAndSubscription gets a watch record from the database using the passed subscription-id and record_id to match it
func GetWatchByRecordAndSubscription(database *sql.DB, subscriptionID string, recordID string, url string) (*datastore.WatchReturn, error, int) {
	log.Println(tlog.Log()+"subscriptionID:", subscriptionID, "recordID:", recordID, "url:", url)

	var watch datastore.WatchGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow("SELECT "+
			WATCH_COLUMN_RECORD_ID+","+
			WATCH_COLUMN_SUBSCRIPTION_ID+","+
			WATCH_COLUMN_KIND+","+
			WATCH_COLUMN_PATH+","+
			WATCH_COLUMN_CRN_FULL+","+
			WATCH_COLUMN_WILDCARDS+","+
			WATCH_COLUMN_RECORD_ID_TO_WATCH+","+
			WATCH_COLUMN_SUBSCRIPTION_EMAIL+" FROM "+
			WATCH_TABLE_NAME+" WHERE "+
			WATCH_COLUMN_RECORD_ID+" = $1 AND "+
			WATCH_COLUMN_SUBSCRIPTION_ID+" = $2",
			recordID, subscriptionID).Scan(&watch.RecordID,
			&watch.SubscriptionRecordID,
			&watch.Kind,
			&watch.Path,
			&watch.CRNFull,
			&watch.Wildcards,
			&watch.RecordIDToWatch,
			&watch.SubscriptionEmail)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Watch with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			log.Printf("Watch is %s\n", watch.RecordID)
			watchReturn := datastore.ConvertWatchGetToWatchReturn(watch, url)
			return &watchReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusOK

}

// GetWatchesByRecordID gets all the watches associated with a subscription
func GetWatchesByKind(database *sql.DB, subscriptionID string, kind string, url string) ([]datastore.WatchReturn, error, int) {
	log.Println(tlog.Log()+"subscriptionID:", subscriptionID, "kind:", kind, "url:", url)

	var watchesToReturn []datastore.WatchReturn
	var watch datastore.WatchGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			WATCH_COLUMN_RECORD_ID+","+
			WATCH_COLUMN_SUBSCRIPTION_ID+","+
			WATCH_COLUMN_KIND+","+
			WATCH_COLUMN_PATH+","+
			WATCH_COLUMN_CRN_FULL+","+
			WATCH_COLUMN_WILDCARDS+","+
			WATCH_COLUMN_RECORD_ID_TO_WATCH+","+
			WATCH_COLUMN_SUBSCRIPTION_EMAIL+" FROM "+
			WATCH_TABLE_NAME+" WHERE "+
			WATCH_COLUMN_SUBSCRIPTION_ID+" = $1 AND "+
			WATCH_COLUMN_KIND+"= $2", subscriptionID, kind)

		if err != nil {
			log.Println(tlog.Log()+": rowcount error: ", err.Error())
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		}
		defer rows.Close()

		watchesToReturn = []datastore.WatchReturn{}
		for rows.Next() {
			watch = datastore.WatchGet{}
			err = rows.Scan(&watch.RecordID,
				&watch.SubscriptionRecordID,
				&watch.Kind,
				&watch.Path,
				&watch.CRNFull,
				&watch.Wildcards,
				&watch.RecordIDToWatch,
				&watch.SubscriptionEmail)
			if err != nil {
				log.Println(tlog.Log()+": Error in GetSubscriptionAlls: ", err)
				return nil, err, http.StatusInternalServerError
			}
			watchReturn := datastore.ConvertWatchGetToWatchReturn(watch, url)

			watchesToReturn = append(watchesToReturn, watchReturn)

		}
		if err = rows.Err(); err != nil {
			log.Printf(tlog.Log()+"rows.Err(): ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				if err = rows.Close(); err != nil {
					log.Println(tlog.Log()+"Error in rows.Close ", err)
				}
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}
		if err = rows.Close(); err != nil {
			log.Println(tlog.Log()+"Error in rows.Close ", err)
		}
		//log.Printf(tlog.Log()+": Watches are %s\n", watchesToReturn)
		retry = false
	}
	return watchesToReturn, nil, http.StatusOK

}

func GetWatchesBySubscriptionIdStatement(database *sql.DB, subscriptionID string, apiURL string) ([]datastore.WatchReturn, error, int) {
	log.Println(tlog.Log()+"subscriptionID:", subscriptionID, "apiUrl:", apiURL)

	var watchesToReturn []datastore.WatchReturn
	var watch datastore.WatchGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			WATCH_COLUMN_RECORD_ID+","+
			WATCH_COLUMN_SUBSCRIPTION_ID+","+
			WATCH_COLUMN_KIND+","+
			WATCH_COLUMN_PATH+","+
			WATCH_COLUMN_CRN_FULL+","+
			WATCH_COLUMN_WILDCARDS+","+
			WATCH_COLUMN_RECORD_ID_TO_WATCH+","+
			WATCH_COLUMN_SUBSCRIPTION_EMAIL+" FROM "+
			WATCH_TABLE_NAME+" WHERE "+
			WATCH_COLUMN_SUBSCRIPTION_ID+" = $1", subscriptionID)

		if err != nil {
			log.Println(tlog.Log()+": rowcount error: ", err.Error())
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		}
		defer rows.Close()

		watchesToReturn = []datastore.WatchReturn{}
		for rows.Next() {
			watch = datastore.WatchGet{}
			err = rows.Scan(&watch.RecordID,
				&watch.SubscriptionRecordID,
				&watch.Kind,
				&watch.Path,
				&watch.CRNFull,
				&watch.Wildcards,
				&watch.RecordIDToWatch,
				&watch.SubscriptionEmail)
			if err != nil {
				log.Println(tlog.Log()+": Error: ", err)
				return nil, err, http.StatusInternalServerError
			}
			watchReturn := datastore.ConvertWatchGetToWatchReturn(watch, apiURL)
			watchesToReturn = append(watchesToReturn, watchReturn)

		}
		if err = rows.Err(); err != nil {
			log.Printf(tlog.Log()+"rows.Err(): ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				if err = rows.Close(); err != nil {
					log.Println(tlog.Log()+"Error in rows.Close ", err)
				}
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}
		if err = rows.Close(); err != nil {
			log.Println(tlog.Log()+"Error in rows.Close ", err)
		}
		retry = false
		//log.Printf(tlog.Log()+": Watches are %s\n", watchesToReturn)
	}
	return watchesToReturn, nil, http.StatusOK

}

// GetWatchesAll gets all the watches associated with a subscription
func GetWatchesAll(database *sql.DB, apiURL string) ([]datastore.WatchReturn, error, int) {
	log.Println(tlog.Log()+"apiUrl:", apiURL)

	var watchesToReturn []datastore.WatchReturn
	var watch datastore.WatchGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT " +
			WATCH_COLUMN_RECORD_ID + "," +
			WATCH_COLUMN_SUBSCRIPTION_ID + "," +
			WATCH_COLUMN_KIND + "," +
			WATCH_COLUMN_PATH + "," +
			WATCH_COLUMN_CRN_FULL + "," +
			WATCH_COLUMN_WILDCARDS + "," +
			WATCH_COLUMN_RECORD_ID_TO_WATCH + "," +
			WATCH_COLUMN_SUBSCRIPTION_EMAIL + " FROM " +
			WATCH_TABLE_NAME)

		if err != nil {
			log.Println(tlog.Log()+": rowcount error: ", err.Error())
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		}
		defer rows.Close()

		watchesToReturn = []datastore.WatchReturn{}
		for rows.Next() {
			watch = datastore.WatchGet{}
			err = rows.Scan(&watch.RecordID,
				&watch.SubscriptionRecordID,
				&watch.Kind,
				&watch.Path,
				&watch.CRNFull,
				&watch.Wildcards,
				&watch.RecordIDToWatch,
				&watch.SubscriptionEmail)
			if err != nil {
				log.Println(tlog.Log()+": Error in GetWatchesAll: ", err)
				return nil, err, http.StatusInternalServerError
			}
			watchReturn := datastore.ConvertWatchGetToWatchReturn(watch, apiURL)
			watchesToReturn = append(watchesToReturn, watchReturn)

		}
		if err = rows.Err(); err != nil {
			log.Printf(tlog.Log()+"rows.Err(): ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// explicitly do rows close
				if err = rows.Close(); err != nil {
					log.Println(tlog.Log()+"Error in rows.Close ", err)
				}
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}
		// explicitly do rows close
		if err = rows.Close(); err != nil {
			log.Println(tlog.Log()+"Error in rows.Close ", err)
		}
		retry = false
		//log.Printf(tlog.Log()+": Watches are %s\n", watchesToReturn)
	}
	return watchesToReturn, nil, http.StatusOK

}
func GetWatchesByRecordIDToWatch(database *sql.DB, recordID string, apiURL string) ([]datastore.WatchReturn, error, int) {
	log.Println(tlog.Log()+"recordID:", recordID, "apiUrl:", apiURL)

	var watchesToReturn []datastore.WatchReturn
	var watch datastore.WatchGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			WATCH_COLUMN_RECORD_ID+","+
			WATCH_COLUMN_SUBSCRIPTION_ID+","+
			WATCH_COLUMN_KIND+","+
			WATCH_COLUMN_PATH+","+
			WATCH_COLUMN_CRN_FULL+","+
			WATCH_COLUMN_WILDCARDS+","+
			WATCH_COLUMN_RECORD_ID_TO_WATCH+","+
			WATCH_COLUMN_SUBSCRIPTION_EMAIL+" FROM "+
			WATCH_TABLE_NAME+" WHERE "+
			WATCH_COLUMN_RECORD_ID_TO_WATCH+" = $1",
			recordID)

		if err != nil {
			log.Println(tlog.Log()+": rowcount error: ", err.Error())
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}
		defer rows.Close()

		watchesToReturn = []datastore.WatchReturn{}
		for rows.Next() {
			watch = datastore.WatchGet{}
			err = rows.Scan(&watch.RecordID,
				&watch.SubscriptionRecordID,
				&watch.Kind,
				&watch.Path,
				&watch.CRNFull,
				&watch.Wildcards,
				&watch.RecordIDToWatch,
				&watch.SubscriptionEmail)
			if err != nil {
				log.Println(tlog.Log()+": Error in GetSubscriptionAlls: ", err)
				return nil, err, http.StatusInternalServerError
			}
			watchToReturn := datastore.ConvertWatchGetToWatchReturn(watch, apiURL)
			watchesToReturn = append(watchesToReturn, watchToReturn)

		}
		if err = rows.Err(); err != nil {
			log.Printf(tlog.Log()+"rows.Err(): ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// explicitly do rows close
				if err = rows.Close(); err != nil {
					log.Println(tlog.Log()+"Error in rows.Close ", err)
				}
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}
		// explicitly do rows close
		if err = rows.Close(); err != nil {
			log.Println(tlog.Log()+"Error in rows.Close ", err)
		}
		retry = false
		//log.Printf(tlog.Log()+": Watches are %s\n", watchesToReturn)
	}
	return watchesToReturn, nil, http.StatusOK

}

// GetWatchesByQuery query is a string that has a format like "x=1&y=2&y=3", where the key like x,y,z, etc. must be one of the following values:
// crn, version, cname, ctype, service_name, location, scope, service_instance, resource_type, resource, kind, subscription_id.
//
// When key is crn, it must be in crn string format, and it can only support single value, for example,
// "crn=crn:version:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"
// If a crn attribute is missing, wildcard is applied in the search.
// If you want an exact matching, not wildcard matching, then specify all crn sttributes,
// e.g. "version=v1&cname=cname1&ctype=ctype1&service-name=service-name1&location=location1&scope=scope1&service-instance=service-instance1&resource_type=&resource="
//
// When key is version, cname, ctype, service_name, location, scope, service_instance, resource_type, resource, kind, or subscription_id,
// it can support multiple values with comma separated, for example, query can be
// "kind=incident,maintenance&subscription_id=3c793c70-adfb-405a-918b-d14e10fd32b5,1d909eb3-1e44-4f81-904f-1cc52bfecd51&cnmae=cname1,cname2"
//
// If the query string cannot be parsed by net/url/ParseQuery function, error will be returned.
// limit must be > 0, offset must be >= 0. If limit is <=0, will be no limit
//
// Example 1: GetWatchesByQuery("crn=crn:v1:cname1:ctype1:service-name1:::::", 50, 25) will return all watches that matches "crn:v1:cname1:ctype1:service-name1:::::" crn string with wildcard is ON.
//            The WHERE clause used in the database search is "(version='v1'} AND (cname='cname1') AND (ctype='ctype1') AND (service_name='service-name1')"
//            The function will return 3 outputs: 1) *[]datastore.WatchReturn which contains 50 records starting at offset 25,
//                                                2) total number of records that satisfy the conditions in the query
//                                                3) Error occurs in this function call
//
// Example 2: GetWatchesByQuery("version=v1&cname=cname1&ctype=ctype1&service_name1=service-name1&location=location1&scope=&service_instance=&resource_type=&resource-type1&resource=", 0, 0)
//            will return all watches that matches all the crn attributes.
//            When all crn attributes are specified in the query, it will be an exact match (not wildcard), if any crn attribute is missing, the missing attribute will become a wildcard match.
//            The WHERE clause used in the database search is "(version='v1'} AND (cname='cname1') AND (ctype='ctype1') AND (service_name='service-name1') AND (location='location1) AND (scope='') AND (service_instance='') AND (resource_type='resource-type1') AND (resource='')"
//            The function will return 3 outputs: 1) *[]datastore.WatchReturn which contains all records starting at offset 0,
//                                                2) total number of records that satisfy the conditions in the query
//                                                3) Error occurs in this function call
//
// Example 3: GetWatchesByQuery("crn=crn:v1:cname1:ctype1:service-name1:::::&kind=incident", 10, 0)
//            will return all watches that matches "crn:v1:cname1:ctype1:service-name1:::::" crn string with wildcard is ON, and kind is incident
//            The WHERE clause used in the database search is "(version='v1'} AND (cname='cname1') AND (ctype='ctype1') AND (service_name='service-name1') AND (kind='incident')"
//            The function will return 3 outputs: 1) *[]datastore.WatchReturn which contains 10 records starting at offset 0,
//                                                2) total number of records that satisfy the conditions in the query
//                                                3) Error occurs in this function call
//
// Example 4: GetWatchesByQuery("crn=crn:v1:cname1:ctype1:service-name1:::::&kind=incident,maintenance&subscription_id=3c793c70-adfb-405a-918b-d14e10fd32b5", 0, 1)
//            will return all watches that matches "crn:v1:cname1:ctype1:service-name1:::::" crn string with wildcard is ON, and kind is 'incident' or 'maintenance', and SubscriptionRecordID is '3c793c70-adfb-405a-918b-d14e10fd32b5'
//            The WHERE clause used in the database search is "(version='v1'} AND (cname='cname1') AND (ctype='ctype1') AND (service_name='service-name1') AND (kind='incident' OR kind='maintenance') AND (subscription_id='3c793c70-adfb-405a-918b-d14e10fd32b5')
//            The function will return 3 outputs: 1) *[]datastore.WatchReturn which contains all records starting at offset 1,
//                                                2) total number of records that satisfy the conditions in the query
//                                                3) Error occurs in this function call
//
func GetWatchesByQuery(database *sql.DB, query string, limit int, offset int, apiURL string) (*[]datastore.WatchReturn, int, error, int) {
	log.Println(tlog.Log()+"query:", query, "limit:", limit, "offset:", offset, "apiUrl:", apiURL)

	queryParms, err := url.ParseQuery(query)
	if err != nil {
		//log.Println(tlog.Log()+"Error: ", err)
		return nil, 0, err, http.StatusBadRequest
	}

	hasQuery := false
	hasCRNQuery := false
	hasOtherQuery := false

	// checking query
	if len(queryParms[RESOURCE_QUERY_VERSION]) == 0 && len(queryParms[RESOURCE_QUERY_CNAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CTYPE]) == 0 && len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_LOCATION]) == 0 && len(queryParms[RESOURCE_QUERY_SCOPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) == 0 && len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_RESOURCE]) == 0 && len(queryParms[RESOURCE_QUERY_CRN]) == 0 &&
		len(queryParms[WATCH_QUERY_SUBSCRIPTION_ID]) == 0 && len(queryParms[WATCH_QUERY_KIND]) == 0 {

		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters. : query:[" + query + "]"
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	// ======== check offset ========
	if offset < 0 {
		// handle error
		msg := tlog.Log() + "ERROR: offset must be a non-negative integer."
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	if len(queryParms[RESOURCE_QUERY_VERSION]) > 0 || len(queryParms[RESOURCE_QUERY_CNAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_CTYPE]) > 0 || len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_LOCATION]) > 0 || len(queryParms[RESOURCE_QUERY_SCOPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) > 0 || len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_RESOURCE]) > 0 {

		if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
			msg := tlog.Log() + "ERROR: You cannot have crn string and also crn attribute in the query. : query:[" + query + "]"
			//log.Println(msg)
			return nil, 0, errors.New(msg), http.StatusBadRequest
		}
		hasCRNQuery = true
	} else if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
		hasCRNQuery = true
	}

	if len(queryParms[WATCH_QUERY_SUBSCRIPTION_ID]) > 0 || len(queryParms[WATCH_QUERY_KIND]) > 0 {

		hasOtherQuery = true
	}

	whereClause := ""
	orderBy := ""
	var crnErr error

	if hasCRNQuery {
		whereClause, orderBy, crnErr = CreateCRNSearchWhereClauseFromQuery(queryParms, RESOURCE_TABLE_PREFIX)
		orderBy = WATCH_COLUMN_RECORD_ID // order by RecordID
		if crnErr != nil {
			msg := "Error: cannot create where clause for CRN query"
			//log.Println(tlog.Log()+msg, err)
			return nil, 0, errors.New(msg), http.StatusInternalServerError
		} else if strings.TrimSpace(whereClause) == "" {
			hasQuery = false
		} else {
			hasQuery = true
		}
	}

	if hasOtherQuery {
		// subscription record id
		subscriptionIDQueryList := ParseQueryCondition(queryParms, WATCH_QUERY_SUBSCRIPTION_ID, false)

		// kind
		kindQueryList := ParseQueryCondition(queryParms, WATCH_QUERY_KIND, false)

		// ======== Form where clause ========
		orderBy = WATCH_COLUMN_RECORD_ID // order by RecordID
		// where subscriptio_id
		whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, subscriptionIDQueryList, WATCH_COLUMN_SUBSCRIPTION_ID, WATCH_TABLE_PREFIX)

		// where kind
		whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, kindQueryList, WATCH_COLUMN_KIND, WATCH_TABLE_PREFIX)
	}

	orderByClause := ""
	limitClause := ""
	offsetClause := ""
	selectStmt := ""
	totalCount := 0
	watchesToReturn := []datastore.WatchReturn{}

	if strings.TrimSpace(whereClause) == "" {
		hasQuery = false
	} else {
		hasQuery = true
		if orderBy != "" {
			orderByClause = " ORDER BY " + orderBy
		}

		// ======== LIMIT =========
		if limit > 0 {
			limitClause = " LIMIT " + strconv.Itoa(limit)
		}

		// ======== OFFSET =========
		if offset > 0 {
			offsetClause = " OFFSET " + strconv.Itoa(offset)
		}
	}

	// ========= database Query =========
	//log.Println("whereClause: ", whereClause)
	//log.Println("orderByClause: ", orderByClause)
	//log.Println("limitClause: ", limitClause)
	//log.Println("offsetClause: ", offsetClause)

	// Commented out https://github.ibm.com/cloud-sre/pnp-abstraction/issues/646
	// if hasCRNQuery {

	// 	selectStmt = "SELECT DISTINCT " + SELECT_WATCH_COLUMNS + " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS +
	// 		" inner join " + WATCH_JUNCTION_TABLE_NAME + " " + JUNCTION_TABLE_ALIAS + " on " +
	// 		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "=" + JUNCTION_TABLE_PREFIX + WATCHJUNCTION_COLUMN_RESOURCE_ID +
	// 		" inner join " + WATCH_TABLE_NAME + " " + WATCH_TABLE_ALIAS + " on " +
	// 		JUNCTION_TABLE_PREFIX + WATCHJUNCTION_COLUMN_WATCH_ID + "=" + WATCH_TABLE_PREFIX + WATCH_COLUMN_RECORD_ID +
	// 		" WHERE " + whereClause

	// } else if hasOtherQuery {

	// do not have CRN query, only have subscription_id and/or kind
	selectStmt = "SELECT " + SELECT_WATCH_COLUMNS + " FROM " + WATCH_TABLE_NAME + " " + WATCH_TABLE_ALIAS + " WHERE " + whereClause

	// }

	//log.Println("selectStmt: ", selectStmt)
	executeStmt := "WITH cte AS (" + selectStmt + ") SELECT * FROM ( TABLE cte " + orderByClause + limitClause + offsetClause + ") sub RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true;"

	log.Println(tlog.Log()+"DEBUG: executeStmt: ", executeStmt)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(executeStmt)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Watches found.")
			return nil, 0, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, 0, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			watchesToReturn = []datastore.WatchReturn{}
			for rows.Next() {
				var w datastore.WatchGetNull

				err = rows.Scan(&w.RecordID, &w.SubscriptionRecordID, &w.Kind, &w.Path, &w.CRNFull, &w.Wildcards, &w.RecordIDToWatch, &w.SubscriptionEmail, &totalCount)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to WatchGetNull Error: %v", err)
					return nil, 0, err, http.StatusInternalServerError
				}

				// if RecordID is not null, then rescan into WatchGet struct
				if w.RecordID.Valid && totalCount > 0 {
					var wg datastore.WatchGet
					err = rows.Scan(&wg.RecordID, &wg.SubscriptionRecordID, &wg.Kind, &wg.Path, &wg.CRNFull, &wg.Wildcards, &wg.RecordIDToWatch, &wg.SubscriptionEmail, &totalCount)

					if err != nil {
						log.Printf(tlog.Log()+"Row Scan to IncidentGet Error: %v", err)
						return nil, 0, err, http.StatusInternalServerError
					}

					watchReturn := datastore.ConvertWatchGetToWatchReturn(wg, apiURL)
					watchesToReturn = append(watchesToReturn, watchReturn)
				}
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, 0, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}
		}

		if err != nil {
			return &watchesToReturn, totalCount, err, http.StatusInternalServerError
		}
		retry = false
	}
	return &watchesToReturn, totalCount, err, http.StatusOK
}

// GetSubscriptionByRecordIDStatement is called from pnp-subscrition get a subscription record by record id
func GetSubscriptionByRecordIDStatement(database *sql.DB, recordID string) (*datastore.SubscriptionReturn, error, int) {
	log.Println(tlog.Log()+"recordID:", recordID)

	var subscription datastore.SubscriptionGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {
		err := database.QueryRow("SELECT "+
			SUBSCRIPTION_COLUMN_RECORD_ID+","+
			SUBSCRIPTION_COLUMN_NAME+","+
			SUBSCRIPTION_COLUMN_TARGET_ADDRESS+","+
			SUBSCRIPTION_COLUMN_TARGET_TOKEN+","+
			SUBSCRIPTION_COLUMN_EXPIRATION+" FROM "+
			SUBSCRIPTION_TABLE_NAME+" WHERE "+
			SUBSCRIPTION_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&subscription.RecordID,
			&subscription.Name,
			&subscription.TargetAddress,
			&subscription.TargetToken,
			&subscription.Expiration)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Subscription with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			//log.Printf("Subscription is %s\n", subscription.RecordID)
			subscriptionToReturn := datastore.ConvertSubGetToSubReturn(&subscription)
			return subscriptionToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusInternalServerError

}

// GetSubscriptionAll is called from pnp-subscrition gets all subscription record in the database
func GetSubscriptionAll(dbConnection *sql.DB) ([]datastore.SubscriptionReturn, error, int) {
	log.Println(tlog.Log())

	var Subscriptions []datastore.SubscriptionReturn
	var subscription datastore.SubscriptionGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := dbConnection.Query("SELECT " +
			SUBSCRIPTION_COLUMN_RECORD_ID + "," +
			SUBSCRIPTION_COLUMN_NAME + "," +
			SUBSCRIPTION_COLUMN_TARGET_ADDRESS + "," +
			SUBSCRIPTION_COLUMN_TARGET_TOKEN + "," +
			SUBSCRIPTION_COLUMN_EXPIRATION + " FROM " +
			SUBSCRIPTION_TABLE_NAME + " WHERE " +
			SUBSCRIPTION_COLUMN_NAME + " like '%'")

		if err != nil {
			log.Println(tlog.Log()+"rowcount error: ", err.Error())
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}
		defer rows.Close()

		Subscriptions = []datastore.SubscriptionReturn{}
		for rows.Next() {
			err = rows.Scan(&subscription.RecordID,
				&subscription.Name,
				&subscription.TargetAddress,
				&subscription.TargetToken,
				&subscription.Expiration)
			if err != nil {
				log.Println(tlog.Log()+"Error: ", err)
				return nil, err, http.StatusInternalServerError
			}
			subscriptionToReturn := datastore.ConvertSubGetToSubReturn(&subscription)
			Subscriptions = append(Subscriptions, *subscriptionToReturn)
		}
		if err = rows.Err(); err != nil {
			log.Printf(tlog.Log()+"rows.Err(): ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				// explicitly do rows close
				if err = rows.Close(); err != nil {
					log.Println(tlog.Log()+"Error in rows.Close ", err)
				}
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		}
		// explicitly do rows close
		if err = rows.Close(); err != nil {
			log.Println(tlog.Log()+"Error in rows.Close ", err)
		}
		retry = false
	}
	return Subscriptions, nil, http.StatusOK

}

func getVisibilityByRecordIDStatement(database *sql.DB, recordID string) (*datastore.VisibilityGet, error, int) {
	var visibilityToReturn datastore.VisibilityGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow("SELECT "+
			VISIBILITY_COLUMN_RECORD_ID+","+
			VISIBILITY_COLUMN_NAME+","+
			VISIBILITY_COLUMN_DESCRIPTION+" FROM "+
			VISIBILITY_TABLE_NAME+" WHERE "+
			VISIBILITY_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&visibilityToReturn.RecordID,
			&visibilityToReturn.Name,
			&visibilityToReturn.Description)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Visibility with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			//log.Printf(tlog.Log()+"Visibility is %s\n", visibilityToReturn.RecordID)
			return &visibilityToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusInternalServerError

}

func getTagByRecordIDStatement(database *sql.DB, recordID string) (*datastore.TagGet, error, int) {
	var tagToReturn datastore.TagGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow("SELECT "+
			TAG_COLUMN_RECORD_ID+","+
			TAG_COLUMN_ID+" FROM "+
			TAG_TABLE_NAME+" WHERE "+
			TAG_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&tagToReturn.RecordID,
			&tagToReturn.ID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Tag with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			//log.Printf(tlog.Log()+"Tag is %s\n", tagToReturn.RecordID)
			return &tagToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusInternalServerError

}

func GetAllSNMaintenances(database *sql.DB) ([]*datastore.MaintenanceReturn, error, int) {
	log.Println(tlog.Log())

	maintenancesToReturn := []*datastore.MaintenanceReturn{}

	//	var disruptive string
	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			MAINTENANCE_COLUMN_RECORD_ID+","+
			MAINTENANCE_COLUMN_SOURCE_ID+","+
			MAINTENANCE_COLUMN_SOURCE+","+
			MAINTENANCE_COLUMN_RECORD_HASH+" FROM "+
			MAINTENANCE_TABLE_NAME+" WHERE "+
			MAINTENANCE_COLUMN_SOURCE+" = $1 AND "+
			MAINTENANCE_COLUMN_PNP_REMOVED+" = $2",
			"servicenow", "false")

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Maintenance found.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			maintenancesToReturn = []*datastore.MaintenanceReturn{}
			for rows.Next() {
				/* var n datastore.MaintenanceGetNull
				err = rows.Scan(&n.RecordID, &n.SourceID, &n.Source, &n.RecordHash)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to MaintenanceGetNull Error: %v", err)
					return nil, 0, err, http.StatusInternalServerError
				}

				// if RecordID is not null, then rescan into MaintenanceGet struct
				if n.RecordID.Valid {
				*/
				var m datastore.MaintenanceGet
				err = rows.Scan(&m.RecordID, &m.SourceID, &m.Source, &m.RecordHash)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to MaintenanceGet Error: %v", err)
					return nil, err, http.StatusInternalServerError
				}

				maintenanceReturn := ConvertMaintenanceGetToMaintenanceReturn(m)
				maintenancesToReturn = append(maintenancesToReturn, &maintenanceReturn)
				/* } else {
					// the whole row has Null values, except total_count. Could be because offset is out of result set range
					break
				} */
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

		}

		if err != nil {
			return maintenancesToReturn, err, http.StatusInternalServerError
		}
		retry = false
	}
	return maintenancesToReturn, nil, http.StatusOK

}

func GetMaintenanceByRecordIDStatement(database *sql.DB, recordID string) (*datastore.MaintenanceReturn, error, int) {
	log.Println(tlog.Log()+"recordID:", recordID)

	var maintenanceGet datastore.MaintenanceGet

	//	var disruptive string

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow("SELECT "+
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
			MAINTENANCE_COLUMN_TARGETED_URL+","+
			MAINTENANCE_COLUMN_COMPLETION_CODE+","+
			MAINTENANCE_COLUMN_AUDIENCE+
			" FROM "+MAINTENANCE_TABLE_NAME+
			" WHERE "+
			MAINTENANCE_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&maintenanceGet.RecordID,
			&maintenanceGet.PnpCreationTime,
			&maintenanceGet.PnpUpdateTime,
			&maintenanceGet.SourceCreationTime,
			&maintenanceGet.SourceUpdateTime,
			&maintenanceGet.PlannedStartTime,
			&maintenanceGet.PlannedEndTime,
			&maintenanceGet.ShortDescription,
			&maintenanceGet.LongDescription,
			&maintenanceGet.CRNFull,
			&maintenanceGet.State,
			&maintenanceGet.Disruptive,
			&maintenanceGet.SourceID,
			&maintenanceGet.Source,
			&maintenanceGet.RecordHash,
			&maintenanceGet.MaintenanceDuration,
			&maintenanceGet.DisruptionType,
			&maintenanceGet.DisruptionDescription,
			&maintenanceGet.DisruptionDuration,
			&maintenanceGet.RegulatoryDomain,
			&maintenanceGet.PnPRemoved,
			&maintenanceGet.TargetedURL,
			&maintenanceGet.CompletionCode,
			&maintenanceGet.Audience)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Maintenance with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			//log.Printf("Maintenance is %s\n", maintenanceGet.RecordID)
			maintenanceToReturn := ConvertMaintenanceGetToMaintenanceReturn(maintenanceGet)
			return &maintenanceToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusInternalServerError

}

// GetMaintenanceBySourceID get a maaintenace_table record from the database using the record_id, the record_id is obtained by using
// the passing source and source_id and calling CreateRecordIDFromSourceSourceID
func GetMaintenanceBySourceID(database *sql.DB, source string, sourceID string) (*datastore.MaintenanceReturn, error, int) {

	if strings.TrimSpace(source) == "" {
		return nil, errors.New("source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(sourceID) == "" {
		return nil, errors.New("sourceID cannot be empty"), http.StatusBadRequest
	}

	recordID := CreateRecordIDFromSourceSourceID(source, sourceID)

	return GetMaintenanceByRecordIDStatement(database, recordID)
}

// GetMaintenanceByQuery -query is a string that has a format like "x=1&y=2&y=3", where the key like x,y,z, etc. must be one of the following values:
// crn, version, cname, ctype, service_name, location, scope, service_instance, resource_type, resource, state, disruptive, creation_time_start, creation_time_end, update_time_start, update_time_end, planned_start_start, planned_start_end, source, planned_end_start, planned_end_end,
// When key is crn, it must be crn string format, and it can only support single value, cannot be comma separated values, for example,
// "crn=crn:version:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"
// If a crn attribute is missing, wildcard is applied in matching.
// If you want an exact matching, not wildcard matching, then specify all crn sttributes. When an attribute is empty, you still have to specify the key,
// e.g. "version=v1&cname=cname1&ctype=ctype1&service-name=service-name1&location=location1&scope=scope1&service-instance=service-instance1&resource_type=&resource="
//
// For all the timestamp related keys, like creation_time_start, creation_time_end, update_time_start, update_time_end, planned_start_start, planned_start_end,
// they can only be single value, cannot be comma separated values.
//
// For other keys, their values can be comma separated. For example, "cname=abc,def&state=xxx,yyy"
//
// The rows returned are sorted by source + source_id.
// If the query string cannot be parsed by net/url/ParseQuery function, error will be returned.
// limit must be > 0, offset must be >= 0. If limit is <=0, will be no limit
//
// See GetWatchesByQuery function for examples for examples on crn wildcard and non-wildcard search.
//
func GetMaintenanceByQuery(database *sql.DB, query string, limit int, offset int) (*[]datastore.MaintenanceReturn, int, error, int) {
	log.Println(tlog.Log()+"query:", query, "limit:", limit, "offset:", offset)

	queryParms, err := url.ParseQuery(query)
	if err != nil {
		//log.Println(tlog.Log()+"Error: ", err)
		return nil, 0, err, http.StatusInternalServerError
	}

	hasQuery := false
	hasCRNQuery := false
	hasOtherQuery := false
	hasPnPRemovedQuery := false

	// checking query
	if len(queryParms[RESOURCE_QUERY_VERSION]) == 0 && len(queryParms[RESOURCE_QUERY_CNAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CTYPE]) == 0 && len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_LOCATION]) == 0 && len(queryParms[RESOURCE_QUERY_SCOPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) == 0 && len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_RESOURCE]) == 0 && len(queryParms[RESOURCE_QUERY_CRN]) == 0 &&
		len(queryParms[MAINTENANCE_QUERY_CREATION_TIME_START]) == 0 && len(queryParms[MAINTENANCE_QUERY_CREATION_TIME_END]) == 0 &&
		len(queryParms[MAINTENANCE_QUERY_UPDATE_TIME_START]) == 0 && len(queryParms[MAINTENANCE_QUERY_UPDATE_TIME_END]) == 0 &&
		len(queryParms[MAINTENANCE_QUERY_PNP_CREATION_TIME_START]) == 0 && len(queryParms[MAINTENANCE_QUERY_PNP_CREATION_TIME_END]) == 0 &&
		len(queryParms[MAINTENANCE_QUERY_PNP_UPDATE_TIME_START]) == 0 && len(queryParms[MAINTENANCE_QUERY_PNP_UPDATE_TIME_END]) == 0 &&
		len(queryParms[MAINTENANCE_QUERY_PLANNED_START_START]) == 0 && len(queryParms[MAINTENANCE_QUERY_PLANNED_START_END]) == 0 &&
		len(queryParms[MAINTENANCE_QUERY_SOURCE]) == 0 && len(queryParms[MAINTENANCE_COLUMN_PNP_REMOVED]) == 0 &&
		len(queryParms[MAINTENANCE_QUERY_PLANNED_END_START]) == 0 && len(queryParms[MAINTENANCE_QUERY_PLANNED_END_END]) == 0 {

		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters. : query:[" + query + "]"
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	// ======== check offset ========
	if offset < 0 {
		// handle error
		msg := tlog.Log() + "ERROR: offset must be a non-negative integer."
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	if len(queryParms[RESOURCE_QUERY_VERSION]) > 0 || len(queryParms[RESOURCE_QUERY_CNAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_CTYPE]) > 0 || len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_LOCATION]) > 0 || len(queryParms[RESOURCE_QUERY_SCOPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) > 0 || len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_RESOURCE]) > 0 {

		if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
			msg := tlog.Log() + "ERROR: You cannot have crn string and also crn attribute in the query. : query:[" + query + "]"
			//log.Println(msg)
			return nil, 0, errors.New(msg), http.StatusBadRequest
		}
		hasCRNQuery = true
	} else if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
		hasCRNQuery = true
	}

	if len(queryParms[MAINTENANCE_QUERY_CREATION_TIME_START]) > 0 ||
		len(queryParms[MAINTENANCE_QUERY_PNP_CREATION_TIME_START]) > 0 || len(queryParms[MAINTENANCE_QUERY_PNP_CREATION_TIME_END]) > 0 ||
		len(queryParms[MAINTENANCE_QUERY_PNP_UPDATE_TIME_START]) > 0 || len(queryParms[MAINTENANCE_QUERY_PNP_UPDATE_TIME_END]) > 0 ||
		len(queryParms[MAINTENANCE_QUERY_CREATION_TIME_END]) > 0 || len(queryParms[MAINTENANCE_QUERY_UPDATE_TIME_START]) > 0 ||
		len(queryParms[MAINTENANCE_QUERY_UPDATE_TIME_END]) > 0 || len(queryParms[MAINTENANCE_QUERY_PLANNED_START_START]) > 0 ||
		len(queryParms[MAINTENANCE_QUERY_PLANNED_START_END]) > 0 || len(queryParms[MAINTENANCE_COLUMN_STATE]) > 0 ||
		len(queryParms[MAINTENANCE_COLUMN_DISRUPTIVE]) > 0 || len(queryParms[MAINTENANCE_QUERY_SOURCE]) > 0 ||
		len(queryParms[MAINTENANCE_QUERY_PLANNED_END_START]) > 0 || len(queryParms[MAINTENANCE_QUERY_PLANNED_END_END]) > 0 ||
		len(queryParms[MAINTENANCE_COLUMN_PNP_REMOVED]) > 0 || len(queryParms[MAINTENANCE_COLUMN_TARGETED_URL]) > 0 ||
		len(queryParms[MAINTENANCE_COLUMN_AUDIENCE]) > 0 {
		hasOtherQuery = true
	}

	if len(queryParms[MAINTENANCE_COLUMN_PNP_REMOVED]) > 0 {
		hasPnPRemovedQuery = true
	}

	whereClause := ""
	orderBy := ""
	var crnErr error

	if hasCRNQuery {
		whereClause, orderBy, crnErr = CreateCRNSearchWhereClauseFromQuery(queryParms, RESOURCE_TABLE_PREFIX)
		orderBy = MAINTENANCE_COLUMN_SOURCE + ", " + MAINTENANCE_COLUMN_SOURCE_ID // order by Source, SourceID
		if crnErr != nil {
			msg := "Error: cannot create where clause for CRN query"
			//log.Println(tlog.Log()+msg, err)
			return nil, 0, errors.New(msg), http.StatusBadRequest
		} else if strings.TrimSpace(whereClause) == "" {
			hasQuery = false
		} else {
			hasQuery = true
		}
	}

	if hasOtherQuery {
		// creation_time_start
		createStartQueryList, createStartErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_CREATION_TIME_START)
		if createStartErr != nil {
			//log.Println(tlog.Log() + createStartErr.Error())
			return nil, 0, createStartErr, http.StatusBadRequest
		}

		// creation_time_end
		createEndQueryList, createEndErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_CREATION_TIME_END)
		if createEndErr != nil {
			//log.Println(tlog.Log() + createEndErr.Error())
			return nil, 0, createEndErr, http.StatusBadRequest
		}

		// pnp_creation_time_start
		pnpCreateStartQueryList, pnpCreateStartErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PNP_CREATION_TIME_START)
		if pnpCreateStartErr != nil {
			//log.Println(tlog.Log() + pnpCreateStartErr.Error())
			return nil, 0, pnpCreateStartErr, http.StatusBadRequest
		}

		// pnp_creation_time_end
		pnpCreateEndQueryList, pnpCreateEndErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PNP_CREATION_TIME_END)
		if pnpCreateEndErr != nil {
			//log.Println(tlog.Log() + pnpCreateEndErr.Error())
			return nil, 0, pnpCreateEndErr, http.StatusBadRequest
		}

		// pnp_update_time_start
		pnpUpdateStartQueryList, pnpUpdateStartErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PNP_UPDATE_TIME_START)
		if pnpUpdateStartErr != nil {
			//log.Println(tlog.Log() + pnpUpdateStartErr.Error())
			return nil, 0, pnpUpdateStartErr, http.StatusBadRequest
		}

		// pnp_update_time_end
		pnpUpdateEndQueryList, pnpUpdateEndErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PNP_UPDATE_TIME_END)
		if pnpUpdateEndErr != nil {
			//log.Println(tlog.Log() + pnpUpdateEndErr.Error())
			return nil, 0, pnpUpdateEndErr, http.StatusBadRequest
		}

		// update_time_start
		updateStartQueryList, updateStartErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_UPDATE_TIME_START)
		if updateStartErr != nil {
			//log.Println(tlog.Log() + updateStartErr.Error())
			return nil, 0, updateStartErr, http.StatusBadRequest
		}

		// update_time_end
		updateEndQueryList, updateEndErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_UPDATE_TIME_END)
		if updateEndErr != nil {
			//log.Println(tlog.Log() + updateEndErr.Error())
			return nil, 0, updateEndErr, http.StatusBadRequest
		}

		// planned_start_start
		plannedStartStartQueryList, plannedStartStartErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PLANNED_START_START)
		if plannedStartStartErr != nil {
			//log.Println(tlog.Log() + plannedStartStartErr.Error())
			return nil, 0, plannedStartStartErr, http.StatusBadRequest
		}

		// planned_start_end
		plannedStartEndQueryList, plannedStartEndErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PLANNED_START_END)
		if plannedStartEndErr != nil {
			//log.Println(tlog.Log() + plannedStartEndErr.Error())
			return nil, 0, plannedStartEndErr, http.StatusBadRequest
		}

		// planned_end_start
		plannedEndStartQueryList, plannedEndStartErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PLANNED_END_START)
		if plannedEndStartErr != nil {
			//log.Println(tlog.Log() + plannedEndStartErr.Error())
			return nil, 0, plannedEndStartErr, http.StatusBadRequest
		}

		// planned_end_end
		plannedEndEndQueryList, plannedEndEndErr := ParseQueryTimeCondition(queryParms, MAINTENANCE_QUERY_PLANNED_END_END)
		if plannedEndEndErr != nil {
			//log.Println(tlog.Log() + plannedEndEndErr.Error())
			return nil, 0, plannedEndEndErr, http.StatusBadRequest
		}

		stateQueryList := ParseQueryCondition(queryParms, MAINTENANCE_COLUMN_STATE, true)

		disruptiveQueryList := ParseQueryCondition(queryParms, MAINTENANCE_COLUMN_DISRUPTIVE, true)

		sourceQueryList := ParseQueryCondition(queryParms, MAINTENANCE_COLUMN_SOURCE, false)

		pnpRemovedQueryList := ParseQueryCondition(queryParms, MAINTENANCE_COLUMN_PNP_REMOVED, true)

		// =========== other query where clause ===========
		// where source_creation_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, createStartQueryList, MAINTENANCE_COLUMN_SOURCE_CREATION_TIME, MAINTENANCE_TABLE_PREFIX, ">=")

		// where source_creation_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, createEndQueryList, MAINTENANCE_COLUMN_SOURCE_CREATION_TIME, MAINTENANCE_TABLE_PREFIX, "<=")

		// where pnp_create_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpCreateStartQueryList, MAINTENANCE_COLUMN_PNP_CREATION_TIME, MAINTENANCE_TABLE_PREFIX, ">=")

		// where pnp_create_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpCreateEndQueryList, MAINTENANCE_COLUMN_PNP_CREATION_TIME, MAINTENANCE_TABLE_PREFIX, "<=")

		// where pnp_update_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpUpdateStartQueryList, MAINTENANCE_COLUMN_PNP_UPDATE_TIME, MAINTENANCE_TABLE_PREFIX, ">=")

		// where pnp_update_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpUpdateEndQueryList, MAINTENANCE_COLUMN_PNP_UPDATE_TIME, MAINTENANCE_TABLE_PREFIX, "<=")

		// where source_update_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, updateStartQueryList, MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME, MAINTENANCE_TABLE_PREFIX, ">=")

		// where source_update_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, updateEndQueryList, MAINTENANCE_COLUMN_SOURCE_UPDATE_TIME, MAINTENANCE_TABLE_PREFIX, "<=")

		// where start_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, plannedStartStartQueryList, MAINTENANCE_COLUMN_START_TIME, MAINTENANCE_TABLE_PREFIX, ">=")

		// where start_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, plannedStartEndQueryList, MAINTENANCE_COLUMN_START_TIME, MAINTENANCE_TABLE_PREFIX, "<=")

		// where end_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, plannedEndStartQueryList, MAINTENANCE_COLUMN_END_TIME, MAINTENANCE_TABLE_PREFIX, ">=")

		// where end_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, plannedEndEndQueryList, MAINTENANCE_COLUMN_END_TIME, MAINTENANCE_TABLE_PREFIX, "<=")

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, stateQueryList, MAINTENANCE_COLUMN_STATE, MAINTENANCE_TABLE_PREFIX)

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, disruptiveQueryList, MAINTENANCE_COLUMN_DISRUPTIVE, MAINTENANCE_TABLE_PREFIX)

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, sourceQueryList, MAINTENANCE_COLUMN_SOURCE, MAINTENANCE_TABLE_PREFIX)

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, pnpRemovedQueryList, MAINTENANCE_COLUMN_PNP_REMOVED, MAINTENANCE_TABLE_PREFIX)

		orderBy = MAINTENANCE_COLUMN_SOURCE + ", " + MAINTENANCE_COLUMN_SOURCE_ID // order by Source, SourceID
	}

	// if pnpRemoved is not specified in the query, default to query false only
	if !hasPnPRemovedQuery {
		pnpRemovedQueryList := []string{"false"}
		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, pnpRemovedQueryList, MAINTENANCE_COLUMN_PNP_REMOVED, MAINTENANCE_TABLE_PREFIX)
		orderBy = MAINTENANCE_COLUMN_SOURCE + ", " + MAINTENANCE_COLUMN_SOURCE_ID // order by Source, SourceID
	}

	orderByClause := ""
	limitClause := ""
	offsetClause := ""
	selectStmt := ""
	totalCount := 0
	maintenanceToReturn := []datastore.MaintenanceReturn{}

	if strings.TrimSpace(whereClause) == "" {
		hasQuery = false
	} else {
		hasQuery = true
		if orderBy != "" {
			orderByClause = " ORDER BY " + orderBy
		}

		// ======== LIMIT =========
		if limit > 0 {
			limitClause = " LIMIT " + strconv.Itoa(limit)
		}

		// ======== OFFSET =========
		if offset > 0 {
			offsetClause = " OFFSET " + strconv.Itoa(offset)
		}
	}

	// ========= database Query =========
	//log.Println("whereClause: ", whereClause)
	//log.Println("orderByClause: ", orderByClause)
	//log.Println("limitClause: ", limitClause)
	//log.Println("offsetClause: ", offsetClause)

	selectStmt = "SELECT DISTINCT " + SELECT_MAINTENANCE_COLUMNS +
		" FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS +
		" inner join " + MAINTENANCE_JUNCTION_TABLE_NAME + " " + JUNCTION_TABLE_ALIAS +
		" on " + RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "=" + JUNCTION_TABLE_PREFIX + MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID +
		" inner join " + MAINTENANCE_TABLE_NAME + " " + MAINTENANCE_TABLE_ALIAS +
		" on " + JUNCTION_TABLE_PREFIX + MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID + "=" + MAINTENANCE_TABLE_PREFIX + MAINTENANCE_COLUMN_RECORD_ID +
		" WHERE " + whereClause

	//log.Println("selectStmt: ", selectStmt)
	executeStmt := "WITH cte AS (" + selectStmt + ") SELECT * FROM ( TABLE cte " + orderByClause + limitClause + offsetClause + ") sub RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true;"

	log.Println("DEBUG: executeStmt: ", executeStmt)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(executeStmt)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Maintenance found.")
			return nil, 0, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, 0, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			maintenanceToReturn = []datastore.MaintenanceReturn{}
			for rows.Next() {
				var n datastore.MaintenanceGetNull
				err = rows.Scan(&n.RecordID, &n.PnpCreationTime, &n.PnpUpdateTime, &n.SourceCreationTime, &n.SourceUpdateTime, &n.PlannedStartTime,
					&n.PlannedEndTime, &n.ShortDescription, &n.LongDescription, &n.State, &n.Disruptive,
					&n.SourceID, &n.Source, &n.RegulatoryDomain, &n.RecordHash, &n.MaintenanceDuration, &n.DisruptionType,
					&n.DisruptionDescription, &n.DisruptionDuration, &n.CompletionCode, &n.CRNFull, &n.PnPRemoved, &n.TargetedURL, &n.Audience, &totalCount)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to MaintenanceGetNull Error: %v", err)
					return nil, 0, err, http.StatusInternalServerError
				}

				// if RecordID is not null, then rescan into MaintenanceGet struct
				if n.RecordID.Valid && totalCount > 0 {
					var m datastore.MaintenanceGet
					err = rows.Scan(&m.RecordID, &m.PnpCreationTime, &m.PnpUpdateTime, &m.SourceCreationTime, &m.SourceUpdateTime, &m.PlannedStartTime,
						&m.PlannedEndTime, &m.ShortDescription, &m.LongDescription, &m.State, &m.Disruptive, &m.SourceID, &m.Source, &m.RegulatoryDomain,
						&m.RecordHash, &m.MaintenanceDuration, &m.DisruptionType, &m.DisruptionDescription, &m.DisruptionDuration, &m.CompletionCode,
						&m.CRNFull, &m.PnPRemoved, &m.TargetedURL, &m.Audience, &totalCount)

					if err != nil {
						log.Printf(tlog.Log()+"Row Scan to MaintenanceGet Error: %v", err)
						return nil, 0, err, http.StatusInternalServerError
					}

					maintenanceReturn := ConvertMaintenanceGetToMaintenanceReturn(m)
					maintenanceToReturn = append(maintenanceToReturn, maintenanceReturn)
				} //else {
				// the whole row has Null values, except total_count. Could be because offset is out of result set range
				// break
				//}
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, 0, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

		}

		if err != nil {
			return &maintenanceToReturn, totalCount, err, http.StatusInternalServerError
		}
		retry = false
	}
	return &maintenanceToReturn, totalCount, err, http.StatusOK
}

// GetIncidentByRecordIDStatement get a incident_table record matching the record_id passed
func GetIncidentByRecordIDStatement(database *sql.DB, recordID string) (*datastore.IncidentReturn, error, int) {
	log.Println(tlog.Log()+"recordID:", recordID)

	var incidentGet datastore.IncidentGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow("SELECT "+
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
			INCIDENT_COLUMN_SOURCE_ID+","+
			INCIDENT_COLUMN_SOURCE+","+
			INCIDENT_COLUMN_REGULATORY_DOMAIN+","+
			INCIDENT_COLUMN_CRN_FULL+","+
			INCIDENT_COLUMN_AFFECTED_ACTIVITY+","+
			INCIDENT_COLUMN_PNP_REMOVED+","+
			INCIDENT_COLUMN_TARGETED_URL+","+
			INCIDENT_COLUMN_CUSTOMER_IMPACT_DESCRIPTION+","+
			INCIDENT_COLUMN_AUDIENCE+
			" FROM "+INCIDENT_TABLE_NAME+
			" WHERE "+
			INCIDENT_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&incidentGet.RecordID,
			&incidentGet.PnpCreationTime,
			&incidentGet.PnpUpdateTime,
			&incidentGet.SourceCreationTime,
			&incidentGet.SourceUpdateTime,
			&incidentGet.OutageStartTime,
			&incidentGet.OutageEndTime,
			&incidentGet.ShortDescription,
			&incidentGet.LongDescription,
			&incidentGet.State,
			&incidentGet.Classification,
			&incidentGet.Severity,
			&incidentGet.SourceID,
			&incidentGet.Source,
			&incidentGet.RegulatoryDomain,
			&incidentGet.CRNFull,
			&incidentGet.AffectedActivity,
			&incidentGet.PnPRemoved,
			&incidentGet.TargetedURL,
			&incidentGet.CustomerImpactDescription,
			&incidentGet.Audience)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Incident with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			//log.Printf("Incident is %s\n", incidentGet.RecordID)
			incidentToReturn := ConvertIncidentGetToIncidentReturn(incidentGet)
			return &incidentToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusInternalServerError

}

// GetIncidentBySourceID returns an incident_table record matching record_id, first get the record id by using the source and source_id and calling CreateRecordIDFromSourceSourceID funtion
func GetIncidentBySourceID(database *sql.DB, source string, sourceID string) (*datastore.IncidentReturn, error, int) {
	log.Println(tlog.Log()+"source:", source, "sourceID:", sourceID)

	if strings.TrimSpace(source) == "" {
		return nil, errors.New("source cannot be empty"), http.StatusBadRequest
	}

	if strings.TrimSpace(sourceID) == "" {
		return nil, errors.New("sourceID cannot be empty"), http.StatusBadRequest
	}

	recordID := CreateRecordIDFromSourceSourceID(source, sourceID)

	return GetIncidentByRecordIDStatement(database, recordID)
}

// GetIncidentByQuery query is a string that has a format like "x=1&y=2&y=3", where the key like x,y,z, etc. must be one of the following values:
// crn, version, cname, ctype, service_name, location, scope, service_instance, resource_type, resource, state, classification, severity, creation_time_start, creation_time_end, update_time_start, update_time_end, outage_start_start, outage_start_end, pnp_removed
// When key is crn, it must be crn string format, and it can only support single value, cannot be comma separated values, for example,
// "crn=crn:version:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"
// If a crn attribute is missing, wildcard is applied in the search.
// If you want an exact matching, not wildcard matching, then specify all crn sttributes,
// e.g. "version=v1&cname=cname1&ctype=ctype1&service-name=service-name1&location=location1&scope=scope1&service-instance=service-instance1&resource_type=&resource="
//
// For all the timestamp related keys, like creation_time_start, creation_time_end, update_time_start, update_time_end, outage_start_start, outage_start_end,
// they can only be single value, cannot be comma separated values.
//
// For other keys, their values can be comma separated. For example, "cname=abc,def&state=xxx,yyy". If "pnp_removed=true,false" is eqivalent to pnpRemoved="both" in PnP StatusApi
//
// The rows returned are sorted by source + source_id.
// If the query string cannot be parsed by net/url/ParseQuery function, error will be returned.
// limit must be > 0, offset must be >= 0. If limit is <=0, will be no limit
//
// See GetWatchesByQuery function for examples on crn wildcard and non-wildcard search.
//
func GetIncidentByQuery(database *sql.DB, query string, limit int, offset int) (*[]datastore.IncidentReturn, int, error, int) {
	log.Println(tlog.Log()+"query:", query, "limit:", limit, "offset:", offset)

	queryParms, err := url.ParseQuery(query)
	if err != nil {
		//log.Println(tlog.Log()+"Error: ", err)
		return nil, 0, err, http.StatusInternalServerError
	}

	hasQuery := false
	hasCRNQuery := false
	hasOtherQuery := false
	hasPnPRemovedQuery := false

	// checking query
	if len(queryParms[RESOURCE_QUERY_VERSION]) == 0 && len(queryParms[RESOURCE_QUERY_CNAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CTYPE]) == 0 && len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_LOCATION]) == 0 && len(queryParms[RESOURCE_QUERY_SCOPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) == 0 && len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_RESOURCE]) == 0 && len(queryParms[RESOURCE_QUERY_CRN]) == 0 &&
		len(queryParms[INCIDENT_QUERY_CREATION_TIME_START]) == 0 && len(queryParms[INCIDENT_QUERY_CREATION_TIME_END]) == 0 &&
		len(queryParms[INCIDENT_QUERY_UPDATE_TIME_START]) == 0 && len(queryParms[INCIDENT_QUERY_UPDATE_TIME_END]) == 0 &&
		len(queryParms[INCIDENT_QUERY_PNP_CREATION_TIME_START]) == 0 && len(queryParms[INCIDENT_QUERY_PNP_CREATION_TIME_END]) == 0 &&
		len(queryParms[INCIDENT_QUERY_PNP_UPDATE_TIME_START]) == 0 && len(queryParms[INCIDENT_QUERY_PNP_UPDATE_TIME_END]) == 0 &&
		len(queryParms[INCIDENT_QUERY_OUTAGE_START_START]) == 0 && len(queryParms[INCIDENT_QUERY_OUTAGE_START_END]) == 0 &&
		len(queryParms[INCIDENT_COLUMN_PNP_REMOVED]) == 0 {

		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters. : query:[" + query + "]"
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	// ======== check offset ========
	if offset < 0 {
		// handle error
		msg := tlog.Log() + "ERROR: offset must be a non-negative integer."
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	if len(queryParms[RESOURCE_QUERY_VERSION]) > 0 || len(queryParms[RESOURCE_QUERY_CNAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_CTYPE]) > 0 || len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_LOCATION]) > 0 || len(queryParms[RESOURCE_QUERY_SCOPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) > 0 || len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_RESOURCE]) > 0 {

		if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
			msg := tlog.Log() + "ERROR: You cannot have crn string and also crn attribute in the query. : query:[" + query + "]"
			//log.Println(msg)
			return nil, 0, errors.New(msg), http.StatusBadRequest
		}
		hasCRNQuery = true
	} else if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
		hasCRNQuery = true
	}

	if len(queryParms[INCIDENT_QUERY_CREATION_TIME_START]) > 0 || len(queryParms[INCIDENT_QUERY_CREATION_TIME_END]) > 0 ||
		len(queryParms[INCIDENT_QUERY_UPDATE_TIME_START]) > 0 || len(queryParms[INCIDENT_QUERY_UPDATE_TIME_END]) > 0 ||
		len(queryParms[INCIDENT_QUERY_PNP_CREATION_TIME_START]) > 0 || len(queryParms[INCIDENT_QUERY_PNP_CREATION_TIME_END]) > 0 ||
		len(queryParms[INCIDENT_QUERY_PNP_UPDATE_TIME_START]) > 0 || len(queryParms[INCIDENT_QUERY_PNP_UPDATE_TIME_END]) > 0 ||
		len(queryParms[INCIDENT_QUERY_OUTAGE_START_START]) > 0 || len(queryParms[INCIDENT_QUERY_OUTAGE_START_END]) > 0 ||
		len(queryParms[INCIDENT_COLUMN_STATE]) > 0 || len(queryParms[INCIDENT_COLUMN_CLASSIFICATION]) > 0 ||
		len(queryParms[INCIDENT_COLUMN_SEVERITY]) > 0 || len(queryParms[INCIDENT_COLUMN_PNP_REMOVED]) > 0 ||
		len(queryParms[INCIDENT_COLUMN_AUDIENCE]) > 0 {

		hasOtherQuery = true
	}

	if len(queryParms[INCIDENT_COLUMN_PNP_REMOVED]) > 0 {
		hasPnPRemovedQuery = true
	}

	whereClause := ""
	orderBy := ""
	var crnErr error

	if hasCRNQuery {
		whereClause, orderBy, crnErr = CreateCRNSearchWhereClauseFromQuery(queryParms, RESOURCE_TABLE_PREFIX)
		orderBy = INCIDENT_COLUMN_SOURCE + ", " + INCIDENT_COLUMN_SOURCE_ID // order by Source, SourceID
		if crnErr != nil {
			msg := "Error: cannot create where clause for CRN query"
			//log.Println(tlog.Log()+msg, err)
			return nil, 0, errors.New(msg), http.StatusBadRequest
		} else if strings.TrimSpace(whereClause) == "" {
			hasQuery = false
		} else {
			hasQuery = true
		}
	}

	if hasOtherQuery {
		// creation_time_start
		createStartQueryList, createStartErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_CREATION_TIME_START)
		if createStartErr != nil {
			//log.Println(tlog.Log() + createStartErr.Error())
			return nil, 0, createStartErr, http.StatusBadRequest
		}

		// creation_time_end
		createEndQueryList, createEndErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_CREATION_TIME_END)
		if createEndErr != nil {
			//log.Println(tlog.Log() + createEndErr.Error())
			return nil, 0, createEndErr, http.StatusBadRequest
		}

		// pnp_creation_time_start
		pnpCreateStartQueryList, pnpCreateStartErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_PNP_CREATION_TIME_START)
		if pnpCreateStartErr != nil {
			//log.Println(tlog.Log() + pnpCreateStartErr.Error())
			return nil, 0, pnpCreateStartErr, http.StatusBadRequest
		}

		// pnp_creation_time_end
		pnpCreateEndQueryList, pnpCreateEndErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_PNP_CREATION_TIME_END)
		if pnpCreateEndErr != nil {
			//log.Println(tlog.Log() + pnpCreateEndErr.Error())
			return nil, 0, pnpCreateEndErr, http.StatusBadRequest
		}

		// pnp_update_time_start
		pnpUpdateStartQueryList, pnpUpdateStartErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_PNP_UPDATE_TIME_START)
		if pnpUpdateStartErr != nil {
			//log.Println(tlog.Log() + pnpUpdateStartErr.Error())
			return nil, 0, pnpUpdateStartErr, http.StatusBadRequest
		}

		// pnp_update_time_end
		pnpUpdateEndQueryList, pnpUpdateEndErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_PNP_UPDATE_TIME_END)
		if pnpUpdateEndErr != nil {
			//log.Println(tlog.Log() + pnpUpdateEndErr.Error())
			return nil, 0, pnpUpdateEndErr, http.StatusBadRequest
		}

		// update_time_start
		updateStartQueryList, updateStartErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_UPDATE_TIME_START)
		if updateStartErr != nil {
			//log.Println(tlog.Log() + updateStartErr.Error())
			return nil, 0, updateStartErr, http.StatusBadRequest
		}

		// update_time_end
		updateEndQueryList, updateEndErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_UPDATE_TIME_END)
		if updateEndErr != nil {
			//log.Println(tlog.Log() + updateEndErr.Error())
			return nil, 0, updateEndErr, http.StatusBadRequest
		}

		// outage_start_start
		outageStartQueryList, outageStartErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_OUTAGE_START_START)
		if outageStartErr != nil {
			//log.Println(tlog.Log() + outageStartErr.Error())
			return nil, 0, outageStartErr, http.StatusBadRequest
		}

		// outage_start_end
		outageEndQueryList, outageEndErr := ParseQueryTimeCondition(queryParms, INCIDENT_QUERY_OUTAGE_START_END)
		if outageEndErr != nil {
			//log.Println(tlog.Log() + outageEndErr.Error())
			return nil, 0, outageEndErr, http.StatusBadRequest
		}

		stateQueryList := ParseQueryCondition(queryParms, INCIDENT_COLUMN_STATE, true)

		classificationQueryList := ParseQueryCondition(queryParms, INCIDENT_COLUMN_CLASSIFICATION, true)

		severityQueryList := ParseQueryCondition(queryParms, INCIDENT_COLUMN_SEVERITY, true)

		pnpRemovedQueryList := ParseQueryCondition(queryParms, INCIDENT_COLUMN_PNP_REMOVED, true)

		// =========== other query where clause ===========
		// where source_creation_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, createStartQueryList, INCIDENT_COLUMN_SOURCE_CREATION_TIME, INCIDENT_TABLE_PREFIX, ">=")

		// where source_creation_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, createEndQueryList, INCIDENT_COLUMN_SOURCE_CREATION_TIME, INCIDENT_TABLE_PREFIX, "<=")

		// where pnp_create_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpCreateStartQueryList, INCIDENT_COLUMN_PNP_CREATION_TIME, INCIDENT_TABLE_PREFIX, ">=")

		// where pnp_create_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpCreateEndQueryList, INCIDENT_COLUMN_PNP_CREATION_TIME, INCIDENT_TABLE_PREFIX, "<=")

		// where pnp_update_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpUpdateStartQueryList, INCIDENT_COLUMN_PNP_UPDATE_TIME, INCIDENT_TABLE_PREFIX, ">=")

		// where pnp_update_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, pnpUpdateEndQueryList, INCIDENT_COLUMN_PNP_UPDATE_TIME, INCIDENT_TABLE_PREFIX, "<=")

		// where source_update_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, updateStartQueryList, INCIDENT_COLUMN_SOURCE_UPDATE_TIME, INCIDENT_TABLE_PREFIX, ">=")

		// where source_update_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, updateEndQueryList, INCIDENT_COLUMN_SOURCE_UPDATE_TIME, INCIDENT_TABLE_PREFIX, "<=")

		// where start_time start
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, outageStartQueryList, INCIDENT_COLUMN_START_TIME, INCIDENT_TABLE_PREFIX, ">=")

		// where start_time end
		whereClause += CreateWhereWithSingleConditionWithOperators(&hasQuery, &orderBy, outageEndQueryList, INCIDENT_COLUMN_START_TIME, INCIDENT_TABLE_PREFIX, "<=")

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, stateQueryList, INCIDENT_COLUMN_STATE, INCIDENT_TABLE_PREFIX)

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, classificationQueryList, INCIDENT_COLUMN_CLASSIFICATION, INCIDENT_TABLE_PREFIX)

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, severityQueryList, INCIDENT_COLUMN_SEVERITY, INCIDENT_TABLE_PREFIX)

		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, pnpRemovedQueryList, INCIDENT_COLUMN_PNP_REMOVED, INCIDENT_TABLE_PREFIX)

		// filter out incidents created by heartbeat, where sourceID = INC0000001-<region>, like INC0000001-us-east, INC0000001-eu-gb, INC0000001-us-south
		whereClause += CreateWhereWithMultipleNotEqualConditionsExOrderBy(&hasQuery, []string{"INC0000001-us-east", "INC0000001-eu-gb", "INC0000001-us-south"}, INCIDENT_COLUMN_SOURCE_ID, INCIDENT_TABLE_PREFIX)

		orderBy = INCIDENT_COLUMN_SOURCE + ", " + INCIDENT_COLUMN_SOURCE_ID // order by Source, SourceID
	}

	// if pnpRemoved is not specified in the query, default to query false only
	if !hasPnPRemovedQuery {
		pnpRemovedQueryList := []string{"false"}
		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, pnpRemovedQueryList, INCIDENT_COLUMN_PNP_REMOVED, INCIDENT_TABLE_PREFIX)
		orderBy = INCIDENT_COLUMN_SOURCE + ", " + INCIDENT_COLUMN_SOURCE_ID // order by Source, SourceID
	}

	orderByClause := ""
	limitClause := ""
	offsetClause := ""
	selectStmt := ""
	totalCount := 0
	incidentsToReturn := []datastore.IncidentReturn{}

	if strings.TrimSpace(whereClause) == "" {
		hasQuery = false
	} else {
		hasQuery = true
		if orderBy != "" {
			orderByClause = " ORDER BY " + orderBy
		}

		// ======== LIMIT =========
		if limit > 0 {
			limitClause = " LIMIT " + strconv.Itoa(limit)
		}

		// ======== OFFSET =========
		if offset > 0 {
			offsetClause = " OFFSET " + strconv.Itoa(offset)
		}
	}

	// ========= database Query =========
	//log.Println("whereClause: ", whereClause)
	//log.Println("orderByClause: ", orderByClause)
	//log.Println("limitClause: ", limitClause)
	//log.Println("offsetClause: ", offsetClause)

	//	if hasCRNQuery {

	// With inner joins, only incidents that have at least one of its crn-masks found in the resource_table will be returned
	selectStmt = "SELECT DISTINCT " + SELECT_INCIDENT_COLUMNS + " FROM " + RESOURCE_TABLE_NAME + " " + RESOURCE_TABLE_ALIAS +
		" inner join " + INCIDENT_JUNCTION_TABLE_NAME + " " + JUNCTION_TABLE_ALIAS + " on " +
		RESOURCE_TABLE_PREFIX + RESOURCE_COLUMN_RECORD_ID + "=" + JUNCTION_TABLE_PREFIX + INCIDENTJUNCTION_COLUMN_RESOURCE_ID +
		" inner join " + INCIDENT_TABLE_NAME + " " + INCIDENT_TABLE_ALIAS + " on " +
		JUNCTION_TABLE_PREFIX + INCIDENTJUNCTION_COLUMN_INCIDENT_ID + "=" + INCIDENT_TABLE_PREFIX + INCIDENT_COLUMN_RECORD_ID +
		" WHERE " + whereClause

	//  https://github.ibm.com/cloud-sre/pnp-abstraction/issues/523
	//	} else if hasOtherQuery {
	//
	//		// do not have CRN query, only other time query
	//		selectStmt = "SELECT " + SELECT_INCIDENT_COLUMNS + " FROM " + INCIDENT_TABLE_NAME + " " + INCIDENT_TABLE_ALIAS +
	//			" WHERE " + whereClause
	//	}

	//log.Println("selectStmt: ", selectStmt)
	executeStmt := "WITH cte AS (" + selectStmt + ") SELECT * FROM ( TABLE cte " + orderByClause + limitClause + offsetClause + ") sub RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true;"

	log.Println(tlog.Log()+"DEBUG: executeStmt: ", executeStmt)

	// For example:
	// WITH cte AS (SELECT DISTINCT i.record_id,i.pnp_creation_time,i.pnp_update_time,i.source_creation_time,i.source_update_time,i.start_time,
	// i.end_time,i.short_description,i.long_description,i.state,i.classification,i.severity,i.source_id,i.source,i.regulatory_domain,i.crn_full,
	// i.affected_activity,i.customer_impact_description,i.pnp_removed FROM
	// resource_table r inner join incident_junction_table j on r.record_id=j.resource_id
	// inner join incident_table i on j.incident_id=i.record_id WHERE
	// (i.start_time>='2019-05-05T00:00:00Z') AND (i.start_time<='2019-06-04T00:00:00Z') AND
	// (i.source_id!='INC0000001-us-east' AND i.source_id!='INC0000001-eu-gb' AND i.source_id!='INC0000001-us-south') AND
	// (i.pnp_removed='false')) SELECT * FROM ( TABLE cte ORDER BY source, source_id LIMIT 200) sub RIGHT JOIN (SELECT count(*) FROM cte) c(total_count) ON true;

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(executeStmt)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Incidents found.")
			return nil, 0, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, 0, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			incidentsToReturn = []datastore.IncidentReturn{}
			for rows.Next() {
				var in datastore.IncidentGetNull
				err = rows.Scan(&in.RecordID, &in.PnpCreationTime, &in.PnpUpdateTime, &in.SourceCreationTime, &in.SourceUpdateTime, &in.OutageStartTime,
					&in.OutageEndTime, &in.ShortDescription, &in.LongDescription, &in.State, &in.Classification, &in.Severity, &in.SourceID, &in.Source,
					&in.RegulatoryDomain, &in.CRNFull, &in.AffectedActivity, &in.CustomerImpactDescription, &in.PnPRemoved, &in.TargetedURL, &in.Audience,
					&totalCount)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to IncidentGetNull Error: %v", err)
					return nil, 0, err, http.StatusInternalServerError
				}

				// if RecordID is not null, then rescan into MaintenanceGet struct
				if in.RecordID.Valid && totalCount > 0 {
					var ig datastore.IncidentGet
					err = rows.Scan(&ig.RecordID, &ig.PnpCreationTime, &ig.PnpUpdateTime, &ig.SourceCreationTime, &ig.SourceUpdateTime, &ig.OutageStartTime,
						&ig.OutageEndTime, &ig.ShortDescription, &ig.LongDescription, &ig.State, &ig.Classification, &ig.Severity, &ig.SourceID, &ig.Source,
						&ig.RegulatoryDomain, &ig.CRNFull, &ig.AffectedActivity, &ig.CustomerImpactDescription, &ig.PnPRemoved, &ig.TargetedURL, &ig.Audience,
						&totalCount)

					if err != nil {
						log.Printf(tlog.Log()+"Row Scan to IncidentGet Error: %v", err)
						return nil, 0, err, http.StatusInternalServerError
					}

					incidentReturn := ConvertIncidentGetToIncidentReturn(ig)
					incidentsToReturn = append(incidentsToReturn, incidentReturn)
				} //else {
				// the whole row has Null values, except total_count. Could be because offset is out of result set range
				//break
				//}
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, 0, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

		}

		if err != nil {
			return &incidentsToReturn, totalCount, err, http.StatusInternalServerError
		}
		retry = false
	}
	return &incidentsToReturn, totalCount, nil, http.StatusOK
}

// GetNotificationByQuery query is a string that has a format like "x=1&y=2&y=3", where the key like x,y,z, etc. must be one of the following values:
// crn, version, cname, ctype, service_name, location, scope, service_instance, resource_type, resource, type, creation_time_start, creation_time_end, source, source_id, type, pnp_removed
// When key is crn, it must be crn string format, and it can only support single value, cannot be comma separated values, for example,
// "crn=crn:version:cname:ctype:service-name:location:scope:service-instance:resource-type:resource"
// If a crn attribute is missing, wildcard is applied in the search.
// If you want an exact matching, not wildcard matching, then specify all crn attributes even if they are blanks,
// e.g. "version=v1&cname=cname1&ctype=ctype1&service-name=service-name1&location=location1&scope=scope1&service-instance=service-instance1&resource_type=&resource="
//
// For the following keys, like all the timestamp related keys, they can only be single value, they cannot be comma separated values:
// crn, source, source_id, creation_time_start, creation_time_end
//
// For the rest, their values can be comma separated. For example, "cname=abc,def&type=announcement,security"
//
// The rows returned are sorted by record_id.
// If the query string cannot be parsed by net/url/ParseQuery function, error will be returned.
// limit must be > 0, offset must be >= 0. If limit is <=0, will be no limit
//
// See GetWatchesByQuery function for examples.
func GetNotificationByQuery(database *sql.DB, query string, limit int, offset int) (*[]datastore.NotificationReturn, int, error, int) {
	log.Println(tlog.Log()+"query:", query, "limit:", limit, "offset:", offset)

	queryParms, err := url.ParseQuery(query)
	if err != nil {
		//log.Println(tlog.Log()+"Error: ", err)
		return nil, 0, err, http.StatusInternalServerError
	}

	hasQuery := false

	// checking query
	if len(queryParms[NOTIFICATION_QUERY_VERSION]) == 0 && len(queryParms[NOTIFICATION_QUERY_CNAME]) == 0 &&
		len(queryParms[NOTIFICATION_QUERY_CTYPE]) == 0 && len(queryParms[NOTIFICATION_QUERY_SERVICE_NAME]) == 0 &&
		len(queryParms[NOTIFICATION_QUERY_LOCATION]) == 0 && len(queryParms[NOTIFICATION_QUERY_SCOPE]) == 0 &&
		len(queryParms[NOTIFICATION_QUERY_SERVICE_INSTANCE]) == 0 && len(queryParms[NOTIFICATION_QUERY_RESOURCE_TYPE]) == 0 &&
		len(queryParms[NOTIFICATION_QUERY_RESOURCE]) == 0 && len(queryParms[NOTIFICATION_QUERY_CRN]) == 0 &&
		len(queryParms[NOTIFICATION_QUERY_CREATION_TIME_START]) == 0 && len(queryParms[NOTIFICATION_QUERY_CREATION_TIME_END]) == 0 &&
		len(queryParms[NOTIFICATION_QUERY_PNP_CREATION_TIME_START]) == 0 && len(queryParms[NOTIFICATION_QUERY_PNP_CREATION_TIME_END]) == 0 &&
		len(queryParms[NOTIFICATION_QUERY_PNP_UPDATE_TIME_START]) == 0 && len(queryParms[NOTIFICATION_QUERY_PNP_UPDATE_TIME_END]) == 0 &&
		len(queryParms[NOTIFICATION_COLUMN_SOURCE]) == 0 && len(queryParms[NOTIFICATION_COLUMN_SOURCE_ID]) == 0 &&
		len(queryParms[NOTIFICATION_COLUMN_TYPE]) == 0 && len(queryParms[NOTIFICATION_COLUMN_PNP_REMOVED]) == 0 {

		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters. : query:[" + query + "]"
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}

	// ======== check offset ========
	if offset < 0 {
		// handle error
		msg := tlog.Log() + "ERROR: offset must be a non-negative integer."
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}
	//log.Printf("queryParms: %+v", queryParms)
	if len(queryParms[NOTIFICATION_QUERY_VERSION]) > 0 || len(queryParms[NOTIFICATION_QUERY_CNAME]) > 0 ||
		len(queryParms[NOTIFICATION_QUERY_CTYPE]) > 0 || len(queryParms[NOTIFICATION_QUERY_SERVICE_NAME]) > 0 ||
		len(queryParms[NOTIFICATION_QUERY_LOCATION]) > 0 || len(queryParms[NOTIFICATION_QUERY_SCOPE]) > 0 ||
		len(queryParms[NOTIFICATION_QUERY_SERVICE_INSTANCE]) > 0 || len(queryParms[NOTIFICATION_QUERY_RESOURCE_TYPE]) > 0 ||
		len(queryParms[NOTIFICATION_QUERY_RESOURCE]) > 0 {

		if len(queryParms[NOTIFICATION_QUERY_CRN]) > 0 {
			msg := tlog.Log() + "ERROR: You cannot have crn string and also crn attribute in the query : query:[" + query + "]"
			//log.Println(msg)
			return nil, 0, errors.New(msg), http.StatusBadRequest
		}
	}

	whereClause := ""
	orderBy := ""
	var crnErr error

	whereClause, orderBy, crnErr = CreateCRNSearchWhereClauseFromQuery(queryParms, NOTIFICATION_TABLE_PREFIX)
	if crnErr != nil {
		msg := "Error: cannot create where clause for CRN query"
		//log.Println(tlog.Log()+msg, err)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	} else if strings.TrimSpace(whereClause) == "" {
		hasQuery = false
	} else {
		hasQuery = true
	}

	// creation_time_start
	createStartQueryList, createStartErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_CREATION_TIME_START)
	if createStartErr != nil {
		//log.Println(tlog.Log() + createStartErr.Error())
		return nil, 0, createStartErr, http.StatusBadRequest
	}

	// creation_time_end
	createEndQueryList, createEndErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_CREATION_TIME_END)
	if createEndErr != nil {
		//log.Println(tlog.Log() + createEndErr.Error())
		return nil, 0, createEndErr, http.StatusBadRequest
	}

	// update_time_start
	updateStartQueryList, updateStartErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_UPDATE_TIME_START)
	if updateStartErr != nil {
		//log.Println(tlog.Log() + updateStartErr.Error())
		return nil, 0, updateStartErr, http.StatusBadRequest
	}

	// update_time_end
	updateEndQueryList, updateEndErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_UPDATE_TIME_END)
	if updateEndErr != nil {
		//log.Println(tlog.Log() + updateEndErr.Error())
		return nil, 0, updateEndErr, http.StatusBadRequest
	}

	// pnp_creation_time_start
	pnpCreateStartQueryList, pnpCreateStartErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_PNP_CREATION_TIME_START)
	if pnpCreateStartErr != nil {
		//log.Println(tlog.Log() + pnpCreateStartErr.Error())
		return nil, 0, pnpCreateStartErr, http.StatusBadRequest
	}

	// pnp_creation_time_end
	pnpCreateEndQueryList, pnpCreateEndErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_PNP_CREATION_TIME_END)
	if pnpCreateEndErr != nil {
		//log.Println(tlog.Log() + pnpCreateEndErr.Error())
		return nil, 0, pnpCreateEndErr, http.StatusBadRequest
	}

	// pnp_update_time_start
	pnpUpdateStartQueryList, pnpUpdateStartErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_PNP_UPDATE_TIME_START)
	if pnpUpdateStartErr != nil {
		//log.Println(tlog.Log() + pnpUpdateStartErr.Error())
		return nil, 0, pnpUpdateStartErr, http.StatusBadRequest
	}

	// pnp_update_time_end
	pnpUpdateEndQueryList, pnpUpdateEndErr := ParseQueryTimeCondition(queryParms, NOTIFICATION_QUERY_PNP_UPDATE_TIME_END)
	if pnpUpdateEndErr != nil {
		//log.Println(tlog.Log() + pnpUpdateEndErr.Error())
		return nil, 0, pnpUpdateEndErr, http.StatusBadRequest
	}

	typeQueryList := ParseQueryCondition(queryParms, NOTIFICATION_COLUMN_TYPE, true)
	sourceQueryList := ParseQueryCondition(queryParms, NOTIFICATION_COLUMN_SOURCE, false)
	sourceIDQueryList := ParseQueryCondition(queryParms, NOTIFICATION_COLUMN_SOURCE_ID, false)
	pnpRemovedQueryList := ParseQueryCondition(queryParms, NOTIFICATION_COLUMN_PNP_REMOVED, true)

	// if pnpRemoved is not specified in the query, default to query false only
	if len(pnpRemovedQueryList) == 0 {
		pnpRemovedQueryList = append(pnpRemovedQueryList, "false")
	}

	// =========== other query where clause ===========
	// where source
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, sourceQueryList, NOTIFICATION_COLUMN_SOURCE, NOTIFICATION_TABLE_PREFIX, "=")

	// where source_id
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, sourceIDQueryList, NOTIFICATION_COLUMN_SOURCE_ID, NOTIFICATION_TABLE_PREFIX, "=")

	// where pnp_creation_time start
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, pnpCreateStartQueryList, NOTIFICATION_COLUMN_PNP_CREATION_TIME, NOTIFICATION_TABLE_PREFIX, ">=")

	// where pnp_creation_time end
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, pnpCreateEndQueryList, NOTIFICATION_COLUMN_PNP_CREATION_TIME, NOTIFICATION_TABLE_PREFIX, "<=")

	// where pnp_update_time start
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, pnpUpdateStartQueryList, NOTIFICATION_COLUMN_PNP_UPDATE_TIME, NOTIFICATION_TABLE_PREFIX, ">=")

	// where pnp_update_time end
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, pnpUpdateEndQueryList, NOTIFICATION_COLUMN_PNP_UPDATE_TIME, NOTIFICATION_TABLE_PREFIX, "<=")

	// where source_creation_time start
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, createStartQueryList, NOTIFICATION_COLUMN_SOURCE_CREATION_TIME, NOTIFICATION_TABLE_PREFIX, ">=")

	// where source_creation_time end
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, createEndQueryList, NOTIFICATION_COLUMN_SOURCE_CREATION_TIME, NOTIFICATION_TABLE_PREFIX, "<=")

	// where source_update_time start
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, updateStartQueryList, NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME, NOTIFICATION_TABLE_PREFIX, ">=")

	// where source_update_time end
	whereClause += CreateWhereWithSingleConditionWithOperatorsExOrderBy(&hasQuery, updateEndQueryList, NOTIFICATION_COLUMN_SOURCE_UPDATE_TIME, NOTIFICATION_TABLE_PREFIX, "<=")

	// where type
	whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, typeQueryList, NOTIFICATION_COLUMN_TYPE, NOTIFICATION_TABLE_PREFIX)

	// where pnp_removed
	whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, pnpRemovedQueryList, NOTIFICATION_COLUMN_PNP_REMOVED, NOTIFICATION_TABLE_PREFIX)

	// if pnpRemoved is not specified in the query, default to query false only
	if len(pnpRemovedQueryList) == 0 {
		log.Println("pnpRemovedQueryList len == 0")
		pnpRemovedQueryList := []string{"false"}
		whereClause += CreateWhereWithMultipleORConditionsExOrderBy(&hasQuery, pnpRemovedQueryList, NOTIFICATION_COLUMN_PNP_REMOVED, NOTIFICATION_TABLE_PREFIX)
	}

	orderBy = NOTIFICATION_COLUMN_RECORD_ID // order by RecordID

	orderByClause := ""
	limitClause := ""
	offsetClause := ""
	selectStmt := ""

	if strings.TrimSpace(whereClause) == "" {
		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters."
		//log.Println(msg)
		return nil, 0, errors.New(msg), http.StatusBadRequest
	}
	if orderBy != "" {
		orderByClause = " ORDER BY " + orderBy
	}

	// ======== LIMIT =========
	if limit > 0 {
		limitClause = " LIMIT " + strconv.Itoa(limit)
	}

	// ======== OFFSET =========
	if offset > 0 {
		offsetClause = " OFFSET " + strconv.Itoa(offset)
	}

	// ========= database Query =========
	//log.Println("whereClause: ", whereClause)
	//log.Println("orderByClause: ", orderByClause)
	//log.Println("limitClause: ", limitClause)
	//log.Println("offsetClause: ", offsetClause)

	// WITH cte AS (SELECT n.record_id FROM notification_table n WHERE (cname='cname1') )
	// SELECT DISTINCT * FROM (
	//                          SELECT DISTINCT n.record_id,n.pnp_creation_time,n.pnp_update_time,n.source_creation_time,n.type,n.category,n.source,n.source_id,n.crn_full,nd.long_description,nd.language
	//                          FROM notification_table n
	//                          full join notification_description_table nd on nd.notification_id=n.record_id
	//                          WHERE n.record_id = ANY
	//                                                   (SELECT * FROM cte ORDER BY record_id LIMIT 2 OFFSET 0)) AS n2
	//                          RIGHT JOIN (SELECT count(*) FROM cte) c(total_count) ON true ORDER BY record_id;

	// do not have CRN query, only other time query
	selectStmt = "SELECT " + NOTIFICATION_TABLE_PREFIX + NOTIFICATION_COLUMN_RECORD_ID + " FROM " + NOTIFICATION_TABLE_NAME + " " + NOTIFICATION_TABLE_ALIAS +
		" WHERE " + whereClause

	//log.Println("selectStmt: ", selectStmt)

	executeStmt := "WITH cte AS (" + selectStmt + ") " +
		"SELECT DISTINCT * FROM ( " + SELECT_NOTIFICATION_BY_RECORDID_STATEMENT + " ANY " +
		"(SELECT * FROM cte " + orderByClause + limitClause + offsetClause +
		") ) AS n2 RIGHT JOIN  (SELECT count(*) FROM cte) c(total_count) ON true ORDER BY record_id;"

	log.Println(tlog.Log()+"DEBUG: executeStmt: ", executeStmt)
	recordsReturn := []datastore.NotificationReturn{}
	totalCount := 0

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(executeStmt)
		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Notifications found.")
			return nil, 0, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, 0, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			//			totalCount, notificationsToReturn, err, status := getMultipleNotificationsToReturnFromRows(rows)

			recordsReturn = []datastore.NotificationReturn{}
			recordToReturn := datastore.NotificationReturn{}
			rg := datastore.NotificationGet{}
			totalCount = 0
			rowsReturned := 0
			currentRecordID := ""
			var notifDescriptionLongDescription, notifDescriptionLanguage sql.NullString
			var status int

			for rows.Next() {
				rowsReturned++
				r := datastore.NotificationGetNull{}

				err := rows.Scan(&r.RecordID, &r.PnpCreationTime, &r.PnpUpdateTime, &r.SourceCreationTime, &r.SourceUpdateTime,
					&r.EventTimeStart, &r.EventTimeEnd, &r.Source, &r.SourceID, &r.Type, &r.Category, &r.IncidentID,
					&r.CRNFull, &r.ResourceDisplayNames, &r.ShortDescription, &r.Tags, &r.RecordRetractionTime, &r.PnPRemoved, &r.ReleaseNoteUrl,
					&notifDescriptionLongDescription, &notifDescriptionLanguage, &totalCount)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to NotificationGetNull Error: %v", err)
					return nil, totalCount, err, http.StatusInternalServerError
				}

				if r.RecordID.Valid && currentRecordID != r.RecordID.String {
					if currentRecordID != "" {
						recordsReturn = append(recordsReturn, recordToReturn)
						recordToReturn = datastore.NotificationReturn{}
					}
					rg = datastore.NotificationGet{}
					err = rows.Scan(&rg.RecordID, &rg.PnpCreationTime, &rg.PnpUpdateTime, &rg.SourceCreationTime, &rg.SourceUpdateTime,
						&rg.EventTimeStart, &rg.EventTimeEnd, &rg.Source, &rg.SourceID, &rg.Type, &rg.Category, &rg.IncidentID,
						&rg.CRNFull, &rg.ResourceDisplayNames, &rg.ShortDescription, &rg.Tags, &rg.RecordRetractionTime, &rg.PnPRemoved, &rg.ReleaseNoteUrl,
						&notifDescriptionLongDescription, &notifDescriptionLanguage, &totalCount)

					if err != nil {
						log.Printf(tlog.Log()+"Row Scan to NotificationGet Error: %v", err)
						return nil, totalCount, err, http.StatusInternalServerError
					}

					recordToReturn = ConvertNotificationGetToNotificationReturn(rg)
					currentRecordID = recordToReturn.RecordID

				} else if !r.RecordID.Valid && totalCount > 0 && currentRecordID != "" {
					recordsReturn = append(recordsReturn, recordToReturn)
					return &recordsReturn, totalCount, nil, http.StatusOK

				} else if !r.RecordID.Valid {
					return &recordsReturn, totalCount, nil, http.StatusOK
				}

				if notifDescriptionLongDescription.Valid && notifDescriptionLanguage.Valid && !ContainDisplayName(recordToReturn.LongDescription, notifDescriptionLongDescription.String, notifDescriptionLanguage.String) {
					dn := datastore.DisplayName{Name: notifDescriptionLongDescription.String, Language: notifDescriptionLanguage.String}
					recordToReturn.LongDescription = append(recordToReturn.LongDescription, dn)

				} else if notifDescriptionLongDescription.Valid && !notifDescriptionLanguage.Valid && !ContainDisplayName(recordToReturn.LongDescription, notifDescriptionLongDescription.String, "") {
					dn := datastore.DisplayName{Name: notifDescriptionLongDescription.String, Language: ""}
					recordToReturn.LongDescription = append(recordToReturn.LongDescription, dn)

				} else if !notifDescriptionLongDescription.Valid && notifDescriptionLanguage.Valid && !ContainDisplayName(recordToReturn.LongDescription, "", notifDescriptionLanguage.String) {
					dn := datastore.DisplayName{Name: "", Language: notifDescriptionLanguage.String}
					recordToReturn.LongDescription = append(recordToReturn.LongDescription, dn)
				}
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return &recordsReturn, 0, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			if rowsReturned == 0 {
				err = sql.ErrNoRows
				status = http.StatusOK
			} else if currentRecordID == recordToReturn.RecordID {
				recordsReturn = append(recordsReturn, recordToReturn)
				status = http.StatusOK
			}

			if err != nil && status != http.StatusOK {
				log.Printf(tlog.Log()+"Row Scan to recordId, total_count Error: %v", err)
				return nil, 0, err, status
			}
			retry = false
		}
	}
	return &recordsReturn, totalCount, nil, http.StatusOK
}

// GetNotificationByRecordID gets a notification_table record from the database matching the notification_table.record_id
func GetNotificationByRecordID(database *sql.DB, recordID string) (*datastore.NotificationReturn, error, int) {
	log.Println(tlog.Log()+"recordID:", recordID)

	recordToReturn := datastore.NotificationReturn{}

	selectStmt := SELECT_NOTIFICATION_BY_RECORDID_STATEMENT + "'" + recordID + "'"

	//log.Println(tlog.Log()+"selectStmt: ", selectStmt)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query(selectStmt)
		rowsReturned := 0

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Notification found.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			recordToReturn = datastore.NotificationReturn{}
			for rows.Next() {
				rowsReturned++
				r := datastore.NotificationGet{}
				var notifDescriptionLongDescription, notifDescriptionLanguage sql.NullString

				err := rows.Scan(&r.RecordID, &r.PnpCreationTime, &r.PnpUpdateTime, &r.SourceCreationTime, &r.SourceUpdateTime,
					&r.EventTimeStart, &r.EventTimeEnd, &r.Source, &r.SourceID, &r.Type, &r.Category, &r.IncidentID,
					&r.CRNFull, &r.ResourceDisplayNames, &r.ShortDescription, &r.Tags, &r.RecordRetractionTime, &r.PnPRemoved, &r.ReleaseNoteUrl,
					&notifDescriptionLongDescription, &notifDescriptionLanguage)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
					return nil, err, http.StatusInternalServerError
				}

				if recordToReturn.RecordID == "" { // only do it once
					recordToReturn = ConvertNotificationGetToNotificationReturn(r)
				}

				if notifDescriptionLongDescription.Valid && notifDescriptionLanguage.Valid && !ContainDisplayName(recordToReturn.LongDescription, notifDescriptionLongDescription.String, notifDescriptionLanguage.String) {
					dn := datastore.DisplayName{Name: notifDescriptionLongDescription.String, Language: notifDescriptionLanguage.String}
					recordToReturn.LongDescription = append(recordToReturn.LongDescription, dn)

				} else if notifDescriptionLongDescription.Valid && !notifDescriptionLanguage.Valid && !ContainDisplayName(recordToReturn.LongDescription, notifDescriptionLongDescription.String, "") {
					dn := datastore.DisplayName{Name: notifDescriptionLongDescription.String, Language: ""}
					recordToReturn.LongDescription = append(recordToReturn.LongDescription, dn)

				} else if !notifDescriptionLongDescription.Valid && notifDescriptionLanguage.Valid && !ContainDisplayName(recordToReturn.LongDescription, "", notifDescriptionLanguage.String) {
					dn := datastore.DisplayName{Name: "", Language: notifDescriptionLanguage.String}
					recordToReturn.LongDescription = append(recordToReturn.LongDescription, dn)
				}
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}
			if rowsReturned == 0 {
				return nil, sql.ErrNoRows, http.StatusOK
			}

			retry = false
		}
	}
	return &recordToReturn, nil, http.StatusOK
}

// GetNotificationsByIncidentID get a notification_table record from the database matching the incident_record_id and pnp_removed = false
func GetNotificationsByIncidentID(database *sql.DB, incidentID string) (*[]datastore.NotificationReturn, error, int) {
	log.Println(tlog.Log()+"incidentID:", incidentID)

	recordsReturn := []datastore.NotificationReturn{}

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
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
			NOTIFICATION_COLUMN_CRN_FULL+","+
			NOTIFICATION_COLUMN_RESOURCE_DISPLAY_NAMES+","+
			NOTIFICATION_COLUMN_SHORT_DESCRIPTION+","+
			NOTIFICATION_COLUMN_TAGS+","+
			NOTIFICATION_COLUMN_RECORD_RETRACTION_TIME+","+
			NOTIFICATION_COLUMN_PNP_REMOVED+","+
			NOTIFICATION_COLUMN_RELEASE_NOTE_URL+" FROM "+
			NOTIFICATION_TABLE_NAME+" WHERE "+
			NOTIFICATION_COLUMN_INCIDENT_ID+" = $1 AND "+
			NOTIFICATION_COLUMN_PNP_REMOVED+" = $2",
			incidentID, "false")

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Notification found.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			recordsReturn = []datastore.NotificationReturn{}
			recordToReturn := datastore.NotificationReturn{}
			rg := datastore.NotificationGet{}
			rowsReturned := 0
			currentRecordID := ""
			//var notifDescriptionLongDescription, notifDescriptionLanguage sql.NullString
			var status int

			for rows.Next() {
				rowsReturned++
				r := datastore.NotificationGetNull{}

				err := rows.Scan(&r.RecordID, &r.PnpCreationTime, &r.PnpUpdateTime, &r.SourceCreationTime, &r.SourceUpdateTime,
					&r.EventTimeStart, &r.EventTimeEnd, &r.Source, &r.SourceID, &r.Type, &r.Category, &r.IncidentID,
					&r.CRNFull, &r.ResourceDisplayNames, &r.ShortDescription, &r.Tags, &r.RecordRetractionTime, &r.PnPRemoved, &r.ReleaseNoteUrl)

				if err != nil {
					log.Printf(tlog.Log()+"Row Scan to NotificationGetNull Error: %v", err)
					return nil, err, http.StatusInternalServerError
				}

				if r.RecordID.Valid && currentRecordID != r.RecordID.String {
					if currentRecordID != "" {
						recordsReturn = append(recordsReturn, recordToReturn)
						recordToReturn = datastore.NotificationReturn{}
					}
					rg = datastore.NotificationGet{}
					err = rows.Scan(&rg.RecordID, &rg.PnpCreationTime, &rg.PnpUpdateTime, &rg.SourceCreationTime, &rg.SourceUpdateTime,
						&rg.EventTimeStart, &rg.EventTimeEnd, &rg.Source, &rg.SourceID, &rg.Type, &rg.Category, &rg.IncidentID,
						&rg.CRNFull, &rg.ResourceDisplayNames, &rg.ShortDescription, &rg.Tags, &rg.RecordRetractionTime, &rg.PnPRemoved, &rg.ReleaseNoteUrl)

					if err != nil {
						log.Printf(tlog.Log()+"Row Scan to NotificationGet Error: %v", err)
						return nil, err, http.StatusInternalServerError
					}

					recordToReturn = ConvertNotificationGetToNotificationReturn(rg)
					currentRecordID = recordToReturn.RecordID

				} else if !r.RecordID.Valid && currentRecordID != "" {
					recordsReturn = append(recordsReturn, recordToReturn)
					return &recordsReturn, nil, http.StatusOK

				} else if !r.RecordID.Valid {
					return &recordsReturn, nil, http.StatusOK
				}
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return &recordsReturn, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			if rowsReturned == 0 {
				err = sql.ErrNoRows
				status = http.StatusOK
			} else if currentRecordID == recordToReturn.RecordID {
				recordsReturn = append(recordsReturn, recordToReturn)
				status = http.StatusOK
			}

			if err != nil && status != http.StatusOK {
				log.Printf(tlog.Log()+"Row Scan to recordId, total_count Error: %v", err)
				return nil, err, status
			}
			retry = false
		}
	}
	return &recordsReturn, nil, http.StatusOK
}

/*
func GetResourceIDByCRNFullStatement(database *sql.DB, crnFull string) (string, error) {

	FCT := "GetResourceIDByCRNFullStatement: "

	var resourceIDToReturn string

	err := database.QueryRow("SELECT "+
		RESOURCE_COLUMN_SOURCE_ID+" FROM "+
		RESOURCE_TABLE_NAME+" WHERE "+
		RESOURCE_COLUMN_CRN_FULL+" = $1",
		crnFull).Scan(&resourceIDToReturn)
	switch {
	case err == sql.ErrNoRows:
		log.Println(FCT + "No SourceID with CRNFull: " + crnFull)
	case err != nil:
		log.Println(FCT+"Error: ", err)
	default:
		log.Printf(FCT+"resourceIDToReturn is %s\n", resourceIDToReturn)
		return resourceIDToReturn, err
	}
	return resourceIDToReturn, err

}
*/
// getResourceByRecordIDStatement only get the resource record from resource_table, it does not get data
// from other tables like visibility_table, display_names_table, member_table.
func getResourceByRecordIDStatement(database *sql.DB, recordID string) (*datastore.ResourceReturn, error, int) {
	var resourceGet datastore.ResourceGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := database.QueryRow("SELECT "+
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
			RESOURCE_COLUMN_IS_CATALOG_PARENT+","+
			RESOURCE_COLUMN_CATALOG_PARENT_ID+","+
			RESOURCE_COLUMN_RECORD_HASH+" FROM "+
			RESOURCE_TABLE_NAME+" WHERE "+
			RESOURCE_COLUMN_RECORD_ID+" = $1",
			recordID).Scan(&resourceGet.RecordID,
			&resourceGet.PnpCreationTime,
			&resourceGet.PnpUpdateTime,
			&resourceGet.SourceCreationTime,
			&resourceGet.SourceUpdateTime,
			&resourceGet.CRNFull,
			&resourceGet.State,
			&resourceGet.OperationalStatus,
			&resourceGet.Source,
			&resourceGet.SourceID,
			&resourceGet.Status,
			&resourceGet.StatusUpdateTime,
			&resourceGet.RegulatoryDomain,
			&resourceGet.CategoryID,
			&resourceGet.CategoryParent,
			&resourceGet.IsCatalogParent,
			&resourceGet.CatalogParentID,
			&resourceGet.RecordHash)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Resource with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			//log.Printf("Resource is %s\n", resourceGet.RecordID)
			resourceToReturn := ConvertResourceGetToResourceReturn(resourceGet)
			return &resourceToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusInternalServerError

}

func getVisibilityByNameStatement(db *sql.DB, name string) (*datastore.VisibilityGet, error, int) {
	var visibilityToReturn datastore.VisibilityGet

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		err := db.QueryRow("SELECT "+
			VISIBILITY_COLUMN_RECORD_ID+","+
			VISIBILITY_COLUMN_NAME+","+
			VISIBILITY_COLUMN_DESCRIPTION+" FROM "+
			VISIBILITY_TABLE_NAME+" WHERE "+
			VISIBILITY_COLUMN_NAME+" = $1",
			name).Scan(&visibilityToReturn.RecordID,
			&visibilityToReturn.Name,
			&visibilityToReturn.Description)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Visibility with that ID.")
			return nil, err, http.StatusOK
		case err != nil:
			log.Println(tlog.Log()+"Error: ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			//log.Printf(tlog.Log()+"Visibility is %s\n", visibilityToReturn.RecordID)
			return &visibilityToReturn, err, http.StatusOK
		}
	}
	return nil, nil, http.StatusInternalServerError
}

func getVisibilityJunctionByResourceIDStatement(database *sql.DB, resourceRecordID string) ([]datastore.VisibilityJunctionGet, error) {
	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			VISIBILITYJUNCTION_COLUMN_RESOURCE_ID+","+
			VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID+" FROM "+
			VISIBILITY_JUNCTION_TABLE_NAME+" WHERE "+
			VISIBILITYJUNCTION_COLUMN_RESOURCE_ID+" = $1",
			resourceRecordID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Visibility Junction with that resource record ID found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err
			}

		default:
			defer rows.Close()

			visibilityJunctionToReturn := []datastore.VisibilityJunctionGet{}
			for rows.Next() {
				var r datastore.VisibilityJunctionGet
				err = rows.Scan(&r.ResourceID, &r.VisibilityID)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
				}
				//log.Printf(tlog.Log()+"VisibilityJunction resourceID is %s\n", r.ResourceID)
				visibilityJunctionToReturn = append(visibilityJunctionToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return visibilityJunctionToReturn, err
		}
		retry = false
	}
	return nil, nil
}

func getTagJunctionByResourceIDStatement(database *sql.DB, resourceRecordID string) ([]datastore.TagJunctionGet, error) {
	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			TAGJUNCTION_COLUMN_RESOURCE_ID+","+
			TAGJUNCTION_COLUMN_TAG_ID+" FROM "+
			TAG_JUNCTION_TABLE_NAME+" WHERE "+
			TAGJUNCTION_COLUMN_RESOURCE_ID+" = $1",
			resourceRecordID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Tag Junction with that resource record ID found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err
			}

		default:
			defer rows.Close()

			tagJunctionToReturn := []datastore.TagJunctionGet{}
			for rows.Next() {
				var r datastore.TagJunctionGet
				err = rows.Scan(&r.ResourceID, &r.TagID)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
				}
				//log.Printf(tlog.Log()+"TagJunction resourceID is %s\n", r.ResourceID)
				tagJunctionToReturn = append(tagJunctionToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return tagJunctionToReturn, err
		}
		retry = false
	}
	return nil, nil
}

func getVisibilityJunctionByVisibilityIDStatement(database *sql.DB, visibilityRecordID string) ([]datastore.VisibilityJunctionGet, error) {
	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			VISIBILITYJUNCTION_COLUMN_RESOURCE_ID+","+
			VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID+" FROM "+
			VISIBILITY_JUNCTION_TABLE_NAME+" WHERE "+
			VISIBILITYJUNCTION_COLUMN_VISIBILITY_ID+" = $1",
			visibilityRecordID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Visibility Junction with that visibility record ID found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err
			}

		default:
			defer rows.Close()

			visibilityJunctionToReturn := []datastore.VisibilityJunctionGet{}
			for rows.Next() {
				var r datastore.VisibilityJunctionGet
				err = rows.Scan(&r.ResourceID, &r.VisibilityID)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
				}
				visibilityJunctionToReturn = append(visibilityJunctionToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return visibilityJunctionToReturn, err
		}
		retry = false
	}
	return nil, nil
}

func getTagJunctionByTagIDStatement(database *sql.DB, tagRecordID string) ([]datastore.TagJunctionGet, error) {
	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			TAGJUNCTION_COLUMN_RESOURCE_ID+","+
			TAGJUNCTION_COLUMN_TAG_ID+" FROM "+
			TAG_JUNCTION_TABLE_NAME+" WHERE "+
			TAGJUNCTION_COLUMN_TAG_ID+" = $1",
			tagRecordID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Tag Junction with that tag record ID found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err
			}

		default:
			defer rows.Close()

			tagJunctionToReturn := []datastore.TagJunctionGet{}
			for rows.Next() {
				var r datastore.TagJunctionGet
				err = rows.Scan(&r.ResourceID, &r.TagID)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
				}
				tagJunctionToReturn = append(tagJunctionToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return tagJunctionToReturn, err
		}
		retry = false
	}
	return nil, nil
}

func getDisplayNamesByResourceIDStatement(database *sql.DB, resourceRecordID string) (*[]datastore.DisplayName, error) {
	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			DISPLAYNAMES_COLUMN_NAME+","+
			DISPLAYNAMES_COLUMN_LANGUAGE+" FROM "+
			DISPLAY_NAMES_TABLE_NAME+" WHERE "+
			DISPLAYNAMES_COLUMN_RESOURCE_ID+" = $1",
			resourceRecordID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Display Names with that resource record ID found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err
			}

		default:
			defer rows.Close()

			displayNamesToReturn := []datastore.DisplayName{}
			for rows.Next() {
				var r datastore.DisplayName
				err = rows.Scan(&r.Name, &r.Language)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
				}
				displayNamesToReturn = append(displayNamesToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return &displayNamesToReturn, err
		}
		retry = false
	}
	return nil, nil
}

//getNotificationDescriptionByNotificationIDStatement  returns the record of notification_description_table related with notification_table by notification_id
func getNotificationDescriptionByNotificationIDStatement(database *sql.DB, notificationRecordID string) (*[]datastore.NotificationDescriptionGet, error) {
	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			NOTIFICATIONDESCRIPTION_COLUMN_RECORD_ID+","+
			NOTIFICATIONDESCRIPTION_COLUMN_LONG_DESCRIPTION+","+
			NOTIFICATIONDESCRIPTION_COLUMN_LANGUAGE+" FROM "+
			NOTIFICATION_DESCRIPTION_TABLE_NAME+" WHERE "+
			NOTIFICATIONDESCRIPTION_COLUMN_NOTIFICATION_ID+" = $1",
			notificationRecordID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Notification Descriptions with that notification record ID found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err
			}

		default:
			defer rows.Close()

			ndToReturn := []datastore.NotificationDescriptionGet{}
			for rows.Next() {
				var r datastore.NotificationDescriptionGet
				err = rows.Scan(&r.RecordID, &r.LongDescription, &r.Language)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
				}
				ndToReturn = append(ndToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return &ndToReturn, err
		}
		retry = false
	}
	return nil, nil
}

//GetMaintenanceJunctionByMaintenanceID  returns the record of maintenance_juntion_table related with maintenance_table by record_id
func GetMaintenanceJunctionByMaintenanceID(database *sql.DB, maintenanceID string) (*[]datastore.MaintenanceJunctionGet, error, int) {
	log.Println(tlog.Log()+"maintenanceId:", maintenanceID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			MAINTENANCEJUNCTION_COLUMN_RECORD_ID+","+
			MAINTENANCEJUNCTION_COLUMN_RESOURCE_ID+","+
			MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID+" FROM "+
			MAINTENANCE_JUNCTION_TABLE_NAME+" WHERE "+
			MAINTENANCEJUNCTION_COLUMN_MAINTENANCE_ID+" = $1",
			maintenanceID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Maintenance with that maintenanceId found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}
		default:
			defer rows.Close()

			maintenanceToReturn := []datastore.MaintenanceJunctionGet{}
			for rows.Next() {
				var r datastore.MaintenanceJunctionGet
				err = rows.Scan(&r.RecordID, &r.ResourceID, &r.MaintenanceID)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
					return nil, err, http.StatusInternalServerError
				}
				maintenanceToReturn = append(maintenanceToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return &maintenanceToReturn, err, http.StatusOK
		}
		retry = false
	}
	return nil, nil, http.StatusOK
}

//GetIncidentJunctionByIncidentID  returns the record of incident_juntion_table related with incident_table by record_id
func GetIncidentJunctionByIncidentID(database *sql.DB, incidentID string) (*[]datastore.IncidentJunctionGet, error, int) {
	log.Println(tlog.Log()+"incidentId:", incidentID)

	// Retry transaction in case Postgres crashes, and wait for failover
	retry := true
	for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

		rows, err := database.Query("SELECT "+
			INCIDENTJUNCTION_COLUMN_RECORD_ID+","+
			INCIDENTJUNCTION_COLUMN_RESOURCE_ID+","+
			INCIDENTJUNCTION_COLUMN_INCIDENT_ID+" FROM "+
			INCIDENT_JUNCTION_TABLE_NAME+" WHERE "+
			INCIDENTJUNCTION_COLUMN_INCIDENT_ID+" = $1",
			incidentID)

		switch {
		case err == sql.ErrNoRows:
			log.Println(tlog.Log() + "No Incident with that incidentId found.")
		case err != nil:
			log.Println(tlog.Log()+"Error : ", err)
			if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
				Delay()
				continue
			} else {
				return nil, err, http.StatusInternalServerError
			}

		default:
			defer rows.Close()

			incidentToReturn := []datastore.IncidentJunctionGet{}
			for rows.Next() {
				var r datastore.IncidentJunctionGet
				err = rows.Scan(&r.RecordID, &r.ResourceID, &r.IncidentID)
				if err != nil {
					log.Printf(tlog.Log()+"Row Scan Error: %v", err)
					return &incidentToReturn, err, http.StatusInternalServerError
				}
				incidentToReturn = append(incidentToReturn, r)
			}
			if err = rows.Err(); err != nil {
				log.Printf(tlog.Log()+"rows.Err(): ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					// explicitly do rows close
					if err = rows.Close(); err != nil {
						log.Println(tlog.Log()+"Error in rows.Close ", err)
					}
					Delay()
					continue
				} else {
					return nil, err, http.StatusInternalServerError
				}
			}
			// explicitly do rows close
			if err = rows.Close(); err != nil {
				log.Println(tlog.Log()+"Error in rows.Close ", err)
			}

			return &incidentToReturn, err, http.StatusOK
		}
		retry = false
	}
	return nil, nil, http.StatusOK
}

// GetResourceByQuerySimple  Internal function to get Resource based on CRN. Does not support Visibility, DisplayName, Member.
//
// if record not found returns sql.ErrNoRows as error and http.StatusOK
func GetResourceByQuerySimple(database *sql.DB, query string) (*[]datastore.ResourceReturn, error, int) {
	log.Println(tlog.Log()+"query:", query)

	queryParms, err := url.ParseQuery(query)
	if err != nil {
		//log.Println(tlog.Log()+"Error: ", err)
		return nil, err, http.StatusBadRequest
	}

	// checking query
	if len(queryParms[RESOURCE_QUERY_VERSION]) == 0 && len(queryParms[RESOURCE_QUERY_CNAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CTYPE]) == 0 && len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_LOCATION]) == 0 && len(queryParms[RESOURCE_QUERY_SCOPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) == 0 && len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_RESOURCE]) == 0 && //len(queryParms[RESOURCE_QUERY_VISIBILITY]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CRN]) == 0 {

		msg := tlog.Log() + "ERROR: No recognized query values found, please set one or more query filters. : query:[" + query + "]"
		log.Println(tlog.Log(), msg)
		return nil, errors.New(msg), http.StatusBadRequest

	}

	hasCRNQuery := false
	//hasVisibilityQuery := false

	if len(queryParms[RESOURCE_QUERY_VERSION]) > 0 || len(queryParms[RESOURCE_QUERY_CNAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_CTYPE]) > 0 || len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) > 0 ||
		len(queryParms[RESOURCE_QUERY_LOCATION]) > 0 || len(queryParms[RESOURCE_QUERY_SCOPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) > 0 || len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) > 0 ||
		len(queryParms[RESOURCE_QUERY_RESOURCE]) > 0 {

		if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
			msg := tlog.Log() + "ERROR: You cannot have crn string and also crn attribute in the query. : query:[" + query + "]"
			log.Println(tlog.Log(), msg)
			return nil, errors.New(msg), http.StatusBadRequest
		}
		hasCRNQuery = true
	} else if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
		hasCRNQuery = true
	}

	// ======== Form where clause ========
	whereClause, orderBy, err := CreateCRNSearchWhereClauseFromQuery(queryParms, RESOURCE_TABLE_PREFIX)
	orderBy = RESOURCE_COLUMN_SOURCE + ", " + RESOURCE_COLUMN_SOURCE_ID // order by Source, SourceID

	// if there is query on visibility, we have to do further searching
	if strings.TrimSpace(whereClause) == "" {
		hasCRNQuery = false
	} else {
		hasCRNQuery = true
	}

	resourcesToReturn := []datastore.ResourceReturn{}

	if hasCRNQuery {

		// ======== ORDER BY =========
		orderByClause := ""
		if orderBy != "" {
			orderByClause = " ORDER BY " + orderBy
		}

		// ========= database Query =========
		//log.Println("whereClause: ", whereClause)
		//log.Println("orderByClause: ", orderByClause)

		selectStmt := "SELECT " + SELECT_RESOURCE_COLUMNS + SELECT_RESOURCE_FROM + whereClause + orderByClause

		// Retry transaction in case Postgres crashes, and wait for failover
		retry := true
		for nretry := 0; retry && nretry < DefaultMaxRetries; nretry++ {

			// Added the call below slq.Query does not return sql.ErrNoRows if not records will never hit the first case of the switch
			// if not rows don't need to call a bigger query that will consume more resources and not need to reach the for loop
			// retuned resourcesToReturn since it is expected nil returns an error as origal error handler
			// rowCnt := GetRecordsCount(database, SELECT_RESOURCE_FROM, whereClause)
			// if rowCnt == 0 {
			// 	log.Println(tlog.Log() + sql.ErrNoRows.Error())
			// 	//return &resourcesToReturn, sql.ErrNoRows, http.StatusOK
			// }
			// The above lines were created for debugging use, I forget to remove them
			log.Println(tlog.Log()+"select string: ", selectStmt)
			rows, err := database.Query(selectStmt)

			switch {
			case err == sql.ErrNoRows:
				log.Println(tlog.Log() + "No Resource found.")
				return nil, err, http.StatusOK
			case err != nil:
				log.Println(tlog.Log()+"Error : ", err)
				if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
					Delay()
					continue
				} else {
					return nil, err, http.StatusInternalServerError
				}

			default:
				defer rows.Close()

				resourcesToReturn = []datastore.ResourceReturn{}
				for rows.Next() {
					var r datastore.ResourceGet
					err = rows.Scan(&r.RecordID, &r.PnpCreationTime, &r.PnpUpdateTime, &r.SourceCreationTime,
						&r.SourceUpdateTime, &r.State, &r.OperationalStatus, &r.Source, &r.SourceID, &r.Status,
						&r.StatusUpdateTime, &r.RegulatoryDomain, &r.CategoryID, &r.CategoryParent, &r.CRNFull,
						&r.IsCatalogParent, &r.CatalogParentID, &r.RecordHash)
					if err != nil {
						log.Printf(tlog.Log()+"Row Scan Error: %v", err)
						return nil, err, http.StatusInternalServerError
					}
					log.Println(tlog.Log(), r.CRNFull, r.Status, r.SourceID, r.RecordID)
					resReturn := ConvertResourceGetToResourceReturn(r)
					resourcesToReturn = append(resourcesToReturn, resReturn)
				}
				if err = rows.Err(); err != nil {
					log.Printf(tlog.Log()+"rows.Err(): ", err)
					if nretry < DefaultMaxRetries-1 && !IsPostgresNonFailoverError(err) {
						// explicitly do rows close
						if err = rows.Close(); err != nil {
							log.Println(tlog.Log()+"Error in rows.Close ", err)
						}
						Delay()
						continue
					} else {
						log.Println(tlog.Log()+"Error: ", http.StatusInternalServerError)
						return nil, err, http.StatusInternalServerError
					}
				}
				// explicitly do rows close
				if err = rows.Close(); err != nil {
					log.Println(tlog.Log()+"Error in rows.Close ", err)
				}

				retry = false
			}
		}
		return &resourcesToReturn, nil, http.StatusOK
	}

	return &resourcesToReturn, nil, http.StatusOK
}

// ConvertIncidentGetToIncidentReturn recieves the incidentGet object and validate all data before to send the data back using IncidentReturn
func ConvertIncidentGetToIncidentReturn(incidentGet datastore.IncidentGet) datastore.IncidentReturn {

	incidentReturn := datastore.IncidentReturn{}
	incidentReturn.RecordID = incidentGet.RecordID
	incidentReturn.PnpCreationTime = incidentGet.PnpCreationTime
	incidentReturn.PnpUpdateTime = incidentGet.PnpUpdateTime
	incidentReturn.SourceCreationTime = incidentGet.SourceCreationTime
	//incidentReturn.Audience = incidentGet.Audience

	if incidentGet.SourceUpdateTime.Valid {
		incidentReturn.SourceUpdateTime = incidentGet.SourceUpdateTime.String
	}

	if incidentGet.OutageStartTime.Valid {
		incidentReturn.OutageStartTime = incidentGet.OutageStartTime.String
	}

	if incidentGet.OutageEndTime.Valid {
		incidentReturn.OutageEndTime = incidentGet.OutageEndTime.String
	}

	if incidentGet.ShortDescription.Valid {
		incidentReturn.ShortDescription = incidentGet.ShortDescription.String
	}

	if incidentGet.LongDescription.Valid {
		incidentReturn.LongDescription = incidentGet.LongDescription.String
	}

	incidentReturn.State = incidentGet.State
	incidentReturn.Classification = incidentGet.Classification
	incidentReturn.Severity = incidentGet.Severity

	incidentReturn.SourceID = incidentGet.SourceID
	incidentReturn.Source = incidentGet.Source

	if incidentGet.RegulatoryDomain.Valid {
		incidentReturn.RegulatoryDomain = incidentGet.RegulatoryDomain.String
	}

	if incidentGet.CRNFull.Valid {
		incidentReturn.CRNFull = SplitGetCRNStringToArray(incidentGet.CRNFull.String)
	}

	if incidentGet.AffectedActivity.Valid {
		incidentReturn.AffectedActivity = incidentGet.AffectedActivity.String
	}

	if incidentGet.CustomerImpactDescription.Valid {
		incidentReturn.CustomerImpactDescription = incidentGet.CustomerImpactDescription.String
	}

	if strings.ToLower(incidentGet.PnPRemoved) == "true" {
		incidentReturn.PnPRemoved = true
	} else {
		incidentReturn.PnPRemoved = false
	}

	if incidentGet.TargetedURL.Valid {
		incidentReturn.TargetedURL = incidentGet.TargetedURL.String
	}

	if incidentGet.Audience.Valid {
		incidentReturn.Audience = incidentGet.Audience.String
	}

	return incidentReturn
}

// ConvertMaintenanceGetToMaintenanceReturn recieves the MaintenanceGet object and validate all data before to send the data back using MaintenanceReturn
func ConvertMaintenanceGetToMaintenanceReturn(maintenanceGet datastore.MaintenanceGet) datastore.MaintenanceReturn {

	maintenanceReturn := datastore.MaintenanceReturn{}
	maintenanceReturn.RecordID = maintenanceGet.RecordID
	maintenanceReturn.PnpCreationTime = maintenanceGet.PnpCreationTime
	maintenanceReturn.PnpUpdateTime = maintenanceGet.PnpUpdateTime
	maintenanceReturn.SourceCreationTime = maintenanceGet.SourceCreationTime

	if maintenanceGet.SourceUpdateTime.Valid {
		maintenanceReturn.SourceUpdateTime = maintenanceGet.SourceUpdateTime.String
	}

	if maintenanceGet.PlannedStartTime.Valid {
		maintenanceReturn.PlannedStartTime = maintenanceGet.PlannedStartTime.String
	}

	if maintenanceGet.PlannedEndTime.Valid {
		maintenanceReturn.PlannedEndTime = maintenanceGet.PlannedEndTime.String
	}

	if maintenanceGet.ShortDescription.Valid {
		maintenanceReturn.ShortDescription = maintenanceGet.ShortDescription.String
	}

	if maintenanceGet.LongDescription.Valid {
		maintenanceReturn.LongDescription = maintenanceGet.LongDescription.String
	}

	maintenanceReturn.State = maintenanceGet.State

	if strings.ToLower(maintenanceGet.Disruptive) == "true" {
		maintenanceReturn.Disruptive = true
	} else {
		maintenanceReturn.Disruptive = false
	}
	maintenanceReturn.SourceID = maintenanceGet.SourceID
	maintenanceReturn.Source = maintenanceGet.Source

	if maintenanceGet.RegulatoryDomain.Valid {
		maintenanceReturn.RegulatoryDomain = maintenanceGet.RegulatoryDomain.String
	} else {
		maintenanceReturn.RegulatoryDomain = ""
	}

	temp, err := strconv.Atoi(maintenanceGet.MaintenanceDuration)
	if err != nil {
		maintenanceReturn.MaintenanceDuration = 0
	} else {
		maintenanceReturn.MaintenanceDuration = temp
	}

	temp, err = strconv.Atoi(maintenanceGet.DisruptionDuration)
	if err != nil {
		maintenanceReturn.DisruptionDuration = 0
	} else {
		maintenanceReturn.DisruptionDuration = temp
	}

	if maintenanceGet.DisruptionType.Valid {
		maintenanceReturn.DisruptionType = maintenanceGet.DisruptionType.String
	} else {
		maintenanceReturn.DisruptionType = ""
	}

	if maintenanceGet.DisruptionDescription.Valid {
		maintenanceReturn.DisruptionDescription = maintenanceGet.DisruptionDescription.String
	} else {
		maintenanceReturn.DisruptionDescription = ""
	}

	if maintenanceGet.CRNFull.Valid {
		maintenanceReturn.CRNFull = SplitGetCRNStringToArray(maintenanceGet.CRNFull.String)
	}

	if maintenanceGet.RecordHash.Valid {
		maintenanceReturn.RecordHash = maintenanceGet.RecordHash.String
	}

	if maintenanceGet.CompletionCode.Valid {
		maintenanceReturn.CompletionCode = maintenanceGet.CompletionCode.String
	}

	if strings.ToLower(maintenanceGet.PnPRemoved) == "true" {
		maintenanceReturn.PnPRemoved = true
	} else {
		maintenanceReturn.PnPRemoved = false
	}

	if maintenanceGet.TargetedURL.Valid {
		maintenanceReturn.TargetedURL = maintenanceGet.TargetedURL.String
	}

	if maintenanceGet.Audience.Valid {
		maintenanceReturn.Audience = maintenanceGet.Audience.String
	}

	log.Println(tlog.Log(), maintenanceReturn)

	return maintenanceReturn
}

// ConvertResourceGetToResourceReturn recieves the ResourceGet object and validate all data before to send the data back using ResourceReturn
func ConvertResourceGetToResourceReturn(resourceGet datastore.ResourceGet) datastore.ResourceReturn {

	resourceReturn := datastore.ResourceReturn{}
	resourceReturn.RecordID = resourceGet.RecordID
	resourceReturn.PnpCreationTime = resourceGet.PnpCreationTime
	resourceReturn.PnpUpdateTime = resourceGet.PnpUpdateTime

	if resourceGet.SourceCreationTime.Valid {
		resourceReturn.SourceCreationTime = resourceGet.SourceCreationTime.String
	}

	if resourceGet.SourceUpdateTime.Valid {
		resourceReturn.SourceUpdateTime = resourceGet.SourceUpdateTime.String
	}

	resourceReturn.CRNFull = resourceGet.CRNFull

	//resourceReturn.State = resourceGet.State
	if resourceGet.State.Valid {
		resourceReturn.State = resourceGet.State.String
	}

	//resourceReturn.OperationalStatus = resourceGet.OperationalStatus
	if resourceGet.OperationalStatus.Valid {
		resourceReturn.OperationalStatus = resourceGet.OperationalStatus.String
	}

	resourceReturn.Source = resourceGet.Source
	resourceReturn.SourceID = resourceGet.SourceID
	if resourceGet.RecordHash.Valid {
		resourceReturn.RecordHash = resourceGet.RecordHash.String
	}

	//resourceReturn.Status = resourceGet.Status
	if resourceGet.Status.Valid {
		resourceReturn.Status = resourceGet.Status.String
	}

	if resourceGet.StatusUpdateTime.Valid {
		resourceReturn.StatusUpdateTime = resourceGet.StatusUpdateTime.String
	}

	if resourceGet.RegulatoryDomain.Valid {
		resourceReturn.RegulatoryDomain = resourceGet.RegulatoryDomain.String
	}

	if resourceGet.CategoryID.Valid {
		resourceReturn.CategoryID = resourceGet.CategoryID.String
	}

	if strings.ToLower(resourceGet.CategoryParent) == "true" {
		resourceReturn.CategoryParent = true
	} else {
		resourceReturn.CategoryParent = false
	}

	resourceReturn.IsCatalogParent = resourceGet.IsCatalogParent

	if resourceGet.CatalogParentID.Valid {
		resourceReturn.CatalogParentID = resourceGet.CatalogParentID.String
	}

	if resourceGet.RecordHash.Valid {
		resourceReturn.RecordHash = resourceGet.RecordHash.String
	}

	return resourceReturn
}

// ConvertNotificationGetToNotificationReturn recieves the NotificationGet object and validate all data before to send the data back using NotificationReturn
func ConvertNotificationGetToNotificationReturn(notificationGet datastore.NotificationGet) datastore.NotificationReturn {

	notificationReturn := datastore.NotificationReturn{}
	notificationReturn.RecordID = notificationGet.RecordID
	notificationReturn.PnpCreationTime = notificationGet.PnpCreationTime
	notificationReturn.PnpUpdateTime = notificationGet.PnpUpdateTime

	if notificationGet.SourceCreationTime.Valid {
		notificationReturn.SourceCreationTime = notificationGet.SourceCreationTime.String
	}

	if notificationGet.SourceUpdateTime.Valid {
		notificationReturn.SourceUpdateTime = notificationGet.SourceUpdateTime.String
	}

	if notificationGet.EventTimeStart.Valid {
		notificationReturn.EventTimeStart = notificationGet.EventTimeStart.String
	}

	if notificationGet.EventTimeEnd.Valid {
		notificationReturn.EventTimeEnd = notificationGet.EventTimeEnd.String
	}

	notificationReturn.Source = notificationGet.Source
	notificationReturn.SourceID = notificationGet.SourceID
	notificationReturn.Type = notificationGet.Type
	notificationReturn.CRNFull = notificationGet.CRNFull
	notificationReturn.ReleaseNoteUrl = notificationGet.ReleaseNoteUrl.String

	if notificationGet.Category.Valid {
		notificationReturn.Category = notificationGet.Category.String
	}

	if notificationGet.IncidentID.Valid {
		notificationReturn.IncidentID = notificationGet.IncidentID.String
	}

	if notificationGet.ResourceDisplayNames.Valid {
		notificationReturn.ResourceDisplayNames = SplitGetNotificationDisplayNameToArray(notificationGet.ResourceDisplayNames.String)
	}

	if notificationGet.ShortDescription.Valid {
		notificationReturn.ShortDescription = SplitGetNotificationDisplayNameToArray(notificationGet.ShortDescription.String)
	}

	if notificationGet.Tags.Valid {
		notificationReturn.Tags = notificationGet.Tags.String
	}

	if notificationGet.RecordRetractionTime.Valid {
		notificationReturn.RecordRetractionTime = notificationGet.RecordRetractionTime.String
	}

	if strings.ToLower(notificationGet.PnPRemoved) == "true" {
		notificationReturn.PnPRemoved = true
	} else {
		notificationReturn.PnPRemoved = false
	}

	return notificationReturn
}

// ConvertPnpDeploymentToResourceInsert recieves the PnpDeployment object and validate and formats all data before to send the data back using ResourceInsert
func ConvertPnpDeploymentToResourceInsert(v *datastore.PnpDeployment, displayNames datastore.CatalogDisplayName) (resInsert *datastore.ResourceInsert) {

	resInsert = &(datastore.ResourceInsert{})
	resInsert.CategoryParent = v.Parent
	resInsert.OperationalStatus = v.OperationalStatus
	resInsert.State = v.State
	resInsert.Status = v.Status
	resInsert.Visibility = v.Visibility
	resInsert.Source = v.Source
	resInsert.SourceID = v.SourceID
	resInsert.SourceCreationTime = v.CreationTime
	resInsert.SourceUpdateTime = v.UpdateTime
	resInsert.CRNFull = v.Crn
	resInsert.CategoryID = v.CategoryId
	resInsert.OperationalStatus = v.OperationalStatus
	for _, val := range v.Tags {
		resInsert.Tags = append(resInsert.Tags, datastore.Tag{ID: val})
	}

	for i := range displayNames {
		displayName := datastore.DisplayName{Name: v.DisplayName[i].Text, Language: v.DisplayName[i].Language}
		resInsert.DisplayNames = append(resInsert.DisplayNames, displayName)
	}
	resInsert.IsCatalogParent = v.IsCatalogParent
	resInsert.CatalogParentID = v.CatalogParentID
	return resInsert
}

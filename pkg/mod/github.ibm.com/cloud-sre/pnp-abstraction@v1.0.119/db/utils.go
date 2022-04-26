package db

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/oss-globals/tlog"
	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-abstraction/utils"
)

// CRN Cloud Resource Name structure
type CRN struct {
	Version         string `json:"version"`
	Cname           string `json:"cname"`
	Ctype           string `json:"ctype"`
	ServiceName     string `json:"serviceName"`
	Location        string `json:"location"`
	Scope           string `json:"scope"`
	ServiceInstance string `json:"serviceInstance"`
	ResourceType    string `json:"resourceType"`
	Resource        string `json:"resource"`
}

// CreateRecordIDFromSourceSourceID creates a checksum string using source and sourceID values
// to generate unique record_id
func CreateRecordIDFromSourceSourceID(source string, sourceID string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s+%s", source, sourceID)))
	return fmt.Sprintf("%x", sum)
}

// CreateRecordIDFromString creates a checksum string using name value to generate a unique record_id
func CreateRecordIDFromString(name string) string {
	sum := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%x", sum)
}

// CreateNotificationRecordID Provides a unique id for SN change notifications that is different than all others
func CreateNotificationRecordID(source string, sourceID string, crn string, incidentID string, kind string) string {
	if source == "servicenow" && kind == "maintenance" {
		return CreateRecordIDFromString(source + sourceID + crn + incidentID)
	}
	return CreateRecordIDFromString(source + sourceID + crn)
}

// ParseCRNString parse CRN string which is in format crn:version:cname:ctype:service-name:location:scope:service-instance:resource-type:resource
// DEPRECATED Use crn.Parse(string) or crn.ParseAll(string) from github.ibm.com/cloud-sre/osscatalog/crn instead
func ParseCRNString(CRNFull string) (*CRN, bool) {
	// ensure CRNFull is in lowercase
	CRNFull = strings.ToLower(CRNFull)

	if !strings.HasPrefix(CRNFull, "crn:") {
		return nil, false
	}
	crnStr := CRNFull[4:] // remove "crn:" prefix
	var crnStruct CRN
	crnArray := strings.Split(crnStr, ":")
	if len(crnArray) != 9 {
		return nil, false
	}
	crnStruct.Version = crnArray[0]
	crnStruct.Cname = crnArray[1]
	crnStruct.Ctype = crnArray[2]
	crnStruct.ServiceName = crnArray[3]
	crnStruct.Location = crnArray[4]
	crnStruct.Scope = crnArray[5]
	crnStruct.ServiceInstance = crnArray[6]
	crnStruct.ResourceType = crnArray[7]
	crnStruct.Resource = crnArray[8]

	return &crnStruct, true
}

//CRNStringToQueryParms  To be used by GetResourceByQuery to convert CRN string to queryParms for wildcard matching
func CRNStringToQueryParms(CRNFull string) (queryParms url.Values, ok bool) {
	queryParms = make(map[string][]string)

	// ensure CRNFull is in lowercase
	CRNFull = strings.ToLower(CRNFull)

	if !strings.HasPrefix(CRNFull, "crn:") {
		return nil, false
	}
	crnStr := CRNFull[4:] // remove "crn:" prefix
	crnArray := strings.Split(crnStr, ":")
	if len(crnArray) != 9 {
		return nil, false
	}

	if strings.TrimSpace(crnArray[0]) != "" {
		queryParms[RESOURCE_QUERY_VERSION] = []string{crnArray[0]}
	}
	if strings.TrimSpace(crnArray[1]) != "" {
		queryParms[RESOURCE_QUERY_CNAME] = []string{crnArray[1]}
	}
	if strings.TrimSpace(crnArray[2]) != "" {
		queryParms[RESOURCE_QUERY_CTYPE] = []string{crnArray[2]}
	}
	if strings.TrimSpace(crnArray[3]) != "" {
		queryParms[RESOURCE_QUERY_SERVICE_NAME] = []string{crnArray[3]}
	}
	if strings.TrimSpace(crnArray[4]) != "" {
		queryParms[RESOURCE_QUERY_LOCATION] = []string{crnArray[4]}
	}
	if strings.TrimSpace(crnArray[5]) != "" {
		queryParms[RESOURCE_QUERY_SCOPE] = []string{crnArray[5]}
	}
	if strings.TrimSpace(crnArray[6]) != "" {
		queryParms[RESOURCE_QUERY_SERVICE_INSTANCE] = []string{crnArray[6]}
	}
	if strings.TrimSpace(crnArray[7]) != "" {
		queryParms[RESOURCE_QUERY_RESOURCE_TYPE] = []string{crnArray[7]}
	}
	if strings.TrimSpace(crnArray[8]) != "" {
		queryParms[RESOURCE_QUERY_RESOURCE] = []string{crnArray[8]}
	}

	return queryParms, true
}

// VerifyAndConvertTimestamp Verify that the timestamp specified in create_start and create_end are one of these format:
//    yyyy-MM-ddTHH:mm:ssZ
//    yyyy-MM-ddTHH:mm:ss.sssZ
//    yyyy-MM-ddTHH:mm:ss-hhmm
//    yyyy-MM-ddTHH:mm:ss+hhmm
//
// And convert yyyy-MM-ddTHH:mm:ss-hhmm or yyyy-MM-ddTHH:mm:ss+hhmm to yyyy-MM-ddTHH:mm:ssZ UTC timestamp
// When yyyy-MM-ddTHH:mm:ss-hhmm, add hhmm to yyyy-MM-ddTHH:mm:ss
// When yyyy-MM-ddTHH:mm:ss-hhmm, minus hhmm from yyyy-MM-ddTHH:mm:ss
func VerifyAndConvertTimestamp(timestamp string) (utcTimestamp string, err error) {
	return utils.VerifyAndConvertTimestamp(timestamp)
}

// ParseQueryCondition - if tolower is true, will convert the queryParams[paramKey] values to lowercase,
// this is especially for crn and crn attributes
func ParseQueryCondition(queryParams url.Values, paramKey string, tolower bool) (paramValueList []string) {

	if tolower {
		queryParams[paramKey] = StringArrayToLower(queryParams[paramKey])
	}

	if len(queryParams[paramKey]) == 1 {
		// Parse the source IDs if it is a comma separated list
		paramValueList = strings.Split(queryParams[paramKey][0], ",")
	} else if len(queryParams[paramKey]) > 1 {
		// If they by chance used separate parameters, then just copy all of those
		paramValueList = append(paramValueList, queryParams[paramKey]...)
	}
	return paramValueList
}

// ParseQueryTimeCondition For parsing query condition that is in a range of time. This function especially used by time condition,
// to ensure there is only one value specified for time condition, and the timestamp value is in correct format.
// For example, if query string is
// "creation_time_start=2018-07-30T15:04:00Z&creation_end_start=2018-07-30T15:04:30Z"
// then a single condition is "creation_time_start=2018-07-30T15:04:00Z" and "creation_end_start=2018-07-30T15:04:30Z".
// We cannot have multiple value for "creation_time_start" and "creation_time_end".
func ParseQueryTimeCondition(queryParams url.Values, paramKey string) (paramValueList []string, err error) {
	paramValueList = ParseQueryCondition(queryParams, paramKey, false)
	if len(paramValueList) > 1 {
		// multiple values found
		msg := "ERROR: multiple " + paramKey + " found in the query. "
		err = errors.New(msg)
	} else if len(paramValueList) == 1 {
		paramValueList[0], err = VerifyAndConvertTimestamp(paramValueList[0])
		if err != nil {
			msg := "ERROR: " + paramKey + " has incorrect timestamp format. Error message:" + err.Error()
			err = errors.New(msg)
		}
	}
	return paramValueList, err
}

// CreateWhereWithMultipleORConditions The return string will be something like:
// " (originator='originator 1' OR originator='originator 2')" if hasQuery is false
// " AND (originator='originator 1' OR originator='originator 2')" if hasQuery is true
func CreateWhereWithMultipleORConditions(hasQuery *bool, orderBy *string, condValueList []string, conditionKey string, tablePrefix string) (whereClauseCondition string) {
	whereClauseCondition = ""

	if len(condValueList) > 0 {
		if *hasQuery {
			whereClauseCondition += " AND ("
		} else {
			whereClauseCondition += " ("
			*orderBy = conditionKey
		}
		whereClauseCondition += tablePrefix + conditionKey + "='" + condValueList[0] + "'"
		*hasQuery = true
		for i := 1; i < len(condValueList); i++ {
			whereClauseCondition += " OR " + tablePrefix + conditionKey + "='" + condValueList[i] + "'"
		}
		whereClauseCondition += ")"
	}

	return whereClauseCondition
}

// CreateWhereWithMultipleORConditionsExOrderBy Similar to CreateWhereWithMultipleORConditions but does not include Order by
func CreateWhereWithMultipleORConditionsExOrderBy(hasQuery *bool, condValueList []string, conditionKey string, tablePrefix string) (whereClauseCondition string) {
	whereClauseCondition = ""

	if len(condValueList) > 0 {
		if *hasQuery {
			whereClauseCondition += " AND ("
		} else {
			whereClauseCondition += " ("
		}
		whereClauseCondition += tablePrefix + conditionKey + "='" + condValueList[0] + "'"
		*hasQuery = true
		for i := 1; i < len(condValueList); i++ {
			whereClauseCondition += " OR " + tablePrefix + conditionKey + "='" + condValueList[i] + "'"
		}
		whereClauseCondition += ")"
	}

	return whereClauseCondition
}

// CreateWhereWithMultipleNotEqualConditionsExOrderBy similar to CreateWhereWithMultipleORConditionsExOrderBy
// but this one uses not equal instead
func CreateWhereWithMultipleNotEqualConditionsExOrderBy(hasQuery *bool, condValueList []string, conditionKey string, tablePrefix string) (whereClauseCondition string) {
	whereClauseCondition = ""

	if len(condValueList) > 0 {
		if *hasQuery {
			whereClauseCondition += " AND ("
		} else {
			whereClauseCondition += " ("
		}
		whereClauseCondition += tablePrefix + conditionKey + "!='" + condValueList[0] + "'"
		*hasQuery = true
		for i := 1; i < len(condValueList); i++ {
			whereClauseCondition += " AND " + tablePrefix + conditionKey + "!='" + condValueList[i] + "'"
		}
		whereClauseCondition += ")"
	}

	return whereClauseCondition
}

// CreateWhereWithSingleConditionWithOperators The return string will be something like:
// "(source_creation_time>='2018-07-30T15:04:00Z')" if hasQuery is false
// " AND (source_creation_time<='2018-07-30T15:04:30Z')" if hasQuery is true
func CreateWhereWithSingleConditionWithOperators(hasQuery *bool, orderBy *string, condValueList []string, conditionKey string, tablePrefix string, operator string) (whereClauseCondition string) {
	whereClauseCondition = ""

	if len(condValueList) > 0 {
		if *hasQuery {
			whereClauseCondition += " AND ("
		} else {
			whereClauseCondition += " ("
			*orderBy = conditionKey
		}
		whereClauseCondition += tablePrefix + conditionKey + operator + "'" + condValueList[0] + "')"
		*hasQuery = true
	}

	return whereClauseCondition
}

// CreateWhereWithSingleConditionWithOperatorsExOrderBy similar to CreateWhereWithSingleConditionWithOperators but does not include Order by
func CreateWhereWithSingleConditionWithOperatorsExOrderBy(hasQuery *bool, condValueList []string, conditionKey string, tablePrefix string, operator string) (whereClauseCondition string) {
	whereClauseCondition = ""

	if len(condValueList) > 0 {
		if *hasQuery {
			whereClauseCondition += " AND ("
		} else {
			whereClauseCondition += " ("
		}
		whereClauseCondition += tablePrefix + conditionKey + operator + "'" + condValueList[0] + "')"
		*hasQuery = true
	}

	return whereClauseCondition
}

// IntersectionResourceAndVisibilityJunction Find the matching of resource sourceID and VisibilityJunctionGet.ResourceID
// Return the matched ResourceReturn records.
func IntersectionResourceAndVisibilityJunction(resourceSourceIds []string, visibilityJunctions []datastore.VisibilityJunctionGet) (c []string) {
	m := make(map[string]int)

	for i, resource := range resourceSourceIds {
		m[resource] = i + 1
	}

	idx := -1
	for _, vj := range visibilityJunctions {
		idx = m[vj.ResourceID]
		if idx > 0 {
			c = append(c, resourceSourceIds[idx-1])
		}
	}
	return c
}

// CreateCRNSearchWhereClauseFromQuery Create a Where statement if GaaS removes CRN
func CreateCRNSearchWhereClauseFromQuery(queryParms url.Values, tablePrefix string) (whereClause string, orderBy string, err error) {
	hasQuery := false

	if len(queryParms[RESOURCE_QUERY_CRN]) > 0 {
		//crnStruct, _ := ParseCRNString(queryParms[RESOURCE_QUERY_CRN][0]) Using ParseAll from osscatalog instead
		crnStruct, _ := crn.ParseAll(queryParms[RESOURCE_QUERY_CRN][0])
		isGaasResource := api.IsCrnGaasResource(queryParms[RESOURCE_QUERY_CRN][0], ctxt.Context{})

		log.Println(tlog.Log()+"IsGaasResource: ", isGaasResource, ", crn: ", queryParms[RESOURCE_QUERY_CRN][0])
		if isGaasResource {
			// is Gaas resource query, delete "crn" query, and add crn attributes query for non-wildcard search to
			// avoid picking up IBM Public Cloud (bluemix:public) resources which has the same service-name
			delete(queryParms, RESOURCE_QUERY_CRN)
			queryParms[RESOURCE_QUERY_VERSION] = []string{crnStruct.Version}
			queryParms[RESOURCE_QUERY_CNAME] = []string{""}
			queryParms[RESOURCE_QUERY_CTYPE] = []string{""}
			queryParms[RESOURCE_QUERY_SERVICE_NAME] = []string{crnStruct.ServiceName}
			queryParms[RESOURCE_QUERY_LOCATION] = []string{""}
			queryParms[RESOURCE_QUERY_SCOPE] = []string{""}
			queryParms[RESOURCE_QUERY_SERVICE_INSTANCE] = []string{""}
			queryParms[RESOURCE_QUERY_RESOURCE_TYPE] = []string{""}
			queryParms[RESOURCE_QUERY_RESOURCE] = []string{""}
		} else {
			// is IBM Public Cloud resource query
			queryParms2, ok := CRNStringToQueryParms(queryParms[RESOURCE_QUERY_CRN][0])
			if !ok {
				msg := tlog.Log() + "Error: crn string in query is invalid."
				log.Println(msg)
				return "", "", errors.New(msg)
			}
			if len(queryParms[RESOURCE_QUERY_VISIBILITY]) > 0 {
				queryParms2[RESOURCE_QUERY_VISIBILITY] = queryParms[RESOURCE_QUERY_VISIBILITY]
			}
			queryParms = queryParms2
			//log.Println(FCT+"queryParms: ", queryParms)
		}
	}

	// checking query
	if len(queryParms[RESOURCE_QUERY_VERSION]) == 0 && len(queryParms[RESOURCE_QUERY_CNAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_CTYPE]) == 0 && len(queryParms[RESOURCE_QUERY_SERVICE_NAME]) == 0 &&
		len(queryParms[RESOURCE_QUERY_LOCATION]) == 0 && len(queryParms[RESOURCE_QUERY_SCOPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_SERVICE_INSTANCE]) == 0 && len(queryParms[RESOURCE_QUERY_RESOURCE_TYPE]) == 0 &&
		len(queryParms[RESOURCE_QUERY_RESOURCE]) == 0 && len(queryParms[RESOURCE_QUERY_CATALOG_PARENT_ID]) == 0 {

		//msg := FCT + "ERROR: No recognized query values found, please set one or more query filters."
		//log.Println(msg)
		return "", "", nil
	}

	// ======= Parse query =======
	// remember it may be only one value or a comma separated list of source IDs
	// version
	versionQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_VERSION, true)

	// cname
	cnameQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_CNAME, true)

	// ctype
	ctypeQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_CTYPE, true)

	// service_name
	serviceNameQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_SERVICE_NAME, true)

	// location
	locationQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_LOCATION, true)

	// scope
	scopeQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_SCOPE, true)

	// service_instance
	serviceInstanceQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_SERVICE_INSTANCE, true)

	// resource_type
	resourceTypeQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_RESOURCE_TYPE, true)

	// resource
	resourceQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_RESOURCE, true)

	// catalog parent id
	catalogParentIDQueryList := ParseQueryCondition(queryParms, RESOURCE_QUERY_CATALOG_PARENT_ID, true)

	if len(versionQueryList) == 0 && len(cnameQueryList) == 0 && len(ctypeQueryList) == 0 &&
		len(serviceNameQueryList) == 0 && len(locationQueryList) == 0 && len(scopeQueryList) == 0 &&
		len(serviceInstanceQueryList) == 0 && len(resourceTypeQueryList) == 0 &&
		len(resourceQueryList) == 0 && len(catalogParentIDQueryList) == 0 {

		msg := tlog.Log() + "ERROR: either one or more filters are invalid, or no query filters are found for crn string or crn attributes."
		log.Println(msg)
		return "", "", errors.New(msg)
	}

	// ======== Form where clause ========

	// where cname
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, cnameQueryList, RESOURCE_COLUMN_CNAME, tablePrefix)

	// where ctype
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, ctypeQueryList, RESOURCE_COLUMN_CTYPE, tablePrefix)

	// where service_name
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, serviceNameQueryList, RESOURCE_COLUMN_SERVICE_NAME, tablePrefix)

	// where location
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, locationQueryList, RESOURCE_COLUMN_LOCATION, tablePrefix)

	// where scope
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, scopeQueryList, RESOURCE_COLUMN_SCOPE, tablePrefix)

	// where service_instance
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, serviceInstanceQueryList, RESOURCE_COLUMN_SERVICE_INSTANCE, tablePrefix)

	// where resource_type
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, resourceTypeQueryList, RESOURCE_COLUMN_RESOURCE_TYPE, tablePrefix)

	// where resource
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, resourceQueryList, RESOURCE_COLUMN_RESOURCE, tablePrefix)

	// where resource
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, catalogParentIDQueryList, RESOURCE_COLUMN_CATALOG_PARENT_ID, tablePrefix)

	// where version (place version at the end, for we do not want to search by version, or orderBy version)
	whereClause += CreateWhereWithMultipleORConditions(&hasQuery, &orderBy, versionQueryList, RESOURCE_COLUMN_VERSION, tablePrefix)

	return whereClause, orderBy, nil
}

// SplitGetCRNStringToArray when we get CRNFull from database, it is returned as a string, even though it is an array of string, e.g. the data returned is something like
// "{crn:1:cname1:ctype1:service_name1:location1::::,crn:1:cname1:ctype1:service_name1:location2::::}", then this function will split it into an array
// array[0]="crn:1:cname1:ctype1:service_name1:location1::::",
// array[1]="crn:1:cname1:ctype1:service_name1:location2::::"
func SplitGetCRNStringToArray(getCRNFullStr string) []string {

	if strings.HasPrefix(getCRNFullStr, "{") {
		getCRNFullStr = getCRNFullStr[1:]
	}

	if strings.HasSuffix(getCRNFullStr, "}") {
		getCRNFullStr = getCRNFullStr[0 : len(getCRNFullStr)-1]
	}

	return strings.Split(getCRNFullStr, ",")
}

// SplitGetNotificationDisplayNameToArray when we get "DisplayName" type of items from database, it is returned as a string, even though it is an array of string, e.g. the data returned is something like
// '{"en                  resource display name1","fr                  resource display name2"}', then this function will split it into an array of DisplayName struct
// array[0]={name: resource display name1, language: en}
// array[1]={name: resource display name2, language: fr}
func SplitGetNotificationDisplayNameToArray(getStr string) (displayNames []datastore.DisplayName) {
	if getStr == "{}" {
		return displayNames
	}

	if strings.HasPrefix(getStr, "{") {
		getStr = getStr[1:]
	}

	if strings.HasSuffix(getStr, "}") {
		getStr = getStr[0 : len(getStr)-1]
	}

	strArr := strings.Split(getStr, "\",\"")

	for _, str := range strArr {
		str = strings.Trim(str, "\"")
		dn := datastore.DisplayName{}
		dn.Language = strings.TrimSpace(str[0:NOTIFICATION_LANGUAGE_LENGTH])
		dn.Name = str[NOTIFICATION_LANGUAGE_LENGTH:]
		displayNames = append(displayNames, dn)
	}

	return displayNames
}

// CreateNonWildcardCRNQueryString Create a query string for non-wildcard search from a CRNFull string
func CreateNonWildcardCRNQueryString(crnFullStr string) (string, error) {
	// ensure crnFullStr is in lowercase
	crnFullStr = strings.ToLower(crnFullStr)

	//crn, b := ParseCRNString(crnFullStr) using ParseAll from osscatalog instead
	crn, err := crn.ParseAll(crnFullStr)
	if err != nil {
		return "", err
	}

	q := RESOURCE_QUERY_VERSION + "=" + crn.Version + "&" + RESOURCE_QUERY_CNAME + "=" + crn.CName + "&" + RESOURCE_QUERY_CTYPE + "=" + crn.CType +
		"&" + RESOURCE_QUERY_SERVICE_NAME + "=" + crn.ServiceName + "&" + RESOURCE_QUERY_LOCATION + "=" + crn.Location + "&" + RESOURCE_QUERY_SCOPE + "=" +
		crn.Scope + "&" + RESOURCE_QUERY_SERVICE_INSTANCE + "=" + crn.ServiceInstance + "&" + RESOURCE_QUERY_RESOURCE_TYPE + "=" + crn.ResourceType +
		"&" + RESOURCE_QUERY_RESOURCE + "=" + crn.Resource

	return q, nil
}

// NewNullString Call this function when performing SQL statements to insert any string values as SQL NULL if string is nil
func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// StringArrayToLower Convert all the strings in a string array to lowercase
func StringArrayToLower(strArray []string) []string {
	for i := range strArray {
		strArray[i] = strings.ToLower(strArray[i])
	}
	return strArray
}

// Contains tells whether string array a contains string x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// ContainDisplayName tells whether DisplayName array a contains name and language.
func ContainDisplayName(a []datastore.DisplayName, name string, language string) bool {
	for _, dn := range a {
		if name == dn.Name && language == dn.Language {
			return true
		}
	}
	return false
}

// ContainTag tells whether Tag array a contains id.
func ContainTag(a []datastore.Tag, id string) bool {
	for _, t := range a {
		if id == t.ID {
			return true
		}
	}
	return false
}

// PadRight Pad str with pad, to make str for a size of length
// e.g. str := PadRight("en", " ", 20), str will be "en                  "
func PadRight(str string, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}

// Delay sleps a transaction a number of seconds decalred at DefaultRetryDelay constant
func Delay() {
	log.Print(tlog.Log()+"Transaction failed. Going to retry after ", DefaultRetryDelay, " seconds.")
	time.Sleep(time.Second * time.Duration(DefaultRetryDelay))
}

// IsPostgresNonFailoverError Postgres errors that could not be due to failover
func IsPostgresNonFailoverError(err error) bool {
	// Postgres errors that could not be due to failover
	postgresNonFailoverErrors := []string{"pq: duplicate key value", "violates foreign key constraint", "pq: password authentication failed"} // the string is only the beginning of the error, not the entire error

	match := false
	for _, perr := range postgresNonFailoverErrors {
		if strings.Contains(err.Error(), perr) {
			match = true
			break
		}
	}
	return match
}

// ComputeMaintenanceRecordHashUsingReturn This is needed to make sure we don't miss any fields that are not necessarily populated
func ComputeMaintenanceRecordHashUsingReturn(record *datastore.MaintenanceInsert, recordFromDb *datastore.MaintenanceReturn) (*datastore.MaintenanceInsert, string) {
	//FCT := "CreateMaintenanceInsertForComputeHash:  "
	values := reflect.ValueOf(*record)
	fields := reflect.TypeOf(*record)

	numFields := fields.NumField()
	for i := 0; i < numFields; i++ {
		field := fields.Field(i)
		var value interface{}
		if values.Field(i).CanInterface() {
			value = values.Field(i).Interface()
			switch v := value.(type) {
			case string:
				// If value is empty and db value is not empty, use db value.
				if value == "" && reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).CanInterface() {
					valFromDb := reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).String()
					if valFromDb != "" {
						reflect.ValueOf(record).Elem().FieldByName(field.Name).SetString(valFromDb)
					}
				}
			case int:
				// If value is empty and db value is not empty, use db value.
				if value == "" && reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).CanInterface() {
					valFromDb := reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).Int()
					if valFromDb > 0 {
						reflect.ValueOf(record).Elem().FieldByName(field.Name).SetInt(valFromDb)
					}
				}
			case bool:
				// Just use the value of the new record
			default:
				log.Print(tlog.Log()+"Type is ", v)
				rt := reflect.TypeOf(value)
				switch rt.Kind() {
				case reflect.Slice, reflect.Array:
					if reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).CanInterface() {
						t := reflect.ValueOf(record).Elem().FieldByName(field.Name)
						if t.CanSet() {
							t.Set(reflect.ValueOf(value))
						} else {
							log.Panic("Could not set ", reflect.ValueOf(value))
						}

					}
				}
			}
		}
	}
	record.RecordHash = ComputeMaintenanceRecordHash(record)
	//log.Printf(FCT+"Record: %+v", record)
	return record, record.RecordHash
}

// ComputeMaintenanceRecordHash creates a string checksum make sure we don't miss any fields that are not necessarily populated
func ComputeMaintenanceRecordHash(recordOut *datastore.MaintenanceInsert) string {
	sort.Strings(recordOut.CRNFull)

	convertedSourceCreationTime, _ := utils.VerifyAndConvertTimestamp(recordOut.SourceCreationTime)
	convertedSourceUpdateTime, _ := utils.VerifyAndConvertTimestamp(recordOut.SourceUpdateTime)
	convertedPlannedStartTime, _ := utils.VerifyAndConvertTimestamp(recordOut.PlannedStartTime)
	convertedPlannedEndTime, _ := utils.VerifyAndConvertTimestamp(recordOut.PlannedEndTime)

	preHash := convertedSourceCreationTime +
		convertedSourceUpdateTime +
		convertedPlannedStartTime +
		convertedPlannedEndTime +
		recordOut.ShortDescription +
		recordOut.LongDescription +
		fmt.Sprintf("%s", recordOut.CRNFull) +
		recordOut.State +
		strconv.FormatBool(recordOut.Disruptive) +
		recordOut.SourceID +
		recordOut.Source +
		recordOut.RegulatoryDomain +
		strconv.Itoa(recordOut.MaintenanceDuration) +
		recordOut.DisruptionType +
		recordOut.DisruptionDescription +
		strconv.Itoa(recordOut.DisruptionDuration) +
		recordOut.CompletionCode +
		strconv.FormatBool(recordOut.PnPRemoved)

	sum := sha256.Sum256([]byte(preHash))
	return fmt.Sprintf("%x", sum)
}

// ComputeResourceRecordHashUsingReturn This is needed to make sure we don't miss any fields that are not necessarily populated
func ComputeResourceRecordHashUsingReturn(record *datastore.ResourceInsert, recordFromDb *datastore.ResourceReturn) (*datastore.ResourceInsert, string) {
	//FCT := "CreateResourceInsertForComputeHash:  "
	values := reflect.ValueOf(*record)
	fields := reflect.TypeOf(*record)
	fmt.Println("Types of input", reflect.TypeOf(*record), reflect.TypeOf(*recordFromDb))

	numFields := fields.NumField()
	for i := 0; i < numFields; i++ {
		field := fields.Field(i)
		var value interface{}
		if values.Field(i).CanInterface() {
			value = values.Field(i).Interface()
			switch v := value.(type) {
			case string:
				// If value is empty and db value is not empty, use db value.
				if value == "" && reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).CanInterface() {
					valFromDb := reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).String()
					if valFromDb != "" {
						reflect.ValueOf(record).Elem().FieldByName(field.Name).SetString(valFromDb)
					}
				}
			case *int:

				// If value of insert is empty and db value is not empty, use db value.
				if value.(*int) == nil && reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).CanInterface() {
					// if the db value is nil, we don't use it
					dbVal := reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).Interface()
					fmt.Printf("%v", dbVal)
					if dbVal.(*int) != nil {
						val := reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name)
						reflect.ValueOf(record).Elem().FieldByName(field.Name).Set(val)
						newVal := reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).Interface()
						fmt.Print(newVal)
					}
				}
			case bool:
				// Just use the value of the new record
			default:
				log.Print(tlog.Log()+"Type is ", v)
				rt := reflect.TypeOf(value)
				switch rt.Kind() {
				case reflect.Slice, reflect.Array:
					if reflect.ValueOf(recordFromDb).Elem().FieldByName(field.Name).CanInterface() {
						t := reflect.ValueOf(record).Elem().FieldByName(field.Name)
						if t.CanSet() {
							t.Set(reflect.ValueOf(value))
						} else {
							log.Panic("Could not set ", reflect.ValueOf(value))
						}

					}
				}
			}
		}
	}

	record.RecordHash = ComputeResourceRecordHash(record)
	//log.Printf(FCT+"Record: %+v", record)
	return record, record.RecordHash
}

// ComputeResourceRecordHash creates a string checksum make sure we don't miss any fields that are not necessarily populated
func ComputeResourceRecordHash(recordOut *datastore.ResourceInsert) string {
	sort.Strings(recordOut.Visibility)
	sort.Slice(recordOut.Tags, func(i, j int) bool {
		return recordOut.Tags[i].ID < recordOut.Tags[j].ID
	})
	sort.Slice(recordOut.DisplayNames, func(i, j int) bool {
		return recordOut.DisplayNames[i].Name < recordOut.DisplayNames[j].Name
	})

	convertedSourceCreationTime, _ := utils.VerifyAndConvertTimestamp(recordOut.SourceCreationTime)
	convertedSourceUpdateTime, _ := utils.VerifyAndConvertTimestamp(recordOut.SourceUpdateTime)

	preHash := convertedSourceCreationTime +
		convertedSourceUpdateTime +
		recordOut.CRNFull +
		recordOut.State +
		recordOut.OperationalStatus +
		recordOut.Source +
		recordOut.SourceID +

		// Don't need status update since that would get updated within PnP
		// recordOut.Status +
		//recordOut.StatusUpdateTime +
		recordOut.RegulatoryDomain +
		recordOut.CategoryID +
		strconv.FormatBool(recordOut.CategoryParent) +
		recordOut.CatalogParentID +
		strconv.FormatBool(recordOut.IsCatalogParent) +
		fmt.Sprintf("%s", recordOut.DisplayNames) +
		fmt.Sprintf("%s", recordOut.Visibility) +
		fmt.Sprintf("%s", recordOut.Tags)

	sum := sha256.Sum256([]byte(preHash))
	log.Println(tlog.Log(), "Pre Hash:", preHash, "\n:", fmt.Sprintf("%x", sum))
	return fmt.Sprintf("%x", sum)
}

// VerifyCRNArray checks CRN is valid in format and context
func VerifyCRNArray(crnArray []string) error {

	for i := 0; i < len(crnArray); i++ {
		// store crn's in database in lowercase
		crnArray[i] = strings.ToLower(crnArray[i])

		//crnStruct, ok := ParseCRNString(crnArray[i])
		crnStruct, err := crn.ParseAll(crnArray[i])
		if err != nil {
			return err
		}
		if crnStruct.Version == "" {
			return errors.New(ERR_NO_CRN_VERSION + crnArray[i])
		}
		if crnStruct.ServiceName == "" {
			return errors.New(ERR_NO_SERVICE + crnArray[i])
		}

		// If cname is bluemix, then ctype and location should have values;
		// otherwise it is referring to a Gaas resource, which may not have cname, ctype and location
		if crnStruct.CName == DefaultIBMCloudCname {
			if crnStruct.CType == "" {
				return errors.New(ERR_NO_CTYPE + crnArray[i])
			}
			if crnStruct.Location == "" {
				return errors.New(ERR_NO_LOCATION + crnArray[i])
			}
		}
	}
	return nil
}

// ConvertToResourceLookupCRN If crn is Gaas resource, return true, and there is no cname, ctype and location in resource lookup
// Otherwise return false, and no change in the crn for resource lookup;
func ConvertToResourceLookupCRN(crnIn string) (bool, string, error) {

	//crnMask, err := crn.Parse(crn)
	crnMask, err := crn.Parse(crnIn)
	if err != nil {
		return false, crnIn, err
	}
	isGaasResource := api.IsCrnGaasResource(crnIn, ctxt.Context{})

	log.Println(tlog.Log()+"ConvertToResourceLookupCRN: original crn:", crnIn, "isGaasResource:", isGaasResource)

	if isGaasResource {
		// Gaas resource
		return isGaasResource, "crn:v1:::" + crnMask.ServiceName + ":::::", nil
	}

	// not Gaas resource, and no change in crn

	return isGaasResource, crnIn, nil
}

// CreateCrnFilter Create the query string from the crn in incident/maintenance/watch for GetIncidentByQuery, GetMaintenanceByQuery, InsertWatchByTypeForSubscriptionStatement
// watchWildcards is WatchInsert.Wildcards only applies to watch
func CreateCrnFilter(crn string, watchWildcards string) (string, error, int) {
	//FCT := "CreateCrnFilter: "

	queryStr := ""
	isGaasResource, crnToLookup, err := ConvertToResourceLookupCRN(crn)
	if err != nil {
		log.Println(tlog.Log()+"ERROR: ConvertToResourceLookupCRN returns error: ", err)
		return queryStr, err, http.StatusBadRequest
	}
	if watchWildcards == "true" || !isGaasResource { //watchWildcards is for InsertWatchByTypeForSubscriptionStatement
		// wildcard search
		queryStr = "crn=" + crnToLookup
	} else {
		// non-wildcard search, no cname, ctype and location for Gaas resource
		queryStr, err = CreateNonWildcardCRNQueryString(crnToLookup)
		if err != nil {
			log.Println(tlog.Log()+"ERROR CreateNonWildcardCRNQueryString returns error: ", err)
			return queryStr, err, http.StatusInternalServerError
		}
	}
	log.Println(tlog.Log()+"DEBUG: crn: ", crn, "isGaasResource:", isGaasResource, "queryStr:", queryStr)
	return queryStr, nil, http.StatusOK
}

// CreateCrnFilterQry Create the query string from the crn in incident/maintenance/watch for GetIncidentByQuery, GetMaintenanceByQuery, InsertWatchByTypeForSubscriptionStatement
// watchWildcards is WatchInsert.Wildcards only applies to watch
func CreateCrnFilterQry(crn string, watchWildcards string) (crnQryFilter string, errCode int, isGaasResource bool, err error) {
	queryStr := ""
	isGaasResource, crnToLookup, err := ConvertToResourceLookupCRN(crn)
	if err != nil {
		log.Println(tlog.Log()+"ERROR: ConvertToResourceLookupCRN returns error: ", err)
		return "", http.StatusBadRequest, isGaasResource, err
	}
	if watchWildcards == "true" || !isGaasResource { //watchWildcards is for InsertWatchByTypeForSubscriptionStatement
		// wildcard search
		queryStr = "crn=" + crnToLookup
	} else {
		// non-wildcard search, no cname, ctype and location for Gaas resource
		queryStr, err = CreateNonWildcardCRNQueryString(crnToLookup)
		if err != nil {
			log.Println(tlog.Log()+"ERROR CreateNonWildcardCRNQueryString returns error: ", err)
			return queryStr, http.StatusInternalServerError, isGaasResource, err
		}
	}
	log.Println(tlog.Log()+"DEBUG: crn: ", crn, "isGaasResource:", isGaasResource, "queryStr:", queryStr)
	return queryStr, http.StatusOK, isGaasResource, nil
}

// GetRecordsCount gets the number of records retuned by the query passed
// func GetRecordsCount(database *sql.DB, tableName string, whereClause string) (rowCnt int) {

// 	strQryCnt := "select count(*) from " + tableName + whereClause
// 	_ = database.QueryRow(strQryCnt).Scan(&rowCnt) // get the number of records
// 	log.Println(tlog.Log(), strQryCnt, "rows:", rowCnt)
// 	return rowCnt
// }

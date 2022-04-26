package servicenow

// Data structures and functions for interacting with ServiceNow
// See the file ossrecord.go for definitions of the various field types

//go:generate easytags $GOFILE

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

type cachedSNRecordsType = map[ossrecord.ServiceNowSysid]*ConfigurationItem

// The cachedSNRecords variables cache the SN records obtained directly from the batch API,
// to avoid having to fetch them again individually
// (or to compare them against the individually fetched records if verifyCachedRecords is true)
var cachedSNRecords = make(map[ServiceNowDomain]cachedSNRecordsType)
var cachedSNRecordsTest = make(map[ServiceNowDomain]cachedSNRecordsType)

func getCachedSNRecords(testMode bool, snDomain ServiceNowDomain) cachedSNRecordsType {
	if testMode {
		cachedInstance := cachedSNRecordsTest[snDomain]
		if cachedInstance == nil {
			cachedInstance = make(cachedSNRecordsType)
			cachedSNRecordsTest[snDomain] = cachedInstance
		}
		return cachedInstance
	}

	cachedInstance := cachedSNRecords[snDomain]
	if cachedInstance == nil {
		cachedInstance = make(cachedSNRecordsType)
		cachedSNRecords[snDomain] = cachedInstance
	}
	return cachedInstance
}

var verifyCachedRecords = false /* XXX */

// ServiceNowDomain represents the different domains that are valid for ServiceNow
type ServiceNowDomain string

const (
	// WATSON is the ServiceNowDomain for the watson (commercial) domain
	WATSON ServiceNowDomain = "watson"

	// CLOUDFED is the ServiceNowDomain for the cloudfed (US FedRamp) regulated domain
	CLOUDFED ServiceNowDomain = "cloudfed"
)

// ReadServiceNowRecord reads one service entry through the ServiceNow API, given its name
func ReadServiceNowRecord(name ossrecord.CRNServiceName, testMode bool, snDomain ServiceNowDomain) (*ConfigurationItem, []*ossvalidation.ValidationIssue, error) {
	var testModeString string
	if testMode {
		testModeString = testModePrefix
	}
	sysids, found := LookupServiceNowSysids(name, testMode, snDomain)
	if !found || len(sysids) == 0 {
		err := rest.MakeHTTPError(nil, nil, true, "%sCannot find a ServiceNow sysid for \"%s\"", testModeString, name)
		return nil, nil, err
	}

	if len(sysids) == 1 {
		// Common case
		rec, err := ReadServiceNowRecordBySysid(sysids[0], testMode, snDomain)
		return rec, nil, err
	}

	var issues []*ossvalidation.ValidationIssue
	retireds := make([]*ConfigurationItem, 0, len(sysids))
	nonRetireds := make([]*ConfigurationItem, 0, len(sysids))
	for _, id := range sysids {
		rec, err := ReadServiceNowRecordBySysid(id, testMode, snDomain)
		if err != nil {
			return nil, nil, debug.WrapError(err, "%sError while checking multiple sysids (%d) for \"%s\"", testModeString, len(sysids), name)
		}
		status, _ := ParseOperationalStatus(string(rec.GeneralInfo.OperationalStatus))
		//		fmt.Printf("DEBUG: found entry(%s) - status=%s/%s\n", name, rec.GeneralInfo.OperationalStatus, status)
		if status == ossrecord.RETIRED {
			retireds = append(retireds, rec)
		} else if status == ossrecord.THIRDPARTY {
			// FIXME: *TEMPORARY: we treat status=THIRDPARTY the same as RETIRED because of a bug in ServiceNow
			// see issue https://github.ibm.com/cloud-sre/osscatalog/issues/98
			if testMode {
				v := ossvalidation.NewIssue(ossvalidation.WARNING, testModePrefix+"Found ServiceNow entry/sysid with state=THIRDPARTY and other entries with the same name -- assuming the THIRDPARTY entry is actually RETIRED (see ServiceNow bug, issue #98)", `name="%s"   id=%s`, name, id).TagConsistency()
				issues = append(issues, v)
				debug.PrintError(strings.TrimSpace(v.String()))
			} else {
				v := ossvalidation.NewIssue(ossvalidation.WARNING, "Found ServiceNow entry/sysid with state=THIRDPARTY and other entries with the same name -- assuming the THIRDPARTY entry is actually RETIRED (see ServiceNow bug, issue #98)", `name="%s"   id=%s`, name, id).TagConsistency()
				issues = append(issues, v)
				debug.PrintError(strings.TrimSpace(v.String()))
			}
			retireds = append(retireds, rec)
		} else {
			nonRetireds = append(nonRetireds, rec)
		}
	}
	switch {
	case len(nonRetireds) == 1:
		if len(retireds) > 0 {
			if testMode {
				v := ossvalidation.NewIssue(ossvalidation.INFO, testModePrefix+"Found 1 non-retired ServiceNow entry/sysid and some retired entries/sysids for the same name", `retired=%d   name="%s"`, len(retireds), name).TagConsistency()
				issues = append(issues, v)
			} else {
				v := ossvalidation.NewIssue(ossvalidation.INFO, "Found 1 non-retired ServiceNow entry/sysid and some retired entries/sysids for the same name", `retired=%d   name="%s"`, len(retireds), name).TagConsistency()
				issues = append(issues, v)
			}
			return nonRetireds[0], issues, nil
		}
		return nonRetireds[0], issues, nil
	case len(nonRetireds) == 0:
		if len(retireds) > 0 {
			return nil, issues, fmt.Errorf("%sFound more than one retired and no non-retired ServiceNow entres/sysids for \"%s\" : non-retired=%d   retired=%d", testModeString, name, len(nonRetireds), len(retireds))
		} else if len(retireds) == 1 {
			return retireds[0], issues, nil
		} else {
			panic(fmt.Sprintf("%sFound no retired or non-retired ServiceNow entries/sysids for \"%s\" but sysids list has %d entries", testModeString, name, len(sysids)))
		}
	case len(nonRetireds) > 0:
		return nil, nil, fmt.Errorf("%sFound more than one non-retired ServiceNow entry/sysid for \"%s\" : non-retired=%d   retired=%d", testModeString, name, len(nonRetireds), len(retireds))
	}
	return nil, nil, nil // unreachable
}

// ReadServiceNowRecordBySysid reads one service entry through the ServiceNow API, given its ServiceNow sysid
func ReadServiceNowRecordBySysid(sysid ossrecord.ServiceNowSysid, testMode bool, snDomain ServiceNowDomain) (*ConfigurationItem, error) {
	var cachedEntry, entry *ConfigurationItem
	var ok bool
	var testModeString string
	if testMode {
		testModeString = testModePrefix
	}
	if cachedEntry, ok = getCachedSNRecords(testMode, snDomain)[sysid]; !ok || verifyCachedRecords {
		actualURL, headers := getServiceNowCallInfo(string(sysid), 1, 1, testMode, snDomain)
		token, err := getServiceNowToken(testMode, snDomain)
		if err != nil {
			err = debug.WrapError(err, "%sCannot get token for ServiceNow", testModeString)
			return nil, err
		}

		var resultContainer struct {
			Result ConfigurationItem `json:"result"`
		}
		err = rest.DoHTTPGet(actualURL, "Bearer "+token, headers, "ServiceNow", debug.ServiceNow, &resultContainer)
		if err != nil {
			if cachedEntry != nil {
				return nil, debug.WrapError(err, "%sError re-reading cached entry %s", testModeString, cachedEntry.String())
			}
			return nil, err
		}
		entry = &resultContainer.Result
		entry.normalize()
	} else {
		entry = cachedEntry
	}

	if cachedEntry != nil && entry != nil && cachedEntry != entry {
		out := compare.Output{}
		compare.DeepCompare("batch.load", cachedEntry, "individual.load", entry, &out)
		if out.NumDiffs() > 0 {
			debug.PrintError("%sFound differences in ServiceNow records for %s fetch by batch load vs individual load:\n%s", testModeString, cachedEntry.String(), out.StringWithPrefix("      "))
		}
	}

	return entry, nil
}

const testModePrefix = "(Test Mode): "

// ListServiceNowRecords lists all ServicNow entries and calls the special handler function for each entry
func ListServiceNowRecords(pattern *regexp.Regexp, testMode bool, snDomain ServiceNowDomain, handler func(e *ConfigurationItem, issues []*ossvalidation.ValidationIssue)) error {
	resourceCount := 0
	rawEntries := 0
	totalEntries := 0
	errorCount := 0

	var testModeString string
	if testMode {
		testModeString = testModePrefix
	}
	// If sysidTable is not empty at this point, it means that we must have pre-loaded it from a -sysid input file
	// Otherwise, attempt a batch load of all ServiceNow records
	if len(getSysidTable(testMode, snDomain)) == 0 {
		offset := 0
		const size = 100
		token, err := getServiceNowToken(testMode, snDomain)
		if err != nil {
			err = debug.WrapError(err, "%sCannot get token for ServiceNow", testModeString)
			return err
		}
	FETCH:
		for {
			actualURL, headers := getServiceNowCallInfo("", offset, size, testMode, snDomain)
			var resultContainer struct {
				Result struct {
					Resources []*ConfigurationItem `json:"resources"`
					Count     float64              `json:"count"`
					Limit     float64              `json:"limit"`
					Offset    float64              `json:"offset"`
					Query     string               `json:"query"`
				} `json:"result"`
			}

			switch apiVersion {
			case apiV1:
				var resultContainerV1 struct {
					Result []*ConfigurationItem `json:"result"`
				}
				err = rest.DoHTTPGet(actualURL, "Bearer "+token, headers, "ServiceNow", debug.ServiceNow, &resultContainerV1)
				resultContainer.Result.Resources = resultContainerV1.Result
			case apiV2:
				err = rest.DoHTTPGet(actualURL, "Bearer "+token, headers, "ServiceNow", debug.ServiceNow, &resultContainer)
			default:
				panic(fmt.Sprintf("Unknown ServiceNow API version: %d", apiVersion))
			}
			if err != nil {
				return err
			}

			cachedSNRecords := getCachedSNRecords(testMode, snDomain)
			for _, e := range resultContainer.Result.Resources {
				resourceCount++
				e.normalize()
				if testMode {
					e.SysID = testModePrefix + e.SysID
				}
				RecordSysid(ossrecord.CRNServiceName(e.CRNServiceName), ossrecord.ServiceNowSysid(e.SysID), testMode, snDomain)
				cachedSNRecords[ossrecord.ServiceNowSysid(e.SysID)] = e
			}

			switch apiVersion {
			case apiV1:
				if len(resultContainer.Result.Resources) >= apiV1BatchSize {
					debug.PrintError("%sListing entries directly through ServiceNow API V1: found %d entries -- ** limit reached; there might be additional entries **\n", testModeString, len(resultContainer.Result.Resources))
				} else {
					debug.Info("%sListing entries directly through ServiceNow API V1: found %d entries\n", testModeString, len(resultContainer.Result.Resources))
				}
				break FETCH
			case apiV2:
				if len(resultContainer.Result.Resources) >= size {
					offset += len(resultContainer.Result.Resources)
				} else {
					if resourceCount != int(resultContainer.Result.Count) {
						return fmt.Errorf("%sListServiceNowRecords(): ServiceNow API returned %d actual resources but claims the total count should be %d (url=%s)", testModeString, resourceCount, int(resultContainer.Result.Count), actualURL)
					}
					break FETCH
				}
			default:
				panic(fmt.Sprintf("Unknown ServiceNow API version: %d", apiVersion))
			}
		}
	}

	for name := range getSysidTable(testMode, snDomain) {
		if (rawEntries % 30) == 0 {
			debug.Info("%sLoading one batch of entries from ServiceNow (%d/%d entries so far)", testModeString, totalEntries, rawEntries)
		}
		rawEntries++

		if pattern.FindString(string(name)) == "" {
			continue
		}

		entry, issues, err := ReadServiceNowRecord(name, testMode, snDomain)
		if err != nil {
			// Since each ServiceNow call is independent, and some sysids might be imvalid, we can afford to proceed even after an error
			errorCount++
			if errorCount < 10 /* XXX */ {
				debug.PrintError("%sListServiceNowRecords(): Error reading one ServiceNow entry: %v", testModeString, err)
				continue
			} else {
				return debug.WrapError(err, "%sListServiceNowRecords(): Too many errors(%d) reading ServiceNow entries - aborting", testModeString, errorCount)
			}
		}
		handler(entry, issues)
		totalEntries++
	}

	debug.Info("%sRead %d entries from ServiceNow", testModeString, totalEntries)
	return nil
}

// String returns a short string representation of this ServiceNow record
func (r *ConfigurationItem) String() string {
	if debug.IsDebugEnabled(debug.Fine) {
		return fmt.Sprintf(`ServiceNow{Name:"%s", Status:"%s", Type:"%s", SysID:"%s", Addr=%p}`, r.CRNServiceName, r.GeneralInfo.OperationalStatus, r.GeneralInfo.EntryType, r.SysID, r)
	}
	return fmt.Sprintf(`ServiceNow{Name:"%s", Status:"%s", Type:"%s", SysID:"%s"}`, r.CRNServiceName, r.GeneralInfo.OperationalStatus, r.GeneralInfo.EntryType, r.SysID)
}

var _ fmt.Stringer = &ConfigurationItem{}

// SyncServiceNowRecord sync servicenow CI info through servicenow API by given service
func SyncServiceNowRecord(token string, service *ossrecord.OSSService, testMode bool, snDomain ServiceNowDomain) (*ossrecord.OSSService, error) {
	var entry *ConfigurationItem
	var testModeString string
	if testMode {
		testModeString = testModePrefix
	}
	// get token
	var err error
	if token == "" {
		token, err = getServiceNowToken(testMode, snDomain)
		if err != nil {
			err = debug.WrapError(err, "%sCannot get token for ServiceNow", testModeString)
			return nil, err
		}
	}

	name := string(service.ReferenceResourceName)
	actualURL, headers := getServiceNowCallInfo(string(name), 1, 1, testMode, snDomain)

	var resultContainer struct {
		Result ConfigurationItem `json:"result"`
	}
	err = rest.DoHTTPGet(actualURL, "Bearer "+token, headers, "ServiceNow", debug.ServiceNow, &resultContainer)
	if err != nil {
		return nil, err
	}
	entry = &resultContainer.Result
	entry.normalize()

	return MergeCIToOSSService(entry, service, snDomain)
}

func CreateServiceNowRecordFromOSSService(token string, service *ossrecord.OSSService, testMode bool, snDomain ServiceNowDomain) (*ossrecord.OSSService, error) {
	ci := GetCIFromOSSService(service)
	ci, err := CreateServiceNowRecord(token, ci, testMode, snDomain)
	if err == nil {
		return MergeCIToOSSService(ci, service, snDomain)
	}
	debug.Info(err.Error())
	return service, err
}

// CreateServiceNowRecord create a new servicenow CI through the ServiceNow API
func CreateServiceNowRecord(token string, ci *ConfigurationItem, testMode bool, snDomain ServiceNowDomain) (entry *ConfigurationItem, err error) {
	var testModeString string
	if testMode {
		testModeString = testModePrefix
	}

	// get token
	if token == "" {
		token, err = getServiceNowToken(testMode, snDomain)
		if err != nil {
			err = debug.WrapError(err, "%sCannot get token for ServiceNow", testModeString)
			return nil, err
		}
	}

	var resultContainer struct {
		Result ConfigurationItem `json:"result"`
	}

	// create CI
	actualURL, headers := getServiceNowCallInfo("", -1, -1, testMode, snDomain)
	err = rest.DoHTTPPostOrPut(http.MethodPost, actualURL, "Bearer "+token, headers, ci, &resultContainer, "ServiceNow", debug.ServiceNow)
	if err != nil {
		err = debug.WrapError(err, "%sCreate ServiceNow ConfigurationItem with name \"%s\" failed.", testModeString, ci.CRNServiceName)
		return nil, err
	}
	entry = &resultContainer.Result
	entry.normalize()

	return entry, nil
}

// UpdateServiceNowRecordFromOSSService update servicenow CI from given ossservice
func UpdateServiceNowRecordFromOSSService(token string, service *ossrecord.OSSService, testMode bool, snDomain ServiceNowDomain) (result *ossrecord.OSSService, err error) {
	ci := GetCIFromOSSService(service)
	ci, err = UpdateServiceNowRecordByName(token, ci, testMode, snDomain)
	if err == nil {
		return MergeCIToOSSService(ci, service, snDomain)
	}
	return service, err

}

// UpdateServiceNowRecordByName update servicenow CI through the ServiceNow API
func UpdateServiceNowRecordByName(token string, ci interface{}, testMode bool, snDomain ServiceNowDomain) (entry *ConfigurationItem, err error) {
	var testModeString string
	if testMode {
		testModeString = testModePrefix
	}

	// get token
	if token == "" {
		token, err = getServiceNowToken(testMode, snDomain)
		if err != nil {
			err = debug.WrapError(err, "%sCannot get token for ServiceNow", testModeString)
			return nil, err
		}
	}

	serviceName := ""
	// TODO: do we really need to use reflection here? Why not simply a type switch?
	switch reflect.TypeOf(ci).Kind() {
	case reflect.Struct:
		v := reflect.ValueOf(ci).FieldByName("CRNServiceName")
		serviceName = v.Interface().(string)
		break
	case reflect.Ptr:
		ciVal := reflect.ValueOf(ci).Elem()
		v := ciVal.FieldByName("CRNServiceName")
		serviceName = v.Interface().(string)
		break
	}

	if serviceName == "" {
		err = debug.WrapError(err, "%sService name can't be empty", testModeString)
		return nil, err
	}

	actualURL, headers := getServiceNowCallInfo(serviceName, 1, 1, testMode, snDomain)
	var resultContainer struct {
		Result ConfigurationItem `json:"result"`
	}
	// update CI
	err = rest.DoHTTPPostOrPut(http.MethodPatch, actualURL, "Bearer "+token, headers, ci, &resultContainer, "ServiceNow", debug.ServiceNow)
	if err != nil {
		err = debug.WrapError(err, "%sUpdate ServiceNow ConfigurationItem with name \"%s\" failed.", testModeString, serviceName)
		return nil, err
	}
	entry = &resultContainer.Result
	entry.normalize()

	return entry, nil
}

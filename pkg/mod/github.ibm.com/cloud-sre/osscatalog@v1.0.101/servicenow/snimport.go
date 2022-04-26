package servicenow

import (
	"fmt"
	"os"
	"regexp"

	"github.com/gocarina/gocsv"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// SNImport represents one record from the SN import csv file
type SNImport struct {
	Name                                string `csv:"name"`                                   // 0
	DisplayName                         string `csv:"u_catalog_name"`                         // 1
	FullCRN                             string `csv:"u_cloud_resource_name"`                  // 2
	StatusPageNotificationCategoryID    string `csv:"u_status_page_notification_category_id"` // 3
	Tier1SupportAssignmentGroup         string `csv:"support_group"`                          // 4
	Tier1OperationsAssignmentGroup      string `csv:"assignment_group"`                       // 5
	Tier2SupportEscalationType          string `csv:"u_support_tier2_escalation_type"`        // 6
	Tier2SupportAssignmentGroup         string `csv:"u_tier2_support_group"`                  // 7
	Tier2SupportEscalationGitHubRepo    string `csv:"u_support_tier2_esc_github_repo"`        // 8
	Tier2OperationsEscalationType       string `csv:"u_ops_tier2_escalation_type"`            // 9
	Tier2OperationsAssignmentGroup      string `csv:"u_tier2_assignment_group"`               // 10
	Tier2OperationsEscalationGitHubRepo string `csv:"u_ops_tier2_esc_github_repo"`            // 11
	EntryType                           string `csv:"sys_class_name"`                         // 12
	ClientExperience                    string `csv:"u_client_experience"`                    // 13
	CustomerFacing                      bool   `csv:"u_customer_facing"`                      // 14
	TOCEnabled                          bool   `csv:"u_toc_enabled"`                          // 15
	OperationalStatus                   string `csv:"operational_status"`                     // 16
	OfferingManager                     string `csv:"owned_by"`                               // 17
	StatusPageNotificationGroup         string `csv:"u_status_page_notification_group"`       // 18
	CreatedBy                           string `csv:"sys_created_by"`                         // 19
	UpdatedBy                           string `csv:"sys_updated_by"`                         // 20
	Segment                             string `csv:"u_segment"`                              // 21
	Tribe                               string `csv:"u_tribe"`                                // 22
	OperationsManager                   string `csv:"managed_by"`                             // 23
	SupportManager                      string `csv:"supported_by"`                           // 24
}

var allImportedRecords map[string]*SNImport

// HasServiceNowImport return true if we have imported records from the ServiceNow import file; false otherwise
func HasServiceNowImport() bool {
	return allImportedRecords != nil
}

// ReadServiceNowImportFile reads a file exported from Configuration Items report in ServiceNow
func ReadServiceNowImportFile(filename string) error {
	file, err := os.Open(filename) // #nosec G304
	if err != nil {
		return debug.WrapError(err, "Cannot open ServiceNow import file %s", filename)
	}
	defer file.Close() // #nosec G307

	var imported = []*SNImport{}
	gocsv.FailIfDoubleHeaderNames = true
	gocsv.FailIfUnmatchedStructTags = true
	err = gocsv.UnmarshalFile(file, &imported)
	if err != nil {
		return debug.WrapError(err, "Cannot parse ServiceNow import file %s", filename)
	}

	allImportedRecords = make(map[string]*SNImport)

	countRecords := 0
	for _, entry := range imported {
		countRecords++
		debug.Debug(debug.ServiceNow, "Updating Record %3d: %v", countRecords, entry)
		RegisterServiceNowImport(entry)
	}
	debug.Info("Completed reading the ServiceNow import file %s :  %d records\n", filename, countRecords)
	return nil

	/*
		reader := csv.NewReader(file)
		countLocalErrors := 0
		countRecords := 0
		//reader.FieldsPerRecord = 3
		firstTime := true
		for {
			record, err := reader.Read()
			if err == io.EOF {
				debug.Info("Completed reading the ServiceNow import file %s :  %d records;  %d errors\n", filename, countRecords, countLocalErrors)
				return nil
			}
			if err != nil {
				debug.PrintError("Error while parsing one record in ServiceNow import file: %v", err)
				countLocalErrors++
				continue
			}
			if firstTime {
				// Skip the title row
				firstTime = false
				continue
			}
			countRecords++
			// TODO: initialize remaining SN import fields
			entry := new(SNImport)
			entry.Name = record[0]
			entry.DisplayName = record[1]
			entry.FullCRN = record[2]
			entry.StatusPageNotificationCategoryID = record[3]
			entry.OfferingManager = record[17]
			entry.Segment = record[21]
			entry.Tribe = record[22]
			entry.OperationsManager = record[23]
			entry.SupportManager = record[24]
			debug.Debug(debug.ServiceNow, "Updating Record %3d: %#v  %v", countRecords, record, entry)
			prior, found := allImportedRecords[entry.Name]
			if found {
				panic(fmt.Sprintf(`ReadServiceNowImportFile() found duplicate ServiceNow entry:  (1)=%v  (2)=%v`, prior, entry))
			}
			allImportedRecords[entry.Name] = entry
		}
	*/
}

// GetServiceNowImport returns the ServiceNow import record associated with a given CRN service name
func GetServiceNowImport(name ossrecord.CRNServiceName) (*SNImport, bool) {
	result, ok := allImportedRecords[string(name)]
	return result, ok
}

// ListServiceNowImport lists all entries from the ServiceNow import file (not the API!) and calls the special handler function for each entry
func ListServiceNowImport(pattern *regexp.Regexp, handler func(e *SNImport)) error {
	for _, e := range allImportedRecords {
		if pattern.FindString(string(e.Name)) == "" {
			continue
		}
		handler(e)
	}
	return nil
}

// RegisterServiceNowImport registers one SNImport record in the table for future lookup
// Used only internally from this package, and for unit tests from other packages
func RegisterServiceNowImport(entry *SNImport) {
	prior, found := allImportedRecords[entry.Name]
	if found {
		if prior.OperationalStatus == "Retired" {
			debug.Info(`RegisterServiceNowImport() ignoring Retired entry "%s" in favor of another non-retired entry with the same name (1)`, prior.Name)
			// fall through and re-register
		} else if entry.OperationalStatus == "Retired" {
			debug.Info(`RegisterServiceNowImport() ignoring Retired entry "%s" in favor of another non-retired entry with the same name (2)`, prior.Name)
			return
		} else {
			panic(fmt.Errorf(`RegisterServiceNowImport() found duplicate ServiceNow entry:  (1)=%v  (2)=%v`, prior, entry))
		}
	}
	allImportedRecords[entry.Name] = entry
}

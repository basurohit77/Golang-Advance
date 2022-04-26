// Functions for dealing with the "legacy" CRN validation script results (python-based)

package legacy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tealeg/xlsx"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// Entry represents the information from one entry in the legacy input file (CRN validation spreadsheet)
type Entry struct {
	Name              ossrecord.CRNServiceName
	EntryType         ossrecord.EntryType
	OperationalStatus ossrecord.OperationalStatus
	Exceptions        string
	ValidationStatus  string
	DuplicateOf       string
	Notes             string
	AllSources        map[string]string
}

var legacyInputFile string

const (
	colKey = iota
	colName
	colException
	colPreviousValidationStatus
	colValidationStatus
	colNotes
	colDuplicateOf
	colNumErrors
	colNumWarnings
	colNumIgnored
	colNumDeferred
	colEntryType
	colOperationalStatus
	colSegment
	colTribe
	colOfferingManager
	colServiceNowEntryCreatedBy
	colServiceNowEntryLastUpdatedBy
	colEntriesUsingCanonicalName
	colNameVariant1
	colEntriesUsingNameVariant1
	colNameVariant2
	colEntriesUsingNameVariant2
	colNameVariant3
	colEntriesUsingNameVariant3
)

// SetCRNValidationFile defines the CRN Validation spreadsheet file to use for loading Legacy records
func SetCRNValidationFile(fname string) error {
	legacyInputFile = fname
	return nil
}

func checkColTitle(row *xlsx.Row, index int, expected string, output *strings.Builder) {
	if row.Cells[index].Value != expected {
		output.WriteString(fmt.Sprintf(`ListLegacyRecords(): unexpected column title(%2d): expected "%s"   found "%s"\n`, index, expected, row.Cells[index].Value))
	}
}

// ListLegacyRecords lists all Legacy records from the CRN validation spreadsheet and calls the special handler function for each entry
func ListLegacyRecords(pattern *regexp.Regexp, handler func(e *Entry)) error {
	if legacyInputFile == "" {
		debug.Info("no legacy input file specified")
		return nil
	}

	var countRecords = 0

	file, err := xlsx.OpenFile(legacyInputFile)
	if err != nil {
		return debug.WrapError(err, "Cannot open the legacy input file")
	}

	sheet, ok := file.Sheet["CRN Master List"]
	if !ok {
		return fmt.Errorf(`Cannot find sheet "CRN Master List" in spreadsheet "%s"`, legacyInputFile)
	}

	for ix, row := range sheet.Rows {
		if len(row.Cells) < 14 {
			debug.PrintError(`ListLegacyRecords(): ignoring row with not enough columns: %d: %v`, ix, row)
			continue
		}
		if ix == 0 {
			// Special checks for title row
			output := &strings.Builder{}
			checkColTitle(row, colName, "Canonical Name", output)
			checkColTitle(row, colException, "Exception to validation rules (if any)", output)
			checkColTitle(row, colValidationStatus, "New Validation Status", output)
			checkColTitle(row, colNotes, "Notes", output)
			checkColTitle(row, colDuplicateOf, "Duplicate of", output)
			checkColTitle(row, colEntryType, "Type", output)
			checkColTitle(row, colOperationalStatus, "Operational Status", output)
			checkColTitle(row, colEntriesUsingCanonicalName, "Entries using Canonical Name", output)
			checkColTitle(row, colNameVariant1, "Name Variant #1", output)
			checkColTitle(row, colEntriesUsingNameVariant1, "Entries using Name Variant #1", output)
			checkColTitle(row, colNameVariant2, "Name Variant #2", output)
			checkColTitle(row, colEntriesUsingNameVariant2, "Entries using Name Variant #2", output)
			checkColTitle(row, colNameVariant3, "Name Variant #3", output)
			checkColTitle(row, colEntriesUsingNameVariant3, "Entries using Name Variant #3", output)
			str := output.String()
			if len(str) > 0 {
				return fmt.Errorf("ListLegacyRecords(): cannot parse the input\n%s", str)
			}
		} else {
			cells := row.Cells
			name := ossrecord.CRNServiceName(strings.TrimSpace(cells[colName].Value))
			if pattern != nil && pattern.FindString(string(name)) == "" {
				continue
			}

			e := Entry{}
			e.Name = name
			e.EntryType, _ = ParseEntryType(strings.TrimSpace(cells[colEntryType].Value))
			e.OperationalStatus, _ = ParseOperationalStatus(strings.TrimSpace(cells[colOperationalStatus].Value))
			e.Exceptions = strings.TrimSpace(cells[colException].Value)
			e.ValidationStatus = strings.TrimSpace(cells[colValidationStatus].Value)
			e.DuplicateOf = strings.TrimSpace(cells[colDuplicateOf].Value)
			e.Notes = strings.TrimSpace(cells[colNotes].Value)
			e.AllSources = make(map[string]string)
			if v := strings.TrimSpace(cells[colEntriesUsingCanonicalName].Value); v != "" {
				e.AllSources[string(e.Name)] = v
			}
			if v := strings.TrimSpace(cells[colEntriesUsingNameVariant1].Value); v != "" {
				e.AllSources[strings.TrimSpace(cells[colNameVariant1].Value)] = v
			}
			if v := strings.TrimSpace(cells[colEntriesUsingNameVariant2].Value); v != "" {
				e.AllSources[strings.TrimSpace(cells[colNameVariant2].Value)] = v
			}
			if v := strings.TrimSpace(cells[colEntriesUsingNameVariant3].Value); v != "" {
				e.AllSources[strings.TrimSpace(cells[colNameVariant3].Value)] = v
			}
			handler(&e)
			countRecords++
		}
	}
	debug.Info("Completed reading the Legacy CRN Validation report %s :  %d records\n", legacyInputFile, countRecords)

	return nil
}

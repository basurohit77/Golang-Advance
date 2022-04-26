// Code to read the "OSSCAT1" report from ClearingHouse

package clearinghouse

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gocarina/gocsv"
	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// DeliverableID is the unique identifier for one ClearingHouse entry
type DeliverableID string

// CHSummaryEntry represents the information from one entry in the input file
type CHSummaryEntry struct {
	Name             string        `csv:"Name"`
	CodeName         string        `csv:"Code Name"`
	DeliverableID    DeliverableID `csv:"Deliverable ID"`
	OfficialName     string        `csv:"Official Name"`
	ShortName        string        `csv:"Short Name"`
	PIDs             string        `csv:"PID/MTM"`
	XMajorUnitUTL10  string        `csv:"Major Unit (UTL10)"` // Disabled - we want to use the full record obtained from GetFullRecordByID()
	XMinorUnitUTL15  string        `csv:"Minor Unit (UTL15)"` // Disabled - we want to use the full record obtained from GetFullRecordByID()
	XMarketUTL17     string        `csv:"Market (UTL17)"`     // Disabled - we want to use the full record obtained from GetFullRecordByID()
	XPortfolioUTL20  string        `csv:"Portfolio (UTL20)"`  // Disabled - we want to use the full record obtained from GetFullRecordByID()
	XOfferingClassL1 string        `csv:"Offer Class L1"`     // Disabled - we want to use the full record obtained from GetFullRecordByID()
	XOfferingClassL2 string        `csv:"Offer Class L2"`     // Disabled - we want to use the full record obtained from GetFullRecordByID()
	CRNServiceName   string        `csv:"CRN Name"`
}

// String returns a short string identifier for the ClearingHouse entry
func (ch *CHSummaryEntry) String() string {
	return MakeCHLabel(ch.Name, ch.DeliverableID)
}

var allCHSummaryEntriesByPID map[string][]*CHSummaryEntry
var allCHSummaryEntriesByClearingHouseID map[DeliverableID]*CHSummaryEntry
var allCHSummaryEntriesByCRNServiceName map[string][]*CHSummaryEntry

var missingCHSummaryEntriesByPID map[string][]string
var missingCHSummaryEntriesByClearingHouseID map[DeliverableID][]string
var missingCHSummaryEntriesByCRNServiceName map[string][]string

// ReadCHInputFile reads a file exported from ClearingHouse ("OSSCAT1" report)
func ReadCHInputFile(filename string) error {
	countEmptyPIDs := 0

	file, err := os.Open(filename) // #nosec G304
	if err != nil {
		return debug.WrapError(err, "Cannot open ClearingHouse input file %s", filename)
	}
	defer file.Close() // #nosec G307

	var imported = []*CHSummaryEntry{}
	gocsv.FailIfDoubleHeaderNames = true
	gocsv.FailIfUnmatchedStructTags = true
	err = gocsv.UnmarshalFile(file, &imported)
	if err != nil {
		return debug.WrapError(err, "Cannot parse ClearingHouse input file %s", filename)
	}

	allCHSummaryEntriesByPID = make(map[string][]*CHSummaryEntry)
	allCHSummaryEntriesByClearingHouseID = make(map[DeliverableID]*CHSummaryEntry)
	allCHSummaryEntriesByCRNServiceName = make(map[string][]*CHSummaryEntry)
	missingCHSummaryEntriesByPID = make(map[string][]string)
	missingCHSummaryEntriesByClearingHouseID = make(map[DeliverableID][]string)
	missingCHSummaryEntriesByCRNServiceName = make(map[string][]string)

	countRecords := 0
	for _, entry := range imported {
		countRecords++
		debug.Debug(debug.ClearingHouse, "Updating Record %3d: %+v", countRecords, entry)
		pids := strings.Split(entry.PIDs, ",")
		for _, pid := range pids {
			pid := ossrecord.NormalizeProductID(pid)
			//		debug.Debug(debug.ClearingHouse, "Updating Record %3d: PID:%s", countRecords, pid)
			if pid != "" {
				if elist, found := allCHSummaryEntriesByPID[pid]; found {
					elist = append(elist, entry)
					allCHSummaryEntriesByPID[pid] = elist
				} else {
					allCHSummaryEntriesByPID[pid] = []*CHSummaryEntry{entry}
				}
			} else {
				countEmptyPIDs++
				debug.Debug(debug.ClearingHouse, "Found ClearingHouse entry with empty PID: %+v", entry)
			}
		}
		if prior, found := allCHSummaryEntriesByClearingHouseID[entry.DeliverableID]; found {
			panic(fmt.Sprintf(`Found ClearingHouse entry with duplicate DeliverableID: new=%+v    prior=%+v`, entry, prior))
		}
		allCHSummaryEntriesByClearingHouseID[entry.DeliverableID] = entry
		entry.CRNServiceName = strings.TrimSpace(entry.CRNServiceName)
		if entry.CRNServiceName != "" {
			if elist, found := allCHSummaryEntriesByCRNServiceName[entry.CRNServiceName]; found {
				elist = append(elist, entry)
				allCHSummaryEntriesByCRNServiceName[entry.CRNServiceName] = elist
			} else {
				allCHSummaryEntriesByCRNServiceName[entry.CRNServiceName] = []*CHSummaryEntry{entry}
			}
		}
	}
	if countEmptyPIDs > 0 {
		debug.Warning("Found %d ClearingHouse entries with empty PIDs in input file %s", countEmptyPIDs, filename)

	}
	debug.Info("Completed reading the ClearingHouse input %s :  %d records\n", filename, countRecords)
	return nil
}

// LookupSummaryEntryByPID returns the CHSummaryEntry record(s) associated with a given PID
// The "hint" optional parameter is used only for debugging, to provide any information available about what we "think" the missing entry could have been
func LookupSummaryEntryByPID(pid string, hint string) ([]*CHSummaryEntry, bool) {
	e, ok := allCHSummaryEntriesByPID[pid]
	if !ok {
		hint0, _ := missingCHSummaryEntriesByPID[pid]
		missingCHSummaryEntriesByPID[pid], _ = collections.AppendSliceStringNoDups(hint0, hint)
	}
	return e, ok
}

// LookupSummaryEntryByID returns the CHSummaryEntry record associated with a given ClearingHouse Deliverable ID
// The "hint" optional parameter is used only for debugging, to provide any information available about what we "think" the missing entry could have been
func LookupSummaryEntryByID(chid DeliverableID, hint string) (*CHSummaryEntry, bool) {
	e, ok := allCHSummaryEntriesByClearingHouseID[chid]
	if !ok {
		hint0, _ := missingCHSummaryEntriesByClearingHouseID[chid]
		missingCHSummaryEntriesByClearingHouseID[chid], _ = collections.AppendSliceStringNoDups(hint0, hint)
	}
	return e, ok
}

// LookupSummaryEntryByCRNServiceName returns the CHSummaryEntry record(s) associated with a given CRN Service Name
// The "hint" optional parameter is used only for debugging, to provide any information available about what we "think" the missing entry could have been
func LookupSummaryEntryByCRNServiceName(crnServiceName string, hint string) ([]*CHSummaryEntry, bool) {
	e, ok := allCHSummaryEntriesByCRNServiceName[crnServiceName]
	if !ok {
		hint0, _ := missingCHSummaryEntriesByCRNServiceName[crnServiceName]
		missingCHSummaryEntriesByCRNServiceName[crnServiceName], _ = collections.AppendSliceStringNoDups(hint0, hint)
	}
	return e, ok
}

// HasCHInfo returns true if the table of ClearingHouse records has been initialized and loaded
func HasCHInfo() bool {
	return allCHSummaryEntriesByPID != nil
}

// ListSummaryEntries invokes the handler function on all ClearingHouse entries whose name matches the specified pattern
// (or all entries, if the pattern is empty)
func ListSummaryEntries(pattern *regexp.Regexp, handler func(ch *CHSummaryEntry)) error {
	for _, ch := range allCHSummaryEntriesByClearingHouseID {
		if pattern == nil || pattern.FindString(ch.Name) != "" {
			handler(ch)
		}
	}
	return nil
}

// DumpMissingCHSummaryEntries returns a multi-line buffer that lists
// all the entries that have been looked-up in this run of the program but were not found
func DumpMissingCHSummaryEntries() string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("    - %d missing entries by PID:\n", len(missingCHSummaryEntriesByPID)))
	pids := make([]string, 0, len(missingCHSummaryEntriesByPID))
	for pid := range missingCHSummaryEntriesByPID {
		pids = append(pids, pid)
	}
	sort.Strings(pids)
	for _, pid := range pids {
		buf.WriteString(fmt.Sprintf("        - %-40s    hint: %q\n", pid, missingCHSummaryEntriesByPID[pid]))
	}
	buf.WriteString(fmt.Sprintf("    - %d missing entries by ClearingHouse DeliverableID:\n", len(missingCHSummaryEntriesByClearingHouseID)))
	chids := make([]string, 0, len(missingCHSummaryEntriesByClearingHouseID))
	for chid := range missingCHSummaryEntriesByClearingHouseID {
		chids = append(chids, string(chid))
	}
	sort.Strings(chids)
	for _, chid := range chids {
		buf.WriteString(fmt.Sprintf("        - %-40s    hint: %q\n", chid, missingCHSummaryEntriesByClearingHouseID[DeliverableID(chid)]))
	}
	buf.WriteString(fmt.Sprintf("    - %d missing entries by CRN Service Name:\n", len(missingCHSummaryEntriesByCRNServiceName)))
	/*
		// No point listing missing CRN Service Names while most entries are actually missing
		crns := make([]string, 0, len(missingCHSummaryEntriesByCRNServiceName))
		for crn := range missingCHSummaryEntriesByCRNServiceName {
			crns = append(crns, string(crn))
		}
		sort.Strings(crns)
		for _, crn := range crns {
			buf.WriteString(fmt.Sprintf("        - %-40s    hint: %q\n", crn, missingCHSummaryEntriesByCRNServiceName[crn]))
		}
	*/
	return buf.String()
}

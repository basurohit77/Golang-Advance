package servicenow

// snsysyidfile contains the code for reading the csv file containing ServiceNow CI-sysid mappings

import (
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

type sysidTableType = map[ossrecord.CRNServiceName][]ossrecord.ServiceNowSysid

var sysidTables = make(map[ServiceNowDomain]sysidTableType)
var sysidTablesTest = make(map[ServiceNowDomain]sysidTableType)

func getSysidTable(testMode bool, snDomain ServiceNowDomain) sysidTableType {
	if testMode {
		sysidTable := sysidTablesTest[snDomain]
		if sysidTable == nil {
			sysidTable = make(sysidTableType)
			sysidTablesTest[snDomain] = sysidTable
		}
		return sysidTable
	}

	sysidTable := sysidTables[snDomain]
	if sysidTable == nil {
		sysidTable = make(sysidTableType)
		sysidTables[snDomain] = sysidTable
	}
	return sysidTable
}

// RecordSysid records one (name,sysid) pair in the table mapping service names to ServiceNow sysids
func RecordSysid(name ossrecord.CRNServiceName, sysid ossrecord.ServiceNowSysid, testMode bool, snDomain ServiceNowDomain) {
	t := getSysidTable(testMode, snDomain)
	t[name] = append(t[name], sysid)
}

// LookupServiceNowSysids returns the ServiceNow sysids associated with a given CRN service name
// There is usually only one sysid for each service name, but in some cases there may be more than one
// (e.g. if some of the ServiceNow entries are Retired)
func LookupServiceNowSysids(name ossrecord.CRNServiceName, testMode bool, snDomain ServiceNowDomain) ([]ossrecord.ServiceNowSysid, bool) {
	sysids, ok := getSysidTable(testMode, snDomain)[name]
	return sysids, ok
}

package ossreportsregistry

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/ossreports/allossentries"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/catalogga"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/cataloghidden"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/catalogsummary"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/catalogvisibilitygroups"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/clearinghousenames"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/dependencies"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/legacy"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/ownership"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/services4edb"
	"github.ibm.com/cloud-sre/osscatalog/ossreports/services4pnp"
	"github.ibm.com/cloud-sre/osscatalog/supportcenter/supportreport"
)

// Registry of all known reports

// ReportFunc represents the standard function to be invoked to run any report
type ReportFunc func(w io.Writer, timestamp string, pattern *regexp.Regexp) error

// Report represents one report in the global reports registry
type Report struct {
	Name          string
	FileExtension string
	ReportFunc    ReportFunc
}

// RegisterReport registers one report in the global reports registry
func RegisterReport(name string, fileExtension string, reportFunc ReportFunc) *Report {
	name = strings.TrimSpace(name)
	trimmed := strings.ToLower(name)
	if prior, found := reportsRegistry[trimmed]; found {
		panic(fmt.Sprintf("Duplicate report registration %s <-> %s", prior.Name, name))
	}
	r := &Report{
		Name:          name,
		FileExtension: fileExtension,
		ReportFunc:    reportFunc,
	}
	reportsRegistry[trimmed] = r
	return r
}

// LookupReport returns a report record associated with the given name, or nil if not found
func LookupReport(name string) *Report {
	trimmed := strings.ToLower(name)
	return reportsRegistry[trimmed]
}

var reportsRegistry = make(map[string]*Report)

func init() {
	RegisterReport("CatalogGA", "txt", catalogga.CatalogGA)
	RegisterReport("CatalogHidden", "txt", cataloghidden.CatalogHidden)
	RegisterReport("CatalogSummaryXL", "xlsx", catalogsummary.RunReport)
	RegisterReport("CatalogVisibilityGroups", "txt", catalogvisibilitygroups.CatalogVisibilityGroups)
	RegisterReport("ClearingHouseNames", "xlsx", clearinghousenames.ClearingHouseNames)
	RegisterReport("Dependencies", "xlsx", dependencies.Dependencies)
	RegisterReport("LegacyXL", "xlsx", legacy.RunReport)
	RegisterReport("Ownership", "txt", ownership.RunReport)
	RegisterReport("OwnershipXL", "xlsx", ownership.RunReportXL)
	RegisterReport("Services4PnP", "xlsx", services4pnp.Services4PnP)
	RegisterReport("Services4EDB", "xlsx", services4edb.Services4EDB)
	RegisterReport("AllOSSEntries", "xlsx", allossentries.RunReport)
	RegisterReport("SupportCenter", "xlsx", supportreport.RunReport)
}

package catalogapi

import (
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// IsIaaSRegionsPlaceholder returns true if this Catalog Resource is one of the special placeholder entries
// used to keep track of region/datacenter information for IaaS offerings
// In addition, if the "verbose" parameter is true, it generates a warning if it encounters a malformed record.
func (r *Resource) IsIaaSRegionsPlaceholder(verbose bool) bool {
	var hasTag bool
	for _, t := range r.Tags {
		if t == "iaasregions" {
			hasTag = true
			break
		}
	}

	if strings.Contains(r.Name, "placeholder") {
		if hasTag {
			return true
		}
		if verbose {
			debug.Warning(`Catalog entry name contains the word "placeholder" but does not have the "iaasregions" tag: %s`, r.String())
		}
	} else {
		if hasTag {
			debug.Warning(`Catalog entry has the "iaasregions" tag but the name does not contain the word "placeholder" : %s`, r.String())
		}
	}
	return false
}

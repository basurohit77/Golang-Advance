// Package ossrunactions tracks various options actions that can be taken during a merge.
package ossrunactions

import (
	"fmt"
	"sort"
	"strings"
)

// RunAction represents one optional action that can be taken during a merge
type RunAction struct {
	name    string
	parent  *RunAction
	enabled bool
}

var allValidRunActionsMap = make(map[string]*RunAction)
var allValidRunActionsNames = make([]string, 0, 20)
var allValidRunActionsList = make([]*RunAction, 0, 20)

// Declarations for all possible RunActions
// Note that we never sort the list of RunActions. It will appear in the output in the order that it was declared here
var (
	Services                  = newRunAction("Services", true, nil)                                // Refresh services/components OSS records (main function of osscatimporter)
	Tribes                    = newRunAction("Tribes", true, nil)                                  // Refresh segment/tribe OSS records
	Environments              = newRunAction("Environments", false, nil)                           // Refresh Environments (regions/datacenters/etc.) OSS records
	EnvironmentsNative        = newRunAction("Environments-Native", false, Environments)           // Refresh Environments (regions/datacenters/etc.) OSS records -- Native mode: do not create explicit OSS entry in Global Catalog if there is already a Main Catalog entry
	Deployments               = newRunAction("Deployments", false, nil)                            // Fetch and consolidate deployments information from services in Global Catalog
	Monitoring                = newRunAction("Monitoring", false, nil)                             // Fetch and consolidate information about Availability Monitoring metrics
	ProductInfoParts          = newRunAction("ProductInfo-Parts", false, nil)                      // Fetch pricing and part numbers information (add for new records only; copy for old records)
	ProductInfoPartsRefresh   = newRunAction("ProductInfo-Parts-Refresh", false, ProductInfoParts) // Refresh pricing and part numbers information for all records (old and new)
	ProductInfoClearingHouse  = newRunAction("ProductInfo-ClearingHouse", false, nil)              // Fetch ClearingHouse information (via CH export file)
	DependenciesClearingHouse = newRunAction("Dependencies-ClearingHouse", false, nil)             // Fetch dependency information from ClearingHouse (via CH API)
	ScorecardV1               = newRunAction("ScorecardV1", false, nil)                            // Load services information from ScorecardV1
	RMC                       = newRunAction("RMC", true, nil)                                     // Load information from RMC records, but only if the OSS entry is already known to have a RMC record
	RMCRescan                 = newRunAction("RMC-Rescan", false, RMC)                             // Load information from RMC records for all OSS entries (re-check if an OSS entry has a RMC record)
	Doctor                    = newRunAction("Doctor", false, nil)                                 // Load Segment/Tribe and Environment information from Doctor
	//	IAM                      = newRunAction("IAM", false, nil)                                    // Check IAM records
)

// newRunAction created a new RunAction and registers with the system
func newRunAction(name string, defaultEnabled bool, parent *RunAction) *RunAction {
	folded := strings.TrimSpace(strings.ToLower(name))
	if _, found := allValidRunActionsMap[folded]; found {
		panic(fmt.Sprintf("Found duplicate ossrunactions.RunAction: %s (%s)", name, folded))
	}
	ra := RunAction{name: name, enabled: defaultEnabled, parent: parent}
	allValidRunActionsMap[folded] = &ra
	allValidRunActionsMap[name] = &ra
	allValidRunActionsNames = append(allValidRunActionsNames, ra.Name())
	allValidRunActionsList = append(allValidRunActionsList, &ra)
	return &ra
}

// ParseRunAction  parses a string that represents a OSS RunAction and returns the actual RunAction
func ParseRunAction(s string) (ra *RunAction, ok bool) {
	folded := strings.TrimSpace(strings.ToLower(s))
	ra, ok = allValidRunActionsMap[folded]
	return ra, ok
}

// Enable takes an input or RunAction names,
// and records them all as enabled for this run of the merge tool
// If calling both Enable() and Disable() for the same actions, the last setting prevails.
func Enable(actions []string) error {
	var invalid []string
	var valid []*RunAction
	for _, s := range actions {
		if s == "" {
			continue
		}
		ra, ok := ParseRunAction(s)
		if ok {
			valid = append(valid, ra)
		} else {
			invalid = append(invalid, s)
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		return fmt.Errorf("Invalid RunAction(s): %q -- allowed values: %q", invalid, ListValidRunActionNames())
	}
	for _, ra := range valid {
		ra.enabled = true
	}
	return nil
}

// Disable takes an input list of RunAction names,
// and records them all as disabled for this run of the merge tool.
// If calling both Enable() and Disable() for the same actions, the last setting prevails.
func Disable(actions []string) error {
	var invalid []string
	var valid []*RunAction
	for _, s := range actions {
		if s == "" {
			continue
		}
		ra, ok := ParseRunAction(s)
		if ok {
			valid = append(valid, ra)
		} else {
			invalid = append(invalid, s)
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		return fmt.Errorf("Invalid RunAction(s): %q -- allowed values: %q", invalid, ListValidRunActionNames())
	}
	for _, ra := range valid {
		ra.enabled = false
	}
	return nil
}

// ListValidRunActions returns a list of all valid RunActions
func ListValidRunActions() []*RunAction {
	return allValidRunActionsList
}

// ListValidRunActionNames returns a list of all valid RunAction names
func ListValidRunActionNames() []string {
	return allValidRunActionsNames
}

// Name returns the name of this RunAcion
func (ra *RunAction) Name() string {
	return ra.name
}

// Parent returns the parent of this RunAcion, or nil if it has no parent
func (ra *RunAction) Parent() *RunAction {
	return ra.parent
}

// IsEnabled returns true if this RunAction is enabled for this run of the merge tool
func (ra *RunAction) IsEnabled() bool {
	return ra.enabled
}

// ListEnabledRunActionNames returns a list with the names of all the RunActions that are currently enabled
func ListEnabledRunActionNames() []string {
	result := make([]string, 0, len(allValidRunActionsList))
	for _, ra := range allValidRunActionsList {
		if ra.IsEnabled() {
			result = append(result, ra.Name())
		}
	}
	//	sort.Strings(result)
	return result
}

// ListDisabledRunActionNames returns a list with the names of all the RunActions that are currently disabled
func ListDisabledRunActionNames() []string {
	result := make([]string, 0, len(allValidRunActionsList))
	for _, ra := range allValidRunActionsList {
		if !ra.IsEnabled() {
			result = append(result, ra.Name())
		}
	}
	//	sort.Strings(result)
	return result
}

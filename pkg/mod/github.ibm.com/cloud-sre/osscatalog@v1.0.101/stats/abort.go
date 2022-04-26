package stats

import (
	"flag"
	"fmt"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
)

var thresholdPnpEnabledAdd = flag.Int("threshold-pnpEnabled-add", 5, "if the number of entries added to PnPEnabled list is exceeded, switch to read-only mode")
var thresholdPnpEnabledRemove = flag.Int("threshold-pnpEnabled-remove", 0, "if the number of entries removed from the PnPEnabled list is exceeded, switch to read-only mode")
var thresholdClientFacingAdd = flag.Int("threshold-clientFacing-add", 5, "if the number of entries added to ClientFacing list is exceeded, switch to read-only mode")
var thresholdClientFacingRemove = flag.Int("threshold-clientFacing-remove", 0, "if the number of entries removed from the ClientFacing list is exceeded, switch to read-only mode")
var thresholdTIPOnboardedAdd = flag.Int("threshold-tipOnboarded-add", 5, "if the number of entries added to TIPOnboarded list is exceeded, switch to read-only mode")
var thresholdTIPOnboardedRemove = flag.Int("threshold-tipOnboarded-remove", 0, "if the number of entries removed from the TIPOnboarded list is exceeded, switch to read-only mode")
var thresholdEntriesCreate = flag.Int("threshold-entries-create", 5, "if the number of new OSS entries is exceeded, switch to read-only mode")
var thresholdEntriesDelete = flag.Int("threshold-entries-delete", 0, "if the number of OSS entries deleted is exceeded, switch to read-only mode")
var thresholdErrors = flag.Int("threshold-errors", 0, "if the number of errors during the run is exceeded, switch to read-only mode")

// CheckForAbort checks the available stats to determine if the run should be aborted.
// It returns an abort message, or an empty string if the run does not have to be aborted
func (d *Data) CheckForAbort() string {
	allPnPEnabledAdded := len(d.AllPnPEnabledAdded)
	allPnPEnabledRemoved := len(d.AllPnPEnabledRemoved)
	allClientFacingAdded := len(d.AllClientFacingAdded)
	allClientFacingRemoved := len(d.AllClientFacingRemoved)
	allTIPOnboardedAdded := len(d.AllTIPOnboardedAdded)
	allTIPOnboardedRemoved := len(d.AllTIPOnboardedRemoved)
	allAdded := len(d.AllAdded)
	allRemoved := len(d.AllRemoved)
	allUnRetired := len(d.AllUnRetired)
	allNewRetired := len(d.AllNewRetired)
	abortMessage := strings.Builder{}
	if debug.CountCriticals() > 0 {
		abortMessage.WriteString(fmt.Sprintf("Encountered CRITICAL issues during run: %d\n", debug.CountCriticals()))
	}
	if debug.CountErrors() > *thresholdErrors {
		abortMessage.WriteString(fmt.Sprintf("Encountered ERRORS during run: %d / %d\n", debug.CountErrors(), *thresholdErrors))
	}
	if allPnPEnabledAdded > *thresholdPnpEnabledAdd {
		abortMessage.WriteString(fmt.Sprintf("OSS entries added to PnPEnabled threshold is exceeded: %d / %d\n", allPnPEnabledAdded, *thresholdPnpEnabledAdd))
	}
	if allPnPEnabledRemoved > *thresholdPnpEnabledRemove {
		abortMessage.WriteString(fmt.Sprintf("OSS entries removed from PnPEnabled threshold is exceeded: %d / %d\n", allPnPEnabledRemoved, *thresholdPnpEnabledRemove))
	}
	if allClientFacingAdded > *thresholdClientFacingAdd {
		abortMessage.WriteString(fmt.Sprintf("OSS entries added to ClientFacing threshold is exceeded: %d / %d\n", allClientFacingAdded, *thresholdClientFacingAdd))
	}
	if allClientFacingRemoved > *thresholdClientFacingRemove {
		abortMessage.WriteString(fmt.Sprintf("OSS entries removed from ClientFacing threshold is exceeded: %d / %d\n", allClientFacingRemoved, *thresholdClientFacingRemove))
	}
	if allTIPOnboardedAdded > *thresholdTIPOnboardedAdd {
		abortMessage.WriteString(fmt.Sprintf("OSS entries added to TIPOnboarded threshold is exceeded: %d / %d\n", allTIPOnboardedAdded, *thresholdTIPOnboardedAdd))
	}
	if allTIPOnboardedRemoved > *thresholdTIPOnboardedRemove {
		abortMessage.WriteString(fmt.Sprintf("OSS entries removed from TIPOnboarded threshold is exceeded: %d / %d\n", allTIPOnboardedRemoved, *thresholdTIPOnboardedRemove))
	}
	if allAdded > *thresholdEntriesCreate {
		abortMessage.WriteString(fmt.Sprintf("OSS entries created threshold is exceeded: %d / %d\n", allAdded, *thresholdEntriesCreate))
	}
	if allRemoved > *thresholdEntriesDelete {
		abortMessage.WriteString(fmt.Sprintf("OSS entries deleted threshold is exceeded: %d / %d\n", allRemoved, *thresholdEntriesDelete))
	}
	if allUnRetired > *thresholdEntriesCreate {
		abortMessage.WriteString(fmt.Sprintf("OSS entries un-Retired threshold is exceeded: %d / %d\n", allUnRetired, *thresholdEntriesCreate))
	}
	if allNewRetired > *thresholdEntriesDelete {
		abortMessage.WriteString(fmt.Sprintf("OSS entries newly Retired threshold is exceeded: %d / %d\n", allNewRetired, *thresholdEntriesDelete))
	}

	return abortMessage.String()
}

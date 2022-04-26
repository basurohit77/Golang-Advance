// Keeping track of various statistics from one run of the OSS tools
// (osscatimporter/osscatpublisher)

package stats

import (
	"fmt"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"

	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// actualStatsSingleton is a global object to keep statistics for the entire run of this program
var actualStatsSingleton *Data

// GetGlobalActualStats return the singleton object used to keep statistics for the entire run of this program
func GetGlobalActualStats() *Data {
	if actualStatsSingleton == nil {
		actualStatsSingleton = newData()
	}
	return actualStatsSingleton
}

// prepareStatsSingleton is a global object to keep statistics for the "prepare" phase of this program
// (prior to determining if any abort thresholds have been exceeded)
var prepareStatsSingleton *Data

// GetGlobalPrepareStats return the singleton object used to keep statistics for "prepare" phase of this program
// (prior to determining if any abort thresholds have been exceeded)
func GetGlobalPrepareStats() *Data {
	if prepareStatsSingleton == nil {
		prepareStatsSingleton = newData()
	}
	return prepareStatsSingleton
}

// Action represents the type of action (create/update/delete) performed on each record
type Action string

// Possible Action values
const (
	ActionCreate        Action = "CREATE"
	ActionUpdate        Action = "UPDATE"
	ActionDelete        Action = "DELETE"
	ActionNotModified   Action = "NOT-MODIFIED"
	ActionCatalogNative Action = "CATALOG-NATIVE"
	ActionError         Action = "ERROR"
	ActionLocked        Action = "LOCKED"
	actionIgnoreKey     Action = "Ignoring entry" // common prefix/key for all ActionIgnore instances
)

// ActionIgnore initializes a dummy action that indicates that a particular entry should be ignored
func ActionIgnore(str string) Action {
	return Action(fmt.Sprintf(`%s: %s`, actionIgnoreKey, str))
}

// IsActionIgnore return true if this particular action is an instance of the "ActionIgnore" class of actions
func (action Action) IsActionIgnore() bool {
	return strings.HasPrefix(string(action), string(actionIgnoreKey))
}

// Message produce a string message for each action, taking into account the ro/rw mode and force mode
func (action Action) Message(runMode options.RunMode, forceMode bool) string {
	switch action {
	case ActionCreate:
		if runMode != options.RunModeRW {
			return "Would create OSS entry (read-only mode)"
		} else if forceMode {
			return "Creating OSS entry (force mode)"
		} else {
			return "Creating OSS entry"
		}
	case ActionUpdate:
		if runMode != options.RunModeRW {
			return "Would update OSS entry (read-only mode)"
		} else if forceMode {
			return "Updating OSS entry (force mode)"
		} else {
			return "Updating OSS entry"
		}
	case ActionDelete:
		if runMode != options.RunModeRW {
			return "Would delete OSS entry (read-only mode)"
		}
		return "Deleting OSS entry"
	case ActionNotModified:
		return "Not updating OSS entry - no changes"
	case ActionCatalogNative:
		return "Not updating OSS entry - native Catalog entry"
	case ActionError:
		return "ERROR OCCURRED: Cannot update / create"
	case ActionLocked:
		return "OSS entry is locked"
	default:
		if action.IsActionIgnore() {
			return string(action)
		}
		panic(fmt.Sprintf("Unkown Action: %v", action))
	}
}

// Data tracks all the statistics from one run of a OSS tool
type Data struct {
	ActionCounts                     map[string]int
	NumServicesRaw                   int
	NumSegmentsRaw                   int
	NumTribesRaw                     int
	NumEnvironmentsRaw               int
	NumSchemaRaw                     int
	NumServicesActual                int
	NumSegmentsActual                int
	NumTribesActual                  int
	NumEnvironmentsActual            int
	NumSchemaActual                  int
	AllAdded                         []string
	AllRemoved                       []string
	AllNewRetired                    []string
	AllUnRetired                     []string
	NumPnPEnabled                    int
	AllPnPEnabledAdded               []string
	AllPnPEnabledRemoved             []string
	AllClientFacingAdded             []string
	AllClientFacingRemoved           []string
	AllTIPOnboardedAdded             []string
	AllTIPOnboardedRemoved           []string
	ServicesOneCloudValidationStatus map[osstags.Tag]*StatusCounts
	numReadOnlyActions               int // internal
	numReadWriteActions              int // internal
}

// newData allocates a new Data record to track all statistics
func newData() *Data {
	d := new(Data)
	d.ActionCounts = make(map[string]int)
	d.ServicesOneCloudValidationStatus = make(map[osstags.Tag]*StatusCounts)
	return d
}

type targetTypeType int

const (
	typeService targetTypeType = iota
	typeSegment
	typeTribe
	typeEnvironment
	typeOSSResourceClassification
)

// RecordAction records one Action performed on a given target OSS entry
func (d *Data) RecordAction(action Action, runMode options.RunMode, forceMode bool, target ossrecord.OSSEntry, prior ossrecord.OSSEntry) {
	if target == nil {
		return
	}

	var targetType targetTypeType

	// Update the counts of OSS entry types
	switch target.(type) {
	case *ossrecord.OSSService, *ossrecordextended.OSSServiceExtended:
		d.NumServicesRaw++
		targetType = typeService
	case *ossrecord.OSSSegment, *ossrecordextended.OSSSegmentExtended:
		d.NumSegmentsRaw++
		targetType = typeSegment
	case *ossrecord.OSSTribe, *ossrecordextended.OSSTribeExtended:
		d.NumTribesRaw++
		targetType = typeTribe
	case *ossrecord.OSSEnvironment, *ossrecordextended.OSSEnvironmentExtended:
		d.NumEnvironmentsRaw++
		targetType = typeEnvironment
	case *ossrecord.OSSResourceClassification:
		d.NumSchemaRaw++
		targetType = typeOSSResourceClassification
	default:
		panic(fmt.Sprintf(`RecordAction("%s", "%s", "%v"): Unexpected entry type %T`, action, target.String(), prior, target))
	}

	// Update the counts of Actions
	if action.IsActionIgnore() {
		count := d.ActionCounts[string(actionIgnoreKey)]
		d.ActionCounts[string(actionIgnoreKey)] = count + 1
	} else {
		actionString := action.Message(runMode, forceMode)
		count := d.ActionCounts[actionString]
		d.ActionCounts[actionString] = count + 1
	}

	// Update the lists of entries added/removed
	var pnpBefore, pnpAfter bool
	var clientFacingBefore, clientFacingAfter bool
	var tipBefore, tipAfter bool
	var activeBefore, activeAfter bool
	switch action {
	case ActionCreate:
		/* FIXME: see https://github.ibm.com/cloud-sre/osscatalog/issues/122
		if prior != nil {
			panic(fmt.Sprintf(`RecordAction("%s", "%s", "%v"): Unexpected non-nil prior entry`, action, target.String(), prior))
		}
		*/
		d.AllAdded = append(d.AllAdded, target.String())
		pnpBefore = false
		clientFacingBefore = false
		tipBefore = false
		pnpAfter = target.GetOSSTags().Contains(osstags.PnPEnabled)
		clientFacingAfter, tipAfter, activeAfter = getTargetAttributes(target)
		activeBefore = activeAfter // Change will be captured as a true "added" event
		if runMode != options.RunModeRW {
			d.numReadOnlyActions++
		} else {
			d.numReadWriteActions++
		}
	case ActionDelete:
		if prior == nil {
			panic(fmt.Sprintf(`RecordAction("%s", "%s", "%v"): Unexpected nil prior entry`, action, target.String(), prior))
		}
		d.AllRemoved = append(d.AllRemoved, target.String())
		pnpBefore = prior.GetOSSTags().Contains(osstags.PnPEnabled)
		clientFacingBefore, tipBefore, activeBefore = getTargetAttributes(prior)
		pnpAfter = false
		clientFacingAfter = false
		tipAfter = false
		activeAfter = activeBefore // Change will be captured as a true "removed" event
		if runMode != options.RunModeRW {
			d.numReadOnlyActions++
		} else {
			d.numReadWriteActions++
		}
	case ActionUpdate:
		if prior == nil {
			panic(fmt.Sprintf(`RecordAction("%s", "%s", "%v"): Unexpected nil prior entry`, action, target.String(), prior))
		}
		pnpBefore = prior.GetOSSTags().Contains(osstags.PnPEnabled)
		clientFacingBefore, tipBefore, activeBefore = getTargetAttributes(prior)
		pnpAfter = target.GetOSSTags().Contains(osstags.PnPEnabled)
		clientFacingAfter, tipAfter, activeAfter = getTargetAttributes(target)
		if runMode != options.RunModeRW {
			d.numReadOnlyActions++
		} else {
			d.numReadWriteActions++
		}
	case ActionNotModified:
		pnpBefore = prior.GetOSSTags().Contains(osstags.PnPEnabled)
		clientFacingBefore, tipBefore, activeBefore = getTargetAttributes(prior)
		pnpAfter = target.GetOSSTags().Contains(osstags.PnPEnabled)
		clientFacingAfter, tipAfter, activeAfter = getTargetAttributes(target)
	case ActionLocked:
		// Ignore
	case ActionCatalogNative, ActionError:
		// XXX ignore
	}

	// Check for inconsistencies read-only/read-write
	if d.numReadOnlyActions > 0 && d.numReadWriteActions > 0 {
		panic(fmt.Sprintf(`RecordAction("%s", "%s", "%v"): got a mix of read-only and read-write actions`, action, target.String(), prior))
	}

	// Update PnP and other counts
	if targetType == typeService {
		if pnpAfter {
			d.NumPnPEnabled++
		}
		switch {
		case pnpBefore && !pnpAfter:
			d.AllPnPEnabledRemoved = append(d.AllPnPEnabledRemoved, target.String())
		case !pnpBefore && pnpAfter:
			d.AllPnPEnabledAdded = append(d.AllPnPEnabledAdded, target.String())
		}
		switch {
		case clientFacingBefore && !clientFacingAfter:
			d.AllClientFacingRemoved = append(d.AllClientFacingRemoved, target.String())
		case !clientFacingBefore && clientFacingAfter:
			d.AllClientFacingAdded = append(d.AllClientFacingAdded, target.String())
		}
		switch {
		case tipBefore && !tipAfter:
			d.AllTIPOnboardedRemoved = append(d.AllTIPOnboardedRemoved, target.String())
		case !tipBefore && tipAfter:
			d.AllTIPOnboardedAdded = append(d.AllTIPOnboardedAdded, target.String())
		}
	}
	switch {
	case activeBefore && !activeAfter:
		d.AllNewRetired = append(d.AllNewRetired, target.String())
	case !activeBefore && activeAfter:
		d.AllUnRetired = append(d.AllUnRetired, target.String())
	}

	// Update actual entry counts and status counts
	switch action {
	case ActionCreate, ActionUpdate, ActionNotModified, ActionLocked:
		switch targetType {
		case typeService:
			d.NumServicesActual++
			oneCloudTag := target.GetOSSTags().GetTagByGroup(osstags.GroupOneCloud)
			if oneCloudTag == "" {
				oneCloudTag = "~not_onecloud"
			}
			statusCounts := d.ServicesOneCloudValidationStatus[oneCloudTag]
			if statusCounts == nil {
				statusCounts = new(StatusCounts)
				d.ServicesOneCloudValidationStatus[oneCloudTag] = statusCounts
			}
			statusCounts.updateStatus(target)
		case typeSegment:
			d.NumSegmentsActual++
		case typeTribe:
			d.NumTribesActual++
		case typeEnvironment:
			d.NumEnvironmentsActual++
		case typeOSSResourceClassification:
			d.NumSchemaActual++
		}
	case ActionDelete, ActionError:
		// ignore
	case ActionCatalogNative:
		if targetType != typeEnvironment {
			panic(fmt.Sprintf(`RecordAction("%s", "%s", "%v"): expected OSSEnvironment entry but got %T`, action, target.String(), prior, target))
		}
		d.NumEnvironmentsActual++ // TODO: should we count CatalogNative records separately
	}
}

// NumEntries returns the number of OSS entries for which an action was recorded
func (d *Data) NumEntries() int {
	return d.NumServicesRaw + d.NumSegmentsRaw + d.NumTribesRaw + d.NumEnvironmentsRaw + d.NumSchemaRaw
}

// Finalize performs any necessary computations that need to happen after all Actions have been recorded in this Data object
// (e.g. compute totals)
func (d *Data) Finalize() {
	const totalKey osstags.Tag = "ALL"

	if _, ok := d.ServicesOneCloudValidationStatus[totalKey]; !ok {
		total := new(StatusCounts)
		for _, sc := range d.ServicesOneCloudValidationStatus {
			total.add(sc)
		}
		d.ServicesOneCloudValidationStatus[totalKey] = total
	}
}

// Report generates a report (as a string) containing all the key statistics
func (d *Data) Report(prefix string) string {
	d.Finalize()
	keys := make([]osstags.Tag, 0, len(d.ServicesOneCloudValidationStatus))
	for k := range d.ServicesOneCloudValidationStatus {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("%s%-52s: services/components:%-4d  segments:%-4d  tribes:%-4d  environments:%-4d  schema:%-4d\n", prefix, "Raw number of OSS records processed", d.NumServicesRaw, d.NumSegmentsRaw, d.NumTribesRaw, d.NumEnvironmentsRaw, d.NumSchemaRaw))
	buf.WriteString(fmt.Sprintf("%s%-52s: services/components:%-4d  segments:%-4d  tribes:%-4d  environments:%-4d  schema:%-4d\n", prefix, "Final number of valid OSS records", d.NumServicesActual, d.NumSegmentsActual, d.NumTribesActual, d.NumEnvironmentsActual, d.NumSchemaActual))
	for _, k := range keys {
		sc := d.ServicesOneCloudValidationStatus[k]
		buf.WriteString(fmt.Sprintf("%s%-52s: Total:%-4d    CRN:Green:%-4d        CRN:Yellow:%-4d        CRN:Red:%-4d        CRN:Unknown:%-4d\n", prefix, fmt.Sprintf("CRN validation status for %s services", k), (sc.CRN.Green + sc.CRN.Yellow + sc.CRN.Red + sc.CRN.Unknown), sc.CRN.Green, sc.CRN.Yellow, sc.CRN.Red, sc.CRN.Unknown))
	}
	for _, k := range keys {
		sc := d.ServicesOneCloudValidationStatus[k]
		buf.WriteString(fmt.Sprintf("%s%-52s: Total:%-4d    Overall:Green:%-4d    Overall:Yellow:%-4d    Overall:Red:%-4d    Overall:Unknown:%-4d\n", prefix, fmt.Sprintf("Overall validation status for %s services", k), (sc.Overall.Green + sc.Overall.Yellow + sc.Overall.Red + sc.Overall.Unknown), sc.Overall.Green, sc.Overall.Yellow, sc.Overall.Red, sc.Overall.Unknown))
	}
	buf.WriteString(fmt.Sprintf("%s%-52s: %d\n", prefix, "Total OSS entries that are PnPEnabled", d.NumPnPEnabled))
	buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries added to PnPEnabled list", len(d.AllPnPEnabledAdded), d.AllPnPEnabledAdded))
	buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries removed from PnPEnabled list", len(d.AllPnPEnabledRemoved), d.AllPnPEnabledRemoved))
	buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries added to ClientFacing list", len(d.AllClientFacingAdded), d.AllClientFacingAdded))
	buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries removed from ClientFacing list", len(d.AllClientFacingRemoved), d.AllClientFacingRemoved))
	buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries added to TIPOnboarded list", len(d.AllTIPOnboardedAdded), d.AllTIPOnboardedAdded))
	buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries removed from TIPOnboarded list", len(d.AllTIPOnboardedRemoved), d.AllTIPOnboardedRemoved))
	actions := make([]string, 0, len(d.ActionCounts))
	for a := range d.ActionCounts {
		actions = append(actions, a)
	}
	sort.Slice(actions, func(i, j int) bool {
		return actions[i] < actions[j]
	})
	for _, a := range actions {
		buf.WriteString(fmt.Sprintf("%s%-52s: %d\n", prefix, "Action: "+a, d.ActionCounts[a]))
	}
	if d.numReadOnlyActions > 0 {
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries created (read-only mode)", len(d.AllAdded), d.AllAdded))
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries deleted (read-only mode)", len(d.AllRemoved), d.AllRemoved))
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries un-Retired (read-only mode)", len(d.AllUnRetired), d.AllUnRetired))
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries newly Retired (read-only mode)", len(d.AllNewRetired), d.AllNewRetired))
	} else {
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries created", len(d.AllAdded), d.AllAdded))
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries deleted", len(d.AllRemoved), d.AllRemoved))
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries un-Retired", len(d.AllUnRetired), d.AllUnRetired))
		buf.WriteString(fmt.Sprintf("%s%-52s: %-4d  %v\n", prefix, "OSS entries newly Retired", len(d.AllNewRetired), d.AllNewRetired))
	}
	buf.WriteString(fmt.Sprintf("%s%-52s: %v\n", prefix, "Total warnings", debug.CountWarnings()))
	buf.WriteString(fmt.Sprintf("%s%-52s: %v\n", prefix, "Total errors", debug.CountErrors()))
	buf.WriteString(fmt.Sprintf("%s%-52s: %v\n", prefix, "Total critical issues ***", debug.CountCriticals()))
	return buf.String()
}

func getTargetAttributes(target ossrecord.OSSEntry) (clientFacing, tip, active bool) {
	switch t1 := target.(type) {
	case *ossrecord.OSSService:
		clientFacing = t1.GeneralInfo.ClientFacing
		tip = t1.Operations.TIPOnboarded
		active = (t1.GeneralInfo.OperationalStatus != ossrecord.RETIRED)
	case *ossrecordextended.OSSServiceExtended:
		clientFacing = t1.OSSService.GeneralInfo.ClientFacing
		tip = t1.OSSService.Operations.TIPOnboarded
		active = (t1.OSSService.GeneralInfo.OperationalStatus != ossrecord.RETIRED)
	case *ossrecord.OSSEnvironment:
		clientFacing = false
		tip = false
		active = (t1.Status != ossrecord.EnvironmentDecommissioned)
	case *ossrecordextended.OSSEnvironmentExtended:
		clientFacing = false
		tip = false
		active = (t1.OSSEnvironment.Status != ossrecord.EnvironmentDecommissioned)
	default:
		clientFacing = false
		tip = false
		active = true
	}
	return clientFacing, tip, active
}

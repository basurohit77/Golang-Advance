package ossmerge

// Functions for merging data structures

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// MergeStruct copies all the matching fields found in the "from" structure into the "to" structure
func MergeStruct(to interface{}, from interface{}) {
	buffer, err := json.Marshal(from)
	if err != nil {
		panic(debug.WrapError(err, "MergeStruct() Cannot Marshal struct: %v", from))
	}
	err = json.Unmarshal(buffer, to)
	if err != nil {
		panic(debug.WrapError(err, "MergeStruct() Cannot Unmarshal struct: %v into %v", from, to))
	}
}

type parameter interface {
	isAMergeParameter()
}

// RMC is a merge parameter used to pass a value read from a RMC record (main record)
type RMC struct {
	V interface{}
}

func (p RMC) isAMergeParameter() {}

// RMCOSS is a merge parameter used to pass a value read from the Operations (OSS) tab of a RMC record
type RMCOSS struct {
	V interface{}
}

func (p RMCOSS) isAMergeParameter() {}

// Catalog is a merge parameter used to pass a value read from a Catalog main record
type Catalog struct {
	V interface{}
}

func (p Catalog) isAMergeParameter() {}

// Custom is a merge parameter used to pass a custom value - must specify a source name explicitly, and there is no check if that source is valid
// This can be used, for example, to pass computed values derived indirectly from the non-existence of a Catalog entry (Catalog shadow)
type Custom struct {
	N string
	V interface{}
}

func (p Custom) isAMergeParameter() {}

// PriorOSS is a merge parameter used to pass a value read from a prior OSS record
type PriorOSS struct {
	V interface{}
}

func (p PriorOSS) isAMergeParameter() {}

// ServiceNow is a merge parameter used to pass a value read from a ServiceNow record
type ServiceNow struct {
	V interface{}
}

func (p ServiceNow) isAMergeParameter() {}

// ScorecardV1 is a merge parameter used to pass a value read from a ScorecardV1 record
type ScorecardV1 struct {
	V interface{}
}

func (p ScorecardV1) isAMergeParameter() {}

// OverrideProperty is a merge parameter used to pass a value read from a OSSMergeControl.Overrides entry
type OverrideProperty struct {
	N string
}

func (p OverrideProperty) isAMergeParameter() {}

// OverrideTag is a merge parameter used to pass a value read from a OSSMergeControl.OSSTags entry
type OverrideTag struct {
	N osstags.Tag
	V interface{}
}

func (p OverrideTag) isAMergeParameter() {}

// SeverityIfMissing is a merge parameter used to specify the severity of the ValidationIssue generated if the value from a particular source is missing
type SeverityIfMissing struct {
	V ossvalidation.Severity
}

func (p SeverityIfMissing) isAMergeParameter() {}

// TagIfMissing is a merge parameter used to used to specify a tag for the ValidationIssue generated if the value from a particular source is missing
type TagIfMissing struct {
	V ossvalidation.Tag
}

func (p TagIfMissing) isAMergeParameter() {}

// SeverityIfMismatch is a merge parameter used to specify the severity of the ValidationIssue generated if there is a value mismatch between multiple sources
type SeverityIfMismatch struct {
	V ossvalidation.Severity
}

func (p SeverityIfMismatch) isAMergeParameter() {}

// TagIfMismatch is a merge parameter used to specify the severity of the ValidationIssue generated if there is a value mismatch between multiple sources
type TagIfMismatch struct {
	V ossvalidation.Tag
}

func (p TagIfMismatch) isAMergeParameter() {}

// MergeValues merges values from multiple sources according to variable struct parameters
// of type Catalog{}, ServiceNow{}, ScorecardV1, SeverityIfMissing{}, SeverityIfMismatch{}, etc.
func (si *ServiceInfo) MergeValues(item string, params ...parameter) interface{} {
	var severityIfMissing = ossvalidation.WARNING
	var severityIfMismatch = ossvalidation.WARNING
	var tagsIfMissing = []ossvalidation.Tag{ossvalidation.TagDataMissing}
	var tagsIfMismatch = []ossvalidation.Tag{ossvalidation.TagDataMismatch}
	var returnType reflect.Type
	var returnTypeSource string
	var zero interface{}
	var mergedVal interface{}
	var mergedValSource string
	var numPotentialSources = 0
	var numValidSources = 0
	var allParams = make(map[string]bool)
	var haveSeenPriorOSS = false
	var haveSeenRMCOSS = false
	var haveScorecardV1Disabled = false
	var haveRMCDisabled = false
	var haveRMCOSSDisabled = false
	for _, p := range params {
		if p == nil {
			continue
		}
		pName := reflect.TypeOf(p).Name()
		if _, found := allParams[pName]; found {
			panic(fmt.Sprintf("ossmerge.MergeValues(): duplicate parameter of the same type %T", p))
		}
		allParams[pName] = true
		var sourceName string
		var sourceVal interface{}
		var isValid = true
		var isOverride = false
		var isPriorOSS = false
		switch pp := p.(type) {
		case SeverityIfMissing:
			severityIfMissing = pp.V
			continue
		case TagIfMissing:
			tagsIfMissing = append(tagsIfMissing, pp.V)
			continue
		case SeverityIfMismatch:
			severityIfMismatch = pp.V
			continue
		case TagIfMismatch:
			tagsIfMismatch = append(tagsIfMismatch, pp.V)
			continue
		case ScorecardV1:
			if ossrunactions.ScorecardV1.IsEnabled() {
				if si.HasSourceScorecardV1Detail() {
					numValidSources++
				} else {
					isValid = false
				}
				sourceName = "ScorecardV1"
				sourceVal = pp.V
				numPotentialSources++
			} else {
				haveScorecardV1Disabled = true
				continue
			}
		case ServiceNow:
			if si.HasSourceServiceNow() {
				numValidSources++
			} else {
				isValid = false
			}
			sourceName = "ServiceNow"
			sourceVal = pp.V
			numPotentialSources++
		case Catalog:
			if si.HasSourceMainCatalog() {
				numValidSources++
			} else {
				isValid = false
			}
			sourceName = "Catalog"
			sourceVal = pp.V
			numPotentialSources++
		case RMC:
			if ossrunactions.RMC.IsEnabled() {
				if si.HasSourceRMC() {
					numValidSources++
				} else {
					isValid = false
				}
				sourceName = "RMC"
				sourceVal = pp.V
				numPotentialSources++
			} else {
				haveRMCDisabled = true
				continue
			}
		case RMCOSS:
			if si.HasPriorOSS() && si.GetPriorOSS().GeneralInfo.OSSOnboardingPhase != "" {
				numValidSources++
			} else {
				haveRMCOSSDisabled = true
				isValid = false
			}
			sourceName = "RMCOSS"
			sourceVal = pp.V
			numPotentialSources++
			haveSeenRMCOSS = true
		case Custom:
			// Do not bother checking the validity of the source - assume it is valid
			sourceName = pp.N
			sourceVal = pp.V
			numPotentialSources++
			numValidSources++
		case PriorOSS:
			// This complex logic is a "catch-all" to get a last-change value from the Prior OSS record, if we cannot find it from anywhere else.
			// In a nutshell:
			//   - if the list of MergeParameters include RMCOSS, we always will get a prior OSS value (either through the RMCOSS parameter itself, or through this catch-all)
			//   - if the list of MergeParameters does not include RMCOSS but the OSSOnboardingPhase is not empty (meaning someone has been editing the record in RMC),
			//     we ill get a prior OSS value through this catch-all
			// As a table:
			// |                | with RMCOSS                 | without RMCOSS             |
			// |----------------+-----------------------------+----------------------------|
			// | oss=raw        | use (haveRMCOSSDisabled)    | ignore (no conditions met) |
			// | oss=onboarding | use RMCOSS not PriorOSS     | use (!haveseenRMCOSS)      |
			//
			// In more detail (based on other MergeParameters present):
			// | Input MergeParameters         | PriorOSS Current Impl                                    |
			// |-------------------------------+----------------------------------------------------------|
			// | other=notzero                 | ignore (numValidSources>0)                               |
			// | other=zero                    | ignore (numValidSources>0)                               |
			// | other=missing; oss=onboarding | use (!haveSeenRMCOSS)                                    |
			// | other=missing; oss=raw        | ignore (no conditions met)                               |
			// | RMCOSS=onboarding             | ignore (numValidSources>0); get value from RMCOSS itself |
			// | RMCOSS=raw                    | use (haveRMCOSSDisabled)                                 |
			// | none                          | use (numPotentialSources==0)                             |
			if si.HasPriorOSS() && numValidSources == 0 && (numPotentialSources == 0 || haveScorecardV1Disabled || haveRMCDisabled || haveRMCOSSDisabled || (si.HasPriorOSS() && si.GetPriorOSS().GeneralInfo.OSSOnboardingPhase != "" && !haveSeenRMCOSS)) {
				numValidSources++ // XXX Do we really want to count this?
			} else {
				isValid = false
				// Ignore PriorOSS
			}
			sourceName = "prior OSS record"
			sourceVal = pp.V
			haveSeenPriorOSS = true
			isPriorOSS = true
		case OverrideProperty:
			if si.OSSMergeControl == nil {
				isValid = false
				continue
			}
			val, err := si.OSSMergeControl.GetOverride(pp.N)
			if err != nil {
				panic(fmt.Sprintf("ossmerge.MergeValues(): error getting OverrideProperty(%s): %v", pp.N, err))
			}
			if val == nil {
				isValid = false
				continue
			}
			isOverride = true
			sourceName = "OSSMergeControl.Overrides." + pp.N
			sourceVal = val
			numPotentialSources++
			numValidSources++
		case OverrideTag:
			if si.OSSMergeControl == nil || !si.OSSMergeControl.OSSTags.Contains(pp.N) {
				isValid = false
				continue
			}
			if pp.V == nil {
				panic(fmt.Sprintf("ossmerge.MergeValues(): error getting OverrideTag(%s): %v", pp.N, pp.V))
			}
			isOverride = true
			sourceName = "OSSMergeControl.OSSTags." + string(pp.N)
			sourceVal = pp.V
			numPotentialSources++
			numValidSources++
		default:
			panic(fmt.Sprintf("ossmerge.MergeValues(): unknown parameter type %T", p))
		}
		if sourceVal == nil {
			panic(fmt.Sprintf("ossmerge.MergeValues(): missing value for parameter %s", sourceName))
			// TODO: Try to validate that the value is actually a field within the corresponding source record (or derived from it)
		}
		if returnType == nil {
			returnType = reflect.TypeOf(sourceVal)
			returnTypeSource = sourceName
			zero = reflect.Zero(returnType).Interface()
		} else if reflect.TypeOf(sourceVal) != returnType {
			panic(fmt.Sprintf("ossmerge.MergeValues() attempting to merge two values of different types:  %s=%s   %s=%T",
				returnTypeSource, returnType.Name(), sourceName, sourceVal))
		}
		if !isValid {
			continue
		}
		if !isPriorOSS && haveSeenPriorOSS {
			panic(fmt.Sprintf("ossmerge.MergeValues(): PriorOSS parameter is not in last position (found %T after PriorOSS)", p))
		}
		if returnType.Kind() == reflect.Bool || isOverride || !compareValues(sourceVal, zero) {
			if mergedVal == nil {
				mergedVal = sourceVal
				mergedValSource = sourceName
				if isOverride {
					si.AddValidationIssue(ossvalidation.MINOR, fmt.Sprintf("%s value overriden with OSS MergeControl record", item),
						`%s="%v"`, mergedValSource, mergedVal).TagControlOverride()
				}
				/*
					if isPriorOSS {
						// DONE: remove this validation issue in Production
						si.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("%s value taken from PriorOSS record", item),
							`%s="%v"`, mergedValSource, mergedVal).TagPriorOSS()
					}
				*/
			} else if !compareValues(mergedVal, sourceVal) {
				// TODO: Avoid generating multiple "mismatch" ValidationIssues if there are more than 2 sources
				si.AddValidationIssue(severityIfMismatch, fmt.Sprintf("%s has different values from different sources (first source prevails)", item),
					`%s="%v"   %s="%v"`, mergedValSource, mergedVal, sourceName, sourceVal).AddTag(tagsIfMismatch...)
			}
		} else {
			si.AddValidationIssue(severityIfMissing, fmt.Sprintf("%s is missing from %s", item, sourceName), "").AddTag(tagsIfMissing...)
		}
	}

	if !haveSeenPriorOSS {
		panic(fmt.Sprintf("ossmerge.MergeValues(): no PriorOSS parameter specified"))
	}
	if severityIfMissing == "" {
		panic(fmt.Sprintf("ossmerge.MergeValues(): SeverityIfMissing is not set"))
	}
	if severityIfMismatch == "" {
		panic(fmt.Sprintf("ossmerge.MergeValues(): SeverityIfMismatch is not set"))
	}
	if returnType == nil {
		panic(fmt.Sprintf("ossmerge.MergeValues(): No sources specified (returnType==nil)"))
	}

	if mergedVal == nil {
		if numValidSources > 0 {
			si.AddValidationIssue(severityIfMissing, fmt.Sprintf("%s cannot be set in OSS record from any available source", item), "").AddTag(tagsIfMissing...)
		} else {
			si.AddValidationIssue(ossvalidation.IGNORE, fmt.Sprintf("%s cannot be set in OSS record because there are no source records containing this attribute", item), "").AddTag(tagsIfMissing...)
		}
		// TODO: Could "zero" might be uninitialized if we found no valid sources at all?
		// TODO: Generate a validation warning if an OverrideProperty or OverrideTag is unnecessary because all main sources already match it
		return zero
	}
	return mergedVal
}

// compareValues compares two values represented as interfaces,
// including support for comparing osstags.TagSet and string arrays
func compareValues(a, b interface{}) bool {
	switch aa := a.(type) {
	case osstags.TagSet:
		bb := b.(osstags.TagSet)
		if len(aa) != len(bb) {
			return false
		}
		for _, t := range aa {
			if !bb.Contains(t) {
				return false
			}
		}
		return true
	case []string:
		bb := b.([]string)
		if len(aa) != len(bb) {
			return false
		}
		for _, ar := range aa {
			found := false
			for _, br := range bb {
				if br == ar {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// newStringNoDups obtains a value for a new string, comparing it against an old value
// to ensure that the old value is either empty or equal to the new value.
// It return a boolean indicator if there is a conflict.
func newStringNoDups(prior, val string) (newVal string, isDup bool) {
	prior = strings.TrimSpace(prior)
	val = strings.TrimSpace(val)
	if val == "" {
		return prior, false
	}
	if prior == "" {
		return val, false
	}
	if val != prior {
		return prior, true
	}
	return prior, false
}

// DeepCopy performs a deep copy of a struct
// It relies on the struct having valid json tags for all members to be copied
func DeepCopy(dest, src interface{}) error {
	buffer, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buffer, dest)
	return err
}

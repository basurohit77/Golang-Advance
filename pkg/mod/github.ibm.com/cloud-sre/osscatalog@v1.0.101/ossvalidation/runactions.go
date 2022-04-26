package ossvalidation

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
)

// RunActionToTag returns the Validation Tag associated with this RunAction
func RunActionToTag(ra *ossrunactions.RunAction) Tag {
	for ; ra.Parent() != nil; ra = ra.Parent() {
	}
	return Tag(fmt.Sprintf("%s", ra.Name()))
}

// RecordRunAction records the last time that a given RunAction was executed in this OSSValidation record
func (ossv *OSSValidation) RecordRunAction(ra *ossrunactions.RunAction) {
	//ossv.LastRunActions[ra.Name()] = "<current>"
	// The LastRunTimestamp for this entire OSSValidation counts as the last timestamp for this RunAction
	// We do not update the timestamp to avoid causing extra rewrites of the OSS records when there are no other changes
}

// CopyRunAction copies all the ValidationIssues for a given RunAction from a prior OSSValidation record
// into this one
func (ossv *OSSValidation) CopyRunAction(prior *OSSValidation, ra *ossrunactions.RunAction) {
	if prior != nil {
		issues := prior.GetIssues([]Tag{RunActionToTag(ra)})
		for _, v := range issues {
			ossv.AddIssuePreallocated(v)
		}
		if ossv.LastRunActions == nil {
			ossv.LastRunActions = make(map[string]string)
		}
		if lastRun, found := prior.LastRunActions[ra.Name()]; found {
			ossv.LastRunActions[ra.Name()] = lastRun
		} else if prior.LastRunTimestamp != "" {
			ossv.LastRunActions[ra.Name()] = prior.LastRunTimestamp
		} else {
			ossv.LastRunActions[ra.Name()] = "<never>"
		}
	} else {
		if ossv.LastRunActions == nil {
			ossv.LastRunActions = make(map[string]string)
		}
		ossv.LastRunActions[ra.Name()] = "<never>"
	}
}

// TagRunAction is a utility function to add the tag associated with a given RunAction to a ValidationIssue
func (v *ValidationIssue) TagRunAction(ra *ossrunactions.RunAction) *ValidationIssue {
	return v.AddTag(RunActionToTag(ra))
}

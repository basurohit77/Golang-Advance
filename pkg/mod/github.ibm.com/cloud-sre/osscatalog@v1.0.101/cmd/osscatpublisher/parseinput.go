package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"

	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

type input struct {
	OSSService                *ossrecord.OSSService                `json:"oss_service,omitempty"`
	OSSMergeControl           *ossmergecontrol.OSSMergeControl     `json:"oss_merge_control,omitempty"`
	OSSValidation             *ossvalidation.OSSValidation         `json:"oss_validation,omitempty"`
	OSSSegment                *ossrecord.OSSSegment                `json:"oss_segment,omitempty"`
	OSSTribe                  *ossrecord.OSSTribe                  `json:"oss_tribe,omitempty"`
	OSSEnvironment            *ossrecord.OSSEnvironment            `json:"oss_environment,omitempty"`
	OSSResourceClassification *ossrecord.OSSResourceClassification `json:"oss_resource_classification,omitempty"`
}

func (in *input) ossEntry() ossrecord.OSSEntry {
	return nil
}

func parseJSONInput(reader io.Reader, pattern *regexp.Regexp, handler func(r ossrecord.OSSEntry)) (numEntries int, err error) {
	dec := json.NewDecoder(reader)
	token, err := dec.Token()
	if err != nil {
		return 0, debug.WrapError(err, "Error reading initial token from input file")
	}
	if token0, ok := token.(json.Delim); !ok || token0.String() != `[` {
		return 0, fmt.Errorf(`Expected initial token "[" in input file, got "%#v"`, token)
	}
	for dec.More() {
		var e input
		err := dec.Decode(&e)
		if err != nil {
			details := parseJSONError(err)
			return 0, debug.WrapError(err, "Error reading entry from input file: %s", details)
		}
		switch {
		case e.OSSService != nil:
			if pattern.FindString(string(e.OSSService.ReferenceResourceName)) != "" {
				if *stagingOnly {
					handler(&ossrecordextended.OSSServiceExtended{
						OSSService:      *e.OSSService,
						OSSMergeControl: e.OSSMergeControl,
						OSSValidation:   e.OSSValidation,
					})
				} else {
					handler(e.OSSService)
				}
				numEntries++
			}
		case e.OSSSegment != nil:
			if pattern.FindString(string(e.OSSSegment.DisplayName)) != "" {
				if *stagingOnly {
					handler(&ossrecordextended.OSSSegmentExtended{
						OSSSegment:    *e.OSSSegment,
						OSSValidation: e.OSSValidation,
					})
				} else {
					handler(e.OSSSegment)
				}
				numEntries++
			}
		case e.OSSTribe != nil:
			if pattern.FindString(string(e.OSSTribe.DisplayName)) != "" {
				if *stagingOnly {
					handler(&ossrecordextended.OSSTribeExtended{
						OSSTribe:      *e.OSSTribe,
						OSSValidation: e.OSSValidation,
					})
				} else {
					handler(e.OSSTribe)
				}
				numEntries++
			}
		case e.OSSEnvironment != nil:
			if pattern.FindString(string(e.OSSEnvironment.EnvironmentID)) != "" {
				if *stagingOnly {
					handler(&ossrecordextended.OSSEnvironmentExtended{
						OSSEnvironment: *e.OSSEnvironment,
						OSSValidation:  e.OSSValidation,
					})
				} else {
					handler(e.OSSEnvironment)
				}
				numEntries++
			}
		case e.OSSResourceClassification != nil:
			handler(e.OSSResourceClassification)
			numEntries++
		default:
			if options.GlobalOptions().Lenient {
				debug.Warning("(Lenient Mode): Ignoring empty source entry in inputfile: %+v", e)
			} else {
				debug.PrintError("Ignoring empty source entry in inputfile: %+v", e)
			}
		}
	}
	token, err = dec.Token()
	if err != nil {
		return numEntries, debug.WrapError(err, "Error reading final token from input file")
	}
	if token0, ok := token.(json.Delim); !ok || token0.String() != `]` {
		return numEntries, fmt.Errorf(`Expected final token "]" in input file, got "%#v"`, token)
	}

	return numEntries, nil
}

func parseJSONError(err error) string {
	switch err1 := err.(type) {
	case *json.SyntaxError:
		return fmt.Sprintf("json.SyntaxError near offset %d", err1.Offset)
	case *json.UnmarshalTypeError:
		return fmt.Sprintf("json.UnmarshalTypeError near offset %d", err1.Offset)
	default:
		return fmt.Sprintf("%T (offset unknown)", err)
	}
}

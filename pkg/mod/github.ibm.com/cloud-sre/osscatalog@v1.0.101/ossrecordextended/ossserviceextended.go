package ossrecordextended

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// OSSServiceExtended contains all the information being stored in each OSSService record in the Catalog
// (including information that is involved in the merge operation but not really part of the exported OSS data)
type OSSServiceExtended struct {
	ossrecord.OSSService                                  // Main OSSService entry
	OSSMergeControl      *ossmergecontrol.OSSMergeControl // Control information for the merge
	OSSValidation        *ossvalidation.OSSValidation     // Validation output from the merge
	Created              string                           // Creation time of the OSS entry
	Updated              string                           // Last update time of the OSS entry
}

var _ ossrecord.OSSEntry = &OSSServiceExtended{} // Check interface definition

// NewOSSServiceExtended creates an empty OSSServiceExtended
func NewOSSServiceExtended(name ossrecord.CRNServiceName) *OSSServiceExtended {
	result := OSSServiceExtended{}
	result.Created = time.Now().String()
	result.Updated = time.Now().String()
	result.OSSService.SchemaVersion = ossrecord.OSSCurrentSchema
	result.OSSService.ReferenceResourceName = name
	result.OSSMergeControl = ossmergecontrol.New(string(name))
	result.OSSValidation = ossvalidation.New(string(name), options.GlobalOptions().LogTimeStamp)
	return &result
}

// CheckConsistency verifies that the several sub-records in this OSSServiceExtended are internally consistent
// (for example, that they all reference the same resource name)
// It returns an error if it finds an issue, or panics if the "panicIfError" flag is true
func (ossrec *OSSServiceExtended) CheckConsistency(panicIfError bool) error {
	var ok = true
	var err error
	if ossrec.OSSMergeControl != nil && string(ossrec.OSSService.ReferenceResourceName) != ossrec.OSSMergeControl.CanonicalName {
		ok = false
	}
	if ossrec.OSSValidation != nil && string(ossrec.OSSService.ReferenceResourceName) != ossrec.OSSValidation.CanonicalName {
		ok = false
	}
	if !ok {
		switch {
		case ossrec.OSSMergeControl != nil && ossrec.OSSValidation != nil:
			err = fmt.Errorf(`Attempting to combine OSS record "%s" with OSSMergeControl for "%s" and OSSValidation for "%s"`, ossrec.OSSService.String(), ossrec.OSSMergeControl.CanonicalName, ossrec.OSSValidation.CanonicalName)
		case ossrec.OSSMergeControl != nil && ossrec.OSSValidation == nil:
			err = fmt.Errorf(`Attempting to combine OSS record "%s" with OSSMergeControl for "%s" (OSSValidation not included)`, ossrec.OSSService.String(), ossrec.OSSMergeControl.CanonicalName)
		case ossrec.OSSMergeControl == nil && ossrec.OSSValidation != nil:
			err = fmt.Errorf(`Attempting to combine OSS record "%s" with OSSValidation for "%s" (OSSMergeControl not included)`, ossrec.OSSService.String(), ossrec.OSSValidation.CanonicalName)
		default:
			err = fmt.Errorf(`CheckConsistency() Unexpected error OSS record "%s" (neither OSSMergeControl nor OSSValidation included)`, ossrec.OSSService.String())
		}
	}
	if err != nil {
		if panicIfError {
			panic(err)
		} else {
			return err
		}
	}
	return nil
}

/*
// String returns a short string representation of this OSSServiceExtended record
func (ossrec *OSSServiceExtended) String() string {
	/*
		var hasOSSMergeControl = "nil"
		if ossrec.OSSMergeControl != nil {
			hasOSSMergeControl = string(ossrec.OSSMergeControl.CanonicalName)
		}
		var hasOSSValidation = "nil"
		if ossrec.OSSValidation != nil {
			hasOSSValidation = string(ossrec.OSSValidation.CanonicalName)
		}
		return fmt.Sprintf(`OSSServiceExtended{ OSS: %s, OSSMergeControl: %s, OSSValidation: %s }`, ossrec.OSSService.ReferenceResourceName, hasOSSMergeControl, hasOSSValidation)
	//END INNER COMMENT

	return ossrec.OSSService.String()
}
*/

// JSON returns a JSON-style string representation of the entire OSSServiceExtended record (with no header), multi-line
func (ossrec *OSSServiceExtended) JSON() string {
	var result strings.Builder
	result.WriteString("  {\n")
	result.WriteString(`    "oss_service": `)
	json1, _ := json.MarshalIndent(ossrec.OSSService, "    ", "    ")
	_, err := result.Write(json1)
	if err != nil {
		panic(err)
	}
	if ossrec.OSSMergeControl != nil {
		result.WriteString(",\n")
		result.WriteString(`    "oss_merge_control": `)
		json2, _ := json.MarshalIndent(ossrec.OSSMergeControl, "    ", "    ")
		_, err := result.Write(json2)
		if err != nil {
			panic(err)
		}
	}
	if ossrec.OSSValidation != nil {
		result.WriteString(",\n")
		result.WriteString(`    "oss_validation": `)
		json2, _ := json.MarshalIndent(ossrec.OSSValidation, "    ", "    ")
		_, err := result.Write(json2)
		if err != nil {
			panic(err)
		}
	}
	result.WriteString("\n  }")
	return result.String()
}

// SetTimes sets the Created and Updated times of this OSSEntry
func (ossrec *OSSServiceExtended) SetTimes(created string, updated string) {
	ossrec.Created = created
	ossrec.Updated = updated
}

// IsDeletable returns true if this OSS record is a candidate for deletion (i.e it has no source, and no non-zero OSSMergeControl or RMC data)
func (ossrec *OSSServiceExtended) IsDeletable() bool {
	if ossrec.OSSMergeControl.OSSTags.Contains(osstags.OSSOnly) ||
		ossrec.OSSMergeControl.OSSTags.Contains(osstags.OSSTest) ||
		(ossrec.OSSService.GeneralInfo.OSSOnboardingPhase != "" /* && ossrec.OSSService.GeneralInfo.OSSOnboardingPhase != ossrecord.INVALID XXX do not delete invalid RMC entries for now */) ||
		ossrec.OSSMergeControl.OSSTags.Contains(osstags.TypeIAMOnly) {
		return false
	}
	trueSources := ossrec.OSSValidation.TrueSources()
	if len(trueSources) == 0 {
		return true
	}
	if ossrec.GeneralInfo.EntryType == ossrecord.IAMONLY && len(trueSources) <= 2 {
		// OK to delete IAMONLY but only if there are no other sources beside a Catalog and RMC entry
		for _, s := range trueSources {
			switch s {
			case ossvalidation.CATALOG, ossvalidation.RMC, ossvalidation.CATALOGIGNORED:
				continue
			default:
				return false
			}
		}
		return true
	}
	return false
}

// CleanEntryForCompare cleans the content of an OSSEntry so that it can be compared with other entries
// without flagging irrelevant differences
func (ossrec *OSSServiceExtended) CleanEntryForCompare() {
	ossrec.Created = ""
	ossrec.Updated = ""
	ossrec.OSSMergeControl = nil
	ossrec.OSSValidation = nil
}

// ResetForRMC resets the validation info (status and issues) that might become obsolete when an entry is being edited in RMC
func (ossrec *OSSServiceExtended) ResetForRMC() {
	ossrec.OSSService.ResetForRMC()
	if ossrec.OSSValidation != nil {
		ossrec.OSSValidation.ResetForRMC()
	}
}

package ossrecordextended

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// OSSEnvironmentExtended contains all the information being stored in each OSSEnvironment record in the Catalog
// (including information that is involved in the merge operation but not really part of the exported OSS data)
type OSSEnvironmentExtended struct {
	ossrecord.OSSEnvironment                              // Main OSSEnvironment entry
	OSSValidation            *ossvalidation.OSSValidation // Validation output from the merge
	Created                  string                       // Creation time of the OSS entry
	Updated                  string                       // Last update time of the OSS entry
}

var _ ossrecord.OSSEntry = &OSSEnvironmentExtended{} // Check interface definition

// NewOSSEnvironmentExtended creates an empty OSSEnvironmentExtended
func NewOSSEnvironmentExtended(id ossrecord.EnvironmentID) *OSSEnvironmentExtended {
	result := OSSEnvironmentExtended{}
	result.OSSEnvironment.SchemaVersion = ossrecord.OSSCurrentSchema
	result.OSSEnvironment.EnvironmentID = id
	result.OSSValidation = ossvalidation.New(string(id), options.GlobalOptions().LogTimeStamp)
	result.Created = time.Now().String()
	result.Updated = time.Now().String()
	return &result
}

// CheckConsistency verifies that the several sub-records in this OSSEnvironmentExtended are internally consistent
// (for example, that they all reference the same OSSEntry ID)
// It returns an error if it finds an issue, or panics if the "panicIfError" flag is true
func (ossrec *OSSEnvironmentExtended) CheckConsistency(panicIfError bool) error {
	var ok = true
	var err error
	if ossrec.OSSValidation != nil && string(ossrec.OSSEnvironment.EnvironmentID) != ossrec.OSSValidation.CanonicalName {
		ok = false
	}
	if !ok {
		switch {
		case ossrec.OSSValidation != nil:
			err = fmt.Errorf(`Attempting to combine OSS record "%s" with OSSValidation for "%s" (full record=%s)`, ossrec.OSSEnvironment.String(), ossrec.OSSValidation.CanonicalName, ossrec.OSSEnvironment.JSON())
		default:
			err = fmt.Errorf(`CheckConsistency() Unexpected error in OSS record "%s" (neither OSSMergeControl nor OSSValidation included) (full record=%s)`, ossrec.OSSEnvironment.String(), ossrec.OSSEnvironment.JSON())
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

// CleanEntryForCompare cleans the content of an OSSEntry so that it can be compared with other entries
// without flagging irrelevant differences
func (ossrec *OSSEnvironmentExtended) CleanEntryForCompare() {
	ossrec.Created = ""
	ossrec.Updated = ""
	ossrec.OSSValidation = nil
}

// JSON returns a JSON-style string representation of the entire OSSEnvironmentExtended record (with no header), multi-line
func (ossrec *OSSEnvironmentExtended) JSON() string {
	var result strings.Builder
	result.WriteString("  {\n")
	result.WriteString(`    "oss_environment": `)
	json1, _ := json.MarshalIndent(ossrec.OSSEnvironment, "    ", "    ")
	_, err := result.Write(json1)
	if err != nil {
		panic(err)
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
func (ossrec *OSSEnvironmentExtended) SetTimes(created string, updated string) {
	ossrec.Created = created
	ossrec.Updated = updated
}

// ResetForRMC resets the validation info (status and issues) that might become obsolete when an entry is being edited in RMC
func (ossrec *OSSEnvironmentExtended) ResetForRMC() {
	ossrec.OSSEnvironment.ResetForRMC()
	if ossrec.OSSValidation != nil {
		ossrec.OSSValidation.ResetForRMC()
	}
}

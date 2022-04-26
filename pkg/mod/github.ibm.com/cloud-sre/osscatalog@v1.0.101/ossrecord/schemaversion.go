package ossrecord

import (
	"fmt"

	goversion "github.com/hashicorp/go-version"
	"github.ibm.com/cloud-sre/osscatalog/debug"
)

// OSSCurrentSchema is an identifier for the current version of the schema used to represent OSS records in this library
const OSSCurrentSchema = "1.0.14"

// NullSchemaVersionError indicates that CheckSchemaVersion encountered a null schema version in a OSS entry
// We use a distinguished error to allow the caller to ignore this particular, if desired
type NullSchemaVersionError struct {
	error
}

var currentSchemaVersion *goversion.Version

func init() {
	var err error
	currentSchemaVersion, err = goversion.NewVersion(OSSCurrentSchema)
	if err != nil {
		panic(debug.WrapError(err, "OSSCurrentSchema(%s) has invalid format", OSSCurrentSchema))
	}
}

func checkSchemaVersion(entry OSSEntry, version *string) error {
	if *version == "" {
		return &NullSchemaVersionError{fmt.Errorf("OSS Resource %s has empty Schema version", entry.String())}
	}
	parsed, err := goversion.NewVersion(*version)
	if err != nil {
		return debug.WrapError(err, "OSS Resource %s has invalid format Schema version (%s)", entry.String(), *version)
	}
	if parsed.GreaterThan(currentSchemaVersion) {
		err := fmt.Errorf("OSS Resource %s has incompatible Schema version: current library OSSSchema %q / got %q", entry.String(), OSSCurrentSchema, *version)
		*version = fmt.Sprintf(`Mismatch{record:%q  library:%q}`, *version, OSSCurrentSchema)
		return err
	}
	return nil
}

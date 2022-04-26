package doctor

import (
	"fmt"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// EnvironmentEntry is the data structure returned from the
// https://cloud-oss-metadata.bluemix.net/cloud-oss/metadata/allEnv in Doctor
type EnvironmentEntry struct {
	CRN        string `json:"crn"`
	Deployment string `json:"deployment"`
	Doctor     string `json:"doctor"`
	EnvName    string `json:"env_name"`
	Monitor    string `json:"monitor"`
	NewCRN     string `json:"new_crn"`
	RegionID   string `json:"region_id"`
	UCD        string `json:"ucd"`
}

// String returns a short string representation of this Doctor EnvironmentEntry
func (e *EnvironmentEntry) String() string {
	return fmt.Sprintf("DoctorEnvironment(%s/%s[%s]", e.EnvName, e.Doctor, e.NewCRN)
}

// ListEnvironments lists all Environment entries registered in Doctor
// and calls the special handler function for each entry
func ListEnvironments(pattern *regexp.Regexp, handler func(e *EnvironmentEntry)) error {
	rawEntries := 0
	totalEntries := 0

	actualURL := doctorAllEnvURL

	// TODO: Fix TLS for Doctor access
	oldTLS := rest.SetDisableTLSVerify(true)
	defer rest.SetDisableTLSVerify(oldTLS)

	var result []*EnvironmentEntry
	err := rest.DoHTTPGet(actualURL, "", nil, "Doctor environments", debug.Doctor, &result)
	if err != nil {
		return err
	}

	if len(result) == 0 {
		err = fmt.Errorf("Doctor environments GET: empty result  (URL=%s)", actualURL)
		return err
	}

	for _, e := range result {
		rawEntries++
		if (rawEntries % 30) == 0 {
			debug.Info("Loading one batch of environment entries from Doctor (%d/%d entries so far)", totalEntries, rawEntries)
		}
		if pattern != nil && pattern.FindString(e.Doctor) == "" {
			continue
		}
		handler(e)
		totalEntries++
	}

	debug.Info("Read %d environment entries from Doctor (rawEntries=%d)", totalEntries, rawEntries)
	return nil
}

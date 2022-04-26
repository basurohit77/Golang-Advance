package supportcenter

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
)

var candidatesInputFile *os.File

// SetCandidatesInputFile defines the input file for loading Support Center candidates
func SetCandidatesInputFile(fname string) error {
	var err error
	candidatesInputFile, err = os.Open(fname) // #nosec G304
	if err != nil {
		return debug.WrapError(err, "Cannot open Support Center candidates input file %s", fname)
	}
	return err
}

// HasCandidatesInputFile returns true if an input file for loading Support Center candidates has been specified
func HasCandidatesInputFile() bool {
	return candidatesInputFile != nil
}

// Candidate represents the information from one entry in the Support Center candidates input file
type Candidate struct {
	CRNServiceName string `json:"crn_service_name"`
	DisplayName    string `json:"display_name"`
	Parent         string `json:"parent"`
}

// LoadSupportCenterCandidates loads a ServiceInfo entry for each recored in the list of candidates for the Support Center
func LoadSupportCenterCandidates(registry ossmergemodel.ModelRegistry) error {
	if candidatesInputFile == nil {
		panic("supportcenter.LoadSupportCenterCandidates() called but no input file specified")
	}
	defer candidatesInputFile.Close() // #nosec G307
	count := 0
	rawData, err := ioutil.ReadAll(candidatesInputFile)
	if err != nil {
		return debug.WrapError(err, "Error reading Support Center candidates input file %v", candidatesInputFile)
	}
	var candidates []*Candidate
	err = json.Unmarshal(rawData, &candidates)
	if err != nil {
		return debug.WrapError(err, "Error parsing Support Center candidates input file %v", candidatesInputFile)
	}
	for _, e := range candidates {
		si, _ := LookupService(registry, e.CRNServiceName, true)
		sci := si.GetSupportCenterInfo(NewSupportCenterInfo)
		if sci.Candidate == nil {
			sci.Candidate = e
			count++
		} else {
			debug.PrintError(`LoadSupportCenterCandidates() found duplicate entry "%s" (model="%s")`, e.CRNServiceName, si.Model.String())
		}
	}

	debug.Info("Completed reading the Support Center candidates file %s: %d entries", candidatesInputFile.Name(), count)
	return nil
}

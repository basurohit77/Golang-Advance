package iam

import (
	"fmt"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// Service represents one service entry in IAM
type Service struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Enabled     bool   `json:"enabled"`
	// Lots of other attributes that we are not interested in
}

type servicesGet struct {
	Services []*Service `json:"services"`
}

// String returns a short string reprentation of this IAM service entry
func (e *Service) String() string {
	return fmt.Sprintf(`IAM.Service{Name:"%s", Enabled:%v}`, e.Name, e.Enabled)

}

// ListIAMServices lists all services registered in IAM and calls the special handler function for each entry
func ListIAMServices(pattern *regexp.Regexp, handler func(e *Service)) error {
	rawEntries := 0
	totalEntries := 0

	actualURL := IAMServicesURL

	token, err := rest.GetToken(IAMServicesKeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get IAM token for IAM")
		return err
	}

	var result = new(servicesGet)
	debug.Info("Loading one batch of entries from IAM (%d/%d entries so far)", totalEntries, rawEntries)
	err = rest.DoHTTPGet(actualURL, token, nil, "IAM", debug.IAM, result)
	if err != nil {
		return err
	}

	/*
		if result.Result != "success" {
			err = fmt.Errorf("Scorecard GET: result=%s  (URL=%s)", result.Result, actualURL)
			return err
		}
	*/
	rawEntries += len(result.Services)

	for i := 0; i < len(result.Services); i++ {
		rawEntries++
		if (rawEntries % 30) == 0 {
			debug.Info("Loading one batch of entries from IAM (%d/%d entries so far)", totalEntries, rawEntries)
		}
		e := result.Services[i]
		if pattern.FindString(e.Name) == "" {
			continue
		}
		handler(e)
		totalEntries++
	}

	debug.Info("Read %d entries from IAM", totalEntries)
	return nil
}

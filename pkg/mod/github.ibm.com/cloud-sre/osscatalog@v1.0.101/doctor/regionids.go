package doctor

import (
	"fmt"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// RegionID is the data structure returned from the
// https://api-oss.bluemix.net/doctorapi/api/doctor/regionids in Doctor
type RegionID struct {
	CRN            string `json:"crn"`
	Decommissioned string `json:"decommissioned"`
	ID             string `json:"id"`
	Name           string `json:"name"`
}

type regionIDGet struct {
	Resources []*RegionID `json:"resources"`
}

// String returns a short string representation of this Doctor RegionID
func (r *RegionID) String() string {
	return fmt.Sprintf("DoctorRegionID(%s[%s]/mccpid=%s)", r.Name, r.CRN, r.ID)
}

// ListRegionIDs lists all RegionID entries registered in Doctor
// and calls the special handler function for each entry
func ListRegionIDs(pattern *regexp.Regexp, handler func(e *RegionID)) error {
	rawEntries := 0
	totalEntries := 0

	actualURL := doctorRegionIDURL

	key, err := rest.GetKey(doctorKeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for Doctor")
		return err
	}

	// TODO: Fix TLS for Doctor access
	oldTLS := rest.SetDisableTLSVerify(true)
	defer rest.SetDisableTLSVerify(oldTLS)

	var result regionIDGet
	err = rest.DoHTTPGet(actualURL, key, nil, "Doctor RegionIDs", debug.Doctor, &result)
	if err != nil {
		return err
	}

	if len(result.Resources) == 0 {
		err = fmt.Errorf("Doctor RegionIDs GET: empty result  (URL=%s)", actualURL)
		return err
	}

	for _, e := range result.Resources {
		rawEntries++
		if (rawEntries % 30) == 0 {
			debug.Info("Loading one batch of regionID entries from Doctor (%d/%d entries so far)", totalEntries, rawEntries)
		}
		if pattern != nil && pattern.FindString(e.Name) == "" {
			continue
		}
		handler(e)
		totalEntries++
	}

	debug.Info("Read %d regionID entries from Doctor (rawEntries=%d)", totalEntries, rawEntries)
	return nil
}

// CRNMask returns the CRN mask associated with a given RegionID record.
// It accounts for the fact that the "CRN" attribute might be empty, and attempts to construct a CRN mask from the MCCP ID and other info if necessary
func (r *RegionID) CRNMask() (crn.Mask, error) {
	if r.CRN != "" {
		ret, err := crn.Parse(r.CRN)
		if err != nil {
			return crn.Mask{}, err
		}
		return ret, nil
	}
	comps := strings.Split(r.ID, ":")
	if len(comps) != 3 || comps[0] == "" || comps[2] == "" {
		return crn.Mask{}, fmt.Errorf(`Cannot construct CRN for Doctor RegionID: CRN attribute is empty and MCCPID not of the form "<cname>:<something>:<location>"  -- input=%s`, r.String())
	}
	ret := crn.Mask{
		CName:    comps[0],
		CType:    "dedicated", // XXX can it really only be dedicated?
		Location: comps[2],
	}
	return ret, nil
}

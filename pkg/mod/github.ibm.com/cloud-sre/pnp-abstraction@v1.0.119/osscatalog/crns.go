package osscatalog

import (
	"log"
	"regexp"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
)

// CatalogCheckBypass allows bypassing of some of the functions to avoid
// calls to OSS catalog. This is helpful to bypass for test scenarios
var CatalogCheckBypass = false

// IsCRNPnPEnabled is an all encompassing function that will examine
// any URL and indicate if the environment plus service name equate
// to a PNP enabled resource
func IsCRNPnPEnabled(crn string) (bool, error) {

	if CatalogCheckBypass {
		return true, nil
	}

	ok, err := IsServicePnPEnabled(GetServiceFromCRN(crn))
	if err != nil {
		log.Println(crn, "is not service PnP enabled due to error:", err)
		return false, err
	}

	ok = ok && IsEnvPnPEnabled(crn)

	return ok, nil
}

// IsServicePnPEnabled will indicate if the given service name is PnP Enabled
// If an error occurs, false will be returned with the error
func IsServicePnPEnabled(service string) (bool, error) {

	if CatalogCheckBypass {
		return true, nil
	}

	if service == "" { // Not sure what this service is
		return false, nil
	}

	r, err := ServiceNameToOSSRecord(ctxt.Context{}, service)
	if err != nil {
		return false, err
	}

	return r.GeneralInfo.OSSTags.Contains(osstags.PnPEnabled), nil

}

// IsGaaSCRN will indicate if the passed CRN represents a GaaS resource
func IsGaaSCRN(crn string) (bool, error) {

	if CatalogCheckBypass {
		return false, nil
	}

	// Note there is not concept of a GaaS environment right now,
	// so we just check the service.
	return IsGaaSService(GetServiceFromCRN(crn))
}

// IsGaaSService will indicate if the provided service is considered a GaaS service
func IsGaaSService(service string) (bool, error) {

	if CatalogCheckBypass {
		return false, nil
	}

	if service == "" { // Not sure what this service is
		return false, nil
	}

	r, err := ServiceNameToOSSRecord(ctxt.Context{}, service)
	if err != nil {
		return false, err
	}

	return r.GeneralInfo.EntryType == ossrecord.GAAS, nil

}

// GetServiceFromCRN will return the service-name included in the provide
// CRN. If there is no service name or the CRN is formatted incorrectly
// then the empty string is returned.
func GetServiceFromCRN(crn string) string {
	FCT := "GetServiceFromCRN"

	bits := strings.Split(crn, ":")
	if len(bits) != 10 {
		log.Printf("%s: Bad CRN: %s", FCT, crn)
		return ""
	}
	return bits[4]
}

// IsValidCRNFormat merely indicates if the provided crn appears to
// be in a valid format.  It does not check the validity of the actual
// data provide in the CRN.
func IsValidCRNFormat(crn string) bool {
	FCT := "IsValidCRNFormat"

	// First validate that the crn meets all the basic character requirements
	matched, err := regexp.MatchString(`^[a-z0-9:-]+$`, crn)
	if err != nil {
		log.Println(FCT, "Error matching crn", err)
		return false
	}
	if !matched {
		log.Println(FCT, "CRN has invalid characters")
		return false
	}

	bits := strings.Split(crn, ":")
	if len(bits) != 10 {
		log.Println(FCT, "CRN does not have enough colons")
		return false
	}

	if bits[0] != "crn" || bits[1] != "v1" {
		log.Println(FCT, "CRN does not start correctly. Must be crn:v1")
		return false
	}

	return true
}

// ParsedCRN is a handy struct to name the segments of a CRN
type ParsedCRN struct {
	Version         string
	Cname           string
	Ctype           string
	ServiceName     string
	Location        string
	Scope           string
	ServiceInstance string
	ResourceType    string
	Resource        string
}

// ParseCRN will parse the provided CRN into its component parts
// If the CRN cannot be parsed for any reason or the CRN format
// looks invalid, then nil is returned
func ParseCRN(crn string) *ParsedCRN {

	if !IsValidCRNFormat(crn) {
		return nil
	}

	bits := strings.Split(crn, ":")

	return &ParsedCRN{Version: bits[1], Cname: bits[2], Ctype: bits[3], ServiceName: bits[4], Location: bits[5], Scope: bits[6], ServiceInstance: bits[7], ResourceType: bits[8], Resource: bits[9]}

}

// String will create a crn string from a parsed CRN
func (pc *ParsedCRN) String() string {
	return "crn:" + pc.Version + ":" + pc.Cname + ":" + pc.Ctype + ":" + pc.ServiceName + ":" + pc.Location + ":" + pc.Scope + ":" + pc.ServiceInstance + ":" + pc.ResourceType + ":" + pc.Resource
}

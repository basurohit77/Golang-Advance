package ossrecord

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var validEntryTypes = map[string]EntryType{
	"SERVICE":            SERVICE,
	"RUNTIME":            RUNTIME,
	"TEMPLATE":           TEMPLATE,
	"IAAS":               IAAS,
	"PLATFORM_COMPONENT": PLATFORMCOMPONENT,
	"SUBCOMPONENT":       SUBCOMPONENT,
	"SUPERCOMPONENT":     SUPERCOMPONENT,
	"COMPOSITE":          COMPOSITE,
	"CONTENT":            CONTENT,
	"CONSULTING":         CONSULTING,
	"SERVICE/VMWARE":     VMWARE,
	"GAAS":               GAAS,
	"OTHEROSS":           OTHEROSS,
	"IAM_ONLY":           IAMONLY,
	"INTERNAL":           INTERNALSERVICE,
}

// ParseEntryType converts a string into a valid EntryType value, or returns an error if the string does not map
// to a known EntryType
func ParseEntryType(input string) (EntryType, error) {
	if result, ok := validEntryTypes[input]; ok {
		return result, nil
	}
	return "", fmt.Errorf(`Invalid string value for ossrecord.EntryType: "%s"`, input)
}

var validOperationalStatuses = map[string]OperationalStatus{
	"GA":                  GA,
	"BETA":                BETA,
	"EXPERIMENTAL":        EXPERIMENTAL,
	"THIRDPARTY":          THIRDPARTY,
	"COMMUNITY":           COMMUNITY,
	"DEPRECATED":          DEPRECATED,
	"SELECTAVAILABILITY":  SELECTAVAILABILITY,
	"LIMITEDAVAILABILITY": SELECTAVAILABILITY, // For backward compatibility
	"LIMITEDAVAILABILIY":  SELECTAVAILABILITY, // For backward compatibility
	"RETIRED":             RETIRED,
	"INTERNAL":            INTERNAL,
	"NOTREADY":            NOTREADY,
	"<unknown>":           OperationalStatusUnknown,
}

// ParseOperationalStatus converts a string into a valid OperationalStatus value, or returns an error if the string does not map
// to a known OperationalStatus
func ParseOperationalStatus(input string) (OperationalStatus, error) {
	if result, ok := validOperationalStatuses[input]; ok {
		return result, nil
	}
	return "", fmt.Errorf(`Invalid string value for ossrecord.OperationalStatus: "%s"`, input)
}

var validTier2EscalationTypes = map[string]Tier2EscalationType{
	"SERVICENOW":      SERVICENOW,
	"GITHUB":          GITHUB,
	"RTC":             RTC,
	"OTHERESCALATION": OTHERESCALATION,
}

// ParseTier2EscalationType converts a string into a valid Tier2EscalationType value, or returns an error if the string does not map
// to a known Tier2EscalationType
func ParseTier2EscalationType(input string) (Tier2EscalationType, error) {
	if result, ok := validTier2EscalationTypes[input]; ok {
		return result, nil
	}
	return "", fmt.Errorf(`Invalid string value for ossrecord.Tier2EscalationType: "%s"`, input)
}

var validClientExperiences = map[string]ClientExperience{
	"ACS_SUPPORTED":   ACSSUPPORTED,
	"DSET_SUPPORTED":  ACSSUPPORTED,
	"TRIBE_SUPPORTED": TRIBESUPPORTED,
}

// ParseClientExperience converts a string into a valid ClientExperience value, or returns an error if the string does not map
// to a known ClientExperience
func ParseClientExperience(input string) (ClientExperience, error) {
	if result, ok := validClientExperiences[input]; ok {
		return result, nil
	}
	return "", fmt.Errorf(`Invalid string value for ossrecord.ClientExperience: "%s"`, input)
}

var emailRegexp = regexp.MustCompile(`^(.*?)(\b[\w+.-]+@[\w+.-]+\b)(.*)$`)

// ConstructPerson constructs a full Person record from a simple string obtained from other tools
// (which might contain a name, an email address or both)
func ConstructPerson(info string) Person {
	m := emailRegexp.FindStringSubmatch(info)
	if m != nil {
		return Person{
			W3ID: strings.TrimSpace(m[2]),
			Name: strings.TrimSpace(m[1] + m[3]),
		}
	}
	return Person{
		Name: strings.TrimSpace(info),
	}
}

var normalizePIDPattern = regexp.MustCompile(`[ -]`)

// NormalizeProductID normalizes a product ID (PID) by removing spaces, dashes, etc.
// to allow for easy comparison
func NormalizeProductID(pid string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(pid))
	result := normalizePIDPattern.ReplaceAllLiteralString(trimmed, "")
	return result
}

// IsProductInfoNone returns true if a slice from the ProductInfo record (PartNumbers, ProductIDs, etc) is equal to the ProductInfoNone value
func IsProductInfoNone(slice []string) bool {
	if len(slice) == 1 && slice[0] == ProductInfoNone {
		return true
	}
	return false
}

// DeepCopy performs a deep copy of this OSSService object, returning a new OSSService object
func (oss *OSSService) DeepCopy() *OSSService {
	buffer, err := json.Marshal(oss)
	if err != nil {
		panic(err)
	}
	dest := new(OSSService)
	err = json.Unmarshal(buffer, dest)
	if err != nil {
		panic(err)
	}
	return dest
}

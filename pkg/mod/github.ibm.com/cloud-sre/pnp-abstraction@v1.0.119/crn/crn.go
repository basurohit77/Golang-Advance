package crn

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/pborman/uuid"
	"github.ibm.com/cloud-sre/oss-globals/consts"
	"github.ibm.com/cloud-sre/oss-globals/tlog"
)

// CRN holds the crn string in its parts
type CRN struct {
	Scheme          string
	Version         string
	CName           string
	CType           string
	ServiceName     string
	Region          string
	ScopeType       string
	Scope           string
	ServiceInstance string
	ResourceType    string
	Resource        string
}

// moved to github.ibm.com/cloud-sre/oss-globals/consts/crn.go
// const (
// 	//CrnScheme The base canonical format
// 	CrnScheme = "crn"
// 	//CrnSeparator set the CRN separator character
// 	CrnSeparator = ":"
// 	//ScopeSeparator set the scope separator character
// 	ScopeSeparator = "/"
// 	//CrnVersion current default version of CRN
// 	CrnVersion = "v1"
// 	//CrnDedicated CRN type of cloud instance dedicated
// 	CrnDedicated = "dedicated"
// 	//CrnPublic CRN type of cloud instance  public
// 	CrnPublic = "public"
// 	//CrnLocal CRN type of cloud instance local
// 	CrnLocal = "local"
// 	//Bluemix cloud instance that contains the resource
// 	Bluemix = "bluemix"
// 	//SofLayer cloud instance that contains the resource
// 	SofLayer = "softlayer"
// 	//IBMCloud cloud instance that contains the resource
// 	IBMCloud = "ibmcloud"
// )

var (
	// ErrMalformedCRN is an error for malformed CRN
	ErrMalformedCRN = errors.New("malformed CRN")
	// ErrMalformedScope is an error for malformed scopes
	ErrMalformedScope = errors.New("malformed scope in CRN")
	// ErrEmptyCRN is an error for no CRN to process
	ErrEmptyCRN = errors.New("no CRN to process")
)

// ResourceType describes the available resource types for a CRN
type ResourceType string

const (
	// BackupType is the Backup ResourceType for a CRN
	BackupType ResourceType = "backup"
	// TaskType is the Task ResourceType for a CRN
	TaskType ResourceType = "task"
)

// Segment describes the available segments in a CRN that can be queried
type Segment int

const (
	// Unknown is the Zero-Value for a Segment
	Unknown Segment = iota
	// ServiceInstance segment of a CRN
	ServiceInstance Segment = iota
	// Resource segment of a CRN
	Resource Segment = iota
)

// New creates a CRN from the given string or returns the string and an error if unsuccessful.
func New(s string) (CRN, error) {
	if s == "" {
		return CRN{}, ErrEmptyCRN
	}

	segments := strings.Split(s, consts.CrnSeparator)

	if len(segments) != consts.CrnLen || segments[consts.Schema] != consts.CrnScheme {
		return CRN{}, ErrMalformedCRN
	}

	crn := CRN{
		Scheme:          segments[0],
		Version:         segments[1],
		CName:           segments[2],
		CType:           segments[3],
		ServiceName:     segments[4],
		Region:          segments[5],
		ServiceInstance: segments[7],
		ResourceType:    segments[8],
		Resource:        segments[9],
	}

	scopeSegments := segments[consts.Scope]
	if scopeSegments != "" {

		scopeParts := strings.Split(scopeSegments, consts.ScopeSeparator)
		if len(scopeParts) != 2 {
			return CRN{}, ErrMalformedScope
		}
		crn.ScopeType = scopeParts[0]
		crn.Scope = scopeParts[1]

	}

	return crn, nil
}

// UpdateResource updates the CRN's resource type and resource.
func (crn *CRN) UpdateResource(rt ResourceType, r string) *CRN {
	crn.ResourceType = string(rt)
	crn.Resource = r

	return crn
}

// RemoveResource removes the CRN's resource type and resource.
func (crn *CRN) RemoveResource() *CRN {
	crn.ResourceType = ""
	crn.Resource = ""

	return crn
}

// IsValid checks to see if the identifier
// is a valid crn string
func IsValid(identifier string) bool {
	_, err := New(identifier)
	if err != nil {
		return false
	}

	return true
}

// IsValidAndReturn checks to see if the identifier
// is a valid crn string and returns the CRN
func IsValidAndReturn(identifier string) *CRN {
	CRNObj, err := New(identifier)
	if err != nil {
		return nil
	}

	return &CRNObj
}

// GetSegment takes a string and returns the string if it is not a CRN. Otherwise, it parses the CRN
// and returns the string for the specified segment of the CRN.
func GetSegment(seg Segment, crn string) string {
	if strings.HasPrefix(crn, consts.CrnScheme) {
		parsedCRN, err := New(crn)
		if err != nil {
			return crn
		}

		switch seg {
		case ServiceInstance:
			return parsedCRN.ServiceInstance
		case Resource:
			return parsedCRN.Resource
		}
	}

	return crn
}

// String returns the stringified version of a CRN
func (crn CRN) String() string {
	return strings.Join([]string{
		crn.Scheme,
		crn.Version,
		crn.CName,
		crn.CType,
		crn.ServiceName,
		crn.Region,
		crn.scopeSegment(),
		crn.ServiceInstance,
		crn.ResourceType,
		crn.Resource,
	}, consts.CrnSeparator)
}

// GenerateFake creates a fake CRN that is similar to what an actual instance might use. This is
// only meant to be used for testing.
func GenerateFake(shape string) CRN {
	return CRN{
		Scheme:          consts.CrnScheme,
		Version:         consts.CrnVersion,
		CName:           "testing",
		CType:           consts.CrnPublic,
		ServiceName:     fmt.Sprintf("databases-for-%s", shape),
		Region:          "us-south",
		ScopeType:       "a",
		Scope:           "b9552134280015ebfde430a819fa4bb3",
		ServiceInstance: uuid.New(),
	}
}

// ScopeSegment returns the scope portion of a CRN
// as a string
func (crn CRN) scopeSegment() string {
	if crn.ScopeType == "" && crn.Scope == "" {
		return ""
	}
	return crn.ScopeType + consts.ScopeSeparator + crn.Scope
}

// GetCloudName returns the cloud name from the CRN
func (crn *CRN) GetCloudName() string {
	return crn.CName
}

// GetCloudType returns the cloud type from the CRN
func (crn *CRN) GetCloudType() string {
	return crn.CType
}

// GetServiceName returns the service name from the CRN
func (crn *CRN) GetServiceName() string {
	return crn.ServiceName
}

// GetRegion returns the region from the CRN
func (crn *CRN) GetRegion() string {
	return crn.Region
}

// GetScope returns the scope from the CRN
/* func (crn *CRN) GetScope() scope {
	return BuildScope(crn.scopePrefix, crn.scopeID)
} */

// GetServiceInstance returns the service instance from the CRN
func (crn *CRN) GetServiceInstance() string {
	return crn.ServiceInstance
}

// GetResourceType returns the resource type from the CRN
func (crn *CRN) GetResourceType() string {
	return crn.ResourceType
}

// GetResourceID returns the resource from the CRN
func (crn CRN) GetResourceID() string {
	return crn.Resource
}

/* const crnFormatStr = "%s" + segmentDelimiter + "%s" + segmentDelimiter + "%s" + segmentDelimiter + "%s" + segmentDelimiter + "%s" + segmentDelimiter + "%s" + segmentDelimiter + "%s" + segmentDelimiter + "%s" + segmentDelimiter + "%s"

// GetRepresentation returns the string representation from a CRN type
func (crn *CRN) GetRepresentation() string {
	return fmt.Sprintf(crnFormatStr, c.Scheme, crn.cloudName, crn.cloudType, crn.serviceName, crn.region, BuildScope(crn.scopePrefix, crn.scopeID), crn.serviceInstance, crn.resourceType, crn.resourceID)
} */

/* / NewCRN will create and return a CRN object for you
func NewCRN(template *Template) (*CRN, error) {
	// Validate Template
	var err error
	if template.CloudName == "" {
		return nil, errors.New("cloud name is required for CRN")
	}

	if template.CloudType == "" {
		return nil, errors.New("cloud type is required for CRN")
	}

	if template.ServiceName == "" {
		return nil, errors.New("service name is required for CRN")
	}

	if template.ServiceInstance == "" {
		return nil, errors.New("service instance is required for CRN")
	}

	if template.ResourceID == "" {
		return nil, errors.New("resource id is required for CRN")
	}

	var scopePrefix scopePrefixEnum
	var scopeID string
	if template.Scope != "" {
		if scopePrefix, scopeID, err = ParseScope(template.Scope); err != nil {
			return nil, err
		}
	}

	crn := new(CRN)
	crn.cloudName = template.CloudName
	crn.cloudType = template.CloudType
	crn.serviceName = template.ServiceName
	crn.region = template.Region
	crn.scopeID = scopeID
	crn.scopePrefix = scopePrefix
	crn.serviceInstance = template.ServiceInstance
	crn.resourceType = template.ResourceType
	crn.resourceID = template.ResourceID

	return crn, nil
} */

// Segment indexes for the CRN string
const (
	cloudNameIndex = iota + 2
	cloudTypeIndex
	serviceNameIndex
	regionIndex
	scopeIndex
	serviceInstanceIndex
	resourceTypeIndex
	resourceIDIndex
)

func isValidCloudType(cloudType string) bool {
	switch cloudType {
	case consts.CrnPublic:
		return true
	case consts.CrnLocal:
		return true
	case consts.CrnDedicated:
		return true
	}

	return false
}

const serviceNamePattern = `^[a-z0-9-]{2,32}$`
const regionNamePattern = `^[a-zA-Z0-9-]{4,32}$`
const resourceTypePattern = `^[a-zA-Z0-9-]{2,32}$`
const resourceIDPattern = `^[a-zA-Z0-9-/]{2,32}$`

func isValidServiceName(serviceName string) bool {
	return regexp.MustCompile(serviceNamePattern).MatchString(serviceName)
}

func isValidResourceID(resourceID string) bool {
	return regexp.MustCompile(resourceIDPattern).MatchString(resourceID)
}

func isValidRegionName(regionName string) bool {
	return !(!regexp.MustCompile(regionNamePattern).MatchString(regionName) && regionName != "")
}

func isValidResourceType(resourceType string) bool {
	return !(!regexp.MustCompile(resourceTypePattern).MatchString(resourceType) && resourceType != "")
}

// Gets the field value of the struct object.  Useful for dynamic referencing
func field(t interface{}, key string) string {
	strs := strings.Split(key, ".")
	v := reflect.Indirect(reflect.ValueOf(t))
	for _, s := range strs[1:] {
		v = v.FieldByName(s)
		fmt.Println(v.Interface())

	}
	return v.Interface().(string)
}

// IsCRNMatchFast determines whether the crnString (candidate) is a match for the matchString.  Simple string compare
func IsCRNMatchFast(candidateString string, matchString string, isWildcard bool) bool {
	matches := true

	if candidateString == "" ||
		candidateString == "crn:v1::::::::" ||
		matchString == "null" ||
		matchString == "" ||
		matchString == "crn:v1::::::::" {
		return matches
	}
	log.Println(tlog.Log(), "candidate(from msg): ", candidateString, "\t matchString(from watchmap): ", matchString, "\t wildcard(from watchmap): ", isWildcard)
	matchString = strings.ToLower(matchString)
	candidateString = strings.ToLower(candidateString)
	matchCRN := strings.Split(matchString, consts.CrnSeparator)
	candidateCRN := strings.Split(candidateString, consts.CrnSeparator)
	if len(matchCRN) != len(candidateCRN) || len(candidateCRN) != consts.CrnLen {
		log.Println(tlog.Log(), "Incorrect CRN lengths detected")
		return false
	}

	for i, v := range matchCRN {
		if isWildcard && strings.Index(candidateCRN[i], v) != 0 {
			return false
		} else if !isWildcard && candidateCRN[i] != v {
			log.Println(tlog.Log(), "candidate value does not match: ", candidateCRN[i], " : ", v)
			return false
		}
	}

	return matches
}

// IsCRNMatch determines whether the crnString (candidate) is a match for the matchString
func IsCRNMatch(crnString string, matchString string) bool {
	matches := true

	CRNCandidate := IsValidAndReturn(crnString)
	if CRNCandidate == nil {
		return false
	}

	CRNSearch := IsValidAndReturn(matchString)
	if CRNSearch == nil {
		return false
	}

	v := reflect.Indirect(reflect.ValueOf(CRNSearch))
	typeOfT := v.Type()
	//values := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())

		valueOfCandidateField := field(CRNCandidate, "CRNCandidate."+typeOfT.Field(i).Name)
		valueOfSearchField := f.Interface().(string)

		if (typeOfT.Field(i).Name == "Region" &&
			strings.Index(valueOfCandidateField, CRNSearch.Region) == 0) ||
			len(valueOfSearchField) == 0 ||
			valueOfCandidateField == valueOfSearchField {

			fmt.Println(typeOfT.Field(i).Name + " for this matches")
		} else {

			fmt.Println(typeOfT.Field(i).Name + " for this DOES NOT match")
			matches = false
			break
		}

		//values[i] = v.Field(i).Interface()
	}

	fmt.Printf("Completed %v\n", CRNCandidate)
	return matches

}

// GetCRNMatches returns matches to the crnString
func GetCRNMatches(crnString string) []string {
	// Parse the string into a manageable CRN
	parsedCRN, err := New(crnString)
	if err != nil {
		return []string{"Error parsing crnString : " + crnString + "\n\t" + err.Error()}

	}

	// check CRN for whether there are fields and create query
	query := "select * from database where  "
	firstWhere := true

	if len(parsedCRN.CName) > 0 {
		if !firstWhere {
			query = query + " and "
		}
		query = query + " CName is " + parsedCRN.CName

		firstWhere = false

	}
	if len(parsedCRN.CType) > 0 {
		if !firstWhere {
			query = query + " and "
		}
		query = query + " CType is " + parsedCRN.CType

		firstWhere = false

	}
	if len(parsedCRN.ServiceName) > 0 {
		if !firstWhere {
			query = query + " and "
		}
		query = query + " ServiceName is " + parsedCRN.ServiceName
	}
	if len(parsedCRN.Region) > 0 {
		if !firstWhere {
			query = query + " and "
		}
		query = query + " Region like " + parsedCRN.Region + "%"

		firstWhere = false
	}
	if len(parsedCRN.ScopeType) > 0 {
		if !firstWhere {
			query = query + " and "
		}
		query = query + " ScopeType is " + parsedCRN.ScopeType

		firstWhere = false
	}
	if len(parsedCRN.Scope) > 0 {
		if !firstWhere {
			query = query + " and "
		}
		query = query + " Scope is " + parsedCRN.Scope

		firstWhere = false
	}
	if len(parsedCRN.ServiceInstance) > 0 {
		if !firstWhere {
			query = query + " and "
		}
		//fmt.Println("Length of service instance : " + strconv.FormatBool(len(parsedCRN.ServiceInstance) > 0))
		query = query + " ServiceInstance is " + parsedCRN.ServiceInstance

	}

	fmt.Println("query : " + query)
	// Make the request to the db

	// Parse the db response into a string array

	// Return

	return []string{"This is a test", "test2"}

}

package ossrecord

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ClearingHouseReference represents one linkage of an OSS entry to a ClearingHouse entry
type ClearingHouseReference struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// ClearingHouseReferences is a list of ClearingHouseReference records
type ClearingHouseReferences []ClearingHouseReference

// Constant definitions for common tags associated with ClearingHouseReferences
// Note: there may be additional, free-form tags
const (
	ClearingHouseReferenceTagSource             = "Src:"             // How was this ClearingHouseReference found
	ClearingHouseReferenceTagSourcePID          = "Src:PID"          // This ClearingHouseReference was found by matching ProductIDs (PIDs)
	ClearingHouseReferenceTagSourceNames        = "Src:Names"        // This ClearingHouseReference was found by matching service names
	ClearingHouseReferenceTagSourceCRNAttribute = "Src:CRNAttribute" // This ClearingHouseReference was found by matching the CRN service-name attribute in ClearingHouse records
)

// AddClearingHouseReference adds a CH reference to a slice containing a list of CH references, while avoiding duplicates
// This function returns true if the CH reference was not already present, false otherwise
func (refs *ClearingHouseReferences) AddClearingHouseReference(name, id string, tags ...string) (isNew bool) {
	var foundID bool
	for i := range *refs {
		if id == (*refs)[i].ID {
			foundID = true
			for _, t1 := range tags {
				var foundTag bool
				for _, t2 := range (*refs)[i].Tags {
					if t1 == t2 {
						foundTag = true
						break
					}
				}
				if !foundTag {
					(*refs)[i].Tags = append((*refs)[i].Tags, t1)
					sort.Strings((*refs)[i].Tags)
				}
			}
			break
		}
	}
	if !foundID {
		r1 := ClearingHouseReference{id, name, tags}
		sort.Strings(r1.Tags)
		*refs = append(*refs, r1)
	}
	sort.Slice(*refs, func(i, j int) bool {
		return (*refs)[i].Name < (*refs)[j].Name
	})
	return !foundID
}

// FindTag checks if a given tag is present in this ClearingHouseReference.
// If the tag exists and ends in a ":", the value after the ":" is returned
// If the tag exists and does not end in a ":", the entire tag string is returned
// If the tag does not exist, the empty string is returned
func (ref *ClearingHouseReference) FindTag(tag string) string {
	var pattern *regexp.Regexp

	if strings.HasSuffix(tag, ":") {
		pattern = regexp.MustCompile(fmt.Sprintf(`^%s(\S+)$`, tag))
	}

	for _, t := range ref.Tags {
		if pattern != nil {
			if m := pattern.FindStringSubmatch(t); m != nil {
				return m[1]
			}
		} else {
			if t == tag {
				return t
			}
		}
	}
	return ""
}

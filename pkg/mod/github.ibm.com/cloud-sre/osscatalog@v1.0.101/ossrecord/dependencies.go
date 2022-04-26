package ossrecord

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Dependency represents one product dependency (inbound or outbound)
type Dependency struct {
	Service string   `json:"service"`
	Tags    []string `json:"tags"`
}

// Dependencies is a list of Dependency records
type Dependencies []*Dependency

// Constant definitions for common tags associated with Dependencies
// Note: there may be additional, free-form tags
const (
	DependencyTagSource              = "Src:"              // How was this Dependency found
	DependencyTagSourceClearingHouse = "Src:ClearingHouse" // This Dependency was found in ClearingHouse
	DependencyTagType                = "Type:"             // Type of this Dependency (from ClearingHouse)
	DependencyTagStatus              = "Status:"           // Status of this Dependency (from ClearingHouse)
	DependencyTagIssues              = "Issues:"           // Number of issues encountered while computing this Dependency
	DependencyTagNotOSS              = "NotOSS"            // This dependency is not a OSS entity
)

// AddDependency adds a dependency to a slice containing a list of dependencies, while avoiding duplicates
func (deps *Dependencies) AddDependency(service string, tags ...string) {
	var foundService bool
	for i := range *deps {
		if service == (*deps)[i].Service {
			foundService = true
			for _, t1 := range tags {
				var foundTag bool
				for _, t2 := range (*deps)[i].Tags {
					if t1 == t2 {
						foundTag = true
						break
					}
				}
				if !foundTag {
					//					d.Tags = append(d.Tags, t1)
					(*deps)[i].Tags = append((*deps)[i].Tags, t1)
					//					sort.Strings(d.Tags)
					sort.Strings((*deps)[i].Tags)
				}
			}
			break
		}
	}
	if !foundService {
		d1 := Dependency{service, tags}
		sort.Strings(d1.Tags)
		*deps = append(*deps, &d1)
	}
	sort.Slice(*deps, func(i, j int) bool {
		return (*deps)[i].Service < (*deps)[j].Service
	})
}

// CountTag returns the number of Dependencies that contain a given tag.
// The tag may either be an exact match, or a prefix match if the specified tag ends in ":"
func (deps *Dependencies) CountTag(tag string) int {
	var count int
	var byPrefix bool
	if strings.HasSuffix(tag, ":") {
		byPrefix = true
	}
	for _, d := range *deps {
		for _, t := range d.Tags {
			if byPrefix {
				if strings.HasPrefix(t, tag) {
					count++
				}
			} else {
				if t == tag {
					count++
				}
			}
		}
	}
	return count
}

// FindTag checks if a given tag is present in this dependency.
// If the tag exists and ends in a ":", the value after the ":" is returned
// If the tag exists and does not end in a ":", the entire tag string is returned
// If the tag does not exist, the empty string is returned
func (d *Dependency) FindTag(tag string) string {
	var pattern *regexp.Regexp

	if strings.HasSuffix(tag, ":") {
		pattern = regexp.MustCompile(fmt.Sprintf(`^%s(\S+)$`, tag))
	}

	for _, t := range d.Tags {
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

// ComparableString forces comparisons of Dependency records to be on a single line with all attributes together
func (d *Dependency) ComparableString() string {
	return fmt.Sprintf("%s %q", d.Service, d.Tags)
}

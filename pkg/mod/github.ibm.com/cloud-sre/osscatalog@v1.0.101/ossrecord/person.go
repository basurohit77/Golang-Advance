package ossrecord

import (
	"fmt"
	"strings"
)

// Person represents the information to identify one person who has some role in a OSS record (e.g. Owner, etc.)
type Person struct {
	W3ID string `json:"w3id"`
	Name string `json:"name"`
}

// IsValid returns true if this Person record contain a valid user
func (p *Person) IsValid() bool {
	// TODO: should check that the W3ID is valid in Bluepages
	if p == nil {
		return false
	}
	if p.Name == "" && p.W3ID == "" {
		return false
	}
	if !strings.Contains(p.W3ID, "@") {
		return false
	}
	return true
}

// IsEmpty returns true if this Person record contain no information at all
// Note: not-empty does not necessarily mean that that IsValid() == true
func (p *Person) IsEmpty() bool {
	if p == nil {
		return true
	}
	if p.Name == "" && p.W3ID == "" {
		return true
	}
	return false
}

// ComparableString forces comparisons of Person records to be on a single line with all attributes together
func (p *Person) ComparableString() string {
	return p.String()
}

// PersonListEntry represents one entry in a list of individuals, each of whom can be associated with some tags
// to reflect their role within the list.
// Used for example for the change_approvers list associated with a OSSTribe
type PersonListEntry struct {
	Member Person   `json:"member"`
	Tags   []string `json:"tags"`
}

// ComparableString forces comparisons of PersonListEntry records to be on a single line with all attributes together
func (pl *PersonListEntry) ComparableString() string {
	return fmt.Sprintf("%s %q", pl.Member.String(), pl.Tags)
}

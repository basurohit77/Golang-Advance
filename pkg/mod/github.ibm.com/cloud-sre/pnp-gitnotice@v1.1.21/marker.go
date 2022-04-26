package gitnotice

import (
	"regexp"
	"strings"
)

const (
	// MarkerBeginningStr is the text that starts any marker
	MarkerBeginningStr = `**====`
	// MarkerEndingStr is the text that ends any marker
	MarkerEndingStr = `====**`
	// MarkerBeginningRegex is the regex to match the start of any marker
	MarkerBeginningRegex = `[*]{2}==== *`
	// MarkerEndingRegex is the regex to match the end of any marker
	MarkerEndingRegex = ` *====[*]{2}`
	// StartMarker is the marker containing the start time
	StartMarker = "START"
	// ServiceMarker is the marker containing the list of services
	ServiceMarker = "SERVICE"
	// LocationMarker contains the list of affected locations
	LocationMarker = "IMPACTED LOCATIONS"
	// TitleMarker is the Title of the notification
	TitleMarker = "TITLE"
	// AudienceMarker is the audience for the notification
	AudienceMarker = "AUDIENCE"
	// DescriptionMarker is the description for the notification
	DescriptionMarker = "DESCRIPTION"
)

// MarkerSection is a parsed section of the description field in the issue
// The content will have all of the section including the marker headers
type MarkerSection struct {
	MarkerType string // Something like StartMarker or TitleMarker
	Content    string
}

// ParseDescription will parse the full text of a description and return an array
// of sections repsenting each of the different sections
func ParseDescription(input string) (list map[string]*MarkerSection) {

	list = make(map[string]*MarkerSection)

	addSection(TitleMarker, input, list)
	addSection(DescriptionMarker, input, list)
	addSection(StartMarker, input, list)
	addSection(LocationMarker, input, list)
	addSection(ServiceMarker, input, list)
	addSection(AudienceMarker, input, list)

	return list
}

// GetContent returns the content of the section without the first marker line.
func (m *MarkerSection) GetContent() string {
	if m == nil || len(m.Content) == 0 {
		return ""
	}

	i := strings.Index(m.Content, "\n")

	return m.Content[i+1:]
}

func addSection(marker, input string, list map[string]*MarkerSection) {
	list[marker] = &MarkerSection{MarkerType: marker, Content: getMarkerContent(marker, input)}
}

func getMarkerContent(marker, input string) (output string) {

	regex := regexp.MustCompile(MarkerBeginningRegex + marker + MarkerEndingRegex)

	loc := regex.FindStringIndex(input)
	if loc == nil {
		return ""
	}

	end := strings.Index(input[loc[0]+1:], MarkerBeginningStr) // Find beginning of next marker

	if end > 0 {
		output = input[loc[0] : loc[0]+end]
	} else { // Probably the last section in the description
		output = input[loc[0]:]
	}

	return output
}

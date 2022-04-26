package gitnotice

import (
	"strings"
	"time"
)

// Notice is the data that can be parsed from a git issue
type Notice struct {
	Start       time.Time
	Services    []string
	Locations   []CRN
	Title       string
	Description string
	Audience    string
	Err         *Error
}

// BuildNotice will create a notice from a given description
func BuildNotice(description string) (notice *Notice, err *Error) {

	notice = new(Notice)

	sections := ParseDescription(description)

	notice.Title = notice.getStringContent(sections, TitleMarker, true)
	notice.Description = notice.getStringContent(sections, DescriptionMarker, true)
	notice.Audience = notice.getStringContent(sections, AudienceMarker, false)
	notice.Start = notice.getTimeContent(sections, StartMarker)
	notice.Services = notice.getStringListContent(sections, ServiceMarker)
	notice.Locations = notice.getCRNListContent(sections, LocationMarker)

	return notice, notice.Err
}

func (notice *Notice) getTimeContent(sections map[string]*MarkerSection, marker string) (result time.Time) {

	var err error

	sec := sections[marker]

	if sec == nil || len(strings.TrimSpace(sec.GetContent())) == 0 {
		// [2019-08-05] Bill W. said he doesn't want a start date for the event, so suppress the error
		// notice.Err = notice.Err.Add(nil, "No "+marker+" provided. Be sure to use the template.")
		return result
	}

	tStr := strings.TrimSpace(sec.GetContent())

	result, err = time.Parse(time.RFC3339, tStr)

	if err != nil || result.IsZero() {
		notice.Err = notice.Err.Add(nil, "Could not parse "+marker+". Incorrect time format. Error= "+err.Error())
		return result
	}

	return result
}

func (notice *Notice) getStringContent(sections map[string]*MarkerSection, marker string, required bool) string {

	sec := sections[marker]
	if sec == nil || len(sec.GetContent()) == 0 {
		if required {
			notice.Err = notice.Err.Add(nil, "No "+marker+" provided. Be sure to use the template.")
		}
		return ""
	}

	return sec.GetContent()

}

func (notice *Notice) getCRNListContent(sections map[string]*MarkerSection, marker string) (result []CRN) {

	strList := notice.getStringListContent(sections, marker)

	for _, str := range strList {

		crn, err := ToCRN(str)

		if err != nil {
			notice.Err = notice.Err.Meld(err)
		} else {
			result = append(result, crn)
		}
	}
	return result

}

var gResult []string

func (notice *Notice) getStringListContent(sections map[string]*MarkerSection, marker string) []string {

	result := []string{}

	sec := sections[marker]

	if sec == nil || len(sec.GetContent()) == 0 {
		notice.Err = notice.Err.Add(nil, "No "+marker+" provided. Be sure to use the template.")
		return nil
	}

	list := strings.Split(sec.GetContent(), "\n")

	for _, s := range list {
		tt := strings.TrimSpace(s)
		if len(tt) > 0 {
			result = append(result, tt)
		}
	}

	return result
}

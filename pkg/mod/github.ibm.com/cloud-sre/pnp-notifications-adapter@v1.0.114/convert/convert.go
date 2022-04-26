package convert

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/api"
)

const (
	retractedTag = "retract"
)

// CnToPGniList converts a list of Common Notifications to a list of Postgres NotificationInserts
func CnToPGniList(inputList *api.NotificationList) (outputList []datastore.NotificationInsert) {

	outputList = make([]datastore.NotificationInsert, 0, 5)
	for _, in := range inputList.Items {
		out := CnToPGni(in)
		outputList = append(outputList, out...)
	}

	return outputList

}

// PGnrToCnList means Postgres NotificationReturns to Common Notification
// This converts between lists of these objects.  It does the right thing to compensate for CRNs
func PGnrToCnList(input []datastore.NotificationReturn, tagsParamValue string) (output *api.NotificationList) {

	if input == nil {
		return &api.NotificationList{Items: []*api.Notification{}}
	}

	// a map of source ID to a list of notifications - used to collate based on CRN
	// SourceID should be the ID of the cloudant document that contained the notification
	collator := make(map[string][]*api.Notification)

	for nr := range input {

		n := PGnrToCn(&input[nr])

		list := collator[n.SourceID]

		if list == nil {
			list = make([]*api.Notification, 0, 1)
		}

		if tagsParamValue != "" {
			if len(n.Tags) > 0 {
				tags := strings.Split(tagsParamValue, ",")

				// singleTagSearch is used when a single tag is used in the search. i.e. ?tags=retract-1
				var singleTagSearch = false
				// singleTagMatch is used to identify if there is a match for a record when there is a single input tag
				var singleTagMatch = false

				if len(tags) == 1 {
					singleTagSearch = true
				}

				var tagsToReturn []string

				for _, tag := range tags {

					for _, ntag := range n.Tags {
						if singleTagSearch {
							// if only one tag is used in the search, search the input tag in all the tags available in the notification
							if ntag == tag {
								singleTagMatch = true
							}
						} else {
							// gets in here only when multiple tags in the search
							// if tags match, add it to slice to be returned in the tags JSON attribute
							if ntag == tag {
								tagsToReturn = append(tagsToReturn, ntag)
							}
						}
					}

				}

				if singleTagMatch {
					if strings.Contains(tags[0], "-") {
						splitTag := strings.Split(tags[0], "-")
						n.Tags = []string{splitTag[0], tags[0]}
					} else {
						n.Tags = tags
					}
				}

				// both the input tags passed in and the tags found in each notification record match
				// this can only be true when 2 or more tags are used in the tags filter
				multipleTagMatch := len(tagsToReturn) == len(tags)

				if multipleTagMatch {
					n.Tags = tagsToReturn
				}

				if multipleTagMatch || singleTagMatch {
					list = append(list, n)
					collator[n.SourceID] = list
				}

			}
		} else {

			retracted := false
			for _, s := range n.Tags {
				if strings.Index(s, retractedTag) == 0 {
					retracted = true
					break
				}
			}

			if !retracted {
				list = append(list, n)
				collator[n.SourceID] = list
			}
		}
	}

	// [2019-04-01] BSPNs are handled in a special way because in ServiceNow, BSPNs are not modified.
	// Users create new BSPNs to replace the old. So to make life simpler for PnP clients, we will pull
	// out the older BSPNs so as to only return the most recent for each incident.
	collator = removeOldBSPNs(collator)

	// Now create the result to send back to the caller
	output = new(api.NotificationList)

	for _, v := range collator {

		// For ServiceNow maintenance records, we don't want to collate based on Publications.
		// The current strategy is to leave all of these as separate notification records
		if len(v) > 0 && v[0].NotificationType == "maintenance" && v[0].Source == "servicenow" {

			for _, s := range v {
				output.Items = append(output.Items, s)
			}

		} else {

			sort.Sort(ByRecordID(v)) // This is important to get some consistency with Record IDs on notifications

			var final *api.Notification
			for _, s := range v {
				if final == nil {
					final = s
				} else {
					final.CRNs = append(final.CRNs, s.CRNs...)
				}
			}
			output.Items = append(output.Items, final)
		}
	}

	return output

}

// removeOldBSPNs will look at the input map which is keyed by sourceID and will pull out all the old
// BSPN entries.  The basic assumption is for any one incident, there is only ever one most recent; and
// consequently the only active BSPN.  In other words, you don't have 2 BSPNs active for the same CIE.
func removeOldBSPNs(input map[string][]*api.Notification) map[string][]*api.Notification {

	result := make(map[string][]*api.Notification)

	incToSrc := make(map[string][]string) // maps incident ID to bspn IDs

	for sourceID, values := range input { // Note sourceID is the BSPN id for incident types

		if len(values) > 0 && values[0].NotificationType == "incident" {

			incID := values[0].IncidentID // incident ID is the same for all BSPNs in the group

			list := incToSrc[incID]
			if list == nil {
				list = make([]string, 0)
			}
			list = append(list, sourceID) // Matching all unique BSPXXX records to single incident records

			incToSrc[incID] = list

		} else {
			result[sourceID] = values // just throw the non-incidents on the output list
		}
	}

	// Walk the BSPNs for each incident and select the newest and discard the others.
	for _, bspns := range incToSrc {
		max := ""
		var maxTime time.Time

		for _, bspn := range bspns {

			time1, err := time.Parse(time.RFC3339, input[bspn][0].UpdateTime)
			if err != nil {
				log.Println("removeOldBSPNs", err) // Should not occur
				continue
			}

			if max == "" || time1.After(maxTime) {
				max = bspn
				maxTime = time1
			}

		}

		if max != "" && input[max] != nil {
			result[max] = input[max]
		}
	}

	return result
}

// PGnrToCn means Postgres NotificationReturns to Common Notification
func PGnrToCn(input *datastore.NotificationReturn) (output *api.Notification) {

	if input == nil {
		return nil
	}

	output = new(api.Notification)

	output.RecordID = input.RecordID
	output.EventTimeStart = input.EventTimeStart
	output.EventTimeEnd = input.EventTimeEnd
	output.IncidentID = input.IncidentID
	output.CreationTime = input.SourceCreationTime
	output.UpdateTime = input.SourceUpdateTime
	output.PNPCreationTime = input.PnpCreationTime
	output.PNPUpdateTime = input.PnpUpdateTime
	output.Category = input.Category
	output.Source = input.Source
	output.SourceID = input.SourceID
	output.NotificationType = input.Type
	output.CRNs = []string{input.CRNFull}
	output.PnPRemoved = input.PnPRemoved

	if input.Tags != "" {
		output.Tags = strings.Split(input.Tags, ",")
	} else {
		output.Tags = []string{}
	}

	for _, dn := range input.ResourceDisplayNames {
		output.DisplayName = append(output.DisplayName, &api.TranslatedString{Text: dn.Name, Language: dn.Language})
	}

	if len(input.ShortDescription) > 0 {
		output.ShortDescription = input.ShortDescription[0].Name
	}

	if len(input.LongDescription) > 0 {
		output.LongDescription = input.LongDescription[0].Name
	}

	return output
}

// CnToPGni means Common Notification to Postgres NotificationInsert
// Note you will get multiple records back if the input has more than one CRN
func CnToPGni(input *api.Notification) (output []datastore.NotificationInsert) {

	if input == nil {
		return nil
	}

	output = make([]datastore.NotificationInsert, 0, len(input.CRNs))
	for _, crn := range input.CRNs {
		output = append(output, *crnCnToPGni(crn, input))
	}
	return output
}

func crnCnToPGni(crn string, input *api.Notification) (output *datastore.NotificationInsert) {

	if input == nil {
		return nil
	}

	output = new(datastore.NotificationInsert)

	output.SourceCreationTime = input.CreationTime
	output.SourceUpdateTime = input.UpdateTime
	output.EventTimeStart = input.EventTimeStart
	output.EventTimeEnd = input.EventTimeEnd
	output.Source = input.Source
	output.SourceID = input.SourceID
	output.Type = input.NotificationType
	output.Category = input.Category
	output.IncidentID = input.IncidentID
	output.CRNFull = crn
	output.PnPRemoved = input.PnPRemoved

	if len(input.Tags) > 0 {
		output.Tags = strings.Join(input.Tags, ",")
	}

	output.ShortDescription = make([]datastore.DisplayName, 1)
	output.ShortDescription[0].Name = input.ShortDescription
	output.ShortDescription[0].Language = "en"

	output.LongDescription = make([]datastore.DisplayName, 1)
	output.LongDescription[0].Name = input.LongDescription
	output.LongDescription[0].Language = "en"

	for _, dn := range input.DisplayName {
		output.ResourceDisplayNames = append(output.ResourceDisplayNames, datastore.DisplayName{Name: dn.Text, Language: dn.Language})
	}

	return output
}

// NInsertToNReturn converts a notification insert to a notification return
func NInsertToNReturn(ni *datastore.NotificationInsert) *datastore.NotificationReturn {

	nr := new(datastore.NotificationReturn)

	//nr.RecordID = fmt.Sprintf("%s+%s+%s", ni.Source, ni.SourceID, ni.CRNFull)
	nr.PnpCreationTime = ni.SourceCreationTime
	nr.PnpUpdateTime = ni.SourceUpdateTime
	nr.SourceCreationTime = ni.SourceCreationTime
	nr.SourceUpdateTime = ni.SourceUpdateTime
	nr.EventTimeStart = ni.EventTimeStart
	nr.EventTimeEnd = ni.EventTimeEnd
	nr.Source = ni.Source
	nr.SourceID = ni.SourceID
	nr.Type = ni.Type
	nr.Category = ni.Category
	nr.IncidentID = ni.IncidentID
	nr.CRNFull = ni.CRNFull
	nr.ResourceDisplayNames = append(nr.ResourceDisplayNames, ni.ResourceDisplayNames...)
	nr.ShortDescription = append(nr.ShortDescription, ni.ShortDescription...)
	nr.LongDescription = append(nr.LongDescription, ni.LongDescription...)

	return nr
}

// ByRecordID sort is used to select the record ID that will be used on the collated notification that
// is returned. Background: The notification logic will "collate" multiple notification records
// into a single record. This is nice for the clients because they get a single notification record
// with multiple CRNs.  This record can be created no matter which record ID is requested for any
// the the notification records in our database.  When a generic query is sent in, we just want to
// attempt to return a consistent record ID to cut down on variability.
type ByRecordID []*api.Notification

func (a ByRecordID) Len() int           { return len(a) }
func (a ByRecordID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRecordID) Less(i, j int) bool { return a[i].RecordID < a[j].RecordID }

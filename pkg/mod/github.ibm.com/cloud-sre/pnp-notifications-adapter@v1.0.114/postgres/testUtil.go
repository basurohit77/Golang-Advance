package postgres

import (
	"fmt"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

// This file contains only unit test utilties.  It must be in a non "*_test.go" file because
// these functions are referenced outside of this package in other tests.

// CreateNotificationReturn will create a mocked up notification return
func CreateNotificationReturn(source, sourceID, nType, nCategory, incidentID, crn string, cPnpTime, uPnpTime, cSrcTime, uSrcTime, sEvtTime, eEvtTime time.Duration, resourceName, shortDesc, longDesc string) datastore.NotificationReturn {

	result := &datastore.NotificationReturn{Source: source, SourceID: sourceID, Type: nType, Category: nCategory, IncidentID: incidentID, CRNFull: crn}

	result.RecordID = fmt.Sprintf("%s+%s+%s", source, sourceID, crn)

	result.PnpCreationTime = time.Now().Add(cPnpTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.PnpUpdateTime = time.Now().Add(uPnpTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.SourceCreationTime = time.Now().Add(cSrcTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.SourceUpdateTime = time.Now().Add(uSrcTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.EventTimeStart = time.Now().Add(sEvtTime * -1).Format("2006-01-02T15:04:05:000Z")
	result.EventTimeEnd = time.Now().Add(eEvtTime * -1).Format("2006-01-02T15:04:05:000Z")

	result.ResourceDisplayNames = append(result.ResourceDisplayNames, datastore.DisplayName{Name: resourceName, Language: "en"})
	result.ShortDescription = append(result.ShortDescription, datastore.DisplayName{Name: shortDesc, Language: "en"})
	result.LongDescription = append(result.LongDescription, datastore.DisplayName{Name: longDesc, Language: "en"})
	result.Tags = ""

	return *result
}

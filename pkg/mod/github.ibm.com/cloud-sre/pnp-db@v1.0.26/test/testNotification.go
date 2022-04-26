package test

import (
	"database/sql"
	"log"
//	"net/http"
	"strings"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

// Test Maintenance for IBM Public Cloud
func TestNotification(database *sql.DB) int {
	FCT := "TestNotification "
	failed := 0

	log.Println("======== "+FCT+" Test 1 ========")

	log.Println("Test simple InsertNotification and GetNotificationByRecordID")

	resourceDisplayNames := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "resource display, name1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "resource display, name2",
			Language: "fr",
		},
	}

	shortDescription := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "short, description 1",
			Language: "fr",
		},
		datastore.DisplayName {
			Name: "Third-party service deprecations: MMMMMM, FFFFFFFFF.io, and DDDDDDD",
			Language: "en",
		},
	}

	longDescription := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "We would like to inform you that we are retiring the following services from the MMM on September 27, 2018:\n<ul><li>MMMMMM</li><li>FFFFFFFFF.io</li><li>DDDDDDD</li></ul>\nFor more information on how to continue to use MMMMMM services, see <u><a href=\"https://mmmm.com/\" target=\"_blank\">https://mmmm.com/</a></u>.<br><br>Here’s what you need to know:<ul><li>You cannot provision new service instances. However, existing instances will continue to be supported until the End of Support date.</li><li>The End of Support date is September 27, 2018.</li><li>Through September 27, 2018, all existing instances will continue to be available through the command line.</li><li>Any instance of these services that is still provisioned as of the End of Support date will be deleted.</li><li>Delete your service instances before the End of Support date.</li></ul>\nThis information originated in the <u><a href=\"https://www.mmm.com/blog/test/\" target=\"_blank\">Deprecations</a></u> article within the MMM blog.",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "long, description 2\nWith some longer explanation, of the announcement",
			Language: "fr",
		},
	}

	notification := datastore.NotificationInsert {
		SourceCreationTime:   "2018-06-07T22:01:01Z",
		EventTimeStart:       "2018-06-07T21:01:00Z",
		Source:               "test_source1",
		SourceID:             "source_id1",
		Type:                 "announcement",
		Category:             "runtimes",
		CRNFull:              "crn:v1:Bluemix:Public:Service-Name1:location1::Service-Instance1:Resource-Type1:",
		ResourceDisplayNames: resourceDisplayNames,
		ShortDescription:     shortDescription,
		LongDescription:      longDescription,
		Tags:                 "retract",
		RecordRetractionTime: "2018-06-07T22:02:01Z",
		ReleaseNoteUrl:       "https://cloud.ibm.com/docs/overview?topic=whates-new",
	}

	record_id, err,_ := db.InsertNotification(database, &notification)
	if err != nil {
		log.Println("Insert Notification failed: ", err.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id)
		log.Println("Insert Notification passed")
	}
	log.Println("Test get Notification")
	notificationReturn, err1,_ := db.GetNotificationByRecordID(database, record_id)
	if err1 != nil {
		log.Println("Get Notification failed: ", err1.Error())
		failed++
	} else {
		log.Println("Get result: ", notificationReturn)
		if notificationReturn.CRNFull == strings.ToLower(notification.CRNFull) &&
			notificationReturn.Type == notification.Type &&
			notificationReturn.Category == notification.Category &&
			notificationReturn.SourceID == notification.SourceID &&
			notificationReturn.Source == notification.Source &&
			notificationReturn.Tags == notification.Tags &&
			notificationReturn.ReleaseNoteUrl == notification.ReleaseNoteUrl &&
			len(notificationReturn.ResourceDisplayNames) == 2 && 
			((notificationReturn.ResourceDisplayNames[0].Name == resourceDisplayNames[0].Name && notificationReturn.ResourceDisplayNames[0].Language == resourceDisplayNames[0].Language) || 
			 (notificationReturn.ResourceDisplayNames[0].Name == resourceDisplayNames[1].Name && notificationReturn.ResourceDisplayNames[0].Language == resourceDisplayNames[1].Language)) &&
			((notificationReturn.ResourceDisplayNames[1].Name == resourceDisplayNames[0].Name && notificationReturn.ResourceDisplayNames[1].Language == resourceDisplayNames[0].Language) || 
			 (notificationReturn.ResourceDisplayNames[1].Name == resourceDisplayNames[1].Name && notificationReturn.ResourceDisplayNames[1].Language == resourceDisplayNames[1].Language)) &&
			len(notificationReturn.ShortDescription) == 2 && 
			((notificationReturn.ShortDescription[0].Name == shortDescription[0].Name && notificationReturn.ShortDescription[0].Language == shortDescription[0].Language) || 
			 (notificationReturn.ShortDescription[0].Name == shortDescription[1].Name && notificationReturn.ShortDescription[0].Language == shortDescription[1].Language)) &&
			((notificationReturn.ShortDescription[1].Name == shortDescription[0].Name && notificationReturn.ShortDescription[1].Language == shortDescription[0].Language) || 
			 (notificationReturn.ShortDescription[1].Name == shortDescription[1].Name && notificationReturn.ShortDescription[1].Language == shortDescription[1].Language)) &&
			len(notificationReturn.LongDescription) == 2 && 
			((notificationReturn.LongDescription[0].Name == longDescription[0].Name && notificationReturn.LongDescription[0].Language == longDescription[0].Language) || 
			 (notificationReturn.LongDescription[0].Name == longDescription[1].Name && notificationReturn.LongDescription[0].Language == longDescription[1].Language)) &&
			((notificationReturn.LongDescription[1].Name == longDescription[0].Name && notificationReturn.LongDescription[1].Language == longDescription[0].Language) || 
			 (notificationReturn.LongDescription[1].Name == longDescription[1].Name && notificationReturn.LongDescription[1].Language == longDescription[1].Language)) {

			log.Println("Get Notification passed")
		} else {
			log.Println("Get Notification failed")
			failed++
		}
	}
	


	log.Println("======== "+FCT+" Test 2 ========")

	log.Println("Test InsertNotification - same source_id has 2 crn, GetNotificationByQuery where crn=crn:v1:Bluemix:public:Service-Name1:::::")

	resourceDisplayNames2 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "resource display2, name1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "resource display2, name2",
			Language: "fr",
		},
	}

	shortDescription2 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "short2, description 1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "short2, description 2",
			Language: "fr",
		},
	}

	longDescription2 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "long2, description 1\nWith some longer explanation, of the announcement",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "long2, description 2\nWith some longer explanation, of the announcement",
			Language: "fr",
		},
	}

	notification2 := datastore.NotificationInsert {
		SourceCreationTime:   "2018-04-07T22:01:01Z",
		SourceUpdateTime:     "2018-04-07T22:01:01Z",
		EventTimeStart:       "2018-04-07T21:01:00Z",
		Source:               "test_source1",
		SourceID:             "source_id1",
		Type:                 "announcement",
		Category:             "runtimes",
		CRNFull:              "crn:v1:bLuemix:public:Service-Name1:location2::Service-Instance2:Resource-Type2:",
		ResourceDisplayNames: resourceDisplayNames2,
		ShortDescription:     shortDescription2,
		LongDescription:      longDescription2,
		ReleaseNoteUrl:       "https://cloud.ibm.com/docs/overview?topic=whates-new",
	}

	record_id2, err2,_ := db.InsertNotification(database, &notification2)
	if err2 != nil {
		log.Println("Insert Notification failed: ", err2.Error())
		failed++
	} else {
		log.Println("record_id2: ", record_id2)
		log.Println("Insert Notification passed")
	}
	log.Println("Test get Notification")
	queryStr2 := db.NOTIFICATION_QUERY_CRN + "=crn:v1:Bluemix:public:Service-Name1:::::"
	notificationReturn2, total_count2, err2a,_ := db.GetNotificationByQuery(database, queryStr2, 0, 0)
	log.Println("total_count2: ", total_count2)
	log.Println("notificationReturn2: ", notificationReturn2)

	if err2a != nil {
		log.Println("Get Notification failed: ", err2a.Error())
		failed++
	} else if total_count2 == 2 {

		if len(*notificationReturn2) != 2 {
			log.Printf("Error: Number of returned: %d, expecting: 2\n", len(*notificationReturn2))
			failed++
		} else {
			match := 0
			for _,r := range *notificationReturn2 {
				if r.CRNFull == strings.ToLower(notification.CRNFull) &&
					r.Type == notification.Type &&
					r.Category == notification.Category &&
					r.SourceID == notification.SourceID &&
					r.Source == notification.Source &&
					r.Tags == notification.Tags &&
					r.ReleaseNoteUrl == notification.ReleaseNoteUrl &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription[0].Name && r.ShortDescription[0].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription[1].Name && r.ShortDescription[0].Language == shortDescription[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription[0].Name && r.ShortDescription[1].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription[1].Name && r.ShortDescription[1].Language == shortDescription[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription[0].Name && r.LongDescription[0].Language == longDescription[0].Language) ||
					 (r.LongDescription[0].Name == longDescription[1].Name && r.LongDescription[0].Language == longDescription[1].Language)) &&
					((r.LongDescription[1].Name == longDescription[0].Name && r.LongDescription[1].Language == longDescription[0].Language) ||
					 (r.LongDescription[1].Name == longDescription[1].Name && r.LongDescription[1].Language == longDescription[1].Language)) {

					log.Println("Row 0 matched")
					match++
				}

				if r.CRNFull == strings.ToLower(notification2.CRNFull) &&
					r.Type == notification2.Type &&
					r.Category == notification2.Category &&
					r.SourceID == notification2.SourceID &&
					r.Source == notification2.Source &&
					r.ReleaseNoteUrl == notification2.ReleaseNoteUrl &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription2[0].Name && r.ShortDescription[0].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription2[1].Name && r.ShortDescription[0].Language == shortDescription2[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription2[0].Name && r.ShortDescription[1].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription2[1].Name && r.ShortDescription[1].Language == shortDescription2[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription2[0].Name && r.LongDescription[0].Language == longDescription2[0].Language) ||
					 (r.LongDescription[0].Name == longDescription2[1].Name && r.LongDescription[0].Language == longDescription2[1].Language)) &&
					((r.LongDescription[1].Name == longDescription2[0].Name && r.LongDescription[1].Language == longDescription2[0].Language) ||
					 (r.LongDescription[1].Name == longDescription2[1].Name && r.LongDescription[1].Language == longDescription2[1].Language)) {

					log.Println("Row 1 matched")
					match++
				}
			}
			if (match != 2){
				log.Println("There are rows unmatched, failed")
				failed++
			} else {
				log.Println("Get Notification passed")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count2)
		failed++
	}

	log.Println("======== "+FCT+" Test 3 ========")

	log.Println("Test InsertNotification, GetNotificationByQuery where crn=crn:v1:Bluemix:public::::::&type=announcement, offset=1, limit=3")

	resourceDisplayNames3 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "resource display3, name1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "resource display3, name2",
			Language: "fr",
		},
	}

	shortDescription3 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "short3, description 1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "short3, description 2",
			Language: "fr",
		},
	}

	longDescription3 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "long3, description 1\nWith some longer explanation, of the announcement",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "long3, description 2\nWith some longer explanation, of the announcement",
			Language: "fr",
		},
	}

	notification3 := datastore.NotificationInsert {
		SourceCreationTime:   "2018-07-07T22:01:01Z",
		EventTimeStart:       "2018-07-07T21:01:00Z",
		Source:               "test_source1",
		SourceID:             "source_id3",
		Type:                 "security",
		Category:             "runtimes",
		CRNFull:              "crn:v1:bluemix:public:Service-Name3:location3::Service-Instance3:Resource-Type3:",
		ResourceDisplayNames: resourceDisplayNames3,
		ShortDescription:     shortDescription3,
		LongDescription:      longDescription3,
		ReleaseNoteUrl:       "https://cloud.ibm.com/docs/overview?topic=whates-new",
	}

	record_id3, err3,_ := db.InsertNotification(database, &notification3)
	if err3 != nil {
		log.Println("Insert Notification failed: ", err3.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id3)
		log.Println("Insert Notification passed")
	}
	log.Println("Test get Notification")
	queryStr3 := db.NOTIFICATION_QUERY_CRN + "=crn:v1:Bluemix:public::::::&" + db.NOTIFICATION_COLUMN_TYPE + "=announcement"
	notificationReturn3, total_count3, err3a,_ := db.GetNotificationByQuery(database, queryStr3, 3, 1)
	log.Println("total_count3: ", total_count3)
	log.Println("notificationReturn3: ", notificationReturn3)

	if err3a != nil {
		log.Println("Get Notification failed: ", err3a.Error())
		failed++
	} else if total_count3 == 2 {

		if len(*notificationReturn3) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*notificationReturn3))
			failed++
		} else {
			match := 0
			for _,r := range *notificationReturn3 {
				if r.CRNFull == strings.ToLower(notification.CRNFull) &&
					r.Type == notification.Type &&
					r.Category == notification.Category &&
					r.SourceID == notification.SourceID &&
					r.Source == notification.Source &&
					r.ReleaseNoteUrl == notification.ReleaseNoteUrl &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription[0].Name && r.ShortDescription[0].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription[1].Name && r.ShortDescription[0].Language == shortDescription[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription[0].Name && r.ShortDescription[1].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription[1].Name && r.ShortDescription[1].Language == shortDescription[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription[0].Name && r.LongDescription[0].Language == longDescription[0].Language) ||
					 (r.LongDescription[0].Name == longDescription[1].Name && r.LongDescription[0].Language == longDescription[1].Language)) &&
					((r.LongDescription[1].Name == longDescription[0].Name && r.LongDescription[1].Language == longDescription[0].Language) ||
					 (r.LongDescription[1].Name == longDescription[1].Name && r.LongDescription[1].Language == longDescription[1].Language)) {

					log.Println("Row 0 matched")
					match++
				}

				if r.CRNFull == strings.ToLower(notification2.CRNFull) &&
					r.Type == notification2.Type &&
					r.Category == notification2.Category &&
					r.SourceID == notification2.SourceID &&
					r.Source == notification2.Source &&
					r.ReleaseNoteUrl == notification2.ReleaseNoteUrl &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription2[0].Name && r.ShortDescription[0].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription2[1].Name && r.ShortDescription[0].Language == shortDescription2[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription2[0].Name && r.ShortDescription[1].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription2[1].Name && r.ShortDescription[1].Language == shortDescription2[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription2[0].Name && r.LongDescription[0].Language == longDescription2[0].Language) ||
					 (r.LongDescription[0].Name == longDescription2[1].Name && r.LongDescription[0].Language == longDescription2[1].Language)) &&
					((r.LongDescription[1].Name == longDescription2[0].Name && r.LongDescription[1].Language == longDescription2[0].Language) ||
					 (r.LongDescription[1].Name == longDescription2[1].Name && r.LongDescription[1].Language == longDescription2[1].Language)) {

					log.Println("Row 1 matched")
					match++
				}
			}
			if (match != 1){
				log.Println("There are rows unmatched, failed")
				failed++
			}
			log.Println("Get Notification passed")
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count3)
		failed++
	}
	
	log.Println("======== "+FCT+" Test 4 ========")

	log.Println("Test InsertNotification, GetNotificationByQuery where crn=crn:v1:bluemix:public::::::&type=announcement, offset=1, limit=3")

	resourceDisplayNames4 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "resource display4, name1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "resource display4, name2",
			Language: "fr",
		},
	}

	shortDescription4 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "short4, description 1",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "short4, description 2",
			Language: "fr",
		},
	}

	longDescription4 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "long4, description 1\nWith some longer explanation, of the announcement",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "long4, description 2\nWith some longer explanation, of the announcement",
			Language: "fr",
		},
	}

	notification4 := datastore.NotificationInsert {
		SourceCreationTime:   "2018-08-07T22:01:01Z",
		SourceUpdateTime:     "2018-08-07T22:01:01Z",
		EventTimeStart:       "2018-08-07T21:01:00Z",
		Source:               "test_source1",
		SourceID:             "source_id4",
		Type:                 "announcement",
		Category:             "runtimes",
		CRNFull:              "crn:v1:bluemix:public:Service-Name1:location4::Service-Instance4:Resource-Type4:",
		ResourceDisplayNames: resourceDisplayNames4,
		ShortDescription:     shortDescription4,
		LongDescription:      longDescription4,
		ReleaseNoteUrl:       "https://cloud.ibm.com/docs/overview?topic=whates-new",
	}

	record_id4, err4,_ := db.InsertNotification(database, &notification4)
	if err4 != nil {
		log.Println("Insert Notification failed: ", err4.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id4)
		log.Println("Insert Notification passed")
	}
	log.Println("Test get Notification")
	queryStr4 := db.NOTIFICATION_QUERY_CRN + "=crn:v1:bluemix:public::::::&" + db.NOTIFICATION_COLUMN_TYPE + "=announcement"
	notificationReturn4, total_count4, err4a,_ := db.GetNotificationByQuery(database, queryStr4, 2, 1)
	log.Println("total_count4: ", total_count4)
	log.Println("notificationReturn4: ", notificationReturn4)

	if err4a != nil {
		log.Println("Get Notification failed: ", err4a.Error())
		failed++
	} else if total_count4 == 3 {

		if len(*notificationReturn4) != 2 {
			log.Printf("Error: Number of returned: %d, expecting: 2\n", len(*notificationReturn4))
			failed++
		} else {
			match := 0
			for _,r := range *notificationReturn4 {
				if r.CRNFull == strings.ToLower(notification.CRNFull) &&
					r.Type == notification.Type &&
					r.Category == notification.Category &&
					r.SourceID == notification.SourceID &&
					r.Source == notification.Source &&
					r.ReleaseNoteUrl ==  notification.ReleaseNoteUrl&&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription[0].Name && r.ShortDescription[0].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription[1].Name && r.ShortDescription[0].Language == shortDescription[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription[0].Name && r.ShortDescription[1].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription[1].Name && r.ShortDescription[1].Language == shortDescription[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription[0].Name && r.LongDescription[0].Language == longDescription[0].Language) ||
					 (r.LongDescription[0].Name == longDescription[1].Name && r.LongDescription[0].Language == longDescription[1].Language)) &&
					((r.LongDescription[1].Name == longDescription[0].Name && r.LongDescription[1].Language == longDescription[0].Language) ||
					 (r.LongDescription[1].Name == longDescription[1].Name && r.LongDescription[1].Language == longDescription[1].Language)) {

					log.Println("Row 0 matched")
					match++
				}

				if r.CRNFull == strings.ToLower(notification2.CRNFull) &&
					r.Type == notification2.Type &&
					r.Category == notification2.Category &&
					r.SourceID == notification2.SourceID &&
					r.Source == notification2.Source &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription2[0].Name && r.ShortDescription[0].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription2[1].Name && r.ShortDescription[0].Language == shortDescription2[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription2[0].Name && r.ShortDescription[1].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription2[1].Name && r.ShortDescription[1].Language == shortDescription2[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription2[0].Name && r.LongDescription[0].Language == longDescription2[0].Language) ||
					 (r.LongDescription[0].Name == longDescription2[1].Name && r.LongDescription[0].Language == longDescription2[1].Language)) &&
					((r.LongDescription[1].Name == longDescription2[0].Name && r.LongDescription[1].Language == longDescription2[0].Language) ||
					 (r.LongDescription[1].Name == longDescription2[1].Name && r.LongDescription[1].Language == longDescription2[1].Language)) {

					log.Println("Row 1 matched")
					match++
				}

				if r.CRNFull == strings.ToLower(notification4.CRNFull) &&
					r.Type == notification4.Type &&
					r.Category == notification4.Category &&
					r.SourceID == notification4.SourceID &&
					r.Source == notification4.Source &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames4[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames4[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames4[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames4[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames4[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames4[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames4[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames4[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription4[0].Name && r.ShortDescription[0].Language == shortDescription4[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription4[1].Name && r.ShortDescription[0].Language == shortDescription4[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription4[0].Name && r.ShortDescription[1].Language == shortDescription4[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription4[1].Name && r.ShortDescription[1].Language == shortDescription4[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription4[0].Name && r.LongDescription[0].Language == longDescription4[0].Language) ||
					 (r.LongDescription[0].Name == longDescription4[1].Name && r.LongDescription[0].Language == longDescription4[1].Language)) &&
					((r.LongDescription[1].Name == longDescription4[0].Name && r.LongDescription[1].Language == longDescription4[0].Language) ||
					 (r.LongDescription[1].Name == longDescription4[1].Name && r.LongDescription[1].Language == longDescription4[1].Language)) {

					log.Println("Row 2 matched")
					match++
				}
			}
			if (match != 2){
				log.Println("There are rows unmatched, failed")
				failed++
			}
			log.Println("Get Notification passed")
		}
	} else {
		log.Printf("total_count is %d, expecting 3, failed", total_count4)
		failed++
	}

	log.Println("======== "+FCT+" Test 5 ========")

	log.Println("Test InsertNotification, GetNotificationByQuery where cname=bluemix...not wildcard, offset=0, limit=0")

	resourceDisplayNames5 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "resource display5, name1",
			Language: "en",
		},
	}

	shortDescription5 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "short5, description 1",
			Language: "en",
		},
	}

	longDescription5 := []datastore.DisplayName{
		datastore.DisplayName {
			Name: "long5, description 1\nWith some longer explanation, of the announcement",
			Language: "en",
		},
	}

	notification5 := datastore.NotificationInsert {
		SourceCreationTime:   "2018-08-17T22:01:01Z",
		SourceUpdateTime:     "2018-08-17T22:01:01Z",
		EventTimeStart:       "2018-08-17T21:01:00Z",
		Source:               "test_source1",
		SourceID:             "source_id5",
		IncidentID:           "inc0001",
		Type:                 "incident",
		Category:             "services",
		CRNFull:              "crn:v1:aa:dedicated:Service-Name5:location5::Service-Instance5::",
		ResourceDisplayNames: resourceDisplayNames5,
		ShortDescription:     shortDescription5,
		LongDescription:      longDescription5,
	}

	record_id5, err5,_ := db.InsertNotification(database, &notification5)
	if err5 != nil {
		log.Println("Insert Notification failed: ", err5.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id5)
		log.Println("Insert Notification passed")
	}
	log.Println("Test get Notification")
	queryStr5 := db.NOTIFICATION_QUERY_VERSION + "=v1&" + db.NOTIFICATION_QUERY_CNAME + "=AA&" + db.NOTIFICATION_QUERY_CTYPE + "=Dedicated&" + db.NOTIFICATION_QUERY_SERVICE_NAME + "=Service-Name5&" +
	db.NOTIFICATION_QUERY_LOCATION + "=location5&" + db.NOTIFICATION_QUERY_SERVICE_INSTANCE + "=Service-Instance5&" + db.NOTIFICATION_QUERY_RESOURCE_TYPE + "=&" +
	db.NOTIFICATION_QUERY_RESOURCE + "="
	notificationReturn5, total_count5, err5a,_ := db.GetNotificationByQuery(database, queryStr5, 0, 0)
	log.Println("total_count5: ", total_count5)
	log.Println("notificationReturn5: ", notificationReturn5)

	if err5a != nil {
		log.Println("Get Notification failed: ", err5a.Error())
		failed++
	} else if total_count5 == 1 {

		if len(*notificationReturn5) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*notificationReturn5))
			failed++
		} else {
			match := 0
			for _,r := range *notificationReturn5 {
				if r.CRNFull == strings.ToLower(notification5.CRNFull) &&
					r.Type == notification5.Type &&
					r.Category == notification5.Category &&
					r.SourceID == notification5.SourceID &&
					r.Source == notification5.Source &&
					len(r.ResourceDisplayNames) == 1 &&
					(r.ResourceDisplayNames[0].Name == resourceDisplayNames5[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames5[0].Language) &&
					len(r.ShortDescription) == 1 &&
					(r.ShortDescription[0].Name == shortDescription5[0].Name && r.ShortDescription[0].Language == shortDescription5[0].Language) &&
					len(r.LongDescription) == 1 &&
					(r.LongDescription[0].Name == longDescription5[0].Name && r.LongDescription[0].Language == longDescription5[0].Language) {

					log.Println("Row 0 matched")
					match++
				}
			}
			if (match != 1){
				log.Println("There are rows unmatched, failed")
				failed++
			}
			log.Println("Get Notification passed")
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count5)
		failed++
	}


	// Update PnPRemoved=true
	notification5.PnPRemoved = true

	err5b,_ := db.UpdateNotification(database, &notification5)
	if err5b != nil {
		failed++
		log.Println("Update Notification failed")
	}
	notificationReturn5c, total_count5c, err5c,_ := db.GetNotificationByQuery(database, queryStr5, 0, 0)
	log.Println("total_count5c: ", total_count5c)
	log.Println("notificationReturn5c: ", notificationReturn5c)

	if err5c != nil {
		log.Println("Get Notification failed: ", err5c.Error())
		failed++
	} else if total_count5c == 0 {

		log.Println("Get Notification passed")

	} else {
		log.Printf("total_count is %d, expecting 0, failed", total_count5c)
		failed++
	}


	log.Println("======== "+FCT+" Test 6 ========")

	log.Println("Test GetNotificationByQuery - query not found")

	queryStr6 := db.NOTIFICATION_QUERY_CRN + "=crn:v1:bluemix:public::::::&" + db.NOTIFICATION_COLUMN_TYPE + "=incident"
	log.Println("query: ", queryStr6)
	notificationReturn6, total_count6, err6,_ := db.GetNotificationByQuery(database, queryStr6, 0, 0)
	if err6 != nil {
		log.Println("Get Notification failed: ", err6.Error())
		failed++
	} else if total_count6 !=0 {
		log.Printf("Error: total_count: %d, expecting: 0\n", total_count6)
		failed++
	} else if len(*notificationReturn6) != 0 {
		log.Printf("Error: Number of returned: %d, expecting: 0\n", len(*notificationReturn6))
		failed++
	} else {
		log.Println("Get Notification passed")
	}

	log.Println("======== "+FCT+" Test 7 ========")

	log.Println("Test GetResourceByQuery where source=source1&source_id=source_id1")

	log.Println("Test get Notification")
	queryStr7 := db.NOTIFICATION_COLUMN_SOURCE + "=test_source1&" + db.NOTIFICATION_COLUMN_SOURCE_ID + "=source_id1"
	notificationReturn7, total_count7, err7a,_ := db.GetNotificationByQuery(database, queryStr7, 0, 0)
	log.Println("total_count7: ", total_count7)
	log.Println("notificationReturn7: ", notificationReturn7)

	if err7a != nil {
		log.Println("Get Notification failed: ", err7a.Error())
		failed++
	} else if total_count7 == 2 {

		if len(*notificationReturn7) != 2 {
			log.Printf("Error: Number of returned: %d, expecting: 2\n", len(*notificationReturn7))
			failed++
		} else {
			match := 0
			for _,r := range *notificationReturn7 {
				if r.CRNFull == strings.ToLower(notification.CRNFull) &&
					r.Type == notification.Type &&
					r.Category == notification.Category &&
					r.SourceID == notification.SourceID &&
					r.Source == notification.Source &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription[0].Name && r.ShortDescription[0].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription[1].Name && r.ShortDescription[0].Language == shortDescription[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription[0].Name && r.ShortDescription[1].Language == shortDescription[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription[1].Name && r.ShortDescription[1].Language == shortDescription[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription[0].Name && r.LongDescription[0].Language == longDescription[0].Language) ||
					 (r.LongDescription[0].Name == longDescription[1].Name && r.LongDescription[0].Language == longDescription[1].Language)) &&
					((r.LongDescription[1].Name == longDescription[0].Name && r.LongDescription[1].Language == longDescription[0].Language) ||
					 (r.LongDescription[1].Name == longDescription[1].Name && r.LongDescription[1].Language == longDescription[1].Language)) {

					log.Println("Row 0 matched")
					match++
				}

				if  r.CRNFull == strings.ToLower(notification2.CRNFull) &&
					r.Type == notification2.Type &&
					r.Category == notification2.Category &&
					r.SourceID == notification2.SourceID &&
					r.Source == notification2.Source &&
					len(r.ResourceDisplayNames) == 2 &&
					((r.ResourceDisplayNames[0].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[0].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[0].Language == resourceDisplayNames2[1].Language)) &&
					((r.ResourceDisplayNames[1].Name == resourceDisplayNames2[0].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[0].Language) ||
					 (r.ResourceDisplayNames[1].Name == resourceDisplayNames2[1].Name && r.ResourceDisplayNames[1].Language == resourceDisplayNames2[1].Language)) &&
					len(r.ShortDescription) == 2 &&
					((r.ShortDescription[0].Name == shortDescription2[0].Name && r.ShortDescription[0].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[0].Name == shortDescription2[1].Name && r.ShortDescription[0].Language == shortDescription2[1].Language)) &&
					((r.ShortDescription[1].Name == shortDescription2[0].Name && r.ShortDescription[1].Language == shortDescription2[0].Language) ||
					 (r.ShortDescription[1].Name == shortDescription2[1].Name && r.ShortDescription[1].Language == shortDescription2[1].Language)) &&
					len(r.LongDescription) == 2 &&
					((r.LongDescription[0].Name == longDescription2[0].Name && r.LongDescription[0].Language == longDescription2[0].Language) ||
					 (r.LongDescription[0].Name == longDescription2[1].Name && r.LongDescription[0].Language == longDescription2[1].Language)) &&
					((r.LongDescription[1].Name == longDescription2[0].Name && r.LongDescription[1].Language == longDescription2[0].Language) ||
					 (r.LongDescription[1].Name == longDescription2[1].Name && r.LongDescription[1].Language == longDescription2[1].Language)) {

					log.Println("Row 1 matched")
					match++
				}
			}
			if (match != 2){
				log.Println("There are rows unmatched, failed")
				failed++
			}
			log.Println("Get Notification passed")
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count2)
		failed++
	}

	log.Println("======== "+FCT+" Test 8 ========")

	log.Println("Test Update Notification - Category, ShortDescription, LongDescription")

	notification.ShortDescription = []datastore.DisplayName{
		datastore.DisplayName {
			Name: "short, description 2",
			Language: "fr",
		},
		datastore.DisplayName {
			Name: "Third-party service deprecations: MMMMMM, FFFFFFFFF.io, ABC and DDDDDDD",
			Language: "en",
		},
	}

	notification.LongDescription = []datastore.DisplayName{
		datastore.DisplayName {
			Name: "We would like to inform you that we are retiring the following services from the MMM on October 27, 2018:\n<ul><li>MMMMMM</li><li>FFFFFFFFF.io</li><li>DDDDDDD</li></ul>\nFor more information on how to continue to use MMMMMM services, see <u><a href=\"https://mmmm.com/\" target=\"_blank\">https://mmmm.com/</a></u>.<br><br>Here’s what you need to know:<ul><li>You cannot provision new service instances. However, existing instances will continue to be supported until the End of Support date.</li><li>The End of Support date is October 27, 2018.</li><li>Through October 27, 2018, all existing instances will continue to be available through the command line.</li><li>Any instance of these services that is still provisioned as of the End of Support date will be deleted.</li><li>Delete your service instances before the End of Support date.</li></ul>\nThis information originated in the <u><a href=\"https://www.mmm.com/blog/test/\" target=\"_blank\">Deprecations</a></u> article within the MMM blog.",
			Language: "en",
		},
		datastore.DisplayName {
			Name: "long, description 1\nWith some longer explanation, of the announcement",
			Language: "es",
		},
	}

	notification.Category = "platform"
	notification.PnPRemoved = true

	err8,_ := db.UpdateNotification(database, &notification)
	if err8 != nil {
		failed++
		log.Println("Update Notification failed")
	} else {
		nGet8a, err8a, _ := db.GetNotificationByRecordID(database, record_id)
		log.Println("nGet8a: ", nGet8a)

		if err8a==nil && nGet8a != nil {
			if nGet8a.Category == notification.Category && nGet8a.PnPRemoved == true &&
				((nGet8a.ShortDescription[0].Name == notification.ShortDescription[0].Name && nGet8a.ShortDescription[0].Language == notification.ShortDescription[0].Language) ||
				 (nGet8a.ShortDescription[0].Name == notification.ShortDescription[1].Name && nGet8a.ShortDescription[0].Language == notification.ShortDescription[1].Language)) &&
				((nGet8a.ShortDescription[1].Name == notification.ShortDescription[0].Name && nGet8a.ShortDescription[1].Language == notification.ShortDescription[0].Language) ||
				 (nGet8a.ShortDescription[1].Name == notification.ShortDescription[1].Name && nGet8a.ShortDescription[1].Language == notification.ShortDescription[1].Language)) &&
				((nGet8a.LongDescription[0].Name == notification.LongDescription[0].Name && nGet8a.LongDescription[0].Language == notification.LongDescription[0].Language) ||
				 (nGet8a.LongDescription[0].Name == notification.LongDescription[1].Name && nGet8a.LongDescription[0].Language == notification.LongDescription[1].Language)) &&
				((nGet8a.LongDescription[1].Name == notification.LongDescription[0].Name && nGet8a.LongDescription[1].Language == notification.LongDescription[0].Language) ||
				 (nGet8a.LongDescription[1].Name == notification.LongDescription[1].Name && nGet8a.LongDescription[1].Language == notification.LongDescription[1].Language)) {

				log.Println("Update Notification passed")
			} else {
				failed++
				log.Println("Update Notification failed")
			}
		} else {
			failed++
			log.Println("Update Notification failed, err8a: ", err8a.Error())
		}
	}


	log.Println("======== "+FCT+" Test 9 ========")

	log.Println("Test DeleteOldAnnouncmentNotifications")

	err9, _ := db.DeleteOldAnnouncmentNotifications(database, 90)
	if err9 != nil {
		log.Println("DeleteOldAnnouncmentNotifications failed: ", err9.Error())
	}

	nGet9, _, _ := db.GetNotificationByRecordID(database, record_id2)
	if nGet9 != nil {
		//record_id2 should have been deleted from database
		failed++
		log.Println("Test 9: DeleteOldAnnouncmentNotifications failed")
	}

	if nGet9 == nil {
		log.Println("Test 9: DeleteOldAnnouncmentNotifications passed")
	}



	log.Println("======== "+FCT+" Test 10 ========")

	log.Println("Test DeleteOldIncidentNotifications")

	err10, _ := db.DeleteOldIncidentNotifications(database, 30)
	if err10 != nil {
		log.Println("DeleteOldIncidentNotifications failed: ", err10.Error())
	}

	nGet10, _, _ := db.GetNotificationByRecordID(database, record_id5)
	if nGet10 != nil {
		//record_id5 should have been deleted from database
		failed++
		log.Println("Test 10: DeleteOldIncidentNotifications failed")
	}

	if nGet10 == nil {
		log.Println("Test 10: DeleteOldIncidentNotifications passed")
	}


	log.Println("======== "+FCT+" Test 11 ========")

	log.Println("Test DeleteOldMaintenanceNotifications")

	notification11 := datastore.NotificationInsert {
		SourceCreationTime:   "2018-08-07T22:01:01Z",
		SourceUpdateTime:     "2018-08-07T22:01:01Z",
		EventTimeStart:       "2018-08-07T21:01:00Z",
		Source:               "test_source11",
		SourceID:             "source_id11",
		Type:                 "maintenance",
		Category:             "runtimes",
		CRNFull:              "crn:v1:bluemix:public:Service-Name1:location4::Service-Instance4:Resource-Type4:",
		ResourceDisplayNames: resourceDisplayNames4,
		ShortDescription:     shortDescription4,
		LongDescription:      longDescription4,
	}

	record_id11, err11,_ := db.InsertNotification(database, &notification11)
	if err11 != nil {
		log.Println("Insert Notification failed: ", err11.Error())
		failed++
	} else {
		log.Println("record_id: ", record_id11)
		log.Println("Insert Notification passed")
	}

	err11b, _ := db.DeleteOldMaintenanceNotifications(database, 30)
	if err11b != nil {
		log.Println("DeleteOldMaintenanceNotifications failed: ", err11b.Error())
	}

	nGet11, _, _ := db.GetNotificationByRecordID(database, record_id11)
	if nGet11 != nil {
		//record_id4 should have been deleted from database
		failed++
		log.Println("Test 11: DeleteOldMaintenanceNotifications failed")
	}

	if nGet11 == nil {
		log.Println("Test 11: DeleteOldMaintenanceNotifications passed")
	}


/*
	log.Println("======== "+FCT+" Test 10 ========")
	// check panic error in getting Notification LongDescription
	rows, err := database.Query("SELECT record_id FROM notification_table")
	rowsReturned := 0

	switch {
	case err == sql.ErrNoRows:
		log.Println(FCT + "No records found.")
	case err != nil:
		log.Println(FCT+"Error : ", err)
	default:
		defer rows.Close()

		recordsGet := []datastore.NotificationGet{}
		for rows.Next() {
			rowsReturned++
			r := datastore.NotificationGet{}

			err = rows.Scan(&r.RecordID)
			if err != nil {
				log.Printf(FCT+"Row Scan Error: %v", err)
				break
			}
			recordsGet = append(recordsGet, r)
		}
		for i := 0; i < len(recordsGet); i++ {
			//log.Println(FCT + "---------------------- recordID: ", recordsGet[i].RecordID)
			notificationReturn, err1, _ := db.GetNotificationByRecordID(database, recordsGet[i].RecordID)

			if err1 != nil {
				log.Println("Get Notification failed: record_id: ", recordsGet[i].RecordID, ", ", err1.Error())
				failed++
			} else {
				//log.Println("Get result: ", notificationReturn)
				//log.Println(FCT + "len(notificationReturn.LongDescription): ", len(notificationReturn.LongDescription))
				for j:=0; j< len(notificationReturn.LongDescription); j++{
					log.Println("LongDescription.Name", notificationReturn.LongDescription[j].Name)
					log.Println("LongDescription.Language", notificationReturn.LongDescription[j].Language)
				}
			}
		}
	}
*/

	log.Println("**** SUMMARY ****")
	if failed > 0 {
		log.Println("**** "+FCT+"FAILED ****")
	} else {
		log.Println("**** "+FCT+"PASSED ****")
	}

	return failed
}

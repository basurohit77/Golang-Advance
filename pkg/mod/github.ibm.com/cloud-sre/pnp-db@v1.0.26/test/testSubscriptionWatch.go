package test

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.ibm.com/cloud-sre/pnp-abstraction/db"
	"github.ibm.com/cloud-sre/pnp-abstraction/datastore"
)

const (
	API_URL = "apiUrl"
)

func TestSubscriptionWatch(database *sql.DB) int {
	FCT := "TestSubscriptionWatch "
	failed := 0

	log.Println("======== "+FCT+" Test 1 ========")
	subscription := datastore.SubscriptionInsert {
		Name:            "test_name1",
		TargetAddress:   "target1",
		TargetToken:     "token1",
		Expiration:      "2018-08-01T10:00:00Z",
	}

	subscription_id, err, status := db.InsertSubscriptionStatement(database, &subscription)
	if status != http.StatusOK {
		log.Println("Insert Subscription failed: ", err.Error())
		failed++
	} else {
		log.Println("record_id: ", subscription_id)
		log.Println("Insert Subscription passed")
	}


	log.Println("Test get Subscription")
	subscriptionReturn, err1,_ := db.GetSubscriptionByRecordIDStatement(database, subscription_id)
	if err1 != nil {
		log.Println("Get Subscription failed: ", err1.Error())
		failed++
	} else {

		log.Println("Get result: ", subscriptionReturn)

		if subscriptionReturn.Name == subscription.Name &&
			subscriptionReturn.TargetAddress == subscription.TargetAddress &&
			subscriptionReturn.TargetToken == subscription.TargetToken { //&&
//			subscriptionReturn.Expiration == subscription.Expiration  {

			log.Println("Test 1: Get Subscription PASSED")
		} else {
			log.Println("Test 1: Get Subscription FAILED")
			failed++
		}

	}


	log.Println("======== "+FCT+" Test 2 ========")
	crnFull2 := []string{"crn:v1:BLUEMIX:PUBLIC:service-NAME2:location2::::",
						 "crn:v1:aa:dedicated:service-name2:LOCATION3::::",
				}
	recordIDToWatch2 := []string {"INC00001"}

	watch2 := datastore.WatchInsert {
		SubscriptionRecordID: subscription_id,
		Kind:                 "incident",
		Path:                 "new",
		CRNFull:              crnFull2,
		Wildcards:            "true",
		RecordIDToWatch:      recordIDToWatch2,
		SubscriptionEmail:	  "tester1@ibm.com",
	}

	watchReturn2, err2,_ := db.InsertWatchByTypeForSubscriptionStatement(database, &watch2, API_URL)
	if err2 != nil {
		log.Println("Insert Watch failed: ", err2.Error())
		failed++
	} else {
		log.Println("record_id: ", watchReturn2.RecordID)
		log.Println("Insert Watch passed")
	}

	log.Println("Test get Watch")
	watchReturn2b, err2b,_ := db.GetWatchByRecordIDStatement(database, watchReturn2.RecordID, API_URL)
	if err2b != nil {
		log.Println("Get Watch failed: ", err2b.Error())
		failed++
	} else {
		log.Println("Get result: ", watchReturn2b)

		if strings.Contains(watchReturn2b.SubscriptionURL.URL, watch2.SubscriptionRecordID) &&
			watchReturn2b.Kind == watch2.Kind+"Watch" &&
			watchReturn2b.Path == watch2.Path &&
			len(watchReturn2b.CRNFull) == len(watch2.CRNFull) &&
			watchReturn2b.CRNFull[0] == strings.ToLower(watch2.CRNFull[0]) &&
			watchReturn2b.CRNFull[1] == strings.ToLower(watch2.CRNFull[1]) &&
			watchReturn2b.Wildcards == watch2.Wildcards &&
			len(watchReturn2b.RecordIDToWatch) == len(watch2.RecordIDToWatch) &&
			watchReturn2b.RecordIDToWatch[0] == watch2.RecordIDToWatch[0] &&
			watchReturn2b.SubscriptionEmail == watch2.SubscriptionEmail {

			log.Println("Test 2: Get Watch PASSED")
		} else {
			log.Println("Test 2: Get Watch FAILED")
			failed++
		}
	}

	log.Println("======== "+FCT+" Test 3 ========")
	crnFull3 := []string{"crn:v1:bluemix:public:service-name1:location1::service-instance1:resource-type1:",
				}
	recordIDToWatch3 := []string {db.CreateRecordIDFromSourceSourceID("source1", "INC00011")}

	watch3 := datastore.WatchInsert {
		SubscriptionRecordID: subscription_id,
		Kind:                 "maintenance",
		Path:                 "new",
		CRNFull:              crnFull3,
		Wildcards:            "false",
		RecordIDToWatch:      recordIDToWatch3,
		SubscriptionEmail:	  "tester2@ibm.com",
	}

	watchReturn3, err3,_ := db.InsertWatchByTypeForSubscriptionStatement(database, &watch3, API_URL)
	if err3 != nil {
		log.Println("Insert Watch failed: ", err3.Error())
		failed++
	} else {
		log.Println("record_id: ", watchReturn3.RecordID)
		log.Println("Insert Watch passed")
	}

	log.Println("Test get Watch")
	watchReturn3b, err3b,_ := db.GetWatchByRecordIDStatement(database, watchReturn3.RecordID, API_URL)
	if err3b != nil {
		log.Println("Get Watch failed: ", err3b.Error())
		failed++
	} else {
		log.Println("Get result: ", watchReturn3b)

		if strings.Contains(watchReturn3b.SubscriptionURL.URL, watch3.SubscriptionRecordID) &&
			watchReturn3b.Kind == watch3.Kind+"Watch" &&
			watchReturn3b.Path == watch3.Path &&
			len(watchReturn3b.CRNFull) == len(watch3.CRNFull) &&
			watchReturn3b.CRNFull[0] == strings.ToLower(watch3.CRNFull[0]) &&
			watchReturn3b.Wildcards == watch3.Wildcards &&
			len(watchReturn3b.RecordIDToWatch) == len(watch3.RecordIDToWatch) &&
			watchReturn3b.RecordIDToWatch[0] == watch3.RecordIDToWatch[0] &&
			watchReturn3b.SubscriptionEmail == watch3.SubscriptionEmail {

			log.Println("Test 3: Get Watch PASSED")
		} else {
			log.Println("Test 3: Get Watch FAILED")
			failed++
		}
	}


	log.Println("======== "+FCT+" Test 4 ========")
	queryStr4 := db.RESOURCE_QUERY_CRN + "=crn:v1:aa:dedicated:service-name2:LOCation3::::"
	log.Println("query: ", queryStr4)
	watchReturn4, total_count4, err4,_ := db.GetWatchesByQuery(database, queryStr4, 0, 0, API_URL)

	if err4 != nil {
		log.Println("Get Watches failed: ", err4.Error())
		log.Println("Test 4: Get Watches FAILED")
		failed++
	} else if total_count4 == 1 {
		log.Println("total_count4: ", total_count4)
		log.Println("watchReturn4: ", watchReturn4)

		if len(*watchReturn4) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*watchReturn4))
			log.Println("Test 4: Get Watches FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *watchReturn4 {

				if strings.Contains(r.SubscriptionURL.URL, watch2.SubscriptionRecordID) &&
					r.Kind == watch2.Kind+"Watch" &&
					r.Path == watch2.Path &&
					len(r.CRNFull) == len(watch2.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(watch2.CRNFull[0]) &&
					r.CRNFull[1] == strings.ToLower(watch2.CRNFull[1]) &&
					r.Wildcards == watch2.Wildcards &&
					len(r.RecordIDToWatch) == len(watch2.RecordIDToWatch) &&
					r.RecordIDToWatch[0] == watch2.RecordIDToWatch[0] &&
					r.SubscriptionEmail == watch2.SubscriptionEmail {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 4: Get Watches FAILED")
				failed++
			} else {
				log.Println("Test 4: Get Watches PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count4)
		log.Println("Test 4: Get Watches FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 5 ========")
	queryStr5, err5a := db.CreateNonWildcardCRNQueryString(crnFull3[0])
	if err5a != nil {
		log.Println("Error: ", err5a)
	}
	log.Println("query: ", queryStr5)

	watchReturn5, total_count5, err5,_ := db.GetWatchesByQuery(database, queryStr5, 0, 0, API_URL)

	if err5 != nil {
		log.Println("Get Watches failed: ", err5.Error())
		log.Println("Test 5: Get Watches FAILED")
		failed++
	} else if total_count5 == 1 {
		log.Println("total_count5: ", total_count5)
		log.Println("watchReturn5: ", watchReturn5)

		if len(*watchReturn5) != 1 {
			log.Printf("Error: Number of returned: %d, expecting: 1\n", len(*watchReturn5))
			log.Println("Test 5: Get Watches FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *watchReturn5 {

				if strings.Contains(r.SubscriptionURL.URL, watch3.SubscriptionRecordID) &&
					r.Kind == watch3.Kind+"Watch" &&
					r.Path == watch3.Path &&
					len(r.CRNFull) == len(watch3.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(watch3.CRNFull[0]) &&
					r.Wildcards == watch3.Wildcards &&
					len(r.RecordIDToWatch) == len(watch3.RecordIDToWatch) &&
					r.RecordIDToWatch[0] == watch3.RecordIDToWatch[0] &&
					r.SubscriptionEmail == watch3.SubscriptionEmail {

					match++
				}
			}

			if (match <= 0){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 5: Get Watches FAILED")
				failed++
			} else {
				log.Println("Test 5: Get Watches PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 1, failed", total_count5)
		log.Println("Test 5: Get Watches FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 6 ========")
	subscription6 := datastore.SubscriptionInsert {
		Name:            "test_name6",
		TargetAddress:   "target6",
		TargetToken:     "token6",
		Expiration:      "2018-12-01T06:06:00Z",
	}

	subscription_id6, err6,_ := db.InsertSubscriptionStatement(database, &subscription6)
	if err6 != nil {
		log.Println("Insert Subscription failed: ", err6.Error())
		failed++
	} else {
		log.Println("record_id: ", subscription_id6)
		log.Println("Insert Subscription passed")
	}


	log.Println("Test get Subscription")
	subscriptionReturn6, err61,_ := db.GetSubscriptionByRecordIDStatement(database, subscription_id6)
	if err61 != nil {
		log.Println("Get Subscription failed: ", err61.Error())
		failed++
	} else {

		log.Println("Get result: ", subscriptionReturn6)

		if subscriptionReturn6.Name == subscription6.Name &&
			subscriptionReturn6.TargetAddress == subscription6.TargetAddress &&
			subscriptionReturn6.TargetToken == subscription6.TargetToken { //&&
			//subscriptionReturn6.Expiration == subscription6.Expiration  {

			log.Println("Get Subscription passed")
		} else {
			log.Println("Get Subscription failed")
			failed++
		}

	}

	crnFull6 := []string{"crn:v1:blueMIX:public:service-naME1:location1::::",
				}
	recordIDToWatch6 := []string {db.CreateRecordIDFromSourceSourceID("source1", "INC00006")}

	watch6 := datastore.WatchInsert {
		SubscriptionRecordID: subscription_id6,
		Kind:                 "incident",
		Path:                 "new",
		CRNFull:              crnFull6,
		Wildcards:            "true",
		RecordIDToWatch:      recordIDToWatch6,
		SubscriptionEmail: 	  "tester3@ibm.com",
	}

	watchReturn6, err62,_ := db.InsertWatchByTypeForSubscriptionStatement(database, &watch6, API_URL)
	if err62 != nil {
		log.Println("Insert Watch failed: ", err62.Error())
		failed++
	} else {
		log.Println("record_id: ", watchReturn6.RecordID)
		log.Println("Insert Watch passed")
	}

	log.Println("Test get Watch")
	queryStr6 := db.RESOURCE_QUERY_CRN + "=crn:v1:bluemix:public:service-name1:::::&" + db.WATCH_QUERY_KIND + "=incident,maintenance"
	log.Println("query: ", queryStr6)
	watchReturn6b, total_count6, err6b,_ := db.GetWatchesByQuery(database, queryStr6, 0, 0, API_URL)

	if err6b != nil {
		log.Println("Get Watches failed: ", err6b.Error())
		log.Println("Test 6: Get Watches FAILED")
		failed++
	} else if total_count6 == 2 {
		log.Println("total_count6: ", total_count6)
		log.Println("watchReturn6: ", watchReturn6b)

		if len(*watchReturn6b) != 2 {
			log.Printf("Error: Number of returned: %d, expecting: 2\n", len(*watchReturn6b))
			log.Println("Test 6: Get Watches FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *watchReturn6b {

				if strings.Contains(r.SubscriptionURL.URL, watch3.SubscriptionRecordID) &&
					r.Kind == watch3.Kind+"Watch" &&
					r.Path == watch3.Path &&
					len(r.CRNFull) == len(watch3.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(watch3.CRNFull[0]) &&
					r.Wildcards == watch3.Wildcards &&
					len(r.RecordIDToWatch) == len(watch3.RecordIDToWatch) &&
					r.RecordIDToWatch[0] == watch3.RecordIDToWatch[0] &&
					r.SubscriptionEmail == watch3.SubscriptionEmail {

					match++
				} else if strings.Contains(r.SubscriptionURL.URL, watch6.SubscriptionRecordID) &&
					r.Kind == watch6.Kind+"Watch" &&
					r.Path == watch6.Path &&
					len(r.CRNFull) == len(watch6.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(watch6.CRNFull[0]) &&
					r.Wildcards == watch6.Wildcards &&
					len(r.RecordIDToWatch) == len(watch6.RecordIDToWatch) &&
					r.RecordIDToWatch[0] == watch6.RecordIDToWatch[0] &&
					r.SubscriptionEmail == watch6.SubscriptionEmail {

					match++
				}
			}

			if (match <= 1){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 6: Get Watches FAILED")
				failed++
			} else {
				log.Println("Test 6: Get Watches PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count6)
		log.Println("Test 6: Get Watches FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 7 ========")
	log.Println("Test get Watch")
	queryStr7 := db.WATCH_QUERY_KIND + "=incident,maintenance&" + db.WATCH_QUERY_SUBSCRIPTION_ID + "=" + subscription_id
	log.Println("query: ", queryStr7)
	watchReturn7, total_count7, err7,_ := db.GetWatchesByQuery(database, queryStr7, 0, 0, API_URL)

	if err7 != nil {
		log.Println("Get Watches failed: ", err7.Error())
		log.Println("Test 7: Get Watches FAILED")
		failed++
	} else if total_count7 == 2 {
		log.Println("total_count7: ", total_count7)
		log.Println("watchReturn7: ", watchReturn7)

		if len(*watchReturn7) != 2 {
			log.Printf("Error: Number of returned: %d, expecting: 2\n", len(*watchReturn7))
			log.Println("Test 7: Get Watches FAILED")
			failed++
		} else {
			match := 0
			for _,r := range *watchReturn7 {

				if strings.Contains(r.SubscriptionURL.URL, watch2.SubscriptionRecordID) &&
					r.Kind == watch2.Kind+"Watch" &&
					r.Path == watch2.Path &&
					len(r.CRNFull) == len(watch2.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(watch2.CRNFull[0]) &&
					r.Wildcards == watch2.Wildcards &&
					len(r.RecordIDToWatch) == len(watch2.RecordIDToWatch) &&
					r.RecordIDToWatch[0] == watch2.RecordIDToWatch[0] &&
					r.SubscriptionEmail == watch2.SubscriptionEmail {

					match++
				} else if strings.Contains(r.SubscriptionURL.URL, watch3.SubscriptionRecordID) &&
					r.Kind == watch3.Kind+"Watch" &&
					r.Path == watch3.Path &&
					len(r.CRNFull) == len(watch3.CRNFull) &&
					r.CRNFull[0] == strings.ToLower(watch3.CRNFull[0]) &&
					r.Wildcards == watch3.Wildcards &&
					len(r.RecordIDToWatch) == len(watch3.RecordIDToWatch) &&
					r.RecordIDToWatch[0] == watch3.RecordIDToWatch[0] &&
					r.SubscriptionEmail == watch3.SubscriptionEmail {

					match++
				}
			}

			if (match <= 1){
				log.Println("There are rows unmatched, failed")
				log.Println("Test 7: Get Watches FAILED")
				failed++
			} else {
				log.Println("Test 7: Get Watches PASSED")
			}
		}
	} else {
		log.Printf("total_count is %d, expecting 2, failed", total_count7)
		log.Println("Test 7: Get Watches FAILED")
		failed++
	}

	log.Println("======== "+FCT+" Test 8 ========")

	log.Println("Test delete expired subscription")
	err8,_ := db.DeleteExpiredSubscriptions(database)
	if err8 != nil {
		log.Println("Delete expired subscription failed: ", err8.Error())
		failed++
	}
	subscriptionReturn8, err8c,_ := db.GetSubscriptionByRecordIDStatement(database, subscription_id)
	if subscriptionReturn8 != nil {
		log.Println("Delete Maintenance failed, Maintenance still exists: ", err8c.Error())
		log.Println("Test 8: Delete expired subscription FAILED")
		failed++
	} else {
		log.Println("Test 8: Delete expired subscription PASSED")
	}


	// Summary
	log.Println("**** SUMMARY ****")
	if failed > 0 {
		log.Println("**** "+FCT+"FAILED ****")
	} else {
		log.Println("**** "+FCT+"PASSED ****")
	}
	return failed
}

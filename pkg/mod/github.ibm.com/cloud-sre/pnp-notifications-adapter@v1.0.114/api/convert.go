package api

var (
// EliminateBadOSSRecords will remove any notifications for resources that don't have good OSS status
//EliminateBadOSSRecords = true
)

// ConvertNotifications will convert cloudant structured notifications to something more generic
//func ConvertNotifications(ctx ctxt.Context, inputList *cloudant.NotificationsResult, creds *SourceConfig) (outputList *NotificationList, err error) {
//
//	METHOD := "api/ConvertNotifications"
//
//	if inputList == nil {
//		return nil, fmt.Errorf("ERROR (%s): Empty input list recieved. No cloudant result", METHOD)
//	}
//
//	eliminatedCount := 0
//	examinedCount := 0
//	outputList = new(NotificationList)
//	outputList.Items = make([]*Notification, 0, inputList.TotalRows)
//
//	for _, r := range inputList.Rows {
//
//		// Do not include any cloudant records without a notification ID.
//		// All entries will have an ID except some cloudant records not related to notifications.
//		if r.Doc.SubCategory != "" {
//
//			//createTime, err := time.Parse("2006-01-02T15:04:05.000Z", r.Doc.Creation.Time)
//			//if err != nil {
//			//	log.Println("Received time is not parsable. Could be bad record. Time=", r.Doc.Creation.Time)
//			//}
//
//			// Knock out alerts that are pre-2015 because there were some bad ones in 2014 with unrecognized IDs
//			//if err == nil && createTime.After(time.Date(2014, 12, 31, 0, 0, 0, 0, time.UTC)) {
//
//			// We keep all security notes, but others we only keep for 90 days
//			maxAge := time.Hour * 24 * 91
//			updateTime, err := time.Parse("2006-01-02T15:04:05.000Z", r.Doc.LastUpdate.Time)
//			expiredTime := time.Now()
//			if err != nil {
//				log.Println("Received last update time is not parsable. Could be bad record. Time=", r.Doc.LastUpdate.Time)
//			} else {
//				expiredTime = updateTime.Add(maxAge)
//			}
//
//			// We only want to deal with security and announcements. Things like incident are old coming from cloudant
//			if r.Doc.Type == "SECURITY" || (time.Now().Before(expiredTime) && r.Doc.Type == "ANNOUNCEMENT") {
//
//				// We only want to deal with these specific categories for now
//				if r.Doc.Category == "PLATFORM" || r.Doc.Category == "RUNTIMES" || r.Doc.Category == "SERVICES" {
//
//					examinedCount++
//
//					n := new(Notification)
//
//					n.ShortDescription = r.Doc.Title
//					n.LongDescription = r.Doc.Text
//
//					n.UpdateTime = r.Doc.LastUpdate.Time
//					n.CreationTime = r.Doc.Creation.Time
//					n.EventTimeStart = r.Doc.EventTime.Start
//					n.EventTimeEnd = r.Doc.EventTime.End
//					n.Category = strings.ToLower(r.Doc.Category)
//					n.CategoryNotificationID = r.Doc.SubCategory
//					n.Source = "cloudant"
//					n.SourceID = r.ID
//					n.NotificationType = strings.ToLower(r.Doc.Type)
//
//					crnData, err := GetCRNInfoForCloudantNote(ctx, r, creds)
//					if err != nil {
//						log.Printf("WARNING (%s): Could not locate CRN info for categoryID [%s]. creationtime [%s]. title [%s]. Error=[%s]", METHOD, n.CategoryNotificationID, n.CreationTime, n.ShortDescription, err.Error())
//						eliminatedCount++
//					} else {
//						if crnData == nil {
//							log.Println("Should Not Occur: Bad CRN info Pointer!!!")
//						} else {
//							if !EliminateBadOSSRecords || osscatalog.CategoryIDToOSSCompliance(ctx, n.CategoryNotificationID) == "ok" {
//								n.DisplayName = append(n.DisplayName, crnData.DisplayNames...)
//								n.CRNs = append(n.CRNs, crnData.CRNMasks...)
//
//								outputList.Items = append(outputList.Items, n)
//							} else {
//								eliminatedCount++
//							}
//						}
//					}
//				}
//			}
//			//}
//		}
//	}
//
//	// Count up the records here and output log message with information
//	secCount := 0
//	annCount := 0
//	for _, vvv := range outputList.Items {
//		if vvv.NotificationType == "announcement" {
//			annCount++
//		}
//		if vvv.NotificationType == "security" {
//			secCount++
//		}
//	}
//	log.Printf("INFO (%s): Input %d records. Examined %d records. Generated %d announcements and %d security notifications. Eliminated %d based on OSS status.", METHOD, len(inputList.Rows), examinedCount, annCount, secCount, eliminatedCount)
//
//	return outputList, err
//}

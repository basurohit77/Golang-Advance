package datastore

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

// WatchInsert type used to insert records into the db
type WatchInsert struct {
	SubscriptionRecordID string   `json:"subscription_record_id,omitempty" col:"subscription_id"`
	Kind                 string   `json:"kind,omitempty" col:"kind"`
	Path                 string   `json:"path,omitempty" col:"path"`
	CRNFull              []string `json:"crnMasks,omitempty" col:"crn_full"`
	Wildcards            string   `json:"wildcards,omitempty" col:"wildcards"`
	RecordIDToWatch      []string `json:"recordIDToWatch,omitempty" col:"record_id_to_watch"`
	SubscriptionEmail    string   `json:"subscriptionEmail,omitempty" col:"subscription_email"`
}

// WatchGet type used to get records from the db
type WatchGet struct {
	RecordID             string         `json:"record_id,omitempty"`
	SubscriptionRecordID string         `json:"subscription_record_id,omitempty"`
	Kind                 sql.NullString `json:"kind,omitempty"`
	Path                 sql.NullString `json:"path,omitempty"`
	CRNFull              sql.NullString `json:"crnMasks,omitempty"`
	Wildcards            sql.NullBool   `json:"wildcards,omitempty"`
	RecordIDToWatch      sql.NullString `json:"recordIDToWatch,omitempty"`
	SubscriptionEmail    sql.NullString `json:"subscriptionEmail,omitempty"`
}

// This is for the case when offset is beyond the result set, and ends up with all Null in all columns in the returned row, except total_count
type WatchGetNull struct {
	RecordID             sql.NullString `json:"record_id,omitempty"`
	SubscriptionRecordID sql.NullString `json:"subscription_record_id,omitempty"`
	Kind                 sql.NullString `json:"kind,omitempty"`
	Path                 sql.NullString `json:"path,omitempty"`
	CRNFull              sql.NullString `json:"crnMasks,omitempty"`
	Wildcards            sql.NullBool   `json:"wildcards,omitempty"`
	RecordIDToWatch      sql.NullString `json:"recordIDToWatch,omitempty"`
	SubscriptionEmail    sql.NullString `json:"subscriptionEmail,omitempty"`
}

// WatchReturn type used to pass records back to the API request
type WatchReturn struct {
	Href            string   `json:"href,omitempty"`
	RecordID        string   `json:"record_id,omitempty"`
	SubscriptionURL Href     `json:"subscription,omitempty"`
	Kind            string   `json:"kind,omitempty"`
	Path            string   `json:"path,omitempty"`
	CRNFull         []string `json:"crnMasks,omitempty"`
	Wildcards       string   `json:"wildcards,omitempty"`
	RecordIDToWatch []string `json:"recordIDToWatch,omitempty"`
	SubscriptionEmail string `json:"subscriptionEmail,omitempty"`
}

// WatchResponse type used to multiple records back to an API request
type WatchResponse struct {
	WatchURL  string        `json:"href,omitempty"`
	Resources []WatchReturn `json:"resources"`
}

// ConvertWatchReturnToWatchIns converts from a return type to an insert type
func ConvertWatchReturnToWatchIns(watchReturn WatchReturn) WatchInsert {
	//FCT := "ConvertWatchReturnToWatchIns: "
	tmpWatchInsert := WatchInsert{}
	tmpWatchInsert.SubscriptionRecordID = strings.Split(strings.Split(watchReturn.SubscriptionURL.URL, "/subscriptions/")[0], "/")[0]
	tmpWatchInsert.Kind = watchReturn.Kind
	tmpWatchInsert.Path = watchReturn.Path
	tmpWatchInsert.CRNFull = watchReturn.CRNFull
	tmpWatchInsert.RecordIDToWatch = watchReturn.RecordIDToWatch
	tmpWatchInsert.SubscriptionEmail = watchReturn.SubscriptionEmail

	tmpWatchInsert.Wildcards = watchReturn.Wildcards
	return tmpWatchInsert

}

// ConvertWatchGetToWatchIns from the Get type to an Insert type
func ConvertWatchGetToWatchIns(watchGet WatchGet) WatchInsert {
	FCT := "ConvertWatchGetToWatchIns: "
	tmpWatchInsert := WatchInsert{}
	tmpWatchInsert.SubscriptionRecordID = watchGet.SubscriptionRecordID
	if watchGet.Kind.Valid {
		tmpWatchInsert.Kind = watchGet.Kind.String
	}
	if watchGet.Path.Valid {
		tmpWatchInsert.Path = watchGet.Path.String
	}
	if watchGet.CRNFull.Valid {
		var tmpArray []string
		err := json.Unmarshal([]byte(watchGet.CRNFull.String), &tmpArray)
		if err != nil {
			log.Println(FCT+" Error on unmarshal :", err.Error())
			tmpWatchInsert.CRNFull = []string{}
		}
		tmpWatchInsert.CRNFull = tmpArray
	} else {
		tmpWatchInsert.CRNFull = []string{}
	}
	if watchGet.RecordIDToWatch.Valid {

		var tmpArr []string
		err := json.Unmarshal([]byte(watchGet.RecordIDToWatch.String), &tmpArr)
		if err != nil {
			log.Print(FCT+": Error : ", err.Error())
			return WatchInsert{}
		}
		tmpWatchInsert.RecordIDToWatch = tmpArr
	} else {
		tmpWatchInsert.RecordIDToWatch = []string{}
	}
	if watchGet.SubscriptionEmail.Valid {
		tmpWatchInsert.SubscriptionEmail = watchGet.SubscriptionEmail.String
	}
	if watchGet.Wildcards.Valid {
		tmpWatchInsert.Wildcards = strconv.FormatBool(watchGet.Wildcards.Bool)
	} else {
		tmpWatchInsert.Wildcards = "false"
	}
	return tmpWatchInsert

}

// ConvertWatchGetToWatchReturn converts from a Get type to a Return type
func ConvertWatchGetToWatchReturn(watchGet WatchGet, url string) WatchReturn {
	FCT := "ConvertWatchGetToWatchReturn: "
	url = strings.TrimRight(url, "/") + "/"
	tmpWatchReturn := WatchReturn{}
	tmpWatchReturn.Href = url + watchGet.SubscriptionRecordID + "/watches/" + watchGet.RecordID
	tmpWatchReturn.RecordID = watchGet.RecordID
	tmpWatchReturn.SubscriptionURL = Href{URL: url + watchGet.SubscriptionRecordID}
	if watchGet.Kind.Valid {
		tmpWatchReturn.Kind = strings.TrimRight(watchGet.Kind.String, "s") + "Watch"
	}
	if watchGet.Path.Valid {
		tmpWatchReturn.Path = watchGet.Path.String
	}
	if watchGet.CRNFull.Valid {
		watchGet.CRNFull.String = strings.Trim(watchGet.CRNFull.String, "{}")
		tmp := strings.Split(watchGet.CRNFull.String, ",")
		crnsString := ""
		if watchGet.CRNFull.String != "" {
			for i, v := range tmp {
				crnsString += "\"" + v + "\""
				if i < len(tmp)-1 {
					crnsString += ","
				}
			}
		}
		watchGet.CRNFull.String = "[" + crnsString + "]"
		err := json.Unmarshal([]byte(watchGet.CRNFull.String), &tmpWatchReturn.CRNFull)
		if err != nil {
			log.Print(FCT+": Error : ", err.Error())
			return WatchReturn{}
		}

	} else {
		tmpWatchReturn.CRNFull = []string{}
	}
	if watchGet.RecordIDToWatch.Valid {
		watchGet.RecordIDToWatch.String = strings.Trim(watchGet.RecordIDToWatch.String, "{}")
		tmp := strings.Split(watchGet.RecordIDToWatch.String, ",")

		recordString := ""
		for i, v := range tmp {
			if v != "" {
				recordString += "\"" + v + "\""
				if i < len(tmp)-1 {
					recordString += ","
				}
			}
		}

		watchGet.RecordIDToWatch.String = "[" + recordString + "]"
		err := json.Unmarshal([]byte(watchGet.RecordIDToWatch.String), &tmpWatchReturn.RecordIDToWatch)
		if err != nil {
			log.Print(FCT+": Error : ", err.Error())
			return WatchReturn{}
		}
	} else {

		tmpWatchReturn.RecordIDToWatch = []string{}
	}
	if watchGet.SubscriptionEmail.Valid {
		tmpWatchReturn.SubscriptionEmail = watchGet.SubscriptionEmail.String
	}
	if watchGet.Wildcards.Valid {
		tmpWatchReturn.Wildcards = strconv.FormatBool(watchGet.Wildcards.Bool)
	} else {
		tmpWatchReturn.Wildcards = "false"
	}
	return tmpWatchReturn

}

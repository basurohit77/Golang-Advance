package datastore

import (
	"database/sql"
)

// Href is used to handle URLs
type Href struct {
	URL string `json:"href,omitempty"`
}

// SubscriptionInsert is used to insert into the DB
type SubscriptionInsert struct {
	Name          string `json:"name,omitempty" col:"name"`
	TargetAddress string `json:"targetAddress,omitempty" col:"target_address"`
	TargetToken   string `json:"targetToken,omitempty" col:"target_token"`
	Expiration    string `json:"expiration,omitempty" col:"expiration"`
}

// SubscriptionReturn is used to return API requests
type SubscriptionReturn struct {
	RecordID      string `json:"recordId,omitempty"`
	Name          string `json:"name,omitempty"`
	TargetAddress string `json:"targetAddress,omitempty"`
	TargetToken   string `json:"targetToken,omitempty"`
	Expiration    string `json:"expiration,omitempty"`
}

// SubscriptionGet used to get data from the DB
type SubscriptionGet struct {
	RecordID      string         `json:"recordId,omitempty"`
	Name          string         `json:"name,omitempty"`
	TargetAddress string         `json:"targetAddress,omitempty"`
	TargetToken   sql.NullString `json:"targetToken,omitempty"`
	Expiration    sql.NullString `json:"expiration,omitempty"`
}

// SubscriptionResponse used to return multiple records back to an API request
type SubscriptionResponse struct {
	SubscriptionURL string                       `json:"href,omitempty"`
	Resources       []SubscriptionSingleResponse `json:"resources,omitempty"`
}

// SubscriptionSingleResponse used to return a record back to a post API request
type SubscriptionSingleResponse struct {
	SubscriptionURL   string `json:"href,omitempty"`
	RecordID          string `json:"recordId,omitempty"`
	Name              string `json:"name,omitempty"`
	TargetAddress     string `json:"targetAddress,omitempty"`
	TargetToken       string `json:"targetToken,omitempty"`
	Expiration        string `json:"expiration,omitempty"`
	Watches           Href   `json:"watches,omitempty"`
	IncidentWatch     Href   `json:"incidentWatch,omitempty"`
	MaintenanceWatch  Href   `json:"maintenanceWatch,omitempty"`
	ResourceWatch     Href   `json:"resourceWatch,omitempty"`
	CaseWatch         Href   `json:"caseWatch,omitempty"`
	NotificationWatch Href   `json:"notificationWatch,omitempty"`
}

// ConvertSubSingleToSubReturn  converts from a get type to a return type
func ConvertSubSingleToSubReturn(subSingle *SubscriptionSingleResponse) *SubscriptionReturn {
	tmpSubscriptionReturn := SubscriptionReturn{}
	tmpSubscriptionReturn.RecordID = subSingle.RecordID
	tmpSubscriptionReturn.Name = subSingle.Name
	tmpSubscriptionReturn.TargetAddress = subSingle.TargetAddress
	tmpSubscriptionReturn.Expiration = subSingle.Expiration
	tmpSubscriptionReturn.TargetToken = subSingle.TargetToken

	return &tmpSubscriptionReturn

}

// ConvertSubGetToSubReturn  converts from a get type to a return type
func ConvertSubGetToSubReturn(subGet *SubscriptionGet) *SubscriptionReturn {
	tmpSubscriptionReturn := SubscriptionReturn{}
	tmpSubscriptionReturn.RecordID = subGet.RecordID
	tmpSubscriptionReturn.Name = subGet.Name
	tmpSubscriptionReturn.TargetAddress = subGet.TargetAddress
	if subGet.Expiration.Valid {
		tmpSubscriptionReturn.Expiration = subGet.Expiration.String
	}
	// Remove the target token from the output https://github.ibm.com/cloud-sre/pnp-subscription/issues/124
	// Restore need to investigate more PuP reported is not getting an Authorization header
	if subGet.TargetToken.Valid {
		tmpSubscriptionReturn.TargetToken = subGet.TargetToken.String
	}
	return &tmpSubscriptionReturn

}

// ConvertSubGetToSubIns from a get type to an  insert type
func ConvertSubGetToSubIns(subGet SubscriptionGet) SubscriptionInsert {
	tmpSubscriptionInsert := SubscriptionInsert{}
	tmpSubscriptionInsert.Name = subGet.Name
	tmpSubscriptionInsert.TargetAddress = subGet.TargetAddress
	if subGet.Expiration.Valid {
		tmpSubscriptionInsert.Expiration = subGet.Expiration.String
	}
	if subGet.TargetToken.Valid {
		tmpSubscriptionInsert.TargetToken = subGet.TargetToken.String
	}
	return tmpSubscriptionInsert

}

// ConvertSubreturnsToSubresponse converts from a return type to a response type
func ConvertSubreturnsToSubresponse(subRets []SubscriptionReturn, url string) SubscriptionResponse {
	retVal := SubscriptionResponse{}
	if len(subRets) > 0 {
		retVal.SubscriptionURL = url
		for i := range subRets {
			subResponse := ConvertSubreturnToSubresponse(subRets[i], url+"/"+subRets[i].RecordID)
			retVal.Resources = append(retVal.Resources, subResponse)
		}
	}
	return retVal
}

// ConvertSubinsertToSubresponse - used for posting a subscription and getting the subscription info back
func ConvertSubinsertToSubresponse(subIns SubscriptionInsert, url string, recordID string) SubscriptionSingleResponse {
	retVal := SubscriptionSingleResponse{}
	retVal.SubscriptionURL = url
	retVal.RecordID = recordID
	retVal.TargetAddress = subIns.TargetAddress
	retVal.TargetToken = subIns.TargetToken
	retVal.Name = subIns.Name
	retVal.Expiration = subIns.Expiration
	retVal.Watches = Href{URL: url + "/watches"}
	retVal.IncidentWatch = Href{URL: url + "/watches/incidents"}
	retVal.MaintenanceWatch = Href{URL: url + "/watches/maintenance"}
	retVal.ResourceWatch = Href{URL: url + "/watches/resources"}
	retVal.CaseWatch = Href{URL: url + "/watches/case"}
	retVal.NotificationWatch = Href{URL: url + "/watches/notifications"}

	return retVal

}

// ConvertSubreturnToSubresponse - used for posting a subscription and getting the subscription info back
func ConvertSubreturnToSubresponse(subRet SubscriptionReturn, url string) SubscriptionSingleResponse {
	retVal := SubscriptionSingleResponse{}
	retVal.SubscriptionURL = url
	retVal.RecordID = subRet.RecordID
	retVal.TargetAddress = subRet.TargetAddress
	// Remove the target token from the output https://github.ibm.com/cloud-sre/pnp-subscription/issues/124
	//retVal.TargetToken = subRet.TargetToken
	retVal.Name = subRet.Name
	retVal.Expiration = subRet.Expiration
	retVal.Watches = Href{URL: url + "/watches"}
	retVal.IncidentWatch = Href{URL: url + "/watches/incidents"}
	retVal.MaintenanceWatch = Href{URL: url + "/watches/maintenance"}
	retVal.ResourceWatch = Href{URL: url + "/watches/resources"}
	retVal.CaseWatch = Href{URL: url + "/watches/case"}
	retVal.NotificationWatch = Href{URL: url + "/watches/notifications"}

	return retVal

}

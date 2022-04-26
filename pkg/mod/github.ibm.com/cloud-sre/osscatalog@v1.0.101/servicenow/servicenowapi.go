package servicenow

import (
	"fmt"
	"net/http"
	"strconv"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

const (
	apiV1 = 1 << iota
	apiV2
)

// ConfigurationItem is the data expected back from ServiceNow after creating or updating a Configuration Item (POST or PUT)
// or doing a GET for a Configuration Item
type ConfigurationItem struct {
	SysID          string `json:"sys_id"`
	CRNServiceName string `json:"crn_service_name"`
	DisplayName    string `json:"display_name"`

	GeneralInfo struct {
		RMCNumber         string                      `json:"rmc_number"`
		SOPNumber         string                      `json:"sop_number"` /* V2 only */
		OperationalStatus ossrecord.OperationalStatus `json:"operational_status"`
		ClientFacing      bool                        `json:"client_facing"`
		EntryType         ossrecord.EntryType         `json:"entry_type"`
		FullCRN           string                      `json:"full_crn"`
		OSSDescription    string                      `json:"oss_description"`
		ServiceNowCIURL   string                      `json:"service_now_ciurl"`
	} `json:"general_info"`

	Ownership struct {
		OfferingManager ossrecord.Person `json:"offering_manager"`
		SegmentName     string           `json:"segment_name"`
		ServiceOffering string           `json:"service_offering"`
		TribeName       string           `json:"tribe_name"`
		// TribeOwnerContainer PersonContainer  `json:"tribe_owner"` /* different V1 vs. V2 */
		TribeOwner ossrecord.Person `json:"tribe_owner"`
		TribeID    string           `json:"tribe_id"` /* V2 only */
	} `json:"ownership"`

	Support struct {
		Manager             ossrecord.Person              `json:"manager"`
		ClientExperience    ossrecord.ClientExperience    `json:"client_experience"`
		SpecialInstructions string                        `json:"special_instructions"`
		Tier2EscalationType ossrecord.Tier2EscalationType `json:"tier2_escalation_type"`
		Slack               ossrecord.SlackChannel        `json:"slack"`
		Tier2Repo           ossrecord.GHRepo              `json:"tier2_repo"`
		Tier2RTC            ossrecord.RTCCategory         `json:"tier2rtc"`
	} `json:"support"`

	Operations struct {
		Manager             ossrecord.Person              `json:"manager"`
		SpecialInstructions string                        `json:"special_instructions"`
		Tier2EscalationType ossrecord.Tier2EscalationType `json:"tier2_escalation_type"`
		Slack               ossrecord.SlackChannel        `json:"slack"`
		Tier2Repo           ossrecord.GHRepo              `json:"tier2_repo"`
		Tier2RTC            ossrecord.RTCCategory         `json:"tier2rtc"`
	} `json:"operations"`

	StatusPage struct {
		Group                string `json:"group"`
		CategoryID           string `json:"category_id"`
		CategoryIDMisspelled string `json:"cateogry_id"` // XXX V1 only - temporary
	} `json:"status_page"`

	Compliance struct {
		SNSupportVerified    bool `json:"sn_support_verified"`
		SNOperationsVerified bool `json:"sn_operations_verified"`
	} `json:"compliance"`

	ServiceNowInfo struct { // "ghosting" info from SN
		SupportTier1AG          string `json:"support_tier1ag"`
		SupportTier2AG          string `json:"support_tier2ag"`
		OperationsTier1AG       string `json:"operations_tier1ag"`
		OperationsTier2AG       string `json:"operations_tier2ag"`
		RCAApprovalGroup        string `json:"rca_approval_group"`
		ERCAApprovalGroup       string `json:"erca_approval_group"`       /* V2 only */
		TargetedCommunication   string `json:"targeted_communication"`    /* V2 only */
		CIEPageout              string `json:"cie_pageout"`               /* V2 only */
		BackupContacts          string `json:"backup_contacts"`           /* V2 only */
		GBT30                   string `json:"gbt30"`                     /* V2 only */
		District                string `json:"district"`                  /* V2 only */
		SupportNotApplicable    string `json:"support_not_applicable"`    /* V2 only */
		OperationsNotApplicable string `json:"operations_not_applicable"` /* V2 only */
	} `json:"service_now_info"`
}

// PersonContainer is a container for information about one person, returned in a JSON record
// either full ossrecord.Person for the V1 API
// or a simple string (w3id) for the V2 API
//type PersonContainer interface{}

// normalize the contents of a ConfigurationItem to account for differences
// between API V1 and V2 (esp. Person records)
func (ci *ConfigurationItem) normalize() {
	/*
		// Fixup Ownership.TribeOwner (API V2 returns a simple string not a Person record)
		src := ci.Ownership.TribeOwnerContainer
		dst := &ci.Ownership.TribeOwner
		var err error
		switch data := src.(type) {
		case string:
			dst.Name = data
		case map[string]interface{}:
			for k, v := range data {
				switch k {
				case "name":
					if v1, ok := v.(string); ok {
						if dst.Name == "" {
							dst.Name = v1
						} else {
							err = fmt.Errorf(`multiple "name" entries`)
							break
						}
					} else {
						err = fmt.Errorf(`"name" entry is not a string`)
						break
					}
				case "w3id":
					if v1, ok := v.(string); ok {
						if dst.W3ID == "" {
							dst.W3ID = v1
						} else {
							err = fmt.Errorf(`multiple "w3id" entries`)
							break
						}
					} else {
						err = fmt.Errorf(`"w3id" entry is not a string`)
						break
					}
				default:
					err = fmt.Errorf(`unexpected entry "%s"`, k)
					break
				}
			}
		default:
			err = fmt.Errorf("unexpected type")
		}
		if err != nil {
			panic(fmt.Sprintf("Unexpected value received from ServiceNow API for Ownership.TribeOwner: %T  %+v   (entry name=%s  id=%s): %v", src, src, ci.CRNServiceName, ci.SysID, err))
		}
	*/

	// Fixup StatusPage.CategoryID (misspelled in API V1)
	if ci.StatusPage.CategoryID == "" && ci.StatusPage.CategoryIDMisspelled != "" {
		ci.StatusPage.CategoryID = ci.StatusPage.CategoryIDMisspelled
	}
}

const apiV1BatchSize = 600

func getServiceNowCallInfo(sysid string, offset, size int, testMode bool, domain ServiceNowDomain) (actualURL string, headers http.Header) {
	switch apiVersion {
	case apiV1:
		if sysid != "" {
			if testMode {
				return fmt.Sprintf("%s/%s", serviceNowTestURLv1, sysid), nil
			}
			return fmt.Sprintf("%s/%s", serviceNowURLv1, sysid), nil
		}
		if testMode {
			return fmt.Sprintf("%s?sysparm_limit=%d", serviceNowTestURLv1, apiV1BatchSize), nil
		}
		return fmt.Sprintf("%s?sysparm_limit=%d", serviceNowURLv1, apiV1BatchSize), nil
	case apiV2:
		headers = make(http.Header)
		headers.Set("updated_by", serviceNowUser)
		headers.Set("use_gc_format", "true")
		headers.Set("include_all", "true")
		var serviceNowURL string
		if testMode && domain == WATSON {
			serviceNowURL = serviceNowWatsonTestURLv2
		} else if !testMode && domain == WATSON {
			serviceNowURL = serviceNowWatsonURLv2
		} else if testMode && domain == CLOUDFED {
			serviceNowURL = serviceNowCloudfedTestURLv2
		} else if !testMode && domain == CLOUDFED {
			serviceNowURL = serviceNowCloudfedURLv2
		} else {
			panic(fmt.Sprintf("Unknown ServiceNow domain: %T (%v)", domain, domain))
		}
		if sysid != "" {
			return fmt.Sprintf("%s/%s", serviceNowURL, sysid), headers
		}
		if offset >= 0 {
			headers.Set("offset", strconv.Itoa(offset))
		}
		if size >= 0 {
			headers.Set("limit", strconv.Itoa(size))
		}
		return fmt.Sprintf("%s", serviceNowURL), headers
	default:
		panic(fmt.Sprintf("Unknown ServiceNow API version: %d", apiVersion))
	}
}

func getServiceNowToken(testMode bool, domain ServiceNowDomain) (token string, err error) {
	tokenName := getTokenName(testMode, domain)
	token, err = rest.GetToken(tokenName)
	if err != nil {
		return "", err
	}
	return token, nil
}

func getTokenName(testMode bool, domain ServiceNowDomain) (tokenName string) {
	if testMode && domain == WATSON {
		tokenName = serviceNowWatsonTestTokenName
	} else if !testMode && domain == WATSON {
		tokenName = serviceNowWatsonTokenName
	} else if testMode && domain == CLOUDFED {
		tokenName = serviceNowCloudfedTestTokenName
	} else if !testMode && domain == CLOUDFED {
		tokenName = serviceNowCloudfedTokenName
	} else {
		panic(fmt.Sprintf("Unknown ServiceNow domain: %T (%v)", domain, domain))
	}
	return tokenName
}

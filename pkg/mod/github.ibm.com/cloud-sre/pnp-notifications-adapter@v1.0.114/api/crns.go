package api

import (
	"fmt"
	"github.ibm.com/cloud-sre/pnp-abstraction/ctxt"
	"github.ibm.com/cloud-sre/pnp-abstraction/osscatalog"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/cloudant"
	"github.ibm.com/cloud-sre/pnp-notifications-adapter/globalcatalog"
	"log"
	"regexp"
	"strings"
)

// CRNData provides CRN information returned by key methods
type CRNData struct {
	CRNMasks     []string
	DisplayNames []*TranslatedString
}

// GetCRNInfoForCloudantNote will retrieve CRN masks for the given cloudant notification record
//func GetCRNInfoForCloudantNote(ctx ctxt.Context, n *cloudant.NotificationsRecord, creds *SourceConfig) (result *CRNData, err error) {
//
//	METHOD := "GetCRNMaskForCloudantNote"
//
//	if n == nil {
//		return nil, fmt.Errorf("ERROR (%s): Record is nil", METHOD)
//	}
//
//	categoryID := n.Doc.SubCategory
//	if categoryID == "" {
//		return nil, fmt.Errorf("ERROR (%s): Category is empty. [%s]", METHOD, n.Doc.Title)
//	}
//
//	serviceName, err := osscatalog.CategoryIDToServiceName(ctx, categoryID)
//	if err != nil {
//		log.Printf("ERROR (%s) query to osscatalog to get service name for categoryID (%s) failed with error: %s", METHOD, categoryID, err.Error())
//	}
//
//	var cloudantNameMap *cloudant.NameMap
//	if serviceName == "" {
//		// Log an error indicating OSS does not have it
//
//		// Find servicename in manually maintained database (maintained by original cloud status page team)
//		serviceName, cloudantNameMap = getServiceNameFromCloudant(categoryID, creds)
//
//		if serviceName != "" {
//
//			errmsg := fmt.Sprintf("INFO (%s): Could not find service name in OSS records for categoryID [%s]. But successfully found in manual db [%s]", METHOD, categoryID, serviceName)
//			if !utils.LogSquelch(errmsg, time.Hour) {
//				log.Printf(errmsg)
//			}
//		} else {
//			// Sometimes the manual db has a record without a srvice name, just a display name
//			// Example here is 'Workflow'.  These are really old records.  So I think in these cases if I do
//			// have a display name, then I should just makeup the serviceName from the display name since
//			// these represent components that don't exist anymore
//			displayName, _ := getDisplayNameFromCloudant(categoryID, creds, cloudantNameMap)
//			if displayName != "" {
//				serviceName = NormalizeServiceName(displayName)
//
//				errmsg := fmt.Sprintf("INFO (%s): Could not find service name in OSS records or manual db for categoryID [%s]. But manual db has a display name [%s] -> [%s]", METHOD, categoryID, displayName, serviceName)
//				if !utils.LogSquelch(errmsg, time.Hour) {
//					log.Printf(errmsg)
//				}
//			}
//		}
//	}
//
//	if serviceName != "" {
//
//		result = new(CRNData)
//		result.DisplayNames = make([]*TranslatedString, 0, 1)
//
//		result.AddDisplayNames(ctx, serviceName, categoryID, creds, cloudantNameMap)
//		result.addCRNs(serviceName, n)
//
//	} else {
//		// Log error that service name cannot be located
//		// Will need to return nil so this record will be unused.
//		err = fmt.Errorf("ERROR (%s): could not locate service name for notification category ID: %s", METHOD, categoryID)
//	}
//
//	return result, err
//}

// NormalizeServiceName will take a name and force it to follow naming rules for services
func NormalizeServiceName(name string) string {
	n := strings.ToLower(name)

	re := regexp.MustCompile("[^0-9a-z]")
	n = re.ReplaceAllString(n, "-")

	return n
}

func (result *CRNData) addCRNs(serviceName string, note *cloudant.NotificationsRecord) {

	// This will need to be improved in the future.
	baseCRNMask := "crn:v1:bluemix:public:%s:%s::::"

	for _, region := range note.Doc.RegionsAffected {
		crnMask := fmt.Sprintf(baseCRNMask, NormalizeServiceName(serviceName), NormalizeServiceName(region.ID))
		result.CRNMasks = append(result.CRNMasks, crnMask)
	}
}

// AddDisplayNames will add display names to the CRNData object
func (result *CRNData) AddDisplayNames(ctx ctxt.Context, serviceName, categoryID string, creds *SourceConfig, cloudantNameMap *cloudant.NameMap) {
	METHOD := "AddDisplayNames"

	// Find service name in global catalog
	gcResources, err := globalcatalog.GetCloudResourcesCache(ctx, creds.GlobalCatalog.ResourcesURL)
	if err != nil {
		log.Printf("ERROR (%s) getting global catalog resource records: %s", METHOD, err.Error())
	}

	if gcResources != nil && gcResources.Resources[serviceName] != nil { // If found in the global catalog, use display names from there

		cr := gcResources.Resources[serviceName]

		// First get the english string since this will be the default
		defaultDn := serviceName
		for _, ls := range cr.LanguageStrings {
			if ls.Language == "en" {
				defaultDn = ls.DisplayName
				break
			}
		}

		// Now fill in other languages and use english default where display name absent
		for _, ls := range cr.LanguageStrings {
			lang := ls.Language
			dn := ls.DisplayName
			if dn == "" {
				dn = defaultDn
			}

			result.DisplayNames = append(result.DisplayNames, &TranslatedString{Language: lang, Text: dn})
		}

	} else { // If not found in the global catalog, use display name from OSS record or manual database

		displayName, err := osscatalog.CategoryIDToDisplayName(ctx, categoryID)
		if err != nil {
			log.Printf("ERROR (%s) query to osscatalog to get display name for categoryID (%s) failed with error: %s", METHOD, categoryID, err.Error())
		}

		//if displayName == "" && cloudantNameMap != nil {
		//	displayName, _ = getDisplayNameFromCloudant(categoryID, creds, cloudantNameMap)
		//}

		// No display name found. Just default to the service name
		if displayName == "" {
			displayName = serviceName
		}

		result.DisplayNames = append(result.DisplayNames, &TranslatedString{Language: "en", Text: displayName})
	}
}

//func getServiceNameFromCloudant(categoryID string, creds *SourceConfig) (serviceName string, cloudantNameMap *cloudant.NameMap) {
//
//	METHOD := "getServiceNameFromCloudant"
//
//	cloudantNameMap, err := cloudant.NewNameMap(creds.Cloudant.AccountID, creds.Cloudant.Password, creds.Cloudant.ServiceNamesURL, creds.Cloudant.RuntimeNamesURL, creds.Cloudant.PlatformNamesURL)
//
//	if err != nil {
//		// Log this error and move on
//		log.Printf("ERROR (%s) getting cloudant name map: %s", METHOD, err.Error())
//	} else {
//		nameComponent := cloudantNameMap.MatchCategoryID(categoryID)
//		if nameComponent != nil {
//			serviceName = nameComponent.ServiceName
//		}
//	}
//
//	return serviceName, cloudantNameMap
//}

//func getDisplayNameFromCloudant(categoryID string, creds *SourceConfig, cNameMap *cloudant.NameMap) (displayName string, cloudantNameMap *cloudant.NameMap) {
//	METHOD := "getDisplayNameFromCloudant"
//	var err error
//
//	cloudantNameMap = cNameMap
//
//	if cloudantNameMap == nil {
//		cloudantNameMap, err = cloudant.NewNameMap(creds.Cloudant.AccountID, creds.Cloudant.Password, creds.Cloudant.ServiceNamesURL, creds.Cloudant.RuntimeNamesURL, creds.Cloudant.PlatformNamesURL)
//
//		if err != nil {
//			// Log this error and move on
//			log.Printf("ERROR (%s) getting cloudant name map: %s", METHOD, err.Error())
//		}
//	}
//
//	if cloudantNameMap != nil {
//		nameComponent := cloudantNameMap.MatchCategoryID(categoryID)
//		if nameComponent != nil {
//			displayName = nameComponent.DisplayName
//		}
//	}
//
//	return displayName, cloudantNameMap
//}

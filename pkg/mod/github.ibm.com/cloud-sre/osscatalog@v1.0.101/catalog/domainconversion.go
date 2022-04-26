package catalog

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// The code in this file provides support to obtain domain specific versions of OSS service records.

// createNewOSSServiceExtendedWithService returns a new OSS service extended instance with the provided OSS service set.
// The provided service extended and service records are not changed by this function.
func createNewOSSServiceExtendedWithService(serviceExt *ossrecordextended.OSSServiceExtended, service *ossrecord.OSSService) ossrecordextended.OSSServiceExtended {
	updatedServiceExtended := *serviceExt
	updatedServiceExtended.OSSService = *service
	return updatedServiceExtended
}

// getDomainSpecificOSSEntries returns all the OSS entries that match the domains in the provided include
// options. The commercial domain will be included by default if no specific domains are included in the
// include options. Currently only supported for OSS services or OSS service extended entries - other types
// will simply be returned in the response array. The provided entry is not changed by this function.
func getDomainSpecificOSSEntries(entry *ossrecord.OSSEntry, incl IncludeOptions) []ossrecord.OSSEntry {
	var result []ossrecord.OSSEntry
	_, service := getOSSServiceFromOSSEntry(entry)
	if service != nil {
		// Add the US regulated OSS service to result if requested:
		if (incl & IncludeServicesDomainUSRegulated) != 0 {
			includeOptions := IncludeServicesDomainUSRegulated
			if (incl & IncludeServicesDomainOverrides) != 0 {
				includeOptions |= IncludeServicesDomainOverrides
			}
			entryForDomain, err := getOSSEntryForDomain(entry, includeOptions)
			if err == nil {
				result = append(result, entryForDomain)
			}
		}

		// If commercial services is requested, or US regulated not requested, then include commercial:
		if ((incl & IncludeServicesDomainCommercial) != 0) || ((incl & IncludeServicesDomainUSRegulated) == 0) {
			includeOptions := IncludeServicesDomainCommercial
			if (incl & IncludeServicesDomainOverrides) != 0 {
				includeOptions |= IncludeServicesDomainOverrides
			}
			entryForDomain, err := getOSSEntryForDomain(entry, includeOptions)
			if err == nil {
				result = append(result, entryForDomain)
			}
		}
	} else {
		result = append(result, *entry)
	}
	return result
}

// getOSSEntryForDomain returns a new OSS entry for the domain defined in the include options. If
// no update is required, the OSS entry is passed back unchanged. Currently only supported for OSS services
// or OSS service extended entries - other types will simply be returned back unchanged. The provided OSS
// entry is not changed by this function.
func getOSSEntryForDomain(entry *ossrecord.OSSEntry, incl IncludeOptions) (ossrecord.OSSEntry, error) {
	serviceExtended, service := getOSSServiceFromOSSEntry(entry)
	if service != nil {
		var err error

		// Assume no changes needed:
		updatedService := *service

		// Check if we need to convert to US regulated:
		if (incl & IncludeServicesDomainUSRegulated) != 0 {
			updatedService, err = getOSSServiceRecordForDomain(*service, ossrecord.USREGULATED)
			if err != nil {
				return nil, err
			}
		}

		// Check whether the overrides should be kept:
		if updatedService.Overrides != nil && (incl&IncludeServicesDomainOverrides) == 0 {
			updatedService.Overrides = nil
		}

		if serviceExtended != nil {
			updatedServiceExtended := createNewOSSServiceExtendedWithService(serviceExtended, &updatedService)
			var updatedEntry ossrecord.OSSEntry
			updatedEntry = &updatedServiceExtended
			return updatedEntry, nil
		}
		return &updatedService, nil
	}
	return *entry, nil
}

// getOSSServiceFromOSSEntry returns the OSS service extended and OSS service for the provided OSS entry if any.
// The provided entry is not modified by this function.
func getOSSServiceFromOSSEntry(entry *ossrecord.OSSEntry) (*ossrecordextended.OSSServiceExtended, *ossrecord.OSSService) {
	switch entry0 := (*entry).(type) {
	case *ossrecordextended.OSSServiceExtended:
		return entry0, &entry0.OSSService
	case *ossrecord.OSSService:
		return nil, entry0
	default:
		return nil, nil
	}
}

// getOSSServiceRecordForDomain returns an new updated OSS service given the provided domain
func getOSSServiceRecordForDomain(service ossrecord.OSSService, targetDomain ossrecord.Domain) (ossrecord.OSSService, error) {
	if service.GeneralInfo.Domain == targetDomain {
		return service, nil
	}
	for _, override := range service.Overrides {
		if override.GeneralInfo.Domain == targetDomain {
			// Apply overrides:
			if override.GeneralInfo.OSSTags.Contains(osstags.ServiceNowApproved) {
				service.GeneralInfo.OSSTags.AddTag(osstags.ServiceNowApproved)
			}
			service.GeneralInfo.ServiceNowSysid = override.GeneralInfo.ServiceNowSysid
			service.GeneralInfo.ServiceNowCIURL = override.GeneralInfo.ServiceNowCIURL
			service.GeneralInfo.Domain = override.GeneralInfo.Domain
			service.Compliance.ServiceNowOnboarded = override.Compliance.ServiceNowOnboarded
			return service, nil
		}
	}
	return ossrecord.OSSService{}, fmt.Errorf("getOSSServiceRecordForDomain(%s, %s): returning with domain %s does not exist for service %s", service.ReferenceResourceName, targetDomain, targetDomain, service.ReferenceResourceName)
}

// GetOrCreateOverride returns the index of the override object for the provided service and domain. If
// a new override is created, it is added to the provided service.
func GetOrCreateOverride(service *ossrecord.OSSService, domain ossrecord.Domain) (int, error) {
	// Ensure domain is valid:
	if !isValidDomain(domain) {
		return -2, fmt.Errorf("GetOrCreateOverride(%s, %s): returned with invalid domain %s", service, domain, domain)
	}

	// Check if domain is commercial in which case we don't need an override:
	if domain == ossrecord.COMMERCIAL {
		return -1, nil
	}

	// See if the override already exists, if so return the index:
	for i, currentOverride := range service.Overrides {
		if currentOverride.GeneralInfo.Domain == domain {
			return i, nil
		}
	}
	// If override does not exist, create the override, add the override to the provided service, and return the index of the override:
	var override = ossrecord.OSSServiceOverride{}
	override.GeneralInfo.Domain = domain
	service.Overrides = append(service.Overrides, override)
	return 0, nil
}

// isValidDomain returns true if the provided domain is a valid known domain, false otherwise
func isValidDomain(domain ossrecord.Domain) bool {
	return domain == ossrecord.COMMERCIAL || domain == ossrecord.USREGULATED
}

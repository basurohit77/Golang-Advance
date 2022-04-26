package catalog

import (
	"fmt"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
)

var categoryTags map[string]string
var unknownCatalogTags = make(map[string]int)

var freePricingTags map[string]string
var unknownPricingTags = make(map[string]int)

func init() {
	// TODO: Validate how Global Catalog tags map to Categories in the Client-facing catalog
	categoryTags = map[string]string{
		"apidocs_enabled":                "",
		"ai":                             "",           // As of 190825 "AI"
		"analytics":                      "*Analytics", // As of 200814: CAT - some in CAT but missing GC tag?
		"api":                            "",           // As of 190825 "Integration"
		"apis":                           "",           // "APIs",
		"api docs enabled":               "",
		"apps-services":                  "",                             // "Application Services",
		"app_services":                   "Web and Application",          // As of 190825
		"architecture":                   "Compute>VMware Architectures", // As of 190825
		"beryllium":                      "",
		"big_data":                       "", // "Data & Analytics" // As of 200814: RMC
		"billing.composite":              "",
		"blockchain":                     "*Blockchain",            // As of 200814: CAT
		"business_analytics":             "*Analytics",             // As of 200814: RMC; CAT(merge into analytics)
		"cloud_foundry":                  "*Compute>Cloud Foundry", // As of 200814: CAT - does not account for `codeengine` or CF platform
		"clusters":                       "",                       // As of 190825 /* XXX also "Containers" */
		"community":                      "",
		"compliance:fedramp":             "",
		"compliance:fedramp:pending":     "",
		"compute":                        "**Compute>Product",           // As of 200814: RMC; CAT(only tag for Product - `websphereappsvr`)
		"compute_baremetal":              "*Compute>Bare Metal Servers", // As of 200814: CAT
		"containers":                     "*Containers",                 // As of 200814: RMC, CAT
		"content":                        "Compute>Cloud Images",        // As of 190825
		"couchdb":                        "",
		"data":                           "",                 // "Data & Analytics",
		"database":                       "*Databases",       // As of 200814: CAT(superseded by `data_management`? - except for `dashdb-for-transactions, `esql(Elephant)`` and `informix`)
		"data_analytics":                 "*Analytics",       // As of 200814: CAT(merge into analytics)
		"data_management":                "*Databases",       // As of 200814: RMC; CAT (has merges)
		"dedicated":                      "",                 // As of 200814: RMC
		"dev_ops":                        "*Developer Tools", // As of 200814: RMC; CAT = maybe some dups in `devops`?
		"devops":                         "*Developer Tools", // As of 200814: CAT (superseded by `dev_ops`?)
		"disable_reclamation":            "",
		"eu_access":                      "",
		"eu_location":                    "",
		"finance":                        "", // "Finance",
		"financial services":             "",
		"freesia":                        "",
		"fs cloud":                       "",
		"fs_ready":                       "",
		"fss":                            "",
		"g1":                             "",
		"gc_migrate":                     "",
		"group":                          "",
		"hipaa":                          "",
		"iaas":                           "", // "Infrastructure",
		"ibm cloud for vmware solutions": "",
		"ibm_beta":                       "",
		"ibm_community":                  "",
		"ibm_created":                    "",
		"ibm_dedicated_public":           "",
		"ibm_deprecated":                 "",
		"ibm_experimental":               "",
		"ibm_iaas_g0":                    "", // "Infrastructure",
		"ibm_iaas_g1":                    "", // "Infrastructure",
		"ibm_iaas_g2":                    "", // "Infrastructure",
		"ibm_internal":                   "",
		"ibm_release":                    "",
		"ibm_third_party":                "",
		"integration":                    "*Integration",        // As of 200814: RMC; CAT
		"internet_of_things":             "*Internet of Things", // As of 200814: RMC; CAT
		"is.composite":                   "",
		"lite":                           "",                                   // TODO: Should "lite" as non-pricing Catalog tag be interpreted as pricing anyway? // As of 200814: RMC
		"local":                          "",                                   // As of 200814: RMC
		"logging_monitoring":             "*Logging and Monitoring",            // As of 200814: CAT
		"mobile":                         "*Mobile",                            // As of 200814: RMC; CAT (minus "Twilio Programmable SMS" - has web_and_app but not specific enough)
		"network":                        "**Networking>Product",               // As of 200814: RMC; CAT(only tag for Product - `is-flow-log-collector`)
		"network_classic":                "*Networking>Classic Infrastructure", // As of 200814: CAT
		"network_edge":                   "*Networking>Edge",                   // Ad of 200814: CAT
		"network_interconnectivity":      "*Networking>Interconnectivity",      // Ad of 200814: CAT
		"network_vpc":                    "*Networking>VPC Infrastructure",     // Ad of 200814: CAT
		"nosql":                          "",
		"openwhisk":                      "",                  // As of 190825
		"payments":                       "",                  // As of 200814: RMC
		"platform_service":               "Platform Services", // As of 190825
		"private":                        "",                  // As of 200814: RMC
		"rc_compatible":                  "",
		"registry":                       "",                       // As of 190825 /* XXX also "Containers" */
		"runtime":                        "*Compute>Cloud Foundry", // As of 200814: RMC
		"satellite_enabled":              "",
		"security":                       "*Security",                   // As of 200814: RMC; CAT
		"serverless":                     "*Compute>Serverless Compute", // As of 200814: CAT
		"service":                        "",
		"service_endpoint_supported":     "",
		"servicegroup.billing.account":   "",
		"servicegroup.billing.invoicing": "",
		"softlayer":                      "",
		"starterkit":                     "",
		"storage":                        "**Storage>Product",               // As of 200814: RMC; CAT(Product)
		"storage_classic":                "*Storage>Classic Infrastructure", // As of 200814: CAT
		"storage_datamovement":           "*Storage>Data Movement",          // As of 200814: CAT
		"storage_vpc":                    "*Storage>VPC Infrastructure",     // As of 200814: CAT
		"supports_syndication":           "",
		"target_iks":                     "",
		"target_roks":                    "",
		"template":                       "Starter Kits",                        // As of 190825
		"virtualservers":                 "*Compute>Virtual Machines",           // As of 200814: CAT
		"virtual_data_center":            "Compute>VMware Virtual Data Centers", // As of 190825
		"vmware":                         "Compute>Containers and VMs",          // As of 200814: RMC
		"vmware_managed_service":         "Compute>VMware Managed Services",     // As of 190825
		"vmware_service":                 "Compute>VMware Services",             // As of 190825
		"vpc":                            "VPC Infrastructure",                  // As of 190825
		"watson":                         "*AI / Machine Learning",              // As of 200814: RMC; CAT
		"web_and_app":                    "Web and Mobile",                      // As of 200814: RMC
		"whisk":                          "",                                    // "Functions",
	}

	freePricingTags = map[string]string{
		"allow_internal_users": "",
		"free":                 "Free",
		"invoiced":             "",
		"lite":                 "Lite",
		"paid":                 "",
		"paid_only":            "",
		"paygo":                "",
		"standard":             "",
		"subscription":         "",
		"subscription_only":    "",
		"trial":                "Trial",
	}

}

// ScanCategoryTags figures out the Catalog category from all the tags and other info
// (note that not all tags represent a category, hence we must enumerate them)
func ScanCategoryTags(r *catalogapi.Resource) string {
	var categories []string

	for _, t := range r.Tags {
		if cat, ok := categoryTags[t]; ok {
			if cat != "" {
				categories = append(categories, cat)
			} else {
				// Not a category
			}
		} else {
			unknownCatalogTags[t] = unknownCatalogTags[t] + 1
		}
	}

	switch r.Kind {
	case "boilerplate":
	case "template":
		categories = append(categories, "Boilerplates")
	case "runtime":
		categories = append(categories, "Cloud Foundry Apps")
	}

	switch len(categories) {
	case 0:
		return ""
	case 1:
		return categories[0]
	default:
		// remove duplicates
		var allCats = make(map[string]bool)
		var sorted []string
		for _, c := range categories {
			if _, found := allCats[c]; !found {
				allCats[c] = true
				sorted = append(sorted, c)
			}
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i] < sorted[j]
		})
		return strings.Join(sorted, ", ")
	}
}

// ScanFreePricing checks if there are any free plans, from all the tags and other info
func ScanFreePricing(r *catalogapi.Resource) string {
	var pricing []string

	for _, t := range r.PricingTags {
		if cat, ok := freePricingTags[t]; ok {
			if cat != "" {
				pricing = append(pricing, cat)
			} else {
				// Not a free pricing category
			}
		} else {
			unknownPricingTags[t] = unknownPricingTags[t] + 1
		}
	}
	sort.Slice(pricing, func(i, j int) bool {
		return pricing[i] < pricing[j]
	})
	return strings.Join(pricing, ", ")
}

// ListUnknownTags generates a list of unknown tags, as an error message
func ListUnknownTags() {
	if len(unknownCatalogTags) > 0 {
		few, many := generateUnknownList(unknownCatalogTags)
		if len(few) > 0 {
			debug.Info("Unknown Catalog tags with few instances: %v", few)
		}
		if len(many) > 0 {
			debug.PrintError("Unknown Catalog tags with many instances: %v", many)
		}
	}
	if len(unknownPricingTags) > 0 {
		few, many := generateUnknownList(unknownPricingTags)
		if len(few) > 0 {
			debug.Info("Unknown Catalog Pricing tags with few instances: %v", few)
		}
		if len(many) > 0 {
			debug.PrintError("Unknown Catalog Pricing tags with many instances: %v", many)
		}
	}
}

func generateUnknownList(m map[string]int) (string, string) {
	var unknownFew = make([]string, 0, len(m))
	var unknownMany = make([]string, 0, len(m))
	for tag, count := range m {
		if count < 10 {
			unknownFew = append(unknownFew, fmt.Sprintf("%s(%d)", tag, count))
		} else {
			unknownMany = append(unknownMany, fmt.Sprintf("%s(%d)", tag, count))
		}
	}
	sort.Slice(unknownFew, func(i, j int) bool {
		return unknownFew[i] < unknownFew[j]
	})
	sort.Slice(unknownMany, func(i, j int) bool {
		return unknownMany[i] < unknownMany[j]
	})
	return strings.Join(unknownFew, ", "), strings.Join(unknownMany, ", ")
}

// SearchTags searches the Catalog tags to determine if they contain one particular tag
func SearchTags(r *catalogapi.Resource, tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

package catalogapi

// Data type definitions for the Global Catalog.
// In a separate package from the Catalog access functions, to avoid a import cycle

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// GetResponse represents the response from GET calls in Global Catalog (in JSON)
type GetResponse struct {
	Count         int64      `json:"count"`
	Limit         int64      `json:"limit"`
	Offset        int64      `json:"offset"`
	Next          string     `json:"next"`
	ResourceCount int64      `json:"resource_count"`
	Resources     []Resource `json:"resources"`
}

// Resource represents the key items that we are about in a Global Catalog entry (in JSON)
type Resource struct {
	ID         string `json:"id" mapstructure:"id"`
	Kind       string `json:"kind" mapstructure:"kind"`
	CatalogCRN string `json:"catalog_crn,omitempty" mapstructure:"catalog_crn,omitempty"`
	URL        string `json:"url,omitempty" mapstructure:"url,omitempty"`
	Name       string `json:"name" mapstructure:"name"`
	Images     struct {
		Image string `json:"image,omitempty" mapstructure:"image,omitempty"`
	} `json:"images,omitempty" mapstructure:"images,omitempty"`
	Active      bool     `json:"active" mapstructure:"active"`
	Disabled    bool     `json:"disabled" mapstructure:"disabled"`
	Tags        []string `json:"tags" mapstructure:"tags"`
	GeoTags     []string `json:"geo_tags" mapstructure:"geo_tags"`
	PricingTags []string `json:"pricing_tags" mapstructure:"pricing_tags"`
	Group       bool     `json:"group" mapstructure:"group"`
	ParentID    string   `json:"parent_id,omitempty" mapstructure:"parent_id,omitempty"`
	ParentURL   string   `json:"parent_url,omitempty" mapstructure:"parent_url,omitempty"`
	ChildrenURL string   `json:"children_url" mapstructure:"children_url"`
	OverviewUI  struct {
		En struct {
			DisplayName     string `json:"display_name" mapstructure:"display_name"`
			Description     string `json:"description" mapstructure:"description"`
			LongDescription string `json:"long_description" mapstructure:"long_description"`
		} `json:"en" mapstructure:"en"`
	} `json:"overview_ui" mapstructure:"overview_ui"`
	Created  string `json:"created,omitempty" mapstructure:"created,omitempty"`
	Updated  string `json:"updated,omitempty" mapstructure:"updated,omitempty"`
	Provider struct {
		Name         string `json:"name" mapstructure:"name"`
		Email        string `json:"email" mapstructure:"email"`
		Contact      string `json:"contact" mapstructure:"contact"`
		SupportEmail string `json:"support_email" mapstructure:"support_email"`
		Phone        string `json:"phone" mapstructure:"phone"`
	} `json:"provider" mapstructure:"provider"`
	ObjectMetaData struct {
		Other struct {
			OSS                       *ossrecord.OSSService                `json:"oss,omitempty" mapstructure:"oss,omitempty"`
			OSSMergeControl           *ossmergecontrol.OSSMergeControl     `json:"oss_merge_control,omitempty" mapstructure:"oss_merge_control,omitempty"`
			OSSValidation             *ossvalidation.OSSValidation         `json:"oss_validation_info,omitempty" mapstructure:"oss_validation_info,omitempty"`
			OSSSegment                *ossrecord.OSSSegment                `json:"oss_segment,omitempty" mapstructure:"oss_segment,omitempty"`
			OSSTribe                  *ossrecord.OSSTribe                  `json:"oss_tribe,omitempty" mapstructure:"oss_tribe,omitempty"`
			OSSEnvironment            *ossrecord.OSSEnvironment            `json:"oss_environment,omitempty" mapstructure:"oss_environment,omitempty"`
			OSSResourceClassification *ossrecord.OSSResourceClassification `json:"oss_resource_classification,omitempty" mapstructure:"oss_resource_classification,omitempty"`
			Composite                 *Composite                           `json:"composite,omitempty" mapstructure:"composite,omitempty"`
		} `json:"other" mapstructure:"other"`
		UI struct {
			Hidden bool `json:"hidden,omitempty" mapstructure:"hidden,omitempty"`
		} `json:"ui,omitempty" mapstructure:"ui,omitempty"`
		OriginalName string `json:"original_name" mapstructure:"original_name"`
		RCCompatible bool   `json:"rc_compatible" mapstructure:"rc_compatible"`
		Service      *struct {
			AsyncProvisioningSupported   bool `json:"async_provisioning_supported" mapstructure:"async_provisioning_supported"`
			AsyncUnprovisioningSupported bool `json:"async_unprovisioning_supported" mapstructure:"async_unprovisioning_supported"`
			Bindable                     bool `json:"bindable" mapstructure:"bindable"`
			IAMCompatible                bool `json:"iam_compatible" mapstructure:"iam_compatible"`
			RCProvisionable              bool `json:"rc_provisionable" mapstructure:"rc_provisionable"`
			ServiceCheckEnabled          bool `json:"service_check_enabled" mapstructure:"service_check_enabled"`
			TestCheckInterval            int  `json:"test_check_interval" mapstructure:"test_check_interval"`
			// other attributes omitted
		} `json:"service" mapstructure:"service"`
		Pricing *struct {
			//			Origin string `json:"origin"`
			//			Type   string `json:"type"`
			URL string `json:"url" mapstructure:"url"`
			//			StartingPrice map[string]interface{} `json:"starting_price"`
			//			Metrics       map[string]interface{} `json:"metrics"`
		} `json:"pricing" mapstructure:"pricing"`
		Deployment *struct {
			Location    string `json:"location" mapstructure:"location"`
			LocationURL string `json:"location_url" mapstructure:"location_url"`
			TargetCRN   string `json:"target_crn" mapstructure:"target_crn"`
			Broker      *struct {
				Name string `json:"name" mapstructure:"name"`
				GUID string `json:"guid" mapstructure:"guid"`
			} `json:"broker" mapstructure:"broker"`
			MCCPID string `json:"mccp_id" mapstructure:"mccp_id"`
		} `json:"deployment" mapstructure:"deployment"`
		Callbacks *struct {
			BrokerURL         string `json:"broker_url" mapstructure:"broker_url"`
			ServiceMonitorAPI string `json:"service_monitor_api" mapstructure:"service_monitor_api"`
			ServiceMonitorApp string `json:"service_monitor_app" mapstructure:"service_monitor_app"`
			// other attributes omitted
		} `json:"callbacks" mapstructure:"callbacks"`
		Plan *struct {
			ServiceCheckEnabled bool `json:"service_check_enabled" mapstructure:"service_check_enabled"`
			TestCheckInterval   int  `json:"test_check_interval" mapstructure:"test_check_interval"`
			// other attributes omitted
		} `json:"plan" mapstructure:"plan"`
	} `json:"metadata" mapstructure:"metadata"`
	Visibility          Visibility `json:"visibility" mapstructure:"visibility"`
	CatalogPath         string     `json:"-"` // INTERNAL info about the path in the Catalog that lead to this entry -- never stored in the Catalog itself
	EffectiveVisibility Visibility `json:"-"` // INTERNAL info about the effective Visibility of this entry, accounting for possible visibility restrictions on parents -- never stored in the Catalog itself
}

// Possible values for Catalog "kind" attribute (partial list)
const (
	KindAvailabilityZone  string = "zone"  // Availability Zone
	KindLegacyCName       string = "cname" // deprecated
	KindComposite         string = "composite"
	KindCountry           string = "country"
	KindDatacenter        string = "dc" // Datacenter
	KindDeployment        string = "deployment"
	KindLegacyEnvironment string = "environment" // deprecated
	KindFlavor            string = "flavor"
	KindGeography         string = "geography"
	KindIaaS              string = "iaas"
	KindMetro             string = "metro"
	KindOSSService        string = "oss"
	KindOSSSegment        string = "oss_segment"
	KindOSSTribe          string = "oss_tribe"
	KindOSSEnvironment    string = "oss_environment"
	KindOSSSchema         string = "oss_schema"
	KindPlan              string = "plan"
	KindPlatformService   string = "platform_service"
	KindPlugin            string = "plugin"
	KindPOP               string = "pop" // Point of Presence
	KindProfile           string = "profile"
	KindRegion            string = "region"
	KindRuntime           string = "runtime"
	KindService           string = "service"
	KindTemplate          string = "template"
	KindSatellite         string = "satellite"
	KindLocation          string = "location"
	KindISProfiles        string = "is.profiles"
)

// VisibilityRestrictions represents the value of the Visibility.Restrictions attribute in the Catalog REST API
type VisibilityRestrictions string

// Possible values for VisibilityRestrictions
const (
	VisibilityPublic  VisibilityRestrictions = "public"
	VisibilityIBMOnly VisibilityRestrictions = "ibm_only"
	VisibilityPrivate VisibilityRestrictions = "private"
)

// Visibility represents the entry visibility information from the Catalog REST API (in JSON)
type Visibility struct {
	Restrictions string   `json:"restrictions"`
	Owner        string   `json:"owner"`
	Owners       []string `json:"owners,omitempty"`
	Include      struct {
		Accounts map[string]string `json:"accounts"`
	} `json:"include,omitempty"`
	Exclude struct {
		Accounts map[string]string `json:"accounts"`
	} `json:"exclude,omitempty"`
	Approved bool `json:"approved,omitempty"`
}

// MergeVisibility updates a Visibility entry to account for possibly more restrictive visibilty attributes from a parent
func (v *Visibility) MergeVisibility(v0 *Visibility) {
	if v0 == nil {
		return
	}
	switch VisibilityRestrictions(v.Restrictions) {
	case VisibilityPublic:
		if v0.Restrictions != string(VisibilityPublic) {
			v.Restrictions = v0.Restrictions
		}
	case VisibilityIBMOnly:
		if v0.Restrictions == string(VisibilityPrivate) {
			v.Restrictions = v0.Restrictions
		}
	case VisibilityPrivate:
		// let it be
	default:
		panic(fmt.Sprintf(`Unknown Visibility.Restrictions: "%s"`, v.Restrictions))
	}
	// TODO: Merge Visibilty Include and Exclude lists
}

// Composite represents the Catalog metadata associated with an entry of kind=composite
type Composite struct {
	Children []struct {
		Kind string `json:"kind" mapstructure:"kind"`
		Name string `json:"name" mapstructure:"name"`
	}
	CompositeKind string `json:"composite_kind" mapstructure:"composite_kind"`
	CompositeTag  string `json:"composite_tag" mapstructure:"composite_tag"`
}

// Pricing represents the pricing information associated with a Catalog entry
// (accessed indirectly from ObjectMetaData.Pricing.URL)
type Pricing struct {
	Origin string `json:"origin"`
	Type   string `json:"type"`
	//	I18N           map[string]interface{} `json:"i18n"`
	//	StartingPrice  map[string]interface{} `json:"starting_price"`
	//	EffectiveFrom  string                 `json:"effective_from"`
	//	EffectiveUntil string                 `json:"effective_until"`
	Metrics []struct {
		PartRef             string `json:"part_ref"`
		MetricID            string `json:"metric_id"`
		TierModel           string `json:"tier_model"`
		ResourceDisplayName string `json:"resource_display_name"`
		// More attributes omitted
	} `json:"metrics"`
}

// String returns a short string representation of this Pricing record, esp. MetricIDs (part numbers)
func (p *Pricing) String() string {
	names := collections.NewStringSet()
	parts := collections.NewStringSet()
	for _, m := range p.Metrics {
		names.Add(m.ResourceDisplayName)
		parts.Add(m.MetricID)
	}
	return fmt.Sprintf("Pricing(%q, %v)", names.Slice(), parts.Slice())
}

// String returns a short string representation of this Catalog Resource record
func (r *Resource) String() string {
	if r.CatalogPath != "" && r.CatalogPath != "/" {
		return fmt.Sprintf(`Catalog{Path:"%s", Kind:"%s", ID:"%s"}`, r.CatalogPath, r.Kind, r.ID)
	}
	return fmt.Sprintf(`Catalog{Name:"%s", Kind:"%s", ID:"%s"}`, r.Name, r.Kind, r.ID)
}

var _ fmt.Stringer = &Resource{}

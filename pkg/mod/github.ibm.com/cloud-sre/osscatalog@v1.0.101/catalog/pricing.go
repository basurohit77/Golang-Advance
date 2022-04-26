package catalog

import (
	"fmt"
	"net/http"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

const includeOptionsPricing = "metadata.pricing&languages=en"

// ListPricingInfoFromCatalog finds all the Pricing records associated with a given Catalog resource object and its children,
// *within the Catalog itself, which caches info from the Pricing Catalog*
// and calls the handler function with each Pricing record found
func ListPricingInfoFromCatalog(r *catalogapi.Resource, handler func(p *catalogapi.Pricing)) error {
	ctx, err := setupContextForMainEntries(productionFlagReadOnly, true) // Use a refreshable token because this may be a very slow operation
	if err != nil {
		return err
	}
	_, _, token, err := readContextForMainEntries(ctx, false)
	if err != nil {
		return err
	}

	var recurse func(r *catalogapi.Resource) error
	recurse = func(r *catalogapi.Resource) error {
		debug.Debug(debug.Pricing, "Fetching pricing info from entry name=%s  kind=%s  url=%s  pricing=%v", r.Name, r.Kind, r.URL, r.ObjectMetaData.Pricing)
		if r.ObjectMetaData.Pricing != nil && r.ObjectMetaData.Pricing.URL != "" {
			var pricing = new(catalogapi.Pricing)
			//	r.ObjectMetaData.Pricing.URL = r.ObjectMetaData.Pricing.URL + "/XXX"
			debug.Debug(debug.Pricing, "Fetching pricing record at URL %s", r.ObjectMetaData.Pricing.URL)
			err = rest.DoHTTPGet(r.ObjectMetaData.Pricing.URL, string(token), nil, "Catalog.Pricing", debug.Catalog, pricing)
			if err != nil {
				debug.Debug(debug.Pricing, "No pricing record for %s: %v", r.Name, err)
				if httpErr, ok := err.(rest.HTTPError); ok && httpErr.GetHTTPStatusCode() == http.StatusNotFound {
					debug.Debug(debug.Pricing, "No pricing info for %s(%s): %s", r.Name, r.ObjectMetaData.Pricing.URL, err.Error())
					handler(&catalogapi.Pricing{})
				} else {
					return err
				}
			} else {
				debug.Debug(debug.Pricing, "Got pricing record for %s: %+v", r.Name, pricing)
				handler(pricing)
			}
		}
		if url := r.ChildrenURL; url != "" {
			actualURL := fmt.Sprintf("%s?include=%s", url, includeOptionsPricing)
			var response = new(catalogapi.GetResponse)
			err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.Pricing", debug.Catalog, response)
			if err != nil {
				return err
			}
			if len(response.Resources) > 0 {
				// Refresh the tokens (as multiple recursions can take longer than the normal token expiration interval)
				_, _, token, err = readContextForMainEntries(ctx, false)
				if err != nil {
					return err
				}
			}
			for i := 0; i < len(response.Resources); i++ {
				r := &response.Resources[i]
				err = recurse(r)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	return recurse(r)
}

// BSSPricing is the record type returned from the Pricing Catalog
type BSSPricing struct {
	ID                   string `json:"_id"`
	Rev                  string `json:"_rev"`
	ResourceID           string `json:"resource_id"`
	PlanID               string `json:"plan_id"`
	RegionPricingEnabled bool   `json:"region_pricing_enabled"`
	Free                 bool   `json:"free"`
	Lite                 bool   `json:"lite"`
	Metrics              []struct {
		ChargeUnit            string  `json:"charge_unit"`
		ChargeUnitName        string  `json:"charge_unit_name"`
		ChargeUnitDisplayName string  `json:"charge_unit_display_name"`
		ResourceDisplayName   string  `json:"resource_display_name"`
		UsageCapQty           float64 `json:"usage_cap_qty"`
		DisplayCap            float64 `json:"display_cap"`
		PartNumber            string  `json:"part_number"`
	} `json:"metrics"`
	PromoteState   string `json:"promote_state"`
	EffectiveFrom  string `json:"effective_from"`
	EffectiveUntil string `json:"effective_until"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// String returns a short string representation of this BSSPricing record, esp. MetricIDs (part numbers)
func (p *BSSPricing) String() string {
	names := collections.NewStringSet()
	parts := collections.NewStringSet()
	for _, m := range p.Metrics {
		names.Add(m.ResourceDisplayName)
		parts.Add(m.PartNumber)
	}
	return fmt.Sprintf("Pricing(%q, %v)", names.Slice(), parts.Slice())
}

// ListPricingInfo finds all the Pricing records associated with a given Catalog resource object and its children,
// *directly from the Pricing Catalog*
// and calls the handler function with each Pricing record found
func ListPricingInfo(r *catalogapi.Resource, handler func(p *BSSPricing)) error {
	totalEntries := 0

	ctx, err := setupContextForMainEntries(productionFlagReadOnly, false)
	if err != nil {
		return err
	}
	_, _, ctoken, err := readContextForMainEntries(ctx, false)
	if err != nil {
		return err
	}
	ptoken, err := rest.GetToken(pricingKeyName)
	if err != nil {
		return err
	}

	var recurse func(r *catalogapi.Resource) error
	recurse = func(r *catalogapi.Resource) error {
		debug.Debug(debug.Pricing, "Fetching pricing info from entry name=%s  kind=%s  url=%s  pricing=%v", r.Name, r.Kind, r.URL, r.ObjectMetaData.Pricing)
		if r.Kind == "plan" || r.Kind == "flavor" /* || r.Kind == "deployment" */ {
			var pricing []BSSPricing
			totalEntries++
			if totalEntries > 0 && (totalEntries%10) == 0 {
				debug.Info("    Loading one batch of Pricing entries from Pricing Catalog from %s (%d entries so far)", r.CatalogPath, totalEntries)
			}
			actualURL := fmt.Sprintf(pricingURL, r.ID)
			debug.Debug(debug.Pricing, "Get pricing record at URL %s", actualURL)
			err = rest.DoHTTPGet(actualURL, ptoken, nil, "PricingCatalog", debug.Catalog, &pricing)
			if err != nil {
				if httpErr, ok := err.(rest.HTTPError); ok && httpErr.GetHTTPStatusCode() == http.StatusNotFound {
					if freeTags := ScanFreePricing(r); len(freeTags) > 0 {
						debug.Debug(debug.Pricing, "No pricing info for %s(%s): %s  -- freePricingTags=%v", r.Name, actualURL, err.Error(), freeTags)
					} else {
						return err
					}
				} else {
					return err
				}
			} else {
				for _, p := range pricing {
					p := p
					handler(&p)
				}
			}
		} else { // No need to recurse *inside* a plan
			if url := r.ChildrenURL; url != "" {
				childrenURL := fmt.Sprintf("%s?include=%s", url, includeOptionsPricing)
				var response = new(catalogapi.GetResponse)
				err = rest.DoHTTPGet(childrenURL, string(ctoken), nil, "Catalog.Pricing", debug.Catalog, response)
				if err != nil {
					return err
				}
				for i := 0; i < len(response.Resources); i++ {
					r := &response.Resources[i]
					// Sleep for rate limiting
					time.Sleep(time.Second)
					err := recurse(r)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	return recurse(r)
}

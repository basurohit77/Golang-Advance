package catalog

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// mainIncludes contains all the Catalog "include" options that we need to fetch Main Catalog entries
const mainIncludes = "&include=metadata.other.oss:metadata.ui.hidden:metadata.pricing:metadata.deployment&languages=en&noLocations=true"

//const mainIncludes = "&include=metadata.other.oss:metadata.ui.hidden:metadata.pricing:metadata.deployment&languages=en"

// ReadMainCatalogEntry reads an entire entry (full record) from the Global Catalog
func ReadMainCatalogEntry(name ossrecord.CRNServiceName) (*catalogapi.Resource, error) {
	ctx, err := setupContextForMainEntries(productionFlagReadOnly, false)
	if err != nil {
		return nil, err
	}
	return ReadMainCatalogEntryWithContext(ctx, name)
}

// ReadMainCatalogEntryWithContext reads an entire entry (full record) from the Global Catalog
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
// TODO: This function only works for top-level entries, not entries under a Group
func ReadMainCatalogEntryWithContext(ctx context.Context, name ossrecord.CRNServiceName) (*catalogapi.Resource, error) {
	catalogURL, scope, token, err := readContextForMainEntries(ctx, false)
	if err != nil {
		return nil, err
	}
	actualURL := fmt.Sprintf("%s?q=name:%s%s%s", catalogURL, name, mainIncludes, scope)
	var result = catalogapi.GetResponse{}
	err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.Main", debug.Catalog, &result)
	if err != nil {
		return nil, err
	}

	count := int(result.ResourceCount)
	if count == 0 {
		err := rest.MakeHTTPError(nil, nil, true, "Resource \"%s\" not found in Global Catalog", name)
		return nil, err
	} else if count > 1 {
		panic(fmt.Sprintf("Found %d entries in Global Catalog for the same name \"%s\"", count, name))
	} else {
		ret := &result.Resources[0]
		debug.Debug(debug.Visibility, `Visibility for entry "%s": %+v  (top-level entry, effective visibiity is same as local)`, ret.Name, &ret.Visibility)
		// TODO: If we someday support non top-level entries, we need to recurse through the parents to determine the effective visibility
		return ret, nil
	}
}

//CountMainEntries keeps track of all the counts of OSS entries encountered during ListMainEntries() and its variants
type CountMainEntries struct {
	rawEntries   int
	countsByKind map[string]int
}

func newCountMainEntries() *CountMainEntries {
	ret := &CountMainEntries{}
	ret.countsByKind = make(map[string]int)
	return ret
}

// Total returns the total number of entries returned from ListMainEntries() and its variants (after applying the filter pattern)
func (c *CountMainEntries) Total() int {
	var total int
	for _, n := range c.countsByKind {
		total += n
	}
	return total
}

// Increment incemenents the count of entries of one given Catalog Kind returned from ListMainEntries() and its variants
func (c *CountMainEntries) Increment(kind string) {
	c.countsByKind[kind] = c.countsByKind[kind] + 1
}

// String returns a string representation of all the counts of entries encountered during ListMainEntries() and its variants
func (c *CountMainEntries) String() string {
	var keys = make([]string, 0, len(c.countsByKind))
	for k := range c.countsByKind {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf := strings.Builder{}
	buf.WriteString("{ ")
	buf.WriteString(fmt.Sprintf("rawEntries=%d", c.rawEntries))
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf(", %s=%d", k, c.countsByKind[k]))
	}
	buf.WriteString(" }")
	return buf.String()
}

// ListMainCatalogEntries lists all main Global Catalog entries (for services, runtimes, iaas, etc.) and calls the special handler function for each entry
func ListMainCatalogEntries(pattern *regexp.Regexp, handler func(r *catalogapi.Resource)) error {
	// Note that we use a refreshable token in the Context, because the list operation may take a very long time
	// (e.g. when also fetching Pricing info)
	ctx, err := setupContextForMainEntries(productionFlagReadOnly, true)
	if err != nil {
		return err
	}
	return ListMainCatalogEntriesWithContext(ctx, pattern, handler)
}

// ListMainCatalogEntriesWithContext lists all main Global Catalog entries (for services, runtimes, iaas, etc.) and calls the special handler function for each entry
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func ListMainCatalogEntriesWithContext(ctx context.Context, pattern *regexp.Regexp, handler func(r *catalogapi.Resource)) error {
	var path = ""
	var parentVisibility *catalogapi.Visibility
	var counts = newCountMainEntries()
	catalogURL, _, _, err := readContextForMainEntries(ctx, false)
	if err != nil {
		return err
	}

	err = processEntries(ctx, catalogURL, path, parentVisibility, pattern, counts, handler)
	if err != nil {
		return err
	}

	debug.Info("Read %d Main entries from Global Catalog: %s", counts.Total(), counts.String())
	return nil
}

// processEntries processes a list of entries received from Global Catalog,
// for the implementation of ListMainCatalogEntries()
func processEntries(ctx context.Context, url string, path string, parentVisibility *catalogapi.Visibility, pattern *regexp.Regexp, counts *CountMainEntries, handler func(r *catalogapi.Resource)) error {
	var offset = 0
	for {
		if path == "/" || path == "" {
			debug.Info("Loading one batch of Main entries from Global Catalog (%d/%d entries so far - %s)", counts.Total(), counts.rawEntries, counts.String())
		}
		// Refresh the token (as multiple recursions can take longer than the normal token expiration interval)
		_, scope, token, err := readContextForMainEntries(ctx, false)
		if err != nil {
			return err
		}
		actualURL := fmt.Sprintf("%s?_offset=%d%s%s", url, offset, mainIncludes, scope)
		var result = new(catalogapi.GetResponse)
		err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.Main", debug.Catalog, result)
		if err != nil {
			return err
		}

		count := int(result.ResourceCount)
		if count != len(result.Resources) {
			panic(fmt.Sprintf("ListMainCatalogEntries(%s): GET Main entries from Catalog: Offset=%d  ResourceCount=%d   len(Resources)=%d", actualURL, offset, count, len(result.Resources)))
		}
		debug.Debug(debug.Catalog, `ListMainCatalogEntries.processEntries(%s) -> %d results`, url, count)
		if count == 0 {
			break
		}
		for i := 0; i < len(result.Resources); i++ {
			counts.rawEntries++
			r := &result.Resources[i]
			newPath := fmt.Sprintf("%s/%s", path, r.Name)
			if pattern != nil && pattern.FindString(newPath) == "" {
				continue
			}
			r.CatalogPath = newPath
			r.EffectiveVisibility = r.Visibility
			r.EffectiveVisibility.MergeVisibility(parentVisibility)
			debug.Debug(debug.Visibility, `Visibility for entry "%s": local=%+v  /  effective=%+v`, r.String(), &r.Visibility, &r.EffectiveVisibility)
			switch r.Kind {
			case catalogapi.KindRuntime, catalogapi.KindTemplate, catalogapi.KindPlatformService, catalogapi.KindDeployment, "platform_service_environment": // TODO: What to do about "platform_service_environment" type
				handler(r)
				counts.Increment(r.Kind)

			case catalogapi.KindService, catalogapi.KindIaaS, catalogapi.KindComposite:
				// Note that we must call the handler for a parent entry before any child entries,
				// to allow our client to easily keep track of children
				handler(r)
				counts.Increment(r.Kind)

				// Go through children of all entries (not just Group) because we want to see deployments and plans, and children of composites
				if r.Group || r.Active || r.IsIaaSRegionsPlaceholder(false) {
					// Note: Do not pass the name filter pattern to the second level
					debug.Debug(debug.Catalog, `ListMainCatalogEntries() loading children of service entry %s`, r.String())
					err := processEntries(ctx, r.ChildrenURL, newPath, &r.EffectiveVisibility, nil, counts, handler)
					if err != nil {
						return err
					}
				} else {
					debug.Debug(debug.Catalog, `ListMainCatalogEntries() ignoring children of Inactive entry %s`, r.String())
				}

			case catalogapi.KindPlan, catalogapi.KindFlavor, catalogapi.KindProfile:
				// Note that we must call the handler for a parent entry before any child entries,
				// to allow our client to easily keep track of children
				handler(r)
				counts.Increment(r.Kind)

				if ossrunactions.Deployments.IsEnabled() {
					// Go through children of all entries (not just Group) because we want to see deployments and plans, and children of composites
					if r.Group || r.Active || r.IsIaaSRegionsPlaceholder(false) {
						// Note: Do not pass the name filter pattern to the second level
						debug.Debug(debug.Catalog, `ListMainCatalogEntries() loading children of plan entry %s`, r.String())
						err := processEntries(ctx, r.ChildrenURL, newPath, &r.EffectiveVisibility, nil, counts, handler)
						if err != nil {
							return err
						}
					} else {
						debug.Debug(debug.Catalog, `ListMainCatalogEntries() ignoring children of Inactive entry %s`, r.String())
					}
				} else {
					debug.Debug(debug.Catalog, `ListMainCatalogEntries() ignoring children of plan/flavor/profile when ossrunactions.Deployments is disabled: %s`, r.String())
				}

			case catalogapi.KindRegion, catalogapi.KindDatacenter, catalogapi.KindAvailabilityZone, catalogapi.KindPOP,
				catalogapi.KindLegacyCName, catalogapi.KindLegacyEnvironment, catalogapi.KindSatellite:
				handler(r)
				counts.Increment(r.Kind)
				// Note: Do not pass the name filter pattern to the second level
				err := processEntries(ctx, r.ChildrenURL, newPath, &r.EffectiveVisibility, nil, counts, handler)
				if err != nil {
					return err
				}

			case catalogapi.KindGeography, catalogapi.KindCountry, catalogapi.KindMetro:
				// Examine children but do not return this entry itself
				// Note: Do not pass the name filter pattern to the second level
				err := processEntries(ctx, r.ChildrenURL, newPath, &r.EffectiveVisibility, nil, counts, handler)
				if err != nil {
					return err
				}

			case catalogapi.KindLocation:
				// ignore

			case "dataset", "ui-dashboard", "ssm-service", "alias",
				"instance.profile", "extension-point", "volume.profile", "application", "catalog_schema", "buildpack", "cfee-buildpacks",
				"index-version-key", "broker", "catalog_root", "dedicated-host.profile", "share.profile", "load-balancer.profile", "bare-metal-server.profile",
				catalogapi.KindPlugin, catalogapi.KindISProfiles:
				// ignore

				/*
					case catalogapi.KindLegacyEnvironment, catalogapi.KindLegacyCName:
						debug.Debug(debug.Catalog, `Ignoring old-style legacy entry %s  --> keep traversing the hierarchy`, r.String())
						if r.Visibility.Restrictions != string(catalogapi.VisibilityPublic) {
							catalogName := getCatalogNameFromContext(ctx)
							debug.Info("Ignoring non-public location entry in %s Global Catalog: %s - visibility=%s", catalogName, r.String(), r.Visibility.Restrictions)
							continue
						}
						lookForChildren = true
						// ignore but keep going to look for children
				*/

			case catalogapi.KindOSSService, catalogapi.KindOSSSegment, catalogapi.KindOSSTribe, catalogapi.KindOSSEnvironment, catalogapi.KindOSSSchema:
				// ignore OSS entries in Production Catalog -- this call is focused on the *Main* entries

			default:
				debug.PrintError(`ListMainCatalogEntries() Found unknown entry kind in Global Catalog: "%s"  (name=%s  ID=%q)`, r.Kind, r.Name, r.ID)
			}
		}
		offset += count
	}
	return nil
}

// GetMainCatalogEntryUI returns a URL for viewing the Main entry with the given ID in the Global Catalog UI
func GetMainCatalogEntryUI(id ossrecord.CatalogID) (string, error) {
	url0, err := url.Parse(catalogMainURL)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("GetMainCatalogEntryUI(): no ID for Catalog resource")
		//		return fmt.Sprintf("%s://%s/search?q=name:%s", url0.Scheme, url0.Host, url.QueryEscape(string(name))), nil
	}
	return fmt.Sprintf("%s://%s/update/%s", url0.Scheme, url0.Host, url.PathEscape(string(id))), nil
}

// ListPlans lists all the Plan records associated with a given Main Catalog Entry and calls the handler function for each
func ListPlans(r *catalogapi.Resource, handler func(r *catalogapi.Resource)) error {
	ctx, err := setupContextForMainEntries(productionFlagReadOnly, false)
	if err != nil {
		return err
	}
	return ListPlansWithContext(ctx, r, handler)
}

// ListPlansWithContext lists all the Plan records associated with a given Main Catalog Entry and calls the handler function for each
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func ListPlansWithContext(ctx context.Context, r *catalogapi.Resource, handler func(r *catalogapi.Resource)) error {
	if r.ChildrenURL != "" {
		_, scope, token, err := readContextForMainEntries(ctx, false)
		if err != nil {
			return err
		}
		actualURL := fmt.Sprintf("%s?_offset=%d%s", r.ChildrenURL, 0, scope)
		var response = new(catalogapi.GetResponse)
		err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.Plans", debug.Catalog, response)
		if err != nil {
			return err
		}
		for i := 0; i < len(response.Resources); i++ {
			r := &response.Resources[i]
			if r.Kind != "plan" {
				continue
			}
			handler(r)
		}
	}
	return nil
}

// GetDeployments gets a Map of all the deployment records associated with a given Main Catalog Entry
func GetDeployments(r *catalogapi.Resource, deploymentMap map[string]catalogapi.Resource, serviceName string) (map[string]catalogapi.Resource, error) {
	ctx, err := setupContextForMainEntries(productionFlagReadOnly, false)
	if err != nil {
		return nil, err
	}
	return GetDeploymentsWithContext(ctx, r, deploymentMap, serviceName)
}

// GetDeploymentsWithContext gets a Map of all the deployment records associated with a given Main Catalog Entry
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func GetDeploymentsWithContext(ctx context.Context, r *catalogapi.Resource, deploymentMap map[string]catalogapi.Resource, serviceName string) (map[string]catalogapi.Resource, error) {
	if deploymentMap == nil || &deploymentMap == nil {
		deploymentMap = map[string]catalogapi.Resource{}
	}

	if r.ChildrenURL != "" {
		_, scope, token, err := readContextForMainEntries(ctx, false)
		if err != nil {
			return nil, err
		}
		actualURL := fmt.Sprintf("%s?_offset=%d%s", r.ChildrenURL, 0, scope)

		for len(actualURL) > 0 {
			var response = new(catalogapi.GetResponse)
			err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.Deployments", debug.Catalog, response)
			if err != nil {
				return nil, err
			}
			// Anything at the first level is likely a Plan
			for i := 0; i < len(response.Resources); i++ {
				dr := &response.Resources[i]
				// log.Println("pr.Child: " + dr.ChildrenURL)
				if dr.Kind == catalogapi.KindPlan {
					_, err := GetDeployments(dr, deploymentMap, serviceName)
					if err != nil {
						return nil, err
					}
				} else if dr.Kind == catalogapi.KindDeployment {
					crnSegments := strings.Split(dr.CatalogCRN, ":")
					// After splitting the catalog crn, which we use for the base CRN for our purposes
					// we need want to build a crn based on the service name
					// and the geotag associated with the deployment record.
					crnSegments[4] = string(serviceName)
					if len(dr.GeoTags) > 0 {
						crnSegments[5] = dr.GeoTags[0]
					}

					crnString := strings.Join(crnSegments[0:6], ":") + "::::"
					// Add deployment if it is not already there
					dm := deploymentMap
					if _, ok := dm[crnString]; !ok {
						dm[crnString] = *dr
					}
				} else {
					continue
				}

			}
			actualURL = response.Next
		}
	}
	return deploymentMap, nil
}

// ListDataCenterEntries lists all DataCenter entries in Global Catalog entries and calls the special handler function for each entry
func ListDataCenterEntries(handler func(r *catalogapi.Resource)) error {
	ctx, err := setupContextForMainEntries(productionFlagReadOnly, false)
	if err != nil {
		return err
	}
	return ListDataCenterEntriesWithContext(ctx, handler)
}

// ListDataCenterEntriesWithContext lists all DataCenter entries in Global Catalog entries and calls the special handler function for each entry
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func ListDataCenterEntriesWithContext(ctx context.Context, handler func(r *catalogapi.Resource)) error {
	catalogURL, scope, token, err := readContextForMainEntries(ctx, false)
	if err != nil {
		return err
	}
	rawEntries := 0
	totalEntries := 0
	offset := 0

	for {
		debug.Info("Loading one batch of DC entries from Global Catalog (%d/%d entries so far)", totalEntries, rawEntries)
		actualURL := fmt.Sprintf("%s?_offset=%d%s&q=kind:geography%s", catalogURL, offset, mainIncludes, scope)
		var result = new(catalogapi.GetResponse)
		err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.Main", debug.Catalog, result)
		if err != nil {
			return err
		}

		count := int(result.ResourceCount)
		if count != len(result.Resources) {
			panic(fmt.Sprintf("ListDataCenterEntries(): GET Main entries from Catalog: Offset=%d  ResourceCount=%d   len(Resources)=%d", offset, count, len(result.Resources)))
		}
		if count == 0 {
			break
		}
		rawEntries += count
		entryCount, err := getDatacenters(result.Resources, string(token), handler)
		totalEntries += entryCount
		if err != nil {
			return err
		}

		offset += count
	}

	debug.Info("Read %d DC entries from Global Catalog", totalEntries)
	return nil
}

func getDatacenters(resources []catalogapi.Resource, token string, handler func(r *catalogapi.Resource)) (numEntries int, err error) {
	numEntries = 0
	for i := 0; i < len(resources); i++ {

		r := &resources[i]
		//log.Print("Processing ", r.Name, " -> ", r.Kind)
		if r.Kind == "dc" {
			numEntries++
			handler(r)
		} else {
			if r.ChildrenURL != "" {
				/*
					_, scope, _, err := readContextForMainEntries(ctx, false)
					if err != nil {
						return nil, err
					}
				*/
				actualURL := fmt.Sprintf("%s?_offset=%d%s", r.ChildrenURL, 0 /* scope */, "")
				var result = new(catalogapi.GetResponse)
				err = rest.DoHTTPGet(actualURL, token, nil, "Catalog.Main", debug.Catalog, result)
				if err != nil {
					return 0, err
				}
				cnt, err := getDatacenters(result.Resources, token, handler)
				numEntries += cnt
				if err != nil {
					return 0, err
				}
			} else {
				debug.PrintError(`getDatacenters: entry with nil children_url for name="%s"  kind=\"%s"`, r.Name, r.Kind)
			}

		}

	}
	return numEntries, nil

}

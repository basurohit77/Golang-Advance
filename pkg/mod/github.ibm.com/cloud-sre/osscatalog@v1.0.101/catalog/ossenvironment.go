package catalog

import (
	"context"
	"fmt"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// Functions for reading the list of OSSEnvironments from the Global Catalog

/*
// catalogURLEnvironments is the URL to use to access cname and environment records in Global Catalog
// const catalogURLEnvironments = catalogMainURL
const catalogURLEnvironments = catalogOSSURLStaging

// catalogKeyNameEnvironment is the name of the API key from which to obtain a IAM token for access to cname and environment records in Global Catalog
// const catalogKeyNameEnvironment = catalogMainKeyName
const catalogKeyNameEnvironment = catalogOSSKeyNameStaging
*/

/*
var patternPublicCRN = regexp.MustCompile(`^crn:v1:[a-z0-9-]*:(public|):`)
var patternDedicatedCRN = regexp.MustCompile(`^crn:v1:[a-z0-9-]+:dedicated:`)
var patternLocalCRN = regexp.MustCompile(`^crn:v1:[a-z0-9-]+:local:`)
var patternStagingCRN = regexp.MustCompile(`^crn:v1:[a-z0-9-]+:staging:`)
*/

// ListOSSEnvironments lists all OSSEnvironment entries in Global Catalog
// from the Staging instance (which contains the most recently updated data)
func ListOSSEnvironments(pattern *regexp.Regexp, incl IncludeOptions, handler func(r ossrecord.OSSEntry)) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	return ListOSSEnvironmentsWithContext(ctx, pattern, incl, handler)
}

// ListOSSEnvironmentsProduction lists OSSEnvironment entries in Global Catalog
// from the Production instance (which contains the stable data)
func ListOSSEnvironmentsProduction(pattern *regexp.Regexp, incl IncludeOptions, handler func(r ossrecord.OSSEntry)) error {
	ctx, err := setupContextForOSSEntries(productionFlagReadOnly)
	if err != nil {
		return err
	}
	return ListOSSEnvironmentsWithContext(ctx, pattern, incl, handler)
}

// ListOSSEnvironmentsWithContext lists all OSSEnvironment entries found by scanning underlying
// region, datacenter, zone and oss_environment entries in Global Catalog
// and calls the special handler function for each entry.
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func ListOSSEnvironmentsWithContext(ctx context.Context, pattern *regexp.Regexp, incl IncludeOptions, handler func(r ossrecord.OSSEntry)) error {
	incl |= IncludeEnvironments
	return ListOSSEntriesWithContext(ctx, pattern, incl, handler)
}

// ReadOSSEnvironment reads an OSSEnvironment entry from the Global Catalog, given its EnvironmentID,
// in the Staging instance (which contains the most recently updated data)
func ReadOSSEnvironment(id ossrecord.EnvironmentID) (env *ossrecord.OSSEnvironment, err error) {
	id2 := ossrecord.MakeOSSEnvironmentID(id)
	e, err := ReadOSSEntryByID(id2, IncludeEnvironments)
	if err != nil {
		return nil, err
	}
	var ok bool
	if env, ok = e.(*ossrecord.OSSEnvironment); !ok {
		err := fmt.Errorf("ReadOSSEnvironment(%s) returned Catalog Resource of unexpected type %T (%v)", id, e, e)
		return nil, err
	}
	if env.EnvironmentID != id {
		err := fmt.Errorf("ReadOSSEnvironment(%s) returned entry with unexpected id %s", id, env.EnvironmentID)
		return nil, err
	}
	return env, nil
}

// ReadOSSEnvironmentProduction reads an OSSEnvironment entry from the Global Catalog, given its EnvironmentID,
// in the Production instance (which contains the stable data)
func ReadOSSEnvironmentProduction(id ossrecord.EnvironmentID) (env *ossrecord.OSSEnvironment, err error) {
	id2 := ossrecord.MakeOSSEnvironmentID(id)
	e, err := ReadOSSEntryByIDProduction(id2, IncludeEnvironments)
	if err != nil {
		return nil, err
	}
	var ok bool
	if env, ok = e.(*ossrecord.OSSEnvironment); !ok {
		err := fmt.Errorf("ReadOSSEnvironment(%s) returned Catalog Resource of unexpected type %T (%v)", id, e, e)
		return nil, err
	}
	if env.EnvironmentID != id {
		err := fmt.Errorf("ReadOSSEnvironment(%s) returned entry with unexpected id %s", id, env.EnvironmentID)
		return nil, err
	}
	return env, nil
}

// makeNativeOSSEnvironmentEntry creates a "ghost" OSSEnvironment entry that is not represented by a physical OSS entry in Global Catalog
// but insteas is populated on the fly from information contained in a "Main" Catalog entry for a Region or Datacenter.
// Used only with the ossrunactions.EnvironmentsNative
func makeNativeOSSEnvironmentEntry(r *catalogapi.Resource, parentID string, path string) (entry ossrecord.OSSEntry, newParentID string, err error) {
	env := &ossrecord.OSSEnvironment{}
	entry = env
	env.SchemaVersion = ossrecord.OSSCurrentSchema
	env.OSSTags.AddTag(osstags.CatalogNative)
	if r.ObjectMetaData.Deployment == nil {
		err := fmt.Errorf(`Region/Environment resource found in Global Catalog has no "deployment" metadata field: %s`, r.String())
		return nil, "", err
	}
	normalizedCRN, err := crn.ParseAndNormalize(r.ObjectMetaData.Deployment.TargetCRN,
		crn.ExpectEnvironment, crn.AllowBlankCName, crn.AllowSoftLayerCName /* XXX Copied from ossmerge.mergenvironments.go */)
	if err != nil {
		err := fmt.Errorf("ossmerge.LoadAllEntries.MainCatalog: found region-related entry with invalid target_crn: %s  -- %v", r.String(), err)
		return nil, "", err
	}
	env.EnvironmentID = ossrecord.EnvironmentID(normalizedCRN.ToCRNString())
	env.ParentID = ossrecord.EnvironmentID(parentID)
	env.DisplayName = r.OverviewUI.En.DisplayName
	if normalizedCRN.IsIBMPublicCloud() {
		switch r.Kind {
		case catalogapi.KindRegion:
			env.Type = ossrecord.EnvironmentIBMCloudRegion
			newParentID = string(normalizedCRN.ToCRNString())
		case catalogapi.KindDatacenter:
			env.Type = ossrecord.EnvironmentIBMCloudDatacenter
			newParentID = string(normalizedCRN.ToCRNString())
		case catalogapi.KindAvailabilityZone:
			env.Type = ossrecord.EnvironmentIBMCloudZone
			newParentID = string(normalizedCRN.ToCRNString())
		case catalogapi.KindPOP:
			env.Type = ossrecord.EnvironmentIBMCloudPOP
			newParentID = string(normalizedCRN.ToCRNString())
		case catalogapi.KindSatellite:
			env.Type = ossrecord.EnvironmentIBMCloudSatellite
			newParentID = string(normalizedCRN.ToCRNString())
		default:
			panic(fmt.Sprintf(`ListOSSEnvironmentsWithContext() unexpected entry kind in Global Catalog: "%s"  (entry=%s)`, r.Kind, r.String()))
		}
	} else {
		debug.PrintError(`ListOSSEnvironmentsWithContext() entry has target_crn that does not match a known pattern for Public: "%s" (entry=%s)`, r.ObjectMetaData.Deployment.TargetCRN, r.String())
	}
	env.Status = ossrecord.EnvironmentActive // TODO: check Catalog visibility
	env.ReferenceCatalogID = ossrecord.CatalogID(r.ID)
	env.ReferenceCatalogPath = r.CatalogPath
	env.OwningSegment = "XXX"
	env.Description = fmt.Sprintf("%s    [catalog_path=%s]", r.OverviewUI.En.Description, path)
	env.LegacyMCCPID = r.ObjectMetaData.Deployment.MCCPID
	//env.LegacyIMSID = ""
	env.Status = ossrecord.EnvironmentActive
	return env, newParentID, nil
}

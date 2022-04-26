package catalog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"

	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/options"

	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// IncludeOptions represents options for including various type of records and OSSMergeControl/OSSValidation info in read/write operations
type IncludeOptions uint32

// Possible values for IncludeOptions
const (
	IncludeNone                      IncludeOptions = 0
	IncludeOSSMergeControl           IncludeOptions = 0x1
	IncludeOSSValidation             IncludeOptions = 0x2
	IncludeOSSTimestamps             IncludeOptions = 0x4
	IncludeServices                  IncludeOptions = 0x1000
	IncludeTribes                    IncludeOptions = 0x2000
	IncludeEnvironments              IncludeOptions = 0x4000
	IncludeEnvironmentsNative        IncludeOptions = 0x8000
	IncludeServicesDomainOverrides   IncludeOptions = 0x10000
	IncludeServicesDomainCommercial  IncludeOptions = 0x20000
	IncludeServicesDomainUSRegulated IncludeOptions = 0x40000
	IncludeOSSResourceClassification IncludeOptions = 0x80000

	IncludeAll IncludeOptions = IncludeOSSMergeControl | IncludeOSSValidation | IncludeOSSTimestamps | IncludeServices | IncludeTribes | IncludeEnvironments
)

// parentIDMap keeps track of the parentID of every OSSResource read from the Catalog, to determine if we need to specify the "move" parameter on update
type parentIDMapType struct {
	theMap map[string]string
	lock   sync.Mutex
}

var parentIDMap = parentIDMapType{
	theMap: make(map[string]string),
}

func (p *parentIDMapType) Set(catalogOSSURL string, childID string, parentID string) {
	p.lock.Lock()
	parentIDMap.theMap[fmt.Sprintf("%s/%s", catalogOSSURL, childID)] = parentID
	p.lock.Unlock()
}

func (p *parentIDMapType) UnSet(catalogOSSURL string, childID string) {
	p.lock.Lock()
	delete(parentIDMap.theMap, fmt.Sprintf("%s/%s", catalogOSSURL, childID))
	p.lock.Unlock()
}

func (p *parentIDMapType) Lookup(catalogOSSURL string, childID string) (parentID string, found bool) {
	p.lock.Lock()
	parentID, found = parentIDMap.theMap[fmt.Sprintf("%s/%s", catalogOSSURL, childID)]
	p.lock.Unlock()
	return parentID, found
}

func (opt *IncludeOptions) normalize() {
	if *opt&(IncludeServices|IncludeTribes|IncludeEnvironments|IncludeEnvironmentsNative) == 0 {
		*opt |= IncludeServices | IncludeTribes | IncludeEnvironments
	}
	if (*opt & (IncludeOSSMergeControl | IncludeOSSValidation)) != 0 {
		*opt |= IncludeOSSTimestamps
	}
	cleaned := *opt & (IncludeOSSMergeControl | IncludeOSSValidation | IncludeOSSTimestamps | IncludeServices | IncludeTribes | IncludeEnvironments | IncludeEnvironmentsNative | IncludeServicesDomainOverrides | IncludeServicesDomainCommercial | IncludeServicesDomainUSRegulated | IncludeOSSResourceClassification)
	if cleaned != *opt {
		panic(fmt.Sprintf("Unknown IncludeOptions: 0x%x  (full options=0x%x)", cleaned, *opt))
	}
	if (cleaned & (IncludeEnvironments | IncludeEnvironmentsNative)) == (IncludeEnvironments | IncludeEnvironmentsNative) {
		panic(fmt.Sprintf(`IncludeOptions "IncludeEnvironments"(=0x%x) and "IncludeEnvironmentsNative"(=0x%x) are incompatible  (full options=0x%x)`, IncludeEnvironments, IncludeEnvironmentsNative, *opt))
	}
}

// GetOSSProductionURL returns the URL used to access the Production Catalog for OSS records
func GetOSSProductionURL(readwrite bool) string {
	var ctx context.Context
	var err error
	if readwrite {
		ctx, err = setupContextForOSSEntries(productionFlagReadWrite)
	} else {
		ctx, err = setupContextForOSSEntries(productionFlagReadOnly)
	}
	if err != nil {
		errstr := fmt.Sprintf("Cannot get URL for Production Catalog: %v", err)
		debug.PrintError(errstr)
		return "*** " + errstr + " ***"
	}
	url0, _, _, err := readContextForOSSEntries(ctx, readwrite)
	if err != nil {
		errstr := fmt.Sprintf("Cannot get URL for Production Catalog: %v", err)
		debug.PrintError(errstr)
		return "*** " + errstr + " ***"
	}
	return url0
}

// GetOSSStagingURL returns the URL used to access the Staging Catalog for OSS records
func GetOSSStagingURL(readwrite bool) string {
	var ctx context.Context
	var err error
	ctx, err = setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		errstr := fmt.Sprintf("Cannot get URL for Staging Catalog: %v", err)
		debug.PrintError(errstr)
		return "*** " + errstr + " ***"
	}
	url0, _, _, err := readContextForOSSEntries(ctx, readwrite)
	if err != nil {
		errstr := fmt.Sprintf("Cannot get URL for Staging Catalog: %v", err)
		debug.PrintError(errstr)
		return "*** " + errstr + " ***"
	}
	return url0
}

var canonicalNamePattern = regexp.MustCompile(`[^a-z0-9-]`)

// makeOSSServiceName creates a name for a OSS service/component record in Global Catalog.
// We cannot use the CRNServiceName itself, because that one may be used by the main entry, not the OSS entry
func makeOSSServiceName(service *ossrecord.OSSService) string {
	return "oss." + string(service.ReferenceResourceName)
}

// makeOSSSegmentName creates a name for a OSS Segment entry in Global Catalog.
// We cannot use the Segment name itself, because that one may be used by the main entry, not the OSS entry
func makeOSSSegmentName(seg *ossrecord.OSSSegment) string {
	canonical := canonicalNamePattern.ReplaceAllLiteralString(strings.ToLower(seg.DisplayName), "-")
	if strings.HasPrefix(canonical, "test-record-") {
		// Special hack to compensate for rennaming of "*TEST RECORD*" to "TEST RECORD:"
		// See https://github.ibm.com/cloud-sre/osscatalog/issues/268
		canonical = "-" + canonical
	}
	return "oss-segment." + canonical
}

// makeOSSTribeName creates a name for a OSS Tribe entry in Global Catalog.
// We cannot use the Tribe name itself, because that one may be used by the main entry, not the OSS entry
func makeOSSTribeName(tr *ossrecord.OSSTribe) string {
	canonical := canonicalNamePattern.ReplaceAllLiteralString(strings.ToLower(tr.DisplayName), "-")
	if strings.HasPrefix(canonical, "test-record-") {
		// Special hack to compensate for rennaming of "*TEST RECORD*" to "TEST RECORD:"
		// See https://github.ibm.com/cloud-sre/osscatalog/issues/268
		canonical = "-" + canonical
	}
	return "oss-tribe." + canonical
}

// makeOSSEnvironmentName creates a name for a OSS Environment entry in Global Catalog.
// We cannot use the Environment name itself, because that one may be used by the main entry, not the OSS entry
func makeOSSEnvironmentName(env *ossrecord.OSSEnvironment) string {
	crnMask, _ := crn.Parse(string(env.EnvironmentID))
	canonical := fmt.Sprintf("%s-%s-%s", crnMask.CName, crnMask.CType, crnMask.Location)
	//	canonical := canonicalNamePattern.ReplaceAllLiteralString(strings.ToLower(env.DisplayName), "-")
	return "oss-environment." + canonical
}

// makeOSSCatalogResource allocates and initializes the basic information for a OSS entry for the Global Catalog
func makeOSSCatalogResource(e ossrecord.OSSEntry, incl IncludeOptions) (*catalogapi.Resource, error) {
	err := e.CheckSchemaVersion()
	if err != nil {
		if options.GlobalOptions().Lenient {
			debug.PrintError(`(Lenient Mode): Allowing WRITE with invalid entry schema version: %v"`, err)
		} else {
			return nil, err
		}
	}

	ossresource := &catalogapi.Resource{
		//Kind:
		//Name:
		// Images:
		Active:      true,
		Disabled:    false,
		Tags:        []string{""},
		GeoTags:     []string{""},
		PricingTags: []string{""},
		Group:       false,
		//Provider:
	}
	ossresource.Images.Image = "https://www.ibm.com/cloud-computing/images/Cloud_Bluemix_banner-1.png" //TODO: get a real image for OSS entries
	ossresource.Provider.Name = "OSS Catalog Tooling"                                                  //TODO: - use actual Provider name for OSS entries
	ossresource.Provider.Email = "osscat@us.ibm.com"
	// Cannot set visibility directly from the main entry POST/PUT; need to use the special /visibility endpoint
	// ossresource.Visibility.Restrictions = string(VisibilityIBMOnly)

	switch e0 := e.(type) {
	case *ossrecord.OSSService, *ossrecordextended.OSSServiceExtended:
		var ex *ossrecordextended.OSSServiceExtended
		switch e2 := e0.(type) {
		case *ossrecord.OSSService:
			ex = &ossrecordextended.OSSServiceExtended{OSSService: *e2}
		case *ossrecordextended.OSSServiceExtended:
			ex = e2
		}
		ex.GeneralInfo.Domain = ossrecord.COMMERCIAL // default domain for a service record
		_ = ex.CheckConsistency(true)
		entryName := ex.OSSService.ReferenceResourceName
		ossresource.Kind = catalogapi.KindOSSService
		ossresource.ID = string(ex.GetOSSEntryID())
		ossresource.Name = makeOSSServiceName(&ex.OSSService)
		fullName := fmt.Sprintf("%s(%s)", ex.OSSService.ReferenceDisplayName, entryName)
		ossresource.OverviewUI.En.DisplayName = "OSS Record: " + fullName
		ossresource.OverviewUI.En.Description = "OSS service/component record for " + fullName
		buf := strings.Builder{}
		buf.WriteString("OSS service/component record for ")
		buf.WriteString(fullName)
		buf.WriteString("\n")
		buf.WriteString("\n")
		buf.WriteString(ex.OSSService.Header())
		if ex.OSSValidation != nil {
			buf.WriteString(ex.OSSValidation.Header())
		}
		ossresource.OverviewUI.En.LongDescription = buf.String()
		err := ex.OSSService.GeneralInfo.OSSTags.Validate(true)
		if err != nil {
			return nil, debug.WrapError(err, `Invalid OSSService.OSSTags in "%s"`, entryName)
		}
		ossresource.ObjectMetaData.Other.OSS = &ex.OSSService
		if (incl & IncludeOSSMergeControl) != 0 {
			if ex.OSSMergeControl == nil {
				if options.GlobalOptions().Lenient {
					debug.Warning(`(Lenient Mode): Cannot include nil OSSMergeControl in OSS record for "%s"`, entryName)
				} else {
					panic(fmt.Sprintf(`Cannot include nil OSSMergeControl in OSS record for "%s"`, entryName))
				}
			} else {
				err := ex.OSSMergeControl.OSSTags.Validate(false)
				if err != nil {
					return nil, debug.WrapError(err, `Invalid OSSMergeControl.OSSTags in "%s"`, entryName)
				}
				ossresource.ObjectMetaData.Other.OSSMergeControl = ex.OSSMergeControl
			}
		}
		if (incl & IncludeOSSValidation) != 0 {
			if ex.OSSValidation == nil {
				if options.GlobalOptions().Lenient {
					debug.Warning(`(Lenient Mode): Cannot include nil OSSValidation in OSS record for "%s"`, entryName)
					ossresource.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(ex.OSSService.ReferenceResourceName), options.GlobalOptions().LogTimeStamp)
				} else {
					panic(fmt.Sprintf(`Cannot include nil OSSValidation in OSS record for "%s"`, entryName))
				}
			} else {
				ossresource.ObjectMetaData.Other.OSSValidation = ex.OSSValidation
			}
		}
	case *ossrecord.OSSSegment, *ossrecordextended.OSSSegmentExtended:
		var ex *ossrecordextended.OSSSegmentExtended
		switch e2 := e0.(type) {
		case *ossrecord.OSSSegment:
			ex = &ossrecordextended.OSSSegmentExtended{OSSSegment: *e2}
		case *ossrecordextended.OSSSegmentExtended:
			ex = e2
		}
		entryName := ex.OSSSegment.String()
		ossresource.Kind = catalogapi.KindOSSSegment
		ossresource.ID = string(ex.GetOSSEntryID())
		ossresource.Name = makeOSSSegmentName(&ex.OSSSegment)
		ossresource.OverviewUI.En.DisplayName = "OSS Segment:  " + ex.OSSSegment.DisplayName
		ossresource.OverviewUI.En.Description = "OSS Segment entry for " + string(entryName)
		buf := strings.Builder{}
		buf.WriteString("OSS Segment entry for ")
		buf.WriteString(string(entryName))
		buf.WriteString("\n")
		buf.WriteString(ex.OSSSegment.String())
		buf.WriteString("\n")
		if ex.OSSValidation != nil {
			buf.WriteString(ex.OSSValidation.Header())
		}
		ossresource.OverviewUI.En.LongDescription = buf.String()
		err := ex.OSSSegment.OSSTags.Validate(true)
		if err != nil {
			return nil, debug.WrapError(err, `Invalid OSSSegment.OSSTags in "%s"`, entryName)
		}
		ossresource.ObjectMetaData.Other.OSSSegment = &ex.OSSSegment
		if (incl & IncludeOSSValidation) != 0 {
			if ex.OSSValidation == nil {
				debug.Info("Omit writing OSSValidation (nil) for %s", ex.String())
				ossresource.ObjectMetaData.Other.OSSValidation = nil
				/* XXX
				if options.GlobalOptions().Lenient {
					debug.Warning(`(Lenient Mode): Cannot include nil OSSValidation in OSS record for "%s"`, entryName)
					ossresource.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(ex.OSSSegment.SegmentID), options.GlobalOptions().LogTimeStamp)
				} else {
					panic(fmt.Sprintf(`Cannot include nil OSSValidation in OSS record for "%s"`, entryName))
				}
				*/
			} else {
				ossresource.ObjectMetaData.Other.OSSValidation = ex.OSSValidation
			}
		}
	case *ossrecord.OSSTribe, *ossrecordextended.OSSTribeExtended:
		var ex *ossrecordextended.OSSTribeExtended
		switch e2 := e0.(type) {
		case *ossrecord.OSSTribe:
			ex = &ossrecordextended.OSSTribeExtended{OSSTribe: *e2}
		case *ossrecordextended.OSSTribeExtended:
			ex = e2
		}
		entryName := ex.OSSTribe.String()
		ossresource.Kind = catalogapi.KindOSSTribe
		ossresource.ID = string(ex.GetOSSEntryID())
		ossresource.ParentID = string(ossrecord.MakeOSSSegmentID(ex.OSSTribe.SegmentID))
		ossresource.Name = makeOSSTribeName(&ex.OSSTribe)
		ossresource.OverviewUI.En.DisplayName = "OSS Tribe: " + ex.OSSTribe.DisplayName
		ossresource.OverviewUI.En.Description = "OSS Tribe entry for " + string(entryName) + " (description)"
		buf := strings.Builder{}
		buf.WriteString("OSS Tribe entry for ")
		buf.WriteString(string(entryName))
		buf.WriteString("\n")
		buf.WriteString(ex.OSSTribe.String())
		buf.WriteString("\n")
		// TODO: find a way to display the Segment name in the long description for a Tribe OSS record
		buf.WriteString("Segment ID: ")
		buf.WriteString(string(ex.OSSTribe.SegmentID))
		buf.WriteString("\n")
		if ex.OSSValidation != nil {
			buf.WriteString(ex.OSSValidation.Header())
		}
		ossresource.OverviewUI.En.LongDescription = buf.String()
		err := ex.OSSTribe.OSSTags.Validate(true)
		if err != nil {
			return nil, debug.WrapError(err, `Invalid OSSTribe.OSSTags in "%s"`, entryName)
		}
		ossresource.ObjectMetaData.Other.OSSTribe = &ex.OSSTribe
		if (incl & IncludeOSSValidation) != 0 {
			if ex.OSSValidation == nil {
				debug.Info("Omit writing OSSValidation (nil) for %s", ex.String())
				ossresource.ObjectMetaData.Other.OSSValidation = nil
				/*
					if options.GlobalOptions().Lenient {
						debug.Warning(`(Lenient Mode): Cannot include nil OSSValidation in OSS record for "%s"`, entryName)
						//ossresource.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(ex.OSSTribe.TribeID), options.GlobalOptions().LogTimeStamp)
					} else {
						panic(fmt.Sprintf(`Cannot include nil OSSValidation in OSS record for "%s"`, entryName))
					}
				*/
			} else {
				ossresource.ObjectMetaData.Other.OSSValidation = ex.OSSValidation
			}
		}
	case *ossrecord.OSSEnvironment, *ossrecordextended.OSSEnvironmentExtended:
		var ex *ossrecordextended.OSSEnvironmentExtended
		switch e2 := e0.(type) {
		case *ossrecord.OSSEnvironment:
			ex = &ossrecordextended.OSSEnvironmentExtended{OSSEnvironment: *e2}
		case *ossrecordextended.OSSEnvironmentExtended:
			ex = e2
		}
		entryName := ex.OSSEnvironment.String()
		if ex.OSSTags.Contains(osstags.CatalogNative) {
			panic(fmt.Sprintf(`Attempting to create a new Catalog Resource for OSSEnvironment that has the "%s" tag: %s`, osstags.CatalogNative, entryName))
		}
		ossresource.Kind = catalogapi.KindOSSEnvironment
		ossresource.ID = string(ex.GetOSSEntryID())
		if ex.OSSEnvironment.OwningSegment != "" {
			ossresource.ParentID = string(ossrecord.MakeOSSSegmentID(ex.OSSEnvironment.OwningSegment))
		} else {
			ossresource.ParentID = ""
		}
		// XXX		ossresource.ParentID = string(ossrecord.MakeOSSEnvironmentID(e1.ParentID))
		ossresource.Name = makeOSSEnvironmentName(&ex.OSSEnvironment)
		ossresource.OverviewUI.En.DisplayName = "OSS Environment: " + ex.OSSEnvironment.DisplayName
		ossresource.OverviewUI.En.Description = "OSS Environment entry for " + string(entryName) + " (description)"
		buf := strings.Builder{}
		buf.WriteString("OSS Environment entry for ")
		buf.WriteString(string(entryName))
		buf.WriteString("\n")
		buf.WriteString(ex.OSSEnvironment.String())
		buf.WriteString("\n")
		buf.WriteString("Parent ID: ")
		buf.WriteString(string(ex.OSSEnvironment.ParentID))
		buf.WriteString("\n")
		if ex.OSSValidation != nil {
			buf.WriteString(ex.OSSValidation.Header())
		}
		ossresource.OverviewUI.En.LongDescription = buf.String()
		err := ex.OSSEnvironment.OSSTags.Validate(true)
		if err != nil {
			return nil, debug.WrapError(err, `Invalid OSSEnvironment.OSSTags in "%s"`, entryName)
		}
		ossresource.ObjectMetaData.Other.OSSEnvironment = &ex.OSSEnvironment
		if (incl & IncludeOSSValidation) != 0 {
			if ex.OSSValidation == nil {
				if options.GlobalOptions().Lenient {
					debug.Warning(`(Lenient Mode): Cannot include nil OSSValidation in OSS record for "%s"`, entryName)
					ossresource.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(ex.OSSEnvironment.EnvironmentID), options.GlobalOptions().LogTimeStamp)
				} else {
					panic(fmt.Sprintf(`Cannot include nil OSSValidation in OSS record for "%s"`, entryName))
				}
			} else {
				ossresource.ObjectMetaData.Other.OSSValidation = ex.OSSValidation
			}
		}
	case *ossrecord.OSSResourceClassification:
		ossresource.Kind = catalogapi.KindOSSSchema
		ossresource.ID = string(e.GetOSSEntryID())
		ossresource.Name = ossrecord.OSSResourceClassificationEntryName
		ossresource.OverviewUI.En.DisplayName = e.String()
		ossresource.OverviewUI.En.Description = e.String()
		buf := strings.Builder{}
		buf.WriteString(e.String())
		buf.WriteString("\n")
		buf.WriteString("This singleton entry contains metadata describing all supported resource types in the OSS registry")
		buf.WriteString("\n")
		ossresource.OverviewUI.En.LongDescription = buf.String()
		ossresource.ObjectMetaData.Other.OSSResourceClassification = e0
	default:
		panic(fmt.Sprintf("Unknown type of OSSEntry: %#v", e))
	}

	return ossresource, nil
}

// makeOSSEntry allocates and initializes a OSSEntry (OSSServiceExtended, OSSSegment, OSSTribe, OSSEnvironment), based on information found in a (raw) Global Catalog Resource record
func makeOSSEntry(r *catalogapi.Resource, incl IncludeOptions) (result ossrecord.OSSEntry, err error) {
	if options.GlobalOptions().CheckOwner {
		if r.Visibility.Owner != osscatOwnerStaging && r.Visibility.Owner != osscatOwnerProduction {
			err := fmt.Errorf("Ignoring OSS Resource %s found in Global Catalog but not owned by osscat@us.ibm.com - actual owner=%q", r.String(), r.Visibility.Owner)
			return nil, err
		}
	}

	// Allow for errors if the entry is a TEST entry (it will be ignored anyway)
	var rawOSSTags string
	defer func() {
		if err != nil && !options.GlobalOptions().Lenient && !options.GlobalOptions().TestMode && strings.Contains(rawOSSTags, string(osstags.OSSTest)) {
			err = debug.WrapError(err, `Ignoring TEST entry while not in test mode, that contains format errors: %s`, r.String())
			result = nil
		}
	}()

	// Check the presence and absence of the right OSS MetaData fields
	var unexpectedData []string
	var missingData []string
	if r.ObjectMetaData.Other.OSS != nil {
		rawOSSTags = rawOSSTags + r.ObjectMetaData.Other.OSS.GeneralInfo.OSSTags.String()
		if r.Kind != catalogapi.KindOSSService {
			unexpectedData = append(unexpectedData, fmt.Sprintf("OSSService(%s)", r.ObjectMetaData.Other.OSS.String()))
		}
	} else if r.Kind == catalogapi.KindOSSService {
		missingData = append(missingData, "OSSService")
		r.ObjectMetaData.Other.OSS = &ossrecord.OSSService{ReferenceResourceName: "*NONAME*"} // just to allow error recovery
	}
	if r.ObjectMetaData.Other.OSSSegment != nil {
		rawOSSTags = rawOSSTags + r.ObjectMetaData.Other.OSSSegment.OSSTags.String()
		if r.Kind != catalogapi.KindOSSSegment {
			unexpectedData = append(unexpectedData, fmt.Sprintf("%s", r.ObjectMetaData.Other.OSSSegment.String()))
		}
	} else if r.Kind == catalogapi.KindOSSSegment {
		missingData = append(missingData, "OSSSegment")
	}
	if r.ObjectMetaData.Other.OSSTribe != nil {
		rawOSSTags = rawOSSTags + r.ObjectMetaData.Other.OSSTribe.OSSTags.String()
		if r.Kind != catalogapi.KindOSSTribe {
			unexpectedData = append(unexpectedData, fmt.Sprintf("%s", r.ObjectMetaData.Other.OSSTribe.String()))
		}
	} else if r.Kind == catalogapi.KindOSSTribe {
		missingData = append(missingData, "OSSTribe")
	}
	if r.ObjectMetaData.Other.OSSEnvironment != nil {
		rawOSSTags = rawOSSTags + r.ObjectMetaData.Other.OSSEnvironment.OSSTags.String()
		if r.Kind != catalogapi.KindOSSEnvironment {
			unexpectedData = append(unexpectedData, fmt.Sprintf("%s", r.ObjectMetaData.Other.OSSEnvironment.String()))
		}
	} else if r.Kind == catalogapi.KindOSSEnvironment {
		missingData = append(missingData, "OSSEnvironment")
	}
	if r.ObjectMetaData.Other.OSSResourceClassification != nil {
		if r.Kind != catalogapi.KindOSSSchema {
			unexpectedData = append(unexpectedData, fmt.Sprintf("%s", r.ObjectMetaData.Other.OSSResourceClassification.String()))
		}
	} else if r.Kind == catalogapi.KindOSSSchema {
		missingData = append(missingData, "OSSResourceClassification")
	}
	if r.ObjectMetaData.Other.OSSMergeControl != nil {
		if /* (incl&IncludeOSSMergeControl) == 0 || */ r.Kind != catalogapi.KindOSSService {
			unexpectedData = append(unexpectedData, fmt.Sprintf("OSSMergeControl(%s)", r.ObjectMetaData.Other.OSSMergeControl.CanonicalName))
		}
	} else if (incl&IncludeOSSMergeControl) != 0 && r.Kind == catalogapi.KindOSSService {
		if options.GlobalOptions().Lenient {
			debug.Warning("(Lenient Mode): OSS Resource %s found in Global Catalog but does not contain a OSSMergeControl record", r.String())
			r.ObjectMetaData.Other.OSSMergeControl = ossmergecontrol.New(string(r.ObjectMetaData.Other.OSS.ReferenceResourceName))
		} else if r.ObjectMetaData.Other.OSS.GeneralInfo.OSSOnboardingPhase != "" { // XXX FIXME: temporary workaround for RMC/GCOR bug https://github.ibm.com/cloud-sre/operations-ui/issues/478
			debug.Warning("Ignoring missing OSSMergeControl for RMC-managed entry: OSS Resource %s found in Global Catalog but does not contain a OSSMergeControl record", r.String())
			r.ObjectMetaData.Other.OSSMergeControl = ossmergecontrol.New(string(r.ObjectMetaData.Other.OSS.ReferenceResourceName))
		} else {
			missingData = append(missingData, "OSSMergeControl")
		}
	}
	if r.ObjectMetaData.Other.OSSValidation != nil {
		if /* (incl&IncludeOSSValidation) == 0 || */ r.Kind != catalogapi.KindOSSService && r.Kind != catalogapi.KindOSSEnvironment && r.Kind != catalogapi.KindOSSSegment && r.Kind != catalogapi.KindOSSTribe {
			unexpectedData = append(unexpectedData, fmt.Sprintf("OSSValidation(%s)", r.ObjectMetaData.Other.OSSValidation.CanonicalName))
		}
	} else if (incl & IncludeOSSValidation) != 0 {
		if r.Kind == catalogapi.KindOSSSegment || r.Kind == catalogapi.KindOSSTribe {
			debug.Info("Missing optional OSSValidation for %s", r.String())
			r.ObjectMetaData.Other.OSSValidation = nil
		} else if r.Kind == catalogapi.KindOSSSchema {
			r.ObjectMetaData.Other.OSSValidation = nil
		} else {
			if options.GlobalOptions().Lenient {
				debug.Warning("(Lenient Mode): OSS Resource %s found in Global Catalog but does not contain a OSSValidation record", r.String())
				switch r.Kind {
				case catalogapi.KindOSSService:
					r.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(r.ObjectMetaData.Other.OSS.ReferenceResourceName), "")
				case catalogapi.KindOSSSegment:
					r.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(r.ObjectMetaData.Other.OSSSegment.SegmentID), "")
				case catalogapi.KindOSSTribe:
					r.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(r.ObjectMetaData.Other.OSSTribe.TribeID), "")
				case catalogapi.KindOSSEnvironment:
					r.ObjectMetaData.Other.OSSValidation = ossvalidation.New(string(r.ObjectMetaData.Other.OSSEnvironment.EnvironmentID), "")
				default:
					r.ObjectMetaData.Other.OSSValidation = ossvalidation.New("", "")
				}
			} else {
				missingData = append(missingData, "OSSValidation")
			}
		}
	}
	if len(unexpectedData) > 0 || len(missingData) > 0 {
		err := fmt.Errorf("OSS Resource %s found in Global Catalog has unexpected metadata fields: %v or is missing some necessary metadata fields: %v", r.String(), unexpectedData, missingData)
		return nil, err
	}

	// Create the actual OSS entry, from the data in the Catalog Resource
	var expectedName string
	switch r.Kind {
	case catalogapi.KindOSSService:
		if incl&IncludeServices != 0 {
			if incl&(IncludeOSSMergeControl|IncludeOSSValidation|IncludeOSSTimestamps) != 0 {
				ossrec := &ossrecordextended.OSSServiceExtended{}
				result = ossrec
				ossrec.Created = r.Created
				ossrec.Updated = r.Updated
				ossrec.OSSService = *r.ObjectMetaData.Other.OSS
				// Convert OperationalStatus for backward compatibility with LIMITEDAVAILABILITY
				// - issue https://github.ibm.com/cloud-sre/osscatalog/issues/375
				if st, err := ossrecord.ParseOperationalStatus(string(ossrec.OSSService.GeneralInfo.OperationalStatus)); err == nil {
					ossrec.OSSService.GeneralInfo.OperationalStatus = st
				} else {
					if options.GlobalOptions().Lenient {
						debug.Warning("(Lenient Mode): OSS Resource %s found in Global Catalog has invalid OperationalStatus attribute: %s", r.String(), ossrec.OSSService.GeneralInfo.OperationalStatus)
						ossrec.OSSService.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
					} else {
						err2 := debug.WrapError(err, "OSS Resource %s found in Global Catalog has invalid OperationalStatus attribute: %s", r.String(), ossrec.OSSService.GeneralInfo.OperationalStatus)
						return nil, err2
					}
				}
				if (incl & IncludeOSSMergeControl) != 0 {
					ossrec.OSSMergeControl = r.ObjectMetaData.Other.OSSMergeControl
				}
				if (incl & IncludeOSSValidation) != 0 {
					ossrec.OSSValidation = r.ObjectMetaData.Other.OSSValidation
				}
				expectedName = string(makeOSSServiceName(&ossrec.OSSService))
				err := ossrec.CheckConsistency(false)
				if err != nil {
					if options.GlobalOptions().Lenient {
						debug.Warning(`(Lenient Mode): Allowing OSSServiceExtended with consistency error: %v"`, err)
					} else {
						return nil, err
					}
				}
			} else {
				oss := r.ObjectMetaData.Other.OSS
				// Convert OperationalStatus for backward compatibility with LIMITEDAVAILABILITY
				// - issue https://github.ibm.com/cloud-sre/osscatalog/issues/375
				if st, err := ossrecord.ParseOperationalStatus(string(oss.GeneralInfo.OperationalStatus)); err == nil {
					oss.GeneralInfo.OperationalStatus = st
				} else {
					if options.GlobalOptions().Lenient {
						debug.Warning("(Lenient Mode): OSS Resource %s found in Global Catalog has invalid OperationalStatus attribute: %s", r.String(), oss.GeneralInfo.OperationalStatus)
						oss.GeneralInfo.OperationalStatus = ossrecord.OperationalStatusUnknown
					} else {
						err2 := debug.WrapError(err, "OSS Resource %s found in Global Catalog has invalid OperationalStatus attribute: %s", r.String(), oss.GeneralInfo.OperationalStatus)
						return nil, err2
					}
				}
				result = oss
				expectedName = string(makeOSSServiceName(oss))
			}
		} else {
			return nil, nil
		}
	case catalogapi.KindOSSSegment:
		if incl&IncludeTribes != 0 {
			if incl&(IncludeOSSMergeControl|IncludeOSSValidation|IncludeOSSTimestamps) != 0 {
				ossrec := &ossrecordextended.OSSSegmentExtended{}
				result = ossrec
				ossrec.Created = r.Created
				ossrec.Updated = r.Updated
				ossrec.OSSSegment = *r.ObjectMetaData.Other.OSSSegment
				if (incl & IncludeOSSValidation) != 0 {
					ossrec.OSSValidation = r.ObjectMetaData.Other.OSSValidation
				}
				expectedName = string(makeOSSSegmentName(&ossrec.OSSSegment))
				err := ossrec.CheckConsistency(false)
				if err != nil {
					if options.GlobalOptions().Lenient {
						debug.Warning(`(Lenient Mode): Allowing OSSSegmentExtended with consistency error: %v"`, err)
					} else {
						return nil, err
					}
				}
			} else {
				seg := r.ObjectMetaData.Other.OSSSegment
				result = seg
				expectedName = makeOSSSegmentName(seg)
			}
		} else {
			return nil, nil
		}
	case catalogapi.KindOSSTribe:
		if incl&IncludeTribes != 0 {
			if incl&(IncludeOSSMergeControl|IncludeOSSValidation|IncludeOSSTimestamps) != 0 {
				ossrec := &ossrecordextended.OSSTribeExtended{}
				result = ossrec
				ossrec.Created = r.Created
				ossrec.Updated = r.Updated
				ossrec.OSSTribe = *r.ObjectMetaData.Other.OSSTribe
				if (incl & IncludeOSSValidation) != 0 {
					ossrec.OSSValidation = r.ObjectMetaData.Other.OSSValidation
				}
				expectedName = string(makeOSSTribeName(&ossrec.OSSTribe))
				err := ossrec.CheckConsistency(false)
				if err != nil {
					if options.GlobalOptions().Lenient {
						debug.Warning(`(Lenient Mode): Allowing OSSTribeExtended with consistency error: %v"`, err)
					} else {
						return nil, err
					}
				}
			} else {
				tr := r.ObjectMetaData.Other.OSSTribe
				result = tr
				expectedName = makeOSSTribeName(tr)
			}
		} else {
			return nil, nil
		}
	case catalogapi.KindOSSEnvironment:
		if incl&(IncludeEnvironments|IncludeEnvironmentsNative) != 0 {
			if incl&(IncludeOSSMergeControl|IncludeOSSValidation|IncludeOSSTimestamps) != 0 {
				ossrec := &ossrecordextended.OSSEnvironmentExtended{}
				result = ossrec
				ossrec.Created = r.Created
				ossrec.Updated = r.Updated
				ossrec.OSSEnvironment = *r.ObjectMetaData.Other.OSSEnvironment
				if (incl & IncludeOSSValidation) != 0 {
					ossrec.OSSValidation = r.ObjectMetaData.Other.OSSValidation
				}
				expectedName = makeOSSEnvironmentName(&ossrec.OSSEnvironment)
				err := ossrec.CheckConsistency(false)
				if err != nil {
					if options.GlobalOptions().Lenient {
						debug.Warning(`(Lenient Mode): Allowing OSSEnvironmentExtended with consistency error: %v"`, err)
					} else {
						return nil, err
					}
				}
			} else {
				env := r.ObjectMetaData.Other.OSSEnvironment
				result = env
				expectedName = makeOSSEnvironmentName(env)
			}
		} else {
			return nil, nil
		}
	case catalogapi.KindOSSSchema:
		if incl&IncludeOSSResourceClassification != 0 {
			rc := r.ObjectMetaData.Other.OSSResourceClassification
			result = rc
			expectedName = ossrecord.OSSResourceClassificationEntryName
		} else {
			return nil, nil
		}
	default:
		err := fmt.Errorf("Unknown Kind found for OSS Resource %s in Global Catalog", r.String())
		return nil, err
	}

	// Final consistency checks
	err = result.CheckSchemaVersion()
	if err != nil {
		if options.GlobalOptions().Lenient {
			debug.Warning(`(Lenient Mode): Allowing READ with invalid entry schema version: %v"`, err)
		} else {
			return nil, err
		}
	}
	expectedID := string(result.GetOSSEntryID())
	if (r.Name != expectedName) || (r.ID != expectedID) {
		if options.GlobalOptions().Lenient {
			debug.Warning("(Lenient Mode): Unexpected Kind/Name/ID for OSS Resource %s found in Global Catalog; expected %s", r.String(), (&catalogapi.Resource{Name: expectedName, Kind: r.Kind, ID: expectedID}).String())
		} else {
			err := fmt.Errorf("Unexpected Kind/Name/ID for OSS Resource %s found in Global Catalog; expected %s", r.String(), (&catalogapi.Resource{Name: expectedName, Kind: r.Kind, ID: expectedID}).String())
			return nil, err
		}
	}
	return result, nil
}

// computeIncludeOptions computes the "include" option string for loading entries from the Global Catalog
// (must explicity include the OSS-related records in the metadata.other section)
func computeIncludeOptions(incl IncludeOptions) string {
	result := strings.Builder{}
	result.WriteString("metadata.other.oss:metadata.other.oss_segment:metadata.other.oss_tribe:metadata.other.oss_environment")
	if (incl & IncludeOSSMergeControl) != 0 {
		result.WriteString(":metadata.other.oss_merge_control")
	}
	if (incl & IncludeOSSValidation) != 0 {
		result.WriteString(":metadata.other.oss_validation_info")
	}
	if (incl & IncludeEnvironmentsNative) != 0 {
		result.WriteString(":metadata.deployment")
	}
	return result.String()
}

/*
// readOSSCatalogEntryByName reads an OSS entry from the Global Catalog, given its entry name
func readOSSCatalogResourceByName(ossname string, incl IncludeOptions) (ossresource *catalogapi.Resource, err error) {
	actualURL := fmt.Sprintf("%s?q=name:%s&include=%s", catalogOSSURL, ossname, computeIncludeOptions(incl))
	var result = catalogapi.GetResponse{}
	token, err := rest.GetToken(catalogOSSKeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get IAM token for Global Catalog (OSS entries)")
		return nil, err
	}
	err = rest.DoHTTPGet(actualURL, token, nil, "Catalog", debug.Catalog, &result)
	if err != nil {
		return nil, err
	}

	count := int(result.ResourceCount)
	if count == 0 {
		err := rest.MakeHTTPError(nil, true, "Resource \"%s\" not found in Global Catalog (OSS entries)", ossname)
		return nil, err
	} else if count > 1 {
		panic(fmt.Sprintf("Found %d entries in Global Catalog for the same name \"%s\" (OSS entries)", count, ossname))
	}
	ossresource = &result.Resources[0]
	return ossresource, nil
}
*/

// ReadOSSEntryByID reads an OSS entry from the Global Catalog, given its ID,
// in the Staging instance (which contains the most recently updated data)
func ReadOSSEntryByID(id ossrecord.OSSEntryID, incl IncludeOptions) (entry ossrecord.OSSEntry, err error) {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return nil, err
	}
	return ReadOSSEntryByIDWithContext(ctx, id, incl)
}

// ReadOSSEntryByIDProduction reads an OSS entry from the Global Catalog, given its ID,
// in the Production instance (which contains the stable data)
func ReadOSSEntryByIDProduction(id ossrecord.OSSEntryID, incl IncludeOptions) (entry ossrecord.OSSEntry, err error) {
	ctx, err := setupContextForOSSEntries(productionFlagReadOnly)
	if err != nil {
		return nil, err
	}
	return ReadOSSEntryByIDWithContext(ctx, id, incl)
}

// ReadOSSEntryByIDWithContext reads an OSS entry from the Global Catalog, given its ID.
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func ReadOSSEntryByIDWithContext(ctx context.Context, id ossrecord.OSSEntryID, incl IncludeOptions) (entry ossrecord.OSSEntry, err error) {
	incl.normalize()
	if (incl&IncludeEnvironmentsNative != 0) && ossrecord.IsOSSEnvironmentID(id) {
		panic(fmt.Sprintf("ReadOSSEntryByIDWithContext() not supported for OSSEnvironment entries with IncludeEnvironmentsNative flag (id=%s", id))
	}
	catalogOSSURL, _, token, err := readContextForOSSEntries(ctx, false)
	if err != nil {
		return nil, err
	}
	actualURL := fmt.Sprintf("%s/%s?include=%s", catalogOSSURL, id, computeIncludeOptions(incl))
	var result = catalogapi.Resource{}
	err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.OSS", debug.Catalog, &result)
	if err != nil {
		parentIDMap.UnSet(catalogOSSURL, string(id))
		if httpErr, ok := err.(rest.HTTPError); ok && httpErr.GetHTTPStatusCode() == http.StatusNotFound {
			err1 := rest.MakeHTTPError(err, nil, true, "Catalog OSSEntry not found - id=%s", id)
			return nil, err1
		}
		return nil, err
	}

	ossresource := &result

	// TODO: need to be able to read CatalogNative OSSEnvironment entries

	// In the case of a service, get the overrrides as well so we can convert the record to domain specific below if needed:
	entry, err = makeOSSEntry(ossresource, incl|IncludeServicesDomainOverrides)
	if err != nil {
		parentIDMap.UnSet(catalogOSSURL, ossresource.ID)
		return nil, err
	}

	// Convert the OSS entry based on the domain includes if applicable:
	entry, err = getOSSEntryForDomain(&entry, incl)
	if err != nil {
		return nil, err
	}

	if ossresource.Visibility.Restrictions != string(catalogapi.VisibilityIBMOnly) {
		// XXX Is the visibility always expected to be IBM? What if changed by a user?
		debug.PrintError(`ReadOSSCatalogEntry(%s) returned unexpected Visibility: "%v" (expected "ibm_only")`, ossresource.String(), ossresource.Visibility.Restrictions)
	}

	parentIDMap.Set(catalogOSSURL, ossresource.ID, ossresource.ParentID)

	return entry, nil
}

// ReadOSSService reads an OSSService entry from the Global Catalog, given its name,
// in the Staging instance (which contains the most recently updated data)
func ReadOSSService(name ossrecord.CRNServiceName) (oss *ossrecord.OSSService, err error) {
	return ReadOSSServiceWithOptions(name, IncludeServices)
}

// ReadOSSServiceWithOptions reads an OSSService entry from the Global Catalog, given its name
// and provided options, in the Staging instance (which contains the most recently updated data)
func ReadOSSServiceWithOptions(name ossrecord.CRNServiceName, incl IncludeOptions) (oss *ossrecord.OSSService, err error) {
	id2 := ossrecord.MakeOSSServiceID(name)
	e, err := ReadOSSEntryByID(id2, incl|IncludeServices)
	if err != nil {
		return nil, err
	}
	var ok bool
	if oss, ok = e.(*ossrecord.OSSService); !ok {
		err := fmt.Errorf("ReadOSSService(%s) returned Catalog Resource of unexpected type %T (%v)", name, e, e)
		return nil, err
	}
	if oss.ReferenceResourceName != name {
		err := fmt.Errorf("ReadOSSService(%s) returned entry with unexpected name %s", name, oss.ReferenceResourceName)
		return nil, err
	}
	return oss, nil
}

// ReadOSSServiceProduction reads an OSSService entry from the Global Catalog, given its name,
// in the Production instance (which contains the stable data)
func ReadOSSServiceProduction(name ossrecord.CRNServiceName) (oss *ossrecord.OSSService, err error) {
	return ReadOSSServiceProductionWithOptions(name, IncludeServices)
}

// ReadOSSServiceProductionWithOptions reads an OSSService entry from the Global Catalog, given its name
// and provided options, in the Production instance (which contains the stable data)
func ReadOSSServiceProductionWithOptions(name ossrecord.CRNServiceName, incl IncludeOptions) (oss *ossrecord.OSSService, err error) {
	id2 := ossrecord.MakeOSSServiceID(name)
	e, err := ReadOSSEntryByIDProduction(id2, incl|IncludeServices)
	if err != nil {
		return nil, err
	}
	var ok bool
	if oss, ok = e.(*ossrecord.OSSService); !ok {
		err := fmt.Errorf("ReadOSSServiceProduction(%s) returned Catalog Resource of unexpected type %T (%v)", name, e, e)
		return nil, err
	}
	if oss.ReferenceResourceName != name {
		err := fmt.Errorf("ReadOSSServiceProduction(%s) returned entry with unexpected id %s", name, oss.ReferenceResourceName)
		return nil, err
	}
	return oss, nil
}

// ReadOSSSegment reads an OSSSegment entry from the Global Catalog, given its SegmentID,
// in the Staging instance (which contains the most recently updated data)
func ReadOSSSegment(id ossrecord.SegmentID) (seg *ossrecord.OSSSegment, err error) {
	id2 := ossrecord.MakeOSSSegmentID(id)
	e, err := ReadOSSEntryByID(id2, IncludeTribes)
	if err != nil {
		return nil, err
	}
	var ok bool
	if seg, ok = e.(*ossrecord.OSSSegment); !ok {
		err := fmt.Errorf("ReadOSSSegment(%s) returned Catalog Resource of unexpected type %T (%v)", id, e, e)
		return nil, err
	}
	if seg.SegmentID != id {
		err := fmt.Errorf("ReadOSSSegment(%s) returned entry with unexpected id %s", id, seg.SegmentID)
		return nil, err
	}
	return seg, nil
}

// ReadOSSSegmentProduction reads an OSSSegment entry from the Global Catalog, given its SegmentID,
// in the Production instance (which contains the stable data)
func ReadOSSSegmentProduction(id ossrecord.SegmentID) (seg *ossrecord.OSSSegment, err error) {
	id2 := ossrecord.MakeOSSSegmentID(id)
	e, err := ReadOSSEntryByIDProduction(id2, IncludeTribes)
	if err != nil {
		return nil, err
	}
	var ok bool
	if seg, ok = e.(*ossrecord.OSSSegment); !ok {
		err := fmt.Errorf("ReadOSSSegmentProduction(%s) returned Catalog Resource of unexpected type %T (%v)", id, e, e)
		return nil, err
	}
	if seg.SegmentID != id {
		err := fmt.Errorf("ReadOSSSegmentProduction(%s) returned entry with unexpected id %s", id, seg.SegmentID)
		return nil, err
	}
	return seg, nil
}

// ReadOSSTribe reads an OSSTribe entry from the Global Catalog, given its TribeID,
// in the Staging instance (which contains the most recently updated data)
func ReadOSSTribe(id ossrecord.TribeID) (tr *ossrecord.OSSTribe, err error) {
	id2 := ossrecord.MakeOSSTribeID(id)
	e, err := ReadOSSEntryByID(id2, IncludeTribes)
	if err != nil {
		return nil, err
	}
	var ok bool
	if tr, ok = e.(*ossrecord.OSSTribe); !ok {
		err := fmt.Errorf("ReadOSSTribe(%s) returned Catalog Resource of unexpected type %T (%v)", id, e, e)
		return nil, err
	}
	if tr.TribeID != id {
		err := fmt.Errorf("ReadOSSTribe(%s) returned entry with unexpected id %s", id, tr.TribeID)
		return nil, err
	}
	return tr, nil
}

// ReadOSSTribeProduction reads an OSSTribe entry from the Global Catalog, given its TribeID,
// in the Production instance (which contains the stable data)
func ReadOSSTribeProduction(id ossrecord.TribeID) (tr *ossrecord.OSSTribe, err error) {
	id2 := ossrecord.MakeOSSTribeID(id)
	e, err := ReadOSSEntryByIDProduction(id2, IncludeTribes)
	if err != nil {
		return nil, err
	}
	var ok bool
	if tr, ok = e.(*ossrecord.OSSTribe); !ok {
		err := fmt.Errorf("ReadOSSTribeProduction(%s) returned Catalog Resource of unexpected type %T (%v)", id, e, e)
		return nil, err
	}
	if tr.TribeID != id {
		err := fmt.Errorf("ReadOSSTribeProduction(%s) returned entry with unexpected id %s", id, tr.TribeID)
		return nil, err
	}
	return tr, nil
}

// CreateOSSEntry creates a new OSS entry in the Global Catalog
// in the Staging instance (which contains the most recently updated data)
func CreateOSSEntry(e ossrecord.OSSEntry, incl IncludeOptions) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	return CreateOSSEntryWithContext(ctx, e, incl)
}

// CreateOSSEntryProduction creates a new OSS entry in the Global Catalog
// in the Production instance (which contains the stable data)
func CreateOSSEntryProduction(e ossrecord.OSSEntry, incl IncludeOptions) error {
	ctx, err := setupContextForOSSEntries(productionFlagReadWrite)
	if err != nil {
		return err
	}
	return CreateOSSEntryWithContext(ctx, e, incl)
}

// CreateOSSEntryWithContext creates a new OSS entry in the Global Catalog
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func CreateOSSEntryWithContext(ctx context.Context, e ossrecord.OSSEntry, incl IncludeOptions) error {
	incl.normalize()
	catalogOSSURL, _, token, err := readContextForOSSEntries(ctx, true)
	if err != nil {
		return err
	}
	ossresource, err := makeOSSCatalogResource(e, incl)
	if err != nil {
		return err
	}
	actualURL := fmt.Sprintf("%s", catalogOSSURL)

	ossresourceOut := &catalogapi.Resource{}

	err = rest.DoHTTPPostOrPut("POST", actualURL, string(token), nil, ossresource, ossresourceOut, "Catalog.OSS.Write", debug.CatalogWrite)
	if err != nil {
		parentIDMap.UnSet(catalogOSSURL, ossresource.ID)
		return err
	}

	visibility := catalogapi.Visibility{Restrictions: string(catalogapi.VisibilityIBMOnly)}
	err = SetOSSVisibilityWithContext(ctx, ossrecord.CatalogID(ossresource.ID), &visibility)

	parentIDMap.Set(catalogOSSURL, ossresource.ID, ossresource.ParentID)

	e.SetTimes(ossresourceOut.Created, ossresourceOut.Updated)

	return err
}

// UpdateOSSEntry updates an existing OSS entry in the Global Catalog
// in the Staging instance (which contains the most recently updated data)
func UpdateOSSEntry(e ossrecord.OSSEntry, incl IncludeOptions) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	return UpdateOSSEntryWithContext(ctx, e, incl)
}

// UpdateOSSEntryProduction updates an existing OSS entry in the Global Catalog
// in the Production instance (which contains the stable data)
func UpdateOSSEntryProduction(e ossrecord.OSSEntry, incl IncludeOptions) error {
	ctx, err := setupContextForOSSEntries(productionFlagReadWrite)
	if err != nil {
		return err
	}
	return UpdateOSSEntryWithContext(ctx, e, incl)
}

// UpdateOSSEntryWithContext updates an existing OSS entry in the Global Catalog
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func UpdateOSSEntryWithContext(ctx context.Context, e ossrecord.OSSEntry, incl IncludeOptions) error {
	incl.normalize()
	catalogOSSURL, _, token, err := readContextForOSSEntries(ctx, true)
	if err != nil {
		return err
	}
	ossresource, err := makeOSSCatalogResource(e, incl)
	if err != nil {
		return err
	}
	var actualURL string
	if previousParent, found := parentIDMap.Lookup(catalogOSSURL, ossresource.ID); found {
		if previousParent == ossresource.ParentID {
			actualURL = fmt.Sprintf("%s/%s", catalogOSSURL, ossresource.ID)
		} else {
			actualURL = fmt.Sprintf("%s/%s?move=true", catalogOSSURL, ossresource.ID)
		}
	} else {
		return fmt.Errorf("Attempting to update OSS Entry that was not previously loaded: %s   %s", e.String(), ossresource.String())
	}

	ossresourceOut := &catalogapi.Resource{}

	err = rest.DoHTTPPostOrPut("PUT", actualURL, string(token), nil, ossresource, ossresourceOut, "Catalog.OSS.Write", debug.CatalogWrite)
	if err != nil {
		parentIDMap.UnSet(catalogOSSURL, ossresource.ID)
		return err
	}

	parentIDMap.Set(catalogOSSURL, ossresource.ID, ossresource.ParentID)

	if ossresourceOut.Visibility.Restrictions != string(catalogapi.VisibilityIBMOnly) {
		// XXX Should we really be modifying the visibility on an update?
		debug.PrintError("UpdateOSSCatalogEntry(%s) returned unexpected Visibility: %v - attempting to fix", ossresource.String(), ossresourceOut.Visibility.Restrictions)
		visibility := catalogapi.Visibility{Restrictions: string(catalogapi.VisibilityIBMOnly)}
		err = SetOSSVisibilityWithContext(ctx, ossrecord.CatalogID(ossresource.ID), &visibility)
	}

	e.SetTimes(ossresourceOut.Created, ossresourceOut.Updated)

	return err
}

// DeleteOSSEntryByIDWithContext deletes an existing OSS entry in the Global Catalog
// Note: this call succeeds even if the entry did not exist
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func DeleteOSSEntryByIDWithContext(ctx context.Context, id ossrecord.CatalogID) error {
	catalogOSSURL, _, token, err := readContextForOSSEntries(ctx, true)
	if err != nil {
		return err
	}
	actualURL := fmt.Sprintf("%s/%s", catalogOSSURL, id)

	err = rest.DoHTTPPostOrPut("DELETE", actualURL, string(token), nil, nil, nil, "Catalog.OSS.Write", debug.CatalogWrite)

	if err != nil {
		parentIDMap.UnSet(catalogOSSURL, string(id))
		if httpErr, ok := err.(rest.HTTPError); ok && httpErr.GetHTTPStatusCode() == http.StatusNotFound {
			err1 := rest.MakeHTTPError(err, nil, true, "Catalog OSSEntry not found - id=%s", id)
			debug.Debug(debug.Catalog, "Ignoring error while deleting Catalog OSS entry: %v", err1)
			return nil
		}
	}
	parentIDMap.UnSet(catalogOSSURL, string(id))
	return err
}

// DeleteOSSService deletes an OSS service/component record from the Global Catalog, given its name,
// in the Staging instance (which contains the most recently updated data)
// Note: this call succeeds even if the entry did not exist
func DeleteOSSService(name ossrecord.CRNServiceName) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	catalogID := ossrecord.CatalogID(ossrecord.MakeOSSServiceID(name))
	return DeleteOSSEntryByIDWithContext(ctx, catalogID)
}

// DeleteOSSSegment deletes an OSSSegment entry from the Global Catalog, given its name,
// in the Staging instance (which contains the most recently updated data)
// Note: this call succeeds even if the entry did not exist
func DeleteOSSSegment(segmentID ossrecord.SegmentID) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	catalogID := ossrecord.CatalogID(ossrecord.MakeOSSSegmentID(segmentID))
	return DeleteOSSEntryByIDWithContext(ctx, catalogID)
}

// DeleteOSSTribe deletes an OSSTribe entry from the Global Catalog, given its name,
// in the Staging instance (which contains the most recently updated data)
// Note: this call succeeds even if the entry did not exist
func DeleteOSSTribe(tribeID ossrecord.TribeID) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	catalogID := ossrecord.CatalogID(ossrecord.MakeOSSTribeID(tribeID))
	return DeleteOSSEntryByIDWithContext(ctx, catalogID)
}

// DeleteOSSEntry deletes an OSS from the Global Catalog
// in the Staging instance (which contains the most recently updated data)
// Note: this call succeeds even if the entry did not exist
func DeleteOSSEntry(e ossrecord.OSSEntry) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	catalogID := ossrecord.CatalogID(e.GetOSSEntryID())
	return DeleteOSSEntryByIDWithContext(ctx, catalogID)
}

// DeleteOSSEntryProduction deletes an OSS from the Global Catalog
// in the Production instance (which contains the stable data)
// Note: this call succeeds even if the entry did not exist
func DeleteOSSEntryProduction(e ossrecord.OSSEntry) error {
	ctx, err := setupContextForOSSEntries(productionFlagReadWrite)
	if err != nil {
		return err
	}
	catalogID := ossrecord.CatalogID(e.GetOSSEntryID())
	return DeleteOSSEntryByIDWithContext(ctx, catalogID)
}

// ReadOSSVisibility reads the Visibility information for one OSS Catalog entry
// in the Staging instance (which contains the most recently updated data)
func ReadOSSVisibility(id ossrecord.CatalogID) (*catalogapi.Visibility, error) {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return nil, err
	}
	return ReadOSSVisibilityWithContext(ctx, id)
}

// ReadOSSVisibilityProduction reads the Visibility information for one OSS Catalog entry
// in the Production instance (which contains the stable data)
func ReadOSSVisibilityProduction(id ossrecord.CatalogID) (*catalogapi.Visibility, error) {
	ctx, err := setupContextForOSSEntries(productionFlagReadOnly)
	if err != nil {
		return nil, err
	}
	return ReadOSSVisibilityWithContext(ctx, id)
}

// ReadOSSVisibilityWithContext reads the Visibility information for one OSS Catalog entry
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func ReadOSSVisibilityWithContext(ctx context.Context, id ossrecord.CatalogID) (*catalogapi.Visibility, error) {
	catalogOSSURL, _, token, err := readContextForOSSEntries(ctx, false)
	if err != nil {
		return nil, err
	}
	actualURL := fmt.Sprintf("%s/%s/visibility", catalogOSSURL, id)
	var result = catalogapi.Visibility{}
	err = rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.OSS", debug.Catalog, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SetOSSVisibility sets the Visibility information for one OSS Catalog entry
// in the Staging instance (which contains the most recently updated data)
func SetOSSVisibility(id ossrecord.CatalogID, visibility *catalogapi.Visibility) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	return SetOSSVisibilityWithContext(ctx, id, visibility)
}

// SetOSSVisibilityProduction sets the Visibility information for one OSS Catalog entry
// in the Production instance (which contains the stable data)
func SetOSSVisibilityProduction(id ossrecord.CatalogID, visibility *catalogapi.Visibility) error {
	ctx, err := setupContextForOSSEntries(productionFlagReadWrite)
	if err != nil {
		return err
	}
	return SetOSSVisibilityWithContext(ctx, id, visibility)
}

// SetOSSVisibilityWithContext sets the Visibility information for one OSS Catalog entry
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func SetOSSVisibilityWithContext(ctx context.Context, id ossrecord.CatalogID, visibility *catalogapi.Visibility) error {
	catalogOSSURL, _, token, err := readContextForOSSEntries(ctx, true)
	if err != nil {
		return err
	}
	actualURL := fmt.Sprintf("%s/%s/visibility", catalogOSSURL, id)
	err = rest.DoHTTPPostOrPut("PUT", actualURL, string(token), nil, visibility, nil, "Catalog.OSS.Write", debug.CatalogWrite)

	// TODO: Should not need a retry for SetOSSVisibility
	if err != nil {
		delay := time.Duration(1)
		debug.PrintError("*** Retrying SetOSSVisibility(%s) after %d second delay (previous error: %v)\n", id, delay, err)
		time.Sleep(delay * time.Second)
		err = rest.DoHTTPPostOrPut("PUT", actualURL, string(token), nil, visibility, nil, "Catalog.OSS.Write", debug.CatalogWrite)
	}

	return err
}

//CountOSSEntries keeps track of all the counts of OSS entries encountered during ListOSSEntries() and its variants
type CountOSSEntries struct {
	rawEntries         int
	segments           int
	tribes             int
	environments       int
	services           int
	nativeEnvironments int
	schema             int
}

func newCountOSSEntries() *CountOSSEntries {
	ret := &CountOSSEntries{}
	return ret
}

// Total returns the total number of entries returned from ListOSSEntries() and its variants (after applying the filter pattern)
func (c *CountOSSEntries) Total() int {
	return c.segments + c.tribes + c.environments + c.services + c.nativeEnvironments + c.schema
}

// String returns a string representation of all the counts of OSS entries encountered during ListOSSEntries() and its variants
func (c *CountOSSEntries) String() string {
	if c.nativeEnvironments > 0 {
		return fmt.Sprintf("{ rawEntries=%d. segments=%d, tribes=%d, environments=%d, nativeEnvironments=%d, services=%d, schema=%d }", c.rawEntries, c.segments, c.tribes, c.environments, c.nativeEnvironments, c.services, c.schema)
	}
	return fmt.Sprintf("{ rawEntries=%d, segments=%d, tribes=%d, environments=%d, services=%d, schema=%d }", c.rawEntries, c.segments, c.tribes, c.environments, c.services, c.schema)
}

// ListOSSEntries lists all OSS entries of all types in Global Catalog (service/components, segments, tribes)
// from the Staging instance (which contains the most recently updated data)
func ListOSSEntries(pattern *regexp.Regexp, incl IncludeOptions, handler func(r ossrecord.OSSEntry)) error {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return err
	}
	return ListOSSEntriesWithContext(ctx, pattern, incl, handler)
}

// ListOSSEntriesProduction lists all OSS entries of all types in Global Catalog (service/components, segments, tribes)
// from the Production instance (which contains the stable data)
func ListOSSEntriesProduction(pattern *regexp.Regexp, incl IncludeOptions, handler func(r ossrecord.OSSEntry)) error {
	ctx, err := setupContextForOSSEntries(productionFlagReadOnly)
	if err != nil {
		return err
	}
	return ListOSSEntriesWithContext(ctx, pattern, incl, handler)
}

// ListOSSEntriesWithContext lists all OSS entries of all types in Global Catalog (service/components, segments, tribes).
// The Context parameter specifies whether to access the Production or Staging instance, and the IAM token to use
func ListOSSEntriesWithContext(ctx context.Context, pattern *regexp.Regexp, incl IncludeOptions, handler func(r ossrecord.OSSEntry)) error {
	// TODO: Should we look for OSS entries as childen of other non-OSS entries
	incl.normalize()
	var counts = newCountOSSEntries()
	var offset int
	catalogOSSURL, _, token, err := readContextForOSSEntries(ctx, false)
	if err != nil {
		return err
	}
	catalogName := getCatalogNameFromContext(ctx)
	includeOptionsString := computeIncludeOptions(incl)
	queryParams := make([]string, 0, 10)
	queryParams = append(queryParams, "kind:"+catalogapi.KindOSSSegment) // Always need to check, because Services, Tribes and Environments may be inside it
	queryParams = append(queryParams, "kind:"+catalogapi.KindOSSTribe)   // Should not be at top, but let's check anyway
	if incl&IncludeServices != 0 {
		queryParams = append(queryParams, "kind:"+catalogapi.KindOSSService)
	}
	if incl&IncludeTribes != 0 {
		// Already covered unconditionally
	}
	if incl&IncludeEnvironments != 0 {
		queryParams = append(queryParams, "kind:"+catalogapi.KindOSSEnvironment)
	}
	if incl&IncludeEnvironmentsNative != 0 {
		queryParams = append(queryParams, "kind:"+catalogapi.KindOSSEnvironment)
		queryParams = append(queryParams, "kind:"+catalogapi.KindGeography)
		//		queryParams = append(queryParams, "kind:"+catalogapi.KindLegacyEnvironment)
		//		queryParams = append(queryParams, "kind:"+catalogapi.KindLegacyCName)
	}
	if incl&IncludeOSSResourceClassification != 0 {
		queryParams = append(queryParams, "kind:"+catalogapi.KindOSSSchema)
	}
	queryParamsString := strings.Join(queryParams, "+")

	var count int
	for {
		debug.Info("Loading one batch of OSS entries from %s Global Catalog (%d/%d entries so far - %s) (last_offset=%d  last_count=%d)", catalogName, counts.Total(), counts.rawEntries, counts.String(), offset, count)
		offset += count
		actualURL := fmt.Sprintf("%s?_offset=%d&include=%s&q=%s", catalogOSSURL, offset, includeOptionsString, queryParamsString)
		// actualURL = fmt.Sprintf("%s?_offset=%d&include=%s", catalogOSSURL, offset, includeOptionsString) // override the use of kind= query filters
		var result = new(catalogapi.GetResponse)
		err := rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.OSS", debug.Catalog, result)
		if err != nil {
			return err
		}

		count = int(result.ResourceCount)
		if count != len(result.Resources) {
			panic(fmt.Sprintf("ListOSSEntries(): GET OSS entries from Catalog: Offset=%d  ResourceCount=%d   len(Resources)=%d", offset, count, len(result.Resources)))
		}
		if count == 0 {
			break
		}
		_, err = processOSSEntries(ctx, result.Resources, pattern, incl, "", "", counts, handler)
		if err != nil {
			return err
		}
	}

	debug.Info("Read %d OSS entries from %s Global Catalog: %s (last_offset=%d  last_count=%d)", counts.Total(), catalogName, counts.String(), offset, count)
	if counts.Total() == 0 || counts.rawEntries == 0 {
		return fmt.Errorf("Read 0 OSS entries from %s Global Catalog: %s", catalogName, counts.String())
	}
	return nil
}

// processOSSEntries processes a list of OSS entries received from Global Catalog,
// for the implementation of ListOSSEntriesWithContext()
func processOSSEntries(ctx context.Context, resources []catalogapi.Resource, pattern *regexp.Regexp, incl IncludeOptions, parentID string, path string, counts *CountOSSEntries, handler func(r ossrecord.OSSEntry)) (entryCount int, err error) {
	var token Token
	var catalogOSSURL string
	entryCount = 0
	if len(resources) > 0 {
		counts.rawEntries += len(resources)
		catalogOSSURL, _, token, err = readContextForOSSEntries(ctx, false)
		if err != nil {
			return 0, err
		}
	}
	for i := 0; i < len(resources); i++ {
		entryCount++
		r := &resources[i]
		var lookForChildren bool
		var newParentID string
		var newPath = fmt.Sprintf("%s/%s", path, r.Name)
		if pattern != nil && pattern.FindString(newPath) == "" {
			//			debug.Debug(debug.Catalog, `Skipping entry %s  --> does not match the name pattern`, r.String())
			continue
		}
		switch r.Kind {
		case catalogapi.KindOSSService, catalogapi.KindOSSSegment, catalogapi.KindOSSTribe, catalogapi.KindOSSEnvironment, catalogapi.KindOSSSchema:
			// In the case of a service, get the overrrides as well so we can convert the record to domain specific below if needed:
			entry, err := makeOSSEntry(r, incl|IncludeServicesDomainOverrides)
			if err != nil { //XXX should use golang error wrapping
				if strings.Contains(err.Error(), `found in Global Catalog but not owned by osscat@us.ibm.com`) {
					debug.Warning(`ListOSSEntries(): %v`, err)
				} else if strings.Contains(err.Error(), `Ignoring TEST entry while not in test mode`) {
					debug.Warning(`ListOSSEntries(): %v`, err)
				} else {
					debug.PrintError(`ListOSSEntries(): %v`, err)
				}
			} else if entry != nil {
				// Look at incl options to determine what domains to included in the response
				parentIDMap.Set(catalogOSSURL, r.ID, r.ParentID)
				entries := getDomainSpecificOSSEntries(&entry, incl)
				for _, entryToPassToHandler := range entries {
					handler(entryToPassToHandler)
				}
			}
			switch r.Kind {
			case catalogapi.KindOSSSegment:
				// keep going to look for children
				lookForChildren = true
				counts.segments++
			case catalogapi.KindOSSTribe:
				// keep going to look for children
				lookForChildren = true
				counts.tribes++
			case catalogapi.KindOSSService:
				counts.services++
			case catalogapi.KindOSSEnvironment:
				counts.environments++
			case catalogapi.KindOSSSchema:
				counts.schema++
			}

		case catalogapi.KindRegion, catalogapi.KindDatacenter, catalogapi.KindAvailabilityZone, catalogapi.KindPOP, catalogapi.KindSatellite:
			if incl&IncludeEnvironmentsNative != 0 {
				debug.Debug(debug.Catalog, `Processing entry %s  --> creating catalog_native OSSEnvironment entry`, r.String())
				debug.Debug(debug.Visibility, `Visibility for entry %s: %v`, r.String(), &r.Visibility)
				if r.Visibility.Restrictions != string(catalogapi.VisibilityPublic) {
					catalogName := getCatalogNameFromContext(ctx)
					debug.Info("Ignoring non-public location entry in %s Global Catalog: %s - visibility=%s", catalogName, r.String(), r.Visibility.Restrictions)
					continue
				}
				var entry ossrecord.OSSEntry
				var err error
				entry, newParentID, err = makeNativeOSSEnvironmentEntry(r, parentID, path)
				if err != nil {
					debug.PrintError(`ListOSSEntries(): %v`, err)
				} else {
					handler(entry)
					counts.nativeEnvironments++
				}
				lookForChildren = true
				// keep going to look for children
			}

		case catalogapi.KindGeography, catalogapi.KindMetro, catalogapi.KindCountry:
			if incl&IncludeEnvironmentsNative != 0 {
				debug.Debug(debug.Catalog, `Processing entry %s  --> keep traversing the hierarchy`, r.String())
				if r.Visibility.Restrictions != string(catalogapi.VisibilityPublic) {
					catalogName := getCatalogNameFromContext(ctx)
					debug.Info("Ignoring non-public location entry in %s Global Catalog: %s - visibility=%s", catalogName, r.String(), r.Visibility.Restrictions)
					continue
				}
				lookForChildren = true
				// ignore but keep going to look for children
			}

		default:
			debug.Debug(debug.Catalog, `Processing entry %s  --> unknown kind`, r.String())
			//			debug.PrintError(`ListOSSEnvironmentsWithContext() Found unknown entry kind %s in Global Catalog: %s`, r.Kind, r.String())
		}

		// Note that we traverse the groups even if the group name itself does not match the pattern or the type itself is not included -- some children might match
		// TODO: Should try to be clever and not traverse parts of the tree that cannot contain any entry types not specified in the IncludeOptions
		if r.Group || lookForChildren {
			// TODO: Do we need to iterate over an offset to get children entries
			url := r.ChildrenURL
			if url != "" {
				var offset int
				for {
					actualURL := fmt.Sprintf("%s?_offset=%d&include=%s", url, offset, computeIncludeOptions(incl))
					var result = new(catalogapi.GetResponse)
					err := rest.DoHTTPGet(actualURL, string(token), nil, "Catalog.OSS", debug.Catalog, result)
					if err != nil {
						return entryCount, err
					}
					count := int(result.ResourceCount)
					if count != len(result.Resources) {
						panic(fmt.Sprintf("ListOSSEntries(): GET OSS entries from Catalog: Offset=%d  ResourceCount=%d   len(Resources)=%d", offset, count, len(result.Resources)))
					}
					if count == 0 {
						break
					}
					children, err := processOSSEntries(ctx, result.Resources, pattern, incl, newParentID, newPath, counts, handler)
					if err != nil {
						return entryCount, err
					}
					if r.Group {
						debug.PrintError(`ListOSSEntries() Found unexpected GROUP OSS entry in Global Catalog: %s  with %d valid OSS children entries`, r.String(), children)
					}
					switch r.Kind {
					case catalogapi.KindOSSSegment:
						if children == 0 {
							debug.Warning(`ListOSSEntries() Segment OSS entry with no valid OSS children entries in Global Catalog: %s`, r.String())
						}
					case catalogapi.KindOSSService, catalogapi.KindOSSEnvironment, catalogapi.KindOSSSchema:
						if children > 0 {
							debug.PrintError(`ListOSSEntries() OSS entry with unexpected %d OSS children entries in Global Catalog: %s`, children, r.String())
						}
					case catalogapi.KindOSSTribe:
						// ignore
					}
					offset += count
				}
			} else {
				debug.PrintError(`ListOSSEntries() children_url is nil: %s  group=%v`, r.String(), r.Group)
			}
		}
	}
	return entryCount, nil
}

// GetOSSEntryUI returns a URL for viewing the specified OSS entry in the Global Catalog UI
// in the Staging instance (which contains the most recently updated data)
func GetOSSEntryUI(entry ossrecord.OSSEntry) (string, error) {
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		return "", err
	}
	catalogOSSURL, _, _, err := readContextForOSSEntries(ctx, false)
	if err != nil {
		return "", err
	}
	if catalogOSSURL == "" {
		return "", fmt.Errorf("GetOSSEntryUI(): empty URL from Context")
	}
	url0, err := url.Parse(catalogOSSURL)
	if err != nil {
		return "", err
	}
	id := entry.GetOSSEntryID()
	str := fmt.Sprintf("%s://%s/update/%s", url0.Scheme, url0.Host, url.PathEscape(string(id)))
	return str, nil
}

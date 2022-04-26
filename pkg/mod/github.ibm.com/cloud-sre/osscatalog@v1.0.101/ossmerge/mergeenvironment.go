package ossmerge

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/collections"
	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/doctor"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// Sets of CRN validation keys to use when parsing CRN masks found in Catalog and Doctor entries
var (
	crnFromCatalog = []crn.ValidationKey{crn.ExpectEnvironment, crn.AllowBlankCName, crn.AllowSoftLayerCName, crn.StripOtherSegments}
	crnFromDoctor  = []crn.ValidationKey{crn.ExpectEnvironment, crn.AllowSoftLayerCName}
)

// checkSourceCatalogValid checks that the given SourceMainCatalog entry is valid for this environment
// If it is valid, it returns an empty string; if it is not valid, it return a short string with a reason why
func (env *EnvironmentInfo) checkSourceCatalogValid(cat *catalogapi.Resource) string {
	if cat.Kind == catalogapi.KindLegacyCName || cat.Kind == catalogapi.KindLegacyEnvironment {
		return invalidDedicatedLocal
	}
	if cat.Name != env.ComparableCRNMask.Location || cat.ID != env.ComparableCRNMask.Location {
		return invalidNameIDMismatch
	}
	return ""
}

const (
	invalidDedicatedLocal = "dedicated/local"
	invalidNameIDMismatch = "name/id/location mismatch"
)

// makeSourceCatalogSortKey returns a sortable key for all the SourceMainCatalog entries withing in EnvironmentInfo
func (env *EnvironmentInfo) makeSourceCatalogSortKey(cat *catalogapi.Resource) string {
	key := strings.Builder{}
	if invalid := env.checkSourceCatalogValid(cat); invalid == "" {
		key.WriteString("1.Valid/")
	} else {
		key.WriteString("2.Invalid/")
	}
	if cat.Active {
		key.WriteString("1.Active/")
	} else {
		key.WriteString("2.Inactive/")
	}
	key.WriteString(cat.ID)
	return key.String()
}

// LoadDoctorEnvironments loads environment records from Doctor into the model
func LoadDoctorEnvironments() error {
	err := doctor.ListEnvironments(nil, func(e *doctor.EnvironmentEntry) {
		debug.Debug(debug.Environments, "LoadAllEntries().Doctor Environments: processing %s", e.String())
		normalizedCRN, err := crn.ParseAndNormalize(e.NewCRN, crnFromDoctor...)
		if err != nil {
			debug.PrintError("ossmerge.LoadAllEntries.Doctor Environments: found entry with invalid new_crn: %s  -- %v", e.String(), err)
			return
		}
		env, _ := LookupEnvironment(normalizedCRN, true)
		env.SourceDoctorEnvironments = append(env.SourceDoctorEnvironments, e)
	})
	if err != nil {
		return (debug.WrapError(err, "Error listing Environments from Doctor"))
	}
	return nil
}

// LoadDoctorRegionIDs loads regionID record from Doctor into the model
func LoadDoctorRegionIDs() error {
	err := doctor.ListRegionIDs(nil, func(e *doctor.RegionID) {
		debug.Debug(debug.Environments, "LoadAllEntries().Doctor RegionIDs: processing %s", e.String())
		rawCRN, err := e.CRNMask()
		if err != nil {
			debug.PrintError("ossmerge.LoadAllEntries.Doctor RegionIDs: found entry with no CRN: %s  -- %v", e.String(), err)
			return
		}
		normalizedCRN, err := rawCRN.Normalize(crnFromDoctor...)
		if err != nil {
			debug.PrintError("ossmerge.LoadAllEntries.Doctor RegionIDs: found entry with invalid CRN: %s  -- %v", e.String(), err)
			return
		}
		env, _ := LookupEnvironment(normalizedCRN, true)
		env.SourceDoctorRegionIDs = append(env.SourceDoctorRegionIDs, e)
	})
	if err != nil {
		return (debug.WrapError(err, "Error listing RegionIDs from Doctor"))
	}
	return nil
}

//var imsIDRegex = regexp.MustCompile(`^SoftLayer\s+[A-Za-z]+[0-9]+\s+\(([0-9]+)\)\s*$)`)
var imsIDRegex = regexp.MustCompile(`^SoftLayer\s+[A-Za-z]+[0-9]+\s+\(([0-9]+)\)\s*$`)

//var doctorEnvRegex = regexp.MustCompile(`DoctorEnvironment\([^\]]+\]`)
var doctorEnvRegex = regexp.MustCompile(`DoctorEnvironment\(.+?\[.+?\]`)

//var doctorRegionIDRegex = regexp.MustCompile(`DoctorRegionID\([^\)]+\)`)
var doctorRegionIDRegex = regexp.MustCompile(`DoctorRegionID\(.+?/mccpid=.+?(\(\d+\))?\)`)

// MergePhaseOne consolidates and merges the information from availables sources in a
// EnvironmentInfo record and constructs the OSSEnvironment record.
// First phase of the overall merge.
// This phase operates on individual records; while operating in one record, we cannot assume anything about the state of
// any other records.
func (env *EnvironmentInfo) mergePhaseOne() {
	if debug.IsDebugEnabled(debug.Environments) {
		debug.Debug(debug.Environments, "EnvironmentInfo.mergePhaseOne() starting -> %s", env.String())
	}
	oss := &env.OSSEnvironment

	// Compute valid MainCatalog sources
	//TODO: account for visibility of Catalog environment entries when selecting the "best" entry
	if env.HasSourceMainCatalog() {
		if len(env.AdditionalMainCatalog) > 0 {
			// Sort the AdditionalMainCatalog list to ensure consistent results
			list := env.AdditionalMainCatalog
			var originalCat = /* catalogapi.Resource  */ *env.GetSourceMainCatalog() // Force a copy
			list = append(list, &originalCat)
			sort.Slice(list, func(i, j int) bool {
				return env.makeSourceCatalogSortKey(list[i]) < env.makeSourceCatalogSortKey(list[j])
			})
			env.SourceMainCatalog = *list[0]
			env.AdditionalMainCatalog = list[1:]
		}
		cat := env.GetSourceMainCatalog()
		if invalid := env.checkSourceCatalogValid(cat); invalid != "" {
			var originalCat = /* catalogapi.Resource  */ *env.GetSourceMainCatalog() // Force a copy
			env.AdditionalMainCatalog = append([]*catalogapi.Resource{&originalCat}, env.AdditionalMainCatalog...)
			debug.Debug(debug.Environments, "EnvironmentInfo.mergePhaseOne() ignoring SourceMainCatalog(%s) -> %s", cat.String(), env.String())
			cat.Name = ""
		}
		for _, cat0 := range env.AdditionalMainCatalog {
			invalid := env.checkSourceCatalogValid(cat0)
			if invalid == invalidDedicatedLocal {
				env.OSSValidation.AddSource(cat0.ObjectMetaData.Deployment.TargetCRN, ossvalidation.CATALOGIGNORED)
				env.AddValidationIssue(ossvalidation.IGNORE, "Ignoring Legacy Catalog entry for Dedicated/Local environment", "%s  TargetCRN=%s", cat0.String(), cat0.ObjectMetaData.Deployment.TargetCRN).TagEnvironments()
			} else if invalid != "" {
				env.OSSValidation.AddSource(cat0.ObjectMetaData.Deployment.TargetCRN, ossvalidation.CATALOGIGNORED)
				env.AddValidationIssue(ossvalidation.SEVERE, "Ignoring invalid Catalog entry", "(%s)  %s  TargetCRN=%s", invalid, cat0.String(), cat0.ObjectMetaData.Deployment.TargetCRN).TagEnvironments()
			} else if cat0.Active == false {
				env.OSSValidation.AddSource(cat0.ObjectMetaData.Deployment.TargetCRN, ossvalidation.CATALOGIGNORED)
				env.AddValidationIssue(ossvalidation.CRITICAL, "Found additional Main Catalog entry (inactive) with the same EnvironmentID", "%s  TargetCRN=%s", cat0.String(), cat0.ObjectMetaData.Deployment.TargetCRN).TagEnvironments()
			} else {
				env.OSSValidation.AddSource(cat0.ObjectMetaData.Deployment.TargetCRN, ossvalidation.CATALOGIGNORED)
				env.AddValidationIssue(ossvalidation.CRITICAL, "Found additional Main Catalog entry (active) with the same EnvironmentID", "%s  TargetCRN=%s", cat0.String(), cat0.ObjectMetaData.Deployment.TargetCRN).TagEnvironments()
			}
		}
	}

	var doctorEnvString, doctorEnvCRN string
	var doctorRegionIDString, doctorRegionIDCRN string
	if ossrunactions.Doctor.IsEnabled() {
		env.OSSValidation.RecordRunAction(ossrunactions.Doctor)

		// Compute valid Doctor Environment sources
		// XXX for now we just take the first Doctor Environment source and ignore the restx
		if len(env.SourceDoctorEnvironments) > 1 {
			sort.Slice(env.SourceDoctorEnvironments, func(i, j int) bool {
				return (env.SourceDoctorEnvironments[i].NewCRN + env.SourceDoctorEnvironments[i].EnvName) < (env.SourceDoctorEnvironments[j].NewCRN + env.SourceDoctorEnvironments[j].EnvName)
			})
			for _, dr0 := range env.SourceDoctorEnvironments[:1] {
				env.OSSValidation.AddSource(dr0.NewCRN, ossvalidation.DOCTORENVIGNORED)
				env.AddValidationIssue(ossvalidation.CRITICAL, "Found additional Doctor Environment entry with the same EnvironmentID", "%s", dr0.String()).TagEnvironments()
			}
		}

		// Compute valid Doctor RegionID sources
		// XXX for now we just take the first Doctor Environment source and ignore the rest
		if len(env.SourceDoctorRegionIDs) > 1 {
			sort.Slice(env.SourceDoctorRegionIDs, func(i, j int) bool {
				crn1, _ := env.SourceDoctorRegionIDs[i].CRNMask()
				crn2, _ := env.SourceDoctorRegionIDs[j].CRNMask()
				return (string(crn1.ToCRNString()) + env.SourceDoctorRegionIDs[i].Name) < (string(crn2.ToCRNString()) + env.SourceDoctorRegionIDs[j].Name)
			})
			for _, dr0 := range env.SourceDoctorRegionIDs[1:] {
				rawCRN, _ := dr0.CRNMask()
				env.OSSValidation.AddSource(string(rawCRN.ToCRNString()), ossvalidation.DOCTORREGIONIDIGNORED)
				env.AddValidationIssue(ossvalidation.CRITICAL, "Found additional Doctor RegionID entry with the same EnvironmentID", "%s", dr0.String()).TagEnvironments()
			}
		}

	} else {
		env.OSSValidation.CopyRunAction(env.GetPriorOSSValidation(), ossrunactions.Doctor)
		if env.HasPriorOSS() && env.HasPriorOSSValidation() {
			// Copy the Doctor source names from prior run
			priorOSSVal := env.GetPriorOSSValidation()
			priorDesc := env.GetPriorOSS().Description
			doctorEnvString = doctorEnvRegex.FindString(priorDesc)
			doctorRegionIDString = doctorRegionIDRegex.FindString(priorDesc)
			namesEnv := collections.NewStringSet()
			namesEnv.Add(priorOSSVal.SourceNames(ossvalidation.DOCTORENV)...)
			namesEnv.Add(priorOSSVal.SourceNames(ossvalidation.DOCTORENVIGNORED)...)
			namesEnv.Add(priorOSSVal.SourceNames(ossvalidation.DOCTORENVDISABLED)...)
			namesRegion := collections.NewStringSet()
			namesRegion.Add(priorOSSVal.SourceNames(ossvalidation.DOCTORREGIONID)...)
			namesRegion.Add(priorOSSVal.SourceNames(ossvalidation.DOCTORREGIONIDIGNORED)...)
			namesRegion.Add(priorOSSVal.SourceNames(ossvalidation.DOCTORREGIONIDDISABLED)...)
			if namesEnv.Len() > 0 || namesRegion.Len() > 0 {
				env.AddValidationIssue(ossvalidation.INFO, "Copying prior info from Doctor (but not actually fetching new data from Doctor)", "%q %q", namesEnv.Slice(), namesRegion.Slice()).TagCRN().TagPriorOSS()
				for i, n := range namesEnv.Slice() {
					env.OSSValidation.AddSource(n, ossvalidation.DOCTORENVDISABLED)
					if i == 0 {
						doctorEnvCRN = n
					}
				}
				for i, n := range namesRegion.Slice() {
					env.OSSValidation.AddSource(n, ossvalidation.DOCTORREGIONIDDISABLED)
					if i == 0 {
						doctorRegionIDCRN = n
					}
				}
			}
		}
	}

	// Set SchemaVersion
	// (unconditional)
	env.OSSEnvironment.SchemaVersion = ossrecord.OSSCurrentSchema

	// Set OSSTags
	// Do this early, so that we can refer to those tags to determine other attributes
	if env.HasPriorOSS() {
		oss.OSSTags = env.GetPriorOSS().OSSTags.Copy()
		oss.OSSOnboardingPhase = env.GetPriorOSS().OSSOnboardingPhase
		oss.OSSOnboardingApprover = env.GetPriorOSS().OSSOnboardingApprover
		oss.OSSOnboardingApprovalDate = env.GetPriorOSS().OSSOnboardingApprovalDate
	}

	// Set OwningClient
	if env.HasPriorOSS() {
		oss.OwningClient = env.GetPriorOSS().OwningClient
	}

	// Set EnvironmentID and sources
	if env.HasPriorOSS() {
		// Note: we take the Environment from PriorOSS first, because otherwise we might be left with an orphan OSS record in the Catalog if we end-up picking a different EnvironmentID.
		// This is different from the handling of Type, Status and most other attributes, where we take the value from the PriorOSS record last.
		oss.EnvironmentID = env.GetPriorOSS().EnvironmentID
		var err error
		env.mergeWorkArea.finalCRNMask, err = crn.Parse(string(oss.EnvironmentID))
		if err != nil {
			panic(err) // Should not be possible to have an invalid CRN in a PriorOSS entry; we should have rejected that during loading
		}
		env.OSSValidation.AddSource(string(env.GetPriorOSS().EnvironmentID), ossvalidation.PRIOROSS)
		env.AddValidationIssue(ossvalidation.INFO, "Source: PriorOSS", "%s", env.GetPriorOSS().String()).TagEnvironments()
		if len(env.AdditionalPriorOSS) > 0 {
			env.AddValidationIssue(ossvalidation.CRITICAL, "More than one PriorOSS record with the same EnvironmentID", "%v %v", env.GetPriorOSS(), env.AdditionalPriorOSS).TagEnvironments()
		}
	}
	if env.HasSourceMainCatalog() {
		cat := env.GetSourceMainCatalog()
		depl := cat.ObjectMetaData.Deployment
		normalized, err := crn.ParseAndNormalize(depl.TargetCRN, crnFromCatalog...)
		if err != nil {
			panic(err) // We should have already validated this while loading the Catalog entries into the Model
		}
		if oss.EnvironmentID == "" {
			oss.EnvironmentID = ossrecord.EnvironmentID(normalized.ToCRNString())
			env.mergeWorkArea.finalCRNMask = normalized
		}
		env.OSSValidation.AddSource(depl.TargetCRN, ossvalidation.CATALOG)
		env.AddValidationIssue(ossvalidation.INFO, "Source: Main Catalog", "%s  TargetCRN=%s", cat.String(), depl.TargetCRN).TagEnvironments()
		env.mergeWorkArea.catalogCRNMask = normalized
		if !cat.IsPublicVisibleInactiveOK() {
			env.OSSValidation.RecordCatalogVisibility(cat.EffectiveVisibility.Restrictions, cat.Visibility.Restrictions, cat.Active, cat.ObjectMetaData.UI.Hidden, cat.Disabled)
			env.AddValidationIssue(ossvalidation.INFO, "Main Catalog entry is not public", "%+v", &env.OSSValidation.CatalogVisibility).TagEnvironments()
		}
		if cat.Active == false {
			env.AddValidationIssue(ossvalidation.IGNORE, "Main Catalog entry is Inactive", "%s", cat.String()).TagEnvironments()
		}
		if depl.TargetCRN != string(oss.EnvironmentID) {
			env.AddValidationIssue(ossvalidation.IGNORE, "Main Catalog entry has TargetCRN  that is different from the canonical CRN Mask for this Environment", "%s  TargetCRN=%s", cat.String(), depl.TargetCRN).TagEnvironments().TagDataMismatch()
		}
		// TODO: validate that depl.TargetCRN matches depl.Location and cat.Name
	}
	if ossrunactions.Doctor.IsEnabled() {
		if env.HasSourceDoctorEnvironment() {
			dr := env.GetSourceDoctorEnvironment()
			normalized, err := crn.ParseAndNormalize(dr.NewCRN, crnFromDoctor...)
			if err != nil {
				panic(err) // We should have already validated this while loading the Doctor Environment entries into the Model
			}
			if oss.EnvironmentID == "" {
				oss.EnvironmentID = ossrecord.EnvironmentID(normalized.ToCRNString())
				env.mergeWorkArea.finalCRNMask = normalized
			}
			env.OSSValidation.AddSource(dr.NewCRN, ossvalidation.DOCTORENV)
			env.AddValidationIssue(ossvalidation.INFO, "Source: Doctor Environment", "%s  CRN Mask=%s", dr.String(), dr.NewCRN).TagEnvironments()
			env.mergeWorkArea.doctorEnvironmentCRNMask = normalized
			if dr.NewCRN != string(oss.EnvironmentID) {
				env.AddValidationIssue(ossvalidation.MINOR, "Doctor Environment entry has CRN Mask (new_crn) that is different from the canonical CRN Mask for this Environment", "%s  CRN=%s", dr.String(), dr.NewCRN).TagEnvironments().TagDataMismatch()
			}
		}
		if env.HasSourceDoctorRegionID() {
			dr := env.GetSourceDoctorRegionID()
			rawCRN, err := dr.CRNMask()
			if err != nil {
				panic(err) // We should have already validated this while loading the Doctor RegionID entries into the Model
			}
			normalized, err := rawCRN.Normalize(crnFromDoctor...)
			if err != nil {
				panic(err) // We should have already validated this while loading the Doctor RegionID entries into the Model
			}
			if oss.EnvironmentID == "" {
				oss.EnvironmentID = ossrecord.EnvironmentID(normalized.ToCRNString())
				env.mergeWorkArea.finalCRNMask = normalized
			}
			rawCRNstr := string(rawCRN.ToCRNString())
			env.OSSValidation.AddSource(rawCRNstr, ossvalidation.DOCTORREGIONID)
			env.AddValidationIssue(ossvalidation.INFO, "Source: Doctor Region ID", "%s  CRN Mask=%s", dr.String(), rawCRNstr).TagEnvironments()
			//	env.mergeWorkArea.doctorRegionIDCRNMask = normalized
			if rawCRNstr != string(oss.EnvironmentID) {
				env.AddValidationIssue(ossvalidation.MINOR, "Doctor Region ID entry has CRN Mask that is different from the canonical CRN Mask for this Environment", "%s  CRN=%s", dr.String(), rawCRNstr).TagEnvironments().TagDataMismatch()
			}
			if dr.CRN == "" {
				env.AddValidationIssue(ossvalidation.MINOR, "Doctor Region ID entry has empty CRN attribute -- constructing one from the MCCPID", "%s  constructed.CRN=%s", dr.String(), rawCRNstr).TagEnvironments().TagDataMissing()
			} else if dr.CRN != rawCRNstr {
				// We should never get here with current implementation, because dr.CRNMask() will normally return the main CRN attribute if not empty and not construct one from the MCCPID
				env.AddValidationIssue(ossvalidation.SEVERE, "Doctor Region ID entry has CRN attribute different from the CRN constructed from the MCCPID", "%s  constructed.CRN=%s", dr.String(), rawCRNstr).TagEnvironments().TagDataMismatch()
			}
		}
	} else {
		if doctorEnvString != "" {
			env.AddValidationIssue(ossvalidation.INFO, "Source: Doctor Environment", "%s  CRN Mask=%s", doctorEnvString, doctorEnvCRN).TagEnvironments()
		}
		if doctorRegionIDString != "" {
			env.AddValidationIssue(ossvalidation.INFO, "Source: Doctor Region ID", "%s  CRN Mask=%s", doctorRegionIDString, doctorRegionIDCRN).TagEnvironments()
		}
	}
	/* Not needed - entry will be handled by EnvironmentInfo.IsDeletable()
	if oss.EnvironmentID == "" && env.HasPriorOSS() && env.GetPriorOSS().OSSTags.Contains(osstags.OSSOnly) {
		oss.EnvironmentID = env.GetPriorOSS().EnvironmentID
		env.mergeWorkArea.finalCRNMask, err = normalized
	}
	*/
	if oss.EnvironmentID != "" {
		env.OSSValidation.SetSourceNameCanonical(string(oss.EnvironmentID))
		env.AddValidationIssue(ossvalidation.INFO, "Merge Sources", "%v", env.OSSValidation.AllSources()).TagEnvironments()
	}

	if !env.IsValid() {
		// Bail out
		return
	}

	// Set ParentID
	// --> deferred to Merge Phase Two

	// Set OwningSegment
	// Must happen before we set the type, in order to find GaaS environments
	if env.HasPriorOSS() {
		if oss.OSSTags.Contains(osstags.IBMCloudDefaultSegment) {
			oss.OSSTags.RemoveTag(osstags.IBMCloudDefaultSegment)
			oss.OwningSegment = ""
		} else {
			oss.OwningSegment = env.GetPriorOSS().OwningSegment
		}
	}
	if oss.OwningSegment == "" {
		switch len(IBMCloudDefaultSegment) {
		case 1:
			oss.OwningSegment = IBMCloudDefaultSegment[0].OSSSegment.SegmentID
			oss.OSSTags.AddTag(osstags.IBMCloudDefaultSegment)
			env.AddValidationIssue(ossvalidation.INFO, "Setting the OwningSegment to the default IBM Cloud segment", "%s", IBMCloudDefaultSegment[0].OSSSegment.String()).TagEnvironments()
		case 0:
			env.AddValidationIssue(ossvalidation.SEVERE, "No OwningSegment explicitly set and cannot use the default IBM Cloud segment because it itself is not defined", "").TagEnvironments()
		default:
			env.AddValidationIssue(ossvalidation.CRITICAL, "No OwningSegment explicitly set and cannot use the default IBM Cloud segment because more than one is defined (incorrectly)", "%v", IBMCloudDefaultSegment).TagEnvironments()
		}
	}

	// Set Type
	if oss.OwningSegment != "" && !oss.OSSTags.Contains(osstags.IBMCloudDefaultSegment) {
		if seg, found := LookupSegment(oss.OwningSegment, false); found {
			switch seg.OSSSegment.SegmentType {
			case ossrecord.SegmentTypeIBMPublicCloud:
				// keep blank for now
			case ossrecord.SegmentTypeGaaS:
				oss.Type = ossrecord.EnvironmentGAAS
			default:
				env.AddValidationIssue(ossvalidation.SEVERE, "Unknown type for OwningSegment - cannot determine if this is a GaaS environment", "%s   Type=%s", seg.OSSSegment.String(), seg.OSSSegment.SegmentType).TagEnvironments()
			}
		} else {
			env.AddValidationIssue(ossvalidation.SEVERE, "Unknown OwningSegment - cannot determine if this is a GaaS environment", "SegmentID=%s", oss.OwningSegment).TagEnvironments()
		}
	}
	if env.HasSourceMainCatalog() {
		cat := env.GetSourceMainCatalog()
		if oss.Type == ossrecord.EnvironmentGAAS {
			env.AddValidationIssue(ossvalidation.SEVERE, "Environment is of type GaaS but has a Main Catalog entry", "%s", cat.String()).TagEnvironments()
		} else {
			if env.mergeWorkArea.catalogCRNMask.IsIBMPublicCloud() {
				switch cat.Kind {
				case catalogapi.KindRegion:
					oss.Type = ossrecord.EnvironmentIBMCloudRegion
				case catalogapi.KindDatacenter:
					oss.Type = ossrecord.EnvironmentIBMCloudDatacenter
				case catalogapi.KindAvailabilityZone:
					oss.Type = ossrecord.EnvironmentIBMCloudZone
				case catalogapi.KindPOP:
					oss.Type = ossrecord.EnvironmentIBMCloudPOP
				case catalogapi.KindSatellite:
					oss.Type = ossrecord.EnvironmentIBMCloudSatellite
					oss.OSSTags.AddTag(osstags.OSSStaging) // XXX Remove when we are ready to use in Production
				default:
					panic(fmt.Sprintf(`EnvironmentInfo.mergePhaseOne() unexpected entry kind in Global Catalog: "%s"  (entry=%s)`, cat.Kind, cat.String()))
				}
			} else {
				env.AddValidationIssue(ossvalidation.SEVERE, "Catalog Entry TargetCRN does not represent a IBM Public Cloud environment", "TargetCRN=%s  /  normalized=%s", env.GetSourceMainCatalog().ObjectMetaData.Deployment.TargetCRN, env.mergeWorkArea.catalogCRNMask.ToCRNString()).TagEnvironments()
			}
		}
	}
	if env.HasSourceDoctorEnvironment() {
		if oss.Type == ossrecord.EnvironmentGAAS {
			env.AddValidationIssue(ossvalidation.SEVERE, "Environment is of type GaaS but has a Doctor Environment entry", "%s", env.GetSourceDoctorEnvironment().String()).TagEnvironments()
		} else {
			var doctorType ossrecord.EnvironmentType
			switch {
			case env.mergeWorkArea.doctorEnvironmentCRNMask.IsIBMPublicCloud():
				if oss.Type == ossrecord.EnvironmentIBMCloudRegion || oss.Type == ossrecord.EnvironmentIBMCloudDatacenter || oss.Type == ossrecord.EnvironmentIBMCloudZone || oss.Type == ossrecord.EnvironmentIBMCloudSatellite {
					// compatible types -- no issue
				} else if oss.Type != "" {
					doctorType = "<ibm-public-cloud>"
				} else {
					// TODO: parse the location in Doctor CRN to determine type of public environment
					env.AddValidationIssue(ossvalidation.SEVERE, "Cannot determine exact Environment type for public region/datacenter/zone from Doctor Environment entry alone", "").TagEnvironments().TagDataMissing()
				}
			case env.mergeWorkArea.doctorEnvironmentCRNMask.IsIBMCloudDedicated():
				doctorType = ossrecord.EnvironmentIBMCloudDedicated
			case env.mergeWorkArea.doctorEnvironmentCRNMask.IsIBMCloudLocal():
				doctorType = ossrecord.EnvironmentIBMCloudLocal
			case env.mergeWorkArea.doctorEnvironmentCRNMask.IsIBMCloudStaging():
				doctorType = ossrecord.EnvironmentIBMCloudStaging
			default:
				// TODO: use other attributes in Doctor entry to determine environment type
				env.AddValidationIssue(ossvalidation.SEVERE, "Cannot determine Environment type from Doctor Environment entry CRN", "%s  (normalized=%s)", env.GetSourceDoctorEnvironment().NewCRN, env.mergeWorkArea.doctorEnvironmentCRNMask.ToCRNString()).TagEnvironments().TagDataMissing()
			}
			if oss.Type != "" {
				if doctorType != "" && oss.Type != doctorType {
					env.AddValidationIssue(ossvalidation.SEVERE, "Environment type determined from Main Catalog entry does not match type determined from Doctor Environment entry -- resetting", "Catalog=%s   Doctor=%s", oss.Type, doctorType).TagEnvironments().TagDataMismatch()
					oss.Type = ossrecord.EnvironmentTypeUnknown
				}
			} else {
				oss.Type = doctorType
			}
		}
	}
	if env.HasSourceDoctorRegionID() {
		// TODO: try to infer Type from Doctor RegionID
	}
	if oss.Type == "" && env.HasPriorOSS() {
		if env.GetPriorOSS().Type != "" && env.GetPriorOSS().Type != ossrecord.EnvironmentTypeUnknown {
			oss.Type = env.GetPriorOSS().Type
			if ossrunactions.Doctor.IsEnabled() {
				if oss.Type == ossrecord.EnvironmentGAAS {
					env.AddValidationIssue(ossvalidation.MINOR, "No Environment Type from either Main Catalog or Doctor for a GaaS environment -- using value from prior OSS record", "%s", oss.Type).TagEnvironments()
				} else {
					env.AddValidationIssue(ossvalidation.WARNING, "Cannot find Environment Type from either Main Catalog or Doctor -- using value from prior OSS record", "%s", oss.Type).TagEnvironments().TagDataMismatch()
				}
			}
		}
	}
	if oss.Type == "" {
		env.AddValidationIssue(ossvalidation.SEVERE, "Cannot set Environment Type from any available sources", "").TagEnvironments().TagDataMissing()
		oss.Type = ossrecord.EnvironmentTypeUnknown
	}

	// Set Status
	var overrideStatus bool
	if oss.OSSTags.Contains(osstags.NotReady) {
		oss.Status = ossrecord.EnvironmentNotReady
		env.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Forcing environment status to %s with %s OSS tag", ossrecord.EnvironmentNotReady, osstags.NotReady), "").TagEnvironments()
		overrideStatus = true
	} else if oss.OSSTags.Contains(osstags.SelectAvailability) {
		oss.Status = ossrecord.EnvironmentSelectAvailability
		env.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Forcing environment status to %s with %s OSS tag", ossrecord.EnvironmentSelectAvailability, osstags.SelectAvailability), "").TagEnvironments()
		overrideStatus = true
	} else if oss.OSSTags.Contains(osstags.Retired) {
		oss.Status = ossrecord.EnvironmentDecommissioned
		env.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Forcing environment status to %s with %s OSS tag", ossrecord.EnvironmentDecommissioned, osstags.Retired), "").TagEnvironments()
		overrideStatus = true
	}
	var catalogStatus ossrecord.EnvironmentStatus
	if env.HasSourceMainCatalog() {
		if !env.GetSourceMainCatalog().IsPublicVisibleInactiveOK() {
			catalogStatus = ossrecord.EnvironmentSelectAvailability
			if overrideStatus {
				env.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Environment would be marked %s because the Main Catalog entry is not public -- but already overridden by a OSS tag", catalogStatus), "").TagEnvironments()
			} else {
				env.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Marking Environment %s because the Main Catalog entry is not public", catalogStatus), "").TagEnvironments()
			}
		} else {
			catalogStatus = ossrecord.EnvironmentActive
		}
		if oss.Status == "" {
			oss.Status = catalogStatus
		}
	} else {
		switch oss.Type {
		// Public Environments MUST be represented in Global Catalog to be valid
		case ossrecord.EnvironmentIBMCloudRegion,
			ossrecord.EnvironmentIBMCloudZone,
			ossrecord.EnvironmentIBMCloudDatacenter,
			ossrecord.EnvironmentIBMCloudPOP,
			ossrecord.EnvironmentIBMCloudSatellite:
			if oss.EnvironmentID == "crn:v1:bluemix:public::satellite::::" { // Special case for the dummy entry for Satellite
				break
			}
			catalogStatus = ossrecord.EnvironmentDecommissioned
			if overrideStatus {
				env.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Public Environment would be marked %s because the Main Catalog entry is not found -- but already overridden by a OSS tag", catalogStatus), "").TagEnvironments()
			} else {
				env.AddValidationIssue(ossvalidation.INFO, fmt.Sprintf("Marking Public Environment %s because the Main Catalog entry is not found", catalogStatus), "").TagEnvironments()
			}
			if oss.Status == "" {
				oss.Status = catalogStatus
			}
		}
	}
	if env.HasSourceDoctorEnvironment() {
		var doctorStatus ossrecord.EnvironmentStatus
		if strings.HasSuffix(strings.TrimSpace(strings.ToLower(env.GetSourceDoctorEnvironment().EnvName)), "decommissioned") {
			doctorStatus = ossrecord.EnvironmentDecommissioned
		} else {
			doctorStatus = ossrecord.EnvironmentActive
		}
		if oss.Status != "" {
			if catalogStatus != doctorStatus {
				if overrideStatus {
					env.AddValidationIssue(ossvalidation.MINOR, "Environment status determined from Main Catalog entry does not match status determined from Doctor Environment entry -- but already overridden by a OSS tag", "Catalog=%s   Doctor=%s", catalogStatus, doctorStatus).TagEnvironments().TagDataMismatch()
				} else {
					env.AddValidationIssue(ossvalidation.SEVERE, "Environment status determined from Main Catalog entry does not match status determined from Doctor Environment entry -- resetting", "Catalog=%s   Doctor=%s", catalogStatus, doctorStatus).TagEnvironments().TagDataMismatch()
					oss.Type = ossrecord.EnvironmentTypeUnknown
				}
			}
		} else {
			oss.Status = doctorStatus
		}
	}
	if env.HasSourceDoctorRegionID() {
		// TODO: try to infer Status from Doctor RegionID in more cases
		if oss.Status == "" && strings.HasSuffix(strings.TrimSpace(strings.ToLower(env.GetSourceDoctorRegionID().Name)), "decommissioned") {
			oss.Status = ossrecord.EnvironmentDecommissioned
		}
	}
	if oss.Status == "" && env.HasPriorOSS() {
		if env.GetPriorOSS().Status != "" && env.GetPriorOSS().Status != ossrecord.EnvironmentStatusUnknown {
			oss.Status = env.GetPriorOSS().Status
			if ossrunactions.Doctor.IsEnabled() {
				if oss.Type == ossrecord.EnvironmentGAAS {
					env.AddValidationIssue(ossvalidation.INFO, "No Environment Status from either Main Catalog or Doctor for a GaaS environment -- using value from prior OSS record", "%s", oss.Status).TagEnvironments()
				} else {
					env.AddValidationIssue(ossvalidation.WARNING, "Cannot find Environment Status from either Main Catalog or Doctor -- using value from prior OSS record", "%s", oss.Status).TagEnvironments().TagDataMismatch()
				}
			}
		}
	}
	if oss.Status == "" {
		env.AddValidationIssue(ossvalidation.SEVERE, "Cannot set Environment Status from any available sources", "").TagEnvironments().TagDataMissing()
		oss.Status = ossrecord.EnvironmentStatusUnknown
	}

	// Set DisplayName
	if env.HasSourceMainCatalog() {
		oss.DisplayName = env.GetSourceMainCatalog().OverviewUI.En.DisplayName
	}
	if env.HasSourceDoctorEnvironment() {
		if oss.DisplayName == "" {
			oss.DisplayName = env.GetSourceDoctorEnvironment().EnvName
		} else if oss.DisplayName != env.GetSourceDoctorEnvironment().EnvName {
			env.AddValidationIssue(ossvalidation.MINOR, "Display name is different between Main Catalog and Doctor Environment entry", `Catalog=%q   Doctor=%q`, oss.DisplayName, env.GetSourceDoctorEnvironment().EnvName).TagEnvironments().TagDataMismatch()
		}
	}
	if env.HasSourceDoctorRegionID() {
		if oss.DisplayName == "" {
			oss.DisplayName = env.GetSourceDoctorRegionID().Name
		} else if oss.DisplayName != env.GetSourceDoctorRegionID().Name {
			env.AddValidationIssue(ossvalidation.MINOR, "Display name is different in Doctor RegionID entry", `Catalog=%q   Doctor=%q`, oss.DisplayName, env.GetSourceDoctorRegionID().Name).TagEnvironments().TagDataMismatch()
		}
	}
	if oss.DisplayName == "" && env.HasPriorOSS() && env.GetPriorOSS().DisplayName != "" {
		oss.DisplayName = env.GetPriorOSS().DisplayName
		if ossrunactions.Doctor.IsEnabled() {
			if oss.Type == ossrecord.EnvironmentGAAS {
				env.AddValidationIssue(ossvalidation.INFO, "No Display name in either Main Catalog or Doctor for a GaaS environment -- using value from prior OSS record", "").TagEnvironments()
			} else {
				env.AddValidationIssue(ossvalidation.WARNING, "Cannot find a Display name in either Main Catalog or Doctor -- using value from prior OSS record", "").TagEnvironments().TagDataMismatch()
			}
		}
	}
	if oss.DisplayName == "" {
		env.AddValidationIssue(ossvalidation.SEVERE, "Cannot set Display Name from any available sources", "").TagDataMissing()
		oss.DisplayName = fmt.Sprintf("[%s]", oss.EnvironmentID)
	}
	osstags.CheckOSSTestTag(&oss.DisplayName, &oss.OSSTags)

	// Set ReferenceCatalogID
	if env.HasSourceMainCatalog() {
		oss.ReferenceCatalogID = ossrecord.CatalogID(env.GetSourceMainCatalog().ID)
		RecordEnvironmentByReferenceCatalogID(oss.ReferenceCatalogID, env)
		oss.ReferenceCatalogPath = env.GetSourceMainCatalog().CatalogPath
	}

	// Set Description
	desc := strings.Builder{}
	desc.WriteString(fmt.Sprintf("%s\n", oss.DisplayName))
	desc.WriteString("\n")
	if env.HasSourceMainCatalog() {
		desc.WriteString(fmt.Sprintf("    %s\n", env.GetSourceMainCatalog().String()))
	}
	if ossrunactions.Doctor.IsEnabled() {
		if env.HasSourceDoctorEnvironment() {
			desc.WriteString(fmt.Sprintf("    %s\n", env.GetSourceDoctorEnvironment().String()))
		}
		if env.HasSourceDoctorRegionID() {
			desc.WriteString(fmt.Sprintf("    %s\n", env.GetSourceDoctorRegionID().String()))
		}
	} else if env.HasPriorOSS() {
		if doctorEnvString != "" {
			desc.WriteString(fmt.Sprintf("    %s\n", doctorEnvString))
		}
		if doctorRegionIDString != "" {
			desc.WriteString(fmt.Sprintf("    %s\n", doctorRegionIDString))
		}
	}
	oss.Description = desc.String()

	// Set LegacyIMSID
	// TODO: extract IMSID for Doctor environment name
	if env.HasSourceDoctorEnvironment() {
		if m := imsIDRegex.FindStringSubmatch(env.GetSourceDoctorEnvironment().EnvName); m != nil {
			oss.LegacyIMSID = m[1]
		}
	}
	if env.HasSourceDoctorRegionID() {
		if m := imsIDRegex.FindStringSubmatch(env.GetSourceDoctorRegionID().Name); m != nil {
			if oss.LegacyIMSID == "" {
				oss.LegacyIMSID = m[1]
			} else if oss.LegacyIMSID != m[1] {
				env.AddValidationIssue(ossvalidation.SEVERE, "Environment Legacy IMSID determined from Doctor Environment entry does not match that determined from Doctor RegionID entry -- Environment entry prevails", "Doctor Enviroment=%s   Doctor RegionID=%s", oss.LegacyIMSID, m[1]).TagEnvironments().TagDataMismatch()
			}
		}
	}
	if oss.LegacyIMSID == "" && env.HasPriorOSS() && env.GetPriorOSS().LegacyIMSID != "" {
		oss.LegacyIMSID = env.GetPriorOSS().LegacyIMSID
		if ossrunactions.Doctor.IsEnabled() {
			env.AddValidationIssue(ossvalidation.INFO, "Cannot find a Legacy IMSID in either Main Catalog or Doctor -- using value from prior OSS record", "").TagEnvironments().TagDataMismatch()
		}
	}
	if oss.LegacyIMSID == "" {
		//		env.AddValidationIssue(ossvalidation.MINOR, "Cannot set Legacy IMSID from any available sources", "").TagDataMissing()
	}

	// Set LegacyMCCPID
	if env.HasSourceMainCatalog() {
		oss.LegacyMCCPID = env.GetSourceMainCatalog().ObjectMetaData.Deployment.MCCPID
	}
	if env.HasSourceDoctorEnvironment() {
		doctorMCCPID := env.GetSourceDoctorEnvironment().RegionID
		if oss.LegacyMCCPID != "" {
			if doctorMCCPID != "" && oss.LegacyMCCPID != doctorMCCPID {
				env.AddValidationIssue(ossvalidation.SEVERE, "Environment Legacy MCCPIP determined from Main Catalog entry does not match that determined from Doctor Environment entry -- Catalog prevails", "Catalog=%s   Doctor=%s", oss.LegacyIMSID, doctorMCCPID).TagEnvironments().TagDataMismatch()
			}
		} else {
			doctorMCCPID := env.GetSourceDoctorEnvironment().RegionID
			oss.LegacyMCCPID = doctorMCCPID
		}
	}
	if env.HasSourceDoctorRegionID() {
		doctorMCCPID := env.GetSourceDoctorRegionID().ID
		if oss.LegacyMCCPID != "" && oss.LegacyMCCPID != doctorMCCPID {
			env.AddValidationIssue(ossvalidation.SEVERE, "Environment Legacy MCCPIP determined from Doctor RegionID entry does not match that determined from other sources", "MCCPID=%s   Doctor.RegionID.MCCPIP=%s", oss.LegacyIMSID, doctorMCCPID).TagEnvironments().TagDataMismatch()
		} else {
			oss.LegacyMCCPID = doctorMCCPID
		}
	}
	if oss.LegacyMCCPID == "" && env.HasPriorOSS() && env.GetPriorOSS().LegacyMCCPID != "" {
		oss.LegacyMCCPID = env.GetPriorOSS().LegacyMCCPID
		if ossrunactions.Doctor.IsEnabled() {
			env.AddValidationIssue(ossvalidation.INFO, "Cannot find a Legacy MCCPID in either Main Catalog or Doctor -- using value from prior OSS record", "").TagEnvironments().TagDataMismatch()
		}
	}
	if oss.LegacyMCCPID == "" {
		//		env.AddValidationIssue(ossvalidation.MINOR, "Cannot set Legacy MCCPID from any available sources", "").TagDataMissing()
	}

	// Set LegacyDoctorCRN
	if env.HasSourceDoctorRegionID() {
		oss.LegacyDoctorCRN = env.GetSourceDoctorRegionID().CRN
	}
	if env.HasSourceDoctorEnvironment() {
		newCRN := env.GetSourceDoctorEnvironment().NewCRN
		if oss.LegacyDoctorCRN != "" {
			if newCRN != "" && oss.LegacyDoctorCRN != newCRN {
				env.AddValidationIssue(ossvalidation.SEVERE, `Legacy CRN from Doctor RegionID entry does not match the "new_crn" attribute from Doctor Environment entry -- Doctor RegionID prevails`, "Doctor.RegionID=%s   Doctor.Environment=%s", oss.LegacyDoctorCRN, newCRN).TagEnvironments().TagDataMismatch()
			}
		} else {
			oss.LegacyDoctorCRN = newCRN
		}
	}
	if oss.LegacyDoctorCRN == "" && env.HasPriorOSS() && env.GetPriorOSS().LegacyDoctorCRN != "" {
		oss.LegacyDoctorCRN = env.GetPriorOSS().LegacyDoctorCRN
		if ossrunactions.Doctor.IsEnabled() {
			env.AddValidationIssue(ossvalidation.INFO, "Cannot find a Legacy CRN from Doctor -- using value from prior OSS record", "").TagEnvironments().TagDataMismatch()
		}
	}
	if oss.LegacyMCCPID == "" {
		//		env.AddValidationIssue(ossvalidation.MINOR, "Cannot set Legacy MCCPID from any available sources", "").TagDataMissing()
	}
	if oss.LegacyDoctorCRN != "" && oss.LegacyDoctorCRN != string(oss.EnvironmentID) {
		env.AddValidationIssue(ossvalidation.IGNORE, "Legacy CRN from Doctor is not identical to the canonical EnvironmentID", "LegacyDoctorCRN=%s", oss.LegacyDoctorCRN).TagEnvironments()
	}

	// Final sanity check
	if env.IsValid() {
		if oss.EnvironmentID != ossrecord.EnvironmentID(env.mergeWorkArea.finalCRNMask.ToCRNString()) {
			env.AddValidationIssue(ossvalidation.CRITICAL, "EnvironmentInfo.mergePhaseOne() mismatch environment ID", "CRNMask=%s   NewOSS=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.EnvironmentID).TagCRN()
			//panic(fmt.Sprintf("EnvironmentInfo.mergePhaseOne() mismatch environment ID: CRNMask=%s   NewOSS=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.EnvironmentID))
		}
		switch oss.Type {
		case ossrecord.EnvironmentIBMCloudRegion, ossrecord.EnvironmentIBMCloudDatacenter, ossrecord.EnvironmentIBMCloudZone, ossrecord.EnvironmentIBMCloudPOP, ossrecord.EnvironmentIBMCloudSatellite:
			if !env.mergeWorkArea.finalCRNMask.IsIBMPublicCloud() {
				env.AddValidationIssue(ossvalidation.CRITICAL, "Final CRN Mask for IBM Public Cloud environment does not match the final Type", "CRNMask=%s   FinalType=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type).TagCRN()
				//panic(fmt.Sprintf("Merging environments: final CRN Mask=%s for IBM Public Cloud environment does not match the final Type=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type))
			}
		case ossrecord.EnvironmentIBMCloudDedicated:
			if !env.mergeWorkArea.finalCRNMask.IsIBMCloudDedicated() {
				env.AddValidationIssue(ossvalidation.CRITICAL, "Final CRN Mask for IBM Cloud Dedicated environment does not match the final Type", "CRNMask=%s   FinalType=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type).TagCRN()
				//panic(fmt.Sprintf("Merging environments: final CRN Mask=%s for IBM Cloud Dedicated environment does not match the final Type=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type))
			}
		case ossrecord.EnvironmentIBMCloudLocal:
			if !env.mergeWorkArea.finalCRNMask.IsIBMCloudLocal() {
				env.AddValidationIssue(ossvalidation.CRITICAL, "Final CRN Mask for IBM Cloud Local environment does not match the final Type", "CRNMask=%s   FinalType=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type).TagCRN()
				//panic(fmt.Sprintf("Merging environments: final CRN Mask=%s for IBM Cloud Local environment does not match the final Type=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type))
			}
		case ossrecord.EnvironmentIBMCloudStaging:
			if !env.mergeWorkArea.finalCRNMask.IsIBMCloudStaging() {
				env.AddValidationIssue(ossvalidation.CRITICAL, "Final CRN Mask for IBM Cloud Staging environment does not match the final Type", "CRNMask=%s   FinalType=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type).TagCRN()
				//panic(fmt.Sprintf("Merging environments: final CRN Mask=%s for IBM Cloud Staging environment does not match the final Type=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type))
			}
		case ossrecord.EnvironmentGAAS:
			if env.mergeWorkArea.finalCRNMask.IsAnyIBMCloud() {
				env.AddValidationIssue(ossvalidation.CRITICAL, "Final CRN Mask for IBM Cloud environment does not match the final Type", "CRNMask=%s   FinalType=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type).TagCRN()
				//panic(fmt.Sprintf("Merging environments: final CRN Mask=%s for IBM Cloud environment does not match the final Type=%s", env.mergeWorkArea.finalCRNMask.ToCRNString(), oss.Type))
			}
		case ossrecord.EnvironmentTypeSpecial:
			// ignore
		case ossrecord.EnvironmentTypeUnknown:
			// ignore
		default:
			panic(fmt.Sprintf("Merging environments: unexpected final Type=%s for environment with final CRN Mask=%s", oss.Type, env.mergeWorkArea.finalCRNMask.ToCRNString()))
		}
	}
}

// mergePhaseTwo continues the merge of OSSEnvironments. It handles relationships between multiple Environment records,
// and relies on the fact that each individual record has already been constructed (without relationships) in Phase EnvironmentsOne
func (env *EnvironmentInfo) mergePhaseTwo() {
	if !env.IsValid() {
		// Bail out
		return
	}

	if skip, _ := env.SkipMerge(); !skip {
		oss := &env.OSSEnvironment
		if env.HasSourceMainCatalog() && env.GetSourceMainCatalog().ParentID != "" {
			parentID := ossrecord.CatalogID(env.GetSourceMainCatalog().ParentID)
			if parent, found := LookupEnvironmentByReferenceCatalogID(parentID); found {
				oss.ParentID = parent.OSSEnvironment.EnvironmentID
				if env.GetSourceMainCatalog().Kind != catalogapi.KindAvailabilityZone {
					env.AddValidationIssue(ossvalidation.WARNING, "Main Catalog entry for something other than an Availability Zone has a parent that is itself another environment-type entry", "parent=%s", parent.OSSEnvironment.String()).TagEnvironments()
				}
			} else {
				if env.GetSourceMainCatalog().Kind == catalogapi.KindAvailabilityZone {
					env.AddValidationIssue(ossvalidation.SEVERE, "Main Catalog entry for this environment specifies a parent, but that parent is not itself found as an environment-type entry", "parentID=%s", parentID).TagEnvironments()
				} else {
					// Regions or Datacenters or POPs normally have a parent that is a geography/country/metro not itself another environment
					// We assume this is OK but we cannot verify at this point
					// TODO: verify that parent of region/datacenter is a geography/country/metro
				}
			}
		}
	} else {
		debug.Debug(debug.Merge, `ossmerge.EnvironmentsPhaseTwo(): Skipping merge of OSSEnvironment record "%s" that has the SkipMerge flag`, env.String())
	}

	// Take care of any side effects of IsDeletable() (e.g. new ValidationIssues)
	_ = env.IsDeletable()

	// TODO: Compute OSSEnvironment validation status (red/yellow/green)

	if env.OSSEnvironment.OSSOnboardingPhase == ossrecord.EDIT {
		env.OSSEnvironmentExtended.ResetForRMC()
	}

	// Sort the ValidationIssues and the tags
	if env.OSSValidation != nil {
		env.OSSValidation.Sort()
	}
	err := env.OSSEnvironment.OSSTags.Validate(true)
	if err != nil {
		panic(err)
	}
}

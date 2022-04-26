package ossmerge

import (
	"fmt"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/ossmergecontrol"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/supportcenter"

	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"

	"github.ibm.com/cloud-sre/osscatalog/monitoringinfo"

	"github.ibm.com/cloud-sre/osscatalog/legacy"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/iam"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/scorecardv1"
	"github.ibm.com/cloud-sre/osscatalog/servicenow"
)

var servicesByCatalogID = make(map[string]*ServiceInfo)
var servicesByPlanID = make(map[string]*ServiceInfo)
var inactivePlanIDs = make(map[string]string)

// LoadAllEntries loads all service/components, segment and tribe entries from all sources into the Model, that match a given name pattern,
// in preparation for merging.
func LoadAllEntries(pattern *regexp.Regexp) error {
	// Load the segments/tribes first, as this is the call that is most likely to fail (i.e. fail early)
	if ossrunactions.Tribes.IsEnabled() && ossrunactions.Doctor.IsEnabled() {
		err := LoadScorecardV1SegmentTribes()
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error loading Segments and Tribes from ScorecardV1"))
		}
	} else {
		debug.Info("Skip reloading Segment and Tribe Info from ScorecardV1")
	}

	// Load environments from Doctor -- should be a fairly short call, might as well fail early if at all
	if (ossrunactions.Environments.IsEnabled() || ossrunactions.EnvironmentsNative.IsEnabled()) && ossrunactions.Doctor.IsEnabled() {
		err := LoadDoctorEnvironments()
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error loading Environments from Doctor"))
		}
	}

	// Load additional environments from Doctor through the "regionid" URL -- should be a fairly short call, might as well fail early if at all
	if (ossrunactions.Environments.IsEnabled() || ossrunactions.EnvironmentsNative.IsEnabled()) && ossrunactions.Doctor.IsEnabled() {
		err := LoadDoctorRegionIDs()
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error loading RegionIDs from Doctor"))
		}
	}

	// TODO: When switching to Production, we should be reading OSS entries from the Production Global Catalog
	ossIncludeOptions := catalog.IncludeOSSMergeControl | catalog.IncludeOSSValidation
	if ossrunactions.Services.IsEnabled() {
		ossIncludeOptions |= catalog.IncludeServices
	}
	if ossrunactions.Tribes.IsEnabled() {
		ossIncludeOptions |= catalog.IncludeTribes
	}
	if ossrunactions.Environments.IsEnabled() {
		ossIncludeOptions |= catalog.IncludeEnvironments
	}
	if ossrunactions.EnvironmentsNative.IsEnabled() {
		ossIncludeOptions |= catalog.IncludeEnvironmentsNative
	}
	ossIncludeOptions |= catalog.IncludeServicesDomainOverrides // include service domain overrides to get the full service record
	err := catalog.ListOSSEntries(pattern, ossIncludeOptions, func(r ossrecord.OSSEntry) {
		switch r1 := r.(type) {
		case *ossrecordextended.OSSServiceExtended:
			if len(r1.OSSMergeControl.DoNotMergeNames) > 0 {
				registerDoNotMergeNames(r1.OSSMergeControl, r1.OSSMergeControl.DoNotMergeNames...)
			}
			comparableName := MakeComparableName(string(r1.OSSService.ReferenceResourceName))
			si, _ := LookupService(comparableName, true)
			if !si.HasPriorOSS() {
				si.PriorOSS = r1.OSSService
				si.OSSMergeControl = r1.OSSMergeControl
				si.PriorOSSValidationChecksum = r1.OSSValidation.Checksum()
				si.PriorOSSValidation = r1.OSSValidation
				si.Created = r1.Created
				si.Updated = r1.Updated
				debug.Debug(debug.Merge, "ossmerge.LoadAllEntries() -> loaded %s", si.String())
			} else {
				panic(fmt.Sprintf(`Found more than one OSS service/component record with the same comparableName="%s"  (name1="%s" / name2="%s")`, comparableName, si.PriorOSS.ReferenceResourceName, r1.OSSService.ReferenceResourceName))
			}
		case *ossrecordextended.OSSSegmentExtended:
			seg, _ := LookupSegment(r1.OSSSegment.SegmentID, true)
			if !seg.HasPriorOSS() {
				seg.PriorOSS = r1.OSSSegment
				seg.Created = r1.Created
				seg.Updated = r1.Updated
				if r1.OSSValidation != nil {
					seg.PriorOSSValidationChecksum = r1.OSSValidation.Checksum()
					seg.PriorOSSValidation = r1.OSSValidation
				}
				debug.Debug(debug.Merge, "ossmerge.LoadAllEntries() -> loaded %s", seg.String())
			} else {
				panic(fmt.Sprintf(`Found more than one OSSSegment entry with the same segmentID  %s - %s`, seg.PriorOSS.String(), r1.String()))
			}
		case *ossrecordextended.OSSTribeExtended:
			// TODO: Handle the case where the parent OSSSegment for a OSSTribe is never found in Catalog
			seg, _ := LookupSegment(r1.OSSTribe.SegmentID, true)
			tr, _ := seg.LookupTribe(r1.OSSTribe.TribeID, true)
			if !tr.HasPriorOSS() {
				tr.PriorOSS = r1.OSSTribe
				tr.Created = r1.Created
				tr.Updated = r1.Updated
				if r1.OSSValidation != nil {
					tr.PriorOSSValidationChecksum = r1.OSSValidation.Checksum()
					tr.PriorOSSValidation = r1.OSSValidation
				}
				debug.Debug(debug.Merge, "ossmerge.LoadAllEntries() -> loaded %s", tr.String())
			} else {
				panic(fmt.Sprintf(`Found more than one OSSTribe entry with the same segmentID  %s - %s`, tr.PriorOSS.String(), r1.String()))
			}
		case *ossrecordextended.OSSEnvironmentExtended:
			crnMask, err := crn.Parse(string(r1.OSSEnvironment.EnvironmentID))
			if err != nil {
				debug.PrintError("ossmerge.LoadAllEntries.OSS: found OSS entry with invalid environment ID: %s  -- %v", r1.String(), err)
				return
			}
			env, _ := LookupEnvironment(crnMask, true)
			if !env.HasPriorOSS() {
				env.PriorOSS = r1.OSSEnvironment
				env.PriorOSSValidationChecksum = r1.OSSValidation.Checksum()
				env.PriorOSSValidation = r1.OSSValidation
				env.Created = r1.Created
				env.Updated = r1.Updated
				debug.Debug(debug.Environments, "ossmerge.LoadAllEntries() -> loaded %s", env.String())
			} else {
				env.AdditionalPriorOSS = append(env.AdditionalPriorOSS, &r1.OSSEnvironment)
			}
		default:
			panic(fmt.Sprintf("catalog.ListOSSEntries(): Unexpected entry type: %#v\n", r))
		}
	})
	if err != nil {
		return (debug.WrapError(err, "LoadAllEntries(): error listing OSS entries from Global Catalog"))
	}

	if ossrunactions.Services.IsEnabled() || ossrunactions.Environments.IsEnabled() || ossrunactions.EnvironmentsNative.IsEnabled() {
		if options.GlobalOptions().RefreshPricing {
			debug.Info("Loading all Main Catalog entries including Pricing information (refresh pricing for all entries from BSS Pricing Catalog)")
		} else if options.GlobalOptions().IncludePricing {
			debug.Info("Loading all Main Catalog entries including Pricing information (pricing for new entries only from BSS Pricing Catalog)")
		}
		err = catalog.ListMainCatalogEntries(pattern, func(r *catalogapi.Resource) {
			switch r.Kind {
			// TODO: Handle plans/deployments associated with AdditionalMainCatalog entries instead of SourceMainCatalog
			case catalogapi.KindPlan, catalogapi.KindFlavor, catalogapi.KindProfile:
				if !ossrunactions.Services.IsEnabled() {
					return
				}
				debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: processing plan %s -- ParentID=%q", r.String(), r.ParentID)
				if r.Active || r.IsIaaSRegionsPlaceholder(true) {
					if r.ParentID == "" {
						debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: Found plan with empty ParentID: %s", r.String())
					} else if si, found := servicesByCatalogID[r.ParentID]; found {
						if si.GetSourceMainCatalog().Active {
							cex := si.GetCatalogExtra(true)
							cex.ChildrenKinds.Add(r.Kind)
							si.AddPlan(r.OverviewUI.En.DisplayName, r.ID)
							debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: adding plan %s to service %s", r.OverviewUI.En.DisplayName, si.String())
						} else {
							si.AddValidationIssue("Found Active Catalog plan under an Inactive service entry", "Child=%s   Parent=%s", r.String(), si.GetSourceMainCatalog().String()).TagCatalogConsistency()
						}
						if prior, found2 := servicesByPlanID[r.ID]; !found2 {
							servicesByPlanID[r.ID] = si
						} else {
							debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: Found duplicate service for plan %s: %s <-> %s", r.String(), si.SourceMainCatalog.String(), prior.SourceMainCatalog.String())
						}
					} else {
						debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: No service found for plan %s -- ParentID=%q", r.String(), r.ParentID)
					}
				} else {
					inactivePlanIDs[r.ID] = r.String()
					debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: recording inactive plan %s (%s)", r.OverviewUI.En.DisplayName, r.String())
				}
			case catalogapi.KindDeployment:
				if !ossrunactions.Services.IsEnabled() {
					return
				}
				debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: processing deployment %s -- ParentID=%q", r.String(), r.ParentID)
				if r.Active || r.IsIaaSRegionsPlaceholder(true) {
					if r.ParentID == "" {
						debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: Found deployment with empty ParentID: %s", r.String())
					} else if _, found := inactivePlanIDs[r.ParentID]; found {
						// ignore
						debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: ignoring location %s from inactive plan %s", r.ObjectMetaData.Deployment.Location, r.ParentID)
					} else if si, found := servicesByPlanID[r.ParentID]; found {
						if si.GetSourceMainCatalog().Active {
							if r.ObjectMetaData.Deployment != nil {
								cex := si.GetCatalogExtra(true)
								cex.Locations.Add(r.ObjectMetaData.Deployment.Location)
								debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: adding location %s to service %s (through plan %s)", r.ObjectMetaData.Deployment.Location, si.String(), r.ParentID)
							} else {
								si.AddValidationIssue(ossvalidation.WARNING, "Found a deployment record with no deployment meta-data", "%s", r.String()).TagCatalogConsistency()
							}
						} else {
							si.AddValidationIssue("Found Active Catalog deployment under an Inactive service entry", "(lookup by plan) - Child=%s   Parent=%s", r.String(), si.GetSourceMainCatalog().String()).TagCatalogConsistency()
						}
					} else if si, found := servicesByCatalogID[r.ParentID]; found {
						if si.GetSourceMainCatalog().Active {
							si.AddValidationIssue(ossvalidation.WARNING, "Found a deployment record that is a direct parent of a service, not of a plan", "%s", r.String()).TagCatalogConsistency()
							cex := si.GetCatalogExtra(true)
							cex.ChildrenKinds.Add(r.Kind)
							if r.ObjectMetaData.Deployment != nil {
								cex.Locations.Add(r.ObjectMetaData.Deployment.Location)
								debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: adding location %s to service %s (direct parent)", r.ObjectMetaData.Deployment.Location, si.String())
							} else {
								si.AddValidationIssue(ossvalidation.WARNING, "Found a deployment record with no deployment meta-data", "%s", r.String()).TagCatalogConsistency()
							}
						} else {
							si.AddValidationIssue("Found Active Catalog deployment under an Inactive service entry", "(lookup by service) - Child=%s   Parent=%s", r.String(), si.GetSourceMainCatalog().String()).TagCatalogConsistency()
						}
					} else {
						debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: No service found for deployment %s -- ParentID=%q", r.String(), r.ParentID)
					}
				} else {
					debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: skipping inactive deployment for location %s", r.ObjectMetaData.Deployment.Location)
				}
			case catalogapi.KindRegion, catalogapi.KindDatacenter, catalogapi.KindAvailabilityZone, catalogapi.KindPOP, catalogapi.KindLegacyCName, catalogapi.KindLegacyEnvironment, catalogapi.KindSatellite:
				if !ossrunactions.Environments.IsEnabled() && !ossrunactions.EnvironmentsNative.IsEnabled() {
					return
				}
				if r.ObjectMetaData.Deployment == nil {
					debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: found region-related entry without deployment meta-data: %s", r.String())
					return
				}
				debug.Debug(debug.Environments, "LoadAllEntries().Environments: processing %s    targetCRN=%s", r.String(), r.ObjectMetaData.Deployment.TargetCRN)
				normalizedCRN, err := crn.ParseAndNormalize(r.ObjectMetaData.Deployment.TargetCRN, crnFromCatalog...)
				if err != nil {
					debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: found region-related entry with invalid target_crn: %s  -- %v", r.String(), err)
					return
				}
				env, _ := LookupEnvironment(normalizedCRN, true)
				if !env.HasSourceMainCatalog() {
					env.SourceMainCatalog = *r
				} else {
					env.AdditionalMainCatalog = append(env.AdditionalMainCatalog, r)
				}

			default:
				if !ossrunactions.Services.IsEnabled() {
					return
				}
				comparableName := MakeComparableName(r.Name)
				si, _ := LookupService(comparableName, true)
				if !si.HasSourceMainCatalog() {
					si.SourceMainCatalog = *r
				} else {
					si.AdditionalMainCatalog = append(si.AdditionalMainCatalog, r)
				}
				if options.GlobalOptions().RefreshPricing {
					RecordPricingInfo(si, r)
				} else if options.GlobalOptions().IncludePricing && (!si.HasPriorOSS() || len(si.GetPriorOSS().ProductInfo.PartNumbers) == 0) {
					RecordPricingInfo(si, r)
				}
				if prior, found := servicesByCatalogID[r.ID]; !found {
					debug.Debug(debug.Merge, "ossmerge.LoadAllEntries.MainCatalog: recording entry %s -- ParentID=%q service=%s", r.String(), r.ParentID, si.String())
					servicesByCatalogID[r.ID] = si
				} else {
					debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: Found duplicate Catalog ID=%q for Main resource: %s <-> %s", r.ID, r.String(), prior.SourceMainCatalog.String())
				}
				if r.ParentID != "" {
					if parentSI, found := servicesByCatalogID[r.ParentID]; found {
						parentCX := parentSI.GetCatalogExtra(true)
						parentCX.ChildrenKinds.Add(r.Kind)
					} else {
						debug.PrintError("ossmerge.LoadAllEntries.MainCatalog: No service found for ParentID=%q  -- child=%s", r.ParentID, r.String())
					}
				}
			}
		})
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error listing Main entries from Global Catalog"))
		}
	}

	if ossrunactions.Services.IsEnabled() && ossrunactions.ScorecardV1.IsEnabled() {
		err = scorecardv1.ListScorecardV1Details(pattern, func(e *scorecardv1.DetailEntry) {
			comparableName := MakeComparableName(e.Name)
			si, _ := LookupService(comparableName, true)
			if !si.HasSourceScorecardV1Detail() {
				si.SourceScorecardV1Detail = *e
			} else {
				si.AdditionalScorecardV1Detail = append(si.AdditionalScorecardV1Detail, e)
			}
		})
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error listing Detail entries from ScorecardV1"))
		}
	} else {
		debug.Info("Skip reloading Services info from ScorecardV1")
	}

	if ossrunactions.Services.IsEnabled() {
		err = servicenow.ListServiceNowRecords(pattern, false, servicenow.WATSON, func(e *servicenow.ConfigurationItem, issues []*ossvalidation.ValidationIssue) {
			comparableName := MakeComparableName(e.CRNServiceName)
			si, _ := LookupService(comparableName, true)
			si.normalizeServiceNowEntry(e)
			if !si.HasSourceServiceNow() {
				si.SourceServiceNow = *e
			} else {
				si.AdditionalServiceNow = append(si.AdditionalServiceNow, e)
			}
			if len(issues) > 0 {
				for _, v := range issues {
					si.OSSValidation.AddIssuePreallocated(v)
				}
			}
		})
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error listing entries from ServiceNow"))
		}

		// Override the ServiceNow entries from API with entries from the csv import file, and find any missing ServiceNow entries
		// This must happen after we have loaded from all other sources, so that we can accurately report
		// if an entry from the ServiceNow import file is not found anywhere else
		err = processServiceNowImport(pattern)
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error processing the ServiceNow csv import file"))
		}

		if options.GlobalOptions().TestMode {
			err = servicenow.ListServiceNowRecords(pattern, true, servicenow.WATSON, func(e *servicenow.ConfigurationItem, issues []*ossvalidation.ValidationIssue) {
				comparableName := MakeComparableName(e.CRNServiceName)
				si, _ := LookupService(comparableName, true)
				si.normalizeServiceNowEntry(e)
				displayName := e.DisplayName
				dummyTags := osstags.TagSet{}
				if !si.HasSourceServiceNow() {
					if si.HasPriorOSS() || osstags.CheckOSSTestTag(&displayName, &dummyTags) { // XXX Do not create new test records for everything in ServiceNow -- only items that already have a (possibly empty) OSS record
						si.SourceServiceNow = *e
						if si.OSSMergeControl == nil {
							si.OSSMergeControl = ossmergecontrol.New("")
						}
						si.OSSMergeControl.OSSTags.AddTag(osstags.OSSTest)
						si.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSTest)
						debug.Info("Test Mode: loading test entry from ServiceNow: %s", e.CRNServiceName)
						si.AddValidationIssue(ossvalidation.INFO, "Entry loaded from test instance of ServiceNow", "").TagTest()
					} else {
						debug.Info("Test Mode: ignoring test entry from ServiceNow that does not have a prior OSS entry: %s", e.CRNServiceName)
						si.AddValidationIssue(ossvalidation.INFO, "Ignoring test entry from ServiceNow that does not have a prior OSS entry", "").TagTest()
					}
				}
				if len(issues) > 0 {
					for _, v := range issues {
						si.OSSValidation.AddIssuePreallocated(v)
					}
				}
			})
			if err != nil {
				return (debug.WrapError(err, "(Test Mode): LoadAllEntries(): error listing entries from ServiceNow"))
			}
		}
	}

	if ossrunactions.Services.IsEnabled() {
		err = legacy.ListLegacyRecords(pattern, func(e *legacy.Entry) {
			comparableName := MakeComparableName(string(e.Name))
			si, _ := LookupService(comparableName, true)
			if si.Legacy == nil {
				si.Legacy = e
			} else {
				debug.PrintError(`LoadAllEntries(): found duplicate legacy entry "%s`, e.Name)
			}
		})
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error listing legacy entries"))
		}
	}

	if ossrunactions.Services.IsEnabled() {
		err = iam.ListIAMServices(pattern, func(e *iam.Service) {
			comparableName := MakeComparableName(e.Name)
			si, _ := LookupService(comparableName, true)
			if !si.HasSourceIAM() {
				si.SourceIAM = *e
			} else {
				si.AdditionalIAM = append(si.AdditionalIAM, e)
			}
		})
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error listing Service entries from IAM"))
		}
	}

	if ossrunactions.Monitoring.IsEnabled() {
		err := monitoringinfo.LoadAllServices(serviceInfoRegistrySingleton, pattern)
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error listing Monitoring information"))
		}
	}

	if supportcenter.HasCandidatesInputFile() {
		err := supportcenter.LoadSupportCenterCandidates(serviceInfoRegistrySingleton)
		if err != nil {
			return (debug.WrapError(err, "LoadAllEntries(): error listing Support Center candidates"))
		}
	}

	err = processDuplicateNames()
	if err != nil {
		return (debug.WrapError(err, "LoadAllEntries(): error processing duplicates"))
	}

	return nil
}

// LoadOneService loads the records from various sources from a single service/component, into the Model,
// in preparation for merging.
func LoadOneService(name string) (*ServiceInfo, error) {
	// TODO: Should pass errors back to the caller

	serviceName := ossrecord.CRNServiceName(name)

	// Create an empty ServiceInfo record
	var si *ServiceInfo

	// Load all available records from multiple sources; include service domain overrides to get the full service record
	entry, err := catalog.ReadOSSEntryByID(ossrecord.MakeOSSServiceID(serviceName), catalog.IncludeAll|catalog.IncludeServicesDomainOverrides)
	if err != nil {
		if rest.IsEntryNotFound(err) {
			debug.Info("No OSS Catalog entry found for \"%s\": %v", serviceName, err)
		} else {
			debug.PrintError("Error reading OSS Catalog entry for \"%s\": %v", serviceName, err)
		}
		comparableName := MakeComparableName(name)
		si, _ = LookupService(comparableName, true)
	} else {
		if ossrec, ok := entry.(*ossrecordextended.OSSServiceExtended); ok {
			if len(ossrec.OSSMergeControl.DoNotMergeNames) > 0 {
				registerDoNotMergeNames(ossrec.OSSMergeControl, ossrec.OSSMergeControl.DoNotMergeNames...)
			}
			comparableName := MakeComparableName(name)
			si, _ = LookupService(comparableName, true)
			si.PriorOSS = ossrec.OSSService
			si.OSSMergeControl = ossrec.OSSMergeControl
			si.PriorOSSValidationChecksum = ossrec.OSSValidation.Checksum()
			si.PriorOSSValidation = ossrec.OSSValidation
		} else {
			panic(fmt.Sprintf("ossmerge.LoadOneService(%s) returne unexpected type %T   %s", serviceName, entry, entry))
		}
	}
	cat, err := catalog.ReadMainCatalogEntry(serviceName)
	if err != nil {
		if rest.IsEntryNotFound(err) {
			debug.Info("No Main Catalog entry found for \"%s\": %v", serviceName, err)
		} else {
			debug.PrintError("Error reading Main Catalog entry for \"%s\": %v", serviceName, err)
		}
	} else {
		si.SourceMainCatalog = *cat
	}
	sn, issues, err := servicenow.ReadServiceNowRecord(serviceName, false, servicenow.WATSON)
	if err != nil {
		if rest.IsEntryNotFound(err) {
			debug.Info("No ServiceNow entry found for \"%s\": %v", serviceName, err)
		} else {
			debug.PrintError("Error reading ServiceNow entry for \"%s\": %v", serviceName, err)
		}
	} else {
		si.normalizeServiceNowEntry(sn)
		si.SourceServiceNow = *sn
		if len(issues) > 0 {
			for _, v := range issues {
				si.OSSValidation.AddIssuePreallocated(v)
			}
		}
	}
	if ossrunactions.ScorecardV1.IsEnabled() {
		sc, err := scorecardv1.ReadScorecardV1Detail(serviceName)
		if err != nil {
			if rest.IsEntryNotFound(err) {
				debug.Info("No ScorecardV1 entry found for \"%s\": %v", serviceName, err)
			} else {
				debug.PrintError("Error reading ScorecardV1 entry for \"%s\": %v", serviceName, err)
			}
		} else {
			si.SourceScorecardV1Detail = *sc
		}
	} else {
		debug.Info("Skip reloading Services info from ScorecardV1")
	}

	return si, nil
}

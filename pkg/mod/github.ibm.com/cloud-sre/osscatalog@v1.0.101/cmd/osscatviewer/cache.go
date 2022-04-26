package main

// Cache for all records loaded in the most recent refresh

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalog"
	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

var allRecords AllRecordsType

const (
	diffLabelStagingOnly    = "staging-only"
	diffLabelProductionOnly = "production-only"
	diffLabelDiffs          = "diffs"
)

// AllRecordsType is a cache for all records loaded in the most recent refresh
type AllRecordsType struct {
	updateMutex     sync.Mutex
	readMutex       sync.Mutex
	services        []*cachedService
	segments        []*cachedSegment
	environments    []*cachedEnvironment
	servicesMap     map[ossrecord.CRNServiceName]*cachedService
	segmentsMap     map[ossrecord.SegmentID]*cachedSegment
	tribesMap       map[ossrecord.TribeID]*cachedTribe
	environmentsMap map[ossrecord.EnvironmentID]*cachedEnvironment
	lastRefresh     time.Time
}

type cachedService struct {
	FilterString            string
	ReferenceResourceName   ossrecord.CRNServiceName
	ReferenceDisplayName    string
	EntryType               ossrecord.EntryType
	OperationalStatusString string
	OSSCRNStatus            osstags.Tag
	OSSOverallStatus        osstags.Tag
	OSSOnboardingPhase      string
	OSSTagsString           string
	OSSMergeTagsString      string
	OSSMergeNotes           string
	OSSMergeLastUpdate      string
	TribeID                 ossrecord.TribeID
	ProductionDiffID        ossrecord.CatalogID
	ProductionDiffLabel     string
	Parent                  ossrecord.CRNServiceName
	Children                []ossrecord.CRNServiceName
	//	Record       *catalog.OSSServiceExtended
}

type cachedSegment struct {
	SegmentID           ossrecord.SegmentID
	SegmentType         ossrecord.SegmentType
	DisplayName         string
	Tribes              []*cachedTribe
	NumServices         int
	OSSOnboardingPhase  ossrecord.OSSOnboardingPhase
	OSSTagsString       string
	IsPlaceholder       bool
	ProductionDiffID    ossrecord.CatalogID
	ProductionDiffLabel string
}

type cachedTribe struct {
	TribeID             ossrecord.TribeID
	DisplayName         string
	SegmentID           ossrecord.SegmentID
	Services            []*cachedService
	OSSOnboardingPhase  ossrecord.OSSOnboardingPhase
	OSSTagsString       string
	IsPlaceholder       bool
	ProductionDiffID    ossrecord.CatalogID
	ProductionDiffLabel string
	// Segment     *cachedSegment
	// SegmentName string
}

type cachedEnvironment struct {
	EnvironmentID       ossrecord.EnvironmentID
	Type                ossrecord.EnvironmentType
	Status              ossrecord.EnvironmentStatus
	DisplayName         string
	OSSOnboardingPhase  ossrecord.OSSOnboardingPhase
	OSSTagsString       string
	IsPlaceholder       bool
	ProductionDiffID    ossrecord.CatalogID
	ProductionDiffLabel string
}

const unknownSegmentName = "* UNKNOWN SEGMENT *"
const unknownTribeName = "* UNKNOWN TRIBE *"

const missingSegmentName = "* MISSING SEGMENT *"
const missingSegmentID ossrecord.SegmentID = "<empty>"
const missingTribeName = "* MISSING TRIBE *"
const missingTribeID ossrecord.TribeID = "<empty>"

var allPattern = regexp.MustCompile(".*")

// refresh refreshes all the cached data about OSS entries
func (ar *AllRecordsType) refresh() error {
	ar.updateMutex.Lock()
	defer ar.updateMutex.Unlock()

	now := time.Now()
	if now.Sub(ar.lastRefresh) < 5*time.Minute {
		debug.Audit("Refreshing OSS data from Global Catalog - less than 5 min since last refresh - returning")
		return nil
	}
	debug.Info("Refreshing OSS data from Global Catalog - starting")

	var numEntriesStaging int
	var numEntriesProduction int

	var numServices int
	var numSegments int
	var numTribes int
	var numEnvironments int

	// Read through all OSS entries in the Production catalog and cache them temporarily,
	// to compute differences with the Staging Catalog
	var allProduction map[ossrecord.OSSEntryID]ossrecord.OSSEntry
	if !*noDiffs {
		allProduction = make(map[ossrecord.OSSEntryID]ossrecord.OSSEntry)
		err := catalog.ListOSSEntriesProduction(allPattern, catalog.IncludeNone,
			func(e ossrecord.OSSEntry) {
				numEntriesProduction++
				e.CleanEntryForCompare()
				allProduction[e.GetOSSEntryID()] = e
			})
		if err != nil {
			debug.PrintError("Refreshing OSS data from Production Global Catalog - error: %v", err)
			return err
		}
	}

	// First read through all OSS entries in the catalog
	newServicesMap := make(map[ossrecord.CRNServiceName]*cachedService)
	newSegmentsMap := make(map[ossrecord.SegmentID]*cachedSegment)
	newTribesMap := make(map[ossrecord.TribeID]*cachedTribe)
	newEnvironmentsMap := make(map[ossrecord.EnvironmentID]*cachedEnvironment)
	err := catalog.ListOSSEntries(allPattern, catalog.IncludeServices|catalog.IncludeTribes|catalog.IncludeEnvironments|catalog.IncludeOSSMergeControl, // Do not include OSSValidation, for performance
		func(e ossrecord.OSSEntry) {
			numEntriesStaging++
			switch r := e.(type) {
			case *ossrecordextended.OSSServiceExtended:
				numServices++
				filterString := makeFilterString(&r.OSSService)
				service := &cachedService{
					FilterString:          filterString,
					ReferenceResourceName: r.OSSService.ReferenceResourceName,
					ReferenceDisplayName:  r.OSSService.ReferenceDisplayName,
					EntryType:             r.OSSService.GeneralInfo.EntryType,
					//					OperationalStatus:     r.OSSService.GeneralInfo.OperationalStatus,
					OSSCRNStatus:       r.OSSService.GeneralInfo.OSSTags.GetCRNStatus(),
					OSSOverallStatus:   r.OSSService.GeneralInfo.OSSTags.GetOverallStatus(),
					OSSOnboardingPhase: string(r.OSSService.GeneralInfo.OSSOnboardingPhase),
					OSSMergeNotes:      r.OSSMergeControl.Notes,
					OSSMergeLastUpdate: r.OSSMergeControl.LastUpdate,
					TribeID:            r.OSSService.Ownership.TribeID,
					Parent:             r.OSSService.GeneralInfo.ParentResourceName,
				}
				if r.OSSService.GeneralInfo.FutureOperationalStatus != "" && r.OSSService.GeneralInfo.FutureOperationalStatus != r.OSSService.GeneralInfo.OperationalStatus {
					service.OperationalStatusString = fmt.Sprintf("%.18s",
						fmt.Sprintf("%s->%s", r.OSSService.GeneralInfo.OperationalStatus.ShortString(), r.OSSService.GeneralInfo.FutureOperationalStatus.ShortString()))
				} else {
					service.OperationalStatusString = fmt.Sprintf("%.18s", r.OSSService.GeneralInfo.OperationalStatus.ShortString())
				}
				combinedTags := r.OSSService.GeneralInfo.OSSTags.WithoutStatus()
				if r.OSSService.GeneralInfo.OSSTags.Contains(osstags.PnPEnabled) {
					combinedTags.AddTag(osstags.PnPEnabled) // OK to simply add here because osstags.WithoutStatus() above already made a copy of the TagSet
				}
				if r.OSSMergeControl.LastUpdate != "" && r.OSSMergeControl.LastUpdate > r.Updated {
					service.OSSTagsString = ""
					service.OSSMergeTagsString = combinedTags.String()
				} else {
					service.OSSTagsString = combinedTags.String()
					service.OSSMergeTagsString = ""
				}
				if !*noDiffs {
					entryID := r.GetOSSEntryID()
					if prod, found := allProduction[entryID]; found {
						r.CleanEntryForCompare()
						out := compare.Output{}
						compare.DeepCompare("", &r.OSSService, "", prod, &out)
						if out.NumDiffs() > 0 {
							service.ProductionDiffID = ossrecord.CatalogID(entryID)
							service.ProductionDiffLabel = diffLabelDiffs
						}
						delete(allProduction, entryID)
					} else {
						service.ProductionDiffID = ossrecord.CatalogID(entryID)
						service.ProductionDiffLabel = diffLabelStagingOnly
					}
				}
				newServicesMap[service.ReferenceResourceName] = service
			case *ossrecordextended.OSSSegmentExtended:
				numSegments++
				if seg, found := newSegmentsMap[r.SegmentID]; !found {
					seg = &cachedSegment{
						SegmentID:          r.SegmentID,
						SegmentType:        r.SegmentType,
						DisplayName:        r.DisplayName,
						Tribes:             make([]*cachedTribe, 0, 5),
						OSSOnboardingPhase: r.OSSOnboardingPhase,
						OSSTagsString:      r.OSSTags.String(),
					}
					if !*noDiffs {
						entryID := r.GetOSSEntryID()
						if prod, found := allProduction[entryID]; found {
							r.CleanEntryForCompare()
							out := compare.Output{}
							compare.DeepCompare("", &r.OSSSegment, "", prod, &out)
							if out.NumDiffs() > 0 {
								seg.ProductionDiffID = ossrecord.CatalogID(entryID)
								seg.ProductionDiffLabel = diffLabelDiffs
							}
							delete(allProduction, entryID)
						} else {
							seg.ProductionDiffID = ossrecord.CatalogID(entryID)
							seg.ProductionDiffLabel = diffLabelStagingOnly
						}
					}
					newSegmentsMap[r.SegmentID] = seg
				} else {
					debug.PrintError(`Found duplicate segment id: Segment(%s[%s]) / Segment(%s[%s])`, seg.DisplayName, seg.SegmentID, r.DisplayName, r.SegmentID)
				}
			case *ossrecordextended.OSSTribeExtended:
				numTribes++
				if tr, found := newTribesMap[r.TribeID]; !found {
					tr := &cachedTribe{
						TribeID:     r.TribeID,
						DisplayName: r.DisplayName,
						SegmentID:   r.SegmentID,
						// SegmentName: unknownSegmentName,
						Services:           make([]*cachedService, 0, 10),
						OSSOnboardingPhase: r.OSSOnboardingPhase,
						OSSTagsString:      r.OSSTags.String(),
					}
					if !*noDiffs {
						entryID := r.GetOSSEntryID()
						if prod, found := allProduction[entryID]; found {
							r.CleanEntryForCompare()
							out := compare.Output{}
							compare.DeepCompare("", &r.OSSTribe, "", prod, &out)
							if out.NumDiffs() > 0 {
								tr.ProductionDiffID = ossrecord.CatalogID(entryID)
								tr.ProductionDiffLabel = diffLabelDiffs
							}
							delete(allProduction, entryID)
						} else {
							tr.ProductionDiffID = ossrecord.CatalogID(entryID)
							tr.ProductionDiffLabel = diffLabelStagingOnly
						}
					}
					newTribesMap[r.TribeID] = tr
				} else {
					debug.PrintError(`Found duplicate tribe id: Tribe(%s[%s]) / Tribe(%s[%s])`, tr.DisplayName, tr.TribeID, r.DisplayName, r.TribeID)
				}
			case *ossrecordextended.OSSEnvironmentExtended:
				numEnvironments++
				if env, found := newEnvironmentsMap[r.OSSEnvironment.EnvironmentID]; !found {
					env = &cachedEnvironment{
						EnvironmentID:      r.OSSEnvironment.EnvironmentID,
						Type:               r.OSSEnvironment.Type,
						Status:             r.OSSEnvironment.Status,
						DisplayName:        r.OSSEnvironment.DisplayName,
						OSSOnboardingPhase: r.OSSEnvironment.OSSOnboardingPhase,
					}
					if r.OSSEnvironment.OSSTags.Contains(osstags.IBMCloudDefaultSegment) {
						strippedTags := r.OSSEnvironment.OSSTags.Copy()
						strippedTags.RemoveTag(osstags.IBMCloudDefaultSegment)
						env.OSSTagsString = strippedTags.String()
					} else {
						env.OSSTagsString = r.OSSEnvironment.OSSTags.String()
					}
					if !*noDiffs {
						entryID := r.GetOSSEntryID()
						if prod, found := allProduction[entryID]; found {
							r.CleanEntryForCompare()
							out := compare.Output{}
							compare.DeepCompare("", &r.OSSEnvironment, "", prod, &out)
							if out.NumDiffs() > 0 {
								env.ProductionDiffID = ossrecord.CatalogID(entryID)
								env.ProductionDiffLabel = diffLabelDiffs
							}
							delete(allProduction, entryID)
						} else {
							env.ProductionDiffID = ossrecord.CatalogID(entryID)
							env.ProductionDiffLabel = diffLabelStagingOnly
						}
					}
					newEnvironmentsMap[r.EnvironmentID] = env
				} else {
					debug.PrintError(`Found duplicate environment id: Environment(%s[%s]) / Environment(%s[%s])`, env.DisplayName, env.EnvironmentID, r.OSSEnvironment.DisplayName, r.OSSEnvironment.EnvironmentID)
				}
			default:
				panic(fmt.Sprintf("Ignoring unknown entry type from from Staging Catalog %T   %v", e, e))
				// ignore all other entries
			}
		})
	ar.lastRefresh = time.Now()
	if err != nil {
		debug.PrintError("Refreshing OSS data from Staging Global Catalog - error: %v", err)
		return err
	}

	// Check for entries found only in Production but not in Staging
	if !*noDiffs {
		for id, prod := range allProduction {
			switch r := prod.(type) {
			case *ossrecord.OSSService:
				filterString := makeFilterString(r)
				service := &cachedService{
					FilterString:          filterString,
					ReferenceResourceName: r.ReferenceResourceName,
					ReferenceDisplayName:  "PRODUCTION ONLY: " + r.ReferenceDisplayName,
					TribeID:               r.Ownership.TribeID,
				}
				service.ProductionDiffID = ossrecord.CatalogID(id)
				service.ProductionDiffLabel = diffLabelProductionOnly
				newServicesMap[r.ReferenceResourceName] = service
			case *ossrecord.OSSSegment:
				if seg, found := newSegmentsMap[r.SegmentID]; !found {
					seg = &cachedSegment{
						SegmentID:   r.SegmentID,
						SegmentType: r.SegmentType,
						DisplayName: "PRODUCTION ONLY: " + r.DisplayName,
						Tribes:      make([]*cachedTribe, 0, 5),
					}
					seg.ProductionDiffID = ossrecord.CatalogID(id)
					seg.ProductionDiffLabel = diffLabelProductionOnly
					newSegmentsMap[r.SegmentID] = seg
				} else {
					debug.PrintError(`Found duplicate segment id in Production: Segment(%s[%s]) / Segment(%s[%s])`, seg.DisplayName, seg.SegmentID, r.DisplayName, r.SegmentID)
				}
			case *ossrecord.OSSTribe:
				if tr, found := newTribesMap[r.TribeID]; !found {
					tr := &cachedTribe{
						TribeID:     r.TribeID,
						DisplayName: "PRODUCTION ONLY: " + r.DisplayName,
						SegmentID:   r.SegmentID,
						// SegmentName: unknownSegmentName,
						Services: make([]*cachedService, 0, 10),
					}
					tr.ProductionDiffID = ossrecord.CatalogID(id)
					tr.ProductionDiffLabel = diffLabelProductionOnly
					newTribesMap[r.TribeID] = tr
				} else {
					debug.PrintError(`Found duplicate tribe id in Production: Tribe(%s[%s]) / Tribe(%s[%s])`, tr.DisplayName, tr.TribeID, r.DisplayName, r.TribeID)
				}
			case *ossrecord.OSSEnvironment:
				if env, found := newEnvironmentsMap[r.EnvironmentID]; !found {
					env = &cachedEnvironment{
						EnvironmentID: r.EnvironmentID,
						Type:          r.Type,
						Status:        r.Status,
						DisplayName:   "PRODUCTION ONLY: " + r.DisplayName,
					}
					env.ProductionDiffID = ossrecord.CatalogID(id)
					env.ProductionDiffLabel = diffLabelProductionOnly
					newEnvironmentsMap[r.EnvironmentID] = env
				} else {
					debug.PrintError(`Found duplicate environment id in Production: Environment(%s[%s]) / Environment(%s[%s])`, env.DisplayName, env.EnvironmentID, r.DisplayName, r.EnvironmentID)
				}
			default:
				panic(fmt.Sprintf("Ignoring unknown entry type from from Production Catalog %T   %v", prod, prod))
				// ignore all other entries
			}
		}
	}

	// Setup all the links between segment, tribe and service records in the cache
	linkAllCachedRecords(newServicesMap, newSegmentsMap, newTribesMap)

	// Created sorted slices before loading them into the cache
	newServices := make([]*cachedService, 0, len(newServicesMap))
	for _, svc := range newServicesMap {
		newServices = append(newServices, svc)
		if svc.Children != nil {
			sort.Slice(svc.Children, func(i, j int) bool {
				return svc.Children[i] < svc.Children[j]
			})
		}
	}
	sort.Slice(newServices, func(i, j int) bool {
		return newServices[i].ReferenceResourceName < newServices[j].ReferenceResourceName
	})
	debug.Info("Total services/components: %d/%d", numServices, len(newServices))
	newSegments := make([]*cachedSegment, 0, len(newSegmentsMap))
	for _, seg := range newSegmentsMap {
		newSegments = append(newSegments, seg)
		sort.Slice(seg.Tribes, func(i, j int) bool {
			return seg.Tribes[i].DisplayName < seg.Tribes[j].DisplayName
		})
	}
	sort.Slice(newSegments, func(i, j int) bool {
		return newSegments[i].DisplayName < newSegments[j].DisplayName
	})
	debug.Info("Total segments: %d/%d", numSegments, len(newSegments))
	for _, tr := range newTribesMap {
		sort.Slice(tr.Services, func(i, j int) bool {
			return tr.Services[i].ReferenceResourceName < tr.Services[j].ReferenceResourceName
		})
	}
	newEnvironments := make([]*cachedEnvironment, 0, len(newEnvironmentsMap))
	for _, env := range newEnvironmentsMap {
		newEnvironments = append(newEnvironments, env)
	}
	sort.Slice(newEnvironments, func(i, j int) bool {
		return newEnvironments[i].EnvironmentID < newEnvironments[j].EnvironmentID
	})
	debug.Info("Total environments: %d/%d", numEnvironments, len(newEnvironments))

	// Load the data into the cache
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	// Note that we MUST assign all news slices/maps, because we may have returned the old slices/maps
	// to some callers, who no longer hold any mutex
	ar.services = newServices
	ar.segments = newSegments
	ar.servicesMap = newServicesMap
	ar.segmentsMap = newSegmentsMap
	ar.tribesMap = newTribesMap
	ar.environments = newEnvironments
	ar.environmentsMap = newEnvironmentsMap
	debug.Audit("Refreshing OSS data from Global Catalog - done. Number of entries: Production=%d / Staging=%d", numEntriesProduction, numEntriesStaging)
	return nil
}

// linkAllCachedRecords initializes all the links between segment, tribe and service records in the cache
func linkAllCachedRecords(
	newServicesMap map[ossrecord.CRNServiceName]*cachedService,
	newSegmentsMap map[ossrecord.SegmentID]*cachedSegment,
	newTribesMap map[ossrecord.TribeID]*cachedTribe) {

	// Scan through all service/component records and assign to tribes; also check for parent references
	for _, svc := range newServicesMap {
		var tribeID ossrecord.TribeID
		var tr *cachedTribe
		var found bool
		if svc.TribeID == "" {
			// debug.PrintError(`Found a service with empty TribeID: "%s"`, svc.ReferenceResourceName) // not necessarily an error - e.g. for third-party services
			tribeID = missingTribeID
		} else {
			tribeID = svc.TribeID
		}
		if tr, found = newTribesMap[tribeID]; !found {
			tr = &cachedTribe{
				TribeID:   tribeID,
				SegmentID: missingSegmentID,
				// SegmentName: unknownSegmentName,
				Services:      make([]*cachedService, 0, 10),
				IsPlaceholder: true,
			}
			if tribeID == missingTribeID {
				tr.DisplayName = missingTribeName
			} else {
				debug.PrintError(`Found a service with unknown TribeID: "%s" - Tribe([%s])`, svc.ReferenceResourceName, svc.TribeID)
				tr.DisplayName = unknownTribeName
			}
			debug.Debug(debug.Main, `Creating placeholder Tribe cache entry %#v for Service "%s`, tr, svc.ReferenceResourceName)
			newTribesMap[tribeID] = tr
		}
		tr.Services = append(tr.Services, svc)

		if svc.Parent != "" {
			if parent, ok := newServicesMap[svc.Parent]; ok {
				parent.Children = append(parent.Children, svc.ReferenceResourceName)
			}
		}
	}

	// Link the tribes to their parent segments
	for _, tr := range newTribesMap {
		var segmentID ossrecord.SegmentID
		var seg *cachedSegment
		var found bool
		if tr.SegmentID == "" {
			debug.PrintError(`Found a tribe with empty SegmentID: Tribe(%s[%s])`, tr.DisplayName, tr.TribeID)
			segmentID = missingSegmentID
		} else {
			segmentID = tr.SegmentID
		}
		if seg, found = newSegmentsMap[segmentID]; !found {
			seg = &cachedSegment{
				SegmentID:     tr.SegmentID,
				Tribes:        make([]*cachedTribe, 0, 5),
				IsPlaceholder: true,
			}
			if segmentID == missingSegmentID {
				seg.DisplayName = missingSegmentName
			} else {
				debug.PrintError(`Found a tribe with unknown SegmentID: Tribe(%s[%s]) - Segment([%s]))`, tr.DisplayName, tr.TribeID, tr.SegmentID)
				seg.DisplayName = unknownSegmentName
			}
			debug.Debug(debug.Main, `Creating placeholder Segment cache entry %#v for Tribe "%s`, seg, tr.DisplayName)
			newSegmentsMap[tr.SegmentID] = seg
		}
		//		tr.SegmentName = seg.DisplayName
		//		tr.Segment = seg
		seg.Tribes = append(seg.Tribes, tr)
		seg.NumServices += len(tr.Services)
	}
}

// getAllServices returns a slice of all service entries in the cache
func (ar *AllRecordsType) getAllServices() (services []*cachedService, lastRefresh time.Time) {
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	services = ar.services
	lastRefresh = ar.lastRefresh
	return services, lastRefresh
}

// getAllSegments returns a slice of all segment entries in the cache
func (ar *AllRecordsType) getAllSegments() (segments []*cachedSegment, lastRefresh time.Time) {
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	segments = ar.segments
	lastRefresh = ar.lastRefresh
	return segments, lastRefresh
}

// getSegmentName returns the name associated with a given SegmentID
func (ar *AllRecordsType) getSegmentName(id ossrecord.SegmentID) string {
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	if seg, found := ar.segmentsMap[id]; found {
		return seg.DisplayName
	}
	return unknownSegmentName
}

// getServicesForTribeID returns a slice with all the service entries that belong to a given TribeID
func (ar *AllRecordsType) getServicesForTribeID(id ossrecord.TribeID) (services []*cachedService, lastRefresh time.Time) {
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	if tr, found := ar.tribesMap[id]; found {
		return tr.Services, ar.lastRefresh
	}
	return nil, ar.lastRefresh
}

// getTribesForSegmentID returns a map of all the Tribes associated with a given SegmentID
func (ar *AllRecordsType) getTribesForSegmentID(id ossrecord.SegmentID) (tribes []*cachedTribe, lastRefresh time.Time) {
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	if seg, found := ar.segmentsMap[id]; found {
		return seg.Tribes, ar.lastRefresh
	}
	return nil, ar.lastRefresh
}

// getPlaceHolderTribe returns a placeholder OSSTribeExtended record from the cache, if there is one, instead a real OSSTribe record found in the Catalog
func (ar *AllRecordsType) getPlaceHolderTribe(id ossrecord.TribeID) *ossrecordextended.OSSTribeExtended {
	if e, found := ar.tribesMap[id]; found && e.IsPlaceholder {
		tr := &ossrecordextended.OSSTribeExtended{
			OSSTribe: ossrecord.OSSTribe{
				TribeID:     e.TribeID,
				DisplayName: e.DisplayName,
			},
		}
		return tr
	}
	return nil
}

// getPlaceHolderSegment returns a placeholder OSSSegmentExtended record from the cache, if there is one, instead a real OSSSegment record found in the Catalog
func (ar *AllRecordsType) getPlaceHolderSegment(id ossrecord.SegmentID) *ossrecordextended.OSSSegmentExtended {
	if e, found := ar.segmentsMap[id]; found && e.IsPlaceholder {
		seg := &ossrecordextended.OSSSegmentExtended{
			OSSSegment: ossrecord.OSSSegment{
				SegmentID:   e.SegmentID,
				DisplayName: e.DisplayName,
			},
		}
		return seg
	}
	return nil
}

// getServiceChildren returns a slice with all the service names that are children of the given service (as defined by the ParentResourceName attribute)
func (ar *AllRecordsType) getServiceChildren(name ossrecord.CRNServiceName) (children []ossrecord.CRNServiceName, lastRefresh time.Time) {
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	if svc, found := ar.servicesMap[name]; found {
		return svc.Children, ar.lastRefresh
	}
	return nil, ar.lastRefresh
}

// getAllEnvironments returns a slice of all environment entries in the cache
func (ar *AllRecordsType) getAllEnvironments() (environments []*cachedEnvironment, lastRefresh time.Time) {
	ar.readMutex.Lock()
	defer ar.readMutex.Unlock()
	environments = ar.environments
	lastRefresh = ar.lastRefresh
	return environments, lastRefresh
}

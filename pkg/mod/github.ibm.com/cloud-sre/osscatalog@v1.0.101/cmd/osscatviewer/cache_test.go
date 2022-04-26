package main

import (
	"fmt"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
)

// createTestCache creates and populates a dummy cache, to use in other tests
func createTestCache() *AllRecordsType {
	ar := &AllRecordsType{
		services:        make([]*cachedService, 0, 10),
		segments:        make([]*cachedSegment, 0, 10),
		segmentsMap:     make(map[ossrecord.SegmentID]*cachedSegment),
		tribesMap:       make(map[ossrecord.TribeID]*cachedTribe),
		environmentsMap: make(map[ossrecord.EnvironmentID]*cachedEnvironment),
		environments:    make([]*cachedEnvironment, 0, 10),
		lastRefresh:     time.Now(),
	}

	newServices := make([]*cachedService, 0, 10)
	newSegments := make([]*cachedSegment, 0, 10)
	newServicesMap := make(map[ossrecord.CRNServiceName]*cachedService)
	newSegmentsMap := make(map[ossrecord.SegmentID]*cachedSegment)
	newTribesMap := make(map[ossrecord.TribeID]*cachedTribe)
	newEnvironments := make([]*cachedEnvironment, 0, 10)
	newEnvironmentsMap := make(map[ossrecord.EnvironmentID]*cachedEnvironment)

	var testSegmentID ossrecord.SegmentID
	var testTribeID ossrecord.TribeID

	for i := 1; i <= 3; i++ {
		seg := &cachedSegment{
			SegmentID:   ossrecord.SegmentID(fmt.Sprintf("oss_segment.oss-test-%d", i)),
			DisplayName: fmt.Sprintf("Segment Name %d", i),
			Tribes:      make([]*cachedTribe, 0, 5),
		}
		if i == 1 {
			seg.ProductionDiffID = ossrecord.CatalogID(seg.SegmentID)
			seg.ProductionDiffLabel = diffLabelDiffs
		}
		newSegmentsMap[seg.SegmentID] = seg
		newSegments = append(newSegments, seg)
		if testSegmentID == "" {
			testSegmentID = seg.SegmentID
		}
		for j := 1; j <= 3; j++ {
			tr := &cachedTribe{
				TribeID:     ossrecord.TribeID(fmt.Sprintf("oss_tribe.oss-test-%d-%d", i, j)),
				DisplayName: fmt.Sprintf("Tribe Name %d/%d", i, j),
				SegmentID:   seg.SegmentID,
				// SegmentName: unknownSegmentName,
				Services: make([]*cachedService, 0, 10),
			}
			if j == 2 {
				tr.ProductionDiffID = ossrecord.CatalogID(tr.TribeID)
				tr.ProductionDiffLabel = diffLabelStagingOnly
			}
			newTribesMap[tr.TribeID] = tr
			if testTribeID == "" {
				testTribeID = tr.TribeID
			}
		}
	}

	for i := 0; i <= 2; i++ {
		oss := ossrecord.CreateTestRecord()
		oss.Ownership.SegmentID = testSegmentID
		oss.Ownership.TribeID = testTribeID
		baseName := oss.ReferenceResourceName
		r := ossrecordextended.NewOSSServiceExtended(baseName)
		r.OSSService = *oss
		r.OSSService.ReferenceResourceName = ossrecord.CRNServiceName(fmt.Sprintf("%s.%d", baseName, i))
		var productionDiff ossrecord.CatalogID
		switch i {
		case 0:
			r.OSSService.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNGreen)
			r.OSSService.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusGreen)
			r.OSSService.GeneralInfo.OSSTags.AddTag(osstags.NotReady)
			r.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSOnly)
			r.OSSMergeControl.OSSTags.AddTag(osstags.OSSOnly)
			r.OSSMergeControl.OSSTags.AddTag(osstags.NotReady)
			r.OSSMergeControl.OSSTags.AddTag(osstags.OneCloudWave2)
			r.OSSMergeControl.Notes = "Note1"
			r.OSSService.GeneralInfo.ParentResourceName = ossrecord.CRNServiceName(fmt.Sprintf("%s.%d", baseName, 1))
		case 1:
			r.OSSService.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNRed)
			r.OSSService.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusYellow)
			r.OSSService.GeneralInfo.OSSTags.AddTag(osstags.NotReady)
			r.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSOnly)
			r.OSSService.GeneralInfo.EntryType = ossrecord.PLATFORMCOMPONENT
			r.OSSService.GeneralInfo.OperationalStatus = ossrecord.EXPERIMENTAL
			r.OSSMergeControl.Notes = "Note2: Some notes about this record"
			r.OSSMergeControl.LastUpdate = "2006-01-02 15:04 MST"
			productionDiff = ossrecord.CatalogID(r.GetOSSEntryID())
		case 2:
			r.OSSService.GeneralInfo.OSSTags.SetCRNStatus(osstags.StatusCRNYellow)
			r.OSSService.GeneralInfo.OSSTags.SetOverallStatus(osstags.StatusRed)
			r.OSSService.GeneralInfo.OSSTags.AddTag(osstags.NotReady)
			r.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSOnly)
			r.OSSMergeControl.OSSTags.AddTag(osstags.OSSOnly)
			r.OSSMergeControl.OSSTags.AddTag(osstags.OneCloudWave1)
			r.OSSService.GeneralInfo.EntryType = ossrecord.SUBCOMPONENT
			r.OSSService.GeneralInfo.OperationalStatus = ossrecord.SELECTAVAILABILITY
			r.OSSService.GeneralInfo.FutureOperationalStatus = ossrecord.BETA
			r.OSSService.GeneralInfo.OSSOnboardingPhase = ossrecord.PRODUCTION
			r.OSSService.GeneralInfo.ParentResourceName = ossrecord.CRNServiceName(fmt.Sprintf("%s.%d", baseName, 1))
			productionDiff = ossrecord.CatalogID(r.GetOSSEntryID())
		}
		filterString := makeFilterString(&r.OSSService)
		entry := &cachedService{
			FilterString:            filterString,
			ReferenceResourceName:   r.OSSService.ReferenceResourceName,
			ReferenceDisplayName:    r.OSSService.ReferenceDisplayName,
			EntryType:               r.OSSService.GeneralInfo.EntryType,
			OperationalStatusString: fmt.Sprintf("%.18s", r.OSSService.GeneralInfo.OperationalStatus),
			OSSOnboardingPhase:      string(r.GeneralInfo.OSSOnboardingPhase),
			OSSCRNStatus:            r.OSSService.GeneralInfo.OSSTags.GetCRNStatus(),
			OSSOverallStatus:        r.OSSService.GeneralInfo.OSSTags.GetOverallStatus(),
			OSSTagsString:           r.OSSService.GeneralInfo.OSSTags.WithoutStatus().String(),
			OSSMergeNotes:           r.OSSMergeControl.Notes,
			OSSMergeLastUpdate:      r.OSSMergeControl.LastUpdate,
			TribeID:                 r.OSSService.Ownership.TribeID,
			ProductionDiffID:        productionDiff,
			ProductionDiffLabel:     diffLabelProductionOnly,
			Parent:                  r.OSSService.GeneralInfo.ParentResourceName,
		}
		mt := r.OSSMergeControl.OSSTags.WithoutStatus().String()
		if mt != entry.OSSTagsString {
			if mt == "[]" {
				mt = "<removed>"
			}
			entry.OSSMergeTagsString = mt
		}
		newServicesMap[entry.ReferenceResourceName] = entry
		newServices = append(newServices, entry)
	}

	for i := 1; i <= 3; i++ {
		env := &cachedEnvironment{
			EnvironmentID: ossrecord.EnvironmentID(fmt.Sprintf("crn:v1:bluemix:public::dal0%d::::", i)),
			DisplayName:   fmt.Sprintf("Dallas 0%d", i),
			Type:          ossrecord.EnvironmentIBMCloudDatacenter,
			Status:        ossrecord.EnvironmentActive,
			OSSTagsString: (&osstags.TagSet{osstags.IBMCloudDefaultSegment}).String(),
		}
		newEnvironmentsMap[env.EnvironmentID] = env
		newEnvironments = append(newEnvironments, env)
	}

	linkAllCachedRecords(newServicesMap, newSegmentsMap, newTribesMap)

	ar.services = newServices
	ar.segments = newSegments
	ar.servicesMap = newServicesMap
	ar.segmentsMap = newSegmentsMap
	ar.tribesMap = newTribesMap
	ar.environments = newEnvironments
	ar.environmentsMap = newEnvironmentsMap

	return ar
}

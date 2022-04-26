package services4edb

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossreports"
)

// mainRecordType represents one row in the main sheet of the report
type mainRecordType struct {
	CRNServiceName      ossreports.ExcelLink `column:"CRN service-name,15"`
	DisplayName         string               `column:"Display Name,20"`
	EntryType           string               `column:"Type"`
	OperationalStatus   string               `column:"Operational Status,10"`
	OneCloud            string               `column:"OneCloud,6"`
	ClientFacingFlag    string               `column:"Client-facing Flag,6"`
	CatalogClientFacing string               `column:"Client Facing (from Catalog only),7"`
	//	ClientFacingComputed string               `column:"Client-facing (Computed),10"`
	HasEDBProvisioning ossreports.ExcelLink `column:"Has EDB Provisioning Data,10"`
	HasEDBConsumption  ossreports.ExcelLink `column:"Has EDB Consumption Data,10"`
	HasOtherMetrics    ossreports.ExcelLink `column:"Has Other Metrics Data,10"`
	EDBTags            string               `column:"EDB Tags,10"`
	EDBMissing         string               `column:"EDB Missing?,36"`
	Segment            string               `column:"Segment"`
	Tribe              string               `column:"Tribe"`
	CRNStatus          string               `column:"CRN Validation Status,10"`
	Notes              string               `column:"Merge Control Notes,20"`
	OSSOnboardingPhase string               `column:"OSS Onboarding Phase,12"`
	OSSTags            string               `column:"OSS Tags"`
}

// Labels for columns in the spreadsheet
const (
	emptycell = " "
	YES       = "yes"
	NO        = "no"
	EMPTY     = "*EMPTY*"
)

// URLs for tools linked from the report
const osscatviewer = "https://osscatviewer.us-south.cf.test.appdomain.cloud"
const scorecard = "https://cloud.ibm.com/scorecard"

func makeScorecardLink(oss *ossrecord.OSSService) ossreports.ExcelLink {
	if len(oss.Ownership.SegmentName) > 100 || len(oss.Ownership.TribeName) > 100 || strings.Contains(oss.Ownership.SegmentName, "/") || strings.Contains(oss.Ownership.TribeName, "/") {
		return ossreports.ExcelLink{
			Text: YES,
		}
	}
	return ossreports.ExcelLink{
		URL:  fmt.Sprintf("%s/resources/%s/%s/%s", scorecard, url.PathEscape(oss.Ownership.SegmentName), url.PathEscape(oss.Ownership.TribeName), oss.ReferenceResourceName),
		Text: YES,
	}
}

// Services4EDB generates a summary report of EDB metrics enablement per service/component (as an Excel spreadsheet)
func Services4EDB(w io.Writer, timeStamp string, pattern *regexp.Regexp) error {
	var err error
	mainData := make([]interface{}, 0, 500)

	err = ossmerge.ListAllServices(pattern, func(si *ossmerge.ServiceInfo) {
		if !si.IsValid() || si.IsDeletable() {
			return
		}
		if si.OSSService.GeneralInfo.OSSTags.Contains(osstags.OSSTest) && !options.GlobalOptions().TestMode {
			return
		}
		//var issues int
		oss := &si.OSSService

		r := mainRecordType{}
		r.CRNServiceName = ossreports.ExcelLink{
			URL:  fmt.Sprintf("%s/view/%s", osscatviewer, oss.ReferenceResourceName),
			Text: string(oss.ReferenceResourceName),
		}
		if oss.ReferenceDisplayName != "" {
			r.DisplayName = oss.ReferenceDisplayName
		} else {
			r.DisplayName = EMPTY
		}
		r.EntryType = string(oss.GeneralInfo.EntryType)
		r.OperationalStatus = string(oss.GeneralInfo.OperationalStatus)
		r.OneCloud = strconv.FormatBool(oss.GeneralInfo.OSSTags.Contains(osstags.OneCloud) || oss.GeneralInfo.OSSTags.Contains(osstags.OneCloudComponent))
		r.ClientFacingFlag = strconv.FormatBool(oss.GeneralInfo.ClientFacing)
		r.CatalogClientFacing = strconv.FormatBool(oss.CatalogInfo.CatalogClientFacing)
		//		r.ClientFacingComputed = "???"
		var numEDBProvisioning, numEDBConsumption, numOtherMetrics int
		for _, m := range oss.MonitoringInfo.Metrics {
			if m.Type == ossrecord.MetricProvisioning && m.FindTag(ossrecord.MetricTagSourceEDB) != "" {
				numEDBProvisioning++
			} else if m.Type == ossrecord.MetricConsumption && m.FindTag(ossrecord.MetricTagSourceEDB) != "" {
				numEDBConsumption++
			} else {
				numOtherMetrics++
			}
		}
		if numEDBProvisioning > 0 {
			r.HasEDBProvisioning = makeScorecardLink(oss)
		} else {
			r.HasEDBProvisioning = ossreports.ExcelLink{
				Text: NO,
			}
		}
		if numEDBConsumption > 0 {
			r.HasEDBConsumption = makeScorecardLink(oss)
		} else {
			r.HasEDBConsumption = ossreports.ExcelLink{
				Text: NO,
			}
		}
		if numOtherMetrics > 0 {
			r.HasOtherMetrics = makeScorecardLink(oss)
		} else {
			r.HasOtherMetrics = ossreports.ExcelLink{
				Text: NO,
			}
		}

		r.EDBTags = string(oss.GeneralInfo.OSSTags.GetTagByGroup(osstags.GroupEDBControl))

		if si.OSSValidation != nil {
			if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.EDBMissingButRequired); found {
				r.EDBMissing = "Missing: required by edb_include tag"
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.EDBMissingClientFacing); found {
				r.EDBMissing = "Missing: client-facing service or component"
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.EDBMissingClientFacingNonStandard); found {
				r.EDBMissing = "(minor) Missing: client-facing service or component (non-standard type or status)"
				/*
					} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.EDBMissingNotClientFacing); found {
						r.EDBMissing = "(minor) Missing: non-client-facing service or component"
				*/
			} else if _, found := si.OSSValidation.FindNamedIssue(ossvalidation.EDBFoundButExcluded); found {
				r.EDBMissing = "*** Unexpected: excluded by edb_exclude tag"
			} else {
				r.EDBMissing = emptycell
			}
		} else {
			r.EDBMissing = emptycell
		}

		r.Segment = si.Ownership.SegmentName
		r.Tribe = si.Ownership.TribeName
		r.CRNStatus = oss.GeneralInfo.OSSTags.GetCRNStatus().StringStatus()
		if si.OSSMergeControl != nil {
			r.Notes = si.OSSMergeControl.Notes
		}
		r.OSSOnboardingPhase = string(oss.GeneralInfo.OSSOnboardingPhase)
		r.OSSTags = oss.GeneralInfo.OSSTags.String()

		mainData = append(mainData, r)
	})
	if err != nil {
		return err
	}

	sort.Slice(mainData, func(i, j int) bool {
		return mainData[i].(mainRecordType).CRNServiceName.Text < mainData[j].(mainRecordType).CRNServiceName.Text
	})

	xl := ossreports.CreateExcel(w, "Services4EDB")
	err = xl.AddSheet(mainData, 90.0, "Services + Components")
	if err != nil {
		return err
	}

	err = xl.Finalize()
	if err != nil {
		return err
	}

	return nil
}

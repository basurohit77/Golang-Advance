package monitoringinfo

import (
	"fmt"
	"regexp"
	"sort"

	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrunactions"
	"github.ibm.com/cloud-sre/osscatalog/osstags"
	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"
)

// LoadAllServices loads all monitoring information for all services/components into the common merge registry
func LoadAllServices(registry ossmergemodel.ModelRegistry, pattern *regexp.Regexp) error {
	err := LoadEDBMetrics(registry)
	return err
}

// MergeOneService merges all the monitoring information collected from various sources for one particular service/component
func MergeOneService(model ossmergemodel.Model) {
	si := ServiceInfo{Model: model}
	oss := si.GetOSSServiceExtended()
	mi := si.GetMonitoringInfo(nil)
	et := oss.GeneralInfo.EntryType
	os := oss.GeneralInfo.OperationalStatus

	oss.OSSService.MonitoringInfo.Metrics = nil

	if mi != nil {
		for _, md := range mi.AllMetrics {
			// TODO: for now we copy all EDB Metrics directly into the OSS record
			// Once we process other metric sources, we should do an actual merge
			md.Metric.Environment = string(md.Environment.ToCRNString())
			if md.SourceEDBMetricData != nil {
				md.Metric.AddTags(ossrecord.MetricTagSourceEDB)
				if md.SourceEDBMetricData.Service != string(oss.OSSService.ReferenceResourceName) {
					issue := ossvalidation.NewIssue(ossvalidation.CRITICAL, "Service name in EDB metric does not match the service name of the parent OSS entry", `metric=%s   expected name="%s"`, md.SourceEDBMetricData.String(), oss.OSSService.ReferenceResourceName).TagRunAction(ossrunactions.Monitoring)
					debug.PrintError("monitoringinfo.MergeOneService(service=%s): %s", si.String(), issue.String())
					oss.OSSValidation.AddIssuePreallocated(issue)
					md.NumIssues++
					// TODO: check for other mismatches, e.g. on EntryType and OperationalStatus
				}
			}
			if md.NumIssues > 0 {
				md.Metric.AddTags(fmt.Sprintf("%s%d", ossrecord.MetricTagIssues, md.NumIssues))
			}
			oss.OSSService.MonitoringInfo.Metrics = append(oss.OSSService.MonitoringInfo.Metrics, &md.Metric)
		}
		sort.Slice(oss.OSSService.MonitoringInfo.Metrics, func(i, j int) bool {
			return oss.OSSService.MonitoringInfo.Metrics[i].String() < oss.OSSService.MonitoringInfo.Metrics[j].String()
		})
	}

	if oss.OSSService.GeneralInfo.OSSTags.Contains(osstags.EDBExclude) {
		if len(oss.OSSService.MonitoringInfo.Metrics) > 0 {
			oss.OSSValidation.AddNamedIssue(ossvalidation.EDBFoundButExcluded, "")
		} else {
			oss.OSSValidation.AddNamedIssue(ossvalidation.EDBExcludedOK, "")
		}
	} else if oss.OSSService.GeneralInfo.OSSTags.Contains(osstags.EDBInclude) {
		if len(oss.OSSService.MonitoringInfo.Metrics) > 0 {
			oss.OSSValidation.AddNamedIssue(ossvalidation.EDBIncludedOK, "")
		} else {
			oss.OSSValidation.AddNamedIssue(ossvalidation.EDBMissingButRequired, "")
		}
	} else if et != ossrecord.VMWARE &&
		et != ossrecord.TEMPLATE &&
		et != ossrecord.SUPERCOMPONENT &&
		et != ossrecord.COMPOSITE &&
		et != ossrecord.GAAS &&
		et != ossrecord.OTHEROSS &&
		et != ossrecord.CONTENT &&
		et != ossrecord.CONSULTING &&
		/*  We do expect EDB monitoring for internal services 		et != ossrecord.INTERNALSERVICE && */
		os != ossrecord.THIRDPARTY &&
		os != ossrecord.DEPRECATED &&
		os != ossrecord.RETIRED {
		if len(oss.OSSService.MonitoringInfo.Metrics) > 0 {
			// OK, nothing to say
		} else {
			if oss.OSSService.GeneralInfo.ClientFacing {
				oss.OSSValidation.AddNamedIssue(ossvalidation.EDBMissingClientFacing, "")
			} else {
				oss.OSSValidation.AddNamedIssue(ossvalidation.EDBMissingNotClientFacing, "")
			}
		}
	} else if len(oss.OSSService.MonitoringInfo.Metrics) > 0 {
		// OK, nothing to say
	} else {
		if oss.OSSService.GeneralInfo.ClientFacing {
			oss.OSSValidation.AddNamedIssue(ossvalidation.EDBMissingClientFacingNonStandard, "")
		} else {
			// OK, nothing to say
		}
	}
}

// LoadEDBMetrics loads all the recent metrics from EDB into MonitoringInfo records for each service
func LoadEDBMetrics(registry ossmergemodel.ModelRegistry) error {
	err := ListEDBMetricData(nil, func(e *EDBMetricData) {
		si, _ := LookupService(registry, e.Service, true)
		mi := si.GetMonitoringInfo(NewMonitoringInfo)
		environment := crn.Mask{}.SetPublicCloud(e.Location)
		md, _ := mi.LookupMetricData(e.Type.MetricType(), e.Plan, environment, true)
		md.SourceEDBMetricData = e
	})
	return err
}

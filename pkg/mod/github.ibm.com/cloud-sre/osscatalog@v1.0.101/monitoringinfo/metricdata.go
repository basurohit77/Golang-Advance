package monitoringinfo

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// MetricData represents all the merge information for one particular metric for one service (merged from multiple sources)
type MetricData struct {
	ossrecord.Metric
	Environment crn.Mask
	//	Source     MetricSource
	//	SourceName string
	//	SourceURL  string
	NumIssues           int
	SourceEDBMetricData *EDBMetricData
	//	SourceEDBEstadoMapping *EDBEstadoMapping
	//	SourceEstadoV1Weblinks *EstadoV1Weblink
	//	SourceEstadoV2Weblinks *EstadoV2Weblink
}

// MetricSource represents the source of one Metric (e.g. Estado V1, Estado V2, EDB, etc.)
type MetricSource string

// Constants for MetricSource
const (
	MetricEstadoV1  MetricSource = "EstadoV1"
	MetricEstadoV2  MetricSource = "EstadoV2"
	MetricEDBDirect MetricSource = "EDB-direct"
)

// MetricKey is the type for a unique key for a MetricData object, suitable for lookup and sorting
type MetricKey string

// MakeMetricKey returns a MetricKey containing the specified individual attributes
func MakeMetricKey(typ ossrecord.MetricType, plan string, environment crn.Mask) MetricKey {
	return MetricKey(fmt.Sprintf("%s/%s/%s", typ, plan, environment.ToCRNString()))
}

// Key returns a MetricKey for this Metric object
func (md *MetricData) Key() MetricKey {
	return MakeMetricKey(md.Type, md.PlanOrLabel, md.Environment)
}

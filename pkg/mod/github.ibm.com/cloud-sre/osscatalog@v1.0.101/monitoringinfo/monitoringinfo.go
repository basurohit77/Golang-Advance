package monitoringinfo

import (
	"reflect"

	"github.ibm.com/cloud-sre/osscatalog/crn"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
)

// MonitoringInfo contains all the monitoring-related information collected from various sources,
// for merging and consolidating
type MonitoringInfo struct {
	// Service                ossrecord.CRNServiceName
	AllMetrics map[MetricKey]*MetricData
}

// NewMonitoringInfo initializes a new empty MonitoringInfo record
func NewMonitoringInfo() *MonitoringInfo {
	mi := &MonitoringInfo{}
	mi.AllMetrics = make(map[MetricKey]*MetricData)
	return mi
}

// SubRecordType is the key for the Accessor for the MonitoringInfo sub-record inside a containing ServiceInfo
var SubRecordType = reflect.TypeOf(&MonitoringInfo{})

// Accessor is the function to access the MonitoringInfo sub-record inside a containing ServiceInfo
var Accessor ossmergemodel.AccessorFunc

// LookupService returns the ServiceInfo record associated with a given "comparable name" or creates a new record if appropriate.
// If no record exists for a comprableName and the parameter 'createIfNeeded' is false, 'nil' is returned.
func LookupService(registry ossmergemodel.ModelRegistry, name string, createIfNeeded bool) (si *ServiceInfo, found bool) {
	model, found := registry.LookupModel(name, createIfNeeded)
	if found {
		return &ServiceInfo{Model: model}, true
	}
	return nil, false
}

// ServiceInfo is a wrapper for ossmerge.Model, used to add some specialized methods in this package
type ServiceInfo struct {
	ossmergemodel.Model
	cachedMonitoringInfo *MonitoringInfo
}

// GetMonitoringInfo returns the MonitoringInfo sub-record contained in this ServiceInfo
func (si *ServiceInfo) GetMonitoringInfo(createFunc func() *MonitoringInfo) *MonitoringInfo {
	if si.cachedMonitoringInfo == nil {
		if Accessor == nil {
			Accessor = si.GetAccessor(SubRecordType)
		}
		var ret interface{}
		if createFunc != nil {
			f := func() interface{} {
				return createFunc()
			}
			ret = Accessor(si.Model, f)
		} else {
			ret = Accessor(si.Model, nil)
		}
		si.cachedMonitoringInfo = ret.(*MonitoringInfo)
	}
	return si.cachedMonitoringInfo
}

// GetOSSServiceExtended returns the OSSServiceExtended sub-record contained in this ServiceInfo
func (si *ServiceInfo) GetOSSServiceExtended() *ossrecordextended.OSSServiceExtended {
	return si.Model.OSSEntry().(*ossrecordextended.OSSServiceExtended)
}

// LookupMetricData returns one particular MetricData object contained in this MonitoringInfo
func (mi *MonitoringInfo) LookupMetricData(typ ossrecord.MetricType, plan string, environment crn.Mask, createIfNeeded bool) (md *MetricData, found bool) {
	key := MakeMetricKey(typ, plan, environment)
	if md, found = mi.AllMetrics[key]; found {
		return md, found
	} else if createIfNeeded {
		md = &MetricData{}
		md.Type = typ
		md.PlanOrLabel = plan
		md.Environment = environment
		mi.AllMetrics[key] = md
		return md, true
	}
	return nil, false
}

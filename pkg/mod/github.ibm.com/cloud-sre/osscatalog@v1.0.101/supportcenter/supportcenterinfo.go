package supportcenter

import (
	"reflect"

	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
)

// Info contains all the SupportCenter related information collected from various sources,
// for merging and consolidating
type Info struct {
	Candidate *Candidate
}

// NewSupportCenterInfo initializes a new empty supportcenter.Info record
func NewSupportCenterInfo() *Info {
	sci := &Info{}
	return sci
}

// SubRecordType is the key for the Accessor for the supportcenter.Info sub-record inside a containing ServiceInfo
var SubRecordType = reflect.TypeOf(&Info{})

// Accessor is the function to access the supportcenter.Info sub-record inside a containing ServiceInfo
var Accessor ossmergemodel.AccessorFunc

// LookupService returns the ServiceInfo record associated with a given "comparable name" or creates a new record if appropriate.
// If no record exists for a comparableName and the parameter 'createIfNeeded' is false, 'nil' is returned.
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
	cachedSupportCenterInfo *Info
}

// GetSupportCenterInfo returns the supportcenter.Info sub-record contained in this ServiceInfo
func (si *ServiceInfo) GetSupportCenterInfo(createFunc func() *Info) *Info {
	if si.cachedSupportCenterInfo == nil {
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
		si.cachedSupportCenterInfo = ret.(*Info)
	}
	return si.cachedSupportCenterInfo
}

// GetOSSServiceExtended returns the OSSServiceExtended sub-record contained in this ServiceInfo
func (si *ServiceInfo) GetOSSServiceExtended() *ossrecordextended.OSSServiceExtended {
	return si.Model.OSSEntry().(*ossrecordextended.OSSServiceExtended)
}

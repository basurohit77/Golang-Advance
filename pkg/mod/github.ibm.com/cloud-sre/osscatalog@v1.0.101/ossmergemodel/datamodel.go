package ossmergemodel

import (
	"reflect"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

// Model is the common base for all record types used in the data model for ossmerge functions
// (ServiceInfo, SegmentInfo, TribeInfo)
type Model interface {
	OSSEntry() ossrecord.OSSEntry
	IsDeletable() bool
	IsValid() bool
	IsUpdatable() bool
	HasPriorOSS() bool
	PriorOSSEntry() ossrecord.OSSEntry
	String() string
	Header() string
	Details() string
	Diffs() *compare.Output
	GlobalSortKey() string
	GetAccessor(reflect.Type) AccessorFunc
	GetOSSValidation() *ossvalidation.OSSValidation
	SkipMerge() (bool, string)
}

// AccessorFunc is a type of function that returns one sub-record inside an instance of a Model
// When a CreateFunc parameter is specified, that function is called to initialize
// a new sub-record if one was not already present; otherwise the AccessorFunc returns nil.
type AccessorFunc func(Model, CreateFunc) interface{}

// CreateFunc is a type of function used to create one new instance of a sub-record within an instance of a Model
type CreateFunc func() interface{}

// ModelRegistry is the common interface used to lookup Model instances for ossmerge functions
type ModelRegistry interface {
	LookupModel(id string, createIfNeeded bool) (m Model, ok bool)
	ListAllModels(pattern *regexp.Regexp, handler func(m Model)) error
}

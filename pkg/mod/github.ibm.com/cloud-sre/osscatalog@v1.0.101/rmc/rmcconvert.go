package rmc

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
)

var mapMaturity = map[string]ossrecord.OperationalStatus{
	"GA":           ossrecord.GA,
	"ga":           ossrecord.GA,
	"Beta":         ossrecord.BETA,
	"beta":         ossrecord.BETA,
	"Experimental": ossrecord.EXPERIMENTAL,
	"experimental": ossrecord.EXPERIMENTAL,
	"one_cloud":    ossrecord.GA, // XXX
	// "Other": ossrecord.OperationalStatusUnknown,
}

var mapType = map[string]ossrecord.EntryType{
	"service":          ossrecord.SERVICE,
	"platform_service": ossrecord.PLATFORMCOMPONENT,
	// "platform-service": ossrecord.PLATFORMCOMPONENT, // misspelled
	"composite": ossrecord.COMPOSITE,
	// "iaas":             ossrecord.IAAS, // unused
	"operations_only": ossrecord.EntryType(ossrecord.GAAS), // we use GaaS to match the default behavior of the RMC Operations tab
}

// GetOperationalStatus converts the Maturity field from a RMC Summary record into a standard ossrecord.OperationalStatus value.
// It returns an error if the conversion fails.
func (se *SummaryEntry) GetOperationalStatus() (ossrecord.OperationalStatus, error) {
	if result, ok := mapMaturity[se.Maturity]; ok {
		return result, nil
	}
	return ossrecord.OperationalStatusUnknown, fmt.Errorf(`Unknown Maturity="%s" in RMC record "%s" -- cannot map to OSS OperationalStatus`, se.Maturity, se.CRNServiceName)
}

// GetEntryType converts the Type field from a RMC Summary record into a standard ossrecord.EntryType value.
// It returns an error if the conversion fails.
func (se *SummaryEntry) GetEntryType() (ossrecord.EntryType, error) {
	if result, ok := mapType[se.Type]; ok {
		return result, nil
	}
	return ossrecord.EntryTypeUnknown, fmt.Errorf(`Unknown Type="%s" in RMC record "%s" -- cannot map to OSS EntryType`, se.Type, se.CRNServiceName)
}

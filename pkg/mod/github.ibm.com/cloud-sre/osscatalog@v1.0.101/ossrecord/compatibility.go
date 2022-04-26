package ossrecord

// OSSMetaData is an alias for OSSService, for backward compatibility
type OSSMetaData = OSSService

// MakeOSSRecordID creates a OSSEntryID for a OSS service/component record
// This function is an alias of MakeOSSServiceID, provided for backward compatibility
func MakeOSSRecordID(name CRNServiceName) OSSEntryID {
	return MakeOSSServiceID(name)
}

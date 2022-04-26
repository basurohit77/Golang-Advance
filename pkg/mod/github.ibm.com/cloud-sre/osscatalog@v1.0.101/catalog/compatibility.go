package catalog

import (
	"fmt"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
)

// OSSRecordExtended is an alias for OSSServiceExtended, for backward compatibility
type OSSRecordExtended = ossrecordextended.OSSServiceExtended

// ReadOSSRecord reads an OSS service/component record from the Global Catalog, given its name
// in the Staging instance (which contains the most recently updated data)
// This function is an extended version of ReadOSSService, provided for backward compatibility
func ReadOSSRecord(name ossrecord.CRNServiceName, incl IncludeOptions) (ossrec *ossrecordextended.OSSServiceExtended, err error) {
	id := ossrecord.MakeOSSServiceID(name)
	e, err := ReadOSSEntryByID(id, incl)
	if err != nil {
		return nil, err
	}
	switch e1 := e.(type) {
	case *ossrecord.OSSService:
		ossrec = &ossrecordextended.OSSServiceExtended{OSSService: *e1}
	case *ossrecordextended.OSSServiceExtended:
		ossrec = e1
	default:
		err := fmt.Errorf("ReadOSSRecord(%s) returned Catalog Resource of unexpected type %T (%v)", name, e, e)
		return nil, err
	}
	if ossrec.ReferenceResourceName != name {
		err := fmt.Errorf("ReadOSSRecord(%s) returned entry with unexpected name %s", name, ossrec.ReferenceResourceName)
		return nil, err
	}
	return ossrec, nil
}

// ReadOSSRecordProduction reads an OSS service/component record from the Global Catalog, given its name,
// in the Production instance (which contains the stable data)
// This function is an extended version of ReadOSSService, provided for backward compatibility
func ReadOSSRecordProduction(name ossrecord.CRNServiceName, incl IncludeOptions) (ossrec *ossrecordextended.OSSServiceExtended, err error) {
	id := ossrecord.MakeOSSServiceID(name)
	e, err := ReadOSSEntryByIDProduction(id, incl)
	if err != nil {
		return nil, err
	}
	switch e1 := e.(type) {
	case *ossrecord.OSSService:
		ossrec = &ossrecordextended.OSSServiceExtended{OSSService: *e1}
	case *ossrecordextended.OSSServiceExtended:
		ossrec = e1
	default:
		err := fmt.Errorf("ReadOSSRecordProduction(%s) returned Catalog Resource of unexpected type %T (%v)", name, e, e)
		return nil, err
	}
	if ossrec.ReferenceResourceName != name {
		err := fmt.Errorf("ReadOSSRecordProduction(%s) returned entry with unexpected name %s", name, ossrec.ReferenceResourceName)
		return nil, err
	}
	return ossrec, nil
}

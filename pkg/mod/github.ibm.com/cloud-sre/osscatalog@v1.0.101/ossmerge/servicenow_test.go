package ossmerge

import (
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"

	"github.ibm.com/cloud-sre/osscatalog/ossrecord"

	"github.ibm.com/cloud-sre/osscatalog/servicenow"
)

func TestMergeServiceNowImport(t *testing.T) {

	// Read a dummy SN import file to force initialization of data structures
	filename := "testdata/snimport_empty.csv"
	err := servicenow.ReadServiceNowImportFile(filename)
	testhelper.AssertError(t, err)

	entryName := "oss-test-record"
	si := &ServiceInfo{}
	si.SourceServiceNow.CRNServiceName = entryName
	si.OSSServiceExtended = *ossrecordextended.NewOSSServiceExtended(ossrecord.CRNServiceName(entryName))

	sni := &servicenow.SNImport{
		Name:                                "oss-test-record",
		DisplayName:                         "val_DisplayName",
		FullCRN:                             "val_FullCRN",
		StatusPageNotificationCategoryID:    "val_StatusPageNotificationCategoryID",
		Tier1SupportAssignmentGroup:         "val_Tier1SupportAssignmentGroup",
		Tier1OperationsAssignmentGroup:      "val_Tier1OperationsAssignmentGroup",
		Tier2SupportEscalationType:          "GitHub",
		Tier2SupportAssignmentGroup:         "val_Tier2SupportAssignmentGroup",
		Tier2SupportEscalationGitHubRepo:    "val_Tier2SupportEscalationGitHubRepo",
		Tier2OperationsEscalationType:       "Other",
		Tier2OperationsAssignmentGroup:      "val_Tier2OperationsAssignmentGroup",
		Tier2OperationsEscalationGitHubRepo: "val_Tier2OperationsEscalationGitHubRepo",
		//EntryType:                           "Cloud Service",
		EntryType:                   "SERVICE",
		ClientExperience:            "IBM Cloud Supported",
		CustomerFacing:              true,
		TOCEnabled:                  true,
		OperationalStatus:           "BETA IBM",
		OfferingManager:             "val_OfferingManager",
		StatusPageNotificationGroup: "val_StatusPageNotificationGroup",
		CreatedBy:                   "val_CreatedBy",
		UpdatedBy:                   "val_UpdatedBy",
		Segment:                     "val_Segment",
		Tribe:                       "val_Tribe",
		OperationsManager:           "val_OperationsManager",
		SupportManager:              "val_SupportManager",
	}
	servicenow.RegisterServiceNowImport(sni)

	si.mergeServiceNowImport(&si.SourceServiceNow, sni)

	testhelper.AssertEqual(t, "Name", entryName, si.SourceServiceNow.CRNServiceName)
	testhelper.AssertEqual(t, "DisplayName", "val_DisplayName", si.SourceServiceNow.DisplayName)
	testhelper.AssertEqual(t, "EntryType", ossrecord.SERVICE, si.SourceServiceNow.GeneralInfo.EntryType)
	testhelper.AssertEqual(t, "OperationalStatus", ossrecord.BETA, si.SourceServiceNow.GeneralInfo.OperationalStatus)
	testhelper.AssertEqual(t, "ClientExperience", ossrecord.ACSSUPPORTED, si.SourceServiceNow.Support.ClientExperience)
	testhelper.AssertEqual(t, "ClientFacing", true, si.SourceServiceNow.GeneralInfo.ClientFacing)
	testhelper.AssertEqual(t, "Operations.Manager", "val_OperationsManager", si.SourceServiceNow.Operations.Manager.Name)
	testhelper.AssertEqual(t, "Support.Tier2EscalationType", ossrecord.GITHUB, si.SourceServiceNow.Support.Tier2EscalationType)
	testhelper.AssertEqual(t, "Operations.Tier2EscalationType", ossrecord.OTHERESCALATION, si.SourceServiceNow.Operations.Tier2EscalationType)

	//	si.OSSValidation.Sort()
	//	fmt.Println(si.OSSValidation.Details())
}

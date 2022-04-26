package rmc

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.ibm.com/cloud-sre/osscatalog/compare"
	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// InvalidNameChars is a list of characters that are known to be invalid in RMC entry names (result in error 400)
const InvalidNameChars = "()/&"

// SummaryEntry is the format of one service/component entry stored in RMC
// The data in this struct is parsed from the RMC summary API, but does not directly
// map to this raw API, because it contains a complex multi-type "SoP" array
type SummaryEntry struct {
	CRNServiceName  ossrecord.CRNServiceName `json:"crnName"`
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	DisplayName     string                   `json:"displayName"`
	Type            string                   `json:"type"`
	Maturity        string                   `json:"maturity"`
	OneCloudService bool                     `json:"oneCloudService"`
	TargetGADate    string                   `json:"targetGADate"`
	ManagedBy       string                   `json:"managedBy"`
	OwningUser      string                   `json:"owningUser"`
	//	ParentCompositeService interface{} `json:"parentCompositeService,omitempty"`
	ParentCompositeService struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"parentCompositeService,omitempty"`
	ContainedResourceTypesRAWVALUE interface{} `json:"containedResourceTypes,omitempty"` // either a string or []string
	ContainedResourceTypes         []string    `json:"containedResourceTypesCORRECTED,omitempty"`
	Contributors                   []struct {
		Email       string `json:"email,omitempty"`
		Name        string `json:"name,omitempty"`
		Org         string `json:"org,omitempty"`
		Role        string `json:"role,omitempty"`
		MemberEmail string `json:"memberEmail,omitempty"`
		MemberName  string `json:"memberName,omitempty"`
	} `json:"contributors"`
	Progress struct {
		BCDRNotCompleted                bool `json:"BCDRNotCompleted"`
		ArchitectureNotCompleted        bool `json:"architectureNotCompleted"`
		BrokersNotCompleted             bool `json:"brokersNotCompleted"`
		BssNotCompleted                 bool `json:"bssNotCompleted"`
		DesignNotCompleted              bool `json:"designNotCompleted"`
		DocumentationNotCompleted       bool `json:"documentationNotCompleted"`
		IamNotCompleted                 bool `json:"iamNotCompleted"`
		LegalNotCompleted               bool `json:"legalNotCompleted"`
		MetadataNotCompleted            bool `json:"metadataNotCompleted"`
		OperationMonitoringNotCompleted bool `json:"operationMonitoringNotCompleted"`
		ProductionPushNotCompleted      bool `json:"productionPushNotCompleted"`
		SecurityNotCompleted            bool `json:"securityNotCompleted"`
		SummaryNotCompleted             bool `json:"summaryNotCompleted"`
		TranslationNotCompleted         bool `json:"translationNotCompleted"`
	} `json:"progress"`
	SOPBSS            *SOPBSS            `json:"bss,omitempty"`                                // Placeholder, never actually in the JSON
	SOPLegal          *SOPLegal          `json:"legal,omitempty"`                              // Placeholder, never actually in the JSON
	SOPSecurity       *SOPSecurity       `json:"security,omitempty"`                           // Placeholder, never actually in the JSON
	SOPOperations     *SOPOperations     `json:"operations,omitempty"`                         // Placeholder, never actually in the JSON
	SOPDesign         *SOPDesign         `json:"design,omitempty"`                             // Placeholder, never actually in the JSON
	SOPDocumentation  *SOPDocumentation  `json:"documentation,omitempty"`                      // Placeholder, never actually in the JSON
	SOPDedicated      *SOPDedicated      `json:"dedicated,omitempty"`                          // Placeholder, never actually in the JSON
	SOPArchitecture   *SOPArchitecture   `json:"architecture,omitempty"`                       // Placeholder, never actually in the JSON
	SOPGTM            *SOPGTM            `json:"gtm,omitempty"`                                // Placeholder, never actually in the JSON
	SOPFinalPlayback  *SOPFinalPlayback  `json:"finalPlayback,omitempty"`                      // Placeholder, never actually in the JSON
	SOPGoLiveApproval *SOPGoLiveApproval `json:"goLiveApproval,omitempty"`                     // Placeholder, never actually in the JSON
	SOPIAM            *SOPIAM            `json:"iam,omitempty"`                                // Placeholder, never actually in the JSON
	SOPBCDR           *SOPBCDR           `json:"businessContinuityDisasterRecovery,omitempty"` // Placeholder, never actually in the JSON
	SOPIMS            *SOPIMS            `json:"ims,omitempty"`                                // Placeholder, never actually in the JSON
}

// SOPCommon contains the common fields in every SOP sub-record
type SOPCommon struct {
	ID                   string        `json:"id"`
	Index                string        `json:"index"`
	Step                 string        `json:"step"`
	OwnerEmail           string        `json:"ownerEmail"`
	OwnerName            string        `json:"ownerName"`
	HumanTask            string        `json:"humanTask"`
	SystemFunction       string        `json:"systemFunction"`
	DependentOn          []interface{} `json:"dependentOn"`   // TODO: fix DependentOn type
	InnerWorkflow        []interface{} `json:"innerWorkflow"` // TODO: fix InnerWorkflow type
	State                string        `json:"state"`
	Contact              string        `json:"contact"`
	ChecklistContact     string        `json:"checklistContact,omitempty"`
	ChecklistStatus      string        `json:"checklistStatus,omitempty"`
	ChecklistStatusNotes string        `json:"checklistStatusNotes,omitempty"`
	Status               string        `json:"status,omitempty"`
	Notes                string        `json:"notes,omitempty"`
	ApprovedBy           string        `json:"approvedBy,omitempty"`
	ApprovedDateRAWVALUE interface{}   `json:"approvedDate,omitempty"` // either a string or float64
	ApprovedDate         string        `json:"approvedDateCORRECTED,omitempty"`
}

// Special hack to be able to obtain the SOPCommon embedded struct out of every SOP record
type sopCommonContainer interface {
	getSOPCommon() *SOPCommon
}

func (s *SOPCommon) getSOPCommon() *SOPCommon {
	return s
}

// SOPBSS is the "BSS" SOP sub-record
type SOPBSS struct {
	SOPCommon `json:",squash"`
}

// SOPLegal is the "Legal" SOP sub-record
type SOPLegal struct {
	SOPCommon         `json:",squash"`
	Coo               string `json:"coo"`
	MiniSDApprovedNew string `json:"miniSDApprovedNew"`
	MiniSDDocID       string `json:"miniSDDocID"`
	Selfcertify       bool   `json:"selfcertify,omitempty"`
	Export            string `json:"export"`
}

// SOPSecurity is the "Security" SOP sub-record
type SOPSecurity struct {
	SOPCommon                    `json:",squash"`
	PSIRTProductID               string `json:"PSIRTProductID,omitempty"`
	CumulusPageURL               string `json:"cumulusPageURL,omitempty"`
	Sec030Issue                  string `json:"sec030Issue,omitempty"`
	SecurityContact              string `json:"securityContact,omitempty"`
	SecurityReviewGitHubEpic     string `json:"securityReviewGitHubEpic,omitempty"`
	AOSOS                        string `json:"AOSOS,omitempty"`
	AOSOSNotes                   string `json:"AOSOSNotes,omitempty"`
	CSIRTLeads                   string `json:"CSIRTLeads,omitempty"`
	FFIEC                        string `json:"FFIEC,omitempty"`
	FFIECNotes                   string `json:"FFIECNotes,omitempty"`
	FISMA                        string `json:"FISMA,omitempty"`
	FISMANotes                   string `json:"FISMANotes,omitempty"`
	HIPAA                        string `json:"HIPAA,omitempty"`
	HIPAANotes                   string `json:"HIPAANotes,omitempty"`
	IBMID                        string `json:"IBMID,omitempty"`
	IBMIDNotes                   string `json:"IBMIDNotes,omitempty"`
	ISO270001                    string `json:"ISO270001,omitempty"`
	ISO270001Notes               string `json:"ISO270001Notes,omitempty"`
	ITCS300Compliance            string `json:"ITCS300Compliance,omitempty"`
	ITCS300ComplianceNotes       string `json:"ITCS300ComplianceNotes,omitempty"`
	PCI                          string `json:"PCI,omitempty"`
	PCINotes                     string `json:"PCINotes,omitempty"`
	PIIRegulated                 string `json:"PIIRegulated,omitempty"`
	PIIRegulatedNotes            string `json:"PIIRegulatedNotes,omitempty"`
	PSIRTProductName             string `json:"PSIRTProductName,omitempty"`
	PSIRTResponder               string `json:"PSIRTResponder,omitempty"`
	PSIRTResponderNotes          string `json:"PSIRTResponderNotes,omitempty"`
	SDcomplete                   string `json:"SDcomplete,omitempty"`
	SDcompleteNotes              string `json:"SDcompleteNotes,omitempty"`
	SHPS                         string `json:"SHPS,omitempty"`
	SHPSNotes                    string `json:"SHPSNotes,omitempty"`
	SOC2                         string `json:"SOC2,omitempty"`
	SOC2Notes                    string `json:"SOC2Notes,omitempty"`
	SOD                          string `json:"SOD,omitempty"`
	SODNotes                     string `json:"SODNotes,omitempty"`
	BestPractices                string `json:"bestPractices,omitempty"`
	BestPracticesNotes           string `json:"bestPracticesNotes,omitempty"`
	CodeScan                     string `json:"codeScan,omitempty"`
	CodeScanNotes                string `json:"codeScanNotes,omitempty"`
	ComplianceBluemix            string `json:"complianceBluemix,omitempty"`
	ComplianceBluemixNotes       string `json:"complianceBluemixNotes,omitempty"`
	CurrentComplianceDate        string `json:"currentComplianceDate,omitempty"`
	FullComplianceDate           string `json:"fullComplianceDate,omitempty"`
	DeployedInDedicated          bool   `json:"deployedInDedicated,omitempty"`
	Encryption                   string `json:"encryption,omitempty"`
	EncryptionNotes              string `json:"encryptionNotes,omitempty"`
	FocalPoint                   string `json:"focalPoint,omitempty"`
	InfraSecurity                string `json:"infraSecurity,omitempty"`
	InfraSecurityNotes           string `json:"infraSecurityNotes,omitempty"`
	Labeling                     string `json:"labeling,omitempty"`
	LabelingNotes                string `json:"labelingNotes,omitempty"`
	Logging                      string `json:"logging,omitempty"`
	LoggingNotes                 string `json:"loggingNotes,omitempty"`
	Monitoring                   string `json:"monitoring,omitempty"`
	MonitoringNotes              string `json:"monitoringNotes,omitempty"`
	NessusScan                   string `json:"nessusScan,omitempty"`
	NessusScanNotes              string `json:"nessusScanNotes,omitempty"`
	Networking                   string `json:"networking,omitempty"`
	NetworkingNotes              string `json:"networkingNotes,omitempty"`
	Patching                     string `json:"patching,omitempty"`
	PatchingNotes                string `json:"patchingNotes,omitempty"`
	Pattern                      string `json:"pattern,omitempty"`
	PreProdTesting               string `json:"preProdTesting,omitempty"`
	PreProdTestingNotes          string `json:"preProdTestingNotes,omitempty"`
	ProdTesting                  string `json:"prodTesting,omitempty"`
	ProdTestingNotes             string `json:"prodTestingNotes,omitempty"`
	SecurityScan                 string `json:"securityScan,omitempty"`
	SecurityScanNotes            string `json:"securityScanNotes,omitempty"`
	ServiceContractNumber        string `json:"serviceContractNumber,omitempty"`
	StandardPolicy               string `json:"standardPolicy,omitempty"`
	StandardPolicyNotes          string `json:"standardPolicyNotes,omitempty"`
	ThirdPartyManagement         string `json:"thirdPartyManagement,omitempty"`
	ThirdPartyManagementNotes    string `json:"thirdPartyManagementNotes,omitempty"`
	ThreatModel                  string `json:"threatModel,omitempty"`
	ThreatModelNotes             string `json:"threatModelNotes,omitempty"`
	UsingSoftlayer               string `json:"usingSoftlayer,omitempty"`
	UsingSoftlayerNotes          string `json:"usingSoftlayerNotes,omitempty"`
	HistoricalNotesLog           string `json:"historicalNotesLog,omitempty"`
	ProductDescription           string `json:"productDescription,omitempty"`
	ArchitectureDiagram          string `json:"architectureDiagram,omitempty"`
	LastUpdateDate               string `json:"lastUpdateDate,omitempty"`
	CheckListStatusNotesRAWVALUE string `json:"checkListStatusNotes,omitempty"` // Note misspelled capitalization of checkList XXX
	DeployedInDedicatedTotal     string `json:"deployedInDedicated_total,omitempty"`
	DeployedInDedicatedGreen     string `json:"deployedInDedicated_green,omitempty"`
}

// SOPOperations is the "Operations" SOP sub-record
type SOPOperations struct {
	SOPCommon    `json:",squash"`
	Availability *struct {
		Regions struct {
			China string `json:"china"`
			Lyp   string `json:"lyp"`
			Syp   string `json:"syp"`
			Yp    string `json:"yp"`
		} `json:"regions"`
	} `json:"availability"`
	Capacity *struct {
		ProdAcctNum   string `json:"ProdAcctNum"`
		StageAcctNum  string `json:"StageAcctNum"`
		OrderDate     string `json:"orderDate"`
		RequestsSheet string `json:"requestsSheet"`
		VlanYesNo     string `json:"vlanYesNo"`
	} `json:"capacity"`
	Devops *struct {
		Checklist struct {
			Status string `json:"Status"`
		} `json:"Checklist"`
		AlertType              string `json:"alertType"`
		Contact                string `json:"contact"`
		EstadoServiceName      string `json:"estadoServiceName"`
		MonitoringProvisioning string `json:"monitoringProvisioning"`
		Notes                  string `json:"notes"`
		OpsPlanComplete        string `json:"opsPlanComplete"`
		State                  string `json:"state"`
		Status                 string `json:"status"`
		Step                   string `json:"step,omitempty"`
	} `json:"devops"`
	DoctorGithubTrackerURL string `json:"doctorGithubTrackerUrl,omitempty"`
	Focal                  string `json:"focal,omitempty"`
	Monitoring             *struct {
		Checklist struct {
			Status string `json:"Status"`
		} `json:"Checklist"`
		ConsumptionMonitoringType string `json:"consumptionMonitoringType"`
		Contact                   string `json:"contact"`
		State                     string `json:"state"`
		Status                    string `json:"status"`
		Step                      string `json:"step,omitempty"`
	} `json:"monitoring"`
	Notification *struct {
		Group                string `json:"group"`
		ID                   string `json:"id"`
		NotificationCategory string `json:"notificationCategory,omitempty"`
		Step                 string `json:"step,omitempty"`
	} `json:"notification"`
	Support *struct {
		Checklist struct {
			Status string `json:"Status"`
		} `json:"Checklist"`
		EscalationsContact         string `json:"EscalationsContact"`
		AlertTarget                string `json:"alertTarget"`
		AlertType                  string `json:"alertType"`
		BizAPContact1              string `json:"bizAPContact1"`
		BizAPContact2              string `json:"bizAPContact2"`
		BizAmericasContact1        string `json:"bizAmericasContact1"`
		BizAmericasContact2        string `json:"bizAmericasContact2"`
		BizEMEAContact1            string `json:"bizEMEAContact1"`
		BizEMEAContact2            string `json:"bizEMEAContact2"`
		ChecklistStatusMain        string `json:"checklistStatusMain"`
		ChecklistStatusTech1       string `json:"checklistStatusTech1"`
		Contact                    string `json:"contact"`
		ContactNotesL              string `json:"contactNotesL"`
		CoverageGeneral            string `json:"coverageGeneral"`
		CoverageNotes              string `json:"coverageNotes"`
		CoverageSevirity1          string `json:"coverageSevirity1"`
		DsetRoutingTarget          string `json:"dsetRoutingTarget"`
		ForumChecklistStatusBefore string `json:"forumChecklistStatusBefore"`
		ForumContacts              string `json:"forumContacts"`
		ForumOtherTags             string `json:"forumOtherTags"`
		ForumPrimaryTag            string `json:"forumPrimaryTag"`
		Kbinfo                     string `json:"kbinfo"`
		ParatureCategory           string `json:"paratureCategory"`
		Pdinfo                     string `json:"pdinfo"`
		RoutingNotes               string `json:"routingNotes"`
		RoutingTarget              string `json:"routingTarget"`
		SkillsTransferDate         string `json:"skillsTransferDate"`
		SkillsTransferNotes        string `json:"skillsTransferNotes"`
		State                      string `json:"state"`
		Status                     string `json:"status"`
		Support247Contact1         string `json:"support247Contact1"`
		Support247Contact2         string `json:"support247Contact2"`
		SupportRoutingType         string `json:"supportRoutingType"`
		TriageInfo                 string `json:"triageInfo"`
		DSETRoutingType            string `json:"dsetRoutingType,omitempty"`
		Step                       string `json:"step,omitempty"`
	} `json:"support"`
}

// SOPDesign is the "Design" SOP sub-record
type SOPDesign struct {
	SOPCommon                `json:",squash"`
	Nps                      bool   `json:"NPS"`
	AchievedNPS              bool   `json:"achievedNPS"`
	AgreeToMaintainNPS       bool   `json:"agreeToMaintainNPS"`
	Bullets                  bool   `json:"bullets"`
	CatalogDetailsContent    bool   `json:"catalogDetailsContent"`
	CompletedUserTests       bool   `json:"completedUserTests"`
	DesignFocalEmail         string `json:"designFocalEmail,omitempty"`
	EmbeddedInBluemix        bool   `json:"embeddedInBluemix"`
	FollowsStyleGuide        bool   `json:"followsStyleGuide"`
	HaveReadPlaybackGuide    bool   `json:"haveReadPlaybackGuide"`
	NpsUsedFeedback          bool   `json:"npsUsedFeedback"`
	ProxyPath                string `json:"proxyPath,omitempty"`
	Screenshots              bool   `json:"screenshots"`
	ServiceConsole           bool   `json:"serviceConsole"`
	ServiceIcon              bool   `json:"serviceIcon"`
	ServiceIconGitHubIssue   string `json:"serviceIconGitHubIssue,omitempty"`
	ShortAndLongDescriptions bool   `json:"shortAndLongDescriptions"`
	UIDevelopmentFocalEmail  string `json:"uiDevelopmentFocalEmail,omitempty"`
	UIGitHubRepo             string `json:"uiGitHubRepo,omitempty"`
	UserTesting              bool   `json:"userTesting"`
	UserTestingUsedFeedback  bool   `json:"userTestingUsedFeedback"`
	UsesAPITemplate          bool   `json:"usesAPITemplate"`
}

// SOPDocumentation is the "Documentation" SOP sub-record
type SOPDocumentation struct {
	SOPCommon          `json:",squash"`
	Gheid              string `json:"GHEID,omitempty"`
	Ghid               string `json:"GHID,omitempty"`
	DocContentContact  string `json:"docContentContact,omitempty"`
	DocumentationNotes string `json:"documentationNotes,omitempty"`
	TechnicalOwner     string `json:"technicalOwner,omitempty"`
	Selfcertify        bool   `json:"selfcertify,omitempty"`
}

// SOPDedicated is the "Dedicated" SOP sub-record
type SOPDedicated struct {
	SOPCommon `json:",squash"`
}

// SOPArchitecture is the "Architecture" SOP sub-record
type SOPArchitecture struct {
	SOPCommon         `json:",squash"`
	ArchGitHubRepo    string `json:"archGitHubRepo,omitempty"`
	GitHubBacklogRepo string `json:"gitHubBacklogRepo,omitempty"`
	FocalEmail        string `json:"focalEmail,omitempty"`
}

// SOPGTM is the "GTM" SOP sub-record
type SOPGTM struct {
	SOPCommon `json:",squash"`
}

// SOPFinalPlayback is the "finalPlayback" SOP sub-record
type SOPFinalPlayback struct {
	SOPCommon `json:",squash"`
}

// SOPGoLiveApproval is the "goLiveApproval" SOP sub-record
type SOPGoLiveApproval struct {
	SOPCommon `json:",squash"`
}

// SOPIAM is the "IAM" SOP sub-record
type SOPIAM struct {
	SOPCommon `json:",squash"`
}

// SOPBCDR is the "businessContinuityDisasterRecovery" SOP sub-record
type SOPBCDR struct {
	SOPCommon    `json:",squash"`
	GitHubRepo   string `json:"GitHubRepo,omitempty"`
	ContactEmail string `json:"contactEmail,omitempty"`
}

// SOPIMS is the "ims" SOP sub-record
type SOPIMS struct {
	SOPCommon `json:",squash"`
	Certify   []string `json:"certify,omitempty"`
}

/*
// sopRaw is the raw struct from decoding the SOP sub-records from the RMC summary API
// Note: we do not actually use this struct; instead we explicitly parse each SOP sub-record type
type sopRaw struct {
	ID             string        `json:"id"`
	Index          string        `json:"index"`
	Step           string        `json:"step"`
	OwnerEmail     string        `json:"ownerEmail"`
	OwnerName      string        `json:"ownerName"`
	HumanTask      string        `json:"humanTask"`
	SystemFunction string        `json:"systemFunction"`
	DependentOn    []interface{} `json:"dependentOn"`
	InnerWorkflow  []interface{} `json:"innerWorkflow"`
	State          string        `json:"state"`
	Contact        string        `json:"contact"`
	// Remaining fields Sop depend on the type of entry (ID)
	Gheid              string `json:"GHEID"`
	Ghid               string `json:"GHID"`
	GitHubRepo         string `json:"GitHubRepo"`
	Nps                bool   `json:"NPS"`
	PSIRTProductID     string `json:"PSIRTProductID"`
	AchievedNPS        bool   `json:"achievedNPS"`
	AgreeToMaintainNPS bool   `json:"agreeToMaintainNPS"`
	ArchGitHubRepo     string `json:"archGitHubRepo"`
	Availability       *struct {
		Regions struct {
			China string `json:"china"`
			Lyp   string `json:"lyp"`
			Syp   string `json:"syp"`
			Yp    string `json:"yp"`
		} `json:"regions"`
	} `json:"availability"`
	Bullets  bool `json:"bullets"`
	Capacity *struct {
		ProdAcctNum   string `json:"ProdAcctNum"`
		StageAcctNum  string `json:"StageAcctNum"`
		OrderDate     string `json:"orderDate"`
		RequestsSheet string `json:"requestsSheet"`
		VlanYesNo     string `json:"vlanYesNo"`
	} `json:"capacity"`
	CatalogDetailsContent bool     `json:"catalogDetailsContent"`
	Certify               []string `json:"certify"`
	CompletedUserTests    bool     `json:"completedUserTests"`
	ContactEmail          string   `json:"contactEmail"`
	Coo                   string   `json:"coo"`
	CumulusPageURL        string   `json:"cumulusPageURL"`
	DesignFocalEmail      string   `json:"designFocalEmail"`
	Devops                *struct {
		Checklist struct {
			Status string `json:"Status"`
		} `json:"Checklist"`
		AlertType              string `json:"alertType"`
		Contact                string `json:"contact"`
		EstadoServiceName      string `json:"estadoServiceName"`
		MonitoringProvisioning string `json:"monitoringProvisioning"`
		Notes                  string `json:"notes"`
		OpsPlanComplete        string `json:"opsPlanComplete"`
		State                  string `json:"state"`
		Status                 string `json:"status"`
	} `json:"devops"`
	DocContentContact      string `json:"docContentContact"`
	DoctorGithubTrackerURL string `json:"doctorGithubTrackerUrl"`
	DocumentationNotes     string `json:"documentationNotes"`
	EmbeddedInBluemix      bool   `json:"embeddedInBluemix"`
	Export                 string `json:"export"`
	Focal                  string `json:"focal"`
	FocalEmail             string `json:"focalEmail"`
	FollowsStyleGuide      bool   `json:"followsStyleGuide"`
	GitHubBacklogRepo      string `json:"gitHubBacklogRepo"`
	HaveReadPlaybackGuide  bool   `json:"haveReadPlaybackGuide"`
	MiniSDApprovedNew      string `json:"miniSDApprovedNew"`
	MiniSDDocID            string `json:"miniSDDocID"`
	Monitoring             *struct {
		Checklist struct {
			Status string `json:"Status"`
		} `json:"Checklist"`
		ConsumptionMonitoringType string `json:"consumptionMonitoringType"`
		Contact                   string `json:"contact"`
		State                     string `json:"state"`
		Status                    string `json:"status"`
		Step                      string `json:"step"`
	} `json:"monitoring"`
	Notes        string `json:"notes"`
	Notification *struct {
		Group string `json:"group"`
		ID    string `json:"id"`
	} `json:"notification"`
	NpsUsedFeedback          bool   `json:"npsUsedFeedback"`
	ProxyPath                string `json:"proxyPath"`
	Screenshots              bool   `json:"screenshots"`
	Sec030Issue              string `json:"sec030Issue"`
	SecurityContact          string `json:"securityContact"`
	SecurityReviewGitHubEpic string `json:"securityReviewGitHubEpic"`
	Selfcertify              bool   `json:"selfcertify"`
	ServiceConsole           bool   `json:"serviceConsole"`
	ServiceIcon              bool   `json:"serviceIcon"`
	ServiceIconGitHubIssue   string `json:"serviceIconGitHubIssue"`
	ShortAndLongDescriptions bool   `json:"shortAndLongDescriptions"`
	Status                   string `json:"status"`
	Support                  *struct {
		Checklist struct {
			Status string `json:"Status"`
		} `json:"Checklist"`
		EscalationsContact         string `json:"EscalationsContact"`
		AlertTarget                string `json:"alertTarget"`
		AlertType                  string `json:"alertType"`
		BizAPContact1              string `json:"bizAPContact1"`
		BizAPContact2              string `json:"bizAPContact2"`
		BizAmericasContact1        string `json:"bizAmericasContact1"`
		BizAmericasContact2        string `json:"bizAmericasContact2"`
		BizEMEAContact1            string `json:"bizEMEAContact1"`
		BizEMEAContact2            string `json:"bizEMEAContact2"`
		ChecklistStatusMain        string `json:"checklistStatusMain"`
		ChecklistStatusTech1       string `json:"checklistStatusTech1"`
		Contact                    string `json:"contact"`
		ContactNotesL              string `json:"contactNotesL"`
		CoverageGeneral            string `json:"coverageGeneral"`
		CoverageNotes              string `json:"coverageNotes"`
		CoverageSevirity1          string `json:"coverageSevirity1"`
		DsetRoutingTarget          string `json:"dsetRoutingTarget"`
		ForumChecklistStatusBefore string `json:"forumChecklistStatusBefore"`
		ForumContacts              string `json:"forumContacts"`
		ForumOtherTags             string `json:"forumOtherTags"`
		ForumPrimaryTag            string `json:"forumPrimaryTag"`
		Kbinfo                     string `json:"kbinfo"`
		ParatureCategory           string `json:"paratureCategory"`
		Pdinfo                     string `json:"pdinfo"`
		RoutingNotes               string `json:"routingNotes"`
		RoutingTarget              string `json:"routingTarget"`
		SkillsTransferDate         string `json:"skillsTransferDate"`
		SkillsTransferNotes        string `json:"skillsTransferNotes"`
		State                      string `json:"state"`
		Status                     string `json:"status"`
		Support247Contact1         string `json:"support247Contact1"`
		Support247Contact2         string `json:"support247Contact2"`
		SupportRoutingType         string `json:"supportRoutingType"`
		TriageInfo                 string `json:"triageInfo"`
	} `json:"support"`
	TechnicalOwner          string `json:"technicalOwner"`
	UIDevelopmentFocalEmail string `json:"uiDevelopmentFocalEmail"`
	UIGitHubRepo            string `json:"uiGitHubRepo"`
	UserTesting             bool   `json:"userTesting"`
	UserTestingUsedFeedback bool   `json:"userTestingUsedFeedback"`
	UsesAPITemplate         bool   `json:"usesAPITemplate"`
}
*/

/*
// summaryMessage defines the "message" sub-record in a result from the RMC summary API
// Note: we do not actually use this struct; we assume that the "message" has the same
// structure as "data", which is represented by the SummaryEntry type
type summaryMessage struct {
	Sop []struct {
		Gheid              string `json:"GHEID"`
		Ghid               string `json:"GHID"`
		GitHubRepo         string `json:"GitHubRepo"`
		Nps                bool   `json:"NPS"`
		PSIRTProductID     string `json:"PSIRTProductID"`
		AchievedNPS        bool   `json:"achievedNPS"`
		AgreeToMaintainNPS bool   `json:"agreeToMaintainNPS"`
		ArchGitHubRepo     string `json:"archGitHubRepo"`
		Availability       struct {
			Regions struct {
				China string `json:"china"`
				Lyp   string `json:"lyp"`
				Syp   string `json:"syp"`
				Yp    string `json:"yp"`
			} `json:"regions"`
		} `json:"availability"`
		Bullets  bool `json:"bullets"`
		Capacity struct {
			ProdAcctNum   string `json:"ProdAcctNum"`
			StageAcctNum  string `json:"StageAcctNum"`
			OrderDate     string `json:"orderDate"`
			RequestsSheet string `json:"requestsSheet"`
			VlanYesNo     string `json:"vlanYesNo"`
		} `json:"capacity"`
		CatalogDetailsContent bool          `json:"catalogDetailsContent"`
		Certify               []string      `json:"certify"`
		ChecklistStatus       string        `json:"checklistStatus"`
		ChecklistStatusNotes  string        `json:"checklistStatusNotes"`
		CompletedUserTests    bool          `json:"completedUserTests"`
		Contact               string        `json:"contact"`
		ContactEmail          string        `json:"contactEmail"`
		Coo                   string        `json:"coo"`
		CumulusPageURL        string        `json:"cumulusPageURL"`
		DependentOn           []interface{} `json:"dependentOn"`
		DesignFocalEmail      string        `json:"designFocalEmail"`
		Devops                struct {
			Checklist struct {
				Status string `json:"Status"`
			} `json:"Checklist"`
			AlertType              string `json:"alertType"`
			Contact                string `json:"contact"`
			EstadoServiceName      string `json:"estadoServiceName"`
			MonitoringProvisioning string `json:"monitoringProvisioning"`
			Notes                  string `json:"notes"`
			OpsPlanComplete        string `json:"opsPlanComplete"`
			State                  string `json:"state"`
			Status                 string `json:"status"`
		} `json:"devops"`
		DocContentContact      string        `json:"docContentContact"`
		DoctorGithubTrackerURL string        `json:"doctorGithubTrackerUrl"`
		DocumentationNotes     string        `json:"documentationNotes"`
		EmbeddedInBluemix      bool          `json:"embeddedInBluemix"`
		Export                 string        `json:"export"`
		Focal                  string        `json:"focal"`
		FocalEmail             string        `json:"focalEmail"`
		FollowsStyleGuide      bool          `json:"followsStyleGuide"`
		GitHubBacklogRepo      string        `json:"gitHubBacklogRepo"`
		HaveReadPlaybackGuide  bool          `json:"haveReadPlaybackGuide"`
		HumanTask              string        `json:"humanTask"`
		ID                     string        `json:"id"`
		Index                  string        `json:"index"`
		InnerWorkflow          []interface{} `json:"innerWorkflow"`
		MiniSDApprovedNew      string        `json:"miniSDApprovedNew"`
		MiniSDDocID            string        `json:"miniSDDocID"`
		Monitoring             struct {
			Checklist struct {
				Status string `json:"Status"`
			} `json:"Checklist"`
			ConsumptionMonitoringType string `json:"consumptionMonitoringType"`
			Contact                   string `json:"contact"`
			State                     string `json:"state"`
			Status                    string `json:"status"`
		} `json:"monitoring"`
		Notes        string `json:"notes"`
		Notification struct {
			Group string `json:"group"`
			ID    string `json:"id"`
		} `json:"notification"`
		NpsUsedFeedback          bool   `json:"npsUsedFeedback"`
		OwnerEmail               string `json:"ownerEmail"`
		OwnerName                string `json:"ownerName"`
		ProxyPath                string `json:"proxyPath"`
		Screenshots              bool   `json:"screenshots"`
		Sec030Issue              string `json:"sec030Issue"`
		SecurityContact          string `json:"securityContact"`
		SecurityReviewGitHubEpic string `json:"securityReviewGitHubEpic"`
		Selfcertify              bool   `json:"selfcertify"`
		ServiceConsole           bool   `json:"serviceConsole"`
		ServiceIcon              bool   `json:"serviceIcon"`
		ServiceIconGitHubIssue   string `json:"serviceIconGitHubIssue"`
		ShortAndLongDescriptions bool   `json:"shortAndLongDescriptions"`
		State                    string `json:"state"`
		Status                   string `json:"status"`
		Step                     string `json:"step"`
		Support                  struct {
			Checklist struct {
				Status string `json:"Status"`
			} `json:"Checklist"`
			EscalationsContact         string `json:"EscalationsContact"`
			AlertTarget                string `json:"alertTarget"`
			AlertType                  string `json:"alertType"`
			BizAPContact1              string `json:"bizAPContact1"`
			BizAPContact2              string `json:"bizAPContact2"`
			BizAmericasContact1        string `json:"bizAmericasContact1"`
			BizAmericasContact2        string `json:"bizAmericasContact2"`
			BizEMEAContact1            string `json:"bizEMEAContact1"`
			BizEMEAContact2            string `json:"bizEMEAContact2"`
			ChecklistStatusMain        string `json:"checklistStatusMain"`
			ChecklistStatusTech1       string `json:"checklistStatusTech1"`
			Contact                    string `json:"contact"`
			ContactNotesL              string `json:"contactNotesL"`
			CoverageGeneral            string `json:"coverageGeneral"`
			CoverageNotes              string `json:"coverageNotes"`
			CoverageSevirity1          string `json:"coverageSevirity1"`
			DsetRoutingTarget          string `json:"dsetRoutingTarget"`
			ForumChecklistStatusBefore string `json:"forumChecklistStatusBefore"`
			ForumContacts              string `json:"forumContacts"`
			ForumOtherTags             string `json:"forumOtherTags"`
			ForumPrimaryTag            string `json:"forumPrimaryTag"`
			Kbinfo                     string `json:"kbinfo"`
			ParatureCategory           string `json:"paratureCategory"`
			Pdinfo                     string `json:"pdinfo"`
			RoutingNotes               string `json:"routingNotes"`
			RoutingTarget              string `json:"routingTarget"`
			SkillsTransferDate         string `json:"skillsTransferDate"`
			SkillsTransferNotes        string `json:"skillsTransferNotes"`
			State                      string `json:"state"`
			Status                     string `json:"status"`
			Support247Contact1         string `json:"support247Contact1"`
			Support247Contact2         string `json:"support247Contact2"`
			SupportRoutingType         string `json:"supportRoutingType"`
			TriageInfo                 string `json:"triageInfo"`
		} `json:"support"`
		SystemFunction          string `json:"systemFunction"`
		TechnicalOwner          string `json:"technicalOwner"`
		UIDevelopmentFocalEmail string `json:"uiDevelopmentFocalEmail"`
		UIGitHubRepo            string `json:"uiGitHubRepo"`
		UserTesting             bool   `json:"userTesting"`
		UserTestingUsedFeedback bool   `json:"userTestingUsedFeedback"`
		UsesAPITemplate         bool   `json:"usesAPITemplate"`
	} `json:"SOP"`
	ContainedResourceTypes string `json:"containedResourceTypes"`
	Contributors           []struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	} `json:"contributors"`
	ID                     string      `json:"id"`
	ManagedBy              string      `json:"managedBy"`
	Maturity               string      `json:"maturity"`
	Name                   string      `json:"name"`
	OneCloudService        bool        `json:"oneCloudService"`
	OwningUser             string      `json:"owningUser"`
	ParentCompositeService interface{} `json:"parentCompositeService"`
	Progress               struct {
		BCDRNotCompleted                bool `json:"BCDRNotCompleted"`
		ArchitectureNotCompleted        bool `json:"architectureNotCompleted"`
		BrokersNotCompleted             bool `json:"brokersNotCompleted"`
		BssNotCompleted                 bool `json:"bssNotCompleted"`
		DesignNotCompleted              bool `json:"designNotCompleted"`
		DocumentationNotCompleted       bool `json:"documentationNotCompleted"`
		IamNotCompleted                 bool `json:"iamNotCompleted"`
		LegalNotCompleted               bool `json:"legalNotCompleted"`
		MetadataNotCompleted            bool `json:"metadataNotCompleted"`
		OperationMonitoringNotCompleted bool `json:"operationMonitoringNotCompleted"`
		ProductionPushNotCompleted      bool `json:"productionPushNotCompleted"`
		SecurityNotCompleted            bool `json:"securityNotCompleted"`
		SummaryNotCompleted             bool `json:"summaryNotCompleted"`
		TranslationNotCompleted         bool `json:"translationNotCompleted"`
	} `json:"progress"`
	TargetGADate string `json:"targetGADate"`
	Type         string `json:"type"`
}
*/

// summaryResult is a container for the raw result from the RMC summary API
// (prior to parsing the various SOP sub-records)
type summaryResult struct {
	Data *struct {
		SummaryEntry
		SOP []map[string]interface{} `json:"SOP"`
	} `json:"data"`
	//	Message     *summaryMessage `json:"message"`
	/*
		Message *struct {
			SummaryEntry
			SOP []map[string]interface{} `json:"SOP"`
		} `json:"message"`
	*/
	RequestedAt string `json:"requestedAt"`
	ResponsedAt string `json:"responsedAt"`
	StatusCode  int64  `json:"statusCode"`
}

// ReadRMCSummaryEntry reads one service summary entry from RMC, given its name
func ReadRMCSummaryEntry(name ossrecord.CRNServiceName, testMode bool) (*SummaryEntry, error) {
	var actualURL string
	if testMode {
		actualURL = fmt.Sprintf(rmcTestSummaryURL, string(name))
	} else {
		actualURL = fmt.Sprintf(rmcSummaryURL, string(name))
	}
	key, err := rest.GetToken(rmcKeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for RMC")
		return nil, err
	}
	var result = new(summaryResult)
	err = rest.DoHTTPGet(actualURL, key, nil, "RMC", debug.RMC, result)
	if err != nil {
		if rest.GetHTTPStatusCode(err) == http.StatusInternalServerError || rest.GetHTTPStatusCode(err) == http.StatusNotFound {
			errString := err.Error()
			if strings.Contains(errString, `doesn't exists in RMC`) || strings.Contains(errString, `does not exist in RMC`) || strings.Contains(errString, `HTTPError code=404 Not Found : {"error":"Error: Not found"}`) {
				return nil, rest.MakeHTTPError(err, nil, true, `Entry "%s" not found in RMC`, name)
			}
		}
		if rest.GetHTTPStatusCode(err) == http.StatusBadRequest {
			errString := err.Error()
			if strings.Contains(errString, `"msg":"Invalid value","param":"serviceName","location":"params"}]}`) && strings.ContainsAny(string(name), InvalidNameChars) {
				return nil, rest.MakeHTTPError(err, nil, true, `Entry "%s" not found in RMC (invalid name syntax)`, name)
			}
		}
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		err = fmt.Errorf("Non-OK status code: %v", result.StatusCode)
		return nil, err
	}

	switch result.Data.SummaryEntry.CRNServiceName {
	case name:
		// normal case

	case "":
		// Copy the CRNServiceName, which is actually not part of the JSON payload
		// XXX Note this is *not* the same as the "Name" attribute in the JSON payload
		result.Data.SummaryEntry.CRNServiceName = name

	default:
		// XXX Name mismatch. Should generate some warning here ... or leave it to the merge logic
		//result.Data.SummaryEntry.CRNServiceName = name
	}

	// Fixup raw values
	if result.Data.ContainedResourceTypesRAWVALUE != nil {
		switch val := result.Data.ContainedResourceTypesRAWVALUE.(type) {
		case string:
			result.Data.ContainedResourceTypes = []string{val}
		case []string:
			result.Data.ContainedResourceTypes = val
		case []interface{}:
			for _, elem := range val {
				switch e1 := elem.(type) {
				case string:
					result.Data.ContainedResourceTypes = append(result.Data.ContainedResourceTypes, e1)
				default:
					err = fmt.Errorf("Unexpected element type %T for ContainedResourceTypes: %#v", elem, result.Data.ContainedResourceTypesRAWVALUE)
					return nil, err
				}
			}
		default:
			err = fmt.Errorf("Unexpected type %T for ContainedResourceTypes: %#v", result.Data.ContainedResourceTypesRAWVALUE, result.Data.ContainedResourceTypesRAWVALUE)
			return nil, err
		}
	}

	err = parseSOPRecords(fmt.Sprintf(`ReadRMCSummaryEntry(%s).Data`, result.Data.CRNServiceName), result.Data.SOP, &result.Data.SummaryEntry)
	if err != nil {
		return nil, err
	}

	debug.Debug(debug.RMC, `RMC entry %s{Name:%q  Type:%q  Maturity:%q  OneCloudService:%v  ManagedBy:%q  OwningUser:%q  ParentComposite:%q  ID:%q}`,
		result.Data.CRNServiceName, result.Data.Name, result.Data.Type, result.Data.Maturity, result.Data.OneCloudService,
		result.Data.ManagedBy, result.Data.OwningUser, result.Data.ParentCompositeService.Name, result.Data.ID)
	// XXX Should we also check the Messages sub-record?

	return &result.Data.SummaryEntry, nil
}

func parseSOPRecords(label string, maps []map[string]interface{}, entry *SummaryEntry) error {
	decode := func(label string, in map[string]interface{}, out interface{}) error {
		decoderConfig := mapstructure.DecoderConfig{
			Result:      out,
			TagName:     `json`,
			ErrorUnused: true,
			Metadata:    &mapstructure.Metadata{},
		}
		decoder, err := mapstructure.NewDecoder(&decoderConfig)
		if err != nil {
			return err
		}
		err = decoder.Decode(in)
		if len(decoderConfig.Metadata.Unused) > 0 {
			debug.Debug(debug.RMC|debug.Fine, `%s: Unused keys in RMC SOP record: %q`, label, decoderConfig.Metadata.Unused)
		}
		if err != nil {
			return err
		}
		if debug.IsDebugEnabled(debug.RMC | debug.Fine) {
			diffs := compare.Output{IncludeEqual: true}
			compare.DeepCompare("result."+label, out, "json."+label, in, &diffs)
			str := diffs.StringWithPrefix("DEBUG:")
			debug.Debug(debug.RMC|debug.Fine, "*** Comparing parsed result with raw JSON object for SOP \"%s\"\n%s", label, str)
		}

		// Fixup raw values
		var checklistStatusNotesRAWVALUE string
		if sopSecurity, ok := out.(*SOPSecurity); ok {
			checklistStatusNotesRAWVALUE = sopSecurity.CheckListStatusNotesRAWVALUE
		}
		if commonContainer, ok := out.(sopCommonContainer); ok {
			common := commonContainer.getSOPCommon()
			if common.ApprovedDateRAWVALUE != nil {
				switch val := common.ApprovedDateRAWVALUE.(type) {
				case float64:
					common.ApprovedDate = fmt.Sprintf("%v", val) // TODO: should we convert the ApprovedDate float64 date into readable format?
				case string:
					common.ApprovedDate = val
				default:
				}
			}
			if checklistStatusNotesRAWVALUE != "" {
				if common.ChecklistStatusNotes == "" {
					common.ChecklistStatusNotes = checklistStatusNotesRAWVALUE
				} else if common.ChecklistStatusNotes == checklistStatusNotesRAWVALUE {
					// Not sure why this happens, but it does. Not a problem.
					debug.Debug(debug.RMC, `%s: Two equal fields for common.ChecklistStatusNotes and security.CheckListStatusNotes:  value="%s"`, label, common.ChecklistStatusNotes)
				} else {
					debug.Warning(`%s: Two conflicting fields for ChecklistStatusNotes in RMC:  common.ChecklistStatusNotes="%s"  security.CheckListStatusNotes="%s"`, label, cleanStringForErrorMessage(common.ChecklistStatusNotes), cleanStringForErrorMessage(checklistStatusNotesRAWVALUE))
					debug.Debug(debug.RMC, `     --> common.ChecklistStatusNotes  ="%s"`, common.ChecklistStatusNotes)
					debug.Debug(debug.RMC, `     --> security.CheckListStatusNotes="%s"`, checklistStatusNotesRAWVALUE)
					//err := fmt.Errorf(`%s: Two conflicting fields for ChecklistStatusNotes in RMC:  common.ChecklistStatusNotes="%s"  security.CheckListStatusNotes="%s"`, label, common.ChecklistStatusNotes, checklistStatusNotesRAWVALUE)
					//debug.PrintError("%v", err)
					//return err
				}
			}
		} else {
			err := fmt.Errorf(`%s: Cannot obtain SOPCommon record in RMC entry:  %T  %#v`, label, out, out)
			debug.PrintError("%v", err)
			//return err XXX
		}
		return nil
	}

	for _, m := range maps {
		if id, ok := m[`id`]; ok {
			var err error
			if idString, ok2 := id.(string); ok2 {
				label2 := fmt.Sprintf("%s.%s", label, idString)
				switch idString {
				case `bss`:
					entry.SOPBSS = new(SOPBSS)
					err = decode(label2, m, entry.SOPBSS)
				case `legal`:
					entry.SOPLegal = new(SOPLegal)
					err = decode(label2, m, entry.SOPLegal)
				case `security`:
					entry.SOPSecurity = new(SOPSecurity)
					err = decode(label2, m, entry.SOPSecurity)
				case `operations`:
					entry.SOPOperations = new(SOPOperations)
					err = decode(label2, m, entry.SOPOperations)
				case `design`:
					entry.SOPDesign = new(SOPDesign)
					err = decode(label2, m, entry.SOPDesign)
				case `documentation`:
					entry.SOPDocumentation = new(SOPDocumentation)
					err = decode(label2, m, entry.SOPDocumentation)
				case `dedicated`:
					entry.SOPDedicated = new(SOPDedicated)
					err = decode(label2, m, entry.SOPDedicated)
				case `architecture`:
					entry.SOPArchitecture = new(SOPArchitecture)
					err = decode(label2, m, entry.SOPArchitecture)
				case `gtm`:
					entry.SOPGTM = new(SOPGTM)
					err = decode(label2, m, entry.SOPGTM)
				case `finalPlayback`:
					entry.SOPFinalPlayback = new(SOPFinalPlayback)
					err = decode(label2, m, entry.SOPFinalPlayback)
				case `goLiveApproval`:
					entry.SOPGoLiveApproval = new(SOPGoLiveApproval)
					err = decode(label2, m, entry.SOPGoLiveApproval)
				case `iam`:
					entry.SOPIAM = new(SOPIAM)
					err = decode(label2, m, entry.SOPIAM)
				case `businessContinuityDisasterRecovery`:
					entry.SOPBCDR = new(SOPBCDR)
					err = decode(label2, m, entry.SOPBCDR)
				case `ims`:
					entry.SOPIMS = new(SOPIMS)
					err = decode(label2, m, entry.SOPIMS)
				default:
					err := fmt.Errorf(`%s: found SOP sub-record with unknown id: "%s"`, label, idString)
					//debug.PrintError("%v", err)
					return err
				}
				if err != nil {
					err := debug.WrapError(err, `%s: Error parsing SOP sub-record`, label2)
					//debug.PrintError("%v", err)
					return err
				}
			} else {
				err := fmt.Errorf(`%s: found SOP sub-record with non-string id: %T "%#v"`, label, id, id)
				//debug.PrintError("%v", err)
				return err
			}
		} else {
			return fmt.Errorf(`%s: found SOP sub-record with no id: %v`, label, m)
		}
	}
	return nil
}

func cleanStringForErrorMessage(in string) string {
	result := strings.ReplaceAll(in, "\n", `\n`)
	if len(result) > 40 {
		return result[0:40] + "..."
	}
	return result
}

// ListRMCEntries lists all known RMC entries
// FIXME: work in progress
func ListRMCEntries(testMode bool) error {
	var actualURL string
	if testMode {
		actualURL = "https://api-test.rmc.test.cloud.ibm.com/v1/resources"
	} else {
		actualURL = "https://api.rmc.test.cloud.ibm.com/v1/resources"
	}
	key, err := rest.GetToken(rmcKeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for RMC")
		return err
	}
	var result = new(summaryResult)
	err = rest.DoHTTPGet(actualURL, key, nil, "RMC", debug.RMC, result)
	return err
}

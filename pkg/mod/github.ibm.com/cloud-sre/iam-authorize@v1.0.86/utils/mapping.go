package utils

import (
	"net/http"
	"strings"
)

type elem struct {
	IamResource   string
	WriteResource string
	Links         map[string]elem
}

var (
	iamResourceMap   = map[string]elem{}
	iamPermissionMap = map[string]string{
		http.MethodGet:    "pnp-api-oss.rest.get",
		http.MethodPost:   "pnp-api-oss.rest.post",
		http.MethodPatch:  "pnp-api-oss.rest.patch",
		http.MethodPut:    "pnp-api-oss.rest.put",
		http.MethodDelete: "pnp-api-oss.rest.delete",
	}
)

func init() {

	iamResourceMap["api"] = elem{IamResource: "", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["info"] = elem{IamResource: "", Links: make(map[string]elem)}

	iamResourceMap["api"].Links["v1"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"] = elem{IamResource: "", Links: map[string]elem{}}

	// catalog
	iamResourceMap["api"].Links["catalog"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["healthz"] = elem{IamResource: "", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["impls"] = elem{IamResource: "", WriteResource: "catalog1-impls-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["impl"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["impl"].Links["id"] = elem{IamResource: "", WriteResource: "catalog1-impls-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["subscriptions"] = elem{IamResource: "catalog1-subscriptions", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["subscription"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["subscription"].Links["id"] = elem{IamResource: "catalog1-subscriptions-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["categories"] = elem{IamResource: "", WriteResource: "catalog1-categories", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["category"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["catalog"].Links["category"].Links["id"] = elem{IamResource: "catalog1-categories-id", Links: make(map[string]elem)}

	// segmenttribes

	iamResourceMap["api"].Links["segmenttribe"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"] = elem{IamResource: "segmenttribe1-segments", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"] = elem{IamResource: "segmenttribe1-segments-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"].Links["tribes"] = elem{IamResource: "segmenttribe1-segments-id-tribes", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"].Links["tribes"].Links["id"] = elem{IamResource: "segmenttribe1-segments-id-tribes-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"].Links["tribes"].Links["id"].Links["services"] = elem{IamResource: "segmenttribe1-segments-id-tribes-id-services", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"].Links["tribes"].Links["id"].Links["services"].Links["id"] = elem{IamResource: "segmenttribe1-segments-id-tribes-id-services-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["services"] = elem{IamResource: "segmenttribe1-services", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["tribes"] = elem{IamResource: "segmenttribe1-tribes", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["production_readiness"] = elem{IamResource: "segmenttribe1-pr", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"].Links["production_readiness"] = elem{IamResource: "segmenttribe1-segments-id-pr", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"].Links["tribes"].Links["id"].Links["production_readiness"] = elem{IamResource: "segmenttribe1-segments-id-tribes-id-pr", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["segments"].Links["id"].Links["tribes"].Links["id"].Links["services"].Links["id"].Links["production_readiness"] = elem{IamResource: "segmenttribe1-segments-id-tribes-id-services-id-pr", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["segmenttribe"].Links["v1"].Links["rawservices"] = elem{IamResource: "segmenttribe1-rawservices", Links: make(map[string]elem)}

	// keyservice
	iamResourceMap["api"].Links["auth"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["auth"].Links["healthz"] = elem{IamResource: "", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["auth"].Links["tokens"] = elem{IamResource: "keys1-token", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["auth"].Links["apikeys"] = elem{IamResource: "keys1-apikeys", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["auth"].Links["apikeys"].Links["id"] = elem{IamResource: "keys1-apikeys-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["auth"].Links["publickey"] = elem{IamResource: "keys1-publickey", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["auth"].Links["logins"] = elem{IamResource: "keys1-logins", Links: make(map[string]elem)}

	// eventmgmt
	iamResourceMap["api"].Links["eventmgmt"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["eventmgmt"].Links["events"] = elem{IamResource: "eventmgmt1-events", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["eventmgmt"].Links["event"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["eventmgmt"].Links["event"].Links["id"] = elem{IamResource: "eventmgmt1-event-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["eventmgmt"].Links["event"].Links["id"].Links["audit_logs"] = elem{IamResource: "eventmgmt1-event-id-audit", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["eventmgmt"].Links["cicd"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["eventmgmt"].Links["cicd"].Links["events_flatten"] = elem{IamResource: "eventmgmt1-cicd-eventsflatten", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["eventmgmt"].Links["cicd"].Links["events_sql"] = elem{IamResource: "eventmgmt1-cicd-eventssql", Links: make(map[string]elem)}

	// incidentmgmt
	iamResourceMap["api"].Links["v1"].Links["incidentmgmt"] = elem{IamResource: "forbidden", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["incidentmgmt"].Links["concerns"] = elem{IamResource: "incidentmgmt1-concerns", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["incidentmgmt"].Links["journal"] = elem{IamResource: "incidentmgmt1-concerns", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["incidentmgmt"].Links["healthz"] = elem{IamResource: "", Links: map[string]elem{}}

	iamResourceMap["api"].Links["v1"].Links["incidentmgmtx"] = elem{IamResource: "forbidden", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["incidentmgmtx"].Links["chronology"] = elem{IamResource: "incidentmgmt1-concerns", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["incidentmgmtx"].Links["healthz"] = elem{IamResource: "", Links: map[string]elem{}}

	// doctor
	iamResourceMap["api"].Links["doctor"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["doctor"].Links["regionids"] = elem{IamResource: "doctor1-regionids", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["doctor"].Links["drs"] = elem{IamResource: "doctor1-drs", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["doctor"].Links["clientrecords"] = elem{IamResource: "doctor1-clientrecords", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["doctor"].Links["services"] = elem{IamResource: "doctor1-services", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["doctor"].Links["healthz"] = elem{IamResource: "", Links: make(map[string]elem)}

	// gcor
	iamResourceMap["api"].Links["v1"].Links["gcor"] = elem{IamResource: "forbidden", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["segments"] = elem{IamResource: "gcor1-segments", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["segments"].Links["id"] = elem{IamResource: "gcor1-segments-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["segments"].Links["id"].Links["services"] = elem{IamResource: "gcor1-segments-id-services", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["segments"].Links["id"].Links["tribes"] = elem{IamResource: "gcor1-segments-id-tribes", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["segments"].Links["id"].Links["productionReadiness"] = elem{IamResource: "gcor1-segments-id-prc", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["segmentExtended"] = elem{IamResource: "gcor1-segmentExtended", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["segmentExtended"].Links["id"] = elem{IamResource: "gcor1-segmentExtended-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["services"] = elem{IamResource: "gcor1-services", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["services"].Links["id"] = elem{IamResource: "gcor1-services-name", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["services"].Links["id"].Links["children"] = elem{IamResource: "gcor1-services-name-children", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["services"].Links["id"].Links["productionReadiness"] = elem{IamResource: "gcor1-services-name-prc", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["tribes"] = elem{IamResource: "gcor1-tribes", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["tribes"].Links["id"] = elem{IamResource: "gcor1-tribes-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["tribes"].Links["id"].Links["services"] = elem{IamResource: "gcor1-tribes-id-services", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["tribes"].Links["id"].Links["productionReadiness"] = elem{IamResource: "gcor1-tribes-id-prc", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["tribeExtended"] = elem{IamResource: "gcor1-tribeExtended", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["tribeExtended"].Links["id"] = elem{IamResource: "gcor1-tribeExtended-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["environments"] = elem{IamResource: "gcor1-envs", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["environments"].Links["id"] = elem{IamResource: "gcor1-envs-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["environmentExtended"] = elem{IamResource: "gcor1-envExtended", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["environmentExtended"].Links["id"] = elem{IamResource: "gcor1-envExtended-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["productionReadiness"] = elem{IamResource: "gcor1-prc", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["serviceExtended"] = elem{IamResource: "gcor1-serviceExtended", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["serviceExtended"].Links["id"] = elem{IamResource: "gcor1-serviceExtended-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["scorecardData"] = elem{IamResource: "gcor1-scorecardData", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["servicenowCI"] = elem{IamResource: "gcor1-servicenowci", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["gcor"].Links["servicenowCI"].Links["id"] = elem{IamResource: "gcor1-servicenowci-name", Links: map[string]elem{}}

	// concernsubscr
	iamResourceMap["api"].Links["concern"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["concern"].Links["subscriptions"] = elem{IamResource: "concern1-subscriptions", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["concern"].Links["subscriptions"].Links["id"] = elem{IamResource: "concern1-subscriptions-id", Links: make(map[string]elem)}
	iamResourceMap["api"].Links["concern"].Links["subscriptions"].Links["id"].Links["status"] = elem{IamResource: "concern1-subscriptions-id-status", Links: make(map[string]elem)}

	iamResourceMap["api"].Links["v1"].Links["pnp"] = elem{IamResource: "", Links: map[string]elem{}}

	// pnp-status
	iamResourceMap["api"].Links["v1"].Links["pnp"].Links["status"] = elem{IamResource: "forbidden", Links: map[string]elem{}}
	pnpStatus := iamResourceMap["api"].Links["v1"].Links["pnp"].Links["status"]

	pnpStatus.Links["healthz"] = elem{IamResource: "", Links: map[string]elem{}}
	pnpStatus.Links["resources"] = elem{IamResource: "pnp1-status-resources", Links: map[string]elem{}}
	pnpStatus.Links["resources"].Links["id"] = elem{IamResource: "pnp1-status-resources-id", Links: map[string]elem{}}
	pnpStatus.Links["incidents"] = elem{IamResource: "pnp1-status-incidents", Links: map[string]elem{}}
	pnpStatus.Links["incidents"].Links["id"] = elem{IamResource: "pnp1-status-incidents-id", Links: map[string]elem{}}
	// ATR added https://github.ibm.com/cloud-sre/toolsplatform/issues/8414
	pnpStatus.Links["incidents"].Links["id"].Links["affects"] = elem{IamResource: "pnp1-status-incidents-id-affects", Links: map[string]elem{}}
	// Everything that comes after the affects is authorized based on affects itself
	pnpStatus.Links["incidents"].Links["id"].Links["affects"].Links["ok-to-end"] = elem{IamResource: "pnp1-status-incidents-id-affects", Links: map[string]elem{}}
	pnpStatus.Links["maintenances"] = elem{IamResource: "pnp1-status-maintenances", Links: map[string]elem{}}
	pnpStatus.Links["maintenances"].Links["id"] = elem{IamResource: "pnp1-status-maintenances-id", Links: map[string]elem{}}
	// ATR added https://github.ibm.com/cloud-sre/toolsplatform/issues/8414
	pnpStatus.Links["maintenances"].Links["id"].Links["affects"] = elem{IamResource: "pnp1-status-maintenances-id-affects", Links: map[string]elem{}}
	pnpStatus.Links["maintenances"].Links["id"].Links["affects"].Links["ok-to-end"] = elem{IamResource: "pnp1-status-maintenances-id-affects", Links: map[string]elem{}}
	pnpStatus.Links["notifications"] = elem{IamResource: "pnp1-status-notifications", Links: map[string]elem{}}
	pnpStatus.Links["notifications"].Links["id"] = elem{IamResource: "pnp1-status-notifications-id", Links: map[string]elem{}}
	pnpStatus.Links["notifications"].Links["id"].Links["result"] = elem{IamResource: "pnp1-status-notifications-id-result", Links: map[string]elem{}}

	// pnp-subscriptions
	iamResourceMap["api"].Links["v1"].Links["pnp"].Links["subscriptions"] = elem{IamResource: "pnp1-subscriptions", Links: map[string]elem{}}
	pnpSub := iamResourceMap["api"].Links["v1"].Links["pnp"].Links["subscriptions"]

	pnpSub.Links["healthz"] = elem{IamResource: "", Links: map[string]elem{}}
	pnpSub.Links["id"] = elem{IamResource: "pnp1-subscriptions-id", Links: map[string]elem{}}
	pnpSubID := pnpSub.Links["id"]
	pnpSubID.Links["watches"] = elem{IamResource: "pnp1-subscriptions-id-watches", Links: map[string]elem{}}
	pnpSubWatch := pnpSubID.Links["watches"]
	pnpSubWatch.Links["id"] = elem{IamResource: "pnp1-subscriptions-id-watches-id", Links: map[string]elem{}}
	pnpSubWatch.Links["incidents"] = elem{IamResource: "pnp1-subscriptions-id-watches-incidents", Links: map[string]elem{}}
	pnpSubWatch.Links["maintenances"] = elem{IamResource: "pnp1-subscriptions-id-watches-maintenances", Links: map[string]elem{}}
	pnpSubWatch.Links["resources"] = elem{IamResource: "pnp1-subscriptions-id-watches-resources", Links: map[string]elem{}}
	pnpSubWatch.Links["case"] = elem{IamResource: "pnp1-subscriptions-id-watches-case", Links: map[string]elem{}}
	pnpSubWatch.Links["notifications"] = elem{IamResource: "pnp1-subscriptions-id-watches-notifications", Links: map[string]elem{}}

	// pnp-case
	iamResourceMap["api"].Links["v1"].Links["pnp"].Links["cases"] = elem{IamResource: "pnp1-cases", Links: map[string]elem{}}
	pnpCases := iamResourceMap["api"].Links["v1"].Links["pnp"].Links["cases"]
	pnpCases.Links["id"] = elem{IamResource: "pnp1-cases-id", Links: map[string]elem{}}
	pnpCasesID := pnpCases.Links["id"]
	pnpCasesID.Links["attachments"] = elem{IamResource: "pnp1-cases-id-attachments", Links: map[string]elem{}}
	pnpCasesIdattachments := pnpCasesID.Links["attachments"]
	pnpCasesIdattachments.Links["id"] = elem{IamResource: "pnp1-cases-id-attachments-id", Links: map[string]elem{}}
	pnpCasesID.Links["comments"] = elem{IamResource: "pnp1-cases-id-comments", Links: map[string]elem{}}
	pnpCasesID.Links["acceptance"] = elem{IamResource: "pnp1-cases-id-acceptance", Links: map[string]elem{}}
	pnpCases.Links["healthz"] = elem{IamResource: "", Links: map[string]elem{}}

	// pnp-ops-api - api/v1/pnp/ops/report
	iamResourceMap["api"].Links["v1"].Links["pnp"].Links["ops"] = elem{IamResource: "", Links: map[string]elem{}}
	pnpOpsAPI := iamResourceMap["api"].Links["v1"].Links["pnp"].Links["ops"]
	pnpOpsAPI.Links["report"] = elem{IamResource: "pnpops1-report", Links: map[string]elem{}}

	// issuecreator
	iamResourceMap["api"].Links["v1"].Links["issuecreator"] = elem{IamResource: "", Links: map[string]elem{}}
	issueCreator := iamResourceMap["api"].Links["v1"].Links["issuecreator"]
	issueCreator.Links["ahamapping"] = elem{IamResource: "issuecreator1-ahamapping", Links: map[string]elem{}}
	issueCreator.Links["createIssues"] = elem{IamResource: "issuecreator1-createIssues", Links: map[string]elem{}}
	issueCreator.Links["getSFIssues"] = elem{IamResource: "issuecreator1-getSFIssues", Links: map[string]elem{}}
	issueCreator.Links["getSFServices"] = elem{IamResource: "issuecreator1-getSFServices", Links: map[string]elem{}}
	issueCreator.Links["getSFIssuesByPillarFilter"] = elem{IamResource: "issuecreator1-getSFIssuesByPillarFilter", Links: map[string]elem{}}
	issueCreator.Links["getSFIssuesByPillar"] = elem{IamResource: "issuecreator1-getSFIssuesByPillar", Links: map[string]elem{}}
	issueCreator.Links["getWorkflows"] = elem{IamResource: "issuecreator1-getWorkflows", Links: map[string]elem{}}
	issueCreator.Links["healthz"] = elem{IamResource: "", Links: map[string]elem{}}

	// scorecardbackend
	iamResourceMap["api"].Links["v1"].Links["scorecardbackend"] = elem{IamResource: "", Links: map[string]elem{}}
	scorecardBackend := iamResourceMap["api"].Links["v1"].Links["scorecardbackend"]
	scorecardBackend.Links["CIEAvailability"] = elem{IamResource: "scorecardbackend1-CIEAvailability", Links: map[string]elem{}}
	scorecardBackend.Links["MappedCIEAvailability"] = elem{IamResource: "scorecardbackend1-MappedCIEAvailability", Links: map[string]elem{}}
	scorecardBackend.Links["edbAggregatedRollingMetrics"] = elem{IamResource: "scorecardbackend1-edbAggregatedRollingMetrics", Links: map[string]elem{}}
	scorecardBackend.Links["edbDailyAvailability"] = elem{IamResource: "scorecardbackend1-edbDailyAvailability", Links: map[string]elem{}}
	scorecardBackend.Links["edbMetricsCertification"] = elem{IamResource: "scorecardbackend1-edbMetricsCertification", Links: map[string]elem{}}
	scorecardBackend.Links["getReport"] = elem{IamResource: "scorecardbackend1-getReport", Links: map[string]elem{}}
	scorecardBackend.Links["sendReport"] = elem{IamResource: "scorecardbackend1-sendReport", Links: map[string]elem{}}
	scorecardBackend.Links["edbReportURLs"] = elem{IamResource: "scorecardbackend1-edbReportURLs", Links: map[string]elem{}}
	scorecardBackend.Links["edbRollingAvailability"] = elem{IamResource: "scorecardbackend1-edbRollingAvailability", Links: map[string]elem{}}
	scorecardBackend.Links["getOSSRecords"] = elem{IamResource: "scorecardbackend1-getOSSRecords", Links: map[string]elem{}}
	scorecardBackend.Links["getOSSValidations"] = elem{IamResource: "scorecardbackend1-getOSSValidations", Links: map[string]elem{}}
	scorecardBackend.Links["getTIPOnboardStatus"] = elem{IamResource: "scorecardbackend1-getTIPOnboardStatus", Links: map[string]elem{}}
	scorecardBackend.Links["certHealthStatus"] = elem{IamResource: "scorecardbackend1-certHealthStatus", Links: map[string]elem{}}
	scorecardBackend.Links["serviceCertList"] = elem{IamResource: "scorecardbackend1-serviceCertList", Links: map[string]elem{}}
	scorecardBackend.Links["serviceCertHealthStatus"] = elem{IamResource: "scorecardbackend1-serviceCertHealthStatus", Links: map[string]elem{}}
	scorecardBackend.Links["sendDailyMetrics"] = elem{IamResource: "scorecardbackend1-sendDailyMetrics", Links: map[string]elem{}}
	scorecardBackend.Links["healthz"] = elem{IamResource: "", Links: map[string]elem{}}

	// oauth
	iamResourceMap["api"].Links["v1"].Links["oauth"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["oauth"].Links["ibmcloud"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["oauth"].Links["ibmcloud"].Links["token"] = elem{IamResource: "oauth1-ibmcloud-token"}

	// tip-hooks
	iamResourceMap["hooks"] = elem{IamResource: "", Links: make(map[string]elem)}
	iamResourceMap["hooks"].Links["tip-alert"] = elem{IamResource: "tiphooks-tipalert", Links: make(map[string]elem)}
	iamResourceMap["hooks"].Links["tip-incident"] = elem{IamResource: "tiphooks-tipincident", Links: make(map[string]elem)}
	iamResourceMap["hooks"].Links["tip-newrelic"] = elem{IamResource: "tiphooks-tipnewrelic", Links: make(map[string]elem)}
	iamResourceMap["hooks"].Links["tip-prom"] = elem{IamResource: "tiphooks-tipprom", Links: make(map[string]elem)}
	iamResourceMap["hooks"].Links["tip-tokens"] = elem{IamResource: "tiphooks-tiptokens", Links: make(map[string]elem)}
	iamResourceMap["hooks"].Links["tip-sysdig"] = elem{IamResource: "tiphooks-tipsysdig", Links: make(map[string]elem)}
	iamResourceMap["hooks"].Links["tip-instana"] = elem{IamResource: "tiphooks-tipinstana", Links: make(map[string]elem)}

	// change
	iamResourceMap["api"].Links["v3"].Links["change_requests"] = elem{IamResource: "pnp3-changes", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"] = elem{IamResource: "pnp3-changes-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["comments"] = elem{IamResource: "pnp3-changes-id-comments", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["user_approvals"] = elem{IamResource: "pnp3-changes-id-userapprovals", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["group_approvals"] = elem{IamResource: "pnp3-changes-id-groupapprovals", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["processes"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["processes"].Links["request_approval"] = elem{IamResource: "pnp3-changes-id-proc-requestapproval", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["processes"].Links["implement"] = elem{IamResource: "pnp3-changes-id-proc-implement", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["processes"].Links["close"] = elem{IamResource: "pnp3-changes-id-proc-close", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["change_tasks"] = elem{IamResource: "pnp3-changes-id-tasks", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["change_tasks"].Links["id"] = elem{IamResource: "pnp3-changes-id-tasks-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_requests"].Links["id"].Links["publications"] = elem{IamResource: "pnp3-changes-id-publications", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["change_tasks"] = elem{IamResource: "pnp3-changetasks", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["publications"] = elem{IamResource: "pnp3-publications", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["publications"].Links["id"] = elem{IamResource: "pnp3-publications-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["publications"].Links["id"].Links["comments"] = elem{IamResource: "pnp3-publications-id-comments", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["publications"].Links["id"].Links["processes"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["publications"].Links["id"].Links["processes"].Links["request_approval"] = elem{IamResource: "pnp3-publications-id-proc-requestapproval", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["publications"].Links["id"].Links["processes"].Links["send"] = elem{IamResource: "pnp3-publications-id-proc-send", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["publications"].Links["id"].Links["change_requests"] = elem{IamResource: "pnp3-publications-id-changes", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["releases"] = elem{IamResource: "pnp3-releases", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["releases"].Links["id"] = elem{IamResource: "pnp3-releases-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["releases"].Links["id"].Links["change_requests"] = elem{IamResource: "pnp3-releases-id-changes", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["releases"].Links["id"].Links["change_tasks"] = elem{IamResource: "pnp3-releases-id-tasks", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["releases"].Links["id"].Links["change_tasks"].Links["id"] = elem{IamResource: "pnp3-releases-id-tasks-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v3"].Links["windows"] = elem{IamResource: "pnp3-windows", Links: map[string]elem{}}

	// bastion
	iamResourceMap["bastion"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["bastion"].Links["api"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}
	iamResourceMap["bastion"].Links["api"].Links["v1"] = elem{IamResource: "forbidden", Links: make(map[string]elem)}

	iamResourceMap["bastion"].Links["api"].Links["info"] = elem{IamResource: "bastion1-apiinfo", Links: map[string]elem{}}
	iamResourceMap["bastion"].Links["api"].Links["v1"].Links["accesscheck"] = elem{IamResource: "bastion1-accesscheck", Links: map[string]elem{}}
	iamResourceMap["bastion"].Links["api"].Links["v1"].Links["healthz"] = elem{IamResource: "bastion1-healthz", Links: map[string]elem{}}

	// phe
	iamResourceMap["api"].Links["v1"].Links["rules"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["rules"].Links["operations"] = elem{IamResource: "phe1-rules-operations", Links: map[string]elem{}}

	iamResourceMap["api"].Links["v1"].Links["phe"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["phe"].Links["datapoints"] = elem{IamResource: "phengine1-phe-datapoints", Links: map[string]elem{}}

	iamResourceMap["api"].Links["v1"].Links["phe"].Links["debug"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["phe"].Links["debug"].Links["pprof"] = elem{IamResource: "phengine1-phe-debug-pprof", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["phe"].Links["debug"].Links["dump"] = elem{IamResource: "phengine1-phe-debug-dump", Links: map[string]elem{}}

	// requested_items
	iamResourceMap["v1"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["v1"].Links["requested_items"] = elem{IamResource: "requesteditems1", Links: map[string]elem{}}
	iamResourceMap["v1"].Links["requested_items"].Links["id"] = elem{IamResource: "requesteditems1-id", Links: map[string]elem{}}
	iamResourceMap["v1"].Links["requested_items"].Links["id"].Links["user_approvals"] = elem{IamResource: "requesteditems1-id-userapprovals", Links: map[string]elem{}}
	iamResourceMap["v1"].Links["requested_items"].Links["id"].Links["group_approvals"] = elem{IamResource: "requesteditems1-id-groupapprovals", Links: map[string]elem{}}
	iamResourceMap["v1"].Links["requested_items"].Links["healthz"] = elem{IamResource: "requesteditems1-healthz", Links: map[string]elem{}}

	// OSS Kube Plugin: audit
	iamResourceMap["api"].Links["v1"].Links["audit"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["audit"].Links["entries"] = elem{IamResource: "auditentries1", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["audit"].Links["healthz"] = elem{IamResource: "auditentries1-healthz", Links: map[string]elem{}}

	// OSS Kube Plugin: chart
	iamResourceMap["api"].Links["v1"].Links["chart"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["chart"].Links["namespaces"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["chart"].Links["namespaces"].Links["id"] = elem{IamResource: "", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["chart"].Links["namespaces"].Links["id"].Links["releases"] = elem{IamResource: "chart1-namespaces-id-releases", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["chart"].Links["namespaces"].Links["id"].Links["releases"].Links["id"] = elem{IamResource: "chart1-namespaces-id-releases-id", Links: map[string]elem{}}
	iamResourceMap["api"].Links["v1"].Links["chart"].Links["healthz"] = elem{IamResource: "chart1-healthz", Links: map[string]elem{}}

	// End Mapping
}

func getIAMResource(req *http.Request) string {

	urlPath := req.URL.Path
	urlSegments := strings.Split(urlPath, "/")
	start := false // just tracks whether the url to map tracking has started

	var (
		currentMap  = iamResourceMap
		lastElement *elem
	)

	for index, val := range urlSegments {
		var element elem
		var ok bool

		if element, ok = currentMap[val]; ok {
			currentMap = currentMap[val].Links
			if !start {
				start = true
			}
		} else if element, ok = currentMap["id"]; ok {
			currentMap = currentMap["id"].Links
		} else if element, ok = currentMap["ok-to-end"]; ok {
			// Special case for PNP when we append TURL data
			return element.IamResource
		} else if start && val == "" {
			if req.Method != http.MethodGet && element.WriteResource != "" {
				return lastElement.WriteResource
			}
			return lastElement.IamResource
		} else if start {
			return "Error"
		}

		if index == (len(urlSegments) - 1) {
			if req.Method != http.MethodGet && element.WriteResource != "" {
				return element.WriteResource
			}
			return element.IamResource
		}
		lastElement = &element
	}

	// log.Print(currentMap)
	return "Error"
}

// GetIAMResourceAndPermission based on the request URI passed will parce it and make sure
// the resource exist and it is valid using the iamResourceMap
func GetIAMResourceAndPermission(req *http.Request) (resource string, permission string) {
	resource = getIAMResource(req)
	if resource == "" {
		return "", ""
	}
	return resource, iamPermissionMap[req.Method]
}

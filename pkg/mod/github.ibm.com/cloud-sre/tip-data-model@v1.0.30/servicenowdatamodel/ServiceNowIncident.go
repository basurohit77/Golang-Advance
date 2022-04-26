package servicenowdatamodel

// ServiceNowIncident - generated from SN (maybe outdated)
type ServiceNowIncident struct {
	Action                 string `json:"_action"`
	RecordType             string `json:"_record_type"`
	Active                 string `json:"active"`
	ActivityDue            string `json:"activity_due"`
	AdditionalAssigneeList string `json:"additional_assignee_list"`
	Approval               string `json:"approval"`
	ApprovalHistory        string `json:"approval_history"`
	ApprovalSet            string `json:"approval_set"`
	AssignedTo             struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"assigned_to"`
	AssignmentGroup struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"assignment_group"`
	BusinessDuration string `json:"business_duration"`
	BusinessService  struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"business_service"`
	BusinessStc      string `json:"business_stc"`
	CalendarDuration string `json:"calendar_duration"`
	CalendarStc      string `json:"calendar_stc"`
	CallerID         struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"caller_id"`
	Category string `json:"category"`
	CausedBy struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"caused_by"`
	ChildIncidents string `json:"child_incidents"`
	CloseCode      string `json:"close_code"`
	CloseNotes     string `json:"close_notes"`
	ClosedAt       string `json:"closed_at"`
	ClosedBy       struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"closed_by"`
	CmdbCi struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"cmdb_ci"`
	Comments             string `json:"comments"`
	CommentsAndWorkNotes string `json:"comments_and_work_notes"`
	Company              struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"company"`
	ContactType string `json:"contact_type"`
	Contract    struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"contract"`
	CorrelationDisplay string `json:"correlation_display"`
	CorrelationID      string `json:"correlation_id"`
	DeliveryPlan       struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"delivery_plan"`
	DeliveryTask struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"delivery_task"`
	Description   string `json:"description"`
	DueDate       string `json:"due_date"`
	Escalation    string `json:"escalation"`
	ExpectedStart string `json:"expected_start"`
	FollowUp      string `json:"follow_up"`
	GroupList     string `json:"group_list"`
	HoldReason    string `json:"hold_reason"`
	Impact        string `json:"impact"`
	IncidentState string `json:"incident_state"`
	Knowledge     string `json:"knowledge"`
	Location      struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"location"`
	MadeSLA  string `json:"made_sla"`
	Notify   string `json:"notify"`
	Number   string `json:"number"`
	OpenedAt string `json:"opened_at"`
	OpenedBy struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"opened_by"`
	Order  string `json:"order"`
	Parent struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"parent"`
	ParentIncident struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"parent_incident"`
	Priority  string `json:"priority"`
	ProblemID struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"problem_id"`
	ReassignmentCount string `json:"reassignment_count"`
	RejectionGoto     struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"rejection_goto"`
	ReopenCount string `json:"reopen_count"`
	ResolvedAt  string `json:"resolved_at"`
	ResolvedBy  struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"resolved_by"`
	Rfc struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"rfc"`
	Severity          string `json:"severity"`
	ShortDescription  string `json:"short_description"`
	Skills            string `json:"skills"`
	SLADue            string `json:"sla_due"`
	State             string `json:"state"`
	Subcategory       string `json:"subcategory"`
	SysClassName      string `json:"sys_class_name"`
	SysCreatedBy      string `json:"sys_created_by"`
	SysCreatedOn      string `json:"sys_created_on"`
	SysDomain         string `json:"sys_domain"`
	SysDomainPath     string `json:"sys_domain_path"`
	SysID             string `json:"sys_id"`
	SysModCount       string `json:"sys_mod_count"`
	SysUpdatedBy      string `json:"sys_updated_by"`
	SysUpdatedOn      string `json:"sys_updated_on"`
	TimeWorked        string `json:"time_worked"`
	UAffectedActivity string `json:"u_affected_activity"`
	UAudience         string `json:"u_audience"`
	UCase             struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"u_case"`
	UCurrentStatus             string `json:"u_current_status"`
	UCustomerImpactedQuestion  string `json:"u_customer_impacted_question"`
	UCustomerNames             string `json:"u_customer_names"`
	UCustomersImpacted         string `json:"u_customers_impacted"`
	UDataCenter                string `json:"u_data_center"`
	UDescriptionCustomerImpact string `json:"u_description_customer_impact"`
	UDetectionSource           string `json:"u_detection_source,omitempty"`
	UDisruptionBegan           string `json:"u_disruption_began"`
	UDisruptionEnded           string `json:"u_disruption_ended"`
	UDisruptionTime            string `json:"u_disruption_time"`
	UEnvironment               string `json:"u_environment"`
	UMonitoringIncidentNumber  string `json:"u_monitoring_incident_number,omitempty"`
	UMonitoringSituation       string `json:"u_monitoring_situation,omitempty"`
	UNcAlertkey                string `json:"u_nc_alertkey"`
	UNcApplication             string `json:"u_nc_application"`
	UNcNode                    string `json:"u_nc_node"`
	UNcPdEscalationPolicy      string `json:"u_nc_pd_escalation_policy"`
	UNcPdIncidentAssign        string `json:"u_nc_pd_incident_assign"`
	UNcPdIncidentNumber        string `json:"u_nc_pd_incident_number"`
	UNcPdOncall                string `json:"u_nc_pd_oncall"`
	UOpenAlert                 string `json:"u_open_alert,omitempty"`
	UOpenRunbook               string `json:"u_open_runbook,omitempty"`
	URecurringEvent            string `json:"u_recurring_event"`
	URegion                    string `json:"u_region"`
	USummary                   string `json:"u_summary"`
	UTribe                     string `json:"u_tribe"`
	UponApproval               string `json:"upon_approval"`
	UponReject                 string `json:"upon_reject"`
	Urgency                    string `json:"urgency"`
	UserInput                  string `json:"user_input"`
	Variables                  string `json:"variables"`
	WatchList                  string `json:"watch_list"`
	WfActivity                 struct {
		Link  string `json:"link"`
		Value string `json:"value"`
	} `json:"wf_activity"`
	WorkEnd                               string `json:"work_end"`
	WorkNotes                             string `json:"work_notes"`
	WorkNotesList                         string `json:"work_notes_list"`
	WorkStart                             string `json:"work_start"`
	XPdIntegrationIncident                string `json:"x_pd_integration_incident"`
	XPdIntegrationIncidentKey             string `json:"x_pd_integration_incident_key"`
	XPdIntegrationNotesIds                string `json:"x_pd_integration_notes_ids"`
	XPdIntegrationPagerdutyDisableTrigger string `json:"x_pd_integration_pagerduty_disable_trigger,omitempty"`
}

type ServiceNowIncidents []ServiceNowIncident

type ServiceNowResult struct {
	Result ServiceNowIncidents `json:"result"`
}

type ServiceNowIncidents2 []ServiceNowIncident2

type ServiceNowResult2 struct {
	Result ServiceNowIncidents2 `json:"results"`
}

type DisplayValue struct {
	DisplayValue string `json:"display_value"`
	Value        string `json:"value"`
	Link         string `json:"link,omitempty"`
}

// A safe action on pointer values to retrieve the 'Value'
func (dv *DisplayValue) GetValue() string {
	if dv == nil {
		return ""
	}
	return dv.Value
}

// A safe action on pointer values to retrieve the 'DisplayValue'
func (dv *DisplayValue) GetDisplayValue() string {
	if dv == nil {
		return ""
	}
	return dv.DisplayValue
}

// ServiceNowDisplayIncident - Incident information from service now when
// sysparm_display_value=all is specified on the query
type ServiceNowDisplayIncident struct {
	UAffectedActivity                     *DisplayValue `json:"u_affected_activity"`
	AssignedTo                            *DisplayValue `json:"assigned_to,omitempty"`
	UAudience                             *DisplayValue `json:"u_audience,omitempty"`
	CloseCode                             *DisplayValue `json:"close_code,omitempty"`
	CloseNotes                            *DisplayValue `json:"close_notes,omitempty"`
	CorrelationId                         *DisplayValue `json:"correlation_id,omitempty"`
	CmdbCi                                *DisplayValue `json:"cmdb_ci,omitempty"`
	Description                           *DisplayValue `json:"description,omitempty"`
	Number                                *DisplayValue `json:"number,omitempty"`
	Priority                              *DisplayValue `json:"priority,omitempty"`
	ResolvedBy                            *DisplayValue `json:"resolved_by,omitempty"`
	ShortDescription                      *DisplayValue `json:"short_description,omitempty"`
	State                                 *DisplayValue `json:"state,omitempty"`
	SysID                                 *DisplayValue `json:"sys_id,omitempty"`
	UCustomerImpactedQuestion             *DisplayValue `json:"u_customer_impacted_question,omitempty"`
	UCustomerNames                        *DisplayValue `json:"u_customer_names"`
	UDescriptionCustomerImpact            *DisplayValue `json:"u_description_customer_impact"`
	UDetectionSource                      *DisplayValue `json:"u_detection_source,omitempty"`
	UDisruptionBegan                      *DisplayValue `json:"u_disruption_began"`
	UDisruptionEnded                      *DisplayValue `json:"u_disruption_ended"`
	UMonitoringIncidentNumber             *DisplayValue `json:"u_monitoring_incident_number,omitempty"`
	UMonitoringSituation                  *DisplayValue `json:"u_monitoring_situation,omitempty"`
	UOpenAlert                            *DisplayValue `json:"u_open_alert,omitempty"`
	UOpenRunbook                          *DisplayValue `json:"u_open_runbook,omitempty"`
	UStatus                               *DisplayValue `json:"u_status,omitempty"`
	UUserFeedback                         *DisplayValue `json:"u_user_feedback,omitempty"`
	XPdIntegrationPagerdutyDisableTrigger *DisplayValue `json:"x_pd_integration_pagerduty_disable_trigger,omitempty"`
	SysCreatedOn                          *DisplayValue `json:"sys_created_on"`
}

type ServiceNowIncident2 struct {
	AssignedTo                            *string  `json:"assigned_to,omitempty"`
	AssignedToUserName                    *string  `json:"assigned_to_user_name,omitempty"`
	CloseCode                             *string  `json:"close_code,omitempty"`
	CloseNotes                            *string  `json:"close_notes,omitempty"`
	CmdbCi                                *string  `json:"cmdb_ci,omitempty"`
	CmdbCiName                            *string  `json:"cmdb_ci_name,omitempty"`
	CorrelationId                         *string  `json:"correlation_id,omitempty"`
	Description                           *string  `json:"description,omitempty"`
	Number                                *string  `json:"number,omitempty"`
	Priority                              *string  `json:"priority,omitempty"`
	ResolvedBy                            *string  `json:"resolved_by,omitempty"`
	ResolvedByUserName                    *string  `json:"resolved_by_user_name,omitempty"`
	ShortDescription                      *string  `json:"short_description,omitempty"`
	State                                 *string  `json:"state,omitempty"`
	SysCreatedOn                          *string  `json:"sys_created_on"`
	SysID                                 *string  `json:"sys_id,omitempty"`
	UAffectedActivity                     *string  `json:"u_affected_activity"`
	UAudience                             *string  `json:"u_audience,omitempty"`
	UCrns                                 []string `json:"u_crn,omitempty"`
	UCustomerImpactedQuestion             *string  `json:"u_customer_impacted_question,omitempty"`
	UCustomerNames                        *string  `json:"u_customer_names"`
	UDescriptionCustomerImpact            *string  `json:"u_description_customer_impact"`
	UDetectionSource                      *string  `json:"u_detection_source,omitempty"`
	UDisruptionBegan                      *string  `json:"u_disruption_began"`
	UDisruptionEnded                      *string  `json:"u_disruption_ended"`
	UImpactAdjustmentFactor               *string  `json:"u_impact_adjustment_factor,omitempty"`
	UMonitoringIncidentNumber             *string  `json:"u_monitoring_incident_number,omitempty"`
	UMonitoringSituation                  *string  `json:"u_monitoring_situation,omitempty"`
	UOpenAlert                            *string  `json:"u_open_alert,omitempty"`
	UOpenRunbook                          *string  `json:"u_open_runbook,omitempty"`
	UStatus                               *string  `json:"u_status,omitempty"`
	UTotalUnits                           *string  `json:"u_total_units,omitempty"`
	UTribeName                            *string  `json:"u_tribe_name,omitempty"`
	UUnitsAffected                        *string  `json:"u_units_affected,omitempty"`
	UUnitType                             *string  `json:"u_unit_type,omitempty"`
	UUserFeedback                         *string  `json:"u_user_feedback,omitempty"`
	XPdIntegrationPagerdutyDisableTrigger *string  `json:"x_pd_integration_pagerduty_disable_trigger,omitempty"`
}

type ServiceNowDisplayIncidents []ServiceNowDisplayIncident

type ServiceNowDisplayResult struct {
	Result ServiceNowDisplayIncidents `json:"result"`
}

//
// ServiceNow Payload for create/update of an incident in SN
// Only fields below are send to ServiceNow.
//
type ServiceNowPayload struct {
	AssignedTo                 string `json:"assigned_to,omitempty"`
	AssignmentGroup            string `json:"assignment_group,omitempty"`
	CloseCode                  string `json:"close_code,omitempty"`
	CloseNotes                 string `json:"close_notes,omitempty"`
	CorrelationID              string `json:"correlation_id,omitempty"`
	CmdbCi                     string `json:"cmdb_ci,omitempty"`
	Comments                   string `json:"comments,omitempty"`
	Description                string `json:"description,omitempty"`
	DisablePager               string `json:"x_pd_integration_pagerduty_disable_trigger,omitempty"`
	Priority                   int32  `json:"priority,omitempty"`
	ResolvedBy                 string `json:"resolved_by,omitempty"`
	ShortDescription           string `json:"short_description,omitempty"`
	State                      int32  `json:"state,omitempty"`
	UAffectedActivity          string `json:"u_affected_activity,omitempty"`
	UAudience                  string `json:"u_audience,omitempty"`
	UCustomerNames             string `json:"u_customer_names,omitempty"`
	UCustomerImpactedQuestion  string `json:"u_customer_impacted_question,omitempty"`
	UDescriptionCustomerImpact string `json:"u_description_customer_impact,omitempty"`
	UDisruptionBegan           string `json:"u_disruption_began,omitempty"`
	UDisruptionEnded           string `json:"u_disruption_ended,omitempty"`
	UEnvironment               string `json:"u_environment,omitempty"`
	UImpactAdjustmentFactor    string `json:"u_impact_adjustment_factor,omitempty"`
	UMonitoringSituation       string `json:"u_monitoring_situation,omitempty"`
	UMonitoringIncidentNumber  string `json:"u_monitoring_incident_number,omitempty"`
	UOpenAlert                 string `json:"u_open_alert,omitempty"`
	UOpenRunbook               string `json:"u_open_runbook,omitempty"`
	UpdatedBy                  string `json:"updated_by,omitempty"`
	UStatus                    int32  `json:"u_status,omitempty"`
	UDetectionSource           string `json:"u_detection_source,omitempty"`
	UTotalUnits                string `json:"u_total_units,omitempty"`
	UUnitsAffected             string `json:"u_units_affected,omitempty"`
	UUnitType                  string `json:"u_unit_type,omitempty"`
	UUserFeedback              string `json:"u_user_feedback,omitempty"`
	WorkNotes                  string `json:"work_notes,omitempty"`
}

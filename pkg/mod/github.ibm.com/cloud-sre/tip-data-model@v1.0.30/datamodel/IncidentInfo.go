/*
 * Incident Management
 *
 * The Technical Integration Point (TIP) Incident Management API.  This specification covers the incident management interface.
 *
 * OpenAPI spec version: 0.0.2
 *
 * Generated by: https://github.com/swagger-api/swagger-codegen.git
 */

// Note: modified package from the generated "TipIncApi" to a more generic one
package datamodel

// Errors that come from using addEscToIncident in the IMAPI
type IncidentError struct {
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	Detail   string `json:"detail,omitempty"`
	PagerURL string `json:"pager_url,omitempty"`
}

// Returned response from the IMAPI
type IncidentInfo struct {
	IncidentId     string               `json:"incidentId,omitempty"`
	UiUrl          string               `json:"uiUrl,omitempty"`
	IncidentErrors []IncidentError      `json:"errors,omitempty"`
	Extras         []IncidentInfoExtras `json:"extras,omitempty"`
	PagerURL       string               `json:"pager_url,omitempty"`
}

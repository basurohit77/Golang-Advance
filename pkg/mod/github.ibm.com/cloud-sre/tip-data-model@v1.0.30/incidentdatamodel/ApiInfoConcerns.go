/*
 * Incident Management
 *
 * The Technical Integration Point (TIP) Incident Management API.  This specification covers the incident management interface.
 *
 * OpenAPI spec version: 0.0.2
 *
 * Generated by: https://github.com/swagger-api/swagger-codegen.git
 */

package incidentdatamodel

type ApiInfoConcerns struct {

	// The URL for the 'https://&lt;domain&gt;/api/incidentmgmt/concerns' interface.
	Href string `json:"href,omitempty"`
}
type HealthzInfo struct {
	Href      string          `json:"href"`
	Code 	  int             `json:"code"`
	Description string        `json:"description"`
}
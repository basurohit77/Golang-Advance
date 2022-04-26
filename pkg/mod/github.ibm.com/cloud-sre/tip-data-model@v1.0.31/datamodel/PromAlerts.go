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

// REQUIRED. The CRN of the affected service if known.  See format is https&colon;//github.ibm.com/ibmcloud/builders-guide/blob/master/specifications/crn/CRN.md
type Alerts struct {
        Annotations PromAnnotations `json:"annotations,omitempty"`
}

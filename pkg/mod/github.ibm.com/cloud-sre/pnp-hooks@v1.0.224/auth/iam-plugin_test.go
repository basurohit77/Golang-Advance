package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetIAMResourceFromRequestForBspn tests the GetIAMResourceFromRequest function with /api/v1/snow/bspn
func TestGetIAMResourceFromRequestForBspn(t *testing.T) {
	plugin := IAMResourceActionPluginType{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snow/bspn", nil)
	resource := plugin.GetIAMResourceFromRequest(req)
	AssertEqual(t, "bspn resource", "pnphook1-snow-bspn", resource)
}

// TestGetIAMResourceFromRequestForCases tests the GetIAMResourceFromRequest function with /api/v1/snow/cases
func TestGetIAMResourceFromRequestForCases(t *testing.T) {
	plugin := IAMResourceActionPluginType{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snow/cases", nil)
	resource := plugin.GetIAMResourceFromRequest(req)
	AssertEqual(t, "cases resource", "pnphook1-snow-cases", resource)
}

// TestGetIAMResourceFromRequestForChanges tests the GetIAMResourceFromRequest function with /api/v1/snow/changes
func TestGetIAMResourceFromRequestForChanges(t *testing.T) {
	plugin := IAMResourceActionPluginType{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snow/changes", nil)
	resource := plugin.GetIAMResourceFromRequest(req)
	AssertEqual(t, "changes resource", "pnphook1-snow-chgs", resource)
}

// TestGetIAMResourceFromRequestForIncidents tests the GetIAMResourceFromRequest function with /api/v1/snow/incidents
func TestGetIAMResourceFromRequestForIncidents(t *testing.T) {
	plugin := IAMResourceActionPluginType{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snow/incidents", nil)
	resource := plugin.GetIAMResourceFromRequest(req)
	AssertEqual(t, "incidents resource", "pnphook1-snow-incs", resource)
}

// TestGetIAMResourceFromRequestNoResourceFound tests the GetIAMResourceFromRequest function when no resource is found
func TestGetIAMResourceFromRequestNoResourceFound(t *testing.T) {
	plugin := IAMResourceActionPluginType{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snow", nil)
	resource := plugin.GetIAMResourceFromRequest(req)
	AssertEqual(t, "no resource found", "", resource)
}

// TestGetIAMActionFromRequest tests the GetIAMActionFromRequest function
func TestGetIAMActionFromRequest(t *testing.T) {
	plugin := IAMResourceActionPluginType{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/snow/bspn", nil)
	action := plugin.GetIAMActionFromRequest(req)
	AssertEqual(t, "action", "pnp-api-oss.rest.post", action)
}

// TestGetIAMActionFromRequestNoActionFound tests the GetIAMActionFromRequest function when no action is found
func TestGetIAMActionFromRequestNoActionFound(t *testing.T) {
	plugin := IAMResourceActionPluginType{}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/snow/bspn", nil)
	action := plugin.GetIAMActionFromRequest(req)
	AssertEqual(t, "no action found", "", action)
}

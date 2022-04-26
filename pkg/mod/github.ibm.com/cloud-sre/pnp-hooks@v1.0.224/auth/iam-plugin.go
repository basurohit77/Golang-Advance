package auth

import (
	"net/http"
	"strings"
)

var iamPermissionMap = map[string]string{http.MethodPost: "pnp-api-oss.rest.post"}

type IAMResourceActionPluginType struct{}

func (iam IAMResourceActionPluginType) GetIAMResourceFromRequest(req *http.Request) (resource string) {
	urlPath := req.URL.Path
	switch {
	case strings.HasSuffix(urlPath, `/api/v1/snow/bspn`):
		return `pnphook1-snow-bspn`
	case strings.HasSuffix(urlPath, `/api/v1/snow/cases`):
		return `pnphook1-snow-cases`
	case strings.HasSuffix(urlPath, `/api/v1/snow/changes`):
		return `pnphook1-snow-chgs`
	case strings.HasSuffix(urlPath, `/api/v1/snow/incidents`):
		return `pnphook1-snow-incs`
	}

	return ``
}

func (iam IAMResourceActionPluginType) GetIAMActionFromRequest(req *http.Request) (action string) {
	return iamPermissionMap[req.Method]
}

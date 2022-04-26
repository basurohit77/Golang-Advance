package plugins

import (
	"net/http"
)

// IAMResourceActionPlugin - optional plugin that allows iam-authorize client code
// to override the default code used to obtain an IAM resource and action from a request.
type IAMResourceActionPlugin interface {

	// GetIAMResourceFromRequest - given the request, return the associated IAM resource:
	GetIAMResourceFromRequest(req *http.Request) (resource string)

	// GetIAMActionFromRequest - give the request, return the associated IAM action:
	GetIAMActionFromRequest(req *http.Request) (action string)

}

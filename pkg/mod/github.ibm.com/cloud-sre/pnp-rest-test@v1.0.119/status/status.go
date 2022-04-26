package status

import "github.ibm.com/cloud-sre/pnp-rest-test/catalog"

const (
	// APICategory is the category in the API platform for the status API
	APICategory = "pnpstatus"
)

// API is context for the status API
type API struct {
	apiInfo *APIInfo
	cat     *catalog.Catalog
}

// NewStatusAPI initializes a status API
func NewStatusAPI(cat *catalog.Catalog) (*API, error) {
	METHOD := "NewStatusAPI"

	url, err := cat.GetAPIImpl(METHOD, APICategory)
	if err != nil {
		return nil, err
	}

	info := new(APIInfo)
	err = cat.Server.GetAndDecode("NewStatusAPI", "APIInfo", url, info)
	if err != nil {
		return nil, err
	}

	return &API{cat: cat, apiInfo: info}, nil
}

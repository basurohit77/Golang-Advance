package catalog

import (
	"errors"

	"github.ibm.com/cloud-sre/pnp-rest-test/lg"
	"github.ibm.com/cloud-sre/pnp-rest-test/rest"
)

// Catalog is a API catalog representation
type Catalog struct {
	endpoint     string
	apiLookupURL string
	Server       *rest.Server
}

// NewCatalog returns a catalog object to use
func NewCatalog(endpoint string, server *rest.Server) *Catalog {
	return &Catalog{endpoint: endpoint, Server: server}
}

// GetAPIImpl returns the URL for an API info for a given category
func (cat *Catalog) GetAPIImpl(fct, category string) (url string, err error) {
	METHOD := fct + "->GetAPIImpl"

	if cat.apiLookupURL == "" {
		err = cat.getCatalogAPIInfo(METHOD)
		if err != nil {
			return "", err
		}
	}

	impls := new(RoleImplList)
	err = cat.Server.GetAndDecode(METHOD, "catalog.RoleImplList", cat.apiLookupURL+"?category="+category, impls)
	if err != nil {
		return "", err
	}

	if len(impls.Impls) != 1 {
		return "", lg.Err(METHOD, errors.New("api implementations problem"), "Catalog should have one implementation for category '%s' but has %d", category, len(impls.Impls))
	}

	if impls.Impls[0].SourceInfo.Href == "" {
		return "", lg.Err(METHOD, errors.New("no href"), "No sourceInfo found on API catalog")
	}

	return impls.Impls[0].SourceInfo.Href, nil
}

// getCatalogAPIInfo will retrieve the implementations URL from the catalog
func (cat *Catalog) getCatalogAPIInfo(fct string) error {
	METHOD := fct + "->getCatalogAPIInfo"

	info := new(APIInfo)
	err := cat.Server.GetAndDecode(METHOD, "catalog.APIInfo", cat.endpoint, info)
	if err != nil {
		return err
	}

	cat.apiLookupURL = info.APIImplementations.Href
	return nil
}

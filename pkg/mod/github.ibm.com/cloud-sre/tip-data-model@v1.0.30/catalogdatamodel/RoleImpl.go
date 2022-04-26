package catalogdatamodel

//
// Catalog API definitions for use of GET/POST/PATCH /api/catalog/impls
//
type Source struct {
	Href string `json:"href"`
}

type RoleImpl struct {
	Href       string   `json:"href"`
	Id         string   `json:"id"`
	ClientId   string   `json:"clientId"`
	Categories []string `json:"categories"`
	SourceInfo *Source  `json:"sourceInfo"`
}

type RoleImplList struct {
	Href  string      `json:"href"`
	Impls []*RoleImpl `json:"impls"`
}

type RoleImplInfo struct {
	ClientId   string   `json:"clientId"`
	Categories []string `json:"categories"`
	SourceInfo *Source  `json:"sourceInfo"`
}





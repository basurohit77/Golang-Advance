package catalogdatamodel

// Category represents a catagory in the API Catalog
type Category struct {
	HREF         string      `json:"href"`
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
}

// CategoryInfo represents a category to be registered with the API Catalog
type CategoryInfo struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
}

package cloudant

// NotificationsResult represents the full query result returned with all records
type NotificationsResult struct {
	TotalRows int                    `json:"total_rows"`
	Offset    int                    `json:"offset"`
	Rows      []*NotificationsRecord `json:"rows"`
}

// NotificationsRecord represents the individual raw record received from Cloudant when we query a notification
type NotificationsRecord struct {
	ID string `json:"id"`

	Doc struct {
		Title           string `json:"title"`
		Type            string `json:"type"` // This is like SECURITY, ANNOUNCEMENTS, etc
		Text            string `json:"text"`
		Category        string `json:"category"`    // This is like services, runtimes, etc
		SubCategory     string `json:"subCategory"` // This is the notification category ID
		RegionsAffected []*Region
		Archived        bool `json:"archived"`

		EventTime struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"eventTime"`

		Creation struct {
			Time string `json:"time"`
			// Not including email at this time since it identifies a user. However, if we
			// ever do, I found that sometimes email is not a string, but it is a struct.
			// Will need to write custom marshalling code to deal with that situation.
			//Email string `json:"email"`
		} `json:"creation"`

		LastUpdate struct {
			Time string `json:"time"`
			//Email string `json:"email"`
		} `json:"lastUpdate"`
	} `json:"doc"`
}

// Region represents a region tag in the cloudant record
type Region struct {
	ID string `json:"id"`
}

// NameMapping is a mapping between service names and notification category ID
type NameMapping struct {
	DisplayName string           `json:"displayName"`
	Components  []*NameComponent `json:"components"`
	Type        string           `json:"type"`
	ID          string           `json:"id"`
}

// NameComponent is an individual mapping entry
type NameComponent struct {
	DisplayName     string `json:"displayName"`
	EstadoDedicated string `json:"estado_dedicated"`
	ID              string `json:"id"`
	ServiceName     string `json:"serviceName"`
}

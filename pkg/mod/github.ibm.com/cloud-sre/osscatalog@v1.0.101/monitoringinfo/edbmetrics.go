package monitoringinfo

import (
	"fmt"
	"regexp"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// Reading the metrics collected in EDB

// edbMetricsURL is the URL from which to query all the available metric numbers from EDB, for one given date
// (date format: 2019-05-01)
//const edbMetricsURL = "https://pnp-api-oss.test.cloud.ibm.com/scorecardbackend/api/v1/edbDailyAvailability?date=%s"
//const edbMetricsURL = "https://pnp-api-oss.cloud.ibm.com/scorecardbackend/api/v1/edbRollingAvailability"
const edbMetricsURL = "https://pnp-api-oss.cloud.ibm.com/scorecardbackend/api/v1/scorecardbackend/edbRollingAvailability"

//const edbMetricsKeyName = "oss-apiplatform"
const edbMetricsKeyName = "catalog-yp"

// EDBMetricData represents one metric data point stored in EDB
type EDBMetricData struct {
	Service          string           `json:"service"`
	Location         string           `json:"location"`
	Plan             string           `json:"plan"`
	Component        string           `json:"component"`
	Type             EDBMetricType    `json:"type"`
	Segment          string           `json:"segment"`
	SegmentID        string           `json:"segmentID"`
	SegmentOwner     ossrecord.Person `json:"segmentOwner"`
	Tribe            string           `json:"tribe"`
	TribeID          string           `json:"tribeID"`
	TribeOwner       ossrecord.Person `json:"tribeOwner"`
	Status           string           `json:"status"`             // edbRollingAvailability only
	EntryType        string           `json:"entryType"`          // edbRollingAvailability only
	LastUpdate       string           `json:"last_update"`        // edbRollingAvailability only
	LastReport       float64          `json:"last_report"`        // edbRollingAvailability only
	Last1h           float64          `json:"last1h"`             // edbRollingAvailability only
	Last4h           float64          `json:"last4h"`             // edbRollingAvailability only
	Last24h          float64          `json:"last24h"`            // edbRollingAvailability only
	Last7d           float64          `json:"last7d"`             // edbRollingAvailability only
	Last30d          float64          `json:"last30d"`            // edbRollingAvailability only
	Uptime           float64          `json:"uptime"`             // edbRollingAvailability only
	Downtime         float64          `json:"downtime"`           // edbRollingAvailability only
	NotCollectedTime float64          `json:"not_collected_time"` // edbRollingAvailability only
	NoSLASLOImpact   bool             `json:"no_sla_slo_impact"`
	ServiceOwner     ossrecord.Person `json:"serviceOwner"`
	SNCIMapping      []string         `json:"sn_ci_mapping"`
	//Date             string           `json:"date"`
	//Availability     float64          `json:"availability"`       // edbDailyAvailability only
}

// EDBMetricType is the type of EDBMetric (provisioning or consumption)
type EDBMetricType string

// EDBMetricType is the type of EDBMetric (provisioning or consumption)
// Possible values for
const (
	EDBMetricTypeProvisioning EDBMetricType = "provisioning"
	EDBMetricTypeConsumption  EDBMetricType = "consumption"
)

// MetricType converts a EDBMetricType to ossrecord.MetricType
func (mt EDBMetricType) MetricType() ossrecord.MetricType {
	switch mt {
	case EDBMetricTypeProvisioning:
		return ossrecord.MetricProvisioning
	case EDBMetricTypeConsumption:
		return ossrecord.MetricConsumption
	default:
		return ossrecord.MetricOther
	}
}

// String returns a short string identifier for this EDBMetricData object
func (emd *EDBMetricData) String() string {
	return fmt.Sprintf("EDBMetricData(%s/%s/%s/%s)", emd.Service, emd.Plan, emd.Location, emd.Type)
}

// edbMetricDataGet is the container for the data returned by a GET on the EDBMetricsURL
type edbMetricDataGet struct {
	Resources struct {
		Provisioning []*EDBMetricData `json:"provisioning"`
		Consumption  []*EDBMetricData `json:"consumption"`
	} `json:"resources"`
}

// ListEDBMetricData lists all EDB Metrics entries and calls the specified handler function for each entry
func ListEDBMetricData(pattern *regexp.Regexp, handler func(e *EDBMetricData)) error {
	/*
		date := time.Now().Truncate(24 * time.Hour).Add(-24 * time.Hour).Format("2006-01-02")
		actualURL := fmt.Sprintf(edbMetricsURL, date)
	*/
	key, err := rest.GetKey(edbMetricsKeyName)
	if err != nil {
		err = debug.WrapError(err, "Cannot get key for EDB")
		return err
	}
	actualURL := edbMetricsURL
	var result = new(edbMetricDataGet)
	err = rest.DoHTTPGet(actualURL, key, nil, "EDB Metrics", debug.Monitoring, result)
	if err != nil {
		return err
	}

	for _, e := range result.Resources.Provisioning {
		if e.Type != EDBMetricTypeProvisioning {
			debug.PrintError("ListEDBMetricData(): Unexpected EDB metric type in EDB metric list: expected %s got %s", EDBMetricTypeProvisioning, e.String())
		}
		if pattern != nil && pattern.FindString(e.Service) == "" {
			continue
		}
		handler(e)
	}
	for _, e := range result.Resources.Consumption {
		if e.Type != EDBMetricTypeConsumption {
			debug.PrintError("ListEDBMetricData(): Unexpected EDB metric type in EDB metric list: expected %s got %s", EDBMetricTypeConsumption, e.String())
		}
		if pattern != nil && pattern.FindString(e.Service) == "" {
			continue
		}
		handler(e)
	}

	return nil
}

/*
https://pnp-api-oss.test.cloud.ibm.com/scorecardbackend/api/v1/edbDailyAvailability?date=2019-05-01
{
  "resources": {
    "provisioning": [
      {
        "service": "twilio-authy",
        "location": "eu-de",
        "plan": "user-provided",
        "type": "provisioning",
        "date": "2019-05-01",
        "segment": "",
        "tribe": "",
        "availability": 1
      },
      {
        "service": "twilio-authy",
        "location": "us-south",
        "plan": "user-provided",
        "type": "provisioning",
        "date": "2019-05-01",
        "segment": "",
        "tribe": "",
        "availability": 1
	  },
	],
	"consumption": [
      {
        "service": "is-network-acl",
        "location": "us-south",
        "plan": "",
        "type": "consumption",
        "date": "2019-05-01",
        "segment": "Infrastructure Services",
        "tribe": "Network",
        "availability": 0
      },
      {
        "service": "is-network-acl",
        "location": "us-south",
        "plan": "starter",
        "type": "consumption",
        "date": "2019-05-01",
        "segment": "Infrastructure Services",
        "tribe": "Network",
        "availability": 1
	  },
	]
  }
}
*/

/*
{
resources: {
provisioning: [
{
service: "fss-portfolio-service",
status: "EXPERIMENTAL",
entryType: "SERVICE",
location: "us-south",
plan: "fss-portfolio-service-free-plan",
type: "provisioning",
segment: "IBM Analytics",
tribe: "Financial Services",
last_update: "2019-05-03T11:30:24.086Z",
last_report: 1,
last1h: 1,
last4h: 1,
last24h: 1,
last7d: -1,
last30d: -1,
uptime: 98306.086319459,
downtime: 0,
not_collected_time: 0
},
consumption: [
{
service: "is-vpc",
status: "LIMITEDAVAILABILIY",
entryType: "SERVICE",
location: "us-south",
plan: "starter",
type: "consumption",
segment: "Infrastructure Services",
tribe: "Network",
last_update: "2019-05-03T11:30:17.893Z",
last_report: 1,
last1h: 1,
last4h: 1,
last24h: 1,
last7d: -1,
last30d: -1,
uptime: 276327.89344864,
downtime: 0,
not_collected_time: 0
},
*/

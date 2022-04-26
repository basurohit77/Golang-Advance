package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-hooks/health"
	"github.ibm.com/cloud-sre/tip-data-model/incidentdatamodel"
)

// HealthzHandler - healthz handler data
type HealthzHandler struct {
	ClientID             string
	KongURL              string
	APIRequestPathPrefix string
	SubPath              string
	catalogHealth        *health.CatalogHealth
	rabbitMQHealth       *health.RabbitMQHealth
	ciebotHealth   		 *health.CiebotHealth
}

// NewHealthzHandler - constructor for healthz handler
func NewHealthzHandler() *HealthzHandler {
	healthzHandler := &HealthzHandler{}
	healthzHandler.catalogHealth = health.NewCatalogHealth(os.Getenv("API_CATALOG_HEALTHZ_URL"))
	healthzHandler.rabbitMQHealth = health.NewRabbitMQHealth()
	healthzHandler.ciebotHealth = health.NewCiebotHealth(
		os.Getenv("CIEBOT_CONSUMER_HEALTHZ_URL"),
		os.Getenv("CIEBOT_WEBHOOK_HEALTHZ_URL"),
		os.Getenv("CIEBOT_HANDLER_HEALTHZ_URL"),
		os.Getenv("CIEBOT_SKIP_HEALTH_CHECK") == "true")
	return healthzHandler
}

// ServeHealthz - handle GET /healthz requests
//   Interface used by NewRelic monitoring
//
// Returns status 200 unless
//       - API Catalog /healthz is not reachable or returns code != 0, then status 503 is returned.
//       - RabbitMQ connection can not be made, then status 503 is returned.
//       - Internal error occurs, then status 500 is returned.
//
func (healthzHandler HealthzHandler) ServeHealthz(resp http.ResponseWriter, req *http.Request) {
	// intercept panics: print error and stacktrace
	defer api.HandlePanics(resp)

	// defer req.Body.Close()

	code, description := healthzHandler.getOverallStatus()

	resp.Header().Set("Content-Type", "application/json")
	if code == 0 {
		resp.WriteHeader(http.StatusOK)
	} else {
		resp.WriteHeader(http.StatusServiceUnavailable)
	}

	healthz := &incidentdatamodel.HealthzInfo{Href: healthzHandler.KongURL + "/" + healthzHandler.APIRequestPathPrefix + healthzHandler.SubPath, Code: code, Description: description}

	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(healthz); err != nil {
		// this would be a programming defect
		log.Fatal(healthzHandler.ClientID, " Could not encode catalog healthz info ", err)
		resp.WriteHeader(http.StatusInternalServerError)
	}

	if _, err := resp.Write(buffer.Bytes()); err != nil {
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

func (healthzHandler HealthzHandler) getOverallStatus() (int, string) {
	isHealthy, healthDescription := healthzHandler.rabbitMQHealth.CheckRabbitMQHealth()
	if !isHealthy {
		return 1, "MQ health check failed: " + healthDescription
	}

	isHealthy, healthDescription = healthzHandler.catalogHealth.CheckCatalogHealth()
	if !isHealthy {
		return 1, "API Catalog health check failed: " + healthDescription
	}

	isHealthy, healthDescription = healthzHandler.ciebotHealth.CheckCiebotHealth()
	if !isHealthy {
		// log.Println("INFO: Ciebot health check failed: " + healthDescription)
		return 1, "Ciebot health check failed: " + healthDescription
	}

	return 0, "The API is available and operational."
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.ibm.com/cloud-sre/pnp-abstraction/api"
	"github.ibm.com/cloud-sre/pnp-abstraction/monitoring"
	"github.ibm.com/cloud-sre/pnp-hooks/auth"
	"github.ibm.com/cloud-sre/pnp-hooks/handlers"
	"github.ibm.com/cloud-sre/pnp-hooks/shared"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	// connects to RabbitMQ
	connectRabbitmq()

	server := createServerObject()

	// create the authorization instance that will be used to authorize API calls:
	iamAuth := auth.NewIAMAuth(monitoring.GetNRApplication())
	authInstance := auth.NewAuth(iamAuth)

	// configure handlers
	setupHandlers(server, authInstance)

	server.StartWithoutMonitor()
}

func createServerObject() *api.Server {
	environment := os.Getenv("KUBE_APP_DEPLOYED_ENV")
	region := os.Getenv("KUBE_CLUSTER_REGION")
	monitoringKey := os.Getenv("NR_LICENSE")
	monitoringAppName := os.Getenv("MONITORING_APP_NAME")

	server := api.NewServerWithMonitor(monitoringAppName, monitoringKey, region, environment)
	server.Port = os.Getenv("PORT")
	server.SkipCatalogRegistration = true
	server.KongURL = os.Getenv("KONG_URL")
	// server.CatalogURL = os.Getenv("CATALOG_URL")
	// server.CatalogTimeout = getIntEnvVarValue("CATALOG_TIMEOUT")
	// server.CatalogCatagoryID = os.Getenv("CATALOG_CATEGORY_ID")
	// server.CatalogCatagoryName = os.Getenv("CATALOG_CATEGORY_NAME")
	// server.CatalogCatagoryDescription = os.Getenv("CATALOG_CATEGORY_DESCRIPTION")
	// server.CatalogClientID = os.Getenv("CATALOG_CLIENT_ID")
	// server.CatalogCheckRegInterval = getIntEnvVarValue("CATALOG_CHECK_REG_INTERVAL")
	return server
}

func createHealthzHandler(server *api.Server) *handlers.HealthzHandler {
	healthzHandler := handlers.NewHealthzHandler()
	// healthzHandler.ClientID = server.CatalogClientID
	healthzHandler.KongURL = server.KongURL
	healthzHandler.APIRequestPathPrefix = "pnphooks"
	healthzHandler.SubPath = shared.APIHealthzPath
	return healthzHandler
}

func connectRabbitmq() {
	err := shared.ConnectProducerRabbitMQ()
	if err != nil {
		log.Fatalln(err)
	}
}

func setupHandlers(server *api.Server, authInstance auth.Auth) {
	validMethods := make(map[string]string)
	validMethods[http.MethodPost] = http.MethodPost

	var endpointHandlers []api.EndpointHandler

	// Add Doctor /maintenances
	endpointHandlers = append(endpointHandlers, *api.CreateEndpointHandler(shared.APIDoctorMaintenancePath, handlers.ServeDoctorMaintenance, validMethods))

	// Add SNow /cases
	endpointHandlers = append(endpointHandlers, *api.CreateEndpointHandler(shared.APISnowCasesPath, authInstance.Middleware(handlers.ServeSnowCases), validMethods))

	// Add SNow /incidents
	endpointHandlers = append(endpointHandlers, *api.CreateEndpointHandler(shared.APISnowIncidentsPath, authInstance.Middleware(handlers.ServeSnowIncidents), validMethods))

	// Add SNow /bspn
	endpointHandlers = append(endpointHandlers, *api.CreateEndpointHandler(shared.APISnowBSPNPath, authInstance.Middleware(handlers.ServeSnowBSPN), validMethods))

	// Add SNow /changes
	endpointHandlers = append(endpointHandlers, *api.CreateEndpointHandler(shared.APISnowChangesPath, authInstance.Middleware(handlers.ServeSnowChanges), validMethods))

	// Add Ghe /announcement
	endpointHandlers = append(endpointHandlers, *api.CreateEndpointHandler(shared.APIGheAnnouncementsPath, handlers.ServeGheAnnouncements, validMethods))

	// Add /healthz
	healthzValidMethos := make(map[string]string)
	healthzValidMethos[http.MethodGet] = http.MethodGet
	healthzHandler := createHealthzHandler(server)
	endpointHandlers = append(endpointHandlers, *api.CreateEndpointHandler(shared.APIHealthzPath, healthzHandler.ServeHealthz, healthzValidMethos))

	server.EndpointHandlers = endpointHandlers

}

package api

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.ibm.com/cloud-sre/pnp-abstraction/monitoring"

	newrelic "github.com/newrelic/go-agent"
)

const (
	skipCatalogRegistrationOverride = true
)

// Server - Data, including options, for the server
type Server struct {
	Region                     string
	Environment                string
	Port                       string
	monitoringKey              string
	MonitoringAppName          string
	MonitoringApp              newrelic.Application
	KongURL                    string
	SkipCatalogRegistration    bool
	CatalogURL                 string
	CatalogTimeout             int
	CatalogCatagoryID          string
	CatalogCatagoryName        string
	CatalogCatagoryDescription string
	CatalogClientID            string
	CatalogCheckRegInterval    int
	EndpointHandlers           []EndpointHandler
	stacktraceBuffer           []byte
}

// EndpointHandler - handler for a specific path
type EndpointHandler struct {
	Path         string
	HandlerFunc  func(resp http.ResponseWriter, req *http.Request)
	ValidMethods map[string]string
}

// NewServer - constructor for creating a new API Server.
func NewServer(monitoringKey string) *Server {
	server := &Server{
		monitoringKey: monitoringKey,
	}
	return server
}

// NewServerWithMonitor - constructor for creating a new API Server with a new New Relic application.
func NewServerWithMonitor(monitoringAppName string, monitoringKey string, region string, environment string) *Server {
	monitoring.Initialize(monitoringAppName, monitoringKey, region, environment)
	server := &Server{
		MonitoringAppName: monitoringAppName,
		monitoringKey:     monitoringKey,
		MonitoringApp:     monitoring.GetNRApplication(),
		Region:            region,
		Environment:       environment,
	}
	return server
}

// Start - starts the new server
func (server *Server) Start() {
	HandleUnixSignals()
	printOutServerData(server)
	if !server.SkipCatalogRegistration && !skipCatalogRegistrationOverride {
		go registerWithCatalogAndMonitor(server.KongURL, server.CatalogURL, server.CatalogTimeout, server.CatalogCatagoryID, server.CatalogCatagoryName, server.CatalogCatagoryDescription, server.CatalogClientID, server.CatalogCheckRegInterval)
	}
	monitoring.Initialize(server.MonitoringAppName, server.monitoringKey, server.Region, server.Environment)
	registerEndPointHandlers(server.EndpointHandlers)
	startServer(server.Port, server.stacktraceBuffer)
}

// StartWithoutMonitor - starts the new server without a monitor
func (server *Server) StartWithoutMonitor() {
	HandleUnixSignals()
	printOutServerData(server)
	if !server.SkipCatalogRegistration && !skipCatalogRegistrationOverride {
		go registerWithCatalogAndMonitor(server.KongURL, server.CatalogURL, server.CatalogTimeout, server.CatalogCatagoryID, server.CatalogCatagoryName, server.CatalogCatagoryDescription, server.CatalogClientID, server.CatalogCheckRegInterval)
	}
	registerEndPointHandlers(server.EndpointHandlers)
	startServer(server.Port, server.stacktraceBuffer)
}

func printOutServerData(server *Server) {
	log.Print("SERVER DATA:")
	log.Print("Region: ", server.Region)
	log.Print("Environment: ", server.Environment)
	log.Print("Port: ", server.Port)
	log.Print("Monitoring key: ", server.monitoringKey[0:1]+"***...")
	log.Print("Monitoring app name: ", server.MonitoringAppName)
	log.Print("Kong URL: ", server.KongURL)
	log.Print("Skip catalog registration: ", server.SkipCatalogRegistration)
	log.Print("Catalog URL: ", server.CatalogURL)
	log.Print("Catalog timeout (sec): ", server.CatalogTimeout)
	log.Print("Catalog category id: ", server.CatalogCatagoryID)
	log.Print("Catalog category name: ", server.CatalogCatagoryName)
	log.Print("Catalog category description: ", server.CatalogCatagoryDescription)
	log.Print("Catalog client id: ", server.CatalogClientID)
	log.Print("Catalog check internal (sec): ", server.CatalogCheckRegInterval)
	log.Print("Endpoint handlers: ", server.EndpointHandlers)
}

func registerWithCatalogAndMonitor(kongURL string, catalogURL string, catalogTimeout int, catalogCategoryID string, catalogCategoryName string, catalogCategoryDescription string, catalogClientID string, catalogCheckRegInterval int) {
	const FCT = "monitor: "
	for {
		// On interval, check if still registered with catalog:
		registrationResult := RegisterAPIWithCatalog(kongURL, catalogURL, catalogTimeout, catalogCategoryID, catalogCategoryName, catalogCategoryDescription, catalogClientID)
		if !registrationResult {
			// TODO Can we send failure result to NewRelic?
			log.Print(FCT + "Failed to register sevice with API Catalog")
		}

		time.Sleep(time.Duration(catalogCheckRegInterval) * time.Second)

		// Allow other goroutines to proceed.
		runtime.Gosched()
	}
}

func registerEndPointHandlers(endpointHandlers []EndpointHandler) {
	const FCT = "registerEndPointHandlers: "
	log.Println(FCT + "Going to register handlers")
	for _, endpointHandler := range endpointHandlers {
		log.Println(FCT + "Registering handler for " + endpointHandler.Path)
		wrappedEndpointHandler := NewWrappedEndpointHandler(endpointHandler.HandlerFunc, endpointHandler.ValidMethods)
		monitoring.RegisterAndMonitorHandlerFunc(endpointHandler.Path, wrappedEndpointHandler.wrappedHandler)
	}
}

func startServer(port string, stacktraceBuffer []byte) {
	defer HandlePanicsForServer()
	log.Println("Starting server on port:", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Println("ERROR: API server has terminated, err=", err)
		time.Sleep(5 * time.Second)
		os.Exit(-1)
	}
}

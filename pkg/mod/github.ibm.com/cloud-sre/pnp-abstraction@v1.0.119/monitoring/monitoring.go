package monitoring

import (
	"github.com/newrelic/go-agent"
	loglm "github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
)

/**
 * This monitoring package provides support to monitor a Go application. The functions in this
 * package provide higher level abstractions that attempt to hide the actual monitoring system
 * that is being used. Currently, the monitoring system being used is NewRelic.
 *
 * An application first needs to call the Initialize function, and then is free to call any
 * other public function.
 **/

// Wrapper for a transaction:
type Transaction struct {
	Value newrelic.Transaction
}

// Keep track of the NewRelic application:
var (
	app          newrelic.Application
	region       string
	environment  string
	nrAppNameEnv = os.Getenv("NR_APPNAME")
)

// Keep track of the application name in NewRelic:
var appName string

/**
 * Initializes the monitoring package. The applicationName parameter is used to identify the
 * application in the monitoring system. The key parameter is the key needed to connect to
 * the monitoring system.
 **/
func Initialize(applicationName string, key string, regionArg string, environmentArg string) bool {
	const FCT = "monitoring Initialize: "
	region = regionArg
	environment = environmentArg
	if applicationName != "" && key != "" {
		if nrAppNameEnv != "" {
			applicationName = nrAppNameEnv
		}
		log.Println("Initializing monitoring for " + applicationName + " ...")
		config := newrelic.NewConfig(applicationName, key)
		newRelicApp, err := newrelic.NewApplication(config)
		if err != nil {
			loglm.Errorln(FCT+"ERROR: Failed to instantiate NewRelic. err: ", err)
			app = nil
			appName = ""
			return false
		} else {
			app = newRelicApp
			appName = applicationName
		}

	} else {
		// no error case, return true as parameters not specified
		log.Println("Not initializing monitoring for since application name or key was not provided")
		app = nil
		appName = ""
		return false
	}
	return true
}

/**
 * Registers the provided HTTP handler and monitors requests to the handler.
 *
 * The path parameter is the URL path to listen for, and the handler parameter is the http
 * handler function that should be called when a request for the path comes in.
 *
 **/
func RegisterAndMonitorHandlerFunc(path string, handler func(http.ResponseWriter, *http.Request)) {
	const FCT = "monitoring RegisterAndMonitorHandlerFunc: "
	if app != nil {
		log.Println("Monitoring handler function for: " + path)
		http.HandleFunc(newrelic.WrapHandleFunc(app, path, handler))
	} else {
		http.HandleFunc(path, handler)
	}
}

/**
 * Get transaction from response.
 **/
func GetTransaction(resp http.ResponseWriter) Transaction {
	const FCT = "monitoring GetTransaction: "
	if app != nil {
		txn, _ := resp.(newrelic.Transaction)
		return Transaction{txn}
	}
	return Transaction{nil}
}

/**
 * Adds a monitor onto the provided client so that the call can be measured in the
 * monitoring system as an external call.
 **/
func MonitorExternalCall(txn *Transaction, client *http.Client) {
	const FCT = "monitoring MonitorExternalCall: "
	if app != nil && txn != nil && txn.Value != nil {
		client.Transport = newrelic.NewRoundTripper(txn.Value, nil)
	}
}

/**
 * Adds a custom attribute to the transaction associated with the provided resp parameter.
 **/
func AddCustomAttribute(resp http.ResponseWriter, key string, value string) {
	const FCT = "monitoring AddCustomAttribute: "
	if app != nil {
		txn, _ := resp.(newrelic.Transaction)
		AddCustomAttributeTxn(txn, key, value)
	}
}

/**
 * Adds a custom attribute to the transaction associated with the provided resp parameter.
 **/
func AddInt32CustomAttribute(resp http.ResponseWriter, key string, value int32) {
	const FCT = "monitoring AddInt32CustomAttribute: "
	if app != nil {
		txn, _ := resp.(newrelic.Transaction)
		AddInt32CustomAttributeTxn(txn, key, value)
	}
}

/**
 * Adds a custom attribute to the transaction associated with the provided resp parameter.
 **/
func AddInt64CustomAttribute(resp http.ResponseWriter, key string, value int64) {
	const FCT = "monitoring AddInt64CustomAttribute: "
	if app != nil {
		txn, _ := resp.(newrelic.Transaction)
		AddInt64CustomAttributeTxn(txn, key, value)
	}
}

/**
 * Add a custom attribute to a transaction
 **/
func AddCustomAttributeTxn(txn newrelic.Transaction, key string, value string) {
	if txn != nil {
		if err := txn.AddAttribute(key, value); err != nil {
			log.Println(err)
		}
	}
}

/**
 * Adds a custom attribute to a transaction
 **/
func AddInt32CustomAttributeTxn(txn newrelic.Transaction, key string, value int32) {
	if txn != nil {
		if err := txn.AddAttribute(key, value); err != nil {
			log.Println(err)
		}
	}
}

/**
 * Adds a custom attribute to a transaction
 **/
func AddInt64CustomAttributeTxn(txn newrelic.Transaction, key string, value int64) {
	if txn != nil {
		if err := txn.AddAttribute(key, value); err != nil {
			log.Println(err)
		}
	}
}

/**
 * Adds a custom attribute to a transaction
 **/
func AddBoolCustomAttributeTxn(txn newrelic.Transaction, key string, value bool) {
	if txn != nil {
		if err := txn.AddAttribute(key, value); err != nil {
			log.Println(err)
		}
	}
}

// AddKubeAttributes add default attributes to a NewRelic APM transaction
// Kubernetes cluster region and Application environment as attributes to send to NewRelic
func AddKubeAttributes(res http.ResponseWriter) {
	if app != nil {
		txn, _ := res.(newrelic.Transaction)
		if txn != nil {
			err := txn.AddAttribute("apiKubeClusterRegion", region)
			if err != nil {
				log.Println(err)
			}

			err = txn.AddAttribute("apiKubeAppDeployedEnv", environment)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// GetNRApplication - get newRelic.Application
func GetNRApplication() newrelic.Application {
	return app
}

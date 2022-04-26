package exmon

import (
	"errors"
	"log"
	"os"
	"time"

	newrelic "github.com/newrelic/go-agent"
	"github.ibm.com/cloud-sre/pnp-abstraction/monitoring"
)

const (
	pnp              = "pnp-"
	notification     = "notification-"
	cloudantPkg      = "cloudant-"
	postgresPkg      = "postgres-"
	osscatalogPkg    = "osscatalog-"
	mqPkg            = "mq-"
	globalcatalogPkg = "globalcatalog-"
	durationAttr     = "Duration"
	starttimeAttr    = "StartTime"

	// CloudantGetNotifications is the transaction name for the method in the cloudant package for getting notifications.
	CloudantGetNotifications = TransactionName(pnp + notification + cloudantPkg + "GetNotifications")

	// PostgresGetNotifications is the transaction name for the method in the postgres package for getting notifications.
	PostgresGetNotifications = TransactionName(pnp + notification + postgresPkg + "GetNotifications")

	// OSSCatGetRecords is the transaction name for the method in the osscatalog package that retrieves records.
	OSSCatGetRecords = TransactionName(pnp + notification + osscatalogPkg + "GetRecords")

	// GlobalCatalogGetRecords is the transaction name for the method in the global catalog package that retrieves records.
	GlobalCatalogGetRecords = TransactionName(pnp + notification + globalcatalogPkg + "GetRecords")

	// MQProduceNotification is the transaction name for the method in the mq package that posts notifications.
	MQProduceNotification = TransactionName(pnp + notification + mqPkg + "SendNotifications")

	// PnPStartTime is a integer value that specifies the start time of the transaction.
	PnPStartTime = AttributeName(pnp + starttimeAttr)

	// PnPDuration is a integer value that specifies the duration of the transaction.
	PnPDuration = AttributeName(pnp + durationAttr)

	// OSSCatalogFailed is a bool attribute that indicates if there was a failure with the OSS catalog
	OSSCatalogFailed = AttributeName(pnp + "db-failed")

	// CloudantFailed is a bool attribute that indicates if there was a failure with the cloudant database
	CloudantFailed = AttributeName(pnp + "db-failed")

	// PostgresFailed is a bool attribute that indicates if there was a failure with the postgres database
	PostgresFailed = AttributeName(pnp + "db-failed")

	// GlobalCatalogFailed is a bool attribute that indicates if there was a failure with the global catalog database
	GlobalCatalogFailed = AttributeName(pnp + "db-failed")

	// MQPostFailed is a bool attribute that indicates if the post to MQ failed
	MQPostFailed = AttributeName(pnp + "mq-post-failed")

	// HttpGetFailed is a bool attribute that indicates if the http request failed
	HttpGetFailed = AttributeName(pnp + "http-get-failed")

	// HttpPostFailed is a bool attribute that indicates if the post to Http failed
	HttpPostFailed = AttributeName(pnp + "http-post-failed")

	// EncryptFailed is a bool attribute that indicates if the encryption failed
	EncryptFailed = AttributeName(pnp + "encrypt-failed")

	// DecryptFailed is a bool attribute that indicates if the decryption failed
	DecryptFailed = AttributeName(pnp + "decrypt-failed")

	// WatchMapFailed is a bool attribute that indicates if the WatchMap operation failed
	WatchMapFailed = AttributeName(pnp + "watchmap-failed")

	// WatchMapError is a string attribute that indicates how the WatchMap operation failed
	WatchMapError = AttributeName(pnp + "watchmap-error")

	// ParseFailed is a bool attribute indicates that a response was recieved, but it could not be parsed as expected.
	ParseFailed = AttributeName(pnp + "parse-failed")

	// RecordCount is an int32 attribute containing the count of records queried from the database.
	RecordCount = AttributeName(pnp + "record-count")

	// Source is the name of the source from inside the notification
	Source = AttributeName(pnp + "Source")

	// SourceID is the source ID from inside the notification
	SourceID = AttributeName(pnp + "SourceID")

	// Operation is the operation being performed such as insert or update
	Operation = AttributeName(pnp + "Operation")

	// APIKubeAppDeployedEnv is the deployed env
	APIKubeAppDeployedEnv = AttributeName("apiKubeAppDeployedEnv")

	// APIKubeClusterRegion is the cluster region
	APIKubeClusterRegion = AttributeName("apiKubeClusterRegion")

	// Kind describes the kind of record
	Kind = AttributeName("pnp-Kind")

	// OperationQuery is the operation when doing queries
	OperationQuery = "query"

	// OperationInsert is the operation when doing inserts
	OperationInsert = "insert"

	// OperationUpdate is the operation when doing updates
	OperationUpdate = "update"

	// KindNotification is the constant for the notification kind
	KindNotification = "notification"
)

// TransactionName is a type which is the transaction name
type TransactionName string

// AttributeName is a type specifying an attribute name
type AttributeName string

var (
	monitoringKey      = os.Getenv("NR_LICENSE")
	monitoringAppName  = os.Getenv("NR_APPNAME")
	kubeClusterRegion  = os.Getenv("KUBE_CLUSTER_REGION")
	kubeAppDeployedEnv = os.Getenv("KUBE_APP_DEPLOYED_ENV")
)

// Monitor is the structure to hold the monitor information.
type Monitor struct {
	created  bool
	NRConfig newrelic.Config
	NRApp    newrelic.Application
}

// ITransaction is an interface for a transaction object
type ITransaction interface {
	End()
}

// Transaction is an abstraction of the new relic transaction
type Transaction struct {
	Name  TransactionName
	NRTxn newrelic.Transaction
	Start time.Time
	Ended bool
}

// End will clean up a transaction
func (txn Transaction) End() {
	METHOD := "exmon.End"
	if txn.Name != "" && !txn.Ended {
		txn.Ended = true
		// Capture end time:
		endTime := time.Now()

		// Determine duration to report:
		duration := (endTime.UnixNano() - txn.Start.UnixNano()) / 1000000 // Milliseconds
		txn.AddInt64CustomAttribute(PnPDuration, duration)
		err := txn.NRTxn.End()
		if err != nil {
			log.Printf("ERROR (%s): %s", METHOD, err.Error())
		}
		//log.Printf("INFO (%s): Completed transaction %s", METHOD, txn.Name.String())
	}
}

// CreateMonitor creates a new monitor - primarily for new relic monitoring, but this allows
// unit testing
func CreateMonitor() (mon *Monitor, err error) {
	if newConfig == nil || newApplication == nil {
		return nil, errors.New("ERROR: (CreateMonitor): Adapter code not initialized. Please call initadapter.Initialize()")
	}
	nrConfig := newConfig(monitoringAppName, monitoringKey)
	nrApp, err := newApplication(nrConfig)

	if err != nil {
		return mon, err
	}

	mon = new(Monitor)
	mon.NRConfig = nrConfig
	mon.NRApp = nrApp

	return mon, nil
}

// StartTransaction is called to start a particular transaction, or entry in monitoring
func (mon *Monitor) StartTransaction(name TransactionName) (result Transaction) {
	if mon == nil {
		return result
	}
	nrtxn := mon.NRApp.StartTransaction(name.String(), nil, nil)

	txn := Transaction{Name: name, NRTxn: nrtxn, Start: time.Now()}
	txn.AddInt64CustomAttribute(PnPStartTime, txn.Start.Unix())

	txn.AddCustomAttribute(APIKubeClusterRegion, kubeClusterRegion)
	txn.AddCustomAttribute(APIKubeAppDeployedEnv, kubeAppDeployedEnv)

	return txn
}

func (tn TransactionName) String() string {
	return string(tn)
}

func (an AttributeName) String() string {
	return string(an)
}

// AddInt64CustomAttribute will add a custom attribute to a transaction
func (txn Transaction) AddInt64CustomAttribute(key AttributeName, value int64) {
	if txn.Name != "" {
		addInt64CustomAttribute(txn.NRTxn, key.String(), value)
	}
}

// AddIntCustomAttribute will add a custom attribute to a transaction
func (txn Transaction) AddIntCustomAttribute(key AttributeName, value int) {
	if txn.Name != "" {
		addInt32CustomAttribute(txn.NRTxn, key.String(), int32(value))
	}
}

// AddCustomAttribute will add a custom attribute to a transaction
func (txn Transaction) AddCustomAttribute(key AttributeName, value string) {
	if txn.Name != "" {
		addCustomAttribute(txn.NRTxn, key.String(), value)
	}
}

// AddBoolCustomAttribute will add a custom attribute to a transaction
func (txn Transaction) AddBoolCustomAttribute(key AttributeName, value bool) {
	if txn.Name != "" {
		addBoolCustomAttribute(txn.NRTxn, key.String(), value)
	}
}

// SetupMonitorFunctions is called to set up the default list of monitoring functions
func SetupMonitorFunctions() {
	newConfig = newrelic.NewConfig
	newApplication = newrelic.NewApplication
	addInt64CustomAttribute = monitoring.AddInt64CustomAttributeTxn
	addInt32CustomAttribute = monitoring.AddInt32CustomAttributeTxn
	addCustomAttribute = monitoring.AddCustomAttributeTxn
	addBoolCustomAttribute = monitoring.AddBoolCustomAttributeTxn
}

// SetFunctionPointers is a convenience method used primarily for unit testing purposes to set the various function pointers
// No need to call this during normal operation
func SetFunctionPointers(f1 NewConfigFunc, f2 NewApplicationFunc, f3 AddInt64CustomAttributeFunc, f4 AddInt32CustomAttributeFunc, f5 AddCustomAttributeFunc, f6 AddBoolCustomAttributeFunc) {
	newConfig = f1
	newApplication = f2
	addInt64CustomAttribute = f3
	addInt32CustomAttribute = f4
	addCustomAttribute = f5
	addBoolCustomAttribute = f6
}

// NewConfigFunc is a type of function used by new relic
type NewConfigFunc func(appname, license string) newrelic.Config

// NewConfig is the function pointer to the new relic function
var newConfig NewConfigFunc

// NewApplicationFunc is a type of function from new relic for monitoring
type NewApplicationFunc func(c newrelic.Config) (newrelic.Application, error)

// NewApplication is the pointer to the new relic function
var newApplication NewApplicationFunc

// AddInt64CustomAttributeFunc is a type of function used to add attributes to monitoring
type AddInt64CustomAttributeFunc func(txn newrelic.Transaction, key string, value int64)

// AddInt64CustomAttribute is the function pointer to the
var addInt64CustomAttribute AddInt64CustomAttributeFunc

// AddInt32CustomAttributeFunc is a type of function used to add attributes to monitoring
type AddInt32CustomAttributeFunc func(txn newrelic.Transaction, key string, value int32)

// AddInt32CustomAttribute is the function pointer to the
var addInt32CustomAttribute AddInt32CustomAttributeFunc

// AddCustomAttributeFunc is a type of function used to add attributes to monitoring
type AddCustomAttributeFunc func(txn newrelic.Transaction, key string, value string)

// AddCustomAttribute is a type of function used to add attributes to monitoring
var addCustomAttribute AddCustomAttributeFunc

// AddBoolCustomAttributeFunc is a type of function used to add attributes to monitoring
type AddBoolCustomAttributeFunc func(txn newrelic.Transaction, key string, value bool)

// AddBoolCustomAttribute is a type of function used to add attributes to monitoring
var addBoolCustomAttribute AddBoolCustomAttributeFunc

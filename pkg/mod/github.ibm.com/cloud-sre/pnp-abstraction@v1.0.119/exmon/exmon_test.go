package exmon

import (
	"net/http"
	"testing"
	"time"

	newrelic "github.com/newrelic/go-agent"
)

func TestSetup(t *testing.T) {
	SetupMonitorFunctions()

}

func TestBasic(t *testing.T) {
	SetFunctionPointers(exutNewConfig, exutNewApplication, exutAddInt64CustomAttribute, exutAddInt32CustomAttribute, exutAddCustomAttribute, exutAddBoolCustomAttribute)

	mon, err := CreateMonitor()
	if err != nil {
		t.Fatal(err)
	}

	txn := mon.StartTransaction(PostgresGetNotifications)
	txn.AddIntCustomAttribute(RecordCount, 8)
	txn.AddBoolCustomAttribute(PostgresFailed, true)
	txn.End()

	mon = nil
	mon.StartTransaction(PostgresGetNotifications) // This should not create an error
}

func exutNewConfig(appname, license string) (conf newrelic.Config) {
	return conf
}

func exutNewApplication(c newrelic.Config) (app newrelic.Application, err error) {
	return exutApplication{}, err
}

func exutAddInt64CustomAttribute(txn newrelic.Transaction, key string, value int64) {
}

func exutAddInt32CustomAttribute(txn newrelic.Transaction, key string, value int32) {
}

func exutAddCustomAttribute(txn newrelic.Transaction, key string, value string) {
}

func exutAddBoolCustomAttribute(txn newrelic.Transaction, key string, value bool) {
}

// -------- Application test type ---------------------
type exutApplication struct {
}

func (app exutApplication) StartTransaction(name string, w http.ResponseWriter, r *http.Request) newrelic.Transaction {
	return exutTransaction{}
}

func (app exutApplication) RecordCustomEvent(eventType string, params map[string]interface{}) error {
	return nil
}

func (app exutApplication) RecordCustomMetric(name string, value float64) error {
	return nil
}

func (app exutApplication) WaitForConnection(timeout time.Duration) error {
	return nil
}

func (app exutApplication) Shutdown(timeout time.Duration) {
}

// -------- Transaction test type ---------------------
type exutTransaction struct {
	//http.ResponseWriter
	newrelic.Transaction
}

func (txn exutTransaction) End() error {
	return nil
}

func (txn exutTransaction) Ignore() error {
	return nil
}

func (txn exutTransaction) SetName(name string) error {
	return nil
}

func (txn exutTransaction) NoticeError(err error) error {
	return nil
}

func (txn exutTransaction) AddAttribute(key string, value interface{}) error {
	return nil
}

func (txn exutTransaction) StartSegmentNow() (seg newrelic.SegmentStartTime) {
	return seg
}

func (txn exutTransaction) CreateDistributedTracePayload() newrelic.DistributedTracePayload {
	return nil
}

func (txn exutTransaction) AcceptDistributedTracePayload(tn newrelic.TransportType, payload interface{}) error {
	return nil
}

func (txn exutTransaction) NewGoroutine() newrelic.Transaction {
	return nil
}

func (txn exutTransaction) SetWebRequest(request newrelic.WebRequest) error {
	return nil
}
func (txn exutTransaction) SetWebResponse(http.ResponseWriter) newrelic.Transaction {
	return nil
}
func (txn exutTransaction) Application() (app newrelic.Application) {
	return app
}
func (txn exutTransaction) BrowserTimingHeader() (*newrelic.BrowserTimingHeader, error) {
	return nil, nil
}

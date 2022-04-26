package monitoring

import (
	"context"
	"testing"

	instana "github.com/instana/go-sensor"
	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func TestNewSpan(t *testing.T) {
	// No parent span, create a new span
	sensor := instana.NewSensor("my-sensor")
	span, ctx := NewSpan(context.Background(), sensor, "my-span")

	nrConfig := newrelicPreModules.Config{
		AppName: "my-nr-app",
	}
	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)
	txnPreModules := nrApp.StartTransaction("my-transaction", nil, nil)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("my-nr-app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)
	txn := nrAppV3.StartTransaction("my-transaction")

	ctx = newrelicPreModules.NewContext(ctx, txnPreModules)
	ctx = newrelic.NewContext(ctx, txn)
	defer txnPreModules.End()
	defer txn.End()

	if span == nil {
		t.Error("cannot create a parent span")
	}

	SetTag(ctx, "key", "value")

	SetTagsKV(ctx, "key1", "value1", "key2", "value2")
}

func TestNewSpanChild(t *testing.T) {
	// Has parent span, create a child span
	sensor := instana.NewSensor("my-sensor")
	parentSpan := sensor.Tracer().StartSpan("my-span-parent")
	childSpan, ctx := NewSpan(instana.ContextWithSpan(context.Background(), parentSpan), sensor, "my-span-child")

	nrConfig := newrelicPreModules.Config{
		AppName: "my-nr-app",
	}
	nrApp, _ := newrelicPreModules.NewApplication(nrConfig)
	txnPreModules := nrApp.StartTransaction("my-transaction", nil, nil)

	nrAppV3, _ := newrelic.NewApplication(
		newrelic.ConfigAppName("my-nr-app"),
		newrelic.ConfigLicense("1234567890123456789012345678901234567890"),
	)
	txn := nrAppV3.StartTransaction("my-transaction")

	ctx = newrelicPreModules.NewContext(ctx, txnPreModules)
	ctx = newrelic.NewContext(ctx, txn)
	defer txnPreModules.End()
	defer txn.End()

	if childSpan == parentSpan || childSpan == nil {
		t.Error("cannot create a child span")
	}

	SetTag(ctx, "key", nil) // value is nil

	SetTagsKV(nil, "key1", "value1", "key2", "value2") // ctx is nil
	SetTagsKV(ctx, "key1", "value1", "key2", "value2", "key3NoValue") // count is not even
	SetTagsKV(ctx, "key1", "value1", 2, "value2") // non-string key
}

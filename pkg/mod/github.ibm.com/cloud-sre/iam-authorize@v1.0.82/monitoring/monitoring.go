package monitoring

import (
	"context"
	"log"

	instana "github.com/instana/go-sensor"
	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// AppEnv represents the environment of the monitoring application.
type AppEnv string

const (
	// AppEnvProd represents the production environment
	AppEnvProd AppEnv = "prod"
	// AppEnvStage represents the staging environment
	AppEnvStage AppEnv = "staging"
	// AppEnvDev represents the development environment
	AppEnvDev AppEnv = "dev"
)

// NewDefaultMonitoringApp creates a monitoring application that will report metrics under the
// iam-authorize-iam-health-dev, iam-authorize-iam-health-stage, or iam-authorize-iam-health monitoring
// application name depending on the env value. If there is a problem creating the monitoring application,
// an error is returned.
func NewDefaultMonitoringApp(env AppEnv) *instana.Sensor {
	monitoringAppName := "iam-authorize-iam-health"
	if env == AppEnvProd {
		// no update needed
	} else if env == AppEnvStage {
		monitoringAppName = monitoringAppName + "-" + "stage"
	} else if env == AppEnvDev {
		monitoringAppName = monitoringAppName + "-" + "dev"
	} else {
		monitoringAppName = monitoringAppName + "-" + "test"
	}

	return instana.NewSensor(monitoringAppName)
}

// NewSpan is a helper function to create a span as a child of a parent or
// if there is no parent then create a new span. The span is included in the returned context.
func NewSpan(ctx context.Context, sensor *instana.Sensor, name string) (opentracing.Span, context.Context) {
	var span opentracing.Span

	parent, ok := instana.SpanFromContext(ctx)
	if ok {
		// Use the parent span
		log.Printf("%s: use the parent span", name)
		span = parent.Tracer().StartSpan(name, opentracing.ChildOf(parent.Context()))
	} else {
		// Create a new span
		if sensor != nil {
			log.Printf("%s: create new span", name)
			span = sensor.Tracer().StartSpan(name).SetTag(string(ext.SpanKind), "entry")
		}
	}

	// Add the span to the context
	ctx = instana.ContextWithSpan(ctx, span)

	return span, ctx
}

// SetTag adds a key:value pair tag to an Instana/OpenTracing span and
// to a New Relic transaction.
func SetTag(ctx context.Context, key string, value interface{}) {
	if ctx == nil || value == nil {
		return
	}

	// OpenTracing (Instana)
	span, ok := instana.SpanFromContext(ctx)
	if ok {
		span.SetTag(key, value)
	}

	// New Relic
	txnPreModules, txn := getNewRelicTransaction(ctx)
	if txnPreModules != nil {
		_ = txnPreModules.AddAttribute(key, value)
	}
	txn.AddAttribute(key, value)
}

// SetTagsKV is a concise, readable way to record key:value tags about a Span.
// An example:
//		SetTagsKV(ctx,
//			"k8sEnv", k8sEnv,
//			"k8sRegion", k8sRegion,
//			"hooksHandler", handlerName,
//			"tip_severity", 123)
func SetTagsKV(ctx context.Context, keyValues ...interface{}) {
	if ctx == nil {
		return
	}

	if len(keyValues)%2 != 0 {
		log.Printf("SetTagsKV: tag count is not even")
		return
	}

	// OpenTracing (Instana)
	span, _ := instana.SpanFromContext(ctx)

	// New Relic
	txnPreModules, txn := getNewRelicTransaction(ctx)

	// Set each KV
	for i := 0; i*2 < len(keyValues); i++ {
		key, ok := keyValues[i*2].(string)
		if !ok {
			log.Printf("SetTagsKV: non-string key (pair #%d): %T", i, keyValues[i*2])
			return
		}
		value := keyValues[i*2+1]

		if span != nil {
			span.SetTag(key, value) // Instana
		}
		if txnPreModules != nil {
			_ = txnPreModules.AddAttribute(key, value) // New Relic pre modules
		}
		txn.AddAttribute(key, value) // New Relic
	}
}


// getNewRelicTransaction retrieves txnPreModules and txn from context
func getNewRelicTransaction(ctx context.Context) (newrelicPreModules.Transaction, *newrelic.Transaction) {
	txnPreModules := newrelicPreModules.FromContext(ctx)
	txn := newrelic.FromContext(ctx)
	return txnPreModules, txn
}

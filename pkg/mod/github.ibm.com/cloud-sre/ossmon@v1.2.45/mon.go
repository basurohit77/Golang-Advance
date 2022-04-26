// IBM Confidential OCO Source Materials
// (C) Copyright and Licensed by IBM Corp. 2021, 2022
//
// The source code for this program is not published or otherwise
// divested of its trade secrets, irrespective of what has
// been deposited with the U.S. Copyright Office.

// Package ossmon is a package for monitoring with both Instana and New Relic.
package ossmon

import (
	"context"
	"errors"
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
	newrelicPreModules "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	// Common tag names
	// Add here so we get consistency across all microservices
	TagDomain = "k8sDomain"
	TagEnv    = "k8sEnv"
	TagRegion = "k8sRegion"
	TagZone   = "k8sZone"
)

// OSSMon contains the monitoring references to New Relic and Instana.
// Use either NewRelicApp for pre Go modules or NewRelicAppV3 for version 3 with modules.
type OSSMon struct {
	NewRelicApp         newrelicPreModules.Application // Deprecated: Use the New Relic go-agent with Go modules instead.
	Sensor              *instana.Sensor
	NewRelicApplication *newrelic.Application // New Relic go-agent with modules
}

// Sets the end timestamp and finalizes the span state.
func FinishSpan(ctx context.Context) {
	span, ok := instana.SpanFromContext(ctx)
	if ok {
		span.Finish()
	}
}

// NewSpan is a helper function to create a span as a child of a parent or
// if there is no parent then create a new span. The span is included in
// the returned context.
func NewSpan(ctx context.Context, sensor *instana.Sensor, name string) (opentracing.Span, context.Context) {
	var span opentracing.Span

	parent, ok := instana.SpanFromContext(ctx)
	if ok {
		// Use the parent span
		span = parent.Tracer().StartSpan(name, opentracing.ChildOf(parent.Context()))
	} else {
		// Create a new span
		if sensor != nil {
			span = sensor.Tracer().StartSpan(name).SetTag(string(ext.SpanKind), "entry")
		} else {
			log.Printf("sensor=nil")
		}
	}

	// Add the span to the context
	ctx = instana.ContextWithSpan(ctx, span)

	return span, ctx
}

// StartChildSpan creates a child span; the caller is expected to have an existing span in the context.
func StartChildSpan(ctx context.Context, name string) (opentracing.Span, context.Context) {
	var span opentracing.Span

	parent, ok := instana.SpanFromContext(ctx)
	if ok {
		// Use the parent span
		span = parent.Tracer().StartSpan(name, opentracing.ChildOf(parent.Context()))
	} else {
		// No parent span
		log.Println("ERROR. Invalid usage. A parent span must already be present.")
		return nil, ctx
	}

	// Add the span to the context
	ctx = instana.ContextWithSpan(ctx, span)

	return span, ctx
}

// StartParentSpan creates a parent span.
func StartParentSpan(ctx context.Context, m OSSMon, name string) (opentracing.Span, context.Context) {
	var span opentracing.Span

	// Create a new span
	if m.Sensor != nil {
		span = m.Sensor.Tracer().StartSpan(name).SetTag(string(ext.SpanKind), "entry")
	} else {
		log.Printf("sensor=nil")
	}

	// Add the span to the context
	ctx = instana.ContextWithSpan(ctx, span)

	return span, ctx
}

// SetError raises an error with Instana and New Relic. Typically this is
// called for server 5xx errors and not client 4xx errors.
func SetError(ctx context.Context, message string) {
	if ctx == nil {
		return
	}

	// Instana
	span, ok := instana.SpanFromContext(ctx)
	if ok {
		span.SetTag("error", true).SetTag("message", message)
	}

	// New Relic
	txnPreModules, txn := getNewRelicTransaction(ctx)
	if txnPreModules != nil {
		_ = txnPreModules.NoticeError(errors.New(message))
	}
	if txn != nil {
		txn.NoticeError(errors.New(message))
	}
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
	if txn != nil {
		txn.AddAttribute(key, value)
	}
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
		log.Printf("errTagCountNotEven")
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
			log.Printf("non-string key (pair #%d): %T", i, keyValues[i*2])
			return
		}
		value := keyValues[i*2+1]

		if span != nil {
			span.SetTag(key, value) // Instana
		}
		if txnPreModules != nil {
			_ = txnPreModules.AddAttribute(key, value) // New Relic pre modules
		}
		if txn != nil {
			txn.AddAttribute(key, value) // New Relic
		}
	}

}

// Gets the New Relic transaction from the context.  New Relic pre Go modules
// and with modules store different objects in the context.  The expectation
// is that one or the other will be nil.
func getNewRelicTransaction(ctx context.Context) (newrelicPreModules.Transaction, *newrelic.Transaction) {
	txnPreModules := newrelicPreModules.FromContext(ctx)
	txn := newrelic.FromContext(ctx)
	return txnPreModules, txn
}

// WrapWithMonitoring adds the middleware wraping for New Relic and Instana HTTP requests.
// Only one New Relic version is supported at a time (pre go modules or v3 with modules).
func WrapWithMonitoring(m OSSMon, path string, handler http.HandlerFunc) http.HandlerFunc {
	var inst http.HandlerFunc

	if m.NewRelicApplication != nil {
		_, nf := newrelic.WrapHandleFunc(m.NewRelicApplication, path, handler)
		inst = instana.TracingHandlerFunc(m.Sensor, path, nf)
	} else {
		_, nf := newrelicPreModules.WrapHandleFunc(m.NewRelicApp, path, handler)
		inst = instana.TracingHandlerFunc(m.Sensor, path, nf)
	}

	// need ability to pass in the pathTemplate?
	// inst := instana.TracingHandlerFunc(sensor, "thePathTemplate", nf)

	return inst
}

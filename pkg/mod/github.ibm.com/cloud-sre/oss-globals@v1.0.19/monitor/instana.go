package monitor

// this file contains instana specific variables that are common to all handlers.
// Thye are used inside 'span.SetTag(key, value)' calls within each handler function.
// See document here: https://github.ibm.com/cloud-sre/instana/blob/master/APM_monitoring_go.md

import (
	"net/http"
	"os"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	k8sRegion = os.Getenv("KUBE_CLUSTER_REGION")
	k8sEnv    = os.Getenv("KUBE_APP_DEPLOYED_ENV")
	appName   = os.Getenv("MONITORING_APP_NAME")
)

// addToSpan takes a handler name (usually Function name) and the incoming request to
// add a span to the context. The initial span comes from the call to
// instana.TracingHandlerFunc as part of the context.
func AddToSpan(handlerName string, req *http.Request) opentracing.Span {
	var span opentracing.Span
	parent, ok := instana.SpanFromContext(req.Context())
	if ok {
		span = parent.Tracer().StartSpan(handlerName, opentracing.ChildOf(parent.Context()))
	} else {
		sensor := instana.NewSensor(appName)
		span = sensor.Tracer().StartSpan(handlerName).SetTag(string(ext.SpanKind), "entry")

	}

	span.SetTag("appName", appName)
	span.SetTag("k8sEnv", k8sEnv)
	span.SetTag("k8sRegion", k8sRegion)

	return span
}

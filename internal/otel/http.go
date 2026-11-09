package otel

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Semantic-convention attribute keys for HTTP server spans. Defined inline to
// stay independent of semconv version naming (otelhttp uses the same keys).
const (
	attrHTTPMethod     = attribute.Key("http.method")
	attrHTTPScheme     = attribute.Key("http.scheme")
	attrHTTPRoute      = attribute.Key("http.route")
	attrHTTPStatusCode = attribute.Key("http.status_code")
	attrHTTPRequestLen = attribute.Key("http.request.content_length")
	attrHTTPServerName = attribute.Key("http.server.name")
	attrHTTPFlavor     = attribute.Key("http.flavor")
	attrHTTPUserAgent  = attribute.Key("http.user_agent")
)

// HTTPMiddleware returns a gin middleware that starts a server span for each
// request, propagates the trace context into the request context, and records
// the HTTP attributes and status code when the request completes.
//
// It is an OTel-native equivalent of otelhttp.NewMiddleware, built on the core
// trace API because the contrib otelhttp module is not available here.
func HTTPMiddleware(tracer trace.Tracer, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		ctx, span := tracer.Start(
			c.Request.Context(),
			route+" "+c.Request.Method,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attrHTTPMethod.String(c.Request.Method),
				attrHTTPScheme.String(requestScheme(c)),
				attrHTTPRoute.String(route),
				attrHTTPServerName.String(serviceName),
				attrHTTPFlavor.String(c.Request.Proto[0:2]), // e.g. "1.1"
				attrHTTPRequestLen.Int64(c.Request.ContentLength),
			),
		)
		if ua := c.Request.UserAgent(); ua != "" {
			span.SetAttributes(attrHTTPUserAgent.String(ua))
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Record status + duration after the handler runs.
		status := c.Writer.Status()
		span.SetAttributes(attrHTTPStatusCode.Int(status))
		if status >= 500 {
			span.SetStatus(codes.Error, "server error")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}
}

// requestScheme determines the request scheme, honoring X-Forwarded-Proto
// behind a proxy and falling back to TLS presence then "http". http.Request has
// no Scheme field, so it is derived here.
func requestScheme(c *gin.Context) string {
	if p := c.Request.Header.Get("X-Forwarded-Proto"); p != "" {
		return p
	}
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

// HTTPMetricsMiddleware records per-route request counts and durations as
// Prometheus metrics via the provided meter. It is an OTel-native equivalent
// of gin-metrics.
func HTTPMetricsMiddleware(meter metric.Meter) gin.HandlerFunc {
	if meter == nil {
		// No meter configured: still run the chain, just skip metrics.
		return func(c *gin.Context) { c.Next() }
	}

	reqCount, err := meter.Int64Counter(
		"http.server.request.count",
		metric.WithUnit("requests"),
		metric.WithDescription("Number of HTTP requests processed."),
	)
	if err != nil {
		reqCount = nil
	}

	reqDuration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of HTTP requests processed."),
	)
	if err != nil {
		reqDuration = nil
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		route := c.FullPath()
		status := c.Writer.Status()
		attrs := []attribute.KeyValue{
			attrHTTPRoute.String(route),
			attrHTTPStatusCode.Int(status),
		}

		if reqCount != nil {
			reqCount.Add(context.Background(), 1, metric.WithAttributes(attrs...))
		}
		if reqDuration != nil {
			reqDuration.Record(context.Background(), dur.Seconds(), metric.WithAttributes(attrs...))
		}
	}
}

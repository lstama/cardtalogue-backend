package otel

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TracerProviderConfig bundles the pieces needed to build a tracer provider.
type TracerProviderConfig struct {
	Config   Config
	Resource *sdkresource.Resource
	Logger   *slog.Logger
	Enabled  bool
}

// NewTracerProvider builds a trace.TracerProvider from configuration.
//
// Exporters:
//   - "otlp":    OTLP exporter (HTTP/protobuf or gRPC) to the collector/Grafana.
//   - "stdout":  human-readable console exporter (useful for local debugging).
//   - "none":    no-op provider (instrumentation runs but nothing is exported).
//
// When no exporter is configured a no-op provider is returned so the rest of
// the system still works.
func NewTracerProvider(ctx context.Context, cfg TracerProviderConfig) (*sdktrace.TracerProvider, error) {
	if !cfg.Enabled {
		cfg.Logger.Info("otel: tracing disabled (traces_exporter=none)")
		// A provider with no span processors exports nothing (no-op) but still
		// exposes Shutdown, so it is safe to return here.
		return sdktrace.NewTracerProvider(), nil
	}

	var exp sdktrace.SpanExporter
	switch cfg.Config.TracesExporter {
	case "otlp":
		if cfg.Config.protocolIsGRPC() {
			// gRPC exporter requires go.opentelemetry.io/otel/exporters/otlp/
			// otlptrace/otlptracegrpc, which is not available here; fall back to
			// HTTP/protobuf and log it.
			cfg.Logger.Warn("otel: gRPC OTLP exporter unavailable, using HTTP/protobuf")
		}
		var err error
		if exp, err = otlptracehttp.New(ctx); err != nil {
			return nil, err
		}
	case "stdout":
		var err error
		if exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint()); err != nil {
			return nil, err
		}
	case "none", "":
		cfg.Logger.Info("otel: tracing disabled (traces_exporter=none)")
		return sdktrace.NewTracerProvider(), nil
	default:
		return nil, errUnsupportedExporter("traces", cfg.Config.TracesExporter)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp,
			sdktrace.WithBatchTimeout(cfg.Config.ScheduleDelay),
			sdktrace.WithMaxQueueSize(cfg.Config.MaxQueueSize),
			sdktrace.WithMaxExportBatchSize(cfg.Config.MaxExportBatchSize),
		),
		sdktrace.WithResource(cfg.Resource),
	)
	cfg.Logger.Info("otel: tracer provider configured",
		"service", cfg.Config.ServiceName,
		"exporter", cfg.Config.TracesExporter,
		"endpoint", cfg.Config.Endpoint,
		"protocol", cfg.Config.Protocol,
	)
	return tp, nil
}

// TextMapPropagator returns a text map propagator using W3C trace context
// (traceparent/tracestate) plus baggage, suitable for the HTTP middleware.
func TextMapPropagator(_ trace.TracerProvider) propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func errUnsupportedExporter(kind, value string) error {
	return fmt.Errorf("unsupported %s exporter %q (want otlp|stdout|none)", kind, value)
}

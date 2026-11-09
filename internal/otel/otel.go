package otel

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lstama/cardtalogue-backend/internal/config"
	"go.opentelemetry.io/otel/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Provider bundles the initialized OpenTelemetry providers and handles. Callers
// keep the *Provider for the lifetime of the app and call Shutdown on exit.
type Provider struct {
	ServiceName    string
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  metric.MeterProvider
	Tracer         trace.Tracer
	Meter          metric.Meter
	Resource       *sdkresource.Resource
	metricsHandler http.Handler
}

// Init builds the tracer and meter providers from configuration. It returns a
// *Provider to retain and a cleanup func (bound to ctx) to call on shutdown.
//
// When tracing or metrics are disabled (exporter=none) the corresponding
// provider is a no-op but the *Provider is still returned so callers can wire
// instrumentation unconditionally.
func Init(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Provider, func(context.Context) error, error) {
	resource, err := NewResource(ConfigFromEnv(cfg), cfg.Environment)
	if err != nil {
		return nil, nil, err
	}

	tp, err := NewTracerProvider(ctx, TracerProviderConfig{
		Config:   ConfigFromEnv(cfg),
		Resource: resource,
		Logger:   logger,
		Enabled:  cfg.OTEL_TRACES_EXPORTER != "none",
	})
	if err != nil {
		return nil, nil, err
	}

	mpResult, err := NewMeterProvider(ctx, MeterProviderConfig{
		Config:   ConfigFromEnv(cfg),
		Resource: resource,
		Logger:   logger,
		Enabled:  cfg.OTEL_METRICS_EXPORTER != "none",
	})
	if err != nil {
		return nil, nil, err
	}

	p := &Provider{
		ServiceName:    cfg.OTEL_SERVICE_NAME,
		TracerProvider: tp,
		MeterProvider:  mpResult.MeterProvider(),
		Tracer:         tp.Tracer(cfg.OTEL_SERVICE_NAME),
		Meter:          mpResult.MeterProvider().Meter(cfg.OTEL_SERVICE_NAME),
		Resource:       resource,
	}
	p.setMetricsHandler(mpResult.Handler())

	// Wrap so trace shutdown runs even if meter shutdown errors.
	cleanup := func(ctx context.Context) error {
		err := mpResult.Shutdown(ctx)
		if terr := tp.Shutdown(ctx); terr != nil && err == nil {
			err = terr
		}
		return err
	}

	return p, cleanup, nil
}

// MetricsHandler returns the Prometheus HTTP handler for the local /metrics
// endpoint. It always returns a non-nil handler (a minimal 200 handler when no
// meter provider is configured) so the route can be registered unconditionally.
func (p *Provider) MetricsHandler() http.Handler {
	if p == nil || p.MeterProvider == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# no metrics available\n"))
		})
	}
	return p.metricsHandler
}

// setMetricsHandler records the Prometheus handler produced during Init.
func (p *Provider) setMetricsHandler(h http.Handler) {
	p.metricsHandler = h
}

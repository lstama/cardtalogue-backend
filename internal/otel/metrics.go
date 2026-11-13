package otel

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// MeterProviderConfig configures the metrics provider.
type MeterProviderConfig struct {
	Config   Config
	Resource *resource.Resource
	Logger   *slog.Logger
	Enabled  bool
}

// meterProviderResult bundles the SDK meter provider with the exporter used to
// render Prometheus output.
type meterProviderResult struct {
	provider    *sdkmetric.MeterProvider
	exporter    *prometheusExporter
	metricsImpl *metricsProvider
}

// NewMeterProvider builds a meter provider with a periodic reader feeding the
// custom Prometheus exporter. When disabled it returns a no-op provider.
func NewMeterProvider(ctx context.Context, cfg MeterProviderConfig) (*meterProviderResult, error) {
	if !cfg.Enabled {
		return &meterProviderResult{provider: sdkmetric.NewMeterProvider()}, nil
	}
	exporter := newPrometheusExporter()
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(cfg.Config.ScheduleDelay),
		sdkmetric.WithMaxExportBatchSize(cfg.Config.MaxExportBatchSize),
	)
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(cfg.Resource),
		sdkmetric.WithReader(reader),
	)
	return &meterProviderResult{
		provider:    provider,
		exporter:    exporter,
		metricsImpl: &metricsProvider{provider: provider, exporter: exporter},
	}, nil
}

// MeterProvider returns the SDK meter provider.
func (m *meterProviderResult) MeterProvider() metric.MeterProvider {
	return m.provider
}

// Handler serves the current metrics in Prometheus text exposition format.
func (m *meterProviderResult) Handler() http.Handler {
	return m.metricsImpl.Handler()
}

// Shutdown flushes and stops the meter provider.
func (m *meterProviderResult) Shutdown(ctx context.Context) error {
	return m.provider.Shutdown(ctx)
}

// metricsProvider holds the SDK meter provider and the exporter used to render
// Prometheus text output.
type metricsProvider struct {
	provider *sdkmetric.MeterProvider
	exporter *prometheusExporter
}

// Meter returns the meter used by the HTTP and DB instrumentation.
func (m *metricsProvider) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return m.provider.Meter(name, opts...)
}

// Handler serves the current metrics in Prometheus text exposition format.
func (m *metricsProvider) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(m.exporter.Render()))
	})
}

// Shutdown flushes and stops the meter provider.
func (m *metricsProvider) Shutdown(ctx context.Context) error {
	return m.provider.Shutdown(ctx)
}

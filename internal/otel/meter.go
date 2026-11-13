package otel

import "go.opentelemetry.io/otel/metric"

// meterHandle is the subset of meter.Meter used by the instrumentation
// middleware and the database driver. It keeps those components decoupled from
// the concrete metric.Meter type and lets tests supply a stub.
type meterHandle interface {
	Int64Counter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error)
	Int64Histogram(name string, opts ...metric.Int64HistogramOption) (metric.Int64Histogram, error)
}

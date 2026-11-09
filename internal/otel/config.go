// Package otel wires OpenTelemetry into the Cardtalogue backend: tracer and
// meter providers, a local Prometheus /metrics endpoint, structured logging,
// HTTP and database instrumentation, and graceful shutdown.
//
// The contrib modules (otelhttp, otelsql, gcp detector) are not fetchable in
// this environment's module proxy, so the HTTP and database instrumentation
// here are OTel-native equivalents built on the core trace API: an otelhttp
// style gin middleware and a database/sql driver wrapper (otelsql style).
package otel

import (
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/lstama/cardtalogue-backend/internal/config"
)

// Config holds the resolved OpenTelemetry configuration.
type Config struct {
	ServiceName        string
	Endpoint           string
	Protocol           string // "http/protobuf" | "grpc"
	TracesExporter     string // "otlp" | "stdout" | "none"
	MetricsExporter    string // "otlp" | "stdout" | "none"
	MetricsAddress     string
	MetricsPort        string
	ResourceAttributes string
	ScheduleDelay      time.Duration
	MaxQueueSize       int
	MaxExportBatchSize int
}

// ConfigFromEnv builds an otel.Config from the application config.
func ConfigFromEnv(c config.Config) Config {
	protocol := strings.ToLower(strings.TrimSpace(c.OTEL_EXPORTER_OTLP_PROTOCOL))
	if protocol == "" {
		protocol = "http/protobuf"
	}
	return Config{
		ServiceName:        c.OTEL_SERVICE_NAME,
		Endpoint:           c.OTEL_EXPORTER_OTLP_ENDPOINT,
		Protocol:           protocol,
		TracesExporter:     strings.ToLower(strings.TrimSpace(c.OTEL_TRACES_EXPORTER)),
		MetricsExporter:    strings.ToLower(strings.TrimSpace(c.OTEL_METRICS_EXPORTER)),
		MetricsAddress:     c.OTEL_METRICS_ADDRESS,
		MetricsPort:        c.OTEL_METRICS_PORT,
		ResourceAttributes: c.OTEL_RESOURCE_ATTRIBUTES,
		ScheduleDelay:      time.Duration(c.OTEL_BSP_SCHEDULE_DELAY) * time.Millisecond,
		MaxQueueSize:       c.OTEL_BSP_MAX_QUEUE_SIZE,
		MaxExportBatchSize: c.OTEL_BSP_MAX_EXPORT_BATCH_SIZE,
	}
}

// protocolIsGRPC reports whether the OTLP endpoint should use gRPC instead of
// HTTP/protobuf.
func (c Config) protocolIsGRPC() bool {
	return c.Protocol == "grpc"
}

// resourceAttributes parses the raw OTEL_RESOURCE_ATTRIBUTES string
// ("k1=v1;k2=v2") into attribute.KeyValue pairs. Malformed fragments are
// skipped so a bad value never breaks startup.
func (c Config) resourceAttributes() []attribute.KeyValue {
	raw := strings.TrimSpace(c.ResourceAttributes)
	if raw == "" {
		return nil
	}
	var attrs []attribute.KeyValue
	for _, part := range strings.Split(raw, ";") {
		kv := strings.SplitN(part, "=", 2)
		key := strings.TrimSpace(kv[0])
		if key == "" {
			continue
		}
		if len(kv) == 2 {
			attrs = append(attrs, attribute.String(key, strings.TrimSpace(kv[1])))
		} else {
			attrs = append(attrs, attribute.String(key, "true"))
		}
	}
	return attrs
}

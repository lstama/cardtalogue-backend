package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

// NewResource builds a resource from configuration. The environment string is
// reserved for future use (e.g. distinguishing prod/staging) and currently
// only affects the deployment.environment attribute when non-empty.
func NewResource(cfg Config, environment string) (*sdkresource.Resource, error) {
	attrs := cfg.resourceAttributes()
	if environment != "" {
		attrs = append(attrs, attribute.String("deployment.environment", environment))
	}
	return sdkresource.New(context.Background(), sdkresource.WithAttributes(attrs...))
}

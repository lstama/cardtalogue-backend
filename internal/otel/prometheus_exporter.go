package otel

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// prometheusExporter is a minimal sdk/metric.Exporter that stores the collected
// ResourceMetrics and renders them in the Prometheus text exposition format on
// demand. The bundled prometheus exporter (v0.69.0) is stripped in this build
// (it lacks Export/Flush/Shutdown/Temporality/Aggregation), so we render the
// SDK data model directly.
type prometheusExporter struct {
	mu       sync.Mutex
	resource *metricdata.ResourceMetrics
}

func newPrometheusExporter() *prometheusExporter {
	return &prometheusExporter{}
}

func (e *prometheusExporter) Temporality(_ metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (e *prometheusExporter) Aggregation(_ metric.InstrumentKind) metric.Aggregation {
	return nil
}

func (e *prometheusExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	e.mu.Lock()
	e.resource = rm
	e.mu.Unlock()
	return ctx.Err()
}

func (e *prometheusExporter) ForceFlush(ctx context.Context) error {
	return ctx.Err()
}

func (e *prometheusExporter) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	e.resource = nil
	e.mu.Unlock()
	return ctx.Err()
}

// Render returns the current metrics in Prometheus text exposition format.
func (e *prometheusExporter) Render() string {
	e.mu.Lock()
	rm := e.resource
	e.mu.Unlock()

	var b strings.Builder
	if rm == nil {
		return b.String()
	}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			renderMetric(&b, m)
		}
	}
	return b.String()
}

func renderMetric(b *strings.Builder, m metricdata.Metrics) {
	name := sanitizeName(m.Name)
	if m.Unit != "" {
		name += ":" + sanitizeName(m.Unit)
	}
	switch data := m.Data.(type) {
	case metricdata.Gauge[float64]:
		for _, dp := range data.DataPoints {
			writeSample(b, name, dp.Attributes, floatVal(dp.Value))
		}
	case metricdata.Gauge[int64]:
		for _, dp := range data.DataPoints {
			writeSample(b, name, dp.Attributes, strconv.FormatInt(dp.Value, 10))
		}
	case metricdata.Sum[float64]:
		for _, dp := range data.DataPoints {
			writeSample(b, name+"_total", dp.Attributes, floatVal(dp.Value))
		}
	case metricdata.Sum[int64]:
		for _, dp := range data.DataPoints {
			writeSample(b, name+"_total", dp.Attributes, strconv.FormatInt(dp.Value, 10))
		}
	case metricdata.Histogram[float64]:
		writeHistogram(b, name, data.DataPoints)
	case metricdata.Histogram[int64]:
		writeHistogram(b, name, data.DataPoints)
	}
}

func writeHistogram[N int64 | float64](b *strings.Builder, name string, dps []metricdata.HistogramDataPoint[N]) {
	for _, dp := range dps {
		labels := renderLabels(dp.Attributes)
		for i, bound := range dp.Bounds {
			le := strconv.FormatFloat(bound, 'g', -1, 64)
			fmt.Fprintf(b, "%s_bucket%s{le=%q}%s\n", name, labels, le, formatNumber(float64(dp.BucketCounts[i])))
		}
		// +Inf bucket (implicit upper bound).
		infIdx := len(dp.Bounds)
		if infIdx < len(dp.BucketCounts) {
			fmt.Fprintf(b, "%s_bucket%s{le=\"+Inf\"}%s\n", name, labels, formatNumber(float64(dp.BucketCounts[infIdx])))
		}
		fmt.Fprintf(b, "%s_sum%s%s\n", name, labels, formatNumber(float64(dp.Sum)))
		fmt.Fprintf(b, "%s_count%s%s\n", name, labels, formatNumber(float64(dp.Count)))
	}
}

func writeSample(b *strings.Builder, name string, labels attribute.Set, value string) {
	fmt.Fprintf(b, "%s%s%s\n", name, renderLabels(labels), value)
}

// sanitizeName converts an OTel metric name to a Prometheus-compatible name:
// dots become colons (Prometheus convention for OTel metrics).
func sanitizeName(name string) string {
	return strings.ReplaceAll(name, ".", ":")
}

func renderLabels(set attribute.Set) string {
	pairs := make([]string, 0, set.Len())
	for _, kv := range set.ToSlice() {
		pairs = append(pairs, fmt.Sprintf("%s=%q", sanitizeLabel(kv.Key), escapeValue(kv.Value.AsString())))
	}
	sort.Strings(pairs)
	if len(pairs) == 0 {
		return ""
	}
	return "{" + strings.Join(pairs, ",") + "}"
}

func sanitizeLabel(k attribute.Key) string {
	s := string(k)
	s = strings.ReplaceAll(s, ".", ":")
	return s
}

func escapeValue(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return r.Replace(v)
}

func floatVal[N int64 | float64](v N) string {
	return formatNumber(float64(v))
}

func formatNumber(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

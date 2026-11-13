# OpenTelemetry Instrumentation Plan — cardtalogue-backend (Go)

**Status:** DRAFT — for review.
**Scope:** Add proper OpenTelemetry instrumentation to the **existing Go backend**
(`cardtalogue-backend/`). This is separate from the Rust rewrite plan
(`cardtalogue-backend-rust/cardtalogue-rewrite-rust.md`, §14), which already folds OTel in.
**Why:** The Go backend is the live production system. It has fragmented observability and
**no distributed tracing**. This plan brings it to the same OTLP-based standard the Rust rewrite
targets, without waiting for the rewrite.

---

## 1. Current state (verified from the code, 2026-11-05)

The Go backend's observability today is **incomplete and non-unified**:

| Signal | Current implementation | Gap |
|---|---|---|
| **Logging** | Gin built-in `gin.Logger()` + `gin.Recovery()` + `log.Printf` (startup only) | Not structured/OTel; no trace correlation; no log export |
| **Metrics** | `github.com/penglongli/gin-metrics/ginmetrics` → `/metrics` (Prometheus) | Standalone; not OTLP; no custom/business metrics |
| **Traces** | **None** | No HTTP/DB/external-call tracing at all |
| **Health** | Single `/health` → `{"status":"healthy"}` | No readiness/liveness; not exposed as a metric |

Notes that shape this plan:
- Stack is **Gin** + **GORM** (mysql/postgres/sqlite dialectors) + env-var config (`joho/godotenv`).
- One external HTTP call: the card recognition/scan API in
  `internal/service/card_recognition_service.go` (`client.Do(req)` → `CARD_SCAN_API_URL`).
- Routing is **flat** (`/api/v1/cards`, `/api/v1/games`, …) and diverges from the API contract's
  nested paths — so this instruments the **real** Go backend, not the contract.

**Goal:** unified **traces + metrics + logs** exported over **OTLP**, vendor-neutral (via an OTel
Collector), with trace↔log↔metric correlation.

---

## 2. Instrumentation points

| Layer | Where | How |
|---|---|---|
| **HTTP server** | Gin router (`cmd/api/main.go`) | `otelgin` middleware → request spans (route, method, status, duration) + HTTP metrics |
| **External HTTP** | `card_recognition_service.go` (`client.Do`) | Wrap the `http.Client` with `otelhttp.NewClient` → outbound spans (http.route, network.peer) |
| **Database** | GORM (`internal/database`) | **Manual spans via GORM hooks** (see §3) — no official instrumentation exists (verified) |
| **Runtime** | process-wide | `go.opentelemetry.io/contrib/instrumentation/runtime` → goroutines, memory, GC |
| **Logs** | replace `log.Printf`/Gin logger sink | Structured logger emitting OTel log records with `trace_id`/`span_id` |
| **Health** | `/health` handler | Expose `health_status` gauge (and add readiness/liveness per contract) |

---

## 3. Database spans — important caveat

**No maintained GORM instrumentation module exists** in `opentelemetry-go-contrib` as of Oct 2026
(module proxy returns empty; the repo has no `databases/gorm`). So DB spans are done **manually**:
- Register GORM **hooks** (`BeforeQuery`/`AfterQuery`, or `BeforeCreate`/`AfterCreate`/…) that
  start/end an OTel span per query with semantic attributes: `db.system`, `db.operation`
  (`select`/`insert`/…), `db.sql.table`, and the SQL text (sanitized — no PII/tokenized params).
- Alternatively wrap `db.Transaction(...)` calls in a span so a whole unit of work is one child tree.
- Keep it in one place (`internal/database/otel.go`) so it's opt-in via env (off by default in dev
  if export latency is a concern).

*Verify at build time:* re-check the module proxy for `go.opentelemetry.io/contrib/databases/gorm`
in case an official module is republished; if so, prefer it over the manual hooks.

---

## 4. Architecture

```
                 ┌─────────────────────────────────────────┐
   Go backend    │  OTel SDK (in-process)                  │
  ┌───────────┐  │  • TracerProvider  → otlptracehttp  │   │
  │  Gin app  │──│  • MeterProvider   → otlpmetrichttp │   │
  │  + GORM   │  │  • LoggerProvider→ otlploghttp       │   │
  └─────┬─────┘  │  Resource: service.name/version/env   │
        │ OTLP/HTTP└─────────────────────┬───────────────┘
        │                        ┌───────▼────────┐
        │                        │ OTel Collector │  OTLP receiver
        │                        │  processors    │  (batch, retry,
        │                        │  exporters     │   attributes)
        ▼                        └───────┬────────┘
   Backend (select one):   ┌──────────────▼───────────┐
   • Grafana all-in-one    │ tempo (traces)            │
   • or Tempo+Prometheus+Loki│ prometheus (metrics)    │
                           │ loki (logs)               │
                           └───────────────────────────┘
```

- **Vendor-neutral:** the app only speaks OTLP/HTTP; the backend is chosen via Collector config,
  never app config.
- **Backward compatible:** keep `gin-metrics` `/metrics` during the transition; the Collector can
  also scrape Prometheus if needed, then deprecate `gin-metrics` once OTel metrics are verified.

---

## 5. Verified module versions (Go module proxy, 2026-11-05)

Pin these (or newer compatible). All from `proxy.golang.org/<module>/@latest`.

| Module | Version | Role |
|---|---|---|
| `go.opentelemetry.io/otel` | **v1.47.0** | Core SDK/API |
| `go.opentelemetry.io/otel/sdk` | **v1.47.0** | Tracer/Meter/Logger providers |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` | **v1.47.0** | OTLP/HTTP trace exporter |
| `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp` | **v1.47.0** | OTLP/HTTP metric exporter |
| `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp` | **v0.23.0** | OTLP/HTTP log exporter (log SDK still v0.x) |
| `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin` | **v0.72.0** | Gin request instrumentation |
| `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` | **v0.72.0** | Outbound HTTP (card recognition) |
| `go.opentelemetry.io/contrib/instrumentation/runtime` | **v0.72.0** | Go runtime metrics |

*Note:* `otlploghttp` is v0.x — the OTel log data model is newer; pin the log SDK modules together
and treat logs as the lowest-maturity signal.

---

## 6. Configuration (new env vars)

Add to `internal/config/config.go` + `.env.example`:

| Var | Default | Purpose |
|---|---|---|
| `OTEL_SERVICE_NAME` | `cardtalogue-backend` | Service name (resource attribute) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | *(empty = disabled)* | Collector OTLP endpoint, e.g. `http://otel-collector:4318` |
| `OTEL_EXPORTER_OTLP_SCHEME` | `http` | `http` or `grpc` |
| `OTEL_RESOURCE_ATTRIBUTES` | *(empty)* | extra resource attrs (e.g. `deployment.environment=production`) |
| `OTEL_GO_X_CAPTURE_SPAN_ERRORS` | `true` | mark spans errored on panic/return error |
| `OTEL_ENABLED` | `false` | master switch (so dev/local stays off unless opted in) |

`OTEL_ENABLED=false` keeps the app unchanged in dev; flip it on in staging/prod. Exporters are
created lazily only when enabled, so there is zero runtime cost when off.

---

## 7. Phased implementation

**Phase 0 — Foundation (SDK + Collector)**
- Add OTel modules to `go.mod`; create `internal/otel/` (tracer/meter/log provider init, OTLP
  exporters, resource, context helpers).
- Add `OTEL_ENABLED` master switch; providers created only when enabled.
- Add **OTel Collector** to `docker-compose.yml` (OTLP/HTTP receiver → batch/retry → backend
  exporters).
- Wire the TracerProvider into `main.go`; pass a `trace.Tracer` to handlers/services (via a
  shared field or context).

**Phase 1 — Traces (highest value, do first)**
- `otelgin` middleware for inbound request spans.
- `otelhttp.NewClient` around the card recognition `http.Client` for outbound spans.
- GORM hooks (§3) for DB spans; sanitize SQL params.
- Verify: a single card-scan request produces one trace spanning HTTP → DB → external API.

**Phase 2 — Metrics via OTLP**
- HTTP metrics (request count, duration, errors) from the OTel MeterProvider.
- Expose `health_status` gauge; add readiness/liveness gauges (contract gap).
- Decide on `gin-metrics`: keep `/metrics` during transition; add a Prometheus exporter in the
  Collector, then deprecate `gin-metrics` once verified.

**Phase 3 — Logs (OTel log model)**
- Replace ad-hoc `log.Printf`/Gin logger sink with a structured logger emitting OTel log records
  via `otlploghttp`; inject `trace_id`/`span_id` into every log line.
- Correlate logs with traces in the backend.

**Phase 4 — Runtime + polish**
- `runtime` instrumentation (goroutines, memory, GC).
- Resource detection (host/container/OS) via Collector processors.
- End-to-end verification: full card-scan trace + correlated logs/metrics.

---

## 8. docker-compose addition

Add an `otel-collector` service (official `otel/opentelemetry-collector` image) with an OTLP/HTTP
receiver on `4318`, a batch processor, and exporters to the chosen backend. App points
`OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318`. Backend selection (Grafana all-in-one vs
Tempo+Prometheus+Loki) is a Collector-config change only — no app rebuild.

---

## 9. Verification & rollout
- **Smoke:** enable `OTEL_ENABLED=true` locally against a local Collector; confirm spans/metrics/logs
  arrive.
- **E2E:** trace one `/api/v1/...` request end-to-end; confirm DB + external-call child spans.
- **Rollout:** staging first (flip `OTEL_ENABLED`), then prod. Keep `gin-metrics` `/metrics` live
  until OTel metrics are confirmed ingested.
- **Dashboards:** request/metric latency+error rate, DB span duration, external-call failures.

---

## 10. Open questions
- **Backend:** Grafana all-in-one (simplest) or Tempo + Prometheus + Loki (separated)?
- **`gin-metrics` fate:** deprecate after OTel metrics land, or keep both during a grace period?
- **Log maturity:** is the v0.x OTel log SDK acceptable now, or hold logs for Phase 4 and start with
  traces + metrics only?
- **Scope:** instrument the Go backend now, or is the priority the Rust rewrite (in which case this
  becomes a maintenance-mode hardening pass)?

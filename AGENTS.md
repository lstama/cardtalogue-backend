# AGENTS.md — Cardtalogue Backend

Instructions for agents working in this repo.

## Contract-first wiring (Phase 3)

This backend consumes the **contract** in `cardtalogue-api-contract/`. The
generated Go types + Gin/strict server interface live in
`internal/api/gen/` and are **committed** (do not hand-edit them).

- `internal/api/gen/backend.yaml` — the pinned contract, synced from the
  contract repo by `make sync-contract`.
- `internal/api/gen/openapi.go` — generated code (DO NOT EDIT header).
- `internal/api/gen/oapi-codegen.yaml` — committed generator config.
- `internal/api/gen/spec.go` — embeds `backend.yaml` so the validation
  middleware can load it at runtime.
- `CONTRACT_VERSION` — pins the contract version (matches the contract repo).

### Commands

```bash
make sync-contract   # copy the pinned contract into internal/api/gen/backend.yaml
make generate        # regenerate openapi.go from the contract
make verify-contract # regenerate and fail if the committed code is out of sync
```

Local dev override (use the sibling repo instead of a fetched tag):

```bash
CONTRACT_PATH=../cardtalogue-api-contract make generate
```

`make generate` needs `oapi-codegen` on `PATH` (the Makefile adds `GOPATH/bin`).
Pin version via the go.mod `tool` directive:
`github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen`.

### Generated code invariants

- Regenerating must produce **no diff** (`make verify-contract`). If it does,
  the contract changed — re-sync, regenerate, and review the diff.
- The generated Gin handlers are **not** wired into the router yet. Existing
  handlers in `internal/handlers/` remain the runtime implementation. Wiring the
  generated `ServerInterface`/`StrictHandler` into `cmd/api/main.go` is the next
  step (see below).

## Contract validation middleware (opt-in)

`internal/middleware/contract_validation.go` validates each request against the
contract using `kin-openapi`. It is **opt-in** and defaults to a **no-op** so it
never changes runtime behavior:

```bash
API_CONTRACT_VALIDATE=1 ./bin/api   # enable contract validation
./bin/api                           # default: validation disabled (no-op)
```

- Requests that match a contract operation are validated (params, query, body).
- Requests that do **not** match any contract route pass through unchanged, so
  new/unlisted routes are never blocked.
- Auth is **not** enforced by this middleware (`NoopAuthenticationFunc`); the
  existing JWT middleware owns authentication.
- If the spec fails to load/validate, the middleware logs and disables itself
  rather than crashing the server.

The middleware reconciles the app's `/api/v1` mount point via `apiPrefix`
(the contract paths are relative to the API root).

## Do not

- Do not hand-edit `internal/api/gen/openapi.go` — edit the contract and
  regenerate.
- Do not wire the generated handlers into the router without a breaking-change
  review (the contract is backwards-compatible by policy).

## Remaining work (not done in Phase 3)

- Replace the existing `internal/handlers/*` implementations with the generated
  `ServerInterface`/`StrictHandler` (strict-server wiring) and drop the manual
  route registration in `cmd/api/main.go`.
- Enable response validation (`ResponseValidationOptions`) once handlers return
  contract-shaped responses.
- Turn `API_CONTRACT_VALIDATE=1` on in the run config once discrepancies are
  resolved.

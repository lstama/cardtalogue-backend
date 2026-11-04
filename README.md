# Cardtalogue Backend

Go + [Gin](https://gin-gonic.com/) HTTP API for the Cardtalogue card-management platform.
This service owns card, pack, game, language, rarity, storage, user, and media data, and is the
single source of truth the frontend consumes.

## Design

- **Contract-first.** The API is authored as an OpenAPI 3.0 document in
  [`contracts/backend.yaml`](https://github.com/lstama/cardtalogue-api-contract/blob/master/contracts/backend.yaml).
  Handlers, models, and client stubs are generated from that contract with
  [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen). The contract is the spec; this
  repo is one implementation of it.
- **Storage.** SQLite by default (zero external services). Postgres is supported via
  `docker-compose.yml` when `DATABASE_URL` points at a Postgres instance.
- **Auth & RBAC.** JWT bearer tokens (`Authorization: Bearer <token>`). Roles: `admin`, `editor`,
  `viewer`. Authenticated endpoints enforce roles via `RequireRole`; some read endpoints are public
  or require only a valid token.
- **Validation.** A generated contract-validation middleware rejects requests/responses that
  violate the OpenAPI schema.

## Endpoints

Base path: `/api/v1`.

| Group | Methods | Auth | Notes |
| --- | --- | --- | --- |
| `/auth/login`, `/auth/refresh`, `/auth/logout` | POST | public | Issue / rotate / revoke tokens. |
| `/cards`, `/cards/:id`, `/cards/search` | GET | token | Read cards. |
| `/cards` (POST), `/cards/:id` (PUT/DELETE) | — | editor / admin | Create / update / delete. Delete is admin-only. |
| `/cards/protected`, `/cards/protected/:id` | GET | editor + | Sensitive read view. |
| `/games`, `/languages`, `/packs`, `/rarities` | GET/POST/PUT/DELETE | token / editor / admin | Reference-data CRUD. |
| `/storages` | GET/POST/PUT/DELETE | token / editor / admin | Storage CRUD. |
| `/users` | GET/POST/PUT/DELETE | admin | User management. |
| `/profile`, `/profile/password` | GET/PUT | token | Current-user profile. |
| `/media/upload` | POST | editor / admin | Upload card images. |
| `/media/:filename` | GET | public | Serve stored media. |
| `/media/scan` | POST | captcha (proof-of-work) | Trigger ORB card scan (see the [scanner repo](https://github.com/lstama/cardtalogue-yugioh-scanner)). |
| `/health` | GET | public | Liveness. |
| `/swagger/*any` | — | public | Interactive Swagger UI for the contract. |

## Prerequisites

- Go 1.24+
- (Optional) Docker, for the Postgres option
- (Optional) `oapi-codegen` for regenerating code from the contract

## Quick start (SQLite)

```bash
# 1. Copy the environment template and set a real JWT secret.
cp .env.example .env
#    Edit .env: change JWT_SECRET (and the default admin credentials if you like).

# 2. Apply migrations and seed reference data.
go run ./cmd/migrate

# 3. Run the server.
go run ./cmd/api
```

The server listens on `:8080` by default (override with `PORT`).

Swagger UI: <http://localhost:8080/swagger>

### Creating the first admin

The seed inserts a default admin (`ADMIN_EMAIL` / `ADMIN_PASSWORD` from `.env`, default
`admin@example.com` / `admin123`). **Change these before any non-local deployment.** Log in at
`POST /api/v1/auth/login` with email + password to obtain a JWT.

## Docker (Postgres) option

```bash
# Override credentials via the environment (they are NOT hardcoded at runtime).
export POSTGRES_USER=cardtalogue
export POSTGRES_PASSWORD=change-me
export POSTGRES_DB=cardtalogue

docker compose up -d postgres
# Point the API at Postgres:
export DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable"
go run ./cmd/migrate
go run ./cmd/api
```

`POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` in `docker-compose.yml` are read from the
environment with safe defaults — do not commit real values.

## Environment variables

See [`.env.example`](.env.example) for the full list. Key variables:

- `JWT_SECRET` — **must** be a long random string in any real deployment.
- `ADMIN_EMAIL` / `ADMIN_PASSWORD` — seed admin credentials (dev only).
- `DATABASE_URL` — leave unset for SQLite; set to a `postgres://` URL for Postgres.
- `PORT` — listen port (default `8080`).

## How to run

```bash
cp .env.example .env          # set a real JWT_SECRET (and default admin creds if you like)
go run ./cmd/migrate          # migrations + seed reference data
go run ./cmd/api              # server on :8080, Swagger at http://localhost:8080/swagger
```

SQLite by default (no external services). For Postgres, see [Docker (Postgres) option](#docker-postgres-option).

## How to contribute

This repo is **contract-first** — the API lives in the [contract repo](https://github.com/lstama/cardtalogue-api-contract),
not here. To change the API:

1. Edit `contracts/backend.yaml` in the contract repo, bump `CONTRACT_VERSION`.
2. `make generate` here (syncs the pinned contract and regenerates local types/docs).
3. `make test` plus the breaking-change check, then open a PR.

See [Contract-first workflow](#contract-first-workflow) for the full details.

## Contract-first workflow

This repo consumes the shared contract. To change the API:

1. Edit the OpenAPI document in the [contract repo's `contracts/backend.yaml`](https://github.com/lstama/cardtalogue-api-contract/blob/master/contracts/backend.yaml).
2. Bump `CONTRACT_VERSION` in that repo and run its `make generate` to regenerate shared types.
3. Rebuild this repo (`go mod tidy` if the generated types changed), then regenerate local docs:
   `swag init` (or `make docs`).
4. Run `make test` and the breaking-change check before merging.

See [`cardtalogue-api-contract/README.md`](https://github.com/lstama/cardtalogue-api-contract/blob/master/README.md) for details.

## Development

```bash
make lint      # staticcheck + go vet
make test      # go test ./...
make docs      # regenerate swagger docs from annotations
make generate  # (contract repo) regenerate shared types from the OpenAPI doc
```

## License

Licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE).

Yu-Gi-Oh! card names, descriptions, and related data are used for identification and reference
purposes only. Yu-Gi-Oh! is a trademark of Konami Group Corporation. This project does not
distribute card images; image URLs are placeholders.

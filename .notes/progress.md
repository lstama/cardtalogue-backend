# Progress

## Goal
Onboard as new backend engineer: understand how to change the DB schema/models in this Go backend and whether changes can be rolled back. Scope: DB + models only.

## Done
- Read live code: `internal/models/*.go` (8 models), `internal/database/{connection,migrations}.go`, `cmd/{api,migrate}/main.go`, `internal/config/config.go`, Makefile, docker-compose.yml, `.env`, `sample-data/initialize.sql`, README.
- Confirmed GORM AutoMigrate semantics + full `Migrator` API (AddColumn/DropColumn/RenameColumn/ChangeColumnType/DropForeignKey/AddRemoveNotNull; SQLite dialector does table-copy rebuilds) against gorm.io/docs/migration.html.
- Confirmed GORM has NO schema versioning/rollback: `SavePoint`/`RollbackTo` in dialectors are transaction savepoints; `go-gormigrate` README states "Gorm misses proper schema versioning and migration rollback support".
- Wrote full reference in `scratch/db-schema-model.md` and the user-facing doc `DB_SCHEMA.md` at repo root.
- Wrote production plan (SQLite-only requirement) in `agent/PLAN.md` at repo root.
- Wrote `agent/PLAN-POSTGRES.md`: Part 1 deltas for an EXISTING deployment (user confirmed one exists), Part 2 future PostgreSQL move.

## Answer (short form)
- Schema source of truth = Go structs in `internal/models/`. `cmd/api/main.go` calls `RunMigrations` on every start -> `AutoMigrate` over a fixed 8-model list. `migrations/` is empty and `MIGRATIONS_PATH` is ignored.
- Rollback: NOT possible automatically. `RollbackMigrations` hard-errors; Makefile migrate targets are echo stubs; no version tracking. Only manual SQL or destructive DB reset.

## Next steps
- Plan is written but NOT implemented. Implementation order is in `agent/PLAN.md` §5 (12 steps: .gitkeep -> .env -> volume -> sqlite3 driver name -> baseline SQL -> single DB lifecycle -> FK pragma+test -> gate AutoMigrate -> real migrate CLI -> recipes -> snapshots -> tests/CI).
- User answered Q1: an existing deployment exists -> `agent/PLAN-POSTGRES.md` Part 1 supersedes the `PLAN.md` §5 order.
- Still open (ask before implementing): where does the live SQLite file live (image layer / volume / bind mount) and is the host FS local? has the container been restarted since data was written (data may already be lost)? is AUTO_MIGRATE acceptable in dev? do you want FK enforcement load-bearing now?
- Nothing implemented yet; no `go` commands run.

## Open problems (blocking for a SQLite deploy)
- `Dockerfile:39-41` already runs SQLite (`DATABASE_URL=file:/app/data/cardtalogue.db`) but there is NO `VOLUME` -> DB lost on every new container; compose only volumes `postgres_data`.
- `Dockerfile:31` `COPY migrations /app/migrations` breaks on a clean clone (`migrations/` empty + untracked by git).
- `.env` has a postgres `DATABASE_URL` and no `DATABASE_DRIVER`; `config.go:47` defaults driver to `sqlite` -> driver/URL mismatch.
- `cmd/migrate/main.go` `sql.Open(databaseDriver="sqlite")` -> unknown driver; golang-migrate registers `sqlite3` (mattn/go-sqlite3).
- `RunMigrations` (`migrations.go:17`) opens a second `gorm.DB`, never closed -> two SQLite handles on one file.
- FK constraints are emitted (`connection.go:25` `gorm.Config{}` does not disable them) but SQLite ignores them unless `PRAGMA foreign_keys = ON`.
- `CreateDefaultSeedData` runs inside the migration step (`migrations.go:58`) -> dev fixtures in prod DB.
- No tests at all in the repo; README `:10`/`:215` promise up/down migrations that do not exist; `make fill-sample` psql flags wrong.

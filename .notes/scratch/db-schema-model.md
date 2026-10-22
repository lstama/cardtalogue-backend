# DB / model reference (cardtalogue-backend)

Verified by reading live files only (no git history, no `go` commands run).

## Single source of truth = Go structs
- `internal/models/{user,game,language,pack,rarity,storage,media,card}.go` — 8 GORM models.
- GORM v2: `gorm.io/gorm v1.31.x` (go.mod), dialectors `gorm.io/driver/{postgres,mysql,sqlite}`.
- Connection: `internal/database/connection.go:NewConnection(databaseURL, databaseDriver)`; `gorm.Open(dialector, &gorm.Config{})` + `sqlDB.Ping()`.
- `Close(db)` exists in same file; `cmd/api/main.go` uses `defer database.Close(db)`.

## Tables (GORM default inflection, no `Table()` override anywhere)
users, games, languages, packs, rarities, storages, media, cards.
Verify live with `psql -c '\d cards'` rather than trusting inflection.

## FK wiring (explicit column + foreignKey options)
- `Pack.GameID` / `Pack.Game` -> `packs.game_id`; `Pack.LanguageID`/`Language` -> `packs.language_id`
- `Card` FKs (card.go:26-37): `game_id` (not null), `language_id` (not null), `pack_id` (not null), `rarity_id` (not null, default:1), `storage_id` (NULLABLE), `media_id` (NULLABLE, `*uint`)
- `Card.MediaID *uint` + `Card.Media *Media gorm:"column:media_id;foreignKey:MediaID;..."` (card.go:36/37) -> FK column is `media_id` (NOT `provider_id`)
- `Media.Provider` (media.go:14) is an unrelated plain `string` column on the `media` table, `not null` with no default
- All FK fields `not null`; no `nil` allowed.
- `gorm.Config{}` does NOT set `DisableForeignKeyConstraintWhenMigrating: true` -> AutoMigrate emits real `fk_*` CHECK constraints (GORM v2 default).

## Soft delete
Only `User` (user.go:18) and `Card` (card.go:24) have `DeletedAt gorm.DeletedAt`. Game/Language/Pack/Rarity/Storage/Media hard-delete.

## How schema is applied (live path)
1. `cmd/api/main.go` -> `database.RunMigrations(cfg.DatabaseURL, cfg.DatabaseDriver, cfg.MigrationsPath)` on EVERY start.
2. `internal/database/migrations.go:RunMigrations` -> `_ = migrationsPath` (ignored), opens a SECOND connection via `NewConnection`, then `AutoMigrate(db)`.
3. `AutoMigrate` list order is dependency-first: User, Game, Language, Pack, Rarity, Storage, Media, Card. New models with FKs must be appended after their dependencies.
4. `CreateDefaultSeedData` runs in the same transaction (ensure* Default game/language/pack/rarity/storage) — idempotent.
5. `CreateDefaultAdmin(db, email, password)` called separately in main (logs warning on failure).
6. `migrations/` directory is EMPTY. No SQL migration files are read by the app.

## GORM AutoMigrate capabilities (gorm.io/docs/migration.html, fetched 2026-07-09)
- Creates missing tables/columns/indexes/foreign-key constraints.
- Can change column type, size, and add NOT NULL.
- Will NOT remove unused columns / foreign tables / indexes / FK constraints.
- Will NOT remove NOT NULL constraint.
- Docs: "only designed for development and simple schema updates"; use 3rd-party tools for complex changes.
- Postgres: `ALTER TABLE ... ADD/DROP COLUMN` is not atomic -> AutoMigrate can leave DB in intermediate state on failure.
- Not supported by AutoMigrate: CHECK constraints, foreign-key constraints as a *removal*, `ON UPDATE` auto-increment, partial unique constraints.

## Rollback: NOT available
- `internal/database/migrations.go:RollbackMigrations` -> always `return fmt.Errorf("rollback is not supported when using GORM auto-migration")`.
- `GetMigrationVersion` -> always `(0, false, nil)`; `cmd/migrate` prints `Current version: 0, Dirty: false` (meaningless).
- Makefile `migrate-up|down|rollback|create|version|force` are `@echo` stubs; they never invoke code.
- `cmd/migrate/main.go`: `up` = GORM shim; `down` = always errors; `create` writes `migrations/<ts>_<name>.{up,down}.sql` that nothing reads; `force` = the only real golang-migrate path, but `migrations/` is empty.
- `force` sqlite path is broken: `sql.Open("sqlite", url)` (no `sqlite` database/sql driver registered; golang-migrate expects `sqlite3`) and `NewWithDatabaseInstance(migrationsPath, databaseDriver, ...)` passes `"sqlite"` as golang-migrate's database name.
- No version tracking => no history, no undo. Drift is one-way.

## Rollback options (manual)
- psql into the container: `docker exec -it cardtalogue_psql_container bash` then hand-written inverse SQL (`ALTER TABLE ... DROP COLUMN`, `DropConstraint` for `fk_*`).
- Destructive reset: `make delete-db` (podman) / `make delete-db-docker` (docker) -> `docker volume rm -f cardtalogue-backend_postgres_data`; then `make db-up` recreates an empty DB and the app re-seeds defaults.
- Named volume is `postgres_data` in `docker-compose.yml` (runtime name `<dir>_postgres_data`). `make clean` does NOT touch `data.db` (only `bin/`, `coverage.out`) -> sqlite reset must be done by hand.
- Re-introduce golang-migrate: add SQL to `migrations/`, rewire `cmd/migrate` `up`/`down` to golang-migrate, and stop auto-migrate in `cmd/api/main.go` (two sources of truth otherwise conflict).

## Broken / stale plumbing (flag to user)
- `.env`: `DATABASE_URL=postgres://user:password@172.17.0.1:5432/dbname?sslmode=disable` but NO `DATABASE_DRIVER` -> `config.Load()` default is `sqlite` -> sqlite dialect gets a `postgres://` string. Must set `DATABASE_DRIVER=postgres`.
- Makefile defaults: `DATABASE_DRIVER ?= sqlite`, `DATABASE_URL ?= file:./data.db`; `run` exports those three vars, so use `make run DATABASE_DRIVER=postgres DATABASE_URL=...`.
- `make fill-sample`: `psql -h 172.17.0.1 -U user -d dbname -a -f sample-data/initialize.sql` — `-d dbname` is a literal db name and `-a` is not a psql flag -> target is broken.
- README "Database Schema" section documents Users + Items only; stale vs live models (cards/games/languages/packs/rarities/storages/media).
- README "Run Migrations" says `make migrate-up` runs migrations; it does not.
- `sample-data/initialize.sql` must match GORM-generated columns exactly (uses `media_id`, `game_id`, `language_id`, `pack_id`, `rarity_id`, `storage_id`) — consistent with the `column:media_id` override on `Card.Media`.
- `RunMigrations` opens a second `gorm.DB` and never closes it (only the main one is closed).

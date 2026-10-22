# Decisions (append-only)

## 2026-07-09 — DB/model onboarding research
- Read live files only; ignored git history and did not run any `go` command (user constraint). Facts verified by reading source + Makefile.
- Confirmed GORM AutoMigrate capability/limitation list from gorm.io/docs/migration.html (web_browse / Lightpanda, per repo instruction to use Lightpanda for browsing).
- Decision: treat `internal/models/*.go` as the only schema source of truth; treat `migrations/` as dead scaffolding.
  - Rejected: writing SQL migrations now — would create two competing schema sources of truth while `cmd/api/main.go` still auto-migrates on every boot.
- Decision: rollback is out of scope for the current design; recommend additive-only model edits + manual SQL for removals/renames.
  - Rejected: switching to golang-migrate immediately — bigger change, needs `cmd/api` auto-migrate removed and `cmd/migrate` rewired; do it only if/when production schema control is required.
- Flagged (not fixed): `.env` missing `DATABASE_DRIVER`, broken `make fill-sample`, unclosed second DB handle in `RunMigrations`, stale README.

## 2026-07-09 — Production plan: SQLite-only (user requires SQLite to save cost)
- Wrote `agent/PLAN.md` (plan only, nothing implemented).
- Decision: production schema source of truth = numbered SQL files `migrations/NNNN_slug.{up,down}.sql` run by golang-migrate (already vendored in go.mod); `AutoMigrate` demoted to dev-only behind `AUTO_MIGRATE` (default false in prod).
  - Rejected: keep AutoMigrate + snapshots only (no schema versioning, no schema rollback, prod schema churn on every boot).
  - Rejected: go-gormigrate (extra dep layer; SQLite ALTER limits still apply; inverses still hand-written) — revisit if team prefers Go steps over SQL.
  - Rejected: Atlas / atlas-provider-gorm (new toolchain, SQLite caveats) — best drift detection, deferred for cost.
  - Rejected: Postgres (out of scope by user constraint).
- Decision: rollback = two tiers: (1) hand-written inverse SQL per migration, (2) whole-file SQLite snapshots (`data.db` is one file) + one `.dump` text dump per release. Never rely on AutoMigrate (structs describe the NEW schema; the old one exists nowhere in code).
- Decision: SQLite change discipline = additive-only; renames/type changes = add -> copy -> drop; column removal only if not PK/indexed/UNIQUE/FK/CHECK/trigger, else table rebuild (GORM Migrator already implements the rebuild for SQLite).
- Fact check: GORM `SavePoint`/`RollbackTo` (gorm_sqlite >= v1.0.4) are TRANSACTION savepoints, not schema rollback. Earlier claim that SQLite could roll back via GORM was wrong; corrected here.
- Fact check: SQLite FK enforcement is per-connection and OFF by default; `gorm.io/driver/sqlite` (C++ dialector) and `mattn/go-sqlite3` (golang-migrate's `sqlite3` driver) are different SQLite builds -> do not assume shared URI option syntax; guard with an assertion test instead.

## 2026-07-09 — Existing deployment assumed + future Postgres move -> `agent/PLAN-POSTGRES.md`
- User fixed the assumption: an existing deployed SQLite DB exists (answers PLAN.md §7 Q1). No DB file in the repo (`find -name '*.db'` -> none, `bin/` empty) -> live file is only in the deployed container/host.
- Decision: Part 1 reorders PLAN.md -> locate/rescue the live DB FIRST (Dockerfile has no VOLUME, /app/data is an image layer -> data may already be lost on redeploy), snapshot + `.dump` audit, schema drift audit (live `.schema` vs structs), FK-violation scan, THEN baseline.
- Decision: baseline `0001_baseline.up.sql` is derived from the LIVE `.schema` (cleaned), not generated from structs; adopt the version with golang-migrate `Force(version)` (v4 has no SetCurrentVersion/Skip; `force` subcommand already exists in cmd/migrate).
- Decision: FK enforcement is a behaviour change on live data -> scan for dangling FKs BEFORE enabling the pragma. Verified on scratch DB: FK insert with a dangling target succeeds (rc=0) with FKs off, and `NOT EXISTS (SELECT 1 FROM users AS u WHERE u.id = c.game_id)` reports it.
- Decision (Part 2, future Postgres): keep ONE canonical numbered migration tree in a portable SQL subset; engine tuning (PRAGMA/WAL) stays in Go/config, never in SQL; keep the postgres dialector branch (connection.go:15-16) and delete only mysql (reverses PLAN.md C4); declare explicit `size:`/`type:` on strings now; make the PK explicitly auto-incrementing (`autoIncrement`) because SQLite gives `INTEGER PRIMARY KEY` auto-assignment implicitly and Postgres does not (app relies on it: `internal/repository/card_repository.go:35` `db.Create(card)`; `sample-data/initialize.sql` inserts without an `id` column).
- Verified SQLite-only artefacts that must not enter a shared tree: `.dump` emits `PRAGMA foreign_keys=OFF;` and standalone `CREATE FOREIGN KEY fk_1 (cards.game_id) REFERENCES (users.id);` (invalid in Postgres); `.dump` INSERT lines use single-quoted identifiers (invalid in Postgres) -> bulk data moves via CSV or a Go port, never via `.dump`.
- Fact: `postgresql.conf` at repo root already tunes Postgres to shared_buffers=64MB / max_connections=20 -> Postgres cost here is the service+volume+ops, not RAM tuning.

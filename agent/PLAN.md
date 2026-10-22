# Production plan — SQLite-only database strategy for cardtalogue-backend

Status: **plan only, nothing implemented yet.**
Scope: DB + models only. Read together with `../DB_SCHEMA.md` (schema reference).
Constraint honoured: SQLite is the required target (cost). Postgres/MySQL paths get deleted or gated, not maintained.

---

## 0. Decision (TL;DR)

| Question | Answer |
|---|---|
| Keep GORM `AutoMigrate` as the schema source of truth? | **No.** Demote it to a dev-only convenience. |
| Schema source of truth in production? | **Numbered, versioned SQL files** in `migrations/` (`NNNN_slug.up.sql` + `.down.sql`), run by `golang-migrate` (already a dependency in `go.mod`). |
| Rollback possible today? | **No** — `internal/database/migrations.go:25` returns a hard error, `:34` reports version `0, false`. Makefile `migrate-down/rollback/version/force` are `@echo` stubs (`Makefile:59-72`). |
| Rollback after this plan? | **Yes, two tiers:** (1) hand-written inverse SQL per migration, (2) whole-file SQLite snapshots (`data.db` is one file → copy/restore). |
| Migration discipline | **Additive-only** for columns; renames/type changes are expressed as *add → copy → drop*, never as `ALTER` on SQLite. |
| Enforcement | CI guard: production boot must **refuse** to auto-migrate; plus a `schema-check` diff and an FK-enforcement assertion test. |

Why not "keep AutoMigrate and just add snapshots": AutoMigrate derives the schema from Go structs, and the structs describe the **new** schema. After a change, the old schema no longer exists anywhere in the code, so AutoMigrate can never undo it. GORM has no schema versioning or schema rollback (`SavePoint`/`RollbackTo` in the dialectors are *transaction* savepoints, not schema rollback; `go-gormigrate` exists precisely because "Gorm misses proper schema versioning and migration rollback support"). So rollback requires an artifact that is not the structs: numbered SQL.

---

## 1. Current state (verified facts, with evidence)

| # | Fact | Evidence |
|---|---|---|
| 1 | Schema is generated from Go structs; `migrationsPath` is discarded | `internal/database/migrations.go:15` `_ = migrationsPath` |
| 2 | `migrations/` is **empty** and **untracked by git** | `git ls-files migrations` → empty |
| 3 | `Dockerfile:31` does `COPY migrations /app/migrations` → **build breaks on a clean clone** (empty dir is not in git) | `Dockerfile:31` |
| 4 | Production already runs SQLite: `Dockerfile:39-41` `DATABASE_DRIVER=sqlite`, `DATABASE_URL=file:/app/data/cardtalogue.db` | `Dockerfile` |
| 5 | **No `VOLUME` in `Dockerfile`** and `/app/data` is an image layer dir → the SQLite DB is **lost on every new container**; compose only volumes `postgres_data` | `grep VOLUME Dockerfile` → none; `docker-compose.yml:12` |
| 6 | `docker-compose.yml` still starts Postgres 15 (`docker-compose.yml:4`) while prod is SQLite → two DBs, drift, and a wasted service | `docker-compose.yml` |
| 7 | `cmd/api/main.go:56` runs migrations **on every boot** | `cmd/api/main.go:56` |
| 8 | `RunMigrations` opens a **second** `gorm.DB` (`migrations.go:17`) and never closes it → two SQLite instances on one file; the app then runs on the first one | `migrations.go:17`, `cmd/api/main.go:49` |
| 9 | Seed data is injected inside the migration step (`migrations.go:58` `CreateDefaultSeedData`) → dev fixtures land in the prod DB | `migrations.go:58` |
| 10 | `.env` sets a **postgres** `DATABASE_URL` but no `DATABASE_DRIVER`; `config.go:47` defaults driver to `sqlite` → driver/URL mismatch | `.env:1`, `config.go:47` |
| 11 | `cmd/migrate/main.go` calls `sql.Open(databaseDriver, …)` with `"sqlite"`; golang-migrate's sqlite driver registers as **`"sqlite3"`** → `database/sql` unknown-driver error, so the sqlite path in the migrate binary is broken | `cmd/migrate/main.go` |
| 12 | `gorm.Config{}` (`connection.go:25`) does **not** set `DisableForeignKeyConstraintWhenMigrating`, so FK constraints are emitted — but SQLite ignores FK clauses unless `PRAGMA foreign_keys = ON` | `connection.go:25` |
| 13 | No tests exist in the repo (`find -name '*_test.go'` → none) → nothing asserts schema or FK behaviour | — |
| 14 | README advertises "Database migrations with up/down support" (`README.md:10`) — false | `README.md:10` |
| 15 | 8 GORM models → 8 tables (`users, games, languages, packs, rarities, storages, media, cards`); `internal/models/scan.go` holds JSON API types only, not a GORM model | `migrations.go:43-52`, `scan.go` |

Items 5, 8, 10, 11, 12 are **blocking bugs for a SQLite deployment** and must be fixed before any migration tooling is trusted.

---

## 2. Is SQLite actually adequate here? (cost/scale check)

Adequate, with named limits. This app is a card catalogue: read-heavy, single-node, media stored on disk (`uploads/`, `uploads/media`), no multi-region orchestration, no per-tenant isolation. SQLite comfortably handles this to ~10⁷ rows and hundreds of concurrent readers.

Accept the following limits explicitly (they drive the rules in §4):

1. **One writer at a time.** WAL (`_journal_mode=wal`) buys concurrent readers, not concurrent writers. Keep writes single-flighted; the API is already effectively sequential per request.
2. **No NFS/network filesystems** (SQLite byte-range locking). The DB file must be on a real local FS (`tmpfs` also disallowed for the DB).
3. **FK enforcement is opt-in.** Default off → today's `fk_*` constraints are decorative.
4. **`ALTER TABLE` is narrow.** `ADD COLUMN` cannot be PK/UNIQUE/`NOT NULL` w/o a non-NULL default; `DROP COLUMN` fails if the column is indexed, PK, UNIQUE, in an FK, in a `CHECK`, or in a trigger. Renaming a column does not exist.
5. **No per-column type change** without a table rebuild.
6. **Whole-file state** = good for backups, bad for partial-failure isolation (a bad write can corrupt the one file).

If a future requirement breaks any of 1–6, that is the trigger to revisit Postgres — record it in `decisions.md`, do not pre-pay for it now.

---

## 3. Target architecture

```
one DB file, one owner, one writer
  /app/data/cardtalogue.db   (named volume, persists across containers)

schema = migrations/NNNN_slug.{up,down}.sql      ← single source of truth, in git
  applied by cmd/migrate (golang-migrate, sqlite3 driver)
  version stored in the DB by golang-migrate

Go structs (internal/models/*.go) = runtime shape only
  must match the SQL; drift is detected, never auto-repaired in prod

AutoMigrate = dev-only, on a throwaway DB, behind an explicit flag
```

Three binaries/roles:
- `cmd/api` — serves. **Never mutates schema** in production.
- `cmd/migrate` — the only schema mutator. `up`, `down N`, `rollback`, `version`, `snapshot`, `restore`.
- dev flow (`make dev`) — may AutoMigrate, but only into `./data.db` under `dev/`.

---

## 4. The plan, in order

### Phase A — make the SQLite runtime trustworthy (do first, blocks everything else)

A1. **Persist the DB.** Add a named volume in `docker-compose.yml` for the app and mount it at `/app/data`; keep `DATABASE_URL=file:/app/data/cardtalogue.db`. Delete the Postgres service from compose (prod is SQLite); if a Postgres dev box is still wanted, put it in a separate `docker-compose.dev.yml`.
A2. **Fix `.env`** (and `.env.example`): add `DATABASE_DRIVER=sqlite`, make `DATABASE_URL` a `file:` path, keep `MIGRATIONS_PATH=file://./migrations`.
A3. **One connection, one lifecycle.** Refactor `cmd/api/main.go:49` + `migrations.go:17` so migrations run against the *same* `gorm.DB` (or run to completion and close before the app opens). On SQLite, two live handles on one file is a lock hazard.
A4. **Turn FK enforcement on, and prove it.** After `gorm.Open`, `db.Exec(nil, "PRAGMA foreign_keys = ON;")` (or the driver's URI option) — then add a test that inserts a `cards` row with a non-existent `game_id` and asserts failure. Note the asymmetry to verify: `gorm.io/driver/sqlite` is a C++ dialector, whereas golang-migrate's sqlite driver is `mattn/go-sqlite3` (registers `"sqlite3"`, supports `?_foreign_keys=on&_journal_mode=wal&_busy_timeout=5000&_synchronous=normal` in the `file:` URI). Do not assume the option syntax is shared; the assertion test is the guard.
A5. **Fix `cmd/migrate/main.go` driver name**: `sql.Open("sqlite3", path)` for the sqlite case (keep `"postgres"`/`"mysql"` branches only if those engines survive; otherwise delete them).
A6. **Move seed data out of the migration step.** `migrations.go:58 CreateDefaultSeedData` → explicit `migrations/NNNN_seed_data.up.sql` (or a dedicated `make seed` step). Production DB must not receive dev fixtures implicitly.
A7. **Make `migrations/` exist in git** (`.gitkeep`) so `Dockerfile:31` stops breaking clean clones.

Definition of done for A: `make migrate-version` prints a real version; FK test passes; DB survives container restart.

### Phase B — freeze the current schema as an explicit baseline

B1. Dump the live schema once (`sqlite3 ./data.db .schema`, or generate from the structs while AutoMigrate is still authoritative) and write it as **`migrations/0001_baseline.up.sql`** — 8 `CREATE TABLE`s + FK/index/`CHECK` clauses, in dependency order (`users, games, languages, packs, rarities, storages, media, cards`).
B2. Write the true inverse **`0001_baseline.down.sql`** (drop the 8 tables in reverse order).
B3. Add `0002_seed_data.up.sql` for the fixed reference tables (`games`, `languages`, `rarities`, `storages`) that today come from `CreateDefaultSeedData`; inverse = delete those rows.
B4. Cross-check the baseline against `DB_SCHEMA.md` (FK columns `game_id/language_id/pack_id/rarity_id` NOT NULL, `storage_id`/`media_id` nullable `*uint`, `rarity_id` default 1, `DeletedAt` soft-delete only on `users` and `cards`).

Rule from here on: **the SQL files are the schema.** Structs must be edited to match SQL, not the other way round.

### Phase C — stop AutoMigrate in production

C1. Add a config switch, default **off**: `AUTO_MIGRATE=false`. `cmd/api/main.go:56` calls `RunMigrations` only when it is true; when false and the schema does not match the structs, **fail fast** rather than repair.
C2. `make run` / `make dev` set `AUTO_MIGRATE=true` against `./data.db` only; `Dockerfile:39-41` sets it false.
C3. Keep `AutoMigrate` as a *development convenience*, documented as "throwaway DB only".
C4. Delete the `postgres`/`mysql` dialector branches (`connection.go:15-18`) or gate them behind a build flag, so nobody reintroduces a second schema dialect by accident. (SQLite-only is a stated requirement.)

### Phase D — versioned migrations become the workflow

D1. `cmd/migrate` gains real subcommands: `up [N]`, `down N`, `rollback`, `version`, `status`, `snapshot <tag>`, `restore <tag>`.
D2. Makefile stubs (`Makefile:55-72`) become real targets wired to `cmd/migrate` (they are `@echo` today).
D3. Every schema change ships as a **pair** in one PR: `NNNN_what_changed.up.sql` + `NNNN_what_changed.down.sql`, both reviewed.
D4. SQLite-safe change recipes (this is the part that prevents silent breakage):

| Change | Do this | Never do this |
|---|---|---|
| Add a field | `ALTER TABLE t ADD COLUMN c type default NULL;` | `ADD COLUMN … NOT NULL` w/o constant default; PK/UNIQUE on a new column |
| Rename a field | add new column → `UPDATE t AS c_new = c_old` (via a `;`-separated SQL step or a Go step) → drop old column | expect `ALTER … RENAME` to work |
| Change a type | same add/copy/drop recipe | `ALTER COLUMN SET TYPE` |
| Remove a column | allowed only if it is **not** PK, indexed, UNIQUE, or referenced by an FK/`CHECK`/trigger; otherwise rebuild the table (`CREATE new → COPY → DROP old → RENAME`) | assume `DROP COLUMN` succeeds |
| Add an FK | `ALTER … ADD COLUMN … fk` (default NULL) then a separate step to fill values | `ADD COLUMN … REFERENCES` with a non-NULL default |
| Change PK / FK of an existing column | table-rebuild only, with a written down-migration | hand `ALTER` |
| Add an index | `CREATE UNIQUE INDEX …` — but see "Remove a column" (an indexed column cannot be dropped later) | index a column you plan to delete |

D5. Optional, for the rebuild cases: express them through GORM's `Migrator` API (`tx.Migrator().DropColumn / RenameColumn / ChangeColumnType / DropForeignKey / AddRemoveNotNull …`) inside a Go step of `cmd/migrate`. The SQLite dialector already implements the table-copy rebuild ("SQLite doesn't support ALTER/DROP COLUMN — GORM creates a new table, copies data, drops the old one, renames"), so you get correct rebuilds without hand-writing SQLite's fiddly sequence. Trade-off: the step is Go, not SQL, so it is not auto-inverted — its inverse must also be written.

### Phase E — rollback design (the actual answer to "can we go back?")

**Tier 1 — exact inverse SQL (primary).** `make migrate-down 1` / `migrate-rollback` runs `NNNN_*.down.sql`. Requires D3 discipline. Catches 95% of cases and is the only mechanism that restores *schema* semantics.

**Tier 2 — whole-file snapshots (belt and braces).** Because SQLite is one file:
```
make db-snapshot TAG=$(shell date +%Y%m%dT%H%M%S)   # cp data.db -> snapshots/data.<TAG>.db  (only when no writer active)
make db-restore  TAG=...                            # stop app, copy back, run schema-check
```
Cheap, total, and covers cases Tier 1 cannot (bad data written by a correct migration). Keep a bounded number of snapshots (e.g. 10) and one `.dump` text dump per release (`sqlite3 data.db .dump > dump.sql`) as an FS-independent rescue.

**Tier 3 — schema drift detection (catches the case where rollback is "too late").** `make schema-check` diffs the live DB schema against the schema implied by `migrations/` up to `version`; a mismatch fails CI. This is what makes "we forgot to roll back" visible instead of silent.

**Never** expect `AutoMigrate` to roll anything back; `migrations.go:25/34` stay as the loud errors they are (or get replaced by the real golang-migrate calls).

### Phase F — guards and CI

F1. New test `internal/database/schema_test.go`: FK enforcement assertion (D4/A4), soft-delete columns present on `users`/`cards`, FK columns NOT NULL as documented.
F2. New test: `AutoMigrate` must not run when `AUTO_MIGRATE` is false and the schema is missing a table → assert the boot fails.
F3. CI: `make schema-check` + `make test` + `go build` must pass; a PR that changes `internal/models/*.go` without a matching `migrations/NNNN_*` pair must fail.
F4. Fix docs: `README.md:10` and `README.md:215` currently promise up/down migrations that do not exist; update after Phase D lands.

---

## 5. Sequenced task list (smallest safe order)

```
1  A7 .gitkeep in migrations/            (unblocks Dockerfile:31)
2  A2 .env / .env.example SQLite-correct
3  A1 named volume for /app/data         (data survives restarts)
4  A5 cmd/migrate sqlite3 driver name    (migrate binary usable at all)
5  B1-B3 baseline + seed SQL             (schema becomes explicit)
6  A3 single DB lifecycle                (no two handles on one file)
7  A4 PRAGMA foreign_keys + FK test      (constraints stop being decorative)
8  C1-C4 AutoMigrate gated to dev        (prod stops mutating schema)
9  D1-D3 real migrate CLI + Makefile     (up/down/version/rollback)
10 D4 SQLite change recipes documented   (in this file + docs/)
11 E1-E3 snapshot/restore + schema-check
12 F1-F4 tests + CI + README fix
```

Verification commands to run at each step (SQLite CLI is 3.53.4 on this box):
```
sqlite3 ./data.db 'SELECT 1;'
sqlite3 ./data.db .schema
sqlite3 ./data.db 'PRAGMA foreign_keys;'
sqlite3 ./data.db 'SELECT (name) FROM cards WHERE 1;'   # per-table existence
make migrate-version && make migrate-up && make migrate-down 1 && make migrate-version
```

---

## 6. Alternatives considered, and why not now

| Option | Verdict |
|---|---|
| **Keep AutoMigrate + file snapshots only** | Cheapest, but no versioning, no schema rollback, seed data still implicit, prod schema changes on every boot. Rejected: it does not answer "can we roll back?" at the schema level. |
| **`go-gormigrate` (gormigrate)** | Adds versioned Go migrations over GORM. Attractive (same structs), but it is an extra dependency layer and its steps are Go, so SQLite-specific `ALTER` limits still bite and inverses are still hand-written. Consider later if the team prefers Go over SQL files. |
| **Atlas (declarative → versioned) / `atlas-provider-gorm`** | Best-in-class drift detection and generated migrations, but a new toolchain and SQLite support caveats. Rejected now for cost; revisit if drift keeps happening. |
| **Postgres anyway** | Explicitly out of scope (user wants SQLite for cost). |
| **Hand-written SQL, no tool** | Possible (SQLite is one file), but reinvents versioning that `golang-migrate` already provides and is already vendored in `go.mod`. |

---

## 7. Open questions (answer before Phase B)

1. Is there an existing deployed SQLite DB with data to preserve, or is a fresh baseline acceptable? (Decides whether B1 is a *dump of live* schema or a *generated* baseline.)
2. Where does the DB file live in the real deployment (volume vs bind mount, host FS type)? NFS/network FS would invalidate §2 rule 2.
3. Is `AUTO_MIGRATE=true` acceptable in dev at all, or should dev also use SQL files (single path for everyone)?
4. Are Postgres/MySQL code paths (`connection.go:15-18`, `postgresql.conf`, compose Postgres) safe to delete, or does something external still depend on them?

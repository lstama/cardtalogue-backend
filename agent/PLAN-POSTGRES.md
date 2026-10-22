# Part 1 — an existing deployment exists · Part 2 — moving to PostgreSQL later

Companion to `PLAN.md`. Same scope: DB + models only. Nothing implemented.
Assumption now fixed by the user: **there is an existing deployed SQLite DB with data in it.**
That invalidates several steps of `PLAN.md` as written. Deltas are marked `PLAN.md →`.

---

## Part 1 — What changes because the deployment already exists

### 1.1 The live DB is not in this repo (verified)

`find . -name '*.db'` → nothing; `bin/` is empty; `uploads/` has content but no DB. So the live SQLite file lives only in the deployed container/host. Every step below starts by finding it, not by generating a schema.

### 1.2 D1 — locate and rescue the file before anything else (blocking)

`Dockerfile:39-41` sets `DATABASE_URL=file:/app/data/cardtalogue.db`, `/app/data` is created by a `RUN mkdir` layer, and there is **no `VOLUME`** (`grep VOLUME Dockerfile` → none). Consequences, in order of severity:

| Situation | Meaning | Action |
|---|---|---|
| DB inside the image layer, container restarted/redeployed | Data was **silently reset** at each new container start | Data may already be gone. Check `docker exec <c> ls -l /app/data` timestamps against the deploy date |
| DB inside the image layer, container still running | Data exists but is trapped in a writable layer | `docker cp <container>:/app/data/cardtalogue.db ./live.db` immediately, then move to a volume |
| Bind mount / named volume already in place | Fine | Confirm host FS type is local (not NFS/tmpfs) |

Do this first. `PLAN.md` A1 (add the volume) is now a **data-critical migration**, not a config tweak: copy the file out, mount a named volume, restart once, verify `sqlite3 live.db 'SELECT (name) FROM cards;'` counts match before/after.

### 1.3 D2 — snapshot before touching anything

```
sqlite3 live.db .dump > audit/live-dump.sql      # text, diffable, FS-independent
cp live.db snapshots/live.<UTCstamp>.db          # whole-file snapshot
sqlite3 live.db 'SELECT 1;'                      # integrity check on both
```
Never run a migration, a `down`, or a `truncate` against the live file. All rehearsal happens on a scratch copy.

### 1.4 D3 — audit drift (this is the reason the baseline must come from the DB, not the structs)

`AutoMigrate` ran on **every boot** (`cmd/api/main.go:56`), against whatever structs were current at that deploy. So the live schema is the union of every struct version that ever shipped, minus nothing (AutoMigrate never removes columns, tables, or FK constraints). Expect the live DB to contain artefacts the current structs do not:

- orphan columns (added by an older struct, absent from `internal/models/*.go`)
- widened/retyped columns (`size:`/type changes applied in place)
- FK constraints that were **never enforced** (SQLite FKs are off by default; `connection.go:25` `gorm.Config{}` emits them but does not enable the pragma)
- FK violations in the data that no check ever caught
- dev seed rows injected by `CreateDefaultSeedData` (`migrations.go:58`)

Audit commands (no `go` needed):
```
sqlite3 live.db .schema
sqlite3 live.db .tables
diff <(sqlite3 live.db .schema) <(expected-schema-from-DB_SCHEMA.md)
```
Resolve every diff **before** writing the baseline: orphan columns are either kept (and documented) or removed by an explicit migration. Do not delete them reflexively — they may hold data an older client still reads.

### 1.5 D4 — FK violations must be found before enforcement is turned on

`PLAN.md` A4 turns FK enforcement on. On an existing DB that is a **behaviour change**, not a no-op: the pragma does not retro-validate existing rows, but new writes that touch violating rows will start failing. Run a one-time scan per FK pair and fix or accept each violation first:

```
sqlite3 live.db 'SELECT c.id, c.game_id    FROM cards AS c WHERE NOT EXISTS (SELECT 1 FROM users    AS u WHERE u.id = c.game_id);'
sqlite3 live.db 'SELECT c.id, c.language_id FROM cards AS c WHERE NOT EXISTS (SELECT 1 FROM languages AS l WHERE l.id = c.language_id);'
sqlite3 live.db 'SELECT c.id, c.pack_id     FROM cards AS c WHERE NOT EXISTS (SELECT 1 FROM packs     AS p WHERE p.id = c.pack_id);'
sqlite3 live.db 'SELECT c.id, c.rarity_id   FROM cards AS c WHERE NOT EXISTS (SELECT 1 FROM rarities  AS r WHERE r.id = c.rarity_id);'
sqlite3 live.db 'SELECT c.id, c.storage_id  FROM cards AS c WHERE c.storage_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM storages AS s WHERE s.id = c.storage_id);'
sqlite3 live.db 'SELECT c.id, c.media_id    FROM cards AS c WHERE c.media_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM media AS m WHERE m.id = c.media_id);'
```
Verified on a scratch SQLite 3.53.4 DB: `INSERT INTO cards (id, game_id) VALUES (11, 999);` **succeeded** (rc=0) despite no `users.id = 999`, and the scan then returned `11|999`. That is the proof that today's `fk_*` constraints are decorative. Nullable FKs (`storage_id`, `media_id`, the `*uint` fields in `card.go`) need the `IS NOT NULL` guard or every NULL FK is reported as a violation.

### 1.6 D5 — baseline and version adoption

- `0001_baseline.up.sql` = **the live `.schema`, cleaned**, not a generated guess. Cleaning is mandatory: `.dump`/`.schema` output contains SQLite-only artefacts — verified on a scratch DB, `.dump` emits `PRAGMA foreign_keys=OFF;` and standalone `CREATE FOREIGN KEY fk_1 (cards.game_id) REFERENCES (users.id);`. Strip pragmas; keep FK constraints but be aware that standalone form is SQLite-only (see Part 2 §P2).
- Do **not** run `0001` against the live DB (it would `CREATE` existing tables → error). Adopt the version instead: golang-migrate v4 has `Force(version)` (already wired in `cmd/migrate/main.go` as the `force` subcommand) — set current version to `1` without executing it. v4's `Migrate` methods are `Close, Down, Drop, Force, Migrate, Run, Steps, Up, Version`; there is no `SetCurrentVersion`/`Skip` in v4.20.1, so `Force` is the adoption mechanism.
- `Version()` returns `ErrNilVersion` on a fresh DB — so the same code path handles both a fresh install (run `0001`) and the existing deployment (adopt `1`). Keep that branch explicit in `cmd/migrate`.
- `0002_seed_data.up.sql` (`PLAN.md` B3) must be **skipped** for the existing deployment (its rows already exist) or written idempotently. Seed data is a fresh-install concern only.

### 1.7 D6 — rollback tiers get a hard rule

With live data, Tier 1 (inverse SQL) is the only acceptable mechanism for schema rollback on the deployed file; Tier 2 (file snapshot) is insurance, and Tier 3 (`schema-check`) is what catches "we forgot to roll back". `truncate`/`down` rehearsal on a scratch copy only. Add to the CI guard: a migration PR must ship a down file that, applied to a scratch copy of the **live** DB, returns it to the previous `schema-check` output.

### 1.8 Revised order (replaces `PLAN.md` §5)

```
1  locate live DB file; rescue to a named volume if it is in an image layer   (data-critical)
2  snapshot + .dump audit; integrity check
3  schema audit: live .schema vs structs vs DB_SCHEMA.md; list every artefact
4  FK-violation scan; fix or document each violation
5  .gitkeep in migrations/; fix .env (DATABASE_DRIVER=sqlite, file: URL)
6  fix cmd/migrate sqlite3 driver name; add Force-based version adoption
7  write 0001_baseline from the LIVE schema (pragmas stripped)
8  single DB lifecycle (one gorm.DB, closed correctly)
9  turn PRAGMA foreign_keys ON + assertion test (after step 4 is clean)
10 gate AutoMigrate to dev-only; add schema-check as the drift detector
11 real migrate CLI + Makefile targets; test down-migrations on scratch copies only
12 snapshots/restore + CI + README/docs correction
```

Steps 3 and 4 must finish before 7 and 9. `PLAN.md`'s order assumed a fresh DB.

---

## Part 2 — Future move to PostgreSQL: what must change

### P0. Cost framing (decide the trigger, not the mechanism)

`postgresql.conf` at the repo root already tunes Postgres down to `shared_buffers = 64MB`, `max_connections = 20` — a low-RAM Postgres was already intended here. So the honest cost of Postgres is not RAM tuning; it is: an extra service to run and supervise, a volume to manage, and a second schema dialect to keep in sync. Move only when a real SQLite limit is hit:

1. concurrent writers contending (SQLite = one writer at a time)
2. DB file on a network/tmpfs filesystem (locking invalid)
3. need per-user/per-tenant isolation or a second DB
4. need real types (dates, decimals, JSON, arrays) or FK/CHECK enforcement as a correctness guarantee
5. DB size/growth makes one-file snapshots an unacceptable backup unit

If none apply, stay on SQLite and keep the properties below so the move stays cheap.

### P1. Design now, pay later — six decisions that make the Postgres move a swap instead of a rewrite

| # | Decide now | Why it saves the later move |
|---|---|---|
| 1 | **One canonical set of numbered migration files**, written in a portable SQL subset (P2 table) | The same `NNNN_slug.{up,down}.sql` tree runs on both engines; only the driver changes |
| 2 | **Keep the postgres dialector branch** in `connection.go:15-18` | `PLAN.md` C4 said delete postgres/mysql — revise: delete **mysql only**, keep postgres |
| 3 | Keep `postgresql.conf` and the compose Postgres service as a **future dev harness** (or move it to `docker-compose.dev.yml`) | The harness you need on day one already exists |
| 4 | Engine tuning (`PRAGMA foreign_keys`, WAL, `busy_timeout`, `VACUUM`) lives in **Go/config**, never in migration SQL | Pragmas are SQLite-only and would poison the shared tree (verified: `.dump` emits `PRAGMA foreign_keys=OFF;`) |
| 5 | Declare explicit `size:`/`type:` on every string column in `internal/models/*.go` | SQLite is dynamically typed: a wrong/absent type is invisible today and rejected by Postgres later. Fixing it now is free |
| 6 | Make the PK explicitly auto-incrementing (`gorm:"column:id;primaryKey,autoIncrement"`) after verifying it on SQLite | SQLite gives `INTEGER PRIMARY KEY` auto-assignment implicitly; Postgres does **not** — a plain `bigint primary_key` rejects inserts with an unset ID. The app relies on this: `internal/repository/card_repository.go:35` does `db.Create(card)` with a client-supplied struct, and `sample-data/initialize.sql` inserts **without** an `id` column |

Item 6 is the single biggest silent breakage in a SQLite→Postgres move. Verify before changing anything: a test that creates a `Card` with `ID` unset and asserts the assigned ID on both engines.

### P2. Portability table — what breaks in the move

| Construct | SQLite (here) | Postgres | Rule for this repo |
|---|---|---|---|
| `CREATE FOREIGN KEY fk_1 (cards.game_id) REFERENCES (users.id);` (standalone) | valid — this is what the SQLite dialector emits (verified in a scratch `.dump`) | **invalid** | Write FKs inline in `CREATE TABLE` (`FOREIGN KEY fk_1 (col) REFERENCES (tbl.col)`) — valid in both — or keep the SQLite-only form only in a SQLite-only tree |
| `PRAGMA …` | SQLite-only | rejected | Never in migration files (P1.4) |
| `.dump` output (`INSERT INTO users('id','name') VALUES('1','a');`) | SQLite's own format | **invalid** — single-quoted identifiers | Never feed `.dump` to psql; use it for audit only |
| Bulk data transfer | one file | CSV + `\.` / `COPY FROM` | Data moves via CSV or a Go port using the existing structs, not via `.dump` |
| `ALTER TABLE` | exactly **one** clause per statement | many clauses allowed | Always write one clause per statement → portable both ways |
| `ADD COLUMN … DEFAULT NULL` | fine | fine (O(1) with constant default) | Portable; keep it |
| `DROP COLUMN` | fails if PK / indexed / UNIQUE / FK / `CHECK` / trigger | fine, but drops dependent indexes/constraints | Portable in syntax, **not** in semantics → keep the SQLite precondition in the file header comment |
| Column rename | does not exist | `ALTER … RENAME COLUMN` exists | Keep the add → copy → drop recipe (works on both); do not use Postgres-only rename |
| Type change | table rebuild | table rewrite | Keep add → copy → drop |
| `NOT NULL`, `PRIMARY KEY`, `UNIQUE`, `CHECK`, `CREATE UNIQUE INDEX` | supported | supported | Portable; keep |
| FK actions `ON DELETE CASCADE` / `ON UPDATE SET NULL` | supported (SQLite ≥ 3.31) **but only when FK enforcement is on** | always enforced | If cascade semantics are wanted, put them in SQL — note that `card.go:37` `constraint:OnUpdate:CASCADE,OnDelete:SET NULL` is a **GORM-level** constraint, not SQL, so it is not in the schema today |
| Type names (`bigint`, `integer`, `varchar(n)`, `text`, `boolean`, `timestamp`) | any name accepted (dynamic typing) | strict whitelist | Keep a whitelist in the migration file header; a typo is a silent SQLite bug and a hard Postgres failure |
| `boolean` storage | `0`/`1` | `true`/`false` | Pin the representation in migrations, not in Go defaults |
| Date/time columns | no date type; GORM `time` maps per dialect | `timestamptz` | `CreatedAt`/`UpdatedAt`/`DeletedAt` (`gorm.DeletedAt`, `user.go:18`, `card.go:24`) need an explicit conversion pass during the move |
| Identifier case | case-insensitive | unquoted identifiers fold to **lowercase** | snake_case only (GORM already does this); reject camelCase in review |
| Reserved words | smaller set | larger set | Quote or rename on review; check before merge |
| FK enforcement | opt-in, off by default | always on | Today's `fk_*` constraints are decorative; the move makes them load-bearing, so Part 1 §1.4's violation scan is a **precondition** for the move, not just for the pragma |

### P3. The move, when it happens (10 steps)

```
1  freeze: no schema change in flight; snapshot the live SQLite file
2  audit: FK-violation scan (Part 1 §1.4) must be clean
3  schema translation pass: strip pragmas; inline FKs; pin types from the whitelist;
   PK -> bigserial/IDENTITY; dates -> timestamptz; booleans -> boolean
4  create empty Postgres DB; apply the translated 0001_baseline as Postgres SQL
5  bulk data copy per table: sqlite3 -csv -> psql COPY FROM (never .dump)
6  re-validate: FK scan, NOT NULL scan, row counts per table vs SQLite
7  switch: DATABASE_DRIVER=postgres, DATABASE_URL=postgres://… (connection.go:15-18 already supports it)
8  migrations: same numbered tree, new driver; if any file needed SQLite-only syntax,
   split into migrations/sqlite/ + migrations/postgres/ with identical numbering
9  re-run the whole test suite (FK assertion, soft-delete, PK auto-increment, type widths)
10 keep the SQLite file + snapshots as the rollback target for one release, then retire
```

Step 6 is the gate: a Postgres move is where previously-invisible FK violations become fatal. Fix them on SQLite first (Part 1 §1.4), not during the move.

### P4. What must NOT change, or the move is not cheap

Versioned numbered SQL files; additive-only column discipline; `cmd/migrate` as the only schema mutator; `AutoMigrate` off in production; FK-enforcement assertion test; `schema-check` drift detector; whole-file snapshots. All of these are in `PLAN.md` and are engine-neutral — that is the whole point of doing the SQLite plan properly now.

### P5. One migration tree or two

| Option | When | Verdict |
|---|---|---|
| **One tree, portable subset** (P2 rules enforced in review) | default | Preferred: version numbers stay aligned across engines, so the move is a driver swap |
| `migrations/sqlite/` + `migrations/postgres/`, same numbering | only if a file genuinely needs engine-only syntax | Acceptable fallback; costs a rule "both trees must contain the same NNNN" |
| Per-engine generated trees (Atlas) | if drift keeps happening | Deferred, as in `PLAN.md` §6 |

Enforce the portable subset with a lint, not with prose: grep migration files for `PRAGMA`, standalone `CREATE FOREIGN KEY`, camelCase identifiers, and type names outside the whitelist. Make it a CI check next to `schema-check`.

### P6. Open questions (answer before Part 1 step 7)

1. Where does the live SQLite file actually live (image layer / named volume / bind mount), and is its host FS local?
2. Has the deployed container been restarted since data was written (i.e. is the data already lost)?
3. Which struct versions have been deployed since the DB was created — do we know the schema history well enough to explain every orphan column found in the audit?
4. Is the FK-violation scan clean, or does the catalogue already contain dangling references?
5. Do you want FK/CHECK enforcement to be load-bearing now (recommended) or only after the Postgres move?

# Notes Index

- `progress.md` — current goal, state, next steps.
- `decisions.md` — dated decisions + rejected alternatives (append-only).
- `scratch/db-schema-model.md` — full DB/model reference: where schema lives, how it is applied, GORM AutoMigrate limits, rollback status, broken/stale DB plumbing.
- `../DB_SCHEMA.md` (repo root) — user-facing onboarding doc: how to change schema/models + rollback status.
- `../agent/PLAN.md` (repo root) — production plan for SQLite-only DB (fresh-install assumption): phases A-F, 12-step order, SQLite ALTER limits, rollback tiers.
- `../agent/PLAN-POSTGRES.md` (repo root) — Part 1: deltas when an existing SQLite deployment exists (rescue file, drift audit, FK-violation scan, version adoption via Force). Part 2: what to change now so a future PostgreSQL move is a driver swap (portable SQL subset, type/PK rules, SQLite-only artefacts to avoid).

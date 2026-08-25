# Data Model: Go Core Skeleton (F101)

**Feature**: `003-go-core-skeleton`  
**Scope**: Port existing governance DDL only — no new tables

F101 does not introduce new domain tables. It ports Alembic revision
`0001_lore_audit` to goose so Go persistence slices (F103+) can use sqlc
against the same schema Python uses today.

## Tables (unchanged semantics)

### `lore_entries`

| Column | Type | Notes |
|--------|------|-------|
| id | varchar(36) PK | UUID string |
| statement | text | max 8000 enforced in domain |
| scope_kind | varchar(64) | e.g. repository, team |
| scope_key | varchar(512) | trimmed scope identifier |
| origin | varchar(64) | human_authored in current slices |
| verification_status | varchar(32) | unverified \| verified |
| evidence | jsonb | `[{type, value}, ...]` |
| created_by | varchar(256) | actor id |
| created_at | timestamptz | |
| verified_by | varchar(256) nullable | |
| verified_at | timestamptz nullable | |
| updated_at | timestamptz | |

**Index**: `ix_lore_entries_scope_created (scope_kind, scope_key, created_at)`

### `audit_records`

| Column | Type | Notes |
|--------|------|-------|
| id | varchar(36) PK | |
| target_id | varchar(36) | lore entry id |
| action | varchar(32) | create \| verify |
| actor_id | varchar(256) | |
| created_at | timestamptz | |

**Indexes**:

- `ix_audit_records_target_id (target_id)`
- `ix_audit_records_target_created (target_id, created_at, id)`

## Source of truth during migration

| Tool | Role in F101 |
|------|----------------|
| Alembic `0001` | Python app (legacy, still active) |
| goose `00001` | Go module (parity copy) |

Future schema changes SHOULD be applied to goose first once Go owns migrations,
with Alembic updated or retired per slice.

## Out of scope (future features)

- `outbox_events`
- `users`, `teams`, `projects`, `repositories`
- authority factor columns on lore
- graph-service tables (none in Postgres for Graphiti data)

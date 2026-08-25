# Feature Specification: Go PostgreSQL Persistence (F103)

**Feature Branch**: `005-go-postgres-persistence`  
**Created**: 2026-08-25  
**Depends on**: F101 (skeleton), F102 (domain)

## Goal

Implement Go governance-plane persistence for lore entries and audit records using
pgx + sqlc, with application ports matching the Python hexagonal layout.

## Acceptance Criteria

- Application ports: `LoreRepository`, `AuditRepository`, `UnitOfWork`
- sqlc queries: insert/get/update/list for lore; insert/list for audit
- Repositories map between `internal/domain` types and PostgreSQL rows
- Evidence stored as JSONB `[{type, value}, ...]` matching Python
- `list_by_scope` orders by `created_at DESC`; audits by `created_at ASC, id ASC`
- Integration test (build tag `integration`) round-trips create + verify + list when Postgres is up
- Python tests unchanged and passing

## Out of Scope

- Application command handlers (F104)
- HTTP/MCP adapters
- Connection pooling bootstrap / config (minimal DSN helper OK)

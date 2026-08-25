# Feature Specification: Go REST Lore CRUD (F104)

**Branch**: `006-go-rest-lore-crud`  
**Depends on**: F102 (domain), F103 (persistence)

## Goal

Expose governance-plane lore CRUD/verify/audit/list via Go REST `/v1/lore-entries`,
matching the Python contract in `specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`.

## Acceptance Criteria

- Application handlers mirror Python create/verify/get/list/audits behavior
- chi HTTP adapter with error envelope `validation_error` / `not_found`
- `X-Memlore-Actor` required on mutating routes
- Go contract tests mirror `tests/contract/test_create_lore_entry.py` and `test_lore_flow.py`
- `memlore serve` starts Go HTTP server (Python `uv run memlore serve` remains available)
- Integration test optional with Postgres

## Out of Scope

- MCP (F105)
- Disabling Python REST

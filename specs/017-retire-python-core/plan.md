# Implementation Plan: Retire legacy Python governance core

**Branch**: `017-retire-python-core` | **Date**: 2026-08-31 | **Spec**: [spec.md](./spec.md)

## Summary

Delete the strangler leftover (`src/memlore/` REST/MCP/Alembic/root pytest).
Keep `graph-service/`. Make Go the only documented/CI core path. Align
constitution and architecture docs with ADR-0005 after cutover.

## Technical Context

**Language/Version**: Go 1.25+ core; Python 3.12 only in `graph-service/`  
**Primary Dependencies**: none new; remove FastAPI/SQLAlchemy/MCP-Python from root  
**Storage**: PostgreSQL via goose (unchanged)  
**Testing**: `go test ./...`; graph-service pytest unchanged  
**Target Platform**: MemLore Core + graph-service  
**Project Type**: hexagonal Go service + thin Python graph service  
**Performance Goals**: n/a (deletion)  
**Constraints**: Do not change Go behavior; do not touch graph-service code  
**Scale/Scope**: Delete legacy tree + docs/CI/constitution sync

## Constitution Check

- [x] TDD: no new core behavior; deletion verified by `go test ./...` and CI yaml review
- [x] Spec-driven: removal scope and keep-list encoded
- [x] Architecture integrity: governance stays Go; knowledge plane stays graph-service
- [x] Documentation: README, setup, ADRs, constitution, architecture
- [x] Authority & provenance: no change to evaluator
- [x] Temporal correctness: no change
- [x] Secure by default: no new auth surface
- [x] Observability: no change
- [x] Dependency policy: remove unused Python deps
- [x] Simplicity: no compatibility shim

## Project Structure

```text
specs/017-retire-python-core/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md

# Removed
src/memlore/
tests/          # root pytest only
alembic/
pyproject.toml
uv.lock
alembic.ini
.python-version

# Kept
cmd/ internal/ migrations/ graph-service/ docs/
```

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

See [data-model.md](./data-model.md), [quickstart.md](./quickstart.md).

## Constitution re-check (post-design)

Gates pass if constitution is amended to Go core + graph-service Python, and
Alembic is no longer required for governance schema.

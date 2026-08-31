# Tasks: Retire legacy Python governance core

**Input**: `specs/017-retire-python-core/`

## Phase 1: Setup

- [x] T001 Confirm branch `017-retire-python-core` and spec artifacts

## Phase 2: Delete legacy tree

- [x] T002 Remove `src/memlore/`, root `tests/`, `alembic/`, `alembic.ini`, root `pyproject.toml`, `uv.lock`, `.python-version`

## Phase 3: CI + env

- [x] T003 Remove CI `quality` job in `.github/workflows/ci.yml`; keep go + graph-service
- [x] T004 Use `postgresql://` DSN in `.env.example`

## Phase 4: Docs + constitution

- [x] T005 Update README, `docs/development/setup.md`, `testing.md`, `contributing.md`, `docs/operations/migrations.md`
- [x] T006 Update `docs/architecture/overview.md`, `containers.md`; ADR-0002/0005 notes; FEATURE_DEVELOPMENT; specify-rules
- [x] T007 Amend constitution to v1.1.0 (Go core + graph-service Python; goose); sync plan/tasks templates

## Phase 5: Verify

- [x] T008 `go test ./...`; confirm `graph-service/` tree intact

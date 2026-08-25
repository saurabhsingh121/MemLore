# Tasks: Go Core Project Skeleton (F101)

**Input**: Design documents from `/specs/003-go-core-skeleton/`  
**Prerequisites**: plan.md, spec.md, research.md, ADR-0005  
**Tests**: TDD for verifiable behaviors (version command, migration parity)

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [ ] T001 Create `go.mod` (`github.com/memlore/memlore`, `go 1.25`) and `go.sum`
- [ ] T002 [P] Add `sqlc.yaml` pointing at `db/queries/` and postgres engine
- [ ] T003 [P] Add package stubs: `internal/domain`, `application`, `adapters`, `infrastructure/postgres` with `doc.go` only

## Phase 2: Foundational

- [ ] T004 RED: Failing test that `cmd/memlore` prints version and exits 0 in `cmd/memlore/main_test.go`
- [ ] T005 GREEN: Implement `main.go` with `version` subcommand using slog
- [ ] T006 [P] Port Alembic `0001` to `migrations/00001_lore_audit.sql` (goose format)
- [ ] T007 [P] Add `db/queries/smoke.sql` (`SELECT 1 AS ok`) and run `sqlc generate`
- [ ] T008 RED: Test migration SQL contains `lore_entries` and `audit_records` with expected indexes (parse file test in `migrations/migration_test.go`)
- [ ] T009 GREEN: Fix migration file until T008 passes

## Phase 3: User Story 1 — Toolchain health (P1)

- [ ] T010 [US1] RED: CI workflow lacks Go job — add failing expectation or local script check
- [ ] T011 [US1] GREEN: Add `go-test` job to `.github/workflows/ci.yml` (`go test`, `go vet`)
- [ ] T012 [US1] Verify `uv run pytest` still passes (no Python changes)

**DONE WHEN**: `go test ./...` and `go vet ./...` pass locally and in CI.

## Phase 4: User Story 2 — Migration scaffold (P1)

- [ ] T013 [US2] RED: Optional integration test (build tag `integration`) applies goose up against testcontainers Postgres
- [ ] T014 [US2] GREEN: Implement test or document manual quickstart verification; table parity asserted

**DONE WHEN**: Fresh DB + goose up creates expected tables.

## Phase 5: User Story 3 — Layout contract (P1)

- [ ] T015 [US3] RED: Test that required paths from `contracts/go-module-layout.md` exist
- [ ] T016 [US3] GREEN: Add any missing paths; ensure no forbidden empty packages

**DONE WHEN**: Layout contract test passes.

## Phase 6: Polish

- [ ] T017 [P] Update `docs/adr/README.md` and mark ADR-0002 superseded for core
- [ ] T018 [P] Update `docs/development/FEATURE_DEVELOPMENT.md` (F002 DONE, F101 READY→IN DEVELOPMENT)
- [ ] T019 [P] Update `docs/development/migration-inventory.md` (Go skeleton row)
- [ ] T020 [P] Update `.cursor/rules/specify-rules.mdc` active plan → `003-go-core-skeleton`
- [ ] T021 Run full quality gates: `go test`, `go vet`, ruff, mypy, pytest

**Feature complete when**: SC-001 through SC-004 satisfied; Python runtime unchanged.

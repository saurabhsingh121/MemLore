# Tasks: Go PostgreSQL Persistence (F103)

## Phase 1: Ports and queries

- [x] T001 Create `specs/005-go-postgres-persistence/` and application ports
- [x] T002 Add sqlc queries in `db/queries/lore.sql` and `audit.sql`
- [x] T003 Generate or commit sqlc output for lore/audit queries

## Phase 2: Mapping and repositories

- [x] T004 RED: `mapping_test.go` — domain ↔ row conversion including evidence JSON
- [x] T005 GREEN: `mapping.go`
- [x] T006 RED: `lore_repository_test.go` — unit tests with mock Queries interface or pgx mock
- [x] T007 GREEN: `lore_repository.go`
- [x] T008 GREEN: `audit_repository.go`

## Phase 3: Unit of work

- [x] T009 RED: `unit_of_work_test.go` — commit/rollback with integration or tx test
- [x] T010 GREEN: `unit_of_work.go` using pgx transaction

## Phase 4: Integration

- [x] T011 RED: `repository_integration_test.go` (tag `integration`) — parity with Python `test_lore_postgres.py`
- [x] T012 GREEN: wire repositories; make integration test pass

## Phase 5: Polish

- [x] T013 Update FEATURE_DEVELOPMENT, migration-inventory, specify-rules
- [x] T014 Run `go test ./...`, `go vet ./...`, `uv run pytest`

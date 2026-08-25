# Tasks: Go Domain Primitives (F102)

## Phase 1: Setup

- [x] T001 Create `specs/004-go-domain-primitives/` artifacts and branch `004-go-domain-primitives`
- [x] T002 Add `internal/domain/characterization_test.go` documenting Python parity sources

## Phase 2: Errors and enums

- [x] T003 RED: `errors_test.go` — ValidationError type and message formatting
- [x] T004 GREEN: `errors.go`
- [x] T005 RED: `enums_test.go` — all enum string values match Python
- [x] T006 GREEN: `enums.go` with parse helpers

## Phase 3: Scope and evidence

- [x] T007 RED: `scope_test.go` — trim, blank key, max length (from `test_scope_evidence.py`)
- [x] T008 GREEN: `scope.go`
- [x] T009 RED: `evidence_test.go` — trim, blank, max length
- [x] T010 GREEN: `evidence.go`

## Phase 4: Lore, audit, verification

- [x] T011 RED: `lore_test.go` — defaults, oversized statement, origin rule (`test_lore_entry.py`)
- [x] T012 GREEN: `lore.go`
- [x] T013 RED: `audit_test.go` — actor/target validation
- [x] T014 GREEN: `audit.go`
- [x] T015 RED: `verification_test.go` — apply verify + idempotent (`verification.py`)
- [x] T016 GREEN: `verification.go`

## Phase 5: Polish

- [x] T017 Remove `internal/domain/doc.go` placeholder; package is real
- [x] T018 Update FEATURE_DEVELOPMENT, migration-inventory, specify-rules
- [x] T019 Run `go test ./...`, `go vet ./...`, `uv run pytest`

# Tasks: Governance lifecycle — invalidate + supersede (F110)

**Input**: Design documents from `/specs/013-supersede-invalidate/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/lifecycle-lore.md

**Tests**: Constitution TDD — RED → GREEN → REFACTOR for every behavioral task.

## Phase 1: Setup

- [x] T001 Confirm branch `013-supersede-invalidate` and spec artifacts under `specs/013-supersede-invalidate/`

---

## Phase 2: Foundational (domain enums + persistence columns)

- [x] T002 [P] RED: Extend `internal/domain/enums_test.go` for `invalidated` verification status and `invalidate`/`supersede` audit actions
- [x] T003 GREEN: Add enum constants in `internal/domain/enums.go`
- [x] T004 [P] RED: Domain tests in `internal/domain/lore_test.go` asserting new LoreEntry fields default null
- [x] T005 GREEN: Add `SupersededByID`, `InvalidatedBy`, `InvalidatedAt` to `internal/domain/lore.go`
- [x] T006 RED: Migration content test in `migrations/migration_test.go` for `00003` columns
- [x] T007 GREEN: Add `migrations/00003_supersede_invalidate.sql`
- [x] T008 Update `db/queries/lore.sql`; regenerate sqlc; map fields in `internal/infrastructure/postgres/mapping.go`
- [x] T009 Update inline integration schemas with `ADD COLUMN IF NOT EXISTS`

---

## Phase 3: User Story 1 — Invalidate (P1)

- [x] T010 [P] [US1] RED: `internal/domain/lifecycle_test.go` invalidate cases
- [x] T011 [P] [US1] RED: `internal/application/commands/invalidate_lore_test.go`
- [x] T012 [P] [US1] RED: `ApplyVerification` rejects invalidated/superseded
- [x] T013 [US1] GREEN: `ApplyInvalidation` in `internal/domain/lifecycle.go`
- [x] T014 [US1] GREEN: Reject invalidated/superseded in `internal/domain/verification.go`
- [x] T015 [US1] GREEN: `InvalidateLoreHandler` in `internal/application/commands/invalidate_lore.go`

---

## Phase 4: User Story 2 — Supersede (P1)

- [x] T016 [P] [US2] RED: supersession domain tests in `internal/domain/lifecycle_test.go`
- [x] T017 [P] [US2] RED: `internal/application/commands/supersede_lore_test.go`
- [x] T018 [US2] GREEN: `ApplySupersession` in `internal/domain/lifecycle.go`
- [x] T019 [US2] GREEN: `SupersedeLoreHandler` in `internal/application/commands/supersede_lore.go`

---

## Phase 5: User Story 3 — REST + MCP parity (P2)

- [x] T020 [P] [US3] RED: HTTP contract tests in `internal/adapters/http/lore_contract_test.go`
- [x] T021 [P] [US3] RED: MCP contract tests in `internal/adapters/mcp/mcp_contract_test.go`
- [x] T022 [US3] GREEN: Presenter fields in `internal/adapters/presenters/lore.go`
- [x] T023 [US3] GREEN: REST handlers + routes in `internal/adapters/http/handlers.go` and `dto.go`
- [x] T024 [US3] GREEN: MCP tools + registration in `internal/adapters/mcp/tools.go` and `server.go`

---

## Phase 6: Polish & docs

- [x] T025 [P] Update `docs/api/mcp.md` (deferred → implemented)
- [x] T026 [P] Update `specs/001-scoped-lore-entry/contracts/rest-lore-entries.md` and `docs/api/rest.md`
- [x] T027 [P] Update `docs/development/FEATURE_DEVELOPMENT.md`; `.cursor/rules/specify-rules.mdc`
- [x] T028 Run `go test ./...`; run integration tests if Postgres is up
- [x] T029 Optional dogfood: create + supersede via REST on local Postgres (MCP parity covered by contract tests)

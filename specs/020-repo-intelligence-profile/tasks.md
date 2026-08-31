# Tasks: Repository Intelligence Profile (F020)

**Input**: Design documents from `/specs/020-repo-intelligence-profile/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD (RED → GREEN → REFACTOR) for behavioral work.

## Phase 1: Setup

- [x] T001 Create spec artifacts under `specs/020-repo-intelligence-profile/` and pin `.specify/feature.json`

## Phase 2: Foundational

- [x] T002 [P] RED/GREEN: `internal/application/context/profile.go` classifier + `profile_test.go`
- [x] T003 Handler types: `internal/application/queries/repository_profile.go` (list current lore + graph search → rank → budget → classify)
- [x] T004 Presenter: `internal/adapters/presenters/repository_profile.go`

**Checkpoint**: Classifier + handler usable from adapters

## Phase 3: User Story 1 — Compact briefing (P1)

- [x] T005 RED/GREEN: `internal/application/queries/repository_profile_test.go` — sections, omit empty, empty repo OK, ADR → decisions
- [x] T006 REST `POST /v1/repository-profile` in `internal/adapters/http/{handlers,dto}.go`
- [x] T007 RED/GREEN: `internal/adapters/http/repository_profile_contract_test.go`

## Phase 4: User Story 2 — Trust and conflicts (P1)

- [x] T008 RED/GREEN: handler tests — stale omitted, conflicts surfaced, verified ranks above unverified

## Phase 5: User Story 3 — REST / MCP / CLI parity (P2)

- [x] T009 MCP `memlore.repo_profile` in `internal/adapters/mcp/{server,tools}.go`
- [x] T010 Update MCP contract tests to 10 tools + repo_profile call (`mcp_contract_test.go`, `membership_contract_test.go`)
- [x] T011 CLI `memlore profile` — `internal/adapters/cli/profile.go` + `cmd/memlore/main.go`
- [x] T012 CLI tests: usage, missing `--repository`, format briefing

## Phase 6: Polish

- [x] T013 [P] Update `docs/api/rest.md`, `docs/api/mcp.md`, README CLI, `docs/architecture/target-architecture.md` MCP table
- [x] T014 Update `docs/development/FEATURE_DEVELOPMENT.md` F020 status; `.cursor/rules/specify-rules.mdc`
- [x] T015 `go test ./...` and `go vet ./...` green

## Parallel notes

T002 can start immediately. T006 depends on T003–T004. T009 depends on T003–T004. T011 depends on T003 and T004.

# Tasks: Agent Context Bootstrap / richer get_for_task (F021)

**Input**: Design documents from `/specs/021-agent-context-bootstrap/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD (RED → GREEN → REFACTOR) for behavioral work.

## Phase 1: Setup

- [x] T001 Create spec artifacts under `specs/021-agent-context-bootstrap/` and pin `.specify/feature.json`

## Phase 2: Foundational

- [x] T002 [P] RED/GREEN: packet classify (`task_context`, briefing ids, omit empty) in `internal/application/context/profile.go` + `profile_test.go`
- [x] T003 Extend `CompileContextQuery`/`Result` and `NewCompileContextHandler(search, list)` in `internal/application/queries/compile_context.go`; update existing call sites to pass `list` or `nil`
- [x] T004 Presenter additive fields in `internal/adapters/presenters/context_packet.go`

**Checkpoint**: Packet classify + compile handler types ready for adapters

## Phase 3: User Story 1 — Named task briefing (P1)

- [x] T005 RED/GREEN: `internal/application/queries/compile_context_test.go` — sections populated, empty omitted, items kept, stale omitted, verified architecture outranks unverified, conflicts listed, repository briefing-class merge
- [x] T006 REST additive JSON on `POST /v1/context/compile` in `internal/adapters/http/{handlers,dto}.go`
- [x] T007 RED/GREEN: `internal/adapters/http/context_compile_contract_test.go` — v1 body 200; sections/sources when data exists

## Phase 4: User Story 2 — Richer optional inputs (P1)

- [x] T008 RED/GREEN: handler tests — omitted optionals do not fail; files/ticket influence `task_context`; `agent_id` does not change scores; search text includes files/ticket; token budget still caps

## Phase 5: User Story 3 — REST / MCP / CLI parity (P2)

- [x] T009 MCP `memlore.get_for_task` additive args in `internal/adapters/mcp/{server,tools}.go` (tool count stays 10)
- [x] T010 RED/GREEN: MCP contract tests — parity with REST fixture; tool count 10 (`mcp_contract_test.go`)
- [x] T011 CLI `memlore context` — `internal/adapters/cli/context.go` + `cmd/memlore/main.go`
- [x] T012 CLI tests: usage, missing `--task`/`--repository`, format briefing (`internal/adapters/cli/context_test.go`, `cmd/memlore/main_test.go`)

## Phase 6: Polish

- [x] T013 [P] Update `docs/api/rest.md`, `docs/api/mcp.md`, README CLI, architecture MCP notes if present
- [x] T014 Update `docs/development/FEATURE_DEVELOPMENT.md` F021 DONE; next F022 or F030 per tracker; `.cursor/rules/specify-rules.mdc`
- [x] T015 `go test ./...` and `go vet ./...` green

## Parallel notes

T002 can start immediately. T005 depends on T003. T006 depends on T003–T004. T009 depends on T003–T004. T011 depends on T003 and T004.

## Implementation strategy

MVP is US1 (named sections on existing compile). US2 is additive inputs. US3 is surface parity. Do not add an MCP tool or persist packets.

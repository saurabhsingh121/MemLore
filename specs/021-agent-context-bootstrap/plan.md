# Implementation Plan: Agent Context Bootstrap / richer get_for_task (F021)

**Branch**: `021-agent-context-bootstrap` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Extend F007 compile (`CompileContextHandler`, REST `POST /v1/context/compile`,
MCP `memlore.get_for_task`) with additive optional inputs and named packet
sections. Reuse F020 `ClassifyItem` / section ids. Merge repository
list-by-scope briefing lore (architecture, decisions, conventions, gotchas)
with task search so agents get both repo intent and task-specific hits.
Keep `items[]` for backward compatibility. Add CLI `memlore context`. No new
tables, no new MCP tool, no F022/F023.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi, pgx, sqlc, goose, Go MCP SDK (no new libs)  
**Storage**: PostgreSQL + graph-service as today — no migration  
**Testing**: `go test ./...`; unit tests for packet classification + compile handler; REST/MCP contract tests; CLI usage/format tests  
**Target Platform**: `memlore serve`, `memlore mcp`, `memlore context`  
**Project Type**: CLI + REST + MCP governance service  
**Performance Goals**: Same retrieval limit/budget as compile v1 (default 4096 tokens, retrieval limit 20)  
**Constraints**: TDD; domain/application pure of HTTP/MCP; F007 v1 fields remain; MCP tool count stays 10; `graph-service/` unchanged  
**Scale/Scope**: One compile request; on-read packet; no persistence of packets

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for packet classify, compile handler, REST/MCP contracts, CLI
- [x] Spec-driven: FR-001–FR-017 and SC-001–SC-008; no remaining clarifications
- [x] Architecture integrity: packet mapping in application/context; adapters thin; no Graphiti types; no distributed TX
- [x] Documentation: rest.md, mcp.md, FEATURE_DEVELOPMENT F021 → DONE; specify-rules; README CLI line
- [x] Authority & provenance: reuse EvaluateAuthority; agent_id never an authority factor; no silent canonicalization
- [x] Temporal correctness: FilterCurrent + DetectConflicts reused; no invented drift; stale omitted
- [x] Secure by default: same scope membership gate as compile
- [x] Observability: reuse graph_service_unavailable warning; slog on CLI/handler errors
- [x] Engineering intelligence: named task packet sections, not a generic memory dump
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: extend compile handler; reuse F020 classifier; no packet table

**Post-design re-check**: Pass — additive query fields, packet classify helper, three adapters, CLI format.

## Project Structure

### Documentation (this feature)

```text
specs/021-agent-context-bootstrap/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/context-packet.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/application/context/profile.go          # add SectionTaskContext; packet classify
internal/application/context/profile_test.go
internal/application/queries/compile_context.go  # optional inputs + list merge
internal/application/queries/compile_context_test.go
internal/adapters/presenters/context_packet.go   # additive sections/sources/echo
internal/adapters/http/{handlers,dto}.go         # additive compile JSON
internal/adapters/http/context_compile_contract_test.go
internal/adapters/mcp/{server,tools}.go          # get_for_task additive args
internal/adapters/mcp/mcp_contract_test.go
cmd/memlore/main.go                              # context subcommand
internal/adapters/cli/context.go                 # human-readable format
docs/api/{rest,mcp}.md
docs/development/FEATURE_DEVELOPMENT.md
```

**Structure Decision**: Go core only. `graph-service/` unchanged.

## Complexity Tracking

> No constitution violations requiring justification.

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

See [data-model.md](./data-model.md), [contracts/context-packet.md](./contracts/context-packet.md),
[quickstart.md](./quickstart.md).

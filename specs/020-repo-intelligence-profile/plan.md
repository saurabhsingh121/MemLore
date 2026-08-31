# Implementation Plan: Repository Intelligence Profile (F020)

**Branch**: `020-repo-intelligence-profile` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Compile a token-budgeted **RepositoryProfile** on read from current
governance + graph knowledge for a `repository` scope. Classify ranked items
into named engineering sections; omit empty sections and unmatched items.
Expose REST `POST /v1/repository-profile`, MCP `memlore.repo_profile`, and
CLI `memlore profile --repository`. Reuse F007 compile pipeline (search →
temporal filter → conflicts → authority rank → token budget). No new tables.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi, pgx, sqlc, goose, Go MCP SDK (no new libs)  
**Storage**: PostgreSQL + graph-service as today — no migration  
**Testing**: `go test ./...`; unit tests for classifier + handler; REST/MCP contract tests; CLI usage/format tests  
**Target Platform**: `memlore serve`, `memlore mcp`, `memlore profile`  
**Project Type**: CLI + REST + MCP governance service  
**Performance Goals**: Same retrieval limit/budget as compile v1 (default 4096 tokens, retrieval limit 20)  
**Constraints**: TDD; domain/application pure of HTTP/MCP; F007/F114/F112 behavior unchanged for existing tools; MCP tool count 9 → 10  
**Scale/Scope**: One repository per request; compiled overview, not ingest

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for classifier, handler, REST/MCP contracts, CLI format
- [x] Spec-driven: FR-001–FR-014 and SC-001–SC-006; no remaining clarifications
- [x] Architecture integrity: classifier + handler in application; adapters thin; no Graphiti types; no distributed TX
- [x] Documentation: rest.md, mcp.md, FEATURE_DEVELOPMENT F020 → IN DEVELOPMENT/DONE; README CLI line
- [x] Authority & provenance: reuse EvaluateAuthority; items keep evidence/factors; no silent canonicalization
- [x] Temporal correctness: FilterCurrent + DetectConflicts reused; stale omitted
- [x] Secure by default: same scope membership gate as compile
- [x] Observability: reuse graph_service_unavailable warning; slog on CLI/handler errors
- [x] Engineering intelligence: named engineering sections, omit unmatched dump
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: compile-then-classify; no profile table

**Post-design re-check**: Pass — one query handler, pure classifier, three adapters.

## Project Structure

### Documentation (this feature)

```text
specs/020-repo-intelligence-profile/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/repository-profile.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/application/context/profile.go          # section ids + Classify
internal/application/context/profile_test.go
internal/application/queries/repository_profile.go
internal/application/queries/repository_profile_test.go
internal/adapters/presenters/repository_profile.go
internal/adapters/http/{handlers,dto}.go         # POST /v1/repository-profile
internal/adapters/http/repository_profile_contract_test.go
internal/adapters/mcp/{server,tools}.go          # memlore.repo_profile
internal/adapters/mcp/mcp_contract_test.go       # tool count 10
cmd/memlore/main.go                              # profile subcommand
internal/adapters/cli/profile.go                 # human-readable format
docs/api/{rest,mcp}.md
docs/development/FEATURE_DEVELOPMENT.md
```

**Structure Decision**: Go core only. `graph-service/` unchanged.

## Complexity Tracking

> No constitution violations requiring justification.

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

See [data-model.md](./data-model.md), [contracts/repository-profile.md](./contracts/repository-profile.md),
[quickstart.md](./quickstart.md).

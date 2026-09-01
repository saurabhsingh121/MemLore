# Implementation Plan: First-Class Decision Model (F040)

**Branch**: `040-decision-model` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Treat engineering decisions as a dedicated domain, not lore snippets.
Human create persists a `decisions` row (structured question, choice,
rationale, alternatives, consequences, owner, date, components) **and** a
linked lore entry with the **same id** (verified `human_authored`, optional
evidence). Current F032 accepted-ADR lore is **projected** as Decisions at
read time (public id = lore id, source kind `adr`) so the choice is not
copied as a second current fact. Supersede writes a new Decision + lore and
applies F110 lore supersession; history remains gettable. Compile /
`get_for_task` keep section id `decisions`; first-class Decision lore is
classified into that section without changing ranking formulas. Surfaces:
CLI `memlore decision …` and REST `/v1/decisions`. MCP stays at 10 tools.
F035 Accept does not create Decisions.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi, pgx, sqlc, goose, slog; no new libraries  
**Storage**: PostgreSQL — goose migration `00009_decisions.sql` for `decisions` + `decision_alternatives` + `decision_components`; lore_entries reused; ADR projection is read-side (no F032 table changes)  
**Testing**: `go test ./...`; `go vet ./...`; domain tests for create/supersede/history/ADR projection; handler tests with memory UoW; REST + membership contract tests; CLI tests; compile characterization  
**Target Platform**: `memlore decision create|get|list|supersede`, `memlore serve`, existing `memlore worker` (outbox only)  
**Project Type**: CLI + REST governance service (MCP unchanged)  
**Performance Goals**: List-current is one-scope lore list plus decisions-by-scope; v1 in-process mutate  
**Constraints**: TDD; domain independent of HTTP/SQL; no 11th MCP tool; do not reopen F007/F020/F021/F030/F031/F032/F035 except additive decision routes/classify flag; no graph-service change (reuse `episode.ingest`); do not implement F041/F043/F044/F022/F033/F034/F050/F120; do not auto-wrap F035 Accept  
**Scale/Scope**: One repository per list; mutate one Decision per command

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for Decision aggregate, constructors, create/supersede commands, ADR projection query, REST/CLI, compile characterization, membership
- [x] Spec-driven: FR-001–FR-018 and SC-001–SC-009; defaults recorded in Assumptions
- [x] Architecture integrity: commands/queries in application; postgres implements Decision repo; no Graphiti types; Decision+lore+audit+outbox in one UoW; no distributed TX
- [x] Documentation: rest.md, FEATURE_DEVELOPMENT F040 → DONE when implemented, specify-rules, contributing/README as needed
- [x] Authority & provenance: human Decisions distinguishable from ADR projections; F032 trusted-source untouched; observational/F035 not auto-upgraded; popularity unused; `NewLoreEntry` not loosened
- [x] Temporal correctness: supersede new version, not in-place rewrite; predecessors gettable; drift (F050) not invented
- [x] Secure by default: write+membership to create/supersede; read+membership to get/list; local-mode actor; cross-tenant get-by-id 404
- [x] Observability: slog on create/get/list/supersede with repo, id, actor, outcome
- [x] Engineering intelligence: first-class decision model so agents can later answer “why Kafka?” — not generic memory
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: no 11th MCP tool; no web UI; no second knowledge plane; worker stays outbox-only; alternatives are fields not F042 product

**Post-design re-check**: Pass — dedicated `decisions` table for human-recorded rows (id = lore id); ADR lore projected at get/list; `NewDecisionLoreEntry` (verified `human_authored`, optional evidence) does not loosen `NewLoreEntry`; compile sets `FirstClassDecision` on RankedItem then ClassifyItem maps it to section `decisions`; ranking formulas unchanged.

## Project Structure

### Documentation (this feature)

```text
specs/040-decision-model/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/decisions.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/decision.go
internal/domain/decision_test.go
internal/domain/lore.go                          # NewDecisionLoreEntry
internal/application/commands/create_decision.go
internal/application/commands/supersede_decision.go
internal/application/queries/get_decision.go
internal/application/queries/list_decisions.go
internal/application/ports/decision.go
internal/application/ports/repositories.go       # UnitOfWork.Decisions()
internal/application/context/profile.go          # ClassifyItem FirstClassDecision
internal/application/queries/compile_context.go  # mark first-class decision items
internal/infrastructure/postgres/
internal/infrastructure/memory/
migrations/00009_decisions.sql
db/queries/decisions.sql
internal/adapters/http/decision.go
internal/adapters/cli/decision.go
cmd/memlore/main.go
docs/api/rest.md
docs/development/FEATURE_DEVELOPMENT.md
```

**Structure Decision**: MemLore Core is Go. Python `graph-service/` unchanged.
Reuse F110 supersede + F030 outbox pattern. Do not overload
`lore_review_decisions` or ingest cursor tables. Do not fold writes into
`memlore worker`.

## Complexity Tracking

> None — a dedicated Decision table plus ADR read-side projection are required
> by the spec (structured fields without duplicating trusted ADR facts).
> Materializing ADR rows on ingest would reopen F032. A second knowledge
> plane or 11th MCP tool would be speculative.

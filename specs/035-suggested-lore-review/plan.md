# Implementation Plan: Suggested Lore Review Queue (F035)

**Branch**: `035-suggested-lore-review` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Give humans an Accept / Edit / Reject path for git and PR observational lore
so automatically extracted knowledge cannot silently become canonical.
Pending items are **current observational lore** (`repository_observation` +
`commit`/`pr` evidence). A dedicated `lore_review_decisions` overlay stores
accept/reject by extract identity (scope + evidence type/value + statement
checksum). Accept **supersedes** the observation with a new current entry:
`human_verified` when the statement is unchanged, `human_authored` when the
reviewer edits it; evidence is copied; verification is `verified`. Reject
does not erase the observation. F032 accepted-ADR lore is never a queue item.
Surfaces: CLI `memlore review …` and REST `/v1/review-queue`. MCP stays at 10
tools. Compile ranking formulas stay frozen; add characterization tests only.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi, pgx, sqlc, goose, slog; stdlib hash for statement checksum (no new libraries)  
**Storage**: PostgreSQL — goose migration `00008_lore_review.sql` for `lore_review_decisions`; lore_entries reused  
**Testing**: `go test ./...`; `go vet ./...`; domain tests for accept/edit/reject/idempotency; handler tests with memory UoW; REST + membership contract tests; CLI contract tests; compile characterization  
**Target Platform**: `memlore review list|accept|reject`, `memlore serve`, existing `memlore worker` (outbox only)  
**Project Type**: CLI + REST governance service (MCP unchanged)  
**Performance Goals**: List pending by listing lore in one repository scope and joining review decisions; v1 in-process mutate  
**Constraints**: TDD; domain independent of HTTP/SQL; no MCP write/list tool; do not reopen F007/F020/F021/F030/F031/F032 except additive review routes; no graph-service change (reuse `episode.ingest`); do not implement F033/F034/F040/F022/F120; do not invent confidence/reason  
**Scale/Scope**: One repository per list; mutate one observational item per command

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for human_verified constructor, accept/reject domain, review commands/queries, REST/CLI, membership
- [x] Spec-driven: FR-001–FR-018 and SC-001–SC-010; defaults recorded in Assumptions
- [x] Architecture integrity: review command in application; postgres implements ReviewDecision repo; no Graphiti types; lore+audit+outbox in one UoW on Accept; no distributed TX
- [x] Documentation: rest.md, FEATURE_DEVELOPMENT F035 → DONE when implemented, specify-rules, contributing.md as needed
- [x] Authority & provenance: Accept writes `human_verified` or `human_authored` + verified + copied evidence; verify-in-place unchanged; git/PR stay observational until Accept; F032 ADR trusted-source untouched; popularity unused
- [x] Temporal correctness: observational predecessor superseded not overwritten; Reject does not delete history; conflicts not auto-resolved
- [x] Secure by default: write+membership to mutate; read+membership to list; local-mode actor header/flag; cross-tenant hard fail
- [x] Observability: slog on list/accept/reject with repo, item id, actor, outcome; review_decisions row is the durable decision
- [x] Engineering intelligence: promotion path from candidate to trusted knowledge, not a generic memory dump
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: no second knowledge plane; no 11th MCP tool; no web UI; no fake confidence; worker stays outbox-only

**Post-design re-check**: Pass — `NewHumanVerifiedLoreEntry` (do not loosen `NewLoreEntry`); Accept uses `ApplySupersessionWithSuccessor` with a pre-built successor and copied evidence; Reject is overlay-only; queue is a projection over current observational lore plus decisions.

## Project Structure

### Documentation (this feature)

```text
specs/035-suggested-lore-review/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/review-queue.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/lore.go                               # NewHumanVerifiedLoreEntry
internal/domain/review.go                             # extract identity, AcceptSuggestedLore, reject overlay
internal/application/commands/accept_review.go
internal/application/commands/reject_review.go
internal/application/queries/list_review_queue.go
internal/application/ports/review.go                  # ReviewDecisionRepository
internal/application/ports/repositories.go            # UnitOfWork.ReviewDecisions()
internal/infrastructure/postgres/                     # sqlc + mapping
internal/infrastructure/memory/                       # test doubles
migrations/00008_lore_review.sql
db/queries/review.sql
internal/adapters/http/review.go
internal/adapters/cli/review.go
cmd/memlore/main.go
docs/api/rest.md
docs/development/FEATURE_DEVELOPMENT.md
```

**Structure Decision**: MemLore Core is Go. Python `graph-service/` unchanged.
Reuse F110 supersede + F030 outbox pattern. Do not overload ingest cursor
tables. Do not fold review into `memlore worker`.

## Complexity Tracking

> None — a review-decision overlay plus a human_verified constructor are
> required by the spec. Projecting the queue from observational lore avoids a
> second knowledge plane. A generalized workflow engine would be speculative.

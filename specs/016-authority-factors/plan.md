# Implementation Plan: Authority factor model + evaluation (F003)

**Branch**: `016-authority-factors` | **Date**: 2026-08-31 | **Spec**: [spec.md](./spec.md)

## Summary

Replace F109’s hardcoded verified/unverified ranking bases with a pure
Authority Evaluation: explicit factors → deterministic score + trust band.
Surface factors and `trust_band` on compile/`get_for_task` items and on
`memlore.explain` (plus REST GET explain). Evaluate ephemerally at read time.
Do not persist scores. Do not break F112 filtering/conflicts or F111 local-mode
auth.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: chi, pgx/v5, sqlc, goose, MCP Go SDK (no new libs)  
**Storage**: PostgreSQL governance — **no schema change** (ephemeral evaluation)  
**Testing**: `go test ./...`; domain golden matrix; ranking regressions; compile/explain handlers; HTTP/MCP contracts  
**Target Platform**: MemLore core (`memlore serve`, `memlore mcp`)  
**Project Type**: hexagonal Go service  
**Performance Goals**: Evaluation is O(factors) per candidate; compile retrieval limit remains 20  
**Constraints**: TDD; no new MCP tools (stay at 9); no F110/F112 semantic changes; agent inference cannot reach canonical; no opaque-only score storage  
**Scale/Scope**: Pure evaluator + wire into ranking, compile, explain presenters, contracts, docs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED → GREEN → REFACTOR for evaluator, ranking, compile, explain, contracts
- [x] Spec-driven: clarifications encoded (ephemeral, bands, pipeline order)
- [x] Architecture integrity: domain evaluator has no HTTP/DB/Graphiti imports; adapters only map I/O
- [x] Documentation: authority-model, REST/MCP API, FEATURE_DEVELOPMENT, contracts, specify-rules
- [x] Authority & provenance: factors remain explainable; agent inference cannot silently gain canonical/high without verification
- [x] Temporal correctness: F112 filter/conflicts unchanged; stale still explainable
- [x] Secure by default: no auth changes; explain/compile remain reader operations
- [x] Observability: reuse existing compile/explain logs; no new telemetry required for v1
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: small pure evaluator; no evaluations table; no ML

## Project Structure

### Documentation (this feature)

```text
specs/016-authority-factors/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── authority-evaluation.md
└── tasks.md
```

### Source Code

```text
internal/domain/authority.go                 # TrustBand, SourceType, EvaluateAuthority
internal/domain/authority_test.go            # Golden band/score matrix
internal/application/authority/evaluate.go   # LoreEntry / GraphFact adapters
internal/application/context/ranking.go      # Call evaluator; map factors + trust_band
internal/application/queries/compile_context.go
internal/application/queries/explain_lore.go # Get + audits + evaluation
internal/adapters/presenters/context_packet.go
internal/adapters/presenters/lore.go         # ExplainResult authority fields
internal/adapters/http/handlers.go           # GET /v1/lore-entries/{id}/explain
internal/adapters/mcp/tools.go               # enrich memlore.explain
docs/architecture/authority-model.md
docs/architecture/target-architecture.md     # pipeline: filter then evaluate
docs/concepts/authority.md
docs/api/mcp.md
docs/api/rest.md
docs/development/FEATURE_DEVELOPMENT.md
.cursor/rules/specify-rules.mdc
```

**Structure Decision**: Domain owns pure evaluation from factor inputs.
Application `authority` maps governance/graph types. Ranking calls that
mapper. No new packages beyond `internal/application/authority`. No goose
migration.

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/authority-evaluation.md](./contracts/authority-evaluation.md)
- [quickstart.md](./quickstart.md)

Also update existing compile contract:
`specs/012-context-compiler/contracts/context-compile.md`.

## Phase 2 — Tasks

See [tasks.md](./tasks.md) (`/speckit-tasks`).

## Constitution re-check (post-design)

Gates still pass: pure evaluator; no schema; factors always accompany score;
agent origin hard-capped; F112 pipeline preserved with evaluate after filter;
REST/MCP parity on explain; no speculative reputation system.

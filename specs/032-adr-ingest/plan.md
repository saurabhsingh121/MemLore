# Implementation Plan: ADR Auto-Ingestion (F032)

**Branch**: `032-adr-ingest` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Ingest Architecture Decision Records from a **local working copy** bound to an
existing repository scope. Discover markdown under default dirs `docs/adr/`,
`adr/`, `architecture/decisions/` plus optional extra dirs. Conservatively
parse MADR/Nygard headings. Persist **accepted/adopted/approved** ADRs as
governed lore with origin `architecture_decision`, evidence type `adr`, and
**verified** status (constitution V trusted-source policy). Draft/proposed/
rejected are skipped. Deprecated/superseded are stored then invalidated so
history exists but they are not current canonical. Additive operational tables
hold ADR ingest runs, cursors, and processed path+checksum — not git SHA or PR
number tables. Operator surfaces: CLI `memlore ingest adr` / `ingest status
--kind adr` and REST trigger/status/candidates (`evidence_type=adr`). MCP
unchanged (10 tools). Compile ranking unchanged; F003 already boosts ADR
source type — add a characterization test that ingested accepted ADR lore
outranks git/PR observations.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi, pgx, sqlc, goose, slog; **stdlib** filesystem + heading parse (no ADR SDK)  
**Storage**: PostgreSQL — new goose migration `00007_adr_ingest.sql` for ADR ingest runs/cursors/processed files; lore_entries reused  
**Testing**: `go test ./...`; `go vet ./...`; unit tests for parser + trusted-source constructor; handler tests with fake ADRReader; REST/CLI contract tests; temp-dir adapter tests  
**Target Platform**: `memlore ingest adr`, `memlore ingest status --kind adr`, `memlore serve`, existing `memlore worker` (outbox only)  
**Project Type**: CLI + REST governance service (MCP unchanged)  
**Performance Goals**: Full scan of configured ADR dirs per run; idempotency via path+checksum; v1 in-process run  
**Constraints**: TDD; domain independent of OS paths; no MCP write tool; do not reopen F007/F020/F021/F030/F031 ranking or git/PR routes except additive ADR fields/routes; no graph-service change (reuse `episode.ingest`); do not implement F033/F035/F040  
**Scale/Scope**: One repository per run; conservative extraction (≤1 lore entry per ADR file)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for trusted-source constructor, ADR parser, ingest handler, REST/CLI, filesystem adapter
- [x] Spec-driven: FR-001–FR-021 and SC-001–SC-010; defaults recorded in Assumptions
- [x] Architecture integrity: ADRReader port; ingest command in application; postgres implements ADR ingest repos; no Graphiti types; lore+audit+outbox in one UoW per accepted item; no distributed TX
- [x] Documentation: rest.md, FEATURE_DEVELOPMENT F032 → DONE, specify-rules, contributing.md
- [x] Authority & provenance: origin `architecture_decision`, verified for accepted ADRs; evidence `adr`; CreateLore/remember stay human-authored; git/PR observations not upgraded; popularity unused
- [x] Temporal correctness: unchanged checksum is no-op; content change supersedes prior ingest-created lore (no overwrite); human-authored not auto-superseded; conflicts without a link not auto-resolved
- [x] Secure by default: write+membership to trigger; read+membership to list; local-mode actor header/flag
- [x] Observability: ADR ingest run records + slog start/complete/fail with repo, run id, counts
- [x] Engineering intelligence: extract *decisions* from ADRs, not dump every markdown file
- [x] Dependency policy: no new markdown/ADR SDK
- [x] Simplicity: no suggested_lore table; no GitHub Contents API; no 11th MCP tool; worker stays outbox-only; no F033/F035/F040

**Post-design re-check**: Pass — new `NewArchitectureDecisionLoreEntry` (do not loosen `NewLoreEntry` or reuse observational constructor); ADR ingest UoW port; stdlib filesystem adapter; REST+CLI only; supersession helper that accepts a pre-built successor (ApplySupersession today calls `NewLoreEntry`).

## Project Structure

### Documentation (this feature)

```text
specs/032-adr-ingest/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/adr-ingest.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/lore.go                               # NewArchitectureDecisionLoreEntry
internal/domain/lifecycle.go                          # supersede with pre-built successor
internal/domain/ingest.go                             # ADR snapshot, ADRIngestRun, ProcessedADR
internal/application/ingest/extract_adr.go            # skip + parse + status policy
internal/application/commands/ingest_adr.go           # IngestADRsHandler
internal/application/queries/ingest_status.go         # list ADR runs; evidence_type=adr
internal/application/ports/adr.go                     # ADRReader
internal/application/ports/ingest.go                  # ADRIngestRepository
internal/infrastructure/fsadr/reader.go               # local filesystem adapter
internal/infrastructure/postgres/                     # sqlc + mapping
internal/infrastructure/memory/                       # test doubles
migrations/00007_adr_ingest.sql
db/queries/adr_ingest.sql
internal/adapters/http/ingest_adr.go
internal/adapters/cli/ingest.go                       # ingest adr + status --kind adr
cmd/memlore/main.go
docs/api/rest.md
docs/development/FEATURE_DEVELOPMENT.md
```

**Structure Decision**: MemLore Core is Go. Python `graph-service/` unchanged.
Reuse F030/F031 write + outbox pattern; do not fold ADR scanning into
`memlore worker`. Do not overload git or PR ingest tables.

## Complexity Tracking

> None — additive ADR tables, a filesystem ADRReader, and a trusted-source
> constructor are required by the spec. A generalized ingest store or ADR SDK
> would be the speculative extra.

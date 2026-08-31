# Implementation Plan: Git Commit Ingestion (F030)

**Branch**: `030-git-commit-ingest` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Ingest commits from a **local git directory** bound to an existing repository
scope. Extract at most one **observational** lore candidate per SHA (heuristic
on commit message; skip noisy merges/chores). Persist as governed
`lore_entries` with origin `repository_observation`, verification `unverified`,
and evidence type `commit` (full SHA). Operational tables hold ingest runs,
per-repo cursor, and processed SHAs for idempotency — not a second knowledge
store. Operator surfaces: CLI `memlore ingest git` / `memlore ingest status`
and REST trigger/status/candidates. MCP unchanged (10 tools). Compile ranking
unchanged; F003 already downranks repo observations vs verified architecture.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi, pgx, sqlc, goose, slog; **git CLI** via `os/exec` (no go-git)  
**Storage**: PostgreSQL — new goose migration for ingest runs/cursors/processed SHAs; lore_entries reused  
**Testing**: `go test ./...`; `go vet ./...`; unit tests for extract + observational create; handler tests with fake GitReader; REST contract tests; CLI parse/format tests; temp-dir git adapter test  
**Target Platform**: `memlore ingest git`, `memlore ingest status`, `memlore serve`, existing `memlore worker` (outbox only)  
**Project Type**: CLI + REST governance service (MCP unchanged)  
**Performance Goals**: Incremental by cursor; optional `--max-commits`; v1 in-process run (local git)  
**Constraints**: TDD; domain independent of git/HTTP; no MCP write tool; do not reopen F007/F020/F021 ranking; no graph-service change unless episode contract is proven insufficient (reuse `episode.ingest`)  
**Scale/Scope**: One repository per run; conservative extraction (≤1 candidate/SHA)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for evidence type, observational lore, extractor, ingest handler, REST/CLI
- [x] Spec-driven: FR-001–FR-018 and SC-001–SC-008; defaults recorded in Assumptions
- [x] Architecture integrity: GitReader port; ingest command in application; postgres implements ingest repos; no Graphiti types; lore+audit+outbox in one UoW per candidate; no distributed TX
- [x] Documentation: rest.md, FEATURE_DEVELOPMENT F030 → DONE, specify-rules, contributing.md
- [x] Authority & provenance: origin `repository_observation`, unverified; CreateLore/remember stay human-authored; no auto-verify; no trusted-source policy
- [x] Temporal correctness: re-ingest does not overwrite existing candidates; SHA already processed is left as-is
- [x] Secure by default: write+membership to trigger; read+membership to list; local-mode actor header/flag
- [x] Observability: ingest run records + slog start/complete/fail with repo, run id, counts
- [x] Engineering intelligence: extract *why* from commits, not dump every message
- [x] Dependency policy: no new third-party libraries (git CLI, not go-git)
- [x] Simplicity: no suggested_lore knowledge table; no GitHub adapter; no 11th MCP tool; worker stays outbox-only

**Post-design re-check**: Pass — observational constructor (do not loosen `NewLoreEntry`), ingest UoW port, git CLI adapter, REST+CLI only.

## Project Structure

### Documentation (this feature)

```text
specs/030-git-commit-ingest/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/git-commit-ingest.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/enums.go                         # EvidenceTypeCommit
internal/domain/lore.go                          # NewObservationalLoreEntry
internal/application/ingest/extract.go           # skip + candidate heuristic
internal/application/commands/ingest_git.go      # IngestGitCommitsHandler
internal/application/queries/ingest_status.go    # list runs / candidates
internal/application/ports/git.go                # GitReader
internal/application/ports/ingest.go             # ingest run/cursor/sha repos
internal/infrastructure/gitcli/reader.go         # git log adapter
internal/infrastructure/postgres/                # sqlc + mapping
internal/infrastructure/memory/                  # test doubles
migrations/00005_git_ingest.sql
db/queries/ingest.sql
internal/adapters/http/{handlers,dto}.go         # POST/GET ingest
internal/adapters/cli/ingest.go
cmd/memlore/main.go                              # ingest subcommands
docs/api/rest.md
docs/development/FEATURE_DEVELOPMENT.md
docs/development/contributing.md
```

**Structure Decision**: Go core only. `graph-service/` unchanged — existing
`episode.ingest` payload (statement + scope + episode_id) is sufficient.

## Complexity Tracking

> No constitution violations requiring justification.

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

See [data-model.md](./data-model.md), [contracts/git-commit-ingest.md](./contracts/git-commit-ingest.md),
[quickstart.md](./quickstart.md).

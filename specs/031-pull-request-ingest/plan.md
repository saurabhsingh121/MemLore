# Implementation Plan: Pull Request Ingestion (F031)

**Branch**: `031-pull-request-ingest` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Ingest **merged** pull requests from **GitHub** for an existing repository
scope (`github.com/{owner}/{repo}` → `{owner}/{repo}`). Extract at most one
**observational** lore candidate per PR (heuristic on title, body, and human
review comments; skip unmerged, bots, and noisy chore/dependabot bumps).
Persist as governed `lore_entries` with origin `repository_observation`,
verification `unverified`, and evidence type `pr` (`{owner}/{repo}#{number}`).
Additive operational tables hold PR ingest runs, per-repo cursor, and processed
PR numbers — not a second knowledge store, and not `git_ingest_shas`.
Operator surfaces: CLI `memlore ingest pr` / `ingest status --kind pr` and
REST trigger/status/candidates. MCP unchanged (10 tools). Compile ranking
unchanged; F003 already downranks repo observations vs verified architecture.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi, pgx, sqlc, goose, slog; **stdlib `net/http`** for GitHub REST (no GitHub SDK)  
**Storage**: PostgreSQL — new goose migration for PR ingest runs/cursors/processed PRs; lore_entries reused  
**Testing**: `go test ./...`; `go vet ./...`; unit tests for extract + observational PR evidence; handler tests with fake PullRequestReader; REST contract tests; CLI parse/format tests; httptest GitHub adapter tests  
**Target Platform**: `memlore ingest pr`, `memlore ingest status --kind pr`, `memlore serve`, existing `memlore worker` (outbox only)  
**Project Type**: CLI + REST governance service (MCP unchanged)  
**Performance Goals**: Incremental by merged_at cursor; optional `--max-prs` / `--pr`; v1 in-process run  
**Constraints**: TDD; domain independent of GitHub/HTTP; no MCP write tool; do not reopen F007/F020/F021/F030 ranking or git routes; no graph-service change (reuse `episode.ingest`); never log GitHub tokens  
**Scale/Scope**: One repository per run; conservative extraction (≤1 candidate/PR)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for evidence type `pr`, observational lore, PR extractor, ingest handler, REST/CLI, GitHub adapter
- [x] Spec-driven: FR-001–FR-021 and SC-001–SC-008; defaults recorded in Assumptions
- [x] Architecture integrity: PullRequestReader port; ingest command in application; postgres implements PR ingest repos; no Graphiti types; lore+audit+outbox in one UoW per candidate; no distributed TX
- [x] Documentation: rest.md, FEATURE_DEVELOPMENT F031 → DONE, specify-rules, contributing.md, README
- [x] Authority & provenance: origin `repository_observation`, unverified; CreateLore/remember stay human-authored; no auto-verify; no trusted-source policy
- [x] Temporal correctness: re-ingest does not overwrite existing candidates; processed PR is left as-is
- [x] Secure by default: write+membership to trigger; read+membership to list; local-mode actor header/flag; token never logged
- [x] Observability: PR ingest run records + slog start/complete/fail with repo, run id, counts (no token)
- [x] Engineering intelligence: extract *why* from PRs, not dump every description
- [x] Dependency policy: no new third-party GitHub SDK (stdlib HTTP)
- [x] Simplicity: no suggested_lore table; no forge abstraction; no 11th MCP tool; worker stays outbox-only; no F054/F074

**Post-design re-check**: Pass — extend observational constructor (do not loosen `NewLoreEntry`), PR ingest UoW port, GitHub HTTP adapter, REST+CLI only.

## Project Structure

### Documentation (this feature)

```text
specs/031-pull-request-ingest/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/pull-request-ingest.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/domain/enums.go                              # EvidenceTypePR
internal/domain/lore.go                               # NewObservationalLoreEntry accepts pr evidence
internal/domain/ingest.go                             # PR snapshot, PRIngestRun, ProcessedPR
internal/application/ingest/extract_pr.go             # PR skip + candidate heuristic
internal/application/commands/ingest_pr.go            # IngestPullRequestsHandler
internal/application/queries/ingest_pr_status.go      # list PR runs / filter candidates
internal/application/ports/github.go                  # PullRequestReader
internal/application/ports/ingest.go                  # PR ingest repos (or sibling file)
internal/infrastructure/githubhttp/reader.go          # GitHub REST adapter
internal/infrastructure/postgres/                     # sqlc + mapping
internal/infrastructure/memory/                       # test doubles
migrations/00006_pr_ingest.sql
db/queries/pr_ingest.sql
internal/adapters/http/ingest_pr.go
internal/adapters/cli/ingest.go                       # ingest pr + status --kind
cmd/memlore/main.go
docs/api/rest.md
docs/development/FEATURE_DEVELOPMENT.md
```

**Structure Decision**: MemLore Core is Go. Python `graph-service/` unchanged.
Reuse F030 observational write + outbox; do not fold PR polling into `memlore worker`.

## Complexity Tracking

> None — additive PR tables and a GitHub HTTP port are required by the spec; a
> generalized ingest store would be the speculative extra.

# Research: Repository Intelligence Profile (F020)

## Decision 1 — Compile then classify (no new store)

**Decision**: Reuse `SearchKnowledgeHandler` + compile pipeline
(`FilterCurrent` → `DetectConflicts` → `RankAndDedup` → `ApplyTokenBudget`),
then assign each budgeted item to at most one named section.

**Rationale**: F007 already ranks, budgets, and filters stale. A persisted
profile would drift from lore and violate "compile on read."

**Alternatives rejected**: (a) New `repository_profiles` table — stale copies.
(b) LLM summarization — invents text, violates constitution IX.

## Decision 2 — Default overview query

**Decision**: Profile search uses a fixed overview query covering section
themes (`architecture decisions conventions technologies ownership gotchas
migrations risks dependencies`) plus the repository key as `task` for compile
reuse. No caller-supplied query in this slice.

**Rationale**: Profile is repository-wide, not task-specific (that's F021).
A stable query keeps tests deterministic.

## Decision 3 — Deterministic first-match classification

**Decision**: Cue lists and priority order live in application/context
(pure function). First matching rule wins. Unmatched items increment
`unclassified_count` and are omitted from sections.

**Rationale**: Conservative classification beats a generic "other" dump
(constitution IX). ADR evidence type and `architecture_decision` origin map
to `decisions` without keyword luck.

**Alternatives rejected**: ML classifier; forcing unmatched into architecture.

## Decision 4 — Surfaces

**Decision**: REST `POST /v1/repository-profile` (body like compile);
MCP `memlore.repo_profile`; CLI `memlore profile --repository` using the same
Postgres/graph wiring as `memlore mcp`.

**Rationale**: Matches existing POST compile/search style. CLI is a
constitution-required developer surface for F020; talking to local DSN avoids
requiring `serve` to be up.

## Decision 5 — MCP tool count

**Decision**: 9 → 10 tools. Update contract tests that hard-code nine.

**Rationale**: New primary agent read path; not an enrichment of
`get_for_task` (that's F021).

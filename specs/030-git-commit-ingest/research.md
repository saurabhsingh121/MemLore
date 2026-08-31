# Research: Git Commit Ingestion (F030)

## Decision 1 — Persist candidates as lore_entries, not suggested_lore

**Decision**: Store extracted candidates as governed lore with
`origin=repository_observation` and `verification_status=unverified`. Add
operational tables for ingest runs, cursor, and processed SHAs only.

**Rationale**: Constitution V requires git observations to remain
distinguishable; F003 already downranks `repository_observation`. Compile /
`get_for_task` already read lore_entries. A second knowledge table would be
invisible to F021 unless compile were changed (out of scope). F035 can later
list unverified observational lore without a data migration.

**Alternatives rejected**: `suggested_lore` table as the knowledge store;
status enum `candidate` on lore (would fork verification); stuffing extracts
only into ingest-run JSON (compile could not see them).

## Decision 2 — Local git CLI, not GitHub API or go-git

**Decision**: Operator supplies a local working copy (`--path` / REST `path`).
Read history with the `git` binary (`git -C <dir> log …`). Define a
`GitReader` port so unit tests fake commits.

**Rationale**: Spec v1 is local-clone. `os/exec` + stdlib keeps the dependency
surface empty and is easy to test with `git init` in `t.TempDir()`. go-git
adds a large module for no v1 capability. GitHub API is Epic H / F031.

**Alternatives rejected**: go-git; clone-from-URL; GitHub commits API.

## Decision 3 — New evidence type `commit`

**Decision**: Add `EvidenceTypeCommit = "commit"` with value = full SHA
(40 hex). Optional extra `path` evidence for changed files. Do not encode SHA
as `url`.

**Rationale**: Existing types cannot represent a SHA honestly. `url` would
force a forge URL MemLore may not have. `path` is for files. Additive parse
path keeps old evidence valid.

**Alternatives rejected**: SHA in `url`; SHA only on ingest_sha row without
lore evidence (compile/explain would lose the link).

## Decision 4 — Separate observational constructor; keep CreateLore human-only

**Decision**: Do not loosen `NewLoreEntry` (still requires `human_authored`).
Add `NewObservationalLoreEntry` that requires origin
`repository_observation`, unverified, and at least one `commit` evidence.
`CreateLoreCommand` / REST create / MCP `remember` stay human-authored.
Ingest uses a new `IngestGitCommits` command that writes observational lore +
audit + `episode.ingest` outbox per accepted candidate.

**Rationale**: Silently reusing CreateLore would make git extracts look like
human ADRs (constitution V). Outbox reuse keeps graph-service in sync without
a new episode shape.

**Alternatives rejected**: Optional origin on CreateLore; MCP remember for
git; new graph-service endpoint.

## Decision 5 — Dedicated ingest command, not memlore worker git polling

**Decision**: `memlore ingest git` runs extraction in-process and records an
`IngestRun`. `memlore worker` remains outbox → graph-service only.

**Rationale**: Constitution VIII wants ingest observable; a one-shot operator
command is the smallest vertical. Mixing git polling into the outbox worker
couples unrelated cadences. REST POST is the same in-process handler.

**Alternatives rejected**: Worker-loop git poll; async job queue (Redis) for v1.

## Decision 6 — REST paths without repository key in the URL

**Decision**:

- `POST /v1/ingest/git` body `{ scope, path, max_commits? }`
- `GET /v1/ingest/runs?scope_kind=&scope_key=`
- `GET /v1/ingest/runs/{id}`
- `GET /v1/ingest/candidates?scope_kind=&scope_key=`

**Rationale**: Repository keys are `github.com/org/repo` (slashes). Existing
list-by-scope already uses query params. `{key}` path segments would break.

**Alternatives rejected**: `/v1/repositories/{key}/ingest/...`.

## Decision 7 — Conservative deterministic extractor

**Decision**: Pure function on snapshot (message + parents + paths). Skip
wins over cues. At most one candidate per SHA: trimmed subject + body as the
statement when a rationale cue is present and skip rules do not match. No LLM.

**Rationale**: Spec: cannot invent decisions; volume loses to conservatism;
must be testable.

**Alternatives rejected**: LLM rewrite; one lore per changed file; dumping
every message.

## Decision 8 — Per-SHA transactions + processed-SHA unique key

**Decision**: Insert a `running` ingest run (partial unique index: one active
run per repository). Process SHAs oldest-first; each accepted/skipped SHA is
its own UoW (lore+audit+outbox+processed row, or processed-only if skipped).
Unique `(scope, sha)` prevents duplicates on retry. Cursor updated to last
successfully processed commit.

**Rationale**: Spec: failed retry must not duplicate accepted lore. A single
transaction for the whole run would roll back successful SHAs on later git
errors.

**Alternatives rejected**: One giant transaction; deterministic UUIDs only
(would not record skipped SHAs).

## Decision 9 — No MCP tool; no compile contract change

**Decision**: Status is CLI + REST. MCP stays at 10 tools. F007/F021 ranking
untouched; US4 is a characterization test that existing authority already
orders verified architecture above unverified `repository_observation`.

**Rationale**: User: do not add an 11th tool; do not reopen compile.

**Alternatives rejected**: `memlore.ingest_status` tool; compile origin filter
changes.

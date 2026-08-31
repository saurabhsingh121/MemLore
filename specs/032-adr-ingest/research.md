# Research: ADR Auto-Ingestion (F032)

## Decision 1 — Trusted-source constructor; do not loosen NewLoreEntry

**Decision**: Add `NewArchitectureDecisionLoreEntry`. It requires origin
`architecture_decision` (default if empty), at least one `adr` evidence, and
sets `verification_status=verified` with `VerifiedBy`/`VerifiedAt` = triggering
actor / now. `NewLoreEntry` remains human-authored only.
`NewObservationalLoreEntry` still requires `commit` or `pr` evidence and MUST
NOT accept `adr` as a substitute.

**Rationale**: Constitution V allows an explicit trusted-source policy for
accepted ADRs. Silently reusing `CreateLore` would make ingest look like a
human write. Reusing the observational constructor would make ADRs look like
git/PR capture (the opposite of F032). Auto-verify is limited to this
constructor and the accepted-status policy.

**Alternatives rejected**: Loosen `NewLoreEntry` to any origin; origin
`imported_source` + unverified (fallback the user allowed only if auto-verify
were too strong — it is not, for accepted ADRs); stuffing ADRs through
`NewObservationalLoreEntry`.

## Decision 2 — Local filesystem; no GitHub Contents API

**Decision**: `ADRReader` lists markdown under configured directories of an
operator-supplied local path (F030-style `--path`). Production adapter uses
stdlib `os`/`filepath`. Unit tests use a fake; adapter tests use `t.TempDir()`
fixtures.

**Rationale**: User default: smallest vertical, testable without forge
credentials. GitHub Contents would duplicate F031’s token/auth surface for a
directory walk F030 already solved locally.

**Alternatives rejected**: GitHub Contents API; clone-from-URL; recursive
generic doc crawl (F033).

## Decision 3 — Additive ADR tables, do not overload git/PR identity

**Decision**: Goose migration `00007_adr_ingest.sql` with `adr_ingest_runs`,
`adr_ingest_cursors`, `adr_ingest_files`. Unique processed key is
`(scope_kind, scope_key, relative_path, checksum)`.

**Rationale**: Git identity is SHA; PR identity is number. ADR identity is
path + content. Overloading would mix concurrency and make checksum-change
vs same-SHA retries ambiguous. Independent tables let git, PR, and ADR ingest
run concurrently for the same repository.

**Alternatives rejected**: Polymorphic ingest store; storing paths in
`git_ingest_shas.sha`; unique on path only (cannot distinguish unchanged vs
changed content).

## Decision 4 — Idempotency: path + checksum; change supersedes

**Decision**: Full scan of configured dirs each run. If `(path, checksum)`
already exists → no-op. If the same relative path exists with a **different**
checksum and prior ingest-created lore → create a new architecture-decision
entry and supersede the previous ingest-created lore (never overwrite the
row). Cursor records last successful scan metadata for status, not as the skip
mechanism.

**Rationale**: ADR dirs are small. Full scan + unique (path, checksum) is
simpler than a git-log watermark and matches “unchanged checksum → no-op;
changed accepted ADR → new version + supersede.”

**Alternatives rejected**: Skip-if-unchanged without a new lore version on
change (would hide updates); silent overwrite of the lore row (constitution
VI); cursor-only incremental (would miss in-place edits of older files).

## Decision 5 — Status policy

**Decision**:

| Status tokens (case-insensitive) | Action |
|----------------------------------|--------|
| accepted, adopted, approved | Trusted-source: verified `architecture_decision` |
| draft, proposed, rejected, withdrawn | Skip (no lore) |
| deprecated, superseded, superceded | Store as `architecture_decision` + `adr` evidence, then **invalidate** in the same UoW (history, not current) |
| missing / unknown | Skip if no Decision either; skip unknown tokens (do not guess) |

Front matter `status:` is accepted as a stated status (MADR).

**Rationale**: User defaults. Invalidation (not unverified-current) keeps
deprecated files out of `IsCurrent` / compile current set so F003 ADR source
strength cannot make them look canonical.

**Alternatives rejected**: Skip deprecated (loses history); keep deprecated
verified-and-current; unverified-but-current (F003 still boosts ADR evidence).

## Decision 6 — Conservative parser (stdlib); one lore per file

**Decision**: Parse common MADR/Nygard headings (Status, Context, Decision,
Consequences, Alternatives, Supersedes) plus YAML-ish front matter `status:`.
Statement = title + Decision, with Context / Alternatives / Consequences
appended when they fit `MaxStatementLength`; skip if Decision cannot be
represented faithfully. Skip README, `template.md`, `NNNN-title.md`. No LLM.
No ADR SDK.

**Rationale**: Spec forbids inventing decisions and splitting into F040
aggregates. Headings are enough for v1.

**Alternatives rejected**: markdown AST library; one lore row per section;
reuse commit/PR extractors unchanged.

## Decision 7 — Supersession: ingest-created only; pre-built successor

**Decision**: When an accepted ADR states “Supersedes ADR-0003” (or equivalent)
and current ingest-created lore exists with matching `adr` evidence, map onto
lore supersession. Discriminator: predecessor origin MUST be
`architecture_decision` (human create cannot produce that origin today). Do
**not** call `ApplySupersession` as-is — it uses `NewLoreEntry` (human). Add
`ApplySupersessionWithSuccessor(predecessor, successor, actor, now)` that
marks the predecessor and writes both audits.

Content-change of the same ADR identity uses the same helper (new successor
supersedes prior ingest lore for that path).

**Rationale**: Constitution VI + user default: only chain ingest-created ADR
lore; leave human ADRs for humans/F035/F040.

**Alternatives rejected**: Auto-supersede any lore with matching `adr`
evidence; reuse ApplySupersession (would create human_authored successors).

## Decision 8 — REST without repository key in the URL

**Decision**:

- `POST /v1/ingest/adr` body `{ scope, path, adr_dirs? }`
- `GET /v1/ingest/adr-runs?scope_kind=&scope_key=`
- `GET /v1/ingest/adr-runs/{id}`
- `GET /v1/ingest/candidates?scope_kind=&scope_key=&evidence_type=adr`

Existing git/PR routes unchanged. Default candidates list (no `evidence_type`)
still lists `repository_observation` only. `evidence_type=adr` lists current
`architecture_decision` lore that has `adr` evidence (ingest-derived).

**Rationale**: Same slash-in-key constraint as F030/F031. Separate ADR run
collection avoids mixing git/PR clients. Additive `evidence_type=adr` is the
smallest list surface.

**Alternatives rejected**: `/v1/repositories/{key}/ingest/adr`; listing ADR
lore from `GET /v1/lore-entries` only (operators still need ingest-run
status).

## Decision 9 — Dedicated ingest command; worker stays outbox-only

**Decision**: `memlore ingest adr --repository --path [--adr-dir …]` runs
extraction in-process. `memlore worker` remains outbox → graph-service only.

**Rationale**: Same as F030/F031. Mixing ADR filesystem walks into the outbox
worker couples unrelated cadences.

**Alternatives rejected**: Worker-loop ADR poll; async job queue for v1.

## Decision 10 — No MCP tool; no compile ranking change

**Decision**: Status is CLI + REST. MCP stays at 10 tools. F007/F021 ranking
untouched; US5 is a characterization test that existing F003 authority already
orders verified `architecture_decision` + `adr` evidence above unverified
`repository_observation` (git or PR).

**Rationale**: User: do not add an 11th tool; do not reopen compile formulas.

**Alternatives rejected**: `memlore.ingest_status` tool; compile origin filter
changes; auto-verify git/PR lore.

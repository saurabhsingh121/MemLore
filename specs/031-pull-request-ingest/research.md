# Research: Pull Request Ingestion (F031)

## Decision 1 — Same observational origin; new evidence type `pr`

**Decision**: PR extracts use origin `repository_observation` and
`verification_status=unverified`, same as F030. Add evidence type `pr` with
value `{owner}/{repo}#{number}` (e.g. `acme/payments#1842`). Optional `url`
for review comments actually used and for linked issues; optional `path` for
files.

**Rationale**: Constitution V requires PR extracts to stay distinguishable from
human-authored knowledge. F003 already downranks `repository_observation`. A
new origin would require compile/authority changes (out of scope). `url` alone
cannot be the PR identity because compile/explain should see a first-class PR
ref; HTML URLs are still useful for comments and tickets.

**Alternatives rejected**: New origin `pr_observation`; encoding the PR only as
`url`; stuffing PR identity only on the ingest row (compile would lose the link).

## Decision 2 — Additive PR tables, do not overload git_ingest_shas

**Decision**: New goose migration with `pr_ingest_runs`, `pr_ingest_cursors`,
`pr_ingest_prs`. Unique processed key is `(scope_kind, scope_key, pr_number)`.
Cursor is last `merged_at` + PR number.

**Rationale**: `git_ingest_shas` is SHA-shaped (`local_path`, `cursor_sha`).
PR identity is owner/repo + number (or node id). Overloading would mix
concurrency (one active git run vs one active PR run) and make retries
ambiguous. Independent tables let git and PR ingest run concurrently for the
same repository.

**Alternatives rejected**: Generalize F030 tables into a polymorphic ingest
store (speculative; F032 ADR can add its own tables); store PR ids in
`git_ingest_shas.sha`.

## Decision 3 — GitHub REST via stdlib HTTP; PullRequestReader port

**Decision**: Implement a `PullRequestReader` port. Production adapter uses
GitHub REST (`net/http`) against `https://api.github.com` (overridable base URL
for tests). Token from `MEMLORE_GITHUB_TOKEN`, then `GITHUB_TOKEN`. Never log
the token. Unit tests use a fake; adapter tests use `httptest` fixtures.

**Rationale**: F031 needs GitHub as the PR source (Epic H). stdlib HTTP avoids
a third-party SDK. `gh` CLI is harder to fake without network. GraphQL can wait;
REST list/get/files/comments is enough for v1.

**Alternatives rejected**: go-github SDK; `gh` CLI; GitHub App webhooks (F054);
generic forge abstraction.

## Decision 4 — Merged PRs only

**Decision**: Skip open, draft, and closed-unmerged PRs. Do not persist them as
lore labeled “not-landed”.

**Rationale**: User default for the public contract. Treating an unmerged PR as
landed implementation would lie to compile. Labeling unmerged as observational
“intent” is F035/F040-adjacent and would change ranking semantics.

**Alternatives rejected**: Ingest unmerged with a not-landed flag; ingest
closed-unmerged as rejected approaches.

## Decision 5 — Extend NewObservationalLoreEntry; keep CreateLore human-only

**Decision**: Do not loosen `NewLoreEntry`. Extend
`NewObservationalLoreEntry` to require at least one `commit` **or** `pr`
evidence (F030 still supplies commit). Ingest uses `IngestPullRequests`
command writing observational lore + audit + `episode.ingest` outbox per
accepted candidate.

**Rationale**: Same as F030: silently reusing CreateLore would make PR extracts
look like human ADRs. Outbox reuse keeps graph-service in sync.

**Alternatives rejected**: Third constructor; optional origin on CreateLore;
MCP remember for PRs; new graph-service endpoint.

## Decision 6 — Dedicated ingest command; worker stays outbox-only

**Decision**: `memlore ingest pr` runs extraction in-process and records a
`PRIngestRun`. `memlore worker` remains outbox → graph-service only.

**Rationale**: Same as F030. Mixing GitHub polling into the outbox worker
couples unrelated cadences and needs long-lived tokens in the worker process.

**Alternatives rejected**: Worker-loop PR poll; async job queue for v1.

## Decision 7 — REST without repository key in the URL

**Decision**:

- `POST /v1/ingest/pr` body `{ scope, pr?, max_prs? }`
- `GET /v1/ingest/pr-runs?scope_kind=&scope_key=`
- `GET /v1/ingest/pr-runs/{id}`
- `GET /v1/ingest/candidates?scope_kind=&scope_key=&evidence_type=pr`

Existing git routes unchanged. `GET /v1/ingest/candidates` without
`evidence_type` still lists all `repository_observation` lore (git + PR).

**Rationale**: Repository keys contain slashes. Separate PR run collection
avoids changing F030 git list JSON. Additive `evidence_type` filter is the
smallest way to list PR-derived candidates.

**Alternatives rejected**: `/v1/repositories/{key}/ingest/pr`; unioning PR runs
into `GET /v1/ingest/runs` as default (would mix F030 clients).

## Decision 8 — Conservative PR extractor (not the commit function)

**Decision**: New `ExtractPRCandidate` on the PR snapshot. Skip unmerged and
bots first. Noisy conventional titles skip unless body or a human review
comment has a rationale cue. Statement is trimmed title+body when that text
has a cue; otherwise the human review comment(s) that have cues (still one
candidate). No LLM. At most one candidate per PR.

**Rationale**: Spec forbids stuffing PR bodies through the commit-message
extractor unchanged. Review discussion is often the real decision record.

**Alternatives rejected**: Reuse `ExtractCandidate(commit)` by concatenating
the PR into a fake commit; one lore per review comment; LLM rewrite.

## Decision 9 — Per-PR transactions + processed-PR unique key

**Decision**: Insert a `running` PR ingest run (partial unique index: one
active PR run per repository). Process merged PRs oldest-first; each
accepted/skipped PR is its own UoW. Unique `(scope, pr_number)` prevents
duplicates on retry. Cursor updated to last successfully processed merged PR.

**Rationale**: Same retry safety as F030. A single transaction for the whole
run would roll back successful PRs on later GitHub errors.

**Alternatives rejected**: One giant transaction; GraphQL node id as the only
unique key (number is what operators use with `--pr`).

## Decision 10 — No MCP tool; no compile ranking change

**Decision**: Status is CLI + REST. MCP stays at 10 tools. F007/F021 ranking
untouched; US4 is a characterization test that existing authority already
orders verified architecture above unverified `repository_observation`
(PR evidence instead of commit).

**Rationale**: User: do not add an 11th tool; do not reopen compile.

**Alternatives rejected**: `memlore.ingest_status` tool; compile origin filter
changes.

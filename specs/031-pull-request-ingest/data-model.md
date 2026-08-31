# Data Model: Pull Request Ingestion (F031)

## PullRequestSnapshot (not persisted as a row)

| Field | Notes |
|-------|-------|
| Number | GitHub PR number |
| NodeID | Optional GraphQL node id (stored on processed row when present) |
| Owner / Repo | GitHub owner and repository name |
| Title | PR title |
| Body | PR description (markdown) |
| AuthorLogin | GitHub login |
| AuthorIsBot | true when GitHub user type is Bot or login matches skip-list bots |
| Merged | true only when merged |
| MergedAt | Merge timestamp (UTC); nil if unmerged |
| HTMLURL | PR HTML URL |
| Files | Changed paths (name-only) |
| ReviewComments | Review (diff) comments: HTMLURL, Body, AuthorLogin, AuthorIsBot |
| IssueComments | Conversation comments on the PR issue: same shape |
| LinkedIssueURLs | Issue/ticket URLs parsed from body (`Fixes`/`Closes`/`Resolves` #N plus existing URLs) |

## PRCandidate → LoreEntry

Reuses `lore_entries`. Created via `NewObservationalLoreEntry`:

| Field | Value |
|-------|-------|
| statement | Faithful extract of title+body and/or used review comments |
| scope | `{ kind: repository, key }` |
| origin | `repository_observation` |
| verification_status | `unverified` |
| evidence | `[{ type: "pr", value: "owner/repo#N" }, ... optional url/path]` |
| created_by | Triggering actor |
| verified_* | null |

`NewObservationalLoreEntry` requires at least one `commit` **or** `pr`
evidence. Human `NewLoreEntry` is unchanged.

## EvidenceType (additive)

| Type | Value |
|------|-------|
| `url` | existing (review comment HTML URL; linked issue URL) |
| `path` | existing (changed file path; cap 8 like git ingest) |
| `adr` | existing |
| `commit` | existing (F030) |
| `pr` | **new** — `{owner}/{repo}#{number}` |

## PRIngestRun (`pr_ingest_runs`)

| Field | Notes |
|-------|-------|
| id | UUID |
| scope_kind / scope_key | Repository scope |
| actor_id | Triggering actor |
| pr_number | Optional single-PR filter (0/null = incremental list) |
| status | `running` \| `succeeded` \| `failed` |
| prs_seen | PRs read from GitHub for this run |
| prs_skipped | Processed but no candidate |
| candidates_stored | New lore rows this run |
| cursor_pr | Last processed PR number (nullable) |
| cursor_at | That PR's merged_at (nullable) |
| error_summary | Set when failed (never include the token) |
| started_at / finished_at | Observability |

**Invariant**: At most one row with `status=running` per `(scope_kind, scope_key)`.

## PRIngestCursor (`pr_ingest_cursors`)

| Field | Notes |
|-------|-------|
| scope_kind / scope_key | PK |
| last_pr | Last successfully processed PR number |
| last_merged_at | That PR's merged timestamp |
| updated_at | |

Used to bound the next GitHub list (`merged_at` after this watermark).
Idempotency still depends on ProcessedPR.

## ProcessedPR (`pr_ingest_prs`)

| Field | Notes |
|-------|-------|
| scope_kind / scope_key / pr_number | PK |
| node_id | Optional |
| lore_entry_id | Set when a candidate was stored; null if skipped |
| skipped | true when no candidate |
| skip_reason | `unmerged`, `bot`, `noisy`, `empty`, `no_rationale`, `too_long`, … |
| processed_at | |

Re-insert of the same PK is a conflict/no-op (idempotent).

**Why not `git_ingest_shas`**: that table's identity is a git SHA and its
sibling cursor is `cursor_sha`. PR numbers are a different unique key and
must not collide with or block git ingest concurrency.

## State transitions

```text
PRIngestRun:  (create) → running → succeeded
                           ↘ failed
ProcessedPR: absent → {skipped | stored}  (never rewritten)
LoreEntry from ingest: unverified observational → (F035/verify later; not this slice)
```

## Validation

- Scope kind MUST be `repository`.
- Scope key MUST match `github.com/{owner}/{repo}` (exactly two path segments).
- PR evidence MUST be non-empty `{owner}/{repo}#{positive integer}`.
- Statement length ≤ `MaxStatementLength`; otherwise skip with `too_long`.
- `max_prs` omitted or ≤0 means no cap (still bounded by unread merged PRs after cursor).
- Concurrent second `running` PR ingest run for the same scope → conflict (HTTP 409).
- Git ingest `running` for the same scope is independent.

## Relationships

- PRIngestRun belongs to a repository scope.
- ProcessedPR optionally references lore_entries.id.
- PR cursor is 1:1 with repository scope.
- Compile continues to read lore_entries only.

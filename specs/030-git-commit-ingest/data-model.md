# Data Model: Git Commit Ingestion (F030)

## GitCommitSnapshot (not persisted as a row)

| Field | Notes |
|-------|-------|
| SHA | Full hex object id |
| Author | Commit author name (and email if git provides; store name in v1) |
| CommittedAt | Commit timestamp (UTC) |
| Subject | First line of message |
| Body | Remainder of message |
| Message | Full message (subject + body) |
| ParentCount | Used to detect merges (`> 1`) |
| Paths | Changed paths (name-only) |

## CommitCandidate → LoreEntry

Reuses `lore_entries`. Created via `NewObservationalLoreEntry`:

| Field | Value |
|-------|-------|
| statement | Faithful extract of commit message (subject + body, trimmed) |
| scope | `{ kind: repository, key }` |
| origin | `repository_observation` |
| verification_status | `unverified` |
| evidence | `[{ type: "commit", value: "<full sha>" }, ... optional path refs]` |
| created_by | Triggering actor |
| verified_* | null |

Human `NewLoreEntry` is unchanged (origin must be `human_authored`).

## EvidenceType (additive)

| Type | Value |
|------|-------|
| `url` | existing |
| `path` | existing (changed file path) |
| `adr` | existing |
| `commit` | **new** — full SHA |

## IngestRun (`git_ingest_runs`)

| Field | Notes |
|-------|-------|
| id | UUID |
| scope_kind / scope_key | Repository scope |
| actor_id | Triggering actor |
| local_path | Path supplied for the run |
| status | `running` \| `succeeded` \| `failed` |
| commits_seen | Commits read from git for this run |
| commits_skipped | Processed but no candidate (noisy or already seen this run's skip) |
| candidates_stored | New lore rows this run |
| cursor_sha / cursor_at | Watermark after the run (nullable if none) |
| error_summary | Set when failed |
| started_at / finished_at | Observability |

**Invariant**: At most one row with `status=running` per `(scope_kind, scope_key)`.

## IngestCursor (`git_ingest_cursors`)

| Field | Notes |
|-------|-------|
| scope_kind / scope_key | PK |
| last_sha | Last successfully processed SHA |
| last_committed_at | That commit's timestamp |
| updated_at | |

Used to bound `git log --since` on the next run. Idempotency still depends on ProcessedSHA.

## ProcessedSHA (`git_ingest_shas`)

| Field | Notes |
|-------|-------|
| scope_kind / scope_key / sha | PK |
| lore_entry_id | Set when a candidate was stored; null if skipped |
| skipped | true when no candidate |
| skip_reason | `merge`, `empty`, `noisy`, `no_rationale`, `too_long`, … |
| processed_at | |

Re-insert of the same PK is a no-op (idempotent).

## State transitions

```text
IngestRun:  (create) → running → succeeded
                          ↘ failed
ProcessedSHA: absent → {skipped | stored}  (never rewritten)
LoreEntry from ingest: unverified observational → (F035/verify later; not this slice)
```

## Validation

- Scope kind MUST be `repository`.
- SHA evidence MUST be non-empty; v1 prefers 40-hex but accepts git's full object name.
- Statement length ≤ `MaxStatementLength`; otherwise skip with `too_long` (do not invent a shorter claim).
- `max_commits` omitted or ≤0 means no cap (still bounded by unread history after cursor).
- Concurrent second `running` run for the same scope → conflict (HTTP 409).

## Relationships

- IngestRun belongs to a repository scope.
- ProcessedSHA optionally references lore_entries.id.
- Cursor is 1:1 with repository scope.
- Compile continues to read lore_entries only.

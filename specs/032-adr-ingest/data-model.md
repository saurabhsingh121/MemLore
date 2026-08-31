# Data Model: ADR Auto-Ingestion (F032)

## ADRSnapshot (not persisted as a row)

| Field | Notes |
|-------|-------|
| RelativePath | Path relative to the working-copy root (slash-separated, no leading `/`) |
| Checksum | SHA-256 hex of file bytes |
| Title | First markdown `#` heading or filename slug |
| StatusRaw | Status heading or front-matter `status:` (trimmed) |
| StatusClass | `accepted` \| `skip` \| `historical` \| `unknown` |
| Decision | Decision section body |
| Context | Context section body (optional) |
| Alternatives | Alternatives section body (optional) |
| Consequences | Consequences section body (optional) |
| Supersedes | Parsed ADR ids/slugs this file claims to replace |
| Components | Affected components when a Components/Affected heading exists |
| ADRID | Identity slug for evidence (filename stem, or ADR-NNNN from title) |

## ADRDerivedLore → LoreEntry

Reuses `lore_entries`. Created via `NewArchitectureDecisionLoreEntry`:

| Field | Accepted / adopted / approved | Deprecated / superseded |
|-------|------------------------------|-------------------------|
| statement | Faithful title + Decision (+ optional sections) | Same |
| scope | `{ kind: repository, key }` | Same |
| origin | `architecture_decision` | `architecture_decision` |
| verification_status | `verified` | created verified then **invalidated** in the same UoW |
| evidence | `[{ type: "adr", value: "<slug>" }, optional path]` | Same |
| created_by | Triggering actor | Same |
| verified_* | actor / now | actor / now, then invalidate audit |

`NewLoreEntry` is unchanged (human_authored only).
`NewObservationalLoreEntry` is unchanged (commit or pr only).

## EvidenceType (no new type)

| Type | Value |
|------|-------|
| `adr` | existing — ADR id/path slug (e.g. `0001-use-postgres`) |
| `path` | optional extra — relative file path |
| `url`, `commit`, `pr` | unchanged; git/PR observations stay those types |

## ADRIngestRun (`adr_ingest_runs`)

| Field | Notes |
|-------|-------|
| id | UUID |
| scope_kind / scope_key | Repository scope |
| actor_id | Triggering actor |
| local_path | Working-copy root |
| extra_dirs | Optional extra ADR dirs (text; empty if defaults only) |
| status | `running` \| `succeeded` \| `failed` |
| files_seen | Markdown files considered |
| files_skipped | Processed but no current lore stored (skip policy or historical-only) |
| lore_stored | New lore rows this run (includes historical-then-invalidated) |
| lore_superseded | Predecessors superseded this run |
| error_summary | Set when failed |
| started_at / finished_at | Observability |

**Invariant**: At most one row with `status=running` per `(scope_kind, scope_key)`
on this table (independent of git/PR running uniqueness).

## ADRIngestCursor (`adr_ingest_cursors`)

| Field | Notes |
|-------|-------|
| scope_kind / scope_key | PK |
| last_path | Last successfully processed relative path (nullable) |
| last_checksum | That file's checksum (nullable) |
| updated_at | Last successful run finish |

Status watermark only. Idempotency depends on ProcessedADR.

## ProcessedADR (`adr_ingest_files`)

| Field | Notes |
|-------|-------|
| scope_kind / scope_key / relative_path / checksum | PK |
| adr_id | Evidence slug |
| lore_entry_id | Set when lore was stored; null if skipped |
| skipped | true when no lore (draft/template/uncertain) |
| skip_reason | `template`, `readme`, `draft`, `rejected`, `no_decision`, `unknown_status`, `too_long`, `not_adr`, … |
| processed_at | |

Insert of the same PK is a conflict/no-op (idempotent). A **new** checksum for
the same `relative_path` is a new row. Predecessor lore is the latest
non-skipped row for that path (or matching `adr_id`) with a lore_entry_id.

**Why not `git_ingest_shas` / `pr_ingest_prs`**: those identities are SHA and
PR number. ADR identity is path + checksum.

## State transitions

```text
ADRIngestRun:  (create) → running → succeeded
                            ↘ failed

ProcessedADR: absent → {skipped | stored} for (path, checksum)
              same (path, checksum) → no-op
              same path, new checksum → new row + maybe supersede prior lore

LoreEntry (accepted): verified architecture_decision, current
LoreEntry (content change / explicit supersedes): prior ingest lore superseded
LoreEntry (deprecated/superseded file): stored then invalidated (not current)
```

## Validation

- Scope kind MUST be `repository`.
- Local path MUST be a readable directory (else failed run, no lore).
- ADR evidence MUST be non-empty slug.
- Statement length ≤ `MaxStatementLength`; otherwise skip with `too_long`.
- Concurrent second `running` ADR ingest run for the same scope → conflict (HTTP 409).
- Git or PR ingest `running` for the same scope is independent.
- Human-authored lore MUST NOT be a supersession predecessor in this slice.

## Relationships

- ADRIngestRun belongs to a repository scope.
- ProcessedADR optionally references lore_entries.id.
- ADR cursor is 1:1 with repository scope.
- Compile continues to read lore_entries only.

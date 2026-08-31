# Data Model: Suggested Lore Review Queue (F035)

## SuggestedLoreItem (projected, not a knowledge table)

Pending queue row derived from a **current** `LoreEntry`.

| Field | Notes |
|-------|-------|
| id | Observational lore entry id |
| statement | Lore statement (the extract) |
| evidence | Predecessor evidence list (`commit` and/or `pr`, plus optional path/url) |
| source_type | `commit` if primary evidence is commit, else `pr` |
| origin | Always `repository_observation` while pending |
| verification_status | Usually `unverified`; may be `verified` if someone used in-place verify (still pending until Accept/Reject) |
| scope | Repository scope |
| confidence | Omitted in v1 |
| reason | Omitted in v1 |

**Eligibility (pending)**:

- `IsCurrent` (not superseded, not invalidated)
- origin `repository_observation`
- primary evidence type `commit` or `pr`
- no `lore_review_decisions` row for this extract identity (accepted items are also non-current after supersede)

**Never eligible**: origin `architecture_decision`; F032 ADR evidence as the only identity; `human_authored` remember rows.

## ExtractIdentity

| Field | Notes |
|-------|-------|
| scope_kind / scope_key | Repository scope |
| evidence_type | `commit` or `pr` (primary) |
| evidence_value | SHA or `{owner}/{repo}#{n}` |
| statement_checksum | SHA-256 hex of `strings.TrimSpace(statement)` |

Primary evidence: first non-empty `commit` ref, else first non-empty `pr` ref.

## ReviewDecision (`lore_review_decisions`)

| Field | Notes |
|-------|-------|
| id | UUID |
| scope_kind / scope_key | Repository scope |
| evidence_type / evidence_value | Primary observational evidence |
| statement_checksum | SHA-256 hex |
| lore_entry_id | Observational lore id at decision time |
| successor_lore_id | Set on accept; null on reject |
| status | `accepted` \| `rejected` |
| actor_id | Reviewer |
| decided_at | UTC |

**Unique**: `(scope_kind, scope_key, evidence_type, evidence_value, statement_checksum)`.

Insert of the same key with the same status is a no-op (idempotent). Opposite
status → validation error (cannot Accept after Reject or Reject after Accept).

## Accepted successor → LoreEntry

| Path | Origin | Verification | Statement | Evidence | Actor |
|------|--------|--------------|-----------|----------|-------|
| Accept as stated | `human_verified` | `verified` | extract (trimmed) | copied from predecessor | reviewer |
| Edit then Accept | `human_authored` | `verified` | edited text | copied from predecessor | reviewer |

Created via `NewHumanVerifiedLoreEntry` or verified-human-authored helper, then
`ApplySupersessionWithSuccessor`. Predecessor `superseded_by_id` points at
successor. Create + supersede audits. Successor emits existing
`episode.ingest` outbox event. No new outbox types.

`NewLoreEntry` unchanged (unverified `human_authored` for remember).
`NewObservationalLoreEntry` unchanged.
`NewArchitectureDecisionLoreEntry` unchanged.

## Same-statement rule

Trim space on both sides. If equal, Accept-as-stated (`human_verified`). If
the request omits statement, treat as as-stated. If empty after trim when
explicitly sent → validation error.

## State transitions

```text
Observational current + no decision     → pending (in queue)
pending + Accept                        → predecessor superseded; successor current human-side verified; decision accepted
pending + Reject                        → observational still current; decision rejected; not in queue
accepted + Accept                       → idempotent (return existing successor)
rejected + Reject                       → idempotent
accepted + Reject / rejected + Accept   → validation_error
ADR / human remember / non-current      → not a review item (not found or validation_error)
```

## Validation

- Scope kind MUST be `repository` for list (v1 queue is repository-scoped, matching producers).
- Mutate requires non-empty actor.
- Edited statement length ≤ `MaxStatementLength`.
- Predecessor MUST be eligible pending (or idempotent already-decided as above).
- Concurrent double-accept: unique decision key + superseded predecessor → one successor.
- Cross-tenant: membership gate before list/get/mutate; get-by-id of foreign lore → not_found (existence leak, same as other get-by-id).

## Relationships

- ReviewDecision optionally references lore_entries.id (predecessor) and successor_lore_id.
- Compile continues to read lore_entries only.
- Ingest processed SHA/PR/ADR tables are untouched.
- `GET /v1/ingest/candidates` unchanged.

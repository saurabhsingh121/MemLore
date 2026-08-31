# Research: F110 invalidate + supersede

## R1 — Verification vs supersession as orthogonal states

**Decision**: Keep `verification_status` (`unverified` | `verified` | `invalidated`)
separate from supersession. Current vs superseded is derived from
`superseded_by_id IS NULL`.

**Rationale**: Authority model lists both factors independently. A verified
rule can later be superseded without becoming invalidated. Invalidated is the
terminal untrusted verification state; superseded is a history pointer.

**Rejected**: Combined enum (`current_verified`, `superseded`, …) — explodes
states and breaks existing verify tests. Separate `supersession_status`
column — redundant with the FK.

## R2 — Invalidate actor/time on the row vs audit-only

**Decision**: Store `invalidated_by` / `invalidated_at` on `lore_entries`
plus an `invalidate` audit (parity with `verified_by` / `verified_at`).

**Rationale**: GET without explain still shows who invalidated. Audit remains
the chronological source of truth.

## R3 — Supersede audit shape

**Decision**: Dual audits in one UoW: `supersede` on predecessor, `create` on
successor.

**Rationale**: Explain-on-predecessor shows the replacement event; explain-on-
successor looks like a normal create. Matches "preserve history" without a
new audit metadata column.

**Rejected**: Single audit with successor id in a payload column — requires
schema change to `audit_records`.

## R4 — Idempotency

**Decision**:
- Invalidate of already-invalidated: no-op, no audit (like verify).
- Invalidate of superseded: `validation_error`.
- Supersede of superseded or invalidated: `validation_error`, no successor.

**Rationale**: Silent re-supersede would fork history. Silent re-invalidate is
safe because the terminal state is already reached.

## R5 — Outbox

**Decision**: Successor uses existing `NewEpisodeIngestOutboxEvent` (create
path). No new event types for invalidate or predecessor supersede.

**Rationale**: Spec and F010: create-only outbox in v1. Graph-service
invalidate/supersede is deferred (F009). Filtering retrieval is F112.

## R6 — Verify interaction

**Decision**: `ApplyVerification` rejects `invalidated` and non-null
`superseded_by_id` with `validation_error`.

**Rationale**: Verify must not revive untrusted or historical entries.

## R7 — REST status codes

**Decision**: Invalidate `200` + predecessor body (like verify). Supersede
`201` + successor body (new resource, like create).

## R8 — sqlc regeneration

**Decision**: Update `db/queries/lore.sql` column lists; run
`docker run --rm -v "$PWD":/src -w /src sqlc/sqlc:1.28.0 generate` (or local
sqlc 1.28.0) and commit generated files.

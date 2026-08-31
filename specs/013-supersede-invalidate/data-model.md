# Data Model: F110 invalidate + supersede

## LoreEntry (extended)

Existing fields unchanged. Additions:

| Field | Type | Rules |
|-------|------|--------|
| `verification_status` | enum | `unverified` \| `verified` \| `invalidated` |
| `superseded_by_id` | uuid or null | Null = current. When set, points at successor `lore_entries.id` |
| `invalidated_by` | string or null | Set on first successful invalidate; unchanged on idempotent re-invalidate |
| `invalidated_at` | timestamptz or null | UTC; set on first successful invalidate |

### Invariants

- Statement, scope, origin, evidence, `created_by`/`created_at`, `verified_by`/`verified_at` are never cleared by invalidate or supersede.
- Successor: `origin = human_authored`, `verification_status = unverified`, same `scope` as predecessor, `created_by` = supersede actor, `superseded_by_id = null`.
- Predecessor after supersede: `superseded_by_id = successor.id`; verification status unchanged.
- An entry MUST NOT be both created as a successor and left without a predecessor pointer on the old row (transactional).

### State transitions

```text
invalidate:
  current + (unverified|verified) → invalidated   [audit invalidate]
  already invalidated             → no-op
  superseded                      → validation_error

supersede:
  current + not invalidated → predecessor.superseded_by_id = successor
                            + new successor (unverified, same scope)
  superseded or invalidated → validation_error

verify:
  current + unverified → verified   [existing]
  already verified     → no-op      [existing]
  invalidated or superseded → validation_error
```

## AuditRecord

`action` extends to: `create` | `verify` | `invalidate` | `supersede`.

No new audit columns. Target of `supersede` is the predecessor. Target of
successor `create` is the successor.

## Persistence (goose 00003)

```sql
ALTER TABLE lore_entries
  ADD COLUMN superseded_by_id VARCHAR(36) NULL
    REFERENCES lore_entries (id),
  ADD COLUMN invalidated_by VARCHAR(256) NULL,
  ADD COLUMN invalidated_at TIMESTAMPTZ NULL;
```

`verification_status` remains VARCHAR; application enum adds `invalidated`.
No CHECK constraint in v1 (matches existing verify).

Down: drop the three columns.

## Mapping

sqlc `LoreEntry` and postgres mapping include the three new columns on
insert, update, get, and list-by-scope.

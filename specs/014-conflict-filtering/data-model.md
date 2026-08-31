# Data Model: F112 temporal filtering + conflict detection

No new persisted tables. Lifecycle fields from F110 are inputs.

## Current vs stale (derived)

| Condition | Classification |
|-----------|----------------|
| `superseded_by_id` set | Stale (superseded) |
| `verification_status = invalidated` | Stale (invalidated) |
| Neither | **Current** |

An entry can theoretically be both superseded and invalidated only if prior
rules allowed it; F110 rejects invalidating superseded entries. Currentness
requires **both** not superseded **and** not invalidated.

## ConflictGroup (ephemeral, response-only)

| Field | Type | Notes |
|-------|------|-------|
| scope | `{ kind, key }` | Shared scope of the group |
| entry_ids | string[] | All current disagreeing entry ids in the retrieval set for that scope |
| statements | string[] | Distinct statements (original casing as stored), aligned or listed with ids |

**Rules**:
- Built only from current entries after temporal filter
- Grouping key: exact scope (`kind` + `key`)
- Conflict iff ≥2 distinct normalized statements in that scope within the set
- Not stored in PostgreSQL

## ContextPacket additions

| Field | Type | Default |
|-------|------|---------|
| conflicts | `ConflictGroup[]` | `[]` when none |

Existing `items`, `meta`, `warnings` unchanged in meaning. `items` never
include stale governance on default compile.

## include_stale (request)

| Surface | Field | Default | Effect |
|---------|-------|---------|--------|
| GET list / MCP search | `include_stale` | false | Include superseded/invalidated in list |
| knowledge_search | `include_stale` | false | Include in governance array |
| compile / get_for_task | n/a | — | Always current-only items |

## Unchanged entities

- `LoreEntry` fields and F110 transitions
- Audit records
- Outbox events
- Graph facts

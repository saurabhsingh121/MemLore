# REST Contract: Lore Entries (v1)

Base path: `/v1`  
Content-Type: `application/json`  
Actor header (mutating): `X-Memlore-Actor: <non-empty actor id>`

Error envelope:

```json
{
  "error": {
    "code": "validation_error|not_found|internal_error",
    "message": "human-readable summary",
    "details": []
  }
}
```

## POST /v1/lore-entries

Create human-authored lore entry.

**Headers**: `X-Memlore-Actor` required  
**Request**:

```json
{
  "statement": "string (1..8000)",
  "scope": { "kind": "team|repository|organization|project|feature|task", "key": "string" },
  "evidence": [
    { "type": "url|path|adr", "value": "string" }
  ]
}
```

`evidence` optional (default `[]`). Origin is always `human_authored` (not client-settable).

**Responses**:
- `201` → `LoreEntryResponse`
- `400` → validation (missing actor/statement/scope, bad evidence)
- `422` → schema validation

## GET /v1/lore-entries/{id}

**Responses**:
- `200` → `LoreEntryResponse`
- `404` → not found

## POST /v1/lore-entries/{id}/verify

**Headers**: `X-Memlore-Actor` required  
**Body**: empty object `{}` or none  
**Responses**:
- `200` → `LoreEntryResponse` (verified; idempotent if already verified)
- `400` → missing actor, or entry is invalidated/superseded
- `404` → not found

## POST /v1/lore-entries/{id}/invalidate

**Headers**: `X-Memlore-Actor` required  
**Body**: empty or `{}`  
**Responses**:
- `200` → `LoreEntryResponse` (`verification_status=invalidated`; idempotent if already invalidated)
- `400` → missing actor, or entry is superseded
- `404` → not found

## POST /v1/lore-entries/{id}/supersede

**Headers**: `X-Memlore-Actor` required  
**Body**:

```json
{
  "statement": "string (1..8000)",
  "evidence": [{ "type": "url|path|adr", "value": "string" }]
}
```

`evidence` optional (default `[]`). Scope is taken from the predecessor.

**Responses**:
- `201` → successor `LoreEntryResponse`
- `400` → missing actor, empty statement, superseded or invalidated predecessor
- `404` → not found

## GET /v1/lore-entries

List by exact scope.

**Query**:
- `scope_kind` (required)
- `scope_key` (required)

**Responses**:
- `200` → `{ "items": [LoreEntryResponse, ...] }` (possibly empty)
- `400` → missing/invalid query params

## GET /v1/lore-entries/{id}/audits

**Responses**:
- `200` → `{ "items": [AuditRecordResponse, ...] }` chronological ascending
- `404` → lore entry not found (never return an empty `items` list when the
  entry id does not exist)

## Schemas

### LoreEntryResponse

```json
{
  "id": "uuid",
  "statement": "string",
  "scope": { "kind": "repository", "key": "github.com/acme/app" },
  "origin": "human_authored",
  "verification_status": "unverified|verified|invalidated",
  "evidence": [{ "type": "adr", "value": "0001-dual-plane" }],
  "created_by": "alice",
  "created_at": "2026-08-25T12:00:00Z",
  "verified_by": null,
  "verified_at": null,
  "invalidated_by": null,
  "invalidated_at": null,
  "superseded_by_id": null,
  "updated_at": "2026-08-25T12:00:00Z"
}
```

### AuditRecordResponse

```json
{
  "id": "uuid",
  "target_id": "uuid",
  "action": "create|verify|invalidate|supersede",
  "actor_id": "alice",
  "created_at": "2026-08-25T12:00:00Z"
}
```

## Contract test expectations

| Case | Expect |
|------|--------|
| Create valid | 201 + provenance + create audit |
| Create duplicate statement/scope | 201 new id |
| Create missing actor/statement/scope | 400, no row |
| Get unknown | 404 |
| Verify then get | verified_by/at set; origin unchanged |
| Verify twice | 200; single verify audit |
| Invalidate current | 200; status invalidated; one invalidate audit |
| Invalidate twice | 200; single invalidate audit |
| Supersede current | 201 successor; predecessor `superseded_by_id` set |
| Supersede superseded | 400; no additional successor |
| List scope | only matching kind+key |
| Audits unknown id | 404 |

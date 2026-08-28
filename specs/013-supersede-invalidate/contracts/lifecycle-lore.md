# Contract: Invalidate + supersede (REST + MCP)

Extends [`specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`](../../001-scoped-lore-entry/contracts/rest-lore-entries.md)
and [`specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md`](../../002-mcp-lore-tools/contracts/mcp-lore-tools.md).

Error envelope and codes unchanged: `validation_error`, `not_found`.
MCP message format: `{code}: {message}`.

## LoreEntry fields (additive)

```json
{
  "superseded_by_id": null,
  "invalidated_by": null,
  "invalidated_at": null
}
```

`verification_status` values: `unverified` | `verified` | `invalidated`.

Audit `action` values: `create` | `verify` | `invalidate` | `supersede`.

---

## REST

### POST `/v1/lore-entries/{id}/invalidate`

**Headers**: `X-Memlore-Actor` required  
**Body**: empty or `{}`

**Responses**:
- `200` → updated predecessor `LoreEntryResponse`
- `400` → missing actor, or entry is superseded
- `404` → unknown id

Idempotent if already `invalidated`.

### POST `/v1/lore-entries/{id}/supersede`

**Headers**: `X-Memlore-Actor` required  
**Body**:

```json
{
  "statement": "string (1..8000)",
  "evidence": [{ "type": "url|path|adr", "value": "string" }]
}
```

`evidence` optional (default `[]`). Scope is not client-settable.

**Responses**:
- `201` → successor `LoreEntryResponse`
- `400` → missing actor, empty statement, superseded predecessor, invalidated predecessor
- `404` → unknown id

GET predecessor afterwards: `superseded_by_id` equals successor `id`.

---

## MCP

Advertised lore tools become **nine**: existing seven plus:

| Name | readOnlyHint | idempotentHint |
|------|--------------|----------------|
| `memlore.invalidate` | false | true |
| `memlore.supersede` | false | false |

### `memlore.invalidate`

```json
{ "id": "uuid", "actor_id": "non-empty" }
```

**Success**: predecessor `LoreEntry`.  
**Errors**: `validation_error`, `not_found`.

### `memlore.supersede`

```json
{
  "id": "uuid-of-entry-to-supersede",
  "statement": "replacement statement",
  "actor_id": "non-empty",
  "evidence": [{ "type": "adr", "value": "..." }]
}
```

**Success**: successor `LoreEntry`.  
**Errors**: `validation_error`, `not_found`.

---

## Contract test expectations

| Case | Expect |
|------|--------|
| Invalidate current | status `invalidated`; one `invalidate` audit; evidence preserved |
| Invalidate twice | no second audit; `invalidated_by` unchanged |
| Invalidate missing actor | `validation_error` |
| Invalidate unknown id | `not_found` |
| Invalidate superseded | `validation_error` |
| Supersede current | successor 201/tool result; predecessor `superseded_by_id` set; dual audits |
| Supersede superseded | `validation_error`; no new entry |
| Supersede invalidated | `validation_error` |
| MCP tools/list | exactly 9 tools including invalidate and supersede |
| Explain predecessor after supersede | audits include `supersede` in chronological order |

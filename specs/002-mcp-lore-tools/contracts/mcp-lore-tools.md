# MCP Contract: Lore Tools

**Server name**: `memlore`  
**Transport (this feature)**: stdio via CLI `memlore mcp`  
**Protocol**: MCP (official Python SDK v2 `MCPServer`)  
**Out of scope**: Streamable HTTP, Graphiti/Neo4j tools,
`memlore.get_for_task`, `memlore.supersede`, `memlore.invalidate`

JSON field names and enums on success match REST
[`LoreEntryResponse` / `AuditRecordResponse`](../../001-scoped-lore-entry/contracts/rest-lore-entries.md).

## Advertised tools

The server MUST list **exactly** these five tools (no Graphiti internals):

| Name | Role |
|------|------|
| `memlore.remember` | Create human-authored lore entry |
| `memlore.get` | Fetch by id |
| `memlore.verify` | Verify (idempotent) |
| `memlore.explain` | Entry fields + chronological audits |
| `memlore.search` | List by exact scope |

Recommended MCP annotations (not a substitute for tests):

| Tool | readOnlyHint | idempotentHint |
|------|--------------|----------------|
| remember | false | false (duplicates create a new id) |
| get | true | true |
| verify | false | true |
| explain | true | true |
| search | true | true |

## Error result

Business failures MUST be tool errors (`is_error: true`), not JSON-RPC
protocol failures.

**Message format**: `{code}: {message}`

| code | When |
|------|------|
| `validation_error` | Missing/blank `actor_id`, incomplete scope, invalid statement/evidence/enums |
| `not_found` | Unknown id on get / verify / explain |

There is **no** `unavailable` error code. Unexpected infrastructure failures
(e.g. database unreachable) stay on the SDK crash path: generic `is_error`
for the agent, traceback on **stderr** only.

Do not leak stack traces, SQL, or Graphiti/Neo4j internals in `content`.

Omitted required arguments are invalid at the MCP schema layer; the call MUST
NOT persist data.

---

## `memlore.remember`

Create a scoped human-authored lore entry (same rules as `POST /v1/lore-entries`).

**Arguments**:

```json
{
  "statement": "string (1..8000)",
  "scope": { "kind": "team|repository|organization|project|feature|task", "key": "string" },
  "actor_id": "string (required, non-empty)",
  "evidence": [
    { "type": "url|path|adr", "value": "string" }
  ]
}
```

`evidence` optional (default `[]`). Origin is always `human_authored` (not
client-settable).

**Success**: `LoreEntry` object (see schema below).  
**Errors**: `validation_error` (including missing/blank `actor_id`).

---

## `memlore.get`

**Arguments**:

```json
{ "id": "uuid-string" }
```

**Success**: `LoreEntry`  
**Errors**: `not_found`

---

## `memlore.verify`

Same semantics as `POST /v1/lore-entries/{id}/verify` (self-verify allowed;
re-verify is a no-op and does not add a second `verify` audit).

**Arguments**:

```json
{
  "id": "uuid-string",
  "actor_id": "string (required, non-empty)"
}
```

**Success**: `LoreEntry` (verified; original `verified_by` / `verified_at` on
re-verify)  
**Errors**: `validation_error` (missing/blank actor), `not_found`

---

## `memlore.explain`

**Arguments**:

```json
{ "id": "uuid-string" }
```

**Success**: `ExplainResult` — all `LoreEntry` fields plus chronological
`audits` and an ephemeral authority evaluation:

```json
{
  "audits": [ { "id": "uuid", "target_id": "uuid", "action": "create|verify", "actor_id": "string", "created_at": "RFC3339" } ],
  "trust_band": "medium",
  "authority_score": 0.58,
  "authority_factors": { "verification_status": "unverified", "origin": "human_authored" },
  "factor_breakdown": ["verification_status=unverified", "origin=human_authored"]
}
```

`audits` MUST be chronological ascending. No generated natural-language
summary field. See `specs/016-authority-factors/contracts/authority-evaluation.md`.

**Errors**: `not_found` (never an empty success payload for unknown id)

---

## `memlore.search`

Exact scope list (same as `GET /v1/lore-entries?scope_kind=&scope_key=`).
Not semantic/vector search.

**Arguments**:

```json
{
  "scope": { "kind": "team|repository|organization|project|feature|task", "key": "string" }
}
```

**Success**:

```json
{ "items": [ "LoreEntry", "..." ] }
```

Empty `items` when none match.  
**Errors**: `validation_error` when `scope.kind` or `scope.key` missing/invalid.

---

## Schemas

### LoreEntry

```json
{
  "id": "uuid",
  "statement": "string",
  "scope": { "kind": "repository", "key": "github.com/acme/app" },
  "origin": "human_authored",
  "verification_status": "unverified|verified",
  "evidence": [{ "type": "adr", "value": "0001-dual-plane" }],
  "created_by": "alice",
  "created_at": "2026-08-25T12:00:00Z",
  "verified_by": null,
  "verified_at": null,
  "updated_at": "2026-08-25T12:00:00Z"
}
```

### AuditRecord

```json
{
  "id": "uuid",
  "target_id": "uuid",
  "action": "create|verify",
  "actor_id": "alice",
  "created_at": "2026-08-25T12:00:00Z"
}
```

---

## CLI / stdio

```text
uv run memlore mcp
```

- stdin/stdout: MCP protocol only  
- stderr: logs  
- Process uses `MEMLORE_POSTGRES_DSN` (same default as REST)

Example agent config (Cursor / Claude-compatible):

```json
{
  "mcpServers": {
    "memlore": {
      "command": "uv",
      "args": ["run", "memlore", "mcp"],
      "cwd": "/path/to/memlore"
    }
  }
}
```

---

## Contract test expectations

| Case | Expect |
|------|--------|
| `tools/list` | exactly the five `memlore.*` names; no Graphiti/Neo4j tools |
| remember valid | success LoreEntry; origin `human_authored`; create audit exists (unit: `ListAuditsHandler`; full MCP: `memlore.explain` after US2) |
| remember duplicate statement/scope | success; **new** id |
| remember missing/blank `actor_id` | error; no row |
| remember invalid statement/scope | `validation_error`; no row |
| get unknown | `not_found` (`is_error`) |
| get existing | full LoreEntry fields |
| explain existing | LoreEntry fields + chronological audits |
| explain unknown | `not_found`, not empty audits |
| verify unverified | `verified`; origin unchanged; one verify audit |
| verify twice | still verified; original verifier/time; single verify audit |
| verify missing actor / unknown id | `validation_error` / `not_found` |
| search scope A vs B | only matching kind+key |
| search empty scope | `items: []` |
| search incomplete scope | `validation_error` |
| stdio CLI five-tool path | remember → get → verify → explain → search against governance DB (e2e; skip if Postgres down). Wall-clock “under 2 minutes” is a local quickstart check (SC-001), not a CI timer. |

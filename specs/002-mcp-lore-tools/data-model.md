# Data Model: MCP Lore Tools

This feature **does not add persisted entities or migrations**. Governance
data remains the 001 model in PostgreSQL. MCP introduces a **transport view**
over that model.

Canonical persisted definitions: [001 data-model](../001-scoped-lore-entry/data-model.md).

## Reused persisted entities

Unchanged from 001:

| Entity | MCP usage |
|--------|-----------|
| **LoreEntry** | remember creates; get/verify/explain/search return |
| **Scope** | remember + search (`kind` + `key`) |
| **EvidenceReference** | optional on remember |
| **AuditRecord** | create/verify still append; explain lists them |
| **Actor** | string `actor_id` on remember/verify only (not a table) |

### Invariants that MUST stay true over MCP

- Create origin is always `human_authored` (not client-settable).
- Duplicates allowed (same statement + scope → new id).
- Verify is idempotent; re-verify does not append a second `verify` audit.
- Statement, origin, scope, evidence, `created_by` / `created_at` never change
  on verify.
- Unknown id → not found (get, verify, explain), never an empty success body.
- Search with valid empty scope → empty `items`, not an error.

## Transport entity: MCP Tool Invocation

Not persisted. One agent call into the stdio server.

| Field | Type | Rules |
|-------|------|-------|
| tool | enum | `memlore.remember` \| `memlore.get` \| `memlore.verify` \| `memlore.explain` \| `memlore.search` |
| arguments | object | per-tool schema in [contracts/mcp-lore-tools.md](./contracts/mcp-lore-tools.md) |
| result | object \| error | success payload or `ToolError` (`validation_error` / `not_found`); infrastructure crashes are generic `is_error`, not a third code |

**Mutating tools** (`remember`, `verify`): `actor_id` required, non-empty after
strip. No environment default.

**Read tools** (`get`, `explain`, `search`): no actor argument.

## View: ExplainResult (not stored)

Composition of one `LoreEntry` plus its audits.

| Field | Type | Rules |
|-------|------|-------|
| *(all LoreEntry response fields)* | same as 001 `LoreEntryResponse` | required |
| audits | list[AuditRecord] | chronological (`created_at` ASC, then `id` ASC); empty list only if the entry exists and has no audits (should not happen after remember, which writes `create`) |

No natural-language `summary` field.

## Validation (adapter + existing domain)

| Input | Rule |
|-------|------|
| `actor_id` on remember/verify | required; non-empty after strip |
| `statement` | trimmed; 1..8000 chars (domain) |
| `scope.kind` / `scope.key` | same ScopeKind / key rules as REST |
| `evidence[]` | optional; each `type`+`value` as 001 |
| `id` on get/verify/explain | required string; unknown → not found |
| search scope incomplete | validation error |
| search scope valid, no rows | empty list |

## State transitions

Same as 001; MCP does not add transitions:

```text
[remember] --> unverified
unverified --verify--> verified
verified --verify--> verified  (no-op; no new audit)
```

Out of scope: supersede, invalidate, agent-authored origins, Graphiti nodes.

## Indexes / schema

None. Reuse `lore_entries (scope_kind, scope_key, created_at DESC)` and
`audit_records (target_id, created_at ASC, id ASC)`.

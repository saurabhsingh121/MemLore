# Suggested Lore Review Queue Contract (F035)

Authz matches ingest trigger / supersede (not admin-only verify):

- **Local mode**: mutating routes require `X-Memlore-Actor`; membership off.
- **OIDC**: Bearer JWT; list/get require `read` + F114 membership on the
  repository scope; accept/reject require `write` (writer/admin) + membership.
- Cross-tenant list/trigger of a known scope the caller cannot access:
  `403 forbidden`. Get-by-id of lore the caller cannot access: `404 not_found`.

CLI uses the same local DSN as `memlore ingest git` (not HTTP to `serve`).
`--actor` required for accept/reject (or `MEMLORE_ACTOR` env). `review list`
is a local read (no actor required).

MCP: unchanged. Tool count remains 10. No review tool.

`POST /v1/lore-entries` remains human-authored only.
`POST /v1/lore-entries/{id}/verify` remains origin-preserving verify.

F030–F032 ingest routes remain unchanged.

Pending items never include invented `confidence` or `reason` fields.

## CLI

```text
memlore review list --repository <key>
memlore review accept <id> [--statement <text>] [--actor <id>]
memlore review reject <id> [--actor <id>]
```

`review list` example:

```text
Suggested Lore (github.com/acme/payments)
  id: <uuid>
  statement: Payment events use transactional outbox.
  evidence: pr acme/payments#1842
  source: pr
  [pending]
```

Confidence/reason lines are omitted when absent. No JSON required on CLI.

Failed mutate prints `review accept failed: …` / `review reject failed: …`.

## REST — `GET /v1/review-queue`

Query: `scope_kind` (required, must be `repository`), `scope_key` (required),
optional `status` (`pending` default). Other status values may be omitted in
v1 (pending-only list is sufficient).

Authz: read + membership.

### Response `200`

```json
{
  "items": [
    {
      "id": "<lore-entry-id>",
      "statement": "Payment events use transactional outbox.",
      "scope": { "kind": "repository", "key": "github.com/acme/payments" },
      "origin": "repository_observation",
      "verification_status": "unverified",
      "evidence": [{ "type": "pr", "value": "acme/payments#1842" }],
      "source_type": "pr"
    }
  ]
}
```

Do **not** include `confidence` or `reason` keys in v1. F032 ADR lore MUST NOT
appear. Rejected and already-accepted extracts MUST NOT appear in `pending`.

Missing scope → `400`. Non-repository kind → `400`.

## REST — `GET /v1/review-queue/{id}`

Authz: read + membership on the item’s scope.

`200` pending item (same object as list element, plus `"status": "pending"`).
If a decision exists, include `"status": "accepted"|"rejected"` and on accept
`"successor_id"`. Unknown id or unauthorized → `404`.

## REST — `POST /v1/review-queue/{id}/accept`

Authz: write + membership.

### Request

```json
{
  "statement": "Payment events MUST use the transactional outbox."
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| statement | string | no | If omitted or equal to extract after trim → Accept-as-stated (`human_verified`). If different → Edit-then-Accept (`human_authored`). Empty string → `400`. |

### Response `200`

Successor lore entry body (existing lore presenter: id, statement, origin,
verification_status, evidence, scope, …). Origin is `human_verified` or
`human_authored`. Status is `verified`. Evidence includes predecessor refs.

Idempotent re-accept of the same extract returns the existing successor `200`.

Not pending (ADR, human remember, already rejected, superseded by someone
else): `400 validation_error` except idempotent re-accept.

Unknown id / unauthorized: `404`.

## REST — `POST /v1/review-queue/{id}/reject`

Authz: write + membership. Empty body.

### Response `200`

```json
{
  "id": "<lore-entry-id>",
  "status": "rejected",
  "actor_id": "alice"
}
```

Idempotent re-reject returns the same `200`. Accept-after-reject → `400`.
Unknown id / unauthorized → `404`.

## Errors

| Code | When |
|------|------|
| 400 | validation (blank actor, empty statement, wrong scope kind, not eligible, reject-after-accept) |
| 401 | missing actor (local mutate) or missing/invalid bearer (OIDC) |
| 403 | authenticated but no membership on list/mutate-by-scope |
| 404 | unknown id; cross-tenant get-by-id |
| 409 | reserved; v1 prefers idempotent 200 on double-accept |

## Dual-plane

Accept: lore successor + predecessor supersede + audits + outbox `episode.ingest`
in one UoW. Reject: decision row only (no lore mutation, no outbox). Graph
service is not called. `memlore worker` unchanged.

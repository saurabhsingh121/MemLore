# First-Class Decision Contract (F040)

Authz matches lore write / F035 review:

- **Local mode**: mutating routes require `X-Memlore-Actor`; membership off.
- **OIDC**: Bearer JWT; get/list require `read` + F114 membership on the
  repository scope; create/supersede require `write` (writer/admin) +
  membership.
- Cross-tenant list of a known scope the caller cannot access: `403
  forbidden`. Get-by-id the caller cannot access: `404 not_found`.

CLI uses the same local DSN as `memlore review` (not HTTP to `serve`).
`--actor` required for create/supersede (or `MEMLORE_ACTOR` env). `decision
get` / `decision list` are local reads (no actor required).

MCP: unchanged. Tool count remains 10. No decision tool.

`POST /v1/lore-entries` / `memlore.remember` remain human-authored lore only
(unverified snippets), not Decision create.

F030–F035 ingest and review routes remain unchanged.

## CLI

```text
memlore decision create --repository <key> --question <text> --choice <text> --owner <id> [--rationale <text>] [--consequences <text>] [--alternative <label>] [--alternative-note <note>] [--component <name>] [--date <RFC3339>] [--evidence <type:value>] [--actor <id>]
memlore decision get <id>
memlore decision list --repository <key>
memlore decision supersede <id> --question <text> --choice <text> --owner <id> [--rationale <text>] [--consequences <text>] [--alternative <label>] [--component <name>] [--date <RFC3339>] [--evidence <type:value>] [--actor <id>]
```

`--alternative` may repeat. If `--alternative-note` is used, it applies to
the last `--alternative` (v1: one note per alternative flag pair is not
required; notes may be omitted). `--component` and `--evidence` may repeat.
`--evidence` format `type:value` with existing evidence types (`url`,
`path`, `adr`, `commit`, `pr`).

`decision list` example:

```text
Decisions (github.com/acme/payments)
  id: <uuid>
  question: How should payment events be published?
  decision: Transactional outbox
  owner: alice
  source: human
  [current]
  id: <uuid>
  decision: Use PostgreSQL as the system of record
  owner: <ingest-actor>
  source: adr
  evidence: adr 0001-use-postgres
  [current]
```

No JSON required on CLI. Failed mutate prints `decision create failed: …` /
`decision supersede failed: …`.

## Decision JSON object

```json
{
  "id": "<uuid>",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "question": "How should payment events be published?",
  "decision": "Transactional outbox",
  "rationale": "Exactly-once publish with the DB transaction.",
  "alternatives": [{ "label": "Dual-write to the topic", "note": "Lost updates on crash" }],
  "consequences": "All payment services use the outbox worker.",
  "owner": "alice",
  "decided_at": "2026-09-01T12:00:00Z",
  "affected_components": ["payments-api", "outbox-worker"],
  "evidence": [{ "type": "url", "value": "https://wiki.example/outbox" }],
  "source_kind": "human",
  "superseded_by_id": null,
  "current": true,
  "created_by": "alice",
  "created_at": "2026-09-01T12:00:00Z"
}
```

Omit empty `alternatives` / `affected_components` as `[]`. Empty strings
for optional text are allowed (`""`). ADR projections MAY omit `question`
as `""` and use empty alternatives. `source_kind` is `human` or `adr`.
JSON field `decision` is the choice (avoid colliding with the resource
name in prose by using `choice` internally).

## REST — `POST /v1/decisions`

Authz: write + membership.

### Request

```json
{
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "question": "How should payment events be published?",
  "decision": "Transactional outbox",
  "rationale": "Exactly-once publish with the DB transaction.",
  "alternatives": [{ "label": "Dual-write to the topic", "note": "Lost updates on crash" }],
  "consequences": "All payment services use the outbox worker.",
  "owner": "alice",
  "decided_at": "2026-09-01T12:00:00Z",
  "affected_components": ["payments-api"],
  "evidence": [{ "type": "url", "value": "https://wiki.example/outbox" }]
}
```

| Field | Required | Notes |
|-------|----------|-------|
| scope | yes | `kind` must be `repository` |
| question | yes | non-empty |
| decision | yes | non-empty; lore statement |
| owner | yes | non-empty |
| rationale, alternatives, consequences, decided_at, affected_components, evidence | no | |

### Response `201`

Decision JSON object (`current: true`, `source_kind: "human"`).

## REST — `GET /v1/decisions/{id}`

Authz: read + membership on the item’s scope.

`200` Decision JSON (human row or ADR projection, including historical).
Unknown id or unauthorized → `404`.

## REST — `GET /v1/decisions`

Query: `scope_kind` (required, must be `repository`), `scope_key` (required).
List **current** only.

Authz: read + membership.

### Response `200`

```json
{
  "items": [ { "...Decision JSON..." } ]
}
```

Human and ADR-projected current Decisions. Git/PR observations and F035
pending items MUST NOT appear. Each current ADR choice appears at most once.

Missing scope → `400`. Non-repository kind → `400`.

## REST — `POST /v1/decisions/{id}/supersede`

Authz: write + membership.

Body: same fields as create **except** `scope` (successor stays in the
predecessor’s scope). `question`, `decision`, `owner` required.

### Response `201`

Successor Decision JSON (`source_kind: "human"`, `current: true`).
Predecessor is no longer current; get predecessor shows `current: false`
and `superseded_by_id` of the successor.

Already superseded/invalidated → `400 validation_error`.
Unknown id / unauthorized → `404`.

## Errors

| Code | When |
|------|------|
| 400 | validation (blank actor, missing required fields, wrong scope kind, not current) |
| 401 | missing actor (local mutate) or missing/invalid bearer (OIDC) |
| 403 | authenticated but no membership on list/create-by-scope |
| 404 | unknown id; cross-tenant get-by-id |
| 409 | reserved; v1 uses 400 if predecessor is no longer current |

## Dual-plane

Create: lore + decision row + children + create audit + outbox
`episode.ingest` in one UoW.

Supersede: successor lore + successor decision + predecessor lore
supersede + predecessor decision `superseded_by_id` (if a row exists) +
audits + successor outbox in one UoW.

Graph service is not called. `memlore worker` unchanged.

## Compile / MCP

`POST /v1/context/compile` and `memlore.get_for_task` keep section id
`decisions`. Current first-class Decision lore is classified into that
section. Ranking formulas unchanged. MCP tool list remains the existing
ten names.

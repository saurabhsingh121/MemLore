# Git Commit Ingest Contract (F030)

Authz matches other lore operations:

- **Local mode**: mutating routes require `X-Memlore-Actor`; membership off.
- **OIDC**: Bearer JWT; trigger requires `write` (writer/admin) + F114 membership
  on the repository scope; list/get require `read` + membership.
- Cross-tenant: `403 forbidden` (list/trigger) or `404 not_found` only when
  matching existing get-by-id leak-avoidance — ingest list/trigger use **403**
  like other collection operations for a known scope the caller cannot access.

CLI uses the same local DSN as `memlore profile` / `memlore context` (not HTTP
to `serve`). `--actor` required for `ingest git` (or `MEMLORE_ACTOR` env).
`ingest status` is a local read (no actor required, same as `profile`).

MCP: unchanged. Tool count remains 10.

`POST /v1/lore-entries` remains human-authored only.

## CLI

```text
memlore ingest git --repository <key> --path <dir> [--max-commits N] [--actor <id>]
memlore ingest status --repository <key>
```

Human-readable status example:

```text
Repository: github.com/acme/payments
Latest run: succeeded
  id: <uuid>
  commits seen: 12
  skipped: 10
  candidates stored: 2
  cursor: abcdef… @ 2026-08-01T12:00:00Z
```

Failed runs include `error: …`. No JSON required on CLI.

## REST — `POST /v1/ingest/git`

### Request

```json
{
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "path": "/var/repos/payments",
  "max_commits": 100
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| scope | object | yes | `kind` MUST be `repository` |
| path | string | yes | Local git directory |
| max_commits | integer | no | Positive cap; omit or ≤0 = no cap |

Authz: write + membership.

### Response `200`

Completed run body (`IngestRun`). If git/path validation fails before any SHA
work, `400 validation_error` and no run, **or** a `failed` run with
`error_summary` — prefer a persisted **failed** run when the actor and scope
were valid so status is observable (path-not-git). Invalid JSON / missing
scope / non-repository kind → `400` without a run.

### Response `409`

Another ingest is `running` for this repository.

## REST — `GET /v1/ingest/runs`

Query: `scope_kind=repository&scope_key=<key>` (both required).

Authz: read + membership.

Response `200`:

```json
{
  "items": [ { "id": "uuid", "status": "succeeded", "scope": { "kind": "repository", "key": "…" }, "commits_seen": 12, "commits_skipped": 10, "candidates_stored": 2, "cursor_sha": "…", "cursor_at": "…", "error_summary": "", "started_at": "…", "finished_at": "…" } ]
}
```

Newest first. Empty list if none.

## REST — `GET /v1/ingest/runs/{id}`

Authz: read + membership on the run's scope. Unknown id → `404`.

## REST — `GET /v1/ingest/candidates`

Query: `scope_kind=repository&scope_key=<key>` (both required).

Authz: read + membership.

Response `200`: current lore entries in that scope with
`origin=repository_observation` (same lore JSON shape as `GET /v1/lore-entries`).
Superseded/invalidated omitted (current-only), matching default list.

## Observability

Structured logs on trigger complete/fail: `repository_id` (scope key),
`run_id`, `commits_seen`, `candidates_stored`, `error` when failed.

## Out of contract

- MCP tools
- Web UI
- Promoting candidates to verified
- GitHub webhooks
- Changing compile JSON

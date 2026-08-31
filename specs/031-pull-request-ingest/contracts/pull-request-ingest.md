# Pull Request Ingest Contract (F031)

Authz matches other lore operations and F030:

- **Local mode**: mutating routes require `X-Memlore-Actor`; membership off.
- **OIDC**: Bearer JWT; trigger requires `write` (writer/admin) + F114 membership
  on the repository scope; list/get require `read` + membership.
- Cross-tenant: `403 forbidden` for list/trigger of a known scope the caller
  cannot access.

CLI uses the same local DSN as `memlore profile` / `memlore ingest git` (not HTTP
to `serve`). `--actor` required for `ingest pr` (or `MEMLORE_ACTOR` env).
`ingest status` is a local read (no actor required).

GitHub token: `MEMLORE_GITHUB_TOKEN`, else `GITHUB_TOKEN`. Never log the value.

MCP: unchanged. Tool count remains 10.

`POST /v1/lore-entries` remains human-authored only.

F030 git routes remain:

- `POST /v1/ingest/git`
- `GET /v1/ingest/runs`
- `GET /v1/ingest/runs/{id}`

## Repository mapping

Scope key `github.com/acme/payments` → GitHub owner `acme`, repo `payments`.
Any other shape → `400 validation_error`.

## CLI

```text
memlore ingest pr --repository <key> [--pr N] [--max-prs N] [--actor <id>]
memlore ingest status --repository <key> [--kind git|pr]
```

`--kind` default is `git` (F030 output unchanged). `--kind pr` example:

```text
Repository: github.com/acme/payments
Latest PR run: succeeded
  id: <uuid>
  prs seen: 12
  skipped: 10
  candidates stored: 2
  cursor: #1842 @ 2026-08-01T12:00:00Z
```

Failed runs include `error: …`. No JSON required on CLI.

## REST — `POST /v1/ingest/pr`

### Request

```json
{
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "pr": 1842,
  "max_prs": 100
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| scope | object | yes | `kind` MUST be `repository`; key MUST be `github.com/{owner}/{repo}` |
| pr | integer | no | Single PR number; omit for incremental merged list |
| max_prs | integer | no | Positive cap; omit or ≤0 = no cap |

Authz: write + membership.

Token is **not** in the JSON body (environment of the `serve` process).

### Response `200`

Completed PR run body (`PRIngestRun`). Invalid JSON / missing scope /
non-repository kind / unmappable key → `400` without a run. Missing GitHub
token or GitHub API failure after a valid actor/scope → persisted **failed**
run with `error_summary` (observable).

### Response `409`

Another **PR** ingest is `running` for this repository.

## REST — `GET /v1/ingest/pr-runs`

Query: `scope_kind=repository&scope_key=<key>` (both required).

Authz: read + membership.

Response `200`:

```json
{
  "items": [
    {
      "id": "uuid",
      "status": "succeeded",
      "scope": { "kind": "repository", "key": "github.com/acme/payments" },
      "pr": 0,
      "prs_seen": 12,
      "prs_skipped": 10,
      "candidates_stored": 2,
      "cursor_pr": 1842,
      "cursor_at": "2026-08-01T12:00:00Z",
      "error_summary": "",
      "started_at": "…",
      "finished_at": "…"
    }
  ]
}
```

Newest first. Empty list if none. `pr` is the single-PR filter used for the
run (`0` or omitted when incremental).

## REST — `GET /v1/ingest/pr-runs/{id}`

Authz: read + membership on the run's scope. Unknown id → `404`.

## REST — `GET /v1/ingest/candidates`

Query: `scope_kind=repository&scope_key=<key>` (both required).
Optional `evidence_type=pr` (or `commit`) filters observational lore that
has that evidence type.

Authz: read + membership.

Response `200`: current lore entries in that scope with
`origin=repository_observation` (same lore JSON shape as `GET /v1/lore-entries`),
optionally filtered. Superseded/invalidated omitted (current-only).

Without `evidence_type`, git-derived and PR-derived observational lore both
appear (F030 default preserved).

## Observability

Structured logs on trigger complete/fail: `repository_id` (scope key),
`run_id`, `prs_seen`, `candidates_stored`, `error` when failed. Never log
authorization headers or token env values.

## Out of contract

- MCP tools
- Web UI
- Promoting candidates to verified
- GitHub webhooks / check runs / review bot
- Changing compile JSON
- Changing F030 git request/response fields

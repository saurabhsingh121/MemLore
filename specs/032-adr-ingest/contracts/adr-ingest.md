# ADR Ingest Contract (F032)

Authz matches other lore operations and F030/F031:

- **Local mode**: mutating routes require `X-Memlore-Actor`; membership off.
- **OIDC**: Bearer JWT; trigger requires `write` (writer/admin) + F114 membership
  on the repository scope; list/get require `read` + membership.
- Cross-tenant: `403 forbidden` for list/trigger of a known scope the caller
  cannot access.

CLI uses the same local DSN as `memlore profile` / `memlore ingest git` (not HTTP
to `serve`). `--actor` required for `ingest adr` (or `MEMLORE_ACTOR` env).
`ingest status` is a local read (no actor required).

MCP: unchanged. Tool count remains 10.

`POST /v1/lore-entries` remains human-authored only.

F030 git and F031 PR routes remain:

- `POST /v1/ingest/git`
- `GET /v1/ingest/runs`
- `GET /v1/ingest/runs/{id}`
- `POST /v1/ingest/pr`
- `GET /v1/ingest/pr-runs`
- `GET /v1/ingest/pr-runs/{id}`

Default ADR directories (relative to `path`): `docs/adr`, `adr`,
`architecture/decisions`. Extra dirs are additive, not a generic doc crawl.

## CLI

```text
memlore ingest adr --repository <key> --path <dir> [--adr-dir <rel> ...] [--actor <id>]
memlore ingest status --repository <key> [--kind git|pr|adr]
```

`--kind` default is `git` (F030 output unchanged). `--kind adr` example:

```text
Repository: github.com/acme/payments
Latest ADR run: succeeded
  id: <uuid>
  files seen: 4
  skipped: 3
  lore stored: 1
  superseded: 0
```

Failed runs include `error: …`. No JSON required on CLI.

Missing `--path` root → validation error before or as a failed observable run
(no lore). Missing configured ADR subdirs inside a valid root → succeeded run
with zero lore.

## REST — `POST /v1/ingest/adr`

### Request

```json
{
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "path": "/abs/or/relative/working-copy",
  "adr_dirs": ["architecture/records"]
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| scope | object | yes | `kind` MUST be `repository` |
| path | string | yes | Local working-copy root |
| adr_dirs | string[] | no | Extra dirs relative to `path`; defaults always scanned |

Authz: write + membership.

### Response `200`

Completed ADR run body (`ADRIngestRun`). Invalid JSON / missing scope /
non-repository kind / missing path → `400` without a run. Missing directory
after a valid actor/scope → persisted **failed** run with `error_summary`.

### Response `409`

Another **ADR** ingest is `running` for this repository.

## REST — `GET /v1/ingest/adr-runs`

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
      "path": "/working-copy",
      "files_seen": 4,
      "files_skipped": 3,
      "lore_stored": 1,
      "lore_superseded": 0,
      "error_summary": "",
      "started_at": "…",
      "finished_at": "…"
    }
  ]
}
```

Newest first. Empty list if none.

## REST — `GET /v1/ingest/adr-runs/{id}`

Authz: read + membership on the run's scope. Unknown id → `404`.

## REST — `GET /v1/ingest/candidates`

Query: `scope_kind=repository&scope_key=<key>` (both required).
Optional `evidence_type=commit|pr|adr`.

Authz: read + membership.

| `evidence_type` | Result |
|-----------------|--------|
| omitted | Current `repository_observation` lore (F030/F031 default; git + PR) |
| `commit` or `pr` | Current observational lore with that evidence type |
| `adr` | Current `architecture_decision` lore that has `adr` evidence |

Superseded/invalidated omitted (current-only). Same lore JSON shape as
`GET /v1/lore-entries`.

Human-authored lore with `adr` evidence is **not** listed here (not
ingest-derived). Git/PR observations MUST NOT appear when `evidence_type=adr`.

## Observability

Structured logs on trigger complete/fail: `repository_id` (scope key),
`run_id`, `files_seen`, `lore_stored`, `error` when failed.

## Out of contract

- MCP tools
- Web UI
- F033 documentation ingest
- F035 accept/reject queue
- F040 Decision aggregate
- Changing compile JSON / ranking formulas
- Changing F030/F031 request/response fields
- Auto-verifying git/PR observational lore

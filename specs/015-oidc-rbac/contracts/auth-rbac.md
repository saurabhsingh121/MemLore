# Contract: Auth + RBAC (F111)

## Modes

### Local (OIDC unset)

- REST mutating: `X-Memlore-Actor` required (unchanged)
- MCP: `actor_id` required on mutating tools (unchanged)
- Principal role: `admin`

### OIDC configured

- REST: `Authorization: Bearer <jwt>` required for all `/v1/*` routes
- Actor header ignored for identity
- MCP: Bearer via `access_token` argument or `MEMLORE_ACCESS_TOKEN` env;
  `actor_id` ignored for identity
- `/health` remains open

## JWT claims

| Claim | Use |
|-------|-----|
| `sub` | Actor id |
| `iss` | Must match config issuer |
| `aud` | Must include config audience |
| `exp` | Must be valid |
| role claim (default `memlore_role`) | `reader` \| `writer` \| `admin` (string or array) |

## Errors

```json
{ "error": { "code": "unauthorized", "message": "...", "details": [] } }
{ "error": { "code": "forbidden", "message": "...", "details": [] } }
```

MCP: `{code}: {message}` with `unauthorized` / `forbidden`.

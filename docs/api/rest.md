# REST API

Authentication:
- **Local mode** (OIDC unset): mutating routes require `X-Memlore-Actor`
- **OIDC mode** (`MEMLORE_OIDC_ISSUER` + JWKS URL or HMAC secret): all `/v1/*`
  routes require `Authorization: Bearer <jwt>`; actor comes from `sub`; role
  from `memlore_role` (configurable). `/health` stays open.
- Errors: `unauthorized` (401), `forbidden` (403)

See [`specs/015-oidc-rbac/contracts/auth-rbac.md`](../../specs/015-oidc-rbac/contracts/auth-rbac.md).

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/health` | Liveness payload (`status`, `service`, `version`) |
| `POST` | `/v1/lore-entries` | Create human-authored lore (`X-Memlore-Actor` or Bearer) |
| `GET` | `/v1/lore-entries/{id}` | Get by id |
| `POST` | `/v1/lore-entries/{id}/verify` | Verify (admin when OIDC on) |
| `POST` | `/v1/lore-entries/{id}/invalidate` | Invalidate (admin when OIDC on) |
| `POST` | `/v1/lore-entries/{id}/supersede` | Supersede (writer/admin when OIDC on) |
| `GET` | `/v1/lore-entries` | List by `scope_kind` + `scope_key` (current only; `include_stale=true` for history) |
| `GET` | `/v1/lore-entries/{id}/audits` | List audits (404 if entry missing) |
| `POST` | `/v1/knowledge-search` | Dual-plane knowledge search (governance + graph; optional `include_stale`) |
| `POST` | `/v1/context/compile` | Compile token-budgeted context for a task (`conflicts` array) |

Contract details:
[`specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`](../../specs/001-scoped-lore-entry/contracts/rest-lore-entries.md).
Knowledge search: [`specs/011-graph-retrieval-orchestration/contracts/knowledge-search.md`](../../specs/011-graph-retrieval-orchestration/contracts/knowledge-search.md).
Context compile: [`specs/012-context-compiler/contracts/context-compile.md`](../../specs/012-context-compiler/contracts/context-compile.md).
Invalidate / supersede: [`specs/013-supersede-invalidate/contracts/lifecycle-lore.md`](../../specs/013-supersede-invalidate/contracts/lifecycle-lore.md).
Temporal filter + conflicts: [`specs/014-conflict-filtering/contracts/`](../../specs/014-conflict-filtering/contracts/).
Auth + RBAC: [`specs/015-oidc-rbac/contracts/auth-rbac.md`](../../specs/015-oidc-rbac/contracts/auth-rbac.md).

Example create (local mode):

```bash
curl -sS -X POST http://127.0.0.1:8080/v1/lore-entries \
  -H 'Content-Type: application/json' \
  -H 'X-Memlore-Actor: alice' \
  -d '{
    "statement": "Payment events must use the transactional outbox.",
    "scope": {"kind": "repository", "key": "github.com/acme/payments"},
    "evidence": [{"type": "adr", "value": "0001-dual-plane-architecture"}]
  }'
```

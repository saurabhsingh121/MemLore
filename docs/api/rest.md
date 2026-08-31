# REST API

Authentication:
- **Local mode** (OIDC unset): mutating routes require `X-Memlore-Actor`;
  membership checks are off (trusted actor is admin).
- **OIDC mode** (`MEMLORE_OIDC_ISSUER` + JWKS URL or HMAC secret): all `/v1/*`
  routes require `Authorization: Bearer <jwt>`; actor comes from `sub`; role
  from `memlore_role` (configurable). Membership enforcement is **on**:
  non-admin principals may only access lore in scopes they belong to.
  JWT `admin` bypasses membership. `/health` stays open.
- Errors: `unauthorized` (401), `forbidden` (403), `not_found` (404; also used
  for cross-tenant get-by-id to avoid existence leaks)

See [`specs/015-oidc-rbac/contracts/auth-rbac.md`](../../specs/015-oidc-rbac/contracts/auth-rbac.md)
and [`specs/018-membership-authz/contracts/membership-authz.md`](../../specs/018-membership-authz/contracts/membership-authz.md).

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/health` | Liveness payload (`status`, `service`, `version`) |
| `POST` | `/v1/lore-entries` | Create human-authored lore (`X-Memlore-Actor` or Bearer) |
| `GET` | `/v1/lore-entries/{id}` | Get by id |
| `GET` | `/v1/lore-entries/{id}/explain` | Entry + audits + authority evaluation (parity with `memlore.explain`) |
| `POST` | `/v1/lore-entries/{id}/verify` | Verify (admin when OIDC on) |
| `POST` | `/v1/lore-entries/{id}/invalidate` | Invalidate (admin when OIDC on) |
| `POST` | `/v1/lore-entries/{id}/supersede` | Supersede (writer/admin when OIDC on) |
| `GET` | `/v1/lore-entries` | List by `scope_kind` + `scope_key` (current only; `include_stale=true` for history) |
| `GET` | `/v1/lore-entries/{id}/audits` | List audits (404 if entry missing) |
| `POST` | `/v1/knowledge-search` | Dual-plane knowledge search (governance + graph; optional `include_stale`) |
| `POST` | `/v1/context/compile` | Compile token-budgeted named context packet for a task (`sections`, `sources`; `items` retained; additive optional files/ticket/agent_id) |
| `POST` | `/v1/repository-profile` | Compile a repository intelligence profile (named sections; omit empty) |
| `POST` | `/v1/admin/teams` | Create team (admin) |
| `POST` | `/v1/admin/projects` | Create project (admin; optional `team_key`) |
| `POST` | `/v1/admin/teams/{key}/members` | Add team member (admin) |
| `DELETE` | `/v1/admin/teams/{key}/members/{subject}` | Remove team member (admin) |
| `POST` | `/v1/admin/projects/{key}/members` | Add project member (admin) |
| `DELETE` | `/v1/admin/projects/{key}/members/{subject}` | Remove project member (admin) |
| `POST` | `/v1/admin/scope-bindings` | Bind repository/feature/task scope to a project (admin) |
| `DELETE` | `/v1/admin/scope-bindings` | Unbind child scope (`scope_kind` + `scope_key` query) (admin) |

Contract details:
[`specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`](../../specs/001-scoped-lore-entry/contracts/rest-lore-entries.md).
Knowledge search: [`specs/019-semantic-graph-retrieval/contracts/knowledge-search-v2.md`](../../specs/019-semantic-graph-retrieval/contracts/knowledge-search-v2.md)
(extends F108). Governance hits are query-relevant; optional `graph_receipt`
when a graph fact collapses onto lore; scope omitted searches membership-allowed
lore (local: all).
Context compile: [`specs/012-context-compiler/contracts/context-compile.md`](../../specs/012-context-compiler/contracts/context-compile.md)
(v1 fields). F021 packet: [`specs/021-agent-context-bootstrap/contracts/context-packet.md`](../../specs/021-agent-context-bootstrap/contracts/context-packet.md).
Repository profile: [`specs/020-repo-intelligence-profile/contracts/repository-profile.md`](../../specs/020-repo-intelligence-profile/contracts/repository-profile.md).
Invalidate / supersede: [`specs/013-supersede-invalidate/contracts/lifecycle-lore.md`](../../specs/013-supersede-invalidate/contracts/lifecycle-lore.md).
Temporal filter + conflicts: [`specs/014-conflict-filtering/contracts/`](../../specs/014-conflict-filtering/contracts/).
Auth + RBAC: [`specs/015-oidc-rbac/contracts/auth-rbac.md`](../../specs/015-oidc-rbac/contracts/auth-rbac.md).
Membership authz: [`specs/018-membership-authz/contracts/membership-authz.md`](../../specs/018-membership-authz/contracts/membership-authz.md).
Authority evaluation: [`specs/016-authority-factors/contracts/authority-evaluation.md`](../../specs/016-authority-factors/contracts/authority-evaluation.md).

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

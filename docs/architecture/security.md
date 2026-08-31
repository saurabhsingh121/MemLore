# Security

Baseline security posture for MemLore:

- **Authentication**: Optional OIDC/OAuth2 Bearer JWT for humans and service
  principals (F111). Local mode (OIDC unset) trusts `X-Memlore-Actor` /
  MCP `actor_id` as admin for dogfood/CI.
- **Authorization**: explicit, scope-aware (organization → team → project →
  repository → task). Role answers verbs (`reader` / `writer` / `admin`);
  membership answers which scopes (F010 / F114). Every read is
  authorization-aware; every write is auditable. JWT `admin` may bypass
  membership (platform operator).
- **Tenant isolation**: workspace/team boundaries enforced in the Go
  governance plane before knowledge retrieval or mutation (not in
  graph-service). Cross-tenant get-by-id returns `not_found` (no existence
  leak).
- **Secrets**: never commit secrets; use `.env` locally; never log tokens,
  credentials, or sensitive raw prompts unnecessarily.
- **Untrusted agent context**: treat stored agent observations/inferences as
  untrusted input (prompt injection, malicious stored context).
- **Input validation**: validate all REST/MCP boundaries at adapters.
- **Dependency hygiene**: justify new third-party libraries; prefer maintained
  packages with acceptable licenses.

See constitution principle VII, ADR 0004, and
`specs/018-membership-authz/contracts/membership-authz.md`.

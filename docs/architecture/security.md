# Security

Baseline security posture for MemLore:

- **Authentication**: OIDC/OAuth2 planned for humans and service principals.
- **Authorization**: explicit, scope-aware (organization → team → project →
  repository → task). Every read is authorization-aware; every write is
  auditable.
- **Tenant isolation**: workspace/team boundaries enforced in the governance
  plane before knowledge retrieval.
- **Secrets**: never commit secrets; use `.env` locally; never log tokens,
  credentials, or sensitive raw prompts unnecessarily.
- **Untrusted agent context**: treat stored agent observations/inferences as
  untrusted input (prompt injection, malicious stored context).
- **Input validation**: validate all REST/MCP boundaries with Pydantic schemas.
- **Dependency hygiene**: justify new third-party libraries; prefer maintained
  packages with acceptable licenses.

See constitution principle VII and ADR 0004 when auth lands.

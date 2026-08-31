# Research: F111 OIDC + RBAC

## R1 — Optional until configured

**Decision**: OIDC off unless issuer + (JWKS URL or HMAC secret) set.

**Rationale**: Clarification Q1-A; keeps CI/dogfood green.

## R2 — Role matrix

**Decision**: reader / writer / admin as in spec Q2-A.

| Operation | reader | writer | admin |
|-----------|--------|--------|-------|
| get, list, search, knowledge_search, compile, explain, health* | ✓ | ✓ | ✓ |
| create / remember / supersede | | ✓ | ✓ |
| verify / invalidate | | | ✓ |

\*health always public.

## R3 — JWT library

**Decision**: `github.com/golang-jwt/jwt/v5` + stdlib JWKS JSON fetch.

**Alternatives**: go-oidc (heavier); manual crypto only (error-prone).

## R4 — Test credentials

**Decision**: Support `MEMLORE_OIDC_HMAC_SECRET` for HS256 tokens in tests and
simple deployments; production typically uses JWKS URL.

## R5 — MCP token

**Decision**: When OIDC on: `access_token` tool field OR `MEMLORE_ACCESS_TOKEN`
env; ignore `actor_id` for identity. Local mode: existing `actor_id`.

## R6 — Local role

**Decision**: Local-mode principals are `admin` so existing contracts need no
role claims.

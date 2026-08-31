# Data Model: F111 auth (no new tables)

## Principal (request-scoped)

| Field | Notes |
|-------|-------|
| Subject | Token `sub` or local actor string |
| Role | reader \| writer \| admin |

## AuthConfig

| Field | Env | Required when OIDC on |
|-------|-----|------------------------|
| Issuer | `MEMLORE_OIDC_ISSUER` | yes |
| Audience | `MEMLORE_OIDC_AUDIENCE` | yes |
| JWKSURL | `MEMLORE_OIDC_JWKS_URL` | one of JWKS or HMAC |
| HMACSecret | `MEMLORE_OIDC_HMAC_SECRET` | one of JWKS or HMAC |
| RoleClaim | `MEMLORE_OIDC_ROLE_CLAIM` | default `memlore_role` |

Configured iff Issuer set AND (JWKSURL or HMACSecret) set.

## Errors

| Code | HTTP | Meaning |
|------|------|---------|
| `unauthorized` | 401 | Missing/invalid token when OIDC on |
| `forbidden` | 403 | Authenticated but role lacks permission |
| `validation_error` | 400 | Local mode missing actor (unchanged) |

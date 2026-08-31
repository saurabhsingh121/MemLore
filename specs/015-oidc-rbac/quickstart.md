# Quickstart: F111 auth

## Local mode (default)

Unset OIDC env vars. Existing `X-Memlore-Actor` / `actor_id` flows work.

```bash
go test ./...
```

## OIDC-on smoke (HMAC)

```bash
export MEMLORE_OIDC_ISSUER=http://localhost/memlore
export MEMLORE_OIDC_AUDIENCE=memlore
export MEMLORE_OIDC_HMAC_SECRET=test-secret
# Issue HS256 JWT with sub=alice, memlore_role=admin, iss/aud/exp set
curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' ...
```

Reader token cannot verify; admin can.

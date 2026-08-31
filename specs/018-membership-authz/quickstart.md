# Quickstart: Membership-scoped authorization

## Local mode (default)

No membership seed required. Existing lore curl / MCP dogfood unchanged:

```bash
# OIDC unset
curl -sS -X POST http://127.0.0.1:8080/v1/lore-entries \
  -H 'Content-Type: application/json' \
  -H 'X-Memlore-Actor: alice' \
  -d '{"statement":"…","scope":{"kind":"team","key":"alpha"},"evidence":[]}'
```

## OIDC + membership (enforcement on)

1. Set F111 OIDC env (`MEMLORE_OIDC_ISSUER`, HMAC or JWKS, audience).
2. Run migrations (`memlore migrate`) — includes `00004_membership`.
3. As admin, create team/project, add members, bind child scopes.
4. As writer JWT for a member subject, create/list only allowed scopes.

```bash
# Admin: create team + member (Bearer admin JWT)
curl -sS -X POST http://127.0.0.1:8080/v1/admin/teams \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H 'Content-Type: application/json' \
  -d '{"key":"alpha","name":"Alpha"}'

curl -sS -X POST http://127.0.0.1:8080/v1/admin/teams/alpha/members \
  -H "Authorization: Bearer $ADMIN_JWT" \
  -H 'Content-Type: application/json' \
  -d '{"subject":"alice"}'

# Writer alice: create lore in alpha
curl -sS -X POST http://127.0.0.1:8080/v1/lore-entries \
  -H "Authorization: Bearer $ALICE_WRITER_JWT" \
  -H 'Content-Type: application/json' \
  -d '{"statement":"…","scope":{"kind":"team","key":"alpha"},"evidence":[]}'
```

## Tests

```bash
go test ./...
# Membership allow/deny lives in domain + HTTP/MCP contract suites with HMAC OIDC fixtures
```

## sqlc after query changes

```bash
sqlc generate
```

Committed output: `internal/infrastructure/postgres/sqlc/`.

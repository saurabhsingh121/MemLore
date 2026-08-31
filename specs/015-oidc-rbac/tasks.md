# Tasks: OIDC + RBAC (F111)

**Input**: `specs/015-oidc-rbac/`  
**Tests**: TDD RED → GREEN → REFACTOR

## Phase 1: Setup

- [x] T001 Confirm branch `015-oidc-rbac` and spec artifacts

## Phase 2: Domain RBAC

- [x] T002 [P] RED: role permission matrix tests
- [x] T003 GREEN: `Role`, `Permission`, `Authorize` in `internal/domain/rbac.go`
- [x] T004 [P] RED/GREEN: `UnauthorizedError` / `ForbiddenError` in `errors.go`

## Phase 3: Authenticator

- [x] T005 [P] RED: local authenticator + JWT HMAC verifier tests
- [x] T006 GREEN: `internal/application/auth` + `internal/adapters/auth` (local, JWT)
- [x] T007 GREEN: config from env (`AuthConfig.Enabled()`)

## Phase 4: HTTP + MCP

- [x] T008 [P] RED: HTTP OIDC-on unauthorized/forbidden/actor-from-sub tests
- [x] T009 GREEN: middleware + `requirePrincipal` + error mapping
- [x] T010 [P] RED: MCP OIDC-on remember/verify auth tests
- [x] T011 GREEN: MCP token resolve + authorize; local mode unchanged
- [x] T012 Wire config in `cmd/memlore` and HTTP/MCP constructors

## Phase 5: Docs

- [x] T013 Update `docs/api/rest.md`, `docs/api/mcp.md`, FEATURE_DEVELOPMENT, specify-rules
- [x] T014 `go test ./...` green

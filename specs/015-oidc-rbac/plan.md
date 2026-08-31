# Implementation Plan: OIDC + RBAC (F111)

**Branch**: `015-oidc-rbac` | **Date**: 2026-08-31 | **Spec**: [spec.md](./spec.md)

## Summary

Add optional OIDC Bearer JWT authentication and coarse RBAC (`reader` /
`writer` / `admin`). When OIDC is unset, preserve today’s actor header /
`actor_id` local mode (implicit admin). When configured, derive actor from
`sub`, authorize by role claim, and ignore spoofable actor fields.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: chi; `github.com/golang-jwt/jwt/v5` (JWT); stdlib JWKS fetch  
**Storage**: No new tables in v1  
**Testing**: `go test ./...`; HMAC JWT fixtures for OIDC-on contract tests  
**Target Platform**: `memlore serve`, `memlore mcp`  
**Constraints**: TDD; no team/project membership; secrets not logged  
**Scale/Scope**: Auth gateway + RBAC matrix; REST + MCP parity

## Constitution Check

- [x] TDD RED→GREEN→REFACTOR
- [x] Spec-driven (clarifications Q1–Q3 encoded)
- [x] Architecture: domain RBAC pure; OIDC in adapters; ports for verifier
- [x] Documentation: REST/MCP API + contracts + FEATURE_DEVELOPMENT
- [x] Authority ≠ authorization (knowledge authority unchanged)
- [x] Temporal correctness: N/A beyond not changing lore semantics
- [x] Secure by default: fail closed when OIDC configured; least privilege roles
- [x] Observability: auth failures as structured errors (no token logging)
- [x] Dependency policy: golang-jwt justified for JWT parse/verify
- [x] Simplicity: optional-until-configured; no membership schema

## Project Structure

```text
specs/015-oidc-rbac/
├── plan.md, research.md, data-model.md, quickstart.md, tasks.md
└── contracts/auth-rbac.md

internal/domain/rbac.go, errors.go          # roles, permissions, auth errors
internal/application/auth/                  # Authenticator port + resolve helpers
internal/adapters/auth/                     # local + JWT (HMAC/JWKS) verifiers
internal/adapters/http/                     # middleware, requirePrincipal
internal/adapters/mcp/                      # Bearer/env token + authorize
cmd/memlore/main.go                         # wire auth config from env
```

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/auth-rbac.md](./contracts/auth-rbac.md)
- [quickstart.md](./quickstart.md)

## Phase 2 — Tasks

See [tasks.md](./tasks.md).

# Feature Specification: Go Domain Primitives (F102)

**Feature Branch**: `004-go-domain-primitives`  
**Created**: 2026-08-25  
**Status**: Draft  
**Input**: Port Python governance domain types to Go with characterization parity

**Depends on**: F101 (Go skeleton)

## User Scenarios & Testing

### User Story 1 — Typed enums and errors (Priority: P1)

A contributor uses Go domain enums and errors that serialize to the same string
values as the Python domain and distinguish validation from not-found failures.

**Independent Test**: Enum string values and `ValidationError` messages match
Python characterization fixtures.

**Acceptance Scenarios**:

1. **Given** `ScopeKind`, `EvidenceType`, `KnowledgeOrigin`, `VerificationStatus`,
   and `AuditAction`, **When** string values are compared to Python, **Then** they
   match exactly (e.g. `team`, `human_authored`, `unverified`).
2. **Given** invalid enum input strings, **When** parsed, **Then** a validation
   error is returned.

---

### User Story 2 — Scope and evidence value objects (Priority: P1)

A contributor constructs `Scope` and `EvidenceReference` with trimming and length
rules identical to Python.

**Acceptance Scenarios**:

1. **Given** a scope key with surrounding whitespace, **When** constructed,
   **Then** the key is trimmed.
2. **Given** a blank scope key or evidence value, **When** constructed,
   **Then** validation fails with the same message as Python.
3. **Given** evidence value over 2048 chars, **When** constructed, **Then**
   validation fails.

---

### User Story 3 — Lore entry and verification rules (Priority: P1)

A contributor creates human-authored lore entries and applies verification with
the same invariants as Python `LoreEntry` and `apply_verification`.

**Acceptance Scenarios**:

1. **Given** valid statement, scope, and creator, **When** a lore entry is
   created, **Then** defaults are `human_authored`, `unverified`, non-empty id.
2. **Given** oversize statement or non-human origin on create, **When** created,
   **Then** validation fails (same messages as Python).
3. **Given** an unverified entry, **When** `ApplyVerification` runs, **Then**
   status becomes verified, timestamps and actor set, verify audit returned.
4. **Given** an already verified entry, **When** `ApplyVerification` runs again,
   **Then** entry unchanged and no audit returned (idempotent).

## Requirements

- **FR-001**: Go package `internal/domain` MUST NOT import adapters, HTTP, pgx, or MCP.
- **FR-002**: Enum string values MUST match Python `enums.py`.
- **FR-003**: Validation error messages MUST match Python for characterized cases.
- **FR-004**: `MaxStatementLength` = 8000, `MaxScopeKeyLength` = 512,
  `MaxEvidenceValueLength` = 2048.
- **FR-005**: Create lore MUST require `human_authored` origin.
- **FR-006**: Characterization tests MUST reference Python test sources in comments.

## Success Criteria

- **SC-001**: `go test ./internal/domain/...` passes with table-driven characterization tests.
- **SC-002**: Python `uv run pytest tests/unit/domain` still passes unchanged.
- **SC-003**: No new third-party domain dependencies beyond `google/uuid` (optional).

## Out of Scope

- Application handlers, repositories, REST/MCP (F103–F105)
- Authority scoring, supersession, conflicts (F003, F008, F009)

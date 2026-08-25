# Feature Specification: Scoped Human-Authored Lore Entry

**Feature Branch**: `001-scoped-lore-entry`  
**Created**: 2026-08-25  
**Status**: Draft  
**Input**: User description: "Store, retrieve and verify a scoped human-authored Lore entry with provenance"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remember scoped lore (Priority: P1)

An engineer records a durable piece of engineering knowledge (for example, an
architectural decision or coding convention) against a clear scope such as a
repository or team, including who authored it and optional evidence pointers.

**Why this priority**: Without the ability to store scoped, attributed knowledge,
MemLore cannot fulfill its core purpose.

**Independent Test**: Create a lore entry with scope, statement, human authorship,
and optional evidence; confirm it is persisted and returned with a stable
identifier and provenance fields.

**Acceptance Scenarios**:

1. **Given** a valid scope and a non-empty knowledge statement, **When** a human
   actor records the lore entry as human-authored, **Then** the system stores it
   and returns a stable identifier plus provenance (actor, origin, created time,
   scope).
2. **Given** a knowledge statement with one or more evidence references (for
   example path, URL, or decision record id), **When** the entry is recorded,
   **Then** those evidence references are retained with the entry and visible on
   retrieval.
3. **Given** a missing statement or missing scope, **When** a create attempt is
   made, **Then** the system rejects the request with a clear validation error
   and stores nothing.

---

### User Story 2 - Retrieve lore with provenance (Priority: P1)

An engineer or coding-agent operator fetches a previously recorded lore entry by
identifier and sees the statement, scope, authorship origin, verification state,
and evidence.

**Why this priority**: Storage without trustworthy retrieval does not deliver
usable memory.

**Independent Test**: After creating an entry, retrieve it by id and confirm all
required provenance and content fields are present and accurate.

**Acceptance Scenarios**:

1. **Given** an existing lore entry id, **When** a caller retrieves it, **Then**
   the system returns the statement, scope, origin (`human_authored`),
   verification status, evidence, and timestamps.
2. **Given** an unknown lore entry id, **When** a caller retrieves it, **Then**
   the system reports that the entry was not found without leaking internal
   details.

---

### User Story 3 - Verify human-authored lore (Priority: P1)

A human reviewer marks a lore entry as verified, raising its trust posture
without deleting history or changing authorship origin.

**Why this priority**: Authority and verification are first-class MemLore
promises; unverified and verified knowledge must remain distinguishable.

**Independent Test**: Verify an unverified entry and confirm status changes to
verified while provenance and original statement remain intact; attempting to
treat unverified knowledge as verified without this action must fail.

**Acceptance Scenarios**:

1. **Given** an unverified human-authored lore entry, **When** an authorized
   human actor verifies it, **Then** verification status becomes verified, the
   verifier and verification time are recorded, and the original statement and
   authorship remain unchanged.
2. **Given** an already verified lore entry, **When** a caller verifies it again,
   **Then** the system remains verified and does not invent a conflicting second
   “canonical” identity for the same entry (idempotent or clearly rejected as
   already verified—either is acceptable if documented in behavior).
3. **Given** a lore entry that is not verified, **When** a caller asks whether it
   is verified/canonical, **Then** the system reports it as unverified and does
   not present it as verified knowledge.

---

### User Story 4 - List lore within a scope (Priority: P2)

An engineer lists lore entries for a given scope to discover what has already
been recorded.

**Why this priority**: Valuable for early usability, but store/get/verify alone
already form a complete vertical slice.

**Independent Test**: Create multiple entries in two scopes; listing one scope
returns only that scope’s entries.

**Acceptance Scenarios**:

1. **Given** lore entries in scope A and scope B, **When** a caller lists scope A,
   **Then** only scope A entries are returned.
2. **Given** a scope with no entries, **When** listed, **Then** the result is an
   empty collection (not an error).

---

### Edge Cases

- Creating an entry with an empty or whitespace-only statement is rejected.
- Evidence references that are empty strings are rejected; omitting evidence is
  allowed.
- Retrieving or verifying a deleted-or-unknown id yields not-found (hard delete
  is out of scope; this slice has no delete operation).
- Verification by a missing/blank actor identity is rejected.
- Extremely long statements beyond a documented maximum length are rejected with
  a validation error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow a human actor to create a lore entry with a
  non-empty statement, a scope, and origin `human_authored`.
- **FR-002**: System MUST accept optional evidence references on create and
  retain them with the entry.
- **FR-003**: System MUST assign each lore entry a stable unique identifier.
- **FR-004**: System MUST persist and return provenance for every entry: author
  actor identity, origin, created timestamp, scope, verification status, and
  evidence.
- **FR-005**: System MUST allow retrieval of a lore entry by its identifier.
- **FR-006**: System MUST allow an authorized human actor to verify an unverified
  human-authored lore entry, recording verifier identity and verification time.
- **FR-007**: System MUST NOT present unverified lore as verified.
- **FR-008**: System MUST reject create/verify requests that fail validation
  (missing statement, scope, or actor) without partial persistence.
- **FR-009**: System MUST support listing lore entries filtered by a single
  scope (P2).
- **FR-010**: System MUST distinguish human-authored origin from agent origins in
  stored metadata even though agent-authored creation is out of scope for this
  feature.
- **FR-011**: Mutating operations (create, verify) MUST be auditable with actor,
  action type, target id, and timestamp available for later inspection.

### Key Entities

- **Lore Entry**: A governed unit of engineering knowledge with statement, scope,
  origin, verification status, evidence, and timestamps.
- **Scope**: A hierarchical context boundary for the entry (organization, team,
  project, repository, or feature/task). This slice requires at least repository
  and team scope kinds.
- **Evidence Reference**: A pointer supporting the claim (path, URL, decision
  record id, ticket id, or similar), not a full document copy.
- **Actor**: The human identity performing create or verify.
- **Verification**: A recorded confirmation action that changes verification
  status without altering authorship origin.
- **Audit Record**: An immutable record of a mutating action for traceability.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can record a scoped human-authored lore entry and receive a
  stable id in under 30 seconds in a local development environment.
- **SC-002**: A user can retrieve an entry by id and see statement, scope,
  origin, verification status, and evidence with 100% field completeness for
  required attributes on successful create.
- **SC-003**: After verification, 100% of sampled entries show verified status
  plus verifier identity and time, while retaining the original statement and
  `human_authored` origin.
- **SC-004**: 100% of invalid create attempts (missing statement or scope) fail
  without creating an entry.
- **SC-005**: Listing by scope returns only in-scope entries (precision 100% in
  fixture-based acceptance tests).
- **SC-006**: Unverified entries are never labeled as verified in retrieval
  responses.

## Assumptions

- This feature is the first product vertical slice: governance-plane store,
  retrieve, and verify for human-authored lore. Semantic graph ingestion,
  conflict detection, supersession, invalidation, and context compilation are
  out of scope.
- Agent-authored or agent-inferred creation is out of scope; origin values for
  agents are reserved in the model only.
- Full enterprise authentication (OIDC) is out of scope; callers supply an
  explicit human actor identity for create/verify, validated as present and
  non-empty.
- Soft authorization (actor must be present) is sufficient for this slice;
  fine-grained RBAC and tenant isolation policies will follow in a later feature.
- At least `team` and `repository` scope kinds are supported; other hierarchy
  levels may be accepted if provided but are not required to unlock this MVP.
- Primary interaction for this slice is through MemLore’s human/automation
  interface (not requiring MCP tools yet). MCP `remember` / `get` / `verify`
  parity may follow without changing these behavioral requirements.
- Maximum statement length defaults to 8,000 characters unless later revised.
- Re-verifying an already verified entry is treated as a successful no-op that
  preserves the original verification metadata.

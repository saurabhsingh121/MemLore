# Feature Specification: Scoped Human-Authored Lore Entry

**Feature Branch**: `001-scoped-lore-entry`  
**Created**: 2026-08-25  
**Status**: Draft  
**Input**: User description: "Store, retrieve and verify a scoped human-authored Lore entry with provenance"

## Clarifications

### Session 2026-08-25

- Q: May the author verify their own lore entry in this slice? → A: Yes — self-verify is allowed (author and verifier may be the same actor).
- Q: How is scope identity represented for create and list? → A: Structured scope with `kind` + `key` (kinds include at least `team` and `repository`).
- Q: How are evidence references shaped? → A: Structured evidence items with `type` + `value` (types include at least `url`, `path`, `adr`).
- Q: How is audit inspection exposed in this slice? → A: Persist audit records on create/verify and provide a read API to list audits for a lore entry id.
- Q: Are duplicate statements allowed in the same scope? → A: Yes — duplicates allowed; identical statement + scope may create multiple entries (each with its own id).
- Q: (Analyze remediation) Unknown lore id on audit list? → A: Always not-found (404); never empty list for missing entry.

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

1. **Given** a valid scope (`kind` + `key`) and a non-empty knowledge statement,
   **When** a human actor records the lore entry as human-authored, **Then** the
   system stores it and returns a stable identifier plus provenance (actor,
   origin, created time, scope kind and key).
2. **Given** a knowledge statement with one or more structured evidence
   references (`type` + `value`, e.g. path, URL, or ADR id), **When** the entry
   is recorded, **Then** those evidence references are retained with the entry
   and visible on retrieval.
3. **Given** a missing statement, missing scope kind, or missing scope key,
   **When** a create attempt is made, **Then** the system rejects the request
   with a clear validation error and stores nothing.
4. **Given** an existing lore entry with statement S in scope X, **When** a
   caller creates another entry with the same statement S in the same scope X,
   **Then** the system creates a new entry with a distinct identifier (duplicates
   are allowed).

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

1. **Given** an unverified human-authored lore entry, **When** a human actor
   verifies it (including the original author), **Then** verification status
   becomes verified, the verifier and verification time are recorded, and the
   original statement and authorship remain unchanged.
2. **Given** an already verified lore entry, **When** a caller verifies it again,
   **Then** the system remains verified and does not invent a conflicting second
   identity for the same entry (idempotent no-op that preserves original
   verification metadata).
3. **Given** a lore entry that is not verified, **When** a caller retrieves it,
   **Then** the system reports it as unverified and does not present it as
   verified knowledge.
4. **Given** an unverified entry authored by actor A, **When** actor A verifies
   it, **Then** the entry becomes verified (self-verify is allowed in this
   slice).

---

### User Story 4 - List lore within a scope (Priority: P2)

An engineer lists lore entries for a given scope to discover what has already
been recorded.

**Why this priority**: Valuable for early usability, but store/get/verify alone
already form a complete vertical slice.

**Independent Test**: Create multiple entries in two scopes; listing one scope
returns only that scope’s entries.

**Acceptance Scenarios**:

1. **Given** lore entries in scope A and scope B (distinct `kind`+`key` pairs),
   **When** a caller lists scope A, **Then** only scope A entries are returned.
2. **Given** a scope (`kind`+`key`) with no entries, **When** listed, **Then**
   the result is an empty collection (not an error).
3. **Given** two scopes that share a key but differ in kind (or share a kind but
   differ in key), **When** listing one, **Then** entries from the other are not
   included.

---

### User Story 5 - Inspect audit trail (Priority: P1)

An engineer inspects the audit history of a lore entry to see who created and
verified it and when.

**Why this priority**: Provenance and auditability are constitutionally
first-class; create/verify without inspectable audit fails FR-level acceptance.

**Independent Test**: Create and verify an entry, then list audits by entry id
and confirm create and verify actions appear with actor and timestamps.

**Acceptance Scenarios**:

1. **Given** a lore entry that was created, **When** a caller lists audits for
   that entry id, **Then** at least one `create` audit record is returned with
   actor and timestamp.
2. **Given** a lore entry that was created and then verified, **When** audits are
   listed for that id, **Then** both `create` and `verify` actions appear in
   chronological order.
3. **Given** an unknown lore entry id, **When** audits are listed, **Then** the
   system returns not found (HTTP 404 equivalent); it MUST NOT return an empty
   audit list for a missing entry.

---

### Edge Cases

- Creating an entry that duplicates an existing statement in the same scope is
  allowed and yields a new distinct identifier.
- Creating an entry with an unsupported or blank scope `kind`, or blank scope
  `key`, is rejected.
- Evidence omitted is allowed. Evidence items with blank `type`, blank `value`,
  unsupported `type`, or empty-string fields are rejected.
- Retrieving or verifying a deleted-or-unknown id yields not-found (hard delete
  is out of scope; this slice has no delete operation).
- Verification by a missing/blank actor identity is rejected.
- Extremely long statements beyond **8,000** characters are rejected with a
  validation error.
- Scope `key` longer than **512** characters, or evidence `value` longer than
  **2,048** characters, is rejected.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow a human actor to create a lore entry with a
  non-empty statement, a scope identified by `kind` and `key`, and origin
  `human_authored`. Supported kinds for this slice MUST include at least `team`
  and `repository`.
- **FR-002**: System MUST accept optional evidence references on create, each as
  a structured `type` + `value` pair. Supported types for this slice MUST
  include at least `url`, `path`, and `adr`. Evidence is retained with the entry.
- **FR-003**: System MUST assign each lore entry a stable unique identifier.
  Identical statement text within the same scope MUST NOT prevent creation of
  additional entries (duplicates allowed; each receives its own id).
- **FR-004**: System MUST persist and return provenance for every entry: author
  actor identity, origin, created timestamp, scope `kind` and `key`, verification
  status, and evidence.
- **FR-005**: System MUST allow retrieval of a lore entry by its identifier.
- **FR-006**: System MUST allow a human actor to verify an unverified
  human-authored lore entry, recording verifier identity and verification time.
  The verifier MAY be the same actor as the author (self-verify allowed).
- **FR-007**: System MUST NOT present unverified lore as verified.
- **FR-008**: System MUST reject create/verify requests that fail validation
  (missing statement, scope kind, scope key, actor, or invalid evidence
  `type`/`value`) without partial persistence.
- **FR-009**: System MUST support listing lore entries filtered by an exact
  scope match on both `kind` and `key` (P2).
- **FR-010**: System MUST distinguish human-authored origin from agent origins in
  stored metadata even though agent-authored creation is out of scope for this
  feature.
- **FR-011**: Mutating operations (create, verify) MUST write immutable audit
  records with actor, action type, target lore entry id, and timestamp.
- **FR-012**: System MUST allow listing audit records for a lore entry by its
  identifier, ordered chronologically ascending.

### Key Entities

- **Lore Entry**: A governed unit of engineering knowledge with statement, scope,
  origin, verification status, evidence, and timestamps.
- **Scope**: A hierarchical context boundary identified by `kind` + `key`.
  Required kinds for this slice: `team` and `repository`. Optional additional
  kinds (`organization`, `project`, `feature`/`task`) may be accepted if provided
  but are not required to unlock the MVP. Equality for listing is exact match on
  both fields.
- **Evidence Reference**: A structured pointer supporting the claim, identified
  by `type` + `value` (not a full document copy). Required types for this slice:
  `url`, `path`, `adr`. Additional types may be accepted later without changing
  this MVP contract.
- **Actor**: The human identity performing create or verify.
- **Verification**: A recorded confirmation action that changes verification
  status without altering authorship origin.
- **Audit Record**: An immutable record of a mutating action (at least `create`
  and `verify`) for traceability, queryable by target lore entry id.

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
- **SC-004**: 100% of invalid create attempts (missing statement, scope kind, or
  scope key) fail without creating an entry.
- **SC-005**: Listing by exact `kind`+`key` returns only matching entries
  (precision 100% in fixture-based acceptance tests).
- **SC-006**: Unverified entries are never labeled as verified in retrieval
  responses.
- **SC-007**: After create and verify, listing audits for the entry returns both
  actions with actor and timestamp in 100% of fixture-based acceptance tests.

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
  Self-verification is allowed under this soft-auth model.
- At least `team` and `repository` scope kinds are supported via structured
  `kind` + `key` identity; other hierarchy levels may be accepted if provided
  but are not required to unlock this MVP.
- Primary interaction for this slice is through MemLore’s human/automation
  interface (not requiring MCP tools yet). MCP `remember` / `get` / `verify`
  parity may follow without changing these behavioral requirements.
- Maximum statement length is **8,000** characters; scope key max **512**;
  evidence value max **2,048** (aligned with data-model).
- Re-verifying an already verified entry is treated as a successful no-op that
  preserves the original verification metadata.
- Deduplication and conflict detection across similar statements are out of
  scope; duplicate statement+scope pairs are permitted.

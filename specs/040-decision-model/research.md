# Research: First-Class Decision Model (F040)

## Decision 1 — Dedicated `decisions` table; ADR is a read-side projection

**Decision**: Persist human-recorded Decisions in additive tables
(`decisions`, `decision_alternatives`, `decision_components`). Public id
equals the linked lore entry id. Current F032 accepted-ADR lore is **not**
copied into `decisions`. Get/list **project** current (and gettable
historical) ADR lore as Decisions with source kind `adr` and the lore id.

**Rationale**: Spec: structured fields must not live only in `statement`;
ADR-backed choices must be queryable without a second current fact.
Materializing ADR rows on ingest would reopen F032. A view-only Decision
with no table cannot store alternatives/consequences for human create.
Overloading `lore_review_decisions` would collide F035 promotion with
engineering decisions.

**Alternatives rejected**: Stuffing alternatives into lore statement;
eager ADR dual-write in F032 ingest; Decision as a SQL view only;
enriching lore_entries with JSON columns (mixes snippet and decision
schemas).

## Decision 2 — Same identity for human Decision and linked lore

**Decision**: Human create allocates one UUID used as both `decisions.id`
and `lore_entries.id`. Get-by-id, `explain`, and compile item id stay
aligned. Statement of the linked lore is the **choice** text.

**Rationale**: Spec assumption. Agents already `get`/`explain` by lore id.
Two ids would force an 11th tool or leak mapping into every packet item.

**Alternatives rejected**: Separate decision id + lore_id pointer (extra
join, two public identities); Decision without lore (breaks compile/
authority).

## Decision 3 — `NewDecisionLoreEntry` is verified human_authored

**Decision**: Add `NewDecisionLoreEntry`: origin `human_authored`,
`verification_status=verified`, VerifiedBy/VerifiedAt = creating operator /
now, evidence optional. `NewLoreEntry` remains unverified human-authored
only (remember / POST lore-entries). Do not reuse
`NewArchitectureDecisionLoreEntry` (would make human decisions look like
ADRs). Do not reuse F035 `NewHumanVerifiedLoreEntry` (that origin means
Accept-as-stated of an extract).

**Rationale**: Spec: a writer recording an explicit decision is not an
extract and is not trusted-source ADR ingest. Verified so compile does not
treat it as an unverified snippet. Distinguishable from
`architecture_decision` + `adr`.

**Alternatives rejected**: Unverified remember then admin verify (wrong
permission, two steps); origin `architecture_decision` without adr evidence
(lies about ingest); origin `human_verified` (reserved for F035 Accept).

## Decision 4 — ADR projection rules

**Decision**: A lore entry projects as a Decision iff origin is
`architecture_decision` and it has at least one `adr` evidence ref.
Choice = statement. Owner = `CreatedBy`. Date = `CreatedAt`. Question,
rationale, alternatives, consequences, components may be empty. Source
kind = `adr`. Current iff lore `IsCurrent`. Superseded-by = lore
`SupersededByID`. Do not re-parse ADR files. List-current = current
`decisions` rows whose linked lore is current **union** current projecting
ADR lore whose id is **not** already a `decisions` row.

**Rationale**: F032 already stored the trusted fact. Empty optional fields
are honest (FR-004, assumptions). If a human later supersedes an ADR
decision, the successor is a `decisions` row; the ADR lore is superseded
and drops out of list-current while remaining gettable.

**Alternatives rejected**: Re-ingest ADRs to fill alternatives; skip ADR
until a mapper feature; duplicate ADR statement into `decisions` on first
list (creates a second write of the same fact).

## Decision 5 — Supersede is a new Decision + F110 lore supersession

**Decision**: Supersede builds a new human Decision (new id) with caller
fields, `NewDecisionLoreEntry`, then `ApplySupersessionWithSuccessor` on
the predecessor lore (ADR or human). Persist successor `decisions` row;
if predecessor was a `decisions` row, set `superseded_by_id`. Predecessor
Decision remains gettable (row or ADR projection). Reject already
superseded/invalidated lore. Concurrent second supersede → validation
error (predecessor no longer current).

**Rationale**: Constitution VI / F110. Same UoW as Accept: lore + audits +
outbox `episode.ingest`. No new outbox types.

**Alternatives rejected**: In-place UPDATE of choice; invalidate on
replace; silent no-op on re-supersede.

## Decision 6 — F035 Accept is not a Decision

**Decision**: No hook from Accept/Reject into `decisions`. Observational
lore and pending queue items never project as Decisions (projection
requires `architecture_decision` + `adr`, or a `decisions` row).

**Rationale**: Spec FR-005. Promotion ≠ decision model.

**Alternatives rejected**: Auto-wrap every Accept; optional flag on Accept
in this slice (deferred).

## Decision 7 — Compile: additive `FirstClassDecision` flag, formulas frozen

**Decision**: After ranked items exist, mark `RankedItem.FirstClassDecision`
when the item id is a current Decision (human row or ADR projection).
`ClassifyItem` returns section `decisions` when that flag is set (before
heuristics). Dedup in the `decisions` section by item id. Do not add a new
section id. Do not change ranking formulas. Characterization tests:

1. Current human Decision outranks leftover unverified `repository_observation`.
2. Existing ingested-ADR vs git/PR observation test still passes.
3. The same lore id is not listed twice in `decisions`.

**Rationale**: Human decision statements may not contain keywords
(`decision`, `we chose`). ADR items already classify via origin/evidence;
the flag makes classification explicit and supports dedup. F007 freeze.

**Alternatives rejected**: New packet section `decision_records`; new
evidence type `decision` (would expand the Python-parity evidence enum);
changing origin weights.

## Decision 8 — Authz matches lore write / F114

**Decision**: Create/supersede = `PermWrite` (writer/admin) + membership.
Get/list = `PermRead` + membership. Local mode: mutating routes require
`X-Memlore-Actor`; membership off. Cross-tenant get-by-id → `404
not_found`. List of a known scope without membership → `403 forbidden`
(same as review-queue).

**Rationale**: User default. Same as F035 mutate/list.

**Alternatives rejected**: Admin-only create; reader create.

## Decision 9 — REST `/v1/decisions`; CLI `memlore decision`; MCP stays at 10

**Decision**:

- `POST /v1/decisions`
- `GET /v1/decisions/{id}`
- `GET /v1/decisions?scope_kind=&scope_key=` (current only)
- `POST /v1/decisions/{id}/supersede`

CLI: `memlore decision create|get|list|supersede` on local DSN (not HTTP
to `serve`). No 11th MCP tool. Agents see decisions via `get_for_task`
section `decisions` and `get`/`explain` of the lore id.

**Rationale**: Spec FR-011–FR-014. Repository keys must not appear in URL
paths. `get_for_task` already has a `decisions` section so a read tool is
unnecessary.

**Alternatives rejected**: 11th `memlore.get_decision`; MCP write;
overloading `GET /v1/ingest/candidates`.

## Decision 10 — Alternatives as child rows (F042-in-slice)

**Decision**: `decision_alternatives (decision_id, position, label, note)`.
Label required, note optional. Consequences are a single optional text
column on `decisions`. Affected components: `decision_components
(decision_id, position, name)`. No separate alternatives API.

**Rationale**: Spec F042-as-fields. Cheap, ordered, queryable later
without a second feature.

**Alternatives rejected**: JSON blob on `decisions` (harder to constrain
label); separate F042 service.

## Decision 11 — sqlc v1.28.0 style if generate is unavailable

**Decision**: Prefer `sqlc generate`. If the machine has no `sqlc` binary,
hand-write `internal/infrastructure/postgres/sqlc/decisions.sql.go` and
model fields to match existing v1.28.0 output (same as F035).

**Rationale**: User constraint. Must not block the slice.

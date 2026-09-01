# Data Model: First-Class Decision Model (F040)

## Decision (domain aggregate)

| Field | Notes |
|-------|-------|
| id | Same as linked lore entry id |
| scope | Repository scope (`kind=repository`) |
| question | Required on human create; may be empty on ADR projection |
| choice | Required; lore `statement` is this text |
| rationale | Optional (empty string if omitted) |
| alternatives | Ordered `DecisionAlternative` list |
| consequences | Optional text |
| owner | Required on human create; ADR projection uses lore `CreatedBy` |
| decided_at | Required; defaults to now if omitted on create |
| affected_components | Ordered names |
| evidence | From linked lore (not duplicated in `decisions`) |
| source_kind | `human` \| `adr` |
| superseded_by_id | Successor decision/lore id when superseded |
| current | Derived: not superseded and linked lore not invalidated |
| created_by / created_at | Actor and time of this version |

## DecisionAlternative

| Field | Notes |
|-------|-------|
| label | Required, non-empty after trim |
| note | Optional |

## DecisionComponent

| Field | Notes |
|-------|-------|
| name | Required, non-empty after trim |

## DecisionSourceKind

- `human` — row in `decisions` created via F040 create/supersede
- `adr` — projected from F032 lore (`architecture_decision` + `adr` evidence)

## Persistence — `decisions` (human-recorded only)

| Column | Notes |
|--------|-------|
| id | VARCHAR(36) PK, equals lore_entries.id |
| scope_kind / scope_key | Repository |
| question | TEXT NOT NULL (may be empty only if we never persist ADR here — human create rejects empty) |
| choice | TEXT NOT NULL |
| rationale | TEXT NOT NULL DEFAULT '' |
| consequences | TEXT NOT NULL DEFAULT '' |
| owner | TEXT NOT NULL |
| decided_at | TIMESTAMPTZ NOT NULL |
| source_kind | `human` (ADR not stored here) |
| superseded_by_id | VARCHAR(36) NULL |
| created_by | TEXT NOT NULL |
| created_at / updated_at | TIMESTAMPTZ |

Indexes: `(scope_kind, scope_key)` for list; optional `(superseded_by_id)`.

Do **not** FK-enforce lore_entries (same pattern as other governance tables
using VARCHAR ids). Do **not** write to `lore_review_decisions` or ingest
cursor tables.

## Persistence — `decision_alternatives`

`(decision_id, position)` PK; `label` NOT NULL; `note` NOT NULL DEFAULT ''.

## Persistence — `decision_components`

`(decision_id, position)` PK; `name` NOT NULL.

Replace children on write (delete + insert in the same UoW).

## Linked lore (human create)

| Path | Origin | Verification | Statement | Evidence | Actor |
|------|--------|--------------|-----------|----------|-------|
| Create / supersede successor | `human_authored` | `verified` | choice | caller evidence or empty | creating operator |

Constructor: `NewDecisionLoreEntry`. `NewLoreEntry` unchanged.

Create audits: `create` on the new lore id (reuse existing audit action).
Supersede: F110 dual audits (`supersede` predecessor, `create` successor)
plus successor `episode.ingest` outbox. No new outbox types.

## ADR projection (not a table row)

Eligibility:

- origin `architecture_decision`
- at least one evidence type `adr` with non-empty value

Mapped fields: see Research Decision 4. Never eligible: observational lore,
F035 accept successors unless they independently exist as `decisions` rows
(they do not, in this slice).

## Current vs history

```text
human create                         → decisions row + lore; both current
ADR ingest (F032, unchanged)         → lore only; get/list project Decision
list-current                         → current human rows (lore current)
                                       ∪ current ADR projections without a
                                       decisions row for that id
supersede human                      → new row+lore; old row superseded_by_id;
                                       old lore superseded
supersede ADR projection             → new human row+lore; ADR lore superseded;
                                       no decisions row for the ADR id
get superseded ADR id                → projected Decision, current=false
invalidate linked lore (existing)    → Decision not current
F035 Accept / git / PR               → not a Decision
```

## Validation

- Scope kind MUST be `repository` on create/list.
- Human create: question, choice, owner, actor required (non-empty trim).
- Alternative labels and component names required if the element is present.
- Choice length ≤ `MaxStatementLength` (it is the lore statement).
- Question/rationale/consequences: reasonable text; trim; empty rationale/consequences allowed.
- Mutate requires non-empty actor.
- Predecessor MUST be current (human row or projecting ADR lore).
- Already superseded or invalidated → `validation_error`.
- Unknown id / unauthorized get → `not_found`.
- Non-repository list/create → `validation_error`.

## Relationships

- Decision.id = LoreEntry.id (human) or LoreEntry.id (ADR projection).
- Compile continues to rank lore; `FirstClassDecision` is a classification hint.
- F032 ADR ingest tables unchanged.
- F035 `lore_review_decisions` unchanged.

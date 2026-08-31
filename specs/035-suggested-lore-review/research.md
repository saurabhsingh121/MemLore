# Research: Suggested Lore Review Queue (F035)

## Decision 1 — Supersede on Accept; never mutate origin in place

**Decision**: Accept creates a new current lore entry and supersedes the
observational predecessor via `ApplySupersessionWithSuccessor`. The
predecessor row keeps origin `repository_observation`. `ApplyVerification` is
not Accept and is not called as the promotion path.

**Rationale**: Constitution VI — do not overwrite historical observational
truth. Constitution V — do not pretend an extract was always human-side.
Existing `verify` only flips `verification_status`; origin stays observational.
F110 already models successor + predecessor pointer.

**Alternatives rejected**: In-place origin rewrite; Accept implemented as
`verify`; invalidate the observation on Accept (would erase capture history).

## Decision 2 — human_verified constructor for Accept-as-stated

**Decision**: Add `NewHumanVerifiedLoreEntry`. It requires origin
`human_verified` (default if empty), non-empty evidence (copied from the
predecessor), and sets `verification_status=verified` with
`VerifiedBy`/`VerifiedAt` = reviewer / now. `NewLoreEntry` remains
human-authored only and must not become the Accept-as-stated path (it would
mislabel origin and, if evidence were omitted, drop provenance).

Edit-then-Accept (normalized statement differs) builds a **verified**
`human_authored` successor with the same evidence copy — a dedicated helper
that does not go through unverified `NewLoreEntry` then a separate verify
command. Direct `remember` / `POST /v1/lore-entries` stay unverified
`human_authored` until explicit verify.

**Rationale**: `KnowledgeOriginHumanVerified` exists in F003 with human-side
weight but had no writer. Accept-as-stated is that writer. Edit means the
human supplied the trusted wording (`human_authored`).

**Alternatives rejected**: Reuse `NewLoreEntry` for as-stated (wrong origin);
reuse `NewArchitectureDecisionLoreEntry` (would make git/PR look like ADRs);
born-unverified successor (queue Accept must be verified).

## Decision 3 — Queue is a projection + decision overlay

**Decision**: Pending items **are** current observational lore entries (git
`commit` or PR `pr` evidence, origin `repository_observation`, not superseded,
not invalidated). Table `lore_review_decisions` stores Accept/Reject keyed by
extract identity. Listing pending = lore in scope filtered by eligibility
MINUS identities with a decision (or hide accepted because they are already
non-current). Ingest candidates listing is unchanged.

**Rationale**: F030/F031 already persist candidates as lore so compile can see
them. A parallel `suggested_lore` knowledge table would be a second plane.
Overloading `GET /v1/ingest/candidates` would mix F032 trusted ADRs and
review lifecycle. Additive overlay avoids reopening ingest cursor tables.

**Alternatives rejected**: Eager enqueue rows on F030/F031 ingest (reopens
producers); treat ingest candidates as the review UI; in-lore review_status
column (collides with verification_status and ADR listing).

## Decision 4 — Extract identity for Reject idempotency

**Decision**: Key = `scope_kind` + `scope_key` + primary evidence type +
primary evidence value + SHA-256 hex of `strings.TrimSpace(statement)`.
Primary evidence is the first `commit` evidence if present, else the first
`pr` evidence. Re-ingest of the same SHA/PR already no-ops at F030/F031;
the decision key additionally prevents a future same extract from becoming
pending.

**Rationale**: Spec: “scope + evidence identity + checksum/statement hash.”
Whitespace-normalized statement distinguishes Accept-as-stated from Edit
without lowercasing (code/identifiers are case-sensitive).

**Alternatives rejected**: Key on lore entry id only (re-created lore with the
same extract could resurrect); key on evidence only (a later different
statement for the same SHA could never be reviewed — F030 will not re-extract
anyway, but identity should follow the extract).

## Decision 5 — Reject does not invalidate observational lore

**Decision**: Reject inserts/updates a `rejected` decision. Observational lore
remains current and observational. It may still appear in compile at
observational rank. It MUST NOT appear in the pending queue.

**Rationale**: Reject means “do not promote,” not “this commit never said
that.” Invalidating would mix F110 invalidation (knowledge is wrong) with
review (humans declined to trust the extract).

**Alternatives rejected**: Invalidate on reject; delete lore; hide rejected
observations from compile in this slice (F007 freeze).

## Decision 6 — Confidence and reason omitted

**Decision**: Queue JSON/CLI omit `confidence` and `reason` in v1. Do not
persist additive columns until a producer supplies them. Do not default to
`0.84` or derive a fake reason from the statement.

**Rationale**: F030/F031 store the statement only. Invented scores violate
constitution V (false precision) and IX (generic decoration).

**Alternatives rejected**: Hard-code 0.5; duplicate statement as reason.

## Decision 7 — Authz: write + membership, not PermVerify

**Decision**: List = `PermRead` + F114 membership. Accept/Reject = `PermWrite`
(writer or admin) + membership. Same as ingest trigger and supersede. Do not
require `PermVerify` (admin-only). Local mode: actor header/flag; membership
off.

**Rationale**: User default. Verify remaining admin-only is an F010 contract
this slice does not reopen. Accept is promotion with a new row (writer-class),
not the verify-in-place permission.

**Alternatives rejected**: Admin-only Accept (would stall the flywheel);
reader Accept.

## Decision 8 — REST without repository key in the URL

**Decision**:

- `GET /v1/review-queue?scope_kind=&scope_key=` (optional `status=pending`, default pending)
- `GET /v1/review-queue/{id}` where `{id}` is the observational lore entry id
- `POST /v1/review-queue/{id}/accept` body `{ "statement": "…" }` optional
- `POST /v1/review-queue/{id}/reject`

**Rationale**: Same slash-in-key constraint as F030–F032. Item id in the path
matches lore-entry mutate routes.

**Alternatives rejected**: `/v1/repositories/{key}/review`; overloading
`/v1/ingest/candidates/{id}/accept`.

## Decision 9 — CLI `memlore review`; MCP stays at 10

**Decision**: `memlore review list|accept|reject` on the local DSN (same as
ingest). No 11th MCP tool. Agents keep search/get/explain. Mutating review is
human CLI/REST.

**Rationale**: User: prefer 10 tools unless agents cannot work without listing
candidates. They can already see observational lore via search/get; they must
not Accept.

**Alternatives rejected**: `memlore.review_list` 11th tool; MCP accept/reject.

## Decision 10 — Uncertain ADRs stay skipped; F032 unchanged

**Decision**: Do not enqueue F032 skipped drafts/uncertain parses. Do not
change ADR ingest. Queue identity is producer-agnostic so F033/F034 can reuse
it later (new evidence types would be additive).

**Rationale**: Spec: optional uncertain ADRs are out of this slice; do not
undo trusted-source auto-verify for accepted ADRs.

**Alternatives rejected**: Change F032 skip → pending; auto-upgrade git/PR
because ADRs are trusted.

## Decision 11 — Compile characterization only

**Decision**: Do not change F007 ranking formulas. Add tests:

1. `human_verified` + verified + commit/pr evidence outranks unverified
   `repository_observation`.
2. Existing `TestCompileContextIngestedADROutranksGitAndPRObservation` still
   passes (F032 ADR still first vs leftover observations).

**Rationale**: F003 already gives `human_verified` the same origin adjustment
as `architecture_decision`, but ADR **evidence** strength remains higher, so
accepted git/PR lore will not outrank accepted ADRs. That is correct.

**Alternatives rejected**: Tweaking origin weights; boosting commit evidence
to ADR strength.

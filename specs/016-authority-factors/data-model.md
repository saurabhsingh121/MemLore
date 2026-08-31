# Data Model: F003 authority evaluation

No new persisted tables. Inputs are existing `LoreEntry` fields and
`GraphFact.Score`.

## Authority Evaluation (ephemeral)

| Field | Type | Notes |
|-------|------|-------|
| score | float 0–1 | Derived from factors; used for ranking |
| trust_band | enum | `canonical` \| `high` \| `medium` \| `low` \| `untrusted` |
| factors | FactorSet | Always present with the score |
| factor_breakdown | string[] | Short `key=value` lines for explain |

Never stored. Never returned as score-only.

## FactorSet

| Factor | Governance | Graph | Notes |
|--------|------------|-------|-------|
| origin | LoreEntry.Origin | omit | |
| verification_status | LoreEntry.VerificationStatus | omit | |
| supersession_status | `current` or `superseded` | omit | Invalidated is not a supersession value |
| recency_boost | 0.00–0.10 from CreatedAt | omit | Same decay as F109 |
| evidence_strength | 1.0 / 0.6 / 0.0 | omit | ADR / any / none |
| source_type | derived | `graph` | See spec FR-004 |
| scope_match | 1.0 / 0.5 / 0.0 | 1.0 if fact scope matches request else 0.0 | |
| graph_score | omit | GraphFact.Score | Raw graph score, not the capped ranking score |
| source_reliability | omit v1 | omit | Deferred |

## TrustBand assignment

First-match (spec FR-008). Independent of numeric score.

| Band | Meaning |
|------|---------|
| canonical | Current verified ADR-class human-side knowledge |
| high | Current verified, weaker/no evidence; or verified agent origin |
| medium | Current unverified human-side; or superseded (capped) |
| low | Graph-only; unverified agent; unverified repo observation |
| untrusted | Invalidated |

## Context item additions

| Field | Type |
|-------|------|
| trust_band | string (required on each item) |
| authority_factors | FactorSet (existing object, richer keys, omitempty) |
| authority_score | float (existing) |

## Explain Result additions

Existing LoreEntry fields + audits, plus evaluation fields above and
`factor_breakdown`.

## Unchanged

- LoreEntry persistence and F110 transitions
- Audit records
- Outbox
- Graph-service fact model
- ConflictGroup

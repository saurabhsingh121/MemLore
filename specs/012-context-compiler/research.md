# Research: Context Compiler (F109)

## Authority scoring v1

**Decision**: Weighted score 0–1 from verification, origin, recency (governance)
or graph relevance score (graph).

| Source | Base score | Boost |
|--------|------------|-------|
| Governance verified | 0.85 | +0.10 recency (newer = higher) |
| Governance unverified | 0.55 | +0.10 recency |
| Graph | graph `score` × 0.80 | — |

Verified governance always beats graph-only at equal recency.

## Dedup v1

**Decision**: Normalize statement (`strings.ToLower` + trim); skip graph hit if
governance statement already present.

Cross-plane ID dedup deferred (different id spaces).

## Token budgeting

**Decision**: Estimate `(len(statement) + 80) / 4` tokens per item; default
budget 4096; pack in authority rank order.

**Alternatives**: tiktoken (deferred — no new deps in v1).

## Retrieval limit

**Decision**: Pass `limit: 20` to SearchKnowledge; compiler trims to token budget.

## REST path

**Decision**: `POST /v1/context/compile` (not `/v1/get-for-task`).

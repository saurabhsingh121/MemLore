# Authority model

Authority in MemLore is **explainable**. It is derived from explicit factors,
not a single opaque magic score.

## Factors (baseline)

| Factor | Role |
|--------|------|
| source_type | ADR, human statement, agent observation, repo observation, import, … |
| origin | human_authored, human_verified, agent_observation, agent_inference, … |
| verification_status | unverified, verified, rejected, invalidated |
| evidence_strength | presence/quality of linked evidence |
| recency | freshness relative to the task |
| scope_match | how well the fact matches requested scopes |
| supersession_status | current vs superseded |
| source_reliability | historical reliability of the source |

A ranking score may be computed from these factors, but the factors MUST remain
available so every important retrieval can answer:

- Why was this returned?
- Who said it?
- Where did it come from?
- What evidence supports it?
- How authoritative is it?
- Is it current?
- Has something superseded it?

## Hard rules

- Agent inference MUST NOT silently gain human authority.
- Repository evidence is strong observational evidence, not automatic intent.
- Conflicts are preserved and surfaced.

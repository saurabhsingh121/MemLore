# Lore

**Lore** is MemLore's name for preserved engineering knowledge: decisions,
conventions, observations, and context that help humans and agents understand
*why* the system is the way it is.

A **Lore Entry** is a governed unit of knowledge with statement, scope
(`kind` + `key`), origin, verification status, evidence (`type` + `value`), and
timestamps. In the first product slice, human-authored entries can be created,
retrieved, verified (including self-verify), listed by scope, and audited.
Duplicates in the same scope are allowed.

Automatically extracted knowledge (git, PRs, ADRs, agent sessions) is
**candidate lore** until a human accepts it. Agent observations and inferences
MUST remain labeled until verified. Engineering **decisions** are planned as a
dedicated domain concept (F040), not only generic memories. See
[FEATURE_DEVELOPMENT.md](../development/FEATURE_DEVELOPMENT.md).

# MCP API (planned)

MemLore exposes a **domain** MCP surface. Graphiti internals are not the
product contract.

| Tool | Purpose |
|------|---------|
| `memlore.get_for_task` | Preferred agent entry: compiled context packet |
| `memlore.search` | Scoped search |
| `memlore.remember` | Store knowledge with provenance |
| `memlore.get` | Fetch by id |
| `memlore.verify` | Human/policy verification |
| `memlore.supersede` | Replace while preserving history |
| `memlore.invalidate` | Mark invalid without deleting evidence |
| `memlore.explain` | Provenance, authority, evidence |

Status: **not implemented** in bootstrap. Contract tests will land with the
first MCP feature.

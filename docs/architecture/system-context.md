# System context

MemLore sits between engineering work and coding agents.

**Users**

- Software engineers and reviewers (web UI, CLI, IDE)
- Coding agents (MCP-compatible: Cursor, Claude Code, Codex, others)
- Optional source integrations (GitHub/GitLab, Jira, Confluence, CI, ADR repos)

**Primary interactions**

1. Humans and agents **remember** engineering knowledge with provenance.
2. Agents call **get_for_task** / **search** to receive a compiled context packet.
3. Humans **verify**, **supersede**, or **invalidate** knowledge.
4. The system surfaces **conflicts** and **stale** knowledge rather than hiding them.

MemLore is **not** a replacement for Git, issue trackers, wiki systems, or an
autonomous coding agent runtime.

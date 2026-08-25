from __future__ import annotations

import logging
import sys

from mcp.server import MCPServer

from memlore.adapters.mcp.tools import (
    explain_lore,
    get_lore,
    remember_lore,
    search_lore,
    verify_lore,
)
from memlore.adapters.rest.schemas import EvidenceIn, ScopeIn
from memlore.bootstrap.container import AppContainer


def configure_mcp_logging() -> None:
    """Send MemLore logs to stderr so stdout stays MCP-protocol-only."""

    logger = logging.getLogger("memlore")
    logger.handlers.clear()
    handler = logging.StreamHandler(sys.stderr)
    handler.setFormatter(
        logging.Formatter("%(asctime)s %(levelname)s %(name)s %(message)s")
    )
    logger.addHandler(handler)
    logger.setLevel(logging.INFO)
    logger.propagate = False


def create_mcp_server(container: AppContainer) -> MCPServer:
    server = MCPServer("memlore")

    @server.tool(name="memlore.remember")
    def remember(
        statement: str,
        scope: ScopeIn,
        actor_id: str,
        evidence: list[EvidenceIn] | None = None,
    ) -> dict[str, object]:
        """Store a human-authored scoped lore entry."""

        return remember_lore(
            container,
            statement=statement,
            scope=scope,
            actor_id=actor_id,
            evidence=evidence,
        ).model_dump(mode="json")

    @server.tool(name="memlore.get")
    def get(id: str) -> dict[str, object]:
        """Fetch a lore entry by id."""

        return get_lore(container, id=id).model_dump(mode="json")

    @server.tool(name="memlore.verify")
    def verify(id: str, actor_id: str) -> dict[str, object]:
        """Verify a lore entry (idempotent)."""

        return verify_lore(container, id=id, actor_id=actor_id).model_dump(mode="json")

    @server.tool(name="memlore.explain")
    def explain(id: str) -> dict[str, object]:
        """Return lore entry fields plus chronological audits."""

        return explain_lore(container, id=id).model_dump(mode="json")

    @server.tool(name="memlore.search")
    def search(scope: ScopeIn) -> dict[str, object]:
        """List lore entries by exact scope kind and key."""

        return search_lore(container, scope=scope).model_dump(mode="json")

    return server

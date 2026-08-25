from __future__ import annotations

import asyncio
import logging
from datetime import UTC, datetime

import pytest
from tests.support.fakes import InMemoryUnitOfWork

from memlore.adapters.mcp.server import configure_mcp_logging, create_mcp_server
from memlore.bootstrap.container import build_memory_container
from memlore.infrastructure.clock import FixedClock


def test_create_mcp_server_is_named_memlore_without_graphiti_tools() -> None:
    container = build_memory_container(
        InMemoryUnitOfWork(),
        clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC)),
    )
    server = create_mcp_server(container)
    assert server.name == "memlore"
    tools = asyncio.run(server.list_tools())
    names = [tool.name for tool in tools]
    joined = " ".join(names).lower()
    assert "graphiti" not in joined
    assert "neo4j" not in joined


def test_configure_mcp_logging_writes_to_stderr(
    capsys: pytest.CaptureFixture[str],
) -> None:
    logger = logging.getLogger("memlore")
    previous_handlers = list(logger.handlers)
    previous_propagate = logger.propagate
    try:
        configure_mcp_logging()
        logger.info("mcp-log-probe")
        captured = capsys.readouterr()
        assert "mcp-log-probe" in captured.err
        assert "mcp-log-probe" not in captured.out
    finally:
        logger.handlers[:] = previous_handlers
        logger.propagate = previous_propagate

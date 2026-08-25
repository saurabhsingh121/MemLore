from __future__ import annotations

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from datetime import UTC, datetime

from mcp import Client
from mcp.types import CallToolResult

from memlore.adapters.mcp.server import create_mcp_server
from memlore.bootstrap.container import build_memory_container
from memlore.infrastructure.clock import FixedClock
from tests.support.fakes import InMemoryUnitOfWork


def tool_text(result: CallToolResult) -> str:
    return "\n".join(
        block.text for block in result.content if getattr(block, "text", None)
    )


@asynccontextmanager
async def mcp_session() -> AsyncIterator[tuple[Client, InMemoryUnitOfWork]]:
    uow = InMemoryUnitOfWork()
    container = build_memory_container(
        uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )
    server = create_mcp_server(container)
    async with Client(server) as client:
        yield client, uow

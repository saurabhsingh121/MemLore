from __future__ import annotations

from typing import NoReturn

from mcp.server.mcpserver.exceptions import ToolError

from memlore.domain.exceptions import NotFoundError, ValidationError


def raise_as_tool_error(exc: BaseException) -> NoReturn:
    """Map domain errors to agent-visible MCP ToolError codes."""

    if isinstance(exc, ValidationError):
        raise ToolError(f"validation_error: {exc}") from exc
    if isinstance(exc, NotFoundError):
        raise ToolError(f"not_found: {exc}") from exc
    raise exc

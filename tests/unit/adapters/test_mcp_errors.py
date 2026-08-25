from __future__ import annotations

import pytest
from mcp.server.mcpserver.exceptions import ToolError

from memlore.adapters.mcp.errors import raise_as_tool_error
from memlore.domain.exceptions import NotFoundError, ValidationError


def test_validation_error_becomes_tool_error_with_code() -> None:
    with pytest.raises(ToolError) as exc_info:
        raise_as_tool_error(ValidationError("actor must be non-empty"))
    assert str(exc_info.value).startswith("validation_error:")
    assert "actor must be non-empty" in str(exc_info.value)
    assert "Traceback" not in str(exc_info.value)


def test_not_found_error_becomes_tool_error_with_code() -> None:
    with pytest.raises(ToolError) as exc_info:
        raise_as_tool_error(NotFoundError("lore entry missing"))
    assert str(exc_info.value).startswith("not_found:")
    assert "lore entry missing" in str(exc_info.value)
    assert "sql" not in str(exc_info.value).lower()


def test_other_exceptions_are_not_rewritten() -> None:
    with pytest.raises(RuntimeError, match="boom"):
        raise_as_tool_error(RuntimeError("boom"))

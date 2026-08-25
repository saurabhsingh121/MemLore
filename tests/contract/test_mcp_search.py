from __future__ import annotations

from tests.support.mcp_client import mcp_session, tool_text


async def test_search_exact_scope_empty_and_incomplete() -> None:
    async with mcp_session() as (mcp_client, _uow):
        await mcp_client.call_tool(
            "memlore.remember",
            {
                "statement": "Alpha rule",
                "scope": {"kind": "repository", "key": "alpha"},
                "actor_id": "alice",
            },
        )
        await mcp_client.call_tool(
            "memlore.remember",
            {
                "statement": "Beta rule",
                "scope": {"kind": "repository", "key": "beta"},
                "actor_id": "alice",
            },
        )

        found = await mcp_client.call_tool(
            "memlore.search", {"scope": {"kind": "repository", "key": "alpha"}}
        )
        assert found.is_error is False
        assert found.structured_content is not None
        statements = [item["statement"] for item in found.structured_content["items"]]
        assert statements == ["Alpha rule"]

        empty = await mcp_client.call_tool(
            "memlore.search", {"scope": {"kind": "team", "key": "nobody"}}
        )
        assert empty.is_error is False
        assert empty.structured_content is not None
        assert empty.structured_content["items"] == []

        incomplete = await mcp_client.call_tool(
            "memlore.search", {"scope": {"kind": "team"}}
        )
        assert incomplete.is_error is True
        text = tool_text(incomplete)
        assert "validation_error:" in text or "required" in text.lower()

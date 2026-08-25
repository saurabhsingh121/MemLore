from __future__ import annotations

from tests.support.mcp_client import mcp_session, tool_text


async def test_get_and_explain_existing_and_unknown() -> None:
    async with mcp_session() as (mcp_client, _uow):
        created = await mcp_client.call_tool(
            "memlore.remember",
            {
                "statement": "Prefer explicit actor_id",
                "scope": {"kind": "repository", "key": "r1"},
                "actor_id": "alice",
            },
        )
        assert created.structured_content is not None
        entry_id = created.structured_content["id"]

        got = await mcp_client.call_tool("memlore.get", {"id": entry_id})
        assert got.is_error is False
        assert got.structured_content is not None
        assert got.structured_content["id"] == entry_id
        assert got.structured_content["origin"] == "human_authored"
        assert got.structured_content["created_by"] == "alice"

        explained = await mcp_client.call_tool("memlore.explain", {"id": entry_id})
        assert explained.is_error is False
        assert explained.structured_content is not None
        assert "summary" not in explained.structured_content
        assert explained.structured_content["statement"] == "Prefer explicit actor_id"
        actions = [item["action"] for item in explained.structured_content["audits"]]
        assert actions == ["create"]

        missing = "00000000-0000-0000-0000-000000000000"
        get_missing = await mcp_client.call_tool("memlore.get", {"id": missing})
        assert get_missing.is_error is True
        assert "not_found:" in tool_text(get_missing)

        explain_missing = await mcp_client.call_tool("memlore.explain", {"id": missing})
        assert explain_missing.is_error is True
        assert "not_found:" in tool_text(explain_missing)
        assert "Traceback" not in tool_text(explain_missing)

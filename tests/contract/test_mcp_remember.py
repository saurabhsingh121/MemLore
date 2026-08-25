from __future__ import annotations

from tests.support.mcp_client import mcp_session, tool_text


async def test_remember_success_and_duplicate() -> None:
    async with mcp_session() as (mcp_client, _uow):
        args = {
            "statement": "Payment events must use the transactional outbox.",
            "scope": {"kind": "repository", "key": "github.com/acme/payments"},
            "actor_id": "alice",
            "evidence": [{"type": "adr", "value": "0001-dual-plane"}],
        }
        first = await mcp_client.call_tool("memlore.remember", args)
        assert first.is_error is False
        assert first.structured_content is not None
        payload = first.structured_content
        assert payload["origin"] == "human_authored"
        assert payload["statement"] == args["statement"]
        assert payload["created_by"] == "alice"
        assert payload["verification_status"] == "unverified"

        second = await mcp_client.call_tool("memlore.remember", args)
        assert second.is_error is False
        assert second.structured_content is not None
        assert second.structured_content["id"] != payload["id"]


async def test_remember_missing_or_blank_actor_is_validation_error() -> None:
    async with mcp_session() as (mcp_client, mcp_uow):
        missing = await mcp_client.call_tool(
            "memlore.remember",
            {
                "statement": "Rule",
                "scope": {"kind": "team", "key": "t1"},
            },
        )
        assert missing.is_error is True
        assert mcp_uow.lore_entries._items == {}

        blank = await mcp_client.call_tool(
            "memlore.remember",
            {
                "statement": "Rule",
                "scope": {"kind": "team", "key": "t1"},
                "actor_id": "  ",
            },
        )
        assert blank.is_error is True
        assert "validation_error:" in tool_text(blank)
        assert mcp_uow.lore_entries._items == {}


async def test_remember_invalid_scope_is_validation_error() -> None:
    async with mcp_session() as (mcp_client, mcp_uow):
        result = await mcp_client.call_tool(
            "memlore.remember",
            {
                "statement": "Rule",
                "scope": {"kind": "not-a-scope", "key": "t1"},
                "actor_id": "alice",
            },
        )
        assert result.is_error is True
        assert mcp_uow.lore_entries._items == {}

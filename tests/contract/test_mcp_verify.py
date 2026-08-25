from __future__ import annotations

from tests.support.mcp_client import mcp_session, tool_text


async def test_verify_success_idempotent_and_errors() -> None:
    async with mcp_session() as (mcp_client, _uow):
        created = await mcp_client.call_tool(
            "memlore.remember",
            {
                "statement": "Rule",
                "scope": {"kind": "team", "key": "core"},
                "actor_id": "alice",
            },
        )
        assert created.structured_content is not None
        entry_id = created.structured_content["id"]

        verified = await mcp_client.call_tool(
            "memlore.verify", {"id": entry_id, "actor_id": "alice"}
        )
        assert verified.is_error is False
        assert verified.structured_content is not None
        assert verified.structured_content["verification_status"] == "verified"
        assert verified.structured_content["origin"] == "human_authored"
        assert verified.structured_content["verified_by"] == "alice"

        again = await mcp_client.call_tool(
            "memlore.verify", {"id": entry_id, "actor_id": "bob"}
        )
        assert again.is_error is False
        assert again.structured_content is not None
        assert again.structured_content["verified_by"] == "alice"

        blank = await mcp_client.call_tool(
            "memlore.verify", {"id": entry_id, "actor_id": ""}
        )
        assert blank.is_error is True
        assert "validation_error:" in tool_text(blank)

        missing = await mcp_client.call_tool(
            "memlore.verify",
            {
                "id": "00000000-0000-0000-0000-000000000000",
                "actor_id": "alice",
            },
        )
        assert missing.is_error is True
        assert "not_found:" in tool_text(missing)

from __future__ import annotations

from tests.support.mcp_client import mcp_session


async def test_list_tools_is_exactly_the_five_lore_tools() -> None:
    async with mcp_session() as (mcp_client, _uow):
        listed = await mcp_client.list_tools()
        names = [tool.name for tool in listed.tools]
        assert set(names) == {
            "memlore.remember",
            "memlore.get",
            "memlore.verify",
            "memlore.explain",
            "memlore.search",
        }
        assert len(names) == 5
        joined = " ".join(names).lower()
        assert "graphiti" not in joined
        assert "neo4j" not in joined
        assert "get_for_task" not in joined
        assert "supersede" not in joined
        assert "invalidate" not in joined

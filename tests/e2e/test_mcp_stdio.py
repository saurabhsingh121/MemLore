from __future__ import annotations

import os
from pathlib import Path
from uuid import uuid4

import pytest
from mcp import Client
from mcp.client.stdio import StdioServerParameters

from memlore.infrastructure.postgres.models import Base
from memlore.infrastructure.postgres.session import create_db_engine

REPO_ROOT = Path(__file__).resolve().parents[2]


@pytest.mark.integration
@pytest.mark.e2e
async def test_stdio_mcp_five_tool_path(
    require_postgres: None, database_url: str
) -> None:
    engine = create_db_engine(database_url)
    Base.metadata.create_all(engine)
    engine.dispose()

    env = {key: value for key, value in os.environ.items() if value is not None}
    env["MEMLORE_POSTGRES_DSN"] = database_url
    scope_key = f"e2e-{uuid4()}"
    params = StdioServerParameters(
        command="uv",
        args=["run", "memlore", "mcp"],
        cwd=str(REPO_ROOT),
        env=env,
    )
    async with Client(params) as client:
        listed = await client.list_tools()
        names = {tool.name for tool in listed.tools}
        assert names == {
            "memlore.remember",
            "memlore.get",
            "memlore.verify",
            "memlore.explain",
            "memlore.search",
        }

        remembered = await client.call_tool(
            "memlore.remember",
            {
                "statement": "E2E: use memlore mcp over stdio.",
                "scope": {"kind": "repository", "key": scope_key},
                "actor_id": "e2e-actor",
            },
        )
        assert remembered.is_error is False
        assert remembered.structured_content is not None
        entry_id = remembered.structured_content["id"]

        got = await client.call_tool("memlore.get", {"id": entry_id})
        assert got.is_error is False
        assert got.structured_content is not None
        assert got.structured_content["id"] == entry_id

        verified = await client.call_tool(
            "memlore.verify", {"id": entry_id, "actor_id": "e2e-actor"}
        )
        assert verified.is_error is False
        assert verified.structured_content is not None
        assert verified.structured_content["verification_status"] == "verified"

        explained = await client.call_tool("memlore.explain", {"id": entry_id})
        assert explained.is_error is False
        assert explained.structured_content is not None
        actions = [item["action"] for item in explained.structured_content["audits"]]
        assert actions.count("create") == 1
        assert actions.count("verify") == 1

        found = await client.call_tool(
            "memlore.search",
            {"scope": {"kind": "repository", "key": scope_key}},
        )
        assert found.is_error is False
        assert found.structured_content is not None
        ids = [item["id"] for item in found.structured_content["items"]]
        assert entry_id in ids

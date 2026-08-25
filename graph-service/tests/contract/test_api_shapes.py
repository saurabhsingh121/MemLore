from __future__ import annotations

from typing import Any

import pytest
from httpx import ASGITransport, AsyncClient

from graph_service.adapters.http.routes import create_app
from graph_service.bootstrap.settings import Settings

FORBIDDEN_KEYS = frozenset(
    {"EntityEdge", "group_id", "fact_embedding", "episodes", "uuid"}
)


@pytest.fixture
def contract_app(mock_knowledge_graph: Any) -> Any:
    settings = Settings(openai_api_key="test-key")
    app = create_app(settings)
    app.state.knowledge_graph = mock_knowledge_graph
    return app


@pytest.fixture
async def contract_client(contract_app: Any) -> AsyncClient:
    transport = ASGITransport(app=contract_app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac


def _walk_keys(obj: Any) -> set[str]:
    keys: set[str] = set()
    if isinstance(obj, dict):
        for key, value in obj.items():
            keys.add(str(key))
            keys.update(_walk_keys(value))
    elif isinstance(obj, list):
        for item in obj:
            keys.update(_walk_keys(item))
    return keys


@pytest.mark.asyncio
async def test_search_response_has_no_graphiti_field_names(
    contract_client: AsyncClient,
) -> None:
    await contract_client.post(
        "/episodes",
        json={
            "statement": "Outbox pattern for payments.",
            "scope": {"kind": "repository", "key": "github.com/acme/payments"},
        },
    )
    response = await contract_client.post("/search", json={"query": "outbox"})
    assert response.status_code == 200
    keys = _walk_keys(response.json())
    leaked = keys.intersection(FORBIDDEN_KEYS)
    assert not leaked, f"Graphiti fields leaked: {leaked}"

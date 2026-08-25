from __future__ import annotations

from typing import Any

import pytest
from httpx import ASGITransport, AsyncClient

from graph_service.adapters.http.routes import create_app
from graph_service.bootstrap.settings import Settings


@pytest.fixture
def test_app(mock_knowledge_graph: Any) -> Any:
    settings = Settings(
        neo4j_uri="bolt://localhost:7687",
        neo4j_user="neo4j",
        neo4j_password="test",
        openai_api_key="test-key",
    )
    app = create_app(settings)
    app.state.knowledge_graph = mock_knowledge_graph
    return app


@pytest.fixture
async def client(test_app: Any) -> AsyncClient:
    transport = ASGITransport(app=test_app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac


@pytest.mark.asyncio
async def test_health_ok(client: AsyncClient, mock_knowledge_graph: Any) -> None:
    mock_knowledge_graph.neo4j_ok = True
    response = await client.get("/health")
    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    assert body["neo4j"] == "ok"
    assert body["service"] == "graph-service"


@pytest.mark.asyncio
async def test_create_episode_returns_id(client: AsyncClient) -> None:
    payload = {
        "statement": "Payment events must use the transactional outbox.",
        "scope": {"kind": "repository", "key": "github.com/acme/payments"},
        "provenance_refs": ["lore-1"],
    }
    response = await client.post("/episodes", json=payload)
    assert response.status_code == 201
    body = response.json()
    assert body["episode_id"].startswith("ep-")


@pytest.mark.asyncio
async def test_search_returns_memlore_shape(client: AsyncClient) -> None:
    await client.post(
        "/episodes",
        json={
            "statement": "Use outbox for payment events.",
            "scope": {"kind": "repository", "key": "github.com/acme/payments"},
        },
    )
    response = await client.post("/search", json={"query": "outbox"})
    assert response.status_code == 200
    body = response.json()
    assert "results" in body
    assert len(body["results"]) >= 1
    item = body["results"][0]
    assert set(item.keys()) == {
        "id",
        "statement",
        "score",
        "scope",
        "provenance_refs",
    }
    assert "group_id" not in item
    assert "EntityEdge" not in body


@pytest.mark.asyncio
async def test_get_fact_not_found(client: AsyncClient) -> None:
    response = await client.get("/facts/missing-id")
    assert response.status_code == 404

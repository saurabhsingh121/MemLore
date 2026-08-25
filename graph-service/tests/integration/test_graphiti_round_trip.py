from __future__ import annotations

import pytest
from httpx import ASGITransport, AsyncClient

from graph_service.adapters.graphiti import GraphitiKnowledgeGraph
from graph_service.adapters.http.routes import create_app
from graph_service.bootstrap.settings import Settings


@pytest.mark.integration
@pytest.mark.asyncio
async def test_episode_ingest_and_search_round_trip(
    neo4j_settings: dict[str, str],
    require_graph_stack: None,
) -> None:
    settings = Settings(
        neo4j_uri=neo4j_settings["uri"],
        neo4j_user=neo4j_settings["user"],
        neo4j_password=neo4j_settings["password"],
    )
    app = create_app(settings)
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        health = await client.get("/health")
        assert health.status_code == 200
        assert health.json()["neo4j"] == "ok"

        statement = "MemLore integration test: transactional outbox for payments."
        create = await client.post(
            "/episodes",
            json={
                "statement": statement,
                "scope": {
                    "kind": "repository",
                    "key": "github.com/acme/payments-integration",
                },
            },
        )
        assert create.status_code == 201
        episode_id = create.json()["episode_id"]
        assert episode_id

        search = await client.post(
            "/search",
            json={"query": "transactional outbox payments"},
        )
        assert search.status_code == 200
        results = search.json()["results"]
        assert isinstance(results, list)

    graph = app.state.knowledge_graph
    if isinstance(graph, GraphitiKnowledgeGraph):
        await graph.close()

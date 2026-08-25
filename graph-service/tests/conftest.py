from __future__ import annotations

import os
from typing import Any

import pytest


def pytest_configure(config: pytest.Config) -> None:
    config.addinivalue_line(
        "markers",
        "integration: tests that require Neo4j and OPENAI_API_KEY",
    )


@pytest.fixture
def neo4j_settings() -> dict[str, str]:
    return {
        "uri": os.environ.get("MEMLORE_NEO4J_URI", "bolt://localhost:7687"),
        "user": os.environ.get("MEMLORE_NEO4J_USER", "neo4j"),
        "password": os.environ.get("MEMLORE_NEO4J_PASSWORD", "memlore-dev-password"),
    }


@pytest.fixture
def openai_available() -> bool:
    return bool(os.environ.get("OPENAI_API_KEY"))


@pytest.fixture
def require_graph_stack(neo4j_settings: dict[str, str], openai_available: bool) -> None:
    if not openai_available:
        pytest.skip("OPENAI_API_KEY not set — Graphiti ingest/search unavailable")
    try:
        from neo4j import GraphDatabase

        driver = GraphDatabase.driver(
            neo4j_settings["uri"],
            auth=(neo4j_settings["user"], neo4j_settings["password"]),
        )
        driver.verify_connectivity()
        driver.close()
    except Exception as exc:
        pytest.skip(f"Neo4j unavailable: {exc}")


@pytest.fixture
def mock_knowledge_graph() -> Any:
    from tests.support.fake_graph import FakeKnowledgeGraph

    return FakeKnowledgeGraph()

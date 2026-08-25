from __future__ import annotations

from typing import Protocol

from graph_service.domain.models import EpisodeIngest, GraphFact, SearchQuery


class KnowledgeGraphPort(Protocol):
    """Application port for knowledge graph operations."""

    async def ingest_episode(self, episode: EpisodeIngest) -> str:
        """Persist an episode and return its id."""

    async def search(self, query: SearchQuery) -> list[GraphFact]:
        """Search facts matching the query."""

    async def get_fact(self, fact_id: str) -> GraphFact | None:
        """Retrieve a fact by id, or None if not found."""

    async def ping_neo4j(self) -> bool:
        """Return True when Neo4j is reachable."""

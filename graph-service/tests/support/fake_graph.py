from __future__ import annotations

from graph_service.domain.models import EpisodeIngest, GraphFact, SearchQuery


class FakeKnowledgeGraph:
    """In-memory fake for unit and contract tests."""

    def __init__(self) -> None:
        self.episodes: list[EpisodeIngest] = []
        self.facts: dict[str, GraphFact] = {}
        self.neo4j_ok = True

    async def ingest_episode(self, episode: EpisodeIngest) -> str:
        episode_id = episode.episode_id or f"ep-{len(self.episodes) + 1}"
        stored = EpisodeIngest(
            statement=episode.statement,
            scope=episode.scope,
            metadata=episode.metadata,
            provenance_refs=episode.provenance_refs,
            reference_time=episode.reference_time,
            episode_id=episode_id,
        )
        self.episodes.append(stored)
        fact_id = f"fact-{episode_id}"
        self.facts[fact_id] = GraphFact(
            id=fact_id,
            statement=episode.statement,
            score=1.0,
            scope=episode.scope,
            provenance_refs=list(episode.provenance_refs),
        )
        return episode_id

    async def search(self, query: SearchQuery) -> list[GraphFact]:
        results = [
            fact
            for fact in self.facts.values()
            if query.query.lower() in fact.statement.lower()
        ]
        if query.scope is not None:
            results = [
                f
                for f in results
                if f.scope is not None
                and f.scope.kind == query.scope.kind
                and f.scope.key == query.scope.key
            ]
        return results[: query.limit]

    async def get_fact(self, fact_id: str) -> GraphFact | None:
        return self.facts.get(fact_id)

    async def ping_neo4j(self) -> bool:
        return self.neo4j_ok

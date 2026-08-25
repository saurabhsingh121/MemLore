from __future__ import annotations

import json
import logging
import os
import uuid
from datetime import UTC, datetime

from graphiti_core import Graphiti
from graphiti_core.nodes import EpisodeType

from graph_service.domain.models import EpisodeIngest, GraphFact, Scope, SearchQuery

logger = logging.getLogger(__name__)

FORBIDDEN_RESULT_KEYS = frozenset(
    {
        "EntityEdge",
        "group_id",
        "fact_embedding",
        "episodes",
        "uuid",
    }
)


class GraphitiUnavailableError(RuntimeError):
    """Raised when Graphiti cannot connect or OpenAI is not configured."""


class GraphitiKnowledgeGraph:
    """Thin adapter — all graphiti_core imports stay in this module."""

    def __init__(
        self,
        uri: str,
        user: str,
        password: str,
        openai_api_key: str | None = None,
    ) -> None:
        self._uri = uri
        self._user = user
        self._password = password
        self._openai_api_key = openai_api_key
        self._graphiti: Graphiti | None = None
        self._initialized = False

    async def _client(self) -> Graphiti:
        if self._graphiti is None:
            if not self._openai_api_key:
                raise GraphitiUnavailableError(
                    "OPENAI_API_KEY is required for Graphiti ingest/search"
                )
            os.environ.setdefault("OPENAI_API_KEY", self._openai_api_key)
            self._graphiti = Graphiti(
                self._uri,
                self._user,
                self._password,
            )
        if not self._initialized:
            await self._graphiti.build_indices_and_constraints()
            self._initialized = True
        return self._graphiti

    async def ping_neo4j(self) -> bool:
        try:
            client = await self._client()
            driver = client.driver
            async with driver.session() as session:
                result = await session.run("RETURN 1 AS ok")
                record = await result.single()
                return record is not None and record.get("ok") == 1
        except Exception:
            logger.exception("neo4j ping failed")
            return False

    async def ingest_episode(self, episode: EpisodeIngest) -> str:
        client = await self._client()
        episode_id = episode.episode_id or str(uuid.uuid4())
        reference_time = episode.reference_time or datetime.now(UTC)
        body = {
            "statement": episode.statement,
            "metadata": episode.metadata,
            "provenance_refs": episode.provenance_refs,
        }
        await client.add_episode(
            name=episode_id,
            episode_body=json.dumps(body),
            source=EpisodeType.json,
            source_description="memlore-episode",
            reference_time=reference_time,
            group_id=episode.scope.group_id(),
        )
        return episode_id

    async def search(self, query: SearchQuery) -> list[GraphFact]:
        client = await self._client()
        group_ids = [query.scope.group_id()] if query.scope else None
        edges = await client.search(
            query.query,
            group_ids=group_ids,
            num_results=query.limit,
        )
        facts: list[GraphFact] = []
        for edge in edges:
            scope = None
            group_id = getattr(edge, "group_id", None)
            if group_id:
                scope = Scope.from_group_id(str(group_id))
            score = float(getattr(edge, "score", 1.0) or 1.0)
            facts.append(
                GraphFact(
                    id=str(edge.uuid),
                    statement=str(edge.fact),
                    score=score,
                    scope=scope,
                    provenance_refs=[],
                )
            )
        return facts

    async def get_fact(self, fact_id: str) -> GraphFact | None:
        client = await self._client()
        driver = client.driver
        async with driver.session() as session:
            result = await session.run(
                """
                MATCH ()-[e:RELATES_TO]->()
                WHERE e.uuid = $fact_id
                RETURN e.uuid AS uuid, e.fact AS fact, e.group_id AS group_id
                LIMIT 1
                """,
                fact_id=fact_id,
            )
            record = await result.single()
            if record is None:
                return None
            scope = None
            group_id = record.get("group_id")
            if group_id:
                scope = Scope.from_group_id(str(group_id))
            return GraphFact(
                id=str(record["uuid"]),
                statement=str(record["fact"]),
                score=1.0,
                scope=scope,
                provenance_refs=[],
            )

    async def close(self) -> None:
        if self._graphiti is not None:
            await self._graphiti.close()  # type: ignore[no-untyped-call]
            self._graphiti = None
            self._initialized = False


def assert_memlore_result_shape(payload: dict[str, object]) -> None:
    """Contract helper: ensure no Graphiti field names leak to API JSON."""
    for key in payload:
        if key in FORBIDDEN_RESULT_KEYS:
            raise ValueError(f"forbidden Graphiti field in response: {key}")

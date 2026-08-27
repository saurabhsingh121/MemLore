from __future__ import annotations

from datetime import datetime
from typing import Annotated, Any, cast

from fastapi import APIRouter, Depends, FastAPI, HTTPException, Request
from pydantic import BaseModel, Field

from graph_service.adapters.graphiti import (
    GraphitiKnowledgeGraph,
    GraphitiUnavailableError,
)
from graph_service.application.ports.knowledge_graph import KnowledgeGraphPort
from graph_service.bootstrap.settings import Settings, load_settings
from graph_service.domain.models import EpisodeIngest, GraphFact, Scope, SearchQuery

router = APIRouter()


class ScopeSchema(BaseModel):
    kind: str = Field(min_length=1, max_length=64)
    key: str = Field(min_length=1, max_length=512)


class EpisodeRequest(BaseModel):
    statement: str = Field(min_length=1, max_length=8000)
    scope: ScopeSchema
    episode_id: str | None = Field(default=None, max_length=36)
    metadata: dict[str, Any] = Field(default_factory=dict)
    provenance_refs: list[str] = Field(default_factory=list)
    reference_time: datetime | None = None


class EpisodeResponse(BaseModel):
    episode_id: str


class SearchRequest(BaseModel):
    query: str = Field(min_length=1, max_length=2000)
    scope: ScopeSchema | None = None
    limit: int = Field(default=10, ge=1, le=50)


class GraphFactResponse(BaseModel):
    id: str
    statement: str
    score: float
    scope: ScopeSchema | None = None
    provenance_refs: list[str] = Field(default_factory=list)


class SearchResponse(BaseModel):
    results: list[GraphFactResponse]


class HealthResponse(BaseModel):
    status: str
    service: str = "graph-service"
    neo4j: str


class ErrorBody(BaseModel):
    code: str
    message: str


class ErrorEnvelope(BaseModel):
    error: ErrorBody


def get_knowledge_graph(request: Request) -> KnowledgeGraphPort:
    graph = cast(KnowledgeGraphPort | None, request.app.state.knowledge_graph)
    if graph is None:
        raise HTTPException(
            status_code=503,
            detail={
                "error": {
                    "code": "service_unavailable",
                    "message": "knowledge graph not configured",
                }
            },
        )
    return graph


KnowledgeGraphDep = Annotated[KnowledgeGraphPort, Depends(get_knowledge_graph)]


def _to_scope(schema: ScopeSchema) -> Scope:
    return Scope(kind=schema.kind, key=schema.key)


def _fact_response(fact: GraphFact) -> GraphFactResponse:
    scope_out = None
    if fact.scope is not None:
        scope_out = ScopeSchema(kind=fact.scope.kind, key=fact.scope.key)
    return GraphFactResponse(
        id=fact.id,
        statement=fact.statement,
        score=fact.score,
        scope=scope_out,
        provenance_refs=list(fact.provenance_refs),
    )


@router.get("/health", response_model=HealthResponse)
async def health(
    graph: KnowledgeGraphDep,
) -> HealthResponse:
    neo4j_ok = await graph.ping_neo4j()
    neo4j_status = "ok" if neo4j_ok else "unavailable"
    status = "ok" if neo4j_ok else "degraded"
    return HealthResponse(status=status, neo4j=neo4j_status)


@router.post("/episodes", response_model=EpisodeResponse, status_code=201)
async def create_episode(
    body: EpisodeRequest,
    graph: KnowledgeGraphDep,
) -> EpisodeResponse:
    episode = EpisodeIngest(
        statement=body.statement,
        scope=_to_scope(body.scope),
        metadata=body.metadata,
        provenance_refs=body.provenance_refs,
        reference_time=body.reference_time,
        episode_id=body.episode_id,
    )
    try:
        episode_id = await graph.ingest_episode(episode)
    except GraphitiUnavailableError as exc:
        raise HTTPException(
            status_code=503,
            detail={"error": {"code": "service_unavailable", "message": str(exc)}},
        ) from exc
    except Exception as exc:
        raise HTTPException(
            status_code=503,
            detail={
                "error": {
                    "code": "service_unavailable",
                    "message": "graph ingest failed",
                }
            },
        ) from exc
    return EpisodeResponse(episode_id=episode_id)


@router.post("/search", response_model=SearchResponse)
async def search_facts(
    body: SearchRequest,
    graph: KnowledgeGraphDep,
) -> SearchResponse:
    scope = _to_scope(body.scope) if body.scope else None
    query = SearchQuery(query=body.query, scope=scope, limit=body.limit)
    try:
        facts = await graph.search(query)
    except GraphitiUnavailableError as exc:
        raise HTTPException(
            status_code=503,
            detail={"error": {"code": "service_unavailable", "message": str(exc)}},
        ) from exc
    except Exception as exc:
        raise HTTPException(
            status_code=503,
            detail={
                "error": {
                    "code": "service_unavailable",
                    "message": "graph search failed",
                }
            },
        ) from exc
    return SearchResponse(results=[_fact_response(f) for f in facts])


@router.get("/facts/{fact_id}", response_model=GraphFactResponse)
async def get_fact(
    fact_id: str,
    graph: KnowledgeGraphDep,
) -> GraphFactResponse:
    try:
        fact = await graph.get_fact(fact_id)
    except GraphitiUnavailableError as exc:
        raise HTTPException(
            status_code=503,
            detail={"error": {"code": "service_unavailable", "message": str(exc)}},
        ) from exc
    if fact is None:
        raise HTTPException(
            status_code=404,
            detail={"error": {"code": "not_found", "message": "fact not found"}},
        )
    return _fact_response(fact)


def create_app(settings: Settings | None = None) -> FastAPI:
    cfg = settings or load_settings()
    app = FastAPI(title="MemLore Graph Service", version="0.1.0")
    app.state.settings = cfg
    app.state.knowledge_graph = GraphitiKnowledgeGraph(
        uri=cfg.neo4j_uri,
        user=cfg.neo4j_user,
        password=cfg.neo4j_password,
        openai_api_key=cfg.openai_api_key,
    )

    @app.on_event("shutdown")
    async def shutdown() -> None:
        graph = app.state.knowledge_graph
        if isinstance(graph, GraphitiKnowledgeGraph):
            await graph.close()

    app.include_router(router)
    return app

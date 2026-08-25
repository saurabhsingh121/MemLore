from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime
from typing import Any


@dataclass(frozen=True)
class Scope:
    kind: str
    key: str

    def group_id(self) -> str:
        return f"{self.kind}:{self.key}"

    @staticmethod
    def from_group_id(group_id: str) -> Scope | None:
        if ":" not in group_id:
            return None
        kind, key = group_id.split(":", 1)
        if not kind or not key:
            return None
        return Scope(kind=kind, key=key)


@dataclass(frozen=True)
class EpisodeIngest:
    statement: str
    scope: Scope
    metadata: dict[str, Any] = field(default_factory=dict)
    provenance_refs: list[str] = field(default_factory=list)
    reference_time: datetime | None = None
    episode_id: str | None = None


@dataclass(frozen=True)
class SearchQuery:
    query: str
    scope: Scope | None = None
    limit: int = 10


@dataclass(frozen=True)
class GraphFact:
    id: str
    statement: str
    score: float
    scope: Scope | None = None
    provenance_refs: list[str] = field(default_factory=list)

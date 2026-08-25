from __future__ import annotations

from typing import Protocol

from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope


class LoreRepository(Protocol):
    def add(self, entry: LoreEntry) -> None: ...

    def get(self, entry_id: str) -> LoreEntry | None: ...

    def save(self, entry: LoreEntry) -> None: ...

    def list_by_scope(self, scope: Scope) -> list[LoreEntry]: ...

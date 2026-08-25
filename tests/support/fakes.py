from __future__ import annotations

from dataclasses import dataclass, field

from memlore.domain.models.audit_record import AuditRecord
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope


@dataclass
class InMemoryLoreRepository:
    _items: dict[str, LoreEntry] = field(default_factory=dict)

    def add(self, entry: LoreEntry) -> None:
        self._items[entry.id] = entry

    def get(self, entry_id: str) -> LoreEntry | None:
        return self._items.get(entry_id)

    def save(self, entry: LoreEntry) -> None:
        self._items[entry.id] = entry

    def list_by_scope(self, scope: Scope) -> list[LoreEntry]:
        return [
            e
            for e in self._items.values()
            if e.scope.kind == scope.kind and e.scope.key == scope.key
        ]


@dataclass
class InMemoryAuditRepository:
    _items: list[AuditRecord] = field(default_factory=list)

    def add(self, record: AuditRecord) -> None:
        self._items.append(record)

    def list_by_target(self, target_id: str) -> list[AuditRecord]:
        return sorted(
            [r for r in self._items if r.target_id == target_id],
            key=lambda r: (r.created_at, r.id),
        )


@dataclass
class InMemoryUnitOfWork:
    lore_entries: InMemoryLoreRepository = field(default_factory=InMemoryLoreRepository)
    audits: InMemoryAuditRepository = field(default_factory=InMemoryAuditRepository)
    committed: bool = False

    def commit(self) -> None:
        self.committed = True

    def rollback(self) -> None:
        self.committed = False

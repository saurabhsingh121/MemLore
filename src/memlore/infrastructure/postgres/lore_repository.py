from __future__ import annotations

from sqlalchemy import select
from sqlalchemy.orm import Session

from memlore.domain.models.enums import (
    EvidenceType,
    KnowledgeOrigin,
    ScopeKind,
    VerificationStatus,
)
from memlore.domain.models.evidence import EvidenceReference
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope
from memlore.infrastructure.postgres.models import LoreEntryRow


class SqlAlchemyLoreRepository:
    def __init__(self, session: Session) -> None:
        self._session = session

    def add(self, entry: LoreEntry) -> None:
        self._session.add(_to_row(entry))

    def get(self, entry_id: str) -> LoreEntry | None:
        row = self._session.get(LoreEntryRow, entry_id)
        return _from_row(row) if row else None

    def save(self, entry: LoreEntry) -> None:
        row = self._session.get(LoreEntryRow, entry.id)
        if row is None:
            self.add(entry)
            return
        row.statement = entry.statement
        row.scope_kind = entry.scope.kind.value
        row.scope_key = entry.scope.key
        row.origin = entry.origin.value
        row.verification_status = entry.verification_status.value
        row.evidence = [
            {"type": e.type.value, "value": e.value} for e in entry.evidence
        ]
        row.created_by = entry.created_by
        row.created_at = entry.created_at
        row.verified_by = entry.verified_by
        row.verified_at = entry.verified_at
        row.updated_at = entry.updated_at

    def list_by_scope(self, scope: Scope) -> list[LoreEntry]:
        stmt = (
            select(LoreEntryRow)
            .where(
                LoreEntryRow.scope_kind == scope.kind.value,
                LoreEntryRow.scope_key == scope.key,
            )
            .order_by(LoreEntryRow.created_at.desc())
        )
        return [_from_row(row) for row in self._session.scalars(stmt)]


def _to_row(entry: LoreEntry) -> LoreEntryRow:
    return LoreEntryRow(
        id=entry.id,
        statement=entry.statement,
        scope_kind=entry.scope.kind.value,
        scope_key=entry.scope.key,
        origin=entry.origin.value,
        verification_status=entry.verification_status.value,
        evidence=[{"type": e.type.value, "value": e.value} for e in entry.evidence],
        created_by=entry.created_by,
        created_at=entry.created_at,
        verified_by=entry.verified_by,
        verified_at=entry.verified_at,
        updated_at=entry.updated_at,
    )


def _from_row(row: LoreEntryRow) -> LoreEntry:
    return LoreEntry(
        id=row.id,
        statement=row.statement,
        scope=Scope(kind=ScopeKind(row.scope_kind), key=row.scope_key),
        origin=KnowledgeOrigin(row.origin),
        verification_status=VerificationStatus(row.verification_status),
        evidence=[
            EvidenceReference(type=EvidenceType(item["type"]), value=item["value"])
            for item in row.evidence
        ],
        created_by=row.created_by,
        created_at=row.created_at,
        verified_by=row.verified_by,
        verified_at=row.verified_at,
        updated_at=row.updated_at,
    )

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import UTC, datetime
from uuid import uuid4

from memlore.domain.exceptions import ValidationError
from memlore.domain.models.enums import KnowledgeOrigin, VerificationStatus
from memlore.domain.models.evidence import EvidenceReference
from memlore.domain.models.scope import Scope

MAX_STATEMENT_LENGTH = 8000


@dataclass(slots=True)
class LoreEntry:
    statement: str
    scope: Scope
    created_by: str
    id: str = field(default_factory=lambda: str(uuid4()))
    origin: KnowledgeOrigin = KnowledgeOrigin.HUMAN_AUTHORED
    verification_status: VerificationStatus = VerificationStatus.UNVERIFIED
    evidence: list[EvidenceReference] = field(default_factory=list)
    created_at: datetime = field(default_factory=lambda: datetime.now(UTC))
    verified_by: str | None = None
    verified_at: datetime | None = None
    updated_at: datetime = field(default_factory=lambda: datetime.now(UTC))

    def __post_init__(self) -> None:
        statement = self.statement.strip()
        created_by = self.created_by.strip()
        if not statement:
            raise ValidationError("statement must be non-empty")
        if len(statement) > MAX_STATEMENT_LENGTH:
            raise ValidationError(
                f"statement must be at most {MAX_STATEMENT_LENGTH} characters"
            )
        if not created_by:
            raise ValidationError("created_by must be non-empty")
        if self.origin != KnowledgeOrigin.HUMAN_AUTHORED:
            raise ValidationError("create origin must be human_authored")
        self.statement = statement
        self.created_by = created_by

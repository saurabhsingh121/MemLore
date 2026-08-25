from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, Field

from memlore.domain.models.enums import (
    AuditAction,
    EvidenceType,
    KnowledgeOrigin,
    ScopeKind,
    VerificationStatus,
)
from memlore.domain.models.evidence import MAX_EVIDENCE_VALUE_LENGTH
from memlore.domain.models.lore_entry import MAX_STATEMENT_LENGTH
from memlore.domain.models.scope import MAX_SCOPE_KEY_LENGTH


class ScopeIn(BaseModel):
    kind: ScopeKind
    key: str = Field(min_length=1, max_length=MAX_SCOPE_KEY_LENGTH)


class EvidenceIn(BaseModel):
    type: EvidenceType
    value: str = Field(min_length=1, max_length=MAX_EVIDENCE_VALUE_LENGTH)


class CreateLoreRequest(BaseModel):
    statement: str = Field(min_length=1, max_length=MAX_STATEMENT_LENGTH)
    scope: ScopeIn
    evidence: list[EvidenceIn] = Field(default_factory=list)


class ScopeOut(BaseModel):
    kind: ScopeKind
    key: str


class EvidenceOut(BaseModel):
    type: EvidenceType
    value: str


class LoreEntryResponse(BaseModel):
    id: str
    statement: str
    scope: ScopeOut
    origin: KnowledgeOrigin
    verification_status: VerificationStatus
    evidence: list[EvidenceOut]
    created_by: str
    created_at: datetime
    verified_by: str | None
    verified_at: datetime | None
    updated_at: datetime


class LoreEntryListResponse(BaseModel):
    items: list[LoreEntryResponse]


class AuditRecordResponse(BaseModel):
    id: str
    target_id: str
    action: AuditAction
    actor_id: str
    created_at: datetime


class AuditListResponse(BaseModel):
    items: list[AuditRecordResponse]

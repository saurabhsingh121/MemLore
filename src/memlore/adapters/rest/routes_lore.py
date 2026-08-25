from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Depends, Query

from memlore.adapters.rest.deps import require_actor
from memlore.adapters.rest.schemas import (
    AuditListResponse,
    AuditRecordResponse,
    CreateLoreRequest,
    EvidenceOut,
    LoreEntryListResponse,
    LoreEntryResponse,
    ScopeOut,
)
from memlore.application.commands.create_lore import CreateLoreCommand
from memlore.application.commands.verify_lore import VerifyLoreCommand
from memlore.bootstrap.container import AppContainer
from memlore.domain.models.audit_record import AuditRecord
from memlore.domain.models.enums import ScopeKind
from memlore.domain.models.evidence import EvidenceReference
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope
from memlore.infrastructure.telemetry.logging import get_logger, log_operation

logger = get_logger()
router = APIRouter(prefix="/v1")


def get_container() -> AppContainer:
    raise RuntimeError("AppContainer dependency is not configured")


ContainerDep = Annotated[AppContainer, Depends(get_container)]
ActorDep = Annotated[str, Depends(require_actor)]


def _to_response(entry: LoreEntry) -> LoreEntryResponse:
    return LoreEntryResponse(
        id=entry.id,
        statement=entry.statement,
        scope=ScopeOut(kind=entry.scope.kind, key=entry.scope.key),
        origin=entry.origin,
        verification_status=entry.verification_status,
        evidence=[EvidenceOut(type=e.type, value=e.value) for e in entry.evidence],
        created_by=entry.created_by,
        created_at=entry.created_at,
        verified_by=entry.verified_by,
        verified_at=entry.verified_at,
        updated_at=entry.updated_at,
    )


def _audit_to_response(record: AuditRecord) -> AuditRecordResponse:
    return AuditRecordResponse(
        id=record.id,
        target_id=record.target_id,
        action=record.action,
        actor_id=record.actor_id,
        created_at=record.created_at,
    )


@router.post("/lore-entries", response_model=LoreEntryResponse, status_code=201)
def create_lore_entry(
    body: CreateLoreRequest,
    actor: ActorDep,
    container: ContainerDep,
) -> LoreEntryResponse:
    with container.unit_of_work() as uow:
        entry = container.create_lore_handler(uow).handle(
            CreateLoreCommand(
                statement=body.statement,
                scope=Scope(kind=body.scope.kind, key=body.scope.key),
                actor_id=actor,
                evidence=tuple(
                    EvidenceReference(type=e.type, value=e.value) for e in body.evidence
                ),
            )
        )
    log_operation(
        logger, operation="lore.create", actor_id=actor, lore_entry_id=entry.id
    )
    return _to_response(entry)


@router.get("/lore-entries/{entry_id}", response_model=LoreEntryResponse)
def get_lore_entry(entry_id: str, container: ContainerDep) -> LoreEntryResponse:
    with container.unit_of_work() as uow:
        entry = container.get_lore_handler(uow).handle(entry_id)
    log_operation(logger, operation="lore.get", lore_entry_id=entry.id)
    return _to_response(entry)


@router.post("/lore-entries/{entry_id}/verify", response_model=LoreEntryResponse)
def verify_lore_entry(
    entry_id: str,
    actor: ActorDep,
    container: ContainerDep,
) -> LoreEntryResponse:
    with container.unit_of_work() as uow:
        entry = container.verify_lore_handler(uow).handle(
            VerifyLoreCommand(entry_id=entry_id, actor_id=actor)
        )
    log_operation(
        logger, operation="lore.verify", actor_id=actor, lore_entry_id=entry.id
    )
    return _to_response(entry)


@router.get("/lore-entries", response_model=LoreEntryListResponse)
def list_lore_entries(
    container: ContainerDep,
    scope_kind: Annotated[ScopeKind, Query()],
    scope_key: Annotated[str, Query(min_length=1)],
) -> LoreEntryListResponse:
    scope = Scope(kind=scope_kind, key=scope_key)
    with container.unit_of_work() as uow:
        items = container.list_lore_handler(uow).handle(scope)
    log_operation(
        logger,
        operation="lore.list",
        scope_kind=scope_kind.value,
        scope_key=scope.key,
        count=len(items),
    )
    return LoreEntryListResponse(items=[_to_response(e) for e in items])


@router.get(
    "/lore-entries/{entry_id}/audits",
    response_model=AuditListResponse,
)
def list_lore_audits(entry_id: str, container: ContainerDep) -> AuditListResponse:
    with container.unit_of_work() as uow:
        items = container.list_audits_handler(uow).handle(entry_id)
    log_operation(logger, operation="lore.audits", lore_entry_id=entry_id)
    return AuditListResponse(items=[_audit_to_response(a) for a in items])

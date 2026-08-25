from __future__ import annotations

from typing import Any

from memlore.adapters.mcp.errors import raise_as_tool_error
from memlore.adapters.rest.schemas import (
    AuditRecordResponse,
    EvidenceIn,
    EvidenceOut,
    LoreEntryListResponse,
    LoreEntryResponse,
    ScopeIn,
    ScopeOut,
)
from memlore.application.commands.create_lore import CreateLoreCommand
from memlore.application.commands.verify_lore import VerifyLoreCommand
from memlore.bootstrap.container import AppContainer
from memlore.domain.exceptions import NotFoundError, ValidationError
from memlore.domain.models.audit_record import AuditRecord
from memlore.domain.models.evidence import EvidenceReference
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope
from memlore.infrastructure.telemetry.logging import get_logger, log_operation

logger = get_logger()


class ExplainLoreResult(LoreEntryResponse):
    audits: list[AuditRecordResponse]


def lore_entry_response(entry: LoreEntry) -> LoreEntryResponse:
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


def lore_entry_payload(entry: LoreEntry) -> dict[str, Any]:
    return lore_entry_response(entry).model_dump(mode="json")


def audit_record_response(record: AuditRecord) -> AuditRecordResponse:
    return AuditRecordResponse(
        id=record.id,
        target_id=record.target_id,
        action=record.action,
        actor_id=record.actor_id,
        created_at=record.created_at,
    )


def audit_record_payload(record: AuditRecord) -> dict[str, Any]:
    return audit_record_response(record).model_dump(mode="json")


def _require_actor_id(actor_id: str) -> str:
    actor = actor_id.strip()
    if not actor:
        raise ValidationError("actor must be non-empty")
    return actor


def _to_scope(scope: ScopeIn) -> Scope:
    return Scope(kind=scope.kind, key=scope.key)


def _to_evidence(evidence: list[EvidenceIn] | None) -> tuple[EvidenceReference, ...]:
    if not evidence:
        return ()
    return tuple(
        EvidenceReference(type=item.type, value=item.value) for item in evidence
    )


def remember_lore(
    container: AppContainer,
    *,
    statement: str,
    scope: ScopeIn,
    actor_id: str,
    evidence: list[EvidenceIn] | None = None,
) -> LoreEntryResponse:
    try:
        actor = _require_actor_id(actor_id)
        with container.unit_of_work() as uow:
            entry = container.create_lore_handler(uow).handle(
                CreateLoreCommand(
                    statement=statement,
                    scope=_to_scope(scope),
                    actor_id=actor,
                    evidence=_to_evidence(evidence),
                )
            )
    except (ValidationError, NotFoundError) as exc:
        raise_as_tool_error(exc)
    log_operation(
        logger, operation="mcp.remember", actor_id=actor, lore_entry_id=entry.id
    )
    return lore_entry_response(entry)


def get_lore(container: AppContainer, *, id: str) -> LoreEntryResponse:
    try:
        with container.unit_of_work() as uow:
            entry = container.get_lore_handler(uow).handle(id)
    except (ValidationError, NotFoundError) as exc:
        raise_as_tool_error(exc)
    log_operation(logger, operation="mcp.get", lore_entry_id=entry.id)
    return lore_entry_response(entry)


def explain_lore(container: AppContainer, *, id: str) -> ExplainLoreResult:
    try:
        with container.unit_of_work() as uow:
            entry = container.get_lore_handler(uow).handle(id)
            audits = container.list_audits_handler(uow).handle(id)
    except (ValidationError, NotFoundError) as expl_exc:
        raise_as_tool_error(expl_exc)
    log_operation(logger, operation="mcp.explain", lore_entry_id=entry.id)
    base = lore_entry_response(entry)
    return ExplainLoreResult(
        **base.model_dump(),
        audits=[audit_record_response(item) for item in audits],
    )


def verify_lore(
    container: AppContainer, *, id: str, actor_id: str
) -> LoreEntryResponse:
    try:
        actor = _require_actor_id(actor_id)
        with container.unit_of_work() as uow:
            entry = container.verify_lore_handler(uow).handle(
                VerifyLoreCommand(entry_id=id, actor_id=actor)
            )
    except (ValidationError, NotFoundError) as exc:
        raise_as_tool_error(exc)
    log_operation(
        logger, operation="mcp.verify", actor_id=actor, lore_entry_id=entry.id
    )
    return lore_entry_response(entry)


def search_lore(
    container: AppContainer, *, scope: ScopeIn | None
) -> LoreEntryListResponse:
    if scope is None:
        raise_as_tool_error(ValidationError("scope is required"))
    try:
        domain_scope = _to_scope(scope)
        with container.unit_of_work() as uow:
            items = container.list_lore_handler(uow).handle(domain_scope)
    except (ValidationError, NotFoundError) as exc:
        raise_as_tool_error(exc)
    log_operation(
        logger,
        operation="mcp.search",
        scope_kind=domain_scope.kind.value,
        scope_key=domain_scope.key,
        count=len(items),
    )
    return LoreEntryListResponse(items=[lore_entry_response(item) for item in items])

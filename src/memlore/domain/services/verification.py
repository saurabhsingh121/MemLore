from __future__ import annotations

from datetime import UTC, datetime

from memlore.domain.exceptions import ValidationError
from memlore.domain.models.audit_record import AuditRecord
from memlore.domain.models.enums import AuditAction, VerificationStatus
from memlore.domain.models.lore_entry import LoreEntry


def apply_verification(
    entry: LoreEntry,
    *,
    actor_id: str,
    now: datetime,
) -> tuple[LoreEntry, AuditRecord | None]:
    actor = actor_id.strip()
    if not actor:
        raise ValidationError("actor must be non-empty")

    if entry.verification_status == VerificationStatus.VERIFIED:
        return entry, None

    entry.verification_status = VerificationStatus.VERIFIED
    entry.verified_by = actor
    entry.verified_at = now
    entry.updated_at = now
    audit = AuditRecord(
        target_id=entry.id,
        action=AuditAction.VERIFY,
        actor_id=actor,
        created_at=now,
    )
    return entry, audit


def ensure_utc(value: datetime) -> datetime:
    if value.tzinfo is None:
        return value.replace(tzinfo=UTC)
    return value.astimezone(UTC)

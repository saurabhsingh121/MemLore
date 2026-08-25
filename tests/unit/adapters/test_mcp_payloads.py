from __future__ import annotations

from datetime import UTC, datetime

from memlore.adapters.mcp.tools import audit_record_payload, lore_entry_payload
from memlore.adapters.rest.schemas import AuditRecordResponse, LoreEntryResponse
from memlore.domain.models.audit_record import AuditRecord
from memlore.domain.models.enums import AuditAction, KnowledgeOrigin, ScopeKind
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope


def test_lore_entry_payload_matches_rest_response_fields() -> None:
    now = datetime(2026, 8, 25, 12, 0, tzinfo=UTC)
    entry = LoreEntry(
        statement="Use the outbox",
        scope=Scope(kind=ScopeKind.REPOSITORY, key="github.com/acme/app"),
        created_by="alice",
        created_at=now,
        updated_at=now,
    )
    payload = lore_entry_payload(entry)
    expected = LoreEntryResponse.model_validate(payload).model_dump(mode="json")
    assert set(payload) == set(expected)
    assert payload["origin"] == KnowledgeOrigin.HUMAN_AUTHORED.value
    assert payload["statement"] == "Use the outbox"
    assert payload["scope"] == {"kind": "repository", "key": "github.com/acme/app"}
    assert payload["created_by"] == "alice"
    assert "audits" not in payload


def test_audit_record_payload_matches_rest_response_fields() -> None:
    now = datetime(2026, 8, 25, 12, 0, tzinfo=UTC)
    record = AuditRecord(
        target_id="00000000-0000-0000-0000-000000000001",
        action=AuditAction.CREATE,
        actor_id="alice",
        created_at=now,
    )
    payload = audit_record_payload(record)
    expected = AuditRecordResponse.model_validate(payload).model_dump(mode="json")
    assert set(payload) == set(expected)
    assert payload["action"] == "create"
    assert payload["actor_id"] == "alice"

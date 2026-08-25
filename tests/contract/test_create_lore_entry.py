from __future__ import annotations

from datetime import UTC, datetime

from fastapi.testclient import TestClient

from memlore.adapters.rest.app import create_app
from memlore.bootstrap.container import build_memory_container
from memlore.domain.models.lore_entry import MAX_STATEMENT_LENGTH
from memlore.infrastructure.clock import FixedClock
from tests.support.fakes import InMemoryUnitOfWork


def _client() -> TestClient:
    uow = InMemoryUnitOfWork()
    app = create_app(
        build_memory_container(uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC)))
    )
    return TestClient(app)


def test_create_lore_entry_contract() -> None:
    client = _client()
    response = client.post(
        "/v1/lore-entries",
        headers={"X-Memlore-Actor": "alice"},
        json={
            "statement": "Payment events must use the transactional outbox.",
            "scope": {"kind": "repository", "key": "github.com/acme/payments"},
            "evidence": [{"type": "adr", "value": "0001-dual-plane"}],
        },
    )
    assert response.status_code == 201
    body = response.json()
    assert body["origin"] == "human_authored"
    assert body["verification_status"] == "unverified"
    assert body["created_by"] == "alice"
    assert body["scope"]["kind"] == "repository"


def test_create_rejects_missing_actor_and_oversized_statement() -> None:
    client = _client()
    missing = client.post(
        "/v1/lore-entries",
        json={
            "statement": "x",
            "scope": {"kind": "team", "key": "t1"},
        },
    )
    assert missing.status_code == 400

    oversized = client.post(
        "/v1/lore-entries",
        headers={"X-Memlore-Actor": "alice"},
        json={
            "statement": "x" * (MAX_STATEMENT_LENGTH + 1),
            "scope": {"kind": "team", "key": "t1"},
        },
    )
    assert oversized.status_code in {400, 422}


def test_create_duplicate_statement_allowed() -> None:
    client = _client()
    payload = {
        "statement": "Same",
        "scope": {"kind": "team", "key": "t1"},
    }
    a = client.post(
        "/v1/lore-entries", headers={"X-Memlore-Actor": "alice"}, json=payload
    )
    b = client.post(
        "/v1/lore-entries", headers={"X-Memlore-Actor": "alice"}, json=payload
    )
    assert a.status_code == 201
    assert b.status_code == 201
    assert a.json()["id"] != b.json()["id"]

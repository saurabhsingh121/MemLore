from __future__ import annotations

from datetime import UTC, datetime

from fastapi.testclient import TestClient

from memlore.adapters.rest.app import create_app
from memlore.bootstrap.container import build_memory_container
from memlore.infrastructure.clock import FixedClock
from tests.support.fakes import InMemoryUnitOfWork


def _client_with_entry() -> tuple[TestClient, str]:
    uow = InMemoryUnitOfWork()
    app = create_app(
        build_memory_container(uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC)))
    )
    client = TestClient(app)
    created = client.post(
        "/v1/lore-entries",
        headers={"X-Memlore-Actor": "alice"},
        json={
            "statement": "Rule",
            "scope": {"kind": "repository", "key": "r1"},
        },
    )
    return client, created.json()["id"]


def test_get_and_verify_and_audits_flow() -> None:
    client, entry_id = _client_with_entry()

    got = client.get(f"/v1/lore-entries/{entry_id}")
    assert got.status_code == 200
    assert got.json()["verification_status"] == "unverified"

    missing = client.get("/v1/lore-entries/00000000-0000-0000-0000-000000000000")
    assert missing.status_code == 404

    verified = client.post(
        f"/v1/lore-entries/{entry_id}/verify",
        headers={"X-Memlore-Actor": "alice"},
    )
    assert verified.status_code == 200
    assert verified.json()["verification_status"] == "verified"
    assert verified.json()["origin"] == "human_authored"

    again = client.post(
        f"/v1/lore-entries/{entry_id}/verify",
        headers={"X-Memlore-Actor": "bob"},
    )
    assert again.status_code == 200
    assert again.json()["verified_by"] == "alice"

    audits = client.get(f"/v1/lore-entries/{entry_id}/audits")
    assert audits.status_code == 200
    actions = [a["action"] for a in audits.json()["items"]]
    assert actions == ["create", "verify"]

    audits_missing = client.get(
        "/v1/lore-entries/00000000-0000-0000-0000-000000000000/audits"
    )
    assert audits_missing.status_code == 404


def test_list_by_scope() -> None:
    client, _ = _client_with_entry()
    client.post(
        "/v1/lore-entries",
        headers={"X-Memlore-Actor": "alice"},
        json={
            "statement": "Other",
            "scope": {"kind": "team", "key": "r1"},
        },
    )
    listed = client.get(
        "/v1/lore-entries",
        params={"scope_kind": "repository", "scope_key": "r1"},
    )
    assert listed.status_code == 200
    assert len(listed.json()["items"]) == 1

    empty = client.get(
        "/v1/lore-entries",
        params={"scope_kind": "repository", "scope_key": "none"},
    )
    assert empty.status_code == 200
    assert empty.json()["items"] == []

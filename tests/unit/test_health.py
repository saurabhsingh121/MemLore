from __future__ import annotations

from fastapi.testclient import TestClient

from memlore import __version__
from memlore.adapters.rest.app import create_app


def test_health_endpoint_returns_ok_status_and_version() -> None:
    client = TestClient(create_app())

    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {
        "status": "ok",
        "service": "memlore",
        "version": __version__,
    }

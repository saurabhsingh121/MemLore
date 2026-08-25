from __future__ import annotations

from memlore.domain.models.health import HealthStatus


def test_health_status_defaults_to_ok_for_memlore_service() -> None:
    status = HealthStatus(version="0.1.0")

    assert status.status == "ok"
    assert status.service == "memlore"
    assert status.version == "0.1.0"

from __future__ import annotations

from graph_service.adapters.graphiti.adapter import assert_memlore_result_shape
from graph_service.domain.models import Scope


def test_scope_group_id_round_trip() -> None:
    scope = Scope(kind="repository", key="github.com/acme/payments")
    assert scope.group_id() == "repository:github.com/acme/payments"
    parsed = Scope.from_group_id(scope.group_id())
    assert parsed == scope


def test_assert_memlore_result_shape_rejects_graphiti_keys() -> None:
    try:
        assert_memlore_result_shape({"group_id": "x"})
        raise AssertionError("expected ValueError")
    except ValueError as exc:
        assert "group_id" in str(exc)


def test_assert_memlore_result_shape_allows_memlore_keys() -> None:
    assert_memlore_result_shape(
        {
            "id": "f1",
            "statement": "test",
            "score": 1.0,
            "scope": {"kind": "team", "key": "t1"},
            "provenance_refs": [],
        }
    )

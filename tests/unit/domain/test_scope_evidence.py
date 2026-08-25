from __future__ import annotations

import pytest

from memlore.domain.exceptions import ValidationError
from memlore.domain.models.enums import EvidenceType, ScopeKind
from memlore.domain.models.evidence import EvidenceReference
from memlore.domain.models.scope import Scope


def test_scope_trims_key_and_accepts_repository() -> None:
    scope = Scope(kind=ScopeKind.REPOSITORY, key="  github.com/acme/app  ")
    assert scope.key == "github.com/acme/app"


def test_scope_rejects_blank_key() -> None:
    with pytest.raises(ValidationError):
        Scope(kind=ScopeKind.TEAM, key="   ")


def test_evidence_reference_valid() -> None:
    ref = EvidenceReference(type=EvidenceType.ADR, value=" 0001-dual-plane ")
    assert ref.value == "0001-dual-plane"


def test_evidence_rejects_blank_value() -> None:
    with pytest.raises(ValidationError):
        EvidenceReference(type=EvidenceType.URL, value="")

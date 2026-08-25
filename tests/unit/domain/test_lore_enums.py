from __future__ import annotations

from memlore.domain.models.enums import (
    AuditAction,
    EvidenceType,
    KnowledgeOrigin,
    ScopeKind,
    VerificationStatus,
)


def test_scope_kind_includes_required_mvp_values() -> None:
    assert ScopeKind.TEAM == "team"
    assert ScopeKind.REPOSITORY == "repository"


def test_evidence_type_includes_url_path_adr() -> None:
    assert {t.value for t in EvidenceType} == {"url", "path", "adr"}


def test_knowledge_origin_reserves_agent_values() -> None:
    assert KnowledgeOrigin.HUMAN_AUTHORED == "human_authored"
    assert KnowledgeOrigin.AGENT_INFERENCE == "agent_inference"


def test_verification_and_audit_enums() -> None:
    assert VerificationStatus.UNVERIFIED == "unverified"
    assert VerificationStatus.VERIFIED == "verified"
    assert AuditAction.CREATE == "create"
    assert AuditAction.VERIFY == "verify"

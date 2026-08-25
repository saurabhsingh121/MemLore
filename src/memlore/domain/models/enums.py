from __future__ import annotations

from enum import StrEnum


class ScopeKind(StrEnum):
    TEAM = "team"
    REPOSITORY = "repository"
    ORGANIZATION = "organization"
    PROJECT = "project"
    FEATURE = "feature"
    TASK = "task"


class EvidenceType(StrEnum):
    URL = "url"
    PATH = "path"
    ADR = "adr"


class KnowledgeOrigin(StrEnum):
    HUMAN_AUTHORED = "human_authored"
    HUMAN_VERIFIED = "human_verified"
    AGENT_OBSERVATION = "agent_observation"
    AGENT_INFERENCE = "agent_inference"
    REPOSITORY_OBSERVATION = "repository_observation"
    IMPORTED_SOURCE = "imported_source"
    ARCHITECTURE_DECISION = "architecture_decision"


class VerificationStatus(StrEnum):
    UNVERIFIED = "unverified"
    VERIFIED = "verified"


class AuditAction(StrEnum):
    CREATE = "create"
    VERIFY = "verify"

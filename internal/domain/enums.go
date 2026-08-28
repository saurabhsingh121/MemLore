package domain

const (
	MaxStatementLength     = 8000
	MaxScopeKeyLength      = 512
	MaxEvidenceValueLength = 2048
)

// ScopeKind identifies the type of scope key.
type ScopeKind string

const (
	ScopeKindOrganization ScopeKind = "organization"
	ScopeKindTeam         ScopeKind = "team"
	ScopeKindProject      ScopeKind = "project"
	ScopeKindRepository   ScopeKind = "repository"
	ScopeKindFeature      ScopeKind = "feature"
	ScopeKindTask         ScopeKind = "task"
)

func ParseScopeKind(value string) (ScopeKind, error) {
	switch ScopeKind(value) {
	case ScopeKindOrganization, ScopeKindTeam, ScopeKindProject,
		ScopeKindRepository, ScopeKindFeature, ScopeKindTask:
		return ScopeKind(value), nil
	default:
		return "", validationError("invalid scope kind")
	}
}

// EvidenceType classifies an evidence reference.
type EvidenceType string

const (
	EvidenceTypeURL  EvidenceType = "url"
	EvidenceTypePath EvidenceType = "path"
	EvidenceTypeADR  EvidenceType = "adr"
)

func ParseEvidenceType(value string) (EvidenceType, error) {
	switch EvidenceType(value) {
	case EvidenceTypeURL, EvidenceTypePath, EvidenceTypeADR:
		return EvidenceType(value), nil
	default:
		return "", validationError("invalid evidence type")
	}
}

// KnowledgeOrigin records who or what produced knowledge.
type KnowledgeOrigin string

const (
	KnowledgeOriginHumanAuthored         KnowledgeOrigin = "human_authored"
	KnowledgeOriginHumanVerified         KnowledgeOrigin = "human_verified"
	KnowledgeOriginAgentObservation      KnowledgeOrigin = "agent_observation"
	KnowledgeOriginAgentInference        KnowledgeOrigin = "agent_inference"
	KnowledgeOriginRepositoryObservation KnowledgeOrigin = "repository_observation"
	KnowledgeOriginImportedSource        KnowledgeOrigin = "imported_source"
	KnowledgeOriginArchitectureDecision  KnowledgeOrigin = "architecture_decision"
)

// VerificationStatus is the trust posture of a lore entry.
type VerificationStatus string

const (
	VerificationUnverified  VerificationStatus = "unverified"
	VerificationVerified    VerificationStatus = "verified"
	VerificationInvalidated VerificationStatus = "invalidated"
)

// AuditAction records what happened to a lore entry.
type AuditAction string

const (
	AuditActionCreate     AuditAction = "create"
	AuditActionVerify     AuditAction = "verify"
	AuditActionInvalidate AuditAction = "invalidate"
	AuditActionSupersede  AuditAction = "supersede"
)

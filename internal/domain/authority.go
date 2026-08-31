package domain

import (
	"fmt"
	"time"
)

const (
	recencyHorizonDays = 365.0
	maxRecencyBoost    = 0.10

	baseVerified    = 0.72
	baseUnverified  = 0.48
	baseInvalidated = 0.12
	invalidatedCap  = 0.20

	evidenceWeight    = 0.10
	scopeWeight       = 0.05
	supersededPenalty = -0.25
	graphScoreCap     = 0.45

	evidenceStrengthADR   = 1.0
	evidenceStrengthOther = 0.6
	evidenceStrengthNone  = 0.0
	scopeMatchExact       = 1.0
	scopeMatchKindOnly    = 0.5
	scopeMatchNone        = 0.0
)

// TrustBand is a discrete, explainable authority class.
type TrustBand string

const (
	TrustBandCanonical TrustBand = "canonical"
	TrustBandHigh      TrustBand = "high"
	TrustBandMedium    TrustBand = "medium"
	TrustBandLow       TrustBand = "low"
	TrustBandUntrusted TrustBand = "untrusted"
)

// SourceType is the derived evidence class of a candidate.
type SourceType string

const (
	SourceTypeADR              SourceType = "adr"
	SourceTypeHumanStatement   SourceType = "human_statement"
	SourceTypeAgentObservation SourceType = "agent_observation"
	SourceTypeAgentInference   SourceType = "agent_inference"
	SourceTypeRepoObservation  SourceType = "repo_observation"
	SourceTypeImport           SourceType = "import"
	SourceTypeGraph            SourceType = "graph"
)

// FactorInputs are the explicit inputs to authority evaluation.
type FactorInputs struct {
	Origin             KnowledgeOrigin
	VerificationStatus VerificationStatus
	Superseded         bool
	CreatedAt          time.Time
	Now                time.Time
	Evidence           []EvidenceReference
	EntryScope         Scope
	RequestedScope     *Scope
	FromGraph          bool
	GraphScore         float64
}

// FactorSet is the explainable factor breakdown (never score-only).
type FactorSet struct {
	Origin             string
	VerificationStatus string
	SupersessionStatus string
	RecencyBoost       *float64
	EvidenceStrength   *float64
	SourceType         string
	ScopeMatch         *float64
	GraphScore         *float64
}

// Evaluation is an ephemeral authority result.
type Evaluation struct {
	Score     float64
	Band      TrustBand
	Factors   FactorSet
	Breakdown []string
}

// EvaluateAuthority computes score, band, and factors from explicit inputs.
func EvaluateAuthority(in FactorInputs) Evaluation {
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	if in.FromGraph {
		return evaluateGraph(in)
	}
	return evaluateGovernance(in, now)
}

func evaluateGraph(in FactorInputs) Evaluation {
	raw := in.GraphScore
	score := raw * graphScoreCap
	if score > graphScoreCap {
		score = graphScoreCap
	}
	if score < 0 {
		score = 0
	}
	scopeMatch := graphScopeMatch(in)
	gs := raw
	sm := scopeMatch
	factors := FactorSet{
		SourceType: string(SourceTypeGraph),
		ScopeMatch: &sm,
		GraphScore: &gs,
	}
	return Evaluation{
		Score:     score,
		Band:      TrustBandLow,
		Factors:   factors,
		Breakdown: graphBreakdown(factors),
	}
}

func evaluateGovernance(in FactorInputs, now time.Time) Evaluation {
	sourceType := deriveSourceType(in.Origin, in.Evidence)
	evidenceStrength := deriveEvidenceStrength(in.Evidence)
	scopeMatch := deriveScopeMatch(in.EntryScope, in.RequestedScope)
	recency := recencyBoost(in.CreatedAt, now)

	base := baseUnverified
	switch in.VerificationStatus {
	case VerificationVerified:
		base = baseVerified
	case VerificationInvalidated:
		base = baseInvalidated
	}

	score := base + originAdjustment(in.Origin) + evidenceWeight*evidenceStrength +
		scopeWeight*scopeMatch + recency
	if in.Superseded {
		score += supersededPenalty
	}
	score = clamp01(score)
	if in.VerificationStatus == VerificationInvalidated && score > invalidatedCap {
		score = invalidatedCap
	}

	supersession := "current"
	if in.Superseded {
		supersession = "superseded"
	}

	recencyCopy := recency
	evidenceCopy := evidenceStrength
	scopeCopy := scopeMatch
	factors := FactorSet{
		Origin:             string(in.Origin),
		VerificationStatus: string(in.VerificationStatus),
		SupersessionStatus: supersession,
		RecencyBoost:       &recencyCopy,
		EvidenceStrength:   &evidenceCopy,
		SourceType:         string(sourceType),
		ScopeMatch:         &scopeCopy,
	}
	band := assignTrustBand(in, sourceType, evidenceStrength)
	return Evaluation{
		Score:     score,
		Band:      band,
		Factors:   factors,
		Breakdown: governanceBreakdown(factors, band),
	}
}

func assignTrustBand(in FactorInputs, sourceType SourceType, evidenceStrength float64) TrustBand {
	if in.VerificationStatus == VerificationInvalidated {
		return TrustBandUntrusted
	}
	if isAgentOrigin(in.Origin) {
		if in.VerificationStatus != VerificationVerified {
			return TrustBandLow
		}
		return TrustBandHigh
	}
	if in.Superseded {
		return TrustBandMedium
	}
	if in.VerificationStatus == VerificationVerified &&
		(sourceType == SourceTypeADR || evidenceStrength == evidenceStrengthADR) {
		return TrustBandCanonical
	}
	if in.VerificationStatus == VerificationVerified {
		return TrustBandHigh
	}
	if isHumanSideOrigin(in.Origin) {
		return TrustBandMedium
	}
	return TrustBandLow
}

func isAgentOrigin(origin KnowledgeOrigin) bool {
	return origin == KnowledgeOriginAgentInference || origin == KnowledgeOriginAgentObservation
}

func isHumanSideOrigin(origin KnowledgeOrigin) bool {
	switch origin {
	case KnowledgeOriginHumanAuthored, KnowledgeOriginHumanVerified,
		KnowledgeOriginArchitectureDecision, KnowledgeOriginImportedSource:
		return true
	default:
		return false
	}
}

func deriveSourceType(origin KnowledgeOrigin, evidence []EvidenceReference) SourceType {
	if origin == KnowledgeOriginArchitectureDecision || hasADREvidence(evidence) {
		return SourceTypeADR
	}
	switch origin {
	case KnowledgeOriginHumanAuthored, KnowledgeOriginHumanVerified:
		return SourceTypeHumanStatement
	case KnowledgeOriginAgentObservation:
		return SourceTypeAgentObservation
	case KnowledgeOriginAgentInference:
		return SourceTypeAgentInference
	case KnowledgeOriginRepositoryObservation:
		return SourceTypeRepoObservation
	case KnowledgeOriginImportedSource:
		return SourceTypeImport
	default:
		return SourceTypeHumanStatement
	}
}

func hasADREvidence(evidence []EvidenceReference) bool {
	for _, ref := range evidence {
		if ref.Type == EvidenceTypeADR {
			return true
		}
	}
	return false
}

func deriveEvidenceStrength(evidence []EvidenceReference) float64 {
	if hasADREvidence(evidence) {
		return evidenceStrengthADR
	}
	if len(evidence) > 0 {
		return evidenceStrengthOther
	}
	return evidenceStrengthNone
}

func deriveScopeMatch(entry Scope, requested *Scope) float64 {
	if entry.Kind == "" || entry.Key == "" {
		return scopeMatchNone
	}
	if requested == nil {
		return scopeMatchExact
	}
	if requested.Kind == entry.Kind && requested.Key == entry.Key {
		return scopeMatchExact
	}
	if requested.Kind == entry.Kind {
		return scopeMatchKindOnly
	}
	return scopeMatchNone
}

func graphScopeMatch(in FactorInputs) float64 {
	if in.RequestedScope == nil || in.EntryScope.Kind == "" || in.EntryScope.Key == "" {
		return scopeMatchNone
	}
	if in.RequestedScope.Kind == in.EntryScope.Kind && in.RequestedScope.Key == in.EntryScope.Key {
		return scopeMatchExact
	}
	if in.RequestedScope.Kind == in.EntryScope.Kind {
		return scopeMatchKindOnly
	}
	return scopeMatchNone
}

func originAdjustment(origin KnowledgeOrigin) float64 {
	switch origin {
	case KnowledgeOriginArchitectureDecision, KnowledgeOriginHumanVerified:
		return 0.08
	case KnowledgeOriginHumanAuthored:
		return 0.05
	case KnowledgeOriginImportedSource:
		return 0.02
	case KnowledgeOriginRepositoryObservation:
		return 0
	case KnowledgeOriginAgentObservation:
		return -0.16
	case KnowledgeOriginAgentInference:
		return -0.22
	default:
		return 0
	}
}

func recencyBoost(createdAt, now time.Time) float64 {
	if createdAt.IsZero() {
		return 0
	}
	age := now.Sub(createdAt.UTC())
	if age < 0 {
		age = 0
	}
	days := age.Hours() / 24
	fraction := days / recencyHorizonDays
	if fraction > 1 {
		fraction = 1
	}
	return maxRecencyBoost * (1 - fraction)
}

func clamp01(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func governanceBreakdown(factors FactorSet, band TrustBand) []string {
	lines := []string{
		fmt.Sprintf("verification_status=%s", factors.VerificationStatus),
		fmt.Sprintf("origin=%s", factors.Origin),
		fmt.Sprintf("supersession_status=%s", factors.SupersessionStatus),
		fmt.Sprintf("source_type=%s", factors.SourceType),
	}
	if factors.EvidenceStrength != nil {
		lines = append(lines, fmt.Sprintf("evidence_strength=%.2f", *factors.EvidenceStrength))
	}
	if factors.ScopeMatch != nil {
		lines = append(lines, fmt.Sprintf("scope_match=%.2f", *factors.ScopeMatch))
	}
	if factors.RecencyBoost != nil {
		lines = append(lines, fmt.Sprintf("recency_boost=%.2f", *factors.RecencyBoost))
	}
	lines = append(lines, fmt.Sprintf("trust_band=%s", band))
	return lines
}

func graphBreakdown(factors FactorSet) []string {
	lines := []string{
		fmt.Sprintf("source_type=%s", factors.SourceType),
	}
	if factors.GraphScore != nil {
		lines = append(lines, fmt.Sprintf("graph_score=%.2f", *factors.GraphScore))
	}
	if factors.ScopeMatch != nil {
		lines = append(lines, fmt.Sprintf("scope_match=%.2f", *factors.ScopeMatch))
	}
	lines = append(lines, fmt.Sprintf("trust_band=%s", TrustBandLow))
	return lines
}

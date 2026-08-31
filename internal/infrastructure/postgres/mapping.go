package postgres

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

type evidenceJSON struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func loreEntryToInsertParams(entry domain.LoreEntry) (sqlc.InsertLoreEntryParams, error) {
	evidence, err := marshalEvidence(entry.Evidence)
	if err != nil {
		return sqlc.InsertLoreEntryParams{}, err
	}
	return sqlc.InsertLoreEntryParams{
		ID:                 entry.ID,
		Statement:          entry.Statement,
		ScopeKind:          string(entry.Scope.Kind),
		ScopeKey:           entry.Scope.Key,
		Origin:             string(entry.Origin),
		VerificationStatus: string(entry.VerificationStatus),
		Evidence:           evidence,
		CreatedBy:          entry.CreatedBy,
		CreatedAt:          timestamptzFromTime(entry.CreatedAt),
		VerifiedBy:         textFromPtr(entry.VerifiedBy),
		VerifiedAt:         timestamptzFromPtr(entry.VerifiedAt),
		UpdatedAt:          timestamptzFromTime(entry.UpdatedAt),
		SupersededByID:     textFromPtr(entry.SupersededByID),
		InvalidatedBy:      textFromPtr(entry.InvalidatedBy),
		InvalidatedAt:      timestamptzFromPtr(entry.InvalidatedAt),
	}, nil
}

func loreEntryToUpdateParams(entry domain.LoreEntry) (sqlc.UpdateLoreEntryParams, error) {
	evidence, err := marshalEvidence(entry.Evidence)
	if err != nil {
		return sqlc.UpdateLoreEntryParams{}, err
	}
	return sqlc.UpdateLoreEntryParams{
		ID:                 entry.ID,
		Statement:          entry.Statement,
		ScopeKind:          string(entry.Scope.Kind),
		ScopeKey:           entry.Scope.Key,
		Origin:             string(entry.Origin),
		VerificationStatus: string(entry.VerificationStatus),
		Evidence:           evidence,
		CreatedBy:          entry.CreatedBy,
		CreatedAt:          timestamptzFromTime(entry.CreatedAt),
		VerifiedBy:         textFromPtr(entry.VerifiedBy),
		VerifiedAt:         timestamptzFromPtr(entry.VerifiedAt),
		UpdatedAt:          timestamptzFromTime(entry.UpdatedAt),
		SupersededByID:     textFromPtr(entry.SupersededByID),
		InvalidatedBy:      textFromPtr(entry.InvalidatedBy),
		InvalidatedAt:      timestamptzFromPtr(entry.InvalidatedAt),
	}, nil
}

func loreEntryFromRow(row sqlc.LoreEntry) (domain.LoreEntry, error) {
	kind, err := domain.ParseScopeKind(row.ScopeKind)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	scope := domain.Scope{Kind: kind, Key: row.ScopeKey}
	evidence, err := unmarshalEvidence(row.Evidence)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	return domain.LoreEntry{
		ID:                 row.ID,
		Statement:          row.Statement,
		Scope:              scope,
		Origin:             domain.KnowledgeOrigin(row.Origin),
		VerificationStatus: domain.VerificationStatus(row.VerificationStatus),
		Evidence:           evidence,
		CreatedBy:          row.CreatedBy,
		CreatedAt:          timeFromTimestamptz(row.CreatedAt),
		VerifiedBy:         ptrFromText(row.VerifiedBy),
		VerifiedAt:         ptrFromTimestamptz(row.VerifiedAt),
		UpdatedAt:          timeFromTimestamptz(row.UpdatedAt),
		SupersededByID:     ptrFromText(row.SupersededByID),
		InvalidatedBy:      ptrFromText(row.InvalidatedBy),
		InvalidatedAt:      ptrFromTimestamptz(row.InvalidatedAt),
	}, nil
}

func auditRecordToInsertParams(record domain.AuditRecord) sqlc.InsertAuditRecordParams {
	return sqlc.InsertAuditRecordParams{
		ID:        record.ID,
		TargetID:  record.TargetID,
		Action:    string(record.Action),
		ActorID:   record.ActorID,
		CreatedAt: timestamptzFromTime(record.CreatedAt),
	}
}

func auditRecordFromRow(row sqlc.AuditRecord) domain.AuditRecord {
	return domain.AuditRecord{
		ID:        row.ID,
		TargetID:  row.TargetID,
		Action:    domain.AuditAction(row.Action),
		ActorID:   row.ActorID,
		CreatedAt: timeFromTimestamptz(row.CreatedAt),
	}
}

func marshalEvidence(items []domain.EvidenceReference) ([]byte, error) {
	payload := make([]evidenceJSON, len(items))
	for i, item := range items {
		payload[i] = evidenceJSON{Type: string(item.Type), Value: item.Value}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal evidence: %w", err)
	}
	return raw, nil
}

func unmarshalEvidence(raw []byte) ([]domain.EvidenceReference, error) {
	if len(raw) == 0 {
		return []domain.EvidenceReference{}, nil
	}
	var payload []evidenceJSON
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal evidence: %w", err)
	}
	out := make([]domain.EvidenceReference, len(payload))
	for i, item := range payload {
		evidenceType, err := domain.ParseEvidenceType(item.Type)
		if err != nil {
			return nil, err
		}
		out[i] = domain.EvidenceReference{Type: evidenceType, Value: item.Value}
	}
	return out, nil
}

func timestamptzFromTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timestamptzFromPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timeFromTimestamptz(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func ptrFromTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func textFromPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func ptrFromText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	s := value.String
	return &s
}

func textFromString(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: value, Valid: true}
}

func outboxEventToInsertParams(event domain.OutboxEvent) (sqlc.InsertOutboxEventParams, error) {
	return sqlc.InsertOutboxEventParams{
		ID:             event.ID,
		AggregateType:  event.AggregateType,
		AggregateID:    event.AggregateID,
		EventType:      string(event.EventType),
		Payload:        event.Payload,
		Status:         string(event.Status),
		Attempts:       int32(event.Attempts),
		IdempotencyKey: event.IdempotencyKey,
		CreatedAt:      timestamptzFromTime(event.CreatedAt),
		ProcessedAt:    timestamptzFromPtr(event.ProcessedAt),
		LastError:      textFromString(event.LastError),
	}, nil
}

func outboxEventFromRow(row sqlc.OutboxEvent) (domain.OutboxEvent, error) {
	return domain.OutboxEvent{
		ID:             row.ID,
		AggregateType:  row.AggregateType,
		AggregateID:    row.AggregateID,
		EventType:      domain.OutboxEventType(row.EventType),
		Payload:        row.Payload,
		Status:         domain.OutboxStatus(row.Status),
		Attempts:       int(row.Attempts),
		IdempotencyKey: row.IdempotencyKey,
		CreatedAt:      timeFromTimestamptz(row.CreatedAt),
		ProcessedAt:    ptrFromTimestamptz(row.ProcessedAt),
		LastError:      stringFromText(row.LastError),
	}, nil
}

func stringFromText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func ingestRunToInsertParams(run domain.IngestRun) sqlc.InsertGitIngestRunParams {
	return sqlc.InsertGitIngestRunParams{
		ID:               run.ID,
		ScopeKind:        string(run.Scope.Kind),
		ScopeKey:         run.Scope.Key,
		ActorID:          run.ActorID,
		LocalPath:        run.LocalPath,
		Status:           string(run.Status),
		CommitsSeen:      int32(run.CommitsSeen),
		CommitsSkipped:   int32(run.CommitsSkipped),
		CandidatesStored: int32(run.CandidatesStored),
		CursorSha:        textFromString(run.CursorSHA),
		CursorAt:         timestamptzFromPtr(run.CursorAt),
		ErrorSummary:     textFromString(run.ErrorSummary),
		StartedAt:        timestamptzFromTime(run.StartedAt),
		FinishedAt:       timestamptzFromPtr(run.FinishedAt),
	}
}

func ingestRunToUpdateParams(run domain.IngestRun) sqlc.UpdateGitIngestRunParams {
	return sqlc.UpdateGitIngestRunParams{
		ID:               run.ID,
		Status:           string(run.Status),
		CommitsSeen:      int32(run.CommitsSeen),
		CommitsSkipped:   int32(run.CommitsSkipped),
		CandidatesStored: int32(run.CandidatesStored),
		CursorSha:        textFromString(run.CursorSHA),
		CursorAt:         timestamptzFromPtr(run.CursorAt),
		ErrorSummary:     textFromString(run.ErrorSummary),
		FinishedAt:       timestamptzFromPtr(run.FinishedAt),
	}
}

func ingestRunFromRow(row sqlc.GitIngestRun) (domain.IngestRun, error) {
	kind, err := domain.ParseScopeKind(row.ScopeKind)
	if err != nil {
		return domain.IngestRun{}, err
	}
	return domain.IngestRun{
		ID:               row.ID,
		Scope:            domain.Scope{Kind: kind, Key: row.ScopeKey},
		ActorID:          row.ActorID,
		LocalPath:        row.LocalPath,
		Status:           domain.IngestRunStatus(row.Status),
		CommitsSeen:      int(row.CommitsSeen),
		CommitsSkipped:   int(row.CommitsSkipped),
		CandidatesStored: int(row.CandidatesStored),
		CursorSHA:        stringFromText(row.CursorSha),
		CursorAt:         ptrFromTimestamptz(row.CursorAt),
		ErrorSummary:     stringFromText(row.ErrorSummary),
		StartedAt:        timeFromTimestamptz(row.StartedAt),
		FinishedAt:       ptrFromTimestamptz(row.FinishedAt),
	}, nil
}

func ingestCursorFromRow(row sqlc.GitIngestCursor) domain.IngestCursor {
	return domain.IngestCursor{
		Scope:           domain.Scope{Kind: domain.ScopeKind(row.ScopeKind), Key: row.ScopeKey},
		LastSHA:         row.LastSha,
		LastCommittedAt: timeFromTimestamptz(row.LastCommittedAt),
		UpdatedAt:       timeFromTimestamptz(row.UpdatedAt),
	}
}

func processedSHAToInsertParams(row domain.ProcessedSHA) sqlc.InsertGitIngestSHAParams {
	return sqlc.InsertGitIngestSHAParams{
		ScopeKind:   string(row.Scope.Kind),
		ScopeKey:    row.Scope.Key,
		Sha:         row.SHA,
		LoreEntryID: textFromString(row.LoreEntryID),
		Skipped:     row.Skipped,
		SkipReason:  textFromString(row.SkipReason),
		ProcessedAt: timestamptzFromTime(row.ProcessedAt),
	}
}

func processedSHAFromRow(row sqlc.GitIngestSha) domain.ProcessedSHA {
	return domain.ProcessedSHA{
		Scope:       domain.Scope{Kind: domain.ScopeKind(row.ScopeKind), Key: row.ScopeKey},
		SHA:         row.Sha,
		LoreEntryID: stringFromText(row.LoreEntryID),
		Skipped:     row.Skipped,
		SkipReason:  stringFromText(row.SkipReason),
		ProcessedAt: timeFromTimestamptz(row.ProcessedAt),
	}
}

func int4FromInt(value int) pgtype.Int4 {
	if value == 0 {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(value), Valid: true}
}

func intFromInt4(value pgtype.Int4) int {
	if !value.Valid {
		return 0
	}
	return int(value.Int32)
}

func prIngestRunToInsertParams(run domain.PRIngestRun) sqlc.InsertPRIngestRunParams {
	return sqlc.InsertPRIngestRunParams{
		ID:               run.ID,
		ScopeKind:        string(run.Scope.Kind),
		ScopeKey:         run.Scope.Key,
		ActorID:          run.ActorID,
		PrNumber:         int32(run.PRNumber),
		Status:           string(run.Status),
		PrsSeen:          int32(run.PRsSeen),
		PrsSkipped:       int32(run.PRsSkipped),
		CandidatesStored: int32(run.CandidatesStored),
		CursorPr:         int4FromInt(run.CursorPR),
		CursorAt:         timestamptzFromPtr(run.CursorAt),
		ErrorSummary:     textFromString(run.ErrorSummary),
		StartedAt:        timestamptzFromTime(run.StartedAt),
		FinishedAt:       timestamptzFromPtr(run.FinishedAt),
	}
}

func prIngestRunToUpdateParams(run domain.PRIngestRun) sqlc.UpdatePRIngestRunParams {
	return sqlc.UpdatePRIngestRunParams{
		ID:               run.ID,
		Status:           string(run.Status),
		PrsSeen:          int32(run.PRsSeen),
		PrsSkipped:       int32(run.PRsSkipped),
		CandidatesStored: int32(run.CandidatesStored),
		CursorPr:         int4FromInt(run.CursorPR),
		CursorAt:         timestamptzFromPtr(run.CursorAt),
		ErrorSummary:     textFromString(run.ErrorSummary),
		FinishedAt:       timestamptzFromPtr(run.FinishedAt),
	}
}

func prIngestRunFromRow(row sqlc.PrIngestRun) (domain.PRIngestRun, error) {
	kind, err := domain.ParseScopeKind(row.ScopeKind)
	if err != nil {
		return domain.PRIngestRun{}, err
	}
	return domain.PRIngestRun{
		ID:               row.ID,
		Scope:            domain.Scope{Kind: kind, Key: row.ScopeKey},
		ActorID:          row.ActorID,
		PRNumber:         int(row.PrNumber),
		Status:           domain.IngestRunStatus(row.Status),
		PRsSeen:          int(row.PrsSeen),
		PRsSkipped:       int(row.PrsSkipped),
		CandidatesStored: int(row.CandidatesStored),
		CursorPR:         intFromInt4(row.CursorPr),
		CursorAt:         ptrFromTimestamptz(row.CursorAt),
		ErrorSummary:     stringFromText(row.ErrorSummary),
		StartedAt:        timeFromTimestamptz(row.StartedAt),
		FinishedAt:       ptrFromTimestamptz(row.FinishedAt),
	}, nil
}

func prIngestCursorFromRow(row sqlc.PrIngestCursor) domain.PRIngestCursor {
	return domain.PRIngestCursor{
		Scope:        domain.Scope{Kind: domain.ScopeKind(row.ScopeKind), Key: row.ScopeKey},
		LastPR:       int(row.LastPr),
		LastMergedAt: timeFromTimestamptz(row.LastMergedAt),
		UpdatedAt:    timeFromTimestamptz(row.UpdatedAt),
	}
}

func processedPRToInsertParams(row domain.ProcessedPR) sqlc.InsertPRIngestPRParams {
	return sqlc.InsertPRIngestPRParams{
		ScopeKind:   string(row.Scope.Kind),
		ScopeKey:    row.Scope.Key,
		PrNumber:    int32(row.PRNumber),
		NodeID:      textFromString(row.NodeID),
		LoreEntryID: textFromString(row.LoreEntryID),
		Skipped:     row.Skipped,
		SkipReason:  textFromString(row.SkipReason),
		ProcessedAt: timestamptzFromTime(row.ProcessedAt),
	}
}

func processedPRFromRow(row sqlc.PrIngestPr) domain.ProcessedPR {
	return domain.ProcessedPR{
		Scope:       domain.Scope{Kind: domain.ScopeKind(row.ScopeKind), Key: row.ScopeKey},
		PRNumber:    int(row.PrNumber),
		NodeID:      stringFromText(row.NodeID),
		LoreEntryID: stringFromText(row.LoreEntryID),
		Skipped:     row.Skipped,
		SkipReason:  stringFromText(row.SkipReason),
		ProcessedAt: timeFromTimestamptz(row.ProcessedAt),
	}
}

func joinExtraDirs(dirs []string) string {
	if len(dirs) == 0 {
		return ""
	}
	return strings.Join(dirs, "\n")
}

func splitExtraDirs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func adrIngestRunToInsertParams(run domain.ADRIngestRun) sqlc.InsertADRIngestRunParams {
	return sqlc.InsertADRIngestRunParams{
		ID:             run.ID,
		ScopeKind:      string(run.Scope.Kind),
		ScopeKey:       run.Scope.Key,
		ActorID:        run.ActorID,
		LocalPath:      run.LocalPath,
		ExtraDirs:      joinExtraDirs(run.ExtraDirs),
		Status:         string(run.Status),
		FilesSeen:      int32(run.FilesSeen),
		FilesSkipped:   int32(run.FilesSkipped),
		LoreStored:     int32(run.LoreStored),
		LoreSuperseded: int32(run.LoreSuperseded),
		ErrorSummary:   textFromString(run.ErrorSummary),
		StartedAt:      timestamptzFromTime(run.StartedAt),
		FinishedAt:     timestamptzFromPtr(run.FinishedAt),
	}
}

func adrIngestRunToUpdateParams(run domain.ADRIngestRun) sqlc.UpdateADRIngestRunParams {
	return sqlc.UpdateADRIngestRunParams{
		ID:             run.ID,
		Status:         string(run.Status),
		FilesSeen:      int32(run.FilesSeen),
		FilesSkipped:   int32(run.FilesSkipped),
		LoreStored:     int32(run.LoreStored),
		LoreSuperseded: int32(run.LoreSuperseded),
		ErrorSummary:   textFromString(run.ErrorSummary),
		FinishedAt:     timestamptzFromPtr(run.FinishedAt),
	}
}

func adrIngestRunFromRow(row sqlc.AdrIngestRun) (domain.ADRIngestRun, error) {
	kind, err := domain.ParseScopeKind(row.ScopeKind)
	if err != nil {
		return domain.ADRIngestRun{}, err
	}
	return domain.ADRIngestRun{
		ID:             row.ID,
		Scope:          domain.Scope{Kind: kind, Key: row.ScopeKey},
		ActorID:        row.ActorID,
		LocalPath:      row.LocalPath,
		ExtraDirs:      splitExtraDirs(row.ExtraDirs),
		Status:         domain.IngestRunStatus(row.Status),
		FilesSeen:      int(row.FilesSeen),
		FilesSkipped:   int(row.FilesSkipped),
		LoreStored:     int(row.LoreStored),
		LoreSuperseded: int(row.LoreSuperseded),
		ErrorSummary:   stringFromText(row.ErrorSummary),
		StartedAt:      timeFromTimestamptz(row.StartedAt),
		FinishedAt:     ptrFromTimestamptz(row.FinishedAt),
	}, nil
}

func adrIngestCursorFromRow(row sqlc.AdrIngestCursor) domain.ADRIngestCursor {
	return domain.ADRIngestCursor{
		Scope:        domain.Scope{Kind: domain.ScopeKind(row.ScopeKind), Key: row.ScopeKey},
		LastPath:     stringFromText(row.LastPath),
		LastChecksum: stringFromText(row.LastChecksum),
		UpdatedAt:    timeFromTimestamptz(row.UpdatedAt),
	}
}

func processedADRToInsertParams(row domain.ProcessedADR) sqlc.InsertADRIngestFileParams {
	return sqlc.InsertADRIngestFileParams{
		ScopeKind:    string(row.Scope.Kind),
		ScopeKey:     row.Scope.Key,
		RelativePath: row.RelativePath,
		Checksum:     row.Checksum,
		AdrID:        textFromString(row.ADRID),
		LoreEntryID:  textFromString(row.LoreEntryID),
		Skipped:      row.Skipped,
		SkipReason:   textFromString(row.SkipReason),
		ProcessedAt:  timestamptzFromTime(row.ProcessedAt),
	}
}

func processedADRFromRow(row sqlc.AdrIngestFile) domain.ProcessedADR {
	return domain.ProcessedADR{
		Scope:        domain.Scope{Kind: domain.ScopeKind(row.ScopeKind), Key: row.ScopeKey},
		RelativePath: row.RelativePath,
		Checksum:     row.Checksum,
		ADRID:        stringFromText(row.AdrID),
		LoreEntryID:  stringFromText(row.LoreEntryID),
		Skipped:      row.Skipped,
		SkipReason:   stringFromText(row.SkipReason),
		ProcessedAt:  timeFromTimestamptz(row.ProcessedAt),
	}
}

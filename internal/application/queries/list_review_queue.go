package queries

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

const ReviewStatusPending = "pending"

// SuggestedLoreItem is a review-queue projection of observational lore.
type SuggestedLoreItem struct {
	Entry       domain.LoreEntry
	SourceType  string
	Status      string
	SuccessorID *string
	ActorID     string
}

// ListReviewQueueQuery lists pending suggested lore for a repository.
type ListReviewQueueQuery struct {
	Scope domain.Scope
}

// GetReviewItemQuery loads one review item by observational lore id.
type GetReviewItemQuery struct {
	ID string
}

// ListReviewQueueHandler projects pending git/PR observational lore.
type ListReviewQueueHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListReviewQueueHandler(begin ports.UnitOfWorkFactory) *ListReviewQueueHandler {
	return &ListReviewQueueHandler{begin: begin}
}

func (h *ListReviewQueueHandler) Handle(ctx context.Context, q ListReviewQueueQuery) ([]SuggestedLoreItem, error) {
	if q.Scope.Kind != domain.ScopeKindRepository {
		return nil, &domain.ValidationError{Message: "review queue scope kind must be repository"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	entries, err := uow.LoreEntries().ListByScope(ctx, q.Scope)
	if err != nil {
		return nil, err
	}
	decisions, err := uow.ReviewDecisions().ListByScope(ctx, q.Scope)
	if err != nil {
		return nil, err
	}
	blocked := make(map[string]struct{}, len(decisions))
	for _, d := range decisions {
		blocked[identityKey(d)] = struct{}{}
	}

	out := make([]SuggestedLoreItem, 0)
	for _, entry := range entries {
		if !domain.IsReviewEligible(entry) {
			continue
		}
		id, err := domain.ExtractIdentityOf(entry)
		if err != nil {
			continue
		}
		if _, ok := blocked[id.Key()]; ok {
			continue
		}
		out = append(out, SuggestedLoreItem{
			Entry:      entry,
			SourceType: domain.ReviewSourceType(entry),
			Status:     ReviewStatusPending,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Entry.CreatedAt.Equal(out[j].Entry.CreatedAt) {
			return out[i].Entry.CreatedAt.After(out[j].Entry.CreatedAt)
		}
		return out[i].Entry.ID < out[j].Entry.ID
	})
	return out, nil
}

// GetReviewItemHandler loads a pending or decided review item.
type GetReviewItemHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewGetReviewItemHandler(begin ports.UnitOfWorkFactory) *GetReviewItemHandler {
	return &GetReviewItemHandler{begin: begin}
}

func (h *GetReviewItemHandler) Handle(ctx context.Context, q GetReviewItemQuery) (SuggestedLoreItem, error) {
	id := strings.TrimSpace(q.ID)
	if id == "" {
		return SuggestedLoreItem{}, &domain.ValidationError{Message: "id must be non-empty"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return SuggestedLoreItem{}, err
	}
	defer uow.Rollback(ctx)

	entry, err := uow.LoreEntries().Get(ctx, id)
	if err != nil {
		return SuggestedLoreItem{}, &domain.NotFoundError{Message: fmt.Sprintf("review item %s not found", id)}
	}
	decision, ok, err := uow.ReviewDecisions().GetByLoreID(ctx, entry.ID)
	if err != nil {
		return SuggestedLoreItem{}, err
	}
	if ok {
		return SuggestedLoreItem{
			Entry:       entry,
			SourceType:  domain.ReviewSourceType(entry),
			Status:      string(decision.Status),
			SuccessorID: decision.SuccessorLoreID,
			ActorID:     decision.ActorID,
		}, nil
	}
	if !domain.IsReviewEligible(entry) {
		return SuggestedLoreItem{}, &domain.NotFoundError{Message: fmt.Sprintf("review item %s not found", id)}
	}
	return SuggestedLoreItem{
		Entry:      entry,
		SourceType: domain.ReviewSourceType(entry),
		Status:     ReviewStatusPending,
	}, nil
}

func identityKey(d domain.ReviewDecision) string {
	return domain.ExtractIdentity{
		Scope:             d.Scope,
		EvidenceType:      d.EvidenceType,
		EvidenceValue:     d.EvidenceValue,
		StatementChecksum: d.StatementChecksum,
	}.Key()
}

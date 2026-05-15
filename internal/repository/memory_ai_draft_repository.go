package repository

import (
	"context"
	"time"

	"requirement-splitting/internal/domain"
)

type MemoryAIDraftRepository struct {
	drafts map[string]domain.AIDraft
	now    func() time.Time
	newID  func() string
}

func NewMemoryAIDraftRepository() *MemoryAIDraftRepository {
	return &MemoryAIDraftRepository{
		drafts: map[string]domain.AIDraft{},
		now:    time.Now,
		newID:  newMemoryID,
	}
}

func (r *MemoryAIDraftRepository) CreateAIDraft(ctx context.Context, draft domain.AIDraft) (domain.AIDraft, error) {
	if draft.ID == "" {
		draft.ID = r.newID()
	}
	now := r.now()
	draft.CreatedAt = now
	draft.UpdatedAt = now
	r.drafts[draft.ID] = draft
	return draft, nil
}

func (r *MemoryAIDraftRepository) ListAIDrafts(ctx context.Context, projectID string) ([]domain.AIDraft, error) {
	drafts := make([]domain.AIDraft, 0)
	for _, draft := range r.drafts {
		if draft.ProjectID == projectID {
			drafts = append(drafts, draft)
		}
	}
	return drafts, nil
}

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/pramodksahoo/kubechat/backend/internal/plan"
	"github.com/maypok86/otter/v2"
)

type PlanRecord struct {
	Plan      plan.PlanDraft `json:"plan"`
	StoredAt  time.Time      `json:"storedAt"`
	ExpiresAt time.Time      `json:"expiresAt"`
}

type ErrPlanNotFound struct {
	ID string
}

func (e ErrPlanNotFound) Error() string {
	return fmt.Sprintf("plan %s not found", e.ID)
}

type PlanRepository struct {
	cache *otter.Cache[string, any]
	ttl   time.Duration
}

func NewPlanRepository(cache *otter.Cache[string, any], ttl time.Duration) *PlanRepository {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &PlanRepository{cache: cache, ttl: ttl}
}

func (r *PlanRepository) key(id string) string {
	return "plan:" + id
}

func (r *PlanRepository) Save(ctx context.Context, draft plan.PlanDraft) (PlanRecord, error) {
	_ = ctx
	now := time.Now().UTC()
	record := PlanRecord{
		Plan:      draft,
		StoredAt:  now,
		ExpiresAt: now.Add(r.ttl),
	}

	key := r.key(draft.ID)
	r.cache.Set(key, record)
	r.cache.SetExpiresAfter(key, r.ttl)
	return record, nil
}

func (r *PlanRepository) Get(ctx context.Context, id string) (PlanRecord, error) {
	_ = ctx
	value, ok := r.cache.GetIfPresent(r.key(id))
	if !ok {
		return PlanRecord{}, ErrPlanNotFound{ID: id}
	}
	record, ok := value.(PlanRecord)
	if !ok {
		return PlanRecord{}, fmt.Errorf("invalid plan record type for %s", id)
	}
	return record, nil
}

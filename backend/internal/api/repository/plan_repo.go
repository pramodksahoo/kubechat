package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/pramodksahoo/kubechat/backend/internal/plan"
)

type PlanChange struct {
	Field        string `json:"field"`
	Before       any    `json:"before,omitempty"`
	After        any    `json:"after,omitempty"`
	Resource     string `json:"resource,omitempty"`
	StepSequence int    `json:"stepSequence,omitempty"`
}

type PlanRevision struct {
	Version   int          `json:"version"`
	UpdatedAt time.Time    `json:"updatedAt"`
	UpdatedBy string       `json:"updatedBy,omitempty"`
	Changes   []PlanChange `json:"changes"`
}

type PlanRecord struct {
	Plan      plan.PlanDraft `json:"plan"`
	StoredAt  time.Time      `json:"storedAt"`
	ExpiresAt time.Time      `json:"expiresAt"`
	Revisions []PlanRevision `json:"revisions,omitempty"`
}

type PlanUpdate struct {
	TargetNamespace  *string
	Labels           map[string]string
	ReplicaOverrides map[string]int
	UpdatedBy        string
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

	plan.ApplyParameters(&draft, draft.Parameters)

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

func (r *PlanRepository) Update(ctx context.Context, id string, update PlanUpdate) (PlanRecord, bool, error) {
	_ = ctx
	key := r.key(id)
	value, ok := r.cache.GetIfPresent(key)
	if !ok {
		return PlanRecord{}, false, ErrPlanNotFound{ID: id}
	}
	record, ok := value.(PlanRecord)
	if !ok {
		return PlanRecord{}, false, fmt.Errorf("invalid plan record type for %s", id)
	}

	originalParams := cloneParameters(record.Plan.Parameters)
	updatedPlan := clonePlan(record.Plan)

	newParams := cloneParameters(updatedPlan.Parameters)

	if update.TargetNamespace != nil {
		newParams.Namespace = strings.TrimSpace(*update.TargetNamespace)
	}
	if update.Labels != nil {
		newParams.Labels = copyStringMap(update.Labels)
	}
	if update.ReplicaOverrides != nil {
		newParams.ReplicaOverrides = copyIntMap(update.ReplicaOverrides)
	}

	plan.ApplyParameters(&updatedPlan, newParams)

	changes := diffParameters(originalParams, updatedPlan.Parameters, updatedPlan.Steps)
	if len(changes) == 0 {
		return record, false, nil
	}

	now := time.Now().UTC()
	updatedBy := strings.TrimSpace(update.UpdatedBy)
	if updatedBy == "" {
		updatedBy = "unknown"
	}

	revision := PlanRevision{
		Version:   len(record.Revisions) + 1,
		UpdatedAt: now,
		UpdatedBy: updatedBy,
		Changes:   changes,
	}

	record.Plan = updatedPlan
	record.Revisions = append(copyRevisions(record.Revisions), revision)
	record.ExpiresAt = now.Add(r.ttl)

	r.cache.Set(key, record)
	r.cache.SetExpiresAfter(key, r.ttl)

	return record, true, nil
}

func diffParameters(old, current plan.Parameters, steps []plan.PlanStep) []PlanChange {
	var changes []PlanChange
	if old.Namespace != current.Namespace {
		changes = append(changes, PlanChange{
			Field:  "targetNamespace",
			Before: old.Namespace,
			After:  current.Namespace,
		})
	}

	labelChanges := diffStringMaps(old.Labels, current.Labels, "label")
	changes = append(changes, labelChanges...)

	replicaChanges := diffIntMaps(old.ReplicaOverrides, current.ReplicaOverrides, steps)
	changes = append(changes, replicaChanges...)

	return changes
}

func diffStringMaps(old, current map[string]string, field string) []PlanChange {
	keys := unionKeysString(old, current)
	var changes []PlanChange
	for _, key := range keys {
		before, beforeOk := old[key]
		after, afterOk := current[key]
		if beforeOk && afterOk && before == after {
			continue
		}
		change := PlanChange{
			Field:    field,
			Resource: key,
		}
		if beforeOk {
			change.Before = before
		}
		if afterOk {
			change.After = after
		}
		changes = append(changes, change)
	}
	return changes
}

func diffIntMaps(old, current map[string]int, steps []plan.PlanStep) []PlanChange {
	keys := unionKeysInt(old, current)
	var changes []PlanChange
	stepIndexByResource := map[string]int{}
	for _, step := range steps {
		stepIndexByResource[step.Target.Resource] = step.Sequence
	}

	for _, key := range keys {
		before, beforeOk := old[key]
		after, afterOk := current[key]
		if beforeOk && afterOk && before == after {
			continue
		}
		change := PlanChange{
			Field:    "replicaOverride",
			Resource: key,
		}
		if seq, ok := stepIndexByResource[key]; ok {
			change.StepSequence = seq
		}
		if beforeOk {
			change.Before = before
		}
		if afterOk {
			change.After = after
		}
		changes = append(changes, change)
	}
	return changes
}

func unionKeysString(left, right map[string]string) []string {
	keys := make(map[string]struct{})
	for k := range left {
		keys[k] = struct{}{}
	}
	for k := range right {
		keys[k] = struct{}{}
	}
	return sortedKeys(keys)
}

func unionKeysInt(left, right map[string]int) []string {
	keys := make(map[string]struct{})
	for k := range left {
		keys[k] = struct{}{}
	}
	for k := range right {
		keys[k] = struct{}{}
	}
	return sortedKeys(keys)
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func copyStringMap(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyIntMap(in map[string]int) map[string]int {
	if in == nil {
		return map[string]int{}
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneParameters(params plan.Parameters) plan.Parameters {
	return plan.Parameters{
		Namespace:        params.Namespace,
		Labels:           copyStringMap(params.Labels),
		ReplicaOverrides: copyIntMap(params.ReplicaOverrides),
	}
}

func copyRevisions(in []PlanRevision) []PlanRevision {
	if len(in) == 0 {
		return nil
	}
	out := make([]PlanRevision, len(in))
	copy(out, in)
	return out
}

func clonePlan(in plan.PlanDraft) plan.PlanDraft {
	out := in
	out.ScopeSignals = copyStringMap(in.ScopeSignals)
	out.Parameters = cloneParameters(in.Parameters)
	out.RiskSummary = plan.RiskSummary{
		Level:          in.RiskSummary.Level,
		Justifications: append([]string(nil), in.RiskSummary.Justifications...),
	}
	out.Steps = make([]plan.PlanStep, len(in.Steps))
	for i, step := range in.Steps {
		out.Steps[i] = clonePlanStep(step)
	}
	return out
}

func clonePlanStep(step plan.PlanStep) plan.PlanStep {
	out := step
	out.AffectedResources = append([]string(nil), step.AffectedResources...)
	out.Target = plan.TargetDescriptor{
		Cluster:   step.Target.Cluster,
		Namespace: step.Target.Namespace,
		Resource:  step.Target.Resource,
	}
	if step.DiffPreview != nil {
		out.DiffPreview = cloneAny(step.DiffPreview).(map[string]any)
	}
	return out
}

func cloneAny(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = cloneAny(val)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = cloneAny(val)
		}
		return out
	default:
		return v
	}
}

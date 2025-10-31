package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/pramodksahoo/kubechat/backend/internal/plan"
)

func TestPlanRepositoryUpdatePersistsParametersAndRevisions(t *testing.T) {
	cache := otter.Must(&otter.Options[string, any]{MaximumSize: 64})
	repo := NewPlanRepository(cache, time.Hour)

	draft := plan.PlanDraft{
		ID:              "plan-123",
		Prompt:          "scale prod deployments",
		TargetCluster:   "prod",
		TargetNamespace: "default",
		ScopeSignals:    map[string]string{"target_namespace": "default"},
		Steps: []plan.PlanStep{
			{
				Sequence: 1,
				Title:    "Inspect pods",
				Command:  "kubectl get pods --namespace=default --context=prod",
				Target:   plan.TargetDescriptor{Cluster: "prod", Namespace: "default", Resource: "pods"},
			},
			{
				Sequence: 2,
				Title:    "Scale deployment",
				Command:  "kubectl scale deploy/foo --replicas=<desired> --namespace=default --context=prod",
				Target:   plan.TargetDescriptor{Cluster: "prod", Namespace: "default", Resource: "deployments/foo"},
				DiffPreview: map[string]any{
					"replicas": map[string]any{"from": 2},
				},
			},
		},
		Parameters: plan.Parameters{
			Namespace:        "default",
			Labels:           map[string]string{},
			ReplicaOverrides: map[string]int{},
		},
	}

	if _, err := repo.Save(context.Background(), draft); err != nil {
		t.Fatalf("failed to save plan: %v", err)
	}

	ns := "payments"
	record, changed, err := repo.Update(context.Background(), draft.ID, PlanUpdate{
		TargetNamespace:  &ns,
		Labels:           map[string]string{"app": "checkout"},
		ReplicaOverrides: map[string]int{"deployments/foo": 5},
		UpdatedBy:        "unit-test",
	})
	if err != nil {
		t.Fatalf("unexpected error updating plan: %v", err)
	}
	if !changed {
		t.Fatal("expected update to report changes")
	}
	if record.Plan.TargetNamespace != "payments" {
		t.Fatalf("expected namespace to update, got %s", record.Plan.TargetNamespace)
	}
	if record.Plan.Parameters.Namespace != "payments" {
		t.Fatalf("expected parameters namespace to update, got %s", record.Plan.Parameters.Namespace)
	}
	if record.Plan.ScopeSignals["label_selector"] != "app=checkout" {
		t.Fatalf("expected scope signals to reflect labels, got %+v", record.Plan.ScopeSignals)
	}
	if len(record.Plan.Steps) == 0 {
		t.Fatalf("expected plan steps to persist")
	}
	scaleStep := record.Plan.Steps[1]
	if scaleStep.Target.Namespace != "payments" {
		t.Fatalf("expected step namespace to update, got %s", scaleStep.Target.Namespace)
	}
	if !containsAll(scaleStep.Command, []string{"--namespace=payments", "--replicas=5"}) {
		t.Fatalf("expected updated scale command, got %s", scaleStep.Command)
	}
	replicas, ok := scaleStep.DiffPreview["replicas"].(map[string]any)
	if !ok {
		t.Fatalf("expected replica diff preview map, got %T", scaleStep.DiffPreview["replicas"])
	}
	if replicas["to"] != 5 {
		t.Fatalf("expected replica override to update diff preview, got %+v", replicas)
	}

	if len(record.Revisions) != 1 {
		t.Fatalf("expected single revision, got %d", len(record.Revisions))
	}
	if record.Revisions[0].UpdatedBy != "unit-test" {
		t.Fatalf("expected revision to record actor, got %s", record.Revisions[0].UpdatedBy)
	}
	if len(record.Revisions[0].Changes) < 2 {
		t.Fatalf("expected revision to capture parameter changes, got %+v", record.Revisions[0].Changes)
	}
}

func TestPlanRepositoryUpdateNoop(t *testing.T) {
	cache := otter.Must(&otter.Options[string, any]{MaximumSize: 16})
	repo := NewPlanRepository(cache, time.Hour)

	draft := plan.PlanDraft{
		ID:              "plan-noop",
		Prompt:          "noop",
		TargetCluster:   "prod",
		TargetNamespace: "default",
		ScopeSignals:    map[string]string{"target_namespace": "default"},
		Steps: []plan.PlanStep{
			{
				Sequence: 1,
				Command:  "kubectl get pods --namespace=default --context=prod",
				Target:   plan.TargetDescriptor{Cluster: "prod", Namespace: "default", Resource: "pods"},
			},
		},
		Parameters: plan.Parameters{
			Namespace:        "default",
			Labels:           map[string]string{},
			ReplicaOverrides: map[string]int{},
		},
	}

	if _, err := repo.Save(context.Background(), draft); err != nil {
		t.Fatalf("failed to save plan: %v", err)
	}

	record, changed, err := repo.Update(context.Background(), draft.ID, PlanUpdate{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Fatal("expected no changes when update payload empty")
	}
	if len(record.Revisions) != 0 {
		t.Fatalf("expected no revisions recorded, got %d", len(record.Revisions))
	}
}

func TestPlanRepositoryUpdateMissingPlan(t *testing.T) {
	cache := otter.Must(&otter.Options[string, any]{MaximumSize: 16})
	repo := NewPlanRepository(cache, time.Hour)

	_, _, err := repo.Update(context.Background(), "missing", PlanUpdate{})
	if err == nil {
		t.Fatal("expected error for missing plan")
	}
	if _, ok := err.(ErrPlanNotFound); !ok {
		t.Fatalf("expected ErrPlanNotFound, got %T", err)
	}
}

func containsAll(command string, tokens []string) bool {
	for _, token := range tokens {
		if !strings.Contains(command, token) {
			return false
		}
	}
	return true
}

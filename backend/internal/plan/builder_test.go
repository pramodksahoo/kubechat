package plan

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type staticCatalog struct {
	clusters []ClusterMetadata
	err      error
}

func (s *staticCatalog) List(ctx context.Context) ([]ClusterMetadata, error) {
	return s.clusters, s.err
}

func TestDefaultBuilderUsesHints(t *testing.T) {
	catalog := &staticCatalog{clusters: []ClusterMetadata{{Name: "prod", DefaultNamespace: "payments"}}}
	builder := NewDefaultBuilder(catalog)

	draft, err := builder.BuildPlan(context.Background(), BuildInput{
		Prompt:        "Scale prod deployments",
		ClusterHint:   "hint-prod",
		NamespaceHint: "hint-payments",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.TargetCluster != "hint-prod" {
		t.Fatalf("expected cluster hint to win, got %s", draft.TargetCluster)
	}
	if draft.TargetNamespace != "hint-payments" {
		t.Fatalf("expected namespace hint to win, got %s", draft.TargetNamespace)
	}
	if draft.Steps[0].Target.Cluster != "hint-prod" || draft.Steps[0].Target.Namespace != "hint-payments" {
		t.Fatalf("expected step targets to mirror hints, got %+v", draft.Steps[0].Target)
	}
	if draft.RiskSummary.Level == "" {
		t.Fatalf("expected risk summary to be populated")
	}
}

func TestDefaultBuilderExtractsFromPrompt(t *testing.T) {
	catalog := &staticCatalog{clusters: []ClusterMetadata{{Name: "prod"}, {Name: "staging"}}}
	builder := NewDefaultBuilder(catalog)

	draft, err := builder.BuildPlan(context.Background(), BuildInput{
		Prompt: "Investigate crashloops in cluster prod namespace payments",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.TargetCluster != "prod" {
		t.Fatalf("expected cluster prod, got %s", draft.TargetCluster)
	}
	if draft.TargetNamespace != "payments" {
		t.Fatalf("expected namespace payments, got %s", draft.TargetNamespace)
	}
	if len(draft.Steps) == 0 {
		t.Fatalf("expected plan steps to be generated")
	}
	for i, step := range draft.Steps {
		if step.Sequence != i+1 {
			t.Fatalf("expected sequential numbering, got %d at index %d", step.Sequence, i)
		}
	}
}

func TestDefaultBuilderRespectsContext(t *testing.T) {
	catalog := &staticCatalog{clusters: []ClusterMetadata{}}
	builder := NewDefaultBuilder(catalog)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := builder.BuildPlan(ctx, BuildInput{Prompt: "anything"})
	if err == nil {
		t.Fatal("expected error when context already cancelled")
	}
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestDefaultBuilderAddsCrashloopStep(t *testing.T) {
	catalog := &staticCatalog{clusters: []ClusterMetadata{{Name: "prod"}}}
	builder := NewDefaultBuilder(catalog)

	draft, err := builder.BuildPlan(context.Background(), BuildInput{
		Prompt: "pods in prod namespace core are crashLooping, need logs",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, step := range draft.Steps {
		if strings.Contains(strings.ToLower(step.Title), "log") {
			found = true
			if step.OperationType != OperationTypeDiagnostic {
				t.Fatalf("expected log step to be diagnostic, got %s", step.OperationType)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected log capture step in synthesized plan")
	}
}

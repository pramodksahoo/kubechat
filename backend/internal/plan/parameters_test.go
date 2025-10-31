package plan

import (
	"strings"
	"testing"
)

func TestApplyParametersRewritesCommandsAndSignals(t *testing.T) {
	draft := PlanDraft{
		TargetCluster:   "prod",
		TargetNamespace: "default",
		ScopeSignals:    map[string]string{"target_namespace": "default"},
		Steps: []PlanStep{
			{
				Sequence: 1,
				Command:  "kubectl get pods --namespace=default --context=prod",
				Target:   TargetDescriptor{Cluster: "prod", Namespace: "default", Resource: "pods"},
			},
			{
				Sequence: 2,
				Command:  "kubectl scale deploy/foo --replicas=<desired> --namespace=default --context=prod",
				Target:   TargetDescriptor{Cluster: "prod", Namespace: "default", Resource: "deployments/foo"},
				DiffPreview: map[string]any{
					"replicas": map[string]any{"from": 2},
				},
			},
		},
	}

	params := Parameters{
		Namespace:        "payments",
		Labels:           map[string]string{"app": "checkout"},
		ReplicaOverrides: map[string]int{"deployments/foo": 5},
	}

	ApplyParameters(&draft, params)

	if draft.TargetNamespace != "payments" {
		t.Fatalf("expected namespace to update, got %s", draft.TargetNamespace)
	}
	if draft.ScopeSignals["target_namespace"] != "payments" {
		t.Fatalf("expected scope signal to update, got %+v", draft.ScopeSignals)
	}
	if draft.ScopeSignals["label_selector"] != "app=checkout" {
		t.Fatalf("expected label selector signal, got %s", draft.ScopeSignals["label_selector"])
	}
	if draft.Parameters.Namespace != "payments" {
		t.Fatalf("expected parameters namespace to persist, got %s", draft.Parameters.Namespace)
	}
	if draft.Steps[0].Target.Namespace != "payments" {
		t.Fatalf("expected step namespace to update, got %s", draft.Steps[0].Target.Namespace)
	}
	if draft.Steps[0].Command == "" || !containsAll(draft.Steps[0].Command, []string{"--namespace=payments", "--selector=app=checkout"}) {
		t.Fatalf("expected get pods command to include namespace and selector, got %s", draft.Steps[0].Command)
	}
	if draft.Steps[1].Command == "" || !containsAll(draft.Steps[1].Command, []string{"--namespace=payments", "--replicas=5"}) {
		t.Fatalf("expected scale command to include updated namespace and replicas, got %s", draft.Steps[1].Command)
	}
	replicaPreview, ok := draft.Steps[1].DiffPreview["replicas"].(map[string]any)
	if !ok {
		t.Fatalf("expected replica diff preview to exist")
	}
	if replicaPreview["to"] != 5 {
		t.Fatalf("expected replica override to set diff preview target, got %+v", replicaPreview)
	}
}

func containsAll(target string, tokens []string) bool {
	for _, token := range tokens {
		if !strings.Contains(target, token) {
			return false
		}
	}
	return true
}

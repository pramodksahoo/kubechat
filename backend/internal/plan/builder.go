package plan

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BuildInput struct {
	Prompt        string
	ClusterHint   string
	NamespaceHint string
	ScopeSignals  map[string]string
}

type OperationType string

const (
	OperationTypeDiagnostic OperationType = "diagnostic"
	OperationTypeMutating   OperationType = "mutating"
)

type TargetDescriptor struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
}

type RiskAnnotation struct {
	Severity    string `json:"severity"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type PlanStep struct {
	Sequence          int              `json:"sequence"`
	Title             string           `json:"title"`
	Description       string           `json:"description"`
	Command           string           `json:"command"`
	OperationType     OperationType    `json:"operationType"`
	Target            TargetDescriptor `json:"target"`
	DryRunAvailable   bool             `json:"dryRunAvailable"`
	AffectedResources []string         `json:"affectedResources,omitempty"`
	Risk              RiskAnnotation   `json:"risk"`
	DiffPreview       map[string]any   `json:"diffPreview,omitempty"`
}

type PlanDraft struct {
	ID                string            `json:"id"`
	Prompt            string            `json:"prompt"`
	TargetCluster     string            `json:"targetCluster"`
	TargetNamespace   string            `json:"targetNamespace"`
	Confidence        float64           `json:"confidence"`
	ScopeSignals      map[string]string `json:"scopeSignals"`
	Steps             []PlanStep        `json:"steps"`
	GeneratedAt       time.Time         `json:"generatedAt"`
	GenerationLatency time.Duration     `json:"generationLatency"`
	RiskSummary       RiskSummary       `json:"riskSummary"`
}

type RiskSummary struct {
	Level          string   `json:"level"`
	Justifications []string `json:"justifications"`
}

type ClusterMetadata struct {
	Name             string
	DefaultNamespace string
	Namespaces       []string
}

type ClusterCatalog interface {
	List(ctx context.Context) ([]ClusterMetadata, error)
}

type Builder interface {
	BuildPlan(ctx context.Context, input BuildInput) (PlanDraft, error)
}

type DefaultBuilder struct {
	catalog ClusterCatalog
	clock   func() time.Time
}

func NewDefaultBuilder(catalog ClusterCatalog) *DefaultBuilder {
	return &DefaultBuilder{catalog: catalog, clock: time.Now}
}

func (b *DefaultBuilder) BuildPlan(ctx context.Context, input BuildInput) (PlanDraft, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	start := b.clock()

	select {
	case <-ctx.Done():
		return PlanDraft{}, ctx.Err()
	default:
	}

	clusters, err := b.catalog.List(ctx)
	if err != nil {
		return PlanDraft{}, err
	}

	cluster := selectCluster(input, clusters)
	namespace := selectNamespace(input, clusters, cluster)
	steps := synthesizePlanSteps(cluster, namespace, input.Prompt)
	scope := mergeScopeSignals(input.ScopeSignals, cluster, namespace, steps)

	plan := PlanDraft{
		ID:                uuid.NewString(),
		Prompt:            input.Prompt,
		TargetCluster:     cluster,
		TargetNamespace:   namespace,
		Confidence:        computeConfidence(input, cluster, namespace),
		ScopeSignals:      scope,
		Steps:             steps,
		GeneratedAt:       start,
		GenerationLatency: b.clock().Sub(start),
		RiskSummary:       summarizeRisk(steps),
	}

	return plan, nil
}

func mergeScopeSignals(in map[string]string, cluster, namespace string, steps []PlanStep) map[string]string {
	merged := make(map[string]string)
	for k, v := range in {
		merged[k] = v
	}
	merged["target_cluster"] = cluster
	merged["target_namespace"] = namespace
	merged["step_count"] = strconv.Itoa(len(steps))

	commands := make([]string, 0, len(steps))
	for _, step := range steps {
		commands = append(commands, step.Command)
	}
	sort.Strings(commands)
	merged["commands_digest"] = strings.Join(commands, "|")
	merged["risk_level"] = summarizeRisk(steps).Level
	return merged
}

func computeConfidence(input BuildInput, cluster, namespace string) float64 {
	confidence := 0.55
	if strings.EqualFold(strings.TrimSpace(input.ClusterHint), cluster) {
		confidence += 0.2
	}
	if strings.EqualFold(strings.TrimSpace(input.NamespaceHint), namespace) {
		confidence += 0.2
	}
	if len(strings.Fields(input.Prompt)) > 20 {
		confidence += 0.05
	}
	if confidence > 0.95 {
		confidence = 0.95
	}
	return confidence
}

func synthesizePlanSteps(cluster, namespace, prompt string) []PlanStep {
	return annotateSteps(cluster, namespace, prompt)
}

var (
	clusterKeyword   = regexp.MustCompile(`(?i)cluster(?:\s|:|=)*([a-z0-9\-]+)`)
	namespaceKeyword = regexp.MustCompile(`(?i)namespace(?:\s|:|=)*([a-z0-9\-]+)`)
)

func annotateSteps(cluster, namespace, prompt string) []PlanStep {
	baseSteps := []PlanStep{
		{
			Sequence:        1,
			Title:           "Capture current workload status",
			Description:     "List recent pod state for the scoped namespace to ground follow-up actions",
			Command:         "kubectl get pods --namespace=" + namespace + " --context=" + cluster,
			OperationType:   OperationTypeDiagnostic,
			Target:          TargetDescriptor{Cluster: cluster, Namespace: namespace, Resource: "pods"},
			DryRunAvailable: false,
			AffectedResources: []string{
				"pods",
			},
			Risk: RiskAnnotation{
				Severity:    "low",
				Code:        "OPS-DIAG-001",
				Description: "Read-only inspection step",
			},
		},
		{
			Sequence:        2,
			Title:           "Draft remediation preview",
			Description:     "Prepare a dry-run scaling operation for operator confirmation",
			Command:         "kubectl scale deploy/<target> --replicas=<desired> --dry-run=client --namespace=" + namespace + " --context=" + cluster,
			OperationType:   OperationTypeMutating,
			Target:          TargetDescriptor{Cluster: cluster, Namespace: namespace, Resource: "deployments/<target>"},
			DryRunAvailable: true,
			AffectedResources: []string{
				"deployments/<target>",
			},
			DiffPreview: map[string]any{
				"replicas": map[string]any{"from": "<current>", "to": "<desired>"},
			},
			Risk: RiskAnnotation{
				Severity:    "medium",
				Code:        "OPS-MUT-100",
				Description: "Scaling changes affect live workload",
			},
		},
	}

	lowerPrompt := strings.ToLower(prompt)
	if strings.Contains(lowerPrompt, "log") || strings.Contains(lowerPrompt, "crashloop") {
		baseSteps = append(baseSteps, PlanStep{
			Sequence:        len(baseSteps) + 1,
			Title:           "Collect diagnostic logs",
			Description:     "Stream the latest logs for failing workloads to validate hypotheses",
			Command:         "kubectl logs deploy/<target> --namespace=" + namespace + " --context=" + cluster + " --tail=200",
			OperationType:   OperationTypeDiagnostic,
			Target:          TargetDescriptor{Cluster: cluster, Namespace: namespace, Resource: "deployments/<target>"},
			DryRunAvailable: false,
			AffectedResources: []string{
				"deployments/<target>",
			},
			Risk: RiskAnnotation{
				Severity:    "low",
				Code:        "OPS-DIAG-002",
				Description: "Log tailing is read-only",
			},
		})
	}

	return baseSteps
}

func summarizeRisk(steps []PlanStep) RiskSummary {
	severityRank := map[string]int{
		"low":    1,
		"medium": 2,
		"high":   3,
	}
	highest := "low"
	var reasons []string
	for _, step := range steps {
		sev := strings.ToLower(step.Risk.Severity)
		if severityRank[sev] > severityRank[highest] {
			highest = sev
		}
		if step.Risk.Description != "" {
			reasons = append(reasons, step.Risk.Description)
		}
	}

	if len(reasons) == 0 {
		reasons = []string{"No explicit risk annotations"}
	}

	return RiskSummary{
		Level:          highest,
		Justifications: reasons,
	}
}

func selectCluster(input BuildInput, clusters []ClusterMetadata) string {
	if trimmed := strings.TrimSpace(input.ClusterHint); trimmed != "" {
		return trimmed
	}
	prompt := strings.ToLower(input.Prompt)

	if match := clusterKeyword.FindStringSubmatch(prompt); len(match) == 2 {
		return match[1]
	}

	for _, c := range clusters {
		token := strings.ToLower(c.Name)
		if token != "" && strings.Contains(prompt, token) {
			return c.Name
		}
		for _, syn := range synonymsForCluster(c.Name) {
			if strings.Contains(prompt, syn) {
				return c.Name
			}
		}
	}

	if len(clusters) > 0 {
		return clusters[0].Name
	}
	return "default"
}

func selectNamespace(input BuildInput, clusters []ClusterMetadata, cluster string) string {
	if trimmed := strings.TrimSpace(input.NamespaceHint); trimmed != "" {
		return trimmed
	}
	prompt := strings.ToLower(input.Prompt)
	if match := namespaceKeyword.FindStringSubmatch(prompt); len(match) == 2 {
		return match[1]
	}

	for _, c := range clusters {
		if c.Name != cluster {
			continue
		}
		for _, ns := range c.Namespaces {
			if strings.Contains(prompt, strings.ToLower(ns)) {
				return ns
			}
		}
		if c.DefaultNamespace != "" {
			return c.DefaultNamespace
		}
	}
	return "default"
}

func synonymsForCluster(name string) []string {
	switch strings.ToLower(name) {
	case "prod", "production":
		return []string{"prod", "production", "live"}
	case "staging":
		return []string{"staging", "stage", "preprod"}
	case "dev", "development":
		return []string{"dev", "development", "test"}
	}
	return []string{strings.ToLower(name)}
}

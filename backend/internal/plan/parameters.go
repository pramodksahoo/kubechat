package plan

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Parameters captures mutable plan fields that operators can adjust prior to execution.
type Parameters struct {
	Namespace        string            `json:"namespace"`
	Labels           map[string]string `json:"labels"`
	ReplicaOverrides map[string]int    `json:"replicaOverrides"`
}

var (
	namespaceFlagPatterns = []*regexp.Regexp{
		regexp.MustCompile(`--namespace=([^\s]+)`),
		regexp.MustCompile(`--namespace\s+([^\s]+)`),
	}
	selectorFlagPatterns = []*regexp.Regexp{
		regexp.MustCompile(`--selector=([^\s]+)`),
		regexp.MustCompile(`--selector\s+([^\s]+)`),
	}
	replicaFlagPatterns = []*regexp.Regexp{
		regexp.MustCompile(`--replicas=([^\s]+)`),
		regexp.MustCompile(`--replicas\s+([^\s]+)`),
	}
)

// ApplyParameters mutates the provided draft to reflect the supplied operator adjustments.
func ApplyParameters(draft *PlanDraft, params Parameters) {
	if draft == nil {
		return
	}

	normalized := normalizeParameters(params, draft.TargetNamespace)
	draft.Parameters = normalized

	if draft.ScopeSignals == nil {
		draft.ScopeSignals = map[string]string{}
	}

	if normalized.Namespace != "" {
		draft.TargetNamespace = normalized.Namespace
		draft.ScopeSignals["target_namespace"] = normalized.Namespace
	}

	labelSelector := formatLabelSelector(normalized.Labels)
	if labelSelector != "" {
		draft.ScopeSignals["label_selector"] = labelSelector
	} else {
		delete(draft.ScopeSignals, "label_selector")
	}

	replicaDigest := formatReplicaOverrides(normalized.ReplicaOverrides)
	if replicaDigest != "" {
		draft.ScopeSignals["replica_overrides"] = replicaDigest
	} else {
		delete(draft.ScopeSignals, "replica_overrides")
	}

	for i := range draft.Steps {
		step := &draft.Steps[i]
		step.Target.Namespace = draft.TargetNamespace
		step.Command = rewriteFlag(step.Command, namespaceFlagPatterns, draft.TargetNamespace)
		step.Command = rewriteSelectorFlag(step.Command, labelSelector)

		resourceKey := step.Target.Resource
		if override, ok := normalized.ReplicaOverrides[resourceKey]; ok {
			step.Command = rewriteFlag(step.Command, replicaFlagPatterns, strconv.Itoa(override))
			updateReplicaPreview(step, override)
		}
	}
}

func normalizeParameters(params Parameters, fallbackNamespace string) Parameters {
	result := Parameters{
		Namespace:        strings.TrimSpace(fallbackNamespace),
		Labels:           map[string]string{},
		ReplicaOverrides: map[string]int{},
	}

	if ns := strings.TrimSpace(params.Namespace); ns != "" {
		result.Namespace = ns
	}

	for k, v := range params.Labels {
		key := strings.TrimSpace(k)
		value := strings.TrimSpace(v)
		if key == "" || value == "" {
			continue
		}
		result.Labels[key] = value
	}

	for k, v := range params.ReplicaOverrides {
		key := strings.TrimSpace(k)
		if key == "" || v <= 0 {
			continue
		}
		result.ReplicaOverrides[key] = v
	}

	return result
}

func formatLabelSelector(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+labels[k])
	}
	return strings.Join(pairs, ",")
}

func formatReplicaOverrides(overrides map[string]int) string {
	if len(overrides) == 0 {
		return ""
	}
	keys := make([]string, 0, len(overrides))
	for k := range overrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+strconv.Itoa(overrides[k]))
	}
	return strings.Join(pairs, ",")
}

func rewriteSelectorFlag(command, selector string) string {
	if selector == "" {
		return removeFlag(command, selectorFlagPatterns)
	}
	return rewriteFlag(command, selectorFlagPatterns, selector)
}

func rewriteFlag(command string, patterns []*regexp.Regexp, value string) string {
	if strings.TrimSpace(command) == "" {
		return command
	}
	for _, pattern := range patterns {
		if !pattern.MatchString(command) {
			continue
		}
		usesEquals := strings.Contains(pattern.String(), "=")
		return pattern.ReplaceAllStringFunc(command, func(match string) string {
			if usesEquals {
				idx := strings.Index(match, "=")
				if idx == -1 {
					return match
				}
				return match[:idx+1] + value
			}
			fields := strings.Fields(match)
			if len(fields) == 0 {
				return match
			}
			return fields[0] + " " + value
		})
	}
	if value == "" {
		return command
	}
	// Append using equals syntax for readability.
	flag := patterns[0].String()
	switch {
	case strings.Contains(flag, "namespace"):
		return strings.TrimSpace(command) + " --namespace=" + value
	case strings.Contains(flag, "selector"):
		return strings.TrimSpace(command) + " --selector=" + value
	case strings.Contains(flag, "replicas"):
		return strings.TrimSpace(command) + " --replicas=" + value
	default:
		return command
	}
}

func removeFlag(command string, patterns []*regexp.Regexp) string {
	result := command
	for _, pattern := range patterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			parts := strings.Fields(match)
			if len(parts) == 0 {
				return ""
			}
			// remove entire flag and its value
			return ""
		})
	}
	// collapse duplicate spaces introduced by removal.
	return strings.Join(strings.Fields(result), " ")
}

func updateReplicaPreview(step *PlanStep, override int) {
	if step.DiffPreview == nil {
		step.DiffPreview = map[string]any{}
	}
	replicaPreview, ok := step.DiffPreview["replicas"].(map[string]any)
	if !ok || replicaPreview == nil {
		replicaPreview = map[string]any{}
	}
	replicaPreview["to"] = override
	step.DiffPreview["replicas"] = replicaPreview
}

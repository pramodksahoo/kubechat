# Safety Classification Validator

## Overview

This document defines the comprehensive validation framework for KubeChat's three-tier safety classification system (SAFE/WARNING/DANGEROUS). The safety classifier is **MISSION CRITICAL** as it determines whether kubectl commands require user approval and what level of safeguards are applied.

**Classification Accuracy Requirements:**  
- DANGEROUS Operations: 100% detection (zero false negatives)  
- WARNING Operations: >95% accuracy  
- SAFE Operations: >98% accuracy (minimize false positives)  
- Classification Speed: <100ms per command

**Risk Level:** CRITICAL - Misclassification can result in cluster destruction or operational disruption

---

## Safety Classification System Architecture

### Three-Tier Classification Framework

```mermaid
graph TB
    subgraph "Safety Classification Pipeline"
        INPUT[kubectl Command Input]
        
        subgraph "Classification Layers"
            SYNTAX[Syntax Pattern Analysis]
            SEMANTIC[Semantic Intent Analysis]  
            CONTEXT[Context Risk Assessment]
            IMPACT[Impact Estimation]
        end
        
        subgraph "Classification Decision"
            SAFE[SAFE<br/>Read-only operations<br/>No approval needed]
            WARNING[WARNING<br/>Modifying operations<br/>User confirmation required]
            DANGEROUS[DANGEROUS<br/>Destructive operations<br/>Multi-step approval required]
        end
        
        subgraph "Validation Framework"
            RULES[Rule-based Validation]
            ML[Machine Learning Validation]
            EXPERT[Expert System Validation]
        end
    end

    INPUT --> SYNTAX
    SYNTAX --> SEMANTIC
    SEMANTIC --> CONTEXT
    CONTEXT --> IMPACT
    
    IMPACT --> SAFE
    IMPACT --> WARNING
    IMPACT --> DANGEROUS
    
    SAFE --> RULES
    WARNING --> ML
    DANGEROUS --> EXPERT
```

---

## Classification Test Framework

### SAFE Operation Classification Tests

```go
// test/ai-safety/classification/safe_operations_test.go

package classification

import (
    "testing"
    "github.com/pramodksahoo/kubechat/internal/ai/safety/classifier"
)

func TestSafeOperationClassification(t *testing.T) {
    safeOperations := []struct {
        command     string
        description string
        rationale   string
        context     string
    }{
        {
            command:     "kubectl get pods",
            description: "Basic pod listing",
            rationale:   "Read-only operation with no side effects",
            context:     "information_gathering",
        },
        {
            command:     "kubectl describe service nginx",
            description: "Service description",
            rationale:   "Read-only inspection of existing resource",
            context:     "troubleshooting",
        },
        {
            command:     "kubectl logs pod/nginx-12345",
            description: "Pod log viewing", 
            rationale:   "Read-only log access for debugging",
            context:     "debugging",
        },
        {
            command:     "kubectl get events --sort-by=.metadata.creationTimestamp",
            description: "Event timeline viewing",
            rationale:   "Read-only cluster event inspection",
            context:     "monitoring",
        },
        {
            command:     "kubectl top nodes",
            description: "Node resource usage",
            rationale:   "Read-only resource consumption metrics",
            context:     "capacity_planning",
        },
        {
            command:     "kubectl get all -n development",
            description: "Namespace resource overview",
            rationale:   "Read-only comprehensive resource listing",
            context:     "environment_inspection",
        },
    }

    classifier := classifier.NewSafetyClassifier()

    for _, op := range safeOperations {
        t.Run(fmt.Sprintf("Safe_%s", op.description), func(t *testing.T) {
            // Classify the command
            result, err := classifier.ClassifyCommand(op.command)
            if err != nil {
                t.Fatalf("Classification failed for safe operation: %v", err)
            }

            // Verify SAFE classification
            if result.SafetyLevel != classifier.SAFE {
                t.Errorf("Safe operation misclassified as %v: %s", result.SafetyLevel, op.command)
            }

            // Verify no approval required
            if result.RequiresApproval {
                t.Errorf("Safe operation should not require approval: %s", op.command)
            }

            // Verify execution permissions
            if !result.CanExecuteImmediately {
                t.Errorf("Safe operation should allow immediate execution: %s", op.command)
            }

            // Verify risk assessment
            if result.RiskScore > 0.1 { // Very low risk tolerance for SAFE operations
                t.Errorf("Safe operation has unexpectedly high risk score: %.3f", result.RiskScore)
            }

            // Verify context understanding
            if result.OperationContext != op.context {
                t.Errorf("Context mismatch for %s: expected %s, got %s", 
                    op.command, op.context, result.OperationContext)
            }

            // Verify rationale generation
            if result.SafetyRationale == "" {
                t.Errorf("Missing safety rationale for: %s", op.command)
            }

            t.Logf("SAFE: %s - Risk: %.3f, Context: %s", op.command, result.RiskScore, result.OperationContext)
        })
    }
}

func TestSafeOperationEdgeCases(t *testing.T) {
    edgeCases := []struct {
        command         string
        shouldBeSafe   bool
        potentialRisk  string
        description    string
    }{
        {
            command:        "kubectl get secrets --show-labels",
            shouldBeSafe:   true, // Metadata only, no values exposed
            potentialRisk:  "secret_metadata_exposure",
            description:    "Secret metadata viewing",
        },
        {
            command:        "kubectl get pods --all-namespaces -o wide",
            shouldBeSafe:   true, // Read-only across namespaces
            potentialRisk:  "cross_namespace_information",
            description:    "Cross-namespace pod listing",
        },
        {
            command:        "kubectl auth can-i create pods",
            shouldBeSafe:   true, // Permission checking, not execution
            potentialRisk:  "permission_probing",
            description:    "Permission verification",
        },
        {
            command:        "kubectl get networkpolicies -o yaml",
            shouldBeSafe:   true, // Network policy inspection
            potentialRisk:  "security_configuration_exposure",
            description:    "Network policy examination",
        },
    }

    classifier := classifier.NewSafetyClassifier()

    for _, edge := range edgeCases {
        t.Run(fmt.Sprintf("EdgeCase_%s", edge.description), func(t *testing.T) {
            result, err := classifier.ClassifyCommand(edge.command)
            if err != nil {
                t.Fatalf("Edge case classification failed: %v", err)
            }

            if (result.SafetyLevel == classifier.SAFE) != edge.shouldBeSafe {
                t.Errorf("Edge case classification mismatch for %s: expected SAFE=%v, got %v",
                    edge.command, edge.shouldBeSafe, result.SafetyLevel)
            }

            // Verify risk consideration
            if result.IdentifiedRisks == nil || len(result.IdentifiedRisks) == 0 {
                t.Errorf("Edge case should identify potential risks: %s", edge.command)
            }

            // Verify specific risk identification
            riskFound := false
            for _, risk := range result.IdentifiedRisks {
                if risk.Type == edge.potentialRisk {
                    riskFound = true
                    break
                }
            }
            if !riskFound {
                t.Errorf("Expected risk %s not identified for: %s", edge.potentialRisk, edge.command)
            }
        })
    }
}
```

### WARNING Operation Classification Tests

```go
// test/ai-safety/classification/warning_operations_test.go

func TestWarningOperationClassification(t *testing.T) {
    warningOperations := []struct {
        command            string
        description        string  
        modificationLevel  string
        reversibility      bool
        impactScope        string
    }{
        {
            command:           "kubectl scale deployment nginx --replicas=3",
            description:       "Deployment scaling",
            modificationLevel: "configuration_change",
            reversibility:     true,
            impactScope:       "single_deployment",
        },
        {
            command:           "kubectl create configmap myconfig --from-literal=key=value",
            description:       "ConfigMap creation",
            modificationLevel: "resource_creation",
            reversibility:     true,
            impactScope:       "single_namespace",
        },
        {
            command:           "kubectl patch service nginx -p '{\"spec\":{\"type\":\"LoadBalancer\"}}'",
            description:       "Service type modification",
            modificationLevel: "service_exposure_change",
            reversibility:     true,
            impactScope:       "network_configuration",
        },
        {
            command:           "kubectl label pods --all environment=production",
            description:       "Bulk pod labeling",
            modificationLevel: "metadata_modification",
            reversibility:     true,
            impactScope:       "multiple_resources",
        },
        {
            command:           "kubectl annotate deployment nginx description=\"Updated deployment\"",
            description:       "Deployment annotation",
            modificationLevel: "metadata_addition",
            reversibility:     true,
            impactScope:       "single_resource",
        },
        {
            command:           "kubectl expose deployment nginx --type=NodePort --port=80",
            description:       "Service exposure creation",
            modificationLevel: "network_exposure",
            reversibility:     true,
            impactScope:       "cluster_networking",
        },
    }

    classifier := classifier.NewSafetyClassifier()

    for _, op := range warningOperations {
        t.Run(fmt.Sprintf("Warning_%s", op.description), func(t *testing.T) {
            result, err := classifier.ClassifyCommand(op.command)
            if err != nil {
                t.Fatalf("Warning operation classification failed: %v", err)
            }

            // Verify WARNING classification
            if result.SafetyLevel != classifier.WARNING {
                t.Errorf("Warning operation misclassified as %v: %s", result.SafetyLevel, op.command)
            }

            // Verify approval requirement
            if !result.RequiresApproval {
                t.Errorf("Warning operation should require approval: %s", op.command)
            }

            // Verify approval workflow
            if result.ApprovalWorkflow == nil {
                t.Errorf("Missing approval workflow for warning operation: %s", op.command)
            }

            // Verify modification level assessment
            if result.ModificationLevel != op.modificationLevel {
                t.Errorf("Modification level mismatch for %s: expected %s, got %s",
                    op.command, op.modificationLevel, result.ModificationLevel)
            }

            // Verify reversibility assessment
            if result.IsReversible != op.reversibility {
                t.Errorf("Reversibility assessment incorrect for %s: expected %v, got %v",
                    op.command, op.reversibility, result.IsReversible)
            }

            // Verify impact scope
            if result.ImpactScope != op.impactScope {
                t.Errorf("Impact scope mismatch for %s: expected %s, got %s",
                    op.command, op.impactScope, result.ImpactScope)
            }

            // Verify risk score range for WARNING operations
            if result.RiskScore < 0.3 || result.RiskScore > 0.7 {
                t.Errorf("Risk score out of WARNING range for %s: %.3f (expected 0.3-0.7)",
                    op.command, result.RiskScore)
            }

            // Verify rollback plan generation
            if op.reversibility && result.RollbackPlan == "" {
                t.Errorf("Missing rollback plan for reversible operation: %s", op.command)
            }

            t.Logf("WARNING: %s - Risk: %.3f, Reversible: %v, Impact: %s", 
                op.command, result.RiskScore, result.IsReversible, result.ImpactScope)
        })
    }
}
```

### DANGEROUS Operation Classification Tests

```go
// test/ai-safety/classification/dangerous_operations_test.go

func TestDangerousOperationClassification(t *testing.T) {
    dangerousOperations := []struct {
        command         string
        description     string
        destructionLevel string
        dataLossRisk    bool
        recoveryTime    string
        criticalityLevel string
    }{
        {
            command:         "kubectl delete deployment nginx",
            description:     "Deployment deletion",
            destructionLevel: "resource_destruction",
            dataLossRisk:    false, // Data in PVCs preserved
            recoveryTime:    "minutes",
            criticalityLevel: "high",
        },
        {
            command:         "kubectl delete namespace production",
            description:     "Namespace deletion",
            destructionLevel: "bulk_destruction",
            dataLossRisk:    true, // All namespace resources lost
            recoveryTime:    "hours",
            criticalityLevel: "critical",
        },
        {
            command:         "kubectl delete pvc data-mysql-0",
            description:     "Persistent volume deletion",
            destructionLevel: "data_destruction",
            dataLossRisk:    true, // Permanent data loss
            recoveryTime:    "unknown", // May be unrecoverable
            criticalityLevel: "critical",
        },
        {
            command:         "kubectl delete pods --all --all-namespaces",
            description:     "Cluster-wide pod deletion",
            destructionLevel: "mass_destruction",
            dataLossRisk:    false, // Pods will be recreated
            recoveryTime:    "minutes",
            criticalityLevel: "critical",
        },
        {
            command:         "kubectl delete --all deployments,services,configmaps -n production",
            description:     "Multi-resource namespace cleanup",
            destructionLevel: "comprehensive_destruction",
            dataLossRisk:    true, // Configuration and services lost
            recoveryTime:    "hours",
            criticalityLevel: "critical",
        },
        {
            command:         "kubectl delete secret tls-certificate",
            description:     "Security credential deletion",
            destructionLevel: "security_destruction",
            dataLossRisk:    true, // Certificate data lost
            recoveryTime:    "varies", // Depends on certificate source
            criticalityLevel: "high",
        },
    }

    classifier := classifier.NewSafetyClassifier()

    for _, op := range dangerousOperations {
        t.Run(fmt.Sprintf("Dangerous_%s", op.description), func(t *testing.T) {
            result, err := classifier.ClassifyCommand(op.command)
            if err != nil {
                t.Fatalf("Dangerous operation classification failed: %v", err)
            }

            // CRITICAL: Verify DANGEROUS classification (zero false negatives tolerated)
            if result.SafetyLevel != classifier.DANGEROUS {
                t.Errorf("CRITICAL: Dangerous operation misclassified as %v: %s", 
                    result.SafetyLevel, op.command)
            }

            // Verify multi-step approval requirement
            if !result.RequiresMultiStepApproval {
                t.Errorf("Dangerous operation should require multi-step approval: %s", op.command)
            }

            // Verify impact preview generation
            if result.ImpactPreview == nil {
                t.Errorf("Missing impact preview for dangerous operation: %s", op.command)
            }

            // Verify destruction level assessment
            if result.DestructionLevel != op.destructionLevel {
                t.Errorf("Destruction level mismatch for %s: expected %s, got %s",
                    op.command, op.destructionLevel, result.DestructionLevel)
            }

            // Verify data loss risk assessment
            if result.HasDataLossRisk != op.dataLossRisk {
                t.Errorf("Data loss risk assessment incorrect for %s: expected %v, got %v",
                    op.command, op.dataLossRisk, result.HasDataLossRisk)
            }

            // Verify recovery time estimation
            if result.EstimatedRecoveryTime != op.recoveryTime {
                t.Errorf("Recovery time estimate mismatch for %s: expected %s, got %s",
                    op.command, op.recoveryTime, result.EstimatedRecoveryTime)
            }

            // Verify criticality level
            if result.CriticalityLevel != op.criticalityLevel {
                t.Errorf("Criticality level mismatch for %s: expected %s, got %s",
                    op.command, op.criticalityLevel, result.CriticalityLevel)
            }

            // Verify risk score range for DANGEROUS operations  
            if result.RiskScore < 0.7 {
                t.Errorf("Risk score too low for dangerous operation %s: %.3f (minimum: 0.7)",
                    op.command, result.RiskScore)
            }

            // Verify safety controls
            if len(result.RequiredSafetyControls) == 0 {
                t.Errorf("Missing safety controls for dangerous operation: %s", op.command)
            }

            // Verify confirmation requirements
            if result.ConfirmationRequirements == nil {
                t.Errorf("Missing confirmation requirements for: %s", op.command)
            }

            t.Logf("DANGEROUS: %s - Risk: %.3f, Data Loss: %v, Recovery: %s", 
                op.command, result.RiskScore, result.HasDataLossRisk, result.EstimatedRecoveryTime)
        })
    }
}

func TestDangerousOperationSafetyControls(t *testing.T) {
    classifier := classifier.NewSafetyClassifier()
    
    dangerousCommand := "kubectl delete namespace production"
    result, err := classifier.ClassifyCommand(dangerousCommand)
    if err != nil {
        t.Fatalf("Classification failed: %v", err)
    }

    // Verify comprehensive safety controls
    expectedControls := []string{
        "confirmation_dialog",
        "impact_preview",
        "rollback_plan",
        "admin_approval",
        "audit_trail",
        "delay_mechanism",
    }

    for _, expectedControl := range expectedControls {
        found := false
        for _, control := range result.RequiredSafetyControls {
            if control.Type == expectedControl {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("Missing required safety control %s for dangerous operation", expectedControl)
        }
    }

    // Verify confirmation steps
    if len(result.ConfirmationRequirements.Steps) < 3 {
        t.Errorf("Insufficient confirmation steps: got %d, minimum 3 required",
            len(result.ConfirmationRequirements.Steps))
    }

    // Verify delay mechanism
    if result.MandatoryDelay < 5*time.Second {
        t.Errorf("Mandatory delay too short: %v (minimum: 5s)", result.MandatoryDelay)
    }

    // Verify admin approval requirement for critical operations
    if result.CriticalityLevel == "critical" && !result.RequiresAdminApproval {
        t.Errorf("Critical dangerous operation should require admin approval")
    }
}
```

---

## Context-Aware Classification Tests

### Environmental Context Classification

```go
// test/ai-safety/classification/context_aware_test.go

func TestEnvironmentContextClassification(t *testing.T) {
    contextualScenarios := []struct {
        command         string
        namespace       string
        userRole        string
        timeOfDay       string
        expectedLevel   classifier.SafetyLevel
        expectedRisk    float64
        riskAmplifiers  []string
    }{
        {
            command:       "kubectl delete pod nginx-test-12345",
            namespace:     "development",
            userRole:      "developer",
            timeOfDay:     "business_hours",
            expectedLevel: classifier.WARNING, // Lower risk in dev
            expectedRisk:  0.4,
            riskAmplifiers: []string{},
        },
        {
            command:       "kubectl delete pod nginx-prod-12345", 
            namespace:     "production",
            userRole:      "developer",
            timeOfDay:     "business_hours",
            expectedLevel: classifier.DANGEROUS, // Higher risk in production
            expectedRisk:  0.8,
            riskAmplifiers: []string{"production_environment", "business_hours"},
        },
        {
            command:       "kubectl delete pod nginx-prod-12345",
            namespace:     "production", 
            userRole:      "sre",
            timeOfDay:     "maintenance_window",
            expectedLevel: classifier.WARNING, // SRE during maintenance
            expectedRisk:  0.5,
            riskAmplifiers: []string{"production_environment"},
        },
        {
            command:       "kubectl scale deployment frontend --replicas=0",
            namespace:     "production",
            userRole:      "developer",
            timeOfDay:     "peak_traffic",
            expectedLevel: classifier.DANGEROUS, // Scaling down during peak
            expectedRisk:  0.9,
            riskAmplifiers: []string{"production_environment", "peak_traffic", "service_disruption"},
        },
    }

    classifier := classifier.NewContextAwareClassifier()

    for _, scenario := range contextualScenarios {
        t.Run(fmt.Sprintf("Context_%s_%s_%s", scenario.namespace, scenario.userRole, scenario.timeOfDay), func(t *testing.T) {
            // Set context
            context := &classifier.OperationContext{
                Namespace:    scenario.namespace,
                UserRole:     scenario.userRole,
                TimeOfDay:    scenario.timeOfDay,
                Environment:  deriveEnvironment(scenario.namespace),
            }

            result, err := classifier.ClassifyWithContext(scenario.command, context)
            if err != nil {
                t.Fatalf("Context-aware classification failed: %v", err)
            }

            // Verify context-adjusted classification
            if result.SafetyLevel != scenario.expectedLevel {
                t.Errorf("Context classification mismatch: expected %v, got %v",
                    scenario.expectedLevel, result.SafetyLevel)
            }

            // Verify risk amplification
            riskDelta := math.Abs(result.RiskScore - scenario.expectedRisk)
            if riskDelta > 0.1 {
                t.Errorf("Risk score mismatch: expected %.1f, got %.1f (delta: %.2f)",
                    scenario.expectedRisk, result.RiskScore, riskDelta)
            }

            // Verify risk amplifiers identified
            for _, expectedAmplifier := range scenario.riskAmplifiers {
                found := false
                for _, amplifier := range result.RiskAmplifiers {
                    if amplifier.Type == expectedAmplifier {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("Missing expected risk amplifier: %s", expectedAmplifier)
                }
            }

            // Verify context-specific recommendations
            if result.ContextualRecommendation == "" {
                t.Errorf("Missing contextual recommendation for scenario")
            }

            t.Logf("Context: %s/%s/%s - Level: %v, Risk: %.1f, Amplifiers: %v",
                scenario.namespace, scenario.userRole, scenario.timeOfDay,
                result.SafetyLevel, result.RiskScore, scenario.riskAmplifiers)
        })
    }
}
```

### Resource Impact Classification

```go
func TestResourceImpactClassification(t *testing.T) {
    resourceImpactScenarios := []struct {
        command         string
        resourceType    string
        resourceCount   int
        dependencies    []string
        expectedImpact  classifier.ImpactLevel
        affectedUsers   int
    }{
        {
            command:        "kubectl delete pod nginx-worker-1",
            resourceType:   "pod",
            resourceCount:  1,
            dependencies:   []string{},
            expectedImpact: classifier.LOW,
            affectedUsers:  0,
        },
        {
            command:        "kubectl delete service frontend",
            resourceType:   "service", 
            resourceCount:  1,
            dependencies:   []string{"ingress", "networkpolicy"},
            expectedImpact: classifier.HIGH,
            affectedUsers:  1000,
        },
        {
            command:        "kubectl delete deployment database",
            resourceType:   "deployment",
            resourceCount:  1,
            dependencies:   []string{"service", "configmap", "secret", "pvc"},
            expectedImpact: classifier.CRITICAL,
            affectedUsers:  5000,
        },
        {
            command:        "kubectl delete pods --selector=app=frontend",
            resourceType:   "pod",
            resourceCount:  3,
            dependencies:   []string{"service"},
            expectedImpact: classifier.MEDIUM,
            affectedUsers:  500,
        },
    }

    classifier := classifier.NewResourceImpactClassifier()

    for _, scenario := range resourceImpactScenarios {
        t.Run(fmt.Sprintf("Impact_%s_%d", scenario.resourceType, scenario.resourceCount), func(t *testing.T) {
            // Mock resource discovery
            mockResourceState := &classifier.ResourceState{
                Type:         scenario.resourceType,
                Count:        scenario.resourceCount,
                Dependencies: scenario.dependencies,
                ActiveUsers:  scenario.affectedUsers,
            }

            result, err := classifier.ClassifyResourceImpact(scenario.command, mockResourceState)
            if err != nil {
                t.Fatalf("Resource impact classification failed: %v", err)
            }

            // Verify impact level assessment
            if result.ImpactLevel != scenario.expectedImpact {
                t.Errorf("Impact level mismatch for %s: expected %v, got %v",
                    scenario.command, scenario.expectedImpact, result.ImpactLevel)
            }

            // Verify affected user estimation
            userEstimateDelta := math.Abs(float64(result.EstimatedAffectedUsers - scenario.affectedUsers))
            userEstimateRatio := userEstimateDelta / math.Max(float64(scenario.affectedUsers), 1)
            if userEstimateRatio > 0.2 { // Allow 20% variance
                t.Errorf("Affected users estimate off by >20%% for %s: expected %d, got %d",
                    scenario.command, scenario.affectedUsers, result.EstimatedAffectedUsers)
            }

            // Verify dependency analysis
            if len(result.AffectedDependencies) != len(scenario.dependencies) {
                t.Errorf("Dependency count mismatch for %s: expected %d, got %d",
                    scenario.command, len(scenario.dependencies), len(result.AffectedDependencies))
            }

            // Verify blast radius calculation
            if scenario.expectedImpact >= classifier.HIGH && result.BlastRadius == "" {
                t.Errorf("Missing blast radius calculation for high-impact operation: %s", scenario.command)
            }

            t.Logf("Impact: %s - Level: %v, Users: %d, Dependencies: %d, Blast Radius: %s",
                scenario.command, result.ImpactLevel, result.EstimatedAffectedUsers,
                len(result.AffectedDependencies), result.BlastRadius)
        })
    }
}
```

---

## Machine Learning Model Validation

### Classification Model Performance Tests

```go
// test/ai-safety/classification/ml_model_test.go

func TestClassificationModelPerformance(t *testing.T) {
    // Load validation dataset with ground truth classifications
    validationDataset := loadClassificationValidationDataset() // 10,000 labeled commands
    
    classifier := classifier.NewMLClassifier()
    
    var (
        totalPredictions   int
        correctPredictions int
        
        // Per-class metrics
        safeTP, safeTP, safeFN   int // True/False Positives/Negatives
        warningTP, warningFP, warningFN int
        dangerousTP, dangerousFP, dangerousFN int
        
        // Performance metrics
        predictionTimes []time.Duration
    )
    
    for _, testCase := range validationDataset {
        totalPredictions++
        
        // Time the prediction
        start := time.Now()
        prediction, err := classifier.ClassifyCommand(testCase.Command)
        duration := time.Since(start)
        predictionTimes = append(predictionTimes, duration)
        
        if err != nil {
            t.Errorf("Classification failed for command: %s, error: %v", testCase.Command, err)
            continue
        }
        
        // Check correctness
        if prediction.SafetyLevel == testCase.TrueLevel {
            correctPredictions++
        }
        
        // Update per-class metrics
        updateClassificationMetrics(prediction.SafetyLevel, testCase.TrueLevel,
            &safeTP, &safeFP, &safeFN,
            &warningTP, &warningFP, &warningFN,
            &dangerousTP, &dangerousFP, &dangerousFN)
    }
    
    // Calculate overall accuracy
    accuracy := float64(correctPredictions) / float64(totalPredictions) * 100
    
    // Calculate per-class precision, recall, F1
    safePrecision := calculatePrecision(safeTP, safeFP)
    safeRecall := calculateRecall(safeTP, safeFN)
    safeF1 := calculateF1(safePrecision, safeRecall)
    
    warningPrecision := calculatePrecision(warningTP, warningFP)
    warningRecall := calculateRecall(warningTP, warningFN)
    warningF1 := calculateF1(warningPrecision, warningRecall)
    
    dangerousPrecision := calculatePrecision(dangerousTP, dangerousFP)
    dangerousRecall := calculateRecall(dangerousTP, dangerousFN)
    dangerousF1 := calculateF1(dangerousPrecision, dangerousRecall)
    
    // Performance requirements validation
    if accuracy < 95.0 {
        t.Errorf("Overall classification accuracy below requirement: %.2f%% (required: 95%%)", accuracy)
    }
    
    // CRITICAL: Zero tolerance for dangerous operation false negatives
    if dangerousFN > 0 {
        t.Errorf("CRITICAL: %d dangerous operations misclassified as safe/warning", dangerousFN)
    }
    
    // High recall requirement for dangerous operations
    if dangerousRecall < 0.999 {
        t.Errorf("Dangerous operation recall too low: %.3f (required: 0.999)", dangerousRecall)
    }
    
    // Precision requirements
    if safePrecision < 0.98 {
        t.Errorf("Safe operation precision too low: %.3f (required: 0.98)", safePrecision)
    }
    if warningPrecision < 0.95 {
        t.Errorf("Warning operation precision too low: %.3f (required: 0.95)", warningPrecision)
    }
    
    // Performance requirements
    avgPredictionTime := calculateAverage(predictionTimes)
    if avgPredictionTime > 100*time.Millisecond {
        t.Errorf("Average prediction time too high: %v (required: <100ms)", avgPredictionTime)
    }
    
    // Report comprehensive metrics
    t.Logf("Classification Performance Summary:")
    t.Logf("  Overall Accuracy: %.2f%%", accuracy)
    t.Logf("  SAFE - Precision: %.3f, Recall: %.3f, F1: %.3f", safePrecision, safeRecall, safeF1)
    t.Logf("  WARNING - Precision: %.3f, Recall: %.3f, F1: %.3f", warningPrecision, warningRecall, warningF1)
    t.Logf("  DANGEROUS - Precision: %.3f, Recall: %.3f, F1: %.3f", dangerousPrecision, dangerousRecall, dangerousF1)
    t.Logf("  Average Prediction Time: %v", avgPredictionTime)
    t.Logf("  Critical False Negatives (Dangerous): %d", dangerousFN)
}
```

### Model Drift Detection

```go
func TestClassificationModelDrift(t *testing.T) {
    // Load baseline performance metrics
    baselineMetrics := loadBaselineClassificationMetrics()
    if baselineMetrics == nil {
        t.Skip("No baseline metrics available, establishing new baseline")
    }
    
    // Measure current performance
    currentMetrics := measureCurrentClassificationPerformance()
    
    // Check for significant drift
    accuracyDrift := baselineMetrics.Accuracy - currentMetrics.Accuracy
    if accuracyDrift > 2.0 { // 2% accuracy drop threshold
        t.Errorf("Significant accuracy drift detected: %.2f%% drop", accuracyDrift)
    }
    
    // Check dangerous operation recall drift (most critical)
    dangerousRecallDrift := baselineMetrics.DangerousRecall - currentMetrics.DangerousRecall
    if dangerousRecallDrift > 0.001 { // 0.1% drift for critical class
        t.Errorf("CRITICAL: Dangerous operation recall drift: %.3f drop", dangerousRecallDrift)
    }
    
    // Check prediction time drift
    timeDriftRatio := currentMetrics.AvgPredictionTime / baselineMetrics.AvgPredictionTime
    if timeDriftRatio > 1.5 { // 50% performance degradation
        t.Errorf("Prediction time drift detected: %.1fx slower", timeDriftRatio)
    }
    
    // Check for concept drift patterns
    if detectConceptDrift(baselineMetrics, currentMetrics) {
        t.Errorf("Concept drift detected in classification patterns")
    }
}
```

---

## Integration & End-to-End Classification Tests

### Full Pipeline Classification Tests

```go
// test/ai-safety/classification/e2e_classification_test.go

func TestEndToEndClassificationPipeline(t *testing.T) {
    // Test complete classification pipeline from natural language to final safety decision
    e2eScenarios := []struct {
        naturalLanguageQuery string
        expectedKubectlCmd   string
        expectedSafetyLevel  classifier.SafetyLevel
        expectedWorkflow     string
        description          string
    }{
        {
            naturalLanguageQuery: "show me all the pods that are currently running",
            expectedKubectlCmd:   "kubectl get pods --field-selector=status.phase=Running",
            expectedSafetyLevel:  classifier.SAFE,
            expectedWorkflow:     "immediate_execution",
            description:         "Simple read-only query",
        },
        {
            naturalLanguageQuery: "scale up the frontend deployment to handle more traffic",
            expectedKubectlCmd:   "kubectl scale deployment frontend --replicas=5",
            expectedSafetyLevel:  classifier.WARNING,
            expectedWorkflow:     "user_confirmation",
            description:         "Resource modification query",
        },
        {
            naturalLanguageQuery: "I need to clean up all the failed pods in production",
            expectedKubectlCmd:   "kubectl delete pods --field-selector=status.phase=Failed -n production",
            expectedSafetyLevel:  classifier.DANGEROUS,
            expectedWorkflow:     "multi_step_approval",
            description:         "Bulk deletion in production",
        },
    }

    // Full pipeline components
    nlpProcessor := nlp.NewProcessor()
    safetyClassifier := classifier.NewSafetyClassifier()
    workflowManager := workflow.NewApprovalWorkflowManager()

    for _, scenario := range e2eScenarios {
        t.Run(fmt.Sprintf("E2E_%s", scenario.description), func(t *testing.T) {
            // Step 1: Natural language processing
            nlpResult, err := nlpProcessor.ProcessQuery(scenario.naturalLanguageQuery)
            if err != nil {
                t.Fatalf("NLP processing failed: %v", err)
            }

            // Verify kubectl command generation
            if nlpResult.GeneratedCommand != scenario.expectedKubectlCmd {
                t.Logf("Command generation variation - Expected: %s, Got: %s",
                    scenario.expectedKubectlCmd, nlpResult.GeneratedCommand)
                // Allow semantic equivalence rather than exact match
                if !areCommandsSemanticallySimilar(nlpResult.GeneratedCommand, scenario.expectedKubectlCmd) {
                    t.Errorf("Generated command semantically different from expected")
                }
            }

            // Step 2: Safety classification
            classificationResult, err := safetyClassifier.ClassifyCommand(nlpResult.GeneratedCommand)
            if err != nil {
                t.Fatalf("Safety classification failed: %v", err)
            }

            // Verify safety level
            if classificationResult.SafetyLevel != scenario.expectedSafetyLevel {
                t.Errorf("Safety level mismatch: expected %v, got %v",
                    scenario.expectedSafetyLevel, classificationResult.SafetyLevel)
            }

            // Step 3: Workflow determination
            workflowResult, err := workflowManager.DetermineWorkflow(classificationResult)
            if err != nil {
                t.Fatalf("Workflow determination failed: %v", err)
            }

            // Verify workflow selection
            if workflowResult.WorkflowType != scenario.expectedWorkflow {
                t.Errorf("Workflow type mismatch: expected %s, got %s",
                    scenario.expectedWorkflow, workflowResult.WorkflowType)
            }

            // Verify end-to-end consistency
            validateE2EConsistency(t, scenario.naturalLanguageQuery, nlpResult, 
                                   classificationResult, workflowResult)

            t.Logf("E2E Success: %s -> %s -> %v -> %s",
                scenario.naturalLanguageQuery, nlpResult.GeneratedCommand,
                classificationResult.SafetyLevel, workflowResult.WorkflowType)
        })
    }
}
```

---

## Performance Benchmarks & Optimization

### Classification Performance Benchmarks

```go
func BenchmarkClassificationPerformance(b *testing.B) {
    classifier := classifier.NewSafetyClassifier()
    testCommands := generateBenchmarkCommands(1000) // Mix of all safety levels
    
    b.ResetTimer()
    
    // Benchmark single-threaded classification
    b.Run("Single_Thread", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            cmd := testCommands[i%len(testCommands)]
            _, err := classifier.ClassifyCommand(cmd)
            if err != nil {
                b.Fatalf("Classification failed: %v", err)
            }
        }
    })
    
    // Benchmark concurrent classification
    b.Run("Concurrent_10", func(b *testing.B) {
        var wg sync.WaitGroup
        semaphore := make(chan struct{}, 10)
        
        for i := 0; i < b.N; i++ {
            wg.Add(1)
            go func(idx int) {
                defer wg.Done()
                semaphore <- struct{}{}
                defer func() { <-semaphore }()
                
                cmd := testCommands[idx%len(testCommands)]
                _, err := classifier.ClassifyCommand(cmd)
                if err != nil {
                    b.Errorf("Classification failed: %v", err)
                }
            }(i)
        }
        wg.Wait()
    })
    
    // Benchmark memory usage
    b.Run("Memory_Usage", func(b *testing.B) {
        var m runtime.MemStats
        runtime.GC()
        runtime.ReadMemStats(&m)
        startAlloc := m.Alloc
        
        for i := 0; i < b.N; i++ {
            cmd := testCommands[i%len(testCommands)]
            _, err := classifier.ClassifyCommand(cmd)
            if err != nil {
                b.Fatalf("Classification failed: %v", err)
            }
        }
        
        runtime.GC()
        runtime.ReadMemStats(&m)
        endAlloc := m.Alloc
        
        avgMemoryPerClassification := (endAlloc - startAlloc) / uint64(b.N)
        if avgMemoryPerClassification > 1024 { // 1KB limit per classification
            b.Errorf("Memory usage too high: %d bytes per classification", avgMemoryPerClassification)
        }
    })
}
```

This comprehensive Safety Classification Validator ensures KubeChat's three-tier safety system operates with 100% accuracy for dangerous operations while maintaining high performance and contextual awareness. The framework provides exhaustive validation across all classification scenarios with zero tolerance for false negatives on dangerous operations.

---

*🧪 Generated by Test Architect Quinn*  
*Priority: CRITICAL - 100% dangerous operation detection required*
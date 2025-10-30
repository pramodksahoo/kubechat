type PlanRiskAnnotation = {
  severity: string;
  code: string;
  description: string;
};

type PlanTargetDescriptor = {
  cluster: string;
  namespace: string;
  resource: string;
};

type PlanStep = {
  sequence: number;
  title: string;
  description: string;
  command: string;
  operationType: string;
  target: PlanTargetDescriptor;
  dryRunAvailable: boolean;
  affectedResources?: string[];
  risk: PlanRiskAnnotation;
  diffPreview?: Record<string, unknown>;
};

type PlanRiskSummary = {
  level: string;
  justifications: string[];
};

type PlanDraft = {
  id: string;
  prompt: string;
  targetCluster: string;
  targetNamespace: string;
  confidence: number;
  scopeSignals: Record<string, string>;
  steps: PlanStep[];
  generatedAt: string;
  generationLatency: number;
  riskSummary: PlanRiskSummary;
};

type PlanRecord = {
  plan: PlanDraft;
  storedAt?: string;
  expiresAt?: string;
};

type PlanPromptResponse = {
  plan: PlanDraft;
  metrics: {
    generationDurationMs: number;
    capturedAt: string;
  };
  storedAt?: string;
  expiresAt?: string;
};

export type {
  PlanDraft,
  PlanRecord,
  PlanPromptResponse,
  PlanRiskAnnotation,
  PlanRiskSummary,
  PlanStep,
  PlanTargetDescriptor,
};

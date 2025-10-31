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

type PlanParameters = {
  namespace: string;
  labels: Record<string, string>;
  replicaOverrides: Record<string, number>;
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
  parameters: PlanParameters;
};

type PlanRevisionChange = {
  field: string;
  before?: unknown;
  after?: unknown;
  resource?: string;
  stepSequence?: number;
};

type PlanRevision = {
  version: number;
  updatedAt: string;
  updatedBy?: string;
  changes: PlanRevisionChange[];
};

type PlanRecord = {
  plan: PlanDraft;
  storedAt?: string;
  expiresAt?: string;
  revisions?: PlanRevision[];
};

type PlanPromptResponse = {
  plan: PlanDraft;
  metrics: {
    generationDurationMs: number;
    capturedAt: string;
  };
  storedAt?: string;
  expiresAt?: string;
  revisions?: PlanRevision[];
};

export type {
  PlanDraft,
  PlanParameters,
  PlanRecord,
  PlanPromptResponse,
  PlanRevision,
  PlanRevisionChange,
  PlanRiskAnnotation,
  PlanRiskSummary,
  PlanStep,
  PlanTargetDescriptor,
};

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import {
  applyOptimisticPlanUpdate,
  closePlan,
  fetchPlanById,
  setPlanRecord,
  rollbackPlanUpdate,
  updatePlanParameters,
} from "@/data/Plans/PlanPreviewSlice";
import { useAppDispatch, useAppSelector } from "@/redux/hooks";
import { useMediaQuery } from "@/hooks/use-media-query";
import { useRouterState } from "@tanstack/react-router";
import { Copy, Loader2, TriangleAlert } from "lucide-react";
import { PlanRecord, PlanRevision, PlanStep } from "@/types";
import { toast } from "sonner";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { API_VERSION, PLANS_ENDPOINT } from "@/constants";
import { cn } from "@/lib/utils";

const DESKTOP_BREAKPOINT_QUERY = "(min-width: 1024px)";
const DRAWER_WIDTH = 480;

const LoadingState = () => (
  <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
    <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
    <p className="text-sm text-muted-foreground">Preparing plan preview…</p>
    <Skeleton className="h-6 w-3/4" />
    <Skeleton className="h-4 w-1/2" />
    <Skeleton className="h-24 w-full" />
  </div>
);

const EmptyState = () => (
  <div className="flex h-full flex-col items-center justify-center gap-2 p-6 text-center text-muted-foreground">
    <TriangleAlert className="h-8 w-8" />
    <p>No plan selected. Generate a plan from Chat or choose a visual action to preview.</p>
  </div>
);

const ErrorState = ({ message }: { message: string }) => (
  <div className="flex h-full flex-col items-center justify-center gap-2 p-6 text-center text-destructive">
    <TriangleAlert className="h-8 w-8" />
    <p>{message}</p>
  </div>
);

type ResourceStat = {
  key: string;
  label: string;
  occurrences: number;
  cluster: string;
  namespace: string;
};

const buildResourceStats = (steps: PlanStep[]): ResourceStat[] => {
  const map = new Map<string, ResourceStat>();

  steps.forEach((step) => {
    (step.affectedResources ?? []).forEach((resource) => {
      const key = `${step.target.cluster}|${step.target.namespace}|${resource}`;
      const existing = map.get(key);
      if (existing) {
        existing.occurrences += 1;
      } else {
        map.set(key, {
          key,
          label: resource,
          occurrences: 1,
          cluster: step.target.cluster,
          namespace: step.target.namespace,
        });
      }
    });
  });

  return Array.from(map.values()).sort((a, b) => a.label.localeCompare(b.label));
};

const collapseWhitespace = (value: string) => value.replace(/\s+/g, " ").trim();

const formatLabelsForTextarea = (labels: Record<string, string>) =>
  Object.entries(labels)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, value]) => `${key}=${value}`)
    .join("\n");

const parseLabelsFromTextarea = (
  input: string,
): { labels: Record<string, string>; error?: string } => {
  const trimmed = input.trim();
  if (!trimmed) {
    return { labels: {} };
  }
  const labels: Record<string, string> = {};
  const lines = trimmed.split(/\n+/);
  for (const line of lines) {
    const candidate = line.trim();
    if (!candidate) {
      continue;
    }
    const parts = candidate.split("=");
    if (parts.length !== 2) {
      return { labels: {}, error: `Label "${candidate}" must use key=value format.` };
    }
    const key = parts[0]?.trim();
    const value = parts[1]?.trim();
    if (!key || !value) {
      return { labels: {}, error: `Label "${candidate}" must include both key and value.` };
    }
    labels[key] = value;
  }
  return { labels };
};

const formatLabelSelector = (labels: Record<string, string>) =>
  Object.entries(labels)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, value]) => `${key}=${value}`)
    .join(",");

const formatReplicaScopeSignal = (replicaOverrides: Record<string, number>) =>
  Object.entries(replicaOverrides)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([resource, value]) => `${resource}=${value}`)
    .join(",");

const setFlagValue = (command: string, flag: string, value: string | null) => {
  let updated = command;
  const flagEquals = new RegExp(`${flag}=([^\\s]+)`);
  const flagSpace = new RegExp(`${flag}\\s+([^\\s]+)`);

  if (!value) {
    updated = updated.replace(flagEquals, "");
    updated = updated.replace(flagSpace, "");
    return collapseWhitespace(updated);
  }

  if (flagEquals.test(updated)) {
    return collapseWhitespace(updated.replace(flagEquals, `${flag}=${value}`));
  }
  if (flagSpace.test(updated)) {
    return collapseWhitespace(updated.replace(flagSpace, `${flag} ${value}`));
  }
  return `${collapseWhitespace(updated)} ${flag}=${value}`;
};

const clonePlanRecord = (record: PlanRecord): PlanRecord => {
  if (typeof structuredClone === "function") {
    return structuredClone(record);
  }
  return JSON.parse(JSON.stringify(record)) as PlanRecord;
};

const deriveReplicaInputs = (plan: PlanRecord) => {
  const overrides = plan.plan.parameters?.replicaOverrides ?? {};
  const inputs: Record<string, string> = {};

  plan.plan.steps.forEach((step) => {
    if (!step.diffPreview || !("replicas" in step.diffPreview)) {
      return;
    }
    const resource = step.target.resource;
    if (overrides[resource] !== undefined) {
      inputs[resource] = String(overrides[resource]);
      return;
    }
    const diff = step.diffPreview["replicas"] as Record<string, unknown> | undefined;
    const toValue = diff?.to;
    if (typeof toValue === "number" && Number.isFinite(toValue)) {
      inputs[resource] = String(toValue);
      return;
    }
    if (typeof toValue === "string" && toValue.trim().length > 0 && !Number.isNaN(Number.parseInt(toValue, 10))) {
      inputs[resource] = String(Number.parseInt(toValue, 10));
      return;
    }
    inputs[resource] = "";
  });

  return inputs;
};

const prepareReplicaOverrides = (
  inputs: Record<string, string>,
): { overrides: Record<string, number>; error?: string } => {
  const overrides: Record<string, number> = {};
  for (const [resource, value] of Object.entries(inputs)) {
    const trimmed = value.trim();
    if (!trimmed) {
      continue;
    }
    const parsed = Number.parseInt(trimmed, 10);
    if (!Number.isFinite(parsed) || parsed <= 0) {
      return { overrides: {}, error: `Replica override for ${resource} must be a positive number.` };
    }
    overrides[resource] = parsed;
  }
  return { overrides };
};

const buildOptimisticPlan = (
  current: PlanRecord,
  params: { namespace: string; labels: Record<string, string>; replicaOverrides: Record<string, number> },
) => {
  const clone = clonePlanRecord(current);
  const namespace = params.namespace.trim() || current.plan.targetNamespace;
  const selector = formatLabelSelector(params.labels);
  const replicaSignal = formatReplicaScopeSignal(params.replicaOverrides);

  clone.plan.targetNamespace = namespace;
  clone.plan.parameters = {
    namespace,
    labels: params.labels,
    replicaOverrides: params.replicaOverrides,
  };

  clone.plan.scopeSignals = {
    ...(clone.plan.scopeSignals ?? {}),
    target_namespace: namespace,
  };
  if (selector) {
    clone.plan.scopeSignals.label_selector = selector;
  } else {
    delete clone.plan.scopeSignals.label_selector;
  }
  if (replicaSignal) {
    clone.plan.scopeSignals.replica_overrides = replicaSignal;
  } else {
    delete clone.plan.scopeSignals.replica_overrides;
  }

  clone.plan.steps = clone.plan.steps.map((step) => {
    const updated: PlanStep = {
      ...step,
      target: { ...step.target, namespace },
      affectedResources: step.affectedResources ? [...step.affectedResources] : undefined,
      diffPreview: step.diffPreview ? { ...step.diffPreview } : undefined,
    };

    updated.command = setFlagValue(step.command, "--namespace", namespace);
    updated.command = selector ? setFlagValue(updated.command, "--selector", selector) : setFlagValue(updated.command, "--selector", null);

    const override = params.replicaOverrides[updated.target.resource];
    if (override !== undefined) {
      updated.command = setFlagValue(updated.command, "--replicas", String(override));
      const diff = (step.diffPreview?.["replicas"] as Record<string, unknown> | undefined) ?? {};
      const nextDiff = { ...diff, to: override };
      if (!updated.diffPreview) {
        updated.diffPreview = {};
      }
      updated.diffPreview.replicas = nextDiff;
    }

    return updated;
  });

  return clone;
};

const StepItem = ({ step }: { step: PlanStep }) => {
  const handleCopy = () => {
    navigator.clipboard
      .writeText(step.command)
      .then(() => toast.success("Command copied", { description: step.command }))
      .catch(() => toast.error("Failed to copy command"));
  };

  return (
    <div className="rounded-lg border border-border/60 bg-card p-4 shadow-floating">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-medium uppercase text-muted-foreground">Step {step.sequence}</p>
          <h3 className="text-base font-semibold leading-tight">{step.title}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{step.description}</p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-xs capitalize">
            {step.operationType}
          </Badge>
          <Badge
            variant={step.risk.severity.toLowerCase() === "high" ? "destructive" : "secondary"}
            className="text-xs capitalize"
          >
            {step.risk.severity}
          </Badge>
        </div>
      </div>

      <div className="mt-3 rounded-md bg-muted/60 p-3 text-sm">
        <p className="text-xs uppercase text-muted-foreground">Command</p>
        <div className="mt-1 flex items-start justify-between gap-2">
          <code className="flex-1 whitespace-pre-wrap break-words text-xs text-foreground/90">{step.command}</code>
          <Button size="icon" variant="ghost" className="h-8 w-8" onClick={handleCopy}>
            <Copy className="h-4 w-4" />
            <span className="sr-only">Copy command</span>
          </Button>
        </div>
      </div>

      <div className="mt-3 grid gap-2 rounded-md border border-border/60 p-3 text-xs text-muted-foreground">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline" className="bg-background text-xs font-medium">
            Cluster: {step.target.cluster}
          </Badge>
          <Badge variant="outline" className="bg-background text-xs font-medium">
            Namespace: {step.target.namespace || "(default)"}
          </Badge>
          <Badge variant="outline" className="bg-background text-xs font-medium">
            Resource: {step.target.resource}
          </Badge>
          {step.dryRunAvailable ? (
            <Badge variant="outline" className="bg-success/15 text-success">
              Dry-run supported
            </Badge>
          ) : (
            <Badge variant="outline" className="bg-warning/20 text-warning">
              Dry-run unavailable
            </Badge>
          )}
        </div>
        <p className="text-xs italic">{step.risk.description}</p>
      </div>

      <div className="mt-3 rounded-md border border-border/60 border-dashed p-3 text-xs text-muted-foreground">
        Dry-run output will appear here once executed.
      </div>
    </div>
  );
};

const PlanParameterEditor = ({ plan, isUpdating }: { plan: PlanRecord; isUpdating: boolean }) => {
  const dispatch = useAppDispatch();
  const [namespace, setNamespace] = useState(plan.plan.parameters?.namespace ?? plan.plan.targetNamespace ?? "");
  const [labelsInput, setLabelsInput] = useState(formatLabelsForTextarea(plan.plan.parameters?.labels ?? {}));
  const [replicaValues, setReplicaValues] = useState<Record<string, string>>(deriveReplicaInputs(plan));
  const [validationError, setValidationError] = useState<string | null>(null);

  useEffect(() => {
    setNamespace(plan.plan.parameters?.namespace ?? plan.plan.targetNamespace ?? "");
    setLabelsInput(formatLabelsForTextarea(plan.plan.parameters?.labels ?? {}));
    setReplicaValues(deriveReplicaInputs(plan));
    setValidationError(null);
  }, [plan, plan.plan.id, plan.plan.parameters, plan.plan.targetNamespace]);

  const replicaTargets = useMemo(
    () =>
      plan.plan.steps
        .filter((step) => step.diffPreview && "replicas" in step.diffPreview)
        .map((step) => ({ sequence: step.sequence, title: step.title, resource: step.target.resource })),
    [plan.plan.steps],
  );

  const hasChanges = useMemo(() => {
    const trimmedNamespace = namespace.trim();
    const originalNamespace = plan.plan.parameters?.namespace ?? plan.plan.targetNamespace ?? "";
    if (trimmedNamespace !== originalNamespace) {
      return true;
    }

    const formattedOriginalLabels = formatLabelsForTextarea(plan.plan.parameters?.labels ?? {});
    if (collapseWhitespace(labelsInput) !== collapseWhitespace(formattedOriginalLabels)) {
      return true;
    }

    const prepared = prepareReplicaOverrides(replicaValues);
    if (prepared.error) {
      return true;
    }
    const originalOverrides = plan.plan.parameters?.replicaOverrides ?? {};
    const keys = new Set([...Object.keys(originalOverrides), ...Object.keys(prepared.overrides)]);
    for (const key of keys) {
      if ((originalOverrides[key] ?? 0) !== (prepared.overrides[key] ?? 0)) {
        return true;
      }
    }
    return false;
  }, [labelsInput, namespace, plan.plan.parameters?.labels, plan.plan.parameters?.namespace, plan.plan.parameters?.replicaOverrides, plan.plan.targetNamespace, replicaValues]);

  const handleReplicaChange = useCallback((resource: string, value: string) => {
    setReplicaValues((prev) => ({ ...prev, [resource]: value }));
  }, []);

  const handleReset = useCallback(() => {
    setNamespace(plan.plan.parameters?.namespace ?? plan.plan.targetNamespace ?? "");
    setLabelsInput(formatLabelsForTextarea(plan.plan.parameters?.labels ?? {}));
    setReplicaValues(deriveReplicaInputs(plan));
    setValidationError(null);
  }, [plan]);

  const handleSave = useCallback(async () => {
    const parsedLabels = parseLabelsFromTextarea(labelsInput);
    if (parsedLabels.error) {
      setValidationError(parsedLabels.error);
      toast.error("Invalid labels", { description: parsedLabels.error });
      return;
    }

    const preparedReplicas = prepareReplicaOverrides(replicaValues);
    if (preparedReplicas.error) {
      setValidationError(preparedReplicas.error);
      toast.error("Invalid replica override", { description: preparedReplicas.error });
      return;
    }

    setValidationError(null);

    const trimmedNamespace = namespace.trim() || plan.plan.targetNamespace;
    const optimistic = buildOptimisticPlan(plan, {
      namespace: trimmedNamespace,
      labels: parsedLabels.labels,
      replicaOverrides: preparedReplicas.overrides,
    });

    dispatch(applyOptimisticPlanUpdate(optimistic));

    try {
      await dispatch(
        updatePlanParameters({
          planId: plan.plan.id,
          targetNamespace: trimmedNamespace,
          labels: parsedLabels.labels,
          replicaOverrides: preparedReplicas.overrides,
          updatedBy: "ui",
        }),
      ).unwrap();
      toast.success("Plan parameters saved", {
        description: `Namespace set to ${trimmedNamespace}`,
      });
    } catch (error) {
      dispatch(rollbackPlanUpdate());
      toast.error("Failed to update plan", {
        description: (error as { message?: string })?.message ?? "Please retry.",
      });
    }
  }, [dispatch, labelsInput, namespace, plan, replicaValues]);

  return (
    <section className="space-y-3 rounded-lg border border-border/60 bg-card p-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">Plan parameters</h3>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={handleReset} disabled={isUpdating}>
            Reset
          </Button>
          <Button size="sm" onClick={handleSave} disabled={isUpdating || !hasChanges}>
            {isUpdating ? <Loader2 className="mr-2 h-3 w-3 animate-spin" /> : null}
            Save
          </Button>
        </div>
      </div>

      <div className="space-y-3">
        <div className="space-y-1">
          <Label htmlFor="plan-namespace" className="text-xs uppercase text-muted-foreground">
            Target namespace
          </Label>
          <Input
            id="plan-namespace"
            placeholder="payments"
            value={namespace}
            onChange={(event) => setNamespace(event.target.value)}
          />
        </div>

        <div className="space-y-1">
          <Label htmlFor="plan-labels" className="text-xs uppercase text-muted-foreground">
            Label selectors (key=value per line)
          </Label>
          <Textarea
            id="plan-labels"
            rows={3}
            placeholder="app=checkout"
            value={labelsInput}
            onChange={(event) => setLabelsInput(event.target.value)}
          />
        </div>

        {replicaTargets.length > 0 ? (
          <div className="space-y-2">
            <p className="text-xs uppercase text-muted-foreground">Replica overrides</p>
            <div className="grid gap-2">
              {replicaTargets.map((target) => (
                <div key={target.resource} className="flex items-center gap-3">
                  <Label htmlFor={`replica-${target.resource}`} className="flex-1 text-sm">
                    Step {target.sequence}: {target.title}
                  </Label>
                  <Input
                    id={`replica-${target.resource}`}
                    type="number"
                    min={1}
                    className="w-32"
                    value={replicaValues[target.resource] ?? ""}
                    onChange={(event) => handleReplicaChange(target.resource, event.target.value)}
                  />
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {validationError ? <p className="text-xs text-destructive">{validationError}</p> : null}
      </div>

      <p className="text-xs text-muted-foreground">
        Updates persist for all viewers and trigger a server broadcast via plan_update events.
      </p>
    </section>
  );
};

const formatRevisionValue = (value: unknown) => {
  if (value === undefined) {
    return "–";
  }
  if (value === null) {
    return "null";
  }
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number") {
    return value.toString();
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  return String(value);
};

const PlanRevisionHistory = ({ revisions }: { revisions?: PlanRevision[] }) => {
  if (!revisions || revisions.length === 0) {
    return null;
  }

  return (
    <section className="space-y-3 rounded-lg border border-border/60 bg-card p-4">
      <h3 className="text-sm font-semibold">Revision history</h3>
      <div className="space-y-2 text-xs">
        {revisions
          .slice()
          .reverse()
          .map((revision) => (
            <div key={revision.version} className="rounded-md border border-border/50 px-3 py-2">
              <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
                <span className="font-medium">Version {revision.version}</span>
                <span className="text-muted-foreground">{new Date(revision.updatedAt).toLocaleString()}</span>
              </div>
              {revision.updatedBy ? (
                <p className="text-muted-foreground">Updated by {revision.updatedBy}</p>
              ) : null}
              <ul className="mt-2 space-y-1">
                {revision.changes.map((change, index) => (
                  <li key={`${revision.version}-${index}`} className="rounded bg-muted px-2 py-1">
                    <span className="font-medium">{change.field}</span>
                    {change.resource ? <span className="ml-1 text-muted-foreground">({change.resource})</span> : null}
                    <span className="ml-2 text-muted-foreground">
                      {formatRevisionValue(change.before)} → {formatRevisionValue(change.after)}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          ))}
      </div>
    </section>
  );
};

const PlanPreviewContent = ({ plan, onClose, isUpdating }: { plan: PlanRecord; onClose: () => void; isUpdating: boolean }) => {
  const resourceStats = useMemo(() => buildResourceStats(plan.plan.steps), [plan.plan.steps]);

  return (
    <div className="flex h-full flex-col">
      <div className="border-b p-4">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs uppercase text-muted-foreground">Plan Preview</p>
            <h2 className="text-lg font-semibold leading-tight">{plan.plan.prompt}</h2>
          </div>
          <Badge
            variant={plan.plan.riskSummary.level.toLowerCase() === "high" ? "destructive" : "secondary"}
            className="capitalize"
          >
            {plan.plan.riskSummary.level || "unknown"} risk
          </Badge>
        </div>
        <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
          <Badge variant="outline" className="bg-background">
            Cluster: {plan.plan.targetCluster}
          </Badge>
          <Badge variant="outline" className="bg-background">
            Namespace: {plan.plan.targetNamespace || "(default)"}
          </Badge>
          <span>
            Generated {new Date(plan.plan.generatedAt).toLocaleString()} • Latency {plan.plan.generationLatency} ms
          </span>
        </div>
      </div>

      <ScrollArea className="flex-1 p-4">
        <div className="space-y-4">
          <PlanParameterEditor plan={plan} isUpdating={isUpdating} />
          <section className="rounded-lg border border-border/60 bg-card p-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold">Affected resources</h3>
              <span className="text-xs text-muted-foreground">{resourceStats.length} unique</span>
            </div>
            <Collapsible>
              <CollapsibleTrigger className="mt-3 flex w-full items-center justify-between rounded-md border border-border/50 px-3 py-2 text-sm">
                View resources
              </CollapsibleTrigger>
              <CollapsibleContent>
                <ul className="mt-2 space-y-1">
                  {resourceStats.map((resource) => (
                    <li
                      key={resource.key}
                      className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border/40 px-3 py-2 text-xs"
                    >
                      <div className="flex flex-col">
                        <span className="font-medium text-foreground">{resource.label}</span>
                        <span className="text-muted-foreground">
                          Cluster: {resource.cluster} · Namespace: {resource.namespace || "(default)"}
                        </span>
                      </div>
                      <Badge variant="secondary">{resource.occurrences} step(s)</Badge>
                    </li>
                  ))}
                </ul>
              </CollapsibleContent>
            </Collapsible>
          </section>

          <PlanRevisionHistory revisions={plan.revisions} />

          <section className="space-y-3">
            <h3 className="text-sm font-semibold">Steps</h3>
            {plan.plan.steps.map((step) => (
              <StepItem key={step.sequence} step={step} />
            ))}
          </section>
        </div>
      </ScrollArea>

      <Separator />
      <div className="flex gap-2 p-4">
        <Button
          variant="outline"
          className="flex-1"
          onClick={() =>
            toast.info("Dry-run requested", {
              description: "Dry-run execution will be wired up when backend support lands (AC3).",
            })
          }
        >
          Dry-run
        </Button>
        <Button className="flex-1" disabled>
          Approve &amp; Execute
        </Button>
        <Button
          variant="ghost"
          className="flex-1"
          onClick={() => {
            toast.message("Plan preview hidden", {
              description: "Re-open from chat history or future action previews.",
            });
            onClose();
          }}
        >
          Close
        </Button>
      </div>
    </div>
  );
};

const DesktopDrawer = ({ isOpen, children }: { isOpen: boolean; children: React.ReactNode }) => (
  <aside
    className={cn(
      "hidden h-screen flex-col border-l border-border/60 bg-background shadow-feature transition-all duration-300 lg:flex",
      isOpen ? "pointer-events-auto" : "pointer-events-none",
    )}
    style={{ width: isOpen ? DRAWER_WIDTH : 0 }}
    aria-hidden={!isOpen}
  >
    {isOpen ? children : null}
  </aside>
);

const PlanPreviewDrawer = () => {
  const dispatch = useAppDispatch();
  const { plan, isOpen, loading, error, updateInFlight } = useAppSelector((state) => state.planPreview);
  const isDesktop = useMediaQuery(DESKTOP_BREAKPOINT_QUERY);
  const router = useRouterState();
  const eventSourceRef = useRef<EventSource | null>(null);

  const planIdFromSearch = (router.location.search as Record<string, unknown>)?.plan as string | undefined;

  useEffect(() => {
    if (planIdFromSearch && plan?.plan.id !== planIdFromSearch) {
      dispatch(fetchPlanById(planIdFromSearch));
    }
  }, [dispatch, planIdFromSearch, plan?.plan.id]);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }
    const url = new URL(window.location.href);
    if (isOpen && plan?.plan.id) {
      url.searchParams.set("plan", plan.plan.id);
    } else {
      url.searchParams.delete("plan");
    }
    window.history.replaceState(null, "", url.toString());
  }, [isOpen, plan?.plan.id]);

  useEffect(() => {
    if (!isOpen) {
      return;
    }
    const listener = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        dispatch(closePlan());
      }
    };
    window.addEventListener("keydown", listener);
    return () => window.removeEventListener("keydown", listener);
  }, [dispatch, isOpen]);

  useEffect(() => {
    if (!isOpen || !plan?.plan.id) {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      return;
    }

    const source = new EventSource(`${API_VERSION}/${PLANS_ENDPOINT}/${plan.plan.id}/stream`);
    eventSourceRef.current = source;

    const handler = (event: Event) => {
      const message = event as MessageEvent<string>;
      try {
        const payload = JSON.parse(message.data) as PlanRecord;
        dispatch(setPlanRecord(payload));
      } catch (err) {
        console.error("Failed to parse plan_update event", err);
      }
    };

    source.addEventListener("plan_update", handler);
    source.onerror = () => {
      source.close();
      if (eventSourceRef.current === source) {
        eventSourceRef.current = null;
      }
    };

    return () => {
      source.removeEventListener("plan_update", handler);
      source.close();
      if (eventSourceRef.current === source) {
        eventSourceRef.current = null;
      }
    };
  }, [dispatch, isOpen, plan?.plan.id]);

  const handleClose = () => {
    dispatch(closePlan());
  };

  const content = (
    <div className="flex h-full flex-col">
      {loading ? (
        <LoadingState />
      ) : error ? (
        <ErrorState message={error.message || "Failed to load plan."} />
      ) : plan ? (
        <PlanPreviewContent plan={plan} onClose={handleClose} isUpdating={updateInFlight} />
      ) : (
        <EmptyState />
      )}
    </div>
  );

  if (!isDesktop) {
    return (
      <Sheet open={isOpen} onOpenChange={(open) => (open ? undefined : handleClose())}>
        <SheetContent side="right" className="w-full max-w-md p-0">
          {content}
        </SheetContent>
      </Sheet>
    );
  }

  return <DesktopDrawer isOpen={isOpen}>{content}</DesktopDrawer>;
};

export { PlanPreviewDrawer };

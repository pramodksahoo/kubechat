import { useEffect, useMemo } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { closePlan, fetchPlanById } from "@/data/Plans/PlanPreviewSlice";
import { useAppDispatch, useAppSelector } from "@/redux/hooks";
import { useMediaQuery } from "@/hooks/use-media-query";
import { useNavigate, useRouterState } from "@tanstack/react-router";
import { Copy, Loader2, TriangleAlert } from "lucide-react";
import { PlanRecord, PlanStep } from "@/types";
import { toast } from "sonner";
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

const StepItem = ({ step }: { step: PlanStep }) => {
  const handleCopy = () => {
    navigator.clipboard
      .writeText(step.command)
      .then(() => toast.success("Command copied", { description: step.command }))
      .catch(() => toast.error("Failed to copy command"));
  };

  return (
    <div className="rounded-lg border bg-card p-4 shadow-sm">
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

      <div className="mt-3 grid gap-2 rounded-md border p-3 text-xs text-muted-foreground">
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
            <Badge variant="outline" className="bg-emerald-500/10 text-emerald-700 dark:text-emerald-300">
              Dry-run supported
            </Badge>
          ) : (
            <Badge variant="outline" className="bg-amber-500/10 text-amber-700 dark:text-amber-300">
              Dry-run unavailable
            </Badge>
          )}
        </div>
        <p className="text-xs italic">{step.risk.description}</p>
      </div>

      <div className="mt-3 rounded-md border border-dashed p-3 text-xs text-muted-foreground">
        Dry-run output will appear here once executed.
      </div>
    </div>
  );
};

const PlanPreviewContent = ({ plan, onClose }: { plan: PlanRecord; onClose: () => void }) => {
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
          <section className="rounded-lg border bg-card p-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold">Affected resources</h3>
              <span className="text-xs text-muted-foreground">{resourceStats.length} unique</span>
            </div>
            <Collapsible>
              <CollapsibleTrigger className="mt-3 flex w-full items-center justify-between rounded-md border px-3 py-2 text-sm">
                View resources
              </CollapsibleTrigger>
              <CollapsibleContent>
                <ul className="mt-2 space-y-1">
                  {resourceStats.map((resource) => (
                    <li
                      key={resource.key}
                      className="flex flex-wrap items-center justify-between gap-2 rounded-md border px-3 py-2 text-xs"
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
      "hidden h-screen flex-col border-l bg-background shadow-lg transition-all duration-300 lg:flex",
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
  const { plan, isOpen, loading, error } = useAppSelector((state) => state.planPreview);
  const isDesktop = useMediaQuery(DESKTOP_BREAKPOINT_QUERY);
  const router = useRouterState();
  const navigate = useNavigate();

  const planIdFromSearch = (router.location.search as Record<string, unknown>)?.plan as string | undefined;

  useEffect(() => {
    if (planIdFromSearch && plan?.plan.id !== planIdFromSearch) {
      dispatch(fetchPlanById(planIdFromSearch));
    }
  }, [dispatch, planIdFromSearch, plan?.plan.id]);

  useEffect(() => {
    if (!plan || !plan.plan.id || !isOpen) {
      return;
    }
    navigate({
      replace: true,
      search: (prev) => ({
        ...prev,
        plan: plan.plan.id,
      }),
    });
  }, [isOpen, plan, navigate]);

  useEffect(() => {
    if (!isOpen) {
      navigate({
        replace: true,
        search: (prev) => {
          if (!("plan" in prev)) {
            return prev;
          }
          const { plan: _plan, ...rest } = prev as Record<string, unknown>;
          return rest;
        },
      });
    }
  }, [isOpen, navigate]);

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
        <PlanPreviewContent plan={plan} onClose={handleClose} />
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

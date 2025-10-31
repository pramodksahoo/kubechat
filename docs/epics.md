# kubechat - Epic Breakdown

**Author:** pramoda
**Date:** 2025-10-30T08:06:50.280372Z
**Project Level:** 4
**Target Scale:** 5–50 clusters (50–1,000 nodes; 5k–30k pods) across cloud + on-prem

---

## Overview

This document provides the detailed epic breakdown for kubechat, expanding on the high-level epic list in the [PRD](./PRD.md).

Each epic includes:

- Expanded goal and value proposition
- Complete story breakdown with user stories
- Acceptance criteria for each story
- Story sequencing and dependencies

> Kubechat is built as a fork of Kubechat; epics explicitly account for reusing and extending upstream modules while keeping a clean merge path.

**Epic Sequencing Principles:**

- Epic 1 establishes foundational infrastructure and initial functionality
- Subsequent epics build progressively, each delivering significant end-to-end value
- Stories within epics are vertically sliced and sequentially ordered
- No forward dependencies - each story builds only on previous work

---

## E1. AI-Assisted Command Planning
**Expanded Goal:** Deliver an explainable natural-language to kubectl planning engine that scopes commands accurately and highlights risk before execution.

**Foundational Alignment:** Builds on the Kubechat fork by extending core planning capabilities without breaking upstream compatibility, documenting divergences per FR-0 before layering new AI-assisted features. [Source: PRD.md §FR-0]

**Story E1.1: Capture Natural-Language Intents**

**Foundation Alignment:** Story E1.1 must reinforce the Kubechat baseline by integrating with existing modules before expanding NL planning logic, capturing adjustments required for future merges. [Source: PRD.md §FR-0]

As a SRE,
I want to describe crashlooping payment pods in prod,
So that I see an exact plan with commands, scope, and risks before anything runs.

**Acceptance Criteria:**
1. System accepts NL inputs and maps them to intent/target clusters within 5 seconds.
2. Plan preview lists exact kubectl/API calls with namespace, cluster, and affected objects.
3. User can edit parameters (namespace, labels, replica counts) before approving the plan.

**Prerequisites:** None

**Story E1.2: Surface Risk & Blast Radius Insights**

As a SRE,
I want to understand the blast radius of proposed actions,
So that I can quickly judge safety before execution.

**Acceptance Criteria:**
1. Each plan step displays risk level, affected resource counts, and dependencies.
2. Warnings surface when operations touch production or cross cluster boundaries.
3. Risk annotations update immediately when the user edits parameters.

**Prerequisites:** Story E1.1

**Story E1.3: Suggest Diagnostic Steps**

As a SRE,
I want to review system-recommended diagnostics,
So that I avoid missing critical observations before acting.

**Acceptance Criteria:**
1. Plans include recommended read-only diagnostics (logs, events, metrics) ahead of mutations.
2. User can accept/skip diagnostics and reordered plan reflects choice.
3. Diagnostics execute in scoped clusters/namespaces with guardrail status.

**Prerequisites:** Stories E1.1-E1.2

**Story E1.4: Post-Execution Summaries**

As a SRE,
I want to see AI summary of execution results,
So that I understand impact and next steps without parsing raw output.

**Acceptance Criteria:**
1. Upon completion, the system summarizes outputs, highlighting success/failure and suggested follow-ups.
2. Summary links to captured logs/metrics and stores in the incident timeline.
3. Users can annotate or edit the summary before it is archived.

**Prerequisites:** Stories E1.1-E1.3

## E2. Guarded Execution & Approvals
**Expanded Goal:** Enforce preview, approval, and rollback safeguards so every mutating action is auditable and reversible.

**Story E2.1: Dry-Run Previews for Mutations**

As a Platform Engineer,
I want to preview apply/scale/delete operations,
So that I avoid accidental destructive changes.

**Acceptance Criteria:**
1. Mutating operations default to dry-run unless explicitly overridden with justification.
2. Preview displays diff of objects touched and expected outcomes.
3. Dry-run output is logged in the timeline prior to execution.

**Prerequisites:** Stories E1.1-E1.2

**Story E2.2: RBAC & Policy Evaluation**

As a Platform Engineer,
I want to run commands within policy boundaries,
So that I get clear allow/deny feedback before execution.

**Acceptance Criteria:**
1. System evaluates RBAC permissions and custom guardrails before executing commands.
2. Denials return actionable remediation guidance (missing role, disallowed namespace, time window).
3. Policy decisions are logged with rule IDs in the audit trail.

**Prerequisites:** Story E2.1

**Story E2.3: Two-Person Approval Workflow**

As a SRE,
I want to request approvals for production changes,
So that I can escalate risky actions without leaving the console.

**Acceptance Criteria:**
1. Users can mark steps as requiring approval; approvers receive context-rich notifications.
2. Approvals record approver identity, timestamp, and comments in the timeline.
3. System blocks execution until required approvals are received or overridden with policy exception.

**Prerequisites:** Story E2.2

**Story E2.4: Rollback Snapshot & Execution**

As a SRE,
I want to revert failed operations quickly,
So that I can restore service with minimal manual effort.

**Acceptance Criteria:**
1. For each mutation, system captures pre-change state (replica counts, images, configs).
2. Rollback command is generated and previewed alongside the original action.
3. Executing rollback updates audit logs and summaries just like forward actions.

**Prerequisites:** Stories E2.1-E2.3

## E3. Multi-Cluster Visibility Console
**Expanded Goal:** Expose a synchronized chat and visual console covering multi-cluster diagnostics with clear context cues.

**Story E3.1: Cluster & Namespace Selector**

As a SRE,
I want to switch clusters safely,
So that I always know where commands will run.

**Acceptance Criteria:**
1. Context banner displays current cluster/namespace with warning badges when scope changes.
2. Keyboard shortcuts cycle between saved cluster sets (prod, staging, canary).
3. All commands inherit selected scope unless explicitly overridden.

**Prerequisites:** Stories E1.1-E1.2

**Story E3.2: Aggregated Resource Views**

As a Platform Engineer,
I want to view pods/nodes across clusters,
So that I spot issues without opening multiple dashboards.

**Acceptance Criteria:**
1. Virtualized tables display pods/nodes/events aggregated by selected clusters with cluster tags.
2. Views support filtering by namespace, label, and status with server-side pagination.
3. Clicking an item focuses corresponding diagnostics in the chat thread.

**Prerequisites:** Story E3.1

**Story E3.3: Scoped Action Shortcuts**

As a Platform Engineer,
I want to execute describe/logs against filtered sets,
So that I can run read-only diagnostics instantly.

**Acceptance Criteria:**
1. Inline actions allow describe/log/tail/exec against selected resources and respect guardrails.
2. Results stream into the chat timeline with cluster badges.
3. Users can promote read-only actions into command plans for follow-up.

**Prerequisites:** Stories E3.1-E3.2

**Story E3.4: Timeline & Event Overlay**

As a SRE,
I want to see combined operations and alerts,
So that I maintain shared context during incidents.

**Acceptance Criteria:**
1. Unified timeline merges operator actions, AI summaries, and imported alerts (Prometheus, Argo deploys).
2. Users can annotate timeline entries and link to tickets/runbooks.
3. Timeline supports export to incident reports and audit logs.

**Prerequisites:** Stories E3.1-E3.3

## E4. Streaming Observability & Insights
**Expanded Goal:** Provide integrated logs, exec sessions, metrics, and AI insights within the conversational workflow.

**Story E4.1: Integrated Log Tail**

As a SRE,
I want to stream pod logs inside chat,
So that I avoid juggling terminals.

**Acceptance Criteria:**
1. Users can initiate log tails with filters (container, regex, since) and view streams inline.
2. Streams are resumable and annotate timeline entries with cluster/pod IDs.
3. UI highlights anomalies (error spikes) using lightweight heuristics.

**Prerequisites:** Story E3.3

**Story E4.2: Secure Exec Sessions**

As a Platform Engineer,
I want to run guarded exec commands,
So that I can execute troubleshooting commands securely.

**Acceptance Criteria:**
1. Exec sessions require explicit confirmation and optional approval paths.
2. Session transcript is recorded with redacted secrets and linked to audit logs.
3. System auto-terminates idle sessions after configurable timeout.

**Prerequisites:** Stories E2.2, E4.1

**Story E4.3: Telemetry Correlation**

As a SRE,
I want to compare recent deploys and metrics,
So that I understand causal signals faster.

**Acceptance Criteria:**
1. Console surfaces recent deploys, config changes, and alert spikes related to targeted resources.
2. AI generates short insight cards (e.g., OOM restarts align with missing limits).
3. Users can pin insights to the incident timeline for handoff.

**Prerequisites:** Stories E3.4, E4.1

**Story E4.4: Post-Incident Note Automation**

As a Ops Manager,
I want to export timeline and insights,
So that I deliver clean summaries to stakeholders.

**Acceptance Criteria:**
1. Users can generate draft incident notes containing timeline, command summaries, and follow-ups.
2. Output supports markdown export and attachment to ticketing systems via webhook.
3. Notes capture outstanding action items tagged to owners.

**Prerequisites:** Stories E4.1-E4.3

## E5. Audit, Compliance & Policy Framework
**Expanded Goal:** Deliver tamper-evident audit logging, policy enforcement, and evidence exports for regulated teams.

**Story E5.1: Audit Log Ingestion**

As a Compliance Lead,
I want to review privileged actions,
So that I can trace who did what, where, when, and why.

**Acceptance Criteria:**
1. Every action logs operator identity, command payload, scope, timestamps, and approvals.
2. Logs are stored in append-only format with integrity checks.
3. Users can filter logs by cluster, user, action type, and time range.

**Prerequisites:** Stories E2.1-E2.3

**Story E5.2: Evidence Export**

As a Compliance Lead,
I want to export audit data,
So that I satisfy audit requests in under two hours.

**Acceptance Criteria:**
1. System generates CSV/JSON evidence packages aligned to audit scopes (time range, action types).
2. Exports include hash and metadata verifying integrity.
3. Packages can be delivered via download or pushed to security storage.

**Prerequisites:** Story E5.1

**Story E5.3: Policy Management**

As a Platform Manager,
I want to configure guardrail rules,
So that I enforce org standards consistently.

**Acceptance Criteria:**
1. Admins define policies for allowed namespaces, label selectors, time windows, and approval requirements.
2. Policy violations trigger alerts and are logged with remediation guidance.
3. Policies versioned with change history and rollback support.

**Prerequisites:** Stories E2.2, E5.1

**Story E5.4: Runbook & Recipe Catalog**

As a Platform Manager,
I want to standardize approved actions,
So that I reduce variance in live ops.

**Acceptance Criteria:**
1. Users can publish curated command recipes referencing policies and approvals.
2. Recipes track usage analytics and feedback for continuous improvement.
3. Catalog integrates with command planning so recipes seed AI suggestions.

**Prerequisites:** Stories E5.1-E5.3

**Story E5.5: Secrets Rotation & Credential Playbook**

As a Platform Manager,
I want a documented process for rotating multi-cluster credentials,
So that we maintain compliance and avoid configuration drift.

**Acceptance Criteria:**
1. Runbook covers rotation steps for per-cluster and hub installations.
2. Guidance includes secret storage (Secrets/SealedSecrets) and audit anchoring impacts.
3. Checklist reviewed with security/compliance stakeholders.

**Prerequisites:** Stories E5.1-E5.4

## E6. Packaging, Deployment & AI Provider Abstraction
**Expanded Goal:** Ship production-ready packages with flexible AI provider support and offline defaults while maintaining alignment with the upstream Kubechat project.

**Story E6.1: Single Binary Build**

As a DevOps Engineer,
I want to run Kubechat locally,
So that I can evaluate features without complex setup.

**Acceptance Criteria:**
1. Build process produces a self-contained binary bundling backend + UI assets.
2. Binary supports configuration via YAML/env for cluster access and AI providers.
3. Local mode defaults to SQLite storage and Ollama provider.

**Prerequisites:** Stories E1.1-E1.2

**Story E6.2: Helm Chart with CRDs**

As a Platform Engineer,
I want to deploy to cluster,
So that I can install and manage Kubechat using Kubernetes tooling.

**Acceptance Criteria:**
1. Helm chart defines AIProvider and AuditPolicy CRDs with validation schemas.
2. Chart supports Postgres or SQLite backend selection via values file.
3. Default values disable outbound egress; toggles exist for optional integrations.

**Prerequisites:** Story E6.1

**Story E6.3: AI Provider Abstraction Layer**

As a Platform Engineer,
I want to switch AI engines,
So that I can choose Ollama, GPT, Claude, or Gemini based on environment.

**Acceptance Criteria:**
1. Abstraction exposes provider-specific adapters with consistent configuration surface.
2. Policies enforce prompt/response redaction before external calls.
3. Provider health/status visible in UI with failover to local fallback.

**Prerequisites:** Stories E1.1, E6.1

**Story E6.4: CI/CD and Release Pipeline**

As a Ops Manager,
I want to trust deliveries,
So that I rely on signed multi-arch images and automated tests.

**Acceptance Criteria:**
1. Pipeline runs unit/integration/security tests, vulnerability scans, and SBOM generation.
2. Releases produce signed container images and checksums for binaries.
3. Release notes include compatibility matrix (Kubernetes N-3) and upgrade guides.

**Prerequisites:** Stories E6.1-E6.3

**Story E6.5: Redis/Object Storage Enablement Guides**

As a Platform Engineer,
I want clear guidance for enabling Redis and object storage add-ons,
So that customers can scale caching and export retention safely.

**Acceptance Criteria:**
1. Helm values documented for enabling Redis and S3/MinIO integrations.
2. Operational responsibilities assigned (infra team ownership noted).
3. Monitoring/alerting recommendations captured for optional components.

**Prerequisites:** Stories E6.1-E6.3

**Story E6.6: Provision cosign/OTEL Credentials for CI/CD**

As a DevOps Engineer,
I want cosign and OpenTelemetry credentials wired into the pipeline,
So that signed artifacts and trace exports function from day one.

**Acceptance Criteria:**
1. Cosign key management documented and secrets injected in GitHub Actions (or self-hosted runner).
2. OTEL endpoint/credentials configurable via pipeline secrets.
3. CI pipeline smoke test verifies signed image and OTEL export.

**Prerequisites:** Stories E6.1-E6.3



---

## Story Guidelines Reference

**Story Format:**

```
**Story [EPIC.N]: [Story Title]**

As a [user type],
I want [goal/desire],
So that [benefit/value].

**Acceptance Criteria:**
1. [Specific testable criterion]
2. [Another specific criterion]
3. [etc.]

**Prerequisites:** [Dependencies on previous stories, if any]
```

**Story Requirements:**

- **Vertical slices** - Complete, testable functionality delivery
- **Sequential ordering** - Logical progression within epic
- **No forward dependencies** - Only depend on previous work
- **AI-agent sized** - Completable in 2-4 hour focused session
- **Value-focused** - Integrate technical enablers into value-delivering stories

---

**For implementation:** Use the `create-story` workflow to generate individual story implementation plans from this epic breakdown.

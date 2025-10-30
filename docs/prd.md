# kubechat Product Requirements Document (PRD)

**Author:** pramoda
**Date:** 2025-10-30T08:05:58.188389Z
**Project Level:** 4
**Target Scale:** 5–50 clusters (50–1,000 nodes; 5k–30k pods) across cloud + on-prem

---

## Goals and Background Context

### Goals

- Reduce MTTR across P1/P2 incidents by 30–50%.
- Convert 70–90% of production clusters to guarded workflows with previews/approvals and RBAC enforcement.
- Cut audit preparation from days to under two hours via exportable, consolidated logs.
- Launch OSS MVP + Helm chart, onboard 10–20 design partners, reach 30+ external contributors, and exceed 200 weekly active operators.
- Enable an enterprise tier (Postgres, SSO, policy packs) by months 9–12 to unlock monetization.

### Background Context

- A 4–6 person SRE/Platform core team plus 5–10 on-call developers manages 24×7 Kubernetes operations entirely through manual `kubectl`, ad-hoc scripts, and siloed dashboards.
- Live changes lack standardized preview/approval gates; wrong-context or overly broad commands surface roughly 1–2 times per quarter despite informal peer checks.
- Weekly CLI mistakes and inconsistent scripting add 20–60 minutes to remediation, with cross-cluster triage consuming an additional 10–20 minutes and requiring multiple terminals.
- 60–80% of direct cluster activity lacks a concise, human-readable audit trail, making “who ran what, where, when, why” investigations slow and compliance-heavy.
- Process gaps translate into ~4–6 engineer-hours of rework each month and 1–3 avoidable customer-impacting incidents per year.

---

## Requirements

### Functional Requirements

- **FR-0 Kubechat Foundation**
  - Fork and extend the Kubechat (https://github.com/pramodksahoo/kubechat) codebase as the starting point for Kubechat.
  - Preserve upstream compatibility to consume future Kubechat improvements while layering Kubechat-specific capabilities.
  - Document divergences and merge strategy so engineering teams can track deltas across releases.

- **FR-1 Natural-Language Operations Console**
  - Accept natural-language intents and generate explicit, explainable kubectl/API plans with scoped cluster/namespace targeting.
  - Provide per-step risk annotations, blast-radius estimates, and editable parameters before execution.

- **FR-2 Guarded Execution Pipeline**
  - Enforce dry-run previews and require explicit confirmation (including optional two-person approvals) for destructive or production-scoped actions.
  - Perform RBAC checks prior to execution and surface actionable denial reasons.
  - Capture rollback context (previous replicas/images/manifests) and expose one-click revert paths.

- **FR-3 Multi-Cluster Visibility & Control**
  - Aggregate diagnostics (events, logs, metrics, pod status) across selected clusters with clear cluster badges and fast pivots.
  - Support safe scoped actions (describe, logs, scale, restart) across clusters with guardrails applied per target.

- **FR-4 Observability & Streaming Feedback**
  - Stream live command output, rollout status, and log/exec sessions inside the chat console with AI-authored summaries (e.g., suspected root cause, follow-up suggestions).
  - Correlate telemetry (recent deploys, alerts, config changes) into the incident thread.

- **FR-5 Audit & Reporting Layer**
  - Maintain tamper-evident audit logs capturing operator identity, command inputs, approvals, outputs, and context tags.
  - Provide exportable evidence packages (CSV/JSON) suitable for SOC 2/ISO/PCI/HIPAA reviews and link to generated incident notes.

- **FR-6 Packaging & Deployment Experience**
  - Deliver a single binary (embedding the React UI) for local use and a Helm chart with CRDs (AIProvider, AuditPolicy) for cluster deployment.
  - Provide offline-first AI provider abstraction: default to Ollama with configurable GPT/Claude/Gemini connectors.

- **FR-7 Workflow Governance & Runbooks**
  - Allow creation and reuse of approved command “recipes” or runbooks with guardrails.
  - Expose post-incident summaries and follow-up suggestions directly in the console.

### Non-Functional Requirements

- **Performance & Responsiveness**: Initial UI load <3 seconds on target hardware; API latency p95 <200 ms under standard load; command planning turnaround <5 seconds with local Ollama models.
- **Reliability**: Service availability ≥99.9%; graceful degradation when optional integrations (external LLMs, observability backends) are unreachable.
- **Scalability**: Support 5–50 clusters (5k–30k pods) with server-side pagination/filtering; accommodate growth to 1,000 nodes without manual tuning.
- **Security & Compliance**: Enforce RBAC, TLS, secret redaction, and tamper-evident audit logs; preserve data residency within the user’s cluster; operate in air-gapped/offline contexts by default.
- **Privacy**: Redact PII/secrets from prompts/logs; provide opt-in controls for any telemetry or cloud egress; default deny outbound network access.
- **Extensibility**: Enable pluggable AI providers and future policy packs without core rewrites; offer signed plugin surface for custom resource panels.
- **Operability**: Provide health endpoints, structured logs, metrics, and documentation for deployment/upgrade/rollback workflows.
- **Maintainability**: Preserve merge path with upstream Kubechat; enforce PR-only workflow with Codex-generated code reviewed by humans.

---

## User Journeys

1. **Incident Commander Executes Guarded Remediation**
   - Alert fires in PagerDuty → Commander opens Kubechat → Issues NL prompt “Investigate crashloop pods for payment in prod”.
   - System returns plan with diagnostics, proposed remedial commands, risk notes, and blast radius.
   - Commander reviews dry-run output, requests peer approval, executes command; live stream confirms rollout success and AI summarizes root cause.
   - Audit entry and post-incident note automatically generated with follow-up tasks.

2. **On-Call Developer Performs Guided Investigation**
   - Developer joins incident channel → filters affected clusters/namespaces via multi-cluster view.
   - Uses chat to request logs, metrics, and recent deploy history without leaving console.
   - Identifies misconfiguration, triggers pre-approved rollback recipe, and documents outcome via auto-generated summary.

3. **Compliance Lead Conducts Weekly Review**
   - Compliance user opens audit dashboard → filters privileged actions by cluster and operator.
  - Exports CSV evidence covering commands, approvals, and context for the week.
  - Flags any out-of-policy actions and assigns follow-up in ticketing system via webhook.

---

## UX Design Principles

- Prioritize explainability: every AI-generated action must show underlying commands, risks, and approvals.
- Guardrail-first interactions: defaults favor safety (read-only, dry-run) with explicit escalation for writes.
- Maintain shared situational awareness through synchronized chat + visual context with timeline markers.
- Optimize for speed under pressure: keyboard acceleration, quick cluster pivots, and minimal modal friction.
- Design for offline/air-gapped use, eliminating assumptions about external connectivity.
- Make compliance effortless with built-in transparency and exportable evidence.

---

## User Interface Design Goals

- Persistent cluster/context banner with warning badges when scope changes.
- Dual-pane layout: conversational thread alongside context-aware resource panels.
- Inline diff/previews for proposed mutations with clear accept/reject controls.
- Streaming log/output viewer with AI annotations and quick filters.
- Audit timeline widget highlighting who/what/when/how for recent actions.
- Responsive virtualized tables for large pod lists and event streams.

---

## Epic List

- **E1. AI-Assisted Command Planning** – Transform natural-language intents into explainable kubectl plans with risk analysis and editable parameters.
- **E2. Guarded Execution & Approvals** – Enforce dry-run previews, RBAC checks, and approval workflows before mutating operations run.
- **E3. Multi-Cluster Visibility Console** – Provide aggregated diagnostics, scoped actions, and synchronized UI/chat state across clusters.
- **E4. Streaming Observability & Insights** – Deliver live logs, exec sessions, rollout status, and AI-written summaries inside the chat experience.
- **E5. Audit, Compliance & Policy Framework** – Capture tamper-evident audit trails, export evidence, and manage guardrail policies.
- **E6. Packaging, Deployment & AI Provider Abstraction** – Ship single-binary and Helm packages with pluggable AI providers and baseline configuration tooling.

> **Note:** Detailed epic breakdown with full story specifications is available in [epics.md](./epics.md)

---

## Out of Scope

- Cross-cluster rollout orchestration (waves, canaries, advanced health gating).
- Topology/graph visualizations and deep dependency mapping.
- Cost/right-sizing insights, SLO/error-budget analytics, and advanced dashboards.
- Policy packs marketplace, plugin SDK, and deep runbook automation tooling.
- Enterprise SSO/SAML/OIDC, multi-step managerial approvals, and executive dashboards.
- Proactive AI recommendations beyond targeted hints; rich third-party integrations (Slack, PagerDuty, Jira) beyond initial webhooks.

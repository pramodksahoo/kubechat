# Epic Technical Specification: E1 – AI-Assisted Command Planning

Date: 2025-10-30
Author: pramoda
Epic ID: E1
Status: Draft

---

## Overview

Kubechat must transform operator intent expressed in natural language into a safe, reviewable kubectl execution plan. This epic delivers the planning pipeline that interprets prompts, scopes the target cluster/namespace, composes diagnostics, highlights risk, and produces auditable output the Guarded Execution epic can consume. ([Source: PRD.md §3 Functional Requirements])

## Objectives and Scope

**In Scope**
- Natural-language parsing and intent classification for Kubernetes operations.
- Plan assembly with explicit command steps, affected resources, risk annotations, and suggested diagnostics.
- Editing and preview support so operators can adjust parameters before approval.
- Streaming summary of execution outcome and auto-generated incident notes. ([Source: PRD.md §3, User Journeys])

**Out of Scope**
- Command execution safeguards (covered by Epic E2).
- Cross-cluster rollout orchestration (Epic E6 future phase).
- UI refinements beyond the command palette/timeline already defined in UX spec. ([Source: PRD.md §Out of Scope])

## System Architecture Alignment

- Reuses `internal/api` for prompt ingestion and plan persistence. ([Source: architecture.md §5])
- Delegates NL intent resolution to `internal/ai` adapters (Ollama default, cloud providers optional). ([Source: architecture-decisions.json: AI Providers])
- Persists Plan entities and risk metadata in PostgreSQL (`plans`, `audit_log`). ([Source: architecture.md §9 Data Model])
- Streams plan updates through `internal/streaming` SSE endpoints consumed by the React timeline. ([Source: architecture.md §8 Data Flow])

## Detailed Design

### Services and Modules

| Component | Location | Responsibility | Notes |
| --- | --- | --- | --- |
| PromptController | `internal/api/prompts` | REST endpoints (`POST /api/v1/prompts`, `GET /api/v1/plans/{id}`) and validation | Enforces JSON schema, guards against unbounded prompt size |
| PlanBuilder | `internal/ai` | Interfaces with provider adapters, normalizes plan structure | Uses provider registry with rate limits, redaction hooks ([architecture-decisions.json]) |
| RiskAssessor | `internal/ai/risk` | Computes blast radius, dependency graph, policy interactions | Consumes RBAC metadata and architecture constraints |
| DiagnosticsAdvisor | `internal/ai/diagnostics` | Suggests read-only diagnostics before mutating steps | Pulls templates from configuration and past incidents |
| PlanRepository | `internal/api/repository/plan_repo.go` | CRUD for `plans` table, links to audit entries | Ensures append-only history, retains JSON body |
| PlanStreamHub | `internal/streaming/plan_hub.go` | SSE topics `plan_update`, `execution_status` | Backs off on slow consumers, logs analytics |

### Data Models and Contracts

**Plan** (`plans` table)
- `id` UUID
- `prompt` TEXT (original NL input)
- `plan_json` JSONB (array of `PlanStep`)
- `summary` TEXT (AI post-execution synopsis)
- `risk_level` ENUM(`low`,`medium`,`high`)
- `created_by` UUID (operator)
- `created_at` TIMESTAMP

**PlanStep** (stored in `plan_json`)
- `sequence` INT
- `command` STRING (kubectl/CLI statement)
- `operation_type` ENUM (`diagnostic`,`mutating`)
- `target` STRUCT { cluster, namespace, resource }
- `dry_run_available` BOOL
- `diff_preview` JSON (server-side apply diff snippet)
- `risks` ARRAY { code, description, severity }

**DiagnosticsSuggestion**
- `name` STRING
- `action` STRING (e.g., `kubectl describe pod ...`)
- `rationale` STRING

([Source: architecture.md §9])

### APIs and Interfaces

| Endpoint | Method | Description | Request / Response |
| --- | --- | --- | --- |
| `/api/v1/prompts` | POST | Submit NL prompt and receive plan ID | **Request**: `{ prompt, scope?, options }` · **Response**: `{ planId }` |
| `/api/v1/plans/{id}` | GET | Fetch plan with steps, diagnostics, risk | Returns Plan JSON with steps |
| `/api/v1/stream/plans/{id}` | SSE | Stream `plan_update`, `execution_status`, `summary` events | Event payloads include timestamps, actor, diff |
| Provider interface | Go interface `Provider.GeneratePlan(ctx, Prompt) (Plan, error)` | Adapter contract for Ollama/GPT/Claude | Enforces redaction before logging ([architecture-decisions.json]) |

### Workflows and Sequencing

1. API receives prompt, validates scope, writes seed audit entry.
2. PlanBuilder sends normalized prompt to provider adapter (falls back to templates if provider unavailable).
3. RiskAssessor enriches steps with blast radius (namespaces, resource counts) using `internal/k8s` discovery. ([Source: architecture.md §8])
4. DiagnosticsAdvisor adds read-only actions before mutating steps.
5. PlanRepository stores plan JSON; PlanStreamHub emits `plan_update` over SSE.
6. Operator edits plan parameters (cluster, namespace, limits). Edits persisted via PATCH to `/api/v1/plans/{id}`.
7. On execution completion (Epic E2), summary is appended, archived in audit log, and streamed via SSE.

## Non-Functional Requirements

### Performance
- Plan generation median < 5s; p95 < 10s with Ollama. ([Source: PRD.md §NFR])
- SSE event delivery within 1s of plan update; clients tolerate 30s reconnect.

### Security
- Inputs sanitized to prevent prompt-injection leaking secrets. ([architecture-decisions.json: Security])
- Enforce RBAC pre-check for suggested command tags; risk level high if policy gaps.
- Audit log records prompt, plan, edits with hash chain.

### Reliability/Availability
- Provider adapter retries (max 2) with exponential backoff; degrade with "plan unavailable" gracefully. ([architecture-decisions.json: AI Providers])
- Persist plan steps before emitting SSE to handle client reconnects.

### Observability
- Metrics: `plan_generation_duration`, `plan_risk_level_total`, `plan_edit_count`. ([Source: architecture.md §12])
- Structured logs include `planId`, `riskLevel`, `provider`.
- Trace spans for provider calls and risk assessment.

## Dependencies and Integrations

- AI providers: Ollama (local), optional GPT/Claude/Gemini connectors via configuration. ([architecture-decisions.json])
- PostgreSQL 15 for plan storage; migrations tracked in backend.
- SSE/WebSocket infrastructure in `internal/streaming`.
- RBAC/policy metadata supplied by Guarded Execution epic components.

## Acceptance Criteria (Authoritative)

1. System converts NL prompt into ordered plan steps with command, type, target, and diff preview. ([Source: PRD.md §FR-1])
2. Plan includes risk annotations (severity, affected namespaces/resources) and suggested diagnostics before mutating steps. ([Source: PRD.md §FR-1, User Journeys])
3. Operators can edit plan parameters prior to approval and persisted edits survive page reload. ([Source: PRD.md §Goals])
4. Plan update events broadcast over SSE for UI timeline consumption. ([Source: UX Design Specification §Workflows])
5. Execution summary appended post-run and available via API/stream for incident note generation. ([Source: PRD.md §User Journey A])

## Traceability Mapping

| AC | PRD / UX Source | Architecture Component | Test Idea |
| --- | --- | --- | --- |
| 1 | PRD §FR-1, UX Command Palette | `PromptController`, `PlanBuilder` | Unit tests for provider adapters; API contract tests for POST `/api/v1/prompts` |
| 2 | PRD §FR-1, Architecture §9 | `RiskAssessor`, `DiagnosticsAdvisor` | Integration tests with seeded cluster metadata verifying risk levels |
| 3 | PRD §Goals, UX §Command Editing | `PlanRepository`, React Editor | E2E test editing plan and verifying persistence via GET |
| 4 | UX §Timeline, Architecture §8 | `PlanStreamHub`, SSE endpoint | Web test ensuring SSE emits `plan_update` on plan creation |
| 5 | PRD §User Journey A | `PlanRepository`, `AuditService`, `PlanStreamHub` | E2E test verifying summary appended after simulated execution |

## Risks, Assumptions, Open Questions

- **Risk:** Provider latency spikes causing >10s plan builds → Mitigate with circuit breaker and fallback templates.
- **Risk:** Risk assessment inaccurate without up-to-date cluster metadata → Ensure sync job runs before plan generation.
- **Assumption:** Policy metadata supplied by Epic E2 in time for integration.
- **Open Question:** Need confirmation on tenant-specific AI rate limits.

## Test Strategy Summary

- Unit tests: `internal/ai` adapters (Ollama, GPT), `RiskAssessor` heuristics, `PlanRepository` persistence.
- Contract/API tests: POST `/api/v1/prompts`, GET `/api/v1/plans/{id}`, SSE stream contract.
- Integration tests: plan generation with seeded cluster state, verifying diff preview and diagnostics.
- End-to-end tests: Playwright flows covering command prompt, plan editing, SSE updates (reusing fixture architecture and network-first patterns).
- Performance tests: Load plan generation with representative prompts to validate <10s p95.

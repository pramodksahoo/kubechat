# KubeChat

**Natural Language Kubernetes Management Platform**

> Transform complex kubectl operations into intuitive conversations with your cluster while maintaining enterprise-grade security and compliance.

![KubeChat Banner](https://img.shields.io/badge/KubeChat-Natural_Language_K8s_Management-blue?style=for-the-badge&logo=kubernetes)

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18+-61DAFB.svg?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.6+-3178C6.svg?logo=typescript)](https://www.typescriptlang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28+-blue.svg)](https://kubernetes.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg?logo=docker)](https://www.docker.com/)
[![Helm](https://img.shields.io/badge/Helm-3.15+-0F1689.svg?logo=helm)](https://helm.sh/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-336791.svg?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7.4+-DC382D.svg?logo=redis)](https://redis.io/)
[![Ollama](https://img.shields.io/badge/Ollama-Local_AI-000000.svg)](https://ollama.ai/)
[![OpenAI](https://img.shields.io/badge/OpenAI-Optional-412991.svg?logo=openai)](https://openai.com/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-3.4+-38B2AC.svg?logo=tailwind-css)](https://tailwindcss.com/)

## 🌟 What is KubeChat?

KubeChat is an **open-source Natural Language Kubernetes Management Platform** that bridges the gap between Kubernetes complexity and operational efficiency. Instead of memorizing kubectl commands, simply chat with your cluster:

```
"Show me all pods with high memory usage in production"
"Scale the payment-service deployment to handle weekend traffic" 
"What's causing the API gateway to be unhealthy?"
```

**🎯 Perfect for:** DevOps teams, SREs, and platform engineers who need simplified Kubernetes management without compromising security or compliance standards.

## 🚀 Why KubeChat?

### ✨ **Natural Language First**
- Chat with your cluster using plain English
- No more memorizing complex kubectl syntax
- Contextual suggestions and intelligent command translation

### 🔒 **Enterprise-Grade Security**
- **Air-gapped deployment** with Ollama (runs completely offline)
- Built-in RBAC integration and audit logging
- Zero external API calls required for AI processing
- Complete compliance-ready audit trails

### 🎛️ **Modern Web Interface**
- Beautiful, responsive dashboard with real-time cluster monitoring
- Multi-user collaborative troubleshooting
- Progressive disclosure for beginners to experts
- WebSocket-powered live updates

### 🧠 **Multi-LLM Support**
- **Ollama** (default) - Local, air-gapped AI processing
- **OpenAI** - Cloud-powered enhanced capabilities  
- **Intelligent fallback** between providers
- No vendor lock-in

### 🔧 **Production-Ready**
- Kubernetes-native deployment with Helm charts
- Supports all major K8s distributions (EKS, GKE, AKS, OpenShift, Rancher)
- Horizontal scaling and high availability
- Built-in monitoring and observability

## 🏗️ Architecture


---

## ✨ Current Capabilities (MVP)

| Area | Status | Description |
|------|--------|-------------|
| Natural-language prompt ingestion | ✅ | `POST /api/v1/prompts` accepts user intents, classifies cluster/namespace scope, and emits structured plans with per-step metadata. |
| Plan builder heuristics | ✅ | `internal/plan` enriches steps with operation types, target descriptors, risk notes, and affected resources derived from prompt context. |
| Plan persistence | ✅ | Plans are cached via an in-memory repository so they can be reloaded (`GET /api/v1/plans/{id}`) and rehydrated in the UI. |
| Metrics & logging | ✅ | `plan_generation_duration_seconds` Prometheus histogram plus structured logs (request ID, cluster, namespace, risk level). |
| Plan preview UI | ✅ | React-based drawer showing plan header, namespace/cluster chips, affected-resource summaries, and copyable step commands. |
| Chat integration | ✅ | The chat workflow triggers plan creation automatically and deep-links the preview drawer using `?plan=<id>`. |

> These features satisfy Acceptance Criteria 1 and partially satisfy Acceptance Criteria 2 of story `1-1-capture-natural-language-intents`.

---

## 🗺️ Roadmap Alignment (from PRD / Epics)

| Epic / FR | Description | Status |
|-----------|-------------|--------|
| **E1 / FR-1** AI-Assisted Command Planning | Natural-language intents → explainable kubectl plan | 🚧 In progress (MVP delivered, editing/approval workflows forthcoming) |
| **E2 / FR-2** Guarded Execution & Approvals | Dry-run enforcement, approvals, RBAC, rollback | ⏳ Planned |
| **E3 / FR-3** Multi-Cluster Visibility | Aggregated diagnostics, scoped actions | ⏳ Planned |
| **E4 / FR-4** Streaming Observability | Live logs, rollout status, AI summaries | ⏳ Planned |
| **E5 / FR-5** Audit & Reporting | Tamper-evident audit trails, exports | ⏳ Planned |
| **E6 / FR-6** Packaging & AI Abstraction | Single binary, Helm chart, provider plugins | ⏳ Planned |
| **FR-7** Workflow Governance | Guardrailed recipes, post-incident notes | ⏳ Planned |

For full requirements and design principles, see:
- [Product Requirements Document](docs/PRD.md)
- [Epic definitions](docs/epics.md)
- [Architecture overview](docs/architecture.md)

---

## 🧱 Architecture Snapshot

```
┌────────────────────────┐   REST + SSE    ┌────────────────────────┐
│ React UI (Plan Drawer) │ ◄──────────────►│ Go API (internal/api)   │
└────────────────────────┘                │  - PromptController     │
                                          │  - PlanRepository       │
                                          │  - Telemetry exposure   │
                                          └────────────┬───────────┘
                                                       ▼
                                            ┌────────────────────┐
                                            │ Plan Builder       │
                                            │ (internal/plan)    │
                                            └────────────────────┘
```

- **Backend:** Go 1.24 module (`backend/`) with Echo router, SSE server, and Prometheus metrics.
- **Frontend:** React + TypeScript + Tailwind drawer rendered within the shell layout (`client/`).
- **Telemetrics:** Histogram and structured logging keyed by `request_id`.
- **Persistence:** In-memory repository (future stories will wire PostgreSQL per PRD).

---

## 🛠️ Local Development

```bash
# Backend
cd backend
Go env GOWORK=off go mod tidy
Go env GOWORK=off go test ./internal/...

# Frontend
cd ../client
pnpm install
pnpm run lint   # Note: legacy warnings remain; see eslint output for backlog items
pnpm run dev    # Starts Vite dev server with the React plan drawer
```

> ⚠️ Running `go test ./...` in restricted environments may fail when the MCP handler attempts to bind to localhost. Use `go test ./internal/...` for targeted coverage.

---

## 📁 Repository Structure

```
backend/                      # Go services (API, plan builder, telemetry)
client/                       # React application (plan drawer, chat integration)
third_party/github.com/kubechat/sse/v2
                              # Vendored SSE fork used via go.mod replace
charts/                       # Helm assets (placeholder, upcoming stories)
docs/                         # PRD, epics, architecture, and sprint docs
```

---

## 🤝 Contributing

1. Fork the repository at [github.com/pramodksahoo/kubechat](https://github.com/pramodksahoo/kubechat)
2. Create a feature branch: `git checkout -b feature/plan-editor`
3. Run Go and frontend unit tests relevant to your change
4. Submit a pull request referencing the story / acceptance criteria

Please review the [code of conduct](CODE_OF_CONDUCT.md) before contributing.

---

## 📣 Support & Feedback

- Issues & feature requests: [GitHub Issues](https://github.com/pramodksahoo/kubechat/issues)
- Discussions & design questions: [GitHub Discussions](https://github.com/pramodksahoo/kubechat/discussions)
- Documentation hub: [`docs/`](docs/)

---

## ⭐ Project Vision

Kubechat aims to make Kubernetes safer and more approachable by combining conversational interfaces with governed execution. The current milestone proves out plan generation; upcoming stories will layer in approvals, audit trails, and packaging aligned with the PRD roadmap.

> Built with ❤️ for platform engineers and SREs who want confidence before they hit “apply”.

# Kubechat Technical Decisions Log

**Last Updated:** 2025-10-30T08:36:00.772324Z
**Author:** PM Team

---

## Architecture Foundations

- Build atop Kubechat while preserving upstream merge compatibility; maintain PR-only workflow with human review.
- Enforce RBAC, TLS, and secret redaction across services; store tamper-evident audit logs within the customer cluster.
- Disable outbound egress by default to support air-gapped deployments; require explicit policy to enable external integrations.
- Support Kubernetes N-3 minor versions using server-side pagination/filtering; avoid alpha APIs.

## Platform & Packaging Choices

- Backend: Go 1.23 with client-go; surfaces REST, WebSocket/SSE interfaces for UI and streaming.
- Frontend: React + Vite with virtualized data grids and synchronized chat/visual context.
- Storage: SQLite for local/dev footprints; PostgreSQL for production clusters.
- Distribution: Single binary bundling frontend assets for local use; Helm chart with AIProvider and AuditPolicy CRDs for cluster installs.
- CI/CD: Automated tests, security scans, SBOM generation, and signed multi-arch container images per release.

## AI & Guardrail Strategy

- Default AI provider: Ollama running locally or in-cluster for offline capability.
- Pluggable provider abstraction to support GPT, Claude, Gemini; enforce prompt/response redaction before external calls.
- Risk scoring and blast-radius analysis embedded in command planning; tie approval workflows to policy packs.

## Observability & Compliance

- Integrate Prometheus/Grafana/Loki/ELK signals into the chat timeline for correlated insights.
- Provide exportable audit evidence (CSV/JSON) with cryptographic hashes for integrity verification.
- Maintain rollback metadata (replicas, image digests, manifests) alongside mutation history to support one-click reverts.

## Performance Targets

- UI first paint <3s on supported hardware; AI plan generation <5s using local models.
- Service availability ≥99.9% with graceful degradation when optional integrations fail.
- Support 5–50 clusters (50–1,000 nodes; 5k–30k pods) with roadmap to scale higher via pagination and caching.

---

_This log captures cross-functional decisions surfaced during PRD planning; update as architecture and implementation evolve._

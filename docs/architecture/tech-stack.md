# KubeChat Technology Stack

## Overview

This document defines the definitive technology choices for KubeChat development. All development must use these exact versions and tools.

**Last Updated:** 2025-01-10  
**Repository:** https://github.com/pramodksahoo/kubechat  
**Branch:** develop

---

## Core Technologies

### Frontend Stack

| Technology | Version | Purpose | Installation |
|------------|---------|---------|--------------|
| **Node.js** | 20+ | Runtime for frontend tooling | `fnm install 20 && fnm use 20` |
| **TypeScript** | 5.6+ | Type-safe frontend development | `pnpm add -D typescript@^5.6.0` |
| **React** | 18.3+ | Component-based UI framework | `pnpm add react@^18.3.0` |
| **Next.js** | 14+ | Full-stack React framework | `pnpm add next@^14.0.0` |
| **Tailwind CSS** | 3.4+ | Utility-first CSS framework | `pnpm add -D tailwindcss@^3.4.0` |
| **Headless UI** | 2.1+ | Accessible UI components | `pnpm add @headlessui/react@^2.1.0` |
| **Zustand** | 4.5+ | Lightweight state management | `pnpm add zustand@^4.5.0` |
| **Vite** | 5.4+ | Build tool and dev server | `pnpm add -D vite@^5.4.0` |

### Backend Stack

| Technology | Version | Purpose | Installation |
|------------|---------|---------|--------------|
| **Go** | 1.23+ | Backend language | Download from https://go.dev/dl/ |
| **Gin** | 1.10+ | HTTP web framework | `go get github.com/gin-gonic/gin@v1.10.0` |
| **client-go** | 0.30+ | Kubernetes API client | `go get k8s.io/client-go@v0.30.0` |
| **PostgreSQL** | 16+ | Primary database | Via Helm chart (see deployment) |
| **Redis** | 7.4+ | Caching and sessions | Via Helm chart (see deployment) |

### AI/ML Stack

| Technology | Version | Purpose | Configuration |
|------------|---------|---------|---------------|
| **Ollama** | 0.3+ | Local AI inference | Deployed via Helm chart |
| **Phi-3.5-mini** | Latest | Primary Kubernetes model | Auto-downloaded in container |
| **CodeLlama** | 34B | Code generation model | Optional, for advanced queries |
| **OpenAI SDK** | 5.0+ | Cloud AI fallback | `go get github.com/sashabaranov/go-openai@v1.28.0` |

### Development Tools

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **PNPM** | 8+ | Package manager | `curl -fsSL https://get.pnpm.io/install.sh \| sh` |
| **Docker** | Latest | Container runtime | Included with Rancher Desktop |
| **Kubernetes** | 1.28+ | Container orchestration | Rancher Desktop |
| **Helm** | 3.15+ | Kubernetes package manager | `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \| bash` |
| **kubectl** | 1.28+ | Kubernetes CLI | Included with Rancher Desktop |

### Testing Stack

| Tool | Version | Purpose | Usage |
|------|---------|---------|-------|
| **Vitest** | 2.0+ | Frontend unit testing | `pnpm add -D vitest@^2.0.0` |
| **Testing Library** | 16+ | React component testing | `pnpm add -D @testing-library/react@^16.0.0` |
| **Go testing** | Built-in | Backend unit testing | Native Go testing package |
| **Testify** | 1.9+ | Go assertion library | `go get github.com/stretchr/testify@v1.9.0` |
| **Playwright** | 1.47+ | E2E testing | `pnpm add -D @playwright/test@^1.47.0` |

---

## Architecture Decisions

### Why These Choices?

#### **Go 1.23+ Backend**
- **Native Kubernetes Integration:** client-go provides optimal performance
- **Concurrency:** Goroutines handle multiple AI requests efficiently  
- **Enterprise Reliability:** Mature ecosystem with excellent tooling
- **Memory Efficiency:** Critical for air-gapped deployments

#### **React 18.3+ with TypeScript**
- **Type Safety:** Prevents runtime errors, essential for enterprise
- **Component Reusability:** Atomic design supports consistent UI
- **Concurrent Features:** Improved performance for real-time updates
- **Enterprise Ecosystem:** Extensive library support

#### **Ollama-First AI Strategy**
- **Air-Gapped Capability:** Unique competitive advantage
- **Data Sovereignty:** No external API dependencies
- **Cost Control:** No per-token pricing for high usage
- **Kubernetes Optimization:** Models optimized for command generation

#### **Zustand State Management**
- **Simplicity:** Less boilerplate than Redux
- **TypeScript Integration:** Excellent type inference
- **Bundle Size:** Minimal impact on application size
- **Learning Curve:** Easy adoption for team members

#### **Helm-Native Deployment**
- **Kubernetes Standard:** Industry best practice
- **Air-Gapped Support:** Bundle entire application stack
- **Version Management:** Controlled rollouts and rollbacks
- **Configuration Management:** Environment-specific values

---

## Development Environment

### Container-First Requirements

**❌ Prohibited:**
- `pnpm run dev` - No local frontend processes
- `go run` - No local backend processes  
- Direct database connections from host
- Local AI model execution

**✅ Required:**
- All development through containers in Kubernetes
- Port-forwarding for service access
- Container rebuilds for code changes
- Helm charts for all dependencies

### Repository Structure

```plaintext
kubechat/
├── apps/
│   ├── web/           # React TypeScript frontend
│   └── api/           # Go backend services
├── packages/
│   ├── shared/        # Shared TypeScript types
│   ├── ui/           # Component library
│   └── config/       # Shared configuration
├── infrastructure/
│   ├── helm/         # Kubernetes deployment
│   ├── docker/       # Container configurations
│   └── scripts/      # Build automation
└── docs/
    └── architecture/  # Technical documentation
```

---

## Version Compatibility

### Kubernetes Compatibility

| KubeChat Version | Kubernetes Versions | Go client-go | Notes |
|------------------|-------------------|--------------|-------|
| 1.0.x | 1.28, 1.29, 1.30 | v0.30.x | Current development target |
| Future 1.1.x | 1.29, 1.30, 1.31 | v0.31.x | Planned upgrade path |

### Node.js Compatibility

| Frontend Tool | Node.js Versions | Notes |
|---------------|------------------|-------|
| Next.js 14.x | 18.17+, 20+ | Recommended: Node 20 LTS |
| Vite 5.x | 18+, 20+ | Optimal performance with Node 20 |
| PNPM 8.x | 16.14+, 18+, 20+ | Latest features require Node 18+ |

### Database Compatibility

| Component | PostgreSQL | Redis | Notes |
|-----------|------------|-------|-------|
| Audit Schema | 15+, 16+ | 7.0+ | JSONB features require 15+ |
| Performance | 16+ preferred | 7.4+ preferred | Latest versions for optimization |
| Air-gapped | 15.5+, 16+ | 7.2+ | Tested in offline environments |

---

## Security Requirements

### Dependency Management

```bash
# Audit frontend dependencies
pnpm audit

# Audit Go dependencies  
go mod tidy && go list -m -u -mod=readonly all

# Update dependencies (controlled process)
pnpm update --latest
go get -u ./...
```

### Container Security

```dockerfile
# Use specific versions, never 'latest'
FROM node:20.10-alpine AS frontend
FROM golang:1.23.2-alpine AS backend

# Non-root user in all containers
RUN adduser -D -s /bin/sh kubechat
USER kubechat
```

### Vulnerability Scanning

```yaml
# GitHub Actions security scanning
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    scan-ref: '.'
```

---

## Performance Requirements

### Build Performance

| Metric | Target | Measurement |
|--------|--------|-------------|
| Frontend Build | < 30 seconds | `pnpm build` in CI |
| Backend Build | < 15 seconds | `go build` in container |
| Container Build | < 2 minutes | Full Docker build |
| Helm Deploy | < 1 minute | Local Kubernetes deploy |

### Runtime Performance

| Component | Target | Notes |
|-----------|--------|-------|
| API Response | < 200ms P95 | Excluding AI processing |
| AI Processing | < 3s P95 | Ollama local inference |
| Frontend Load | < 2s | Initial page load |
| Bundle Size | < 500KB | Initial JavaScript bundle |

---

## Migration Strategy

### Upgrading Dependencies

#### Frontend Dependencies
```bash
# Check outdated packages
pnpm outdated

# Upgrade incrementally
pnpm update @types/react@latest
pnpm update tailwindcss@latest

# Test after each major upgrade
pnpm test && pnpm build
```

#### Backend Dependencies  
```bash
# Check for updates
go list -u -m all

# Upgrade Go version
go mod edit -go=1.23

# Upgrade specific dependencies
go get -u github.com/gin-gonic/gin@latest
go get -u k8s.io/client-go@latest
```

#### Database Migrations
```bash
# Test migration with backup
kubectl exec -it postgres-0 -- pg_dump kubechat > backup.sql

# Apply migrations
make dev-db-migrate

# Rollback if needed
kubectl exec -it postgres-0 -- psql kubechat < backup.sql
```

---

## Troubleshooting

### Common Version Issues

**Problem:** TypeScript errors after React upgrade
```bash
# Solution: Update @types packages
pnpm update @types/react@latest @types/react-dom@latest
```

**Problem:** Go build fails after client-go upgrade
```bash
# Solution: Update all Kubernetes dependencies together
go get -u k8s.io/api@v0.30.0 k8s.io/apimachinery@v0.30.0 k8s.io/client-go@v0.30.0
```

**Problem:** Container build fails with dependency conflicts
```bash
# Solution: Clear build cache and rebuild
docker system prune -f
make dev-rebuild-api
```

### Version Verification

```bash
# Verify all tool versions
node --version    # Should show v20.x.x
go version       # Should show go1.23.x
kubectl version  # Should show client v1.28+
helm version     # Should show v3.15+
```

---

## Next Steps

1. **Validate Environment:** Run `make dev-info` to verify all prerequisites
2. **Deploy Development:** Execute `make dev-deploy` for full container environment  
3. **Verify Versions:** Check all tools match the specifications above
4. **Start Development:** Begin with Epic 1 implementation priorities

For questions about technology choices or upgrade paths, consult the architecture team or create an issue in the repository.

---

*📚 Generated with [Claude Code](https://claude.ai/code)*

*Co-Authored-By: Winston (Architect) <architect@kubechat.dev>*
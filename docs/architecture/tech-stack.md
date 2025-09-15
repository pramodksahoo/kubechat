# KubeChat Technology Stack

This document outlines the complete technology stack used in KubeChat development and deployment.

## Required Technology Versions

### Frontend Technologies
- **Node.js:** 20+ for frontend tooling and build processes
- **TypeScript:** 5.6+ for type-safe development across all packages
- **React:** 18.3+ with modern hooks and concurrent features
- **Next.js:** 14+ for server-side rendering and routing
- **Tailwind CSS:** 3.4+ for utility-first styling
- **Headless UI:** 2.1+ for accessible, unstyled UI components

### Backend Technologies  
- **Go:** 1.23+ for backend services and APIs
- **Gin:** 1.10+ HTTP web framework for REST APIs
- **client-go:** v0.30+ for Kubernetes integration and cluster management

### Database & Storage
- **PostgreSQL:** 16+ primary relational database
- **Redis:** 7.4+ for caching, sessions, and real-time features

### Container & Orchestration
- **Docker:** Latest for containerization
- **Kubernetes:** 1.28+ for container orchestration
- **Helm:** 3.15+ for Kubernetes package management
- **kubectl:** 1.28+ Kubernetes command-line interface
- **Rancher Desktop:** Recommended local Kubernetes development environment

### Development Tools
- **PNPM:** 8+ for efficient package management in monorepo
- **Vitest:** 2.0+ for frontend testing framework
- **Testing Library:** 16+ for React component testing
- **Playwright:** 1.47+ for end-to-end browser testing
- **Testify:** 1.9+ for Go test assertions and mocking

### Build & Development
- **Air:** 1.49+ for Go hot-reload development
- **ESLint:** 8+ for JavaScript/TypeScript linting
- **Prettier:** Latest for code formatting
- **TypeScript:** 5.6+ compiler and type checking

## Development Architecture

### Container-First Development Rules

**❌ Prohibited Commands:**
```bash
pnpm run dev          # No local frontend development server
go run ./cmd/server   # No local backend processes  
npm start            # No direct Node.js processes
yarn dev             # No local development servers
```

**✅ Required Development Workflow:**
```bash
make dev-build        # Build containers with latest code
make dev-deploy       # Deploy to Kubernetes cluster
make dev-logs         # View application logs
make dev-shell-api    # Access backend container
make dev-shell-web    # Access frontend container
```

### Deployment Stack

**Local Development:**
- **Rancher Desktop** with Kubernetes enabled
- **NodePort** services for local access
- **local-path** storage class for persistent volumes
- **Development images** with hot-reload capabilities

**Production:**
- **LoadBalancer** or **Ingress** for external access
- **Persistent volumes** for database storage
- **Resource limits** and **health checks**
- **Production-optimized** container images

## Package Management

### Monorepo Structure
```yaml
pnpm-workspace.yaml:
  packages:
    - "apps/*"      # Main applications
    - "packages/*"  # Shared packages
```

### Dependency Management
- **Workspace dependencies** for shared packages
- **Semantic versioning** for all packages
- **Locked versions** in package.json files
- **No catalog references** in package files

## Testing Framework

### Frontend Testing
- **Vitest** for unit and integration tests
- **Testing Library** for React component testing  
- **jsdom** environment for browser simulation
- **MSW** for API mocking

### Backend Testing
- **Go testing** package for unit tests
- **Testify** for assertions and test suites
- **httptest** for HTTP handler testing
- **Test containers** for integration testing

### End-to-End Testing
- **Playwright** for browser automation
- **Multi-browser** testing (Chrome, Firefox, Safari)
- **Container-based** test execution
- **Screenshot** and **video** capture on failures

## Security & Quality

### Container Security
- **Non-root** user in all containers
- **Specific versions** (never 'latest') in Dockerfiles
- **Multi-stage builds** for minimal production images
- **Vulnerability scanning** for dependencies

### Code Quality
- **TypeScript strict mode** for type safety
- **ESLint rules** for code consistency
- **Go formatting** with gofmt and goimports
- **Pre-commit hooks** for quality gates

### Authentication & Authorization
- **JWT tokens** for session management
- **RBAC integration** with Kubernetes
- **Secure headers** and CORS configuration
- **Input validation** and sanitization

## Development Environment

### Required Tools Installation

| Tool | Version | Installation Command |
|------|---------|---------------------|
| **Docker** | Latest | Included with Rancher Desktop |
| **Kubernetes** | 1.28+ | Rancher Desktop |  
| **Helm** | 3.15+ | `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \| bash` |
| **kubectl** | 1.28+ | Included with Rancher Desktop |
| **PNPM** | 8+ | `curl -fsSL https://get.pnpm.io/install.sh \| sh` |
| **Go** | 1.23+ | Download from https://go.dev/dl/ |

### Development Workflow

1. **Code Changes:** Edit source files in your IDE
2. **Container Build:** `make dev-rebuild-api` or `make dev-rebuild-web`  
3. **Deploy Changes:** `helm upgrade kubechat-dev infrastructure/helm/kubechat --namespace kubechat --values infrastructure/helm/kubechat/values-dev.yaml`
4. **Verify Deployment:** `kubectl get pods -n kubechat`
5. **Test Changes:** Access services via NodePort URLs

## Performance Considerations

### Frontend Optimization
- **Tree shaking** for unused code elimination
- **Code splitting** for lazy loading
- **Image optimization** with Next.js
- **Bundle analysis** and size monitoring

### Backend Optimization  
- **Connection pooling** for database access
- **Redis caching** for frequently accessed data
- **Go routines** for concurrent processing
- **Profiling tools** for performance monitoring

### Infrastructure Optimization
- **Resource limits** for containers
- **Horizontal Pod Autoscaling** for traffic spikes
- **Persistent volume** optimization
- **Network policies** for security

This technology stack ensures KubeChat maintains high performance, security, and developer productivity while following modern cloud-native development practices.

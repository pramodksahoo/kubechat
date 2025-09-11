# KubeChat Development Guide

> **Container-First Development Environment**  
> Complete guide for developing KubeChat using container-first principles with Rancher Desktop.

## 🎯 Overview

KubeChat follows a **container-first development approach** where all development happens inside containers deployed to Kubernetes. This ensures:

- **Production Parity**: Development environment matches production exactly
- **Consistency**: All developers work in identical environments
- **Isolation**: No conflicts with local system dependencies
- **Air-Gap Ready**: Supports offline development scenarios

## 🚀 Quick Start

### Prerequisites

Ensure you have these tools installed:

| Tool | Version | Installation |
|------|---------|--------------|
| **Docker** | Latest | [Rancher Desktop](https://rancherdesktop.io/) (Recommended) |
| **Kubernetes** | 1.28+ | Included with Rancher Desktop |
| **Helm** | 3.15+ | `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \| bash` |
| **kubectl** | 1.28+ | Included with Rancher Desktop |
| **Node.js** | 20+ | [nodejs.org](https://nodejs.org/) |
| **PNPM** | 8+ | `curl -fsSL https://get.pnpm.io/install.sh \| sh` |
| **Go** | 1.23+ | [go.dev](https://go.dev/dl/) |

### 1️⃣ Environment Validation

```bash
# Clone the repository
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Validate your development environment
./infrastructure/scripts/validate-env.sh

# If validation passes, you're ready to proceed!
```

### 2️⃣ One-Command Setup

```bash
# Initialize and deploy KubeChat (first time)
make init

# This will:
# ✅ Validate prerequisites
# ✅ Build containers  
# ✅ Deploy to Kubernetes
# ✅ Set up databases
# ✅ Start all services
```

### 3️⃣ Access KubeChat

Once deployed, access KubeChat at:

| Service | URL | Description |
|---------|-----|-------------|
| **Frontend** | http://localhost:30001 | Main KubeChat interface |
| **API** | http://localhost:30080 | Backend API endpoints |
| **PgAdmin** | http://localhost:30050 | Database management |
| **Redis Commander** | http://localhost:30081 | Cache management |

**Default Login:**
- Username: `admin`
- Password: `password`

## 🛠️ Development Workflow

### Core Commands

```bash
# Environment Management
make dev-info          # Show system status and prerequisites
make dev-setup         # Initialize development environment
make dev-deploy        # Deploy complete stack to Kubernetes
make dev-clean         # Clean up development environment

# Development Tools
make dev-logs          # View application logs
make dev-shell-api     # Shell into API container
make dev-shell-web     # Shell into web container  
make dev-port-forward  # Setup port forwarding

# Container Management
make dev-build         # Build all containers
make dev-rebuild-api   # Rebuild API container only
make dev-rebuild-web   # Rebuild web container only

# Database Management
make dev-db-connect    # Connect to development database
make dev-db-migrate    # Run database migrations
make dev-db-seed       # Seed development data

# Testing
make dev-test          # Run all tests
make dev-test-unit     # Run unit tests only
make dev-test-e2e      # Run end-to-end tests
```

### Container-First Development Cycle

**Important**: KubeChat uses container-first development - all code changes require rebuilding containers and redeploying to Kubernetes.

```bash
# 1. Make code changes in your editor
# 2. Rebuild specific container
make dev-rebuild-api   # For backend changes
make dev-rebuild-web   # For frontend changes

# 3. Redeploy to Kubernetes (REQUIRED for changes to take effect)
helm upgrade kubechat-dev infrastructure/helm/kubechat \
  --namespace kubechat \
  --values infrastructure/helm/kubechat/values-dev.yaml

# 4. Verify deployment
kubectl get pods -n kubechat
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=kubechat-dev -n kubechat

# 5. View logs
make dev-logs

# 6. Test changes
curl http://localhost:30080/health  # API health check
curl http://localhost:30001         # Frontend check
```

**For Any Story Development:**

Whenever working on stories or features, follow this process:

1. **Code Changes**: Edit source files
2. **Container Rebuild**: `make dev-rebuild-api` or `make dev-rebuild-web`
3. **Helm Upgrade**: Deploy changes to Kubernetes cluster
4. **Verification**: Ensure pods are running and services accessible
5. **Testing**: Validate functionality before proceeding

### 🚫 What NOT to Do

KubeChat enforces container-first development:

```bash
# ❌ NEVER run these commands
pnpm run dev           # No local frontend processes
go run ./cmd/server    # No local backend processes
npm start             # No local development servers

# ✅ ALWAYS use these instead
make dev-deploy        # Deploy to Kubernetes
make dev-shell-web     # Access frontend container
make dev-shell-api     # Access backend container
```

## 📁 Project Structure

```
kubechat/
├── apps/                          # Application services
│   ├── web/                       # React TypeScript frontend
│   │   ├── src/                   # Source code
│   │   ├── public/                # Static assets
│   │   └── package.json           # Dependencies
│   └── api/                       # Go backend services
│       ├── cmd/                   # Application entry points
│       ├── internal/              # Private application code
│       ├── pkg/                   # Public Go packages
│       └── go.mod                 # Go dependencies
│
├── packages/                      # Shared packages
│   ├── shared/                    # TypeScript types
│   ├── ui/                        # Component library
│   └── config/                    # Shared configuration
│
├── infrastructure/                # DevOps and deployment
│   ├── helm/                      # Kubernetes deployment charts
│   │   └── kubechat/             # Main Helm chart
│   ├── docker/                    # Container configurations
│   ├── scripts/                   # Automation scripts
│   └── k8s/                       # Raw Kubernetes manifests
│
├── docs/                          # Project documentation
├── tests/                         # Integration tests
├── .github/                       # CI/CD workflows
├── Makefile                       # Development commands
└── DEVELOPMENT.md                 # This file
```

## 🔧 Configuration

### Environment Variables

KubeChat uses Kubernetes ConfigMaps and Secrets for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `NODE_ENV` | `development` | Runtime environment |
| `GIN_MODE` | `debug` | Gin framework mode |
| `DB_HOST` | `kubechat-dev-postgresql` | Database service host |
| `REDIS_HOST` | `kubechat-dev-redis-master` | Redis service host |
| `LOG_LEVEL` | `debug` | Logging level |
| `DEBUG` | `true` | Enable debug mode |

**Note**: Service names include the Helm release prefix `kubechat-dev-` in development.

### Customization

Edit configuration files in `infrastructure/helm/kubechat/`:

- `values.yaml` - Production defaults
- `values-dev.yaml` - Development overrides  
- `templates/configmap.yaml` - Application configuration

## 🧪 Testing

### Test Structure

```
tests/
├── unit/                         # Component/function tests
├── integration/                  # Service integration tests
├── e2e/                          # End-to-end scenarios
└── performance/                  # Load and performance tests
```

### Running Tests

```bash
# All tests
make dev-test

# Specific test types
make dev-test-unit                # Unit tests only
make dev-test-e2e                 # End-to-end tests only

# Frontend tests in container
make dev-shell-web
npm test

# Backend tests in container  
make dev-shell-api
go test ./...
```

## 🛟 Troubleshooting

### Common Issues

#### Environment Validation Failures

```bash
# Re-run validation with verbose output
./infrastructure/scripts/validate-env.sh --verbose

# Check specific tool versions
make dev-info
```

#### Container Build Failures

```bash
# Clean Docker cache
docker system prune -f

# Rebuild from scratch
make dev-clean
make dev-build

# For Rancher Desktop image import issues
docker save kubechat/web:dev | nerdctl --address /var/run/docker/containerd/containerd.sock --namespace k8s.io load
docker save kubechat/api:dev | nerdctl --address /var/run/docker/containerd/containerd.sock --namespace k8s.io load
```

#### Deployment Issues

```bash
# Check deployment status
make dev-status

# View detailed logs
make dev-logs

# Debug specific pods
kubectl describe pods -n kubechat

# Check events
kubectl get events -n kubechat --sort-by='.firstTimestamp'

# Common fixes for deployment issues
helm uninstall kubechat-dev -n kubechat  # Clean slate
make dev-deploy                          # Redeploy fresh

# For stuck deployments
kubectl rollout restart deployment/kubechat-dev-api -n kubechat
kubectl rollout restart deployment/kubechat-dev-web -n kubechat
```

#### Database Connection Issues

```bash
# Check database pod
kubectl get pods -n kubechat -l app.kubernetes.io/name=postgresql

# Test database connectivity
make dev-db-connect

# Reset database if needed
make dev-db-reset
```

### Getting Help

1. **Check Logs**: `make dev-logs`
2. **Validate Environment**: `./infrastructure/scripts/validate-env.sh`
3. **Health Check**: `./infrastructure/scripts/health-check.sh`
4. **Debug Information**: `make dev-debug`

## 🔒 Security

KubeChat implements enterprise-grade security:

### Features

- **🔐 RBAC Integration**: Kubernetes role-based access control
- **📋 Complete Audit Trail**: All actions logged with cryptographic integrity
- **🏠 Air-Gapped Deployment**: No external dependencies required
- **🔑 JWT Authentication**: Secure session management
- **🛡️ Input Validation**: SQL injection and XSS protection

### Security Configuration

```bash
# View security settings
kubectl get rbac -n kubechat

# Check audit logs
make dev-db-connect
SELECT * FROM audit_log_entries ORDER BY created_at DESC LIMIT 10;

# Review security policies
cat infrastructure/helm/kubechat/templates/rbac.yaml
```

## 🤝 Contributing

### Development Process

1. **Fork and Clone**: Fork the repository and clone your fork
2. **Environment Setup**: Run `make init` to set up your development environment
3. **Create Branch**: Create a feature branch for your changes
4. **Develop**: Make changes following our container-first approach
5. **Test**: Run `make dev-test` to ensure all tests pass
6. **Submit**: Create a pull request with a clear description

### Code Standards

- **Frontend**: TypeScript, React 18+, Tailwind CSS
- **Backend**: Go 1.23+, Gin framework, clean architecture
- **Testing**: Unit tests for all components, integration tests for APIs
- **Documentation**: Update docs for any new features or changes

### Coding Standards

Follow the guidelines in [docs/architecture/coding-standards.md](docs/architecture/coding-standards.md):

- **Type Safety**: Always define types in `packages/shared`
- **API Communication**: Never make direct HTTP calls - use service layer
- **Environment Configuration**: Access only through config objects
- **Error Handling**: All API routes must use standard error handler
- **State Management**: Never mutate state directly - use Zustand patterns
- **Container-First**: Never run local processes - use containers only

## 📋 Development Checklist

### Before Starting Development

- [ ] Rancher Desktop installed and Kubernetes enabled
- [ ] All prerequisites validated with `./infrastructure/scripts/validate-env.sh`
- [ ] Development environment initialized with `make init`
- [ ] Can access all services via NodePort URLs

### For Each Feature

- [ ] Create feature branch from `develop`
- [ ] Follow container-first development (no local processes)
- [ ] Make code changes in your editor
- [ ] Rebuild containers: `make dev-rebuild-api` or `make dev-rebuild-web`
- [ ] Deploy changes: `helm upgrade kubechat-dev infrastructure/helm/kubechat --namespace kubechat --values infrastructure/helm/kubechat/values-dev.yaml`
- [ ] Verify pods are running: `kubectl get pods -n kubechat`
- [ ] Test functionality: API (`curl http://localhost:30080/health`) and Web (`curl http://localhost:30001`)
- [ ] Add tests for new functionality
- [ ] Update documentation if needed
- [ ] Run full test suite with `make dev-test`
- [ ] Check code quality with linting tools

### Before Pull Request

- [ ] All tests passing in container environment
- [ ] No security vulnerabilities introduced
- [ ] Documentation updated for user-facing changes
- [ ] Follows established coding standards
- [ ] Database migrations tested if applicable

## 📚 Additional Resources

- **Architecture**: [docs/architecture/](docs/architecture/)
- **Deployment Guide**: [docs/architecture/deployment.md](docs/architecture/deployment.md)
- **Coding Standards**: [docs/architecture/coding-standards.md](docs/architecture/coding-standards.md)
- **Technology Stack**: [docs/architecture/tech-stack.md](docs/architecture/tech-stack.md)
- **Source Tree**: [docs/architecture/source-tree.md](docs/architecture/source-tree.md)

---

**Built with ❤️ for container-first development**

For questions about the development environment or container-first approach, please create an issue or check our troubleshooting guide.
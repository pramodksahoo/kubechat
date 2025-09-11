# KubeChat Deployment Guide

This document provides comprehensive deployment instructions for KubeChat across different environments.

## Development Environment Configuration

### Prerequisites

**Required Tools with Installation:**

| Tool | Version | Installation Command |
|------|---------|---------------------|
| **Docker** | Latest | Included with Rancher Desktop |
| **Kubernetes** | 1.28+ | Rancher Desktop |
| **Helm** | 3.15+ | `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \| bash` |
| **kubectl** | 1.28+ | Included with Rancher Desktop |
| **PNPM** | 8+ | `curl -fsSL https://get.pnpm.io/install.sh \| sh` |
| **Go** | 1.23+ | Download from https://go.dev/dl/ |

### Rancher Desktop Setup

1. **Download and Install:**
   ```bash
   # macOS
   brew install --cask rancher
   
   # Or download from https://rancherdesktop.io/
   ```

2. **Configure Rancher Desktop:**
   - Enable Kubernetes
   - Set Container Runtime to dockerd
   - Allocate 8GB+ RAM and 4+ CPU cores
   - Enable port forwarding

3. **Verify Installation:**
   ```bash
   # Check Kubernetes cluster
   kubectl cluster-info
   
   # Verify nodes are ready
   kubectl get nodes
   
   # Check storage classes
   kubectl get storageclass
   ```

### Expected Validation Output

**Kubernetes Control Plane:**
```
Kubernetes control plane is running at https://127.0.0.1:6443
```

**Node Status:**
```
NAME                   STATUS   ROLES           AGE   VERSION
lima-rancher-desktop   Ready    control-plane   1d    v1.28.x
```

**Storage Classes:**
```
NAME                   PROVISIONER                     RECLAIMPOLICY
local-path (default)   rancher.io/local-path           Delete
```

## Complete Makefile Commands

### Environment Management
```bash
make dev-info          # Show system information and prerequisites
make dev-setup         # Initial development environment setup  
make dev-clean         # Clean development environment
```

### Container Management
```bash
make dev-build         # Build all application containers
make dev-rebuild-api   # Rebuild API container only
make dev-rebuild-web   # Rebuild web container only
make dev-push          # Push containers to local registry
```

### Deployment Management
```bash
make dev-deploy        # Deploy complete development stack
make dev-upgrade       # Upgrade existing deployment
make dev-status        # Show deployment status
make dev-undeploy      # Remove development deployment
make dev-rollback      # Restore previous working state
```

### Development Tools
```bash
make dev-logs          # View aggregated application logs
make dev-logs-api      # View API service logs only
make dev-logs-web      # View web service logs only
make dev-shell-api     # Shell into API container
make dev-shell-web     # Shell into web container
make dev-port-forward  # Setup port forwarding for local access
```

### Database Management
```bash
make dev-db-connect    # Connect to development database
make dev-db-migrate    # Run database migrations
make dev-db-seed       # Seed development data
make dev-db-reset      # Reset database to clean state
```

### Testing
```bash
make dev-test          # Run all tests in containers
make dev-test-unit     # Run unit tests only
make dev-test-e2e      # Run end-to-end tests
```

## Port Forwarding Configuration

### Application Services
```bash
localhost:3001  -> kubechat-web       # Frontend application
localhost:8080  -> kubechat-api       # Backend API  
localhost:11434 -> ollama             # AI inference service
```

### Development Tools
```bash
localhost:5432  -> postgresql         # Database direct access
localhost:6379  -> redis              # Redis direct access
localhost:5050  -> pgadmin            # Database admin UI
localhost:8081  -> redis-commander    # Redis admin UI
```

## Environment Verification Commands

### Check Kubernetes Cluster Status
```bash
kubectl cluster-info
```

### Verify Nodes are Ready
```bash
kubectl get nodes
```

### Check Available Storage Classes
```bash
kubectl get storageclass
```

### Validate KubeChat Deployment
```bash
# Check all pods are running
kubectl get pods -n kubechat

# Verify services are accessible
curl http://localhost:8080/health
curl http://localhost:3001
```

## Container-First Development Workflow

### 1. Initial Setup
```bash
# Clone repository
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Initialize environment (one-time setup)
make init
```

### 2. Daily Development Workflow
```bash
# Start development session
make dev-status          # Check current deployment status

# Make code changes in your editor

# Rebuild affected containers
make dev-rebuild-api     # For backend changes
make dev-rebuild-web     # For frontend changes

# Deploy changes to Kubernetes
helm upgrade kubechat-dev infrastructure/helm/kubechat \
  --namespace kubechat \
  --values infrastructure/helm/kubechat/values-dev.yaml

# Verify deployment
kubectl get pods -n kubechat
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=kubechat-dev -n kubechat

# View logs
make dev-logs

# Test changes
curl http://localhost:8080/health  # API health check
curl http://localhost:3001         # Frontend check
```

### 3. Troubleshooting
```bash
# Check deployment status
make dev-status

# View detailed logs
make dev-logs

# Debug specific pods
kubectl describe pods -n kubechat

# Check events
kubectl get events -n kubechat --sort-by='.firstTimestamp'

# Reset environment if needed
make dev-clean
make dev-deploy
```

## Production Deployment

### Prerequisites
- Kubernetes cluster (1.28+)
- Helm 3.15+
- kubectl configured for target cluster
- Persistent storage provisioner

### Production Values Configuration
```bash
# Copy and customize production values
cp infrastructure/helm/kubechat/values.yaml values-production.yaml

# Edit values-production.yaml:
# - Set appropriate resource limits
# - Configure persistent storage
# - Set production environment variables
# - Configure ingress for external access
```

### Production Deployment Commands
```bash
# Add Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Install dependencies
cd infrastructure/helm/kubechat
helm dependency update

# Deploy to production
helm upgrade --install kubechat infrastructure/helm/kubechat \
  --namespace kubechat \
  --create-namespace \
  --values values-production.yaml \
  --wait \
  --timeout 600s
```

## Staging Environment

### Staging-Specific Configuration
```bash
# Use staging values
helm upgrade --install kubechat-staging infrastructure/helm/kubechat \
  --namespace kubechat-staging \
  --create-namespace \
  --values infrastructure/helm/kubechat/values-staging.yaml \
  --wait
```

## Environment Variables

### Development Environment
```bash
NODE_ENV=development
GIN_MODE=debug
DB_HOST=kubechat-dev-postgresql
REDIS_HOST=kubechat-dev-redis-master
LOG_LEVEL=debug
DEBUG=true
```

### Production Environment
```bash
NODE_ENV=production
GIN_MODE=release
DB_HOST=kubechat-postgresql
REDIS_HOST=kubechat-redis-master
LOG_LEVEL=info
```

## Security Considerations

### Container Security
- All containers run as non-root users
- Specific version tags (never 'latest')
- Multi-stage builds for minimal attack surface
- Regular vulnerability scanning

### Kubernetes Security
- RBAC enabled with minimal permissions
- Network policies for pod-to-pod communication
- Secret management for sensitive data
- Resource quotas and limits

### Application Security
- JWT authentication for API access
- CORS configuration for cross-origin requests
- Input validation and sanitization
- Audit logging for all operations

This deployment guide ensures consistent, secure, and reliable KubeChat deployments across all environments while maintaining the container-first development approach.

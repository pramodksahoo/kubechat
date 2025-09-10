# KubeChat Deployment Guide

## Overview

This document provides comprehensive deployment instructions for KubeChat across different environments, from local development to production air-gapped deployments.

**Last Updated:** 2025-01-10  
**Repository:** https://github.com/pramodksahoo/kubechat  
**Branch:** develop

---

## Prerequisites

### Required Tools

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Docker** | Latest | Container runtime | Included with Rancher Desktop |
| **Kubernetes** | 1.28+ | Container orchestration | Rancher Desktop |
| **Helm** | 3.15+ | Kubernetes package manager | `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \| bash` |
| **kubectl** | 1.28+ | Kubernetes CLI | Included with Rancher Desktop |
| **PNPM** | 8+ | Package manager | `curl -fsSL https://get.pnpm.io/install.sh \| sh` |
| **Go** | 1.23+ | Backend build | Download from https://go.dev/dl/ |

### Rancher Desktop Setup

**Verify Rancher Desktop is running:**

```bash
# Check Kubernetes cluster status
kubectl cluster-info

# Verify nodes are ready
kubectl get nodes

# Check available storage classes
kubectl get storageclass
```

**Expected Output:**
```bash
Kubernetes control plane is running at https://127.0.0.1:6443
CoreDNS is running at https://127.0.0.1:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

NAME                          STATUS   ROLES                  AGE     VERSION
rancher-desktop              Ready    control-plane,master   1d      v1.28.5+k3s1

NAME         PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  1d
```

---

## Development Deployment

### Quick Start

```bash
# Clone repository
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Switch to develop branch
git checkout develop

# Verify development environment
make dev-info

# Deploy full development stack
make dev-deploy

# Monitor deployment status
make dev-status

# View application logs
make dev-logs
```

### Makefile Commands

**Core Development Commands:**

```bash
# Environment Management
make dev-info          # Show system information and prerequisites
make dev-setup         # Initial development environment setup
make dev-clean         # Clean development environment

# Container Management  
make dev-build         # Build all application containers
make dev-rebuild-api   # Rebuild API container only
make dev-rebuild-web   # Rebuild web container only
make dev-push          # Push containers to local registry

# Deployment Management
make dev-deploy        # Deploy complete development stack
make dev-upgrade       # Upgrade existing deployment
make dev-status        # Show deployment status
make dev-undeploy      # Remove development deployment

# Development Tools
make dev-logs          # View aggregated application logs
make dev-logs-api      # View API service logs only
make dev-logs-web      # View web service logs only
make dev-shell-api     # Shell into API container
make dev-shell-web     # Shell into web container
make dev-port-forward  # Setup port forwarding for local access

# Database Management
make dev-db-connect    # Connect to development database
make dev-db-migrate    # Run database migrations
make dev-db-seed       # Seed development data
make dev-db-reset      # Reset database to clean state

# Testing
make dev-test          # Run all tests in containers
make dev-test-unit     # Run unit tests only
make dev-test-e2e      # Run end-to-end tests
```

### Development Stack Components

**Services Deployed:**

```yaml
# KubeChat Application Stack
- kubechat-api         # Go backend API service
- kubechat-web         # React TypeScript frontend
- kubechat-ai          # AI processing service

# Infrastructure Dependencies
- postgresql           # Primary database
- redis                # Cache and sessions
- ollama               # Local AI inference

# Development Tools
- pgadmin              # Database administration (dev only)
- redis-commander      # Redis management (dev only)
```

**Port Forwarding Configuration:**

```bash
# Application Services
localhost:3000  -> kubechat-web       # Frontend application
localhost:8080  -> kubechat-api       # Backend API
localhost:11434 -> ollama             # AI inference service

# Development Tools  
localhost:5432  -> postgresql         # Database direct access
localhost:6379  -> redis              # Redis direct access
localhost:5050  -> pgadmin            # Database admin UI
localhost:8081  -> redis-commander    # Redis admin UI
```

---

## Production Deployment

### Production Prerequisites

```bash
# Verify production Kubernetes cluster
kubectl cluster-info
kubectl get nodes

# Verify Helm is configured
helm version
helm repo list

# Check required storage classes
kubectl get storageclass

# Verify ingress controller
kubectl get ingress-class
```

### Production Values Configuration

**Create production values file:**

```yaml
# values-production.yaml
global:
  environment: production
  domain: kubechat.yourdomain.com
  
# Application Configuration
app:
  replicas:
    api: 3
    web: 2
  resources:
    api:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "2Gi"
        cpu: "1000m"
    web:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "1Gi"
        cpu: "500m"

# Database Configuration
postgresql:
  enabled: true
  persistence:
    enabled: true
    size: 100Gi
    storageClass: "fast-ssd"
  resources:
    requests:
      memory: "1Gi"
      cpu: "500m"
    limits:
      memory: "4Gi" 
      cpu: "2000m"
  auth:
    database: kubechat
    username: kubechat
    existingSecret: kubechat-db-secret

# Redis Configuration
redis:
  enabled: true
  auth:
    enabled: true
    existingSecret: kubechat-redis-secret
  persistence:
    enabled: true
    size: 10Gi
    storageClass: "fast-ssd"

# AI Configuration
ollama:
  enabled: true
  persistence:
    enabled: true
    size: 50Gi
    storageClass: "fast-ssd"
  models:
    - phi3.5-mini
    - codellama:34b
  resources:
    requests:
      memory: "8Gi"
      cpu: "2000m"
    limits:
      memory: "16Gi"
      cpu: "4000m"

# Ingress Configuration
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
  hosts:
    - host: kubechat.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: kubechat-tls
      hosts:
        - kubechat.yourdomain.com

# Monitoring
monitoring:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true
    ingress:
      enabled: true
      host: grafana.yourdomain.com

# Security
security:
  networkPolicies:
    enabled: true
  podSecurityPolicy:
    enabled: true
  rbac:
    create: true
    
# High Availability
autoscaling:
  enabled: true
  api:
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
  web:
    minReplicas: 2
    maxReplicas: 5
    targetCPUUtilizationPercentage: 80

# Pod Disruption Budgets
podDisruptionBudget:
  api:
    minAvailable: 1
  web:
    minAvailable: 1
```

### Production Secrets Management

**Create required secrets:**

```bash
# Database credentials
kubectl create secret generic kubechat-db-secret \
  --from-literal=postgres-password=YOUR_SECURE_PASSWORD \
  --from-literal=password=YOUR_SECURE_PASSWORD

# Redis credentials  
kubectl create secret generic kubechat-redis-secret \
  --from-literal=redis-password=YOUR_SECURE_REDIS_PASSWORD

# Application secrets
kubectl create secret generic kubechat-app-secret \
  --from-literal=jwt-secret=YOUR_JWT_SECRET \
  --from-literal=api-key=YOUR_API_KEY

# TLS certificates (if not using cert-manager)
kubectl create secret tls kubechat-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key
```

### Production Deployment Commands

```bash
# Add Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Create namespace
kubectl create namespace kubechat

# Deploy ingress controller (if needed)
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace

# Deploy cert-manager (if using TLS)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Deploy KubeChat
helm install kubechat ./infrastructure/helm/kubechat \
  --namespace kubechat \
  --values values-production.yaml \
  --wait \
  --timeout 10m

# Verify deployment
kubectl get all -n kubechat
kubectl get ingress -n kubechat
kubectl get certificates -n kubechat
```

### Production Health Checks

```bash
# Check pod status
kubectl get pods -n kubechat

# Check service endpoints
kubectl get endpoints -n kubechat

# Check ingress configuration
kubectl describe ingress kubechat -n kubechat

# Test application health
curl -k https://kubechat.yourdomain.com/health
curl -k https://kubechat.yourdomain.com/api/health

# Check resource utilization
kubectl top pods -n kubechat
kubectl top nodes
```

---

## Staging Deployment

### Staging Configuration

**Create staging values file:**

```yaml
# values-staging.yaml
global:
  environment: staging
  domain: staging.kubechat.yourdomain.com

app:
  replicas:
    api: 2
    web: 1
  
postgresql:
  persistence:
    size: 20Gi
    
redis:
  persistence:
    size: 2Gi
    
ollama:
  persistence:
    size: 20Gi
  models:
    - phi3.5-mini  # Staging uses smaller model set
```

**Deploy staging environment:**

```bash
# Deploy to staging namespace
kubectl create namespace kubechat-staging

helm install kubechat-staging ./infrastructure/helm/kubechat \
  --namespace kubechat-staging \
  --values values-staging.yaml \
  --wait
```

---

## Troubleshooting

### Common Issues

**1. Pod Startup Issues**

```bash
# Check pod status and events
kubectl describe pod <pod-name> -n kubechat

# Check container logs
kubectl logs <pod-name> -n kubechat -c <container-name>

# Check resource constraints
kubectl top pods -n kubechat
kubectl describe node <node-name>
```

**2. Database Connection Issues**

```bash
# Test database connectivity
kubectl exec -it deployment/kubechat-api -n kubechat -- sh
nc -zv kubechat-postgresql 5432

# Check database pod logs
kubectl logs -f kubechat-postgresql-0 -n kubechat

# Verify database credentials
kubectl get secret kubechat-db-secret -n kubechat -o yaml
```

**3. AI Service Issues**

```bash
# Check Ollama model status
kubectl exec -it deployment/kubechat-ollama -n kubechat -- ollama list

# Test AI service connectivity  
kubectl exec -it deployment/kubechat-api -n kubechat -- sh
curl http://kubechat-ollama:11434/api/tags

# Check Ollama logs
kubectl logs -f deployment/kubechat-ollama -n kubechat
```

**4. Ingress Issues**

```bash
# Check ingress controller
kubectl get pods -n ingress-nginx
kubectl logs -f deployment/ingress-nginx-controller -n ingress-nginx

# Verify ingress configuration
kubectl describe ingress kubechat -n kubechat

# Check TLS certificates
kubectl get certificates -n kubechat
kubectl describe certificate kubechat-tls -n kubechat
```

### Performance Tuning

**Database Optimization:**

```yaml
# postgresql values for high performance
postgresql:
  primary:
    configuration: |
      shared_buffers = 256MB
      effective_cache_size = 1GB
      maintenance_work_mem = 64MB
      checkpoint_completion_target = 0.9
      wal_buffers = 16MB
      default_statistics_target = 100
      random_page_cost = 1.1
      effective_io_concurrency = 200
```

**API Service Optimization:**

```yaml
app:
  api:
    env:
      GOGC: "100"
      GOMEMLIMIT: "1.5GiB"
      GOMAXPROCS: "2"
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "2Gi"
        cpu: "1000m"
```

### Monitoring Commands

```bash
# Watch deployment rollout
kubectl rollout status deployment/kubechat-api -n kubechat

# Monitor resource usage
watch kubectl top pods -n kubechat

# Check application metrics
kubectl port-forward svc/kubechat-api 8080:8080 -n kubechat
curl http://localhost:8080/metrics

# View recent events
kubectl get events -n kubechat --sort-by=.metadata.creationTimestamp
```

---

## Backup and Recovery

### Database Backup

```bash
# Create database backup
kubectl exec kubechat-postgresql-0 -n kubechat -- \
  pg_dump -U kubechat kubechat > kubechat-backup-$(date +%Y%m%d).sql

# Automated backup with CronJob
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  name: kubechat-db-backup
  namespace: kubechat
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: db-backup
            image: postgres:16
            command:
            - /bin/bash
            - -c
            - pg_dump -h kubechat-postgresql -U kubechat kubechat > /backup/kubechat-\$(date +%Y%m%d).sql
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: kubechat-db-secret
                  key: postgres-password
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: kubechat-backups
          restartPolicy: OnFailure
EOF
```

### Application Recovery

```bash
# Restore from backup
kubectl exec -i kubechat-postgresql-0 -n kubechat -- \
  psql -U kubechat kubechat < kubechat-backup-20250110.sql

# Rolling restart after recovery
kubectl rollout restart deployment/kubechat-api -n kubechat
kubectl rollout restart deployment/kubechat-web -n kubechat
```

---

## Scaling and Maintenance

### Horizontal Scaling

```bash
# Scale API replicas
kubectl scale deployment kubechat-api --replicas=5 -n kubechat

# Scale web replicas  
kubectl scale deployment kubechat-web --replicas=3 -n kubechat

# Enable autoscaling
kubectl autoscale deployment kubechat-api \
  --cpu-percent=70 \
  --min=2 \
  --max=10 \
  -n kubechat
```

### Update Procedures

```bash
# Update application images
helm upgrade kubechat ./infrastructure/helm/kubechat \
  --namespace kubechat \
  --values values-production.yaml \
  --set app.image.tag=v1.2.0

# Monitor rollout
kubectl rollout status deployment/kubechat-api -n kubechat
kubectl rollout status deployment/kubechat-web -n kubechat

# Rollback if needed
helm rollback kubechat -n kubechat
```

### Maintenance Windows

```bash
# Drain node for maintenance
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# Cordon node (prevent scheduling)
kubectl cordon <node-name>

# Uncordon node after maintenance
kubectl uncordon <node-name>
```

---

## Next Steps

1. **Security Hardening:** Implement network policies and pod security standards
2. **Monitoring Setup:** Deploy Prometheus and Grafana for comprehensive monitoring
3. **Backup Strategy:** Implement automated backup and disaster recovery procedures
4. **Load Testing:** Perform thorough load testing before production deployment

For deployment issues or questions, consult the troubleshooting section or create an issue in the repository.

---

*📚 Generated with [Claude Code](https://claude.ai/code)*

*Co-Authored-By: Winston (Architect) <architect@kubechat.dev>*
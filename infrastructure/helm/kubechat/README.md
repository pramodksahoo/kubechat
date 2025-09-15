# KubeChat Helm Chart

This Helm chart deploys KubeChat, a Natural Language Kubernetes Management Platform with Enterprise Production Features, to a Kubernetes cluster.

## Prerequisites

- **Kubernetes**: 1.25+ (tested with 1.28+)
- **Helm**: 3.15+
- **cert-manager**: Required for TLS certificate management (production)
- **Ingress Controller**:
  - Production: nginx with SSL/TLS support
  - Development: traefik (default for local k8s)
- **StorageClass**:
  - Production: `fast-ssd` for high-performance storage
  - Development: `local-path` (Rancher Desktop default)

## Chart Components

- **Web Frontend**: Next.js application (port 3000)
- **API Backend**: Go-based REST API (port 8080)
- **PostgreSQL**: Primary database with HA support (Bitnami chart)
- **Redis**: Caching and session storage with replication (Bitnami chart)
- **Ollama**: AI service for natural language processing (enabled in development)
- **Dev Tools**: pgAdmin and Redis Commander (development only)

## Installation

### Development Environment

```bash
# Add required Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Update chart dependencies
helm dependency update

# Install with development configuration
helm install kubechat-dev ./kubechat -f values-dev.yaml

# Access services (NodePort configuration):
# Web UI: http://localhost:30001 or http://kubechat.local
# API: http://localhost:30080
# pgAdmin: http://localhost:30050 (admin@kubechat.dev / dev-admin)
# Redis Commander: http://localhost:30081
# Ollama: http://localhost:30434
```

### Production Deployment

```bash
# Update dependencies
helm dependency update

# Install with production configuration
helm install kubechat ./kubechat -f values.yaml \
  --namespace kubechat-production \
  --create-namespace

# Access via configured ingress:
# Web UI: https://kubechat.com
# API: https://kubechat.com/api
```

### Custom Domain Installation

```bash
# Install with custom domain and TLS
helm install kubechat ./kubechat -f values.yaml \
  --set ingress.hosts[0].host=kubechat.yourdomain.com \
  --set tls.certManager.issuer.email=admin@yourdomain.com \
  --set ingress.tls[0].hosts[0]=kubechat.yourdomain.com
```

## Configuration

### Required Values

```yaml
# Ingress configuration
ingress:
  hosts:
    - host: kubechat.yourdomain.com

# TLS configuration
tls:
  certManager:
    issuer:
      email: admin@yourdomain.com
```

### Environment-Specific Configurations

The chart includes two pre-configured value files:

- `values.yaml` - Production configuration with high availability, security, and TLS
- `values-dev.yaml` - Development configuration with debugging tools and relaxed security

### Key Configuration Sections

#### Web Frontend
**Development:**
```yaml
web:
  replicaCount: 1
  image:
    repository: kubechat/web
    tag: "dev"
  service:
    type: NodePort
    nodePort: 30001
  env:
    NODE_ENV: "development"
    NEXT_PUBLIC_API_BASE_URL: http://localhost:30080/api
```

**Production:**
```yaml
web:
  replicaCount: 5
  image:
    repository: kubechat/web
    tag: "1.0.0"
  service:
    type: ClusterIP
  env:
    NODE_ENV: "production"
    NEXT_PUBLIC_API_BASE_URL: https://api.kubechat.com
```

#### API Backend
**Development:**
```yaml
api:
  replicaCount: 1
  image:
    repository: kubechat/api
    tag: "dev"
  service:
    type: NodePort
    nodePort: 30080
  env:
    GIN_MODE: "debug"
    LOG_LEVEL: debug
```

**Production:**
```yaml
api:
  replicaCount: 5
  image:
    repository: kubechat/api
    tag: "1.0.0"
  service:
    type: ClusterIP
  env:
    GIN_MODE: "release"
    LOG_LEVEL: warn
```

#### Database (PostgreSQL)
**Development:**
```yaml
postgresql:
  enabled: true
  auth:
    postgresPassword: "dev-postgres"
    username: kubechat
    password: "dev-password"
    database: kubechat
  primary:
    persistence:
      size: 5Gi
      storageClass: "local-path"
```

**Production:**
```yaml
postgresql:
  enabled: true
  architecture: repmgr  # High availability
  auth:
    existingSecret: kubechat-postgresql-secret
    database: kubechat_production
  primary:
    persistence:
      size: 100Gi
      storageClass: "fast-ssd"
  readReplicas:
    replicaCount: 2
```

#### Redis Cache
**Development:**
```yaml
redis:
  enabled: true
  auth:
    enabled: true
    password: "dev-redis"
  master:
    persistence:
      size: 2Gi
      storageClass: "local-path"
  replica:
    replicaCount: 1
```

**Production:**
```yaml
redis:
  enabled: true
  architecture: replication  # High availability
  auth:
    enabled: true
    existingSecret: kubechat-redis-secret
  master:
    persistence:
      size: 20Gi
      storageClass: "fast-ssd"
  replica:
    replicaCount: 2
```

## Monitoring

The chart includes comprehensive monitoring capabilities:

### Prometheus Integration
```yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
  prometheusRules:
    enabled: true
```

### Grafana Dashboards
```yaml
monitoring:
  grafanaDashboards:
    enabled: true
    labels:
      grafana_dashboard: "1"
```

## Security

### RBAC
```yaml
rbac:
  create: true
  rules:
    - apiGroups: [""]
      resources: ["pods", "services"]
      verbs: ["get", "list", "watch"]
```

### Network Policies
```yaml
networkPolicy:
  enabled: true
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            name: ingress-nginx
```

### Pod Security Standards
```yaml
podSecurityStandards:
  enabled: true
  enforce: restricted
```

## Backup & Recovery

### Database Backup
```yaml
backup:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  retention:
    days: 30
    weeks: 8
```

## High Availability

### Horizontal Pod Autoscaling
```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70
```

### Pod Disruption Budget
```yaml
podDisruptionBudget:
  enabled: true
  minAvailable: 50%
```

## Troubleshooting

### Common Issues

1. **Certificate Issues**
   ```bash
   # Check certificate status
   kubectl get certificate -n kubechat

   # Check cert-manager logs
   kubectl logs -n cert-manager deploy/cert-manager
   ```

2. **Database Connection Issues**
   ```bash
   # Check PostgreSQL status
   kubectl get pods -n kubechat -l app.kubernetes.io/name=postgresql

   # View connection logs
   kubectl logs -n kubechat deployment/kubechat-api
   ```

3. **Template Validation**
   ```bash
   # Validate templates
   helm template kubechat ./kubechat --values values.yaml
   ```

### Health Checks

All services include comprehensive health checks:

- **Startup Probes**: Ensure services start correctly
- **Liveness Probes**: Restart unhealthy containers
- **Readiness Probes**: Route traffic only to ready pods

### Monitoring & Alerting

The chart includes Prometheus alerts for:

- High CPU/Memory usage
- Pod crash looping
- Service availability
- High response times

## Upgrading

### Standard Upgrade
```bash
# Update dependencies
helm dependency update

# Upgrade release
helm upgrade kubechat ./kubechat -f values.yaml
```

### Rolling Back
```bash
# View release history
helm history kubechat

# Rollback to previous version
helm rollback kubechat
```

## Uninstallation

```bash
# Uninstall release
helm uninstall kubechat

# Clean up persistent volumes (if needed)
kubectl delete pvc -l app.kubernetes.io/instance=kubechat
```

## Values Reference

For a complete list of configurable values, see:
- `values.yaml` - Production configuration with HA, security, and enterprise features
- `values-dev.yaml` - Development configuration with debugging tools and relaxed security

### Key Differences Between Environments

| Feature | Development | Production |
|---------|-------------|------------|
| **Replicas** | 1 per service | 5 web, 5 api |
| **Image Tags** | `dev` | `1.0.0` |
| **Service Type** | NodePort | ClusterIP |
| **Storage** | `local-path` (5-10Gi) | `fast-ssd` (100Gi+) |
| **Security** | Relaxed | Strict (Pod Security Standards) |
| **TLS/HTTPS** | Disabled | Let's Encrypt + cert-manager |
| **Ingress** | traefik (kubechat.local) | nginx (kubechat.com) |
| **Database** | Single instance | HA with read replicas |
| **Redis** | Single master + 1 replica | Master + 2 replicas |
| **Backup** | Disabled | Daily with 30-day retention |
| **Monitoring** | Basic | Full Prometheus integration |
| **Dev Tools** | pgAdmin, Redis Commander | Disabled |
| **Ollama AI** | Enabled | Disabled |
| **Autoscaling** | Disabled | Enabled (3-50 pods) |

## Support

For issues and questions:
- Review the troubleshooting section above
- Check the [KubeChat documentation](https://kubechat.dev)
- Open an issue in the GitHub repository

## Version Compatibility

| Chart Version | KubeChat Version | Kubernetes Version |
|---------------|------------------|--------------------|
| 1.0.0         | 1.0.0           | 1.25+              |

## License

This chart is licensed under the MIT License.
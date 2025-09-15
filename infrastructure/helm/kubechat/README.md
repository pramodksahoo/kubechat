# KubeChat Helm Chart

This Helm chart deploys KubeChat, a Natural Language Kubernetes Management Platform, to a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.25+
- Helm 3.15+
- cert-manager (for TLS certificates)
- Ingress controller (nginx recommended)
- StorageClass for persistent volumes

## Installation

### Quick Start (Development)

```bash
# Add required Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Install with default development settings
helm install kubechat ./kubechat
```

### Production Deployment

```bash
# Install with production values
helm install kubechat ./kubechat -f values.yaml

# Or use staging configuration
helm install kubechat-staging ./kubechat -f values-staging.yaml
```

### Custom Installation

```bash
# Install with custom domain and TLS
helm install kubechat ./kubechat \
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

The chart includes three pre-configured value files:

- `values.yaml` - Production configuration
- `values-staging.yaml` - Staging configuration
- `values-production.yaml` - Alternative production configuration

### Key Configuration Sections

#### Web Frontend
```yaml
web:
  enabled: true
  replicaCount: 2
  image:
    repository: kubechat/web
    tag: "1.0.0"
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
```

#### API Backend
```yaml
api:
  enabled: true
  replicaCount: 3
  image:
    repository: kubechat/api
    tag: "1.0.0"
  resources:
    limits:
      cpu: 2000m
      memory: 2Gi
```

#### Database (PostgreSQL)
```yaml
postgresql:
  enabled: true
  auth:
    database: kubechat
    username: kubechat
  primary:
    persistence:
      size: 20Gi
```

#### Redis Cache
```yaml
redis:
  enabled: true
  auth:
    enabled: true
  master:
    persistence:
      size: 8Gi
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
- `values.yaml` - All available options with defaults
- `values-staging.yaml` - Staging-specific overrides
- `values-production.yaml` - Production-specific overrides

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
# KubeChat Deployment Guide

This guide provides step-by-step instructions for deploying KubeChat to your Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (1.19+) with at least:
  - 4GB RAM available for Ollama
  - 2 CPU cores available
  - 10GB persistent storage
- Helm 3.x installed
- kubectl configured to access your cluster

## Quick Deployment

### 1. Deploy KubeChat with In-Cluster Ollama

```bash
# Clone the repository
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Deploy using Helm
helm install kubechat ./chart \
  --set llm.provider=ollama \
  --set llm.ollama.enabled=true \
  --set llm.ollama.deploy=true \
  --create-namespace \
  --namespace kubechat
```

### 2. Wait for Deployment

```bash
# Watch the deployment progress
kubectl get pods -n kubechat -w

# Expected pods:
# - kubechat-xxxxxxxxx-xxxxx (KubeChat main application)
# - kubechat-ollama-xxxxxxxxx-xxxxx (Ollama LLM service)
# - kubechat-ollama-init-xxxxx (Job to pull the model)
```

### 3. Access the Application

```bash
# Port forward to access locally
kubectl port-forward svc/kubechat 8080:8080 -n kubechat

# Open in browser
open http://localhost:8080
```

## Configuration Options

### Resource Requirements

The default configuration is optimized for development/testing:

```yaml
# KubeChat application
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"

# Ollama service  
ollama:
  resources:
    requests:
      memory: "2Gi"
      cpu: "1000m"
    limits:
      memory: "4Gi"
      cpu: "2000m"
```

### Production Deployment

For production use, create a custom values file:

```yaml
# values-production.yaml
replicaCount: 2

resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"

llm:
  ollama:
    resources:
      requests:
        memory: "4Gi"
        cpu: "2000m"
      limits:
        memory: "8Gi"
        cpu: "4000m"
    persistence:
      size: "50Gi"
      storageClass: "fast-ssd"

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: kubechat.yourcompany.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: kubechat-tls
      hosts:
        - kubechat.yourcompany.com

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

Deploy with custom values:

```bash
helm install kubechat ./chart \
  --values values-production.yaml \
  --namespace kubechat \
  --create-namespace
```

## Alternative LLM Providers

### Using OpenAI

```bash
# Create secret for OpenAI API key
kubectl create secret generic openai-secret \
  --from-literal=api-key="sk-your-openai-key" \
  -n kubechat

# Deploy with OpenAI
helm install kubechat ./chart \
  --set llm.provider=openai \
  --set llm.openai.enabled=true \
  --set llm.openai.existingSecret=openai-secret \
  --set llm.ollama.enabled=false \
  --namespace kubechat \
  --create-namespace
```

### Using External Ollama

```bash
# Deploy with external Ollama service
helm install kubechat ./chart \
  --set llm.provider=ollama \
  --set llm.ollama.enabled=true \
  --set llm.ollama.deploy=false \
  --set llm.ollama.url="http://external-ollama.example.com:11434" \
  --namespace kubechat \
  --create-namespace
```

## Monitoring and Troubleshooting

### Check Deployment Status

```bash
# Check all resources
kubectl get all -n kubechat

# Check logs
kubectl logs -f deployment/kubechat -n kubechat
kubectl logs -f deployment/kubechat-ollama -n kubechat

# Check Ollama model initialization
kubectl logs -f job/kubechat-ollama-init -n kubechat
```

### Test API Endpoints

```bash
# Port forward
kubectl port-forward svc/kubechat 8080:8080 -n kubechat

# Test health endpoint
curl http://localhost:8080/api/health

# Test cluster info
curl http://localhost:8080/api/clusters

# Test natural language queries
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "show me all pods"}'

curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "list all services in the kubechat namespace"}'

# Test Ollama connectivity (if deployed in-cluster)
kubectl port-forward svc/kubechat-ollama 11434:11434 -n kubechat
curl http://localhost:11434/api/tags

# Test Ollama model directly
curl -X POST http://localhost:11434/api/generate \
  -H "Content-Type: application/json" \
  -d '{"model": "phi3:mini", "prompt": "Hello", "stream": false}'
```

### Common Issues

#### Ollama Pod Stuck in Pending

```bash
# Check node resources
kubectl describe nodes

# Check persistent volume
kubectl get pv,pvc -n kubechat

# Scale down if needed
kubectl scale deployment kubechat-ollama --replicas=0 -n kubechat
```

#### Model Pull Job Fails

```bash
# Check job logs
kubectl logs job/kubechat-ollama-init -n kubechat

# Manually pull model (use phi3:mini for better memory efficiency)
kubectl exec -it deployment/kubechat-ollama -n kubechat -- ollama pull phi3:mini
```

## Memory Optimization

### Using Smaller Models

For clusters with limited memory (< 8GB), use the phi3:mini model instead of llama2:

```bash
# Update the model in values.yaml
helm upgrade kubechat ./chart \
  --set llm.ollama.model=phi3:mini \
  -n kubechat

# Or patch the running deployment
kubectl patch deployment kubechat -n kubechat \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"kubechat","env":[{"name":"OLLAMA_MODEL","value":"phi3:mini"}]}]}}}}'
```

### Model Specifications

| Model | Size | Memory Required | Performance |
|-------|------|----------------|-------------|
| llama2 | 3.8GB | ~6GB | High quality responses |
| phi3:mini | 2.2GB | ~3GB | Good quality, faster |
| qwen2:0.5b | 0.5GB | ~1GB | Basic responses |

### Managing Models

```bash
# List available models
kubectl exec -n kubechat deployment/kubechat-ollama -- ollama list

# Remove unused models to save space
kubectl exec -n kubechat deployment/kubechat-ollama -- ollama rm llama2

# Pull specific model version
kubectl exec -n kubechat deployment/kubechat-ollama -- ollama pull phi3:mini
```

## Tested Configuration

This deployment has been tested successfully on:

- **Platform**: Rancher Desktop on macOS
- **Kubernetes Version**: 1.28+
- **Memory**: 8GB+ allocated to Rancher Desktop
- **Model**: phi3:mini (2.2GB)
- **Test Queries**:
  - "show me all pods" → `kubectl get pods`
  - "list all services in kubechat namespace" → `kubectl get svc -n kubechat`

## Cleanup

```bash
# Uninstall the release
helm uninstall kubechat -n kubechat

# Delete namespace (optional)
kubectl delete namespace kubechat
```

## Security Considerations

- KubeChat uses RBAC with minimal required permissions
- All LLM communication stays within the cluster when using in-cluster Ollama
- Audit logs are maintained for all operations
- No sensitive data is logged or stored by default

For production deployments, consider:
- Network policies to restrict pod-to-pod communication
- Pod security policies or pod security standards
- Regular security scans of container images
- Monitoring and alerting on unusual activity
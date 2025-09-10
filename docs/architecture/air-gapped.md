# KubeChat Air-Gapped Installation Guide

## Overview

This document provides comprehensive instructions for deploying KubeChat in air-gapped environments where internet access is restricted or unavailable. This is a critical competitive advantage for enterprise and government deployments.

**Last Updated:** 2025-01-10  
**Repository:** https://github.com/pramodksahoo/kubechat  
**Branch:** develop

---

## Air-Gapped Architecture Benefits

### Competitive Advantages

**🔒 Data Sovereignty**
- All AI processing happens locally
- No external API calls or data transmission
- Complete data locality compliance

**🏢 Enterprise Security** 
- Meets strict security requirements (SOX, HIPAA, SOC 2)
- Government and defense contractor compliant
- Zero external network dependencies

**💰 Cost Control**
- No per-token pricing for AI queries
- Predictable infrastructure costs
- No bandwidth charges for AI processing

**🚀 Performance**
- Local AI inference reduces latency
- No network bottlenecks for core functionality
- Optimized for Kubernetes command generation

---

## Prerequisites

### Infrastructure Requirements

**Minimum Hardware Specifications:**

| Component | Development | Staging | Production |
|-----------|-------------|---------|------------|
| **CPU** | 8 cores | 16 cores | 32+ cores |
| **Memory** | 32 GB | 64 GB | 128+ GB |
| **Storage** | 500 GB SSD | 1 TB SSD | 2+ TB SSD |
| **GPU** | Optional | Recommended | Required for large models |

**Kubernetes Cluster Requirements:**

```yaml
# Minimum cluster specifications
nodes: 3              # High availability
kubernetes: "1.28+"   # Supported version
storage: "persistent" # For AI models and data
network: "CNI"        # Container networking
ingress: "nginx"      # Load balancing
```

### Required Files

**Pre-Download Requirements:**
- Container images (Docker registry export)
- Helm charts with dependencies
- AI models (Ollama format)
- Operating system packages
- SSL certificates

---

## Preparation Phase (Connected Environment)

### 1. Container Image Export

**Export all required container images:**

```bash
# Create air-gap preparation directory
mkdir -p kubechat-airgap
cd kubechat-airgap

# Build and export KubeChat images
make airgap-build-images

# Export application images
docker save kubechat/api:latest kubechat/web:latest kubechat/ai:latest \
  -o kubechat-app-images.tar

# Export dependency images
docker save \
  postgres:16-alpine \
  redis:7.4-alpine \
  ollama/ollama:0.3 \
  nginx:1.25-alpine \
  bitnami/postgresql:16 \
  bitnami/redis:7.4 \
  -o kubechat-deps-images.tar

# Export monitoring images (optional)
docker save \
  prom/prometheus:latest \
  grafana/grafana:latest \
  prom/node-exporter:latest \
  -o kubechat-monitoring-images.tar

# Verify image exports
ls -lah *.tar
```

### 2. Helm Charts Bundle

**Package Helm charts with all dependencies:**

```bash
# Update Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Package main application chart
helm package ./infrastructure/helm/kubechat \
  --dependency-update \
  --destination ./charts/

# Package dependency charts
helm pull bitnami/postgresql --version 13.2.24 --destination ./charts/
helm pull bitnami/redis --version 18.1.5 --destination ./charts/
helm pull prometheus-community/kube-prometheus-stack --version 51.2.0 --destination ./charts/

# Create chart index
helm repo index ./charts/

# Bundle all charts
tar -czf kubechat-helm-charts.tar.gz charts/
```

### 3. AI Models Download

**Download and package AI models:**

```bash
# Create models directory
mkdir -p models

# Download Phi-3.5-mini model (primary Kubernetes model)
ollama pull phi3.5-mini
ollama save phi3.5-mini models/phi3.5-mini.bin

# Download CodeLlama model (code generation)
ollama pull codellama:34b
ollama save codellama:34b models/codellama-34b.bin

# Download backup models
ollama pull llama3.2:3b
ollama save llama3.2:3b models/llama3.2-3b.bin

# Package models
tar -czf kubechat-ai-models.tar.gz models/

# Verify model package
ls -lah kubechat-ai-models.tar.gz
```

### 4. Operating System Packages

**Package required system dependencies:**

```bash
# For Ubuntu/Debian systems
mkdir -p packages/ubuntu
apt-get download \
  curl \
  wget \
  gnupg \
  ca-certificates \
  software-properties-common \
  apt-transport-https

# For RHEL/CentOS systems  
mkdir -p packages/rhel
yumdownloader \
  curl \
  wget \
  gnupg2 \
  ca-certificates

# Package system dependencies
tar -czf kubechat-system-packages.tar.gz packages/
```

### 5. SSL Certificates

**Prepare SSL certificates:**

```bash
mkdir -p certificates

# Copy your SSL certificates
cp /path/to/your/domain.crt certificates/
cp /path/to/your/domain.key certificates/
cp /path/to/ca-bundle.crt certificates/

# Or generate self-signed certificates for testing
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certificates/kubechat.key \
  -out certificates/kubechat.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=kubechat.local"

# Package certificates
tar -czf kubechat-certificates.tar.gz certificates/
```

### 6. Create Air-Gap Bundle

**Create complete installation bundle:**

```bash
# Create installation scripts
cat > install-airgap.sh << 'EOF'
#!/bin/bash
set -e

echo "KubeChat Air-Gapped Installation"
echo "================================"

# Verify prerequisites
./scripts/verify-prerequisites.sh

# Load container images
./scripts/load-images.sh

# Setup local registry
./scripts/setup-registry.sh

# Install AI models
./scripts/install-models.sh

# Deploy KubeChat
./scripts/deploy-kubechat.sh

echo "Installation completed successfully!"
EOF

chmod +x install-airgap.sh

# Package complete bundle
tar -czf kubechat-airgap-bundle.tar.gz \
  kubechat-app-images.tar \
  kubechat-deps-images.tar \
  kubechat-monitoring-images.tar \
  kubechat-helm-charts.tar.gz \
  kubechat-ai-models.tar.gz \
  kubechat-system-packages.tar.gz \
  kubechat-certificates.tar.gz \
  install-airgap.sh \
  scripts/ \
  manifests/ \
  docs/

echo "Air-gap bundle created: kubechat-airgap-bundle.tar.gz"
ls -lah kubechat-airgap-bundle.tar.gz
```

---

## Installation Phase (Air-Gapped Environment)

### 1. Prerequisites Verification

**Verify air-gapped environment readiness:**

```bash
# Extract installation bundle
tar -xzf kubechat-airgap-bundle.tar.gz
cd kubechat-airgap

# Verify Kubernetes cluster
kubectl cluster-info
kubectl get nodes
kubectl get storageclass

# Check available resources
kubectl describe nodes | grep -A 3 "Allocated resources"

# Verify persistent storage
kubectl get pv
```

### 2. Local Container Registry Setup

**Setup local Docker registry for images:**

```bash
# Create registry namespace
kubectl create namespace container-registry

# Deploy local registry
cat > registry-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: docker-registry
  namespace: container-registry
spec:
  replicas: 1
  selector:
    matchLabels:
      app: docker-registry
  template:
    metadata:
      labels:
        app: docker-registry
    spec:
      containers:
      - name: registry
        image: registry:2
        ports:
        - containerPort: 5000
        env:
        - name: REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY
          value: "/var/lib/registry"
        volumeMounts:
        - name: registry-storage
          mountPath: /var/lib/registry
      volumes:
      - name: registry-storage
        persistentVolumeClaim:
          claimName: registry-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: registry-pvc
  namespace: container-registry
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
---
apiVersion: v1
kind: Service
metadata:
  name: docker-registry
  namespace: container-registry
spec:
  selector:
    app: docker-registry
  ports:
  - port: 5000
    targetPort: 5000
  type: ClusterIP
EOF

kubectl apply -f registry-deployment.yaml

# Wait for registry to be ready
kubectl wait --for=condition=ready pod -l app=docker-registry -n container-registry --timeout=300s
```

### 3. Load Container Images

**Load and push images to local registry:**

```bash
# Load images into Docker
docker load -i kubechat-app-images.tar
docker load -i kubechat-deps-images.tar
docker load -i kubechat-monitoring-images.tar

# Setup port forward to local registry
kubectl port-forward svc/docker-registry 5000:5000 -n container-registry &
REGISTRY_PID=$!

# Wait for port forward
sleep 5

# Tag and push images to local registry
REGISTRY_URL="localhost:5000"

# Push KubeChat application images
docker tag kubechat/api:latest ${REGISTRY_URL}/kubechat/api:latest
docker tag kubechat/web:latest ${REGISTRY_URL}/kubechat/web:latest
docker tag kubechat/ai:latest ${REGISTRY_URL}/kubechat/ai:latest

docker push ${REGISTRY_URL}/kubechat/api:latest
docker push ${REGISTRY_URL}/kubechat/web:latest
docker push ${REGISTRY_URL}/kubechat/ai:latest

# Push dependency images
docker tag postgres:16-alpine ${REGISTRY_URL}/postgres:16-alpine
docker tag redis:7.4-alpine ${REGISTRY_URL}/redis:7.4-alpine
docker tag ollama/ollama:0.3 ${REGISTRY_URL}/ollama/ollama:0.3

docker push ${REGISTRY_URL}/postgres:16-alpine
docker push ${REGISTRY_URL}/redis:7.4-alpine
docker push ${REGISTRY_URL}/ollama/ollama:0.3

# Kill port forward
kill $REGISTRY_PID
```

### 4. Install AI Models

**Setup Ollama with pre-downloaded models:**

```bash
# Extract AI models
tar -xzf kubechat-ai-models.tar.gz

# Create models ConfigMap
kubectl create configmap kubechat-ai-models \
  --from-file=models/ \
  --namespace=kubechat

# Deploy Ollama with models
cat > ollama-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubechat-ollama
  namespace: kubechat
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubechat-ollama
  template:
    metadata:
      labels:
        app: kubechat-ollama
    spec:
      initContainers:
      - name: model-loader
        image: localhost:5000/ollama/ollama:0.3
        command: ["/bin/sh", "-c"]
        args:
        - |
          cp /models/* /root/.ollama/models/ || true
          ollama serve &
          sleep 10
          ollama load phi3.5-mini /models/phi3.5-mini.bin
          ollama load codellama:34b /models/codellama-34b.bin
          ollama load llama3.2:3b /models/llama3.2-3b.bin
        volumeMounts:
        - name: model-storage
          mountPath: /root/.ollama
        - name: models-data
          mountPath: /models
      containers:
      - name: ollama
        image: localhost:5000/ollama/ollama:0.3
        ports:
        - containerPort: 11434
        env:
        - name: OLLAMA_HOST
          value: "0.0.0.0:11434"
        volumeMounts:
        - name: model-storage
          mountPath: /root/.ollama
        resources:
          requests:
            memory: "8Gi"
            cpu: "2000m"
          limits:
            memory: "16Gi"
            cpu: "4000m"
      volumes:
      - name: model-storage
        persistentVolumeClaim:
          claimName: ollama-models-pvc
      - name: models-data
        configMap:
          name: kubechat-ai-models
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ollama-models-pvc
  namespace: kubechat
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
---
apiVersion: v1
kind: Service
metadata:
  name: kubechat-ollama
  namespace: kubechat
spec:
  selector:
    app: kubechat-ollama
  ports:
  - port: 11434
    targetPort: 11434
EOF

kubectl apply -f ollama-deployment.yaml

# Wait for Ollama to be ready
kubectl wait --for=condition=ready pod -l app=kubechat-ollama -n kubechat --timeout=600s
```

### 5. Deploy KubeChat Application

**Deploy with air-gapped configuration:**

```bash
# Extract Helm charts
tar -xzf kubechat-helm-charts.tar.gz

# Create air-gapped values file
cat > values-airgap.yaml << 'EOF'
global:
  environment: airgap
  imageRegistry: "localhost:5000"
  
app:
  image:
    registry: localhost:5000
    repository: kubechat
    tag: latest
    pullPolicy: IfNotPresent
    
  replicas:
    api: 2
    web: 1
    
# Use local registry for all images
postgresql:
  enabled: true
  image:
    registry: localhost:5000
    repository: postgres
    tag: 16-alpine
  persistence:
    enabled: true
    size: 20Gi

redis:
  enabled: true
  image:
    registry: localhost:5000
    repository: redis
    tag: 7.4-alpine
  persistence:
    enabled: true
    size: 5Gi

# Ollama is deployed separately above
ollama:
  enabled: false
  externalService: "kubechat-ollama"

# Disable external dependencies
ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
  hosts:
    - host: kubechat.local
      paths:
        - path: /
          pathType: Prefix

# Security for air-gapped
security:
  networkPolicies:
    enabled: true
    denyAll: true  # Block all external traffic
  podSecurityContext:
    runAsNonRoot: true
    runAsUser: 10001
    fsGroup: 10001

# Monitoring (local only)
monitoring:
  enabled: true
  external: false
  
# Disable telemetry and external integrations
telemetry:
  enabled: false
  
integrations:
  external: false
  openai: false
  anthropic: false
EOF

# Install KubeChat
helm install kubechat ./charts/kubechat-*.tgz \
  --namespace kubechat \
  --create-namespace \
  --values values-airgap.yaml \
  --wait \
  --timeout 15m

# Verify deployment
kubectl get all -n kubechat
```

### 6. SSL Certificate Configuration

**Configure SSL certificates:**

```bash
# Extract certificates
tar -xzf kubechat-certificates.tar.gz

# Create TLS secret
kubectl create secret tls kubechat-tls \
  --cert=certificates/kubechat.crt \
  --key=certificates/kubechat.key \
  --namespace kubechat

# Update ingress for TLS
kubectl patch ingress kubechat -n kubechat --type='merge' -p='
{
  "spec": {
    "tls": [
      {
        "hosts": ["kubechat.local"],
        "secretName": "kubechat-tls"
      }
    ]
  }
}'
```

---

## Post-Installation Configuration

### 1. Network Policies

**Implement strict network isolation:**

```yaml
# Create network policies for air-gapped security
cat > network-policies.yaml << 'EOF'
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kubechat-deny-all
  namespace: kubechat
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kubechat-allow-internal
  namespace: kubechat
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: kubechat
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kubechat
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
EOF

kubectl apply -f network-policies.yaml
```

### 2. Access Configuration

**Setup local access:**

```bash
# Get ingress IP
kubectl get ingress kubechat -n kubechat

# Add to hosts file (replace IP with actual ingress IP)
echo "192.168.1.100 kubechat.local" | sudo tee -a /etc/hosts

# Or setup port forwarding
kubectl port-forward svc/kubechat-web 3000:80 -n kubechat &
kubectl port-forward svc/kubechat-api 8080:80 -n kubechat &
```

### 3. Verification Tests

**Verify air-gapped functionality:**

```bash
# Test AI model access
kubectl exec -it deployment/kubechat-ollama -n kubechat -- \
  curl -X POST http://localhost:11434/api/generate \
  -d '{"model":"phi3.5-mini","prompt":"List Kubernetes pods"}'

# Test API connectivity
curl -k https://kubechat.local/api/health
curl -k https://kubechat.local/api/models

# Test web interface
curl -k https://kubechat.local

# Verify no external network calls
kubectl logs -l app=kubechat-api -n kubechat | grep -E "(http|https)://" | grep -v localhost
```

---

## Air-Gapped Operations

### Model Management

**Managing AI models in air-gapped environment:**

```bash
# List available models
kubectl exec -it deployment/kubechat-ollama -n kubechat -- ollama list

# Model performance testing
kubectl exec -it deployment/kubechat-ollama -n kubechat -- \
  ollama run phi3.5-mini "How do I list pods in Kubernetes?"

# Model resource monitoring
kubectl top pods -l app=kubechat-ollama -n kubechat
```

### Backup Procedures

**Air-gapped backup strategy:**

```bash
# Database backup
kubectl exec kubechat-postgresql-0 -n kubechat -- \
  pg_dump -U kubechat kubechat > kubechat-backup-$(date +%Y%m%d).sql

# AI models backup
kubectl exec -it deployment/kubechat-ollama -n kubechat -- \
  tar -czf /tmp/models-backup.tar.gz /root/.ollama/models

kubectl cp kubechat-ollama-pod:/tmp/models-backup.tar.gz \
  ./models-backup-$(date +%Y%m%d).tar.gz

# Application configuration backup
kubectl get configmap,secret -n kubechat -o yaml > kubechat-config-backup.yaml
```

### Updates and Maintenance

**Updating air-gapped installation:**

```bash
# Prepare update bundle (connected environment)
make airgap-update-bundle VERSION=v1.1.0

# Apply update (air-gapped environment)
tar -xzf kubechat-update-v1.1.0.tar.gz

# Load new images
docker load -i kubechat-update-images.tar

# Push to local registry
./scripts/update-local-registry.sh

# Helm upgrade
helm upgrade kubechat ./charts/kubechat-1.1.0.tgz \
  --namespace kubechat \
  --values values-airgap.yaml

# Verify update
kubectl rollout status deployment/kubechat-api -n kubechat
kubectl rollout status deployment/kubechat-web -n kubechat
```

---

## Security Hardening

### Pod Security Standards

```yaml
# Implement pod security standards
apiVersion: v1
kind: Namespace
metadata:
  name: kubechat
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### RBAC Configuration

```yaml
# Minimal RBAC for air-gapped environment
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubechat-airgap-readonly
rules:
- apiGroups: [""]
  resources: ["pods", "services", "nodes", "events"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]
```

### Audit Configuration

**Enable comprehensive audit logging:**

```yaml
# Kubernetes audit policy for air-gapped
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
  namespaces: ["kubechat"]
  resources:
  - group: ""
    resources: ["*"]
- level: Request
  users: ["kubechat-api"]
  verbs: ["create", "update", "patch", "delete"]
```

---

## Troubleshooting Air-Gapped Issues

### Common Problems

**1. Image Pull Failures**
```bash
# Check local registry status
kubectl get pods -n container-registry
kubectl logs deployment/docker-registry -n container-registry

# Verify image availability
curl http://localhost:5000/v2/_catalog
curl http://localhost:5000/v2/kubechat/api/tags/list
```

**2. AI Model Loading Issues**
```bash
# Check model loading status
kubectl logs deployment/kubechat-ollama -n kubechat

# Verify model files
kubectl exec -it deployment/kubechat-ollama -n kubechat -- \
  ls -la /root/.ollama/models/

# Test model functionality
kubectl exec -it deployment/kubechat-ollama -n kubechat -- \
  ollama run phi3.5-mini "test query"
```

**3. Network Connectivity Issues**
```bash
# Test internal connectivity
kubectl exec -it deployment/kubechat-api -n kubechat -- \
  nc -zv kubechat-postgresql 5432

kubectl exec -it deployment/kubechat-api -n kubechat -- \
  nc -zv kubechat-redis 6379

kubectl exec -it deployment/kubechat-api -n kubechat -- \
  nc -zv kubechat-ollama 11434
```

### Performance Optimization

**Air-gapped specific optimizations:**

```yaml
# Optimize for local processing
app:
  api:
    env:
      AI_CACHE_SIZE: "1000"
      AI_PARALLEL_REQUESTS: "4"
      QUERY_TIMEOUT: "30s"
    resources:
      requests:
        memory: "1Gi"
        cpu: "500m"
      limits:
        memory: "4Gi"
        cpu: "2000m"

ollama:
  resources:
    requests:
      memory: "8Gi"
      cpu: "4000m"
    limits:
      memory: "16Gi"
      cpu: "8000m"
```

---

## Compliance and Validation

### Security Compliance Checklist

**Air-gapped security validation:**

- [ ] ✅ No external network connections from application pods
- [ ] ✅ All container images from local registry
- [ ] ✅ AI models loaded locally without external downloads
- [ ] ✅ Network policies block external traffic
- [ ] ✅ Pod security standards enforced
- [ ] ✅ RBAC minimal permissions applied
- [ ] ✅ Audit logging enabled and configured
- [ ] ✅ Secrets managed securely within cluster
- [ ] ✅ TLS encryption for internal communication
- [ ] ✅ Database encryption at rest enabled

### Validation Tests

**Comprehensive air-gapped validation:**

```bash
# Run validation test suite
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: kubechat-airgap-validation
  namespace: kubechat
spec:
  template:
    spec:
      containers:
      - name: validator
        image: localhost:5000/kubechat/validator:latest
        command: ["/bin/sh", "-c"]
        args:
        - |
          echo "Testing air-gapped functionality..."
          
          # Test AI model access
          curl -f http://kubechat-ollama:11434/api/tags || exit 1
          
          # Test database connectivity
          nc -zv kubechat-postgresql 5432 || exit 1
          
          # Test Redis connectivity
          nc -zv kubechat-redis 6379 || exit 1
          
          # Test API endpoints
          curl -f http://kubechat-api/health || exit 1
          
          # Verify no external network access
          ! curl -f https://google.com --max-time 5 || exit 1
          
          echo "All air-gapped validation tests passed!"
      restartPolicy: OnFailure
EOF

# Monitor validation
kubectl logs -f job/kubechat-airgap-validation -n kubechat
```

---

## Next Steps

1. **Security Review:** Conduct thorough security assessment of air-gapped deployment
2. **Performance Testing:** Validate AI model performance under load
3. **Backup Strategy:** Implement comprehensive backup and recovery procedures
4. **Documentation:** Create environment-specific operational procedures
5. **Training:** Train operations team on air-gapped maintenance procedures

For air-gapped deployment support or questions, consult the troubleshooting section or contact the KubeChat support team through secure channels.

---

*📚 Generated with [Claude Code](https://claude.ai/code)*

*Co-Authored-By: Winston (Architect) <architect@kubechat.dev>*
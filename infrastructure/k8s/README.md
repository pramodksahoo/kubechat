# Kubernetes Raw Manifests

This directory contains raw Kubernetes YAML manifests for direct kubectl deployment.

## Directory Structure

### `/base/` - Base Kubernetes Resources
Common Kubernetes manifests that can be used across environments:
- Namespace definitions
- Service accounts and RBAC
- ConfigMaps and Secrets templates
- Base deployments and services

### `/overlays/` - Environment-Specific Configurations
Environment-specific overlays using Kustomize:

- `/overlays/development/` - Development environment configurations
- `/overlays/staging/` - Staging environment configurations  
- `/overlays/production/` - Production environment configurations

## Usage

### Direct kubectl Deployment
```bash
# Deploy base resources
kubectl apply -f infrastructure/k8s/base/

# Deploy environment-specific resources
kubectl apply -k infrastructure/k8s/overlays/development/
```

### Kustomize Integration
```bash
# Preview changes
kubectl kustomize infrastructure/k8s/overlays/development/

# Apply with kustomize
kubectl apply -k infrastructure/k8s/overlays/development/
```

## Relationship to Helm

These raw manifests complement the Helm charts in `/infrastructure/helm/`:

- **Helm Charts**: Primary deployment method with templating and values
- **Raw Manifests**: Direct deployment for specific scenarios or debugging
- **Kustomize Overlays**: Environment-specific configurations without templating

## When to Use

- **Use Helm** (recommended): For normal development and production deployments
- **Use Raw Manifests**: For debugging, testing, or environments where Helm is not available
- **Use Kustomize**: For environment-specific configurations with raw manifests

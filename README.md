# KubeChat

**Natural Language Kubernetes Management Platform**

KubeChat transforms complex kubectl operations into intuitive conversational interfaces, enabling teams to manage Kubernetes infrastructure through natural language commands while maintaining enterprise-grade security and compliance.

## 🚀 Features

- **Natural Language Interface**: Chat with your Kubernetes cluster using plain English
- **Multi-LLM Support**: Works with Ollama (local/air-gapped) and OpenAI
- **Enterprise-Grade Security**: Built-in RBAC integration and audit logging
- **Real-time Dashboard**: Modern web interface with live cluster monitoring
- **Intelligent Command Translation**: Converts natural language to safe kubectl operations
- **Audit Trail**: Complete logging of all queries and operations for compliance

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   React UI      │    │   Go Backend     │    │   Kubernetes    │
│   (Frontend)    │◄──►│   (API Server)   │◄──►│   Cluster       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   LLM Provider   │
                       │ (Ollama/OpenAI)  │
                       └──────────────────┘
```

## 🛠️ Quick Start

### Prerequisites

- **Kubernetes Cluster**: Rancher Desktop, minikube, or any K8s cluster (with 4GB+ RAM available)
- **Helm 3.x**: For deployment
- **kubectl**: Configured to access your cluster
- **Go 1.21+**: For building the backend (development only)
- **Node.js 18+**: For building the frontend (development only)

### 1. Prerequisites Check

```bash
# Verify Kubernetes cluster is running
kubectl cluster-info

# Check available resources (Ollama needs ~4GB RAM)
kubectl top nodes

# Verify you have Helm installed
helm version
```

### 2. Deploy to Kubernetes

```bash
# Clone the repository
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Deploy with Helm (includes Ollama in-cluster)
helm install kubechat ./chart \
  --create-namespace \
  --namespace kubechat

# Wait for deployment (Ollama model pull takes ~5 minutes)
kubectl get pods -n kubechat -w
```

### 3. Access KubeChat

```bash
# Port forward to access locally
kubectl port-forward svc/kubechat 8080:8080 -n kubechat

# Open in browser
open http://localhost:8080
```

## 🛠️ Development Mode

For local development without Kubernetes deployment:

```bash
# Prerequisites: Local Ollama installation
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull llama2

# Quick setup
make quickstart

# Run development servers
make dev-backend    # Terminal 1
make dev-frontend   # Terminal 2
```

## 💬 Example Queries

Try these natural language commands:

```
"Show me all pods in the default namespace"
"Which nodes are running in my cluster?"
"Get services in the kube-system namespace"
"Show me pods with high CPU usage"
"Display the logs for the nginx pod"
```

## 🐳 Docker Development

For testing the container image locally:

```bash
# Build Docker image
make docker-build

# Run with Docker (requires external Ollama)
make docker-run
```

## ☸️ Advanced Kubernetes Deployment

### Default Deployment (Ollama In-Cluster)

```bash
# Standard deployment - everything runs in Kubernetes
helm install kubechat ./chart --namespace kubechat --create-namespace
```

### Alternative LLM Providers

```bash
# External Ollama
helm install kubechat ./chart \
  --set llm.ollama.deploy=false \
  --set llm.ollama.url="http://your-ollama-service:11434" \
  --namespace kubechat --create-namespace

# OpenAI
helm install kubechat ./chart \
  --set llm.provider=openai \
  --set llm.openai.enabled=true \
  --set llm.openai.apiKey="your-api-key" \
  --namespace kubechat --create-namespace
```

### Production Deployment

```bash
# With ingress and resource optimization
helm install kubechat ./chart \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=kubechat.yourcompany.com \
  --set resources.requests.memory="512Mi" \
  --set llm.ollama.resources.requests.memory="4Gi" \
  --namespace kubechat --create-namespace
```

### Access the Application

```bash
# Local access via port-forward
kubectl port-forward svc/kubechat 8080:8080 -n kubechat

# Or configure ingress (see DEPLOYMENT.md for details)
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | LLM provider (`ollama` or `openai`) | `ollama` |
| `OLLAMA_URL` | Ollama service URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | Ollama model name | `llama2` |
| `OPENAI_API_KEY` | OpenAI API key | `""` |
| `OPENAI_MODEL` | OpenAI model name | `gpt-4` |
| `LLM_FALLBACK` | Enable fallback between providers | `true` |

### Helm Values

See `chart/values.yaml` for comprehensive configuration options including:
- Resource limits and requests
- Ingress configuration
- Security settings
- Persistence options
- RBAC settings

## 🔒 Security & RBAC

KubeChat is designed with security-first principles:

- **In-Cluster LLM**: Ollama runs within your cluster - no external AI API calls
- **RBAC Integration**: Respects existing Kubernetes RBAC policies
- **Minimal Permissions**: Service account with read-only access to core resources
- **Complete Audit Trail**: All queries and commands logged with user attribution
- **Air-Gap Ready**: Fully functional without internet connectivity
- **Command Validation**: All LLM-generated commands validated before execution

## 📊 Monitoring & Observability

### Health Endpoints

- `GET /api/health` - Application health status
- `GET /api/audit` - Recent audit logs
- `GET /api/clusters` - Cluster information

### Logs

```bash
# View application logs
kubectl logs -f deployment/kubechat -n kubechat

# View Ollama logs
kubectl logs -f deployment/kubechat-ollama -n kubechat

# View audit logs via API (after port-forward)
curl http://localhost:8080/api/audit?limit=10
```

## 🧪 Development

### Project Structure

```
├── cmd/server/          # Main application entry point
├── internal/            # Go backend packages
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP handlers
│   ├── k8s/           # Kubernetes client wrapper
│   ├── llm/           # LLM integration layer
│   ├── translator/    # Natural language translation engine
│   └── audit/         # Audit logging
├── web/                # React frontend
├── chart/              # Helm chart
└── docs/               # Documentation
```

### Adding New Commands

1. Add command handling in `internal/translator/engine.go`
2. Extend the LLM prompt in `internal/llm/service.go`
3. Add corresponding Kubernetes API calls in `internal/k8s/client.go`

### Testing

```bash
# Run all tests
make test

# Run specific tests
go test ./internal/translator/...
```

## 🚧 Roadmap

### Phase 1 (Current - PoC)
- [x] Basic natural language to kubectl translation
- [x] Ollama and OpenAI integration
- [x] Web dashboard with chat interface
- [x] RBAC integration and audit logging
- [x] Helm chart deployment

### Phase 2 (Enterprise Features)
- [ ] Multi-cluster management
- [ ] Advanced export & integration APIs
- [ ] Enterprise SSO/LDAP integration
- [ ] Advanced approval workflows
- [ ] Sophisticated knowledge learning

### Phase 3 (Advanced AI)
- [ ] Predictive operations recommendations
- [ ] Automated incident response
- [ ] Custom model fine-tuning
- [ ] Industry-specific templates

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/pramodksahoo/kubechat/issues)
- **Discussions**: [GitHub Discussions](https://github.com/pramodksahoo/kubechat/discussions)

---

## 🚀 Quick Start Summary

**For Production/Testing (Kubernetes):**
```bash
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat
helm install kubechat ./chart --namespace kubechat --create-namespace
kubectl port-forward svc/kubechat 8080:8080 -n kubechat
# Open http://localhost:8080
```

**For Development (Local):**
```bash
# Setup Ollama locally first
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull llama2

# Then run KubeChat
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat
make quickstart
make dev-backend    # Terminal 1  
make dev-frontend   # Terminal 2
```

**Key Features:**
- 🤖 **In-Cluster AI**: Ollama runs inside your Kubernetes cluster
- 🔒 **Air-Gap Ready**: No external API calls required  
- 🎯 **Natural Language**: "show me all pods" → kubectl get pods
- 📋 **Enterprise Audit**: Complete compliance logging
- ⚡ **Multi-LLM**: Ollama (default) + OpenAI support

---

**Built with ❤️ for the Kubernetes community**

*KubeChat makes Kubernetes accessible through the power of natural language AI while maintaining enterprise-grade security and compliance.*
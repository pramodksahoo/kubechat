# KubeChat

**Natural Language Kubernetes Management Platform**

> Transform complex kubectl operations into intuitive conversations with your cluster while maintaining enterprise-grade security and compliance.

![KubeChat Banner](https://img.shields.io/badge/KubeChat-Natural_Language_K8s_Management-blue?style=for-the-badge&logo=kubernetes)

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18+-61DAFB.svg?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6.svg?logo=typescript)](https://www.typescriptlang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.24+-blue.svg)](https://kubernetes.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg?logo=docker)](https://www.docker.com/)
[![Helm](https://img.shields.io/badge/Helm-3.x-0F1689.svg?logo=helm)](https://helm.sh/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Ready-336791.svg?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-Ready-DC382D.svg?logo=redis)](https://redis.io/)
[![Ollama](https://img.shields.io/badge/Ollama-Local_AI-000000.svg)](https://ollama.ai/)
[![OpenAI](https://img.shields.io/badge/OpenAI-Optional-412991.svg?logo=openai)](https://openai.com/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-3.x-38B2AC.svg?logo=tailwind-css)](https://tailwindcss.com/)

## 🌟 What is KubeChat?

KubeChat is an **open-source Natural Language Kubernetes Management Platform** that bridges the gap between Kubernetes complexity and operational efficiency. Instead of memorizing kubectl commands, simply chat with your cluster:

```
"Show me all pods with high memory usage in production"
"Scale the payment-service deployment to handle weekend traffic" 
"What's causing the API gateway to be unhealthy?"
```

**🎯 Perfect for:** DevOps teams, SREs, and platform engineers who need simplified Kubernetes management without compromising security or compliance standards.

## 🚀 Why KubeChat?

### ✨ **Natural Language First**
- Chat with your cluster using plain English
- No more memorizing complex kubectl syntax
- Contextual suggestions and intelligent command translation

### 🔒 **Enterprise-Grade Security**
- **Air-gapped deployment** with Ollama (runs completely offline)
- Built-in RBAC integration and audit logging
- Zero external API calls required for AI processing
- Complete compliance-ready audit trails

### 🎛️ **Modern Web Interface**
- Beautiful, responsive dashboard with real-time cluster monitoring
- Multi-user collaborative troubleshooting
- Progressive disclosure for beginners to experts
- WebSocket-powered live updates

### 🧠 **Multi-LLM Support**
- **Ollama** (default) - Local, air-gapped AI processing
- **OpenAI** - Cloud-powered enhanced capabilities  
- **Intelligent fallback** between providers
- No vendor lock-in

### 🔧 **Production-Ready**
- Kubernetes-native deployment with Helm charts
- Supports all major K8s distributions (EKS, GKE, AKS, OpenShift, Rancher)
- Horizontal scaling and high availability
- Built-in monitoring and observability

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   React UI      │    │   Go Backend     │    │   Kubernetes    │
│   (Dashboard)   │◄──►│   Microservices  │◄──►│   Cluster       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        
                                ▼                        
                       ┌──────────────────┐              
                       │   AI Processing  │              
                       │  Ollama/OpenAI   │              
                       └──────────────────┘              
```

**🏛️ Technical Stack:**
- **Backend:** Go microservices with Gin framework
- **Frontend:** React + TypeScript + Tailwind CSS  
- **Database:** PostgreSQL (audit logs), Redis (caching)
- **AI:** Ollama (local), OpenAI (optional cloud)
- **Deployment:** Helm charts, Docker containers
- **K8s Integration:** Native client-go library

## 🛠️ Quick Start

### Prerequisites

- **Kubernetes cluster** with 4GB+ available RAM
- **kubectl** configured for cluster access  
- **Helm 3.x** for deployment

### 1️⃣ Deploy with Helm

```bash
# Clone the repository
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Deploy to Kubernetes (includes Ollama)
helm install kubechat ./chart \
  --create-namespace \
  --namespace kubechat

# Wait for deployment (Ollama model download takes ~5 minutes)
kubectl get pods -n kubechat -w
```

### 2️⃣ Access KubeChat

```bash
# Port forward to access locally
kubectl port-forward svc/kubechat 8080:8080 -n kubechat

# Open in your browser
open http://localhost:8080
```

### 3️⃣ Start Chatting!

Try these example queries:
```
"Show me all pods in the default namespace"
"Which nodes are running in my cluster?"
"Display logs for pods with errors"
"Get resource usage for the kube-system namespace"
```

## 🐳 Development Setup

For local development without full Kubernetes deployment:

```bash
# Prerequisites: Docker Desktop with Kubernetes enabled
# Install Ollama locally for AI processing
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull llama2

# Clone and setup
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Quick development setup
make quickstart

# Run development servers (separate terminals)
make dev-backend    # Go API server with hot reload
make dev-frontend   # React development server
```

## ⚙️ Configuration Options

### Multi-LLM Setup

```bash
# Use external Ollama instance
helm install kubechat ./chart \
  --set llm.ollama.deploy=false \
  --set llm.ollama.url="http://your-ollama-service:11434"

# Enable OpenAI (requires API key)
helm install kubechat ./chart \
  --set llm.provider=openai \
  --set llm.openai.enabled=true \
  --set llm.openai.apiKey="your-api-key"
```

### Production Deployment

```bash
# Production setup with ingress and resource optimization
helm install kubechat ./chart \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=kubechat.yourcompany.com \
  --set resources.requests.memory="512Mi" \
  --set llm.ollama.resources.requests.memory="4Gi" \
  --namespace kubechat --create-namespace
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | AI provider (`ollama` or `openai`) | `ollama` |
| `OLLAMA_URL` | Ollama service URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | Ollama model name | `llama2` |
| `OPENAI_API_KEY` | OpenAI API key | `""` |
| `LOG_LEVEL` | Logging level | `info` |

## 🔒 Security & Compliance

KubeChat is built with **security-first principles** for enterprise environments:

- ✅ **Air-gap ready** - Complete offline operation with Ollama
- ✅ **RBAC integration** - Respects existing Kubernetes permissions  
- ✅ **Audit logging** - Complete command and query trails
- ✅ **Input validation** - Prevents injection attacks and malicious prompts
- ✅ **Encrypted storage** - Audit logs and sensitive data protection
- ✅ **Zero external calls** - No data leaves your cluster (with Ollama)

### Compliance Features

- **SOX, HIPAA, SOC 2** audit trail support
- **Complete operational history** with user attribution
- **Command preview** and approval workflows  
- **Role-based access** control integration
- **Data sovereignty** with local AI processing

## 📊 Monitoring & Observability

Built-in monitoring capabilities:

```bash
# Health check endpoints
curl http://localhost:8080/api/health
curl http://localhost:8080/api/audit?limit=10

# Application logs
kubectl logs -f deployment/kubechat -n kubechat

# Ollama AI service logs
kubectl logs -f deployment/kubechat-ollama -n kubechat
```

## 🤝 Contributing

We welcome contributions! KubeChat thrives on community input.

### Getting Started
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and test thoroughly
4. Submit a pull request with a clear description

### Development Guidelines
- **Backend:** Go 1.21+, follow effective Go practices
- **Frontend:** React + TypeScript, component-based architecture  
- **Testing:** Include unit tests for new features
- **Documentation:** Update docs for user-facing changes

### Code of Conduct
We maintain a welcoming, inclusive community. Please review our [Code of Conduct](CODE_OF_CONDUCT.md).

## 📋 Roadmap

**🎯 Current Focus (MVP)**
- ✅ Natural language to kubectl translation
- ✅ Real-time cluster monitoring dashboard
- ✅ Multi-LLM support (Ollama + OpenAI)
- ✅ Enterprise-grade audit logging
- 🔄 Advanced RBAC integration

**🚀 Coming Soon**
- Multi-cluster management capabilities
- Custom dashboard creation and sharing
- Advanced troubleshooting workflows
- Performance optimization recommendations
- Enhanced compliance reporting

## 📄 License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

## 🆘 Support & Community

- **📚 Documentation:** [docs/](docs/)
- **🐛 Bug Reports:** [GitHub Issues](https://github.com/pramodksahoo/kubechat/issues)
- **💬 Discussions:** [GitHub Discussions](https://github.com/pramodksahoo/kubechat/discussions)
- **🌐 Website:** Coming soon
- **📧 Contact:** [hello@kubechat.dev](mailto:hello@kubechat.dev)

## ⭐ Show Your Support

If KubeChat helps simplify your Kubernetes operations, please consider:
- ⭐ **Starring** the repository
- 🐛 **Reporting issues** you encounter
- 🤝 **Contributing** code or documentation
- 📢 **Sharing** with your DevOps community

---

## 🎯 Quick Summary

**For Kubernetes Administrators:**
```bash
helm install kubechat ./chart --namespace kubechat --create-namespace
kubectl port-forward svc/kubechat 8080:8080 -n kubechat
# Visit http://localhost:8080 and start chatting with your cluster
```

**For Developers:**
```bash
make quickstart && make dev  # Start developing in minutes
```

**Key Features:**
- 🤖 **Natural Language** → kubectl commands
- 🔒 **Air-gap Ready** with local Ollama AI  
- 📊 **Real-time Dashboard** with cluster monitoring
- ✅ **Enterprise Security** and compliance logging
- 🌐 **Multi-user** collaborative troubleshooting

---

**Built with ❤️ for the Kubernetes community**

*Making Kubernetes accessible through the power of natural language AI while maintaining enterprise-grade security and compliance.*
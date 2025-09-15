# KubeChat System Architecture Overview

## Introduction

KubeChat is a natural language Kubernetes management platform designed with enterprise-grade security, compliance, and operational excellence. This document provides a comprehensive overview of the system architecture, component interactions, and design decisions.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           KubeChat Platform                        │
│                                                                     │
│ ┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐ │
│ │   React Web     │    │   Go Backend     │    │   Kubernetes    │ │
│ │   Frontend      │◄──►│   Services       │◄──►│   Cluster       │ │
│ │                 │    │                  │    │                 │ │
│ │ • Dashboard     │    │ • Auth Service   │    │ • RBAC          │ │
│ │ • Chat UI       │    │ • Query Service  │    │ • Audit Logs    │ │
│ │ • Audit Trail   │    │ • K8s Client     │    │ • Workloads     │ │
│ │ • Real-time     │    │ • Audit Service  │    │ • Resources     │ │
│ └─────────────────┘    └──────────────────┘    └─────────────────┘ │
│          │                       │                       │         │
│          │              ┌──────────────────┐             │         │
│          └─────────────►│   Data Layer     │◄────────────┘         │
│                         │                  │                       │
│                         │ • PostgreSQL 16+ │                       │
│                         │ • Redis 7.4+     │                       │
│                         │ • Audit Store    │                       │
│                         └──────────────────┘                       │
│                                  │                                  │
│                         ┌──────────────────┐                       │
│                         │   AI Processing  │                       │
│                         │                  │                       │
│                         │ • Ollama (Local) │                       │
│                         │ • OpenAI (Cloud) │                       │
│                         │ • Safety Engine  │                       │
│                         └──────────────────┘                       │
└─────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Frontend Layer (React/TypeScript)

**Technology Stack:**
- React 18.3+ with TypeScript 5.6+
- Next.js 14+ for SSR and routing
- Tailwind CSS 3.4+ for styling
- Headless UI 2.1+ for accessible components

**Key Components:**
- **Dashboard**: Real-time cluster monitoring and health visualization
- **Chat Interface**: Natural language query input with command preview
- **Audit Trail Viewer**: Compliance-ready audit log interface
- **User Management**: RBAC integration and session management
- **Settings Panel**: Configuration and preferences management

**Architecture Patterns:**
- Component-based architecture with shared UI library
- State management using Zustand for predictable state updates
- Service layer abstraction for API communication
- WebSocket integration for real-time updates

### 2. Backend Layer (Go Microservices)

**Technology Stack:**
- Go 1.23+ with Gin 1.10+ web framework
- client-go v0.30+ for Kubernetes integration
- PostgreSQL 16+ with GORM for ORM
- Redis 7.4+ for caching and sessions

**Microservices Architecture:**

#### Auth Service
```go
// Handles authentication, authorization, and session management
type AuthService interface {
    Authenticate(ctx context.Context, credentials LoginRequest) (*AuthResult, error)
    ValidateSession(ctx context.Context, token string) (*User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
    RevokeSession(ctx context.Context, sessionID string) error
}
```

#### Query Service
```go
// Processes natural language queries and generates kubectl commands
type QueryService interface {
    ProcessQuery(ctx context.Context, req ProcessQueryRequest) (*QueryResult, error)
    ValidateCommand(ctx context.Context, command string) (*SafetyAssessment, error)
    ExecuteCommand(ctx context.Context, req ExecuteCommandRequest) (*ExecutionResult, error)
}
```

#### Kubernetes Client Service
```go
// Interfaces with Kubernetes clusters using client-go
type KubernetesService interface {
    GetClusterInfo(ctx context.Context) (*ClusterInfo, error)
    ExecuteKubectl(ctx context.Context, command KubectlCommand) (*KubectlResult, error)
    WatchResources(ctx context.Context, resourceType string) (<-chan ResourceEvent, error)
    ValidateRBAC(ctx context.Context, user User, action string) (bool, error)
}
```

#### Audit Service
```go
// Manages comprehensive audit trails for compliance
type AuditService interface {
    LogQuery(ctx context.Context, entry QueryAuditEntry) error
    LogExecution(ctx context.Context, entry ExecutionAuditEntry) error
    GetAuditTrail(ctx context.Context, filters AuditFilters) ([]*AuditEntry, error)
    ValidateIntegrity(ctx context.Context, startTime, endTime time.Time) (*IntegrityReport, error)
}
```

### 3. Data Layer

#### PostgreSQL Database Schema

**Users and Authentication:**
```sql
-- User management with RBAC integration
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    roles TEXT[] NOT NULL DEFAULT '{}',
    clusters TEXT[] NOT NULL DEFAULT '{}',
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Session management for secure authentication
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Immutable Audit Trail:**
```sql
-- Cryptographically secure audit logging
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES user_sessions(id),
    query_text TEXT NOT NULL,
    generated_command TEXT NOT NULL,
    safety_level VARCHAR(20) NOT NULL,
    execution_result JSONB,
    execution_status VARCHAR(20) NOT NULL,
    cluster_context VARCHAR(255),
    namespace_context VARCHAR(255),
    timestamp TIMESTAMP DEFAULT NOW() NOT NULL,
    ip_address INET,
    user_agent TEXT,
    -- Cryptographic integrity protection
    checksum VARCHAR(64) NOT NULL,
    previous_checksum VARCHAR(64),
    -- Compliance metadata
    compliance_tags JSONB DEFAULT '{}'
);

-- Ensure immutability
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit logs are immutable and cannot be modified';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_immutable_trigger
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();
```

#### Redis Caching Strategy

**Session Storage:**
```
sessions:{session_id} -> {user_data, permissions, expires}
```

**Query Caching:**
```
query_cache:{query_hash} -> {command, safety_level, cached_at}
```

**Cluster State Cache:**
```
cluster:{cluster_id}:resources -> {resource_state, last_updated}
```

### 4. AI Processing Layer

#### Ollama Integration (Local AI)
```go
type OllamaClient struct {
    baseURL string
    model   string
    timeout time.Duration
}

func (c *OllamaClient) GenerateCommand(ctx context.Context, query string) (*AIResponse, error) {
    prompt := c.buildPrompt(query)
    response, err := c.sendRequest(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("ollama request failed: %w", err)
    }
    return c.parseResponse(response), nil
}
```

#### OpenAI Integration (Cloud AI)
```go
type OpenAIClient struct {
    apiKey  string
    model   string
    baseURL string
}

func (c *OpenAIClient) GenerateCommand(ctx context.Context, query string) (*AIResponse, error) {
    // Circuit breaker pattern for external API calls
    return c.circuitBreaker.Execute(func() (*AIResponse, error) {
        return c.makeAPICall(ctx, query)
    })
}
```

#### Safety Classification Engine
```go
type SafetyClassifier struct {
    rules []SafetyRule
}

func (s *SafetyClassifier) ClassifyCommand(command string) SafetyLevel {
    for _, rule := range s.rules {
        if rule.Matches(command) {
            return rule.Level
        }
    }
    return SafetyLevelSafe
}

const (
    SafetyLevelSafe      = "safe"      // Read operations, harmless commands
    SafetyLevelWarning   = "warning"   // Write operations requiring caution
    SafetyLevelDangerous = "dangerous" // Destructive operations requiring approval
)
```

## Security Architecture

### Authentication & Authorization Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   User      │    │  Frontend   │    │  Auth API   │    │ Kubernetes  │
│  Browser    │    │   (React)   │    │   (Go)      │    │  Cluster    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                  │                  │                  │
       │ 1. Login         │                  │                  │
       ├─────────────────►│                  │                  │
       │                  │ 2. Auth Request  │                  │
       │                  ├─────────────────►│                  │
       │                  │                  │ 3. RBAC Check    │
       │                  │                  ├─────────────────►│
       │                  │                  │ 4. Permissions   │
       │                  │                  │◄─────────────────┤
       │                  │ 5. JWT Token     │                  │
       │                  │◄─────────────────┤                  │
       │ 6. Auth Success  │                  │                  │
       │◄─────────────────┤                  │                  │
```

### RBAC Integration

KubeChat respects and integrates with Kubernetes RBAC:

```yaml
# Example ClusterRole for KubeChat users
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubechat-user
rules:
  # Read-only access to common resources
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps", "secrets"]
    verbs: ["get", "list", "watch"]
  # Namespace-specific access
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
    resourceNames: [] # Controlled by namespace
```

### Audit Trail Integrity

**Cryptographic Chain Validation:**
```go
func (s *AuditService) ValidateChain(entries []*AuditEntry) error {
    for i, entry := range entries {
        expectedChecksum := s.calculateChecksum(entry)
        if entry.Checksum != expectedChecksum {
            return fmt.Errorf("integrity violation at entry %d", i)
        }

        if i > 0 && entry.PreviousChecksum != entries[i-1].Checksum {
            return fmt.Errorf("chain break between entries %d and %d", i-1, i)
        }
    }
    return nil
}
```

## Container-First Development Architecture

### Development Environment

**Container Orchestration:**
```yaml
# Development deployment with hot-reload
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubechat-dev-api
spec:
  template:
    spec:
      containers:
      - name: api
        image: kubechat-api:dev
        env:
        - name: GIN_MODE
          value: "debug"
        - name: HOT_RELOAD
          value: "true"
        volumeMounts:
        - name: source-code
          mountPath: /app/src
      volumes:
      - name: source-code
        hostPath:
          path: ./apps/api
```

**Development Workflow:**
1. Code changes on host machine
2. Hot-reload in development containers
3. Kubernetes deployment updates
4. Real-time testing in cluster environment

### Build and Deployment Pipeline

**Multi-stage Dockerfile (API):**
```dockerfile
# Development stage with hot-reload
FROM golang:1.23-alpine AS development
RUN go install github.com/cosmtrek/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
CMD ["air"]

# Production stage - minimal and secure
FROM alpine:latest AS production
RUN adduser -D -s /bin/sh appuser
COPY --from=build /app/server /server
USER appuser
EXPOSE 8080
CMD ["/server"]
```

## Extension Points and Integration

### AI Provider Interface

**Plugin Architecture:**
```go
type AIProvider interface {
    Name() string
    GenerateCommand(ctx context.Context, query string) (*AIResponse, error)
    ValidateConnection(ctx context.Context) error
    GetCapabilities() AICapabilities
}

// Register new AI providers
func RegisterAIProvider(name string, provider AIProvider) {
    aiProviders[name] = provider
}
```

### Webhook Integration

**Event Webhook System:**
```go
type WebhookManager struct {
    hooks map[string][]WebhookEndpoint
}

func (w *WebhookManager) TriggerEvent(event string, data interface{}) error {
    for _, hook := range w.hooks[event] {
        go w.sendWebhook(hook, data)
    }
    return nil
}
```

### Custom Dashboard Widgets

**Widget Framework:**
```typescript
interface DashboardWidget {
  id: string;
  title: string;
  component: React.ComponentType<WidgetProps>;
  configuration: WidgetConfig;
  dataSource: DataSourceConfig;
}

// Register custom widgets
export function registerWidget(widget: DashboardWidget): void {
  widgetRegistry.set(widget.id, widget);
}
```

## Technology Stack Decisions

### Frontend Technology Choices

**React + TypeScript:**
- **Rationale**: Strong ecosystem, excellent TypeScript support, component reusability
- **Alternative Considered**: Vue.js, Angular
- **Decision Factors**: Team expertise, enterprise adoption, security track record

**Tailwind CSS:**
- **Rationale**: Utility-first approach, consistent design system, excellent performance
- **Alternative Considered**: Styled Components, Material-UI
- **Decision Factors**: Bundle size optimization, design system flexibility

### Backend Technology Choices

**Go with Gin Framework:**
- **Rationale**: Excellent Kubernetes client library support, performance, memory efficiency
- **Alternative Considered**: Node.js, Python, Rust
- **Decision Factors**: Kubernetes ecosystem fit, static compilation, security

**PostgreSQL:**
- **Rationale**: ACID compliance for audit trails, JSON support, enterprise features
- **Alternative Considered**: MongoDB, ClickHouse
- **Decision Factors**: Compliance requirements, data integrity, SQL compatibility

### AI Integration Decisions

**Ollama (Primary):**
- **Rationale**: Local processing, air-gap compatibility, privacy, cost control
- **Alternative Considered**: Exclusively cloud-based solutions
- **Decision Factors**: Enterprise security requirements, data sovereignty

**OpenAI (Optional):**
- **Rationale**: Advanced capabilities, well-tested models, good API design
- **Alternative Considered**: Anthropic Claude, Google Vertex AI
- **Decision Factors**: API stability, model performance, integration complexity

## Performance Considerations

### Frontend Optimization

**Code Splitting:**
```typescript
// Lazy loading for route-based code splitting
const Dashboard = lazy(() => import('./components/Dashboard'));
const AuditTrail = lazy(() => import('./components/AuditTrail'));
```

**State Management Optimization:**
```typescript
// Optimized Zustand store with selectors
const useClusterData = create<ClusterState>((set, get) => ({
  clusters: [],
  selectedCluster: null,
  updateCluster: (cluster) => set(
    produce((state) => {
      const index = state.clusters.findIndex(c => c.id === cluster.id);
      if (index >= 0) state.clusters[index] = cluster;
    })
  ),
}));
```

### Backend Performance

**Database Query Optimization:**
```sql
-- Optimized audit query with proper indexing
CREATE INDEX CONCURRENTLY idx_audit_logs_user_timestamp
ON audit_logs(user_id, timestamp DESC);

CREATE INDEX CONCURRENTLY idx_audit_logs_cluster_namespace
ON audit_logs(cluster_context, namespace_context);
```

**Caching Strategy:**
```go
// Multi-level caching with TTL
type CacheManager struct {
    memory cache.Cache
    redis  *redis.Client
}

func (c *CacheManager) Get(key string) (interface{}, bool) {
    // L1: Memory cache
    if val, found := c.memory.Get(key); found {
        return val, true
    }

    // L2: Redis cache
    if val := c.redis.Get(key); val.Err() == nil {
        c.memory.Set(key, val.Val(), cache.DefaultExpiration)
        return val.Val(), true
    }

    return nil, false
}
```

## Monitoring and Observability

### Application Metrics

**Prometheus Integration:**
```go
var (
    queriesProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "kubechat_queries_processed_total",
            Help: "Total number of queries processed",
        },
        []string{"user", "safety_level", "status"},
    )

    queryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "kubechat_query_duration_seconds",
            Help: "Time spent processing queries",
        },
        []string{"type"},
    )
)
```

### Health Checks

**Service Health Endpoints:**
```go
func (s *Server) healthCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        health := HealthStatus{
            Status:    "healthy",
            Timestamp: time.Now(),
            Services: map[string]ServiceHealth{
                "database":   s.checkDatabase(),
                "kubernetes": s.checkKubernetes(),
                "ai":        s.checkAI(),
            },
        }

        c.JSON(http.StatusOK, health)
    }
}
```

## Conclusion

KubeChat's architecture prioritizes security, compliance, and operational excellence while maintaining developer productivity through container-first development practices. The modular design enables enterprise adoption while supporting community extensibility through well-defined interfaces and extension points.

Key architectural strengths:
- **Security-first design** with comprehensive audit trails
- **Air-gap capability** through local AI processing
- **Enterprise compliance** with immutable audit logs
- **Developer experience** through container-first development
- **Extensibility** through plugin interfaces and webhooks
- **Performance** through intelligent caching and optimization

This architecture supports KubeChat's mission of making Kubernetes accessible through natural language while maintaining the security and operational standards required for enterprise environments.
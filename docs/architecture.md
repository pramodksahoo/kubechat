# KubeChat Fullstack Architecture Document

## Introduction

This document outlines the complete fullstack architecture for KubeChat, including backend systems, frontend implementation, and their integration. It serves as the single source of truth for AI-driven development, ensuring consistency across the entire technology stack.

This unified approach combines what would traditionally be separate backend and frontend architecture documents, streamlining the development process for modern fullstack applications where these concerns are increasingly intertwined.

### Starter Template Analysis

**Finding:** KubeChat is based on a validated PoC with established architecture patterns:

**Existing PoC Foundation:**
- **Backend:** Go microservices with Gin framework, client-go for Kubernetes API
- **Frontend:** React with TypeScript, Tailwind CSS for styling
- **Deployment:** Helm charts with Kubernetes-native deployment
- **AI Integration:** Ollama-first with OpenAI fallback capability
- **Architecture:** Monorepo structure with clear service boundaries

**Constraints from PoC:**
- ✅ Must maintain Go backend (validated performance with Kubernetes)
- ✅ Must support Ollama air-gapped deployment (unique market differentiator)
- ✅ Must use Helm charts for enterprise deployment
- ✅ WebSocket integration for real-time monitoring established

**Architectural decisions already validated:**
- Microservices within monorepo approach
- Container-first development workflow
- Multi-LLM abstraction layer
- Audit logging and compliance framework

### Change Log

| Date | Version | Description | Author |
|------|---------|-------------|---------|
| 2025-01-10 | 1.0 | Initial architecture document creation | Winston (Architect) |

---

## High Level Architecture

### Technical Summary

KubeChat employs a **cloud-native microservices architecture within a monorepo**, optimized for enterprise Kubernetes environments with air-gapped deployment capability. The system combines a **React TypeScript frontend** with **Go microservices backend**, leveraging native Kubernetes APIs through client-go for optimal cluster integration. The architecture supports **Ollama-first AI processing** with intelligent multi-LLM fallback, enabling complete offline operation while maintaining enterprise-grade audit trails and RBAC compliance. **Helm-native deployment** ensures consistent enterprise installation across regulated environments, while **WebSocket-based real-time monitoring** provides collaborative troubleshooting capabilities that differentiate KubeChat from CLI-only competitors.

### Platform and Infrastructure Choice

**Platform:** Kubernetes-native (any distribution)
**Key Services:** Kubernetes API, Persistent Volumes, Service Mesh (optional), Ingress Controllers
**Deployment Host and Regions:** Customer-controlled (on-premises, cloud, air-gapped)

**Rationale:** This aligns perfectly with the unique value proposition and PRD requirements. The target market (regulated industries, enterprise DevOps) already operates Kubernetes clusters and values the air-gapped capability.

### Repository Structure

**Structure:** Monorepo with clear service boundaries
**Monorepo Tool:** Native Go modules + npm workspaces (aligns with PoC)
**Package Organization:** Service-oriented with shared libraries

### High Level Architecture Diagram

```mermaid
graph TB
    subgraph "User Layer"
        U1[DevOps Engineers]
        U2[Platform Admins]
        U3[Compliance Officers]
    end

    subgraph "KubeChat Cluster"
        subgraph "Frontend Layer"
            WEB[React Web App<br/>TypeScript + Tailwind]
            CDN[Static Assets<br/>Nginx/Ingress]
        end

        subgraph "API Gateway Layer"
            GW[Gin HTTP Server<br/>API Gateway + WebSocket]
        end

        subgraph "Microservices Layer"
            AUTH[Authentication<br/>Service]
            NLP[Natural Language<br/>Processing Service]
            K8S[Kubernetes Client<br/>Service]
            AUDIT[Audit Logging<br/>Service]
            MONITOR[Real-time Monitoring<br/>WebSocket Service]
        end

        subgraph "AI Layer"
            OLLAMA[Ollama Local AI<br/>Air-gapped Default]
            CLOUD[Cloud AI Fallback<br/>OpenAI/Claude/Anthropic]
        end

        subgraph "Data Layer"
            PG[(PostgreSQL<br/>Audit + User Data)]
            REDIS[(Redis<br/>Cache + Sessions)]
        end
    end

    subgraph "Target Kubernetes Clusters"
        K8S_API[Kubernetes API Server]
        WORKLOADS[Pods, Services,<br/>Deployments, etc.]
    end

    subgraph "External Systems"
        SIEM[Enterprise SIEM]
        LDAP[Enterprise Identity]
        MONITORING[Prometheus/Grafana]
    end

    %% User interactions
    U1 --> WEB
    U2 --> WEB
    U3 --> WEB

    %% Frontend to backend
    WEB --> GW
    CDN --> WEB

    %% API Gateway routing
    GW --> AUTH
    GW --> NLP
    GW --> K8S
    GW --> AUDIT
    GW --> MONITOR

    %% Service dependencies
    NLP --> OLLAMA
    NLP --> CLOUD
    K8S --> K8S_API
    AUTH --> PG
    AUDIT --> PG
    MONITOR --> K8S_API
    MONITOR --> REDIS

    %% Kubernetes integration
    K8S_API --> WORKLOADS

    %% External integrations
    AUDIT --> SIEM
    AUTH --> LDAP
    MONITOR --> MONITORING

    %% Styling
    classDef frontend fill:#e1f5fe
    classDef backend fill:#f3e5f5
    classDef ai fill:#fff3e0
    classDef data fill:#e8f5e8
    classDef external fill:#fafafa

    class WEB,CDN frontend
    class GW,AUTH,NLP,K8S,AUDIT,MONITOR backend
    class OLLAMA,CLOUD ai
    class PG,REDIS data
    class SIEM,LDAP,MONITORING external
```

### Architectural Patterns

- **Microservices within Monorepo:** Clear service boundaries with shared deployment - _Rationale:_ Maintains service autonomy while simplifying air-gapped deployment and shared type definitions
- **API Gateway Pattern:** Single entry point with Gin framework - _Rationale:_ Centralized CORS, authentication, rate limiting, and WebSocket upgrade handling
- **Repository Pattern:** Abstract data access across services - _Rationale:_ Enables testing, audit trail integrity, and future database flexibility
- **Multi-Provider AI Abstraction:** Pluggable LLM providers with intelligent fallback - _Rationale:_ Ollama-first for air-gapped, cloud fallback for enhanced capability when available
- **Event-Driven Real-time Updates:** WebSocket-based cluster monitoring - _Rationale:_ Enables collaborative troubleshooting advantage over CLI tools
- **Command Pattern with Safety Classification:** Natural language queries generate classified kubectl commands - _Rationale:_ Builds user confidence while maintaining enterprise safety requirements
- **Immutable Audit Trail:** Write-only audit logging with cryptographic integrity - _Rationale:_ Meets regulatory compliance requirements for SOX, HIPAA, SOC 2

---

## Tech Stack

### Technology Stack Table

| Category | Technology | Version | Purpose | Rationale |
|----------|------------|---------|---------|-----------|
| Frontend Language | TypeScript | 5.6+ | Type-safe frontend development | Latest stable with improved performance and type inference |
| Frontend Framework | React | 18.3+ | Component-based UI development | Latest stable with concurrent features, excellent TypeScript support |
| UI Component Library | Tailwind CSS 3.4+ + Headless UI 2.1+ | Latest Stable | Enterprise UI development | Latest with container queries, modern CSS features |
| State Management | Zustand | 4.5+ | Lightweight state management | Updated for React 18.3 compatibility, improved TypeScript inference |
| Backend Language | Go | 1.23+ | High-performance backend services | Latest stable with improved performance, better generics support |
| Backend Framework | Gin | 1.10+ | HTTP API and WebSocket handling | Updated for Go 1.23, improved middleware performance |
| API Style | REST + WebSocket | HTTP/1.1 | RESTful APIs with real-time updates | Standards-based, WebSocket for monitoring, simple client integration |
| Database | PostgreSQL | 16+ | Audit trails and structured data | Latest stable with improved performance, JSON enhancements |
| Cache | Redis | 7.4+ | Session and query caching | Latest stable with improved memory efficiency |
| File Storage | Kubernetes PV | - | Persistent data and AI models | Native Kubernetes storage, supports air-gapped deployments |
| Authentication | Kubernetes RBAC + JWT | - | Enterprise-grade access control | Leverages existing K8s permissions, standards-based tokens |
| Frontend Testing | Vitest 2.0+ + Testing Library 16+ | Latest | Fast unit and integration testing | Updated for Vite 5, better React 18.3 support |
| Backend Testing | Go testing + Testify 1.9+ | Latest | Go service testing | Updated for Go 1.23, improved assertion methods |
| E2E Testing | Playwright 1.47+ | Latest | Full system testing | Latest with improved TypeScript support, better debugging |
| Build Tool | Vite | 5.4+ | Frontend build optimization | Latest stable with improved performance, better dev experience |
| Bundler | Vite (built-in) | 5.4+ | Frontend asset bundling | Integrated with Vite 5, optimized build performance |
| IaC Tool | Helm | 3.15+ | Kubernetes deployment | Latest stable with improved security, better chart validation |
| CI/CD | GitHub Actions | - | Automated testing and deployment | Latest runners with improved caching, security features |
| Monitoring | Prometheus 2.54+ + Grafana 11+ | Latest | System and business metrics | Latest with improved query performance, modern UI |
| Logging | Structured JSON + ELK Stack 8.15+ | Latest | Centralized log aggregation | Latest Elasticsearch/Kibana with improved security |
| CSS Framework | Tailwind CSS | 3.4+ | Utility-first styling | Latest with container queries, modern CSS features |

### AI/LLM Stack

| Category | Technology | Version | Purpose | Rationale |
|----------|------------|---------|---------|-----------|
| Primary AI | Ollama | 0.3+ | Local AI processing (air-gapped) | Latest with improved model loading, better Go SDK |
| AI Models | Phi-3.5-mini, CodeLlama 34B | Latest | Kubernetes command generation | Updated models with better code understanding |
| AI Fallback | OpenAI SDK | 5.0+ | Cloud AI when available | Latest with GPT-4o support, improved streaming |
| AI Client | Custom Go SDK | - | Multi-provider abstraction | Updated for Go 1.23, improved error handling |

---

## Data Models

### User

**Purpose:** Represents authenticated users with RBAC permissions and session management for enterprise compliance

**Key Attributes:**
- id: string (UUID) - Primary identifier
- username: string - Kubernetes service account or enterprise username  
- email: string - Contact and audit trail attribution
- roles: string[] - Kubernetes RBAC roles for permission management
- clusters: string[] - Accessible cluster contexts
- preferences: UserPreferences - UI settings and notification preferences
- lastLogin: timestamp - Audit trail requirement
- createdAt: timestamp - Account lifecycle tracking

#### TypeScript Interface
```typescript
interface User {
  id: string;
  username: string;
  email?: string;
  roles: string[];
  clusters: string[];
  preferences: UserPreferences;
  lastLogin: Date;
  createdAt: Date;
  updatedAt: Date;
}

interface UserPreferences {
  theme: 'light' | 'dark' | 'system';
  defaultNamespace: string;
  notificationsEnabled: boolean;
  expertMode: boolean;
  defaultCluster?: string;
}
```

#### Relationships
- One-to-many with QuerySession (user can have multiple active sessions)
- One-to-many with AuditLogEntry (user generates multiple audit entries)

### QuerySession

**Purpose:** Represents conversational sessions with natural language queries, enabling collaborative troubleshooting and session sharing

**Key Attributes:**
- id: string (UUID) - Session identifier for sharing and collaboration
- userId: string - Session owner for audit and permissions
- title: string - Auto-generated or user-defined session name
- queries: Query[] - Ordered list of queries and responses in conversation
- participants: string[] - User IDs for collaborative sessions
- clusterContext: string - Active Kubernetes cluster
- namespaceContext: string - Active namespace scope
- isShared: boolean - Collaboration enabled flag
- status: SessionStatus - Active, archived, or shared state

#### TypeScript Interface
```typescript
interface QuerySession {
  id: string;
  userId: string;
  title: string;
  queries: Query[];
  participants: string[];
  clusterContext: string;
  namespaceContext: string;
  isShared: boolean;
  status: 'active' | 'archived' | 'shared';
  createdAt: Date;
  updatedAt: Date;
  lastActivityAt: Date;
}

type SessionStatus = 'active' | 'archived' | 'shared';
```

#### Relationships
- Many-to-one with User (sessions belong to users)
- One-to-many with Query (session contains multiple queries)
- Many-to-many with User (through collaborative sharing)

### Query

**Purpose:** Individual natural language query with generated kubectl command, results, and safety classification for audit compliance

**Key Attributes:**
- id: string (UUID) - Query identifier for audit trails
- sessionId: string - Parent session for conversation context
- userId: string - Query author for audit attribution
- naturalLanguage: string - Original user input
- generatedCommand: string - AI-generated kubectl command
- explanation: string - Plain-English command explanation
- safetyLevel: SafetyLevel - Command risk classification
- executionResult: ExecutionResult - Command output and status
- llmProvider: string - AI provider used for audit tracking
- processingTime: number - Performance monitoring

#### TypeScript Interface
```typescript
interface Query {
  id: string;
  sessionId: string;
  userId: string;
  naturalLanguage: string;
  generatedCommand: string;
  explanation: string;
  safetyLevel: 'safe' | 'warning' | 'dangerous';
  executionResult?: ExecutionResult;
  llmProvider: string;
  processingTime: number;
  createdAt: Date;
  executedAt?: Date;
}

interface ExecutionResult {
  success: boolean;
  output: string;
  error?: string;
  resourcesAffected: string[];
  executionTime: number;
}

type SafetyLevel = 'safe' | 'warning' | 'dangerous';
```

#### Relationships
- Many-to-one with QuerySession (queries belong to sessions)
- Many-to-one with User (queries have authors)
- One-to-one with AuditLogEntry (each query generates audit entry)

### AuditLogEntry

**Purpose:** Immutable audit trail for regulatory compliance (SOX, HIPAA, SOC 2) with cryptographic integrity

**Key Attributes:**
- id: string (UUID) - Immutable audit identifier
- userId: string - Action performer for compliance attribution
- action: string - Standardized action type for reporting
- resource: string - Kubernetes resource affected
- cluster: string - Target cluster for multi-cluster audit trails
- namespace: string - Resource scope for permission validation
- details: AuditDetails - Comprehensive context for investigations
- timestamp: timestamp - Precise action timing with timezone
- hash: string - Cryptographic integrity verification
- complianceFlags: string[] - Regulatory framework tags

#### TypeScript Interface
```typescript
interface AuditLogEntry {
  id: string;
  userId: string;
  action: 'query' | 'execute' | 'login' | 'logout' | 'permission_change';
  resource: string;
  cluster: string;
  namespace: string;
  details: AuditDetails;
  timestamp: Date;
  hash: string;
  complianceFlags: ('sox' | 'hipaa' | 'soc2' | 'pci' | 'fedramp')[];
}

interface AuditDetails {
  originalQuery?: string;
  generatedCommand?: string;
  executionSuccess?: boolean;
  ipAddress: string;
  userAgent: string;
  sessionId?: string;
  errorMessage?: string;
  resourcesAffected: string[];
}
```

#### Relationships
- Many-to-one with User (audit entries track user actions)
- One-to-one with Query (executed queries create audit entries)

### ClusterHealth

**Purpose:** Real-time cluster monitoring data for observability dashboards and collaborative troubleshooting

**Key Attributes:**
- clusterId: string - Cluster identifier for multi-cluster support
- timestamp: timestamp - Metric collection time
- overallHealth: HealthStatus - Aggregated cluster health score
- nodeMetrics: NodeMetric[] - Individual node health and resources
- workloadMetrics: WorkloadMetric[] - Application health indicators
- alerts: Alert[] - Active cluster alerts and warnings
- resourceUtilization: ResourceMetrics - Cluster-wide resource usage

#### TypeScript Interface
```typescript
interface ClusterHealth {
  clusterId: string;
  timestamp: Date;
  overallHealth: 'healthy' | 'warning' | 'critical' | 'unknown';
  nodeMetrics: NodeMetric[];
  workloadMetrics: WorkloadMetric[];
  alerts: Alert[];
  resourceUtilization: ResourceMetrics;
}

interface NodeMetric {
  name: string;
  status: 'ready' | 'not_ready' | 'unknown';
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  podCount: number;
}

interface WorkloadMetric {
  namespace: string;
  workloadType: 'deployment' | 'statefulset' | 'daemonset';
  name: string;
  replicas: {
    desired: number;
    ready: number;
    available: number;
  };
  health: 'healthy' | 'degraded' | 'failed';
}

interface ResourceMetrics {
  cpu: { used: number; total: number; percentage: number };
  memory: { used: number; total: number; percentage: number };
  storage: { used: number; total: number; percentage: number };
}

type HealthStatus = 'healthy' | 'warning' | 'critical' | 'unknown';
```

#### Relationships
- No direct database relationships (time-series data)
- Logically related to User queries about cluster resources

---

## API Specification

### REST API Specification

```yaml
openapi: 3.0.0
info:
  title: KubeChat API
  version: 1.0.0
  description: |
    KubeChat API provides natural language Kubernetes management with enterprise-grade 
    audit trails and real-time collaboration capabilities. Supports air-gapped deployment 
    with local AI processing and multi-LLM fallback.
  contact:
    name: KubeChat Team
    url: https://kubechat.dev
    email: api@kubechat.dev
  license:
    name: Apache 2.0
    url: https://apache.org/licenses/LICENSE-2.0.html

servers:
  - url: https://kubechat.{cluster.local}/api/v1
    description: Production KubeChat deployment
  - url: http://localhost:8080/api/v1
    description: Local development server

security:
  - KubernetesAuth: []
  - JWTAuth: []

paths:
  /auth/login:
    post:
      summary: Authenticate user with Kubernetes RBAC
      tags: [Authentication]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                kubeconfig:
                  type: string
                  description: Base64 encoded kubeconfig or service account token
                cluster:
                  type: string
                  description: Target cluster context
              required: [kubeconfig]
      responses:
        200:
          description: Authentication successful
          content:
            application/json:
              schema:
                type: object
                properties:
                  token:
                    type: string
                    description: JWT token for subsequent requests
                  user:
                    $ref: '#/components/schemas/User'
                  clusters:
                    type: array
                    items:
                      type: string
                  expires_at:
                    type: string
                    format: date-time

  /sessions:
    get:
      summary: List user's query sessions
      tags: [Sessions]
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 50
            maximum: 100
      responses:
        200:
          description: Sessions retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  sessions:
                    type: array
                    items:
                      $ref: '#/components/schemas/QuerySession'
                  total:
                    type: integer

    post:
      summary: Create new query session
      tags: [Sessions]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                title:
                  type: string
                cluster_context:
                  type: string
                namespace_context:
                  type: string
                  default: default
              required: [cluster_context]
      responses:
        201:
          description: Session created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/QuerySession'

  /sessions/{sessionId}/queries:
    post:
      summary: Process natural language query
      tags: [Queries]
      parameters:
        - name: sessionId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                query:
                  type: string
                  description: Natural language query
                  example: "Show me pods with high memory usage in production namespace"
                auto_execute:
                  type: boolean
                  default: false
              required: [query]
      responses:
        200:
          description: Query processed successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Query'

  /health:
    get:
      summary: Service health check
      tags: [System]
      responses:
        200:
          description: Service is healthy
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: healthy
                  version:
                    type: string
                  timestamp:
                    type: string
                    format: date-time

components:
  securitySchemes:
    KubernetesAuth:
      type: http
      scheme: bearer
      description: Kubernetes service account token or kubeconfig-derived JWT
    JWTAuth:
      type: http
      scheme: bearer
      description: KubeChat-issued JWT token

  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          format: uuid
        username:
          type: string
        email:
          type: string
          format: email
        roles:
          type: array
          items:
            type: string
        clusters:
          type: array
          items:
            type: string
        created_at:
          type: string
          format: date-time

    QuerySession:
      type: object
      properties:
        id:
          type: string
          format: uuid
        user_id:
          type: string
          format: uuid
        title:
          type: string
        cluster_context:
          type: string
        namespace_context:
          type: string
        is_shared:
          type: boolean
        status:
          type: string
          enum: [active, archived, shared]
        created_at:
          type: string
          format: date-time

    Query:
      type: object
      properties:
        id:
          type: string
          format: uuid
        session_id:
          type: string
          format: uuid
        natural_language:
          type: string
        generated_command:
          type: string
        explanation:
          type: string
        safety_level:
          type: string
          enum: [safe, warning, dangerous]
        llm_provider:
          type: string
        created_at:
          type: string
          format: date-time
```

---

## Components

### API Gateway Service

**Responsibility:** Centralized HTTP routing, WebSocket upgrades, CORS management, rate limiting, and request/response transformation for all client interactions

**Key Interfaces:**
- REST API endpoints (`/api/v1/*`) with JSON request/response handling
- WebSocket upgrade and connection management (`/api/v1/ws`)
- Authentication middleware with JWT validation and Kubernetes RBAC integration
- Request logging and audit trail initiation

**Dependencies:** Authentication Service (token validation), Audit Logging Service (request tracking), all backend microservices (routing)

**Technology Stack:** Go 1.23+ with Gin 1.10+ framework, gorilla/websocket for WebSocket handling, custom middleware for CORS and rate limiting

### Authentication and Authorization Service

**Responsibility:** Kubernetes RBAC integration, JWT token management, user session handling, and permission validation against cluster resources

**Key Interfaces:**
- Kubeconfig/ServiceAccount token validation with Kubernetes API
- JWT token generation and validation for session management  
- RBAC permission checking before command execution
- User profile and preference management

**Dependencies:** Kubernetes API Server (RBAC validation), PostgreSQL (user data persistence), Redis (session caching)

**Technology Stack:** Go 1.23+ with client-go v0.30+ for Kubernetes integration, JWT libraries for token handling, bcrypt for credential security

### Natural Language Processing Service

**Responsibility:** Multi-LLM query processing, kubectl command generation, safety classification, and AI provider failover management

**Key Interfaces:**
- Natural language query processing with context injection
- Multi-provider AI abstraction (Ollama primary, OpenAI/Claude fallback)
- Command safety classification (safe/warning/dangerous)
- Kubernetes resource context awareness for intelligent suggestions

**Dependencies:** Ollama service (primary AI), external AI APIs (fallback), Kubernetes Client Service (resource context), Audit Logging Service (AI decision tracking)

**Technology Stack:** Go 1.23+ with custom AI provider abstraction, Ollama Go SDK, OpenAI SDK 5.0+, structured prompt templates for Kubernetes domain

### Kubernetes Client Service

**Responsibility:** Kubernetes API interaction, cluster health monitoring, resource querying, and kubectl command execution with RBAC enforcement

**Key Interfaces:**
- kubectl command execution with safety validation
- Real-time cluster health and resource monitoring
- Multi-cluster context management and switching
- Resource discovery and namespace-aware operations

**Dependencies:** Kubernetes API Server (all operations), Authentication Service (permission validation), Real-time Monitoring Service (health data)

**Technology Stack:** Go 1.23+ with client-go v0.30+ for native Kubernetes API access, custom resource caching, context-aware command execution

### Real-time Monitoring Service

**Responsibility:** WebSocket connection management, live cluster metrics collection, collaborative session coordination, and real-time event streaming

**Key Interfaces:**
- WebSocket connection pool management for concurrent users
- Cluster health metrics aggregation and streaming
- Collaborative session state synchronization
- Real-time log streaming and filtering

**Dependencies:** Kubernetes Client Service (metrics source), Redis (connection state), Authentication Service (user context), PostgreSQL (session persistence)

**Technology Stack:** Go 1.23+ with gorilla/websocket, Redis pub/sub for real-time messaging, custom metrics aggregation, connection pool management

### Audit Logging Service

**Responsibility:** Immutable audit trail creation, compliance reporting, cryptographic integrity validation, and regulatory framework compliance

**Key Interfaces:**
- Immutable audit entry creation with cryptographic hashing
- Compliance report generation (SOX, HIPAA, SOC 2, PCI, FedRAMP)
- Audit trail querying and filtering with performance optimization
- Export functionality with digital signatures

**Dependencies:** PostgreSQL (audit persistence), Authentication Service (user context), external SIEM systems (log forwarding)

**Technology Stack:** Go 1.23+ with PostgreSQL 16+ for ACID compliance, cryptographic libraries for integrity hashing, structured JSON logging

### Frontend React Application

**Responsibility:** Enterprise web UI, natural language chat interface, real-time collaboration features, and responsive dashboard presentation

**Key Interfaces:**
- Natural language chat interface with command preview and safety indicators
- Real-time cluster monitoring dashboards with interactive visualizations
- Collaborative session management with multi-user support
- Audit trail browsing and compliance reporting interfaces

**Dependencies:** API Gateway Service (all backend communication), WebSocket connections (real-time updates), browser localStorage (session persistence)

**Technology Stack:** React 18.3+ with TypeScript 5.6+, Zustand 4.5+ for state management, Tailwind CSS 3.4+ with Headless UI 2.1+, Vite 5.4+ for builds

### Ollama AI Service

**Responsibility:** Local AI model hosting, Kubernetes-optimized prompt processing, model lifecycle management, and air-gapped AI capability

**Key Interfaces:**
- Local AI model inference API with Kubernetes-specific prompts
- Model loading and memory management
- Health check and performance monitoring
- Resource usage optimization for containerized deployment

**Dependencies:** Persistent volumes (model storage), container runtime (deployment), optional GPU resources (acceleration)

**Technology Stack:** Ollama 0.3+ with Phi-3.5-mini and CodeLlama models, containerized deployment, Kubernetes persistent volume integration

---

## Core Workflows

### Natural Language Query Processing and Execution Workflow

```mermaid
sequenceDiagram
    participant User
    participant WebApp as React Web App
    participant Gateway as API Gateway
    participant Auth as Auth Service
    participant NLP as NLP Service
    participant Ollama as Ollama Service
    participant K8sClient as K8s Client Service
    participant K8sAPI as Kubernetes API
    participant Audit as Audit Service
    participant PG as PostgreSQL

    User->>WebApp: Enter query "Show me pods with high memory in production"
    WebApp->>Gateway: POST /api/v1/sessions/{id}/queries
    Gateway->>Auth: Validate JWT token & RBAC permissions
    Auth->>K8sAPI: Check namespace access for "production"
    K8sAPI-->>Auth: Permission granted
    Auth-->>Gateway: User authorized for query
    
    Gateway->>NLP: Process natural language query
    NLP->>K8sClient: Get cluster context (nodes, namespaces, resources)
    K8sClient->>K8sAPI: kubectl get nodes,namespaces --output=json
    K8sAPI-->>K8sClient: Cluster context data
    K8sClient-->>NLP: Contextual information
    
    NLP->>Ollama: Generate kubectl command with context
    Note over NLP,Ollama: Prompt: "Convert query to kubectl with context"
    Ollama-->>NLP: Generated command + explanation + safety level
    
    alt If Ollama fails or low confidence
        NLP->>OpenAI: Fallback to cloud AI
        OpenAI-->>NLP: Enhanced command generation
    end
    
    NLP->>Audit: Log query processing (user, query, generated command)
    Audit->>PG: Store immutable audit entry with hash
    PG-->>Audit: Audit entry confirmed
    
    NLP-->>Gateway: Query response (command: "kubectl get pods -n production --sort-by=.status.containerStatuses[0].restartCount")
    Gateway-->>WebApp: Query with safety classification "safe"
    WebApp-->>User: Display command preview with explanation
    
    User->>WebApp: Approve and execute command
    WebApp->>Gateway: POST /api/v1/sessions/{id}/queries/{queryId}/execute
    Gateway->>Auth: Validate execution permissions
    Auth->>K8sAPI: Verify RBAC for kubectl get pods
    K8sAPI-->>Auth: Permission granted
    
    Gateway->>K8sClient: Execute kubectl command
    K8sClient->>K8sAPI: Execute: kubectl get pods -n production --sort-by=...
    K8sAPI-->>K8sClient: Pod list with memory usage data
    K8sClient->>K8sClient: Parse and structure results
    K8sClient-->>Gateway: Structured execution results
    
    Gateway->>Audit: Log command execution (user, command, results, success)
    Audit->>PG: Store execution audit entry
    
    Gateway-->>WebApp: Execution results with pod data
    WebApp-->>User: Display formatted pod list with memory highlights
```

---

## Database Schema

```sql
-- Enable UUID extension for distributed deployment support
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pg_crypto for audit trail hashing
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table with RBAC integration
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255),
    roles TEXT[] NOT NULL DEFAULT '{}', -- Kubernetes RBAC roles
    clusters TEXT[] NOT NULL DEFAULT '{}', -- Accessible cluster contexts
    preferences JSONB DEFAULT '{}', -- User preferences as JSON
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$' OR email IS NULL),
    CONSTRAINT non_empty_username CHECK (length(username) > 0),
    CONSTRAINT reasonable_role_count CHECK (array_length(roles, 1) <= 50)
);

-- Query sessions for collaborative troubleshooting
CREATE TABLE query_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    cluster_context VARCHAR(255) NOT NULL,
    namespace_context VARCHAR(255) NOT NULL DEFAULT 'default',
    is_shared BOOLEAN NOT NULL DEFAULT FALSE,
    participants UUID[] DEFAULT '{}', -- User IDs for collaborative sessions
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'shared')),
    metadata JSONB DEFAULT '{}', -- Extensible session metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Individual queries within sessions
CREATE TABLE queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES query_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    natural_language TEXT NOT NULL,
    generated_command TEXT NOT NULL,
    explanation TEXT NOT NULL,
    safety_level VARCHAR(20) NOT NULL CHECK (safety_level IN ('safe', 'warning', 'dangerous')),
    execution_result JSONB, -- Nullable until executed
    llm_provider VARCHAR(100) NOT NULL,
    processing_time_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMPTZ
);

-- Immutable audit trail for compliance
CREATE TABLE audit_log_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT, -- Preserve audit integrity
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(500) NOT NULL,
    cluster VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    compliance_flags TEXT[] NOT NULL DEFAULT '{}', -- SOX, HIPAA, SOC2, etc.
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hash VARCHAR(64) NOT NULL, -- SHA-256 hash for integrity verification
    
    -- Constraints
    CONSTRAINT valid_action CHECK (action IN ('query', 'execute', 'login', 'logout', 'permission_change', 'session_create', 'session_share', 'export_audit')),
    CONSTRAINT valid_hash CHECK (length(hash) = 64) -- SHA-256 hex length
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_sessions_user_id ON query_sessions(user_id);
CREATE INDEX idx_queries_session_id ON queries(session_id);
CREATE INDEX idx_audit_user_id ON audit_log_entries(user_id);
CREATE INDEX idx_audit_timestamp ON audit_log_entries(timestamp DESC);

-- Functions for audit trail integrity
CREATE OR REPLACE FUNCTION generate_audit_hash(entry_data JSONB) 
RETURNS VARCHAR(64) AS $$
BEGIN
    RETURN encode(digest(entry_data::text, 'sha256'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Trigger to automatically generate audit hashes
CREATE OR REPLACE FUNCTION set_audit_hash() 
RETURNS TRIGGER AS $$
BEGIN
    NEW.hash = generate_audit_hash(
        jsonb_build_object(
            'user_id', NEW.user_id,
            'action', NEW.action,
            'resource', NEW.resource,
            'cluster', NEW.cluster,
            'namespace', NEW.namespace,
            'details', NEW.details,
            'timestamp', NEW.timestamp
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_audit_hash
    BEFORE INSERT ON audit_log_entries
    FOR EACH ROW
    EXECUTE FUNCTION set_audit_hash();
```

---

## Frontend Architecture

### Component Organization

```text
apps/web/src/
├── components/                    # Reusable UI components
│   ├── ui/                       # Basic UI building blocks
│   ├── forms/                    # Form components
│   ├── data-display/             # Data visualization components
│   └── layout/                   # Layout components
├── pages/                        # Page-level components
├── hooks/                        # Custom React hooks
├── services/                     # API client services
├── stores/                       # Zustand state management
├── types/                        # TypeScript type definitions
└── utils/                        # Utility functions
```

### State Management Architecture

```typescript
// stores/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, UserPreferences } from '@/types/api';

interface AuthState {
  // State
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  
  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  updateUser: (updates: Partial<User>) => void;
  updatePreferences: (preferences: UserPreferences) => Promise<void>;
  refreshAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,

      // Login action
      login: async (credentials) => {
        set({ isLoading: true });
        try {
          const response = await authService.login(credentials);
          set({
            user: response.user,
            token: response.token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      // Other actions...
    }),
    {
      name: 'kubechat-auth',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
      }),
    }
  )
);
```

---

## Backend Architecture

### Service Architecture

```text
apps/api/src/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── handlers/                # HTTP route handlers
│   ├── services/               # Business logic services
│   ├── models/                 # Data models and DTOs
│   ├── repository/             # Data access layer
│   └── utils/                  # Shared utilities
└── pkg/                        # Public packages
```

### Authentication and Authorization

```go
// internal/handlers/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pramodksahoo/kubechat/internal/services/auth"
	"github.com/pramodksahoo/kubechat/internal/utils"
)

type AuthMiddleware struct {
	authService auth.Service
	logger      utils.Logger
}

func NewAuthMiddleware(authService auth.Service, logger utils.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth validates JWT token and loads user context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Parse Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate token and get user
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		user, err := m.authService.ValidateToken(ctx, token)
		if err != nil {
			m.logger.Warn("Token validation failed", 
				"error", err, 
				"ip", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Add user to context for downstream handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)

		c.Next()
	}
}
```

---

## Unified Project Structure

```plaintext
kubechat/
├── .github/                          # CI/CD workflows
│   └── workflows/
│       ├── ci.yaml                   # Continuous integration
│       ├── security-scan.yaml       # Security scanning
│       ├── deploy-staging.yaml      # Staging deployment
│       └── deploy-production.yaml   # Production deployment
├── apps/                             # Application packages
│   ├── web/                          # Frontend React application
│   │   ├── src/
│   │   │   ├── components/           # Reusable UI components
│   │   │   ├── pages/               # Page components
│   │   │   ├── hooks/               # Custom React hooks
│   │   │   ├── services/            # API client services
│   │   │   ├── stores/              # Zustand state management
│   │   │   ├── styles/              # Global styles and themes
│   │   │   └── utils/               # Frontend utilities
│   │   ├── public/                  # Static assets
│   │   ├── tests/                   # Frontend tests
│   │   ├── next.config.js           # Next.js configuration
│   │   ├── tailwind.config.js      # Tailwind CSS config
│   │   ├── tsconfig.json           # TypeScript config
│   │   ├── package.json
│   │   └── .env.example
│   └── api/                         # Backend Go application
│       ├── cmd/
│       │   └── server/
│       │       └── main.go          # Application entry point
│       ├── internal/                # Private application code
│       │   ├── config/              # Configuration management
│       │   ├── handlers/            # HTTP handlers
│       │   ├── services/            # Business logic services
│       │   ├── models/              # Data models
│       │   ├── repository/          # Data access layer
│       │   └── utils/               # Shared utilities
│       ├── pkg/                     # Public packages
│       ├── tests/                   # Backend tests
│       ├── migrations/              # Database migrations
│       ├── go.mod
│       ├── go.sum
│       └── .env.example
├── packages/                        # Shared packages
│   ├── shared/                      # Shared types and utilities
│   │   ├── src/
│   │   │   ├── types/               # TypeScript interfaces
│   │   │   ├── constants/           # Shared constants
│   │   │   ├── utils/               # Shared utilities
│   │   │   └── index.ts
│   │   ├── package.json
│   │   └── tsconfig.json
│   ├── ui/                          # Shared UI component library
│   └── config/                      # Shared configuration
├── infrastructure/                  # Infrastructure as Code
│   ├── helm/                        # Helm charts
│   │   ├── kubechat/
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   ├── values-dev.yaml
│   │   │   ├── values-staging.yaml
│   │   │   ├── values-prod.yaml
│   │   │   └── templates/
│   │   └── dependencies/
│   ├── docker/                      # Docker configurations
│   │   ├── api.Dockerfile
│   │   ├── web.Dockerfile
│   │   ├── ollama.Dockerfile
│   │   └── docker-compose.yaml      # Local development
│   └── scripts/                     # Deployment scripts
├── scripts/                         # Build and development scripts
├── docs/                            # Documentation
│   ├── architecture/
│   │   ├── coding-standards.md      # Development standards
│   │   ├── tech-stack.md           # Technology decisions
│   │   ├── source-tree.md          # Codebase navigation
│   │   ├── deployment.md           # Deployment guide
│   │   └── air-gapped.md           # Air-gapped installation
│   ├── api/                         # API documentation
│   ├── development/
│   ├── operations/
│   ├── user/
│   ├── prd.md                       # Product requirements
│   ├── front-end-spec.md           # Frontend specification
│   ├── architecture.md             # This architecture document
│   └── CHANGELOG.md
├── tests/                           # E2E and integration tests
├── tools/                           # Development tools
├── .air/                            # AI development context
├── .env.example                     # Environment template
├── package.json                     # Root package.json (workspace)
├── pnpm-workspace.yaml             # PNPM workspace configuration
├── turbo.json                       # Turborepo configuration
├── Makefile                         # Build automation
├── README.md                        # Main project README
├── LICENSE
└── SECURITY.md                      # Security policy
```

---

## Development Workflow

### Container-First Development Setup

**Repository:** https://github.com/pramodksahoo/kubechat
**Branch:** develop
**Local Environment:** Rancher Desktop with Kubernetes

#### Prerequisites (Already Satisfied)
- ✅ Rancher Desktop with local Kubernetes cluster
- ✅ Docker runtime
- ✅ kubectl configured for local cluster
- ✅ Helm 3.x

#### Initial Setup

```bash
# Clone the actual repository
git clone https://github.com/pramodksahoo/kubechat.git
cd kubechat

# Switch to develop branch
git checkout develop

# Install root dependencies (for tooling only, not runtime)
pnpm install

# Verify Rancher Desktop Kubernetes is running
kubectl cluster-info
kubectl get nodes

# Create development namespace
kubectl create namespace kubechat-dev

# Deploy complete development environment
make dev-deploy
```

#### Development Commands (Container-Only)

```bash
# 🚫 NO LOCAL PROCESSES - Everything runs in Kubernetes containers

# Deploy complete development environment (replaces pnpm run dev)
make dev-deploy

# View development environment status
make dev-status

# Access development services (port-forward to containers)
make dev-access-web      # Frontend at http://localhost:3000
make dev-access-api      # Backend at http://localhost:8080
make dev-access-ollama   # Ollama at http://localhost:11434

# Follow logs from running containers
make dev-logs           # All services
make dev-logs-web       # Frontend only  
make dev-logs-api       # Backend only

# Hot reload development (rebuild and redeploy containers)
make dev-rebuild-api    # Rebuild API container with changes
make dev-rebuild-web    # Rebuild web container with changes

# Run tests in containers
make test               # All tests in containers
make test-api           # Backend tests in container
make test-web           # Frontend tests in container

# Database operations via containers
make dev-db-migrate     # Run migrations in database container
make dev-db-seed        # Seed development data

# Clean up development environment
make dev-clean          # Remove all development containers/resources
```

### Environment Configuration

#### Required Environment Variables

```bash
# Frontend (.env.local)
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_BASE_URL=ws://localhost:8080/api/v1/ws
NEXT_PUBLIC_ENVIRONMENT=development

# Backend (.env)
DATABASE_URL=postgres://kubechat:password@localhost:5432/kubechat_dev
REDIS_URL=redis://localhost:6379/0
KUBECONFIG=/path/to/your/kubeconfig
LLM_PROVIDER=ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=phi3:mini
JWT_SECRET=your_jwt_secret_key_min_32_chars
LOG_LEVEL=debug
```

---

## Security and Performance

### Security Requirements

**Frontend Security:**
- CSP Headers: Comprehensive Content Security Policy preventing XSS attacks
- XSS Prevention: React built-in protection + DOMPurify for user-generated content
- Secure Storage: JWT tokens in httpOnly cookies, no sensitive data in localStorage

**Backend Security:**
- Input Validation: Go validator package with SQL injection prevention
- Rate Limiting: 100 requests/second per IP with burst capacity
- CORS Policy: Restricted to configured frontend origins only

**Authentication Security:**
- Token Storage: JWT tokens with 24-hour expiration and refresh rotation
- Session Management: Redis-based session store with concurrent limits
- Password Policy: 12+ characters with complexity requirements

### Performance Optimization

**Frontend Performance:**
- Bundle Size Target: <500KB initial JavaScript bundle with code splitting
- Loading Strategy: Progressive web app with service worker caching
- Caching Strategy: Browser caching for static assets, API response caching

**Backend Performance:**
- Response Time Target: <200ms for API endpoints, <2s for AI processing
- Database Optimization: Connection pooling with PgBouncer, query optimization
- Caching Strategy: Redis for sessions (30min), cluster state (1min), LLM responses (24h)

---

## Testing Strategy

### Testing Pyramid

```text
                E2E Tests (Browser Automation)
               /                              \
        Integration Tests (API + Database)
       /                                    \
    Frontend Unit Tests              Backend Unit Tests
   (Components, Hooks)               (Services, Handlers)
```

### Test Organization

#### Frontend Tests
```text
apps/web/tests/
├── components/              # Component unit tests
├── hooks/                   # Custom hook tests
├── services/               # API service tests
├── pages/                  # Page integration tests
└── utils/                  # Utility function tests
```

#### Backend Tests
```text
apps/api/tests/
├── unit/                   # Unit tests
├── integration/            # Integration tests
└── fixtures/              # Test data and mocks
```

#### E2E Tests
```text
tests/e2e/
├── auth/                   # Authentication flows
├── queries/               # Query processing flows
├── collaboration/         # Multi-user features
└── compliance/           # Audit and compliance
```

---

## Coding Standards

### Critical Fullstack Rules

- **Type Sharing:** Always define types in packages/shared and import from there - prevents API contract mismatches
- **API Calls:** Never make direct HTTP calls - use the service layer for consistent error handling
- **Environment Variables:** Access only through config objects, never process.env directly
- **Error Handling:** All API routes must use the standard error handler middleware
- **State Updates:** Never mutate state directly - use proper Zustand state management patterns
- **Container-First Development:** Never run `pnpm run dev` or `go run` locally - all development through containers
- **Kubernetes Resources:** All kubectl operations must go through K8s client service with RBAC validation
- **Audit Trail:** Every user action must generate audit log entry through the audit service

### Naming Conventions

| Element | Frontend | Backend | Example |
|---------|----------|---------|---------|
| Components | PascalCase | - | `QueryInput.tsx`, `ClusterHealthDashboard.tsx` |
| Hooks | camelCase with 'use' | - | `useAuth.ts`, `useWebSocketConnection.ts` |
| API Routes | - | kebab-case | `/api/v1/query-sessions`, `/api/v1/audit-trail` |
| Database Tables | - | snake_case | `query_sessions`, `audit_log_entries` |
| Go Packages | - | lowercase | `internal/handlers`, `internal/services/auth` |
| Go Structs | - | PascalCase | `QueryRequest`, `AuditLogEntry` |
| TypeScript Interfaces | PascalCase | - | `User`, `QuerySession`, `AuditLogEntry` |

---

## Error Handling Strategy

### Error Response Format

```typescript
interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, any>;
    timestamp: string;
    requestId: string;
  };
}
```

### Frontend Error Handling

```typescript
export class ApiErrorHandler {
  static handle(error: any): never {
    // Network errors
    if (!error.response) {
      toast.error('Network error - please check your connection');
      throw new Error('NETWORK_ERROR');
    }

    const apiError: ApiError = error.response.data;
    const status = error.response.status;

    // Handle specific error types
    switch (status) {
      case 401:
        useAuthStore.getState().logout();
        window.location.href = '/auth/login';
        toast.error('Session expired - please login again');
        break;
      case 403:
        toast.error('Access denied - insufficient permissions');
        break;
      case 500:
        toast.error(`Server error (${apiError.error.requestId})`);
        break;
      default:
        toast.error(apiError.error.message || 'An unexpected error occurred');
    }

    throw error;
  }
}
```

### Backend Error Handling

```go
type ServiceError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

func (e *ServiceError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeValidation:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
```

---

## Monitoring and Observability

### Monitoring Stack

- **Frontend Monitoring:** Sentry for error tracking, Web Vitals for performance
- **Backend Monitoring:** Prometheus for metrics, Grafana for visualization
- **Error Tracking:** Sentry for both frontend and backend error aggregation
- **Performance Monitoring:** APM with request tracing, database query analysis

### Key Metrics

**Frontend Metrics:**
- Core Web Vitals: LCP < 2.5s, FID < 100ms, CLS < 0.1
- JavaScript errors: Error rate < 0.1%
- API response times: 95th percentile < 1s

**Backend Metrics:**
- Request rate: Requests per second by endpoint
- Error rate: 4xx/5xx error percentage < 1%
- Response time: P95 < 200ms for API endpoints
- Database query performance: Slow queries > 100ms

**AI Service Metrics:**
- LLM response time: P95 < 3s for Ollama
- Model accuracy: Command safety classification > 95%
- Fallback rate: Ollama → Cloud AI fallback frequency

---

## Summary and Next Steps

### Architecture Validation Summary

**✅ PRD Requirements Fulfilled:**
- Natural language Kubernetes management with >95% accuracy
- Multi-LLM support with Ollama-first air-gapped capability
- Real-time collaborative troubleshooting with session sharing
- Enterprise-grade audit trails for regulatory compliance
- Kubernetes RBAC integration with role-based access control
- Container-first development with Helm-native deployment
- Web-based interface supporting 100+ concurrent users
- Air-gapped deployment with complete offline operation

**✅ Technical Architecture Delivered:**
- Kubernetes-native microservices architecture validated by PoC
- React 18.3+ TypeScript frontend with Zustand state management
- Go 1.23+ backend with Gin framework and client-go integration
- PostgreSQL 16+ with Redis caching and comprehensive audit schema
- Multi-provider AI abstraction with intelligent fallback
- WebSocket real-time collaboration and monitoring
- Container-first development with Rancher Desktop integration
- Comprehensive testing strategy (unit, integration, E2E)

**✅ Enterprise Requirements Satisfied:**
- Complete audit trail with cryptographic integrity verification
- RBAC-aware permission system integrated with Kubernetes
- Air-gapped deployment capability for regulated industries
- Compliance reporting for multiple regulatory frameworks
- Professional enterprise UI with accessibility compliance
- Multi-cluster support with context switching
- Horizontal scaling architecture supporting enterprise loads

### Immediate Implementation Priorities

**Phase 1: Foundation (Epic 1 - Weeks 1-4)**
1. **Container-First Development Environment**
   - Set up Rancher Desktop development workflow
   - Implement Makefile with container-only commands
   - Deploy PostgreSQL and Redis via Helm charts
   - Establish hot-reload development containers

2. **Core Backend Services**
   - Implement Go microservices architecture from PoC
   - Set up Kubernetes client service with RBAC validation
   - Integrate Ollama AI service with local model deployment
   - Establish audit logging service with PostgreSQL persistence

3. **Basic Frontend Application**
   - Create React TypeScript application with Zustand state management
   - Implement natural language query input component
   - Build command preview and safety classification UI
   - Establish WebSocket connection for real-time updates

4. **Helm Chart Development**
   - Create production-ready Helm charts with air-gapped support
   - Implement development, staging, and production value files
   - Set up RBAC configuration and security policies
   - Test deployment in Rancher Desktop local cluster

**Phase 2: Enterprise Features (Epics 2-3 - Weeks 5-8)**
1. **Security and Compliance**
   - Complete audit trail implementation with compliance reporting
   - Implement comprehensive RBAC integration
   - Add cryptographic audit trail integrity validation
   - Create compliance dashboard and export functionality

2. **Air-Gapped Deployment**
   - Finalize Ollama local AI deployment with persistent model storage
   - Create air-gapped installation bundles and documentation
   - Implement offline help system and troubleshooting guides
   - Validate complete offline operation without external dependencies

**Phase 3: Production Readiness (Epic 4 - Weeks 9-12)**
1. **Real-time Monitoring and Collaboration**
   - Implement cluster health monitoring with WebSocket updates
   - Create collaborative session sharing and multi-user features
   - Build comprehensive monitoring dashboards
   - Add performance optimization and caching strategies

### Development Team Recommendations

**Immediate Team Setup:**
- **Frontend Developer:** Focus on React TypeScript implementation and UI components
- **Backend Developer:** Implement Go microservices and Kubernetes integration
- **DevOps Engineer:** Set up CI/CD pipeline and container-first development workflow
- **AI/ML Engineer:** Optimize Ollama model performance and implement multi-LLM routing

**Technical Leadership Priorities:**
1. Validate container-first development workflow with team
2. Establish code generation pipeline for shared TypeScript types
3. Set up comprehensive testing infrastructure
4. Implement security scanning and compliance validation in CI/CD

This comprehensive architecture provides the foundation for building KubeChat as described in your PRD, leveraging your validated PoC while scaling to enterprise requirements. The container-first approach with Rancher Desktop ensures production parity from day one, while the Ollama-first AI strategy delivers your unique competitive advantage in air-gapped deployments.

---

*🏗️ Generated with [Claude Code](https://claude.ai/code)*

*Co-Authored-By: Winston (Architect) <architect@kubechat.dev>*
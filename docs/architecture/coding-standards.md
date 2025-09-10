# KubeChat Coding Standards

## Overview

This document defines the critical coding standards for KubeChat development. These rules prevent common architectural mistakes and ensure consistency across the fullstack codebase.

**Repository:** https://github.com/pramodksahoo/kubechat  
**Branch:** develop  
**Scope:** Frontend (React/TypeScript) + Backend (Go) + Infrastructure (Helm/Docker)

---

## Critical Fullstack Rules

### 🔒 Type Safety Rules

**Rule:** Always define types in `packages/shared` and import from there
```typescript
// ❌ BAD - Defining types locally
interface User {
  id: string;
  name: string;
}

// ✅ GOOD - Import from shared package
import type { User } from '@kubechat/shared/types';
```
**Rationale:** Prevents API contract mismatches between frontend and backend. Ensures type definitions are single source of truth.

---

### 🌐 API Communication Rules

**Rule:** Never make direct HTTP calls - use the service layer
```typescript
// ❌ BAD - Direct axios call
const response = await axios.post('/api/v1/queries', { query });

// ✅ GOOD - Use service layer
import { queryService } from '@/services/api';
const response = await queryService.submitQuery(sessionId, query);
```
**Rationale:** Service layer provides consistent error handling, authentication, and request/response transformation.

---

### 🔧 Environment Configuration Rules

**Rule:** Access only through config objects, never process.env directly
```typescript
// ❌ BAD - Direct environment access
const apiUrl = process.env.NEXT_PUBLIC_API_BASE_URL;

// ✅ GOOD - Through config object
import { config } from '@/config';
const apiUrl = config.api.baseUrl;
```
**Rationale:** Enables type safety, validation, and centralized configuration management.

---

### 🛡️ Error Handling Rules

**Rule:** All API routes must use the standard error handler
```go
// ❌ BAD - Custom error handling
func (h *Handler) ProcessQuery(c *gin.Context) {
    if err != nil {
        c.JSON(500, gin.H{"error": "something went wrong"})
        return
    }
}

// ✅ GOOD - Standard error handler
func (h *Handler) ProcessQuery(c *gin.Context) {
    if err != nil {
        serviceErr := utils.NewServiceError(utils.ErrCodeAI, "Query processing failed")
        panic(serviceErr) // Handled by error middleware
    }
}
```
**Rationale:** Consistent error format, automatic audit logging, and proper request ID tracking.

---

### 🔄 State Management Rules

**Rule:** Never mutate state directly - use proper Zustand patterns
```typescript
// ❌ BAD - Direct mutation
const { user } = useAuthStore();
user.preferences.theme = 'dark'; // Mutates state directly

// ✅ GOOD - Proper state update
const { updateUserPreferences } = useAuthStore();
updateUserPreferences({ theme: 'dark' });
```
**Rationale:** Maintains state consistency, enables proper re-rendering, and supports state persistence.

---

### 🐳 Container-First Development Rules

**Rule:** Never run `pnpm run dev` or `go run` locally - use containers only
```bash
# ❌ BAD - Local development processes
pnpm run dev
go run ./cmd/server

# ✅ GOOD - Container-first development
make dev-deploy
make dev-access-web
```
**Rationale:** Ensures production parity, consistent environment, and supports air-gapped deployment testing.

---

### ☸️ Kubernetes Integration Rules

**Rule:** All kubectl operations must go through K8s client service with RBAC
```go
// ❌ BAD - Direct kubectl execution
cmd := exec.Command("kubectl", "get", "pods")
output, _ := cmd.Output()

// ✅ GOOD - Through K8s client service
pods, err := h.k8sClient.GetPods(ctx, namespace)
if err != nil {
    return utils.HandleKubernetesError(err)
}
```
**Rationale:** Proper RBAC validation, error handling, and audit trail integration.

---

### 📋 Audit Trail Rules

**Rule:** Every user action must generate audit log entry
```go
// ❌ BAD - No audit logging
result := h.executeCommand(command)
return c.JSON(200, result)

// ✅ GOOD - Audit trail included
result := h.executeCommand(command)

// Log to audit trail
h.auditService.LogExecution(audit.ExecutionEntry{
    UserID:    user.ID,
    Command:   command,
    Success:   result.Success,
    Timestamp: time.Now(),
})

return c.JSON(200, result)
```
**Rationale:** Required for enterprise compliance (SOX, HIPAA, SOC 2). Ensures all actions are traceable.

---

## Frontend Coding Standards

### Component Organization

```typescript
// ✅ Component file structure
export interface QueryInputProps {
  sessionId: string;
  onQuerySubmitted?: (query: Query) => void;
  disabled?: boolean;
}

export const QueryInput: React.FC<QueryInputProps> = ({ 
  sessionId, 
  onQuerySubmitted, 
  disabled = false 
}) => {
  // Component implementation
};

export default QueryInput;
```

### Hook Patterns

```typescript
// ✅ Custom hook structure
export const useQuerySession = (sessionId: string) => {
  const [queries, setQueries] = useState<Query[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submitQuery = useCallback(async (query: string) => {
    // Implementation
  }, [sessionId]);

  return {
    queries,
    isLoading,
    error,
    submitQuery,
  };
};
```

### State Store Patterns

```typescript
// ✅ Zustand store structure
interface AuthState {
  // State properties
  user: User | null;
  isAuthenticated: boolean;
  
  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  isAuthenticated: false,
  
  login: async (credentials) => {
    // Implementation with proper error handling
  },
  
  logout: () => {
    set({ user: null, isAuthenticated: false });
  },
}));
```

---

## Backend Coding Standards

### Service Architecture

```go
// ✅ Service interface pattern
type QueryService interface {
    ProcessQuery(ctx context.Context, req ProcessQueryRequest) (*QueryResult, error)
    ExecuteCommand(ctx context.Context, req ExecuteCommandRequest) (*ExecutionResult, error)
}

type queryService struct {
    nlpService    NLPService
    k8sClient     K8sClient
    auditService  AuditService
    logger        Logger
}

func NewQueryService(
    nlpService NLPService,
    k8sClient K8sClient,
    auditService AuditService,
    logger Logger,
) QueryService {
    return &queryService{
        nlpService:   nlpService,
        k8sClient:    k8sClient,
        auditService: auditService,
        logger:       logger,
    }
}
```

### Error Handling Patterns

```go
// ✅ Standard error handling
func (s *queryService) ProcessQuery(ctx context.Context, req ProcessQueryRequest) (*QueryResult, error) {
    // Validate input
    if req.Query == "" {
        return nil, utils.NewServiceError(utils.ErrCodeValidation, "Query cannot be empty")
    }

    // Process with proper error handling
    result, err := s.nlpService.GenerateCommand(ctx, req.Query)
    if err != nil {
        return nil, utils.HandleAIError(err, "ollama")
    }

    // Log successful operation
    s.logger.Info("Query processed successfully",
        "user_id", req.UserID,
        "processing_time", time.Since(startTime),
    )

    return result, nil
}
```

### Database Repository Patterns

```go
// ✅ Repository interface
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
}

// ✅ Repository implementation
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    query := `
        INSERT INTO users (id, username, email, roles, clusters, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`

    _, err := r.db.ExecContext(ctx, query,
        user.ID,
        user.Username,
        user.Email,
        pq.Array(user.Roles),
        pq.Array(user.Clusters),
        user.CreatedAt,
        user.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    return nil
}
```

---

## Naming Conventions

### Frontend Conventions

| Element | Convention | Example | Notes |
|---------|------------|---------|-------|
| **Components** | PascalCase | `QueryInput`, `ClusterHealthDashboard` | React component files |
| **Hooks** | camelCase with 'use' prefix | `useAuth`, `useWebSocket` | Custom React hooks |
| **Stores** | camelCase with Store suffix | `authStore`, `sessionStore` | Zustand store files |
| **Services** | camelCase with Service suffix | `queryService`, `authService` | API service files |
| **Types** | PascalCase | `User`, `QuerySession` | TypeScript interfaces |
| **Constants** | SCREAMING_SNAKE_CASE | `API_BASE_URL`, `MAX_QUERY_LENGTH` | Configuration constants |
| **CSS Classes** | kebab-case | `query-input`, `cluster-health-card` | Tailwind utility classes |

### Backend Conventions

| Element | Convention | Example | Notes |
|---------|------------|---------|-------|
| **Packages** | lowercase | `handlers`, `services`, `models` | Go package names |
| **Structs** | PascalCase | `QueryRequest`, `AuditLogEntry` | Public structs |
| **Interfaces** | PascalCase | `QueryService`, `UserRepository` | Service interfaces |
| **Methods** | PascalCase | `ProcessQuery`, `GetUserByID` | Public methods |
| **Variables** | camelCase | `userID`, `processingTime` | Local variables |
| **Constants** | PascalCase or camelCase | `DefaultTimeout`, `maxRetries` | Based on visibility |

### Database Conventions

| Element | Convention | Example | Notes |
|---------|------------|---------|-------|
| **Tables** | snake_case | `query_sessions`, `audit_log_entries` | PostgreSQL tables |
| **Columns** | snake_case | `user_id`, `created_at` | Database columns |
| **Indexes** | snake_case with prefix | `idx_users_username`, `idx_audit_timestamp` | Database indexes |
| **Constraints** | snake_case descriptive | `valid_email`, `non_empty_username` | Table constraints |

### API Conventions

| Element | Convention | Example | Notes |
|---------|------------|---------|-------|
| **Endpoints** | kebab-case | `/api/v1/query-sessions`, `/api/v1/audit-trail` | REST API paths |
| **Parameters** | snake_case | `user_id`, `session_id` | URL and query parameters |
| **JSON Fields** | snake_case | `natural_language`, `generated_command` | API request/response |

---

## File Organization Standards

### Frontend File Structure

```text
src/
├── components/
│   ├── ui/              # Basic UI components
│   │   ├── Button/
│   │   │   ├── Button.tsx
│   │   │   ├── Button.test.tsx
│   │   │   ├── Button.stories.tsx
│   │   │   └── index.ts
│   │   └── index.ts
│   ├── forms/          # Form-specific components
│   └── layout/         # Layout components
├── pages/              # Page components (Next.js routing)
├── hooks/              # Custom React hooks
├── services/           # API service layer
├── stores/             # Zustand state stores
├── types/              # Local TypeScript types (import from shared)
└── utils/              # Utility functions
```

### Backend File Structure

```text
internal/
├── handlers/           # HTTP request handlers
├── services/          # Business logic services
│   ├── auth/
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── types.go
│   └── nlp/
├── models/            # Data models
├── repository/        # Data access layer
└── utils/             # Shared utilities
```

---

## Testing Standards

### Frontend Testing

```typescript
// ✅ Component test structure
describe('QueryInput', () => {
  const mockSubmitQuery = jest.fn();
  
  beforeEach(() => {
    mockSubmitQuery.mockClear();
  });

  it('should submit query when form is submitted', async () => {
    render(<QueryInput sessionId="test-session" onQuerySubmitted={mockSubmitQuery} />);
    
    const textarea = screen.getByPlaceholderText(/Ask KubeChat/);
    const submitButton = screen.getByRole('button', { name: /Ask/ });
    
    fireEvent.change(textarea, { target: { value: 'show me pods' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockSubmitQuery).toHaveBeenCalledWith(expect.objectContaining({
        naturalLanguage: 'show me pods'
      }));
    });
  });
});
```

### Backend Testing

```go
// ✅ Service test structure
func TestQueryService_ProcessQuery(t *testing.T) {
    // Setup
    mockNLP := &mocks.MockNLPService{}
    mockAudit := &mocks.MockAuditService{}
    service := NewQueryService(mockNLP, mockAudit, &mocks.MockLogger{})

    // Test case
    t.Run("successful query processing", func(t *testing.T) {
        // Arrange
        mockNLP.On("GenerateCommand", mock.Anything, "show me pods").Return(&NLPResult{
            Command: "kubectl get pods",
            Safety:  "safe",
        }, nil)

        // Act
        result, err := service.ProcessQuery(ctx, ProcessQueryRequest{
            Query:  "show me pods",
            UserID: testUserID,
        })

        // Assert
        assert.NoError(t, err)
        assert.Equal(t, "kubectl get pods", result.Command)
        mockNLP.AssertExpectations(t)
    })
}
```

---

## Security Standards

### Input Validation

```go
// ✅ Input validation pattern
func ValidateQueryRequest(req *QueryRequest) error {
    if req.Query == "" {
        return utils.NewServiceError(utils.ErrCodeValidation, "Query is required")
    }
    
    if len(req.Query) > 2000 {
        return utils.NewServiceError(utils.ErrCodeValidation, "Query too long (max 2000 characters)")
    }
    
    // Sanitize input to prevent injection attacks
    req.Query = strings.TrimSpace(req.Query)
    req.Query = regexp.MustCompile(`[^\w\s\-\.\:\/]`).ReplaceAllString(req.Query, "")
    
    return nil
}
```

### Authentication Patterns

```typescript
// ✅ Frontend auth pattern
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuthStore();
  
  if (isLoading) {
    return <LoadingSpinner />;
  }
  
  if (!isAuthenticated) {
    return <Navigate to="/auth/login" replace />;
  }
  
  return <>{children}</>;
};
```

---

## Performance Standards

### Frontend Performance

```typescript
// ✅ Component optimization
const QueryHistory = React.memo(({ queries }: QueryHistoryProps) => {
  const memoizedQueries = useMemo(() => 
    queries.filter(q => q.status === 'completed'), 
    [queries]
  );
  
  return (
    <VirtualizedList 
      items={memoizedQueries}
      renderItem={({ item }) => <QueryItem query={item} />}
    />
  );
});

// ✅ Proper dependency array
const useClusterHealth = (clusterId: string) => {
  const [health, setHealth] = useState<ClusterHealth | null>(null);
  
  useEffect(() => {
    const fetchHealth = async () => {
      const data = await healthService.getClusterHealth(clusterId);
      setHealth(data);
    };
    
    fetchHealth();
  }, [clusterId]); // Correct dependency
  
  return health;
};
```

### Backend Performance

```go
// ✅ Database query optimization
func (r *queryRepository) GetQueriesBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.Query, error) {
    // Use prepared statement with proper indexing
    query := `
        SELECT id, session_id, user_id, natural_language, generated_command, safety_level, created_at
        FROM queries 
        WHERE session_id = $1 
        ORDER BY created_at DESC 
        LIMIT $2`

    rows, err := r.db.QueryContext(ctx, query, sessionID, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to query queries: %w", err)
    }
    defer rows.Close()

    // Process results efficiently
    queries := make([]*models.Query, 0, limit)
    for rows.Next() {
        var q models.Query
        if err := rows.Scan(&q.ID, &q.SessionID, &q.UserID, &q.NaturalLanguage, &q.GeneratedCommand, &q.SafetyLevel, &q.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan query: %w", err)
        }
        queries = append(queries, &q)
    }

    return queries, nil
}
```

---

## Documentation Standards

### Code Comments

```go
// ✅ Function documentation
// ProcessQuery converts natural language into kubectl commands using AI services.
// It validates user permissions, processes the query through the NLP service,
// and logs all actions for audit compliance.
//
// Parameters:
//   ctx: Request context with timeout and cancellation
//   req: Query request containing user input and context
//
// Returns:
//   QueryResult with generated command and safety classification
//   Error if processing fails or user lacks permissions
func (s *queryService) ProcessQuery(ctx context.Context, req ProcessQueryRequest) (*QueryResult, error) {
    // Implementation
}
```

```typescript
// ✅ Component documentation
/**
 * QueryInput provides natural language input for Kubernetes commands.
 * Includes safety validation, command preview, and execution controls.
 * 
 * @param sessionId - Active query session identifier
 * @param onQuerySubmitted - Callback when query is successfully processed
 * @param disabled - Disables input during processing
 */
export const QueryInput: React.FC<QueryInputProps> = ({
  sessionId,
  onQuerySubmitted,
  disabled = false
}) => {
  // Implementation
};
```

---

## Enforcement

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: frontend-lint
        name: Frontend Linting
        entry: pnpm lint:frontend
        language: system
        files: ^apps/web/
        
      - id: backend-lint
        name: Backend Linting  
        entry: golangci-lint run
        language: system
        files: ^apps/api/
        
      - id: type-check
        name: TypeScript Check
        entry: pnpm type-check
        language: system
        files: \.(ts|tsx)$
```

### CI/CD Integration

```yaml
# GitHub Actions
- name: Validate Coding Standards
  run: |
    pnpm lint:frontend --max-warnings 0
    pnpm type-check:frontend
    cd apps/api && golangci-lint run --timeout 5m
    cd apps/api && go vet ./...
```

### Code Review Checklist

**Frontend Review:**
- [ ] Components use proper TypeScript interfaces
- [ ] API calls go through service layer
- [ ] State updates use Zustand patterns
- [ ] Error handling follows standard patterns
- [ ] Tests cover happy path and error cases

**Backend Review:**
- [ ] All errors use ServiceError pattern
- [ ] Database queries use repository pattern
- [ ] RBAC validation for Kubernetes operations
- [ ] Audit logging for user actions
- [ ] Context cancellation respected

**General Review:**
- [ ] No direct process.env access
- [ ] Naming conventions followed
- [ ] File organization correct
- [ ] Documentation updated
- [ ] No container-first violations

---

For questions about coding standards or to propose changes, create an issue in the repository with the `coding-standards` label.

---

*📝 Generated with [Claude Code](https://claude.ai/code)*

*Co-Authored-By: Winston (Architect) <architect@kubechat.dev>*
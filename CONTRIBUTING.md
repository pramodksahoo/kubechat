# Contributing to KubeChat

Thank you for your interest in contributing to KubeChat! This guide will help you get started with our container-first development approach and contribution process.

## 🎯 Overview

KubeChat is built with a **container-first development philosophy**. All development, testing, and deployment happens inside containers running on Kubernetes. This ensures consistency, reproducibility, and production parity across all environments.

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have the following tools installed:

| Tool | Version | Purpose |
|------|---------|---------|
| **Docker** | Latest | Container runtime (via Rancher Desktop) |
| **Kubernetes** | 1.28+ | Container orchestration |
| **Helm** | 3.15+ | Kubernetes package manager |
| **kubectl** | 1.28+ | Kubernetes CLI |
| **Node.js** | 20+ | Frontend tooling |
| **PNPM** | 8+ | Package manager |
| **Go** | 1.23+ | Backend development |

**Recommended Setup**: [Rancher Desktop](https://rancherdesktop.io/) provides Docker, Kubernetes, and kubectl in a single installation.

### Environment Setup

1. **Fork the Repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/kubechat.git
   cd kubechat
   ```

2. **Validate Your Environment**
   ```bash
   # Run comprehensive environment validation
   ./infrastructure/scripts/validate-env.sh
   
   # Should show all green checkmarks ✅
   ```

3. **Initialize Development Environment**
   ```bash
   # One-command setup
   make init
   
   # This will:
   # - Build all containers
   # - Deploy to local Kubernetes
   # - Set up databases
   # - Seed development data
   ```

4. **Verify Installation**
   ```bash
   # Check deployment status
   make dev-status
   
   # Access the application
   # Frontend: http://localhost:30001
   # API: http://localhost:30080
   ```

## 🛠️ Development Workflow

### Container-First Development Rules

**✅ DO:**
- Use `make dev-deploy` to run the application
- Use `make dev-shell-api` or `make dev-shell-web` to access containers
- Make code changes on your host machine (hot reload enabled)
- Run tests inside containers with `make dev-test`

**❌ DON'T:**
- Run `pnpm run dev` or `npm start` locally
- Run `go run` or direct Go commands locally  
- Install dependencies directly on your host machine
- Connect to databases directly from your host

### Daily Development Cycle

```bash
# 1. Start your day
git pull origin develop
make dev-deploy

# 2. Make code changes in your editor
# Files are automatically synced to containers

# 3. Rebuild containers after significant changes
make dev-rebuild-api    # For backend changes
make dev-rebuild-web    # For frontend changes

# 4. View logs and debug
make dev-logs           # All services
make dev-logs-api       # API only
make dev-logs-web       # Frontend only

# 5. Run tests
make dev-test           # All tests
make dev-test-unit      # Unit tests only

# 6. Access containers for debugging
make dev-shell-api      # Shell into API container
make dev-shell-web      # Shell into web container
```

## 📋 Contribution Process

### 1. Choose Your Contribution Type

**🐛 Bug Fixes**
- Fix existing functionality that isn't working correctly
- Include reproduction steps and test cases
- Update documentation if behavior changes

**✨ New Features**
- Add new functionality to KubeChat
- Discuss in GitHub Issues before starting large features
- Include comprehensive tests and documentation

**📚 Documentation**
- Improve existing docs or add new documentation
- Ensure accuracy with current container-first approach
- Include examples and code snippets

**🔧 Infrastructure**
- Improve build processes, CI/CD, or deployment
- Enhance development experience
- Optimize container configurations

### 2. Create a Feature Branch

```bash
# Create and switch to a new branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/bug-description
```

### 3. Make Your Changes

Follow our coding standards and architecture guidelines:

- **Frontend**: React + TypeScript, use shared types from `packages/shared`
- **Backend**: Go with clean architecture, follow error handling patterns
- **Database**: Use migration scripts, follow naming conventions
- **Tests**: Add unit and integration tests for new functionality

### 4. Test Your Changes

```bash
# Run all tests
make dev-test

# Run specific test types
make dev-test-unit      # Unit tests
make dev-test-e2e       # End-to-end tests

# Lint code
make dev-lint

# Check for security issues
make dev-security-scan
```

### 5. Document Your Changes

- Update relevant documentation files
- Add docstrings/comments for new functions
- Update API documentation if endpoints changed
- Include migration guides for breaking changes

### 6. Submit a Pull Request

```bash
# Push your branch
git push origin feature/your-feature-name

# Create a pull request on GitHub with:
# - Clear title and description
# - Link to related issues
# - Screenshots/demos for UI changes
# - Testing instructions
```

## 🏗️ Architecture Guidelines

### Project Structure

Follow the established monorepo structure:

```
kubechat/
├── apps/web/              # React frontend
├── apps/api/              # Go backend
├── packages/shared/       # Shared TypeScript types
├── packages/ui/           # UI component library
├── infrastructure/        # Deployment configurations
├── tests/                 # Integration tests
└── docs/                  # Documentation
```

### Coding Standards

See [docs/architecture/coding-standards.md](docs/architecture/coding-standards.md) for detailed guidelines:

#### Frontend Standards

```typescript
// ✅ Good: Use shared types
import type { User, Query } from '@kubechat/shared/types';

// ✅ Good: Use service layer for API calls
import { queryService } from '@/services/api';
const result = await queryService.submitQuery(sessionId, query);

// ❌ Bad: Direct axios calls
const response = await axios.post('/api/queries', { query });
```

#### Backend Standards

```go
// ✅ Good: Use standard error handling
func (s *service) ProcessQuery(ctx context.Context, req ProcessQueryRequest) (*QueryResult, error) {
    if req.Query == "" {
        return nil, utils.NewServiceError(utils.ErrCodeValidation, "Query cannot be empty")
    }
    // ... implementation
}

// ❌ Bad: Custom error handling
func (h *Handler) ProcessQuery(c *gin.Context) {
    if err != nil {
        c.JSON(500, gin.H{"error": "something went wrong"})
        return
    }
}
```

### Database Guidelines

```sql
-- ✅ Good: Use migrations for schema changes
-- File: infrastructure/scripts/database/migrations/002_add_user_preferences.sql
ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}';
CREATE INDEX idx_users_preferences ON users USING GIN(preferences);
```

## 🧪 Testing Guidelines

### Test Structure

```
tests/
├── unit/                 # Component/function level tests
├── integration/          # Service integration tests  
├── e2e/                  # End-to-end user scenarios
└── performance/          # Load and performance tests
```

### Writing Tests

**Frontend Tests:**
```typescript
// Component test example
describe('QueryInput', () => {
  it('should submit query when form is valid', async () => {
    render(<QueryInput sessionId="test" onSubmit={mockSubmit} />);
    
    fireEvent.change(screen.getByRole('textbox'), { 
      target: { value: 'show me pods' } 
    });
    fireEvent.click(screen.getByRole('button', { name: /submit/i }));
    
    expect(mockSubmit).toHaveBeenCalledWith('show me pods');
  });
});
```

**Backend Tests:**
```go
func TestQueryService_ProcessQuery(t *testing.T) {
    // Setup
    mockNLP := &mocks.MockNLPService{}
    service := NewQueryService(mockNLP, mockLogger)

    // Test
    result, err := service.ProcessQuery(ctx, ProcessQueryRequest{
        Query: "show me pods",
    })

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "kubectl get pods", result.Command)
}
```

### Running Tests in Containers

```bash
# All tests run inside containers
make dev-test

# Access test containers for debugging
make dev-shell-api
go test -v ./internal/services/...

make dev-shell-web  
npm test -- --coverage
```

## 🔒 Security Guidelines

### Security Principles

1. **Input Validation**: Validate all user inputs
2. **Authentication**: Use JWT tokens for API access
3. **Authorization**: Implement RBAC for Kubernetes operations
4. **Audit Logging**: Log all user actions
5. **Data Protection**: Encrypt sensitive data at rest

### Security Checklist

- [ ] Input validation on all endpoints
- [ ] SQL injection prevention
- [ ] XSS protection in frontend
- [ ] CSRF tokens for state-changing operations
- [ ] Secure headers in API responses
- [ ] No secrets in logs or error messages

### Security Vulnerability Reporting

**🚨 IMPORTANT**: Do NOT create public GitHub issues for security vulnerabilities.

If you discover a security vulnerability:

1. **Email**: security@kubechat.dev (GPG key available on request)
2. **GitHub**: Use [private security reporting](https://github.com/pramodksahoo/kubechat/security/advisories/new)
3. **Response Time**: We aim to respond within 24 hours
4. **Disclosure**: We follow coordinated disclosure practices

Include in your report:
- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if available)

### Security Review Process

All contributions undergo security review:
- Automated security scanning in CI/CD
- Manual review for security-sensitive changes
- Dependency vulnerability scanning
- Container image security analysis

## 📚 Documentation Standards

### Documentation Types

1. **Code Documentation**: Inline comments and docstrings
2. **API Documentation**: OpenAPI specifications
3. **User Documentation**: How-to guides and tutorials
4. **Architecture Documentation**: Design decisions and patterns

### Writing Guidelines

- Use clear, concise language
- Include code examples
- Provide context for decisions
- Update docs with code changes
- Test documentation accuracy

## 🚀 Release Process

### Version Management

- Follow [Semantic Versioning](https://semver.org/)
- Tag releases with `vX.Y.Z` format
- Maintain changelog with notable changes

### Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] Breaking changes documented
- [ ] Migration guides provided
- [ ] Security review completed

## 🤝 Community Guidelines

### Code of Conduct

We are committed to providing a welcoming and inclusive environment:

- Be respectful and constructive
- Welcome newcomers and help them learn
- Focus on technical merit of contributions
- Maintain professionalism in all interactions

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Pull Requests**: Code review and collaboration

### Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

## ❓ Getting Help

### Troubleshooting

1. **Environment Issues**: Run `./infrastructure/scripts/validate-env.sh`
2. **Build Problems**: Try `make dev-clean && make dev-build`
3. **Deployment Issues**: Check `make dev-status` and `make dev-logs`

### Resources

- **Development Guide**: [DEVELOPMENT.md](DEVELOPMENT.md)
- **Architecture Docs**: [docs/architecture/](docs/architecture/)
- **Troubleshooting**: [docs/troubleshooting/](docs/troubleshooting/)

### Support

- Create GitHub Issues for bugs or feature requests
- Use GitHub Discussions for questions
- Review existing issues before creating new ones

## 🎉 Thank You!

Your contributions help make KubeChat better for everyone. Whether it's code, documentation, testing, or feedback, every contribution is valuable and appreciated!

---

**Questions?** Feel free to reach out via GitHub Issues or Discussions. We're here to help you succeed with your contributions to KubeChat.
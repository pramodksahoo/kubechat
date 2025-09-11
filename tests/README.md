# KubeChat Testing Suite

This directory contains comprehensive tests for the KubeChat application.

## Test Organization

### `/unit/` - Component/Function Level Tests
- Component tests for React UI components
- Function tests for utility functions
- Service tests with mocked dependencies
- Fast execution, isolated testing

### `/integration/` - Service Integration Tests  
- API endpoint tests
- Database integration tests
- Service-to-service communication tests
- Container-based test execution

### `/e2e/` - End-to-End Scenarios
- Full user journey testing
- Browser automation with Playwright
- Complete system integration tests
- Production-like environment testing

### `/performance/` - Load and Performance Tests
- Load testing for API endpoints
- Database performance tests
- Memory and CPU usage profiling
- Scalability testing

## Running Tests

All tests run inside containers following container-first development principles:

```bash
# Run all tests
make dev-test

# Run specific test types
make dev-test-unit       # Unit tests only
make dev-test-e2e        # End-to-end tests only

# Run tests in specific containers
make dev-shell-web       # Frontend test environment
make dev-shell-api       # Backend test environment
```

## Test Requirements

- **Framework Requirements:**
  - Frontend: Vitest 2.0+ and Testing Library 16+
  - Backend: Go testing package with Testify 1.9+
  - E2E: Playwright 1.47+

- **Container Testing:**
  - All tests must run inside containers
  - No local test execution allowed
  - Test results accessible through Makefile commands

## Test Standards

- Component tests with proper mocking patterns
- Service tests with dependency injection
- Integration tests covering API endpoints
- Container-based test execution only

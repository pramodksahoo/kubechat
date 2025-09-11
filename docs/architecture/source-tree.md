# KubeChat Source Tree Documentation

This document describes the complete monorepo structure for KubeChat, following container-first development principles.

## Complete Monorepo Structure

```
kubechat/
├── apps/                          # Application services
│   ├── web/                       # React TypeScript frontend
│   │   ├── src/                   # Source code
│   │   │   ├── components/        # React components
│   │   │   ├── pages/             # Next.js pages
│   │   │   ├── hooks/             # Custom React hooks
│   │   │   ├── utils/             # Frontend utilities
│   │   │   └── types/             # Frontend-specific types
│   │   ├── public/                # Static assets
│   │   ├── tests/                 # Frontend tests
│   │   ├── package.json           # Frontend dependencies
│   │   ├── next.config.js         # Next.js configuration
│   │   ├── tailwind.config.js     # Tailwind CSS config
│   │   └── vitest.config.ts       # Test configuration
│   └── api/                       # Go backend services
│       ├── cmd/                   # Application entry points
│       │   └── server/            # Main server entry point
│       ├── internal/              # Private application code
│       │   ├── handlers/          # HTTP handlers
│       │   ├── services/          # Business logic
│       │   ├── repositories/      # Data access layer
│       │   └── models/            # Data models
│       ├── pkg/                   # Public Go packages
│       │   ├── auth/              # Authentication utilities
│       │   ├── kubernetes/        # Kubernetes client wrapper
│       │   └── utils/             # Common utilities
│       ├── tests/                 # Backend tests
│       ├── go.mod                 # Go dependencies
│       └── go.sum                 # Go dependency checksums
│
├── packages/                      # Shared packages  
│   ├── shared/                    # Shared TypeScript types
│   │   ├── types/                 # Common type definitions
│   │   ├── utils/                 # Shared utilities
│   │   └── package.json           # Shared package config
│   ├── ui/                        # Component library
│   │   ├── components/            # Reusable UI components
│   │   ├── hooks/                 # Shared UI hooks
│   │   └── package.json           # UI package config
│   └── config/                    # Shared configuration
│       ├── env/                   # Environment configuration
│       ├── constants/             # Application constants
│       └── package.json           # Config package config
│
├── infrastructure/                # DevOps and deployment
│   ├── helm/                      # Kubernetes deployment charts
│   │   └── kubechat/             # Main Helm chart
│   │       ├── templates/         # Kubernetes templates
│   │       ├── charts/            # Sub-charts
│   │       ├── values.yaml        # Production values
│   │       └── values-dev.yaml    # Development values
│   ├── docker/                    # Container configurations
│   │   ├── web/                   # Frontend Dockerfile
│   │   └── api/                   # Backend Dockerfile
│   ├── scripts/                   # Build automation
│   │   ├── validate-env.sh        # Environment validation
│   │   ├── health-check.sh        # Health monitoring
│   │   └── database/              # Database scripts
│   └── k8s/                       # Raw Kubernetes manifests
│       ├── base/                  # Base resources
│       └── overlays/              # Environment overlays
│
├── docs/                          # Project documentation
│   ├── architecture/              # Architecture documentation
│   │   ├── source-tree.md         # This file
│   │   ├── tech-stack.md          # Technology stack
│   │   ├── deployment.md          # Deployment guide
│   │   └── coding-standards.md    # Coding standards
│   ├── stories/                   # User stories
│   └── troubleshooting/           # Troubleshooting guides
│
├── tests/                         # Cross-service integration tests
│   ├── unit/                      # Component/function tests
│   ├── integration/               # Service integration tests
│   ├── e2e/                       # End-to-end scenarios
│   └── performance/               # Load and performance tests
│
├── .github/                       # GitHub configuration
│   ├── workflows/                 # CI/CD workflows
│   └── ISSUE_TEMPLATE/            # Issue templates
│
├── pnpm-workspace.yaml           # PNPM workspace config
├── Makefile                      # Development commands
├── .gitignore                    # Git ignore rules
├── README.md                     # Project overview
├── DEVELOPMENT.md                # Development guide
└── CONTRIBUTING.md               # Contribution guidelines
```

## Directory Purposes

### `/apps/` - Application Services
Contains the main applications that make up KubeChat:
- **Frontend**: React TypeScript application with Next.js
- **Backend**: Go API services with Gin framework

### `/packages/` - Shared Code
Monorepo packages shared across applications:
- **shared**: Common types and utilities
- **ui**: Reusable React components
- **config**: Configuration utilities

### `/infrastructure/` - DevOps
All deployment and infrastructure code:
- **helm**: Primary deployment method
- **docker**: Container definitions
- **scripts**: Automation tools
- **k8s**: Raw Kubernetes manifests

### `/docs/` - Documentation
Project documentation including:
- Architecture decisions
- User stories and requirements
- Troubleshooting guides

### `/tests/` - Testing
Cross-service integration and end-to-end tests:
- **unit**: Component/function level tests
- **integration**: Service integration tests
- **e2e**: Full user journey tests
- **performance**: Load and performance tests

## Key Design Principles

### Container-First Development
- All development happens in containers
- No local `pnpm run dev` or `go run` processes
- Development through Kubernetes deployment

### Monorepo Organization
- Shared code in `/packages/`
- Clear separation of concerns
- Consistent tooling across all projects

### Infrastructure as Code
- Helm charts for deployment
- Raw Kubernetes manifests for debugging
- Automated environment setup

### Testing Strategy
- Tests at multiple levels
- Container-based test execution
- Integration and E2E testing

This structure ensures scalability, maintainability, and consistency across the KubeChat project.

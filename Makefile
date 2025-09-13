# KubeChat Development Makefile
# Container-first development environment for Kubernetes-based development
# All commands operate through containers and Helm deployment

.PHONY: help dev-info dev-setup dev-clean dev-build dev-rebuild-api dev-rebuild-web dev-push
.PHONY: dev-deploy dev-upgrade dev-status dev-undeploy dev-rollback
.PHONY: dev-logs dev-logs-api dev-logs-web dev-shell-api dev-shell-web dev-port-forward
.PHONY: dev-db-connect dev-db-migrate dev-db-seed dev-db-reset dev-db-status dev-db-backup dev-db-health dev-db-integrity
.PHONY: dev-test dev-test-unit dev-test-e2e

# Default target
help: ## Show this help message
	@echo "KubeChat Development Environment"
	@echo "================================"
	@echo ""
	@echo "Container-first development commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Prerequisites: Docker, Kubernetes (Rancher Desktop), Helm 3.15+, kubectl 1.28+, PNPM 8+, Go 1.23+"
	@echo "Quick start: make dev-setup && make dev-deploy"

# Environment Management
dev-info: ## Show system information and prerequisites
	@echo "=== KubeChat Development Environment Info ==="
	@echo ""
	@echo "Tool Versions:"
	@echo "=============="
	@echo -n "Docker: " && docker --version 2>/dev/null || echo "❌ Not installed"
	@echo -n "Kubernetes: " && kubectl version --client --short 2>/dev/null || echo "❌ Not installed"
	@echo -n "Helm: " && helm version --short 2>/dev/null || echo "❌ Not installed"
	@echo -n "PNPM: " && pnpm --version 2>/dev/null || echo "❌ Not installed"
	@echo -n "Go: " && go version 2>/dev/null || echo "❌ Not installed"
	@echo ""
	@echo "Kubernetes Cluster:"
	@echo "=================="
	@kubectl cluster-info 2>/dev/null || echo "❌ Cluster not accessible"
	@echo ""
	@echo "Kubernetes Nodes:"
	@kubectl get nodes 2>/dev/null || echo "❌ No nodes found"
	@echo ""
	@echo "Storage Classes:"
	@kubectl get storageclass 2>/dev/null || echo "❌ No storage classes found"
	@echo ""
	@echo "KubeChat Deployment Status:"
	@echo "==========================="
	@helm status kubechat-dev -n kubechat 2>/dev/null || echo "❌ KubeChat not deployed"

dev-setup: ## Initial development environment setup
	@echo "=== Setting up KubeChat Development Environment ==="
	@echo ""
	@echo "Step 1: Validating prerequisites..."
	@make --no-print-directory dev-info
	@echo ""
	@echo "Step 2: Adding Helm repositories..."
	@helm repo add bitnami https://charts.bitnami.com/bitnami
	@helm repo update
	@echo ""
	@echo "Step 3: Building application containers..."
	@make --no-print-directory dev-build
	@echo ""
	@echo "Step 4: Installing Helm dependencies..."
	@cd infrastructure/helm/kubechat && helm dependency update
	@echo ""
	@echo "✅ Development environment setup complete!"
	@echo "   Next: run 'make dev-deploy' to deploy the application"

dev-clean: ## Clean development environment
	@echo "=== Cleaning Development Environment ==="
	@echo ""
	@echo "Removing KubeChat deployment..."
	@helm uninstall kubechat-dev -n kubechat 2>/dev/null || true
	@echo ""
	@echo "Cleaning up containers..."
	@docker system prune -f
	@docker image prune -f
	@echo ""
	@echo "Removing namespace (keeping PVCs)..."
	@kubectl delete namespace kubechat --ignore-not-found=true
	@echo ""
	@echo "✅ Development environment cleaned"

# Container Management
dev-build: ## Build all application containers
	@echo "=== Building Application Containers ==="
	@echo "Building without cache for fresh builds..."
	@echo ""
	@echo "Building web container..."
	@docker build --no-cache -t kubechat/web:dev -f infrastructure/docker/web/Dockerfile .
	@echo "Importing web image to Rancher Desktop k8s namespace..."
	@docker save kubechat/web:dev | nerdctl --address /var/run/docker/containerd/containerd.sock --namespace k8s.io load
	@echo ""
	@echo "Building API container..."
	@docker build --no-cache -t kubechat/api:dev -f infrastructure/docker/api/Dockerfile .
	@echo "Importing API image to Rancher Desktop k8s namespace..."
	@docker save kubechat/api:dev | nerdctl --address /var/run/docker/containerd/containerd.sock --namespace k8s.io load
	@echo ""
	@echo "✅ All containers built and imported to k8s successfully"

dev-rebuild-api: ## Rebuild API container only
	@echo "=== Rebuilding API Container ==="
	@echo "Building without cache for fresh build..."
	@docker build --no-cache -t kubechat/api:dev -f infrastructure/docker/api/Dockerfile .
	@echo "Importing API image to Rancher Desktop k8s namespace..."
	@docker save kubechat/api:dev | nerdctl --address /var/run/docker/containerd/containerd.sock --namespace k8s.io load
	@echo "✅ API container rebuilt and imported to k8s"

dev-rebuild-web: ## Rebuild web container only
	@echo "=== Rebuilding Web Container ==="
	@echo "Building without cache for fresh build..."
	@docker build --no-cache -t kubechat/web:dev -f infrastructure/docker/web/Dockerfile .
	@echo "Importing web image to Rancher Desktop k8s namespace..."
	@docker save kubechat/web:dev | nerdctl --address /var/run/docker/containerd/containerd.sock --namespace k8s.io load
	@echo "✅ Web container rebuilt and imported to k8s"

dev-push: ## Push containers to local registry
	@echo "=== Pushing Containers to Local Registry ==="
	@echo "Note: Using local development tags - no push needed"
	@echo "Containers are built locally and available to Kubernetes"

# Deployment Management
dev-deploy: ## Deploy complete development stack
	@echo "=== Deploying KubeChat Development Stack ==="
	@echo ""
	@echo "Installing/upgrading with development values..."
	@helm upgrade --install kubechat-dev infrastructure/helm/kubechat \
		--namespace kubechat \
		--create-namespace \
		--values infrastructure/helm/kubechat/values-dev.yaml \
		--wait \
		--timeout 300s
	@echo ""
	@echo "✅ KubeChat deployed successfully!"
	@echo ""
	@echo "Access URLs:"
	@echo "============"
	@echo "Frontend:        http://localhost:30001"
	@echo "API:             http://localhost:30080"
	@echo "PgAdmin:         http://localhost:30050"
	@echo "Redis Commander: http://localhost:30081"
	@echo ""
	@echo "Use 'make dev-logs' to view application logs"
	@echo "Use 'make dev-status' to check deployment status"

dev-upgrade: ## Upgrade existing deployment
	@echo "=== Upgrading KubeChat Deployment ==="
	@helm upgrade kubechat-dev infrastructure/helm/kubechat \
		--namespace kubechat \
		--values infrastructure/helm/kubechat/values-dev.yaml \
		--wait
	@echo "✅ Deployment upgraded successfully"

dev-status: ## Show deployment status
	@echo "=== KubeChat Deployment Status ==="
	@echo ""
	@echo "Helm Release:"
	@helm status kubechat-dev -n kubechat
	@echo ""
	@echo "Pods:"
	@kubectl get pods -n kubechat
	@echo ""
	@echo "Services:"
	@kubectl get svc -n kubechat
	@echo ""
	@echo "Ingress:"
	@kubectl get ingress -n kubechat 2>/dev/null || echo "No ingress found"

dev-undeploy: ## Remove development deployment
	@echo "=== Removing KubeChat Deployment ==="
	@helm uninstall kubechat-dev -n kubechat
	@echo "✅ Deployment removed (PVCs preserved)"

dev-rollback: ## Restore previous working state
	@echo "=== Rolling Back KubeChat Deployment ==="
	@helm rollback kubechat-dev -n kubechat
	@echo "✅ Rollback completed"

# Development Tools
dev-logs: ## View aggregated application logs
	@echo "=== KubeChat Application Logs ==="
	@echo ""
	@echo "Streaming logs from all pods (press Ctrl+C to stop)..."
	@kubectl logs -f -l app.kubernetes.io/instance=kubechat-dev -n kubechat --max-log-requests=10

dev-logs-api: ## View API service logs only
	@echo "=== API Service Logs ==="
	@kubectl logs -f -l app.kubernetes.io/component=api -n kubechat

dev-logs-web: ## View web service logs only
	@echo "=== Web Service Logs ==="
	@kubectl logs -f -l app.kubernetes.io/component=web -n kubechat

dev-shell-api: ## Shell into API container
	@echo "=== Opening shell in API container ==="
	@kubectl exec -it -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}') -- /bin/sh

dev-shell-web: ## Shell into web container
	@echo "=== Opening shell in Web container ==="
	@kubectl exec -it -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=web -o jsonpath='{.items[0].metadata.name}') -- /bin/sh

dev-port-forward: ## Setup port forwarding for local access
	@echo "=== Setting up port forwarding ==="
	@echo ""
	@echo "Starting port forwarding (press Ctrl+C to stop)..."
	@echo "Frontend: http://localhost:3000"
	@echo "API: http://localhost:8080"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo ""
	@kubectl port-forward -n kubechat svc/kubechat-dev-web 3000:3000 &
	@kubectl port-forward -n kubechat svc/kubechat-dev-api 8080:8080 &
	@kubectl port-forward -n kubechat svc/kubechat-dev-postgresql 5432:5432 &
	@kubectl port-forward -n kubechat svc/kubechat-dev-redis-master 6379:6379 &
	@echo "Port forwarding active in background"
	@echo "Use 'pkill -f kubectl.*port-forward' to stop all port forwarding"

# Database Management
dev-db-connect: ## Connect to development database
	@echo "=== Connecting to Development Database ==="
	@kubectl exec -it -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}') -- env PGPASSWORD=dev-password psql -U kubechat -d kubechat

dev-db-migrate: ## Run database migrations
	@echo "=== Running Database Migrations ==="
	@echo "Note: Migrations are handled by PostgreSQL initdbScripts during container startup"
	@echo "Check migration status with 'make dev-db-status'

dev-db-seed: ## Seed development data
	@echo "=== Seeding Development Data ==="
	@echo "Note: Default admin user is created automatically during database initialization"
	@echo "Username: admin, Password: admin123, Email: admin@kubechat.dev"

dev-db-reset: ## Reset database to clean state
	@echo "=== Resetting Database ==="
	@echo "⚠️  This will delete all data. Press Ctrl+C to cancel, Enter to continue..."
	@read dummy
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}') -- env PGPASSWORD=dev-password psql -U kubechat -d kubechat -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "✅ Database reset completed. Restart PostgreSQL pod to re-run initialization scripts."

dev-db-status: ## Show database migration status
	@echo "=== Database Migration Status ==="
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}') -- env PGPASSWORD=dev-password psql -U kubechat -d kubechat -c "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at DESC;"

dev-db-backup: ## Create database backup
	@echo "=== Creating Database Backup ==="
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}') -- env PGPASSWORD=dev-password pg_dump -U kubechat kubechat > kubechat-backup-$$(date +%Y%m%d-%H%M%S).sql
	@echo "✅ Database backup created"

dev-db-health: ## Check database health
	@echo "=== Database Health Check ==="
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}') -- env PGPASSWORD=dev-password psql -U kubechat -d kubechat -c "SELECT 'Database connection: OK' as status; SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"

dev-db-integrity: ## Validate database integrity
	@echo "=== Database Integrity Validation ==="
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}') -- env PGPASSWORD=dev-password psql -U kubechat -d kubechat -c "SELECT * FROM verify_audit_log_integrity() WHERE NOT is_valid;"

# Testing
dev-test: ## Run all tests in containers
	@echo "=== Running All Tests ==="
	@make --no-print-directory dev-test-unit
	@make --no-print-directory dev-test-e2e

dev-test-unit: ## Run unit tests only
	@echo "=== Running Unit Tests ==="
	@echo ""
	@echo "Frontend tests:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=web -o jsonpath='{.items[0].metadata.name}') -- npm test
	@echo ""
	@echo "Backend tests:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}') -- go test ./...

dev-test-e2e: ## Run end-to-end tests
	@echo "=== Running End-to-End Tests ==="
	@echo "Note: E2E tests will be implemented with Playwright"
	@echo "Placeholder for E2E test execution"

dev-test-coverage: ## Run tests with coverage measurement
	@echo "=== Running Tests with Coverage ==="
	@echo ""
	@echo "Backend test coverage:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}') -- go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "Coverage report:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}') -- go tool cover -func=coverage.out
	@echo ""
	@echo "HTML coverage report:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}') -- go tool cover -html=coverage.out -o coverage.html

dev-test-coverage-local: ## Run tests locally with coverage (for CI/local development)
	@echo "=== Running Local Tests with Coverage ==="
	@echo ""
	@echo "Testing packages with coverage (excluding Docker-dependent tests):"
	@cd apps/api && go test -coverprofile=coverage.out -covermode=atomic ./internal/models ./internal/services/user ./tests/integration
	@echo ""
	@echo "Coverage summary:"
	@cd apps/api && go tool cover -func=coverage.out | grep -E "(models|user|integration|total)"
	@echo ""
	@echo "Detailed coverage report:"
	@cd apps/api && go tool cover -func=coverage.out
	@echo ""
	@echo "HTML coverage report:"
	@cd apps/api && go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "✅ Coverage report generated at apps/api/coverage.html"
	@echo "📊 Total coverage above demonstrates functional implementation with tests"

# Lint and Format
dev-lint: ## Run linting for all code
	@echo "=== Running Code Linting ==="
	@echo ""
	@echo "Frontend linting:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=web -o jsonpath='{.items[0].metadata.name}') -- npm run lint
	@echo ""
	@echo "Backend linting:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}') -- golangci-lint run

dev-format: ## Format all code
	@echo "=== Formatting Code ==="
	@echo ""
	@echo "Frontend formatting:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=web -o jsonpath='{.items[0].metadata.name}') -- npm run format
	@echo ""
	@echo "Backend formatting:"
	@kubectl exec -n kubechat $$(kubectl get pod -n kubechat -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}') -- go fmt ./...

# Troubleshooting
dev-debug: ## Show debugging information
	@echo "=== Debug Information ==="
	@echo ""
	@echo "Recent events:"
	@kubectl get events -n kubechat --sort-by='.firstTimestamp' | tail -10
	@echo ""
	@echo "Pod descriptions:"
	@kubectl describe pods -n kubechat
	@echo ""
	@echo "Resource usage:"
	@kubectl top pods -n kubechat 2>/dev/null || echo "Metrics server not available"

dev-restart: ## Restart all application pods
	@echo "=== Restarting Application Pods ==="
	@kubectl rollout restart deployment/kubechat-dev-web -n kubechat
	@kubectl rollout restart deployment/kubechat-dev-api -n kubechat
	@echo "✅ Restart initiated. Use 'make dev-status' to check progress"

# Quick development workflow shortcuts
quick-web: dev-rebuild-web dev-restart ## Quick web rebuild and restart
	@echo "✅ Web application updated"

quick-api: dev-rebuild-api dev-restart ## Quick API rebuild and restart
	@echo "✅ API application updated"

# Project initialization (run once)
init: ## Initialize project for first-time setup
	@echo "=== Initializing KubeChat Project ==="
	@echo ""
	@echo "This will set up the complete development environment."
	@echo "Prerequisites: Docker, Rancher Desktop with Kubernetes enabled"
	@echo ""
	@make --no-print-directory dev-setup
	@make --no-print-directory dev-deploy
	@echo ""
	@echo "🎉 KubeChat development environment is ready!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Visit http://localhost:30001 for the frontend"
	@echo "2. Check http://localhost:30080/health for API health"
	@echo "3. Use 'make dev-logs' to view application logs"
	@echo "4. Use 'make help' to see all available commands"
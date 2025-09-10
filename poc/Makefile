# KubeChat Makefile

.PHONY: help build run test clean docker-build docker-run install-deps

# Default target
help:
	@echo "KubeChat - Natural Language Kubernetes Management Platform"
	@echo ""
	@echo "Available commands:"
	@echo "  install-deps    Install Go and Node.js dependencies"
	@echo "  build          Build Go backend and React frontend"
	@echo "  run            Run the application locally"
	@echo "  test           Run tests"
	@echo "  clean          Clean build artifacts"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo "  helm-install   Install using Helm chart"
	@echo "  helm-uninstall Uninstall Helm release"

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download
	@echo "Installing Node.js dependencies..."
	cd web && npm install

# Build the application
build: build-frontend build-backend

build-frontend:
	@echo "Building React frontend..."
	cd web && npm run build

build-backend:
	@echo "Building Go backend..."
	go build -o bin/kubechat ./cmd/server

# Run the application locally
run: build
	@echo "Starting KubeChat server..."
	@echo "Frontend will be served at: http://localhost:8080"
	@echo "API endpoints available at: http://localhost:8080/api/*"
	./bin/kubechat

# Run development servers
dev-frontend:
	@echo "Starting React development server..."
	cd web && npm start

dev-backend:
	@echo "Starting Go development server..."
	go run ./cmd/server

# Run tests
test:
	@echo "Running Go tests..."
	go test ./...
	@echo "Running frontend tests..."
	cd web && npm test -- --coverage --watchAll=false

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/build/
	rm -rf web/node_modules/

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t kubechat:latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --rm \
		-e LLM_PROVIDER=ollama \
		-e OLLAMA_URL=http://host.docker.internal:11434 \
		kubechat:latest

# Helm commands
helm-install:
	@echo "Installing KubeChat using Helm..."
	helm install kubechat ./chart \
		--set llm.provider=ollama \
		--set llm.ollama.enabled=true \
		--set llm.ollama.deploy=true

helm-uninstall:
	@echo "Uninstalling KubeChat..."
	helm uninstall kubechat

# Setup Ollama (for local development)
setup-ollama:
	@echo "Setting up Ollama..."
	@echo "1. Install Ollama: https://ollama.ai"
	@echo "2. Pull a model: ollama pull llama2"
	@echo "3. Verify: curl http://localhost:11434/api/tags"

# Quick start
quickstart: install-deps build
	@echo ""
	@echo "🚀 KubeChat is ready!"
	@echo ""
	@echo "To start the application:"
	@echo "  make run"
	@echo ""
	@echo "To run in development mode:"
	@echo "  Terminal 1: make dev-backend"
	@echo "  Terminal 2: make dev-frontend"
	@echo ""
	@echo "Make sure you have:"
	@echo "  - Ollama running: http://localhost:11434"
	@echo "  - Kubernetes cluster accessible (Rancher Desktop)"
	@echo ""
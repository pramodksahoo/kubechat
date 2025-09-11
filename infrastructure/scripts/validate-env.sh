#!/bin/bash
# KubeChat Development Environment Validation Script
# Validates all prerequisites and environment setup for KubeChat development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Validation results
VALIDATION_PASSED=0
VALIDATION_FAILED=0

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    VALIDATION_PASSED=$((VALIDATION_PASSED + 1))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    VALIDATION_FAILED=$((VALIDATION_FAILED + 1))
}

# Validation functions
check_command() {
    local cmd=$1
    local name=$2
    local required_version=$3
    
    if command -v $cmd >/dev/null 2>&1; then
        local version=$($cmd --version 2>/dev/null | head -n1 || echo "unknown")
        log_success "$name is installed: $version"
        return 0
    else
        log_error "$name is not installed (required: $required_version)"
        return 1
    fi
}

check_docker() {
    log_info "Checking Docker..."
    
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker is not installed"
        log_error "Install Docker Desktop or Rancher Desktop"
        return 1
    fi
    
    local docker_version=$(docker --version 2>/dev/null)
    log_success "Docker is installed: $docker_version"
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon is not running"
        log_error "Start Docker Desktop or Rancher Desktop"
        return 1
    fi
    
    log_success "Docker daemon is running"
    return 0
}

check_kubernetes() {
    log_info "Checking Kubernetes..."
    
    if ! command -v kubectl >/dev/null 2>&1; then
        log_error "kubectl is not installed"
        log_error "Install kubectl 1.28+ or use Rancher Desktop"
        return 1
    fi
    
    local kubectl_version=$(kubectl version --client --short 2>/dev/null)
    log_success "kubectl is installed: $kubectl_version"
    
    # Check if Kubernetes cluster is accessible
    if ! kubectl cluster-info >/dev/null 2>&1; then
        log_error "Kubernetes cluster is not accessible"
        log_error "Start Rancher Desktop and enable Kubernetes"
        return 1
    fi
    
    local cluster_info=$(kubectl cluster-info 2>/dev/null | head -n1)
    log_success "Kubernetes cluster is accessible: $cluster_info"
    
    # Check nodes
    local nodes=$(kubectl get nodes --no-headers 2>/dev/null | wc -l)
    if [ $nodes -eq 0 ]; then
        log_error "No Kubernetes nodes found"
        return 1
    fi
    
    log_success "Kubernetes nodes found: $nodes"
    
    # Check for Ready nodes
    local ready_nodes=$(kubectl get nodes --no-headers 2>/dev/null | grep -c "Ready" || echo "0")
    if [ $ready_nodes -eq 0 ]; then
        log_error "No Ready Kubernetes nodes found"
        return 1
    fi
    
    log_success "Ready Kubernetes nodes: $ready_nodes"
    return 0
}

check_helm() {
    log_info "Checking Helm..."
    
    if ! command -v helm >/dev/null 2>&1; then
        log_error "Helm is not installed"
        log_error "Install Helm 3.15+ using: curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
        return 1
    fi
    
    local helm_version=$(helm version --short 2>/dev/null)
    log_success "Helm is installed: $helm_version"
    
    # Check Helm version (should be 3.x)
    local major_version=$(helm version --short 2>/dev/null | grep -o "v3\." || echo "")
    if [ -z "$major_version" ]; then
        log_warning "Helm version should be 3.15+ for best compatibility"
    fi
    
    return 0
}

check_node_pnpm() {
    log_info "Checking Node.js and PNPM..."
    
    # Check Node.js
    if ! command -v node >/dev/null 2>&1; then
        log_error "Node.js is not installed"
        log_error "Install Node.js 20+ from https://nodejs.org/"
        return 1
    fi
    
    local node_version=$(node --version 2>/dev/null)
    log_success "Node.js is installed: $node_version"
    
    # Check Node.js version (should be 20+)
    local node_major=$(node --version 2>/dev/null | grep -o "v[0-9]*" | grep -o "[0-9]*")
    if [ "$node_major" -lt 20 ]; then
        log_warning "Node.js version should be 20+ for optimal compatibility"
    fi
    
    # Check PNPM
    if ! command -v pnpm >/dev/null 2>&1; then
        log_error "PNPM is not installed"
        log_error "Install PNPM using: curl -fsSL https://get.pnpm.io/install.sh | sh"
        return 1
    fi
    
    local pnpm_version=$(pnpm --version 2>/dev/null)
    log_success "PNPM is installed: v$pnpm_version"
    
    return 0
}

check_go() {
    log_info "Checking Go..."
    
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go is not installed"
        log_error "Install Go 1.23+ from https://go.dev/dl/"
        return 1
    fi
    
    local go_version=$(go version 2>/dev/null)
    log_success "Go is installed: $go_version"
    
    # Check Go version (should be 1.23+)
    local go_ver=$(go version 2>/dev/null | grep -o "go[0-9]\+\.[0-9]\+" | grep -o "[0-9]\+\.[0-9]\+")
    local go_major=$(echo $go_ver | cut -d. -f1)
    local go_minor=$(echo $go_ver | cut -d. -f2)
    
    if [ "$go_major" -lt 1 ] || ([ "$go_major" -eq 1 ] && [ "$go_minor" -lt 23 ]); then
        log_warning "Go version should be 1.23+ for optimal compatibility"
    fi
    
    return 0
}

check_storage_class() {
    log_info "Checking Kubernetes storage classes..."
    
    local storage_classes=$(kubectl get storageclass --no-headers 2>/dev/null | wc -l)
    if [ $storage_classes -eq 0 ]; then
        log_error "No storage classes found"
        log_error "Ensure Rancher Desktop or your Kubernetes cluster has storage classes configured"
        return 1
    fi
    
    log_success "Storage classes found: $storage_classes"
    
    # Check for local-path storage class (Rancher Desktop default)
    if kubectl get storageclass local-path >/dev/null 2>&1; then
        log_success "local-path storage class found (Rancher Desktop)"
    else
        log_warning "local-path storage class not found - using default storage class"
    fi
    
    # Show default storage class
    local default_sc=$(kubectl get storageclass 2>/dev/null | grep "(default)" | awk '{print $1}' || echo "none")
    if [ "$default_sc" != "none" ]; then
        log_success "Default storage class: $default_sc"
    else
        log_warning "No default storage class set"
    fi
    
    return 0
}

check_rancher_desktop() {
    log_info "Checking Rancher Desktop specific configuration..."
    
    # Check if this looks like Rancher Desktop
    local context=$(kubectl config current-context 2>/dev/null || echo "")
    if [[ $context == "rancher-desktop" ]]; then
        log_success "Rancher Desktop context detected: $context"
        
        # Check Rancher Desktop specific features
        if kubectl get nodes -o wide 2>/dev/null | grep -q "rancher-desktop"; then
            log_success "Rancher Desktop node configuration detected"
        fi
        
        # Check if local registry is available (if configured)
        if docker images | grep -q "localhost:5000" 2>/dev/null; then
            log_success "Local registry images found"
        else
            log_info "No local registry images found (this is normal)"
        fi
        
    else
        log_warning "Current context is not rancher-desktop: $context"
        log_warning "KubeChat is optimized for Rancher Desktop but should work with other Kubernetes clusters"
    fi
    
    return 0
}

check_network_connectivity() {
    log_info "Checking network connectivity..."
    
    # Check if we can reach Docker Hub (for image pulls)
    if curl -s --connect-timeout 5 https://hub.docker.com >/dev/null 2>&1; then
        log_success "Docker Hub is accessible"
    else
        log_warning "Docker Hub is not accessible - container image pulls may fail"
    fi
    
    # Check if we can reach Helm repository
    if curl -s --connect-timeout 5 https://charts.bitnami.com >/dev/null 2>&1; then
        log_success "Bitnami Helm repository is accessible"
    else
        log_warning "Bitnami Helm repository is not accessible - Helm dependency installation may fail"
    fi
    
    return 0
}

check_ports() {
    log_info "Checking port availability..."
    
    local ports=(3000 8080 5432 6379 30001 30080 30050 30081)
    local blocked_ports=()
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            blocked_ports+=($port)
        fi
    done
    
    if [ ${#blocked_ports[@]} -eq 0 ]; then
        log_success "All required ports are available: ${ports[*]}"
    else
        log_warning "Some ports are in use: ${blocked_ports[*]}"
        log_warning "KubeChat may have port conflicts. Consider stopping other services."
    fi
    
    return 0
}

check_disk_space() {
    log_info "Checking disk space..."
    
    local available_gb=$(df -h . | awk 'NR==2 {print $4}' | sed 's/G.*//')
    if [ "$available_gb" -lt 10 ]; then
        log_warning "Low disk space: ${available_gb}GB available"
        log_warning "KubeChat development requires at least 10GB free space"
    else
        log_success "Sufficient disk space: ${available_gb}GB available"
    fi
    
    return 0
}

check_memory() {
    log_info "Checking system memory..."
    
    if command -v free >/dev/null 2>&1; then
        local memory_gb=$(free -g | awk 'NR==2 {print $2}')
        if [ "$memory_gb" -lt 8 ]; then
            log_warning "Low system memory: ${memory_gb}GB total"
            log_warning "KubeChat development works best with 8GB+ RAM"
        else
            log_success "Sufficient system memory: ${memory_gb}GB total"
        fi
    elif command -v sysctl >/dev/null 2>&1; then
        # macOS
        local memory_bytes=$(sysctl -n hw.memsize 2>/dev/null || echo "0")
        local memory_gb=$((memory_bytes / 1024 / 1024 / 1024))
        if [ "$memory_gb" -lt 8 ]; then
            log_warning "Low system memory: ${memory_gb}GB total"
            log_warning "KubeChat development works best with 8GB+ RAM"
        else
            log_success "Sufficient system memory: ${memory_gb}GB total"
        fi
    else
        log_info "Cannot determine system memory"
    fi
    
    return 0
}

# Main validation function
run_validation() {
    log_info "=== KubeChat Development Environment Validation ==="
    echo
    
    # Core tools validation
    check_docker
    check_kubernetes
    check_helm
    check_node_pnpm
    check_go
    
    echo
    log_info "=== Kubernetes Environment ==="
    check_storage_class
    check_rancher_desktop
    
    echo
    log_info "=== System Resources ==="
    check_disk_space
    check_memory
    check_ports
    check_network_connectivity
    
    echo
    log_info "=== Validation Summary ==="
    log_success "Passed validations: $VALIDATION_PASSED"
    if [ $VALIDATION_FAILED -gt 0 ]; then
        log_error "Failed validations: $VALIDATION_FAILED"
        echo
        log_error "Please fix the failed validations before proceeding with KubeChat development"
        echo
        log_info "For installation instructions, run: make help"
        exit 1
    else
        log_success "All validations passed! ✅"
        echo
        log_success "Your environment is ready for KubeChat development!"
        echo
        log_info "Next steps:"
        echo "  1. Run 'make dev-setup' to initialize the development environment"
        echo "  2. Run 'make dev-deploy' to deploy KubeChat"
        echo "  3. Visit http://localhost:30001 to access the application"
        echo
    fi
}

# Help function
show_help() {
    cat << EOF
KubeChat Development Environment Validation

Usage: $0 [options]

Options:
    --help, -h     Show this help message
    --quiet, -q    Quiet mode (only show errors)
    --verbose, -v  Verbose mode (show detailed information)

This script validates that your development environment meets all
requirements for KubeChat development:

Required Tools:
- Docker (latest)
- Kubernetes 1.28+ (via Rancher Desktop recommended)
- Helm 3.15+
- kubectl 1.28+
- Node.js 20+
- PNPM 8+
- Go 1.23+

System Requirements:
- 8GB+ RAM
- 10GB+ free disk space
- Available ports: 3000, 8080, 5432, 6379, 30001, 30080, 30050, 30081

Network Requirements:
- Access to Docker Hub for container images
- Access to Helm repositories for dependencies

Examples:
    $0              # Run full validation
    $0 --quiet      # Show only errors
    $0 --verbose    # Show detailed information
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --quiet|-q)
            # Redirect info and success messages to /dev/null
            exec 3>&1
            exec 1>/dev/null
            ;;
        --verbose|-v)
            set -x
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
    shift
done

# Run the validation
run_validation
#!/bin/bash
# KubeChat Service Health Check Script
# Monitors the health of all KubeChat services and dependencies

set -e

# Configuration
NAMESPACE=${NAMESPACE:-kubechat}
TIMEOUT=${TIMEOUT:-30}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[HEALTHY]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[UNHEALTHY]${NC} $1"
}

# Health check functions
check_namespace() {
    log_info "Checking namespace: $NAMESPACE"
    
    if kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
        log_success "Namespace $NAMESPACE exists"
        return 0
    else
        log_error "Namespace $NAMESPACE does not exist"
        return 1
    fi
}

check_helm_release() {
    log_info "Checking Helm release: kubechat-dev"
    
    if helm status kubechat-dev -n $NAMESPACE >/dev/null 2>&1; then
        local status=$(helm status kubechat-dev -n $NAMESPACE -o json | jq -r '.info.status' 2>/dev/null || echo "unknown")
        if [ "$status" = "deployed" ]; then
            log_success "Helm release is deployed successfully"
        else
            log_warning "Helm release status: $status"
        fi
        return 0
    else
        log_error "Helm release kubechat-dev not found"
        return 1
    fi
}

check_pods() {
    log_info "Checking pod status..."
    
    local pods=$(kubectl get pods -n $NAMESPACE --no-headers 2>/dev/null)
    if [ -z "$pods" ]; then
        log_error "No pods found in namespace $NAMESPACE"
        return 1
    fi
    
    local total_pods=$(echo "$pods" | wc -l)
    local running_pods=$(echo "$pods" | grep -c "Running" || echo "0")
    local ready_pods=$(echo "$pods" | awk '{if ($2 ~ /^[0-9]+\/[0-9]+$/ && $2 != "0/0") print $2}' | awk -F'/' '{if ($1 == $2) count++} END {print count+0}')
    
    log_info "Total pods: $total_pods"
    log_info "Running pods: $running_pods"
    log_info "Ready pods: $ready_pods"
    
    if [ "$running_pods" -eq "$total_pods" ] && [ "$ready_pods" -eq "$total_pods" ]; then
        log_success "All pods are running and ready"
        return 0
    else
        log_warning "Some pods are not ready"
        kubectl get pods -n $NAMESPACE
        return 1
    fi
}

check_services() {
    log_info "Checking services..."
    
    local services=$(kubectl get svc -n $NAMESPACE --no-headers 2>/dev/null)
    if [ -z "$services" ]; then
        log_error "No services found in namespace $NAMESPACE"
        return 1
    fi
    
    local service_count=$(echo "$services" | wc -l)
    log_success "Services found: $service_count"
    
    # Check specific services
    local required_services=("kubechat-dev-web" "kubechat-dev-api" "kubechat-dev-postgresql" "kubechat-dev-redis-master")
    for service in "${required_services[@]}"; do
        if kubectl get svc $service -n $NAMESPACE >/dev/null 2>&1; then
            log_success "Service $service is available"
        else
            log_warning "Service $service not found"
        fi
    done
    
    return 0
}

check_endpoints() {
    log_info "Checking service endpoints..."
    
    # Web service endpoint
    if kubectl get endpoints kubechat-dev-web -n $NAMESPACE >/dev/null 2>&1; then
        local web_endpoints=$(kubectl get endpoints kubechat-dev-web -n $NAMESPACE -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null)
        if [ -n "$web_endpoints" ]; then
            log_success "Web service has endpoints: $(echo $web_endpoints | wc -w) addresses"
        else
            log_warning "Web service has no endpoints"
        fi
    fi
    
    # API service endpoint
    if kubectl get endpoints kubechat-dev-api -n $NAMESPACE >/dev/null 2>&1; then
        local api_endpoints=$(kubectl get endpoints kubechat-dev-api -n $NAMESPACE -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null)
        if [ -n "$api_endpoints" ]; then
            log_success "API service has endpoints: $(echo $api_endpoints | wc -w) addresses"
        else
            log_warning "API service has no endpoints"
        fi
    fi
    
    return 0
}

check_database() {
    log_info "Checking PostgreSQL database..."
    
    local postgres_pod=$(kubectl get pod -n $NAMESPACE -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -z "$postgres_pod" ]; then
        log_error "PostgreSQL pod not found"
        return 1
    fi
    
    # Check if pod is ready
    if kubectl wait --for=condition=ready pod/$postgres_pod -n $NAMESPACE --timeout=10s >/dev/null 2>&1; then
        log_success "PostgreSQL pod is ready: $postgres_pod"
    else
        log_warning "PostgreSQL pod is not ready: $postgres_pod"
        return 1
    fi
    
    # Check database connectivity
    if kubectl exec -n $NAMESPACE $postgres_pod -- pg_isready -U kubechat >/dev/null 2>&1; then
        log_success "PostgreSQL database is accepting connections"
    else
        log_warning "PostgreSQL database is not accepting connections"
        return 1
    fi
    
    return 0
}

check_redis() {
    log_info "Checking Redis cache..."
    
    local redis_pod=$(kubectl get pod -n $NAMESPACE -l app.kubernetes.io/name=redis,app.kubernetes.io/component=master -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -z "$redis_pod" ]; then
        log_error "Redis master pod not found"
        return 1
    fi
    
    # Check if pod is ready
    if kubectl wait --for=condition=ready pod/$redis_pod -n $NAMESPACE --timeout=10s >/dev/null 2>&1; then
        log_success "Redis pod is ready: $redis_pod"
    else
        log_warning "Redis pod is not ready: $redis_pod"
        return 1
    fi
    
    # Check Redis connectivity
    if kubectl exec -n $NAMESPACE $redis_pod -- redis-cli ping >/dev/null 2>&1; then
        log_success "Redis is responding to ping"
    else
        log_warning "Redis is not responding to ping"
        return 1
    fi
    
    return 0
}

check_ingress() {
    log_info "Checking ingress configuration..."
    
    if kubectl get ingress -n $NAMESPACE >/dev/null 2>&1; then
        local ingress_count=$(kubectl get ingress -n $NAMESPACE --no-headers | wc -l)
        if [ "$ingress_count" -gt 0 ]; then
            log_success "Ingress resources found: $ingress_count"
            
            # Check if ingress has addresses
            local ingress_addresses=$(kubectl get ingress -n $NAMESPACE -o jsonpath='{.items[*].status.loadBalancer.ingress[*].ip}' 2>/dev/null)
            if [ -n "$ingress_addresses" ]; then
                log_success "Ingress has load balancer addresses"
            else
                log_info "Ingress load balancer addresses pending (this is normal for local development)"
            fi
        else
            log_info "No ingress resources found (using NodePort services)"
        fi
    else
        log_info "No ingress resources configured"
    fi
    
    return 0
}

check_volumes() {
    log_info "Checking persistent volumes..."
    
    local pvcs=$(kubectl get pvc -n $NAMESPACE --no-headers 2>/dev/null)
    if [ -z "$pvcs" ]; then
        log_warning "No persistent volume claims found"
        return 0
    fi
    
    local total_pvcs=$(echo "$pvcs" | wc -l)
    local bound_pvcs=$(echo "$pvcs" | grep -c "Bound" || echo "0")
    
    log_info "Total PVCs: $total_pvcs"
    log_info "Bound PVCs: $bound_pvcs"
    
    if [ "$bound_pvcs" -eq "$total_pvcs" ]; then
        log_success "All persistent volumes are bound"
    else
        log_warning "Some persistent volumes are not bound"
        kubectl get pvc -n $NAMESPACE
    fi
    
    return 0
}

check_application_health() {
    log_info "Checking application health endpoints..."
    
    # Check API health endpoint via port-forward
    local api_pod=$(kubectl get pod -n $NAMESPACE -l app.kubernetes.io/component=api -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -n "$api_pod" ]; then
        log_info "Testing API health endpoint..."
        
        # Start port-forward in background
        kubectl port-forward -n $NAMESPACE pod/$api_pod 8080:8080 >/dev/null 2>&1 &
        local port_forward_pid=$!
        
        # Wait a moment for port-forward to establish
        sleep 2
        
        # Test health endpoint
        if curl -s --connect-timeout 5 http://localhost:8080/health >/dev/null 2>&1; then
            log_success "API health endpoint is responding"
        else
            log_warning "API health endpoint is not responding (this is normal if the app is not fully started)"
        fi
        
        # Clean up port-forward
        kill $port_forward_pid 2>/dev/null || true
        wait $port_forward_pid 2>/dev/null || true
    else
        log_warning "API pod not found - cannot test health endpoint"
    fi
    
    return 0
}

check_resource_usage() {
    log_info "Checking resource usage..."
    
    # Check if metrics server is available
    if kubectl top pods -n $NAMESPACE >/dev/null 2>&1; then
        log_success "Metrics server is available"
        kubectl top pods -n $NAMESPACE
    else
        log_info "Metrics server not available (this is normal in development)"
    fi
    
    return 0
}

generate_status_report() {
    log_info "=== KubeChat Status Report ==="
    echo
    
    log_info "Deployment Information:"
    echo "Namespace: $NAMESPACE"
    echo "Helm Release: kubechat-dev"
    echo "Timestamp: $(date)"
    echo
    
    log_info "Quick Status:"
    kubectl get all -n $NAMESPACE
    
    echo
    log_info "Recent Events:"
    kubectl get events -n $NAMESPACE --sort-by='.firstTimestamp' | tail -10
    
    echo
    log_info "Access Information:"
    echo "Frontend: http://localhost:30001"
    echo "API: http://localhost:30080"
    echo "PgAdmin: http://localhost:30050 (admin@kubechat.dev / dev-admin)"
    echo "Redis Commander: http://localhost:30081"
    echo
    echo "Port Forward Commands:"
    echo "  make dev-port-forward  # Start all port forwarding"
    echo "  make dev-logs         # View application logs"
    echo "  make dev-status       # View deployment status"
}

# Main health check function
run_health_check() {
    local exit_code=0
    
    log_info "=== KubeChat Health Check ==="
    echo
    
    # Basic infrastructure checks
    check_namespace || exit_code=1
    check_helm_release || exit_code=1
    
    echo
    log_info "=== Pod and Service Health ==="
    check_pods || exit_code=1
    check_services || exit_code=1
    check_endpoints || exit_code=1
    
    echo
    log_info "=== Database and Cache Health ==="
    check_database || exit_code=1
    check_redis || exit_code=1
    
    echo
    log_info "=== Infrastructure Health ==="
    check_ingress || exit_code=1
    check_volumes || exit_code=1
    
    echo
    log_info "=== Application Health ==="
    check_application_health || exit_code=1
    check_resource_usage || exit_code=1
    
    echo
    if [ $exit_code -eq 0 ]; then
        log_success "All health checks passed! ✅"
    else
        log_warning "Some health checks failed or showed warnings ⚠️"
    fi
    
    return $exit_code
}

# Help function
show_help() {
    cat << EOF
KubeChat Service Health Check

Usage: $0 [options] [command]

Commands:
    check         Run all health checks (default)
    status        Generate detailed status report
    quick         Run quick health check (pods and services only)
    
Options:
    --namespace, -n    Kubernetes namespace (default: kubechat)
    --timeout, -t      Timeout for operations in seconds (default: 30)
    --help, -h         Show this help message

Examples:
    $0                    # Run full health check
    $0 status            # Generate status report
    $0 quick             # Quick check
    $0 -n prod check     # Check production namespace
EOF
}

# Parse command line arguments
COMMAND="check"

while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace|-n)
            NAMESPACE="$2"
            shift 2
            ;;
        --timeout|-t)
            TIMEOUT="$2"
            shift 2
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        check|status|quick)
            COMMAND="$1"
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Execute the requested command
case $COMMAND in
    check)
        run_health_check
        ;;
    status)
        generate_status_report
        ;;
    quick)
        check_namespace
        check_pods
        check_services
        ;;
    *)
        log_error "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac
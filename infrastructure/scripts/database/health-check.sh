#!/bin/bash

# Database Health Check Script
# Performs comprehensive health checks for KubeChat database
# Date: 2025-01-11
# Author: James (Full Stack Developer Agent)

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-kubechat}"
POSTGRES_USER="${POSTGRES_USER:-kubechat}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-kubechat}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Health check results
HEALTH_STATUS="healthy"
ISSUES_FOUND=()

# Function to add issue
add_issue() {
    ISSUES_FOUND+=("$1")
    if [[ "$HEALTH_STATUS" == "healthy" ]]; then
        HEALTH_STATUS="degraded"
    fi
}

# Function to mark as unhealthy
mark_unhealthy() {
    HEALTH_STATUS="unhealthy"
    ISSUES_FOUND+=("$1")
}

# Check database connectivity
check_connectivity() {
    log_info "Checking database connectivity..."
    
    if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT 1;" &> /dev/null; then
        log_success "Database connectivity OK"
    else
        mark_unhealthy "Cannot connect to database"
        return 1
    fi
}

# Check required tables exist
check_schema() {
    log_info "Checking database schema..."
    
    local required_tables=("users" "user_sessions" "audit_logs" "cluster_configs")
    local missing_tables=()
    
    for table in "${required_tables[@]}"; do
        local exists=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = 'public' 
                AND table_name = '${table}'
            );
        " | xargs)
        
        if [[ "${exists}" != "t" ]]; then
            missing_tables+=("${table}")
        fi
    done
    
    if [[ ${#missing_tables[@]} -eq 0 ]]; then
        log_success "All required tables exist"
    else
        mark_unhealthy "Missing tables: $(IFS=', '; echo "${missing_tables[*]}")"
    fi
}

# Check audit log integrity
check_audit_integrity() {
    log_info "Checking audit log integrity..."
    
    local integrity_results=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM verify_audit_log_integrity(NULL) WHERE NOT is_valid;
    " 2>/dev/null | xargs || echo "error")
    
    if [[ "${integrity_results}" == "error" ]]; then
        add_issue "Could not verify audit log integrity"
    elif [[ "${integrity_results}" == "0" ]]; then
        log_success "Audit log integrity OK"
    else
        mark_unhealthy "Found ${integrity_results} audit log integrity violations"
    fi
}

# Check database performance
check_performance() {
    log_info "Checking database performance..."
    
    # Test query response time
    local start_time=$(date +%s%N)
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT COUNT(*) FROM users;" &> /dev/null
    local end_time=$(date +%s%N)
    local duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
    
    if [[ $duration -lt 100 ]]; then
        log_success "Query performance OK (${duration}ms)"
    elif [[ $duration -lt 1000 ]]; then
        log_warning "Query performance slow (${duration}ms)"
        add_issue "Slow query performance: ${duration}ms"
    else
        add_issue "Very slow query performance: ${duration}ms"
    fi
}

# Check connection pool
check_connections() {
    log_info "Checking database connections..."
    
    local active_connections=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT count(*) FROM pg_stat_activity WHERE datname = '${POSTGRES_DB}';
    " | xargs)
    
    local max_connections=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SHOW max_connections;
    " | xargs)
    
    local connection_ratio=$((active_connections * 100 / max_connections))
    
    if [[ $connection_ratio -lt 70 ]]; then
        log_success "Connection usage OK (${active_connections}/${max_connections})"
    elif [[ $connection_ratio -lt 90 ]]; then
        log_warning "High connection usage (${active_connections}/${max_connections})"
        add_issue "High connection usage: ${connection_ratio}%"
    else
        add_issue "Critical connection usage: ${connection_ratio}%"
    fi
}

# Check disk space (if running locally)
check_disk_space() {
    log_info "Checking database disk usage..."
    
    local db_size=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT pg_size_pretty(pg_database_size('${POSTGRES_DB}'));
    " | xargs)
    
    log_info "Database size: ${db_size}"
    
    # Get table sizes
    local large_tables=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT schemaname||'.'||tablename as table, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
        FROM pg_tables 
        WHERE schemaname = 'public'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC 
        LIMIT 5;
    ")
    
    log_info "Largest tables:"
    echo "${large_tables}"
}

# Check for expired sessions
check_expired_sessions() {
    log_info "Checking for expired sessions..."
    
    local expired_count=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM user_sessions WHERE expires_at < NOW();
    " | xargs)
    
    if [[ $expired_count -eq 0 ]]; then
        log_success "No expired sessions found"
    elif [[ $expired_count -lt 10 ]]; then
        log_warning "Found ${expired_count} expired sessions"
        add_issue "Expired sessions need cleanup: ${expired_count}"
    else
        add_issue "Many expired sessions need cleanup: ${expired_count}"
    fi
}

# Check dangerous operations
check_dangerous_operations() {
    log_info "Checking for recent dangerous operations..."
    
    local dangerous_count=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM audit_logs 
        WHERE safety_level = 'dangerous' 
        AND timestamp > NOW() - INTERVAL '24 hours';
    " | xargs)
    
    if [[ $dangerous_count -eq 0 ]]; then
        log_success "No dangerous operations in last 24 hours"
    elif [[ $dangerous_count -lt 5 ]]; then
        log_warning "Found ${dangerous_count} dangerous operations in last 24 hours"
    else
        log_warning "Found ${dangerous_count} dangerous operations in last 24 hours - review audit logs"
        add_issue "High number of dangerous operations: ${dangerous_count}"
    fi
}

# Generate health report
generate_report() {
    echo ""
    echo "========================================="
    echo "         DATABASE HEALTH REPORT         "
    echo "========================================="
    echo "Timestamp: $(date)"
    echo "Host: ${POSTGRES_HOST}:${POSTGRES_PORT}"
    echo "Database: ${POSTGRES_DB}"
    echo ""
    
    case $HEALTH_STATUS in
        "healthy")
            log_success "Overall Status: HEALTHY ✓"
            ;;
        "degraded")
            log_warning "Overall Status: DEGRADED ⚠"
            ;;
        "unhealthy")
            log_error "Overall Status: UNHEALTHY ✗"
            ;;
    esac
    
    if [[ ${#ISSUES_FOUND[@]} -gt 0 ]]; then
        echo ""
        echo "Issues Found:"
        for issue in "${ISSUES_FOUND[@]}"; do
            echo "  - ${issue}"
        done
    fi
    
    echo ""
    echo "Health check completed."
    echo "========================================="
}

# Main execution
main() {
    echo "Starting KubeChat database health check..."
    echo ""
    
    # Run all health checks
    check_connectivity || exit 1
    check_schema
    check_audit_integrity
    check_performance
    check_connections
    check_disk_space
    check_expired_sessions
    check_dangerous_operations
    
    # Generate final report
    generate_report
    
    # Exit with appropriate code
    case $HEALTH_STATUS in
        "healthy")
            exit 0
            ;;
        "degraded")
            exit 1
            ;;
        "unhealthy")
            exit 2
            ;;
    esac
}

# Show help
show_help() {
    cat << EOF
KubeChat Database Health Check Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    --help, -h          Show this help message

ENVIRONMENT VARIABLES:
    POSTGRES_HOST       Database host (default: localhost)
    POSTGRES_PORT       Database port (default: 5432)
    POSTGRES_DB         Database name (default: kubechat)
    POSTGRES_USER       Database user (default: kubechat)
    POSTGRES_PASSWORD   Database password (default: kubechat)

EXIT CODES:
    0                   Healthy
    1                   Degraded (warnings found)
    2                   Unhealthy (critical issues)

EXAMPLES:
    $0                  # Run health check with default settings
    POSTGRES_HOST=db.example.com $0  # Check remote database
EOF
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
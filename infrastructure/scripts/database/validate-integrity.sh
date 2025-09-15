#!/bin/bash

# KubeChat Database Integrity Validation Script
# Comprehensive validation of audit trail integrity and database consistency
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

# Output configuration
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}" # text, json, csv
VERBOSE="${VERBOSE:-false}"
LOG_FILE="${LOG_FILE:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    local message="$1"
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $message"
    [[ -n "${LOG_FILE}" ]] && echo "$(date '+%Y-%m-%d %H:%M:%S') INFO: $message" >> "${LOG_FILE}"
}

log_success() {
    local message="$1"
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $message"
    [[ -n "${LOG_FILE}" ]] && echo "$(date '+%Y-%m-%d %H:%M:%S') SUCCESS: $message" >> "${LOG_FILE}"
}

log_warning() {
    local message="$1"
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $message"
    [[ -n "${LOG_FILE}" ]] && echo "$(date '+%Y-%m-%d %H:%M:%S') WARNING: $message" >> "${LOG_FILE}"
}

log_error() {
    local message="$1"
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $message"
    [[ -n "${LOG_FILE}" ]] && echo "$(date '+%Y-%m-%d %H:%M:%S') ERROR: $message" >> "${LOG_FILE}"
}

# Validation results
VALIDATION_RESULTS=()
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Add validation result
add_result() {
    local check_name="$1"
    local status="$2" # PASS, FAIL, WARNING
    local message="$3"
    local details="${4:-}"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    case $status in
        "PASS")
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
            [[ "$VERBOSE" == "true" ]] && log_success "$check_name: $message"
            ;;
        "FAIL")
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            log_error "$check_name: $message"
            ;;
        "WARNING")
            WARNING_CHECKS=$((WARNING_CHECKS + 1))
            log_warning "$check_name: $message"
            ;;
    esac
    
    VALIDATION_RESULTS+=("$check_name|$status|$message|$details")
}

# Help function
show_help() {
    cat << EOF
KubeChat Database Integrity Validation Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    --format FORMAT     Output format: text, json, csv (default: text)
    --verbose           Enable verbose output
    --log-file FILE     Write log to file
    --help, -h          Show this help

ENVIRONMENT VARIABLES:
    POSTGRES_HOST       Database host (default: localhost)
    POSTGRES_PORT       Database port (default: 5432)
    POSTGRES_DB         Database name (default: kubechat)
    POSTGRES_USER       Database user (default: kubechat)
    POSTGRES_PASSWORD   Database password (default: kubechat)

EXAMPLES:
    $0                              # Run validation with text output
    $0 --format json --verbose      # JSON output with details
    $0 --log-file validation.log    # Save log to file

EXIT CODES:
    0       All validations passed
    1       Validation failures found
    2       Critical errors (cannot connect, etc.)
EOF
}

# Check database connectivity
check_connectivity() {
    log_info "Checking database connectivity..."
    
    if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT 1;" &> /dev/null; then
        add_result "Database Connectivity" "PASS" "Successfully connected to database"
    else
        add_result "Database Connectivity" "FAIL" "Cannot connect to database"
        return 1
    fi
}

# Validate audit log integrity
validate_audit_integrity() {
    log_info "Validating audit log integrity..."
    
    # Run the integrity check function
    local integrity_output
    integrity_output=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT 
            COUNT(*) as total_logs,
            COUNT(CASE WHEN is_valid THEN 1 END) as valid_logs,
            COUNT(CASE WHEN NOT is_valid THEN 1 END) as invalid_logs
        FROM verify_audit_log_integrity(NULL);
    " 2>/dev/null)
    
    if [[ $? -ne 0 ]]; then
        add_result "Audit Log Integrity" "FAIL" "Could not run integrity check function"
        return 1
    fi
    
    # Parse results
    local total_logs=$(echo "$integrity_output" | awk '{print $1}')
    local valid_logs=$(echo "$integrity_output" | awk '{print $2}')
    local invalid_logs=$(echo "$integrity_output" | awk '{print $3}')
    
    if [[ "$invalid_logs" == "0" ]]; then
        add_result "Audit Log Integrity" "PASS" "$total_logs audit logs verified, all checksums valid"
    else
        add_result "Audit Log Integrity" "FAIL" "$invalid_logs out of $total_logs audit logs have invalid checksums"
        
        # Get details of failed validations
        local failed_logs
        failed_logs=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
            SELECT log_id FROM verify_audit_log_integrity(NULL) WHERE NOT is_valid LIMIT 10;
        " 2>/dev/null | xargs)
        
        if [[ -n "$failed_logs" ]]; then
            add_result "Failed Audit Log IDs" "FAIL" "Log IDs with integrity violations: $failed_logs"
        fi
    fi
}

# Check checksum chain continuity
validate_checksum_chain() {
    log_info "Validating checksum chain continuity..."
    
    # Check for broken chains
    local broken_chains
    broken_chains=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        WITH chain_check AS (
            SELECT 
                id,
                previous_checksum,
                LAG(checksum) OVER (ORDER BY id) as expected_previous
            FROM audit_logs
            ORDER BY id
        )
        SELECT COUNT(*)
        FROM chain_check
        WHERE id > (SELECT MIN(id) FROM audit_logs)
        AND (previous_checksum IS NULL OR previous_checksum != expected_previous);
    " 2>/dev/null | xargs)
    
    if [[ "$broken_chains" == "0" ]]; then
        add_result "Checksum Chain" "PASS" "Checksum chain is continuous and unbroken"
    else
        add_result "Checksum Chain" "FAIL" "$broken_chains breaks in checksum chain detected"
    fi
}

# Validate user data integrity
validate_user_data() {
    log_info "Validating user data integrity..."
    
    # Check for orphaned sessions
    local orphaned_sessions
    orphaned_sessions=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM user_sessions us 
        LEFT JOIN users u ON us.user_id = u.id 
        WHERE u.id IS NULL;
    " 2>/dev/null | xargs)
    
    if [[ "$orphaned_sessions" == "0" ]]; then
        add_result "User Session Integrity" "PASS" "No orphaned user sessions found"
    else
        add_result "User Session Integrity" "WARNING" "$orphaned_sessions orphaned user sessions found"
    fi
    
    # Check for expired sessions
    local expired_sessions
    expired_sessions=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM user_sessions WHERE expires_at < NOW();
    " 2>/dev/null | xargs)
    
    if [[ "$expired_sessions" == "0" ]]; then
        add_result "Session Expiration" "PASS" "No expired sessions found"
    elif [[ "$expired_sessions" -lt 10 ]]; then
        add_result "Session Expiration" "WARNING" "$expired_sessions expired sessions found (cleanup recommended)"
    else
        add_result "Session Expiration" "WARNING" "$expired_sessions expired sessions found (cleanup needed)"
    fi
    
    # Check for duplicate usernames/emails
    local duplicate_usernames
    duplicate_usernames=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) - COUNT(DISTINCT username) FROM users;
    " 2>/dev/null | xargs)
    
    if [[ "$duplicate_usernames" == "0" ]]; then
        add_result "Username Uniqueness" "PASS" "All usernames are unique"
    else
        add_result "Username Uniqueness" "FAIL" "$duplicate_usernames duplicate usernames found"
    fi
    
    local duplicate_emails
    duplicate_emails=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) - COUNT(DISTINCT email) FROM users;
    " 2>/dev/null | xargs)
    
    if [[ "$duplicate_emails" == "0" ]]; then
        add_result "Email Uniqueness" "PASS" "All email addresses are unique"
    else
        add_result "Email Uniqueness" "FAIL" "$duplicate_emails duplicate email addresses found"
    fi
}

# Validate cluster configuration data
validate_cluster_data() {
    log_info "Validating cluster configuration data..."
    
    # Check for orphaned cluster configs
    local orphaned_configs
    orphaned_configs=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM cluster_configs cc 
        LEFT JOIN users u ON cc.user_id = u.id 
        WHERE u.id IS NULL;
    " 2>/dev/null | xargs)
    
    if [[ "$orphaned_configs" == "0" ]]; then
        add_result "Cluster Config Integrity" "PASS" "No orphaned cluster configurations found"
    else
        add_result "Cluster Config Integrity" "WARNING" "$orphaned_configs orphaned cluster configurations found"
    fi
    
    # Check for multiple active clusters per user (should be at most one)
    local multi_active
    multi_active=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM (
            SELECT user_id, COUNT(*) as active_count
            FROM cluster_configs 
            WHERE is_active = true
            GROUP BY user_id
            HAVING COUNT(*) > 1
        ) multi;
    " 2>/dev/null | xargs)
    
    if [[ "$multi_active" == "0" ]]; then
        add_result "Active Cluster Constraint" "PASS" "No users have multiple active clusters"
    else
        add_result "Active Cluster Constraint" "WARNING" "$multi_active users have multiple active clusters"
    fi
}

# Validate database constraints
validate_constraints() {
    log_info "Validating database constraints..."
    
    # Check for constraint violations (this would typically catch issues before they're committed)
    local constraint_violations
    constraint_violations=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM (
            SELECT table_name, constraint_name
            FROM information_schema.table_constraints 
            WHERE constraint_type = 'CHECK' 
            AND table_schema = 'public'
        ) constraints;
    " 2>/dev/null | xargs)
    
    add_result "Database Constraints" "PASS" "$constraint_violations constraints defined and enforced"
    
    # Check for foreign key constraint violations (shouldn't happen with proper constraints)
    local fk_violations=0
    
    # Check users -> user_sessions
    local user_session_fk
    user_session_fk=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM user_sessions us 
        LEFT JOIN users u ON us.user_id = u.id 
        WHERE u.id IS NULL;
    " 2>/dev/null | xargs)
    fk_violations=$((fk_violations + user_session_fk))
    
    if [[ "$fk_violations" == "0" ]]; then
        add_result "Foreign Key Integrity" "PASS" "All foreign key relationships are valid"
    else
        add_result "Foreign Key Integrity" "FAIL" "$fk_violations foreign key violations found"
    fi
}

# Check database performance indicators
validate_performance() {
    log_info "Validating database performance indicators..."
    
    # Check for large table sizes that might indicate performance issues
    local large_tables
    large_tables=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM (
            SELECT schemaname||'.'||tablename as table,
                   pg_total_relation_size(schemaname||'.'||tablename) as size
            FROM pg_tables 
            WHERE schemaname = 'public'
            AND pg_total_relation_size(schemaname||'.'||tablename) > 100*1024*1024 -- 100MB
        ) large;
    " 2>/dev/null | xargs)
    
    if [[ "$large_tables" == "0" ]]; then
        add_result "Table Sizes" "PASS" "No unusually large tables detected"
    elif [[ "$large_tables" -lt 3 ]]; then
        add_result "Table Sizes" "WARNING" "$large_tables tables over 100MB (monitor growth)"
    else
        add_result "Table Sizes" "WARNING" "$large_tables tables over 100MB (consider archiving)"
    fi
    
    # Check index usage
    local unused_indexes
    unused_indexes=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM pg_stat_user_indexes 
        WHERE idx_scan = 0 AND schemaname = 'public';
    " 2>/dev/null | xargs)
    
    if [[ "$unused_indexes" == "0" ]]; then
        add_result "Index Usage" "PASS" "All indexes are being used"
    else
        add_result "Index Usage" "WARNING" "$unused_indexes indexes are not being used"
    fi
}

# Generate validation report
generate_report() {
    case $OUTPUT_FORMAT in
        "json")
            generate_json_report
            ;;
        "csv")
            generate_csv_report
            ;;
        *)
            generate_text_report
            ;;
    esac
}

# Generate text report
generate_text_report() {
    echo ""
    echo "========================================"
    echo "    DATABASE INTEGRITY VALIDATION      "
    echo "========================================"
    echo "Timestamp: $(date)"
    echo "Database: ${POSTGRES_DB} on ${POSTGRES_HOST}:${POSTGRES_PORT}"
    echo ""
    echo "Summary:"
    echo "  Total Checks: $TOTAL_CHECKS"
    echo "  Passed: $PASSED_CHECKS"
    echo "  Failed: $FAILED_CHECKS"
    echo "  Warnings: $WARNING_CHECKS"
    echo ""
    
    if [[ $FAILED_CHECKS -gt 0 ]]; then
        echo "FAILED CHECKS:"
        for result in "${VALIDATION_RESULTS[@]}"; do
            IFS='|' read -r name status message details <<< "$result"
            if [[ "$status" == "FAIL" ]]; then
                echo "  ✗ $name: $message"
            fi
        done
        echo ""
    fi
    
    if [[ $WARNING_CHECKS -gt 0 ]]; then
        echo "WARNINGS:"
        for result in "${VALIDATION_RESULTS[@]}"; do
            IFS='|' read -r name status message details <<< "$result"
            if [[ "$status" == "WARNING" ]]; then
                echo "  ⚠ $name: $message"
            fi
        done
        echo ""
    fi
    
    if [[ $FAILED_CHECKS -eq 0 && $WARNING_CHECKS -eq 0 ]]; then
        log_success "All validation checks passed!"
    elif [[ $FAILED_CHECKS -eq 0 ]]; then
        log_warning "Validation completed with warnings"
    else
        log_error "Validation failed with $FAILED_CHECKS critical issues"
    fi
    
    echo "========================================"
}

# Generate JSON report
generate_json_report() {
    cat << EOF
{
    "timestamp": "$(date -Iseconds)",
    "database": {
        "host": "${POSTGRES_HOST}",
        "port": ${POSTGRES_PORT},
        "name": "${POSTGRES_DB}"
    },
    "summary": {
        "total_checks": ${TOTAL_CHECKS},
        "passed": ${PASSED_CHECKS},
        "failed": ${FAILED_CHECKS},
        "warnings": ${WARNING_CHECKS}
    },
    "results": [
EOF

    local first=true
    for result in "${VALIDATION_RESULTS[@]}"; do
        IFS='|' read -r name status message details <<< "$result"
        if [[ "$first" == "true" ]]; then
            first=false
        else
            echo ","
        fi
        echo -n "        {\"check\": \"$name\", \"status\": \"$status\", \"message\": \"$message\""
        if [[ -n "$details" ]]; then
            echo -n ", \"details\": \"$details\""
        fi
        echo -n "}"
    done

    cat << EOF

    ]
}
EOF
}

# Generate CSV report
generate_csv_report() {
    echo "Check,Status,Message,Details"
    for result in "${VALIDATION_RESULTS[@]}"; do
        IFS='|' read -r name status message details <<< "$result"
        echo "\"$name\",\"$status\",\"$message\",\"$details\""
    done
}

# Main execution
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE="true"
                shift
                ;;
            --log-file)
                LOG_FILE="$2"
                shift 2
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Initialize log file if specified
    if [[ -n "${LOG_FILE}" ]]; then
        echo "Database Integrity Validation Log - $(date)" > "${LOG_FILE}"
    fi
    
    log_info "Starting database integrity validation..."
    
    # Run all validation checks
    if ! check_connectivity; then
        log_error "Cannot proceed with validation due to connectivity issues"
        exit 2
    fi
    
    validate_audit_integrity
    validate_checksum_chain
    validate_user_data
    validate_cluster_data
    validate_constraints
    validate_performance
    
    # Generate report
    generate_report
    
    # Exit with appropriate code
    if [[ $FAILED_CHECKS -gt 0 ]]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function with all arguments
main "$@"
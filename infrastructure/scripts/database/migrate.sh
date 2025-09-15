#!/bin/bash

# Database Migration Management Script
# Compatible with golang-migrate and container-first development
# Date: 2025-01-11
# Author: James (Full Stack Developer Agent)

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="${SCRIPT_DIR}/migrations"
DATABASE_URL="${DATABASE_URL:-}"
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

# Help function
show_help() {
    cat << EOF
KubeChat Database Migration Management Script

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    up [N]           Apply all pending migrations or N migrations
    down [N]         Rollback N migrations (default: 1)
    version          Show current migration version
    force VERSION    Force set migration version (use with caution)
    create NAME      Create new migration files
    status           Show migration status
    validate         Validate migration files
    reset            Reset database (drop all tables)
    help             Show this help

EXAMPLES:
    $0 up                    # Apply all pending migrations
    $0 up 1                  # Apply next 1 migration
    $0 down                  # Rollback 1 migration
    $0 down 2                # Rollback 2 migrations
    $0 create add_user_table # Create new migration
    $0 status                # Show current status
    $0 validate              # Validate all migrations

ENVIRONMENT VARIABLES:
    DATABASE_URL            Full database connection string
    POSTGRES_HOST           Database host (default: localhost)
    POSTGRES_PORT           Database port (default: 5432)
    POSTGRES_DB             Database name (default: kubechat)
    POSTGRES_USER           Database user (default: kubechat)
    POSTGRES_PASSWORD       Database password (default: kubechat)

CONTAINER-FIRST USAGE:
    # From container with kubectl port-forward
    export POSTGRES_HOST=localhost
    export POSTGRES_PORT=5432
    $0 up

    # Direct kubernetes pod execution
    kubectl exec -it deployment/kubechat-api -- /app/scripts/migrate.sh up
EOF
}

# Build database URL if not provided
build_database_url() {
    if [[ -z "${DATABASE_URL}" ]]; then
        DATABASE_URL="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
    fi
}

# Check if psql is available
check_dependencies() {
    if ! command -v psql &> /dev/null; then
        log_error "psql command not found. Please install PostgreSQL client."
        exit 1
    fi
}

# Check database connectivity
check_database_connection() {
    log_info "Checking database connectivity..."
    build_database_url
    
    if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT 1;" &> /dev/null; then
        log_success "Database connection successful"
    else
        log_error "Cannot connect to database. Please check your connection parameters."
        log_info "Host: ${POSTGRES_HOST}, Port: ${POSTGRES_PORT}, DB: ${POSTGRES_DB}, User: ${POSTGRES_USER}"
        exit 1
    fi
}

# Create schema_migrations table if it doesn't exist
ensure_migrations_table() {
    log_info "Ensuring schema_migrations table exists..."
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version BIGINT PRIMARY KEY,
            dirty BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT NOW()
        );
        CREATE INDEX IF NOT EXISTS idx_schema_migrations_version ON schema_migrations(version);
        CREATE INDEX IF NOT EXISTS idx_schema_migrations_dirty ON schema_migrations(dirty);
    " &> /dev/null
}

# Get current migration version
get_current_version() {
    ensure_migrations_table
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE NOT dirty;
    " 2>/dev/null | xargs || echo "0"
}

# Get pending migrations
get_pending_migrations() {
    local current_version=$(get_current_version)
    find "${MIGRATIONS_DIR}" -name "*.sql" -not -name "*.down.sql" | \
        sed 's/.*\/\([0-9]*\)_.*/\1/' | \
        sort -n | \
        awk -v current="${current_version}" '$1 > current'
}

# Apply migration
apply_migration() {
    local version=$1
    local migration_file="${MIGRATIONS_DIR}/${version}_*.sql"
    local actual_file=$(ls ${migration_file} 2>/dev/null | head -1)
    
    if [[ ! -f "${actual_file}" ]]; then
        log_error "Migration file not found: ${migration_file}"
        return 1
    fi
    
    log_info "Applying migration ${version}: $(basename "${actual_file}")"
    
    # Mark as dirty first
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "
        INSERT INTO schema_migrations (version, dirty) VALUES (${version}, TRUE) 
        ON CONFLICT (version) DO UPDATE SET dirty = TRUE;
    " || return 1
    
    # Apply migration
    if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -f "${actual_file}"; then
        # Mark as clean
        PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "
            UPDATE schema_migrations SET dirty = FALSE WHERE version = ${version};
        "
        log_success "Migration ${version} applied successfully"
        return 0
    else
        log_error "Migration ${version} failed"
        return 1
    fi
}

# Rollback migration
rollback_migration() {
    local version=$1
    local down_file="${MIGRATIONS_DIR}/${version}_*.down.sql"
    local actual_file=$(ls ${down_file} 2>/dev/null | head -1)
    
    if [[ ! -f "${actual_file}" ]]; then
        log_error "Rollback file not found: ${down_file}"
        return 1
    fi
    
    log_info "Rolling back migration ${version}: $(basename "${actual_file}")"
    
    # Apply rollback
    if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -f "${actual_file}"; then
        # Remove from migrations table
        PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "
            DELETE FROM schema_migrations WHERE version = ${version};
        "
        log_success "Migration ${version} rolled back successfully"
        return 0
    else
        log_error "Rollback ${version} failed"
        return 1
    fi
}

# Commands
cmd_up() {
    local limit=${1:-}
    check_database_connection
    
    local pending_migrations=($(get_pending_migrations))
    
    if [[ ${#pending_migrations[@]} -eq 0 ]]; then
        log_info "No pending migrations"
        return 0
    fi
    
    if [[ -n "${limit}" ]]; then
        pending_migrations=(${pending_migrations[@]:0:${limit}})
    fi
    
    log_info "Applying ${#pending_migrations[@]} migration(s)"
    
    for version in "${pending_migrations[@]}"; do
        if ! apply_migration "${version}"; then
            log_error "Migration failed. Database may be in inconsistent state."
            exit 1
        fi
    done
    
    log_success "All migrations applied successfully"
}

cmd_down() {
    local count=${1:-1}
    check_database_connection
    
    local current_version=$(get_current_version)
    if [[ "${current_version}" == "0" ]]; then
        log_info "No migrations to rollback"
        return 0
    fi
    
    # Get last N applied migrations in reverse order
    local applied_migrations=($(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT version FROM schema_migrations WHERE NOT dirty ORDER BY version DESC LIMIT ${count};
    " | xargs))
    
    log_info "Rolling back ${#applied_migrations[@]} migration(s)"
    
    for version in "${applied_migrations[@]}"; do
        if ! rollback_migration "${version}"; then
            log_error "Rollback failed. Database may be in inconsistent state."
            exit 1
        fi
    done
    
    log_success "Rollback completed successfully"
}

cmd_version() {
    check_database_connection
    local current_version=$(get_current_version)
    echo "Current migration version: ${current_version}"
}

cmd_status() {
    check_database_connection
    
    local current_version=$(get_current_version)
    local pending_migrations=($(get_pending_migrations))
    
    echo "Current migration version: ${current_version}"
    echo "Pending migrations: ${#pending_migrations[@]}"
    
    if [[ ${#pending_migrations[@]} -gt 0 ]]; then
        echo "Pending:"
        for version in "${pending_migrations[@]}"; do
            local file=$(ls "${MIGRATIONS_DIR}/${version}_"*.sql 2>/dev/null | head -1)
            echo "  ${version} - $(basename "${file}")"
        done
    fi
    
    # Check for dirty migrations
    local dirty_count=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT COUNT(*) FROM schema_migrations WHERE dirty;
    " 2>/dev/null | xargs || echo "0")
    
    if [[ "${dirty_count}" -gt 0 ]]; then
        log_warning "Found ${dirty_count} dirty migration(s). Database may be in inconsistent state."
    fi
}

cmd_create() {
    local name=$1
    if [[ -z "${name}" ]]; then
        log_error "Migration name is required"
        exit 1
    fi
    
    local timestamp=$(date +%s)
    local up_file="${MIGRATIONS_DIR}/${timestamp}_${name}.sql"
    local down_file="${MIGRATIONS_DIR}/${timestamp}_${name}.down.sql"
    
    cat > "${up_file}" << EOF
-- Migration ${timestamp}: ${name}
-- Date: $(date '+%Y-%m-%d')
-- Author: Generated by migrate.sh

-- Add your migration SQL here

EOF

    cat > "${down_file}" << EOF
-- Migration ${timestamp} DOWN: ${name}
-- Date: $(date '+%Y-%m-%d')
-- Author: Generated by migrate.sh

-- Add your rollback SQL here

EOF

    log_success "Created migration files:"
    echo "  Up:   ${up_file}"
    echo "  Down: ${down_file}"
}

cmd_validate() {
    log_info "Validating migration files..."
    
    local error_count=0
    
    # Check for migration files
    local migration_files=($(find "${MIGRATIONS_DIR}" -name "*.sql" -not -name "*.down.sql" | sort))
    
    if [[ ${#migration_files[@]} -eq 0 ]]; then
        log_warning "No migration files found"
        return 0
    fi
    
    # Validate each migration has corresponding down file
    for up_file in "${migration_files[@]}"; do
        local base_name=$(basename "${up_file}" .sql)
        local down_file="${MIGRATIONS_DIR}/${base_name}.down.sql"
        
        if [[ ! -f "${down_file}" ]]; then
            log_error "Missing down migration: ${down_file}"
            ((error_count++))
        fi
    done
    
    if [[ ${error_count} -eq 0 ]]; then
        log_success "All migrations are valid"
    else
        log_error "Found ${error_count} validation error(s)"
        exit 1
    fi
}

cmd_force() {
    local version=$1
    if [[ -z "${version}" ]]; then
        log_error "Version is required"
        exit 1
    fi
    
    check_database_connection
    log_warning "Force setting migration version to ${version}"
    
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "
        DELETE FROM schema_migrations;
        INSERT INTO schema_migrations (version, dirty) VALUES (${version}, FALSE);
    "
    
    log_success "Migration version set to ${version}"
}

cmd_reset() {
    check_database_connection
    log_warning "This will drop all tables and reset the database!"
    read -p "Are you sure? (yes/no): " confirm
    
    if [[ "${confirm}" == "yes" ]]; then
        log_info "Resetting database..."
        PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "
            DROP SCHEMA public CASCADE;
            CREATE SCHEMA public;
            GRANT ALL ON SCHEMA public TO ${POSTGRES_USER};
            GRANT ALL ON SCHEMA public TO public;
        "
        log_success "Database reset completed"
    else
        log_info "Reset cancelled"
    fi
}

# Main execution
main() {
    check_dependencies
    
    case "${1:-help}" in
        up)
            cmd_up "${2:-}"
            ;;
        down)
            cmd_down "${2:-1}"
            ;;
        version)
            cmd_version
            ;;
        status)
            cmd_status
            ;;
        create)
            cmd_create "${2:-}"
            ;;
        validate)
            cmd_validate
            ;;
        force)
            cmd_force "${2:-}"
            ;;
        reset)
            cmd_reset
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
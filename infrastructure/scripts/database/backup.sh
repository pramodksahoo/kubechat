#!/bin/bash

# KubeChat Database Backup Script
# Automated backup with retention policies and integrity validation
# Date: 2025-01-11
# Author: James (Full Stack Developer Agent)

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/kubechat}"
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-kubechat}"
POSTGRES_USER="${POSTGRES_USER:-kubechat}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-kubechat}"

# Backup configuration
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
BACKUP_COMPRESSION="${BACKUP_COMPRESSION:-true}"
BACKUP_VERIFY="${BACKUP_VERIFY:-true}"
S3_BUCKET="${S3_BUCKET:-}"
S3_PREFIX="${S3_PREFIX:-kubechat-backups}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Help function
show_help() {
    cat << EOF
KubeChat Database Backup Script

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    backup              Create a new backup
    restore FILE        Restore from backup file
    list                List available backups
    cleanup             Remove old backups according to retention policy
    verify FILE         Verify backup integrity
    help                Show this help

OPTIONS:
    --compress          Enable compression (default: true)
    --no-verify         Skip backup verification
    --retention DAYS    Set retention period in days (default: 30)

ENVIRONMENT VARIABLES:
    BACKUP_DIR              Backup directory (default: /var/backups/kubechat)
    POSTGRES_HOST           Database host (default: localhost)
    POSTGRES_PORT           Database port (default: 5432)
    POSTGRES_DB             Database name (default: kubechat)
    POSTGRES_USER           Database user (default: kubechat)
    POSTGRES_PASSWORD       Database password (default: kubechat)
    BACKUP_RETENTION_DAYS   Retention period in days (default: 30)
    S3_BUCKET              S3 bucket for remote backup storage
    S3_PREFIX              S3 key prefix (default: kubechat-backups)

EXAMPLES:
    $0 backup                           # Create a new backup
    $0 restore backup-20250111-120000.sql.gz  # Restore from backup
    $0 cleanup                          # Clean old backups
    $0 verify backup-20250111-120000.sql.gz   # Verify backup
EOF
}

# Ensure backup directory exists
ensure_backup_dir() {
    if [[ ! -d "${BACKUP_DIR}" ]]; then
        log_info "Creating backup directory: ${BACKUP_DIR}"
        mkdir -p "${BACKUP_DIR}"
    fi
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v pg_dump &> /dev/null; then
        missing_deps+=("pg_dump")
    fi
    
    if ! command -v psql &> /dev/null; then
        missing_deps+=("psql")
    fi
    
    if [[ "${BACKUP_COMPRESSION}" == "true" ]] && ! command -v gzip &> /dev/null; then
        missing_deps+=("gzip")
    fi
    
    if [[ -n "${S3_BUCKET}" ]] && ! command -v aws &> /dev/null; then
        missing_deps+=("aws cli")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: $(IFS=', '; echo "${missing_deps[*]}")"
        exit 1
    fi
}

# Test database connection
test_connection() {
    log_info "Testing database connection..."
    
    if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT 1;" &> /dev/null; then
        log_success "Database connection successful"
    else
        log_error "Cannot connect to database"
        exit 1
    fi
}

# Get backup filename
get_backup_filename() {
    local timestamp=$(date '+%Y%m%d-%H%M%S')
    local filename="kubechat-backup-${timestamp}.sql"
    
    if [[ "${BACKUP_COMPRESSION}" == "true" ]]; then
        filename="${filename}.gz"
    fi
    
    echo "${filename}"
}

# Create database backup
create_backup() {
    ensure_backup_dir
    test_connection
    
    local backup_file="${BACKUP_DIR}/$(get_backup_filename)"
    local temp_file="${backup_file}.tmp"
    
    log_info "Creating backup: $(basename "${backup_file}")"
    
    # Create the backup
    local pg_dump_cmd="PGPASSWORD='${POSTGRES_PASSWORD}' pg_dump -h '${POSTGRES_HOST}' -p '${POSTGRES_PORT}' -U '${POSTGRES_USER}' -d '${POSTGRES_DB}' --verbose --clean --if-exists --create"
    
    if [[ "${BACKUP_COMPRESSION}" == "true" ]]; then
        eval "${pg_dump_cmd}" | gzip > "${temp_file}"
    else
        eval "${pg_dump_cmd}" > "${temp_file}"
    fi
    
    # Check if backup was successful
    if [[ $? -eq 0 ]]; then
        mv "${temp_file}" "${backup_file}"
        log_success "Backup created successfully: $(basename "${backup_file}")"
        
        # Get backup size
        local backup_size=$(du -h "${backup_file}" | cut -f1)
        log_info "Backup size: ${backup_size}"
        
        # Verify backup if enabled
        if [[ "${BACKUP_VERIFY}" == "true" ]]; then
            verify_backup "${backup_file}"
        fi
        
        # Upload to S3 if configured
        if [[ -n "${S3_BUCKET}" ]]; then
            upload_to_s3 "${backup_file}"
        fi
        
        # Create metadata file
        create_backup_metadata "${backup_file}"
        
        echo "${backup_file}"
    else
        log_error "Backup failed"
        rm -f "${temp_file}"
        exit 1
    fi
}

# Create backup metadata
create_backup_metadata() {
    local backup_file="$1"
    local metadata_file="${backup_file}.meta"
    
    cat > "${metadata_file}" << EOF
{
    "backup_file": "$(basename "${backup_file}")",
    "timestamp": "$(date -Iseconds)",
    "database": "${POSTGRES_DB}",
    "host": "${POSTGRES_HOST}",
    "port": ${POSTGRES_PORT},
    "size_bytes": $(stat -f%z "${backup_file}" 2>/dev/null || stat -c%s "${backup_file}"),
    "checksum": "$(sha256sum "${backup_file}" | cut -d' ' -f1)",
    "compressed": $([ "${BACKUP_COMPRESSION}" == "true" ] && echo "true" || echo "false"),
    "verified": $([ "${BACKUP_VERIFY}" == "true" ] && echo "true" || echo "false")
}
EOF
    
    log_info "Created metadata file: $(basename "${metadata_file}")"
}

# Verify backup integrity
verify_backup() {
    local backup_file="$1"
    
    log_info "Verifying backup integrity: $(basename "${backup_file}")"
    
    if [[ ! -f "${backup_file}" ]]; then
        log_error "Backup file not found: ${backup_file}"
        return 1
    fi
    
    # Test if file is readable and not corrupted
    if [[ "${backup_file}" == *.gz ]]; then
        if gzip -t "${backup_file}"; then
            log_success "Backup file compression is valid"
        else
            log_error "Backup file is corrupted (compression test failed)"
            return 1
        fi
        
        # Check if SQL content is valid
        if gzip -dc "${backup_file}" | head -n 10 | grep -q "PostgreSQL database dump"; then
            log_success "Backup content appears valid"
        else
            log_error "Backup content does not appear to be a valid PostgreSQL dump"
            return 1
        fi
    else
        if head -n 10 "${backup_file}" | grep -q "PostgreSQL database dump"; then
            log_success "Backup content appears valid"
        else
            log_error "Backup content does not appear to be a valid PostgreSQL dump"
            return 1
        fi
    fi
    
    log_success "Backup verification completed successfully"
}

# Upload backup to S3
upload_to_s3() {
    local backup_file="$1"
    local s3_key="${S3_PREFIX}/$(basename "${backup_file}")"
    
    log_info "Uploading backup to S3: s3://${S3_BUCKET}/${s3_key}"
    
    if aws s3 cp "${backup_file}" "s3://${S3_BUCKET}/${s3_key}"; then
        log_success "Backup uploaded to S3 successfully"
        
        # Upload metadata file too
        local metadata_file="${backup_file}.meta"
        if [[ -f "${metadata_file}" ]]; then
            aws s3 cp "${metadata_file}" "s3://${S3_BUCKET}/${s3_key}.meta"
        fi
    else
        log_error "Failed to upload backup to S3"
    fi
}

# List available backups
list_backups() {
    ensure_backup_dir
    
    log_info "Available backups in ${BACKUP_DIR}:"
    echo ""
    
    local backups=($(find "${BACKUP_DIR}" -name "kubechat-backup-*.sql*" -not -name "*.meta" -not -name "*.tmp" | sort -r))
    
    if [[ ${#backups[@]} -eq 0 ]]; then
        echo "No backups found."
        return
    fi
    
    printf "%-30s %-10s %-20s %-10s\n" "Backup File" "Size" "Date" "Verified"
    printf "%-30s %-10s %-20s %-10s\n" "----------" "----" "----" "--------"
    
    for backup in "${backups[@]}"; do
        local filename=$(basename "${backup}")
        local size=$(du -h "${backup}" | cut -f1)
        local date=$(stat -f%Sm -t "%Y-%m-%d %H:%M" "${backup}" 2>/dev/null || stat -c%y "${backup}" | cut -d' ' -f1,2 | cut -d'.' -f1)
        local verified="No"
        
        # Check if backup was verified
        local metadata_file="${backup}.meta"
        if [[ -f "${metadata_file}" ]] && grep -q '"verified": true' "${metadata_file}"; then
            verified="Yes"
        fi
        
        printf "%-30s %-10s %-20s %-10s\n" "${filename}" "${size}" "${date}" "${verified}"
    done
}

# Restore from backup
restore_backup() {
    local backup_file="$1"
    
    if [[ ! -f "${backup_file}" ]]; then
        # Try to find backup in backup directory
        if [[ -f "${BACKUP_DIR}/${backup_file}" ]]; then
            backup_file="${BACKUP_DIR}/${backup_file}"
        else
            log_error "Backup file not found: ${backup_file}"
            exit 1
        fi
    fi
    
    log_warning "This will COMPLETELY REPLACE the current database!"
    log_warning "Database: ${POSTGRES_DB} on ${POSTGRES_HOST}:${POSTGRES_PORT}"
    log_warning "Backup file: $(basename "${backup_file}")"
    echo ""
    read -p "Are you absolutely sure you want to continue? (yes/no): " confirm
    
    if [[ "${confirm}" != "yes" ]]; then
        log_info "Restore cancelled"
        exit 0
    fi
    
    log_info "Verifying backup before restore..."
    if ! verify_backup "${backup_file}"; then
        log_error "Backup verification failed. Aborting restore."
        exit 1
    fi
    
    log_info "Starting database restore from: $(basename "${backup_file}")"
    test_connection
    
    # Restore the backup
    if [[ "${backup_file}" == *.gz ]]; then
        log_info "Decompressing and restoring backup..."
        if gzip -dc "${backup_file}" | PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres; then
            log_success "Database restore completed successfully"
        else
            log_error "Database restore failed"
            exit 1
        fi
    else
        log_info "Restoring backup..."
        if PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d postgres -f "${backup_file}"; then
            log_success "Database restore completed successfully"
        else
            log_error "Database restore failed"
            exit 1
        fi
    fi
    
    log_info "Verifying restored database..."
    test_connection
    
    # Run basic validation
    local table_count=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "
        SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';
    " | xargs)
    
    log_info "Restored database has ${table_count} tables"
    
    if [[ $table_count -gt 0 ]]; then
        log_success "Database restore verification passed"
    else
        log_error "Database restore verification failed - no tables found"
        exit 1
    fi
}

# Cleanup old backups
cleanup_backups() {
    ensure_backup_dir
    
    log_info "Cleaning up backups older than ${BACKUP_RETENTION_DAYS} days..."
    
    local old_backups=($(find "${BACKUP_DIR}" -name "kubechat-backup-*.sql*" -mtime +${BACKUP_RETENTION_DAYS}))
    
    if [[ ${#old_backups[@]} -eq 0 ]]; then
        log_info "No old backups to clean up"
        return
    fi
    
    log_info "Found ${#old_backups[@]} old backup(s) to remove:"
    for backup in "${old_backups[@]}"; do
        echo "  - $(basename "${backup}")"
    done
    
    read -p "Remove these backups? (yes/no): " confirm
    
    if [[ "${confirm}" == "yes" ]]; then
        for backup in "${old_backups[@]}"; do
            rm -f "${backup}"
            rm -f "${backup}.meta"
            log_info "Removed: $(basename "${backup}")"
        done
        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# Main execution
main() {
    check_dependencies
    
    case "${1:-help}" in
        backup)
            create_backup
            ;;
        restore)
            if [[ -z "${2:-}" ]]; then
                log_error "Backup file is required for restore"
                show_help
                exit 1
            fi
            restore_backup "$2"
            ;;
        list)
            list_backups
            ;;
        cleanup)
            cleanup_backups
            ;;
        verify)
            if [[ -z "${2:-}" ]]; then
                log_error "Backup file is required for verification"
                show_help
                exit 1
            fi
            verify_backup "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: ${1:-}"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
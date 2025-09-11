#!/bin/bash
# KubeChat Database Migration Script
# Handles database migrations and schema updates

set -e

# Configuration
NAMESPACE=${NAMESPACE:-kubechat}
POSTGRES_POD=$(kubectl get pod -n $NAMESPACE -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}')
DB_USER=${DB_USER:-kubechat}
DB_NAME=${DB_NAME:-kubechat}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if PostgreSQL pod is running
check_postgres() {
    log_info "Checking PostgreSQL pod status..."
    if [ -z "$POSTGRES_POD" ]; then
        log_error "PostgreSQL pod not found in namespace $NAMESPACE"
        exit 1
    fi
    
    kubectl wait --for=condition=ready pod/$POSTGRES_POD -n $NAMESPACE --timeout=60s
    log_info "PostgreSQL pod is ready: $POSTGRES_POD"
}

# Execute SQL file
execute_sql() {
    local sql_file=$1
    local description=$2
    
    log_info "$description"
    if kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -f /tmp/$(basename $sql_file); then
        log_info "✅ $description completed successfully"
    else
        log_error "❌ $description failed"
        exit 1
    fi
}

# Copy SQL file to pod
copy_sql_to_pod() {
    local sql_file=$1
    
    if [ ! -f "$sql_file" ]; then
        log_error "SQL file not found: $sql_file"
        exit 1
    fi
    
    log_info "Copying $sql_file to PostgreSQL pod..."
    kubectl cp $sql_file $NAMESPACE/$POSTGRES_POD:/tmp/$(basename $sql_file)
}

# Initialize database
init_database() {
    log_info "=== Initializing KubeChat Database ==="
    copy_sql_to_pod "infrastructure/scripts/database/init.sql"
    execute_sql "/tmp/init.sql" "Creating database schema and tables"
}

# Seed database with development data
seed_database() {
    log_info "=== Seeding Database with Development Data ==="
    copy_sql_to_pod "infrastructure/scripts/database/seed.sql"
    execute_sql "/tmp/seed.sql" "Inserting development seed data"
}

# Run database migrations
run_migrations() {
    log_info "=== Running Database Migrations ==="
    
    # Create migrations table if it doesn't exist
    kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -c "
        CREATE TABLE IF NOT EXISTS migrations (
            id SERIAL PRIMARY KEY,
            filename VARCHAR(255) UNIQUE NOT NULL,
            executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );"
    
    # Find all migration files
    migrations_dir="infrastructure/scripts/database/migrations"
    if [ -d "$migrations_dir" ]; then
        for migration_file in $migrations_dir/*.sql; do
            if [ -f "$migration_file" ]; then
                filename=$(basename "$migration_file")
                
                # Check if migration has already been run
                already_run=$(kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM migrations WHERE filename='$filename';" | tr -d ' ')
                
                if [ "$already_run" -eq "0" ]; then
                    log_info "Running migration: $filename"
                    copy_sql_to_pod "$migration_file"
                    execute_sql "/tmp/$filename" "Executing migration $filename"
                    
                    # Record migration as executed
                    kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -c "INSERT INTO migrations (filename) VALUES ('$filename');"
                else
                    log_info "Migration already executed: $filename"
                fi
            fi
        done
    else
        log_warn "No migrations directory found at $migrations_dir"
    fi
}

# Reset database (WARNING: This will delete all data)
reset_database() {
    log_warn "=== RESETTING DATABASE (THIS WILL DELETE ALL DATA) ==="
    echo "Are you sure you want to reset the database? This will delete all data!"
    echo "Type 'YES' to confirm:"
    read -r confirmation
    
    if [ "$confirmation" != "YES" ]; then
        log_info "Database reset cancelled"
        exit 0
    fi
    
    log_warn "Dropping all tables..."
    kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -c "
        DROP SCHEMA IF EXISTS kubechat CASCADE;
        DROP TABLE IF EXISTS migrations;
    "
    
    log_info "Reinitializing database..."
    init_database
    
    log_info "Reseeding database..."
    seed_database
    
    log_info "✅ Database reset completed"
}

# Backup database
backup_database() {
    local backup_name="kubechat-backup-$(date +%Y%m%d-%H%M%S).sql"
    log_info "=== Creating Database Backup: $backup_name ==="
    
    kubectl exec -n $NAMESPACE $POSTGRES_POD -- pg_dump -U $DB_USER $DB_NAME > "$backup_name"
    log_info "✅ Backup created: $backup_name"
}

# Show database status
show_status() {
    log_info "=== Database Status ==="
    echo
    log_info "PostgreSQL Version:"
    kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -c "SELECT version();"
    
    echo
    log_info "Database Size:"
    kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -c "
        SELECT 
            pg_size_pretty(pg_database_size('$DB_NAME')) AS database_size;"
    
    echo
    log_info "Tables:"
    kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -c "
        SELECT 
            schemaname,
            tablename,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
        FROM pg_tables 
        WHERE schemaname = 'kubechat'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
    
    echo
    log_info "Recent Migrations:"
    kubectl exec -n $NAMESPACE $POSTGRES_POD -- psql -U $DB_USER -d $DB_NAME -c "
        SELECT filename, executed_at 
        FROM migrations 
        ORDER BY executed_at DESC 
        LIMIT 10;" 2>/dev/null || log_warn "No migrations table found"
}

# Main function
main() {
    local command=${1:-help}
    
    case $command in
        init)
            check_postgres
            init_database
            ;;
        seed)
            check_postgres
            seed_database
            ;;
        migrate)
            check_postgres
            run_migrations
            ;;
        reset)
            check_postgres
            reset_database
            ;;
        backup)
            check_postgres
            backup_database
            ;;
        status)
            check_postgres
            show_status
            ;;
        full)
            check_postgres
            init_database
            run_migrations
            seed_database
            ;;
        help|*)
            cat << EOF
KubeChat Database Migration Tool

Usage: $0 [command]

Commands:
    init      Initialize database schema and tables
    seed      Insert development seed data
    migrate   Run pending migrations
    reset     Reset database (WARNING: deletes all data)
    backup    Create database backup
    status    Show database status and information
    full      Full setup (init + migrate + seed)
    help      Show this help message

Environment Variables:
    NAMESPACE   Kubernetes namespace (default: kubechat)
    DB_USER     Database user (default: kubechat)
    DB_NAME     Database name (default: kubechat)

Examples:
    $0 init                  # Initialize database
    $0 migrate              # Run migrations
    $0 seed                 # Add development data
    $0 full                 # Complete setup
    NAMESPACE=prod $0 status # Check production database
EOF
            ;;
    esac
}

# Run main function
main "$@"
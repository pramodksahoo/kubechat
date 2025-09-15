#!/bin/bash
# update-dependencies.sh - Update Helm chart dependencies with version locking
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CHART_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${CHART_DIR}/backup"

echo -e "${GREEN}KubeChat Helm Dependency Update Script${NC}"
echo "Chart directory: ${CHART_DIR}"

# Function to log messages
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

# Validate helm is installed
if ! command -v helm &> /dev/null; then
    error "Helm is not installed. Please install Helm 3.15+ first."
    exit 1
fi

# Check helm version
HELM_VERSION=$(helm version --short --client | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')
log "Using Helm version: ${HELM_VERSION}"

# Create backup directory
mkdir -p "${BACKUP_DIR}"

# Backup existing Chart.lock if it exists
if [[ -f "${CHART_DIR}/Chart.lock" ]]; then
    BACKUP_FILE="${BACKUP_DIR}/Chart.lock.$(date +%Y%m%d_%H%M%S)"
    cp "${CHART_DIR}/Chart.lock" "${BACKUP_FILE}"
    log "Backed up existing Chart.lock to ${BACKUP_FILE}"
fi

# Change to chart directory
cd "${CHART_DIR}"

# Update Helm repositories
log "Updating Helm repositories..."
helm repo add bitnami https://charts.bitnami.com/bitnami --force-update
helm repo update

# Verify chart dependencies
log "Verifying Chart.yaml dependencies..."
if ! helm dependency list . 2>/dev/null; then
    warning "No dependencies found or Chart.yaml issues detected"
fi

# Update dependencies
log "Updating dependencies..."
if helm dependency update .; then
    log "Dependencies updated successfully"
else
    error "Failed to update dependencies"
    exit 1
fi

# Verify the updated Chart.lock
if [[ -f "Chart.lock" ]]; then
    log "Chart.lock file updated:"
    cat Chart.lock
else
    warning "Chart.lock file was not created"
fi

# Download dependency charts
log "Downloading dependency charts..."
if helm dependency build .; then
    log "Dependency charts downloaded successfully"
else
    error "Failed to download dependency charts"
    exit 1
fi

# Verify charts directory
if [[ -d "charts" ]]; then
    log "Downloaded charts:"
    ls -la charts/
else
    warning "Charts directory not found"
fi

# Lint the updated chart
log "Linting chart with updated dependencies..."
if helm lint .; then
    log "Chart linting passed"
else
    error "Chart linting failed"
    exit 1
fi

# Template validation for each environment
ENVIRONMENTS=("dev" "staging" "production")
for env in "${ENVIRONMENTS[@]}"; do
    VALUES_FILE="values-${env}.yaml"
    if [[ -f "${VALUES_FILE}" ]]; then
        log "Validating templates for ${env} environment..."
        if helm template "kubechat-${env}" . -f "${VALUES_FILE}" > /dev/null; then
            log "Template validation passed for ${env}"
        else
            error "Template validation failed for ${env}"
            exit 1
        fi
    else
        if [[ "${env}" == "dev" ]]; then
            VALUES_FILE="values.yaml"
            log "Validating templates for ${env} environment (using values.yaml)..."
            if helm template "kubechat-${env}" . -f "${VALUES_FILE}" > /dev/null; then
                log "Template validation passed for ${env}"
            else
                error "Template validation failed for ${env}"
                exit 1
            fi
        else
            warning "Values file ${VALUES_FILE} not found, skipping validation"
        fi
    fi
done

# Dependency security check
log "Checking dependency versions for security issues..."
POSTGRESQL_VERSION=$(grep -A 1 "name: postgresql" Chart.lock | grep "version:" | awk '{print $2}')
REDIS_VERSION=$(grep -A 1 "name: redis" Chart.lock | grep "version:" | awk '{print $2}')

log "PostgreSQL chart version: ${POSTGRESQL_VERSION}"
log "Redis chart version: ${REDIS_VERSION}"

# Check for minimum required versions
MIN_POSTGRESQL_VERSION="15.0.0"
MIN_REDIS_VERSION="19.0.0"

version_gt() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1";
}

if version_gt "${MIN_POSTGRESQL_VERSION}" "${POSTGRESQL_VERSION}"; then
    error "PostgreSQL chart version ${POSTGRESQL_VERSION} is below minimum required ${MIN_POSTGRESQL_VERSION}"
    exit 1
fi

if version_gt "${MIN_REDIS_VERSION}" "${REDIS_VERSION}"; then
    error "Redis chart version ${REDIS_VERSION} is below minimum required ${MIN_REDIS_VERSION}"
    exit 1
fi

log "All dependency versions meet security requirements"

# Generate dependency report
REPORT_FILE="${CHART_DIR}/dependency-report.txt"
cat > "${REPORT_FILE}" << EOF
KubeChat Helm Chart Dependency Report
Generated: $(date)
Chart Version: $(grep '^version:' Chart.yaml | awk '{print $2}')

Dependencies:
$(helm dependency list . 2>/dev/null || echo "No dependencies found")

Chart.lock Content:
$(cat Chart.lock 2>/dev/null || echo "Chart.lock not found")

Security Status:
- PostgreSQL: ${POSTGRESQL_VERSION} (>= ${MIN_POSTGRESQL_VERSION}) ✓
- Redis: ${REDIS_VERSION} (>= ${MIN_REDIS_VERSION}) ✓

Validation Results:
- Chart lint: PASSED ✓
- Template validation: PASSED ✓
EOF

log "Dependency report generated: ${REPORT_FILE}"

echo
log "Dependency update completed successfully!"
log "Next steps:"
echo "  1. Review the Chart.lock file for dependency versions"
echo "  2. Test deployment with: helm upgrade --install kubechat-dev . -f values-dev.yaml"
echo "  3. Commit the updated Chart.lock to version control"
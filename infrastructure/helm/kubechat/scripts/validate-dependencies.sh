#!/bin/bash
# validate-dependencies.sh - Validate Helm chart dependencies and security
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CHART_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${BLUE}KubeChat Helm Dependency Validation Script${NC}"
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

info() {
    echo -e "${BLUE}[INFO] $1${NC}"
}

# Validation results
VALIDATION_ERRORS=0
VALIDATION_WARNINGS=0

# Function to increment error count
add_error() {
    ((VALIDATION_ERRORS++))
    error "$1"
}

# Function to increment warning count
add_warning() {
    ((VALIDATION_WARNINGS++))
    warning "$1"
}

# Change to chart directory
cd "${CHART_DIR}"

log "Starting dependency validation..."

# Check if Chart.yaml exists
if [[ ! -f "Chart.yaml" ]]; then
    add_error "Chart.yaml not found"
    exit 1
fi

# Check if Chart.lock exists
if [[ ! -f "Chart.lock" ]]; then
    add_error "Chart.lock not found. Run 'helm dependency update' first."
    exit 1
fi

# Validate Chart.yaml syntax
log "Validating Chart.yaml syntax..."
if helm show chart . > /dev/null 2>&1; then
    log "Chart.yaml syntax is valid"
else
    add_error "Chart.yaml syntax is invalid"
fi

# Check dependency versions in Chart.yaml vs Chart.lock
log "Checking version consistency between Chart.yaml and Chart.lock..."

# Extract PostgreSQL version from Chart.yaml
PG_CHART_VERSION=$(grep -A 5 "name: postgresql" Chart.yaml | grep "version:" | awk '{print $2}' | head -1)
# Extract PostgreSQL version from Chart.lock
PG_LOCK_VERSION=$(grep -A 2 "name: postgresql" Chart.lock | grep "version:" | awk '{print $2}')

if [[ "${PG_CHART_VERSION}" != "${PG_LOCK_VERSION}" ]]; then
    add_error "PostgreSQL version mismatch: Chart.yaml=${PG_CHART_VERSION}, Chart.lock=${PG_LOCK_VERSION}"
else
    log "PostgreSQL version consistent: ${PG_CHART_VERSION}"
fi

# Extract Redis version from Chart.yaml
REDIS_CHART_VERSION=$(grep -A 5 "name: redis" Chart.yaml | grep "version:" | awk '{print $2}' | head -1)
# Extract Redis version from Chart.lock
REDIS_LOCK_VERSION=$(grep -A 2 "name: redis" Chart.lock | grep "version:" | awk '{print $2}')

if [[ "${REDIS_CHART_VERSION}" != "${REDIS_LOCK_VERSION}" ]]; then
    add_error "Redis version mismatch: Chart.yaml=${REDIS_CHART_VERSION}, Chart.lock=${REDIS_LOCK_VERSION}"
else
    log "Redis version consistent: ${REDIS_CHART_VERSION}"
fi

# Check for security vulnerabilities in dependency versions
log "Checking for known security issues in dependency versions..."

# Known vulnerable versions (examples - should be updated with real CVE data)
declare -A VULNERABLE_VERSIONS
VULNERABLE_VERSIONS["postgresql"]="14.0.0|14.1.0|15.0.0"
VULNERABLE_VERSIONS["redis"]="18.0.0|18.1.0"

version_in_range() {
    local version="$1"
    local vulnerable_range="$2"
    echo "${vulnerable_range}" | grep -q "${version}"
}

# Check PostgreSQL for vulnerabilities
if version_in_range "${PG_LOCK_VERSION}" "${VULNERABLE_VERSIONS[postgresql]:-}"; then
    add_warning "PostgreSQL version ${PG_LOCK_VERSION} has known security vulnerabilities"
else
    log "PostgreSQL version ${PG_LOCK_VERSION} appears secure"
fi

# Check Redis for vulnerabilities
if version_in_range "${REDIS_LOCK_VERSION}" "${VULNERABLE_VERSIONS[redis]:-}"; then
    add_warning "Redis version ${REDIS_LOCK_VERSION} has known security vulnerabilities"
else
    log "Redis version ${REDIS_LOCK_VERSION} appears secure"
fi

# Validate chart dependencies are downloadable
log "Validating dependencies are downloadable..."
if helm dependency list . | grep -q "missing"; then
    add_error "Some dependencies are missing. Run 'helm dependency update'"
else
    log "All dependencies are available"
fi

# Check charts directory
if [[ -d "charts" ]]; then
    CHART_COUNT=$(find charts -name "*.tgz" | wc -l)
    log "Found ${CHART_COUNT} dependency chart(s) in charts/ directory"

    # List downloaded charts
    info "Downloaded dependency charts:"
    find charts -name "*.tgz" -exec basename {} \; | sort
else
    add_warning "Charts directory not found. Dependencies may not be downloaded."
fi

# Validate repository URLs are accessible
log "Validating repository accessibility..."
while IFS= read -r repo_url; do
    if [[ -n "${repo_url}" ]]; then
        info "Checking repository: ${repo_url}"
        if curl -s --head "${repo_url}/index.yaml" > /dev/null; then
            log "Repository accessible: ${repo_url}"
        else
            add_warning "Repository may be inaccessible: ${repo_url}"
        fi
    fi
done < <(grep "repository:" Chart.yaml | awk '{print $2}' | sort | uniq)

# Check dependency aliases
log "Validating dependency aliases..."
if grep -q "alias:" Chart.yaml; then
    info "Found dependency aliases:"
    grep -A 1 "alias:" Chart.yaml | grep "alias:" | awk '{print "  -", $2}'
    log "Dependency aliases configured correctly"
else
    add_warning "No dependency aliases found. Consider using aliases for better organization."
fi

# Check import-values configuration
log "Validating import-values configuration..."
if grep -q "import-values:" Chart.yaml; then
    info "Found import-values configuration:"
    grep -A 5 "import-values:" Chart.yaml | grep -E "(child:|parent:)" | sed 's/^/  /'
    log "Import-values configuration found"
else
    add_warning "No import-values configuration found. Consider importing key values from dependencies."
fi

# Validate dependency conditions and tags
log "Validating dependency conditions and tags..."
DEPS_WITH_CONDITIONS=$(grep -c "condition:" Chart.yaml || true)
DEPS_WITH_TAGS=$(grep -c "tags:" Chart.yaml || true)

info "Dependencies with conditions: ${DEPS_WITH_CONDITIONS}"
info "Dependencies with tags: ${DEPS_WITH_TAGS}"

if [[ "${DEPS_WITH_CONDITIONS}" -eq 0 ]]; then
    add_warning "No dependency conditions found. Consider adding conditions for optional dependencies."
fi

if [[ "${DEPS_WITH_TAGS}" -eq 0 ]]; then
    add_warning "No dependency tags found. Consider adding tags for dependency grouping."
fi

# Check Chart.lock digest
log "Validating Chart.lock digest..."
if grep -q "digest:" Chart.lock; then
    DIGEST=$(grep "digest:" Chart.lock | awk '{print $2}')
    log "Chart.lock digest: ${DIGEST}"
else
    add_warning "No digest found in Chart.lock. This may indicate an incomplete dependency update."
fi

# Summary
echo
echo "=============================================="
echo -e "${BLUE}Dependency Validation Summary${NC}"
echo "=============================================="

if [[ ${VALIDATION_ERRORS} -eq 0 ]] && [[ ${VALIDATION_WARNINGS} -eq 0 ]]; then
    echo -e "${GREEN}✓ All dependency validations passed!${NC}"
    EXIT_CODE=0
elif [[ ${VALIDATION_ERRORS} -eq 0 ]]; then
    echo -e "${YELLOW}⚠ Validation completed with ${VALIDATION_WARNINGS} warning(s)${NC}"
    EXIT_CODE=0
else
    echo -e "${RED}✗ Validation failed with ${VALIDATION_ERRORS} error(s) and ${VALIDATION_WARNINGS} warning(s)${NC}"
    EXIT_CODE=1
fi

echo
echo "Dependency Information:"
echo "- PostgreSQL: ${PG_LOCK_VERSION}"
echo "- Redis: ${REDIS_LOCK_VERSION}"
echo
echo "Chart Information:"
echo "- Chart Version: $(grep '^version:' Chart.yaml | awk '{print $2}')"
echo "- App Version: $(grep '^appVersion:' Chart.yaml | awk '{print $2}')"

if [[ ${EXIT_CODE} -ne 0 ]]; then
    echo
    echo "Recommended actions:"
    echo "1. Run 'helm dependency update' to refresh dependencies"
    echo "2. Check Chart.yaml for version consistency"
    echo "3. Verify repository URLs are accessible"
    echo "4. Address any security warnings"
fi

exit ${EXIT_CODE}
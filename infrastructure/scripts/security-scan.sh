#!/bin/bash
# KubeChat Container Security Scanning Script
# Implements comprehensive security scanning in build process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCAN_RESULTS_DIR="./security-scan-results"
FAIL_ON_HIGH="true"
FAIL_ON_CRITICAL="true"

echo -e "${BLUE}🔒 KubeChat Container Security Scanner${NC}"
echo "========================================"

# Create results directory
mkdir -p "$SCAN_RESULTS_DIR"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install trivy if not present
install_trivy() {
    if ! command_exists trivy; then
        echo -e "${YELLOW}📦 Installing Trivy security scanner...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install trivy
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
        else
            echo -e "${RED}❌ Unsupported OS for automatic Trivy installation${NC}"
            echo "Please install Trivy manually: https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
            exit 1
        fi
    fi
}

# Function to scan image for vulnerabilities
scan_image() {
    local image_name="$1"
    local scan_type="$2"

    echo -e "${BLUE}🔍 Scanning $image_name for $scan_type...${NC}"

    local output_file="$SCAN_RESULTS_DIR/${image_name//\//_}_${scan_type}_scan.json"
    local summary_file="$SCAN_RESULTS_DIR/${image_name//\//_}_${scan_type}_summary.txt"

    # Run trivy scan
    if trivy image --format json --output "$output_file" --scanners "$scan_type" "$image_name"; then
        # Generate human-readable summary
        trivy image --format table --output "$summary_file" --scanners "$scan_type" "$image_name"

        # Extract critical findings
        local critical_count=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL") | .VulnerabilityID' "$output_file" 2>/dev/null | wc -l || echo "0")
        local high_count=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH") | .VulnerabilityID' "$output_file" 2>/dev/null | wc -l || echo "0")
        local medium_count=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "MEDIUM") | .VulnerabilityID' "$output_file" 2>/dev/null | wc -l || echo "0")
        local low_count=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "LOW") | .VulnerabilityID' "$output_file" 2>/dev/null | wc -l || echo "0")

        echo -e "${GREEN}✅ Scan completed for $image_name${NC}"
        echo -e "   Critical: ${RED}$critical_count${NC}"
        echo -e "   High: ${YELLOW}$high_count${NC}"
        echo -e "   Medium: $medium_count"
        echo -e "   Low: $low_count"

        # Check if we should fail the build
        if [[ "$FAIL_ON_CRITICAL" == "true" && "$critical_count" -gt 0 ]]; then
            echo -e "${RED}❌ Build FAILED: Critical vulnerabilities found${NC}"
            return 1
        fi

        if [[ "$FAIL_ON_HIGH" == "true" && "$high_count" -gt 0 ]]; then
            echo -e "${RED}❌ Build FAILED: High severity vulnerabilities found${NC}"
            return 1
        fi

        return 0
    else
        echo -e "${RED}❌ Scan failed for $image_name${NC}"
        return 1
    fi
}

# Function to scan for secrets
scan_secrets() {
    local image_name="$1"

    echo -e "${BLUE}🔐 Scanning $image_name for secrets...${NC}"

    local output_file="$SCAN_RESULTS_DIR/${image_name//\//_}_secrets_scan.json"

    if trivy image --format json --output "$output_file" --scanners secret "$image_name"; then
        local secrets_count=$(jq -r '.Results[]?.Secrets[]? | .RuleID' "$output_file" 2>/dev/null | wc -l || echo "0")

        if [[ "$secrets_count" -gt 0 ]]; then
            echo -e "${RED}❌ Found $secrets_count potential secrets in $image_name${NC}"
            # List the secrets found
            jq -r '.Results[]?.Secrets[]? | "- \(.RuleID): \(.Title)"' "$output_file" 2>/dev/null || true
            return 1
        else
            echo -e "${GREEN}✅ No secrets detected in $image_name${NC}"
            return 0
        fi
    else
        echo -e "${RED}❌ Secret scan failed for $image_name${NC}"
        return 1
    fi
}

# Function to verify image signature (placeholder for future implementation)
verify_signature() {
    local image_name="$1"

    echo -e "${BLUE}🔏 Verifying signature for $image_name...${NC}"

    # For now, this is a placeholder - in production you would use cosign or similar
    # cosign verify --key cosign.pub "$image_name"

    echo -e "${YELLOW}⚠️  Signature verification not implemented (placeholder)${NC}"
    return 0
}

# Function to generate security report
generate_report() {
    local report_file="$SCAN_RESULTS_DIR/security_report.md"

    echo "# KubeChat Container Security Report" > "$report_file"
    echo "Generated on: $(date)" >> "$report_file"
    echo "" >> "$report_file"

    echo "## Scan Results" >> "$report_file"
    echo "" >> "$report_file"

    for result_file in "$SCAN_RESULTS_DIR"/*.txt; do
        if [[ -f "$result_file" ]]; then
            echo "### $(basename "$result_file")" >> "$report_file"
            echo '```' >> "$report_file"
            cat "$result_file" >> "$report_file"
            echo '```' >> "$report_file"
            echo "" >> "$report_file"
        fi
    done

    echo -e "${GREEN}📊 Security report generated: $report_file${NC}"
}

# Main scanning function
main() {
    local images=()
    local scan_vulnerabilities=true
    local scan_secrets_flag=true
    local verify_signatures=false

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --image)
                images+=("$2")
                shift 2
                ;;
            --no-vulnerabilities)
                scan_vulnerabilities=false
                shift
                ;;
            --no-secrets)
                scan_secrets_flag=false
                shift
                ;;
            --verify-signatures)
                verify_signatures=true
                shift
                ;;
            --fail-on-high)
                FAIL_ON_HIGH="$2"
                shift 2
                ;;
            --fail-on-critical)
                FAIL_ON_CRITICAL="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --image IMAGE           Image to scan (can be used multiple times)"
                echo "  --no-vulnerabilities    Skip vulnerability scanning"
                echo "  --no-secrets           Skip secret scanning"
                echo "  --verify-signatures    Verify image signatures"
                echo "  --fail-on-high BOOL    Fail build on high severity (default: true)"
                echo "  --fail-on-critical BOOL Fail build on critical severity (default: true)"
                echo "  --help                 Show this help message"
                exit 0
                ;;
            *)
                echo -e "${RED}❌ Unknown option: $1${NC}"
                exit 1
                ;;
        esac
    done

    # Default images if none specified
    if [[ ${#images[@]} -eq 0 ]]; then
        images=("kubechat/web:dev" "kubechat/api:dev")
    fi

    # Install trivy if needed
    install_trivy

    local overall_exit_code=0

    # Scan each image
    for image in "${images[@]}"; do
        echo -e "${BLUE}🚀 Processing image: $image${NC}"

        # Check if image exists
        if ! docker image inspect "$image" >/dev/null 2>&1; then
            echo -e "${RED}❌ Image $image not found locally${NC}"
            overall_exit_code=1
            continue
        fi

        # Vulnerability scanning
        if [[ "$scan_vulnerabilities" == true ]]; then
            if ! scan_image "$image" "vuln"; then
                overall_exit_code=1
            fi
        fi

        # Secret scanning
        if [[ "$scan_secrets_flag" == true ]]; then
            if ! scan_secrets "$image"; then
                overall_exit_code=1
            fi
        fi

        # Signature verification
        if [[ "$verify_signatures" == true ]]; then
            if ! verify_signature "$image"; then
                overall_exit_code=1
            fi
        fi

        echo ""
    done

    # Generate report
    generate_report

    # Final result
    if [[ $overall_exit_code -eq 0 ]]; then
        echo -e "${GREEN}🎉 All security scans passed!${NC}"
    else
        echo -e "${RED}💥 Security scan failures detected${NC}"
    fi

    exit $overall_exit_code
}

# Run main function with all arguments
main "$@"
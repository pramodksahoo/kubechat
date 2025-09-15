# KubeChat Security Architecture

## Overview

KubeChat is designed with a security-first approach, prioritizing data protection, access control, and compliance with enterprise security standards. This document outlines the comprehensive security architecture, threat model, and defensive measures implemented throughout the platform.

## Security Principles

### 1. Defense in Depth
Multiple layers of security controls protect against various threat vectors:
- Network security (TLS, network policies)
- Application security (input validation, authentication)
- Data security (encryption, access controls)
- Infrastructure security (container security, RBAC)

### 2. Zero Trust Architecture
No implicit trust is granted to any component:
- All communications are authenticated and encrypted
- Every request is validated and authorized
- Principle of least privilege is enforced

### 3. Privacy by Design
Data protection is built into the system architecture:
- Local AI processing to prevent data leakage
- Minimal data collection and retention
- Transparent data handling practices

## Threat Model

### Assets

**Primary Assets:**
- Kubernetes cluster access and credentials
- User authentication credentials and sessions
- Natural language queries and generated commands
- Audit logs and compliance data
- Application source code and secrets

**Secondary Assets:**
- Configuration data
- Cached responses and temporary data
- Log files and debugging information

### Threat Actors

**External Attackers:**
- Opportunistic attackers seeking unauthorized access
- Advanced persistent threats targeting infrastructure
- Automated scanning and exploitation tools

**Insider Threats:**
- Malicious insiders with legitimate access
- Compromised user accounts
- Accidental misuse or misconfiguration

**Supply Chain Threats:**
- Compromised dependencies or container images
- Malicious code injection through CI/CD
- Third-party service provider compromises

### Attack Vectors

**Network-Based Attacks:**
- Man-in-the-middle attacks on API communications
- Network reconnaissance and service enumeration
- Denial of service attacks

**Application-Layer Attacks:**
- Injection attacks (SQL, command, prompt injection)
- Cross-site scripting (XSS) and CSRF attacks
- Authentication and session management vulnerabilities

**Infrastructure Attacks:**
- Container escape and privilege escalation
- Kubernetes RBAC bypass attempts
- Host system compromise

## Security Architecture Components

### 1. Authentication and Authorization

#### Multi-Factor Authentication Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Browser   │    │  Frontend   │    │  Auth API   │    │  Identity   │
│             │    │             │    │             │    │  Provider   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                  │                  │                  │
       │ 1. Login Request │                  │                  │
       ├─────────────────►│                  │                  │
       │                  │ 2. Auth Request  │                  │
       │                  ├─────────────────►│                  │
       │                  │                  │ 3. Verify User   │
       │                  │                  ├─────────────────►│
       │                  │                  │ 4. User Info     │
       │                  │                  │◄─────────────────┤
       │                  │                  │ 5. MFA Challenge │
       │                  │                  ├─────────────────►│
       │                  │ 6. MFA Prompt    │                  │
       │                  │◄─────────────────┤                  │
       │ 7. MFA Response  │                  │                  │
       ├─────────────────►│                  │                  │
       │                  │ 8. MFA Verify    │                  │
       │                  ├─────────────────►│                  │
       │                  │                  │ 9. Validate MFA  │
       │                  │                  ├─────────────────►│
       │                  │                  │ 10. MFA Success  │
       │                  │                  │◄─────────────────┤
       │                  │ 11. JWT Token    │                  │
       │                  │◄─────────────────┤                  │
       │ 12. Auth Success │                  │                  │
       │◄─────────────────┤                  │                  │
```

#### JWT Token Structure

```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "kubechat-key-1"
  },
  "payload": {
    "sub": "user-uuid",
    "iss": "kubechat-auth",
    "aud": "kubechat-api",
    "exp": 1640995200,
    "iat": 1640908800,
    "roles": ["cluster-admin", "audit-viewer"],
    "clusters": ["prod-cluster", "staging-cluster"],
    "session_id": "session-uuid",
    "mfa_verified": true
  }
}
```

#### RBAC Integration

```go
type RBACValidator struct {
    k8sClient kubernetes.Interface
}

func (r *RBACValidator) ValidateAction(ctx context.Context, user *User, action string, resource string) error {
    // Create subject access review
    sar := &authorizationv1.SubjectAccessReview{
        Spec: authorizationv1.SubjectAccessReviewSpec{
            User: user.Username,
            ResourceAttributes: &authorizationv1.ResourceAttributes{
                Verb:     action,
                Resource: resource,
                Group:    "",
            },
        },
    }

    // Check with Kubernetes RBAC
    result, err := r.k8sClient.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("RBAC validation failed: %w", err)
    }

    if !result.Status.Allowed {
        return fmt.Errorf("action %s on %s denied: %s", action, resource, result.Status.Reason)
    }

    return nil
}
```

### 2. Input Validation and Sanitization

#### Query Validation Pipeline

```go
type QueryValidator struct {
    maxLength int
    patterns  []SecurityPattern
}

type SecurityPattern struct {
    Pattern     *regexp.Regexp
    Risk        RiskLevel
    Description string
}

func (v *QueryValidator) ValidateQuery(query string) (*ValidationResult, error) {
    result := &ValidationResult{
        Original:  query,
        Sanitized: query,
        Risks:     []Risk{},
    }

    // Length validation
    if len(query) > v.maxLength {
        return nil, fmt.Errorf("query exceeds maximum length of %d characters", v.maxLength)
    }

    // Pattern-based risk assessment
    for _, pattern := range v.patterns {
        if pattern.Pattern.MatchString(query) {
            result.Risks = append(result.Risks, Risk{
                Level:       pattern.Risk,
                Description: pattern.Description,
                Pattern:     pattern.Pattern.String(),
            })
        }
    }

    // Sanitize input
    result.Sanitized = v.sanitizeInput(query)

    return result, nil
}

func (v *QueryValidator) sanitizeInput(input string) string {
    // Remove null bytes
    input = strings.ReplaceAll(input, "\x00", "")

    // Normalize whitespace
    input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")

    // Remove control characters except newlines and tabs
    input = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(input, "")

    return strings.TrimSpace(input)
}
```

#### Command Safety Classification

```go
type SafetyClassifier struct {
    rules []SafetyRule
}

type SafetyRule struct {
    Pattern     *regexp.Regexp
    Level       SafetyLevel
    Description string
    Examples    []string
}

const (
    SafetyLevelSafe      SafetyLevel = "safe"      // Read operations
    SafetyLevelWarning   SafetyLevel = "warning"   // Write operations
    SafetyLevelDangerous SafetyLevel = "dangerous" // Destructive operations
)

var DefaultSafetyRules = []SafetyRule{
    {
        Pattern:     regexp.MustCompile(`^kubectl\s+get\s+`),
        Level:       SafetyLevelSafe,
        Description: "Read-only operations",
        Examples:    []string{"kubectl get pods", "kubectl get services"},
    },
    {
        Pattern:     regexp.MustCompile(`^kubectl\s+(apply|create|patch)\s+`),
        Level:       SafetyLevelWarning,
        Description: "Resource modification operations",
        Examples:    []string{"kubectl apply -f", "kubectl create deployment"},
    },
    {
        Pattern:     regexp.MustCompile(`^kubectl\s+(delete|destroy)\s+`),
        Level:       SafetyLevelDangerous,
        Description: "Destructive operations",
        Examples:    []string{"kubectl delete namespace", "kubectl delete pod"},
    },
}
```

### 3. Encryption and Data Protection

#### Encryption at Rest

**Database Encryption:**
```sql
-- Enable transparent data encryption
ALTER SYSTEM SET ssl = on;
ALTER SYSTEM SET ssl_cert_file = '/etc/ssl/certs/server.crt';
ALTER SYSTEM SET ssl_key_file = '/etc/ssl/private/server.key';

-- Encrypt sensitive columns
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Store encrypted API keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    key_name VARCHAR(255) NOT NULL,
    encrypted_key BYTEA NOT NULL, -- pgp_sym_encrypt(key, passphrase)
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Kubernetes Secrets Encryption:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubechat-secrets
  namespace: kubechat
type: Opaque
data:
  database-password: <base64-encoded-encrypted-value>
  jwt-signing-key: <base64-encoded-encrypted-value>
  openai-api-key: <base64-encoded-encrypted-value>
```

#### Encryption in Transit

**TLS Configuration:**
```go
func setupTLS() *tls.Config {
    return &tls.Config{
        MinVersion: tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        },
        PreferServerCipherSuites: true,
        CurvePreferences: []tls.CurveID{
            tls.CurveP521,
            tls.CurveP384,
            tls.CurveP256,
        },
    }
}
```

### 4. Audit Trail Security

#### Cryptographic Integrity

```go
type AuditEntry struct {
    ID               int64     `json:"id"`
    UserID           string    `json:"user_id"`
    Query            string    `json:"query"`
    Command          string    `json:"command"`
    Timestamp        time.Time `json:"timestamp"`
    Checksum         string    `json:"checksum"`
    PreviousChecksum string    `json:"previous_checksum"`
}

func (a *AuditEntry) CalculateChecksum(previousChecksum string) string {
    data := fmt.Sprintf("%d|%s|%s|%s|%d|%s",
        a.ID, a.UserID, a.Query, a.Command, a.Timestamp.Unix(), previousChecksum)

    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

func (s *AuditService) ValidateChain(entries []*AuditEntry) error {
    for i, entry := range entries {
        var expectedPrevious string
        if i > 0 {
            expectedPrevious = entries[i-1].Checksum
        }

        expectedChecksum := entry.CalculateChecksum(expectedPrevious)
        if entry.Checksum != expectedChecksum {
            return fmt.Errorf("integrity violation at entry %d: expected %s, got %s",
                entry.ID, expectedChecksum, entry.Checksum)
        }

        if entry.PreviousChecksum != expectedPrevious {
            return fmt.Errorf("chain break at entry %d: expected previous %s, got %s",
                entry.ID, expectedPrevious, entry.PreviousChecksum)
        }
    }
    return nil
}
```

#### Tamper Detection

```go
type IntegrityMonitor struct {
    auditService *AuditService
    alerter      *SecurityAlerter
    checkInterval time.Duration
}

func (i *IntegrityMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(i.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := i.performIntegrityCheck(); err != nil {
                i.alerter.SendAlert("CRITICAL", "Audit trail integrity violation", err)
            }
        }
    }
}

func (i *IntegrityMonitor) performIntegrityCheck() error {
    // Get recent audit entries
    entries, err := i.auditService.GetRecentEntries(time.Hour * 24)
    if err != nil {
        return fmt.Errorf("failed to retrieve audit entries: %w", err)
    }

    // Validate cryptographic chain
    return i.auditService.ValidateChain(entries)
}
```

### 5. Container and Infrastructure Security

#### Secure Container Images

**Multi-stage Dockerfile with Security:**
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Production stage - minimal and secure
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app/server /server

# Run as non-root user
USER 65534:65534

EXPOSE 8080
ENTRYPOINT ["/server"]
```

#### Kubernetes Security Policies

**Pod Security Standards:**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: kubechat
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

**Network Policies:**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kubechat-network-policy
  namespace: kubechat
spec:
  podSelector:
    matchLabels:
      app: kubechat
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443  # HTTPS to Kubernetes API
  - to: []  # Allow DNS
    ports:
    - protocol: UDP
      port: 53
```

### 6. AI Security

#### Prompt Injection Protection

```go
type PromptSanitizer struct {
    blacklistedPatterns []string
    maxTokens          int
}

func (p *PromptSanitizer) SanitizePrompt(userInput string) (string, error) {
    // Check for prompt injection patterns
    for _, pattern := range p.blacklistedPatterns {
        if matched, _ := regexp.MatchString(pattern, userInput); matched {
            return "", fmt.Errorf("potential prompt injection detected")
        }
    }

    // Limit token count to prevent resource exhaustion
    tokens := strings.Fields(userInput)
    if len(tokens) > p.maxTokens {
        return "", fmt.Errorf("input exceeds maximum token limit")
    }

    // Escape special characters
    sanitized := html.EscapeString(userInput)

    return sanitized, nil
}

var DefaultBlacklistedPatterns = []string{
    `(?i)ignore\s+previous\s+instructions`,
    `(?i)disregard\s+the\s+above`,
    `(?i)act\s+as\s+a\s+different`,
    `(?i)pretend\s+to\s+be`,
    `(?i)system\s*:\s*`,
    `(?i)assistant\s*:\s*`,
}
```

#### AI Response Validation

```go
type ResponseValidator struct {
    commandPattern *regexp.Regexp
    safetyChecker  *SafetyClassifier
}

func (r *ResponseValidator) ValidateResponse(response *AIResponse) error {
    // Ensure response contains valid kubectl command
    if !r.commandPattern.MatchString(response.Command) {
        return fmt.Errorf("invalid command format in AI response")
    }

    // Check command safety
    safety := r.safetyChecker.ClassifyCommand(response.Command)
    if safety == SafetyLevelDangerous && !response.RequiresApproval {
        return fmt.Errorf("dangerous command must require approval")
    }

    // Validate explanation exists for non-safe commands
    if safety != SafetyLevelSafe && len(response.Explanation) < 50 {
        return fmt.Errorf("insufficient explanation for non-safe command")
    }

    return nil
}
```

## Security Monitoring and Alerting

### Real-time Security Monitoring

```go
type SecurityMonitor struct {
    alerter        *AlertManager
    rateLimit     *RateLimiter
    anomalyDetector *AnomalyDetector
}

func (s *SecurityMonitor) MonitorQuery(ctx context.Context, query *QueryRequest) {
    // Rate limiting check
    if s.rateLimit.Exceeded(query.UserID) {
        s.alerter.SendAlert("WARNING", "Rate limit exceeded", map[string]interface{}{
            "user_id": query.UserID,
            "ip":      query.IPAddress,
        })
    }

    // Anomaly detection
    if s.anomalyDetector.IsAnomalous(query) {
        s.alerter.SendAlert("INFO", "Anomalous query pattern detected", map[string]interface{}{
            "user_id": query.UserID,
            "query":   query.Query,
            "risk":    s.anomalyDetector.CalculateRisk(query),
        })
    }
}
```

### Security Metrics Collection

```go
var (
    securityEventsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "kubechat_security_events_total",
            Help: "Total number of security events detected",
        },
        []string{"type", "severity", "user"},
    )

    failedAuthAttemptsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "kubechat_failed_auth_attempts_total",
            Help: "Total number of failed authentication attempts",
        },
        []string{"source_ip", "username"},
    )

    dangerousCommandsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "kubechat_dangerous_commands_total",
            Help: "Total number of dangerous commands executed",
        },
        []string{"user", "approved"},
    )
)
```

## Compliance and Governance

### SOC 2 Type II Compliance

**Control Implementation:**
- **Security:** Multi-factor authentication, encryption, access controls
- **Availability:** High availability deployment, monitoring, backup procedures
- **Confidentiality:** Data encryption, access logging, secure communication
- **Processing Integrity:** Input validation, audit trails, error handling

### GDPR Compliance

**Data Protection Measures:**
```go
type DataProtectionService struct {
    encryptionKey []byte
    retentionPolicy *RetentionPolicy
}

func (d *DataProtectionService) HandleDataSubjectRequest(request *DataSubjectRequest) error {
    switch request.Type {
    case "access":
        return d.exportUserData(request.UserID)
    case "deletion":
        return d.deleteUserData(request.UserID)
    case "rectification":
        return d.correctUserData(request.UserID, request.Corrections)
    case "portability":
        return d.exportPortableData(request.UserID)
    default:
        return fmt.Errorf("unsupported request type: %s", request.Type)
    }
}
```

### Audit Compliance (SOX, HIPAA)

**Compliance Reporting:**
```go
type ComplianceReporter struct {
    auditService *AuditService
    templates    map[string]*ReportTemplate
}

func (c *ComplianceReporter) GenerateSOXReport(period TimePeriod) (*ComplianceReport, error) {
    entries, err := c.auditService.GetEntriesForPeriod(period.Start, period.End)
    if err != nil {
        return nil, err
    }

    report := &ComplianceReport{
        Period:      period,
        TotalEntries: len(entries),
        UserActions: c.aggregateUserActions(entries),
        RiskEvents:  c.identifyRiskEvents(entries),
        IntegrityCheck: c.validateIntegrity(entries),
    }

    return report, nil
}
```

## Security Testing and Validation

### Automated Security Testing

**Security Test Suite:**
```go
func TestSecurityControls(t *testing.T) {
    tests := []struct {
        name     string
        testFunc func(t *testing.T)
    }{
        {"SQL Injection Protection", testSQLInjection},
        {"XSS Protection", testXSSProtection},
        {"CSRF Protection", testCSRFProtection},
        {"Authentication Bypass", testAuthBypass},
        {"Privilege Escalation", testPrivilegeEscalation},
        {"Data Exposure", testDataExposure},
        {"Rate Limiting", testRateLimiting},
        {"Input Validation", testInputValidation},
    }

    for _, tt := range tests {
        t.Run(tt.name, tt.testFunc)
    }
}
```

### Penetration Testing Guidelines

**Regular Security Assessments:**
1. **Quarterly automated scans** using tools like OWASP ZAP, Nessus
2. **Annual penetration testing** by third-party security firms
3. **Continuous security monitoring** with SIEM integration
4. **Bug bounty program** for community security research

## Incident Response

### Security Incident Response Plan

```go
type IncidentResponse struct {
    alerter     *AlertManager
    forensics   *ForensicsCollector
    containment *ContainmentManager
}

func (i *IncidentResponse) HandleSecurityIncident(incident *SecurityIncident) error {
    // 1. Immediate containment
    if err := i.containment.ContainThreat(incident); err != nil {
        return fmt.Errorf("containment failed: %w", err)
    }

    // 2. Evidence collection
    evidence, err := i.forensics.CollectEvidence(incident)
    if err != nil {
        return fmt.Errorf("evidence collection failed: %w", err)
    }

    // 3. Alert stakeholders
    i.alerter.SendIncidentAlert(incident, evidence)

    // 4. Begin recovery procedures
    return i.initiateRecovery(incident)
}
```

## Conclusion

KubeChat's security architecture implements comprehensive defense-in-depth strategies to protect against modern threats while maintaining usability and performance. The combination of cryptographic integrity, comprehensive monitoring, and compliance-ready audit trails ensures that KubeChat meets enterprise security requirements while providing a seamless user experience.

Key security features:
- **Multi-layered authentication** with MFA and RBAC integration
- **Cryptographic audit trail integrity** with tamper detection
- **Comprehensive input validation** and prompt injection protection
- **End-to-end encryption** for data at rest and in transit
- **Real-time security monitoring** with anomaly detection
- **Compliance-ready audit trails** for SOX, HIPAA, and SOC 2
- **Container security** with minimal attack surface
- **AI safety controls** with command classification and approval workflows

This security architecture enables organizations to deploy KubeChat with confidence in enterprise environments while maintaining the flexibility needed for effective Kubernetes management.
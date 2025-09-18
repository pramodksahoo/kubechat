// Admin credential management types for KubeChat
// AC 8: Built-in Admin Credential Management - K8s Secret Integration
// Enterprise security compliance for SOC 2 Type II, GDPR

import type { CredentialRotationResult, ComplianceValidationResult } from './admin';

export interface K8sSecretCredentials {
  admin_username: string;
  admin_password: string;
  admin_email: string;
  password_created: string;
  password_expires: string;
  last_rotation: string;
  rotation_count: string;
  min_password_length: string;
  max_password_age_days: string;
  password_complexity: string;
  sync_enabled: string;
  last_db_sync: string;
  sync_status: string;
}

export interface AdminCredentialSyncRecord {
  id: string;
  sync_timestamp: string;
  sync_source: 'k8s_secret' | 'database' | 'rotation' | 'emergency';
  sync_type: 'bootstrap' | 'update' | 'rotation' | 'verification';
  sync_status: 'success' | 'failed' | 'pending' | 'partial';
  username: string;
  password_hash: string;
  email: string;
  k8s_secret_name: string;
  k8s_namespace: string;
  k8s_resource_version?: string;
  password_created_at: string;
  password_expires_at: string;
  rotation_count: number;
  compliance_status: 'compliant' | 'non_compliant' | 'pending_review';
  compliance_frameworks: string[];
  sync_initiated_by: string;
  error_message?: string;
  notes?: string;
}

export interface AdminCredentialAuditEntry {
  id: string;
  audit_timestamp: string;
  event_type: 'credential_access' | 'password_change' | 'sync_operation' | 'rotation' | 'emergency_access';
  actor_type: 'user' | 'system' | 'k8s_controller';
  actor_id: string;
  operation: string;
  target_username: string;
  source_system?: 'k8s_secret' | 'database' | 'admin_ui' | 'cli';
  destination_system?: string;
  ip_address?: string;
  user_agent?: string;
  session_id?: string;
  risk_level: 'low' | 'medium' | 'high' | 'critical';
  compliance_impact: string[];
  success: boolean;
  error_code?: string;
  error_message?: string;
  audit_checksum: string;
  previous_audit_checksum?: string;
  retention_policy: string;
  legal_hold_exempt: boolean;
  pii_data_mask: boolean;
}

export interface EmergencyAccessRecord {
  id: string;
  emergency_timestamp: string;
  emergency_type: 'password_reset' | 'account_unlock' | 'credential_recovery' | 'bypass_mfa';
  requested_by: string;
  approved_by?: string;
  justification: string;
  business_impact: string;
  approval_status: 'pending' | 'approved' | 'denied' | 'expired';
  approval_timestamp?: string;
  emergency_username?: string;
  emergency_password_hash?: string;
  emergency_token?: string;
  expires_at?: string;
  ip_whitelist?: string[];
  session_restrictions?: Record<string, any>;
  mfa_bypass_approved: boolean;
  used_at?: string;
  revoked_at?: string;
  revocation_reason?: string;
  compliance_review_required: boolean;
  compliance_review_completed: boolean;
  audit_checksum: string;
}

export interface CredentialSyncConfig {
  enabled: boolean;
  sync_interval_minutes: number;
  auto_rotation_enabled: boolean;
  rotation_warning_days: number;
  max_sync_retries: number;
  sync_timeout_seconds: number;
  compliance_validation_enabled: boolean;
  audit_retention_days: number;
  emergency_access_enabled: boolean;
  mfa_required_for_emergency: boolean;
}

export interface PasswordPolicy {
  min_length: number;
  max_age_days: number;
  complexity_requirements: {
    uppercase: boolean;
    lowercase: boolean;
    numbers: boolean;
    special_chars: boolean;
    no_common_passwords: boolean;
    no_user_info: boolean;
  };
  history_length: number;
  lockout_after_attempts: number;
  lockout_duration_minutes: number;
}

export interface ComplianceReport {
  report_id: string;
  generated_at: string;
  generated_by: string;
  report_period: {
    start: string;
    end: string;
  };
  compliance_frameworks: string[];
  overall_status: 'compliant' | 'non_compliant' | 'partial_compliance';
  findings: {
    check_id: string;
    check_name: string;
    status: 'pass' | 'fail' | 'warning' | 'not_applicable';
    severity: 'low' | 'medium' | 'high' | 'critical';
    description: string;
    evidence: string[];
    remediation_steps: string[];
    compliance_frameworks: string[];
  }[];
  statistics: {
    total_checks: number;
    passed: number;
    failed: number;
    warnings: number;
    not_applicable: number;
  };
  recommendations: string[];
  next_review_date: string;
}

// Service interfaces for admin credential management
export interface AdminCredentialSyncService {
  // K8s Secret operations
  getK8sSecretCredentials(): Promise<K8sSecretCredentials>;
  updateK8sSecret(credentials: Partial<K8sSecretCredentials>): Promise<void>;
  validateK8sSecretExists(): Promise<boolean>;

  // Database sync operations
  syncFromK8sSecret(): Promise<AdminCredentialSyncRecord>;
  syncToK8sSecret(): Promise<AdminCredentialSyncRecord>;
  getSyncHistory(limit?: number): Promise<AdminCredentialSyncRecord[]>;
  getLatestSyncStatus(): Promise<AdminCredentialSyncRecord | null>;

  // Credential rotation
  rotateCredentials(newPassword?: string): Promise<CredentialRotationResult>;
  scheduleRotation(rotationDate: string): Promise<void>;
  cancelScheduledRotation(): Promise<void>;

  // Compliance and audit
  validateCompliance(): Promise<ComplianceValidationResult>;
  generateComplianceReport(frameworks: string[], period: { start: string; end: string }): Promise<ComplianceReport>;
  getAuditTrail(filters?: {
    event_type?: string;
    actor_id?: string;
    date_range?: { start: string; end: string };
  }): Promise<AdminCredentialAuditEntry[]>;

  // Emergency access
  requestEmergencyAccess(request: {
    emergency_type: string;
    justification: string;
    business_impact: string;
  }): Promise<EmergencyAccessRecord>;
  approveEmergencyAccess(requestId: string, approval: {
    approved_by: string;
    mfa_bypass_approved: boolean;
    ip_whitelist?: string[];
    session_restrictions?: Record<string, any>;
  }): Promise<EmergencyAccessRecord>;
  revokeEmergencyAccess(requestId: string, revocation_reason: string): Promise<void>;

  // Configuration
  getConfig(): Promise<CredentialSyncConfig>;
  updateConfig(config: Partial<CredentialSyncConfig>): Promise<CredentialSyncConfig>;
  getPasswordPolicy(): Promise<PasswordPolicy>;
  updatePasswordPolicy(policy: Partial<PasswordPolicy>): Promise<PasswordPolicy>;

  // Health monitoring
  checkSyncHealth(): Promise<{
    status: 'healthy' | 'degraded' | 'unhealthy';
    last_successful_sync: string;
    next_scheduled_sync: string;
    pending_rotations: number;
    compliance_violations: number;
    details: Record<string, any>;
  }>;
}

// UI Component Props for admin credential management
export interface AdminCredentialSettingsProps {
  syncStatus: AdminCredentialSyncRecord | null;
  complianceStatus: ComplianceValidationResult | null;
  onSync: () => Promise<void>;
  onRotate: (newPassword?: string) => Promise<void>;
  onValidateCompliance: () => Promise<void>;
  isLoading: boolean;
  error: string | null;
}

export interface CredentialAuditLogProps {
  auditEntries: AdminCredentialAuditEntry[];
  onLoadMore: () => Promise<void>;
  onExport: (format: 'csv' | 'json' | 'pdf') => Promise<void>;
  filters: {
    event_type?: string;
    actor_id?: string;
    date_range?: { start: string; end: string };
  };
  onFiltersChange: (filters: any) => void;
  isLoading: boolean;
}

export interface EmergencyAccessPanelProps {
  emergencyRequests: EmergencyAccessRecord[];
  onRequest: (request: { emergency_type: string; justification: string; business_impact: string }) => Promise<void>;
  onApprove: (requestId: string, approval: any) => Promise<void>;
  onRevoke: (requestId: string, reason: string) => Promise<void>;
  userRole: string;
  isLoading: boolean;
}
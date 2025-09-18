// Admin Credential Sync Service
// Following coding standards from docs/architecture/coding-standards.md
// AC 8: Built-in Admin Credential Management - K8s Secret Integration

import { httpClient } from '../api';
import type {
  K8sSecretCredentials,
  AdminCredentialSyncRecord,
  AdminCredentialAuditEntry,
  EmergencyAccessRecord,
  CredentialSyncConfig,
  PasswordPolicy,
  ComplianceReport,
  AdminCredentialSyncService as IAdminCredentialSyncService
} from '../../types/adminCredentials';
import type { CredentialRotationResult, ComplianceValidationResult } from '../../types/admin';

class AdminCredentialSyncService implements IAdminCredentialSyncService {

  // K8s Secret operations
  async getK8sSecretCredentials(): Promise<K8sSecretCredentials> {
    try {
      const response = await httpClient.get('/api/v1/admin/credentials/k8s-secret');

      if (!response.data) {
        throw new Error('Failed to retrieve K8s Secret credentials');
      }

      return response.data as K8sSecretCredentials;
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to retrieve K8s Secret credentials');
    }
  }

  async updateK8sSecret(credentials: Partial<K8sSecretCredentials>): Promise<void> {
    try {
      await httpClient.put('/api/v1/admin/credentials/k8s-secret', credentials);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to update K8s Secret');
    }
  }

  async validateK8sSecretExists(): Promise<boolean> {
    try {
      const response = await httpClient.get('/api/v1/admin/credentials/k8s-secret/validate');
      return (response.data as any)?.exists || false;
    } catch (error: any) {
      console.error('K8s Secret validation failed:', error);
      return false;
    }
  }

  // Database sync operations
  async syncFromK8sSecret(): Promise<AdminCredentialSyncRecord> {
    try {
      const response = await httpClient.post('/api/v1/admin/credentials/sync/from-k8s');

      if (!response.data) {
        throw new Error('Failed to sync from K8s Secret');
      }

      return this.transformSyncRecord(response.data);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to sync credentials from K8s Secret');
    }
  }

  async syncToK8sSecret(): Promise<AdminCredentialSyncRecord> {
    try {
      const response = await httpClient.post('/api/v1/admin/credentials/sync/to-k8s');

      if (!response.data) {
        throw new Error('Failed to sync to K8s Secret');
      }

      return this.transformSyncRecord(response.data);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to sync credentials to K8s Secret');
    }
  }

  async getSyncHistory(limit = 50): Promise<AdminCredentialSyncRecord[]> {
    try {
      const response = await httpClient.get(`/api/v1/admin/credentials/sync/history?limit=${limit}`);

      return ((response.data as any))?.sync_records?.map(this.transformSyncRecord) || [];
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to retrieve sync history');
    }
  }

  async getLatestSyncStatus(): Promise<AdminCredentialSyncRecord | null> {
    try {
      const response = await httpClient.get('/api/v1/admin/credentials/sync/status');

      if (!response.data) {
        return null;
      }

      return this.transformSyncRecord(response.data);
    } catch (error: any) {
      console.error('Failed to get sync status:', error);
      return null;
    }
  }

  // Credential rotation
  async rotateCredentials(newPassword?: string): Promise<CredentialRotationResult> {
    try {
      const response = await httpClient.post('/api/v1/admin/credentials/rotate', {
        new_password: newPassword
      });

      if (!response.data) {
        throw new Error('Failed to rotate credentials');
      }

      const data = response.data as any;
      return {
        success: data.success,
        rotationId: data.rotation_id,
        newPasswordExpiry: data.new_password_expiry,
        rotationCount: data.rotation_count,
        message: data.message,
        auditTrailId: data.audit_trail_id
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to rotate admin credentials');
    }
  }

  async scheduleRotation(rotationDate: string): Promise<void> {
    try {
      await httpClient.post('/api/v1/admin/credentials/schedule-rotation', {
        rotation_date: rotationDate
      });
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to schedule credential rotation');
    }
  }

  async cancelScheduledRotation(): Promise<void> {
    try {
      await httpClient.delete('/api/v1/admin/credentials/scheduled-rotation');
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to cancel scheduled rotation');
    }
  }

  // Compliance and audit
  async validateCompliance(): Promise<ComplianceValidationResult> {
    try {
      const response = await httpClient.get('/api/v1/admin/credentials/compliance/validate');

      if (!response.data) {
        throw new Error('Failed to validate compliance');
      }

      const data = response.data as any;
      return {
        overall: data.overall_status,
        checks: data.checks?.map((check: any) => ({
          checkName: check.check_name,
          status: check.status,
          details: check.details,
          remediation: check.remediation
        })) || [],
        lastValidated: data.last_validated,
        nextValidationDue: data.next_validation_due
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to validate credential compliance');
    }
  }

  async generateComplianceReport(
    frameworks: string[],
    period: { start: string; end: string }
  ): Promise<ComplianceReport> {
    try {
      const response = await httpClient.post('/api/v1/admin/credentials/compliance/report', {
        frameworks,
        period
      });

      if (!response.data) {
        throw new Error('Failed to generate compliance report');
      }

      const data = response.data as any;
      return data;
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to generate compliance report');
    }
  }

  async getAuditTrail(filters?: {
    event_type?: string;
    actor_id?: string;
    date_range?: { start: string; end: string };
  }): Promise<AdminCredentialAuditEntry[]> {
    try {
      const params = new URLSearchParams();
      if (filters?.event_type) params.append('event_type', filters.event_type);
      if (filters?.actor_id) params.append('actor_id', filters.actor_id);
      if (filters?.date_range?.start) params.append('start_date', filters.date_range.start);
      if (filters?.date_range?.end) params.append('end_date', filters.date_range.end);

      const response = await httpClient.get(`/api/v1/admin/credentials/audit?${params.toString()}`);

      return ((response.data as any))?.audit_entries?.map(this.transformAuditEntry) || [];
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to retrieve audit trail');
    }
  }

  // Emergency access
  async requestEmergencyAccess(request: {
    emergency_type: string;
    justification: string;
    business_impact: string;
  }): Promise<EmergencyAccessRecord> {
    try {
      const response = await httpClient.post('/api/v1/admin/credentials/emergency-access/request', request);

      if (!response.data) {
        throw new Error('Failed to request emergency access');
      }

      return this.transformEmergencyRecord(response.data);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to request emergency access');
    }
  }

  async approveEmergencyAccess(
    requestId: string,
    approval: {
      approved_by: string;
      mfa_bypass_approved: boolean;
      ip_whitelist?: string[];
      session_restrictions?: Record<string, any>;
    }
  ): Promise<EmergencyAccessRecord> {
    try {
      const response = await httpClient.post(
        `/api/v1/admin/credentials/emergency-access/${requestId}/approve`,
        approval
      );

      if (!response.data) {
        throw new Error('Failed to approve emergency access');
      }

      return this.transformEmergencyRecord(response.data);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to approve emergency access');
    }
  }

  async revokeEmergencyAccess(requestId: string, revocation_reason: string): Promise<void> {
    try {
      await httpClient.post(`/api/v1/admin/credentials/emergency-access/${requestId}/revoke`, {
        revocation_reason
      });
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to revoke emergency access');
    }
  }

  // Configuration
  async getConfig(): Promise<CredentialSyncConfig> {
    try {
      const response = await httpClient.get('/api/v1/admin/credentials/config');

      const data = response.data as any;
      return data || this.getDefaultConfig();
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to retrieve credential sync configuration');
    }
  }

  async updateConfig(config: Partial<CredentialSyncConfig>): Promise<CredentialSyncConfig> {
    try {
      const response = await httpClient.put('/api/v1/admin/credentials/config', config);

      if (!response.data) {
        throw new Error('Failed to update configuration');
      }

      const data = response.data as any;
      return data;
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to update credential sync configuration');
    }
  }

  async getPasswordPolicy(): Promise<PasswordPolicy> {
    try {
      const response = await httpClient.get('/api/v1/admin/credentials/password-policy');

      const data = response.data as any;
      return data || this.getDefaultPasswordPolicy();
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to retrieve password policy');
    }
  }

  async updatePasswordPolicy(policy: Partial<PasswordPolicy>): Promise<PasswordPolicy> {
    try {
      const response = await httpClient.put('/api/v1/admin/credentials/password-policy', policy);

      if (!response.data) {
        throw new Error('Failed to update password policy');
      }

      const data = response.data as any;
      return data;
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to update password policy');
    }
  }

  // Health monitoring
  async checkSyncHealth(): Promise<{
    status: 'healthy' | 'degraded' | 'unhealthy';
    last_successful_sync: string;
    next_scheduled_sync: string;
    pending_rotations: number;
    compliance_violations: number;
    details: Record<string, any>;
  }> {
    try {
      const response = await httpClient.get('/api/v1/admin/credentials/health');

      const data = response.data as any;
      return data || {
        status: 'unhealthy',
        last_successful_sync: '',
        next_scheduled_sync: '',
        pending_rotations: 0,
        compliance_violations: 0,
        details: {}
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to check sync health');
    }
  }

  // Private helper methods
  private transformSyncRecord(data: any): AdminCredentialSyncRecord {
    return {
      id: data.id,
      sync_timestamp: data.sync_timestamp,
      sync_source: data.sync_source,
      sync_type: data.sync_type,
      sync_status: data.sync_status,
      username: data.username,
      password_hash: data.password_hash,
      email: data.email,
      k8s_secret_name: data.k8s_secret_name,
      k8s_namespace: data.k8s_namespace,
      k8s_resource_version: data.k8s_resource_version,
      password_created_at: data.password_created_at,
      password_expires_at: data.password_expires_at,
      rotation_count: data.rotation_count,
      compliance_status: data.compliance_status,
      compliance_frameworks: data.compliance_frameworks || [],
      sync_initiated_by: data.sync_initiated_by,
      error_message: data.error_message,
      notes: data.notes
    };
  }

  private transformAuditEntry(data: any): AdminCredentialAuditEntry {
    return {
      id: data.id,
      audit_timestamp: data.audit_timestamp,
      event_type: data.event_type,
      actor_type: data.actor_type,
      actor_id: data.actor_id,
      operation: data.operation,
      target_username: data.target_username,
      source_system: data.source_system,
      destination_system: data.destination_system,
      ip_address: data.ip_address,
      user_agent: data.user_agent,
      session_id: data.session_id,
      risk_level: data.risk_level,
      compliance_impact: data.compliance_impact || [],
      success: data.success,
      error_code: data.error_code,
      error_message: data.error_message,
      audit_checksum: data.audit_checksum,
      previous_audit_checksum: data.previous_audit_checksum,
      retention_policy: data.retention_policy,
      legal_hold_exempt: data.legal_hold_exempt,
      pii_data_mask: data.pii_data_mask
    };
  }

  private transformEmergencyRecord(data: any): EmergencyAccessRecord {
    return {
      id: data.id,
      emergency_timestamp: data.emergency_timestamp,
      emergency_type: data.emergency_type,
      requested_by: data.requested_by,
      approved_by: data.approved_by,
      justification: data.justification,
      business_impact: data.business_impact,
      approval_status: data.approval_status,
      approval_timestamp: data.approval_timestamp,
      emergency_username: data.emergency_username,
      emergency_password_hash: data.emergency_password_hash,
      emergency_token: data.emergency_token,
      expires_at: data.expires_at,
      ip_whitelist: data.ip_whitelist || [],
      session_restrictions: data.session_restrictions,
      mfa_bypass_approved: data.mfa_bypass_approved,
      used_at: data.used_at,
      revoked_at: data.revoked_at,
      revocation_reason: data.revocation_reason,
      compliance_review_required: data.compliance_review_required,
      compliance_review_completed: data.compliance_review_completed,
      audit_checksum: data.audit_checksum
    };
  }

  private getDefaultConfig(): CredentialSyncConfig {
    return {
      enabled: true,
      sync_interval_minutes: 60,
      auto_rotation_enabled: true,
      rotation_warning_days: 7,
      max_sync_retries: 3,
      sync_timeout_seconds: 30,
      compliance_validation_enabled: true,
      audit_retention_days: 2555, // 7 years for SOX compliance
      emergency_access_enabled: true,
      mfa_required_for_emergency: true
    };
  }

  private getDefaultPasswordPolicy(): PasswordPolicy {
    return {
      min_length: 24,
      max_age_days: 90,
      complexity_requirements: {
        uppercase: true,
        lowercase: true,
        numbers: true,
        special_chars: true,
        no_common_passwords: true,
        no_user_info: true
      },
      history_length: 12,
      lockout_after_attempts: 5,
      lockout_duration_minutes: 30
    };
  }

  private handleApiError(error: any, defaultMessage: string): Error {
    if (error.response?.data?.message) {
      return new Error(error.response.data.message);
    }

    if (error.response?.status === 401) {
      return new Error('Unauthorized access - admin privileges required');
    }

    if (error.response?.status === 403) {
      return new Error('Forbidden - insufficient permissions for credential management');
    }

    if (error.response?.status === 404) {
      return new Error('Credential management endpoint not found');
    }

    if (error.response?.status >= 500) {
      return new Error('Server error in credential management - please try again later');
    }

    return new Error(error.message || defaultMessage);
  }
}

export const adminCredentialSyncService = new AdminCredentialSyncService();
// Admin interface types for KubeChat
// Following coding standards from docs/architecture/coding-standards.md
// AC 1-7: Admin User Management & RBAC Interface Types

import type { EmergencyAccessRecord } from './adminCredentials';

export interface AdminUser {
  id: string;
  username: string;
  email: string;
  role: string;
  permissions?: string[];
  clusters?: string[];
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
  isActive: boolean;
  mfaEnabled?: boolean;
  accountLocked?: boolean;
  failedLoginAttempts?: number;
  lastPasswordChange?: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role: string;
  permissions?: string[];
  clusters?: string[];
  requirePasswordChange?: boolean;
  mfaRequired?: boolean;
}

export interface UpdateUserRequest {
  email?: string;
  role?: string;
  permissions?: string[];
  clusters?: string[];
  isActive?: boolean;
  mfaEnabled?: boolean;
  accountLocked?: boolean;
  resetPassword?: boolean;
}

export interface Role {
  id: string;
  name: string;
  description: string;
  permissions: string[];
  isSystemRole: boolean;
  createdAt: string;
  updatedAt: string;
  userCount?: number;
}

export interface Permission {
  id: string;
  name: string;
  description: string;
  resource: string;
  action: string;
  scope?: string;
  category: 'system' | 'cluster' | 'namespace' | 'user';
}

export interface UserSession {
  id: string;
  userId: string;
  username: string;
  sessionToken: string;
  ipAddress: string;
  userAgent: string;
  createdAt: string;
  expiresAt: string;
  lastActivity: string;
  isActive: boolean;
  location?: {
    country?: string;
    city?: string;
    timezone?: string;
  };
}

export interface AuditLogEntry {
  id: string;
  userId?: string;
  username?: string;
  sessionId?: string;
  timestamp: string;
  action: string;
  resource: string;
  resourceId?: string;
  ipAddress: string;
  userAgent: string;
  success: boolean;
  errorMessage?: string;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  complianceFrameworks?: string[];
  metadata?: Record<string, any>;
}

export interface SecurityPolicy {
  id: string;
  name: string;
  description: string;
  type: 'password' | 'session' | 'access' | 'mfa';
  enabled: boolean;
  configuration: Record<string, any>;
  createdAt: string;
  updatedAt: string;
  lastModifiedBy: string;
}

// Admin credential management types (AC 8)
export interface AdminCredentialSyncStatus {
  syncStatus: 'synced' | 'out_of_sync' | 'rotation_pending' | 'error' | 'pending_initial_sync';
  lastSync: string;
  passwordExpiry: string;
  rotationCount: number;
  complianceStatus: 'compliant' | 'non_compliant' | 'pending_review';
  k8sSecretVersion?: string;
  nextRotationDue?: string;
  syncSource?: 'k8s_secret' | 'database' | 'rotation';
}

export interface CredentialRotationResult {
  success: boolean;
  rotationId: string;
  newPasswordExpiry: string;
  rotationCount: number;
  message: string;
  auditTrailId?: string;
}

export interface ComplianceValidationResult {
  overall: 'compliant' | 'non_compliant' | 'pending_review';
  checks: {
    checkName: string;
    status: 'passed' | 'failed' | 'warning';
    details: string;
    remediation?: string;
  }[];
  lastValidated: string;
  nextValidationDue: string;
}

export interface EmergencyAccessRequest {
  id: string;
  requestedBy: string;
  justification: string;
  businessImpact: string;
  approvalStatus: 'pending' | 'approved' | 'denied' | 'expired';
  approvedBy?: string;
  approvalTimestamp?: string;
  emergencyToken?: string;
  expiresAt?: string;
  usedAt?: string;
  revokedAt?: string;
  revocationReason?: string;
  auditTrail: string[];
}

// Admin state management types
export interface AdminState {
  // User management
  users: AdminUser[];
  selectedUser: AdminUser | null;
  userSearchQuery: string;
  userFilters: {
    role?: string;
    status?: 'active' | 'inactive' | 'locked';
    cluster?: string;
  };

  // Role management
  roles: Role[];
  permissions: Permission[];
  selectedRole: Role | null;

  // Session management
  activeSessions: UserSession[];
  sessions: UserSession[];
  sessionFilters: {
    userId?: string;
    ipAddress?: string;
    location?: string;
  };

  // Audit logs
  auditLogs: AdminAuditResponse;
  auditFilters: {
    userId?: string;
    action?: string;
    dateRange?: {
      start: string;
      end: string;
    };
    riskLevel?: string;
  };

  // Security policies
  securityPolicies: SecurityPolicy[];

  // Admin credentials (AC 8)
  adminCredentials: AdminCredentialSyncStatus;

  // UI state
  isLoading: boolean;
  error: string | null;
  currentView: 'dashboard' | 'users' | 'roles' | 'sessions' | 'security' | 'audit' | 'credentials';

  // Actions
  loadUsers: () => Promise<void>;
  createUser: (userData: CreateUserRequest) => Promise<AdminUser>;
  updateUser: (userId: string, updates: UpdateUserRequest) => Promise<AdminUser>;
  deleteUser: (userId: string) => Promise<void>;
  loadRoles: () => Promise<void>;
  loadPermissions: () => Promise<void>;
  createRole: (roleData: Omit<Role, 'id' | 'createdAt' | 'updatedAt'>) => Promise<Role>;
  updateRole: (roleId: string, updates: Partial<Role>) => Promise<Role>;
  deleteRole: (roleId: string) => Promise<void>;
  loadActiveSessions: () => Promise<void>;
  loadSessions: () => Promise<void>;
  terminateSession: (sessionId: string) => Promise<void>;
  loadAuditLogs: () => Promise<void>;
  loadSecurityPolicies: () => Promise<void>;
  updateSecurityPolicy: (policyId: string, updates: Partial<SecurityPolicy>) => Promise<SecurityPolicy>;

  // Admin credential management actions (AC 8)
  syncAdminCredentials: () => Promise<void>;
  rotateAdminPassword: (newPassword?: string) => Promise<CredentialRotationResult>;
  validateCredentialCompliance: () => Promise<ComplianceValidationResult>;
  requestEmergencyAccess: (justification: string, businessImpact: string) => Promise<EmergencyAccessRecord>;

  // UI actions
  setCurrentView: (view: AdminState['currentView']) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

// API Response types
export interface AdminUsersResponse {
  users: AdminUser[];
  total: number;
  page: number;
  limit: number;
}

export interface AdminRolesResponse {
  roles: Role[];
  total: number;
}

export interface AdminSessionsResponse {
  sessions: UserSession[];
  total: number;
  active: number;
}

export interface AdminAuditResponse {
  logs: AuditLogEntry[];
  total: number;
  page: number;
  limit: number;
}

export interface AdminError {
  message: string;
  code?: string;
  field?: string;
  details?: Record<string, any>;
}
export interface User {
  id: string;
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  avatar?: string;
  roles: Role[];
  permissions: Permission[];
  clusters: ClusterAccess[];
  preferences: UserPreferences;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
  isActive: boolean;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions: Permission[];
  isSystemRole: boolean;
}

export interface Permission {
  id: string;
  name: string;
  description?: string;
  category: 'cluster' | 'audit' | 'user' | 'system';
  actions: string[];
}

export interface ClusterAccess {
  clusterId: string;
  clusterName: string;
  permissions: Permission[];
  roles: Role[];
  canExecuteCommands: boolean;
  canApproveCommands: boolean;
  canViewAuditLogs: boolean;
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'system';
  language: string;
  timezone: string;
  notifications: {
    email: boolean;
    browser: boolean;
    mobile: boolean;
    commandApprovals: boolean;
    systemAlerts: boolean;
    auditAlerts: boolean;
  };
  dashboard: {
    defaultCluster?: string;
    layout: 'compact' | 'comfortable';
    refreshInterval: number;
  };
}

export interface Session {
  id: string;
  userId: string;
  token: string;
  refreshToken: string;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
  ipAddress?: string;
  userAgent?: string;
  isActive: boolean;
}

export interface SecurityEvent {
  id: string;
  type: 'login' | 'logout' | 'failed_login' | 'permission_denied' | 'suspicious_activity';
  userId?: string;
  sessionId?: string;
  description: string;
  ipAddress?: string;
  userAgent?: string;
  metadata?: Record<string, any>;
  severity: 'low' | 'medium' | 'high' | 'critical';
  timestamp: string;
}
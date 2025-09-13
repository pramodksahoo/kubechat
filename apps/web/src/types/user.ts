export interface UserPreferences {
  theme: {
    mode: 'light' | 'dark' | 'system';
    primaryColor: string;
    fontSize: 'small' | 'medium' | 'large';
  };
  notifications: {
    email: boolean;
    desktop: boolean;
    sound: boolean;
    webhooks: boolean;
  };
  security: {
    sessionTimeout: number;
    requireTwoFactor: boolean;
    allowRememberMe: boolean;
    logoutOnClose: boolean;
  };
  dashboard: {
    defaultView: 'grid' | 'list';
    refreshInterval: number;
    showMetrics: boolean;
    compactMode: boolean;
    defaultCluster?: string;
  };
  accessibility: {
    highContrast: boolean;
    reducedMotion: boolean;
    screenReader: boolean;
    keyboardNavigation: boolean;
  };
  language: string;
  timezone: string;
}

export interface Permission {
  id: string;
  name: string;
  description: string;
  resource: string;
  action: string;
  category: 'cluster' | 'namespace' | 'system' | 'user';
}

export interface Role {
  id: string;
  name: string;
  description: string;
  permissions: Permission[];
  isSystem: boolean;
}

export interface User {
  id: string;
  email: string;
  username: string;
  firstName?: string;
  lastName?: string;
  avatar?: string;
  role: 'admin' | 'user' | 'viewer';
  roles: Role[];
  permissions: Permission[];
  preferences: UserPreferences;
  createdAt: Date;
  updatedAt: Date;
  lastLoginAt?: Date;
  isActive: boolean;
  clusters: ClusterAccess[];
}

export interface ClusterAccess {
  clusterId: string;
  clusterName: string;
  namespace: string;
  permissions: Permission[];
  roles: Role[];
  canExecuteCommands: boolean;
  canApproveCommands: boolean;
  canViewAuditLogs: boolean;
  expiresAt?: Date;
}
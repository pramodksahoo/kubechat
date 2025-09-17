// Admin state management store using Zustand
// Following coding standards from docs/architecture/coding-standards.md
// AC 1-7: Admin User Management & RBAC Interface State

import { create } from 'zustand';
import type { AdminState } from '../types/admin';
import { adminUserService } from '../services/admin/userManagementService';
import { adminRoleService } from '../services/admin/roleManagementService';
import { adminSessionService } from '../services/admin/sessionService';
import { adminAuditService } from '../services/admin/auditService';
import { adminCredentialSyncService } from '../services/admin/adminCredentialSyncService';

export const useAdminStore = create<AdminState>()((set, get) => ({
  // Initial state
  users: [],
  selectedUser: null,
  userSearchQuery: '',
  userFilters: {},
  roles: [],
  permissions: [],
  selectedRole: null,
  activeSessions: [],
  sessions: [],
  sessionFilters: {},
  auditLogs: { logs: [], total: 0, page: 1, limit: 50 },
  auditFilters: {},
  securityPolicies: [],
  adminCredentials: {
    syncStatus: 'pending_initial_sync',
    lastSync: '',
    passwordExpiry: '',
    rotationCount: 0,
    complianceStatus: 'pending_review'
  },
  isLoading: false,
  error: null,
  currentView: 'dashboard',

  // User management actions
  loadUsers: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await adminUserService.getAllUsers();
      set({ users: response.users, isLoading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to load users', isLoading: false });
      throw error;
    }
  },

  createUser: async (userData) => {
    set({ isLoading: true, error: null });
    try {
      const user = await adminUserService.createUser(userData);
      set(state => ({
        users: [...state.users, user],
        isLoading: false
      }));
      return user;
    } catch (error: any) {
      set({ error: error.message || 'Failed to create user', isLoading: false });
      throw error;
    }
  },

  updateUser: async (userId, updates) => {
    set({ isLoading: true, error: null });
    try {
      const user = await adminUserService.updateUser(userId, updates);
      set(state => ({
        users: state.users.map(u => u.id === userId ? user : u),
        selectedUser: state.selectedUser?.id === userId ? user : state.selectedUser,
        isLoading: false
      }));
      return user;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update user', isLoading: false });
      throw error;
    }
  },

  deleteUser: async (userId) => {
    set({ isLoading: true, error: null });
    try {
      await adminUserService.deleteUser(userId);
      set(state => ({
        users: state.users.filter(u => u.id !== userId),
        selectedUser: state.selectedUser?.id === userId ? null : state.selectedUser,
        isLoading: false
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to delete user', isLoading: false });
      throw error;
    }
  },

  // Role management actions
  loadRoles: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await adminRoleService.getAllRoles();
      set({ roles: response.roles, isLoading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to load roles', isLoading: false });
      throw error;
    }
  },

  loadPermissions: async () => {
    set({ isLoading: true, error: null });
    try {
      const permissions = await adminRoleService.getAllPermissions();
      set({ permissions, isLoading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to load permissions', isLoading: false });
      throw error;
    }
  },

  createRole: async (roleData) => {
    set({ isLoading: true, error: null });
    try {
      const role = await adminRoleService.createRole(roleData);
      set(state => ({
        roles: [...state.roles, role],
        isLoading: false
      }));
      return role;
    } catch (error: any) {
      set({ error: error.message || 'Failed to create role', isLoading: false });
      throw error;
    }
  },

  updateRole: async (roleId, updates) => {
    set({ isLoading: true, error: null });
    try {
      const role = await adminRoleService.updateRole(roleId, updates);
      set(state => ({
        roles: state.roles.map(r => r.id === roleId ? role : r),
        selectedRole: state.selectedRole?.id === roleId ? role : state.selectedRole,
        isLoading: false
      }));
      return role;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update role', isLoading: false });
      throw error;
    }
  },

  deleteRole: async (roleId) => {
    set({ isLoading: true, error: null });
    try {
      await adminRoleService.deleteRole(roleId);
      set(state => ({
        roles: state.roles.filter(r => r.id !== roleId),
        selectedRole: state.selectedRole?.id === roleId ? null : state.selectedRole,
        isLoading: false
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to delete role', isLoading: false });
      throw error;
    }
  },

  // Session management actions
  loadActiveSessions: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await adminSessionService.getActiveSessions();
      set({ activeSessions: response.sessions, isLoading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to load sessions', isLoading: false });
      throw error;
    }
  },

  loadSessions: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await adminSessionService.getActiveSessions();
      set({ sessions: response.sessions, isLoading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to load sessions', isLoading: false });
      throw error;
    }
  },

  terminateSession: async (sessionId) => {
    set({ isLoading: true, error: null });
    try {
      await adminSessionService.terminateSession(sessionId);
      set(state => ({
        activeSessions: state.activeSessions.filter(s => s.id !== sessionId),
        isLoading: false
      }));
    } catch (error: any) {
      set({ error: error.message || 'Failed to terminate session', isLoading: false });
      throw error;
    }
  },

  // Audit log actions
  loadAuditLogs: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await adminAuditService.getAuditLogs();
      set({ auditLogs: response, isLoading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to load audit logs', isLoading: false });
      throw error;
    }
  },

  // Security policy actions
  loadSecurityPolicies: async () => {
    set({ isLoading: true, error: null });
    try {
      const policies = await adminUserService.getSecurityPolicies();
      set({ securityPolicies: policies, isLoading: false });
    } catch (error: any) {
      set({ error: error.message || 'Failed to load security policies', isLoading: false });
      throw error;
    }
  },

  updateSecurityPolicy: async (policyId, updates) => {
    set({ isLoading: true, error: null });
    try {
      const policy = await adminUserService.updateSecurityPolicy(policyId, updates);
      set(state => ({
        securityPolicies: state.securityPolicies.map(p => p.id === policyId ? policy : p),
        isLoading: false
      }));
      return policy;
    } catch (error: any) {
      set({ error: error.message || 'Failed to update security policy', isLoading: false });
      throw error;
    }
  },

  // Admin credential management actions (AC 8)
  syncAdminCredentials: async () => {
    set({ isLoading: true, error: null });
    try {
      const syncResult = await adminCredentialSyncService.syncFromK8sSecret();
      const latestStatus = await adminCredentialSyncService.getLatestSyncStatus();

      if (latestStatus) {
        set(state => ({
          adminCredentials: {
            syncStatus: latestStatus.sync_status as any,
            lastSync: latestStatus.sync_timestamp,
            passwordExpiry: latestStatus.password_expires_at,
            rotationCount: latestStatus.rotation_count,
            complianceStatus: latestStatus.compliance_status as any,
            k8sSecretVersion: latestStatus.k8s_resource_version,
            syncSource: latestStatus.sync_source as any
          },
          isLoading: false
        }));
      }
    } catch (error: any) {
      set({ error: error.message || 'Failed to sync admin credentials', isLoading: false });
      throw error;
    }
  },

  rotateAdminPassword: async (newPassword) => {
    set({ isLoading: true, error: null });
    try {
      const result = await adminCredentialSyncService.rotateCredentials(newPassword);

      // Update admin credentials status
      const latestStatus = await adminCredentialSyncService.getLatestSyncStatus();
      if (latestStatus) {
        set(state => ({
          adminCredentials: {
            ...state.adminCredentials,
            lastSync: latestStatus.sync_timestamp,
            passwordExpiry: latestStatus.password_expires_at,
            rotationCount: latestStatus.rotation_count,
            syncStatus: latestStatus.sync_status as any
          },
          isLoading: false
        }));
      }

      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to rotate admin password', isLoading: false });
      throw error;
    }
  },

  validateCredentialCompliance: async () => {
    set({ isLoading: true, error: null });
    try {
      const result = await adminCredentialSyncService.validateCompliance();

      set(state => ({
        adminCredentials: {
          ...state.adminCredentials,
          complianceStatus: result.overall as any
        },
        isLoading: false
      }));

      return result;
    } catch (error: any) {
      set({ error: error.message || 'Failed to validate compliance', isLoading: false });
      throw error;
    }
  },

  requestEmergencyAccess: async (justification, businessImpact) => {
    set({ isLoading: true, error: null });
    try {
      const request = await adminCredentialSyncService.requestEmergencyAccess({
        emergency_type: 'credential_recovery',
        justification,
        business_impact: businessImpact
      });

      set({ isLoading: false });
      return request;
    } catch (error: any) {
      set({ error: error.message || 'Failed to request emergency access', isLoading: false });
      throw error;
    }
  },

  // UI actions
  setCurrentView: (view) => {
    set({ currentView: view, error: null });
  },

  setError: (error) => {
    set({ error });
  },

  clearError: () => {
    set({ error: null });
  }
}));

// Initialize admin store with current user context
export const initializeAdminStore = async () => {
  const { loadUsers, loadRoles, syncAdminCredentials } = useAdminStore.getState();

  try {
    // Load initial data in parallel
    await Promise.all([
      loadUsers(),
      loadRoles(),
      syncAdminCredentials()
    ]);
  } catch (error) {
    console.error('Failed to initialize admin store:', error);
    useAdminStore.getState().setError('Failed to initialize admin interface');
  }
};
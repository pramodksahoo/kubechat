// Admin User Management Service
// Following coding standards from docs/architecture/coding-standards.md
// AC 2: User Management CRUD Interface with API integration

import { httpClient } from '../api';
import type {
  AdminUser,
  CreateUserRequest,
  UpdateUserRequest,
  AdminUsersResponse,
  SecurityPolicy,
  AdminError
} from '../../types/admin';

class AdminUserManagementService {
  // User CRUD operations
  async getAllUsers(page = 1, limit = 50, filters?: {
    role?: string;
    status?: string;
    search?: string;
  }): Promise<AdminUsersResponse> {
    try {
      const params = new URLSearchParams({
        page: page.toString(),
        limit: limit.toString(),
        ...(filters?.role && { role: filters.role }),
        ...(filters?.status && { status: filters.status }),
        ...(filters?.search && { search: filters.search })
      });

      const response = await httpClient.get(`/api/v1/admin/users?${params.toString()}`);

      if (!response.data) {
        throw new Error('Invalid response from user management API');
      }

      const data = response.data as any;
      return {
        users: data.users.map(this.transformUserData),
        total: data.total || data.users.length,
        page: data.page || page,
        limit: data.limit || limit
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load users');
    }
  }

  async getUserById(userId: string): Promise<AdminUser> {
    try {
      const response = await httpClient.get(`/api/v1/admin/users/${userId}`);

      if (!response.data) {
        throw new Error('User not found');
      }

      return this.transformUserData(response.data);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load user');
    }
  }

  async createUser(userData: CreateUserRequest): Promise<AdminUser> {
    try {
      // Validate user data
      this.validateUserData(userData);

      const response = await httpClient.post('/api/v1/admin/users', {
        username: userData.username,
        email: userData.email,
        password: userData.password,
        role: userData.role,
        permissions: userData.permissions || [],
        clusters: userData.clusters || [],
        require_password_change: userData.requirePasswordChange || false,
        mfa_required: userData.mfaRequired || false
      });

      if (!response.data) {
        throw new Error('Failed to create user - invalid response');
      }

      return this.transformUserData(response.data);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to create user');
    }
  }

  async updateUser(userId: string, updates: UpdateUserRequest): Promise<AdminUser> {
    try {
      // Validate update data
      this.validateUpdateData(updates);

      const response = await httpClient.put(`/api/v1/admin/users/${userId}`, {
        email: updates.email,
        role: updates.role,
        permissions: updates.permissions,
        clusters: updates.clusters,
        is_active: updates.isActive,
        mfa_enabled: updates.mfaEnabled,
        account_locked: updates.accountLocked,
        reset_password: updates.resetPassword || false
      });

      if (!response.data) {
        throw new Error('Failed to update user - invalid response');
      }

      return this.transformUserData(response.data);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to update user');
    }
  }

  async deleteUser(userId: string): Promise<void> {
    try {
      await httpClient.delete(`/api/v1/admin/users/${userId}`);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to delete user');
    }
  }

  async resetUserPassword(userId: string, newPassword?: string): Promise<{ tempPassword?: string }> {
    try {
      const response = await httpClient.post(`/api/v1/admin/users/${userId}/reset-password`, {
        new_password: newPassword
      });

      return {
        tempPassword: ((response.data as any))?.temp_password
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to reset user password');
    }
  }

  async lockUser(userId: string, reason?: string): Promise<void> {
    try {
      await httpClient.post(`/api/v1/admin/users/${userId}/lock`, {
        reason: reason || 'Administrative action'
      });
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to lock user account');
    }
  }

  async unlockUser(userId: string): Promise<void> {
    try {
      await httpClient.post(`/api/v1/admin/users/${userId}/unlock`);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to unlock user account');
    }
  }

  // Security policy management
  async getSecurityPolicies(): Promise<SecurityPolicy[]> {
    try {
      const response = await httpClient.get('/api/v1/admin/security/policies');

      return ((response.data as any))?.policies || [];
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load security policies');
    }
  }

  async updateSecurityPolicy(policyId: string, updates: Partial<SecurityPolicy>): Promise<SecurityPolicy> {
    try {
      const response = await httpClient.put(`/api/v1/admin/security/policies/${policyId}`, updates);

      if (!response.data) {
        throw new Error('Failed to update security policy');
      }

      return response.data as SecurityPolicy;
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to update security policy');
    }
  }

  // Bulk operations
  async bulkUpdateUsers(userIds: string[], updates: Partial<UpdateUserRequest>): Promise<AdminUser[]> {
    try {
      const response = await httpClient.post('/api/v1/admin/users/bulk-update', {
        user_ids: userIds,
        updates: updates
      });

      return ((response.data as any))?.users?.map(this.transformUserData) || [];
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to bulk update users');
    }
  }

  async bulkDeleteUsers(userIds: string[]): Promise<void> {
    try {
      await httpClient.post('/api/v1/admin/users/bulk-delete', {
        user_ids: userIds
      });
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to bulk delete users');
    }
  }

  // User statistics
  async getUserStatistics(): Promise<{
    total: number;
    active: number;
    inactive: number;
    locked: number;
    byRole: Record<string, number>;
    recentLogins: number;
  }> {
    try {
      const response = await httpClient.get('/api/v1/admin/users/statistics');

      const data = response.data as any;
      return data || {
        total: 0,
        active: 0,
        inactive: 0,
        locked: 0,
        byRole: {},
        recentLogins: 0
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load user statistics');
    }
  }

  // Private helper methods
  private transformUserData(userData: any): AdminUser {
    return {
      id: userData.id,
      username: userData.username,
      email: userData.email,
      role: userData.role,
      permissions: userData.permissions || [],
      clusters: userData.clusters || [],
      createdAt: userData.created_at || userData.createdAt,
      updatedAt: userData.updated_at || userData.updatedAt,
      lastLoginAt: userData.last_login_at || userData.lastLoginAt,
      isActive: userData.is_active !== undefined ? userData.is_active : userData.isActive !== false,
      mfaEnabled: userData.mfa_enabled || userData.mfaEnabled || false,
      accountLocked: userData.account_locked || userData.accountLocked || false,
      failedLoginAttempts: userData.failed_login_attempts || userData.failedLoginAttempts || 0,
      lastPasswordChange: userData.last_password_change || userData.lastPasswordChange
    };
  }

  private validateUserData(userData: CreateUserRequest): void {
    if (!userData.username || userData.username.trim().length < 3) {
      throw new Error('Username must be at least 3 characters long');
    }

    if (!userData.email || !this.isValidEmail(userData.email)) {
      throw new Error('Valid email address is required');
    }

    if (!userData.password || userData.password.length < 8) {
      throw new Error('Password must be at least 8 characters long');
    }

    if (!userData.role) {
      throw new Error('User role is required');
    }

    // Check for valid role
    const validRoles = ['admin', 'user', 'viewer', 'auditor', 'compliance_officer'];
    if (!validRoles.includes(userData.role)) {
      throw new Error(`Invalid role. Must be one of: ${validRoles.join(', ')}`);
    }
  }

  private validateUpdateData(updates: UpdateUserRequest): void {
    if (updates.email && !this.isValidEmail(updates.email)) {
      throw new Error('Valid email address is required');
    }

    if (updates.role) {
      const validRoles = ['admin', 'user', 'viewer', 'auditor', 'compliance_officer'];
      if (!validRoles.includes(updates.role)) {
        throw new Error(`Invalid role. Must be one of: ${validRoles.join(', ')}`);
      }
    }
  }

  private isValidEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  private handleApiError(error: any, defaultMessage: string): AdminError {
    if (error.response?.data?.message) {
      return new Error(error.response.data.message);
    }

    if (error.response?.status === 401) {
      return new Error('Unauthorized access - admin privileges required');
    }

    if (error.response?.status === 403) {
      return new Error('Forbidden - insufficient permissions');
    }

    if (error.response?.status === 404) {
      return new Error('Resource not found');
    }

    if (error.response?.status >= 500) {
      return new Error('Server error - please try again later');
    }

    return new Error(error.message || defaultMessage);
  }
}

export const adminUserService = new AdminUserManagementService();
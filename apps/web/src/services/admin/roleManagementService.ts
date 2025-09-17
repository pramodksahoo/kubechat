// Admin Role Management Service
// Following coding standards from docs/architecture/coding-standards.md
// AC 4: RBAC Management Interface

import { httpClient } from '../api';
import type { Role, Permission, AdminRolesResponse } from '../../types/admin';

class AdminRoleManagementService {
  async getAllRoles(): Promise<AdminRolesResponse> {
    try {
      const response = await httpClient.get('/api/v1/admin/roles');

      return {
        roles: ((response.data as any))?.roles || [],
        total: ((response.data as any))?.total || 0
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load roles');
    }
  }

  async createRole(roleData: Omit<Role, 'id' | 'createdAt' | 'updatedAt'>): Promise<Role> {
    try {
      const response = await httpClient.post('/api/v1/admin/roles', roleData);
      return response.data as Role;
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to create role');
    }
  }

  async updateRole(roleId: string, updates: Partial<Role>): Promise<Role> {
    try {
      const response = await httpClient.put(`/api/v1/admin/roles/${roleId}`, updates);
      return response.data as Role;
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to update role');
    }
  }

  async deleteRole(roleId: string): Promise<void> {
    try {
      await httpClient.delete(`/api/v1/admin/roles/${roleId}`);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to delete role');
    }
  }

  async getAllPermissions(): Promise<Permission[]> {
    try {
      const response = await httpClient.get('/api/v1/admin/permissions');
      return ((response.data as any))?.permissions || [];
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load permissions');
    }
  }

  private handleApiError(error: any, defaultMessage: string): Error {
    return new Error(error.response?.data?.message || error.message || defaultMessage);
  }
}

export const adminRoleService = new AdminRoleManagementService();
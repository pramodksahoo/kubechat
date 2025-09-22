// Admin Audit Service
// Following coding standards from docs/architecture/coding-standards.md
// AC 7: Audit Integration Interface

import { httpClient } from '../api';
import type { AuditLogEntry, AdminAuditResponse } from '../../types/admin';

class AdminAuditService {
  async getAuditLogs(page = 1, limit = 50): Promise<AdminAuditResponse> {
    try {
      const response = await httpClient.get(`/audit/logs?page=${page}&limit=${limit}`);

      return {
        logs: ((response.data as any))?.logs || [],
        total: ((response.data as any))?.total || 0,
        page: ((response.data as any))?.page || page,
        limit: ((response.data as any))?.limit || limit
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load audit logs');
    }
  }

  private handleApiError(error: any, defaultMessage: string): Error {
    return new Error(error.response?.data?.message || error.message || defaultMessage);
  }
}

export const adminAuditService = new AdminAuditService();
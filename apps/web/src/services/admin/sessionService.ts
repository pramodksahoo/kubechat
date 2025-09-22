// Admin Session Management Service
// Following coding standards from docs/architecture/coding-standards.md
// AC 5: Session Management Dashboard

import { httpClient } from '../api';
import type { UserSession, AdminSessionsResponse } from '../../types/admin';

class AdminSessionService {
  async getActiveSessions(): Promise<AdminSessionsResponse> {
    try {
      const response = await httpClient.get('/api/v1/auth/admin/sessions');

      return {
        sessions: ((response.data as any))?.sessions || [],
        total: ((response.data as any))?.total || 0,
        active: ((response.data as any))?.active || 0
      };
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to load sessions');
    }
  }

  async terminateSession(sessionId: string): Promise<void> {
    try {
      await httpClient.delete(`/api/v1/auth/admin/sessions/${sessionId}`);
    } catch (error: any) {
      throw this.handleApiError(error, 'Failed to terminate session');
    }
  }

  private handleApiError(error: any, defaultMessage: string): Error {
    return new Error(error.response?.data?.message || error.message || defaultMessage);
  }
}

export const adminSessionService = new AdminSessionService();
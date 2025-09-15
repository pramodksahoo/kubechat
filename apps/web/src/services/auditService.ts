import { api } from './api';
import {
  AuditLogEntry,
  AuditSummary,
  AuditFilter,
  ExportFormat,
  ChainIntegrityResult,
  TamperAlert,
  SuspiciousActivity,
  LegalHold,
  LegalHoldRequest,
  ComplianceReport,
  ComplianceReportRequest,
  RetentionPolicy,
  ArchivalResult,
  RealTimeMonitoringStatus
} from '@kubechat/shared/types';

export interface AuditMetrics {
  totalEvents: number;
  successEvents: number;
  failureEvents: number;
  errorEvents: number;
  criticalEvents: number;
  averageResponseTime: number;
  topUsers: {
    userId: string;
    userName?: string;
    eventCount: number;
  }[];
  topActions: {
    action: string;
    eventCount: number;
  }[];
  timeline: {
    date: string;
    eventCount: number;
    successRate: number;
  }[];
}

export interface AuditExportOptions {
  format: 'json' | 'csv' | 'pdf';
  filters?: AuditFilter;
  includeMetadata?: boolean;
  startDate?: string;
  endDate?: string;
}

class AuditService {
  // Audit Log Management
  async getAuditLogs(filters?: AuditFilter): Promise<AuditLogEntry[]> {
    try {
      const response = await api.audit.getLogs({
        startTime: filters?.startDate,
        endTime: filters?.endDate,
        limit: 100,
        page: 1
      });

      return this.mapAuditLogsFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch audit logs:', error);
      return [];
    }
  }

  async getAuditLogEntry(entryId: string): Promise<AuditLogEntry | null> {
    try {
      const response = await api.audit.getLog(entryId);
      return this.mapAuditLogFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch audit log entry:', error);
      return null;
    }
  }

  // Audit Metrics and Summary
  async getAuditSummary(filters?: AuditFilter): Promise<AuditSummary> {
    try {
      const response = await api.audit.getMetrics();

      return this.mapAuditSummaryFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch audit summary:', error);
      return this.getDefaultAuditSummary();
    }
  }

  async getAuditMetrics(period: '1h' | '24h' | '7d' | '30d' = '24h'): Promise<AuditMetrics> {
    try {
      const response = await api.audit.getMetrics();
      return this.mapAuditMetricsFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch audit metrics:', error);
      return this.getDefaultAuditMetrics();
    }
  }

  // User Activity Analysis
  async getUserActivity(userId?: string, limit: number = 50): Promise<AuditLogEntry[]> {
    try {
      // Since getUserActivity doesn't exist, use getLogs instead
      const response = await api.audit.getLogs({
        limit,
        page: 1
      });

      return this.mapAuditLogsFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch user activity:', error);
      return [];
    }
  }

  async getTopUsers(limit: number = 10, period: '24h' | '7d' | '30d' = '7d'): Promise<{
    userId: string;
    userName?: string;
    eventCount: number;
    lastActivity: string;
  }[]> {
    try {
      // Since getTopUsers doesn't exist, return empty array
      return [];
    } catch (error) {
      console.error('Failed to fetch top users:', error);
      return [];
    }
  }

  // Export Functionality
  async exportAuditLogs(options: AuditExportOptions): Promise<Blob> {
    try {
      // Since exportLogs doesn't exist, create export from getLogs data
      const response = await api.audit.getLogs({
        startTime: options.startDate,
        endTime: options.endDate,
        limit: 1000,
        page: 1
      });

      // Handle different response types based on format
      if (options.format === 'json') {
        return new Blob([JSON.stringify(response.data, null, 2)], { type: 'application/json' });
      } else if (options.format === 'csv') {
        // Convert to CSV format
        const csvData = this.convertToCSV(response.data);
        return new Blob([csvData], { type: 'text/csv' });
      } else if (options.format === 'pdf') {
        // For PDF, return JSON for now
        return new Blob([JSON.stringify(response.data, null, 2)], { type: 'application/json' });
      }

      return new Blob([JSON.stringify(response.data)], { type: 'application/json' });
    } catch (error) {
      console.error('Failed to export audit logs:', error);
      throw error;
    }
  }

  // Real-time Audit Stream
  async getRecentAuditEvents(limit: number = 20): Promise<AuditLogEntry[]> {
    try {
      const response = await api.audit.getLogs({ limit, page: 1 });
      return this.mapAuditLogsFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch recent audit events:', error);
      return [];
    }
  }

  // Search and Filter
  async searchAuditLogs(query: string, filters?: Partial<AuditFilter>): Promise<AuditLogEntry[]> {
    try {
      // Since searchLogs doesn't exist, use getLogs and filter client-side
      const response = await api.audit.getLogs({ limit: 100, page: 1 });
      const logs = this.mapAuditLogsFromBackend(response.data);

      // Simple client-side filtering by query
      return logs.filter(log =>
        log.action?.toLowerCase().includes(query.toLowerCase()) ||
        log.resource?.toLowerCase().includes(query.toLowerCase()) ||
        log.userName?.toLowerCase().includes(query.toLowerCase())
      );
    } catch (error) {
      console.error('Failed to search audit logs:', error);
      return [];
    }
  }

  // Data mapping functions
  private mapAuditLogsFromBackend(data: any): AuditLogEntry[] {
    if (!Array.isArray(data)) {
      // Handle case where data is wrapped in a result object
      if (data && Array.isArray(data.logs)) {
        data = data.logs;
      } else if (data && Array.isArray(data.entries)) {
        data = data.entries;
      } else if (data && Array.isArray(data.data)) {
        data = data.data;
      } else {
        return [];
      }
    }

    return data.map((entry: any) => this.mapAuditLogFromBackend(entry)).filter(Boolean) as AuditLogEntry[];
  }

  private mapAuditLogFromBackend(data: any): AuditLogEntry | null {
    if (!data || typeof data !== 'object') {
      return null;
    }

    return {
      id: data.id || data._id || data.entryId || `audit-${Date.now()}`,
      timestamp: data.timestamp || data.createdAt || data.eventTime || new Date().toISOString(),
      userId: data.userId || data.user_id || data.user?.id || 'unknown',
      userName: data.userName || data.user_name || data.user?.name || data.user?.username || 'Unknown User',
      sessionId: data.sessionId || data.session_id || data.session || '',
      action: data.action || data.operation || data.event || 'unknown_action',
      resource: data.resource || data.target || data.object || 'unknown_resource',
      resourceId: data.resourceId || data.resource_id || data.targetId,
      method: this.mapHttpMethod(data.method || data.httpMethod || data.verb),
      result: this.mapResult(data.result || data.status || data.outcome),
      statusCode: data.statusCode || data.status_code || data.responseCode,
      duration: data.duration || data.responseTime || data.elapsed,
      ipAddress: data.ipAddress || data.ip_address || data.clientIP || data.remoteAddr,
      userAgent: data.userAgent || data.user_agent || data.clientAgent,
      endpoint: data.endpoint || data.url || data.path || data.api,
      command: data.command || data.kubectl || data.commandLine,
      clusterId: data.clusterId || data.cluster_id || data.cluster?.id,
      clusterName: data.clusterName || data.cluster_name || data.cluster?.name,
      severity: this.mapSeverity(data.severity || data.level || data.priority),
      errorMessage: data.errorMessage || data.error_message || data.error || data.failure_reason,
      metadata: data.metadata || data.additional_data || data.context || {}
    };
  }

  private mapAuditSummaryFromBackend(data: any): AuditSummary {
    if (!data || typeof data !== 'object') {
      return this.getDefaultAuditSummary();
    }

    return {
      totalEvents: data.totalEvents || data.total_events || data.count || 0,
      successEvents: data.successEvents || data.success_events || data.successful || 0,
      failureEvents: data.failureEvents || data.failure_events || data.failed || 0,
      errorEvents: data.errorEvents || data.error_events || data.errors || 0,
      criticalEvents: data.criticalEvents || data.critical_events || data.critical || 0,
      topUsers: this.mapTopUsersFromBackend(data.topUsers || data.top_users || []),
      topActions: this.mapTopActionsFromBackend(data.topActions || data.top_actions || []),
      timeline: this.mapTimelineFromBackend(data.timeline || data.activity_timeline || [])
    };
  }

  private mapAuditMetricsFromBackend(data: any): AuditMetrics {
    if (!data || typeof data !== 'object') {
      return this.getDefaultAuditMetrics();
    }

    return {
      totalEvents: data.totalEvents || data.total || 0,
      successEvents: data.successEvents || data.successful || 0,
      failureEvents: data.failureEvents || data.failed || 0,
      errorEvents: data.errorEvents || data.errors || 0,
      criticalEvents: data.criticalEvents || data.critical || 0,
      averageResponseTime: data.averageResponseTime || data.avg_response_time || 0,
      topUsers: this.mapTopUsersFromBackend(data.topUsers || data.users || []),
      topActions: this.mapTopActionsFromBackend(data.topActions || data.actions || []),
      timeline: this.mapTimelineFromBackend(data.timeline || data.history || [])
    };
  }

  private mapTopUsersFromBackend(data: any[]): { userId: string; userName: string; eventCount: number; }[] {
    if (!Array.isArray(data)) return [];

    return data.map(user => ({
      userId: user.userId || user.user_id || user.id || 'unknown',
      userName: user.userName || user.user_name || user.name || user.username || 'Unknown',
      eventCount: user.eventCount || user.event_count || user.count || 0
    }));
  }

  private mapTopActionsFromBackend(data: any[]): { action: string; eventCount: number; }[] {
    if (!Array.isArray(data)) return [];

    return data.map(action => ({
      action: action.action || action.operation || action.event || 'unknown',
      eventCount: action.eventCount || action.event_count || action.count || 0
    }));
  }

  private mapTimelineFromBackend(data: any[]): { date: string; eventCount: number; successRate: number; }[] {
    if (!Array.isArray(data)) return [];

    return data.map(day => ({
      date: day.date || day.timestamp || new Date().toISOString().split('T')[0],
      eventCount: day.eventCount || day.event_count || day.count || 0,
      successRate: day.successRate !== undefined ? day.successRate :
                   day.success_rate !== undefined ? day.success_rate :
                   day.eventCount > 0 ? ((day.successful || 0) / day.eventCount) * 100 : 0
    }));
  }

  // Helper mapping functions
  private mapHttpMethod(method: any): 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' {
    if (!method) return 'GET';

    const methodStr = method.toString().toUpperCase();
    const validMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'];

    return validMethods.includes(methodStr) ? methodStr as any : 'GET';
  }

  private mapResult(result: any): 'success' | 'failure' | 'error' {
    if (!result) return 'success';

    const resultStr = result.toString().toLowerCase();

    if (resultStr.includes('success') || resultStr.includes('ok') || resultStr === '200' || resultStr === 'completed') {
      return 'success';
    } else if (resultStr.includes('error') || resultStr.includes('exception') || resultStr.includes('5')) {
      return 'error';
    } else if (resultStr.includes('fail') || resultStr.includes('denied') || resultStr.includes('4')) {
      return 'failure';
    }

    return 'success';
  }

  private mapSeverity(severity: any): 'low' | 'medium' | 'high' | 'critical' {
    if (!severity) return 'low';

    const severityStr = severity.toString().toLowerCase();

    if (severityStr.includes('critical') || severityStr.includes('urgent')) return 'critical';
    if (severityStr.includes('high') || severityStr.includes('severe')) return 'high';
    if (severityStr.includes('medium') || severityStr.includes('warn')) return 'medium';

    return 'low';
  }

  // Default data functions
  private getDefaultAuditSummary(): AuditSummary {
    return {
      totalEvents: 0,
      successEvents: 0,
      failureEvents: 0,
      errorEvents: 0,
      criticalEvents: 0,
      topUsers: [],
      topActions: [],
      timeline: []
    };
  }

  private getDefaultAuditMetrics(): AuditMetrics {
    return {
      totalEvents: 0,
      successEvents: 0,
      failureEvents: 0,
      errorEvents: 0,
      criticalEvents: 0,
      averageResponseTime: 0,
      topUsers: [],
      topActions: [],
      timeline: []
    };
  }

  // Enhanced methods for Story 1.8

  /**
   * Verify audit chain integrity
   */
  async verifyChainIntegrity(): Promise<ChainIntegrityResult> {
    try {
      const response = await fetch('/api/v1/audit/chain-integrity');
      if (!response.ok) {
        throw new Error(`Failed to verify chain integrity: ${response.statusText}`);
      }
      const data = await response.json();
      return data.data;
    } catch (error) {
      console.error('Failed to verify chain integrity:', error);
      return {
        isValid: false,
        totalChecked: 0,
        violations: [],
        lastValidated: new Date().toISOString(),
        integrityScore: 0
      };
    }
  }

  /**
   * Get suspicious activities
   */
  async getSuspiciousActivities(timeWindow: string = '24h'): Promise<SuspiciousActivity[]> {
    try {
      const response = await fetch(`/api/v1/audit/suspicious-activities?time_window=${timeWindow}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch suspicious activities: ${response.statusText}`);
      }
      const data = await response.json();
      return data.data || [];
    } catch (error) {
      console.error('Failed to fetch suspicious activities:', error);
      return [];
    }
  }

  /**
   * Export audit logs with enhanced format support
   */
  async exportAuditLogsEnhanced(format: ExportFormat, filters: AuditFilter): Promise<void> {
    try {
      const params = new URLSearchParams();
      params.append('format', format);

      if (filters.startDate) params.append('start_time', filters.startDate);
      if (filters.endDate) params.append('end_time', filters.endDate);
      if (filters.userId) params.append('user_id', filters.userId);
      if (filters.clusterId) params.append('cluster_id', filters.clusterId);

      const response = await fetch(`/api/v1/audit/export?${params}`);
      if (!response.ok) {
        throw new Error(`Failed to export audit logs: ${response.statusText}`);
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `audit-logs-${format}-${new Date().toISOString().split('T')[0]}.${format === 'csv' ? 'csv' : format === 'pdf' ? 'pdf' : 'json'}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Failed to export audit logs:', error);
      throw error;
    }
  }

  /**
   * Generate compliance report
   */
  async generateComplianceReport(request: ComplianceReportRequest): Promise<ComplianceReport> {
    try {
      const response = await fetch('/api/v1/audit/compliance-report', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        throw new Error(`Failed to generate compliance report: ${response.statusText}`);
      }

      const data = await response.json();
      return data.data;
    } catch (error) {
      console.error('Failed to generate compliance report:', error);
      throw error;
    }
  }

  /**
   * Create legal hold
   */
  async createLegalHold(request: LegalHoldRequest): Promise<LegalHold> {
    try {
      const response = await fetch('/api/v1/audit/legal-hold', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        throw new Error(`Failed to create legal hold: ${response.statusText}`);
      }

      const data = await response.json();
      return data.data;
    } catch (error) {
      console.error('Failed to create legal hold:', error);
      throw error;
    }
  }

  /**
   * Release legal hold
   */
  async releaseLegalHold(holdId: string): Promise<void> {
    try {
      const response = await fetch(`/api/v1/audit/legal-hold/${holdId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error(`Failed to release legal hold: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Failed to release legal hold:', error);
      throw error;
    }
  }

  /**
   * Get all legal holds
   */
  async getLegalHolds(): Promise<LegalHold[]> {
    try {
      const response = await fetch('/api/v1/audit/legal-holds');
      if (!response.ok) {
        throw new Error(`Failed to fetch legal holds: ${response.statusText}`);
      }

      const data = await response.json();
      return data.data || [];
    } catch (error) {
      console.error('Failed to fetch legal holds:', error);
      return [];
    }
  }

  /**
   * Apply retention policy
   */
  async applyRetentionPolicy(policy: RetentionPolicy): Promise<void> {
    try {
      const response = await fetch('/api/v1/audit/retention-policy', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(policy),
      });

      if (!response.ok) {
        throw new Error(`Failed to apply retention policy: ${response.statusText}`);
      }
    } catch (error) {
      console.error('Failed to apply retention policy:', error);
      throw error;
    }
  }

  /**
   * Archive old logs
   */
  async archiveOldLogs(cutoffDate: string): Promise<ArchivalResult> {
    try {
      const response = await fetch('/api/v1/audit/archive', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ cutoff_date: cutoffDate }),
      });

      if (!response.ok) {
        throw new Error(`Failed to archive logs: ${response.statusText}`);
      }

      const data = await response.json();
      return data.data;
    } catch (error) {
      console.error('Failed to archive logs:', error);
      throw error;
    }
  }

  /**
   * Start real-time monitoring
   */
  async startRealTimeMonitoring(): Promise<RealTimeMonitoringStatus> {
    try {
      const response = await fetch('/api/v1/audit/monitor');
      if (!response.ok) {
        throw new Error(`Failed to start real-time monitoring: ${response.statusText}`);
      }

      return {
        isActive: true,
        connectedSessions: 1,
        eventsPerSecond: 0,
        lastEventAt: new Date().toISOString(),
      };
    } catch (error) {
      console.error('Failed to start real-time monitoring:', error);
      return {
        isActive: false,
        connectedSessions: 0,
        eventsPerSecond: 0,
      };
    }
  }

  /**
   * Get audit service health
   */
  async getAuditHealth(): Promise<{ status: string; message: string }> {
    try {
      const response = await fetch('/api/v1/audit/health');
      if (!response.ok) {
        throw new Error(`Audit service health check failed: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Audit service health check failed:', error);
      return { status: 'unhealthy', message: 'Service unavailable' };
    }
  }

  // CSV conversion helper
  private convertToCSV(data: any): string {
    if (!Array.isArray(data) || data.length === 0) {
      return 'timestamp,userId,action,resource,result\n';
    }

    const headers = ['timestamp', 'userId', 'action', 'resource', 'result'];
    const csvRows = [headers.join(',')];

    data.forEach(item => {
      const row = headers.map(header => {
        const value = item[header] || '';
        return `"${value.toString().replace(/"/g, '""')}"`;
      });
      csvRows.push(row.join(','));
    });

    return csvRows.join('\n');
  }

  // Utility functions for download
  downloadBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  generateFilename(format: string, filters?: AuditFilter): string {
    const timestamp = new Date().toISOString().split('T')[0];
    const baseFilename = `audit-logs-${timestamp}`;

    if (filters?.userId) {
      return `${baseFilename}-user-${filters.userId}.${format}`;
    }
    if (filters?.clusterId) {
      return `${baseFilename}-cluster-${filters.clusterId}.${format}`;
    }

    return `${baseFilename}.${format}`;
  }
}

export const auditService = new AuditService();
export default auditService;
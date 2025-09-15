import React, { useState, useEffect, useCallback } from 'react';
import {
  AuditLogEntry,
  AuditFilter,
  AuditSummary,
  AuditMetrics,
  ExportFormat,
  ChainIntegrityResult,
  TamperAlert,
  SuspiciousActivity
} from '@kubechat/shared/types';
import { AuditSearchPanel } from './AuditSearchPanel';
import { AuditExportManager } from './AuditExportManager';
import { IntegrityMonitor } from './IntegrityMonitor';
import { ComplianceReportManager } from '../compliance/ComplianceReportManager';
import { LegalHoldManager } from '../compliance/LegalHoldManager';
import { TamperDetectionMonitor } from '../security/TamperDetectionMonitor';
import { LogRetentionManager } from './LogRetentionManager';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { Icon } from '../ui/Icon';
import { format } from 'date-fns';

interface AuditTrailDashboardProps {
  className?: string;
}

export function AuditTrailDashboard({ className = '' }: AuditTrailDashboardProps) {
  const [auditEntries, setAuditEntries] = useState<AuditLogEntry[]>([]);
  const [summary, setSummary] = useState<AuditSummary | null>(null);
  const [metrics, setMetrics] = useState<AuditMetrics | null>(null);
  const [filters, setFilters] = useState<AuditFilter>({
    startDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    endDate: new Date().toISOString(),
  });
  const [loading, setLoading] = useState(false);
  const [selectedView, setSelectedView] = useState<'logs' | 'integrity' | 'analytics' | 'compliance' | 'legal-holds' | 'tamper-detection' | 'retention'>('logs');
  const [integrityResult, setIntegrityResult] = useState<ChainIntegrityResult | null>(null);
  const [tamperAlerts] = useState<TamperAlert[]>([]);
  const [suspiciousActivities, setSuspiciousActivities] = useState<SuspiciousActivity[]>([]);

  // Load audit data
  const loadAuditData = useCallback(async (newFilters: AuditFilter) => {
    setLoading(true);
    try {
      // Simulated API calls - replace with actual API integration
      const response = await fetch('/api/v1/audit/logs?' + new URLSearchParams({
        start_time: newFilters.startDate || '',
        end_time: newFilters.endDate || '',
        user_id: newFilters.userId || '',
        action: newFilters.action || '',
        severity: newFilters.severity || '',
        limit: '50'
      }));

      if (response.ok) {
        const data = await response.json();
        setAuditEntries(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load audit data:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load summary and metrics
  const loadSummaryAndMetrics = useCallback(async () => {
    try {
      const [summaryResponse, metricsResponse] = await Promise.all([
        fetch('/api/v1/audit/summary'),
        fetch('/api/v1/audit/metrics')
      ]);

      if (summaryResponse.ok) {
        const summaryData = await summaryResponse.json();
        setSummary(summaryData.data);
      }

      if (metricsResponse.ok) {
        const metricsData = await metricsResponse.json();
        setMetrics(metricsData.data);
      }
    } catch (error) {
      console.error('Failed to load summary and metrics:', error);
    }
  }, []);

  // Verify chain integrity
  const verifyChainIntegrity = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/audit/chain-integrity');
      if (response.ok) {
        const data = await response.json();
        setIntegrityResult(data.data);
      }
    } catch (error) {
      console.error('Failed to verify chain integrity:', error);
    }
  }, []);

  // Load suspicious activities
  const loadSuspiciousActivities = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/audit/suspicious-activities?time_window=24h');
      if (response.ok) {
        const data = await response.json();
        setSuspiciousActivities(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load suspicious activities:', error);
    }
  }, []);

  // Handle export
  const handleExport = useCallback(async (format: ExportFormat, exportFilters: AuditFilter) => {
    try {
      const response = await fetch('/api/v1/audit/export?' + new URLSearchParams({
        format,
        start_time: exportFilters.startDate || '',
        end_time: exportFilters.endDate || '',
        user_id: exportFilters.userId || '',
      }));

      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `audit-logs-${format}-${format === 'pdf' ? '.pdf' : format === 'csv' ? '.csv' : '.json'}`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      }
    } catch (error) {
      console.error('Failed to export audit logs:', error);
    }
  }, []);

  useEffect(() => {
    loadAuditData(filters);
    loadSummaryAndMetrics();
    verifyChainIntegrity();
    loadSuspiciousActivities();
  }, [filters, loadAuditData, loadSummaryAndMetrics, verifyChainIntegrity, loadSuspiciousActivities]);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'high': return 'orange';
      case 'medium': return 'yellow';
      case 'low': return 'green';
      default: return 'gray';
    }
  };

  const getResultColor = (result: string) => {
    switch (result) {
      case 'success': return 'green';
      case 'failure': return 'red';
      case 'error': return 'red';
      default: return 'gray';
    }
  };

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Audit Trail Dashboard</h1>
          <p className="text-gray-600">Monitor and analyze system audit logs with advanced compliance features</p>
        </div>

        {metrics && (
          <div className="flex space-x-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">{metrics.complianceScore.toFixed(1)}%</div>
              <div className="text-sm text-gray-500">Compliance Score</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">{metrics.totalLogsCreated.toLocaleString()}</div>
              <div className="text-sm text-gray-500">Total Logs</div>
            </div>
          </div>
        )}
      </div>

      {/* Navigation Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'logs', label: 'Audit Logs', icon: 'document-text' },
            { id: 'integrity', label: 'Integrity Monitor', icon: 'shield-check' },
            { id: 'tamper-detection', label: 'Tamper Detection', icon: 'exclamation-triangle' },
            { id: 'compliance', label: 'Compliance Reports', icon: 'clipboard-check' },
            { id: 'legal-holds', label: 'Legal Holds', icon: 'scale' },
            { id: 'retention', label: 'Log Retention', icon: 'archive' },
            { id: 'analytics', label: 'Analytics', icon: 'chart-bar' }
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setSelectedView(tab.id as any)}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                selectedView === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center space-x-2">
                <Icon name={tab.icon} className="w-4 h-4" />
                <span>{tab.label}</span>
              </div>
            </button>
          ))}
        </nav>
      </div>

      {/* Alert Banner for Suspicious Activity */}
      {suspiciousActivities.length > 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
          <div className="flex items-start">
            <Icon name="exclamation-triangle" className="w-5 h-5 text-yellow-600 mt-0.5" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-yellow-800">
                Suspicious Activity Detected
              </h3>
              <p className="text-sm text-yellow-700 mt-1">
                {suspiciousActivities.length} suspicious activities detected in the last 24 hours.
                <button className="ml-2 font-medium underline">Review now</button>
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Content based on selected view */}
      {selectedView === 'logs' && (
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Search and Filters */}
          <div className="lg:col-span-1">
            <AuditSearchPanel
              filters={filters}
              onFiltersChange={setFilters}
              onSearch={loadAuditData}
              loading={loading}
            />
            <div className="mt-6">
              <AuditExportManager
                filters={filters}
                onExport={handleExport}
              />
            </div>
          </div>

          {/* Audit Logs Table */}
          <div className="lg:col-span-3">
            <Card>
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-lg font-medium">Recent Audit Logs</h2>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => loadAuditData(filters)}
                    disabled={loading}
                  >
                    <Icon name="refresh" className="w-4 h-4 mr-2" />
                    Refresh
                  </Button>
                </div>

                {loading ? (
                  <div className="flex justify-center py-8">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Timestamp
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            User
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Action
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Resource
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Result
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Severity
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {auditEntries.map((entry) => (
                          <tr key={entry.id} className="hover:bg-gray-50">
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {format(new Date(entry.timestamp), 'MMM dd, HH:mm:ss')}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {entry.userName}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {entry.action}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {entry.resource}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <Badge variant={getResultColor(entry.result)}>
                                {entry.result}
                              </Badge>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <Badge variant={getSeverityColor(entry.severity)}>
                                {entry.severity}
                              </Badge>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </Card>
          </div>
        </div>
      )}

      {selectedView === 'integrity' && (
        <IntegrityMonitor
          integrityResult={integrityResult}
          tamperAlerts={tamperAlerts}
          onVerifyIntegrity={verifyChainIntegrity}
        />
      )}

      {selectedView === 'tamper-detection' && (
        <TamperDetectionMonitor />
      )}

      {selectedView === 'compliance' && (
        <ComplianceReportManager />
      )}

      {selectedView === 'legal-holds' && (
        <LegalHoldManager />
      )}

      {selectedView === 'retention' && (
        <LogRetentionManager />
      )}

      {selectedView === 'analytics' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {summary && (
            <>
              <Card>
                <div className="p-6">
                  <h3 className="text-lg font-medium mb-4">Event Distribution</h3>
                  <div className="space-y-4">
                    <div className="flex justify-between">
                      <span>Success Events</span>
                      <span className="font-medium text-green-600">{summary.successEvents}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Failure Events</span>
                      <span className="font-medium text-red-600">{summary.failureEvents}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Error Events</span>
                      <span className="font-medium text-red-600">{summary.errorEvents}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Critical Events</span>
                      <span className="font-medium text-red-600">{summary.criticalEvents}</span>
                    </div>
                  </div>
                </div>
              </Card>

              <Card>
                <div className="p-6">
                  <h3 className="text-lg font-medium mb-4">Top Actions</h3>
                  <div className="space-y-3">
                    {summary.topActions.map((action, index) => (
                      <div key={index} className="flex justify-between items-center">
                        <span className="text-sm">{action.action}</span>
                        <Badge variant="blue">{action.eventCount}</Badge>
                      </div>
                    ))}
                  </div>
                </div>
              </Card>
            </>
          )}
        </div>
      )}
    </div>
  );
}
import React, { useState, useCallback } from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';
import { AuditTrailSummary } from '@/components/compliance';
import { auditService, AuditExportOptions } from '@/services/auditService';
import { AuditLogEntry, AuditSummary, AuditFilter } from '@kubechat/shared/types';

export default function AuditPage() {
  const [auditEntries, setAuditEntries] = useState<AuditLogEntry[]>([]);
  const [auditSummary, setAuditSummary] = useState<AuditSummary>({
    totalEvents: 0,
    successEvents: 0,
    failureEvents: 0,
    errorEvents: 0,
    criticalEvents: 0,
    topUsers: [],
    topActions: [],
    timeline: []
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadAuditData = useCallback(async (filters: AuditFilter) => {
    setLoading(true);
    setError(null);
    try {
      const [entries, summary] = await Promise.all([
        auditService.getAuditLogs(filters),
        auditService.getAuditSummary(filters)
      ]);

      setAuditEntries(entries);
      setAuditSummary(summary);
    } catch (error) {
      console.error('Failed to load audit data:', error);
      setError('Failed to load audit data');
    } finally {
      setLoading(false);
    }
  }, []);

  const handleExportAuditLog = useCallback(async (filters: AuditFilter) => {
    try {
      const exportOptions: AuditExportOptions = {
        format: 'csv', // Default format
        filters,
        includeMetadata: true,
        startDate: filters.startDate,
        endDate: filters.endDate
      };

      const blob = await auditService.exportAuditLogs(exportOptions);
      const filename = auditService.generateFilename('csv', filters);
      auditService.downloadBlob(blob, filename);
    } catch (error) {
      console.error('Failed to export audit log:', error);
      setError('Failed to export audit data');
    }
  }, []);

  const handleViewEntryDetails = useCallback(async (entryId: string): Promise<AuditLogEntry> => {
    try {
      const entry = await auditService.getAuditLogEntry(entryId);
      if (entry) {
        return entry;
      }

      // Fallback to finding in current entries
      const existingEntry = auditEntries.find(e => e.id === entryId);
      if (existingEntry) {
        return existingEntry;
      }

      // Return default entry if not found
      return {
        id: entryId,
        timestamp: new Date().toISOString(),
        userId: 'unknown',
        userName: 'Unknown User',
        sessionId: '',
        action: 'unknown_action',
        resource: 'unknown_resource',
        method: 'GET',
        result: 'success',
        ipAddress: '',
        userAgent: '',
        metadata: {},
        severity: 'low'
      };
    } catch (error) {
      console.error('Failed to fetch entry details:', error);
      throw error;
    }
  }, [auditEntries]);

  return (
    <>
      <Head>
        <title>Audit Trail - KubeChat</title>
        <meta name="description" content="View audit logs and security events" />
      </Head>

      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Audit Trail</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              Monitor and review system activities, commands, and security events
            </p>
          </div>

          {error && (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800 dark:text-red-200">
                    Unable to load audit data
                  </h3>
                  <div className="mt-2 text-sm text-red-700 dark:text-red-300">
                    <p>{error}</p>
                  </div>
                  <div className="mt-4">
                    <button
                      type="button"
                      onClick={() => loadAuditData({
                        startDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                        endDate: new Date().toISOString()
                      })}
                      className="bg-red-50 dark:bg-red-900/20 px-3 py-2 rounded-md text-sm font-medium text-red-800 dark:text-red-200 hover:bg-red-100 dark:hover:bg-red-900/30"
                    >
                      Retry
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          <AuditTrailSummary
            auditEntries={auditEntries}
            auditSummary={auditSummary}
            onLoadAuditData={loadAuditData}
            onExportAuditLog={handleExportAuditLog}
            onViewEntryDetails={handleViewEntryDetails}
            loading={loading}
          />
        </div>
      </MainLayout>
    </>
  );
}
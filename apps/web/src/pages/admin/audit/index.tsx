// Admin Audit Integration Interface
// Following coding standards from docs/architecture/coding-standards.md
// AC 7: Audit Integration Interface

import React, { useEffect, useState } from 'react';
import { AdminLayout } from '../../../components/admin/AdminLayout';
import { useAdminStore } from '../../../stores/adminStore';
import type { AuditLogEntry } from '../../../types/admin';

export default function AdminAuditPage() {
  const {
    auditLogs,
    isLoading,
    error,
    loadAuditLogs,
    setError
  } = useAdminStore();

  const [currentPage, setCurrentPage] = useState(1);
  const [filters, setFilters] = useState({
    eventType: '',
    userId: '',
    dateRange: {
      start: '',
      end: ''
    },
    riskLevel: '',
    success: ''
  });
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLogs, setSelectedLogs] = useState<string[]>([]);
  const [showFilters, setShowFilters] = useState(false);

  const itemsPerPage = 50;

  useEffect(() => {
    loadAuditLogs();
  }, [loadAuditLogs]);

  const handleFilterChange = (key: string, value: any) => {
    setFilters(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const handleDateRangeChange = (type: 'start' | 'end', value: string) => {
    setFilters(prev => ({
      ...prev,
      dateRange: {
        ...prev.dateRange,
        [type]: value
      }
    }));
  };

  const applyFilters = () => {
    setCurrentPage(1);
    // In a real implementation, these filters would be sent to the backend
    loadAuditLogs();
  };

  const clearFilters = () => {
    setFilters({
      eventType: '',
      userId: '',
      dateRange: { start: '', end: '' },
      riskLevel: '',
      success: ''
    });
    setSearchTerm('');
    setCurrentPage(1);
    loadAuditLogs();
  };

  const filteredLogs = React.useMemo(() => {
    return auditLogs.logs.filter(log => {
      if (searchTerm) {
        const searchLower = searchTerm.toLowerCase();
        return (
          log.action.toLowerCase().includes(searchLower) ||
          log.userId?.toLowerCase().includes(searchLower) ||
          log.ipAddress?.toLowerCase().includes(searchLower) ||
          (log.metadata?.details && log.metadata.details.toLowerCase().includes(searchLower))
        );
      }
      return true;
    });
  }, [auditLogs.logs, searchTerm]);

  const getRiskLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
      case 'high': return 'text-red-800 bg-red-100';
      case 'medium': return 'text-yellow-800 bg-yellow-100';
      case 'low': return 'text-green-800 bg-green-100';
      default: return 'text-gray-800 bg-gray-100';
    }
  };

  const getEventTypeIcon = (action: string) => {
    switch (action.toLowerCase()) {
      case 'login': return '🔑';
      case 'logout': return '🚪';
      case 'user_create': return '👤';
      case 'user_update': return '✏️';
      case 'user_delete': return '🗑️';
      case 'role_change': return '🎭';
      case 'permission_change': return '🛡️';
      case 'password_reset': return '🔄';
      case 'session_terminate': return '⏹️';
      case 'security_violation': return '🚨';
      case 'credential_rotation': return '🔐';
      case 'emergency_access': return '🆘';
      default: return '📝';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return {
      date: date.toLocaleDateString(),
      time: date.toLocaleTimeString()
    };
  };

  const exportAuditLogs = () => {
    const csvContent = [
      ['Timestamp', 'Event Type', 'User ID', 'IP Address', 'Risk Level', 'Success', 'Details'].join(','),
      ...filteredLogs.map(log => [
        log.timestamp,
        log.action,
        log.userId || '',
        log.ipAddress || '',
        log.riskLevel || '',
        log.success ? 'Yes' : 'No',
        `"${(log.metadata?.details || '').replace(/"/g, '""')}"`
      ].join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `audit-logs-${new Date().toISOString().split('T')[0]}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (error) {
    return (
      <AdminLayout title="Audit Logs">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error loading audit logs</h3>
              <p className="text-sm text-red-700 mt-1">{error}</p>
              <button
                onClick={() => {
                  setError(null);
                  loadAuditLogs();
                }}
                className="mt-2 text-sm bg-red-100 text-red-800 px-3 py-1 rounded hover:bg-red-200"
              >
                Try Again
              </button>
            </div>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout title="Audit Logs">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Audit Trail</h1>
            <p className="text-gray-600">Monitor and investigate system activities</p>
          </div>
          <div className="flex space-x-3">
            {filteredLogs.length > 0 && (
              <button
                onClick={exportAuditLogs}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Export CSV
              </button>
            )}
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`inline-flex items-center px-4 py-2 border rounded-md text-sm font-medium ${
                showFilters
                  ? 'border-blue-300 text-blue-700 bg-blue-50'
                  : 'border-gray-300 text-gray-700 bg-white hover:bg-gray-50'
              }`}
            >
              <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.414A1 1 0 013 6.707V4z" />
              </svg>
              Filters
            </button>
            <button
              onClick={() => loadAuditLogs()}
              disabled={isLoading}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
            >
              {isLoading ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600 mr-2"></div>
              ) : (
                <svg className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              )}
              Refresh
            </button>
          </div>
        </div>

        {/* Audit Statistics */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-blue-100 rounded-md flex items-center justify-center">
                    <span className="text-lg">📊</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">Total Events</dt>
                    <dd className="text-lg font-medium text-gray-900">{auditLogs.total}</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-red-100 rounded-md flex items-center justify-center">
                    <span className="text-lg">🚨</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">High Risk</dt>
                    <dd className="text-lg font-medium text-gray-900">
                      {auditLogs.logs.filter(log => log.riskLevel === 'high').length}
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-red-100 rounded-md flex items-center justify-center">
                    <span className="text-lg">❌</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">Failed Events</dt>
                    <dd className="text-lg font-medium text-gray-900">
                      {auditLogs.logs.filter(log => !log.success).length}
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-green-100 rounded-md flex items-center justify-center">
                    <span className="text-lg">👥</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">Unique Users</dt>
                    <dd className="text-lg font-medium text-gray-900">
                      {new Set(auditLogs.logs.map(log => log.userId)).size}
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Filters Panel */}
        {showFilters && (
          <div className="bg-white shadow rounded-lg p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4">
              <div>
                <label htmlFor="event-type" className="block text-sm font-medium text-gray-700 mb-1">
                  Event Type
                </label>
                <select
                  id="event-type"
                  value={filters.eventType}
                  onChange={(e) => handleFilterChange('eventType', e.target.value)}
                  className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Events</option>
                  <option value="login">Login</option>
                  <option value="logout">Logout</option>
                  <option value="user_create">User Create</option>
                  <option value="user_update">User Update</option>
                  <option value="user_delete">User Delete</option>
                  <option value="role_change">Role Change</option>
                  <option value="security_violation">Security Violation</option>
                </select>
              </div>

              <div>
                <label htmlFor="risk-level" className="block text-sm font-medium text-gray-700 mb-1">
                  Risk Level
                </label>
                <select
                  id="risk-level"
                  value={filters.riskLevel}
                  onChange={(e) => handleFilterChange('riskLevel', e.target.value)}
                  className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Levels</option>
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                </select>
              </div>

              <div>
                <label htmlFor="success-filter" className="block text-sm font-medium text-gray-700 mb-1">
                  Status
                </label>
                <select
                  id="success-filter"
                  value={filters.success}
                  onChange={(e) => handleFilterChange('success', e.target.value)}
                  className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Status</option>
                  <option value="true">Success</option>
                  <option value="false">Failed</option>
                </select>
              </div>

              <div>
                <label htmlFor="start-date" className="block text-sm font-medium text-gray-700 mb-1">
                  Start Date
                </label>
                <input
                  type="date"
                  id="start-date"
                  value={filters.dateRange.start}
                  onChange={(e) => handleDateRangeChange('start', e.target.value)}
                  className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              <div>
                <label htmlFor="end-date" className="block text-sm font-medium text-gray-700 mb-1">
                  End Date
                </label>
                <input
                  type="date"
                  id="end-date"
                  value={filters.dateRange.end}
                  onChange={(e) => handleDateRangeChange('end', e.target.value)}
                  className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              <div className="flex items-end space-x-2">
                <button
                  onClick={applyFilters}
                  className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                >
                  Apply
                </button>
                <button
                  onClick={clearFilters}
                  className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  Clear
                </button>
              </div>
            </div>

            <div className="mt-4">
              <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-1">
                Search Events
              </label>
              <input
                type="text"
                id="search"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search by event, user ID, IP address, or details..."
                className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>
        )}

        {/* Audit Logs Table */}
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <div className="px-4 py-5 sm:p-6">
            <div className="mb-4 flex justify-between items-center">
              <div>
                <h3 className="text-lg font-medium text-gray-900">Audit Events</h3>
                <p className="text-sm text-gray-500">
                  Showing {filteredLogs.length} of {auditLogs.total} events
                </p>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Event
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      User
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Timestamp
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      IP Address
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Risk Level
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Details
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {filteredLogs.map((log) => {
                    const timestamp = formatTimestamp(log.timestamp);
                    return (
                      <tr key={log.id} className="hover:bg-gray-50">
                        <td className="px-6 py-4">
                          <div className="flex items-center">
                            <span className="text-lg mr-2">{getEventTypeIcon(log.action)}</span>
                            <div>
                              <div className="text-sm font-medium text-gray-900">
                                {log.action.replace(/_/g, ' ').replace(/\b\w/g, (l: string) => l.toUpperCase())}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-gray-900">{log.userId}</div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-gray-900">{timestamp.date}</div>
                          <div className="text-xs text-gray-500">{timestamp.time}</div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-gray-900 font-mono">
                            {log.ipAddress || 'Unknown'}
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          {log.riskLevel && (
                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getRiskLevelColor(log.riskLevel)}`}>
                              {log.riskLevel.toUpperCase()}
                            </span>
                          )}
                        </td>
                        <td className="px-6 py-4">
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            log.success ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                          }`}>
                            {log.success ? 'Success' : 'Failed'}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-gray-900 max-w-xs truncate" title={log.metadata?.details || log.errorMessage || ''}>
                            {log.metadata?.details || log.errorMessage || 'No details'}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {auditLogs.total > itemsPerPage && (
              <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
                <div className="flex-1 flex justify-between sm:hidden">
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                    className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setCurrentPage(Math.min(Math.ceil(auditLogs.total / itemsPerPage), currentPage + 1))}
                    disabled={currentPage === Math.ceil(auditLogs.total / itemsPerPage)}
                    className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                  >
                    Next
                  </button>
                </div>
                <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                  <div>
                    <p className="text-sm text-gray-700">
                      Showing <span className="font-medium">{(currentPage - 1) * itemsPerPage + 1}</span> to{' '}
                      <span className="font-medium">
                        {Math.min(currentPage * itemsPerPage, auditLogs.total)}
                      </span>{' '}
                      of <span className="font-medium">{auditLogs.total}</span> results
                    </p>
                  </div>
                  <div>
                    <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                      <button
                        onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                        disabled={currentPage === 1}
                        className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                      >
                        Previous
                      </button>
                      <button
                        onClick={() => setCurrentPage(Math.min(Math.ceil(auditLogs.total / itemsPerPage), currentPage + 1))}
                        disabled={currentPage === Math.ceil(auditLogs.total / itemsPerPage)}
                        className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                      >
                        Next
                      </button>
                    </nav>
                  </div>
                </div>
              </div>
            )}

            {filteredLogs.length === 0 && !isLoading && (
              <div className="text-center py-8">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900">No audit logs found</h3>
                <p className="mt-1 text-sm text-gray-500">
                  {searchTerm || Object.values(filters).some(f => f)
                    ? 'Try adjusting your search or filters.'
                    : 'No audit events are currently available.'
                  }
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </AdminLayout>
  );
}
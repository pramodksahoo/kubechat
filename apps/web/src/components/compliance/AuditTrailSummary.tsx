import { useState, useEffect } from 'react';
import { AuditLogEntry, AuditSummary, AuditFilter } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
import { formatDistanceToNow, format } from 'date-fns';

interface AuditTrailSummaryProps {
  auditEntries: AuditLogEntry[];
  auditSummary: AuditSummary;
  onLoadAuditData: (filters: AuditFilter) => Promise<void>;
  onExportAuditLog: (filters: AuditFilter) => Promise<void>;
  onViewEntryDetails: (entryId: string) => Promise<AuditLogEntry>;
  loading?: boolean;
  className?: string;
}

export function AuditTrailSummary({
  auditEntries,
  auditSummary,
  onLoadAuditData,
  onExportAuditLog,
  onViewEntryDetails,
  loading = false,
  className = '',
}: AuditTrailSummaryProps) {
  const [filters, setFilters] = useState<AuditFilter>({
    startDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    endDate: new Date().toISOString(),
  });
  const [showFilters, setShowFilters] = useState(false);
  const [showEntryDetails, setShowEntryDetails] = useState(false);
  const [selectedEntry, setSelectedEntry] = useState<AuditLogEntry | null>(null);
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d' | '30d' | 'custom'>('7d');

  useEffect(() => {
    onLoadAuditData(filters);
  }, [filters, onLoadAuditData]);

  const updateTimeRange = (range: '1h' | '24h' | '7d' | '30d' | 'custom') => {
    setTimeRange(range);
    
    if (range === 'custom') {
      setShowFilters(true);
      return;
    }

    const now = new Date();
    let startDate: Date;

    switch (range) {
      case '1h':
        startDate = new Date(now.getTime() - 60 * 60 * 1000);
        break;
      case '24h':
        startDate = new Date(now.getTime() - 24 * 60 * 60 * 1000);
        break;
      case '7d':
        startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case '30d':
        startDate = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
      default:
        startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    }

    setFilters(prev => ({
      ...prev,
      startDate: startDate.toISOString(),
      endDate: now.toISOString(),
    }));
  };

  const updateFilters = (newFilters: Partial<AuditFilter>) => {
    setFilters(prev => ({ ...prev, ...newFilters }));
  };

  const getResultIcon = (result: string) => {
    switch (result) {
      case 'success': return '✅';
      case 'failure': return '❌';
      case 'error': return '⚠️';
      default: return '📝';
    }
  };

  const getResultColor = (result: string) => {
    switch (result) {
      case 'success': return 'text-green-600 dark:text-green-400';
      case 'failure': return 'text-red-600 dark:text-red-400';
      case 'error': return 'text-yellow-600 dark:text-yellow-400';
      default: return 'text-gray-600 dark:text-gray-400';
    }
  };

  const getSeverityBadge = (severity: string) => {
    const colors = {
      low: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      high: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
      critical: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    };

    return (
      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${colors[severity as keyof typeof colors] || colors.medium}`}>
        {severity.toUpperCase()}
      </span>
    );
  };

  const getMethodBadge = (method: string) => {
    const colors = {
      GET: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
      POST: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      PUT: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      DELETE: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
      PATCH: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
    };

    return (
      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${colors[method as keyof typeof colors] || colors.GET}`}>
        {method}
      </span>
    );
  };

  const handleViewDetails = async (entry: AuditLogEntry) => {
    try {
      const fullEntry = await onViewEntryDetails(entry.id);
      setSelectedEntry(fullEntry);
      setShowEntryDetails(true);
    } catch (error) {
      console.error('Failed to load entry details:', error);
      setSelectedEntry(entry);
      setShowEntryDetails(true);
    }
  };

  const getSuccessRate = () => {
    const total = auditSummary.totalEvents;
    const success = auditSummary.successEvents;
    return total > 0 ? Math.round((success / total) * 100) : 0;
  };

  const getFailureRate = () => {
    const total = auditSummary.totalEvents;
    const failures = auditSummary.failureEvents + auditSummary.errorEvents;
    return total > 0 ? Math.round((failures / total) * 100) : 0;
  };

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
              <span className="text-blue-600 dark:text-blue-400 text-xl">📊</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {auditSummary.totalEvents.toLocaleString()}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Total Events</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
              <span className="text-green-600 dark:text-green-400 text-xl">✅</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {getSuccessRate()}%
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Success Rate</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
              <span className="text-red-600 dark:text-red-400 text-xl">🚨</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {auditSummary.criticalEvents}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Critical Events</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-orange-100 dark:bg-orange-900 rounded-lg">
              <span className="text-orange-600 dark:text-orange-400 text-xl">⚠️</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {getFailureRate()}%
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Failure Rate</div>
            </div>
          </div>
        </Card>
      </div>

      {/* Time Range Selector */}
      <Card>
        <div className="p-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Audit Trail Timeline
            </h3>
            
            <div className="flex items-center space-x-2">
              {['1h', '24h', '7d', '30d', 'custom'].map((range) => (
                <button
                  key={range}
                  onClick={() => updateTimeRange(range as '1h' | '24h' | '7d' | '30d' | 'custom')}
                  className={`px-3 py-1 text-sm font-medium rounded transition-colors ${
                    timeRange === range
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
                  }`}
                >
                  {range === 'custom' ? 'Custom' : range.toUpperCase()}
                </button>
              ))}
            </div>
          </div>

          {/* Activity Timeline Chart Placeholder */}
          <div className="h-32 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
            <div className="text-gray-500 dark:text-gray-400 text-center">
              <div className="text-2xl mb-2">📈</div>
              <div className="text-sm">Activity timeline visualization would be rendered here</div>
            </div>
          </div>
        </div>
      </Card>

      {/* Controls */}
      <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            onClick={() => setShowFilters(!showFilters)}
            variant="secondary"
            size="sm"
          >
            🔍 {showFilters ? 'Hide' : 'Show'} Filters
          </Button>
          
          <Button
            onClick={() => onExportAuditLog(filters)}
            variant="secondary"
            size="sm"
          >
            📥 Export
          </Button>
        </div>

        <div className="flex items-center space-x-2">
          <span className="text-sm text-gray-600 dark:text-gray-400">
            {auditEntries.length} entries shown
          </span>
          
          <Button
            onClick={() => onLoadAuditData(filters)}
            variant="secondary"
            size="sm"
            loading={loading}
          >
            🔄 Refresh
          </Button>
        </div>
      </div>

      {/* Advanced Filters */}
      {showFilters && (
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Advanced Filters
            </h3>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  User ID
                </label>
                <Input
                  value={filters.userId || ''}
                  onChange={(value) => updateFilters({ userId: value || undefined })}
                  placeholder="Filter by user ID"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Action
                </label>
                <Input
                  value={filters.action || ''}
                  onChange={(value) => updateFilters({ action: value || undefined })}
                  placeholder="Filter by action"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Resource
                </label>
                <Input
                  value={filters.resource || ''}
                  onChange={(value) => updateFilters({ resource: value || undefined })}
                  placeholder="Filter by resource"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Result
                </label>
                <select
                  value={filters.result || ''}
                  onChange={(e) => updateFilters({ result: e.target.value as 'success' | 'failure' | 'error' || undefined })}
                  className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                >
                  <option value="">All Results</option>
                  <option value="success">Success</option>
                  <option value="failure">Failure</option>
                  <option value="error">Error</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Severity
                </label>
                <select
                  value={filters.severity || ''}
                  onChange={(e) => updateFilters({ severity: e.target.value as 'low' | 'medium' | 'high' | 'critical' || undefined })}
                  className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                >
                  <option value="">All Severities</option>
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                  <option value="critical">Critical</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Search Query
                </label>
                <Input
                  value={filters.searchQuery || ''}
                  onChange={(value) => updateFilters({ searchQuery: value || undefined })}
                  placeholder="Search in commands, endpoints..."
                />
              </div>
            </div>

            <div className="flex items-center justify-between mt-4">
              <Button
                onClick={() => setFilters({
                  startDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                  endDate: new Date().toISOString(),
                })}
                variant="secondary"
                size="sm"
              >
                Clear Filters
              </Button>
              
              <div className="text-sm text-gray-600 dark:text-gray-400">
                Filters applied: {Object.values(filters).filter(v => v !== undefined && v !== '').length}
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* Top Activity Summary */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Top Users
            </h3>
            <div className="space-y-3">
              {auditSummary.topUsers.slice(0, 5).map((user, index) => (
                <div key={user.userId} className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      {index + 1}.
                    </span>
                    <span className="text-sm text-gray-900 dark:text-white">
                      {user.userName || user.userId}
                    </span>
                  </div>
                  <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                    {user.eventCount}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </Card>

        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Top Actions
            </h3>
            <div className="space-y-3">
              {auditSummary.topActions.slice(0, 5).map((action, index) => (
                <div key={action.action} className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      {index + 1}.
                    </span>
                    <span className="text-sm text-gray-900 dark:text-white">
                      {action.action}
                    </span>
                  </div>
                  <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                    {action.eventCount}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </Card>

        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Activity Trend
            </h3>
            <div className="space-y-3">
              {auditSummary.timeline.slice(-5).map((day) => (
                <div key={day.date} className="flex items-center justify-between">
                  <span className="text-sm text-gray-900 dark:text-white">
                    {format(new Date(day.date), 'MMM dd')}
                  </span>
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                      {day.eventCount}
                    </span>
                    <span className={`text-xs font-medium ${
                      day.successRate >= 95 ? 'text-green-600 dark:text-green-400' :
                      day.successRate >= 90 ? 'text-yellow-600 dark:text-yellow-400' :
                      'text-red-600 dark:text-red-400'
                    }`}>
                      {Math.round(day.successRate)}%
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </div>

      {/* Audit Entries List */}
      <Card>
        <div className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Recent Audit Entries
          </h3>

          <div className="space-y-3">
            {auditEntries.slice(0, 20).map((entry) => (
              <div
                key={entry.id}
                className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer"
                onClick={() => handleViewDetails(entry)}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center space-x-3">
                    <span className="text-lg">{getResultIcon(entry.result)}</span>
                    <div>
                      <div className="flex items-center space-x-2">
                        <span className="font-medium text-gray-900 dark:text-white">
                          {entry.action}
                        </span>
                        {getMethodBadge(entry.method)}
                        {getSeverityBadge(entry.severity)}
                      </div>
                      <div className="text-sm text-gray-600 dark:text-gray-400">
                        {entry.userName || entry.userId} • {entry.resource}
                        {entry.clusterName && ` • ${entry.clusterName}`}
                      </div>
                    </div>
                  </div>
                  
                  <div className="text-right">
                    <div className={`text-sm font-medium ${getResultColor(entry.result)}`}>
                      {entry.result.toUpperCase()}
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      {formatDistanceToNow(new Date(entry.timestamp), { addSuffix: true })}
                    </div>
                    {entry.duration && (
                      <div className="text-xs text-gray-500 dark:text-gray-400">
                        {entry.duration}ms
                      </div>
                    )}
                  </div>
                </div>

                {entry.command && (
                  <div className="bg-gray-50 dark:bg-gray-800 p-2 rounded text-sm font-mono">
                    {entry.command}
                  </div>
                )}

                {entry.errorMessage && (
                  <div className="bg-red-50 dark:bg-red-900/20 p-2 rounded text-sm text-red-700 dark:text-red-300 mt-2">
                    {entry.errorMessage}
                  </div>
                )}
              </div>
            ))}
          </div>

          {auditEntries.length === 0 && (
            <div className="text-center py-8">
              <div className="text-gray-500 dark:text-gray-400">
                <span className="text-4xl mb-4 block">🔍</span>
                <p className="text-lg font-medium mb-2">No audit entries found</p>
                <p className="text-sm">Try adjusting your filters or time range</p>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Entry Details Modal */}
      <Modal
        isOpen={showEntryDetails}
        onClose={() => setShowEntryDetails(false)}
        title="Audit Entry Details"
      >
        {selectedEntry && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Timestamp
                </label>
                <p className="text-gray-900 dark:text-white">
                  {format(new Date(selectedEntry.timestamp), 'MMM dd, yyyy HH:mm:ss')}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Result
                </label>
                <div className="flex items-center space-x-2">
                  <span>{getResultIcon(selectedEntry.result)}</span>
                  <span className={`font-medium ${getResultColor(selectedEntry.result)}`}>
                    {selectedEntry.result.toUpperCase()}
                  </span>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  User
                </label>
                <p className="text-gray-900 dark:text-white">
                  {selectedEntry.userName || selectedEntry.userId}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Session ID
                </label>
                <p className="text-gray-900 dark:text-white font-mono text-sm">
                  {selectedEntry.sessionId}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Method
                </label>
                {getMethodBadge(selectedEntry.method)}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Severity
                </label>
                {getSeverityBadge(selectedEntry.severity)}
              </div>

              {selectedEntry.clusterId && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Cluster
                  </label>
                  <p className="text-gray-900 dark:text-white">
                    {selectedEntry.clusterName || selectedEntry.clusterId}
                  </p>
                </div>
              )}

              {selectedEntry.duration && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Duration
                  </label>
                  <p className="text-gray-900 dark:text-white">
                    {selectedEntry.duration}ms
                  </p>
                </div>
              )}

              {selectedEntry.statusCode && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Status Code
                  </label>
                  <p className="text-gray-900 dark:text-white">
                    {selectedEntry.statusCode}
                  </p>
                </div>
              )}

              {selectedEntry.ipAddress && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    IP Address
                  </label>
                  <p className="text-gray-900 dark:text-white">
                    {selectedEntry.ipAddress}
                  </p>
                </div>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Action
              </label>
              <p className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded">
                {selectedEntry.action}
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Resource
              </label>
              <p className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded">
                {selectedEntry.resource}
                {selectedEntry.resourceId && ` (ID: ${selectedEntry.resourceId})`}
              </p>
            </div>

            {selectedEntry.endpoint && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Endpoint
                </label>
                <p className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded font-mono text-sm">
                  {selectedEntry.endpoint}
                </p>
              </div>
            )}

            {selectedEntry.command && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Command
                </label>
                <pre className="text-gray-900 dark:text-white bg-gray-900 text-green-400 p-3 rounded font-mono text-sm overflow-auto">
                  {selectedEntry.command}
                </pre>
              </div>
            )}

            {selectedEntry.errorMessage && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Error Message
                </label>
                <p className="text-red-700 dark:text-red-300 bg-red-50 dark:bg-red-900/20 p-3 rounded">
                  {selectedEntry.errorMessage}
                </p>
              </div>
            )}

            {selectedEntry.userAgent && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  User Agent
                </label>
                <p className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded font-mono text-sm break-all">
                  {selectedEntry.userAgent}
                </p>
              </div>
            )}

            {selectedEntry.metadata && Object.keys(selectedEntry.metadata).length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Additional Metadata
                </label>
                <pre className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded font-mono text-sm overflow-auto">
                  {JSON.stringify(selectedEntry.metadata, null, 2)}
                </pre>
              </div>
            )}

            <div className="flex space-x-3">
              <Button
                onClick={() => setShowEntryDetails(false)}
                variant="primary"
                className="flex-1"
              >
                Close
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
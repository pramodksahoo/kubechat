import { useState, useEffect } from 'react';
interface SecurityEvent {
  id: string;
  userId: string;
  type: string;
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  timestamp: Date;
  ipAddress: string;
  sessionId: string;
  userAgent: string;
  metadata?: Record<string, unknown>;
}
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
// import { formatDistanceToNow, format } from 'date-fns';

interface SecurityEventsMonitorProps {
  events: SecurityEvent[];
  onLoadEvents: (filters: EventFilters) => Promise<void>;
  onAcknowledgeEvent: (eventId: string) => Promise<void>;
  onAcknowledgeMultiple: (eventIds: string[]) => Promise<void>;
  onExportEvents: (filters: EventFilters) => Promise<void>;
  onCreateAlert: (eventType: string, threshold: number) => Promise<void>;
  loading?: boolean;
  className?: string;
}

interface EventFilters {
  type?: string;
  severity?: string;
  dateFrom?: string;
  dateTo?: string;
  userId?: string;
  ipAddress?: string;
  limit?: number;
  offset?: number;
}

export function SecurityEventsMonitor({
  events,
  onLoadEvents,
  onAcknowledgeEvent,
  onAcknowledgeMultiple,
  onExportEvents,
  onCreateAlert,
  loading = false,
  className = '',
}: SecurityEventsMonitorProps) {
  const [filters, setFilters] = useState<EventFilters>({
    limit: 50,
    offset: 0,
  });
  const [selectedEvents, setSelectedEvents] = useState<Set<string>>(new Set());
  const [showFilters, setShowFilters] = useState(false);
  const [showEventDetails, setShowEventDetails] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<SecurityEvent | null>(null);
  const [showCreateAlert, setShowCreateAlert] = useState(false);
  const [alertConfig, setAlertConfig] = useState({
    eventType: '',
    threshold: 5,
  });

  useEffect(() => {
    onLoadEvents(filters);
  }, [filters, onLoadEvents]);

  const updateFilters = (newFilters: Partial<EventFilters>) => {
    setFilters(prev => ({ ...prev, ...newFilters, offset: 0 }));
  };

  const getEventIcon = (type: string) => {
    switch (type) {
      case 'login': return '✅';
      case 'logout': return '👋';
      case 'failed_login': return '❌';
      case 'permission_denied': return '🚫';
      case 'suspicious_activity': return '⚠️';
      default: return '📝';
    }
  };

  // const getSeverityColor = (severity: string) => {
  //   switch (severity) {
  //     case 'low': return 'text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20';
  //     case 'medium': return 'text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/20';
  //     case 'high': return 'text-orange-600 dark:text-orange-400 bg-orange-50 dark:bg-orange-900/20';
  //     case 'critical': return 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20';
  //     default: return 'text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800';
  //   }
  // };

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

  const handleSelectEvent = (eventId: string) => {
    const newSelected = new Set(selectedEvents);
    if (newSelected.has(eventId)) {
      newSelected.delete(eventId);
    } else {
      newSelected.add(eventId);
    }
    setSelectedEvents(newSelected);
  };

  const handleSelectAll = () => {
    if (selectedEvents.size === events.length) {
      setSelectedEvents(new Set());
    } else {
      setSelectedEvents(new Set(events.map(e => e.id)));
    }
  };

  const handleBulkAcknowledge = async () => {
    if (selectedEvents.size > 0) {
      await onAcknowledgeMultiple(Array.from(selectedEvents));
      setSelectedEvents(new Set());
    }
  };

  const handleCreateAlert = async () => {
    await onCreateAlert(alertConfig.eventType, alertConfig.threshold);
    setShowCreateAlert(false);
    setAlertConfig({ eventType: '', threshold: 5 });
  };

  // Commented out unused function
  // const getEventTypeFrequency = () => {
  //   const frequency: Record<string, number> = {};
  //   events.forEach(event => {
  //     frequency[event.type] = (frequency[event.type] || 0) + 1;
  //   });
  //   return frequency;
  // };

  const getSeverityBreakdown = () => {
    const breakdown: Record<string, number> = {};
    events.forEach(event => {
      breakdown[event.severity] = (breakdown[event.severity] || 0) + 1;
    });
    return breakdown;
  };

  // const eventTypeFrequency = getEventTypeFrequency();
  const severityBreakdown = getSeverityBreakdown();

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header with Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">
              {events.length}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Total Events</div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-red-600 dark:text-red-400">
              {severityBreakdown.critical || 0}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Critical</div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">
              {severityBreakdown.high || 0}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">High Priority</div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600 dark:text-green-400">
              {events.filter(e => new Date(e.timestamp).getTime() > Date.now() - 24 * 60 * 60 * 1000).length}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Last 24h</div>
          </div>
        </Card>
      </div>

      {/* Controls */}
      <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            onClick={() => setShowFilters(!showFilters)}
            variant="secondary"
            size="sm"
          >
            🔍 Filters
          </Button>
          
          <Button
            onClick={() => onExportEvents(filters)}
            variant="secondary"
            size="sm"
          >
            📥 Export
          </Button>
          
          <Button
            onClick={() => setShowCreateAlert(true)}
            variant="secondary"
            size="sm"
          >
            🔔 Create Alert
          </Button>
        </div>

        <div className="flex items-center space-x-4">
          {selectedEvents.size > 0 && (
            <div className="flex items-center space-x-2">
              <span className="text-sm text-gray-600 dark:text-gray-400">
                {selectedEvents.size} selected
              </span>
              <Button
                onClick={handleBulkAcknowledge}
                variant="primary"
                size="sm"
              >
                Acknowledge Selected
              </Button>
            </div>
          )}
          
          <Button
            onClick={() => onLoadEvents(filters)}
            variant="secondary"
            size="sm"
            loading={loading}
          >
            🔄 Refresh
          </Button>
        </div>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Event Filters
            </h3>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Event Type
                </label>
                <select
                  value={filters.type || ''}
                  onChange={(e) => updateFilters({ type: e.target.value || undefined })}
                  className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                >
                  <option value="">All Types</option>
                  <option value="login">Login</option>
                  <option value="logout">Logout</option>
                  <option value="failed_login">Failed Login</option>
                  <option value="permission_denied">Permission Denied</option>
                  <option value="suspicious_activity">Suspicious Activity</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Severity
                </label>
                <select
                  value={filters.severity || ''}
                  onChange={(e) => updateFilters({ severity: e.target.value || undefined })}
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
                  IP Address
                </label>
                <Input
                  value={filters.ipAddress || ''}
                  onChange={(value) => updateFilters({ ipAddress: value || undefined })}
                  placeholder="Filter by IP address"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Date From
                </label>
                <input
                  type="datetime-local"
                  value={filters.dateFrom || ''}
                  onChange={(e) => updateFilters({ dateFrom: e.target.value || undefined })}
                  className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Date To
                </label>
                <input
                  type="datetime-local"
                  value={filters.dateTo || ''}
                  onChange={(e) => updateFilters({ dateTo: e.target.value || undefined })}
                  className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                />
              </div>

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
            </div>

            <div className="flex items-center justify-between mt-4">
              <Button
                onClick={() => setFilters({ limit: 50, offset: 0 })}
                variant="secondary"
                size="sm"
              >
                Clear Filters
              </Button>
              
              <div className="text-sm text-gray-600 dark:text-gray-400">
                Showing {events.length} events
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* Events List */}
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Security Events
            </h3>
            
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={selectedEvents.size === events.length && events.length > 0}
                onChange={handleSelectAll}
                className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <label className="text-sm text-gray-600 dark:text-gray-400">
                Select All
              </label>
            </div>
          </div>

          <div className="space-y-3">
            {events.map((event) => (
              <div
                key={event.id}
                className={`border rounded-lg p-4 transition-colors ${'bg-gray-50 dark:bg-gray-800'} ${
                  selectedEvents.has(event.id) ? 'ring-2 ring-blue-500' : ''
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <input
                      type="checkbox"
                      checked={selectedEvents.has(event.id)}
                      onChange={() => handleSelectEvent(event.id)}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    
                    <span className="text-xl">{getEventIcon(event.type)}</span>
                    
                    <div>
                      <div className="flex items-center space-x-2">
                        <h4 className="font-medium text-gray-900 dark:text-white">
                          {event.description}
                        </h4>
                        {getSeverityBadge(event.severity)}
                      </div>
                      
                      <div className="text-sm text-gray-600 dark:text-gray-400 mt-1 space-x-4">
                        <span>{new Date(event.timestamp).toLocaleString()}</span>
                        {event.userId && <span>User: {event.userId}</span>}
                        {event.ipAddress && <span>IP: {event.ipAddress}</span>}
                        {event.sessionId && <span>Session: {event.sessionId.slice(0, 8)}...</span>}
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-2">
                    <Button
                      onClick={() => {
                        setSelectedEvent(event);
                        setShowEventDetails(true);
                      }}
                      variant="secondary"
                      size="sm"
                    >
                      Details
                    </Button>
                    
                    <Button
                      onClick={() => onAcknowledgeEvent(event.id)}
                      variant="primary"
                      size="sm"
                    >
                      Acknowledge
                    </Button>
                  </div>
                </div>

                {/* Quick metadata preview */}
                {event.userAgent && (
                  <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                    User Agent: {event.userAgent.slice(0, 80)}...
                  </div>
                )}
              </div>
            ))}
          </div>

          {events.length === 0 && (
            <div className="text-center py-8">
              <div className="text-gray-500 dark:text-gray-400">
                <span className="text-4xl mb-4 block">🔍</span>
                <p className="text-lg font-medium mb-2">No events found</p>
                <p className="text-sm">Try adjusting your filters or check back later</p>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Event Details Modal */}
      <Modal
        isOpen={showEventDetails}
        onClose={() => setShowEventDetails(false)}
        title="Event Details"
      >
        {selectedEvent && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Event Type
                </label>
                <div className="flex items-center space-x-2">
                  <span>{getEventIcon(selectedEvent.type)}</span>
                  <span className="text-gray-900 dark:text-white">{selectedEvent.type}</span>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Severity
                </label>
                {getSeverityBadge(selectedEvent.severity)}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Timestamp
                </label>
                <p className="text-gray-900 dark:text-white">
                  {new Date(selectedEvent.timestamp).toLocaleString()}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Event ID
                </label>
                <p className="text-gray-900 dark:text-white font-mono text-sm">
                  {selectedEvent.id}
                </p>
              </div>

              {selectedEvent.userId && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    User ID
                  </label>
                  <p className="text-gray-900 dark:text-white">{selectedEvent.userId}</p>
                </div>
              )}

              {selectedEvent.sessionId && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Session ID
                  </label>
                  <p className="text-gray-900 dark:text-white font-mono text-sm">
                    {selectedEvent.sessionId}
                  </p>
                </div>
              )}

              {selectedEvent.ipAddress && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    IP Address
                  </label>
                  <p className="text-gray-900 dark:text-white">{selectedEvent.ipAddress}</p>
                </div>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Description
              </label>
              <p className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded">
                {selectedEvent.description}
              </p>
            </div>

            {selectedEvent.userAgent && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  User Agent
                </label>
                <p className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded font-mono text-sm">
                  {selectedEvent.userAgent}
                </p>
              </div>
            )}

            {selectedEvent.metadata && Object.keys(selectedEvent.metadata).length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Additional Metadata
                </label>
                <pre className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded font-mono text-sm overflow-auto">
                  {JSON.stringify(selectedEvent.metadata, null, 2)}
                </pre>
              </div>
            )}

            <div className="flex space-x-3">
              <Button
                onClick={() => {
                  onAcknowledgeEvent(selectedEvent.id);
                  setShowEventDetails(false);
                }}
                variant="primary"
                className="flex-1"
              >
                Acknowledge Event
              </Button>
              <Button
                onClick={() => setShowEventDetails(false)}
                variant="secondary"
                className="flex-1"
              >
                Close
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Create Alert Modal */}
      <Modal
        isOpen={showCreateAlert}
        onClose={() => setShowCreateAlert(false)}
        title="Create Security Alert"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Event Type
            </label>
            <select
              value={alertConfig.eventType}
              onChange={(e) => setAlertConfig(prev => ({ ...prev, eventType: e.target.value }))}
              className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
            >
              <option value="">Select event type</option>
              <option value="failed_login">Failed Login</option>
              <option value="permission_denied">Permission Denied</option>
              <option value="suspicious_activity">Suspicious Activity</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Alert Threshold (events per hour)
            </label>
            <input
              type="number"
              value={alertConfig.threshold}
              onChange={(e) => setAlertConfig(prev => ({ ...prev, threshold: parseInt(e.target.value) }))}
              min="1"
              max="100"
              className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
            />
          </div>

          <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
            <p className="text-sm text-blue-700 dark:text-blue-300">
              You&apos;ll receive notifications when the selected event type occurs more than {alertConfig.threshold} times within an hour.
            </p>
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleCreateAlert}
              variant="primary"
              className="flex-1"
              disabled={!alertConfig.eventType || alertConfig.threshold < 1}
            >
              Create Alert
            </Button>
            <Button
              onClick={() => setShowCreateAlert(false)}
              variant="secondary"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
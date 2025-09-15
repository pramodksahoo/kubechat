import { useState, useEffect } from 'react';
import { User } from '../../types/user';
import { securityService, SecurityAlert, SecurityEvent, SecurityMetrics, SecurityScan, ComplianceResult } from '../../services/securityService';
import { authService } from '../../services/authService';
import { useRealTimeUpdates, useSystemNotifications } from '../../services/realTimeService';


import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';
// import { formatDistanceToNow } from 'date-fns';

interface SecurityDashboardProps {
  user: User;
  className?: string;
}

export function SecurityDashboard({
  user,
  className = '',
}: SecurityDashboardProps) {
  const [activeTab, setActiveTab] = useState<'overview' | 'sessions' | 'events' | 'alerts' | 'compliance' | 'scans' | 'settings'>('overview');
  const [showTerminateModal, setShowTerminateModal] = useState(false);
  const [sessionToTerminate, setSessionToTerminate] = useState<any>(null);
  const [showTerminateAllModal, setShowTerminateAllModal] = useState(false);
  const [eventFilter, setEventFilter] = useState<'all' | 'high' | 'critical'>('all');
  const [alertFilter, setAlertFilter] = useState<'all' | 'high' | 'critical'>('all');

  // State for real backend data
  const [sessions, setSessions] = useState<any[]>([]);
  const [securityEvents, setSecurityEvents] = useState<SecurityEvent[]>([]);
  const [securityAlerts, setSecurityAlerts] = useState<SecurityAlert[]>([]);
  const [securityMetrics, setSecurityMetrics] = useState<SecurityMetrics | null>(null);
  const [complianceResults, setComplianceResults] = useState<ComplianceResult[]>([]);
  const [securityScans, setSecurityScans] = useState<SecurityScan[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Real-time updates
  const { lastUpdate, isConnected } = useRealTimeUpdates(['security', 'audit', 'system']);
  const { notifications } = useSystemNotifications();

  useEffect(() => {
    loadSecurityData();
  }, []);

  // Handle real-time updates
  useEffect(() => {
    if (lastUpdate) {
      switch (lastUpdate.type) {
        case 'security':
          if (lastUpdate.action === 'create' || lastUpdate.action === 'update') {
            if (lastUpdate.data.alert) {
              setSecurityAlerts(prev => [lastUpdate.data.alert, ...prev.slice(0, 49)]);
            }
            if (lastUpdate.data.event) {
              setSecurityEvents(prev => [lastUpdate.data.event, ...prev.slice(0, 49)]);
            }
            if (lastUpdate.data.metrics) {
              setSecurityMetrics(lastUpdate.data.metrics);
            }
          }
          break;
        case 'audit':
          if (lastUpdate.action === 'alert' && lastUpdate.severity === 'critical') {
            // Auto-refresh security data on critical audit events
            loadSecurityData();
          }
          break;
      }
    }
  }, [lastUpdate]);

  const loadSecurityData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [sessionsData, eventsData, alertsData, metricsData, complianceData, scansData] = await Promise.all([
        authService.getSessions().catch(() => []),
        securityService.getEvents({ limit: 50 }).catch(() => []),
        securityService.getAlerts({ limit: 50 }).catch(() => []),
        securityService.getMetrics('24h').catch(() => null),
        securityService.getComplianceResults().catch(() => []),
        securityService.getScans({ limit: 20 }).catch(() => [])
      ]);

      setSessions(sessionsData);
      setSecurityEvents(eventsData);
      setSecurityAlerts(alertsData);
      setSecurityMetrics(metricsData);
      setComplianceResults(complianceData);
      setSecurityScans(scansData);
    } catch (error) {
      console.error('Failed to load security data:', error);
      setError('Failed to load security data');
    } finally {
      setLoading(false);
    }
  };

  const onRefreshSessions = async () => {
    try {
      const sessionsData = await authService.getSessions();
      setSessions(sessionsData);
    } catch (error) {
      console.error('Failed to refresh sessions:', error);
    }
  };

  const onTerminateSession = async (sessionId: string) => {
    try {
      await authService.terminateSession(sessionId);
      await onRefreshSessions();
    } catch (error) {
      console.error('Failed to terminate session:', error);
      throw error;
    }
  };

  const onTerminateAllSessions = async () => {
    try {
      await authService.terminateAllSessions();
      await onRefreshSessions();
    } catch (error) {
      console.error('Failed to terminate all sessions:', error);
      throw error;
    }
  };

  const onAcknowledgeEvent = async (eventId: string) => {
    // Security events don't have acknowledge in the current implementation
    // This would typically update the event status
    console.log('Acknowledge event:', eventId);
  };

  const onAcknowledgeAlert = async (alertId: string) => {
    try {
      await securityService.acknowledgeAlert(alertId);
      await loadSecurityData(); // Refresh data
    } catch (error) {
      console.error('Failed to acknowledge alert:', error);
    }
  };

  const onResolveAlert = async (alertId: string, resolution: string) => {
    try {
      await securityService.resolveAlert(alertId, resolution);
      await loadSecurityData(); // Refresh data
    } catch (error) {
      console.error('Failed to resolve alert:', error);
    }
  };

  // const getCurrentSession = () => {
  //   // In a real implementation, this would be based on the current session token
  //   return sessions.find(session => session.isActive) || sessions[0];
  // };

  const getActiveSessionsCount = () => sessions.filter(session => session.isActive).length;
  const getSuspiciousEventsCount = () => securityEvents.filter(event =>
    event.type === 'authentication' && event.result === 'failure'
  ).length + securityAlerts.filter(alert =>
    alert.severity === 'high' || alert.severity === 'critical'
  ).length;
  const getRecentLoginAttempts = () => securityEvents.filter(event =>
    event.type === 'authentication' && event.result === 'failure' &&
    new Date(event.timestamp).getTime() > Date.now() - 24 * 60 * 60 * 1000
  ).length;
  const getActiveAlertsCount = () => securityAlerts.filter(alert => alert.status === 'active').length;

  const filteredEvents = securityEvents.filter(event => {
    if (eventFilter === 'all') return true;
    // Map security event result to severity for filtering
    if (eventFilter === 'critical' && event.result === 'blocked') return true;
    if (eventFilter === 'high' && event.result === 'failure') return true;
    return false;
  });

  const filteredAlerts = securityAlerts.filter(alert => {
    if (alertFilter === 'all') return true;
    return alert.severity === alertFilter;
  });

  const handleTerminateSession = async (session: any) => {
    setSessionToTerminate(session);
    setShowTerminateModal(true);
  };

  const confirmTerminateSession = async () => {
    if (sessionToTerminate) {
      try {
        await onTerminateSession(sessionToTerminate.id);
        setShowTerminateModal(false);
        setSessionToTerminate(null);
      } catch (error) {
        console.error('Failed to terminate session:', error);
      }
    }
  };

  const confirmTerminateAllSessions = async () => {
    try {
      await onTerminateAllSessions();
      setShowTerminateAllModal(false);
    } catch (error) {
      console.error('Failed to terminate all sessions:', error);
    }
  };

  const getSessionDevice = (userAgent?: string) => {
    if (!userAgent) return 'Unknown Device';
    
    if (userAgent.includes('Mobile')) return '📱 Mobile';
    if (userAgent.includes('Chrome')) return '💻 Chrome';
    if (userAgent.includes('Firefox')) return '🦊 Firefox';
    if (userAgent.includes('Safari')) return '🌐 Safari';
    if (userAgent.includes('Edge')) return '📘 Edge';
    
    return '💻 Desktop';
  };

  const getSessionLocation = (ipAddress?: string) => {
    // In a real implementation, this would use a geo-IP service
    if (!ipAddress) return 'Unknown Location';
    if (ipAddress === '127.0.0.1' || ipAddress === 'localhost') return '🏠 Local';
    return '🌍 External';
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
  //     case 'low': return 'text-green-600 dark:text-green-400';
  //     case 'medium': return 'text-yellow-600 dark:text-yellow-400';
  //     case 'high': return 'text-orange-600 dark:text-orange-400';
  //     case 'critical': return 'text-red-600 dark:text-red-400';
  //     default: return 'text-gray-600 dark:text-gray-400';
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

  // const currentSession = getCurrentSession();
  const currentSession = sessions.find(s => s.isActive);

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Real-time Connection Status */}
      {notifications.length > 0 && (
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <span className="text-amber-500 text-xl">🔔</span>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-amber-800 dark:text-amber-200">
                {notifications.length} Security Notification{notifications.length !== 1 ? 's' : ''}
              </h3>
              <div className="mt-1 text-sm text-amber-700 dark:text-amber-300">
                {notifications[0]?.title}: {notifications[0]?.message}
                {notifications.length > 1 && ` and ${notifications.length - 1} more`}
              </div>
            </div>
            <div className={`ml-auto flex items-center space-x-2 text-xs ${
              isConnected ? 'text-emerald-600' : 'text-amber-600'
            }`}>
              <div className={`h-2 w-2 rounded-full ${
                isConnected ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500 animate-bounce'
              }`} />
              <span>{isConnected ? 'Live updates' : 'Reconnecting...'}</span>
            </div>
          </div>
        </div>
      )}

      {/* Security Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
              <span className="text-green-600 dark:text-green-400 text-xl">🛡️</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {getActiveSessionsCount()}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Active Sessions</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
              <span className="text-red-600 dark:text-red-400 text-xl">⚠️</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {getActiveAlertsCount()}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Active Alerts</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
              <span className="text-yellow-600 dark:text-yellow-400 text-xl">🔒</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {getRecentLoginAttempts()}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Failed Logins (24h)</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
              <span className="text-blue-600 dark:text-blue-400 text-xl">👤</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {user.lastLoginAt ? new Date(user.lastLoginAt).toLocaleDateString() : 'Never'}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Last Login</div>
            </div>
          </div>
        </Card>
      </div>

      {/* Navigation Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'overview', label: 'Overview', icon: '📊' },
            { id: 'sessions', label: 'Sessions', icon: '🔐', count: getActiveSessionsCount() },
            { id: 'alerts', label: 'Security Alerts', icon: '🚨', count: getActiveAlertsCount() },
            { id: 'events', label: 'Events', icon: '📋', count: getSuspiciousEventsCount() },
            { id: 'compliance', label: 'Compliance', icon: '📜' },
            { id: 'scans', label: 'Scans', icon: '🔍' },
            { id: 'settings', label: 'Settings', icon: '⚙️' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as 'overview' | 'sessions' | 'events' | 'alerts' | 'compliance' | 'scans' | 'settings')}
              className={`py-2 px-1 border-b-2 font-medium text-sm flex items-center space-x-2 ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
              }`}
            >
              <span>{tab.icon}</span>
              <span>{tab.label}</span>
              {'count' in tab && tab.count! > 0 && (
                <span className="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200 text-xs px-2 py-0.5 rounded-full">
                  {tab.count!}
                </span>
              )}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="space-y-6">
        {/* Overview Tab */}
        {activeTab === 'overview' && (
          <div className="space-y-6">
            {/* User Permissions Summary */}
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  User Permissions & Access
                </h3>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                      {user.roles.length}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Roles</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                      {user.permissions.length}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Permissions</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">
                      {user.clusters.length}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Clusters</div>
                  </div>
                  <div className="text-center">
                    <div className={`text-2xl font-bold ${user.isActive ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                      {user.isActive ? 'Active' : 'Inactive'}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Status</div>
                  </div>
                </div>
                <div className="space-y-3">
                  <div>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">Current Roles: </span>
                    <div className="mt-1 flex flex-wrap gap-2">
                      {user.roles.map((role: any, index: number) => (
                        <span
                          key={index}
                          className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                        >
                          {role.name || role}
                        </span>
                      ))}
                      {user.roles.length === 0 && (
                        <span className="text-sm text-gray-500 dark:text-gray-400">No roles assigned</span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </Card>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Current Session
                </h3>
                {currentSession && (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600 dark:text-gray-400">Device</span>
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {getSessionDevice(currentSession.userAgent)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600 dark:text-gray-400">Location</span>
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {getSessionLocation(currentSession.ipAddress)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600 dark:text-gray-400">Started</span>
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {new Date().toLocaleDateString()}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600 dark:text-gray-400">Expires</span>
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {new Date().toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Recent Security Events
                </h3>
                <div className="space-y-3">
                  {(securityAlerts.length > 0 ? securityAlerts : securityEvents).slice(0, 5).map((item) => {
                    const isAlert = 'severity' in item && 'title' in item;
                    return (
                      <div key={item.id} className="flex items-center space-x-3">
                        <span className="text-lg">{isAlert ? '🚨' : getEventIcon((item as SecurityEvent).type)}</span>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                            {isAlert ? (item as SecurityAlert).title : (item as SecurityEvent).action}
                          </p>
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            {new Date(item.timestamp).toLocaleDateString()}
                          </p>
                        </div>
                        {isAlert ? getSeverityBadge((item as SecurityAlert).severity) :
                         <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                           (item as SecurityEvent).result === 'success' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                           (item as SecurityEvent).result === 'failure' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' :
                           'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
                         }`}>
                           {((item as SecurityEvent).result || 'unknown').toUpperCase()}
                         </span>}
                      </div>
                    );
                  })}
                </div>
              </div>
            </Card>
            </div>
          </div>
        )}

        {/* Sessions Tab */}
        {activeTab === 'sessions' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Active Sessions ({getActiveSessionsCount()})
              </h3>
              <div className="flex space-x-3">
                <Button
                  onClick={onRefreshSessions}
                  variant="secondary"
                  size="sm"
                  loading={loading}
                >
                  Refresh
                </Button>
                <Button
                  onClick={() => setShowTerminateAllModal(true)}
                  variant="danger"
                  size="sm"
                  disabled={getActiveSessionsCount() <= 1}
                >
                  Terminate All Others
                </Button>
              </div>
            </div>

            <div className="space-y-4">
              {sessions.filter(session => session.isActive).map((session) => {
                const isCurrentSession = session.id === currentSession?.id;
                return (
                  <Card key={session.id} className={isCurrentSession ? 'border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-900/20' : ''}>
                    <div className="p-6">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                          <div className="p-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
                            <span className="text-xl">{getSessionDevice(session.userAgent).split(' ')[0]}</span>
                          </div>
                          <div>
                            <div className="flex items-center space-x-2">
                              <h4 className="font-medium text-gray-900 dark:text-white">
                                {getSessionDevice(session.userAgent)}
                              </h4>
                              {isCurrentSession && (
                                <span className="px-2 py-1 text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 rounded font-medium">
                                  Current Session
                                </span>
                              )}
                            </div>
                            <div className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                              <p>{getSessionLocation(session.ipAddress)} • {session.ipAddress}</p>
                              <p>Started {new Date().toLocaleDateString()}</p>
                              <p>Expires {new Date().toLocaleDateString()}</p>
                            </div>
                          </div>
                        </div>
                        
                        <div className="flex items-center space-x-3">
                          <div className="text-right text-sm">
                            <div className="text-gray-900 dark:text-white font-medium">
                              {new Date().toLocaleDateString()}
                            </div>
                            <div className="text-gray-500 dark:text-gray-400">Last activity</div>
                          </div>
                          
                          {!isCurrentSession && (
                            <Button
                              onClick={() => handleTerminateSession(session)}
                              variant="danger"
                              size="sm"
                            >
                              Terminate
                            </Button>
                          )}
                        </div>
                      </div>
                    </div>
                  </Card>
                );
              })}
            </div>
          </div>
        )}

        {/* Events Tab */}
        {activeTab === 'events' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Security Events
              </h3>
              <select
                value={eventFilter}
                onChange={(e) => setEventFilter(e.target.value as 'all' | 'high' | 'critical')}
                className="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-1 text-sm text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                <option value="all">All Events</option>
                <option value="high">High Severity</option>
                <option value="critical">Critical Only</option>
              </select>
            </div>

            <div className="space-y-3">
              {filteredEvents.map((event) => (
                <Card key={event.id}>
                  <div className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <span className="text-xl">{getEventIcon(event.type)}</span>
                        <div>
                          <div className="flex items-center space-x-2">
                            <h4 className="font-medium text-gray-900 dark:text-white">
                              {event.action || event.type}
                            </h4>
                            {getSeverityBadge((event as any).severity || 'low')}
                          </div>
                          <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {new Date().toLocaleDateString()}
                            {event.ipAddress && ` • ${event.ipAddress}`}
                          </div>
                        </div>
                      </div>
                      
                      <Button
                        onClick={() => onAcknowledgeEvent(event.id)}
                        variant="secondary"
                        size="sm"
                      >
                        Acknowledge
                      </Button>
                    </div>
                    
                    {event.details && Object.keys(event.details).length > 0 && (
                      <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                        <details className="text-sm">
                          <summary className="cursor-pointer text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">
                            Additional Details
                          </summary>
                          <div className="mt-2 bg-gray-50 dark:bg-gray-800 p-3 rounded font-mono text-xs">
                            <pre>{JSON.stringify(event.details, null, 2)}</pre>
                          </div>
                        </details>
                      </div>
                    )}
                  </div>
                </Card>
              ))}
            </div>
          </div>
        )}

        {/* Settings Tab */}
        {activeTab === 'settings' && (
          <div className="space-y-6">
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Session Security
                </h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-white">Auto-logout on idle</h4>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Automatically sign out after 30 minutes of inactivity
                      </p>
                    </div>
                    <button className="relative inline-flex h-6 w-11 items-center rounded-full bg-blue-600">
                      <span className="inline-block h-4 w-4 transform rounded-full bg-white transition-transform translate-x-6" />
                    </button>
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-white">Email on new login</h4>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Send email notifications for new device logins
                      </p>
                    </div>
                    <button className="relative inline-flex h-6 w-11 items-center rounded-full bg-blue-600">
                      <span className="inline-block h-4 w-4 transform rounded-full bg-white transition-transform translate-x-6" />
                    </button>
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-white">Suspicious activity alerts</h4>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Notify on unusual login patterns or failed attempts
                      </p>
                    </div>
                    <button className="relative inline-flex h-6 w-11 items-center rounded-full bg-blue-600">
                      <span className="inline-block h-4 w-4 transform rounded-full bg-white transition-transform translate-x-6" />
                    </button>
                  </div>
                </div>
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Account Security
                </h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-white">Two-Factor Authentication</h4>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Secure your account with an additional verification step
                      </p>
                    </div>
                    <Button variant="primary" size="sm">
                      Setup 2FA
                    </Button>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-white">Backup Codes</h4>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Generate backup codes for account recovery
                      </p>
                    </div>
                    <Button variant="secondary" size="sm">
                      Generate Codes
                    </Button>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        )}

        {/* Alerts Tab */}
        {activeTab === 'alerts' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Security Alerts ({securityAlerts.length})
              </h3>
              <select
                value={alertFilter}
                onChange={(e) => setAlertFilter(e.target.value as 'all' | 'high' | 'critical')}
                className="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-1 text-sm text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                <option value="all">All Alerts</option>
                <option value="high">High Severity</option>
                <option value="critical">Critical Only</option>
              </select>
            </div>

            {loading && (
              <div className="text-center py-8">
                <div className="text-gray-600 dark:text-gray-400">Loading security alerts...</div>
              </div>
            )}

            {error && (
              <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
                <p className="text-red-800 dark:text-red-200">{error}</p>
              </div>
            )}

            <div className="space-y-3">
              {filteredAlerts.map((alert) => (
                <Card key={alert.id}>
                  <div className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <span className="text-xl">🚨</span>
                        <div>
                          <div className="flex items-center space-x-2">
                            <h4 className="font-medium text-gray-900 dark:text-white">
                              {alert.title}
                            </h4>
                            {getSeverityBadge(alert.severity)}
                            <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                              alert.status === 'active' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' :
                              alert.status === 'acknowledged' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' :
                              alert.status === 'resolved' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                              'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
                            }`}>
                              {alert.status.toUpperCase()}
                            </span>
                          </div>
                          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {alert.description}
                          </p>
                          <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {new Date(alert.timestamp).toLocaleDateString()}
                            {alert.source && ` • ${alert.source}`}
                            {alert.cluster && ` • ${alert.cluster}`}
                          </div>
                        </div>
                      </div>

                      <div className="flex space-x-2">
                        {alert.status === 'active' && (
                          <>
                            <Button
                              onClick={() => onAcknowledgeAlert(alert.id)}
                              variant="secondary"
                              size="sm"
                            >
                              Acknowledge
                            </Button>
                            <Button
                              onClick={() => onResolveAlert(alert.id, 'Resolved manually')}
                              variant="primary"
                              size="sm"
                            >
                              Resolve
                            </Button>
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                </Card>
              ))}
              {filteredAlerts.length === 0 && !loading && (
                <div className="text-center py-8">
                  <div className="text-gray-600 dark:text-gray-400">No security alerts found.</div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Compliance Tab */}
        {activeTab === 'compliance' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Compliance Status
              </h3>
              <Button onClick={loadSecurityData} variant="secondary" size="sm" loading={loading}>
                Refresh
              </Button>
            </div>

            {securityMetrics && (
              <Card>
                <div className="p-6">
                  <h4 className="font-semibold text-gray-900 dark:text-white mb-4">Overall Compliance Score</h4>
                  <div className="flex items-center space-x-4">
                    <div className="text-4xl font-bold text-green-600 dark:text-green-400">
                      {securityMetrics.complianceScore}%
                    </div>
                    <div className="flex-1">
                      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                        <div
                          className="bg-green-600 dark:bg-green-400 h-2 rounded-full"
                          style={{ width: `${securityMetrics.complianceScore}%` }}
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            )}

            <div className="space-y-3">
              {complianceResults.map((result) => (
                <Card key={result.id}>
                  <div className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <span className="text-lg">
                          {result.status === 'compliant' ? '✅' :
                           result.status === 'non_compliant' ? '❌' :
                           result.status === 'not_applicable' ? '➖' : '⏳'}
                        </span>
                        <div>
                          <div className="flex items-center space-x-2">
                            <h4 className="font-medium text-gray-900 dark:text-white">
                              {result.title}
                            </h4>
                            <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                              {result.framework}
                            </span>
                            {getSeverityBadge(result.severity)}
                          </div>
                          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {result.description}
                          </p>
                          <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            Control: {result.control} • Last checked: {new Date(result.lastChecked).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>
              ))}
              {complianceResults.length === 0 && !loading && (
                <div className="text-center py-8">
                  <div className="text-gray-600 dark:text-gray-400">No compliance results available.</div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Scans Tab */}
        {activeTab === 'scans' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                Security Scans
              </h3>
              <Button onClick={loadSecurityData} variant="secondary" size="sm" loading={loading}>
                Refresh
              </Button>
            </div>

            <div className="space-y-3">
              {securityScans.map((scan) => (
                <Card key={scan.id}>
                  <div className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <span className="text-lg">
                          {scan.status === 'completed' ? '✅' :
                           scan.status === 'running' ? '⏳' :
                           scan.status === 'failed' ? '❌' : '⏸️'}
                        </span>
                        <div>
                          <div className="flex items-center space-x-2">
                            <h4 className="font-medium text-gray-900 dark:text-white">
                              {scan.type.toUpperCase()} Scan
                            </h4>
                            <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                              scan.status === 'completed' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                              scan.status === 'running' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' :
                              scan.status === 'failed' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' :
                              'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
                            }`}>
                              {scan.status.toUpperCase()}
                            </span>
                          </div>
                          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            Target: {scan.target}
                          </p>
                          <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            Started: {new Date(scan.startTime).toLocaleDateString()}
                            {scan.endTime && ` • Completed: ${new Date(scan.endTime).toLocaleDateString()}`}
                          </div>
                          {scan.status === 'completed' && (
                            <div className="mt-2 flex space-x-4 text-sm">
                              <span className="text-red-600 dark:text-red-400">Critical: {scan.findings.critical}</span>
                              <span className="text-orange-600 dark:text-orange-400">High: {scan.findings.high}</span>
                              <span className="text-yellow-600 dark:text-yellow-400">Medium: {scan.findings.medium}</span>
                              <span className="text-green-600 dark:text-green-400">Low: {scan.findings.low}</span>
                            </div>
                          )}
                          {scan.status === 'running' && scan.progress !== undefined && (
                            <div className="mt-2">
                              <div className="flex items-center space-x-2">
                                <div className="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                                  <div
                                    className="bg-blue-600 dark:bg-blue-400 h-2 rounded-full transition-all"
                                    style={{ width: `${scan.progress}%` }}
                                  ></div>
                                </div>
                                <span className="text-sm text-gray-600 dark:text-gray-400">{scan.progress}%</span>
                              </div>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>
              ))}
              {securityScans.length === 0 && !loading && (
                <div className="text-center py-8">
                  <div className="text-gray-600 dark:text-gray-400">No security scans available.</div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Terminate Session Modal */}
      <Modal
        isOpen={showTerminateModal}
        onClose={() => setShowTerminateModal(false)}
        title="Terminate Session"
      >
        <div className="space-y-4">
          <div className="bg-yellow-50 dark:bg-yellow-900/20 p-4 rounded-lg">
            <p className="text-yellow-800 dark:text-yellow-200">
              Are you sure you want to terminate this session? The user will be signed out immediately.
            </p>
          </div>

          {sessionToTerminate && (
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="space-y-2 text-sm">
                <div><strong>Device:</strong> {getSessionDevice(sessionToTerminate.userAgent)}</div>
                <div><strong>Location:</strong> {getSessionLocation(sessionToTerminate.ipAddress)}</div>
                <div><strong>Started:</strong> {new Date().toLocaleDateString()}</div>
              </div>
            </div>
          )}

          <div className="flex space-x-3">
            <Button
              onClick={confirmTerminateSession}
              variant="danger"
              className="flex-1"
            >
              Terminate Session
            </Button>
            <Button
              onClick={() => setShowTerminateModal(false)}
              variant="secondary"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>

      {/* Terminate All Sessions Modal */}
      <Modal
        isOpen={showTerminateAllModal}
        onClose={() => setShowTerminateAllModal(false)}
        title="Terminate All Other Sessions"
      >
        <div className="space-y-4">
          <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
            <p className="text-red-800 dark:text-red-200">
              This will sign out all other active sessions except your current one. 
              Users on those sessions will need to sign in again.
            </p>
          </div>

          <div className="text-sm text-gray-600 dark:text-gray-400">
            <p><strong>Sessions to terminate:</strong> {getActiveSessionsCount() - 1}</p>
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={confirmTerminateAllSessions}
              variant="danger"
              className="flex-1"
            >
              Terminate All Others
            </Button>
            <Button
              onClick={() => setShowTerminateAllModal(false)}
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
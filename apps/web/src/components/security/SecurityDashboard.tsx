import { useState, useEffect } from 'react';
import { User } from '../../types/user';

interface Session {
  id: string;
  userId: string;
  deviceType: string;
  browser: string;
  location: string;
  ipAddress: string;
  userAgent: string;
  isActive: boolean;
  lastActivity: Date;
  createdAt: Date;
  updatedAt: Date;
  expiresAt: Date;
}

interface SecurityEvent {
  id: string;
  userId: string;
  type: string;
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  timestamp: Date;
  ipAddress: string;
  metadata?: Record<string, unknown>;
}

import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';
// import { formatDistanceToNow } from 'date-fns';

interface SecurityDashboardProps {
  user: User;
  sessions: Session[];
  securityEvents: SecurityEvent[];
  onTerminateSession: (sessionId: string) => Promise<void>;
  onTerminateAllSessions: () => Promise<void>;
  onAcknowledgeEvent: (eventId: string) => Promise<void>;
  onRefreshSessions: () => Promise<void>;
  loading?: boolean;
  className?: string;
}

export function SecurityDashboard({
  user,
  sessions,
  securityEvents,
  onTerminateSession,
  onTerminateAllSessions,
  onAcknowledgeEvent,
  onRefreshSessions,
  loading = false,
  className = '',
}: SecurityDashboardProps) {
  const [activeTab, setActiveTab] = useState<'overview' | 'sessions' | 'events' | 'settings'>('overview');
  const [showTerminateModal, setShowTerminateModal] = useState(false);
  const [sessionToTerminate, setSessionToTerminate] = useState<Session | null>(null);
  const [showTerminateAllModal, setShowTerminateAllModal] = useState(false);
  const [eventFilter, setEventFilter] = useState<'all' | 'high' | 'critical'>('all');

  useEffect(() => {
    onRefreshSessions();
  }, [onRefreshSessions]);

  // const getCurrentSession = () => {
  //   // In a real implementation, this would be based on the current session token
  //   return sessions.find(session => session.isActive) || sessions[0];
  // };

  const getActiveSessionsCount = () => sessions.filter(session => session.isActive).length;
  const getSuspiciousEventsCount = () => securityEvents.filter(event => 
    event.severity === 'high' || event.severity === 'critical'
  ).length;
  const getRecentLoginAttempts = () => securityEvents.filter(event => 
    event.type === 'failed_login' && 
    new Date(event.timestamp).getTime() > Date.now() - 24 * 60 * 60 * 1000
  ).length;

  const filteredEvents = securityEvents.filter(event => {
    if (eventFilter === 'all') return true;
    return event.severity === eventFilter;
  });

  const handleTerminateSession = async (session: Session) => {
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
                {getSuspiciousEventsCount()}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Security Alerts</div>
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
            { id: 'sessions', label: 'Active Sessions', icon: '🔐', count: getActiveSessionsCount() },
            { id: 'events', label: 'Security Events', icon: '📋', count: getSuspiciousEventsCount() },
            { id: 'settings', label: 'Security Settings', icon: '⚙️' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as 'overview' | 'sessions' | 'events' | 'settings')}
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
                  {securityEvents.slice(0, 5).map((event) => (
                    <div key={event.id} className="flex items-center space-x-3">
                      <span className="text-lg">{getEventIcon(event.type)}</span>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                          {event.description}
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          {new Date().toLocaleDateString()}
                        </p>
                      </div>
                      {getSeverityBadge(event.severity)}
                    </div>
                  ))}
                </div>
              </div>
            </Card>
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
                              {event.description}
                            </h4>
                            {getSeverityBadge(event.severity)}
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
                    
                    {event.metadata && Object.keys(event.metadata).length > 0 && (
                      <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                        <details className="text-sm">
                          <summary className="cursor-pointer text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white">
                            Additional Details
                          </summary>
                          <div className="mt-2 bg-gray-50 dark:bg-gray-800 p-3 rounded font-mono text-xs">
                            <pre>{JSON.stringify(event.metadata, null, 2)}</pre>
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
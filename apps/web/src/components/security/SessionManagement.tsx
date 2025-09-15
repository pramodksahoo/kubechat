import { useState, useEffect } from 'react';
import { Session, User } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
import { formatDistanceToNow, format } from 'date-fns';

interface SessionManagementProps {
  currentUser: User;
  allSessions: Session[];
  onTerminateSession: (sessionId: string) => Promise<void>;
  // onTerminateUserSessions: (userId: string) => Promise<void>;
  onExtendSession: (sessionId: string, hours: number) => Promise<void>;
  onRefreshSessions: () => Promise<void>;
  onUpdateSessionSettings: (settings: SessionSettings) => Promise<void>;
  sessionSettings: SessionSettings;
  isAdmin?: boolean;
  loading?: boolean;
  className?: string;
}

interface SessionSettings {
  maxSessionDuration: number; // hours
  idleTimeout: number; // minutes
  maxConcurrentSessions: number;
  requireReauth: boolean;
  allowRememberMe: boolean;
  sessionWarningMinutes: number;
}

export function SessionManagement({
  currentUser,
  allSessions,
  onTerminateSession,
  // onTerminateUserSessions,
  onExtendSession,
  onRefreshSessions,
  onUpdateSessionSettings,
  sessionSettings,
  isAdmin = false,
  loading = false,
  className = '',
}: SessionManagementProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'expired'>('all');
  const [selectedSessions, setSelectedSessions] = useState<Set<string>>(new Set());
  const [showTerminateModal, setShowTerminateModal] = useState(false);
  const [showExtendModal, setShowExtendModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [sessionToExtend, setSessionToExtend] = useState<Session | null>(null);
  const [extensionHours, setExtensionHours] = useState(2);
  const [localSettings, setLocalSettings] = useState<SessionSettings>(sessionSettings);

  useEffect(() => {
    onRefreshSessions();
  }, [onRefreshSessions]);

  useEffect(() => {
    setLocalSettings(sessionSettings);
  }, [sessionSettings]);

  const getFilteredSessions = () => {
    return allSessions.filter(session => {
      const matchesSearch = searchTerm === '' || 
        session.userId.toLowerCase().includes(searchTerm.toLowerCase()) ||
        session.ipAddress?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        session.userAgent?.toLowerCase().includes(searchTerm.toLowerCase());

      const matchesStatus = filterStatus === 'all' || 
        (filterStatus === 'active' && session.isActive) ||
        (filterStatus === 'expired' && !session.isActive);

      return matchesSearch && matchesStatus;
    });
  };

  const getSessionStats = () => {
    const total = allSessions.length;
    const active = allSessions.filter(s => s.isActive).length;
    const expired = total - active;
    const expiringSoon = allSessions.filter(s => 
      s.isActive && new Date(s.expiresAt).getTime() < Date.now() + (60 * 60 * 1000) // 1 hour
    ).length;

    return { total, active, expired, expiringSoon };
  };

  const getDeviceIcon = (userAgent?: string) => {
    if (!userAgent) return '💻';
    
    if (userAgent.includes('Mobile')) return '📱';
    if (userAgent.includes('Tablet')) return '📱';
    if (userAgent.includes('iPhone')) return '📱';
    if (userAgent.includes('Android')) return '📱';
    
    return '💻';
  };

  const getBrowserInfo = (userAgent?: string) => {
    if (!userAgent) return 'Unknown Browser';
    
    if (userAgent.includes('Chrome')) return 'Chrome';
    if (userAgent.includes('Firefox')) return 'Firefox';
    if (userAgent.includes('Safari')) return 'Safari';
    if (userAgent.includes('Edge')) return 'Edge';
    
    return 'Unknown Browser';
  };

  const getLocationInfo = (ipAddress?: string) => {
    if (!ipAddress) return 'Unknown Location';
    if (ipAddress === '127.0.0.1' || ipAddress === 'localhost') return 'Local';
    // In real implementation, this would use geo-IP lookup
    return 'External';
  };

  const handleSelectSession = (sessionId: string) => {
    const newSelected = new Set(selectedSessions);
    if (newSelected.has(sessionId)) {
      newSelected.delete(sessionId);
    } else {
      newSelected.add(sessionId);
    }
    setSelectedSessions(newSelected);
  };

  const handleSelectAll = () => {
    const filteredSessions = getFilteredSessions();
    if (selectedSessions.size === filteredSessions.length) {
      setSelectedSessions(new Set());
    } else {
      setSelectedSessions(new Set(filteredSessions.map(s => s.id)));
    }
  };

  const handleBulkTerminate = async () => {
    for (const sessionId of selectedSessions) {
      await onTerminateSession(sessionId);
    }
    setSelectedSessions(new Set());
    setShowTerminateModal(false);
  };

  const handleExtendSession = async () => {
    if (sessionToExtend) {
      await onExtendSession(sessionToExtend.id, extensionHours);
      setShowExtendModal(false);
      setSessionToExtend(null);
      setExtensionHours(2);
    }
  };

  const handleSaveSettings = async () => {
    await onUpdateSessionSettings(localSettings);
    setShowSettingsModal(false);
  };

  const isSessionExpiringSoon = (session: Session) => {
    return new Date(session.expiresAt).getTime() < Date.now() + (60 * 60 * 1000); // 1 hour
  };

  const filteredSessions = getFilteredSessions();
  const stats = getSessionStats();

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
              <span className="text-blue-600 dark:text-blue-400 text-xl">📊</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {stats.total}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Total Sessions</div>
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
                {stats.active}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Active Sessions</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
              <span className="text-yellow-600 dark:text-yellow-400 text-xl">⏰</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {stats.expiringSoon}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Expiring Soon</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
              <span className="text-gray-600 dark:text-gray-400 text-xl">❌</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {stats.expired}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Expired</div>
            </div>
          </div>
        </Card>
      </div>

      {/* Controls */}
      <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
        <div className="flex items-center space-x-4 w-full sm:w-auto">
          <div className="flex-1 sm:w-64">
            <Input
              value={searchTerm}
              onChange={(value) => setSearchTerm(value)}
              placeholder="Search sessions..."
              className="w-full"
            />
          </div>
          
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value as 'all' | 'active' | 'expired')}
            className="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
          >
            <option value="all">All Sessions</option>
            <option value="active">Active Only</option>
            <option value="expired">Expired Only</option>
          </select>
        </div>

        <div className="flex items-center space-x-2">
          {selectedSessions.size > 0 && (
            <Button
              onClick={() => setShowTerminateModal(true)}
              variant="danger"
              size="sm"
            >
              Terminate Selected ({selectedSessions.size})
            </Button>
          )}
          
          {isAdmin && (
            <Button
              onClick={() => setShowSettingsModal(true)}
              variant="secondary"
              size="sm"
            >
              ⚙️ Settings
            </Button>
          )}
          
          <Button
            onClick={onRefreshSessions}
            variant="secondary"
            size="sm"
            loading={loading}
          >
            🔄 Refresh
          </Button>
        </div>
      </div>

      {/* Sessions List */}
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Session Management
            </h3>
            
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={selectedSessions.size === filteredSessions.length && filteredSessions.length > 0}
                onChange={handleSelectAll}
                className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <label className="text-sm text-gray-600 dark:text-gray-400">
                Select All
              </label>
            </div>
          </div>

          <div className="space-y-3">
            {filteredSessions.map((session) => {
              const isCurrentUser = session.userId === currentUser.id;
              const isExpiring = isSessionExpiringSoon(session);
              
              return (
                <div
                  key={session.id}
                  className={`border rounded-lg p-4 transition-colors ${
                    selectedSessions.has(session.id) ? 'ring-2 ring-blue-500' : ''
                  } ${
                    isCurrentUser ? 'border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-900/20' : ''
                  } ${
                    isExpiring && session.isActive ? 'border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-900/20' : ''
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      {isAdmin && (
                        <input
                          type="checkbox"
                          checked={selectedSessions.has(session.id)}
                          onChange={() => handleSelectSession(session.id)}
                          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                        />
                      )}
                      
                      <div className="p-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
                        <span className="text-xl">{getDeviceIcon(session.userAgent)}</span>
                      </div>
                      
                      <div>
                        <div className="flex items-center space-x-2">
                          <h4 className="font-medium text-gray-900 dark:text-white">
                            {session.userId}
                          </h4>
                          
                          {isCurrentUser && (
                            <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded font-medium">
                              Current Session
                            </span>
                          )}
                          
                          <span className={`px-2 py-1 text-xs rounded font-medium ${
                            session.isActive
                              ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                              : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                          }`}>
                            {session.isActive ? 'Active' : 'Expired'}
                          </span>
                          
                          {isExpiring && session.isActive && (
                            <span className="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 rounded font-medium">
                              Expiring Soon
                            </span>
                          )}
                        </div>
                        
                        <div className="text-sm text-gray-600 dark:text-gray-400 mt-1 space-y-1">
                          <div className="flex items-center space-x-4">
                            <span>{getBrowserInfo(session.userAgent)}</span>
                            <span>{getLocationInfo(session.ipAddress)} • {session.ipAddress}</span>
                          </div>
                          <div className="flex items-center space-x-4">
                            <span>Started: {new Date().toLocaleDateString()}</span>
                            <span>Expires: {new Date().toLocaleDateString()}</span>
                            <span>Last active: {new Date().toLocaleDateString()}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                    
                    <div className="flex items-center space-x-2">
                      {session.isActive && isAdmin && (
                        <Button
                          onClick={() => {
                            setSessionToExtend(session);
                            setShowExtendModal(true);
                          }}
                          variant="secondary"
                          size="sm"
                        >
                          Extend
                        </Button>
                      )}
                      
                      {session.isActive && (
                        <Button
                          onClick={() => onTerminateSession(session.id)}
                          variant="danger"
                          size="sm"
                          disabled={isCurrentUser && !isAdmin}
                        >
                          Terminate
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {filteredSessions.length === 0 && (
            <div className="text-center py-8">
              <div className="text-gray-500 dark:text-gray-400">
                <span className="text-4xl mb-4 block">🔍</span>
                <p className="text-lg font-medium mb-2">No sessions found</p>
                <p className="text-sm">Try adjusting your search or filters</p>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Terminate Multiple Sessions Modal */}
      <Modal
        isOpen={showTerminateModal}
        onClose={() => setShowTerminateModal(false)}
        title="Terminate Sessions"
      >
        <div className="space-y-4">
          <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
            <p className="text-red-800 dark:text-red-200">
              Are you sure you want to terminate {selectedSessions.size} selected session(s)? 
              Users will be signed out immediately.
            </p>
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleBulkTerminate}
              variant="danger"
              className="flex-1"
            >
              Terminate Sessions
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

      {/* Extend Session Modal */}
      <Modal
        isOpen={showExtendModal}
        onClose={() => setShowExtendModal(false)}
        title="Extend Session"
      >
        <div className="space-y-4">
          {sessionToExtend && (
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="space-y-2 text-sm">
                <div><strong>User:</strong> {sessionToExtend.userId}</div>
                <div><strong>Current Expiry:</strong> {new Date().toLocaleDateString()}</div>
                <div><strong>Device:</strong> {getBrowserInfo(sessionToExtend.userAgent)}</div>
              </div>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Extend by (hours)
            </label>
            <input
              type="number"
              value={extensionHours}
              onChange={(e) => setExtensionHours(parseInt(e.target.value))}
              min="1"
              max="24"
              className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
            />
          </div>

          <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
            <p className="text-sm text-blue-700 dark:text-blue-300">
              Session will be extended until {sessionToExtend && 
                format(new Date(Date.now() + extensionHours * 60 * 60 * 1000), 'MMM dd, yyyy HH:mm')
              }
            </p>
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleExtendSession}
              variant="primary"
              className="flex-1"
            >
              Extend Session
            </Button>
            <Button
              onClick={() => setShowExtendModal(false)}
              variant="secondary"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>

      {/* Session Settings Modal */}
      <Modal
        isOpen={showSettingsModal}
        onClose={() => setShowSettingsModal(false)}
        title="Session Settings"
      >
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Max Session Duration (hours)
              </label>
              <input
                type="number"
                value={localSettings.maxSessionDuration}
                onChange={(e) => setLocalSettings(prev => ({ 
                  ...prev, 
                  maxSessionDuration: parseInt(e.target.value) 
                }))}
                min="1"
                max="168"
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Idle Timeout (minutes)
              </label>
              <input
                type="number"
                value={localSettings.idleTimeout}
                onChange={(e) => setLocalSettings(prev => ({ 
                  ...prev, 
                  idleTimeout: parseInt(e.target.value) 
                }))}
                min="5"
                max="240"
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Max Concurrent Sessions
              </label>
              <input
                type="number"
                value={localSettings.maxConcurrentSessions}
                onChange={(e) => setLocalSettings(prev => ({ 
                  ...prev, 
                  maxConcurrentSessions: parseInt(e.target.value) 
                }))}
                min="1"
                max="10"
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Session Warning (minutes before expiry)
              </label>
              <input
                type="number"
                value={localSettings.sessionWarningMinutes}
                onChange={(e) => setLocalSettings(prev => ({ 
                  ...prev, 
                  sessionWarningMinutes: parseInt(e.target.value) 
                }))}
                min="1"
                max="60"
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>
          </div>

          <div className="space-y-4">
            {[
              { key: 'requireReauth', label: 'Require Re-authentication for Sensitive Operations' },
              { key: 'allowRememberMe', label: 'Allow "Remember Me" Option' },
            ].map(({ key, label }) => (
              <div key={key} className="flex items-center justify-between">
                <span className="text-sm text-gray-700 dark:text-gray-300">{label}</span>
                <button
                  onClick={() => setLocalSettings(prev => ({
                    ...prev,
                    [key]: !prev[key as keyof SessionSettings],
                  }))}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    localSettings[key as keyof SessionSettings]
                      ? 'bg-blue-600'
                      : 'bg-gray-200 dark:bg-gray-700'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      localSettings[key as keyof SessionSettings]
                        ? 'translate-x-6'
                        : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>
            ))}
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleSaveSettings}
              variant="primary"
              className="flex-1"
            >
              Save Settings
            </Button>
            <Button
              onClick={() => setShowSettingsModal(false)}
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
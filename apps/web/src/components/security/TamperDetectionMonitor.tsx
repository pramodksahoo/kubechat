import React, { useState, useEffect, useCallback } from 'react';
import { TamperAlert, SuspiciousActivity } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { Icon } from '../ui/Icon';
import { Modal } from '../ui/Modal';
import { format } from 'date-fns';

interface TamperDetectionMonitorProps {
  className?: string;
}

interface TamperDetectionConfig {
  enableRealTimeMonitoring: boolean;
  alertThresholds: {
    failedLoginAttempts: number;
    suspiciousApiCalls: number;
    timeWindow: number; // minutes
  };
  monitoringScope: {
    auditLogs: boolean;
    userActions: boolean;
    systemEvents: boolean;
    apiCalls: boolean;
  };
}

export function TamperDetectionMonitor({ className = '' }: TamperDetectionMonitorProps) {
  const [tamperAlerts, setTamperAlerts] = useState<TamperAlert[]>([]);
  const [suspiciousActivities, setSuspiciousActivities] = useState<SuspiciousActivity[]>([]);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [showConfig, setShowConfig] = useState(false);
  const [showAlertDetails, setShowAlertDetails] = useState(false);
  const [selectedAlert, setSelectedAlert] = useState<TamperAlert | null>(null);
  const [loading, setLoading] = useState(false);

  const [config, setConfig] = useState<TamperDetectionConfig>({
    enableRealTimeMonitoring: false,
    alertThresholds: {
      failedLoginAttempts: 5,
      suspiciousApiCalls: 10,
      timeWindow: 15
    },
    monitoringScope: {
      auditLogs: true,
      userActions: true,
      systemEvents: true,
      apiCalls: true
    }
  });

  // Real-time monitoring
  useEffect(() => {
    let interval: NodeJS.Timeout;

    if (isMonitoring && config.enableRealTimeMonitoring) {
      interval = setInterval(() => {
        loadTamperAlerts();
        loadSuspiciousActivities();
      }, 5000); // Check every 5 seconds
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isMonitoring, config.enableRealTimeMonitoring]);

  // Load tamper alerts
  const loadTamperAlerts = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/audit/tamper-alerts?limit=50');
      if (response.ok) {
        const data = await response.json();
        setTamperAlerts(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load tamper alerts:', error);
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

  // Start/stop monitoring
  const toggleMonitoring = useCallback(async () => {
    setLoading(true);
    try {
      const action = isMonitoring ? 'stop' : 'start';
      const response = await fetch(`/api/v1/audit/tamper-detection/${action}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(config),
      });

      if (response.ok) {
        setIsMonitoring(!isMonitoring);
        if (!isMonitoring) {
          // Load initial data when starting monitoring
          loadTamperAlerts();
          loadSuspiciousActivities();
        }
      }
    } catch (error) {
      console.error('Failed to toggle monitoring:', error);
    } finally {
      setLoading(false);
    }
  }, [isMonitoring, config, loadTamperAlerts, loadSuspiciousActivities]);

  // Update configuration
  const updateConfig = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/audit/tamper-detection/config', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(config),
      });

      if (response.ok) {
        setShowConfig(false);
      }
    } catch (error) {
      console.error('Failed to update configuration:', error);
    }
  }, [config]);

  // Acknowledge alert
  const acknowledgeAlert = useCallback(async (alertId: string) => {
    try {
      const response = await fetch(`/api/v1/audit/tamper-alerts/${alertId}/acknowledge`, {
        method: 'POST',
      });

      if (response.ok) {
        setTamperAlerts(prev => prev.filter(alert => alert.id !== alertId));
      }
    } catch (error) {
      console.error('Failed to acknowledge alert:', error);
    }
  }, []);

  useEffect(() => {
    loadTamperAlerts();
    loadSuspiciousActivities();
  }, [loadTamperAlerts, loadSuspiciousActivities]);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'high': return 'orange';
      case 'medium': return 'yellow';
      case 'low': return 'green';
      default: return 'gray';
    }
  };

  const getStatusIcon = (isActive: boolean) => {
    return isActive ? 'play' : 'pause';
  };

  return (
    <>
      <div className={`space-y-6 ${className}`}>
        {/* Header and Controls */}
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Tamper Detection & Monitoring</h2>
            <p className="text-gray-600">Real-time security monitoring and threat detection</p>
          </div>
          <div className="flex space-x-3">
            <Button
              variant="secondary"
              onClick={() => setShowConfig(true)}
            >
              <Icon name="cog" className="w-4 h-4 mr-2" />
              Configure
            </Button>
            <Button
              variant={isMonitoring ? 'danger' : 'primary'}
              onClick={toggleMonitoring}
              disabled={loading}
            >
              {loading ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              ) : (
                <Icon name={getStatusIcon(isMonitoring)} className="w-4 h-4 mr-2" />
              )}
              {isMonitoring ? 'Stop Monitoring' : 'Start Monitoring'}
            </Button>
          </div>
        </div>

        {/* Monitoring Status */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Monitoring Status</h3>
                  <p className="text-sm text-gray-600">Real-time detection</p>
                </div>
                <div className={`w-3 h-3 rounded-full ${isMonitoring ? 'bg-green-500 animate-pulse' : 'bg-gray-300'}`}></div>
              </div>
              <div className="mt-4">
                <Badge variant={isMonitoring ? 'green' : 'gray'}>
                  {isMonitoring ? 'ACTIVE' : 'INACTIVE'}
                </Badge>
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Active Alerts</h3>
                  <p className="text-sm text-gray-600">Requires attention</p>
                </div>
                <Icon name="exclamation-triangle" className={`w-8 h-8 ${tamperAlerts.length > 0 ? 'text-red-600' : 'text-green-600'}`} />
              </div>
              <div className="mt-4">
                <div className={`text-3xl font-bold ${tamperAlerts.length > 0 ? 'text-red-600' : 'text-green-600'}`}>
                  {tamperAlerts.length}
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Suspicious Activities</h3>
                  <p className="text-sm text-gray-600">Last 24 hours</p>
                </div>
                <Icon name="eye" className="w-8 h-8 text-yellow-600" />
              </div>
              <div className="mt-4">
                <div className="text-3xl font-bold text-yellow-600">
                  {suspiciousActivities.length}
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Threat Level</h3>
                  <p className="text-sm text-gray-600">Current assessment</p>
                </div>
                <Icon name="shield-check" className="w-8 h-8 text-blue-600" />
              </div>
              <div className="mt-4">
                <Badge variant={tamperAlerts.filter(a => a.severity === 'critical').length > 0 ? 'red' : 'green'}>
                  {tamperAlerts.filter(a => a.severity === 'critical').length > 0 ? 'HIGH' : 'LOW'}
                </Badge>
              </div>
            </div>
          </Card>
        </div>

        {/* Active Tamper Alerts */}
        {tamperAlerts.length > 0 && (
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-medium mb-4 text-red-600">Active Tamper Alerts</h3>
              <div className="space-y-4">
                {tamperAlerts.slice(0, 10).map((alert) => (
                  <div
                    key={alert.id}
                    className="border border-red-200 rounded-md p-4 bg-red-50 cursor-pointer hover:bg-red-100"
                    onClick={() => {
                      setSelectedAlert(alert);
                      setShowAlertDetails(true);
                    }}
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Badge variant={getSeverityColor(alert.severity)}>
                            {alert.severity.toUpperCase()}
                          </Badge>
                          <span className="text-sm text-gray-600">
                            {format(new Date(alert.detectedAt), 'MMM dd, HH:mm:ss')}
                          </span>
                        </div>
                        <h4 className="font-medium text-red-800 mb-1">{alert.violationType}</h4>
                        <p className="text-sm text-red-700">{alert.description}</p>
                        <p className="text-xs text-gray-600 mt-1">
                          Affected Log: {alert.affectedLogId}
                        </p>
                      </div>
                      <div className="flex space-x-2">
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => acknowledgeAlert(alert.id)}
                        >
                          Acknowledge
                        </Button>
                        <Icon name="chevron-right" className="w-5 h-5 text-gray-400" />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        )}

        {/* Suspicious Activities */}
        {suspiciousActivities.length > 0 && (
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-medium mb-4 text-yellow-600">Suspicious Activities</h3>
              <div className="space-y-3">
                {suspiciousActivities.slice(0, 5).map((activity, index) => (
                  <div key={index} className="border border-yellow-200 rounded-md p-3 bg-yellow-50">
                    <div className="flex justify-between items-start">
                      <div>
                        <h4 className="font-medium text-yellow-800">{activity.type}</h4>
                        <p className="text-sm text-yellow-700">{activity.description}</p>
                        <p className="text-xs text-gray-600 mt-1">
                          User: {activity.userId || 'System'} • {format(new Date(activity.detectedAt), 'MMM dd, HH:mm:ss')}
                        </p>
                      </div>
                      <Badge variant="yellow">
                        Risk: {activity.riskLevel}
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        )}
      </div>

      {/* Configuration Modal */}
      <Modal
        isOpen={showConfig}
        onClose={() => setShowConfig(false)}
        title="Tamper Detection Configuration"
        size="lg"
      >
        <div className="space-y-6">
          <div>
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={config.enableRealTimeMonitoring}
                onChange={(e) => setConfig(prev => ({
                  ...prev,
                  enableRealTimeMonitoring: e.target.checked
                }))}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <span className="ml-2 text-sm font-medium text-gray-700">Enable Real-time Monitoring</span>
            </label>
          </div>

          <div>
            <h4 className="font-medium text-gray-900 mb-3">Alert Thresholds</h4>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Failed Login Attempts
                </label>
                <input
                  type="number"
                  value={config.alertThresholds.failedLoginAttempts}
                  onChange={(e) => setConfig(prev => ({
                    ...prev,
                    alertThresholds: {
                      ...prev.alertThresholds,
                      failedLoginAttempts: parseInt(e.target.value) || 0
                    }
                  }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Suspicious API Calls
                </label>
                <input
                  type="number"
                  value={config.alertThresholds.suspiciousApiCalls}
                  onChange={(e) => setConfig(prev => ({
                    ...prev,
                    alertThresholds: {
                      ...prev.alertThresholds,
                      suspiciousApiCalls: parseInt(e.target.value) || 0
                    }
                  }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>
          </div>

          <div>
            <h4 className="font-medium text-gray-900 mb-3">Monitoring Scope</h4>
            <div className="space-y-2">
              {Object.entries(config.monitoringScope).map(([key, value]) => (
                <label key={key} className="flex items-center">
                  <input
                    type="checkbox"
                    checked={value}
                    onChange={(e) => setConfig(prev => ({
                      ...prev,
                      monitoringScope: {
                        ...prev.monitoringScope,
                        [key]: e.target.checked
                      }
                    }))}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span className="ml-2 text-sm text-gray-700 capitalize">
                    {key.replace(/([A-Z])/g, ' $1').toLowerCase()}
                  </span>
                </label>
              ))}
            </div>
          </div>

          <div className="flex justify-end space-x-3">
            <Button variant="secondary" onClick={() => setShowConfig(false)}>
              Cancel
            </Button>
            <Button variant="primary" onClick={updateConfig}>
              Save Configuration
            </Button>
          </div>
        </div>
      </Modal>

      {/* Alert Details Modal */}
      <Modal
        isOpen={showAlertDetails}
        onClose={() => setShowAlertDetails(false)}
        title="Tamper Alert Details"
        size="lg"
      >
        {selectedAlert && (
          <div className="space-y-4">
            <div className="bg-red-50 border border-red-200 rounded-md p-4">
              <div className="flex items-center space-x-2 mb-2">
                <Badge variant={getSeverityColor(selectedAlert.severity)}>
                  {selectedAlert.severity.toUpperCase()}
                </Badge>
                <span className="text-sm text-gray-600">
                  {format(new Date(selectedAlert.detectedAt), 'MMM dd, yyyy HH:mm:ss')}
                </span>
              </div>
              <h4 className="font-medium text-red-800 mb-2">{selectedAlert.violationType}</h4>
              <p className="text-sm text-red-700">{selectedAlert.description}</p>
            </div>

            <div>
              <h5 className="font-medium mb-2">Alert Details:</h5>
              <dl className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <dt className="text-gray-600">Alert ID:</dt>
                  <dd className="font-medium">{selectedAlert.id}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-gray-600">Affected Log:</dt>
                  <dd className="font-medium">{selectedAlert.affectedLogId}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-gray-600">Detection Method:</dt>
                  <dd className="font-medium">Cryptographic Validation</dd>
                </div>
              </dl>
            </div>

            <div className="flex justify-end space-x-3">
              <Button variant="secondary" onClick={() => setShowAlertDetails(false)}>
                Close
              </Button>
              <Button
                variant="danger"
                onClick={() => {
                  acknowledgeAlert(selectedAlert.id);
                  setShowAlertDetails(false);
                }}
              >
                Acknowledge Alert
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}
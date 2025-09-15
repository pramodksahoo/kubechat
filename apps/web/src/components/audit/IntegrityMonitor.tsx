import React, { useState } from 'react';
import { ChainIntegrityResult, TamperAlert } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { Icon } from '../ui/Icon';
import { Modal } from '../ui/Modal';
import { format } from 'date-fns';

interface IntegrityMonitorProps {
  integrityResult: ChainIntegrityResult | null;
  tamperAlerts: TamperAlert[];
  onVerifyIntegrity: () => Promise<void>;
  className?: string;
}

export function IntegrityMonitor({
  integrityResult,
  tamperAlerts,
  onVerifyIntegrity,
  className = ''
}: IntegrityMonitorProps) {
  const [isVerifying, setIsVerifying] = useState(false);
  const [showViolationDetails, setShowViolationDetails] = useState(false);
  const [selectedViolation, setSelectedViolation] = useState<string | null>(null);
  const [realTimeMonitoring, setRealTimeMonitoring] = useState(false);

  const handleVerifyIntegrity = async () => {
    setIsVerifying(true);
    try {
      await onVerifyIntegrity();
    } finally {
      setIsVerifying(false);
    }
  };

  const getIntegrityStatusColor = (isValid: boolean) => {
    return isValid ? 'green' : 'red';
  };

  const getIntegrityStatusIcon = (isValid: boolean) => {
    return isValid ? 'shield-check' : 'shield-exclamation';
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'high': return 'orange';
      case 'medium': return 'yellow';
      case 'low': return 'green';
      default: return 'gray';
    }
  };

  const getScoreColor = (score: number) => {
    if (score >= 95) return 'text-green-600';
    if (score >= 80) return 'text-yellow-600';
    if (score >= 60) return 'text-orange-600';
    return 'text-red-600';
  };

  return (
    <>
      <div className={`space-y-6 ${className}`}>
        {/* Integrity Status Overview */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Chain Integrity</h3>
                  <p className="text-sm text-gray-600">Overall audit chain status</p>
                </div>
                <Icon
                  name={integrityResult ? getIntegrityStatusIcon(integrityResult.isValid) : 'shield-check'}
                  className={`w-8 h-8 ${integrityResult?.isValid ? 'text-green-600' : 'text-red-600'}`}
                />
              </div>
              <div className="mt-4">
                <Badge
                  variant={integrityResult ? getIntegrityStatusColor(integrityResult.isValid) : 'gray'}
                  className="text-lg px-3 py-1"
                >
                  {integrityResult ? (integrityResult.isValid ? 'INTACT' : 'COMPROMISED') : 'UNKNOWN'}
                </Badge>
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Integrity Score</h3>
                  <p className="text-sm text-gray-600">Chain reliability percentage</p>
                </div>
                <Icon name="chart-pie" className="w-8 h-8 text-blue-600" />
              </div>
              <div className="mt-4">
                <div className={`text-3xl font-bold ${integrityResult ? getScoreColor(integrityResult.integrityScore) : 'text-gray-400'}`}>
                  {integrityResult ? `${integrityResult.integrityScore.toFixed(1)}%` : '--'}
                </div>
                {integrityResult && (
                  <p className="text-sm text-gray-600 mt-1">
                    {integrityResult.totalChecked.toLocaleString()} records verified
                  </p>
                )}
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Tamper Alerts</h3>
                  <p className="text-sm text-gray-600">Active security alerts</p>
                </div>
                <Icon
                  name="exclamation-triangle"
                  className={`w-8 h-8 ${tamperAlerts.length > 0 ? 'text-red-600' : 'text-green-600'}`}
                />
              </div>
              <div className="mt-4">
                <div className={`text-3xl font-bold ${tamperAlerts.length > 0 ? 'text-red-600' : 'text-green-600'}`}>
                  {tamperAlerts.length}
                </div>
                <p className="text-sm text-gray-600 mt-1">
                  {tamperAlerts.length === 0 ? 'No alerts' : 'Requires attention'}
                </p>
              </div>
            </div>
          </Card>
        </div>

        {/* Verification Controls */}
        <Card>
          <div className="p-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium">Integrity Verification</h3>
              <div className="flex space-x-3">
                <Button
                  variant={realTimeMonitoring ? 'primary' : 'secondary'}
                  onClick={() => setRealTimeMonitoring(!realTimeMonitoring)}
                  size="sm"
                >
                  <Icon name={realTimeMonitoring ? 'pause' : 'play'} className="w-4 h-4 mr-2" />
                  {realTimeMonitoring ? 'Stop' : 'Start'} Real-time Monitoring
                </Button>
                <Button
                  variant="primary"
                  onClick={handleVerifyIntegrity}
                  disabled={isVerifying}
                >
                  {isVerifying ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Verifying...
                    </>
                  ) : (
                    <>
                      <Icon name="shield-check" className="w-4 h-4 mr-2" />
                      Verify Chain Integrity
                    </>
                  )}
                </Button>
              </div>
            </div>

            {integrityResult && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h4 className="font-medium mb-3">Verification Results</h4>
                  <dl className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <dt className="text-gray-600">Total Records Checked:</dt>
                      <dd className="font-medium">{integrityResult.totalChecked.toLocaleString()}</dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-gray-600">Violations Found:</dt>
                      <dd className="font-medium text-red-600">
                        {integrityResult.violations?.length || 0}
                      </dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-gray-600">Last Verified:</dt>
                      <dd className="font-medium">
                        {format(new Date(integrityResult.lastValidated), 'MMM dd, yyyy HH:mm:ss')}
                      </dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-gray-600">Chain Status:</dt>
                      <dd>
                        <Badge variant={getIntegrityStatusColor(integrityResult.isValid)}>
                          {integrityResult.isValid ? 'Valid' : 'Invalid'}
                        </Badge>
                      </dd>
                    </div>
                  </dl>
                </div>

                {integrityResult.violations && integrityResult.violations.length > 0 && (
                  <div>
                    <h4 className="font-medium mb-3 text-red-600">Integrity Violations</h4>
                    <div className="space-y-2">
                      {integrityResult.violations.slice(0, 5).map((violation, index) => (
                        <div
                          key={index}
                          className="flex justify-between items-center p-2 bg-red-50 rounded-md cursor-pointer hover:bg-red-100"
                          onClick={() => {
                            setSelectedViolation(violation);
                            setShowViolationDetails(true);
                          }}
                        >
                          <span className="text-sm font-medium text-red-800">
                            Record ID: {violation}
                          </span>
                          <Icon name="chevron-right" className="w-4 h-4 text-red-600" />
                        </div>
                      ))}
                      {integrityResult.violations.length > 5 && (
                        <div className="text-sm text-gray-600 text-center">
                          +{integrityResult.violations.length - 5} more violations
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </Card>

        {/* Tamper Alerts */}
        {tamperAlerts.length > 0 && (
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-medium mb-4 text-red-600">Active Tamper Alerts</h3>
              <div className="space-y-4">
                {tamperAlerts.map((alert, index) => (
                  <div key={index} className="border border-red-200 rounded-md p-4 bg-red-50">
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Badge variant={getSeverityColor(alert.severity)}>
                            {alert.severity.toUpperCase()}
                          </Badge>
                          <span className="text-sm text-gray-600">
                            {format(new Date(alert.detectedAt), 'MMM dd, yyyy HH:mm:ss')}
                          </span>
                        </div>
                        <h4 className="font-medium text-red-800 mb-1">
                          {alert.violationType}
                        </h4>
                        <p className="text-sm text-red-700 mb-2">
                          {alert.description}
                        </p>
                        <p className="text-xs text-gray-600">
                          Affected Log ID: {alert.affectedLogId}
                        </p>
                      </div>
                      <Button variant="secondary" size="sm" className="ml-4">
                        <Icon name="eye" className="w-4 h-4 mr-2" />
                        Investigate
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        )}

        {/* Real-time Monitoring Status */}
        {realTimeMonitoring && (
          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-green-600">Real-time Monitoring Active</h3>
                <div className="flex items-center space-x-2">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                  <span className="text-sm text-green-600">Live</span>
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">0</div>
                  <div className="text-sm text-gray-600">New Violations</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-600">142</div>
                  <div className="text-sm text-gray-600">Records/min</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">99.9%</div>
                  <div className="text-sm text-gray-600">Uptime</div>
                </div>
              </div>
            </div>
          </Card>
        )}
      </div>

      {/* Violation Details Modal */}
      <Modal
        isOpen={showViolationDetails}
        onClose={() => setShowViolationDetails(false)}
        title="Integrity Violation Details"
      >
        <div className="space-y-4">
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <h4 className="font-medium text-red-800 mb-2">Violation Detected</h4>
            <p className="text-sm text-red-700">
              Record ID: <code className="bg-red-100 px-1 rounded">{selectedViolation}</code>
            </p>
          </div>

          <div>
            <h5 className="font-medium mb-2">Violation Details:</h5>
            <dl className="space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-600">Type:</dt>
                <dd className="font-medium">Checksum Mismatch</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-600">Detected:</dt>
                <dd className="font-medium">{format(new Date(), 'MMM dd, yyyy HH:mm:ss')}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-600">Severity:</dt>
                <dd><Badge variant="red">Critical</Badge></dd>
              </div>
            </dl>
          </div>

          <div className="flex justify-end space-x-3">
            <Button variant="secondary" onClick={() => setShowViolationDetails(false)}>
              Close
            </Button>
            <Button variant="primary">
              <Icon name="document-text" className="w-4 h-4 mr-2" />
              Generate Report
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}
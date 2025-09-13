import { useState, useEffect } from 'react';
import { ComplianceStatus, ComplianceFinding, AuditSummary } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';
import { formatDistanceToNow, format } from 'date-fns';

interface ComplianceDashboardProps {
  complianceStatuses: ComplianceStatus[];
  auditSummary: AuditSummary;
  recentFindings: ComplianceFinding[];
  onGenerateReport: (standard: string, dateRange: { start: string; end: string }) => Promise<void>;
  onUpdateFinding: (findingId: string, updates: Partial<ComplianceFinding>) => Promise<void>;
  onRefreshData: () => Promise<void>;
  loading?: boolean;
  className?: string;
}

export function ComplianceDashboard({
  complianceStatuses,
  auditSummary,
  recentFindings,
  onGenerateReport,
  onUpdateFinding,
  onRefreshData,
  loading = false,
  className = '',
}: ComplianceDashboardProps) {
  const [selectedStandard, setSelectedStandard] = useState<ComplianceStatus | null>(null);
  const [showStandardDetails, setShowStandardDetails] = useState(false);
  const [showReportGenerator, setShowReportGenerator] = useState(false);
  const [reportConfig, setReportConfig] = useState({
    standard: '',
    startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    endDate: new Date().toISOString().split('T')[0],
  });

  useEffect(() => {
    onRefreshData();
  }, [onRefreshData]);

  const getComplianceStatusColor = (status: string) => {
    switch (status) {
      case 'compliant': return 'text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20';
      case 'warning': return 'text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/20';
      case 'non_compliant': return 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20';
      case 'unknown': return 'text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800';
      default: return 'text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800';
    }
  };

  const getComplianceStatusIcon = (status: string) => {
    switch (status) {
      case 'compliant': return '✅';
      case 'warning': return '⚠️';
      case 'non_compliant': return '❌';
      case 'unknown': return '❓';
      default: return '❓';
    }
  };

  const getStandardBadgeColor = (standard: string) => {
    const colors = {
      'SOX': 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
      'HIPAA': 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
      'SOC2': 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      'GDPR': 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      'PCI-DSS': 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
      'ISO27001': 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200',
    };
    return colors[standard as keyof typeof colors] || colors.SOX;
  };

  const getFindingSeverityColor = (severity: string) => {
    switch (severity) {
      case 'low': return 'text-green-600 dark:text-green-400';
      case 'medium': return 'text-yellow-600 dark:text-yellow-400';
      case 'high': return 'text-orange-600 dark:text-orange-400';
      case 'critical': return 'text-red-600 dark:text-red-400';
      default: return 'text-gray-600 dark:text-gray-400';
    }
  };

  const getFindingTypeIcon = (type: string) => {
    switch (type) {
      case 'violation': return '🚨';
      case 'risk': return '⚠️';
      case 'recommendation': return '💡';
      default: return '📝';
    }
  };

  const getOverallComplianceScore = () => {
    if (complianceStatuses.length === 0) return 0;
    const totalScore = complianceStatuses.reduce((sum, status) => sum + status.score, 0);
    return Math.round(totalScore / complianceStatuses.length);
  };

  const getCriticalFindings = () => {
    return recentFindings.filter(finding => 
      finding.severity === 'critical' && finding.status === 'open'
    ).length;
  };

  const getUpcomingAssessments = () => {
    return complianceStatuses.filter(status => {
      const nextAssessment = new Date(status.nextAssessment);
      const thirtyDaysFromNow = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000);
      return nextAssessment <= thirtyDaysFromNow;
    });
  };

  const handleGenerateReport = async () => {
    await onGenerateReport(reportConfig.standard, {
      start: reportConfig.startDate,
      end: reportConfig.endDate,
    });
    setShowReportGenerator(false);
  };

  const overallScore = getOverallComplianceScore();
  const criticalFindings = getCriticalFindings();
  const upcomingAssessments = getUpcomingAssessments();

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className={`p-2 rounded-lg ${
              overallScore >= 90 ? 'bg-green-100 dark:bg-green-900' :
              overallScore >= 75 ? 'bg-yellow-100 dark:bg-yellow-900' :
              'bg-red-100 dark:bg-red-900'
            }`}>
              <span className={`text-xl ${
                overallScore >= 90 ? 'text-green-600 dark:text-green-400' :
                overallScore >= 75 ? 'text-yellow-600 dark:text-yellow-400' :
                'text-red-600 dark:text-red-400'
              }`}>
                📊
              </span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {overallScore}%
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Overall Compliance</div>
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
                {criticalFindings}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Critical Findings</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
              <span className="text-blue-600 dark:text-blue-400 text-xl">📋</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {complianceStatuses.length}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Standards Tracked</div>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
              <span className="text-yellow-600 dark:text-yellow-400 text-xl">📅</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {upcomingAssessments.length}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Upcoming Assessments</div>
            </div>
          </div>
        </Card>
      </div>

      {/* Compliance Standards Overview */}
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
              Compliance Standards
            </h2>
            <div className="flex space-x-3">
              <Button
                onClick={() => setShowReportGenerator(true)}
                variant="secondary"
                size="sm"
              >
                📊 Generate Report
              </Button>
              <Button
                onClick={onRefreshData}
                variant="secondary"
                size="sm"
                loading={loading}
              >
                🔄 Refresh
              </Button>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {complianceStatuses.map((status) => (
              <div
                key={status.id}
                className={`border-2 rounded-lg p-4 cursor-pointer transition-colors hover:shadow-lg ${getComplianceStatusColor(status.status)}`}
                onClick={() => {
                  setSelectedStandard(status);
                  setShowStandardDetails(true);
                }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className={`px-3 py-1 text-sm font-medium rounded-full ${getStandardBadgeColor(status.standard)}`}>
                    {status.standard}
                  </span>
                  <span className="text-2xl">{getComplianceStatusIcon(status.status)}</span>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600 dark:text-gray-400">Score</span>
                    <span className="text-lg font-bold text-gray-900 dark:text-white">
                      {status.score}%
                    </span>
                  </div>

                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className={`h-2 rounded-full ${
                        status.score >= 90 ? 'bg-green-500' :
                        status.score >= 75 ? 'bg-yellow-500' :
                        'bg-red-500'
                      }`}
                      style={{ width: `${status.score}%` }}
                    ></div>
                  </div>

                  <div className="space-y-1 text-xs text-gray-600 dark:text-gray-400">
                    <div>Last assessed: {formatDistanceToNow(new Date(status.lastAssessed), { addSuffix: true })}</div>
                    <div>Next assessment: {format(new Date(status.nextAssessment), 'MMM dd, yyyy')}</div>
                    <div>
                      {status.findings.filter(f => f.status === 'open').length} open findings
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Card>

      {/* Recent Findings */}
      <Card>
        <div className="p-6">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">
            Recent Compliance Findings
          </h2>

          <div className="space-y-4">
            {recentFindings.slice(0, 10).map((finding) => (
              <div
                key={finding.id}
                className="border border-gray-200 dark:border-gray-700 rounded-lg p-4"
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center space-x-3">
                    <span className="text-xl">{getFindingTypeIcon(finding.type)}</span>
                    <div>
                      <h3 className="font-medium text-gray-900 dark:text-white">
                        {finding.title}
                      </h3>
                      <div className="flex items-center space-x-2 mt-1">
                        <span className={`text-sm font-medium ${getFindingSeverityColor(finding.severity)}`}>
                          {finding.severity.toUpperCase()}
                        </span>
                        <span className="text-sm text-gray-500 dark:text-gray-400">•</span>
                        <span className="text-sm text-gray-500 dark:text-gray-400">
                          {finding.type.replace('_', ' ').toUpperCase()}
                        </span>
                        {finding.requirement && (
                          <>
                            <span className="text-sm text-gray-500 dark:text-gray-400">•</span>
                            <span className="text-sm text-gray-500 dark:text-gray-400">
                              {finding.requirement}
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${
                      finding.status === 'open' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' :
                      finding.status === 'in_progress' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' :
                      finding.status === 'resolved' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                      'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                    }`}>
                      {finding.status.replace('_', ' ').toUpperCase()}
                    </span>
                    
                    {finding.dueDate && (
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        Due: {format(new Date(finding.dueDate), 'MMM dd')}
                      </span>
                    )}
                  </div>
                </div>

                <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
                  {finding.description}
                </p>

                {finding.remediation && (
                  <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded text-sm">
                    <div className="font-medium text-blue-900 dark:text-blue-200 mb-1">
                      Recommended Action:
                    </div>
                    <div className="text-blue-800 dark:text-blue-300">
                      {finding.remediation}
                    </div>
                  </div>
                )}

                <div className="flex items-center justify-between mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    Created {formatDistanceToNow(new Date(finding.createdAt), { addSuffix: true })}
                    {finding.assignedTo && ` • Assigned to ${finding.assignedTo}`}
                  </div>
                  
                  {finding.status === 'open' && (
                    <Button
                      onClick={() => onUpdateFinding(finding.id, { status: 'in_progress' })}
                      variant="primary"
                      size="sm"
                    >
                      Start Remediation
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>

          {recentFindings.length === 0 && (
            <div className="text-center py-8">
              <div className="text-gray-500 dark:text-gray-400">
                <span className="text-4xl mb-4 block">✅</span>
                <p className="text-lg font-medium mb-2">No recent findings</p>
                <p className="text-sm">Your compliance posture is looking good!</p>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Audit Activity Summary */}
      <Card>
        <div className="p-6">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-6">
            Audit Activity Summary
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-4">
              <h3 className="font-medium text-gray-900 dark:text-white">Event Breakdown</h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600 dark:text-gray-400">Total Events</span>
                  <span className="font-medium text-gray-900 dark:text-white">
                    {auditSummary.totalEvents.toLocaleString()}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600 dark:text-gray-400">Success Rate</span>
                  <span className="font-medium text-green-600 dark:text-green-400">
                    {Math.round((auditSummary.successEvents / auditSummary.totalEvents) * 100)}%
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600 dark:text-gray-400">Critical Events</span>
                  <span className="font-medium text-red-600 dark:text-red-400">
                    {auditSummary.criticalEvents}
                  </span>
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="font-medium text-gray-900 dark:text-white">Top Users</h3>
              <div className="space-y-2">
                {auditSummary.topUsers.slice(0, 5).map((user, index) => (
                  <div key={user.userId} className="flex items-center justify-between">
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      {index + 1}. {user.userName || user.userId}
                    </span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {user.eventCount}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="font-medium text-gray-900 dark:text-white">Top Actions</h3>
              <div className="space-y-2">
                {auditSummary.topActions.slice(0, 5).map((action, index) => (
                  <div key={action.action} className="flex items-center justify-between">
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      {index + 1}. {action.action}
                    </span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {action.eventCount}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Standard Details Modal */}
      <Modal
        isOpen={showStandardDetails}
        onClose={() => setShowStandardDetails(false)}
        title={`${selectedStandard?.standard} Compliance Details`}
      >
        {selectedStandard && (
          <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Overall Status
                </label>
                <div className="flex items-center space-x-2">
                  <span className="text-xl">{getComplianceStatusIcon(selectedStandard.status)}</span>
                  <span className="capitalize font-medium text-gray-900 dark:text-white">
                    {selectedStandard.status.replace('_', ' ')}
                  </span>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Compliance Score
                </label>
                <div className="text-2xl font-bold text-gray-900 dark:text-white">
                  {selectedStandard.score}%
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Last Assessment
                </label>
                <div className="text-gray-900 dark:text-white">
                  {format(new Date(selectedStandard.lastAssessed), 'MMM dd, yyyy')}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Next Assessment
                </label>
                <div className="text-gray-900 dark:text-white">
                  {format(new Date(selectedStandard.nextAssessment), 'MMM dd, yyyy')}
                </div>
              </div>
            </div>

            <div>
              <h3 className="font-medium text-gray-900 dark:text-white mb-3">
                Requirements Status ({selectedStandard.requirements.length})
              </h3>
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {selectedStandard.requirements.map((req) => (
                  <div key={req.id} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded">
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-gray-900 dark:text-white text-sm">
                        {req.requirement}
                      </div>
                      <div className="text-xs text-gray-600 dark:text-gray-400 truncate">
                        {req.description}
                      </div>
                    </div>
                    <span className={`px-2 py-1 text-xs font-medium rounded ml-3 ${
                      req.status === 'met' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' :
                      req.status === 'partially_met' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' :
                      req.status === 'not_met' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' :
                      'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                    }`}>
                      {req.status.replace('_', ' ').toUpperCase()}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <h3 className="font-medium text-gray-900 dark:text-white mb-3">
                Active Findings ({selectedStandard.findings.filter(f => f.status === 'open').length})
              </h3>
              <div className="space-y-2 max-h-40 overflow-y-auto">
                {selectedStandard.findings.filter(f => f.status === 'open').map((finding) => (
                  <div key={finding.id} className="p-3 bg-gray-50 dark:bg-gray-800 rounded">
                    <div className="flex items-center justify-between">
                      <div className="font-medium text-gray-900 dark:text-white text-sm">
                        {finding.title}
                      </div>
                      <span className={`text-xs font-medium ${getFindingSeverityColor(finding.severity)}`}>
                        {finding.severity.toUpperCase()}
                      </span>
                    </div>
                    <div className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                      {finding.description}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="flex space-x-3">
              <Button
                onClick={() => {
                  setReportConfig(prev => ({ ...prev, standard: selectedStandard.standard }));
                  setShowStandardDetails(false);
                  setShowReportGenerator(true);
                }}
                variant="primary"
                className="flex-1"
              >
                Generate Report
              </Button>
              <Button
                onClick={() => setShowStandardDetails(false)}
                variant="secondary"
                className="flex-1"
              >
                Close
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Report Generator Modal */}
      <Modal
        isOpen={showReportGenerator}
        onClose={() => setShowReportGenerator(false)}
        title="Generate Compliance Report"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Compliance Standard
            </label>
            <select
              value={reportConfig.standard}
              onChange={(e) => setReportConfig(prev => ({ ...prev, standard: e.target.value }))}
              className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
            >
              <option value="">All Standards</option>
              {complianceStatuses.map((status) => (
                <option key={status.standard} value={status.standard}>
                  {status.standard}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Start Date
              </label>
              <input
                type="date"
                value={reportConfig.startDate}
                onChange={(e) => setReportConfig(prev => ({ ...prev, startDate: e.target.value }))}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                End Date
              </label>
              <input
                type="date"
                value={reportConfig.endDate}
                onChange={(e) => setReportConfig(prev => ({ ...prev, endDate: e.target.value }))}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>
          </div>

          <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
            <p className="text-sm text-blue-700 dark:text-blue-300">
              The report will include compliance status, findings, remediation progress, 
              and audit trail data for the selected period.
            </p>
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleGenerateReport}
              variant="primary"
              className="flex-1"
            >
              Generate Report
            </Button>
            <Button
              onClick={() => setShowReportGenerator(false)}
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
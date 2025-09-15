import React, { useState, useEffect } from 'react';
import { ComplianceReport, ComplianceFramework } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Select } from '../ui/Select';
import { Badge } from '../ui/Badge';
import { Icon } from '../ui/Icon';
import { Modal } from '../ui/Modal';
import { format } from 'date-fns';

interface ComplianceReportManagerProps {
  className?: string;
}

export function ComplianceReportManager({ className = '' }: ComplianceReportManagerProps) {
  const [reports, setReports] = useState<ComplianceReport[]>([]);
  const [selectedFramework, setSelectedFramework] = useState<ComplianceFramework>('sox');
  const [selectedDateRange, setSelectedDateRange] = useState({
    startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
    endDate: new Date().toISOString()
  });
  const [isGenerating, setIsGenerating] = useState(false);
  const [showReportDetails, setShowReportDetails] = useState(false);
  const [selectedReport, setSelectedReport] = useState<ComplianceReport | null>(null);
  const [loading, setLoading] = useState(false);

  // Load existing reports
  const loadReports = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/audit/compliance-reports');
      if (response.ok) {
        const data = await response.json();
        setReports(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load compliance reports:', error);
    } finally {
      setLoading(false);
    }
  };

  // Generate new compliance report
  const generateReport = async () => {
    setIsGenerating(true);
    try {
      const response = await fetch('/api/v1/audit/compliance-reports', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          framework: selectedFramework,
          startTime: selectedDateRange.startDate,
          endTime: selectedDateRange.endDate,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        setReports(prev => [data.data, ...prev]);
      }
    } catch (error) {
      console.error('Failed to generate compliance report:', error);
    } finally {
      setIsGenerating(false);
    }
  };

  useEffect(() => {
    loadReports();
  }, []);

  const getFrameworkColor = (framework: ComplianceFramework) => {
    switch (framework) {
      case 'sox': return 'blue';
      case 'hipaa': return 'green';
      case 'soc2': return 'blue';
      default: return 'gray';
    }
  };

  const getComplianceScoreColor = (score: number) => {
    if (score >= 95) return 'text-green-600';
    if (score >= 80) return 'text-yellow-600';
    if (score >= 60) return 'text-orange-600';
    return 'text-red-600';
  };

  const getComplianceScoreBadge = (score: number) => {
    if (score >= 95) return 'green';
    if (score >= 80) return 'yellow';
    if (score >= 60) return 'orange';
    return 'red';
  };

  return (
    <>
      <div className={`space-y-6 ${className}`}>
        {/* Report Generation */}
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-medium mb-4 flex items-center">
              <Icon name="document-text" className="w-5 h-5 mr-2" />
              Generate Compliance Report
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Compliance Framework
                </label>
                <Select
                  value={selectedFramework}
                  onChange={(e) => setSelectedFramework(e.target.value as ComplianceFramework)}
                  className="w-full"
                >
                  <option value="sox">Sarbanes-Oxley (SOX)</option>
                  <option value="hipaa">Health Insurance Portability (HIPAA)</option>
                  <option value="soc2">Service Organization Control 2 (SOC2)</option>
                </Select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Start Date
                </label>
                <input
                  type="date"
                  value={selectedDateRange.startDate.split('T')[0]}
                  onChange={(e) => setSelectedDateRange(prev => ({
                    ...prev,
                    startDate: new Date(e.target.value).toISOString()
                  }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  End Date
                </label>
                <input
                  type="date"
                  value={selectedDateRange.endDate.split('T')[0]}
                  onChange={(e) => setSelectedDateRange(prev => ({
                    ...prev,
                    endDate: new Date(e.target.value).toISOString()
                  }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>

            <Button
              onClick={generateReport}
              disabled={isGenerating}
              className="w-full md:w-auto"
            >
              {isGenerating ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Generating Report...
                </>
              ) : (
                <>
                  <Icon name="document-plus" className="w-4 h-4 mr-2" />
                  Generate {selectedFramework.toUpperCase()} Report
                </>
              )}
            </Button>
          </div>
        </Card>

        {/* Reports List */}
        <Card>
          <div className="p-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium">Recent Compliance Reports</h3>
              <Button
                variant="secondary"
                size="sm"
                onClick={loadReports}
                disabled={loading}
              >
                <Icon name="refresh" className="w-4 h-4 mr-2" />
                Refresh
              </Button>
            </div>

            {loading ? (
              <div className="flex justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : reports.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Icon name="document-text" className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>No compliance reports generated yet.</p>
                <p className="text-sm">Generate your first report above.</p>
              </div>
            ) : (
              <div className="space-y-4">
                {reports.map((report) => (
                  <div
                    key={report.id}
                    className="border border-gray-200 rounded-md p-4 hover:bg-gray-50 cursor-pointer"
                    onClick={() => {
                      setSelectedReport(report);
                      setShowReportDetails(true);
                    }}
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3 mb-2">
                          <Badge variant={getFrameworkColor(report.framework)}>
                            {report.framework.toUpperCase()}
                          </Badge>
                          <Badge variant={getComplianceScoreBadge(report.complianceScore)}>
                            {report.complianceScore.toFixed(1)}% Compliance
                          </Badge>
                          <span className="text-sm text-gray-500">
                            {format(new Date(report.generatedAt), 'MMM dd, yyyy HH:mm')}
                          </span>
                        </div>

                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                          <div>
                            <span className="text-gray-600">Violations:</span>
                            <span className={`ml-2 font-medium ${report.violations.length > 0 ? 'text-red-600' : 'text-green-600'}`}>
                              {report.violations.length}
                            </span>
                          </div>
                          <div>
                            <span className="text-gray-600">Recommendations:</span>
                            <span className="ml-2 font-medium text-blue-600">
                              {report.recommendations.length}
                            </span>
                          </div>
                          <div>
                            <span className="text-gray-600">Period:</span>
                            <span className="ml-2 font-medium">
                              {Math.ceil((new Date(report.periodEnd).getTime() - new Date(report.periodStart).getTime()) / (1000 * 60 * 60 * 24))} days
                            </span>
                          </div>
                          <div>
                            <span className="text-gray-600">Status:</span>
                            <span className={`ml-2 font-medium ${report.complianceScore >= 80 ? 'text-green-600' : 'text-red-600'}`}>
                              {report.complianceScore >= 80 ? 'Compliant' : 'Non-Compliant'}
                            </span>
                          </div>
                        </div>
                      </div>

                      <Icon name="chevron-right" className="w-5 h-5 text-gray-400" />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Report Details Modal */}
      <Modal
        isOpen={showReportDetails}
        onClose={() => setShowReportDetails(false)}
        title={`${selectedReport?.framework.toUpperCase()} Compliance Report`}
        size="lg"
      >
        {selectedReport && (
          <div className="space-y-6">
            {/* Report Summary */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="text-center">
                <div className={`text-3xl font-bold ${getComplianceScoreColor(selectedReport.complianceScore)}`}>
                  {selectedReport.complianceScore.toFixed(1)}%
                </div>
                <div className="text-sm text-gray-600">Compliance Score</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-red-600">
                  {selectedReport.violations.length}
                </div>
                <div className="text-sm text-gray-600">Violations</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-blue-600">
                  {selectedReport.recommendations.length}
                </div>
                <div className="text-sm text-gray-600">Recommendations</div>
              </div>
            </div>

            {/* Violations */}
            {selectedReport.violations.length > 0 && (
              <div>
                <h4 className="font-medium mb-3 text-red-600">Compliance Violations</h4>
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {selectedReport.violations.map((violation, index) => (
                    <div key={index} className="bg-red-50 border border-red-200 rounded-md p-3">
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="font-medium text-red-800">{violation.type}</div>
                          <div className="text-sm text-red-700">{violation.description}</div>
                        </div>
                        <Badge variant="red" className="ml-2">
                          {violation.severity}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Recommendations */}
            {selectedReport.recommendations.length > 0 && (
              <div>
                <h4 className="font-medium mb-3 text-blue-600">Recommendations</h4>
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {selectedReport.recommendations.map((recommendation, index) => (
                    <div key={index} className="bg-blue-50 border border-blue-200 rounded-md p-3">
                      <div className="text-sm text-blue-800">{recommendation}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Actions */}
            <div className="flex justify-end space-x-3">
              <Button variant="secondary" onClick={() => setShowReportDetails(false)}>
                Close
              </Button>
              <Button variant="primary">
                <Icon name="download" className="w-4 h-4 mr-2" />
                Download Report
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}
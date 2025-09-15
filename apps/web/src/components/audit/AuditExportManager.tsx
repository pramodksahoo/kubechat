import React, { useState } from 'react';
import { AuditFilter, ExportFormat } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Select } from '../ui/Select';
import { Icon } from '../ui/Icon';
import { Modal } from '../ui/Modal';

interface AuditExportManagerProps {
  filters: AuditFilter;
  onExport: (format: ExportFormat, filters: AuditFilter) => Promise<void>;
  className?: string;
}

export function AuditExportManager({
  filters,
  onExport,
  className = ''
}: AuditExportManagerProps) {
  const [exportFormat, setExportFormat] = useState<ExportFormat>('json');
  const [showExportModal, setShowExportModal] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [exportProgress, setExportProgress] = useState(0);

  const handleExport = async () => {
    setIsExporting(true);
    setExportProgress(0);

    try {
      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setExportProgress((prev) => Math.min(prev + 10, 90));
      }, 100);

      await onExport(exportFormat, filters);

      clearInterval(progressInterval);
      setExportProgress(100);

      setTimeout(() => {
        setShowExportModal(false);
        setIsExporting(false);
        setExportProgress(0);
      }, 1000);
    } catch (error) {
      console.error('Export failed:', error);
      setIsExporting(false);
      setExportProgress(0);
    }
  };

  const getFormatIcon = (format: ExportFormat) => {
    switch (format) {
      case 'csv': return 'table';
      case 'json': return 'code';
      case 'pdf': return 'document';
      default: return 'document';
    }
  };

  const getFormatDescription = (format: ExportFormat) => {
    switch (format) {
      case 'csv': return 'Comma-separated values, ideal for spreadsheet analysis';
      case 'json': return 'JavaScript Object Notation, structured data format';
      case 'pdf': return 'Portable Document Format, suitable for reports and archival';
      default: return '';
    }
  };

  return (
    <>
      <Card className={className}>
        <div className="p-6">
          <h3 className="text-lg font-medium mb-4 flex items-center">
            <Icon name="download" className="w-5 h-5 mr-2" />
            Export Audit Logs
          </h3>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Export Format
              </label>
              <Select
                value={exportFormat}
                onChange={(e) => setExportFormat(e.target.value as ExportFormat)}
                className="w-full"
              >
                <option value="json">JSON</option>
                <option value="csv">CSV</option>
                <option value="pdf">PDF</option>
              </Select>
              <p className="text-xs text-gray-500 mt-1">
                {getFormatDescription(exportFormat)}
              </p>
            </div>

            <Button
              onClick={() => setShowExportModal(true)}
              className="w-full"
              disabled={isExporting}
            >
              <Icon name={getFormatIcon(exportFormat)} className="w-4 h-4 mr-2" />
              Export as {exportFormat.toUpperCase()}
            </Button>

            {/* Export Statistics */}
            <div className="mt-4 p-3 bg-gray-50 rounded-md">
              <h4 className="text-sm font-medium text-gray-900 mb-2">Export Summary:</h4>
              <div className="text-xs text-gray-600 space-y-1">
                <div>Format: {exportFormat.toUpperCase()}</div>
                <div>Time Range: {filters.startDate && filters.endDate ?
                  `${new Date(filters.startDate).toLocaleDateString()} - ${new Date(filters.endDate).toLocaleDateString()}` :
                  'All time'
                }</div>
                {filters.userId && <div>User: {filters.userId}</div>}
                {filters.action && <div>Action: {filters.action}</div>}
                {filters.severity && <div>Severity: {filters.severity}</div>}
              </div>
            </div>

            {/* Quick Export Options */}
            <div className="border-t pt-4">
              <h4 className="text-sm font-medium text-gray-900 mb-3">Quick Export:</h4>
              <div className="space-y-2">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    setExportFormat('csv');
                    handleExport();
                  }}
                  className="w-full text-left"
                  disabled={isExporting}
                >
                  <Icon name="table" className="w-4 h-4 mr-2" />
                  Export Current View as CSV
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    setExportFormat('pdf');
                    handleExport();
                  }}
                  className="w-full text-left"
                  disabled={isExporting}
                >
                  <Icon name="document" className="w-4 h-4 mr-2" />
                  Generate PDF Report
                </Button>
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Export Confirmation Modal */}
      <Modal
        isOpen={showExportModal}
        onClose={() => setShowExportModal(false)}
        title="Export Audit Logs"
      >
        <div className="space-y-4">
          <div className="flex items-center space-x-3">
            <Icon name={getFormatIcon(exportFormat)} className="w-8 h-8 text-blue-600" />
            <div>
              <h4 className="text-lg font-medium">Export as {exportFormat.toUpperCase()}</h4>
              <p className="text-sm text-gray-600">{getFormatDescription(exportFormat)}</p>
            </div>
          </div>

          <div className="bg-gray-50 p-4 rounded-md">
            <h5 className="font-medium mb-2">Export Details:</h5>
            <dl className="space-y-1 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-600">Format:</dt>
                <dd className="font-medium">{exportFormat.toUpperCase()}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-600">Time Range:</dt>
                <dd className="font-medium">
                  {filters.startDate && filters.endDate ?
                    `${new Date(filters.startDate).toLocaleDateString()} - ${new Date(filters.endDate).toLocaleDateString()}` :
                    'All time'
                  }
                </dd>
              </div>
              {filters.userId && (
                <div className="flex justify-between">
                  <dt className="text-gray-600">User Filter:</dt>
                  <dd className="font-medium">{filters.userId}</dd>
                </div>
              )}
              {filters.action && (
                <div className="flex justify-between">
                  <dt className="text-gray-600">Action Filter:</dt>
                  <dd className="font-medium">{filters.action}</dd>
                </div>
              )}
              {filters.severity && (
                <div className="flex justify-between">
                  <dt className="text-gray-600">Severity Filter:</dt>
                  <dd className="font-medium">{filters.severity}</dd>
                </div>
              )}
            </dl>
          </div>

          {isExporting && (
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Exporting...</span>
                <span>{exportProgress}%</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${exportProgress}%` }}
                ></div>
              </div>
            </div>
          )}

          <div className="flex justify-end space-x-3">
            <Button
              variant="secondary"
              onClick={() => setShowExportModal(false)}
              disabled={isExporting}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleExport}
              disabled={isExporting}
            >
              {isExporting ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Exporting...
                </>
              ) : (
                <>
                  <Icon name="download" className="w-4 h-4 mr-2" />
                  Start Export
                </>
              )}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}
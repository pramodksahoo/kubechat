import { useState } from 'react';
import { CommandPreview } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';

interface CommandPreviewCardProps {
  preview: CommandPreview;
  onApprove: () => void;
  onReject: (reason: string) => void;
  className?: string;
}

export function CommandPreviewCard({ 
  preview, 
  onApprove, 
  onReject, 
  className = '' 
}: CommandPreviewCardProps) {
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [rejectReason, setRejectReason] = useState('');

  const getSafetyColor = () => {
    switch (preview.safetyLevel) {
      case 'safe': return 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20';
      case 'warning': return 'border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-900/20';
      case 'dangerous': return 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/20';
      default: return 'border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800';
    }
  };

  const getSafetyIcon = () => {
    switch (preview.safetyLevel) {
      case 'safe':
        return (
          <div className="flex items-center space-x-2 text-green-600 dark:text-green-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">Safe Operation</span>
          </div>
        );
      case 'warning':
        return (
          <div className="flex items-center space-x-2 text-yellow-600 dark:text-yellow-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">Potentially Risky</span>
          </div>
        );
      case 'dangerous':
        return (
          <div className="flex items-center space-x-2 text-red-600 dark:text-red-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">Dangerous Operation</span>
          </div>
        );
      default:
        return null;
    }
  };

  const getConfidenceColor = () => {
    if (preview.confidence >= 0.8) return 'text-green-600 dark:text-green-400';
    if (preview.confidence >= 0.6) return 'text-yellow-600 dark:text-yellow-400';
    return 'text-red-600 dark:text-red-400';
  };

  const handleReject = () => {
    onReject(rejectReason);
    setShowRejectModal(false);
    setRejectReason('');
  };

  return (
    <>
      <Card className={`${getSafetyColor()} border-2 ${className}`}>
        <div className="p-4">
          {/* Header */}
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-3">
              {getSafetyIcon()}
              <div className={`text-sm font-medium ${getConfidenceColor()}`}>
                {Math.round(preview.confidence * 100)}% confidence
              </div>
            </div>
            {preview.approvalRequired && (
              <div className="flex items-center space-x-1 text-orange-600 dark:text-orange-400">
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
                </svg>
                <span className="text-xs font-medium">Approval Required</span>
              </div>
            )}
          </div>

          {/* Natural Language */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Your Request:
            </label>
            <p className="text-gray-900 dark:text-white bg-white dark:bg-gray-800 p-3 rounded-lg border">
              {preview.naturalLanguage}
            </p>
          </div>

          {/* Generated Command */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Generated Command:
            </label>
            <div className="bg-gray-900 text-green-400 p-3 rounded-lg font-mono text-sm overflow-x-auto">
              <code>{preview.generatedCommand}</code>
            </div>
          </div>

          {/* Explanation */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              What this does:
            </label>
            <p className="text-gray-700 dark:text-gray-300 text-sm">
              {preview.explanation}
            </p>
          </div>

          {/* Impact Assessment */}
          {preview.potentialImpact.length > 0 && (
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Potential Impact:
              </label>
              <ul className="space-y-1">
                {preview.potentialImpact.map((impact, index) => (
                  <li key={index} className="flex items-start space-x-2 text-sm">
                    <span className="text-gray-400 mt-1.5 w-1 h-1 bg-current rounded-full flex-shrink-0"></span>
                    <span className="text-gray-700 dark:text-gray-300">{impact}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Required Permissions */}
          {preview.requiredPermissions.length > 0 && (
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Required Permissions:
              </label>
              <div className="flex flex-wrap gap-2">
                {preview.requiredPermissions.map((permission, index) => (
                  <span 
                    key={index}
                    className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                  >
                    {permission}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex space-x-3">
            <Button
              onClick={onApprove}
              variant="primary"
              className="flex-1"
              disabled={preview.safetyLevel === 'dangerous' && preview.approvalRequired}
            >
              <div className="flex items-center justify-center space-x-2">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                <span>
                  {preview.approvalRequired ? 'Request Approval' : 'Execute Command'}
                </span>
              </div>
            </Button>

            <Button
              onClick={() => setShowRejectModal(true)}
              variant="secondary"
              className="flex-1"
            >
              <div className="flex items-center justify-center space-x-2">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
                <span>Cancel</span>
              </div>
            </Button>
          </div>
        </div>
      </Card>

      {/* Reject Modal */}
      <Modal
        isOpen={showRejectModal}
        onClose={() => setShowRejectModal(false)}
        title="Cancel Command"
      >
        <div className="space-y-4">
          <p className="text-gray-700 dark:text-gray-300">
            Why are you canceling this command? (Optional)
          </p>
          
          <textarea
            value={rejectReason}
            onChange={(e) => setRejectReason(e.target.value)}
            placeholder="Enter reason for cancellation..."
            rows={3}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
          />

          <div className="flex space-x-3">
            <Button
              onClick={handleReject}
              variant="primary"
              className="flex-1"
            >
              Confirm Cancel
            </Button>
            <Button
              onClick={() => setShowRejectModal(false)}
              variant="secondary"
              className="flex-1"
            >
              Keep Command
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}
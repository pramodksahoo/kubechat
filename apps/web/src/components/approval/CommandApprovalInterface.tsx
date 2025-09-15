import { useState, useEffect } from 'react';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';
import { StatusIndicator } from '../ui/StatusIndicator';
import { commandApprovalService, ApprovalRequest, ApprovalStep } from '@/services/commandApprovalService';

interface CommandApprovalInterfaceProps {
  className?: string;
}

export function CommandApprovalInterface({ className = '' }: CommandApprovalInterfaceProps) {
  const [pendingRequests, setPendingRequests] = useState<ApprovalRequest[]>([]);
  const [selectedRequest, setSelectedRequest] = useState<ApprovalRequest | null>(null);
  const [showApprovalModal, setShowApprovalModal] = useState(false);
  const [showRejectionModal, setShowRejectionModal] = useState(false);
  const [rejectionReason, setRejectionReason] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Load initial pending requests
    loadPendingRequests();

    // Subscribe to approval updates
    const unsubscribe = commandApprovalService.onApprovalUpdate(() => {
      loadPendingRequests();
    });

    return unsubscribe;
  }, []);

  const loadPendingRequests = async () => {
    try {
      const requests = await commandApprovalService.getPendingApprovalRequests();
      setPendingRequests(requests);
    } catch (error) {
      console.error('Failed to load pending requests:', error);
    }
  };

  const handleApproveRequest = async (requestId: string, stepId?: string) => {
    try {
      setLoading(true);
      const currentUserId = 'current-user'; // In real app, get from auth context

      await commandApprovalService.approveRequest(requestId, currentUserId, stepId);

      if (!stepId) {
        setShowApprovalModal(false);
        setSelectedRequest(null);
      }
    } catch (error) {
      console.error('Failed to approve request:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRejectRequest = async () => {
    if (!selectedRequest) return;

    try {
      setLoading(true);
      const currentUserId = 'current-user'; // In real app, get from auth context

      await commandApprovalService.rejectRequest(selectedRequest.id, currentUserId, rejectionReason);
      setShowRejectionModal(false);
      setShowApprovalModal(false);
      setSelectedRequest(null);
      setRejectionReason('');
    } catch (error) {
      console.error('Failed to reject request:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStepStatusIcon = (step: ApprovalStep) => {
    switch (step.status) {
      case 'completed':
        return (
          <svg className="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
          </svg>
        );
      case 'pending':
        return (
          <svg className="w-5 h-5 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
          </svg>
        );
      default:
        return (
          <svg className="w-5 h-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
        );
    }
  };

  const getSafetyLevelBadge = (safetyLevel: string) => {
    const configs = {
      safe: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      dangerous: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
    };

    const config = configs[safetyLevel as keyof typeof configs] || configs.safe;

    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config}`}>
        {safetyLevel.toUpperCase()}
      </span>
    );
  };

  const formatTimeAgo = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diffMs = now.getTime() - time.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffMins > 0) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  if (pendingRequests.length === 0) {
    return (
      <Card className={`p-6 text-center ${className}`}>
        <div className="flex flex-col items-center space-y-4">
          <svg className="w-16 h-16 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white">No Pending Approvals</h3>
            <p className="text-gray-500 dark:text-gray-400">All command approval requests have been processed.</p>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <div className={`space-y-4 ${className}`}>
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
          Command Approval Requests
        </h2>
        <StatusIndicator
          status={pendingRequests.length > 0 ? 'warning' : 'healthy'}
          label={`${pendingRequests.length} pending`}
          size="sm"
        />
      </div>

      <div className="space-y-4">
        {pendingRequests.map((request) => (
          <Card key={request.id} className="p-4">
            <div className="space-y-4">
              {/* Request Header */}
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-3 mb-2">
                    <h3 className="text-base font-medium text-gray-900 dark:text-white">
                      Command Approval Request
                    </h3>
                    {getSafetyLevelBadge(request.commandPreview.safetyLevel)}
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                    <p>Requested by: {request.requestedBy}</p>
                    <p>Requested: {formatTimeAgo(request.requestedAt)}</p>
                    <p>Expires: {formatTimeAgo(request.expiresAt)}</p>
                  </div>
                </div>
                <Button
                  onClick={() => {
                    setSelectedRequest(request);
                    setShowApprovalModal(true);
                  }}
                  variant="primary"
                  size="sm"
                >
                  Review
                </Button>
              </div>

              {/* Command Preview */}
              <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <div className="space-y-2">
                  <div>
                    <span className="text-xs font-medium text-gray-600 dark:text-gray-400">Request:</span>
                    <p className="text-sm text-gray-900 dark:text-white">{request.commandPreview.naturalLanguage}</p>
                  </div>
                  <div>
                    <span className="text-xs font-medium text-gray-600 dark:text-gray-400">Command:</span>
                    <code className="block text-xs bg-gray-900 text-green-400 p-2 rounded mt-1">
                      {request.commandPreview.generatedCommand}
                    </code>
                  </div>
                </div>
              </div>

              {/* Approval Steps */}
              {request.approvalSteps && request.approvalSteps.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Approval Steps
                  </h4>
                  <div className="space-y-2">
                    {request.approvalSteps
                      .sort((a, b) => a.order - b.order)
                      .map((step) => (
                        <div key={step.id} className="flex items-center space-x-3">
                          {getStepStatusIcon(step)}
                          <div className="flex-1">
                            <div className="flex items-center justify-between">
                              <span className="text-sm text-gray-900 dark:text-white">
                                {step.name}
                              </span>
                              {step.status === 'completed' && step.completedBy && (
                                <span className="text-xs text-gray-500 dark:text-gray-400">
                                  by {step.completedBy}
                                </span>
                              )}
                            </div>
                            <p className="text-xs text-gray-600 dark:text-gray-400">
                              {step.description}
                            </p>
                          </div>
                        </div>
                      ))}
                  </div>
                </div>
              )}
            </div>
          </Card>
        ))}
      </div>

      {/* Approval Modal */}
      {selectedRequest && (
        <Modal
          isOpen={showApprovalModal}
          onClose={() => {
            setShowApprovalModal(false);
            setSelectedRequest(null);
          }}
          title="Review Command Approval Request"
          size="lg"
        >
          <div className="space-y-6">
            {/* Request Details */}
            <div className="space-y-4">
              <div className="flex items-center space-x-3">
                <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                  Command Details
                </h3>
                {getSafetyLevelBadge(selectedRequest.commandPreview.safetyLevel)}
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="font-medium text-gray-700 dark:text-gray-300">Requested by:</span>
                  <p className="text-gray-900 dark:text-white">{selectedRequest.requestedBy}</p>
                </div>
                <div>
                  <span className="font-medium text-gray-700 dark:text-gray-300">Requested:</span>
                  <p className="text-gray-900 dark:text-white">{formatTimeAgo(selectedRequest.requestedAt)}</p>
                </div>
              </div>

              {/* Command Preview */}
              <div>
                <span className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  User Request:
                </span>
                <p className="text-gray-900 dark:text-white bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
                  {selectedRequest.commandPreview.naturalLanguage}
                </p>
              </div>

              <div>
                <span className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Generated Command:
                </span>
                <code className="block bg-gray-900 text-green-400 p-3 rounded-lg text-sm">
                  {selectedRequest.commandPreview.generatedCommand}
                </code>
              </div>

              <div>
                <span className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Explanation:
                </span>
                <p className="text-gray-700 dark:text-gray-300 text-sm">
                  {selectedRequest.commandPreview.explanation}
                </p>
              </div>

              {/* Risk Assessment for Dangerous Operations */}
              {selectedRequest.commandPreview.safetyLevel === 'dangerous' && (
                <div className="p-4 bg-red-50 dark:bg-red-900/10 border-l-4 border-red-400 rounded-r-lg">
                  <h4 className="text-red-800 dark:text-red-200 font-semibold text-sm mb-2">
                    ⚠️ HIGH RISK OPERATION
                  </h4>
                  <ul className="text-red-700 dark:text-red-300 text-sm space-y-1">
                    {selectedRequest.commandPreview.potentialImpact.map((impact, index) => (
                      <li key={index} className="flex items-start space-x-2">
                        <span className="text-red-500 mt-1.5 w-1.5 h-1.5 bg-current rounded-full flex-shrink-0"></span>
                        <span>{impact}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>

            {/* Approval Steps */}
            {selectedRequest.approvalSteps && selectedRequest.approvalSteps.length > 0 && (
              <div>
                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                  Approval Workflow
                </h4>
                <div className="space-y-3">
                  {selectedRequest.approvalSteps
                    .sort((a, b) => a.order - b.order)
                    .map((step) => (
                      <div key={step.id} className="flex items-start space-x-3 p-3 border rounded-lg">
                        {getStepStatusIcon(step)}
                        <div className="flex-1">
                          <div className="flex items-center justify-between mb-1">
                            <h5 className="text-sm font-medium text-gray-900 dark:text-white">
                              {step.name}
                            </h5>
                            {step.status === 'pending' && step.required && (
                              <Button
                                onClick={() => handleApproveRequest(selectedRequest.id, step.id)}
                                variant="primary"
                                size="sm"
                                loading={loading}
                              >
                                Complete Step
                              </Button>
                            )}
                          </div>
                          <p className="text-xs text-gray-600 dark:text-gray-400 mb-2">
                            {step.description}
                          </p>
                          {step.status === 'completed' && step.completedBy && (
                            <p className="text-xs text-green-600 dark:text-green-400">
                              Completed by {step.completedBy} at {new Date(step.completedAt!).toLocaleString()}
                            </p>
                          )}
                        </div>
                      </div>
                    ))}
                </div>
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex space-x-3 pt-4 border-t border-gray-200 dark:border-gray-700">
              <Button
                onClick={() => handleApproveRequest(selectedRequest.id)}
                variant={selectedRequest.commandPreview.safetyLevel === 'dangerous' ? 'danger' : 'primary'}
                className="flex-1"
                loading={loading}
              >
                {selectedRequest.commandPreview.safetyLevel === 'dangerous'
                  ? 'Approve Dangerous Operation'
                  : 'Approve Request'
                }
              </Button>
              <Button
                onClick={() => setShowRejectionModal(true)}
                variant="secondary"
                className="flex-1"
              >
                Reject
              </Button>
            </div>
          </div>
        </Modal>
      )}

      {/* Rejection Modal */}
      <Modal
        isOpen={showRejectionModal}
        onClose={() => setShowRejectionModal(false)}
        title="Reject Approval Request"
      >
        <div className="space-y-4">
          <p className="text-gray-700 dark:text-gray-300">
            Please provide a reason for rejecting this approval request:
          </p>

          <textarea
            value={rejectionReason}
            onChange={(e) => setRejectionReason(e.target.value)}
            placeholder="Enter rejection reason..."
            rows={4}
            className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
          />

          <div className="flex space-x-3">
            <Button
              onClick={handleRejectRequest}
              variant="danger"
              className="flex-1"
              loading={loading}
              disabled={!rejectionReason.trim()}
            >
              Confirm Rejection
            </Button>
            <Button
              onClick={() => {
                setShowRejectionModal(false);
                setRejectionReason('');
              }}
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
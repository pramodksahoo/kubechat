import { useState, useEffect } from 'react';
import { ApprovalRequest } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Modal } from '../ui/Modal';
import { formatDistanceToNow } from 'date-fns';

interface CommandApprovalInterfaceProps {
  request: ApprovalRequest;
  onApprove: (requestId: string, comments?: string) => void;
  onReject: (requestId: string, reason: string) => void;
  onViewDetails: (requestId: string) => void;
  userRole: 'requester' | 'approver' | 'admin';
  className?: string;
}

export function CommandApprovalInterface({
  request,
  onApprove,
  onReject,
  onViewDetails,
  userRole,
  className = '',
}: CommandApprovalInterfaceProps) {
  const [showApprovalModal, setShowApprovalModal] = useState(false);
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [approvalComments, setApprovalComments] = useState('');
  const [rejectionReason, setRejectionReason] = useState('');
  const [timeRemaining, setTimeRemaining] = useState('');

  useEffect(() => {
    const updateTimeRemaining = () => {
      const expiresAt = new Date(request.expiresAt);
      const now = new Date();
      
      if (expiresAt <= now) {
        setTimeRemaining('Expired');
      } else {
        setTimeRemaining(formatDistanceToNow(expiresAt, { addSuffix: true }));
      }
    };

    updateTimeRemaining();
    const interval = setInterval(updateTimeRemaining, 60000); // Update every minute

    return () => clearInterval(interval);
  }, [request.expiresAt]);

  const getStatusColor = () => {
    switch (request.status) {
      case 'pending': return 'border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-900/20';
      case 'approved': return 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20';
      case 'rejected': return 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/20';
      case 'expired': return 'border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800';
      default: return 'border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800';
    }
  };

  const getStatusIcon = () => {
    switch (request.status) {
      case 'pending':
        return (
          <div className="flex items-center space-x-2 text-yellow-600 dark:text-yellow-400">
            <svg className="w-5 h-5 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span className="font-medium">Pending Approval</span>
          </div>
        );
      case 'approved':
        return (
          <div className="flex items-center space-x-2 text-green-600 dark:text-green-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">Approved</span>
          </div>
        );
      case 'rejected':
        return (
          <div className="flex items-center space-x-2 text-red-600 dark:text-red-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">Rejected</span>
          </div>
        );
      case 'expired':
        return (
          <div className="flex items-center space-x-2 text-gray-600 dark:text-gray-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">Expired</span>
          </div>
        );
      default:
        return null;
    }
  };

  const getSafetyBadge = () => {
    const level = request.commandPreview.safetyLevel;
    const colors = {
      safe: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      dangerous: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    };

    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${colors[level]}`}>
        {level.toUpperCase()}
      </span>
    );
  };

  const handleApprove = () => {
    onApprove(request.id, approvalComments || undefined);
    setShowApprovalModal(false);
    setApprovalComments('');
  };

  const handleReject = () => {
    onReject(request.id, rejectionReason);
    setShowRejectModal(false);
    setRejectionReason('');
  };

  const canApprove = userRole === 'approver' || userRole === 'admin';
  const canReject = userRole === 'approver' || userRole === 'admin';
  const isExpired = new Date(request.expiresAt) <= new Date();
  const isPending = request.status === 'pending' && !isExpired;

  return (
    <>
      <Card className={`${getStatusColor()} border-2 ${className}`}>
        <div className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-4">
            {getStatusIcon()}
            <div className="flex items-center space-x-3">
              {getSafetyBadge()}
              <div className="text-sm text-gray-500 dark:text-gray-400">
                Expires {timeRemaining}
              </div>
            </div>
          </div>

          {/* Request Details */}
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Requested Operation:
              </label>
              <p className="text-gray-900 dark:text-white bg-white dark:bg-gray-800 p-3 rounded-lg border">
                {request.commandPreview.naturalLanguage}
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Generated Command:
              </label>
              <div className="bg-gray-900 text-green-400 p-3 rounded-lg font-mono text-sm overflow-x-auto">
                <code>{request.commandPreview.generatedCommand}</code>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Requested By:
                </label>
                <p className="text-gray-700 dark:text-gray-300">{request.requestedBy}</p>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Requested At:
                </label>
                <p className="text-gray-700 dark:text-gray-300">
                  {formatDistanceToNow(new Date(request.requestedAt), { addSuffix: true })}
                </p>
              </div>
            </div>

            {/* Impact Assessment */}
            {request.commandPreview.potentialImpact.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Potential Impact:
                </label>
                <ul className="space-y-1">
                  {request.commandPreview.potentialImpact.map((impact, index) => (
                    <li key={index} className="flex items-start space-x-2 text-sm">
                      <span className="text-gray-400 mt-1.5 w-1 h-1 bg-current rounded-full flex-shrink-0"></span>
                      <span className="text-gray-700 dark:text-gray-300">{impact}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Approval/Rejection Details */}
            {request.status === 'approved' && request.approvedBy && (
              <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg">
                <p className="text-sm text-green-700 dark:text-green-300">
                  <strong>Approved by:</strong> {request.approvedBy}
                </p>
                {request.approvedAt && (
                  <p className="text-sm text-green-600 dark:text-green-400">
                    {formatDistanceToNow(new Date(request.approvedAt), { addSuffix: true })}
                  </p>
                )}
              </div>
            )}

            {request.status === 'rejected' && request.rejectedBy && (
              <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
                <p className="text-sm text-red-700 dark:text-red-300">
                  <strong>Rejected by:</strong> {request.rejectedBy}
                </p>
                {request.rejectionReason && (
                  <p className="text-sm text-red-600 dark:text-red-400 mt-1">
                    <strong>Reason:</strong> {request.rejectionReason}
                  </p>
                )}
                {request.rejectedAt && (
                  <p className="text-sm text-red-500 dark:text-red-400 mt-1">
                    {formatDistanceToNow(new Date(request.rejectedAt), { addSuffix: true })}
                  </p>
                )}
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div className="flex space-x-3 mt-6">
            {isPending && canApprove && (
              <Button
                onClick={() => setShowApprovalModal(true)}
                variant="primary"
                className="flex-1"
              >
                <div className="flex items-center justify-center space-x-2">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  <span>Approve</span>
                </div>
              </Button>
            )}

            {isPending && canReject && (
              <Button
                onClick={() => setShowRejectModal(true)}
                variant="danger"
                className="flex-1"
              >
                <div className="flex items-center justify-center space-x-2">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  <span>Reject</span>
                </div>
              </Button>
            )}

            <Button
              onClick={() => onViewDetails(request.id)}
              variant="secondary"
              className={isPending && (canApprove || canReject) ? '' : 'flex-1'}
            >
              View Details
            </Button>
          </div>
        </div>
      </Card>

      {/* Approval Modal */}
      <Modal
        isOpen={showApprovalModal}
        onClose={() => setShowApprovalModal(false)}
        title="Approve Command"
      >
        <div className="space-y-4">
          <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg">
            <p className="text-sm text-green-700 dark:text-green-300">
              You are about to approve the execution of this command. This action cannot be undone.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Approval Comments (Optional):
            </label>
            <textarea
              value={approvalComments}
              onChange={(e) => setApprovalComments(e.target.value)}
              placeholder="Add any comments about this approval..."
              rows={3}
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
            />
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleApprove}
              variant="primary"
              className="flex-1"
            >
              Confirm Approval
            </Button>
            <Button
              onClick={() => setShowApprovalModal(false)}
              variant="secondary"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>

      {/* Rejection Modal */}
      <Modal
        isOpen={showRejectModal}
        onClose={() => setShowRejectModal(false)}
        title="Reject Command"
      >
        <div className="space-y-4">
          <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
            <p className="text-sm text-red-700 dark:text-red-300">
              You are about to reject this command request. Please provide a reason for the rejection.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Rejection Reason (Required):
            </label>
            <textarea
              value={rejectionReason}
              onChange={(e) => setRejectionReason(e.target.value)}
              placeholder="Explain why this command is being rejected..."
              rows={3}
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              required
            />
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleReject}
              variant="danger"
              className="flex-1"
              disabled={!rejectionReason.trim()}
            >
              Confirm Rejection
            </Button>
            <Button
              onClick={() => setShowRejectModal(false)}
              variant="secondary"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}
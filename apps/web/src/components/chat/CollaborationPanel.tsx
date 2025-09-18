// Collaboration Panel Component for Story 2.2
// Enables team collaboration features for chat sessions

import React, { useState, useEffect } from 'react';
import { useAuthStore } from '../../stores/authStore';
import { useWebSocketStore } from '../../stores/websocketStore';
import { chatSessionService } from '../../services/chat';

interface CollaborationUser {
  id: string;
  username: string;
  email: string;
  isOnline: boolean;
  isTyping: boolean;
  lastSeen: Date;
}

interface CollaborationPanelProps {
  sessionId: string;
  isVisible: boolean;
  onToggle: () => void;
  className?: string;
}

export function CollaborationPanel({
  sessionId,
  isVisible,
  onToggle,
  className = '',
}: CollaborationPanelProps) {
  const [collaborators, setCollaborators] = useState<CollaborationUser[]>([]);
  const [shareLink, setShareLink] = useState<string>('');
  const [isSharing, setIsSharing] = useState(false);
  const [shareError, setShareError] = useState<string>('');
  const [inviteEmail, setInviteEmail] = useState('');

  const { user } = useAuthStore();
  const {
    connected: wsConnected,
    joinCollaborativeSession,
    sendTypingIndicator,
  } = useWebSocketStore();

  // Join collaborative session when component mounts
  useEffect(() => {
    if (sessionId && wsConnected) {
      joinCollaborativeSession(sessionId).catch(console.error);
    }
  }, [sessionId, wsConnected, joinCollaborativeSession]);

  // Mock collaborators data - in real implementation, this would come from WebSocket
  useEffect(() => {
    const mockCollaborators: CollaborationUser[] = [
      {
        id: user?.id || 'current-user',
        username: user?.username || 'You',
        email: user?.email || 'you@example.com',
        isOnline: true,
        isTyping: false,
        lastSeen: new Date(),
      },
    ];
    setCollaborators(mockCollaborators);
  }, [user]);

  const handleShareSession = async () => {
    if (!sessionId) return;

    try {
      setIsSharing(true);
      setShareError('');

      // Get user IDs to share with (in real implementation, this would be a user picker)
      const userIds: string[] = []; // Empty for now

      const shareResult = await chatSessionService.shareSession(sessionId, userIds);
      setShareLink(shareResult.shareUrl);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to share session';
      setShareError(errorMessage);
    } finally {
      setIsSharing(false);
    }
  };

  const handleInviteUser = async () => {
    if (!inviteEmail.trim()) return;

    try {
      // In real implementation, this would send an email invitation
      console.log('Inviting user:', inviteEmail);
      setInviteEmail('');
    } catch (error) {
      console.error('Failed to invite user:', error);
    }
  };

  const handleTypingStart = () => {
    if (wsConnected) {
      sendTypingIndicator(sessionId, true);
    }
  };

  const handleTypingStop = () => {
    if (wsConnected) {
      sendTypingIndicator(sessionId, false);
    }
  };

  const copyShareLink = () => {
    if (shareLink) {
      navigator.clipboard.writeText(shareLink);
      // Could show a toast notification
    }
  };

  const formatLastSeen = (date: Date) => {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  };

  if (!isVisible) {
    return (
      <button
        onClick={onToggle}
        className="fixed right-4 top-1/2 transform -translate-y-1/2 bg-primary-600 hover:bg-primary-700 text-white p-2 rounded-l-lg shadow-lg z-40"
        title="Show collaboration panel"
      >
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
      </button>
    );
  }

  return (
    <div className={`bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Collaboration</h3>
        <button
          onClick={onToggle}
          className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          title="Hide collaboration panel"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="p-4 space-y-6">
        {/* Active Collaborators */}
        <div>
          <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
            Active Users ({collaborators.filter(c => c.isOnline).length})
          </h4>
          <div className="space-y-2">
            {collaborators.map((collaborator) => (
              <div
                key={collaborator.id}
                className="flex items-center space-x-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                <div className="relative">
                  <div className="w-8 h-8 bg-gradient-to-br from-primary-400 to-primary-600 rounded-full flex items-center justify-center text-white text-sm font-medium">
                    {collaborator.username.charAt(0).toUpperCase()}
                  </div>
                  {collaborator.isOnline && (
                    <div className="absolute -bottom-1 -right-1 w-3 h-3 bg-green-500 border-2 border-white dark:border-gray-800 rounded-full"></div>
                  )}
                </div>

                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium text-gray-900 dark:text-white truncate">
                    {collaborator.username}
                    {collaborator.id === user?.id && (
                      <span className="text-xs text-gray-500 dark:text-gray-400 ml-1">(You)</span>
                    )}
                  </div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    {collaborator.isTyping ? (
                      <span className="text-blue-600 dark:text-blue-400 animate-pulse">Typing...</span>
                    ) : collaborator.isOnline ? (
                      'Online'
                    ) : (
                      `Last seen ${formatLastSeen(collaborator.lastSeen)}`
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Share Session */}
        <div>
          <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">Share Session</h4>

          {/* Generate Share Link */}
          {!shareLink ? (
            <button
              onClick={handleShareSession}
              disabled={isSharing}
              className="w-full px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400 text-white rounded-lg text-sm font-medium transition-colors"
            >
              {isSharing ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Generating...
                </div>
              ) : (
                'Generate Share Link'
              )}
            </button>
          ) : (
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  value={shareLink}
                  readOnly
                  className="flex-1 px-3 py-2 text-xs bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded text-gray-600 dark:text-gray-300"
                />
                <button
                  onClick={copyShareLink}
                  className="px-3 py-2 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300 rounded text-xs"
                  title="Copy link"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
              <p className="text-xs text-gray-500 dark:text-gray-400">
                Anyone with this link can join the session
              </p>
            </div>
          )}

          {shareError && (
            <div className="mt-2 text-xs text-red-600 dark:text-red-400">
              {shareError}
            </div>
          )}
        </div>

        {/* Invite by Email */}
        <div>
          <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">Invite User</h4>
          <div className="flex space-x-2">
            <input
              type="email"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
              placeholder="Enter email address"
              className="flex-1 px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
              onKeyPress={(e) => {
                if (e.key === 'Enter') {
                  handleInviteUser();
                }
              }}
            />
            <button
              onClick={handleInviteUser}
              disabled={!inviteEmail.trim()}
              className="px-3 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-300 dark:disabled:bg-gray-600 text-white rounded-lg text-sm"
            >
              Invite
            </button>
          </div>
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Send an invitation email to collaborate
          </p>
        </div>

        {/* Connection Status */}
        <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-2">
            <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-gray-400'}`}></div>
            <span className="text-xs text-gray-500 dark:text-gray-400">
              {wsConnected ? 'Real-time sync active' : 'Disconnected'}
            </span>
          </div>
        </div>

        {/* Typing Indicator Test */}
        <div className="pt-2">
          <div className="flex space-x-2">
            <button
              onClick={handleTypingStart}
              className="px-2 py-1 text-xs bg-blue-100 hover:bg-blue-200 dark:bg-blue-900 dark:hover:bg-blue-800 text-blue-700 dark:text-blue-300 rounded"
            >
              Start Typing
            </button>
            <button
              onClick={handleTypingStop}
              className="px-2 py-1 text-xs bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded"
            >
              Stop Typing
            </button>
          </div>
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Test typing indicators
          </p>
        </div>
      </div>
    </div>
  );
}
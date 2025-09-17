// Session Selector Component for Story 2.2
// Allows switching between multiple concurrent chat sessions

import React, { useState, useEffect } from 'react';
import { ChatSession } from '../../types/chat';
import { chatSessionService } from '../../services/chat';
import { useAuthStore } from '../../stores/authStore';

interface SessionSelectorProps {
  currentSession: ChatSession | null;
  onSessionSelect: (session: ChatSession) => void;
  onNewSession: () => void;
  className?: string;
}

export function SessionSelector({
  currentSession,
  onSessionSelect,
  onNewSession,
  className = '',
}: SessionSelectorProps) {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [loading, setLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const { isAuthenticated } = useAuthStore();

  // Load available sessions
  useEffect(() => {
    if (isAuthenticated) {
      loadSessions();
    }
  }, [isAuthenticated]);

  const loadSessions = async () => {
    try {
      setLoading(true);
      const sessionList = await chatSessionService.getSessions({ status: 'active', limit: 20 });
      setSessions(sessionList);
    } catch (error) {
      console.error('Failed to load sessions:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSessionSelect = (session: ChatSession) => {
    onSessionSelect(session);
    setShowDropdown(false);
  };

  const handleNewSession = () => {
    onNewSession();
    setShowDropdown(false);
  };

  const formatSessionTitle = (session: ChatSession) => {
    const maxLength = 30;
    const title = session.title || `Session ${session.id.slice(-6)}`;
    return title.length > maxLength ? `${title.slice(0, maxLength)}...` : title;
  };

  const formatLastActive = (session: ChatSession) => {
    const date = new Date(session.updatedAt);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  return (
    <div className={`relative ${className}`}>
      {/* Current Session Display */}
      <button
        onClick={() => setShowDropdown(!showDropdown)}
        className="flex items-center justify-between w-full px-4 py-2 text-left bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
        disabled={loading}
      >
        <div className="flex items-center space-x-3 min-w-0 flex-1">
          <div className="flex-shrink-0">
            <div className="w-8 h-8 bg-primary-100 dark:bg-primary-900 rounded-full flex items-center justify-center">
              <svg className="w-4 h-4 text-primary-600 dark:text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
          </div>

          <div className="min-w-0 flex-1">
            <div className="text-sm font-medium text-gray-900 dark:text-white truncate">
              {currentSession ? formatSessionTitle(currentSession) : 'No Session'}
            </div>
            {currentSession && (
              <div className="text-xs text-gray-500 dark:text-gray-400">
                {currentSession.messageCount} messages • {formatLastActive(currentSession)}
              </div>
            )}
          </div>
        </div>

        <div className="flex-shrink-0">
          <svg
            className={`w-5 h-5 text-gray-400 transition-transform ${showDropdown ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </button>

      {/* Dropdown Menu */}
      {showDropdown && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 max-h-96 overflow-y-auto">
          {/* New Session Button */}
          <button
            onClick={handleNewSession}
            className="w-full px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-700 border-b border-gray-200 dark:border-gray-700"
          >
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center">
                <svg className="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              </div>
              <span className="text-sm font-medium text-gray-900 dark:text-white">New Chat Session</span>
            </div>
          </button>

          {/* Loading State */}
          {loading && (
            <div className="px-4 py-3 text-center">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-600 mx-auto"></div>
              <span className="text-sm text-gray-500 dark:text-gray-400 mt-2 block">Loading sessions...</span>
            </div>
          )}

          {/* Session List */}
          {!loading && sessions.length === 0 && (
            <div className="px-4 py-3 text-center text-sm text-gray-500 dark:text-gray-400">
              No active sessions found
            </div>
          )}

          {!loading && sessions.map((session) => (
            <button
              key={session.id}
              onClick={() => handleSessionSelect(session)}
              className={`w-full px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-700 ${
                currentSession?.id === session.id
                  ? 'bg-primary-50 dark:bg-primary-900/20 border-l-4 border-primary-500'
                  : ''
              }`}
            >
              <div className="flex items-center space-x-3">
                <div className="flex-shrink-0">
                  <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
                    session.status === 'active'
                      ? 'bg-blue-100 dark:bg-blue-900'
                      : 'bg-gray-100 dark:bg-gray-700'
                  }`}>
                    <svg className={`w-4 h-4 ${
                      session.status === 'active'
                        ? 'text-blue-600 dark:text-blue-400'
                        : 'text-gray-400'
                    }`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                    </svg>
                  </div>
                </div>

                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium text-gray-900 dark:text-white truncate">
                    {formatSessionTitle(session)}
                  </div>
                  <div className="text-xs text-gray-500 dark:text-gray-400 flex items-center space-x-2">
                    <span>{session.messageCount} messages</span>
                    <span>•</span>
                    <span>{formatLastActive(session)}</span>
                    {session.clusterName && (
                      <>
                        <span>•</span>
                        <span className="text-blue-600 dark:text-blue-400">{session.clusterName}</span>
                      </>
                    )}
                  </div>
                </div>

                {session.status === 'active' && (
                  <div className="flex-shrink-0">
                    <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  </div>
                )}
              </div>
            </button>
          ))}

          {/* View All Sessions Link */}
          {!loading && sessions.length > 0 && (
            <div className="border-t border-gray-200 dark:border-gray-700">
              <button
                onClick={() => {
                  setShowDropdown(false);
                  // Could navigate to a sessions management page
                }}
                className="w-full px-4 py-3 text-sm text-primary-600 dark:text-primary-400 hover:bg-gray-50 dark:hover:bg-gray-700 text-center"
              >
                View All Sessions
              </button>
            </div>
          )}
        </div>
      )}

      {/* Backdrop */}
      {showDropdown && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setShowDropdown(false)}
        />
      )}
    </div>
  );
}
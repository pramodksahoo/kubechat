// Enhanced Chat Interface for Story 2.2
// Integrates all authenticated services with modern UI and collaboration features

import React, { useState, useEffect, useRef } from 'react';
import { ChatMessage, ChatSession, CommandPreview } from '../../types/chat';
import { useAuthStore } from '../../stores/authStore';
import { useCommandStore } from '../../stores/commandStore';
import { useWebSocketStore } from '../../stores/websocketStore';
import { chatSessionService } from '../../services/chat';

// Import existing components
import { ChatMessageList } from './ChatMessageList';
import { ChatInput } from './ChatInput';
import { CommandPreviewCard } from './CommandPreviewCard';
import { SessionSelector } from './SessionSelector';
import { CollaborationPanel } from './CollaborationPanel';
import { ErrorDisplay } from './ErrorDisplay';
import { LoadingIndicator } from './LoadingIndicator';

interface EnhancedChatInterfaceProps {
  initialSessionId?: string;
  showCollaboration?: boolean;
  className?: string;
}

export function EnhancedChatInterface({
  initialSessionId,
  showCollaboration = true,
  className = '',
}: EnhancedChatInterfaceProps) {
  // Local state
  const [currentSession, setCurrentSession] = useState<ChatSession | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCollaborationPanel, setShowCollaborationPanel] = useState(false);
  const [inputValue, setInputValue] = useState('');

  // Store state
  const { isAuthenticated, user } = useAuthStore();
  const {
    currentPreview,
    showPreview,
    executeCommand,
    generatePreview,
    approveCommand,
    setCurrentPreview,
    showCommandPreview,
    clearError: clearCommandError,
  } = useCommandStore();
  const {
    connect: connectWebSocket,
    connected: wsConnected,
    subscribeToSession,
    unsubscribeFromSession,
  } = useWebSocketStore();

  // Refs
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Initialize chat interface
  useEffect(() => {
    if (isAuthenticated) {
      initializeChatInterface();
    }
  }, [isAuthenticated, initialSessionId]);

  // Connect WebSocket when authenticated
  useEffect(() => {
    if (isAuthenticated && !wsConnected) {
      connectWebSocket({ sessionId: currentSession?.id }).catch(console.error);
    }
  }, [isAuthenticated, wsConnected, currentSession?.id, connectWebSocket]);

  // Subscribe to session updates
  useEffect(() => {
    if (currentSession && wsConnected) {
      subscribeToSession(currentSession.id);
      return () => {
        unsubscribeFromSession(currentSession.id);
      };
    }
  }, [currentSession, wsConnected, subscribeToSession, unsubscribeFromSession]);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const initializeChatInterface = async () => {
    try {
      setLoading(true);
      setError(null);

      let session: ChatSession | null = null;

      if (initialSessionId) {
        // Load specific session
        try {
          session = await chatSessionService.getSession(initialSessionId);
        } catch (error) {
          console.warn('Failed to load initial session:', error);
        }
      }

      if (!session) {
        // Try to restore last session or create new one
        session = await chatSessionService.restoreSession();
        if (!session) {
          session = await chatSessionService.createSession();
        }
      }

      setCurrentSession(session);

      // Load messages for the session
      if (session) {
        await loadSessionMessages(session.id);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to initialize chat';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const loadSessionMessages = async (sessionId: string) => {
    try {
      // In a real implementation, this would load messages from the chat service
      // For now, we'll use empty array as messages are managed elsewhere
      setMessages([]);
    } catch (error) {
      console.error('Failed to load messages:', error);
    }
  };

  const handleSessionSelect = async (session: ChatSession) => {
    try {
      setLoading(true);
      setCurrentSession(session);
      await loadSessionMessages(session.id);

      // Clear any existing preview
      setCurrentPreview(null);
      showCommandPreview(false);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to switch session';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleNewSession = async () => {
    try {
      setLoading(true);
      const newSession = await chatSessionService.createSession();
      setCurrentSession(newSession);
      setMessages([]);

      // Clear any existing preview
      setCurrentPreview(null);
      showCommandPreview(false);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to create session';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleSendMessage = async (message: string) => {
    if (!currentSession || !message.trim()) return;

    try {
      clearError();
      clearCommandError();

      // Clear input
      setInputValue('');

      // Check if message looks like a command request
      if (isCommandLike(message)) {
        // Generate command preview
        await generatePreview(message, currentSession.id);
      }

      // Add user message to local state (would be handled by chat service in real implementation)
      const userMessage: ChatMessage = {
        id: `msg-${Date.now()}`,
        sessionId: currentSession.id,
        userId: user?.id || 'current-user',
        type: 'user',
        content: message,
        timestamp: new Date().toISOString(),
      };

      setMessages(prev => [...prev, userMessage]);

      // Send message through chat service (implementation would depend on existing chat service)
      // await chatService.sendMessage(currentSession.id, message);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to send message';
      setError(errorMessage);
    }
  };

  const handleCommandApprove = async () => {
    if (!currentPreview) return;

    try {
      await approveCommand(currentPreview.id);

      // Add system message about command execution
      const systemMessage: ChatMessage = {
        id: `msg-${Date.now()}`,
        sessionId: currentSession?.id || 'unknown',
        userId: 'system',
        type: 'system',
        content: currentPreview.approvalRequired
          ? 'Approval request submitted. Waiting for administrator approval.'
          : 'Command executed successfully.',
        timestamp: new Date().toISOString(),
        metadata: {
          command: currentPreview.generatedCommand,
          safetyLevel: currentPreview.safetyLevel,
        },
      };

      setMessages(prev => [...prev, systemMessage]);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Command execution failed';
      setError(errorMessage);
    }
  };

  const handleCommandReject = (reason?: string) => {
    setCurrentPreview(null);
    showCommandPreview(false);

    if (reason && currentSession) {
      const rejectMessage: ChatMessage = {
        id: `msg-${Date.now()}`,
        sessionId: currentSession.id,
        userId: 'system',
        type: 'system',
        content: `Command cancelled: ${reason}`,
        timestamp: new Date().toISOString(),
      };

      setMessages(prev => [...prev, rejectMessage]);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const clearError = () => {
    setError(null);
  };

  // Helper function to detect command-like messages
  const isCommandLike = (message: string): boolean => {
    const commandKeywords = [
      'kubectl', 'create', 'delete', 'deploy', 'scale', 'restart',
      'get pods', 'describe', 'logs', 'exec', 'apply', 'rollout',
      'show me', 'list', 'check', 'update', 'remove'
    ];

    const lowerMessage = message.toLowerCase();
    return commandKeywords.some(keyword => lowerMessage.includes(keyword));
  };

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-gray-600 dark:text-gray-400">Please log in to access the chat interface.</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`flex h-full ${className}`}>
      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header with Session Selector */}
        <div className="flex-shrink-0 p-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
          <div className="flex items-center justify-between">
            <div className="flex-1 max-w-md">
              <SessionSelector
                currentSession={currentSession}
                onSessionSelect={handleSessionSelect}
                onNewSession={handleNewSession}
              />
            </div>

            <div className="flex items-center space-x-4">
              {/* Connection Status */}
              <div className="flex items-center space-x-2">
                <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`}></div>
                <span className="text-xs text-gray-500 dark:text-gray-400">
                  {wsConnected ? 'Online' : 'Offline'}
                </span>
              </div>

              {/* Collaboration Toggle */}
              {showCollaboration && currentSession && (
                <button
                  onClick={() => setShowCollaborationPanel(!showCollaborationPanel)}
                  className={`p-2 rounded-lg transition-colors ${
                    showCollaborationPanel
                      ? 'bg-primary-100 dark:bg-primary-900 text-primary-600 dark:text-primary-400'
                      : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
                  }`}
                  title="Toggle collaboration panel"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Error Display */}
        {error && (
          <div className="flex-shrink-0 p-4">
            <ErrorDisplay
              type="general"
              error={error}
              onRetry={clearError}
              onDismiss={clearError}
            />
          </div>
        )}

        {/* Messages Area */}
        <div className="flex-1 overflow-hidden">
          {loading ? (
            <div className="flex items-center justify-center h-full">
              <LoadingIndicator />
            </div>
          ) : currentSession ? (
            <ChatMessageList
              messages={messages}
              loading={false}
              className="h-full"
            />
          ) : (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <p className="text-gray-600 dark:text-gray-400">No chat session active.</p>
                <button
                  onClick={handleNewSession}
                  className="mt-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg"
                >
                  Start New Chat
                </button>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Command Preview */}
        {showPreview && currentPreview && (
          <div className="flex-shrink-0 p-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
            <CommandPreviewCard
              preview={currentPreview}
              onApprove={handleCommandApprove}
              onReject={handleCommandReject}
            />
          </div>
        )}

        {/* Chat Input */}
        {currentSession && (
          <div className="flex-shrink-0 p-4 border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
            <ChatInput
              value={inputValue}
              onChange={setInputValue}
              onSend={handleSendMessage}
              loading={loading}
              placeholder="Type your message or describe what you want to do with Kubernetes..."
            />
          </div>
        )}
      </div>

      {/* Collaboration Panel */}
      {showCollaboration && currentSession && (
        <CollaborationPanel
          sessionId={currentSession.id}
          isVisible={showCollaborationPanel}
          onToggle={() => setShowCollaborationPanel(!showCollaborationPanel)}
          className={`transition-all duration-300 ${
            showCollaborationPanel ? 'w-80' : 'w-0'
          }`}
        />
      )}
    </div>
  );
}
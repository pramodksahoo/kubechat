import { useState, useEffect, useCallback } from 'react';
import { ChatMessage, ChatSession, CommandPreview } from '../types/chat';
import { chatService } from '@/services/chatService';

export interface UseChatOptions {
  sessionId?: string;
  autoLoadHistory?: boolean;
  persistLocally?: boolean;
}

export function useChat(options: UseChatOptions = {}) {
  const { sessionId: initialSessionId, autoLoadHistory = true, persistLocally = true } = options;

  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ChatSession | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [connected, setConnected] = useState(false);
  const [websocket, setWebsocket] = useState<WebSocket | null>(null);

  // Load sessions on mount
  useEffect(() => {
    const loadSessions = async () => {
      try {
        setLoading(true);

        // Try to load from server first
        try {
          const serverSessions = await chatService.getSessions();
          setSessions(serverSessions);

          // Find or create session
          let session = serverSessions.find(s => s.id === initialSessionId);
          if (!session && serverSessions.length > 0) {
            session = serverSessions[0]; // Use most recent
          }

          if (!session) {
            // Create new session
            session = await chatService.createSession();
            setSessions(prev => [session!, ...prev]);
          }

          setCurrentSession(session);

          if (persistLocally) {
            chatService.saveSessionLocally(session);
          }
        } catch (serverError) {
          console.warn('Failed to load from server, using local storage:', serverError);

          // Fallback to local storage
          const localSessions = chatService.getLocalSessions();
          setSessions(localSessions);

          if (localSessions.length > 0) {
            const session = localSessions.find(s => s.id === initialSessionId) || localSessions[0];
            setCurrentSession(session);
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load chat sessions');
      } finally {
        setLoading(false);
      }
    };

    loadSessions();
  }, [initialSessionId, persistLocally]);

  // Load messages when session changes
  useEffect(() => {
    const loadMessages = async () => {
      if (!currentSession || !autoLoadHistory) return;

      try {
        setLoading(true);

        // Try server first
        try {
          const serverMessages = await chatService.getMessages(currentSession.id);
          setMessages(serverMessages);

          if (persistLocally) {
            chatService.saveMessagesLocally(currentSession.id, serverMessages);
          }
        } catch (serverError) {
          console.warn('Failed to load messages from server, using local storage:', serverError);

          // Fallback to local storage
          const localMessages = chatService.getLocalMessages(currentSession.id);
          setMessages(localMessages);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load messages');
      } finally {
        setLoading(false);
      }
    };

    loadMessages();
  }, [currentSession?.id, autoLoadHistory, persistLocally]);

  // Setup WebSocket connection
  useEffect(() => {
    if (!currentSession) return;

    try {
      chatService.connectWebSocket(currentSession.id, (data) => {
        if (data.type === 'new_message') {
          const newMessage: ChatMessage = data.message;
          setMessages(prev => {
            const updated = [...prev, newMessage];
            if (persistLocally && currentSession) {
              chatService.saveMessagesLocally(currentSession.id, updated);
            }
            return updated;
          });
        }
      });

      setConnected(true);
      setError(null);

      return () => {
        // WebSocket cleanup handled by service
        setConnected(false);
      };
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setConnected(false);
    }
  }, [currentSession?.id, persistLocally]);

  const sendMessage = useCallback(async (content: string) => {
    if (!currentSession || !content.trim()) return;

    const userMessage: ChatMessage = {
      id: `msg-${Date.now()}`,
      sessionId: currentSession.id,
      content: content.trim(),
      type: 'user',
      timestamp: new Date().toISOString(),
      userId: 'current-user', // TODO: Get from auth
    };

    // Optimistically add message
    setMessages(prev => {
      const updated = [...prev, userMessage];
      if (persistLocally) {
        chatService.saveMessagesLocally(currentSession.id, updated);
      }
      return updated;
    });

    try {
      setLoading(true);

      // Send to server
      const serverMessage = await chatService.sendMessage(currentSession.id, content);

      // Replace optimistic message with server response
      setMessages(prev => {
        const withoutOptimistic = prev.filter(m => m.id !== userMessage.id);
        const updated = [...withoutOptimistic, serverMessage];

        if (persistLocally) {
          chatService.saveMessagesLocally(currentSession.id, updated);
        }

        return updated;
      });

    } catch (err) {
      // Remove optimistic message on error
      setMessages(prev => prev.filter(m => m.id !== userMessage.id));
      setError(err instanceof Error ? err.message : 'Failed to send message');
    } finally {
      setLoading(false);
    }
  }, [currentSession, persistLocally]);

  const generateCommandPreview = useCallback(async (naturalLanguage: string): Promise<CommandPreview | null> => {
    if (!currentSession) return null;

    try {
      return await chatService.generateCommandPreview(naturalLanguage, currentSession.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate command preview');
      return null;
    }
  }, [currentSession]);

  const createNewSession = useCallback(async (clusterId?: string) => {
    try {
      setLoading(true);
      const newSession = await chatService.createSession(clusterId);

      setSessions(prev => [newSession, ...prev]);
      setCurrentSession(newSession);
      setMessages([]);

      if (persistLocally) {
        chatService.saveSessionLocally(newSession);
      }

      return newSession;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create session');
      return null;
    } finally {
      setLoading(false);
    }
  }, [persistLocally]);

  const switchSession = useCallback((sessionId: string) => {
    const session = sessions.find(s => s.id === sessionId);
    if (session) {
      setCurrentSession(session);
    }
  }, [sessions]);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const clearLocalData = useCallback(() => {
    chatService.clearLocalData();
    setSessions([]);
    setCurrentSession(null);
    setMessages([]);
  }, []);

  return {
    // State
    sessions,
    currentSession,
    messages,
    loading,
    error,
    connected,

    // Actions
    sendMessage,
    generateCommandPreview,
    createNewSession,
    switchSession,
    clearError,
    clearLocalData,
  };
}
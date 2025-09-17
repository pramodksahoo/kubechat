// WebSocket Connection Store for Story 2.2
// State management for authenticated real-time WebSocket communication

import { create } from 'zustand';
import { authenticatedWebSocketService } from '../services/chat';
import { useAuthStore } from './authStore';

interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: number;
  sessionId?: string;
  userId?: string;
}

interface WebSocketState {
  // Connection state
  connected: boolean;
  connecting: boolean;
  connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

  // Subscriptions
  subscribedSessions: Set<string>;
  subscribedCommands: Set<string>;

  // Message history
  recentMessages: WebSocketMessage[];

  // Error state
  error: string | null;

  // Statistics
  messageCount: number;
  lastMessageTime: number | null;

  // Actions
  connect: (options?: { sessionId?: string }) => Promise<void>;
  disconnect: () => void;

  // Session subscriptions
  subscribeToSession: (sessionId: string) => Promise<void>;
  unsubscribeFromSession: (sessionId: string) => Promise<void>;

  // Command subscriptions
  subscribeToCommand: (executionId: string) => Promise<void>;
  unsubscribeFromCommand: (executionId: string) => Promise<void>;

  // Collaboration
  joinCollaborativeSession: (sessionId: string) => Promise<void>;
  sendTypingIndicator: (sessionId: string, isTyping: boolean) => Promise<void>;

  // Message handling
  sendMessage: (message: Omit<WebSocketMessage, 'timestamp'>) => void;
  clearMessages: () => void;

  // Error handling
  clearError: () => void;
}

export const useWebSocketStore = create<WebSocketState>((set, get) => ({
  // Initial state
  connected: false,
  connecting: false,
  connectionStatus: 'disconnected',
  subscribedSessions: new Set(),
  subscribedCommands: new Set(),
  recentMessages: [],
  error: null,
  messageCount: 0,
  lastMessageTime: null,

  // Connect to WebSocket with authentication
  connect: async (options = {}) => {
    const { connected, connecting } = get();
    if (connected || connecting) return;

    set({ connecting: true, error: null });

    try {
      // Ensure user is authenticated
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated) {
        throw new Error('Authentication required for WebSocket connection');
      }

      // Set up message handler
      const unsubscribeMessage = authenticatedWebSocketService.onMessage((message) => {
        set(state => ({
          recentMessages: [message, ...state.recentMessages.slice(0, 99)], // Keep last 100 messages
          messageCount: state.messageCount + 1,
          lastMessageTime: Date.now(),
        }));

        // Handle specific message types
        handleWebSocketMessage(message, get, set);
      });

      // Set up error handler
      const unsubscribeError = authenticatedWebSocketService.onError((error) => {
        console.error('WebSocket error:', error);
        set({
          error: 'WebSocket connection error',
          connected: false,
          connecting: false,
          connectionStatus: 'disconnected',
        });
      });

      // Set up connection status handler
      const unsubscribeConnection = authenticatedWebSocketService.onConnectionChange((connected) => {
        set({
          connected,
          connecting: false,
          connectionStatus: connected ? 'connected' : 'disconnected',
          error: connected ? null : get().error,
        });
      });

      // Connect with options
      await authenticatedWebSocketService.connect({
        sessionId: options.sessionId,
        autoReconnect: true,
        maxReconnectAttempts: 5,
        reconnectDelay: 1000,
      });

      // Store cleanup functions for later use
      (get() as any)._cleanupFunctions = [
        unsubscribeMessage,
        unsubscribeError,
        unsubscribeConnection,
      ];

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Connection failed';
      set({
        error: errorMessage,
        connected: false,
        connecting: false,
        connectionStatus: 'disconnected',
      });
      throw error;
    }
  },

  // Disconnect from WebSocket
  disconnect: () => {
    // Clean up event handlers
    const cleanupFunctions = (get() as any)._cleanupFunctions;
    if (cleanupFunctions) {
      cleanupFunctions.forEach((cleanup: () => void) => cleanup());
    }

    authenticatedWebSocketService.disconnect();

    set({
      connected: false,
      connecting: false,
      connectionStatus: 'disconnected',
      subscribedSessions: new Set(),
      subscribedCommands: new Set(),
      error: null,
    });
  },

  // Subscribe to session updates
  subscribeToSession: async (sessionId) => {
    try {
      await authenticatedWebSocketService.subscribeToSession(sessionId);
      set(state => ({
        subscribedSessions: new Set([...state.subscribedSessions, sessionId]),
      }));
    } catch (error) {
      console.error('Failed to subscribe to session:', error);
      throw error;
    }
  },

  // Unsubscribe from session updates
  unsubscribeFromSession: async (sessionId) => {
    try {
      await authenticatedWebSocketService.unsubscribeFromSession(sessionId);
      set(state => {
        const newSubscriptions = new Set(state.subscribedSessions);
        newSubscriptions.delete(sessionId);
        return { subscribedSessions: newSubscriptions };
      });
    } catch (error) {
      console.error('Failed to unsubscribe from session:', error);
    }
  },

  // Subscribe to command execution updates
  subscribeToCommand: async (executionId) => {
    try {
      await authenticatedWebSocketService.subscribeToCommandUpdates(executionId);
      set(state => ({
        subscribedCommands: new Set([...state.subscribedCommands, executionId]),
      }));
    } catch (error) {
      console.error('Failed to subscribe to command updates:', error);
      throw error;
    }
  },

  // Unsubscribe from command execution updates
  unsubscribeFromCommand: async (executionId) => {
    try {
      await authenticatedWebSocketService.unsubscribeFromCommandUpdates(executionId);
      set(state => {
        const newSubscriptions = new Set(state.subscribedCommands);
        newSubscriptions.delete(executionId);
        return { subscribedCommands: newSubscriptions };
      });
    } catch (error) {
      console.error('Failed to unsubscribe from command updates:', error);
    }
  },

  // Join collaborative session
  joinCollaborativeSession: async (sessionId) => {
    try {
      const authState = useAuthStore.getState();
      await authenticatedWebSocketService.joinCollaborativeSession(
        sessionId,
        authState.user?.id || 'unknown'
      );
    } catch (error) {
      console.error('Failed to join collaborative session:', error);
      throw error;
    }
  },

  // Send typing indicator
  sendTypingIndicator: async (sessionId, isTyping) => {
    try {
      await authenticatedWebSocketService.sendTypingIndicator(sessionId, isTyping);
    } catch (error) {
      console.error('Failed to send typing indicator:', error);
    }
  },

  // Send custom message
  sendMessage: (message) => {
    try {
      authenticatedWebSocketService.send({
        ...message,
        timestamp: Date.now(),
      });
    } catch (error) {
      console.error('Failed to send message:', error);
      set({ error: 'Failed to send message' });
    }
  },

  // Clear message history
  clearMessages: () => {
    set({
      recentMessages: [],
      messageCount: 0,
      lastMessageTime: null,
    });
  },

  // Clear error
  clearError: () => {
    set({ error: null });
  },
}));

// Handle incoming WebSocket messages
function handleWebSocketMessage(
  message: WebSocketMessage,
  getState: () => WebSocketState,
  setState: (partial: Partial<WebSocketState>) => void
) {
  switch (message.type) {
    case 'NEW_MESSAGE':
      // Handle new chat messages - this would typically update the chat store
      console.log('New chat message received:', message.data);
      break;

    case 'COMMAND_STATUS_UPDATE':
      // Handle command execution status updates - update command store
      console.log('Command status update:', message.data);
      break;

    case 'USER_JOINED':
      // Handle user joining collaborative session
      console.log('User joined session:', message.data);
      break;

    case 'USER_LEFT':
      // Handle user leaving collaborative session
      console.log('User left session:', message.data);
      break;

    case 'TYPING_INDICATOR':
      // Handle typing indicators
      console.log('Typing indicator:', message.data);
      break;

    case 'SESSION_SHARED':
      // Handle session sharing notifications
      console.log('Session shared:', message.data);
      break;

    case 'APPROVAL_REQUESTED':
      // Handle approval request notifications
      console.log('Approval requested:', message.data);
      break;

    case 'SYSTEM_NOTIFICATION':
      // Handle system-wide notifications
      console.log('System notification:', message.data);
      break;

    default:
      console.log('Unhandled WebSocket message type:', message.type, message.data);
  }
}

// Auto-connect when authenticated
useAuthStore.subscribe((state) => {
  const webSocketState = useWebSocketStore.getState();

  if (state.isAuthenticated && !webSocketState.connected && !webSocketState.connecting) {
    // Auto-connect when user becomes authenticated
    webSocketState.connect().catch(console.error);
  } else if (!state.isAuthenticated && webSocketState.connected) {
    // Disconnect when user logs out
    webSocketState.disconnect();
  }
});
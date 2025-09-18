// Authenticated WebSocket Service for Story 2.2
// Implements secure real-time communication with JWT authentication

import { useAuthStore } from '../../stores/authStore';
import { errorHandlingService } from '../errorHandlingService';

interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: number;
  sessionId?: string;
  userId?: string;
}

interface ConnectionOptions {
  sessionId?: string;
  autoReconnect?: boolean;
  maxReconnectAttempts?: number;
  reconnectDelay?: number;
}

type MessageHandler = (message: WebSocketMessage) => void;
type ErrorHandler = (error: Event) => void;
type ConnectionHandler = (connected: boolean) => void;

export class AuthenticatedWebSocketService {
  private ws: WebSocket | null = null;
  private connectionOptions: ConnectionOptions = {};
  private messageHandlers = new Set<MessageHandler>();
  private errorHandlers = new Set<ErrorHandler>();
  private connectionHandlers = new Set<ConnectionHandler>();

  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private pingInterval: NodeJS.Timeout | null = null;
  private isManualClose = false;
  private connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'reconnecting' = 'disconnected';

  // Task 4.1: Implement authenticated WebSocket connection with JWT token validation
  async connect(options: ConnectionOptions = {}): Promise<void> {
    try {
      // Ensure user is authenticated
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated || !authState.tokens) {
        throw new Error('Authentication required for WebSocket connection');
      }

      this.connectionOptions = {
        autoReconnect: true,
        maxReconnectAttempts: 5,
        reconnectDelay: 1000,
        ...options,
      };

      this.maxReconnectAttempts = this.connectionOptions.maxReconnectAttempts || 5;
      this.reconnectDelay = this.connectionOptions.reconnectDelay || 1000;

      await this.establishConnection();
    } catch (error) {
      const errorDetails = await errorHandlingService.handleError(error as Error, {
        context: {
          operation: 'websocket-connect',
          component: 'AuthenticatedWebSocketService',
          sessionId: options.sessionId,
        },
        logToConsole: true,
      });

      this.notifyErrorHandlers(new ErrorEvent('connection-failed', { message: errorDetails.type }));
      throw error;
    }
  }

  private async establishConnection(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.setConnectionStatus('connecting');

        // Get authentication token
        const authState = useAuthStore.getState();
        const token = authState.tokens?.accessToken;

        if (!token) {
          reject(new Error('No authentication token available'));
          return;
        }

        // Task 4.4: Implement WebSocket reconnection handling with authentication renewal
        if (this.isTokenExpired(token)) {
          // Try to refresh token before connecting
          authState.refreshToken().then(() => {
            this.establishConnection().then(resolve).catch(reject);
          }).catch(reject);
          return;
        }

        // Create WebSocket connection with authentication
        const wsUrl = this.buildWebSocketUrl();
        this.ws = new WebSocket(wsUrl);

        // Set up event handlers
        this.ws.onopen = () => {
          console.log('WebSocket connected successfully');
          this.setConnectionStatus('connected');
          this.reconnectAttempts = 0;
          this.isManualClose = false;

          // Send authentication message
          this.sendAuthenticationMessage(token);

          // Start ping/pong to keep connection alive
          this.startPing();

          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event);
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.notifyErrorHandlers(error);
          reject(error);
        };

        this.ws.onclose = (event) => {
          this.handleConnectionClose(event);
        };

        // Connection timeout
        setTimeout(() => {
          if (this.connectionStatus === 'connecting') {
            reject(new Error('WebSocket connection timeout'));
          }
        }, 10000);

      } catch (error) {
        reject(error);
      }
    });
  }

  // Task 4.2: Create real-time chat message synchronization across sessions
  async subscribeToSession(sessionId: string): Promise<void> {
    if (!this.isConnected()) {
      throw new Error('WebSocket not connected');
    }

    const subscribeMessage: WebSocketMessage = {
      type: 'SUBSCRIBE_SESSION',
      data: { sessionId },
      timestamp: Date.now(),
      sessionId,
      userId: useAuthStore.getState().user?.id,
    };

    this.send(subscribeMessage);
  }

  async unsubscribeFromSession(sessionId: string): Promise<void> {
    if (!this.isConnected()) return;

    const unsubscribeMessage: WebSocketMessage = {
      type: 'UNSUBSCRIBE_SESSION',
      data: { sessionId },
      timestamp: Date.now(),
      sessionId,
      userId: useAuthStore.getState().user?.id,
    };

    this.send(unsubscribeMessage);
  }

  // Task 4.3: Add live command execution status updates with progress indicators
  async subscribeToCommandUpdates(executionId: string): Promise<void> {
    if (!this.isConnected()) {
      throw new Error('WebSocket not connected');
    }

    const subscribeMessage: WebSocketMessage = {
      type: 'SUBSCRIBE_COMMAND',
      data: { executionId },
      timestamp: Date.now(),
      userId: useAuthStore.getState().user?.id,
    };

    this.send(subscribeMessage);
  }

  async unsubscribeFromCommandUpdates(executionId: string): Promise<void> {
    if (!this.isConnected()) return;

    const unsubscribeMessage: WebSocketMessage = {
      type: 'UNSUBSCRIBE_COMMAND',
      data: { executionId },
      timestamp: Date.now(),
      userId: useAuthStore.getState().user?.id,
    };

    this.send(unsubscribeMessage);
  }

  // Task 4.5: Create real-time collaboration features for team chat sessions
  async joinCollaborativeSession(sessionId: string, userId: string): Promise<void> {
    if (!this.isConnected()) {
      throw new Error('WebSocket not connected');
    }

    const joinMessage: WebSocketMessage = {
      type: 'JOIN_COLLABORATION',
      data: { sessionId, userId },
      timestamp: Date.now(),
      sessionId,
      userId: useAuthStore.getState().user?.id,
    };

    this.send(joinMessage);
  }

  async sendTypingIndicator(sessionId: string, isTyping: boolean): Promise<void> {
    if (!this.isConnected()) return;

    const typingMessage: WebSocketMessage = {
      type: 'TYPING_INDICATOR',
      data: { sessionId, isTyping },
      timestamp: Date.now(),
      sessionId,
      userId: useAuthStore.getState().user?.id,
    };

    this.send(typingMessage);
  }

  // Task 4.6: Add WebSocket security and rate limiting integration
  private sendAuthenticationMessage(token: string): void {
    const authMessage: WebSocketMessage = {
      type: 'AUTHENTICATE',
      data: { token },
      timestamp: Date.now(),
      userId: useAuthStore.getState().user?.id,
    };

    this.send(authMessage);
  }

  // Send message through WebSocket
  send(message: WebSocketMessage): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected, message not sent:', message);
      return;
    }

    try {
      this.ws.send(JSON.stringify(message));
    } catch (error) {
      console.error('Failed to send WebSocket message:', error);
    }
  }

  // Message handlers
  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler);
    return () => this.errorHandlers.delete(handler);
  }

  onConnectionChange(handler: ConnectionHandler): () => void {
    this.connectionHandlers.add(handler);
    return () => this.connectionHandlers.delete(handler);
  }

  // Handle incoming messages
  private handleMessage(event: MessageEvent): void {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);

      // Handle authentication responses
      if (message.type === 'AUTH_SUCCESS') {
        console.log('WebSocket authentication successful');
        return;
      }

      if (message.type === 'AUTH_FAILED') {
        console.error('WebSocket authentication failed');
        this.disconnect();
        return;
      }

      // Handle ping/pong
      if (message.type === 'PING') {
        this.send({
          type: 'PONG',
          data: message.data,
          timestamp: Date.now(),
        });
        return;
      }

      // Notify message handlers
      this.messageHandlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error('Error in message handler:', error);
        }
      });
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  // Handle connection close
  private handleConnectionClose(event: CloseEvent): void {
    console.log('WebSocket connection closed:', event.code, event.reason);
    this.setConnectionStatus('disconnected');
    this.stopPing();

    // Auto-reconnect if enabled and not manual close
    if (this.connectionOptions.autoReconnect &&
        !this.isManualClose &&
        this.reconnectAttempts < this.maxReconnectAttempts) {

      this.setConnectionStatus('reconnecting');
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);

      console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`);

      setTimeout(() => {
        this.reconnectAttempts++;
        this.establishConnection().catch(error => {
          console.error('Reconnection failed:', error);
          if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnection attempts reached');
            this.setConnectionStatus('disconnected');
          }
        });
      }, delay);
    }
  }

  // Connection management
  disconnect(): void {
    this.isManualClose = true;
    this.stopPing();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.setConnectionStatus('disconnected');
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  getConnectionStatus(): string {
    return this.connectionStatus;
  }

  // Ping/pong for connection keep-alive
  private startPing(): void {
    this.pingInterval = setInterval(() => {
      if (this.isConnected()) {
        this.send({
          type: 'PING',
          data: { timestamp: Date.now() },
          timestamp: Date.now(),
        });
      }
    }, 30000); // Ping every 30 seconds
  }

  private stopPing(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  // Utility methods
  private buildWebSocketUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const baseUrl = process.env.NEXT_PUBLIC_WS_BASE_URL || `${protocol}//${window.location.host}`;
    return `${baseUrl}/api/v1/ws`;
  }

  private isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return Date.now() >= payload.exp * 1000;
    } catch {
      return true;
    }
  }

  private setConnectionStatus(status: typeof this.connectionStatus): void {
    if (this.connectionStatus !== status) {
      this.connectionStatus = status;
      this.notifyConnectionHandlers(status === 'connected');
    }
  }

  private notifyErrorHandlers(error: Event): void {
    this.errorHandlers.forEach(handler => {
      try {
        handler(error);
      } catch (err) {
        console.error('Error in error handler:', err);
      }
    });
  }

  private notifyConnectionHandlers(connected: boolean): void {
    this.connectionHandlers.forEach(handler => {
      try {
        handler(connected);
      } catch (error) {
        console.error('Error in connection handler:', error);
      }
    });
  }
}

// Singleton instance
export const authenticatedWebSocketService = new AuthenticatedWebSocketService();
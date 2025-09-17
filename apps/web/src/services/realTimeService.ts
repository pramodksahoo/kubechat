import { webSocketService, api } from './api';
import { WebSocketMessage } from '../hooks/useWebSocket';
import { useState, useEffect, useCallback } from 'react';

export interface RealTimeUpdate {
  type: 'dashboard' | 'security' | 'cluster' | 'audit' | 'chat' | 'system';
  action: 'create' | 'update' | 'delete' | 'alert' | 'status_change';
  data: any;
  timestamp: string;
  source: string;
  severity?: 'low' | 'medium' | 'high' | 'critical';
}

export interface SystemNotification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: string;
  category: 'security' | 'cluster' | 'audit' | 'system' | 'user';
  persistent?: boolean;
  actionUrl?: string;
  metadata?: Record<string, unknown>;
}

export interface RealTimeSubscription {
  id: string;
  topics: string[];
  callback: (update: RealTimeUpdate) => void;
  filters?: {
    types?: string[];
    severity?: string[];
    sources?: string[];
  };
}

class RealTimeService {
  private subscriptions: Map<string, RealTimeSubscription> = new Map();
  private isConnected = false;
  private connectionListeners: ((connected: boolean) => void)[] = [];
  private notificationListeners: ((notification: SystemNotification) => void)[] = [];
  private wsUrl: string;

  constructor() {
    // Use the same WebSocket URL pattern as the existing service
    const protocol = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = typeof window !== 'undefined' ? window.location.host : 'localhost:30080';
    this.wsUrl = `${protocol}//${host}/ws`;

    this.initializeWebSocket();
  }

  // WebSocket Connection Management
  initializeWebSocket() {
    if (typeof window === 'undefined') {
      return; // Skip on server-side rendering
    }

    webSocketService.connect(
      (data) => this.handleWebSocketMessage(data),
      (error) => this.handleWebSocketError(error)
    );

    this.subscribeToConnectionStatus();
  }

  private handleWebSocketMessage(data: unknown) {
    try {
      const message = data as WebSocketMessage & RealTimeUpdate;

      // Handle different message types
      switch (message.type) {
        case 'dashboard':
          this.broadcastDashboardUpdate(message);
          break;
        case 'security':
          this.broadcastSecurityUpdate(message);
          break;
        case 'cluster':
          this.broadcastClusterUpdate(message);
          break;
        case 'audit':
          this.broadcastAuditUpdate(message);
          break;
        case 'chat':
          this.broadcastChatUpdate(message);
          break;
        case 'system':
          this.broadcastSystemUpdate(message);
          break;
        default:
          console.log('Unknown real-time message type:', message.type);
      }
    } catch (error) {
      console.error('Failed to handle WebSocket message:', error);
    }
  }

  private handleWebSocketError(error: Event) {
    console.error('WebSocket connection error:', error);
    this.isConnected = false;
    this.notifyConnectionListeners(false);
  }

  // Subscription Management
  subscribe(
    topics: string[],
    callback: (update: RealTimeUpdate) => void,
    filters?: RealTimeSubscription['filters']
  ): string {
    const id = `subscription-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;

    const subscription: RealTimeSubscription = {
      id,
      topics,
      callback,
      filters
    };

    this.subscriptions.set(id, subscription);

    // Send subscription request to backend
    if (this.isConnected) {
      webSocketService.send({
        type: 'SUBSCRIBE',
        topics,
        filters,
        subscriptionId: id
      });
    }

    return id;
  }

  unsubscribe(subscriptionId: string): void {
    const subscription = this.subscriptions.get(subscriptionId);
    if (subscription) {
      // Send unsubscribe request to backend
      if (this.isConnected) {
        webSocketService.send({
          type: 'UNSUBSCRIBE',
          subscriptionId
        });
      }

      this.subscriptions.delete(subscriptionId);
    }
  }

  // Helper method to check WebSocket connection
  private checkWebSocketConnection(): boolean {
    // Since we don't have direct access to WebSocket state, we'll track it via onopen/onclose events
    // For now, we'll use a simple check
    try {
      return webSocketService !== null && typeof webSocketService.send === 'function';
    } catch {
      return false;
    }
  }

  // Connection Status Management
  private subscribeToConnectionStatus() {
    // Monitor WebSocket connection status
    setInterval(() => {
      const connected = this.checkWebSocketConnection();
      if (connected !== this.isConnected) {
        this.isConnected = connected;
        this.notifyConnectionListeners(connected);

        if (connected) {
          this.resubscribeAll();
        }
      }
    }, 1000);
  }

  private notifyConnectionListeners(connected: boolean) {
    this.connectionListeners.forEach(listener => {
      try {
        listener(connected);
      } catch (error) {
        console.error('Error in connection listener:', error);
      }
    });
  }

  onConnectionChange(listener: (connected: boolean) => void): () => void {
    this.connectionListeners.push(listener);

    // Return unsubscribe function
    return () => {
      const index = this.connectionListeners.indexOf(listener);
      if (index > -1) {
        this.connectionListeners.splice(index, 1);
      }
    };
  }

  // Resubscribe all subscriptions after reconnection
  private resubscribeAll() {
    for (const [id, subscription] of this.subscriptions) {
      webSocketService.send({
        type: 'SUBSCRIBE',
        topics: subscription.topics,
        filters: subscription.filters,
        subscriptionId: id
      });
    }
  }

  // Broadcast methods for different update types
  private broadcastDashboardUpdate(update: RealTimeUpdate) {
    this.broadcastToSubscribers(['dashboard', 'system'], update);
  }

  private broadcastSecurityUpdate(update: RealTimeUpdate) {
    this.broadcastToSubscribers(['security', 'system'], update);

    // Create system notification for critical security events
    if (update.severity === 'critical' || update.severity === 'high') {
      this.createSystemNotification({
        type: 'warning',
        title: 'Security Alert',
        message: `${update.action} detected: ${update.data?.title || 'Security event'}`,
        category: 'security',
        persistent: update.severity === 'critical'
      });
    }
  }

  private broadcastClusterUpdate(update: RealTimeUpdate) {
    this.broadcastToSubscribers(['cluster', 'system'], update);

    // Create system notification for cluster issues
    if (update.action === 'alert' && update.severity && ['high', 'critical'].includes(update.severity)) {
      this.createSystemNotification({
        type: 'error',
        title: 'Cluster Alert',
        message: `Cluster issue detected: ${update.data?.message || 'Cluster event'}`,
        category: 'cluster',
        persistent: true
      });
    }
  }

  private broadcastAuditUpdate(update: RealTimeUpdate) {
    this.broadcastToSubscribers(['audit', 'system'], update);
  }

  private broadcastChatUpdate(update: RealTimeUpdate) {
    this.broadcastToSubscribers(['chat'], update);

    // Handle command execution status updates (AC: 7)
    if (update.action === 'status_change' && update.data?.executionId) {
      this.broadcastToSubscribers(['command_execution'], update);
    }
  }

  private broadcastSystemUpdate(update: RealTimeUpdate) {
    this.broadcastToSubscribers(['system'], update);

    // Create system notification for important system events
    this.createSystemNotification({
      type: 'info',
      title: 'System Update',
      message: update.data?.message || 'System status changed',
      category: 'system'
    });
  }

  private broadcastToSubscribers(topics: string[], update: RealTimeUpdate) {
    for (const subscription of this.subscriptions.values()) {
      // Check if subscription matches topics
      const topicMatch = subscription.topics.some(topic => topics.includes(topic));
      if (!topicMatch) continue;

      // Apply filters if present
      if (subscription.filters) {
        if (subscription.filters.types && !subscription.filters.types.includes(update.type)) {
          continue;
        }
        if (subscription.filters.severity && update.severity && !subscription.filters.severity.includes(update.severity)) {
          continue;
        }
        if (subscription.filters.sources && !subscription.filters.sources.includes(update.source)) {
          continue;
        }
      }

      // Call subscription callback
      try {
        subscription.callback(update);
      } catch (error) {
        console.error('Error in subscription callback:', error);
      }
    }
  }

  // System Notifications
  private createSystemNotification(notification: Omit<SystemNotification, 'id' | 'timestamp'>) {
    const fullNotification: SystemNotification = {
      id: `notification-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
      timestamp: new Date().toISOString(),
      ...notification
    };

    this.notificationListeners.forEach(listener => {
      try {
        listener(fullNotification);
      } catch (error) {
        console.error('Error in notification listener:', error);
      }
    });
  }

  onSystemNotification(listener: (notification: SystemNotification) => void): () => void {
    this.notificationListeners.push(listener);

    // Return unsubscribe function
    return () => {
      const index = this.notificationListeners.indexOf(listener);
      if (index > -1) {
        this.notificationListeners.splice(index, 1);
      }
    };
  }

  // Command execution specific methods (AC: 7)
  subscribeToCommandExecution(
    executionId: string,
    callback: (update: RealTimeUpdate) => void
  ): string {
    return this.subscribe(['command_execution'], (update) => {
      if (update.data?.executionId === executionId) {
        callback(update);
      }
    });
  }

  subscribeToAllCommandExecutions(callback: (update: RealTimeUpdate) => void): string {
    return this.subscribe(['command_execution'], callback);
  }

  // Request real-time updates for a specific command execution
  requestCommandExecutionUpdates(executionId: string): void {
    if (this.isConnected) {
      webSocketService.send({
        type: 'SUBSCRIBE_COMMAND_EXECUTION',
        executionId,
        timestamp: new Date().toISOString()
      });
    }
  }

  // Manual trigger methods for testing and development
  async broadcastToSystem(message: any) {
    if (this.isConnected) {
      webSocketService.send({
        type: 'BROADCAST',
        data: message,
        timestamp: new Date().toISOString()
      });
    }
  }

  async broadcastToTopic(topic: string, message: any) {
    try {
      await api.websocket.broadcastToTopic(topic, message);
    } catch (error) {
      console.error('Failed to broadcast to topic:', error);
    }
  }

  async broadcastToUser(userId: string, message: any) {
    try {
      await api.websocket.broadcastToUser(userId, message);
    } catch (error) {
      console.error('Failed to broadcast to user:', error);
    }
  }

  // Utility methods
  isWebSocketConnected(): boolean {
    return this.isConnected;
  }

  getSubscriptionCount(): number {
    return this.subscriptions.size;
  }

  getActiveTopics(): string[] {
    const topics = new Set<string>();
    for (const subscription of this.subscriptions.values()) {
      subscription.topics.forEach(topic => topics.add(topic));
    }
    return Array.from(topics);
  }

  // WebSocket Health Check
  async checkWebSocketHealth() {
    try {
      const health = await api.websocket.getHealth();
      return health.data;
    } catch (error) {
      console.error('WebSocket health check failed:', error);
      return null;
    }
  }

  async getWebSocketMetrics() {
    try {
      const metrics = await api.websocket.getMetrics();
      return metrics.data;
    } catch (error) {
      console.error('Failed to get WebSocket metrics:', error);
      return null;
    }
  }
}

// React hook for using real-time updates
export function useRealTimeUpdates(
  topics: string[],
  filters?: RealTimeSubscription['filters']
) {
  const [updates, setUpdates] = useState<RealTimeUpdate[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [lastUpdate, setLastUpdate] = useState<RealTimeUpdate | null>(null);

  useEffect(() => {
    const handleUpdate = (update: RealTimeUpdate) => {
      setLastUpdate(update);
      setUpdates(prev => [...prev.slice(-49), update]); // Keep last 50 updates
    };

    const subscriptionId = realTimeService.subscribe(topics, handleUpdate, filters);

    const unsubscribeConnectionStatus = realTimeService.onConnectionChange(setIsConnected);

    return () => {
      realTimeService.unsubscribe(subscriptionId);
      unsubscribeConnectionStatus();
    };
  }, [topics.join(','), JSON.stringify(filters)]);

  const clearUpdates = useCallback(() => {
    setUpdates([]);
    setLastUpdate(null);
  }, []);

  return {
    updates,
    lastUpdate,
    isConnected,
    clearUpdates
  };
}

// React hook for command execution tracking (AC: 7)
export function useCommandExecutionTracking(executionId?: string) {
  const [status, setStatus] = useState<string>('pending');
  const [progress, setProgress] = useState<number>(0);
  const [phase, setPhase] = useState<string>('initializing');
  const [output, setOutput] = useState<string>('');
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!executionId) return;

    const handleUpdate = (update: RealTimeUpdate) => {
      if (update.data?.executionId === executionId) {
        if (update.data.status) setStatus(update.data.status);
        if (update.data.progress !== undefined) setProgress(update.data.progress);
        if (update.data.phase) setPhase(update.data.phase);
        if (update.data.output) setOutput(update.data.output);
        if (update.data.error) setError(update.data.error);
      }
    };

    const subscriptionId = realTimeService.subscribeToCommandExecution(executionId, handleUpdate);
    const unsubscribeConnectionStatus = realTimeService.onConnectionChange(setIsConnected);

    // Request real-time updates for this execution
    realTimeService.requestCommandExecutionUpdates(executionId);

    return () => {
      realTimeService.unsubscribe(subscriptionId);
      unsubscribeConnectionStatus();
    };
  }, [executionId]);

  return {
    status,
    progress,
    phase,
    output,
    error,
    isConnected
  };
}

// React hook for system notifications
export function useSystemNotifications() {
  const [notifications, setNotifications] = useState<SystemNotification[]>([]);

  useEffect(() => {
    const handleNotification = (notification: SystemNotification) => {
      setNotifications(prev => [...prev, notification]);
    };

    const unsubscribe = realTimeService.onSystemNotification(handleNotification);

    return unsubscribe;
  }, []);

  const dismissNotification = useCallback((notificationId: string) => {
    setNotifications(prev => prev.filter(n => n.id !== notificationId));
  }, []);

  const clearAllNotifications = useCallback(() => {
    setNotifications([]);
  }, []);

  return {
    notifications,
    dismissNotification,
    clearAllNotifications
  };
}

// Singleton instance
export const realTimeService = new RealTimeService();
export default realTimeService;
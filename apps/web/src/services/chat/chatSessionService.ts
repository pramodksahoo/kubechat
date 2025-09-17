// Authenticated Chat Session Service for Story 2.2
// Integrates with Story 2.1 authentication system and verified backend APIs

import { ChatMessage, ChatSession } from '../../types/chat';
import { api } from '../api';
import { useAuthStore } from '../../stores/authStore';
import { errorHandlingService } from '../errorHandlingService';

interface CreateSessionData {
  clusterId?: string;
  title?: string;
}

interface SessionFilters {
  status?: 'active' | 'completed' | 'archived';
  clusterId?: string;
  limit?: number;
  offset?: number;
}

export class ChatSessionService {
  private sessions: ChatSession[] = [];
  private currentSession: ChatSession | null = null;

  constructor() {
    this.loadLocalSessions();
  }

  // Task 1.1: Implement ChatSessionService using authenticated /api/v1/chat/sessions API
  async createSession(data: CreateSessionData = {}): Promise<ChatSession> {
    try {
      // Ensure user is authenticated
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated || !authState.user) {
        throw new Error('User must be authenticated to create chat sessions');
      }

      // Call authenticated backend API
      const response = await api.chat.createSession({
        clusterId: data.clusterId,
      });

      // Transform backend response to our ChatSession interface
      const sessionData = response.data;
      const session: ChatSession = {
        id: sessionData.data.id,
        userId: authState.user.id,
        title: sessionData.data.title || data.title || `Chat Session ${new Date().toLocaleString()}`,
        clusterId: data.clusterId,
        clusterName: data.clusterId ? await this.getClusterName(data.clusterId) : undefined,
        createdAt: sessionData.data.createdAt,
        updatedAt: sessionData.data.updatedAt,
        status: sessionData.data.isActive ? 'active' : 'completed',
        messageCount: sessionData.data.messageCount || 0,
      };

      // Store locally for offline access and quick loading
      this.sessions.unshift(session);
      this.currentSession = session;
      this.saveLocalSessions();

      return session;
    } catch (error) {
      const errorDetails = await errorHandlingService.handleError(error as Error, {
        context: {
          operation: 'create-chat-session',
          component: 'ChatSessionService',
        },
        logToConsole: true,
      });

      // Create fallback local session if backend fails
      const fallbackSession: ChatSession = {
        id: `local-session-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        userId: useAuthStore.getState().user?.id || 'unknown',
        title: data.title || `Offline Chat Session ${new Date().toLocaleString()}`,
        clusterId: data.clusterId,
        clusterName: data.clusterId ? 'Unknown Cluster' : undefined,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        status: 'active',
        messageCount: 0,
      };

      this.sessions.unshift(fallbackSession);
      this.currentSession = fallbackSession;
      this.saveLocalSessions();

      console.warn('Created offline session due to backend error:', errorDetails);
      return fallbackSession;
    }
  }

  // Task 1.2: Create chat session initialization with JWT token validation
  async initializeSession(sessionId?: string): Promise<ChatSession> {
    try {
      // Validate authentication
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated) {
        throw new Error('Authentication required for chat session access');
      }

      if (sessionId) {
        // Load specific session
        const session = await this.getSession(sessionId);
        this.currentSession = session;
        return session;
      } else {
        // Create new session if none specified
        return await this.createSession();
      }
    } catch (error) {
      console.error('Session initialization failed:', error);
      throw error;
    }
  }

  // Task 1.3: Implement session persistence and restoration after page refresh
  async restoreSession(): Promise<ChatSession | null> {
    try {
      // First, try to restore from localStorage
      const localSessions = this.getLocalSessions();
      if (localSessions.length > 0) {
        const lastActiveSession = localSessions.find(s => s.status === 'active') || localSessions[0];

        // Validate session with backend if online
        try {
          const validatedSession = await this.getSession(lastActiveSession.id);
          this.currentSession = validatedSession;
          return validatedSession;
        } catch {
          // Backend validation failed, use local session
          this.currentSession = lastActiveSession;
          return lastActiveSession;
        }
      }

      // No local sessions, try to load from backend
      const sessions = await this.getSessions({ limit: 1 });
      if (sessions.length > 0) {
        this.currentSession = sessions[0];
        return sessions[0];
      }

      return null;
    } catch (error) {
      console.error('Session restoration failed:', error);
      return null;
    }
  }

  // Get sessions with authentication
  async getSessions(filters: SessionFilters = {}): Promise<ChatSession[]> {
    try {
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated) {
        return this.getLocalSessions();
      }

      const response = await api.chat.getSessions();
      const backendSessions = response.data.sessions;

      // Transform backend sessions to our interface
      const sessions: ChatSession[] = backendSessions.map((s: any) => ({
        id: s.id,
        userId: s.userId,
        title: `Chat Session ${new Date(s.createdAt).toLocaleString()}`,
        clusterId: s.clusterId,
        clusterName: s.clusterId ? 'Connected Cluster' : undefined,
        createdAt: s.createdAt,
        updatedAt: s.updatedAt,
        status: s.status as 'active' | 'completed' | 'archived',
        messageCount: s.messageCount || 0,
      }));

      // Apply client-side filters
      let filteredSessions = sessions;
      if (filters.status) {
        filteredSessions = filteredSessions.filter(s => s.status === filters.status);
      }
      if (filters.clusterId) {
        filteredSessions = filteredSessions.filter(s => s.clusterId === filters.clusterId);
      }
      if (filters.limit) {
        filteredSessions = filteredSessions.slice(0, filters.limit);
      }

      // Update local cache
      this.sessions = sessions;
      this.saveLocalSessions();

      return filteredSessions;
    } catch (error) {
      console.error('Failed to load sessions from backend:', error);
      return this.getLocalSessions();
    }
  }

  // Get specific session with authentication
  async getSession(sessionId: string): Promise<ChatSession> {
    try {
      const response = await api.chat.getSession(sessionId);
      const sessionData = response.data;

      const session: ChatSession = {
        id: sessionData.id,
        userId: sessionData.userId,
        title: `Chat Session ${new Date(sessionData.createdAt).toLocaleString()}`,
        clusterId: sessionData.clusterId,
        clusterName: sessionData.clusterId ? await this.getClusterName(sessionData.clusterId) : undefined,
        createdAt: sessionData.createdAt,
        updatedAt: sessionData.updatedAt,
        status: sessionData.status as 'active' | 'completed' | 'archived',
        messageCount: sessionData.messageCount || 0,
      };

      // Update local cache
      const existingIndex = this.sessions.findIndex(s => s.id === sessionId);
      if (existingIndex >= 0) {
        this.sessions[existingIndex] = session;
      } else {
        this.sessions.push(session);
      }
      this.saveLocalSessions();

      return session;
    } catch (error) {
      // Try local cache as fallback
      const localSession = this.sessions.find(s => s.id === sessionId);
      if (localSession) {
        return localSession;
      }
      throw new Error(`Session ${sessionId} not found`);
    }
  }

  // Task 1.4: Add session management for multiple concurrent chat conversations
  async switchToSession(sessionId: string): Promise<ChatSession> {
    const session = await this.getSession(sessionId);
    this.currentSession = session;
    return session;
  }

  async getActiveSessions(): Promise<ChatSession[]> {
    return this.getSessions({ status: 'active' });
  }

  getCurrentSession(): ChatSession | null {
    return this.currentSession;
  }

  // Task 1.5: Implement chat session cleanup and termination workflows
  async terminateSession(sessionId: string): Promise<void> {
    try {
      // Mark session as completed in backend
      await api.chat.updateSession(sessionId, { status: 'completed' });

      // Update local state
      const sessionIndex = this.sessions.findIndex(s => s.id === sessionId);
      if (sessionIndex >= 0) {
        this.sessions[sessionIndex].status = 'completed';
        this.sessions[sessionIndex].updatedAt = new Date().toISOString();
      }

      // If this was the current session, clear it
      if (this.currentSession?.id === sessionId) {
        this.currentSession = null;
      }

      this.saveLocalSessions();
    } catch (error) {
      console.error('Failed to terminate session:', error);
      throw error;
    }
  }

  async deleteSession(sessionId: string): Promise<void> {
    try {
      // Delete from backend
      await api.chat.deleteSession(sessionId);

      // Remove from local state
      this.sessions = this.sessions.filter(s => s.id !== sessionId);

      // Clear current session if it was deleted
      if (this.currentSession?.id === sessionId) {
        this.currentSession = null;
      }

      this.saveLocalSessions();
      this.clearLocalMessages(sessionId);
    } catch (error) {
      console.error('Failed to delete session:', error);
      throw error;
    }
  }

  // Task 1.6: Add session sharing and collaboration features for team workflows
  async shareSession(sessionId: string, userIds: string[]): Promise<{ shareId: string; shareUrl: string }> {
    try {
      // Create share link through backend API
      const response = await api.chat.shareSession(sessionId, { userIds });
      const shareData = response.data as any;

      return {
        shareId: shareData.shareId || 'share-' + Date.now(),
        shareUrl: shareData.shareUrl || `${window.location.origin}/chat/shared/${shareData.shareId}`,
      };
    } catch (error) {
      console.error('Failed to share session:', error);
      throw error;
    }
  }

  async joinSharedSession(shareId: string): Promise<ChatSession> {
    try {
      const response = await api.chat.joinSharedSession(shareId);
      const sessionData = response.data as any;

      const session: ChatSession = {
        id: sessionData.session?.id || 'shared-' + Date.now(),
        userId: sessionData.session?.userId || 'unknown',
        title: sessionData.session?.title || `Shared Session ${new Date().toLocaleString()}`,
        clusterId: sessionData.session?.clusterId,
        clusterName: sessionData.session?.clusterName,
        createdAt: sessionData.session?.createdAt || new Date().toISOString(),
        updatedAt: sessionData.session?.updatedAt || new Date().toISOString(),
        status: 'active',
        messageCount: sessionData.session?.messageCount || 0,
      };

      // Add to local sessions
      this.sessions.unshift(session);
      this.currentSession = session;
      this.saveLocalSessions();

      return session;
    } catch (error) {
      console.error('Failed to join shared session:', error);
      throw error;
    }
  }

  // Local storage management for offline functionality
  private saveLocalSessions(): void {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        const data = {
          sessions: this.sessions,
          currentSessionId: this.currentSession?.id || null,
          lastUpdated: new Date().toISOString(),
        };
        localStorage.setItem('kubechat_chat_sessions', JSON.stringify(data));
      }
    } catch (error) {
      console.error('Failed to save sessions locally:', error);
    }
  }

  private loadLocalSessions(): void {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        const stored = localStorage.getItem('kubechat_chat_sessions');
        if (stored) {
          const data = JSON.parse(stored);
          this.sessions = data.sessions || [];

          // Restore current session
          if (data.currentSessionId) {
            this.currentSession = this.sessions.find(s => s.id === data.currentSessionId) || null;
          }
        }
      }
    } catch (error) {
      console.error('Failed to load local sessions:', error);
    }
  }

  getLocalSessions(): ChatSession[] {
    return [...this.sessions];
  }

  clearLocalSessions(): void {
    this.sessions = [];
    this.currentSession = null;

    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.removeItem('kubechat_chat_sessions');
    }
  }

  private clearLocalMessages(sessionId: string): void {
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.removeItem(`kubechat_messages_${sessionId}`);
    }
  }

  // Helper method to get cluster name
  private async getClusterName(clusterId: string): Promise<string> {
    try {
      const response = await api.clusters.getCluster();
      const clusterData = response.data as any;
      return clusterData?.name || `Cluster ${clusterId}`;
    } catch {
      return `Cluster ${clusterId}`;
    }
  }
}

// Singleton instance
export const chatSessionService = new ChatSessionService();
import { ChatMessage, ChatSession, CommandPreview, CommandExecution } from '../types/chat';

export class ChatService {
  private baseUrl: string;

  constructor(baseUrl: string = process.env.NEXT_PUBLIC_API_URL || '/api/v1') {
    this.baseUrl = baseUrl;
  }

  // Session Management
  async createSession(clusterId?: string): Promise<ChatSession> {
    const response = await fetch(`${this.baseUrl}/chat/sessions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
      body: JSON.stringify({ clusterId }),
    });

    if (!response.ok) {
      throw new Error(`Failed to create chat session: ${response.statusText}`);
    }

    return response.json();
  }

  async getSessions(): Promise<ChatSession[]> {
    const response = await fetch(`${this.baseUrl}/chat/sessions`, {
      headers: {
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch chat sessions: ${response.statusText}`);
    }

    return response.json();
  }

  async getSession(sessionId: string): Promise<ChatSession> {
    const response = await fetch(`${this.baseUrl}/chat/sessions/${sessionId}`, {
      headers: {
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch chat session: ${response.statusText}`);
    }

    return response.json();
  }

  async deleteSession(sessionId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/chat/sessions/${sessionId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to delete chat session: ${response.statusText}`);
    }
  }

  // Message Management
  async getMessages(sessionId: string, limit = 50, offset = 0): Promise<ChatMessage[]> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });

    const response = await fetch(`${this.baseUrl}/chat/sessions/${sessionId}/messages?${params}`, {
      headers: {
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch messages: ${response.statusText}`);
    }

    return response.json();
  }

  async sendMessage(sessionId: string, content: string): Promise<ChatMessage> {
    const response = await fetch(`${this.baseUrl}/chat/sessions/${sessionId}/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
      body: JSON.stringify({
        content,
        type: 'user',
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to send message: ${response.statusText}`);
    }

    return response.json();
  }

  // Command Management
  async generateCommandPreview(naturalLanguage: string, sessionId: string): Promise<CommandPreview> {
    const response = await fetch(`${this.baseUrl}/commands/preview`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
      body: JSON.stringify({
        naturalLanguage,
        sessionId,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to generate command preview: ${response.statusText}`);
    }

    return response.json();
  }

  async executeCommand(previewId: string): Promise<CommandExecution> {
    const response = await fetch(`${this.baseUrl}/commands/execute`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
      body: JSON.stringify({ previewId }),
    });

    if (!response.ok) {
      throw new Error(`Failed to execute command: ${response.statusText}`);
    }

    return response.json();
  }

  async requestApproval(previewId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/commands/approve`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
      body: JSON.stringify({ previewId }),
    });

    if (!response.ok) {
      throw new Error(`Failed to request approval: ${response.statusText}`);
    }
  }

  async getCommandHistory(sessionId?: string, limit = 20): Promise<CommandExecution[]> {
    const params = new URLSearchParams({
      limit: limit.toString(),
    });

    if (sessionId) {
      params.append('sessionId', sessionId);
    }

    const response = await fetch(`${this.baseUrl}/commands/history?${params}`, {
      headers: {
        'Authorization': `Bearer ${this.getAuthToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch command history: ${response.statusText}`);
    }

    return response.json();
  }

  // WebSocket for real-time updates
  createWebSocket(sessionId: string): WebSocket {
    const wsUrl = this.baseUrl.replace(/^http/, 'ws');
    const token = this.getAuthToken();
    return new WebSocket(`${wsUrl}/chat/sessions/${sessionId}/ws?token=${token}`);
  }

  // Local Storage for offline persistence
  saveSessionLocally(session: ChatSession): void {
    const sessions = this.getLocalSessions();
    const existingIndex = sessions.findIndex(s => s.id === session.id);
    
    if (existingIndex >= 0) {
      sessions[existingIndex] = session;
    } else {
      sessions.push(session);
    }

    localStorage.setItem('kubechat_sessions', JSON.stringify(sessions));
  }

  saveMessagesLocally(sessionId: string, messages: ChatMessage[]): void {
    localStorage.setItem(`kubechat_messages_${sessionId}`, JSON.stringify(messages));
  }

  getLocalSessions(): ChatSession[] {
    try {
      const stored = localStorage.getItem('kubechat_sessions');
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error('Failed to parse local sessions:', error);
      return [];
    }
  }

  getLocalMessages(sessionId: string): ChatMessage[] {
    try {
      const stored = localStorage.getItem(`kubechat_messages_${sessionId}`);
      return stored ? JSON.parse(stored) : [];
    } catch (error) {
      console.error('Failed to parse local messages:', error);
      return [];
    }
  }

  clearLocalData(): void {
    const keys = Object.keys(localStorage).filter(key => 
      key.startsWith('kubechat_')
    );
    keys.forEach(key => localStorage.removeItem(key));
  }

  private getAuthToken(): string {
    // In a real implementation, this would get the token from auth service
    return localStorage.getItem('auth_token') || '';
  }
}

export const chatService = new ChatService();
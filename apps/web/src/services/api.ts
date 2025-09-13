// KubeChat API Service Layer for Kubernetes Service-to-Service Communication
// Uses relative URLs that get proxied to backend service via Next.js rewrites

import { withRetry, withCircuitBreaker } from '../lib/resilience';

// API Configuration
const API_CONFIG = {
  baseURL: '', // Use relative URLs for Kubernetes service communication
  timeout: 30000,
  retries: 3,
  retryDelay: 1000,
} as const;

// API Response Types
export interface ApiResponse<T = unknown> {
  data: T;
  success: boolean;
  message?: string;
  error?: string;
}

export interface ApiError {
  message: string;
  status: number;
  code?: string;
  details?: unknown;
}

// HTTP Client with retry and circuit breaker patterns
class HttpClient {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${API_CONFIG.baseURL}${endpoint}`;
    
    const defaultHeaders = {
      'Content-Type': 'application/json',
      ...this.getAuthHeaders(),
    };

    const config: RequestInit = {
      ...options,
      headers: {
        ...defaultHeaders,
        ...options.headers,
      },
      signal: AbortSignal.timeout(API_CONFIG.timeout),
    };

    const response = await fetch(url, config);

    if (!response.ok) {
      const error: ApiError = {
        message: response.statusText || 'API request failed',
        status: response.status,
      };

      try {
        const errorData = await response.json();
        error.message = errorData.message || error.message;
        error.code = errorData.code;
        error.details = errorData.details;
      } catch {
        // Use default error message if response body is not JSON
      }

      throw error;
    }

    const data = await response.json();
    return data;
  }

  private getAuthHeaders(): Record<string, string> {
    // Get auth token from localStorage, cookies, or auth context
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('auth_token');
      if (token) {
        return { Authorization: `Bearer ${token}` };
      }
    }
    return {};
  }

  // HTTP Methods with resilience patterns
  async get<T>(endpoint: string, options?: RequestInit): Promise<ApiResponse<T>> {
    const requestWithResilience = withCircuitBreaker(
      () => withRetry(
        () => this.request<T>(endpoint, { ...options, method: 'GET' }),
        { maxAttempts: API_CONFIG.retries, baseDelay: API_CONFIG.retryDelay }
      ),
      `api-${endpoint}`
    );

    return requestWithResilience();
  }

  async post<T>(endpoint: string, body?: unknown, options?: RequestInit): Promise<ApiResponse<T>> {
    const requestWithResilience = withCircuitBreaker(
      () => withRetry(
        () => this.request<T>(endpoint, {
          ...options,
          method: 'POST',
          body: body ? JSON.stringify(body) : undefined,
        }),
        { maxAttempts: API_CONFIG.retries, baseDelay: API_CONFIG.retryDelay }
      ),
      `api-${endpoint}`
    );

    return requestWithResilience();
  }

  async put<T>(endpoint: string, body?: unknown, options?: RequestInit): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async delete<T>(endpoint: string, options?: RequestInit): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

// Singleton HTTP client instance
export const httpClient = new HttpClient();

// Health Check API
export const healthApi = {
  check: () => httpClient.get<{ status: string; timestamp: number }>('/health'),
};

// Authentication API
export const authApi = {
  login: (credentials: { username: string; password: string }) =>
    httpClient.post<{ user: unknown; token: string }>('/api/auth/login', credentials),
  
  logout: () => httpClient.post('/api/auth/logout'),
  
  me: () => httpClient.get<{ id: string; username: string; email: string; role: string }>('/api/auth/me'),
  
  refresh: () => httpClient.post<{ token: string }>('/api/auth/refresh'),
};

// Commands API
export const commandsApi = {
  execute: (command: { command: string; cluster?: string }) =>
    httpClient.post<{
      id: string;
      command: string;
      status: string;
      output: string;
      exitCode: number | null;
      executedAt: number;
    }>('/api/commands/execute', command),

  getExecution: (id: string) =>
    httpClient.get<{
      id: string;
      command: string;
      status: string;
      output: string;
      exitCode: number | null;
      executedAt: number;
      completedAt?: number;
    }>(`/api/commands/${id}`),

  listExecutions: (params?: { page?: number; limit?: number; status?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.status) searchParams.set('status', params.status);
    
    const query = searchParams.toString();
    return httpClient.get<{
      executions: unknown[];
      pagination: { page: number; limit: number; total: number; pages: number };
    }>(`/api/commands/executions${query ? `?${query}` : ''}`);
  },

  approve: (id: string) =>
    httpClient.post(`/api/commands/${id}/approve`),

  reject: (id: string, reason?: string) =>
    httpClient.post(`/api/commands/${id}/reject`, { reason }),
};

// Clusters API
export const clustersApi = {
  list: () =>
    httpClient.get<{
      clusters: Array<{
        id: string;
        name: string;
        status: string;
        version: string;
        nodes: number;
        pods: number;
      }>;
    }>('/api/kubernetes/clusters'),

  getNamespaces: (clusterId?: string) => {
    const params = clusterId ? `?cluster=${clusterId}` : '';
    return httpClient.get<{ namespaces: string[] }>(`/api/kubernetes/namespaces${params}`);
  },

  getResources: (type: string, namespace?: string) => {
    const params = new URLSearchParams();
    if (namespace) params.set('namespace', namespace);
    const query = params.toString();
    return httpClient.get(`/api/kubernetes/resources/${type}${query ? `?${query}` : ''}`);
  },
};

// Audit API
export const auditApi = {
  getEvents: (params?: { page?: number; limit?: number; startTime?: string; endTime?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.startTime) searchParams.set('startTime', params.startTime);
    if (params?.endTime) searchParams.set('endTime', params.endTime);
    
    const query = searchParams.toString();
    return httpClient.get<{
      events: Array<{
        id: string;
        timestamp: number;
        user: string;
        action: string;
        resource: string;
        status: string;
        ip: string;
      }>;
      pagination: { page: number; limit: number; total: number; pages: number };
    }>(`/api/audit/events${query ? `?${query}` : ''}`);
  },
};

// WebSocket Service for real-time updates
export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private pingInterval: NodeJS.Timeout | null = null;
  private isManualClose = false;

  connect(onMessage?: (data: unknown) => void, onError?: (error: Event) => void) {
    // In Kubernetes, WebSocket connections go directly to the backend service via ingress
    // The ingress controller handles WebSocket upgrades to the backend service
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/commands/ws`;
    
    this.ws = new WebSocket(wsUrl);
    this.isManualClose = false;

    this.ws.onopen = () => {
      console.log('WebSocket connected to backend service');
      this.reconnectAttempts = 0;
      this.startPing();
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage?.(data);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      onError?.(error);
    };

    this.ws.onclose = () => {
      this.stopPing();
      
      if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
        console.log(`WebSocket disconnected. Attempting to reconnect... (${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`);
        setTimeout(() => {
          this.reconnectAttempts++;
          this.connect(onMessage, onError);
        }, this.reconnectDelay * Math.pow(2, this.reconnectAttempts)); // Exponential backoff
      }
    };
  }

  send(message: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  close() {
    this.isManualClose = true;
    this.stopPing();
    this.ws?.close();
    this.ws = null;
  }

  private startPing() {
    this.pingInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send({ type: 'PING', timestamp: Date.now() });
      }
    }, 30000); // Ping every 30 seconds
  }

  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }
}

// Singleton WebSocket service instance
export const webSocketService = new WebSocketService();

// Combined API object for easy importing
export const api = {
  health: healthApi,
  auth: authApi,
  commands: commandsApi,
  clusters: clustersApi,
  audit: auditApi,
  ws: webSocketService,
};
// KubeChat API Service Layer for Kubernetes Service-to-Service Communication
// Uses relative URLs that get proxied to backend service via Next.js rewrites

import { withRetry, withCircuitBreaker } from '../lib/resilience';

// API Configuration - Using environment variables for Kubernetes-native configuration
const API_CONFIG = {
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || '', // Empty for relative URLs in K8s
  timeout: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT || '30000'),
  retries: parseInt(process.env.NEXT_PUBLIC_API_RETRIES || '3'),
  retryDelay: parseInt(process.env.NEXT_PUBLIC_API_RETRY_DELAY || '1000'),
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
  live: () => httpClient.get<{ status: string }>('/health/live'),
  status: () => httpClient.get('/status'),
};

// Authentication API - Updated to match backend endpoints
export const authApi = {
  login: (credentials: { username: string; password: string }) =>
    httpClient.post<{ user: unknown; token: string }>('/auth/login', credentials),

  logout: () => httpClient.post('/auth/logout'),

  register: (userData: { username: string; email: string; password: string }) =>
    httpClient.post('/auth/register', userData),

  me: () => httpClient.get<{ id: string; username: string; email: string; role: string }>('/auth/me'),

  profile: () => httpClient.get('/auth/profile'),

  refresh: () => httpClient.post<{ token: string }>('/auth/refresh'),

  // Admin user management
  getUsers: () => httpClient.get('/auth/admin/users'),

  getUser: (id: string) => httpClient.get(`/auth/admin/users/${id}`),
};

// Commands API - Updated to match backend endpoints
export const commandsApi = {
  execute: (command: { command: string; cluster?: string }) =>
    httpClient.post<{
      id: string;
      command: string;
      status: string;
      output: string;
      exitCode: number | null;
      executedAt: number;
    }>('/api/v1/commands/execute', command),

  getExecution: (id: string) =>
    httpClient.get<{
      id: string;
      command: string;
      status: string;
      output: string;
      exitCode: number | null;
      executedAt: number;
      completedAt?: number;
    }>(`/api/v1/commands/executions/${id}`),

  listExecutions: (params?: { page?: number; limit?: number; status?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.status) searchParams.set('status', params.status);

    const query = searchParams.toString();
    return httpClient.get<{
      executions: unknown[];
      pagination: { page: number; limit: number; total: number; pages: number };
    }>(`/api/v1/commands/executions${query ? `?${query}` : ''}`);
  },

  // Command approval endpoints
  getPendingApprovals: () =>
    httpClient.get('/api/v1/commands/approvals/pending'),

  approve: (approvalData: unknown) =>
    httpClient.post('/api/v1/commands/approve', approvalData),

  // Command stats and health
  getStats: () => httpClient.get('/api/v1/commands/stats'),

  getHealth: () => httpClient.get('/api/v1/commands/health'),

  // Rollback operations
  createRollbackPlan: (executionId: string) =>
    httpClient.post(`/api/v1/commands/executions/${executionId}/rollback/plan`),

  validateRollback: (executionId: string) =>
    httpClient.get(`/api/v1/commands/executions/${executionId}/rollback/validate`),

  executeRollback: (planId: string) =>
    httpClient.post(`/api/v1/rollback/plans/${planId}/execute`),

  getRollbackStatus: (rollbackId: string) =>
    httpClient.get(`/api/v1/rollback/executions/${rollbackId}/status`),
};

// Kubernetes/Clusters API - Updated to match backend endpoints
export const clustersApi = {
  // Cluster information
  getCluster: () => httpClient.get('/kubernetes/cluster'),

  getHealth: () => httpClient.get('/kubernetes/health'),

  validate: (data: unknown) => httpClient.post('/kubernetes/validate', data),

  // Namespace operations
  getNamespaces: () => httpClient.get('/kubernetes/namespaces'),

  // Pod operations
  getPods: (namespace: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/pods`),

  getPod: (namespace: string, name: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/pods/${name}`),

  deletePod: (namespace: string, name: string) =>
    httpClient.delete(`/kubernetes/namespaces/${namespace}/pods/${name}`),

  getPodLogs: (namespace: string, name: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/pods/${name}/logs`),

  // Deployment operations
  getDeployments: (namespace: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/deployments`),

  getDeployment: (namespace: string, name: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/deployments/${name}`),

  restartDeployment: (namespace: string, name: string) =>
    httpClient.post(`/kubernetes/namespaces/${namespace}/deployments/${name}/restart`),

  scaleDeployment: (namespace: string, name: string, replicas: number) =>
    httpClient.put(`/kubernetes/namespaces/${namespace}/deployments/${name}/scale`, { replicas }),

  // Service operations
  getServices: (namespace: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/services`),

  getService: (namespace: string, name: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/services/${name}`),

  // ConfigMap operations
  getConfigMaps: (namespace: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/configmaps`),

  getConfigMap: (namespace: string, name: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/configmaps/${name}`),

  // Secret operations
  getSecrets: (namespace: string) =>
    httpClient.get(`/kubernetes/namespaces/${namespace}/secrets`),
};

// Audit API - Updated to match backend endpoints
export const auditApi = {
  // Audit logs
  getLogs: (params?: { page?: number; limit?: number; startTime?: string; endTime?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.startTime) searchParams.set('startTime', params.startTime);
    if (params?.endTime) searchParams.set('endTime', params.endTime);

    const query = searchParams.toString();
    return httpClient.get(`/audit/logs${query ? `?${query}` : ''}`);
  },

  createLog: (logData: unknown) => httpClient.post('/audit/logs', logData),

  getLog: (id: string) => httpClient.get(`/audit/logs/${id}`),

  // Audit metrics and summaries
  getMetrics: () => httpClient.get('/audit/metrics'),

  getSummary: () => httpClient.get('/audit/summary'),

  // Specialized audit queries
  getDangerousCommands: () => httpClient.get('/audit/dangerous'),

  getFailedCommands: () => httpClient.get('/audit/failed'),

  // Audit system health
  getHealth: () => httpClient.get('/audit/health'),

  // Integrity verification
  verifyIntegrity: () => httpClient.post('/audit/verify-integrity'),
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
    // In Kubernetes, WebSocket connections use relative URLs resolved by ingress
    // The ingress controller handles WebSocket upgrades to the backend service
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`; // Match backend /ws endpoint

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

// Security API
export const securityApi = {
  getAlerts: () => httpClient.get('/security/alerts'),

  getEvents: () => httpClient.get('/security/events'),

  createEvent: (eventData: unknown) => httpClient.post('/security/events', eventData),

  analyzeRequest: (requestData: unknown) => httpClient.post('/security/analyze-request', requestData),

  generateToken: (tokenData: unknown) => httpClient.post('/security/generate-token', tokenData),

  hashPassword: (password: string) => httpClient.post('/security/hash-password', { password }),

  validatePassword: (passwordData: unknown) => httpClient.post('/security/validate-password', passwordData),

  validateHeaders: (headers: unknown) => httpClient.post('/security/validate-headers', headers),

  scan: (scanData: unknown) => httpClient.post('/security/scan', scanData),

  getRateLimit: (identifier: string) => httpClient.get(`/security/rate-limit/${identifier}`),

  checkRateLimit: (identifier: string, data: unknown) =>
    httpClient.post(`/security/rate-limit/${identifier}/check`, data),

  deleteRateLimit: (identifier: string) => httpClient.delete(`/security/rate-limit/${identifier}`),

  cleanupSessions: () => httpClient.delete('/security/sessions/cleanup'),

  getHealth: () => httpClient.get('/security/health'),
};

// NLP API
export const nlpApi = {
  process: (text: string, options?: Record<string, unknown>) =>
    httpClient.post('/nlp/process', { text, ...(options || {}) }),

  classify: (text: string) => httpClient.post('/nlp/classify', { text }),

  validate: (data: unknown) => httpClient.post('/nlp/validate', data),

  getProviders: () => httpClient.get('/nlp/providers'),

  getMetrics: () => httpClient.get('/nlp/metrics'),

  getHealth: () => httpClient.get('/nlp/health'),
};

// Database API
export const databaseApi = {
  getHealth: () => httpClient.get('/database/health'),

  getMetrics: () => httpClient.get('/database/metrics'),

  // Cluster management
  getClusters: () => httpClient.get('/database/clusters'),

  createCluster: (clusterData: unknown) => httpClient.post('/database/clusters', clusterData),

  getCluster: (clusterId: string) => httpClient.get(`/database/clusters/${clusterId}`),

  getClusterHealth: (clusterId: string) => httpClient.get(`/database/clusters/${clusterId}/health`),

  triggerFailover: (clusterId: string) => httpClient.post(`/database/clusters/${clusterId}/failover`),

  getFailoverStatus: (clusterId: string) => httpClient.get(`/database/clusters/${clusterId}/failover-status`),

  // Connection management
  getReadConnection: (clusterId: string) => httpClient.get(`/database/clusters/${clusterId}/connection/read`),

  getWriteConnection: (clusterId: string) => httpClient.get(`/database/clusters/${clusterId}/connection/write`),

  // Backup operations
  createBackup: (backupData: unknown) => httpClient.post('/database/backups', backupData),

  getBackup: (backupId: string) => httpClient.get(`/database/backups/${backupId}`),

  restoreBackup: (backupId: string) => httpClient.post(`/database/backups/${backupId}/restore`),

  // Migration management
  getMigrations: () => httpClient.get('/database/migrations'),

  runMigration: (migrationData: unknown) => httpClient.post('/database/migrations', migrationData),

  // Pool management
  getPools: () => httpClient.get('/database/pools'),

  createPool: (poolData: unknown) => httpClient.post('/database/pools', poolData),

  getPool: (poolId: string) => httpClient.get(`/database/pools/${poolId}`),

  deletePool: (poolId: string) => httpClient.delete(`/database/pools/${poolId}`),

  getPoolStats: (poolId: string) => httpClient.get(`/database/pools/${poolId}/stats`),

  queryPool: (poolId: string, query: unknown) => httpClient.post(`/database/pools/${poolId}/query`, query),

  startTransaction: (poolId: string, transactionData: unknown) =>
    httpClient.post(`/database/pools/${poolId}/transaction`, transactionData),
};

// Performance API
export const performanceApi = {
  getHealth: () => httpClient.get('/performance/health'),

  getMetrics: () => httpClient.get('/performance/metrics'),

  // Cache operations
  getCache: (key: string) => httpClient.get(`/performance/cache/${key}`),

  setCache: (key: string, value: unknown) => httpClient.post(`/performance/cache/${key}`, value),

  deleteCache: (key: string) => httpClient.delete(`/performance/cache/${key}`),

  getCacheStats: () => httpClient.get('/performance/cache/stats'),
};

// WebSocket API
export const webSocketApi = {
  getHealth: () => httpClient.get('/websocket/health'),

  getMetrics: () => httpClient.get('/websocket/metrics'),

  getClients: () => httpClient.get('/websocket/clients'),

  getClientCount: () => httpClient.get('/websocket/clients/count'),

  deleteClient: (id: string) => httpClient.delete(`/websocket/clients/${id}`),

  getClientSubscriptions: (id: string) => httpClient.get(`/websocket/clients/${id}/subscriptions`),

  getTopicSubscriptions: (topic: string) => httpClient.get(`/websocket/subscriptions/${topic}`),

  broadcast: (message: unknown) => httpClient.post('/websocket/broadcast', message),

  broadcastToTopic: (topic: string, message: unknown) =>
    httpClient.post('/websocket/broadcast/topics', { topic, message }),

  broadcastToUser: (userId: string, message: unknown) =>
    httpClient.post(`/websocket/broadcast/user/${userId}`, message),

  notifySystem: (notification: unknown) => httpClient.post('/websocket/notify/system', notification),

  notifyUser: (userId: string, notification: unknown) =>
    httpClient.post(`/websocket/notify/user/${userId}`, notification),
};

// Combined API object for easy importing
export const api = {
  health: healthApi,
  auth: authApi,
  commands: commandsApi,
  clusters: clustersApi,
  audit: auditApi,
  security: securityApi,
  nlp: nlpApi,
  database: databaseApi,
  performance: performanceApi,
  websocket: webSocketApi,
  ws: webSocketService,
};
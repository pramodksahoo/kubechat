// KubeChat Application Constants

// API Configuration
export const API_ROUTES = {
  AUTH: '/api/auth',
  USERS: '/api/users',
  KUBERNETES: '/api/kubernetes',
  CHAT: '/api/chat',
  HEALTH: '/api/health',
} as const;

// HTTP Status Codes
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500,
} as const;

// User Roles
export const USER_ROLES = {
  ADMIN: 'admin',
  USER: 'user',
  VIEWER: 'viewer',
} as const;

// Application Constants
export const APP_CONFIG = {
  NAME: 'KubeChat',
  VERSION: '1.0.0',
  DESCRIPTION: 'Kubernetes Natural Language Interface',
  DEFAULT_NAMESPACE: 'default',
  MAX_CHAT_HISTORY: 100,
  SESSION_TIMEOUT: 24 * 60 * 60 * 1000, // 24 hours in milliseconds
} as const;

// Kubernetes Resource Types
export const KUBERNETES_RESOURCES = {
  POD: 'Pod',
  SERVICE: 'Service',
  DEPLOYMENT: 'Deployment',
  CONFIGMAP: 'ConfigMap',
  SECRET: 'Secret',
  NAMESPACE: 'Namespace',
  INGRESS: 'Ingress',
} as const;

// Query Types
export const QUERY_TYPES = {
  LIST: 'list',
  GET: 'get',
  CREATE: 'create',
  UPDATE: 'update',
  DELETE: 'delete',
  SCALE: 'scale',
  LOGS: 'logs',
} as const;

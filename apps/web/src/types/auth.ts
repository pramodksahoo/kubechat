// Authentication types for KubeChat frontend
// Following coding standards from docs/architecture/coding-standards.md

export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  permissions?: string[];
  createdAt: string;
  lastLoginAt?: string;
}

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken?: string;
  expiresAt: number;
}

export interface AuthState {
  // State properties
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  clearError: () => void;
  setUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
}

export interface AuthError {
  message: string;
  code?: string;
  field?: string;
}

// JWT token payload interface
export interface JWTPayload {
  sub: string; // user ID
  username: string;
  email: string;
  role: string;
  iat: number;
  exp: number;
}

// Login response from API
export interface LoginResponse {
  user: User;
  token: string;
  refreshToken?: string;
  expiresIn: number;
}

// Registration response from API
export interface RegisterResponse {
  user: User;
  message: string;
  requiresVerification?: boolean;
}
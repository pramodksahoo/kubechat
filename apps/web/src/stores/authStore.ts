// Authentication store using Zustand
// Following coding standards from docs/architecture/coding-standards.md

import { create } from 'zustand';
import { api } from '../services/api';
import { tokenService, createAuthTokens } from '../services/tokenService';
import type { AuthState, LoginCredentials, RegisterData, User, AuthTokens } from '../types/auth';

// Helper function to get initial auth state from sessionStorage during SSR
const getInitialAuthState = (): Pick<AuthState, 'isAuthenticated' | 'user' | 'tokens'> => {
  // During SSR, always return unauthenticated state
  if (typeof window === 'undefined') {
    return {
      isAuthenticated: false,
      user: null,
      tokens: null
    };
  }

  // During client-side hydration, check for existing token
  try {
    const token = sessionStorage.getItem('kubechat_auth_token');
    if (token && !tokenService.isTokenExpired(token)) {
      // If we have a valid token, assume authenticated state initially
      // The actual user data will be loaded by initializeAuth
      return {
        isAuthenticated: true,
        user: null, // Will be loaded by initializeAuth
        tokens: createAuthTokens(token)
      };
    }
  } catch (error) {
    console.warn('Failed to read token from sessionStorage:', error);
  }

  return {
    isAuthenticated: false,
    user: null,
    tokens: null
  };
};

export const useAuthStore = create<AuthState>()((set, get) => {
  const initialState = getInitialAuthState();
  
  return {
      // Initial state - SSR-safe with token check
      user: initialState.user,
      tokens: initialState.tokens,
      isAuthenticated: initialState.isAuthenticated,
      isLoading: false,
      error: null,

      // Login action
      login: async (credentials: LoginCredentials) => {
        set({ isLoading: true, error: null });

        try {
          const response = await api.auth.login(credentials);

          if (!response.data) {
            throw new Error('Invalid login response');
          }

          const { user, token } = response.data;
          const typedUser = user as User;

          // Validate token is present in response
          if (!token) {
            throw new Error('Token not found in login response');
          }

          // Create tokens object from response
          const tokens = createAuthTokens(token);

          // Store token in sessionStorage directly for persistence
          if (typeof window !== 'undefined') {
            try {
              sessionStorage.setItem('kubechat_auth_token', token);
            } catch (error) {
              console.warn('Failed to store token in sessionStorage:', error);
            }
          }

          set({
            user: typedUser,
            tokens,
            isAuthenticated: true,
            isLoading: false,
            error: null
          });

        } catch (error: any) {
          const errorMessage = error.message || 'Login failed';
          set({
            user: null,
            tokens: null,
            isAuthenticated: false,
            isLoading: false,
            error: errorMessage
          });
          throw error;
        }
      },

      // Register action
      register: async (data: RegisterData) => {
        set({ isLoading: true, error: null });

        try {
          if (data.password !== data.confirmPassword) {
            throw new Error('Passwords do not match');
          }

          const response = await api.auth.register({
            username: data.username,
            email: data.email,
            password: data.password
          });

          set({
            isLoading: false,
            error: null
          });

        } catch (error: any) {
          const errorMessage = error.message || 'Registration failed';
          set({
            isLoading: false,
            error: errorMessage
          });
          throw error;
        }
      },

      // Logout action
      logout: async () => {
        // Clear tokens from sessionStorage directly
        if (typeof window !== 'undefined') {
          try {
            sessionStorage.removeItem('kubechat_auth_token');
            sessionStorage.removeItem('kubechat_refresh_token');
          } catch (error) {
            console.warn('Failed to clear tokens from sessionStorage:', error);
          }
        }

        // Call logout API endpoint
        api.auth.logout().catch(console.error);

        set({
          user: null,
          tokens: null,
          isAuthenticated: false,
          isLoading: false,
          error: null
        });
      },

      // Refresh token action
      refreshToken: async () => {
        const currentTokens = get().tokens;
        
        if (!currentTokens) {
          await get().logout();
          return;
        }

        try {
          set({ isLoading: true });

          const response = await api.auth.refresh();

          if (!response.data || !response.data.token) {
            throw new Error('Token refresh failed - no token in response');
          }

          const { token } = response.data;
          
          // Store new token in sessionStorage directly
          if (typeof window !== 'undefined') {
            try {
              sessionStorage.setItem('kubechat_auth_token', token);
            } catch (error) {
              console.warn('Failed to store token in sessionStorage:', error);
            }
          }

          const tokens = createAuthTokens(token);

          set({
            tokens,
            isLoading: false,
            error: null
          });

        } catch (error) {
          console.error('Token refresh failed:', error);
          await get().logout();
        }
      },

      // Clear error action
      clearError: () => {
        set({ error: null });
      },

      // Set user action
      setUser: (user: User) => {
        set({ user });
      },

      // Set loading action
      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      }
    };
});

// Initialize auth store with existing token
export const initializeAuth = async (): Promise<boolean> => {
  // Skip during SSR
  if (typeof window === 'undefined') {
    return false;
  }

  // Get token from sessionStorage directly
  let token: string | null = null;
  try {
    token = sessionStorage.getItem('kubechat_auth_token');
  } catch (error) {
    console.warn('Failed to get token from sessionStorage:', error);
    return false;
  }

  if (!token) {
    // No token found, ensure auth state is cleared
    useAuthStore.setState({
      user: null,
      tokens: null,
      isAuthenticated: false,
      isLoading: false,
      error: null
    });
    return false;
  }

  if (tokenService.isTokenExpired(token)) {
    // Token expired, try to refresh
    try {
      await useAuthStore.getState().refreshToken();
      return useAuthStore.getState().isAuthenticated;
    } catch (error) {
      console.error('Token refresh during initialization failed:', error);
      return false;
    }
  }

  try {
    // Set loading state
    useAuthStore.setState({ isLoading: true });

    // Verify token with backend and get user info
    const response = await api.auth.me();

    if (response.data) {
      const tokens = createAuthTokens(token);

      useAuthStore.setState({
        user: response.data as User,
        tokens,
        isAuthenticated: true,
        isLoading: false,
        error: null
      });
      return true;
    } else {
      throw new Error('No user data in response');
    }
  } catch (error) {
    console.error('Auth initialization failed:', error);
    
    // Clear token from sessionStorage and reset auth state
    try {
      sessionStorage.removeItem('kubechat_auth_token');
      sessionStorage.removeItem('kubechat_refresh_token');
    } catch (clearError) {
      console.warn('Failed to clear tokens:', clearError);
    }
    
    useAuthStore.setState({
      user: null,
      tokens: null,
      isAuthenticated: false,
      isLoading: false,
      error: null
    });
    
    return false;
  }
};

// Token refresh interval
let refreshInterval: NodeJS.Timeout | null = null;

export const startTokenRefresh = () => {
  if (refreshInterval) return;

  refreshInterval = setInterval(() => {
    const { tokens, isAuthenticated } = useAuthStore.getState();

    if (!isAuthenticated || !tokens) return;

    // Refresh token 5 minutes before expiry
    const fiveMinutes = 5 * 60 * 1000;
    if (tokens.expiresAt - Date.now() < fiveMinutes) {
      useAuthStore.getState().refreshToken();
    }
  }, 60000); // Check every minute
};

export const stopTokenRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
};

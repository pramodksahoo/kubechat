// Authentication store using Zustand
// Following coding standards from docs/architecture/coding-standards.md

import { create } from 'zustand';
import { api } from '../services/api';
import { tokenService, createAuthTokens } from '../services/tokenService';
import type { AuthState, LoginCredentials, RegisterData, User, AuthTokens } from '../types/auth';

export const useAuthStore = create<AuthState>()((set, get) => ({
      // Initial state
      user: null,
      tokens: null,
      isAuthenticated: false,
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

          const { user } = response.data;
          const typedUser = user as User;

          // Get token from secure cookie (backend sets it automatically)
          const token = await tokenService.getAccessToken();
          if (!token) {
            throw new Error('Token not found after login');
          }

          // Create tokens object
          const tokens = createAuthTokens(token);

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
        // Clear secure tokens
        await tokenService.clearTokens();

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
        const refreshToken = await tokenService.getRefreshToken();

        if (!currentTokens || !refreshToken) {
          await get().logout();
          return;
        }

        try {
          set({ isLoading: true });

          const response = await api.auth.refresh();

          if (!response.data) {
            throw new Error('Token refresh failed');
          }

          // Get token from secure cookie (backend sets it automatically)
          const token = await tokenService.getAccessToken();
          if (!token) {
            throw new Error('Token not found after refresh');
          }

          await tokenService.setTokens(token);

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
    }));

// Initialize auth store with existing token
export const initializeAuth = async () => {
  const token = await tokenService.getAccessToken();

  if (!token) {
    return;
  }

  if (tokenService.isTokenExpired(token)) {
    await useAuthStore.getState().refreshToken();
    return;
  }

  try {
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
    }
  } catch (error) {
    console.error('Auth initialization failed:', error);
    await tokenService.clearTokens();
    await useAuthStore.getState().logout();
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

// Secure token storage service using httpOnly cookies
// Following security best practices for JWT token management

import { AuthTokens } from '../types/auth';

// Cookie names
const TOKEN_COOKIE_NAME = 'kubechat_token';
const REFRESH_TOKEN_COOKIE_NAME = 'kubechat_refresh_token';

// Cookie configuration
const COOKIE_OPTIONS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === 'production',
  sameSite: 'strict' as const,
  path: '/',
  maxAge: 24 * 60 * 60 * 1000, // 24 hours
};

export interface TokenService {
  setTokens(accessToken: string, refreshToken?: string): Promise<void>;
  getAccessToken(): Promise<string | null>;
  getRefreshToken(): Promise<string | null>;
  clearTokens(): Promise<void>;
  isTokenExpired(token: string): boolean;
  decodeToken(token: string): any;
}

// Secure cookie-based token storage implementation
class SecureTokenService implements TokenService {

  /**
   * Store tokens in secure httpOnly cookies via API call
   */
  async setTokens(accessToken: string, refreshToken?: string): Promise<void> {
    try {
      const response = await fetch('/api/auth/set-tokens', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          accessToken,
          refreshToken,
        }),
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to set secure tokens');
      }
    } catch (error) {
      console.error('Error setting secure tokens:', error);
      throw error;
    }
  }

  /**
   * Get access token from secure httpOnly cookie via API call
   */
  async getAccessToken(): Promise<string | null> {
    try {
      const response = await fetch('/api/auth/get-token', {
        method: 'GET',
        credentials: 'include',
      });

      if (!response.ok) {
        return null;
      }

      const data = await response.json();
      return data.accessToken || null;
    } catch (error) {
      console.error('Error getting access token:', error);
      return null;
    }
  }

  /**
   * Get refresh token from secure httpOnly cookie via API call
   */
  async getRefreshToken(): Promise<string | null> {
    try {
      const response = await fetch('/api/auth/get-refresh-token', {
        method: 'GET',
        credentials: 'include',
      });

      if (!response.ok) {
        return null;
      }

      const data = await response.json();
      return data.refreshToken || null;
    } catch (error) {
      console.error('Error getting refresh token:', error);
      return null;
    }
  }

  /**
   * Clear all tokens by calling API to remove httpOnly cookies
   */
  async clearTokens(): Promise<void> {
    try {
      const response = await fetch('/api/auth/clear-tokens', {
        method: 'POST',
        credentials: 'include',
      });

      if (!response.ok) {
        console.warn('Failed to clear secure tokens');
      }
    } catch (error) {
      console.error('Error clearing secure tokens:', error);
    }
  }

  /**
   * Decode JWT token payload
   */
  decodeToken(token: string): any {
    try {
      const payload = token.split('.')[1];
      return JSON.parse(atob(payload));
    } catch (error) {
      console.error('Error decoding token:', error);
      return null;
    }
  }

  /**
   * Check if token is expired
   */
  isTokenExpired(token: string): boolean {
    const decoded = this.decodeToken(token);
    if (!decoded || !decoded.exp) return true;

    // Add 30 second buffer to account for clock skew
    const bufferTime = 30 * 1000;
    return (decoded.exp * 1000) < (Date.now() + bufferTime);
  }
}

// Fallback localStorage implementation for development/testing
class LocalStorageTokenService implements TokenService {
  private readonly TOKEN_KEY = 'kubechat_auth_token';
  private readonly REFRESH_TOKEN_KEY = 'kubechat_refresh_token';

  async setTokens(accessToken: string, refreshToken?: string): Promise<void> {
    if (typeof window === 'undefined') return;

    localStorage.setItem(this.TOKEN_KEY, accessToken);
    if (refreshToken) {
      localStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken);
    }
  }

  async getAccessToken(): Promise<string | null> {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(this.TOKEN_KEY);
  }

  async getRefreshToken(): Promise<string | null> {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  async clearTokens(): Promise<void> {
    if (typeof window === 'undefined') return;

    localStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_TOKEN_KEY);
  }

  decodeToken(token: string): any {
    try {
      const payload = token.split('.')[1];
      return JSON.parse(atob(payload));
    } catch (error) {
      return null;
    }
  }

  isTokenExpired(token: string): boolean {
    const decoded = this.decodeToken(token);
    if (!decoded || !decoded.exp) return true;
    return decoded.exp * 1000 < Date.now();
  }
}

// Export service instance - use secure cookies in production, localStorage for development
export const tokenService: TokenService =
  process.env.NODE_ENV === 'production'
    ? new SecureTokenService()
    : new LocalStorageTokenService();

// Helper function to create AuthTokens object from token
export const createAuthTokens = (accessToken: string): AuthTokens => {
  const decoded = tokenService.decodeToken(accessToken);
  return {
    accessToken,
    expiresAt: decoded ? decoded.exp * 1000 : Date.now() + 24 * 60 * 60 * 1000,
  };
};
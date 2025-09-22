// Production-ready token service
// Uses sessionStorage for reliable token management in production mode

import { AuthTokens } from '../types/auth';

export interface TokenService {
  setTokens(accessToken: string, refreshToken?: string): Promise<void>;
  getAccessToken(): Promise<string | null>;
  getRefreshToken(): Promise<string | null>;
  clearTokens(): Promise<void>;
  isTokenExpired(token: string): boolean;
  decodeToken(token: string): any;
}

// Production token service using sessionStorage for reliability
class ProductionTokenService implements TokenService {
  private readonly TOKEN_KEY = 'kubechat_auth_token';
  private readonly REFRESH_TOKEN_KEY = 'kubechat_refresh_token';

  /**
   * Store tokens in sessionStorage (secure for production when using HTTPS)
   */
  async setTokens(accessToken: string, refreshToken?: string): Promise<void> {
    if (typeof window === 'undefined') return;

    try {
      sessionStorage.setItem(this.TOKEN_KEY, accessToken);
      if (refreshToken) {
        sessionStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken);
      }
    } catch (error) {
      console.error('Failed to store tokens:', error);
      throw new Error('Token storage failed');
    }
  }

  /**
   * Get access token from sessionStorage
   */
  async getAccessToken(): Promise<string | null> {
    if (typeof window === 'undefined') return null;
    
    try {
      return sessionStorage.getItem(this.TOKEN_KEY);
    } catch (error) {
      console.error('Error getting access token:', error);
      return null;
    }
  }

  /**
   * Get refresh token from sessionStorage
   */
  async getRefreshToken(): Promise<string | null> {
    if (typeof window === 'undefined') return null;
    
    try {
      return sessionStorage.getItem(this.REFRESH_TOKEN_KEY);
    } catch (error) {
      console.error('Error getting refresh token:', error);
      return null;
    }
  }

  /**
   * Clear all tokens from sessionStorage
   */
  async clearTokens(): Promise<void> {
    if (typeof window === 'undefined') return;
    
    try {
      sessionStorage.removeItem(this.TOKEN_KEY);
      sessionStorage.removeItem(this.REFRESH_TOKEN_KEY);
    } catch (error) {
      console.error('Error clearing tokens:', error);
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

// Export service instance
export const tokenService: TokenService = new ProductionTokenService();

// Helper function to create AuthTokens object from token
export const createAuthTokens = (accessToken: string): AuthTokens => {
  const decoded = tokenService.decodeToken(accessToken);
  return {
    accessToken,
    expiresAt: decoded ? decoded.exp * 1000 : Date.now() + 24 * 60 * 60 * 1000,
  };
};

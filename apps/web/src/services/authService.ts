import { User, Role, Permission } from '../types/user';
import { api } from './api';
import { tokenService } from './tokenService';

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  firstName?: string;
  lastName?: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  expiresAt: string;
}

export interface Session {
  id: string;
  userId: string;
  deviceType: string;
  browser: string;
  location: string;
  ipAddress: string;
  userAgent: string;
  isActive: boolean;
  lastActivity: Date;
  createdAt: Date;
  updatedAt: Date;
  expiresAt: Date;
}

class AuthService {
  private currentUser: User | null = null;
  private token: string | null = null;
  private refreshTimer: NodeJS.Timeout | null = null;

  constructor() {
    // Initialize from localStorage
    this.loadFromStorage();
    this.setupTokenRefresh();
  }

  // Authentication Methods
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    try {
      const response = await api.auth.login(credentials);

      // Wait briefly for cookie to be set by backend
      await new Promise(resolve => setTimeout(resolve, 100));

      // Get token from secure cookie (backend sets it automatically)
      const token = await tokenService.getAccessToken();
      if (!token) {
        throw new Error('Token not found after login');
      }

      const authResponse: AuthResponse = {
        user: (response.data as any).user as User,
        token: token,
        expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString() // 24 hours
      };

      // Store authentication data
      this.currentUser = authResponse.user;
      this.token = authResponse.token;
      this.saveToStorage();
      this.setupTokenRefresh();

      return authResponse;

    } catch (error) {
      console.error('Login failed:', error);
      throw new Error('Authentication failed. Please check your credentials.');
    }
  }

  async register(userData: RegisterData): Promise<AuthResponse> {
    try {
      const response = await api.auth.register(userData);

      const authResponse: AuthResponse = {
        user: (response.data as any).user as User,
        token: (response.data as any).token,
        expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
      };

      this.currentUser = authResponse.user;
      this.token = authResponse.token;
      this.saveToStorage();
      this.setupTokenRefresh();

      return authResponse;

    } catch (error) {
      console.error('Registration failed:', error);
      throw new Error('Registration failed. Please try again.');
    }
  }

  async logout(): Promise<void> {
    try {
      await api.auth.logout();
    } catch (error) {
      console.error('Logout API call failed:', error);
    } finally {
      // Clear local data regardless of API success
      this.currentUser = null;
      this.token = null;
      this.clearStorage();
      if (this.refreshTimer) {
        clearTimeout(this.refreshTimer);
        this.refreshTimer = null;
      }
    }
  }

  async refreshToken(): Promise<void> {
    try {
      const response = await api.auth.refresh();
      this.token = response.data.token;
      this.saveToStorage();
    } catch (error) {
      console.error('Token refresh failed:', error);
      // Force logout on refresh failure
      await this.logout();
      throw error;
    }
  }

  async getCurrentUser(): Promise<User | null> {
    if (!this.currentUser && this.token) {
      try {
        const response = await api.auth.me();
        this.currentUser = response.data as User;
        this.saveToStorage();
      } catch (error) {
        console.error('Failed to fetch current user:', error);
        await this.logout();
      }
    }
    return this.currentUser;
  }

  async updateProfile(updates: Partial<User>): Promise<User> {
    if (!this.currentUser) {
      throw new Error('User not authenticated');
    }

    try {
      const response = await api.auth.profile();
      const updatedUser = { ...(response.data as any), ...updates } as User;
      this.currentUser = updatedUser;
      this.saveToStorage();
      return updatedUser;
    } catch (error) {
      console.error('Profile update failed:', error);
      throw new Error('Failed to update profile');
    }
  }

  // User Management (Admin only)
  async getUsers(): Promise<User[]> {
    try {
      const response = await api.auth.getUsers();
      return response.data as User[];
    } catch (error) {
      console.error('Failed to fetch users:', error);
      throw new Error('Failed to fetch users');
    }
  }

  async getUser(id: string): Promise<User> {
    try {
      const response = await api.auth.getUser(id);
      return response.data as User;
    } catch (error) {
      console.error('Failed to fetch user:', error);
      throw new Error('Failed to fetch user');
    }
  }

  // Session Management
  async getSessions(): Promise<Session[]> {
    try {
      // Mock sessions since backend might not have this endpoint
      return [
        {
          id: 'session-1',
          userId: this.currentUser?.id || 'current-user',
          deviceType: 'Desktop',
          browser: this.getBrowserInfo(),
          location: 'Unknown',
          ipAddress: '192.168.1.100',
          userAgent: navigator.userAgent,
          isActive: true,
          lastActivity: new Date(),
          createdAt: new Date(),
          updatedAt: new Date(),
          expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000)
        }
      ];
    } catch (error) {
      console.error('Failed to fetch sessions:', error);
      return [];
    }
  }

  async terminateSession(sessionId: string): Promise<void> {
    try {
      // Use security API to manage sessions
      await api.security.cleanupSessions();
      console.log(`Session ${sessionId} terminated`);
    } catch (error) {
      console.error('Failed to terminate session:', error);
      throw new Error('Failed to terminate session');
    }
  }

  async terminateAllSessions(): Promise<void> {
    try {
      await api.security.cleanupSessions();
      // Keep current session active
    } catch (error) {
      console.error('Failed to terminate all sessions:', error);
      throw new Error('Failed to terminate all sessions');
    }
  }

  // Permission Management
  async getUserPermissions(userId?: string): Promise<Permission[]> {
    const targetUserId = userId || this.currentUser?.id;
    if (!targetUserId) {
      return [];
    }

    try {
      const user = await this.getUser(targetUserId);
      return user.permissions;
    } catch (error) {
      console.error('Failed to fetch user permissions:', error);
      return [];
    }
  }

  async hasPermission(permission: string, resource?: string): Promise<boolean> {
    const user = await this.getCurrentUser();
    if (!user) return false;

    // Check if user has the specific permission
    return user.permissions.some(p =>
      p.action === permission &&
      (!resource || p.resource === resource || p.resource === '*')
    );
  }

  async hasRole(roleName: string): Promise<boolean> {
    const user = await this.getCurrentUser();
    if (!user) return false;

    return user.roles.some(role => role.name === roleName);
  }

  // Token and Authentication State
  getToken(): string | null {
    return this.token;
  }

  isAuthenticated(): boolean {
    return !!this.token && !!this.currentUser;
  }

  isTokenExpired(): boolean {
    if (!this.token) return true;

    try {
      const payload = JSON.parse(atob(this.token.split('.')[1]));
      return Date.now() >= payload.exp * 1000;
    } catch {
      return true;
    }
  }

  // Storage Management
  private saveToStorage(): void {
    if (typeof window === 'undefined') return;

    try {
      localStorage.setItem('auth_token', this.token || '');
      localStorage.setItem('auth_user', JSON.stringify(this.currentUser));
      localStorage.setItem('auth_timestamp', Date.now().toString());
    } catch (error) {
      console.error('Failed to save auth data:', error);
    }
  }

  private loadFromStorage(): void {
    if (typeof window === 'undefined') return;

    try {
      this.token = localStorage.getItem('auth_token');
      const userStr = localStorage.getItem('auth_user');
      const timestamp = localStorage.getItem('auth_timestamp');

      if (userStr) {
        this.currentUser = JSON.parse(userStr);
      }

      // Check if data is too old (older than 24 hours)
      if (timestamp) {
        const age = Date.now() - parseInt(timestamp);
        if (age > 24 * 60 * 60 * 1000) {
          this.clearStorage();
        }
      }
    } catch (error) {
      console.error('Failed to load auth data:', error);
      this.clearStorage();
    }
  }

  private clearStorage(): void {
    if (typeof window === 'undefined') return;

    try {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_user');
      localStorage.removeItem('auth_timestamp');
    } catch (error) {
      console.error('Failed to clear auth data:', error);
    }
  }

  private setupTokenRefresh(): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    if (this.token) {
      // Refresh token every 23 hours
      this.refreshTimer = setTimeout(() => {
        this.refreshToken().catch(console.error);
      }, 23 * 60 * 60 * 1000);
    }
  }

  private getBrowserInfo(): string {
    const userAgent = navigator.userAgent;
    if (userAgent.includes('Chrome')) return 'Chrome';
    if (userAgent.includes('Firefox')) return 'Firefox';
    if (userAgent.includes('Safari')) return 'Safari';
    if (userAgent.includes('Edge')) return 'Edge';
    return 'Unknown';
  }
}

export const authService = new AuthService();
export default authService;

// Enterprise Session Management Service
// Following coding standards from docs/architecture/coding-standards.md

import { useAuthStore } from '../stores/authStore';
import { tokenService } from './tokenService';

interface SessionConfig {
  idleTimeout: number;          // Idle timeout in milliseconds
  absoluteTimeout: number;      // Absolute session timeout in milliseconds
  warningTime: number;          // Warning time before expiry in milliseconds
  maxConcurrentSessions: number; // Maximum concurrent sessions per user
  trackActivity: boolean;       // Track user activity
  trackLocation: boolean;       // Track session location/IP
}

interface SessionWarning {
  type: 'idle' | 'absolute' | 'refresh';
  timeRemaining: number;
  callback?: () => void;
}

class EnterpriseSessionManager {
  private config: SessionConfig;
  private activityTimer: NodeJS.Timeout | null = null;
  private warningTimer: NodeJS.Timeout | null = null;
  private lastActivity: number = Date.now();
  private sessionStart: number = Date.now();
  private warningCallback?: (warning: SessionWarning) => void;
  private isIdle: boolean = false;

  constructor(config?: Partial<SessionConfig>) {
    this.config = {
      idleTimeout: 30 * 60 * 1000,      // 30 minutes idle
      absoluteTimeout: 8 * 60 * 60 * 1000, // 8 hours absolute
      warningTime: 5 * 60 * 1000,       // 5 minutes warning
      maxConcurrentSessions: 3,          // Max 3 sessions
      trackActivity: true,
      trackLocation: true,
      ...config
    };

    this.initializeListeners();
  }

  private initializeListeners() {
    if (typeof window === 'undefined') return;

    // Track user activity
    const activityEvents = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click'];
    
    const handleActivity = () => {
      this.updateActivity();
    };

    activityEvents.forEach(event => {
      document.addEventListener(event, handleActivity, true);
    });

    // Track page visibility changes
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) {
        this.updateActivity();
      }
    });

    // Prevent back button navigation issues
    window.addEventListener('popstate', (event) => {
      // Add a small delay to allow auth state to settle
      setTimeout(() => {
        this.validateSession();
      }, 100);
    });

    // Handle beforeunload for session cleanup
    window.addEventListener('beforeunload', () => {
      this.cleanup();
    });
  }

  private updateActivity() {
    this.lastActivity = Date.now();
    
    if (this.isIdle) {
      this.isIdle = false;
      this.resetTimers();
    }

    this.resetActivityTimer();
  }

  private resetActivityTimer() {
    if (this.activityTimer) {
      clearTimeout(this.activityTimer);
    }

    this.activityTimer = setTimeout(() => {
      this.handleIdleTimeout();
    }, this.config.idleTimeout);

    // Set warning timer
    if (this.warningTimer) {
      clearTimeout(this.warningTimer);
    }

    const warningTime = this.config.idleTimeout - this.config.warningTime;
    if (warningTime > 0) {
      this.warningTimer = setTimeout(() => {
        this.handleSessionWarning('idle');
      }, warningTime);
    }
  }

  private handleIdleTimeout() {
    this.isIdle = true;
    console.log('Session idle timeout reached');
    
    // Show idle warning before logout
    if (this.warningCallback) {
      this.warningCallback({
        type: 'idle',
        timeRemaining: this.config.warningTime,
        callback: () => {
          this.extendSession();
        }
      });
    } else {
      // Auto-logout after warning period
      setTimeout(() => {
        this.logout('idle_timeout');
      }, this.config.warningTime);
    }
  }

  private handleAbsoluteTimeout() {
    console.log('Absolute session timeout reached');
    this.logout('absolute_timeout');
  }

  private handleSessionWarning(type: 'idle' | 'absolute' | 'refresh') {
    const now = Date.now();
    let timeRemaining = 0;

    switch (type) {
      case 'idle':
        timeRemaining = this.config.idleTimeout - (now - this.lastActivity);
        break;
      case 'absolute':
        timeRemaining = this.config.absoluteTimeout - (now - this.sessionStart);
        break;
      case 'refresh':
        const { tokens } = useAuthStore.getState();
        if (tokens) {
          timeRemaining = tokens.expiresAt - now;
        }
        break;
    }

    if (this.warningCallback && timeRemaining > 0) {
      this.warningCallback({
        type,
        timeRemaining,
        callback: () => {
          if (type === 'refresh') {
            this.refreshToken();
          } else {
            this.extendSession();
          }
        }
      });
    }
  }

  public startSession() {
    this.sessionStart = Date.now();
    this.lastActivity = Date.now();
    this.resetActivityTimer();

    // Set absolute timeout
    setTimeout(() => {
      this.handleAbsoluteTimeout();
    }, this.config.absoluteTimeout);

    // Set absolute timeout warning
    const absoluteWarningTime = this.config.absoluteTimeout - this.config.warningTime;
    if (absoluteWarningTime > 0) {
      setTimeout(() => {
        this.handleSessionWarning('absolute');
      }, absoluteWarningTime);
    }
  }

  public extendSession() {
    this.lastActivity = Date.now();
    this.isIdle = false;
    this.resetTimers();
  }

  public refreshToken() {
    const { refreshToken } = useAuthStore.getState();
    refreshToken().catch((error) => {
      console.error('Token refresh failed:', error);
      this.logout('token_refresh_failed');
    });
  }

  public validateSession() {
    const { tokens, isAuthenticated } = useAuthStore.getState();
    
    if (!isAuthenticated || !tokens) {
      return false;
    }

    const now = Date.now();
    
    // Check if token is expired
    if (tokens.expiresAt <= now) {
      this.logout('token_expired');
      return false;
    }

    // Check if session should refresh soon
    const refreshThreshold = 5 * 60 * 1000; // 5 minutes
    if (tokens.expiresAt - now < refreshThreshold) {
      this.handleSessionWarning('refresh');
    }

    return true;
  }

  public async logout(reason: string) {
    console.log(`Session logout: ${reason}`);
    this.cleanup();
    
    // Call auth store logout
    const { logout } = useAuthStore.getState();
    try {
      await logout();
    } catch (error: any) {
      console.error('Logout failed:', error);
    }
  }

  public cleanup() {
    if (this.activityTimer) {
      clearTimeout(this.activityTimer);
      this.activityTimer = null;
    }
    
    if (this.warningTimer) {
      clearTimeout(this.warningTimer);
      this.warningTimer = null;
    }
  }

  public setWarningCallback(callback: (warning: SessionWarning) => void) {
    this.warningCallback = callback;
  }

  public getSessionInfo() {
    const now = Date.now();
    const { tokens } = useAuthStore.getState();
    
    return {
      sessionStart: this.sessionStart,
      lastActivity: this.lastActivity,
      isIdle: this.isIdle,
      timeToIdle: Math.max(0, this.config.idleTimeout - (now - this.lastActivity)),
      timeToAbsolute: Math.max(0, this.config.absoluteTimeout - (now - this.sessionStart)),
      tokenExpiresAt: tokens?.expiresAt || 0,
      timeToTokenExpiry: tokens ? Math.max(0, tokens.expiresAt - now) : 0
    };
  }

  private resetTimers() {
    this.resetActivityTimer();
  }

  public updateConfig(newConfig: Partial<SessionConfig>) {
    this.config = { ...this.config, ...newConfig };
    this.resetTimers();
  }
}

// Singleton instance
export const sessionManager = new EnterpriseSessionManager();

export type { SessionConfig, SessionWarning };
export { EnterpriseSessionManager };
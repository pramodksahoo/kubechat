// Error Reporting and Monitoring System for KubeChat Enterprise
// Comprehensive monitoring, error tracking, and performance analytics

import React from 'react';
import { ScreenReaderUtils } from '../design-system/accessibility';

// Simple ID generator to replace uuid
function generateId(): string {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}

// Error severity levels
export enum ErrorSeverity {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  CRITICAL = 'critical',
}

// Error categories
export enum ErrorCategory {
  AUTHENTICATION = 'authentication',
  AUTHORIZATION = 'authorization',
  NETWORK = 'network',
  VALIDATION = 'validation',
  BUSINESS_LOGIC = 'business_logic',
  EXTERNAL_SERVICE = 'external_service',
  KUBERNETES = 'kubernetes',
  SYSTEM = 'system',
  USER_INPUT = 'user_input',
  PERFORMANCE = 'performance',
}

// Performance metrics
export interface PerformanceMetric {
  name: string;
  value: number;
  unit: string;
  timestamp: Date;
  tags?: Record<string, string>;
}

// Error context
export interface ErrorContext {
  userId?: string;
  sessionId?: string;
  userAgent?: string;
  url?: string;
  component?: string;
  action?: string;
  clusterId?: string;
  namespace?: string;
  workload?: string;
  command?: string;
  timestamp: Date;
  additionalData?: Record<string, any>;
}

// Error report
export interface ErrorReport {
  id: string;
  message: string;
  severity: ErrorSeverity;
  category: ErrorCategory;
  stack?: string;
  context: ErrorContext;
  fingerprint: string;
  occurred: Date;
  resolved: boolean;
  resolvedAt?: Date;
  occurrenceCount: number;
  lastOccurrence: Date;
  tags: Record<string, string>;
}

// User feedback on errors
export interface ErrorFeedback {
  errorId: string;
  userId: string;
  helpful: boolean;
  expectedBehavior?: string;
  actualBehavior?: string;
  steps?: string;
  severity?: ErrorSeverity;
  timestamp: Date;
}

// Performance metrics tracking
export interface PerformanceReport {
  id: string;
  metrics: PerformanceMetric[];
  context: ErrorContext;
  timestamp: Date;
}

// Main monitoring service
export class MonitoringService {
  private static instance: MonitoringService;
  private sessionId: string;
  private userId?: string;
  private errorQueue: ErrorReport[] = [];
  private performanceQueue: PerformanceReport[] = [];
  private feedbackQueue: ErrorFeedback[] = [];
  private isOnline = navigator.onLine;
  private flushInterval?: NodeJS.Timeout;

  private constructor() {
    this.sessionId = generateId();
    this.setupEventListeners();
    this.startPerformanceMonitoring();
    this.scheduleFlush();
  }

  static getInstance(): MonitoringService {
    if (!MonitoringService.instance) {
      MonitoringService.instance = new MonitoringService();
    }
    return MonitoringService.instance;
  }

  // Initialize with user context
  initialize(userId: string, additionalContext?: Record<string, any>) {
    this.userId = userId;
    
    // Send initialization event
    this.trackEvent('session_started', {
      userId,
      sessionId: this.sessionId,
      ...additionalContext,
    });
  }

  // Generate error fingerprint for deduplication
  private generateFingerprint(error: Error, context: ErrorContext): string {
    const key = `${error.name}:${error.message}:${context.component}:${context.action}`;
    return btoa(key).substring(0, 16);
  }

  // Report an error
  reportError(
    error: Error,
    severity: ErrorSeverity = ErrorSeverity.MEDIUM,
    category: ErrorCategory = ErrorCategory.SYSTEM,
    context: Partial<ErrorContext> = {}
  ): string {
    const errorId = generateId();
    const fingerprint = this.generateFingerprint(error, context as ErrorContext);
    
    // Check if this error already exists in queue (deduplication)
    const existingError = this.errorQueue.find(e => e.fingerprint === fingerprint);
    
    if (existingError) {
      existingError.occurrenceCount++;
      existingError.lastOccurrence = new Date();
      return existingError.id;
    }

    const errorReport: ErrorReport = {
      id: errorId,
      message: error.message,
      severity,
      category,
      stack: error.stack,
      context: {
        userId: this.userId,
        sessionId: this.sessionId,
        userAgent: navigator.userAgent,
        url: window.location.href,
        timestamp: new Date(),
        ...context,
      },
      fingerprint,
      occurred: new Date(),
      resolved: false,
      occurrenceCount: 1,
      lastOccurrence: new Date(),
      tags: {
        environment: process.env.NODE_ENV || 'development',
        version: process.env.REACT_APP_VERSION || 'unknown',
        ...this.extractTags(context),
      },
    };

    this.errorQueue.push(errorReport);
    
    // Immediate flush for critical errors
    if (severity === ErrorSeverity.CRITICAL) {
      this.flushErrors();
      
      // Notify user of critical error
      ScreenReaderUtils.announce(
        'A critical error occurred and has been reported. Support has been notified.',
        'assertive'
      );
    }

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.group(`🚨 Error Reported [${severity}]`);
      console.error('Error:', error);
      console.log('Context:', errorReport.context);
      console.log('Error ID:', errorId);
      console.groupEnd();
    }

    return errorId;
  }

  // Report performance metrics
  reportPerformance(metrics: PerformanceMetric[], context: Partial<ErrorContext> = {}) {
    const report: PerformanceReport = {
      id: generateId(),
      metrics,
      context: {
        userId: this.userId,
        sessionId: this.sessionId,
        userAgent: navigator.userAgent,
        url: window.location.href,
        timestamp: new Date(),
        ...context,
      },
      timestamp: new Date(),
    };

    this.performanceQueue.push(report);
  }

  // Submit user feedback on error
  submitErrorFeedback(
    errorId: string,
    helpful: boolean,
    details?: {
      expectedBehavior?: string;
      actualBehavior?: string;
      steps?: string;
      severity?: ErrorSeverity;
    }
  ) {
    if (!this.userId) return;

    const feedback: ErrorFeedback = {
      errorId,
      userId: this.userId,
      helpful,
      ...details,
      timestamp: new Date(),
    };

    this.feedbackQueue.push(feedback);
    
    // Immediate flush for feedback
    this.flushFeedback();
  }

  // Track custom events
  trackEvent(event: string, data: Record<string, any> = {}) {
    const eventData = {
      event,
      userId: this.userId,
      sessionId: this.sessionId,
      timestamp: new Date().toISOString(),
      url: window.location.href,
      ...data,
    };

    // Store in queue for batching
    this.performanceQueue.push({
      id: generateId(),
      metrics: [{
        name: 'custom_event',
        value: 1,
        unit: 'count',
        timestamp: new Date(),
        tags: { event, ...data },
      }],
      context: {
        userId: this.userId,
        sessionId: this.sessionId,
        timestamp: new Date(),
      },
      timestamp: new Date(),
    });

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.log('📊 Event Tracked:', eventData);
    }
  }

  // Track page views
  trackPageView(page: string, additionalData?: Record<string, any>) {
    this.trackEvent('page_view', {
      page,
      referrer: document.referrer,
      ...additionalData,
    });
  }

  // Track user interactions
  trackInteraction(action: string, component: string, additionalData?: Record<string, any>) {
    this.trackEvent('user_interaction', {
      action,
      component,
      ...additionalData,
    });
  }

  // Track API calls
  trackApiCall(
    endpoint: string,
    method: string,
    status: number,
    duration: number,
    additionalData?: Record<string, any>
  ) {
    this.reportPerformance([
      {
        name: 'api_call_duration',
        value: duration,
        unit: 'ms',
        timestamp: new Date(),
        tags: {
          endpoint,
          method,
          status: status.toString(),
          ...additionalData,
        },
      },
    ]);

    // Report as error if API call failed
    if (status >= 400) {
      this.reportError(
        new Error(`API call failed: ${method} ${endpoint} - ${status}`),
        status >= 500 ? ErrorSeverity.HIGH : ErrorSeverity.MEDIUM,
        ErrorCategory.NETWORK,
        {
          component: 'api_client',
          action: 'api_call',
          additionalData: { endpoint, method, status, duration },
        }
      );
    }
  }

  // Extract tags from context
  private extractTags(context: Partial<ErrorContext>): Record<string, string> {
    const tags: Record<string, string> = {};
    
    if (context.component) tags.component = context.component;
    if (context.action) tags.action = context.action;
    if (context.clusterId) tags.cluster_id = context.clusterId;
    if (context.namespace) tags.namespace = context.namespace;
    if (context.workload) tags.workload = context.workload;
    
    return tags;
  }

  // Setup event listeners for automatic error capture
  private setupEventListeners() {
    // Global error handler
    window.addEventListener('error', (event) => {
      this.reportError(
        event.error || new Error(event.message),
        ErrorSeverity.HIGH,
        ErrorCategory.SYSTEM,
        {
          component: 'global_error_handler',
          action: 'uncaught_error',
          additionalData: {
            filename: event.filename,
            lineno: event.lineno,
            colno: event.colno,
          },
        }
      );
    });

    // Unhandled promise rejection
    window.addEventListener('unhandledrejection', (event) => {
      this.reportError(
        new Error(`Unhandled Promise Rejection: ${event.reason}`),
        ErrorSeverity.HIGH,
        ErrorCategory.SYSTEM,
        {
          component: 'global_error_handler',
          action: 'unhandled_rejection',
        }
      );
    });

    // Online/offline status
    window.addEventListener('online', () => {
      this.isOnline = true;
      this.flushAll();
      this.trackEvent('network_status_change', { status: 'online' });
    });

    window.addEventListener('offline', () => {
      this.isOnline = false;
      this.trackEvent('network_status_change', { status: 'offline' });
    });

    // Page visibility changes
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'hidden') {
        this.flushAll();
        this.trackEvent('page_visibility_change', { state: 'hidden' });
      } else {
        this.trackEvent('page_visibility_change', { state: 'visible' });
      }
    });

    // Before page unload
    window.addEventListener('beforeunload', () => {
      this.flushAll();
      this.trackEvent('session_ended');
    });
  }

  // Start performance monitoring
  private startPerformanceMonitoring() {
    // Monitor Core Web Vitals
    if ('web-vital' in window) {
      // This would integrate with web-vitals library
      // For now, we'll monitor basic performance metrics
    }

    // Monitor page load performance
    window.addEventListener('load', () => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      
      if (navigation) {
        this.reportPerformance([
          {
            name: 'page_load_time',
            value: navigation.loadEventEnd - navigation.loadEventStart,
            unit: 'ms',
            timestamp: new Date(),
          },
          {
            name: 'dom_content_loaded',
            value: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
            unit: 'ms',
            timestamp: new Date(),
          },
          {
            name: 'first_contentful_paint',
            value: navigation.domContentLoadedEventEnd - navigation.fetchStart,
            unit: 'ms',
            timestamp: new Date(),
          },
        ]);
      }
    });

    // Monitor long tasks
    if ('PerformanceObserver' in window) {
      try {
        const observer = new PerformanceObserver((list) => {
          list.getEntries().forEach((entry) => {
            if (entry.duration > 50) {
              this.reportPerformance([{
                name: 'long_task',
                value: entry.duration,
                unit: 'ms',
                timestamp: new Date(),
                tags: {
                  entry_type: entry.entryType,
                  name: entry.name,
                },
              }]);
            }
          });
        });
        
        observer.observe({ entryTypes: ['longtask'] });
      } catch (error) {
        // PerformanceObserver not supported
      }
    }
  }

  // Schedule periodic flushing
  private scheduleFlush() {
    this.flushInterval = setInterval(() => {
      if (this.isOnline) {
        this.flushAll();
      }
    }, 30000); // Flush every 30 seconds
  }

  // Flush all queues
  private async flushAll() {
    await Promise.all([
      this.flushErrors(),
      this.flushPerformance(),
      this.flushFeedback(),
    ]);
  }

  // Flush error reports
  private async flushErrors() {
    if (this.errorQueue.length === 0 || !this.isOnline) return;

    const errors = [...this.errorQueue];
    this.errorQueue = [];

    try {
      // In production, send to your error reporting service
      // await fetch('/api/errors', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ errors }),
      // });

      console.log('📤 Flushed errors:', errors.length);
    } catch (error) {
      // Restore errors to queue if sending failed
      this.errorQueue.unshift(...errors);
      console.error('Failed to flush errors:', error);
    }
  }

  // Flush performance reports
  private async flushPerformance() {
    if (this.performanceQueue.length === 0 || !this.isOnline) return;

    const reports = [...this.performanceQueue];
    this.performanceQueue = [];

    try {
      // In production, send to your analytics service
      // await fetch('/api/analytics', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ reports }),
      // });

      console.log('📤 Flushed performance reports:', reports.length);
    } catch (error) {
      // Restore reports to queue if sending failed
      this.performanceQueue.unshift(...reports);
      console.error('Failed to flush performance reports:', error);
    }
  }

  // Flush feedback
  private async flushFeedback() {
    if (this.feedbackQueue.length === 0 || !this.isOnline) return;

    const feedback = [...this.feedbackQueue];
    this.feedbackQueue = [];

    try {
      // In production, send to your feedback service
      // await fetch('/api/feedback', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ feedback }),
      // });

      console.log('📤 Flushed feedback:', feedback.length);
    } catch (error) {
      // Restore feedback to queue if sending failed
      this.feedbackQueue.unshift(...feedback);
      console.error('Failed to flush feedback:', error);
    }
  }

  // Get queue stats
  getStats() {
    return {
      errors: this.errorQueue.length,
      performance: this.performanceQueue.length,
      feedback: this.feedbackQueue.length,
      isOnline: this.isOnline,
      sessionId: this.sessionId,
      userId: this.userId,
    };
  }

  // Cleanup
  destroy() {
    if (this.flushInterval) {
      clearInterval(this.flushInterval);
    }
    this.flushAll();
  }
}

// Convenience functions
export const monitoring = MonitoringService.getInstance();

export const reportError = (
  error: Error,
  severity?: ErrorSeverity,
  category?: ErrorCategory,
  context?: Partial<ErrorContext>
) => monitoring.reportError(error, severity, category, context);

export const trackEvent = (event: string, data?: Record<string, any>) =>
  monitoring.trackEvent(event, data);

export const trackPageView = (page: string, data?: Record<string, any>) =>
  monitoring.trackPageView(page, data);

export const trackInteraction = (action: string, component: string, data?: Record<string, any>) =>
  monitoring.trackInteraction(action, component, data);

export const trackApiCall = (
  endpoint: string,
  method: string,
  status: number,
  duration: number,
  data?: Record<string, any>
) => monitoring.trackApiCall(endpoint, method, status, duration, data);

// React hook for monitoring
export function useMonitoring() {
  return {
    reportError,
    trackEvent,
    trackPageView,
    trackInteraction,
    trackApiCall,
    submitFeedback: monitoring.submitErrorFeedback.bind(monitoring),
    getStats: monitoring.getStats.bind(monitoring),
  };
}

// Higher-order component for automatic error reporting
export function withErrorReporting<P extends object>(
  Component: React.ComponentType<P>,
  componentName?: string
) {
  const WrappedComponent = (props: P) => {
    React.useEffect(() => {
      // Track component mount
      trackEvent('component_mounted', { component: componentName || Component.name });
      
      return () => {
        // Track component unmount
        trackEvent('component_unmounted', { component: componentName || Component.name });
      };
    }, []);

    return React.createElement(Component, props);
  };

  WrappedComponent.displayName = `withErrorReporting(${componentName || Component.displayName || Component.name})`;
  
  return WrappedComponent;
}
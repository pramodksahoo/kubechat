export interface ErrorInfo {
  id: string;
  type: 'network' | 'api' | 'validation' | 'auth' | 'system' | 'user';
  code?: string | number;
  message: string;
  description?: string;
  context?: Record<string, any>;
  timestamp: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  retryable: boolean;
  source: string;
  stackTrace?: string;
  userAction?: string;
}

export interface ErrorAction {
  id: string;
  label: string;
  type: 'retry' | 'reload' | 'navigate' | 'custom';
  handler?: () => void | Promise<void>;
  href?: string;
  primary?: boolean;
}

export interface ErrorRecoveryStrategy {
  type: string;
  maxRetries: number;
  retryDelay: number;
  exponentialBackoff: boolean;
  fallbackAction?: () => void;
}

class ErrorService {
  private errorListeners: ((error: ErrorInfo) => void)[] = [];
  private errorHistory: ErrorInfo[] = [];
  private retryAttempts: Map<string, number> = new Map();
  private recoveryStrategies: Map<string, ErrorRecoveryStrategy> = new Map();

  constructor() {
    // Set up global error handlers
    if (typeof window !== 'undefined') {
      window.addEventListener('error', this.handleGlobalError.bind(this));
      window.addEventListener('unhandledrejection', this.handleUnhandledRejection.bind(this));
    }

    // Default recovery strategies
    this.registerRecoveryStrategy('network', {
      type: 'network',
      maxRetries: 3,
      retryDelay: 1000,
      exponentialBackoff: true,
    });

    this.registerRecoveryStrategy('api', {
      type: 'api',
      maxRetries: 2,
      retryDelay: 2000,
      exponentialBackoff: false,
    });
  }

  // Register recovery strategy for error type
  registerRecoveryStrategy(type: string, strategy: ErrorRecoveryStrategy): void {
    this.recoveryStrategies.set(type, strategy);
  }

  // Handle and classify errors
  handleError(error: Error | string, context?: {
    source: string;
    type?: ErrorInfo['type'];
    severity?: ErrorInfo['severity'];
    retryable?: boolean;
    userAction?: string;
  }): ErrorInfo {
    const errorInfo = this.classifyError(error, context);

    // Store in history
    this.errorHistory.unshift(errorInfo);
    if (this.errorHistory.length > 100) {
      this.errorHistory = this.errorHistory.slice(0, 100);
    }

    // Notify listeners
    this.notifyListeners(errorInfo);

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.error('Error handled:', errorInfo);
    }

    // Send to monitoring service in production
    if (process.env.NODE_ENV === 'production') {
      this.sendToMonitoring(errorInfo);
    }

    return errorInfo;
  }

  // Classify error based on content and context
  private classifyError(error: Error | string, context?: any): ErrorInfo {
    const message = error instanceof Error ? error.message : error;
    const stackTrace = error instanceof Error ? error.stack : undefined;

    let type: ErrorInfo['type'] = 'system';
    let severity: ErrorInfo['severity'] = 'medium';
    let retryable = false;
    let userAction = 'Please try again later';

    // Classify based on error message patterns
    if (message.toLowerCase().includes('network') || message.toLowerCase().includes('fetch')) {
      type = 'network';
      severity = 'high';
      retryable = true;
      userAction = 'Check your internet connection and try again';
    } else if (message.toLowerCase().includes('unauthorized') || message.toLowerCase().includes('auth')) {
      type = 'auth';
      severity = 'high';
      retryable = false;
      userAction = 'Please log in again';
    } else if (message.toLowerCase().includes('validation') || message.toLowerCase().includes('invalid')) {
      type = 'validation';
      severity = 'low';
      retryable = false;
      userAction = 'Please check your input and try again';
    } else if (message.toLowerCase().includes('server') || message.toLowerCase().includes('api')) {
      type = 'api';
      severity = 'high';
      retryable = true;
      userAction = 'Service temporarily unavailable. Please try again in a few minutes';
    }

    // Override with context if provided
    if (context) {
      type = context.type || type;
      severity = context.severity || severity;
      retryable = context.retryable !== undefined ? context.retryable : retryable;
      userAction = context.userAction || userAction;
    }

    return {
      id: `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      type,
      message,
      severity,
      retryable,
      timestamp: new Date().toISOString(),
      source: context?.source || 'unknown',
      stackTrace,
      userAction,
      context: context || {}
    };
  }

  // Handle global JavaScript errors
  private handleGlobalError(event: ErrorEvent): void {
    this.handleError(event.error || new Error(event.message), {
      source: 'global-error-handler',
      type: 'system',
      severity: 'high'
    });
  }

  // Handle unhandled promise rejections
  private handleUnhandledRejection(event: PromiseRejectionEvent): void {
    this.handleError(event.reason || new Error('Unhandled promise rejection'), {
      source: 'promise-rejection',
      type: 'system',
      severity: 'high'
    });
  }

  // Retry an operation with exponential backoff
  async retryOperation<T>(
    operation: () => Promise<T>,
    errorType: string,
    operationId?: string
  ): Promise<T> {
    const strategy = this.recoveryStrategies.get(errorType);
    if (!strategy) {
      throw new Error(`No recovery strategy defined for error type: ${errorType}`);
    }

    const id = operationId || `operation-${Date.now()}`;
    const currentAttempts = this.retryAttempts.get(id) || 0;

    try {
      const result = await operation();
      // Success, reset retry count
      this.retryAttempts.delete(id);
      return result;
    } catch (error) {
      if (currentAttempts >= strategy.maxRetries) {
        // Max retries reached, run fallback if available
        if (strategy.fallbackAction) {
          strategy.fallbackAction();
        }

        // Create error info and throw
        const errorInfo = this.handleError(error as Error, {
          source: 'retry-operation',
          type: errorType as ErrorInfo['type'],
          retryable: false,
          userAction: 'Maximum retry attempts reached'
        });

        throw error;
      }

      // Calculate delay with exponential backoff if enabled
      let delay = strategy.retryDelay;
      if (strategy.exponentialBackoff) {
        delay = strategy.retryDelay * Math.pow(2, currentAttempts);
      }

      // Update retry count
      this.retryAttempts.set(id, currentAttempts + 1);

      // Log retry attempt
      console.log(`Retry attempt ${currentAttempts + 1}/${strategy.maxRetries} for ${errorType} after ${delay}ms`);

      // Wait before retry
      await new Promise(resolve => setTimeout(resolve, delay));

      // Recursive retry
      return this.retryOperation(operation, errorType, id);
    }
  }

  // Get error actions based on error type
  getErrorActions(error: ErrorInfo): ErrorAction[] {
    const actions: ErrorAction[] = [];

    switch (error.type) {
      case 'network':
        actions.push({
          id: 'retry-network',
          label: 'Retry',
          type: 'retry',
          primary: true
        });
        actions.push({
          id: 'reload-page',
          label: 'Reload Page',
          type: 'reload',
          handler: () => window.location.reload()
        });
        break;

      case 'auth':
        actions.push({
          id: 'login-again',
          label: 'Login Again',
          type: 'navigate',
          href: '/login',
          primary: true
        });
        break;

      case 'validation':
        actions.push({
          id: 'dismiss-error',
          label: 'Dismiss',
          type: 'custom',
          primary: true
        });
        break;

      case 'api':
        if (error.retryable) {
          actions.push({
            id: 'retry-api',
            label: 'Try Again',
            type: 'retry',
            primary: true
          });
        }
        actions.push({
          id: 'go-dashboard',
          label: 'Go to Dashboard',
          type: 'navigate',
          href: '/'
        });
        break;

      default:
        actions.push({
          id: 'reload-page',
          label: 'Reload Page',
          type: 'reload',
          handler: () => window.location.reload(),
          primary: true
        });
        break;
    }

    return actions;
  }

  // Get error history
  getErrorHistory(): ErrorInfo[] {
    return [...this.errorHistory];
  }

  // Clear error history
  clearErrorHistory(): void {
    this.errorHistory = [];
  }

  // Get errors by severity
  getErrorsBySeverity(severity: ErrorInfo['severity']): ErrorInfo[] {
    return this.errorHistory.filter(error => error.severity === severity);
  }

  // Get critical errors from last hour
  getCriticalErrors(hours: number = 1): ErrorInfo[] {
    const cutoff = new Date(Date.now() - hours * 60 * 60 * 1000).toISOString();
    return this.errorHistory.filter(
      error => error.severity === 'critical' && error.timestamp > cutoff
    );
  }

  // Subscribe to error events
  onError(callback: (error: ErrorInfo) => void): () => void {
    this.errorListeners.push(callback);

    // Return unsubscribe function
    return () => {
      const index = this.errorListeners.indexOf(callback);
      if (index > -1) {
        this.errorListeners.splice(index, 1);
      }
    };
  }

  // Notify all listeners
  private notifyListeners(error: ErrorInfo): void {
    this.errorListeners.forEach(listener => {
      try {
        listener(error);
      } catch (err) {
        console.error('Error in error listener:', err);
      }
    });
  }

  // Send error to monitoring service (mock)
  private sendToMonitoring(error: ErrorInfo): void {
    // In real implementation, send to services like Sentry, DataDog, etc.
    console.log('Sending error to monitoring service:', {
      id: error.id,
      type: error.type,
      message: error.message,
      severity: error.severity,
      source: error.source,
      timestamp: error.timestamp
    });
  }

  // Handle API errors specifically
  handleApiError(response: Response, context?: string): ErrorInfo {
    let message = `API Error: ${response.status} ${response.statusText}`;
    let userAction = 'Please try again later';

    switch (response.status) {
      case 400:
        message = 'Invalid request data';
        userAction = 'Please check your input and try again';
        break;
      case 401:
        message = 'Authentication required';
        userAction = 'Please log in again';
        break;
      case 403:
        message = 'Access denied';
        userAction = 'You do not have permission to perform this action';
        break;
      case 404:
        message = 'Resource not found';
        userAction = 'The requested resource could not be found';
        break;
      case 429:
        message = 'Too many requests';
        userAction = 'Please wait a moment before trying again';
        break;
      case 500:
        message = 'Internal server error';
        userAction = 'Service temporarily unavailable. Please try again later';
        break;
      case 503:
        message = 'Service unavailable';
        userAction = 'Service is temporarily down for maintenance';
        break;
    }

    return this.handleError(new Error(message), {
      source: context || 'api-call',
      type: 'api',
      severity: response.status >= 500 ? 'critical' : response.status >= 400 ? 'high' : 'medium',
      retryable: response.status >= 500 || response.status === 429,
      userAction
    });
  }

  // Create user-friendly error message
  getUserFriendlyMessage(error: ErrorInfo): string {
    const baseMessages = {
      network: 'Unable to connect to the server. Please check your internet connection.',
      api: 'Service is temporarily unavailable. Please try again in a few minutes.',
      auth: 'Your session has expired. Please log in again.',
      validation: 'Please check your input and try again.',
      system: 'An unexpected error occurred. Please try again.',
      user: 'Please try again or contact support if the problem persists.'
    };

    return error.userAction || baseMessages[error.type] || baseMessages.system;
  }
}

export const errorService = new ErrorService();
export default ErrorService;
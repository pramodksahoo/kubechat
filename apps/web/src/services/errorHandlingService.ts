import { ErrorType } from '@/components/chat/ErrorDisplay';

export interface ErrorContext {
  operation: string;
  component?: string;
  userId?: string;
  sessionId?: string;
  timestamp: number;
  retryCount: number;
}

export interface RetryConfig {
  maxAttempts: number;
  baseDelay: number;
  backoffMultiplier: number;
  maxDelay: number;
}

export interface ErrorHandlingOptions {
  showToUser?: boolean;
  logToConsole?: boolean;
  sendToAnalytics?: boolean;
  context?: Partial<ErrorContext>;
  retryConfig?: Partial<RetryConfig>;
}

export class ErrorHandlingService {
  private static instance: ErrorHandlingService;
  private errorHistory: Map<string, ErrorContext[]> = new Map();
  private retryQueue: Map<string, () => Promise<any>> = new Map();

  static getInstance(): ErrorHandlingService {
    if (!ErrorHandlingService.instance) {
      ErrorHandlingService.instance = new ErrorHandlingService();
    }
    return ErrorHandlingService.instance;
  }

  private getDefaultRetryConfig(): RetryConfig {
    return {
      maxAttempts: 3,
      baseDelay: 1000,
      backoffMultiplier: 2,
      maxDelay: 10000
    };
  }

  /**
   * Classify error type based on error details (AC: 5)
   */
  classifyError(error: Error | string, statusCode?: number): ErrorType {
    const errorMessage = error instanceof Error ? error.message : error;
    const lowerMessage = errorMessage.toLowerCase();

    // Network errors
    if (lowerMessage.includes('network') ||
        lowerMessage.includes('fetch') ||
        lowerMessage.includes('connection') ||
        statusCode === 0) {
      return 'network';
    }

    // Authentication errors
    if (statusCode === 401 ||
        lowerMessage.includes('unauthorized') ||
        lowerMessage.includes('authentication') ||
        lowerMessage.includes('token')) {
      return 'authentication';
    }

    // Permission errors
    if (statusCode === 403 ||
        lowerMessage.includes('forbidden') ||
        lowerMessage.includes('permission') ||
        lowerMessage.includes('access denied')) {
      return 'permission';
    }

    // Validation errors
    if (statusCode === 400 ||
        lowerMessage.includes('validation') ||
        lowerMessage.includes('invalid') ||
        lowerMessage.includes('bad request')) {
      return 'validation';
    }

    // Timeout errors
    if (lowerMessage.includes('timeout') ||
        lowerMessage.includes('timed out') ||
        statusCode === 408) {
      return 'timeout';
    }

    // API errors
    if (statusCode && statusCode >= 500) {
      return 'api';
    }

    return 'general';
  }

  /**
   * Handle errors with comprehensive context and recovery options (AC: 5)
   */
  async handleError(
    error: Error | string,
    options: ErrorHandlingOptions = {}
  ): Promise<{
    type: ErrorType;
    context: ErrorContext;
    canRetry: boolean;
    suggestions: string[];
  }> {
    const errorType = this.classifyError(error);
    const context: ErrorContext = {
      operation: options.context?.operation || 'unknown',
      component: options.context?.component,
      userId: options.context?.userId || 'anonymous',
      sessionId: options.context?.sessionId,
      timestamp: Date.now(),
      retryCount: options.context?.retryCount || 0,
      ...options.context
    };

    // Log error history
    const errorKey = `${context.operation}-${context.component || 'global'}`;
    if (!this.errorHistory.has(errorKey)) {
      this.errorHistory.set(errorKey, []);
    }
    this.errorHistory.get(errorKey)?.push(context);

    // Console logging
    if (options.logToConsole !== false) {
      console.error(`[ErrorHandling] ${errorType.toUpperCase()}:`, {
        error,
        context,
        stack: error instanceof Error ? error.stack : undefined
      });
    }

    // Determine if retry is possible
    const canRetry = this.canRetryError(errorType, context);

    // Generate suggestions
    const suggestions = this.generateSuggestions(errorType, error, context);

    return {
      type: errorType,
      context,
      canRetry,
      suggestions
    };
  }

  /**
   * Retry mechanism with exponential backoff (AC: 5)
   */
  async withRetry<T>(
    operation: () => Promise<T>,
    operationName: string,
    retryConfig?: Partial<RetryConfig>
  ): Promise<T> {
    const config = { ...this.getDefaultRetryConfig(), ...retryConfig };
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= config.maxAttempts; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));

        // Don't retry on certain error types
        const errorType = this.classifyError(lastError);
        if (!this.shouldRetryErrorType(errorType)) {
          throw lastError;
        }

        // Don't delay on last attempt
        if (attempt < config.maxAttempts) {
          const delay = Math.min(
            config.baseDelay * Math.pow(config.backoffMultiplier, attempt - 1),
            config.maxDelay
          );

          console.log(`[Retry] Attempt ${attempt}/${config.maxAttempts} failed for ${operationName}. Retrying in ${delay}ms...`);
          await this.delay(delay);
        }
      }
    }

    throw lastError;
  }

  /**
   * Generate user-friendly error suggestions (AC: 5)
   */
  private generateSuggestions(errorType: ErrorType, error: Error | string, context: ErrorContext): string[] {
    const suggestions: string[] = [];
    const errorMessage = error instanceof Error ? error.message : error;

    switch (errorType) {
      case 'network':
        suggestions.push('Check your internet connection');
        suggestions.push('Verify the server is accessible');
        if (context.retryCount < 3) {
          suggestions.push('Try the operation again');
        }
        break;

      case 'authentication':
        suggestions.push('Please log in to your account');
        suggestions.push('Check if your session has expired');
        suggestions.push('Verify your credentials are correct');
        break;

      case 'permission':
        suggestions.push('Contact your administrator for access');
        suggestions.push('Ensure you have the necessary permissions');
        suggestions.push('Try a different action that you have access to');
        break;

      case 'validation':
        suggestions.push('Check that all required fields are filled correctly');
        suggestions.push('Ensure your input follows the expected format');
        if (errorMessage.includes('kubectl')) {
          suggestions.push('Verify the kubectl command syntax');
          suggestions.push('Check if the resource names are correct');
        }
        break;

      case 'api':
        suggestions.push('Try the operation again in a few moments');
        if (context.retryCount === 0) {
          suggestions.push('Contact support if the problem persists');
        }
        suggestions.push('Check if the service is currently available');
        break;

      case 'timeout':
        suggestions.push('Try the operation again with a longer timeout');
        suggestions.push('Check your network connection speed');
        suggestions.push('Consider breaking large operations into smaller ones');
        break;

      default:
        suggestions.push('Try refreshing the page');
        suggestions.push('Clear your browser cache if the problem persists');
        break;
    }

    // Add context-specific suggestions
    if (context.operation === 'query-processing') {
      suggestions.push('Try rephrasing your query');
      suggestions.push('Use simpler language in your request');
    } else if (context.operation === 'command-execution') {
      suggestions.push('Check if the target resources exist');
      suggestions.push('Verify your cluster connection');
    }

    return suggestions;
  }

  /**
   * Check if error type should be retried
   */
  private shouldRetryErrorType(errorType: ErrorType): boolean {
    switch (errorType) {
      case 'network':
      case 'timeout':
      case 'api':
        return true;
      case 'authentication':
      case 'permission':
      case 'validation':
        return false;
      default:
        return true;
    }
  }

  /**
   * Check if error can be retried based on context
   */
  private canRetryError(errorType: ErrorType, context: ErrorContext): boolean {
    if (!this.shouldRetryErrorType(errorType)) {
      return false;
    }

    // Check retry count
    if (context.retryCount >= 3) {
      return false;
    }

    return true;
  }

  /**
   * Get error history for analysis
   */
  getErrorHistory(operationKey?: string): ErrorContext[] {
    if (operationKey) {
      return this.errorHistory.get(operationKey) || [];
    }

    const allErrors: ErrorContext[] = [];
    this.errorHistory.forEach(errors => allErrors.push(...errors));
    return allErrors.sort((a, b) => b.timestamp - a.timestamp);
  }

  /**
   * Clear error history
   */
  clearErrorHistory(operationKey?: string): void {
    if (operationKey) {
      this.errorHistory.delete(operationKey);
    } else {
      this.errorHistory.clear();
    }
  }

  /**
   * Utility method for delays
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Export singleton instance
export const errorHandlingService = ErrorHandlingService.getInstance();
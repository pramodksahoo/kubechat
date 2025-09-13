import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Button, Heading, Body, Code } from '../../design-system';
import { ScreenReaderUtils } from '../../design-system/accessibility';

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
  errorId: string;
}

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo, errorId: string) => void;
  isolate?: boolean;
  resetOnPropsChange?: boolean;
  resetKeys?: Array<string | number>;
}

// Generate unique error ID for tracking
function generateErrorId(): string {
  return `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

// Error reporting service
class ErrorReportingService {
  static report(error: Error, errorInfo: ErrorInfo, errorId: string, context?: Record<string, unknown>) {
    // In production, this would send to your error monitoring service
    // (e.g., Sentry, LogRocket, Bugsnag, etc.)
    console.group(`🚨 Error Boundary Caught Error [${errorId}]`);
    console.error('Error:', error);
    console.error('Error Info:', errorInfo);
    console.error('Component Stack:', errorInfo.componentStack);
    console.error('Context:', context);
    console.groupEnd();

    // Send to monitoring service in production
    if (process.env.NODE_ENV === 'production') {
      // Example: Sentry.captureException(error, { tags: { errorId }, extra: { errorInfo, context } });
    }

    // Announce to screen readers
    ScreenReaderUtils.announce(
      `An error occurred in the application. Error ID: ${errorId}. Please try refreshing the page or contact support.`,
      'assertive'
    );
  }
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  private resetTimeoutId: number | null = null;

  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: '',
    };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return {
      hasError: true,
      error,
      errorId: generateErrorId(),
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const { onError } = this.props;
    const { errorId } = this.state;

    this.setState({ errorInfo });

    // Report the error
    ErrorReportingService.report(error, errorInfo, errorId, {
      props: this.props,
      userAgent: navigator.userAgent,
      url: window.location.href,
      timestamp: new Date().toISOString(),
    });

    // Call custom error handler
    onError?.(error, errorInfo, errorId);
  }

  componentDidUpdate(prevProps: ErrorBoundaryProps) {
    const { resetOnPropsChange, resetKeys } = this.props;
    const { hasError } = this.state;

    // Reset error boundary when specified props change
    if (hasError && resetOnPropsChange && prevProps.children !== this.props.children) {
      this.resetErrorBoundary();
    }

    // Reset when reset keys change
    if (hasError && resetKeys) {
      const prevResetKeys = prevProps.resetKeys || [];
      if (resetKeys.some((key, idx) => key !== prevResetKeys[idx])) {
        this.resetErrorBoundary();
      }
    }
  }

  resetErrorBoundary = () => {
    if (this.resetTimeoutId) {
      clearTimeout(this.resetTimeoutId);
    }

    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: '',
    });

    ScreenReaderUtils.announce('The error has been cleared. The application is ready to use again.');
  };

  handleRetry = () => {
    this.resetErrorBoundary();
  };

  handleReload = () => {
    window.location.reload();
  };

  handleReport = () => {
    const { error, errorId } = this.state;
    if (error && errorId) {
      // Copy error details to clipboard
      const errorDetails = `Error ID: ${errorId}\nError: ${error.message}\nURL: ${window.location.href}\nTimestamp: ${new Date().toISOString()}`;
      
      navigator.clipboard.writeText(errorDetails).then(() => {
        ScreenReaderUtils.announce('Error details copied to clipboard. Please share this with support.');
      }).catch(() => {
        // Fallback: show error details in alert
        alert(`Error details:\n\n${errorDetails}`);
      });
    }
  };

  render() {
    const { hasError, error, errorInfo, errorId } = this.state;
    const { children, fallback, isolate = false } = this.props;

    if (hasError) {
      // Custom fallback UI
      if (fallback) {
        return fallback;
      }

      // Default error UI
      return (
        <div className={`error-boundary ${isolate ? 'isolate' : ''} p-6 max-w-2xl mx-auto`}>
          <div className="enterprise-card p-8 text-center">
            {/* Error Icon */}
            <div className="mx-auto w-16 h-16 mb-6 text-error-500">
              <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </div>

            {/* Error Message */}
            <Heading level={2} className="text-error-700 dark:text-error-400 mb-4">
              Something went wrong
            </Heading>

            <Body size="lg" color="secondary" className="mb-6">
              We apologize for the inconvenience. An unexpected error occurred while loading this part of the application.
            </Body>

            {/* Error ID */}
            <div className="bg-background-secondary p-4 rounded-lg mb-6">
              <Body size="sm" color="tertiary" className="mb-2">
                Error ID for support:
              </Body>
              <Code className="text-error-600 dark:text-error-400 font-semibold">
                {errorId}
              </Code>
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 justify-center">
              <Button 
                variant="primary" 
                onClick={this.handleRetry}
                leftIcon={
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                }
              >
                Try Again
              </Button>

              <Button 
                variant="secondary" 
                onClick={this.handleReload}
                leftIcon={
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                }
              >
                Reload Page
              </Button>

              <Button 
                variant="outline" 
                onClick={this.handleReport}
                leftIcon={
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                }
              >
                Copy Error Details
              </Button>
            </div>

            {/* Development Error Details */}
            {process.env.NODE_ENV === 'development' && error && (
              <details className="mt-8 text-left">
                <summary className="cursor-pointer text-sm font-medium text-text-secondary hover:text-text-primary mb-4">
                  Show Technical Details (Development Only)
                </summary>
                
                <div className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm font-mono overflow-auto">
                  <div className="mb-4">
                    <strong className="text-red-400">Error:</strong>
                    <pre className="mt-1 whitespace-pre-wrap">{error.toString()}</pre>
                  </div>
                  
                  {error.stack && (
                    <div className="mb-4">
                      <strong className="text-red-400">Stack Trace:</strong>
                      <pre className="mt-1 whitespace-pre-wrap text-xs">{error.stack}</pre>
                    </div>
                  )}
                  
                  {errorInfo?.componentStack && (
                    <div>
                      <strong className="text-red-400">Component Stack:</strong>
                      <pre className="mt-1 whitespace-pre-wrap text-xs">{errorInfo.componentStack}</pre>
                    </div>
                  )}
                </div>
              </details>
            )}
          </div>
        </div>
      );
    }

    return children;
  }
}

// Higher-order component for wrapping components with error boundary
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>
) {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  );

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;
  
  return WrappedComponent;
}

// Hook for using error boundary features
export function useErrorHandler() {
  return {
    reportError: (error: Error, context?: Record<string, unknown>) => {
      const errorId = generateErrorId();
      const errorInfo: ErrorInfo = {
        componentStack: 'Manual error report',
      };
      
      ErrorReportingService.report(error, errorInfo, errorId, context);
      return errorId;
    },
    
    throwError: (message: string, context?: Record<string, unknown>) => {
      const error = new Error(message);
      const errorId = generateErrorId();
      const errorInfo: ErrorInfo = {
        componentStack: 'Manual error throw',
      };
      
      ErrorReportingService.report(error, errorInfo, errorId, context);
      throw error;
    },
  };
}

// Specialized error boundaries for different parts of the app

// Route-level error boundary
export function RouteErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary
      onError={(error, errorInfo, errorId) => {
        // Route-specific error handling
        console.log('Route error:', { error, errorInfo, errorId });
      }}
      resetOnPropsChange={true}
    >
      {children}
    </ErrorBoundary>
  );
}

// Component-level error boundary for isolation
export function ComponentErrorBoundary({ 
  children, 
  componentName 
}: { 
  children: ReactNode;
  componentName: string;
}) {
  return (
    <ErrorBoundary
      isolate={true}
      onError={(error, errorInfo, errorId) => {
        console.log(`${componentName} error:`, { error, errorInfo, errorId });
      }}
      fallback={
        <div className="p-4 border border-error-200 dark:border-error-800 rounded-lg bg-error-50 dark:bg-error-900/20">
          <Body size="sm" color="error">
            Failed to load {componentName}. Please try refreshing the page.
          </Body>
        </div>
      }
    >
      {children}
    </ErrorBoundary>
  );
}

// Async operation error boundary
export function AsyncErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary
      onError={(error, errorInfo, errorId) => {
        // Async operation specific error handling
        console.log('Async operation error:', { error, errorInfo, errorId });
      }}
      resetKeys={[Date.now()]} // Auto-reset for retries
    >
      {children}
    </ErrorBoundary>
  );
}
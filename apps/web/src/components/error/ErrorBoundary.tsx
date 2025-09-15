import React, { Component, ErrorInfo, ReactNode } from 'react';
import { errorService } from '@/services/errorService';
import { ErrorDisplay } from './ErrorDisplay';

function generateErrorId(): string {
  return `error-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}

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

    // Use our error service to handle the error
    const managedError = errorService.handleError(error, {
      source: 'react-error-boundary',
      type: 'system',
      severity: 'high',
      retryable: false,
      userAction: 'Please refresh the page or contact support if the problem persists'
    });

    // Call custom error handler
    onError?.(error, errorInfo, errorId || managedError.id);
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
  };

  handleRetry = () => {
    this.resetErrorBoundary();
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    const { hasError, error, errorInfo, errorId } = this.state;
    const { children, fallback } = this.props;

    if (hasError) {
      // Custom fallback UI
      if (fallback) {
        return fallback;
      }

      // Default error UI using ErrorDisplay
      return (
        <ErrorDisplay
          error={error}
          errorInfo={errorInfo}
          errorId={errorId}
          onRetry={this.handleRetry}
          onReload={this.handleReload}
          showDetails={process.env.NODE_ENV === 'development'}
        />
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
      const errorInfo = errorService.handleError(error, {
        source: 'manual-report',
        type: 'user',
        severity: 'medium'
      });
      return errorInfo.id;
    },

    throwError: (message: string, context?: Record<string, unknown>) => {
      const error = new Error(message);
      errorService.handleError(error, {
        source: 'manual-throw',
        type: 'user',
        severity: 'medium'
      });
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
        <div className="p-4 border border-red-200 dark:border-red-800 rounded-lg bg-red-50 dark:bg-red-900/20">
          <p className="text-sm text-red-700 dark:text-red-300">
            Failed to load {componentName}. Please try refreshing the page.
          </p>
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
        console.log('Async operation error:', { error, errorInfo, errorId });
      }}
      resetKeys={[Date.now()]} // Auto-reset for retries
    >
      {children}
    </ErrorBoundary>
  );
}
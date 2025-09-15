import React from 'react';
import { Button } from '../ui/Button';
import { Card } from '../ui/Card';
import { errorService, ErrorInfo as ErrorServiceInfo } from '@/services/errorService';

interface ErrorDisplayProps {
  error: Error | null;
  errorInfo?: React.ErrorInfo | null;
  errorId?: string | null;
  onRetry?: () => void;
  onReload?: () => void;
  showDetails?: boolean;
  className?: string;
}

export function ErrorDisplay({
  error,
  errorInfo,
  errorId,
  onRetry,
  onReload,
  showDetails = false,
  className = ''
}: ErrorDisplayProps) {
  const [detailsVisible, setDetailsVisible] = React.useState(false);

  const handleCopyError = async () => {
    if (!error || !errorId) return;

    const errorDetails = [
      `Error ID: ${errorId}`,
      `Error: ${error.message}`,
      `URL: ${window.location.href}`,
      `Timestamp: ${new Date().toISOString()}`,
      `User Agent: ${navigator.userAgent}`,
      error.stack ? `Stack: ${error.stack}` : ''
    ].filter(Boolean).join('\n');

    try {
      await navigator.clipboard.writeText(errorDetails);
      console.log('Error details copied to clipboard');
    } catch (err) {
      console.error('Failed to copy error details:', err);
    }
  };

  const handleReload = () => {
    if (onReload) {
      onReload();
    } else {
      window.location.reload();
    }
  };

  const getErrorIcon = () => (
    <svg className="w-16 h-16 mx-auto text-red-500 dark:text-red-400" fill="currentColor" viewBox="0 0 20 20">
      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
    </svg>
  );

  return (
    <div className={`error-display p-6 max-w-2xl mx-auto ${className}`}>
      <Card className="p-8 text-center">
        {/* Error Icon */}
        <div className="mb-6">
          {getErrorIcon()}
        </div>

        {/* Error Message */}
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          Something went wrong
        </h2>

        <p className="text-gray-600 dark:text-gray-400 mb-6 text-lg">
          We apologize for the inconvenience. An unexpected error occurred while loading this part of the application.
        </p>

        {/* Error ID */}
        {errorId && (
          <div className="bg-gray-100 dark:bg-gray-800 p-4 rounded-lg mb-6">
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
              Error ID for support:
            </p>
            <code className="text-red-600 dark:text-red-400 font-mono font-semibold text-sm">
              {errorId}
            </code>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-3 justify-center mb-6">
          {onRetry && (
            <Button
              onClick={onRetry}
              variant="primary"
              className="flex items-center justify-center space-x-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <span>Try Again</span>
            </Button>
          )}

          <Button
            onClick={handleReload}
            variant="secondary"
            className="flex items-center justify-center space-x-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span>Reload Page</span>
          </Button>

          <Button
            onClick={handleCopyError}
            variant="ghost"
            className="flex items-center justify-center space-x-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <span>Copy Error Details</span>
          </Button>
        </div>

        {/* Show Details Toggle */}
        {(showDetails || process.env.NODE_ENV === 'development') && error && (
          <div className="text-left">
            <button
              onClick={() => setDetailsVisible(!detailsVisible)}
              className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 mb-4"
            >
              {detailsVisible ? 'Hide' : 'Show'} Technical Details {process.env.NODE_ENV === 'development' ? '(Development Only)' : ''}
            </button>

            {detailsVisible && (
              <div className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm font-mono overflow-auto max-h-96">
                <div className="mb-4">
                  <strong className="text-red-400">Error:</strong>
                  <pre className="mt-1 whitespace-pre-wrap text-white">{error.toString()}</pre>
                </div>

                {error.stack && (
                  <div className="mb-4">
                    <strong className="text-red-400">Stack Trace:</strong>
                    <pre className="mt-1 whitespace-pre-wrap text-xs text-gray-300">{error.stack}</pre>
                  </div>
                )}

                {errorInfo?.componentStack && (
                  <div>
                    <strong className="text-red-400">Component Stack:</strong>
                    <pre className="mt-1 whitespace-pre-wrap text-xs text-gray-300">{errorInfo.componentStack}</pre>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </Card>
    </div>
  );
}

// Inline error display for smaller components
export function InlineErrorDisplay({
  message,
  onRetry,
  className = ''
}: {
  message: string;
  onRetry?: () => void;
  className?: string;
}) {
  return (
    <div className={`inline-error-display p-4 border border-red-200 dark:border-red-800 rounded-lg bg-red-50 dark:bg-red-900/20 ${className}`}>
      <div className="flex items-start space-x-3">
        <svg className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
        </svg>
        <div className="flex-1">
          <p className="text-sm text-red-700 dark:text-red-300">
            {message}
          </p>
          {onRetry && (
            <button
              onClick={onRetry}
              className="mt-2 text-sm text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-200 font-medium"
            >
              Try again
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

// Loading error display
export function LoadingErrorDisplay({
  message = 'Failed to load content',
  onRetry,
  className = ''
}: {
  message?: string;
  onRetry?: () => void;
  className?: string;
}) {
  return (
    <div className={`loading-error-display text-center p-8 ${className}`}>
      <svg className="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>

      <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
        {message}
      </h3>

      <p className="text-gray-600 dark:text-gray-400 mb-4">
        There was a problem loading this content.
      </p>

      {onRetry && (
        <Button
          onClick={onRetry}
          variant="primary"
          size="sm"
        >
          Retry
        </Button>
      )}
    </div>
  );
}
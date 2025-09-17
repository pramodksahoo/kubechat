import { useState } from 'react';

export type ErrorType = 'network' | 'api' | 'validation' | 'authentication' | 'permission' | 'timeout' | 'general';

interface ErrorDisplayProps {
  error: string | Error;
  type?: ErrorType;
  context?: string;
  onRetry?: () => void;
  onDismiss?: () => void;
  showDetails?: boolean;
  className?: string;
}

export function ErrorDisplay({
  error,
  type = 'general',
  context,
  onRetry,
  onDismiss,
  showDetails = false,
  className = ''
}: ErrorDisplayProps) {
  const [expanded, setExpanded] = useState(false);

  const errorMessage = error instanceof Error ? error.message : error;
  const errorStack = error instanceof Error ? error.stack : undefined;

  const getErrorConfig = (errorType: ErrorType) => {
    switch (errorType) {
      case 'network':
        return {
          icon: '🌐',
          title: 'Network Error',
          description: 'Unable to connect to the server. Please check your internet connection.',
          bgColor: 'bg-red-50 dark:bg-red-900/20',
          borderColor: 'border-red-200 dark:border-red-800',
          textColor: 'text-red-800 dark:text-red-200',
          iconColor: 'text-red-600',
          showRetry: true
        };
      case 'api':
        return {
          icon: '⚠️',
          title: 'API Error',
          description: 'The server encountered an error while processing your request.',
          bgColor: 'bg-orange-50 dark:bg-orange-900/20',
          borderColor: 'border-orange-200 dark:border-orange-800',
          textColor: 'text-orange-800 dark:text-orange-200',
          iconColor: 'text-orange-600',
          showRetry: true
        };
      case 'validation':
        return {
          icon: '📝',
          title: 'Validation Error',
          description: 'Please check your input and try again.',
          bgColor: 'bg-yellow-50 dark:bg-yellow-900/20',
          borderColor: 'border-yellow-200 dark:border-yellow-800',
          textColor: 'text-yellow-800 dark:text-yellow-200',
          iconColor: 'text-yellow-600',
          showRetry: false
        };
      case 'authentication':
        return {
          icon: '🔐',
          title: 'Authentication Error',
          description: 'You need to log in to perform this action.',
          bgColor: 'bg-purple-50 dark:bg-purple-900/20',
          borderColor: 'border-purple-200 dark:border-purple-800',
          textColor: 'text-purple-800 dark:text-purple-200',
          iconColor: 'text-purple-600',
          showRetry: false
        };
      case 'permission':
        return {
          icon: '🚫',
          title: 'Permission Denied',
          description: 'You don\'t have permission to perform this action.',
          bgColor: 'bg-red-50 dark:bg-red-900/20',
          borderColor: 'border-red-200 dark:border-red-800',
          textColor: 'text-red-800 dark:text-red-200',
          iconColor: 'text-red-600',
          showRetry: false
        };
      case 'timeout':
        return {
          icon: '⏱️',
          title: 'Request Timeout',
          description: 'The request took too long to complete. Please try again.',
          bgColor: 'bg-blue-50 dark:bg-blue-900/20',
          borderColor: 'border-blue-200 dark:border-blue-800',
          textColor: 'text-blue-800 dark:text-blue-200',
          iconColor: 'text-blue-600',
          showRetry: true
        };
      default:
        return {
          icon: '❌',
          title: 'Error',
          description: 'An unexpected error occurred.',
          bgColor: 'bg-gray-50 dark:bg-gray-900/20',
          borderColor: 'border-gray-200 dark:border-gray-800',
          textColor: 'text-gray-800 dark:text-gray-200',
          iconColor: 'text-gray-600',
          showRetry: true
        };
    }
  };

  const config = getErrorConfig(type);

  const getSuggestions = (errorType: ErrorType, errorMessage: string) => {
    const suggestions: string[] = [];

    switch (errorType) {
      case 'network':
        suggestions.push('Check your internet connection');
        suggestions.push('Try refreshing the page');
        suggestions.push('Check if the server is accessible');
        break;
      case 'api':
        if (errorMessage.includes('404')) {
          suggestions.push('The requested resource was not found');
          suggestions.push('Check if the API endpoint is correct');
        } else if (errorMessage.includes('500')) {
          suggestions.push('Server internal error');
          suggestions.push('Try again in a few moments');
          suggestions.push('Contact support if the problem persists');
        }
        break;
      case 'validation':
        suggestions.push('Check that all required fields are filled');
        suggestions.push('Ensure your input follows the expected format');
        break;
      case 'authentication':
        suggestions.push('Log in to your account');
        suggestions.push('Check if your session has expired');
        break;
      case 'permission':
        suggestions.push('Contact your administrator for access');
        suggestions.push('Ensure you have the necessary permissions');
        break;
      case 'timeout':
        suggestions.push('Try the operation again');
        suggestions.push('Check your internet connection speed');
        suggestions.push('Try with a smaller request if possible');
        break;
    }

    return suggestions;
  };

  const suggestions = getSuggestions(type, errorMessage);

  return (
    <div className={`${config.bgColor} ${config.borderColor} border rounded-lg ${className}`}>
      <div className="p-4">
        {/* Error Header */}
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-3">
            <span className={`text-xl ${config.iconColor}`}>{config.icon}</span>
            <div className="flex-1">
              <h3 className={`font-medium ${config.textColor}`}>
                {config.title}
                {context && <span className="text-sm font-normal"> - {context}</span>}
              </h3>
              <p className={`mt-1 text-sm ${config.textColor} opacity-80`}>
                {config.description}
              </p>
              <div className={`mt-2 text-sm ${config.textColor}`}>
                {errorMessage}
              </div>
            </div>
          </div>

          {/* Dismiss Button */}
          {onDismiss && (
            <button
              onClick={onDismiss}
              className={`${config.textColor} hover:opacity-70 transition-opacity`}
              title="Dismiss error"
            >
              ✕
            </button>
          )}
        </div>

        {/* Error Suggestions */}
        {suggestions.length > 0 && (
          <div className="mt-3">
            <div className={`text-sm font-medium ${config.textColor} mb-2`}>
              Suggestions:
            </div>
            <ul className={`text-sm ${config.textColor} opacity-80 space-y-1`}>
              {suggestions.map((suggestion, index) => (
                <li key={index} className="flex items-start">
                  <span className="mr-2 text-xs">•</span>
                  <span>{suggestion}</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Action Buttons */}
        <div className="mt-4 flex items-center gap-3">
          {config.showRetry && onRetry && (
            <button
              onClick={onRetry}
              className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${
                config.iconColor.replace('text-', 'bg-').replace('600', '600')
              } text-white hover:opacity-90`}
            >
              Try Again
            </button>
          )}

          {showDetails && errorStack && (
            <button
              onClick={() => setExpanded(!expanded)}
              className={`text-sm ${config.textColor} hover:opacity-70 transition-opacity`}
            >
              {expanded ? 'Hide Details' : 'Show Details'}
            </button>
          )}
        </div>

        {/* Error Details */}
        {expanded && errorStack && (
          <div className="mt-3 p-3 bg-black/10 rounded-md">
            <div className="text-xs font-mono text-gray-600 dark:text-gray-400 whitespace-pre-wrap">
              {errorStack}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Specific error components for common scenarios
export function NetworkErrorDisplay({ error, onRetry, className }: {
  error: string | Error;
  onRetry?: () => void;
  className?: string;
}) {
  return (
    <ErrorDisplay
      error={error}
      type="network"
      onRetry={onRetry}
      className={className}
    />
  );
}

export function APIErrorDisplay({ error, context, onRetry, className }: {
  error: string | Error;
  context?: string;
  onRetry?: () => void;
  className?: string;
}) {
  return (
    <ErrorDisplay
      error={error}
      type="api"
      context={context}
      onRetry={onRetry}
      showDetails={true}
      className={className}
    />
  );
}

export function ValidationErrorDisplay({ error, className }: {
  error: string | Error;
  className?: string;
}) {
  return (
    <ErrorDisplay
      error={error}
      type="validation"
      className={className}
    />
  );
}

export function PermissionErrorDisplay({ error, className }: {
  error: string | Error;
  className?: string;
}) {
  return (
    <ErrorDisplay
      error={error}
      type="permission"
      className={className}
    />
  );
}
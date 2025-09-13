import React, { useState, useEffect } from 'react';
import { Button, Heading, Body, Code, StatusBadge } from '../../design-system';
import { cn } from '../../lib/utils';
import { ScreenReaderUtils } from '../../design-system/accessibility';

// Error severity levels
export type ErrorSeverity = 'low' | 'medium' | 'high' | 'critical';

// Error types for different contexts
export type ErrorType = 
  | 'network'
  | 'authentication'
  | 'authorization'
  | 'validation'
  | 'server'
  | 'timeout'
  | 'not_found'
  | 'maintenance'
  | 'kubernetes'
  | 'command_execution'
  | 'unknown';

// Base error interface
export interface AppError {
  id?: string;
  type: ErrorType;
  severity: ErrorSeverity;
  title: string;
  message: string;
  details?: string;
  context?: Record<string, unknown>;
  timestamp?: Date;
  retryable?: boolean;
  actions?: ErrorAction[];
}

// Error action interface
export interface ErrorAction {
  label: string;
  action: () => void | Promise<void>;
  variant?: 'primary' | 'secondary' | 'danger';
  loading?: boolean;
}

// Error icon mapping
const ErrorIcons: Record<ErrorType, React.ReactNode> = {
  network: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M17.778 8.222c-4.296-4.296-11.26-4.296-15.556 0A1 1 0 01.808 6.808c5.076-5.077 13.308-5.077 18.384 0a1 1 0 01-1.414 1.414zM14.95 11.05a7 7 0 00-9.9 0 1 1 0 01-1.414-1.414 9 9 0 0112.728 0 1 1 0 01-1.414 1.414zM12.12 13.88a3 3 0 00-4.242 0 1 1 0 01-1.415-1.415 5 5 0 017.072 0 1 1 0 01-1.415 1.415zM9 16a1 1 0 011-1h.01a1 1 0 110 2H10a1 1 0 01-1-1z" clipRule="evenodd" />
    </svg>
  ),
  authentication: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M18 8a6 6 0 01-7.743 5.743L10 14l-1 1-1 1H6v2H2v-4l4.257-4.257A6 6 0 1118 8zm-6-4a1 1 0 100 2 2 2 0 012 2 1 1 0 102 0 4 4 0 00-4-4z" clipRule="evenodd" />
    </svg>
  ),
  authorization: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
    </svg>
  ),
  validation: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
    </svg>
  ),
  server: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm0 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1v-2z" clipRule="evenodd" />
    </svg>
  ),
  timeout: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
    </svg>
  ),
  not_found: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
    </svg>
  ),
  maintenance: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" />
    </svg>
  ),
  kubernetes: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path d="M10 2L3 7l7 5 7-5-7-5zM3 8l7 5v5l-7-5V8zm14 0v5l-7 5v-5l7-5z" />
    </svg>
  ),
  command_execution: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5zm3.293 1.293a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 01-1.414-1.414L7.586 10 5.293 7.707a1 1 0 010-1.414zM11 12a1 1 0 100 2h3a1 1 0 100-2h-3z" clipRule="evenodd" />
    </svg>
  ),
  unknown: (
    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
    </svg>
  ),
};

// Severity color mapping
const SeverityColors: Record<ErrorSeverity, string> = {
  low: 'text-blue-600 dark:text-blue-400',
  medium: 'text-warning-600 dark:text-warning-400',
  high: 'text-error-600 dark:text-error-400',
  critical: 'text-error-700 dark:text-error-300',
};

// Main error display component
export interface ErrorDisplayProps {
  error: AppError;
  onRetry?: () => void;
  onDismiss?: () => void;
  showDetails?: boolean;
  compact?: boolean;
  className?: string;
  actions?: ErrorAction[];
}

export function ErrorDisplay({
  error,
  onRetry,
  onDismiss,
  showDetails = false,
  compact = false,
  className,
  // actions = [],
}: ErrorDisplayProps) {
  const [isDetailsOpen, setIsDetailsOpen] = useState(showDetails);
  
  useEffect(() => {
    // Announce errors to screen readers
    if (error.severity === 'critical' || error.severity === 'high') {
      ScreenReaderUtils.announce(
        `${error.severity} error: ${error.title}. ${error.message}`,
        'assertive'
      );
    } else {
      ScreenReaderUtils.announce(`Error: ${error.title}`, 'polite');
    }
  }, [error]);

  const icon = ErrorIcons[error.type];
  const severityColor = SeverityColors[error.severity];

  if (compact) {
    return (
      <div className={cn(
        'flex items-center gap-3 p-3 bg-error-50 dark:bg-error-900/20 border border-error-200 dark:border-error-800 rounded-lg',
        className
      )}>
        <div className={cn('w-5 h-5 flex-shrink-0', severityColor)}>
          {icon}
        </div>
        
        <div className="flex-1 min-w-0">
          <Body size="sm" className="text-error-800 dark:text-error-200 truncate">
            {error.message}
          </Body>
        </div>

        {(onRetry || onDismiss) && (
          <div className="flex gap-2">
            {onRetry && error.retryable !== false && (
              <Button size="sm" variant="outline" onClick={onRetry}>
                Retry
              </Button>
            )}
            {onDismiss && (
              <Button size="sm" variant="ghost" onClick={onDismiss}>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </Button>
            )}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className={cn(
      'enterprise-card p-6',
      error.severity === 'critical' && 'border-error-500',
      error.severity === 'high' && 'border-error-400',
      className
    )}>
      {/* Header */}
      <div className="flex items-start gap-4 mb-4">
        <div className={cn('w-8 h-8 flex-shrink-0 mt-1', severityColor)}>
          {icon}
        </div>
        
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 mb-2">
            <Heading level={3} className="text-error-800 dark:text-error-200">
              {error.title}
            </Heading>
            <StatusBadge variant={error.severity === 'critical' ? 'error' : 'warning'}>
              {error.severity}
            </StatusBadge>
          </div>
          
          <Body className="text-error-700 dark:text-error-300 mb-3">
            {error.message}
          </Body>

          {error.id && (
            <div className="mb-3">
              <Body size="sm" color="tertiary" className="mb-1">
                Error ID:
              </Body>
              <Code className="text-error-600 dark:text-error-400">
                {error.id}
              </Code>
            </div>
          )}

          {error.timestamp && (
            <Body size="sm" color="tertiary">
              Occurred at: {error.timestamp.toLocaleString()}
            </Body>
          )}
        </div>

        {onDismiss && (
          <Button variant="ghost" size="sm" onClick={onDismiss}>
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </Button>
        )}
      </div>

      {/* Actions */}
      {(error.actions || onRetry) && (
        <div className="flex flex-wrap gap-3 mb-4">
          {onRetry && error.retryable !== false && (
            <Button variant="primary" onClick={onRetry}>
              Try Again
            </Button>
          )}
          
          {error.actions?.map((action, index) => (
            <Button
              key={index}
              variant={action.variant || 'secondary'}
              onClick={action.action}
              loading={action.loading}
            >
              {action.label}
            </Button>
          ))}
        </div>
      )}

      {/* Details */}
      {error.details && (
        <div className="border-t border-border-primary pt-4">
          <button
            onClick={() => setIsDetailsOpen(!isDetailsOpen)}
            className="flex items-center gap-2 text-sm font-medium text-text-secondary hover:text-text-primary mb-3"
          >
            <svg 
              className={cn('w-4 h-4 transition-transform', isDetailsOpen && 'rotate-90')} 
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
            {isDetailsOpen ? 'Hide' : 'Show'} Technical Details
          </button>
          
          {isDetailsOpen && (
            <div className="bg-background-secondary p-4 rounded-lg">
              <Body size="sm" className="font-mono whitespace-pre-wrap">
                {error.details}
              </Body>
              
              {error.context && (
                <details className="mt-3">
                  <summary className="cursor-pointer text-sm font-medium text-text-secondary">
                    Additional Context
                  </summary>
                  <pre className="mt-2 text-xs text-text-tertiary overflow-auto">
                    {JSON.stringify(error.context, null, 2)}
                  </pre>
                </details>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// Network error component
export interface NetworkErrorProps {
  onRetry?: () => void;
  onGoOffline?: () => void;
}

export function NetworkError({ onRetry, onGoOffline }: NetworkErrorProps) {
  const error: AppError = {
    type: 'network',
    severity: 'high',
    title: 'Network Connection Error',
    message: 'Unable to connect to the server. Please check your internet connection and try again.',
    retryable: true,
  };

  return (
    <ErrorDisplay 
      error={error} 
      onRetry={onRetry}
      actions={onGoOffline ? [{
        label: 'Work Offline',
        action: onGoOffline,
        variant: 'secondary',
      }] : undefined}
    />
  );
}

// Command execution error component
export interface CommandErrorProps {
  command: string;
  exitCode?: number;
  stderr?: string;
  onRetry?: () => void;
  onEditCommand?: () => void;
}

export function CommandError({ 
  command, 
  exitCode, 
  stderr, 
  onRetry, 
  onEditCommand 
}: CommandErrorProps) {
  const error: AppError = {
    type: 'command_execution',
    severity: exitCode === 130 ? 'medium' : 'high', // 130 is user cancellation
    title: 'Command Execution Failed',
    message: `The kubectl command failed${exitCode ? ` with exit code ${exitCode}` : ''}.`,
    details: stderr,
    context: { command, exitCode },
    retryable: true,
  };

  const actions: ErrorAction[] = [];
  
  if (onEditCommand) {
    actions.push({
      label: 'Edit Command',
      action: onEditCommand,
      variant: 'secondary',
    });
  }

  return (
    <div className="space-y-4">
      <div className="bg-background-secondary p-4 rounded-lg">
        <Body size="sm" color="tertiary" className="mb-2">
          Failed Command:
        </Body>
        <Code className="block text-error-600 dark:text-error-400">
          {command}
        </Code>
      </div>
      
      <ErrorDisplay 
        error={error} 
        onRetry={onRetry}
        actions={actions}
        showDetails={!!stderr}
      />
    </div>
  );
}

// Maintenance mode component
export function MaintenanceMode() {
  const error: AppError = {
    type: 'maintenance',
    severity: 'medium',
    title: 'Scheduled Maintenance',
    message: 'KubeChat is currently undergoing scheduled maintenance. We\'ll be back shortly.',
    retryable: false,
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="max-w-md w-full">
        <ErrorDisplay error={error} />
        
        <div className="mt-6 text-center">
          <Body size="sm" color="tertiary">
            Estimated completion time: 15 minutes
          </Body>
        </div>
      </div>
    </div>
  );
}

// Permission denied component
export interface PermissionDeniedProps {
  resource?: string;
  requiredPermission?: string;
  onRequestAccess?: () => void;
}

export function PermissionDenied({ 
  resource, 
  requiredPermission, 
  onRequestAccess 
}: PermissionDeniedProps) {
  const error: AppError = {
    type: 'authorization',
    severity: 'medium',
    title: 'Permission Denied',
    message: `You don't have permission to ${resource ? `access ${resource}` : 'perform this action'}.`,
    retryable: false,
  };

  if (requiredPermission) {
    error.details = `Required permission: ${requiredPermission}`;
  }

  const actions: ErrorAction[] = [];
  
  if (onRequestAccess) {
    actions.push({
      label: 'Request Access',
      action: onRequestAccess,
      variant: 'primary',
    });
  }

  return <ErrorDisplay error={error} actions={actions} />;
}

// Loading failed component
export interface LoadingFailedProps {
  resource: string;
  onRetry?: () => void;
  onGoBack?: () => void;
}

export function LoadingFailed({ resource, onRetry, onGoBack }: LoadingFailedProps) {
  const error: AppError = {
    type: 'server',
    severity: 'medium',
    title: 'Failed to Load',
    message: `Unable to load ${resource}. This might be a temporary issue.`,
    retryable: true,
  };

  const actions: ErrorAction[] = [];
  
  if (onGoBack) {
    actions.push({
      label: 'Go Back',
      action: onGoBack,
      variant: 'secondary',
    });
  }

  return <ErrorDisplay error={error} onRetry={onRetry} actions={actions} />;
}

// Error toast notification
export interface ErrorToastProps {
  error: AppError;
  onDismiss: () => void;
  autoHide?: boolean;
  duration?: number;
}

export function ErrorToast({ 
  error, 
  onDismiss, 
  autoHide = true, 
  duration = 5000 
}: ErrorToastProps) {
  useEffect(() => {
    if (autoHide && error.severity !== 'critical') {
      const timer = setTimeout(onDismiss, duration);
      return () => clearTimeout(timer);
    }
  }, [autoHide, duration, error.severity, onDismiss]);

  return (
    <div className="fixed top-4 right-4 z-notification max-w-md animate-slide-left">
      <ErrorDisplay 
        error={error} 
        onDismiss={onDismiss} 
        compact={true}
        className="shadow-lg"
      />
    </div>
  );
}
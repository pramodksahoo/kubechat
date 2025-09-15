import { useState, useCallback, useEffect, useRef } from 'react';
import { errorService, ErrorInfo, ErrorAction } from '@/services/errorService';

export interface UseErrorHandlingOptions {
  retryLimit?: number;
  showErrorToUser?: boolean;
  autoRetry?: boolean;
  retryDelay?: number;
}

export interface ErrorState {
  error: ErrorInfo | null;
  hasError: boolean;
  isRetrying: boolean;
  retryCount: number;
  actions: ErrorAction[];
}

export function useErrorHandling(options: UseErrorHandlingOptions = {}) {
  const {
    retryLimit = 3,
    showErrorToUser = true,
    autoRetry = false,
    retryDelay = 1000
  } = options;

  const [errorState, setErrorState] = useState<ErrorState>({
    error: null,
    hasError: false,
    isRetrying: false,
    retryCount: 0,
    actions: []
  });

  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const retryCallbackRef = useRef<(() => Promise<void>) | null>(null);

  // Handle error with automatic classification and actions
  const handleError = useCallback((error: Error | string, context?: {
    source?: string;
    retryCallback?: () => Promise<void>;
    severity?: 'low' | 'medium' | 'high' | 'critical';
  }) => {
    const errorInfo = errorService.handleError(error, {
      source: context?.source || 'component',
      severity: context?.severity || 'medium',
      retryable: !!context?.retryCallback
    });

    const actions = errorService.getErrorActions(errorInfo);

    // Store retry callback if provided
    if (context?.retryCallback) {
      retryCallbackRef.current = context.retryCallback;
    }

    setErrorState({
      error: errorInfo,
      hasError: true,
      isRetrying: false,
      retryCount: 0,
      actions
    });

    // Auto-retry if enabled and error is retryable
    if (autoRetry && errorInfo.retryable && context?.retryCallback) {
      setTimeout(() => {
        retry();
      }, retryDelay);
    }
  }, [autoRetry, retryDelay]);

  // Retry the failed operation
  const retry = useCallback(async () => {
    if (!retryCallbackRef.current || errorState.retryCount >= retryLimit) {
      return;
    }

    setErrorState(prev => ({
      ...prev,
      isRetrying: true,
      retryCount: prev.retryCount + 1
    }));

    try {
      await retryCallbackRef.current();

      // Success - clear error
      setErrorState({
        error: null,
        hasError: false,
        isRetrying: false,
        retryCount: 0,
        actions: []
      });
    } catch (error) {
      // Failed again - update error state
      const newRetryCount = errorState.retryCount + 1;
      const canRetryAgain = newRetryCount < retryLimit;

      setErrorState(prev => ({
        ...prev,
        isRetrying: false,
        retryCount: newRetryCount
      }));

      // Auto-retry if enabled and within limit
      if (autoRetry && canRetryAgain) {
        retryTimeoutRef.current = setTimeout(() => {
          retry();
        }, retryDelay * Math.pow(2, newRetryCount)); // Exponential backoff
      }
    }
  }, [errorState.retryCount, retryLimit, autoRetry, retryDelay]);

  // Clear error state
  const clearError = useCallback(() => {
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }

    setErrorState({
      error: null,
      hasError: false,
      isRetrying: false,
      retryCount: 0,
      actions: []
    });

    retryCallbackRef.current = null;
  }, []);

  // Execute an error action
  const executeAction = useCallback(async (actionId: string) => {
    const action = errorState.actions.find(a => a.id === actionId);
    if (!action) return;

    switch (action.type) {
      case 'retry':
        await retry();
        break;
      case 'reload':
        window.location.reload();
        break;
      case 'navigate':
        if (action.href) {
          window.location.href = action.href;
        }
        break;
      case 'custom':
        if (action.handler) {
          await action.handler();
        }
        clearError();
        break;
    }
  }, [errorState.actions, retry, clearError]);

  // Wrap async operations with error handling
  const wrapAsync = useCallback(<T extends any[], R>(
    asyncFn: (...args: T) => Promise<R>,
    options?: {
      source?: string;
      onError?: (error: Error) => void;
      severity?: 'low' | 'medium' | 'high' | 'critical';
    }
  ) => {
    return async (...args: T): Promise<R | undefined> => {
      try {
        clearError(); // Clear previous errors
        return await asyncFn(...args);
      } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        options?.onError?.(err);

        handleError(err, {
          source: options?.source,
          severity: options?.severity,
          retryCallback: async () => { await asyncFn(...args); }
        });

        return undefined;
      }
    };
  }, [handleError, clearError]);

  // Handle network errors specifically
  const handleNetworkError = useCallback((error: Error, retryCallback?: () => Promise<void>) => {
    handleError(error, {
      source: 'network',
      severity: 'high',
      retryCallback
    });
  }, [handleError]);

  // Handle API errors specifically
  const handleApiError = useCallback((response: Response, retryCallback?: () => Promise<void>) => {
    const errorInfo = errorService.handleApiError(response, 'api-call');

    const actions = errorService.getErrorActions(errorInfo);

    if (retryCallback) {
      retryCallbackRef.current = retryCallback;
    }

    setErrorState({
      error: errorInfo,
      hasError: true,
      isRetrying: false,
      retryCount: 0,
      actions
    });
  }, []);

  // Clean up on unmount
  useEffect(() => {
    return () => {
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
    };
  }, []);

  return {
    // State
    ...errorState,
    canRetry: errorState.retryCount < retryLimit && !!retryCallbackRef.current,

    // Actions
    handleError,
    handleNetworkError,
    handleApiError,
    retry,
    clearError,
    executeAction,
    wrapAsync,

    // Utilities
    getUserFriendlyMessage: errorState.error ? errorService.getUserFriendlyMessage(errorState.error) : ''
  };
}

// Hook for global error handling
export function useGlobalErrorHandler() {
  const [errors, setErrors] = useState<ErrorInfo[]>([]);

  useEffect(() => {
    const unsubscribe = errorService.onError((error) => {
      setErrors(prev => [error, ...prev.slice(0, 9)]); // Keep last 10 errors
    });

    return unsubscribe;
  }, []);

  const clearAllErrors = useCallback(() => {
    setErrors([]);
    errorService.clearErrorHistory();
  }, []);

  const dismissError = useCallback((errorId: string) => {
    setErrors(prev => prev.filter(error => error.id !== errorId));
  }, []);

  const getCriticalErrors = useCallback(() => {
    return errors.filter(error => error.severity === 'critical');
  }, [errors]);

  return {
    errors,
    criticalErrors: getCriticalErrors(),
    clearAllErrors,
    dismissError,
    hasErrors: errors.length > 0,
    hasCriticalErrors: getCriticalErrors().length > 0
  };
}

// Hook for component-specific error handling
export function useComponentError(componentName: string) {
  const errorHandling = useErrorHandling({
    showErrorToUser: true,
    autoRetry: false
  });

  const handleComponentError = useCallback((error: Error | string) => {
    errorHandling.handleError(error, {
      source: `component-${componentName}`,
      severity: 'medium'
    });
  }, [errorHandling, componentName]);

  return {
    ...errorHandling,
    handleComponentError
  };
}

// Hook for async operation error handling with loading states
export function useAsyncError() {
  const [loading, setLoading] = useState(false);
  const errorHandling = useErrorHandling({
    autoRetry: true,
    retryLimit: 2,
    retryDelay: 1000
  });

  const executeAsync = useCallback(async <T>(
    operation: () => Promise<T>,
    options?: {
      source?: string;
      onSuccess?: (result: T) => void;
      onError?: (error: Error) => void;
    }
  ): Promise<T | undefined> => {
    setLoading(true);

    try {
      const result = await operation();
      options?.onSuccess?.(result);
      errorHandling.clearError();
      return result;
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      options?.onError?.(err);

      errorHandling.handleError(err, {
        source: options?.source || 'async-operation',
        retryCallback: async () => { await operation(); }
      });

      return undefined;
    } finally {
      setLoading(false);
    }
  }, [errorHandling]);

  return {
    ...errorHandling,
    loading,
    executeAsync
  };
}
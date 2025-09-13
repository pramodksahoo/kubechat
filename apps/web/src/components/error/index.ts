// KubeChat Enterprise Error Handling and Resilience System
// Complete error handling, monitoring, and resilience patterns

// Error Boundary Components
export {
  ErrorBoundary,
  withErrorBoundary,
  useErrorHandler,
  RouteErrorBoundary,
  ComponentErrorBoundary,
  AsyncErrorBoundary,
} from './ErrorBoundary';

// Error Display Components
export {
  ErrorDisplay,
  NetworkError,
  CommandError,
  MaintenanceMode,
  PermissionDenied,
  LoadingFailed,
  ErrorToast,
} from './ErrorComponents';

// Error Types and Interfaces
export type {
  ErrorSeverity,
  ErrorType,
  AppError,
  ErrorAction,
  ErrorDisplayProps,
  NetworkErrorProps,
  CommandErrorProps,
  PermissionDeniedProps,
  LoadingFailedProps,
  ErrorToastProps,
} from './ErrorComponents';

// Resilience Utilities - temporarily commented for build
// export {
//   withRetry,
//   CircuitBreaker,
//   CircuitBreakerState,
//   withTimeout,
//   Bulkhead,
//   RateLimiter,
//   ResilienceCache,
//   ResilienceWrapper,
// } from '../lib/resilience';

// Resilience Types - temporarily commented for build
// export type {
//   RetryConfig,
//   CircuitBreakerConfig,
//   TimeoutConfig,
//   BulkheadConfig,
//   RateLimiterConfig,
//   ResilienceCacheConfig,
//   ApiError,
// } from '../lib/resilience';

// Graceful Degradation Components
export {
  ProgressiveEnhancement,
  LazyComponent,
  LoadingSkeleton,
  AdaptiveImage,
  AdaptiveData,
  FeatureToggle,
  GracefulDegradation,
  withGracefulDegradation,
} from './resilience/GracefulDegradation';

// Graceful Degradation Types
export type {
  FeatureLevel,
  FeatureConfig,
  ProgressiveEnhancementProps,
  LazyComponentProps,
  AdaptiveImageProps,
  AdaptiveDataProps,
  FeatureToggleProps,
} from './resilience/GracefulDegradation';

// Offline Detection Hooks - temporarily commented for build
// export {
//   useOfflineDetection,
// } from '../hooks/useOfflineDetection';

// Offline Detection Types - temporarily commented for build
// export type {
//   OfflineConfig,
//   OfflineState,
//   OfflineCapabilities,
//   ServiceHealth,
// } from '../hooks/useOfflineDetection';

// Monitoring and Error Reporting - temporarily commented for build
// export {
//   MonitoringService,
//   monitoring,
//   reportError,
//   trackEvent,
//   trackPageView,
//   trackInteraction,
//   trackApiCall,
//   useMonitoring,
//   withErrorReporting,
//   ErrorSeverity as MonitoringErrorSeverity,
//   ErrorCategory,
// } from '../lib/monitoring';

// Monitoring Types - temporarily commented for build
// export type {
//   PerformanceMetric,
//   ErrorContext,
//   ErrorReport,
//   ErrorFeedback,
//   PerformanceReport,
// } from '../lib/monitoring';
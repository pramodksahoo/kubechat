import React, { useState, useEffect, useCallback, ReactNode, ComponentType } from 'react';
import Image from 'next/image';
import { Button, Body, Heading, StatusBadge } from '../../design-system';
import { cn } from '../../lib/utils';
import { useOfflineDetection } from '../../hooks/useOfflineDetection';
import { ScreenReaderUtils } from '../../design-system/accessibility';

// Feature flag types
export type FeatureLevel = 'essential' | 'enhanced' | 'experimental';

export interface FeatureConfig {
  level: FeatureLevel;
  fallback?: ReactNode;
  offlineFallback?: ReactNode;
  slowConnectionFallback?: ReactNode;
  errorFallback?: ReactNode;
  dependencies?: string[];
  minimumBandwidth?: number; // Kbps
  requiresAuth?: boolean;
  requiresPermissions?: string[];
}

// Progressive enhancement wrapper
export interface ProgressiveEnhancementProps {
  children: ReactNode;
  config: FeatureConfig;
  className?: string;
}

export function ProgressiveEnhancement({ 
  children, 
  config, 
  className 
}: ProgressiveEnhancementProps) {
  const { isOnline, isSlowConnection } = useOfflineDetection();
  const [isSupported, setIsSupported] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  // Check feature support
  useEffect(() => {
    // Check if all dependencies are available
    if (config.dependencies) {
      const missingDeps = config.dependencies.filter(dep => {
        // Check if dependency is available (could be API, browser feature, etc.)
        return !((window as unknown) as Record<string, unknown>)[dep];
      });
      
      if (missingDeps.length > 0) {
        setIsSupported(false);
        console.warn(`Feature not supported: missing dependencies ${missingDeps.join(', ')}`);
      }
    }
  }, [config.dependencies]);

  // Determine which fallback to show
  const getFallback = () => {
    if (error && config.errorFallback) {
      return config.errorFallback;
    }
    
    if (!isOnline && config.offlineFallback) {
      return config.offlineFallback;
    }
    
    if (isSlowConnection && config.slowConnectionFallback) {
      return config.slowConnectionFallback;
    }
    
    if (!isSupported && config.fallback) {
      return config.fallback;
    }
    
    return null;
  };

  const fallback = getFallback();
  
  if (fallback) {
    return (
      <div className={cn('progressive-enhancement-fallback', className)}>
        {fallback}
      </div>
    );
  }

  // Error boundary for this feature
  return (
    <FeatureErrorBoundary onError={setError} config={config}>
      <div className={cn('progressive-enhancement', className)}>
        {children}
      </div>
    </FeatureErrorBoundary>
  );
}

// Feature error boundary
interface FeatureErrorBoundaryProps {
  children: ReactNode;
  onError: (error: Error) => void;
  config: FeatureConfig;
}

class FeatureErrorBoundary extends React.Component<
  FeatureErrorBoundaryProps,
  { hasError: boolean; error: Error | null }
> {
  constructor(props: FeatureErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Feature error:', error, errorInfo);
    this.props.onError(error);
    
    ScreenReaderUtils.announce(
      `A feature encountered an error and has been disabled. Fallback functionality is available.`
    );
  }

  render() {
    if (this.state.hasError) {
      return this.props.config.errorFallback || (
        <div className="p-4 border border-warning-200 dark:border-warning-800 rounded-lg bg-warning-50 dark:bg-warning-900/20">
          <Body size="sm" color="warning">
            This feature is temporarily unavailable. Basic functionality is still available.
          </Body>
        </div>
      );
    }

    return this.props.children;
  }
}

// Lazy loading with graceful degradation
export interface LazyComponentProps<T = Record<string, never>> {
  loader: () => Promise<{ default: ComponentType<T> }>;
  fallback?: ReactNode;
  errorFallback?: ReactNode;
  props: T;
  timeout?: number;
  retryCount?: number;
}

export function LazyComponent<T = Record<string, never>>({
  loader,
  fallback = <LoadingSkeleton />,
  errorFallback,
  props,
  timeout = 10000,
  retryCount = 3,
}: LazyComponentProps<T>) {
  const [Component, setComponent] = useState<ComponentType<T> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [attempts, setAttempts] = useState(0);

  const loadComponent = useCallback(async () => {
    if (attempts >= retryCount) {
      setError(new Error('Maximum retry attempts reached'));
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);
    setAttempts(prev => prev + 1);

    const timeoutId = setTimeout(() => {
      setError(new Error('Component loading timeout'));
      setLoading(false);
    }, timeout);

    try {
      const loadedModule = await loader();
      clearTimeout(timeoutId);
      setComponent(() => loadedModule.default);
      setLoading(false);
    } catch (err) {
      clearTimeout(timeoutId);
      const error = err instanceof Error ? err : new Error('Component loading failed');
      setError(error);
      setLoading(false);
      
      // Retry with exponential backoff
      if (attempts < retryCount) {
        setTimeout(() => void loadComponent(), 1000 * Math.pow(2, attempts));
      }
    }
  }, [attempts, retryCount, loader, timeout]);

  useEffect(() => {
    void loadComponent();
  }, [loadComponent]);

  if (loading) {
    return <>{fallback}</>;
  }

  if (error) {
    if (errorFallback) {
      return <>{errorFallback}</>;
    }
    
    return (
      <div className="p-4 border border-error-200 dark:border-error-800 rounded-lg bg-error-50 dark:bg-error-900/20">
        <Body size="sm" color="error" className="mb-3">
          Failed to load component: {error.message}
        </Body>
        {attempts < retryCount && (
          <Button size="sm" variant="outline" onClick={loadComponent}>
            Retry ({retryCount - attempts} attempts left)
          </Button>
        )}
      </div>
    );
  }

  if (!Component) {
    return <>{errorFallback || <div>Component not available</div>}</>;
  }

  // @ts-expect-error - Dynamic component rendering with complex generics
  return <Component {...props} />;
}

// Loading skeleton component
export function LoadingSkeleton({ 
  lines = 3, 
  className 
}: { 
  lines?: number; 
  className?: string; 
}) {
  return (
    <div className={cn('animate-pulse space-y-3', className)}>
      {Array.from({ length: lines }).map((_, i) => (
        <div
          key={i}
          className={cn(
            'h-4 bg-gray-200 dark:bg-gray-700 rounded',
            i === lines - 1 && 'w-3/4' // Last line is shorter
          )}
        />
      ))}
    </div>
  );
}

// Adaptive image component with graceful degradation
export interface AdaptiveImageProps {
  src: string;
  alt: string;
  lowQualitySrc?: string;
  placeholder?: string;
  className?: string;
  priority?: boolean;
}

export function AdaptiveImage({
  src,
  alt,
  lowQualitySrc,
  placeholder,
  className,
  priority = false,
}: AdaptiveImageProps) {
  const { isSlowConnection, connectionQuality } = useOfflineDetection();
  const [imageLoaded, setImageLoaded] = useState(false);
  const [imageError, setImageError] = useState(false);
  const [currentSrc, setCurrentSrc] = useState<string>('');

  useEffect(() => {
    // Choose appropriate image quality based on connection
    if (connectionQuality === 'poor' && lowQualitySrc) {
      setCurrentSrc(lowQualitySrc);
    } else {
      setCurrentSrc(src);
    }
  }, [src, lowQualitySrc, connectionQuality]);

  const handleLoad = () => {
    setImageLoaded(true);
    setImageError(false);
  };

  const handleError = () => {
    setImageError(true);
    setImageLoaded(false);
    
    // Try low quality fallback if available
    if (currentSrc === src && lowQualitySrc) {
      setCurrentSrc(lowQualitySrc);
      setImageError(false);
    }
  };

  if (imageError && !lowQualitySrc) {
    return (
      <div className={cn(
        'flex items-center justify-center bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-600',
        className
      )}>
        <svg className="w-8 h-8" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
        </svg>
      </div>
    );
  }

  return (
    <div className={cn('relative', className)}>
      {/* Placeholder */}
      {!imageLoaded && placeholder && (
        <div className="absolute inset-0 bg-gray-100 dark:bg-gray-800">
          <Image
            src={placeholder}
            alt=""
            width={400}
            height={300}
            className="w-full h-full object-cover opacity-50 blur-sm"
          />
        </div>
      )}
      
      {/* Loading skeleton */}
      {!imageLoaded && !placeholder && (
        <div className="absolute inset-0">
          <LoadingSkeleton lines={1} className="h-full" />
        </div>
      )}
      
      {/* Main image */}
      <Image
        src={currentSrc}
        alt={alt}
        width={400}
        height={300}
        className={cn(
          'w-full h-full object-cover transition-opacity duration-300',
          !imageLoaded && 'opacity-0'
        )}
        onLoad={handleLoad}
        onError={handleError}
        priority={priority}
      />
      
      {/* Connection quality indicator */}
      {isSlowConnection && currentSrc === lowQualitySrc && (
        <div className="absolute top-2 right-2">
          <StatusBadge variant="warning">
            Low Quality
          </StatusBadge>
        </div>
      )}
    </div>
  );
}

// Adaptive data loading component
export interface AdaptiveDataProps<T> {
  children: (data: T, isStale: boolean) => ReactNode;
  loader: () => Promise<T>;
  fallbackData?: T;
  cacheKey: string;
  refreshInterval?: number;
  staleTime?: number;
  errorFallback?: (error: Error, retry: () => void) => ReactNode;
}

export function AdaptiveData<T>({
  children,
  loader,
  fallbackData,
  cacheKey,
  refreshInterval = 30000,
  staleTime = 60000,
  errorFallback,
}: AdaptiveDataProps<T>) {
  const { isOnline, connectionQuality } = useOfflineDetection();
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [isStale, setIsStale] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  // Load from cache
  const loadFromCache = useCallback((): T | null => {
    try {
      const cached = localStorage.getItem(`adaptive_data_${cacheKey}`);
      if (cached) {
        const { data: cachedData, timestamp } = JSON.parse(cached);
        const age = Date.now() - timestamp;
        setIsStale(age > staleTime);
        setLastUpdated(new Date(timestamp));
        return cachedData;
      }
    } catch {
      // Cache read failed
    }
    return null;
  }, [cacheKey, staleTime]);

  // Save to cache
  const saveToCache = useCallback((newData: T) => {
    try {
      const cacheItem = {
        data: newData,
        timestamp: Date.now(),
      };
      localStorage.setItem(`adaptive_data_${cacheKey}`, JSON.stringify(cacheItem));
    } catch {
      // Cache write failed - continue without caching
    }
  }, [cacheKey]);

  // Load data
  const loadData = useCallback(async (useCache = true) => {
    setError(null);
    
    // Try cache first if allowed
    if (useCache) {
      const cachedData = loadFromCache();
      if (cachedData) {
        setData(cachedData);
        setLoading(false);
        
        // Load fresh data in background if online
        if (isOnline && connectionQuality !== 'poor') {
          void loadData(false);
        }
        return;
      }
    }

    // Load fresh data
    if (!isOnline) {
      if (fallbackData) {
        setData(fallbackData);
        setIsStale(true);
      } else {
        setError(new Error('No internet connection and no cached data available'));
      }
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      const newData = await loader();
      setData(newData);
      setIsStale(false);
      setLastUpdated(new Date());
      saveToCache(newData);
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Data loading failed');
      setError(error);
      
      // Fallback to cache or fallback data
      const cachedData = loadFromCache();
      if (cachedData) {
        setData(cachedData);
        setIsStale(true);
      } else if (fallbackData) {
        setData(fallbackData);
        setIsStale(true);
      }
    } finally {
      setLoading(false);
    }
  }, [isOnline, connectionQuality, loader, fallbackData, loadFromCache, saveToCache]);

  // Auto-refresh data
  useEffect(() => {
    void loadData();
  }, [isOnline, loadData]);

  // Set up refresh interval
  useEffect(() => {
    if (!isOnline || connectionQuality === 'poor' || !refreshInterval) {
      return;
    }

    const interval = setInterval(() => {
      void loadData(false);
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [isOnline, connectionQuality, refreshInterval, loadData]);

  if (loading && !data) {
    return <LoadingSkeleton />;
  }

  if (error && !data) {
    if (errorFallback) {
      return errorFallback(error, () => loadData(false));
    }
    
    return (
      <div className="p-4 border border-error-200 dark:border-error-800 rounded-lg bg-error-50 dark:bg-error-900/20">
        <Body size="sm" color="error" className="mb-3">
          Failed to load data: {error.message}
        </Body>
        <Button size="sm" variant="outline" onClick={() => loadData(false)}>
          Retry
        </Button>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="p-4 text-center">
        <Body color="tertiary">No data available</Body>
      </div>
    );
  }

  return (
    <>
      {children(data, isStale)}
      
      {/* Stale data indicator */}
      {isStale && (
        <div className="mt-2 p-2 bg-warning-50 dark:bg-warning-900/20 border border-warning-200 dark:border-warning-800 rounded text-center">
          <Body size="sm" color="warning">
            Data may be outdated
            {lastUpdated && ` (last updated: ${lastUpdated.toLocaleTimeString()})`}
            {isOnline && (
              <Button 
                size="sm" 
                variant="outline" 
                className="ml-2" 
                onClick={() => loadData(false)}
              >
                Refresh
              </Button>
            )}
          </Body>
        </div>
      )}
    </>
  );
}

// Feature toggle component
export interface FeatureToggleProps {
  feature: string;
  children: ReactNode;
  fallback?: ReactNode;
  level?: FeatureLevel;
}

export function FeatureToggle({ 
  feature, 
  children, 
  fallback,
  level = 'enhanced' 
}: FeatureToggleProps) {
  const { connectionQuality } = useOfflineDetection();
  
  // Simple feature flag logic (in real app, this would come from a feature flag service)
  const isFeatureEnabled = () => {
    // Disable experimental features on poor connections
    if (level === 'experimental' && connectionQuality === 'poor') {
      return false;
    }
    
    // Check localStorage for feature flags
    try {
      const flags = JSON.parse(localStorage.getItem('feature_flags') || '{}');
      return flags[feature] !== false; // Default to enabled unless explicitly disabled
    } catch {
      return true; // Default to enabled if can't read flags
    }
  };

  if (!isFeatureEnabled()) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}

// Main GracefulDegradation wrapper component
export interface GracefulDegradationProps {
  children: ReactNode;
  fallback?: ReactNode;
}

export function GracefulDegradation({ children, fallback }: GracefulDegradationProps) {
  const { isOnline } = useOfflineDetection();

  // If offline and no fallback provided, show default offline message
  if (!isOnline && !fallback) {
    return (
      <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <Heading level={3} className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
              Connection Lost
            </Heading>
            <Body className="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
              You&apos;re currently offline. Some features may be limited until connection is restored.
            </Body>
          </div>
        </div>
      </div>
    );
  }

  // If offline and fallback provided, show fallback
  if (!isOnline && fallback) {
    return <>{fallback}</>;
  }

  // Show children normally if online
  return <>{children}</>;
}
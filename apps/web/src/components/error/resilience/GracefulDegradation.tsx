import React, { useState, useEffect, useCallback, ReactNode, ComponentType } from 'react';
import Image from 'next/image';
import { cn } from '../../../lib/utils';

// Types
export type FeatureLevel = 'full' | 'limited' | 'minimal' | 'offline';

export interface FeatureConfig {
  level: FeatureLevel;
  description: string;
  capabilities: string[];
  fallbackContent?: ReactNode;
}

export interface ProgressiveEnhancementProps {
  children: ReactNode;
  fallback?: ReactNode;
  level?: FeatureLevel;
  className?: string;
}

export interface LazyComponentProps {
  children: ReactNode;
  loading?: ReactNode;
  error?: ReactNode;
  delay?: number;
  className?: string;
}

export interface LoadingSkeletonProps {
  lines?: number;
  width?: string | number;
  height?: string | number;
  className?: string;
}

export interface AdaptiveImageProps {
  src: string;
  alt: string;
  lowResSrc?: string;
  placeholder?: string;
  className?: string;
  width?: number;
  height?: number;
}

export interface AdaptiveDataProps {
  children: ReactNode;
  fallback?: ReactNode;
  retryButton?: boolean;
  onRetry?: () => void;
  className?: string;
}

export interface FeatureToggleProps {
  feature: string;
  enabled: boolean;
  children: ReactNode;
  fallback?: ReactNode;
  className?: string;
}

export interface GracefulDegradationProps {
  children: ReactNode;
  fallback?: ReactNode;
  level?: FeatureLevel;
  onLevelChange?: (level: FeatureLevel) => void;
  className?: string;
}

// Progressive Enhancement Component
export function ProgressiveEnhancement({
  children,
  fallback,
  level = 'full',
  className,
}: ProgressiveEnhancementProps) {
  const [isSupported, setIsSupported] = useState(true);

  useEffect(() => {
    // Check for browser capabilities
    const checkSupport = () => {
      if (typeof window === 'undefined') return false;
      
      switch (level) {
        case 'full':
          return 'fetch' in window && 'Promise' in window && 'requestAnimationFrame' in window;
        case 'limited':
          return 'fetch' in window && 'Promise' in window;
        case 'minimal':
          return 'XMLHttpRequest' in window;
        default:
          return true;
      }
    };

    setIsSupported(checkSupport());
  }, [level]);

  if (!isSupported && fallback) {
    return <div className={cn('progressive-enhancement-fallback', className)}>{fallback}</div>;
  }

  return <div className={cn('progressive-enhancement', className)}>{children}</div>;
}

// Lazy Component Loader
export function LazyComponent({
  children,
  loading,
  error,
  delay = 0,
  className,
}: LazyComponentProps) {
  const [isLoaded, setIsLoaded] = useState(false);
  const [hasError, setHasError] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      try {
        setIsLoaded(true);
      } catch (err) {
        setHasError(true);
      }
    }, delay);

    return () => clearTimeout(timer);
  }, [delay]);

  if (hasError && error) {
    return <div className={cn('lazy-component-error', className)}>{error}</div>;
  }

  if (!isLoaded && loading) {
    return <div className={cn('lazy-component-loading', className)}>{loading}</div>;
  }

  return <div className={cn('lazy-component', className)}>{children}</div>;
}

// Loading Skeleton
export function LoadingSkeleton({
  lines = 3,
  width = '100%',
  height = '1rem',
  className,
}: LoadingSkeletonProps) {
  return (
    <div className={cn('animate-pulse space-y-2', className)}>
      {Array.from({ length: lines }).map((_, index) => (
        <div
          key={index}
          className="bg-gray-200 dark:bg-gray-700 rounded"
          style={{
            width: typeof width === 'number' ? `${width}px` : width,
            height: typeof height === 'number' ? `${height}px` : height,
          }}
        />
      ))}
    </div>
  );
}

// Adaptive Image
export function AdaptiveImage({
  src,
  alt,
  lowResSrc,
  placeholder,
  className,
  width,
  height,
}: AdaptiveImageProps) {
  const [isLoading, setIsLoading] = useState(true);
  const [hasError, setHasError] = useState(false);
  const [currentSrc, setCurrentSrc] = useState(lowResSrc || placeholder || '');

  useEffect(() => {
    const img = document.createElement('img');
    img.onload = () => {
      setCurrentSrc(src);
      setIsLoading(false);
    };
    img.onerror = () => {
      setHasError(true);
      setIsLoading(false);
    };
    img.src = src;
  }, [src]);

  if (hasError) {
    return (
      <div className={cn('bg-gray-100 dark:bg-gray-800 flex items-center justify-center', className)}>
        <span className="text-gray-500 text-sm">Failed to load image</span>
      </div>
    );
  }

  return (
    <Image
      src={currentSrc}
      alt={alt}
      width={width || 400}
      height={height || 300}
      className={cn(
        'transition-opacity duration-300',
        isLoading ? 'opacity-50' : 'opacity-100',
        className
      )}
    />
  );
}

// Adaptive Data Component
export function AdaptiveData({
  children,
  fallback,
  retryButton = true,
  onRetry,
  className,
}: AdaptiveDataProps) {
  const [hasData, setHasData] = useState(true);
  const [isRetrying, setIsRetrying] = useState(false);

  const handleRetry = useCallback(async () => {
    if (!onRetry) return;
    
    setIsRetrying(true);
    try {
      await onRetry();
      setHasData(true);
    } catch (error) {
      setHasData(false);
    } finally {
      setIsRetrying(false);
    }
  }, [onRetry]);

  if (!hasData) {
    return (
      <div className={cn('adaptive-data-fallback p-4 text-center', className)}>
        {fallback || (
          <div className="space-y-3">
            <p className="text-gray-600 dark:text-gray-400">Data unavailable</p>
            {retryButton && onRetry && (
              <button
                onClick={handleRetry}
                disabled={isRetrying}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
              >
                {isRetrying ? 'Retrying...' : 'Retry'}
              </button>
            )}
          </div>
        )}
      </div>
    );
  }

  return <div className={cn('adaptive-data', className)}>{children}</div>;
}

// Feature Toggle
export function FeatureToggle({
  feature: _feature,
  enabled,
  children,
  fallback,
  className,
}: FeatureToggleProps) {
  if (!enabled) {
    return fallback ? (
      <div className={cn('feature-toggle-fallback', className)}>{fallback}</div>
    ) : null;
  }

  return <div className={cn('feature-toggle', className)}>{children}</div>;
}

// Main Graceful Degradation Component
export function GracefulDegradation({
  children,
  fallback,
  level = 'full',
  onLevelChange,
  className,
}: GracefulDegradationProps) {
  const [currentLevel, setCurrentLevel] = useState<FeatureLevel>(level);

  useEffect(() => {
    setCurrentLevel(level);
    onLevelChange?.(level);
  }, [level, onLevelChange]);

  const shouldShowFallback = currentLevel === 'offline' || currentLevel === 'minimal';

  if (shouldShowFallback && fallback) {
    return <div className={cn('graceful-degradation-fallback', className)}>{fallback}</div>;
  }

  return <div className={cn('graceful-degradation', className)}>{children}</div>;
}

// HOC for graceful degradation
export function withGracefulDegradation<P extends object>(
  Component: ComponentType<P>,
  config: {
    fallback?: ReactNode;
    level?: FeatureLevel;
    className?: string;
  } = {}
) {
  const WrappedComponent = (props: P) => {
    return (
      <GracefulDegradation
        fallback={config.fallback}
        level={config.level}
        className={config.className}
      >
        <Component {...props} />
      </GracefulDegradation>
    );
  };

  WrappedComponent.displayName = `withGracefulDegradation(${Component.displayName || Component.name})`;
  
  return WrappedComponent;
}
import { useState, useEffect, useCallback, useMemo } from 'react';

// Types
export interface OfflineConfig {
  checkInterval?: number;
  timeout?: number;
  maxRetries?: number;
  endpoints?: string[];
  fallbackData?: boolean;
  staleWhileRevalidate?: boolean;
}

export interface ServiceHealth {
  name: string;
  url: string;
  healthy: boolean;
  latency?: number;
  lastChecked: Date;
  error?: string;
}

export interface OfflineCapabilities {
  cache: boolean;
  localStorage: boolean;
  indexedDB: boolean;
  serviceWorker: boolean;
  backgroundSync: boolean;
}

export interface OfflineState {
  isOnline: boolean;
  isSlowConnection: boolean;
  networkType: string;
  connectionQuality: 'excellent' | 'good' | 'fair' | 'poor';
  services: ServiceHealth[];
  capabilities: OfflineCapabilities;
  lastOnline: Date | null;
  isRecovering: boolean;
}

const DEFAULT_CONFIG: Required<OfflineConfig> = {
  checkInterval: 30000, // 30 seconds
  timeout: 5000, // 5 seconds
  maxRetries: 3,
  endpoints: ['/api/health'],
  fallbackData: true,
  staleWhileRevalidate: true,
};

export function useOfflineDetection(config: OfflineConfig = {}) {
  const finalConfig = useMemo(() => ({ ...DEFAULT_CONFIG, ...config }), [config]);
  
  const [state, setState] = useState<OfflineState>({
    isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
    isSlowConnection: false,
    networkType: 'unknown',
    connectionQuality: 'good',
    services: [],
    capabilities: {
      cache: false,
      localStorage: false,
      indexedDB: false,
      serviceWorker: false,
      backgroundSync: false,
    },
    lastOnline: null,
    isRecovering: false,
  });

  // Check browser capabilities
  const checkCapabilities = useCallback((): OfflineCapabilities => {
    if (typeof window === 'undefined') {
      return {
        cache: false,
        localStorage: false,
        indexedDB: false,
        serviceWorker: false,
        backgroundSync: false,
      };
    }

    return {
      cache: 'caches' in window,
      localStorage: 'localStorage' in window,
      indexedDB: 'indexedDB' in window,
      serviceWorker: 'serviceWorker' in navigator,
      backgroundSync: 'serviceWorker' in navigator && 'sync' in window.ServiceWorkerRegistration.prototype,
    };
  }, []);

  // Check service health
  const checkServiceHealth = useCallback(async (endpoint: string): Promise<ServiceHealth> => {
    const startTime = Date.now();
    
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), finalConfig.timeout);
      
      const response = await fetch(endpoint, {
        method: 'HEAD',
        signal: controller.signal,
        cache: 'no-cache',
      });
      
      clearTimeout(timeoutId);
      const latency = Date.now() - startTime;
      
      return {
        name: endpoint,
        url: endpoint,
        healthy: response.ok,
        latency,
        lastChecked: new Date(),
        error: response.ok ? undefined : `HTTP ${response.status}`,
      };
    } catch (error) {
      const latency = Date.now() - startTime;
      return {
        name: endpoint,
        url: endpoint,
        healthy: false,
        latency,
        lastChecked: new Date(),
        error: error instanceof Error ? error.message : 'Network error',
      };
    }
  }, [finalConfig.timeout]);

  // Determine connection quality based on latency and network type
  const determineConnectionQuality = useCallback((services: ServiceHealth[], networkType: string): 'excellent' | 'good' | 'fair' | 'poor' => {
    if (services.length === 0) return 'good';
    
    const averageLatency = services.reduce((sum, service) => sum + (service.latency || 1000), 0) / services.length;
    
    if (networkType.includes('slow-2g') || networkType === '2g') return 'poor';
    if (networkType === '3g' && averageLatency > 1000) return 'poor';
    if (averageLatency > 2000) return 'poor';
    if (averageLatency > 1000) return 'fair';
    if (averageLatency > 300) return 'good';
    return 'excellent';
  }, []);

  // Manual retry function  
  const retry = useCallback(async () => {
    setState(prev => ({ ...prev, isRecovering: true }));
    
    try {
      const services = finalConfig.endpoints.map(async endpoint => {
        return await checkServiceHealth(endpoint);
      });
      
      const results = await Promise.all(services);
      const isOnline = navigator.onLine && results.some(service => service.healthy);
      const connectionQuality = determineConnectionQuality(results, state.networkType);
      
      setState(prev => ({
        ...prev,
        isOnline,
        services: results,
        connectionQuality,
        isRecovering: false,
        lastOnline: isOnline ? new Date() : prev.lastOnline,
      }));
    } catch (error) {
      setState(prev => ({ ...prev, isRecovering: false }));
    }
  }, [finalConfig.endpoints, checkServiceHealth, determineConnectionQuality, state.networkType]);

  // Get offline data capabilities
  const getOfflineCapabilities = useCallback(() => {
    const { capabilities } = state;
    return {
      canCache: capabilities.cache,
      canStoreData: capabilities.localStorage || capabilities.indexedDB,
      canSyncLater: capabilities.serviceWorker && capabilities.backgroundSync,
      hasOfflineSupport: capabilities.localStorage || capabilities.indexedDB || capabilities.cache,
    };
  }, [state.capabilities, state]);

  // Check if service is healthy
  const isServiceHealthy = useCallback((serviceName: string) => {
    const service = state.services.find(s => s.name === serviceName || s.url === serviceName);
    return service?.healthy || false;
  }, [state.services]);

  // Get overall health score
  const getHealthScore = useCallback(() => {
    if (state.services.length === 0) return 1;
    const healthyServices = state.services.filter(s => s.healthy).length;
    return healthyServices / state.services.length;
  }, [state.services]);

  // Initialize capabilities and setup event listeners
  useEffect(() => {
    if (typeof window === 'undefined') return;

    setState(prev => ({
      ...prev,
      capabilities: checkCapabilities(),
    }));

    const handleOnline = () => {
      setState(prev => ({
        ...prev,
        isOnline: true,
        isRecovering: !prev.isOnline,
        lastOnline: new Date(),
      }));
    };

    const handleOffline = () => {
      setState(prev => ({
        ...prev,
        isOnline: false,
        isRecovering: false,
      }));
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [checkCapabilities]);

  return {
    ...state,
    retry,
    getOfflineCapabilities,
    isServiceHealthy,
    getHealthScore,
    config: finalConfig,
  };
}
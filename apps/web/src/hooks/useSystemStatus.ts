import { useState, useEffect, useCallback } from 'react';
import { statusService, SystemStatus, ServiceHealth } from '@/services/statusService';

export interface UseSystemStatusOptions {
  autoStart?: boolean;
  refreshInterval?: number;
}

export function useSystemStatus(options: UseSystemStatusOptions = {}) {
  const { autoStart = true, refreshInterval = 30000 } = options;

  const [systemStatus, setSystemStatus] = useState<SystemStatus | null>(null);
  const [serviceHealth, setServiceHealth] = useState<ServiceHealth[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Manual refresh
  const refresh = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const [status, health] = await Promise.all([
        statusService.checkSystemStatus(),
        statusService.checkServiceHealth(),
      ]);

      setSystemStatus(status);
      setServiceHealth(health);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch system status';
      setError(errorMessage);
      console.error('System status refresh failed:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Start/stop monitoring
  const startMonitoring = useCallback(() => {
    statusService.startMonitoring(refreshInterval);
  }, [refreshInterval]);

  const stopMonitoring = useCallback(() => {
    statusService.stopMonitoring();
  }, []);

  // Set up listeners and initial load
  useEffect(() => {
    // Subscribe to real-time status updates
    const unsubscribeStatus = statusService.onStatusUpdate((status) => {
      setSystemStatus(status);
      setLoading(false);
    });

    const unsubscribeHealth = statusService.onHealthUpdate((health) => {
      setServiceHealth(health);
    });

    // Start monitoring if auto-start enabled
    if (autoStart) {
      startMonitoring();
      // Initial load
      refresh();
    }

    return () => {
      unsubscribeStatus();
      unsubscribeHealth();
      if (autoStart) {
        stopMonitoring();
      }
    };
  }, [autoStart, refresh, startMonitoring, stopMonitoring]);

  // Calculate overall system health
  const overallHealth = useCallback(() => {
    if (!systemStatus) return 'unknown';

    const statuses = [
      systemStatus.cluster.status,
      systemStatus.llm.status,
      systemStatus.api.status,
      systemStatus.database.status,
      systemStatus.websocket.status,
    ];

    if (statuses.includes('error')) return 'error';
    if (statuses.includes('warning')) return 'warning';
    if (statuses.includes('connecting') || statuses.includes('unknown')) return 'connecting';
    if (statuses.every(status => status === 'healthy')) return 'healthy';

    return 'warning';
  }, [systemStatus]);

  // Get status for specific component
  const getComponentStatus = useCallback((component: keyof SystemStatus) => {
    return systemStatus?.[component] || null;
  }, [systemStatus]);

  // Check if specific service is healthy
  const isServiceHealthy = useCallback((serviceName: string) => {
    const service = serviceHealth.find(s => s.name === serviceName);
    return service?.status === 'healthy';
  }, [serviceHealth]);

  // Get services by status
  const getServicesByStatus = useCallback((status: string) => {
    return serviceHealth.filter(service => service.status === status);
  }, [serviceHealth]);

  // Get critical issues
  const getCriticalIssues = useCallback(() => {
    const issues: string[] = [];

    if (systemStatus?.cluster.status === 'error') {
      issues.push(`Cluster: ${systemStatus.cluster.details || 'Connection failed'}`);
    }
    if (systemStatus?.llm.status === 'error') {
      issues.push(`LLM: ${systemStatus.llm.details || 'Service unavailable'}`);
    }
    if (systemStatus?.api.status === 'error') {
      issues.push(`API: ${systemStatus.api.details || 'Service error'}`);
    }
    if (systemStatus?.database.status === 'error') {
      issues.push(`Database: ${systemStatus.database.details || 'Connection failed'}`);
    }

    // Add failed services
    const failedServices = getServicesByStatus('error');
    failedServices.forEach(service => {
      issues.push(`${service.name}: ${service.details || 'Service error'}`);
    });

    return issues;
  }, [systemStatus, getServicesByStatus]);

  return {
    // State
    systemStatus,
    serviceHealth,
    loading,
    error,

    // Actions
    refresh,
    startMonitoring,
    stopMonitoring,

    // Computed values
    overallHealth: overallHealth(),
    criticalIssues: getCriticalIssues(),

    // Helper functions
    getComponentStatus,
    isServiceHealthy,
    getServicesByStatus,
  };
}
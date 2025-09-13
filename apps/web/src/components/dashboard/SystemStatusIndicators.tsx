import React, { useEffect, useState } from 'react';
import { Icon } from '@/components/ui/Icon';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

interface SystemStatus {
  id: string;
  name: string;
  status: 'online' | 'degraded' | 'offline' | 'maintenance';
  lastChecked: string;
  responseTime?: number;
  uptime?: number;
  message?: string;
}

interface ConnectionStatus {
  isConnected: boolean;
  lastUpdate: string;
  reconnectAttempts?: number;
}

export interface SystemStatusIndicatorsProps extends BaseComponentProps {
  systems?: SystemStatus[];
  websocketStatus?: ConnectionStatus;
  autoRefresh?: boolean;
  refreshInterval?: number;
  onStatusClick?: (systemId: string) => void;
  showDetails?: boolean;
}

export const SystemStatusIndicators: React.FC<SystemStatusIndicatorsProps> = ({
  systems = [],
  websocketStatus,
  autoRefresh = true,
  refreshInterval = 30000, // 30 seconds
  onStatusClick,
  showDetails = false,
  className = '',
  'data-testid': dataTestId = 'system-status-indicators'
}) => {
  const [lastUpdated, setLastUpdated] = useState(new Date().toISOString());
  const [isRefreshing, setIsRefreshing] = useState(false);

  const defaultSystems: SystemStatus[] = [
    {
      id: 'kubernetes-api',
      name: 'Kubernetes API',
      status: 'online',
      lastChecked: new Date().toISOString(),
      responseTime: 45,
      uptime: 99.9,
      message: 'All clusters responding normally'
    },
    {
      id: 'llm-service',
      name: 'LLM Service',
      status: 'online',
      lastChecked: new Date().toISOString(),
      responseTime: 120,
      uptime: 98.5,
      message: 'OpenAI GPT-4 responding'
    },
    {
      id: 'database',
      name: 'Database',
      status: 'online',
      lastChecked: new Date().toISOString(),
      responseTime: 12,
      uptime: 99.99,
      message: 'PostgreSQL cluster healthy'
    },
    {
      id: 'redis-cache',
      name: 'Cache Layer',
      status: 'degraded',
      lastChecked: new Date().toISOString(),
      responseTime: 200,
      uptime: 97.2,
      message: 'High latency detected'
    },
    {
      id: 'auth-service',
      name: 'Authentication',
      status: 'online',
      lastChecked: new Date().toISOString(),
      responseTime: 35,
      uptime: 99.8,
      message: 'RBAC and JWT services operational'
    }
  ];

  const defaultWebsocketStatus: ConnectionStatus = {
    isConnected: true,
    lastUpdate: new Date().toISOString(),
    reconnectAttempts: 0
  };

  const displaySystems = systems.length > 0 ? systems : defaultSystems;
  const wsStatus = websocketStatus || defaultWebsocketStatus;

  // Auto-refresh effect
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      setIsRefreshing(true);
      // Simulate refresh
      setTimeout(() => {
        setLastUpdated(new Date().toISOString());
        setIsRefreshing(false);
      }, 500);
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval]);

  const getStatusConfig = (status: string) => {
    switch (status) {
      case 'online':
        return {
          color: 'text-success-500',
          bgColor: 'bg-success-500',
          label: 'Online',
          pulseClass: 'animate-pulse-subtle'
        };
      case 'degraded':
        return {
          color: 'text-warning-500',
          bgColor: 'bg-warning-500',
          label: 'Degraded',
          pulseClass: 'animate-pulse'
        };
      case 'offline':
        return {
          color: 'text-danger-500',
          bgColor: 'bg-danger-500',
          label: 'Offline',
          pulseClass: ''
        };
      case 'maintenance':
        return {
          color: 'text-info-500',
          bgColor: 'bg-info-500',
          label: 'Maintenance',
          pulseClass: 'animate-pulse-slow'
        };
      default:
        return {
          color: 'text-gray-500',
          bgColor: 'bg-gray-500',
          label: 'Unknown',
          pulseClass: ''
        };
    }
  };

  const getOverallStatus = () => {
    const statusCounts = displaySystems.reduce((acc, system) => {
      acc[system.status] = (acc[system.status] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);

    if (statusCounts.offline > 0) return 'critical';
    if (statusCounts.degraded > 0) return 'warning';
    return 'healthy';
  };

  const overallStatus = getOverallStatus();
  const overallConfig = getStatusConfig(overallStatus === 'critical' ? 'offline' : overallStatus === 'warning' ? 'degraded' : 'online');

  const formatUptime = (uptime: number) => {
    return `${uptime.toFixed(2)}%`;
  };

  const formatResponseTime = (time: number) => {
    return `${time}ms`;
  };

  const formatTimeAgo = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diffInSeconds = Math.floor((now.getTime() - time.getTime()) / 1000);
    
    if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
    
    const diffInMinutes = Math.floor(diffInSeconds / 60);
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
    
    const diffInHours = Math.floor(diffInMinutes / 60);
    return `${diffInHours}h ago`;
  };

  return (
    <div className={className} data-testid={dataTestId}>
      {/* Overall status header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center">
          <div className={`h-3 w-3 rounded-full ${overallConfig.bgColor} ${overallConfig.pulseClass} mr-3`} />
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              System Status
            </h3>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {overallStatus === 'critical' 
                ? 'Some services are experiencing issues'
                : overallStatus === 'warning'
                ? 'Some services are degraded'
                : 'All systems operational'
              }
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-3">
          {/* WebSocket status */}
          <div className="flex items-center text-sm text-gray-500 dark:text-gray-400">
            <div className={`h-2 w-2 rounded-full ${wsStatus.isConnected ? 'bg-success-500 animate-pulse-subtle' : 'bg-danger-500'} mr-2`} />
            <span>{wsStatus.isConnected ? 'Real-time' : 'Disconnected'}</span>
          </div>
          
          {/* Last updated */}
          <div className="text-xs text-gray-400">
            <Icon name="clock" className="h-3 w-3 inline mr-1" />
            Updated {formatTimeAgo(lastUpdated)}
          </div>
          
          {/* Refresh indicator */}
          {isRefreshing && (
            <Icon name="spinner" className="h-4 w-4 animate-spin text-primary-500" />
          )}
        </div>
      </div>

      {/* Compact status indicators */}
      {!showDetails && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {displaySystems.map((system) => {
            const statusConfig = getStatusConfig(system.status);
            
            return (
              <div
                key={system.id}
                className="p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-primary-300 dark:hover:border-primary-600 cursor-pointer transition-all duration-200"
                onClick={() => onStatusClick?.(system.id)}
                data-testid={`system-${system.id}`}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className={`h-2 w-2 rounded-full ${statusConfig.bgColor} ${statusConfig.pulseClass}`} />
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {system.responseTime ? formatResponseTime(system.responseTime) : '—'}
                  </span>
                </div>
                <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {system.name}
                </p>
                <p className={`text-xs ${statusConfig.color}`}>
                  {statusConfig.label}
                </p>
              </div>
            );
          })}
        </div>
      )}

      {/* Detailed status view */}
      {showDetails && (
        <div className="space-y-3">
          {displaySystems.map((system) => {
            const statusConfig = getStatusConfig(system.status);
            
            return (
              <div
                key={system.id}
                className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-primary-300 dark:hover:border-primary-600 cursor-pointer transition-all duration-200"
                onClick={() => onStatusClick?.(system.id)}
                data-testid={`system-detail-${system.id}`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <div className={`h-3 w-3 rounded-full ${statusConfig.bgColor} ${statusConfig.pulseClass} mr-3`} />
                    <div>
                      <h4 className="text-sm font-medium text-gray-900 dark:text-white">
                        {system.name}
                      </h4>
                      <p className="text-xs text-gray-500 dark:text-gray-400">
                        {system.message || 'No additional information'}
                      </p>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
                    {system.uptime && (
                      <div className="text-center">
                        <p className="font-medium">{formatUptime(system.uptime)}</p>
                        <p>Uptime</p>
                      </div>
                    )}
                    {system.responseTime && (
                      <div className="text-center">
                        <p className="font-medium">{formatResponseTime(system.responseTime)}</p>
                        <p>Response</p>
                      </div>
                    )}
                    <div className="text-center">
                      <p className={`font-medium ${statusConfig.color}`}>{statusConfig.label}</p>
                      <p>Status</p>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Status legend */}
      <div className="mt-6 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <div className="flex items-center justify-between text-xs">
          <span className="text-gray-600 dark:text-gray-400 font-medium">Status Legend:</span>
          <div className="flex items-center space-x-4">
            <div className="flex items-center">
              <div className="h-2 w-2 rounded-full bg-success-500 mr-1" />
              <span className="text-gray-600 dark:text-gray-400">Online</span>
            </div>
            <div className="flex items-center">
              <div className="h-2 w-2 rounded-full bg-warning-500 mr-1" />
              <span className="text-gray-600 dark:text-gray-400">Degraded</span>
            </div>
            <div className="flex items-center">
              <div className="h-2 w-2 rounded-full bg-danger-500 mr-1" />
              <span className="text-gray-600 dark:text-gray-400">Offline</span>
            </div>
            <div className="flex items-center">
              <div className="h-2 w-2 rounded-full bg-info-500 mr-1" />
              <span className="text-gray-600 dark:text-gray-400">Maintenance</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SystemStatusIndicators;
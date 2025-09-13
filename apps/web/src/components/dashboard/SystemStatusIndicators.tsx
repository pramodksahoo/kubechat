import React, { useEffect, useState } from 'react';
import { Icon } from '@/components/ui';

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
          color: 'text-emerald-600 dark:text-emerald-400',
          bgColor: 'bg-gradient-to-r from-emerald-500 to-green-500',
          ringColor: 'ring-emerald-500/20',
          glowColor: 'shadow-emerald-500/30',
          bgGradient: 'bg-gradient-to-br from-emerald-50 to-green-100 dark:from-emerald-900/20 dark:to-green-900/30',
          borderColor: 'border-emerald-200/50 dark:border-emerald-700/50',
          label: 'Online',
          pulseClass: 'animate-pulse',
          icon: 'check-circle'
        };
      case 'degraded':
        return {
          color: 'text-amber-600 dark:text-amber-400',
          bgColor: 'bg-gradient-to-r from-amber-500 to-orange-500',
          ringColor: 'ring-amber-500/20',
          glowColor: 'shadow-amber-500/30',
          bgGradient: 'bg-gradient-to-br from-amber-50 to-orange-100 dark:from-amber-900/20 dark:to-orange-900/30',
          borderColor: 'border-amber-200/50 dark:border-amber-700/50',
          label: 'Degraded',
          pulseClass: 'animate-pulse',
          icon: 'exclamation-triangle'
        };
      case 'offline':
        return {
          color: 'text-red-600 dark:text-red-400',
          bgColor: 'bg-gradient-to-r from-red-500 to-rose-500',
          ringColor: 'ring-red-500/20',
          glowColor: 'shadow-red-500/30',
          bgGradient: 'bg-gradient-to-br from-red-50 to-rose-100 dark:from-red-900/20 dark:to-rose-900/30',
          borderColor: 'border-red-200/50 dark:border-red-700/50',
          label: 'Offline',
          pulseClass: '',
          icon: 'x-circle'
        };
      case 'maintenance':
        return {
          color: 'text-blue-600 dark:text-blue-400',
          bgColor: 'bg-gradient-to-r from-blue-500 to-indigo-500',
          ringColor: 'ring-blue-500/20',
          glowColor: 'shadow-blue-500/30',
          bgGradient: 'bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-blue-900/20 dark:to-indigo-900/30',
          borderColor: 'border-blue-200/50 dark:border-blue-700/50',
          label: 'Maintenance',
          pulseClass: 'animate-pulse',
          icon: 'cog-6-tooth'
        };
      default:
        return {
          color: 'text-gray-600 dark:text-gray-400',
          bgColor: 'bg-gradient-to-r from-gray-500 to-slate-500',
          ringColor: 'ring-gray-500/20',
          glowColor: 'shadow-gray-500/30',
          bgGradient: 'bg-gradient-to-br from-gray-50 to-slate-100 dark:from-gray-900/20 dark:to-slate-900/30',
          borderColor: 'border-gray-200/50 dark:border-gray-700/50',
          label: 'Unknown',
          pulseClass: '',
          icon: 'question-mark-circle'
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
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center space-x-4">
          <div className="relative">
            <div className={`h-4 w-4 rounded-full ${overallConfig.bgColor} ${overallConfig.pulseClass} shadow-lg ${overallConfig.glowColor}`} />
            <div className={`absolute inset-0 h-4 w-4 rounded-full ${overallConfig.bgColor} opacity-20 ${overallConfig.pulseClass} scale-150`} />
          </div>
          <div>
            <h3 className="text-xl font-bold bg-gradient-to-r from-gray-900 via-gray-700 to-gray-900 dark:from-white dark:via-gray-200 dark:to-white bg-clip-text text-transparent">
              System Status
            </h3>
            <p className="text-sm font-medium mt-1 flex items-center space-x-2">
              <span className={overallConfig.color}>
                {overallStatus === 'critical'
                  ? 'Some services are experiencing issues'
                  : overallStatus === 'warning'
                  ? 'Some services are degraded'
                  : 'All systems operational'
                }
              </span>
              <span className="text-gray-400">•</span>
              <span className="text-gray-500 dark:text-gray-400">{displaySystems.length} services monitored</span>
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-4">
          {/* WebSocket status */}
          <div className="flex items-center px-3 py-2 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50">
            <div className={`h-2 w-2 rounded-full mr-2 ${wsStatus.isConnected ? 'bg-emerald-500 animate-pulse shadow-sm shadow-emerald-500/50' : 'bg-red-500'}`} />
            <span className="text-xs font-medium text-gray-700 dark:text-gray-300">
              {wsStatus.isConnected ? 'Live' : 'Offline'}
            </span>
          </div>

          {/* Last updated */}
          <div className="flex items-center px-3 py-2 rounded-full bg-gray-100/50 dark:bg-gray-800/50 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50">
            <Icon name="clock" className="h-3 w-3 mr-2 text-gray-500" />
            <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
              {formatTimeAgo(lastUpdated)}
            </span>
          </div>

          {/* Refresh indicator */}
          {isRefreshing && (
            <div className="flex items-center px-3 py-2 rounded-full bg-blue-100/50 dark:bg-blue-900/20 backdrop-blur-sm border border-blue-200/50 dark:border-blue-700/50">
              <Icon name="spinner" className="h-3 w-3 animate-spin text-blue-600" />
              <span className="text-xs font-medium text-blue-600 dark:text-blue-400 ml-2">Updating</span>
            </div>
          )}
        </div>
      </div>

      {/* Compact status indicators */}
      {!showDetails && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
          {displaySystems.map((system, index) => {
            const statusConfig = getStatusConfig(system.status);

            return (
              <div
                key={system.id}
                className={`group relative overflow-hidden rounded-2xl p-5 ${statusConfig.bgGradient} border ${statusConfig.borderColor}
                  backdrop-blur-sm cursor-pointer transition-all duration-300 ease-out
                  hover:scale-105 hover:shadow-xl hover:${statusConfig.glowColor}
                  transform hover:-translate-y-1`}
                onClick={() => onStatusClick?.(system.id)}
                data-testid={`system-${system.id}`}
                style={{
                  animationDelay: `${index * 100}ms`,
                  animation: 'fadeInUp 0.6s ease-out forwards'
                }}
              >
                {/* Background decoration */}
                <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                {/* Status indicator and metrics */}
                <div className="relative z-10">
                  <div className="flex items-center justify-between mb-3">
                    <div className="relative">
                      <div className={`h-3 w-3 rounded-full ${statusConfig.bgColor} ${statusConfig.pulseClass} shadow-lg`} />
                      <div className={`absolute inset-0 h-3 w-3 rounded-full ${statusConfig.bgColor} opacity-20 ${statusConfig.pulseClass} scale-150`} />
                    </div>

                    <div className="flex items-center space-x-2">
                      {system.responseTime && (
                        <div className="px-2 py-1 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                          <span className="text-xs font-medium text-gray-700 dark:text-gray-300">
                            {formatResponseTime(system.responseTime)}
                          </span>
                        </div>
                      )}
                      <Icon
                        name={statusConfig.icon}
                        className={`h-4 w-4 ${statusConfig.color} opacity-60 group-hover:opacity-100 transition-opacity duration-300`}
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <h4 className="text-sm font-semibold text-gray-900 dark:text-white group-hover:text-gray-800 dark:group-hover:text-gray-100 transition-colors duration-300">
                      {system.name}
                    </h4>

                    <div className="flex items-center justify-between">
                      <span className={`text-xs font-medium ${statusConfig.color} uppercase tracking-wider`}>
                        {statusConfig.label}
                      </span>
                      {system.uptime && (
                        <span className="text-xs text-gray-500 dark:text-gray-400 font-medium">
                          {formatUptime(system.uptime)} up
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* Hover shine effect */}
                <div className="absolute inset-0 -top-10 -left-10 bg-gradient-to-r from-transparent via-white/10 to-transparent
                  transform skew-x-12 opacity-0 group-hover:opacity-100 transition-all duration-700 group-hover:translate-x-full" />
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
      <div className="mt-8 p-6 rounded-2xl bg-gradient-to-br from-gray-50/50 to-white/50 dark:from-gray-900/50 dark:to-gray-800/50 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between space-y-3 sm:space-y-0">
          <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">Status Legend</span>
          <div className="flex flex-wrap items-center gap-4">
            {[
              { status: 'online', color: 'bg-gradient-to-r from-emerald-500 to-green-500', label: 'Online' },
              { status: 'degraded', color: 'bg-gradient-to-r from-amber-500 to-orange-500', label: 'Degraded' },
              { status: 'offline', color: 'bg-gradient-to-r from-red-500 to-rose-500', label: 'Offline' },
              { status: 'maintenance', color: 'bg-gradient-to-r from-blue-500 to-indigo-500', label: 'Maintenance' }
            ].map(({ status, color, label }) => {
              return (
                <div key={status} className="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                  <div className="relative">
                    <div className={`h-2.5 w-2.5 rounded-full ${color} shadow-sm`} />
                    <div className={`absolute inset-0 h-2.5 w-2.5 rounded-full ${color} opacity-30 animate-pulse scale-150`} />
                  </div>
                  <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{label}</span>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Add keyframes for animations */}
      <style jsx>{`
        @keyframes fadeInUp {
          0% {
            opacity: 0;
            transform: translateY(20px);
          }
          100% {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
};

export default SystemStatusIndicators;
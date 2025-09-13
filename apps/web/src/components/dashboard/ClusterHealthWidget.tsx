import React from 'react';
import { Icon } from '@/components/ui';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

interface ClusterHealthData {
  id: string;
  name: string;
  status: 'healthy' | 'warning' | 'critical' | 'unknown';
  uptime: string;
  nodes: {
    total: number;
    ready: number;
    notReady: number;
  };
  pods: {
    total: number;
    running: number;
    pending: number;
    failed: number;
  };
  resources: {
    cpu: {
      used: number;
      total: number;
      percentage: number;
    };
    memory: {
      used: number;
      total: number;
      percentage: number;
    };
  };
  lastChecked: string;
}

export interface ClusterHealthWidgetProps extends BaseComponentProps {
  clusters?: ClusterHealthData[];
  isLoading?: boolean;
  onRefresh?: () => void;
  onClusterClick?: (clusterId: string) => void;
}

export const ClusterHealthWidget: React.FC<ClusterHealthWidgetProps> = ({
  clusters = [],
  isLoading = false,
  onRefresh,
  onClusterClick,
  className = '',
  'data-testid': dataTestId = 'cluster-health-widget'
}) => {
  const defaultClusters: ClusterHealthData[] = [
    {
      id: 'prod-cluster-1',
      name: 'Production Cluster',
      status: 'healthy',
      uptime: '45 days',
      nodes: { total: 5, ready: 5, notReady: 0 },
      pods: { total: 127, running: 125, pending: 2, failed: 0 },
      resources: {
        cpu: { used: 2.4, total: 8.0, percentage: 30 },
        memory: { used: 14.2, total: 32.0, percentage: 44 }
      },
      lastChecked: new Date().toISOString()
    },
    {
      id: 'staging-cluster-1',
      name: 'Staging Cluster',
      status: 'warning',
      uptime: '12 days',
      nodes: { total: 3, ready: 2, notReady: 1 },
      pods: { total: 45, running: 42, pending: 2, failed: 1 },
      resources: {
        cpu: { used: 1.8, total: 4.0, percentage: 45 },
        memory: { used: 6.8, total: 16.0, percentage: 42 }
      },
      lastChecked: new Date().toISOString()
    }
  ];

  const displayClusters = clusters.length > 0 ? clusters : defaultClusters;

  const getStatusConfig = (status: string) => {
    switch (status) {
      case 'healthy':
        return {
          icon: 'check-circle',
          color: 'text-emerald-600 dark:text-emerald-400',
          bgGradient: 'bg-gradient-to-br from-emerald-50 to-green-100 dark:from-emerald-900/20 dark:to-green-900/30',
          borderColor: 'border-emerald-200/50 dark:border-emerald-700/50',
          badgeClasses: 'bg-gradient-to-r from-emerald-100 to-green-200 text-emerald-800 dark:from-emerald-900/50 dark:to-green-900/50 dark:text-emerald-200',
          dotColor: 'bg-gradient-to-r from-emerald-500 to-green-500',
          glowColor: 'shadow-emerald-500/20'
        };
      case 'warning':
        return {
          icon: 'exclamation-triangle',
          color: 'text-amber-600 dark:text-amber-400',
          bgGradient: 'bg-gradient-to-br from-amber-50 to-orange-100 dark:from-amber-900/20 dark:to-orange-900/30',
          borderColor: 'border-amber-200/50 dark:border-amber-700/50',
          badgeClasses: 'bg-gradient-to-r from-amber-100 to-orange-200 text-amber-800 dark:from-amber-900/50 dark:to-orange-900/50 dark:text-amber-200',
          dotColor: 'bg-gradient-to-r from-amber-500 to-orange-500',
          glowColor: 'shadow-amber-500/20'
        };
      case 'critical':
        return {
          icon: 'x-circle',
          color: 'text-red-600 dark:text-red-400',
          bgGradient: 'bg-gradient-to-br from-red-50 to-rose-100 dark:from-red-900/20 dark:to-rose-900/30',
          borderColor: 'border-red-200/50 dark:border-red-700/50',
          badgeClasses: 'bg-gradient-to-r from-red-100 to-rose-200 text-red-800 dark:from-red-900/50 dark:to-rose-900/50 dark:text-red-200',
          dotColor: 'bg-gradient-to-r from-red-500 to-rose-500',
          glowColor: 'shadow-red-500/20'
        };
      default:
        return {
          icon: 'question-mark-circle',
          color: 'text-gray-600 dark:text-gray-400',
          bgGradient: 'bg-gradient-to-br from-gray-50 to-slate-100 dark:from-gray-900/20 dark:to-slate-900/30',
          borderColor: 'border-gray-200/50 dark:border-gray-700/50',
          badgeClasses: 'bg-gradient-to-r from-gray-100 to-slate-200 text-gray-800 dark:from-gray-900/50 dark:to-slate-900/50 dark:text-gray-200',
          dotColor: 'bg-gradient-to-r from-gray-500 to-slate-500',
          glowColor: 'shadow-gray-500/20'
        };
    }
  };

  const getResourceBarColor = (percentage: number, type: 'cpu' | 'memory') => {
    if (percentage > 80) {
      return type === 'cpu'
        ? 'bg-gradient-to-r from-red-500 to-rose-500'
        : 'bg-gradient-to-r from-red-500 to-pink-500';
    }
    if (percentage > 60) {
      return type === 'cpu'
        ? 'bg-gradient-to-r from-amber-500 to-orange-500'
        : 'bg-gradient-to-r from-amber-500 to-yellow-500';
    }
    return type === 'cpu'
      ? 'bg-gradient-to-r from-blue-500 to-indigo-500'
      : 'bg-gradient-to-r from-emerald-500 to-green-500';
  };

  // Animated progress bar component
  const ProgressBar: React.FC<{ percentage: number; type: 'cpu' | 'memory'; animated?: boolean }> = ({ percentage, type, animated = true }) => {
    const [animatedPercentage, setAnimatedPercentage] = React.useState(0);

    React.useEffect(() => {
      if (animated) {
        const timer = setTimeout(() => {
          setAnimatedPercentage(percentage);
        }, 300);
        return () => clearTimeout(timer);
      } else {
        setAnimatedPercentage(percentage);
      }
    }, [percentage, animated]);

    return (
      <div className="relative w-full bg-gray-200/50 dark:bg-gray-700/50 rounded-full h-2 overflow-hidden backdrop-blur-sm">
        <div
          className={`h-full transition-all duration-1000 ease-out ${getResourceBarColor(percentage, type)} shadow-sm`}
          style={{ width: `${animatedPercentage}%` }}
        />
        <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent opacity-0 hover:opacity-100 transition-opacity duration-300" />
      </div>
    );
  };

  return (
    <div className={`relative overflow-hidden rounded-2xl bg-white/70 dark:bg-gray-900/70 backdrop-blur-xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl ${className}`} data-testid={dataTestId}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-xl bg-gradient-to-br from-blue-100 to-indigo-200 dark:from-blue-800/50 dark:to-indigo-700/50">
              <Icon name="server" className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h3 className="text-xl font-bold bg-gradient-to-r from-gray-900 via-gray-700 to-gray-900 dark:from-white dark:via-gray-200 dark:to-white bg-clip-text text-transparent">
                Cluster Health
              </h3>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
                {displayClusters.length} {displayClusters.length === 1 ? 'cluster' : 'clusters'} monitored
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onRefresh}
            disabled={isLoading}
            className="p-3 rounded-xl text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-blue-500/20 disabled:opacity-50 transition-all duration-200"
            data-testid="refresh-button"
            aria-label="Refresh cluster health"
          >
            <Icon
              name={isLoading ? "spinner" : "refresh"}
              className={`h-5 w-5 ${isLoading ? 'animate-spin' : ''}`}
            />
          </button>
        </div>

        {/* Loading state */}
        {isLoading && (
          <div className="flex items-center justify-center py-8">
            <Icon name="spinner" className="h-8 w-8 animate-spin text-primary-500" />
            <span className="ml-3 text-sm text-gray-500">Loading cluster data...</span>
          </div>
        )}

        {/* Cluster list */}
        {!isLoading && (
          <div className="space-y-6">
            {displayClusters.map((cluster, index) => {
              const statusConfig = getStatusConfig(cluster.status);

              return (
                <div
                  key={cluster.id}
                  className={`group relative overflow-hidden rounded-2xl p-6 ${statusConfig.bgGradient} border ${statusConfig.borderColor}
                    backdrop-blur-sm cursor-pointer transition-all duration-300 ease-out
                    hover:scale-[1.02] hover:shadow-xl hover:${statusConfig.glowColor}
                    transform hover:-translate-y-1`}
                  onClick={() => onClusterClick?.(cluster.id)}
                  data-testid={`cluster-${cluster.id}`}
                  style={{
                    animationDelay: `${index * 200}ms`,
                    animation: 'fadeInUp 0.6s ease-out forwards'
                  }}
                >
                  {/* Background decoration */}
                  <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                  {/* Cluster header */}
                  <div className="relative z-10">
                    <div className="flex items-center justify-between mb-6">
                      <div className="flex items-center space-x-3">
                        <div className="relative">
                          <div className={`h-3 w-3 rounded-full ${statusConfig.dotColor} shadow-lg animate-pulse`} />
                          <div className={`absolute inset-0 h-3 w-3 rounded-full ${statusConfig.dotColor} opacity-20 animate-pulse scale-150`} />
                        </div>
                        <div>
                          <h4 className="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-gray-800 dark:group-hover:text-gray-100 transition-colors duration-300">
                            {cluster.name}
                          </h4>
                          <p className="text-sm text-gray-600 dark:text-gray-400">Uptime: {cluster.uptime}</p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-3">
                        <span className={`inline-flex items-center px-3 py-1.5 rounded-full text-xs font-semibold ${statusConfig.badgeClasses} shadow-sm uppercase tracking-wider`}>
                          <Icon name={statusConfig.icon} className="h-3 w-3 mr-1.5" />
                          {cluster.status}
                        </span>
                      </div>
                    </div>

                    {/* Cluster metrics */}
                    <div className="grid grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
                      <div className="text-center p-3 rounded-xl bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                        <Icon name="server" className="h-5 w-5 text-blue-600 dark:text-blue-400 mx-auto mb-2" />
                        <p className="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wider">Nodes</p>
                        <p className="text-lg font-bold text-gray-900 dark:text-white mt-1">
                          {cluster.nodes.ready}<span className="text-gray-500 dark:text-gray-400">/{cluster.nodes.total}</span>
                        </p>
                      </div>
                      <div className="text-center p-3 rounded-xl bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                        <Icon name="cube" className="h-5 w-5 text-emerald-600 dark:text-emerald-400 mx-auto mb-2" />
                        <p className="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wider">Pods</p>
                        <p className="text-lg font-bold text-gray-900 dark:text-white mt-1">
                          {cluster.pods.running}<span className="text-gray-500 dark:text-gray-400">/{cluster.pods.total}</span>
                        </p>
                      </div>
                      <div className="text-center p-3 rounded-xl bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                        <Icon name="cpu-chip" className="h-5 w-5 text-purple-600 dark:text-purple-400 mx-auto mb-2" />
                        <p className="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wider">CPU</p>
                        <p className="text-lg font-bold text-gray-900 dark:text-white mt-1">
                          {cluster.resources.cpu.percentage}%
                        </p>
                      </div>
                      <div className="text-center p-3 rounded-xl bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                        <Icon name="circle-stack" className="h-5 w-5 text-orange-600 dark:text-orange-400 mx-auto mb-2" />
                        <p className="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase tracking-wider">Memory</p>
                        <p className="text-lg font-bold text-gray-900 dark:text-white mt-1">
                          {cluster.resources.memory.percentage}%
                        </p>
                      </div>
                    </div>

                    {/* Resource usage bars */}
                    <div className="space-y-4">
                      <div className="p-4 rounded-xl bg-white/30 dark:bg-gray-800/30 backdrop-blur-sm">
                        <div className="flex justify-between items-center mb-3">
                          <div className="flex items-center space-x-2">
                            <Icon name="cpu-chip" className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">CPU Usage</span>
                          </div>
                          <span className="text-sm font-semibold text-gray-900 dark:text-white">
                            {cluster.resources.cpu.used}GB / {cluster.resources.cpu.total}GB
                          </span>
                        </div>
                        <ProgressBar percentage={cluster.resources.cpu.percentage} type="cpu" />
                        <div className="flex justify-between items-center mt-2">
                          <span className="text-xs text-gray-500 dark:text-gray-400">0%</span>
                          <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{cluster.resources.cpu.percentage}%</span>
                          <span className="text-xs text-gray-500 dark:text-gray-400">100%</span>
                        </div>
                      </div>

                      <div className="p-4 rounded-xl bg-white/30 dark:bg-gray-800/30 backdrop-blur-sm">
                        <div className="flex justify-between items-center mb-3">
                          <div className="flex items-center space-x-2">
                            <Icon name="circle-stack" className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Memory Usage</span>
                          </div>
                          <span className="text-sm font-semibold text-gray-900 dark:text-white">
                            {cluster.resources.memory.used}GB / {cluster.resources.memory.total}GB
                          </span>
                        </div>
                        <ProgressBar percentage={cluster.resources.memory.percentage} type="memory" />
                        <div className="flex justify-between items-center mt-2">
                          <span className="text-xs text-gray-500 dark:text-gray-400">0%</span>
                          <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{cluster.resources.memory.percentage}%</span>
                          <span className="text-xs text-gray-500 dark:text-gray-400">100%</span>
                        </div>
                      </div>
                    </div>

                    {/* Footer */}
                    <div className="flex justify-between items-center mt-6 pt-4 border-t border-gray-200/50 dark:border-gray-700/50">
                      <div className="flex items-center space-x-4 text-xs">
                        <div className="flex items-center space-x-1">
                          <div className="h-1.5 w-1.5 bg-emerald-500 rounded-full animate-pulse" />
                          <span className="text-gray-600 dark:text-gray-400">Last checked</span>
                        </div>
                        <span className="font-medium text-gray-700 dark:text-gray-300">
                          {new Date(cluster.lastChecked).toLocaleTimeString()}
                        </span>
                      </div>
                      <button className="text-xs font-medium text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors duration-200">
                        View Details →
                      </button>
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

        {/* Empty state */}
        {!isLoading && displayClusters.length === 0 && (
          <div className="text-center py-12">
            <div className="p-6 rounded-2xl bg-gradient-to-br from-gray-50 to-slate-100 dark:from-gray-800/50 dark:to-slate-900/50 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50 max-w-md mx-auto">
              <Icon name="server" className="h-16 w-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
              <p className="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">No clusters connected</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">Connect your first Kubernetes cluster to get started with monitoring</p>
              <button className="mt-4 px-4 py-2 text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors duration-200">
                Connect Cluster \u2192
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Add keyframes for animations */}
      <style jsx>{`
        @keyframes fadeInUp {
          0% {
            opacity: 0;
            transform: translateY(30px);
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

export default ClusterHealthWidget;
import React from 'react';
import { Icon } from '@/components/ui/Icon';
import { Card } from '@/components/ui/Card';

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

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return { icon: 'check-circle', color: 'text-success-500' };
      case 'warning':
        return { icon: 'exclamation-triangle', color: 'text-warning-500' };
      case 'critical':
        return { icon: 'x-circle', color: 'text-danger-500' };
      default:
        return { icon: 'question-mark-circle', color: 'text-gray-500' };
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'bg-success-100 text-success-800 dark:bg-success-900 dark:text-success-200';
      case 'warning':
        return 'bg-warning-100 text-warning-800 dark:bg-warning-900 dark:text-warning-200';
      case 'critical':
        return 'bg-danger-100 text-danger-800 dark:bg-danger-900 dark:text-danger-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300';
    }
  };

  return (
    <Card className={className} data-testid={dataTestId}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center">
            <Icon name="server" className="h-5 w-5 text-primary-500 mr-2" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Cluster Health
            </h3>
          </div>
          <button
            type="button"
            onClick={onRefresh}
            disabled={isLoading}
            className="p-2 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 rounded-md disabled:opacity-50"
            data-testid="refresh-button"
            aria-label="Refresh cluster health"
          >
            <Icon 
              name="refresh" 
              className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} 
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
          <div className="space-y-4">
            {displayClusters.map((cluster) => {
              const statusConfig = getStatusIcon(cluster.status);
              
              return (
                <div
                  key={cluster.id}
                  className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:border-primary-300 dark:hover:border-primary-600 transition-colors cursor-pointer"
                  onClick={() => onClusterClick?.(cluster.id)}
                  data-testid={`cluster-${cluster.id}`}
                >
                  {/* Cluster header */}
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center">
                      <Icon 
                        name={statusConfig.icon} 
                        className={`h-5 w-5 ${statusConfig.color} mr-2`} 
                      />
                      <h4 className="text-sm font-medium text-gray-900 dark:text-white">
                        {cluster.name}
                      </h4>
                    </div>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusBadge(cluster.status)}`}>
                      {cluster.status}
                    </span>
                  </div>

                  {/* Cluster metrics */}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
                    <div>
                      <p className="text-gray-500 dark:text-gray-400">Nodes</p>
                      <p className="font-medium text-gray-900 dark:text-white">
                        {cluster.nodes.ready}/{cluster.nodes.total}
                      </p>
                    </div>
                    <div>
                      <p className="text-gray-500 dark:text-gray-400">Pods</p>
                      <p className="font-medium text-gray-900 dark:text-white">
                        {cluster.pods.running}/{cluster.pods.total}
                      </p>
                    </div>
                    <div>
                      <p className="text-gray-500 dark:text-gray-400">CPU</p>
                      <p className="font-medium text-gray-900 dark:text-white">
                        {cluster.resources.cpu.percentage}%
                      </p>
                    </div>
                    <div>
                      <p className="text-gray-500 dark:text-gray-400">Memory</p>
                      <p className="font-medium text-gray-900 dark:text-white">
                        {cluster.resources.memory.percentage}%
                      </p>
                    </div>
                  </div>

                  {/* Resource usage bars */}
                  <div className="mt-3 space-y-2">
                    <div>
                      <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
                        <span>CPU Usage</span>
                        <span>{cluster.resources.cpu.used}GB / {cluster.resources.cpu.total}GB</span>
                      </div>
                      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                        <div 
                          className="bg-primary-500 h-1.5 rounded-full transition-all duration-300" 
                          style={{ width: `${cluster.resources.cpu.percentage}%` }}
                        />
                      </div>
                    </div>
                    <div>
                      <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
                        <span>Memory Usage</span>
                        <span>{cluster.resources.memory.used}GB / {cluster.resources.memory.total}GB</span>
                      </div>
                      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                        <div 
                          className="bg-secondary-500 h-1.5 rounded-full transition-all duration-300" 
                          style={{ width: `${cluster.resources.memory.percentage}%` }}
                        />
                      </div>
                    </div>
                  </div>

                  {/* Footer */}
                  <div className="flex justify-between items-center mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                      Uptime: {cluster.uptime}
                    </span>
                    <span className="text-xs text-gray-500 dark:text-gray-400">
                      Updated: {new Date(cluster.lastChecked).toLocaleTimeString()}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {/* Empty state */}
        {!isLoading && displayClusters.length === 0 && (
          <div className="text-center py-8">
            <Icon name="server" className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <p className="text-sm text-gray-500">No clusters connected</p>
            <p className="text-xs text-gray-400 mt-1">Connect your first Kubernetes cluster to get started</p>
          </div>
        )}
      </div>
    </Card>
  );
};

export default ClusterHealthWidget;
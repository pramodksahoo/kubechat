import React, { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { ClusterHealthWidget } from './ClusterHealthWidget';
import { RecentActivitiesFeed } from './RecentActivitiesFeed';
import { QuickAccessPanels } from './QuickAccessPanels';
import { SystemStatusIndicators } from './SystemStatusIndicators';
import { PerformanceMonitoringWidget } from './PerformanceMonitoringWidget';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

interface DashboardStats {
  totalClusters: number;
  activePods: number;
  commandsToday: number;
  systemHealth: 'healthy' | 'warning' | 'critical';
}

export interface DashboardViewProps extends BaseComponentProps {
  stats?: DashboardStats;
  userName?: string;
  onRefreshAll?: () => void;
  isLoading?: boolean;
}

export const DashboardView: React.FC<DashboardViewProps> = ({
  stats,
  userName = 'Administrator',
  onRefreshAll,
  isLoading = false,
  className = '',
  'data-testid': dataTestId = 'dashboard-view'
}) => {
  const [refreshingWidget, setRefreshingWidget] = useState<string | null>(null);

  const defaultStats: DashboardStats = {
    totalClusters: 3,
    activePods: 247,
    commandsToday: 18,
    systemHealth: 'healthy'
  };

  const displayStats = stats || defaultStats;

  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return 'Good morning';
    if (hour < 18) return 'Good afternoon';
    return 'Good evening';
  };

  const getHealthStatusConfig = (health: string) => {
    switch (health) {
      case 'healthy':
        return { icon: 'check-circle', color: 'text-success-500', bg: 'bg-success-50 dark:bg-success-900/20' };
      case 'warning':
        return { icon: 'exclamation-triangle', color: 'text-warning-500', bg: 'bg-warning-50 dark:bg-warning-900/20' };
      case 'critical':
        return { icon: 'x-circle', color: 'text-danger-500', bg: 'bg-danger-50 dark:bg-danger-900/20' };
      default:
        return { icon: 'question-mark-circle', color: 'text-gray-500', bg: 'bg-gray-50 dark:bg-gray-800' };
    }
  };

  const healthConfig = getHealthStatusConfig(displayStats.systemHealth);

  const handleWidgetRefresh = async (widgetId: string, refreshFn?: () => void) => {
    setRefreshingWidget(widgetId);
    try {
      if (refreshFn) {
        await refreshFn();
      } else {
        // Simulate refresh
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    } finally {
      setRefreshingWidget(null);
    }
  };

  const statCards = [
    {
      title: 'Total Clusters',
      value: displayStats.totalClusters,
      icon: 'server',
      color: 'primary',
      change: '+2 this month'
    },
    {
      title: 'Active Pods',
      value: displayStats.activePods,
      icon: 'cube',
      color: 'success',
      change: '+12 since yesterday'
    },
    {
      title: 'Commands Today',
      value: displayStats.commandsToday,
      icon: 'terminal',
      color: 'info',
      change: '↑ 25% vs yesterday'
    },
    {
      title: 'System Health',
      value: displayStats.systemHealth.charAt(0).toUpperCase() + displayStats.systemHealth.slice(1),
      icon: healthConfig.icon,
      color: displayStats.systemHealth === 'healthy' ? 'success' : displayStats.systemHealth === 'warning' ? 'warning' : 'danger',
      change: 'All systems operational'
    }
  ];

  const getStatCardClasses = (color: string) => {
    const colorMap = {
      primary: 'border-primary-200 dark:border-primary-800 bg-primary-50 dark:bg-primary-900/20',
      success: 'border-success-200 dark:border-success-800 bg-success-50 dark:bg-success-900/20',
      warning: 'border-warning-200 dark:border-warning-800 bg-warning-50 dark:bg-warning-900/20',
      danger: 'border-danger-200 dark:border-danger-800 bg-danger-50 dark:bg-danger-900/20',
      info: 'border-info-200 dark:border-info-800 bg-info-50 dark:bg-info-900/20'
    };
    return colorMap[color as keyof typeof colorMap] || colorMap.primary;
  };

  const getIconColor = (color: string) => {
    const colorMap = {
      primary: 'text-primary-600 dark:text-primary-400',
      success: 'text-success-600 dark:text-success-400',
      warning: 'text-warning-600 dark:text-warning-400',
      danger: 'text-danger-600 dark:text-danger-400',
      info: 'text-info-600 dark:text-info-400'
    };
    return colorMap[color as keyof typeof colorMap] || colorMap.primary;
  };

  return (
    <div className={`space-y-6 ${className}`} data-testid={dataTestId}>
      {/* Welcome header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {getGreeting()}, {userName}!
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Here&apos;s what&apos;s happening with your Kubernetes clusters today
          </p>
        </div>
        
        <div className="flex items-center space-x-3">
          <button
            type="button"
            onClick={onRefreshAll}
            disabled={isLoading}
            className="flex items-center px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
            data-testid="refresh-all-button"
          >
            <Icon name={isLoading ? "spinner" : "refresh"} className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh All
          </button>
          
          <button
            type="button"
            className="flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
            data-testid="new-chat-button"
          >
            <Icon name="chat-bubble-left-right" className="h-4 w-4 mr-2" />
            New Chat
          </button>
        </div>
      </div>

      {/* Key metrics stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat, index) => (
          <div
            key={index}
            className={`p-6 rounded-lg border ${getStatCardClasses(stat.color)} transition-all duration-200 hover:shadow-md`}
            data-testid={`stat-card-${index}`}
          >
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Icon name={stat.icon} className={`h-8 w-8 ${getIconColor(stat.color)}`} />
              </div>
              <div className="ml-4 flex-1">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
                  {stat.title}
                </p>
                <p className="text-2xl font-semibold text-gray-900 dark:text-white">
                  {typeof stat.value === 'number' ? stat.value.toLocaleString() : stat.value}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  {stat.change}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* System status indicators */}
      <SystemStatusIndicators
        showDetails={false}
        onStatusClick={(systemId) => console.log('Status clicked:', systemId)}
      />

      {/* Main dashboard grid */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        {/* Left column - Cluster health and performance */}
        <div className="xl:col-span-2 space-y-6">
          <ClusterHealthWidget
            isLoading={refreshingWidget === 'cluster-health'}
            onRefresh={() => handleWidgetRefresh('cluster-health')}
            onClusterClick={(clusterId) => console.log('Cluster clicked:', clusterId)}
          />
          
          <PerformanceMonitoringWidget
            autoRefresh={true}
            onTimeRangeChange={(range) => console.log('Time range changed:', range)}
            onMetricClick={(metricId) => console.log('Metric clicked:', metricId)}
          />
        </div>

        {/* Right column - Activities */}
        <div className="space-y-6">
          <RecentActivitiesFeed
            maxItems={8}
            isLoading={refreshingWidget === 'activities'}
            onRefresh={() => handleWidgetRefresh('activities')}
            onActivityClick={(activityId) => console.log('Activity clicked:', activityId)}
            onViewAll={() => console.log('View all activities')}
          />
        </div>
      </div>

      {/* Quick access panels */}
      <QuickAccessPanels
        columns={3}
        onItemClick={(itemId) => console.log('Quick access clicked:', itemId)}
      />

      {/* Footer information */}
      <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            <Icon name="information-circle" className="h-5 w-5 text-info-600 dark:text-info-400 mr-2" />
            <div>
              <p className="text-sm font-medium text-gray-900 dark:text-white">
                Dashboard Information
              </p>
              <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                Data refreshes automatically every 30 seconds. Last updated: {new Date().toLocaleTimeString()}
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
            <div className="flex items-center">
              <div className="h-2 w-2 bg-success-500 rounded-full animate-pulse-subtle mr-2" />
              <span>Real-time updates active</span>
            </div>
            <div className="flex items-center">
              <Icon name="shield-check" className="h-4 w-4 mr-1" />
              <span>Secure connection</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardView;
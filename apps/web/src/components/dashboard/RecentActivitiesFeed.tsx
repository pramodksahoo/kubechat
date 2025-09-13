import React from 'react';
import { Icon } from '@/components/ui/Icon';
import { Card } from '@/components/ui/Card';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

interface ActivityData {
  id: string;
  type: 'command' | 'deployment' | 'security' | 'audit' | 'system';
  title: string;
  description: string;
  user: {
    id: string;
    name: string;
    avatar?: string;
  };
  timestamp: string;
  status: 'success' | 'warning' | 'error' | 'pending';
  cluster?: string;
  metadata?: Record<string, unknown>;
}

export interface RecentActivitiesFeedProps extends BaseComponentProps {
  activities?: ActivityData[];
  isLoading?: boolean;
  maxItems?: number;
  onRefresh?: () => void;
  onActivityClick?: (activityId: string) => void;
  onViewAll?: () => void;
}

export const RecentActivitiesFeed: React.FC<RecentActivitiesFeedProps> = ({
  activities = [],
  isLoading = false,
  maxItems = 10,
  onRefresh,
  onActivityClick,
  onViewAll,
  className = '',
  'data-testid': dataTestId = 'recent-activities-feed'
}) => {
  const defaultActivities: ActivityData[] = [
    {
      id: 'activity-1',
      type: 'command',
      title: 'kubectl get pods executed',
      description: 'Listed all pods in production namespace',
      user: { id: 'user-1', name: 'John Doe' },
      timestamp: new Date(Date.now() - 300000).toISOString(), // 5 minutes ago
      status: 'success',
      cluster: 'prod-cluster-1'
    },
    {
      id: 'activity-2',
      type: 'deployment',
      title: 'Application deployment completed',
      description: 'Successfully deployed webapp v2.1.0 to staging',
      user: { id: 'user-2', name: 'Jane Smith' },
      timestamp: new Date(Date.now() - 900000).toISOString(), // 15 minutes ago
      status: 'success',
      cluster: 'staging-cluster-1'
    },
    {
      id: 'activity-3',
      type: 'security',
      title: 'Failed authentication attempt',
      description: 'Multiple failed login attempts from IP 192.168.1.100',
      user: { id: 'system', name: 'System' },
      timestamp: new Date(Date.now() - 1200000).toISOString(), // 20 minutes ago
      status: 'warning'
    },
    {
      id: 'activity-4',
      type: 'audit',
      title: 'Compliance scan completed',
      description: 'Security compliance audit completed with 2 warnings',
      user: { id: 'user-3', name: 'Admin User' },
      timestamp: new Date(Date.now() - 1800000).toISOString(), // 30 minutes ago
      status: 'warning',
      cluster: 'prod-cluster-1'
    },
    {
      id: 'activity-5',
      type: 'command',
      title: 'kubectl scale deployment',
      description: 'Scaled nginx deployment to 3 replicas',
      user: { id: 'user-1', name: 'John Doe' },
      timestamp: new Date(Date.now() - 2700000).toISOString(), // 45 minutes ago
      status: 'success',
      cluster: 'prod-cluster-1'
    }
  ];

  const displayActivities = (activities.length > 0 ? activities : defaultActivities)
    .slice(0, maxItems);

  const getActivityIcon = (type: string) => {
    const iconMap: Record<string, string> = {
      command: 'terminal',
      deployment: 'rocket-launch',
      security: 'shield-check',
      audit: 'document-check',
      system: 'cog'
    };
    
    return iconMap[type] || 'bell';
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'text-success-500';
      case 'warning':
        return 'text-warning-500';
      case 'error':
        return 'text-danger-500';
      case 'pending':
        return 'text-info-500';
      default:
        return 'text-gray-500';
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-success-100 text-success-800 dark:bg-success-900 dark:text-success-200';
      case 'warning':
        return 'bg-warning-100 text-warning-800 dark:bg-warning-900 dark:text-warning-200';
      case 'error':
        return 'bg-danger-100 text-danger-800 dark:bg-danger-900 dark:text-danger-200';
      case 'pending':
        return 'bg-info-100 text-info-800 dark:bg-info-900 dark:text-info-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300';
    }
  };

  const formatTimeAgo = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diffInMinutes = Math.floor((now.getTime() - time.getTime()) / 60000);
    
    if (diffInMinutes < 1) return 'Just now';
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
    
    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) return `${diffInHours}h ago`;
    
    const diffInDays = Math.floor(diffInHours / 24);
    return `${diffInDays}d ago`;
  };

  return (
    <Card className={className} data-testid={dataTestId}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center">
            <Icon name="clock" className="h-5 w-5 text-primary-500 mr-2" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Recent Activities
            </h3>
          </div>
          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={onRefresh}
              disabled={isLoading}
              className="p-2 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 rounded-md disabled:opacity-50"
              data-testid="refresh-button"
              aria-label="Refresh activities"
            >
              <Icon 
                name="refresh" 
                className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} 
              />
            </button>
            <button
              type="button"
              onClick={onViewAll}
              className="text-sm text-primary-600 dark:text-primary-400 hover:text-primary-500 font-medium"
              data-testid="view-all-button"
            >
              View all
            </button>
          </div>
        </div>

        {/* Loading state */}
        {isLoading && (
          <div className="flex items-center justify-center py-8">
            <Icon name="spinner" className="h-8 w-8 animate-spin text-primary-500" />
            <span className="ml-3 text-sm text-gray-500">Loading activities...</span>
          </div>
        )}

        {/* Activities list */}
        {!isLoading && (
          <div className="space-y-4">
            {displayActivities.map((activity, index) => (
              <div
                key={activity.id}
                className="flex items-start space-x-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer transition-colors"
                onClick={() => onActivityClick?.(activity.id)}
                data-testid={`activity-${activity.id}`}
              >
                {/* Activity icon */}
                <div className={`flex-shrink-0 w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center ${getStatusColor(activity.status)}`}>
                  <Icon 
                    name={getActivityIcon(activity.type)} 
                    className="h-4 w-4" 
                  />
                </div>

                {/* Activity content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                      {activity.title}
                    </p>
                    <div className="flex items-center space-x-2 ml-2">
                      <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusBadge(activity.status)}`}>
                        {activity.status}
                      </span>
                      <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                        {formatTimeAgo(activity.timestamp)}
                      </span>
                    </div>
                  </div>
                  
                  <p className="text-sm text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">
                    {activity.description}
                  </p>
                  
                  <div className="flex items-center justify-between mt-2">
                    <div className="flex items-center text-xs text-gray-500 dark:text-gray-400">
                      <Icon name="user" className="h-3 w-3 mr-1" />
                      <span>{activity.user.name}</span>
                      {activity.cluster && (
                        <>
                          <span className="mx-1">•</span>
                          <Icon name="server" className="h-3 w-3 mr-1" />
                          <span>{activity.cluster}</span>
                        </>
                      )}
                    </div>
                  </div>
                </div>

                {/* Timeline line */}
                {index < displayActivities.length - 1 && (
                  <div className="absolute left-6 w-px h-8 bg-gray-200 dark:bg-gray-700 mt-10" />
                )}
              </div>
            ))}
          </div>
        )}

        {/* Empty state */}
        {!isLoading && displayActivities.length === 0 && (
          <div className="text-center py-8">
            <Icon name="clock" className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <p className="text-sm text-gray-500">No recent activities</p>
            <p className="text-xs text-gray-400 mt-1">Activities will appear here as you interact with your clusters</p>
          </div>
        )}

        {/* Footer with activity count */}
        {!isLoading && displayActivities.length > 0 && (
          <div className="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
              <span>Showing {displayActivities.length} recent activities</span>
              <button
                type="button"
                onClick={onViewAll}
                className="text-primary-600 dark:text-primary-400 hover:text-primary-500 font-medium"
              >
                View activity log →
              </button>
            </div>
          </div>
        )}
      </div>
    </Card>
  );
};

export default RecentActivitiesFeed;
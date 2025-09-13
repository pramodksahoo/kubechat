import React from 'react';
import { Icon } from '@/components/ui';

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

  const getActivityConfig = (type: string, status: string) => {
    const typeConfig = {
      command: {
        icon: 'terminal',
        bgColor: 'bg-gradient-to-br from-blue-100 to-indigo-200 dark:from-blue-800/50 dark:to-indigo-700/50',
        iconColor: 'text-blue-600 dark:text-blue-400',
        borderColor: 'border-blue-200/50 dark:border-blue-700/50'
      },
      deployment: {
        icon: 'rocket-launch',
        bgColor: 'bg-gradient-to-br from-purple-100 to-violet-200 dark:from-purple-800/50 dark:to-violet-700/50',
        iconColor: 'text-purple-600 dark:text-purple-400',
        borderColor: 'border-purple-200/50 dark:border-purple-700/50'
      },
      security: {
        icon: 'shield-check',
        bgColor: 'bg-gradient-to-br from-emerald-100 to-green-200 dark:from-emerald-800/50 dark:to-green-700/50',
        iconColor: 'text-emerald-600 dark:text-emerald-400',
        borderColor: 'border-emerald-200/50 dark:border-emerald-700/50'
      },
      audit: {
        icon: 'document-check',
        bgColor: 'bg-gradient-to-br from-cyan-100 to-blue-200 dark:from-cyan-800/50 dark:to-blue-700/50',
        iconColor: 'text-cyan-600 dark:text-cyan-400',
        borderColor: 'border-cyan-200/50 dark:border-cyan-700/50'
      },
      system: {
        icon: 'cog',
        bgColor: 'bg-gradient-to-br from-gray-100 to-slate-200 dark:from-gray-800/50 dark:to-slate-700/50',
        iconColor: 'text-gray-600 dark:text-gray-400',
        borderColor: 'border-gray-200/50 dark:border-gray-700/50'
      }
    };

    const statusConfig = {
      success: {
        dotColor: 'bg-gradient-to-r from-emerald-500 to-green-500',
        badgeClasses: 'bg-gradient-to-r from-emerald-100 to-green-200 text-emerald-800 dark:from-emerald-900/50 dark:to-green-900/50 dark:text-emerald-200',
        timelineColor: 'bg-gradient-to-b from-emerald-500/20 to-emerald-500/5'
      },
      warning: {
        dotColor: 'bg-gradient-to-r from-amber-500 to-orange-500',
        badgeClasses: 'bg-gradient-to-r from-amber-100 to-orange-200 text-amber-800 dark:from-amber-900/50 dark:to-orange-900/50 dark:text-amber-200',
        timelineColor: 'bg-gradient-to-b from-amber-500/20 to-amber-500/5'
      },
      error: {
        dotColor: 'bg-gradient-to-r from-red-500 to-rose-500',
        badgeClasses: 'bg-gradient-to-r from-red-100 to-rose-200 text-red-800 dark:from-red-900/50 dark:to-rose-900/50 dark:text-red-200',
        timelineColor: 'bg-gradient-to-b from-red-500/20 to-red-500/5'
      },
      pending: {
        dotColor: 'bg-gradient-to-r from-blue-500 to-indigo-500',
        badgeClasses: 'bg-gradient-to-r from-blue-100 to-indigo-200 text-blue-800 dark:from-blue-900/50 dark:to-indigo-900/50 dark:text-blue-200',
        timelineColor: 'bg-gradient-to-b from-blue-500/20 to-blue-500/5'
      }
    };

    const defaultTypeConfig = {
      icon: 'bell',
      bgColor: 'bg-gradient-to-br from-gray-100 to-slate-200 dark:from-gray-800/50 dark:to-slate-700/50',
      iconColor: 'text-gray-600 dark:text-gray-400',
      borderColor: 'border-gray-200/50 dark:border-gray-700/50'
    };

    const defaultStatusConfig = {
      dotColor: 'bg-gradient-to-r from-gray-500 to-slate-500',
      badgeClasses: 'bg-gradient-to-r from-gray-100 to-slate-200 text-gray-800 dark:from-gray-900/50 dark:to-slate-900/50 dark:text-gray-200',
      timelineColor: 'bg-gradient-to-b from-gray-500/20 to-gray-500/5'
    };

    return {
      ...typeConfig[type as keyof typeof typeConfig] || defaultTypeConfig,
      ...statusConfig[status as keyof typeof statusConfig] || defaultStatusConfig
    };
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
    <div className={`relative overflow-hidden rounded-2xl bg-white/70 dark:bg-gray-900/70 backdrop-blur-xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl ${className}`} data-testid={dataTestId}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-xl bg-gradient-to-br from-green-100 to-emerald-200 dark:from-green-800/50 dark:to-emerald-700/50">
              <Icon name="clock" className="h-5 w-5 text-green-600 dark:text-green-400" />
            </div>
            <div>
              <h3 className="text-xl font-bold bg-gradient-to-r from-gray-900 via-gray-700 to-gray-900 dark:from-white dark:via-gray-200 dark:to-white bg-clip-text text-transparent">
                Recent Activities
              </h3>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
                Live activity stream from your clusters
              </p>
            </div>
          </div>
          <div className="flex items-center space-x-3">
            <button
              type="button"
              onClick={onRefresh}
              disabled={isLoading}
              className="p-3 rounded-xl text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-green-500/20 disabled:opacity-50 transition-all duration-200"
              data-testid="refresh-button"
              aria-label="Refresh activities"
            >
              <Icon
                name={isLoading ? "spinner" : "refresh"}
                className={`h-5 w-5 ${isLoading ? 'animate-spin' : ''}`}
              />
            </button>
            <button
              type="button"
              onClick={onViewAll}
              className="px-4 py-2 text-sm font-medium text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 bg-green-100/50 dark:bg-green-900/20 rounded-xl hover:bg-green-200/50 dark:hover:bg-green-900/30 transition-all duration-200"
              data-testid="view-all-button"
            >
              View All →
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

        {/* Timeline Activities */}
        {!isLoading && (
          <div className="relative">
            {/* Timeline line */}
            <div className="absolute left-8 top-0 bottom-0 w-0.5 bg-gradient-to-b from-gray-200 via-gray-300 to-gray-200 dark:from-gray-700 dark:via-gray-600 dark:to-gray-700" />

            <div className="space-y-6">
              {displayActivities.map((activity, index) => {
                const config = getActivityConfig(activity.type, activity.status);

                return (
                  <div
                    key={activity.id}
                    className="group relative cursor-pointer transition-all duration-300 ease-out hover:scale-[1.02] hover:-translate-y-1"
                    onClick={() => onActivityClick?.(activity.id)}
                    data-testid={`activity-${activity.id}`}
                    style={{
                      animationDelay: `${index * 100}ms`,
                      animation: 'fadeInLeft 0.6s ease-out forwards'
                    }}
                  >
                    {/* Timeline dot */}
                    <div className="absolute left-6 top-6 transform -translate-x-1/2">
                      <div className="relative">
                        <div className={`w-4 h-4 rounded-full ${config.dotColor} shadow-lg border-2 border-white dark:border-gray-900 z-10 relative`} />
                        <div className={`absolute inset-0 w-4 h-4 rounded-full ${config.dotColor} opacity-30 animate-pulse scale-150`} />
                      </div>
                    </div>

                    {/* Activity card */}
                    <div className="ml-16 p-6 rounded-2xl bg-gradient-to-br from-white/50 to-gray-50/50 dark:from-gray-800/50 dark:to-gray-900/50 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50 shadow-sm hover:shadow-lg transition-all duration-300">
                      <div className="flex items-start justify-between">
                        <div className="flex items-start space-x-4 flex-1">
                          {/* Activity icon */}
                          <div className={`flex-shrink-0 w-12 h-12 rounded-xl ${config.bgColor} border ${config.borderColor} flex items-center justify-center shadow-sm group-hover:scale-110 transition-transform duration-300`}>
                            <Icon
                              name={config.icon}
                              className={`h-5 w-5 ${config.iconColor}`}
                            />
                          </div>

                          {/* Activity content */}
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between mb-2">
                              <h4 className="text-base font-semibold text-gray-900 dark:text-white group-hover:text-gray-800 dark:group-hover:text-gray-100 transition-colors duration-300">
                                {activity.title}
                              </h4>
                              <div className="flex items-center space-x-2 ml-4">
                                <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold ${config.badgeClasses} shadow-sm uppercase tracking-wider`}>
                                  {activity.status}
                                </span>
                              </div>
                            </div>

                            <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed mb-3">
                              {activity.description}
                            </p>

                            {/* Metadata */}
                            <div className="flex items-center justify-between">
                              <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
                                <div className="flex items-center space-x-1 px-2 py-1 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                                  <Icon name="user" className="h-3 w-3" />
                                  <span className="font-medium">{activity.user.name}</span>
                                </div>
                                {activity.cluster && (
                                  <div className="flex items-center space-x-1 px-2 py-1 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                                    <Icon name="server" className="h-3 w-3" />
                                    <span className="font-medium">{activity.cluster}</span>
                                  </div>
                                )}
                              </div>
                              <div className="flex items-center space-x-2 text-xs text-gray-500 dark:text-gray-400">
                                <Icon name="clock" className="h-3 w-3" />
                                <span className="font-medium">{formatTimeAgo(activity.timestamp)}</span>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Empty state */}
        {!isLoading && displayActivities.length === 0 && (
          <div className="text-center py-16">
            <div className="p-8 rounded-2xl bg-gradient-to-br from-gray-50 to-slate-100 dark:from-gray-800/50 dark:to-slate-900/50 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50 max-w-md mx-auto">
              <Icon name="clock" className="h-16 w-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
              <p className="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">No recent activities</p>
              <p className="text-sm text-gray-500 dark:text-gray-400 leading-relaxed">Activities will appear here as you interact with your clusters and execute commands</p>
              <button className="mt-4 px-4 py-2 text-sm font-medium text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 transition-colors duration-200">
                View All Activities →
              </button>
            </div>
          </div>
        )}

        {/* Enhanced footer */}
        {!isLoading && displayActivities.length > 0 && (
          <div className="mt-8 pt-6 border-t border-gray-200/50 dark:border-gray-700/50">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between space-y-3 sm:space-y-0">
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                  <div className="h-1.5 w-1.5 bg-emerald-500 rounded-full animate-pulse" />
                  <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
                    Showing {displayActivities.length} recent activities
                  </span>
                </div>
                <div className="flex items-center space-x-1">
                  <div className="h-1.5 w-1.5 bg-blue-500 rounded-full animate-pulse" />
                  <span className="text-xs font-medium text-blue-600 dark:text-blue-400">Live updates</span>
                </div>
              </div>

              <div className="flex items-center space-x-2">
                {[
                  { type: 'command', count: displayActivities.filter(a => a.type === 'command').length, color: 'blue' },
                  { type: 'deployment', count: displayActivities.filter(a => a.type === 'deployment').length, color: 'purple' },
                  { type: 'security', count: displayActivities.filter(a => a.type === 'security').length, color: 'emerald' }
                ].filter(({ count }) => count > 0).map(({ type, count, color }) => (
                  <div key={type} className="flex items-center space-x-1 px-2 py-1 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                    <div className={`h-1.5 w-1.5 bg-${color}-500 rounded-full`} />
                    <span className="text-xs font-medium text-gray-700 dark:text-gray-300 capitalize">
                      {type}: {count}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Add keyframes for animations */}
      <style jsx>{`
        @keyframes fadeInLeft {
          0% {
            opacity: 0;
            transform: translateX(-30px);
          }
          100% {
            opacity: 1;
            transform: translateX(0);
          }
        }
      `}</style>
    </div>
  );
};

export default RecentActivitiesFeed;
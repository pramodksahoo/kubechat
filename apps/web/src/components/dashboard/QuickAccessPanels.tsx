import React from 'react';
import { Icon } from '@/components/ui';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

interface QuickAccessItem {
  id: string;
  title: string;
  description: string;
  iconName: string;
  href?: string;
  onClick?: () => void;
  disabled?: boolean;
  badge?: string;
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'danger' | 'info';
  external?: boolean;
}

export interface QuickAccessPanelsProps extends BaseComponentProps {
  items?: QuickAccessItem[];
  columns?: 2 | 3 | 4;
  onItemClick?: (itemId: string) => void;
}

export const QuickAccessPanels: React.FC<QuickAccessPanelsProps> = ({
  items = [],
  columns = 3,
  onItemClick,
  className = '',
  'data-testid': dataTestId = 'quick-access-panels'
}) => {
  const defaultItems: QuickAccessItem[] = [
    {
      id: 'chat',
      title: 'Chat with KubeChat',
      description: 'Ask questions and execute kubectl commands safely',
      iconName: 'chat-bubble-left-right',
      href: '/chat',
      color: 'primary'
    },
    {
      id: 'clusters',
      title: 'Cluster Explorer',
      description: 'Browse and manage your Kubernetes clusters',
      iconName: 'server',
      href: '/clusters',
      color: 'info'
    },
    {
      id: 'deployments',
      title: 'Deployments',
      description: 'View and manage application deployments',
      iconName: 'rocket-launch',
      href: '/deployments',
      color: 'success'
    },
    {
      id: 'pods',
      title: 'Pod Management',
      description: 'Monitor and troubleshoot running pods',
      iconName: 'cube',
      href: '/pods',
      color: 'secondary'
    },
    {
      id: 'security',
      title: 'Security Dashboard',
      description: 'View security status and compliance reports',
      iconName: 'shield-check',
      href: '/security',
      color: 'warning',
      badge: 'New'
    },
    {
      id: 'audit',
      title: 'Audit Trail',
      description: 'Review command history and audit logs',
      iconName: 'document-check',
      href: '/audit',
      color: 'info'
    },
    {
      id: 'settings',
      title: 'Settings',
      description: 'Configure clusters and user preferences',
      iconName: 'cog',
      href: '/settings',
      color: 'secondary'
    },
    {
      id: 'help',
      title: 'Help & Documentation',
      description: 'Get help and view documentation',
      iconName: 'question-mark-circle',
      href: '/help',
      color: 'info'
    },
    {
      id: 'kubectl-docs',
      title: 'kubectl Reference',
      description: 'Access kubectl command documentation',
      iconName: 'book-open',
      href: 'https://kubernetes.io/docs/reference/kubectl/',
      color: 'primary',
      external: true
    }
  ];

  const displayItems = items.length > 0 ? items : defaultItems;

  const getGridCols = () => {
    switch (columns) {
      case 2: return 'grid-cols-1 md:grid-cols-2';
      case 3: return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3';
      case 4: return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4';
      default: return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3';
    }
  };

  const getColorClasses = (color: string) => {
    const colorMap = {
      primary: {
        bg: 'bg-primary-50 dark:bg-primary-900/20',
        border: 'border-primary-200 dark:border-primary-800',
        icon: 'text-primary-600 dark:text-primary-400',
        hover: 'hover:bg-primary-100 dark:hover:bg-primary-900/30'
      },
      secondary: {
        bg: 'bg-secondary-50 dark:bg-secondary-900/20',
        border: 'border-secondary-200 dark:border-secondary-800',
        icon: 'text-secondary-600 dark:text-secondary-400',
        hover: 'hover:bg-secondary-100 dark:hover:bg-secondary-900/30'
      },
      success: {
        bg: 'bg-success-50 dark:bg-success-900/20',
        border: 'border-success-200 dark:border-success-800',
        icon: 'text-success-600 dark:text-success-400',
        hover: 'hover:bg-success-100 dark:hover:bg-success-900/30'
      },
      warning: {
        bg: 'bg-warning-50 dark:bg-warning-900/20',
        border: 'border-warning-200 dark:border-warning-800',
        icon: 'text-warning-600 dark:text-warning-400',
        hover: 'hover:bg-warning-100 dark:hover:bg-warning-900/30'
      },
      danger: {
        bg: 'bg-danger-50 dark:bg-danger-900/20',
        border: 'border-danger-200 dark:border-danger-800',
        icon: 'text-danger-600 dark:text-danger-400',
        hover: 'hover:bg-danger-100 dark:hover:bg-danger-900/30'
      },
      info: {
        bg: 'bg-info-50 dark:bg-info-900/20',
        border: 'border-info-200 dark:border-info-800',
        icon: 'text-info-600 dark:text-info-400',
        hover: 'hover:bg-info-100 dark:hover:bg-info-900/30'
      }
    };

    return colorMap[color as keyof typeof colorMap] || colorMap.primary;
  };

  const handleItemClick = (item: QuickAccessItem) => {
    if (item.disabled) return;
    
    onItemClick?.(item.id);
    
    if (item.onClick) {
      item.onClick();
    } else if (item.href) {
      if (item.external) {
        window.open(item.href, '_blank', 'noopener,noreferrer');
      } else {
        // In a real app, use Next.js router
        window.location.href = item.href;
      }
    }
  };

  return (
    <div className={className} data-testid={dataTestId}>
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center">
          <Icon name="squares-plus" className="h-5 w-5 text-primary-500 mr-2" />
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
            Quick Access
          </h3>
        </div>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Commonly used actions and tools for cluster management
        </p>
      </div>

      {/* Quick access grid */}
      <div className={`grid ${getGridCols()} gap-4`}>
        {displayItems.map((item) => {
          const colorClasses = getColorClasses(item.color || 'primary');
          
          return (
            <div
              key={item.id}
              className={`relative overflow-hidden cursor-pointer transition-all duration-200 bg-white dark:bg-gray-900 rounded-lg border shadow-sm ${
                item.disabled
                  ? 'opacity-50 cursor-not-allowed'
                  : `${colorClasses.hover} hover:shadow-md hover:-translate-y-0.5`
              } ${colorClasses.bg} ${colorClasses.border}`}
              onClick={() => handleItemClick(item)}
              data-testid={`quick-access-${item.id}`}
            >
              <div className="p-6">
                {/* Badge */}
                {item.badge && (
                  <div className="absolute top-3 right-3">
                    <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800 dark:bg-primary-900 dark:text-primary-200">
                      {item.badge}
                    </span>
                  </div>
                )}

                {/* Icon */}
                <div className={`inline-flex items-center justify-center w-12 h-12 rounded-lg mb-4 ${colorClasses.bg}`}>
                  <Icon 
                    name={item.iconName} 
                    className={`h-6 w-6 ${colorClasses.icon}`} 
                  />
                </div>

                {/* Content */}
                <div>
                  <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-2 line-clamp-1">
                    {item.title}
                    {item.external && (
                      <Icon name="arrow-top-right-on-square" className="inline h-4 w-4 ml-1 text-gray-400" />
                    )}
                  </h4>
                  <p className="text-sm text-gray-600 dark:text-gray-300 line-clamp-2">
                    {item.description}
                  </p>
                </div>

                {/* Hover indicator */}
                {!item.disabled && (
                  <div className="absolute inset-0 bg-gradient-to-r from-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-200" />
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Empty state */}
      {displayItems.length === 0 && (
        <div className="text-center py-12">
          <Icon name="squares-plus" className="h-12 w-12 text-gray-300 mx-auto mb-4" />
          <p className="text-sm text-gray-500">No quick access items available</p>
          <p className="text-xs text-gray-400 mt-1">Quick access shortcuts will appear here</p>
        </div>
      )}

      {/* Footer tip */}
      {displayItems.length > 0 && (
        <div className="mt-6 p-4 bg-info-50 dark:bg-info-900/20 rounded-lg border border-info-200 dark:border-info-800">
          <div className="flex items-start">
            <Icon name="light-bulb" className="h-5 w-5 text-info-600 dark:text-info-400 mt-0.5 mr-3 flex-shrink-0" />
            <div>
              <p className="text-sm font-medium text-info-900 dark:text-info-100">
                Pro Tip
              </p>
              <p className="text-sm text-info-700 dark:text-info-200 mt-1">
                Use keyboard shortcuts to quickly access common actions. Press <kbd className="px-2 py-1 bg-white dark:bg-gray-800 border rounded text-xs">Ctrl+K</kbd> to open the command palette.
              </p>
            </div>
          </div>
        </div>
      )}

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

export default QuickAccessPanels;
import React from 'react';
import Image from 'next/image';
import { useNavigationStore } from '@/stores/navigationStore';
import { Icon } from '@/components/ui/Icon';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

export interface SidebarProps extends BaseComponentProps {
  isOpen?: boolean;
  onClose?: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  isOpen = true,
  onClose,
  className = '',
  'data-testid': dataTestId = 'sidebar'
}) => {
  const { navigationItems, activeItemId, setActiveItem } = useNavigationStore();

  const handleItemClick = (itemId: string, href?: string, onClick?: () => void) => {
    setActiveItem(itemId);
    if (onClick) {
      onClick();
    } else if (href) {
      // In a real app, use Next.js router
      window.location.href = href;
    }
    // Close sidebar on mobile after navigation
    if (window.innerWidth < 1024) {
      onClose?.();
    }
  };

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div 
          className="fixed inset-0 z-40 lg:hidden"
          onClick={onClose}
          aria-hidden="true"
        >
          <div className="fixed inset-0 bg-gray-600 bg-opacity-75"></div>
        </div>
      )}

      {/* Sidebar */}
      <div
        className={`fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-900 transform ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        } transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0 shadow-lg ${className}`}
        data-testid={dataTestId}
      >
        <div className="flex flex-col h-full">
          {/* Sidebar header with logo */}
          <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center">
              <div className="h-10 w-10 rounded-lg overflow-hidden">
                <Image 
                  src="/kubechat-logo.png" 
                  alt="KubeChat Logo"
                  width={40}
                  height={40}
                  className="h-full w-full object-contain"
                />
              </div>
              <span className="ml-3 text-lg font-semibold text-gray-900 dark:text-white">
                KubeChat
              </span>
            </div>
            
            {/* Close button for mobile */}
            <button
              type="button"
              className="lg:hidden p-2 rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
              onClick={onClose}
              data-testid="sidebar-close"
              aria-label="Close sidebar"
            >
              <Icon name="close" />
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-4 py-6 space-y-2 overflow-y-auto">
            {navigationItems.map((item) => {
              const isActive = item.id === activeItemId;
              
              return (
                <div key={item.id}>
                  <button
                    onClick={() => handleItemClick(item.id, item.href, item.onClick)}
                    disabled={item.disabled}
                    className={`group flex items-center w-full px-3 py-2 text-sm font-medium rounded-md transition-colors duration-200 ${
                      item.disabled
                        ? 'text-gray-400 cursor-not-allowed'
                        : isActive
                        ? 'bg-primary-50 dark:bg-primary-900 text-primary-700 dark:text-primary-300 border-r-2 border-primary-500'
                        : 'text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-gray-800'
                    }`}
                    data-testid={`nav-${item.id}`}
                    aria-current={isActive ? 'page' : undefined}
                  >
                    {item.iconName && (
                      <Icon 
                        name={item.iconName} 
                        className={`mr-3 flex-shrink-0 ${
                          isActive ? 'text-primary-500' : 'text-gray-400 group-hover:text-gray-500'
                        }`}
                      />
                    )}
                    <span className="truncate">{item.label}</span>
                    {item.badge && (
                      <span className="ml-auto inline-flex items-center px-2 py-1 text-xs font-medium bg-primary-100 text-primary-800 rounded-full dark:bg-primary-900 dark:text-primary-200">
                        {item.badge}
                      </span>
                    )}
                    {item.external && (
                      <svg className="ml-auto h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                      </svg>
                    )}
                  </button>
                  
                  {/* Sub-navigation - Future feature */}
                  {item.children && item.children.length > 0 && isActive && (
                    <div className="ml-8 mt-1 space-y-1">
                      {item.children.map((subItem) => (
                        <button
                          key={subItem.id}
                          onClick={() => handleItemClick(subItem.id, subItem.href, subItem.onClick)}
                          className="group flex items-center w-full px-3 py-1 text-xs font-medium text-gray-600 dark:text-gray-400 rounded-md hover:text-gray-900 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors duration-200"
                          data-testid={`nav-${subItem.id}`}
                        >
                          {subItem.iconName && (
                            <Icon name={subItem.iconName} className="mr-2 h-4 w-4" />
                          )}
                          <span className="truncate">{subItem.label}</span>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </nav>

          {/* Footer with system status */}
          <div className="flex-shrink-0 flex border-t border-gray-200 dark:border-gray-700 p-4">
            <div className="flex-shrink-0 w-full">
              <div className="flex items-center">
                <div className="h-2 w-2 bg-success-500 rounded-full animate-pulse-subtle"></div>
                <div className="ml-3">
                  <p className="text-xs font-medium text-gray-700 dark:text-gray-300">System Status</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">All services operational</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default Sidebar;
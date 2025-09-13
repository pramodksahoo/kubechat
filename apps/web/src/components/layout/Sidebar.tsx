import React, { useState } from 'react';
import Image from 'next/image';
import { useRouter } from 'next/router';
import { useNavigationStore } from '@/stores/navigationStore';
import { Icon } from '@/components/ui';

// Tooltip component for collapsed sidebar
const Tooltip: React.FC<{ content: string; children: React.ReactNode; position?: 'right' | 'left' }> = ({
  content,
  children,
  position = 'right'
}) => {
  const [isVisible, setIsVisible] = useState(false);

  return (
    <div
      className="relative"
      onMouseEnter={() => setIsVisible(true)}
      onMouseLeave={() => setIsVisible(false)}
    >
      {children}
      {isVisible && (
        <div
          className={`absolute z-50 px-3 py-2 text-sm font-medium text-white bg-gray-900 dark:bg-gray-700 rounded-lg shadow-lg whitespace-nowrap ${
            position === 'right'
              ? 'left-full top-1/2 transform -translate-y-1/2 ml-3'
              : 'right-full top-1/2 transform -translate-y-1/2 mr-3'
          }`}
        >
          {content}
          <div
            className={`absolute top-1/2 transform -translate-y-1/2 w-2 h-2 bg-gray-900 dark:bg-gray-700 rotate-45 ${
              position === 'right' ? '-left-1' : '-right-1'
            }`}
          />
        </div>
      )}
    </div>
  );
};

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

export interface SidebarProps extends BaseComponentProps {
  isOpen?: boolean;
  onClose?: () => void;
  isCollapsed?: boolean;
  onToggleCollapse?: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  isOpen = true,
  onClose,
  isCollapsed = false,
  onToggleCollapse,
  className = '',
  'data-testid': dataTestId = 'sidebar'
}) => {
  const { navigationItems, activeItemId, setActiveItem } = useNavigationStore();
  const router = useRouter();

  // Update active item based on current route
  React.useEffect(() => {
    const currentPath = router.asPath;
    const currentItem = navigationItems.find(item => item.href === currentPath || (item.href === '/' && currentPath === '/'));
    if (currentItem && currentItem.id !== activeItemId) {
      setActiveItem(currentItem.id);
    }
  }, [router.asPath, navigationItems, activeItemId, setActiveItem]);

  const handleItemClick = async (itemId: string, href?: string, onClick?: () => void) => {
    setActiveItem(itemId);
    if (onClick) {
      onClick();
    } else if (href) {
      await router.push(href);
    }
    // Close sidebar on mobile after navigation
    if (typeof window !== 'undefined' && window.innerWidth < 1024) {
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
        className={`fixed inset-y-0 left-0 z-40 ${isCollapsed && !isOpen ? 'w-16' : 'w-64'} bg-white/70 dark:bg-gray-900/70 backdrop-blur-xl border-r border-gray-200/50 dark:border-gray-700/50 transform ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        } transition-all duration-300 ease-in-out lg:relative lg:transform-none lg:flex lg:flex-col lg:h-screen shadow-xl ${className}`}
        data-testid={dataTestId}
      >
        <div className="flex flex-col h-full lg:h-screen">
          {/* Header Section */}
          <div className="flex flex-col items-center justify-center py-6 px-4 border-b border-gray-200/50 dark:border-gray-700/50">
            {/* Logo */}
            <Tooltip content="KubeChat Dashboard" position="right">
              <button
                onClick={async () => await router.push('/')}
                className="hover:scale-105 transition-transform duration-200 focus:outline-none w-full flex justify-center"
                data-testid="sidebar-logo"
                aria-label="Go to dashboard"
              >
                <Image
                  src="/kubechat-logo.png"
                  alt="KubeChat Logo"
                  width={isCollapsed ? 48 : 160}
                  height={isCollapsed ? 48 : 160}
                  className={`object-contain transition-all duration-300 ${
                    isCollapsed ? 'h-12 w-12' : 'h-32 w-32 max-w-full'
                  }`}
                />
              </button>
            </Tooltip>

            {/* Collapse/Expand Toggle - Desktop only */}
            {onToggleCollapse && (
              <Tooltip content={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'} position="right">
                <button
                  type="button"
                  className="hidden lg:flex mt-3 p-2 rounded-lg text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-primary-500/20 transition-all duration-200"
                  onClick={onToggleCollapse}
                  data-testid="sidebar-toggle"
                  aria-label={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                >
                  <Icon name={isCollapsed ? 'chevron-right' : 'chevron-left'} className="w-4 h-4" />
                </button>
              </Tooltip>
            )}

            {/* Close button for mobile */}
            <button
              type="button"
              className="lg:hidden absolute top-4 right-4 p-2 rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
              onClick={onClose}
              data-testid="sidebar-close"
              aria-label="Close sidebar"
            >
              <Icon name="close" />
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
            {navigationItems.map((item) => {
              const isActive = item.id === activeItemId;
              const NavButton = (
                <button
                  onClick={() => handleItemClick(item.id, item.href, item.onClick)}
                  disabled={item.disabled}
                  className={`group flex items-center w-full ${isCollapsed ? 'justify-center px-3 py-3' : 'px-3 py-2.5'} text-sm font-medium rounded-xl transition-all duration-200 ${
                    item.disabled
                      ? 'text-gray-400 cursor-not-allowed'
                      : isActive
                      ? 'bg-gradient-to-r from-primary-500/10 to-primary-600/10 text-primary-700 dark:text-primary-300 border border-primary-200 dark:border-primary-800 shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100/50 dark:hover:bg-gray-800/50'
                  }`}
                  data-testid={`nav-${item.id}`}
                  aria-current={isActive ? 'page' : undefined}
                >
                  {item.iconName && (
                    <Icon
                      name={item.iconName}
                      className={`${isCollapsed ? '' : 'mr-3'} flex-shrink-0 w-5 h-5 ${
                        isActive ? 'text-primary-600 dark:text-primary-400' : 'text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300'
                      }`}
                    />
                  )}
                  {!isCollapsed && (
                    <>
                      <span className="truncate font-medium">{item.label}</span>
                      {item.badge && (
                        <span className="ml-auto inline-flex items-center px-2 py-0.5 text-xs font-medium bg-gradient-to-r from-primary-100 to-primary-200 text-primary-800 rounded-full dark:from-primary-900/50 dark:to-primary-800/50 dark:text-primary-200">
                          {item.badge}
                        </span>
                      )}
                      {item.external && (
                        <svg className="ml-auto h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                        </svg>
                      )}
                    </>
                  )}
                </button>
              );

              return (
                <div key={item.id}>
                  {isCollapsed ? (
                    <Tooltip content={item.label} position="right">
                      {NavButton}
                    </Tooltip>
                  ) : (
                    NavButton
                  )}

                  {/* Sub-navigation - Only show when not collapsed */}
                  {!isCollapsed && item.children && item.children.length > 0 && isActive && (
                    <div className="ml-8 mt-1 space-y-1">
                      {item.children.map((subItem) => (
                        <button
                          key={subItem.id}
                          onClick={() => handleItemClick(subItem.id, subItem.href, subItem.onClick)}
                          className="group flex items-center w-full px-3 py-1.5 text-xs font-medium text-gray-600 dark:text-gray-400 rounded-lg hover:text-gray-900 dark:hover:text-white hover:bg-gray-100/50 dark:hover:bg-gray-800/50 transition-all duration-200"
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
          <div className="flex-shrink-0 border-t border-gray-200/50 dark:border-gray-700/50 p-3">
            <div className="flex-shrink-0 w-full">
              {isCollapsed ? (
                <Tooltip content="System Status: Operational" position="right">
                  <div className="flex justify-center">
                    <div className="h-2 w-2 bg-emerald-500 rounded-full animate-pulse shadow-sm"></div>
                  </div>
                </Tooltip>
              ) : (
                <div className="flex items-center px-3 py-2 rounded-xl bg-emerald-50/50 dark:bg-emerald-900/10 border border-emerald-200/50 dark:border-emerald-800/50">
                  <div className="h-2 w-2 bg-emerald-500 rounded-full animate-pulse shadow-sm"></div>
                  <div className="ml-3">
                    <p className="text-xs font-medium text-emerald-700 dark:text-emerald-300">System Status</p>
                    <p className="text-xs text-emerald-600/80 dark:text-emerald-400/80">All services operational</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default Sidebar;
import React from 'react';
import { useRouter } from 'next/router';
import { useNavigationStore } from '@/stores/navigationStore';
import { Icon } from '@/components/ui';


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
        className={`w-64 bg-white dark:bg-gray-950 border-r border-gray-200 dark:border-gray-800 flex-shrink-0 ${
          isOpen ? 'block' : 'hidden'
        } lg:block transition-all duration-300 ease-in-out shadow-lg ${className}`}
        data-testid={dataTestId}
      >
        <div className="flex flex-col h-full" style={{ height: 'calc(100vh - 4rem)' }}>

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

          {/* Navigation */}
          <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
            {navigationItems.map((item) => {
              const isActive = item.id === activeItemId;
              const NavButton = (
                <button
                  onClick={() => handleItemClick(item.id, item.href, item.onClick)}
                  disabled={item.disabled}
                  className={`group flex items-center w-full px-3 py-2.5 text-sm font-medium rounded-xl transition-all duration-200 ${
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
                      className={`mr-3 flex-shrink-0 w-5 h-5 ${
                        isActive ? 'text-primary-600 dark:text-primary-400' : 'text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300'
                      }`}
                    />
                  )}
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
                </button>
              );

              return (
                <div key={item.id}>
                  {NavButton}

                  {/* Sub-navigation */}
                  {item.children && item.children.length > 0 && isActive && (
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
          <div className="flex-shrink-0 border-t border-gray-200 dark:border-gray-800 p-3">
            <div className="flex items-center px-3 py-2 rounded-xl bg-emerald-50 dark:bg-emerald-900/10 border border-emerald-200 dark:border-emerald-800">
              <div className="h-2 w-2 bg-emerald-500 rounded-full animate-pulse shadow-sm"></div>
              <div className="ml-3">
                <p className="text-xs font-medium text-emerald-700 dark:text-emerald-300">System Status</p>
                <p className="text-xs text-emerald-600/80 dark:text-emerald-400/80">All services operational</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default Sidebar;
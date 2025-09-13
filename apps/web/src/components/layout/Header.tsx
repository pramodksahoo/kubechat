import React from 'react';
import Image from 'next/image';
import { UserProfile } from './UserProfile';
import { useNavigationStore } from '@/stores/navigationStore';
import { Icon } from '@/components/ui/Icon';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

export interface HeaderProps extends BaseComponentProps {
  title?: string;
  showUserMenu?: boolean;
  onMenuToggle?: () => void;
}

export const Header: React.FC<HeaderProps> = ({ 
  title = 'KubeChat',
  showUserMenu = true,
  onMenuToggle,
  className = '',
  children,
  'data-testid': dataTestId = 'header'
}) => {
  const { toggleSidebar, breadcrumbs } = useNavigationStore();

  const handleMenuToggle = () => {
    if (onMenuToggle) {
      onMenuToggle();
    } else {
      toggleSidebar();
    }
  };

  return (
    <header 
      className={`bg-white dark:bg-gray-900 shadow-sm border-b border-gray-200 dark:border-gray-700 ${className}`}
      data-testid={dataTestId}
    >
      <div className="px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Left section with logo and menu toggle */}
          <div className="flex items-center">
            <button
              type="button"
              className="lg:hidden -ml-0.5 -mt-0.5 h-12 w-12 inline-flex items-center justify-center rounded-md text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary-500 transition-colors duration-200"
              onClick={handleMenuToggle}
              data-testid="mobile-menu-toggle"
              aria-label="Toggle navigation menu"
            >
              <Icon name="menu" className="text-current" />
            </button>
            
            <div className="flex items-center ml-4 lg:ml-0">
              <div className="flex-shrink-0">
                <div className="flex items-center">
                  {/* KubeChat Logo */}
                  <div className="h-8 w-8 rounded-lg overflow-hidden">
                    <Image 
                      src="/kubechat-icon.png" 
                      alt="KubeChat"
                      width={32}
                      height={32}
                      className="h-full w-full object-contain"
                    />
                  </div>
                  <span className="ml-3 text-xl font-semibold text-gray-900 dark:text-white">
                    {title}
                  </span>
                </div>
              </div>
              
              {/* Breadcrumbs */}
              {breadcrumbs.length > 0 && (
                <nav className="hidden md:flex ml-8" aria-label="Breadcrumb">
                  <ol className="flex items-center space-x-2">
                    {breadcrumbs.map((crumb, index) => (
                      <li key={`breadcrumb-${index}`} className="flex items-center">
                        {index > 0 && (
                          <svg 
                            className="flex-shrink-0 h-5 w-5 text-gray-400 mx-2" 
                            fill="currentColor" 
                            viewBox="0 0 20 20"
                            aria-hidden="true"
                          >
                            <path 
                              fillRule="evenodd" 
                              d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" 
                              clipRule="evenodd" 
                            />
                          </svg>
                        )}
                        {crumb.href ? (
                          <a
                            href={crumb.href}
                            className="text-sm font-medium text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors duration-200"
                          >
                            {crumb.label}
                          </a>
                        ) : (
                          <span className="text-sm font-medium text-gray-900 dark:text-white">
                            {crumb.label}
                          </span>
                        )}
                      </li>
                    ))}
                  </ol>
                </nav>
              )}
            </div>
          </div>

          {/* Center section for search or additional content */}
          <div className="flex-1 flex justify-center px-6 max-w-lg">
            {children}
          </div>

          {/* Right section with system status and user profile */}
          <div className="flex items-center space-x-4">
            {/* Global system status indicator */}
            <div className="flex items-center space-x-2">
              <div className="h-2 w-2 bg-success-500 rounded-full animate-pulse-subtle" aria-hidden="true"></div>
              <span className="text-sm text-gray-500 dark:text-gray-400 hidden sm:inline">
                All Systems Operational
              </span>
            </div>

            {/* Notifications */}
            <button
              type="button"
              className="relative p-1 rounded-full text-gray-400 hover:text-gray-500 dark:text-gray-300 dark:hover:text-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors duration-200"
              data-testid="notifications-button"
              aria-label="View notifications"
            >
              <Icon name="bell" className="text-current" />
              <span 
                className="absolute top-0 right-0 block h-2 w-2 rounded-full bg-warning-400 ring-2 ring-white dark:ring-gray-900"
                aria-hidden="true"
              ></span>
            </button>

            {/* User profile dropdown */}
            {showUserMenu && (
              <UserProfile 
                onSignOut={() => {
                  // Security: Clear sensitive data before sign out
                  if (typeof window !== 'undefined') {
                    localStorage.removeItem('token');
                    sessionStorage.clear();
                  }
                  console.log('Sign out - redirecting to login');
                }}
                onProfile={() => console.log('Navigate to profile')}
                onSettings={() => console.log('Navigate to settings')}
              />
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
import React from 'react';
import Image from 'next/image';
import { useRouter } from 'next/router';
import { UserProfile } from './UserProfile';
import { useNavigationStore } from '@/stores/navigationStore';
import { useTheme } from '@/providers/ThemeProvider';
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
  showUserMenu = true,
  onMenuToggle,
  className = '',
  children,
  'data-testid': dataTestId = 'header'
}) => {
  const { toggleSidebar, breadcrumbs } = useNavigationStore();
  const { theme, resolvedTheme, toggleTheme } = useTheme();
  const router = useRouter();

  const handleMenuToggle = () => {
    if (onMenuToggle) {
      onMenuToggle();
    } else {
      toggleSidebar();
    }
  };

  return (
    <header
      className={`backdrop-blur-xl bg-white dark:bg-gray-950 border-b border-gray-200 dark:border-gray-800 flex-shrink-0 ${className}`}
      data-testid={dataTestId}
    >
      <div className="flex justify-between items-center h-16 px-6">
        {/* Left section - Logo and Mobile menu */}
        <div className="flex items-center space-x-4">
          {/* Big Icon with Brand Text */}
          <button
            onClick={() => router.push('/')}
            className="flex items-center space-x-4 hover:scale-105 transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-primary-500/20 rounded-xl p-2"
            data-testid="header-logo"
            aria-label="Go to dashboard"
          >
            {/* Big KubeChat Icon */}
            <Image
              src="/kubechat-icon.png"
              alt="KubeChat Icon"
              width={48}
              height={48}
              className="object-contain h-12 w-12 drop-shadow-lg"
              priority
            />

            {/* Brand Text with Professional Colors */}
            <div className="hidden sm:block">
              <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-blue-700 dark:from-blue-400 dark:to-blue-500 bg-clip-text text-transparent tracking-tight">
                KubeChat
              </h1>
            </div>
          </button>

          {/* Mobile menu toggle */}
          <button
            type="button"
            className="lg:hidden p-2 rounded-xl text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-primary-500/20 transition-all duration-200"
            onClick={handleMenuToggle}
            data-testid="mobile-menu-toggle"
            aria-label="Toggle navigation menu"
          >
            <Icon name="menu" className="w-6 h-6" />
          </button>
        </div>

        {/* Center section - Breadcrumbs */}
        <div className="flex-1 flex items-center justify-center">
          {breadcrumbs.length > 0 && (
            <nav className="flex items-center" aria-label="Breadcrumb">
              <ol className="flex items-center space-x-3">
                {breadcrumbs.map((crumb, index) => (
                  <li key={`breadcrumb-${index}`} className="flex items-center">
                    {index > 0 && (
                      <svg
                        className="flex-shrink-0 w-4 h-4 text-gray-400 mx-3"
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
                        className="text-sm font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 transition-colors duration-200"
                      >
                        {crumb.label}
                      </a>
                    ) : (
                      <span className="text-sm font-semibold text-gray-900 dark:text-white">
                        {crumb.label}
                      </span>
                    )}
                  </li>
                ))}
              </ol>
            </nav>
          )}
          {children}
        </div>

        {/* Right section - Controls and user menu */}
        <div className="flex items-center space-x-3">
            {/* System status indicator */}
            <div className="hidden sm:flex items-center space-x-3 px-3 py-1.5 rounded-full bg-emerald-50 dark:bg-emerald-900/20">
              <div className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
              <span className="text-xs font-medium text-emerald-700 dark:text-emerald-300">
                Operational
              </span>
            </div>

            {/* Quick actions */}
            <div className="flex items-center space-x-1">
              {/* Search */}
              <button
                type="button"
                className="p-2 rounded-xl text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-primary-500/20 transition-all duration-200"
                aria-label="Search"
              >
                <Icon name="search" className="w-5 h-5" />
              </button>

              {/* Theme Toggle */}
              <button
                type="button"
                onClick={toggleTheme}
                className="p-2 rounded-xl text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-primary-500/20 transition-all duration-200"
                data-testid="theme-toggle"
                aria-label={`Switch to ${resolvedTheme === 'light' ? 'dark' : 'light'} theme`}
                title={`Current: ${theme === 'system' ? `System (${resolvedTheme})` : resolvedTheme} • Click to toggle`}
              >
                {resolvedTheme === 'light' ? (
                  <Icon name="moon" className="w-5 h-5" />
                ) : (
                  <Icon name="sun" className="w-5 h-5" />
                )}
              </button>

              {/* Notifications */}
              <button
                type="button"
                className="relative p-2 rounded-xl text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-primary-500/20 transition-all duration-200"
                data-testid="notifications-button"
                aria-label="View notifications"
              >
                <Icon name="bell" className="w-5 h-5" />
                <span
                  className="absolute -top-0.5 -right-0.5 w-3 h-3 rounded-full bg-gradient-to-r from-orange-400 to-red-500 ring-2 ring-white dark:ring-gray-900"
                  aria-hidden="true"
                ></span>
              </button>

              {/* Settings */}
              <button
                type="button"
                className="p-2 rounded-xl text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100/50 dark:hover:bg-gray-800/50 focus:outline-none focus:ring-2 focus:ring-primary-500/20 transition-all duration-200"
                aria-label="Settings"
              >
                <Icon name="settings" className="w-5 h-5" />
              </button>
            </div>

            {/* User profile */}
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
    </header>
  );
};

export default Header;
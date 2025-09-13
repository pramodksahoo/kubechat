import React, { useState } from 'react';
import Image from 'next/image';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

interface UserType {
  id: string;
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  avatar?: string;
  roles: unknown[];
  permissions: unknown[];
  clusters: unknown[];
  preferences: unknown;
  createdAt: string;
  updatedAt: string;
  isActive: boolean;
}

export interface UserProfileProps extends BaseComponentProps {
  user?: UserType;
  onSignOut?: () => void;
  onProfile?: () => void;
  onSettings?: () => void;
}

export const UserProfile: React.FC<UserProfileProps> = ({
  user,
  onSignOut,
  onProfile,
  onSettings,
  className = '',
  'data-testid': dataTestId = 'user-profile'
}) => {
  const [isOpen, setIsOpen] = useState(false);

  const defaultUser: UserType = {
    id: 'demo-user',
    username: 'admin',
    email: 'admin@kubechat.dev',
    firstName: 'KubeChat',
    lastName: 'Admin',
    roles: [],
    permissions: [],
    clusters: [],
    preferences: {
      theme: 'system',
      language: 'en',
      timezone: 'UTC',
      notifications: {
        email: true,
        browser: true,
        mobile: false,
        commandApprovals: true,
        systemAlerts: true,
        auditAlerts: true
      },
      dashboard: {
        layout: 'comfortable',
        refreshInterval: 30
      }
    },
    createdAt: '2023-01-01T00:00:00Z',
    updatedAt: '2023-01-01T00:00:00Z',
    isActive: true
  };

  const displayUser = user || defaultUser;
  const initials = `${displayUser.firstName?.[0] || 'K'}${displayUser.lastName?.[0] || 'A'}`;

  return (
    <div className={`relative ${className}`} data-testid={dataTestId}>
      <button
        type="button"
        className="flex items-center justify-center p-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 hover:scale-105 transition-all duration-200 rounded-full"
        onClick={() => setIsOpen(!isOpen)}
        data-testid={`${dataTestId}-trigger`}
      >
        {/* Avatar only */}
        <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-lg">
          {displayUser.avatar ? (
            <Image
              className="h-10 w-10 rounded-full object-cover"
              src={displayUser.avatar}
              alt={`${displayUser.firstName} ${displayUser.lastName}`}
              width={40}
              height={40}
            />
          ) : (
            <span className="text-lg font-bold text-white">{initials}</span>
          )}
        </div>
      </button>

      {/* Dropdown menu */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-[9998]"
            onClick={() => setIsOpen(false)}
          />

          {/* Menu */}
          <div
            className="absolute right-0 z-[10001] mt-2 w-80 origin-top-right rounded-2xl bg-white/95 dark:bg-gray-800/95 backdrop-blur-xl py-2 shadow-2xl border border-gray-200/50 dark:border-gray-700/50 focus:outline-none"
            data-testid={`${dataTestId}-menu`}
          >
            {/* User info section */}
            <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
              <div className="flex items-center space-x-3">
                <div className="h-12 w-12 rounded-full bg-primary-500 flex items-center justify-center">
                  {displayUser.avatar ? (
                    <Image
                      className="h-12 w-12 rounded-full object-cover"
                      src={displayUser.avatar}
                      alt={`${displayUser.firstName} ${displayUser.lastName}`}
                      width={48}
                      height={48}
                    />
                  ) : (
                    <span className="text-lg font-medium text-white">{initials}</span>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                    {displayUser.firstName} {displayUser.lastName}
                  </p>
                  <p className="text-sm text-gray-500 dark:text-gray-400 truncate">
                    {displayUser.email}
                  </p>
                  <div className="flex items-center mt-1">
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-success-100 text-success-800 dark:bg-success-900 dark:text-success-200">
                      Online
                    </span>
                    <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">
                      {displayUser.clusters.length} clusters
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Menu items */}
            <div className="py-1">
              <button
                className="group flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                onClick={onProfile}
                data-testid={`${dataTestId}-profile`}
              >
                <svg className="mr-3 h-4 w-4 text-gray-400 group-hover:text-gray-500" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
                </svg>
                Your Profile
              </button>
              
              <button
                className="group flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                onClick={onSettings}
                data-testid={`${dataTestId}-settings`}
              >
                <svg className="mr-3 h-4 w-4 text-gray-400 group-hover:text-gray-500" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a6.759 6.759 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
                </svg>
                Settings
              </button>

              <div className="border-t border-gray-200 dark:border-gray-700 my-1" />

              <div className="px-4 py-2">
                <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">Active Sessions</p>
                <p className="text-xs text-gray-700 dark:text-gray-300">2 active sessions</p>
              </div>

              <div className="border-t border-gray-200 dark:border-gray-700 my-1" />

              <button
                className="group flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                onClick={onSignOut}
                data-testid={`${dataTestId}-signout`}
              >
                <svg className="mr-3 h-4 w-4 text-gray-400 group-hover:text-gray-500" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15M12 9l-3 3m0 0 3 3m-3-3h12.75" />
                </svg>
                Sign out
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default UserProfile;
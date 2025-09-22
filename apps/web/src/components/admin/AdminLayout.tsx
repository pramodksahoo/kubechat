// Admin Layout Component
// Following coding standards from docs/architecture/coding-standards.md
// AC 1: Admin Route Creation with admin-only access

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import Link from 'next/link';
import { useAuthStore } from '../../stores/authStore';
import { useAdminStore } from '../../stores/adminStore';

// Icon component with SVG icons
const Icon: React.FC<{ name: string; className?: string }> = ({ name, className = "w-5 h-5" }) => {
  const icons = {
    ChartBarIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
    ),
    UserGroupIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
      </svg>
    ),
    ShieldCheckIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
      </svg>
    ),
    LinkIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
      </svg>
    ),
    ShieldExclamationIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.618 5.984A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622a12.02 12.02 0 00.382-3.016zM12 9v2m0 4h.01" />
      </svg>
    ),
    DocumentTextIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
    KeyIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
      </svg>
    ),
    ChatBubbleLeftRightIcon: (
      <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
      </svg>
    )
  };
  
  return icons[name as keyof typeof icons] || <div className={className}></div>;
};

export interface AdminLayoutProps {
  children: React.ReactNode;
  title?: string;
}

const AdminNavigation: React.FC = () => {
  const router = useRouter();
  const { setCurrentView } = useAdminStore();

  const navItems = [
    { key: 'dashboard', label: 'Dashboard', href: '/admin', icon: 'ChartBarIcon' },
    { key: 'users', label: 'User Management', href: '/admin/users', icon: 'UserGroupIcon' },
    { key: 'roles', label: 'Role Management', href: '/admin/roles', icon: 'ShieldCheckIcon' },
    { key: 'sessions', label: 'Active Sessions', href: '/admin/sessions', icon: 'LinkIcon' },
    { key: 'security', label: 'Security Settings', href: '/admin/security', icon: 'ShieldExclamationIcon' },
    { key: 'audit', label: 'Audit Logs', href: '/admin/audit', icon: 'DocumentTextIcon' },
    { key: 'credentials', label: 'Admin Credentials', href: '/admin/credentials', icon: 'KeyIcon' },
    { key: 'chat', label: 'Admin Chat', href: '/admin/chat', icon: 'ChatBubbleLeftRightIcon' }
  ];

  const handleNavClick = (key: string) => {
    setCurrentView(key as any);
  };

  return (
    <nav className="bg-gray-800 text-white w-64 min-h-screen p-4">
      <div className="mb-8">
        <h1 className="text-xl font-bold text-blue-300">KubeChat Admin</h1>
        <p className="text-sm text-gray-400">System Administration</p>
      </div>

      <ul className="space-y-2">
        {navItems.map((item) => {
          const isActive = router.pathname === item.href;
          return (
            <li key={item.key}>
              <Link
                href={item.href}
                onClick={() => handleNavClick(item.key)}
                className={`
                  flex items-center space-x-3 px-4 py-3 rounded-lg transition-colors duration-200
                  ${isActive
                    ? 'bg-blue-600 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                  }
                `}
              >
                <Icon name={item.icon} className="w-5 h-5" />
                <span className="font-medium">{item.label}</span>
              </Link>
            </li>
          );
        })}
      </ul>

      <div className="mt-8 pt-8 border-t border-gray-700">
        <div className="text-xs text-gray-400 space-y-2">
          <div className="flex items-center space-x-2">
            <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
            </svg>
            <span>Admin Access Only</span>
          </div>
          <div className="flex items-center space-x-2">
            <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clipRule="evenodd" />
            </svg>
            <span>SOC 2 Compliant</span>
          </div>
          <div className="flex items-center space-x-2">
            <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 1L3 6v4c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V6l-7-5z" clipRule="evenodd" />
            </svg>
            <span>GDPR Protected</span>
          </div>
        </div>
      </div>
    </nav>
  );
};

const AdminHeader: React.FC<{ title?: string }> = ({ title }) => {
  const { user, logout } = useAuthStore();
  const { error, clearError } = useAdminStore();

  const handleLogout = () => {
    logout();
  };

  return (
    <header className="bg-white border-b border-gray-200 px-6 py-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            {title || 'Admin Dashboard'}
          </h1>
          <p className="text-sm text-gray-600">
            Enterprise User Management & Security Controls
          </p>
        </div>

        <div className="flex items-center space-x-4">
          {/* Error notification */}
          {error && (
            <div className="flex items-center space-x-2 bg-red-100 text-red-700 px-3 py-2 rounded-md">
              <span className="text-sm">{error}</span>
              <button
                onClick={clearError}
                className="text-red-500 hover:text-red-700"
                aria-label="Clear error"
              >
                ×
              </button>
            </div>
          )}

          {/* User info */}
          <div className="flex items-center space-x-3">
            <div className="text-right">
              <p className="text-sm font-medium text-gray-900">
                {user?.username}
              </p>
              <p className="text-xs text-gray-500">
                {user?.role} • Admin Access
              </p>
            </div>
            <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center">
              <span className="text-white text-sm font-semibold">
                {user?.username?.charAt(0).toUpperCase()}
              </span>
            </div>
          </div>

          {/* Logout button */}
          <button
            onClick={handleLogout}
            className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
          >
            Logout
          </button>
        </div>
      </div>
    </header>
  );
};

const AdminSecurityBanner: React.FC = () => {
  const [isVisible, setIsVisible] = useState(true);

  if (!isVisible) return null;

  return (
    <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-4">
      <div className="flex items-center justify-between">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="w-5 h-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <p className="text-sm text-yellow-700">
              <strong>Security Notice:</strong> You are accessing administrative functions.
              All actions are logged for compliance and security monitoring.
            </p>
          </div>
        </div>
        <div className="ml-auto pl-3">
          <button
            onClick={() => setIsVisible(false)}
            className="text-yellow-400 hover:text-yellow-600"
          >
            <span className="sr-only">Dismiss</span>
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
};

export const AdminLayout: React.FC<AdminLayoutProps> = ({ children, title }) => {
  const { user, isAuthenticated, isLoading } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    // Redirect non-admin users
    if (!isLoading && isAuthenticated && user && user.role !== 'admin') {
      router.replace('/403'); // Access denied page
      return;
    }

    // Redirect unauthenticated users
    if (!isLoading && !isAuthenticated) {
      router.replace('/auth/login?returnUrl=' + encodeURIComponent(router.asPath));
      return;
    }
  }, [isAuthenticated, isLoading, user, router]);

  // Show loading state
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <p className="text-gray-600">Loading admin interface...</p>
        </div>
      </div>
    );
  }

  // Show access denied for non-admin users
  if (isAuthenticated && user && user.role !== 'admin') {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="w-16 h-16 mx-auto bg-red-100 rounded-full flex items-center justify-center">
            <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900">Admin Access Required</h2>
          <p className="text-gray-600">
            You need administrator privileges to access this page.
          </p>
          <button
            onClick={() => router.back()}
            className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  // Don't render if not authenticated or not admin
  if (!isAuthenticated || !user || user.role !== 'admin') {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar Navigation */}
      <AdminNavigation />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <AdminHeader title={title} />

        {/* Content Area */}
        <main className="flex-1 p-6">
          <AdminSecurityBanner />
          {children}
        </main>

        {/* Footer */}
        <footer className="bg-white border-t border-gray-200 px-6 py-4">
          <div className="flex items-center justify-between text-sm text-gray-500">
            <div>
              <span>KubeChat Admin Console</span>
              <span className="mx-2">•</span>
              <span>Enterprise Security Management</span>
            </div>
            <div>
              <span>SOC 2 Type II Compliant</span>
              <span className="mx-2">•</span>
              <span>GDPR Protected</span>
            </div>
          </div>
        </footer>
      </div>
    </div>
  );
};

export default AdminLayout;
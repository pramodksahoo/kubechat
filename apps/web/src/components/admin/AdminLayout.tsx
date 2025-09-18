// Admin Layout Component
// Following coding standards from docs/architecture/coding-standards.md
// AC 1: Admin Route Creation with admin-only access

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import Link from 'next/link';
import { useAuthStore } from '../../stores/authStore';
import { useAdminStore } from '../../stores/adminStore';

export interface AdminLayoutProps {
  children: React.ReactNode;
  title?: string;
}

const AdminNavigation: React.FC = () => {
  const router = useRouter();
  const { setCurrentView } = useAdminStore();

  const navItems = [
    { key: 'dashboard', label: 'Dashboard', href: '/admin', icon: '📊' },
    { key: 'users', label: 'User Management', href: '/admin/users', icon: '👥' },
    { key: 'roles', label: 'Role Management', href: '/admin/roles', icon: '🔐' },
    { key: 'sessions', label: 'Active Sessions', href: '/admin/sessions', icon: '🔗' },
    { key: 'security', label: 'Security Settings', href: '/admin/security', icon: '🛡️' },
    { key: 'audit', label: 'Audit Logs', href: '/admin/audit', icon: '📋' },
    { key: 'credentials', label: 'Admin Credentials', href: '/admin/credentials', icon: '🔑' }
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
                <span className="text-lg">{item.icon}</span>
                <span className="font-medium">{item.label}</span>
              </Link>
            </li>
          );
        })}
      </ul>

      <div className="mt-8 pt-8 border-t border-gray-700">
        <div className="text-xs text-gray-400 space-y-1">
          <p>🔒 Admin Access Only</p>
          <p>📋 SOC 2 Compliant</p>
          <p>🛡️ GDPR Protected</p>
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
            <span className="text-yellow-400">⚠️</span>
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
            ×
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
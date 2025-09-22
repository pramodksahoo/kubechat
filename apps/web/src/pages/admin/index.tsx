// Admin Dashboard Page
// Following coding standards from docs/architecture/coding-standards.md
// AC 1: Admin Route Creation with dashboard and overview

import React, { useEffect, useState } from 'react';
import { AdminLayout } from '../../components/admin/AdminLayout';
import { useAdminStore } from '../../stores/adminStore';
import { initializeAdminStore } from '../../stores/adminStore';

interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  adminUsers: number;
  activeSessions: number;
  recentAuditLogs: number;
  complianceStatus: string;
  credentialSyncStatus: string;
}

const StatsCard: React.FC<{
  title: string;
  value: string | number;
  icon: string;
  status?: 'success' | 'warning' | 'error';
  description?: string;
}> = ({ title, value, icon, status, description }) => {
  const statusColors = {
    success: 'text-green-600 bg-green-100',
    warning: 'text-yellow-600 bg-yellow-100',
    error: 'text-red-600 bg-red-100'
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-3xl font-bold text-gray-900 mt-2">{value}</p>
          {description && (
            <p className="text-sm text-gray-500 mt-1">{description}</p>
          )}
        </div>
        <div className={`p-3 rounded-full ${status ? statusColors[status] : 'text-blue-600 bg-blue-100'}`}>
          <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" />
          </svg>
        </div>
      </div>
    </div>
  );
};

const RecentActivity: React.FC = () => {
  const { auditLogs, loadAuditLogs, isLoading } = useAdminStore();
  const recentLogs = auditLogs.logs.slice(0, 5);

  useEffect(() => {
    loadAuditLogs().catch(console.error);
  }, [loadAuditLogs]);

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Recent Activity</h3>
        <div className="animate-pulse space-y-3">
          {[1, 2, 3].map(i => (
            <div key={i} className="h-4 bg-gray-200 rounded"></div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-medium text-gray-900">Recent Activity</h3>
        <a href="/admin/audit" className="text-sm text-blue-600 hover:text-blue-800">
          View all →
        </a>
      </div>

      <div className="space-y-3">
        {recentLogs.length > 0 ? recentLogs.map((log) => (
          <div key={log.id} className="flex items-center space-x-3 p-3 bg-gray-50 rounded-lg">
            <div className={`w-2 h-2 rounded-full ${
              log.success ? 'bg-green-500' : 'bg-red-500'
            }`}></div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-gray-900 truncate">
                {log.action}
              </p>
              <p className="text-xs text-gray-500">
                {log.username || 'System'} • {new Date(log.timestamp).toLocaleString()}
              </p>
            </div>
            <div className={`px-2 py-1 text-xs rounded-full ${
              log.riskLevel === 'high' ? 'bg-red-100 text-red-800' :
              log.riskLevel === 'medium' ? 'bg-yellow-100 text-yellow-800' :
              'bg-green-100 text-green-800'
            }`}>
              {log.riskLevel}
            </div>
          </div>
        )) : (
          <p className="text-sm text-gray-500">No recent activity</p>
        )}
      </div>
    </div>
  );
};

const QuickActions: React.FC = () => {
  const actions = [
    { title: 'Create User', href: '/admin/users/create', description: 'Add new user account' },
    { title: 'Manage Roles', href: '/admin/roles', description: 'Configure user roles' },
    { title: 'View Sessions', href: '/admin/sessions', description: 'Monitor active sessions' },
    { title: 'Security Settings', href: '/admin/security', description: 'Configure security policies' },
    { title: 'Audit Logs', href: '/admin/audit', description: 'Review audit trail' },
    { title: 'Admin Credentials', href: '/admin/credentials', description: 'Manage admin access' }
  ];

  const getActionIcon = (title: string) => {
    switch (title) {
      case 'Create User':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        );
      case 'Manage Roles':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        );
      case 'View Sessions':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
        );
      case 'Security Settings':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.618 5.984A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622a12.02 12.02 0 00.382-3.016zM12 9v2m0 4h.01" />
          </svg>
        );
      case 'Audit Logs':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        );
      case 'Admin Credentials':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
          </svg>
        );
      default:
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
        );
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h3>
      <div className="grid grid-cols-2 gap-3">
        {actions.map((action) => (
          <a
            key={action.title}
            href={action.href}
            className="flex items-center space-x-3 p-3 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
          >
            {getActionIcon(action.title)}
            <div>
              <p className="text-sm font-medium text-gray-900">{action.title}</p>
              <p className="text-xs text-gray-500">{action.description}</p>
            </div>
          </a>
        ))}
      </div>
    </div>
  );
};

const ChatPreview: React.FC = () => {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-medium text-gray-900">Admin Chat Interface</h3>
        <a 
          href="/admin/chat"
          className="text-blue-600 hover:text-blue-800 text-sm font-medium"
        >
          Open Full Chat →
        </a>
      </div>
      
      <div className="space-y-3">
        <div className="bg-gray-50 rounded-lg p-3">
          <div className="flex items-center space-x-2 mb-2">
            <svg className="w-4 h-4 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" clipRule="evenodd" />
            </svg>
            <span className="text-sm font-medium text-gray-900">Quick Admin Commands</span>
          </div>
          <div className="space-y-1 text-sm text-gray-600">
            <p>• Check cluster status</p>
            <p>• List failing pods</p>
            <p>• Review security policies</p>
            <p>• Monitor resource usage</p>
          </div>
        </div>
        
        <div className="flex items-center justify-between p-3 bg-blue-50 rounded-lg">
          <div className="flex items-center space-x-2">
            <svg className="w-5 h-5 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 1L3 6v4c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V6l-7-5z" clipRule="evenodd" />
            </svg>
            <span className="text-sm font-medium text-blue-900">Admin Mode</span>
          </div>
          <span className="text-xs text-blue-700">Enhanced Privileges</span>
        </div>
        
        <a
          href="/admin/chat"
          className="block w-full text-center py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
        >
          Start Admin Chat Session
        </a>
      </div>
    </div>
  );
};

const AdminDashboard: React.FC = () => {
  const {
    users,
    activeSessions,
    auditLogs,
    adminCredentials,
    isLoading,
    error
  } = useAdminStore();

  const [stats, setStats] = useState<DashboardStats>({
    totalUsers: 0,
    activeUsers: 0,
    adminUsers: 0,
    activeSessions: 0,
    recentAuditLogs: 0,
    complianceStatus: 'pending',
    credentialSyncStatus: 'pending'
  });

  useEffect(() => {
    // Initialize admin store with data
    initializeAdminStore();
  }, []);

  useEffect(() => {
    // Update stats when data changes
    setStats({
      totalUsers: users.length,
      activeUsers: users.filter(u => u.isActive).length,
      adminUsers: users.filter(u => u.role === 'admin').length,
      activeSessions: activeSessions.length,
      recentAuditLogs: auditLogs.logs.length,
      complianceStatus: adminCredentials.complianceStatus,
      credentialSyncStatus: adminCredentials.syncStatus
    });
  }, [users, activeSessions, auditLogs, adminCredentials]);

  const getStatusFromString = (status: string) => {
    if (status.includes('compliant') || status.includes('success') || status.includes('synced')) {
      return 'success';
    }
    if (status.includes('warning') || status.includes('pending')) {
      return 'warning';
    }
    return 'error';
  };

  return (
    <AdminLayout title="Admin Dashboard">
      <div className="space-y-6">
        {/* Error Display */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <span className="text-red-400">⚠️</span>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Error</h3>
                <p className="mt-1 text-sm text-red-700">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Welcome Section */}
        <div className="bg-gradient-to-r from-blue-600 to-blue-800 rounded-lg p-6 text-white">
          <h1 className="text-2xl font-bold mb-2">Welcome to KubeChat Admin</h1>
          <p className="text-blue-100">
            Enterprise user management and security administration console.
            Monitor users, sessions, and maintain compliance with enterprise security standards.
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <StatsCard
            title="Total Users"
            value={stats.totalUsers}
            icon="👥"
            description={`${stats.activeUsers} active`}
          />
          <StatsCard
            title="Admin Users"
            value={stats.adminUsers}
            icon="🔐"
            status={stats.adminUsers > 0 ? 'success' : 'warning'}
          />
          <StatsCard
            title="Active Sessions"
            value={stats.activeSessions}
            icon="🔗"
            status={stats.activeSessions < 50 ? 'success' : 'warning'}
          />
          <StatsCard
            title="Compliance Status"
            value={stats.complianceStatus}
            icon="✅"
            status={getStatusFromString(stats.complianceStatus)}
          />
        </div>

        {/* Secondary Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <StatsCard
            title="Credential Sync"
            value={stats.credentialSyncStatus}
            icon="🔄"
            status={getStatusFromString(stats.credentialSyncStatus)}
            description="Admin credential management"
          />
          <StatsCard
            title="Audit Entries"
            value={stats.recentAuditLogs}
            icon="📋"
            description="Recent activity logs"
          />
        </div>

        {/* Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-1">
            <RecentActivity />
          </div>
          <div className="lg:col-span-1">
            <QuickActions />
          </div>
          <div className="lg:col-span-1">
            <ChatPreview />
          </div>
        </div>

        {/* Security Notice */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
          <div className="flex items-start space-x-3">
            <svg className="w-6 h-6 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 1L3 6v4c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V6l-7-5z" clipRule="evenodd" />
            </svg>
            <div>
              <h3 className="text-lg font-medium text-blue-900 mb-2">
                Enterprise Security Compliance
              </h3>
              <div className="text-sm text-blue-700 space-y-1">
                <div className="flex items-center space-x-2">
                  <svg className="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span>SOC 2 Type II compliant audit logging</span>
                </div>
                <div className="flex items-center space-x-2">
                  <svg className="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span>GDPR-compliant data handling</span>
                </div>
                <div className="flex items-center space-x-2">
                  <svg className="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span>Enterprise password policies enforced</span>
                </div>
                <div className="flex items-center space-x-2">
                  <svg className="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span>Multi-factor authentication supported</span>
                </div>
                <div className="flex items-center space-x-2">
                  <svg className="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span>Real-time session monitoring</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AdminLayout>
  );
};

export default AdminDashboard;
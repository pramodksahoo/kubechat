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
          <span className="text-2xl">{icon}</span>
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
    { title: 'Create User', href: '/admin/users/create', icon: '👤', description: 'Add new user account' },
    { title: 'Manage Roles', href: '/admin/roles', icon: '🔐', description: 'Configure user roles' },
    { title: 'View Sessions', href: '/admin/sessions', icon: '🔗', description: 'Monitor active sessions' },
    { title: 'Security Settings', href: '/admin/security', icon: '🛡️', description: 'Configure security policies' },
    { title: 'Audit Logs', href: '/admin/audit', icon: '📋', description: 'Review audit trail' },
    { title: 'Admin Credentials', href: '/admin/credentials', icon: '🔑', description: 'Manage admin access' }
  ];

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
            <span className="text-xl">{action.icon}</span>
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
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <RecentActivity />
          <QuickActions />
        </div>

        {/* Security Notice */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
          <div className="flex items-start space-x-3">
            <span className="text-blue-500 text-xl">🛡️</span>
            <div>
              <h3 className="text-lg font-medium text-blue-900 mb-2">
                Enterprise Security Compliance
              </h3>
              <div className="text-sm text-blue-700 space-y-1">
                <p>✅ SOC 2 Type II compliant audit logging</p>
                <p>✅ GDPR-compliant data handling</p>
                <p>✅ Enterprise password policies enforced</p>
                <p>✅ Multi-factor authentication supported</p>
                <p>✅ Real-time session monitoring</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AdminLayout>
  );
};

export default AdminDashboard;
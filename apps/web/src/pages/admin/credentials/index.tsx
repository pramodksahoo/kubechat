// Admin Credentials Management Page
// Provides visibility into admin credential sync status (AC 8)

import React, { useEffect } from 'react';
import { AdminLayout } from '../../../components/admin/AdminLayout';
import { useAdminStore } from '../../../stores/adminStore';

const AdminCredentialsPage: React.FC = () => {
  const {
    adminCredentials,
    syncAdminCredentials,
    isLoading,
    error
  } = useAdminStore();

  useEffect(() => {
    syncAdminCredentials().catch((err) => {
      console.error('Failed to sync admin credentials:', err);
    });
  }, [syncAdminCredentials]);

  const renderStatusBadge = (status: string | undefined, fallback: string) => {
    const value = (status || fallback).toLowerCase();
    let classes = 'bg-gray-100 text-gray-700';

    if (value.includes('success') || value.includes('synced') || value.includes('compliant')) {
      classes = 'bg-green-100 text-green-800';
    } else if (value.includes('pending') || value.includes('warning')) {
      classes = 'bg-yellow-100 text-yellow-800';
    } else if (value.includes('error') || value.includes('out_of_sync')) {
      classes = 'bg-red-100 text-red-800';
    }

    return (
      <span className={`rounded-full px-3 py-1 text-xs font-semibold ${classes}`}>
        {status || fallback}
      </span>
    );
  };

  return (
    <AdminLayout title="Admin Credential Management">
      <div className="space-y-6">
        <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Admin Credential Sync</h1>
              <p className="text-gray-600">
                Monitor Kubernetes Secret ↔ database synchronization health and rotation posture.
              </p>
            </div>
            <button
              onClick={() => syncAdminCredentials().catch(console.error)}
              className="inline-flex items-center rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              disabled={isLoading}
            >
              {isLoading ? 'Syncing...' : 'Refresh Status'}
            </button>
          </div>
        </div>

        {error && (
          <div className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900">Current Status</h2>
            <dl className="mt-4 space-y-3 text-sm text-gray-700">
              <div className="flex items-center justify-between">
                <dt>Sync Status</dt>
                <dd>{renderStatusBadge(adminCredentials.syncStatus, 'unknown')}</dd>
              </div>
              <div className="flex items-center justify-between">
                <dt>Compliance</dt>
                <dd>{renderStatusBadge(adminCredentials.complianceStatus, 'pending_review')}</dd>
              </div>
              <div className="flex items-center justify-between">
                <dt>Last Sync</dt>
                <dd>{adminCredentials.lastSync || 'Not available'}</dd>
              </div>
              <div className="flex items-center justify-between">
                <dt>Password Expiry</dt>
                <dd>{adminCredentials.passwordExpiry || 'Not available'}</dd>
              </div>
              <div className="flex items-center justify-between">
                <dt>Rotation Count</dt>
                <dd>{adminCredentials.rotationCount}</dd>
              </div>
            </dl>
          </div>

          <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900">Rotation Guidance</h2>
            <p className="mt-2 text-sm text-gray-600">
              Use automation or manual secret updates to rotate admin credentials while ensuring
              auditability and compliance with 90-day policies.
            </p>
            <ul className="mt-4 list-disc space-y-2 pl-6 text-sm text-gray-700">
              <li>Generate passwords ≥ 24 characters with mixed-case letters, numbers, and symbols.</li>
              <li>Rotate credentials every 90 days or immediately after suspected compromise.</li>
              <li>Log all credential retrievals for audit readiness.</li>
            </ul>
          </div>
        </div>
      </div>
    </AdminLayout>
  );
};

export default AdminCredentialsPage;

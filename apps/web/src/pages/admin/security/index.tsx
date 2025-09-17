// Admin Security Administration Panel
// Following coding standards from docs/architecture/coding-standards.md
// AC 6: Security Administration Panel

import React, { useEffect, useState } from 'react';
import { AdminLayout } from '../../../components/admin/AdminLayout';
import { useAdminStore } from '../../../stores/adminStore';
import type { SecurityPolicy } from '../../../types/admin';

export default function AdminSecurityPage() {
  const {
    securityPolicies,
    adminCredentials,
    isLoading,
    error,
    loadSecurityPolicies,
    updateSecurityPolicy,
    syncAdminCredentials,
    rotateAdminPassword,
    validateCredentialCompliance,
    requestEmergencyAccess,
    setError
  } = useAdminStore();

  const [activeTab, setActiveTab] = useState<'policies' | 'credentials' | 'emergency' | 'compliance'>('policies');
  const [editingPolicy, setEditingPolicy] = useState<SecurityPolicy | null>(null);
  const [emergencyForm, setEmergencyForm] = useState({
    justification: '',
    businessImpact: ''
  });
  const [rotationPassword, setRotationPassword] = useState('');
  const [complianceResults, setComplianceResults] = useState<any>(null);

  useEffect(() => {
    loadSecurityPolicies();
  }, [loadSecurityPolicies]);

  const handleUpdatePolicy = async (policy: SecurityPolicy, updates: Partial<SecurityPolicy>) => {
    try {
      await updateSecurityPolicy(policy.id, updates);
      setEditingPolicy(null);
    } catch (err) {
      console.error('Failed to update security policy:', err);
    }
  };

  const handleCredentialSync = async () => {
    try {
      await syncAdminCredentials();
    } catch (err) {
      console.error('Failed to sync admin credentials:', err);
    }
  };

  const handlePasswordRotation = async () => {
    try {
      await rotateAdminPassword(rotationPassword || undefined);
      setRotationPassword('');
    } catch (err) {
      console.error('Failed to rotate admin password:', err);
    }
  };

  const handleComplianceValidation = async () => {
    try {
      const results = await validateCredentialCompliance();
      setComplianceResults(results);
    } catch (err) {
      console.error('Failed to validate compliance:', err);
    }
  };

  const handleEmergencyAccess = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await requestEmergencyAccess(emergencyForm.justification, emergencyForm.businessImpact);
      setEmergencyForm({ justification: '', businessImpact: '' });
    } catch (err) {
      console.error('Failed to request emergency access:', err);
    }
  };

  const getSyncStatusColor = (status: string) => {
    switch (status) {
      case 'synced': return 'text-green-800 bg-green-100';
      case 'out_of_sync': return 'text-yellow-800 bg-yellow-100';
      case 'rotation_pending': return 'text-blue-800 bg-blue-100';
      case 'error': return 'text-red-800 bg-red-100';
      default: return 'text-gray-800 bg-gray-100';
    }
  };

  if (error) {
    return (
      <AdminLayout title="Security Administration">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error loading security data</h3>
              <p className="text-sm text-red-700 mt-1">{error}</p>
              <button
                onClick={() => {
                  setError(null);
                  loadSecurityPolicies();
                }}
                className="mt-2 text-sm bg-red-100 text-red-800 px-3 py-1 rounded hover:bg-red-200"
              >
                Try Again
              </button>
            </div>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout title="Security Administration">
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Security Administration</h1>
          <p className="text-gray-600">Manage security policies, credentials, and compliance</p>
        </div>

        {/* Security Status Overview */}
        <div className="bg-gradient-to-r from-red-50 to-yellow-50 border-l-4 border-orange-400 p-4 rounded-md">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-orange-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-orange-800">Security Configuration Required</h3>
              <p className="text-sm text-orange-700 mt-1">
                Some security policies require immediate attention. Review and update configurations as needed.
              </p>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            {[
              { id: 'policies', name: 'Security Policies', icon: '🛡️' },
              { id: 'credentials', name: 'Credential Management', icon: '🔑' },
              { id: 'emergency', name: 'Emergency Access', icon: '🚨' },
              { id: 'compliance', name: 'Compliance', icon: '📋' }
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                } whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm flex items-center space-x-2`}
              >
                <span>{tab.icon}</span>
                <span>{tab.name}</span>
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        <div className="bg-white shadow rounded-lg">
          {/* Security Policies Tab */}
          {activeTab === 'policies' && (
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-lg font-medium text-gray-900">Security Policies</h2>
                <button
                  onClick={loadSecurityPolicies}
                  disabled={isLoading}
                  className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                >
                  {isLoading ? 'Loading...' : 'Refresh'}
                </button>
              </div>

              <div className="space-y-4">
                {securityPolicies.map((policy) => (
                  <div key={policy.id} className="border border-gray-200 rounded-lg p-4">
                    <div className="flex justify-between items-start">
                      <div>
                        <h3 className="text-sm font-medium text-gray-900">{policy.name}</h3>
                        <p className="text-sm text-gray-500 mt-1">{policy.description}</p>
                        <div className="flex items-center mt-2 space-x-4">
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            policy.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                          }`}>
                            {policy.enabled ? 'Enabled' : 'Disabled'}
                          </span>
                          <span className="text-xs text-gray-500">
                            Updated: {new Date(policy.updatedAt).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => setEditingPolicy(policy)}
                          className="text-sm text-blue-600 hover:text-blue-900"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleUpdatePolicy(policy, { enabled: !policy.enabled })}
                          className="text-sm text-gray-600 hover:text-gray-900"
                        >
                          {policy.enabled ? 'Disable' : 'Enable'}
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Credential Management Tab */}
          {activeTab === 'credentials' && (
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-lg font-medium text-gray-900">Admin Credential Management</h2>
                <button
                  onClick={handleCredentialSync}
                  disabled={isLoading}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
                >
                  {isLoading ? 'Syncing...' : 'Sync Credentials'}
                </button>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Credential Status */}
                <div className="space-y-4">
                  <h3 className="text-base font-medium text-gray-900">Current Status</h3>

                  <div className="border border-gray-200 rounded-lg p-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium text-gray-700">Sync Status</span>
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getSyncStatusColor(adminCredentials.syncStatus)}`}>
                        {adminCredentials.syncStatus.replace('_', ' ').toUpperCase()}
                      </span>
                    </div>
                  </div>

                  <div className="border border-gray-200 rounded-lg p-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium text-gray-700">Last Sync</span>
                      <span className="text-sm text-gray-600">
                        {adminCredentials.lastSync ? new Date(adminCredentials.lastSync).toLocaleString() : 'Never'}
                      </span>
                    </div>
                  </div>

                  <div className="border border-gray-200 rounded-lg p-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium text-gray-700">Password Expiry</span>
                      <span className="text-sm text-gray-600">
                        {adminCredentials.passwordExpiry ? new Date(adminCredentials.passwordExpiry).toLocaleDateString() : 'Not set'}
                      </span>
                    </div>
                  </div>

                  <div className="border border-gray-200 rounded-lg p-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium text-gray-700">Rotation Count</span>
                      <span className="text-sm text-gray-600">{adminCredentials.rotationCount}</span>
                    </div>
                  </div>

                  <div className="border border-gray-200 rounded-lg p-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium text-gray-700">Compliance Status</span>
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                        adminCredentials.complianceStatus === 'compliant' ? 'bg-green-100 text-green-800' :
                        adminCredentials.complianceStatus === 'non_compliant' ? 'bg-red-100 text-red-800' :
                        'bg-yellow-100 text-yellow-800'
                      }`}>
                        {adminCredentials.complianceStatus.replace('_', ' ').toUpperCase()}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Password Rotation */}
                <div className="space-y-4">
                  <h3 className="text-base font-medium text-gray-900">Password Rotation</h3>

                  <div className="border border-gray-200 rounded-lg p-4">
                    <div className="space-y-4">
                      <div>
                        <label htmlFor="rotation-password" className="block text-sm font-medium text-gray-700 mb-1">
                          New Password (optional)
                        </label>
                        <input
                          type="password"
                          id="rotation-password"
                          value={rotationPassword}
                          onChange={(e) => setRotationPassword(e.target.value)}
                          placeholder="Leave empty for auto-generated password"
                          className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                      </div>
                      <button
                        onClick={handlePasswordRotation}
                        disabled={isLoading}
                        className="w-full inline-flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700 disabled:opacity-50"
                      >
                        {isLoading ? 'Rotating...' : 'Rotate Password'}
                      </button>
                      <p className="text-xs text-gray-500">
                        This will immediately update the admin password in both the database and K8s Secret.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Emergency Access Tab */}
          {activeTab === 'emergency' && (
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-lg font-medium text-gray-900">Emergency Access Request</h2>
              </div>

              <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4 mb-6">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-yellow-800">Emergency Access</h3>
                    <p className="text-sm text-yellow-700 mt-1">
                      Emergency access requests are audited and require approval. Use only in critical situations.
                    </p>
                  </div>
                </div>
              </div>

              <form onSubmit={handleEmergencyAccess} className="space-y-6">
                <div>
                  <label htmlFor="justification" className="block text-sm font-medium text-gray-700 mb-2">
                    Justification *
                  </label>
                  <textarea
                    id="justification"
                    value={emergencyForm.justification}
                    onChange={(e) => setEmergencyForm({ ...emergencyForm, justification: e.target.value })}
                    required
                    rows={4}
                    placeholder="Provide detailed justification for emergency access request..."
                    className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>

                <div>
                  <label htmlFor="business-impact" className="block text-sm font-medium text-gray-700 mb-2">
                    Business Impact *
                  </label>
                  <textarea
                    id="business-impact"
                    value={emergencyForm.businessImpact}
                    onChange={(e) => setEmergencyForm({ ...emergencyForm, businessImpact: e.target.value })}
                    required
                    rows={4}
                    placeholder="Describe the business impact and urgency..."
                    className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>

                <button
                  type="submit"
                  disabled={isLoading}
                  className="inline-flex justify-center items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-red-600 hover:bg-red-700 disabled:opacity-50"
                >
                  {isLoading ? 'Submitting...' : 'Request Emergency Access'}
                </button>
              </form>
            </div>
          )}

          {/* Compliance Tab */}
          {activeTab === 'compliance' && (
            <div className="p-6">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-lg font-medium text-gray-900">Compliance Validation</h2>
                <button
                  onClick={handleComplianceValidation}
                  disabled={isLoading}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
                >
                  {isLoading ? 'Validating...' : 'Run Compliance Check'}
                </button>
              </div>

              {complianceResults ? (
                <div className="space-y-4">
                  <div className={`border rounded-lg p-4 ${
                    complianceResults.overall === 'compliant' ? 'border-green-200 bg-green-50' :
                    complianceResults.overall === 'non_compliant' ? 'border-red-200 bg-red-50' :
                    'border-yellow-200 bg-yellow-50'
                  }`}>
                    <div className="flex items-center">
                      <div className="flex-shrink-0">
                        {complianceResults.overall === 'compliant' ? (
                          <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                          </svg>
                        ) : complianceResults.overall === 'non_compliant' ? (
                          <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                          </svg>
                        ) : (
                          <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                          </svg>
                        )}
                      </div>
                      <div className="ml-3">
                        <h3 className={`text-sm font-medium ${
                          complianceResults.overall === 'compliant' ? 'text-green-800' :
                          complianceResults.overall === 'non_compliant' ? 'text-red-800' :
                          'text-yellow-800'
                        }`}>
                          Overall Status: {complianceResults.overall.replace('_', ' ').toUpperCase()}
                        </h3>
                        <p className={`text-sm mt-1 ${
                          complianceResults.overall === 'compliant' ? 'text-green-700' :
                          complianceResults.overall === 'non_compliant' ? 'text-red-700' :
                          'text-yellow-700'
                        }`}>
                          Last validated: {complianceResults.lastValidated ? new Date(complianceResults.lastValidated).toLocaleString() : 'Never'}
                        </p>
                      </div>
                    </div>
                  </div>

                  {complianceResults.checks && complianceResults.checks.length > 0 && (
                    <div className="space-y-3">
                      <h3 className="text-base font-medium text-gray-900">Compliance Checks</h3>
                      {complianceResults.checks.map((check: any, index: number) => (
                        <div key={index} className="border border-gray-200 rounded-lg p-4">
                          <div className="flex justify-between items-start">
                            <div>
                              <h4 className="text-sm font-medium text-gray-900">{check.checkName}</h4>
                              <p className="text-sm text-gray-600 mt-1">{check.details}</p>
                              {check.remediation && (
                                <p className="text-xs text-blue-600 mt-2">
                                  Remediation: {check.remediation}
                                </p>
                              )}
                            </div>
                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                              check.status === 'pass' ? 'bg-green-100 text-green-800' :
                              check.status === 'fail' ? 'bg-red-100 text-red-800' :
                              'bg-yellow-100 text-yellow-800'
                            }`}>
                              {check.status.toUpperCase()}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-center py-8">
                  <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <h3 className="mt-2 text-sm font-medium text-gray-900">No compliance data</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    Run a compliance check to view detailed validation results.
                  </p>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Policy Edit Modal */}
        {editingPolicy && (
          <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-11/12 md:w-1/2 shadow-lg rounded-md bg-white">
              <div className="mt-3">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Edit Security Policy</h3>

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Policy Name</label>
                    <input
                      type="text"
                      value={editingPolicy.name}
                      className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm bg-gray-50"
                      readOnly
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea
                      value={editingPolicy.description}
                      className="block w-full border border-gray-300 rounded-md px-3 py-2 text-sm bg-gray-50"
                      rows={3}
                      readOnly
                    />
                  </div>

                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      id="policy-enabled"
                      checked={editingPolicy.enabled}
                      onChange={(e) => setEditingPolicy({ ...editingPolicy, enabled: e.target.checked })}
                      className="rounded border-gray-300"
                    />
                    <label htmlFor="policy-enabled" className="ml-2 text-sm text-gray-900">
                      Enable this security policy
                    </label>
                  </div>
                </div>

                <div className="flex justify-end space-x-3 mt-6">
                  <button
                    onClick={() => setEditingPolicy(null)}
                    className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => handleUpdatePolicy(editingPolicy, { enabled: editingPolicy.enabled })}
                    className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                  >
                    Save Changes
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </AdminLayout>
  );
}
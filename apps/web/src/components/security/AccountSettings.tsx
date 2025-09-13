import { useState } from 'react';
import { UserPreferences } from '../../types/user';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';

interface AccountSettingsProps {
  preferences: UserPreferences;
  onUpdatePreferences: (preferences: UserPreferences) => Promise<void>;
  onExportData: () => Promise<void>;
  onDeleteAccount: (confirmation: string) => Promise<void>;
  onEnableTwoFactor: () => Promise<void>;
  onDisableTwoFactor: () => Promise<void>;
  twoFactorEnabled?: boolean;
  loading?: boolean;
  className?: string;
}

export function AccountSettings({
  preferences,
  onUpdatePreferences,
  onExportData,
  onDeleteAccount,
  onEnableTwoFactor,
  onDisableTwoFactor,
  twoFactorEnabled = false,
  loading = false,
  className = '',
}: AccountSettingsProps) {
  const [localPreferences, setLocalPreferences] = useState<UserPreferences>(preferences);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [deleteConfirmation, setDeleteConfirmation] = useState('');
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  const updatePreference = (updates: Partial<UserPreferences>) => {
    setLocalPreferences(prev => ({ ...prev, ...updates }));
    setHasUnsavedChanges(true);
  };

  const updateNestedPreference = <T extends keyof UserPreferences>(
    category: T,
    updates: Partial<UserPreferences[T]>
  ) => {
    setLocalPreferences(prev => ({
      ...prev,
      [category]: { ...(prev[category] as Record<string, unknown> || {}), ...updates }
    }));
    setHasUnsavedChanges(true);
  };

  const handleSave = async () => {
    try {
      await onUpdatePreferences(localPreferences);
      setHasUnsavedChanges(false);
    } catch (error) {
      console.error('Failed to save preferences:', error);
    }
  };

  const handleReset = () => {
    setLocalPreferences(preferences);
    setHasUnsavedChanges(false);
  };

  const handleDeleteAccount = async () => {
    try {
      await onDeleteAccount(deleteConfirmation);
      setShowDeleteModal(false);
    } catch (error) {
      console.error('Failed to delete account:', error);
    }
  };

  // const getRefreshIntervalLabel = (interval: number) => {
  //   if (interval < 60) return `${interval} seconds`;
  //   if (interval < 3600) return `${interval / 60} minutes`;
  //   return `${interval / 3600} hours`;
  // };

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Navigation and Theme Settings */}
      <Card>
        <div className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">
            Appearance & Interface
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Theme Preference
              </label>
              <select
                value={localPreferences.theme.mode}
                onChange={(e) => updateNestedPreference('theme', { mode: e.target.value as 'light' | 'dark' | 'system' })}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                <option value="light">Light Mode</option>
                <option value="dark">Dark Mode</option>
                <option value="system">Follow System</option>
              </select>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Choose your preferred color scheme
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Dashboard Layout
              </label>
              <select
                value={localPreferences.dashboard.defaultView}
                onChange={(e) => updateNestedPreference('dashboard', { defaultView: e.target.value as 'grid' | 'list' })}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                <option value="grid">Grid</option>
                <option value="list">List</option>
              </select>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Adjust information density
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Default Cluster
              </label>
              <Input
                value={localPreferences.dashboard.defaultCluster || ''}
                onChange={(value) => updateNestedPreference('dashboard', { defaultCluster: value || undefined })}
                placeholder="Select default cluster"
              />
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Automatically connect to this cluster
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Refresh Interval
              </label>
              <select
                value={localPreferences.dashboard.refreshInterval}
                onChange={(e) => updateNestedPreference('dashboard', { refreshInterval: parseInt(e.target.value) })}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                <option value={10}>10 seconds</option>
                <option value={30}>30 seconds</option>
                <option value={60}>1 minute</option>
                <option value={300}>5 minutes</option>
                <option value={900}>15 minutes</option>
              </select>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                How often to refresh dashboard data
              </p>
            </div>
          </div>
        </div>
      </Card>

      {/* Notification Settings */}
      <Card>
        <div className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">
            Notification Preferences
          </h2>

          <div className="space-y-6">
            <div>
              <h3 className="text-base font-medium text-gray-900 dark:text-white mb-4">
                Delivery Methods
              </h3>
              <div className="space-y-4">
                {[
                  { key: 'email', label: 'Email Notifications', description: 'Receive notifications via email' },
                  { key: 'browser', label: 'Browser Notifications', description: 'Show browser push notifications' },
                  { key: 'mobile', label: 'Mobile Notifications', description: 'Send notifications to mobile app' },
                ].map(({ key, label, description }) => (
                  <div key={key} className="flex items-center justify-between">
                    <div>
                      <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{label}</span>
                      <p className="text-xs text-gray-500 dark:text-gray-400">{description}</p>
                    </div>
                    <button
                      onClick={() => updateNestedPreference('notifications', {
                        [key]: !localPreferences.notifications[key as keyof typeof localPreferences.notifications]
                      })}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        localPreferences.notifications[key as keyof typeof localPreferences.notifications]
                          ? 'bg-blue-600'
                          : 'bg-gray-200 dark:bg-gray-700'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          localPreferences.notifications[key as keyof typeof localPreferences.notifications]
                            ? 'translate-x-6'
                            : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <h3 className="text-base font-medium text-gray-900 dark:text-white mb-4">
                Notification Types
              </h3>
              <div className="space-y-4">
                {[
                  { key: 'commandApprovals', label: 'Command Approvals', description: 'When commands require your approval' },
                  { key: 'systemAlerts', label: 'System Alerts', description: 'Critical system events and errors' },
                  { key: 'auditAlerts', label: 'Audit Alerts', description: 'Security and compliance events' },
                ].map(({ key, label, description }) => (
                  <div key={key} className="flex items-center justify-between">
                    <div>
                      <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{label}</span>
                      <p className="text-xs text-gray-500 dark:text-gray-400">{description}</p>
                    </div>
                    <button
                      onClick={() => updateNestedPreference('notifications', {
                        [key]: !localPreferences.notifications[key as keyof typeof localPreferences.notifications]
                      })}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        localPreferences.notifications[key as keyof typeof localPreferences.notifications]
                          ? 'bg-blue-600'
                          : 'bg-gray-200 dark:bg-gray-700'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          localPreferences.notifications[key as keyof typeof localPreferences.notifications]
                            ? 'translate-x-6'
                            : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Localization Settings */}
      <Card>
        <div className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">
            Localization
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Language
              </label>
              <select
                value={localPreferences.language}
                onChange={(e) => updatePreference({ language: e.target.value })}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                <option value="en">English</option>
                <option value="es">Español</option>
                <option value="fr">Français</option>
                <option value="de">Deutsch</option>
                <option value="ja">日本語</option>
                <option value="zh">中文</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Timezone
              </label>
              <select
                value={localPreferences.timezone}
                onChange={(e) => updatePreference({ timezone: e.target.value })}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                <option value="UTC">UTC</option>
                <option value="America/New_York">Eastern Time (UTC-5/-4)</option>
                <option value="America/Chicago">Central Time (UTC-6/-5)</option>
                <option value="America/Denver">Mountain Time (UTC-7/-6)</option>
                <option value="America/Los_Angeles">Pacific Time (UTC-8/-7)</option>
                <option value="Europe/London">London (UTC+0/+1)</option>
                <option value="Europe/Berlin">Berlin (UTC+1/+2)</option>
                <option value="Europe/Paris">Paris (UTC+1/+2)</option>
                <option value="Asia/Tokyo">Tokyo (UTC+9)</option>
                <option value="Asia/Shanghai">Shanghai (UTC+8)</option>
                <option value="Australia/Sydney">Sydney (UTC+10/+11)</option>
              </select>
            </div>
          </div>
        </div>
      </Card>

      {/* Security Settings */}
      <Card>
        <div className="p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">
            Security & Privacy
          </h2>

          <div className="space-y-6">
            <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
              <div>
                <h3 className="font-medium text-gray-900 dark:text-white">Two-Factor Authentication</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Add an extra layer of security to your account
                </p>
              </div>
              <div className="flex items-center space-x-3">
                <span className={`text-sm font-medium ${
                  twoFactorEnabled 
                    ? 'text-green-600 dark:text-green-400' 
                    : 'text-gray-500 dark:text-gray-400'
                }`}>
                  {twoFactorEnabled ? 'Enabled' : 'Disabled'}
                </span>
                <Button
                  onClick={twoFactorEnabled ? onDisableTwoFactor : onEnableTwoFactor}
                  variant={twoFactorEnabled ? 'danger' : 'primary'}
                  size="sm"
                >
                  {twoFactorEnabled ? 'Disable' : 'Enable'}
                </Button>
              </div>
            </div>

            <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
              <div>
                <h3 className="font-medium text-gray-900 dark:text-white">Export Personal Data</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Download a copy of your personal data and account information
                </p>
              </div>
              <Button
                onClick={() => setShowExportModal(true)}
                variant="secondary"
                size="sm"
              >
                Export Data
              </Button>
            </div>

            <div className="flex items-center justify-between p-4 border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 rounded-lg">
              <div>
                <h3 className="font-medium text-red-900 dark:text-red-200">Delete Account</h3>
                <p className="text-sm text-red-700 dark:text-red-300">
                  Permanently delete your account and all associated data
                </p>
              </div>
              <Button
                onClick={() => setShowDeleteModal(true)}
                variant="danger"
                size="sm"
              >
                Delete Account
              </Button>
            </div>
          </div>
        </div>
      </Card>

      {/* Save Actions */}
      {hasUnsavedChanges && (
        <Card className="border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-900/20">
          <div className="p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <svg className="w-5 h-5 text-yellow-600 dark:text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                <span className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                  You have unsaved changes
                </span>
              </div>
              <div className="flex space-x-3">
                <Button
                  onClick={handleReset}
                  variant="secondary"
                  size="sm"
                >
                  Reset
                </Button>
                <Button
                  onClick={handleSave}
                  variant="primary"
                  size="sm"
                  loading={loading}
                >
                  Save Changes
                </Button>
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* Export Data Modal */}
      <Modal
        isOpen={showExportModal}
        onClose={() => setShowExportModal(false)}
        title="Export Personal Data"
      >
        <div className="space-y-4">
          <p className="text-gray-700 dark:text-gray-300">
            This will download a ZIP file containing all your personal data including:
          </p>
          
          <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1 ml-4">
            <li>• Profile information and preferences</li>
            <li>• Chat history and command logs</li>
            <li>• Security events and audit logs</li>
            <li>• Account activity and session data</li>
          </ul>

          <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
            <p className="text-sm text-blue-700 dark:text-blue-300">
              <strong>Note:</strong> The export process may take a few minutes to complete.
              You&apos;ll receive an email when your data is ready for download.
            </p>
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={async () => {
                await onExportData();
                setShowExportModal(false);
              }}
              variant="primary"
              className="flex-1"
            >
              Request Export
            </Button>
            <Button
              onClick={() => setShowExportModal(false)}
              variant="secondary"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Account Modal */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Delete Account"
      >
        <div className="space-y-4">
          <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
            <div className="flex items-start space-x-3">
              <svg className="w-6 h-6 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
              </svg>
              <div>
                <h4 className="font-medium text-red-900 dark:text-red-200">
                  This action cannot be undone
                </h4>
                <p className="text-sm text-red-700 dark:text-red-300 mt-1">
                  Deleting your account will permanently remove all your data, including chat history, 
                  preferences, and access to all clusters. This action is irreversible.
                </p>
              </div>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Type &quot;DELETE&quot; to confirm:
            </label>
            <Input
              value={deleteConfirmation}
              onChange={(value) => setDeleteConfirmation(value)}
              placeholder="DELETE"
              className="font-mono"
            />
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleDeleteAccount}
              variant="danger"
              className="flex-1"
              disabled={deleteConfirmation !== 'DELETE'}
            >
              Delete Account
            </Button>
            <Button
              onClick={() => setShowDeleteModal(false)}
              variant="secondary"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
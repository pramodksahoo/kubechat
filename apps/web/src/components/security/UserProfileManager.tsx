import { useState, useEffect } from 'react';
import Image from 'next/image';
import { User, UserPreferences } from '../../types/user';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
// import { formatDistanceToNow } from 'date-fns';

interface UserProfileManagerProps {
  user: User;
  onUpdateProfile: (updates: Partial<User>) => Promise<void>;
  onUpdatePreferences: (preferences: UserPreferences) => Promise<void>;
  onChangePassword: (currentPassword: string, newPassword: string) => Promise<void>;
  onUploadAvatar: (file: File) => Promise<string>;
  loading?: boolean;
  className?: string;
}

export function UserProfileManager({
  user,
  onUpdateProfile,
  onUpdatePreferences,
  onChangePassword,
  onUploadAvatar,
  loading = false,
  className = '',
}: UserProfileManagerProps) {
  const [activeTab, setActiveTab] = useState<'profile' | 'preferences' | 'security' | 'clusters'>('profile');
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [profileData, setProfileData] = useState({
    firstName: user.firstName || '',
    lastName: user.lastName || '',
    email: user.email,
  });
  const [preferences, setPreferences] = useState<UserPreferences>(user.preferences);
  const [passwordData, setPasswordData] = useState({
    current: '',
    new: '',
    confirm: '',
  });
  const [avatarUploading, setAvatarUploading] = useState(false);

  useEffect(() => {
    setProfileData({
      firstName: user.firstName || '',
      lastName: user.lastName || '',
      email: user.email,
    });
    setPreferences(user.preferences);
  }, [user]);

  const handleProfileSave = async () => {
    try {
      await onUpdateProfile(profileData);
    } catch (error) {
      console.error('Failed to update profile:', error);
    }
  };

  const handlePreferencesSave = async () => {
    try {
      await onUpdatePreferences(preferences);
    } catch (error) {
      console.error('Failed to update preferences:', error);
    }
  };

  const handlePasswordChange = async () => {
    if (passwordData.new !== passwordData.confirm) {
      alert('New passwords do not match');
      return;
    }

    try {
      await onChangePassword(passwordData.current, passwordData.new);
      setShowPasswordModal(false);
      setPasswordData({ current: '', new: '', confirm: '' });
    } catch (error) {
      console.error('Failed to change password:', error);
    }
  };

  const handleAvatarUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Basic validation
    if (!file.type.startsWith('image/')) {
      alert('Please select an image file');
      return;
    }

    if (file.size > 5 * 1024 * 1024) { // 5MB limit
      alert('Image file must be smaller than 5MB');
      return;
    }

    try {
      setAvatarUploading(true);
      const avatarUrl = await onUploadAvatar(file);
      await onUpdateProfile({ avatar: avatarUrl });
    } catch (error) {
      console.error('Failed to upload avatar:', error);
    } finally {
      setAvatarUploading(false);
    }
  };

  const getRoleNames = () => user.roles.map(role => role.name).join(', ');
  const getPermissionCount = () => user.permissions.length;
  const getClusterCount = () => user.clusters.length;

  const tabs = [
    { id: 'profile', label: 'Profile', icon: '👤' },
    { id: 'preferences', label: 'Preferences', icon: '⚙️' },
    { id: 'security', label: 'Security', icon: '🔒' },
    { id: 'clusters', label: 'Cluster Access', icon: '☸️' },
  ] as const;

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div className="flex items-center space-x-4">
        <div className="relative">
          <div className="w-16 h-16 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center overflow-hidden">
            {user.avatar ? (
              <Image
                src={user.avatar}
                alt={`${user.firstName} ${user.lastName}`}
                width={64}
                height={64}
                className="w-full h-full object-cover"
              />
            ) : (
              <span className="text-2xl font-semibold text-gray-600 dark:text-gray-300">
                {(user.firstName?.[0] || user.username[0]).toUpperCase()}
              </span>
            )}
          </div>
          {avatarUploading && (
            <div className="absolute inset-0 rounded-full bg-black bg-opacity-50 flex items-center justify-center">
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            </div>
          )}
        </div>
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {user.firstName && user.lastName ? `${user.firstName} ${user.lastName}` : user.username}
          </h1>
          <p className="text-gray-600 dark:text-gray-400">@{user.username}</p>
          <div className="flex items-center space-x-4 mt-1">
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {getRoleNames()}
            </span>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {getPermissionCount()} permissions
            </span>
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {getClusterCount()} clusters
            </span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
              }`}
            >
              <span className="mr-2">{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="space-y-6">
        {/* Profile Tab */}
        {activeTab === 'profile' && (
          <Card>
            <div className="p-6 space-y-6">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                  Profile Information
                </h2>
                <label className="cursor-pointer text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 text-sm font-medium">
                  Upload Photo
                  <input
                    type="file"
                    accept="image/*"
                    onChange={handleAvatarUpload}
                    className="hidden"
                    disabled={avatarUploading}
                  />
                </label>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    First Name
                  </label>
                  <Input
                    value={profileData.firstName}
                    onChange={(value) => setProfileData(prev => ({ ...prev, firstName: value }))}
                    placeholder="Enter first name"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Last Name
                  </label>
                  <Input
                    value={profileData.lastName}
                    onChange={(value) => setProfileData(prev => ({ ...prev, lastName: value }))}
                    placeholder="Enter last name"
                  />
                </div>

                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Email Address
                  </label>
                  <Input
                    type="email"
                    value={profileData.email}
                    onChange={(value) => setProfileData(prev => ({ ...prev, email: value }))}
                    placeholder="Enter email address"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Username
                  </label>
                  <Input
                    value={user.username}
                    disabled
                    className="bg-gray-50 dark:bg-gray-800"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    Username cannot be changed
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Account Status
                  </label>
                  <div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    user.isActive 
                      ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                      : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                  }`}>
                    {user.isActive ? 'Active' : 'Inactive'}
                  </div>
                </div>
              </div>

              <div className="flex space-x-3">
                <Button
                  onClick={handleProfileSave}
                  variant="primary"
                  loading={loading}
                >
                  Save Changes
                </Button>
                <Button
                  onClick={() => setShowPasswordModal(true)}
                  variant="secondary"
                >
                  Change Password
                </Button>
              </div>
            </div>
          </Card>
        )}

        {/* Preferences Tab */}
        {activeTab === 'preferences' && (
          <Card>
            <div className="p-6 space-y-6">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                User Preferences
              </h2>

              <div className="space-y-6">
                {/* Appearance */}
                <div>
                  <h3 className="text-base font-medium text-gray-900 dark:text-white mb-4">
                    Appearance
                  </h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Theme
                      </label>
                      <select
                        value={preferences.theme.mode}
                        onChange={(e) => setPreferences(prev => ({ 
                          ...prev, 
                          theme: { ...prev.theme, mode: e.target.value as 'light' | 'dark' | 'system' }
                        }))}
                        className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                      >
                        <option value="light">Light</option>
                        <option value="dark">Dark</option>
                        <option value="system">System</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Dashboard Layout
                      </label>
                      <select
                        value={preferences.dashboard.defaultView}
                        onChange={(e) => setPreferences(prev => ({ 
                          ...prev, 
                          dashboard: { 
                            ...prev.dashboard, 
                            layout: e.target.value as 'compact' | 'comfortable' 
                          }
                        }))}
                        className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                      >
                        <option value="compact">Compact</option>
                        <option value="comfortable">Comfortable</option>
                      </select>
                    </div>
                  </div>
                </div>

                {/* Notifications */}
                <div>
                  <h3 className="text-base font-medium text-gray-900 dark:text-white mb-4">
                    Notifications
                  </h3>
                  <div className="space-y-4">
                    {[
                      { key: 'email', label: 'Email Notifications' },
                      { key: 'browser', label: 'Browser Notifications' },
                      { key: 'commandApprovals', label: 'Command Approval Alerts' },
                      { key: 'systemAlerts', label: 'System Alerts' },
                      { key: 'auditAlerts', label: 'Audit Alerts' },
                    ].map(({ key, label }) => (
                      <div key={key} className="flex items-center justify-between">
                        <span className="text-sm text-gray-700 dark:text-gray-300">{label}</span>
                        <button
                          onClick={() => setPreferences(prev => ({
                            ...prev,
                            notifications: {
                              ...prev.notifications,
                              [key]: !prev.notifications[key as keyof typeof prev.notifications],
                            }
                          }))}
                          className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                            preferences.notifications[key as keyof typeof preferences.notifications]
                              ? 'bg-blue-600'
                              : 'bg-gray-200 dark:bg-gray-700'
                          }`}
                        >
                          <span
                            className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                              preferences.notifications[key as keyof typeof preferences.notifications]
                                ? 'translate-x-6'
                                : 'translate-x-1'
                            }`}
                          />
                        </button>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Localization */}
                <div>
                  <h3 className="text-base font-medium text-gray-900 dark:text-white mb-4">
                    Localization
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Language
                      </label>
                      <select
                        value={preferences.language}
                        onChange={(e) => setPreferences(prev => ({ ...prev, language: e.target.value }))}
                        className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                      >
                        <option value="en">English</option>
                        <option value="es">Spanish</option>
                        <option value="fr">French</option>
                        <option value="de">German</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Timezone
                      </label>
                      <select
                        value={preferences.timezone}
                        onChange={(e) => setPreferences(prev => ({ ...prev, timezone: e.target.value }))}
                        className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
                      >
                        <option value="UTC">UTC</option>
                        <option value="America/New_York">Eastern Time</option>
                        <option value="America/Chicago">Central Time</option>
                        <option value="America/Denver">Mountain Time</option>
                        <option value="America/Los_Angeles">Pacific Time</option>
                        <option value="Europe/London">London</option>
                        <option value="Europe/Berlin">Berlin</option>
                        <option value="Asia/Tokyo">Tokyo</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>

              <Button
                onClick={handlePreferencesSave}
                variant="primary"
                loading={loading}
              >
                Save Preferences
              </Button>
            </div>
          </Card>
        )}

        {/* Security Tab */}
        {activeTab === 'security' && (
          <Card>
            <div className="p-6 space-y-6">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                Security Information
              </h2>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Account Created
                    </label>
                    <p className="text-gray-900 dark:text-white">
                      {new Date().toLocaleDateString()}
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Last Login
                    </label>
                    <p className="text-gray-900 dark:text-white">
                      {user.lastLoginAt 
                        ? new Date(user.lastLoginAt).toLocaleDateString()
                        : 'Never'
                      }
                    </p>
                  </div>
                </div>

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Assigned Roles
                    </label>
                    <div className="space-y-2">
                      {user.roles.map((role) => (
                        <div
                          key={role.id}
                          className="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800 rounded"
                        >
                          <span className="text-sm font-medium text-gray-900 dark:text-white">
                            {role.name}
                          </span>
                          <span className={`text-xs px-2 py-1 rounded ${
                            role.isSystem
                              ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
                              : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                          }`}>
                            {role.isSystem ? 'System' : 'Custom'}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
                <Button
                  onClick={() => setShowPasswordModal(true)}
                  variant="secondary"
                >
                  Change Password
                </Button>
              </div>
            </div>
          </Card>
        )}

        {/* Clusters Tab */}
        {activeTab === 'clusters' && (
          <Card>
            <div className="p-6">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-6">
                Cluster Access
              </h2>

              <div className="space-y-4">
                {user.clusters.map((cluster) => (
                  <div
                    key={cluster.clusterId}
                    className="border border-gray-200 dark:border-gray-700 rounded-lg p-4"
                  >
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="font-medium text-gray-900 dark:text-white">
                        {cluster.clusterName}
                      </h3>
                      <div className="flex space-x-2">
                        {cluster.canExecuteCommands && (
                          <span className="px-2 py-1 text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 rounded">
                            Execute
                          </span>
                        )}
                        {cluster.canApproveCommands && (
                          <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded">
                            Approve
                          </span>
                        )}
                        {cluster.canViewAuditLogs && (
                          <span className="px-2 py-1 text-xs bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200 rounded">
                            Audit
                          </span>
                        )}
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                          Roles
                        </label>
                        <div className="flex flex-wrap gap-1">
                          {cluster.roles.map((role) => (
                            <span
                              key={role.id}
                              className="px-2 py-1 text-xs bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200 rounded"
                            >
                              {role.name}
                            </span>
                          ))}
                        </div>
                      </div>

                      <div>
                        <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                          Permissions ({cluster.permissions.length})
                        </label>
                        <div className="text-xs text-gray-600 dark:text-gray-400">
                          {cluster.permissions.slice(0, 3).map(p => p.name).join(', ')}
                          {cluster.permissions.length > 3 && ` +${cluster.permissions.length - 3} more`}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        )}
      </div>

      {/* Password Change Modal */}
      <Modal
        isOpen={showPasswordModal}
        onClose={() => setShowPasswordModal(false)}
        title="Change Password"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Current Password
            </label>
            <Input
              type="password"
              value={passwordData.current}
              onChange={(value) => setPasswordData(prev => ({ ...prev, current: value }))}
              placeholder="Enter current password"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              New Password
            </label>
            <Input
              type="password"
              value={passwordData.new}
              onChange={(value) => setPasswordData(prev => ({ ...prev, new: value }))}
              placeholder="Enter new password"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Confirm New Password
            </label>
            <Input
              type="password"
              value={passwordData.confirm}
              onChange={(value) => setPasswordData(prev => ({ ...prev, confirm: value }))}
              placeholder="Confirm new password"
            />
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handlePasswordChange}
              variant="primary"
              className="flex-1"
              disabled={!passwordData.current || !passwordData.new || !passwordData.confirm}
            >
              Update Password
            </Button>
            <Button
              onClick={() => setShowPasswordModal(false)}
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
import { useState } from 'react';
import { User, Role, Permission } from '../../types/user';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';

interface PermissionsDisplayProps {
  user: User;
  onRequestPermission?: (permission: Permission, justification: string) => Promise<void>;
  onRequestRole?: (role: Role, justification: string) => Promise<void>;
  showRequestButtons?: boolean;
  className?: string;
}

export function PermissionsDisplay({
  user,
  onRequestPermission,
  onRequestRole,
  showRequestButtons = false,
  className = '',
}: PermissionsDisplayProps) {
  const [activeView, setActiveView] = useState<'overview' | 'roles' | 'permissions' | 'clusters'>('overview');
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [showRequestModal, setShowRequestModal] = useState(false);
  const [requestType, setRequestType] = useState<'permission' | 'role'>('permission');
  const [selectedItem, setSelectedItem] = useState<Permission | Role | null>(null);
  const [justification, setJustification] = useState('');

  const filterPermissions = (permissions: Permission[]) => {
    return permissions.filter(permission => {
      const matchesSearch = searchTerm === '' || 
        permission.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        permission.description?.toLowerCase().includes(searchTerm.toLowerCase());
      
      const matchesCategory = selectedCategory === 'all' || permission.category === selectedCategory;
      
      return matchesSearch && matchesCategory;
    });
  };

  const filterRoles = (roles: Role[]) => {
    return roles.filter(role => {
      return searchTerm === '' || 
        role.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        role.description?.toLowerCase().includes(searchTerm.toLowerCase());
    });
  };

  const getPermissionIcon = (category: string) => {
    switch (category) {
      case 'cluster': return '☸️';
      case 'audit': return '📋';
      case 'user': return '👤';
      case 'system': return '⚙️';
      default: return '🔑';
    }
  };

  const getRoleTypeIcon = (isSystemRole: boolean) => {
    return isSystemRole ? '🛡️' : '👥';
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case 'cluster': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
      case 'audit': return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200';
      case 'user': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'system': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200';
    }
  };

  const getAllPermissions = () => {
    const rolePermissions = user.roles.flatMap(role => role.permissions);
    const directPermissions = user.permissions;
    
    // Remove duplicates
    const permissionMap = new Map();
    [...rolePermissions, ...directPermissions].forEach(permission => {
      permissionMap.set(permission.id, permission);
    });
    
    return Array.from(permissionMap.values());
  };

  const getPermissionSource = (permission: Permission) => {
    const fromRoles = user.roles.filter(role => 
      role.permissions.some(p => p.id === permission.id)
    );
    const fromDirect = user.permissions.some(p => p.id === permission.id);
    
    return { fromRoles, fromDirect };
  };

  const handleRequestAccess = async () => {
    if (!selectedItem) return;

    try {
      if (requestType === 'permission') {
        await onRequestPermission?.(selectedItem as Permission, justification);
      } else {
        await onRequestRole?.(selectedItem as Role, justification);
      }
      setShowRequestModal(false);
      setJustification('');
      setSelectedItem(null);
    } catch (error) {
      console.error('Failed to request access:', error);
    }
  };

  const categories = ['all', 'cluster', 'audit', 'user', 'system'];
  const allPermissions = getAllPermissions();
  const filteredPermissions = filterPermissions(allPermissions);
  const filteredRoles = filterRoles(user.roles);

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header with Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">
              {user.roles.length}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Roles</div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">
              {allPermissions.length}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Permissions</div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">
              {user.clusters.length}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Clusters</div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">
              {user.clusters.filter(c => c.canExecuteCommands).length}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Execute Access</div>
          </div>
        </Card>
      </div>

      {/* Navigation Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'overview', label: 'Overview', icon: '📊' },
            { id: 'roles', label: 'Roles', icon: '👥' },
            { id: 'permissions', label: 'Permissions', icon: '🔑' },
            { id: 'clusters', label: 'Cluster Access', icon: '☸️' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveView(tab.id as 'overview' | 'roles' | 'permissions' | 'clusters')}
              className={`py-2 px-1 border-b-2 font-medium text-sm flex items-center space-x-2 ${
                activeView === tab.id
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
              }`}
            >
              <span>{tab.icon}</span>
              <span>{tab.label}</span>
            </button>
          ))}
        </nav>
      </div>

      {/* Search and Filter */}
      {(activeView === 'permissions' || activeView === 'roles') && (
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <Input
              value={searchTerm}
              onChange={(value) => setSearchTerm(value)}
              placeholder={`Search ${activeView}...`}
              className="w-full"
            />
          </div>
          
          {activeView === 'permissions' && (
            <div className="sm:w-48">
              <select
                value={selectedCategory}
                onChange={(e) => setSelectedCategory(e.target.value)}
                className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
              >
                {categories.map(category => (
                  <option key={category} value={category}>
                    {category === 'all' ? 'All Categories' : category.charAt(0).toUpperCase() + category.slice(1)}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
      )}

      {/* Content */}
      <div className="space-y-6">
        {/* Overview Tab */}
        {activeView === 'overview' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Role Summary
                </h3>
                <div className="space-y-3">
                  {user.roles.slice(0, 5).map((role) => (
                    <div key={role.id} className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <span className="text-lg">{getRoleTypeIcon(role.isSystem)}</span>
                        <span className="font-medium text-gray-900 dark:text-white">{role.name}</span>
                      </div>
                      <span className="text-sm text-gray-500 dark:text-gray-400">
                        {role.permissions.length} permissions
                      </span>
                    </div>
                  ))}
                  {user.roles.length > 5 && (
                    <button
                      onClick={() => setActiveView('roles')}
                      className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
                    >
                      View all {user.roles.length} roles →
                    </button>
                  )}
                </div>
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Permission Categories
                </h3>
                <div className="space-y-3">
                  {categories.slice(1).map((category) => {
                    const count = allPermissions.filter(p => p.category === category).length;
                    return (
                      <div key={category} className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <span className="text-lg">{getPermissionIcon(category)}</span>
                          <span className="font-medium text-gray-900 dark:text-white capitalize">
                            {category}
                          </span>
                        </div>
                        <span className={`px-2 py-1 text-xs font-medium rounded ${getCategoryColor(category)}`}>
                          {count}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>
            </Card>
          </div>
        )}

        {/* Roles Tab */}
        {activeView === 'roles' && (
          <div className="space-y-4">
            {filteredRoles.map((role) => (
              <Card key={role.id}>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center space-x-3">
                      <span className="text-2xl">{getRoleTypeIcon(role.isSystem)}</span>
                      <div>
                        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                          {role.name}
                        </h3>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                          {role.description || 'No description available'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className={`px-2 py-1 text-xs font-medium rounded ${
                        role.isSystem
                          ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
                          : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                      }`}>
                        {role.isSystem ? 'System Role' : 'Custom Role'}
                      </span>
                      <span className="text-sm text-gray-500 dark:text-gray-400">
                        {role.permissions.length} permissions
                      </span>
                    </div>
                  </div>

                  <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
                    <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                      Included Permissions:
                    </h4>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                      {role.permissions.map((permission) => (
                        <div
                          key={permission.id}
                          className="flex items-center space-x-2 p-2 bg-gray-50 dark:bg-gray-800 rounded"
                        >
                          <span className="text-sm">{getPermissionIcon(permission.category)}</span>
                          <div className="min-w-0 flex-1">
                            <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                              {permission.name}
                            </p>
                            <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
                              {permission.description}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}

        {/* Permissions Tab */}
        {activeView === 'permissions' && (
          <div className="space-y-4">
            {filteredPermissions.map((permission) => {
              const source = getPermissionSource(permission);
              return (
                <Card key={permission.id}>
                  <div className="p-6">
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center space-x-3">
                        <span className="text-xl">{getPermissionIcon(permission.category)}</span>
                        <div>
                          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                            {permission.name}
                          </h3>
                          <p className="text-sm text-gray-600 dark:text-gray-400">
                            {permission.description || 'No description available'}
                          </p>
                        </div>
                      </div>
                      <span className={`px-3 py-1 text-xs font-medium rounded-full ${getCategoryColor(permission.category)}`}>
                        {permission.category.toUpperCase()}
                      </span>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                          Available Actions:
                        </h4>
                        <div className="flex flex-wrap gap-2">
                          <span className="px-2 py-1 text-xs bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300 rounded">
                            {permission.action}
                          </span>
                        </div>
                      </div>

                      <div>
                        <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                          Granted Through:
                        </h4>
                        <div className="space-y-1">
                          {source.fromDirect && (
                            <div className="flex items-center space-x-2">
                              <span className="w-2 h-2 bg-green-500 rounded-full"></span>
                              <span className="text-sm text-gray-600 dark:text-gray-400">Direct assignment</span>
                            </div>
                          )}
                          {source.fromRoles.map((role) => (
                            <div key={role.id} className="flex items-center space-x-2">
                              <span className="w-2 h-2 bg-blue-500 rounded-full"></span>
                              <span className="text-sm text-gray-600 dark:text-gray-400">Role: {role.name}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>
              );
            })}
          </div>
        )}

        {/* Clusters Tab */}
        {activeView === 'clusters' && (
          <div className="space-y-4">
            {user.clusters.map((cluster) => (
              <Card key={cluster.clusterId}>
                <div className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center space-x-3">
                      <span className="text-2xl">☸️</span>
                      <div>
                        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                          {cluster.clusterName}
                        </h3>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                          Cluster ID: {cluster.clusterId}
                        </p>
                      </div>
                    </div>
                    <div className="flex space-x-2">
                      {cluster.canExecuteCommands && (
                        <span className="px-2 py-1 text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 rounded font-medium">
                          EXECUTE
                        </span>
                      )}
                      {cluster.canApproveCommands && (
                        <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded font-medium">
                          APPROVE
                        </span>
                      )}
                      {cluster.canViewAuditLogs && (
                        <span className="px-2 py-1 text-xs bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200 rounded font-medium">
                          AUDIT
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div>
                      <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                        Cluster Roles ({cluster.roles.length})
                      </h4>
                      <div className="space-y-2">
                        {cluster.roles.map((role) => (
                          <div
                            key={role.id}
                            className="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800 rounded"
                          >
                            <span className="text-sm font-medium text-gray-900 dark:text-white">
                              {role.name}
                            </span>
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                              {role.permissions.length} permissions
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>

                    <div>
                      <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                        Direct Permissions ({cluster.permissions.length})
                      </h4>
                      <div className="space-y-2">
                        {cluster.permissions.slice(0, 5).map((permission) => (
                          <div
                            key={permission.id}
                            className="flex items-center space-x-2 p-2 bg-gray-50 dark:bg-gray-800 rounded"
                          >
                            <span className="text-sm">{getPermissionIcon(permission.category)}</span>
                            <span className="text-sm text-gray-900 dark:text-white truncate">
                              {permission.name}
                            </span>
                          </div>
                        ))}
                        {cluster.permissions.length > 5 && (
                          <p className="text-xs text-gray-500 dark:text-gray-400 px-2">
                            +{cluster.permissions.length - 5} more permissions
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Request Access Button */}
      {showRequestButtons && (
        <Card className="border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-900/20">
          <div className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-medium text-blue-900 dark:text-blue-200">
                  Need additional access?
                </h3>
                <p className="text-sm text-blue-700 dark:text-blue-300">
                  Request additional permissions or roles from your administrator
                </p>
              </div>
              <Button
                onClick={() => setShowRequestModal(true)}
                variant="primary"
                size="sm"
              >
                Request Access
              </Button>
            </div>
          </div>
        </Card>
      )}

      {/* Request Access Modal */}
      <Modal
        isOpen={showRequestModal}
        onClose={() => setShowRequestModal(false)}
        title="Request Additional Access"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Request Type
            </label>
            <select
              value={requestType}
              onChange={(e) => setRequestType(e.target.value as 'permission' | 'role')}
              className="block w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
            >
              <option value="permission">Specific Permission</option>
              <option value="role">Role Assignment</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Business Justification (Required)
            </label>
            <textarea
              value={justification}
              onChange={(e) => setJustification(e.target.value)}
              placeholder="Explain why you need this access and how it relates to your job responsibilities..."
              rows={4}
              className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400"
            />
          </div>

          <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
            <p className="text-sm text-blue-700 dark:text-blue-300">
              <strong>Note:</strong> Access requests will be reviewed by your administrator. 
              You&apos;ll be notified via email once the request has been processed.
            </p>
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={handleRequestAccess}
              variant="primary"
              className="flex-1"
              disabled={!justification.trim()}
            >
              Submit Request
            </Button>
            <Button
              onClick={() => setShowRequestModal(false)}
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
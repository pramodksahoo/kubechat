// Admin RBAC Management Interface
// Following coding standards from docs/architecture/coding-standards.md
// AC 4: RBAC Management Interface

import React, { useEffect, useState } from 'react';
import { AdminLayout } from '../../../components/admin/AdminLayout';
import { useAdminStore } from '../../../stores/adminStore';
import type { Role, Permission } from '../../../types/admin';

export default function AdminRolesPage() {
  const {
    roles,
    permissions,
    isLoading,
    error,
    loadRoles,
    loadPermissions,
    createRole,
    updateRole,
    deleteRole,
    setError
  } = useAdminStore();

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    permissions: [] as string[],
    isSystemRole: false
  });

  useEffect(() => {
    loadRoles();
    loadPermissions();
  }, [loadRoles, loadPermissions]);

  const handleCreateRole = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createRole(formData);
      setShowCreateModal(false);
      resetForm();
    } catch (err) {
      console.error('Failed to create role:', err);
    }
  };

  const handleUpdateRole = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingRole) return;

    try {
      await updateRole(editingRole.id, formData);
      setEditingRole(null);
      resetForm();
    } catch (err) {
      console.error('Failed to update role:', err);
    }
  };

  const handleDeleteRole = async (roleId: string) => {
    if (confirm('Are you sure you want to delete this role? This action cannot be undone.')) {
      try {
        await deleteRole(roleId);
      } catch (err) {
        console.error('Failed to delete role:', err);
      }
    }
  };

  const handleBulkDelete = async () => {
    if (selectedRoles.length === 0) return;

    if (confirm(`Are you sure you want to delete ${selectedRoles.length} role(s)? This action cannot be undone.`)) {
      try {
        for (const roleId of selectedRoles) {
          await deleteRole(roleId);
        }
        setSelectedRoles([]);
      } catch (err) {
        console.error('Failed to delete roles:', err);
      }
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      permissions: [],
      isSystemRole: false
    });
  };

  const handleEditRole = (role: Role) => {
    setEditingRole(role);
    setFormData({
      name: role.name,
      description: role.description || '',
      permissions: role.permissions || [],
      isSystemRole: role.isSystemRole
    });
    setShowCreateModal(true);
  };

  const handlePermissionToggle = (permissionId: string) => {
    setFormData(prev => ({
      ...prev,
      permissions: prev.permissions.includes(permissionId)
        ? prev.permissions.filter(id => id !== permissionId)
        : [...prev.permissions, permissionId]
    }));
  };

  const getRoleUsageCount = (roleId: string) => {
    // This would typically come from an API call
    return 0;
  };

  if (error) {
    return (
      <AdminLayout title="RBAC Management">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error loading RBAC data</h3>
              <p className="text-sm text-red-700 mt-1">{error}</p>
              <button
                onClick={() => {
                  setError(null);
                  loadRoles();
                  loadPermissions();
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
    <AdminLayout title="RBAC Management">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Role-Based Access Control</h1>
            <p className="text-gray-600">Manage roles and permissions for your application</p>
          </div>
          <div className="flex space-x-3">
            {selectedRoles.length > 0 && (
              <button
                onClick={handleBulkDelete}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              >
                Delete Selected ({selectedRoles.length})
              </button>
            )}
            <button
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              Create Role
            </button>
          </div>
        </div>

        {/* Roles Table */}
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <div className="px-4 py-5 sm:p-6">
            <div className="mb-4">
              <h3 className="text-lg font-medium text-gray-900">Roles</h3>
              <p className="text-sm text-gray-500">
                {roles.length} role(s) configured
              </p>
            </div>

            {isLoading ? (
              <div className="flex justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left">
                        <input
                          type="checkbox"
                          className="rounded border-gray-300"
                          checked={selectedRoles.length === roles.length && roles.length > 0}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedRoles(roles.map(r => r.id));
                            } else {
                              setSelectedRoles([]);
                            }
                          }}
                        />
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Role
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Permissions
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Users
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Created
                      </th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {roles.map((role) => (
                      <tr key={role.id} className="hover:bg-gray-50">
                        <td className="px-6 py-4">
                          <input
                            type="checkbox"
                            className="rounded border-gray-300"
                            checked={selectedRoles.includes(role.id)}
                            onChange={(e) => {
                              if (e.target.checked) {
                                setSelectedRoles([...selectedRoles, role.id]);
                              } else {
                                setSelectedRoles(selectedRoles.filter(id => id !== role.id));
                              }
                            }}
                          />
                        </td>
                        <td className="px-6 py-4">
                          <div>
                            <div className="text-sm font-medium text-gray-900">{role.name}</div>
                            {role.description && (
                              <div className="text-sm text-gray-500">{role.description}</div>
                            )}
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-gray-900">
                            {role.permissions?.length || 0} permissions
                          </div>
                          <div className="text-xs text-gray-500">
                            {role.permissions?.slice(0, 3).join(', ')}
                            {(role.permissions?.length || 0) > 3 && '...'}
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-gray-900">
                            {getRoleUsageCount(role.id)} users
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="text-sm text-gray-900">
                            {new Date(role.createdAt).toLocaleDateString()}
                          </div>
                        </td>
                        <td className="px-6 py-4 text-right text-sm font-medium space-x-2">
                          <button
                            onClick={() => handleEditRole(role)}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => handleDeleteRole(role.id)}
                            className="text-red-600 hover:text-red-900"
                            disabled={getRoleUsageCount(role.id) > 0}
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>

        {/* Create/Edit Role Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-1/2 shadow-lg rounded-md bg-white">
              <div className="mt-3">
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  {editingRole ? 'Edit Role' : 'Create New Role'}
                </h3>

                <form onSubmit={editingRole ? handleUpdateRole : handleCreateRole}>
                  <div className="space-y-4">
                    <div>
                      <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                        Role Name
                      </label>
                      <input
                        type="text"
                        id="name"
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        required
                        className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Enter role name"
                      />
                    </div>

                    <div>
                      <label htmlFor="description" className="block text-sm font-medium text-gray-700">
                        Description
                      </label>
                      <textarea
                        id="description"
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                        rows={3}
                        className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Enter role description"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Permissions
                      </label>
                      <div className="border border-gray-300 rounded-md p-3 max-h-60 overflow-y-auto">
                        {permissions.map((permission) => (
                          <div key={permission.id} className="flex items-center space-x-2 py-1">
                            <input
                              type="checkbox"
                              id={`permission-${permission.id}`}
                              checked={formData.permissions.includes(permission.id)}
                              onChange={() => handlePermissionToggle(permission.id)}
                              className="rounded border-gray-300"
                            />
                            <label
                              htmlFor={`permission-${permission.id}`}
                              className="text-sm text-gray-900 cursor-pointer flex-1"
                            >
                              <span className="font-medium">{permission.name}</span>
                              {permission.description && (
                                <span className="text-gray-500 ml-2">- {permission.description}</span>
                              )}
                            </label>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>

                  <div className="flex justify-end space-x-3 mt-6">
                    <button
                      type="button"
                      onClick={() => {
                        setShowCreateModal(false);
                        setEditingRole(null);
                        resetForm();
                      }}
                      className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={isLoading}
                      className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                    >
                      {isLoading ? 'Saving...' : (editingRole ? 'Update Role' : 'Create Role')}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        )}
      </div>
    </AdminLayout>
  );
}
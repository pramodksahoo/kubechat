import React, { useState, useEffect, useCallback } from 'react';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { Icon } from '../ui/Icon';
import { Modal } from '../ui/Modal';
import { Input } from '../ui/Input';
import { Select } from '../ui/Select';
import { format } from 'date-fns';

interface RetentionPolicy {
  id: string;
  name: string;
  description: string;
  retentionPeriod: number; // days
  archivalPeriod: number; // days
  complianceFramework?: string;
  logTypes: string[];
  compressionEnabled: boolean;
  encryptionEnabled: boolean;
  status: 'active' | 'inactive' | 'pending';
  createdAt: string;
  lastExecuted?: string;
}

interface ArchivalJob {
  id: string;
  policyId: string;
  policyName: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  startTime: string;
  endTime?: string;
  recordsProcessed: number;
  archiveSize: number; // bytes
  compressionRatio?: number;
  errorMessage?: string;
}

interface LogRetentionManagerProps {
  className?: string;
}

export function LogRetentionManager({ className = '' }: LogRetentionManagerProps) {
  const [policies, setPolicies] = useState<RetentionPolicy[]>([]);
  const [jobs, setJobs] = useState<ArchivalJob[]>([]);
  const [showCreatePolicy, setShowCreatePolicy] = useState(false);
  const [showJobDetails, setShowJobDetails] = useState(false);
  const [selectedJob, setSelectedJob] = useState<ArchivalJob | null>(null);
  const [loading, setLoading] = useState(false);

  const [newPolicy, setNewPolicy] = useState<Partial<RetentionPolicy>>({
    name: '',
    description: '',
    retentionPeriod: 365,
    archivalPeriod: 30,
    logTypes: ['audit', 'system'],
    compressionEnabled: true,
    encryptionEnabled: true,
    status: 'active'
  });

  // Load retention policies
  const loadPolicies = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/audit/retention-policies');
      if (response.ok) {
        const data = await response.json();
        setPolicies(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load retention policies:', error);
    }
  }, []);

  // Load archival jobs
  const loadJobs = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/audit/archival-jobs?limit=20');
      if (response.ok) {
        const data = await response.json();
        setJobs(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load archival jobs:', error);
    }
  }, []);

  // Create retention policy
  const createPolicy = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/audit/retention-policies', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newPolicy),
      });

      if (response.ok) {
        const data = await response.json();
        setPolicies(prev => [data.data, ...prev]);
        setShowCreatePolicy(false);
        resetNewPolicy();
      }
    } catch (error) {
      console.error('Failed to create retention policy:', error);
    } finally {
      setLoading(false);
    }
  }, [newPolicy]);

  // Execute archival job
  const executeArchival = useCallback(async (policyId: string) => {
    try {
      const response = await fetch(`/api/v1/audit/retention-policies/${policyId}/execute`, {
        method: 'POST',
      });

      if (response.ok) {
        const data = await response.json();
        setJobs(prev => [data.data, ...prev]);
        loadJobs(); // Refresh jobs list
      }
    } catch (error) {
      console.error('Failed to execute archival:', error);
    }
  }, [loadJobs]);

  // Delete policy
  const deletePolicy = useCallback(async (policyId: string) => {
    try {
      const response = await fetch(`/api/v1/audit/retention-policies/${policyId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setPolicies(prev => prev.filter(p => p.id !== policyId));
      }
    } catch (error) {
      console.error('Failed to delete policy:', error);
    }
  }, []);

  const resetNewPolicy = () => {
    setNewPolicy({
      name: '',
      description: '',
      retentionPeriod: 365,
      archivalPeriod: 30,
      logTypes: ['audit', 'system'],
      compressionEnabled: true,
      encryptionEnabled: true,
      status: 'active'
    });
  };

  useEffect(() => {
    loadPolicies();
    loadJobs();
  }, [loadPolicies, loadJobs]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'inactive': return 'gray';
      case 'pending': return 'yellow';
      case 'running': return 'blue';
      case 'completed': return 'green';
      case 'failed': return 'red';
      default: return 'gray';
    }
  };

  const formatBytes = (bytes: number) => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const calculateNextExecution = (policy: RetentionPolicy) => {
    if (!policy.lastExecuted) return 'Next execution: Tonight';
    const lastRun = new Date(policy.lastExecuted);
    const nextRun = new Date(lastRun.getTime() + 24 * 60 * 60 * 1000);
    return `Next execution: ${format(nextRun, 'MMM dd, HH:mm')}`;
  };

  return (
    <>
      <div className={`space-y-6 ${className}`}>
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Log Retention & Archival</h2>
            <p className="text-gray-600">Automated log lifecycle management and compliance archival</p>
          </div>
          <Button onClick={() => setShowCreatePolicy(true)}>
            <Icon name="document-plus" className="w-4 h-4 mr-2" />
            Create Policy
          </Button>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Active Policies</h3>
                  <p className="text-sm text-gray-600">Retention rules</p>
                </div>
                <Icon name="document-text" className="w-8 h-8 text-blue-600" />
              </div>
              <div className="mt-4">
                <div className="text-3xl font-bold text-blue-600">
                  {policies.filter(p => p.status === 'active').length}
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Running Jobs</h3>
                  <p className="text-sm text-gray-600">Active archival</p>
                </div>
                <Icon name="cog" className="w-8 h-8 text-yellow-600" />
              </div>
              <div className="mt-4">
                <div className="text-3xl font-bold text-yellow-600">
                  {jobs.filter(j => j.status === 'running').length}
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Archives Created</h3>
                  <p className="text-sm text-gray-600">Last 30 days</p>
                </div>
                <Icon name="archive" className="w-8 h-8 text-green-600" />
              </div>
              <div className="mt-4">
                <div className="text-3xl font-bold text-green-600">
                  {jobs.filter(j => j.status === 'completed').length}
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">Storage Saved</h3>
                  <p className="text-sm text-gray-600">Compression ratio</p>
                </div>
                <Icon name="chart-bar" className="w-8 h-8 text-purple-600" />
              </div>
              <div className="mt-4">
                <div className="text-3xl font-bold text-purple-600">
                  {Math.round(jobs.reduce((acc, j) => acc + (j.compressionRatio || 0), 0) / Math.max(jobs.length, 1))}%
                </div>
              </div>
            </div>
          </Card>
        </div>

        {/* Retention Policies */}
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-medium mb-4">Retention Policies</h3>
            {policies.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Icon name="document-text" className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>No retention policies configured.</p>
                <p className="text-sm">Create your first policy to start automated archival.</p>
              </div>
            ) : (
              <div className="space-y-4">
                {policies.map((policy) => (
                  <div key={policy.id} className="border border-gray-200 rounded-md p-4 hover:bg-gray-50">
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3 mb-2">
                          <h4 className="font-medium text-gray-900">{policy.name}</h4>
                          <Badge variant={getStatusColor(policy.status)}>
                            {policy.status.toUpperCase()}
                          </Badge>
                          {policy.complianceFramework && (
                            <Badge variant="blue">
                              {policy.complianceFramework.toUpperCase()}
                            </Badge>
                          )}
                        </div>

                        <p className="text-sm text-gray-600 mb-3">{policy.description}</p>

                        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 text-sm">
                          <div>
                            <span className="font-medium text-gray-700">Retention:</span>
                            <div className="text-gray-600">{policy.retentionPeriod} days</div>
                          </div>
                          <div>
                            <span className="font-medium text-gray-700">Archival:</span>
                            <div className="text-gray-600">{policy.archivalPeriod} days</div>
                          </div>
                          <div>
                            <span className="font-medium text-gray-700">Log Types:</span>
                            <div className="text-gray-600">{policy.logTypes.join(', ')}</div>
                          </div>
                          <div>
                            <span className="font-medium text-gray-700">Features:</span>
                            <div className="text-gray-600">
                              {policy.compressionEnabled && '🗜️ '}
                              {policy.encryptionEnabled && '🔐 '}
                            </div>
                          </div>
                          <div>
                            <span className="font-medium text-gray-700">Schedule:</span>
                            <div className="text-gray-600">{calculateNextExecution(policy)}</div>
                          </div>
                        </div>
                      </div>

                      <div className="flex space-x-2">
                        <Button
                          variant="primary"
                          size="sm"
                          onClick={() => executeArchival(policy.id)}
                          disabled={policy.status !== 'active'}
                        >
                          <Icon name="play" className="w-4 h-4 mr-1" />
                          Run Now
                        </Button>
                        <Button
                          variant="danger"
                          size="sm"
                          onClick={() => deletePolicy(policy.id)}
                        >
                          <Icon name="trash" className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </Card>

        {/* Recent Archival Jobs */}
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-medium mb-4">Recent Archival Jobs</h3>
            {jobs.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Icon name="archive" className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>No archival jobs executed yet.</p>
              </div>
            ) : (
              <div className="space-y-3">
                {jobs.slice(0, 10).map((job) => (
                  <div
                    key={job.id}
                    className="border border-gray-200 rounded-md p-3 cursor-pointer hover:bg-gray-50"
                    onClick={() => {
                      setSelectedJob(job);
                      setShowJobDetails(true);
                    }}
                  >
                    <div className="flex justify-between items-center">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <Badge variant={getStatusColor(job.status)}>
                            {job.status.toUpperCase()}
                          </Badge>
                          <span className="font-medium">{job.policyName}</span>
                          <span className="text-sm text-gray-500">
                            {format(new Date(job.startTime), 'MMM dd, HH:mm')}
                          </span>
                        </div>

                        <div className="mt-1 text-sm text-gray-600">
                          Processed: {job.recordsProcessed.toLocaleString()} records •
                          Size: {formatBytes(job.archiveSize)}
                          {job.compressionRatio && ` • Compression: ${job.compressionRatio}%`}
                          {job.status === 'running' && (
                            <span className="ml-2">
                              <div className="inline-block w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
                            </span>
                          )}
                        </div>
                      </div>

                      <Icon name="chevron-right" className="w-5 h-5 text-gray-400" />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Create Policy Modal */}
      <Modal
        isOpen={showCreatePolicy}
        onClose={() => setShowCreatePolicy(false)}
        title="Create Retention Policy"
        size="lg"
      >
        <div className="space-y-4">
          <Input
            label="Policy Name"
            value={newPolicy.name || ''}
            onChange={(value) => setNewPolicy(prev => ({ ...prev, name: value }))}
            placeholder="e.g., SOX Compliance Policy"
            required
          />

          <Input
            label="Description"
            value={newPolicy.description || ''}
            onChange={(value) => setNewPolicy(prev => ({ ...prev, description: value }))}
            placeholder="Brief description of this policy"
          />

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Retention Period (days)
              </label>
              <input
                type="number"
                value={newPolicy.retentionPeriod || 365}
                onChange={(e) => setNewPolicy(prev => ({
                  ...prev,
                  retentionPeriod: parseInt(e.target.value) || 365
                }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                min="1"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Archival After (days)
              </label>
              <input
                type="number"
                value={newPolicy.archivalPeriod || 30}
                onChange={(e) => setNewPolicy(prev => ({
                  ...prev,
                  archivalPeriod: parseInt(e.target.value) || 30
                }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                min="1"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Log Types
            </label>
            <div className="space-y-2">
              {['audit', 'system', 'security', 'application'].map((type) => (
                <label key={type} className="flex items-center">
                  <input
                    type="checkbox"
                    checked={newPolicy.logTypes?.includes(type) || false}
                    onChange={(e) => {
                      const logTypes = newPolicy.logTypes || [];
                      if (e.target.checked) {
                        setNewPolicy(prev => ({
                          ...prev,
                          logTypes: [...logTypes, type]
                        }));
                      } else {
                        setNewPolicy(prev => ({
                          ...prev,
                          logTypes: logTypes.filter(t => t !== type)
                        }));
                      }
                    }}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span className="ml-2 text-sm text-gray-700 capitalize">{type} Logs</span>
                </label>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Archive Options
            </label>
            <div className="space-y-2">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={newPolicy.compressionEnabled || false}
                  onChange={(e) => setNewPolicy(prev => ({
                    ...prev,
                    compressionEnabled: e.target.checked
                  }))}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <span className="ml-2 text-sm text-gray-700">Enable Compression</span>
              </label>

              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={newPolicy.encryptionEnabled || false}
                  onChange={(e) => setNewPolicy(prev => ({
                    ...prev,
                    encryptionEnabled: e.target.checked
                  }))}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <span className="ml-2 text-sm text-gray-700">Enable Encryption</span>
              </label>
            </div>
          </div>

          <div className="flex justify-end space-x-3">
            <Button variant="secondary" onClick={() => setShowCreatePolicy(false)}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={createPolicy}
              disabled={loading || !newPolicy.name}
            >
              {loading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Creating...
                </>
              ) : (
                'Create Policy'
              )}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Job Details Modal */}
      <Modal
        isOpen={showJobDetails}
        onClose={() => setShowJobDetails(false)}
        title={`Archival Job: ${selectedJob?.policyName}`}
      >
        {selectedJob && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <h4 className="font-medium text-gray-900 mb-2">Job Information</h4>
                <dl className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <dt className="text-gray-600">Status:</dt>
                    <dd>
                      <Badge variant={getStatusColor(selectedJob.status)}>
                        {selectedJob.status.toUpperCase()}
                      </Badge>
                    </dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-600">Started:</dt>
                    <dd className="font-medium">{format(new Date(selectedJob.startTime), 'MMM dd, yyyy HH:mm:ss')}</dd>
                  </div>
                  {selectedJob.endTime && (
                    <div className="flex justify-between">
                      <dt className="text-gray-600">Completed:</dt>
                      <dd className="font-medium">{format(new Date(selectedJob.endTime), 'MMM dd, yyyy HH:mm:ss')}</dd>
                    </div>
                  )}
                </dl>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-2">Processing Stats</h4>
                <dl className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <dt className="text-gray-600">Records:</dt>
                    <dd className="font-medium">{selectedJob.recordsProcessed.toLocaleString()}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-gray-600">Archive Size:</dt>
                    <dd className="font-medium">{formatBytes(selectedJob.archiveSize)}</dd>
                  </div>
                  {selectedJob.compressionRatio && (
                    <div className="flex justify-between">
                      <dt className="text-gray-600">Compression:</dt>
                      <dd className="font-medium">{selectedJob.compressionRatio}%</dd>
                    </div>
                  )}
                </dl>
              </div>
            </div>

            {selectedJob.errorMessage && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3">
                <h5 className="font-medium text-red-800 mb-1">Error Details</h5>
                <p className="text-sm text-red-700">{selectedJob.errorMessage}</p>
              </div>
            )}

            <div className="flex justify-end space-x-3">
              <Button variant="secondary" onClick={() => setShowJobDetails(false)}>
                Close
              </Button>
              {selectedJob.status === 'completed' && (
                <Button variant="primary">
                  <Icon name="download" className="w-4 h-4 mr-2" />
                  Download Archive
                </Button>
              )}
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}
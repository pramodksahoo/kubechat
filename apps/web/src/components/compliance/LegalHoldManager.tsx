import React, { useState, useEffect } from 'react';
import { LegalHold, LegalHoldRequest } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Select } from '../ui/Select';
import { Badge } from '../ui/Badge';
import { Icon } from '../ui/Icon';
import { Modal } from '../ui/Modal';
import { format } from 'date-fns';

interface LegalHoldManagerProps {
  className?: string;
}

export function LegalHoldManager({ className = '' }: LegalHoldManagerProps) {
  const [legalHolds, setLegalHolds] = useState<LegalHold[]>([]);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [selectedHold, setSelectedHold] = useState<LegalHold | null>(null);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);

  const [createForm, setCreateForm] = useState<LegalHoldRequest>({
    caseNumber: '',
    description: '',
    startTime: new Date().toISOString(),
    endTime: ''
  });

  // Load legal holds
  const loadLegalHolds = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/audit/legal-holds');
      if (response.ok) {
        const data = await response.json();
        setLegalHolds(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load legal holds:', error);
    } finally {
      setLoading(false);
    }
  };

  // Create legal hold
  const createLegalHold = async () => {
    setCreating(true);
    try {
      const response = await fetch('/api/v1/audit/legal-holds', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(createForm),
      });

      if (response.ok) {
        const data = await response.json();
        setLegalHolds(prev => [data.data, ...prev]);
        setShowCreateModal(false);
        resetCreateForm();
      }
    } catch (error) {
      console.error('Failed to create legal hold:', error);
    } finally {
      setCreating(false);
    }
  };

  // Release legal hold
  const releaseLegalHold = async (holdId: string) => {
    try {
      const response = await fetch(`/api/v1/audit/legal-holds/${holdId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setLegalHolds(prev => prev.map(hold =>
          hold.id === holdId
            ? { ...hold, status: 'released', releasedAt: new Date().toISOString() }
            : hold
        ));
      }
    } catch (error) {
      console.error('Failed to release legal hold:', error);
    }
  };

  const resetCreateForm = () => {
    setCreateForm({
      caseNumber: '',
      description: '',
      startTime: new Date().toISOString(),
      endTime: ''
    });
  };

  useEffect(() => {
    loadLegalHolds();
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'pending': return 'yellow';
      case 'released': return 'gray';
      default: return 'gray';
    }
  };


  return (
    <>
      <div className={`space-y-6 ${className}`}>
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Legal Hold Management</h2>
            <p className="text-gray-600">Manage litigation hold notices and data preservation</p>
          </div>
          <Button onClick={() => setShowCreateModal(true)}>
            <Icon name="document-plus" className="w-4 h-4 mr-2" />
            Create Legal Hold
          </Button>
        </div>

        {/* Legal Holds List */}
        <Card>
          <div className="p-6">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-medium">Active Legal Holds</h3>
              <Button
                variant="secondary"
                size="sm"
                onClick={loadLegalHolds}
                disabled={loading}
              >
                <Icon name="refresh" className="w-4 h-4 mr-2" />
                Refresh
              </Button>
            </div>

            {loading ? (
              <div className="flex justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : legalHolds.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Icon name="scale" className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p>No legal holds in place.</p>
                <p className="text-sm">Create a legal hold to preserve audit data for litigation.</p>
              </div>
            ) : (
              <div className="space-y-4">
                {legalHolds.map((hold) => (
                  <div
                    key={hold.id}
                    className="border border-gray-200 rounded-md p-4 hover:bg-gray-50 cursor-pointer"
                    onClick={() => {
                      setSelectedHold(hold);
                      setShowDetailsModal(true);
                    }}
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3 mb-2">
                          <h4 className="font-medium text-gray-900">Case #{hold.caseNumber}</h4>
                          <Badge variant={getStatusColor(hold.status)}>
                            {hold.status.toUpperCase()}
                          </Badge>
                          {hold.status === 'active' && (
                            <Badge variant="blue">
                              {hold.recordCount} records preserved
                            </Badge>
                          )}
                        </div>

                        <p className="text-sm text-gray-600 mb-3">{hold.description}</p>

                        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-xs text-gray-500">
                          <div>
                            <span className="font-medium">Created By:</span>
                            <div>{hold.createdBy}</div>
                          </div>
                          <div>
                            <span className="font-medium">Created:</span>
                            <div>{format(new Date(hold.createdAt), 'MMM dd, yyyy')}</div>
                          </div>
                          <div>
                            <span className="font-medium">Start Time:</span>
                            <div>{format(new Date(hold.startTime), 'MMM dd, yyyy')}</div>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center space-x-2">
                        {hold.status === 'active' && (
                          <Button
                            variant="danger"
                            size="sm"
                            onClick={() => releaseLegalHold(hold.id)}
                          >
                            Release
                          </Button>
                        )}
                        <Icon name="chevron-right" className="w-5 h-5 text-gray-400" />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Create Legal Hold Modal */}
      <Modal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        title="Create Legal Hold"
        size="lg"
      >
        <div className="space-y-4">
          <Input
            label="Case Number"
            value={createForm.caseNumber}
            onChange={(value) => setCreateForm(prev => ({ ...prev, caseNumber: value }))}
            placeholder="CASE-2025-001"
            required
          />

          <Input
            label="Description"
            value={createForm.description}
            onChange={(value) => setCreateForm(prev => ({ ...prev, description: value }))}
            placeholder="Brief description of the legal matter"
            required
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Start Time
              </label>
              <input
                type="date"
                value={createForm.startTime.split('T')[0]}
                onChange={(e) => setCreateForm(prev => ({
                  ...prev,
                  startTime: new Date(e.target.value).toISOString()
                }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                End Time (Optional)
              </label>
              <input
                type="date"
                value={createForm.endTime ? createForm.endTime.split('T')[0] : ''}
                onChange={(e) => setCreateForm(prev => ({
                  ...prev,
                  endTime: e.target.value ? new Date(e.target.value).toISOString() : ''
                }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>

          <div className="flex justify-end space-x-3">
            <Button variant="secondary" onClick={() => setShowCreateModal(false)}>
              Cancel
            </Button>
            <Button
              onClick={createLegalHold}
              disabled={creating || !createForm.caseNumber || !createForm.description}
            >
              {creating ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Creating...
                </>
              ) : (
                <>
                  <Icon name="scale" className="w-4 h-4 mr-2" />
                  Create Legal Hold
                </>
              )}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Legal Hold Details Modal */}
      <Modal
        isOpen={showDetailsModal}
        onClose={() => setShowDetailsModal(false)}
        title={`Legal Hold: Case #${selectedHold?.caseNumber}`}
        size="lg"
      >
        {selectedHold && (
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <h4 className="font-medium text-gray-900 mb-2">Case Details</h4>
                <dl className="space-y-2 text-sm">
                  <div>
                    <dt className="text-gray-600">Case Number:</dt>
                    <dd className="font-medium">{selectedHold.caseNumber}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-600">Status:</dt>
                    <dd>
                      <Badge variant={getStatusColor(selectedHold.status)}>
                        {selectedHold.status.toUpperCase()}
                      </Badge>
                    </dd>
                  </div>
                  <div>
                    <dt className="text-gray-600">Created By:</dt>
                    <dd className="font-medium">{selectedHold.createdBy}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-600">Created:</dt>
                    <dd className="font-medium">{format(new Date(selectedHold.createdAt), 'MMM dd, yyyy HH:mm')}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-600">Start Time:</dt>
                    <dd className="font-medium">{format(new Date(selectedHold.startTime), 'MMM dd, yyyy HH:mm')}</dd>
                  </div>
                  {selectedHold.endTime && (
                    <div>
                      <dt className="text-gray-600">End Time:</dt>
                      <dd className="font-medium">{format(new Date(selectedHold.endTime), 'MMM dd, yyyy HH:mm')}</dd>
                    </div>
                  )}
                </dl>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-2">Preservation Stats</h4>
                <dl className="space-y-2 text-sm">
                  <div>
                    <dt className="text-gray-600">Records Preserved:</dt>
                    <dd className="font-medium text-blue-600">{selectedHold.recordCount.toLocaleString()}</dd>
                  </div>
                </dl>
              </div>
            </div>

            <div>
              <h4 className="font-medium text-gray-900 mb-2">Description</h4>
              <p className="text-sm text-gray-600">{selectedHold.description}</p>
            </div>

            <div className="flex justify-end space-x-3">
              <Button variant="secondary" onClick={() => setShowDetailsModal(false)}>
                Close
              </Button>
              <Button variant="primary">
                <Icon name="download" className="w-4 h-4 mr-2" />
                Export Preserved Data
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}
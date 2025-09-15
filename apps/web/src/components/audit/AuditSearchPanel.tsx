import React, { useState } from 'react';
import { AuditFilter } from '@kubechat/shared/types';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Select } from '../ui/Select';
import { Icon } from '../ui/Icon';

interface AuditSearchPanelProps {
  filters: AuditFilter;
  onFiltersChange: (filters: AuditFilter) => void;
  onSearch: (filters: AuditFilter) => void;
  loading?: boolean;
  className?: string;
}

export function AuditSearchPanel({
  filters,
  onFiltersChange,
  onSearch,
  loading = false,
  className = ''
}: AuditSearchPanelProps) {
  const [localFilters, setLocalFilters] = useState<AuditFilter>(filters);

  const handleFilterChange = (key: keyof AuditFilter, value: string) => {
    const newFilters = { ...localFilters, [key]: value || undefined };
    setLocalFilters(newFilters);
  };

  const handleSearch = () => {
    onFiltersChange(localFilters);
    onSearch(localFilters);
  };

  const handleClearFilters = () => {
    const clearedFilters: AuditFilter = {
      startDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
      endDate: new Date().toISOString(),
    };
    setLocalFilters(clearedFilters);
    onFiltersChange(clearedFilters);
    onSearch(clearedFilters);
  };

  const setQuickTimeRange = (range: '1h' | '24h' | '7d' | '30d') => {
    const now = new Date();
    let start: Date;

    switch (range) {
      case '1h':
        start = new Date(now.getTime() - 60 * 60 * 1000);
        break;
      case '24h':
        start = new Date(now.getTime() - 24 * 60 * 60 * 1000);
        break;
      case '7d':
        start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case '30d':
        start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
    }

    const newFilters = {
      ...localFilters,
      startDate: start.toISOString(),
      endDate: now.toISOString(),
    };
    setLocalFilters(newFilters);
    onFiltersChange(newFilters);
    onSearch(newFilters);
  };

  return (
    <Card className={className}>
      <div className="p-6">
        <h3 className="text-lg font-medium mb-4 flex items-center">
          <Icon name="search" className="w-5 h-5 mr-2" />
          Search & Filter
        </h3>

        <div className="space-y-4">
          {/* Quick Time Ranges */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Quick Time Range
            </label>
            <div className="grid grid-cols-2 gap-2">
              {[
                { label: '1 Hour', value: '1h' },
                { label: '24 Hours', value: '24h' },
                { label: '7 Days', value: '7d' },
                { label: '30 Days', value: '30d' }
              ].map((range) => (
                <Button
                  key={range.value}
                  variant="secondary"
                  size="sm"
                  onClick={() => setQuickTimeRange(range.value as any)}
                  className="text-xs"
                >
                  {range.label}
                </Button>
              ))}
            </div>
          </div>

          {/* Date Range */}
          <div className="grid grid-cols-1 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Start Date
              </label>
              <Input
                type="datetime-local"
                value={localFilters.startDate ? new Date(localFilters.startDate).toISOString().slice(0, 16) : ''}
                onChange={(value) => handleFilterChange('startDate', value ? new Date(value).toISOString() : '')}
                className="w-full"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                End Date
              </label>
              <Input
                type="datetime-local"
                value={localFilters.endDate ? new Date(localFilters.endDate).toISOString().slice(0, 16) : ''}
                onChange={(value) => handleFilterChange('endDate', value ? new Date(value).toISOString() : '')}
                className="w-full"
              />
            </div>
          </div>

          {/* User ID */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              User ID
            </label>
            <Input
              type="text"
              placeholder="Enter user ID"
              value={localFilters.userId || ''}
              onChange={(value) => handleFilterChange('userId', value)}
              className="w-full"
            />
          </div>

          {/* Cluster ID */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Cluster ID
            </label>
            <Input
              type="text"
              placeholder="Enter cluster ID"
              value={localFilters.clusterId || ''}
              onChange={(value) => handleFilterChange('clusterId', value)}
              className="w-full"
            />
          </div>

          {/* Action */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Action
            </label>
            <Select
              value={localFilters.action || ''}
              onChange={(e) => handleFilterChange('action', e.target.value)}
              className="w-full"
            >
              <option value="">All Actions</option>
              <option value="create">Create</option>
              <option value="update">Update</option>
              <option value="delete">Delete</option>
              <option value="read">Read</option>
              <option value="execute">Execute</option>
              <option value="login">Login</option>
              <option value="logout">Logout</option>
            </Select>
          </div>

          {/* Resource */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Resource
            </label>
            <Input
              type="text"
              placeholder="Enter resource name"
              value={localFilters.resource || ''}
              onChange={(value) => handleFilterChange('resource', value)}
              className="w-full"
            />
          </div>

          {/* Result */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Result
            </label>
            <Select
              value={localFilters.result || ''}
              onChange={(e) => handleFilterChange('result', e.target.value)}
              className="w-full"
            >
              <option value="">All Results</option>
              <option value="success">Success</option>
              <option value="failure">Failure</option>
              <option value="error">Error</option>
            </Select>
          </div>

          {/* Severity */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Severity
            </label>
            <Select
              value={localFilters.severity || ''}
              onChange={(e) => handleFilterChange('severity', e.target.value)}
              className="w-full"
            >
              <option value="">All Severities</option>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </Select>
          </div>

          {/* Search Query */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Search Query
            </label>
            <Input
              type="text"
              placeholder="Search in logs..."
              value={localFilters.searchQuery || ''}
              onChange={(value) => handleFilterChange('searchQuery', value)}
              className="w-full"
            />
          </div>

          {/* Action Buttons */}
          <div className="flex space-x-2">
            <Button
              variant="primary"
              onClick={handleSearch}
              disabled={loading}
              className="flex-1"
            >
              {loading ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              ) : (
                <Icon name="search" className="w-4 h-4 mr-2" />
              )}
              Search
            </Button>
            <Button
              variant="secondary"
              onClick={handleClearFilters}
              disabled={loading}
            >
              <Icon name="x" className="w-4 h-4" />
            </Button>
          </div>

          {/* Active Filters Summary */}
          {(localFilters.userId || localFilters.action || localFilters.severity || localFilters.searchQuery) && (
            <div className="mt-4 p-3 bg-blue-50 rounded-md">
              <h4 className="text-sm font-medium text-blue-900 mb-2">Active Filters:</h4>
              <div className="flex flex-wrap gap-2">
                {localFilters.userId && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    User: {localFilters.userId}
                  </span>
                )}
                {localFilters.action && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    Action: {localFilters.action}
                  </span>
                )}
                {localFilters.severity && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    Severity: {localFilters.severity}
                  </span>
                )}
                {localFilters.searchQuery && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    Search: {localFilters.searchQuery}
                  </span>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}
import { useState, useEffect } from 'react';
import { api } from '@/services/api';

interface CommandResultDisplayProps {
  executionId: string;
}

export function CommandResultDisplay({ executionId }: CommandResultDisplayProps) {
  const [execution, setExecution] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchExecutionDetails = async () => {
      try {
        setLoading(true);
        const response = await api.commands.getExecution(executionId);
        setExecution(response.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch execution details');
      } finally {
        setLoading(false);
      }
    };

    if (executionId) {
      fetchExecutionDetails();
    }
  }, [executionId]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-600 bg-green-100 border-green-300';
      case 'running': return 'text-blue-600 bg-blue-100 border-blue-300';
      case 'failed': return 'text-red-600 bg-red-100 border-red-300';
      case 'pending': return 'text-yellow-600 bg-yellow-100 border-yellow-300';
      case 'cancelled': return 'text-gray-600 bg-gray-100 border-gray-300';
      default: return 'text-gray-600 bg-gray-100 border-gray-300';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return '✅';
      case 'running': return '🔄';
      case 'failed': return '❌';
      case 'pending': return '⏳';
      case 'cancelled': return '🚫';
      default: return '❓';
    }
  };

  if (loading) {
    return (
      <div className="mt-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
          <span className="text-sm text-gray-600 dark:text-gray-400">Loading execution details...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mt-3 p-3 bg-red-50 border border-red-200 rounded-lg">
        <div className="flex items-center gap-2">
          <span className="text-red-600">❌</span>
          <span className="text-red-800 text-sm font-medium">Failed to load execution details</span>
        </div>
        <div className="mt-1 text-red-700 text-sm">{error}</div>
      </div>
    );
  }

  if (!execution) {
    return null;
  }

  return (
    <div className="mt-3 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
      {/* Command Execution Header (AC: 3) */}
      <div className="px-3 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-gray-900 dark:text-white">
              Command Execution
            </span>
            <span className={`px-2 py-1 text-xs font-medium rounded-full border ${getStatusColor(execution.status)}`}>
              {getStatusIcon(execution.status)} {execution.status.toUpperCase()}
            </span>
          </div>
          <div className="text-xs text-gray-500 dark:text-gray-400">
            ID: {executionId}
          </div>
        </div>

        {/* Execution Metadata (AC: 3) */}
        <div className="mt-2 grid grid-cols-2 gap-4 text-xs text-gray-600 dark:text-gray-400">
          <div>
            <span className="font-medium">Started:</span>{' '}
            {execution.executedAt ? new Date(execution.executedAt).toLocaleString() : 'Unknown'}
          </div>
          <div>
            <span className="font-medium">Duration:</span>{' '}
            {execution.completedAt && execution.executedAt
              ? `${execution.completedAt - execution.executedAt}ms`
              : 'In progress'}
          </div>
        </div>
      </div>

      {/* Command Output Display (AC: 2) */}
      {execution.output && (
        <div className="p-3">
          <div className="mb-2">
            <span className="text-sm font-medium text-gray-900 dark:text-white">Output:</span>
            <button
              onClick={() => navigator.clipboard?.writeText(execution.output)}
              className="ml-2 text-xs text-blue-600 hover:text-blue-800 transition-colors"
              title="Copy output"
            >
              📋 Copy
            </button>
          </div>
          <div className="bg-gray-900 text-green-400 rounded-md p-3 font-mono text-sm overflow-x-auto">
            <pre className="whitespace-pre-wrap break-all">{execution.output}</pre>
          </div>
        </div>
      )}

      {/* Error Display */}
      {execution.error && (
        <div className="p-3 border-t border-gray-200 dark:border-gray-700">
          <div className="mb-2">
            <span className="text-sm font-medium text-red-600">Error:</span>
          </div>
          <div className="bg-red-50 border border-red-200 rounded-md p-3">
            <pre className="text-red-800 text-sm whitespace-pre-wrap break-all">{execution.error}</pre>
          </div>
        </div>
      )}

      {/* Exit Code Display */}
      {execution.exitCode !== null && execution.exitCode !== undefined && (
        <div className="px-3 py-2 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between text-xs">
            <span className="text-gray-600 dark:text-gray-400">
              <span className="font-medium">Exit Code:</span> {execution.exitCode}
            </span>
            <span className={`font-medium ${execution.exitCode === 0 ? 'text-green-600' : 'text-red-600'}`}>
              {execution.exitCode === 0 ? 'Success' : 'Failed'}
            </span>
          </div>
        </div>
      )}

      {/* Command Information */}
      <div className="px-3 py-2 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
        <div className="text-xs text-gray-600 dark:text-gray-400">
          <span className="font-medium">Command:</span>{' '}
          <code className="bg-gray-200 dark:bg-gray-700 px-1 py-0.5 rounded text-xs">
            {execution.command}
          </code>
        </div>
      </div>
    </div>
  );
}
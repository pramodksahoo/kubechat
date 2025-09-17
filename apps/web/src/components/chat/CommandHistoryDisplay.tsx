import { useState, useEffect } from 'react';
import { CommandExecution } from '@/types/chat';
import { chatService } from '@/services/chatService';
import { formatDistanceToNow } from 'date-fns';

interface CommandHistoryDisplayProps {
  sessionId?: string;
  limit?: number;
  className?: string;
}

export function CommandHistoryDisplay({
  sessionId,
  limit = 20,
  className = ''
}: CommandHistoryDisplayProps) {
  const [executions, setExecutions] = useState<CommandExecution[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCommandHistory = async () => {
      try {
        setLoading(true);
        const history = await chatService.getCommandHistory(sessionId, limit);
        setExecutions(history);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch command history');
      } finally {
        setLoading(false);
      }
    };

    fetchCommandHistory();
  }, [sessionId, limit]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-600 bg-green-100';
      case 'running': return 'text-blue-600 bg-blue-100';
      case 'failed': return 'text-red-600 bg-red-100';
      case 'pending': return 'text-yellow-600 bg-yellow-100';
      case 'cancelled': return 'text-gray-600 bg-gray-100';
      default: return 'text-gray-600 bg-gray-100';
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
      <div className={`${className}`}>
        <div className="animate-pulse space-y-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-16 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`${className}`}>
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center gap-2">
            <span className="text-red-600">❌</span>
            <span className="text-red-800 font-medium">Failed to load command history</span>
          </div>
          <div className="mt-1 text-red-700 text-sm">{error}</div>
        </div>
      </div>
    );
  }

  if (executions.length === 0) {
    return (
      <div className={`${className}`}>
        <div className="text-center py-8">
          <div className="text-gray-500 dark:text-gray-400">
            <svg
              className="mx-auto h-12 w-12 mb-4 opacity-50"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
              />
            </svg>
            <p className="text-lg font-medium mb-2">No command history</p>
            <p className="text-sm">
              Command executions will appear here once you start running commands.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-3 ${className}`}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-medium text-gray-900 dark:text-white">
          Command History
        </h3>
        <span className="text-sm text-gray-500 dark:text-gray-400">
          {executions.length} execution{executions.length !== 1 ? 's' : ''}
        </span>
      </div>

      {executions.map((execution) => (
        <div
          key={execution.id}
          className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden hover:shadow-md transition-shadow"
        >
          {/* Header with status and timestamp (AC: 3) */}
          <div className="px-4 py-3 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(execution.status)}`}>
                  {getStatusIcon(execution.status)} {execution.status.toUpperCase()}
                </span>
                <span className="text-sm text-gray-600 dark:text-gray-400">
                  ID: {execution.id}
                </span>
              </div>
              <div className="text-sm text-gray-500 dark:text-gray-400">
                {formatDistanceToNow(execution.startedAt, { addSuffix: true })}
              </div>
            </div>
          </div>

          {/* Command and details */}
          <div className="p-4">
            <div className="mb-3">
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Command:</div>
              <div className="bg-gray-900 text-green-400 rounded-md p-2 font-mono text-sm overflow-x-auto">
                {execution.command}
              </div>
            </div>

            {/* Execution metadata (AC: 3) */}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-gray-600 dark:text-gray-400">Executed by:</span>
                <span className="ml-2 font-medium text-gray-900 dark:text-white">
                  {execution.executedBy}
                </span>
              </div>
              <div>
                <span className="text-gray-600 dark:text-gray-400">Duration:</span>
                <span className="ml-2 font-medium text-gray-900 dark:text-white">
                  {execution.completedAt
                    ? `${execution.completedAt.getTime() - execution.startedAt.getTime()}ms`
                    : 'In progress'}
                </span>
              </div>
              {execution.approvedBy && (
                <div className="col-span-2">
                  <span className="text-gray-600 dark:text-gray-400">Approved by:</span>
                  <span className="ml-2 font-medium text-gray-900 dark:text-white">
                    {execution.approvedBy}
                  </span>
                </div>
              )}
            </div>

            {/* Result preview */}
            {execution.output && (
              <div className="mt-3">
                <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Output preview:</div>
                <div className="bg-gray-100 dark:bg-gray-700 rounded-md p-2 text-sm overflow-hidden">
                  <div className="line-clamp-3 text-gray-700 dark:text-gray-300">
                    {execution.output.length > 200
                      ? `${execution.output.substring(0, 200)}...`
                      : execution.output}
                  </div>
                </div>
              </div>
            )}

            {/* Error display */}
            {execution.error && (
              <div className="mt-3">
                <div className="text-sm text-red-600 mb-1">Error:</div>
                <div className="bg-red-50 border border-red-200 rounded-md p-2 text-sm">
                  <div className="text-red-800">
                    {execution.error.length > 200
                      ? `${execution.error.substring(0, 200)}...`
                      : execution.error}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
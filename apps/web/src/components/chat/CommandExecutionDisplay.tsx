import { useState } from 'react';
import { api } from '@/services/api';
import { CommandProgressTracker } from './CommandProgressTracker';
import { CommandExecutionLoader } from './LoadingIndicator';

interface CommandExecutionDisplayProps {
  command: string;
  safetyLevel: 'safe' | 'warning' | 'dangerous';
  confidence?: number;
  explanation?: string;
  potentialImpact?: string[];
  executionId?: string;
}

export function CommandExecutionDisplay({
  command,
  safetyLevel,
  confidence,
  explanation,
  potentialImpact,
  executionId
}: CommandExecutionDisplayProps) {
  const [isExecuting, setIsExecuting] = useState(false);
  const [executionResult, setExecutionResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);

  const handleExecuteCommand = async () => {
    if (isExecuting) return;

    setIsExecuting(true);
    setError(null);

    try {
      // Execute command via /commands/execute API (AC: 2)
      const response = await api.commands.execute({
        command,
        cluster: 'default'
      });

      setExecutionResult(response.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Command execution failed');
    } finally {
      setIsExecuting(false);
    }
  };

  const getSafetyBorderColor = () => {
    switch (safetyLevel) {
      case 'dangerous': return 'border-red-500';
      case 'warning': return 'border-yellow-500';
      case 'safe': return 'border-green-500';
      default: return 'border-gray-300';
    }
  };

  const getSafetyBgColor = () => {
    switch (safetyLevel) {
      case 'dangerous': return 'bg-red-50 dark:bg-red-900/20';
      case 'warning': return 'bg-yellow-50 dark:bg-yellow-900/20';
      case 'safe': return 'bg-green-50 dark:bg-green-900/20';
      default: return 'bg-gray-50 dark:bg-gray-900/20';
    }
  };

  return (
    <div className={`mt-3 border rounded-lg ${getSafetyBorderColor()} ${getSafetyBgColor()}`}>
      {/* Command Header */}
      <div className="px-3 py-2 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between">
          <h4 className="font-medium text-gray-900 dark:text-white">
            Generated Command
          </h4>
          {confidence && (
            <span className="text-xs text-gray-500 dark:text-gray-400">
              Confidence: {Math.round(confidence * 100)}%
            </span>
          )}
        </div>
      </div>

      {/* Command Display with Syntax Highlighting (AC: 2) */}
      <div className="p-3">
        <div className="bg-gray-900 text-green-400 rounded-md p-3 font-mono text-sm overflow-x-auto">
          <div className="flex items-center justify-between mb-2">
            <span className="text-gray-500">$</span>
            <button
              onClick={() => navigator.clipboard?.writeText(command)}
              className="text-xs text-gray-400 hover:text-white transition-colors"
              title="Copy command"
            >
              📋 Copy
            </button>
          </div>
          <div className="whitespace-pre-wrap break-all">
            <SyntaxHighlightedCommand command={command} />
          </div>
        </div>

        {/* Explanation */}
        {explanation && (
          <div className="mt-3 text-sm text-gray-700 dark:text-gray-300">
            <strong>Explanation:</strong> {explanation}
          </div>
        )}

        {/* Potential Impact (AC: 2) */}
        {potentialImpact && potentialImpact.length > 0 && (
          <div className="mt-3">
            <strong className="text-sm text-gray-900 dark:text-white">Potential Impact:</strong>
            <ul className="mt-1 text-sm text-gray-700 dark:text-gray-300 space-y-1">
              {potentialImpact.map((impact, index) => (
                <li key={index} className="flex items-start">
                  <span className="text-gray-400 mr-2">•</span>
                  <span>{impact}</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Execution Controls (AC: 2, 4) */}
        <div className="mt-4 flex items-center gap-3">
          <button
            onClick={handleExecuteCommand}
            disabled={isExecuting || safetyLevel === 'dangerous'}
            className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              safetyLevel === 'dangerous'
                ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                : isExecuting
                ? 'bg-blue-300 text-blue-700 cursor-wait'
                : 'bg-blue-600 text-white hover:bg-blue-700'
            }`}
          >
            {isExecuting ? (
              <span className="flex items-center gap-2">
                <div className="w-3 h-3 border border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                Executing...
              </span>
            ) : safetyLevel === 'dangerous' ? (
              'Requires Approval'
            ) : (
              'Execute Command'
            )}
          </button>

          {safetyLevel === 'dangerous' && (
            <span className="text-xs text-red-600 dark:text-red-400">
              This command requires manual approval due to safety concerns
            </span>
          )}
        </div>

        {/* Execution Error Display */}
        {error && (
          <div className="mt-3 p-3 bg-red-100 border border-red-300 rounded-md">
            <div className="flex items-center gap-2">
              <span className="text-red-600">❌</span>
              <span className="text-red-800 text-sm font-medium">Execution Failed</span>
            </div>
            <div className="mt-1 text-red-700 text-sm">{error}</div>
          </div>
        )}

        {/* Progress Tracking for Executing Commands (AC: 4, 6) */}
        {isExecuting && executionResult?.id && (
          <CommandProgressTracker
            executionId={executionResult.id}
            onProgressUpdate={(phase, progress) => {
              console.log(`Command ${executionResult.id}: ${phase} - ${progress}%`);
            }}
            onComplete={(result) => {
              setIsExecuting(false);
              setExecutionResult(result);
            }}
            onError={(error) => {
              setIsExecuting(false);
              setError(error);
            }}
          />
        )}

        {/* Simple Loading for Command Execution */}
        {isExecuting && !executionResult?.id && (
          <CommandExecutionLoader message="Submitting command for execution..." />
        )}

        {/* Execution Result Preview */}
        {executionResult && !isExecuting && (
          <div className="mt-3 p-3 bg-gray-100 dark:bg-gray-800 rounded-md">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-green-600">✅</span>
              <span className="text-gray-900 dark:text-white text-sm font-medium">
                Command Executed Successfully
              </span>
            </div>
            <div className="text-xs text-gray-600 dark:text-gray-400">
              Execution ID: {executionResult.id}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Simple syntax highlighting for kubectl commands
function SyntaxHighlightedCommand({ command }: { command: string }) {
  const parts = command.split(/(\s+)/);

  return (
    <>
      {parts.map((part, index) => {
        const trimmed = part.trim();

        if (trimmed === 'kubectl') {
          return <span key={index} className="text-blue-400 font-semibold">{part}</span>;
        } else if (['get', 'describe', 'logs', 'exec', 'delete', 'create', 'apply', 'scale', 'restart'].includes(trimmed)) {
          return <span key={index} className="text-yellow-400">{part}</span>;
        } else if (trimmed.startsWith('-')) {
          return <span key={index} className="text-purple-400">{part}</span>;
        } else if (trimmed.includes('/') || trimmed.includes(':')) {
          return <span key={index} className="text-cyan-400">{part}</span>;
        } else {
          return <span key={index} className="text-green-400">{part}</span>;
        }
      })}
    </>
  );
}
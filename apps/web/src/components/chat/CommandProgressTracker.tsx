import { useState, useEffect } from 'react';
import { api } from '@/services/api';
import { useCommandExecutionTracking } from '@/services/realTimeService';

export type CommandPhase = 'parsing' | 'validation' | 'execution' | 'completion';

interface CommandProgressTrackerProps {
  executionId: string;
  onProgressUpdate?: (phase: CommandPhase, progress: number) => void;
  onComplete?: (result: any) => void;
  onError?: (error: string) => void;
}

export function CommandProgressTracker({
  executionId,
  onProgressUpdate,
  onComplete,
  onError
}: CommandProgressTrackerProps) {
  // Use real-time WebSocket tracking (AC: 7)
  const {
    status: realTimeStatus,
    progress: realTimeProgress,
    phase: realTimePhase,
    output: realTimeOutput,
    error: realTimeError,
    isConnected
  } = useCommandExecutionTracking(executionId);

  const [currentPhase, setCurrentPhase] = useState<CommandPhase>('parsing');
  const [progress, setProgress] = useState(0);
  const [isComplete, setIsComplete] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [executionTime, setExecutionTime] = useState(0);
  const [startTime] = useState(Date.now());

  // Update local state based on real-time WebSocket data (AC: 7)
  useEffect(() => {
    if (realTimeStatus) {
      if (realTimeStatus === 'completed' || realTimeStatus === 'failed') {
        setIsComplete(true);
        if (realTimeStatus === 'completed') {
          setProgress(100);
          setCurrentPhase('completion');
          onComplete?.({ executionId, status: realTimeStatus, output: realTimeOutput });
        }
      }
    }
  }, [realTimeStatus, executionId, onComplete, realTimeOutput]);

  useEffect(() => {
    if (realTimeProgress !== undefined) {
      setProgress(realTimeProgress);
      onProgressUpdate?.(currentPhase, realTimeProgress);
    }
  }, [realTimeProgress, currentPhase, onProgressUpdate]);

  useEffect(() => {
    if (realTimePhase) {
      setCurrentPhase(realTimePhase as CommandPhase);
    }
  }, [realTimePhase]);

  useEffect(() => {
    if (realTimeError) {
      setError(realTimeError);
      onError?.(realTimeError);
    }
  }, [realTimeError, onError]);

  // Update execution time
  useEffect(() => {
    const interval = setInterval(() => {
      setExecutionTime(Date.now() - startTime);
    }, 1000);

    return () => clearInterval(interval);
  }, [startTime]);

  const getPhaseDescription = (phase: CommandPhase) => {
    switch (phase) {
      case 'parsing': return 'Parsing command syntax...';
      case 'validation': return 'Validating permissions and safety...';
      case 'execution': return 'Executing command on cluster...';
      case 'completion': return 'Processing results...';
      default: return 'Processing...';
    }
  };

  const getPhaseIcon = (phase: CommandPhase) => {
    switch (phase) {
      case 'parsing': return '📝';
      case 'validation': return '🔍';
      case 'execution': return '⚡';
      case 'completion': return '✅';
      default: return '⏳';
    }
  };

  if (error) {
    return (
      <div className="mt-3 p-3 bg-red-50 border border-red-200 rounded-lg">
        <div className="flex items-center gap-2">
          <span className="text-red-600">❌</span>
          <span className="text-red-800 font-medium">Command Progress Tracking Failed</span>
        </div>
        <div className="mt-1 text-red-700 text-sm">{error}</div>
      </div>
    );
  }

  if (isComplete) {
    return (
      <div className="mt-3 p-3 bg-green-50 border border-green-200 rounded-lg">
        <div className="flex items-center gap-2">
          <span className="text-green-600">✅</span>
          <span className="text-green-800 font-medium">Command Completed</span>
        </div>
        <div className="mt-1 text-green-700 text-sm">
          Execution time: {Math.round(executionTime / 1000)}s
        </div>
      </div>
    );
  }

  return (
    <div className="mt-3 p-3 bg-blue-50 border border-blue-200 rounded-lg">
      {/* Progress Header with WebSocket Status (AC: 7) */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-blue-600">{getPhaseIcon(currentPhase)}</span>
          <span className="text-blue-800 font-medium">
            {getPhaseDescription(currentPhase)}
          </span>
          {!isConnected && (
            <span className="text-xs text-orange-600 bg-orange-100 px-2 py-1 rounded-full">
              ⚠️ Real-time updates disconnected
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-blue-600">
            {Math.round(progress)}%
          </span>
          <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-orange-500'}`} title={isConnected ? 'Real-time connected' : 'Real-time disconnected'}></div>
        </div>
      </div>

      {/* Progress Bar (AC: 4) */}
      <div className="mb-3">
        <div className="w-full bg-blue-200 rounded-full h-2">
          <div
            className="bg-blue-600 h-2 rounded-full transition-all duration-300 ease-out"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
      </div>

      {/* Execution Metrics (AC: 4) */}
      <div className="grid grid-cols-3 gap-4 text-xs text-blue-700">
        <div>
          <span className="font-medium">Phase:</span>
          <div className="capitalize">{currentPhase}</div>
        </div>
        <div>
          <span className="font-medium">Time:</span>
          <div>{Math.round(executionTime / 1000)}s</div>
        </div>
        <div>
          <span className="font-medium">Progress:</span>
          <div>{Math.round(progress)}%</div>
        </div>
      </div>

      {/* Phase Indicators */}
      <div className="mt-3 flex items-center justify-between">
        {(['parsing', 'validation', 'execution', 'completion'] as CommandPhase[]).map((phase, index) => {
          const isActive = phase === currentPhase;
          const isCompleted = ['parsing', 'validation', 'execution', 'completion'].indexOf(currentPhase) > index;

          return (
            <div key={phase} className="flex items-center">
              <div className={`w-3 h-3 rounded-full ${
                isCompleted
                  ? 'bg-green-500'
                  : isActive
                  ? 'bg-blue-500 animate-pulse'
                  : 'bg-gray-300'
              }`}></div>
              <span className={`ml-1 text-xs ${
                isActive ? 'text-blue-700 font-medium' : 'text-gray-500'
              }`}>
                {phase}
              </span>
            </div>
          );
        })}
      </div>

      {/* Cancellation Button (AC: 4) */}
      <div className="mt-3 flex justify-end">
        <button
          onClick={() => {
            // TODO: Implement command cancellation via API
            setError('Command cancellation requested');
            onError?.('Command cancelled by user');
          }}
          className="text-xs text-red-600 hover:text-red-800 transition-colors"
        >
          Cancel Command
        </button>
      </div>
    </div>
  );
}
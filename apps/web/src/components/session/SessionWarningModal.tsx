// Session Warning Modal Component
// Following coding standards from docs/architecture/coding-standards.md

import React, { useEffect, useState } from 'react';
import { sessionManager, SessionWarning } from '../../services/sessionManager';

interface SessionWarningModalProps {
  warning: SessionWarning | null;
  onExtend: () => void;
  onLogout: () => void;
  onDismiss: () => void;
}

const SessionWarningModal: React.FC<SessionWarningModalProps> = ({
  warning,
  onExtend,
  onLogout,
  onDismiss
}) => {
  const [timeRemaining, setTimeRemaining] = useState(0);

  useEffect(() => {
    if (!warning) return;

    setTimeRemaining(warning.timeRemaining);

    const interval = setInterval(() => {
      setTimeRemaining(prev => {
        const newTime = prev - 1000;
        if (newTime <= 0) {
          clearInterval(interval);
          onLogout();
        }
        return newTime;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [warning, onLogout]);

  if (!warning) return null;

  const formatTime = (ms: number) => {
    const minutes = Math.floor(ms / 60000);
    const seconds = Math.floor((ms % 60000) / 1000);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  const getWarningTitle = (type: string) => {
    switch (type) {
      case 'idle':
        return '⏰ Session Idle Warning';
      case 'absolute':
        return '⏰ Session Expiring';
      case 'refresh':
        return '🔄 Token Refresh Required';
      default:
        return '⚠️ Session Warning';
    }
  };

  const getWarningMessage = (type: string) => {
    switch (type) {
      case 'idle':
        return 'Your session will expire due to inactivity.';
      case 'absolute':
        return 'Your session has reached the maximum allowed time.';
      case 'refresh':
        return 'Your authentication token needs to be refreshed.';
      default:
        return 'Your session will expire soon.';
    }
  };

  const getWarningColor = (type: string) => {
    switch (type) {
      case 'idle':
        return 'border-yellow-400 bg-yellow-50';
      case 'absolute':
        return 'border-red-400 bg-red-50';
      case 'refresh':
        return 'border-blue-400 bg-blue-50';
      default:
        return 'border-gray-400 bg-gray-50';
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className={`bg-white rounded-lg shadow-xl border-2 p-6 max-w-md w-full mx-4 ${getWarningColor(warning.type)}`}>
        <div className="flex items-center space-x-3 mb-4">
          <div className="flex-shrink-0">
            <div className="w-10 h-10 bg-yellow-100 rounded-full flex items-center justify-center">
              <svg className="w-6 h-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            </div>
          </div>
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-gray-900">
              {getWarningTitle(warning.type)}
            </h3>
            <p className="text-sm text-gray-600">
              {getWarningMessage(warning.type)}
            </p>
          </div>
        </div>

        <div className="bg-white rounded-lg p-4 mb-4 border">
          <div className="text-center">
            <div className="text-3xl font-bold text-gray-900 mb-1">
              {formatTime(timeRemaining)}
            </div>
            <div className="text-sm text-gray-600">
              Time remaining
            </div>
          </div>
        </div>

        {warning.type === 'refresh' && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-blue-800">
              <strong>Note:</strong> Your session will be automatically refreshed. You can continue working without interruption.
            </p>
          </div>
        )}

        <div className="flex space-x-3">
          {warning.type !== 'absolute' && (
            <button
              onClick={onExtend}
              className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
            >
              {warning.type === 'refresh' ? 'Refresh Now' : 'Extend Session'}
            </button>
          )}
          <button
            onClick={onLogout}
            className="flex-1 bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
          >
            Logout Now
          </button>
          {warning.type === 'refresh' && (
            <button
              onClick={onDismiss}
              className="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors"
            >
              Dismiss
            </button>
          )}
        </div>

        <div className="mt-4 text-xs text-gray-500 text-center">
          For security, your session will automatically logout when the timer reaches zero.
        </div>
      </div>
    </div>
  );
};

// Hook to use session warnings
export const useSessionWarnings = () => {
  const [warning, setWarning] = useState<SessionWarning | null>(null);

  useEffect(() => {
    sessionManager.setWarningCallback((warning: SessionWarning) => {
      setWarning(warning);
    });

    return () => {
      sessionManager.setWarningCallback(() => {});
    };
  }, []);

  const handleExtend = () => {
    if (warning?.callback) {
      warning.callback();
    }
    setWarning(null);
  };

  const handleLogout = () => {
    sessionManager.logout('user_logout');
    setWarning(null);
  };

  const handleDismiss = () => {
    setWarning(null);
  };

  return {
    warning,
    handleExtend,
    handleLogout,
    handleDismiss
  };
};

export default SessionWarningModal;
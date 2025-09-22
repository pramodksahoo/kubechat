// Session Provider Component
// Following coding standards from docs/architecture/coding-standards.md

import React, { useEffect } from 'react';
import { sessionManager } from '../../services/sessionManager';
import { useAuthStore } from '../../stores/authStore';
import SessionWarningModal, { useSessionWarnings } from './SessionWarningModal';

interface SessionProviderProps {
  children: React.ReactNode;
}

const SessionProvider: React.FC<SessionProviderProps> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  const { warning, handleExtend, handleLogout, handleDismiss } = useSessionWarnings();

  useEffect(() => {
    if (isAuthenticated) {
      // Start session management when user is authenticated
      sessionManager.startSession();
      
      // Validate session periodically
      const validationInterval = setInterval(() => {
        sessionManager.validateSession();
      }, 60000); // Check every minute

      return () => {
        clearInterval(validationInterval);
      };
    } else {
      // Cleanup session when user is not authenticated
      sessionManager.cleanup();
    }
  }, [isAuthenticated]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      sessionManager.cleanup();
    };
  }, []);

  return (
    <>
      {children}
      <SessionWarningModal
        warning={warning}
        onExtend={handleExtend}
        onLogout={handleLogout}
        onDismiss={handleDismiss}
      />
    </>
  );
};

export default SessionProvider;
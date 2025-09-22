// Authentication initialization hook
// Handles SSR/hydration and ensures proper auth state on page refresh

import { useEffect, useState } from 'react';
import { useAuthStore, initializeAuth } from '../stores/authStore';

export interface AuthInitializationState {
  isInitializing: boolean;
  isInitialized: boolean;
  initializationError: string | null;
}

export const useAuthInitialization = (): AuthInitializationState => {
  const [state, setState] = useState<AuthInitializationState>({
    isInitializing: true,
    isInitialized: false,
    initializationError: null
  });

  const { isLoading } = useAuthStore();

  useEffect(() => {
    let isMounted = true;

    const initAuth = async () => {
      // Skip during SSR
      if (typeof window === 'undefined') {
        return;
      }

      try {
        setState(prev => ({ 
          ...prev, 
          isInitializing: true, 
          initializationError: null 
        }));

        await initializeAuth();

        if (isMounted) {
          setState(prev => ({
            ...prev,
            isInitializing: false,
            isInitialized: true,
            initializationError: null
          }));
        }
      } catch (error) {
        console.error('Auth initialization failed:', error);
        
        if (isMounted) {
          setState(prev => ({
            ...prev,
            isInitializing: false,
            isInitialized: true, // Still mark as initialized even if failed
            initializationError: error instanceof Error ? error.message : 'Authentication initialization failed'
          }));
        }
      }
    };

    initAuth();

    return () => {
      isMounted = false;
    };
  }, []);

  // Consider initialization complete when either:
  // 1. We're done initializing and not loading, or
  // 2. We're on the server side
  const isComplete = state.isInitialized && !isLoading;

  return {
    ...state,
    isInitializing: state.isInitializing || isLoading,
    isInitialized: isComplete
  };
};
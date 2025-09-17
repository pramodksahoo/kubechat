// Protected Route Component
// Following coding standards from docs/architecture/coding-standards.md

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import { useAuthStore } from '../../stores/authStore';

export interface ProtectedRouteProps {
  children: React.ReactNode;
  requireAuth?: boolean;
  requiredRoles?: string[];
  redirectTo?: string;
  fallback?: React.ReactNode;
}

// Loading spinner component
const LoadingSpinner: React.FC = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="flex flex-col items-center space-y-4">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      <p className="text-gray-600">Loading...</p>
    </div>
  </div>
);

// Access denied component
const AccessDenied: React.FC<{ requiredRoles?: string[] }> = ({ requiredRoles }) => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="text-center space-y-4">
      <div className="w-16 h-16 mx-auto bg-red-100 rounded-full flex items-center justify-center">
        <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
      </div>
      <h2 className="text-xl font-bold text-gray-900">Access Denied</h2>
      <p className="text-gray-600">
        {requiredRoles
          ? `You need one of the following roles: ${requiredRoles.join(', ')}`
          : 'You do not have permission to access this page'
        }
      </p>
      <button
        onClick={() => window.history.back()}
        className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
      >
        Go Back
      </button>
    </div>
  </div>
);

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requireAuth = true,
  requiredRoles = [],
  redirectTo = '/auth/login',
  fallback
}) => {
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuthStore();
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    // Initialize auth store if needed
    const initAuth = async () => {
      try {
        const { initializeAuth } = await import('../../stores/authStore');
        await initializeAuth();
      } catch (error) {
        console.error('Auth initialization failed:', error);
      } finally {
        setIsInitialized(true);
      }
    };

    if (!isInitialized) {
      initAuth();
    }
  }, [isInitialized]);

  useEffect(() => {
    if (!isInitialized || isLoading) return;

    // If authentication is required but user is not authenticated
    if (requireAuth && !isAuthenticated) {
      const currentPath = router.asPath;
      const returnUrl = currentPath !== '/' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
      router.replace(`${redirectTo}${returnUrl}`);
      return;
    }

    // Check role-based access
    if (requireAuth && isAuthenticated && requiredRoles.length > 0 && user) {
      const userRole = user.role;
      const hasRequiredRole = requiredRoles.includes(userRole);

      if (!hasRequiredRole) {
        // User doesn't have required role - show access denied
        return;
      }
    }
  }, [isInitialized, isAuthenticated, isLoading, requireAuth, requiredRoles, user, router, redirectTo]);

  // Show loading state
  if (!isInitialized || isLoading) {
    return fallback || <LoadingSpinner />;
  }

  // If auth is not required, render children
  if (!requireAuth) {
    return <>{children}</>;
  }

  // If not authenticated, don't render anything (redirect will happen)
  if (!isAuthenticated) {
    return fallback || <LoadingSpinner />;
  }

  // Check role-based access
  if (requiredRoles.length > 0 && user) {
    const userRole = user.role;
    const hasRequiredRole = requiredRoles.includes(userRole);

    if (!hasRequiredRole) {
      return <AccessDenied requiredRoles={requiredRoles} />;
    }
  }

  // All checks passed, render children
  return <>{children}</>;
};

// Higher-order component version
export const withAuth = <P extends object>(
  Component: React.ComponentType<P>,
  options: Omit<ProtectedRouteProps, 'children'> = {}
) => {
  const WrappedComponent: React.FC<P> = (props) => (
    <ProtectedRoute {...options}>
      <Component {...props} />
    </ProtectedRoute>
  );

  WrappedComponent.displayName = `withAuth(${Component.displayName || Component.name})`;

  return WrappedComponent;
};

// Hook for checking authentication status
export const useRequireAuth = (options: {
  requireAuth?: boolean;
  requiredRoles?: string[];
  redirectTo?: string;
} = {}) => {
  const {
    requireAuth = true,
    requiredRoles = [],
    redirectTo = '/auth/login'
  } = options;

  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuthStore();

  useEffect(() => {
    if (isLoading) return;

    if (requireAuth && !isAuthenticated) {
      const currentPath = router.asPath;
      const returnUrl = currentPath !== '/' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
      router.replace(`${redirectTo}${returnUrl}`);
      return;
    }

    if (requireAuth && isAuthenticated && requiredRoles.length > 0 && user) {
      const userRole = user.role;
      const hasRequiredRole = requiredRoles.includes(userRole);

      if (!hasRequiredRole) {
        router.replace('/403'); // Forbidden page
        return;
      }
    }
  }, [isAuthenticated, isLoading, requireAuth, requiredRoles, user, router, redirectTo]);

  return {
    isAuthenticated,
    isLoading,
    user,
    hasAccess: !requireAuth || (isAuthenticated && (requiredRoles.length === 0 || (user && requiredRoles.includes(user.role))))
  };
};

export default ProtectedRoute;
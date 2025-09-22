// Login page
// Following coding standards from docs/architecture/coding-standards.md

import React, { useEffect } from 'react';
import { useRouter } from 'next/router';
import { LoginForm } from '../../components/auth';
import { useAuthStore } from '../../stores/authStore';

const LoginPage: React.FC = () => {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    // Redirect authenticated users based on their role
    if (isAuthenticated) {
      const returnUrl = router.query.returnUrl as string;
      if (returnUrl) {
        router.replace(returnUrl);
      } else {
        // Redirect based on user role: admin users to /admin, regular users to /
        const { user } = useAuthStore.getState();
        const defaultRedirect = user?.role === 'admin' ? '/admin' : '/';
        router.replace(defaultRedirect);
      }
    }
  }, [isAuthenticated, router]);

  const handleLoginSuccess = () => {
    const returnUrl = router.query.returnUrl as string;
    if (returnUrl) {
      router.push(returnUrl);
    } else {
      // Role-based redirect will be handled by LoginForm component
      // No need to redirect here as LoginForm handles it
    }
  };

  if (isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <LoginForm onSuccess={handleLoginSuccess} />
      </div>
    </div>
  );
};

export default LoginPage;
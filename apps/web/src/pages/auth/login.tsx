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
    // Redirect authenticated users to dashboard
    if (isAuthenticated) {
      const returnUrl = router.query.returnUrl as string;
      router.replace(returnUrl || '/dashboard');
    }
  }, [isAuthenticated, router]);

  const handleLoginSuccess = () => {
    const returnUrl = router.query.returnUrl as string;
    router.push(returnUrl || '/dashboard');
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
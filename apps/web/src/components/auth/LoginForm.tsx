// Login Form Component
// Following coding standards from docs/architecture/coding-standards.md

import React, { useState } from 'react';
import { useRouter } from 'next/router';
import { useAuthStore } from '../../stores/authStore';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Card } from '../ui/Card';
import type { LoginCredentials } from '../../types/auth';

export interface LoginFormProps {
  onSuccess?: () => void;
  redirectTo?: string;
}

export const LoginForm: React.FC<LoginFormProps> = ({
  onSuccess,
  redirectTo
}) => {
  const router = useRouter();
  const { login, isLoading, error, clearError } = useAuthStore();

  const [credentials, setCredentials] = useState<LoginCredentials>({
    username: '',
    password: ''
  });

  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!credentials.username.trim()) {
      errors.username = 'Username is required';
    } else if (credentials.username.length < 3) {
      errors.username = 'Username must be at least 3 characters';
    }

    if (!credentials.password) {
      errors.password = 'Password is required';
    } else if (credentials.password.length < 6) {
      errors.password = 'Password must be at least 6 characters';
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();

    if (!validateForm()) {
      return;
    }

    try {
      await login(credentials);
      
      // Get the updated user info after login
      const { user } = useAuthStore.getState();
      
      // Determine redirect based on user role if no explicit redirectTo is provided
      let finalRedirect = redirectTo;
      if (!finalRedirect) {
        finalRedirect = user?.role === 'admin' ? '/admin' : '/';
      }
      
      onSuccess?.();
      router.push(finalRedirect);
    } catch (error) {
      // Error is handled by the store
      console.error('Login failed:', error);
    }
  };

  const handleInputChange = (field: keyof LoginCredentials) => (
    value: string
  ) => {
    setCredentials(prev => ({
      ...prev,
      [field]: value
    }));

    // Clear field-specific validation error
    if (validationErrors[field]) {
      setValidationErrors(prev => ({
        ...prev,
        [field]: ''
      }));
    }
  };

  return (
    <Card className="w-full max-w-md mx-auto">
      <div className="p-6 space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900">Sign In</h1>
          <p className="text-sm text-gray-600 mt-2">
            Access your KubeChat dashboard
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
              Username
            </label>
            <Input
              id="username"
              type="text"
              value={credentials.username}
              onChange={handleInputChange('username')}
              placeholder="Enter your username"
              disabled={isLoading}
              aria-describedby={validationErrors.username ? 'username-error' : undefined}
              className={validationErrors.username ? 'border-red-500' : ''}
            />
            {validationErrors.username && (
              <p id="username-error" className="text-sm text-red-600 mt-1">
                {validationErrors.username}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              Password
            </label>
            <Input
              id="password"
              type="password"
              value={credentials.password}
              onChange={handleInputChange('password')}
              placeholder="Enter your password"
              disabled={isLoading}
              aria-describedby={validationErrors.password ? 'password-error' : undefined}
              className={validationErrors.password ? 'border-red-500' : ''}
            />
            {validationErrors.password && (
              <p id="password-error" className="text-sm text-red-600 mt-1">
                {validationErrors.password}
              </p>
            )}
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <Button
            type="submit"
            disabled={isLoading}
            className="w-full"
          >
            {isLoading ? 'Signing in...' : 'Sign In'}
          </Button>
        </form>

        <div className="text-center">
          <p className="text-sm text-gray-600">
            Don't have an account?{' '}
            <button
              type="button"
              onClick={() => router.push('/auth/register')}
              className="font-medium text-blue-600 hover:text-blue-500 focus:outline-none focus:underline"
            >
              Sign up
            </button>
          </p>
        </div>
      </div>
    </Card>
  );
};

export default LoginForm;
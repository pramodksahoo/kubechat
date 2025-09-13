import React from 'react';
import { ButtonProps } from '@kubechat/shared/types';

export const Button: React.FC<ButtonProps> = ({
  variant = 'primary',
  size = 'md',
  disabled = false,
  loading = false,
  onClick,
  type = 'button',
  className = '',
  children,
  'data-testid': dataTestId = 'button'
}) => {
  const baseClasses = 'inline-flex items-center justify-center font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors duration-200 disabled:cursor-not-allowed disabled:opacity-50';

  const variantClasses = {
    primary: 'bg-primary-600 hover:bg-primary-700 focus:ring-primary-500 text-white shadow-sm',
    secondary: 'bg-secondary-100 hover:bg-secondary-200 focus:ring-secondary-500 text-secondary-900 border border-secondary-300 dark:bg-secondary-800 dark:text-secondary-100 dark:border-secondary-600 dark:hover:bg-secondary-700',
    danger: 'bg-danger-600 hover:bg-danger-700 focus:ring-danger-500 text-white shadow-sm',
    ghost: 'hover:bg-gray-100 focus:ring-gray-500 text-gray-700 dark:text-gray-300 dark:hover:bg-gray-800'
  };

  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-sm',
    lg: 'px-6 py-3 text-base'
  };

  const classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`;

  return (
    <button
      type={type}
      className={classes}
      disabled={disabled || loading}
      onClick={onClick}
      data-testid={dataTestId}
    >
      {loading && (
        <svg 
          className="animate-spin -ml-1 mr-2 h-4 w-4" 
          xmlns="http://www.w3.org/2000/svg" 
          fill="none" 
          viewBox="0 0 24 24"
        >
          <circle 
            className="opacity-25" 
            cx="12" 
            cy="12" 
            r="10" 
            stroke="currentColor" 
            strokeWidth="4"
          />
          <path 
            className="opacity-75" 
            fill="currentColor" 
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  );
};
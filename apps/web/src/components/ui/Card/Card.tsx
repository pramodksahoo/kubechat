import React from 'react';
import { CardProps } from '@kubechat/shared/types';

export const Card: React.FC<CardProps> = ({
  title,
  description,
  variant = 'default',
  padding = 'md',
  className = '',
  children,
  'data-testid': dataTestId = 'card'
}) => {
  const baseClasses = 'bg-white dark:bg-gray-900 rounded-lg border shadow-sm';

  const variantClasses = {
    default: 'border-gray-200 dark:border-gray-700',
    highlighted: 'border-primary-200 dark:border-primary-800 bg-primary-50 dark:bg-primary-950',
    error: 'border-danger-200 dark:border-danger-800 bg-danger-50 dark:bg-danger-950',
    warning: 'border-warning-200 dark:border-warning-800 bg-warning-50 dark:bg-warning-950',
    success: 'border-success-200 dark:border-success-800 bg-success-50 dark:bg-success-950'
  };

  const paddingClasses = {
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8'
  };

  const classes = `${baseClasses} ${variantClasses[variant]} ${paddingClasses[padding]} ${className}`;

  return (
    <div className={classes} data-testid={dataTestId}>
      {(title || description) && (
        <div className="mb-4">
          {title && (
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              {title}
            </h3>
          )}
          {description && (
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {description}
            </p>
          )}
        </div>
      )}
      {children}
    </div>
  );
};
import React from 'react';
import { InputProps } from '@kubechat/shared/types';

export const Input: React.FC<InputProps> = ({
  id,
  label,
  placeholder,
  value,
  onChange,
  type = 'text',
  error,
  disabled = false,
  required = false,
  className = '',
  'data-testid': dataTestId = 'input'
}) => {
  const baseClasses = 'block w-full rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors duration-200 disabled:cursor-not-allowed disabled:opacity-50';
  
  const errorClasses = error 
    ? 'border-danger-300 dark:border-danger-600 focus:ring-danger-500 focus:border-danger-500' 
    : 'border-gray-300 dark:border-gray-600 focus:ring-primary-500 focus:border-primary-500';
  
  const backgroundClasses = 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400';
  
  const classes = `${baseClasses} ${errorClasses} ${backgroundClasses} ${className}`;

  return (
    <div>
      {label && (
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          {label}
          {required && <span className="text-danger-500 ml-1">*</span>}
        </label>
      )}
      
      <input
        id={id}
        type={type}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        disabled={disabled}
        required={required}
        className={classes}
        data-testid={dataTestId}
      />
      
      {error && (
        <p className="mt-1 text-sm text-danger-600 dark:text-danger-400" data-testid={`${dataTestId}-error`}>
          {error}
        </p>
      )}
    </div>
  );
};
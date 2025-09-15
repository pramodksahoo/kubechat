import React from 'react';

export type StatusType = 'healthy' | 'warning' | 'error' | 'offline' | 'connecting' | 'unknown';

export interface StatusIndicatorProps {
  status: StatusType;
  label: string;
  description?: string;
  showLabel?: boolean;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  onClick?: () => void;
  lastUpdated?: string;
}

const statusConfig = {
  healthy: {
    color: 'text-green-600 dark:text-green-400',
    bgColor: 'bg-green-100 dark:bg-green-900/20',
    borderColor: 'border-green-200 dark:border-green-800',
    dotColor: 'bg-green-500',
    icon: 'check-circle',
    pulse: true,
  },
  warning: {
    color: 'text-yellow-600 dark:text-yellow-400',
    bgColor: 'bg-yellow-100 dark:bg-yellow-900/20',
    borderColor: 'border-yellow-200 dark:border-yellow-800',
    dotColor: 'bg-yellow-500',
    icon: 'exclamation-triangle',
    pulse: false,
  },
  error: {
    color: 'text-red-600 dark:text-red-400',
    bgColor: 'bg-red-100 dark:bg-red-900/20',
    borderColor: 'border-red-200 dark:border-red-800',
    dotColor: 'bg-red-500',
    icon: 'x-circle',
    pulse: false,
  },
  offline: {
    color: 'text-gray-600 dark:text-gray-400',
    bgColor: 'bg-gray-100 dark:bg-gray-900/20',
    borderColor: 'border-gray-200 dark:border-gray-800',
    dotColor: 'bg-gray-500',
    icon: 'wifi-off',
    pulse: false,
  },
  connecting: {
    color: 'text-blue-600 dark:text-blue-400',
    bgColor: 'bg-blue-100 dark:bg-blue-900/20',
    borderColor: 'border-blue-200 dark:border-blue-800',
    dotColor: 'bg-blue-500',
    icon: 'loading',
    pulse: true,
  },
  unknown: {
    color: 'text-gray-600 dark:text-gray-400',
    bgColor: 'bg-gray-100 dark:bg-gray-900/20',
    borderColor: 'border-gray-200 dark:border-gray-800',
    dotColor: 'bg-gray-400',
    icon: 'question-mark-circle',
    pulse: false,
  },
};

export const StatusIndicator: React.FC<StatusIndicatorProps> = ({
  status,
  label,
  description,
  showLabel = true,
  size = 'md',
  className = '',
  onClick,
  lastUpdated,
}) => {
  const config = statusConfig[status];

  const sizeClasses = {
    sm: {
      dot: 'w-2 h-2',
      text: 'text-xs',
      padding: 'px-2 py-1',
      icon: 'w-3 h-3',
    },
    md: {
      dot: 'w-3 h-3',
      text: 'text-sm',
      padding: 'px-3 py-1.5',
      icon: 'w-4 h-4',
    },
    lg: {
      dot: 'w-4 h-4',
      text: 'text-base',
      padding: 'px-4 py-2',
      icon: 'w-5 h-5',
    },
  };

  const sizeClass = sizeClasses[size];

  const baseClasses = `
    inline-flex items-center space-x-2 rounded-full border
    ${config.bgColor} ${config.borderColor}
    ${sizeClass.padding}
    ${onClick ? 'cursor-pointer hover:opacity-80 transition-opacity' : ''}
    ${className}
  `;

  const content = (
    <>
      <div className="flex items-center space-x-2">
        <div
          className={`
            ${sizeClass.dot} ${config.dotColor} rounded-full
            ${config.pulse ? 'animate-pulse' : ''}
          `}
        />
        {showLabel && (
          <span className={`font-medium ${config.color} ${sizeClass.text}`}>
            {label}
          </span>
        )}
      </div>

      {description && (
        <span className={`${config.color} ${sizeClass.text} opacity-75`}>
          {description}
        </span>
      )}

      {lastUpdated && (
        <span className="text-gray-500 dark:text-gray-400 text-xs">
          {lastUpdated}
        </span>
      )}
    </>
  );

  if (onClick) {
    return (
      <button
        className={baseClasses}
        onClick={onClick}
        type="button"
        aria-label={`${label} status: ${status}${description ? `. ${description}` : ''}`}
      >
        {content}
      </button>
    );
  }

  return (
    <div
      className={baseClasses}
      role="status"
      aria-label={`${label} status: ${status}${description ? `. ${description}` : ''}`}
    >
      {content}
    </div>
  );
};

export default StatusIndicator;
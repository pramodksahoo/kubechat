interface LoadingIndicatorProps {
  type?: 'query' | 'command' | 'general';
  message?: string;
  className?: string;
}

export function LoadingIndicator({
  type = 'general',
  message,
  className = ''
}: LoadingIndicatorProps) {
  const getLoadingConfig = () => {
    switch (type) {
      case 'query':
        return {
          icon: '🧠',
          defaultMessage: 'Processing your query...',
          color: 'text-blue-600',
          bgColor: 'bg-blue-50',
          borderColor: 'border-blue-200'
        };
      case 'command':
        return {
          icon: '⚡',
          defaultMessage: 'Executing command...',
          color: 'text-green-600',
          bgColor: 'bg-green-50',
          borderColor: 'border-green-200'
        };
      default:
        return {
          icon: '⏳',
          defaultMessage: 'Loading...',
          color: 'text-gray-600',
          bgColor: 'bg-gray-50',
          borderColor: 'border-gray-200'
        };
    }
  };

  const config = getLoadingConfig();

  return (
    <div className={`${config.bgColor} ${config.borderColor} border rounded-lg p-3 ${className}`}>
      <div className="flex items-center gap-3">
        {/* Animated Icon */}
        <div className="text-lg animate-bounce">
          {config.icon}
        </div>

        {/* Loading Dots Animation */}
        <div className="flex items-center gap-1">
          <div className={`w-2 h-2 ${config.color.replace('text-', 'bg-')} rounded-full animate-bounce`}></div>
          <div
            className={`w-2 h-2 ${config.color.replace('text-', 'bg-')} rounded-full animate-bounce`}
            style={{ animationDelay: '0.1s' }}
          ></div>
          <div
            className={`w-2 h-2 ${config.color.replace('text-', 'bg-')} rounded-full animate-bounce`}
            style={{ animationDelay: '0.2s' }}
          ></div>
        </div>

        {/* Loading Message */}
        <span className={`text-sm font-medium ${config.color}`}>
          {message || config.defaultMessage}
        </span>
      </div>

      {/* Progress Bar for Commands */}
      {type === 'command' && (
        <div className="mt-3">
          <div className="w-full bg-green-200 rounded-full h-1">
            <div className="bg-green-600 h-1 rounded-full animate-pulse" style={{ width: '60%' }}></div>
          </div>
        </div>
      )}
    </div>
  );
}

// Specific loading components for different use cases
export function QueryProcessingLoader({ message, className }: { message?: string; className?: string }) {
  return (
    <LoadingIndicator
      type="query"
      message={message || 'Analyzing your request with AI...'}
      className={className}
    />
  );
}

export function CommandExecutionLoader({ message, className }: { message?: string; className?: string }) {
  return (
    <LoadingIndicator
      type="command"
      message={message || 'Executing command on Kubernetes cluster...'}
      className={className}
    />
  );
}

export function TypingIndicator({ className }: { className?: string }) {
  return (
    <div className={`flex items-center gap-2 p-3 bg-gray-100 dark:bg-gray-800 rounded-lg max-w-xs ${className}`}>
      <div className="text-gray-500 dark:text-gray-400 text-sm">KubeChat is typing</div>
      <div className="flex space-x-1">
        <div className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce"></div>
        <div className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
        <div className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
      </div>
    </div>
  );
}
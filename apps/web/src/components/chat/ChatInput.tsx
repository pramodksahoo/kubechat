import { useState, KeyboardEvent } from 'react';
import { Button } from '../ui/Button';

interface ChatInputProps {
  value: string;
  onChange: (value: string) => void;
  onSend: (message: string) => void;
  loading?: boolean;
  placeholder?: string;
  className?: string;
}

export function ChatInput({
  value,
  onChange,
  onSend,
  loading = false,
  placeholder = 'Type your message...',
  className = '',
}: ChatInputProps) {
  const [isComposing, setIsComposing] = useState(false);

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey && !isComposing) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleSend = () => {
    if (!value.trim() || loading) return;
    onSend(value.trim());
  };

  const getExampleCommands = () => [
    "List all pods in the default namespace",
    "Scale deployment nginx to 3 replicas",
    "Get logs from pod frontend-abc123",
    "Check cluster node status",
    "Create a deployment with image nginx:latest"
  ];

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Quick Command Suggestions */}
      {value.length === 0 && (
        <div className="flex flex-wrap gap-2">
          {getExampleCommands().slice(0, 3).map((command, index) => (
            <button
              key={index}
              onClick={() => onChange(command)}
              className="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              disabled={loading}
            >
              {command}
            </button>
          ))}
        </div>
      )}

      {/* Input Area */}
      <div className="flex space-x-3">
        <div className="flex-1">
          <textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            onKeyDown={handleKeyDown}
            onCompositionStart={() => setIsComposing(true)}
            onCompositionEnd={() => setIsComposing(false)}
            placeholder={placeholder}
            disabled={loading}
            rows={1}
            className="w-full min-h-[44px] max-h-32 resize-none rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-4 py-3 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:border-blue-500 dark:focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:focus:ring-blue-400 disabled:opacity-50 disabled:cursor-not-allowed"
            style={{
              height: 'auto',
              minHeight: '44px',
            }}
            onInput={(e) => {
              const target = e.target as HTMLTextAreaElement;
              target.style.height = 'auto';
              target.style.height = `${Math.min(target.scrollHeight, 128)}px`;
            }}
          />
        </div>

        <Button
          onClick={handleSend}
          disabled={!value.trim() || loading}
          loading={loading}
          variant="primary"
          className="self-end h-11 px-6"
        >
          {loading ? (
            <div className="flex items-center space-x-2">
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
              <span>Sending</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
              <span>Send</span>
            </div>
          )}
        </Button>
      </div>

      {/* Input Hints */}
      <div className="text-xs text-gray-500 dark:text-gray-400">
        <span>Press Enter to send, Shift+Enter for new line</span>
        {value.trim() && (
          <span className="ml-2">• Character count: {value.trim().length}</span>
        )}
      </div>
    </div>
  );
}
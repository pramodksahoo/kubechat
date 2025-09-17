import { useState } from 'react';
import { formatDistanceToNow } from 'date-fns';
import { ChatMessage } from '@/types/chat';
import { CommandExecutionDisplay } from './CommandExecutionDisplay';
import { CommandResultDisplay } from './CommandResultDisplay';
import { ErrorDisplay } from './ErrorDisplay';

interface ChatMessageBubbleProps {
  message: ChatMessage & { timestamp: Date };
}

export function ChatMessageBubble({ message }: ChatMessageBubbleProps) {
  const isUser = message.type === 'user';
  const isSystem = message.type === 'system';

  // Check if this message contains command execution data
  const hasCommandData = message.metadata?.command;
  const hasExecutionData = message.metadata?.executionId;
  const hasError = message.metadata?.error;

  const getSafetyColor = (safetyLevel?: string) => {
    switch (safetyLevel) {
      case 'dangerous': return 'text-red-600 dark:text-red-400';
      case 'warning': return 'text-yellow-600 dark:text-yellow-400';
      case 'safe': return 'text-green-600 dark:text-green-400';
      default: return '';
    }
  };

  const getSafetyIcon = (safetyLevel?: string) => {
    switch (safetyLevel) {
      case 'dangerous': return '⚠️';
      case 'warning': return '⚡';
      case 'safe': return '✅';
      default: return '';
    }
  };

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}>
      <div className={`max-w-4xl ${
        isUser
          ? 'bg-blue-600 text-white'
          : isSystem
          ? 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-900 dark:text-yellow-100 border border-yellow-200 dark:border-yellow-800'
          : 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white'
      } rounded-lg shadow-sm overflow-hidden`}>

        {/* Message Content */}
        <div className="p-3">
          {/* Safety indicator for commands */}
          {hasCommandData && message.metadata?.safetyLevel && (
            <div className={`flex items-center gap-2 mb-2 text-sm font-medium ${getSafetyColor(message.metadata.safetyLevel)}`}>
              <span>{getSafetyIcon(message.metadata.safetyLevel)}</span>
              <span className="capitalize">{message.metadata.safetyLevel} Command</span>
              {message.metadata.approvalRequired && (
                <span className="bg-red-100 text-red-800 px-2 py-1 rounded-full text-xs">
                  Approval Required
                </span>
              )}
            </div>
          )}

          {/* Main message content with markdown-like rendering */}
          <div className="break-words">
            <MessageContent content={message.content} />
          </div>

          {/* Command execution display (AC: 2, 4) */}
          {hasCommandData && message.metadata?.command && (
            <CommandExecutionDisplay
              command={message.metadata.command}
              safetyLevel={message.metadata.safetyLevel || 'safe'}
              confidence={message.metadata.confidence}
              explanation={message.metadata.explanation}
              potentialImpact={message.metadata.potentialImpact}
              executionId={message.metadata.executionId}
            />
          )}

          {/* Command results display (AC: 2, 3) */}
          {hasExecutionData && message.metadata?.executionId && (
            <CommandResultDisplay
              executionId={message.metadata.executionId}
            />
          )}

          {/* Error display with retry and suggestions (AC: 5) */}
          {hasError && message.metadata?.error && (
            <ErrorDisplay
              error={message.metadata.error}
              type={message.metadata.errorType as any}
              context="Message processing"
              onRetry={message.metadata.canRetry ? () => {
                // TODO: Implement retry functionality
                console.log('Retry requested for message:', message.id);
              } : undefined}
              className="mt-3"
            />
          )}

          {/* Timestamp */}
          <div className="mt-2 text-xs opacity-60">
            {formatDistanceToNow(message.timestamp, { addSuffix: true })}
          </div>
        </div>
      </div>
    </div>
  );
}

// Component to render message content with basic markdown support and copy functionality
function MessageContent({ content }: { content: string }) {
  const [copied, setCopied] = useState(false);

  const copyToClipboard = () => {
    navigator.clipboard?.writeText(content).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  // Split content by code blocks and regular text
  const parts = content.split(/(`[^`]+`|\*\*[^*]+\*\*)/g);

  return (
    <div className="relative group">
      {/* Copy button (Story requirement: copy functionality) */}
      <button
        onClick={copyToClipboard}
        className="absolute top-0 right-0 p-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity"
        title="Copy message"
      >
        {copied ? (
          <span className="text-green-600">✓ Copied!</span>
        ) : (
          <span>📋</span>
        )}
      </button>

      <div>
        {parts.map((part, index) => {
          if (part.startsWith('`') && part.endsWith('`')) {
            // Inline code with copy functionality
            return (
              <code
                key={index}
                className="relative group bg-gray-800 text-green-400 px-2 py-1 rounded text-sm font-mono cursor-pointer hover:bg-gray-700"
                onClick={() => navigator.clipboard?.writeText(part.slice(1, -1))}
                title="Click to copy code"
              >
                {part.slice(1, -1)}
              </code>
            );
          } else if (part.startsWith('**') && part.endsWith('**')) {
            // Bold text
            return (
              <strong key={index} className="font-semibold">
                {part.slice(2, -2)}
              </strong>
            );
          } else {
            // Regular text with line breaks
            return (
              <span key={index} className="whitespace-pre-wrap">
                {part}
              </span>
            );
          }
        })}
      </div>
    </div>
  );
}
import { ChatMessage } from '@/types/chat';
import { ChatMessageBubble } from './ChatMessageBubble';

interface ChatMessageListProps {
  messages: ChatMessage[];
  loading?: boolean;
  className?: string;
}

export function ChatMessageList({ 
  messages, 
  loading = false, 
  className = '' 
}: ChatMessageListProps) {
  return (
    <div className={`overflow-y-auto p-4 space-y-4 ${className}`}>
      {messages.length === 0 && !loading && (
        <div className="text-center py-8">
          <div className="text-gray-500 dark:text-gray-400">
            <svg 
              className="mx-auto h-12 w-12 mb-4 opacity-50" 
              fill="none" 
              viewBox="0 0 24 24" 
              stroke="currentColor"
            >
              <path 
                strokeLinecap="round" 
                strokeLinejoin="round" 
                strokeWidth={1.5} 
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" 
              />
            </svg>
            <p className="text-lg font-medium mb-2">Start a conversation</p>
            <p className="text-sm">
              Ask me anything about your Kubernetes cluster or describe what you&apos;d like to do.
            </p>
          </div>
        </div>
      )}

      {messages.map((message) => (
        <ChatMessageBubble
          key={message.id}
          message={{
            ...message,
            timestamp: new Date(message.timestamp)
          }}
        />
      ))}

      {loading && (
        <div className="flex justify-start">
          <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-3 max-w-xs">
            <div className="flex space-x-1">
              <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
              <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
              <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
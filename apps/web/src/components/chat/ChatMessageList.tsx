import { useState, useEffect, useRef } from 'react';
import { ChatMessage } from '@/types/chat';
import { ChatMessageBubble } from './ChatMessageBubble';
import { TypingIndicator } from './LoadingIndicator';

interface ChatMessageListProps {
  messages: ChatMessage[];
  loading?: boolean;
  className?: string;
  autoScroll?: boolean;
  messagesPerPage?: number;
}

export function ChatMessageList({
  messages,
  loading = false,
  className = '',
  autoScroll = true,
  messagesPerPage = 50
}: ChatMessageListProps) {
  const [displayedMessages, setDisplayedMessages] = useState<ChatMessage[]>([]);
  const [showScrollToBottom, setShowScrollToBottom] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Enhanced auto-scroll with user control (Story requirement)
  useEffect(() => {
    if (autoScroll && messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, autoScroll]);

  // Message pagination (Story requirement)
  useEffect(() => {
    setDisplayedMessages(messages.slice(-messagesPerPage));
  }, [messages, messagesPerPage]);

  // Scroll detection for scroll-to-bottom button
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 100;
      setShowScrollToBottom(!isNearBottom && messages.length > 0);
    };

    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, [messages.length]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const loadMoreMessages = () => {
    if (displayedMessages.length < messages.length) {
      const newCount = Math.min(displayedMessages.length + messagesPerPage, messages.length);
      setDisplayedMessages(messages.slice(-newCount));
    }
  };
  return (
    <div ref={containerRef} className={`relative overflow-y-auto p-4 space-y-4 ${className}`}>
      {/* Load More Messages Button (Story requirement: message pagination) */}
      {displayedMessages.length < messages.length && (
        <div className="text-center pb-4">
          <button
            onClick={loadMoreMessages}
            className="px-4 py-2 text-sm bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
          >
            Load {Math.min(messagesPerPage, messages.length - displayedMessages.length)} more messages
          </button>
        </div>
      )}

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

      {/* Display paginated messages */}
      {displayedMessages.map((message) => (
        <ChatMessageBubble
          key={message.id}
          message={{
            ...message,
            timestamp: new Date(message.timestamp)
          } as ChatMessage & { timestamp: Date }}
        />
      ))}

      {loading && (
        <div className="flex justify-start">
          <TypingIndicator />
        </div>
      )}

      {/* Auto-scroll anchor */}
      <div ref={messagesEndRef} />

      {/* Scroll to bottom button (Story requirement: auto-scroll control) */}
      {showScrollToBottom && (
        <button
          onClick={scrollToBottom}
          className="fixed bottom-20 right-6 p-3 bg-blue-600 text-white rounded-full shadow-lg hover:bg-blue-700 transition-colors z-10"
          title="Scroll to bottom"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
          </svg>
        </button>
      )}
    </div>
  );
}
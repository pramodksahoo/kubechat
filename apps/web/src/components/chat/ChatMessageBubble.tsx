import { formatDistanceToNow } from 'date-fns';

// Simple interface for this component's needs
interface Message {
  id: string;
  userId: string;
  type: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  metadata?: Record<string, unknown>;
}

interface ChatMessageBubbleProps {
  message: Message;
}

export function ChatMessageBubble({ message }: ChatMessageBubbleProps) {
  const isUser = message.type === 'user';

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}>
      <div className={`max-w-3xl ${
        isUser 
          ? 'bg-blue-600 text-white' 
          : 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white'
      } rounded-lg p-3 shadow-sm`}>
        <div className="break-words whitespace-pre-wrap">
          {String(message.content)}
        </div>
        <div className="mt-2 text-xs opacity-60">
          {formatDistanceToNow(message.timestamp, { addSuffix: true })}
        </div>
      </div>
    </div>
  );
}
import { useState, useRef, useEffect } from 'react';
import { ChatMessage, ChatSession, CommandPreview } from '@kubechat/shared/types';
import { CommandPreviewCard } from './CommandPreviewCard';
import { ChatMessageList } from './ChatMessageList';
import { ChatInput } from './ChatInput';

interface ChatInterfaceProps {
  session: ChatSession;
  messages: ChatMessage[];
  onSendMessage: (message: string) => void;
  onCommandPreview?: (preview: CommandPreview) => void;
  onCommandApprove?: (previewId: string) => void;
  onCommandReject?: (previewId: string, reason: string) => void;
  loading?: boolean;
  className?: string;
}

export function ChatInterface({
  session,
  messages,
  onSendMessage,
  onCommandPreview,
  onCommandApprove,
  onCommandReject,
  loading = false,
  className = '',
}: ChatInterfaceProps) {
  const [currentMessage, setCurrentMessage] = useState('');
  const [showCommandPreview, setShowCommandPreview] = useState(false);
  const [pendingCommand, setPendingCommand] = useState<CommandPreview | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = async (message: string) => {
    if (!message.trim() || loading) return;

    setCurrentMessage('');
    onSendMessage(message);

    // Check if message looks like a command
    if (isCommandLike(message)) {
      setShowCommandPreview(true);
      // In real implementation, this would call the API to generate command preview
      const preview = await generateCommandPreview(message);
      setPendingCommand(preview);
      onCommandPreview?.(preview);
    }
  };

  const handleCommandApprove = () => {
    if (pendingCommand) {
      onCommandApprove?.(pendingCommand.id);
      setShowCommandPreview(false);
      setPendingCommand(null);
    }
  };

  const handleCommandReject = (reason: string) => {
    if (pendingCommand) {
      onCommandReject?.(pendingCommand.id, reason);
      setShowCommandPreview(false);
      setPendingCommand(null);
    }
  };

  return (
    <div className={`flex flex-col h-full ${className}`}>
      {/* Chat Header */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
              {session.title || `Chat Session`}
            </h2>
            {session.clusterName && (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Connected to: {session.clusterName}
              </p>
            )}
          </div>
          <div className="flex items-center space-x-2">
            <div className="flex items-center space-x-1">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-xs text-gray-500 dark:text-gray-400">
                {session.status === 'active' ? 'Online' : 'Offline'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-hidden">
        <ChatMessageList 
          messages={messages} 
          loading={loading}
          className="h-full"
        />
        <div ref={messagesEndRef} />
      </div>

      {/* Command Preview */}
      {showCommandPreview && pendingCommand && (
        <div className="p-4 border-t border-gray-200 dark:border-gray-700">
          <CommandPreviewCard
            preview={pendingCommand}
            onApprove={handleCommandApprove}
            onReject={handleCommandReject}
          />
        </div>
      )}

      {/* Chat Input */}
      <div className="p-4 border-t border-gray-200 dark:border-gray-700">
        <ChatInput
          value={currentMessage}
          onChange={setCurrentMessage}
          onSend={handleSendMessage}
          loading={loading}
          placeholder="Type your message or describe what you want to do with Kubernetes..."
        />
      </div>
    </div>
  );
}

// Helper functions (would be moved to utils in real implementation)
function isCommandLike(message: string): boolean {
  const commandKeywords = [
    'kubectl', 'create', 'delete', 'deploy', 'scale', 'restart',
    'get pods', 'describe', 'logs', 'exec', 'apply', 'rollout'
  ];
  
  const lowerMessage = message.toLowerCase();
  return commandKeywords.some(keyword => lowerMessage.includes(keyword));
}

async function generateCommandPreview(message: string): Promise<CommandPreview> {
  // Mock implementation - would call actual API
  return {
    id: `preview-${Date.now()}`,
    naturalLanguage: message,
    generatedCommand: `kubectl get pods -n default`,
    safetyLevel: 'safe',
    confidence: 0.85,
    explanation: 'This command will list all pods in the default namespace',
    potentialImpact: ['Read-only operation', 'No cluster changes'],
    requiredPermissions: ['pods:list'],
    clusterId: 'cluster-1',
    approvalRequired: false,
  };
}
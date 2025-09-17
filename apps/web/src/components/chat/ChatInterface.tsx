import { useState, useRef, useEffect } from 'react';
import { CommandPreview } from '@/types/chat';
import { CommandPreviewCard } from './CommandPreviewCard';
import { ChatMessageList } from './ChatMessageList';
import { ChatInput } from './ChatInput';
import { useChat } from '@/hooks/useChat';
import { commandApprovalService } from '@/services/commandApprovalService';
import { realTimeService } from '@/services/realTimeService';

interface ChatInterfaceProps {
  sessionId?: string;
  onCommandPreview?: (preview: CommandPreview) => void;
  onCommandApprove?: (previewId: string) => void;
  onCommandReject?: (previewId: string, reason: string) => void;
  className?: string;
}

export function ChatInterface({
  sessionId,
  onCommandPreview,
  onCommandApprove,
  onCommandReject,
  className = '',
}: ChatInterfaceProps) {
  const {
    currentSession,
    messages,
    loading,
    error,
    connected,
    sendMessage,
    generateCommandPreview,
    clearError,
  } = useChat({ sessionId, autoLoadHistory: true, persistLocally: true });

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

  // Initialize WebSocket connection for real-time updates (AC: 7)
  useEffect(() => {
    if (!realTimeService.isWebSocketConnected()) {
      realTimeService.initializeWebSocket();
    }

    // Subscribe to chat-related real-time updates
    const subscriptionId = realTimeService.subscribe(['chat'], (update) => {
      console.log('Chat real-time update:', update);
      // Handle real-time chat updates if needed
    });

    return () => {
      realTimeService.unsubscribe(subscriptionId);
    };
  }, []);

  const handleSendMessage = async (message: string) => {
    if (!message.trim() || loading || !currentSession) return;

    setCurrentMessage('');

    // Send message with persistent history
    await sendMessage(message);

    // Check if message looks like a command
    if (isCommandLike(message)) {
      setShowCommandPreview(true);

      // Generate command preview
      const preview = await generateCommandPreview(message);
      if (preview) {
        setPendingCommand(preview);
        onCommandPreview?.(preview);
      }
    }
  };

  const handleCommandApprove = async () => {
    if (pendingCommand) {
      try {
        // For dangerous operations, create approval request
        if (pendingCommand.safetyLevel === 'dangerous' || pendingCommand.approvalRequired) {
          await commandApprovalService.createApprovalRequest(pendingCommand);
          // Notify user that approval request was created
          console.log('Approval request created for dangerous operation');
        } else {
          // For safe operations, execute immediately
          onCommandApprove?.(pendingCommand.id);
        }

        setShowCommandPreview(false);
        setPendingCommand(null);
      } catch (error) {
        console.error('Failed to process command approval:', error);
      }
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
              {currentSession?.title || `Chat Session`}
            </h2>
            {currentSession?.clusterName && (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Connected to: {currentSession.clusterName}
              </p>
            )}
          </div>
          <div className="flex items-center space-x-2">
            {error && (
              <div className="flex items-center space-x-1 px-2 py-1 bg-red-100 dark:bg-red-900/20 rounded-full">
                <div className="w-2 h-2 bg-red-500 rounded-full"></div>
                <span className="text-xs text-red-600 dark:text-red-400">Error</span>
                <button
                  onClick={clearError}
                  className="text-red-500 hover:text-red-700 text-xs ml-1"
                >
                  ×
                </button>
              </div>
            )}
            <div className="flex items-center space-x-1">
              <div className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`}></div>
              <span className="text-xs text-gray-500 dark:text-gray-400">
                {connected ? 'Online' : 'Offline'}
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
  const lowerMessage = message.toLowerCase();

  // Determine safety level and command based on message content
  let safetyLevel: 'safe' | 'warning' | 'dangerous' = 'safe';
  let generatedCommand = 'kubectl get pods -n default';
  let explanation = 'This command will list all pods in the default namespace';
  let potentialImpact = ['Read-only operation', 'No cluster changes'];
  let requiredPermissions = ['pods:list'];
  let approvalRequired = false;

  if (lowerMessage.includes('delete') || lowerMessage.includes('remove')) {
    if (lowerMessage.includes('namespace') || lowerMessage.includes('all') || lowerMessage.includes('cluster')) {
      safetyLevel = 'dangerous';
      generatedCommand = 'kubectl delete namespace production';
      explanation = 'This command will permanently delete the production namespace and all resources within it';
      potentialImpact = [
        'Complete loss of production namespace',
        'All pods, services, and data in namespace will be deleted',
        'Application downtime',
        'Potential data loss',
        'May require full redeployment'
      ];
      requiredPermissions = ['namespaces:delete', 'admin:cluster'];
      approvalRequired = true;
    } else {
      safetyLevel = 'warning';
      generatedCommand = 'kubectl delete pod nginx-deployment-xyz123 -n default';
      explanation = 'This command will delete a specific pod in the default namespace';
      potentialImpact = ['Pod will be terminated', 'Service interruption possible', 'Pod will be recreated by deployment'];
      requiredPermissions = ['pods:delete'];
      approvalRequired = true;
    }
  } else if (lowerMessage.includes('scale') && (lowerMessage.includes('0') || lowerMessage.includes('zero'))) {
    safetyLevel = 'warning';
    generatedCommand = 'kubectl scale deployment nginx-deployment --replicas=0 -n default';
    explanation = 'This command will scale down the deployment to zero replicas, stopping all pods';
    potentialImpact = ['All application instances will stop', 'Service will become unavailable', 'Users will experience downtime'];
    requiredPermissions = ['deployments:update'];
    approvalRequired = true;
  } else if (lowerMessage.includes('restart') || lowerMessage.includes('rollout')) {
    safetyLevel = 'warning';
    generatedCommand = 'kubectl rollout restart deployment/nginx-deployment -n default';
    explanation = 'This command will restart all pods in the deployment with rolling update';
    potentialImpact = ['Pods will be restarted one by one', 'Brief service interruption during rolling restart'];
    requiredPermissions = ['deployments:update'];
    approvalRequired = false;
  }

  return {
    id: `preview-${Date.now()}`,
    naturalLanguage: message,
    generatedCommand,
    safetyLevel,
    confidence: 0.85,
    explanation,
    potentialImpact,
    requiredPermissions,
    clusterId: 'cluster-1',
    approvalRequired,
  };
}
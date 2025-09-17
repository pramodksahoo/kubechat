export interface ChatMessage {
  id: string;
  sessionId: string;
  userId: string;
  type: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  metadata?: {
    queryId?: string;
    command?: string;
    safetyLevel?: 'safe' | 'warning' | 'dangerous';
    confidence?: number;
    explanation?: string;
    potentialImpact?: string[];
    requiredPermissions?: string[];
    executionId?: string;
    approvalRequired?: boolean;
    error?: string;
    errorType?: string;
    canRetry?: boolean;
    suggestions?: string[];
  };
}

export interface ChatSession {
  id: string;
  userId: string;
  title?: string;
  clusterId?: string;
  clusterName?: string;
  createdAt: string;
  updatedAt: string;
  status: 'active' | 'completed' | 'archived';
  messageCount: number;
  lastMessage?: ChatMessage;
}

export interface CommandPreview {
  id: string;
  naturalLanguage: string;
  generatedCommand: string;
  safetyLevel: 'safe' | 'warning' | 'dangerous';
  confidence: number;
  explanation: string;
  potentialImpact: string[];
  requiredPermissions: string[];
  clusterId: string;
  approvalRequired: boolean;
}

export interface CommandExecution {
  id: string;
  sessionId: string;
  previewId: string;
  command: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  output?: string;
  error?: string;
  result?: 'success' | 'failure';
  startedAt: Date;
  completedAt?: Date;
  executedBy: string;
  approvedBy?: string;
}
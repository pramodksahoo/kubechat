export interface ChatMessage {
  id: string;
  sessionId: string;
  type: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  metadata?: {
    command?: string;
    safetyLevel?: 'safe' | 'warning' | 'dangerous';
    executionId?: string;
    approvalRequired?: boolean;
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
  previewId: string;
  command: string;
  status: 'pending' | 'approved' | 'executing' | 'completed' | 'failed' | 'cancelled';
  result?: string;
  error?: string;
  executedAt?: string;
  completedAt?: string;
  executedBy: string;
  approvedBy?: string;
  clusterId: string;
}

export interface ApprovalRequest {
  id: string;
  commandPreview: CommandPreview;
  requestedBy: string;
  requestedAt: string;
  status: 'pending' | 'approved' | 'rejected' | 'expired';
  approvedBy?: string;
  approvedAt?: string;
  rejectedBy?: string;
  rejectedAt?: string;
  rejectionReason?: string;
  expiresAt: string;
}
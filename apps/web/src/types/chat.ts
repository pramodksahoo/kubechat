export interface ChatMessage {
  id: string;
  sessionId: string;
  userId?: string;
  type: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  metadata?: Record<string, unknown>;
}

export interface ChatSession {
  id: string;
  title: string;
  clusterId?: string;
  clusterName?: string;
  createdAt: Date;
  updatedAt: Date;
  messageCount: number;
  lastMessage?: string;
}

export interface CommandPreview {
  id: string;
  sessionId: string;
  command: string;
  description: string;
  risks: string[];
  safeguards: string[];
  estimatedImpact: 'low' | 'medium' | 'high';
  requiresApproval: boolean;
  approvalRequired: boolean;
  generatedCommand: string;
  safetyLevel: 'low' | 'medium' | 'high';
  createdAt: Date;
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
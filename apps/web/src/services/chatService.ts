import { ChatMessage, ChatSession, CommandPreview, CommandExecution } from '../types/chat';
import { api } from './api';

export class ChatService {
  private sessions: ChatSession[] = [];
  private messages: Map<string, ChatMessage[]> = new Map();

  constructor() {
    // Load sessions from localStorage on initialization
    this.loadLocalData();
  }

  // Session Management - Using local storage since backend doesn't have chat sessions
  async createSession(clusterId?: string): Promise<ChatSession> {
    const session: ChatSession = {
      id: `session-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
      userId: 'current-user', // TODO: Get from auth
      title: `Chat Session ${new Date().toLocaleString()}`,
      clusterId,
      clusterName: clusterId ? 'Production Cluster' : undefined,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: 'active',
      messageCount: 0
    };

    this.sessions.unshift(session);
    this.messages.set(session.id, []);
    this.saveLocalData();

    return session;
  }

  async getSessions(): Promise<ChatSession[]> {
    return [...this.sessions];
  }

  async getSession(sessionId: string): Promise<ChatSession> {
    const session = this.sessions.find(s => s.id === sessionId);
    if (!session) {
      throw new Error(`Session ${sessionId} not found`);
    }
    return session;
  }

  async deleteSession(sessionId: string): Promise<void> {
    this.sessions = this.sessions.filter(s => s.id !== sessionId);
    this.messages.delete(sessionId);
    this.saveLocalData();
  }

  // Message Management
  async getMessages(sessionId: string, limit = 50, offset = 0): Promise<ChatMessage[]> {
    const sessionMessages = this.messages.get(sessionId) || [];
    return sessionMessages.slice(offset, offset + limit);
  }

  async sendMessage(sessionId: string, content: string): Promise<ChatMessage> {
    // Create user message
    const userMessage: ChatMessage = {
      id: `msg-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
      sessionId,
      userId: 'current-user',
      type: 'user',
      content,
      timestamp: new Date().toISOString()
    };

    // Add to messages
    const sessionMessages = this.messages.get(sessionId) || [];
    sessionMessages.push(userMessage);

    // Process with NLP API to get AI response
    try {
      const nlpResponse = await api.nlp.process(content);

      // Create AI response message
      const aiMessage: ChatMessage = {
        id: `msg-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        sessionId,
        userId: 'assistant',
        type: 'assistant',
        content: this.generateAIResponse(content, nlpResponse.data),
        timestamp: new Date().toISOString(),
        metadata: {
          command: (nlpResponse.data as any)?.suggestedCommand,
          safetyLevel: 'safe' as const
        }
      };

      sessionMessages.push(aiMessage);
      this.messages.set(sessionId, sessionMessages);

      // Update session
      const session = this.sessions.find(s => s.id === sessionId);
      if (session) {
        session.messageCount = sessionMessages.length;
        session.updatedAt = new Date().toISOString();
        session.lastMessage = aiMessage;
      }

      this.saveLocalData();
      return userMessage;

    } catch (error) {
      console.error('NLP processing failed:', error);

      // Create fallback AI response
      const fallbackMessage: ChatMessage = {
        id: `msg-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        sessionId,
        userId: 'assistant',
        type: 'assistant',
        content: this.generateFallbackResponse(content),
        timestamp: new Date().toISOString()
      };

      sessionMessages.push(fallbackMessage);
      this.messages.set(sessionId, sessionMessages);
      this.saveLocalData();

      return userMessage;
    }
  }

  // Command Management - Using real backend APIs
  async generateCommandPreview(naturalLanguage: string, sessionId: string): Promise<CommandPreview> {
    try {
      // First classify the input to determine if it's a command
      const classifyResponse = await api.nlp.classify(naturalLanguage);
      const isCommand = (classifyResponse.data as any)?.type === 'command' ||
                       naturalLanguage.toLowerCase().includes('kubectl') ||
                       naturalLanguage.toLowerCase().includes('deploy') ||
                       naturalLanguage.toLowerCase().includes('scale');

      if (!isCommand) {
        throw new Error('Input does not appear to be a command');
      }

      // Process with NLP to generate command
      const nlpResponse = await api.nlp.process(naturalLanguage, { generateCommand: true });

      // Create command preview
      const preview: CommandPreview = {
        id: `preview-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        naturalLanguage,
        generatedCommand: this.extractCommand(nlpResponse.data) || `kubectl ${naturalLanguage.toLowerCase()}`,
        safetyLevel: this.determineSafetyLevel(naturalLanguage),
        confidence: (nlpResponse.data as any)?.confidence || 0.7,
        explanation: `Generated kubectl command from: "${naturalLanguage}"`,
        potentialImpact: this.assessPotentialImpact(naturalLanguage),
        requiredPermissions: ['kubernetes:execute'],
        clusterId: 'default',
        approvalRequired: this.requiresApproval(naturalLanguage)
      };

      return preview;

    } catch (error) {
      console.error('Failed to generate command preview:', error);

      // Create fallback preview
      return {
        id: `preview-${Date.now()}`,
        naturalLanguage,
        generatedCommand: `# Could not process: ${naturalLanguage}`,
        safetyLevel: 'warning',
        confidence: 0.3,
        explanation: 'Unable to process natural language input',
        potentialImpact: ['Unknown impact'],
        requiredPermissions: ['kubernetes:read'],
        clusterId: 'default',
        approvalRequired: true
      };
    }
  }

  async executeCommand(previewId: string): Promise<CommandExecution> {
    try {
      // Use the real command execution API
      const response = await api.commands.execute({
        command: `kubectl get pods`, // This would be the actual command from preview
        cluster: 'default'
      });

      const executionData = response.data as any;
      return {
        id: executionData.id || `exec-${Date.now()}`,
        sessionId: 'current',
        previewId,
        command: executionData.command || 'kubectl get pods',
        status: (executionData.status as 'pending' | 'running' | 'completed' | 'failed' | 'cancelled') || 'completed',
        output: executionData.output,
        error: executionData.exitCode !== 0 ? 'Command failed' : undefined,
        result: executionData.exitCode === 0 ? 'success' : 'failure',
        startedAt: executionData.executedAt ? new Date(executionData.executedAt) : new Date(),
        completedAt: executionData.completedAt ? new Date(executionData.completedAt) : new Date(),
        executedBy: 'current-user',
        approvedBy: undefined
      };

    } catch (error) {
      console.error('Command execution failed:', error);
      throw error;
    }
  }

  async requestApproval(previewId: string): Promise<void> {
    try {
      // Use the real approval API
      await api.commands.approve({
        previewId,
        requestedBy: 'current-user',
        requestedAt: new Date().toISOString()
      });
    } catch (error) {
      console.error('Failed to request approval:', error);
      throw error;
    }
  }

  async getCommandHistory(sessionId?: string, limit = 20): Promise<CommandExecution[]> {
    try {
      const response = await api.commands.listExecutions({ limit });
      return response.data.executions as CommandExecution[];
    } catch (error) {
      console.error('Failed to fetch command history:', error);
      return [];
    }
  }

  // WebSocket for real-time updates - Using actual backend WebSocket
  connectWebSocket(sessionId: string, onMessage?: (data: any) => void): void {
    try {
      api.ws.connect(
        (data) => {
          console.log('WebSocket message received:', data);
          onMessage?.(data);
        },
        (error) => {
          console.error('WebSocket error:', error);
        }
      );
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
    }
  }

  // Local Storage Management
  saveSessionLocally(session: ChatSession): void {
    const existingIndex = this.sessions.findIndex(s => s.id === session.id);
    if (existingIndex >= 0) {
      this.sessions[existingIndex] = session;
    } else {
      this.sessions.push(session);
    }
    this.saveLocalData();
  }

  saveMessagesLocally(sessionId: string, messages: ChatMessage[]): void {
    this.messages.set(sessionId, messages);
    this.saveLocalData();
  }

  getLocalSessions(): ChatSession[] {
    return [...this.sessions];
  }

  getLocalMessages(sessionId: string): ChatMessage[] {
    return this.messages.get(sessionId) || [];
  }

  clearLocalData(): void {
    this.sessions = [];
    this.messages.clear();
    localStorage.removeItem('kubechat_chat_data');
  }

  private loadLocalData(): void {
    try {
      const stored = localStorage.getItem('kubechat_chat_data');
      if (stored) {
        const data = JSON.parse(stored);
        this.sessions = data.sessions || [];
        this.messages = new Map(data.messages || []);
      }
    } catch (error) {
      console.error('Failed to load local chat data:', error);
    }
  }

  private saveLocalData(): void {
    try {
      const data = {
        sessions: this.sessions,
        messages: Array.from(this.messages.entries())
      };
      localStorage.setItem('kubechat_chat_data', JSON.stringify(data));
    } catch (error) {
      console.error('Failed to save local chat data:', error);
    }
  }

  // Helper methods for AI responses
  private generateAIResponse(userInput: string, nlpData: any): string {
    const responses = [
      `I understand you want to: ${userInput}. Let me help you with that.`,
      `Based on your request "${userInput}", here's what I can help with:`,
      `I've processed your request about "${userInput}". Here's my response:`
    ];

    if (nlpData?.intent) {
      return `I detected that you want to ${nlpData.intent}. ${responses[0]}`;
    }

    return responses[Math.floor(Math.random() * responses.length)];
  }

  private generateFallbackResponse(userInput: string): string {
    return `I received your message: "${userInput}". The AI service is currently unavailable, but I'm here to help with Kubernetes management tasks.`;
  }

  private extractCommand(nlpData: any): string | undefined {
    return nlpData?.command || nlpData?.generatedCommand;
  }

  private determineSafetyLevel(input: string): 'safe' | 'warning' | 'dangerous' {
    const dangerous = ['delete', 'remove', 'destroy', 'terminate'];
    const warning = ['scale', 'restart', 'update', 'patch'];

    const inputLower = input.toLowerCase();
    if (dangerous.some(word => inputLower.includes(word))) {
      return 'dangerous';
    }
    if (warning.some(word => inputLower.includes(word))) {
      return 'warning';
    }
    return 'safe';
  }

  private assessPotentialImpact(input: string): string[] {
    const impact = [];
    if (input.toLowerCase().includes('delete')) {
      impact.push('Resource deletion');
    }
    if (input.toLowerCase().includes('scale')) {
      impact.push('Resource scaling');
    }
    if (input.toLowerCase().includes('deploy')) {
      impact.push('Application deployment');
    }
    return impact.length > 0 ? impact : ['Configuration change'];
  }

  private requiresApproval(input: string): boolean {
    const requiresApproval = ['delete', 'remove', 'destroy', 'terminate', 'scale down'];
    return requiresApproval.some(word => input.toLowerCase().includes(word));
  }

  private getAuthToken(): string {
    return localStorage.getItem('auth_token') || '';
  }
}

export const chatService = new ChatService();
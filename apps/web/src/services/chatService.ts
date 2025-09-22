import { ChatMessage, ChatSession, CommandPreview, CommandExecution } from '../types/chat';
import { api } from './api';
import { errorHandlingService } from './errorHandlingService';

export class ChatService {
  private sessions: ChatSession[] = [];
  private messages: Map<string, ChatMessage[]> = new Map();

  constructor() {
    // Load sessions from localStorage on initialization
    this.loadLocalData();
  }

  // Session Management - Using backend chat API
  async createSession(clusterId?: string): Promise<ChatSession> {
    try {
      // Call the backend API to create a session
      const response = await api.chat.createSession({
        clusterId: clusterId
      });

      const backendSession = response.data.data;
      
      // Convert backend response to our ChatSession format
      const session: ChatSession = {
        id: backendSession.id,
        userId: 'current-user', // Will be set by backend based on auth
        title: backendSession.title || `Chat Session ${new Date().toLocaleString()}`,
        clusterId: clusterId,
        clusterName: clusterId ? 'Production Cluster' : undefined,
        createdAt: backendSession.createdAt,
        updatedAt: backendSession.updatedAt,
        status: 'active',
        messageCount: backendSession.messageCount || 0
      };

      // Store locally for offline access
      this.sessions.unshift(session);
      this.messages.set(session.id, []);
      this.saveLocalData();

      return session;
    } catch (error) {
      console.error('Backend session creation failed, falling back to local:', error);
      
      // Fallback to local session creation if backend fails
      const session: ChatSession = {
        id: `local-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        userId: 'current-user',
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
    try {
      // Step 1: Send user message to backend chat API
      const userMessageResponse = await api.chat.sendMessage(sessionId, {
        content,
        type: 'user'
      });

      // Create properly formatted user message from backend response
      const userMessage: ChatMessage = {
        id: userMessageResponse.data.data.id,
        sessionId: userMessageResponse.data.data.sessionId,
        userId: 'current-user',
        type: 'user',
        content: userMessageResponse.data.data.content,
        timestamp: userMessageResponse.data.data.timestamp
      };

      // Step 2: Generate AI response locally since backend doesn't do it automatically
      let aiMessage: ChatMessage;

      try {
        // Try to use NLP service (without auth - will fail but gracefully)
        const nlpResponse = await api.nlp.processQuery({
          query: content,
          context: 'default',
        });

        const queryData = nlpResponse.data;
        const aiContent = this.generateAIResponseFromQuery(content, queryData);

        // Send AI response back to backend
        const aiMessageResponse = await api.chat.sendMessage(sessionId, {
          content: aiContent,
          type: 'assistant'
        });

        aiMessage = {
          id: aiMessageResponse.data.data.id,
          sessionId: aiMessageResponse.data.data.sessionId,
          userId: 'assistant',
          type: 'assistant',
          content: aiContent,
          timestamp: aiMessageResponse.data.data.timestamp,
          metadata: {
            queryId: queryData.id,
            command: queryData.generated_command,
            safetyLevel: queryData.safety_level,
            confidence: queryData.confidence,
            explanation: queryData.explanation,
            potentialImpact: queryData.potential_impact || [],
            requiredPermissions: queryData.required_permissions || [],
            approvalRequired: queryData.approval_required || false
          }
        };
      } catch (nlpError) {
        // NLP failed - create simple AI response without advanced processing
        const simpleResponse = this.generateSimpleAIResponse(content);
        const aiMessageResponse = await api.chat.sendMessage(sessionId, {
          content: simpleResponse,
          type: 'assistant'
        });

        aiMessage = {
          id: aiMessageResponse.data.data.id,
          sessionId: aiMessageResponse.data.data.sessionId,
          userId: 'assistant',
          type: 'assistant',
          content: simpleResponse,
          timestamp: aiMessageResponse.data.data.timestamp,
          metadata: {
            error: 'NLP service unavailable - using fallback response',
            safetyLevel: 'warning'
          }
        };
      }

      // Update local state
      const sessionMessages = this.messages.get(sessionId) || [];
      sessionMessages.push(userMessage, aiMessage);
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
      // Enhanced error handling with retry mechanisms (AC: 5)
      const errorDetails = await errorHandlingService.handleError(error as Error, {
        context: {
          operation: 'query-processing',
          component: 'ChatService',
          sessionId: sessionId
        },
        logToConsole: true
      });

      // Create fallback AI response with error context
      const fallbackMessage: ChatMessage = {
        id: `msg-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        sessionId,
        userId: 'assistant',
        type: 'assistant',
        content: this.generateFallbackResponse(content, errorDetails),
        timestamp: new Date().toISOString(),
        metadata: {
          error: error instanceof Error ? error.message : String(error),
          errorType: errorDetails.type,
          canRetry: errorDetails.canRetry,
          suggestions: errorDetails.suggestions
        }
      };

      // Get or create session messages array
      const sessionMessages = this.messages.get(sessionId) || [];
      sessionMessages.push(fallbackMessage);
      this.messages.set(sessionId, sessionMessages);
      this.saveLocalData();

      // Create a basic user message since the main flow failed
      const basicUserMessage: ChatMessage = {
        id: `msg-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        sessionId,
        userId: 'current-user',
        type: 'user',
        content,
        timestamp: new Date().toISOString()
      };

      return basicUserMessage;
    }
  }

  // Command Management - Using real backend APIs
  async generateCommandPreview(naturalLanguage: string, _sessionId: string): Promise<CommandPreview> {
    try {
      // Use the /nlp API to process natural language and generate command preview (AC: 1)
      const nlpResponse = await api.nlp.processQuery({
        query: naturalLanguage,
        context: 'default',
      });

      const queryData = nlpResponse.data;

      // Create command preview from query response
      const preview: CommandPreview = {
        id: `preview-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        naturalLanguage,
        generatedCommand: queryData.generated_command || `kubectl get pods`,
        safetyLevel: queryData.safety_level,
        confidence: queryData.confidence,
        explanation: queryData.explanation || `Generated kubectl command from: "${naturalLanguage}"`,
        potentialImpact: queryData.potential_impact || [],
        requiredPermissions: queryData.required_permissions || [],
        clusterId: 'default',
        approvalRequired: queryData.approval_required || false
      };

      return preview;

    } catch (error) {
      // Enhanced error handling for command preview generation (AC: 5)
      const errorDetails = await errorHandlingService.handleError(error as Error, {
        context: {
          operation: 'command-preview-generation',
          component: 'ChatService'
        },
        logToConsole: true
      });

      // Create fallback preview with error context
      return {
        id: `preview-${Date.now()}`,
        naturalLanguage,
        generatedCommand: `# Error: Could not process "${naturalLanguage}"`,
        safetyLevel: 'warning',
        confidence: 0.1,
        explanation: `${errorDetails.type} error occurred while processing your request. ${errorDetails.suggestions.join(' ')}`,
        potentialImpact: ['Unknown impact - query processing failed', 'Manual review required'],
        requiredPermissions: ['kubernetes:read'],
        clusterId: 'default',
        approvalRequired: true
      };
    }
  }

  async executeCommand(previewId: string, command?: string): Promise<CommandExecution> {
    try {
      // Use the /commands/execute API for actual command execution (AC: 2)
      const response = await api.commands.execute({
        command: command || 'kubectl get pods', // Use provided command or fallback
        cluster: 'default'
      });

      const executionData = response.data;
      return {
        id: executionData.id || `exec-${Date.now()}`,
        sessionId: 'current',
        previewId,
        command: executionData.command,
        status: executionData.status as 'pending' | 'running' | 'completed' | 'failed' | 'cancelled',
        output: executionData.output,
        error: executionData.exitCode !== 0 ? 'Command failed' : undefined,
        result: executionData.exitCode === 0 ? 'success' : 'failure',
        startedAt: executionData.executedAt ? new Date(executionData.executedAt) : new Date(),
        completedAt: (executionData as any).completedAt ? new Date((executionData as any).completedAt) : undefined,
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
      // Use the /commands/executions API for command history (AC: 3)
      const response = await api.commands.listExecutions({ limit, page: 1 });

      // Transform the response to match our CommandExecution interface
      const executions = response.data.executions.map((exec: any) => ({
        id: exec.id,
        sessionId: sessionId || 'unknown',
        previewId: exec.previewId || `preview-${exec.id}`,
        command: exec.command,
        status: exec.status as 'pending' | 'running' | 'completed' | 'failed' | 'cancelled',
        output: exec.output,
        error: exec.error,
        result: exec.exitCode === 0 ? 'success' : 'failure',
        startedAt: exec.executedAt ? new Date(exec.executedAt) : new Date(),
        completedAt: exec.completedAt ? new Date(exec.completedAt) : undefined,
        executedBy: exec.executedBy || 'unknown-user',
        approvedBy: exec.approvedBy,
        executionTimeMs: exec.executionTimeMs
      })) as CommandExecution[];

      return executions;
    } catch (error) {
      console.error('Failed to fetch command history:', error);
      return [];
    }
  }

  // WebSocket for real-time updates - Using actual backend WebSocket
  connectWebSocket(_sessionId: string, onMessage?: (data: any) => void): void {
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

    // Check if we're in a browser environment
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.removeItem('kubechat_chat_data');
    }
  }

  private loadLocalData(): void {
    try {
      // Check if we're in a browser environment
      if (typeof window === 'undefined' || !window.localStorage) {
        return; // Skip localStorage operations during SSR
      }

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
      // Check if we're in a browser environment
      if (typeof window === 'undefined' || !window.localStorage) {
        return; // Skip localStorage operations during SSR
      }

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
  private generateAIResponseFromQuery(userInput: string, queryData: any): string {
    if (queryData.generatedCommand) {
      const safetyIndicator = queryData.safetyLevel === 'dangerous' ? '⚠️ ' :
                              queryData.safetyLevel === 'warning' ? '⚡ ' : '✅ ';

      let response = `${safetyIndicator}I've analyzed your request: "${userInput}"\\n\\n`;
      response += `**Generated Command:** \`${queryData.generatedCommand}\`\\n\\n`;

      if (queryData.explanation) {
        response += `**Explanation:** ${queryData.explanation}\\n\\n`;
      }

      if (queryData.confidence) {
        response += `**Confidence:** ${Math.round(queryData.confidence * 100)}%\\n\\n`;
      }

      if (queryData.potentialImpact?.length > 0) {
        response += `**Potential Impact:**\n${queryData.potentialImpact.map((impact: string) => `• ${impact}`).join('\n')}\n\n`;
      }

      if (queryData.approvalRequired) {
        response += `**⚠️ This command requires approval before execution.**`;
      } else {
        response += `**✅ This command is safe to execute.**`;
      }

      return response;
    }

    return `I've processed your request: "${userInput}". How can I help you with Kubernetes management?`;
  }

  private generateSimpleAIResponse(userInput: string): string {
    // Simple fallback response when NLP service is unavailable
    const responses = [
      `I received your message: "${userInput}". I'm here to help with Kubernetes management.`,
      `Thank you for your question about "${userInput}". Let me help you with that.`,
      `I understand you're asking about "${userInput}". How can I assist you with Kubernetes?`,
      `I've noted your request: "${userInput}". What specific Kubernetes task would you like help with?`
    ];

    // Simple hash to pick consistent response for same input
    const hash = userInput.split('').reduce((a, b) => {
      a = ((a << 5) - a) + b.charCodeAt(0);
      return a & a;
    }, 0);

    return responses[Math.abs(hash) % responses.length];
  }

  private generateFallbackResponse(userInput: string, errorDetails?: any): string {
    if (errorDetails) {
      let response = `I encountered an issue while processing your message: "${userInput}"\\n\\n`;
      response += `**Error:** ${errorDetails.type} error occurred.\\n\\n`;

      if (errorDetails.suggestions && errorDetails.suggestions.length > 0) {
        response += `**Suggestions:**\\n`;
        errorDetails.suggestions.forEach((suggestion: string) => {
          response += `• ${suggestion}\\n`;
        });
        response += '\\n';
      }

      if (errorDetails.canRetry) {
        response += `You can try your request again, or rephrase it if the issue persists.`;
      } else {
        response += `Please check the suggestions above and modify your request.`;
      }

      return response;
    }

    return `I received your message: "${userInput}". The API service is currently unavailable, but I'm here to help with Kubernetes management tasks.`;
  }
}

export const chatService = new ChatService();
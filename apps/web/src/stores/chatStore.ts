import { create } from 'zustand';
import { ChatMessage, ChatSession, CommandPreview } from '../types/chat';
import { chatService } from '../services/chatService';

interface ChatState {
  // Sessions
  sessions: ChatSession[];
  currentSession: ChatSession | null;
  
  // Messages
  messages: ChatMessage[];
  loading: boolean;
  
  // Command preview
  currentPreview: CommandPreview | null;
  showPreview: boolean;
  
  // WebSocket
  wsConnection: WebSocket | null;
  wsConnected: boolean;
  
  // Actions
  loadSessions: () => Promise<void>;
  createSession: (clusterId?: string) => Promise<ChatSession>;
  selectSession: (sessionId: string) => Promise<void>;
  deleteSession: (sessionId: string) => Promise<void>;
  
  sendMessage: (content: string) => Promise<void>;
  loadMessages: (sessionId: string) => Promise<void>;
  
  generatePreview: (naturalLanguage: string) => Promise<void>;
  approveCommand: () => Promise<void>;
  rejectCommand: (reason?: string) => void;
  
  connectWebSocket: (sessionId: string) => void;
  disconnectWebSocket: () => void;
  
  // Utility
  clearAll: () => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  // Initial state
  sessions: [],
  currentSession: null,
  messages: [],
  loading: false,
  currentPreview: null,
  showPreview: false,
  wsConnection: null,
  wsConnected: false,

  // Session actions
  loadSessions: async () => {
    try {
      const sessions = await chatService.getSessions();
      set({ sessions });
    } catch (error) {
      console.error('Failed to load sessions:', error);
      // Fallback to local sessions
      const localSessions = chatService.getLocalSessions();
      set({ sessions: localSessions });
    }
  },

  createSession: async (clusterId?: string) => {
    try {
      const session = await chatService.createSession(clusterId);
      set(state => ({
        sessions: [session, ...state.sessions],
        currentSession: session,
        messages: [],
        currentPreview: null,
        showPreview: false,
      }));
      
      // Save locally for offline access
      chatService.saveSessionLocally(session);
      
      return session;
    } catch (error) {
      console.error('Failed to create session:', error);
      throw error;
    }
  },

  selectSession: async (sessionId: string) => {
    const { sessions } = get();
    const session = sessions.find(s => s.id === sessionId);
    
    if (!session) {
      throw new Error('Session not found');
    }

    set({ 
      currentSession: session, 
      loading: true,
      currentPreview: null,
      showPreview: false,
    });

    try {
      await get().loadMessages(sessionId);
      get().connectWebSocket(sessionId);
    } catch (error) {
      console.error('Failed to load session data:', error);
      // Load from local storage as fallback
      const localMessages = chatService.getLocalMessages(sessionId);
      set({ messages: localMessages, loading: false });
    }
  },

  deleteSession: async (sessionId: string) => {
    try {
      await chatService.deleteSession(sessionId);
      set(state => ({
        sessions: state.sessions.filter(s => s.id !== sessionId),
        currentSession: state.currentSession?.id === sessionId ? null : state.currentSession,
        messages: state.currentSession?.id === sessionId ? [] : state.messages,
      }));
    } catch (error) {
      console.error('Failed to delete session:', error);
      throw error;
    }
  },

  // Message actions
  loadMessages: async (sessionId: string) => {
    set({ loading: true });
    try {
      const messages = await chatService.getMessages(sessionId);
      set({ messages, loading: false });
      
      // Save locally
      chatService.saveMessagesLocally(sessionId, messages);
    } catch (error) {
      console.error('Failed to load messages:', error);
      set({ loading: false });
      throw error;
    }
  },

  sendMessage: async (content: string) => {
    const { currentSession } = get();
    if (!currentSession) {
      throw new Error('No active session');
    }

    const tempMessage: ChatMessage = {
      id: `temp-${Date.now()}`,
      sessionId: currentSession.id,
      userId: 'current-user',
      type: 'user',
      content,
      timestamp: new Date().toISOString(),
    };

    // Optimistically add message
    set(state => ({
      messages: [...state.messages, tempMessage],
    }));

    try {
      const message = await chatService.sendMessage(currentSession.id, content);
      
      // Replace temp message with real one
      set(state => ({
        messages: state.messages.map(m => 
          m.id === tempMessage.id ? message : m
        ),
      }));

      // Check if this might be a command
      if (isCommandLike(content)) {
        await get().generatePreview(content);
      }

    } catch (error) {
      console.error('Failed to send message:', error);
      
      // Remove temp message on error
      set(state => ({
        messages: state.messages.filter(m => m.id !== tempMessage.id),
      }));
      
      throw error;
    }
  },

  // Command preview actions
  generatePreview: async (naturalLanguage: string) => {
    const { currentSession } = get();
    if (!currentSession) return;

    try {
      const preview = await chatService.generateCommandPreview(naturalLanguage, currentSession.id);
      set({ 
        currentPreview: preview, 
        showPreview: true 
      });
    } catch (error) {
      console.error('Failed to generate command preview:', error);
    }
  },

  approveCommand: async () => {
    const { currentPreview, currentSession } = get();
    if (!currentPreview || !currentSession) return;

    try {
      if (currentPreview.approvalRequired) {
        await chatService.requestApproval(currentPreview.id);
        
        // Add system message about approval request
        const approvalMessage: ChatMessage = {
          id: `approval-${Date.now()}`,
          sessionId: currentSession.id,
          userId: 'system',
          type: 'system',
          content: 'Approval request sent. Waiting for administrator approval.',
          timestamp: new Date().toISOString(),
        };

        set(state => ({
          messages: [...state.messages, approvalMessage],
          showPreview: false,
          currentPreview: null,
        }));
      } else {
        // Execute directly
        const execution = await chatService.executeCommand(currentPreview.id);
        
        // Add execution message
        const executionMessage: ChatMessage = {
          id: `execution-${Date.now()}`,
          sessionId: currentSession.id,
          userId: 'assistant',
          type: 'assistant',
          content: `Command executed successfully. Result: ${execution.result || 'No output'}`,
          timestamp: new Date().toISOString(),
          metadata: {
            command: currentPreview.generatedCommand,
            executionId: execution.id,
            safetyLevel: currentPreview.safetyLevel,
          },
        };

        set(state => ({
          messages: [...state.messages, executionMessage],
          showPreview: false,
          currentPreview: null,
        }));
      }
    } catch (error) {
      console.error('Failed to execute command:', error);
      
      // Add error message
      const errorMessage: ChatMessage = {
        id: `error-${Date.now()}`,
        sessionId: currentSession.id,
        userId: 'system',
        type: 'system',
        content: `Failed to execute command: ${error instanceof Error ? error.message : 'Unknown error'}`,
        timestamp: new Date().toISOString(),
      };

      set(state => ({
        messages: [...state.messages, errorMessage],
        showPreview: false,
        currentPreview: null,
      }));
    }
  },

  rejectCommand: (reason?: string) => {
    const { currentSession } = get();
    
    if (reason && currentSession) {
      const rejectMessage: ChatMessage = {
        id: `reject-${Date.now()}`,
        sessionId: currentSession.id,
        userId: 'system',
        type: 'system',
        content: `Command cancelled${reason ? `: ${reason}` : '.'}`,
        timestamp: new Date().toISOString(),
      };

      set(state => ({
        messages: [...state.messages, rejectMessage],
        showPreview: false,
        currentPreview: null,
      }));
    } else {
      set({
        showPreview: false,
        currentPreview: null,
      });
    }
  },

  // WebSocket actions
  connectWebSocket: (sessionId: string) => {
    const { wsConnection } = get();
    
    // Close existing connection
    if (wsConnection) {
      wsConnection.close();
    }

    try {
      chatService.connectWebSocket(sessionId, (data) => {
        // Handle WebSocket messages
        if (data.type === 'new_message') {
          set(state => ({
            messages: [...state.messages, data.message],
          }));
        } else if (data.type === 'command_status') {
          // Handle command execution status updates
          console.log('Command status update:', data);
        }
      });

      set({ wsConnected: true });
      console.log('WebSocket connected');
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
    }
  },

  disconnectWebSocket: () => {
    const { wsConnection } = get();
    if (wsConnection) {
      wsConnection.close();
      set({ wsConnection: null, wsConnected: false });
    }
  },

  // Utility
  clearAll: () => {
    get().disconnectWebSocket();
    chatService.clearLocalData();
    set({
      sessions: [],
      currentSession: null,
      messages: [],
      loading: false,
      currentPreview: null,
      showPreview: false,
      wsConnection: null,
      wsConnected: false,
    });
  },
}));

// Helper function
function isCommandLike(message: string): boolean {
  const commandKeywords = [
    'kubectl', 'create', 'delete', 'deploy', 'scale', 'restart',
    'get pods', 'describe', 'logs', 'exec', 'apply', 'rollout',
    'list', 'show me', 'check', 'restart', 'update'
  ];
  
  const lowerMessage = message.toLowerCase();
  return commandKeywords.some(keyword => lowerMessage.includes(keyword));
}
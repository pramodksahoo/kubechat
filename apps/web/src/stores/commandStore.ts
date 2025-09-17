// Command Execution Store for Story 2.2
// State management for authenticated command execution and monitoring

import { create } from 'zustand';
import { CommandExecution, CommandPreview } from '../types/chat';
import { commandService, commandHistoryService, nlpService } from '../services/chat';
import { useAuthStore } from './authStore';

interface CommandState {
  // Current executions
  activeExecutions: CommandExecution[];
  executionHistory: CommandExecution[];

  // Command previews
  currentPreview: CommandPreview | null;
  showPreview: boolean;

  // Loading states
  loading: boolean;
  executing: boolean;

  // Error state
  error: string | null;

  // Analytics
  analytics: any | null;

  // Actions
  executeCommand: (command: string, options?: {
    clusterId?: string;
    sessionId?: string;
    previewId?: string;
  }) => Promise<CommandExecution>;

  generatePreview: (naturalLanguage: string, sessionId?: string) => Promise<CommandPreview>;

  approveCommand: (previewId: string) => Promise<void>;
  requestApproval: (previewId: string, reason?: string) => Promise<void>;

  monitorExecution: (executionId: string) => Promise<void>;
  cancelExecution: (executionId: string) => Promise<void>;

  getExecutionHistory: (filters?: any) => Promise<void>;
  getExecutionAnalytics: () => Promise<void>;

  // Preview management
  setCurrentPreview: (preview: CommandPreview | null) => void;
  showCommandPreview: (show: boolean) => void;

  // Utility
  clearError: () => void;
  clearHistory: () => void;
}

export const useCommandStore = create<CommandState>((set, get) => ({
  // Initial state
  activeExecutions: [],
  executionHistory: [],
  currentPreview: null,
  showPreview: false,
  loading: false,
  executing: false,
  error: null,
  analytics: null,

  // Execute command with authentication
  executeCommand: async (command, options = {}) => {
    set({ executing: true, error: null });

    try {
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated) {
        throw new Error('Authentication required to execute commands');
      }

      const execution = await commandService.executeCommand({
        command,
        clusterId: options.clusterId,
        sessionId: options.sessionId,
        previewId: options.previewId,
      });

      // Update active executions
      set(state => ({
        activeExecutions: [execution, ...state.activeExecutions],
        executing: false,
      }));

      // Start monitoring if not completed
      if (!['completed', 'failed', 'cancelled'].includes(execution.status)) {
        get().monitorExecution(execution.id);
      }

      return execution;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Command execution failed';
      set({ error: errorMessage, executing: false });
      throw error;
    }
  },

  // Generate command preview
  generatePreview: async (naturalLanguage, sessionId) => {
    set({ loading: true, error: null });

    try {
      const response = await nlpService.processQuery({
        query: naturalLanguage,
        sessionId,
      });

      const preview: CommandPreview = {
        id: response.id,
        naturalLanguage: response.query,
        generatedCommand: response.generatedCommand || '',
        safetyLevel: response.safetyLevel,
        confidence: response.confidence,
        explanation: response.explanation || '',
        potentialImpact: response.potentialImpact || [],
        requiredPermissions: response.requiredPermissions || [],
        clusterId: 'default',
        approvalRequired: response.approvalRequired || false,
      };

      set({
        currentPreview: preview,
        showPreview: true,
        loading: false,
      });

      return preview;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Preview generation failed';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  // Approve command for execution
  approveCommand: async (previewId) => {
    const { currentPreview } = get();
    if (!currentPreview || currentPreview.id !== previewId) {
      throw new Error('Preview not found');
    }

    try {
      if (currentPreview.approvalRequired) {
        // Request approval from administrators
        await get().requestApproval(previewId, 'User approved command execution');
      } else {
        // Execute directly
        await get().executeCommand(currentPreview.generatedCommand, {
          previewId,
        });
      }

      set({
        currentPreview: null,
        showPreview: false,
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Command approval failed';
      set({ error: errorMessage });
      throw error;
    }
  },

  // Request approval for dangerous operations
  requestApproval: async (previewId, reason) => {
    try {
      const authState = useAuthStore.getState();
      await commandService.requestApproval({
        previewId,
        requestedBy: authState.user?.id || 'unknown',
        reason,
        priority: 'medium',
      });

      // Could add to a pending approvals list here
      console.log('Approval request submitted for preview:', previewId);
    } catch (error) {
      console.error('Failed to request approval:', error);
      throw error;
    }
  },

  // Monitor command execution status
  monitorExecution: async (executionId) => {
    try {
      await commandService.monitorExecution(executionId, (updatedExecution) => {
        set(state => ({
          activeExecutions: state.activeExecutions.map(exec =>
            exec.id === executionId ? updatedExecution : exec
          ),
        }));

        // Move to history if completed
        if (['completed', 'failed', 'cancelled'].includes(updatedExecution.status)) {
          set(state => ({
            activeExecutions: state.activeExecutions.filter(exec => exec.id !== executionId),
            executionHistory: [updatedExecution, ...state.executionHistory],
          }));
        }
      });
    } catch (error) {
      console.error('Failed to monitor execution:', error);
    }
  },

  // Cancel running execution
  cancelExecution: async (executionId) => {
    try {
      await commandService.cancelExecution(executionId);

      set(state => ({
        activeExecutions: state.activeExecutions.filter(exec => exec.id !== executionId),
      }));
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Cancellation failed';
      set({ error: errorMessage });
      throw error;
    }
  },

  // Get execution history with filters
  getExecutionHistory: async (filters = {}) => {
    set({ loading: true, error: null });

    try {
      const history = await commandHistoryService.getCommandHistory(filters);
      set({
        executionHistory: history,
        loading: false,
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load execution history';
      set({ error: errorMessage, loading: false });
    }
  },

  // Get execution analytics
  getExecutionAnalytics: async () => {
    set({ loading: true, error: null });

    try {
      const analytics = await commandHistoryService.getCommandAnalytics();
      set({
        analytics,
        loading: false,
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load analytics';
      set({ error: errorMessage, loading: false });
    }
  },

  // Preview management
  setCurrentPreview: (preview) => {
    set({ currentPreview: preview });
  },

  showCommandPreview: (show) => {
    set({ showPreview: show });
  },

  // Utility actions
  clearError: () => {
    set({ error: null });
  },

  clearHistory: () => {
    set({
      activeExecutions: [],
      executionHistory: [],
      analytics: null,
    });
  },
}));
// Authenticated Command Execution Service for Story 2.2
// Integrates with verified backend command APIs with JWT authentication

import { CommandExecution, CommandPreview } from '../../types/chat';
import { api } from '../api';
import { useAuthStore } from '../../stores/authStore';
import { errorHandlingService } from '../errorHandlingService';

interface ExecuteCommandRequest {
  command: string;
  clusterId?: string;
  sessionId?: string;
  previewId?: string;
}

interface ApprovalRequest {
  previewId: string;
  requestedBy: string;
  reason?: string;
  priority?: 'low' | 'medium' | 'high' | 'critical';
}

interface ExecutionFilters {
  status?: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  userId?: string;
  limit?: number;
  page?: number;
}

interface RollbackPlan {
  id: string;
  executionId: string;
  commands: string[];
  description: string;
  estimatedDuration: number;
}

export class CommandService {
  private activeExecutions = new Map<string, CommandExecution>();
  private executionHistory: CommandExecution[] = [];

  // Task 3.1: Implement CommandExecutionService using /api/v1/commands/execute with JWT authentication
  async executeCommand(request: ExecuteCommandRequest): Promise<CommandExecution> {
    try {
      // Ensure user is authenticated
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated || !authState.user) {
        throw new Error('User must be authenticated to execute commands');
      }

      // Task 3.2: Create command execution workflow with user authorization checks
      const hasPermission = await this.validateExecutionPermissions(request.command, authState.user.permissions || []);
      if (!hasPermission) {
        throw new Error('Insufficient permissions to execute this command');
      }

      // Call authenticated backend API
      const response = await api.commands.execute({
        command: request.command,
        cluster: request.clusterId || 'default',
      });

      const executionData = response.data;

      // Transform to our CommandExecution interface
      const execution: CommandExecution = {
        id: executionData.id || `exec-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`,
        sessionId: request.sessionId || 'default',
        previewId: request.previewId || `preview-${Date.now()}`,
        command: executionData.command,
        status: executionData.status as 'pending' | 'running' | 'completed' | 'failed' | 'cancelled',
        output: executionData.output || '',
        error: executionData.exitCode !== 0 ? 'Command execution failed' : undefined,
        result: executionData.exitCode === 0 ? 'success' : 'failure',
        startedAt: executionData.executedAt ? new Date(executionData.executedAt) : new Date(),
        completedAt: executionData.status === 'completed' || executionData.status === 'failed' ? new Date() : undefined,
        executedBy: authState.user.id,
        approvedBy: undefined, // Will be set if approval was required
      };

      // Track active execution
      this.activeExecutions.set(execution.id, execution);

      // Add to history
      this.executionHistory.unshift(execution);
      this.saveExecutionHistory();

      return execution;
    } catch (error) {
      const errorDetails = await errorHandlingService.handleError(error as Error, {
        context: {
          operation: 'command-execution',
          component: 'CommandService',
        },
        logToConsole: true,
      });

      throw new Error(`Command execution failed: ${errorDetails.type} - ${errorDetails.suggestions.join('. ')}`);
    }
  }

  // Task 3.3: Add command approval integration for dangerous operations
  async requestApproval(request: ApprovalRequest): Promise<{ approvalId: string; status: string }> {
    try {
      const response = await api.commands.approve({
        previewId: request.previewId,
        requestedBy: request.requestedBy,
        requestedAt: new Date().toISOString(),
        reason: request.reason,
        priority: request.priority || 'medium',
      });

      return {
        approvalId: (response.data as any)?.id || 'approval-' + Date.now(),
        status: (response.data as any)?.status || 'pending',
      };
    } catch (error) {
      console.error('Failed to request approval:', error);
      throw new Error('Failed to submit approval request');
    }
  }

  async getPendingApprovals(): Promise<any[]> {
    try {
      const response = await api.commands.getPendingApprovals();
      return (response.data as any)?.approvals || [];
    } catch (error) {
      console.error('Failed to get pending approvals:', error);
      return [];
    }
  }

  // Task 3.4: Implement execution status monitoring with real-time updates
  async getExecutionStatus(executionId: string): Promise<CommandExecution> {
    try {
      // Check local cache first
      const cachedExecution = this.activeExecutions.get(executionId);
      if (cachedExecution && cachedExecution.status === 'completed') {
        return cachedExecution;
      }

      // Fetch latest status from backend
      const response = await api.commands.getExecution(executionId);
      const executionData = response.data;

      const execution: CommandExecution = {
        id: executionData.id,
        sessionId: cachedExecution?.sessionId || 'unknown',
        previewId: cachedExecution?.previewId || 'unknown',
        command: executionData.command,
        status: executionData.status as 'pending' | 'running' | 'completed' | 'failed' | 'cancelled',
        output: executionData.output || '',
        error: executionData.exitCode !== 0 ? 'Command execution failed' : undefined,
        result: executionData.exitCode === 0 ? 'success' : 'failure',
        startedAt: executionData.executedAt ? new Date(executionData.executedAt) : new Date(),
        completedAt: (executionData as any).completedAt ? new Date((executionData as any).completedAt) : undefined,
        executedBy: cachedExecution?.executedBy || 'unknown',
        approvedBy: cachedExecution?.approvedBy,
      };

      // Update cache
      this.activeExecutions.set(executionId, execution);

      // Remove from active if completed
      if (['completed', 'failed', 'cancelled'].includes(execution.status)) {
        this.activeExecutions.delete(executionId);
      }

      return execution;
    } catch (error) {
      console.error('Failed to get execution status:', error);
      throw new Error('Failed to retrieve execution status');
    }
  }

  async monitorExecution(executionId: string, onUpdate: (execution: CommandExecution) => void): Promise<void> {
    const pollInterval = setInterval(async () => {
      try {
        const execution = await this.getExecutionStatus(executionId);
        onUpdate(execution);

        // Stop polling if execution is complete
        if (['completed', 'failed', 'cancelled'].includes(execution.status)) {
          clearInterval(pollInterval);
        }
      } catch (error) {
        console.error('Error monitoring execution:', error);
        clearInterval(pollInterval);
      }
    }, 2000); // Poll every 2 seconds

    // Clean up after 10 minutes
    setTimeout(() => {
      clearInterval(pollInterval);
    }, 600000);
  }

  // Task 3.5: Create command cancellation and abort functionality
  async cancelExecution(executionId: string): Promise<void> {
    try {
      await api.commands.cancelExecution(executionId);

      // Update local state
      const execution = this.activeExecutions.get(executionId);
      if (execution) {
        execution.status = 'cancelled';
        execution.completedAt = new Date();
        this.activeExecutions.delete(executionId);
      }
    } catch (error) {
      console.error('Failed to cancel execution:', error);
      throw new Error('Failed to cancel command execution');
    }
  }

  // Task 3.6: Add execution result processing and display with security filtering
  async getExecutionResults(executionId: string, filterSensitive = true): Promise<{
    output: string;
    error?: string;
    metadata: any;
  }> {
    try {
      const execution = await this.getExecutionStatus(executionId);

      let output = execution.output || '';
      let error = execution.error;

      if (filterSensitive) {
        // Filter sensitive information from output
        output = this.filterSensitiveData(output);
        error = error ? this.filterSensitiveData(error) : undefined;
      }

      return {
        output,
        error,
        metadata: {
          executionId: execution.id,
          command: execution.command,
          status: execution.status,
          result: execution.result,
          duration: execution.completedAt && execution.startedAt
            ? execution.completedAt.getTime() - execution.startedAt.getTime()
            : undefined,
          executedBy: execution.executedBy,
          approvedBy: execution.approvedBy,
        },
      };
    } catch (error) {
      console.error('Failed to get execution results:', error);
      throw new Error('Failed to retrieve execution results');
    }
  }

  // Get command execution history with filtering
  async getExecutionHistory(filters: ExecutionFilters = {}): Promise<CommandExecution[]> {
    try {
      const response = await api.commands.listExecutions({
        page: filters.page || 1,
        limit: filters.limit || 50,
        status: filters.status,
      });

      const executions = response.data.executions.map((exec: any) => ({
        id: exec.id,
        sessionId: exec.sessionId || 'unknown',
        previewId: exec.previewId || 'unknown',
        command: exec.command,
        status: exec.status as 'pending' | 'running' | 'completed' | 'failed' | 'cancelled',
        output: exec.output || '',
        error: exec.error,
        result: exec.exitCode === 0 ? 'success' : 'failure',
        startedAt: exec.executedAt ? new Date(exec.executedAt) : new Date(),
        completedAt: exec.completedAt ? new Date(exec.completedAt) : undefined,
        executedBy: exec.executedBy || 'unknown',
        approvedBy: exec.approvedBy,
      })) as CommandExecution[];

      // Update local cache
      this.executionHistory = executions;
      this.saveExecutionHistory();

      return executions;
    } catch (error) {
      console.error('Failed to get execution history:', error);
      return this.getLocalExecutionHistory();
    }
  }

  // Get execution statistics
  async getExecutionStats(): Promise<{
    total: number;
    successful: number;
    failed: number;
    pending: number;
    averageDuration: number;
  }> {
    try {
      const response = await api.commands.getStats();
      return response.data as any;
    } catch (error) {
      console.error('Failed to get execution stats:', error);
      return {
        total: 0,
        successful: 0,
        failed: 0,
        pending: 0,
        averageDuration: 0,
      };
    }
  }

  // Rollback functionality
  async createRollbackPlan(executionId: string): Promise<RollbackPlan> {
    try {
      const response = await api.commands.createRollbackPlan(executionId);
      return response.data as any;
    } catch (error) {
      console.error('Failed to create rollback plan:', error);
      throw new Error('Failed to create rollback plan');
    }
  }

  async executeRollback(planId: string): Promise<CommandExecution> {
    try {
      const response = await api.commands.executeRollback(planId);
      const execution = response.data as any;

      // Track rollback execution
      this.activeExecutions.set(execution.id, execution);

      return execution;
    } catch (error) {
      console.error('Failed to execute rollback:', error);
      throw new Error('Failed to execute rollback');
    }
  }

  // Validate user permissions for command execution
  private async validateExecutionPermissions(command: string, userPermissions: string[]): Promise<boolean> {
    // Basic permission validation
    if (userPermissions.includes('admin') || userPermissions.includes('*')) {
      return true;
    }

    // Check specific permissions based on command
    if (/get|list|describe/i.test(command)) {
      return userPermissions.some(p => p.includes('read') || p.includes('view'));
    }

    if (/delete/i.test(command)) {
      return userPermissions.some(p => p.includes('delete') || p.includes('admin'));
    }

    if (/create|apply/i.test(command)) {
      return userPermissions.some(p => p.includes('write') || p.includes('create'));
    }

    if (/update|scale|restart/i.test(command)) {
      return userPermissions.some(p => p.includes('update') || p.includes('write'));
    }

    // Default to requiring basic kubernetes access
    return userPermissions.some(p => p.includes('kubernetes'));
  }

  // Filter sensitive data from command output
  private filterSensitiveData(text: string): string {
    if (!text) return text;

    return text
      .replace(/password[=:]\s*[^\s]+/gi, 'password=***')
      .replace(/token[=:]\s*[^\s]+/gi, 'token=***')
      .replace(/secret[=:]\s*[^\s]+/gi, 'secret=***')
      .replace(/key[=:]\s*[^\s]+/gi, 'key=***')
      .replace(/[A-Za-z0-9+/]{40,}/g, '***') // Base64 encoded secrets
      .replace(/-----BEGIN [A-Z ]+-----[\s\S]*?-----END [A-Z ]+-----/g, '*** CERTIFICATE ***');
  }

  // Local storage for offline functionality
  private saveExecutionHistory(): void {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        const data = {
          executions: this.executionHistory.slice(0, 100), // Keep only last 100
          lastUpdated: new Date().toISOString(),
        };
        localStorage.setItem('kubechat_execution_history', JSON.stringify(data));
      }
    } catch (error) {
      console.error('Failed to save execution history:', error);
    }
  }

  private getLocalExecutionHistory(): CommandExecution[] {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        const stored = localStorage.getItem('kubechat_execution_history');
        if (stored) {
          const data = JSON.parse(stored);
          return data.executions || [];
        }
      }
    } catch (error) {
      console.error('Failed to load local execution history:', error);
    }
    return [];
  }

  // Get active executions
  getActiveExecutions(): CommandExecution[] {
    return Array.from(this.activeExecutions.values());
  }

  // Clear all data
  clearAll(): void {
    this.activeExecutions.clear();
    this.executionHistory = [];

    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.removeItem('kubechat_execution_history');
    }
  }
}

// Singleton instance
export const commandService = new CommandService();
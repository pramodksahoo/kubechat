// Command History and Audit Service for Story 2.2
// Implements command history with RBAC filtering and audit compliance

import { CommandExecution } from '../../types/chat';
import { api } from '../api';
import { useAuthStore } from '../../stores/authStore';
import { errorHandlingService } from '../errorHandlingService';

interface HistoryFilters {
  userId?: string;
  sessionId?: string;
  status?: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  command?: string;
  startDate?: Date;
  endDate?: Date;
  safetyLevel?: 'safe' | 'warning' | 'dangerous';
  limit?: number;
  offset?: number;
}

interface SearchFilters {
  query: string;
  fields?: ('command' | 'output' | 'error')[];
  fuzzy?: boolean;
}

interface ExecutionAnalytics {
  totalExecutions: number;
  successRate: number;
  failureRate: number;
  averageDuration: number;
  mostUsedCommands: Array<{ command: string; count: number }>;
  errorPatterns: Array<{ pattern: string; count: number; suggestion: string }>;
  userActivity: Array<{ userId: string; executionCount: number }>;
}

interface ExportOptions {
  format: 'json' | 'csv' | 'pdf';
  includeOutput?: boolean;
  includeSensitive?: boolean;
  dateRange?: { start: Date; end: Date };
}

export class CommandHistoryService {
  private historyCache: CommandExecution[] = [];
  private analyticsCache: ExecutionAnalytics | null = null;
  private cacheExpiry: number = 0;
  private readonly CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

  // Task 5.1: Create CommandHistoryService using /api/v1/commands/executions API
  async getCommandHistory(filters: HistoryFilters = {}): Promise<CommandExecution[]> {
    try {
      // Ensure user is authenticated
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated || !authState.user) {
        throw new Error('Authentication required to access command history');
      }

      // Task 5.2: Implement user-specific command history filtering with RBAC
      const rbacFilters = this.applyRBACFilters(filters, authState.user.permissions || [], authState.user.role);

      // Check cache first
      if (this.isCacheValid() && this.historyCache.length > 0) {
        return this.filterHistoryLocally(this.historyCache, rbacFilters);
      }

      // Call backend API
      const response = await api.commands.listExecutions({
        page: Math.floor((rbacFilters.offset || 0) / (rbacFilters.limit || 50)) + 1,
        limit: rbacFilters.limit || 50,
        status: rbacFilters.status,
      });

      // Transform response to our CommandExecution interface
      const executions: CommandExecution[] = response.data.executions.map((exec: any) => ({
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
      }));

      // Apply additional client-side filtering
      const filteredExecutions = this.filterHistoryLocally(executions, rbacFilters);

      // Update cache
      this.historyCache = executions;
      this.cacheExpiry = Date.now() + this.CACHE_DURATION;

      return filteredExecutions;
    } catch (error) {
      const errorDetails = await errorHandlingService.handleError(error as Error, {
        context: {
          operation: 'get-command-history',
          component: 'CommandHistoryService',
        },
        logToConsole: true,
      });

      // Return cached data if available
      if (this.historyCache.length > 0) {
        console.warn('Using cached command history due to API error:', errorDetails);
        return this.filterHistoryLocally(this.historyCache, filters);
      }

      throw new Error(`Failed to retrieve command history: ${errorDetails.type}`);
    }
  }

  // Task 5.3: Add command history search and filtering capabilities
  async searchCommandHistory(searchFilters: SearchFilters, historyFilters: HistoryFilters = {}): Promise<CommandExecution[]> {
    try {
      // Get full history first
      const allHistory = await this.getCommandHistory(historyFilters);

      // Apply search filters
      const searchFields = searchFilters.fields || ['command', 'output', 'error'];
      const query = searchFilters.query.toLowerCase();

      const results = allHistory.filter(execution => {
        return searchFields.some(field => {
          const fieldValue = execution[field as keyof CommandExecution];
          if (typeof fieldValue === 'string') {
            if (searchFilters.fuzzy) {
              return this.fuzzyMatch(fieldValue.toLowerCase(), query);
            } else {
              return fieldValue.toLowerCase().includes(query);
            }
          }
          return false;
        });
      });

      return results;
    } catch (error) {
      console.error('Command history search failed:', error);
      throw new Error('Failed to search command history');
    }
  }

  // Task 5.4: Create execution timeline and status progression display
  async getExecutionTimeline(executionId: string): Promise<{
    execution: CommandExecution;
    timeline: Array<{
      timestamp: Date;
      status: string;
      message: string;
      duration?: number;
    }>;
  }> {
    try {
      const response = await api.commands.getExecution(executionId);
      const executionData = response.data as any;

      const execution: CommandExecution = {
        id: executionData.id,
        sessionId: executionData.sessionId || 'unknown',
        previewId: executionData.previewId || 'unknown',
        command: executionData.command,
        status: (executionData.status === 'completed' || executionData.status === 'pending' || executionData.status === 'running' || executionData.status === 'failed' || executionData.status === 'cancelled')
          ? executionData.status as 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
          : 'completed',
        output: executionData.output || '',
        error: executionData.error || (executionData.exitCode !== 0 ? 'Command failed' : undefined),
        result: executionData.exitCode === 0 ? 'success' : 'failure',
        startedAt: new Date(executionData.executedAt),
        completedAt: executionData.completedAt ? new Date(executionData.completedAt) : undefined,
        executedBy: executionData.executedBy || 'unknown',
        approvedBy: executionData.approvedBy || undefined,
      };

      // Build timeline from execution data
      const timeline = this.buildExecutionTimeline(execution);

      return { execution, timeline };
    } catch (error) {
      console.error('Failed to get execution timeline:', error);
      throw new Error('Failed to retrieve execution timeline');
    }
  }

  // Task 5.5: Implement command history export for audit and compliance
  async exportCommandHistory(filters: HistoryFilters, options: ExportOptions): Promise<{
    data: string | Blob;
    filename: string;
    mimeType: string;
  }> {
    try {
      const history = await this.getCommandHistory(filters);

      // Filter sensitive data if requested
      const exportData = options.includeSensitive
        ? history
        : history.map(exec => this.sanitizeExecutionForExport(exec));

      switch (options.format) {
        case 'json':
          return {
            data: JSON.stringify(exportData, null, 2),
            filename: `command-history-${new Date().toISOString().split('T')[0]}.json`,
            mimeType: 'application/json',
          };

        case 'csv':
          return {
            data: this.convertToCSV(exportData, options.includeOutput),
            filename: `command-history-${new Date().toISOString().split('T')[0]}.csv`,
            mimeType: 'text/csv',
          };

        case 'pdf':
          return await this.generatePDFReport(exportData, options);

        default:
          throw new Error('Unsupported export format');
      }
    } catch (error) {
      console.error('Command history export failed:', error);
      throw new Error('Failed to export command history');
    }
  }

  // Task 5.6: Add command history analytics and pattern recognition
  async getCommandAnalytics(filters: HistoryFilters = {}): Promise<ExecutionAnalytics> {
    try {
      // Check cache first
      if (this.analyticsCache && this.isCacheValid()) {
        return this.analyticsCache;
      }

      const history = await this.getCommandHistory(filters);

      const analytics: ExecutionAnalytics = {
        totalExecutions: history.length,
        successRate: this.calculateSuccessRate(history),
        failureRate: this.calculateFailureRate(history),
        averageDuration: this.calculateAverageDuration(history),
        mostUsedCommands: this.analyzeMostUsedCommands(history),
        errorPatterns: this.analyzeErrorPatterns(history),
        userActivity: this.analyzeUserActivity(history),
      };

      // Cache analytics
      this.analyticsCache = analytics;

      return analytics;
    } catch (error) {
      console.error('Failed to get command analytics:', error);
      throw new Error('Failed to retrieve command analytics');
    }
  }

  // Get command execution patterns
  async getExecutionPatterns(): Promise<{
    timePatterns: Array<{ hour: number; count: number }>;
    commandPatterns: Array<{ pattern: string; count: number; successRate: number }>;
    clusterPatterns: Array<{ clusterId: string; commandCount: number; errorRate: number }>;
  }> {
    try {
      const history = await this.getCommandHistory();

      return {
        timePatterns: this.analyzeTimePatterns(history),
        commandPatterns: this.analyzeCommandPatterns(history),
        clusterPatterns: this.analyzeClusterPatterns(history),
      };
    } catch (error) {
      console.error('Failed to get execution patterns:', error);
      throw new Error('Failed to retrieve execution patterns');
    }
  }

  // RBAC filtering
  private applyRBACFilters(filters: HistoryFilters, permissions: string[], role: string): HistoryFilters {
    const rbacFilters = { ...filters };

    // Users can only see their own executions unless they have admin permissions
    if (!permissions.includes('admin') && !permissions.includes('audit:read')) {
      rbacFilters.userId = useAuthStore.getState().user?.id;
    }

    // Limit history scope based on role
    if (role === 'user') {
      rbacFilters.limit = Math.min(rbacFilters.limit || 50, 100);
    }

    return rbacFilters;
  }

  // Local filtering
  private filterHistoryLocally(executions: CommandExecution[], filters: HistoryFilters): CommandExecution[] {
    return executions.filter(exec => {
      if (filters.userId && exec.executedBy !== filters.userId) return false;
      if (filters.sessionId && exec.sessionId !== filters.sessionId) return false;
      if (filters.status && exec.status !== filters.status) return false;
      if (filters.command && !exec.command.toLowerCase().includes(filters.command.toLowerCase())) return false;
      if (filters.startDate && exec.startedAt < filters.startDate) return false;
      if (filters.endDate && exec.startedAt > filters.endDate) return false;

      return true;
    }).slice(filters.offset || 0, (filters.offset || 0) + (filters.limit || 50));
  }

  // Analytics helpers
  private calculateSuccessRate(history: CommandExecution[]): number {
    if (history.length === 0) return 0;
    const successful = history.filter(exec => exec.result === 'success').length;
    return (successful / history.length) * 100;
  }

  private calculateFailureRate(history: CommandExecution[]): number {
    return 100 - this.calculateSuccessRate(history);
  }

  private calculateAverageDuration(history: CommandExecution[]): number {
    const completedExecutions = history.filter(exec => exec.completedAt);
    if (completedExecutions.length === 0) return 0;

    const totalDuration = completedExecutions.reduce((sum, exec) => {
      return sum + (exec.completedAt!.getTime() - exec.startedAt.getTime());
    }, 0);

    return totalDuration / completedExecutions.length;
  }

  private analyzeMostUsedCommands(history: CommandExecution[]): Array<{ command: string; count: number }> {
    const commandCounts = new Map<string, number>();

    history.forEach(exec => {
      const baseCommand = exec.command.split(' ')[0];
      commandCounts.set(baseCommand, (commandCounts.get(baseCommand) || 0) + 1);
    });

    return Array.from(commandCounts.entries())
      .map(([command, count]) => ({ command, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10);
  }

  private analyzeErrorPatterns(history: CommandExecution[]): Array<{ pattern: string; count: number; suggestion: string }> {
    const errorPatterns = new Map<string, number>();
    const suggestions = new Map<string, string>();

    history.filter(exec => exec.error).forEach(exec => {
      const errorType = this.categorizeError(exec.error!);
      errorPatterns.set(errorType, (errorPatterns.get(errorType) || 0) + 1);
      if (!suggestions.has(errorType)) {
        suggestions.set(errorType, this.getErrorSuggestion(errorType));
      }
    });

    return Array.from(errorPatterns.entries())
      .map(([pattern, count]) => ({
        pattern,
        count,
        suggestion: suggestions.get(pattern) || 'Review command syntax and permissions',
      }))
      .sort((a, b) => b.count - a.count);
  }

  private analyzeUserActivity(history: CommandExecution[]): Array<{ userId: string; executionCount: number }> {
    const userActivity = new Map<string, number>();

    history.forEach(exec => {
      userActivity.set(exec.executedBy, (userActivity.get(exec.executedBy) || 0) + 1);
    });

    return Array.from(userActivity.entries())
      .map(([userId, executionCount]) => ({ userId, executionCount }))
      .sort((a, b) => b.executionCount - a.executionCount);
  }

  // Pattern analysis helpers
  private analyzeTimePatterns(history: CommandExecution[]): Array<{ hour: number; count: number }> {
    const hourCounts = Array(24).fill(0);

    history.forEach(exec => {
      const hour = exec.startedAt.getHours();
      hourCounts[hour]++;
    });

    return hourCounts.map((count, hour) => ({ hour, count }));
  }

  private analyzeCommandPatterns(history: CommandExecution[]): Array<{ pattern: string; count: number; successRate: number }> {
    const patterns = new Map<string, { count: number; successes: number }>();

    history.forEach(exec => {
      const pattern = this.extractCommandPattern(exec.command);
      const current = patterns.get(pattern) || { count: 0, successes: 0 };
      current.count++;
      if (exec.result === 'success') current.successes++;
      patterns.set(pattern, current);
    });

    return Array.from(patterns.entries())
      .map(([pattern, stats]) => ({
        pattern,
        count: stats.count,
        successRate: (stats.successes / stats.count) * 100,
      }))
      .sort((a, b) => b.count - a.count);
  }

  private analyzeClusterPatterns(_history: CommandExecution[]): Array<{ clusterId: string; commandCount: number; errorRate: number }> {
    // This would require cluster information in execution records
    // For now, return empty array as cluster info isn't available in current structure
    return [];
  }

  // Utility methods
  private buildExecutionTimeline(execution: CommandExecution): Array<{
    timestamp: Date;
    status: string;
    message: string;
    duration?: number;
  }> {
    const timeline: Array<{
      timestamp: Date;
      status: string;
      message: string;
      duration?: number;
    }> = [
      {
        timestamp: execution.startedAt,
        status: 'started',
        message: 'Command execution started',
      },
    ];

    if (execution.completedAt) {
      timeline.push({
        timestamp: execution.completedAt,
        status: execution.status,
        message: `Command execution ${execution.status}`,
        duration: execution.completedAt.getTime() - execution.startedAt.getTime(),
      });
    }

    return timeline;
  }

  private sanitizeExecutionForExport(execution: CommandExecution): Partial<CommandExecution> {
    return {
      ...execution,
      output: this.sanitizeSensitiveData(execution.output || ''),
      error: execution.error ? this.sanitizeSensitiveData(execution.error) : undefined,
    };
  }

  private sanitizeSensitiveData(text: string): string {
    return text
      .replace(/password[=:]\s*[^\s]+/gi, 'password=***')
      .replace(/token[=:]\s*[^\s]+/gi, 'token=***')
      .replace(/secret[=:]\s*[^\s]+/gi, 'secret=***')
      .replace(/key[=:]\s*[^\s]+/gi, 'key=***');
  }

  private convertToCSV(data: any[], includeOutput = false): string {
    if (data.length === 0) return '';

    const headers = ['id', 'command', 'status', 'result', 'executedBy', 'startedAt', 'completedAt'];
    if (includeOutput) headers.push('output');

    const csvLines = [headers.join(',')];

    data.forEach(execution => {
      const row = headers.map(header => {
        const value = execution[header];
        if (typeof value === 'string') {
          return `"${value.replace(/"/g, '""')}"`;
        }
        return value || '';
      });
      csvLines.push(row.join(','));
    });

    return csvLines.join('\n');
  }

  private async generatePDFReport(data: any[], _options: ExportOptions): Promise<{
    data: Blob;
    filename: string;
    mimeType: string;
  }> {
    // In a real implementation, this would use a PDF library like jsPDF
    // For now, return a simple text blob
    const content = `Command History Report\n\nGenerated: ${new Date().toISOString()}\n\n${JSON.stringify(data, null, 2)}`;
    return {
      data: new Blob([content], { type: 'text/plain' }),
      filename: `command-history-report-${new Date().toISOString().split('T')[0]}.txt`,
      mimeType: 'text/plain',
    };
  }

  private fuzzyMatch(text: string, query: string): boolean {
    const textChars = text.split('');
    const queryChars = query.split('');
    let queryIndex = 0;

    for (const char of textChars) {
      if (queryIndex < queryChars.length && char === queryChars[queryIndex]) {
        queryIndex++;
      }
    }

    return queryIndex === queryChars.length;
  }

  private categorizeError(error: string): string {
    if (error.includes('permission') || error.includes('forbidden')) return 'Permission Error';
    if (error.includes('not found') || error.includes('404')) return 'Resource Not Found';
    if (error.includes('timeout')) return 'Timeout Error';
    if (error.includes('connection')) return 'Connection Error';
    if (error.includes('syntax')) return 'Syntax Error';
    return 'Unknown Error';
  }

  private getErrorSuggestion(errorType: string): string {
    const suggestions = {
      'Permission Error': 'Check user permissions and RBAC configuration',
      'Resource Not Found': 'Verify resource names and namespaces',
      'Timeout Error': 'Increase timeout or check resource availability',
      'Connection Error': 'Check network connectivity and cluster health',
      'Syntax Error': 'Review command syntax and parameters',
      'Unknown Error': 'Check logs for detailed error information',
    };
    return suggestions[errorType as keyof typeof suggestions] || 'Review command and try again';
  }

  private extractCommandPattern(command: string): string {
    const parts = command.split(' ');
    if (parts.length < 2) return command;
    return `${parts[0]} ${parts[1]}`;
  }

  private isCacheValid(): boolean {
    return Date.now() < this.cacheExpiry;
  }

  // Clear cache
  clearCache(): void {
    this.historyCache = [];
    this.analyticsCache = null;
    this.cacheExpiry = 0;
  }
}

// Singleton instance
export const commandHistoryService = new CommandHistoryService();
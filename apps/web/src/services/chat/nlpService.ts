// Authenticated Natural Language Processing Service for Story 2.2
// Integrates with verified backend NLP APIs with security validation

import { api } from '../api';
import { useAuthStore } from '../../stores/authStore';
import { errorHandlingService } from '../errorHandlingService';

interface NLPProcessRequest {
  query: string;
  context?: string;
  clusterId?: string;
  sessionId?: string;
}

interface NLPProcessResponse {
  id: string;
  query: string;
  generatedCommand?: string;
  explanation?: string;
  safetyLevel: 'safe' | 'warning' | 'dangerous';
  confidence: number;
  potentialImpact?: string[];
  requiredPermissions?: string[];
  approvalRequired?: boolean;
}

interface CommandValidationRequest {
  command: string;
  context?: string;
}

interface CommandValidationResponse {
  valid: boolean;
  safetyLevel: string;
  warnings: string[];
  suggestions: string[];
}

interface CommandClassificationRequest {
  command: string;
  context?: string;
}

interface CommandClassificationResponse {
  classification: string;
  riskLevel: string;
  confidence: number;
}

export class NLPService {
  private readonly contextCache = new Map<string, string>();

  // Task 2.1: Create NLPService using authenticated /api/v1/nlp/process endpoint
  async processQuery(request: NLPProcessRequest): Promise<NLPProcessResponse> {
    try {
      // Ensure user is authenticated
      const authState = useAuthStore.getState();
      if (!authState.isAuthenticated || !authState.user) {
        throw new Error('User must be authenticated to process natural language queries');
      }

      // Task 2.2: Implement query processing with security context and user permissions
      const enhancedRequest = {
        query: this.sanitizeQuery(request.query),
        context: request.context || await this.buildSecurityContext(authState.user.id, request.clusterId),
        cluster_info: request.clusterId ? await this.getClusterInfo(request.clusterId) : undefined,
        provider: 'default', // Use default NLP provider
      };

      // Call authenticated backend API
      const response = await api.nlp.processQuery(enhancedRequest);
      const nlpData = response.data;

      // Task 2.3: Add command safety classification with user-specific validation
      const safetyLevel = await this.validateCommandSafety(
        nlpData.generated_command || '',
        authState.user.permissions || []
      );

      // Task 2.4: Create command explanation and preview with security warnings
      const enhancedResponse: NLPProcessResponse = {
        id: nlpData.id,
        query: nlpData.query,
        generatedCommand: nlpData.generated_command,
        explanation: this.enhanceExplanation(nlpData.explanation || '', safetyLevel),
        safetyLevel: safetyLevel,
        confidence: nlpData.confidence,
        potentialImpact: nlpData.potential_impact || this.generatePotentialImpact(nlpData.generated_command || ''),
        requiredPermissions: nlpData.required_permissions || this.extractRequiredPermissions(nlpData.generated_command || ''),
        approvalRequired: nlpData.approval_required || this.requiresApproval(safetyLevel, authState.user.permissions || []),
      };

      // Cache context for conversation continuity
      if (request.sessionId) {
        this.contextCache.set(request.sessionId, enhancedRequest.context);
      }

      return enhancedResponse;
    } catch (error) {
      const errorDetails = await errorHandlingService.handleError(error as Error, {
        context: {
          operation: 'nlp-query-processing',
          component: 'NLPService',
        },
        logToConsole: true,
      });

      // Task 2.6: Add natural language context management and conversation history
      throw new Error(`NLP processing failed: ${errorDetails.type} - ${errorDetails.suggestions.join('. ')}`);
    }
  }

  // Task 2.5: Implement query validation and sanitization for security
  private sanitizeQuery(query: string): string {
    // Remove potentially dangerous characters and patterns
    let sanitized = query
      .replace(/[<>\"']/g, '') // Remove HTML/XML characters
      .replace(/\${.*?}/g, '') // Remove shell variable substitutions
      .replace(/`.*?`/g, '') // Remove command substitutions
      .replace(/\|.*?;/g, '') // Remove pipe chains
      .trim();

    // Limit query length for security
    if (sanitized.length > 2000) {
      sanitized = sanitized.substring(0, 2000);
    }

    return sanitized;
  }

  // Build security context with user permissions and cluster info
  private async buildSecurityContext(userId: string, clusterId?: string): Promise<string> {
    try {
      const authState = useAuthStore.getState();
      const userRole = authState.user?.role || 'user';
      const permissions = authState.user?.permissions || [];

      let context = `User: ${userId}, Role: ${userRole}, Permissions: ${permissions.join(', ')}`;

      if (clusterId) {
        const clusterInfo = await this.getClusterInfo(clusterId);
        context += `, Cluster: ${clusterInfo}`;
      }

      return context;
    } catch (error) {
      console.error('Failed to build security context:', error);
      return `User: ${userId}, Role: user, Permissions: basic`;
    }
  }

  // Get cluster information for context
  private async getClusterInfo(clusterId: string): Promise<string> {
    try {
      const response = await api.clusters.getCluster();
      const cluster = response.data as any;
      return `${cluster.name || clusterId} (${cluster.version || 'unknown'})`;
    } catch {
      return clusterId;
    }
  }

  // Validate command safety based on user permissions
  private async validateCommandSafety(command: string, userPermissions: string[]): Promise<'safe' | 'warning' | 'dangerous'> {
    if (!command) return 'safe';

    // Check for dangerous operations
    const dangerousPatterns = [
      /delete\s+namespace/i,
      /delete\s+.*--all/i,
      /rm\s+-rf/i,
      /kubectl\s+delete\s+.*cluster/i,
      /destroy/i,
      /terminate/i,
    ];

    const warningPatterns = [
      /delete\s+pod/i,
      /delete\s+deployment/i,
      /scale.*--replicas=0/i,
      /restart/i,
      /rollout/i,
    ];

    // Check if user has admin permissions
    const hasAdminPermissions = userPermissions.some(p =>
      p.includes('admin') || p.includes('cluster:delete') || p.includes('*')
    );

    for (const pattern of dangerousPatterns) {
      if (pattern.test(command)) {
        return hasAdminPermissions ? 'warning' : 'dangerous';
      }
    }

    for (const pattern of warningPatterns) {
      if (pattern.test(command)) {
        return 'warning';
      }
    }

    return 'safe';
  }

  // Enhance explanation with security warnings
  private enhanceExplanation(explanation: string, safetyLevel: 'safe' | 'warning' | 'dangerous'): string {
    let enhanced = explanation;

    switch (safetyLevel) {
      case 'dangerous':
        enhanced += '\n\n⚠️ **DANGER**: This command could cause significant damage to your cluster or data. Admin approval is required.';
        break;
      case 'warning':
        enhanced += '\n\n⚡ **CAUTION**: This command may cause service disruption. Please review carefully before executing.';
        break;
      case 'safe':
        enhanced += '\n\n✅ **SAFE**: This is a read-only operation that will not modify your cluster.';
        break;
    }

    return enhanced;
  }

  // Generate potential impact assessment
  private generatePotentialImpact(command: string): string[] {
    if (!command) return ['No command generated'];

    const impacts: string[] = [];

    if (/get|describe|list/i.test(command)) {
      impacts.push('Read-only operation', 'No cluster changes');
    }

    if (/delete/i.test(command)) {
      if (/namespace/i.test(command)) {
        impacts.push('Complete namespace deletion', 'All resources in namespace will be lost', 'Potential data loss');
      } else if (/pod/i.test(command)) {
        impacts.push('Pod termination', 'Possible service interruption', 'Pod may be recreated by controller');
      }
    }

    if (/scale.*0/i.test(command)) {
      impacts.push('Service unavailability', 'All instances stopped', 'Users will experience downtime');
    }

    if (/restart|rollout/i.test(command)) {
      impacts.push('Rolling restart of pods', 'Brief service interruption during restart');
    }

    if (/apply|create/i.test(command)) {
      impacts.push('New resources created', 'Cluster configuration changes');
    }

    return impacts.length > 0 ? impacts : ['Unknown impact - manual review recommended'];
  }

  // Extract required permissions for command
  private extractRequiredPermissions(command: string): string[] {
    const permissions: string[] = [];

    if (/get|list|describe/i.test(command)) {
      permissions.push('kubernetes:read');
    }

    if (/create|apply/i.test(command)) {
      permissions.push('kubernetes:write');
    }

    if (/delete/i.test(command)) {
      permissions.push('kubernetes:delete');
      if (/namespace/i.test(command)) {
        permissions.push('admin:cluster');
      }
    }

    if (/scale|restart|rollout/i.test(command)) {
      permissions.push('kubernetes:update');
    }

    return permissions.length > 0 ? permissions : ['kubernetes:read'];
  }

  // Determine if approval is required
  private requiresApproval(safetyLevel: 'safe' | 'warning' | 'dangerous', userPermissions: string[]): boolean {
    if (safetyLevel === 'dangerous') return true;
    if (safetyLevel === 'warning') {
      // Warning commands require approval unless user has elevated permissions
      return !userPermissions.some(p => p.includes('admin') || p.includes('elevated'));
    }
    return false;
  }

  // Command validation endpoint
  async validateCommand(request: CommandValidationRequest): Promise<CommandValidationResponse> {
    try {
      const response = await api.nlp.validateCommand(request);
      const data = response.data as any;
      return {
        valid: data.valid,
        safetyLevel: data.safety_level || data.safetyLevel,
        warnings: data.warnings || [],
        suggestions: data.suggestions || [],
      };
    } catch (error) {
      console.error('Command validation failed:', error);
      // Return safe default
      return {
        valid: false,
        safetyLevel: 'dangerous',
        warnings: ['Unable to validate command - treating as dangerous'],
        suggestions: ['Please review the command manually', 'Consider using a simpler alternative'],
      };
    }
  }

  // Command classification endpoint
  async classifyCommand(request: CommandClassificationRequest): Promise<CommandClassificationResponse> {
    try {
      const response = await api.nlp.classifyCommand(request);
      const data = response.data as any;
      return {
        classification: data.classification,
        riskLevel: data.risk_level || data.riskLevel,
        confidence: data.confidence,
      };
    } catch (error) {
      console.error('Command classification failed:', error);
      return {
        classification: 'unknown',
        riskLevel: 'high',
        confidence: 0.0,
      };
    }
  }

  // Get available NLP providers
  async getProviders(): Promise<{ providers: string[]; defaultProvider: string }> {
    try {
      const response = await api.nlp.getProviders();
      const data = response.data as any;
      return {
        providers: data.providers,
        defaultProvider: data.default_provider || data.defaultProvider,
      };
    } catch (error) {
      console.error('Failed to get NLP providers:', error);
      return {
        providers: ['default'],
        defaultProvider: 'default',
      };
    }
  }

  // Check NLP service health
  async getHealth(): Promise<{ status: string; providers: Record<string, string> }> {
    try {
      const response = await api.nlp.getHealth();
      return response.data;
    } catch (error) {
      console.error('NLP health check failed:', error);
      return {
        status: 'unhealthy',
        providers: {},
      };
    }
  }

  // Clear cached context for session
  clearContext(sessionId: string): void {
    this.contextCache.delete(sessionId);
  }

  // Get cached context for session
  getContext(sessionId: string): string | undefined {
    return this.contextCache.get(sessionId);
  }
}

// Singleton instance
export const nlpService = new NLPService();
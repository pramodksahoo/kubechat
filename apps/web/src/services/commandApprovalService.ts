import { CommandPreview } from '@kubechat/shared/types';

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
  approvalSteps?: ApprovalStep[];
}

export interface ApprovalStep {
  id: string;
  name: string;
  description: string;
  status: 'pending' | 'completed' | 'skipped';
  completedBy?: string;
  completedAt?: string;
  required: boolean;
  order: number;
}

export interface ApprovalWorkflow {
  id: string;
  name: string;
  description: string;
  safetyLevel: 'safe' | 'warning' | 'dangerous';
  steps: ApprovalStep[];
  autoExpireAfter: number; // minutes
}

class CommandApprovalService {
  private baseUrl: string;
  private approvalListeners: ((requests: ApprovalRequest[]) => void)[] = [];

  constructor(baseUrl: string = process.env.NEXT_PUBLIC_API_URL || '/api/v1') {
    this.baseUrl = baseUrl;
  }

  // Create approval request for dangerous operations
  async createApprovalRequest(commandPreview: CommandPreview): Promise<ApprovalRequest> {
    try {
      const workflow = this.getWorkflowForSafetyLevel(commandPreview.safetyLevel);

      const approvalRequest: ApprovalRequest = {
        id: `approval-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        commandPreview,
        requestedBy: this.getCurrentUserId(),
        requestedAt: new Date().toISOString(),
        status: 'pending',
        expiresAt: new Date(Date.now() + workflow.autoExpireAfter * 60000).toISOString(),
        approvalSteps: workflow.steps.map(step => ({
          ...step,
          status: 'pending'
        }))
      };

      // For dangerous operations, require senior admin approval
      if (commandPreview.safetyLevel === 'dangerous') {
        await this.requestSeniorAdminApproval(approvalRequest);
      }

      // Store the request (in real implementation, would call API)
      await this.storeApprovalRequest(approvalRequest);

      this.notifyApprovalListeners();
      return approvalRequest;
    } catch (error) {
      console.error('Failed to create approval request:', error);
      throw error;
    }
  }

  // Get approval workflow based on safety level
  private getWorkflowForSafetyLevel(safetyLevel: string): ApprovalWorkflow {
    const workflows: Record<string, ApprovalWorkflow> = {
      safe: {
        id: 'safe-workflow',
        name: 'Safe Operation Workflow',
        description: 'Standard workflow for safe operations',
        safetyLevel: 'safe',
        steps: [
          {
            id: 'user-confirm',
            name: 'User Confirmation',
            description: 'User confirms the operation',
            status: 'pending',
            required: true,
            order: 1
          }
        ],
        autoExpireAfter: 30 // 30 minutes
      },
      warning: {
        id: 'warning-workflow',
        name: 'Caution Required Workflow',
        description: 'Enhanced workflow for potentially risky operations',
        safetyLevel: 'warning',
        steps: [
          {
            id: 'impact-review',
            name: 'Impact Assessment Review',
            description: 'Review potential impact of the operation',
            status: 'pending',
            required: true,
            order: 1
          },
          {
            id: 'admin-approval',
            name: 'Admin Approval',
            description: 'Admin reviews and approves the operation',
            status: 'pending',
            required: true,
            order: 2
          }
        ],
        autoExpireAfter: 60 // 1 hour
      },
      dangerous: {
        id: 'dangerous-workflow',
        name: 'High-Risk Operation Workflow',
        description: 'Multi-step approval for dangerous operations',
        safetyLevel: 'dangerous',
        steps: [
          {
            id: 'senior-admin-approval',
            name: 'Senior Admin Approval',
            description: 'Senior administrator must approve this dangerous operation',
            status: 'pending',
            required: true,
            order: 1
          },
          {
            id: 'impact-assessment',
            name: 'Impact Assessment Review',
            description: 'Comprehensive review of potential impact and risks',
            status: 'pending',
            required: true,
            order: 2
          },
          {
            id: 'final-approval',
            name: 'Final Execution Approval',
            description: 'Final approval before command execution',
            status: 'pending',
            required: true,
            order: 3
          }
        ],
        autoExpireAfter: 120 // 2 hours
      }
    };

    return workflows[safetyLevel] || workflows.safe;
  }

  // Request senior admin approval for dangerous operations
  private async requestSeniorAdminApproval(approvalRequest: ApprovalRequest): Promise<void> {
    // In real implementation, this would:
    // 1. Send notifications to senior admins
    // 2. Create audit log entry
    // 3. Set up escalation if not approved within time limit

    console.log(`Requesting senior admin approval for dangerous operation: ${approvalRequest.id}`);

    // Mock approval request notification
    const notification = {
      type: 'approval_required',
      title: 'Dangerous Operation Approval Required',
      message: `${approvalRequest.requestedBy} has requested approval for a dangerous operation: ${approvalRequest.commandPreview.naturalLanguage}`,
      severity: 'high',
      approvalRequestId: approvalRequest.id,
      timestamp: new Date().toISOString()
    };

    // Would send to senior admins via WebSocket/email/Slack
    this.sendApprovalNotification(notification);
  }

  // Approve an approval request
  async approveRequest(requestId: string, approverId: string, stepId?: string): Promise<ApprovalRequest> {
    try {
      const request = await this.getApprovalRequest(requestId);
      if (!request) {
        throw new Error('Approval request not found');
      }

      if (request.status !== 'pending') {
        throw new Error('Request is no longer pending');
      }

      if (stepId) {
        // Approve specific step
        const step = request.approvalSteps?.find(s => s.id === stepId);
        if (step) {
          step.status = 'completed';
          step.completedBy = approverId;
          step.completedAt = new Date().toISOString();
        }

        // Check if all required steps are completed
        const allRequiredStepsCompleted = request.approvalSteps?.every(
          step => !step.required || step.status === 'completed'
        ) ?? false;

        if (allRequiredStepsCompleted) {
          request.status = 'approved';
          request.approvedBy = approverId;
          request.approvedAt = new Date().toISOString();
        }
      } else {
        // Approve entire request
        request.status = 'approved';
        request.approvedBy = approverId;
        request.approvedAt = new Date().toISOString();

        // Mark all steps as completed
        request.approvalSteps?.forEach(step => {
          step.status = 'completed';
          step.completedBy = approverId;
          step.completedAt = new Date().toISOString();
        });
      }

      await this.updateApprovalRequest(request);
      this.notifyApprovalListeners();

      return request;
    } catch (error) {
      console.error('Failed to approve request:', error);
      throw error;
    }
  }

  // Reject an approval request
  async rejectRequest(requestId: string, rejectedBy: string, reason: string): Promise<ApprovalRequest> {
    try {
      const request = await this.getApprovalRequest(requestId);
      if (!request) {
        throw new Error('Approval request not found');
      }

      request.status = 'rejected';
      request.rejectedBy = rejectedBy;
      request.rejectedAt = new Date().toISOString();
      request.rejectionReason = reason;

      await this.updateApprovalRequest(request);
      this.notifyApprovalListeners();

      return request;
    } catch (error) {
      console.error('Failed to reject request:', error);
      throw error;
    }
  }

  // Get all pending approval requests
  async getPendingApprovalRequests(): Promise<ApprovalRequest[]> {
    try {
      // In real implementation, would call API
      const allRequests = await this.getAllApprovalRequests();
      return allRequests.filter(request => request.status === 'pending');
    } catch (error) {
      console.error('Failed to get pending approval requests:', error);
      return [];
    }
  }

  // Get approval request by ID
  async getApprovalRequest(requestId: string): Promise<ApprovalRequest | null> {
    try {
      // Mock implementation - would call API
      const storedRequests = localStorage.getItem('approval_requests');
      if (!storedRequests) return null;

      const requests: ApprovalRequest[] = JSON.parse(storedRequests);
      return requests.find(request => request.id === requestId) || null;
    } catch (error) {
      console.error('Failed to get approval request:', error);
      return null;
    }
  }

  // Private helper methods
  private async getAllApprovalRequests(): Promise<ApprovalRequest[]> {
    try {
      const storedRequests = localStorage.getItem('approval_requests');
      return storedRequests ? JSON.parse(storedRequests) : [];
    } catch (error) {
      console.error('Failed to get all approval requests:', error);
      return [];
    }
  }

  private async storeApprovalRequest(request: ApprovalRequest): Promise<void> {
    try {
      const allRequests = await this.getAllApprovalRequests();
      allRequests.push(request);
      localStorage.setItem('approval_requests', JSON.stringify(allRequests));
    } catch (error) {
      console.error('Failed to store approval request:', error);
      throw error;
    }
  }

  private async updateApprovalRequest(request: ApprovalRequest): Promise<void> {
    try {
      const allRequests = await this.getAllApprovalRequests();
      const index = allRequests.findIndex(r => r.id === request.id);
      if (index >= 0) {
        allRequests[index] = request;
        localStorage.setItem('approval_requests', JSON.stringify(allRequests));
      }
    } catch (error) {
      console.error('Failed to update approval request:', error);
      throw error;
    }
  }

  private getCurrentUserId(): string {
    // In real implementation, would get from auth context
    return localStorage.getItem('user_id') || 'current-user';
  }

  private sendApprovalNotification(notification: any): void {
    // In real implementation, would send via WebSocket/email/Slack
    console.log('Sending approval notification:', notification);
  }

  private notifyApprovalListeners(): void {
    this.getAllApprovalRequests().then(requests => {
      this.approvalListeners.forEach(listener => {
        try {
          listener(requests);
        } catch (error) {
          console.error('Approval listener error:', error);
        }
      });
    });
  }

  // Subscribe to approval updates
  onApprovalUpdate(callback: (requests: ApprovalRequest[]) => void): () => void {
    this.approvalListeners.push(callback);

    // Return unsubscribe function
    return () => {
      const index = this.approvalListeners.indexOf(callback);
      if (index > -1) {
        this.approvalListeners.splice(index, 1);
      }
    };
  }

  // Check if user can approve requests
  canUserApprove(userId: string, safetyLevel: string): boolean {
    // In real implementation, would check user permissions
    const userRoles = this.getUserRoles(userId);

    switch (safetyLevel) {
      case 'dangerous':
        return userRoles.includes('senior-admin') || userRoles.includes('super-admin');
      case 'warning':
        return userRoles.includes('admin') || userRoles.includes('senior-admin') || userRoles.includes('super-admin');
      default:
        return true;
    }
  }

  private getUserRoles(userId: string): string[] {
    // Mock implementation - would get from auth service
    const storedRoles = localStorage.getItem(`user_roles_${userId}`);
    return storedRoles ? JSON.parse(storedRoles) : ['user'];
  }
}

export const commandApprovalService = new CommandApprovalService();
export default CommandApprovalService;
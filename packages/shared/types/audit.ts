export interface AuditLogEntry {
  id: string;
  timestamp: string;
  userId: string;
  userName: string;
  sessionId: string;
  clusterId?: string;
  clusterName?: string;
  action: string;
  resource: string;
  resourceId?: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  endpoint?: string;
  command?: string;
  result: 'success' | 'failure' | 'error';
  statusCode?: number;
  errorMessage?: string;
  ipAddress?: string;
  userAgent?: string;
  metadata?: Record<string, any>;
  duration?: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
}

export interface ComplianceStatus {
  id: string;
  standard: 'SOX' | 'HIPAA' | 'SOC2' | 'GDPR' | 'PCI-DSS' | 'ISO27001';
  status: 'compliant' | 'warning' | 'non_compliant' | 'unknown';
  lastAssessed: string;
  nextAssessment: string;
  score: number;
  requirements: ComplianceRequirement[];
  findings: ComplianceFinding[];
}

export interface ComplianceRequirement {
  id: string;
  standard: string;
  requirement: string;
  description: string;
  status: 'met' | 'partially_met' | 'not_met' | 'not_applicable';
  evidence?: string[];
  lastVerified?: string;
}

export interface ComplianceFinding {
  id: string;
  type: 'violation' | 'risk' | 'recommendation';
  severity: 'low' | 'medium' | 'high' | 'critical';
  title: string;
  description: string;
  requirement?: string;
  remediation?: string;
  status: 'open' | 'in_progress' | 'resolved' | 'accepted_risk';
  assignedTo?: string;
  dueDate?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuditFilter {
  startDate?: string;
  endDate?: string;
  userId?: string;
  clusterId?: string;
  action?: string;
  resource?: string;
  result?: 'success' | 'failure' | 'error';
  severity?: 'low' | 'medium' | 'high' | 'critical';
  searchQuery?: string;
}

export interface AuditSummary {
  totalEvents: number;
  successEvents: number;
  failureEvents: number;
  errorEvents: number;
  criticalEvents: number;
  topUsers: {
    userId: string;
    userName: string;
    eventCount: number;
  }[];
  topActions: {
    action: string;
    eventCount: number;
  }[];
  timeline: {
    date: string;
    eventCount: number;
    successRate: number;
  }[];
}
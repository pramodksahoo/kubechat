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

// Enhanced types for Story 1.8 implementation

export type ExportFormat = 'csv' | 'json' | 'pdf';

export type ComplianceFramework = 'sox' | 'hipaa' | 'soc2';

export interface AuditEvent {
  id: string;
  timestamp: string;
  userId?: string;
  eventType: string;
  severity: string;
  metadata: Record<string, any>;
}

export interface TamperAlert {
  id: string;
  detectedAt: string;
  affectedLogId: string;
  violationType: string;
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
}

export interface SuspiciousActivity {
  id: string;
  type: string;
  description: string;
  userId?: string;
  detectedAt: string;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  metadata: Record<string, any>;
}

export interface ChainIntegrityResult {
  isValid: boolean;
  totalChecked: number;
  violations?: string[];
  lastValidated: string;
  integrityScore: number;
}

export interface LegalHold {
  id: string;
  caseNumber: string;
  description: string;
  createdBy: string;
  createdAt: string;
  startTime: string;
  endTime?: string;
  status: 'active' | 'released' | 'expired';
  recordCount: number;
}

export interface LegalHoldRequest {
  caseNumber: string;
  description: string;
  startTime: string;
  endTime?: string;
}

export interface ComplianceReport {
  id: string;
  framework: ComplianceFramework;
  generatedAt: string;
  periodStart: string;
  periodEnd: string;
  complianceScore: number;
  totalEvents: number;
  violations: ComplianceViolation[];
  recommendations: string[];
  executiveSummary: string;
  detailedFindings: Record<string, any>;
}

export interface ComplianceViolation {
  id: string;
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  affectedLogIds: string[];
  detectedAt: string;
}

export interface SuspiciousActivity {
  id: string;
  detectedAt: string;
  activityType: string;
  userId?: string;
  description: string;
  riskScore: number;
  affectedRecords: string[];
  patternMatched: string;
}

export interface RetentionPolicy {
  id: string;
  name: string;
  retentionDays: number;
  appliesTo: string;
  createdAt: string;
  lastApplied?: string;
  automatic: boolean;
}

export interface ArchivalResult {
  archiveId: string;
  archivedCount: number;
  archiveSize: number;
  startDate: string;
  endDate: string;
  storageLocation: string;
  createdAt: string;
}

export interface AuditMetrics {
  totalLogsCreated: number;
  totalDangerousOps: number;
  totalFailedOps: number;
  averageResponseTime: number;
  successRate: number;
  integrityChecksPassed: number;
  integrityChecksFailed: number;
  lastIntegrityCheck?: string;
  queueSize: number;
  processedCount: number;
  errorCount: number;
  activeMonitoringSessions: number;
  tamperAlertsTriggered: number;
  activeLegalHolds: number;
  complianceScore: number;
}

export interface ExportRequest {
  format: ExportFormat;
  filter: AuditFilter;
}

export interface ComplianceReportRequest {
  framework: ComplianceFramework;
  startTime: string;
  endTime: string;
}

export interface RealTimeMonitoringStatus {
  isActive: boolean;
  connectedSessions: number;
  eventsPerSecond: number;
  lastEventAt?: string;
}
import { CommandPreview } from '@kubechat/shared/types';
import { Card } from '../ui/Card';

interface ImpactPreviewCardProps {
  preview: CommandPreview;
  expanded?: boolean;
  onToggleExpanded?: () => void;
  className?: string;
}

interface RiskAssessment {
  category: string;
  level: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  mitigations: string[];
}

export function ImpactPreviewCard({ 
  preview, 
  expanded = false, 
  onToggleExpanded,
  className = '' 
}: ImpactPreviewCardProps) {
  
  const getRiskAssessments = (): RiskAssessment[] => {
    // Mock implementation - would be computed by backend
    const assessments: RiskAssessment[] = [];

    // Analyze command for different risk categories
    const command = preview.generatedCommand.toLowerCase();

    // Data Loss Risk
    if (command.includes('delete') || command.includes('rm')) {
      assessments.push({
        category: 'Data Loss',
        level: preview.safetyLevel === 'dangerous' ? 'critical' : 'high',
        description: 'This command may permanently delete data or resources',
        mitigations: [
          'Verify backup availability before execution',
          'Consider using --dry-run flag first',
          'Ensure proper RBAC permissions are in place'
        ]
      });
    }

    // Service Disruption Risk
    if (command.includes('restart') || command.includes('rollout') || command.includes('scale')) {
      assessments.push({
        category: 'Service Disruption',
        level: preview.safetyLevel === 'dangerous' ? 'high' : 'medium',
        description: 'This command may cause temporary service unavailability',
        mitigations: [
          'Execute during maintenance window',
          'Verify replica count for zero-downtime deployment',
          'Monitor service health after execution'
        ]
      });
    }

    // Security Risk
    if (command.includes('exec') || command.includes('port-forward') || command.includes('proxy')) {
      assessments.push({
        category: 'Security',
        level: 'medium',
        description: 'This command may expose cluster resources or create security vulnerabilities',
        mitigations: [
          'Ensure network policies are properly configured',
          'Limit execution time and scope',
          'Audit access logs after execution'
        ]
      });
    }

    // Resource Impact
    if (command.includes('create') || command.includes('apply') || command.includes('scale')) {
      assessments.push({
        category: 'Resource Consumption',
        level: 'low',
        description: 'This command may affect cluster resource utilization',
        mitigations: [
          'Monitor cluster resource usage',
          'Verify resource quotas and limits',
          'Consider cluster capacity before scaling'
        ]
      });
    }

    // Compliance Risk
    if (preview.safetyLevel === 'dangerous') {
      assessments.push({
        category: 'Compliance',
        level: 'high',
        description: 'This operation requires additional oversight due to potential regulatory impact',
        mitigations: [
          'Document business justification',
          'Obtain necessary approvals',
          'Ensure audit trail is maintained'
        ]
      });
    }

    return assessments;
  };

  const getBlastRadiusInfo = () => {
    const command = preview.generatedCommand;
    
    return {
      scope: getCommandScope(command),
      affectedResources: getAffectedResources(command),
      recoveryTime: getEstimatedRecoveryTime(preview.safetyLevel),
      rollbackAvailable: hasRollbackCapability(command),
    };
  };

  const riskAssessments = getRiskAssessments();
  const blastRadius = getBlastRadiusInfo();

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'low': return 'text-green-600 dark:text-green-400';
      case 'medium': return 'text-yellow-600 dark:text-yellow-400';
      case 'high': return 'text-orange-600 dark:text-orange-400';
      case 'critical': return 'text-red-600 dark:text-red-400';
      default: return 'text-gray-600 dark:text-gray-400';
    }
  };

  const getLevelBadge = (level: string) => {
    const colors = {
      low: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
      high: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
      critical: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    };

    return (
      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${colors[level as keyof typeof colors] || colors.medium}`}>
        {level.toUpperCase()}
      </span>
    );
  };

  return (
    <Card className={`${className}`}>
      <div className="p-4">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
            Impact Assessment
          </h3>
          {onToggleExpanded && (
            <button
              onClick={onToggleExpanded}
              className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 text-sm font-medium"
            >
              {expanded ? 'Show Less' : 'Show More'}
            </button>
          )}
        </div>

        {/* Quick Summary */}
        <div className="grid grid-cols-2 gap-4 mb-6">
          <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
            <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Overall Risk
            </div>
            <div className={`font-semibold ${getLevelColor(preview.safetyLevel)}`}>
              {preview.safetyLevel.toUpperCase()}
            </div>
          </div>
          
          <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg">
            <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Approval Required
            </div>
            <div className={`font-semibold ${preview.approvalRequired ? 'text-orange-600 dark:text-orange-400' : 'text-green-600 dark:text-green-400'}`}>
              {preview.approvalRequired ? 'YES' : 'NO'}
            </div>
          </div>
        </div>

        {/* Risk Categories */}
        <div className="space-y-4 mb-6">
          <h4 className="font-medium text-gray-900 dark:text-white">Risk Categories</h4>
          <div className="space-y-3">
            {riskAssessments.map((risk, index) => (
              <div key={index} className="border border-gray-200 dark:border-gray-700 rounded-lg p-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-gray-900 dark:text-white">
                    {risk.category}
                  </span>
                  {getLevelBadge(risk.level)}
                </div>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                  {risk.description}
                </p>
                
                {expanded && (
                  <div>
                    <div className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Recommended Mitigations:
                    </div>
                    <ul className="text-xs text-gray-600 dark:text-gray-400 space-y-1">
                      {risk.mitigations.map((mitigation, idx) => (
                        <li key={idx} className="flex items-start space-x-2">
                          <span className="text-blue-500 mt-1 w-1 h-1 bg-current rounded-full flex-shrink-0"></span>
                          <span>{mitigation}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Blast Radius Analysis */}
        {expanded && (
          <div className="space-y-4">
            <h4 className="font-medium text-gray-900 dark:text-white">Blast Radius Analysis</h4>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-3">
                <div>
                  <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Scope
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">
                    {blastRadius.scope}
                  </div>
                </div>
                
                <div>
                  <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Estimated Recovery Time
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">
                    {blastRadius.recoveryTime}
                  </div>
                </div>
              </div>
              
              <div className="space-y-3">
                <div>
                  <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Affected Resources
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">
                    {blastRadius.affectedResources}
                  </div>
                </div>
                
                <div>
                  <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Rollback Available
                  </div>
                  <div className={`text-sm font-medium ${blastRadius.rollbackAvailable ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                    {blastRadius.rollbackAvailable ? 'YES' : 'NO'}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Confidence Score */}
        <div className="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Command Analysis Confidence
            </span>
            <div className="flex items-center space-x-2">
              <div className="w-20 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                <div 
                  className="bg-blue-600 h-2 rounded-full" 
                  style={{ width: `${preview.confidence * 100}%` }}
                ></div>
              </div>
              <span className="text-sm font-medium text-gray-900 dark:text-white">
                {Math.round(preview.confidence * 100)}%
              </span>
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}

// Helper functions
function getCommandScope(command: string): string {
  if (command.includes('--all-namespaces') || command.includes('-A')) {
    return 'Cluster-wide operation';
  } else if (command.includes('-n ') || command.includes('--namespace')) {
    return 'Single namespace operation';
  } else if (command.includes('node') || command.includes('pv')) {
    return 'Cluster resource operation';
  }
  return 'Default namespace operation';
}

function getAffectedResources(command: string): string {
  const resourceTypes = [];
  
  if (command.includes('pod')) resourceTypes.push('Pods');
  if (command.includes('deployment')) resourceTypes.push('Deployments');
  if (command.includes('service')) resourceTypes.push('Services');
  if (command.includes('configmap')) resourceTypes.push('ConfigMaps');
  if (command.includes('secret')) resourceTypes.push('Secrets');
  if (command.includes('node')) resourceTypes.push('Nodes');
  
  return resourceTypes.length > 0 ? resourceTypes.join(', ') : 'Various Kubernetes resources';
}

function getEstimatedRecoveryTime(safetyLevel: string): string {
  switch (safetyLevel) {
    case 'safe': return '< 5 minutes';
    case 'warning': return '5-30 minutes';
    case 'dangerous': return '30+ minutes or manual intervention required';
    default: return 'Unknown';
  }
}

function hasRollbackCapability(command: string): boolean {
  const rollbackCommands = ['rollout', 'apply', 'create', 'scale'];
  const destructiveCommands = ['delete', 'rm'];
  
  const hasRollback = rollbackCommands.some(cmd => command.includes(cmd));
  const hasDestructive = destructiveCommands.some(cmd => command.includes(cmd));
  
  return hasRollback && !hasDestructive;
}
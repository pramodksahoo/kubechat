interface ComponentTypeLike<P = unknown> { (props: P): any; }

export interface ClusterHealth {
  clusterId: string;
  clusterName: string;
  status: 'healthy' | 'warning' | 'critical' | 'unavailable';
  nodeCount: number;
  healthyNodes: number;
  unhealthyNodes: number;
  podCount: number;
  runningPods: number;
  failedPods: number;
  cpuUsage: number;
  memoryUsage: number;
  storageUsage: number;
  lastUpdated: string;
}

export interface ActivityItem {
  id: string;
  type: 'command' | 'deployment' | 'error' | 'warning' | 'info';
  title: string;
  description?: string;
  timestamp: string;
  userId?: string;
  userName?: string;
  clusterId?: string;
  clusterName?: string;
  severity?: 'low' | 'medium' | 'high' | 'critical';
  metadata?: Record<string, any>;
}

export interface QuickAccessPanel {
  id: string;
  title: string;
  description?: string;
  icon: ComponentTypeLike<any>;
  action: () => void;
  badge?: string | number;
  disabled?: boolean;
  permissions?: string[];
}

export interface PerformanceMetric {
  id: string;
  name: string;
  value: number;
  unit: string;
  trend: 'up' | 'down' | 'stable';
  change?: number;
  changePercent?: number;
  threshold?: {
    warning: number;
    critical: number;
  };
  historical?: {
    timestamp: string;
    value: number;
  }[];
}

export interface SystemHealth {
  api: {
    status: 'healthy' | 'degraded' | 'down';
    responseTime: number;
    uptime: number;
  };
  database: {
    status: 'healthy' | 'degraded' | 'down';
    connections: number;
    maxConnections: number;
  };
  websocket: {
    status: 'healthy' | 'degraded' | 'down';
    activeConnections: number;
  };
  llm: {
    status: 'healthy' | 'degraded' | 'down';
    provider: string;
    responseTime: number;
  };
  kubernetes: {
    status: 'healthy' | 'degraded' | 'down';
    clustersConnected: number;
    totalClusters: number;
  };
}
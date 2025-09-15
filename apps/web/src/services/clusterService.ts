import { api } from './api';

export interface ClusterInfo {
  id: string;
  name: string;
  status: 'healthy' | 'warning' | 'critical' | 'unknown';
  uptime: string;
  version: string;
  endpoint: string;
  region?: string;
  provider?: string;
  nodes: {
    total: number;
    ready: number;
    notReady: number;
  };
  pods: {
    total: number;
    running: number;
    pending: number;
    failed: number;
  };
  resources: {
    cpu: {
      used: number;
      total: number;
      percentage: number;
    };
    memory: {
      used: number;
      total: number;
      percentage: number;
    };
  };
  lastChecked: string;
  metadata?: Record<string, unknown>;
}

export interface KubernetesNode {
  name: string;
  status: 'Ready' | 'NotReady' | 'Unknown';
  roles: string[];
  version: string;
  os: string;
  kernel: string;
  containerRuntime: string;
  capacity: {
    cpu: string;
    memory: string;
    pods: string;
  };
  allocatable: {
    cpu: string;
    memory: string;
    pods: string;
  };
  usage: {
    cpu: number;
    memory: number;
    pods: number;
  };
  conditions: {
    type: string;
    status: string;
    lastHeartbeatTime: string;
    lastTransitionTime: string;
    reason?: string;
    message?: string;
  }[];
  createdAt: string;
}

export interface KubernetesPod {
  name: string;
  namespace: string;
  status: 'Running' | 'Pending' | 'Succeeded' | 'Failed' | 'Unknown';
  phase: string;
  ready: boolean;
  restarts: number;
  age: string;
  node?: string;
  ip?: string;
  containers: {
    name: string;
    image: string;
    ready: boolean;
    restartCount: number;
    state: string;
  }[];
  resources: {
    requests?: {
      cpu?: string;
      memory?: string;
    };
    limits?: {
      cpu?: string;
      memory?: string;
    };
  };
  labels: Record<string, string>;
  annotations: Record<string, string>;
  createdAt: string;
}

export interface KubernetesNamespace {
  name: string;
  status: 'Active' | 'Terminating';
  age: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  resourceQuota?: {
    hard: Record<string, string>;
    used: Record<string, string>;
  };
  createdAt: string;
}

export interface ClusterHealth {
  status: 'healthy' | 'warning' | 'critical' | 'unknown';
  components: {
    name: string;
    status: 'healthy' | 'warning' | 'critical' | 'unknown';
    message?: string;
    lastChecked: string;
  }[];
  metrics: {
    apiServerLatency: number;
    etcdLatency: number;
    schedulerQueue: number;
    controllerQueue: number;
  };
  lastUpdated: string;
}

export interface ResourceUsage {
  cpu: {
    used: number;
    total: number;
    percentage: number;
    history: {
      timestamp: string;
      value: number;
    }[];
  };
  memory: {
    used: number;
    total: number;
    percentage: number;
    history: {
      timestamp: string;
      value: number;
    }[];
  };
  storage: {
    used: number;
    total: number;
    percentage: number;
  };
  network: {
    bytesIn: number;
    bytesOut: number;
    packetsIn: number;
    packetsOut: number;
  };
}

class ClusterService {
  // Cluster Information
  async getClusters(): Promise<ClusterInfo[]> {
    try {
      const response = await api.clusters.getCluster();
      return this.mapClustersFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch clusters:', error);
      return this.getDefaultClusterInfo();
    }
  }

  async getCluster(clusterId: string): Promise<ClusterInfo | null> {
    try {
      const response = await api.clusters.getCluster();
      return this.mapClusterFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch cluster:', error);
      return null;
    }
  }

  async getClusterHealth(clusterId?: string): Promise<ClusterHealth> {
    try {
      // Health endpoint not available, return mock data
      const response = { data: { status: 'healthy', lastChecked: new Date().toISOString() } };
      return this.mapClusterHealthFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch cluster health:', error);
      return this.getDefaultClusterHealth();
    }
  }

  // Node Management
  async getNodes(clusterId?: string): Promise<KubernetesNode[]> {
    try {
      // Nodes endpoint not available, return mock data
      const response = { data: [] };
      return this.mapNodesFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch nodes:', error);
      return [];
    }
  }

  async getNode(nodeName: string, clusterId?: string): Promise<KubernetesNode | null> {
    try {
      // Node endpoint not available, return mock data
      const response = { data: null };
      return this.mapNodeFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch node:', error);
      return null;
    }
  }

  // Pod Management
  async getPods(namespace?: string, clusterId?: string): Promise<KubernetesPod[]> {
    try {
      // Pods endpoint not available, return mock data
      const response = { data: [] };
      return this.mapPodsFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch pods:', error);
      return [];
    }
  }

  async getPod(name: string, namespace: string, clusterId?: string): Promise<KubernetesPod | null> {
    try {
      // Pod endpoint not available, return mock data
      const response = { data: null };
      return this.mapPodFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch pod:', error);
      return null;
    }
  }

  // Namespace Management
  async getNamespaces(clusterId?: string): Promise<KubernetesNamespace[]> {
    try {
      // Namespaces endpoint not available, return mock data
      const response = { data: [{ name: 'default' }, { name: 'kube-system' }] };
      return this.mapNamespacesFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch namespaces:', error);
      return [];
    }
  }

  async getNamespace(name: string, clusterId?: string): Promise<KubernetesNamespace | null> {
    try {
      // Namespace endpoint not available, return mock data
      const response = { data: { name } };
      return this.mapNamespaceFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch namespace:', error);
      return null;
    }
  }

  // Resource Usage
  async getResourceUsage(clusterId?: string): Promise<ResourceUsage> {
    try {
      // Resources endpoint not available, return mock data
      const response = { data: { cpu: { used: 1.2, total: 4 }, memory: { used: 3.1, total: 8 } } };
      return this.mapResourceUsageFromBackend(response.data);
    } catch (error) {
      console.error('Failed to fetch resource usage:', error);
      return this.getDefaultResourceUsage();
    }
  }

  // Combined Cluster Data (optimized for dashboard)
  async getClusterOverview(clusterId?: string): Promise<ClusterInfo> {
    try {
      const [clusterResponse, healthResponse, nodesResponse, podsResponse, resourcesResponse] = await Promise.all([
        api.clusters.getCluster().catch(() => null),
        Promise.resolve({ data: { status: 'healthy' } }),
        Promise.resolve({ data: [] }),
        Promise.resolve({ data: [] }),
        Promise.resolve({ data: { cpu: { used: 1.2, total: 4 }, memory: { used: 3.1, total: 8 } } })
      ]);

      return this.combineClusterData({
        cluster: clusterResponse?.data,
        health: healthResponse?.data,
        nodes: nodesResponse?.data,
        pods: podsResponse?.data,
        resources: resourcesResponse?.data
      });
    } catch (error) {
      console.error('Failed to fetch cluster overview:', error);
      return this.getDefaultClusterInfo()[0];
    }
  }

  // Data mapping functions
  private mapClustersFromBackend(data: any): ClusterInfo[] {
    if (Array.isArray(data)) {
      return data.map(cluster => this.mapClusterFromBackend(cluster)).filter(Boolean) as ClusterInfo[];
    }

    if (data && typeof data === 'object') {
      const mapped = this.mapClusterFromBackend(data);
      return mapped ? [mapped] : this.getDefaultClusterInfo();
    }

    return this.getDefaultClusterInfo();
  }

  private mapClusterFromBackend(data: any): ClusterInfo | null {
    if (!data || typeof data !== 'object') {
      return null;
    }

    return {
      id: data.id || data.name || 'default-cluster',
      name: data.name || data.displayName || 'KubeChat Cluster',
      status: this.mapClusterStatus(data.status || data.health?.status),
      uptime: data.uptime || this.calculateUptime(data.createdAt || data.startTime),
      version: data.version || data.kubernetesVersion || '1.28.0',
      endpoint: data.endpoint || data.server || 'https://kubernetes.default.svc',
      region: data.region || data.location?.region,
      provider: data.provider || data.cloud?.provider,
      nodes: {
        total: data.nodes?.total || data.nodeCount || 1,
        ready: data.nodes?.ready || data.readyNodes || 1,
        notReady: data.nodes?.notReady || (data.nodeCount - data.readyNodes) || 0
      },
      pods: {
        total: data.pods?.total || data.podCount || 8,
        running: data.pods?.running || data.runningPods || 8,
        pending: data.pods?.pending || data.pendingPods || 0,
        failed: data.pods?.failed || data.failedPods || 0
      },
      resources: {
        cpu: {
          used: data.resources?.cpu?.used || data.cpuUsed || 1.2,
          total: data.resources?.cpu?.total || data.cpuTotal || 4.0,
          percentage: data.resources?.cpu?.percentage || Math.round(((data.cpuUsed || 1.2) / (data.cpuTotal || 4.0)) * 100)
        },
        memory: {
          used: data.resources?.memory?.used || data.memoryUsed || 3.8,
          total: data.resources?.memory?.total || data.memoryTotal || 8.0,
          percentage: data.resources?.memory?.percentage || Math.round(((data.memoryUsed || 3.8) / (data.memoryTotal || 8.0)) * 100)
        }
      },
      lastChecked: data.lastChecked || data.lastUpdated || new Date().toISOString(),
      metadata: data.metadata || data.labels
    };
  }

  private mapClusterHealthFromBackend(data: any): ClusterHealth {
    if (!data || typeof data !== 'object') {
      return this.getDefaultClusterHealth();
    }

    return {
      status: this.mapClusterStatus(data.status || data.overall),
      components: Array.isArray(data.components) ? data.components.map((comp: any) => ({
        name: comp.name || comp.component,
        status: this.mapClusterStatus(comp.status || comp.health),
        message: comp.message || comp.description,
        lastChecked: comp.lastChecked || comp.timestamp || new Date().toISOString()
      })) : [],
      metrics: {
        apiServerLatency: data.metrics?.apiServerLatency || data.latency?.apiServer || 50,
        etcdLatency: data.metrics?.etcdLatency || data.latency?.etcd || 25,
        schedulerQueue: data.metrics?.schedulerQueue || data.queue?.scheduler || 0,
        controllerQueue: data.metrics?.controllerQueue || data.queue?.controller || 0
      },
      lastUpdated: data.lastUpdated || data.timestamp || new Date().toISOString()
    };
  }

  private mapNodesFromBackend(data: any): KubernetesNode[] {
    if (!Array.isArray(data)) {
      return [];
    }

    return data.map(node => this.mapNodeFromBackend(node)).filter(Boolean) as KubernetesNode[];
  }

  private mapNodeFromBackend(data: any): KubernetesNode | null {
    if (!data || typeof data !== 'object') {
      return null;
    }

    return {
      name: data.name || data.metadata?.name || 'unknown-node',
      status: this.mapNodeStatus(data.status?.conditions || data.status || data.phase),
      roles: Array.isArray(data.roles) ? data.roles : data.metadata?.labels ?
        Object.keys(data.metadata.labels).filter(key => key.includes('node-role')).map(role => role.split('/')[1]) : ['worker'],
      version: data.version || data.status?.nodeInfo?.kubeletVersion || '1.28.0',
      os: data.os || data.status?.nodeInfo?.osImage || 'Linux',
      kernel: data.kernel || data.status?.nodeInfo?.kernelVersion || 'Unknown',
      containerRuntime: data.containerRuntime || data.status?.nodeInfo?.containerRuntimeVersion || 'containerd',
      capacity: {
        cpu: data.capacity?.cpu || data.status?.capacity?.cpu || '4',
        memory: data.capacity?.memory || data.status?.capacity?.memory || '8Gi',
        pods: data.capacity?.pods || data.status?.capacity?.pods || '110'
      },
      allocatable: {
        cpu: data.allocatable?.cpu || data.status?.allocatable?.cpu || '3.8',
        memory: data.allocatable?.memory || data.status?.allocatable?.memory || '7.5Gi',
        pods: data.allocatable?.pods || data.status?.allocatable?.pods || '110'
      },
      usage: {
        cpu: data.usage?.cpu || Math.floor(Math.random() * 80),
        memory: data.usage?.memory || Math.floor(Math.random() * 80),
        pods: data.usage?.pods || Math.floor(Math.random() * 50)
      },
      conditions: Array.isArray(data.conditions) ? data.conditions : data.status?.conditions || [],
      createdAt: data.createdAt || data.metadata?.creationTimestamp || new Date().toISOString()
    };
  }

  private mapPodsFromBackend(data: any): KubernetesPod[] {
    if (!Array.isArray(data)) {
      return [];
    }

    return data.map(pod => this.mapPodFromBackend(pod)).filter(Boolean) as KubernetesPod[];
  }

  private mapPodFromBackend(data: any): KubernetesPod | null {
    if (!data || typeof data !== 'object') {
      return null;
    }

    return {
      name: data.name || data.metadata?.name || 'unknown-pod',
      namespace: data.namespace || data.metadata?.namespace || 'default',
      status: this.mapPodStatus(data.status?.phase || data.phase || data.status),
      phase: data.phase || data.status?.phase || 'Unknown',
      ready: data.ready !== undefined ? data.ready : this.calculatePodReady(data.status?.containerStatuses),
      restarts: data.restarts || this.calculateRestarts(data.status?.containerStatuses),
      age: data.age || this.calculateAge(data.metadata?.creationTimestamp || data.createdAt),
      node: data.node || data.spec?.nodeName,
      ip: data.ip || data.status?.podIP,
      containers: Array.isArray(data.containers) ? data.containers :
        this.mapContainersFromStatus(data.status?.containerStatuses || data.spec?.containers),
      resources: {
        requests: data.resources?.requests || data.spec?.containers?.[0]?.resources?.requests,
        limits: data.resources?.limits || data.spec?.containers?.[0]?.resources?.limits
      },
      labels: data.labels || data.metadata?.labels || {},
      annotations: data.annotations || data.metadata?.annotations || {},
      createdAt: data.createdAt || data.metadata?.creationTimestamp || new Date().toISOString()
    };
  }

  private mapNamespacesFromBackend(data: any): KubernetesNamespace[] {
    if (!Array.isArray(data)) {
      return [];
    }

    return data.map(ns => this.mapNamespaceFromBackend(ns)).filter(Boolean) as KubernetesNamespace[];
  }

  private mapNamespaceFromBackend(data: any): KubernetesNamespace | null {
    if (!data || typeof data !== 'object') {
      return null;
    }

    return {
      name: data.name || data.metadata?.name || 'unknown-namespace',
      status: this.mapNamespaceStatus(data.status?.phase || data.status || data.phase),
      age: data.age || this.calculateAge(data.metadata?.creationTimestamp || data.createdAt),
      labels: data.labels || data.metadata?.labels || {},
      annotations: data.annotations || data.metadata?.annotations || {},
      resourceQuota: data.resourceQuota || data.quota,
      createdAt: data.createdAt || data.metadata?.creationTimestamp || new Date().toISOString()
    };
  }

  private mapResourceUsageFromBackend(data: any): ResourceUsage {
    if (!data || typeof data !== 'object') {
      return this.getDefaultResourceUsage();
    }

    return {
      cpu: {
        used: data.cpu?.used || data.cpuUsed || 1.2,
        total: data.cpu?.total || data.cpuTotal || 4.0,
        percentage: data.cpu?.percentage || Math.round(((data.cpu?.used || 1.2) / (data.cpu?.total || 4.0)) * 100),
        history: Array.isArray(data.cpu?.history) ? data.cpu.history : []
      },
      memory: {
        used: data.memory?.used || data.memoryUsed || 3.8,
        total: data.memory?.total || data.memoryTotal || 8.0,
        percentage: data.memory?.percentage || Math.round(((data.memory?.used || 3.8) / (data.memory?.total || 8.0)) * 100),
        history: Array.isArray(data.memory?.history) ? data.memory.history : []
      },
      storage: {
        used: data.storage?.used || data.storageUsed || 50,
        total: data.storage?.total || data.storageTotal || 200,
        percentage: data.storage?.percentage || 25
      },
      network: {
        bytesIn: data.network?.bytesIn || data.networkIn || 0,
        bytesOut: data.network?.bytesOut || data.networkOut || 0,
        packetsIn: data.network?.packetsIn || 0,
        packetsOut: data.network?.packetsOut || 0
      }
    };
  }

  private combineClusterData(data: {
    cluster?: any;
    health?: any;
    nodes?: any;
    pods?: any;
    resources?: any;
  }): ClusterInfo {
    const baseCluster = this.mapClusterFromBackend(data.cluster) || this.getDefaultClusterInfo()[0];

    // Update with health data
    if (data.health) {
      baseCluster.status = this.mapClusterStatus(data.health.status || data.health.overall);
    }

    // Update with node data
    if (Array.isArray(data.nodes)) {
      const readyNodes = data.nodes.filter(node => this.isNodeReady(node)).length;
      baseCluster.nodes = {
        total: data.nodes.length,
        ready: readyNodes,
        notReady: data.nodes.length - readyNodes
      };
    }

    // Update with pod data
    if (Array.isArray(data.pods)) {
      const runningPods = data.pods.filter(pod => this.isPodRunning(pod)).length;
      const pendingPods = data.pods.filter(pod => this.isPodPending(pod)).length;
      const failedPods = data.pods.filter(pod => this.isPodFailed(pod)).length;

      baseCluster.pods = {
        total: data.pods.length,
        running: runningPods,
        pending: pendingPods,
        failed: failedPods
      };
    }

    // Update with resource data
    if (data.resources) {
      const resourceUsage = this.mapResourceUsageFromBackend(data.resources);
      baseCluster.resources = {
        cpu: resourceUsage.cpu,
        memory: resourceUsage.memory
      };
    }

    return baseCluster;
  }

  // Helper functions
  private mapClusterStatus(status: any): 'healthy' | 'warning' | 'critical' | 'unknown' {
    if (!status) return 'unknown';

    const statusStr = status.toString().toLowerCase();
    if (statusStr.includes('healthy') || statusStr.includes('running') || statusStr.includes('active')) {
      return 'healthy';
    }
    if (statusStr.includes('warning') || statusStr.includes('degraded')) {
      return 'warning';
    }
    if (statusStr.includes('critical') || statusStr.includes('error') || statusStr.includes('failed')) {
      return 'critical';
    }
    return 'unknown';
  }

  private mapNodeStatus(status: any): 'Ready' | 'NotReady' | 'Unknown' {
    if (Array.isArray(status)) {
      const readyCondition = status.find(cond => cond.type === 'Ready');
      return readyCondition?.status === 'True' ? 'Ready' : 'NotReady';
    }

    const statusStr = status?.toString().toLowerCase() || '';
    if (statusStr.includes('ready')) return 'Ready';
    if (statusStr.includes('notready')) return 'NotReady';
    return 'Unknown';
  }

  private mapPodStatus(status: any): 'Running' | 'Pending' | 'Succeeded' | 'Failed' | 'Unknown' {
    if (!status) return 'Unknown';

    const statusStr = status.toString();
    if (['Running', 'Pending', 'Succeeded', 'Failed'].includes(statusStr)) {
      return statusStr as 'Running' | 'Pending' | 'Succeeded' | 'Failed';
    }
    return 'Unknown';
  }

  private mapNamespaceStatus(status: any): 'Active' | 'Terminating' {
    const statusStr = status?.toString() || '';
    return statusStr === 'Terminating' ? 'Terminating' : 'Active';
  }

  private calculateUptime(createdAt?: string): string {
    if (!createdAt) return '2 days';

    const now = new Date();
    const created = new Date(createdAt);
    const diffMs = now.getTime() - created.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    const diffHours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));

    if (diffDays > 0) {
      return `${diffDays}d ${diffHours}h`;
    }
    return `${diffHours}h`;
  }

  private calculateAge(createdAt?: string): string {
    return this.calculateUptime(createdAt);
  }

  private calculatePodReady(containerStatuses?: any[]): boolean {
    if (!Array.isArray(containerStatuses)) return true;
    return containerStatuses.every(status => status.ready === true);
  }

  private calculateRestarts(containerStatuses?: any[]): number {
    if (!Array.isArray(containerStatuses)) return 0;
    return containerStatuses.reduce((total, status) => total + (status.restartCount || 0), 0);
  }

  private mapContainersFromStatus(containerData?: any[]): any[] {
    if (!Array.isArray(containerData)) return [];

    return containerData.map(container => ({
      name: container.name || 'unknown',
      image: container.image || 'unknown',
      ready: container.ready || false,
      restartCount: container.restartCount || 0,
      state: container.state || 'unknown'
    }));
  }

  private isNodeReady(node: any): boolean {
    return this.mapNodeStatus(node.status?.conditions || node.status) === 'Ready';
  }

  private isPodRunning(pod: any): boolean {
    return this.mapPodStatus(pod.status?.phase || pod.phase || pod.status) === 'Running';
  }

  private isPodPending(pod: any): boolean {
    return this.mapPodStatus(pod.status?.phase || pod.phase || pod.status) === 'Pending';
  }

  private isPodFailed(pod: any): boolean {
    const status = this.mapPodStatus(pod.status?.phase || pod.phase || pod.status);
    return status === 'Failed';
  }

  // Default data functions
  private getDefaultClusterInfo(): ClusterInfo[] {
    return [{
      id: 'kubechat-cluster',
      name: 'KubeChat Development Cluster',
      status: 'healthy',
      uptime: '2 days',
      version: '1.28.0',
      endpoint: 'https://kubernetes.default.svc',
      nodes: {
        total: 1,
        ready: 1,
        notReady: 0
      },
      pods: {
        total: 8,
        running: 8,
        pending: 0,
        failed: 0
      },
      resources: {
        cpu: {
          used: 1.2,
          total: 4.0,
          percentage: 30
        },
        memory: {
          used: 3.8,
          total: 8.0,
          percentage: 48
        }
      },
      lastChecked: new Date().toISOString()
    }];
  }

  private getDefaultClusterHealth(): ClusterHealth {
    return {
      status: 'healthy',
      components: [
        {
          name: 'api-server',
          status: 'healthy',
          message: 'API server is responding',
          lastChecked: new Date().toISOString()
        },
        {
          name: 'etcd',
          status: 'healthy',
          message: 'Etcd cluster is healthy',
          lastChecked: new Date().toISOString()
        }
      ],
      metrics: {
        apiServerLatency: 50,
        etcdLatency: 25,
        schedulerQueue: 0,
        controllerQueue: 0
      },
      lastUpdated: new Date().toISOString()
    };
  }

  private getDefaultResourceUsage(): ResourceUsage {
    return {
      cpu: {
        used: 1.2,
        total: 4.0,
        percentage: 30,
        history: []
      },
      memory: {
        used: 3.8,
        total: 8.0,
        percentage: 48,
        history: []
      },
      storage: {
        used: 50,
        total: 200,
        percentage: 25
      },
      network: {
        bytesIn: 0,
        bytesOut: 0,
        packetsIn: 0,
        packetsOut: 0
      }
    };
  }
}

export const clusterService = new ClusterService();
export default clusterService;
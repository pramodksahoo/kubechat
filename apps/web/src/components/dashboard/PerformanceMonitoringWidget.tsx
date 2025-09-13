import React, { useEffect, useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Card } from '@/components/ui/Card';

interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
  'data-testid'?: string;
}

interface MetricData {
  timestamp: string;
  value: number;
}

interface PerformanceMetric {
  id: string;
  name: string;
  unit: string;
  current: number;
  average: number;
  threshold: {
    warning: number;
    critical: number;
  };
  trend: 'up' | 'down' | 'stable';
  data: MetricData[];
}

interface ResourceUsage {
  cpu: PerformanceMetric;
  memory: PerformanceMetric;
  network: PerformanceMetric;
  storage: PerformanceMetric;
}

export interface PerformanceMonitoringWidgetProps extends BaseComponentProps {
  metrics?: ResourceUsage;
  timeRange?: '1h' | '6h' | '24h' | '7d';
  autoRefresh?: boolean;
  refreshInterval?: number;
  onTimeRangeChange?: (range: string) => void;
  onMetricClick?: (metricId: string) => void;
}

export const PerformanceMonitoringWidget: React.FC<PerformanceMonitoringWidgetProps> = ({
  metrics,
  timeRange = '1h',
  autoRefresh = true,
  refreshInterval = 30000,
  onTimeRangeChange,
  onMetricClick,
  className = '',
  'data-testid': dataTestId = 'performance-monitoring-widget'
}) => {
  const [isLoading, setIsLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(new Date().toISOString());

  // Generate sample data points for demonstration
  const generateSampleData = (baseValue: number, variance: number = 10): MetricData[] => {
    const data: MetricData[] = [];
    const now = new Date();
    const points = timeRange === '1h' ? 12 : timeRange === '6h' ? 36 : timeRange === '24h' ? 96 : 168;
    const intervalMs = timeRange === '1h' ? 5 * 60 * 1000 : timeRange === '6h' ? 10 * 60 * 1000 : timeRange === '24h' ? 15 * 60 * 1000 : 60 * 60 * 1000;
    
    for (let i = points - 1; i >= 0; i--) {
      data.push({
        timestamp: new Date(now.getTime() - (i * intervalMs)).toISOString(),
        value: baseValue + (Math.random() - 0.5) * variance
      });
    }
    return data;
  };

  const defaultMetrics: ResourceUsage = {
    cpu: {
      id: 'cpu',
      name: 'CPU Usage',
      unit: '%',
      current: 34.5,
      average: 28.2,
      threshold: { warning: 70, critical: 90 },
      trend: 'up',
      data: generateSampleData(30, 15)
    },
    memory: {
      id: 'memory',
      name: 'Memory Usage',
      unit: '%',
      current: 68.3,
      average: 65.1,
      threshold: { warning: 80, critical: 95 },
      trend: 'stable',
      data: generateSampleData(65, 10)
    },
    network: {
      id: 'network',
      name: 'Network I/O',
      unit: 'MB/s',
      current: 12.4,
      average: 8.9,
      threshold: { warning: 50, critical: 80 },
      trend: 'up',
      data: generateSampleData(10, 8)
    },
    storage: {
      id: 'storage',
      name: 'Storage Usage',
      unit: '%',
      current: 42.1,
      average: 41.8,
      threshold: { warning: 85, critical: 95 },
      trend: 'stable',
      data: generateSampleData(42, 3)
    }
  };

  const displayMetrics = metrics || defaultMetrics;

  // Auto-refresh effect
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      setIsLoading(true);
      // Simulate data refresh
      setTimeout(() => {
        setLastUpdated(new Date().toISOString());
        setIsLoading(false);
      }, 500);
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval]);

  const getStatusColor = (current: number, threshold: { warning: number; critical: number }) => {
    if (current >= threshold.critical) return 'text-danger-500';
    if (current >= threshold.warning) return 'text-warning-500';
    return 'text-success-500';
  };

  const getStatusBg = (current: number, threshold: { warning: number; critical: number }) => {
    if (current >= threshold.critical) return 'bg-danger-50 dark:bg-danger-900/20 border-danger-200 dark:border-danger-800';
    if (current >= threshold.warning) return 'bg-warning-50 dark:bg-warning-900/20 border-warning-200 dark:border-warning-800';
    return 'bg-success-50 dark:bg-success-900/20 border-success-200 dark:border-success-800';
  };

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return { icon: 'arrow-trending-up', color: 'text-danger-500' };
      case 'down': return { icon: 'arrow-trending-down', color: 'text-success-500' };
      default: return { icon: 'minus', color: 'text-gray-500' };
    }
  };

  const formatValue = (value: number, unit: string) => {
    return `${value.toFixed(1)}${unit}`;
  };

  const timeRangeOptions = [
    { value: '1h', label: '1 Hour' },
    { value: '6h', label: '6 Hours' },
    { value: '24h', label: '24 Hours' },
    { value: '7d', label: '7 Days' }
  ];

  // Simple sparkline SVG generator
  const generateSparkline = (data: MetricData[], width = 120, height = 40) => {
    if (data.length < 2) return '';
    
    const minValue = Math.min(...data.map(d => d.value));
    const maxValue = Math.max(...data.map(d => d.value));
    const range = maxValue - minValue || 1;
    
    const points = data.map((d, i) => {
      const x = (i / (data.length - 1)) * width;
      const y = height - ((d.value - minValue) / range) * height;
      return `${x},${y}`;
    }).join(' ');
    
    return (
      <svg width={width} height={height} className="text-primary-500">
        <polyline
          points={points}
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          className="opacity-70"
        />
      </svg>
    );
  };

  return (
    <Card className={className} data-testid={dataTestId}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center">
            <Icon name="chart-bar" className="h-5 w-5 text-primary-500 mr-2" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Performance Monitoring
            </h3>
          </div>
          
          <div className="flex items-center space-x-3">
            {/* Time range selector */}
            <select
              value={timeRange}
              onChange={(e) => onTimeRangeChange?.(e.target.value)}
              className="text-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1 bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              data-testid="time-range-selector"
            >
              {timeRangeOptions.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            
            {/* Refresh indicator */}
            {isLoading && (
              <Icon name="spinner" className="h-4 w-4 animate-spin text-primary-500" />
            )}
          </div>
        </div>

        {/* Metrics grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {Object.values(displayMetrics).map((metric) => {
            const trendConfig = getTrendIcon(metric.trend);
            const statusColor = getStatusColor(metric.current, metric.threshold);
            const statusBg = getStatusBg(metric.current, metric.threshold);
            
            return (
              <div
                key={metric.id}
                className={`p-4 rounded-lg border cursor-pointer transition-all duration-200 hover:shadow-md ${statusBg}`}
                onClick={() => onMetricClick?.(metric.id)}
                data-testid={`metric-${metric.id}`}
              >
                {/* Metric header */}
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center">
                    <h4 className="text-sm font-medium text-gray-900 dark:text-white">
                      {metric.name}
                    </h4>
                    <Icon 
                      name={trendConfig.icon} 
                      className={`h-4 w-4 ml-2 ${trendConfig.color}`} 
                    />
                  </div>
                  <div className="text-right">
                    <p className={`text-lg font-semibold ${statusColor}`}>
                      {formatValue(metric.current, metric.unit)}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      Avg: {formatValue(metric.average, metric.unit)}
                    </p>
                  </div>
                </div>

                {/* Sparkline chart */}
                <div className="flex items-end justify-between">
                  <div className="flex-1">
                    {generateSparkline(metric.data)}
                  </div>
                </div>

                {/* Threshold indicators */}
                <div className="mt-3 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                  <span>Thresholds:</span>
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center">
                      <div className="h-2 w-2 bg-warning-500 rounded-full mr-1" />
                      <span>{metric.threshold.warning}{metric.unit}</span>
                    </div>
                    <div className="flex items-center">
                      <div className="h-2 w-2 bg-danger-500 rounded-full mr-1" />
                      <span>{metric.threshold.critical}{metric.unit}</span>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Footer */}
        <div className="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
            <div className="flex items-center">
              <Icon name="clock" className="h-3 w-3 mr-1" />
              <span>Last updated: {new Date(lastUpdated).toLocaleTimeString()}</span>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center">
                <div className="h-2 w-2 bg-success-500 rounded-full mr-1" />
                <span>Normal</span>
              </div>
              <div className="flex items-center">
                <div className="h-2 w-2 bg-warning-500 rounded-full mr-1" />
                <span>Warning</span>
              </div>
              <div className="flex items-center">
                <div className="h-2 w-2 bg-danger-500 rounded-full mr-1" />
                <span>Critical</span>
              </div>
            </div>
          </div>
        </div>

        {/* Performance tips */}
        <div className="mt-4 p-3 bg-info-50 dark:bg-info-900/20 rounded-lg border border-info-200 dark:border-info-800">
          <div className="flex items-start">
            <Icon name="light-bulb" className="h-4 w-4 text-info-600 dark:text-info-400 mt-0.5 mr-2 flex-shrink-0" />
            <div>
              <p className="text-xs font-medium text-info-900 dark:text-info-100 mb-1">
                Performance Optimization
              </p>
              <p className="text-xs text-info-700 dark:text-info-200">
                Monitor resource usage patterns to optimize cluster performance. Consider scaling when CPU or memory consistently exceeds 70%.
              </p>
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
};

export default PerformanceMonitoringWidget;
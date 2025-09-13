import React, { useEffect, useState } from 'react';
import { Icon } from '@/components/ui';

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

  const getStatusConfig = (current: number, threshold: { warning: number; critical: number }) => {
    if (current >= threshold.critical) {
      return {
        color: 'text-red-600 dark:text-red-400',
        bgGradient: 'bg-gradient-to-br from-red-50 to-rose-100 dark:from-red-900/20 dark:to-rose-900/30',
        borderColor: 'border-red-200/50 dark:border-red-700/50',
        dotColor: 'bg-gradient-to-r from-red-500 to-rose-500',
        glowColor: 'shadow-red-500/20',
        chartColor: '#ef4444',
        level: 'critical'
      };
    }
    if (current >= threshold.warning) {
      return {
        color: 'text-amber-600 dark:text-amber-400',
        bgGradient: 'bg-gradient-to-br from-amber-50 to-orange-100 dark:from-amber-900/20 dark:to-orange-900/30',
        borderColor: 'border-amber-200/50 dark:border-amber-700/50',
        dotColor: 'bg-gradient-to-r from-amber-500 to-orange-500',
        glowColor: 'shadow-amber-500/20',
        chartColor: '#f59e0b',
        level: 'warning'
      };
    }
    return {
      color: 'text-emerald-600 dark:text-emerald-400',
      bgGradient: 'bg-gradient-to-br from-emerald-50 to-green-100 dark:from-emerald-900/20 dark:to-green-900/30',
      borderColor: 'border-emerald-200/50 dark:border-emerald-700/50',
      dotColor: 'bg-gradient-to-r from-emerald-500 to-green-500',
      glowColor: 'shadow-emerald-500/20',
      chartColor: '#10b981',
      level: 'normal'
    };
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

  // Enhanced interactive sparkline component
  const InteractiveSparkline: React.FC<{
    data: MetricData[];
    color: string;
    width?: number;
    height?: number;
    animated?: boolean;
  }> = ({ data, color, width = 140, height = 60, animated = true }) => {
    const [hoveredPoint, setHoveredPoint] = useState<number | null>(null);
    const [animatedData, setAnimatedData] = useState<MetricData[]>([]);

    React.useEffect(() => {
      if (animated) {
        setAnimatedData([]);
        const timer = setTimeout(() => {
          setAnimatedData(data);
        }, 200);
        return () => clearTimeout(timer);
      } else {
        setAnimatedData(data);
      }
    }, [data, animated]);

    if (animatedData.length < 2) return null;

    const minValue = Math.min(...animatedData.map(d => d.value));
    const maxValue = Math.max(...animatedData.map(d => d.value));
    const range = maxValue - minValue || 1;

    const points = animatedData.map((d, i) => {
      const x = (i / (animatedData.length - 1)) * width;
      const y = height - ((d.value - minValue) / range) * height;
      return { x, y, value: d.value, timestamp: d.timestamp };
    });

    const pathData = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x},${p.y}`).join(' ');
    const areaData = `${pathData} L ${width},${height} L 0,${height} Z`;

    return (
      <div className="relative group">
        <svg
          width={width}
          height={height}
          className="transition-all duration-300"
          onMouseLeave={() => setHoveredPoint(null)}
        >
          {/* Gradient definition */}
          <defs>
            <linearGradient id={`gradient-${color.replace('#', '')}`} x1="0%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" stopColor={color} stopOpacity="0.3" />
              <stop offset="100%" stopColor={color} stopOpacity="0.05" />
            </linearGradient>
          </defs>

          {/* Area fill */}
          <path
            d={areaData}
            fill={`url(#gradient-${color.replace('#', '')})`}
            className="transition-all duration-1000 ease-out"
            style={{
              transform: animated ? 'scaleX(1)' : 'scaleX(0)',
              transformOrigin: 'left'
            }}
          />

          {/* Main line */}
          <path
            d={pathData}
            fill="none"
            stroke={color}
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="transition-all duration-1000 ease-out filter drop-shadow-sm"
            style={{
              strokeDasharray: animated ? 'none' : `${width * 2}`,
              strokeDashoffset: animated ? '0' : `${width * 2}`
            }}
          />

          {/* Data points */}
          {points.map((point, i) => (
            <circle
              key={i}
              cx={point.x}
              cy={point.y}
              r={hoveredPoint === i ? 4 : 2}
              fill={color}
              className="transition-all duration-200 cursor-pointer opacity-0 group-hover:opacity-100"
              onMouseEnter={() => setHoveredPoint(i)}
              onClick={() => console.log('Point clicked:', point)}
            />
          ))}
        </svg>

        {/* Tooltip */}
        {hoveredPoint !== null && (
          <div className="absolute z-10 px-2 py-1 text-xs font-medium text-white bg-gray-900 dark:bg-gray-700 rounded-md shadow-lg transform -translate-x-1/2 -translate-y-full">
            <div className="text-center">
              <div className="font-semibold">{points[hoveredPoint].value.toFixed(1)}</div>
              <div className="text-gray-300 text-xs">
                {new Date(points[hoveredPoint].timestamp).toLocaleTimeString()}
              </div>
            </div>
            <div className="absolute top-full left-1/2 transform -translate-x-1/2 w-2 h-2 bg-gray-900 dark:bg-gray-700 rotate-45" />
          </div>
        )}
      </div>
    );
  };

  return (
    <div className={`relative overflow-hidden rounded-2xl bg-white/70 dark:bg-gray-900/70 backdrop-blur-xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl ${className}`} data-testid={dataTestId}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-xl bg-gradient-to-br from-purple-100 to-indigo-200 dark:from-purple-800/50 dark:to-indigo-700/50">
              <Icon name="chart-bar" className="h-5 w-5 text-purple-600 dark:text-purple-400" />
            </div>
            <div>
              <h3 className="text-xl font-bold bg-gradient-to-r from-gray-900 via-gray-700 to-gray-900 dark:from-white dark:via-gray-200 dark:to-white bg-clip-text text-transparent">
                Performance Monitoring
              </h3>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
                Real-time resource utilization across clusters
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            {/* Time range selector */}
            <div className="flex items-center space-x-1 p-1 rounded-xl bg-gray-100/50 dark:bg-gray-800/50 backdrop-blur-sm">
              {timeRangeOptions.map(option => (
                <button
                  key={option.value}
                  onClick={() => onTimeRangeChange?.(option.value)}
                  className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-all duration-200 ${
                    timeRange === option.value
                      ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
                  }`}
                  data-testid={`time-range-${option.value}`}
                >
                  {option.label}
                </button>
              ))}
            </div>

            {/* Refresh indicator */}
            {isLoading && (
              <div className="flex items-center px-3 py-2 rounded-xl bg-blue-100/50 dark:bg-blue-900/20 backdrop-blur-sm border border-blue-200/50 dark:border-blue-700/50">
                <Icon name="spinner" className="h-4 w-4 animate-spin text-blue-600" />
                <span className="text-xs font-medium text-blue-600 dark:text-blue-400 ml-2">Updating</span>
              </div>
            )}
          </div>
        </div>

        {/* Metrics grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {Object.values(displayMetrics).map((metric, index) => {
            const trendConfig = getTrendIcon(metric.trend);
            const statusConfig = getStatusConfig(metric.current, metric.threshold);

            return (
              <div
                key={metric.id}
                className={`group relative overflow-hidden rounded-2xl p-6 ${statusConfig.bgGradient} border ${statusConfig.borderColor}
                  backdrop-blur-sm cursor-pointer transition-all duration-300 ease-out
                  hover:scale-[1.02] hover:shadow-xl hover:${statusConfig.glowColor}
                  transform hover:-translate-y-1`}
                onClick={() => onMetricClick?.(metric.id)}
                data-testid={`metric-${metric.id}`}
                style={{
                  animationDelay: `${index * 150}ms`,
                  animation: 'fadeInUp 0.6s ease-out forwards'
                }}
              >
                {/* Background decoration */}
                <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                <div className="relative z-10">
                  {/* Metric header */}
                  <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center space-x-3">
                      <div className="relative">
                        <div className={`h-3 w-3 rounded-full ${statusConfig.dotColor} shadow-lg animate-pulse`} />
                        <div className={`absolute inset-0 h-3 w-3 rounded-full ${statusConfig.dotColor} opacity-20 animate-pulse scale-150`} />
                      </div>
                      <div>
                        <h4 className="text-sm font-semibold text-gray-900 dark:text-white uppercase tracking-wider">
                          {metric.name}
                        </h4>
                        <div className="flex items-center space-x-1 mt-1">
                          <Icon
                            name={trendConfig.icon}
                            className={`h-3 w-3 ${trendConfig.color}`}
                          />
                          <span className={`text-xs font-medium ${trendConfig.color} capitalize`}>
                            {metric.trend}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className={`text-2xl font-bold ${statusConfig.color}`}>
                        {formatValue(metric.current, metric.unit)}
                      </p>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                        Avg: {formatValue(metric.average, metric.unit)}
                      </p>
                    </div>
                  </div>

                  {/* Interactive chart */}
                  <div className="mb-6">
                    <InteractiveSparkline
                      data={metric.data}
                      color={statusConfig.chartColor}
                      width={260}
                      height={80}
                    />
                  </div>

                  {/* Progress bar and thresholds */}
                  <div className="space-y-3">
                    <div className="relative">
                      <div className="w-full bg-gray-200/50 dark:bg-gray-700/50 rounded-full h-2 overflow-hidden">
                        <div
                          className={`h-full transition-all duration-1000 ease-out ${statusConfig.dotColor} shadow-sm`}
                          style={{ width: `${Math.min(metric.current, 100)}%` }}
                        />
                      </div>
                      <div className="flex justify-between items-center mt-2">
                        <span className="text-xs text-gray-500 dark:text-gray-400">0{metric.unit}</span>
                        <span className="text-xs font-medium text-gray-700 dark:text-gray-300">
                          Current: {formatValue(metric.current, metric.unit)}
                        </span>
                        <span className="text-xs text-gray-500 dark:text-gray-400">100{metric.unit}</span>
                      </div>
                    </div>

                    {/* Threshold indicators */}
                    <div className="flex items-center justify-between pt-3 border-t border-gray-200/50 dark:border-gray-700/50">
                      <span className="text-xs font-medium text-gray-600 dark:text-gray-400">Thresholds</span>
                      <div className="flex items-center space-x-3">
                        <div className="flex items-center space-x-1">
                          <div className="h-2 w-2 bg-gradient-to-r from-amber-500 to-orange-500 rounded-full shadow-sm" />
                          <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{metric.threshold.warning}{metric.unit}</span>
                        </div>
                        <div className="flex items-center space-x-1">
                          <div className="h-2 w-2 bg-gradient-to-r from-red-500 to-rose-500 rounded-full shadow-sm" />
                          <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{metric.threshold.critical}{metric.unit}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Hover shine effect */}
                <div className="absolute inset-0 -top-10 -left-10 bg-gradient-to-r from-transparent via-white/10 to-transparent
                  transform skew-x-12 opacity-0 group-hover:opacity-100 transition-all duration-700 group-hover:translate-x-full" />
              </div>
            );
          })}
        </div>

        {/* Footer */}
        <div className="mt-8 pt-6 border-t border-gray-200/50 dark:border-gray-700/50">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between space-y-3 sm:space-y-0">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                <Icon name="clock" className="h-3 w-3 text-gray-500" />
                <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
                  Updated {new Date(lastUpdated).toLocaleTimeString()}
                </span>
              </div>
              <div className="flex items-center space-x-1">
                <div className="h-1.5 w-1.5 bg-emerald-500 rounded-full animate-pulse" />
                <span className="text-xs font-medium text-emerald-600 dark:text-emerald-400">Live monitoring</span>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              {[
                { level: 'Normal', color: 'bg-gradient-to-r from-emerald-500 to-green-500', count: Object.values(displayMetrics).filter(m => getStatusConfig(m.current, m.threshold).level === 'normal').length },
                { level: 'Warning', color: 'bg-gradient-to-r from-amber-500 to-orange-500', count: Object.values(displayMetrics).filter(m => getStatusConfig(m.current, m.threshold).level === 'warning').length },
                { level: 'Critical', color: 'bg-gradient-to-r from-red-500 to-rose-500', count: Object.values(displayMetrics).filter(m => getStatusConfig(m.current, m.threshold).level === 'critical').length }
              ].map(({ level, color, count }) => (
                <div key={level} className="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm">
                  <div className={`h-2.5 w-2.5 rounded-full ${color} shadow-sm`} />
                  <span className="text-xs font-medium text-gray-700 dark:text-gray-300">
                    {level}: {count}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Performance insights */}
        <div className="mt-6 p-6 rounded-2xl bg-gradient-to-br from-blue-50/50 to-indigo-100/50 dark:from-blue-900/20 dark:to-indigo-900/30 backdrop-blur-sm border border-blue-200/50 dark:border-blue-700/50">
          <div className="flex items-start space-x-3">
            <div className="p-2 rounded-xl bg-gradient-to-br from-blue-100 to-indigo-200 dark:from-blue-800/50 dark:to-indigo-700/50">
              <Icon name="light-bulb" className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            </div>
            <div className="flex-1">
              <p className="text-sm font-semibold text-blue-900 dark:text-blue-100 mb-2">
                Performance Insights
              </p>
              <p className="text-sm text-blue-700 dark:text-blue-200 leading-relaxed">
                Monitor resource usage patterns to optimize cluster performance. Consider scaling when CPU or memory consistently exceeds 70%.
                <span className="font-medium">Click on any metric</span> for detailed analysis and recommendations.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Add keyframes for animations */}
      <style jsx>{`
        @keyframes fadeInUp {
          0% {
            opacity: 0;
            transform: translateY(30px);
          }
          100% {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
};

export default PerformanceMonitoringWidget;
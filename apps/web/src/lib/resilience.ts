// Resilience utilities for KubeChat Enterprise
// Implementing retry mechanisms, circuit breakers, and fault tolerance patterns

import { ScreenReaderUtils } from '../design-system/accessibility';

// Extended error type for API errors
export interface ApiError extends Error {
  code?: string;
  status?: number;
}

// Helper function to convert unknown errors to ApiError
function toApiError(error: unknown): ApiError {
  if (error instanceof Error) {
    return error as ApiError;
  }
  return new Error(String(error)) as ApiError;
}

// Retry configuration interface
export interface RetryConfig {
  maxAttempts: number;
  baseDelay: number;
  maxDelay: number;
  backoffMultiplier: number;
  jitter: boolean;
  retryCondition?: (error: ApiError) => boolean;
  onRetry?: (attempt: number, error: ApiError) => void;
  onFailure?: (error: ApiError, attempts: number) => void;
}

// Default retry configuration
const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxAttempts: 3,
  baseDelay: 1000,
  maxDelay: 10000,
  backoffMultiplier: 2,
  jitter: true,
  retryCondition: (error) => {
    // Retry on network errors, 5xx errors, and timeouts
    return (
      error?.code === 'NETWORK_ERROR' ||
      error?.code === 'TIMEOUT' ||
      (error?.status !== undefined && error.status >= 500 && error.status < 600) ||
      error?.name === 'NetworkError' ||
      error?.name === 'TimeoutError'
    );
  },
};

// Exponential backoff with jitter
function calculateDelay(attempt: number, config: RetryConfig): number {
  const exponentialDelay = Math.min(
    config.baseDelay * Math.pow(config.backoffMultiplier, attempt - 1),
    config.maxDelay
  );

  if (config.jitter) {
    // Add random jitter to prevent thundering herd
    return exponentialDelay + Math.random() * 1000;
  }

  return exponentialDelay;
}

// Retry mechanism with exponential backoff
export async function withRetry<T>(
  operation: () => Promise<T>,
  config: Partial<RetryConfig> = {}
): Promise<T> {
  const finalConfig = { ...DEFAULT_RETRY_CONFIG, ...config };
  let lastError: ApiError | undefined;

  for (let attempt = 1; attempt <= finalConfig.maxAttempts; attempt++) {
    try {
      return await operation();
    } catch (error) {
      lastError = toApiError(error);

      // Check if we should retry this error
      if (!finalConfig.retryCondition?.(lastError)) {
        throw error;
      }

      // Don't delay on the last attempt
      if (attempt === finalConfig.maxAttempts) {
        break;
      }

      // Calculate delay and wait
      const delay = calculateDelay(attempt, finalConfig);
      
      // Call retry callback
      finalConfig.onRetry?.(attempt, lastError);

      // Wait before next attempt
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }

  // All attempts failed
  const finalError = lastError || new Error('Operation failed after retries') as ApiError;
  finalConfig.onFailure?.(finalError, finalConfig.maxAttempts);
  throw finalError;
}

// Circuit breaker states
export enum CircuitBreakerState {
  CLOSED = 'closed',
  OPEN = 'open',
  HALF_OPEN = 'half-open',
}

// Circuit breaker configuration
export interface CircuitBreakerConfig {
  failureThreshold: number;
  recoveryTimeout: number;
  monitoringPeriod: number;
  halfOpenMaxCalls: number;
  onStateChange?: (state: CircuitBreakerState, error?: unknown) => void;
  isFailure?: (error: unknown) => boolean;
}

// Default circuit breaker configuration
const DEFAULT_CIRCUIT_BREAKER_CONFIG: CircuitBreakerConfig = {
  failureThreshold: 5,
  recoveryTimeout: 60000, // 1 minute
  monitoringPeriod: 10000, // 10 seconds
  halfOpenMaxCalls: 3,
  isFailure: (error) => {
    // Consider 5xx errors and network errors as failures
    const err = error as any;
    return (
      err?.status >= 500 ||
      err?.code === 'NETWORK_ERROR' ||
      err?.code === 'TIMEOUT'
    );
  },
};

// Circuit breaker implementation
export class CircuitBreaker<T> {
  private state: CircuitBreakerState = CircuitBreakerState.CLOSED;
  private failureCount = 0;
  private lastFailureTime = 0;
  private halfOpenCalls = 0;
  private config: CircuitBreakerConfig;
  private monitoringInterval: NodeJS.Timeout | null = null;

  constructor(
    private operation: () => Promise<T>,
    config: Partial<CircuitBreakerConfig> = {}
  ) {
    this.config = { ...DEFAULT_CIRCUIT_BREAKER_CONFIG, ...config };
    this.startMonitoring();
  }

  async execute(): Promise<T> {
    if (this.state === CircuitBreakerState.OPEN) {
      if (Date.now() - this.lastFailureTime >= this.config.recoveryTimeout) {
        this.setState(CircuitBreakerState.HALF_OPEN);
      } else {
        throw new Error('Circuit breaker is OPEN. Service is temporarily unavailable.');
      }
    }

    if (this.state === CircuitBreakerState.HALF_OPEN) {
      if (this.halfOpenCalls >= this.config.halfOpenMaxCalls) {
        throw new Error('Circuit breaker is HALF_OPEN. Maximum test calls exceeded.');
      }
      this.halfOpenCalls++;
    }

    try {
      const result = await this.operation();
      
      // Success - reset failure count
      if (this.state === CircuitBreakerState.HALF_OPEN) {
        this.setState(CircuitBreakerState.CLOSED);
        this.halfOpenCalls = 0;
      }
      
      this.failureCount = 0;
      return result;
    } catch (error) {
      this.handleFailure(error);
      throw error;
    }
  }

  private handleFailure(error: unknown) {
    if (!this.config.isFailure?.(error)) {
      return; // Not considered a failure
    }

    this.failureCount++;
    this.lastFailureTime = Date.now();

    if (this.state === CircuitBreakerState.HALF_OPEN) {
      this.setState(CircuitBreakerState.OPEN);
      this.halfOpenCalls = 0;
    } else if (
      this.state === CircuitBreakerState.CLOSED &&
      this.failureCount >= this.config.failureThreshold
    ) {
      this.setState(CircuitBreakerState.OPEN);
    }
  }

  private setState(newState: CircuitBreakerState) {
    if (this.state !== newState) {
      const previousState = this.state;
      this.state = newState;
      
      console.log(`Circuit breaker state changed: ${previousState} -> ${newState}`);
      this.config.onStateChange?.(newState);

      // Announce state changes to screen readers for critical services
      if (newState === CircuitBreakerState.OPEN) {
        ScreenReaderUtils.announce(
          'Service temporarily unavailable due to errors. Please try again later.',
          'assertive'
        );
      } else if (newState === CircuitBreakerState.CLOSED && previousState === CircuitBreakerState.OPEN) {
        ScreenReaderUtils.announce('Service has been restored and is now available.');
      }
    }
  }

  private startMonitoring() {
    this.monitoringInterval = setInterval(() => {
      // Reset failure count periodically in closed state
      if (this.state === CircuitBreakerState.CLOSED && this.failureCount > 0) {
        this.failureCount = Math.max(0, this.failureCount - 1);
      }
    }, this.config.monitoringPeriod);
  }

  getState(): CircuitBreakerState {
    return this.state;
  }

  getStats() {
    return {
      state: this.state,
      failureCount: this.failureCount,
      lastFailureTime: this.lastFailureTime,
      halfOpenCalls: this.halfOpenCalls,
    };
  }

  destroy() {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
    }
  }
}

// Timeout wrapper
export function withTimeout<T>(
  operation: () => Promise<T>,
  timeoutMs: number,
  timeoutMessage = 'Operation timed out'
): Promise<T> {
  return Promise.race([
    operation(),
    new Promise<never>((_, reject) => {
      setTimeout(() => {
        const error = new Error(timeoutMessage);
        (error as any).code = 'TIMEOUT';
        reject(error);
      }, timeoutMs);
    }),
  ]);
}

// Bulkhead pattern - resource isolation
interface QueueItem<T = unknown> {
  operation: () => Promise<T>;
  resolve: (value: T) => void;
  reject: (error: unknown) => void;
}

export class Bulkhead {
  private activeOperations = 0;
  private queue: QueueItem[] = [];

  constructor(
    private maxConcurrency: number,
    private maxQueueSize: number = 100
  ) {}

  async execute<T>(operation: () => Promise<T>): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      if (this.queue.length >= this.maxQueueSize) {
        reject(new Error('Bulkhead queue is full. Request rejected.'));
        return;
      }

      this.queue.push({
        operation: operation as () => Promise<unknown>,
        resolve: resolve as (value: unknown) => void,
        reject
      });
      this.processQueue();
    });
  }

  private async processQueue() {
    if (this.activeOperations >= this.maxConcurrency || this.queue.length === 0) {
      return;
    }

    const item = this.queue.shift();
    if (!item) return;

    this.activeOperations++;

    try {
      const result = await item.operation();
      item.resolve(result);
    } catch (error) {
      item.reject(error);
    } finally {
      this.activeOperations--;
      // Process next item in queue
      setTimeout(() => this.processQueue(), 0);
    }
  }

  getStats() {
    return {
      activeOperations: this.activeOperations,
      queueLength: this.queue.length,
      maxConcurrency: this.maxConcurrency,
      maxQueueSize: this.maxQueueSize,
    };
  }
}

// Rate limiter using token bucket algorithm
export class RateLimiter {
  private tokens: number;
  private lastRefill: number;

  constructor(
    private maxTokens: number,
    private refillRate: number // tokens per second
  ) {
    this.tokens = maxTokens;
    this.lastRefill = Date.now();
  }

  async acquire(tokens = 1): Promise<boolean> {
    this.refillTokens();

    if (this.tokens >= tokens) {
      this.tokens -= tokens;
      return true;
    }

    return false;
  }

  async waitForTokens(tokens = 1): Promise<void> {
    while (!(await this.acquire(tokens))) {
      // Wait and try again
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  }

  private refillTokens() {
    const now = Date.now();
    const timePassed = (now - this.lastRefill) / 1000;
    const tokensToAdd = timePassed * this.refillRate;

    this.tokens = Math.min(this.maxTokens, this.tokens + tokensToAdd);
    this.lastRefill = now;
  }

  getStats() {
    return {
      tokens: this.tokens,
      maxTokens: this.maxTokens,
      refillRate: this.refillRate,
    };
  }
}

// Cache with TTL and fallback
export class ResilienceCache<T> {
  private cache = new Map<string, { value: T; expiry: number; stale: boolean }>();

  constructor(
    private defaultTTL: number = 300000, // 5 minutes
    private staleWhileRevalidate: number = 600000 // 10 minutes
  ) {}

  async get(
    key: string,
    fetcher: () => Promise<T>,
    ttl: number = this.defaultTTL
  ): Promise<T> {
    const cached = this.cache.get(key);
    const now = Date.now();

    // Cache hit and not expired
    if (cached && now < cached.expiry) {
      return cached.value;
    }

    // Stale data available - return it while fetching fresh data
    if (cached && now < cached.expiry + this.staleWhileRevalidate) {
      // Background refresh
      this.refreshInBackground(key, fetcher, ttl);
      return cached.value;
    }

    // No cache or expired - fetch fresh data
    try {
      const value = await fetcher();
      this.cache.set(key, {
        value,
        expiry: now + ttl,
        stale: false,
      });
      return value;
    } catch (error) {
      // If fetch fails and we have stale data, return it
      if (cached) {
        console.warn(`Failed to refresh cache for key ${key}, returning stale data:`, error);
        return cached.value;
      }
      throw error;
    }
  }

  private async refreshInBackground(
    key: string,
    fetcher: () => Promise<T>,
    ttl: number
  ) {
    try {
      const value = await fetcher();
      this.cache.set(key, {
        value,
        expiry: Date.now() + ttl,
        stale: false,
      });
    } catch (error) {
      console.warn(`Background refresh failed for key ${key}:`, error);
      // Mark existing entry as stale
      const existing = this.cache.get(key);
      if (existing) {
        existing.stale = true;
      }
    }
  }

  delete(key: string) {
    this.cache.delete(key);
  }

  clear() {
    this.cache.clear();
  }

  getStats() {
    return {
      size: this.cache.size,
      keys: Array.from(this.cache.keys()),
    };
  }
}

// Comprehensive resilience wrapper
export class ResilienceWrapper<T> {
  private circuitBreaker: CircuitBreaker<T>;
  private bulkhead?: Bulkhead;
  private rateLimiter?: RateLimiter;
  private cache?: ResilienceCache<T>;

  constructor(
    private operation: () => Promise<T>,
    private config: {
      retry?: Partial<RetryConfig>;
      circuitBreaker?: Partial<CircuitBreakerConfig>;
      timeout?: number;
      bulkhead?: { maxConcurrency: number; maxQueueSize?: number };
      rateLimiter?: { maxTokens: number; refillRate: number };
      cache?: { ttl?: number; staleWhileRevalidate?: number };
    } = {}
  ) {
    this.circuitBreaker = new CircuitBreaker(operation, config.circuitBreaker);
    
    if (config.bulkhead) {
      this.bulkhead = new Bulkhead(
        config.bulkhead.maxConcurrency,
        config.bulkhead.maxQueueSize
      );
    }

    if (config.rateLimiter) {
      this.rateLimiter = new RateLimiter(
        config.rateLimiter.maxTokens,
        config.rateLimiter.refillRate
      );
    }

    if (config.cache) {
      this.cache = new ResilienceCache<T>(
        config.cache.ttl,
        config.cache.staleWhileRevalidate
      );
    }
  }

  async execute(cacheKey?: string): Promise<T> {
    // Rate limiting
    if (this.rateLimiter && !(await this.rateLimiter.acquire())) {
      throw new Error('Rate limit exceeded. Please try again later.');
    }

    // Cache lookup
    if (this.cache && cacheKey) {
      return this.cache.get(cacheKey, () => this.executeOperation());
    }

    return this.executeOperation();
  }

  private async executeOperation(): Promise<T> {
    const operation = async () => {
      if (this.bulkhead) {
        return this.bulkhead.execute(() => this.circuitBreaker.execute());
      }
      return this.circuitBreaker.execute();
    };

    if (this.config.timeout) {
      return withTimeout(
        () => withRetry(operation, this.config.retry),
        this.config.timeout
      );
    }

    return withRetry(operation, this.config.retry);
  }

  getStats() {
    return {
      circuitBreaker: this.circuitBreaker.getStats(),
      bulkhead: this.bulkhead?.getStats(),
      rateLimiter: this.rateLimiter?.getStats(),
      cache: this.cache?.getStats(),
    };
  }

  destroy() {
    this.circuitBreaker.destroy();
  }
}

// Convenience function for circuit breaker pattern
export function withCircuitBreaker<T>(
  operation: () => Promise<T>,
  name: string,
  config?: Partial<CircuitBreakerConfig>
): () => Promise<T> {
  const circuitBreaker = new CircuitBreaker(operation, config);
  return () => circuitBreaker.execute();
}
// KubeChat Shared Types
// TypeScript type definitions shared across frontend and backend

import { z } from 'zod';

// Export UI types
export * from './ui';
export * from './dashboard';
export * from './chat';
export * from './auth';
export * from './audit';

// API Response Types
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// User Management Types
export const UserSchema = z.object({
  id: z.string().uuid(),
  username: z.string().min(3).max(50),
  email: z.string().email(),
  role: z.enum(['admin', 'user', 'viewer']),
  createdAt: z.date(),
  updatedAt: z.date(),
});

export type User = z.infer<typeof UserSchema>;

// Kubernetes Resource Types
export const KubernetesResourceSchema = z.object({
  apiVersion: z.string(),
  kind: z.string(),
  metadata: z.object({
    name: z.string(),
    namespace: z.string().optional(),
    labels: z.record(z.string()).optional(),
    annotations: z.record(z.string()).optional(),
  }),
});

export type KubernetesResource = z.infer<typeof KubernetesResourceSchema>;

// Chat/Query Types  
export const ChatMessageSchema = z.object({
  id: z.string().uuid(),
  userId: z.string().uuid(),
  content: z.string().min(1),
  type: z.enum(['user', 'assistant', 'system']),
  timestamp: z.date(),
  metadata: z.record(z.unknown()).optional(),
});

export type ChatMessage = z.infer<typeof ChatMessageSchema>;

// Query Result Types
export const QueryResultSchema = z.object({
  id: z.string().uuid(),
  query: z.string(),
  result: z.unknown(),
  status: z.enum(['success', 'error', 'pending']),
  executedAt: z.date(),
  executionTime: z.number(),
});

export type QueryResult = z.infer<typeof QueryResultSchema>;

// Configuration Types
export interface DatabaseConfig {
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  ssl?: boolean;
}

export interface RedisConfig {
  host: string;
  port: number;
  password?: string;
  db?: number;
}

export interface AppConfig {
  environment: 'development' | 'staging' | 'production';
  port: number;
  database: DatabaseConfig;
  redis: RedisConfig;
  jwtSecret: string;
  corsOrigins: string[];
}

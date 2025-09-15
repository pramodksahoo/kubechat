// KubeChat Environment Configuration
import { z } from 'zod';

// Environment schema validation
export const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'staging', 'production']).default('development'),
  PORT: z.coerce.number().default(3000),
  API_BASE_URL: z.string().url().optional(),
  DATABASE_URL: z.string().optional(),
  REDIS_URL: z.string().optional(),
  JWT_SECRET: z.string().min(32).optional(),
});

export type Environment = z.infer<typeof envSchema>;

// Parse and validate environment variables
export function parseEnvironment(env: Record<string, string | undefined> = process.env): Environment {
  const result = envSchema.safeParse(env);
  
  if (!result.success) {
    throw new Error(`Invalid environment configuration: ${result.error.message}`);
  }
  
  return result.data;
}

// Get validated environment
export const env = parseEnvironment();

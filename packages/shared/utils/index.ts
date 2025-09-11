// KubeChat Shared Utilities
// Common utility functions used across frontend and backend

// API Response Helpers
export function createSuccessResponse<T>(data: T, message?: string) {
  return {
    success: true,
    data,
    message,
  };
}

export function createErrorResponse(error: string, message?: string) {
  return {
    success: false,
    error,
    message,
  };
}

// Date Utilities
export function formatTimestamp(date: Date): string {
  return date.toISOString();
}

export function parseTimestamp(timestamp: string): Date {
  return new Date(timestamp);
}

// Validation Utilities
export function isValidUUID(uuid: string): boolean {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
  return uuidRegex.test(uuid);
}

export function sanitizeString(input: string): string {
  return input.trim().replace(/[<>]/g, '');
}

// Kubernetes Utilities
export function formatResourceName(name: string): string {
  return name.toLowerCase().replace(/[^a-z0-9-]/g, '-');
}

export function parseKubernetesNamespace(resource: string): { namespace?: string; name: string } {
  const parts = resource.split('/');
  if (parts.length === 2) {
    return { namespace: parts[0], name: parts[1] };
  }
  return { name: resource };
}

// Error Handling Utilities
export class KubeChatError extends Error {
  constructor(
    message: string,
    public code: string,
    public statusCode: number = 500
  ) {
    super(message);
    this.name = 'KubeChatError';
  }
}

export function handleAsyncError<T>(
  promise: Promise<T>
): Promise<[T | null, Error | null]> {
  return promise
    .then<[T, null]>((data: T) => [data, null])
    .catch<[null, Error]>((error: Error) => [null, error]);
}

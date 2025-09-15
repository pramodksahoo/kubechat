# KubeChat API Documentation

## Overview

The KubeChat API provides programmatic access to KubeChat's natural language Kubernetes management capabilities. This RESTful API enables integration with external tools, custom dashboards, and automated workflows while maintaining enterprise-grade security and audit compliance.

## Base URL

```
https://your-kubechat-instance.com/api/v1
```

## Authentication

KubeChat API uses JWT (JSON Web Tokens) for authentication with support for multiple authentication methods.

### Authentication Methods

1. **Username/Password Authentication**
2. **External Identity Provider (OIDC)**
3. **Service Account Tokens**
4. **API Keys** (for programmatic access)

### Obtaining an Access Token

**Username/Password:**
```bash
curl -X POST https://your-kubechat-instance.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your-username",
    "password": "your-password"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### Using Authentication

Include the JWT token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
     https://your-kubechat-instance.com/api/v1/queries
```

## Rate Limiting

API requests are rate-limited to prevent abuse and ensure service availability:

- **Authenticated users**: 1000 requests per hour
- **Unauthenticated requests**: 100 requests per hour
- **Query processing**: 60 queries per minute per user

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Error Handling

The API uses standard HTTP status codes and returns detailed error information:

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Query cannot be empty",
    "details": {
      "field": "query",
      "value": ""
    },
    "request_id": "req_1234567890"
  }
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `AUTHENTICATION_ERROR` | Invalid or expired token |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `VALIDATION_ERROR` | Invalid request parameters |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `QUERY_PROCESSING_ERROR` | AI processing failed |
| `KUBERNETES_ERROR` | Kubernetes operation failed |
| `INTERNAL_ERROR` | Server error |

## Security Headers

All API responses include security headers:

```
Content-Security-Policy: default-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

## CORS Configuration

CORS is configured for secure cross-origin requests:

```javascript
// Allowed origins (configure for your domain)
const allowedOrigins = [
  'https://your-domain.com',
  'http://localhost:3000' // Development only
];

// Allowed methods
const allowedMethods = ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'];

// Allowed headers
const allowedHeaders = ['Authorization', 'Content-Type', 'X-Requested-With'];
```

## WebSocket API

Real-time updates are available via WebSocket connection:

```javascript
const ws = new WebSocket('wss://your-kubechat-instance.com/api/v1/ws');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Real-time update:', data);
};
```

## API Endpoints

### Core Endpoints

| Endpoint | Description |
|----------|-------------|
| [`POST /queries`](endpoints/queries.md) | Process natural language queries |
| [`GET /queries/{id}`](endpoints/queries.md#get-query) | Get query details |
| [`POST /commands/execute`](endpoints/commands.md) | Execute kubectl commands |
| [`GET /audit`](endpoints/audit.md) | Retrieve audit logs |
| [`GET /clusters`](endpoints/clusters.md) | List available clusters |
| [`GET /health`](endpoints/health.md) | Service health check |

### Authentication Endpoints

| Endpoint | Description |
|----------|-------------|
| [`POST /auth/login`](endpoints/auth.md) | User authentication |
| [`POST /auth/refresh`](endpoints/auth.md#refresh) | Refresh access token |
| [`POST /auth/logout`](endpoints/auth.md#logout) | Invalidate session |
| [`GET /auth/me`](endpoints/auth.md#profile) | Get user profile |

### Administration Endpoints

| Endpoint | Description |
|----------|-------------|
| [`GET /admin/users`](endpoints/admin.md) | List users |
| [`POST /admin/users`](endpoints/admin.md#create-user) | Create user |
| [`GET /admin/settings`](endpoints/admin.md#settings) | Get system settings |
| [`PUT /admin/settings`](endpoints/admin.md#update-settings) | Update settings |

## SDKs and Client Libraries

### Official SDKs

- **Go SDK**: [kubechat-go](https://github.com/pramodksahoo/kubechat-go)
- **Python SDK**: [kubechat-python](https://github.com/pramodksahoo/kubechat-python)
- **JavaScript/TypeScript SDK**: [kubechat-js](https://github.com/pramodksahoo/kubechat-js)

### Community SDKs

- **Java SDK**: [kubechat-java](https://github.com/community/kubechat-java)
- **C# SDK**: [kubechat-dotnet](https://github.com/community/kubechat-dotnet)

### Quick Start with Go SDK

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/pramodksahoo/kubechat-go"
)

func main() {
    client := kubechat.NewClient(&kubechat.Config{
        BaseURL: "https://your-kubechat-instance.com",
        Token:   "your-jwt-token",
    })

    result, err := client.ProcessQuery(context.Background(), &kubechat.QueryRequest{
        Query: "show me all pods in the default namespace",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Generated command: %s\n", result.Command)
    fmt.Printf("Safety level: %s\n", result.SafetyLevel)
}
```

## Integration Examples

### Custom Dashboard Integration

```javascript
// Example: Integrating KubeChat into a custom dashboard
import KubeChatClient from '@kubechat/client';

const client = new KubeChatClient({
  baseURL: 'https://your-kubechat-instance.com',
  token: localStorage.getItem('kubechat_token')
});

async function queryCluster(naturalLanguageQuery) {
  try {
    const result = await client.processQuery({
      query: naturalLanguageQuery
    });

    return {
      command: result.command,
      safetyLevel: result.safety_level,
      explanation: result.explanation
    };
  } catch (error) {
    console.error('Query failed:', error);
    throw error;
  }
}
```

### CI/CD Pipeline Integration

```yaml
# Example: GitHub Actions workflow using KubeChat API
name: Deploy with KubeChat
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Authenticate with KubeChat
        run: |
          TOKEN=$(curl -X POST ${{ secrets.KUBECHAT_URL }}/api/v1/auth/login \
            -H "Content-Type: application/json" \
            -d '{"username": "${{ secrets.KUBECHAT_USER }}", "password": "${{ secrets.KUBECHAT_PASS }}"}' \
            | jq -r '.access_token')
          echo "KUBECHAT_TOKEN=$TOKEN" >> $GITHUB_ENV

      - name: Deploy application
        run: |
          curl -X POST ${{ secrets.KUBECHAT_URL }}/api/v1/queries \
            -H "Authorization: Bearer $KUBECHAT_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{"query": "deploy my-app to production namespace with 3 replicas"}'
```

## Webhooks

KubeChat supports webhooks for real-time notifications:

### Webhook Configuration

```json
{
  "url": "https://your-app.com/webhooks/kubechat",
  "events": ["query.processed", "command.executed", "security.alert"],
  "secret": "your-webhook-secret",
  "active": true
}
```

### Webhook Payload Example

```json
{
  "event": "query.processed",
  "timestamp": "2025-01-15T10:30:00Z",
  "data": {
    "query_id": "query_123",
    "user_id": "user_456",
    "query": "show me pod status",
    "command": "kubectl get pods",
    "safety_level": "safe"
  },
  "signature": "sha256=..."
}
```

## OpenAPI Specification

The complete API specification is available in OpenAPI 3.0 format:

- [Download OpenAPI Spec](openapi.yaml)
- [Interactive API Explorer](https://your-kubechat-instance.com/api/docs)

## Best Practices

### Security Best Practices

1. **Use HTTPS**: Always use HTTPS in production
2. **Rotate Tokens**: Regularly rotate JWT tokens and API keys
3. **Validate Input**: Always validate and sanitize input data
4. **Rate Limiting**: Respect rate limits and implement exponential backoff
5. **Error Handling**: Never expose sensitive information in error messages

### Performance Best Practices

1. **Connection Pooling**: Reuse HTTP connections when possible
2. **Caching**: Cache responses when appropriate
3. **Pagination**: Use pagination for large result sets
4. **Compression**: Enable gzip compression for responses
5. **Monitoring**: Monitor API usage and performance metrics

### Integration Best Practices

1. **Idempotency**: Ensure operations are idempotent where possible
2. **Retry Logic**: Implement exponential backoff for retries
3. **Circuit Breaker**: Use circuit breaker pattern for fault tolerance
4. **Logging**: Log all API interactions for debugging
5. **Health Checks**: Implement health checks for your integration

## Support and Resources

### Documentation

- [API Reference](endpoints/)
- [SDK Documentation](sdks/)
- [Integration Guides](integrations/)
- [Troubleshooting](troubleshooting.md)

### Community

- [GitHub Issues](https://github.com/pramodksahoo/kubechat/issues)
- [Community Forum](https://github.com/pramodksahoo/kubechat/discussions)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/kubechat)

### Enterprise Support

For enterprise support and custom integrations:
- Email: [enterprise@kubechat.dev](mailto:enterprise@kubechat.dev)
- Documentation: [Enterprise API Guide](enterprise/)

## Changelog

### v1.0.0 (2025-01-15)
- Initial API release
- Core query processing endpoints
- JWT authentication
- Audit trail API
- WebSocket support

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.
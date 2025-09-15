/** @type {import('next').NextConfig} */
const nextConfig = {
  // Output configuration for container deployment
  output: 'standalone',
  
  // Transpile shared packages
  transpilePackages: ['@kubechat/shared', '@kubechat/ui'],
  
  // Container-first configuration
  env: {
    NEXT_TELEMETRY_DISABLED: '1',
  },
  
  // Experimental features for performance
  experimental: {
    serverComponentsExternalPackages: [],
  },
  
  // Build configuration
  poweredByHeader: false,
  compress: true,
  
  // ESLint configuration - temporarily disabled for production build
  eslint: {
    ignoreDuringBuilds: true,
  },
  
  // Image optimization
  images: {
    unoptimized: false, // Enable optimization for better performance
    domains: [], // Add any external domains here if needed
    formats: ['image/webp', 'image/avif'],
  },
  
  // Webpack configuration
  webpack: (config, { dev, isServer }) => {
    // Optimize bundle for container deployment
    if (!dev && !isServer) {
      config.optimization.splitChunks.chunks = 'all';
    }
    
    return config;
  },
  
  // Headers for security and caching
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY'
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff'
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin'
          },
          {
            key: 'Cache-Control',
            value: 'no-cache, no-store, must-revalidate'
          },
          {
            key: 'Pragma',
            value: 'no-cache'
          },
          {
            key: 'Expires',
            value: '0'
          }
        ]
      },
      {
        source: '/_next/static/(.*)',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable'
          }
        ]
      }
    ];
  },
  
  // API proxying for Kubernetes service-to-service communication
  async rewrites() {
    // Use environment variable for backend service URL - no hardcoding
    const apiUrl = process.env.API_URL || process.env.BACKEND_SERVICE_URL;

    if (!apiUrl) {
      console.warn('API_URL or BACKEND_SERVICE_URL environment variable not set - API proxying will not work');
      return [];
    }

    return [
      // WebSocket endpoint (handled by ingress, but included for completeness)
      {
        source: '/ws',
        destination: `${apiUrl}/ws`
      },
      // API v1 endpoints
      {
        source: '/api/v1/:path*',
        destination: `${apiUrl}/api/v1/:path*`
      },
      // Auth endpoints
      {
        source: '/auth/:path*',
        destination: `${apiUrl}/auth/:path*`
      },
      // Audit endpoints
      {
        source: '/audit/:path*',
        destination: `${apiUrl}/audit/:path*`
      },
      // Kubernetes endpoints
      {
        source: '/kubernetes/:path*',
        destination: `${apiUrl}/kubernetes/:path*`
      },
      // Security endpoints
      {
        source: '/security/:path*',
        destination: `${apiUrl}/security/:path*`
      },
      // NLP endpoints
      {
        source: '/nlp/:path*',
        destination: `${apiUrl}/nlp/:path*`
      },
      // Database endpoints
      {
        source: '/database/:path*',
        destination: `${apiUrl}/database/:path*`
      },
      // Performance endpoints
      {
        source: '/performance/:path*',
        destination: `${apiUrl}/performance/:path*`
      },
      // WebSocket endpoints
      {
        source: '/websocket/:path*',
        destination: `${apiUrl}/websocket/:path*`
      },
      // Communication endpoints
      {
        source: '/communication/:path*',
        destination: `${apiUrl}/communication/:path*`
      },
      // Gateway endpoints
      {
        source: '/gateway/:path*',
        destination: `${apiUrl}/gateway/:path*`
      },
      // External API endpoints
      {
        source: '/external/:path*',
        destination: `${apiUrl}/external/:path*`
      },
      // General API endpoints (fallback)
      {
        source: '/api/:path*',
        destination: `${apiUrl}/api/:path*`
      },
      // Health check endpoints
      {
        source: '/health/:path*',
        destination: `${apiUrl}/health/:path*`
      },
      {
        source: '/health',
        destination: `${apiUrl}/health`
      },
      // Status endpoint
      {
        source: '/status',
        destination: `${apiUrl}/status`
      }
    ];
  }
};

module.exports = nextConfig;
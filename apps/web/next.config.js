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
    const apiUrl = process.env.API_URL;
    
    if (!apiUrl) {
      console.warn('API_URL environment variable not set - API proxying will not work');
      return [];
    }
    
    return [
      // Proxy all API calls to the backend service
      {
        source: '/api/:path*',
        destination: `${apiUrl}/api/:path*`
      },
      // Health check endpoint
      {
        source: '/health',
        destination: `${apiUrl}/api/health`
      }
    ];
  }
};

module.exports = nextConfig;
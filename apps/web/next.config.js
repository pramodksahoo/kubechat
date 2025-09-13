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
  
  // Image optimization
  images: {
    unoptimized: false, // Enable optimization for better performance
    domains: [], // Add any external domains here if needed
    formats: ['image/webp', 'image/avif'],
  },
  
  // Webpack configuration
  webpack: (config, { buildId, dev, isServer, defaultLoaders, webpack }) => {
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
  
  // Redirects for container routing
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: process.env.API_URL ? `${process.env.API_URL}/api/:path*` : '/api/:path*'
      }
    ];
  }
};

module.exports = nextConfig;
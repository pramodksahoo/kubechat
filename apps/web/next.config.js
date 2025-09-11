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
    unoptimized: true, // Disable for container deployments
  },
  
  // Webpack configuration
  webpack: (config, { buildId, dev, isServer, defaultLoaders, webpack }) => {
    // Optimize bundle for container deployment
    if (!dev && !isServer) {
      config.optimization.splitChunks.chunks = 'all';
    }
    
    return config;
  },
  
  // Headers for security
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
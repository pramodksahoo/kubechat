// KubeChat Frontend Vitest Configuration
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@kubechat/shared': path.resolve(__dirname, '../../packages/shared'),
      '@kubechat/ui': path.resolve(__dirname, '../../packages/ui'),
      '@kubechat/config': path.resolve(__dirname, '../../packages/config'),
    },
  },
});

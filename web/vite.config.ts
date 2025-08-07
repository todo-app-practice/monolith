import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // This is required to make the dev server accessible from outside the container
    host: '0.0.0.0',
    // This is required for HMR (Hot Module Replacement) to work correctly in a Docker environment
    hmr: {
      clientPort: 5173,
    },
    // Proxy API requests to the backend service to avoid CORS issues
    proxy: {
      '/api': {
        target: 'http://todo-app:8765',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
});

import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    hmr: {
      // Tell the browser's HMR client to connect on the nginx port (3000),
      // not the internal Vite port (5173).
      clientPort: 3000,
    },
  },
})

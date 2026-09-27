import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true, // Listen on all local IPs
    proxy: {
      '/query': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
